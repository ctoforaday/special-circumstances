// sc-quality-gate is a prosthetic-conscience PostToolUse hook (Go binary).
//
// Contract (Design by Contract), split by cause:
//
//	BEFORE running the quality tool, it MUST confirm the tool is present AND that
//	the project is qlty-initialised; when either is missing it MUST no-op.
//	Infrastructure problems (absent tool, absent config, timeout, a qlty crash)
//	MUST NOT fail the Edit/Write — they exit 0.
//	FINDINGS, by contrast, MUST reach the model: lint issues, or a file that
//	`qlty fmt` rewrote, exit 2 so Claude Code feeds stderr back. Exit 2 at
//	PostToolUse does not undo the write; it is the feedback channel.
//
// It reads the PostToolUse hook JSON on stdin, records that it fired (which proves
// the hook reaches here — including inside a subagent), then runs `qlty fmt` and
// `qlty check` on the edited file.
package qualitygate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hooklog"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

// Modes for SC_QUALITY_GATE. Formatting rewrites files, so it gets its own
// off-switch separate from disabling the gate outright.
const (
	modeFull  = "full"       // qlty fmt, then qlty check
	modeCheck = "check-only" // qlty check only, no rewriting
	modeOff   = "off"
)

const (
	defaultTimeout = 30 * time.Second // cold qlty installs its linters: measured 29s
	maxTimeout     = 10 * time.Minute
	maxFeedbackLen = 4000 // characters of qlty output fed back
	maxFeedbackLn  = 60   // ...and lines, whichever bites first
)

type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	} `json:"tool_input"`
	CWD string `json:"cwd"`
}

// runner is the process boundary, injected so both the healthy and the failing
// branches are reachable in a test without qlty being installed on the machine.
type runner func(ctx context.Context, dir, name string, args ...string) (output string, exitCode int, err error)

// env is everything sc-quality-gate learns from outside itself.
type env struct {
	projectDir  string
	qltyPresent bool
	mode        string
	timeout     time.Duration
	exec        runner
}

// plan is the decision: run qlty, or skip with a reason.
type plan struct {
	run    bool
	format bool
	rel    string // path handed to qlty, relative to the project dir
	skip   string // why not, for the log
	warn   bool   // ...and whether that reason is worth a line on stderr
}

func parseMode(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "off", "0", "false", "disabled", "no":
		return modeOff
	case "check", "check-only", "checkonly":
		return modeCheck
	default:
		return modeFull
	}
}

func parseTimeout(v string) time.Duration {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n <= 0 {
		return defaultTimeout
	}
	if d := time.Duration(n) * time.Second; d < maxTimeout {
		return d
	}
	return maxTimeout
}

// decide is the gate: which of the preconditions holds, and what to run.
// Order matters — the checks run cheapest and most-actionable first.
func decide(e env, file string) plan {
	if e.mode == modeOff {
		return plan{skip: "disabled by SC_QUALITY_GATE=off"}
	}
	if file == "" || file == "unknown" {
		return plan{skip: "no file in the payload"}
	}
	if !e.qltyPresent {
		// Absent tool is broken machine state and fixable: worth one line.
		return plan{skip: "qlty not found — quality gate skipped (run /prosthetic-conscience:doctor --fix). file=" + file, warn: true}
	}
	if e.projectDir == "" {
		return plan{skip: "no CLAUDE_PROJECT_DIR"}
	}
	// qlty refuses to run outside an initialised workspace, and does it by
	// printing a Rust backtrace — never let that reach the operator.
	if _, err := os.Stat(filepath.Join(e.projectDir, ".qlty", "qlty.toml")); err != nil {
		// A project without qlty config chose that; nagging every edit is noise.
		return plan{skip: "project is not qlty-initialised (`qlty init`)"}
	}
	rel, ok := within(e.projectDir, file)
	if !ok {
		return plan{skip: "file is outside the project dir"}
	}
	if st, err := os.Stat(filepath.Join(e.projectDir, rel)); err != nil || st.IsDir() {
		return plan{skip: "not a readable file"}
	}
	return plan{run: true, format: e.mode == modeFull, rel: rel}
}

