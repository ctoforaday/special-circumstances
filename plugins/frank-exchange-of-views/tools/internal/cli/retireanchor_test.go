package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// AN ANCHORED CLAIM LEAVES THROUGH RETIRE, AND ITS ANCHOR LEAVES WITH IT.
//
// An edit may carry an anchor but never drop one, so the way to delete a sentence a finding, computation
// or gap marks is to edit it down to the bare anchor and retire it; a citation never stands bare,
// so a cited sentence loses its citation to a retire first and is then prose an edit may cut. Before the retire named the anchors that
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
		"--quote", sentence, "--url", url, "--title", "Source "+url); err != nil {
		t.Fatalf("cite %q: %v", sentence, err)
	}
	re := regexp.MustCompile(regexp.QuoteMeta(strings.TrimRight(sentence, ".")) + `<!--cite:(C-[0-9a-f]{8})-->`)
	m := re.FindStringSubmatch(readReport(t, runDir))
	if m == nil {
		t.Fatalf("no cite anchor after %q:\n%s", sentence, readReport(t, runDir))
	}
	return m[1]
}

// gutToAnchor edits an anchored sentence down to its bare anchor — the only edit that removes
// the sentence, because an edit may not drop the anchor. A citation's anchor is never left bare
// (TestACitedSentenceIsNotCutToItsAnchor), so the anchor gutted is a computation's.
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

// computeSentence backs one sentence with a computation and returns its anchor id: a marker an edit may
// leave bare, for the retire to take out with its claim.
func computeSentence(t *testing.T, runDir, sentence string) string {
	t.Helper()
	name := "compute-" + strings.Trim(strings.Fields(sentence)[0], ".") + ".js"
	s := script(t, runDir, name, "console.log('it holds');")
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", sentence,
		"--dir", s, "--reason", "the computation settles it", "--", "node", name); err != nil {
		t.Fatalf("compute %q: %v", sentence, err)
	}
	return lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
}

// countEvents is how many events of one type the record holds.
func countEvents(t *testing.T, runDir string, typ recordpb.EventType) int {
	t.Helper()
	var n int
	if err := openRunDB(t, runDir).QueryRow(`SELECT count(*) FROM "events" WHERE "type" = ?`, recordpb.Word(typ)).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
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

// A CITED SENTENCE LEAVES IN TWO ACTS: the retire takes the citation off it, keeping it word for
// word, and the edit then cuts plain prose. The retire names the citation and says the claim
// stayed, the count falls by it, and the assembled report neither numbers nor lists its source.
func TestACitationRetiresOffItsSentenceAndAnEditThenCutsIt(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	before := claimcount.Count(readReport(t, runDir))

	out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", label, "--reason", "the source cannot be read")
	if err != nil {
		t.Fatalf("retire --anchor %s alone: %v", label, err)
	}
	ev := lastBody(t, runDir, &recordpb.Retire{})
	if !slices.Equal(ev.GetAnchors(), []string{label}) || !ev.GetClaimKept() || ev.GetClaim() != "The sky is blue." {
		t.Errorf("the retire recorded anchors %v, claim_kept %v, claim %q — want [%s], kept, the sentence it stood on",
			ev.GetAnchors(), ev.GetClaimKept(), ev.GetClaim(), label)
	}
	if !strings.Contains(out, label) {
		t.Errorf("the retire's answer does not name %s:\n%s", label, out)
	}
	if rep := readReport(t, runDir); !strings.HasPrefix(rep, "# Findings\n\nThe sky is blue. The grass is green.\n") {
		t.Errorf("the citation's sentence did not stay word for word:\n%s", rep)
	}
	if got := claimcount.Count(readReport(t, runDir)); got != before-1 {
		t.Errorf("the claim count = %d after retiring one of %d citations, want %d", got, before, before-1)
	}
	if asm := assembled(t, runDir); strings.Contains(asm, "https://sky/1") || strings.Contains(asm, "[^1]") {
		t.Errorf("the assembled report still numbers or lists the retired citation:\n%s", asm)
	}

	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "The sky is blue. ", "--new", "", "--reason", "the claim cannot stand"); err != nil {
		t.Fatalf("editing out the sentence its citation left: %v", err)
	}
	want := "# Findings\n\nThe grass is green.\n\n- Water is wet.\n- Fire is hot.\n\nClosing paragraph stays.\n"
	if rep := readReport(t, runDir); rep != want {
		t.Errorf("rendered report after the edit:\n got %q\nwant %q", rep, want)
	}
}

