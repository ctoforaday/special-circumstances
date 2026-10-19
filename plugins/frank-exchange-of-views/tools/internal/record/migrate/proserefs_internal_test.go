package migrate

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN ID THE PROSE NAMES AND THE RECORD DOES NOT HOLD STAYS PROSE. The write refuses a field naming
// an id no act placed, so the translation keeps, from what the archived text names, each id the
// migrated record holds at that act's place — once each, in order of first mention.
func TestOnlyHeldProseIDsBecomeFields(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, recordtest.At(t, "red-lens-evidence", "m", &recordpb.Mint{
		GapId: proto.String("G-0000000a"), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"), Class: proto.String("self-attestation"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:  recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}))
	record.Migrating = true
	defer func() { record.Migrating = false }()
	b, err := record.BeginBatch(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = b.Rollback() }()
	d := &recordpb.Declare{Holding: proto.String("G-0000000a holds; G-0000ffff is cited; G-0000000a again")}
	if err := proseRefs(nil, b.Querier(), d); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(d.GetGapIds(), []string{"G-0000000a"}) {
		t.Errorf("gap_ids = %v, want the one held gap, once", d.GetGapIds())
	}
}
