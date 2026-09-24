package capture

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// laneRun declares `lanes` in run-config, writes the cast setup would write, and registers the
// lane seats named.
//
// THE CAST IS PART OF THE FIXTURE NOW, because the lanes are a FIELD on it. A run whose record
// never said which seats are lanes is a run this audit cannot check, and it says SKIP rather than
// guessing from the shape of an id.
func laneRun(t *testing.T, lanes string, registered ...int) string {
	t.Helper()
	run := t.TempDir()
	write(t, filepath.Join(run, "inputs", "run-config.json"), `{"lanes":`+strconv.Quote(lanes)+`}`)
	castN, _ := strconv.Atoi(lanes)
	seats, laneSeats := record.CastFor(nil, castN)
	evs := []*recordpb.Event{recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: seats, LaneSeatIds: laneSeats})}
	for _, n := range registered {
		seat := "blue-lane-" + strconv.Itoa(n)
		evs = append(evs, recordtest.At(t, seat, seat+":register:#1",
			&recordpb.Register{ToolVersion: proto.String("test")}))
	}
	recordtest.Seed(t, run, evs...)
	return run
}

// THE DEFECT, SEEDED: a run configured for three lanes that only ever seated two. Before this
// audit that board was byte-identical to a two-lane run's.
func TestLaneCoverageNamesTheLaneThatNeverRegistered(t *testing.T) {
	got := LaneCoverageAudit(runtest.Open(t, laneRun(t, "3", 1, 2)))
	if got.Verdict != "WARN" {
		t.Fatalf("a lane shortfall must be reported: got %s — %s", got.Verdict, got.Detail)
	}
	for _, want := range []string{"declares 3 lane(s)", "2 registered", "blue-lane-3 never registered", "TWO READINGS"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("detail must carry %q; got:\n%s", want, got.Detail)
		}
	}
	// WHICH lane, not merely how many — that is the half a reader can act on.
	if strings.Contains(got.Detail, "blue-lane-1 never") || strings.Contains(got.Detail, "blue-lane-2 never") {
		t.Errorf("only the missing lane may be named; got:\n%s", got.Detail)
	}
}

// A gap in the middle is still a gap. Counting alone would call 1 and 3 a complete two-lane run.
func TestLaneCoverageCatchesAGapRatherThanOnlyAShortTail(t *testing.T) {
	got := LaneCoverageAudit(runtest.Open(t, laneRun(t, "3", 1, 3)))
	if got.Verdict != "WARN" || !strings.Contains(got.Detail, "blue-lane-2 never registered") {
		t.Fatalf("got %s — %s", got.Verdict, got.Detail)
	}
}

func TestLaneCoveragePassesWhenEveryDeclaredLaneTookItsSeat(t *testing.T) {
	got := LaneCoverageAudit(runtest.Open(t, laneRun(t, "3", 1, 2, 3)))
	if got.Verdict != "PASS" {
		t.Fatalf("got %s — %s", got.Verdict, got.Detail)
	}
	// A re-dispatched lane writes a second register; that is one lane, not two.
	run := laneRun(t, "2", 1, 2)
	recordtest.Seed(t, run, recordtest.At(t, "blue-lane-1", "blue-lane-1:register:#2",
		&recordpb.Register{ToolVersion: proto.String("test")}))
	if got := LaneCoverageAudit(runtest.Open(t, run)); got.Verdict != "PASS" {
		t.Errorf("a re-dispatch is not an extra lane: got %s — %s", got.Verdict, got.Detail)
	}
}

// AN EXCESS IS NOT AMBIGUOUS, so it is not a warning: no dispatch of this config could produce a
// lane the config never asked for.
//
// THE DEFECT IS NOW A DISAGREEMENT BETWEEN TWO RECORDS, and that is the sharper statement of it.
// It used to be arithmetic over names — an index above the declared count. It is the run's CAST
// naming three lanes while its config asked for two: the engine and the config disagree about how
// wide synthesis was, said by two things that both wrote it down.
func TestLaneCoverageFailsOnALaneTheConfigNeverAskedFor(t *testing.T) {
	run := t.TempDir()
	write(t, filepath.Join(run, "inputs", "run-config.json"), `{"lanes":"2"}`)
	seats, laneSeats := record.CastFor(nil, 3)
	evs := []*recordpb.Event{recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: seats, LaneSeatIds: laneSeats})}
	for _, n := range []int{1, 2, 3} {
		seat := "blue-lane-" + strconv.Itoa(n)
		evs = append(evs, recordtest.At(t, seat, seat+":register:#1",
			&recordpb.Register{ToolVersion: proto.String("test")}))
	}
	recordtest.Seed(t, run, evs...)
	got := LaneCoverageAudit(runtest.Open(t, run))
	if got.Verdict != "FAIL" {
		t.Fatalf("a lane beyond the declared count must FAIL: got %s — %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "blue-lane-3 registered beyond the declared count") {
		t.Errorf("detail must name the excess lane; got:\n%s", got.Detail)
	}
}

// THE THREE UNCHECKABLE STATES ARE NOT PASSES. A run that declared nothing, and a record nobody
// could read, must each say so — this audit exists because "looks like two lanes" was allowed to
// stand in for "was two lanes".
func TestLaneCoverageWillNotReportAnUncheckedRunAsAClearOne(t *testing.T) {
	for _, tc := range []struct{ name, cfg, want string }{
		{"no lanes declared", `{}`, "declared no lane count"},
		{"empty string", `{"lanes":""}`, "declared no lane count"},
		{"unparseable config", `{ not json`, "declared no lane count"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := t.TempDir()
			write(t, filepath.Join(run, "inputs", "run-config.json"), tc.cfg)
			recordtest.Seed(t, run)
			got := LaneCoverageAudit(runtest.Open(t, run))
			if got.Verdict != "SKIP" || !strings.Contains(got.Detail, tc.want) {
				t.Errorf("got %s — %s", got.Verdict, got.Detail)
			}
			if !strings.Contains(got.Detail, "NOT a run whose lanes were checked") {
				t.Errorf("the miss must not read as a clean board; got:\n%s", got.Detail)
			}
		})
	}

	// A declared count whose record cannot be read: the count is known, the outcome is not.
	broken := t.TempDir()
	write(t, filepath.Join(broken, "inputs", "run-config.json"), `{"lanes":"3"}`)
	if err := os.MkdirAll(filepath.Join(broken, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "records", "record.db"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := LaneCoverageAudit(runtest.Open(t, broken))
	if got.Verdict != "SKIP" || !strings.Contains(got.Detail, "NOT a run whose lanes all registered") {
		t.Errorf("an unreadable record must not read as full coverage; got %s — %s", got.Verdict, got.Detail)
	}
}
