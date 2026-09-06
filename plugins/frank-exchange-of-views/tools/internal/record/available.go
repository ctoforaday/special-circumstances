package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

// availableOf derives the acts this board affords this seat right now.
//
// EVERY ENTRY NAMES A CONDITION READ OFF THE RECORD, never a general encouragement. "You could
// cite something" is not an affordance, it is advice; "claim c-3 is cited and nobody has verified
// it" is a fact about this board with one act attached to it.
//
// NOT DERIVED, AND SAID RATHER THAN LEFT AS A SILENT HOLE: `motion grade file` and `retire` turn on a judgement
// (whether you disagree, whether the claim should go) that no derivation can make for the seat.
// An affordance line for those would be an expectation that cannot be honestly met, which is the
// defect TestEveryExpectationIsReachableOnItsBoard exists to catch one layer up.
func availableOf(evs []*Event, gaps []WorkGapState, role, seatID string) []Item {
	out := []Item{}
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
		for _, a := range StaleInquiriesOf(evs) {
			add(fmt.Sprintf("line of inquiry %s is at %q and has not moved since round %d — a line declared once and never revisited records an intention rather than a choice", a.ID, a.Status, a.Round))
		}
		// A repair with no receipt is one nobody audited, including its author.
		for _, id := range gapsEditedWithoutManifest(evs, seatID) {
			add("gap " + id + " was answered by an edit and carries no manifest row — the report names a gap YOU repaired that carries no row as a repair nobody audited, including its author")
		}
	case "merge":
		// AN OPEN GAP IS EITHER CLOSED OR DOCKETED, and until the docket was a motion this could
		// not be said here at all — the note above used to record that "`closing` wants the
		// docket, which is not recoverable from board state alone". It is now: a docket motion is
		// an event, so "which gaps are before the bench" is a question the board answers.
		//
		// AN AFFORDANCE, NOT A DUTY. The open-gap row in sitting.go already refuses this seat's
		// PASS, so a second blocking item would be a duplicate — and worse, a seat that OBEYED it
		// would be further from complete than before, because filing adds an unruled motion. What
		// filing buys is not completeness but the gap moving from this seat's undecided pile to a
		// question the bench owes an answer to.
		//
		// NO FLAGS IN THE TEXT. This file's own contract, and the reason is recorded above: an
		// affordance once handed a seat a flag the verb does not have, so the only instruction the
		// duty ever gave could not run.
		// PENDING, NOT EVER-FILED — and the difference is the whole `carried` design.
		//
		// A docket motion that has been RULED is answered. If the gap is still open after that
		// answer, the answer was `carried`: every other disposition closes the gap, so an open gap
		// with a ruled docket motion is EXACTLY the deferral, and the ruling states what would
		// bring it back. Keyed on ever-filed, this affordance went silent permanently at the first
		// filing, so the re-file the deferral asks for was never once offered — the capability the
		// bench deliberately kept alive, and no surface said so. Measured: after M2 was ruled
		// `carried` on R1-2, `show work` listed the gap as blocking PASS and offered nothing to do
		// about it, while a never-docketed R1-3 got the affordance.
		//
		// An UNRULED motion still suppresses it, because filing a second time while the first is
		// pending asks the bench the same question twice.
		pending := map[string]bool{}
		for _, m := range MotionsOf(evs) {
			if m != nil && m.Subject == "docket" && m.GapID != "" && !m.Ruled() {
				pending[m.GapID] = true
			}
		}
		for _, g := range gaps {
			if !g.Open || pending[g.ID] {
				continue
			}
			// ONE ROW PER GAP, BETTER WORDS WHEN THE BOARD KNOWS MORE (#759).
			//
			// A carried gap and a gap nobody has ever docketed both reach this line, and until
			// the view could tell them apart they got the same sentence — true of both and
			// actionable on neither. A SECOND row for the carried case would be the wrong fix:
			// two items naming one gap is a duplicate, and the open-gap row in sitting.go
			// already blocks PASS over it. So the row is REPLACED, not added.
			//
			// The reopens-on condition is the substance. "The bench carried this" tells a seat
			// the history; "the bench carried it until blue reports what the stated direction
			// found" tells it what has to happen, and that sentence is the bench's own words off
			// the record rather than this file's paraphrase of them.
			if g.AwaitingDocket {
				what := "gap " + g.ID + " is open because the BENCH CARRIED it — the motion was answered, so nothing is pending and the gap returns only if it is docketed again"
				if g.DocketReopensOn != "" {
					what += ". The ruling says what reopens it: " + g.DocketReopensOn
				}
				add(what)
				continue
			}
			add("gap " + g.ID + " is open and you have not closed it — `motion docket file` puts it before the bench, which is the channel for a gap you cannot settle yourself")
		}
		if anyClosedGap(gaps) && !seatDid(evs, seatID, recordpb.EventType_EVENT_TYPE_SPOT_CHECK) {
			add("the closure archive is not empty and this sitting has sampled none of it")
		}
		// Accepting a grade motion does not move the grade. Saying so is not doing it.
		for _, id := range gapsWithAcceptedMotionAndNoRegrade(evs) {
			add("gap " + id + " had a grade motion ACCEPTED and no regrade followed it — accepting a dispute does not move the grade, and a grade that moved with no regrade event reads as though the dispute was answered by silence")
		}
	case "lens":
		for _, key := range citedClaimsWithoutVerify(evs) {
			add("citation " + key + " is on the record and nobody has verified it against what the source actually says")
		}
		for _, id := range proofsWithoutReproduce(evs) {
			add("proof " + id + " is recorded and nobody has re-run it — a proof is audited by RE-RUNNING it, not by reading it")
		}
	}
	return out
}

