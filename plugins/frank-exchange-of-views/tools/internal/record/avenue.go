package record

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AVENUES HAVE A LIFECYCLE NOW, BECAUSE THE UNIT IS THE CHOICE, NOT THE ENTRY.
//
// MEASURED, over 86 avenue events across six runs: ZERO lines were ever recorded twice and
// ZERO statuses ever changed. There was no id, no key, no update path — `avenue` was a
// one-shot append. 83 of the 86 landed in round 0.
//
// That makes the corpus's headline number mean something other than it appeared to. 68
// "pursued" is not 68 directions pursued to completion; it is 68 INTENTIONS DECLARED BEFORE
// ANY RESEARCH HAPPENED. If a direction died in round 2 there was no mechanism to say so,
// and the 21% rejection rate measured only what blue could rule out before starting — not
// how hard it looked.
//
// The goal is that blue finds several plausible directions, picks the best, and is SEEN TO
// HAVE DONE SO IN EVIDENCE. A one-shot append records the plan; it cannot record the
// choosing. So an avenue gets what a gap has — an id, a status that moves with a stated
// reason, and an adjudicator.

// AvenueStatuses are the states an avenue may hold. `proposed` is the new one: a direction
// blue has put forward and not yet resolved, which is the state the old shape could not
// express at all (everything had to be declared already-pursued or already-dead).
// `deferred` is the fate that had no name: a direction worth taking, and not by THIS run.
// It is not `declined` (judged not worth it) and not `abandoned` (tried, died) — it is kept,
// and it is the carrier for bootstrapping a later run. Deliberately a PROPOSAL for a human
// to select rather than a seed: a run that queues its own successor is a loop with no human
// in it.
// DERIVED FROM THE ENUM'S OWN `means`, like every other vocabulary with a declaration. This was a
// hand-kept second copy whose glosses had already drifted from the proto's, so a seat's --help and
// the record's vocabulary table described the same five words differently.
var AvenueStatuses = evsOf(recordpb.AvenueStatus(0).Descriptor())

// AvenueStatusNames is the bare vocabulary, for readers that only need the words.
func AvenueStatusNames() []string { return Names(AvenueStatuses) }

// AvenueRulings are red's fates for a proposed direction. Red AUDITS and RULES; it never
// proposes one — directing research is what a gap's required_fix already does, and a second
// spelling for it is the aliasing this vocabulary exists to prevent.
var AvenueRulings = []EnumValue{
	ev("endorsed", "worth this run's time — blue should take it up"),
	ev("out_of_scope", "a real question, but not THIS question"),
	ev("too_thin", "in scope, and the hypothesis does not carry its budget as stated"),
}

// MintAvenueID assigns the next run-unique avenue id (Q1, Q2 …).
//
// Run-unique rather than epoch-scoped, unlike a gap: an avenue OUTLIVES the epoch that
// proposed it — that is the whole point of giving it a lifecycle — so an epoch-scoped id
// would have to be re-minted to survive, which is the bug this replaces.
func MintAvenueID(run Run) (string, error) {
	// A PROPOSAL, NOT A MOVE. `supersedes_status` is PRESENT on a move and absent on a
	// proposal — the schema says so in as many words — so the id counter reads presence (a
	// NULL column), not the string. A move whose marker was written empty would still carry a
	// non-NULL column and is not counted, exactly as the fold's pointer test had it.
	//
	// ONE QUERY, over the proposals that STAND. A read that fails is the caller's error: the open
	// already held this record to the binary's schema, so the failure is a real one (busy, locked,
	// malformed), and an id minted from any other count would be a plausible wrong number.
	var n int
	if _, err := queryRow(run, []any{&n},
		`SELECT count(*) FROM "avenue" a JOIN "live_event" l ON l."event_id" = a."event_id"
		  WHERE COALESCE(a."avenue_id", '') != '' AND a."supersedes_status" IS NULL`); err != nil {
		return "", err
	}
	return fmt.Sprintf("Q%d", n+1), nil
}

