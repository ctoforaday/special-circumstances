// sc-recall-index is a prosthetic-conscience PostToolUse hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE indexing, it MUST confirm qmd is present; absent means silent no-op.
//	It MUST only react to markdown writes — code writes never pay the index cost.
//	It MUST NOT fail the Write/Edit, whatever `qmd update` does.
//
// Purpose: deterministic recall freshness. Agents forget to re-index; a hook
// cannot. Every markdown write triggers a fast FTS `qmd update` (~0.7s measured)
// so BM25 search is never stale. Semantic embeddings refresh separately
// (phase-top `qmd embed`) — this hook keeps the cheap layer current.
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

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

const version = "0.1.0"

const updateTimeout = 30 * time.Second

type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	} `json:"tool_input"`
}

// decide is the pure, unit-tested gate: index only when qmd exists and the
// write touched markdown. The returned reason is for the hook log.
func decide(qmdPresent bool, file string) (run bool, reason string) {
	if !qmdPresent {
		return false, "qmd not found — recall index skipped. file=" + file
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

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("sc-recall-index", version)
		return
	}

	raw, _ := io.ReadAll(os.Stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	file := fileFrom(in)

	run, reason := decide(toolchain.Present("qmd"), file)
	outcome := ""
	if run {
		outcome = " | " + runUpdate()
	}

	if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" {
		logDir := filepath.Join(dir, ".claude")
		if err := os.MkdirAll(logDir, 0o755); err == nil {
			line := fmt.Sprintf("%s sc-recall-index %s -> %s | %s%s\n",
				time.Now().UTC().Format(time.RFC3339), in.ToolName, file, reason, outcome)
			if f, err := os.OpenFile(filepath.Join(logDir, "prosthetic-conscience-hook.log"),
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
				_, _ = f.WriteString(line)
				_ = f.Close()
			}
		}
	}

	// Recall is an optional capability: absence is never worth a per-write nag
	// (sc-toolchain-nudge covers discovery at SessionStart). Never block the tool.
	os.Exit(0)
}
