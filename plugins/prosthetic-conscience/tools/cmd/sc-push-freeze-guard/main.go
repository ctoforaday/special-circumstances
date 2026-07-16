// sc-push-freeze-guard is a prosthetic-conscience PreToolUse hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE a Bash `git push` runs while a research run is LIVE (.run-live marker),
//	the operator MUST see one warning naming the frozen paths. It NEVER blocks —
//	the freeze is a commitment the human may consciously override; the guard's job
//	is making the commitment impossible to forget, not impossible to break.
//	No marker, no output. Non-push commands, no output.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
)

const version = "0.1.0"

type hookInput struct {
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// decide is the pure, unit-tested gate: warn only for a git push while live.
func decide(m *runlive.Marker, command string) string {
	if m == nil || !strings.Contains(command, "git push") {
		return ""
	}
	return fmt.Sprintf("sc-push-freeze-guard: a research run is LIVE (%s, started %s) — pinned paths are FROZEN: %s. Push only if it touches none of them (run-capture lifts the freeze).",
		m.RunDir, m.Started, strings.Join(m.PinnedPaths, ", "))
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("sc-push-freeze-guard", version)
		return
	}

	raw, _ := io.ReadAll(os.Stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)

	if line := decide(runlive.Read(os.Getenv("CLAUDE_PROJECT_DIR")), in.ToolInput.Command); line != "" {
		fmt.Fprintln(os.Stderr, line)
	}
	os.Exit(0)
}
