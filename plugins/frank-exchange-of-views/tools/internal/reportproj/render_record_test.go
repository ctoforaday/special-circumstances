package reportproj

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

func ident(t *testing.T, runDir, seat string) record.Identity {
	t.Helper()
	run := runtest.Open(t, runDir)
	return record.Identity{Run: run, SeatID: seat}
}

func TestRenderFromRecordFoldsBasePlusDiffStack(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, s := range []string{"blue-synthesize", "blue-respond"} {
		if _, _, err := record.RegisterSeat(ident(t, runDir, s), ""); err != nil {
			t.Fatalf("register %s: %v", s, err)
		}
	}
	base := "The value is stable across the whole range of inputs."
	if _, err := record.Append(ident(t, runDir, "blue-synthesize"), &recordpb.BaseIngest{Text: proto.String(base)}); err != nil {
		t.Fatalf("append base: %v", err)
	}
	edits := []Op{
		{Old: "stable", New: "steady"},
		{Old: "the whole range of inputs", New: "every admissible input"},
	}
	for _, e := range edits {
		be := &recordpb.BlueEdit{Old: proto.String(e.Old), New: proto.String(e.New)}
		if _, err := record.Append(ident(t, runDir, "blue-respond"), be); err != nil {
			t.Fatalf("append edit %q: %v", e.Old, err)
		}
	}

	got, err := RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("RenderFromRecord: %v", err)
	}
	want, err := Render(base, edits)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got != want {
		t.Errorf("RenderFromRecord did not reproduce base+stack:\n  want %q\n  got  %q", want, got)
	}
	if want == base {
		t.Fatal("test is vacuous — the edits changed nothing")
	}
}

func TestRenderFromRecordIsLoudWithNoBase(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	_, err := RenderFromRecord(runtest.Open(t, runDir))
	if err == nil {
		t.Fatal("rendering a run with no ingested base must be a loud error, not an empty report")
	}
}

// renderOneEdit records a base and ONE blue edit event as given, and renders the record.
func renderOneEdit(t *testing.T, base string, be *recordpb.BlueEdit) string {
	t.Helper()
	runDir := recordtest.TmpRun(t)
	for _, s := range []string{"blue-synthesize", "blue-respond"} {
		if _, _, err := record.RegisterSeat(ident(t, runDir, s), ""); err != nil {
			t.Fatalf("register %s: %v", s, err)
		}
	}
	if _, err := record.Append(ident(t, runDir, "blue-synthesize"), &recordpb.BaseIngest{Text: proto.String(base)}); err != nil {
		t.Fatalf("append base: %v", err)
	}
	if _, err := record.Append(ident(t, runDir, "blue-respond"), be); err != nil {
		t.Fatalf("append edit: %v", err)
	}
	got, err := RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("RenderFromRecord: %v", err)
	}
	return got
}

// AN EVENT WRITTEN WITHOUT exact_span REPLAYS AS IT ALWAYS DID — including the #861 B3 shape, an
// edit whose trimmed span was a no-op. Recorded before the verb refused those, it rendered nothing
// then and must render nothing now: re-deciding the span at replay would rewrite a report that
// already shipped. The decision is the event's, never the renderer's.
func TestALegacyNoOpEditStillReplaysAsANoOp(t *testing.T) {
	base := "Intro.\n\non their own.).\n"
	got := renderOneEdit(t, base, &recordpb.BlueEdit{Old: proto.String("on their own.)."), New: proto.String("on their own.)")})
	if got != base {
		t.Errorf("a recorded no-op without exact_span replayed as a change:\n  want %q\n  got  %q", base, got)
	}
}

// AN EVENT WITH exact_span REPLAYS ON THE LITERAL SPAN, reproducing what the verb planned — here at
// the end of the document, where no text follows the terminator to quote through.
func TestAnExactSpanEditReplaysOnTheLiteralSpan(t *testing.T) {
	base := "Intro.\n\non their own.).\n"
	old, new := "on their own.).", "on their own.)"
	planned, exact, err := PlanSplice("blue edit", base, old, new)
	if err != nil || !exact {
		t.Fatalf("the planner did not choose the literal span: exact=%v err=%v", exact, err)
	}
	got := renderOneEdit(t, base, &recordpb.BlueEdit{Old: proto.String(old), New: proto.String(new), ExactSpan: proto.Bool(true)})
	if got != planned || got != "Intro.\n\non their own.)\n" {
		t.Errorf("replay drifted from the plan:\n  plan   %q\n  replay %q", planned, got)
	}
}

// A LITERAL QUOTE THAT NO LONGER OCCURS ONCE IS A LOUD REPLAY FAILURE, never a silent fallback to
// the trimmed span — that fallback would render a report no verb ever produced.
func TestAnExactSpanEditThatCannotLocateIsLoud(t *testing.T) {
	_, err := Render("on their own.).\n\non their own.).\n", []Op{{Old: "on their own.).", New: "on their own.)", Exact: true}})
	if err == nil {
		t.Fatal("an exact-span edit whose quote occurs twice replayed without error")
	}
}
