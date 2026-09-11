package blue

import (
	"strings"
	"testing"
)

// THE BASE IS THE LARGEST AUTHORED ARTIFACT IN A RUN AND WAS THE ONE PIECE NEVER VOICE-CHECKED.
//
// #873, measured on the #861 arm-A run: the advisory had exactly one call site, `blue edit`, and
// `blue ingest` wrote 16,874 characters straight to the record past it. Every inline lane tag in
// that run's finished VERIFIED report arrived through this path.
func TestIngestCensusesTheWholeBase(t *testing.T) {
	base := `# On 91

[minority: lane-3] The factorisation is 7 x 13.
This run checked it three ways. [minority: lane-1] The debate also
considered Fermat witnesses. [minority: lane-3] See this report's appendix.`

	census := voiceCensus(base)
	if len(census) == 0 {
		t.Fatal("a base carrying six tells produced no census")
	}
	joined := strings.Join(census, "\n")

	// THE COUNT IS REAL, not "a lane tag is present". This is the whole difference between Find
	// and FindAll, and the reason wiring alone would not have fixed #873.
	if !strings.Contains(joined, "lane-attribution ×3") {
		t.Errorf("the census does not report the lane-tag COUNT:\n%s", joined)
	}
	if !strings.Contains(joined, "process-voice ×3") {
		t.Errorf("the census does not report the process-voice COUNT:\n%s", joined)
	}
	// AND THE PLACES, or an author has a number and nowhere to go.
	if !strings.Contains(joined, "at line") {
		t.Errorf("the census names no lines:\n%s", joined)
	}
	if !strings.Contains(joined, "line 3") {
		t.Errorf("the census does not name the first lane tag's line:\n%s", joined)
	}
}

// A clean base says nothing. An advisory that fires on ordinary subject prose is noise, and noise
// is how a real note stops being read.
func TestACleanBaseCarriesNoNote(t *testing.T) {
	if got := voiceCensus("91 is composite: 7 x 13. Euclid, Elements VII, Def. 11."); got != nil {
		t.Errorf("a clean base produced a census: %v", got)
	}
	if got := (ingestResult{Bytes: 42}).Human(); strings.Contains(got, "NOTE") {
		t.Errorf("a clean ingest carries an advisory:\n%s", got)
	}
}

// The confirmation names the file it froze. B7's said "the file removed" over a 40-byte stub its
// author had never written, and the author read it as a tool fault.
func TestTheConfirmationNamesTheFileItFroze(t *testing.T) {
	if got := (ingestResult{Path: "/run/blue/report.md", Bytes: 40}).Human(); !strings.Contains(got, "/run/blue/report.md frozen into the record (40 bytes)") {
		t.Errorf("the confirmation does not name the file and its size:\n%s", got)
	}
}

// FLAG, DO NOT BLOCK — and the assertion that matters is that the INGEST LANDS. Ingest is
// write-once and has already deleted the file by the time this renders, so a refusal here would
// strand a seat with no file and no base.
func TestTheIngestAdvisoryNotifiesAndDoesNotRefuse(t *testing.T) {
	h := ingestResult{Bytes: 16874, VoiceTells: []string{"lane-attribution ×6 at line 54, 70 — first is \"[minority: lane-3]\" — provenance is the record's"}}.Human()
	if !strings.Contains(h, "frozen into the record") {
		t.Errorf("the advisory replaced the confirmation — the base IS recorded and must say so:\n%s", h)
	}
	if !strings.Contains(h, "not a refusal") {
		t.Errorf("the note does not say it is advice; a seat will read it as a gate:\n%s", h)
	}
	// THE ROUTE OUT. `edit` is the only way to change a frozen base, and a note that does not
	// say so invites a re-ingest the record refuses.
	if !strings.Contains(h, "`edit`") {
		t.Errorf("the note does not name the only path that can act on it:\n%s", h)
	}
	if !strings.Contains(h, "cannot be re-ingested") {
		t.Errorf("the note does not say the base is frozen:\n%s", h)
	}
	// SEPARATION, NEVER DELETION — the same discipline the edit advisory carries. A seat told only
	// "this is wrong" deletes the sentence and loses the epistemic half with the operational one.
	if !strings.Contains(h, "re-voiced") {
		t.Errorf("the note does not say the conclusion-limiting half STAYS:\n%s", h)
	}
}

// A TRUNCATED LIST MUST SAY IT IS TRUNCATED. The cap keeps a badly-voiced base from burying the
// rest of the result; a cap that hid itself would be the same defect this package is about.
func TestTheCensusSaysWhenItCaps(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 20; i++ {
		b.WriteString("[minority: lane-1] a claim.\n")
	}
	joined := strings.Join(voiceCensus(b.String()), "\n")
	if !strings.Contains(joined, "×20") {
		t.Errorf("the count is capped as well as the list — it must be exact:\n%s", joined)
	}
	if !strings.Contains(joined, "more)") {
		t.Errorf("the list was capped and did not say so:\n%s", joined)
	}
}
