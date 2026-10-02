package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
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
	t.Run("under Migrating an unanswered match lands: an archived mint is what a seat DID", func(t *testing.T) {
		dir := withG1(t)
		m, seat := mint("red-lens-computation", "G2", ordinal, nil)
		Migrating = true
		defer func() { Migrating = false }()
		if err := appendAs(t, dir, seat, m); err != nil {
			t.Fatalf("migration re-judged a pre-screen mint under the duplicate screen: %v", err)
		}
		Migrating = false
		// The same mint, live, is still asked — the exemption is migration's, not the screen's.
		m2, seat2 := mint("red-lens-evidence", "G3", ordinal, nil)
		if err := appendAs(t, dir, seat2, m2); err == nil || !strings.Contains(err.Error(), "--distinct-from") {
			t.Fatalf("the live screen let the same defect through after a migrated mint: %v", err)
		}
	})
	t.Run("under Migrating a mint carrying --distinct-from is checked for the reference only: the match is not re-asked", func(t *testing.T) {
		// A replayed mint with a distinction reads the board (the reference must exist) and then
		// leaves: an archived mint that matched a second gap it never named is what the seat DID.
		// Delete the gate after the reference check and this lands on the lexical screen.
		dir := withG1(t)
		other, seat := mint("red-lens-voice", "G2", "The closing section narrates how the run searched instead of addressing the subject.", nil)
		other.Location = proto.String("This run searched three databases.")
		if err := appendAs(t, dir, seat, other); err != nil {
			t.Fatal(err)
		}
		m, seat := mint("red-lens-computation", "G3", "The closing section narrates how the run searched rather than addressing the subject.", func(m *recordpb.Mint) { m.DistinctFrom = []string{"G1"} })
		m.Location = proto.String("This run searched three databases.")
		Migrating = true
		defer func() { Migrating = false }()
		if err := appendAs(t, dir, seat, m); err != nil {
			t.Fatalf("migration re-asked the lexical match of a mint that answered for another gap: %v", err)
		}
		Migrating = false
		// Live, the same mint is asked about G2 — the fixture matches, and only replay is exempt.
		m2, seat2 := mint("red-lens-evidence", "G4", "The closing section narrates how the run searched rather than addressing the subject.", func(m *recordpb.Mint) { m.DistinctFrom = []string{"G1"} })
		m2.Location = proto.String("This run searched three databases.")
		if err := appendAs(t, dir, seat2, m2); err == nil || !strings.Contains(err.Error(), "G2 (open") {
			t.Fatalf("the live screen did not match G2, so the replayed mint above proved nothing: %v", err)
		}
	})
	t.Run("under Migrating a distinction from a gap that does not exist is still refused: that is structure", func(t *testing.T) {
		dir := withG1(t)
		m, seat := mint("red-lens-computation", "G2", ordinal, func(m *recordpb.Mint) { m.DistinctFrom = []string{"G9"} })
		Migrating = true
		defer func() { Migrating = false }()
		if err := appendAs(t, dir, seat, m); err == nil || !strings.Contains(err.Error(), "G9") {
			t.Fatalf("migration admitted a --distinct-from naming no gap: %v", err)
		}
	})
	t.Run("a board that cannot be read refuses the mint with that error, never admits it", func(t *testing.T) {
		// A read failure and an empty board must not share an answer: the screen swallowed every
		// FamilyOf error as "no board yet", and on that path the --distinct-from check went unasked
		// too. The failure is staged in the record itself — a docket ruling whose filing is not on
		// the record, which FamilyOf refuses to fold — so what is tested is the screen's answer to
		// the read, not a stand-in for it.
		dir := withG1(t)
		recordtest.Seed(t, dir, recordtest.At(t, "judge", "judge:motion_rule:M9", &recordpb.MotionRule{
			MotionId: proto.String("M9"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String("ruled"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Principle: proto.String("p"),
				Tension: proto.String("t"), ReviewFlag: proto.String("r"), Settled: proto.String("s"), Final: proto.Bool(true)}},
		}))
		if _, err := FamilyOf(mustRun(t, dir)); err == nil {
			t.Fatal("the staged record still reads — the fixture proves nothing")
		}
		distinct := func(m *recordpb.Mint) { m.DistinctFrom = []string{"G1"} }
		// Under Migrating a mint with a distinction still reads the board for the reference, so the
		// read failure is its refusal too.
		for _, tc := range []struct {
			name      string
			migrating bool
			answer    func(*recordpb.Mint)
		}{{"live, unanswered", false, nil}, {"live, --distinct-from", false, distinct}, {"migrating, --distinct-from", true, distinct}} {
			Migrating = tc.migrating
			m, seat := mint("red-lens-computation", "G2", ordinal, tc.answer)
			err := appendAs(t, dir, seat, m)
			Migrating = false
			if err == nil {
				t.Fatalf("%s: a mint was admitted over a board that could not be read", tc.name)
			}
			// The refusal is the MINT's, coded, and names the read as its cause — not the fold's
			// own sentence about a docket ruling, which two other readers emit word for word.
			if !strings.HasPrefix(err.Error(), "record: mint refused — the board could not be read, so the screen cannot run: ") {
				t.Fatalf("%s: the refusal is not the mint's own, naming the read: %v", tc.name, err)
			}
			if feov.CodeOf(err) != string(feov.Conflict) {
				t.Errorf("%s: the refusal carries code %q, not %q", tc.name, feov.CodeOf(err), feov.Conflict)
			}
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
