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

func readStamp(dir string) time.Time {
	fi, err := os.Stat(filepath.Join(dir, stampName))
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}

func writeStamp(dir string) {
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
		stampDir := filepath.Join(os.Getenv("CLAUDE_PROJECT_DIR"), ".claude")
		if shouldUpdate(time.Now(), readStamp(stampDir), debounceInterval) {
			outcome = " | " + runUpdate()
			writeStamp(stampDir)
		} else {
			outcome = " | debounced (index at most " + debounceInterval.String() + " stale)"
		}
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
