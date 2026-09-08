package dashboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
)

func registered(t *testing.T, seat, agent string) *recordpb.Event {
	t.Helper()
	return recordtest.Event(t, seat, &recordpb.Register{AgentId: proto.String(agent)})
}

// A SEAT'S IDENTITY IS THE RECORD'S, NOT THE TRANSCRIPT HEAD'S. The journal names an agent; the
// register event that agent wrote names the seat, and the record counts which sitting of that
// seat it was and which epoch it sat in. The label is `seat #sitting`. An agent the record never
// bound keeps the bare class from its prompt head, with no number to invent.
func TestSeatsAreLabelledFromTheRecordsRegisters(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir,
		registered(t, "red-lens-evidence", "L1"), // before any chair: epoch 0, sitting 1
		registered(t, "red-chair", "C1"),         // opens epoch 1
		registered(t, "red-chair", "C2"),         // opens epoch 2
		registered(t, "red-lens-evidence", "L2"), // epoch 2, sitting 2
	)
	tr := t.TempDir()
	journal := ""
	for _, id := range []string{"L1", "C1", "C2", "L2", "UNBOUND"} {
		journal += `{"agentId":"` + id + `"}` + "\n"
		head := `{"message":{"role":"user","content":"Red audit — a lens"}}` + "\n"
		if id == "UNBOUND" {
			head = `{"message":{"role":"user","content":"Final assembly"}}` + "\n"
		}
		if err := os.WriteFile(filepath.Join(tr, "agent-"+id+".jsonl"), []byte(head+
			`{"message":{"model":"claude-haiku-4-5","usage":{"input_tokens":1000,"output_tokens":500}}}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tr, "journal.jsonl"), []byte(journal), 0o644); err != nil {
		t.Fatal(err)
	}

	m := BuildModel(runtest.Open(t, runDir), tr, Config{}, 1737000000000)

	want := map[string]struct {
		label          string
		epoch, sitting int
	}{
		"L1":      {"red-lens-evidence #1", 0, 1},
		"C1":      {"red-chair #1", 1, 1},
		"C2":      {"red-chair #2", 2, 2},
		"L2":      {"red-lens-evidence #2", 2, 2},
		"UNBOUND": {"assemble", 0, 0},
	}
	if len(m.Seats) != len(want) {
		t.Fatalf("seats = %d, want %d: %+v", len(m.Seats), len(want), m.Seats)
	}
	for _, s := range m.Seats {
		w := want[s.AgentID]
		if s.Label != w.label || s.Epoch != w.epoch || s.Sitting != w.sitting {
			t.Errorf("%s: label %q epoch %d sitting %d, want %q/%d/%d", s.AgentID, s.Label, s.Epoch, s.Sitting, w.label, w.epoch, w.sitting)
		}
	}

	// AND THE COST BREAKDOWN BUCKETS BY THE SAME EPOCH: the two lenses fall in epochs 0 and 2, the
	// two chair sittings in 1 and 2, the unbound agent in the dash bucket (epoch 0).
	got := map[int]int{}
	for _, r := range m.CostRows {
		got[r.Epoch] += r.Agents
	}
	if got[0] != 2 || got[1] != 1 || got[2] != 2 {
		t.Errorf("cost agents by epoch = %v, want {0:2 1:1 2:2}: %+v", got, m.CostRows)
	}
}
