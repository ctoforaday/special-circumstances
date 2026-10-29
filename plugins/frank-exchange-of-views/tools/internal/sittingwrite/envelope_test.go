package sittingwrite

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE WRITER PRINTS THE RECORD'S REFUSAL OF THE AGENT'S ENVELOPE AND LOGS IT under the agent's seat:
// a bench docket sitting with twin docket motions standing on its gap. An agent bound to no seat gets
// nothing printed and nothing logged, and a run with no record is an error the hook refuses on.
func TestTheWriterPrintsAndLogsAnEnvelopeRefusal(t *testing.T) {
	runDir := newRun(t)
	n := 0
	at := func(seat string, body proto.Message) *recordpb.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	sev := recordpb.Grade_GRADE_LOW
	docket := func(id string) *recordpb.Motion {
		return &recordpb.Motion{MotionId: proto.String(id), Basis: proto.String("b"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}
	}
	recordtest.Seed(t, runDir,
		at(record.HarnessSeat, &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		at(record.HarnessSeat, &recordpb.BaseIngest{Text: proto.String("# report")}),
		at("red-chair", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Register{AgentId: proto.String("agent-lens")}),
		at("red-lens-evidence", &recordpb.Mint{GapId: proto.String("G1"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
		at("blue-respond", docket("M1")),
		at("blue-respond", docket("M2")),
		at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G1"},
			Occasions: []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}}),
		at("judge", &recordpb.Register{AgentId: proto.String("agent-bench"), Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum()}),
	)
	var out bytes.Buffer
	if err := WriteEnvelopeRefusal(runDir, "agent-bench", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "motion M1 ") || !strings.Contains(out.String(), "motion M2 ") {
		t.Fatalf("the writer did not print the refusal naming both twins: %q", out.String())
	}
	var stray bytes.Buffer
	if err := WriteEnvelopeRefusal(runDir, "agent-of-no-seat", &stray); err != nil || stray.Len() != 0 {
		t.Errorf("an agent bound to no seat had its envelope refused (%v): %q", err, stray.String())
	}
	var none bytes.Buffer
	if err := WriteEnvelopeRefusal(newRun(t), "agent-bench", &none); err == nil || none.Len() != 0 {
		t.Errorf("a run with no record was checked as one with nothing to refuse (%v): %q", err, none.String())
	}
	evs, _, err := record.EventsOf(mustRun(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 || evs[0].GetSeatId() != "judge" {
		t.Fatalf("want the bench's one refusal entry, got %d entr(ies)", len(evs))
	}
	if l, _ := recordpb.BodyAs[*recordpb.Log](evs[0]); l.GetType() != recordpb.LogType_LOG_TYPE_REFUSAL || !strings.Contains(l.GetText(), "motion M2 ") {
		t.Errorf("the entry is not the tool's refusal naming the twins: %v", l)
	}
	if got := admissions(t, runDir); len(got) != 0 {
		t.Errorf("a refused envelope was recorded as admitted: %v", got)
	}
}
