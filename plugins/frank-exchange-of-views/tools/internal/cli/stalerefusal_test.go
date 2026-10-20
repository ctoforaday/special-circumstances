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

func findingAbout(t *testing.T, runDir, seat, key, gap string) error {
	t.Helper()
	_, err := run(t, "finding", "--run", runDir, "--seat-id", seat, "--key", key,
		"--about-kind", "gap", "--about", gap, "--reason", "the gap's fix does not reach the defect it names",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium")
	return err
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

// lastCloseID is the events.id of the newest close on the record.
func lastCloseID(t *testing.T, runDir string) int64 {
	t.Helper()
	evs, win, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_CLOSE)
	if err != nil || len(evs) == 0 {
		t.Fatalf("no close on the record: %v", err)
	}
	return win.Of(evs[len(evs)-1]).ID
}

// The adversary's shape: its sitting opens, another seat closes the gap, and it then files a finding
// disputing the gap as if it were open. Refused, naming the close by its event; recorded on a second
// call, which reads from the refusal.
func TestAFindingAboutAGapClosedSinceTheSittingOpenedIsRefusedAndRefilesOnce(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "closed-under", "carry-vs-close")
	sit(t, runDir, staleLens)
	closeRepaired(t, runDir, gap)
	closedAt := lastCloseID(t, runDir)

	err := findingAbout(t, runDir, staleLens, "F1", gap)
	if err == nil {
		t.Fatal("a finding about a gap closed after the seat's sitting opened was recorded")
	}
	for _, want := range []string{gap, fmt.Sprintf("event %d, %s: closed %s as repaired", closedAt, lensSeat, gap)} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not show %q:\n%v", want, err)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING); n != 0 {
		t.Errorf("the refused finding recorded %d finding(s)", n)
	}
	st := lastBody(t, runDir, &recordpb.Stale{})
	if st.GetGapId() != gap || st.GetThrough() != closedAt {
		t.Errorf("the refusal recorded read point %s through %d, want %s through %d", st.GetGapId(), st.GetThrough(), gap, closedAt)
	}
	if err := findingAbout(t, runDir, staleLens, "F1", gap); err != nil {
		t.Fatalf("the same finding, filed again after the refusal showed the close, was refused: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_STALE); n != 1 {
		t.Errorf("%d stale refusals recorded, want 1", n)
	}
}

// The dark-side's shape on a gap: another lens's act on the gap lands after the seat's sitting opened.
// The refusal names it; a change AFTER the refusal is refused again, so the read point moved to what
// was shown and no further.
func TestAnotherSeatsFindingSinceTheSittingOpenedRefusesAndALaterOneRefusesAgain(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "argued-under", "cross-seat-visibility")
	sit(t, runDir, staleLens)
	sit(t, runDir, "red-lens-logic")
	if err := findingAbout(t, runDir, "red-lens-logic", "L1", gap); err != nil {
		t.Fatal(err)
	}
	err := findingAbout(t, runDir, staleLens, "D1", gap)
	if err == nil || !strings.Contains(err.Error(), "red-lens-logic's finding") {
		t.Fatalf("a finding about a gap another lens argued after this sitting opened was not refused naming that finding: %v", err)
	}
	// The minting lens answers the logic finding at a sitting opened after it.
	sit(t, runDir, lensSeat)
	if err := regrade(t, runDir, lensSeat, gap, "high"); err != nil {
		t.Fatalf("the minting lens's regrade after the logic finding: %v", err)
	}
	if err := findingAbout(t, runDir, staleLens, "D1", gap); err == nil || !strings.Contains(err.Error(), "regraded "+gap) {
		t.Fatalf("a regrade after the first refusal did not refuse the re-filed finding: %v", err)
	}
	if err := findingAbout(t, runDir, staleLens, "D1", gap); err != nil {
		t.Fatalf("the finding filed a third time, with nothing new since the second refusal, was refused: %v", err)
	}
}

// An unchanged gap files on the first call, and a sitting that opened after the change has read it.
func TestAFindingAboutAGapUnchangedSinceTheSittingOpenedFiles(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "unchanged", "cross-seat-visibility")
	closeRepaired(t, runDir, gap)
	sit(t, runDir, staleLens)
	if err := findingAbout(t, runDir, staleLens, "U1", gap); err != nil {
		t.Fatalf("a finding about a gap that has not changed since the sitting opened was refused: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_STALE); n != 0 {
		t.Errorf("%d stale refusals recorded on a fresh read", n)
	}
}

// A seat's own acts are never news to it: a lens files two findings about one gap in one sitting.
func TestASeatsOwnActsOnAGapNeverRefuseIt(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "own", "dispute-to-regrade")
	sit(t, runDir, staleLens)
	for _, key := range []string{"O1", "O2"} {
		if err := findingAbout(t, runDir, staleLens, key, gap); err != nil {
			t.Fatalf("finding %s after the lens's own finding on the gap: %v", key, err)
		}
	}
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
			gap := mintGap(t, runDir, "moved-"+strings.ReplaceAll(c.name, " ", "-"), "grade-dispute-visibility")
			sit(t, runDir, c.seat)
			sit(t, runDir, "red-lens-logic")
			if err := findingAbout(t, runDir, "red-lens-logic", "M1", gap); err != nil {
				t.Fatal(err)
			}
			args := c.args(runDir, gap)
			_, err := run(t, args...)
			if err == nil || !strings.Contains(err.Error(), "red-lens-logic's finding") {
				t.Fatalf("%s on a gap another seat argued after the sitting opened was not refused naming it: %v", c.name, err)
			}
			if _, err := run(t, args...); err != nil {
				t.Fatalf("%s filed again after the refusal was refused: %v", c.name, err)
			}
		})
	}
}

