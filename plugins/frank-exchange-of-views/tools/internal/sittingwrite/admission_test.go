package sittingwrite

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

const chair = "frank-exchange-of-views:red-chair"

// admissionRun is a run whose record holds a cast and the report's base, so a seat's work list reads.
func admissionRun(t *testing.T) string {
	t.Helper()
	run := newRun(t)
	recordtest.Seed(t, run,
		recordtest.At(t, record.HarnessSeat, "cast", &recordpb.Cast{SeatIds: []string{"red-chair", "frontier", "blue-respond"}}),
		recordtest.At(t, record.HarnessSeat, "base", &recordpb.ReportPart{Text: proto.String("# report")}))
	return run
}

// admissions is the agents of every envelope_admitted on the record, in order, each checked to be
// the harness's.
func admissions(t *testing.T, run string) []string {
	t.Helper()
	m, err := record.MergedEvents(mustRun(t, run))
	if err != nil {
		t.Fatal(err)
	}
	var agents []string
	for _, e := range m.Events {
		if a, ok := recordpb.BodyAs[*recordpb.EnvelopeAdmitted](e); ok {
			if e.GetSeatId() != HookSeat {
				t.Errorf("an admission is attributed to %q, want %q — the hook observed it, the seat did not claim it", e.GetSeatId(), HookSeat)
			}
			agents = append(agents, a.GetAgentId())
		}
	}
	return agents
}

func write(t *testing.T, s Sitting) {
	t.Helper()
	if err := Write(s, nil); err != nil {
		t.Fatal(err)
	}
}

// AN ENVELOPE WHOSE CALL SUCCEEDED IS RECORDED AS ADMITTED, under the harness and keyed to the agent;
// the same sitting admitted again writes nothing, and its close is then no failure.
func TestAnAdmittedEnvelopeIsRecorded(t *testing.T) {
	run := admissionRun(t)
	write(t, Sitting{RunDir: run, Phase: Open, AgentID: "AG", AgentType: chair})
	for range 2 {
		if err := WriteAdmission(run, "AG"); err != nil {
			t.Fatalf("the chair's accepted envelope was not admitted: %v", err)
		}
	}
	if got := admissions(t, run); len(got) != 1 || got[0] != "AG" {
		t.Fatalf("admissions = %v, want one, of AG", got)
	}
	write(t, Sitting{RunDir: run, Phase: Close, AgentID: "AG", AgentType: chair})
	if f, err := record.FailedSittings(mustRun(t, run)); err != nil || len(f) != 0 {
		t.Errorf("a sitting that closed on an admitted envelope reads failed (%v): %+v", err, f)
	}
}

// A SITTING THAT CLOSED WITH NO ADMITTED ENVELOPE FAILED — for a seat whose work arrives as items.
// A nudged seat closed at its first stop and admitted after it finished; a sitting still open has not
// failed; and the frontier, which returns no envelope the gate asks about, never reads failed.
func TestAClosedSittingWithNoAdmissionReadsFailed(t *testing.T) {
	run := admissionRun(t)
	// The chair's first sitting: closed, never admitted.
	write(t, Sitting{RunDir: run, Phase: Open, AgentID: "C1", AgentType: chair})
	write(t, Sitting{RunDir: run, Phase: Close, AgentID: "C1", AgentType: chair})
	// Its second: nudged — closed, then admitted.
	write(t, Sitting{RunDir: run, Phase: Open, AgentID: "C2", AgentType: chair})
	write(t, Sitting{RunDir: run, Phase: Close, AgentID: "C2", AgentType: chair})
	if err := WriteAdmission(run, "C2"); err != nil {
		t.Fatal(err)
	}
	// Its third: still open.
	write(t, Sitting{RunDir: run, Phase: Open, AgentID: "C3", AgentType: chair})
	// The frontier: registered, closed, no envelope.
	registerAs(t, run, "F1", "frontier")
	write(t, Sitting{RunDir: run, Phase: Close, AgentID: "F1", AgentType: "frank-exchange-of-views:blue-researcher"})

	f, err := record.FailedSittings(mustRun(t, run))
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 1 || f[0].SeatID != "red-chair" || f[0].Sitting != 1 || f[0].AgentID != "C1" {
		t.Fatalf("failed sittings = %+v, want the chair's first sitting alone (agent C1)", f)
	}
}

// THE GATE ADMITS NOTHING. An envelope the record refuses nothing on is let through by the PreToolUse
// gate and still faces the harness's schema check, so the gate's answer writes no admission: the
// sitting is admitted only when the call succeeds (WriteAdmission, from PostToolUse).
func TestTheGateAdmitsNothing(t *testing.T) {
	run := admissionRun(t)
	write(t, Sitting{RunDir: run, Phase: Open, AgentID: "AG", AgentType: chair})
	var out bytes.Buffer
	if err := WriteEnvelopeRefusal(run, "AG", &out); err != nil || out.Len() != 0 {
		t.Fatalf("the chair's envelope was refused (%v): %q", err, out.String())
	}
	if got := admissions(t, run); len(got) != 0 {
		t.Fatalf("the gate's answer wrote admissions %v — an envelope the schema may still reject is not admitted", got)
	}
	write(t, Sitting{RunDir: run, Phase: Close, AgentID: "AG", AgentType: chair})
	if f, err := record.FailedSittings(mustRun(t, run)); err != nil || len(f) != 1 {
		t.Errorf("a sitting whose call never succeeded reads as finished (%v): %+v", err, f)
	}
}

// AN AGENT BOUND TO NO SEAT IS NO SEAT'S ENVELOPE, and its successful call admits nothing.
func TestAnAgentOfNoSeatIsNotAdmitted(t *testing.T) {
	run := admissionRun(t)
	if err := WriteAdmission(run, "agent-of-no-seat"); err != nil {
		t.Fatal(err)
	}
	if got := admissions(t, run); len(got) != 0 {
		t.Errorf("an agent bound to no seat had its envelope admitted: %v", got)
	}
}
