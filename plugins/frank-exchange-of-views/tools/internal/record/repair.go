package record

import (
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
		return "", feov.Errorf(feov.Validation,
			"record: register refused — %s has no sitting to repair: no register of yours opened one. Register without repairing to open this sitting", seat)
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
	refuse := func(format string, a ...any) error {
		return feov.Errorf(feov.Validation, "record: register refused — "+format, a...)
	}
	named := int64(-1)
	for i, e := range evs {
		if e.GetKey() == key {
			named = int64(i)
		}
	}
	if named < 0 {
		return refuse("%q names no act on the record, so it names no sitting of yours to repair", key)
	}
	if b, ok := recordpb.BodyAs[*recordpb.Register](evs[named]); !ok || evs[named].GetSeatId() != seat || b.RepairsSitting != nil {
		return refuse("%q is not a register that opened a sitting of %s, so it names no sitting of yours to repair", key, seat)
	}
	opens := registers[seat]
	if latest := opens[len(opens)-1]; latest != named {
		return refuse("%q opened an earlier sitting of %s; a repair puts on the record what your LATEST sitting owes, and that sitting was opened by %q", key, seat, evs[latest].GetKey())
	}
	if roleOfSeat(seat) != "blue" {
		return refuse("%s owes no position or revision a repair can file: only a blue sitting owes them", seat)
	}
	for _, d := range ds {
		if d.seat == seat && d.at > named {
			return refuse("%s was dispatched again after the sitting %q opened, so this is a new sitting and not a repair of that one — register without repairing", seat, key)
		}
	}
	if seat == blueRespondSeat {
		for _, s := range BlueSittings(evs, WhileRunning) {
			if s.opened != named {
				continue
			}
			if len(s.Open) == 0 {
				return refuse("the sitting %q opened owes nothing: every gap it was engaged on was closed before it sat, so it owes no position and no revision", key)
			}
			owes := s.Owes()
			if len(owes) == 0 {
				return refuse("the sitting %q opened already carries its position and its revision: there is nothing to repair", key)
			}
			return nil
		}
		return refuse("the sitting %q opened was dispatched onto no gap, so it owes no position and no revision", key)
	}
	closer := sittingCloserOf(evs, seq, registers, WhileRunning)
	spans, end, _ := closer.bounds(seat, named)
	for i := named; i < end; i++ {
		if evs[i].GetSeatId() == seat && evs[i].GetType() == recordpb.EventType_EVENT_TYPE_REVISION && holds(spans, i) {
			return refuse("the sitting %q opened already carries its revision: there is nothing to repair", key)
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
