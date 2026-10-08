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
// Nothing here invents an obligation. Each one is refused at a write path (open material gaps and
// unruled motions block `verdict`; a computation gap cannot be closed on prose), is a stated
// epoch-record requirement (W1.7's revision, the bench's terminal outcome), or is enforced by
// dispatch (a seat whose sitting has not opened since the dispatch named it — no hook bracket and
// no register — is readied again; a chair that has not registered for its sitting is refused
// `dispatch next`). Inventing a duty here would make this view disagree with the gates, and a seat
// told it was finished by one surface and refused by another learns to trust neither.

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
func SittingOf(evs []*Event, ids []int64, win WindowIndex, gaps []WorkGapState, role, seatID string) SittingJSON {
	s := SittingJSON{Seat: seatID, Role: role, Open: []Item{}}
	if role == "lens" {
		ls := lastSittingBefore(evs, ids, win, seatID)
		s.LastSitting = &ls
	}
	add := func(what string) { s.Open = append(s.Open, Item{What: what, Blocks: true}) }
	may := func(what string) { s.Open = append(s.Open, Item{What: what, Blocks: false}) }

	// THE AUDIT IS THE LENS'S WORK, AND IT WAS THE ONE THING THIS LIST DID NOT SAY.
	//
	// A lens is dispatched to find what is wrong with the report. That duty is not a gap, so it was
	// in no item, so `open` was empty and `complete` was true for a lens whose report had just moved
	// 116 rows. Measured on the 2026-09-23 run: the voice lens read `complete: true, open: []`,
	// went to the report anyway, and minted two real defects. A seat that cannot trust the work list
	// reads the rest of the surface to find the work — 531 `show` calls across two runs, 331 of them
	// a lens, and the list's own counterparty line told it to go read the board.
	//
	// It is NON-BLOCKING on purpose: finding nothing is a legitimate audit, so this must not hold the
	// turn open the way an owed register does. Present-but-not-blocking is what the Item.Blocks field
	// is for, and the audit is the work it was added to be able to express.
	if s.LastSitting != nil {
		switch s.LastSitting.Kind {
		case "first":
			may("audit the report in full — this is your first sitting on it, so nothing of it is yet yours to have read")
		case "behind":
			// THE PINS ARE NAVIGATION, NOT THE AUDIT SURFACE, and this item said the opposite.
			//
			// It read "the change is your surface, not the whole document", which contradicts the
			// duty it was meant to serve in the words adversarial-audit already uses: "a
			// change-summary is a navigation hint, never the audit surface… This is not permission to
			// audit a fragment — when the report HAS moved, the full re-read stands exactly as
			// written." A lens asked about it had resolved the contradiction in the SKILL's favour
			// and read the report whole, which is the right outcome reached despite this text.
			//
			// So it states the fact and points at the duty, and the seat's constitution keeps saying
			// what the duty is. A work list that argues with a MUST teaches a seat to weigh the two.
			may(fmt.Sprintf("the report moved while you were away — you last sat at %d and it is at %d now, so re-read it in full; the pins are here to show you WHAT changed, not to narrow what you audit",
				s.LastSitting.Pin, s.LastSitting.Head))
		}
	}

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
	// THE SEAT'S OWN ENTRY IS WHAT THIS ASKS FOR. The tool now logs every refusal it gives a seat,
	// so "a log event this sitting" stopped meaning "the seat has spoken" — a refused call would have
	// closed the channel it exists to open. A sitting whose only events are refused calls recorded no
	// act and still met friction, so it is asked too.
	own, refused := logsThisSitting(evs, win, seatID)
	if own == 0 && (refused > 0 || !sittingRecordedNothing(evs, win, seatID)) {
		// THE ITEM NAMES ONLY WHAT CAN BE FILED. It used to end "nor said that nothing blocked you",
		// which no type can say: `nominal` was retired when clean became DERIVED from having sat and
		// filed nothing (LogType's own comment carries the measurement — 40 of 40 and 42 of 42 entries
		// were nominal and carried nothing to act on). A seat reads this list as its duties, so the
		// half it could not discharge got filed under a word that means something else: 8 of 28 entries
		// in universe-m12 and 6 of 35 in m11 assert that nothing was wrong, six of them as `friction`
		// and one as `defect` — in the one field the operator triages the channel by.
		// NOT BLOCKING, and that is the fix rather than the wording. This item used `add`, so a seat
		// that recorded acts and hit nothing could not reach `complete` without filing an entry —
		// while the item said "nothing to report needs no entry". The mechanism overruled the text,
		// and seats padded the log to finish: on universe-m12 audit summaries and arguments about
		// other lenses' gaps, in the channel whose glossary says none of it is debate material.
		// FRICTION, NOT ONLY BLOCKAGE. The item said "a missing capability, a defect in the tooling or
		// an impediment", and on universe-m13 every seat read that as "something that stopped me" and
		// filed nothing across a dozen refusals, guessed flags and workarounds (seven interviews).
		if refused > 0 {
			may(fmt.Sprintf("the log is open, and the tool has already recorded %d refused or failed call(s) of yours this sitting — your entry adds what only you know: what you expected, and where the expectation came from. Any other friction goes there too: a workaround, an act you set aside", refused))
		} else {
			may("the log is open — for anything that cost you a call, a guess or an act. A sitting where nothing cost you anything files nothing")
		}
	}

	// EVERY DISPATCHED SEAT OWES THE SITTING IT WAS DISPATCHED FOR, and this list says so by the
	// predicate dispatch reads (sittingFor): a seat a dispatch names whose sitting has not opened
	// since — no hook bracket, no register — has not sat. Dispatch enforces it — the lens's sitting
	// is not counted, the bench has not sat for the docketing, no exchange is counted for blue or
	// the minting lens — so the seat is readied again, epoch after epoch, until its sitting opens.
	// That covers every seat a dispatch names: the lenses, blue-respond and the bench.
	if d, owed := owedSitting(evs, win, seatID); owed {
		add(fmt.Sprintf("you were dispatched against report head %d and have not registered since — this sitting is not on the record, and dispatch readies you again until it is; register for this sitting", d.pin))
	}
	// THE CHAIR OWES THE SAME REGISTER, AND NO DISPATCH NAMES IT, so the item above never reaches
	// it. A warm chair resuming its session registered once per run on every archived B run, and
	// read register's "first act at the seat" as a binding for its tenure; nothing on this list
	// said otherwise. Its sitting is known to have begun when a party has sat for its last
	// dispatch, and `dispatch next` — the chair's first act — refuses until the register lands,
	// which is the enforcement the rule at the top of this file asks of every item here.
	if seatID == chairSeat {
		if g, owed := unopenedChairSitting(evs, win); owed {
			add(chairRegisterOwed(g) + " — register for this sitting")
		}
	}

	switch role {
	case "blue":
		// A computation demand prose cannot answer. The lens that minted it is REFUSED if it
		// tries to close one unproved, so an unanswered demand does not settle — it carries into
		// the next epoch.
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
		// THERE IS NO PER-LINE AVENUE DUTY HERE ANY MORE, and the deletion is a ruling rather
		// than a dropped check. This arm read `UnsupportedAvenues` — the lines red had voted
		// `unsupported` or `absent`. Presence is not a question: the lines reach the report on
		// the worklist, generated from the record, so blue cannot cut them. Where blue's body
		// genuinely failed to deliver a line's research, red MINTS A GAP, and an open gap already
		// reaches blue through the ordinary route with a grade and a required fix — with the PASS
		// gate behind it when the gap is material by its class or grade. Restoring a second duty here would be the same fact told twice.
		if revisionOwed(evs, win, seatID) && !seatDidThisSitting(evs, win, seatID, recordpb.EventType_EVENT_TYPE_REVISION) {
			add("this sitting's revision is missing — a revision that is not on the record did not happen as far as the run is concerned (W1.7)")
		}
	case "chair":
		// THE GATE'S BLOCKERS ARE THE BLOCKING ITEMS HERE, from the one list the gate refuses on
		// (passBlockersOf), so `complete` agrees with the gate and nothing here blocks what the gate
		// admits. Each item names the seat whose act clears it: an unruled petition still blocks a
		// chair PASS, and the chair is told whose answer it is waiting for. The FAIL-only
		// convergence refusal has no item: this list claims nothing about a FAIL.
		//
		// A gap's item stands in the gap's place on the board, and an open gap that holds nothing
		// is listed there too, as work that is not owed.
		blockers := passBlockersOf(evs, ids, win, blockerGapsOfStates(gaps), freshMaterialOfStates(gaps))
		gapItem := map[string]string{}
		for _, b := range blockers {
			if b.Kind == BlockerStrandedGap || b.Kind == BlockerMaterialGap {
				gapItem[b.Subject] = b.WorkItem()
			}
		}
		for _, g := range gaps {
			if !g.Open {
				continue
			}
			if it, holds := gapItem[g.ID]; holds {
				add(it)
				continue
			}
			grade, _ := g.Severity.(string)
			s.Open = append(s.Open, Item{Blocks: false, What: "gap " + g.ID + " is open and not material (" +
				notMaterialBecause(g.ClassMaterial, grade) + ") — it does not hold PASS; your PASS lists it by class with why it changes no reader decision"})
		}
		for _, b := range blockers {
			if b.Kind != BlockerStrandedGap && b.Kind != BlockerMaterialGap {
				add(b.WorkItem())
			}
		}
		if !seatDid(evs, seatID, recordpb.EventType_EVENT_TYPE_VERDICT) {
			add("your terminal act is missing — the run cannot say from its own record that it was ever verified")
		}
	// THE LENS OWES ONE THING OF ITS OWN: RAISING A CONTRADICTION IT READ. Beyond the two duties
	// every seat holds above this switch (the log channel, and the sitting it was dispatched for),
	// a contradicting source the lens verified and no finding raises is a PASS blocker whose owner
	// is that lens (passBlockersOf), and dispatch readies the lens for it once — so it qualifies
	// under the rule at the top of this file, and it is the kind table's own sentence.
	//
	// Nothing else blocks a lens. The acts a lens genuinely has open to it — verifying a citation
	// nobody checked, corroborating one blue never cited, re-running a proof nobody re-ran — come
	// from availableOf and land on this same list carrying Blocks:false: available work, none of
	// it owed.
	case "lens":
		for _, b := range passBlockersOf(evs, ids, win, blockerGapsOfStates(gaps), freshMaterialOfStates(gaps)) {
			if b.Kind == BlockerContradiction && b.Owner == seatID {
				add(b.WorkItem())
			}
		}
	case "bench":
		// THE BENCH'S MOTIONS ARE THE GATE'S LIST, read through passBlockersOf: every unruled motion
		// whose gavel the schema gives the bench, the same set the outcome refusal names
		// (requireBenchMotionsRuled). A second walk of the motions here could disagree with that
		// refusal, which is the drift the one list exists to end (#1202).
		for _, b := range passBlockersOf(evs, ids, win, blockerGapsOfStates(gaps), freshMaterialOfStates(gaps)) {
			if b.Kind == BlockerUnruledMotion && b.Owner == benchSeat {
				add("motion " + b.Subject + " (" + b.Detail + ") is unruled and the gavel is yours — the run cannot end in a verdict while it stands")
			}
		}
	}
	// THE AFFORDANCES GO ON THE SAME LIST, and they go on it LAST so the blocking items read
	// first. They carry Blocks:false, so they are visible without being owed.
	s.Open = append(s.Open, availableOf(evs, win, gaps, role, seatID)...)

	// AN EMPTY LIST MUST SAY IT IS EMPTY. `complete: true, open: []` is the answer a seat with
	// nothing to do gets, and it is indistinguishable from a command that failed to find anything —
	// so the seat assumes the work is somewhere else and goes looking for it. Measured across two
	// runs: 531 `show` calls, 331 of them a lens, against a work list that had already told several
	// of them they were complete.
	//
	// The statement is an ITEM rather than a field because `open` is what a seat reads; a seat that
	// checks one place and finds a sentence has been answered, where a seat that finds `[]` has not.
	// It is non-blocking, because having nothing to do is not an obligation.
	if len(s.Open) == 0 {
		may("nothing is owed and nothing is available: this list is complete, not empty — no other projection holds work for you, and ending your turn now is the correct act")
	}
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
// nothing to answer, and the sitting its register opened is the whole record of it (gblock's
// ruling). The sitting is record.BlueSittings' — the reading capture's record-parity audit
// holds it to — so the work list and the audit cannot disagree about it.
//
// IT HAS NO NOT-MEASURED ANSWER, AND NEEDS NONE. Whether a sitting owes is its Open set, which is
// fixed at blue's register — the dispatch and the closes before it — and never reads where the
// sitting ended. The latest sitting is unresolved exactly while blue is sitting it, which is when
// this list is read: an in-flight sitting still owes what it found open. Whether the owed revision
// was FILED is seatDidThisSitting's question, read off blue's latest sitting.
func revisionOwed(evs []*Event, win WindowIndex, seatID string) bool {
	if seatID != blueRespondSeat {
		return true
	}
	ss := BlueSittings(evs, win, WhileRunning) // the work list is read while blue sits
	if len(ss) == 0 {
		return true
	}
	return len(ss[len(ss)-1].Open) > 0
}

// seatDidThisSitting is seatDid for a duty the seat owes EVERY SITTING: only its acts in the sitting
// the write path stored them in count (thisSitting). seatDid reads the whole record,
// so the first sitting's act discharged every later one: in B9 blue sat in epoch 5 with G4 open,
// filed no revision, and read `complete` because its epoch-4 revision was on the record. A seat that
// has never sat has no earlier sitting to borrow from, so the whole record is its sitting.
//
// IT ATTRIBUTES, SO A REPAIR OPENS NO WINDOW (#1026). This is a reader of "which sitting does this
// act belong to", so it reads the stored sitting, the one definition every attribution reader
// shares: the write path stores a sitting-record repair, and its acts, in the sitting it repairs.
// Starting the window at the seat's latest register of ANY kind put the repair's own register there, so inside a repair the work list said
// the log channel was open for a sitting that had already filed there. The seat was told to file
// something it had done, on the one surface it reads to find out what is left.
func seatDidThisSitting(evs []*Event, win WindowIndex, seatID string, typ recordpb.EventType) bool {
	for _, e := range thisSitting(evs, win, seatID) {
		if e.GetType() == typ {
			return true
		}
	}
	return false
}

// thisSitting is the seat's own live events in the window seatDidThisSitting reads: the acts the
// write path stored in the seat's latest sitting. That sitting opened at the hook's bracket or at a
// register, and a register under the bracketed agent joined it — so a log filed between the
// bracket and that register is in the sitting, as the record says. A seat that has never sat has
// no earlier sitting to borrow from: its acts are in none, and that is its window.
func thisSitting(evs []*Event, win WindowIndex, seatID string) []*Event {
	latest := win.LatestSittingOf(seatID)
	var out []*Event
	for _, e := range Live(evs) {
		if e.GetSeatId() == seatID && win.Of(e).SittingID == latest {
			out = append(out, e)
		}
	}
	return out
}

// logsThisSitting splits the sitting's log entries by who wrote them: the seat's own, and the
// refusals and failed calls the TOOL recorded against it. They answer different questions — whether the seat has
// spoken, and whether it met friction the tool could see — so neither may stand in for the other.
func logsThisSitting(evs []*Event, win WindowIndex, seatID string) (seatEntries, toolRefusals int) {
	for _, e := range thisSitting(evs, win, seatID) {
		l, ok := recordpb.BodyAs[*recordpb.Log](e)
		if !ok {
			continue
		}
		switch {
		case l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL &&
			(l.GetType() == recordpb.LogType_LOG_TYPE_REFUSAL || l.GetType() == recordpb.LogType_LOG_TYPE_FAILURE):
			toolRefusals++
		case l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL:
			seatEntries++
		}
	}
	return seatEntries, toolRefusals
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
// It reads the same window seatDidThisSitting does (thisSitting), so "this sitting" means one thing
// on this surface: the seat's acts the write path stored in its latest sitting — opened by its
// register, or by the hook's sitting_open where the configuration names one seat.
func sittingRecordedNothing(evs []*Event, win WindowIndex, seatID string) bool {
	if win.LatestSittingOf(seatID) == 0 {
		// No sitting has opened for this seat at all, so there is no empty sitting to excuse —
		// and saying "nothing recorded" here would discharge a duty the seat has not reached.
		return false
	}
	for _, e := range thisSitting(evs, win, seatID) {
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
