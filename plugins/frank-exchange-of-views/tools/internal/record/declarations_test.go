package record

import (
	"encoding/json"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE BOARD STATES A DECLARED GAP'S STATUS, AND THE HOLDING'S PROSE DOES NOT MOVE.
//
// The bench declares while two gaps are open and writes that one "remains open" — the sentence that
// went stale on m25. One gap then closes by its lens and the other by a bench ruling, and every
// seat's list must say so beside the unchanged prose: the defect this exists for is a status read
// from the holding, which keeps reading "open" after the gap has closed.
func TestADeclaredGapCarriesTheBoardsStatusOnEverySeatsList(t *testing.T) {
	const (
		byLens  = "G-0000000a" // the lens closes it
		byBench = "G-0000000b" // the bench rules it closed
		notAGap = "G-0000ffff" // named in the holding, minted by nobody
		// byBench is named by its report token, the form a bench copies out of the report.
		holding  = byLens + " remains open and is material; " + "<!--gap:" + byBench + "--> is construed with it, and " + notAGap + " is cited. " + byLens + " again."
		laterOne = "the later holding names no gap"
	)
	run := corrRun(t)
	chair := sit(t, run, "red-chair")
	lens := sit(t, run, evLens)
	for _, id := range []string{byLens, byBench} {
		m := corrMint(id)
		m.Problem = proto.String(map[string]string{byLens: "the sample size is overstated", byBench: "a quoted figure has no source"}[id])
		mustAppend(t, lens, m)
	}
	sit(t, run, "blue-respond")
	bench := sit(t, run, "judge")
	mustAppend(t, bench, &recordpb.Declare{Holding: proto.String(holding)})

	seats := []struct{ role, seat string }{{"lens", evLens}, {"blue", "blue-respond"}, {"chair", "red-chair"}, {"bench", "judge"}}
	read := func(role, seat string) []DeclarationJSON {
		t.Helper()
		// THE BYTES A SEAT IS HANDED, not the struct: the hook injects WorkJSONBytes.
		b, err := WorkJSONBytes(run, role, seat)
		if err != nil {
			t.Fatal(err)
		}
		var w WorkJSON
		if err := json.Unmarshal(b, &w); err != nil {
			t.Fatal(err)
		}
		return w.Declarations
	}

	before := []DeclaredGapJSON{{ID: byLens, Status: "open"}, {ID: byBench, Status: "open"}, {ID: notAGap, Status: "not_on_board"}}
	for _, s := range seats {
		got := read(s.role, s.seat)
		if len(got) != 1 {
			t.Fatalf("%s's list carries %d declarations, want the one on the record: %+v", s.seat, len(got), got)
		}
		if got[0].Holding != holding {
			t.Errorf("%s's list restates the holding as %q, want the bench's text", s.seat, got[0].Holding)
		}
		if got[0].Epoch < 1 {
			t.Errorf("%s's list carries the declaration at epoch %d", s.seat, got[0].Epoch)
		}
		if !reflect.DeepEqual(got[0].Gaps, before) {
			t.Errorf("%s: before either gap moved, the named gaps read\n%+v\nwant\n%+v", s.seat, got[0].Gaps, before)
		}
	}

	mustAppend(t, lens, &recordpb.Close{GapId: proto.String(byLens), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
		AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("verified")})
	mustAppend(t, chair, docketMotion("M1", byBench))
	bench = sit(t, run, "judge")
	mustAppend(t, bench, docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT))
	mustAppend(t, bench, &recordpb.Declare{Holding: proto.String(laterOne)})

	for _, s := range seats {
		got := read(s.role, s.seat)
		if len(got) != 2 || got[0].Holding != holding || got[1].Holding != laterOne {
			t.Fatalf("%s's list does not carry both declarations in record order, prose unchanged: %+v", s.seat, got)
		}
		g := got[0].Gaps
		if len(g) != 3 {
			t.Fatalf("%s: the first declaration names %d gaps, want 3: %+v", s.seat, len(g), g)
		}
		if g[0].ID != byLens || g[0].Status != "closed" || g[0].Fate != "repaired" || g[0].ClosedBy != "red" || g[0].ClosedEpoch < 1 {
			t.Errorf("%s: the gap its lens closed reads %+v — the holding says it remains open, and the board must say otherwise", s.seat, g[0])
		}
		if g[1].ID != byBench || g[1].Status != "closed" || g[1].Fate != "not_a_defect" || g[1].ClosedBy != "bench" {
			t.Errorf("%s: the gap the bench ruled closed reads %+v", s.seat, g[1])
		}
		if g[2] != (DeclaredGapJSON{ID: notAGap, Status: "not_on_board"}) {
			t.Errorf("%s: the id no gap holds reads %+v", s.seat, g[2])
		}
		if len(got[1].Gaps) != 0 || got[1].Gaps == nil {
			t.Errorf("%s: a holding naming no gap carries gaps %#v, want []", s.seat, got[1].Gaps)
		}
	}
}

// A RECORD WITH NO DECLARATION RENDERS `declarations: []`, never null and never absent: "the bench has
// declared nothing" and "not computed" are different answers.
func TestAListWithNoDeclarationSaysNone(t *testing.T) {
	run := corrRun(t)
	sit(t, run, evLens)
	b, err := WorkJSONBytes(run, "lens", evLens)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["declarations"]) != "[]" {
		t.Errorf("declarations renders %q with none on the record, want []", raw["declarations"])
	}
}
