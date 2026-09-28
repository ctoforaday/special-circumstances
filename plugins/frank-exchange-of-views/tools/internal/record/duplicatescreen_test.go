package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE MINT SCREENS THE BOARD AT THE WRITE, and a match is answered rather than overridden.
//
// universe-m13: red-lens-dark-side minted G7 and red-lens-computation minted G8 on one semiprime
// ordinal defect, ten seconds apart. The second lens had run the separate screen four seconds
// BEFORE G7 existed. Each arm below is one way that answer can go, and one way the screen could be
// wrong.
func TestAMintMatchingAnOpenGapIsAskedWhichItIs(t *testing.T) {
	const ordinal = "91 is stated as the twenty-seventh semiprime and the second of the form 7·q; both ordinals are wrong."
	mint := func(seat, gap, problem string, answer func(*recordpb.Mint)) (*recordpb.Mint, string) {
		m := &recordpb.Mint{
			GapId: proto.String(gap), Class: proto.String("c"),
			Problem:     proto.String(problem),
			Location:    proto.String("91 is the twenty-seventh semiprime."),
			RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("chk"),
			CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}
		if answer != nil {
			answer(m)
		}
		return m, seat
	}
	withG1 := func(t *testing.T) string {
		dir := newRun(t)
		m, seat := mint("red-lens-dark-side", "G1", ordinal, nil)
		if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: seat}, m); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	appendAs := func(t *testing.T, dir, seat string, m *recordpb.Mint) error {
		_, err := Append(Identity{Run: mustRun(t, dir), SeatID: seat}, m)
		return err
	}

	t.Run("the same defect from another lens is refused, naming the gap and who minted it", func(t *testing.T) {
		dir := withG1(t)
		m, seat := mint("red-lens-computation", "G2", "The ordinal position of 91 among semiprimes is stated wrongly: twenty-seventh and second of the form 7·q.", nil)
		err := appendAs(t, dir, seat, m)
		if err == nil {
			t.Fatal("a mint of a defect already open on the board landed unasked")
		}
		for _, want := range []string{"G1", "red-lens-dark-side", "--distinct-from", "finding about that gap"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the refusal does not name %q: %v", want, err)
			}
		}
	})
	t.Run("--distinct-from answers it, and the claim is on the record", func(t *testing.T) {
		dir := withG1(t)
		m, seat := mint("red-lens-computation", "G2", ordinal, func(m *recordpb.Mint) { m.DistinctFrom = []string{"G1"} })
		if err := appendAs(t, dir, seat, m); err != nil {
			t.Fatalf("a mint that answered the screen was refused: %v", err)
		}
		f, err := FamilyOf(mustRun(t, dir))
		if err != nil {
			t.Fatal(err)
		}
		if g := f.Gap("G2"); g == nil || len(g.Mint.GetDistinctFrom()) != 1 || g.Mint.GetDistinctFrom()[0] != "G1" {
			t.Errorf("the distinction is not on the record: %+v", g)
		}
	})
	t.Run("--supersedes answers it too: the new gap replaces the old", func(t *testing.T) {
		dir := withG1(t)
		m, seat := mint("red-lens-computation", "G2", ordinal, func(m *recordpb.Mint) { m.Supersedes = []string{"G1"} })
		if err := appendAs(t, dir, seat, m); err != nil {
			t.Fatalf("a successor naming its ancestor was refused: %v", err)
		}
	})
	t.Run("a distinction from a gap that does not exist is refused", func(t *testing.T) {
		dir := withG1(t)
		m, seat := mint("red-lens-computation", "G2", ordinal, func(m *recordpb.Mint) { m.DistinctFrom = []string{"G1", "G9"} })
		if err := appendAs(t, dir, seat, m); err == nil || !strings.Contains(err.Error(), "G9") {
			t.Fatalf("a --distinct-from naming no gap was accepted: %v", err)
		}
	})
	t.Run("an unrelated defect is not asked", func(t *testing.T) {
		dir := withG1(t)
		m, seat := mint("red-lens-voice", "G2", "The closing section narrates how the run searched instead of addressing the subject.", nil)
		m.Location = proto.String("This run searched three databases.")
		if err := appendAs(t, dir, seat, m); err != nil {
			t.Fatalf("an unrelated mint was refused: %v", err)
		}
	})
	t.Run("a CLOSED match does not block: the screen is of what is open", func(t *testing.T) {
		dir := withG1(t)
		if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "red-lens-dark-side"}, &recordpb.Close{
			GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
			AnchorSeat: proto.String("red-lens-dark-side"), AnchorTool: proto.String("show report"), AnchorTarget: proto.String("the ordinal sentence"),
			Prose: proto.String("verified at the leaf"),
		}); err != nil {
			t.Fatal(err)
		}
		m, seat := mint("red-lens-computation", "G2", ordinal, nil)
		if err := appendAs(t, dir, seat, m); err != nil {
			t.Fatalf("a closed gap blocked a fresh mint: %v", err)
		}
	})
}
