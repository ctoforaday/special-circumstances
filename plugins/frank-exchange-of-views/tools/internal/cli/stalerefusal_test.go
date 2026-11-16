package cli

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN ACT ON A GAP THAT MOVED SINCE THE SEAT'S SITTING OPENED IS REFUSED, AND SHOWS WHAT MOVED (gblock
// ruling B, 2026-10-19). On universe-m30 lenses sat concurrently and two filed defects against the
// board their sittings opened with: the adversary's gaps had been closed by two other lenses after its
// sitting opened, and the dark-side lens read a citation as never verified after another lens had
// verified it. Those were log entries, which are the operator's and out of scope; these are the same
// shapes on the acts that name a gap.

const staleLens = "red-lens-dark-side"

// sit opens a new sitting for seat: the read point every check below measures from.
func sit(t *testing.T, runDir, seat string) {
	t.Helper()
	args := []string{"register", "--run", runDir, "--seat-id", seat}
	if record.SeatOwesOccasion(seat) {
		args = append(args, "--occasion", "docket")
	}
	if _, err := run(t, args...); err != nil {
		t.Fatalf("register %s: %v", seat, err)
	}
}

func closeRepaired(t *testing.T, runDir, gap string) {
	t.Helper()
	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:report.md",
		"--reason", "re-read the sentence; the repair holds"); err != nil {
		t.Fatalf("close %s: %v", gap, err)
	}
}

func regrade(t *testing.T, runDir, seat, gap, severity string) error {
	t.Helper()
	_, err := run(t, "regrade", "--run", runDir, "--seat-id", seat, "--id", gap, "--severity", severity,
		"--reason", "re-audited: the acceptance check still fails on the sentence as it stands")
	return err
}

// The dark-side's shape on a gap: another seat's act on the gap lands after the seat's sitting opened.
// The refusal names it; a change AFTER the refusal is refused again, so the read point moved to what
// was shown and no further.
func TestAnotherSeatsActSinceTheSittingOpenedRefusesAndALaterOneRefusesAgain(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "argued-under", "false_universal")
	sit(t, runDir, "blue-respond")
	if err := regrade(t, runDir, lensSeat, gap, "high"); err != nil {
		t.Fatal(err)
	}
	regradedAt := lastRegradeID(t, runDir)
	err := gradeMotion(t, runDir, gap)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("event %d, %s: regraded %s", regradedAt, lensSeat, gap)) {
		t.Fatalf("a motion on a gap another seat regraded after this sitting opened was not refused naming that regrade: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MOTION); n != 0 {
		t.Errorf("the refused motion recorded %d motion(s)", n)
	}
	if st := lastBody(t, runDir, &recordpb.Stale{}); st.GetGapId() != gap || st.GetThrough() != regradedAt {
		t.Errorf("the refusal recorded read point %s through %d, want %s through %d", st.GetGapId(), st.GetThrough(), gap, regradedAt)
	}
	// The minting lens regrades again at a sitting opened after the refusal.
	sit(t, runDir, lensSeat)
	if err := regrade(t, runDir, lensSeat, gap, "medium"); err != nil {
		t.Fatalf("the minting lens's second regrade: %v", err)
	}
	if err := gradeMotion(t, runDir, gap); err == nil || !strings.Contains(err.Error(), "regraded "+gap) {
		t.Fatalf("a regrade after the first refusal did not refuse the re-filed motion: %v", err)
	}
	if err := gradeMotion(t, runDir, gap); err != nil {
		t.Fatalf("the motion filed a third time, with nothing new since the second refusal, was refused: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_STALE); n != 2 {
		t.Errorf("%d stale refusals recorded, want 2", n)
	}
}

// gradeMotion is blue's grade motion on gap: an act that names a gap, from a seat that did not mint it.
func gradeMotion(t *testing.T, runDir, gap string) error {
	t.Helper()
	_, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond", "--id", gap,
		"--dimension", "severity", "--proposed", "low", "--reason", "the defect is cosmetic")
	return err
}

// A seat's own acts are never news to it: the minting lens regrades its gap, which is a change to it,
// and then files a grade motion on it in the same sitting.
func TestASeatsOwnActsOnAGapNeverRefuseIt(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "own", "false_universal")
	if err := regrade(t, runDir, lensSeat, gap, "high"); err != nil {
		t.Fatalf("the minting lens's regrade: %v", err)
	}
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", lensSeat, "--id", gap,
		"--dimension", "likelihood", "--proposed", "low", "--reason", "the harm needs two failures"); err != nil {
		t.Fatalf("the lens's motion after its own regrade: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_STALE); n != 0 {
		t.Errorf("%d stale refusals recorded on a seat's own acts", n)
	}
}

