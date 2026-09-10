package cli

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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

// A FINDING MARKER IS RED'S: blue's retire takes it out only once red's lifecycle has closed on
// it. While no gap credits the finding, or a crediting gap is open, the marker stays bare and the
// retire says why; after the gap closes, the retire takes it and its emptied bullet, and reading
// the report at that anchor names the retire rather than calling the id stale.
func TestRetireTakesARedMarkerOnlyOnceItsGapIsClosed(t *testing.T) {
	runDir := citedRun(t)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat,
		"--key", "F1", "--quote", "Water is wet.", "--reason", "overclaim",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatalf("finding: %v", err)
	}
	f := lastBody(t, runDir, &recordpb.Finding{})
	id, label := f.GetFindingId(), f.GetLabel()
	tok := "<!--fx:" + id + "-->"
	if !strings.Contains(readReport(t, runDir), "Water is wet"+tok) {
		t.Fatalf("no finding marker on the bullet:\n%s", readReport(t, runDir))
	}
	gutToAnchor(t, runDir, "Water is wet.", tok)

	retireHeld := func(stage, why string) {
		t.Helper()
		out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "Water is wet.", "--reason", "refuted")
		if err != nil {
			t.Fatalf("%s: retire: %v", stage, err)
		}
		if ev := lastBody(t, runDir, &recordpb.Retire{}); len(ev.GetAnchors()) != 0 {
			t.Errorf("%s: blue's retire took red's marker out: named %v", stage, ev.GetAnchors())
		}
		if !strings.Contains(readReport(t, runDir), tok) {
			t.Errorf("%s: the finding marker left the report", stage)
		}
		if !strings.Contains(out, id) || !strings.Contains(out, why) {
			t.Errorf("%s: the retire does not say it kept %s and why (%q):\n%s", stage, id, why, out)
		}
	}
	retireHeld("no gap credits the finding", "no gap credits")

	out, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "water", "--class", "overclaim", "--problem", "the defect", "--fix", "the fix",
		"--check-kind", "document", "--check", "the acceptance check", "--severity", "medium",
		"--likelihood", "medium", "--impact", "medium", "--complexity", "low", "--found-by", label)
	if err != nil {
		t.Fatalf("mint crediting %s: %v", label, err)
	}
	gap := gapID(out)
	retireHeld("an open gap credits the finding", "gap "+gap+" is open")

	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat,
		"--id", gap, "--as", "repaired", "--verified-by", "L1", "--verified-with", "go test",
		"--verified-against", "./internal/x", "--reason", "the check passes"); err != nil {
		t.Fatalf("close: %v", err)
	}
	ev := retireClaim(t, runDir, "Water is wet.")
	if !slices.Equal(ev.GetAnchors(), []string{id}) {
		t.Fatalf("with its gap closed, retire named %v, want [%s]", ev.GetAnchors(), id)
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

	_, err = run(t, "show", "report", "--seat-id", "blue-respond", "--run", runDir, "--anchor", id)
	if err == nil {
		t.Fatal("a window was read around an anchor that left the report")
	}
	if msg := err.Error(); !strings.Contains(msg, "RETIRED") || !regexp.MustCompile(`at event \d+`).MatchString(msg) || strings.Contains(msg, "stale reference or belongs") {
		t.Errorf("reading at a retired anchor must name the retire event, not call it stale:\n%s", msg)
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

// THE MEASURED OVER-CREDIT: two retires of gutted cited claims, a retire of uncited prose, one
// cited sentence gutted with no retire, and a merge of two cited sentences into one. The retires
// each took a citation out of the count, so they account for two of the three that left; the
// unretired gut is the one unaccounted loss. The merge carries both anchors and moves nothing —
// claim_count counts attached citations, not sentences — and the uncited retire credits nothing.
func TestRetireCreditsOnlyTheClaimsItTookOut(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Gamma holds. Delta holds. Epsilon holds. Plain prose here.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{
		"https://a/1": []byte("a"), "https://b/1": []byte("b"), "https://g/1": []byte("g"), "https://d/1": []byte("d"),
		"https://e/1": []byte("e")}})
	a := citeSentence(t, runDir, "Alpha holds.", "https://a/1")
	b := citeSentence(t, runDir, "Beta holds.", "https://b/1")
	g := citeSentence(t, runDir, "Gamma holds.", "https://g/1")
	d := citeSentence(t, runDir, "Delta holds.", "https://d/1")
	e := citeSentence(t, runDir, "Epsilon holds.", "https://e/1")
	start := claimcount.Count(readReport(t, runDir))
	gutToAnchor(t, runDir, "Epsilon holds.", "<!--cite:"+e+"-->") // the real loss: no retire follows

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
	if start != 5 || end != 2 {
		t.Fatalf("claim_count %d → %d, want 5 → 2 — the merge keeps both citations (report %q)", start, end, readReport(t, runDir))
	}
	if got := claimLoss(t, runDir, start, end); got != 1 {
		t.Errorf("unrecorded_claim_loss = %d, want 1 — two retires took two claims out, the unretired gut is the one unaccounted", got)
	}
}

