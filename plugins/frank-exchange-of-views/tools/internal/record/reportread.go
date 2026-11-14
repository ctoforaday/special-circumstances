package record

// A SEAT'S PIN IS THE HEAD OF ITS LATEST COMPLETE DELIVERY in a sitting it sat through
// (report_delivery). The delivery writer appends one report_read when it hands a delivery's last item
// into the seat's context, its head the report head the delivery was cut at: the tool's record of
// what reached the seat, never the seat's claim to have read it. A lens never delivered the report at
// the head is ready at it until it is, or retires; a retired lens behind the head is a stale area the
// chair bears changes on.

// ServedPinNote is what moves a lens's pin, as its work list's audit item states it.
const ServedPinNote = "Your pin moves when the report text owed to you this sitting has all been delivered on your work list, in a sitting you sit through; a read through `show report` records nothing"

// served is one completed delivery to a seat (report_delivery): the sitting it was in, its head, and
// whether that sitting was sat through.
type served struct {
	opening, head int64
	sat           bool
}

// servedOf is each seat's deliveries, in record order, read off the report_delivery view. A delivery
// in no sitting is in none.
func servedOf(win WindowIndex) map[string][]served {
	win.mustViews()
	out := map[string][]served{}
	for _, r := range win.deliveries {
		if r.SittingID == 0 {
			continue
		}
		out[r.SeatID] = append(out[r.SeatID], served{opening: r.SittingID, head: r.Head, sat: r.SatThrough})
	}
	return out
}

// pinOf is the lens's pin: the head of its latest delivery in a sitting it sat through whose opening
// precedes before, 0 where none does — an id comparison over report_delivery's rows.
func pinOf(reads []served, before int64) int64 {
	var pin int64
	for _, r := range reads {
		if r.opening < before && r.sat && r.head > pin {
			pin = r.head
		}
	}
	return pin
}