// Avenue is one direction's state after replay: its latest status, with the history that
// produced it. The history is kept because "chose to abandon this at round 2, having
// pursued it at round 0" is the evidence of choosing, and only the sequence carries it.
type Avenue struct {
	ID         string
	Line       string
	Hypothesis string
	Method     string
	Status     string
	Reason     string   // the reason attached to the CURRENT status
	Epoch      int      // the epoch the current status was set in
	History    []string // "r0 pursued", "r2 abandoned" …
	// EverPursued is whether any event on the line recorded `pursued`. The fate word alone implies a
	// history it does not carry: `abandoned` reads as "tried, then died", and nothing in `move`
	// refuses abandoning a line straight from `proposed` — measured once each in #861's B and B3,
	// both genuine attempts whose seat skipped the `pursued` move. Derived here, from the events,
	// so the report can say exactly what the record holds (`[abandoned before pursuit]`).
	EverPursued bool
	SeatID      string // who last moved it — attribution the one-line row has always carried
	Ruling      string // red's fate, if ruled
	RulingWhy   string
	RuledEpoch  int
	// Contests was the ruling blue moved AGAINST, recorded by `blue avenue` at the moment
	// of the move. Read from the field rather than re-derived from (status, ruling): the write
	// path already decided what counts as contesting, and a second derivation downstream is a
	// second definition that can disagree with it.
	//
	// ITS CARRIER IS THE APPEAL, NOT A FIELD ON THE LINE. The payload key `contests_ruling` is gone
	// — recordpb's key census calls it "the legacy spelling of an appeal", and #344 replaced the
	// mechanism with `motion avenue appeal`. So this is set by the MotionAppeal arm of AvenuesOf
	// below, from the ruling already on the line: an appeal against an unruled line cannot be
	// written (RequireRuledMotion). Its reader is the avenues projection
	// (view.AvenueBody), which ships as avenues.md; judgments.md carries the same appeal
	// with the filer's reason. report.md does not render it — the debate over a direction is not
	// research prose.
	Contests string
	// THERE IS NO PER-LINE SUPPORT VERDICT, AND ITS ABSENCE IS A RULING RATHER THAN AN OMISSION.
	// Three fields here — Support, SupportWhy, SupportRound — carried red's per-epoch answer to
	// "does the report still CARRY this line". That made presence the question. Presence is not a
	// question: the lines reach the report on the WORKLIST, generated from this projection, so
	// blue cannot cut them. What remains — did blue's body deliver the research — is an ordinary
	// GAP, minted and closed like any other, and the per-epoch statement that the read HAPPENED is
	// AvenueReviewDue's business, not this struct's.
}

// AvenuesOf is Avenues over the events themselves, for the run-shaped readers.
// StruckText is a struck act's text, with who struck it and why.
type StruckText struct {
	Text string
	Struck
}

// StruckAvenueTexts is, per avenue, the wording of each proposal or move its seat corrected
// in the sitting — what the avenues LISTING shows struck beside the line that stands.
// AvenuesOf folds the acts that stand; this is the part of the listing that fold drops.
func StruckAvenueTexts(evs []*Event) map[string][]StruckText {
	out := map[string][]StruckText{}
	for _, l := range Listing(evs) {
		a, ok := recordpb.BodyAs[*recordpb.Avenue](l.Event)
		if !ok || l.Struck == nil {
			continue
		}
		text := a.GetLine()
		if text == "" {
			text = a.GetReason()
		}
		out[a.GetAvenueId()] = append(out[a.GetAvenueId()], StruckText{Text: text, Struck: *l.Struck})
	}
	return out
}

