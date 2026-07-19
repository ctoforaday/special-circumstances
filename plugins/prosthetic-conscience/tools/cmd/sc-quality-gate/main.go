// sc-quality-gate is a prosthetic-conscience PostToolUse hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE running the quality tool, it MUST confirm the tool is present.
//	When the tool is absent, it MUST no-op with one warning — it MUST NOT fail the Edit/Write.
//
// It reads the PostToolUse hook JSON on stdin, records that it fired (which proves
// the hook reaches here — including inside a subagent), and gates on `qlty`.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

const version = "0.1.0"

type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	} `json:"tool_input"`
}

// gate is the pure, unit-tested decision: given whether qlty is present and the
// file being written, return the line to record/emit. It never blocks the tool.
func gate(qltyPresent bool, file string) string {
	if !qltyPresent {
		return "qlty not found — quality gate skipped (run /prosthetic-conscience:doctor --fix). file=" + file
	}
	return "qlty present — would run `qlty fmt` + `qlty check` on " + file
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

// appendHookLog records one firing under projectDir/.claude. Instrumentation is
// best-effort by design: a log that cannot be written must never cost the tool call,
// so every error here is deliberately swallowed. An empty projectDir disables it
// rather than writing a relative .claude/ wherever the process happens to be.
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

// run is the hook with its process boundary passed in. qltyPresent is a parameter
// rather than a lookup so the degraded and healthy branches are both reachable in a
// test regardless of what is installed on the machine running it.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, qltyPresent bool) int {
	fs := flag.NewFlagSet("sc-quality-gate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag is never worth failing the Edit/Write over
	}
	if *showVersion {
		fmt.Fprintln(stdout, "sc-quality-gate", version)
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	file := fileFrom(in)

	msg := gate(qltyPresent, file)

	// Instrumentation: record every firing (proves the hook reaches here, incl. subagents).
	appendHookLog(projectDir, fmt.Sprintf("%s sc-quality-gate %s -> %s | %s\n",
		time.Now().UTC().Format(time.RFC3339), in.ToolName, file, msg))

	// Terse warning to stderr only when degraded; never block the tool call.
	if !qltyPresent {
		fmt.Fprintln(stderr, msg)
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), toolchain.Present("qlty")))
}
