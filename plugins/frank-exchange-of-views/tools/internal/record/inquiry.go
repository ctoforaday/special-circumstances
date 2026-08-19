package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// LINES OF INQUIRY HAVE A LIFECYCLE NOW, BECAUSE THE UNIT IS THE CHOICE, NOT THE ENTRY.
//
// MEASURED, over 86 line of inquiry events across six runs: ZERO lines were ever recorded twice and
// ZERO statuses ever changed. There was no id, no key, no update path — `line of inquiry` was a
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
// choosing. So a line of inquiry gets what a gap has — an id, a status that moves with a stated
// reason, and an adjudicator.

// InquiryStatuses are the states a line of inquiry may hold. `proposed` is the new one: a direction
// blue has put forward and not yet resolved, which is the state the old shape could not
// express at all (everything had to be declared already-pursued or already-dead).
// `deferred` is the fate that had no name: a direction worth taking, and not by THIS run.
// It is not `declined` (judged not worth it) and not `abandoned` (tried, died) — it is kept,
// and it is the carrier for bootstrapping a later run. Deliberately a PROPOSAL for a human
// to select rather than a seed: a run that queues its own successor is a loop with no human
// in it.
var InquiryStatuses = []EnumValue{
	Ev("proposed", "put forward and not yet resolved — the default, and the state that owes a move"),
	Ev("pursued", "you are following it, or you followed it; say what you learned in --reason"),
	Ev("declined", "considered and judged not worth this run's time"),
	Ev("abandoned", "you TRIED it and it died — the most valuable fate, because it stops a later run re-walking it"),
	Ev("deferred", "worth taking, and not by THIS run. --reason says what a later run should pick it up FOR; it reaches the report as a proposal a human selects, never an automatic seed"),
}

// InquiryStatusNames is the bare vocabulary, for readers that only need the words.
func InquiryStatusNames() []string { return Names(InquiryStatuses) }

// InquiryRulings are red's fates for a proposed direction. Red AUDITS and RULES; it never
// proposes one — directing research is what a gap's required_fix already does, and a second
// spelling for it is the aliasing this vocabulary exists to prevent.
var InquiryRulings = []EnumValue{
	Ev("endorsed", "worth this run's time — blue should take it up"),
	Ev("out_of_scope", "a real question, but not THIS question"),
	Ev("too_thin", "in scope, and the hypothesis does not carry its budget as stated"),
}

// InquirySupports are red's per-round verdict on whether the REPORT still carries this line.
//
// This is a different question from InquiryRulings and the two must not be confused. A ruling
// answers "is this direction worth the run's time" — red's judgement about the RESEARCH. A support
// verdict answers "is this line present in the report, and does the text still back it as stated"
// — red's leaf read of the ARTIFACT. A line can be endorsed and unsupported at once: red agreed it
// was worth taking and the section that took it has since been cut.
//
// WHY IT EXISTS. A line of inquiry reached the report as an unaudited row: `assemble` generates it
// from the record, so it carries no citation anchor and `lens verify` cannot reach it. "We pursued
// X" was the one claim in the document that nothing checked, which is the shape this repository
// keeps finding — a fact stated where nothing can refuse it.
//
// `weakened` is the middle grade and earns its place the way a `low` corroboration does: it lets
// red say the support has eroded without demanding a repair blue may reasonably decline. Only
// `unsupported` and `absent` put the line on blue's work list.
var InquirySupports = []EnumValue{
	Ev("supported", "the line is in the report and the text still backs it as stated"),
	Ev("weakened", "still there, and the support has eroded — a flag, not a demand; blue is not obliged to act"),
	Ev("unsupported", "the line is in the report and the text no longer backs it — blue owes a repair or a rebuttal"),
	Ev("absent", "the line is NOT in the report at all — the record claims a direction the document does not carry"),
}

// InquirySupportNames is the bare vocabulary.
func InquirySupportNames() []string { return Names(InquirySupports) }

