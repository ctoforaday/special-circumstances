package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// chairBoard is a board at head 2 whose one cast lens has sat twice and found nothing, so it is
// retired at the head, the lens condition of the PASS gate holds, and each test adds only the
// condition it is about. The gaps are minted by outsideLens — a seat not in the cast — so no mint
// makes the cast lens active again.
func chairBoard(t *testing.T, lensSat bool) *stage {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
	if lensSat {
		b.sit(2, evLens).sit(2, evLens)
	}
	return b
}

const outsideLens = "red-lens-logic"

func cmMint(gap string, cm recordpb.ClassMaterial, severity string) *recordpb.Mint {
	sev, _ := GradeOf(severity)
	return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(gap), Class: recordtest.ClassFor(cm).Enum(), Problem: proto.String("p"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		ClassMaterial: cm.Enum()}
}

// openChairSitting opens the chair's sitting after the parties sat and closes its log channel, so
// neither the owed register nor the log holds `complete` false for a reason unrelated to the gate.
func openChairSitting(b *stage) *stage {
	b.register("red-chair")
	return b.add("red-chair", &recordpb.Log{Text: proto.String("nothing blocked"),
		Type: recordpb.LogType_LOG_TYPE_REQUEST.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()})
}

var passGate = &recordpb.Gate{}

func blockingItems(s SittingJSON) []string {
	var out []string
	for _, it := range s.Open {
		if it.Blocks {
			out = append(out, it.What)
		}
	}
	return out
}

func itemNaming(s SittingJSON, sub string) (Item, bool) {
	for _, it := range s.Open {
		if strings.Contains(it.What, sub) {
			return it, true
		}
	}
	return Item{}, false
}

// An open gap of a `never` class, graded high, does not hold PASS: the chair's list names it with
// blocks:false, the PASS appends, and the sitting is complete.
func TestChairWorkListOverOpenNeverClassGap(t *testing.T) {
	b := chairBoard(t, true)
	b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
	run := openChairSitting(b).seed()
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("a PASS over an open never-class gap was refused: %v", err)
	}
	s := sittingOfRunT(t, run, "chair", "red-chair")
	it, ok := itemNaming(s, "gap G1 is open and not material (its class is never material)")
	if !ok || it.Blocks {
		t.Errorf("the never-class gap must be listed, not blocking: %+v in %v", it, s.Open)
	}
	if !s.Complete {
		t.Errorf("sitting.complete is false after the PASS the gate admitted: %v", blockingItems(s))
	}
}

// `complete` IS THE GATE'S ANSWER. For each class default either side of the floor, and for a
// superseded ancestor, the sitting is complete exactly when the PASS appends, and each open row's
// `material` is the view's.
func TestChairWorkListAgreesWithPassGate(t *testing.T) {
	for _, c := range []struct {
		name         string
		build        func(b *stage)
		wantMaterial bool
		wantAppends  bool
	}{
		{"always-low", func(b *stage) { b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, "low")) }, true, false},
		{"never-high", func(b *stage) { b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high")) }, false, true},
		{"by_grade-low", func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
		}, false, true},
		{"by_grade-medium", func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "medium"))
		}, true, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := chairBoard(t, true)
			c.build(b)
			run := openChairSitting(b).seed()
			_, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate)
			appended := err == nil
			if appended != c.wantAppends {
				t.Fatalf("PASS appended=%v, want %v (%v)", appended, c.wantAppends, err)
			}
			s := sittingOfRunT(t, run, "chair", "red-chair")
			if s.Complete != appended {
				t.Errorf("sitting.complete=%v but the PASS appended=%v — the list and the gate disagree: %v", s.Complete, appended, blockingItems(s))
			}
			for _, g := range mustWorkJSONT(t, run).Open {
				if g.ID == "G1" && g.Material != c.wantMaterial {
					t.Errorf("open[G1].material=%v, want %v", g.Material, c.wantMaterial)
				}
			}
		})
	}
}
