package cli

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// ruledSentence is the sentence the ruled quote gap stands on, and nextSentence the paragraph's other.
const (
	ruledSentence = "The index rose by nine points in the spring survey of producers."
	nextSentence  = "Volume recovered later in the year."
)

// quoteImpasseBoard is impasseBoard with its gap quoting ruledSentence, its token placed there.
func quoteImpasseBoard(t *testing.T, runDir string) {
	t.Helper()
	quoteImpasseBoardWith(t, runDir, func(*recordpb.Mint) {})
}

// quoteImpasseBoardWith is quoteImpasseBoard with edit applied to its gap's mint.
func quoteImpasseBoardWith(t *testing.T, runDir string, edit func(*recordpb.Mint)) {
	t.Helper()
	n := 0
	// Each event carries the stamp the write path writes, so the board migrates as a live record does
	// (TestAMigratedRecordWritesEachTombstoneOnce).
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.Stamped(recordtest.At(t, seat, fmt.Sprintf("%s:qboard:%d", seat, n), body),
			fmt.Sprintf("2026-01-01T00:00:%02d.000000000Z", n))
	}
	mint := &recordpb.Mint{Location: proto.String(ruledSentence), GapId: proto.String(boardGap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
	edit(mint)
	evs := []*record.Event{
		at(record.HarnessSeat, &recordpb.Cast{SeatIds: []string{"red-lens-evidence", otherLens, "red-chair", "blue-respond", "judge"}}),
		at(record.HarnessSeat, &recordpb.ReportPart{Text: proto.String("# report\n\n## Findings\n\n" + ruledSentence + " " + nextSentence + "\n")}),
		at("red-chair", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Register{}),
		at("red-lens-evidence", mint),
		at("red-lens-evidence", &recordpb.Anchor{Id: proto.String(boardGap), Location: proto.String(ruledSentence)}),
	}
	// The exchanges impasseBoard writes at k-max 1: two engaged rounds.
	for range 2 {
		evs = append(evs,
			at("red-chair", &recordpb.Register{}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-evidence"), GapIds: []string{boardGap}}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{boardGap}}),
			at("red-lens-evidence", &recordpb.Register{}),
			at("blue-respond", &recordpb.Register{}))
	}
	evs = append(evs, at("red-chair", &recordpb.Register{}))
	recordtest.Seed(t, runDir, evs...)
	writeTerms(t, runDir, 1, 0)
}

// ruleAtImpasse drives runDir's board to the bench's ruling of its impasse docketing — the dispatch
// dockets it, blue states not_a_defect and the lens its argument, the bench sits and rules for blue or,
// remand true, remands it — and returns the docketing.
func ruleAtImpasse(t *testing.T, runDir string, remand bool) string {
	t.Helper()
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	d := lastDocketing(t, runDir, boardGap)
	if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
		t.Fatalf("the docketing = %+v, want impasse", d)
	}
	for _, seat := range []string{"blue-respond", otherLens, "red-lens-evidence"} {
		asSeat(t, runDir, seat, "register")
	}
	asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--as", "not_a_defect", "--reason", "the survey is the source")
	red := []string{"closing", "--id", boardGap, "--reason", "the sentence still misleads"}
	if lastMintAboutSection(t, runDir) {
		asSeat(t, runDir, "red-lens-evidence", red...)
	} else {
		asSeat(t, runDir, "red-lens-evidence", append(red, "--no-text")...)
	}
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	asSeat(t, runDir, "judge", "register", "--occasion", "docket")
	ruling := benchRuling(d, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)
	if remand {
		dr := ruling.GetDocket()
		dr.SidedWith, dr.Final, dr.ReopensOn = nil, nil, nil
		dr.Disposition, dr.Direction = recordpb.Disposition_DISPOSITION_REMANDED.Enum(), proto.String("one more exchange on the survey")
	}
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, ruling); err != nil {
		t.Fatalf("the bench's ruling of %s: %v", d.GetDocketingId(), err)
	}
	return d.GetDocketingId()
}

