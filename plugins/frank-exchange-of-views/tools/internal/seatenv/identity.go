package seatenv

import (
	"os"
	"strconv"
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
// AND THE PART THAT IS NOT TRUE YET, which this file's presence makes easy to misread: nothing
// sets SeatVar or RoundVar anywhere in this repository. These constants have readers and no writer.
// Every round in production is still the regex over the seat id, phantom archive included (#396).
//
// WHAT THIS DELIBERATELY DOES NOT TOUCH. The seat id remains the SHARD KEY and the concurrency
// namespace. A lens index recovered from a seat name turned out to be exactly what made a
// lock-free counter safe under parallel dispatch, and collapsing it once made 39 of 60 disposals
// ambiguous. This moves what is READ; it never moves what IDENTIFIES.

// Identity variables, injected by the PreToolUse hook beside FEOV_RUN. Absent from every --help
// surface for the same reason FEOV_RUN is: a seat never types them, and documenting them would
// invite exactly the hand-typed path this removes.
const (
	SeatVar  = "FEOV_SEAT"
	RoundVar = "FEOV_ROUND"
)

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
func ResolveSeat(flagSeatID string, inferRound func(string) int) (Seat, error) {
	env := strings.TrimSpace(os.Getenv(SeatVar))
	if env != "" && flagSeatID != "" && env != flagSeatID {
		return Seat{}, feov.Errorf(feov.Conflict,
			"--seat-id %q disagrees with the seat this process was dispatched as (%q). "+
				"The engine injects your identity; you do not type it. If the flag is a typo, drop it — "+
				"omitting --seat-id is correct and always right. If you believe the injected value is "+
				"wrong, record it with the friction verb rather than working around it",
			flagSeatID, env)
	}
	id := env
	if id == "" {
		id = flagSeatID
	}
	if id == "" {
		return Seat{Round: -1}, nil
	}

	// An injected round is a FACT the dispatcher knows. Only when none arrives does the legacy
	// regex run, and its answer is then explicitly a guess about string shape.
	if raw := strings.TrimSpace(os.Getenv(RoundVar)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return Seat{}, feov.Errorf(feov.Validation,
				"%s=%q is not a number. The round is injected by the engine as a field; a malformed one is a "+
					"dispatch bug, and guessing from the seat id here is what this replaced", RoundVar, raw)
		}
		if n < 0 {
			return Seat{}, feov.Errorf(feov.Validation, "%s=%d is negative; rounds start at 0 (synthesis)", RoundVar, n)
		}
		return Seat{ID: id, Round: n}, nil
	}
	if inferRound == nil {
		return Seat{ID: id, Round: -1}, nil
	}
	return Seat{ID: id, Round: inferRound(id)}, nil
}
