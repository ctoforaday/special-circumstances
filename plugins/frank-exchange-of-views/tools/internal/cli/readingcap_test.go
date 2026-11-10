package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
)

// 35g. A found-source reading at the cap is held by a standing gap of its class on the sentence, and
// the same reading after that gap closed mints, so the cap refuses it: the hold is the board's at the
// act, and a gap another seat closes in between takes it away.
func TestAFoundSourceReadingAtTheCapIsRefusedOnceItsHoldingGapCloses(t *testing.T) {
	b := readingRunOf(t, rdFillerReport, 1, rdCited, rdBacked)
	const cc = "claim_contradicts_source"
	// M's gap on the sentence is the one its own found-source reading mints, a `document` gap so
	// that its close asks no live-answer demand.
	g := b.mustRead(lensM, "--url", rdU5, "--title", "Found", "--quote", rdFree1, "--as", "refutes", "--confidence", "high",
		"--reason", "read at the leaf", "--class", cc, "--problem", "the source says otherwise than the sentence",
		"--check-kind", "document", "--check", "the sentence states what the source says",
		"--severity", "high", "--likelihood", "high", "--impact", "medium").Gap.ID
	b.fillTo(lensL, 0)
	r := b.mustRead(lensL, urlReading(rdU3, rdFree1, "refutes", cc)...)
	if r.Gap == nil || r.Gap.ID != g || r.Gap.Minted {
		t.Fatalf("at the cap, the first reading = %+v, want %s named", r.Gap, g)
	}
	if _, err := run(t, "close", "--run", b.runDir, "--seat-id", lensM, "--id", g, "--as", "repaired",
		"--verified-by", lensM, "--verified-with", "read", "--verified-against", "the report",
		"--reason", "the sentence now says what the source says"); err != nil {
		t.Fatalf("M's close: %v", err)
	}
	held := b.counts()
	_, err := b.verify(lensL, urlReading(rdU4, rdFree1, "refutes", cc)...)
	mustRefuse(t, err, "", "mint budget admits no more")
	if c := b.counts(); c != held {
		t.Errorf("the refused reading wrote %v over %v", c, held)
	}
}

// 35h. A verified fix of a sentence many citations back keeps their tokens, as the accept requires,
// and the tokens are not words it adds: the authoring bound measures the words.
func TestAVerifiedFixOfAHeavilyBackedSentenceIsMeasuredByItsWords(t *testing.T) {
	const sentence = "Costs rose sharply in the first quarter."
	b := readingRunOf(t, rdReport, 0, sentence, sentence, sentence)
	rep := readReport(t, b.runDir)
	start := strings.Index(rep, "Costs rose")
	end := strings.Index(rep[start:], ".") + start + 1
	marked := rep[start:end]
	if n := strings.Count(marked, "<!--cite:"); n != 3 {
		t.Fatalf("the fixture's sentence carries %d citations, want 3: %q", n, marked)
	}
	// 70 characters of words: inside the bound, and past it only with the tokens counted.
	newText := "Read again, with the measured figures for every month of the quarter, " + marked
	if _, err := b.verifyJSON(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source", "--new", newText)...); err != nil {
		t.Fatalf("a %d-character rewording of a sentence whose tokens it keeps was refused: %v",
			len([]rune(claimcount.StripAnchors(newText)))-len([]rune(sentence)), err)
	}
}
