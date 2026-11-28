package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// boardGap is the gap impasseBoard mints, by its id.
const boardGap = "G-000000b1"

// otherLens is a second lens of the cast: a red seat that is not G1's assignee.
const otherLens = "red-lens-logic"

// writeTerms writes the run's terms where setup records them: k-max, and the epoch limit where limit
// is above zero. One write, because the file is one.
func writeTerms(t *testing.T, runDir string, kMax, limit int) {
	t.Helper()
	terms := map[string]int{"kMax": kMax}
	if limit > 0 {
		terms["maxEpochs"] = limit
	}
	b, err := json.Marshal(terms)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "inputs", "run-config.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// impasseBoard seeds boardGap (G1 below) — minted by the evidence lens, about a section — on a cast of two lenses, with
// exchanges null turns on it: each chair sitting engages the lens and blue on G1, and each sits and
// records nothing; a closing chair register follows. A sitting is closed by that seat's next opening,
// so the last exchange is countable only once the cycle after it opens.
func impasseBoard(t *testing.T, runDir string, kMax, exchanges int) {
	t.Helper()
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:board:%d", seat, n), body)
	}
	evs := []*record.Event{
		at(record.HarnessSeat, &recordpb.Cast{SeatIds: []string{"red-lens-evidence", otherLens, "red-chair", "blue-respond", "judge"}}),
		at(record.HarnessSeat, &recordpb.ReportPart{Text: proto.String("# report\n\n## Findings\n\nThe findings stand here.\n")}),
		at("red-chair", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(boardGap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
		// the section gap's token, at the end of its heading (§III.B0 "Sections")
		at("red-lens-evidence", &recordpb.Anchor{Id: proto.String(boardGap), Location: proto.String("## Findings")}),
	}
	for i := 0; i < exchanges+1; i++ {
		evs = append(evs,
			at("red-chair", &recordpb.Register{}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-evidence"), GapIds: []string{boardGap}}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{boardGap}}),
			at("red-lens-evidence", &recordpb.Register{}),
			at("blue-respond", &recordpb.Register{}),
		)
	}
	evs = append(evs, at("red-chair", &recordpb.Register{}))
	recordtest.Seed(t, runDir, evs...)
	writeTerms(t, runDir, kMax, 0)
}

// filersOfMotions is the seat of every motion row on the record, by motion id.
func filersOfMotions(t *testing.T, runDir string) map[string]string {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_MOTION)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, e := range evs {
		m, _ := recordpb.BodyAs[*recordpb.Motion](e)
		out[m.GetMotionId()] = e.GetSeatId()
	}
	return out
}

// T-B24: THE DISPATCH FILES NO MOTION (§III.B13). After a plan on each route — impasse, a proposal in its
// window, the window's end — every motion row is a filer's own, written by its verb; a docket motion
// with no act is refused from any writer, the chair's raw append included; a standing proposal defers
// the impasse docketing (L28), validate refuses an impasse docketing appended raw while it stands, and
// the window's end writes Docketing{PROPOSAL, motion: M}.
func TestTheDispatchFilesNoMotion(t *testing.T) {
	t.Run("impasse", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 2, 2)
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		g := boardGap
		if len(plan.Docket) != 1 || plan.Docket[0] != g {
			t.Fatalf("plan = %+v, want G1 docketed at impasse", plan)
		}
		if n := motionsOnRecord(t, runDir); n != 0 {
			t.Errorf("the plan at impasse left %d motion(s) on the record, want none", n)
		}
		if d := lastDocketing(t, runDir, boardGap); d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE || d.MotionId != nil {
			t.Errorf("the docketing = %+v, want impasse holding no proposal", d)
		}
	})

	t.Run("a docket motion with no act is refused from every writer", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 3, 1)
		g := boardGap
		noAct := &recordpb.Motion{Basis: proto.String("at impasse"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(g)}}}
		for _, seat := range []string{"red-chair", "blue-respond", otherLens} {
			if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: seat}, proto.Clone(noAct)); err == nil {
				t.Errorf("%s appended a docket motion with no act, and the record admitted it", seat)
			}
		}
		if n := motionsOnRecord(t, runDir); n != 0 {
			t.Errorf("%d motion(s) on the record after every refused append, want none", n)
		}
	})

	// secondExchange is G1's second exchange after one: a third cycle of null turns, whose openings
	// close the second's sittings. At k-max 2 the board is then at impasse, with blue's proposal of a
	// close filed in its second sitting where propose is set.
	secondExchange := func(t *testing.T, propose bool) (runDir, m string) {
		runDir = newRun(t)
		impasseBoard(t, runDir, 2, 1)
		if propose {
			m = proposeClose(t, runDir, "blue-respond", boardGap, "not_a_defect", "not a defect, and the lens will not close it")
		}
		asSeat(t, runDir, "red-chair", "register")
		for _, seat := range []string{"red-lens-evidence", "blue-respond"} {
			if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"},
				&recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(seat), GapIds: []string{boardGap}}); err != nil {
				t.Fatal(err)
			}
		}
		asSeat(t, runDir, "red-lens-evidence", "register")
		asSeat(t, runDir, "blue-respond", "register")
		asSeat(t, runDir, "red-chair", "register")
		return runDir, m
	}
	t.Run("the control: with no proposal the same board is docketed at impasse", func(t *testing.T) {
		runDir, _ := secondExchange(t, false)
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if ds, _ := docketingsOn(t, runDir, boardGap); len(ds) != 1 || ds[0].GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
			t.Fatalf("plan %+v: docketings = %+v, want the one impasse docketing — the proposal row below measures nothing without it", plan, ds)
		}
	})

	t.Run("a proposal outranks the impasse, and its window's end dockets it", func(t *testing.T) {
		runDir, m := secondExchange(t, true)
		g := boardGap
		// L28 AT THE WRITE: an impasse or grade-movement docketing appended raw while the proposal stands
		// is refused, naming it.
		for _, cause := range []recordpb.DocketingCause{recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT} {
			raw := &recordpb.Docketing{DocketingId: proto.String(record.NewID("docketing")), GapId: proto.String(g),
				Cause: cause.Enum(), Basis: proto.String(recordpb.Word(cause))}
			if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"}, raw); err == nil {
				t.Errorf("a %s docketing appended while proposal %s stands was admitted (L28)", recordpb.Word(cause), m)
			} else if !strings.Contains(err.Error(), m) {
				t.Errorf("the refusal of a %s docketing over proposal %s does not name it: %v", recordpb.Word(cause), m, err)
			}
		}
		// The board is at impasse (the control above) and the assignee has sat since the filing: the
		// plan writes the proposal's docketing, and no impasse docketing.
		asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
		ds, _ := docketingsOn(t, runDir, g)
		if len(ds) != 1 || ds[0].GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL || ds[0].GetMotionId() != m {
			t.Fatalf("docketings = %+v, want one PROPOSAL docketing holding %s and no impasse docketing (L28)", ds, m)
		}
		for id, seat := range filersOfMotions(t, runDir) {
			if id != m || seat != "blue-respond" {
				t.Errorf("motion %s on the record under %s, want only %s, blue's own", id, seat, m)
			}
		}
	})

	t.Run("a proposal in its window is docketed for nothing, and its window's end dockets it", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 2, 1)
		m := proposeClose(t, runDir, "blue-respond", boardGap, "not_a_defect", "not a defect, and the lens will not close it")
		asSeat(t, runDir, "red-chair", "register")
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if ds, _ := docketingsOn(t, runDir, boardGap); len(ds) != 0 {
			t.Fatalf("with proposal %s in its window the plan docketed the gap: %+v", m, ds)
		}
		readied := false
		for _, p := range plan.Parties {
			readied = readied || p.SeatID == "red-lens-evidence"
		}
		if !readied {
			t.Errorf("plan = %+v, want the assignee readied for its window", plan)
		}
		if d := docketsUnaccepted(t, runDir, "red-lens-evidence", boardGap); d.GetMotionId() != m {
			t.Errorf("the window's end wrote %+v, want PROPOSAL holding %s", d, m)
		}
		if n := motionsOnRecord(t, runDir); n != 1 {
			t.Errorf("%d motion(s) on the record, want blue's one", n)
		}
	})
}
