package migrate

import (
	"fmt"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// GapPlacement is what bringing an archived run's gaps onto anchors did (F-c): each rewrite counted
// by its shape, and by name the gaps with no place in the report — a quote that never placed, or an
// anchor an edit's replacement holds no prose sentence to carry — and those placed by the fallback.
// The shapes are named where each is counted.
type GapPlacement struct {
	Shapes      map[string]int `json:"shapes,omitempty"`
	NeverPlaced []string       `json:"never_placed,omitempty"`
	Fallback    []string       `json:"fallback,omitempty"`
}

// placer runs in Replay after remap.apply and before Append, on bodies already in the destination's
// spelling, adding only what the source lacks: a gap's Anchor where the source never anchored it,
// gap anchors into an edit's old and new where they now stand, and a placement's quote where only
// its quoted span locates. With gap anchors removed, every rewritten old and new is the archived
// bytes, so the render's prose is the archived render's.
type placer struct {
	dst      record.Run
	anchored map[string]bool // ids the source stream anchors itself
	placed   bool
	census   GapPlacement
}

func newPlacer(dst record.Run, evs []OldEvent) *placer {
	p := &placer{dst: dst, anchored: map[string]bool{}, census: GapPlacement{Shapes: map[string]int{}}}
	for _, ev := range evs {
		if id, _ := ev.Fields["id"].(string); ev.Word == "anchor" && id != "" {
			p.anchored[id] = true
		}
	}
	return p
}

// step rewrites body for the report as the destination now renders it and returns the Anchor to
// append after it, or nil. A correction's replacement re-carries what its act placed and adds none.
func (p *placer) step(body proto.Message, correcting bool) (proto.Message, error) {
	switch b := body.(type) {
	case *recordpb.Mint:
		if correcting || b.GetLocation() == "" || p.anchored[b.GetGapId()] {
			return nil, nil
		}
		text, err := reportproj.RenderFromRecord(p.dst)
		if err != nil {
			return nil, nil // no base: the gap reads unrendered
		}
		if _, err := anchortext.Attach(text, b.GetGapId(), b.GetLocation()); err != nil {
			p.census.NeverPlaced = append(p.census.NeverPlaced, b.GetGapId())
			return nil, nil
		}
		p.placed = true
		p.census.Shapes["mint"]++
		return &recordpb.Anchor{Id: proto.String(b.GetGapId()), Location: proto.String(b.GetLocation())}, nil
	case *recordpb.BlueEdit:
		if p.placed {
			return nil, p.edit(b)
		}
	case *recordpb.Cite:
		b.Location = p.quote(b.Location)
	case *recordpb.Proof:
		b.Location = p.quote(b.Location)
	case *recordpb.Anchor:
		b.Location = p.quote(b.Location)
	case *recordpb.Verify:
		if b.GetLabel() != "" {
			b.Claim = p.quote(b.Claim)
		}
	}
	return nil, nil
}

// edit gives an archived edit every gap anchor that now stands in or against its span, so it
// renders, carries them and records the gaps whose sentence it changed.
func (p *placer) edit(b *recordpb.BlueEdit) error {
	text, err := reportproj.RenderFromRecord(p.dst)
	if err != nil {
		return nil
	}
	old := b.GetOld()
	start, end, ok := locate(text, old, b.GetExactSpan())
	if !ok {
		if old, ok = rewriteOld(text, old, b.GetExactSpan()); ok {
			start, end, ok = locate(text, old, b.GetExactSpan())
		}
		if !ok {
			return fmt.Errorf("migrate: this edit no longer locates in the report its gaps' anchors now stand in — a kept husk or a gap anchor inside its span that no rewrite reaches: %q", b.GetOld())
		}
		b.Old = proto.String(old)
		shape := "abutting" // (a): old given the anchor run that now abuts its span
		if b.GetExactSpan() {
			shape = "literal" // (b): old given the gap anchors now inside it
		}
		p.census.Shapes[shape]++
	}
	b.New = proto.String(p.carry(text[start:end], b.GetNew()))
	after := reportproj.ApplySplice(text, start, end, b.GetNew())
	for _, id := range bluedoc.ReopenedAnchors(text, after) {
		if isGap(id) && !slices.Contains(b.Reopened, id) {
			b.Reopened = append(b.Reopened, id)
			p.census.Shapes["reopened"]++
		}
	}
	return nil
}

// locate is the span replay takes for old.
func locate(text, old string, exact bool) (int, int, bool) {
	if exact {
		s, e, _, ok := bluedoc.LocateLiteral(text, old)
		return s, e, ok
	}
	s, e, err := bluedoc.LocateUniqueReplacing("render", text, old)
	return s, e, err == nil
}

// rewriteOld is old as replay can locate it now that gap anchors stand in the report: an exact
// span's bytes over the one place it occurs with gap anchors skipped (b); otherwise the anchor run
// abutting its span, put in place of old's own run after its last content character (a).
func rewriteOld(text, old string, exact bool) (string, bool) {
	if exact {
		gaps := map[int]int{} // each gap anchor's start in text, to its end
		anchor.Each(text, func(s, e int, id string) {
			if isGap(id) {
				gaps[s] = e
			}
		})
		var kept []byte
		var at []int
		for i := 0; i < len(text); {
			if e, ok := gaps[i]; ok {
				i = e
				continue
			}
			kept, at = append(kept, text[i]), append(at, i)
			i++
		}
		lit := strings.TrimSpace(old)
		if lit == "" || strings.Count(string(kept), lit) != 1 {
			return "", false
		}
		k, i := strings.Index(string(kept), lit), strings.Index(old, lit)
		return old[:i] + text[at[k]:at[k+len(lit)-1]+1] + old[i+len(lit):], true
	}
	_, end, err := bluedoc.LocateUnique("render", text, old)
	if err != nil {
		return "", false
	}
	after := strings.TrimLeft(text[end:], anchortext.TrailingPunct)
	run := after[:anchor.SkipRun(after, 0)]
	ce := anchortext.ContentEnd(old)
	for _, at := range []int{ce, len(old) - len(strings.TrimLeft(old[ce:], anchortext.TrailingPunct))} {
		if oe := anchor.SkipRun(old, at); anchor.Replace(run, gapless) == old[at:oe] && run != old[at:oe] {
			return old[:at] + run + old[oe:], true
		}
	}
	return "", false
}

// carry puts into new each gap anchor of span it lacks: where its sentence survives, by AutoPlace;
// else after the last content character of new's first prose sentence (the fallback, F-c); where new
// holds no prose, into its marker run (bare). Where new's only prose is a heading, a fence or a table
// row, no sentence carries the anchor: the edit takes it out, and the gap reads gone, as one whose
// quote never placed.
func (p *placer) carry(span, new string) string {
	var missing []string
	for _, id := range anchor.IDs(span) {
		if isGap(id) && !strings.Contains(new, anchor.Token(id)) {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return new
	}
	new = bluedoc.AutoPlace(span, new)
	for _, id := range missing {
		tok := anchor.Token(id)
		at := proseEnd(new)
		switch {
		case strings.Contains(new, tok):
			p.census.Shapes["autoplace"]++
		case at >= 0:
			new = new[:at] + tok + new[at:]
			p.census.Shapes["sentence"]++
			if !slices.Contains(p.census.Fallback, id) {
				p.census.Fallback = append(p.census.Fallback, id)
			}
		case claimcount.HasProse(new):
			p.census.Shapes["unplaced"]++
			p.census.NeverPlaced = append(p.census.NeverPlaced, id)
		default:
			new += tok
			p.census.Shapes["bare"]++
		}
	}
	return new
}

// proseEnd is the offset after the last content character of new's first sentence that says
// something in a paragraph or a list item — never a heading, a fence or a table row — or -1.
func proseEnd(new string) int {
	for _, b := range anchor.Blocks(new) {
		if b.Kind != anchor.Paragraph && b.Kind != anchor.ListItem {
			continue
		}
		for _, sp := range b.Sentences {
			if claimcount.HasProse(new[sp[0]:sp[1]]) {
				return sp[0] + anchortext.ContentEnd(new[sp[0]:sp[1]])
			}
		}
	}
	return -1
}

// quote is a placement's location as replay can locate it: unchanged where it locates as written,
// else the text between its first two double quotes where that does — how a placement written
// before every quote was matched as written named its sentence (`§ Foundations: "…"`).
func (p *placer) quote(loc *string) *string {
	if loc == nil {
		return loc
	}
	l := strings.TrimSpace(*loc)
	i := strings.Index(l, `"`)
	j := strings.Index(l[i+1:], `"`)
	if i < 0 || j < 0 {
		return loc
	}
	q := strings.TrimSpace(l[i+1 : i+1+j])
	text, err := reportproj.RenderFromRecord(p.dst)
	if err != nil || q == "" {
		return loc
	}
	if s, _ := anchortext.LocateSpan(text, l); s >= 0 {
		return loc
	}
	if s, _ := anchortext.LocateSpan(text, q); s < 0 {
		return loc
	}
	p.census.Shapes["quote"]++
	return proto.String(q)
}

func isGap(id string) bool { return anchor.Kind(id) == "gap" }

// gapless is an anchor.Replace that takes out gap anchors and keeps every other.
func gapless(tok, id string) string {
	if isGap(id) {
		return ""
	}
	return tok
}
