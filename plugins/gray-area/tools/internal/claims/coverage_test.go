package claims

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
)

type fakeEntry struct {
	name string
	dir  bool
}

func (f fakeEntry) Name() string               { return f.name }
func (f fakeEntry) IsDir() bool                { return f.dir }
func (f fakeEntry) Type() fs.FileMode          { return 0 }
func (f fakeEntry) Info() (fs.FileInfo, error) { return nil, nil }

func dirOf(names ...string) func(string) ([]os.DirEntry, error) {
	return func(string) ([]os.DirEntry, error) {
		var out []os.DirEntry
		for _, n := range names {
			out = append(out, fakeEntry{name: n})
		}
		return out, nil
	}
}

func absent(string) ([]os.DirEntry, error) { return nil, errors.New("no such file or directory") }

// The directory comes from the path the harness handed over, not from a search.
func TestTheSeatDirectoryIsDerivedFromTheHandedOverPath(t *testing.T) {
	got := SubagentDir("/root/.claude/projects/-repo/937047bc.jsonl")
	want := "/root/.claude/projects/-repo/937047bc/subagents"
	if got != want {
		t.Errorf("SubagentDir = %q, want %q", got, want)
	}
	for _, bad := range []string{"", "   ", "/root/projects/937047bc", "/root/projects/937047bc.json"} {
		if d := SubagentDir(bad); d != "" {
			t.Errorf("assembled a directory from an untransformable path %q -> %q — it would read as empty", bad, d)
		}
	}
}

// The answer #469 is about: a transcript on disk that no row names.
func TestAnUnnamedTranscriptIsFound(t *testing.T) {
	c := Reconcile("/p/s.jsonl", []string{"a1", "a2"},
		dirOf("agent-a1.jsonl", "agent-a1.meta.json", "agent-a2.jsonl", "agent-a9.jsonl"))
	if !c.Measured {
		t.Fatalf("not measured: %s", c.Why)
	}
	if len(c.Unnamed) != 1 || c.Unnamed[0] != "agent-a9" {
		t.Errorf("unnamed = %v, want [agent-a9]", c.Unnamed)
	}
	if len(c.Missing) != 0 {
		t.Errorf("missing = %v, want none", c.Missing)
	}
	// The sidecar is not a transcript and must not be counted as one.
	if len(c.OnDisk) != 3 {
		t.Errorf("on_disk = %v — a .meta.json was counted as a transcript", c.OnDisk)
	}
}

// The other direction, reported so that #189's 19/19 stops being an assumption
// the tool cannot contradict.
func TestARowNamingAMissingFileIsReportedToo(t *testing.T) {
	c := Reconcile("/p/s.jsonl", []string{"a1", "gone"}, dirOf("agent-a1.jsonl"))
	if !c.Measured {
		t.Fatalf("not measured: %s", c.Why)
	}
	if len(c.Missing) != 1 || c.Missing[0] != "agent-gone" {
		t.Errorf("missing = %v, want [agent-gone]", c.Missing)
	}
}

// THE DISCRIMINATOR, and the reason Measured exists. A missing directory means
// two completely different things depending on whether any seat was recorded, and
// the wrong one of them is the plausible zero this plugin exists to prevent.
func TestAMissingDirectoryIsNotAutomaticallyACleanBoard(t *testing.T) {
	// Seats recorded, directory unreadable: UNKNOWN, loudly.
	bad := Reconcile("/p/s.jsonl", []string{"a1", "a2"}, absent)
	if bad.Measured {
		t.Fatal("reported a measurement from a directory it could not read")
	}
	if !strings.Contains(bad.Why, "2 seat(s)") || !strings.Contains(bad.Why, "guess") {
		t.Errorf("the refusal does not say what is wrong or why it matters: %q", bad.Why)
	}
	if len(bad.Unnamed) != 0 {
		t.Errorf("an unmeasured result carried findings: %v", bad.Unnamed)
	}

	// No seats and no directory: consistent, and a real (boring) measurement.
	ok := Reconcile("/p/s.jsonl", nil, absent)
	if !ok.Measured {
		t.Errorf("a session with no seats and no directory is consistent, not unknown: %q", ok.Why)
	}
	if ok.Why == "" {
		t.Error("a clean-but-empty answer must still say why it is empty")
	}
}

