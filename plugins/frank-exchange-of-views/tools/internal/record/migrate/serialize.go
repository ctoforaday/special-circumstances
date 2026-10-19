package migrate

import (
	"maps"
	"slices"
	"strconv"
)

// serializeInstances puts an archived run's concurrent lens INSTANCES one after another.
//
// In the W2i era (run-archive/, 2026-08-22 .. 2026-09-02) the citation lens could sit as two
// instances in one epoch — `red-lens-r3-L1` and `red-lens-r3-L2`, dividing the report between
// them — and their events interleave almost completely (the quadratic run: L1 at rows 466-510,
// L2 at 467-509). The roundless record has no concurrent instances of one seat: both are
// `red-lens-evidence`, and a seat's acts belong to its SITTING, the rank of its stored opening
// among that seat's sittings (events_w). Replayed in the archived order, L1's verify would land in
// sitting 2 (L2 had registered by then) and collide with L2's on the once-per-sitting key — 53
// refusals on the quadratic run, and a wrong attribution for every act that was not refused.
//
// So within each epoch, the lowest-numbered instance keeps its place and each later instance's
// events are moved, in number order, to just after its last event. Nothing else moves: the chair, blue and the other lenses keep their places, and
// each instance keeps its own order. The migrated record then reads as what a roundless run
// would have done — the evidence lens sat twice in that epoch — and the manifest says how many
// events were moved (Result.Serialized), because the sequence is the migration's and the
// timestamps stay the archive's.
func serializeInstances(evs []OldEvent) ([]OldEvent, map[string]int) {
	// An instance is a numbered lens, and the seat it sat as is the area its number translates to
	// (lensAreaOfNumber, the table seat() reads). The round is the event's own field, which agrees
	// with the id's r<n> on every archived run.
	type seat struct {
		round int32
		area  string
	}
	instances := map[seat]map[int][]int{} // seat -> instance number -> its events' indices
	for i, ev := range evs {
		m := numberedLens.FindStringSubmatch(ev.SeatID)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		area, ok := lensAreaOfNumber(n)
		if !ok {
			continue // seat() refuses the number when the event is translated
		}
		k := seat{ev.Round, area}
		if instances[k] == nil {
			instances[k] = map[int][]int{}
		}
		instances[k][n] = append(instances[k][n], i)
	}
	moved := map[string]int{}
	pinned := map[int]bool{} // indices that leave their place
	insertAfter := map[int][]int{}
	for _, byN := range instances {
		ns := slices.Sorted(maps.Keys(byN))
		// The lowest-numbered instance keeps its place; each later one follows its last event.
		anchor := byN[ns[0]][len(byN[ns[0]])-1]
		for _, n := range ns[1:] {
			for _, i := range byN[n] {
				pinned[i] = true
				insertAfter[anchor] = append(insertAfter[anchor], i)
				moved[evs[i].SeatID]++
			}
		}
	}
	if len(pinned) == 0 {
		return evs, moved
	}
	out := make([]OldEvent, 0, len(evs))
	for i, ev := range evs {
		if pinned[i] {
			continue
		}
		out = append(out, ev)
		for _, j := range insertAfter[i] {
			out = append(out, evs[j])
		}
	}
	return out, moved
}
