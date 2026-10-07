package anchortext

import (
	"strings"
	"testing"
)

// A LOCATION THAT CONTAINS A QUOTED PHRASE IS NOT A LOCATION FOR THAT PHRASE. #552: a long,
// exact, uniquely-matching quote anchored at an unrelated SHORTER sentence holding its quoted
// phrase, silently, on the surface whose whole job is to say where the evidence is. The quote is
// matched as written, so the anchor lands on the sentence it names.
func TestAQuotedPhraseInsideALocationDoesNotStealTheAnchor(t *testing.T) {
	const report = "# H\n\nThe cost is climbing sharply.\n\nBlue wrote that the cost is \"climbing sharply\" and gave no source for it.\n"
	const location = `Blue wrote that the cost is "climbing sharply" and gave no source for it.`

	got, err := Attach(report, "f-1", location)
	if err != nil {
		t.Fatalf("a location present in the report verbatim was refused: %v", err)
	}
	if !strings.Contains(got, `no source for it<!--fx:f-1-->.`) || strings.Contains(got, `climbing sharply<!--fx:f-1-->.`) {
		t.Errorf("the anchor did not land on the sentence the location names:\n%s", got)
	}
}

// A LABELLED LOCATION IS NOT A QUOTE. `§ Foundations: "The scheduler is preemptive"` is not in
// the report, so it places nothing, at the write and at replay alike: an anchor sits where its
// quote is, and nothing reads a quote out of a decoration.
func TestALabelledLocationIsRefused(t *testing.T) {
	const report = "# H\n\n## Foundations\n\nThe scheduler is preemptive.\n"
	const labelled = `§ Foundations: "The scheduler is preemptive"`
	if _, err := Attach(report, "c-1", labelled); err != ErrMisQuote {
		t.Errorf("Attach(labelled) = %v, want ErrMisQuote", err)
	}
	if _, err := InsertAnchor([]byte(report), labelled, "<!--cite:c-1-->"); err != ErrMisQuote {
		t.Errorf("InsertAnchor(labelled) = %v, want ErrMisQuote", err)
	}
}

// A LOCATION THAT IS NOWHERE IS REFUSED: a marker on absent content is the failure the mis-quote
// refusal exists for.
func TestALocationThatIsNotThereIsStillRefused(t *testing.T) {
	const report = "# H\n\nThe cost is climbing sharply.\n"
	if _, err := Attach(report, "f-2", `Nothing in this document says "anything like this".`); err == nil {
		t.Fatal("a location absent from the report was anchored anyway")
	}
}
