package migrate

import (
	"regexp"
	"sort"
	"strconv"
)

// serializeInstances puts an archived run's concurrent lens INSTANCES one after another.
//
// In the W2i era (run-archive/, 2026-08-22 .. 2026-09-02) the citation lens could sit as two
// instances in one round — `red-lens-r3-L1` and `red-lens-r3-L2`, dividing the report between
// them — and their events interleave almost completely (the quadratic run: L1 at rows 466-510,
// L2 at 467-509). The roundless record has no concurrent instances of one seat: both are
// `red-lens-evidence`, and a seat's acts belong to its SITTING, which is the count of its
// registers so far (events_w). Replayed in the archived order, L1's verify would land in sitting
// 2 (L2 had registered by then) and collide with L2's on the once-per-sitting key — 53 refusals
// on the quadratic run, and a wrong attribution for every act that was not refused.
//
// So within each round, instance 2's events (then 3's, 4's) are moved to just after instance 1's
// last event. Nothing else moves: the chair, blue and the other lenses keep their places, and
// each instance keeps its own order. The migrated record then reads as what a roundless run
// would have done — the evidence lens sat twice in that epoch — and the manifest says how many
// events were moved (Result.Serialized), because the sequence is the migration's and the
// timestamps stay the archive's.
func serializeInstances(evs []OldEvent) ([]OldEvent, map[string]int) {
	instance := regexp.MustCompile(`^red-lens-(r\d+)-L([1-4])$`)
	type key struct{ round, n string }
	first := map[string]int{} // round -> index of instance 1's last event
	others := map[string]map[string][]int{}
	for i, ev := range evs {
		m := instance.FindStringSubmatch(ev.SeatID)
		if m == nil {
			continue
		}
		if m[2] == "1" {
			first[m[1]] = i
			continue
		}
		if others[m[1]] == nil {
			others[m[1]] = map[string][]int{}
		}
		others[m[1]][m[2]] = append(others[m[1]][m[2]], i)
	}
	moved := map[string]int{}
	pinned := map[int]bool{} // indices that leave their place
	insertAfter := map[int][]int{}
	for round, byN := range others {
		anchor, ok := first[round]
		if !ok {
			continue // a second instance with no first: nothing to serialize against
		}
		var ns []string
		for n := range byN {
			ns = append(ns, n)
		}
		sort.Slice(ns, func(a, b int) bool { x, _ := strconv.Atoi(ns[a]); y, _ := strconv.Atoi(ns[b]); return x < y })
		for _, n := range ns {
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
