package cli

import (
	"regexp"
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
	gap := mintGap(t, runDir, "G1", "spec_underspecification")
	other := mintGap(t, runDir, "G2", "derivation_status_overclaim")
	closeWith := func(anchors string) error {
		_, err := runAt(t, "close", "--run", runDir, "--seat-id", lensSeat,
			"--id", "G1", "--as", "repaired", "--verified-by", lensSeat, "--verified-with", "show report",
			"--verified-against", "blue/report.md", "--anchor", anchors, "--reason", "read at the leaf")
		return err
	}
	if err := closeWith("g-0dd40334"); err == nil || !strings.Contains(err.Error(), "--anchor") || !strings.Contains(err.Error(), "is not anchor") {
		t.Fatalf("an anchor in no id's shape must be refused naming --anchor; got %v", err)
	}
	if err := closeWith("G-0dd40334"); err == nil || !strings.Contains(err.Error(), "which no act on this record placed") {
		t.Fatalf("an anchor the record does not hold must be refused at the write; got %v", err)
	}
	if err := closeWith("Q-0dd40334"); err == nil || !strings.Contains(err.Error(), "which no act on this record placed") {
		t.Fatalf("an inquiry the record does not hold must be refused at the write; got %v", err)
	}
	// A GAP IS NOT EVIDENCE OF ITS OWN REPAIR (m30: G-dc877265 named itself, and so did G-97b44713
	// and G-968365ad): the closure is verified WITH something other than what it closes.
	if err := closeWith(anchor.Token(gap)); err == nil || !strings.Contains(err.Error(), "the gap this closes") {
		t.Fatalf("a closure naming the gap it closes as its anchor must be refused at the write; got %v", err)
	}
	if err := closeWith(anchor.Token(other)); err != nil {
		t.Fatalf("a closure naming the token of an anchor the record holds was refused: %v", err)
	}
	var got []string
	for _, e := range events(t, runDir) {
		if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok {
			got = c.GetAnchors()
		}
	}
	if len(got) != 1 || got[0] != other {
		t.Errorf("the closure stores anchors %v, want [%s] — the id, its token unwrapped", got, other)
	}
}

// AN INQUIRY IS SOMETHING A CLOSURE RESTS ON (gblock, 2026-10-19): its fate is a record act, and m30's
// G-33401edc and G-5d729d0a, verified against inquiries, could name no anchor at all.
func TestAClosureMayRestOnAnInquiry(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	mintGap(t, runDir, "G1", "spec_underspecification")
	registerBlue(t, runDir)
	out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", blueSeat, "--question", "whether trial division to 3 settles it")
	if err != nil {
		t.Fatalf("inquiry propose: %v\n%s", err, out)
	}
	q := regexp.MustCompile(`Q-[0-9a-f]{8}`).FindString(out)
	if q == "" {
		t.Fatalf("propose returned no inquiry id: %q", out)
	}
	if _, err := runAt(t, "close", "--run", runDir, "--seat-id", lensSeat,
		"--id", "G1", "--as", "repaired", "--verified-by", lensSeat, "--verified-with", "show inquiries",
		"--verified-against", "the inquiry's recorded fate", "--anchor", q, "--reason", "read at the leaf"); err != nil {
		t.Fatalf("a closure naming an inquiry the record holds was refused: %v", err)
	}
	var got []string
	for _, e := range events(t, runDir) {
		if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok {
			got = c.GetAnchors()
		}
	}
	if len(got) != 1 || got[0] != q {
		t.Errorf("the closure stores anchors %v, want [%s]", got, q)
	}
}

func TestADeclarationsGapsAreHeldToTheBoard(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	gap := mintGap(t, runDir, "G1", "spec_underspecification")
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
