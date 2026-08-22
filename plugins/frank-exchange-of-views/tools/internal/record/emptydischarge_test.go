package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
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
	// The EMPTY BODY of each type, which is what "discharged by nothing" means once the payload
	// map is gone: every prose field unset, and the verb still claiming the duty is met.
	for _, tc := range []struct {
		typ  recordpb.EventType
		body proto.Message
		want string
	}{
		{recordpb.EventType_EVENT_TYPE_FRICTION, &recordpb.Friction{}, "--reason"},
		{recordpb.EventType_EVENT_TYPE_FRICTION_NONE, &recordpb.FrictionNone{}, "--reason"},
		{recordpb.EventType_EVENT_TYPE_POSITION, &recordpb.Position{}, "--reason"},
		{recordpb.EventType_EVENT_TYPE_REVISION, &recordpb.Revision{}, "--reason"},
	} {
		t.Run(recordpb.Word(tc.typ), func(t *testing.T) {
			err := validate(t.TempDir(), "blue-respond-r1", tc.typ, tc.body)
			if err == nil {
				t.Fatalf("%s with no reason was accepted — it records an empty event and counts as discharged", recordpb.Word(tc.typ))
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the refusal does not name %s:\n%v", tc.want, err)
			}
		})
	}
}

func TestARealDischargeIsAccepted(t *testing.T) {
	// EACH TYPE'S OWN BODY. The earlier conversion passed a Revision for all four, so three of the
	// cases validated a message the verb does not write — the assertion held for the wrong reason.
	prose := proto.String("what I reached for and what happened")
	for _, c := range []struct {
		typ  recordpb.EventType
		body proto.Message
	}{
		{recordpb.EventType_EVENT_TYPE_FRICTION, &recordpb.Friction{Text: prose}},
		{recordpb.EventType_EVENT_TYPE_FRICTION_NONE, &recordpb.FrictionNone{Text: prose}},
		{recordpb.EventType_EVENT_TYPE_POSITION, &recordpb.Position{Text: prose}},
		{recordpb.EventType_EVENT_TYPE_REVISION, &recordpb.Revision{Text: prose}},
	} {
		if err := validate(t.TempDir(), "blue-respond-r1", c.typ, c.body); err != nil {
			t.Errorf("%s with a reason was refused: %v", recordpb.Word(c.typ), err)
		}
	}
}

// THE EXPLICIT NEGATIVE NEEDS CONTENT TOO. `--none` is worth more than silence only when it says
// what was looked at; without that it is silence with an event attached.
func TestTheExplicitNegativeCannotBeEmpty(t *testing.T) {
	err := validate(t.TempDir(), "blue-respond-r1", recordpb.EventType_EVENT_TYPE_FRICTION_NONE, &recordpb.FrictionNone{})
	if err == nil || !strings.Contains(err.Error(), "FOUND") {
		t.Errorf("friction --none with no reason should say what the negative is FOR: %v", err)
	}
}

// A RECEIPT NAMING NO GAP CANNOT BE AUDITED. requireGap returns nil on an empty id — an ABSENT
// reference is not a DANGLING one, which is right for the check it makes — so absence needed its
// own check, and without it a bare manifest-row printed "manifest row recorded for ".
func TestAReceiptMustNameItsGapAndSayWhatItChecked(t *testing.T) {
	if err := validate(t.TempDir(), "blue-respond-r1", recordpb.EventType_EVENT_TYPE_MANIFEST_ROW, &recordpb.ManifestRow{}); err == nil {
		t.Error("a manifest row naming no gap was accepted")
	} else if !strings.Contains(err.Error(), "--id") {
		t.Errorf("the refusal does not name --id: %v", err)
	}
	// And with a gap but no row: the receipt is what makes "unaudited repair" countable, so a
	// blank one flatters the count it feeds.
	if err := validate(t.TempDir(), "blue-respond-r1", recordpb.EventType_EVENT_TYPE_MANIFEST_ROW,
		&recordpb.ManifestRow{GapId: proto.String("R1-1")}); err == nil {
		t.Error("a manifest row with no --row was accepted")
	}
}

// spot-check is DELIBERATELY not in this class: its bare form is a documented decision with its
// own test, and its --none --reason exists for the same distinction. Pinned so a later sweep of
// this class does not "fix" it by mistake.
func TestSpotCheckBareStaysAccepted(t *testing.T) {
	if err := validate(t.TempDir(), "red-merge-r1", recordpb.EventType_EVENT_TYPE_SPOT_CHECK, &recordpb.SpotCheck{}); err != nil {
		t.Errorf("a bare spot-check was refused, but an honestly-empty round is a discharge: %v", err)
	}
}