func AvenuesOf(evs []*Event) []*Avenue {
	// The acts that stand: a proposal or move corrected in its sitting is read as its replacement.
	evs = Live(evs)
	byID := map[string]*Avenue{}
	var order []string
	var clk Clock
	for _, e := range evs {
		w := clk.Advance(e)
		body, ok := recordpb.Body(e)
		if !ok {
			// NO BODY IS NOT AN EMPTY ONE. An event the schema carries no body for names no
			// avenue, which is the same outcome the old `Str("avenue_id") == ""` reached —
			// but reached here by asking the question rather than by a lookup that misses.
			continue
		}
		switch t := body.(type) {
		case *recordpb.Avenue:
			// The id is `avenue_id`: the schema kept the pre-rename spelling, and the old
			// `avenue_id` key is the same fact under the newer word.
			id := t.GetAvenueId()
			if id == "" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				a = &Avenue{ID: id}
				byID[id] = a
				order = append(order, id)
			}
			// A creation carries the substance; a MOVE carries only the new status and why,
			// so the substance must not be blanked by it. These stay VALUE tests rather than
			// presence tests on purpose: the question they ask is "did this event bring
			// substance", and a move that carried `line` as an empty string must not blank a
			// proposal's line either. `av.Line != nil` would let it.
			if v := t.GetLine(); v != "" {
				a.Line = v
			}
			if v := t.GetHypothesis(); v != "" {
				a.Hypothesis = v
			}
			if v := t.GetMethod(); v != "" {
				a.Method = v
			}
			// STATUS IS OVERWRITTEN BY EVERY EVENT, including one that carried none — that is
			// what "latest status" means, and the old `Str("status")` did exactly this.
			//
			// `Word` is what makes the absent case survive: an unset status is the enum zero, and
			// Word maps the zero to "" rather than to the literal `unspecified`. Rendering the
			// zero's name would put a line in the avenues projection under a fate no
			// seat ever chose, and into History as "r2 unspecified". The value form is used
			// deliberately — `Word(t.Status)` would pass a typed nil pointer into an interface
			// that is not nil, and panic on the absent case this line exists to handle.
			//
			// AvenueStatus needs no hyphen join: none of proposed/pursued/concluded/deferred/
			// declined/abandoned carries an underscore. AvenueRuling below is the opposite case.
			a.Status = recordpb.Word(t.GetStatus())
			// A CONCLUDED LINE WAS TAKEN: it ends a pursuit, so it is one whether or not the seat
			// recorded `pursued` on the way.
			if st := t.GetStatus(); st == recordpb.AvenueStatus_AVENUE_STATUS_PURSUED || st == recordpb.AvenueStatus_AVENUE_STATUS_CONCLUDED {
				a.EverPursued = true
			}
			a.Reason, a.Epoch, a.SeatID = t.GetReason(), w.Epoch, e.GetSeatId()
			// `contests_ruling` HAS NO FIELD, AND THAT IS THE SCHEMA'S DECISION, NOT THIS
			// CONVERSION'S. It was set as a side effect of moving a line to `pursued` against an
			// adverse ruling; #344 replaced it with `motion avenue appeal`, blue/avenue.go:109
			// records that nothing has written it since, and recordpb's key census calls it "the
			// legacy spelling of an appeal … the one legacy field with no counterpart at all".
			// So the read is dropped rather than converted. THE CONCEPT IS NOT DEAD: its post-#344
			// carrier is a `motion-appeal` event on this line's id, read by the MotionAppeal arm
			// below, which is what sets Avenue.Contests now.
			a.History = append(a.History, fmt.Sprintf("e%d %s", w.Epoch, a.Status))
		case *recordpb.MotionRule:
			// THE CURRENT SPELLING, and reading it here is not optional.
			//
			// A direction motion joins on the avenue's own id, so `motion avenue rule`
			// writes a motion-rule whose motion_id IS a Q-number. Until this arm existed, a ruling
			// made through the new verb never reached `--view avenues` — the projection
			// blue reads to decide whether to pursue, comply or drop. The line simply stayed
			// "Awaiting a decision", which is what an unruled line looks like, so red's ruling
			// was indistinguishable from red not having sat.
			//
			// The subject the CLI spells `avenue` is MOTION_SUBJECT_AVENUE in the schema —
			// the same subject under the schema's word, and the only one whose ruling set is
			// AvenueRuling (endorsed / out-of-scope / too-thin).
			if t.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_AVENUE {
				continue
			}
			id := t.GetMotionId()
			if id == "" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				continue
			}
			// AN ABSENT RULING IS THE EMPTY WORD, and the oneof is what says so: a motion-rule
			// whose `ruling` arm is unset, or is set to another subject's arm, carried no
			// direction ruling and must leave Avenue.Ruling empty — `GetAvenue()` alone
			// returns UNSPECIFIED for all three cases and cannot tell them apart.
			//
			// NO `_` -> `-` JOIN, AND THE COMMENT THAT DEMANDED ONE WAS STALE. It said the seat
			// types `out-of-scope`, that AvenueRulings spells it with a hyphen, and that the
			// underscore form "is a word no surface recognizes". Checked: AvenueRuling spells
			// AVENUE_RULING_OUT_OF_SCOPE, `Word` yields `out_of_scope`, and AvenueRulings
			// carries `out_of_scope` too. The hyphen is what no surface recognizes now, so the
			// word goes through unchanged and there is no third spelling to keep in step.
			a.Ruling = ""
			if d, isAvenue := t.GetRuling().(*recordpb.MotionRule_Avenue); isAvenue {
				a.Ruling = recordpb.Word(d.Avenue)
			}
			// `reason` on the wire is `opinion` on the message — the ruler's argument, which is
			// the field MotionRule carries and the only prose channel it has.
			a.RulingWhy, a.RuledEpoch = t.GetOpinion(), w.Epoch
		case *recordpb.MotionAppeal:
			// BLUE MOVING AGAINST A RULING, which is the post-#344 carrier of `contests_ruling`.
			//
			// The field it replaced was set as a side effect of moving a line to `pursued` against
			// an adverse ruling, and when it was retired this arm was NOT written — so
			// `Avenue.Contests` was always empty and the report's "blue took this line against
			// red's X ruling" line could never render. The comment above recorded that as owed
			// rather than done; this is the doing.
			//
			// What blue contested is the ruling ON THE RECORD, so it is read off the line rather
			// than restated by the appeal: an appeal names the motion, and the motion's ruling is
			// already here. An appeal against a line nobody ruled leaves it empty, because there
			// is nothing to have moved against.
			if t.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_AVENUE {
				continue
			}
			a, ok := byID[t.GetMotionId()]
			if !ok {
				continue
			}
			a.Contests = a.Ruling
			// THERE IS NO AvenueReview ARM, AND THAT IS THE SHAPE RATHER THAN A GAP IN IT. The
			// review is ONE event per epoch about the report as a whole; it names no line, so there
			// is nothing here for it to join to. Its reader is AvenueReviewDue.
		}
	}
	out := make([]*Avenue, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// requireAvenue refuses a reference to an avenue no proposal created — the same discipline
// every other cross-reference gets (refs.go), for the same reason: a dangling reference is
// accepted at write time and dropped at replay.
func RequireAvenueRef(run Run, id string) error {
	if id == "" {
		return nil
	}
	// EVERY Avenue event answers this, proposal or move, exactly as the fold's type test did —
	// the check is that the id was ever WRITTEN by the avenue verb, not that the
	// event was the proposal.
	found, err := recordHas(run, `SELECT 1 FROM "avenue" WHERE "avenue_id" = ? LIMIT 1`, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("record: --id names avenue %s, which no avenue event proposed — a dangling reference is accepted here and dropped at replay", id)
}

// StaleAvenuesOf is StaleAvenues over the events themselves.
func StaleAvenuesOf(evs []*Event) []*Avenue {
	now := CurrentEpochOf(evs)
	var out []*Avenue
	for _, a := range AvenuesOf(evs) {
		switch a.Status {
		case "proposed":
			out = append(out, a)
		case "pursued":
			if a.Epoch < now {
				out = append(out, a)
			}
		}
	}
	return out
}

// AvenueRuling returns red's most recent ruling on an avenue, or "" if it never ruled.
//
// The ruling and the avenue's fate were both on the record and joined NOWHERE, so blue
// pursuing a line red called out-of-scope looked exactly like blue pursuing one red endorsed.
// Red's ruling is an argument rather than a command — blue may pursue anyway — but the
// disagreement should be a fact, not something a reader reconstructs from two lists.
func AvenueRuling(run Run, avenueID string) string {
	// The avenue view carries the whole line, this column included: the newest
	// direction-subject rule decides, and a NULL arm on it is red ruling nothing — "" — not an
	// invitation to read an older ruling instead. A read error folds into "", as the
	// board-read error did. The hyphen join is the surface spelling, so it stays a Go concern.
	var word sql.NullString
	found, err := queryRow(run, []any{&word},
		`SELECT "avenue_ruling" FROM "avenue_state" WHERE "avenue_id" = ?`, avenueID)
	if err != nil || !found {
		return ""
	}
	return strings.ReplaceAll(word.String, "_", "-")
}

// AvenueReviewDueOf is AvenueReviewDue over the events themselves.
func AvenueReviewDueOf(evs []*Event) bool {
	evs = Live(evs)
	if len(AvenuesOf(evs)) == 0 {
		return false
	}
	now := CurrentEpochOf(evs)
	var clk Clock
	for _, e := range evs {
		w := clk.Advance(e)
		// THE BODY IS THE TYPE, as everywhere else in this file: a match on the message cannot go
		// stale against the enum. No field is read — the event's existence in this epoch IS the
		// fact — but the type test still goes through the body so a renamed enum value fails to
		// compile rather than silently matching nothing.
		if _, ok := recordpb.BodyAs[*recordpb.AvenueReview](e); ok && w.Epoch == now {
			return false
		}
	}
	return true
}

// AvenueJSON is one avenue as a machine reads it — the same fields the rendered
// projection groups by fate, with their types intact.
//
// WHY THE SHAPE IS REPEATED HERE rather than tagging Avenue itself: Avenue is the internal fold,
// and tagging it would publish every future field to the wire by default. The wire form is a
// decision each time a field is added, which is what keeps `show avenues --json` a
// contract rather than a dump of whatever the fold happens to carry.
type AvenueJSON struct {
	ID         string `json:"id"`
	Line       string `json:"line"`
	Hypothesis string `json:"hypothesis,omitempty"`
	Method     string `json:"method,omitempty"`
	Status     string `json:"status"`
	// Reason belongs to the CURRENT status, not to the line: a reader pairing it with an earlier
	// status in History would be reading the wrong justification for the wrong move.
	Reason  string   `json:"reason,omitempty"`
	Epoch   int      `json:"epoch"`
	History []string `json:"history,omitempty"`
	// EverPursued distinguishes "tried, then died" from "abandoned before it was ever tried" —
	// which the fate word alone cannot, and which `move` does not refuse.
	EverPursued bool   `json:"ever_pursued"`
	SeatID      string `json:"seat_id"`
	Ruling      string `json:"ruling,omitempty"`
}

// AvenuesJSON is the wire form of the exploration space.
type AvenuesJSON struct {
	Avenues []AvenueJSON `json:"avenues"`
}

// AvenuesJSONOf folds the events into the wire form, in the same order the rendered projection
// walks them, so the two accounts of the same directions cannot disagree about order.
func AvenuesJSONOf(evs []*Event) AvenuesJSON {
	out := AvenuesJSON{Avenues: []AvenueJSON{}}
	for _, a := range AvenuesOf(evs) {
		out.Avenues = append(out.Avenues, AvenueJSON{
			ID: a.ID, Line: a.Line, Hypothesis: a.Hypothesis, Method: a.Method,
			Status: a.Status, Reason: a.Reason, Epoch: a.Epoch, History: a.History,
			EverPursued: a.EverPursued, SeatID: a.SeatID, Ruling: a.Ruling,
		})
	}
	return out
}