// within resolves file against dir and reports whether it stays inside it.
//
// FIX (defect): containment used to be pure path arithmetic — filepath.Rel plus a
// ".." prefix check — which is blind to symlinks by construction. With a link inside
// the project pointing at a sibling tree, Rel reports a clean relative path and this
// gate then hands it to `qlty fmt`, which REWRITES it. That is a modification outside
// the working tree, which agent-guardrails says requires explicit approval, performed
// silently by a hook. internal/checkpoint closed the same class at 67eb8fa; this call
// site kept the arithmetic, so the class outlived the instance fix.
//
// The lexical pass stays: it normalises a relative path against the project and drops
// the obvious climbs cheaply. os.Root is what makes containment a syscall-level fact —
// it refuses a path that leaves the root through ANY component, symlink included.
//
// Absent-but-contained is deliberately NOT an escape. The caller distinguishes
// "outside the project" from "not a readable file", and collapsing the two would make
// this fix report the wrong cause for an ordinary missing file.
func within(dir, file string) (string, bool) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	absFile := file
	if !filepath.IsAbs(absFile) {
		absFile = filepath.Join(absDir, absFile)
	}
	absFile, err = filepath.Abs(absFile)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(absDir, absFile)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	root, err := os.OpenRoot(absDir)
	if err != nil {
		return "", false // a project root we cannot open is one we cannot vouch for
	}
	defer func() { _ = root.Close() }()
	if _, err := root.Stat(rel); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", false
	}
	return rel, true
}

// digest fingerprints a file so a rewrite by `qlty fmt` is detectable. An
// unreadable file yields an empty digest, which compares unequal to nothing.
func digest(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// execQlty is the real process boundary.
func execQlty(ctx context.Context, dir, name string, args ...string) (string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	// Killing qlty on the deadline is not enough: a linter it spawned inherits the
	// output pipe and keeps CombinedOutput blocked until that grandchild exits.
	// WaitDelay force-closes the pipes so the timeout is the real ceiling.
	cmd.WaitDelay = 2 * time.Second
	out, err := cmd.CombinedOutput()
	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	if ctx.Err() != nil {
		return string(out), code, ctx.Err()
	}
	if _, ok := err.(*exec.ExitError); ok {
		err = nil // a non-zero exit is data here, not a failure to run
	}
	return string(out), code, err
}

// clamp keeps a wide finding set from flooding the model's context.
func clamp(s string) string {
	s = strings.TrimSpace(s)
	if lines := strings.Split(s, "\n"); len(lines) > maxFeedbackLn {
		s = strings.Join(lines[:maxFeedbackLn], "\n") + "\n… (truncated; run `qlty check` for the rest)"
	}
	if len(s) > maxFeedbackLen {
		s = s[:maxFeedbackLen] + "\n… (truncated; run `qlty check` for the rest)"
	}
	return s
}

// execute runs the planned qlty commands and returns the feedback for the model
// (empty when there is nothing to say) plus the one-line summary for the log.
func execute(e env, p plan) (feedback, logged string) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	abs := filepath.Join(e.projectDir, p.rel)
	var parts []string

	if p.format {
		before := digest(abs)
		if _, _, err := e.exec(ctx, e.projectDir, "qlty", "fmt", "--no-progress", "--trigger", "agent", p.rel); err != nil {
			return "", "qlty fmt could not run: " + err.Error()
		}
		if after := digest(abs); after != before && after != "" {
			// The model's cached copy of the file is now stale; an Edit against
			// it would fail on a string mismatch. Say so before it tries.
			// Criterion: say what DID happen, and what the agent must do about it.
			parts = append(parts, "qlty fmt rewrote "+p.rel+" on disk. Your cached copy is now stale: RE-READ it before your next Edit, or the edit fails on a string mismatch. Nothing was lost — the formatter only reformatted what you wrote.")
		}
	}

	out, code, err := e.exec(ctx, e.projectDir, "qlty", "check", "--no-progress", "--no-fix", "--no-error", p.rel)
	if err != nil {
		// A timeout or an unlaunchable qlty is infrastructure: report the
		// format rewrite if there was one, but never fail on this.
		return strings.Join(parts, "\n"), "qlty check could not run: " + err.Error()
	}
	if code != 0 {
		// THE THING AN AGENT GETS WRONG HERE. Exit 2 at PostToolUse is a FEEDBACK channel,
		// not a rejection — the write already happened and was not reverted. An agent that
		// reads a non-zero exit as "my edit failed" re-applies it, which duplicates the
		// change or fights the formatter. Say so before the findings, not after.
		parts = append(parts,
			"qlty check found issues in "+p.rel+". YOUR WRITE ALREADY HAPPENED and was not reverted — this is feedback on it, not a rejection, so do NOT re-apply the edit. "+
				"Fix these in a follow-up edit; run `qlty check --no-fix "+p.rel+"` to see the full set:",
			clamp(out))
	}

	switch {
	case len(parts) == 0:
		return "", "clean"
	case code != 0:
		return strings.Join(parts, "\n"), "issues reported"
	default:
		return strings.Join(parts, "\n"), "formatted"
	}
}

