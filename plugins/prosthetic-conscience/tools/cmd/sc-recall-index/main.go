// sc-recall-index is a prosthetic-conscience PostToolUse hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE indexing, it MUST confirm qmd is installed; absent means silent no-op.
//	It MUST only react to markdown writes — code writes never pay the index cost.
//	It MUST NOT fail the Write/Edit, whatever `qmd update` does.
//
// Purpose: deterministic recall freshness. Agents forget to re-index; a hook
// cannot. Markdown writes trigger a fast FTS `qmd update` so lexical search is
// never stale. Semantic embeddings refresh separately (phase-top `qmd embed`).
//
// DEBOUNCED (run-4 wall-clock forensics, 2026-07-17): the synchronous update cost a
// measured ~3.9s x 219 markdown writes = 14 agent-minutes per run, against ZERO in-run
// retrieval that needed per-write freshness. The hook now runs the update at most once
// per debounce interval (timestamp file); writes inside the window are skipped — the
// index is at worst debounceInterval stale, which every retrieval mode tolerates.
//
// Version-skew rule: exactly ONE qmd touches the index — the installed binary.
// The doctor installs it (requirements.json pins the version, consent-gated) and
// the project's MCP server runs the same binary, so the read path (MCP) and the
// write path (this hook) can never disagree about the index schema. There is
// deliberately NO npx fallback: a second resolution path is a second version.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

const version = "0.3.0"

const updateTimeout = 30 * time.Second

// debounceInterval bounds index staleness; stampName is the observable state
// (commitment-as-state again — the debounce survives across hook processes).
const debounceInterval = 60 * time.Second
const stampName = "recall-index-stamp"

// shouldUpdate is the pure, unit-tested debounce gate.
func shouldUpdate(now, stamp time.Time, interval time.Duration) bool {
	return stamp.IsZero() || now.Sub(stamp) >= interval
}

// readStamp returns the last update's completion time, or the zero time when there
// has never been one.
//
// FIX (defect): an empty dir is now rejected instead of resolving to the RELATIVE
// path ".claude/recall-index-stamp". With CLAUDE_PROJECT_DIR unset the hook used to
// read and write its debounce state in whatever directory the process happened to
// start in — polluting unrelated trees and making the debounce depend on the cwd.
// The log writer in this same binary already guarded against exactly that; the
// stamp did not.
func readStamp(dir string) time.Time {
	if dir == "" {
		return time.Time{}
	}
	fi, err := os.Stat(filepath.Join(dir, stampName))
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}

func writeStamp(dir string) {
	if dir == "" {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, stampName), []byte{}, 0o644)
	now := time.Now()
	_ = os.Chtimes(filepath.Join(dir, stampName), now, now)
}

type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	} `json:"tool_input"`
	CWD string `json:"cwd"`
}

// decide is the pure, unit-tested gate: index only when qmd is installed and
// the write touched markdown. The returned reason is for the hook log.
func decide(qmdPresent bool, file string) (run bool, reason string) {
	if !qmdPresent {
		return false, "qmd not installed — recall index skipped (run /prosthetic-conscience:doctor --fix). file=" + file
	}
	if !strings.EqualFold(filepath.Ext(file), ".md") {
		return false, "not markdown — recall index skipped. file=" + file
	}
	return true, "markdown write — running qmd update. file=" + file
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

func runUpdate() string {
	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()
	if err := exec.CommandContext(ctx, "qmd", "update").Run(); err != nil {
		return "qmd update failed: " + err.Error()
	}
	return "qmd update ok"
}

// appendHookLog records one firing under projectDir/.claude. Best-effort by design:
// instrumentation trouble must never cost the Write/Edit, and an empty projectDir
// disables it rather than writing a relative .claude/ wherever the process started.
func appendHookLog(projectDir, line string) {
	if projectDir == "" {
		return
	}
	logDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(logDir, "prosthetic-conscience-hook.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}

// run is the hook with its process boundary passed in. qmdPresent and update are
// parameters rather than lookups so a test can reach every branch without a qmd on
// the machine — and, more importantly, without a test ever mutating a real index.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, qmdPresent bool, now time.Time, update func() string) int {
	fs := flag.NewFlagSet("sc-recall-index", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag is never worth failing the Write/Edit over
	}
	if *showVersion {
		fmt.Fprintln(stdout, "sc-recall-index", version)
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	file := fileFrom(in)

	doRun, reason := decide(qmdPresent, file)
	outcome := ""
	if doRun {
		stampDir := ""
		if projectDir != "" {
			stampDir = filepath.Join(projectDir, ".claude")
		}
		if shouldUpdate(now, readStamp(stampDir), debounceInterval) {
			outcome = " | " + update()
			writeStamp(stampDir)
		} else {
			outcome = " | debounced (index at most " + debounceInterval.String() + " stale)"
		}
	}

	appendHookLog(projectDir, fmt.Sprintf("%s sc-recall-index %s -> %s | %s%s\n",
		now.UTC().Format(time.RFC3339), in.ToolName, file, reason, outcome))

	// Recall is an optional capability: absence is never worth a per-write nag
	// (sc-toolchain-nudge covers discovery at SessionStart). Never block the tool.
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), toolchain.Present("qmd"), time.Now(), runUpdate))
}