func anyClosedGap(gaps []WorkGapState) bool {
	for _, g := range gaps {
		if !g.Open {
			return true
		}
	}
	return false
}

// gapsEditedWithoutManifest finds gaps this seat answered with an edit and never receipted.
//
// THE `id` FALLBACK ON THE RECEIPT SIDE IS GONE, and it was never a fallback. It read
// `Payload.Str("id")` beside `gap_id` on a manifest-row event; `blue manifest-row` writes
// `gap_id` and nothing else joinable (blue/manifest_row.go:27), and ManifestRow carries exactly
// `gap_id` and `row`. So the second arm matched nothing on any event this tool has ever written.
// Its removal changes no outcome, which is the point: the live arm is the one that was always
// doing the work.
func gapsEditedWithoutManifest(evs []*Event, seatID string) []string {
	edited, rowed := map[string]bool{}, map[string]bool{}
	var order []string
	for i := range evs {
		e := evs[i]
		if e.GetSeatId() != seatID {
			continue
		}
		body, ok := recordpb.Body(e)
		if !ok {
			// No body at all, so neither an edit's `answers` nor a receipt's `gap_id` can be on
			// it. The old code reached the same answer by a different route: Payload.Str on an
			// absent key returned "", which both `!= ""` filters dropped.
			continue
		}
		switch x := body.(type) {
		case *recordpb.BlueEdit:
			if id := x.GetAnswers(); id != "" && !edited[id] {
				edited[id] = true
				order = append(order, id)
			}
		case *recordpb.ManifestRow:
			if id := x.GetGapId(); id != "" {
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
//
// THE RULING AND THE GAP ARE ON TWO DIFFERENT EVENTS, AND THIS FUNCTION USED TO READ BOTH OFF ONE.
// It is the same defect citedClaimsWithoutVerify documents below — a join key that is not on the
// event — and it had never fired once, on any vocabulary, in this function's entire lifetime.
//
// What it read, against what a `motion-rule` event has ever carried:
//
//   - `as` and `verdict` for the ruling. The sole writer (cli/motion/verbs.go) has written the key
//     `ruling` since the commit that introduced THIS FILE (487efa2, 2026-08-15) — `as` is the
//     FLAG word, `--as`, which is a different vocabulary from the payload it lands in. Neither
//     `as` nor `verdict` was ever a pre-collapse spelling of a ruling either: the retired
//     grade exchange was `dispute`/`dispute-respond` keyed `response`. So both arms of that
//     condition matched nothing, always.
//   - `gap_id` for the identity. A ruling names the MOTION (`motion_id`); the gap is on the
//     FILING, where `motion grade file --id` lands it (cli/motion/verbs.go payloadKey). No
//     motion-rule event has ever carried a gap_id, so even a corrected ruling check would have
//     produced an empty id and skipped.
//
// The affordance is wanted — MotionVerdicts["grade"] tells the ruler "accepting without regrading
// is a channel with no consequence", and this is the only thing that checks it — so the answer is
// the join the record actually supports, not deletion. MotionRule carries no `gap_id` field and no
// `as` field, so the schema now refuses the old shape outright.
//
// THE JOIN IS Motions(), NOT A SECOND COPY OF IT. Filing and ruling live in different shards and
// replay interleaves, so a ruling can arrive before the motion it answers; motion.go carries the
// two-pass projection that survives that, and its header records shipping the single-pass bug
// once already. A private join here would re-earn it.
func gapsWithAcceptedMotionAndNoRegrade(evs []*Event) []string {
	regraded := map[string]bool{}
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		// The `id` fallback that sat beside this is gone for the same reason as the manifest
		// one: `merge regrade --id` writes `gap_id` (cli/merge/regrade.go:23) and Regrade
		// carries no other identifier, so the second arm never matched.
		if r, isRegrade := body.(*recordpb.Regrade); isRegrade {
			if id := r.GetGapId(); id != "" {
				regraded[id] = true
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for _, m := range MotionsOf(evs) {
		if m.Subject != "grade" || m.Ruling != "accepted" {
			continue
		}
		id := m.Fields["gap_id"]
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
func citedClaimsWithoutVerify(evs []*Event) []string {
	verified := map[string]bool{}
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		if v, isVerify := body.(*recordpb.Verify); isVerify {
			// An INDEPENDENT verify carries no anchor and discharges nothing, and that stays a
			// property of the ANCHOR being absent rather than of the `independent` flag being
			// set — the two are separate fields and the empty-anchor test is the one this
			// function has always made.
			if k := v.GetAnchor(); k != "" {
				verified[k] = true
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		c, isCite := body.(*recordpb.Cite)
		if !isCite {
			continue
		}
		k := c.GetLabel()
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
//
// THE DISCHARGE SIDE JOINED ON A KEY A `reproduce` EVENT HAS NEVER CARRIED, and this one failed
// in the direction that hides best: it produced a NAG THAT COULD NOT BE CLEARED, not a silence.
//
// The re-run set was built from `proof_id`, then `id`, on the reproduce event. `lens reproduce`
// writes proof_sha, reproduced, soundness and the two outputs (cli/lens/reproduce.go:69) and
// Reproduce carries no identifier but `proof_sha` — so the set was ALWAYS EMPTY, and every proof
// on the board was reported as never re-run for the rest of the run no matter how many times red
// re-ran it. The proof side worked, which is what kept it plausible.
//
// A PROOF AND ITS RE-RUN JOIN ON proof_sha. That is the schema's own pairing (Proof.proof_sha,
// Reproduce.proof_sha) and the join evidenceview.go and citationid.go already make; it is also
// the only one available, since the sha is the whole of what the re-runner is given.
//
// The reported token stays the PROOF ID, because that is what the affordance's sentence names and
// what the seat carries away. See the return notes: `lens reproduce` takes `--id <sha256>`, so a
// seat handed a `p-…` label still has to resolve it through `lens show evidence` — the same
// shape of dead-end the `--key` note below records, and left as found rather than redesigned here.
func proofsWithoutReproduce(evs []*Event) []string {
	rerun := map[string]bool{}
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		if r, isRerun := body.(*recordpb.Reproduce); isRerun {
			if sha := r.GetProofSha(); sha != "" {
				rerun[sha] = true
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		p, isProof := body.(*recordpb.Proof)
		if !isProof {
			continue
		}
		id := p.GetProofId()
		if id == "" || seen[id] {
			continue
		}
		// A proof with no sha cannot be re-run at all, so it cannot be discharged: it is
		// reported, which is what the old code did with every proof, rather than dropped into a
		// silence that would read as audited.
		if sha := p.GetProofSha(); sha != "" && rerun[sha] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
