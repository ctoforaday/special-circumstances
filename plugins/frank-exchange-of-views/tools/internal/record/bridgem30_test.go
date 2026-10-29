package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// M30'S BRIDGE OVER AN INVERTED TERM HAS NO HOME NOW. The evidence lens recorded
// supports_with_bridge on C-4f4b8cfa at events 142 and 411 while its own reason said the source
// names the base the liar and the report names the composite — none of the three bridges. Written
// again, live, at the place it was written, each is refused for naming no bridge; the write that
// would carry one must say which of the three it was.
func TestM30BridgesAreRefusedWithoutAKind(t *testing.T) {
	whole := archivedEvents(t, m30Archive)
	for _, n := range []int{142, 411} {
		v := whole[n-1].GetVerify()
		if v.GetAnchor() != "C-4f4b8cfa" || v.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE || v.Bridge != nil {
			t.Fatalf("event %d is not the evidence lens's kindless supports_with_bridge on C-4f4b8cfa: %v", n, whole[n-1])
		}
		if !strings.Contains(v.GetText(), "liar") {
			t.Fatalf("event %d's reason does not describe the inverted term: %q", n, v.GetText())
		}
		run := archiveAt(t, m30Archive, whole, n-1)
		check := func(b *recordpb.Verify) error {
			return validateRun(run, whole[n-1].GetSeatId(), recordpb.EventType_EVENT_TYPE_VERIFY, b)
		}
		again := proto.Clone(v).(*recordpb.Verify)
		if err := check(again); err == nil || !strings.Contains(err.Error(), "supports_with_bridge requires --bridge") {
			t.Errorf("event %d written live: %v, want the refusal naming --bridge", n, err)
		}
		again.Bridge = recordpb.BridgeKind_BRIDGE_KIND_NEAR_RESTATEMENT.Enum()
		if err := check(again); err != nil {
			t.Errorf("event %d with a bridge named: %v, want it admitted", n, err)
		}
	}
}

// THE BRIDGE IS THE OUTCOME'S AND NO OTHER'S, live and under Migrating alike; only its ABSENCE on
// supports_with_bridge is admitted under Migrating, where an archive before the field holds it.
func TestABridgeBelongsToSupportsWithBridgeAlone(t *testing.T) {
	v := func(o recordpb.SourceOutcome, k *recordpb.BridgeKind) *recordpb.Verify {
		return &recordpb.Verify{CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(), Claim: proto.String("c"), Anchor: proto.String("C-00000001"), Outcome: o.Enum(), Bridge: k,
			Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("t")}
	}
	run := mustRun(t, t.TempDir())
	validate := func(b *recordpb.Verify) error {
		return validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_VERIFY, b)
	}
	defer func() { Migrating = false }()
	swb := recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE
	summary := recordpb.BridgeKind_BRIDGE_KIND_SUMMARY.Enum()
	for _, migrating := range []bool{false, true} {
		Migrating = migrating
		if err := validate(v(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS, summary)); err == nil || !strings.Contains(err.Error(), "drop --bridge") {
			t.Errorf("migrating=%v: a bridge on supports: %v, want it refused", migrating, err)
		}
		if err := validate(v(swb, recordpb.BridgeKind_BRIDGE_KIND_UNSPECIFIED.Enum())); err == nil {
			t.Errorf("migrating=%v: an unspecified bridge was admitted", migrating)
		}
		if err := validate(v(swb, summary)); err != nil {
			t.Errorf("migrating=%v: a named bridge: %v", migrating, err)
		}
	}
	Migrating = true
	if err := validate(v(swb, nil)); err != nil {
		t.Errorf("under Migrating, an archived bridge with no kind: %v, want it carried", err)
	}
	Migrating = false
	if err := validate(v(swb, nil)); err == nil {
		t.Error("live, supports_with_bridge with no kind was admitted")
	}
}
