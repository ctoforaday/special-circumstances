package seat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY ACT SAYS WHERE THE SEAT NOW STANDS (#1122).
//
// #1163 delivered the work list at dispatch, which ended the OPENING read and made a sitting with
// nothing to do cost nothing. It could not end the rest: SubagentStart fires once per dispatch, and
// a seat's own acts are what move its list, so a working sitting still had to ask after every act.
//
// THE ANSWER RIDES THE ACT, because the act already ran in a process holding the record. A
// PostToolUse hook would work as a channel — it fires on every successful tool call and injects with
// no loop (plans/hook-surface-spike.md §8) — and would be the wrong instrument: it would re-read the
// record to learn what the process that just wrote it already knew.
//
// # It carries the VERDICT, not the list
//
// Measured on universe-m9's record, a full work list is 4.0-5.0 KB per seat — roughly 1.2k tokens.
// Attached to every write, a ten-act sitting pays ~12k tokens, about an entire lens prompt, to
// re-send a list the seat mostly already has. A seat asks two or three times, not ten, so sending
// the whole list would save round-trips and SPEND more tokens than the asking it replaces.
//
// What a seat needs after acting is whether it may stop, and if not, what is holding it. That is
// `complete`, the counts, and the BLOCKING items — one to three short strings. The non-blocking
// affordances are the long tail and arrived at dispatch.
//
// # The discriminator is the ROUTING, not a declaration
//
// A read is rendered by renderView and returns BEFORE Emit, so every invocation reaching Emit's
// success path has acted; a refusal takes the error path. There is therefore nothing here to test a
// verb against.
//
// A guard on the verb's declared record type sat here first and was removed as dead: every verb that
// reaches Emit carries that annotation — including the motion verbs, which are built from a bare
// cobra.Command and annotate themselves afterwards (cli/motion/verbs.go) — so the guard excluded
// nothing, and no mutation of it could fail a test. A check that cannot be wrong is not protection,
// it is a claim that the routing above is not trusted. What DOES need excluding is an identity that
// holds no position in the debate, which is the operator, checked below.

// StandingJSON is where a seat stands after the act it just recorded.
//
// IT IS NOT A SECOND COMPUTATION OF THE WORK LIST. The fields come from record.SittingOf through the
// same call `show work` makes, so a seat that reads this and a seat that asks cannot be told
// different things — which is the whole reason the list is not re-derived here.
type StandingJSON struct {
	// Complete answers "may I end my turn": false whenever a blocking item is open. It is the
	// question a seat was calling `show work` to settle.
	Complete bool `json:"complete"`
	// Blocking is every item that stops closure, verbatim from the work list. Not a count — a count
	// tells a seat it is held and not by what, which is another read.
	Blocking []string `json:"blocking"`
	// Open is how many items the list holds in total, blocking or not. It is the pointer to the rest:
	// a seat with nothing blocking and items open is complete and not finished, and this is the only
	// thing here that says so.
	Open int `json:"open"`
	// Unreadable is why the standing could not be computed, and it exists so this block's absence is
	// never mistaken for "nothing is open to you". A zero Complete with no reason would read as a
	// seat that is blocked; a true one would read as a seat that may stop. Neither is a measurement.
	Unreadable string `json:"unreadable,omitempty"`
}

// Human renders the standing as the lines that follow a success message.
func (s StandingJSON) Human() string {
	if s.Unreadable != "" {
		return "WHERE YOU STAND IS NOT MEASURED: " + s.Unreadable +
			"\n  this is not a report that nothing is open to you — read your work list before you stop"
	}
	var b strings.Builder
	if len(s.Blocking) == 0 {
		fmt.Fprintf(&b, "you may stop: nothing blocking, %d item(s) open to you", s.Open)
		return b.String()
	}
	fmt.Fprintf(&b, "you may NOT stop yet — %d of %d item(s) block you:", len(s.Blocking), s.Open)
	for _, it := range s.Blocking {
		b.WriteString("\n  · " + it)
	}
	return b.String()
}

// standingAfter computes where the seat stands, or says why it cannot.
//
// A nil return means this invocation is not one that can have moved the seat's position — a read, or
// an operator verb with no seat work at all — and the caller attaches nothing. That is different from
// a StandingJSON carrying Unreadable, which is an act whose standing could not be computed.
func standingAfter(cmd *cobra.Command) *StandingJSON {
	role := roleOf(cmd)
	seatID := Of(cmd).SeatID
	// THE OPERATOR HOLDS NO POSITION IN THE DEBATE, and that is a fact about the run's structure
	// rather than an empty list — the same distinction record.ScorecardOf draws for the card.
	// Attaching a standing here would invent a position for an identity that has none, and the
	// operator's verbs (setup, capture) are not a seat's acts.
	if seatID == "" || role == record.OperatorRole {
		return nil
	}
	run, err := Of(cmd).RequireRun(role)
	if err != nil {
		return &StandingJSON{Unreadable: err.Error()}
	}
	w, err := record.WorkOfSeat(run, role, seatID)
	if err != nil {
		return &StandingJSON{Unreadable: err.Error()}
	}
	out := &StandingJSON{Complete: w.Sitting.Complete, Open: len(w.Sitting.Open), Blocking: []string{}}
	for _, it := range w.Sitting.Open {
		if it.Blocks {
			out.Blocking = append(out.Blocking, it.What)
		}
	}
	return out
}
