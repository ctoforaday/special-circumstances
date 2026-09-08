package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// stageEvents seeds a board the dispatch verb reads: the cast, the ingested report, the chair and
// a lens registered, and one material gap the lens minted.
func stageEvents(t *testing.T, runDir string, more ...*record.Event) {
	t.Helper()
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:stage:%d", seat, n), body)
	}
	evs := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# report")}),
		at("red-chair", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Mint{GapId: proto.String("G1"), Class: proto.String("x"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
	}
	recordtest.Seed(t, runDir, append(evs, more...)...)
}

// The chair's verb records who sits and prints the same plan; the workflow obeys the record.
func TestDispatchNextRecordsThePartiesItNames(t *testing.T) {
	runDir := newRun(t)
	stageEvents(t, runDir)
	out, err := run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, out)
	}
	plan := planOf(t, out)
	if plan.Head != 2 || len(plan.Parties) != 2 || plan.PassPermitted || plan.Ceiling {
		t.Fatalf("plan = %+v, want head 2 and two parties (the lens and blue on G1)", plan)
	}
	evs, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_DISPATCH)
	if err != nil || len(evs) != 2 {
		t.Fatalf("dispatch events on the record = %d (%v), want one per party", len(evs), err)
	}
	for _, e := range evs {
		d, _ := recordpb.BodyAs[*recordpb.Dispatch](e)
		if e.GetSeatId() != "red-chair" || d.GetPin() != 2 {
			t.Errorf("dispatch %+v: want written under red-chair, pinned to head 2", d)
		}
	}
	if _, err := run(t, "dispatch", "later", "--run", runDir, "--seat-id", "red-chair"); err == nil {
		t.Error("dispatch accepted a word other than `next`")
	}
}

// At impasse the verb dockets the gap for the bench and names only the bench.
func TestDispatchNextDocketsAGapAtImpasse(t *testing.T) {
	runDir := newRun(t)
	n := 100
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:impasse:%d", seat, n), body)
	}
	var cycles []*record.Event
	for i := 0; i < 2; i++ { // two stalled exchanges under K = 2
		cycles = append(cycles,
			at("red-chair", &recordpb.Register{}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-evidence"), GapIds: []string{"G1"}}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
			at("red-lens-evidence", &recordpb.Register{}),
			at("blue-respond", &recordpb.Register{}),
		)
	}
	cycles = append(cycles, at("red-chair", &recordpb.Register{}))
	stageEvents(t, runDir, cycles...)
	out, err := run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, out)
	}
	plan := planOf(t, out)
	if len(plan.Docket) != 1 || plan.Docket[0] != "G1" || len(plan.Parties) != 1 || plan.Parties[0].SeatID != "judge" {
		t.Fatalf("plan = %+v, want G1 docketed and the bench alone engaged", plan)
	}
	motions, _ := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_MOTION)
	if len(motions) != 1 {
		t.Fatalf("motions on the record = %d, want the docket motion the verb filed", len(motions))
	}
	m, _ := recordpb.BodyAs[*recordpb.Motion](motions[0])
	if m.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DOCKET || m.GetDocket().GetGapId() != "G1" || !strings.Contains(m.GetBasis(), "impasse") {
		t.Errorf("the docket motion = %+v, want a docket filing on G1 stating impasse", m)
	}
	// Asked again with nothing changed: the docket stands, the bench is still the only party, and
	// no second motion is filed.
	out2, err := run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatal(err)
	}
	plan2 := planOf(t, out2)
	motions, _ = record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_MOTION)
	if len(plan2.Docket) != 0 || len(motions) != 1 || len(plan2.Parties) != 1 {
		t.Errorf("a second ask re-docketed or re-engaged: plan=%+v motions=%d", plan2, len(motions))
	}
}

// planOf reads the verb's plan out of the seat envelope every JSON verb answers in.
func planOf(t *testing.T, out string) record.Plan {
	t.Helper()
	var env struct {
		OK     bool        `json:"ok"`
		Result record.Plan `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || !env.OK {
		t.Fatalf("the plan is not an ok envelope: %v\n%s", err, out)
	}
	return env.Result
}
