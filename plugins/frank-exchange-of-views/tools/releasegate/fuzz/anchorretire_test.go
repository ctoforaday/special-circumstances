package fuzz

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE EVIDENCED-CUT SEED drives both halves of taking an evidenced claim out (cutEvidencedClaim): the
// two rewrites that drop its computation's anchor, or leave it on no words, are refused naming it, and
// the pure cut of the whole sentence takes it. A run not forced drives nothing.
func TestTheEvidencedCutSeedRefusesTheDropsAndCutsTheClaim(t *testing.T) {
	r := stagedLifecycleRunner(t)
	r.cutEvidencedClaim("blue-respond")
	if r.evidencedCutDriven || r.pureCutsRemoved != 0 || r.proseDropRefused != 0 {
		t.Fatalf("a run not forced drove the evidenced cut: driven %v, %d cut(s), %d refusal(s)", r.evidencedCutDriven, r.pureCutsRemoved, r.proseDropRefused)
	}
	r.forceEvidencedCut = true
	r.cutEvidencedClaim("blue-respond")
	if r.evidencedCutErr != "" || r.actErr != "" {
		t.Fatalf("the forced evidenced cut failed: %s %s", r.evidencedCutErr, r.actErr)
	}
	if r.proseDropRefused != 2 || r.pureCutsRemoved != 1 {
		t.Errorf("the forced evidenced cut counted %d refused drop(s) and %d pure cut(s), want 2 and 1", r.proseDropRefused, r.pureCutsRemoved)
	}
	md, err := record.RenderRun(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(md, "Whether the forced claim stands") {
		t.Errorf("the pure cut left the claim in the report:\n%s", md)
	}
}

// BLUE'S DRAWN RETIRE TAKES ONE ANCHOR AND KEEPS ITS SENTENCE (retireOneAnchor): a citation's retire
// lands with the report changed by its token alone, and a retire of an open gap's own marker is refused
// naming the gap (F14), the report unchanged. Each counts its cell.
func TestTheDrawnRetireTakesOneAnchorAndKeepsItsSentence(t *testing.T) {
	r := stagedLifecycleRunner(t)
	const blue, claim = "blue-respond", "A § fuzz sentence to anchor findings"
	cite := r.lifecycleCite(blue, "§ fuzz sentence", "/drawn-retire", "")
	if cite == "" {
		t.Fatalf("blue's cite was refused: %s", r.actErr)
	}
	if !r.openSitting(citationOwner) {
		t.Fatalf("the register of %s was refused", citationOwner)
	}
	g, why := r.lifecycleMint(citationOwner, "document", claim, "", "low")
	if g == "" {
		t.Fatalf("the lens's mint was refused: %s", why)
	}
	before, err := record.RenderRun(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(before, anchor.Token(g)) || !strings.Contains(before, anchor.Token(cite)) {
		t.Fatalf("the report does not hold both anchors:\n%s", before)
	}
	r.retireOneAnchor(blue, g)
	if r.actErr != "" || r.openMarkerRetiresRefused != 1 {
		t.Fatalf("the retire of open gap %s's marker: %d refusal(s) counted (%s), want 1", g, r.openMarkerRetiresRefused, r.actErr)
	}
	if after, _ := record.RenderRun(r.run()); after != before {
		t.Errorf("the refused retire of %s's marker changed the report", g)
	}
	r.retireOneAnchor(blue, cite)
	if r.actErr != "" || r.anchorRetiresKept != 1 {
		t.Fatalf("the retire of citation %s: %d kept-sentence retire(s) counted (%s), want 1", cite, r.anchorRetiresKept, r.actErr)
	}
	if after, _ := record.RenderRun(r.run()); after != strings.Replace(before, anchor.Token(cite), "", 1) {
		t.Errorf("the retire of %s changed the report beyond its token:\n%s", cite, after)
	}
}

// THE BARE-PRESCRIPTION SEED's refusal is read by its words, for EVERY anchor kind (mintBaringCitation):
// a lens prescribing an anchor's token alone as its sentence's replacement is refused naming the token
// left on no words, and the drive counts it per kind rather than reading it as another refusal.
func TestTheBarePrescriptionSeedCountsItsRefusal(t *testing.T) {
	r := stagedLifecycleRunner(t)
	r.forceMintBare = true
	r.citeForBareMint("blue-respond")
	if !r.openSitting(citationOwner) {
		t.Fatalf("the register of %s was refused", citationOwner)
	}
	r.mintBaringCitation(citationOwner)
	if r.mintBareAdmittedErr != "" || r.mintBareOther != "" || r.actErr != "" {
		t.Fatalf("the bare prescription: admitted %q, other %q, act %q; want none", r.mintBareAdmittedErr, r.mintBareOther, r.actErr)
	}
	for _, kind := range anchor.Kinds() {
		if !r.mintBareKinds[kind] {
			t.Errorf("the bare prescription of a %s anchor was not refused naming it", kind)
		}
	}
	if r.mintBareRefused != len(anchor.Kinds()) {
		t.Errorf("the bare prescription: %d refusal(s) counted, want one per kind (%d)", r.mintBareRefused, len(anchor.Kinds()))
	}
}
