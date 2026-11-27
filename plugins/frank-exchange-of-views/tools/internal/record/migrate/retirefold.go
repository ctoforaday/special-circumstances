package migrate

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// ANCHORS FOLLOW THEIR TEXT (EventSchema 82; P12b's translations, #1952, numbered in its own plan and
// distinct from the findings fold's Translations 3-7). A live write leaves no token on no words: a pure cut
// takes the anchors its sentences held and lists them on the edit (BlueEdit.removed_anchors), a
// rewrite carries every anchor, and a retire takes anchors and nothing else. An archive holds two
// shapes the live write no longer makes, and the placer translates each where it replays the edit
// that made it, by what the edit WROTE and the report after it:
//
//   - A pure cut that left a token on no words, with or without a claim retire after it: the cut takes
//     the token (P12b translations 2 and 4). A claim retire whose every token a cut took is folded into the
//     cuts and written as no event; one that carries no anchor is written as no event.
//   - A rewrite that left a token after its sentence's full stop, which no retire takes: the token is
//     re-seated before the full stop of the sentence it trails (P12b translation 6, F20).
//
// Any other token an archived rewrite left bare stays on the claim retire that took it, which becomes
// the anchor-only retire it was in effect — unless it takes an open gap's own marker, which a retire
// may not (F14), and it is written as no event, a lost act. A record whose final report still holds a
// token on no words is refused.

// retireAt is an archived retire as the look-ahead reads it: its place in the source stream, the
// anchors it took as the migrated record spells them, and whether it was a claim retire (the claim
// left the report) rather than an anchor-only one.
type retireAt struct {
	at      int
	anchors []string
	claim   bool
}

// errNoEvent is the placer's answer for a claim retire written as no event because every token it took
// is in the removed_anchors of the cut that bared it: its effect lives in the cut, so it is no lost act.
var errNoEvent = errors.New("written as no event")

// placedLoss is an archived act the placer writes as no event and lists in lost_acts: a claim retire
// that carries no anchor or takes an open gap's own marker (F14), an edit that quoted a token a cut
// took and no longer locates once its quote is rewritten without it. why is the accepted loss.
type placedLoss struct {
	err error
	why string
}

func (l placedLoss) Error() string { return l.err.Error() }

// editLoss is the accepted loss of an archived edit the translations leave nothing to do: a later
// edit that quoted a token a cut took and no longer locates once its quote is rewritten without it.
// One reason per word: the manifest states each word's accepted loss once, and each act's own reason
// is on its lost_acts entry.
const editLoss = "an archived edit the fold leaves nothing to do is written as no event, each listed in lost_acts: a later edit that quoted a token a cut took is rewritten without it, since the token is out of the report from its cut on, and one that no longer locates there is lost"

// retireLoss is the accepted loss of an archived claim retire the record holds no act for: one that
// carried no anchor (a retire takes anchors and nothing else, and a claim leaves with its sentence,
// by a cut), and one that takes an open gap's own marker (F14). One reason per word, as editLoss.
const retireLoss = "an archived claim retire that carried no anchor takes nothing a retire takes, since a claim leaves the report only with its sentence, by a cut; one that takes an open gap's own marker is what a retire may not do (an open gap stands on its marker), every anchor it took staying on its sentence; each is written as no event and listed in lost_acts"

// retiresAhead reads every archived retire in stream order, its anchors in the migrated spelling.
func retiresAhead(evs []OldEvent, migrated map[string]string) (map[int]retireAt, map[string][]retireAt) {
	byAt, byToken := map[int]retireAt{}, map[string][]retireAt{}
	for i, ev := range evs {
		if ev.Word != "retire" {
			continue
		}
		r := retireAt{at: i, claim: !truthy(ev.Fields["claim_kept"])}
		for _, a := range ev.Lists["anchors"] {
			id := cmp.Or(migrated[a], a)
			r.anchors = append(r.anchors, id)
		}
		byAt[i] = r
		for _, id := range r.anchors {
			byToken[id] = append(byToken[id], r)
		}
	}
	return byAt, byToken
}

// retiredAfter is the first archived retire after the event being replayed that takes id.
func (p *placer) retiredAfter(id string) (retireAt, bool) {
	for _, r := range p.retiresOf[id] {
		if r.at > p.at {
			return r, true
		}
	}
	return retireAt{}, false
}

// takenOut is s with every token a cut took earlier taken out: a later edit quoting one is rewritten
// to the report it lands on.
func (p *placer) takenOut(s string) string {
	return anchor.Replace(s, func(tok, id string) string {
		if p.removed[id] {
			return ""
		}
		return tok
	})
}

// movedOn is s with every run P12b translation 6 moved moved the same way: a later edit quoting one is
// rewritten to the report it lands on.
func (p *placer) movedOn(s string) string {
	for _, m := range p.moved {
		s = strings.ReplaceAll(s, m[0], m[1])
	}
	return s
}

// bares is what an archived edit leaves on no words — each token bare after the splice that was not
// bare before it — and what the edit does with each: the pure cut's removed_anchors, or a re-seat.
// It returns the edit's new as replay will splice it, and the tokens it takes, in their list order.
func (p *placer) bares(text string, start, end int, b *recordpb.BlueEdit, unplaced []string) (string, []string) {
	nw := b.GetNew()
	after := reportproj.ApplySplice(text, start, end, nw)
	newly := newlyBare(text, after)
	var taken []string
	if !claimcount.HasProse(nw) {
		taken = newly // a pure cut takes every token it bared
	} else {
		var reseat []string
		for _, id := range newly {
			if _, retired := p.retiredAfter(id); !retired && afterTerminator(nw, id) {
				reseat = append(reseat, id)
			}
		}
		if len(reseat) > 0 {
			nw = p.reseat(nw, reseat)
		}
	}
	for _, id := range unplaced {
		if !slices.Contains(taken, id) {
			taken = append(taken, id)
		}
	}
	p.order(nw, taken)
	return nw, taken
}

