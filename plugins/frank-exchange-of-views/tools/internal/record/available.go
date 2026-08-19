package record

import (
	"fmt"
)

// WHAT THIS BOARD MAKES AVAILABLE, WHICH IS NOT WHAT THIS SEAT OWES.
//
// This file used to open by arguing that the affordances had to be a SECOND list, kept off
// `complete` so the view could not disagree with the gates. The argument was sound and the
// conclusion was wrong, and the difference is worth keeping because it is a shape that recurs.
//
// The rule it was protecting — sitting.go's "nothing here invents an obligation", because a seat
// told it is finished by one surface and refused by another learns to trust neither — is real and
// still holds. But a duty is derived only where omission already carries a mechanical consequence,
// so it can name 4 of blue's verbs, 5 of merge's, 3 of bench's and 2 of lens's. Every act whose
// omission is merely a QUALITY loss — a line of inquiry never revisited, a repair with no manifest
// row, an archive never sampled, a citation nobody verified, a source blue never cited, a proof
// nobody re-ran, a grade accepted and never moved — got no line. Those are exactly the acts the
// probe boards bait and the measured runs skip.
//
// Protecting the rule by making a SEPARATE list made those acts absent from the surface a seat
// consults to ask whether anything is left. Three seats interviewed about verbs they never touched
// all stopped the same way and said so in the same words — the outstanding array emptied — and one
// put the mechanism exactly: the work list "supplied the goal, and a goal crowds out survey.
// Things off the list weren't declined. They were invisible."
//
// The rule never required a second LIST. It required that `complete` be computed from the
// enforced half only. That is a property of one field, not of two surfaces, so it is now a flag
// on each item (record.Item.Blocks) and there is one list. See sitting.go.
//
// The carriage argument that used to close this comment is also settled and is left here because
// it is the measurement, not the opinion: across 24 probe dispatches the duty-carrying projection
// was read 0.33-2.00 times a sitting while `board` was read 2.7-4.3. The answer to a list that
// does not arrive is not a second copy of it somewhere better-trafficked. It is one list, on the
// one command a seat is told to run.

// AvailableOf derives the acts this board affords this seat right now.
//
// EVERY ENTRY NAMES A CONDITION READ OFF THE RECORD, never a general encouragement. "You could
// cite something" is not an affordance, it is advice; "claim c-3 is cited and nobody has verified
// it" is a fact about this board with one act attached to it.
//
// NOT DERIVED, AND SAID RATHER THAN LEFT AS A SILENT HOLE: `closing` wants the docket, which is
// not recoverable from board state alone; `motion grade file` and `retire` turn on a judgement
// (whether you disagree, whether the claim should go) that no derivation can make for the seat.
// An affordance line for those would be an expectation that cannot be honestly met, which is the
// defect TestEveryExpectationIsReachableOnItsBoard exists to catch one layer up.
func AvailableOf(b *Board, role, seatID string) []Item {
	out := []Item{}
	if b == nil {
		return out
	}
	// Blocks is FALSE on every one of these: an affordance is open work, not owed work,
	// and that distinction is the whole content of the rule this file used to split a list
	// in half to honour.
	add := func(what string) { out = append(out, Item{What: what}) }

	switch role {
	case "blue":
		// A line proposed and never revisited records an intention rather than a choice.
		// Measured over six runs: 83 of 86 inquiries were declared in round 0 and not one was
		// ever moved.
		//
		// THE PREDICATE IS StaleInquiries, NOT A SECOND COPY OF IT. This block used to inline
		// `Status == "proposed" || Status == "pursued"`, which is what StaleInquiries also said,
		// so the two drifted together and were wrong together: neither read the round, though
		// both texts promised one. The `How` also named a status set that excluded `pursued`,
		// which is where a followed line comes to REST — so a seat that did the right thing
		// was told to abandon or defer it. Both are fixed at the single predicate now.
		for _, a := range StaleInquiries(b) {
			add(fmt.Sprintf("line of inquiry %s is at %q and has not moved since round %d — a line declared once and never revisited records an intention rather than a choice", a.ID, a.Status, a.Round))
		}
		// A repair with no receipt is one nobody audited, including its author.
		for _, id := range gapsEditedWithoutManifest(b, seatID) {
			add("gap " + id + " was answered by an edit and carries no manifest row — the report names a closed gap with no row as a repair nobody audited, including its author")
		}
	case "merge":
		if anyClosedGap(b) && !seatDid(b, seatID, "spot-check") {
			add("the closure archive is not empty and this sitting has sampled none of it")
		}
		// Accepting a grade motion does not move the grade. Saying so is not doing it.
		for _, id := range gapsWithAcceptedMotionAndNoRegrade(b) {
			add("gap " + id + " had a grade motion ACCEPTED and no regrade followed it — accepting a dispute does not move the grade, and a grade that moved with no regrade event reads as though the dispute was answered by silence")
		}
	case "lens":
		for _, key := range citedClaimsWithoutVerify(b) {
			add("citation " + key + " is on the record and nobody has verified it against what the source actually says")
		}
		for _, id := range proofsWithoutReproduce(b) {
			add("proof " + id + " is recorded and nobody has re-run it — a proof is audited by RE-RUNNING it, not by reading it")
		}
	}
	return out
}

