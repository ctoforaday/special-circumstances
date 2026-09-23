package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHAT THIS SEAT STILL OWES, ON THE READ IT ALREADY DOES.
//
// A seat had no way to know it was finished. Asked directly, the chair named a real mechanism —
// "the `verdict` command either succeeds or fails; if it fails the tool tells me what's blocking
// closure, and I iterate" — because `verdict` refuses over open MATERIAL gaps and unruled motions
// and enumerates them. Blue and the bench answered with things they cannot observe: "red agrees it's
// sound", "the bench has ruled". Those are other seats' future acts. A seat whose completion
// condition is someone else's next move cannot know it is done; it can only stop and hand over.
//
// The fix is NOT another verb. The first thing a seat does is read the board, and the view it
// reads can carry both halves — what is outstanding, and whether anything is. A `can-i-finish`
// command would be a second way to ask a question the first read should already answer, and this
// surface is large enough already.
//
// So the work list becomes the seat's own pending work. It was described as "the merge's
// shrinking working set", and every other role defaulted somewhere unhelpful: blue's default read
// was `changelog` — a record of what it had ALREADY done, handed to it before it had done
// anything.
//
// # Only duties that are enforced or recorded somewhere else
//
// Nothing here invents an obligation. Each one is refused at a write path (open material gaps and unruled
// motions block `verdict`; a computation gap cannot be closed on prose), is a stated epoch-record
// requirement (W1.7's revision, the bench's terminal outcome), or is enforced by dispatch (a seat
// that has not registered for the sitting it was dispatched for is readied again; a chair that has
// not registered for its sitting is refused `dispatch next`). Inventing a
// duty here would make this view disagree with the gates, and a seat told it was finished by one
// surface and refused by another learns to trust neither.

// Duty is one outstanding obligation. It says WHAT is owed, and nothing about how.
//
// # The only page that instructs is the help page
//
// This carried a `how`: an invocation, flag by flag, for each way out. That is a hand-kept SECOND
// COPY of a surface the tool already generates, and it drifted exactly as a second copy does —
// the lens affordance told a seat to run `verify --key <k>`, a flag `lens verify` does not have,
// so the only instruction that duty ever handed over could not run. Nobody caught it because the
// affordance that produces it never fires (see citedClaimsWithoutVerify): a dead path holding a
// wrong instruction, each hiding the other.
//
// A duty is STATE — this is owed, and here is why. `<verb> --help` is where the flags,
// their enums and their required-ness live, generated from the command that enforces them, so it
// cannot disagree with the write path. Naming the verb inside `what` is a pointer; naming its
// flags here would be the same second copy under a shorter name.
type Item struct {
	What string `json:"what"`
	// Blocks answers "is this what is stopping me closing" — the only distinction this list
	// draws, and it is drawn PER ITEM rather than by splitting the list in two.
	//
	// It used to be two lists: `outstanding`, which gated `complete`, and `available`, which was
	// documented as never touching it. The split was defensible one clause at a time — a duty
	// must be enforced at a write path, an affordance is not an obligation, and inventing duties
	// would make this view disagree with the gates. Its consequence was that a seat's completion
	// check could only ever see the mechanically-enforced half, so a petition, an appeal, a
	// corroboration and a re-run were not merely unlisted as owed: they were absent from the one
	// surface a seat consults to ask whether there is anything left. Three seats interviewed
	// about the verbs they never touched gave the same account of stopping — "the `outstanding`
	// array emptied" — and one named the mechanism exactly: "things off the list weren't
	// declined, they were invisible."
	//
	// So both halves are now on one list, and the honest distinction they were split to protect
	// survives as this field. `complete` still agrees with the gates, because it is computed from
	// the blocking items alone; nothing here invents an obligation. What changed is that work a
	// seat may do no longer has to be absent in order to not be owed.
	Blocks bool `json:"blocks"`
}

