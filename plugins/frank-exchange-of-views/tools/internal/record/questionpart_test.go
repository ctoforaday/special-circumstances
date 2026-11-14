package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

const partA = "S-0000000a"

// partProposal is blue's split of the question into a part with the given id.
func partProposal(id string) *recordpb.QuestionPart {
	return &recordpb.QuestionPart{PartId: proto.String(id), Text: proto.String("does the bridge carry the load")}
}

// partCover is blue's coverage record on a part.
func partCover(id string, c recordpb.PartCoverage) *recordpb.QuestionPart {
	return &recordpb.QuestionPart{PartId: proto.String(id), SupersedesCoverage: proto.String("1"), Coverage: c.Enum(), Reason: proto.String("what answers it")}
}

// partCheck is the owner's check of a part.
func partCheck(id string) *recordpb.QuestionPartCheck {
	return &recordpb.QuestionPartCheck{PartId: proto.String(id)}
}

// partGap is a gap about a question part of the given class materiality, graded at sev.
func partGap(gap, part string, cm recordpb.ClassMaterial, sev recordpb.Grade) *recordpb.Mint {
	m := inquiryGap(gap, "", cm, sev)
	m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_QUESTION.Enum(), proto.String(part)
	return m
}

// partBoard is a run with its report ingested, the chair seated, and its question split by hand.
func partBoard(t *testing.T) *stage {
	b := newStage(t).cast(evLens, questionPartOwner, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
	b.unsplit = true
	return b
}

// partKinds is each question blocker on the gate's list, as "kind subject owner".
func partKinds(t *testing.T, run Run) []string {
	t.Helper()
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, b := range bs {
		switch b.Kind {
		case BlockerQuestionUnsplit, BlockerOpenPart, BlockerUncheckedPart, BlockerMaterialGap:
			out = append(out, strings.Join(slices.DeleteFunc([]string{string(b.Kind), b.Subject, b.Owner}, func(s string) bool { return s == "" }), " "))
		}
	}
	return out
}

// partStateOf is a part's judged state on the run, "" where no part has the id.
func partStateOf(t *testing.T, run Run, id string) string {
	t.Helper()
	parts, err := QuestionParts(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range parts {
		if p.ID == id {
			return p.State
		}
	}
	return ""
}

// A QUESTION PART HOLDS ONE JUDGED STATE, read from the question_part_state view, and the gate holds
// exactly what that state owes: unsplit and open are blue's, unchecked the logic lens's, wanting the
// gap's minter's; a check stands only while neither the coverage nor a gap about the part has moved.
func TestAQuestionPartHoldsOneJudgedState(t *testing.T) {
	steps := []struct {
		what  string
		act   func(b *stage)
		state string
		holds []string
	}{
		{"unsplit", func(*stage) {}, "", []string{"question_unsplit blue-respond"}},
		{"split", func(b *stage) { b.add("blue-respond", partProposal(partA)) },
			PartOpen, []string{"open_part " + partA + " blue-respond"}},
		{"covered", func(b *stage) {
			b.add("blue-respond", partCover(partA, recordpb.PartCoverage_PART_COVERAGE_COVERED))
		}, PartUnchecked, []string{"unchecked_part " + partA + " " + questionPartOwner}},
		{"checked", func(b *stage) { b.register(questionPartOwner).add(questionPartOwner, partCheck(partA)) }, PartSound, nil},
		{"coverage recorded again", func(b *stage) {
			b.add("blue-respond", partCover(partA, recordpb.PartCoverage_PART_COVERAGE_NOT_OBTAINED))
		}, PartUnchecked, []string{"unchecked_part " + partA + " " + questionPartOwner}},
		{"checked again", func(b *stage) { b.register(questionPartOwner).add(questionPartOwner, partCheck(partA)) }, PartSound, nil},
		{"a gap below material", func(b *stage) {
			b.add(evLens, partGap("G1", partA, recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, recordpb.Grade_GRADE_LOW))
		}, PartUnchecked, []string{"unchecked_part " + partA + " " + questionPartOwner}},
		{"checked with the gap in hand", func(b *stage) { b.register(questionPartOwner).add(questionPartOwner, partCheck(partA)) }, PartSound, nil},
		{"a material gap", func(b *stage) {
			b.add(evLens, partGap("G2", partA, recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.Grade_GRADE_LOW))
		}, PartWanting, []string{"material_gap G2 " + evLens}},
		{"the material gap closed", func(b *stage) {
			b.add(evLens, &recordpb.Close{GapId: proto.String("G2"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(), Prose: proto.String("p")})
		}, PartUnchecked, []string{"unchecked_part " + partA + " " + questionPartOwner}},
	}
	for n := range steps {
		b := partBoard(t)
		for _, s := range steps[:n+1] {
			s.act(b)
		}
		run := b.seed()
		if got := partStateOf(t, run, partA); got != steps[n].state {
			t.Errorf("after %s: state %q, want %q", steps[n].what, got, steps[n].state)
		}
		if got := partKinds(t, run); !slices.Equal(got, steps[n].holds) {
			t.Errorf("after %s: the gate holds %q, want %q", steps[n].what, got, steps[n].holds)
		}
	}
}

// A RECORD WITH NO CAST OWES NO PART: the question's blockers hold a run whose seats sit, and an
// archived record with no cast derives its outcome as it was stamped (F10).
func TestARecordWithNoCastOwesNoPart(t *testing.T) {
	b := newStage(t)
	b.unsplit = true
	if got := partKinds(t, b.seed()); len(got) != 0 {
		t.Errorf("the gate holds %q on a record with no cast", got)
	}
}

// A QUESTION PART RECORD IS REFUSED WHERE IT IS NOT ONE THING: a split carries its text alone; a
// coverage record names a part that exists, its coverage and its reason, and no text.
func TestAQuestionPartRecordIsOneThing(t *testing.T) {
	covered := recordpb.PartCoverage_PART_COVERAGE_COVERED
	rows := []struct {
		name string
		body *recordpb.QuestionPart
		want string
	}{
		{"a split with no text", &recordpb.QuestionPart{PartId: proto.String("S-0000000b"), Text: proto.String("  ")}, "requires --text"},
		{"a split carrying its coverage", &recordpb.QuestionPart{PartId: proto.String("S-0000000b"), Text: proto.String("t"), Coverage: &covered}, "with its text alone"},
		{"coverage of a part nobody split", partCover("S-0000000b", covered), "S-0000000b"},
		{"coverage with text", &recordpb.QuestionPart{PartId: proto.String(partA), SupersedesCoverage: proto.String("1"), Text: proto.String("t"), Coverage: &covered, Reason: proto.String("r")}, "carries no text"},
		{"coverage with no coverage", &recordpb.QuestionPart{PartId: proto.String(partA), SupersedesCoverage: proto.String("1"), Reason: proto.String("r")}, "requires --as"},
		{"coverage with no reason", &recordpb.QuestionPart{PartId: proto.String(partA), SupersedesCoverage: proto.String("1"), Coverage: &covered, Reason: proto.String(" ")}, "requires --reason"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			b := partBoard(t)
			b.add("blue-respond", partProposal(partA))
			err := validateRun(b.seed(), "blue-respond", recordpb.EventType_EVENT_TYPE_QUESTION_PART, r.body)
			if err == nil || !strings.Contains(err.Error(), r.want) {
				t.Fatalf("%s = %v, want a refusal naming %q", r.name, err, r.want)
			}
		})
	}
}

// THE CHECK IS REFUSED WHERE IT CANNOT JUDGE THE PART AS IT STANDS, each refusal naming what it found.
// Every row is the live write path.
func TestAQuestionPartCheckIsRefusedWhereItCannotJudge(t *testing.T) {
	seat := func(t *testing.T, run Run, id string) Identity {
		i := Identity{Run: run, SeatID: id}
		if _, _, err := RegisterSeat(i, "", ""); err != nil {
			t.Fatal(err)
		}
		return i
	}
	board := func(t *testing.T, cover bool) (Run, Identity, Identity) {
		run := mustRun(t, newRunWithLanes(t, 1))
		seat(t, run, "red-chair")
		recordtest.Seed(t, run.Dir(), recordtest.At(t, "harness", "ingest", &recordpb.ReportPart{Text: proto.String("# report")}))
		blue := seat(t, run, "blue-respond")
		if _, err := Append(blue, partProposal(partA)); err != nil {
			t.Fatal(err)
		}
		if cover {
			if _, err := Append(blue, partCover(partA, recordpb.PartCoverage_PART_COVERAGE_COVERED)); err != nil {
				t.Fatal(err)
			}
		}
		return run, blue, seat(t, run, questionPartOwner)
	}
	rows := []struct {
		name  string
		build func(t *testing.T) Identity
		want  string
	}{
		{"by a seat that is not the owner", func(t *testing.T) Identity {
			run, _, _ := board(t, true)
			return seat(t, run, evLens)
		}, "is " + questionPartOwner + "'s"},
		{"of a part with no coverage", func(t *testing.T) Identity {
			_, _, owner := board(t, false)
			return owner
		}, "has no coverage recorded"},
		{"of a wanting part", func(t *testing.T) Identity {
			run, _, owner := board(t, true)
			m := partGap("G-want", partA, recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.Grade_GRADE_HIGH)
			m.ClassMaterial = nil
			if _, err := Append(seat(t, run, evLens), m); err != nil {
				t.Fatal(err)
			}
			return owner
		}, "is wanting"},
		{"from a sitting opened before its coverage", func(t *testing.T) Identity {
			_, blue, owner := board(t, true)
			if _, err := Append(blue, partCover(partA, recordpb.PartCoverage_PART_COVERAGE_EXCLUDED)); err != nil {
				t.Fatal(err)
			}
			return owner
		}, "after your sitting opened"},
		{"while a check stands", func(t *testing.T) Identity {
			_, _, owner := board(t, true)
			if _, err := Append(owner, partCheck(partA)); err != nil {
				t.Fatal(err)
			}
			return owner
		}, "a check of question part " + partA + " stands"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			id := r.build(t)
			_, err := Append(id, partCheck(partA))
			if err == nil || !strings.Contains(err.Error(), r.want) {
				t.Fatalf("check %s = %v, want a refusal naming %q", r.name, err, r.want)
			}
		})
	}
	// AND THE SAME BOARD ADMITS IT once nothing refuses.
	_, _, owner := board(t, true)
	if _, err := Append(owner, partCheck(partA)); err != nil {
		t.Fatalf("the owner's check of a covered part was refused: %v", err)
	}
	if got := partStateOf(t, owner.Run, partA); got != PartSound {
		t.Errorf("after the owner's check, the part is %q, want %q", got, PartSound)
	}
}

// A PASS OVER A PART NOT COVERED IS QUALIFIED, naming each such part with its coverage; a PASS
// over parts all covered is VERIFIED. Neither is stored: both are derived from the record.
func TestAPassOverAPartNotCoveredIsQualified(t *testing.T) {
	pass := &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}
	for _, c := range []struct {
		coverage recordpb.PartCoverage
		want     string
		why      string
	}{
		{recordpb.PartCoverage_PART_COVERAGE_COVERED, "VERIFIED", ""},
		{recordpb.PartCoverage_PART_COVERAGE_NOT_OBTAINED, "QUALIFIED", partA + " (not_obtained)"},
		{recordpb.PartCoverage_PART_COVERAGE_EXCLUDED, "QUALIFIED", partA + " (excluded)"},
	} {
		t.Run(recordpb.Word(c.coverage), func(t *testing.T) {
			b := partBoard(t)
			b.add("blue-respond", partProposal(partA)).add("blue-respond", partCover(partA, c.coverage)).
				register(questionPartOwner).add(questionPartOwner, partCheck(partA)).add("red-chair", pass)
			got, why, ok, err := DeriveVerdict(b.seed())
			if err != nil || !ok || got != c.want {
				t.Fatalf("got %q (ok=%v, err=%v), want %s: %s", got, ok, err, c.want, why)
			}
			if c.why != "" && !strings.Contains(why, c.why) {
				t.Errorf("the why %q does not name %q", why, c.why)
			}
		})
	}
}

