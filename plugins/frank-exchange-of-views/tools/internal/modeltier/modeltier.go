// Package modeltier is the model vocabulary: which price row a model string belongs to, which of
// two rows is dearer, whether a served model is the configured one (SameModel), and the shape of one
// seat's configured {harness, model} (SeatSpec) with its parse, its per-seat lookup and the one
// expansion of a bulk and a judgment spec into seats. What a run was CONFIGURED to use is the run's
// configuration, read through the record (record.ReadRunConfig, whose `seats` holds SeatSpec).
//
// The tier ladder lives here, apart from the prices in internal/cost, because it is a substring scan
// whose miss returns the DEAREST row: a second copy that falls behind a new model name reports a
// plausible tier rather than an error, so there is one copy. Its one dependency is the seat roster
// (internal/seatclass), which says which seats a run configures and which class each is in.
package modeltier

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// Order is the tier ladder, CHEAPEST FIRST. It is the scan order for Of and the rank for Dearer,
// and those two uses must not come apart: the JS original iterated `Object.keys(PRICES)` in
// insertion order and ranked by input price, which agreed only because the two happened to be
// sorted the same way. Here the ladder IS the order, so agreement is structural.
var Order = []string{"haiku", "sonnet", "opus", "fable"}

// Of picks the tier by substring; an UNRECOGNIZED model falls back to `fable`, the dearest row —
// an unknown model must over-report, never under-report.
func Of(m string) string {
	lower := strings.ToLower(m)
	for _, k := range Order {
		if strings.Contains(lower, k) {
			return k
		}
	}
	return "fable"
}

// Recognized reports whether the model string named a tier at all, which is the distinction Of
// deliberately flattens: Of("") and Of("gpt-9") both answer `fable`, and only this says which of
// those was a measurement.
func Recognized(m string) bool {
	lower := strings.ToLower(m)
	for _, k := range Order {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

// Rank is a tier's position on the ladder. A tier not on it ranks dearest, matching Of's fallback.
func Rank(tier string) int {
	for i, k := range Order {
		if k == tier {
			return i
		}
	}
	return len(Order)
}

// Dearer reports whether tier a costs more than tier b.
func Dearer(a, b string) bool { return Rank(a) > Rank(b) }

// SeatSpec is one seat's configured harness and model: the value of each entry of the run
// configuration's `seats` (record.RunConfig.Seats), which the record reads strictly.
type SeatSpec struct {
	Harness string `json:"harness"`
	Model   string `json:"model"`
}

// String formats SeatSpec as `<harness>:<model>`, or `<model>` if Harness is empty.
func (s SeatSpec) String() string {
	if s.Harness == "" {
		return s.Model
	}
	return s.Harness + ":" + s.Model
}

// ParseSeatSpec parses `<[harness:]model>`, defaulting a bare model string to the `claude` harness.
func ParseSeatSpec(raw string) (SeatSpec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return SeatSpec{}, fmt.Errorf("empty model specification")
	}
	if h, m, ok := strings.Cut(raw, ":"); ok {
		h = strings.TrimSpace(strings.ToLower(h))
		m = strings.TrimSpace(m)
		if h != "claude" && h != "antigravity" {
			return SeatSpec{}, fmt.Errorf("unsupported harness %q in %q (claude | antigravity)", h, raw)
		}
		if m == "" {
			return SeatSpec{}, fmt.Errorf("missing model name after %s:", h)
		}
		return SeatSpec{Harness: h, Model: m}, nil
	}
	return SeatSpec{Harness: "claude", Model: raw}, nil
}

// TieredSeats is the per-seat map a bulk and a judgment spec stand for: every configurable seat
// (Seats) on the spec of its class, the judgment spec for a judgment seat and the bulk spec for
// every other. A zero spec configures no seat of its class. It is the ONE expansion of the two
// class specs into seats: setup's --bulk/--judgment, the engine probe's tiered run, and
// feov-migrate's reshaping of an archived `model`/`judgmentModel` pair all go through it.
func TieredSeats(bulk, judgment SeatSpec) map[string]SeatSpec {
	out := map[string]SeatSpec{}
	for _, id := range Seats() {
		spec := bulk
		if seatclass.ClassOf(baseOf(id)) == "judgment" {
			spec = judgment
		}
		if spec != (SeatSpec{}) {
			out[id] = spec
		}
	}
	return out
}

// Seats is every seat a run configures a model for, sorted: the lane base (blue-lane, which
// LookupSeat resolves each numbered lane through) and every roster seat with a base.
func Seats() []string {
	keys := []string{seatclass.Lane.Base}
	for id, s := range seatclass.Seats {
		if s.Base != "" && id != seatclass.Lane.Base {
			keys = append(keys, id)
		}
	}
	sort.Strings(keys)
	return keys
}

// Configurable reports whether seatID is a seat a run configures a model for (Seats).
func Configurable(seatID string) bool {
	if seatID == seatclass.Lane.Base {
		return true
	}
	s, ok := seatclass.Seats[seatID]
	return ok && s.Base != ""
}

func baseOf(seatID string) string {
	if seatID == seatclass.Lane.Base {
		return seatclass.Lane.Base
	}
	return seatclass.Seats[seatID].Base
}

// LookupSeat returns the SeatSpec for seatID, resolving numbered lane instances (blue-lane-1, ...)
// via the shared blue-lane entry.
func LookupSeat(seats map[string]SeatSpec, seatID string) (SeatSpec, bool) {
	if len(seats) == 0 || seatID == "" {
		return SeatSpec{}, false
	}
	if spec, ok := seats[seatID]; ok {
		return spec, true
	}
	if seatclass.RoleOfSeat(seatID) == seatclass.Lane.Role {
		spec, ok := seats[seatclass.Lane.Base]
		return spec, ok
	}
	return SeatSpec{}, false
}

// SameModel reports whether a served model matches the configured model: Anthropic tier match
// when both are recognized Anthropic tiers, and exact case-insensitive ID match otherwise.
func SameModel(configured, served string) bool {
	if configured == "" || served == "" {
		return false
	}
	if Recognized(configured) && Recognized(served) {
		return Of(configured) == Of(served)
	}
	return strings.EqualFold(strings.TrimSpace(configured), strings.TrimSpace(served))
}
