package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// EVERY ACT SAYS WHERE THE SEAT NOW STANDS (#1122).
//
// #1163 delivered the work list at dispatch. That ended the OPENING read and made an empty sitting
// free, and it could not end the rest: SubagentStart fires once per dispatch, and a seat's own acts
// move its list. So the verdict rides every write.
//
// DRIVEN THROUGH THE CLI, because the subject is what a seat is handed. The unit of this feature is
// not a function that returns a struct; it is the bytes a write verb prints.
func TestEveryWriteSaysWhereTheSeatStands(t *testing.T) {
	runDir := seatRun(t)

	// A WRITE CARRIES IT. `log` is the smallest write on every role's surface.
	out, err := run(t, "log", "--run", runDir, "--seat-id", lensSeat,
		"--reason", "nothing blocked me", "--type", "defect")
	if err != nil {
		t.Fatal(err)
	}
	if !saysWhereYouStand(out) {
		t.Errorf("a write did not say where the seat stands:\n%s", out)
	}
	// AFTER the verb's own line, never instead of it.
	if firstLines(out, 1) == "" || saysWhereYouStand(firstLines(out, 1)) {
		t.Errorf("the standing displaced the act's own result:\n%s", out)
	}

	// A READ CARRIES NOTHING. This is a guard on the ROUTING rather than on a check in standing.go:
	// a read is rendered by renderView and returns before Emit, so there is no code path that could
	// attach one. Asserted anyway, because that routing is what the feature rests on — and a read
	// that started going through Emit would append a summary of the list to the list itself.
	work, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
	if err != nil {
		t.Fatal(err)
	}
	if saysWhereYouStand(work) || strings.Contains(work, "WHERE YOU STAND") {
		t.Errorf("a read carried a standing block:\n%s", work)
	}
}

// THE TWO FORMS ANSWER THE SAME QUESTION. A consumer parsing the envelope and a seat reading the
// line must not be told different things — they are one computation in Emit, and this holds them to
// each other rather than to a literal.
func TestTheStandingIsTheSameInBothForms(t *testing.T) {
	runDir := seatRun(t)
	mintGap(t, runDir, "standing-gap", "spec_underspecification")

	human, err := run(t, "log", "--run", runDir, "--seat-id", lensSeat,
		"--reason", "nothing blocked me", "--type", "defect")
	if err != nil {
		t.Fatal(err)
	}
	// A SECOND RUN, because the log is now written and the standing may legitimately have moved;
	// what must agree is the two renderings of the SAME invocation, so this one is read as JSON.
	raw, err := run(t, "log", "--run", runDir, "--seat-id", lensSeat, "--json",
		"--reason", "still nothing blocked me", "--type", "defect")
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Standing *struct {
			Complete bool     `json:"complete"`
			Blocking []string `json:"blocking"`
			Open     int      `json:"open"`
		} `json:"standing"`
	}
	if e := json.Unmarshal([]byte(raw), &env); e != nil {
		t.Fatalf("the envelope did not parse (%v):\n%s", e, raw)
	}
	if env.Standing == nil {
		t.Fatal("a write's JSON envelope carries no standing — a consumer would have to run a second command for it")
	}
	// The human form of the FIRST call said the same thing about blocking that the envelope says.
	blockedHuman := strings.Contains(human, "STILL TO CLEAR")
	if blockedHuman == env.Standing.Complete {
		t.Errorf("the two forms disagree: human blocked=%v, envelope complete=%v\nhuman:\n%s\njson:\n%s",
			blockedHuman, env.Standing.Complete, human, raw)
	}
	if env.Standing.Open == 0 {
		t.Errorf("the envelope reports no items open at all, which this board does not have:\n%s", raw)
	}
}

// A BLOCKED SEAT IS TOLD WHAT BLOCKS IT, not merely that it is blocked. A count is another read: a
// seat handed "1 item blocks you" has to ask which, which is the call this feature exists to remove.
func TestABlockedSeatIsToldWhatBlocksIt(t *testing.T) {
	runDir := seatRun(t)
	// The chair's terminal act is missing until it records a verdict, which is a blocking item with
	// text of its own — so this drives the blocked arm without inventing a board state.
	out, err := run(t, "position", "--run", runDir, "--seat-id", "red-chair",
		"--reason", "red's account of this epoch")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "STILL TO CLEAR") {
		t.Fatalf("the chair owes its terminal act and was not told it is held:\n%s", out)
	}
	// THE TERMINAL AND THE HELD STATE MUST NOT BOTH READ AS PERMISSION. "ALL CLEAR" is the one
	// sentence that grants it, and a held seat must not see it anywhere in the block.
	if strings.Contains(out, "ALL CLEAR") {
		t.Errorf("a held seat was also told it was clear:\n%s", out)
	}
	if !strings.Contains(out, "·") {
		t.Errorf("the seat was told it is blocked and not by what:\n%s", out)
	}
}

