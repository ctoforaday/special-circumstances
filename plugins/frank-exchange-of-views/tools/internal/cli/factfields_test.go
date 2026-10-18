package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A CLOSURE'S ANCHORS AND A DECLARATION'S GAPS ARE FIELDS THE WRITE REFUSES. Each was read out of the
// seat's prose by an id's shape: an id typed in another shape read as none, and one the record did
// not hold was caught only after the run, if at all. Each now arrives on its own flag, is refused in
// a shape no id has, and is refused unless the record holds it — and what is stored is the id, the
// report's token around it unwrapped.
func TestAClosuresAnchorsAreHeldToTheRecord(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	gap := mintGap(t, runDir, "G1", "scope-creep")
	closeWith := func(anchors string) error {
		_, err := runAt(t, "close", "--run", runDir, "--seat-id", lensSeat,
			"--id", "G1", "--as", "repaired", "--verified-by", lensSeat, "--verified-with", "show report",
			"--verified-against", "blue/report.md", "--anchor", anchors, "--reason", "read at the leaf")
		return err
	}
	if err := closeWith("g-0dd40334"); err == nil || !strings.Contains(err.Error(), "--anchor") || !strings.Contains(err.Error(), "is not anchor") {
		t.Fatalf("an anchor in no id's shape must be refused naming --anchor; got %v", err)
	}
	if err := closeWith("G2"); err == nil || !strings.Contains(err.Error(), "which no act on this record placed") {
		t.Fatalf("an anchor the record does not hold must be refused at the write; got %v", err)
	}
	if err := closeWith(anchor.Token(gap)); err != nil {
		t.Fatalf("a closure naming the token of an anchor the record holds was refused: %v", err)
	}
	var got []string
	for _, e := range events(t, runDir) {
		if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok {
			got = c.GetAnchors()
		}
	}
	if len(got) != 1 || got[0] != gap {
		t.Errorf("the closure stores anchors %v, want [%s] — the id, its token unwrapped", got, gap)
	}
}

func TestADeclarationsGapsAreHeldToTheBoard(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	gap := mintGap(t, runDir, "G1", "scope-creep")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "terminal"); err != nil {
		t.Fatal(err)
	}
	declare := func(ids string) error {
		_, err := runAt(t, "declare", "--run", runDir, "--seat-id", "judge", "--ids", ids,
			"--reason", "the term is construed this way for the rest of the run")
		return err
	}
	if err := declare("g1"); err == nil || !strings.Contains(err.Error(), "--ids") || !strings.Contains(err.Error(), "is not gap-id") {
		t.Fatalf("a gap id in no gap's shape must be refused naming --ids; got %v", err)
	}
	if err := declare("G1,G2"); err == nil || !strings.Contains(err.Error(), "no mint event created") {
		t.Fatalf("a gap the board does not hold must be refused at the write; got %v", err)
	}
	if err := declare("G1"); err != nil {
		t.Fatalf("a declaration about a gap the board holds was refused: %v", err)
	}
	var got []string
	for _, e := range events(t, runDir) {
		if d, ok := recordpb.BodyAs[*recordpb.Declare](e); ok {
			got = d.GetGapIds()
		}
	}
	if len(got) != 1 || got[0] != gap {
		t.Errorf("the declaration stores gap_ids %v, want [%s]", got, gap)
	}
}