// COVERAGE MISSTATED IS THE QUESTION'S DEFECT: its gap is about the question, a part or the whole.
func TestCoverageMisstatedIsAboutTheQuestion(t *testing.T) {
	misstated := func() *recordpb.Mint {
		m := corrMint("G1")
		m.Class = recordpb.GapClass_GAP_CLASS_COVERAGE_MISSTATED.Enum()
		return m
	}
	b := partBoard(t)
	b.add("blue-respond", partProposal(partA))
	run := b.seed()
	mustRefuse(t, validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_MINT, misstated()), "--class coverage_misstated is the question's defect")
	for _, ref := range []string{partA, ""} {
		m := misstated()
		m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_QUESTION.Enum(), proto.String(ref)
		if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_MINT, m); err != nil {
			t.Errorf("coverage_misstated about question %q refused: %v", ref, err)
		}
	}
}

// THE PARTS ARE COUNTED AS THE RECORD STOOD AT A POINT: a coverage recorded after it is not yet
// there, so a certification's count does not move when blue records one later.
func TestPartCoveragesAreReadAtAPoint(t *testing.T) {
	b := partBoard(t)
	b.add("blue-respond", partProposal(partA)).add("blue-respond", partProposal("S-0000000b"))
	split := b.at()
	b.add("blue-respond", partCover(partA, recordpb.PartCoverage_PART_COVERAGE_EXCLUDED))
	run := b.seed()
	for _, c := range []struct {
		point int64
		want  PartCoverages
	}{
		{0, PartCoverages{}},
		{split, PartCoverages{"": 2}},
		{b.at(), PartCoverages{"": 1, "excluded": 1}},
	} {
		got, err := PartCoveragesAt(run, c.point)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(c.want) {
			t.Errorf("at %d: %v, want %v", c.point, got, c.want)
		}
		for k, n := range c.want {
			if got[k] != n {
				t.Errorf("at %d: %v, want %v", c.point, got, c.want)
			}
		}
	}
}
