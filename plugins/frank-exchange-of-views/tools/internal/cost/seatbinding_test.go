package cost

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

func register(t *testing.T, seat, agent string) *recordpb.Event {
	t.Helper()
	reg := &recordpb.Register{}
	if agent != "" {
		reg.AgentId = proto.String(agent)
	}
	return recordtest.Event(t, seat, reg)
}

// THE EPOCH AND THE SITTING ARE COUNTED OFF THE REGISTERS, NOT STAMPED. A lens that sits before
// the chair has ever sat is in epoch 0; the chair's own register opens the epoch it is in; a
// seat's second register is its sitting 2 whatever epoch it lands in.
func TestSeatBindingsCountTheWindowsOffTheRegisters(t *testing.T) {
	evs := []*record.Event{
		register(t, "red-lens-evidence", "L1"), // epoch 0, sitting 1
		register(t, "red-chair", "C1"),         // opens epoch 1, sitting 1
		register(t, "blue-respond", "B1"),      // epoch 1, sitting 1
		register(t, "red-chair", "C2"),         // opens epoch 2, sitting 2
		register(t, "red-lens-evidence", "L2"), // epoch 2, sitting 2
		register(t, "judge", ""),               // no agent id: binds nothing
	}
	got := SeatBindingsOf(evs)
	want := map[string]SeatBinding{
		"L1": {SeatID: "red-lens-evidence", Epoch: 0, Sitting: 1},
		"C1": {SeatID: "red-chair", Epoch: 1, Sitting: 1},
		"B1": {SeatID: "blue-respond", Epoch: 1, Sitting: 1},
		"C2": {SeatID: "red-chair", Epoch: 2, Sitting: 2},
		"L2": {SeatID: "red-lens-evidence", Epoch: 2, Sitting: 2},
	}
	if len(got) != len(want) {
		t.Errorf("bound %d agents, want %d: %+v", len(got), len(want), got)
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("%s = %+v, want %+v", id, got[id], w)
		}
	}
}

// THE LAST REGISTER WINS, as record.SeatOfAgent has it: a resumed seat arrives under a new agent
// id claiming a seat already bound, and an agent that registers twice is bound to its latest claim.
func TestSeatBindingsTakeTheLatestRegister(t *testing.T) {
	got := SeatBindingsOf([]*record.Event{
		register(t, "red-chair", "A"),
		register(t, "red-chair", "A"),
	})
	if got["A"] != (SeatBinding{SeatID: "red-chair", Epoch: 2, Sitting: 2}) {
		t.Errorf("A = %+v, want the second register's window", got["A"])
	}
}

func TestAgentIDOfTranscript(t *testing.T) {
	cases := map[string]string{"agent-abc.jsonl": "abc", "journal.jsonl": "", "agent-x.txt": "", "agent-.jsonl": ""}
	for in, want := range cases {
		if got := AgentIDOfTranscript(in); got != want {
			t.Errorf("AgentIDOfTranscript(%q) = %q, want %q", in, got, want)
		}
	}
}

// THE REPORT'S EPOCH COLUMN COMES FROM THE RECORD. Two transcripts whose heads say nothing about
// a round land in the epochs their registers sat in; a transcript the record never bound prints a
// dash rather than a guessed number. (The head is an unrounded seatclass needle on purpose: the
// seat CLASS is still the prompt head's to say, and this test is about the epoch, not the class.)
func TestReportBindsEachTranscriptToItsEpochFromTheRecord(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir,
		register(t, "red-chair", "C1"),
		register(t, "judge-terminal", "J1"),
		register(t, "red-chair", "C2"),
		register(t, "judge-terminal", "J2"),
	)
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	tr := t.TempDir()
	head := `{"message":{"role":"user","content":"Terminal dispute disposition. FIRST ACTION"}}` + "\n"
	for _, id := range []string{"J1", "J2", "STRAY"} {
		if err := os.WriteFile(filepath.Join(tr, "agent-"+id+".jsonl"),
			[]byte(head+usageLine("claude-haiku-4-5", 1000000, 0, 0, 0)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var b bytes.Buffer
	if err := Report(tr, run, &b); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"## Per seat-epoch", "| epoch | seat |", "| 1 | judge-terminal | haiku | 1 |", "| 2 | judge-terminal | haiku | 1 |", "| — | judge-terminal | haiku | 1 |"} {
		if !strings.Contains(out, want) {
			t.Errorf("cost.md lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "round") {
		t.Errorf("cost.md still speaks of rounds:\n%s", out)
	}
}