// An unchanged gap takes the act on the first call: a sitting that opened after the change has read it.
func TestAnActOnAGapUnchangedSinceTheSittingOpenedLands(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "unchanged", "false_universal")
	if err := regrade(t, runDir, lensSeat, gap, "high"); err != nil {
		t.Fatal(err)
	}
	sit(t, runDir, "blue-respond")
	if err := gradeMotion(t, runDir, gap); err != nil {
		t.Fatalf("a motion on a gap that has not changed since the sitting opened was refused: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_STALE); n != 0 {
		t.Errorf("%d stale refusals recorded on a fresh read", n)
	}
}

// lastRegradeID is the events.id of the newest regrade on the record.
func lastRegradeID(t *testing.T, runDir string) int64 {
	t.Helper()
	evs, win, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_REGRADE)
	if err != nil || len(evs) == 0 {
		t.Fatalf("no regrade on the record: %v", err)
	}
	return win.Of(evs[len(evs)-1]).ID
}

// Every act that names a gap and asserts something about it carries the refusal: the regrade, the
// grade motion and the docket motion, each refused after another seat moved the gap and each recorded
// on the second call.
func TestEveryActThatNamesAGapIsRefusedWhileTheGapMovedUnderIt(t *testing.T) {
	for _, c := range []struct {
		name, seat string
		args       func(runDir, gap string) []string
	}{
		{"regrade", lensSeat, func(runDir, gap string) []string {
			return []string{"regrade", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--severity", "high", "--reason", "re-audited"}
		}},
		{"motion grade file", "blue-respond", func(runDir, gap string) []string {
			return []string{"motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond", "--id", gap,
				"--dimension", "severity", "--proposed", "low", "--reason", "the defect is cosmetic"}
		}},
		{"motion docket file", "red-chair", func(runDir, gap string) []string {
			return []string{"motion", "docket", "file", "--run", runDir, "--seat-id", "red-chair", "--id", gap,
				"--reason", "the lens and blue have exchanged twice and neither moved"}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := seatRun(t)
			gap := mintGap(t, runDir, "moved-"+strings.ReplaceAll(c.name, " ", "-"), "false_universal")
			sit(t, runDir, c.seat)
			// The move is another seat's: the minting lens's regrade, or, where the act is the minting
			// lens's own, the bench's ruling on a docket motion on the gap.
			moved := lensSeat + ": regraded " + gap
			if c.seat == lensSeat {
				m := docketFile(t, runDir, "red-chair", gap, "the lens and blue have exchanged twice and neither moved")
				ruleDocket(t, runDir, m, "remanded")
				moved = "the bench ruled " + gap + " remanded in " + m
			} else if err := regrade(t, runDir, lensSeat, gap, "high"); err != nil {
				t.Fatal(err)
			}
			args := c.args(runDir, gap)
			_, err := run(t, args...)
			if err == nil || !strings.Contains(err.Error(), moved) {
				t.Fatalf("%s on a gap another seat moved after the sitting opened was not refused naming %q: %v", c.name, moved, err)
			}
			if _, err := run(t, args...); err != nil {
				t.Fatalf("%s filed again after the refusal was refused: %v", c.name, err)
			}
		})
	}
}

// THE CHECK IS MADE ON THE RECORD THE ACT LANDS ON: a regrade committing at the instant the motion's
// write opens its transaction is a change the guard reads, as if it had come first.
func TestARegradeLandingAsTheMotionWritesIsSeen(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "raced", "false_universal")
	sit(t, runDir, "blue-respond")
	regradeAct := recordtest.At(t, lensSeat, "", &recordpb.Regrade{GapId: proto.String(gap),
		Severity: recordpb.Grade_GRADE_HIGH.Enum(), Basis: proto.String("re-audited")})
	err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{regradeAct}, func() error {
		return gradeMotion(t, runDir, gap)
	})
	if err == nil || !strings.Contains(err.Error(), "regraded "+gap) {
		t.Fatalf("a regrade that landed as the motion wrote was not seen: %v", err)
	}
}