// tombstoneRows is every Tombstone row on the record, with the type of the act before each.
func tombstoneRows(t *testing.T, runDir string) (rows []*recordpb.Tombstone, after []string) {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	prev := ""
	for _, e := range m.Events {
		if tb, ok := recordpb.BodyAs[*recordpb.Tombstone](e); ok {
			rows, after = append(rows, tb), append(after, prev)
			continue
		}
		typ, _ := recordpb.Body(e)
		prev = fmt.Sprintf("%T", typ)
	}
	return rows, after
}

// mintRefusedNaming runs seat's mint of class at quote, expecting a refusal holding every want.
func mintRefusedNaming(t *testing.T, runDir, what, seat, class, quote string, want []string, extra ...string) {
	t.Helper()
	out, err := tombstoneMint(t, runDir, seat, class, quote, extra...)
	if err == nil {
		t.Errorf("%s was recorded: %s", what, out)
		return
	}
	for _, w := range want {
		if !strings.Contains(err.Error()+out, w) {
			t.Errorf("%s: the refusal does not say %q: %v", what, w, err)
		}
	}
}

// tombstoneMint is seat's mint of class at quote, asked once more where the first is refused only
// because the gap moved under the seat's sitting.
func tombstoneMint(t *testing.T, runDir, seat, class, quote string, extra ...string) (string, error) {
	t.Helper()
	args := append([]string{"--run", runDir, "--seat-id", seat, "--quote", quote, "--class", class, "--problem", "the defect",
		"--fix", "the fix", "--check-kind", "document", "--check", "the acceptance check", "--severity", "medium",
		"--likelihood", "medium", "--impact", "medium", "--complexity", "low"}, extra...)
	out, err := runMint(t, args...)
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runMint(t, args...)
	}
	return out, err
}