// A CITED SENTENCE IS NOT CUT TO ITS ANCHOR. A bare citation backs no sentence, so no lens could
// check it: the edit is refused, naming the citation and the retire that takes it off first, and
// nothing is recorded.
func TestACitedSentenceIsNotCutToItsAnchor(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	tok := anchor.Token(label)
	before := readReport(t, runDir)
	edits := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "The sky is blue"+tok+".", "--new", tok, "--reason", "the claim cannot stand")
	if err == nil {
		t.Fatal("an edit leaving a citation bare was recorded")
	}
	for _, want := range []string{"citation " + label + " standing bare", "Nothing is recorded", "`retire --anchor <its id>`", "keeps the sentence word for word"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q:\n%v", want, err)
		}
	}
	if after := readReport(t, runDir); after != before {
		t.Errorf("the refused edit changed the report:\n got %q\nwant %q", after, before)
	}
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != edits {
		t.Errorf("%d edit(s) recorded by a refused edit", n-edits)
	}
}

// A CITATION RETIRE ANSWERING A GAP IS IN THE GAP'S CHANGE LOG. The retire takes the key an edit
// carries, so `show changes --id <gap>` lists it beside what red asked for in both forms, the
// unscoped log lists it among the changes, and the report's retirements name the gap. A claim retire
// takes no gap — the edit that cut the claim names it — and is refused one.
func TestACitationRetireAnsweringAGapIsInItsChangeLog(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	gap := mintGap(t, runDir, "dead-source", "spec_underspecification")
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", label,
		"--answers", gap, "--reason", "the source cannot be read"); err != nil {
		t.Fatalf("retire --anchor %s --answers %s: %v", label, gap, err)
	}
	if ev := lastBody(t, runDir, &recordpb.Retire{}); ev.GetAnswers() != gap {
		t.Fatalf("the retire recorded answers %q, want %s", ev.GetAnswers(), gap)
	}

	md, err := run(t, "show", "changes", "--id", gap, "--run", runDir, "--seat-id", lensSeat)
	if err != nil {
		t.Fatalf("show changes --id %s: %v", gap, err)
	}
	for _, want := range []string{"## Changes blue recorded against it (1)", "retired citation " + label + ", its sentence kept", "the source cannot be read", "The sky is blue."} {
		if !strings.Contains(md, want) {
			t.Errorf("show changes --id %s does not say %q:\n%s", gap, want, md)
		}
	}
	js, err := run(t, "show", "changes", "--id", gap, "--run", runDir, "--seat-id", lensSeat, "--json")
	if err != nil {
		t.Fatalf("show changes --id %s --json: %v", gap, err)
	}
	var page record.ChangesJSON
	if err := json.Unmarshal([]byte(js), &page); err != nil {
		t.Fatalf("show changes --json is not valid JSON (%v):\n%s", err, js)
	}
	if len(page.Edits) != 1 || page.Edits[0].Retired != label || page.Edits[0].Answers != gap || page.Edits[0].Old != "The sky is blue." {
		t.Errorf("show changes --id %s --json holds %+v, want the one citation retire of %s answering it", gap, page.Edits, label)
	}
	if all, err := run(t, "show", "changes", "--run", runDir, "--seat-id", lensSeat); err != nil || !strings.Contains(all, "answers **"+gap+"** · retired citation "+label) {
		t.Errorf("the unscoped change log does not list the retire answering %s (%v):\n%s", gap, err, all)
	}
	a, err := report.Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if changelog, err := os.ReadFile(filepath.Join(filepath.Dir(a.Report), report.FileChangelog)); err != nil ||
		!strings.Contains(string(changelog), "a source it cited was retired from it, answering **"+gap+"**") {
		t.Errorf("the changelog's retirements do not name the gap the retire answers (%v):\n%s", err, changelog)
	}

	computation := computeSentence(t, runDir, "The grass is green.")
	gutToAnchor(t, runDir, "The grass is green.", anchor.Token(computation))
	_, err = run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "The grass is green.",
		"--answers", gap, "--reason", "refuted")
	if err == nil || !strings.Contains(err.Error(), "the edit that cut it") {
		t.Errorf("a claim retire naming a gap was not refused for it: %v", err)
	}
}