// CONCURRENT SEATS, FUZZED. Seats open sittings, the minting lens closes and regrades, other seats
// file grade motions on the gaps, in a seeded random interleaving. The oracle is independent of the
// tool's fold: for every recorded grade motion, no close or regrade on that gap by another seat lies
// between the filer's read point (its sitting's opening, or the last stale refusal on that gap in the
// sitting) and the motion. And no refusal loops: a refused motion filed again with nothing new in
// between is recorded.
func TestConcurrentSeatsFileNoMotionAgainstAGapThatMovedUnderThem(t *testing.T) {
	filers := []string{"blue-respond", "red-lens-logic", "red-lens-adversary"}
	for seed := int64(1); seed <= 4; seed++ {
		t.Run(fmt.Sprint("seed", seed), func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			runDir := seatRun(t)
			var gaps []string
			for i := 0; i < 3; i++ {
				gaps = append(gaps, mintGap(t, runDir, fmt.Sprintf("fz%d", i), "false_universal"))
			}
			closed := map[string]bool{}
			for _, f := range filers {
				sit(t, runDir, f)
			}
			refusals := 0
			for step := 0; step < 40; step++ {
				gap := gaps[rng.Intn(len(gaps))]
				switch r := rng.Intn(10); {
				case r < 2:
					sit(t, runDir, filers[rng.Intn(len(filers))])
				case r < 3 && !closed[gap]:
					closeRepaired(t, runDir, gap)
					closed[gap] = true
				case r < 4 && !closed[gap]:
					sev := []string{"low", "medium", "high"}[rng.Intn(3)]
					// One regrade per gap per sitting is the record's own rule; a refused one is either.
					if err := regrade(t, runDir, lensSeat, gap, sev); err != nil && !strings.Contains(err.Error(), "changed after your sitting opened") && !strings.Contains(err.Error(), "already recorded a regrade") {
						t.Fatalf("regrade: %v", err)
					}
				case !closed[gap]:
					f := filers[rng.Intn(len(filers))]
					file := func() error {
						_, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", f, "--id", gap,
							"--dimension", "severity", "--proposed", "low", "--reason", fmt.Sprintf("step %d", step))
						return err
					}
					err := file()
					// A motion proposing the grade the gap already holds is refused for that, whatever moved.
					if err == nil || strings.Contains(err.Error(), "asks for no change") {
						continue
					}
					if !strings.Contains(err.Error(), "changed after your sitting opened") {
						t.Fatalf("motion grade file: %v", err)
					}
					refusals++
					if err := file(); err != nil {
						t.Fatalf("a refused motion filed again with nothing new since was refused — a refusal loop: %v", err)
					}
				}
			}
			if bad := staleMotionsOn(t, runDir); len(bad) > 0 {
				t.Errorf("%d motion(s) recorded against a gap that moved under them:\n%s", len(bad), strings.Join(bad, "\n"))
			}
			t.Logf("seed %d: %d stale refusal(s), each re-filed", seed, refusals)
		})
	}
}

// staleMotionsOn is the oracle: every grade motion with another seat's close or regrade on its gap
// after the filer's read point.
func staleMotionsOn(t *testing.T, runDir string) []string {
	t.Helper()
	evs, win, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_CLOSE,
		recordpb.EventType_EVENT_TYPE_REGRADE, recordpb.EventType_EVENT_TYPE_MOTION, recordpb.EventType_EVENT_TYPE_STALE)
	if err != nil {
		t.Fatal(err)
	}
	type act struct {
		id   int64
		seat string
	}
	on := map[string][]act{}
	var bad []string
	motions := 0
	for _, e := range evs {
		id := win.Of(e).ID
		gap := ""
		switch {
		case e.GetClose() != nil:
			gap = e.GetClose().GetGapId()
		case e.GetRegrade() != nil:
			gap = e.GetRegrade().GetGapId()
		case e.GetMotion().GetGrade() != nil:
			gap = e.GetMotion().GetGrade().GetGapId()
			motions++
			sitting := win.Of(e).SittingID
			read := sitting
			for _, p := range evs {
				if s := p.GetStale(); s != nil && s.GetGapId() == gap && p.GetSeatId() == e.GetSeatId() && win.Of(p).SittingID == sitting && win.Of(p).ID < id {
					read = max(read, s.GetThrough())
				}
			}
			for _, a := range on[gap] {
				if a.id > read && a.seat != e.GetSeatId() {
					bad = append(bad, fmt.Sprintf("  event %d (%s) on %s, after event %d by %s (read point %d)", id, e.GetSeatId(), gap, a.id, a.seat, read))
				}
			}
		}
		// A filed motion is no change to the gap; a close or a regrade is.
		if gap != "" && e.GetMotion() == nil {
			on[gap] = append(on[gap], act{id, e.GetSeatId()})
		}
	}
	if motions == 0 {
		t.Fatal("the oracle read no grade motion — it measured nothing")
	}
	return bad
}
