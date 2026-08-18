package record

import (
	"strings"
	"testing"
)

// A DUTY DISCHARGED BY NOTHING IS THE DUTY'S OWN DEFEAT.
//
// Measured before the fix, on a real run: `blue friction` then `blue revision`, no flags at all,
// took a seat from two outstanding duties to `complete: true`, and the friction projection then
// read `total: 1, attested: 0` with `text: ""`. A channel reporting one entry that says nothing
// is worse than one reporting none, because the count is what an audit reads.
//
// The class was found once and fixed at ONE instance — RequiredFields' note on `verify` records
// it ("A VERIFICATION OF NOTHING WAS RECORDABLE") — and the siblings were never swept.

func TestAnEmptyDischargeIsRefused(t *testing.T) {
	for _, tc := range []struct{ typ, want string }{
		{"friction", "--reason"},
		{"friction-none", "--reason"},
		{"position", "--reason"},
		{"revision", "--reason"},
	} {
		t.Run(tc.typ, func(t *testing.T) {
			err := validate(t.TempDir(), "blue-respond-r1", tc.typ, NewPayload())
			if err == nil {
				t.Fatalf("%s with no reason was accepted — it records an empty event and counts as discharged", tc.typ)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the refusal does not name %s:\n%v", tc.want, err)
			}
		})
	}
}

func TestARealDischargeIsAccepted(t *testing.T) {
	for _, typ := range []string{"friction", "friction-none", "position", "revision"} {
		p := NewPayload().Set("reason", "what I reached for and what happened")
		if err := validate(t.TempDir(), "blue-respond-r1", typ, p); err != nil {
			t.Errorf("%s with a reason was refused: %v", typ, err)
		}
	}
}

// THE EXPLICIT NEGATIVE NEEDS CONTENT TOO. `--none` is worth more than silence only when it says
// what was looked at; without that it is silence with an event attached.
func TestTheExplicitNegativeCannotBeEmpty(t *testing.T) {
	err := validate(t.TempDir(), "blue-respond-r1", "friction-none", NewPayload())
	if err == nil || !strings.Contains(err.Error(), "FOUND") {
		t.Errorf("friction --none with no reason should say what the negative is FOR: %v", err)
	}
}

// A RECEIPT NAMING NO GAP CANNOT BE AUDITED. requireGap returns nil on an empty id — an ABSENT
// reference is not a DANGLING one, which is right for the check it makes — so absence needed its
// own check, and without it a bare manifest-row printed "manifest row recorded for ".
func TestAReceiptMustNameItsGapAndSayWhatItChecked(t *testing.T) {
	if err := validate(t.TempDir(), "blue-respond-r1", "manifest-row", NewPayload()); err == nil {
		t.Error("a manifest row naming no gap was accepted")
	} else if !strings.Contains(err.Error(), "--id") {
		t.Errorf("the refusal does not name --id: %v", err)
	}
	// And with a gap but no row: the receipt is what makes "unaudited repair" countable, so a
	// blank one flatters the count it feeds.
	p := NewPayload().Set("gap_id", "R1-1")
	if err := validate(t.TempDir(), "blue-respond-r1", "manifest-row", p); err == nil {
		t.Error("a manifest row with no --row was accepted")
	}
}

// spot-check is DELIBERATELY not in this class: its bare form is a documented decision with its
// own test, and its --none --reason exists for the same distinction. Pinned so a later sweep of
// this class does not "fix" it by mistake.
func TestSpotCheckBareStaysAccepted(t *testing.T) {
	if err := validate(t.TempDir(), "red-merge-r1", "spot-check", NewPayload()); err != nil {
		t.Errorf("a bare spot-check was refused, but an honestly-empty round is a discharge: %v", err)
	}
}
