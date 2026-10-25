package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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
// record requirement (blue-respond's position, the bench's terminal outcome), or is enforced by
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
	// check could only ever see the mechanically-enforced half, so a petition, a
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
	// AnotherSeatClears marks a blocking item this seat's own act cannot clear: it is on the list
	// because it holds what the seat is working toward, and it is the board's state rather than
	// the seat's debt. Only the chair's list carries any — the PASS gate's blockers whose owner is
	// not the chair (Blocker.ChairOwned): an open material gap its minter closes, a ready lens, a
	// motion whose gavel is another seat's. Every other seat's blocking items are its own.
	AnotherSeatClears bool `json:"another_seat_clears,omitempty"`
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
	// LastRead is a lens's pin before the dispatch it is sitting for — the head the whole report was
	// last served to it at (lastReadBefore) — on the read it already does first. Present for every
	// lens, with an explicit kind — first, behind, unchanged or undispatched — never inferred from an
	// absent field; absent for every other role.
	LastRead *LastReadJSON `json:"last_read,omitempty"`
	// MintBudget is a lens's standing against its mint budget — what it holds open, its budget
	// with every term, and what is left — from the reading the write refuses a mint on
	// (lensMintStanding). Present for every lens whose standing reads; absent for every other
	// role, which spends no budget. An item of Open states the same numbers in a sentence.
	MintBudget *MintBudgetJSON `json:"mint_budget,omitempty"`
}

// SittingOf computes what the given seat still owes on this record — the events carry every
// act, and the gap rows carry the view's answers (openness, the proof debt) so no fold decides
// them a second time (plans/board-as-views.md wave 1c). ids are the events' row ids, aligned with
// evs: the PASS gate's lens condition compares recorded pins with the report head, both row ids,
// and the chair's list states that condition from the same fold the gate refuses from.
func SittingOf(evs []*Event, ids []int64, win WindowIndex, gaps []WorkGapState, role, seatID string) SittingJSON {
	return sittingOf(evs, ids, win, gaps, role, seatID, sittingReads{})
}

// sittingReads is what a seat's list needs beyond the stream and the gap rows to say whether its
// sitting can clear an item; a zero value reads every item as clearable.
type sittingReads struct {
	// unverifiable: why no verify can reach a citation (unverifiableIn).
	unverifiable unverifiableFunc
	// mintsLeft: a lens's mints left on its budget, and whether that reads (mintBudgetOf).
	mintsLeft func() (int, bool)
}

// unverifiableFunc says why a citation's sentence cannot be verified from the report as it renders
// now — the reasons `verify` refuses every outcome on — or "" where it can. nil reads every citation
// as verifiable.
type unverifiableFunc func(anchor string) string

