package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// --- punctuation-only repairs and no-op edits, through the real verbs --------------------------
//
// A quote's trailing punctuation is trimmed before the span is located. Measured in #861's arm-B
// rerun: blue quoted `on their own.).` to drop the final period, the span became itself, and the
// verb said "recorded" over a byte-identical report. At the END of a document there is no text
// after the terminator to quote through, so the only route is the literal span — taken, recorded as
// exact_span, and reproduced by replay.

const strayPeriod = "on their own.)."

func editStrayPeriod(t *testing.T, runDir, key string) error {
	t.Helper()
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", key,
		"--quote", strayPeriod, "--new", "on their own.)", "--reason", "drop the stray period after the parenthesis")
	return err
}

func TestAPunctuationRepairAtTheEndOfTheDocumentLands(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nIntro.\n\n(They stand "+strayPeriod+"\n")
	registerBlue(t, runDir)
	if err := editStrayPeriod(t, runDir, "E1"); err != nil {
		t.Fatalf("the repair was refused: %v", err)
	}
	// readReport IS the replay: the report exists only as base + recorded events, rendered.
	got := readReport(t, runDir)
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "(They stand on their own.)") {
		t.Errorf("the stray period is still there, or more moved than it:\n%q", got)
	}
	if !lastBody(t, runDir, &recordpb.BlueEdit{}).GetExactSpan() {
		t.Error("the edit took the literal span but did not record exact_span, so replay cannot know which span it replaced")
	}
}

func TestAPunctuationRepairMidDocumentLands(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\n(They stand "+strayPeriod+"\n\nNext.\n")
	registerBlue(t, runDir)
	if err := editStrayPeriod(t, runDir, "E1"); err != nil {
		t.Fatalf("the repair was refused: %v", err)
	}
	if got := readReport(t, runDir); !strings.Contains(got, "(They stand on their own.)\n\nNext.") {
		t.Errorf("the mid-document repair did not land: %q", got)
	}
}

// A NO-OP THE LITERAL SPAN CANNOT RESCUE IS REFUSED, records nothing, and does not blame the trim.
func TestAWhitespaceOnlyEditIsRefusedAsChangingNothing(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe claim holds.\n")
	registerBlue(t, runDir)
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
		"--quote", "The  claim holds", "--new", "The claim holds", "--reason", "spacing")
	if err == nil {
		t.Fatal("an edit that changes nothing was recorded")
	}
	if !strings.Contains(err.Error(), "changes nothing") || strings.Contains(err.Error(), "trimmed") {
		t.Errorf("want the plain no-change refusal, with no word about the punctuation trim: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 0 {
		t.Errorf("a refused edit still recorded %d blue_edit event(s)", n)
	}
}

// RED MAY NOT VERIFY A FIX THAT CHANGES NOTHING. The pair is distinct, present and legal; located,
// it is a no-op — its trailing punctuation is not the report's, so the literal span cannot stand in.
func TestMintRefusesAPrescriptionThatChangesNothing(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nIntro.\n\n(They stand "+strayPeriod+"\n")
	mintGap(t, runDir, "G0", "overclaim")
	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT)
	_, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "G1", "--class", "overclaim", "--problem", "stray period",
		"--fix", "drop it", "--check-kind", "document", "--check", "no stray period",
		"--severity", "low", "--likelihood", "low", "--impact", "low",
		"--quote", "on their own.);", "--new", "on their own.)")
	if err == nil {
		t.Fatal("a prescription that changes nothing was minted with a verified basis")
	}
	if !strings.Contains(err.Error(), "changes nothing") {
		t.Errorf("the refusal does not say the fix changes nothing: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != before {
		t.Errorf("a refused mint still landed (%d -> %d)", before, n)
	}
}

// THE SAME PLANNER DECIDES FOR RED AND FOR BLUE: a punctuation repair the literal span applies is
// verified at mint, and --accept applies it on that span. Then, once blue has made the change by
// hand, the same prescription changes nothing — and --accept says THAT, not that it is stale.
func TestAcceptOnAPrescriptionThatChangesNothingSaysSo(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nIntro.\n\n(They stand "+strayPeriod+"\n")
	mintGap(t, runDir, "G0", "overclaim")
	gap := mintWithProposal(t, runDir, "G1", strayPeriod, "on their own.)")
	registerBlue(t, runDir)

	// Blue makes the repair itself first; red's recorded fix now has nothing left to do.
	if err := editStrayPeriod(t, runDir, "E0"); err != nil {
		t.Fatalf("blue's own repair: %v", err)
	}
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "A1", "--answers", gap, "--accept", "--reason", "accepting")
	if err == nil {
		t.Fatal("accepted a prescription that changes nothing")
	}
	if !strings.Contains(err.Error(), "changes nothing") || strings.Contains(err.Error(), "stale") {
		t.Errorf("the refusal must say the prescription changes nothing, and must not call it stale: %v", err)
	}
}

func TestAcceptAppliesAPunctuationRepairOnTheLiteralSpan(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nIntro.\n\n(They stand "+strayPeriod+"\n")
	mintGap(t, runDir, "G0", "overclaim")
	gap := mintWithProposal(t, runDir, "G1", strayPeriod, "on their own.)")
	registerBlue(t, runDir)
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "A1", "--answers", gap, "--accept", "--reason", "red is right"); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if got := readReport(t, runDir); strings.Contains(got, strayPeriod) {
		t.Errorf("the accepted repair did not land: %q", got)
	}
	if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetExactSpan() || !b.GetAccepted() {
		t.Errorf("exact_span=%v accepted=%v, want both", b.GetExactSpan(), b.GetAccepted())
	}
}