// An underivable path is never silently treated as an empty directory.
func TestAnUnderivablePathIsUnmeasuredNotEmpty(t *testing.T) {
	c := Reconcile("/p/not-a-transcript", []string{"a1"}, dirOf("agent-a1.jsonl"))
	if c.Measured {
		t.Fatal("measured against a directory it never derived")
	}
	if !strings.Contains(c.Why, "cannot be derived") {
		t.Errorf("why = %q", c.Why)
	}
}

// Empty ids in the manifest must not become a bucket named "agent-".
func TestEmptyAgentIDsAreNotNamed(t *testing.T) {
	c := Reconcile("/p/s.jsonl", []string{"", "a1", ""}, dirOf("agent-a1.jsonl"))
	for _, n := range c.Named {
		if n == "agent-" {
			t.Fatalf("an empty id became a name: %v", c.Named)
		}
	}
	if len(c.Named) != 1 {
		t.Errorf("named = %v, want one", c.Named)
	}
}

// THE FIRST REAL RUN OF THIS TOOL REPRODUCED THE MISCOUNT IT MEASURES.
//
// Filtering on `kind` alone reported 165 seats against 20 transcripts and listed
// 146 phantom MISSING files — #189's miscount, made by the tool built to measure
// its consequences. The code comment predicted it; the code did it anyway. Rows
// written before schema 4 called every SubagentStop a seat, so `kind` on those
// rows is not evidence and must be re-derived from the row's own two fields.
func TestPreSchema4RowsAreReclassifiedFromTheirOwnFields(t *testing.T) {
	m := strings.Join([]string{
		// schema 4: kind is authoritative.
		`{"schema":4,"kind":"seat","agent_id":"s4seat","agent_type":"Explore","resolved":true}`,
		`{"schema":4,"kind":"turn-end","agent_id":"s4turn"}`,
		`{"schema":4,"kind":"session","session_id":"x"}`,
		// pre-4: kind says "seat" for both populations. The conjunction decides.
		`{"schema":1,"kind":"seat","agent_id":"oldseat","agent_type":"claude","resolved":true}`,
		`{"schema":1,"kind":"seat","agent_id":"oldturn","agent_type":"","resolved":false}`,
		// pre-4, no type but the file WAS there: a real trajectory, keep it.
		`{"schema":2,"kind":"seat","agent_id":"oldnameless","agent_type":"","resolved":true}`,
		// pre-4, typed but unresolved: the alarm case — still a seat.
		`{"schema":3,"kind":"seat","agent_id":"oldmissing","agent_type":"red","resolved":false}`,
		`not json at all`,
	}, "\n")
	seats, unclassifiable := SeatIDs([]byte(m))
	want := []string{"oldmissing", "oldnameless", "oldseat", "s4seat"}
	if len(seats) != len(want) {
		t.Fatalf("SeatIDs = %v, want %v", seats, want)
	}
	for i := range want {
		if seats[i] != want[i] {
			t.Fatalf("SeatIDs = %v, want %v", seats, want)
		}
	}
	if unclassifiable != 0 {
		t.Errorf("unclassifiable = %d, want 0", unclassifiable)
	}
	// The specific wrong answer must not recur.
	for _, s := range seats {
		if s == "oldturn" || s == "s4turn" {
			t.Errorf("a turn end was counted as a seat: %v — this is the 165-vs-19 miscount", seats)
		}
	}
}

// A seat row with no id reconciles with nothing. Counted, never folded into
// either side: silently a seat is a phantom MISSING, silently dropped is an
// understated denominator.
func TestSeatRowsWithNoIDAreCountedNotDropped(t *testing.T) {
	m := `{"schema":4,"kind":"seat","agent_id":"a1"}` + "\n" + `{"schema":4,"kind":"seat","agent_id":""}`
	seats, unclassifiable := SeatIDs([]byte(m))
	if len(seats) != 1 || seats[0] != "a1" {
		t.Errorf("seats = %v", seats)
	}
	if unclassifiable != 1 {
		t.Errorf("unclassifiable = %d, want 1", unclassifiable)
	}
}
