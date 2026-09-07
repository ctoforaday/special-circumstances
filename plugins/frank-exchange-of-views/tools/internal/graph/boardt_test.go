package graph

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

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
		return record.NewFamily(nil, nil)
	}
	var ordered []*record.Gap
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g != nil && g.ID == "" {
			g.ID = id
		}
		ordered = append(ordered, g)
	}
	return record.NewFamily(ordered, b.Events)
}
