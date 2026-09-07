package anchortext

import (
	"strings"
	"testing"
)

// A LOCATION THAT CONTAINS A QUOTED PHRASE IS NOT A LOCATION FOR THAT PHRASE.
//
// extractQuote prefers the text between the first two double quotes, which is what makes a
// decorated location — `§ Foundations: "the scheduler is preemptive"` — locatable. Applied
// unconditionally it also fires on a location that IS the sentence and merely CONTAINS a quoted
// phrase, and the anchor then lands on that phrase wherever else it occurs:
//
//	The cost is climbing sharply<!--fx:f-1-->.   <- a different paragraph entirely
//
// Silently, exit 0, on the surface whose whole job is to say where the evidence is. #552 measured
// it from the other side: a long, exact, uniquely-matching quote anchored at an unrelated SHORTER
// sentence while a shorter quote from the same target sentence anchored correctly.
func TestAQuotedPhraseInsideALocationDoesNotStealTheAnchor(t *testing.T) {
	const report = "# H\n\nThe cost is climbing sharply.\n\nBlue wrote that the cost is \"climbing sharply\" and gave no source for it.\n"
	const location = `Blue wrote that the cost is "climbing sharply" and gave no source for it.`

	out, err := InsertAnchor([]byte(report), location, "<!--fx:f-1-->")
	if err != nil {
		t.Fatalf("a location present in the report verbatim was refused: %v", err)
	}
	got := string(out)
	// The marker belongs at the end of the sentence the location NAMES, which is the last one.
	if !strings.Contains(got, `no source for it<!--fx:f-1-->.`) {
		t.Errorf("the anchor did not land on the sentence the location names:\n%s", got)
	}
	if strings.Contains(got, `climbing sharply<!--fx:f-1-->.`) {
		t.Errorf("the anchor landed on the INNER phrase, in a different paragraph:\n%s", got)
	}
}

// THE DECORATED FORM STILL WORKS, and this is the assertion that keeps the fix from being a
// revert: a section label followed by its quoted sentence is not in the report literally, so the
// extracted span is what locates it.
func TestALabelledLocationStillFindsItsQuotedSentence(t *testing.T) {
	const report = "# H\n\n## Foundations\n\nThe scheduler is preemptive.\n"
	out, err := InsertAnchor([]byte(report), `§ Foundations: "The scheduler is preemptive"`, "<!--cite:c-1-->")
	if err != nil {
		t.Fatalf("the labelled form must still locate its sentence: %v", err)
	}
	if !strings.Contains(string(out), "The scheduler is preemptive<!--cite:c-1-->.") {
		t.Errorf("the labelled form did not anchor its sentence:\n%s", out)
	}
}

// A LOCATION THAT IS NOWHERE IS STILL REFUSED. The fallback must not become a way to anchor
// something the report does not contain — a marker on absent content is the failure the
// mis-quote refusal exists for.
func TestALocationThatIsNotThereIsStillRefused(t *testing.T) {
	const report = "# H\n\nThe cost is climbing sharply.\n"
	if _, err := InsertAnchor([]byte(report), `Nothing in this document says "anything like this".`, "<!--fx:f-2-->"); err == nil {
		t.Fatal("a location absent from the report was anchored anyway")
	}
}
