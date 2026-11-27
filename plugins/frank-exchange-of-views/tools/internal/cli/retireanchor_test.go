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

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// AN ANCHORED CLAIM LEAVES WITH ITS ANCHORS, AND AN ANCHOR LEAVES ONLY WHERE AN ACT NAMES IT.
//
// A claim leaves the report through an edit that cuts its sentences whole: the pure cut takes every
// anchor on them and records each in removed_anchors. A rewrite carries every anchor, and no writer
// leaves an anchor standing bare. A retire takes one anchor of any kind off a sentence that stays word
// for word. These drive the real verbs.

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

// gutToAnchor records an edit of an anchored sentence down to its bare anchor, as an archived record
// holds one: `edit` refuses it for an evidence anchor (TestAnEditBaringEvidenceIsRefused), so an
// evidence anchor reaches a claim retire bare only from such a record or the synthesizer's round-0
// write, and the test writes the edit's event as that record carries it.
func gutToAnchor(t *testing.T, runDir, sentence, tok string) {
	t.Helper()
	recordEdit(t, runDir, strings.TrimRight(sentence, ".")+tok+".", tok)
	if !slices.Contains(claimcount.BareAnchorIDs(readReport(t, runDir)), anchor.IDs(tok)[0]) {
		t.Fatalf("the recorded edit did not leave %s bare:\n%s", tok, readReport(t, runDir))
	}
}

// recordEdit writes blue's edit of old to new as an archived record holds it, past `edit`'s guards.
func recordEdit(t *testing.T, runDir, old, new string) {
	t.Helper()
	cut := &recordpb.BlueEdit{Old: proto.String(old), New: proto.String(new), Text: proto.String("the claim cannot stand")}
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: blueSeat}, cut); err != nil {
		t.Fatalf("record the edit of %q: %v", old, err)
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
	if !slices.Equal(ev.GetAnchors(), []string{label}) || ev.GetClaim() != "The sky is blue." {
		t.Errorf("the retire recorded anchors %v, claim %q — want [%s], the sentence it stood on",
			ev.GetAnchors(), ev.GetClaim(), label)
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
// check it: the edit is refused, naming the citation and the two routes — keep words on the sentence,
// or cut it whole, which takes the citation with it — and nothing is recorded.
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
	for _, want := range []string{"anchor " + tok + " standing bare", "Nothing is recorded", "keep words on that sentence", "cut it whole"} {
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
// carries, so `show changes --id <gap>` lists it beside what red asked for in both forms, and the
// unscoped log lists it among the changes. An anchor-only retire changes no text, so the report's
// CHANGELOG, which lists what changed the text, carries no line for it (lead call 11-26).
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
	changelog, err := os.ReadFile(filepath.Join(filepath.Dir(a.Report), report.FileChangelog))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if strings.Contains(string(changelog), label) || strings.Contains(string(changelog), "retired") {
		t.Errorf("the changelog lists the anchor-only retire of %s, which changed no text:\n%s", label, changelog)
	}
}

// markerOfKind anchors an anchor of kind on sentence, with the gap a retire of it bears on where it has one.
func markerOfKind(t *testing.T, runDir, kind, sentence string) (id, gap string) {
	t.Helper()
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
	return id, gap
}

// AN ANCHOR THE QUOTE LEFT OUT IS NOT LEFT BARE. The #1677 smoke's edit cut a sentence's words and
// left out the anchors standing after them. That quote covers every word of the sentence, so the edit
// is a pure cut: it takes the anchor it left out with the sentence, records it in removed_anchors, and
// leaves no token on no words, for every kind. The same sentence cut down to its carried token is a
// rewrite that would leave the anchor bare, and is refused for every kind, nothing recorded.
func TestAnEditBaringAnAnchorItsQuoteLeftOutIsRefused(t *testing.T) {
	const sentence = "Water is wet."
	const words = "Water is wet"
	for _, kind := range anchor.Kinds() {
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			id, _ := markerOfKind(t, runDir, kind, sentence)
			tok := anchor.Token(id)
			before := readReport(t, runDir)
			edits := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
			quote := words + tok + "."
			_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", quote, "--new", tok, "--reason", "the claim cannot stand")
			if err == nil {
				t.Fatalf("an edit leaving the %s anchor %s bare was recorded:\n%s", kind, id, readReport(t, runDir))
			}
			for _, w := range []string{"anchor " + tok + " standing bare", "Nothing is recorded", "keep words on that sentence", "cut it whole"} {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("the refusal does not say %q:\n%v", w, err)
				}
			}
			if after := readReport(t, runDir); after != before {
				t.Errorf("the refused edit changed the report:\n got %q\nwant %q", after, before)
			}
			if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != edits {
				t.Errorf("%d edit(s) recorded by a refused edit", n-edits)
			}

			if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", words, "--new", "", "--reason", "the claim cannot stand"); err != nil {
				t.Fatalf("the cut of the %s's sentence, its quote leaving the anchor out, was refused: %v", kind, err)
			}
			if ev := lastBody(t, runDir, &recordpb.BlueEdit{}); !slices.Equal(ev.GetRemovedAnchors(), []string{id}) {
				t.Errorf("the cut recorded removed_anchors %v, want [%s] — the anchor its quote left out", ev.GetRemovedAnchors(), id)
			}
			rep := readReport(t, runDir)
			if want := "# Findings\n\nThe sky is blue. The grass is green.\n\n- Fire is hot.\n\nClosing paragraph stays.\n"; rep != want {
				t.Errorf("the report after the cut:\n got %q\nwant %q", rep, want)
			}
			if bare := claimcount.BareAnchorIDs(rep); len(bare) != 0 {
				t.Errorf("the cut left %v bare:\n%s", bare, rep)
			}
		})
	}
}

