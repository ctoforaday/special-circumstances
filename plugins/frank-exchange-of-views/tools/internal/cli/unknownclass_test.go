package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A WRONG CLASS WORD IS REFUSED WITH THE NEAR MISS AND THE WORDS ONLY (plans/gap-class-enum.md D9,
// F7). `report-voice` is the archive's spelling of `report_voice`; the refusal names that near miss,
// lists every GapClass word, and carries none of the values' definitions or distinguishers, which
// `mint --help` prints. A refusal that appended the full menu buries the near miss under ~13 KB.
func TestAWrongClassIsRefusedWithTheNearMissAndTheWordsOnly(t *testing.T) {
	runDir := newRun(t)
	registerLensOnce(t, runDir)
	mintArgs := []string{"mint", "--run", runDir, "--seat-id", lensSeat,
		"--class", "report-voice", "--check-kind", "document", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p", "--about-kind", "section", "--about", "Findings"}

	_, err := run(t, mintArgs...)
	if err == nil {
		t.Fatal("a mint with --class report-voice was accepted")
	}
	msg := err.Error()
	if !strings.Contains(msg, `it is "report_voice" with different punctuation`) {
		t.Errorf("the refusal must name report_voice as the near miss:\n%s", msg)
	}
	ed := recordpb.GapClass(0).Descriptor()
	for _, w := range recordpb.Words(ed) {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal must list %q among the words a --class takes:\n%s", w, msg)
		}
	}
	for i := 0; i < ed.Values().Len(); i++ {
		v := ed.Values().Get(i)
		if v.Number() == 0 {
			continue
		}
		means, err := recordpb.EnumValueDoc(v)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(msg, means) {
			t.Errorf("the refusal carries %s's definition, which belongs in mint --help:\n%s", recordpb.Spelling(v), msg)
		}
	}
	if strings.Contains(msg, "nearest:") {
		t.Errorf("the refusal carries a neighbour line, which belongs in mint --help:\n%s", msg)
	}

	// The near miss, corrected, lands.
	mintArgs[6] = "report_voice"
	if _, err := run(t, mintArgs...); err != nil {
		t.Fatalf("the near miss, spelled as the refusal named it, was refused: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Mint{}).GetClass(); got != recordpb.GapClass_GAP_CLASS_REPORT_VOICE {
		t.Errorf("minted class = %v, want report_voice", got)
	}
}

// AN EMPTY --class IS REFUSED, and the refusal names the catch-all a seat reaches for when no class
// fits.
func TestAnEmptyClassIsRefusedNamingTheCatchAll(t *testing.T) {
	runDir := newRun(t)
	registerLensOnce(t, runDir)
	_, err := run(t, "mint", "--about-kind", "section", "--about", "Findings", "--run", runDir, "--seat-id", lensSeat,
		"--class", "", "--check-kind", "document", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err == nil {
		t.Fatal("a mint with an empty --class was accepted")
	}
	if !strings.Contains(err.Error(), "unclassified") {
		t.Errorf("an empty --class must be refused naming the catch-all, got: %v", err)
	}
}
