package fuzz

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FAILED-CHECK DRIVES KEY ON THE CHECK'S STATE (plans/contradiction-is-a-gap.md §III.11, partner
// §V.5.q). A found-against check of a citation has an empty Detail, as an unread one does, so a drive
// that told them apart by Detail would read it as unread: drive an `unreachable` repeat the record
// admits, retire it as unread, and tally its refused close as an unread one. Run outside the release
// gate: one unread check by lens U carried by G-u, one found-against check by lens L carried by G-f
// (L's refuting reading through the built binary); then L's and U's halves, the refused closes of both
// gaps, and blue's remedies — each count names its own state's check alone.
func TestTheFailedCheckDrivesKeyOnTheState(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue, lensU, lensL = "blue-respond", "red-lens-logic", "red-lens-evidence"
	r.sit("blue", blue)
	cite := func(quote, path string) string {
		t.Helper()
		out, ok := r.do("cite", blue).set("--quote", quote).set("--url", sourceURL(path)).set("--title", sourceTitle).
			set("--source-span", sourceBody(path)).bare("--json").admit()
		var env struct {
			Result struct {
				Label string `json:"label"`
			} `json:"result"`
		}
		if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.Label == "" {
			t.Fatalf("blue's cite of %q: %s (%s)", quote, out, r.actErr)
		}
		return env.Result.Label
	}
	cu := cite("§ fuzz sentence", "/state-unread")
	cf := cite("The cost is rising over time", "/state-found-against")

	// U READS NOTHING OF cu's SOURCE, and carries the check by its gap about the anchor.
	r.sit("lens", lensU)
	r.do("verify", lensU).set("--anchor", cu).set("--as", "unreachable").set("--confidence", "high").
		set("--reason", "fuzz: the host did not answer").admit()
	gu := r.mintAbout(t, lensU, cu)
	// L'S READING OF cf FINDS AGAINST IT, and mints G-f in its own act.
	r.sit("lens", lensL)
	out, ok := r.do("verify", lensL).set("--anchor", cf).set("--as", "refutes").set("--confidence", "high").
		set("--reason", "fuzz: the source says the cost fell").bare("--json").
		set("--class", "claim_contradicts_source").set("--problem", "fuzz: the source says the cost fell").
		set("--check-kind", "document").set("--check", "the sentence says what the source says").
		set("--severity", "high").set("--likelihood", "high").set("--impact", "high").admit()
	res, parsed := readingResultOf(out)
	if !ok || !parsed || res.Gap == nil || !res.Gap.Minted {
		t.Fatalf("L's refuting reading minted no gap: %s (%s)", out, r.actErr)
	}
	gf := res.Gap.ID
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
	if s := state(gf); s != record.CheckFoundAgainst {
		t.Fatalf("%s carries a %s check, want found_against", gf, s)
	}

	// EACH LENS'S HALF, a sitting at a time: U's unreadable repeat is driven to its refusal on a coin;
	// L's found-against check owes it nothing, so no `unreachable` of L's is driven as a refused repeat.
	for i := 0; i < 20; i++ {
		r.sit("lens", lensL)
		r.answerFailed(lensL)
		r.sit("lens", lensU)
		r.answerFailed(lensU)
	}
	// THE CLOSE OF EACH GAP IS REFUSED while its check stands, tallied by the check's state.
	for _, g := range []struct{ closer, id string }{{lensU, gu}, {lensL, gf}} {
		b, _ := r.standingFailedCheckOf(g.id)
		r.refuseStandingClose(g.closer, g.id, "repaired", b)
	}
	// BLUE'S REMEDY OF EACH: the unread citation retired, the found-against one edited or retired.
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
		{"foundAgainstClosesRefused", r.foundAgainstClosesRefused, 1},
		{"foundAgainstEdits + foundAgainstRetires", r.foundAgainstEdits + r.foundAgainstRetires, 1},
	} {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d — each counts its own state's check alone", c.name, c.got, c.want)
		}
	}
	if r.unreadRepeats == 0 {
		t.Error("U's unreadable repeat was never driven to its refusal in 20 sittings — the drive this pins is unexercised")
	}
	for _, g := range []string{gu, gf} {
		if _, ok := r.standingFailedCheckOf(g); ok {
			t.Errorf("a failed check carried by %s still stands after blue's remedy", g)
		}
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