// THE BARE REFUSAL NAMES EVERY ANCHOR THE EDIT WOULD LEAVE BARE, carried or left out, in one refusal:
// a rewrite carrying one of a sentence's two markers to no words names both, and a citation and a
// finding carried to no words together are both named, nothing recorded either time.
func TestTheBareRefusalNamesEveryAnchorTheEditBares(t *testing.T) {
	const words = "Water is wet"
	runDir := citedRun(t, "https://water/1")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	f1, _ := markerOfKind(t, runDir, "finding", words+".")
	g, _ := markerOfKind(t, runDir, "gap", words+".")
	first, second := f1, g
	if !strings.Contains(readReport(t, runDir), words+anchor.Token(f1)+anchor.Token(g)) {
		first, second = g, f1
	}
	before := readReport(t, runDir)
	// The quote carries the first anchor of the run and stops there: the second would stand bare beside it.
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", words+anchor.Token(first), "--new", anchor.Token(first), "--reason", "gone")
	if err == nil {
		t.Fatal("a rewrite carrying one anchor to no words and leaving the other bare was recorded")
	}
	if !strings.Contains(err.Error(), "anchors "+anchor.Token(first)+", "+anchor.Token(second)+" standing bare") {
		t.Errorf("the refusal must name both anchors the edit bares, carried and left out, got: %v", err)
	}
	if readReport(t, runDir) != before || countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != 0 {
		t.Error("the refused edit changed the report or the record")
	}

	// A citation and a finding on one sentence, both carried to no words: both named in one refusal.
	runDir = citedRun(t, "https://water/1")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	c := citeSentence(t, runDir, words+".", "https://water/1")
	f, _ := markerOfKind(t, runDir, "finding", words+".")
	both := anchor.Token(c) + anchor.Token(f)
	if !strings.Contains(readReport(t, runDir), words+both+".") {
		both = anchor.Token(f) + anchor.Token(c)
	}
	if !strings.Contains(readReport(t, runDir), words+both+".") {
		t.Fatalf("the sentence does not carry the citation and the finding:\n%s", readReport(t, runDir))
	}
	before = readReport(t, runDir)
	_, err = run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", words+both+".", "--new", both, "--reason", "gone")
	if err == nil {
		t.Fatal("a rewrite baring a citation and a finding was recorded")
	}
	for _, w := range []string{anchor.Token(c), anchor.Token(f), "standing bare", "keep words on that sentence", "cut it whole"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
	if readReport(t, runDir) != before || countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != 0 {
		t.Error("the refused edit changed the report or the record")
	}
}

// ONE LIFECYCLE: an anchor of every kind leaves the report by a cut of its sentence, which takes it,
// or by a retire, which takes it off its sentence word for word. Reading the report at it names the act
// that took it (§III.10c wording site 5); a gap crediting a finding stays open either way; a gap whose own marker its
// sentence's cut took stays open and reads gone (its raiser's turn), and a gap's own marker is retired
// only once the gap is closed (F14). A hold on any kind fails its row.
func TestAnAnchorOfEveryKindLeavesByItsCutOrItsRetire(t *testing.T) {
	const sentence = "Water is wet."
	for _, kind := range anchor.Kinds() {
		for _, byCut := range []bool{true, false} {
			act := map[bool]string{true: "cut", false: "retire"}[byCut]
			t.Run(kind+"/"+act, func(t *testing.T) {
				runDir := citedRun(t, "https://water/1")
				registerChairOnce(t, runDir)
				registerLensOnce(t, runDir)
				id, gap := markerOfKind(t, runDir, kind, sentence)
				tok := anchor.Token(id)
				if !strings.Contains(readReport(t, runDir), "Water is wet"+tok) {
					t.Fatalf("no %s anchor on the bullet:\n%s", kind, readReport(t, runDir))
				}
				said := "because blue's edit cut its sentence"
				if byCut {
					if _, err := cut(t, runDir, "Water is wet"+tok+"."); err != nil {
						t.Fatalf("the cut of the %s's sentence: %v", kind, err)
					}
					if ev := lastBody(t, runDir, &recordpb.BlueEdit{}); !slices.Equal(ev.GetRemovedAnchors(), []string{id}) {
						t.Errorf("the cut took %v, want [%s]", ev.GetRemovedAnchors(), id)
					}
				} else {
					said = "because blue retired it"
					if kind == "gap" {
						closeQuoteGap(t, runDir, id)
					}
					out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--reason", "refuted", "--anchor", id)
					if err != nil {
						t.Fatalf("retire --anchor %s: %v", id, err)
					}
					if ev := lastBody(t, runDir, &recordpb.Retire{}); !slices.Equal(ev.GetAnchors(), []string{id}) || !strings.Contains(out, id) {
						t.Errorf("the retire named %v and said %q, want %s off its sentence", ev.GetAnchors(), out, id)
					}
				}
				rep := readReport(t, runDir)
				if strings.Contains(rep, tok) || !strings.Contains(rep, "- Fire is hot.\n") {
					t.Errorf("the %s anchor survived its %s, or its neighbour moved:\n%s", kind, act, rep)
				}
				if kept := strings.Contains(rep, "- "+sentence+"\n"); kept == byCut {
					t.Errorf("after the %s the sentence kept=%v, want %v:\n%s", act, kept, !byCut, rep)
				}
				_, err := run(t, "show", "report", "--seat-id", "blue-respond", "--run", runDir, "--anchor", id)
				if err == nil || !strings.Contains(err.Error(), "anchor "+id+" is not in the report "+said) {
					t.Errorf("reading at the %s anchor its %s took must name it and the act: %v", kind, act, err)
				}
				if gap != "" && (kind != "gap" || byCut) {
					b := boardOf(t, runDir)
					if len(b.Open) == 0 || b.Open[len(b.Open)-1].ID != gap {
						t.Fatalf("gap %s is not open on the board: %+v", gap, b.Open)
					}
					if kind == "gap" && b.Open[len(b.Open)-1].LocationState != record.LocationGone {
						t.Errorf("the gap whose marker its sentence's cut took reads %q, want gone", b.Open[len(b.Open)-1].LocationState)
					}
				}
			})
		}
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

// A SECOND RETIRE OF THE SAME ANCHOR IS REFUSED: the anchor is already gone, and naming it again
// would be a remove op replay cannot apply. Nothing is recorded, the report_op stream holds one remove
// for it, and replay still succeeds.
func TestRetiringTwiceNamesTheAnchorOnce(t *testing.T) {
	runDir := citedRun(t)
	id := computeSentence(t, runDir, "The sky is blue.")
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--reason", "refuted"); err != nil {
		t.Fatalf("retire --anchor %s: %v", id, err)
	}
	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--reason", "again")
	if err == nil || !strings.Contains(err.Error(), "the report holds no "+anchor.Token(id)) {
		t.Errorf("the second retire of %s = %v, want it refused: the anchor had already left", id, err)
	}
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 1 {
		t.Errorf("%d retire(s) recorded, want the one", n)
	}
	ops, err := record.ReportProjection(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var removes []string
	for _, op := range ops {
		if op.Kind == "remove" {
			removes = append(removes, op.A)
		}
	}
	if !slices.Equal(removes, []string{id}) {
		t.Errorf("report_op removes = %v, want [%s] once", removes, id)
	}
	if rep := readReport(t, runDir); !strings.Contains(rep, "The sky is blue. The grass is green.") { // replay must still succeed
		t.Errorf("the retire did not keep its sentence:\n%s", rep)
	}
}

// THE REFUSAL NAMES THE WAY OUT. A rewrite dropping an anchor is refused, and the refusal names both
// ways an anchor leaves: a cut of its sentence whole, or a retire of the anchor alone.
func TestDroppingAnAnchorTeachesRetire(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	label := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "The sky is blue<!--cite:"+label+"-->.", "--new", "The sky is azure.", "--reason", "reword")
	if err == nil {
		t.Fatal("a rewrite dropping a citation anchor was accepted")
	}
	for _, w := range []string{"cut its sentence whole", "retire it with --anchor"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q, a way an anchored claim or its anchor leaves:\n%v", w, err)
		}
	}
}

// WHAT MOVES THE COUNT: two citations retired off their sentences, a cut of uncited prose, and a
// merge of two cited sentences into one. Each citation retired leaves the count; the merge carries
// both anchors and moves nothing — the count is of attached citations, not sentences — and the
// uncited cut takes out nothing the count held.
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
	// A cut of UNCITED prose removes nothing the count held, and must credit nothing.
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", " Plain prose here.", "--new", "", "--reason", "cut"); err != nil {
		t.Fatalf("edit: %v", err)
	}
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

