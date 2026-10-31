package record

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// RecordRead is what the record says at a seat's close, written onto the seat's envelope by the
// PreToolUse hook and never by the seat (gblock, 2026-10-31). The engine has no other way to read
// the record: a judge's envelope saying "Not halted…" once read as HALTED, because the engine learned
// of a halt from what a seat composed.
//
// THE HOOK OVERWRITES ANYTHING THE SEAT TYPED, so a seat cannot supply or change this. It is not a
// second account of the record in the sense LensEnvelope's comment refuses: the tool wrote it, off
// the record, and capture's envelope-read audit recomputes it at Event.
//
// ABSENT IS NOT EMPTY. An envelope with no field is a hook that did not run, which stops the run; an
// empty BenchOwed is a read that found nothing owed. The field is declared but not required, because
// a required field would have a seat type it when the hook is dead.
type RecordRead struct {
	Event     int64           `json:"event" min:"0" desc:"the events.id the record was read at"`
	BenchOwed []BenchHearing  `json:"bench_owed" desc:"each sitting of the bench that a filing on the record is owed and has not had"`
	Halted    bool            `json:"halted" desc:"the bench has halted the run"`
	Outcome   RecordedVerdict `json:"outcome,omitempty" desc:"the outcome stamped on the record; absent while none is"`
}

// BenchHearing is a bench occasion a filing convenes: petition or position.
type BenchHearing string

// RecordedVerdict is the run's outcome as the record states it, in capitals.
type RecordedVerdict string

// BenchHearings is the occasions a read can report owed, in the order the engine hears them.
func BenchHearings() []string {
	return []string{occasionPetition, occasionPosition}
}

// RecordReadFields is RecordRead's fields, as the walk the plan's schema is generated from reads them.
func RecordReadFields() []PlanField { return fieldsOf(reflect.TypeFor[RecordRead]()) }

// ReadRecord is the record's read as it stands now: one snapshot, so the event it names is the last
// the read saw. A run with no record is an error, as an envelope's check is: the hook refuses on it.
func ReadRecord(run Run) (RecordRead, error) {
	if err := recordPresent(run); err != nil {
		return RecordRead{}, err
	}
	var read RecordRead
	err := readSnapshot(run, func(q recordsql.Querier) error {
		evs, win, err := eventsAt(q)
		if err != nil {
			return err
		}
		read = readOf(evs, win)
		return nil
	})
	return read, err
}

// RecordReadAt recomputes the read a close's envelope carries, from the events up to and including
// event: the same fold, on the prefix the hook read. Capture's envelope-read audit asks it. An event
// the stream does not hold is an error: a read naming one was not read off this record.
func RecordReadAt(evs []*Event, win WindowIndex, event int64) (RecordRead, error) {
	ids := win.IDs(evs)
	n, found := slices.BinarySearch(ids, event)
	if !found && event != 0 {
		return RecordRead{}, fmt.Errorf("the stream holds no event %d", event)
	}
	if found {
		n++
	}
	return readOf(evs[:n], win), nil
}

// BenchOwedBefore is the hearings owed on the record before the event at position pos of evs: what
// capture holds a hearing's opening to. The engine convenes one only where a read reports it.
func BenchOwedBefore(evs []*Event, win WindowIndex, pos int) []string {
	prefix := evs[:pos]
	owed := benchOwedOf(prefix, win.IDs(prefix), win)
	out := make([]string, len(owed))
	for i, o := range owed {
		out[i] = string(o)
	}
	return out
}

// readOf is the read over evs, whose last event is the one it names.
func readOf(evs []*Event, win WindowIndex) RecordRead {
	ids := win.IDs(evs)
	read := RecordRead{BenchOwed: benchOwedOf(evs, ids, win)}
	if len(ids) > 0 {
		read.Event = ids[len(ids)-1]
	}
	for _, e := range evs {
		switch b := mustBodyOf(e).(type) {
		case *recordpb.Halt:
			read.Halted = true
		case *recordpb.Outcome:
			read.Outcome = RecordedVerdict(strings.ToUpper(recordpb.Word(b.GetVerdict())))
		}
	}
	return read
}

// benchOwedOf is each bench hearing a filing stands owed: an unruled motion of that subject that no
// sitting of that occasion has heard since it was filed. One sitting hears a filing, so a filing the
// bench sat on and left unruled is not owed again: nothing re-enters a hearing, and the filing holds
// the PASS until the exit. Never nil: an empty list is a read that found none.
func benchOwedOf(evs []*Event, ids []int64, win WindowIndex) []BenchHearing {
	on := benchOccasionSittings(evs, ids, win)
	owed := map[recordpb.Occasion]bool{}
	for _, b := range unruledMotionBlockers(evs, ids, win) {
		subj, _ := MotionSubjectEnum(b.motionSubject)
		occ := benchSittingFor[subj]
		if b.Owner != benchSeat || occ == recordpb.Occasion_OCCASION_DOCKET || occ == recordpb.Occasion_OCCASION_UNSPECIFIED {
			continue
		}
		if !benchSatOnOccasionSince(on, occ, b.since) {
			owed[occ] = true
		}
	}
	out := []BenchHearing{}
	for _, o := range []recordpb.Occasion{recordpb.Occasion_OCCASION_PETITION, recordpb.Occasion_OCCASION_POSITION} {
		if owed[o] {
			out = append(out, BenchHearing(recordpb.Word(o)))
		}
	}
	return out
}
