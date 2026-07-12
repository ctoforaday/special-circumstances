package main

import (
	"strings"
	"testing"
)

func TestGate(t *testing.T) {
	cases := []struct {
		name   string
		qlty   bool
		file   string
		expect []string
	}{
		{"absent warns and points to doctor", false, "a.go", []string{"skipped", "/sc-doctor", "a.go"}},
		{"present indicates it would run qlty", true, "b.md", []string{"qlty fmt", "qlty check", "b.md"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := gate(c.qlty, c.file)
			for _, want := range c.expect {
				if !strings.Contains(got, want) {
					t.Fatalf("gate(%v,%q) = %q; missing %q", c.qlty, c.file, got, want)
				}
			}
		})
	}
}

func TestFileFrom(t *testing.T) {
	var in hookInput
	in.ToolInput.Path = "p.txt"
	if got := fileFrom(in); got != "p.txt" {
		t.Fatalf("fileFrom should use path: %q", got)
	}
	in.ToolInput.FilePath = "f.txt"
	if got := fileFrom(in); got != "f.txt" {
		t.Fatalf("file_path should win over path: %q", got)
	}
}
