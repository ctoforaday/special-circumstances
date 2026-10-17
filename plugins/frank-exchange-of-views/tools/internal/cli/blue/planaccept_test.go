package blue

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
)

var planGap = anchor.ID("gap", [4]byte{0xa, 0xb, 0xc, 0xd})

const planReport = "# H\n\nFive independent verification approaches agree that 7 is prime.\n\nThe sieve is fast<!--fx:F-00abc123--> and simple.\n"

// planOver is PlanAccept over planReport with gap G's anchor placed at the end of quote, as `lens
// mint` asks it.
func planOver(t *testing.T, quote, fix string) (Accepted, error) {
	t.Helper()
	placed, err := anchortext.Attach(planReport, planGap, quote)
	if err != nil {
		return Accepted{}, err
	}
	return PlanAccept(placed, planGap, quote, fix)
}

// A concrete prescription states an exact, unique span, and the accept plans it — the forced re-read:
// it cannot be written from memory of what the report probably says.
func TestPlanAcceptPlansAnApplyableTextualFix(t *testing.T) {
	a, err := planOver(t, "Five independent verification approaches agree that 7 is prime", "Five verification approaches agree that 7 is prime")
	if err != nil {
		t.Fatalf("a legal textual prescription was refused: %v", err)
	}
	if !strings.Contains(a.Planned, "Five verification approaches agree that 7 is prime"+anchor.Token(planGap)+".") {
		t.Errorf("the planned report does not carry red's text with the gap's anchor at its end:\n%s", a.Planned)
	}
}

// A span the accept cannot locate is a span red did not read.
func TestPlanAcceptRefusesAQuoteThatIsNotThere(t *testing.T) {
	if _, err := PlanAccept(planReport, planGap, "text that was never in the report", "anything"); !errors.Is(err, bluedoc.ErrMisQuote) {
		t.Fatalf("a prescription against text the report does not hold = %v, want ErrMisQuote", err)
	}
}

// A prescription that drops an anchor inside its span is one blue could not apply: the quote is
// matched across the anchor, so the located span CONTAINS F-00abc123.
func TestPlanAcceptRefusesAnAnchorDrop(t *testing.T) {
	if _, err := planOver(t, "The sieve is fast and simple", "The sieve is quick and simple"); err == nil {
		t.Fatal("a prescription that drops a finding's anchor mid-span was planned")
	}
}

// An anchor CARRIED across is legal, for red exactly as for blue.
func TestPlanAcceptAllowsAnchorTransit(t *testing.T) {
	if _, err := planOver(t, "The sieve is fast and simple", "The sieve is quick<!--fx:F-00abc123--> and simple"); err != nil {
		t.Fatalf("carrying an anchor across a prescription was refused: %v", err)
	}
}

// A no-op is not a prescription.
func TestPlanAcceptRefusesAnIdenticalPair(t *testing.T) {
	if _, err := planOver(t, "The sieve is fast", "The sieve is fast"); err == nil {
		t.Fatal("a prescription that changes nothing was planned")
	}
}

// The anchors the tool puts back count: the anchored sentence survives word for word, and the
// prescription stands.
func TestPlanAcceptReadsTheAnchorsPutBack(t *testing.T) {
	if _, err := planOver(t, "The sieve is fast and simple", "It is old. The sieve is fast and simple"); err != nil {
		t.Fatalf("a prescription keeping the anchored sentence word for word was refused: %v", err)
	}
}
