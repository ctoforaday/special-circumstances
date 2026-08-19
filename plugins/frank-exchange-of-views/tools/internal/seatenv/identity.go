package seatenv

import (
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// IDENTITY ARRIVES AS FIELDS, NOT AS A STRING TO PARSE (#348).
//
// THE MEASURED FAILURE. Every event's `Round` was computed by running a regex over the seat id
// at the append path:
//
//	Seq: len(events), TS: nextStamp(runDir), SeatID: seatID, Nonce: nonce, Round: RoundOf(seatID),
//
// That is the hottest recovered-from-a-string fact in the tool, and it has already failed:
// `judge-terminal` carries no round, so RoundOf returned 0, so a bench closure at run END looked
// like a closure BEFORE ROUND 1. It put a phantom entry in the archive and made the W1.8
// spot-check floor demand samples from rounds whose seats had done nothing wrong. It surfaced at
// 1 seed in 60 during #327, entirely by luck; a live run would have failed verify while naming
// innocent rounds.
//
// Role was recovered the same way — `strings.HasPrefix(e.SeatID, "red-merge")` — including in
// the branch deciding whether a position renders as RED or BLUE in the report.
//
// WHY IT IS FIXABLE NOW. Seat identity is deterministic: inherited from the agent id and name,
// unforgeable, and never typed by the LLM. So the harness can inject the STRUCTURED facts and
// the tool can stop inferring them.
//
// MEASURED 2026-08-15 (plans/hook-surface-spike.md §7, #290). That paragraph was a design premise
// when written; the harness half of it is now observed. PreToolUse carries agent_id on subagent
// tool calls — 9 of 9, across Bash, Read and Grep, from six agents; stable across an agent's calls;
// distinct across concurrent agents; equal to the handle the dispatcher already holds; and
// byte-identical to the id the same agent reports at SubagentStop, so the two hooks join on one
// key. session_id and prompt_id are IDENTICAL across the main session and every concurrent
// subagent, so agent_id is the ONLY field that discriminates one seat from another.
//
// AND THE PART THAT WAS NOT TRUE, which this file's presence made easy to misread: nothing set
// FEOV_SEAT or FEOV_ROUND anywhere in this repository. Both had readers and no writer, so both
// branches were scenery — and both are now gone. The seat is bound at `register` and read from the
// record; the round is derived from a seat id whose shape the roster gate validates, and it answers
// NOT-KNOWN rather than 0 where the name carries no round (#327/#396).
//
// WHAT THIS DELIBERATELY DOES NOT TOUCH. The seat id remains the SHARD KEY and the concurrency
// namespace. A lens index recovered from a seat name turned out to be exactly what made a
// lock-free counter safe under parallel dispatch, and collapsing it once made 39 of 60 disposals
// ambiguous. This moves what is READ; it never moves what IDENTIFIES.

// Identity variables, injected by the PreToolUse hook beside FEOV_RUN. Absent from every --help
// surface for the same reason FEOV_RUN is: a seat never types them, and documenting them would
// invite exactly the hand-typed path this removes.
const (
	// AgentVar carries the harness's own handle for the subagent, and it is the one identity
	// variable that HAS a writer: the PreToolUse hook reads agent_id off the payload and exports
	// it beside FEOV_RUN.
	//
	// It is deliberately NOT a seat id. Nothing upstream of `register` knows which agent holds
	// which seat — the dispatcher does not learn the id (Workflow's agent() returns a result, not
	// a handle), and recovering one from the command string would be a fact read back out of
	// prose. So the hook exports the fact it HAS, register writes the mapping as a field on the
	// record, and every later call resolves the seat from that. This is why SeatVar still has no
	// writer and is not going to get one.
	AgentVar = "FEOV_AGENT_ID"
)

// AgentID is the harness handle for this process's subagent, or "" in a main session or any
// context the hook did not rewrite.
func AgentID() string { return strings.TrimSpace(os.Getenv(AgentVar)) }

// Seat is a seat's identity as FACTS rather than as a string other code takes apart.
type Seat struct {
	ID string
	// Round is -1 when nothing supplied one. NOT 0: round 0 is synthesis, a real round in which
	// real events happen, and conflating "no round" with "the first round" is precisely the
	// phantom-archive bug. A caller that needs a number must decide what unknown means.
	Round int
}

// HasRound reports whether the round is known at all.
func (s Seat) HasRound() bool { return s.Round >= 0 }

// ResolveSeat returns the seat's identity for a verb.
//
// Order mirrors Resolve: injected → flag → inference. The DISAGREEMENT is again the point — a
// --seat-id that contradicts the injected identity is REFUSED rather than obeyed or silently
// overridden. Attribution is the one fact a seat must not be able to get wrong: an event filed
// under the wrong seat is credited to the wrong party, and every found_by, estoppel and parity
// check reads it.
//
// inferRound is the legacy path (RoundOf over the id), passed in so this package does not depend
// on record. It is used only when nothing was injected, which is every pre-#348 caller and the
// tests.
func ResolveSeat(flagSeatID string, bound func() (string, error), inferRound func(string) int) (Seat, error) {
	// THE BOUND SEAT IS THE ONE THIS AGENT REGISTERED AS, read from the record. It replaces
	// os.Getenv(SeatVar), which was never anything: FEOV_SEAT had readers and no writer, so the
	// "injected wins" branch below could not fire in any real run, and every seat was in fact
	// running on the flag it typed. The disagreement refusal existed and guarded nothing.
	//
	// A LOOKUP FAILURE IS NOT AN UNBOUND AGENT. bound distinguishes them and the error is
	// returned rather than folded into "", because a record nobody can read would otherwise
	// present as a seat that never registered — and the remedy for those two is opposite.
	var env string
	if bound != nil {
		b, err := bound()
		if err != nil {
			return Seat{}, err
		}
		env = b
	}
	if env != "" && flagSeatID != "" && env != flagSeatID {
		return Seat{}, feov.Errorf(feov.Conflict,
			"--seat-id %q disagrees with the seat this agent registered as (%q). "+
				"Your identity is bound at `register` and read back from the record; you do not retype it. "+
				"If the flag is a typo, drop it — omitting --seat-id is correct and always right once you "+
				"have registered. If you believe the bound value is wrong, record it with the friction verb "+
				"rather than working around it",
			flagSeatID, env)
	}
	id := env
	if id == "" {
		id = flagSeatID
	}
	if id == "" {
		return Seat{Round: -1}, nil
	}

	// THE ROUND IS DERIVED, AND THAT STOPPED BEING A GUESS. FEOV_ROUND used to be read here
	// first, as "a FACT the dispatcher knows" — and the dispatcher never knew it: nothing in
	// production ever set the variable, so the branch was scenery and the derivation below was
	// always the only path.
	//
	// It is now sound rather than merely sole. record.RoundOf reads the round out of a seat id
	// whose SHAPE the roster gate refuses at register unless it is one the engine produces, and
	// it answers NOT-KNOWN for the ids that genuinely carry no round instead of answering 0 —
	// which was the phantom-archive defect (#327/#396).
	if inferRound == nil {
		return Seat{ID: id, Round: -1}, nil
	}
	return Seat{ID: id, Round: inferRound(id)}, nil
}

// Dispatched is the seat id this process is running as, for callers that need it BEFORE flags are
// parsed — the CLI builds its command tree from the seat's role, and cobra resolves the command
// path before it parses any flag.
//
// The RECORD is the real channel: the hook injects an agent handle, `register` binds it, and bound
// resolves the pair. The os.Args scan is for the bootstrap call and for a human or test driving the
// binary by hand with --seat-id; it is a bounded read of two spellings and is not a second flag
// parser. Anything it cannot find yields "", which builds the operator tree — the honest answer for
// a process with no seat.
// THE FLAG WINS HERE, AND THE BINDING WINS AT EXECUTION. They answer different questions.
//
// Tree selection asks WHICH SURFACE AM I LOOKING AT; execution asks WHO IS ACTING. Letting the
// flag choose the tree is what lets any seat read any other seat's surface —
// `--seat-id red-merge-r1 --help` from a blue seat shows merge's verbs — without a second
// mechanism, because cobra answers --help without ever reaching RunE, so nothing is attributed
// and there is nothing to police. The moment a command actually RUNS, ResolveSeat applies the
// rule it always has: a --seat-id that disagrees with the injected identity is refused, because
// attribution is the one fact a seat must not be able to get wrong.
func Dispatched(bound func() (string, error)) string {
	if flag := SeatIDIn(os.Args); flag != "" {
		return flag
	}
	if bound == nil {
		return ""
	}
	// A LOOKUP FAILURE YIELDS NO TREE, not an error. This runs before flags are parsed, so there
	// is nowhere to report to and no command to attach a refusal to — and the honest answer for a
	// process whose seat cannot be established is the same as for one that has none: show the
	// surface that asks who you are. Execution re-resolves through ResolveSeat, which DOES
	// surface the error, so nothing is swallowed twice.
	seat, err := bound()
	if err != nil {
		return ""
	}
	return seat
}

// SeatIDIn reads a --seat-id out of an argument list. Exported for the CLI test harness, which
// drives cobra with an args slice that never reaches os.Args.
func SeatIDIn(args []string) string {
	for i, a := range args {
		if a == "--seat-id" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if v, ok := strings.CutPrefix(a, "--seat-id="); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