// THE COUNT IS THE SEAT'S OWN DEBT. A chair's list holds the PASS gate's blockers, and some are
// another seat's to clear — an open material gap its minter closes. The chair was told "10 STILL TO
// CLEAR before you may end your turn" over two of those and ended its turn, correctly, with all ten
// open (the 2026-10-12 smoke, chair sitting 8). So the count covers what the chair clears, the
// sentence says those hold a PASS, and what another seat clears is stated after as the board's state.
func TestTheChairsDigestCountsOnlyWhatTheChairClears(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "a-material-gap", "spec_underspecification")
	gapLine := "gap " + id + " is open and material"

	out, err := run(t, "position", "--run", runDir, "--seat-id", "red-chair", "--reason", "red's account of this epoch")
	if err != nil {
		t.Fatal(err)
	}
	head, board, found := strings.Cut(out, "ANOTHER SEAT'S TO CLEAR")
	if !found {
		t.Fatalf("the chair's digest does not set apart what another seat clears:\n%s", out)
	}
	if strings.Contains(head, gapLine) || !strings.Contains(board, gapLine) {
		t.Errorf("a gap its minter closes is counted in the chair's own debt, want it under the board's state:\n%s", out)
	}
	if !strings.Contains(head, "your terminal act is missing") {
		t.Errorf("the chair's own item is not in its own count:\n%s", out)
	}
	if strings.Contains(out, "before you may end your turn") {
		t.Errorf("the chair's PASS blockers are said to hold its turn:\n%s", out)
	}
	if !strings.Contains(head, "STILL TO CLEAR before the run can PASS") {
		t.Errorf("the chair is not told its items hold a PASS:\n%s", out)
	}

	// THE ENVELOPE SAYS THE SAME: `blocking` is the chair's own, and the rest is beside it.
	raw, err := run(t, "log", "--run", runDir, "--seat-id", "red-chair", "--json", "--reason", "nothing blocked me", "--type", "defect")
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Standing struct {
			Blocking          []string `json:"blocking"`
			AnotherSeatClears []string `json:"another_seat_clears"`
		} `json:"standing"`
	}
	if e := json.Unmarshal([]byte(raw), &env); e != nil {
		t.Fatalf("the envelope did not parse (%v):\n%s", e, raw)
	}
	if strings.Contains(strings.Join(env.Standing.Blocking, "\n"), gapLine) || !strings.Contains(strings.Join(env.Standing.AnotherSeatClears, "\n"), gapLine) {
		t.Errorf("the envelope counts another seat's blocker as the chair's: %s", raw)
	}

	// A SEAT WHOSE ITEMS ARE ALL ITS OWN IS TOLD THEY HOLD ITS TURN, as before.
	blue, err := run(t, "log", "--run", runDir, "--seat-id", "blue-respond", "--reason", "nothing blocked me", "--type", "defect")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(blue, "ANOTHER SEAT'S TO CLEAR") || strings.Contains(blue, "before a PASS") {
		t.Errorf("a seat that is not the chair is shown the chair's sentence:\n%s", blue)
	}
}

// A MOTION CARRIES IT TOO, and it is here because the motion verbs reach Emit by a different route:
// they are built from a bare cobra.Command rather than through NewKeyed. Filing and ruling
// are writes, a bench ruling is the act most likely to flip `complete`, and nothing else in this file
// drives a verb built that way.
func TestAMotionAlsoSaysWhereTheSeatStands(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "motion-gap", "spec_underspecification")
	out, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond",
		"--id", id, "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation")
	if err != nil {
		t.Fatalf("motion grade file: %v\n%s", err, out)
	}
	if !saysWhereYouStand(out) {
		t.Errorf("a motion is a write and said nothing about where the seat stands:\n%s", out)
	}
}

// saysWhereYouStand matches each opening of the standing block, in ONE place — the phrases are
// what the block is, and a test spelling them separately drifts from the renderer one leg at a time.
func saysWhereYouStand(s string) bool {
	return strings.Contains(s, "ALL CLEAR") || strings.Contains(s, "STILL TO CLEAR") || strings.Contains(s, "NOTHING HERE IS YOURS TO CLEAR")
}