// SittingJSON is the seat's work: everything open to it, whether the sitting is finished, and —
// on each item — whether that item is what is stopping it finishing.
type SittingJSON struct {
	Seat string `json:"seat"`
	Role string `json:"role"`
	// Complete is the answer to "may I end my turn". It is false whenever a BLOCKING item is
	// open, and a seat that ends anyway is ending against a stated list rather than in the dark
	// — which is the difference this exists to make.
	Complete bool   `json:"complete"`
	Open     []Item `json:"open"`
	// LastSitting is a lens's latest sitting strictly before the dispatch it is sitting for, on the
	// read it already does first. Present for every lens, with an explicit kind — first, behind,
	// unchanged or undispatched — never inferred from an absent field; absent for every other role.
	LastSitting *LastSittingJSON `json:"last_sitting,omitempty"`
}

// SittingOf computes what the given seat still owes on this record — the events carry every
// act, and the gap rows carry the view's answers (openness, the proof debt) so no fold decides
// them a second time (plans/board-as-views.md wave 1c). ids are the events' row ids, aligned with
// evs: the PASS gate's lens condition compares recorded pins with the report head, both row ids,
// and the chair's list states that condition from the same fold the gate refuses from.
func SittingOf(evs []*Event, ids []int64, gaps []WorkGapState, role, seatID string) SittingJSON {
	s := SittingJSON{Seat: seatID, Role: role, Open: []Item{}}
	if role == "lens" {
		ls := lastSittingBefore(evs, ids, seatID)
		s.LastSitting = &ls
	}
	add := func(what string) { s.Open = append(s.Open, Item{What: what, Blocks: true}) }

	// EVERY SEAT CLOSES THE LOG CHANNEL. Silence is not the empty case: an absent log reads the
	// same whether the sitting was clean or the channel went unused, and across eighteen recorded
	// sittings it was the second every time.
	//
	// ANY ENTRY DISCHARGES IT, and every surviving type asserts a problem, so the duty is to speak
	// when something is wrong rather than to speak at all.
	//
	// A SITTING THAT RECORDED NOTHING IS ALREADY THE CLEAN CASE, AND SAYING SO COSTS A COMMAND FOR
	// NO INFORMATION. The channel exists because silence is ambiguous — but that argument is about
	// a sitting that DID things and might have hit walls. A sitting with no acts at all is not
	// ambiguous: the hooks bracket it with sitting_open and sitting_close, both carrying the
	// agent's id and type, so that the seat ran is on the record whatever it did. Clean is derived
	// from what happened rather than asserted about it, which is the stronger of the two (#1089).
	//
	// Measured across eight runs: 48% of wakeups recorded nothing and still cost as much as the
	// productive ones — 46% of every command in the run.
	if !seatDidThisSitting(evs, seatID, recordpb.EventType_EVENT_TYPE_LOG) && !sittingRecordedNothing(evs, seatID) {
		add("the log is open — you have neither reported a missing capability nor said that nothing blocked you")
	}

	// EVERY DISPATCHED SEAT OWES THE SITTING IT WAS DISPATCHED FOR, and this list says so by the
	// predicate dispatch reads (sittingFor): a seat a dispatch names that has not registered since
	// has not sat. Dispatch enforces it — the lens's sitting is not counted, the bench has not sat for
	// the docketing, no exchange is counted for blue or the minting lens — so the seat is readied
	// again, epoch after epoch, until it registers. That covers every seat a dispatch names: the
	// lenses, blue-respond and the bench.
	if d, owed := owedSitting(evs, seatID); owed {
		add(fmt.Sprintf("you were dispatched against report head %d and have not registered since — this sitting is not on the record, and dispatch readies you again until it is; register for this sitting", d.pin))
	}
	// THE CHAIR OWES THE SAME REGISTER, AND NO DISPATCH NAMES IT, so the item above never reaches
	// it. A warm chair resuming its session registered once per run on every archived B run, and
	// read register's "first act at the seat" as a binding for its tenure; nothing on this list
	// said otherwise. Its sitting is known to have begun when a party has sat for its last
	// dispatch, and `dispatch next` — the chair's first act — refuses until the register lands,
	// which is the enforcement the rule at the top of this file asks of every item here.
	if seatID == chairSeat {
		if g, owed := unopenedChairSitting(evs); owed {
			add(chairRegisterOwed(g) + " — register for this sitting")
		}
	}

	switch role {
	case "blue":
		// A computation demand prose cannot answer. The chair is REFUSED if it tries to close
		// one unproved, so an unanswered demand does not settle — it carries into the next epoch.
		for _, g := range gaps {
			if !g.AwaitingProof {
				continue
			}
			id := g.ID
			// TWO HONEST ANSWERS, and the second is not a lesser one: there is no verb for
			// contesting a check_kind, so a seat that thinks the demand is wrong argues it in
			// the edit that answers the gap. Naming only `prove` would make disagreement look
			// like non-compliance.
			add("gap " + id + " was minted --check-kind computation and no proof answers it; prose cannot close it")
		}
		// THERE IS NO PER-LINE INQUIRY DUTY HERE ANY MORE, and the deletion is a ruling rather
		// than a dropped check. This arm read `UnsupportedInquiries` — the lines red had voted
		// `unsupported` or `absent`. Presence is not a question: the lines reach the report on
		// the worklist, generated from the record, so blue cannot cut them. Where blue's body
		// genuinely failed to deliver a line's research, red MINTS A GAP, and an open gap already
		// reaches blue through the ordinary route with a grade and a required fix — with the PASS
		// gate behind it when the gap is material by its class or grade. Restoring a second duty here would be the same fact told twice.
		if revisionOwed(evs, seatID) && !seatDidThisSitting(evs, seatID, recordpb.EventType_EVENT_TYPE_REVISION) {
			add("this sitting's revision is missing — a revision that is not on the record did not happen as far as the run is concerned (W1.7)")
		}
	case "chair":
		// EVERY REFUSAL THE GATE MAKES HAS AN ITEM HERE, AND NOTHING HERE BLOCKS WHAT THE GATE
		// ADMITS — so `complete` agrees with the gate. Each item is read from what its refusal
		// reads: the gap view's `stranded` and `material` columns, the motions, the inquiry read,
		// unansweredContradictions and the lens fold. The FAIL-only convergence refusal has no item:
		// this list claims nothing about a FAIL. B9's chair was told two gaps that did not hold the
		// gate refused PASS while dispatch said pass_permitted and the verdict accepted it; it settled
		// the contradiction by trying the verdict.
		for _, g := range gaps {
			if !g.Open {
				continue
			}
			switch {
			case g.Stranded:
				// requireSupersededAreClosed refuses EVERY verdict over it, whatever its grade.
				add("gap " + g.ID + " is open and superseded by " + g.SupersededBy + " — every verdict is refused while it is; close it with `--superseded-by`")
			case g.Material:
				add("gap " + g.ID + " is open and material — PASS is refused while it is")
			default:
				grade, _ := g.Severity.(string)
				s.Open = append(s.Open, Item{Blocks: false, What: "gap " + g.ID + " is open and not material (" +
					notMaterialBecause(g.ClassMaterial, grade) + ") — it does not hold PASS; your PASS lists it by class with why it changes no reader decision"})
			}
		}
		for _, claim := range unansweredContradictions(evs) {
			add("red read a source that contradicts or does not support the claim " + fmt.Sprintf("%q", claim) + " and no finding raises it — PASS is refused until one does")
		}
		for _, st := range passLensGateOf(evs, ids, freshMaterialOfStates(gaps)).statements() {
			add(st)
		}
		// THE VIEW NAMES THE GAVEL BECAUSE THE REFUSAL DOES. requirePassClosesAllMaterialGaps refuses
		// PASS over any unruled motion and says who rules each one; this list said only that the
		// motion stood. A chair seat reading it saw work it appeared to owe, and the item it
		// could not rule looked the same as the ones it could — which is the wedge the refusal's
		// message was rewritten to close, arriving on the other surface.
		//
		// WHO IS BLOCKED DOES NOT CHANGE HERE. An unruled petition still blocks a chair PASS: the
		// run is not finished until the bench answers it. What changes is that the seat is told
		// whose answer it is waiting for.
		for _, m := range MotionsOf(evs) {
			if m != nil && !m.Ruled() {
				phrase, err := rulerPhrase(m.Subject)
				if err != nil {
					// A SCHEMA DEFECT MUST NOT SILENTLY SHORTEN A WORK LIST. SittingOf has no
					// error to return, so the item is still reported and says what it could not
					// resolve — the alternative is dropping a blocking motion from the one list
					// a seat reads to find out what it owes.
					phrase = m.Subject + ", and this binary cannot say who rules it: " + err.Error()
				}
				add("motion " + m.ID + " (" + phrase + ") was filed and never ruled — PASS is refused while it stands")
			}
		}
		// THE LINES OF INQUIRY ARE READ ONCE, EVERY EPOCH.
		//
		// One statement per epoch, not one per line: presence is not the question, because the
		// lines are generated onto the page from the record. What the read owes is a judgement on
		// whether the BODY delivered them, and where it did not, a gap. The report is regenerated
		// each epoch, so a review recorded before this epoch's edits answers a question about a
		// document that no longer exists.
		if InquiryReviewDueOf(evs) {
			add("the report's account of its own research has not been read this epoch — PASS is refused until one `inquiry-support` says what the read found (and any shortfall is minted as a gap)")
		}
		if !seatDid(evs, seatID, recordpb.EventType_EVENT_TYPE_VERDICT) {
			add("your terminal act is missing — the run cannot say from its own record that it was ever verified")
		}
	// THE LENS HAS NO CASE, AND THAT IS THE RULE HOLDING RATHER THAN A GAP IN IT.
	//
	// A lens's blocking duties are the two every seat it shares a moment with holds, above this
	// switch: the log channel, and the sitting it was dispatched for. The second qualifies under
	// the rule at the top of this file because dispatch enforces it — an unregistered lens stays
	// ready. Nothing refuses a sitting over any other missing lens act, and the scorecard scores
	// no lens parity duty, so a lens arm here would be an invented obligation, and `complete:
	// false` on a seat no gate would hold is exactly the disagreement that teaches a seat to trust
	// neither surface.
	//
	// The acts a lens genuinely has open to it — verifying a citation nobody checked,
	// corroborating one blue never cited, re-running a proof nobody re-ran — come from availableOf
	// and land on this same list carrying Blocks:false. A lens that has registered and logged
	// therefore has an EMPTY blocking set and a non-empty work list, which is the accurate
	// statement: available work, none of it owed.
	case "bench":
		// THE SUBJECTS THE BENCH RULES, FROM THE SCHEMA — not the literal "petition" this
		// arm used to test. The gavel is an annotation on the MotionSubject enum, and a
		// hand-written subject name here is a fourth copy of it that goes stale the moment a
		// bench-ruled subject is added: the new subject's unruled motions would simply not appear
		// on the bench's list, which reads as a bench with nothing outstanding.
		for _, m := range MotionsOf(evs) {
			if m == nil || m.Ruled() {
				continue
			}
			subj, known := MotionSubjectEnum(m.Subject)
			if !known {
				continue
			}
			ruler, err := recordpb.SubjectRuler(subj)
			if err != nil || ruler != "bench" {
				continue
			}
			add(m.Subject + " " + m.ID + " is unruled, and the bench's motions are heard BEFORE the debate continues")
		}
	}
	// THE AFFORDANCES GO ON THE SAME LIST, and they go on it LAST so the blocking items read
	// first. They carry Blocks:false, so they are visible without being owed.
	s.Open = append(s.Open, availableOf(evs, gaps, role, seatID)...)
	s.Complete = !s.Blocked()
	return s
}

