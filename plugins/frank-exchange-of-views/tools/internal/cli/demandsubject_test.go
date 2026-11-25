package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// THE DEMAND HOLDS WHILE ITS SUBJECT STANDS IN THE REPORT (ruled by gblock 2026-11-20): a `source` or
// `computation` gap's subject is the anchor it is about, or the claim its marker marks, which leaves
// the report only with that marker. While it stands the close is refused under every disposition;
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

// retireMarkerWithClaim cuts sentence down to gap marker g, then retires the claim with the marker.
func retireMarkerWithClaim(t *testing.T, runDir, sentence, g string) {
	t.Helper()
	gutToAnchor(t, runDir, sentence, anchor.Token(g))
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", g, "--quote", sentence, "--reason", "the claim cannot stand"); err != nil {
		t.Fatalf("retire --anchor %s --quote %q: %v", g, sentence, err)
	}
	if strings.Contains(readReport(t, runDir), anchor.Token(g)) {
		t.Fatalf("the retire left %s's marker in the report:\n%s", g, readReport(t, runDir))
	}
}

// A `--quote` gap's own marker, and a superseded `--quote` ancestor's marker after the heir's mint,
// leaving the report with its claim: the demand is met, and `moot` lands where it was refused.
func TestAQuoteGapClosesOnceItsMarkerLeaves(t *testing.T) {
	const sentence = "Water is wet."
	for _, check := range []string{"source", "computation"} {
		t.Run(check+"/its own marker", func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			g := quoteGap(t, runDir, "quoted", check, sentence)
			wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "moot"), g, check)
			retireMarkerWithClaim(t, runDir, sentence, g)
			if open, answered, awaiting := gapRow(t, runDir, g); !open || !answered || awaiting {
				t.Errorf("after its marker left, %s reads open=%v evidence_answered=%v awaiting_evidence=%v, want open and met — nothing auto-closes", g, open, answered, awaiting)
			}
			if err := lensClose(t, runDir, lensSeat, g, "moot"); err != nil {
				t.Fatalf("the moot close of %s after its marker left with its claim was refused: %v", g, err)
			}
		})
		t.Run(check+"/a superseded ancestor's marker", func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			a := quoteGap(t, runDir, "ancestor", check, sentence)
			h := quoteGap(t, runDir, "heir", check, "Fire is hot.", "--supersedes", a)
			if open, answered, awaiting := gapRow(t, runDir, h); !open || answered || !awaiting {
				t.Fatalf("heir %s reads open=%v evidence_answered=%v awaiting_evidence=%v, want open and awaiting the demand it carries", h, open, answered, awaiting)
			}
			wantDemandRefusal(t, lensClose(t, runDir, lensSeat, h, "moot"), h, check)
			retireMarkerWithClaim(t, runDir, sentence, a)
			if open, answered, awaiting := gapRow(t, runDir, h); !open || !answered || awaiting {
				t.Errorf("after %s's marker left, heir %s reads open=%v evidence_answered=%v awaiting_evidence=%v, want open and met", a, h, open, answered, awaiting)
			}
			if err := lensClose(t, runDir, lensSeat, h, "moot"); err != nil {
				t.Fatalf("the moot close of heir %s after its ancestor's marker left was refused: %v", h, err)
			}
		})
	}
}

// While the subject stands, nothing but the demand's own arms meets it: a `--quote` gap whose marker
// an edit left bare (the claim's words gone, its marker standing until the retire), a gap about an
// anchor that stands, and a gap whose neighbour's claim left the report, are each refused `moot`, and
// the refusal says the demand holds while the subject stands.
func TestTheDemandHoldsWhileItsSubjectStands(t *testing.T) {
	const sentence = "Water is wet."
	for _, c := range []struct{ check, kind string }{{"source", "citation"}, {"computation", "computation"}} {
		t.Run(c.check, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1", "https://sky/1")
			bare := quoteGap(t, runDir, "bare", c.check, sentence)
			id := placeAnswer(t, runDir, c.kind, "The sky is blue.", "https://sky/1", "")
			about := demandGap(t, runDir, "about", c.check, "--about-kind", "anchor", "--about", id)
			neighbour := quoteGap(t, runDir, "neighbour", c.check, "Fire is hot.")

			gutToAnchor(t, runDir, sentence, anchor.Token(bare))
			for _, g := range []string{bare, about, neighbour} {
				wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "moot"), g, c.check, "while its subject stands in the report")
			}
			if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", bare, "--quote", sentence, "--reason", "the claim cannot stand"); err != nil {
				t.Fatalf("retire --anchor %s: %v", bare, err)
			}
			for _, g := range []string{about, neighbour} {
				if _, answered, _ := gapRow(t, runDir, g); answered {
					t.Errorf("the retire of %s's marker met %s's demand, whose subject stands", bare, g)
				}
				wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "moot"), g, c.check)
			}
			if err := lensClose(t, runDir, lensSeat, bare, "moot"); err != nil {
				t.Fatalf("the moot close of %s after its marker left with its claim was refused: %v", bare, err)
			}
		})
	}
}