// T-B19: A BENCH RULING IS FINAL FOR ITS TEXT (#1695 §III.B12), driven as seats run it, on a quote gap
// ruled for blue at impasse: (1) the ruling's transaction holds, right after the ruling, one
// Tombstone{G, RULING, D} row; (12) the report is byte-identical across it; (9) the work list's
// estopped row and the board's ruling_tombstone name it; (2) another lens's mint of G's class on the
// sentence and G's assignee's own are refused naming the ruling, its docketing
// and what ends it; (3) a proposal on G is refused naming the tombstone; (4) a mint of another class on
// the sentence lands; (6) blue's edit rewording G's sentence writes the end row, and (2) then lands on
// the reworded sentence; on a second run, an edit that cuts the sentence writes no row and (3) is still
// refused; (7) a remand writes no row.
func TestABenchRulingIsFinalForItsText(t *testing.T) {
	const class = "spec_underspecification"
	runDir := newRun(t)
	quoteImpasseBoard(t, runDir)
	before := renders(t, runDir)
	d := ruleAtImpasse(t, runDir, false)

	rows, after := tombstoneRows(t, runDir)
	if len(rows) != 1 || rows[0].GetGapId() != boardGap || rows[0].GetKind() != recordpb.TombstoneKind_TOMBSTONE_KIND_RULING ||
		rows[0].GetDocketingId() != d || rows[0].Ends != nil || after[0] != "*recordpb.MotionRule" {
		t.Fatalf("(1) tombstone rows %v after %v, want one ruling row on %s naming %s, right after the ruling", rows, after, boardGap, d)
	}
	if got := renders(t, runDir); got != before {
		t.Errorf("(12) the ruling changed the report:\nbefore %q\nafter  %q", before, got)
	}
	w, err := record.WorkJSONOfRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Estopped) != 1 || w.Estopped[0].ID != boardGap || w.Estopped[0].Ruling.Docketing != d || !strings.Contains(w.Estopped[0].Ends, "reword") {
		t.Errorf("(9) estopped = %+v, want %s under %s, ended by a reword", w.Estopped, boardGap, d)
	}
	for _, g := range boardOf(t, runDir).Closed {
		if g.ID == boardGap && (g.RulingTombstone == nil || g.RulingTombstone.Docketing != d) {
			t.Errorf("(9) the board's row for %s carries ruling_tombstone %+v, want %s", boardGap, g.RulingTombstone, d)
		}
	}

	names := []string{"the bench ruled gap " + boardGap, "docketing " + d, "tombstone", "A reword of its sentence, not a cut"}
	asSeat(t, runDir, otherLens, "register")
	mintRefusedNaming(t, runDir, "(2) another lens's mint of the ruled class on the sentence", otherLens, class, ruledSentence, names)
	asSeat(t, runDir, "red-lens-evidence", "register")
	mintRefusedNaming(t, runDir, "(2) the assignee's own mint of the ruled class", "red-lens-evidence", class, ruledSentence, names)
	actRefusedNaming(t, runDir, "(3) a proposal on the ruled gap", otherLens, []string{"the bench ruled gap " + boardGap, "docketing " + d, "tombstone"},
		"motion", "docket", "file", "regrade", "--id", boardGap, "--severity", "low", "--reason", "the harm needs two failures")
	if out, err := tombstoneMint(t, runDir, otherLens, otherClass, ruledSentence); err != nil {
		t.Errorf("(4) a mint of another class on the ruled sentence: %v %s", err, out)
	}

	// (6) BLUE'S REWORD ENDS IT: an end row after the edit, nothing standing, and the mint lands.
	const reworded = "The index rose by nine points in the producers' spring survey, revised."
	asSeat(t, runDir, "blue-respond", "register")
	md := renders(t, runDir)
	base := strings.TrimSuffix(ruledSentence, ".")
	at := strings.Index(md, base) + len(base)
	tokens := md[at:anchor.SkipRun(md, at)]
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond", "--key", "RW", "--quote", base+tokens+".",
		"--new", strings.TrimSuffix(reworded, ".")+tokens+".", "--reason", "the survey's own wording"); err != nil {
		t.Fatalf("(6) blue's reword: %v %s", err, out)
	}
	rows, after = tombstoneRows(t, runDir)
	if len(rows) != 2 || !rows[1].GetEnds() || rows[1].GetGapId() != boardGap || after[1] != "*recordpb.BlueEdit" {
		t.Fatalf("(6) tombstone rows %v after %v, want the reword's end row on %s", rows, after, boardGap)
	}
	if w, _ := record.WorkJSONOfRun(runtest.Open(t, runDir)); len(w.Estopped) != 0 {
		t.Errorf("(6) estopped after the reword = %+v, want none", w.Estopped)
	}
	asSeat(t, runDir, otherLens, "register")
	if out, err := tombstoneMint(t, runDir, otherLens, class, reworded); err != nil {
		t.Errorf("(6) the ruled class on the reworded sentence, after the end: %v %s", err, out)
	}

	// (6) A CUT ENDS IT, on a second run: the edit takes the sentence out and leaves the token bare,
	// which lists G in its reopened, a change to G's sentence: the end row is written.
	t.Run("a cut", func(t *testing.T) {
		runDir := newRun(t)
		quoteImpasseBoard(t, runDir)
		d := ruleAtImpasse(t, runDir, false)
		asSeat(t, runDir, "blue-respond", "register")
		cut := strings.TrimSuffix(ruledSentence, ".") + anchor.Token(boardGap) + ". "
		if out, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond", "--key", "CUT", "--quote", cut,
			"--new", anchor.Token(boardGap), "--reason", "the sentence goes"); err != nil {
			t.Fatalf("blue's cut: %v %s", err, out)
		}
		if md := renders(t, runDir); strings.Contains(md, "nine points") || !strings.Contains(md, anchor.Token(boardGap)) {
			t.Fatalf("the cut left %q, want the sentence gone and the token bare", md)
		}
		if rows, _ := tombstoneRows(t, runDir); len(rows) != 2 || rows[0].GetDocketingId() != d || !rows[1].GetEnds() || rows[1].GetGapId() != boardGap {
			t.Errorf("the cut wrote tombstone rows %v, want the ruling's place on %s and its end", rows, boardGap)
		}
	})

	// (16) A SECTION GAP'S RULING STANDS FOR THE RUN: a proposal on it is refused, and a mint of its
	// class about the section lands — a section names no subject a mint can relitigate (L21).
	t.Run("a section gap", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 1, 1)
		d := ruleAtImpasse(t, runDir, false)
		asSeat(t, runDir, otherLens, "register")
		actRefusedNaming(t, runDir, "a proposal on the ruled section gap", otherLens, []string{"docketing " + d, "tombstone", "it stands for the run"},
			"motion", "docket", "file", "regrade", "--id", boardGap, "--severity", "low", "--reason", "the harm needs two failures")
		args := []string{"--run", runDir, "--seat-id", otherLens, "--about-kind", "section", "--about", "Findings", "--class", class,
			"--problem", "the defect", "--fix", "the fix", "--check-kind", "document", "--check", "the acceptance check",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low"}
		out, err := runMint(t, args...)
		if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
			out, err = runMint(t, args...)
		}
		if err != nil {
			t.Errorf("a mint of the ruled class about the section: %v %s", err, out)
		}
	})

	// (7) A REMAND WRITES NO ROW.
	t.Run("a remand", func(t *testing.T) {
		runDir := newRun(t)
		quoteImpasseBoard(t, runDir)
		ruleAtImpasse(t, runDir, true)
		if rows, _ := tombstoneRows(t, runDir); len(rows) != 0 {
			t.Errorf("the remand wrote tombstone rows: %v", rows)
		}
	})

	// (1) THE BODY IS THE TOOL'S, met only by validate's shape checks when handed to Append.
	t.Run("shape", func(t *testing.T) {
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}
		for _, c := range []struct {
			what string
			row  *recordpb.Tombstone
		}{
			{"no gap", &recordpb.Tombstone{Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL.Enum()}},
			{"no kind", &recordpb.Tombstone{GapId: proto.String(boardGap)}},
			{"a docketing on an estoppel row", &recordpb.Tombstone{GapId: proto.String(boardGap), Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL.Enum(), DocketingId: proto.String(d)}},
			{"a docketing on an end row", &recordpb.Tombstone{GapId: proto.String(boardGap), Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING.Enum(), DocketingId: proto.String(d), Ends: proto.Bool(true)}},
		} {
			if _, err := record.Append(id, c.row); err == nil {
				t.Errorf("a row with %s was written", c.what)
			}
		}
	})
}

