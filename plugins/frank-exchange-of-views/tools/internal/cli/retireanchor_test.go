package cli

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
)

// AN ANCHORED CLAIM LEAVES THROUGH RETIRE, AND ITS ANCHOR LEAVES WITH IT.
//
// An edit may carry an anchor but never drop one, so the only way to delete an anchored sentence
// is to edit it down to the bare anchor and retire it. Before the retire named the anchors that
// exit with the claim, that bare anchor stood forever: an orphan [^N] with a live bibliography
// entry in the assembled report, an empty `- ?` bullet, and a segment the claim counter still
// counted — so the retire credited a claim that had never left the count and cancelled some
// unrelated real loss. These drive the real verbs.

const retireBase = "# Findings\n\nThe sky is blue. The grass is green.\n\n- Water is wet.\n- Fire is hot.\n\nClosing paragraph stays.\n"

// citeSentence cites one sentence under # Findings and returns the tool-assigned label, read off
// the anchor now standing after the sentence's last word.
func citeSentence(t *testing.T, runDir, sentence, url string) string {
	t.Helper()
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat,
		"--quote", `# Findings: "`+sentence+`"`, "--url", url, "--title", "Source "+url); err != nil {
		t.Fatalf("cite %q: %v", sentence, err)
	}
	re := regexp.MustCompile(regexp.QuoteMeta(strings.TrimRight(sentence, ".")) + `<!--cite:(c-[0-9a-f]+)-->`)
	m := re.FindStringSubmatch(readReport(t, runDir))
	if m == nil {
		t.Fatalf("no cite anchor after %q:\n%s", sentence, readReport(t, runDir))
	}
	return m[1]
}

// gutToAnchor edits an anchored sentence down to its bare anchor — the only edit that removes
// the sentence, because an edit may not drop the anchor.
func gutToAnchor(t *testing.T, runDir, sentence, tok string) {
	t.Helper()
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", strings.TrimRight(sentence, ".")+tok+".", "--new", tok, "--reason", "the claim cannot stand"); err != nil {
		t.Fatalf("edit %q down to its anchor: %v", sentence, err)
	}
}

func retireClaim(t *testing.T, runDir, claim string) *recordpb.Retire {
	t.Helper()
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat,
		"--quote", claim, "--reason", "refuted"); err != nil {
		t.Fatalf("retire %q: %v", claim, err)
	}
	return lastBody(t, runDir, &recordpb.Retire{})
}

func citedRun(t *testing.T, urls ...string) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, retireBase)
	registerBlue(t, runDir)
	resp := map[string][]byte{}
	for _, u := range urls {
		resp[u] = []byte("<html>a source at " + u + "</html>")
	}
	withFetcher(t, &fakeFetcher{resp: resp})
	return runDir
}

func TestRetireTakesABareCiteAnchorOutOfTheReport(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	tok := "<!--cite:" + label + "-->"
	before := claimcount.Count(readReport(t, runDir))

	gutToAnchor(t, runDir, "The sky is blue.", tok)
	// THE BUG'S SHAPE, before the retire: the bare anchor still weaves into an orphan footnote.
	if asm := assembled(t, runDir); !strings.Contains(asm, "https://sky/1") {
		t.Fatalf("precondition: the gutted-but-unretired cite should still reach the bibliography:\n%s", asm)
	}
	// And the gutting edit does not send red to re-verify a sentence that no longer exists.
	if ed := lastBody(t, runDir, &recordpb.BlueEdit{}); slices.Contains(ed.GetReopened(), label) {
		t.Errorf("the edit that left %s bare marked it reopened — there is no sentence to re-verify", label)
	}

	ev := retireClaim(t, runDir, "The sky is blue.")
	if got := ev.GetAnchors(); !slices.Equal(got, []string{label}) {
		t.Errorf("retire named anchors %v, want [%s] — the bare anchor must exit with its claim", got, label)
	}
	if ev.GetRemovalBasis() != record.RemovalVerified {
		t.Errorf("removal_basis = %q, want verified", ev.GetRemovalBasis())
	}

	rep := readReport(t, runDir)
	want := "# Findings\n\nThe grass is green.\n\n- Water is wet.\n- Fire is hot.\n\nClosing paragraph stays.\n"
	if rep != want {
		t.Errorf("rendered report after the retire:\n got %q\nwant %q", rep, want)
	}
	if got := claimcount.Count(rep); got != before-1 {
		t.Errorf("claim_count = %d after retiring one of %d cited claims, want %d", got, before, before-1)
	}
	asm := assembled(t, runDir)
	if strings.Contains(asm, "https://sky/1") || strings.Contains(asm, "[^1]") {
		t.Errorf("the assembled report still numbers or lists the withdrawn citation:\n%s", asm)
	}
}

