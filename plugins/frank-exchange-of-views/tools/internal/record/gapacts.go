package record

import (
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// AN ACT ON A GAP IS A RECORDED EVENT THAT NAMES IT (gblock, 2026-10-25), and A REMAND'S EXCHANGE
// BEGINS ONLY WHEN BLUE RECORDS ONE (ruled by gblock 2026-10-27). A sitting that opens and records
// nothing on the gap is the null turn the exchange fold counts as a turn taken; a remand grants ONE
// more exchange so the direction it states is worked, and a null turn works nothing. The assignee
// sits before blue in an epoch, so its act on a remanded gap is an audit of an answer that does
// not exist yet: on universe-m35 the voice lens's regrade ("Blue has not answered the remand") was
// counted as the exchange, blue's item went with it, and blue never acted. The lens's act still
// clears the lens's own item (remandItemsOf); it begins nothing.
//
// THE AUTHOR ANSWERS A GAP ONLY BY AN ACT THAT NAMES IT (plans/findings-fold.md §III.B3; picked
// 2026-11-24, F-FATE "Raiser's turn"). Any other change to its subject is the raiser's turn to
// re-audit (lensTurnActsOf): no answer, no exchange, no stall reset, no remand discharged. An inquiry
// fate, a question part's coverage or split and a retire that takes out the anchor a gap is about
// change that gap's subject; each answers it only where it names it (`answers`).
//
// The acts, by the field each names the gap in:
//
//   - the raiser: a regrade (one that keeps the grade included), a close, a closing, a check (a
//     reading or a judgement) of a citation or computation placed as its answer;
//   - the author: an edit, an evidence anchor (a citation or a computation) placed as its answer, an
//     evidence-anchor retire, an inquiry fate, a question part's coverage or split, and — on a gap
//     about a gap — its minter's regrade of that gap, each naming the gap (`answers`), and a closing;
//   - the lens whose check it is: a check (`verify` or `reproduce`) of an anchor placed as its answer.
//     A reading reaches a gap only through its citation's answer, never through the gap that carries
//     it: a reading that names its gap moves no turn.
//
// A proposal (a docket motion) is no act on its gap: the raiser answers it by accepting it in its
// window, and the dispatch readies it for that (proposalWindowWhy).
//
// Each is validated at its write against the gap the record holds, so a gap named here is one the
// record names, never one recovered from prose.
type gapAct struct {
	seat string
	at   int64 // the act's events.id
}

// gapActsOf is every act on each gap, in record order. ids are evs's places; win is the whole-record
// index the reaching retires are read off.
func gapActsOf(evs []*Event, ids []int64) map[string][]gapAct {
	out := map[string][]gapAct{}
	answering := map[string]string{} // an evidence anchor → the gap it was placed as the answer to
	for i, e := range evs {
		var on []string
		switch b := mustBody(e).(type) {
		case *recordpb.Regrade:
			on = []string{b.GetGapId(), b.GetAnswers()}
		case *recordpb.Close:
			on = []string{b.GetGapId()}
		case *recordpb.Closing:
			on = []string{b.GetGapId()}
		case *recordpb.BlueEdit:
			on = []string{b.GetAnswers()}
		case *recordpb.Retire:
			on = []string{retireAnswers(e)}
		case *recordpb.Computation:
			answering[b.GetComputationId()] = b.GetAnswers()
			on = []string{b.GetAnswers()}
		case *recordpb.Cite:
			answering[b.GetLabel()] = b.GetAnswers()
			on = []string{b.GetAnswers()}
		case *recordpb.Reproduce:
			on = []string{answering[b.GetComputationId()]}
		case *recordpb.Verify:
			on = []string{answering[b.GetAnchor()]}
		case *recordpb.Inquiry:
			on = []string{b.GetAnswers()}
		case *recordpb.QuestionPart:
			on = []string{b.GetAnswers()}
		}
		for k, g := range on {
			if g != "" && !slices.Contains(on[:k], g) {
				out[g] = append(out[g], gapAct{seat: e.GetSeatId(), at: ids[i]})
			}
		}
	}
	return out
}

// A FATE BLUE CHANGED CHANGES WHAT THE GAPS ABOUT IT ARE ABOUT (ruled by gblock 2026-11-05; its
// turn rule F-FATE, 2026-11-24): each gap anchored on the inquiry, and each gap anchored on a section
// the tool composes that prints the inquiry. The tool composes the report's inquiry sections from
// fates, so recording a fate again is the only act that changes them. On universe-m38 the adversary's
// G-ca807516, about the section "Research areas", saw blue record eight fates again with new reasons
// and methods; its lens was never readied, and the bench ruled it with no closings. Such a change is
// each of those gaps' raiser's turn (lensTurnActsOf), and the answer of the one it names (gapActsOf,
// `answers`). A fate changes nothing a gap about other text is about: a reworded `underway` readies
// no unrelated lens.
//
// CHANGED IS AGAINST THE INQUIRY'S PREVIOUS RECORD: its status, its reason or its method differs. The
// write admits a record that repeats all three, and it answers nothing; a proposal has no record
// before it and is no fate.
//
// THE SECTIONS ARE ASSEMBLY'S: the one the previous record printed the inquiry under and the one this
// record prints it under (reportheading.InquirySectionOf, the table the report is laid out from) — a
// fate that moves an inquiry changes both. A gap's section is its mint's reference, compared by the
// key every reader of headings uses (reportheading.HeadingKey).
//
// fateActsOf is the gaps each of blue's changed fates and question parts changes the subject of: the
// gaps minted before it about its inquiry or about a section that prints it, about the part, or
// about the question where a part is split out.
func fateActsOf(evs []*Event) map[*Event][]string {
	out := map[*Event][]string{}
	anchoredOn := map[string][]string{}   // an inquiry → the gaps minted about it
	aboutSection := map[string][]string{} // a section's heading key → the gaps minted about it
	aboutPart := map[string][]string{}    // a question part's id, "" for the question → the gaps minted about it
	prior := map[string]*recordpb.Inquiry{}
	priorCoverage := map[string]*recordpb.QuestionPart{}
	for _, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			switch b.GetAboutKind() {
			case recordpb.AboutKind_ABOUT_KIND_INQUIRY:
				anchoredOn[b.GetAboutRef()] = append(anchoredOn[b.GetAboutRef()], b.GetGapId())
			case recordpb.AboutKind_ABOUT_KIND_SECTION:
				k := reportheading.HeadingKey(b.GetAboutRef())
				aboutSection[k] = append(aboutSection[k], b.GetGapId())
			case recordpb.AboutKind_ABOUT_KIND_QUESTION:
				aboutPart[b.GetAboutRef()] = append(aboutPart[b.GetAboutRef()], b.GetGapId())
			}
		// A QUESTION PART BLUE RECORDED IS ITS ACT ON THE GAPS ABOUT WHAT IT CHANGED, as a fate is
		// (questionpart.go): a part split out acts on the gaps about the question that name no part —
		// a part the record did not hold — and a coverage that differs from the part's previous one,
		// in its value or its reason, on the gaps about that part.
		case *recordpb.QuestionPart:
			if seatclass.RoleOfSeat(e.GetSeatId()) != "blue" {
				continue
			}
			var on []string
			if b.GetSupersedesCoverage() == "" {
				on = slices.Clone(aboutPart[""])
			} else {
				was := priorCoverage[b.GetPartId()]
				priorCoverage[b.GetPartId()] = b
				if was != nil && was.GetCoverage() == b.GetCoverage() && was.GetReason() == b.GetReason() {
					continue
				}
				on = slices.Clone(aboutPart[b.GetPartId()])
			}
			if len(on) > 0 {
				slices.Sort(on)
				out[e] = slices.Compact(on)
			}
		case *recordpb.Inquiry:
			was := prior[b.GetInquiryId()]
			prior[b.GetInquiryId()] = b
			if was == nil || seatclass.RoleOfSeat(e.GetSeatId()) != "blue" ||
				(was.GetStatus() == b.GetStatus() && was.GetReason() == b.GetReason() && was.GetMethod() == b.GetMethod()) {
				continue
			}
			on := slices.Clone(anchoredOn[b.GetInquiryId()])
			for _, st := range []recordpb.InquiryStatus{was.GetStatus(), b.GetStatus()} {
				on = append(on, aboutSection[reportheading.HeadingKey(reportheading.InquirySectionOf(recordpb.Word(st)))]...)
			}
			slices.Sort(on)
			out[e] = slices.Compact(on)
		}
	}
	return out
}

