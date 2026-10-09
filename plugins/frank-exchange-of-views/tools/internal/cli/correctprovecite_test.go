package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A CORRECTED CITE IS THE CITE, ONCE, WITH ITS NEW TITLE. The replacement carries the original's
// label and quote, so the report places the marker once where it always stood, and the
// Bibliography reads the title that stands — not the struck one first-seen by label. B8's
// synthesizer left a flagged proof note in the report because re-proving would duplicate the anchor.
func TestACorrectedCiteRendersOnceWithItsNewTitle(t *testing.T) {
	runDir := corrFixture(t)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>a source</html>")}})
	quote := "§2 the finding prose lands in a quoted sentence."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", quote, "--url", "https://src/c", "--title", "Source, as read in this run"})
	must(t, runDir, "cite", "--seat-id", "blue-respond", "--quote", quote, "--url", "https://src/c",
		"--title", "Source, section 2", "--corrects", k, "--correction-why", "the title narrated the run")

	if n := len(citeAnchorRe.FindAllString(readReport(t, runDir), -1)); n != 1 {
		t.Fatalf("a corrected cite placed %d markers, want 1:\n%s", n, readReport(t, runDir))
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_ANCHOR); n != 1 {
		t.Fatalf("a cite and its correction appended %d anchors, want the original's one", n)
	}
	srcs, err := record.CitedSources(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Title != "Source, section 2" {
		t.Fatalf("the sources = %+v, want one carrying the corrected title", srcs)
	}

	// What identifies the source is not the seat's wording: a correction that moves it is refused.
	k2 := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", "§1 first — a finding sits in sec 1 here.", "--url", "https://src/c", "--title", "Another"})
	_, err = run(t, "cite", "--run", runDir, "--seat-id", "blue-respond", "--quote", "§1 first — a finding sits in sec 1 here.",
		"--url", "https://src/elsewhere", "--title", "Another", "--corrects", k2, "--correction-why", "wrong source")
	if err == nil || !strings.Contains(err.Error(), "may change only the seat's own wording") {
		t.Fatalf("a correction moving the url was not refused as a frozen field: %v", err)
	}
}

// THE REPLACEMENT REPLAYS WHERE ITS ORIGINAL STOOD. The seat edits the sentence its cite anchors
// — carrying the marker, as an edit must — and then corrects the cite. The correction repeats the
// act's own --quote, which the report no longer reads; replayed at its own later place it would look
// for that quote after the edit and find nothing. Replayed at its original's place, it lands before
// the edit, and the edit carries it as it carried the original.
func TestACiteCorrectedAfterAnEditToItsSentenceRendersOnce(t *testing.T) {
	runDir := corrFixture(t)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/e": []byte("<html>a source</html>")}})
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", "§2 the finding prose lands in a quoted sentence.", "--url", "https://src/e", "--title", "Source, as read in this run"})
	marker := citeAnchorRe.FindString(readReport(t, runDir))
	if marker == "" {
		t.Fatalf("the cite placed no marker:\n%s", readReport(t, runDir))
	}
	must(t, runDir, "edit", "--seat-id", "blue-respond", "--quote", "lands in a quoted sentence"+marker,
		"--new", "rests in a quoted sentence"+marker, "--reason", "a plainer verb")
	must(t, runDir, "cite", "--seat-id", "blue-respond", "--quote", "§2 the finding prose lands in a quoted sentence.",
		"--url", "https://src/e", "--title", "Source, section 2",
		"--corrects", k, "--correction-why", "the title narrated the run")

	report := readReport(t, runDir)
	if n := len(citeAnchorRe.FindAllString(report, -1)); n != 1 || !strings.Contains(report, "rests in a quoted sentence"+marker) {
		t.Fatalf("after an edit and a correction the report carries %d markers, want one on the edited sentence:\n%s", n, report)
	}
	srcs, err := record.CitedSources(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Title != "Source, section 2" {
		t.Fatalf("the sources = %+v, want one carrying the corrected title", srcs)
	}
}

// A CORRECTED PROOF IS THE PROOF, ONCE, WITH ITS NEW NOTE — and it does not run again.
func TestACorrectedProofRendersOnceWithItsNewNote(t *testing.T) {
	runDir := corrFixture(t)
	s := script(t, runDir, "p.js", "console.log(91 % 7)")
	quote := "the parser accepts an empty body in this line."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"prove", "--quote", quote, "--script", s, "--reason", "this run checked it"})
	must(t, runDir, "prove", "--seat-id", "blue-respond", "--quote", quote, "--script", s,
		"--reason", "91 leaves remainder 0 on division by 7", "--corrects", k, "--correction-why", "the note narrated the run")

	if n := strings.Count(readReport(t, runDir), "<!--proof:"); n != 1 {
		t.Fatalf("a corrected proof placed %d markers, want 1:\n%s", n, readReport(t, runDir))
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_ANCHOR); n != 1 {
		t.Fatalf("a proof and its correction appended %d anchors, want the original's one", n)
	}
	proofs, err := record.RecordedProofs(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(proofs) != 1 || proofs[0].Reason != "91 leaves remainder 0 on division by 7" {
		t.Fatalf("the proofs = %+v, want one carrying the corrected note", proofs)
	}
}

// A CORRECTION REPEATS EVERY FLAG, AS ITS HELP SAYS. The correcting invocation is the act run
// again: what the seat types reaches the replacement exactly as it reaches a fresh act, and only
// what the TOOL assigned or computed (the id, the hash, the basis, the exit) is taken from the act
// it corrects. So a flag left out is a flag left out — refused where the verb requires it, refused
// as a change where the act recorded a value the correction may not move, and refused by name where
// it is the seat's own wording, which a correction would otherwise drop without saying so.
func correctWith(t *testing.T, runDir, k string, args ...string) (string, error) {
	t.Helper()
	return runAt(t, append(args, "--run", runDir, "--seat-id", "blue-respond", "--corrects", k, "--correction-why", "the note narrated the run")...)
}

// actsRecorded counts the proofs, cites and corrections on the record — what a refused correction
// must leave unchanged. (The tool's own log of the refusal is not an act.)
func actsRecorded(t *testing.T, runDir string) int {
	t.Helper()
	n := 0
	for _, ev := range events(t, runDir) {
		switch ev.GetType() {
		case recordpb.EventType_EVENT_TYPE_PROOF, recordpb.EventType_EVENT_TYPE_CITE, recordpb.EventType_EVENT_TYPE_CORRECTION:
			n++
		}
	}
	return n
}

// refusedNaming fails unless err is a refusal carrying every one of the words.
func refusedNaming(t *testing.T, what string, err error, words ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: recorded, want a refusal naming %s", what, strings.Join(words, ", "))
	}
	for _, w := range words {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("%s: the refusal does not name %s: %v", what, w, err)
		}
	}
}

