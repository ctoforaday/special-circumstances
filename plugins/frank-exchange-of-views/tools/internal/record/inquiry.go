package record

import "fmt"

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
	Ev("out-of-scope", "a real question, but not THIS question"),
	Ev("too-thin", "in scope, and the hypothesis does not carry its budget as stated"),
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
// `unsupported` and `absent` put the line on blue's worklist.
var InquirySupports = []EnumValue{
	Ev("supported", "the line is in the report and the text still backs it as stated"),
	Ev("weakened", "still there, and the support has eroded — a flag, not a demand; blue is not obliged to act"),
	Ev("unsupported", "the line is in the report and the text no longer backs it — blue owes a repair or a rebuttal"),
	Ev("absent", "the line is NOT in the report at all — the record claims a direction the document does not carry"),
}

// InquirySupportNames is the bare vocabulary.
func InquirySupportNames() []string { return Names(InquirySupports) }

// SupportDemandsBlue reports whether this verdict puts the line on blue's worklist. `weakened` does
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
		if e.Type == "line-of-inquiry" && e.Payload.Str("inquiry_id") != "" && e.Payload.Str("supersedes_status") == "" {
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
	// Contests is the ruling blue moved AGAINST, recorded by `blue line of inquiry` at the moment of
	// the move. Read from the field rather than re-derived from (status, ruling): the write
	// path already decided what counts as contesting, and a second derivation downstream is a
	// second definition that can disagree with it.
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
		id := e.Payload.Str("inquiry_id")
		if id == "" {
			// A direction motion carries the line of inquiry's id under `motion_id`, because to the
			// motion machinery it IS the motion's id — the proposal is the filing, so there is
			// no second identity to mint. Keying only on `inquiry_id` dropped every ruling made
			// through the new verb before it reached the switch below.
			if e.Type == "motion-rule" && e.Payload.Str("subject") == "inquiry" {
				id = e.Payload.Str("motion_id")
			}
			if id == "" {
				continue
			}
		}
		switch e.Type {
		case "line-of-inquiry":
			a, ok := byID[id]
			if !ok {
				a = &Inquiry{ID: id}
				byID[id] = a
				order = append(order, id)
			}
			// A creation carries the substance; a MOVE carries only the new status and why,
			// so the substance must not be blanked by it.
			if v := e.Payload.Str("line"); v != "" {
				a.Line = v
			}
			if v := e.Payload.Str("hypothesis"); v != "" {
				a.Hypothesis = v
			}
			if v := e.Payload.Str("method"); v != "" {
				a.Method = v
			}
			a.Status, a.Reason, a.Round, a.SeatID = e.Payload.Str("status"), e.Payload.Str("reason"), e.Round, e.SeatID
			a.Contests = e.Payload.Str("contests_ruling")
			a.History = append(a.History, fmt.Sprintf("r%d %s", e.Round, a.Status))
		case "motion-rule":
			// THE CURRENT SPELLING, and reading it here is not optional.
			//
			// A direction motion joins on the line of inquiry's own id, so `motion direction rule` writes
			// a motion-rule whose motion_id IS an A-number. Until this arm existed, a ruling
			// made through the new verb never reached `--view lines-of-inquiry` — the projection
			// blue reads to decide whether to pursue, comply or drop. The line simply stayed
			// "Awaiting a decision", which is what an unruled line looks like, so red's ruling
			// was indistinguishable from red not having sat.
			if e.Payload.Str("subject") != "inquiry" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				continue
			}
			a.Ruling, a.RulingWhy, a.RuledRound = e.Payload.Str("ruling"), e.Payload.Str("reason"), e.Round
		case "inquiry-support":
			a, ok := byID[id]
			if !ok {
				continue
			}
			a.Support, a.SupportWhy, a.SupportRound = e.Payload.Str("as"), e.Payload.Str("reason"), e.Round
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
		if e.Type == "line-of-inquiry" && e.Payload.Str("inquiry_id") == id {
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
		if e.Round > max {
			max = e.Round
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
		if e.Type == "motion-rule" && e.Payload.Str("subject") == "inquiry" && e.Payload.Str("motion_id") == inquiryID {
			ruling = e.Payload.Str("ruling")
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

// UnsupportedInquiries returns the lines red's latest verdict puts on BLUE's worklist — the report
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