// SEVERAL ANCHORS EXIT ON ONE CUT. Two cited sentences rewritten down to their citations in one edit
// are refused, naming both. Two proved sentences cut whole in one edit leave with both computations:
// the cut lists them in reading order, the count falls by both, and the report_op view hands replay
// the two removes after the splice in the list's order, whatever order SQLite returns rows in.
func TestACutTakesTwoProvedSentencesOutAtOnce(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Gamma stands. Delta stands. Plain prose here.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://g/1": []byte("g"), "https://d/1": []byte("d")}})
	g := citeSentence(t, runDir, "Gamma stands.", "https://g/1")
	d := citeSentence(t, runDir, "Delta stands.", "https://d/1")
	tok := anchor.Token
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Gamma stands"+tok(g)+". Delta stands"+tok(d)+".", "--new", tok(g)+tok(d), "--reason", "both refuted")
	if err == nil || !strings.Contains(err.Error(), "anchors "+tok(g)+", "+tok(d)+" standing bare") || !strings.Contains(err.Error(), "before them in their sentence") {
		t.Errorf("cutting two cited sentences to their anchors = %v, want a refusal naming both", err)
	}
	a := computeSentence(t, runDir, "Alpha holds.")
	b := computeSentence(t, runDir, "Beta holds.")
	before := claimcount.Count(readReport(t, runDir))
	if _, err := cut(t, runDir, "Alpha holds"+tok(a)+". Beta holds"+tok(b)+"."); err != nil {
		t.Fatalf("the cut of both proved sentences: %v", err)
	}
	ev := lastBody(t, runDir, &recordpb.BlueEdit{})
	if !slices.Equal(ev.GetRemovedAnchors(), []string{a, b}) {
		t.Fatalf("the cut took %v, want [%s %s] — both, in reading order", ev.GetRemovedAnchors(), a, b)
	}
	rep := readReport(t, runDir)
	if rep != "# Findings\n\nGamma stands"+tok(g)+". Delta stands"+tok(d)+". Plain prose here.\n" {
		t.Errorf("rendered report after cutting both:\n got %q", rep)
	}
	if got := claimcount.Count(rep); got != before-2 {
		t.Errorf("the claim count = %d after the cut, want %d", got, before-2)
	}

	ops, err := record.ReportProjection(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var removes []string
	for _, op := range ops {
		if op.Kind == "remove" {
			removes = append(removes, op.A)
		}
	}
	if !slices.Equal(removes, []string{a, b}) {
		t.Errorf("report_op removes = %v, want [%s %s] — one event's rows are ordered by the list's position, not by whatever SQLite returns", removes, a, b)
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
// no round column, a cut and a retire that each took an anchor out — must migrate with no refusal and render the
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
	if _, err := cut(t, runDir, "The sky is blue"+anchor.Token(id)+"."); err != nil {
		t.Fatalf("the cut of the proved sentence: %v", err)
	}
	if ev := lastBody(t, runDir, &recordpb.BlueEdit{}); !slices.Equal(ev.GetRemovedAnchors(), []string{id}) {
		t.Fatalf("precondition: the cut took %v, want [%s]", ev.GetRemovedAnchors(), id)
	}
	kept := computeSentence(t, runDir, "The grass is green.")
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", kept, "--reason", "refuted"); err != nil {
		t.Fatalf("precondition: retire --anchor %s: %v", kept, err)
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
// when its database is created, so one made before anchors were removed by an act has neither the
// retire_anchors table, nor the anchor_removal view over it, nor the report_op branch that replays
// it, and it states the older epoch it was written at. The process that opens it is told so, with the way out, before any verb reads or
// writes — and records nothing.
func TestRetireOnARunOlderThanRetireAnchorsIsRefusedWithTheCause(t *testing.T) {
	runDir := citedRun(t)
	id := computeSentence(t, runDir, "The sky is blue.")

	i := strings.Index(recordsql.ViewsDDL, `CREATE VIEW "report_op" AS`)
	j := strings.Index(recordsql.ViewsDDL[i:], "\n  UNION ALL\n  SELECT ar.\"event_id\", 'remove'")
	if i < 0 || j < 0 {
		t.Fatal("cannot find the report_op view's remove branch to rebuild the pre-change view")
	}
	db := openRunDB(t, runDir)
	for _, q := range []string{`DROP VIEW "report_op"`, `DROP VIEW "anchor_removal"`, `DROP TABLE "retire_anchors"`, recordsql.ViewsDDL[i:i+j] + ";",
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

	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--reason", "refuted")
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

// A BULLET CUT WITH ITS MARKER QUOTED TAKES ITS ANCHORS — universe-m19's ev259, ev260 and ev401.
//
// Blue quoted an open-question bullet with its "- " and cut it, the bullet carrying two finding
// anchors. The list marker lies outside every sentence, so it is no part of the claim: the cut takes
// both anchors, the line goes whole, and no husk — a list marker and punctuation — reaches the report
// or a reader of the assembled report.
func TestACutTakesTheAnchorsOfABulletQuotedWithItsMarker(t *testing.T) {
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
		"--quote", "- Water is wet"+m[1]+".", "--new", "", "--reason", "the question is answered"); err != nil {
		t.Fatalf("the cut of the bullet, its marker quoted: %v", err)
	}
	if ev := lastBody(t, runDir, &recordpb.BlueEdit{}); !slices.Equal(ev.GetRemovedAnchors(), anchor.IDs(m[1])) {
		t.Errorf("the cut took %v, want both finding anchors %v", ev.GetRemovedAnchors(), anchor.IDs(m[1]))
	}
	want := "# Findings\n\nThe sky is blue. The grass is green.\n\n- Fire is hot.\n\nClosing paragraph stays.\n"
	if rep := readReport(t, runDir); rep != want {
		t.Errorf("the report after the cut:\n got %q\nwant %q", rep, want)
	}
	asm := assembled(t, runDir)
	if husk := regexp.MustCompile(`(?m)^[ \t]*(?:(?:[-*+]|\d+[.)])[ \t]*[.,;:!?…]*|[.,;:!?…]+)[ \t]*$`).FindString(asm); husk != "" || !strings.Contains(asm, "- Fire is hot.") {
		t.Errorf("the assembled report prints the emptied bullet %q, or lost its neighbour:\n%s", husk, asm)
	}
}

// findingsOn writes "# Findings" with sentences and a closing plain sentence, then a finding on each
// sentence per keys (keys[i] lists the finding keys on sentences[i]), and returns the findings' ids in
// the order written.
func findingsOn(t *testing.T, sentences []string, keys [][]string) (runDir string, ids []string) {
	t.Helper()
	runDir = newRun(t)
	writeReport(t, runDir, "# Findings\n\n"+strings.Join(sentences, ". ")+". Plain prose here.\n")
	registerBlue(t, runDir)
	registerLensOnce(t, runDir)
	for i, ks := range keys {
		for _, key := range ks {
			if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", sentences[i]+".",
				"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
				t.Fatalf("finding %s: %v", key, err)
			}
			ids = append(ids, lastBody(t, runDir, &recordpb.Finding{}).GetId())
		}
	}
	return runDir, ids
}

// sentenceSpan is the report's text of sentence s with the anchors standing after it, as `show report`
// prints it: from s through the full stop that ends it.
func sentenceSpan(t *testing.T, runDir, s string) string {
	t.Helper()
	rep := readReport(t, runDir)
	i := strings.Index(rep, s)
	if i < 0 {
		t.Fatalf("the report lost %q:\n%s", s, rep)
	}
	j := strings.Index(rep[i:], ".")
	return rep[i : i+j+1]
}

// ONE EDIT CUTTING CLAIMS TAKES EACH SENTENCE'S OWN ANCHORS (m25 ev247-ev367). Two sentences carry three
// finding anchors. A cut of the first sentence takes exactly its two anchors and leaves the second's on
// its sentence; a cut of the second then takes its one. One cut of both takes all three, in reading
// order. No anchor stands bare after any of them.
func TestACutTakesTheAnchorsOfEachSentenceItCuts(t *testing.T) {
	const s1, s2 = "Alpha holds for every input", "Beta holds by the same argument"
	keys := [][]string{{"F1", "F2"}, {"F3"}}

	t.Run("each sentence leaves with its own anchors", func(t *testing.T) {
		runDir, ids := findingsOn(t, []string{s1, s2}, keys)
		f1, f2, f3 := ids[0], ids[1], ids[2]
		if _, err := cut(t, runDir, sentenceSpan(t, runDir, s1)); err != nil {
			t.Fatalf("the cut of the first sentence: %v", err)
		}
		if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetRemovedAnchors(); len(got) != 2 || !slices.Contains(got, f1) || !slices.Contains(got, f2) {
			t.Errorf("cutting the first sentence took %v, want its two anchors %s and %s", got, f1, f2)
		}
		rep := readReport(t, runDir)
		if strings.Contains(rep, f1) || strings.Contains(rep, f2) || !strings.Contains(rep, s2+anchor.Token(f3)+".") {
			t.Errorf("the first cut did not take exactly its own anchors out:\n%s", rep)
		}
		if _, err := cut(t, runDir, sentenceSpan(t, runDir, s2)); err != nil {
			t.Fatalf("the cut of the second sentence: %v", err)
		}
		if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetRemovedAnchors(); !slices.Equal(got, []string{f3}) {
			t.Errorf("cutting the second sentence took %v, want [%s]", got, f3)
		}
		if rep := readReport(t, runDir); rep != "# Findings\n\nPlain prose here.\n" {
			t.Errorf("the report after both cuts:\n got %q", rep)
		}
	})

	t.Run("one cut of both takes all three", func(t *testing.T) {
		runDir, ids := findingsOn(t, []string{s1, s2}, keys)
		rep := readReport(t, runDir)
		span := rep[strings.Index(rep, s1):strings.Index(rep, " Plain prose")]
		if _, err := cut(t, runDir, span); err != nil {
			t.Fatalf("the cut of both sentences: %v", err)
		}
		if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetRemovedAnchors(); !slices.Equal(got, anchor.IDs(span)) || len(got) != 3 {
			t.Errorf("the cut took %v, want all three %v in reading order", got, anchor.IDs(span))
		}
		after := readReport(t, runDir)
		for _, id := range ids {
			if strings.Contains(after, id) {
				t.Errorf("%s stands after the cut of its sentence:\n%s", id, after)
			}
		}
		if bare := claimcount.BareAnchorIDs(after); len(bare) != 0 {
			t.Errorf("the cut left %v bare", bare)
		}
	})
}

// WHOLE SENTENCES OF ONE PARAGRAPH LEAVE TOGETHER, AND A QUOTE ACROSS TWO OF THEM IS NO CUT. Three
// sentences, one finding each: a cut quoting the last two takes both of their anchors and leaves the
// first's; a quote that runs from inside one sentence into the next covers neither whole, so it is a
// rewrite that drops the anchor inside it, refused naming it, with nothing recorded.
func TestACutTakesARunOfWholeSentences(t *testing.T) {
	const s1, s2, s3 = "Alpha holds for every input", "Beta holds by the same argument", "Gamma holds in the limit"
	keys := [][]string{{"F1"}, {"F2"}, {"F3"}}

	t.Run("two sentences leave in one cut", func(t *testing.T) {
		runDir, ids := findingsOn(t, []string{s1, s2, s3}, keys)
		f1, f2, f3 := ids[0], ids[1], ids[2]
		if _, err := cut(t, runDir, sentenceSpan(t, runDir, s2)+" "+sentenceSpan(t, runDir, s3)); err != nil {
			t.Fatalf("the cut of the last two sentences: %v", err)
		}
		if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetRemovedAnchors(); !slices.Equal(got, []string{f2, f3}) {
			t.Errorf("cutting the last two sentences took %v, want [%s %s]", got, f2, f3)
		}
		if rep := readReport(t, runDir); rep != "# Findings\n\n"+s1+anchor.Token(f1)+". Plain prose here.\n" {
			t.Errorf("the cut did not take exactly the two sentences and their anchors out:\n got %q", rep)
		}
	})

	t.Run("a quote across two sentences is no cut", func(t *testing.T) {
		runDir, ids := findingsOn(t, []string{s1, s2, s3}, keys)
		before, edits := readReport(t, runDir), countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
		_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", "every input"+anchor.Token(ids[0])+". Beta holds", "--new", "", "--reason", "unsupported")
		if err == nil || !strings.Contains(err.Error(), ids[0]) {
			t.Fatalf("a quote across two sentences with nothing in its place = %v, want it refused naming %s", err, ids[0])
		}
		if readReport(t, runDir) != before || countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != edits {
			t.Error("the refused edit changed the report or the record")
		}
	})
}
