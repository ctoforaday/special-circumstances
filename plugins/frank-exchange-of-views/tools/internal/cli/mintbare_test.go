package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// quotedBase holds a sentence ending in a closing quotation mark, where an anchor stands past the mark
// and a gap anchored inside the quotation stands before it: two runs, so the accept's pair does not
// carry the one past the mark.
const quotedBase = "# Findings\n\nShe wrote \"water is wet.\" Fire is hot.\n"

func quotedRun(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, quotedBase)
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://water/1": []byte("<html>water is wet</html>")}})
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	return runDir
}

// mintRefusedBare asserts a mint whose --new would leave an anchor bare is refused with want, and
// records nothing.
func mintRefusedBare(t *testing.T, runDir, quote, new string, want ...string) {
	t.Helper()
	before := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_MINT)
	_, err := mintQuote(t, runDir, "P", quote, "--new", new, "--class", "unsourced_claim")
	if err == nil {
		t.Fatalf("a prescription leaving an anchor bare was verified: --quote %q --new %q", quote, new)
	}
	for _, w := range append(want, "standing bare", "blue's accept would refuse it", "NOTHING IS RECORDED", "state the fix as prose in --fix") {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
	if strings.Contains(err.Error(), "retire --anchor") || strings.Contains(err.Error(), "your quote") || strings.Contains(err.Error(), "--answers") {
		t.Errorf("the refusal names blue's acts, not the lens's:\n%v", err)
	}
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != before {
		t.Errorf("%d mint(s) recorded by a refused mint", n-before)
	}
}

// A PRESCRIPTION THAT WOULD BARE AN ANCHOR IS REFUSED AT THE MINT, of every kind. Blue's accept refuses
// an edit that leaves any anchor with no prose before it in its sentence, carried or not (no writer
// leaves a bare anchor); the mint asks the same of the edit the accept would make, so the lens hears it
// while it can still change --new. The refusal names the anchor's token.
func TestAMintPrescriptionBaringEvidenceIsRefused(t *testing.T) {
	for _, kind := range anchor.Kinds() {
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			id, _ := markerOfKind(t, runDir, kind, "Water is wet.")
			tok := anchor.Token(id)
			for _, new := range []string{tok, tok + "."} {
				mintRefusedBare(t, runDir, "Water is wet", new, tok, "Keep words before")
			}
			// Kept on prose, the anchor is admitted.
			if _, err := mintQuote(t, runDir, "K", "Water is wet", "--new", "Water is damp"+tok+".", "--class", "unsourced_claim"); err != nil {
				t.Errorf("a prescription keeping the %s on prose was refused: %v", kind, err)
			}
		})
	}
}

// AN ANCHOR THE QUOTE LEAVES OUT IS BARED ALL THE SAME: past a closing quotation mark, a token outside
// --quote that --new leaves on no words is refused, named by its token; carried onto no words it is
// refused too; a sentence kept before it is admitted, and the accept applies it.
func TestAMintPrescriptionBaringAnAnchorItLeavesOutIsRefused(t *testing.T) {
	const sentence = "She wrote \"water is wet.\""
	const quote = "She wrote \"water is wet"
	for _, kind := range anchor.Kinds() {
		if kind == "citation" || kind == "computation" {
			continue // evidence: the row above
		}
		t.Run(kind, func(t *testing.T) {
			runDir := quotedRun(t)
			id, _ := markerOfKind(t, runDir, kind, sentence)
			tok := anchor.Token(id)
			if rep := readReport(t, runDir); !strings.Contains(rep, sentence+tok) {
				t.Fatalf("the %s anchor does not stand past the closing quotation mark:\n%s", kind, rep)
			}
			mintRefusedBare(t, runDir, quote, ".", tok, "Keep words before")
			// Carried onto no words: refused as well.
			mintRefusedBare(t, runDir, sentence+tok, tok+".", tok, "Keep words before")

			// A sentence kept before it: admitted, and the accept applies it.
			gap, err := mintQuote(t, runDir, "K", quote, "--new", "She wrote \"water is damp", "--class", "unsourced_claim")
			if err != nil {
				t.Fatalf("a prescription keeping a sentence before the anchor was refused: %v", err)
			}
			if err := acceptEdit(t, runDir, "A", gap, "--accept"); err != nil {
				t.Errorf("the admitted prescription was refused at accept: %v", err)
			}
		})
	}
}

// A PRESCRIPTION CARRYING AN ANCHOR ONTO NO WORDS IS REFUSED: the accept moves a finding or gap marker at
// the sentence's end with the gap's, and a carried marker may not stand bare (P12b: no writer leaves a bare anchor).
func TestAMintPrescriptionCarryingAnAnchorOntoNoWordsIsRefused(t *testing.T) {
	for _, kind := range anchor.Kinds() {
		if kind == "citation" || kind == "computation" {
			continue // evidence the accept does not move: carried onto no words in the first test's rows
		}
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			id, _ := markerOfKind(t, runDir, kind, "Water is wet.")
			mintRefusedBare(t, runDir, "Water is wet", ".", anchor.Token(id), "Keep words before")
		})
	}
}
