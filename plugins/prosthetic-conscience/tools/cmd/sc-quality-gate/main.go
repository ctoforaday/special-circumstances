// sc-quality-gate is a prosthetic-conscience PostToolUse hook (Go binary).
//
// Contract (Design by Contract):
//   BEFORE running the quality tool, it MUST confirm the tool is present.
//   When the tool is absent, it MUST no-op with one warning — it MUST NOT fail the Edit/Write.
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
		return "qlty not found — quality gate skipped (run /doctor --fix). file=" + file
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

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("sc-quality-gate", version)
		return
	}

	raw, _ := io.ReadAll(os.Stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	file := fileFrom(in)

	qlty := toolchain.Present("qlty")
	msg := gate(qlty, file)

	// Instrumentation: record every firing (proves the hook reaches here, incl. subagents).
	if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" {
		logDir := filepath.Join(dir, ".claude")
		if err := os.MkdirAll(logDir, 0o755); err == nil {
			line := fmt.Sprintf("%s sc-quality-gate %s -> %s | %s\n",
				time.Now().UTC().Format(time.RFC3339), in.ToolName, file, msg)
			if f, err := os.OpenFile(filepath.Join(logDir, "prosthetic-conscience-hook.log"),
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
				_, _ = f.WriteString(line)
				_ = f.Close()
			}
		}
	}

	// Terse warning to stderr only when degraded; never block the tool call.
	if !qlty {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(0)
}
