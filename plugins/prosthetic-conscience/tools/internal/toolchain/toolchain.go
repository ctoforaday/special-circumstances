// Package toolchain is the single source of truth for "is a required local tool
// present?". Both the hook binaries and the doctor command consult it, so
// knowledge of our local dependencies lives in exactly one place.
//
// See plans/claude-port-plan.md §3a' (Environment preflight).
package toolchain

import (
	"os/exec"
	"strings"
)

// Tool is one declared dependency, mirroring an entry in requirements.json.
type Tool struct {
	Name     string            `json:"name"`
	Purpose  string            `json:"purpose"`
	Tier     string            `json:"tier"` // required | recommended | optional
	CheckCmd string            `json:"check_cmd"`
	Install  map[string]string `json:"install"` // keyed by GOOS: windows|darwin|linux
}

// Status is a tool plus its resolved presence.
type Status struct {
	Tool
	Found bool
}

// Present reports whether the named executable resolves on PATH.
func Present(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Probe resolves presence for each tool, keying off the first token of CheckCmd
// (falling back to Name) so entries like "qlty --version" check the "qlty" binary.
func Probe(tools []Tool) []Status {
	out := make([]Status, 0, len(tools))
	for _, t := range tools {
		name := t.Name
		if f := firstField(t.CheckCmd); f != "" {
			name = f
		}
		out = append(out, Status{Tool: t, Found: Present(name)})
	}
	return out
}

func firstField(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
