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
// So the worklist becomes the seat's own pending work. It was described as "the merge's
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

// Duty is one outstanding obligation, with the act that discharges it.
type Duty struct {
	What string `json:"what"`
	// How names the WAYS OUT, not the way out. A duty a seat cannot act on is a complaint —
	// a blocked seat once searched ten-plus calls for a verb that did not exist — but naming a
	// single command steers: a seat handed one answer takes it and never reads the help or
	// weighs the alternative, and the alternative is often the honest move. Where a duty has
	// more than one legitimate discharge, every one of them is listed.
	How string `json:"how"`
}

// SittingJSON says whether this seat's sitting is finished, and what is left if not.
type SittingJSON struct {
	Seat string `json:"seat"`
	Role string `json:"role"`
	// Complete is the answer to "may I end my turn". It is false whenever anything is
	// outstanding, and a seat that ends anyway is ending against a stated list rather than in
	// the dark — which is the difference this exists to make.
	Complete    bool   `json:"complete"`
	Outstanding []Duty `json:"outstanding"`
}

// SittingOf computes what the given seat still owes on this board.
func SittingOf(b *Board, role, seatID string) SittingJSON {
	s := SittingJSON{Seat: seatID, Role: role, Outstanding: []Duty{}}
	if b == nil {
		return s
	}
	add := func(what, how string) { s.Outstanding = append(s.Outstanding, Duty{What: what, How: how}) }

	// EVERY SEAT CLOSES THE FRICTION CHANNEL. Silence is not the empty case: an absent friction
	// log reads the same whether the sitting was clean or the channel went unused, and across
	// eighteen recorded sittings it was the second every time.
	if !seatDid(b, seatID, "friction") && !seatDid(b, seatID, "friction-none") {
		add("the friction channel is open — you have neither reported a capability gap nor said that nothing blocked you",
			`friction --reason "<what you reached for and could not get>"  |  friction --none --reason "<what you reached for and found>"`)
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
			add("gap "+id+" was minted --check-kind computation and no proof answers it; prose cannot close it",
				`prove --location "<the exact sentence>" --script <path> --answers `+id+
					`   |   if the demand is wrong, say so in the --reason of the edit that answers `+id)
		}
		if !seatDid(b, seatID, "revision") {
			add("the round record is missing — a revision that is not on the record did not happen as far as the debate is concerned (W1.7)",
				`revision --reason "<what you changed this round>"`)
		}
	case "merge":
		// Both of these already REFUSE `verdict --as PASS`. Naming them here is the same list,
		// arriving when the seat can still act on it rather than at the terminal act.
		for _, id := range b.GapOrder {
			if g := b.Gaps[id]; g != nil && g.Open {
				add("gap "+id+" is open — PASS is refused while it is",
					`close --id `+id+` --as <disposition> --reason "..."   (or verdict --as FAIL)`)
			}
		}
		for _, m := range AllMotions(b) {
			if m != nil && !m.Ruled() {
				add("motion "+m.ID+" was filed and never ruled — PASS is refused while it stands",
					`show --view motions   then   motion `+m.Subject+` rule --id `+m.ID+` --as <verdict> --reason "..."`)
			}
		}
		if !seatDid(b, seatID, "verdict") {
			add("your terminal act is missing — the run cannot say from its own record that it was ever verified",
				`verdict --as PASS|FAIL`)
		}
	case "bench":
		for _, m := range AllMotions(b) {
			if m != nil && !m.Ruled() && m.Subject == "petition" {
				add("petition "+m.ID+" is unruled, and petitions are heard BEFORE the debate continues",
					`show --view motions   then   motion petition rule --id `+m.ID+` --as granted|denied --reason "..."`)
			}
		}
	}
	s.Complete = len(s.Outstanding) == 0
	return s
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
