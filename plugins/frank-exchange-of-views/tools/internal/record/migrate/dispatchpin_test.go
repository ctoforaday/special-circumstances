package migrate

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A MIGRATED DISPATCH PINS THE EVENT ITS SOURCE PINNED (#1257). Every archive renumbers — the cast
// is written first, an anchor lands after the act that placed it, a source event may write nothing
// — so a pin carried as written names whatever the migrated record holds at that number: a log, a
// register, another dispatch. The pin is the report head its party audited, and that head is an
// event of the source, so the migrated pin is the id that event was written as.
//
// Asserted per archive that holds dispatch rows, and in order: the n-th migrated dispatch pins an
// event of the type and key the n-th source dispatch's pin named, and that event is a report head.
// The archive is where the renumbering is real; a seeded run writes no cast first and shifts nothing.
func TestAMigratedDispatchPinsTheEventItsSourcePinned(t *testing.T) {
	checked, renumbered := 0, 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			byID := map[int64]OldEvent{}
			var want []OldEvent // the event each source dispatch pins, in dispatch order
			for _, old := range evs {
				byID[old.ID] = old
				if old.Word != "dispatch" {
					continue
				}
				pin, _ := old.Fields["pin"].(int64)
				head, ok := byID[pin]
				if !ok {
					t.Fatalf("source dispatch %d pins %d, which is no earlier event of the source", old.ID, pin)
				}
				want = append(want, head)
			}
			if len(want) == 0 {
				return
			}
			_, dir := MigratedArchive(t, name)
			m, err := record.MergedEvents(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			ids := m.At.IDs(m.Events)
			at := map[int64]*record.Event{}
			for i, e := range m.Events {
				at[ids[i]] = e
			}
			n := 0
			for _, e := range m.Events {
				d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
				if !ok {
					continue
				}
				if n >= len(want) {
					t.Fatalf("the migrated record holds more dispatches than the source's %d", len(want))
				}
				src := want[n]
				got := at[d.GetPin()]
				switch {
				case got == nil:
					t.Errorf("dispatch %d pins %d, which is no event of the migrated record", n, d.GetPin())
				case recordpb.Word(got.GetType()) != src.Word || got.GetKey() != src.Key:
					t.Errorf("dispatch %d pins %s %q; its source pinned %s %q (source id %d)",
						n, recordpb.Word(got.GetType()), got.GetKey(), src.Word, src.Key, src.ID)
				case src.Word != "blue_edit" && src.Word != "base_ingest":
					t.Errorf("dispatch %d pins a %s; a pin names the report head, a blue_edit or the base ingest", n, src.Word)
				}
				if d.GetPin() != src.ID {
					renumbered++
				}
				n++
				checked++
			}
			if n != len(want) {
				t.Errorf("the migrated record holds %d dispatches, the source %d", n, len(want))
			}
		})
	}
	if checked == 0 {
		t.Fatal("no archived run holds a dispatch, so no pin was checked")
	}
	if renumbered == 0 {
		t.Error("no migrated pin differs from its source's number, so nothing here exercised the translation")
	}
}
