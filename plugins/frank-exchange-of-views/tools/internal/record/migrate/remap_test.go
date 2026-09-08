package migrate

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

func TestArchivedSeatIdsBecomeRoundless(t *testing.T) {
	r := newRemap()
	for old, want := range map[string]string{
		"red-lens-r3-L1":              "red-lens-evidence",
		"red-lens-r3-L2":              "red-lens-evidence", // a second citation instance is the same lens
		"red-lens-r1-L5":              "red-lens-logic",
		"red-lens-r4-L6":              "red-lens-dark-side",
		"red-lens-r2-L7":              "red-lens-voice",
		"red-lens-r2-adversary":       "red-lens-adversary",
		"red-merge-r2":                "red-chair",
		"red-chair-r2":                "red-chair",
		"blue-respond-r4":             "blue-respond",
		"judge-r3":                    "judge",
		"judge-petition-red-merge-r5": "judge-petition-red-chair",
		"judge-terminal":              "judge-terminal",
		"blue-lane-2":                 "blue-lane-2",
		"frontier":                    "frontier",
		"red-lens-evidence":           "red-lens-evidence",
	} {
		got, err := r.seat(old)
		if err != nil || got != want {
			t.Errorf("seat(%q) = (%q, %v), want %q", old, got, err, want)
		}
	}
	if _, err := r.seat("red-lens-r1-L9"); err == nil {
		t.Error("lens number 9 was never on a roster; translating it would be a guess, and a guess is a refusal")
	}
}

func TestGapIdsAndLabelsAreRenumberedInReplayOrderAndReferencesFollow(t *testing.T) {
	r := newRemap()
	// Two lenses' findings: L1 and L2 are both the evidence lens, so their labels share one
	// counter; L5 (logic) has its own.
	f1 := &recordpb.Finding{Label: proto.String("L1-F1")}
	r.apply(f1, "red-lens-evidence")
	f2 := &recordpb.Finding{Label: proto.String("L2-F1")}
	r.apply(f2, "red-lens-evidence")
	f3 := &recordpb.Finding{Label: proto.String("L5-F1")}
	r.apply(f3, "red-lens-logic")
	if f1.GetLabel() != "evidence-F1" || f2.GetLabel() != "evidence-F2" || f3.GetLabel() != "logic-F1" {
		t.Fatalf("labels = %q %q %q — two instances of one lens must not both mint F1", f1.GetLabel(), f2.GetLabel(), f3.GetLabel())
	}
	// Mints in replay order; the third supersedes the first and credits L2's finding.
	m1 := &recordpb.Mint{GapId: proto.String("R1-1")}
	m2 := &recordpb.Mint{GapId: proto.String("R1-2")}
	m3 := &recordpb.Mint{GapId: proto.String("R2-1"), Supersedes: []string{"R1-1"}, FoundBy: []string{"L2-F1"}}
	for _, m := range []*recordpb.Mint{m1, m2, m3} {
		r.apply(m, "red-chair")
	}
	if m1.GetGapId() != "G1" || m2.GetGapId() != "G2" || m3.GetGapId() != "G3" {
		t.Fatalf("gap ids = %q %q %q", m1.GetGapId(), m2.GetGapId(), m3.GetGapId())
	}
	if len(m3.GetSupersedes()) != 1 || m3.GetSupersedes()[0] != "G1" || m3.GetFoundBy()[0] != "evidence-F2" {
		t.Errorf("references did not follow: supersedes=%v found_by=%v", m3.GetSupersedes(), m3.GetFoundBy())
	}
	// Every referencing shape: a flat field, a oneof arm, a list, a blue edit's answers.
	c := &recordpb.Close{GapId: proto.String("R1-2"), Successor: proto.String("R2-1")}
	r.apply(c, "red-chair")
	mo := &recordpb.Motion{Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("R2-1")}}}
	r.apply(mo, "red-chair")
	sc := &recordpb.SpotCheck{Ids: []string{"R1-1", "R1-2"}}
	r.apply(sc, "red-chair")
	be := &recordpb.BlueEdit{Answers: proto.String("R1-2")}
	r.apply(be, "blue-respond")
	if c.GetGapId() != "G2" || c.GetSuccessor() != "G3" || mo.GetDocket().GetGapId() != "G3" ||
		sc.GetIds()[0] != "G1" || sc.GetIds()[1] != "G2" || be.GetAnswers() != "G2" {
		t.Errorf("a reference kept the archived id: close=%s/%s docket=%s ids=%v answers=%s",
			c.GetGapId(), c.GetSuccessor(), mo.GetDocket().GetGapId(), sc.GetIds(), be.GetAnswers())
	}
	// An id nothing minted is left alone: the write path says what a dangling reference is worth.
	d := &recordpb.Close{GapId: proto.String("R9-9")}
	r.apply(d, "red-chair")
	if d.GetGapId() != "R9-9" {
		t.Errorf("an unminted reference was rewritten to %q", d.GetGapId())
	}
	if r.gaps["R1-1"] != "G1" || r.labels["L2-F1"] != "evidence-F2" {
		t.Errorf("the tables the manifest carries are wrong: %v %v", r.gaps, r.labels)
	}
}

func TestConcurrentInstancesAreSerializedAfterTheFirstWithinTheirRound(t *testing.T) {
	ev := func(id int, seat string) OldEvent { return OldEvent{ID: int64(id), SeatID: seat, Word: "verify"} }
	in := []OldEvent{
		ev(1, "red-merge-r3"),
		ev(2, "red-lens-r3-L1"), ev(3, "red-lens-r3-L2"), ev(4, "red-lens-r3-L5"), ev(5, "red-lens-r3-L1"),
		ev(6, "red-lens-r3-L2"), ev(7, "blue-respond-r3"),
		ev(8, "red-lens-r4-L2"), // a round with a second instance and no first stays where it is
	}
	out, moved := serializeInstances(in)
	var got []int64
	for _, e := range out {
		got = append(got, e.ID)
	}
	want := []int64{1, 2, 4, 5, 3, 6, 7, 8}
	if len(got) != len(want) {
		t.Fatalf("lost or grew events: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v (L2's events follow L1's last; L5, blue and the chair do not move)", got, want)
		}
	}
	if moved["red-lens-r3-L2"] != 2 || len(moved) != 1 {
		t.Errorf("moved = %v, want red-lens-r3-L2: 2", moved)
	}
}