func TestRetireTakesABareFindingMarkerAndItsEmptiedBullet(t *testing.T) {
	runDir := citedRun(t)
	registerLensOnce(t, runDir)
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat,
		"--key", "F1", "--quote", "Water is wet.", "--reason", "overclaim",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatalf("finding: %v", err)
	}
	m := regexp.MustCompile(`Water is wet(<!--fx:f-[0-9a-f]+-->)`).FindStringSubmatch(readReport(t, runDir))
	if m == nil {
		t.Fatalf("no finding marker on the bullet:\n%s", readReport(t, runDir))
	}
	tok := m[1]
	gutToAnchor(t, runDir, "Water is wet.", tok)

	ev := retireClaim(t, runDir, "Water is wet.")
	if len(ev.GetAnchors()) != 1 {
		t.Errorf("retire named %v, want the one bare finding marker", ev.GetAnchors())
	}
	rep := readReport(t, runDir)
	if strings.Contains(rep, tok) {
		t.Errorf("the bare finding marker survived the retire:\n%s", rep)
	}
	if regexp.MustCompile(`(?m)^\s*-\s*[.?!]*\s*$`).MatchString(rep) {
		t.Errorf("the retire left an empty bullet:\n%s", rep)
	}
	if !strings.Contains(rep, "- Fire is hot.\n") {
		t.Errorf("the neighbouring bullet was disturbed:\n%s", rep)
	}
}

// WITHOUT AN EDIT DOWN TO THE ANCHOR, RETIRE IS EXACTLY WHAT IT WAS: it names nothing, and an
// anchor an edit carried on into surviving prose is never taken out.
func TestRetireWithoutABareAnchorNamesNone(t *testing.T) {
	runDir := citedRun(t, "https://grass/1")
	label := citeSentence(t, runDir, "The grass is green.", "https://grass/1")
	tok := "<!--cite:" + label + "-->"
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "The grass is green"+tok+".", "--new", "Grass is verdant"+tok+".", "--reason", "rephrase"); err != nil {
		t.Fatalf("edit: %v", err)
	}
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Fire is hot.", "--new", "Fire is warm.", "--reason", "narrow"); err != nil {
		t.Fatalf("edit: %v", err)
	}
	before := readReport(t, runDir)
	for _, claim := range []string{"The grass is green.", "Fire is hot."} {
		ev := retireClaim(t, runDir, claim)
		if len(ev.GetAnchors()) != 0 {
			t.Errorf("retiring %q named anchors %v — none was left bare", claim, ev.GetAnchors())
		}
	}
	if ev := lastBody(t, runDir, &recordpb.Retire{}); ev.GetRemovalBasis() != record.RemovalVerified {
		t.Errorf("retiring an edited-out claim: removal_basis = %q, want verified, as before", ev.GetRemovalBasis())
	}
	if after := readReport(t, runDir); after != before {
		t.Errorf("a retire that named no anchor changed the report:\n got %q\nwant %q", after, before)
	}
}

// A SECOND RETIRE OF THE SAME CLAIM NAMES NOTHING: the anchor is already gone, and naming it
// again would be a remove op replay cannot apply.
func TestRetiringTwiceNamesTheAnchorOnce(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	gutToAnchor(t, runDir, "The sky is blue.", "<!--cite:"+label+"-->")
	retireClaim(t, runDir, "The sky is blue.")
	if ev := retireClaim(t, runDir, "The sky is blue."); len(ev.GetAnchors()) != 0 {
		t.Errorf("the second retire named %v — the anchor had already left", ev.GetAnchors())
	}
	readReport(t, runDir) // replay must still succeed
}

