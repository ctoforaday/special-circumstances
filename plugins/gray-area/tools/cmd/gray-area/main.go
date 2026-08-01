// gray-area is the trajectory miner's command-line surface.
//
// Contract (Design by Contract):
//
//	Every row it emits MUST carry provenance — file, line, uuid — so a consumer
//	can drop to the raw trajectory and check. An event that cannot be cited is
//	NEVER emitted; it is counted and reported as suppressed.
//	AFTER a parse that could not account for every line, it MUST say so on
//	stderr, because an answer over a subset that looks like an answer over the
//	whole is the failure this tool exists to prevent.
//
// The rule it enforces, from plans/gray-area.md: exploration may summarize,
// adjudication must cite. This binary is the citing half. It reports what was
// invoked, where, and by which seat — it does not interpret, score, or conclude.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/claims"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/trajectory"
)

const version = "0.2.0"

const usage = `gray-area — trajectory evidence

  gray-area tools <transcript.jsonl>   list every tool invocation, with provenance
  gray-area checkpoint <note.md> [transcript.jsonl]
                                      adjudicate a checkpoint's validation-loop
                                      claims against what the session actually ran.
                                      With no transcript, resolves this session's
                                      from gray-area's own manifest and prints
                                      which row it used

Flags:
  -binary <name>   also resolve shell-aliased invocations of <name>
                   (a seat that runs REC=./tool ; "$REC" verb is invisible to a
                    matcher that greps for the binary name)
  -json            emit one JSON object per row instead of a table
  -version         print version and exit

checkpoint verdicts:
  CITED        a matching invocation is in the trajectory, with its uuid
  STALE        a write under the check's own trigger surface happened AFTER the
               claimed run time — the pass may have been real; it is not current
  NO-EVIDENCE  nothing matched. NOT "did not run": the tokens searched and the
               number of events searched are printed so the reader can judge
  UNCHECKABLE  the claim names no command, so there was nothing to look for

Exit is non-zero when anything is STALE or NO-EVIDENCE.
`

// row is what the tool emits. Provenance fields are first because they are the
// point: a reader quotes them.
type row struct {
	File        string                  `json:"file"`
	Line        int                     `json:"line"`
	UUID        string                  `json:"uuid"`
	Timestamp   string                  `json:"timestamp,omitempty"`
	AgentType   string                  `json:"agent_type,omitempty"`
	Tool        string                  `json:"tool"`
	ToolUseID   string                  `json:"tool_use_id,omitempty"`
	Target      string                  `json:"target,omitempty"`
	Invocations []trajectory.Invocation `json:"invocations,omitempty"`
}

// target is the single most useful identifier of what a call acted on. It is a
// convenience for reading, never a substitute for the input itself — which is
// why the uuid and line travel with it.
func target(e trajectory.Event) string {
	var in struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
		Pattern  string `json:"pattern"`
		URL      string `json:"url"`
	}
	if len(e.ToolInput) > 0 {
		_ = json.Unmarshal(e.ToolInput, &in)
	}
	for _, v := range []string{in.Command, in.FilePath, in.Path, in.Pattern, in.URL} {
		if v != "" {
			return v
		}
	}
	return ""
}

func rowsFrom(res trajectory.Result, binary string) (rows []row, suppressed int) {
	for _, e := range res.ToolUses() {
		if !e.Cited() {
			// Refuse to emit what cannot be checked (see the package contract).
			suppressed++
			continue
		}
		r := row{
			File: e.File, Line: e.Line, UUID: e.UUID, Timestamp: e.Timestamp,
			AgentType: e.AgentType, Tool: e.ToolName, ToolUseID: e.ToolUseID,
			Target: target(e),
		}
		if binary != "" && e.ToolName == "Bash" {
			r.Invocations = trajectory.FindInvocations(target(e), binary)
		}
		rows = append(rows, r)
	}
	return rows, suppressed
}

// splitCommand takes the subcommand from the FRONT, then leaves the rest to the
// flag package — the shape git, go and docker use.
//
// Two bugs found by running this rather than by testing it. First, Go's flag
// package stops parsing at the first non-flag argument, so a plain fs.Parse
// silently ignored `-binary` in `tools -binary X file`, which is the order
// anyone actually types. Second, the obvious fix — hunt for the first non-flag
// token anywhere — mistook a flag's VALUE for the subcommand in
// `-binary feov-record tools file`. Supporting both orders cannot be done
// without knowing which flags take values, so it is not supported: the
// subcommand comes first, and anything else is an error rather than a guess.
func splitCommand(args []string) (cmd string, rest []string) {
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		return args[0], args[1:]
	}
	return "", args
}

