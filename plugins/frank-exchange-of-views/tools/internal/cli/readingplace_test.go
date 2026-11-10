package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// THE REPEATED SENTENCE: one sentence's words in two sections, each copy cited (C-1 on the first, C-2
// on the second), and a third sentence whose citation C-3 backs a clause in its middle. The copies are
// cited while their words differ, and blue's edit then makes them one sentence twice, as an edit can.
const (
	rpRepeated = "The quarter closed with costs rising sharply."
	rpClause   = "Margins held, by the survey's count, through a cold winter."
	rpReport   = "# One\n\n" + rpRepeated + " Other text stands here.\n\n# Two\n\nThe quarter closed with costs rising steeply. More text follows here.\n\n# Three\n\n" + rpClause + "\n"
)

func repeatedRun(t *testing.T) readingBoard {
	t.Helper()
	b := readingRunOf(t, rpReport, rpRepeated, "The quarter closed with costs rising steeply.", "Margins held, by the survey's count")
	if _, err := run(t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--key", "E0", "--quote", "costs rising steeply",
		"--new", "costs rising sharply", "--reason", "one wording"); err != nil {
		t.Fatalf("blue's edit: %v", err)
	}
	// Every lens sits again, so its readings come from a sitting opened after the edit.
	for _, s := range []string{lensL, lensM, lensN, lensP, lensK} {
		b.sit(s)
	}
	return b
}

// section is the text of the report section headed h.
func section(report, h string) string {
	_, after, ok := strings.Cut(report, "# "+h+"\n")
	if !ok {
		return ""
	}
	if i := strings.Index(after, "\n# "); i >= 0 {
		return after[:i]
	}
	return after
}

