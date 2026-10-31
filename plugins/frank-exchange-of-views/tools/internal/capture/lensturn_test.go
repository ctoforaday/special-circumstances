package capture

import (
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
