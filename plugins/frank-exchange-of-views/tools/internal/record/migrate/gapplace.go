package migrate

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// placer runs in Replay after remap.apply and before Append, on bodies already in the destination's
// spelling, adding only what the source lacks: the Anchor of a gap, a citation or a proof where the
// source never anchored it, and gap anchors into an edit's old and new
// where they now stand. With gap anchors removed, every
// rewritten old and new is the archived bytes, so the render's prose is the archived render's. It
// also carries the fold (retirefold.go): the anchors an archived cut left on no words go into its
// removed_anchors, a token a rewrite left past its full stop is re-seated, and a later edit quoting
// either is rewritten to match. It rewrites nothing else: an exact-span edit or a placement that no
// longer places is refused.
//
// It reads the destination through q, the replay's batch (record.Replay.Querier): the batch holds the
// record's one connection, so a render through the run would wait on it forever.
type placer struct {
	q        recordsql.Querier
	anchored map[string]bool // ids the source stream anchors itself, as the migrated record spells them
	placed   bool
	census   migration.GapPlacement

	at        int                   // the source event being replayed, by its place in the stream
	retireAt  map[int]retireAt      // each archived retire, by its place in the stream
	retiresOf map[string][]retireAt // each token's archived retires, in stream order
	removed   map[string]bool       // each token a cut took, out of the report from that cut on
	moved     [][2]string           // each run P12b translation 6 moved, and what it became
}

// newPlacer reads which ids the source stream anchors itself. The bodies step is asked about carry
// migrated ids, so each is keyed on what migrated — the remap's table, filled by now — spells it:
// keyed on the source's spelling, a run that anchored its own acts would gain a second Anchor each.
func newPlacer(q recordsql.Querier, evs []OldEvent, migrated map[string]string) *placer {
	p := &placer{q: q, anchored: map[string]bool{}, census: migration.GapPlacement{Shapes: map[string]int{}}, removed: map[string]bool{}}
	p.retireAt, p.retiresOf = retiresAhead(evs, migrated)
	for _, ev := range evs {
		if id, _ := ev.Fields["id"].(string); ev.Word == "anchor" && id != "" {
			p.anchored[cmp.Or(migrated[id], id)] = true
		}
	}
	return p
}

// step rewrites body for the report as the destination now renders it and returns the Anchor to
// append after it, or nil.
func (p *placer) step(body proto.Message) (proto.Message, error) {
	switch b := body.(type) {
	case *recordpb.Mint:
		if b.GetLocation() == "" || p.anchored[b.GetGapId()] {
			return nil, nil
		}
		text, err := record.RenderAt(p.q)
		if err != nil {
			return nil, nil // no part: the gap reads unrendered
		}
		if _, err := anchortext.Attach(text, b.GetGapId(), b.GetLocation()); err != nil {
			p.census.NeverPlaced = append(p.census.NeverPlaced, b.GetGapId())
			return nil, nil
		}
		p.placed = true
		p.census.Shapes["mint"]++
		return &recordpb.Anchor{Id: proto.String(b.GetGapId()), Location: proto.String(b.GetLocation())}, nil
	case *recordpb.BlueEdit:
		return nil, p.edit(b)
	case *recordpb.Retire:
		return nil, p.retire(b)
	case *recordpb.Cite:
		return p.owed("cite", b.GetLabel(), b.GetLocation())
	case *recordpb.Computation:
		return p.owed("computation", b.GetComputationId(), b.GetLocation())
	case *recordpb.Anchor:
		if b.GetSentenceOf() != "" {
			return nil, p.placesAtSentenceOf(b.GetId(), b.GetSentenceOf())
		}
		return nil, p.places(b.GetId(), b.GetLocation())
	}
	return nil, nil
}

// placesAtSentenceOf refuses an Anchor placed at the end of the sentence anchor `of`'s token sits in
// (a reading of a citation's gap) where the report as it now renders holds no such sentence. Its
// location is empty, so places would pass it unchecked; it is checked by the token's identity, as
// the write places it (anchortext.AttachToSentenceOf).
func (p *placer) placesAtSentenceOf(id, of string) error {
	text, err := record.RenderAt(p.q)
	if err != nil {
		return nil
	}
	if _, err := anchortext.AttachToSentenceOf(text, id, of); err != nil {
		return fmt.Errorf("gap %s's marker ends the sentence citation %s sits in, and the report as the migrated run renders it holds no such sentence: %w", id, of, err)
	}
	return nil
}

// owed is the Anchor an archived citation or computation placed by being recorded: the
// report read its marker off the event itself, so the Anchor appended right after it holds the same
// place in the stream and replay inserts it where it always did. An act that named no id or no
// quote placed nothing. Any other act carrying an id already placed owes a second Anchor, which the
// record refuses.
func (p *placer) owed(shape, id, loc string) (proto.Message, error) {
	if id == "" || loc == "" || p.anchored[id] {
		return nil, nil
	}
	if err := p.places(id, loc); err != nil {
		return nil, err
	}
	p.census.Shapes[shape]++
	return &recordpb.Anchor{Id: proto.String(id), Location: proto.String(loc)}, nil
}

// places refuses a placement replay cannot place in the report as it now renders; a run with no
// render yet has nothing to place against.
func (p *placer) places(id, loc string) error {
	if id == "" || loc == "" {
		return nil
	}
	text, err := record.RenderAt(p.q)
	if err != nil {
		return nil
	}
	if _, err := reportproj.Place(text, loc, id); err != nil {
		return fmt.Errorf("this placement does not place in the report as the migrated run renders it (%v), and migration rewrites no stored location: %q", err, loc)
	}
	return nil
}

