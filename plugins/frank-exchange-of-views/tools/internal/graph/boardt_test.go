package graph

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// boardT keeps the retired Board's fixture SHAPE for these tests: gaps by id with a stated
// order, plus events. fam() is the one bridge to the family every converted reader takes —
// the fold is gone (plans/board-as-views.md wave 7), the fixtures merely kept their spelling.
type boardT struct {
	GapOrder []string
	Gaps     map[string]*record.Gap
	Events   []*record.Event
}

// fam is the family with its events seeded through the write path and loaded back, so every epoch
// and sitting a reader takes off it is the one the record stores (record.WindowIndex). Each gap in
// GapOrder is minted ahead of the events, by the harness, so the record can hold the motions and
// closes that name it; the gaps themselves are the fixture's, as given.
func (b *boardT) fam(t *testing.T) record.Family {
	t.Helper()
	if b == nil {
		return record.NewFamily(nil, nil)
	}
	var ordered []*record.Gap
	var evs []*record.Event
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g != nil && g.ID == "" {
			g.ID = id
		}
		ordered = append(ordered, g)
		evs = append(evs, recordtest.Event(t, record.HarnessSeat, &recordpb.Mint{GapId: proto.String(id), Class: proto.String("x"),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}))
	}
	f := runtest.Family(t, ordered, append(evs, b.Events...)...)
	// The minted gaps are scaffolding for the record's foreign keys, not acts a check reads.
	f.Events = f.Events[len(evs):]
	return f
}
