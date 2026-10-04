package report

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// mintT and closeT are a lens's mint and close as the record holds them, for a famT fixture: the
// write path refuses a close of a gap nobody minted, or one that carries no prose.
func mintT(t *testing.T, lens, gap string) *record.Event {
	return recordtest.Event(t, lens, &recordpb.Mint{GapId: proto.String(gap), Class: proto.String("x"), Problem: proto.String("p"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
}

func closeT(t *testing.T, lens, gap string) *record.Event {
	return recordtest.Event(t, lens, &recordpb.Close{GapId: proto.String(gap),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")})
}

// famT is fam with its events seeded through the write path, for a section that reads which sitting
// an act belongs to — blue's sittings, behind the manifest — and so needs the stored one.
func (b *boardT) famT(t *testing.T) record.Family {
	t.Helper()
	f := b.fam()
	return runtest.Family(t, f.Gaps, f.Events...)
}

// boardT keeps the retired Board's fixture SHAPE for these tests: gaps by id with a stated
// order, plus events. fam() is the one bridge to the family every converted reader takes —
// the fold is gone (plans/board-as-views.md wave 7), the fixtures merely kept their spelling.
type boardT struct {
	GapOrder []string
	Gaps     map[string]*record.Gap
	Events   []*record.Event
}

func (b *boardT) fam() record.Family {
	if b == nil {
		return record.NewFamily(nil, record.Merged{})
	}
	var ordered []*record.Gap
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g != nil && g.ID == "" {
			g.ID = id
		}
		ordered = append(ordered, g)
	}
	return record.NewFamily(ordered, record.Merged{Events: b.Events})
}

// famOf is the family of evs as the write path stores them: every epoch and `seat #N` a section
// prints is the stored window, so a section's fixture is seeded and loaded back.
func famOf(t *testing.T, evs []*record.Event) record.Family {
	t.Helper()
	return runtest.Family(t, nil, evs...)
}

// debateT is debate over evs as the write path stores them.
func debateT(t *testing.T, evs []*record.Event) string {
	t.Helper()
	f := famOf(t, evs)
	return debate(f)
}

// withdrawnClaimsT is withdrawnClaims over evs as the write path stores them.
func withdrawnClaimsT(t *testing.T, evs []*record.Event) string {
	t.Helper()
	f := famOf(t, evs)
	return withdrawnClaims(f)
}

// revisionHistoryT is revisionHistory over evs as the write path stores them.
func revisionHistoryT(t *testing.T, evs []*record.Event) string {
	t.Helper()
	f := famOf(t, evs)
	return revisionHistory(f)
}

// logSectionT is logSection over evs as the write path stores them.
func logSectionT(t *testing.T, evs []*record.Event) string {
	t.Helper()
	f := famOf(t, evs)
	return logSection(f)
}
