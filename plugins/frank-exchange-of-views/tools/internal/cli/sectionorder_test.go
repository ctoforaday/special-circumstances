package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// A SECTION OUT OF THE ORDER THE REPORT SHIPS IN IS REFUSED AT THE WRITE, through the real verbs.
//
// universe-m21: blue wrote the Catechism after its argument ("These four sources…"). Assembly
// placed it ahead of the sections that introduce those sources, and every seat had read it after
// them. The text ships as it is written and the two acts that write it hold the order.

const (
	orderHead = "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n"
	orderCat  = "## The Catechism\n\nThese two factors settle it<!--cite:C-0000000a-->.\n\n"
	orderOwn1 = "## Is 91 divisible by a small prime?\n\nYes: 7 divides it<!--gap:G-0000000b-->. The quotient is 13<!--computation:K-0000000c-->.\n\n"
	orderOwn2 = "## What a Fermat test would have shown\n\nNothing decisive<!--fx:F-0000000d-->.\n\n"
	orderOpen = "## Open questions\n\nNone.\n"
)

func refusesTheOrder(t *testing.T, out string, err error, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("report text holding a section out of order was recorded:\n%s", out)
	}
	msg := err.Error() + "\n" + out
	for _, w := range append(want, report.TextOrder, "NOTHING IS RECORDED", "WHAT TO DO: move the section to its place, whole") {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal does not say %q:\n%s", w, msg)
		}
	}
}

func TestIngestRefusesASectionOutOfOrder(t *testing.T) {
	runDir := newRun(t)
	path := filepath.Join(runDir, "blue", "report.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("register: %v", err)
	}
	for _, tc := range []struct {
		doc  string
		want []string
	}{
		{"# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n## Is 91 divisible by a small prime?\n\nYes.\n\n## The Catechism\n\nThese two factors settle it.\n\n" + orderOpen,
			[]string{`"## The Catechism" after "## Is 91 divisible by a small prime?"`, "nothing in the report's text but the title line (`# …`) and `## TL;DR` stands before it"}},
		{"# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n" + orderOpen + "\n## Appendix\n\nMore.\n",
			[]string{`"## Open questions" before "## Appendix"`, "nothing in the report's text stands after it"}},
		{"# On 91\n\nA framing paragraph.\n\n## TL;DR\n\n91 is 7 x 13.\n",
			[]string{`"## TL;DR" after "A framing paragraph."`}},
	} {
		if err := os.WriteFile(path, []byte(tc.doc), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
		refusesTheOrder(t, out, err, tc.want...)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("the refused ingest did not leave the file for its author to fix: %v", err)
		}
	}
	// What a section stands beside can be a paragraph: the refusal quotes its opening, not all of it.
	long := "A framing paragraph " + strings.Repeat("that runs on ", 30) + "to its last words."
	if err := os.WriteFile(path, []byte("# On 91\n\n"+long+"\n\n## TL;DR\n\n91 is 7 x 13.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	refusesTheOrder(t, out, err, `"## TL;DR" after "A framing paragraph that runs on`, "…")
	if err != nil && strings.Contains(err.Error()+out, "to its last words") {
		t.Errorf("the refusal quotes the whole paragraph the section stands after:\n%v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BASE_INGEST); n != 0 {
		t.Errorf("the refused ingests recorded %d base(s)", n)
	}
	// A report that lacks the three is in order, and the one ingest a run gets is unspent.
	if err := os.WriteFile(path, []byte("# On 91\n\n## What was asked\n\nWhether 91 is prime.\n\n## What was found\n\nIt is not.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("a report with none of the three sections was refused: %v\n%s", err, out)
	}
}

func TestEditRefusesASectionOutOfOrderAndMovesOneInOneEdit(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, orderHead+orderCat+orderOwn1+orderOwn2+orderOpen)
	registerBlue(t, runDir)
	edit := func(key, quote, text string) (string, error) {
		return run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", key, "--quote", quote, "--new", text, "--reason", "the order of the sections")
	}
	span := func(parts ...string) string { return strings.TrimSpace(strings.Join(parts, "")) }

	// The Catechism moved after a section of the report's own; a section written after Open
	// questions; and a heading the edit completes from parts.
	out, err := edit("O1", span(orderCat, orderOwn1), span(orderOwn1, orderCat))
	refusesTheOrder(t, out, err, `"## The Catechism" after "## Is 91 divisible by a small prime?"`)
	out, err = edit("O2", "None.", "None.\n\n## Appendix\n\nMore arithmetic.")
	refusesTheOrder(t, out, err, `"## Open questions" before "## Appendix"`)
	out, err = edit("O3", "What a Fermat test would have shown", "TL;DR")
	refusesTheOrder(t, out, err, `"## TL;DR" after "## TL;DR"`)
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 0 {
		t.Fatalf("the refused edits recorded %d edit(s)", n)
	}

	// A SECTION MOVES IN ONE EDIT, every anchor on its sentence and none reopened: the span is the
	// section and what stands between it and its place, written again with the section moved.
	if out, err := edit("O4", span(orderOwn1, orderOwn2), span(orderOwn2, orderOwn1)); err != nil {
		t.Fatalf("two of the report's own sections were not swapped in one edit: %v\n%s", err, out)
	}
	got, err := reportproj.RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if want := orderHead + orderCat + orderOwn2 + orderOwn1 + orderOpen; got != want {
		t.Errorf("the report after the move reads\n%s\nwant\n%s", got, want)
	}
	if ed := lastBody(t, runDir, &recordpb.BlueEdit{}); len(ed.GetReopened()) != 0 {
		t.Errorf("moving two sections whole reopened %v: no sentence changed", ed.GetReopened())
	}
}

// A RECORD MIGRATED WITH ITS TEXT AS WRITTEN TAKES EVERY EDIT THAT ADDS NOTHING OUT OF ORDER, the
// one that moves its misplaced section to its place included.
func TestARecordHoldingASectionOutOfOrderStillTakesEdits(t *testing.T) {
	runDir := newRun(t)
	id := record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(orderHead + orderOwn1 + orderCat + orderOwn2 + orderOpen)}); err != nil {
		t.Fatal(err)
	}
	registerBlue(t, runDir)
	edit := func(key, quote, text string) (string, error) {
		return run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", key, "--quote", quote, "--new", text, "--reason", "the order of the sections")
	}
	if out, err := edit("L1", "91 is 7 x 13", "91 is 7 times 13"); err != nil {
		t.Fatalf("an edit elsewhere in a record that holds a section out of order was refused: %v\n%s", err, out)
	}
	out, err := edit("L2", "None.", "None.\n\n## Appendix\n\nMore arithmetic.")
	refusesTheOrder(t, out, err, `"## Open questions" before "## Appendix"`)
	one, cat := strings.TrimSpace(orderOwn1), strings.TrimSpace(orderCat)
	if out, err := edit("L3", one+"\n\n"+cat, cat+"\n\n"+one); err != nil {
		t.Fatalf("the edit that moves the misplaced section to its place was refused: %v\n%s", err, out)
	}
	got, err := reportproj.RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if _, found := report.OutOfOrderIn(report.ParseText(""), report.ParseText(got)); found {
		t.Errorf("the report is still out of order after the move:\n%s", got)
	}
	if ed := lastBody(t, runDir, &recordpb.BlueEdit{}); len(ed.GetReopened()) != 0 {
		t.Errorf("moving the section whole reopened %v", ed.GetReopened())
	}
}
