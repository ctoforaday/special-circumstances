package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BIBLIOGRAPHY PRINTS SEAT TEXT, SO PROVE AND CITE MEET THE VOICE ADVISORY.
//
// The census that put edit, ingest and the line-of-inquiry verbs under the advisory exempted cite
// and proof as "landing only in the Bibliography". The 2026-09-11 is-91-prime run's lane drafts and
// frozen base carried no tell; its report.md carried four, every one in a [^PN] footnote — a proof's
// --reason, printed word for word. A cite's --title is printed the same way, as its entry. These
// drive the real verbs, because a guard that exists and is never called survives any test of the
// helper.

const voicedProofNote = "Recomputed trial division from scratch this sitting rather than trusting the lane draft"

func TestProveAdvisesOnAVoicedNote(t *testing.T) {
	runDir := newRun(t)
	seat := proveSeat(t, runDir, "# H\n\nNine is composite, so the protocol rejects a false claim.\n")
	s := script(t, runDir, "nine.js", "console.log('divisors: 3');")
	args := []string{"prove", "--run", runDir, "--seat-id", seat, "--key", "P1",
		"--quote", "Nine is composite, so the protocol rejects a false claim.",
		"--script", s, "--reason", voicedProofNote}

	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("the advisory refused a proof — it must refuse nothing: %v\n%s", err, out)
	}
	for _, want := range []string{"proof p-", "process-voice", "not a refusal", "footnote", "re-voiced"} {
		if !strings.Contains(out, want) {
			t.Errorf("the proof's result does not carry %q:\n%s", want, out)
		}
	}
	// ADVICE, NOT A REWRITE: the record keeps exactly what the seat wrote.
	if got := lastBody(t, runDir, &recordpb.Proof{}).GetText(); got != voicedProofNote {
		t.Errorf("the advisory altered the recorded note: got %q", got)
	}
	// A crash-retry hears the same note, not a fast path that drops it.
	out, err = run(t, args...)
	if err != nil {
		t.Fatalf("retry: %v\n%s", err, out)
	}
	if !strings.Contains(out, "idempotent") || !strings.Contains(out, "process-voice") {
		t.Errorf("the idempotent retry lost the advisory:\n%s", out)
	}
}

func TestACleanProofNoteCarriesNoNote(t *testing.T) {
	runDir := newRun(t)
	seat := proveSeat(t, runDir, "# H\n\nNine is composite, so the protocol rejects a false claim.\n")
	s := script(t, runDir, "nine.js", "console.log('divisors: 3');")
	out, err := run(t, "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "Nine is composite, so the protocol rejects a false claim.",
		"--script", s, "--reason", "a reproducible computation shows 3 divides 9")
	if err != nil {
		t.Fatalf("prove: %v\n%s", err, out)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("a clean proof note carries an advisory:\n%s", out)
	}
}

func TestCiteAdvisesOnAVoicedTitle(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/1": []byte("<html>the sky is blue</html>")}})

	title := "Sky Facts, re-read at the leaf this run"
	out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", "The sky is blue and the grass is green.", "--url", "https://sky/1", "--title", title)
	if err != nil {
		t.Fatalf("the advisory refused a citation — it must refuse nothing: %v\n%s", err, out)
	}
	for _, want := range []string{"citation recorded", "process-voice", "not a refusal", "Bibliography"} {
		if !strings.Contains(out, want) {
			t.Errorf("the citation's result does not carry %q:\n%s", want, out)
		}
	}
	if got := firstCiteEvent(t, runDir).GetTitle(); got != title {
		t.Errorf("the advisory altered the recorded title: got %q", got)
	}
}

func TestACleanCiteTitleCarriesNoNote(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/1": []byte("<html>the sky is blue</html>")}})

	out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", "The sky is blue and the grass is green.", "--url", "https://sky/1", "--title", "Sky Facts, chapter 2")
	if err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("a clean title carries an advisory:\n%s", out)
	}
}
