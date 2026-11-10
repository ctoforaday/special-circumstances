package record

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
)

// THE COMPOSED SECTIONS ARE REPORT TEXT, AND THE VOICE LENS AUDITS THEIR SOURCES (ruled by gblock
// 2026-11-10). Assembly composes Research areas, Future research directions and Alternatives
// considered from the inquiries' fates — each row prints its latest record's reason and method — and
// the risk matrix from the open gaps, each row the first sentence of a gap's problem and of its fix
// (anchor.Concise). None of it is in the report's text, which is all a lens's whole read serves, so
// no lens audited its voice: universe-m40 shipped "A later run should search…" in a deferred
// inquiry's method, "read from cached definitional sources" in a confirmed one's, and "The composed
// row… Re-record the fate…" in a risk-matrix row, after the voice lens had read "the 56-line
// report" sixteen times; the shipped report was 122 lines. m39 shipped the same residue.
//
// So each composed source is the voice lens's to audit, once per cause, as an inquiry's latest fate
// is the logic lens's to check: the record that put the text there readies it (lensFold.readiness)
// until it sits through a sitting opened after that record, and its work list hands it each such
// source in the meantime (composedItemsOf). Narration in an inquiry's reason or method is a
// report_voice gap about the inquiry, and blue answers it by recording the fate again (fateActsOf).
// A gap's problem and fix change only by its minter's successor, so narration there reaches the
// minter as a finding about the gap, which readies it (gapNewsOf) — the voice lens's own gap it
// supersedes itself, and its own mints it does not owe.

// composedSourceAuditor is the lens whose audit of the report covers the composed sections' sources.
const composedSourceAuditor = "red-lens-voice"

// composedSource is one record whose text a composed section prints: an inquiry's latest fate, or an
// open gap's mint. at is the events.id of that record, the cause the auditor owes until it sits
// through a sitting opened after it.
type composedSource struct {
	subject *ItemSubject
	at      int64
	// owed names it among what the auditor owes (lensOwesOf) and in its readiness (lensFold).
	owed string
	// what is its work-list item's sentence (composedItemsOf).
	what string
}

// composedSourcesOf is every composed source on the stream, in record order: each inquiry whose
// latest record gave a reason or a method — a proposal gives neither, and its question and
// hypothesis stand as proposed, so no act could answer a gap about their voice — and each gap of
// open that the auditor did not mint itself.
func composedSourcesOf(evs []*Event, ids []int64, win WindowIndex, open []blockerGap) []composedSource {
	var out []composedSource
	for _, a := range InquiriesOf(evs, win) {
		if a.Reason == "" && a.Method == "" {
			continue
		}
		section := reportheading.InquirySectionOf(a.Status)
		out = append(out, composedSource{
			subject: about(SubjectInquiry, a.ID),
			at:      a.Fate,
			owed:    "the printed fate of inquiry " + a.ID,
			what: fmt.Sprintf("inquiry %s prints under %s with its fate %s, and the reason and the method its record at event %d gave beside it are report text — the inquiries view serves them",
				a.ID, section, a.Status, a.Fate),
		})
	}
	minters := map[string]string{}
	for _, g := range open {
		if g.mintedBy != composedSourceAuditor {
			minters[g.id] = g.mintedBy
		}
	}
	for i, e := range evs {
		m, ok := recordpb.BodyAs[*recordpb.Mint](e)
		if !ok {
			continue
		}
		minter, ok := minters[m.GetGapId()]
		if !ok {
			continue
		}
		out = append(out, composedSource{
			subject: about(SubjectGap, m.GetGapId()),
			at:      ids[i],
			owed:    "the printed problem and fix of gap " + m.GetGapId(),
			what: fmt.Sprintf("gap %s, minted by %s, is open, and the risk matrix prints the first sentence of its problem and of its fix as report text — risk: %q; mitigation: %q",
				m.GetGapId(), minter, anchor.Concise(m.GetProblem()), anchor.Concise(m.GetRequiredFix())),
		})
	}
	slices.SortStableFunc(out, func(a, b composedSource) int { return cmp.Compare(a.at, b.at) })
	return out
}

// composedOwedOf is the composed sources the auditor owes, keyed by its seat: each one whose record
// it has sat through no sitting opened since. Empty where the cast holds no auditor.
func composedOwedOf(evs []*Event, ids []int64, win WindowIndex, open []blockerGap) map[string][]composedSource {
	owed := composedUnhad(evs, ids, win, open, false)
	if len(owed) == 0 {
		return nil
	}
	return map[string][]composedSource{composedSourceAuditor: owed}
}

// composedItemsOf is the work-list items of the composed sources seatID has not had before the
// sitting it is in, each about its inquiry or its gap: none for any seat but the auditor.
//
// THE SITTING IT IS IN IS NOT ONE IT HAD THEM IN. A sitting no agent opened — a register outside any
// subagent — is sat through the moment it opens (satThroughOf), so the debt it ends would empty the
// list of the sitting meant to work it; the list reads the sittings before it, and the debt the
// readiness reads ends with it.
func composedItemsOf(evs []*Event, ids []int64, win WindowIndex, open []blockerGap, seatID string) []Item {
	if seatID != composedSourceAuditor {
		return nil
	}
	var out []Item
	for _, s := range composedUnhad(evs, ids, win, open, true) {
		out = append(out, Item{What: s.what, Kind: ItemComposedSource, Subject: s.subject})
	}
	return out
}

// composedUnhad is the composed sources whose record the auditor has sat through no sitting opened
// since — counting, where before is set, only the sittings that opened before its latest one.
func composedUnhad(evs []*Event, ids []int64, win WindowIndex, open []blockerGap, before bool) []composedSource {
	if !slices.Contains(castOfEvents(evs), composedSourceAuditor) {
		return nil
	}
	_, openings := dispatchLedger(evs, ids, win)
	through := satThroughOf(evs, ids, win, openings).through[composedSourceAuditor]
	if os := openings[composedSourceAuditor]; before && len(os) > 0 {
		through = slices.DeleteFunc(slices.Clone(through), func(o int64) bool { return o >= os[len(os)-1] })
	}
	var out []composedSource
	for _, s := range composedSourcesOf(evs, ids, win, open) {
		if _, sat := firstAfter(through, s.at); !sat {
			out = append(out, s)
		}
	}
	return out
}

// composedNames is the owed names of srcs, in order.
func composedNames(srcs []composedSource) []string {
	out := make([]string, 0, len(srcs))
	for _, s := range srcs {
		out = append(out, s.owed)
	}
	return out
}