// SupportDemandsBlue reports whether this verdict puts the line on blue's work list. `weakened` does
// not: it is red saying "this is thinner than it was", which is an argument blue may answer or
// accept, and a duty that fires on it would make every erosion a blocking repair.
func SupportDemandsBlue(verdict string) bool {
	return verdict == "unsupported" || verdict == "absent"
}

// InquiryRulingNames is the bare vocabulary.
func InquiryRulingNames() []string { return Names(InquiryRulings) }

// MintInquiryID assigns the next run-unique line-of-inquiry id (Q1, Q2 …).
//
// Run-unique rather than round-scoped, unlike a gap: a line of inquiry OUTLIVES the round that
// proposed it — that is the whole point of giving it a lifecycle — so a round-scoped id
// would have to be re-minted to survive, which is the bug this replaces.
func MintInquiryID(runDir string) (string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	n := 0
	for _, e := range m.Events {
		// THE BODY IS THE TYPE. `line-of-inquiry` is carried by the `Avenue` message
		// (EVENT_TYPE_AVENUE) — the schema kept the pre-rename spelling for the message and its
		// id field, so `inquiry_id` is `avenue_id` here. Matching on the body cannot go stale
		// against the enum the way `e.Type == "line-of-inquiry"` could.
		av, ok := recordpb.BodyAs[*recordpb.Avenue](e)
		if !ok {
			continue
		}
		// A PROPOSAL, NOT A MOVE. `supersedes_status` is PRESENT on a move and absent on a
		// proposal — the schema says so in as many words — so the id counter reads the pointer,
		// not the string. `GetSupersedesStatus() == ""` would count a move whose marker was
		// written empty as a fresh proposal and mint an id that already exists.
		if av.GetAvenueId() != "" && av.SupersedesStatus == nil {
			n++
		}
	}
	return fmt.Sprintf("Q%d", n+1), nil
}

// Inquiry is one direction's state after replay: its latest status, with the history that
// produced it. The history is kept because "chose to abandon this at round 2, having
// pursued it at round 0" is the evidence of choosing, and only the sequence carries it.
type Inquiry struct {
	ID         string
	Line       string
	Hypothesis string
	Method     string
	Status     string
	Reason     string   // the reason attached to the CURRENT status
	Round      int      // the round the current status was set in
	History    []string // "r0 pursued", "r2 abandoned" …
	SeatID     string   // who last moved it — attribution the one-line row has always carried
	Ruling     string   // red's fate, if ruled
	RulingWhy  string
	RuledRound int
	// Contests was the ruling blue moved AGAINST, recorded by `blue line-of-inquiry` at the moment
	// of the move. Read from the field rather than re-derived from (status, ruling): the write
	// path already decided what counts as contesting, and a second derivation downstream is a
	// second definition that can disagree with it.
	//
	// IT IS NOW ALWAYS EMPTY, AND SAYING SO HERE IS THE POINT. Its carrier was the payload key
	// `contests_ruling`, which the schema deliberately does not have — recordpb's key census calls
	// it "the legacy spelling of an appeal … the one legacy field with no counterpart at all", and
	// #344 replaced the mechanism with `motion inquiry appeal`. Nothing has written it since;
	// blue/inquiry.go:109 records why. The field stays because report/assemble.go still renders
	// it, and because the CONCEPT is live: its post-#344 carrier is a `motion-appeal` event
	// (MotionAppeal, subject DIRECTION) on this line's id, which this projection has never read.
	// Wiring that is new behaviour rather than a conversion, so it is reported, not done here.
	Contests string
	// Support is red's latest verdict on whether the REPORT still carries this line, with the
	// round it was cast in. SupportRound is what makes "voted THIS round" answerable — the duty
	// is per-round, so a verdict from two rounds ago is not a verdict for this one.
	Support      string
	SupportWhy   string
	SupportRound int
}

