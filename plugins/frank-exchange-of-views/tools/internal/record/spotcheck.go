package record

import (
	"fmt"
	"sort"
	"strings"
)

// THE W1.8 ARCHIVE SPOT-CHECK FLOOR, COMPUTED FROM THE BOARD.
//
// The duty: re-verify sampled archived closures every round, because a closure index is only as
// good as the last time anyone looked. It was born from a real defect — run 5's round-2 merge
// reported that the spot-check "had degraded to same-seat self-attestation", the seat vouching
// for blocks it was about to write itself.
//
// W1.8 specified the repair as an ENFORCEMENT: key the floor on the archive's state at round
// START, so the obligation exists exactly when there is something to sample. What shipped was an
// envelope self-report — the merge wrote `archive_spot_checks[]` and the script compared it
// against `prevArchiveBlocks`, ANOTHER NUMBER THE MERGE REPORTED. That gate was deleted on
// 2026-07-19, correctly, with its own epitaph: "they compared numbers the merge made up (a haiku
// smoke self-reported archive_blocks:22 in a round where the true archived count was 0). The tool
// board is the count authority; capture audits the truth from disk."
//
// Neither replacement authority was built. The `merge spot-check` verb was created to carry the
// receipt, and for a year the receipt went nowhere: the only code that touched the event was the
// validation switch permitting it to be written. So the fix for a self-attestation defect was, in
// net effect, a better place to write the self-attestation.
//
// This is the floor W1.8 asked for. The archive's size at the start of round R is REPLAYED from
// the board — a number no seat can author — and two things are checked against it: a round that
// owed a sample and recorded none, and a round that claimed an empty archive the board says was
// not empty. The second is the direct heir of the run-5 degeneracy.

// SpotCheck is one round's discharge of the duty, joined to what the board says was actually
// available to sample.
type SpotCheck struct {
	Round  int
	SeatID string
	// Sampled are the archived closures the merge says it re-verified.
	Sampled []string
	Notes   string
	// None is a claim that there was nothing to sample.
	None       bool
	NoneReason string
	// Archived is the board's count of closures available at the START of this round — the
	// authority the claim is checked against.
	Archived int
}

// SpotCheckAudit replays the floor: the discharges recorded, the rounds that owed one and
// recorded nothing, and the discharges whose emptiness claim the board contradicts.
//
// ONE COMPUTATION, TWO READERS. verify enforces it and the report renders it; deriving the floor
// twice would be two definitions free to disagree, which is the shape of defect this whole audit
// keeps finding.
func SpotCheckAudit(b *Board) (checks []SpotCheck, debt []int, falseEmpty []SpotCheck) {
	if b == nil {
		return nil, nil, nil
	}
	// The archive at the START of round R: every gap closed in a round strictly before R.
	// Replayed state, not a reported count.
	// A CLOSURE WITH NO ROUND IS NOT AN EARLY CLOSURE. ClosedRound is derived from the closing
	// seat's ID, and the terminal seats carry no round in their name — `judge-terminal` yields 0.
	// Round 0 is synthesis, when no gap exists to close, so 0 here means UNKNOWN, not FIRST.
	//
	// Reading it as "before everything" put a phantom closure in the archive at the start of
	// round 1 and demanded samples for rounds that could not have taken them — the bench's
	// terminal opinion happens after the last round ends. Measured at 1 seed in 60 by the sweep,
	// which is the only reason it was seen at all: a live run would have failed verify with a
	// message naming rounds whose seats had done nothing wrong.
	//
	// This is the string-derived-fact hazard in miniature (facts-are-fields): the round is
	// recovered from a seat-id by shape, and the miss returns a plausible number rather than an
	// error.
	archivedBefore := func(round int) int {
		n := 0
		for _, g := range b.Gaps {
			if g != nil && g.HasClosed && g.ClosedRound > 0 && g.ClosedRound < round {
				n++
			}
		}
		return n
	}

	// A round is only OWED a sample if the merge actually sat in it. Demanding one from a round
	// the merge never entered would fail a run for a duty nobody was there to discharge — the
	// round-number keying W1.8 replaced, in a new spelling.
	mergeSat := map[int]bool{}
	discharged := map[int]bool{}
	for _, e := range b.Events {
		// REGISTERING IS NOT SITTING. A seat announces itself before it does anything, and a
		// round where the merge registered and then the run ended — a ceiling hit, a PASS, a
		// halt between the two — owed a sample it never had the chance to take. The floor is
		// about work the merge DID, so the announcement does not count as work.
		//
		// Found by the sweep at 1 seed in 60 once an unrelated change shifted the RNG stream:
		// rare, real, and exactly the kind of gate that would have fired on a live run months
		// later with nobody able to say why.
		if PartyOf(e) == "merge" && e.Type != "register" {
			mergeSat[e.Round] = true
		}
		if e.Type != "spot-check" {
			continue
		}
		discharged[e.Round] = true
		sc := SpotCheck{
			Round: e.Round, SeatID: e.SeatID,
			Sampled: e.Payload.StrList("ids"), Notes: e.Payload.Str("notes"),
			Archived: archivedBefore(e.Round),
		}
		if v, ok := e.Payload.Get("none"); ok {
			if bv, isBool := v.(bool); isBool && bv {
				sc.None = true
				sc.NoneReason = e.Payload.Str("reason")
			}
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
		return fmt.Sprintf("r%d (%s): **nothing to sample** — %s", s.Round, s.SeatID, s.NoneReason)
	case len(s.Sampled) == 0:
		return fmt.Sprintf("r%d (%s): recorded, naming no closures", s.Round, s.SeatID)
	default:
		line := fmt.Sprintf("r%d (%s): re-verified %s of %d archived closure(s)",
			s.Round, s.SeatID, strings.Join(s.Sampled, ", "), s.Archived)
		if s.Notes != "" {
			line += " — " + s.Notes
		}
		return line
	}
}