func run(args []string, stdout, stderr io.Writer, open func(string) (io.ReadCloser, error)) int {
	cmd, flags := splitCommand(args)

	fs := flag.NewFlagSet("gray-area", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usage) }
	binary := fs.String("binary", "", "resolve shell-aliased invocations of this binary")
	asJSON := fs.Bool("json", false, "emit JSON rows")
	showVersion := fs.Bool("version", false, "print version and exit")
	projectDir := fs.String("project", "", "project root whose .claude/gray-area/ manifest resolves this session's transcript (default: CLAUDE_PROJECT_DIR, else the working directory)")
	if err := fs.Parse(flags); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintln(stdout, "gray-area", version)
		return 0
	}
	if cmd == "" {
		fs.Usage()
		return 2
	}
	if cmd != "tools" && cmd != "checkpoint" {
		fmt.Fprintf(stderr, "gray-area: unknown command %q\n", cmd)
		fs.Usage()
		return 2
	}
	rest := fs.Args()

	if cmd == "checkpoint" {
		if len(rest) < 1 {
			fmt.Fprintln(stderr, "gray-area checkpoint: a note path is required")
			return 2
		}
		trace := ""
		if len(rest) > 1 {
			trace = rest[1]
		}
		return checkpoint(rest[0], trace, stdout, stderr, open, *asJSON, *projectDir)
	}

	if len(rest) < 1 {
		fmt.Fprintln(stderr, "gray-area tools: a transcript path is required")
		return 2
	}

	path := rest[0]
	f, err := open(path)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: cannot open %s: %v\n", path, err)
		return 1
	}
	defer func() { _ = f.Close() }()

	res, err := trajectory.Parse(f, path)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: reading %s: %v\n", path, err)
		return 1
	}

	rows, suppressed := rowsFrom(res, *binary)

	if *asJSON {
		enc := json.NewEncoder(stdout)
		for _, r := range rows {
			if err := enc.Encode(r); err != nil {
				fmt.Fprintln(stderr, "gray-area: encoding row:", err)
				return 1
			}
		}
	} else {
		for _, r := range rows {
			seat := r.AgentType
			if seat == "" {
				seat = "-"
			}
			fmt.Fprintf(stdout, "%s:%d %s %-8s %-6s %s\n",
				r.File, r.Line, short(r.UUID), seat, r.Tool, truncate(r.Target, 80))
			for _, inv := range r.Invocations {
				via := ""
				if inv.Aliased {
					via = fmt.Sprintf(" (aliased via $%s)", inv.Via)
				}
				fmt.Fprintf(stdout, "    -> %s %s%s\n", inv.Binary, inv.Verb, via)
			}
		}
	}

	// Coverage is reported even on success: an answer over a subset that reads
	// like an answer over the whole is exactly what this must not produce.
	if res.BadLines > 0 {
		fmt.Fprintf(stderr, "gray-area: WARNING %d of %d lines did not parse (first at %v) — this listing is over a SUBSET of %s\n",
			res.BadLines, res.Lines, res.BadSample, path)
	}
	if suppressed > 0 {
		fmt.Fprintf(stderr, "gray-area: WARNING %d invocation(s) suppressed for missing provenance — not citable, so not shown\n", suppressed)
	}
	return 0
}

func short(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func truncate(s string, n int) string {
	s = string([]rune(s))
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, func(p string) (io.ReadCloser, error) {
		return os.Open(p)
	}))
}

// defaultProjectDir resolves the project root for the manifest lookup.
//
// A WORKING-DIRECTORY FALLBACK IS CORRECT HERE and wrong in a hook, so the
// divergence is deliberate rather than an oversight. A hook is spawned by the
// harness with no meaningful cwd, which is why prosthetic-conscience's hookenv
// refuses os.Getwd() — there, "somewhere" is unrelated to the session. This is a
// CLI a human runs from inside the repo, where the working directory is exactly
// the statement of which project they mean.
func defaultProjectDir(flagValue, env, wd string) string {
	if flagValue != "" {
		return flagValue
	}
	if env != "" {
		return env
	}
	return wd
}

