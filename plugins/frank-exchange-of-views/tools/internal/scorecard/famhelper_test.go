package scorecard

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// famSeededT is famOfEventsT with the events seeded through the write path, for a metric that reads
// which sitting an act belongs to — blue's sittings, behind manifest coverage — and so needs the
// stored one. The gaps the fixtures name are minted ahead of them, so the record can hold their
// closes and receipts; no blue sitting reads a mint.
func famSeededT(t *testing.T, evs []*record.Event) *record.Family {
	t.Helper()
	var mints []*record.Event
	for _, g := range []string{"G1", "G2", "G3", "G4", "G5"} {
		mints = append(mints, recordtest.Event(t, "red-lens-logic", &recordpb.Mint{GapId: proto.String(g), Class: proto.String("x"),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}))
	}
	f := runtest.Family(t, nil, append(mints, evs...)...)
	return &f
}

// famOfEventsT builds an events-only family fixture; nil stays nil where a test means
// "record unreadable".
func famOfEventsT(evs []*record.Event) *record.Family {
	f := record.NewFamily(nil, evs)
	return &f
}