// Inquiries replays the line of inquiry events into current state, in proposal order.
func Inquiries(b *Board) []*Inquiry {
	byID := map[string]*Inquiry{}
	var order []string
	for _, e := range b.Events {
		body, ok := recordpb.Body(e)
		if !ok {
			// NO BODY IS NOT AN EMPTY ONE. An event the schema carries no body for names no line
			// of inquiry, which is the same outcome the old `Str("inquiry_id") == ""` reached —
			// but reached here by asking the question rather than by a lookup that misses.
			continue
		}
		switch t := body.(type) {
		case *recordpb.Avenue:
			// The id is `avenue_id`: the schema kept the pre-rename spelling, and the old
			// `inquiry_id` key is the same fact under the newer word.
			id := t.GetAvenueId()
			if id == "" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				a = &Inquiry{ID: id}
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
			// zero's name would put a line in the lines-of-inquiry projection under a fate no
			// seat ever chose, and into History as "r2 unspecified". The value form is used
			// deliberately — `Word(t.Status)` would pass a typed nil pointer into an interface
			// that is not nil, and panic on the absent case this line exists to handle.
			//
			// AvenueStatus needs no hyphen join: none of proposed/pursued/deferred/declined/
			// abandoned carries an underscore. DirectionRuling below is the opposite case.
			a.Status = recordpb.Word(t.GetStatus())
			a.Reason, a.Round, a.SeatID = t.GetReason(), int(e.GetRound()), e.GetSeatId()
			// `contests_ruling` HAS NO FIELD, AND THAT IS THE SCHEMA'S DECISION, NOT THIS
			// CONVERSION'S. It was set as a side effect of moving a line to `pursued` against an
			// adverse ruling; #344 replaced it with `motion inquiry appeal`, blue/inquiry.go:109
			// records that nothing has written it since, and recordpb's key census calls it "the
			// legacy spelling of an appeal … the one legacy field with no counterpart at all".
			// So the read is dropped rather than converted, and Inquiry.Contests is now always
			// empty. THE CONCEPT IS NOT DEAD: its post-#344 carrier is a `motion-appeal` event on
			// this line's id, which this projection has never read. Wiring that is new behaviour,
			// not a conversion, so it is reported rather than done here.
			a.History = append(a.History, fmt.Sprintf("r%d %s", e.GetRound(), a.Status))
		case *recordpb.MotionRule:
			// THE CURRENT SPELLING, and reading it here is not optional.
			//
			// A direction motion joins on the line of inquiry's own id, so `motion inquiry rule`
			// writes a motion-rule whose motion_id IS a Q-number. Until this arm existed, a ruling
			// made through the new verb never reached `--view lines-of-inquiry` — the projection
			// blue reads to decide whether to pursue, comply or drop. The line simply stayed
			// "Awaiting a decision", which is what an unruled line looks like, so red's ruling
			// was indistinguishable from red not having sat.
			//
			// The subject the CLI spells `inquiry` is MOTION_SUBJECT_DIRECTION in the schema —
			// the same subject under the schema's word, and the only one whose ruling set is
			// DirectionRuling (endorsed / out-of-scope / too-thin).
			if t.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
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
			// direction ruling and must leave Inquiry.Ruling empty — `GetDirection()` alone
			// returns UNSPECIFIED for all three cases and cannot tell them apart.
			//
			// DIRECTION IS ONE OF THE TWO HYPHENATED VOCABULARIES word.go warns about, and the
			// `_` -> `-` join is not cosmetic. `Spelling` derives the word from the generated
			// constant, so DIRECTION_RULING_OUT_OF_SCOPE reads `out_of_scope`; the seat types
			// `out-of-scope`, InquiryRulings above spells it with a hyphen, `motion inquiry rule
			// --as` validates against the hyphen, and report/assemble renders it into the
			// document. The underscore form is a word no surface recognizes.
			//
			// SHARED CODE, DECLARED, NOT DEFINED HERE (contract rule 2). viewjson.go marks the
			// identical join for Grade as `gradeWord`, and word.go names `record.GradeStr` as
			// where it belongs. Written inline at this file's two sites rather than as a private
			// third copy; named in this agent's return so the lead places one.
			a.Ruling = ""
			if d, isDirection := t.GetRuling().(*recordpb.MotionRule_Direction); isDirection {
				a.Ruling = recordpb.Word(d.Direction)
			}
			// `reason` on the wire is `opinion` on the message — the ruler's argument, which is
			// the field MotionRule carries and the only prose channel it has.
			a.RulingWhy, a.RuledRound = t.GetOpinion(), int(e.GetRound())
		// A LENS'S READ OF THE LINE AGAINST THE CURRENT REPORT.
		//
		// This replaces `merge inquiry-support`, which existed only because a generated row
		// carried no anchor and so no verification could reach it. The row is anchored now and
		// the read is a lens's work, which is what a lens does: check the artifact at the leaf.
		//
		// THREE STATES, and the set is complete by construction rather than by enumeration —
		// present-and-backed, present-and-unbacked, absent. The retired set's fourth value,
		// `weakened`, was a point on a continuum, and a seat asked to place a judgement on a
		// continuum argues the placement instead of making the call.
		case *recordpb.InquiryCheck:
			a, ok := byID[t.GetAvenueId()]
			if !ok {
				continue
			}
			a.Support, a.SupportWhy, a.SupportRound = recordpb.Word(t.GetState()), t.GetReason(), int(e.GetRound())
		}
	}
	out := make([]*Inquiry, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// requireInquiry refuses a reference to a line of inquiry no proposal created — the same discipline
// every other cross-reference gets (refs.go), for the same reason: a dangling reference is
// accepted at write time and dropped at replay.
func RequireInquiryRef(runDir, id string) error {
	if id == "" {
		return nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		// EVERY Avenue event answers this, proposal or move, exactly as the old type test did —
		// the check is that the id was ever WRITTEN by the line-of-inquiry verb, not that the
		// event was the proposal.
		if av, ok := recordpb.BodyAs[*recordpb.Avenue](e); ok && av.GetAvenueId() == id {
			return nil
		}
	}
	return fmt.Errorf("record: --id names line of inquiry %s, which no line of inquiry event proposed — a dangling reference is accepted here and dropped at replay", id)
}

// CurrentRound is the highest round any event on this board reached.
//
// There was no shared accessor for it, which is part of why the predicate below drifted: a
// round-aware check had nothing to be aware WITH, so two callers wrote a status-only test and
// described it in prose as though it looked at the round.
func CurrentRound(b *Board) int {
	max := 0
	for _, e := range b.Events {
		// An event with no round reads as 0, which is what the old `int` field held when the key
		// was absent — round 0 is a real round, so there is nothing here for presence to say.
		if r := int(e.GetRound()); r > max {
			max = r
		}
	}
	return max
}

// StaleInquiries returns the inquiries that owe blue a decision THIS ROUND — the single predicate
// behind the revisit duty, the affordance, and the projection's stale notice.
//
// # It was written twice, both copies status-only, both described as round-aware
//
// This function and `AvailableOf`'s blue case each carried `Status == "proposed" || Status ==
// "pursued"` and nothing else. The affordance's text says a line of inquiry "has no fate THIS ROUND";
// this one's said "a line of inquiry still open LATE IN A RUN" and "nothing ever asked blue to choose
// again once the round-0 plan was written". Neither read a round. `Inquiry.Round` was populated
// on every event and consulted nowhere.
//
// So a line of inquiry blue moved to `pursued` this round, with a --reason saying what it learned, and
// one nobody had touched since round 0 produced the IDENTICAL line. The diligent case and the
// neglected case were the same bytes — and the fix is one predicate rather than two corrected
// copies, because two copies is how this got here.
//
// # `pursued` is not an unresolved state, and treating it as one inverted the intent
//
// The enum defines `pursued` as "you are following it, OR YOU FOLLOWED IT; say what you learned
// in --reason". It is where a line that paid off comes to rest. Firing on it unconditionally
// meant recording exactly the right thing never cleared the duty, and the only statuses that
// did — `declined`, `abandoned`, `deferred` — all mean STOP. The channel could express giving
// up and could not express carrying on, which is the reverse of what it exists for.
//
// # The rule, and the round check applies to ONE status rather than to all of them
//
//	proposed    ALWAYS owes a move. The enum defines it as "put forward and not yet resolved —
//	            the default, AND THE STATE THAT OWES A MOVE". No round condition: a topic nobody
//	            has decided is undecided whenever you ask. A first draft of this fix also
//	            round-gated `proposed`, which made a line of inquiry proposed and abandoned within a
//	            single round invisible for that round — caught by
//	            TestOpenInquiriesAreSurfacedAsOwingADecision, which was right and this was wrong.
//	proposed    is therefore surfaced from the moment it exists. It is an AFFORDANCE and blocks
//	            nothing, so surfacing early costs a line and buys the reminder.
//	pursued     owes a move only when it has not moved THIS round. This is where the round check
//	            belongs and the only place it ever did: re-recording `pursued` with what you
//	            learned is a reaffirmation, and it settles the line for the round.
//	declined,
//	abandoned,
//	deferred    settled, never surfaced. `deferred` is a DECISION ("worth taking, and not by THIS
//	            run"), not an omission.
func StaleInquiries(b *Board) []*Inquiry {
	now := CurrentRound(b)
	var out []*Inquiry
	for _, a := range Inquiries(b) {
		switch a.Status {
		case "proposed":
			out = append(out, a)
		case "pursued":
			if a.Round < now {
				out = append(out, a)
			}
		}
	}
	return out
}

// InquiryRuling returns red's most recent ruling on a line of inquiry, or "" if it never ruled.
//
// The ruling and the line of inquiry's fate were both on the record and joined NOWHERE, so blue
// pursuing a line red called out-of-scope looked exactly like blue pursuing one red endorsed.
// Red's ruling is an argument rather than a command — blue may pursue anyway — but the
// disagreement should be a fact, not something a reader reconstructs from two lists.
func InquiryRuling(runDir, inquiryID string) string {
	b, err := BoardState(runDir)
	if err != nil {
		return ""
	}
	// MOST RECENT WINS. Events are ordered by timestamp across shards, so "last one seen" is the
	// latest ruling. There was a second arm here for the pre-#344 `avenue-rule` spelling; nothing
	// has written it since the motion collapse and the dual-read that justified reading it is gone.
	ruling := ""
	for _, e := range b.Events {
		mr, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || mr.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION || mr.GetMotionId() != inquiryID {
			continue
		}
		// THE LATEST RULING WINS, INCLUDING A LATER ONE THAT CARRIED NONE. The old
		// `Str("ruling")` overwrote with "" for a motion-rule that named no verdict, and the
		// oneof is what preserves that: an unset `ruling` arm, or another subject's arm, is not
		// a direction ruling. `GetDirection()` alone returns UNSPECIFIED for all three and
		// cannot tell "red ruled nothing" from "red is not on this event".
		// The hyphen join, for the reason given at the same read in `Inquiries`.
		ruling = ""
		if d, isDirection := mr.GetRuling().(*recordpb.MotionRule_Direction); isDirection {
			ruling = strings.ReplaceAll(recordpb.Word(d.Direction), "_", "-")
		}
	}
	return ruling
}

// UnvotedInquiries returns the lines red has not cast a support verdict on THIS round.
//
// The vote is per-round by design: the question is whether the report STILL carries the line, and
// a report that changed this round has not been checked by a verdict cast before it did. A verdict
// that carried forward would answer a question about a document that no longer exists.
func UnvotedInquiries(b *Board) []*Inquiry {
	now := CurrentRound(b)
	var out []*Inquiry
	for _, a := range Inquiries(b) {
		if a.Support == "" || a.SupportRound < now {
			out = append(out, a)
		}
	}
	return out
}

// UnsupportedInquiries returns the lines red's latest verdict puts on BLUE's work list — the report
// no longer backs them, or does not carry them at all. `weakened` is deliberately not here: see
// SupportDemandsBlue.
func UnsupportedInquiries(b *Board) []*Inquiry {
	var out []*Inquiry
	for _, a := range Inquiries(b) {
		if SupportDemandsBlue(a.Support) {
			out = append(out, a)
		}
	}
	return out
}
