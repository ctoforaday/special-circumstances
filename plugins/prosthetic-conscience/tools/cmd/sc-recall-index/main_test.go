package main

import (
	"strings"
	"testing"
)

func TestQmdPackageFrom(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string
	}{
		{"pinned npx entry", `{"mcpServers":{"qmd":{"command":"npx","args":["-y","@tobilu/qmd@2.5.3","mcp"]}}}`, "@tobilu/qmd@2.5.3"},
		{"no qmd server", `{"mcpServers":{"pdf-reader":{"command":"npx","args":["-y","x@1"]}}}`, ""},
		{"malformed json", `{not json`, ""},
		{"flags and subcommand skipped", `{"mcpServers":{"qmd":{"command":"npx","args":["--yes","mcp","@tobilu/qmd@3.0.0"]}}}`, "@tobilu/qmd@3.0.0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := qmdPackageFrom([]byte(c.json)); got != c.want {
				t.Fatalf("qmdPackageFrom = %q; want %q", got, c.want)
			}
		})
	}
}

func TestResolveRunner(t *testing.T) {
	if got := resolveRunner(true, true, "@tobilu/qmd@2.5.3"); len(got) != 1 || got[0] != "qmd" {
		t.Fatalf("global binary must win: %v", got)
	}
	if got := resolveRunner(false, true, "@tobilu/qmd@2.5.3"); strings.Join(got, " ") != "npx -y @tobilu/qmd@2.5.3" {
		t.Fatalf("npx fallback must carry the .mcp.json pin: %v", got)
	}
	if got := resolveRunner(false, true, ""); got != nil {
		t.Fatalf("no .mcp.json pin means no runner (the pin is the opt-in): %v", got)
	}
	if got := resolveRunner(false, false, "@tobilu/qmd@2.5.3"); got != nil {
		t.Fatalf("no npx means no runner: %v", got)
	}
}

func TestDecide(t *testing.T) {
	runner := []string{"qmd"}
	cases := []struct {
		name    string
		runner  []string
		file    string
		run     bool
		mention string
	}{
		{"no runner is a silent skip", nil, "report.md", false, "no qmd runner"},
		{"non-markdown never indexes", runner, "main.go", false, "not markdown"},
		{"markdown write triggers update", runner, "blue/report.md", true, "qmd update"},
		{"extension match is case-insensitive", runner, "NOTES.MD", true, "qmd update"},
		{"unknown file (no path in payload) skips", runner, "unknown", false, "not markdown"},
		{"npx runner is named in the log line", []string{"npx", "-y", "@tobilu/qmd@2.5.3"}, "a.md", true, "npx -y @tobilu/qmd@2.5.3 update"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, reason := decide(c.runner, c.file)
			if run != c.run {
				t.Fatalf("decide(%v,%q) run = %v; want %v", c.runner, c.file, run, c.run)
			}
			if !strings.Contains(reason, c.mention) {
				t.Fatalf("decide(%v,%q) reason = %q; missing %q", c.runner, c.file, reason, c.mention)
			}
			if !strings.Contains(reason, c.file) {
				t.Fatalf("reason must carry the file for the hook log: %q", reason)
			}
		})
	}
}

func TestFileFrom(t *testing.T) {
	var in hookInput
	if got := fileFrom(in); got != "unknown" {
		t.Fatalf("empty payload should yield unknown: %q", got)
	}
	in.ToolInput.Path = "p.md"
	if got := fileFrom(in); got != "p.md" {
		t.Fatalf("fileFrom should use path: %q", got)
	}
	in.ToolInput.FilePath = "f.md"
	if got := fileFrom(in); got != "f.md" {
		t.Fatalf("file_path should win over path: %q", got)
	}
}