// THE BARE-CITATION REFUSAL SENDS BLUE TO A RETIRE THAT ANSWERS THE SAME GAP. An edit answering a gap
// that would leave a citation bare names the retire with that gap, so the first half of the repair
// joins the gap as the edit would have.
func TestTheBareCitationRefusalNamesTheGapTheEditAnswers(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	gap := mintGap(t, runDir, "dead-source", "spec_underspecification")
	tok := anchor.Token(label)
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--answers", gap,
		"--quote", "The sky is blue"+tok+".", "--new", tok, "--reason", "the claim cannot stand")
	if err == nil || !strings.Contains(err.Error(), "`retire --anchor <its id> --answers "+gap+"`") {
		t.Errorf("the bare-citation refusal does not name the retire answering %s: %v", gap, err)
	}
}

// ONE LIFECYCLE: an anchor of every kind but a citation, cut to bare and named to the retire, leaves
// with its claim; a citation, named alone, leaves its sentence standing — a finding's while a gap crediting it is open, a gap's while the gap itself is open. The
// retire names it, reading the report at it names the retire, and the gap reads gone. A hold on
// any kind fails its row.
func TestRetireTakesABareMarkerOfEveryKind(t *testing.T) {
	const sentence = "Water is wet."
	for _, kind := range anchor.Kinds() {
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			mint := func(args ...string) string {
				t.Helper()
				out, err := run(t, append([]string{"mint", "--run", runDir, "--seat-id", lensSeat,
					"--class", "spec_underspecification", "--problem", "the defect", "--fix", "the fix",
					"--check-kind", "document", "--check", "the acceptance check", "--severity", "medium",
					"--likelihood", "medium", "--impact", "medium", "--complexity", "low"}, args...)...)
				if err != nil {
					t.Fatalf("mint: %v", err)
				}
				return gapID(out)
			}
			var id, gap string
			switch kind {
			case "finding":
				if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--quote", sentence,
					"--reason", "overclaim", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
					t.Fatalf("finding: %v", err)
				}
				f := lastBody(t, runDir, &recordpb.Finding{})
				id = f.GetId()
				gap = mint("--key", "credits", "--found-by", id, "--about-kind", "section", "--about", "Findings")
			case "citation":
				id = citeSentence(t, runDir, sentence, "https://water/1")
			case "computation":
				s := script(t, runDir, "wet.js", "console.log('wet');")
				if _, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", sentence,
					"--dir", s, "--reason", "the computation settles it", "--", "node", "wet.js"); err != nil {
					t.Fatalf("compute: %v", err)
				}
				id = lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
			case "gap":
				id = mint("--key", "itself", "--quote", sentence)
				gap = id
			default:
				t.Fatalf("no row drives the %s kind — a kind added to the table needs one here", kind)
			}
			tok := anchor.Token(id)
			if !strings.Contains(readReport(t, runDir), "Water is wet"+tok) {
				t.Fatalf("no %s anchor on the bullet:\n%s", kind, readReport(t, runDir))
			}
			args := []string{"retire", "--run", runDir, "--seat-id", blueSeat, "--reason", "refuted", "--anchor", id}
			if kind != "citation" { // a citation never stands bare: it retires alone, its sentence kept
				gutToAnchor(t, runDir, sentence, tok)
				args = append(args, "--quote", sentence)
			}
			out, err := run(t, args...)
			if err != nil {
				t.Fatalf("retire --anchor %s: %v", id, err)
			}
			if ev := lastBody(t, runDir, &recordpb.Retire{}); !slices.Equal(ev.GetAnchors(), []string{id}) || !strings.Contains(out, id) {
				t.Errorf("the retire named %v and said %q, want %s out with its claim", ev.GetAnchors(), out, id)
			}
			if rep := readReport(t, runDir); strings.Contains(rep, tok) || !strings.Contains(rep, "- Fire is hot.\n") {
				t.Errorf("the bare %s anchor survived the retire, or its neighbour moved:\n%s", kind, rep)
			}
			_, err = run(t, "show", "report", "--seat-id", "blue-respond", "--run", runDir, "--anchor", id)
			if err == nil || !strings.Contains(err.Error(), "RETIRED") {
				t.Errorf("reading at the retired %s anchor must name the retire: %v", kind, err)
			}
			if gap != "" {
				b := boardOf(t, runDir)
				if len(b.Open) == 0 || b.Open[len(b.Open)-1].ID != gap {
					t.Fatalf("gap %s is not open on the board: %+v", gap, b.Open)
				}
				if kind == "gap" && b.Open[len(b.Open)-1].LocationState != record.LocationGone {
					t.Errorf("the gap whose anchor retired reads %q, want gone", b.Open[len(b.Open)-1].LocationState)
				}
			}
		})
	}
}

