package migrate

import (
	"slices"
	"strconv"
	"strings"
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
// A dispatch the manifest lists as lost is no source dispatch of this pairing.
// The archive is where the renumbering is real; a seeded run writes no cast first and shifts nothing.
func TestAMigratedDispatchPinsTheEventItsSourcePinned(t *testing.T) {
	checked, renumbered := 0, 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			// A dispatch on a gap a lost act took with it is that same loss, written as no event and
			// listed in the manifest; the rest pair with the migrated dispatches in order.
			res, dir := MigratedArchive(t, name)
			lost := LostIDs(res)
			byID := map[int64]OldEvent{}
			var want []OldEvent      // the event each surviving source dispatch pins, in dispatch order
			var wantDispatch []int64 // each source dispatch's own id, in the same order
			for _, old := range evs {
				byID[old.ID] = old
				if old.Word != "dispatch" || lost[old.ID] {
					continue
				}
				pin, _ := old.Fields["pin"].(int64)
				head, ok := byID[pin]
				if !ok {
					t.Fatalf("source dispatch %d pins %d, which is no earlier event of the source", old.ID, pin)
				}
				want, wantDispatch = append(want, head), append(wantDispatch, old.ID)
			}
			if len(want) == 0 {
				return
			}
			// A SETTLED EDIT IS WRITTEN AS NO EVENT (EventSchema 80), and an edit's key is its seat's
			// count of edits: every later edit of that seat is keyed one lower than its source.
			settledBefore := func(src OldEvent) int {
				n := 0
				for _, id := range res.SettledEdits {
					if s := byID[id]; id < src.ID && s.SeatID == src.SeatID {
						n++
					}
				}
				return n
			}
			m, err := record.MergedEvents(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			ids := m.At.IDs(m.Events)
			at := map[int64]*record.Event{}
			for i, e := range m.Events {
				at[ids[i]] = e
			}
			// AN ENGAGEMENT THAT OWES NONE IS WRITTEN AS NO EVENT (UnowedEngagements), so its place
			// in the source's order is passed over.
			n, dropped := 0, 0
			for _, e := range m.Events {
				d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
				if !ok {
					continue
				}
				if n >= len(want) {
					t.Fatalf("the migrated record holds more dispatches than the source's %d", len(want))
				}
				for n < len(want) && slices.Contains(res.UnowedEngagements, wantDispatch[n]) {
					n++
					dropped++
				}
				if n >= len(want) {
					t.Fatalf("the migrated record holds more dispatches than the source's %d", len(want))
				}
				src := want[n]
				// A pin naming a settled edit names the report head it left standing: the latest
				// source head before it that is not itself settled.
				if slices.Contains(res.SettledEdits, src.ID) {
					src = headBeforeOf(evs, src.ID, res.SettledEdits)
				}
				// AN ARCHIVED REPORT HEAD MAY BE ITS INGESTED TEXT, which the record now holds as the
				// report's one part, under that word in its key.
				word, key := src.Word, src.Key
				if word == "base_ingest" {
					word, key = "report_part", strings.Replace(key, ":base_ingest:", ":report_part:", 1)
				}
				if prefix, n, ok := strings.Cut(key, ":blue_edit:#"); ok && settledBefore(src) > 0 {
					if k, err := strconv.Atoi(n); err == nil {
						key = prefix + ":blue_edit:#" + strconv.Itoa(k-settledBefore(src))
					}
				}
				got := at[d.GetPin()]
				switch {
				case got == nil:
					t.Errorf("dispatch %d pins %d, which is no event of the migrated record", n, d.GetPin())
				case recordpb.Word(got.GetType()) != word || got.GetKey() != key:
					t.Errorf("dispatch %d pins %s %q; its source pinned %s %q (source id %d)",
						n, recordpb.Word(got.GetType()), got.GetKey(), src.Word, src.Key, src.ID)
				case word != "blue_edit" && word != "report_part":
					t.Errorf("dispatch %d pins a %s; a pin names the report head, a blue_edit or a part", n, src.Word)
				}
				if d.GetPin() != src.ID {
					renumbered++
				}
				n++
				checked++
			}
			for n < len(want) && slices.Contains(res.UnowedEngagements, wantDispatch[n]) {
				n++
				dropped++
			}
			if n != len(want) {
				t.Errorf("the migrated record holds %d dispatches, the source %d", n-dropped, len(want))
			}
			if dropped != len(res.UnowedEngagements) {
				t.Errorf("%d source dispatches were passed over as unowed engagements, the manifest states %d", dropped, len(res.UnowedEngagements))
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

// headBeforeOf is the latest source report head before id that is not one of settled.
func headBeforeOf(evs []OldEvent, id int64, settled []int64) OldEvent {
	var head OldEvent
	for _, e := range evs {
		if e.ID >= id {
			break
		}
		switch e.Word {
		case "blue_edit", "base_ingest", "report_part":
			if !slices.Contains(settled, e.ID) {
				head = e
			}
		}
	}
	return head
}