// Blocked reports whether anything open is stopping closure. It is the one place `complete` is
// derived from, so the field and the per-item flag cannot disagree.
func (s SittingJSON) Blocked() bool {
	for _, it := range s.Open {
		if it.Blocks {
			return true
		}
	}
	return false
}

// seatDid reports whether this seat recorded an event of this type in this run.
//
// The type is an EventType rather than a string: a caller that mistypes `"spot_check"` for
// `"spot-check"` used to get a silent false — a duty that reads as undischarged forever, or as
// discharged when it was not, depending on which side of the comparison drifted.
func seatDid(evs []*Event, seatID string, typ recordpb.EventType) bool {
	for _, e := range Live(evs) {
		if e.GetSeatId() == seatID && e.GetType() == typ {
			return true
		}
	}
	return false
}

// revisionOwed says whether this blue sitting owes a revision. Every blue sitting does, except a
// blue-respond sitting that found every gap it was dispatched on closed before it sat: it has
// nothing to answer, and the sitting the harness bracketed is the whole record of it (gblock's
// ruling). The sitting is record.BlueSittings' — the reading capture's record-parity audit
// holds it to — so the work list and the audit cannot disagree about it.
//
// IT HAS NO NOT-MEASURED ANSWER, AND NEEDS NONE. Whether a sitting owes is its Open set, which is
// fixed at blue's register — the dispatch and the closes before it — and never reads where the
// sitting ended. The latest sitting is unresolved exactly while blue is sitting it, which is when
// this list is read: an in-flight sitting still owes what it found open. Whether the owed revision
// was FILED is seatDidThisSitting's question, read from blue's latest register onward.
func revisionOwed(evs []*Event, seatID string) bool {
	if seatID != blueRespondSeat {
		return true
	}
	ss := BlueSittings(evs, WhileRunning) // the work list is read while blue sits
	if len(ss) == 0 {
		return true
	}
	return len(ss[len(ss)-1].Open) > 0
}

