package main

import (
	"strings"
	"testing"
)

func TestDecide(t *testing.T) {
	cases := []struct {
		name    string
		qmd     bool
		file    string
		run     bool
		mention string
	}{
		{"absent qmd is a silent skip", false, "report.md", false, "qmd not found"},
		{"non-markdown never indexes", true, "main.go", false, "not markdown"},
		{"markdown write triggers update", true, "blue/report.md", true, "qmd update"},
		{"extension match is case-insensitive", true, "NOTES.MD", true, "qmd update"},
		{"unknown file (no path in payload) skips", true, "unknown", false, "not markdown"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, reason := decide(c.qmd, c.file)
			if run != c.run {
				t.Fatalf("decide(%v,%q) run = %v; want %v", c.qmd, c.file, run, c.run)
			}
			if !strings.Contains(reason, c.mention) {
				t.Fatalf("decide(%v,%q) reason = %q; missing %q", c.qmd, c.file, reason, c.mention)
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
