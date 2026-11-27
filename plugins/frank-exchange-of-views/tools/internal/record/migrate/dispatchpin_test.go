package migrate

import (
	"fmt"
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
//
// A HEAD MIGRATE WROTE AS NO EVENT IS PINNED AS THE HEAD BEFORE IT. An edit listed in lost_acts
// changed nothing the migrated report still holds, so the report a pin on it audited is the one the
// last head written before it stands at: the expectation is the latest earlier source head that is
// not lost. A key is the write's, numbered from the seat's own count of its word (record.deriveKey),
// and no field names one: an act that wrote no event leaves each later act of its seat and word one
// lower, and the expectation is the source key less the seat's lost acts of that word before it. A
// SETTLED EDIT is written as no event too (res.SettledEdits) and passes over the same way, so a head
// is gone when it is either. m26 is the one archive where either moves (its source edits 219 and 222,
// each written as no event): 10 pins name 222 and land on source 207's edit, and the 14 pins on its
// later edits 321 and 411 name their keys two lower — the stated residue, the same whether an edit
// is gone as lost or as settled.
func TestAMigratedDispatchPinsTheEventItsSourcePinned(t *testing.T) {
	checked, renumbered := 0, 0
	onLostHead, rekeyed := map[string]int{}, map[string]int{}
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
			gone := func(id int64) bool { return lost[id] || slices.Contains(res.SettledEdits, id) }
			byID := map[int64]OldEvent{}
			var heads []OldEvent           // the source's report heads so far, in stream order
			lostBefore := map[string]int{} // seat and word -> that seat's lost acts of the word so far
			keyOf := map[int64]string{}    // source id -> its key as the migrated record numbers it
			var want []OldEvent            // the event each surviving source dispatch pins, in dispatch order
			var wantKey []string
			var wantDispatch []int64 // each source dispatch's own id, in the same order
			for _, old := range evs {
				byID[old.ID] = old
				keyOf[old.ID] = renumberedKey(old.Key, lostBefore[old.SeatID+"\x00"+old.Word])
				if gone(old.ID) {
					lostBefore[old.SeatID+"\x00"+old.Word]++
				}
				if isHead(old.Word) {
					heads = append(heads, old)
				}
				if old.Word != "dispatch" || lost[old.ID] {
					continue
				}
				pin, _ := old.Fields["pin"].(int64)
				head, ok := byID[pin]
				if !ok {
					t.Fatalf("source dispatch %d pins %d, which is no earlier event of the source", old.ID, pin)
				}
				if gone(head.ID) {
					i := slices.IndexFunc(heads, func(h OldEvent) bool { return h.ID == head.ID })
					for i >= 0 && gone(heads[i].ID) {
						i--
					}
					if i < 0 {
						t.Fatalf("source dispatch %d pins head %d, lost with no head before it", old.ID, pin)
					}
					head = heads[i]
					onLostHead[name]++
				}
				if keyOf[head.ID] != head.Key {
					rekeyed[name]++
				}
				want, wantKey, wantDispatch = append(want, head), append(wantKey, keyOf[head.ID]), append(wantDispatch, old.ID)
			}
			if len(want) == 0 {
				return
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
				// AN ARCHIVED REPORT HEAD MAY BE ITS INGESTED TEXT, which the record now holds as the
				// report's one part, under that word in its key.
				word, key := src.Word, wantKey[n]
				if word == "base_ingest" {
					word, key = "report_part", strings.Replace(key, ":base_ingest:", ":report_part:", 1)
				}
				got := at[d.GetPin()]
				switch {
				case got == nil:
					t.Errorf("dispatch %d pins %d, which is no event of the migrated record", n, d.GetPin())
				case recordpb.Word(got.GetType()) != word || got.GetKey() != key:
					t.Errorf("dispatch %d pins %s %q; its source pinned %s %q, numbered %q here (source id %d)",
						n, recordpb.Word(got.GetType()), got.GetKey(), src.Word, src.Key, key, src.ID)
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
	if want := map[string]int{m26Run: 10}; fmt.Sprint(onLostHead) != fmt.Sprint(want) {
		t.Errorf("pins on a head written as no event, by archive: %v, want %v", onLostHead, want)
	}
	if want := map[string]int{m26Run: 14}; fmt.Sprint(rekeyed) != fmt.Sprint(want) {
		t.Errorf("pins on a head whose key an act written as no event renumbered, by archive: %v, want %v", rekeyed, want)
	}
}

// isHead is whether a source word moves the report head: an edit, or the text a part or an ingest
// lays down.
func isHead(word string) bool {
	return word == "blue_edit" || word == "report_part" || word == "base_ingest"
}

// renumberedKey is key as the migrated record numbers it once lower acts of its seat and word wrote
// no event: a key ending `#N` ends `#N-lower`; a labelled key is the act's own and does not move.
func renumberedKey(key string, lower int) string {
	i := strings.LastIndex(key, "#")
	if lower == 0 || i < 0 {
		return key
	}
	n, err := strconv.Atoi(key[i+1:])
	if err != nil {
		return key
	}
	return key[:i+1] + strconv.Itoa(n-lower)
}
