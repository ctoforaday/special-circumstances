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

// registerFor is register with an OCCASION — what the sitting was convened to do, which only the
// bench carries and which the cost report keys on so the assembly's spend stays out of a docket
// ruling's row.
func registerFor(t *testing.T, seat, agent string, occ recordpb.Occasion) *recordpb.Event {
	t.Helper()
	reg := &recordpb.Register{Occasion: occ.Enum()}
	if agent != "" {
		reg.AgentId = proto.String(agent)
	}
	return recordtest.Event(t, seat, reg)
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
		registerFor(t, "judge", "J1", recordpb.Occasion_OCCASION_TERMINAL),
		register(t, "red-chair", "C2"),
		registerFor(t, "judge", "J2", recordpb.Occasion_OCCASION_TERMINAL),
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
	// THE SEAT AND THE SITTING ARE TWO COLUMNS' WORTH OF FACT IN ONE CELL, and both come from the
	// RECORD now. `judge · terminal` is the seat id the register named plus the occasion it
	// carried; the STRAY agent the record never bound has neither an epoch nor an occasion, and
	// reads as a bare `judge` under a dash — which is the honest answer, not a merged row.
	for _, want := range []string{"## Per seat-epoch", "| epoch | seat |", "| 1 | judge · terminal | haiku | 1 |", "| 2 | judge · terminal | haiku | 1 |", "| — | judge | haiku | 1 |"} {
		if !strings.Contains(out, want) {
			t.Errorf("cost.md lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "round") {
		t.Errorf("cost.md still speaks of rounds:\n%s", out)
	}
}

// AN AGENT WHOSE LATEST BRACKET NAMES NO SEAT IS UNBOUND, and its spend reads as such: the dash in
// no epoch, never the seat an earlier bracket named. The rule is the record's (agent_binding); this
// is where the cost report meets it.
func TestReportLeavesAnAgentWhoseLatestBracketNamesNoSeatUnbound(t *testing.T) {
	bracket := func(seat string) *recordpb.Event {
		open := &recordpb.SittingOpen{AgentId: proto.String("E"), AgentType: proto.String("frank-exchange-of-views:x")}
		if seat != "" {
			open.SeatId = proto.String(seat)
		}
		return recordtest.Event(t, record.HarnessSeat, open)
	}
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir, register(t, "red-chair", "C1"), bracket("red-lens-voice"), bracket(""))
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	tr := t.TempDir()
	head := `{"message":{"role":"user","content":"Terminal dispute disposition. FIRST ACTION"}}` + "\n"
	if err := os.WriteFile(filepath.Join(tr, "agent-E.jsonl"), []byte(head+usageLine("claude-haiku-4-5", 1000000, 0, 0, 0)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err := Report(tr, run, &b); err != nil {
		t.Fatal(err)
	}
	if out := b.String(); !strings.Contains(out, "| — | judge | haiku | 1 |") || strings.Contains(out, "\n| 1 | ") {
		// The seat-naming bracket sat in epoch 1, so the earlier seat's binding would read epoch 1.
		t.Errorf("the unbound agent's row must read the dash, in no epoch:\n%s", out)
	}
}
