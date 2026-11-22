package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE LENS-TURN AUDIT FAILS A DISPATCH THAT ENGAGED A LENS BEFORE BLUE (ruled by gblock 2026-10-30),
// names it, and passes a run whose lens sat on the gap only after blue's answer. Capture runs it.
func TestTheLensTurnAuditNamesAnEngagementBeforeBlue(t *testing.T) {
	const lens = "red-lens-evidence"
	seeded := func(t *testing.T, blueFirst bool) record.Run {
		dir := t.TempDir()
		evs := []*recordpb.Event{
			recordtest.At(t, lens, "mint-G1", &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String("G1"),
				Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
				CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
				Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
		}
		if blueFirst {
			evs = append(evs, recordtest.At(t, "blue-respond", "edit-G1", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("b")}))
		}
		evs = append(evs, recordtest.At(t, "red-chair", "dispatch-lens", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(lens), GapIds: []string{"G1"}}))
		recordtest.Seed(t, dir, evs...)
		run, err := record.NewRun(dir)
		if err != nil {
			t.Fatal(err)
		}
		return run
	}
	if a := LensTurnAudit(seeded(t, false)); a.Verdict != "FAIL" || !strings.Contains(a.Detail, lens+" on G1") {
		t.Errorf("a lens engaged on G1 before blue acted on it: %s %s, want FAIL naming it", a.Verdict, a.Detail)
	}
	if a := LensTurnAudit(seeded(t, true)); a.Verdict != "PASS" {
		t.Errorf("a lens engaged on G1 after blue's answer: %s %s, want PASS", a.Verdict, a.Detail)
	}

	run := runtest.New(t, t.TempDir())
	audits, _, _, err := Run(run, t.TempDir(), t.TempDir(), time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range audits {
		found = found || a.Check == "lens-turn"
	}
	if !found {
		t.Error("capture runs no lens-turn audit")
	}
}

// THE DOCKET-ORDER AUDIT FAILS AN IMPASSE DOCKETING WRITTEN WITH THE LENS'S AUDIT STILL TO COME (ruled
// by gblock 2026-10-31), names it, and passes the same record with the lens's sitting between blue's
// answer and the docketing. Capture runs it.
func TestTheDocketOrderAuditNamesADocketBeforeTheLensHasRead(t *testing.T) {
	const lens = "red-lens-evidence"
	d := record.NewID("docketing")
	seeded := func(t *testing.T, lensSat bool) record.Run {
		evs := docketOrderBoard(t, lens, lensSat)
		evs = append(evs, recordtest.At(t, "red-chair", "docket-G1", docketingOf(d, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")))
		return seedKMax1(t, evs)
	}
	if a := DocketOrderAudit(seeded(t, false)); a.Verdict != "FAIL" || !strings.Contains(a.Detail, d+" on G1") {
		t.Errorf("an impasse docketing written with the lens's audit of blue's answer still to come: %s %s, want FAIL naming it", a.Verdict, a.Detail)
	}
	if a := DocketOrderAudit(seeded(t, true)); a.Verdict != "PASS" {
		t.Errorf("an impasse docketing written after the lens's sitting: %s %s, want PASS", a.Verdict, a.Detail)
	}

	run := runtest.New(t, t.TempDir())
	audits, _, _, err := Run(run, t.TempDir(), t.TempDir(), time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range audits {
		found = found || a.Check == "docket-order"
	}
	if !found {
		t.Error("capture runs no docket-order audit")
	}
}

// docketOrderBoard is G1 minted by lens, blue engaged on it and answering it in a sitting a later
// opening closes, and — where lensSat — the lens engaged on it and sitting through its re-audit.
func docketOrderBoard(t *testing.T, lens string, lensSat bool) []*recordpb.Event {
	t.Helper()
	evs := []*recordpb.Event{
		recordtest.At(t, record.HarnessSeat, "cast", &recordpb.Cast{SeatIds: []string{lens, "red-chair", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "report", &recordpb.ReportPart{Text: proto.String("# report")}),
		recordtest.At(t, lens, "register-lens", &recordpb.Register{}),
		recordtest.At(t, lens, "mint-G1", &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String("G1"),
			Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
		recordtest.At(t, "red-chair", "dispatch-blue", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.At(t, "blue-respond", "register-blue", &recordpb.Register{}),
		recordtest.At(t, "blue-respond", "edit-G1", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("b")}),
		// A later opening closes blue's sitting: an open one leaves its exchange unresolved.
		recordtest.At(t, "blue-respond", "register-blue-again", &recordpb.Register{}),
	}
	if lensSat {
		evs = append(evs,
			recordtest.At(t, "red-chair", "dispatch-lens", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(lens), GapIds: []string{"G1"}}),
			recordtest.At(t, lens, "register-lens-audit", &recordpb.Register{}),
			recordtest.At(t, lens, "register-lens-again", &recordpb.Register{}))
	}
	return evs
}

// docketingOf is the dispatch's docketing of G1 with id, for cause, holding motion where cause is
// PROPOSAL.
func docketingOf(id string, cause recordpb.DocketingCause, motion string) *recordpb.Docketing {
	d := &recordpb.Docketing{DocketingId: proto.String(id), GapId: proto.String("G1"), Cause: cause.Enum(), Basis: proto.String(recordpb.Word(cause))}
	if motion != "" {
		d.MotionId = proto.String(motion)
	}
	return d
}

// seedKMax1 seeds evs on a fresh run whose terms are k-max 1.
func seedKMax1(t *testing.T, evs []*recordpb.Event) record.Run {
	t.Helper()
	dir := t.TempDir()
	recordtest.Seed(t, dir, evs...)
	if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "inputs", "run-config.json"), []byte(`{"kMax": 1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	run, err := record.NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	return run
}