// authorActed reports whether author — the gap's subject's author (authorOf) — recorded one of acts
// in the places (after, upto]: whether the round that spans them begins a remand's exchange, and
// whether the lens's turn has come (lensTurnOn).
func authorActed(acts []gapAct, author string, after, upto int64) bool {
	for _, a := range acts {
		if a.at > after && a.at <= upto && a.seat == author {
			return true
		}
	}
	return false
}

// THE LENS'S TURN ON A GAP COMES WITH SOMETHING TO AUDIT (ruled by gblock 2026-10-30: #1603's rule —
// a remand's exchange begins only on blue's act — applied to every exchange). The lens sits before
// blue in an epoch, so a lens readied on a gap nothing has touched since its own last turn there
// audits an answer that does not exist yet: on universe-m36 every fresh gap readied its assignee
// and blue in the next epoch, and seven lens sittings recorded a keep-the-grade regrade ("Blue has not
// edited the sentence since this gap was minted") before blue sat.
//
// lensTurnOn reports whether the assignee's turn on a gap has come by upto: its subject's author
// recorded an entry of turns (lensTurnActsOf) after the place the lens last read it (lensReadAt). The
// plan engages the assignee on the gap only when it holds, and a round whose author's sitting leaves it
// false ends there (exchangeFold), so a gap its author never acts on still reaches impasse.
func lensTurnOn(turns []gapAct, lens, author string, minted, floor, upto int64) bool {
	return authorActed(turns, author, lensReadAt(turns, lens, minted, floor, upto), upto)
}