// boardOf is the board a seat reads, parsed.
func boardOf(t *testing.T, runDir string) record.BoardJSON {
	t.Helper()
	out, err := run(t, "show", "board", "--run", runDir, "--seat-id", blueSeat)
	if err != nil {
		t.Fatalf("show board: %v", err)
	}
	var b record.BoardJSON
	if err := json.Unmarshal([]byte(out), &b); err != nil {
		t.Fatal(err)
	}
	return b
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
	runDir := citedRun(t)
	id := computeSentence(t, runDir, "The sky is blue.")
	gutToAnchor(t, runDir, "The sky is blue.", anchor.Token(id))
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
	if !strings.Contains(err.Error(), "`retire`") {
		t.Errorf("the refusal does not name `retire` as the way an anchored claim leaves:\n%v", err)
	}
}

// WHAT MOVES THE COUNT: two citations retired off their sentences, a retire of uncited prose, and
// a merge of two cited sentences into one. Each citation retired leaves the count; the merge carries
// both anchors and moves nothing — the count is of attached citations, not sentences — and the
// uncited retire takes out nothing the count held.
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
	citeSentence(t, runDir, "Epsilon holds.", "https://e/1")
	start := claimcount.Count(readReport(t, runDir))

	for _, id := range []string{a, b} {
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--reason", "refuted"); err != nil {
			t.Fatalf("retire --anchor %s: %v", id, err)
		}
	}
	// A retire of UNCITED prose removes nothing the count held, and must credit nothing.
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", " Plain prose here.", "--new", "", "--reason", "cut"); err != nil {
		t.Fatalf("edit: %v", err)
	}
	retireClaim(t, runDir, "Plain prose here.")
	// The unrelated merge: two cited sentences become one, both anchors carried, no retire.
	gt, dt := "<!--cite:"+g+"-->", "<!--cite:"+d+"-->"
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Gamma holds"+gt+". Delta holds"+dt, "--new", "Gamma and Delta hold"+gt+dt, "--reason", "chair"); err != nil {
		t.Fatalf("merge edit: %v", err)
	}
	end := claimcount.Count(readReport(t, runDir))
	if start != 5 || end != 3 {
		t.Fatalf("the claim count %d → %d, want 5 → 3 — the merge keeps both citations (report %q)", start, end, readReport(t, runDir))
	}
}