func TestAProveCorrectionRepeatsEveryFlag(t *testing.T) {
	runDir := corrFixture(t)
	s := script(t, runDir, "p.js", "console.log(91 % 7)")
	quote := "the parser accepts an empty body in this line."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"prove", "--quote", quote, "--script", s,
		"--key", "P1", "--answers", "G1", "--reason", "this run checked it"})
	before := actsRecorded(t, runDir)

	// A required flag left out is refused as on any other write.
	_, err := correctWith(t, runDir, k, "prove", "--key", "P1", "--answers", "G1", "--reason", "91 leaves remainder 0")
	refusedNaming(t, "a prove correction omitting --quote and --script", err, `"quote"`, `"script"`, "a correction repeats every flag")

	// A frozen flag left out, or given a different value, is a change — named in the word a seat types.
	_, err = correctWith(t, runDir, k, "prove", "--quote", quote, "--script", s, "--answers", "G1", "--reason", "91 leaves remainder 0")
	refusedNaming(t, "a prove correction omitting the act's --key", err, "may change only the seat's own wording", "this one changes --key,")
	_, err = correctWith(t, runDir, k, "prove", "--quote", quote, "--script", s, "--key", "P1", "--reason", "91 leaves remainder 0")
	refusedNaming(t, "a prove correction omitting the act's --answers", err, "this one changes --answers,")

	// The seat's own wording left out is not a silent clear.
	_, err = correctWith(t, runDir, k, "prove", "--quote", quote, "--script", s, "--key", "P1", "--answers", "G1")
	refusedNaming(t, "a prove correction omitting the note the act holds", err, "omits --reason", "a correction repeats every flag", `--reason ""`)

	if after := actsRecorded(t, runDir); after != before {
		t.Fatalf("the refused corrections wrote %d event(s)", after-before)
	}

	// Every flag repeated, the note changed: recorded, with the new note.
	if _, err := correctWith(t, runDir, k, "prove", "--quote", quote, "--script", s, "--key", "P1", "--answers", "G1",
		"--reason", "91 leaves remainder 0 on division by 7"); err != nil {
		t.Fatalf("a prove correction repeating every flag was refused: %v", err)
	}
	p := lastBody(t, runDir, &recordpb.Proof{})
	if p.GetText() != "91 leaves remainder 0 on division by 7" || p.GetProofKey() != "P1" || p.GetAnswers() != handleText(t, runDir, "G1") || p.GetScript() != s {
		t.Fatalf("the replacement proof = %v, want the new note with every other field as the act recorded it", p)
	}
}

