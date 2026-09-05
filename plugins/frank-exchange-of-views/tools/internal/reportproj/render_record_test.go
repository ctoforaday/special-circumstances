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
	return record.Identity{Run: run, SeatID: seat, Round: record.RoundIn(run)(seat)}
}

func TestRenderFromRecordFoldsBasePlusDiffStack(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, s := range []string{"blue-synthesize", "blue-respond-r1"} {
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
		if _, err := record.Append(ident(t, runDir, "blue-respond-r1"), be); err != nil {
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