// SEVERAL ANCHORS EXIT ON ONE RETIRE. Two proved sentences are cut down to their anchors in one edit —
// whose replacement carries them in DESCENDING id order, so the retire names them in an order replay
// must not depend on. A fragment of what left is refused, naming both; the whole quote takes both,
// the count falls by both, the retire credits both, and the report_op view hands replay the two
// removes in its documented order.
func TestRetireTakesTwoProvedSentencesOutAtOnce(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Gamma stands. Delta stands. Plain prose here.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://g/1": []byte("g"), "https://d/1": []byte("d")}})
	// Two CITED sentences cut in one edit are refused, the refusal naming both citations.
	g := citeSentence(t, runDir, "Gamma stands.", "https://g/1")
	d := citeSentence(t, runDir, "Delta stands.", "https://d/1")
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Gamma stands"+anchor.Token(g)+". Delta stands"+anchor.Token(d)+".", "--new", anchor.Token(g)+anchor.Token(d), "--reason", "both refuted")
	if err == nil || !strings.Contains(err.Error(), "citations "+g+", "+d+" standing bare") || !strings.Contains(err.Error(), "before them in their sentence") {
		t.Errorf("cutting two cited sentences to their anchors = %v, want a refusal naming both", err)
	}
	a := computeSentence(t, runDir, "Alpha holds.")
	b := computeSentence(t, runDir, "Beta holds.")
	hi, lo := a, b
	if hi < lo {
		hi, lo = lo, hi
	}
	tok := anchor.Token
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Alpha holds"+tok(a)+". Beta holds"+tok(b)+".", "--new", tok(hi)+tok(lo), "--reason", "both refuted"); err != nil {
		t.Fatalf("edit both down to their anchors: %v", err)
	}

	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "holds", "--reason", "refuted"); err == nil ||
		!strings.Contains(err.Error(), a) || !strings.Contains(err.Error(), b) {
		t.Errorf("retiring the fragment \"holds\" = %v, want a refusal naming both bare anchors", err)
	}
	if rep := readReport(t, runDir); !strings.Contains(rep, tok(a)) || !strings.Contains(rep, tok(b)) {
		t.Fatalf("a fragment retire took an anchor out:\n%s", rep)
	}

	ev := retireClaim(t, runDir, "Alpha holds. Beta holds.")
	if !slices.Equal(ev.GetAnchors(), []string{hi, lo}) {
		t.Fatalf("retire named %v, want [%s %s] — both, in the order the edit left them", ev.GetAnchors(), hi, lo)
	}
	rep := readReport(t, runDir)
	if rep != "# Findings\n\nGamma stands"+tok(g)+". Delta stands"+tok(d)+". Plain prose here.\n" {
		t.Errorf("rendered report after retiring both:\n got %q", rep)
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
		"--quote", "**Water is wet.**", "--url", "https://w/1", "--title", "Source w"); err != nil {
		t.Fatalf("cite: %v", err)
	}
	rep := readReport(t, runDir)
	if !regexp.MustCompile(`\*\*Water is wet\.\*\*<!--cite:C-[0-9a-f]{8}-->`).MatchString(rep) {
		t.Fatalf("precondition: the anchor is not after the closing emphasis:\n%s", rep)
	}
	if got := claimcount.Count(rep); got != 1 {
		t.Errorf("the claim count = %d, want 1 — the cited emphasized sentence is a claim", got)
	}
	if bare := claimcount.BareAnchorIDs(rep); len(bare) != 0 {
		t.Errorf("the anchor on live cited prose was reported bare: %v", bare)
	}
}

// THE WAY OUT THE OLDER-RUN MESSAGE NAMES MUST WORK ON A ROUNDLESS-ERA RECORD. `migrate` read a
// `round` column the envelope no longer has and synthesized a cast over a record that already
// carried one, so migrating any record written since roundless failed. A current record — cast,
// no round column, a retire that took an anchor out — must migrate with no refusal and render the
// same report.
func TestARoundlessEraRecordMigratesClean(t *testing.T) {
	// A real run's record opens with its cast, before any seat registers — as the 2026-09-10 B3
	// record does. The migration must replay THAT cast, not synthesize a second one over it.
	runDir := newRun(t)
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: []string{"blue-respond", "blue-synthesize"}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	writeReport(t, runDir, retireBase)
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/1": []byte("<html>sky</html>")}})
	id := computeSentence(t, runDir, "The sky is blue.")
	gutToAnchor(t, runDir, "The sky is blue.", anchor.Token(id))
	if ev := retireClaim(t, runDir, "The sky is blue."); len(ev.GetAnchors()) != 1 {
		t.Fatalf("precondition: the retire named %v, want the one anchor", ev.GetAnchors())
	}
	want := readReport(t, runDir)

	dst := filepath.Join(recordtest.TmpRun(t), "migrated") // TmpRun releases the migrated record's handle
	out, err := runMigrate(t, "--from", runDir, "--to", dst)
	if err != nil {
		t.Fatalf("migrating a roundless-era record: %v\n%s", err, out)
	}
	if strings.Contains(out, "REFUSAL") {
		t.Errorf("the migration refused an event:\n%s", out)
	}
	if got := readReport(t, dst); got != want {
		t.Errorf("the migrated report differs:\n got %q\nwant %q", got, want)
	}
}

