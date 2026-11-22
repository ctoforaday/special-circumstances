package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// TRANSLATIONS 3 AND 4 (EventSchema 80, #1695 §III.B13–B14). The bench convenes on a docketing the
// dispatch writes, and a motion is only a party's proposal, which stores an act. No archived docket
// motion stores one, so none is written as a motion:
//
//   - Translation 3. A docket motion the dispatch filed — the chair's, inside the run of the chair's
//     motion and dispatch rows that a chair plan row ends, the transaction `dispatch next` wrote (O3,
//     structural) — is written as the Docketing the dispatch writes now: its basis the motion's, its id
//     spelled from the source, its cause the live rule's on the record replayed so far. Its ruling is
//     that docketing's (DocketRuling.docketing_id), and names no motion. Every other docket motion — the
//     chair's by hand, a minter's own, blue's — and its ruling are written as no event (O2); the closings
//     argued on them answer no docketing and are refusedArgument's.
//   - Translation 4. The bench writes no word or grade of its own; it sides with a party's stated
//     position. An archived ruling's closing word becomes a for-blue answer with blue stating none —
//     {sided_with: BLUE, disposition: not_a_defect}: the archive holds no red text and no blue word, so
//     blue's text standing is the only stated position. A remand stays a remand. A ruling of a docketing
//     a Close of its gap ended first, and every regrade the bench wrote, are written as no event.
//
// A bench opinion is written as no event: the engine convened the bench there and no party stated a
// position, so its gap reads as the opinion found it. A dispatch's docket left standing on a gap a
// successor superseded is written as no event with its ruling: the live write refuses that supersede.
// A spot-check of a gap whose close one of these drops took away names an open gap, which the write
// refuses, and is written as no event too; so is an inquiry check whose inquiry such a gap alone
// holds wanting, which the write refuses as well (requireInquiryCheck). Each drop is stated in the manifest
// (migration.DocketResidue).

// archivedDockets is the source pre-pass: which archived docket motions the dispatch filed, and the
// docketing each becomes.
type archivedDockets struct {
	docketing map[string]string // archived motion id of a dispatch's docket -> its docketing's id
	dropped   map[string]bool   // archived motion ids of every other docket motion (O2)
	// superseded is the archived motion ids of a dispatch's docket whose gap a successor superseded
	// after it, with no ruling of it and no Close of the gap between: the live write refuses that
	// supersede (requireSupersededOffDocket), so no live record holds the docketing standing on a gap a
	// successor closed. The docketing and its ruling are written as no event.
	superseded map[string]bool
	gapOf      map[string]string // archived motion id -> its archived gap id, for every docket motion
}

// docketsOf reads the source stream once. A docket motion is recognised by its translated body, so
// every era's spelling of one is read the same way.
func docketsOf(evs []OldEvent, reg Registry, rm *remap) (archivedDockets, error) {
	out := archivedDockets{docketing: map[string]string{}, dropped: map[string]bool{}, superseded: map[string]bool{}, gapOf: map[string]string{}}
	for i, old := range evs {
		if old.Word != "motion" {
			continue
		}
		bodies, err := reg.Translate(old)
		if err != nil {
			continue // the replay refuses it at its place, naming why
		}
		for _, b := range bodies {
			m, ok := b.(*recordpb.Motion)
			if !ok || m.GetDocket() == nil {
				continue
			}
			out.gapOf[m.GetMotionId()] = m.GetDocket().GetGapId()
			dispatched, err := filedByTheDispatch(evs, i)
			if err != nil {
				return out, err
			}
			if dispatched {
				out.docketing[m.GetMotionId()] = rm.spelling("docketing", fmt.Sprintf("D-mig-%d", old.ID))
				if supersededStanding(evs[i+1:], reg, m.GetMotionId(), m.GetDocket().GetGapId()) {
					out.superseded[m.GetMotionId()] = true
				}
			} else {
				out.dropped[m.GetMotionId()] = true
			}
		}
	}
	return out, nil
}

// filedByTheDispatch is O3's structural rule for the source event at i: the chair's, and followed by
// nothing but the chair's motion and dispatch rows up to a chair plan row — the one transaction
// `dispatch next` writes. Stamps differ within it by microseconds, so the run, not a stamp, is the key.
func filedByTheDispatch(evs []OldEvent, i int) (bool, error) {
	chair := func(e OldEvent) (bool, error) {
		s, err := roundlessSeat(e.SeatID)
		return s == chairSeat, err
	}
	if c, err := chair(evs[i]); err != nil || !c {
		return false, err
	}
	for _, e := range evs[i+1:] {
		c, err := chair(e)
		if err != nil {
			return false, err
		}
		if !c {
			return false, nil
		}
		switch e.Word {
		case "plan":
			return true, nil
		case "motion", "dispatch":
		default:
			return false, nil
		}
	}
	return false, nil
}