// THE CHECK IS MADE ON THE RECORD THE ACT LANDS ON: a close committing at the instant the finding's
// write opens its transaction is a change the guard reads, as if it had come first.
func TestACloseLandingAsTheFindingWritesIsSeen(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "raced", "carry-vs-close")
	sit(t, runDir, staleLens)
	closeAct := recordtest.At(t, lensSeat, "", &recordpb.Close{GapId: proto.String(gap),
		ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(), AnchorSeat: proto.String("L1"),
		AnchorTool: proto.String("git show"), AnchorTarget: proto.String("7bc501e:report.md"), Prose: proto.String("the repair holds")})
	err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{closeAct}, func() error {
		return findingAbout(t, runDir, staleLens, "R1", gap)
	})
	if err == nil || !strings.Contains(err.Error(), "closed "+gap) {
		t.Fatalf("a close that landed as the finding wrote was not seen: %v", err)
	}
}

// CONCURRENT LENSES, FUZZED. Seats open sittings, the minting lens closes and regrades, other lenses
// file findings about the gaps, in a seeded random interleaving. The oracle is independent of the
// tool's fold: for every recorded finding about a gap, no close, regrade or finding on that gap by
// another seat lies between the filer's read point (its sitting's opening, or the last stale refusal
// on that gap in the sitting) and the finding. And no refusal loops: a refused finding filed again with
// nothing new in between is recorded.
func TestConcurrentLensesFileNoFindingAgainstAGapThatMovedUnderThem(t *testing.T) {
	lenses := []string{staleLens, "red-lens-logic", "red-lens-adversary"}
	for seed := int64(1); seed <= 4; seed++ {
		t.Run(fmt.Sprint("seed", seed), func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			runDir := seatRun(t)
			var gaps []string
			for i := 0; i < 3; i++ {
				gaps = append(gaps, mintGap(t, runDir, fmt.Sprintf("fz%d", i), "cross-seat-visibility"))
			}
			closed := map[string]bool{}
			for _, l := range lenses {
				sit(t, runDir, l)
			}
			refusals := 0
			for step := 0; step < 40; step++ {
				gap := gaps[rng.Intn(len(gaps))]
				switch r := rng.Intn(10); {
				case r < 2:
					sit(t, runDir, lenses[rng.Intn(len(lenses))])
				case r < 3 && !closed[gap]:
					closeRepaired(t, runDir, gap)
					closed[gap] = true
				case r < 4 && !closed[gap]:
					sev := []string{"low", "medium", "high"}[rng.Intn(3)]
					// One regrade per gap per sitting is the record's own rule; a refused one is either.
					if err := regrade(t, runDir, lensSeat, gap, sev); err != nil && !strings.Contains(err.Error(), "changed after your sitting opened") && !strings.Contains(err.Error(), "already recorded a regrade") {
						t.Fatalf("regrade: %v", err)
					}
				default:
					l := lenses[rng.Intn(len(lenses))]
					key := fmt.Sprintf("k%d", step)
					err := findingAbout(t, runDir, l, key, gap)
					if err == nil {
						continue
					}
					if !strings.Contains(err.Error(), "changed after your sitting opened") {
						t.Fatalf("finding: %v", err)
					}
					refusals++
					if err := findingAbout(t, runDir, l, key, gap); err != nil {
						t.Fatalf("a refused finding filed again with nothing new since was refused — a refusal loop: %v", err)
					}
				}
			}
			if bad := staleFindingsOn(t, runDir); len(bad) > 0 {
				t.Errorf("%d finding(s) recorded against a gap that moved under them:\n%s", len(bad), strings.Join(bad, "\n"))
			}
			t.Logf("seed %d: %d stale refusal(s), each re-filed", seed, refusals)
		})
	}
}

// staleFindingsOn is the oracle: every finding about a gap with another seat's close, regrade or
// finding on that gap after the filer's read point.
func staleFindingsOn(t *testing.T, runDir string) []string {
	t.Helper()
	evs, win, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_CLOSE,
		recordpb.EventType_EVENT_TYPE_REGRADE, recordpb.EventType_EVENT_TYPE_FINDING, recordpb.EventType_EVENT_TYPE_STALE)
	if err != nil {
		t.Fatal(err)
	}
	type act struct {
		id   int64
		seat string
	}
	on := map[string][]act{}
	var bad []string
	findings := 0
	for _, e := range evs {
		id := win.Of(e).ID
		gap := ""
		switch {
		case e.GetClose() != nil:
			gap = e.GetClose().GetGapId()
		case e.GetRegrade() != nil:
			gap = e.GetRegrade().GetGapId()
		case e.GetFinding() != nil && e.GetFinding().GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_GAP:
			gap = e.GetFinding().GetAboutRef()
			findings++
			sitting := win.Of(e).SittingID
			read := sitting
			for _, p := range evs {
				if s := p.GetStale(); s != nil && s.GetGapId() == gap && p.GetSeatId() == e.GetSeatId() && win.Of(p).SittingID == sitting && win.Of(p).ID < id {
					read = max(read, s.GetThrough())
				}
			}
			for _, a := range on[gap] {
				if a.id > read && a.seat != e.GetSeatId() {
					bad = append(bad, fmt.Sprintf("  event %d (%s) about %s, after event %d by %s (read point %d)", id, e.GetSeatId(), gap, a.id, a.seat, read))
				}
			}
		}
		if gap != "" {
			on[gap] = append(on[gap], act{id, e.GetSeatId()})
		}
	}
	if findings == 0 {
		t.Fatal("the oracle read no finding about a gap — it measured nothing")
	}
	return bad
}
