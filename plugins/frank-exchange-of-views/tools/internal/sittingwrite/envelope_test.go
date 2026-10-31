package sittingwrite

import (
	"bytes"
	"encoding/json"
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
		at(record.HarnessSeat, &recordpb.ReportPart{Text: proto.String("# report")}),
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
	if err := WriteEnvelopeAnswer(runDir, "agent-bench", []byte(`{}`), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "motion M1 ") || !strings.Contains(out.String(), "motion M2 ") {
		t.Fatalf("the writer did not print the refusal naming both twins: %q", out.String())
	}
	var stray bytes.Buffer
	if err := WriteEnvelopeAnswer(runDir, "agent-of-no-seat", []byte(`{}`), &stray); err != nil || stray.Len() != 0 {
		t.Errorf("an agent bound to no seat had its envelope refused (%v): %q", err, stray.String())
	}
	var none bytes.Buffer
	if err := WriteEnvelopeAnswer(newRun(t), "agent-bench", []byte(`{}`), &none); err == nil || none.Len() != 0 {
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

// THE CHAIR'S CALL COMES BACK CARRYING THE PLAN THE RECORD STORED, with nothing logged; one that states
// a field of the plan comes back refused, and the refusal is logged under the seat.
func TestTheWriterAnswersTheChairsEnvelopeWithTheStoredPlan(t *testing.T) {
	run := admissionRun(t)
	write(t, Sitting{RunDir: run, Phase: Open, AgentID: "AG", AgentType: chair})
	plan, _, err := record.DispatchNext(record.Identity{Run: mustRun(t, run), SeatID: "red-chair"})
	if err != nil {
		t.Fatal(err)
	}
	logged := func() int {
		evs, _, err := record.EventsOf(mustRun(t, run), recordpb.EventType_EVENT_TYPE_LOG)
		if err != nil {
			t.Fatal(err)
		}
		return len(evs)
	}
	before := logged()
	var out bytes.Buffer
	if err := WriteEnvelopeAnswer(run, "AG", []byte(`{"notes":"n"}`), &out); err != nil {
		t.Fatal(err)
	}
	var answer EnvelopeAnswer
	if err := json.Unmarshal(out.Bytes(), &answer); err != nil {
		t.Fatalf("the writer printed no answer: %v: %q", err, out.String())
	}
	var fields map[string]any
	if err := json.Unmarshal(answer.Envelope, &fields); err != nil {
		t.Fatalf("the answer carries no envelope: %v: %q", err, out.String())
	}
	if fields["notes"] != "n" {
		t.Errorf("the chair's notes did not stay beside the plan: %v", fields["notes"])
	}
	delete(fields, "notes")
	bare, _ := json.Marshal(fields)
	if _, printed, err := record.ReadPlan(bare); err != nil || printed != plan.Printed() || answer.Refusal != "" {
		t.Errorf("the answer is not the stored plan (%v), refusal %q\n got %s\nwant %s", err, answer.Refusal, printed, plan.Printed())
	}
	if n := logged(); n != before {
		t.Errorf("an envelope the record supplied logged %d entries", n-before)
	}
	out.Reset()
	if err := WriteEnvelopeAnswer(run, "AG", []byte(`{"blockers":[]}`), &out); err != nil {
		t.Fatal(err)
	}
	answer = EnvelopeAnswer{}
	if err := json.Unmarshal(out.Bytes(), &answer); err != nil {
		t.Fatalf("the writer printed no answer: %v: %q", err, out.String())
	}
	if !strings.Contains(answer.Refusal, "`blockers`") || answer.Envelope != nil {
		t.Errorf("a stated field of the plan was not refused naming it: %+v", answer)
	}
	if n := logged(); n != before+1 {
		t.Errorf("the refusal logged %d entries, want one", n-before)
	}
}