// supersededStanding reports whether, in the source events after a dispatch's docket motion on gap,
// a mint superseding gap comes before any ruling of the motion and any Close of gap.
func supersededStanding(after []OldEvent, reg Registry, motion, gap string) bool {
	for _, e := range after {
		if e.Word != "mint" && e.Word != "close" && e.Word != "motion_rule" {
			continue
		}
		bodies, err := reg.Translate(e)
		if err != nil {
			continue
		}
		for _, b := range bodies {
			switch x := b.(type) {
			case *recordpb.Mint:
				if x.GetSupersedes() == gap {
					return true
				}
			case *recordpb.Close:
				if x.GetGapId() == gap {
					return false
				}
			case *recordpb.MotionRule:
				if x.GetMotionId() == motion {
					return false
				}
			}
		}
	}
	return false
}

// chairSeat is the chair's roundless seat id.
const chairSeat = "red-chair"

// docketDrop is why translate wrote a source event as no event.
type docketDrop int

const (
	keptDocket       docketDrop = iota
	partyDocket                 // a docket no dispatch filed, or its ruling (O2)
	supersededDocket            // a dispatch's docket left standing on a gap a successor superseded, or its ruling
)

// translate rewrites the bodies one source event translates to under Translations 3 and 4, before
// they are respelled: a dispatch's docket motion becomes its docketing (its cause filled at the write,
// docketingCause), its ruling the docketing's, and a dropped docket's motion and ruling nothing.
// drop reports why a body was written as no event.
func (d archivedDockets) translate(bodies []proto.Message) (out []proto.Message, drop docketDrop) {
	for _, b := range bodies {
		switch m := b.(type) {
		case *recordpb.Motion:
			if m.GetDocket() == nil {
				break
			}
			if d.dropped[m.GetMotionId()] {
				drop = partyDocket
				continue
			}
			if d.superseded[m.GetMotionId()] {
				drop = supersededDocket
				continue
			}
			if id, ok := d.docketing[m.GetMotionId()]; ok {
				k := &recordpb.Docketing{DocketingId: proto.String(id), GapId: m.GetDocket().GapId}
				if m.Basis != nil {
					k.Basis = proto.String(m.GetBasis())
				}
				out = append(out, k)
				continue
			}
		case *recordpb.MotionRule:
			if d.dropped[m.GetMotionId()] {
				drop = partyDocket
				continue
			}
			if d.superseded[m.GetMotionId()] {
				drop = supersededDocket
				continue
			}
			id, ok := d.docketing[m.GetMotionId()]
			if !ok || m.GetDocket() == nil {
				break
			}
			r := m.GetDocket()
			r.DocketingId = proto.String(id)
			m.MotionId = nil
			if r.GetDisposition() != recordpb.Disposition_DISPOSITION_REMANDED {
				r.SidedWith = recordpb.SidedWith_SIDED_WITH_BLUE.Enum()
				r.Disposition = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
			}
		}
		out = append(out, b)
	}
	return out, drop
}

// docketingCause fills a translated docketing's cause on the record replayed to its place: the cause
// the dispatch's rule gives a docketing written there (record.DocketingCauseAt).
func docketingCause(q recordsql.Querier, k *recordpb.Docketing) error {
	c, err := record.DocketingCauseAt(q, k.GetGapId())
	if err != nil {
		return fmt.Errorf("reading the cause of docketing %s on %s: %w", k.GetDocketingId(), k.GetGapId(), err)
	}
	k.Cause = c.Enum()
	return nil
}

// endedRuling reports whether body is a ruling of a docketing that does not stand where it is
// written: a Close of its gap ended it first (L42), which the trigger
// motion_rule_docket_answers_a_filing refuses.
func endedRuling(q recordsql.Querier, body proto.Message) (bool, error) {
	r, ok := body.(*recordpb.MotionRule)
	if !ok || r.GetDocket().GetDocketingId() == "" {
		return false, nil
	}
	stands, err := record.DocketingStands(q, r.GetDocket().GetDocketingId())
	return !stands, err
}

// benchRegrade reports whether body is a regrade the bench wrote of its own.
func benchRegrade(body proto.Message, seatID string) bool {
	_, ok := body.(*recordpb.Regrade)
	return ok && seatclass.RoleOfSeat(seatID) == "bench"
}

// reopenedSpotCheck reports whether body is a spot-check naming a gap in reopened — one whose close a
// dropped bench act took away — that is open on the record replayed so far.
func reopenedSpotCheck(q recordsql.Querier, body proto.Message, reopened map[string]bool) (bool, error) {
	s, ok := body.(*recordpb.SpotCheck)
	if !ok {
		return false, nil
	}
	for _, g := range s.GetIds() {
		if !reopened[g] {
			continue
		}
		open, err := record.GapOpen(q, g)
		if err != nil || open {
			return open, err
		}
	}
	return false, nil
}

// reopenedInquiryCheck reports whether body is an inquiry check whose inquiry is wanting on the record
// replayed so far only by gaps in reopened — ones whose close a dropped bench act took away. A check
// any other open material gap holds wanting is not dropped: the write refuses it at its place.
func reopenedInquiryCheck(q recordsql.Querier, body proto.Message, reopened map[string]bool) (bool, error) {
	c, ok := body.(*recordpb.InquiryCheck)
	if !ok {
		return false, nil
	}
	wanting, err := record.OpenMaterialGapsAboutInquiry(q, c.GetInquiryId())
	if err != nil || len(wanting) == 0 {
		return false, err
	}
	for _, g := range wanting {
		if !reopened[g] {
			return false, nil
		}
	}
	return true, nil
}
