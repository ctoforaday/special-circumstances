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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
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
	AgentType string `json:"agent_type"`
	// AgentID is the harness's handle for the subagent making this call, and it is the ONLY
	// field on the payload that discriminates one seat from another: session_id and prompt_id
	// are byte-identical across the main session and every concurrent subagent
	// (plans/hook-surface-spike.md §7). Measured present on 9 of 9 subagent calls across three
	// tool types, stable within an agent, distinct across concurrent ones.
	//
	// It is ABSENT on a main-session call, and that absence is meaningful rather than missing
	// data: the main session is not a seat. An empty AgentID injects nothing.
	AgentID   string          `json:"agent_id"`
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

// isBlueReport reports whether a resolved file path targets a blue/report.md BY SHAPE.
//
// Shape alone is not the question the lockdown asks — see insideRun, which is the half that was
// documented from the first commit and implemented in none of them.
func isBlueReport(p string) bool {
	p = filepath.ToSlash(p)
	return strings.HasSuffix(p, "/blue/report.md") || p == "blue/report.md"
}

// insideRun reports whether a path lies within the live run directory.
//
// THIS IS THE CHECK THAT WAS ALWAYS DESCRIBED AND NEVER WRITTEN. Every doc comment on this gate
// said "a run's blue/report.md"; the code matched a path SUFFIX and consulted no run at all.
// PreOutcome has held a resolved runDir since the first commit in this file's history and passed
// it only to the injection arm — the lockdown decision sat one line above it and never took it.
//
// What that cost: any file named blue/report.md anywhere on disk was locked, in any session, run
// or no run — a finished run under research/, a scratch copy, a fixture. The gate was meant to be
// a no-op outside a live run and never was.
//
// An empty runDir means no live run was resolvable, and the answer there is no opinion rather
// than a guess: matching InferRunDir's "say nothing rather than guess".
func insideRun(path, runDir string) bool {
	if runDir == "" || path == "" {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	root, err := filepath.Abs(runDir)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	// ".." leading means the path climbed OUT of the run, which is not inside it.
	return rel != ".." && !strings.HasPrefix(rel, "../")
}

// writesBlueReport resolves whether a tool call WRITES a run's blue/report.md. For the
// structured edit tools the target is tool_input.file_path (airtight); for Bash it is a
// write-position match on the command (best-effort common-shell layer).
// InputUnreadable reports whether a call's tool_input failed to parse. The CLI wrapper writes a
// FRICTION_KIND_TOOL_ERROR event when it is true, so a gate that could not see its target leaves
// a trace instead of a silence. Separate from writesBlueReport because a pure predicate should
// answer one question.
func InputUnreadable(in Input) bool {
	var ti toolInput
	return json.Unmarshal(in.ToolInput, &ti) != nil
}

func writesBlueReport(in Input, runDir string) bool {
	var ti toolInput
	// A GATE THAT CANNOT READ ITS INPUT MUST LEAVE A TRACE. Swallowing the parse failure leaves
	// ti.FilePath empty, makes this return false for the structured-tool arm, and makes Decide
	// answer "no opinion" — so malformed input bypasses the lockdown with nothing recorded.
	//
	// This package is deliberately free of I/O, so it REPORTS rather than acts: InputUnreadable
	// below is the predicate, and internal/cli/hook.go writes the FRICTION_KIND_TOOL_ERROR event
	// and emits an `ask` decision carrying the cause back to the model. The gate's fail direction
	// is unchanged — that is a security-semantics decision, not a cleanup.
	_ = json.Unmarshal(in.ToolInput, &ti) // reported via InputUnreadable; see above
	switch in.ToolName {
	case "Write", "Edit", "MultiEdit", "NotebookEdit":
		return isBlueReport(ti.FilePath) && insideRun(ti.FilePath, runDir)
	case "Bash":
		// The Bash arm recovers the path from the command, so the run test is applied to what it
		// recovered rather than to the command string. A write position that names no path inside
		// the run is not this gate's business, for the same reason a Write to one is not.
		for _, re := range bashWriteRes {
			if m := re.FindString(ti.Command); m != "" {
				if p := blueReportPath.FindString(m); p != "" && insideRun(p, runDir) {
					return true
				}
			}
		}
	}
	return false
}

// PreDecision decides a PreToolUse call: (deny, reason). It denies a write to a run's
// blue/report.md by any agent_type other than the author; a non-report target or an author
// write returns (false, ""). Default-DENY by allowlist: only the author is permitted.
func PreDecision(in Input, runDir string) (bool, string) {
	// OUTSIDE A LIVE RUN THIS GATE HAS NO OPINION, WHICH IS WHAT IT ALWAYS SAID IT DID. The
	// lockdown protects a run's working report; with no run resolvable there is no report to
	// protect, and every path is somebody else's file. This guard comes FIRST — ahead of the
	// fail-closed readability check below — because an unreadable payload outside a run cannot be
	// a write to a run's report either, and denying it was how this gate came to refuse writes to
	// arbitrary files in sessions that had nothing to do with a debate.
	if runDir == "" {
		return false, ""
	}
	// FAIL CLOSED ON INPUT THIS GATE CANNOT READ. A gate bypassed by input it cannot parse is not
	// a gate: an unreadable tool_input leaves FilePath empty, which reads as "not the report" and
	// lets the write through — the one outcome an allowlist must never produce by accident.
	//
	// THE COST IS REAL AND IS ACCEPTED DELIBERATELY (operator, 2026-08-18). If the harness ever
	// sends a shape this parser does not expect, a legitimate write is refused — and the seat most
	// likely to be hit is the ALLOWLISTED AUTHOR, the one seat that should never be blocked. That
	// is the trade against a silent bypass, and it is survivable only because the refusal says
	// what happened: the reason reaches the model, and the caller files a friction event, so a
	// denial is diagnosable rather than mysterious.
	if InputUnreadable(in) {
		return true, unreadableReason
	}
	if !writesBlueReport(in, runDir) {
		return false, "" // not a write to THIS run's report.md → no opinion
	}
	if in.AgentType == AuthorAgentType {
		return false, "" // the allowlisted author may write directly
	}
	return true, denyReason
}

// unreadableReason reaches the MODEL through permissionDecisionReason, so it names the cause and
// the remedy rather than only refusing.
const unreadableReason = "feov-record: this hook could not parse tool_input, so it cannot tell " +
	"whether the call writes blue/report.md — and the blue-report lockdown must not be bypassed " +
	"by input it cannot read. Nothing about your call is known to be wrong. Retry it; if it " +
	"refuses again, the tool_input shape has changed and the gate needs updating — a friction " +
	"event has been filed recording exactly that."

// Outcome is what the PreToolUse hook should do with a tool call.
type Outcome int

const (
	// OutcomeNone: no opinion. The call proceeds untouched.
	OutcomeNone Outcome = iota
	// OutcomeDeny: refuse the call; the string is the reason shown to the seat.
	OutcomeDeny
	// OutcomeRewrite: let the call proceed with a REPLACED command; the string is that command.
	OutcomeRewrite
)

// feovToken matches `feov-record` in COMMAND POSITION — at the start of the command, or
// immediately after a separator (`&&`, `||`, `;`, `|`, or a newline), optionally quoted and
// optionally path-prefixed.
//
// POSITION, NOT SUBSTRING, and the difference is the whole safety argument. A mention is not
// an invocation: `grep -rn "feov-record" plugins/`, a heredoc documenting a verb, a `--reason`
// quoting a command that failed — all contain the token, and a `strings.Contains` rule would
// have prefixed an `export` onto documentation writes and friction messages. The two are
// indistinguishable to any matcher that does not look at where the token sits.
// COMMAND SUBSTITUTION AND SUBSHELLS ARE COMMAND POSITION TOO, and leaving them out cost
// judge-r2 its identity in the same run the heredoc bail cost blue-respond-r1 its
// bibliography: `evidence_cites=$("…/feov-record" show evidence | grep …)` matched nothing,
// so nothing was injected and the tool told a registered seat it had never registered.
//
// `$(`, a backtick and a bare `(` all open a context where the next word is a command. The
// asymmetry that settles the trade: a false positive prepends a harmless export to a command
// that does not use the tool; a false negative destroys a seat's identity and its work.
var feovToken = regexp.MustCompile("(?:^|&&|\\|\\||;|\\||\\n|\\$\\(|`|\\()\\s*['\"]?[^\\s'\";|&]*\\bfeov-record\\b")

// PreOutcome is the SINGLE entry point for the PreToolUse decision, and the ordering is
// structural rather than a convention someone has to remember.
//
// DENY WINS, and it cannot be got around: the deny arm is consulted first and the function
// physically cannot return OutcomeRewrite once it fires. A previous revision left the rewrite
// an exported free function with the ordering living in prose plus one call site — and the
// consequence of getting that wrong is not a cosmetic bug, it is the blue-report lockdown
// silently opening, because the hook protocol permits only ONE hookSpecificOutput document
// and a rewrite emitted in place of a deny is a deny that never happened. A command tripping
// both arms (`cp draft.md …/blue/report.md && "…/feov-record" blue edit …`) is denied.
//
// runDir is a PARAMETER, never a field on Input. Input is unmarshalled straight from the hook
// payload, so every field on it is wire-supplied; a CLI-computed member would leave a reader
// unable to tell a derived value from something the client sent. PostDropped takes its
// injected readers for the same reason.
//
// PreDecision stays exported for its existing callers and cannot emit a rewrite.
func PreOutcome(in Input, runDir string) (Outcome, string) {
	if deny, reason := PreDecision(in, runDir); deny {
		return OutcomeDeny, reason
	}
	if runDir == "" || in.ToolName != "Bash" {
		return OutcomeNone, ""
	}
	var ti toolInput
	if json.Unmarshal(in.ToolInput, &ti) != nil || ti.Command == "" {
		return OutcomeNone, ""
	}
	rewritten, ok := injectEnv(ti.Command, [][2]string{
		{seatenv.Var, runDir},
		// THE IDENTITY, WHICH IS THE HALF THAT WAS MISSING. FEOV_SEAT had readers and no
		// writer, because nothing in the system could produce a seat id: only `register`
		// knows which agent holds which seat, and that is downstream of the first call. What
		// the harness CAN hand over is agent_id, so that is what travels — and the record
		// resolves it to a seat, where the mapping is a field somebody wrote rather than a
		// value recovered from a command string.
		{seatenv.AgentVar, in.AgentID},
	})
	if !ok {
		return OutcomeNone, ""
	}
	return OutcomeRewrite, rewritten
}

// injectEnv prefixes `export VAR='value'; ` for each variable a command does not already
// carry, to a command that invokes feov-record.
//
// `export …;` and NOT the inline `VAR=x cmd` form. Measured: blue's real command was
// `cd C:/… && "…/feov-record" blue manifest-row …`, where an inline prefix binds to `cd` and
// never crosses the `&&`. An export is a statement; it applies to everything after it.
//
// QUOTING IS A SECURITY BOUNDARY, not formatting. Each value is single-quoted with `'` escaped
// as `'\”`, and a value carrying a control character is REFUSED — that variable is dropped —
// because a newline inside the emitted prefix would end the export statement and turn the
// remainder into a command. The run directory comes from a file on disk, and the agent id comes
// off the wire; neither is a trusted input just because we read it ourselves.
//
// PER-VARIABLE, NOT ALL-OR-NOTHING. A command already carrying FEOV_RUN still gets FEOV_AGENT_ID,
// which is what a seat that copied an earlier rewritten command produces. Making the whole
// injection idempotent on one variable would have silently skipped the other — the plausible
// zero, one layer down: identity absent, and the absence looking exactly like a main-session call.
func injectEnv(command string, vars [][2]string) (string, bool) {
	// HEREDOCS ARE A BLIND SPOT, so the whole command is left alone when one is present.
	//
	// A newline counts as a command separator — multi-line scripts are ordinary — but inside a
	// heredoc body a newline introduces DATA, not a command, and telling the two apart needs a
	// shell parser rather than a matcher. Caught by the test that writes a heredoc documenting
	// a verb: it was rewritten, which would have injected an export into the middle of a
	// document a seat was writing. Bailing costs only the injection (the seat still has
	// --run); guessing corrupts the seat's output.
	// A heredoc BODY is data, so it is removed before the match — but the command is no longer
	// abandoned because it contains one.
	//
	// MEASURED 2026-08-22, and it cost a run its bibliography. The bail here read
	// `strings.Contains(command, "<<") -> return "", false`, justified as: "Bailing costs only
	// the injection (the seat still has --run)." That was true when FEOV_RUN was the only
	// passenger. FEOV_AGENT_ID joined this same channel eleven days later and the sentence was
	// never revisited — so the cost quietly became THE SEAT'S IDENTITY.
	//
	// What happened: blue-respond-r1 ran `position --reason-file - <<'EOF'`, the hook bailed, no
	// FEOV_AGENT_ID reached the tool, and the tool answered "this agent has not registered" —
	// FALSE, it had registered fifty calls earlier. The seat believed the refusal and
	// re-registered, which rotates the shard nonce; replay winner selection then kept the newer,
	// thinner shard and orphaned 26 events: 10 citations, 4 report edits, 7 manifest rows. The
	// report still carried the anchors, so its bibliography renders ten
	// "(unresolved citation c-… — no source on the record)" lines. 7 of 8 such refusals that run
	// were heredoc commands, across three seats.
	//
	// AND THE OLD JUSTIFICATION MISDESCRIBED ITS OWN MECHANISM: "it was rewritten, which would
	// have injected an export into the middle of a document a seat was writing". Injection
	// PREPENDS — `prefix + command`, below — so a heredoc body is byte-identical either way. The
	// document could not have been touched.
	//
	// THE RISK PROFILE IS WHY THIS IS SAFE. Stripping heredocs by matcher rather than by shell
	// parser can be wrong, and being wrong costs at most a decision about whether to prepend a
	// harmless export. It can never corrupt output, because nothing is ever inserted INTO the
	// command. Against that: a missed injection destroys a seat's identity and its work. The
	// matcher should lean toward injecting, and now does.
	if !feovToken.MatchString(stripHeredocBodies(command)) {
		return "", false
	}
	var prefix strings.Builder
	for _, kv := range vars {
		name, value := kv[0], kv[1]
		// An empty value is NOT injected as an empty string. `export FEOV_AGENT_ID='';` would
		// make "the main session, which has no agent id" indistinguishable from "an agent whose
		// id is the empty string" at every reader downstream.
		if value == "" {
			continue
		}
		// Idempotent per variable, so a double hook invocation (or a seat that copied a
		// rewritten command) cannot stack prefixes.
		if strings.Contains(command, "export "+name+"=") {
			continue
		}
		if hasControl(value) {
			continue
		}
		prefix.WriteString("export " + name + "='" + strings.ReplaceAll(value, `'`, `'\''`) + "'; ")
	}
	if prefix.Len() == 0 {
		return "", false
	}
	return prefix.String() + command, true
}

// hasControl reports whether a value carries a byte that would break out of the single-quoted
// export statement it is about to be spliced into.
func hasControl(v string) bool {
	for _, r := range v {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
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

// heredocStart matches a heredoc redirection and captures its delimiter: `<<EOF`, `<<-EOF`,
// `<<'EOF'`, `<<"EOF"`. The delimiter ends the body when it appears alone on a line (indented
// too, for the `<<-` form).
var heredocStart = regexp.MustCompile(`<<-?\s*(['"]?)([A-Za-z_][A-Za-z0-9_]*)(['"]?)`)

// stripHeredocBodies blanks the BODY of every heredoc in a command, leaving the surrounding
// shell intact, so a command-position match cannot be satisfied by a line of DATA.
//
// A heredoc body is the one place where a leading newline introduces text rather than a command,
// which is the whole reason feovToken cannot be trusted on a raw command containing one: a
// document whose line begins `feov-record blue cite …` looks exactly like an invocation.
//
// THIS IS A MATCHER, NOT A SHELL PARSER, and it is allowed to be one because of where it sits.
// Its answer decides only whether a harmless `export` is PREPENDED to the command; it never
// edits the command's interior. A mis-parse therefore costs an unnecessary export or a missed
// one — never a corrupted document. Nesting, delimiters carrying quotes, and parameter expansion
// in the delimiter are out of scope and fail toward treating the text as a body.
func stripHeredocBodies(command string) string {
	lines := strings.Split(command, "\n")
	var out []string
	i := 0
	for i < len(lines) {
		line := lines[i]
		out = append(out, line)
		m := heredocStart.FindStringSubmatch(line)
		i++
		if m == nil {
			continue
		}
		// An unterminated heredoc swallows the rest, which is what the shell would do too.
		delim := m[2]
		for i < len(lines) && strings.TrimSpace(lines[i]) != delim {
			out = append(out, "") // the body, blanked — length preserved for nothing, clarity for readers
			i++
		}
		if i < len(lines) {
			out = append(out, lines[i]) // the terminator line itself is shell, not data
			i++
		}
	}
	return strings.Join(out, "\n")
}
