package blue

import (
	"strings"
	"testing"
)

// FLAG FOR RED, DO NOT BLOCK — and the assertion that matters is that the edit LANDS.
//
// A gate here would be the wrong instrument twice over. A pattern cannot tell a report narrating
// its own construction from a report quoting a source that narrates something, so a refusal would
// silently cost the second one; and the residue is frequently load-bearing — "Savage is known only
// through the interested party's summary" is a limit on the CONCLUSION and must survive, while
// "after four hosts refused this container" is a fact about the run and must move. Separation is a
// judgement, and red's voice lens owns it. This only says where to look, while the seat is present.
func TestTheVoiceAdvisoryNotifiesAndDoesNotRefuse(t *testing.T) {
	r := editResult{VoiceTells: []string{`"this run" reads as process-voice — the record already holds the run`}}
	h := r.Human()
	if !strings.Contains(h, "blue edit recorded") {
		t.Errorf("the advisory replaced the confirmation — the edit is recorded and must say so:\n%s", h)
	}
	if !strings.Contains(h, "not a refusal") {
		t.Errorf("the note does not say it is advice; a seat reading it will treat it as a gate:\n%s", h)
	}
	if !strings.Contains(h, "process-voice") {
		t.Errorf("the note does not name what it caught:\n%s", h)
	}
	// SEPARATION, NEVER DELETION. A seat told only "this is wrong" deletes the sentence, which
	// loses the epistemic half with the operational one.
	if !strings.Contains(h, "re-voiced") {
		t.Errorf("the note does not say the conclusion-limiting half STAYS:\n%s", h)
	}
}

// A clean edit says nothing extra. An advisory that fires on ordinary subject prose is noise, and
// noise is how a real note stops being read.
func TestACleanEditCarriesNoNote(t *testing.T) {
	if got := (editResult{}).Human(); strings.Contains(got, "NOTE") {
		t.Errorf("a clean edit carries an advisory: %s", got)
	}
}
