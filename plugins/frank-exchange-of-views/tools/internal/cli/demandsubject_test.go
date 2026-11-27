package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// THE DEMAND HOLDS WHILE ITS SUBJECT STANDS IN THE REPORT (ruled by gblock 2026-11-20): a `source` or
// `computation` gap's subject is the anchor it is about, or the claim its marker marks, which leaves
// the report only with that marker — by a pure cut of its sentence, which takes the marker. While it stands the close is refused under every disposition;
// once it has left, nothing more is owed and the lens's `moot` lands.

// quoteGap mints gap key on sentence, by its quote, with check kind kind, from lensSeat.
func quoteGap(t *testing.T, runDir, key, kind, sentence string, extra ...string) string {
	t.Helper()
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	out, err := runMint(t, append([]string{"--run", runDir, "--seat-id", lensSeat, "--class", "spec_underspecification", "--problem", "p", "--fix", "f",
		"--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--key", key, "--check-kind", kind, "--quote", sentence}, extra...)...)
	if err != nil {
		t.Fatalf("mint %s --quote %q --check-kind %s: %v", key, sentence, kind, err)
	}
	return gapID(out)
}

// cutMarkerWithClaim cuts sentence whole, which takes gap marker g with it (removed_anchors).
func cutMarkerWithClaim(t *testing.T, runDir, sentence, g string) {
	t.Helper()
	quote := strings.TrimSuffix(sentence, ".") + anchor.Token(g) + "."
	if _, err := cut(t, runDir, quote); err != nil {
		t.Fatalf("the cut of %q: %v", quote, err)
	}
	if strings.Contains(readReport(t, runDir), anchor.Token(g)) {
		t.Fatalf("the cut left %s's marker in the report:\n%s", g, readReport(t, runDir))
	}
}

// A `--quote` gap's own marker leaving the report with its claim by a pure cut: the demand is met, and on the lens's turn the cut
// hands it, `moot` lands
// where it was refused.
func TestAQuoteGapClosesOnceItsMarkerLeaves(t *testing.T) {
	const sentence = "Water is wet."
	for _, check := range []string{"source", "computation"} {
		t.Run(check+"/its own marker", func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			g := quoteGap(t, runDir, "quoted", check, sentence)
			wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "moot"), g, check)
			cutMarkerWithClaim(t, runDir, sentence, g)
			lensRereads(t, runDir) // the cut is the raiser's turn: its next sitting is handed the cut
			if open, answered, awaiting := gapRow(t, runDir, g); !open || !answered || awaiting {
				t.Errorf("after its marker left, %s reads open=%v evidence_answered=%v awaiting_evidence=%v, want open and met — nothing auto-closes", g, open, answered, awaiting)
			}
			if err := lensClose(t, runDir, lensSeat, g, "moot"); err != nil {
				t.Fatalf("the moot close of %s after its marker left with its claim was refused: %v", g, err)
			}
		})

	}
}

// While the subject stands, nothing but the demand's own arms meets it: a `--quote` gap whose claim
// stands, a gap about an anchor that stands, and a gap whose neighbour's claim left the report, are each
// refused `moot`, and the refusal says the demand holds while the subject stands. The cut of the first
// gap's claim meets its demand alone.
func TestTheDemandHoldsWhileItsSubjectStands(t *testing.T) {
	const sentence = "Water is wet."
	for _, c := range []struct{ check, kind string }{{"source", "citation"}, {"computation", "computation"}} {
		t.Run(c.check, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1", "https://sky/1")
			cutG := quoteGap(t, runDir, "cut", c.check, sentence)
			id := placeAnswer(t, runDir, c.kind, "The sky is blue.", "https://sky/1", "")
			about := demandGap(t, runDir, "about", c.check, "--about-kind", "anchor", "--about", id)
			neighbour := quoteGap(t, runDir, "neighbour", c.check, "Fire is hot.")

			for _, g := range []string{cutG, about, neighbour} {
				wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "moot"), g, c.check, "while its subject stands in the report")
			}
			cutMarkerWithClaim(t, runDir, sentence, cutG)
			lensRereads(t, runDir) // the cut is the raiser's turn: its next sitting is handed the cut
			for _, g := range []string{about, neighbour} {
				if _, answered, _ := gapRow(t, runDir, g); answered {
					t.Errorf("the cut of %s's marker met %s's demand, whose subject stands", cutG, g)
				}
				wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "moot"), g, c.check)
			}
			if err := lensClose(t, runDir, lensSeat, cutG, "moot"); err != nil {
				t.Fatalf("the moot close of %s after its marker left with its claim was refused: %v", cutG, err)
			}
		})
	}
}
