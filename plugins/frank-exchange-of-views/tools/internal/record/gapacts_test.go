package record

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY ACT ON A GAP IS READ, AND NOTHING ELSE IS (plans/findings-fold.md §III.B3, T-B45). Each row
// of acts is one act a party records on a gap, naming the gap in its own field; gapActsOf must place
// it on that gap. Each row of turns changes what a gap is about without naming it: the raiser's turn
// (lensTurnActsOf), entered under the subject's author, and NO act on the gap — no answer, no exchange
// spent. A row deleted from either fold fails here. The non-acts at the end name no gap, or a gap's
// id only as prose, and a proposal is neither an act nor a turn.
func TestEveryActOnAGapIsRead(t *testing.T) {
	const lens, blue = evLens, "blue-respond"
	const lens2 = "red-lens-logic"
	mint := func(gap string, about recordpb.AboutKind, ref string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(gap), AboutKind: about.Enum(), AboutRef: proto.String(ref)}
	}
	proposal := func(q string) *recordpb.Inquiry {
		return &recordpb.Inquiry{InquiryId: proto.String(q), Question: proto.String("q"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}
	}
	fate := func(q string, s recordpb.InquiryStatus, reason, answers string) *recordpb.Inquiry {
		f := &recordpb.Inquiry{InquiryId: proto.String(q), SupersedesStatus: proto.String("1"), Status: s.Enum(), Reason: proto.String(reason)}
		if answers != "" {
			f.Answers = proto.String(answers)
		}
		return f
	}
	part := func(id string, cov recordpb.PartCoverage, reason, answers string) *recordpb.QuestionPart {
		p := &recordpb.QuestionPart{PartId: proto.String(id), SupersedesCoverage: proto.String("1"), Coverage: cov.Enum(), Reason: proto.String(reason)}
		if answers != "" {
			p.Answers = proto.String(answers)
		}
		return p
	}
	split := func(id, answers string) *recordpb.QuestionPart {
		p := &recordpb.QuestionPart{PartId: proto.String(id), Text: proto.String("a part")}
		if answers != "" {
			p.Answers = proto.String(answers)
		}
		return p
	}
	base := []struct {
		seat string
		body proto.Message
	}{
		{blue, proposal("I1")},
		{blue, proposal("I2")},
		{blue, split("P1", "")},
		{lens, mint("G1", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")},
		{lens, mint("G2", recordpb.AboutKind_ABOUT_KIND_INQUIRY, "I1")},
		// Sections the tool composes from fates, named as a seat might type them: the key folds case
		// and space, as it does for every reader of a heading.
		{lens, mint("G4", recordpb.AboutKind_ABOUT_KIND_SECTION, "research  Areas")},
		{lens, mint("G5", recordpb.AboutKind_ABOUT_KIND_SECTION, "Future research directions")},
		{lens, mint("G7", recordpb.AboutKind_ABOUT_KIND_QUESTION, "P1")},
		{lens, mint("G8", recordpb.AboutKind_ABOUT_KIND_QUESTION, "")},
		// A gap about G1, raised by another lens: G1's minter is its author.
		{lens2, mint("G6", recordpb.AboutKind_ABOUT_KIND_GAP, "G1")},
	}
	type row struct {
		name, seat, gap string
		body            proto.Message
	}
	acts := []row{
		{"a regrade", lens, "G1", &recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String("b")}},
		{"a close", lens, "G1", &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum()}},
		{"a closing", blue, "G1", &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("t")}},
		{"an edit answering it", blue, "G1", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("a")}},
		{"a computation answering it", blue, "G1", &recordpb.Computation{Argv: []string{"python3", "script.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), ComputationId: proto.String("P1"), Answers: proto.String("G1")}},
		{"a judgement of a computation answering it", lens, "G1", &recordpb.Reproduce{ComputationId: proto.String("P1")}},
		{"a citation placed as its answer", blue, "G1", &recordpb.Cite{Label: proto.String("C-1"), Answers: proto.String("G1")}},
		{"a reading of a citation answering it", lens, "G1", &recordpb.Verify{Anchor: proto.String("C-1")}},
		{"a fate naming the gap about its inquiry", blue, "G2", fate("I1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r", "G2")},
		{"a fate naming the gap about a section it prints", blue, "G4", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "x", "G4")},
		{"a fate repeated, naming the gap", blue, "G4", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "x", "G4")},
		{"a coverage naming the gap about its part", blue, "G7", part("P1", recordpb.PartCoverage_PART_COVERAGE_COVERED, "c", "G7")},
		{"a split naming the gap about the question", blue, "G8", split("P2", "G8")},
		// THE MINTER'S REGRADE NAMING A GAP ABOUT ITS GAP is its act on its gap and its answer on the
		// other (Regrade.answers).
		{"a regrade of the gap it is about, naming it", lens, "G6", &recordpb.Regrade{GapId: proto.String("G1"), Answers: proto.String("G6"), Basis: proto.String("b")}},
		{"an amend of the gap it is about, naming it", lens, "G6", &recordpb.Amend{GapId: proto.String("G1"), Answers: proto.String("G6"), Problem: proto.String("p2"), Basis: proto.String("b")}},
		{"the bench's amend of the gap it is about, naming it", "bench", "G6", &recordpb.Amend{GapId: proto.String("G1"), Answers: proto.String("G6"), Problem: proto.String("p6"), Basis: proto.String("b")}},
		{"an amend of its own gap", lens, "G1", &recordpb.Amend{GapId: proto.String("G1"), Problem: proto.String("p3"), Basis: proto.String("b")}},
	}
	turns := []row{
		{"a fate changed on the inquiry it is anchored on", blue, "G2", fate("I1", recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED, "r2", "")},
		{"a fate changed that the section it is anchored on prints", blue, "G4", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "x, narrowed", "")},
		{"a fate moving its inquiry out of the section it is anchored on", blue, "G4", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, "d", "")},
		{"a reason changed that the section it is anchored on prints", blue, "G5", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, "d, narrowed", "")},
		{"a coverage changed on its part", blue, "G7", part("P1", recordpb.PartCoverage_PART_COVERAGE_EXCLUDED, "e", "")},
		{"a part split out of the question", blue, "G8", split("P3", "")},
		// THE BENCH'S APPLICATION IS A CHANGE OF THE SUBJECT: the raiser's turn on every other gap on
		// it, entered under the subject's author. It names G9, the gap its ruling applies.
		{"the bench's edit changing the gap's sentence", "bench", "G1", &recordpb.BlueEdit{Answers: proto.String("G9"), Old: proto.String("a"), New: proto.String("b"), Reopened: []string{"G1"}}},
		// A CHANGE TO THE GAP IT IS ABOUT THAT NAMES IT NOT is its raiser's turn, entered under its
		// author — that gap's minter — whoever wrote it: an amend, a regrade, a close, and the bench's
		// application.
		{"an amend of the gap it is about, naming nothing", lens, "G6", &recordpb.Amend{GapId: proto.String("G1"), Problem: proto.String("p4"), Basis: proto.String("b")}},
		{"a regrade of the gap it is about, naming nothing", lens, "G6", &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordpb.Grade_GRADE_LOW.Enum(), Basis: proto.String("b")}},
		{"the bench's amend of the gap it is about, naming another", "bench", "G6", &recordpb.Amend{GapId: proto.String("G1"), Answers: proto.String("G9"), Problem: proto.String("p5"), Basis: proto.String("b")}},
		{"a close of the gap it is about", lens, "G6", &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum()}},
	}
	var evs []*Event
	var ids []int64
	at := func(seat string, body proto.Message) int64 {
		evs = append(evs, recordtest.At(t, seat, "", body))
		ids = append(ids, int64(len(evs)))
		return ids[len(ids)-1]
	}
	for _, b := range base {
		at(b.seat, b.body)
	}
	actAt := make([]int64, len(acts))
	for i, r := range acts {
		actAt[i] = at(r.seat, r.body)
	}
	turnAt := make([]int64, len(turns))
	for i, r := range turns {
		turnAt[i] = at(r.seat, r.body)
	}
	unrelated := []int64{
		at(lens, &recordpb.Log{Text: proto.String("G1 is wrong")}),
		at(blue, &recordpb.BlueEdit{Old: proto.String("G1"), New: proto.String("b")}),
		at(lens, &recordpb.Reproduce{ComputationId: proto.String("K-unknown")}),
		// A READING MOVES NO TURN ON A GAP: only its citation's answer does.
		at(blue, &recordpb.Cite{Label: proto.String("C-2")}),
		at(lens, &recordpb.Verify{Anchor: proto.String("C-2")}),
		// A fate repeating its inquiry's previous record word for word, naming nothing.
		at(blue, fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, "d, narrowed", "")),
		// A PROPOSAL IS NO ACT AND NO TURN: the raiser answers it in its window.
		at(blue, &recordpb.Motion{MotionId: proto.String("M1"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}),
		at(blue, &recordpb.Motion{MotionId: proto.String("M2"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
			Act: &recordpb.Motion_Regrade{Regrade: &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordpb.Grade_GRADE_LOW.Enum(), Basis: proto.String("b")}}}),
	}
	// No retire is here, so no whole index is needed: a reaching retire is read off the record's rows
	// (TestAReachingRetireIsTheLensTurnOnTheGap).
	got := gapActsOf(evs, ids)
	turned := lensTurnActsOf(evs, ids, WindowIndex{}, got, nil, nil)
	has := func(as []gapAct, at int64, seat string) bool {
		return slices.ContainsFunc(as, func(a gapAct) bool { return a.at == at && a.seat == seat })
	}
	for i, r := range acts {
		if !has(got[r.gap], actAt[i], r.seat) {
			t.Errorf("%s by %s is not read as an act on %s: %+v", r.name, r.seat, r.gap, got[r.gap])
		}
	}
	if !has(got["G1"], actAt[slices.IndexFunc(acts, func(r row) bool { return r.name == "a regrade of the gap it is about, naming it" })], lens) {
		t.Errorf("the minter's regrade naming G6 is not read as its act on its own gap G1: %+v", got["G1"])
	}
	// THE ACT THAT NAMES A GAP ANSWERS IT AND IS NO TURN ON IT: the bench's amend naming G6 is no turn
	// of G6's raiser's entered under G6's author.
	for i, r := range acts {
		if r.seat == "bench" && has(turned[r.gap], actAt[i], lens) {
			t.Errorf("%s is read as %s's raiser's turn as well as its answer: %+v", r.name, r.gap, turned[r.gap])
		}
	}
	// A FATE THAT MOVES ITS INQUIRY CHANGES BOTH SECTIONS: the one it leaves and the one it enters.
	for i, r := range turns {
		if r.name == "a fate moving its inquiry out of the section it is anchored on" && !has(turned["G5"], turnAt[i], blue) {
			t.Errorf("a fate moving I2 into Future research directions is not read as G5's raiser's turn: %+v", turned["G5"])
		}
	}
	for i, r := range turns {
		author := blue
		if r.gap == "G6" {
			author = lens // G6 is about G1, whose minter is its author
		}
		if !has(turned[r.gap], turnAt[i], author) {
			t.Errorf("%s by %s is not read as %s's raiser's turn under its author %s: %+v", r.name, r.seat, r.gap, author, turned[r.gap])
		}
		for g, as := range got {
			// the act on G1 itself is G1's: an amend, a regrade or a close of the gap G6 is about
			if g != "G9" && g != actedOnGap(r.body) && slices.ContainsFunc(as, func(a gapAct) bool { return a.at == turnAt[i] }) {
				t.Errorf("%s names no gap and is read as an act on %s — a change to the subject is the raiser's turn, never the author's answer", r.name, g)
			}
		}
	}
	// A FATE IS NO ACT ON A GAP ABOUT TEXT IT DOES NOT CHANGE: G1 is about a section blue writes.
	for _, as := range [][]gapAct{got["G1"], turned["G1"]} {
		for _, a := range as {
			if _, fate := recordpb.BodyAs[*recordpb.Inquiry](evs[a.at-1]); fate {
				t.Errorf("a fate (event %d) is read as an act or a turn on G1, about a section no fate prints", a.at)
			}
		}
	}
	for _, m := range []map[string][]gapAct{got, turned} {
		for g, as := range m {
			for _, a := range as {
				if slices.Contains(unrelated, a.at) {
					t.Errorf("event %d is neither an act nor a change of a subject and is read on %s", a.at, g)
				}
			}
		}
	}
	if len(got["G3"]) != 0 {
		t.Errorf("a gap's own mint is read as an act on it: %+v", got["G3"])
	}
}
