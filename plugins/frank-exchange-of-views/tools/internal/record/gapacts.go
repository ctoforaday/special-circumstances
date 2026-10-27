package record

import (
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
//     about the gap, a mint superseding it, a judgement of a proof that answers it;
//   - blue: an edit answering it, a proof answering it, a fate on the inquiry it is anchored on, a
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
	proofGap := map[string]string{}     // a proof's anchor → the gap it answers
	anchoredOn := map[string][]string{} // an inquiry → the gaps minted about it
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
		case *recordpb.Proof:
			proofGap[b.GetProofId()] = b.GetAnswers()
			on = []string{b.GetAnswers()}
		case *recordpb.Reproduce:
			on = []string{proofGap[b.GetProofId()]}
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
// round that spans them begins a remand's exchange.
func blueActed(acts []gapAct, after, upto int64) bool {
	for _, a := range acts {
		if a.at > after && a.at <= upto && seatclass.RoleOfSeat(a.seat) == "blue" {
			return true
		}
	}
	return false
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
