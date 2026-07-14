// sc-recall-index is a prosthetic-conscience PostToolUse hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE indexing, it MUST resolve a runner; none resolvable means silent no-op.
//	It MUST only react to markdown writes — code writes never pay the index cost.
//	It MUST NOT fail the Write/Edit, whatever `qmd update` does.
//
// Purpose: deterministic recall freshness. Agents forget to re-index; a hook
// cannot. Every markdown write triggers a fast FTS `qmd update` (~0.7s measured)
// so lexical search is never stale. Semantic embeddings refresh separately
// (phase-top `qmd embed`) — this hook keeps the cheap layer current.
//
// Runner resolution (no separate install required): a global `qmd` binary is a
// pure fast path; otherwise the hook runs the SAME package the project's MCP
// server pins — `npx -y <pkg from .mcp.json>` — so `.mcp.json` stays the single
// source of truth for the qmd version and the hook activates exactly where the
// project declares the recall layer.
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

const version = "0.2.0"

const updateTimeout = 60 * time.Second

type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	} `json:"tool_input"`
}

type mcpConfig struct {
	McpServers map[string]struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	} `json:"mcpServers"`
}

// qmdPackageFrom extracts the pinned qmd package spec (e.g. "@tobilu/qmd@2.5.3")
// from a .mcp.json payload. Empty means the project has not declared qmd.
func qmdPackageFrom(raw []byte) string {
	var cfg mcpConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return ""
	}
	srv, ok := cfg.McpServers["qmd"]
	if !ok {
		return ""
	}
	for _, a := range srv.Args {
		if strings.HasPrefix(a, "-") || a == "mcp" {
			continue
		}
		return a
	}
	return ""
}

// resolveRunner is the pure, unit-tested ladder: global binary first (fast),
// else npx with the project's pinned package, else nothing.
func resolveRunner(qmdPresent, npxPresent bool, pkg string) []string {
	if qmdPresent {
		return []string{"qmd"}
	}
	if npxPresent && pkg != "" {
		return []string{"npx", "-y", pkg}
	}
	return nil
}

// decide is the pure, unit-tested gate: index only when a runner resolved and
// the write touched markdown. The returned reason is for the hook log.
func decide(runner []string, file string) (run bool, reason string) {
	if len(runner) == 0 {
		return false, "no qmd runner (binary or npx+.mcp.json pin) — recall index skipped. file=" + file
	}
	if !strings.EqualFold(filepath.Ext(file), ".md") {
		return false, "not markdown — recall index skipped. file=" + file
	}
	return true, "markdown write — running " + strings.Join(runner, " ") + " update. file=" + file
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

func projectPackage(projectDir string) string {
	if projectDir == "" {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
	if err != nil {
		return ""
	}
	return qmdPackageFrom(raw)
}

func runUpdate(runner []string) string {
	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()
	args := append(runner[1:], "update")
	if err := exec.CommandContext(ctx, runner[0], args...).Run(); err != nil {
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
	projectDir := os.Getenv("CLAUDE_PROJECT_DIR")

	runner := resolveRunner(toolchain.Present("qmd"), toolchain.Present("npx"), projectPackage(projectDir))
	run, reason := decide(runner, file)
	outcome := ""
	if run {
		outcome = " | " + runUpdate(runner)
	}

	if projectDir != "" {
		logDir := filepath.Join(projectDir, ".claude")
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
	// (the .mcp.json pin is the opt-in). Never block the tool.
	os.Exit(0)
}