// seatDidThisSitting is seatDid for a duty the seat owes EVERY SITTING: only its acts since the
// register that OPENED the sitting its acts are attributed to count. seatDid reads the whole record,
// so the first sitting's act discharged every later one: in B9 blue sat in epoch 5 with G4 open,
// filed no revision, and read `complete` because its epoch-4 revision was on the record. A seat that
// has not registered has no earlier sitting to borrow from, so the whole record is its sitting.
//
// IT ATTRIBUTES, SO A REPAIR OPENS NO WINDOW (#1026). This is a reader of "which sitting does this
// act belong to", so the window opens at opensASitting, the one definition every attribution reader
// shares: a sitting-record repair is a turn of its own and its acts are the repaired sitting's,
// which is not what Clock counts. Starting the window at the seat's latest
// register of ANY kind put the repair's own register there, so inside a repair the work list said
// the log channel was open — a duty the repaired sitting had already discharged, and which
// scorecard.channel_closure, reading the same attribution, scored as discharged. The seat was told
// to file something it did not owe, on the one surface it reads to find out what it owes.
func seatDidThisSitting(evs []*Event, seatID string, typ recordpb.EventType) bool {
	live := Live(evs)
	start := 0
	for i, e := range live {
		if s, opens := recordpb.SeatOpeningSitting(e); opens && s == seatID {
			start = i
		}
	}
	for _, e := range live[start:] {
		if e.GetSeatId() == seatID && e.GetType() == typ {
			return true
		}
	}
	return false
}