// checkpoint adjudicates a sealed note's validation-loop claims against the
// trajectory of the session that wrote them.
//
// It exits 1 when any claim is STALE or has NO-EVIDENCE, so this composes into a
// gate. It exits 1 on an unreadable note too: a note whose format has moved out
// from under the parser must not read as "no claims to check".
func checkpoint(notePath, tracePath string, stdout, stderr io.Writer, open func(string) (io.ReadCloser, error), asJSON bool, projectDir string) int {
	// RESOLVED, NOT GUESSED. With no transcript argument the path comes from this
	// plugin's own manifest — written at SessionStart from what the harness handed
	// over — never from a glob of ~/.claude/projects/ (plans/gray-area.md §3).
	var resolved claims.SessionRow
	if tracePath == "" {
		var err error
		wd, _ := os.Getwd()
		root := defaultProjectDir(projectDir, os.Getenv("CLAUDE_PROJECT_DIR"), wd)
		resolved, err = claims.ResolveSession(claims.ManifestDir(root), claims.GlobFS, os.ReadFile, statReadable)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if !resolved.Resolved {
			fmt.Fprintf(stderr, "gray-area: %s:%d names a transcript that is not readable (%s) — adjudicating against it would report NO-EVIDENCE for every claim, which is a lie about the session rather than about the note\n",
				resolved.Manifest, resolved.Line, resolved.CaptureError)
			return 1
		}
		tracePath = resolved.TranscriptPath
		// The pick is a claim like any other, so it is cited.
		fmt.Fprintf(stdout, "resolved this session's trajectory from %s:%d (session %s, captured %s)\n  -> %s\n",
			resolved.Manifest, resolved.Line, short(resolved.SessionID), resolved.CapturedAt, tracePath)
		if resolved.Recovered {
			// Not a warning — a correction, and worth stating because the row still
			// says otherwise. A reader who greps the manifest will see resolved:false
			// and should know why the tool proceeded anyway (§11.9).
			fmt.Fprintf(stdout, "  (the row records resolved:false — the transcript did not exist when the hook fired and does now; re-checked here)\n")
		}
		fmt.Fprintln(stdout)
	}

	nf, err := open(notePath)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: cannot open %s: %v\n", notePath, err)
		return 1
	}
	declared, err := claims.Parse(nf, notePath)
	_ = nf.Close()
	if err != nil {
		fmt.Fprintln(stderr, "gray-area:", err)
		return 1
	}
	if len(declared) == 0 {
		fmt.Fprintf(stderr, "gray-area: %s carries no validation loop — nothing was declared, so there is nothing to adjudicate\n", notePath)
		return 0
	}

	tf, err := open(tracePath)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: cannot open %s: %v\n", tracePath, err)
		return 1
	}
	res, err := trajectory.Parse(tf, tracePath)
	_ = tf.Close()
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: reading %s: %v\n", tracePath, err)
		return 1
	}

	findings := claims.Adjudicate(declared, res)

	if asJSON {
		enc := json.NewEncoder(stdout)
		for _, f := range findings {
			if err := enc.Encode(f); err != nil {
				fmt.Fprintln(stderr, "gray-area: encoding finding:", err)
				return 1
			}
		}
	} else {
		for _, f := range findings {
			fmt.Fprintf(stdout, "%-11s %s:%d  [%s] %s\n", f.Verdict, f.NoteFile, f.NoteLine, f.Index, truncate(f.Claim, 90))
			if f.ClaimedAt != "" || f.Claimed != "" {
				fmt.Fprintf(stdout, "    claimed: %s\n", truncate(strings.TrimSpace(f.Claimed), 100))
			}
			if f.UUID != "" {
				fmt.Fprintf(stdout, "    evidence: %s:%d %s %s  %s\n", f.File, f.Line, short(f.UUID), f.At, truncate(f.Command, 70))
			}
			if len(f.Searched) > 0 {
				fmt.Fprintf(stdout, "    searched: %v across %d citable events\n", f.Searched, f.EventsSeen)
			}
			if f.Note != "" {
				fmt.Fprintf(stdout, "    note: %s\n", f.Note)
			}
		}
	}

	// Coverage, reported the same way `tools` reports it: an adjudication over a
	// subset that reads like one over the whole is the failure to avoid.
	if res.BadLines > 0 {
		fmt.Fprintf(stderr, "gray-area: WARNING %d of %d lines did not parse (first at %v) — this adjudication is over a SUBSET of %s\n",
			res.BadLines, res.Lines, res.BadSample, tracePath)
	}

	for _, f := range findings {
		if f.Verdict == claims.Stale || f.Verdict == claims.NoEvidence {
			return 1
		}
	}
	return 0
}

// statReadable reports nil when a path is a readable regular file. It is the
// production stat for ResolveSession's re-check: a manifest row's `resolved` is
// an observation made at capture time, and on the first SessionStart of a new
// session id the transcript does not exist yet (plans/gray-area.md §11.9).
func statReadable(p string) error {
	st, err := os.Stat(p)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%s is a directory, not a transcript", p)
	}
	return nil
}
