// Package toolchain is the single source of truth for "is a required local tool
// present?". Both the hook binaries and the doctor command consult it, so
// knowledge of our local dependencies lives in exactly one place.
//
// See plans/claude-port-plan.md §3a' (Environment preflight).
package toolchain

import (
	"os"
	"os/exec"
	"strings"
)

// Environment names the kind of place this session runs in, so a manifest can
// declare a tool inapplicable there. Only the distinctions that change a verdict
// belong here; a token nothing keys off is noise.
const (
	// EnvRemote is a Claude Code cloud session (web, mobile, scheduled run).
	EnvRemote = "claude-code-remote"
	// EnvLocal is anything else — a developer's own machine.
	EnvLocal = "local"
)

// Environment resolves the current environment token. Claude Code sets
// CLAUDE_CODE_REMOTE in every cloud session and nowhere else.
func Environment() string {
	if os.Getenv("CLAUDE_CODE_REMOTE") != "" {
		return EnvRemote
	}
	return EnvLocal
}

// Tool is one declared dependency, mirroring an entry in requirements.json.
type Tool struct {
	Name     string            `json:"name"`
	Purpose  string            `json:"purpose"`
	Tier     string            `json:"tier"` // required | recommended | optional
	CheckCmd string            `json:"check_cmd"`
	Install  map[string]string `json:"install"` // keyed by GOOS: windows|darwin|linux
	// NotApplicableIn lists environments where this tool is absent BY DESIGN, so
	// its absence is not a defect and its install string is not advice. `gh` in a
	// cloud session is the motivating case: GitHub is reached through MCP tools
	// there, the CLI is deliberately not shipped, and telling the operator to
	// install it sends them after a tool the environment will not keep.
	//
	// This is deliberately NOT a tier override. "Less important here" and "not a
	// thing here" are different claims, and only the second one is true of gh.
	// A tool that is merely uninstalled — qlty, qmd in a fresh container — stays
	// a real DEGRADED, because installing it is a real remedy.
	NotApplicableIn []string `json:"not_applicable_in,omitempty"`
}

// AppliesIn reports whether this tool is expected to exist in the given environment.
func (t Tool) AppliesIn(env string) bool {
	for _, e := range t.NotApplicableIn {
		if e == env {
			return false
		}
	}
	return true
}

// Status is a tool plus its resolved presence.
type Status struct {
	Tool
	Found bool
	// NotApplicable means the manifest declared this tool out of scope for the
	// current environment. It is independent of Found: a not-applicable tool that
	// happens to be on PATH still reports Found, because the table must not claim
	// a tool is missing when it is sitting right there.
	NotApplicable bool
}

// Present reports whether the named executable resolves on PATH.
func Present(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Probe resolves presence for each tool, keying off the first token of CheckCmd
// (falling back to Name) so entries like "qlty --version" check the "qlty" binary.
func Probe(tools []Tool) []Status {
	return ProbeIn(tools, Environment())
}

// ProbeIn is Probe with the environment passed in, so a test's verdict does not
// depend on where the test happens to run.
func ProbeIn(tools []Tool, env string) []Status {
	out := make([]Status, 0, len(tools))
	for _, t := range tools {
		name := t.Name
		if f := firstField(t.CheckCmd); f != "" {
			name = f
		}
		out = append(out, Status{
			Tool:          t,
			Found:         Present(name),
			NotApplicable: !t.AppliesIn(env),
		})
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
