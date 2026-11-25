package migrate

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A DELIVERY OF REPORT TEXT IS ONE ROW PER SEAT AND SITTING, AND ONLY A LENS'S, THE CHAIR'S OR THE
// BENCH'S (EventSchema 80). The write refuses a second report_read in a seat's sitting and one under
// any other seat, so an archived sitting holding several serves keeps its latest and drops the rest,
// and a serve under any other seat is dropped; the manifest counts each in events_in and none in
// events_out. An archived report_read was a serve of the whole report through stdout, and reads as a
// completed delivery at its head: archived pins stay where they were, and an archived later serve
// counts as the full re-read. Residue: archived rows record serves through stdout, not deliveries
// into the seat's context.
//
// A sitting here is the stretch of a seat's events from one opening to its next — a register under
// the seat, or a sitting_open naming it — read off the source in order, as the write stamps it.
func readsDropped(evs []OldEvent, seat func(string) (string, error)) map[int64]bool {
	drop := map[int64]bool{}
	last := map[string]int64{} // seat -> its latest report_read since its latest opening
	for _, ev := range evs {
		switch ev.Word {
		case "register":
			if s, err := seat(ev.SeatID); err == nil {
				delete(last, s)
			}
		case "sitting_open":
			if id, _ := ev.Fields["seat_id"].(string); id != "" {
				if s, err := seat(id); err == nil {
					delete(last, s)
				}
			}
		case "report_read":
			s, err := seat(ev.SeatID)
			if err != nil {
				continue // the seat is refused where the event is replayed
			}
			switch seatclass.RoleOfSeat(s) {
			case "lens", "chair", "bench":
			default:
				drop[ev.ID] = true
				continue
			}
			if prev, ok := last[s]; ok {
				drop[prev] = true
			}
			last[s] = ev.ID
		}
	}
	return drop
}
