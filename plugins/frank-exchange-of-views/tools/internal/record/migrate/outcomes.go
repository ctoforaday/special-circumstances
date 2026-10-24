package migrate

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RECORD HOLDS ONE OUTCOME, THE LATEST ITS SOURCE STAMPED (gblock, 2026-10-23: "Drop the earlier
// outcome"). The write refuses every seat act once the record holds an outcome
// (record.requireOpenToSeats), a second outcome among them. A source written before that refusal
// may hold a second assembly sitting after the first outcome — its register, then its own outcome
// (three archived runs: research-loop-counterparts, sleeper-service-plan, quadratic-formula) — and
// every reader of the source took the later one (record.RecordedOutcome reads the newest). So an
// outcome the source follows with another is not written: the later sitting's register and outcome
// then land, and the migrated record's one outcome is the one the source's readers reported.
//
// What the dropped outcome's place held stays as written: the bench's declare and certify after it
// land under the live rule either way, and a log is no seat act. An act that names the dropped
// outcome by events.id — a pin, a stale read point, a stored plan, a contradiction — finds no
// event written from it and is refused by name, as every such reference to an unwritten act is.

// supersededOutcomes is every source event that translates to an outcome and is followed in the
// stream by another, keyed by its source id, never nil. A correction chain is written as its head
// at its root's place, so the chain is one outcome here, at the root's place, and what it says is
// its head's — the act the seat meant, and the one the source's readers took. headOf is the act a
// place writes: the head of the chain it begins, or the act itself.
func supersededOutcomes(evs []OldEvent, reg Registry, replaces map[string]bool, headOf func(OldEvent) OldEvent) map[int64]migration.SupersededOutcome {
	type held struct {
		old     OldEvent // the place: the act, or the root of its chain
		head    OldEvent // what the place writes
		verdict string
	}
	var outcomes []held
	for _, old := range evs {
		if replaces[old.Key] {
			continue
		}
		head := headOf(old)
		bodies, _ := reg.Translate(head)
		for _, b := range bodies {
			if o, ok := b.(*recordpb.Outcome); ok {
				outcomes = append(outcomes, held{old, head, recordpb.Word(o.GetVerdict())})
				break
			}
		}
	}
	out := map[int64]migration.SupersededOutcome{}
	if len(outcomes) < 2 {
		return out
	}
	last := outcomes[len(outcomes)-1].old.ID
	for _, h := range outcomes[:len(outcomes)-1] {
		out[h.old.ID] = migration.SupersededOutcome{OldID: h.old.ID, Seat: h.head.SeatID, Verdict: h.verdict, SupersededBy: last}
	}
	return out
}
