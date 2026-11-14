package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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

// 25. A reading of a citation in a repeated sentence is keyed on the anchor, never on the sentence's
// words: the gap minted about it is about the citation, and a plain mint quoting a clause-cited
// sentence lands at its end.
func TestAReadingOfARepeatedSentenceIsKeyedOnTheAnchor(t *testing.T) {
	cc := "claim_contradicts_source"
	b := repeatedRun(t)
	c1, c2 := b.cs[0], b.cs[1]
	t.Run("a: the second copy's citation", func(t *testing.T) {
		b.mustRead(lensL, citeReading(c2, "refutes")...)
		g, out, err := b.mintSource(lensL, cc, about(c2)...)
		if err != nil {
			t.Fatalf("the gap about C-2: %v\n%s", err, out)
		}
		mintedAbout(t, b.runDir, g, c2)
		if rep := readReport(t, b.runDir); strings.Contains(rep, anchor.Token(g)) {
			t.Errorf("a gap about a citation placed a token:\n%s", rep)
		}
	})
	t.Run("c: the first copy's citation is its own subject", func(t *testing.T) {
		b.mustRead(lensN, citeReading(c1, "refutes")...)
		if _, out, err := b.mintSource(lensN, cc, about(c1)...); err != nil {
			t.Fatalf("a gap about C-1 beside the open gap about C-2: %v\n%s", err, out)
		}
	})
	t.Run("d: a plain mint quoting a clause-cited sentence lands at its end", func(t *testing.T) {
		p := repeatedRun(t)
		pg, err := p.plainMint(lensP, rpClause, "spec_underspecification")
		if err != nil {
			t.Fatal(err)
		}
		if rep := readReport(t, p.runDir); !strings.Contains(rep, "through a cold winter"+anchor.Token(pg)+".") {
			t.Errorf("the plain mint's gap lands elsewhere:\n%s", rep)
		}
	})
	t.Run("h: a found-source reading of the repeated words is refused as ambiguous", func(t *testing.T) {
		b := repeatedRun(t)
		_, err := b.verify(lensL, urlReading(rdU3, rpRepeated, "refutes")...)
		mustRefuse(t, err, "", "more than once", `"One"`, `"Two"`)
		if c := b.counts(); c != [3]int{} {
			t.Errorf("the refusal wrote verify/mint/anchor %v", c)
		}
	})
	t.Run("i: the record replays", func(t *testing.T) {
		replays(t, b.runDir)
	})
}

// 26. A READING OVER AN APPLIED PRESCRIPTION RECORDS, AND THE GAP ON IT MEETS THE ESTOPPEL GUARD LIKE
// ANY GAP (F4 (a), ruled): a cited source's gap names the citation and no quote, so the guard never
// asks of it; a found source's claim_contradicts_source is a quote mint, refused naming the
// prescribing gap and admitted with --supersedes it.
func TestAReadingOverAnAppliedPrescriptionRecordsAndItsGapMeetsTheGuard(t *testing.T) {
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
	t.Run("a, b: a citation reading and the gap about the citation", func(t *testing.T) {
		b, _, _ := applied(t)
		if n := estoppels(t, b.runDir); n != 1 {
			t.Fatalf("the control logged %d estoppel entries, want 1", n)
		}
		b.sit(lensL)
		b.mustRead(lensL, citeReading(b.c1, "refutes")...)
		g1, out, err := b.mintSource(lensL, cc, about(b.c1)...)
		if err != nil {
			t.Fatalf("the gap about the citation on the applied sentence: %v\n%s", err, out)
		}
		if n := estoppels(t, b.runDir); n != 1 {
			t.Errorf("a gap about a citation met the estoppel guard: %d entries", n)
		}
		// C-2 is read, so the PASS is held by G1 alone.
		b.mustRead(lensK, citeReading(b.c2, "supports")...)
		if _, err := run(t, "register", "--run", b.runDir, "--seat-id", "red-chair"); err != nil {
			t.Fatal(err)
		}
		_, err = run(t, "verdict", "--run", b.runDir, "--seat-id", "red-chair", "--as", "PASS")
		mustRefuse(t, err, "", g1, "material gap")
	})
	t.Run("c: a found-source reading and its quote gap", func(t *testing.T) {
		b, g0, primeText := applied(t)
		// The prescribing lens reads: a supersession is the minter's own act.
		b.sit(lensP)
		b.mustRead(lensP, urlReading(rdU4, primeText, "refutes")...)
		_, out, err := b.mintSource(lensP, cc, "--quote", primeText)
		mustRefuse(t, err, out, "estoppel", g0)
		if n := estoppels(t, b.runDir); n != 2 {
			t.Errorf("the found source's quote mint logged no estoppel: %d entries", n)
		}
		g2, out, err := b.mintSource(lensP, cc, "--quote", primeText, "--supersedes", g0)
		if err != nil {
			t.Fatalf("the quote mint superseding the prescribing gap: %v\n%s", err, out)
		}
		if og := openGap(t, b.runDir, g2); !og.Open {
			t.Errorf("%s is not open", g2)
		}
	})
}
