package record

// WHAT THIS SEAT STILL OWES, ON THE READ IT ALREADY DOES.
//
// A seat had no way to know it was finished. Asked directly, the merge named a real mechanism —
// "the `verdict` command either succeeds or fails; if it fails the tool tells me what's blocking
// closure, and I iterate" — because `verdict` refuses over open gaps and unruled motions and
// enumerates them. Blue and the bench answered with things they cannot observe: "red agrees it's
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
// Nothing here invents an obligation. Each one is either refused at a write path (open gaps and
// unruled motions block `verdict`; a computation gap cannot be closed on prose) or is a stated
// round-record requirement (W1.7's revision, the bench's terminal outcome). Inventing a duty here
// would make this view disagree with the gates, and a seat told it was finished by one surface
// and refused by another learns to trust neither.

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
}

// SittingOf computes what the given seat still owes on this board.
func SittingOf(b *Board, role, seatID string) SittingJSON {
	s := SittingJSON{Seat: seatID, Role: role, Open: []Item{}}
	if b == nil {
		return s
	}
	add := func(what string) { s.Open = append(s.Open, Item{What: what, Blocks: true}) }

	// EVERY SEAT CLOSES THE FRICTION CHANNEL. Silence is not the empty case: an absent friction
	// log reads the same whether the sitting was clean or the channel went unused, and across
	// eighteen recorded sittings it was the second every time.
	if !seatDid(b, seatID, "friction") && !seatDid(b, seatID, "friction-none") {
		add("the friction channel is open — you have neither reported a capability gap nor said that nothing blocked you")
	}

	switch role {
	case "blue":
		// A computation demand prose cannot answer. The merge is REFUSED if it tries to close
		// one unproved, so an unanswered demand does not settle — it carries into the next round.
		for _, id := range gapsAwaitingProofOn(b) {
			// TWO HONEST ANSWERS, and the second is not a lesser one: there is no verb for
			// contesting a check_kind, so a seat that thinks the demand is wrong argues it in
			// the edit that answers the gap. Naming only `prove` would make disagreement look
			// like non-compliance.
			add("gap " + id + " was minted --check-kind computation and no proof answers it; prose cannot close it")
		}
		// RED SAID THE DOCUMENT DOES NOT CARRY THIS LINE, and that is blue's to answer.
		//
		// `weakened` is deliberately absent: it is red flagging erosion, an argument blue may
		// accept, and a duty firing on it would make every "this is thinner than it was" a
		// blocking repair. SupportDemandsBlue is the one statement of which verdicts bind.
		for _, a := range UnsupportedInquiries(b) {
			add("red voted line of inquiry " + a.ID + " `" + a.Support + "` — the report no longer backs it as stated, or does not carry it at all")
		}
		if !seatDid(b, seatID, "revision") {
			add("the round record is missing — a revision that is not on the record did not happen as far as the debate is concerned (W1.7)")
		}
	case "merge":
		// Both of these already REFUSE `verdict --as PASS`. Naming them here is the same list,
		// arriving when the seat can still act on it rather than at the terminal act.
		for _, id := range b.GapOrder {
			if g := b.Gaps[id]; g != nil && g.Open {
				add("gap " + id + " is open — PASS is refused while it is")
			}
		}
		for _, m := range Motions(b) {
			if m != nil && !m.Ruled() {
				add("motion " + m.ID + " was filed and never ruled — PASS is refused while it stands")
			}
		}
		// EVERY LINE OF INQUIRY IS VOTED, EVERY ROUND.
		//
		// The report is regenerated each round, so a verdict cast before this round's edits
		// answers a question about a document that no longer exists. This is the channel that
		// makes "we pursued X" checkable at all: the row `assemble` generates carries no anchor,
		// so without a vote it is the one claim in the report nothing can refuse.
		for _, a := range UnvotedInquiries(b) {
			add("line of inquiry " + a.ID + " has no support verdict this round — PASS is refused while the report's own account of its research is unchecked")
		}
		if !seatDid(b, seatID, "verdict") {
			add("your terminal act is missing — the run cannot say from its own record that it was ever verified")
		}
	// THE LENS HAS NO CASE, AND THAT IS THE RULE HOLDING RATHER THAN A GAP IN IT.
	//
	// Checked 2026-08-15 rather than assumed: nothing refuses a sitting over a missing lens act,
	// and the scorecard scores no lens parity duty. Under the rule stated at the top of this file
	// — every duty is enforced at a write path or scored at capture — a lens duty would be an
	// invented obligation, and `complete: false` on a seat no gate would hold is exactly the
	// disagreement that teaches a seat to trust neither surface.
	//
	// Written down because the absence reads as an oversight to anyone who has just fixed the
	// roleOf defect and is looking for more of it. The acts a lens genuinely has open to it —
	// verifying a citation nobody checked, corroborating one blue never cited, re-running a proof
	// nobody re-ran — come from AvailableOf and land on this same list carrying Blocks:false. A
	// lens therefore has an EMPTY blocking set and a non-empty work list, which is the accurate
	// statement and was previously unsayable: the old shape could only express it as an empty
	// work list, and a lens seat read that and stopped.
	case "bench":
		for _, m := range Motions(b) {
			if m != nil && !m.Ruled() && m.Subject == "petition" {
				add("petition " + m.ID + " is unruled, and petitions are heard BEFORE the debate continues")
			}
		}
	}
	// THE AFFORDANCES GO ON THE SAME LIST, and they go on it LAST so the blocking items read
	// first. They carry Blocks:false, so they are visible without being owed.
	s.Open = append(s.Open, AvailableOf(b, role, seatID)...)
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
func seatDid(b *Board, seatID, typ string) bool {
	for _, e := range b.Events {
		if e.SeatID == seatID && e.Type == typ {
			return true
		}
	}
	return false
}

// gapsAwaitingProofOn is GapsAwaitingProof against a board the caller already holds.
func gapsAwaitingProofOn(b *Board) []string {
	var out []string
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || !g.Open || g.Mint == nil {
			continue
		}
		if g.Mint.Str("check_kind") == CheckKindComputation && !proofNames(b, id) {
			out = append(out, id)
		}
	}
	return out
}
