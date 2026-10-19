package record_test

import (
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

func familyOf(t *testing.T, evs ...*recordpb.Event) record.Family {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	f, err := record.FamilyOf(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func stop(t *testing.T, agent, path string) *recordpb.Event {
	t.Helper()
	cl := &recordpb.SittingClose{AgentId: proto.String(agent)}
	if path != "" {
		cl.AgentTranscriptPath = proto.String(path)
	}
	return recordtest.Event(t, record.HarnessSeat, cl)
}

func opened(t *testing.T, seat, agent string) *recordpb.Event {
	t.Helper()
	return recordtest.Event(t, seat, &recordpb.Register{AgentId: proto.String(agent)})
}

// THE RECORD SAYS WHERE EACH SEAT'S TRANSCRIPT IS, and every agent it cannot name a readable one
// for is sorted by why: still sitting, stopped with no path, or named and gone from this box. The
// latest stop's path is the one named.
func TestTheRecordNamesEachSeatsTranscriptAndEveryMissSaysWhy(t *testing.T) {
	dir := t.TempDir()
	here := filepath.Join(dir, "lens.jsonl")
	if err := os.WriteFile(here, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gone := filepath.Join(dir, "gone.jsonl")
	f := familyOf(t,
		opened(t, "red-lens-evidence", "L1"),
		stop(t, "L1", filepath.Join(dir, "an-earlier-stop.jsonl")),
		stop(t, "L1", here),
		opened(t, "red-chair", "C1"),
		stop(t, "C1", gone),
		opened(t, "frontier", "F1"),
		stop(t, "F1", ""),
		opened(t, "judge", "J1"), // the bench, still in its sitting
	)

	c := record.TranscriptsOf(f)
	if len(c.Named) != 2 || c.Named[0] != (record.Transcript{AgentID: "C1", Path: gone}) || c.Named[1] != (record.Transcript{AgentID: "L1", Path: here}) {
		t.Errorf("Named = %+v, want C1 at its path and L1 at its LATEST stop's", c.Named)
	}
	if len(c.Unnamed) != 1 || c.Unnamed[0] != "F1" {
		t.Errorf("Unnamed = %v, want [F1]", c.Unnamed)
	}
	if len(c.Sitting) != 1 || c.Sitting[0] != "J1" {
		t.Errorf("Sitting = %v, want [J1]", c.Sitting)
	}

	if p, err := record.TranscriptOf(f, "L1"); err != nil || p != here {
		t.Errorf("TranscriptOf(L1) = %q, %v; want %q", p, err, here)
	}
	for agent, want := range map[string]string{
		"J1": "has not stopped",
		"F1": "names no transcript path",
		"C1": "not archived with the record",
	} {
		p, err := record.TranscriptOf(f, agent)
		if err == nil || p != "" {
			t.Errorf("TranscriptOf(%s) = %q, %v; a miss must be an error and no path", agent, p, err)
			continue
		}
		if !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), agent) {
			t.Errorf("TranscriptOf(%s) must name the agent and say %q: %v", agent, want, err)
		}
	}
}

// A SEAT STILL SITTING HAS NO STOP, SO NOTHING NAMES ITS TRAJECTORY — even with the file on disk.
// The bench assembles run.md inside its own sitting (#1437/#1438); its row says why it is NOT
// MEASURED rather than being answered by a search, and capture re-assembles after the stop.
func TestASeatStillSittingIsNotMeasuredAndSaysSo(t *testing.T) {
	f := familyOf(t,
		opened(t, "red-chair", recordtest.ServedBy(t, "C1", "claude-sonnet-5", "")),
		recordtest.Event(t, record.HarnessSeat, recordtest.StopOf(t, "C1")),
		opened(t, "judge", recordtest.ServedBy(t, "J1", "claude-sonnet-5", "")),
	)
	rows := map[string]record.SeatModel{}
	for _, sm := range record.SeatModels(record.Run{}, f) {
		rows[sm.SeatID] = sm
	}
	if c := rows["red-chair"]; c.Served != "claude-sonnet-5" {
		t.Errorf("the stopped seat is measured from the path its stop names: %+v", c)
	}
	j := rows["judge"]
	if j.Measured() || !strings.Contains(j.NotMeasured, "has not stopped") || !strings.Contains(j.NotMeasured, "J1") {
		t.Errorf("the sitting bench must read NOT MEASURED, naming its agent and why: %+v", j)
	}
}
