package record

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

// PointCounts is the board and the evidence layer in numbers, as the record stood at one point: the
// acts written at or before it, and no other.
type PointCounts struct {
	// GapsMinted is the gaps minted at the point; GapsClosed those whose first closing act is at or
	// before it. The rest were open there.
	GapsMinted, GapsClosed int
	// Citations and Computations are the anchors placed at the point, and the Retired ones those a
	// retire at or before it took out.
	Citations, CitationsRetired       int
	Computations, ComputationsRetired int
	// Readings is every `verify` at the point: of a citation, or of a source the reader found.
	Readings int
}

// GapsOpen is the gaps minted at the point and not closed there.
func (c PointCounts) GapsOpen() int { return c.GapsMinted - c.GapsClosed }

// CountsAt is the record's counts as it stood at the event whose events.id is point.
//
// A GAP IS OPEN AT A POINT UNTIL ITS FIRST CLOSING ACT (Gap.ClosedEvent, gap_state's closed_event) —
// the rule convergence_at judges a verdict by — so at the record's newest event the open and closed
// counts are the board's own. An anchor is retired at a point by the first retire naming it
// (evidence_anchor's retired_event), and a citation with no label is no anchor (EvidenceJSONOf).
//
// The family is the loaded one (FamilyOf): the point of each act is the window index's, and a
// family built by hand has none.
func (f Family) CountsAt(point int64) PointCounts {
	var c PointCounts
	minted, placed, retired := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, e := range f.Events {
		if f.At.Of(e).ID > point {
			continue
		}
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.Mint:
			minted[b.GetGapId()] = true
		case *recordpb.Cite:
			if b.GetLabel() != "" {
				placed[b.GetLabel()] = false
			}
		case *recordpb.Computation:
			placed[b.GetComputationId()] = true
		case *recordpb.Verify:
			c.Readings++
		case *recordpb.Retire:
			for _, a := range b.GetAnchors() {
				retired[a] = true
			}
		}
	}
	c.GapsMinted = len(minted)
	for _, g := range f.Gaps {
		if minted[g.ID] && g.ClosedEvent != 0 && g.ClosedEvent <= point {
			c.GapsClosed++
		}
	}
	for a, computation := range placed {
		switch {
		case computation:
			c.Computations++
			if retired[a] {
				c.ComputationsRetired++
			}
		default:
			c.Citations++
			if retired[a] {
				c.CitationsRetired++
			}
		}
	}
	return c
}
