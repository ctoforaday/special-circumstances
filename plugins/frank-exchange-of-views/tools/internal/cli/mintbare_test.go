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

// A PRESCRIPTION THAT WOULD BARE AN ANCHOR IS REFUSED AT THE MINT. Blue's accept refuses an edit that
// leaves an evidence anchor bare, carried or not, or another anchor bare that it does not carry; the
// mint asks the same of the edit the accept would make, so the lens hears it while it can still change
// --new. The evidence anchor is named by its kind.
func TestAMintPrescriptionBaringEvidenceIsRefused(t *testing.T) {
	for _, kind := range anchor.Kinds() {
		runDir := citedRun(t, "https://water/1")
		registerChairOnce(t, runDir)
		registerLensOnce(t, runDir)
		id, _ := markerOfKind(t, runDir, kind, "Water is wet.")
		if !anchor.Backs(id) {
			continue // the rows below
		}
		t.Run(kind, func(t *testing.T) {
			tok := anchor.Token(id)
			for _, new := range []string{tok, tok + "."} {
				mintRefusedBare(t, runDir, "Water is wet", new, "--new would leave "+kind+" "+id, "Keep the sentence in --new")
			}
			// Kept on prose, the anchor is admitted.
			if _, err := mintQuote(t, runDir, "K", "Water is wet", "--new", "Water is damp"+tok+".", "--class", "unsourced_claim"); err != nil {
				t.Errorf("a prescription keeping the %s on prose was refused: %v", kind, err)
			}
		})
	}
}

func TestAMintPrescriptionBaringAnAnchorItLeavesOutIsRefused(t *testing.T) {
	const sentence = "She wrote \"water is wet.\""
	const quote = "She wrote \"water is wet"
	for _, kind := range anchor.Kinds() {
		if kind == "citation" || kind == "computation" {
			continue // evidence, carried or not: the row above
		}
		t.Run(kind, func(t *testing.T) {
			runDir := quotedRun(t)
			id, _ := markerOfKind(t, runDir, kind, sentence)
			tok := anchor.Token(id)
			if rep := readReport(t, runDir); !strings.Contains(rep, sentence+tok) {
				t.Fatalf("the %s anchor does not stand past the closing quotation mark:\n%s", kind, rep)
			}
			mintRefusedBare(t, runDir, quote, ".", "--new would leave anchor "+tok, "--quote leaves it out", "end --quote with "+tok+" as `show report` prints it, and write "+tok+" into --new")

			// Carried as the refusal says: admitted, and the accept applies it.
			carried, err := mintQuote(t, runDir, "C", sentence+tok, "--new", tok+".", "--class", "unsourced_claim")
			if err != nil {
				t.Fatalf("the carry the refusal names was refused: %v", err)
			}
			if err := acceptEdit(t, runDir, "AC", carried, "--accept"); err != nil {
				t.Errorf("the carried prescription was refused at accept: %v", err)
			}
			runDir = quotedRun(t)
			_, _ = markerOfKind(t, runDir, kind, sentence)

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

// A PRESCRIPTION THAT CARRIES THE ANCHOR IS ADMITTED, bare or not: the accept moves a finding or gap
// anchor at the sentence's end with the gap's, and one carried may stand bare until blue retires the
// claim. Evidence never stands bare (the row above).
func TestAMintPrescriptionCarryingTheAnchorIsAdmitted(t *testing.T) {
	for _, kind := range anchor.Kinds() {
		if kind == "citation" || kind == "computation" {
			continue // evidence never stands bare
		}
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			_, _ = markerOfKind(t, runDir, kind, "Water is wet.")
			gap, err := mintQuote(t, runDir, "P", "Water is wet", "--new", ".", "--class", "unsourced_claim")
			if err != nil {
				t.Fatalf("a prescription carrying the %s anchor was refused: %v", kind, err)
			}
			if err := acceptEdit(t, runDir, "A", gap, "--accept"); err != nil {
				t.Errorf("the admitted prescription was refused at accept: %v", err)
			}
		})
	}
}