func anyClosedGap(b *Board) bool {
	for _, id := range b.GapOrder {
		if g := b.Gaps[id]; g != nil && !g.Open {
			return true
		}
	}
	return false
}

// gapsEditedWithoutManifest finds gaps this seat answered with an edit and never receipted.
func gapsEditedWithoutManifest(b *Board, seatID string) []string {
	edited, rowed := map[string]bool{}, map[string]bool{}
	var order []string
	for _, e := range b.Events {
		if e.SeatID != seatID {
			continue
		}
		switch e.Type {
		case "blue_edit":
			if id := e.Payload.Str("answers"); id != "" && !edited[id] {
				edited[id] = true
				order = append(order, id)
			}
		case "manifest-row":
			if id := e.Payload.Str("gap_id"); id != "" {
				rowed[id] = true
			}
			if id := e.Payload.Str("id"); id != "" {
				rowed[id] = true
			}
		}
	}
	var out []string
	for _, id := range order {
		if !rowed[id] {
			out = append(out, id)
		}
	}
	return out
}

// gapsWithAcceptedMotionAndNoRegrade finds grades argued down and never actually moved.
func gapsWithAcceptedMotionAndNoRegrade(b *Board) []string {
	regraded := map[string]bool{}
	for _, e := range b.Events {
		if e.Type == "regrade" {
			if id := e.Payload.Str("gap_id"); id != "" {
				regraded[id] = true
			}
			if id := e.Payload.Str("id"); id != "" {
				regraded[id] = true
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for _, e := range b.Events {
		if e.Type != "motion-rule" || e.Payload.Str("subject") != "grade" {
			continue
		}
		if e.Payload.Str("as") != "accepted" && e.Payload.Str("verdict") != "accepted" {
			continue
		}
		id := e.Payload.Str("gap_id")
		if id == "" || regraded[id] || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// citedClaimsWithoutVerify finds citations on the record that nobody checked.
//
// THE JOIN IS THE CITATION ANCHOR, and it was neither of the keys this used to read.
//
// It looked for `cite_key`, then `key`, on both sides. A `cite` event carries neither — its keys
// are access_date, claim, label, location, sha256, title, url — so the lookup always missed, the
// slice was always empty, and no lens was ever told about an unverified citation. "Nothing is
// outstanding" and "the join key is not on the event" were the same bytes, and the affordance had
// never fired in the tool's lifetime.
//
// It also hid a second defect: the instruction that dead path handed over named `--key`, a flag
// `lens verify` does not have, so the one time it fired it would have failed. Neither was visible
// because the other kept it off the board.
//
// `blue cite` records the anchor id as `label` (c-<hex>, the token it splices into the report);
// `lens verify` records the citation it checked as `anchor`. An INDEPENDENT verify carries no
// anchor and deliberately discharges nothing: it is a check against a source blue never cited.
func citedClaimsWithoutVerify(b *Board) []string {
	verified := map[string]bool{}
	for _, e := range b.Events {
		if e.Type == "verify" {
			if k := e.Payload.Str("anchor"); k != "" {
				verified[k] = true
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for _, e := range b.Events {
		if e.Type != "cite" {
			continue
		}
		k := e.Payload.Str("label")
		if k == "" || verified[k] || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, k)
	}
	return out
}

// proofsWithoutReproduce finds computations nobody re-ran. `print("7 is prime")` reproduces
// perfectly forever, which is why the re-run is the audit and reading the script is not.
func proofsWithoutReproduce(b *Board) []string {
	rerun := map[string]bool{}
	for _, e := range b.Events {
		if e.Type == "reproduce" {
			if id := e.Payload.Str("proof_id"); id != "" {
				rerun[id] = true
			}
			if id := e.Payload.Str("id"); id != "" {
				rerun[id] = true
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for _, e := range b.Events {
		if e.Type != "proof" {
			continue
		}
		id := e.Payload.Str("proof_id")
		if id == "" {
			id = e.Payload.Str("id")
		}
		if id == "" || rerun[id] || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