// A RUN OLDER THAN THE TABLE IS REFUSED WITH ITS CAUSE, AT THE OPEN. A record's schema is fixed
// when its database is created, so one made before retires carried anchors has neither the
// retire_anchors table nor the report_op branch that replays it, and it states the older epoch it was
// written at. The process that opens it is told so, with the way out, before any verb reads or
// writes — and records nothing.
func TestRetireOnARunOlderThanRetireAnchorsIsRefusedWithTheCause(t *testing.T) {
	runDir := citedRun(t)
	gutToAnchor(t, runDir, "The sky is blue.", anchor.Token(computeSentence(t, runDir, "The sky is blue.")))

	i := strings.Index(recordsql.ViewsDDL, `CREATE VIEW "report_op" AS`)
	j := strings.Index(recordsql.ViewsDDL[i:], "\n  UNION ALL\n  SELECT e.\"id\", 'remove'")
	if i < 0 || j < 0 {
		t.Fatal("cannot find the report_op view's remove branch to rebuild the pre-change view")
	}
	db := openRunDB(t, runDir)
	for _, q := range []string{`DROP VIEW "report_op"`, `DROP TABLE "retire_anchors"`, recordsql.ViewsDDL[i:i+j] + ";",
		fmt.Sprintf(`UPDATE "schema_epoch" SET "epoch" = %d WHERE "id" = 1`, recordpb.EventSchema-1)} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("building the pre-change schema: %v", err)
		}
	}

	// The handle this test's own setup cached passed the open's check before the schema was cut
	// back; the process that meets an older run is one that has not opened it yet.
	if err := recordsql.CloseUnder(runDir); err != nil {
		t.Fatal(err)
	}

	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue.", "--reason", "refuted")
	if err == nil {
		t.Fatal("a retire naming an anchor was recorded on a record with no table to hold it")
	}
	for _, want := range []string{"older binary", fmt.Sprintf("epoch %d", recordpb.EventSchema-1), "`feov-migrate --from <runDir> --to <freshDir>`"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %s:\n%v", want, err)
		}
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "events" WHERE "type" = ?`, recordpb.Word(recordpb.EventType_EVENT_TYPE_RETIRE)).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d retire event(s) recorded — the refused open must leave nothing behind", n)
	}
}

// A BULLET CUT DOWN TO ITS ANCHORS, MARKER QUOTED, RETIRES — universe-m19's ev259, ev260 and ev401.
//
// Blue quoted an open-question bullet with its "- " and replaced it with "- " and the bullet's four
// finding anchors; then retired the question, quoting its words without the marker. The retire
// compared its quote with the edit's whole old span, the marker made the two differ, and it took no
// anchor, said "retired:" and nothing else — twice. The bullet stood as four tokens and a period,
// and the assembled report printed "- ." in a VERIFIED run.
//
// The list marker lies outside every sentence, so it is no part of the claim on either side: the
// retire takes both anchors, the line goes whole, and the claim is recorded as verified.
func TestRetireTakesTheAnchorsOfABulletQuotedWithItsMarker(t *testing.T) {
	runDir := citedRun(t)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	for _, key := range []string{"F1", "F2"} {
		if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", "Water is wet.",
			"--reason", "unsupported "+key, "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
			t.Fatalf("finding: %v", err)
		}
	}
	m := regexp.MustCompile(`- Water is wet((?:<!--fx:F-[0-9a-f]{8}-->){2})\.`).FindStringSubmatch(readReport(t, runDir))
	if m == nil {
		t.Fatalf("the bullet does not hold two finding anchors:\n%s", readReport(t, runDir))
	}
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "- Water is wet"+m[1]+".", "--new", "- "+m[1], "--reason", "the question is answered"); err != nil {
		t.Fatalf("the cut down to the bullet's anchors: %v", err)
	}
	// THE HUSK NEVER REACHES A READER, retired or not: assembly strips the finding tokens, and a line
	// left with a list marker and punctuation goes with them.
	asm := assembled(t, runDir)
	if husk := regexp.MustCompile(`(?m)^[ \t]*(?:(?:[-*+]|\d+[.)])[ \t]*[.,;:!?…]*|[.,;:!?…]+)[ \t]*$`).FindString(asm); husk != "" || !strings.Contains(asm, "- Fire is hot.") {
		t.Errorf("the assembled report prints the emptied bullet %q, or lost its neighbour:\n%s", husk, asm)
	}

	ev := retireClaim(t, runDir, "Water is wet.")
	if len(ev.GetAnchors()) != 2 || ev.GetRemovalBasis() != record.RemovalVerified {
		t.Errorf("the retire took %v on basis %q, want both finding anchors, verified", ev.GetAnchors(), ev.GetRemovalBasis())
	}
	want := "# Findings\n\nThe sky is blue. The grass is green.\n\n- Fire is hot.\n\nClosing paragraph stays.\n"
	if rep := readReport(t, runDir); rep != want {
		t.Errorf("the report after the retire:\n got %q\nwant %q", rep, want)
	}
}

// ONE EDIT CUTTING TWO CLAIMS LEAVES EACH ITS OWN ANCHORS (m25 ev247-ev367). Blue cut two sentences
// carrying three finding anchors down to the anchors in one edit, then retired each sentence with its
// own reason: each retire took nothing, the three anchors stood bare, and the retire naming two of
// them was refused as "no edit that took this claim out left" them — the edit had. A retire of a whole
// sentence of the cut takes the anchors that sentence held; one naming an anchor of the other sentence
// is refused saying where that anchor stood; a fragment is refused while the anchors stand bare.
func TestRetireTakesTheAnchorsOfEachSentenceItsCutTook(t *testing.T) {
	const s1, s2 = "Alpha holds for every input", "Beta holds by the same argument"
	setup := func(t *testing.T) (runDir string, f1, f2, f3 string) {
		t.Helper()
		runDir = newRun(t)
		writeReport(t, runDir, "# Findings\n\n"+s1+". "+s2+". Plain prose here.\n")
		registerBlue(t, runDir)
		registerLensOnce(t, runDir)
		find := func(key, quote string) string {
			t.Helper()
			if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", quote,
				"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
				t.Fatalf("finding %s: %v", key, err)
			}
			return lastBody(t, runDir, &recordpb.Finding{}).GetId()
		}
		f1, f2, f3 = find("F1", s1+"."), find("F2", s1+"."), find("F3", s2+".")
		rep := readReport(t, runDir)
		i, j := strings.Index(rep, s1), strings.Index(rep, "Plain prose")
		if i < 0 || j < 0 {
			t.Fatalf("the report lost its sentences:\n%s", rep)
		}
		cut := strings.TrimRight(rep[i:j], " ")
		var toks string
		for _, id := range anchor.IDs(cut) {
			toks += anchor.Token(id)
		}
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
			"--quote", cut, "--new", toks, "--reason", "both claims leave"); err != nil {
			t.Fatalf("edit both sentences down to their anchors: %v", err)
		}
		return runDir, f1, f2, f3
	}

	t.Run("each sentence leaves with its own anchors", func(t *testing.T) {
		runDir, f1, f2, f3 := setup(t)
		ev := retireClaim(t, runDir, s1)
		if got := ev.GetAnchors(); len(got) != 2 || !slices.Contains(got, f1) || !slices.Contains(got, f2) {
			t.Errorf("retiring the first sentence named %v, want its two anchors %s and %s", got, f1, f2)
		}
		rep := readReport(t, runDir)
		if strings.Contains(rep, f1) || strings.Contains(rep, f2) || !strings.Contains(rep, anchor.Token(f3)) {
			t.Errorf("the first retire did not take exactly its own anchors out:\n%s", rep)
		}
		if ev := retireClaim(t, runDir, s2+"."); !slices.Equal(ev.GetAnchors(), []string{f3}) {
			t.Errorf("retiring the second sentence named %v, want [%s]", ev.GetAnchors(), f3)
		}
		if rep := readReport(t, runDir); rep != "# Findings\n\nPlain prose here.\n" {
			t.Errorf("the report after both retires:\n got %q", rep)
		}
	})

	t.Run("a bare anchor is retired by naming it", func(t *testing.T) {
		runDir, f1, f2, f3 := setup(t)
		_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", s1, "--reason", "unsupported",
			"--anchor", f3)
		if err == nil || !strings.Contains(err.Error(), "did not stand on this claim") || !strings.Contains(err.Error(), s2) {
			t.Fatalf("naming the other sentence's anchor = %v, want a refusal saying it stood on %q", err, s2)
		}
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", s1, "--reason", "unsupported",
			"--anchor", f1+","+f2); err != nil {
			t.Fatalf("naming the first sentence's own anchors: %v", err)
		}
		if ev := lastBody(t, runDir, &recordpb.Retire{}); len(ev.GetAnchors()) != 2 {
			t.Errorf("the named retire took %v, want %s and %s", ev.GetAnchors(), f1, f2)
		}
	})

	t.Run("a fragment is refused while the anchors stand bare", func(t *testing.T) {
		runDir, f1, f2, _ := setup(t)
		before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE)
		_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "for every input", "--reason", "unsupported")
		if err == nil || !strings.Contains(err.Error(), f1) || !strings.Contains(err.Error(), f2) || !strings.Contains(err.Error(), s1) {
			t.Fatalf("a fragment of the cut = %v, want a refusal naming %s, %s and the sentence to quote", err, f1, f2)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != before {
			t.Errorf("a refused retire recorded %d retire(s)", n-before)
		}
	})
}

// WHOLE SENTENCES OF A CUT LEAVE TOGETHER, AND A QUOTE ACROSS TWO OF THEM IS A FRAGMENT. Three
// sentences, one finding each, cut down to their anchors in one edit: a retire quoting the last two
// as one claim takes both of their anchors and leaves the first's; a quote that runs from inside one
// sentence into the next is part of what the edit cut and no whole sentence of it, so it is refused
// naming the anchors still bare, with nothing recorded.
func TestRetireTakesARunOfWholeSentencesOfItsCut(t *testing.T) {
	const s1, s2, s3 = "Alpha holds for every input", "Beta holds by the same argument", "Gamma holds in the limit"
	setup := func(t *testing.T) (runDir string, f1, f2, f3 string) {
		t.Helper()
		runDir = newRun(t)
		writeReport(t, runDir, "# Findings\n\n"+s1+". "+s2+". "+s3+". Plain prose here.\n")
		registerBlue(t, runDir)
		registerLensOnce(t, runDir)
		find := func(key, quote string) string {
			t.Helper()
			if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", quote,
				"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
				t.Fatalf("finding %s: %v", key, err)
			}
			return lastBody(t, runDir, &recordpb.Finding{}).GetId()
		}
		f1, f2, f3 = find("F1", s1+"."), find("F2", s2+"."), find("F3", s3+".")
		rep := readReport(t, runDir)
		i, j := strings.Index(rep, s1), strings.Index(rep, "Plain prose")
		if i < 0 || j < 0 {
			t.Fatalf("the report lost its sentences:\n%s", rep)
		}
		cut := strings.TrimRight(rep[i:j], " ")
		var toks string
		for _, id := range anchor.IDs(cut) {
			toks += anchor.Token(id)
		}
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
			"--quote", cut, "--new", toks, "--reason", "the three claims leave"); err != nil {
			t.Fatalf("edit the three sentences down to their anchors: %v", err)
		}
		return runDir, f1, f2, f3
	}

	t.Run("two sentences leave as one claim", func(t *testing.T) {
		runDir, f1, f2, f3 := setup(t)
		ev := retireClaim(t, runDir, s2+". "+s3+".")
		if got := ev.GetAnchors(); len(got) != 2 || !slices.Contains(got, f2) || !slices.Contains(got, f3) {
			t.Errorf("retiring the last two sentences named %v, want %s and %s", got, f2, f3)
		}
		if rep := readReport(t, runDir); !strings.Contains(rep, anchor.Token(f1)) || strings.Contains(rep, f2) || strings.Contains(rep, f3) {
			t.Errorf("the retire did not take exactly the two sentences' anchors out:\n%s", rep)
		}
	})

	t.Run("a quote across two sentences is a fragment", func(t *testing.T) {
		runDir, f1, f2, f3 := setup(t)
		before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE)
		_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "every input. Beta holds", "--reason", "unsupported")
		if err == nil || !strings.Contains(err.Error(), "fragment") {
			t.Fatalf("a quote across two sentences of the cut = %v, want it refused as a fragment", err)
		}
		for _, id := range []string{f1, f2, f3} {
			if !strings.Contains(err.Error(), id) {
				t.Errorf("the refusal does not name the bare anchor %s: %v", id, err)
			}
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != before {
			t.Errorf("a refused retire recorded %d retire(s)", n-before)
		}
	})
}
