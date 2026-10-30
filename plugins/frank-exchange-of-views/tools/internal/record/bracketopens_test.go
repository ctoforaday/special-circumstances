package record_test

// A HOOK BRACKET OPENS A SITTING AND BINDS ITS AGENT, AND EVERY READER HERE SAYS SO (#1262).
//
// Each test seeds the bracket the way a live run does — the SubagentStart writer
// (sittingwrite.Write), which resolves the configuration to its one seat and appends the
// sitting_open — and then asks a reader that used to count registers. A seat seated by its bracket
// alone is seated, bound and counted, or the reader has a second definition of a sitting.

import (
	"io"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

const (
	lensSeat    = "red-lens-evidence"
	lensConfig  = "frank-exchange-of-views:red-lens-evidence"
	chairConfig = "frank-exchange-of-views:red-chair"
	benchConfig = "frank-exchange-of-views:lead-judge"
)

// bracketedRun is a run as setup leaves it — a cast and an ingested report — with no seat's
// environment leaking in from the shell running the test.
func bracketedRun(t *testing.T) (string, record.Run) {
	t.Helper()
	for _, v := range []string{seatenv.Var, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{SeatIds: []string{lensSeat, "red-chair", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest", &recordpb.ReportPart{Text: proto.String("# report\n\nOne paragraph.\n")}))
	run, err := record.NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, run
}

// bracket is the SubagentStart hook's write for one dispatch of a configuration.
func bracket(t *testing.T, dir, config, agent string) {
	t.Helper()
	if err := sittingwrite.Write(sittingwrite.Sitting{RunDir: dir, Phase: sittingwrite.Open, AgentID: agent, AgentType: config}, io.Discard); err != nil {
		t.Fatalf("the hook's bracket for %s (%s): %v", agent, config, err)
	}
}

// chairDispatch is the chair running `dispatch next`, as the record verb does it.
func chairDispatch(t *testing.T, run record.Run) (record.Plan, bool, error) {
	t.Helper()
	return record.DispatchNext(record.Identity{Run: run, SeatID: "red-chair"})
}

func parties(p record.Plan) []string {
	var out []string
	for _, x := range p.Parties {
		out = append(out, x.SeatID)
	}
	return out
}

// THE CHAIR'S BRACKET OPENS ITS SITTING, and `dispatch next` takes it as opened: the record counts
// it as the chair's sitting (it is the next epoch), so the gate agrees with the record.
func TestAChairSittingItsBracketOpenedMayDispatch(t *testing.T) {
	dir, run := bracketedRun(t)
	bracket(t, dir, chairConfig, "agent_chair_1")
	first, _, err := chairDispatch(t, run)
	if err != nil {
		t.Fatalf("the chair's first sitting: %v", err)
	}
	if !slices.Contains(parties(first), lensSeat) {
		t.Fatalf("the first plan names %v; the test needs the lens dispatched", parties(first))
	}
	bracket(t, dir, lensConfig, "agent_lens_1")   // the lens sits by its bracket alone
	bracket(t, dir, chairConfig, "agent_chair_2") // the workflow comes back: the chair's next sitting
	if _, _, err := chairDispatch(t, run); err != nil {
		t.Fatalf("dispatch next in a chair sitting its bracket opened = %v, want it accepted", err)
	}
}

// SOMEBODY SITTING BY BRACKET SINCE THE DISPATCH IS SOMEBODY SITTING: the same plan asked again is a
// new decision, recorded, not the standing one.
func TestASittingABracketOpenedSinceTheDispatchIsANewDecision(t *testing.T) {
	dir, run := bracketedRun(t)
	bracket(t, dir, chairConfig, "agent_chair_1")
	first, _, err := chairDispatch(t, run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(parties(first), "judge") {
		t.Fatalf("the first plan names the bench (%v); the test needs a seat that is no party", parties(first))
	}
	bracket(t, dir, benchConfig, "agent_bench_1") // a seat no row named sits, by its bracket alone
	again, standing, err := chairDispatch(t, run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(parties(again), parties(first)) {
		t.Fatalf("the plan changed (%v, then %v); the test needs the same plan asked twice", parties(first), parties(again))
	}
	if standing {
		t.Error("the bench sat since the dispatch and the same plan stood — the dispatch after a sitting went unrecorded")
	}
}

// A SEAT BOUND AND SEATED BY ITS BRACKET HAS A SERVED-MODEL ROW, its agent named by the bracket.
func TestABracketSeatedSeatHasAServedModelRow(t *testing.T) {
	dir, run := bracketedRun(t)
	bracket(t, dir, lensConfig, "agent_lens_1")
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	var row *record.SeatModel
	for _, sm := range record.SeatModels(run, fam) {
		if sm.SeatID == lensSeat {
			row = &sm
		}
	}
	if row == nil {
		t.Fatal("a lens seated by its bracket has no served-model row")
	}
	if strings.Contains(row.NotMeasured, "no agent id") {
		t.Errorf("the row reads no agent off the bracket that opened the sitting: %q", row.NotMeasured)
	}
}

// THE SEATS THAT SAT are the stored sittings' owners, a bracket's included.
func TestASeatSeatedByItsBracketHasSat(t *testing.T) {
	dir, run := bracketedRun(t)
	bracket(t, dir, lensConfig, "agent_lens_1")
	seats, err := record.SeatsThatSat(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(seats, lensSeat) {
		t.Errorf("seats that sat = %v, want the lens its bracket seated", seats)
	}
}
