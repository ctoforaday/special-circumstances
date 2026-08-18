package record

import (
	"fmt"
	"os"
	"strings"
)

// WHAT THIS BOARD MAKES AVAILABLE, WHICH IS NOT WHAT THIS SEAT OWES.
//
// # Why this is a second list rather than more duties
//
// sitting.go states the rule that stops the duty list growing: "Nothing here invents an
// obligation… Inventing a duty here would make this view disagree with the gates, and a seat told
// it was finished by one surface and refused by another learns to trust neither." That rule is
// correct and this file does not touch it. `Outstanding` still gates `sitting.complete`, and every
// entry in it is still refused at a write path or scored at capture.
//
// The rule has a consequence nobody chose, though. A duty is derived ONLY where omission already
// carries a mechanical consequence, so the channel can name 4 of blue's 17 verbs, 5 of merge's 17,
// 3 of bench's 14 and 2 of lens's 10. Every verb whose omission is merely a QUALITY loss — a line
// of inquiry never revisited, a repair with no manifest row, an archive never sampled, a citation
// nobody verified, a proof nobody re-ran, a grade accepted and never moved — gets no line at all.
// Those are precisely the acts the probe boards are built to bait, and precisely the ones the
// measured runs skip.
//
// So the fact a seat needs — "this act is open to you, here, now" — is separated from the claim it
// must not be confused with — "you are not finished until you do it". Available NEVER affects
// Complete. A seat reading it is being told what the board affords, not what it owes, and the two
// are different sentences in the JSON as well as in the doc.
//
// # Measured, and it is why the carriage matters as much as the content
//
// Across 24 probe dispatches the duty-carrying projection `worklist` was read 0.33–2.00 times a
// sitting while `board` — described in this tool's own words as "the form a seat acts on" — was
// read 2.7–4.3 times. A list that only rides on the projection a seat rarely opens is a list that
// mostly does not arrive, and its silence is indistinguishable from having nothing to say.

// DutyArm selects how much of this machinery is live. It exists for the probe, and the DEFAULT IS
// THE SHIPPED BEHAVIOUR: unset means exactly what the tool did before this file existed.
//
// An experiment knob inside a production path is a real cost and it is taken deliberately. The
// alternative — a second binary built from a patched tree — measures the patch rather than the
// tool, which is the mistake one level up from the one this is being used to correct. It is
// asserted by TestDutyArmDefaultIsTheShippedSet, so a knob that silently changed the default would
// fail rather than quietly rewrite what every seat reads.
type DutyArm string

const (
	// DutyOff emits neither list — the floor, for asking whether the channel does anything.
	DutyOff DutyArm = "off"
	// DutyShipped is Outstanding only: what the tool has always sent.
	DutyShipped DutyArm = "shipped"
	// DutyAvailable adds the affordance list to the worklist.
	DutyAvailable DutyArm = "available"
	// DutyAvailableOnBoard also carries the sitting on `board`, the projection seats read.
	DutyAvailableOnBoard DutyArm = "available+board"
)

// DutyArmEnv is the variable the probe sets on the seat's process.
const DutyArmEnv = "FEOV_DUTY_ARM"

// CurrentDutyArm reads the arm, defaulting to the shipped behaviour. An unrecognised value is
// TREATED AS SHIPPED rather than as off: a typo that silently emptied the seat's worklist would
// look exactly like a board with nothing outstanding.
func CurrentDutyArm() DutyArm {
	switch DutyArm(strings.TrimSpace(os.Getenv(DutyArmEnv))) {
	case DutyOff:
		return DutyOff
	case DutyAvailable:
		return DutyAvailable
	case DutyAvailableOnBoard:
		return DutyAvailableOnBoard
	default:
		return DutyShipped
	}
}

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
func AvailableOf(b *Board, role, seatID string) []Duty {
	out := []Duty{}
	if b == nil {
		return out
	}
	add := func(what string) { out = append(out, Duty{What: what}) }

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
