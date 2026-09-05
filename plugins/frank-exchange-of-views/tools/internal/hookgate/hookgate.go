// Package hookgate is the decision core of the PreToolUse run/identity injection: the pure logic a
// Claude Code PreToolUse hook runs to prefix every Bash command in a live run with the run directory
// and the calling agent's id, so a seat never mistypes the path and never loses the identity that
// binds it to its seat on the record. It is deliberately free of stdin/CLI plumbing so the rewrite
// rule is unit-tested directly (the CLI wrapper in internal/cli/hook.go handles the I/O).
//
// It USED to also be the blue-report write-lockdown — a PreToolUse deny of a non-author raw write to
// blue/report.md and a PostToolUse backstop for a dropped finding-marker. Under report-as-record
// (#709) there is no blue/report.md file: the report is the record, every change is an appended
// event, and a raw write cannot reach it. The lockdown protected a file that no longer exists, so it
// is gone; only the injection — orthogonal to it from the start — remains.
package hookgate

import (
	"encoding/json"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// Input is the subset of the hook stdin JSON the gate reads.
type Input struct {
	// AgentID is the harness's handle for the subagent making this call, and it is the ONLY
	// field on the payload that discriminates one seat from another: session_id and prompt_id
	// are byte-identical across the main session and every concurrent subagent
	// (plans/hook-surface-spike.md §7). Measured present on 9 of 9 subagent calls across three
	// tool types, stable within an agent, distinct across concurrent ones.
	//
	// It is ABSENT on a main-session call, and that absence is meaningful rather than missing
	// data: the main session is not a seat. An empty AgentID injects nothing.
	AgentID string `json:"agent_id"`

	// AgentType is the harness's name for the agent CONFIGURATION the caller runs under —
	// `frank-exchange-of-views:red-auditor` and its siblings. A DIFFERENT FACT FROM AgentID and
	// it refuses a different thing: the handle says WHICH agent, this says which configuration
	// it was dispatched as, so `register` can reject a seat id from the wrong family.
	//
	// THIS FIELD USED TO BE HERE FOR THE BLUE-REPORT LOCKDOWN, which allowlisted the author's
	// agent_type on every write. #709 made the report a record rather than a file, the lockdown
	// went with it, and the field went with the lockdown. It returns for the identity work,
	// which is now its only reader — so unlike before, nothing else keeps it parsed and the
	// injection test is what holds it.
	//
	// ABSENT on a main-session call, exactly like AgentID, and absent on a SubagentStop fired at
	// the main agent's turn end — which is the property that tells a seat from a turn boundary
	// (plans/hook-surface-spike.md §7a: 19 seats and 50 turn ends, zero exceptions either way).
	AgentType string `json:"agent_type"`

	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

type toolInput struct {
	Command string `json:"command"`
}

// Outcome is what the PreToolUse hook should do with a tool call.
type Outcome int

const (
	// OutcomeNone: no opinion. The call proceeds untouched.
	OutcomeNone Outcome = iota
	// OutcomeRewrite: let the call proceed with a REPLACED command; the string is that command.
	OutcomeRewrite
)

// THERE IS NO MATCHER, AND THAT IS THE POINT.
//
// Injection used to be gated on recognising `feov-record` in command position. That gate was a
// pattern standing in for a schema — "does this command invoke our tool" recovered from string
// shape — and its miss was silent: no injection, no error, and a seat's identity simply absent,
// which is byte-identical to a main-session call. The shape was widened three times, each time
// after a run had already paid: the heredoc bail cost blue-respond-r1 its bibliography,
// command substitution cost judge-r2 its identity, and `RB="…/feov-record"; $RB …` — a seat
// aliasing the long path, which the matcher could never see — cost 21 of 65 registers across
// six runs their agent_id (#510).
//
// THE HARM THE MATCHER GUARDED AGAINST CANNOT OCCUR. Its stated justification was that a
// `strings.Contains` rule "would have prefixed an export onto documentation writes and friction
// messages". Injection PREPENDS — `prefix + command` below — so `echo "run feov-record next"
// >> notes.md` writes the same bytes either way. This file already caught that misdescription
// once, in the heredoc comment, and fixed only the half in front of it.
//
// So the decision is now: a live run directory and a Bash call, and every such command carries
// the run and the identity. A needless `export` on a command that does not use the tool is
// inert. A missing one destroys a seat's work.

// PreOutcome is the SINGLE entry point for the PreToolUse decision: inject the run directory and
// the calling agent's id into every Bash command in a live run, or say nothing.
//
// runDir is a PARAMETER, never a field on Input. Input is unmarshalled straight from the hook
// payload, so every field on it is wire-supplied; a CLI-computed member would leave a reader
// unable to tell a derived value from something the client sent.
func PreOutcome(in Input, runDir string) (Outcome, string) {
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
		// AND THE AGENT'S TYPE, which is a different fact from its handle and refuses a
		// different thing. agent_id says WHICH agent; agent_type says which CONFIGURATION it
		// runs under, so `register` can refuse a seat id from the wrong family — a lead-judge
		// agent claiming `red-merge-r1`. This field has been read here since 0.27.0 for the
		// report lockdown; exporting it changes nothing about how it arrives.
		{seatenv.TypeVar, in.AgentType},
	})
	if !ok {
		return OutcomeNone, ""
	}
	return OutcomeRewrite, rewritten
}

// injectEnv prefixes `export VAR='value'; ` for each variable a command does not already carry.
// EVERY Bash command in a live run, not only the ones that look like they invoke the tool — see
// the note above PreOutcome for why there is no longer anything to look at.
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
	// NOTHING IS EVER INSERTED INTO THE COMMAND. The prefix is prepended, whole, in front of
	// whatever the seat wrote — heredoc bodies, quoted prose, documentation being written to a
	// file, all byte-identical afterwards. That is the property that lets this run unconditionally
	// instead of guessing which commands "really" invoke the tool.
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