// sittingOf is SittingOf with the reads WorkOfSeat supplies (sittingReads): an item those reads say
// the sitting cannot clear is listed, and does not block.
func sittingOf(evs []*Event, ids []int64, win WindowIndex, gaps []WorkGapState, role, seatID string, reads sittingReads) SittingJSON {
	s := SittingJSON{Seat: seatID, Role: role, Open: []Item{}}
	if role == "lens" {
		ls := lastReadBefore(evs, ids, win, seatID)
		s.LastRead = &ls
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
	if s.LastRead != nil {
		switch s.LastRead.Kind {
		case "first":
			may("audit the report in full — no sitting of yours before this one was served it whole, so nothing of it is yet yours to have read. " + ServedPinNote)
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
			//
			// IT SPEAKS OF BEFORE THIS SITTING, because that is what lastReadBefore measures: a whole read
			// in this sitting does not move it, and "it is at H now, re-read it" stayed on the list after
			// the read it asked for (#1565 smoke: 14 work-list reads in 12 sittings).
			may(fmt.Sprintf("before this sitting, the report was last served to you whole at %d, and you are dispatched against head %d, so this sitting's audit reads it in full; the pins are here to show you WHAT changed, not to narrow what you audit. %s",
				s.LastRead.Pin, s.LastRead.Head, ServedPinNote))
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
	// THE BENCH'S ROW CAN CONVENE IT TWICE — a petition and a docket, petition first — and its
	// sitting already registered for one occasion owes the other's register to the NEXT sitting: a
	// second register in this one is refused as a repeat, so this sitting cannot clear it.
	if d, owed := owedSitting(evs, win, seatID); owed && seatID == benchSeat && benchRegisteredSince(evs, win, d) {
		may(fmt.Sprintf("the dispatch against report head %d also convenes you for a sitting of its own you have not opened yet — this sitting is registered for its occasion, and the next is registered for the other", d.pin))
	} else if owed {
		add(fmt.Sprintf("you were dispatched against report head %d and no sitting of yours has opened since — this sitting is not on the record, and dispatch readies you again until it is; register for this sitting", d.pin))
	}
	// THE CHAIR OWES THE SAME REGISTER, AND NO DISPATCH NAMES IT, so the item above never reaches
	// it. A warm chair resuming its session registered once per run on every archived B run, and
	// read register's "first act at the seat" as a binding for its tenure; nothing on this list
	// said otherwise. Its sitting is known to have begun when a party has sat for its last
	// dispatch; where the hook's bracket opened it nothing is owed, and where none did `dispatch
	// next` — the chair's first act — refuses until the register lands, which is the enforcement
	// the rule at the top of this file asks of every item here.
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
		//
		// IT HOLDS THE SITTING ONLY WHERE THIS SITTING IS ENGAGED ON THE GAP, and until a recorded
		// act answers it there (gblock, 2026-10-24): a proof naming it, which ends the debt; a motion
		// blue files on the gap, which puts the demand before its ruler; or a retire that takes the
		// gap's sentence out. Prose does not: an edit arguing the demand is wrong leaves it owed.
		var engaged []string
		if e := engagedOf(evs, win, gaps, seatID); e != nil {
			engaged = e.GapIDs
		}
		moved, retired := answeredByActThisSitting(evs, win, seatID)
		for _, g := range gaps {
			if !g.AwaitingProof {
				continue
			}
			id := g.ID
			const what = " was minted --check-kind computation and no proof answers it; prose cannot close it"
			switch {
			case !slices.Contains(engaged, id):
				may("gap " + id + what + " — this sitting is not engaged on it")
			case moved[id]:
				may("gap " + id + what + " — your motion this sitting on it puts the demand before its ruler, and the lens's close stays refused until a proof names it")
			case retired && g.LocationState == LocationGone:
				may("gap " + id + what + " — your retire this sitting took its sentence out, and the lens's close stays refused until a proof names it")
			default:
				add("gap " + id + what)
			}
		}
		// A REMAND'S EXCHANGE HOLDS BLUE'S SITTING UNTIL BLUE ACTS ON THE GAP (remandItemsOf).
		if seatID == blueRespondSeat {
			s.Open = append(s.Open, remandItemsOf(evs, win, gaps, seatID, true)...)
		}
		// THERE IS NO PER-INQUIRY DUTY HERE, and the absence is a ruling rather than a dropped
		// check. Presence is not a question: the inquiries reach the report on the worklist,
		// generated from the record, so blue cannot cut them. Where blue's body genuinely failed
		// to deliver an inquiry's research, red MINTS A GAP, and an open gap already reaches blue
		// through the ordinary route with a grade and a required fix — with the PASS gate behind
		// it when the gap is material by its class or grade. A second duty here would be the same
		// fact told twice.
		// A GAP THIS SITTING FOUND CLOSED WHOSE SUCCESSOR IS OPEN OWES AN `engage` (successorsOwed):
		// the record names the successor, and blue sits on it a second time this epoch.
		if seatID == blueRespondSeat {
			if owed, _ := successorsOwed(evs, win, openOf(gaps)); len(owed) > 0 {
				var named []string
				for _, o := range owed {
					named = append(named, o.GapID+" (successor of "+o.Of+")")
				}
				add("this sitting found a gap closed whose successor is open and not yet engaged this epoch — " + strings.Join(named, ", ") +
					" — record `engage` on them, then name them in your envelope's successors: you sit on them again this epoch")
			}
		}
		// BLUE-RESPOND OWES ITS CLOSING ON EACH DOCKETED GAP (closingowed.go): the PASS gate's
		// blocker whose owner it is, the reading capture's record-parity audit holds it to, so this
		// list and that audit cannot disagree about which closing is short. A lane, the frontier and
		// the synthesizer are party to no dispute and owe none.
		//
		// IT OWES THE FIX OR RETIRE OF A PROOF A LENS'S STANDING JUDGEMENT CALLS UNSOUND (unsoundproof.go):
		// the PASS gate's blocker, cleared by blue's retire, or by the fresh sound judgement blue's edit
		// of the sentence reopens it for.
		for _, b := range passBlockersOf(evs, ids, win, blockerGapsOfStates(gaps), freshMaterialOfStates(gaps)) {
			switch {
			case b.Kind == BlockerClosingOwed && b.Owner == seatID:
				add(b.WorkItem())
			case b.Kind == BlockerUnsoundProof && seatID == blueRespondSeat:
				add(blueUnsoundItem(b))
			}
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
		// EACH ITEM SAYS WHETHER THE CHAIR CLEARS IT, by the gate's own predicate (ChairOwned), so
		// the digest counts as the chair's debt only what the chair can clear.
		hold := func(b Blocker) {
			s.Open = append(s.Open, Item{What: b.WorkItem(), Blocks: true, AnotherSeatClears: !b.ChairOwned()})
		}
		gapItem := map[string]Blocker{}
		for _, b := range blockers {
			if b.Kind == BlockerStrandedGap || b.Kind == BlockerMaterialGap || b.Kind == BlockerRemandOwed {
				gapItem[b.Subject] = b
			}
		}
		for _, g := range gaps {
			if !g.Open {
				continue
			}
			if b, holds := gapItem[g.ID]; holds {
				hold(b)
				continue
			}
			grade, _ := g.Severity.(string)
			s.Open = append(s.Open, Item{Blocks: false, What: "gap " + g.ID + " is open and not material (" +
				notMaterialBecause(g.ClassMaterial, grade) + ") — it does not hold PASS; your PASS lists it by class with why it changes no reader decision"})
		}
		for _, b := range blockers {
			if b.Kind != BlockerStrandedGap && b.Kind != BlockerMaterialGap && b.Kind != BlockerRemandOwed {
				hold(b)
			}
		}
		if !seatDid(evs, seatID, recordpb.EventType_EVENT_TYPE_VERDICT) {
			add("your terminal act is missing — the run cannot say from its own record that it was ever verified")
		}
	// THE LENS OWES THE PASS BLOCKERS IT OWNS: RAISING A CONTRADICTION IT READ, CHECKING A
	// REOPENED CITATION OR PROOF IT OWNS (anchorOwnerOf), AND ITS CLOSING ON A GAP OF ITS OWN THAT
	// IS DOCKETED (closingowed.go). Beyond the two duties every seat holds above
	// this switch (the log channel, and the sitting it was dispatched for), each is a PASS blocker
	// whose owner is that lens (passBlockersOf), and dispatch readies the lens for it — so it
	// qualifies under the rule at the top of this file, and it is the kind table's own sentence.
	//
	// A REOPENED ITEM IS LISTED ONLY IN A SITTING THAT OPENED AFTER ITS EDIT. A sitting already open
	// when the edit landed cannot clear the item (reopenedEvidence), so holding that sitting open
	// on it would hold it on work it cannot do; the next plan readies the lens, and that sitting
	// lists it.
	//
	// Nothing else blocks a lens. The acts a lens genuinely has open to it — corroborating a source
	// blue never cited and, for the lens that owns the anchor (anchorOwnerOf), verifying a citation
	// nobody checked or re-running a proof nobody re-ran — come from availableOf and land on this same list carrying Blocks:false: available work, none of
	// it owed.
	case "lens":
		// A REMAND'S EXCHANGE HOLDS ITS MINTING LENS'S SITTING UNTIL THE LENS ACTS ON THE GAP
		// (remandItemsOf): the other seat a remand readies, and the only one that may close the gap.
		s.Open = append(s.Open, remandItemsOf(evs, win, gaps, seatID, false)...)
		for _, b := range passBlockersOf(evs, ids, win, blockerGapsOfStates(gaps), freshMaterialOfStates(gaps)) {
			// A PROOF IT JUDGED UNSOUND IS ITS DEBT UNTIL ITS GAP IS MINTED; from then the item stays on
			// the list as not owed — blue's fix or retire clears it, or the lens's own sound judgement
			// if it judges the sentence again.
			if b.Kind == BlockerUnsoundProof && b.Owner == seatID {
				if b.answeredBy == "" {
					add(b.WorkItem())
				} else {
					may(b.WorkItem())
				}
				continue
			}
			if (b.Kind != BlockerContradiction && b.Kind != BlockerReopened && b.Kind != BlockerClosingOwed) || b.Owner != seatID {
				continue
			}
			if b.Kind == BlockerReopened && win.LatestSittingOf(seatID) <= b.since {
				continue
			}
			// A CITATION THIS SITTING READ AS UNREACHABLE IS NOT ITS TO CLEAR: its verify read nothing,
			// so it stands reopened, and no act of this sitting's reads it. A later sitting's verify,
			// or blue's retire of the citation, clears it; the PASS stays held meanwhile.
			if b.Kind == BlockerReopened && unreachableThisSitting(evs, win, seatID)[b.Subject] {
				may(b.WorkItem() + " — your verify this sitting found its source unreachable, so this sitting cannot clear it")
				continue
			}
			// A CITATION NO VERIFY CAN REACH is not its to clear either: `verify` refuses every
			// outcome on it (cli/lens/verify.go citedSentence), and blue's retire of its claim, or a
			// report that renders again, is what moves it.
			if b.Kind == BlockerReopened && b.Detail == "" && reads.unverifiable != nil {
				if why := reads.unverifiable(b.Subject); why != "" {
					may(b.WorkItem() + " — " + why + ", so no verify of this sitting can clear it")
					continue
				}
			}
			// A CONTRADICTION ANOTHER LENS RAISED, AT THE READER'S MINT CAP, is not its to clear: the
			// gap it mints from another seat's finding takes one of its mints (answersOwnRaiseAt), and
			// it has none left. The raiser's own mint from that finding takes none.
			if b.Kind == BlockerContradiction && b.raiser != "" && b.raiser != seatID && reads.mintsLeft != nil {
				if left, read := reads.mintsLeft(); read && left <= 0 {
					may(b.WorkItem() + " — finding " + b.raisedBy + " that raised it is " + b.raiser + "'s, so a gap you mint from it takes a mint and you have none left; " + b.raiser + "'s own mint from it takes none, so this sitting cannot clear it")
					continue
				}
			}
			add(b.WorkItem())
		}
	case "bench":
		// THE BENCH'S MOTIONS ARE THE GATE'S LIST, read through passBlockersOf: every unruled motion
		// whose gavel the schema gives the bench, the same set the outcome refusal names
		// (requireBenchMotionsRuled). A second walk of the motions here could disagree with that
		// refusal, which is the drift the one list exists to end (#1202).
		//
		// ONLY WHAT THIS SITTING IS CONVENED TO RULE BLOCKS (benchHeld): a blocking item refuses the
		// sitting's envelope (EnvelopeRefusal), and a motion another sitting is convened for — a docket
		// on a gap whose closings this sitting was not convened on, a petition filed while it sat — is
		// not this sitting's to rule. It stays on the list, holding the outcome.
		held := benchHeld(evs, ids, win)
		// A REMAND PAST THE LIMIT IS REFUSED (requireRemandWithinLimit), and the motion's item says so
		// before the bench rules, by the write's own predicate (remandPastLimit, WorkGapState.remandLimit).
		limited := map[string]int64{}
		for _, g := range gaps {
			if g.Open && g.remandLimit > 0 {
				limited[g.ID] = g.remandLimit
			}
		}
		for _, b := range passBlockersOf(evs, ids, win, blockerGapsOfStates(gaps), freshMaterialOfStates(gaps)) {
			if b.Kind != BlockerUnruledMotion || b.Owner != benchSeat {
				continue
			}
			limit := ""
			if at, ok := limited[b.About]; ok && b.motionSubject == "docket" {
				limit = fmt.Sprintf(" — a remand of %s is refused: you remanded it at impasse at event %d, and %s; rule a disposition that closes it (%s)",
					b.About, at, RemandLimitRule, strings.Join(ClosingDispositions(), ", "))
			}
			if held[b.Subject] {
				add("motion " + b.Subject + " (" + b.Detail + ") is unruled, the gavel is yours and this sitting is convened to rule it — the run cannot end in a verdict while it stands" + limit)
				continue
			}
			may("motion " + b.Subject + " (" + b.Detail + ") is unruled and the gavel is yours; it waits for the sitting convened for it — the run cannot end in a verdict while it stands" + limit)
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
	//
	// IT IS SAID TO A SEAT WHOSE WORK ARRIVES AS ITEMS, AND TO NO OTHER. The frontier, a lane and
	// the synthesizer — every blue seat but the responding one — are dispatched to write: the
	// inquiries, a draft, the report. No plan names them and nothing they owe is ever an item, so an
	// empty list is what theirs always is, and "ending your turn now is the correct act" tells such
	// a seat to stop before it has started. universe-m21's synthesizer quoted "no other projection
	// holds work for you" as its only evidence about the record, and proposed five approaches the
	// record already held.
	if len(s.Open) == 0 && WorkArrivesAsItems(role, seatID) {
		may("nothing is owed and nothing is available: this list is complete, not empty — no other projection holds work for you, and ending your turn now is the correct act")
	}
	s.Complete = !s.Blocked()
	return s
}

// Owed is the blocking items this seat's own act clears, and Board the blocking items another
// seat must clear, each in list order. Together they are every blocking item, so a reader that
// counts a seat's debt counts Owed and states Board as what the board holds.
func (s SittingJSON) Owed() (owed, board []Item) {
	for _, it := range s.Open {
		switch {
		case !it.Blocks:
		case it.AnotherSeatClears:
			board = append(board, it)
		default:
			owed = append(owed, it)
		}
	}
	return owed, board
}

// HoldsPass reports whether this seat's blocking items hold a PASS rather than the end of its
// turn. The chair's do: its blocking items are the PASS gate's blockers and its verdict, and
// nothing else (the chair arm of SittingOf), so a chair sitting with a PASS out of view ends with
// them open. Every other seat's blocking items are owed in the sitting that lists them.
func (s SittingJSON) HoldsPass() bool { return s.Role == "chair" }

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
	for _, e := range evs {
		if e.GetSeatId() == seatID && e.GetType() == typ {
			return true
		}
	}
	return false
}

// sittingOfBlueRespond is the sitting blue-respond's work list is read in: the latest of
// record.BlueSittings, read while the run is running. sat is false for every other seat, and for a
// blue-respond the record holds no dispatched sitting of. Once blue has registered for its dispatch,
// the gaps the list names as found closed (engagedOf) are read off it — the sitting the manifest's
// owed set reads — so the list cannot excuse a row over one set of gaps and name another.
func sittingOfBlueRespond(evs []*Event, win WindowIndex, seatID string) (BlueSitting, bool) {
	if seatID != blueRespondSeat {
		return BlueSitting{}, false
	}
	ss := BlueSittings(evs, win, WhileRunning) // the work list is read while blue sits
	if len(ss) == 0 {
		return BlueSitting{}, false
	}
	return ss[len(ss)-1], true
}

// EngagedJSON is what the dispatch blue-respond is to sit for engaged it on — the work list's
// `engaged` (WorkJSON.Engaged).
type EngagedJSON struct {
	// GapIDs are the gaps the dispatch named, in its order.
	GapIDs []string `json:"gap_ids"`
	// FoundClosed is each of those a close landed on after the dispatch and before this sitting's
	// register — its lens sat first and closed it — or, where the sitting has not registered yet,
	// before this read. It is what the sitting's envelope names as found closed. Never omitted: an empty list
	// is "every engaged gap is open".
	FoundClosed []FoundClosedJSON `json:"found_closed"`
	// SuccessorsOwed is the successors this sitting owes an `engage` on (successorsOwed): each open
	// successor of a gap it found closed that no dispatch row of this epoch engages blue on, in
	// found-closed order, once each. It is exactly the set the act must name, and it is empty when
	// none is owed and once the act is recorded. Never omitted.
	SuccessorsOwed []string `json:"successors_owed"`
}

// FoundClosedJSON is one engaged gap blue found closed: the seat whose close it was, the fate
// that close gave it (the Disposition vocabulary, as `estopped` carries it), and the open gaps the
// record names as its successors (successorsOf).
type FoundClosedJSON struct {
	ID           string `json:"id"`
	ClosedBySeat string `json:"closed_by_seat"`
	Fate         string `json:"fate"`
	// Successors is each open gap the record names as carrying this one forward — its close's
	// `superseded_by`, or a mint naming it in `supersedes` — in record order. Never omitted: an
	// empty list is "the record names no open successor", and blue sits on nothing more for it.
	Successors []string `json:"successors"`
}

// engagedOf is the engaged block of the dispatch blue-respond is to sit for, or nil where no
// dispatch names it.
//
// IT IS KEYED ON THE DISPATCH, NOT ON THE REGISTER. Blue reads its work list with its dispatch,
// before it registers, and a block read off its latest registered sitting names the dispatch
// before: universe-wave-b's blue was dispatched on one set of gaps and read another in `engaged` in
// four epochs running, and read the whole board to choose between them. So a dispatch the seat has
// not sat for (owedSitting — the reading the "you have not registered" item on this list comes
// from) is the block, split by the closes that have landed since its row (engagementOf); once the
// seat registers, the block is the sitting's own (sittingOfBlueRespond), which is that same split
// read up to the register.
func engagedOf(evs []*Event, win WindowIndex, gaps []WorkGapState, seatID string) *EngagedJSON {
	if seatID != blueRespondSeat {
		return nil
	}
	live := evs
	s, sat := sittingOfBlueRespond(live, win, seatID)
	if d, owed := owedSitting(live, win, seatID); owed {
		s, sat = engagementOf(live, d, int64(len(live))), true
	}
	if !sat {
		return nil
	}
	e := &EngagedJSON{GapIDs: strs(s.Engaged), FoundClosed: []FoundClosedJSON{}, SuccessorsOwed: []string{}}
	succ, open := successorsOf(live), openOf(gaps)
	owed, _ := successorsOwed(live, win, open)
	for _, o := range owed {
		e.SuccessorsOwed = append(e.SuccessorsOwed, o.GapID)
	}
	for _, ev := range s.FoundClosed {
		c, _ := recordpb.BodyAs[*recordpb.Close](ev)
		f := FoundClosedJSON{ID: c.GetGapId(), ClosedBySeat: ev.GetSeatId(), Fate: recordpb.Word(c.GetClosureClass()), Successors: []string{}}
		for _, g := range succ[c.GetGapId()] {
			if open(g) {
				f.Successors = append(f.Successors, g)
			}
		}
		e.FoundClosed = append(e.FoundClosed, f)
	}
	return e
}

// thisSitting is the seat's own live events in its latest sitting: the acts the write path stored
// there. That sitting opened at the hook's bracket or at a
// register, and a register under the bracketed agent joined it — so a log filed between the
// bracket and that register is in the sitting, as the record says. A seat that has never sat has
// no earlier sitting to borrow from: its acts are in none, and that is its window.
func thisSitting(evs []*Event, win WindowIndex, seatID string) []*Event {
	latest := win.LatestSittingOf(seatID)
	var out []*Event
	for _, e := range evs {
		if e.GetSeatId() == seatID && win.Of(e).SittingID == latest {
			out = append(out, e)
		}
	}
	return out
}

// benchRegisteredSince reports whether the bench's current sitting opened after the dispatch row d and
// registered an occasion. d.at is a place in the stream's positions, as owedSitting reads it.
func benchRegisteredSince(evs []*Event, win WindowIndex, d dispatchRow) bool {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	if placesOf(evs, seq, win)[win.LatestSittingOf(benchSeat)] <= d.at {
		return false
	}
	for _, e := range thisSitting(evs, win, benchSeat) {
		if r, ok := recordpb.BodyAs[*recordpb.Register](e); ok && r.Occasion != nil {
			return true
		}
	}
	return false
}

// answeredByActThisSitting is the gaps the seat filed a motion on in its current sitting (a docket
// or grade motion names its gap), and whether it recorded a retire there.
func answeredByActThisSitting(evs []*Event, win WindowIndex, seatID string) (moved map[string]bool, retired bool) {
	moved = map[string]bool{}
	for _, e := range thisSitting(evs, win, seatID) {
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok {
			for _, g := range []string{m.GetDocket().GetGapId(), m.GetGrade().GetGapId()} {
				if g != "" {
					moved[g] = true
				}
			}
		}
		if _, ok := recordpb.BodyAs[*recordpb.Retire](e); ok {
			retired = true
		}
	}
	return moved, retired
}

// unreachableThisSitting is the anchors the seat verified as unreachable in its current sitting.
func unreachableThisSitting(evs []*Event, win WindowIndex, seatID string) map[string]bool {
	out := map[string]bool{}
	for _, e := range thisSitting(evs, win, seatID) {
		if v, ok := recordpb.BodyAs[*recordpb.Verify](e); ok && v.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE {
			out[v.GetAnchor()] = true
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

// sittingRecordedNothing reports whether the seat's current sitting holds no ACTS — nothing of a type
// recordsql.NotActs does not list, the definition seat_metrics counts acts by.
//
// It reads the one window every per-sitting reader does (thisSitting), so "this sitting" means one thing
// on this surface: the seat's acts the write path stored in its latest sitting — opened by its
// register, or by the hook's sitting_open where the configuration names one seat.
func sittingRecordedNothing(evs []*Event, win WindowIndex, seatID string) bool {
	if win.LatestSittingOf(seatID) == 0 {
		// No sitting has opened for this seat at all, so there is no empty sitting to excuse —
		// and saying "nothing recorded" here would discharge a duty the seat has not reached.
		return false
	}
	for _, e := range thisSitting(evs, win, seatID) {
		if !slices.Contains(recordsql.NotActs, e.GetType()) {
			return false
		}
	}
	return true
}

// openOf answers whether a gap is open, off the work list's gap rows: an id the rows do not hold is
// not open.
func openOf(gaps []WorkGapState) func(string) bool {
	open := map[string]bool{}
	for _, g := range gaps {
		open[g.ID] = g.Open
	}
	return func(id string) bool { return open[id] }
}
