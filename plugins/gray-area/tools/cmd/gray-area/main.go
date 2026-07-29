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

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/trajectory"
)

const version = "0.1.0"

const usage = `gray-area — trajectory evidence

  gray-area tools <transcript.jsonl>   list every tool invocation, with provenance

Flags:
  -binary <name>   also resolve shell-aliased invocations of <name>
                   (a seat that runs REC=./tool ; "$REC" verb is invisible to a
                    matcher that greps for the binary name)
  -json            emit one JSON object per row instead of a table
  -version         print version and exit
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
	if cmd != "tools" {
		fmt.Fprintf(stderr, "gray-area: unknown command %q\n", cmd)
		fs.Usage()
		return 2
	}
	rest := fs.Args()
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