// edit stamps an archived edit with the paragraphs it touched, and, once a gap is placed, gives it
// every gap anchor that now stands in or against its span, so it renders and carries them.
//
// A gap anchor standing AGAINST the span is taken into it (the abutting shape). A live edit whose
// quote stops before an anchor run leaves the run where it stands; an archived edit was written
// against a report that held no gap anchor, so its quote says nothing about one, and the translation
// places the gap with the sentence it was minted on — carry's rule — as it does for a gap anchor
// inside the span.
func (p *placer) edit(b *recordpb.BlueEdit) error {
	text, err := record.RenderAt(p.q)
	if err != nil {
		return nil
	}
	// A token a cut took is out of the report from that cut on, and a run P12b translation 6 moved stands
	// where it moved to: a later edit quoting either is rewritten to the report it lands on. Its
	// stored paragraphs keep them, as recorded history.
	b.Old, b.New = proto.String(p.takenOut(b.GetOld())), proto.String(p.takenOut(b.GetNew()))
	b.Old, b.New = proto.String(p.movedOn(b.GetOld())), proto.String(p.movedOn(b.GetNew()))
	old, abuts := b.GetOld(), false
	// Before any gap is placed there is no gap anchor to carry, and the edit is only stamped.
	if p.placed && !b.GetExactSpan() {
		if re, ok := rewriteOld(text, old); ok {
			old, abuts = re, true
		}
	}
	start, end, ok := locate(text, old, b.GetExactSpan())
	// Before any gap is placed an edit that does not locate is the render's to refuse, loudly and
	// naming any refusal before it (replayInOrder): the placer neither stamps nor judges it.
	if !ok && !p.placed {
		return nil
	}
	if !ok && b.GetExactSpan() {
		return fmt.Errorf("this exact-span edit's old text no longer occurs in the report once the gaps' anchors stand in it, and migration rewrites no exact span: %q", old)
	}
	if !ok {
		return fmt.Errorf("this edit no longer locates in the report its gaps' anchors now stand in — a kept husk or a gap anchor inside its span that no rewrite reaches: %q", b.GetOld())
	}
	if abuts {
		b.Old = proto.String(old)
		p.census.Shapes["abutting"]++ // old given the anchor run that now abuts its span
	}
	var unplaced []string
	if p.placed {
		var nw string
		nw, unplaced = p.carry(text[start:end], b.GetNew())
		b.New = proto.String(nw)
	}
	nw, taken := p.bares(text, start, end, b, unplaced)
	b.New = proto.String(nw)
	if len(taken) > 0 {
		b.RemovedAnchors = taken
		for _, id := range taken {
			p.removed[id] = true
		}
	}
	after := reportproj.ApplySplice(text, start, end, b.GetNew())
	// THE EDIT'S TOUCHED PARAGRAPHS, STAMPED ONCE HERE (EventSchema 80): the replay holds the report
	// before the edit and after it, which nothing after migration can rebuild without replaying
	// again, so the archived edit is given what a live edit stamps at its write — and its reopened
	// anchors computed from them as a live edit's are. The paragraphs are the splice's: a token the
	// edit takes stands in them, as the archived edit left it, and the list's removes come after.
	oldParas, newParas := bluedoc.TouchedParagraphs(text, after, start, end)
	b.OldParagraphs, b.NewParagraphs = proto.String(oldParas), proto.String(newParas)
	reopened := bluedoc.ReopenedAnchors(oldParas, newParas)
	for _, id := range reopened {
		if isGap(id) && !slices.Contains(b.Reopened, id) {
			p.census.Shapes["reopened"]++ // a gap whose sentence the archived edit changed
		}
	}
	// A TOKEN THE EDIT TOOK IS NOT REOPENED BY IT. The splice keeps the token, so the re-stamp above
	// reads its proseless sentence as changed; the edit's removed_anchors names it, and a live cut
	// never lists a removed anchor in reopened. Run after the re-stamp, which would list it again.
	b.Reopened = slices.DeleteFunc(reopened, func(id string) bool { return slices.Contains(taken, id) })
	return nil
}

// locate is the span replay takes for old.
func locate(text, old string, exact bool) (int, int, bool) {
	return bluedoc.SpliceSpan(text, old, exact)
}

// rewriteOld is old as replay can locate it now that gap anchors stand in the report: the anchor run
// abutting its span, put in place of old's own run after its last content character.
func rewriteOld(text, old string) (string, bool) {
	_, end, err := bluedoc.LocateUnique(text, old)
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
// holds no prose, into its marker run, which the pure cut then takes (retirefold.go). Where new's
// only prose is a heading, a fence or a table
// row, no sentence carries the anchor: the edit takes it out (it is returned, for the edit's
// removed_anchors), and the gap reads gone, as one whose quote never placed.
func (p *placer) carry(span, new string) (string, []string) {
	var missing []string
	for _, id := range anchor.IDs(span) {
		if isGap(id) && !strings.Contains(new, anchor.Token(id)) {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return new, nil
	}
	var unplaced []string
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
			unplaced = append(unplaced, id) // the splice keeps it; the edit takes it out (removed_anchors)
		default:
			new += tok
			p.census.Shapes["bare"]++
		}
	}
	return new, unplaced
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

func isGap(id string) bool { return anchor.Kind(id) == "gap" }

// gapless is an anchor.Replace that takes out gap anchors and keeps every other.
func gapless(tok, id string) string {
	if isGap(id) {
		return ""
	}
	return tok
}
