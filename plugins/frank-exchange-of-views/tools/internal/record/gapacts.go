package record

import (
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// AN ACT ON A GAP IS A RECORDED EVENT THAT NAMES IT (gblock, 2026-10-25), and A REMAND'S EXCHANGE
// BEGINS ONLY WHEN BLUE RECORDS ONE (ruled by gblock 2026-10-27). A sitting that opens and records
// nothing on the gap is the null turn the exchange fold counts as a turn taken; a remand grants ONE
// more exchange so the direction it states is worked, and a null turn works nothing. The minting
// lens sits before blue in an epoch, so its act on a remanded gap is an audit of an answer that does
// not exist yet: on universe-m35 the voice lens's regrade ("Blue has not answered the remand") was
// counted as the exchange, blue's item went with it, and blue never acted. The lens's act still
// clears the lens's own item (remandItemsOf); it begins nothing.
//
// The acts, by the field each names the gap in:
//
//   - the minting lens: a regrade (one that keeps the grade included), a close, a closing, a finding
//     about the gap, a mint superseding it, a judgement of a computation that answers it;
//   - blue: an edit answering it, a computation answering it, a fate on the inquiry it is anchored on, a
//     closing;
//   - either: a docket or grade motion on it.
//
// Each is validated at its write against the gap the record holds, so a gap named here is one the
// seat named, never one recovered from prose.
type gapAct struct {
	seat string
	at   int64 // the act's events.id
}

// gapActsOf is every act on each gap, in record order. ids are evs's places.
func gapActsOf(evs []*Event, ids []int64) map[string][]gapAct {
	out := map[string][]gapAct{}
	computationGap := map[string]string{} // a computation's anchor → the gap it answers
	anchoredOn := map[string][]string{}   // an inquiry → the gaps minted about it
	for i, e := range evs {
		var on []string
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			if b.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY {
				anchoredOn[b.GetAboutRef()] = append(anchoredOn[b.GetAboutRef()], b.GetGapId())
			}
			on = b.GetSupersedes()
		case *recordpb.Regrade:
			on = []string{b.GetGapId()}
		case *recordpb.Close:
			on = []string{b.GetGapId()}
		case *recordpb.Closing:
			on = []string{b.GetGapId()}
		case *recordpb.Finding:
			if b.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_GAP {
				on = []string{b.GetAboutRef()}
			}
		case *recordpb.BlueEdit:
			on = []string{b.GetAnswers()}
		case *recordpb.Retire:
			on = []string{retireAnswers(e)}
		case *recordpb.Computation:
			computationGap[b.GetComputationId()] = b.GetAnswers()
			on = []string{b.GetAnswers()}
		case *recordpb.Reproduce:
			on = []string{computationGap[b.GetComputationId()]}
		case *recordpb.Inquiry:
			on = anchoredOn[b.GetInquiryId()]
		case *recordpb.Motion:
			on = []string{b.GetDocket().GetGapId(), b.GetGrade().GetGapId()}
		}
		for _, g := range on {
			if g != "" {
				out[g] = append(out[g], gapAct{seat: e.GetSeatId(), at: ids[i]})
			}
		}
	}
	return out
}

// blueActed reports whether blue recorded one of acts in the places (after, upto]: whether the
// round that spans them begins a remand's exchange, and whether the lens's turn has come
// (lensTurnOn).
func blueActed(acts []gapAct, after, upto int64) bool {
	for _, a := range acts {
		if a.at > after && a.at <= upto && seatclass.RoleOfSeat(a.seat) == "blue" {
			return true
		}
	}
	return false
}

// THE LENS'S TURN ON A GAP COMES WITH SOMETHING TO AUDIT (ruled by gblock 2026-10-30: #1603's rule —
// a remand's exchange begins only on blue's act — applied to every exchange). The lens sits before
// blue in an epoch, so a lens readied on a gap nothing has touched since its own last turn there
// audits an answer that does not exist yet: on universe-m36 every fresh gap readied its minting lens
// and blue in the next epoch, and seven lens sittings recorded a keep-the-grade regrade ("Blue has not
// edited the sentence since this gap was minted") before blue sat.
//
// lensTurnOn reports whether the minting lens's turn on a gap has come by upto: blue recorded an
// entry of turns (lensTurnActsOf) after the place the lens last read it (lensReadAt). The plan engages
// the minting lens on the gap only when it holds, and a round whose blue sitting leaves it false ends
// there (exchangeFold), so a gap blue never acts on still reaches impasse.
func lensTurnOn(turns []gapAct, lens string, minted, floor, upto int64) bool {
	return blueActed(turns, lensReadAt(turns, lens, minted, floor, upto), upto)
}

// lensReadAt is the place the minting lens last read a gap by upto: its latest turn there (turns),
// minted, the place of its mint — the mint is not an act on the gap itself — or floor, the gap's latest
// remand ruling (0 for none), whose bench heard every answer before it.
func lensReadAt(turns []gapAct, lens string, minted, floor, upto int64) int64 {
	last := max(minted, floor)
	for _, a := range turns {
		if a.seat == lens && a.at <= upto {
			last = max(last, a.at)
		}
	}
	return last
}

// lensTurnActsOf is what the lens's turn on each gap reads: every act on it (acts), and two more.
//
//   - AN EDIT THAT CHANGED THE GAP'S SENTENCE, whatever gap it answers (ruled by gblock 2026-10-31):
//     the `reopened` that `edited_since` lists. It is the lens's turn and NOT blue's answer. On
//     universe-m37 blue's fix to G-584d152f landed under another gap's name, the lens never read it,
//     and the bench found a gap that was not at impasse.
//   - THE OPENING OF A CLOSED SITTING OF THE LENS ENGAGED ON THE GAP, whether or not it acted: silence
//     is a turn taken, so a null turn does not leave the turn standing for the next epoch. Its opening
//     and not its end, so an act of blue's inside the window still counts as unread. A sitting the
//     record cannot close is no turn taken.
func lensTurnActsOf(evs []*Event, ids []int64, acts map[string][]gapAct, sittings map[string][]partySitting, minted map[string]string) map[string][]gapAct {
	out := make(map[string][]gapAct, len(acts))
	for g, as := range acts {
		out[g] = slices.Clone(as)
	}
	for i, e := range evs {
		if b, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok && b.GetOld() != b.GetNew() {
			for _, g := range b.GetReopened() {
				out[g] = append(out[g], gapAct{seat: e.GetSeatId(), at: ids[i]})
			}
		}
	}
	for g, ss := range sittings {
		for _, s := range ss {
			if s.red && !s.open {
				out[g] = append(out[g], gapAct{seat: minted[g], at: s.start})
			}
		}
	}
	return out
}

// mintPlaces is the place of each gap's mint — its first, as mintedBy reads the minter. ids are
// evs's places.
func mintPlaces(evs []*Event, ids []int64) map[string]int64 {
	out := map[string]int64{}
	for i, e := range evs {
		if m := e.GetMint(); m != nil && m.GetGapId() != "" {
			if _, seen := out[m.GetGapId()]; !seen {
				out[m.GetGapId()] = ids[i]
			}
		}
	}
	return out
}

// seatActedSince reports whether seat recorded one of acts after the place since.
func seatActedSince(acts []gapAct, seat string, since int64) bool {
	for _, a := range acts {
		if a.at > since && a.seat == seat {
			return true
		}
	}
	return false
}