// The note is cleared by saying so.
func TestAProveCorrectionClearsItsNoteOnlyWhenToldTo(t *testing.T) {
	runDir := corrFixture(t)
	s := script(t, runDir, "p.js", "console.log(91 % 7)")
	quote := "the parser accepts an empty body in this line."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"prove", "--quote", quote, "--script", s, "--reason", "this run checked it"})
	if _, err := correctWith(t, runDir, k, "prove", "--quote", quote, "--script", s, "--reason", ""); err != nil {
		t.Fatalf("a prove correction passing --reason empty was refused: %v", err)
	}
	if p := lastBody(t, runDir, &recordpb.Proof{}); p.GetText() != "" {
		t.Fatalf("the replacement proof still carries the note %q", p.GetText())
	}
}

func TestACiteCorrectionRepeatsEveryFlag(t *testing.T) {
	runDir := corrFixture(t)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>a source</html>")}})
	quote := "§2 the finding prose lands in a quoted sentence."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", quote, "--url", "https://src/c",
		"--title", "Source, as read in this run", "--key", "C1", "--source-text", "summary_only", "--reason", "it states the finding"})
	before := actsRecorded(t, runDir)
	all := func(title string, more ...string) []string {
		return append([]string{"cite", "--quote", quote, "--url", "https://src/c", "--title", title}, more...)
	}

	_, err := correctWith(t, runDir, k, "cite", "--key", "C1", "--source-text", "summary_only", "--reason", "it states the finding")
	refusedNaming(t, "a cite correction omitting --quote, --url and --title", err, `"quote"`, `"url"`, `"title"`, "a correction repeats every flag")

	_, err = correctWith(t, runDir, k, all("Source, section 2", "--key", "C2", "--source-text", "summary_only", "--reason", "it states the finding")...)
	refusedNaming(t, "a cite correction under another --key", err, "may change only the seat's own wording", "this one changes --key,")
	_, err = correctWith(t, runDir, k, all("Source, section 2", "--key", "C1", "--reason", "it states the finding")...)
	refusedNaming(t, "a cite correction omitting the act's --source-text", err, "this one changes --source-text,")

	_, err = correctWith(t, runDir, k, all("Source, section 2", "--key", "C1", "--source-text", "summary_only")...)
	refusedNaming(t, "a cite correction omitting the argument the act holds", err, "omits --reason", "a correction repeats every flag", `--reason ""`)

	if after := actsRecorded(t, runDir); after != before {
		t.Fatalf("the refused corrections wrote %d event(s)", after-before)
	}

	if _, err := correctWith(t, runDir, k, all("Source, section 2", "--key", "C1", "--source-text", "summary_only", "--reason", "it states the finding")...); err != nil {
		t.Fatalf("a cite correction repeating every flag was refused: %v", err)
	}
	c := lastBody(t, runDir, &recordpb.Cite{})
	if c.GetTitle() != "Source, section 2" || c.GetText() != "it states the finding" || c.GetCiteKey() != "C1" ||
		c.GetSourceTextRead() != recordpb.SourceTextRead_SOURCE_TEXT_READ_SUMMARY_ONLY {
		t.Fatalf("the replacement cite = %v, want the new title with every other field as the act recorded it", c)
	}
}

// The argument is cleared by saying so, and an act that held none may be corrected without one.
func TestACiteCorrectionClearsItsArgumentOnlyWhenToldTo(t *testing.T) {
	runDir := corrFixture(t)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>a source</html>")}})
	quote := "§2 the finding prose lands in a quoted sentence."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", quote, "--url", "https://src/c",
		"--title", "Source, section 2", "--reason", "this sitting read it"})
	if _, err := correctWith(t, runDir, k, "cite", "--quote", quote, "--url", "https://src/c", "--title", "Source, section 2", "--reason", ""); err != nil {
		t.Fatalf("a cite correction passing --reason empty was refused: %v", err)
	}
	if c := lastBody(t, runDir, &recordpb.Cite{}); c.GetText() != "" {
		t.Fatalf("the replacement cite still carries the argument %q", c.GetText())
	}
}
