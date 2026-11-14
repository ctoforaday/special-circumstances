// Package modeltier is the model-tier vocabulary: which price row a model string belongs to,
// and which of two rows is dearer. What tiers a run was CONFIGURED to use is the run's configuration,
// read through the record (record.ReadRunConfig).
//
// Its one importer is internal/cost, which prices transcripts by tier and judges a seat's served
// model against its configured tier (cost.JudgeServed) for capture's model-tier audit and
// `show tiers`. The ladder lives here, apart from the prices, because it is a substring scan whose
// miss returns the DEAREST row: a second copy that falls behind a new model name reports a
// plausible tier rather than an error, so there is one copy and it imports nothing.
package modeltier

import (
	"strings"
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