// SEVERAL ANCHORS EXIT ON ONE RETIRE, and the quote must be the whole of what left. Two cited
// sentences are cut down to their anchors in one edit — whose replacement carries them in
// DESCENDING id order, so the retire names them in an order replay must not depend on. A fragment
// of what left sweeps nothing; the whole quote takes both, the count falls by both, the retire
// credits both, and the report_op view hands replay the two removes in its documented order.
func TestRetireTakesTwoCitedSentencesOutAtOnce(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Plain prose here.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://a/1": []byte("a"), "https://b/1": []byte("b")}})
	a := citeSentence(t, runDir, "Alpha holds.", "https://a/1")
	b := citeSentence(t, runDir, "Beta holds.", "https://b/1")
	start := claimcount.Count(readReport(t, runDir))
	hi, lo := a, b
	if hi < lo {
		hi, lo = lo, hi
	}
	tok := func(id string) string { return "<!--cite:" + id + "-->" }
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Alpha holds"+tok(a)+". Beta holds"+tok(b)+".", "--new", tok(hi)+tok(lo), "--reason", "both refuted"); err != nil {
		t.Fatalf("edit both down to their anchors: %v", err)
	}

	for _, fragment := range []string{"Alpha holds.", "holds"} {
		if ev := retireClaim(t, runDir, fragment); len(ev.GetAnchors()) != 0 {
			t.Errorf("retiring the fragment %q swept %v — the quote must be the whole of what left", fragment, ev.GetAnchors())
		}
	}
	if rep := readReport(t, runDir); !strings.Contains(rep, tok(a)) || !strings.Contains(rep, tok(b)) {
		t.Fatalf("a fragment retire took an anchor out:\n%s", rep)
	}

	ev := retireClaim(t, runDir, "Alpha holds. Beta holds.")
	if !slices.Equal(ev.GetAnchors(), []string{hi, lo}) {
		t.Fatalf("retire named %v, want [%s %s] — both, in the order the edit left them", ev.GetAnchors(), hi, lo)
	}
	rep := readReport(t, runDir)
	if rep != "# Findings\n\nPlain prose here.\n" {
		t.Errorf("rendered report after retiring both:\n got %q", rep)
	}
	end := claimcount.Count(rep)
	if start != 2 || end != 0 {
		t.Errorf("claim_count %d → %d, want 2 → 0", start, end)
	}
	if got := claimLoss(t, runDir, start, end); got != 0 {
		t.Errorf("unrecorded_claim_loss = %d, want 0 — one retire took both citations out and credits both", got)
	}

	_, _, ops, err := record.ReportProjection(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var removes []string
	for _, op := range ops {
		if op.Kind == "remove" {
			removes = append(removes, op.A)
		}
	}
	if !slices.Equal(removes, []string{lo, hi}) {
		t.Errorf("report_op removes = %v, want [%s %s] — one event's rows are ordered by the anchor id, not by whatever SQLite returns", removes, lo, hi)
	}
}

// AN EMPHASIZED SENTENCE STAYS A CLAIM WHEN CITED. `blue cite` on "**Water is wet.**" places the
// anchor after the closing "**" — a "*" is not trailing punctuation — and the claim counter read
// the "**" segment as no prose: the cited claim stopped counting and was reported bare, so live
// cited prose became a removal candidate.
func TestACitedEmphasizedSentenceStaysACountedClaim(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\n- **Water is wet.**\n- Fire is hot.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://w/1": []byte("w")}})
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat,
		"--quote", `# Findings: "**Water is wet.**"`, "--url", "https://w/1", "--title", "Source w"); err != nil {
		t.Fatalf("cite: %v", err)
	}
	rep := readReport(t, runDir)
	if !regexp.MustCompile(`\*\*Water is wet\.\*\*<!--cite:c-[0-9a-f]+-->`).MatchString(rep) {
		t.Fatalf("precondition: the anchor is not after the closing emphasis:\n%s", rep)
	}
	if got := claimcount.Count(rep); got != 1 {
		t.Errorf("claim_count = %d, want 1 — the cited emphasized sentence is a claim", got)
	}
	if bare := claimcount.BareAnchorIDs(rep); len(bare) != 0 {
		t.Errorf("the anchor on live cited prose was reported bare: %v", bare)
	}
}

// A RUN OLDER THAN THE TABLE IS REFUSED WITH ITS CAUSE. The schema is fixed at a run's creation
// with no migration, so a database made before retires carried anchors has neither the
// retire_anchors table nor the report_op branch that replays it — and the generic body walk reads
// every Retire list table, so no retire recorded there could be read back by this binary. The
// write must say that and record nothing, not fail on SQLite's bare "no such table".
func TestRetireOnARunOlderThanRetireAnchorsIsRefusedWithTheCause(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	tok := "<!--cite:" + label + "-->"
	gutToAnchor(t, runDir, "The sky is blue.", tok)

	i := strings.Index(recordsql.ViewsDDL, `CREATE VIEW "report_op" AS`)
	j := strings.Index(recordsql.ViewsDDL[i:], "\n  UNION ALL\n  SELECT e.\"id\", 'remove'")
	if i < 0 || j < 0 {
		t.Fatal("cannot find the report_op view's remove branch to rebuild the pre-change view")
	}
	db := openRunDB(t, runDir)
	for _, q := range []string{`DROP VIEW "report_op"`, `DROP TABLE "retire_anchors"`, recordsql.ViewsDDL[i:i+j] + ";"} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("building the pre-change schema: %v", err)
		}
	}

	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue.", "--reason", "refuted")
	if err == nil {
		t.Fatal("a retire naming an anchor was recorded on a record with no table to hold it")
	}
	if msg := err.Error(); !strings.Contains(msg, "older binary") || !strings.Contains(msg, "retire_anchors") {
		t.Errorf("the refusal does not name the cause:\n%s", msg)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 0 {
		t.Errorf("%d retire event(s) recorded — the refused write must leave nothing behind", n)
	}
	if !strings.Contains(readReport(t, runDir), tok) {
		t.Error("the anchor left the report on a record with no way to replay its exit")
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