// replays fails the test unless the record's report ops replay from empty to the written render.
func replays(t *testing.T, runDir string) {
	t.Helper()
	ops, err := record.ReportProjection(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	got, err := reportproj.Replay("", false, ops)
	if err != nil {
		t.Fatalf("the record does not replay: %v", err)
	}
	if want := readReport(t, runDir); got != want {
		t.Errorf("the replay differs from the write-time render:\n%s\n---\n%s", got, want)
	}
}

// 25. A citation reading places its gap by the anchor's identity, never by the sentence's words (F9).
func TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor(t *testing.T) {
	cc := "claim_contradicts_source"
	var g string
	b := repeatedRun(t)
	c1, c2, c3 := b.cs[0], b.cs[1], b.cs[2]
	t.Run("a: the second copy's citation", func(t *testing.T) {
		r := b.mustRead(lensL, citeReading(c2, "refutes", cc)...)
		g = mintedBy(t, b, lensL, r, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, bySentenceOf(c2))
		rep := readReport(t, b.runDir)
		if strings.Contains(section(rep, "One"), anchor.Token(g)) || !strings.Contains(section(rep, "Two"), "rising sharply"+anchor.Token(g)+anchor.Token(c2)+".") {
			t.Errorf("G is not at the end of the second copy alone:\n%s", rep)
		}
		og := openGap(t, b.runDir, g)
		if !strings.Contains(og.Passage, "More text follows") {
			t.Errorf("the board puts G in passage %q, want the second section's", og.Passage)
		}
	})
	t.Run("b: a second lens names it", func(t *testing.T) {
		r := b.mustRead(lensM, citeReading(c2, "refutes", cc)...)
		if r.Gap == nil || r.Gap.ID != g || r.Gap.Minted {
			t.Errorf("a second lens's reading of C-2 = %+v, want %s named", r.Gap, g)
		}
	})
	t.Run("c: the first copy's citation mints its own", func(t *testing.T) {
		r := b.mustRead(lensN, citeReading(c1, "refutes", cc)...)
		if r.Gap == nil || !r.Gap.Minted || r.Gap.ID == g {
			t.Errorf("a reading of C-1 = %+v, want a gap of its own", r.Gap)
		}
	})
	t.Run("d: a clause-cited sentence takes its gap at the sentence's end", func(t *testing.T) {
		r := b.mustRead(lensP, citeReading(c3, "refutes", cc)...)
		g3 := r.Gap.ID
		want := "through a cold winter" + anchor.Token(g3) + "."
		if rep := readReport(t, b.runDir); !strings.Contains(rep, want) {
			t.Errorf("the clause-cited sentence's gap is not at its end (%q):\n%s", want, rep)
		}
		// A plain mint quoting the whole sentence lands at the same place.
		p := repeatedRun(t)
		pg, err := p.plainMint(lensP, rpClause, cc)
		if err != nil {
			t.Fatal(err)
		}
		if rep := readReport(t, p.runDir); !strings.Contains(rep, "through a cold winter"+anchor.Token(pg)+".") {
			t.Errorf("the plain mint's gap lands elsewhere:\n%s", rep)
		}
	})
	t.Run("e: a verified fix over the repeated sentence is accepted at the second copy", func(t *testing.T) {
		b := repeatedRun(t)
		c2 := b.cs[1]
		newText := "The quarter closed with costs rising modestly" + anchor.Token(c2) + "."
		r := b.mustRead(lensL, citeReading(c2, "refutes", cc, "--new", newText)...)
		g := r.Gap.ID
		if og := openGap(t, b.runDir, g); og.FixBasis != "verified" {
			t.Errorf("fix_basis %q, want verified", og.FixBasis)
		}
		old, nw := boardPair(t, b.runDir, g)
		first := section(readReport(t, b.runDir), "One")
		if err := acceptEdit(t, b.runDir, "A", g, "--accept"); err != nil {
			t.Fatalf("the accept: %v", err)
		}
		rep := readReport(t, b.runDir)
		if section(rep, "One") != first {
			t.Errorf("the accept moved the first copy:\n%s", rep)
		}
		if !strings.Contains(section(rep, "Two"), "rising modestly") {
			t.Errorf("the accept did not apply at the second copy:\n%s", rep)
		}
		e := lastBody(t, b.runDir, &recordpb.BlueEdit{})
		if !e.GetExactSpan() {
			t.Error("the accept of a repeated span is not recorded with exact_span")
		}
		if e.GetOld() != old || e.GetNew() != nw {
			t.Errorf("the accept applied %q → %q, the board serves %q → %q", e.GetOld(), e.GetNew(), old, nw)
		}
		replays(t, b.runDir)
	})
	t.Run("f: a verified fix on the clause-cited sentence moves its gap to the end", func(t *testing.T) {
		b := repeatedRun(t)
		c3 := b.cs[2]
		newText := "Margins held, by the survey's count" + anchor.Token(c3) + ", through a mild winter."
		g := b.mustRead(lensL, citeReading(c3, "refutes", cc, "--new", newText)...).Gap.ID
		if err := acceptEdit(t, b.runDir, "A", g, "--accept"); err != nil {
			t.Fatalf("the accept: %v", err)
		}
		if rep := readReport(t, b.runDir); !strings.Contains(rep, "through a mild winter"+anchor.Token(g)+".") {
			t.Errorf("G's token is not at the end of --new:\n%s", rep)
		}
		replays(t, b.runDir)
	})
	t.Run("g: a fix the accept would refuse names the reading's subject", func(t *testing.T) {
		b := repeatedRun(t)
		_, err := b.verify(lensL, citeReading(b.cs[1], "refutes", cc, "--new", "Costs were flat.")...)
		mustRefuse(t, err, "", "the sentence your --anchor's token sits in")
		if err != nil && strings.Contains(err.Error(), "--quote") {
			t.Errorf("a citation reading's refusal names --quote: %v", err)
		}
		if c := b.counts(); c != [3]int{} {
			t.Errorf("the refusal wrote verify/mint/anchor %v", c)
		}
	})
	t.Run("h: a found-source reading of the repeated words is refused as ambiguous", func(t *testing.T) {
		b := repeatedRun(t)
		_, err := b.verify(lensL, urlReading(rdU3, rpRepeated, "refutes", cc)...)
		mustRefuse(t, err, "", "more than once", `"One"`, `"Two"`)
		if c := b.counts(); c != [3]int{} {
			t.Errorf("the refusal wrote verify/mint/anchor %v", c)
		}
	})
	t.Run("i: the record replays", func(t *testing.T) {
		replays(t, b.runDir)
	})
}

// 26. A reading over an applied prescription mints its gap: the estoppel binds a seat re-raising a
// prescription by its text, and a reading names a citation or reads a source (F10).
func TestAReadingOverAnAppliedPrescriptionMintsItsGap(t *testing.T) {
	cc := "claim_contradicts_source"
	applied := func(t *testing.T) (readingBoard, string, string) {
		b := readingRun(t)
		prime := "Costs rose modestly in the first quarter" + anchor.Token(b.c1) + "."
		g0, err := b.plainMint(lensP, rdCited, "spec_underspecification", "--new", prime)
		if err != nil {
			t.Fatalf("P's prescription: %v", err)
		}
		if err := acceptEdit(t, b.runDir, "A", g0, "--accept"); err != nil {
			t.Fatalf("blue's accept: %v", err)
		}
		const primeText = "Costs rose modestly in the first quarter."
		_, err = b.plainMint(lensL, primeText, cc)
		mustRefuse(t, err, "", g0)
		return b, g0, primeText
	}
	estoppels := func(t *testing.T, runDir string) int {
		m, err := record.MergedEvents(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		return record.EstoppelRejectionsOf(m.Events)
	}
	t.Run("a, b: a citation reading", func(t *testing.T) {
		b, _, _ := applied(t)
		if n := estoppels(t, b.runDir); n != 1 {
			t.Fatalf("the control logged %d estoppel entries, want 1", n)
		}
		b.sit(lensL)
		r := b.mustRead(lensL, citeReading(b.c1, "refutes", cc)...)
		g1 := mintedBy(t, b, lensL, r, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, bySentenceOf(b.c1))
		if n := estoppels(t, b.runDir); n != 1 {
			t.Errorf("the reading logged an estoppel: %d entries", n)
		}
		fs := failedOf(t, b.runDir, b.c1)
		if len(fs) != 1 || fs[0].CheckState() != record.CheckFoundAgainst || fs[0].AnsweredBy() != g1 {
			t.Errorf("C-1's failed checks = %+v, want one found_against answered by %s", fs, g1)
		}
		// C-2 is read, so the PASS is held by C-1's reading and its gap alone.
		b.mustRead(lensK, citeReading(b.c2, "supports", "")...)
		if _, err := run(t, "register", "--run", b.runDir, "--seat-id", "red-chair"); err != nil {
			t.Fatal(err)
		}
		_, err := run(t, "verdict", "--run", b.runDir, "--seat-id", "red-chair", "--as", "PASS")
		mustRefuse(t, err, "", b.c1)
	})
	t.Run("c: a found-source reading", func(t *testing.T) {
		b, _, primeText := applied(t)
		b.sit(lensM)
		r := b.mustRead(lensM, urlReading(rdU4, primeText, "refutes", cc)...)
		g2 := mintedBy(t, b, lensM, r, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, byLocation(primeText))
		if n := estoppels(t, b.runDir); n != 1 {
			t.Errorf("the reading logged an estoppel: %d entries", n)
		}
		if og := openGap(t, b.runDir, g2); !og.Open {
			t.Errorf("%s is not open", g2)
		}
		// Both citations read from a sitting after the accept, so G2 is what holds the PASS.
		b.sit(lensK)
		b.mustRead(lensK, citeReading(b.c1, "supports", "")...)
		b.mustRead(lensK, citeReading(b.c2, "supports", "")...)
		if _, err := run(t, "register", "--run", b.runDir, "--seat-id", "red-chair"); err != nil {
			t.Fatal(err)
		}
		_, err := run(t, "verdict", "--run", b.runDir, "--seat-id", "red-chair", "--as", "PASS")
		mustRefuse(t, err, "", g2, "material gap")
	})
}
