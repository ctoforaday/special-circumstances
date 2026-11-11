package fuzz

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FAILED-CHECK DRIVES KEY ON THE CHECK'S STATE (plans/contradiction-is-a-gap.md §III.11, partner
// §V.5.q): the ledger's two states are `unread` and `unsound`, and a drive that told a check's kind by
// its Detail would read a citation's empty one as either. Run outside the release gate: one unread
// check by lens U carried by G-u (U's unreachable reading through the built binary), then U's half,
// the refused close of the gap, and blue's remedy — each count names its own state's check alone. A
// citation a lens's reading finds against is on no failed check: that reading is the lens's debt
// (readingdebt_test.go), and the ledger reads the citation as holding.
func TestTheFailedCheckDrivesKeyOnTheState(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue, lensU, lensL = "blue-respond", "red-lens-logic", "red-lens-evidence"
	r.sit("blue", blue)
	cu := r.citeAs(t, blue, "§ fuzz sentence", "/state-unread")
	cf := r.citeAs(t, blue, "The cost is rising over time", "/state-refuted")

	// U READS NOTHING OF cu's SOURCE, and carries the check by its gap about the anchor.
	r.sit("lens", lensU)
	r.do("verify", lensU).set("--anchor", cu).set("--as", "unreachable").set("--confidence", "high").
		set("--reason", "fuzz: the host did not answer").admit()
	gu := r.mintAbout(t, lensU, cu)
	// L'S READING OF cf FINDS AGAINST IT: no failed check, and L's debt until its mint on the reading.
	r.sit("lens", lensL)
	if r.readCited(lensL, cf, "refutes") == 0 {
		t.Fatalf("L's refuting reading: %s", r.actErr)
	}
	if _, stands := r.debtOf(lensL, cf); !stands {
		t.Fatalf("L's refuting reading of %s left no reading debt", cf)
	}
	state := func(g string) record.CheckState {
		b, ok := r.standingFailedCheckOf(g)
		if !ok {
			t.Fatalf("no failed check is carried by %s", g)
		}
		return b.CheckState()
	}
	if s := state(gu); s != record.CheckUnread {
		t.Fatalf("%s carries a %s check, want unread", gu, s)
	}
	bs, err := record.PassBlockers(r.run())
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bs {
		if b.Kind == record.BlockerFailedCheck && b.Subject == cf {
			t.Fatalf("%s, which L's reading finds against, stands on a failed check in state %q", cf, b.CheckState())
		}
	}
	res := r.mintOnReading(lensL, classContradicts, aboutArgs(cf), 0, "about")
	if res.gap == "" || !r.debtDischargedBy(lensL, cf, "mint", "L's mint on its reading") {
		t.Fatalf("L's mint on its reading: %q (%s)", res.out, r.actErr)
	}

	// U'S HALF, a sitting at a time: its unreadable repeat is driven to its refusal on a coin.
	for i := 0; i < 20; i++ {
		r.sit("lens", lensU)
		r.answerFailed(lensU)
	}
	// THE CLOSE OF THE GAP IS REFUSED while its check stands, tallied by the check's state.
	b, _ := r.standingFailedCheckOf(gu)
	r.refuseStandingClose(lensU, gu, "repaired", b)
	// BLUE'S REMEDY: the unread citation retired.
	r.sit("blue", blue)
	r.fixFailed(blue)

	if r.actErr != "" {
		t.Fatalf("a drive answered otherwise than it expects: %s", r.actErr)
	}
	for _, c := range []struct {
		name      string
		got, want int
	}{
		{"unreadRetires", r.unreadRetires, 1},
		{"unreadClosesRefused", r.unreadClosesRefused, 1},
	} {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d — each counts its own state's check alone", c.name, c.got, c.want)
		}
	}
	if r.unreadRepeats == 0 {
		t.Error("U's unreadable repeat was never driven to its refusal in 20 sittings — the drive this pins is unexercised")
	}
	if _, ok := r.standingFailedCheckOf(gu); ok {
		t.Errorf("a failed check carried by %s still stands after blue's remedy", gu)
	}
}

// mintAbout is lens seatID's gap about anchor, whose check of seatID's failed: the gap that carries
// the check to blue.
func (r *runner) mintAbout(t *testing.T, seatID, anchorID string) string {
	t.Helper()
	out, ok := r.admit("--json", "mint", "--seat-id", seatID, "--problem", "fuzz: no lens could read the source "+anchorID+" cites",
		"--check-kind", "document", "--check", "the citation is retired or cites a source a lens can read",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--class", pick(r.rng, mintPoolClasses),
		"--about-kind", "anchor", "--about", anchorID, "--reason", "fuzz: carrying a failed check to blue")
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		t.Fatalf("%s's gap about %s: %s (%s)", seatID, anchorID, out, r.actErr)
	}
	r.noteMint(seatID, env.Result.GapID)
	return env.Result.GapID
}

// A SATISFIED GAP WHOSE CLOSE THE RECORD REFUSED OVER A STANDING CHECK IS NOT WORK LEFT UNFINISHED.
// The terminal-fate oracle read every satisfied gap red sat on and left open as a repair the board
// never recorded; an unread check the gap carries refuses its close until blue's remedy or a backing
// reading, so a run that ends there holds it open by the record's rule. Run outside the release
// gate: an APPLY gap red sat on, its close refused over the check, is not done-but-open; once red's
// next close attempt is not refused over one, it is.
func TestASatisfiedGapHeldOpenByAStandingCheckIsNotDoneButOpen(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue, lensL = "blue-respond", "red-lens-evidence"
	r.sit("blue", blue)
	cite := r.citeAs(t, blue, "The cost is rising over time", "/held-open")
	r.sit("lens", lensL)
	r.do("verify", lensL).set("--anchor", cite).set("--as", "unreachable").set("--confidence", "high").
		set("--reason", "fuzz: the host did not answer").admit()
	out, ok := r.admit("--json", "mint", "--seat-id", lensL, "--problem", "fuzz: no lens could read the source "+cite+" cites",
		"--check-kind", "document", "--check", "the citation is retired or cites a source a lens can read", "--fix", dirApply,
		"--severity", "high", "--likelihood", "high", "--impact", "high", "--class", pick(r.rng, mintPoolClasses),
		"--about-kind", "anchor", "--about", cite, "--reason", "fuzz: carrying a failed check to blue")
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		t.Fatalf("L's gap about the unread citation: %s (%s)", out, r.actErr)
	}
	g := env.Result.GapID
	r.noteMint(lensL, g)
	r.presented.add(g)
	r.evaluated.add(g)
	b, stands := r.standingFailedCheckOf(g)
	if !stands {
		t.Fatalf("%s carries no standing failed check", g)
	}
	r.refuseStandingClose(lensL, g, "repaired", b)
	if r.actErr != "" {
		t.Fatalf("the refused close of %s: %s", g, r.actErr)
	}
	if r.doneButOpen(g, dirApply, true) {
		t.Errorf("%s, its close refused over the check standing on it, reads as work done and never finished", g)
	}
	// RED'S NEXT ATTEMPT IS ITS OWN: an attempt not refused over a check leaves the gap judged again.
	r.heldByCheck.remove(g)
	if !r.doneButOpen(g, dirApply, true) {
		t.Errorf("%s, open with no refusal over a check, does not read as done-but-open", g)
	}
}
