package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
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
		seat string
		typ  recordpb.EventType
		body proto.Message
		want string
	}{
		// TYPED BUT WORDLESS. The type is required too, so an entirely empty body is refused for
		// the type before the prose is ever reached; this case is about the PROSE, so it supplies
		// a type and leaves the words out.
		{"blue-respond", recordpb.EventType_EVENT_TYPE_LOG, &recordpb.Log{Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum()}, "--reason"},
		{"blue-respond", recordpb.EventType_EVENT_TYPE_LOG, &recordpb.Log{Type: recordpb.LogType_LOG_TYPE_REQUEST.Enum()}, "--reason"},
		{"red-chair", recordpb.EventType_EVENT_TYPE_POSITION, &recordpb.Position{}, "--reason"},
	} {
		t.Run(recordpb.Word(tc.typ), func(t *testing.T) {
			err := validateRun(mustRun(t, recordtest.TmpRun(t)), tc.seat, tc.typ, tc.body)
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
	// EACH TYPE'S OWN BODY, so each case validates the message its verb writes.
	prose := proto.String("what I reached for and what happened")
	for _, c := range []struct {
		seat string
		typ  recordpb.EventType
		body proto.Message
	}{
		{"blue-respond", recordpb.EventType_EVENT_TYPE_LOG, &recordpb.Log{Text: prose, Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum()}},
		{"blue-respond", recordpb.EventType_EVENT_TYPE_LOG, &recordpb.Log{Text: prose, Type: recordpb.LogType_LOG_TYPE_REQUEST.Enum()}},
		{"red-chair", recordpb.EventType_EVENT_TYPE_POSITION, &recordpb.Position{Text: prose}},
	} {
		if err := validateRun(mustRun(t, recordtest.TmpRun(t)), c.seat, c.typ, c.body); err != nil {
			t.Errorf("%s with a reason was refused: %v", recordpb.Word(c.typ), err)
		}
	}
}

// A RECEIPT NAMING NO GAP CANNOT BE AUDITED. requireGap returns nil on an empty id — an ABSENT
// reference is not a DANGLING one, which is right for the check it makes — so absence needed its
// own check, and without it a bare manifest-row printed "manifest row recorded for ".
func TestAReceiptMustNameItsGapAndSayWhatItChecked(t *testing.T) {
	if err := validateRun(mustRun(t, recordtest.TmpRun(t)), "blue-respond", recordpb.EventType_EVENT_TYPE_MANIFEST_ROW, &recordpb.ManifestRow{}); err == nil {
		t.Error("a manifest row naming no gap was accepted")
	} else if !strings.Contains(err.Error(), "--id") {
		t.Errorf("the refusal does not name --id: %v", err)
	}
	// And with a gap but no row: the receipt is what makes "unaudited repair" countable, so a
	// blank one flatters the count it feeds.
	if err := validateRun(mustRun(t, recordtest.TmpRun(t)), "blue-respond", recordpb.EventType_EVENT_TYPE_MANIFEST_ROW,
		&recordpb.ManifestRow{GapId: proto.String("G1")}); err == nil {
		t.Error("a manifest row with no --row was accepted")
	}
}

// spot-check is DELIBERATELY not in this class: its bare form is a documented decision with its
// own test, and its --reason says why it named nothing. Pinned so a later sweep of
// this class does not "fix" it by mistake.
func TestSpotCheckBareStaysAccepted(t *testing.T) {
	if err := validateRun(mustRun(t, recordtest.TmpRun(t)), "red-chair", recordpb.EventType_EVENT_TYPE_SPOT_CHECK, &recordpb.SpotCheck{}); err != nil {
		t.Errorf("a bare spot-check was refused, but an honestly-empty round is a discharge: %v", err)
	}
}