// fileFromInput reads the subject out of a raw tool_input, for the merged binary which
// parses the payload once and hands the fragment on.
func fileFromInput(raw json.RawMessage) string {
	var ti struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	}
	_ = json.Unmarshal(raw, &ti)
	var in hookInput
	in.ToolInput.FilePath, in.ToolInput.Path = ti.FilePath, ti.Path
	return fileFrom(in)
}

func fileFrom(in hookInput) string {
	if in.ToolInput.FilePath != "" {
		return in.ToolInput.FilePath
	}
	if in.ToolInput.Path != "" {
		return in.ToolInput.Path
	}
	return "unknown"
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, e env) int {
	fs := flag.NewFlagSet("sc-quality-gate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag is never worth failing the Edit/Write over
	}
	if *showVersion {
		fmt.Fprintln(stdout, buildid.Line("sc-quality-gate"))
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	e.projectDir = hookenv.ProjectDir(e.projectDir, in.CWD)
	file := fileFrom(in)

	out := gate(e, in.ToolName, file, time.Now())
	hooklog.Append(e.projectDir, out.Log)
	if out.Stderr != "" {
		fmt.Fprintln(stderr, out.Stderr)
	}
	return out.Exit
}

// gate is the unit's whole decision, with no I/O of its own: the standalone binary and the
// merged PostToolUse binary both run THIS, so the two paths cannot drift into disagreeing
// about what the gate does.
//
// The log line is RETURNED rather than written. Two processes used to append to one log
// concurrently — proven concurrent by design once hooks were measured to run in parallel
// (hook-surface-spike.md §4b) — and collecting lines for one writer is the fix.
func gate(e env, toolName, file string, now time.Time) hookunit.Result {
	p := decide(e, file)
	msg := p.skip
	feedback := ""
	if p.run {
		feedback, msg = execute(e, p)
	}
	// Instrumentation: record every firing (proves the hook reaches here, incl. subagents).
	r := hookunit.Result{Name: "sc-quality-gate",
		Log: fmt.Sprintf("%s sc-quality-gate %s -> %s | %s\n", now.UTC().Format(time.RFC3339), toolName, file, msg)}
	switch {
	case p.warn:
		r.Stderr = p.skip
	case feedback != "":
		// Exit 2 is how a PostToolUse hook hands stderr back to the model. The write
		// already happened — this reports on it, it does not revoke it.
		r.Stderr, r.Exit = feedback, 2
	}
	return r
}

// Unit exposes the gate to the merged PostToolUse binary, wired to the real environment.
func Unit() hookunit.Unit {
	e := env{
		qltyPresent: toolchain.Present("qlty"),
		mode:        parseMode(os.Getenv("SC_QUALITY_GATE")),
		timeout:     parseTimeout(os.Getenv("SC_QUALITY_GATE_TIMEOUT_SECONDS")),
		exec:        execQlty,
	}
	return hookunit.Unit{
		Name:    "sc-quality-gate",
		Applies: func(c *hookunit.Ctx) bool { return c.ToolName == "Write" || c.ToolName == "Edit" },
		Run: func(c *hookunit.Ctx) hookunit.Result {
			e.projectDir = c.ProjectDir
			return gate(e, c.ToolName, fileFromInput(c.ToolInput), c.Now)
		},
	}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, env{
		projectDir:  os.Getenv("CLAUDE_PROJECT_DIR"),
		qltyPresent: toolchain.Present("qlty"),
		mode:        parseMode(os.Getenv("SC_QUALITY_GATE")),
		timeout:     parseTimeout(os.Getenv("SC_QUALITY_GATE_TIMEOUT_SECONDS")),
		exec:        execQlty,
	})
}
