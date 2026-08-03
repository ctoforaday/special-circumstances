// Package hookgate is the decision core of the blue-report lockdown: the pure logic a
// Claude Code PreToolUse / PostToolUse hook runs to keep blue/report.md writable ONLY by
// the allowlisted author agent_type, and to catch a dropped finding-marker after the fact.
// It is deliberately free of stdin/CLI plumbing so the allow/deny rules are unit-tested
// directly (the CLI wrapper in internal/cli/hook.go handles the I/O).
package hookgate

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// AuthorAgentType is the ONE agent_type on the report.md write allowlist — the round-0
// synthesis author. Every other seat (blue-researcher responders, red, lead, anything) is
// default-DENIED a raw write to blue/report.md and must go through `feov-record blue edit`.
const AuthorAgentType = "frank-exchange-of-views:blue-synthesizer"

// denyReason is the message a denied seat sees — it names the sanctioned path.
const denyReason = "blue/report.md is read-only to this seat. Make every change through " +
	"`feov-record blue edit --old \"<exact current span>\" --new \"<replacement>\" --reason \"<why>\"` — " +
	"it preserves red's finding-markers and records the edit. Direct writes to the report are denied."

// PreDenyReason is the deny message, exported for the CLI's fail-closed path.
func PreDenyReason() string { return denyReason }

// Input is the subset of the hook stdin JSON the gate reads.
type Input struct {
	AgentType string          `json:"agent_type"`
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

type toolInput struct {
	FilePath string `json:"file_path"`
	Command  string `json:"command"`
}

// blueReportPath matches a path token ending in blue/report.md (forward OR back slashes),
// excluding shell metacharacters and quotes so it picks out a single path argument.
var blueReportPath = regexp.MustCompile(`[^\s'"|&;<>()]*[\\/]blue[\\/]report\.md`)

// bashWriteRes are the WRITE positions the PreToolUse Bash arm recognizes. This is the
// "common shell" layer — exotic writers (python -c, awk, heredocs, truncate via a
// variable) are OUT of scope here and caught by the PostToolUse backstop (PostDropped).
var bashWriteRes = []*regexp.Regexp{
	regexp.MustCompile(`>>?\s*` + blueReportPath.String()),                          // > file / >> file
	regexp.MustCompile(`\btee\b[^|&;]*\s` + blueReportPath.String()),                // tee ... file
	regexp.MustCompile(`\b(?:cp|mv|install)\b[^|&;]*\s` + blueReportPath.String()),  // cp/mv/install ... dest
	regexp.MustCompile(`\bsed\b[^|&;]*-i[^|&;]*\s` + blueReportPath.String()),       // sed -i ... file
	regexp.MustCompile(`\bdd\b[^|&;]*of=` + blueReportPath.String()),                // dd of=file
	regexp.MustCompile(`\b(?:truncate|ed|ex)\b[^|&;]*\s` + blueReportPath.String()), // truncate/ed/ex file
}

// isBlueReport reports whether a resolved file path targets a run's blue/report.md.
func isBlueReport(p string) bool {
	p = filepath.ToSlash(p)
	return strings.HasSuffix(p, "/blue/report.md") || p == "blue/report.md"
}

// writesBlueReport resolves whether a tool call WRITES a run's blue/report.md. For the
// structured edit tools the target is tool_input.file_path (airtight); for Bash it is a
// write-position match on the command (best-effort common-shell layer).
func writesBlueReport(in Input) bool {
	var ti toolInput
	_ = json.Unmarshal(in.ToolInput, &ti)
	switch in.ToolName {
	case "Write", "Edit", "MultiEdit", "NotebookEdit":
		return isBlueReport(ti.FilePath)
	case "Bash":
		for _, re := range bashWriteRes {
			if re.MatchString(ti.Command) {
				return true
			}
		}
	}
	return false
}

// PreDecision decides a PreToolUse call: (deny, reason). It denies a write to a run's
// blue/report.md by any agent_type other than the author; a non-report target or an author
// write returns (false, ""). Default-DENY by allowlist: only the author is permitted.
func PreDecision(in Input) (bool, string) {
	if !writesBlueReport(in) {
		return false, "" // not a report.md write → no opinion
	}
	if in.AgentType == AuthorAgentType {
		return false, "" // the allowlisted author may write directly
	}
	return true, denyReason
}

// reportPathFrom extracts a blue/report.md path token referenced by the tool call (a
// file_path or anywhere in a Bash command), or "" if none. The PostToolUse backstop only
// has something to check when the call actually touched a report.md path.
func reportPathFrom(in Input) string {
	var ti toolInput
	_ = json.Unmarshal(in.ToolInput, &ti)
	if isBlueReport(ti.FilePath) {
		return ti.FilePath
	}
	if m := blueReportPath.FindString(ti.Command); m != "" {
		return m
	}
	return ""
}

// PostDropped is the PostToolUse backstop: after a NON-author tool call that referenced a
// run's blue/report.md, it returns the immortal-anchor ids of EITHER class — a finding
// marker (f-…) or a citation anchor (c-…) — that are now MISSING (dropped), via ANY write
// mechanism, including the exotic Bash the PreToolUse arm cannot enumerate. The author is
// never checked; a call that did not reference report.md returns nil. It is presence-only (a
// moved-but-present anchor is red's semantic-audit job, not this check). readReport reads the
// report bytes; anchorIDs supplies the EXPECTED set — the union of finding anchors and
// citation labels (both injected so the logic stays pure and testable).
func PostDropped(in Input, anchorIDs func(runDir string) ([]string, error), readReport func(path string) (string, error)) ([]string, error) {
	if in.AgentType == AuthorAgentType {
		return nil, nil
	}
	path := reportPathFrom(in)
	if path == "" {
		return nil, nil // the call did not touch a report.md → nothing to check
	}
	runDir := filepath.Dir(filepath.Dir(path)) // strip /blue/report.md
	expected, err := anchorIDs(runDir)
	if err != nil {
		return nil, err
	}
	if len(expected) == 0 {
		return nil, nil
	}
	md, err := readReport(path)
	if err != nil {
		// The report should exist if markers were anchored; a read failure here is a real
		// integrity signal, but surfacing it as an error lets the CLI fail-closed.
		return nil, err
	}
	return claimcount.MissingProtectedAnchorIDs(expected, md), nil
}

// DefaultAnchorIDs is the production EXPECTED set for PostDropped: the UNION of finding
// anchors (from `anchor` events) and citation labels (from blue `cite` events), so the
// backstop catches a dropped anchor of either class. Finding ids (f-…) and citation ids
// (c-…) never collide, so the concatenation is a clean union.
func DefaultAnchorIDs(runDir string) ([]string, error) {
	findings, err := record.AnchorIDs(runDir)
	if err != nil {
		return nil, err
	}
	cites, err := record.CitationLabels(runDir)
	if err != nil {
		return nil, err
	}
	return append(findings, cites...), nil
}