// gapsAwaitingProofOn is gone: the gap rows carry awaiting_proof off the view, the same join
// the close gate reads, so the sitting and the gate cannot disagree about what is owed.

// sittingRecordedNothing reports whether the seat's current sitting holds no ACTS — nothing but the
// bookkeeping that brackets it.
//
// THE BOOKKEEPING IS NOT AN ACT. A register is the seat announcing itself, a log is the channel this
// predicate exists to excuse, and sitting_open/sitting_close are the hooks' own. Counting any of
// them would make every sitting look busy and the empty case unreachable.
//
// It reads the same window seatDidThisSitting does, so "this sitting" means one thing on this
// surface: from the event that opened the seat's latest sitting — its register, or the hook's
// sitting_open where the configuration names one seat — to the end of the record.
func sittingRecordedNothing(evs []*Event, seatID string) bool {
	live := Live(evs)
	start := 0
	opened := false
	for i, e := range live {
		if s, opens := recordpb.SeatOpeningSitting(e); opens && s == seatID {
			start, opened = i, true
		}
	}
	if !opened {
		// No sitting has opened for this seat at all, so there is no empty sitting to excuse —
		// and saying "nothing recorded" here would discharge a duty the seat has not reached.
		return false
	}
	for _, e := range live[start:] {
		if e.GetSeatId() != seatID {
			continue
		}
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_REGISTER,
			recordpb.EventType_EVENT_TYPE_LOG,
			recordpb.EventType_EVENT_TYPE_SITTING_OPEN,
			recordpb.EventType_EVENT_TYPE_SITTING_CLOSE:
			continue
		}
		return false
	}
	return true
}