// lensReadAt is the place the assignee last read a gap by upto: its latest turn there (turns),
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

// lensTurnActsOf is what the lens's turn on each gap reads: every act on it (acts), and the changes to
// its subject that do not name it — each the raiser's turn and NOT the author's answer (ruled by gblock
// 2026-10-31, kept 2026-11-23 "Lens's turn, no exchange"; F-FATE 2026-11-24).
//
//   - AN EDIT THAT CHANGED THE GAP'S SENTENCE, whatever gap it answers: the `reopened` that
//     `edited_since` lists. On universe-m37 blue's fix to G-584d152f landed under another gap's name,
//     the lens never read it, and the bench found a gap that was not at impasse.
//   - A CHANGED FATE OR QUESTION PART on what the gap is about (fateActsOf), but the gap it names.
//   - A REACHING RETIRE: a retire that took out the anchor the gap is about (WindowIndex.Reaches), but
//     the gap it names.
//   - THE OPENING OF A CLOSED SITTING OF THE LENS ENGAGED ON THE GAP, whether or not it acted: silence
//     is a turn taken, so a null turn does not leave the turn standing for the next epoch. Its opening
//     and not its end, so an act of blue's inside the window still counts as unread. A sitting the
//     record cannot close is no turn taken, and nor is one that did not finish — failed — which
//     gapSittingsOf does not hold.
//
// A change to the subject is entered under the subject's author (authorOf), whoever wrote it: the
// bench's application of a ruling is a change of the subject, and so the raiser's turn on every other
// gap on it, as the author's own change is.
func lensTurnActsOf(evs []*Event, ids []int64, win WindowIndex, acts map[string][]gapAct, sittings map[string][]partySitting, minted map[string]string) map[string][]gapAct {
	out := make(map[string][]gapAct, len(acts))
	for g, as := range acts {
		out[g] = slices.Clone(as)
	}
	author := authorsOn(evs)
	changed := func(g string, at int64) {
		out[g] = append(out[g], gapAct{seat: author(g), at: at})
	}
	fates := fateActsOf(evs)
	for i, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.BlueEdit:
			if b.GetOld() != b.GetNew() {
				for _, g := range b.GetReopened() {
					changed(g, ids[i])
				}
			}
		case *recordpb.Inquiry:
			for _, g := range fates[e] {
				if g != b.GetAnswers() {
					changed(g, ids[i])
				}
			}
		case *recordpb.QuestionPart:
			for _, g := range fates[e] {
				if g != b.GetAnswers() {
					changed(g, ids[i])
				}
			}
		case *recordpb.Retire:
			for _, r := range win.Reaches(e) {
				if r.gap != retireAnswers(e) {
					changed(r.gap, ids[i])
				}
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

// mintPlaces is the place of each gap's mint — its first, as assignee reads the minter. ids are
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
