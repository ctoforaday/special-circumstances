package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

const keptReason = "the acceptance check still fails: the sentence still names 7 as the only factor"

// keepGrade is the minting lens's regrade that names no grade: it keeps the gap's grade, with why.
func (b *stage) keepGrade(lens, gap, reason string) *stage {
	return b.add(lens, &recordpb.Regrade{GapId: proto.String(gap), Basis: proto.String(reason)})
}

// keptBoard is a lens's open gap G1, minted high, that blue answered with an edit.
func keptBoard(t *testing.T) *stage {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").register(evLens)
	b.mint(evLens, "G1", "high")
	return blueAnswersGap(b, "G1")
}

// THE WRITE PATH ADMITS A REGRADE THAT KEEPS THE GRADE, in both forms — naming no grade, and naming
// the grade the gap already has — and the record reads each as a keep; a regrade that moves a grade
// is not one.
func TestTheWritePathAdmitsARegradeThatKeepsTheGrade(t *testing.T) {
	same := recordpb.Grade_GRADE_HIGH
	lower := recordpb.Grade_GRADE_LOW
	for _, c := range []struct {
		name string
		body *recordpb.Regrade
		keep bool
	}{
		{"naming no grade", &recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String(keptReason)}, true},
		{"naming the grade it has", &recordpb.Regrade{GapId: proto.String("G1"), Severity: &same, Basis: proto.String(keptReason)}, true},
		{"moving a grade", &recordpb.Regrade{GapId: proto.String("G1"), Severity: &lower, Basis: proto.String("conceded")}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := keptBoard(t).register(evLens).seed()
			if _, err := Append(Identity{Run: run, SeatID: evLens}, c.body); err != nil {
				t.Fatalf("the write path refuses the regrade: %v", err)
			}
			m, err := MergedEvents(run)
			if err != nil {
				t.Fatal(err)
			}
			keeps := gradeKeepsOf(m.Events, m.At)
			if got := len(keeps) == 1 && keeps[0].gap == "G1" && keeps[0].reason == c.body.GetBasis(); got != c.keep {
				t.Errorf("read as a keep = %v, want %v: %+v", got, c.keep, keeps)
			}
		})
	}
}

// THE KEEP'S REASON REACHES BLUE: an item on blue's work list names the gap and carries the lens's
// reason, until blue answers the gap again.
func TestBlueReadsTheReasonOfAGradeItsLensKept(t *testing.T) {
	item := func(t *testing.T, b *stage) bool {
		t.Helper()
		s := sittingOfRunT(t, b.register("blue-respond").seed(), "blue", "blue-respond")
		return slices.ContainsFunc(s.Open, func(i Item) bool {
			return strings.Contains(i.What, "gap G1: its lens re-audited it and KEEPS ITS GRADE") && strings.Contains(i.What, keptReason)
		})
	}
	if !item(t, keptBoard(t).sit(2, evLens).keepGrade(evLens, "G1", keptReason)) {
		t.Errorf("blue's work list does not carry the reason its lens kept G1's grade")
	}
	if item(t, keptBoard(t).sit(2, evLens)) {
		t.Errorf("blue's work list carries a kept grade with none on the record")
	}
	if item(t, blueAnswersGap(keptBoard(t).sit(2, evLens).keepGrade(evLens, "G1", keptReason), "G1")) {
		t.Errorf("the kept grade stays on blue's list after blue answered the gap again")
	}
}

// THE KEEP'S REASON REACHES THE BENCH: the debate transcript — its JSON and debate.md — carries it
// as red's argument on the gap, in the epoch it was made.
func TestTheBenchReadsTheReasonOfAGradeALensKept(t *testing.T) {
	run := keptBoard(t).register("red-chair").sit(2, evLens).keepGrade(evLens, "G1", keptReason).seed()
	d, err := debateView.of(run)
	if err != nil {
		t.Fatal(err)
	}
	var keeps []DebateKeepJSON
	for _, e := range d.Epochs {
		keeps = append(keeps, e.RedKeeps...)
	}
	if want := (DebateKeepJSON{GapID: "G1", SeatID: evLens, Text: keptReason}); !slices.Contains(keeps, want) {
		t.Errorf("the debate view's red_keeps = %+v, want %+v", keeps, want)
	}
	// A regrade that moves a grade is not a keep, and is not in the transcript.
	moved := keptBoard(t).register("red-chair").sit(2, evLens).regrade(evLens, "G1", "low").seed()
	if d, err := debateView.of(moved); err != nil {
		t.Fatal(err)
	} else {
		for _, e := range d.Epochs {
			if len(e.RedKeeps) != 0 {
				t.Errorf("a regrade that moved the grade reads as a keep: %+v", e.RedKeeps)
			}
		}
	}
}

// THE LENS IS TOLD WHICH ACT STATES WHAT ITS CHECK STILL FAILS: the item that tells it blue answered
// its gap names the regrade that keeps the grade, and where its statement goes.
func TestTheBlueAnsweredItemNamesTheActThatKeepsTheGrade(t *testing.T) {
	b := keptBoard(t).register(evLens)
	got := blueAnswersT(t, b.evs, []WorkGapState{{ID: "G1", Open: true, lensDue: true}}, evLens)
	if len(got) != 1 || !strings.Contains(got[0], "keep its grade with a regrade that states what your own check still fails") {
		t.Fatalf("the blue-answered item names no act for what the check still fails: %q", got)
	}
}