// order sorts the tokens a cut takes: those a later retire took in that retire's order (F18, F23),
// the rest in their order in the edit's new (P12b translation 4), then the report's.
func (p *placer) order(nw string, taken []string) {
	key := func(id string) (int, int) {
		if r, ok := p.retiredAfter(id); ok {
			return 0, slices.Index(r.anchors, id)
		}
		if i := strings.Index(nw, anchor.Token(id)); i >= 0 {
			return 1, i
		}
		return 2, 0
	}
	slices.SortStableFunc(taken, func(a, b string) int {
		ra, ia := key(a)
		rb, ib := key(b)
		return cmp.Or(cmp.Compare(ra, rb), cmp.Compare(ia, ib))
	})
}

// newlyBare is every token bare in after that was not bare in before, in after's order.
func newlyBare(before, after string) []string {
	was := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(before) {
		was[id] = true
	}
	var out []string
	for _, id := range claimcount.BareAnchorIDs(after) {
		if !was[id] {
			out = append(out, id)
		}
	}
	return out
}

// terminators end a sentence.
const terminators = ".?!"

// runBefore is where the token run holding id begins in s, walking back over whitespace and other
// anchor tokens, and the terminator just before it (-1 where none is).
func runBefore(s, id string) int {
	i := strings.Index(s, anchor.Token(id))
	if i < 0 {
		return -1
	}
	for i > 0 {
		switch c := s[i-1]; {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i--
		case c == '>':
			j := strings.LastIndex(s[:i], "<!--")
			if j < 0 || len(anchor.IDs(s[j:i])) != 1 || anchor.Token(anchor.IDs(s[j:i])[0]) != s[j:i] {
				return -1
			}
			i = j
		case strings.IndexByte(terminators, c) >= 0:
			return i - 1
		default:
			return -1
		}
	}
	return -1
}

// afterTerminator says whether a sentence terminator precedes id in s with only whitespace and other
// anchor tokens between them.
func afterTerminator(s, id string) bool { return runBefore(s, id) >= 0 }

// reseat moves each selected token to just before the terminator its run follows, the tokens in
// their order, and takes out the whitespace between them (P12b translation 6): "x.<A>\n\n<B>" becomes
// "x<A><B>.". Each run is recorded, so a later edit that quotes it is rewritten the same way.
func (p *placer) reseat(nw string, ids []string) string {
	for len(ids) > 0 {
		term := runBefore(nw, ids[0])
		end := 0
		for _, id := range ids {
			if runBefore(nw, id) == term {
				end = max(end, strings.Index(nw, anchor.Token(id))+len(anchor.Token(id)))
			}
		}
		run := nw[term:end]
		var toks strings.Builder
		var moved []string
		anchor.Each(run, func(s, e int, id string) {
			toks.WriteString(run[s:e])
			if slices.Contains(ids, id) {
				moved = append(moved, id)
			}
		})
		to := toks.String() + run[:1]
		nw = nw[:term] + to + nw[end:]
		p.moved = append(p.moved, [2]string{run, to})
		p.census.Shapes["reseated"] += len(moved)
		ids = slices.DeleteFunc(ids, func(id string) bool { return slices.Contains(moved, id) })
	}
	return nw
}

// retire is an archived retire as the record now holds one. An anchor-only retire is written as it
// stands. A claim retire carrying no anchor, or every token of which a cut took, is written as no
// event; one that takes anchors off a kept sentence is the anchor-only retire it was in effect,
// unless it takes an open gap's own marker (F14).
func (p *placer) retire(b *recordpb.Retire) error {
	r, ok := p.retireAt[p.at]
	if !ok || !r.claim {
		return nil
	}
	var kept, folded []string
	for _, id := range b.GetAnchors() {
		if p.removed[id] {
			folded = append(folded, id)
		} else {
			kept = append(kept, id)
		}
	}
	switch {
	case len(b.GetAnchors()) == 0:
		return placedLoss{errors.New("this claim retire carries no anchor, and a retire takes anchors and nothing else; the claim it named left the report by its cut"), retireLoss}
	case len(kept) == 0:
		return errNoEvent
	case len(folded) > 0:
		return fmt.Errorf("this claim retire takes %s, which a cut took, and %s, which stand on a kept sentence: no archived retire mixes the two, and a retire takes only what the report holds", strings.Join(folded, ", "), strings.Join(kept, ", "))
	}
	if err := record.RequireMarkerOfNoOpenGap(p.q, kept); err != nil {
		return placedLoss{err, retireLoss}
	}
	return nil
}

// bareAtEnd refuses a migrated report that still holds a token on no words: no live writer leaves
// one, and the translations above take or re-seat every one an archive left.
func bareAtEnd(text string) error {
	if bare := claimcount.BareAnchorIDs(text); len(bare) > 0 {
		return fmt.Errorf("the migrated report still holds %s on no words — a token no cut took, no retire took and no re-seat moved; no live write leaves one, so the record is refused", strings.Join(bare, ", "))
	}
	return nil
}