// THE REFUSAL NAMES THE WAY OUT. An edit dropping an anchor is refused, and the refusal used to
// say citations are removed "with the tool" while naming no verb that did it.
func TestDroppingAnAnchorTeachesRetire(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "The sky is blue<!--cite:"+label+"-->.", "--new", "", "--reason", "cut")
	if err == nil {
		t.Fatal("an edit dropping a citation anchor was accepted")
	}
	if !strings.Contains(err.Error(), "blue retire") {
		t.Errorf("the refusal does not name `blue retire` as the way an anchored claim leaves:\n%v", err)
	}
}

func claimLoss(t *testing.T, runDir string, counts ...int) int {
	t.Helper()
	var results []map[string]any
	for _, c := range counts {
		results = append(results, map[string]any{"claim_count": float64(c)})
	}
	fam, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range scorecard.Compute(runtest.Open(t, runDir), results, &fam)["blue"] {
		if r.Metric == "unrecorded_claim_loss" {
			v, _ := r.Value.(int)
			return v
		}
	}
	t.Fatal("no unrecorded_claim_loss row")
	return 0
}

// THE MEASURED OVER-CREDIT: two retires of gutted cited claims plus one unrelated merge that
// lowers the count with no retire behind it. The retires each took a claim out of the count, so
// they account for two of the three; the merge is the one unaccounted loss. Before, the gutted
// anchors stayed counted, the drop read 1, and the two retires cancelled it to 0.
func TestRetireCreditsOnlyTheClaimsItTookOut(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Gamma holds. Delta holds. Plain prose here.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{
		"https://a/1": []byte("a"), "https://b/1": []byte("b"), "https://g/1": []byte("g"), "https://d/1": []byte("d")}})
	a := citeSentence(t, runDir, "Alpha holds.", "https://a/1")
	b := citeSentence(t, runDir, "Beta holds.", "https://b/1")
	g := citeSentence(t, runDir, "Gamma holds.", "https://g/1")
	d := citeSentence(t, runDir, "Delta holds.", "https://d/1")
	start := claimcount.Count(readReport(t, runDir))

	gutToAnchor(t, runDir, "Alpha holds.", "<!--cite:"+a+"-->")
	retireClaim(t, runDir, "Alpha holds.")
	gutToAnchor(t, runDir, "Beta holds.", "<!--cite:"+b+"-->")
	retireClaim(t, runDir, "Beta holds.")
	// A retire of UNCITED prose removes nothing the count held, and must credit nothing.
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", " Plain prose here.", "--new", "", "--reason", "cut"); err != nil {
		t.Fatalf("edit: %v", err)
	}
	retireClaim(t, runDir, "Plain prose here.")
	// The unrelated merge: two cited sentences become one, both anchors carried, no retire.
	gt, dt := "<!--cite:"+g+"-->", "<!--cite:"+d+"-->"
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Gamma holds"+gt+". Delta holds"+dt, "--new", "Gamma and Delta hold"+gt+dt, "--reason", "merge"); err != nil {
		t.Fatalf("merge edit: %v", err)
	}
	end := claimcount.Count(readReport(t, runDir))
	if start != 4 || end != 1 {
		t.Fatalf("claim_count %d → %d, want 4 → 1 (report %q)", start, end, readReport(t, runDir))
	}
	if got := claimLoss(t, runDir, start, end); got != 1 {
		t.Errorf("unrecorded_claim_loss = %d, want 1 — two retires took two claims out, the merge is the one unaccounted", got)
	}
}

// A GUTTED CITED SENTENCE WITH NO RETIRE IS NOW SEEN: the bare anchor is not a claim, so the count
// falls, and nothing on the record accounts for it.
func TestAGuttedCitedSentenceWithoutRetireReadsAsLoss(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	before := claimcount.Count(readReport(t, runDir))
	gutToAnchor(t, runDir, "The sky is blue.", "<!--cite:"+label+"-->")
	after := claimcount.Count(readReport(t, runDir))
	if after != before-1 {
		t.Fatalf("claim_count %d → %d after gutting a cited sentence, want a fall of 1", before, after)
	}
	if got := claimLoss(t, runDir, before, after); got != 1 {
		t.Errorf("unrecorded_claim_loss = %d, want 1 — a cited sentence left with no retire", got)
	}
}
