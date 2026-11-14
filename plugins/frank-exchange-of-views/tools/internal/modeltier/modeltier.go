// Package modeltier is the model-tier vocabulary: which price row a model string belongs to,
// which of two rows is dearer, and what tiers a run was CONFIGURED to use.
//
// # Why it is not in internal/cost
//
// It was, and it could not stay there. `cost` prices transcripts, so it imports `view`, which
// imports `record` — and the tier check now has to run at `register`, inside `record` itself.
// Importing cost from there is a cycle.
//
// The alternative was to let `record` re-derive the tier: another `strings.Contains(m, "opus")`
// ladder and another pair of bare `model` / `judgmentModel` keys read out of run-config.json.
// That is the defect this repository names most often — one fact, two hand-kept readers, drifting
// silently — and the tier ladder is exactly the kind that drifts: it is a substring scan whose
// miss returns the DEAREST row, so a copy that falls behind a new model name reports a plausible
// tier rather than an error.
//
// So the vocabulary moved down to a leaf both sides can import, and `cost` keeps only what is
// actually about money: the price rows keyed by the tier this package names.
package modeltier

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// SeatSpec is one seat's configured harness and model in inputs/run-config.json.
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

// SeatConfigs reads the per-seat {harness, model} map from inputs/run-config.json (nil if absent
// or unparseable) — the ONE place the run-config seats map is read.
func SeatConfigs(runDir string) map[string]SeatSpec {
	b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
	if err != nil {
		return nil
	}
	var cfg struct {
		Seats map[string]SeatSpec `json:"seats"`
	}
	if json.Unmarshal(b, &cfg) != nil || len(cfg.Seats) == 0 {
		return nil
	}
	return cfg.Seats
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