// lastMintAboutSection is whether the board's gap is about a section, where red's closing takes no text.
func lastMintAboutSection(t *testing.T, runDir string) bool {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_MINT)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetGapId() == boardGap {
			return m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_SECTION
		}
	}
	return false
}

// redRoad is the text red's closing states for ruledSentence on a for-raiser ruling.
const redRoad = "The index rose by four points in the spring survey of producers, on revised weights."

// quoteImpasseBoardFixing is quoteImpasseBoard with its gap's mint prescribing fix.
func quoteImpasseBoardFixing(t *testing.T, runDir, fix string) {
	t.Helper()
	quoteImpasseBoardWith(t, runDir, func(m *recordpb.Mint) { m.FixNew = proto.String(fix) })
}

// ruleForRedAtImpasse drives quoteImpasseBoard's gap to the bench's for-raiser ruling through the verbs:
// the dispatch dockets it at impasse, blue states not_a_defect, the lens states redRoad, the dispatch
// convenes the bench, and the bench rules for red, which applies redRoad. It returns the docketing.
func ruleForRedAtImpasse(t *testing.T, runDir string) string {
	t.Helper()
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	d := lastDocketing(t, runDir, boardGap)
	if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
		t.Fatalf("the docketing = %+v, want impasse", d)
	}
	for _, seat := range []string{"blue-respond", otherLens, "red-lens-evidence"} {
		asSeat(t, runDir, seat, "register")
	}
	asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--as", "not_a_defect", "--reason", "the survey is the source")
	asSeat(t, runDir, "red-lens-evidence", "closing", "--id", boardGap, "--reason", "the sentence still misleads",
		"--quote", ruledSentence, "--new", redRoad)
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	asSeat(t, runDir, "judge", "register", "--occasion", "docket")
	if out, err := answerDocketing(t, runDir, d.GetDocketingId(), "for-raiser", "--final"); err != nil {
		t.Fatalf("--for-raiser: %v\n%s", err, out)
	}
	if md := renders(t, runDir); !strings.Contains(md, strings.TrimSuffix(redRoad, ".")) {
		t.Fatalf("the for-raiser ruling did not apply red's text: %q", md)
	}
	return d.GetDocketingId()
}

