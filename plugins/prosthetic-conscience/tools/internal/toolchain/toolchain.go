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
	// A tool that is merely uninstalled — qlty in a fresh container — stays
	// a real DEGRADED, because installing it is a real remedy.
	NotApplicableIn []string `json:"not_applicable_in,omitempty"`
	// MinVersion is the lowest version of this tool that can do the job, as dotted numbers
	// ("1.25"). Declaring it is what turns CheckCmd from a label into a check — see version.go.
	// Absent means presence is the whole question, which is true of most tools and is why this
	// is opt-in rather than required.
	MinVersion string `json:"min_version,omitempty"`
	// CheckDir is where CheckCmd must run, relative to the manifest, when the answer depends on
	// it. MEASURED, on the first real run of this check (#521): `go version` reports go1.24.7 at
	// this repository's root and go1.25.0 inside `tools/`, because GOTOOLCHAIN=auto selects a
	// toolchain PER MODULE from that module's own `go` directive. Asked from the wrong place the
	// check declared BLOCKED on a machine where every gate passes — a real number from a
	// location that made it meaningless. Empty means the answer does not depend on where you
	// stand, which is true of every other tool declared here.
	CheckDir string `json:"check_dir,omitempty"`
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

	// VersionLine is the check command's first output line, kept so the table can render what
	// the tool actually said rather than re-running the command to ask again.
	VersionLine string
	// Version is the number parsed out of it, "" when no minimum was declared or none was found.
	Version string
	// TooOld means a minimum was declared, a version was read, and it is below the line. It is
	// its own state and not a flavour of missing: the remedy is upgrade, not install.
	TooOld bool
	// VersionUnmeasured carries WHY a declared minimum could not be checked. Non-empty is never
	// a pass — the alternative is a tool whose version could not be read reporting ✓, which is
	// the same bytes as a tool that met the bar.
	VersionUnmeasured string
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
func ProbeIn(tools []Tool, env string) []Status { return ProbeInDir(tools, env, "") }

// ProbeInDir is ProbeIn with the manifest's own directory, which a tool declaring CheckDir
// needs in order to be asked its question in the right place.
func ProbeInDir(tools []Tool, env, baseDir string) []Status {
	out := make([]Status, 0, len(tools))
	for _, t := range tools {
		name := t.Name
		if f := firstField(t.CheckCmd); f != "" {
			name = f
		}
		st := Status{
			Tool:          t,
			Found:         Present(name),
			NotApplicable: !t.AppliesIn(env),
		}
		// A tool that does not apply here is not interrogated about its version: the manifest
		// has already said its absence is not a defect, so a minimum it cannot meet is not one
		// either.
		if !st.NotApplicable {
			measure(&st, baseDir)
		}
		out = append(out, st)
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
