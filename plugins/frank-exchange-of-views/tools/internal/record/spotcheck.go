package record

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE W1.8 ARCHIVE SPOT-CHECK FLOOR, COMPUTED FROM THE BOARD.
//
// The duty: re-verify sampled archived closures every epoch, because a closure index is only as
// good as the last time anyone looked. It was born from a real defect — run 5's round-2 merge
// reported that the spot-check "had degraded to same-seat self-attestation", the seat vouching
// for blocks it was about to write itself.
//
// W1.8 specified the repair as an ENFORCEMENT: key the floor on the archive's state at epoch
// START, so the obligation exists exactly when there is something to sample. What shipped was an
// envelope self-report — the merge wrote `archive_spot_checks[]` and the script compared it
// against `prevArchiveBlocks`, ANOTHER NUMBER THE MERGE REPORTED. That gate was deleted on
// 2026-07-19, correctly, with its own epitaph: "they compared numbers the merge made up (a haiku
// smoke self-reported archive_blocks:22 in an epoch where the true archived count was 0). The tool
// board is the count authority; capture audits the truth from disk."
//
// Neither replacement authority was built. The `merge spot-check` verb was created to carry the
// receipt, and for a year the receipt went nowhere: the only code that touched the event was the
// validation switch permitting it to be written. So the fix for a self-attestation defect was, in
// net effect, a better place to write the self-attestation.
//
// This is the floor W1.8 asked for. The archive's size at the start of epoch R is REPLAYED from
// the board — a number no seat can author — and two things are checked against it: an epoch that
// owed a sample and recorded none, and an epoch that claimed an empty archive the board says was
// not empty. The second is the direct heir of the run-5 degeneracy.

// SpotCheck is one epoch's discharge of the duty, joined to what the board says was actually
// available to sample.
type SpotCheck struct {
	Epoch   int
	Sitting int
	SeatID  string
	// Sampled are the archived closures the merge says it re-verified.
	Sampled []string
	// Prose is what the sample found, or why there was nothing to sample. ONE channel: it was
	// `notes` and `reason`, two payload keys filled by different branches of one verb, so the
	// report rendered whichever the seat happened to reach for and dropped the other.
	Prose string
	// None is a claim that there was nothing to sample.
	None       bool
	NoneReason string
	// Archived is the board's count of closures available at the START of this epoch — the
	// authority the claim is checked against.
	Archived int
}

// SpotCheckAudit replays the floor: the discharges recorded, the epochs that owed one and
// recorded nothing, and the discharges whose emptiness claim the board contradicts.
//
// ONE COMPUTATION, TWO READERS. verify enforces it and the report renders it; deriving the floor
// twice would be two definitions free to disagree, which is the shape of defect this whole audit
// keeps finding.
func SpotCheckAudit(f Family) (checks []SpotCheck, debt []int, falseEmpty []SpotCheck) {
	if len(f.Gaps) == 0 && len(f.Events) == 0 {
		return nil, nil, nil
	}
	// The archive at the START of epoch R: every gap closed in an epoch strictly before R.
	// Replayed state, not a reported count.
	// A CLOSURE WITH NO EPOCH IS NOT AN EARLY CLOSURE. ClosedEpoch is derived from the closing
	// seat's ID, and the terminal seats carry no epoch in their name — `judge-terminal` yields 0.
	// Synthesis is synthesis, when no gap exists to close, so 0 here means UNKNOWN, not FIRST.
	//
	// Reading it as "before everything" put a phantom closure in the archive at the start of
	// round 1 and demanded samples for rounds that could not have taken them — the bench's
	// terminal opinion happens after the last epoch ends. Measured at 1 seed in 60 by the sweep,
	// which is the only reason it was seen at all: a live run would have failed verify with a
	// message naming epochs whose seats had done nothing wrong.
	//
	// This is the string-derived-fact hazard in miniature (facts-are-fields): the epoch is
	// recovered from a seat-id by shape, and the miss returns a plausible number rather than an
	// error.
	archivedBefore := func(epoch int) int {
		n := 0
		for _, g := range f.Gaps {
			if g != nil && g.HasClosed && g.ClosedEpoch > 0 && g.ClosedEpoch < epoch {
				n++
			}
		}
		return n
	}

	// An epoch is only OWED a sample if the merge actually sat in it. Demanding one from an epoch
	// the merge never entered would fail a run for a duty nobody was there to discharge — the
	// epoch-number keying W1.8 replaced, in a new spelling.
	mergeSat := map[int]bool{}
	discharged := map[int]bool{}
	var clk Clock
	for _, e := range f.Live() {
		w := clk.Advance(e)
		// REGISTERING IS NOT SITTING. A seat announces itself before it does anything, and a
		// epoch where the merge registered and then the run ended — a ceiling hit, a PASS, a
		// halt between the two — owed a sample it never had the chance to take. The floor is
		// about work the merge DID, so the announcement does not count as work.
		//
		// Found by the sweep at 1 seed in 60 once an unrelated change shifted the RNG stream:
		// rare, real, and exactly the kind of gate that would have fired on a live run months
		// later with nobody able to say why.
		if PartyOf(e) == "merge" && e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
			mergeSat[w.Epoch] = true
		}
		// The BODY is the type test. Named `body` because the projection struct this loop fills
		// is ALSO called SpotCheck — record.SpotCheck is the audit row, recordpb.SpotCheck the
		// event body.
		body, ok := recordpb.BodyAs[*recordpb.SpotCheck](e)
		if !ok {
			continue
		}
		round := w.Epoch
		discharged[round] = true
		sc := SpotCheck{
			Epoch: round, Sitting: w.Sitting, SeatID: e.GetSeatId(),
			// ONE PROSE CHANNEL. `notes` and `reason` were two payload keys filled by different
			// branches of one verb; SpotCheck.reason is the field that replaces BOTH, so a
			// record written under `notes` maps here and not to nothing.
			Sampled: body.GetIds(), Prose: body.GetReason(),
			Archived: archivedBefore(round),
		}
		// PRESENT AND TRUE, exactly as before: the old read asked Get("none") for presence and
		// then required the value to be a true bool. An absent `none` is false here, and a
		// recorded `none: false` is a seat NOT claiming an empty archive — the same two answers.
		if body.GetNone() {
			sc.None = true
			sc.NoneReason = body.GetReason()
		}
		checks = append(checks, sc)
		// THE CLAIM THE BOARD CAN REFUSE. "There was nothing to sample" is checkable, and it is
		// the exact shape run 5's degeneracy took.
		if sc.None && sc.Archived > 0 {
			falseEmpty = append(falseEmpty, sc)
		}
	}

	for round := range mergeSat {
		if archivedBefore(round) > 0 && !discharged[round] {
			debt = append(debt, round)
		}
	}
	sort.Ints(debt)
	return checks, debt, falseEmpty
}

// Describe renders one spot-check as a reader-facing line.
func (s SpotCheck) Describe() string {
	switch {
	case s.None:
		return fmt.Sprintf("#%d (%s): **nothing to sample** — %s", s.Sitting, s.SeatID, s.NoneReason)
	case len(s.Sampled) == 0:
		return fmt.Sprintf("#%d (%s): recorded, naming no closures", s.Sitting, s.SeatID)
	default:
		line := fmt.Sprintf("#%d (%s): re-verified %s of %d archived closure(s)",
			s.Sitting, s.SeatID, strings.Join(s.Sampled, ", "), s.Archived)
		if s.Prose != "" {
			line += " — " + s.Prose
		}
		return line
	}
}
