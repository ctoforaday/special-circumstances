package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A SITTING-RECORD REPAIR IS PART OF THE SITTING IT REPAIRS, AND THE REGISTER SAYS WHICH.
//
// The engine re-prompts a blue seat whose sitting did not put its position or revision on the
// record. The re-prompted agent registers, and a register opens a sitting, so the repair's acts
// belonged to no sitting: under the shipped Workflow engine the first agent's stop closes the
// sitting before the re-prompt runs, and capture's record-parity audit could fail the sitting whose
// repair filed exactly what it owed. Joining the two by reading the chair's dispatch rows would make
// a read depend on another seat's write.
//
// So the repair is a FIELD on the register (repairs_sitting), written at the write and refused
// there. The seat asserts it is repairing — the engine's re-prompt tells it to — and the write names
// the sitting, so the seat cannot name a wrong one; checkRepair refuses a claim the record does not
// bear out, so no seat can claim a repair it is not doing.
//
// THE SEAT ASSERTS AND THE WRITER CHECKS; THE WRITER DOES NOT DETECT. A register that finds its
// seat's latest sitting owing, with no dispatch since, looks the same whether it is a re-prompt or
// a crashed sitting dispatched again, so a writer that inferred the repair would record one nobody
// claimed.

// A REFUSED REPAIR HAS TWO MEANINGS AND THE REFUSAL SAYS WHICH (#1026).
//
// The engine's re-prompt branches on what it is told: either there is nothing for a repair to file
// — and the seat then opens a sitting of its own and reports what the record supports — or the claim
// is one the record does not bear out, which is a failure to report. It keyed that branch on one
// refusal's prose, and four refusals that meant the first said something else; the worst left a
// re-prompted seat filing NOTHING when it had no sitting to repair, so the record it was sent to
// complete stayed incomplete.
//
// So the outcome is stated at every refusal site rather than recovered from the sentence. The class
// is a compile-time argument, so a refusal cannot be added without choosing a branch, and the
// "nothing to file" class ends with RepairNothingToFile — one sentence this package writes, which
// the re-prompt names verbatim (TestTheRePromptNamesEveryRepairRefusalsBranch).

// repairOutcome is which branch a refused repair puts the seat on. The set is CLOSED: the re-prompt
// states both, and a third would be a branch no prompt holds.
type repairOutcome int

const (
	// nothingToFile: the sitting the seat would repair owes a repair nothing, so the honest next
	// act is a register that opens a sitting of the seat's own.
	nothingToFile repairOutcome = iota
	// claimUnfounded: the register names a sitting the record does not bear out as this seat's
	// latest. Reachable only where the key comes from somewhere other than repairTarget, which
	// reads it off the record — a forged register, or another write path.
	claimUnfounded
)

// RepairNothingToFile is the sentence every nothingToFile refusal ends with. It is the anchor the
// engine's re-prompt keys on: a tool-placed sentence in a refusal a seat reads, rather than each
// refusal's own wording, which is what the prompt was matching and mostly missing.
const RepairNothingToFile = "There is nothing here for a repair to file: register without repairing, and report what the record supports."

// refuseRepair is the one place a repair refusal is written, so the class and the sentence cannot
// drift apart.
func refuseRepair(out repairOutcome, format string, a ...any) error {
	msg := "record: register refused — " + fmt.Sprintf(format, a...)
	if out == nothingToFile {
		msg += " " + RepairNothingToFile
	}
	return feov.Errorf(feov.Validation, "%s", msg)
}

// repairTarget is the sitting a register by seat repairs — the key of the register that opened
// the seat's latest sitting — once checkRepair admits it.
func repairTarget(evs []*Event, seat string) (string, error) {
	evs = Live(evs)
	var latest *Event
	for _, e := range evs {
		if b, ok := recordpb.BodyAs[*recordpb.Register](e); ok && e.GetSeatId() == seat && b.RepairsSitting == nil {
			latest = e
		}
	}
	if latest == nil {
		return "", refuseRepair(nothingToFile,
			"%s has no sitting to repair: no register of yours opened one.", seat)
	}
	key := latest.GetKey()
	return key, checkRepair(evs, seat, key)
}

// checkRepair refuses a register by seat that names key as the sitting it repairs, unless key is
// the register that opened seat's LATEST sitting, no dispatch has named the seat since, and that
// sitting OWES what a repair files: a blue-respond sitting that found a gap open and lacks its
// position or revision (BlueSitting.Owes), or any other blue seat's sitting that lacks its revision.
func checkRepair(evs []*Event, seat, key string) error {
	evs = Live(evs)
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	ds, registers := dispatchLedger(evs, seq)
	named := int64(-1)
	for i, e := range evs {
		if e.GetKey() == key {
			named = int64(i)
		}
	}
	if named < 0 {
		return refuseRepair(claimUnfounded, "%q names no act on the record, so it names no sitting of yours to repair", key)
	}
	if b, ok := recordpb.BodyAs[*recordpb.Register](evs[named]); !ok || evs[named].GetSeatId() != seat || b.RepairsSitting != nil {
		return refuseRepair(claimUnfounded, "%q is not a register that opened a sitting of %s, so it names no sitting of yours to repair", key, seat)
	}
	opens := registers[seat]
	if latest := opens[len(opens)-1]; latest != named {
		return refuseRepair(claimUnfounded, "%q opened an earlier sitting of %s; a repair puts on the record what your LATEST sitting owes, and that sitting was opened by %q", key, seat, evs[latest].GetKey())
	}
	if roleOfSeat(seat) != "blue" {
		return refuseRepair(nothingToFile, "%s owes no position or revision a repair can file: only a blue sitting owes them.", seat)
	}
	for _, d := range ds {
		if d.seat == seat && d.at > named {
			return refuseRepair(nothingToFile, "%s was dispatched again after the sitting %q opened, so this is a new sitting and not a repair of that one.", seat, key)
		}
	}
	if seat == blueRespondSeat {
		for _, s := range BlueSittings(evs, WhileRunning) {
			if s.opened != named {
				continue
			}
			if len(s.Open) == 0 {
				return refuseRepair(nothingToFile, "the sitting %q opened owes nothing: every gap it was engaged on was closed before it sat, so it owes no position and no revision.", key)
			}
			owes := s.Owes()
			if len(owes) == 0 {
				return refuseRepair(nothingToFile, "the sitting %q opened already carries its position and its revision: there is nothing to repair.", key)
			}
			return nil
		}
		return refuseRepair(nothingToFile, "the sitting %q opened was dispatched onto no gap, so it owes no position and no revision.", key)
	}
	closer := sittingCloserOf(evs, seq, registers, WhileRunning)
	spans, end, _ := closer.bounds(seat, named)
	for i := named; i < end; i++ {
		if evs[i].GetSeatId() == seat && evs[i].GetType() == recordpb.EventType_EVENT_TYPE_REVISION && holds(spans, i) {
			return refuseRepair(nothingToFile, "the sitting %q opened already carries its revision: there is nothing to repair.", key)
		}
	}
	return nil
}

// requireRepairable is checkRepair on the run's record, for a register written with repairs_sitting.
func requireRepairable(run Run, seat, key string) error {
	m, err := MergedEvents(run)
	if err != nil {
		return err
	}
	return checkRepair(m.Events, seat, strings.TrimSpace(key))
}
