package fuzz

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// unseatInquiryCheck is a lens other than the inquiry's owner minting, once a drawn run, a gap below
// material about an inquiry a check stands on, and holding the record to the ruling (gblock
// 2026-11-10): any gap about an inquiry, whatever its materiality, makes the owner's check owed
// again, and the owner's item for it names the gap. universe-m40 shipped a `confirmed` fate over a
// false clause beside such a gap (G-4aa30769): the class graded it below material, and the check
// that judged the fate before it went on standing.
//
// Forced runs take none: the check it re-owes is a sitting of the owner's their bound does not count.
// A mint the tool refuses — a gap of the drawn class already open on the inquiry — is tried again
// at the lens's next sitting.
func (r *runner) unseatInquiryCheck(seatID string) {
	r.mu.Lock()
	due := r.inquiryUnseats == 0 && !r.forcedVerdict() && !r.forceEstoppel
	r.mu.Unlock()
	if !due || seatID == record.InquiryOwner() || !strings.HasPrefix(seatID, "red-lens-") {
		return
	}
	js, ok := r.inquiryJudgements()
	if !ok {
		return
	}
	var sound *record.InquiryJudgement
	for i := range js {
		if js[i].State() == record.InquirySound {
			sound = &js[i]
			break
		}
	}
	if sound == nil {
		return
	}
	checked := sound.Check.ID
	out, err := r.exec("--json", "mint", "--seat-id", seatID,
		"--problem", "fuzz: the fate claims a clause the research did not establish",
		"--check-kind", "document", "--check", "the fate claims only what held", "--fix", dirCounter,
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--class", pick(r.rng, mintPoolClasses),
		"--about-kind", "inquiry", "--about", sound.ID, "--reason", "fuzz: a gap below material about a checked inquiry")
	if err != nil {
		return
	}
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		r.noteActErr("an admitted mint's --json result names no gap_id: " + firstLine(out))
		return
	}
	gap := env.Result.GapID
	r.noteMint(seatID, gap)
	if msg := r.checkOwedAgain(sound.ID, checked, gap); msg != "" {
		r.noteActErr(msg)
		return
	}
	r.mu.Lock()
	r.inquiryUnseats++
	r.mu.Unlock()
}

// checkOwedAgain reads, after gap was minted below material about inquiry id whose check at event
// checked stood, whether the owner owes the check again with the gap named on its item; "" where it
// does. THE GATE'S ROW IS READ, not the judgement's state alone: the item is what the owner's sitting
// is handed. A sibling's act may have moved the inquiry since — the owner checking it again, or a
// material gap making it wanting — and that is no breach; the check that stood before the gap still
// standing is.
func (r *runner) checkOwedAgain(id string, checked int64, gap string) string {
	bs, err := record.PassBlockers(r.run())
	if err != nil {
		return "the blockers do not read after a gap about inquiry " + id + ": " + err.Error()
	}
	for _, b := range bs {
		if b.Kind != record.BlockerUncheckedInquiry || b.Subject != id {
			continue
		}
		if item := b.WorkItem(); !strings.Contains(item, gap+" (") {
			return fmt.Sprintf("inquiry %s's check is owed again after %s, minted about it below material, and the owner's item does not name the gap: %s", id, gap, item)
		}
		return ""
	}
	j, ok := r.inquiryJudgement(id)
	switch {
	case !ok:
		return "inquiry " + id + " has no judgement after a gap was minted about it"
	case j.State() == record.InquiryWanting:
		return ""
	case j.Check != nil && j.Check.ID == checked:
		return fmt.Sprintf("inquiry %s's check at event %d still stands after %s was minted about it below material: any gap about an inquiry makes its check owed again", id, checked, gap)
	case j.Check != nil:
		return ""
	}
	return fmt.Sprintf("inquiry %s is %s after %s was minted about it below material, and the gate holds no unchecked_inquiry row for it", id, j.State(), gap)
}

// A GAP BELOW MATERIAL ABOUT A CHECKED INQUIRY RE-OWES ITS CHECK, run outside the release gate on the
// drive's own sittings: blue settles an inquiry, its owner checks it, and another lens's sitting mints
// a gap below material about it — the check no longer stands, the owner's item names the gap, and the
// owner's next sitting checks it again with the gap in hand.
func TestAGapBelowMaterialReOwesTheInquiryCheck(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1")); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	const blue, owner, lens = "blue-respond", "red-lens-logic", "red-lens-dark-side"
	if owner != record.InquiryOwner() {
		t.Fatalf("the inquiry owner is %s, and the test sits %s", record.InquiryOwner(), owner)
	}
	r.sit("blue", blue)
	out, ok := r.do("inquiry propose", blue).set("--question", "fuzz: does it settle").set("--hypothesis", "fuzz: it settles, and every table lists it").admit()
	id := inquiryIDOf(out)
	if !ok || id == "" {
		t.Fatalf("propose: %s (%s)", out, r.actErr)
	}
	if _, ok := r.do("inquiry record", blue).set("--id", id).set("--as", "confirmed").set("--reason", "fuzz: what the approach found").admit(); !ok {
		t.Fatalf("record: %s", r.actErr)
	}
	r.sit("lens", owner)
	r.checkInquiries(owner)
	if j, ok := r.inquiryJudgement(id); !ok || j.State() != record.InquirySound {
		t.Fatalf("the owner's check left inquiry %s %v (%s)", id, j.State(), r.actErr)
	}
	r.sit("lens", lens)
	r.unseatInquiryCheck(lens)
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if r.inquiryUnseats != 1 {
		t.Fatalf("inquiry unseats %d, want 1: the lens's gap below material about %s did not re-owe its check", r.inquiryUnseats, id)
	}
	if j, _ := r.inquiryJudgement(id); j.State() != record.InquiryUnchecked {
		t.Errorf("after a gap below material, inquiry %s is %s, want %s", id, j.State(), record.InquiryUnchecked)
	}
	r.sit("lens", owner)
	r.checkInquiries(owner)
	if j, _ := r.inquiryJudgement(id); j.State() != record.InquirySound || r.actErr != "" {
		t.Errorf("the owner's next sitting left inquiry %s %s (%s)", id, j.State(), r.actErr)
	}
}