// A RULING BARS ITS GAP FOR EITHER SIDE (gblock 2026-11-22: "Estoppel applies to the judges ruling for
// either side. If it sided with blue, red can't create double jeopardy - it lost and can't reraise that
// gap there."), driven through the verbs: after a for-author and after a for-raiser ruling at impasse, the
// ruling's transaction holds G's ruling tombstone and nothing else; red's re-raising of G — a mint of
// G's class on G's sentence, by another lens and by G's assignee, a proposal on
// G — is refused naming the ruling; a mint of another class on the sentence lands (T-B19 (4)). The
// bench's application of red's text is no blue accept: on a for-raiser ruling it places no estoppel row,
// whether G's mint prescribes nothing or prescribes the very text red's closing states.
func TestARulingBarsItsGapForEitherSide(t *testing.T) {
	const class = "spec_underspecification"
	for _, c := range []struct {
		name     string
		sentence string
		board    func(t *testing.T, runDir string)
		rule     func(t *testing.T, runDir string) string
	}{
		{"for blue", ruledSentence, quoteImpasseBoard, func(t *testing.T, runDir string) string { return ruleAtImpasse(t, runDir, false) }},
		{"for red", redRoad, quoteImpasseBoard, ruleForRedAtImpasse},
		{"for red, G's mint prescribing red's text", redRoad, func(t *testing.T, runDir string) { quoteImpasseBoardFixing(t, runDir, redRoad) }, ruleForRedAtImpasse},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := newRun(t)
			c.board(t, runDir)
			d := c.rule(t, runDir)
			rows, after := tombstoneRows(t, runDir)
			if len(rows) != 1 || rows[0].GetKind() != recordpb.TombstoneKind_TOMBSTONE_KIND_RULING || rows[0].GetGapId() != boardGap ||
				rows[0].GetDocketingId() != d || after[0] != "*recordpb.MotionRule" {
				t.Fatalf("tombstone rows %v after %v, want only the ruling's on %s naming %s, right after it", rows, after, boardGap, d)
			}
			names := []string{"the bench ruled gap " + boardGap, "docketing " + d, "tombstone"}
			asSeat(t, runDir, otherLens, "register")
			mintRefusedNaming(t, runDir, "another lens's mint of G's class on its sentence", otherLens, class, c.sentence, names)
			asSeat(t, runDir, "red-lens-evidence", "register")
			mintRefusedNaming(t, runDir, "the assignee's own mint of G's class", "red-lens-evidence", class, c.sentence, names)
			actRefusedNaming(t, runDir, "a proposal on G", otherLens, names,
				"motion", "docket", "file", "regrade", "--id", boardGap, "--severity", "low", "--reason", "the harm needs two failures")
			if out, err := tombstoneMint(t, runDir, otherLens, otherClass, c.sentence); err != nil {
				t.Errorf("(4) a mint of another class on the ruled sentence: %v %s", err, out)
			}
		})
	}
}
