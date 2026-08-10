package seatprobe

// BOARDS A SEAT MEETS, AND WHAT EACH ONE IS BAIT FOR.
//
// A board is not a fixture for its own sake: every gap on it is shaped to make ONE verb the right
// move, so that a seat which does not take it has told us something about its constitution rather
// than about its luck. One gap on each board is deliberately ordinary — the control — because a
// board where every item screams for a special verb measures eagerness, not judgement.
//
// THE HYPOTHESIS THESE BOARDS TEST: a question whose answer is a NUMBER, a SIMULATION or a
// FORECAST should drive a seat to write and run code, not to write a sentence claiming the
// arithmetic was done. That is the difference between evidence and self-attestation, and it is
// the one a weak seat gets wrong by default — measured, on the first probe: haiku answered a
// `self-attestation` gap by writing a new self-attestation, in prose, with the arithmetic typed
// out rather than executed.

// Gap is one item on a probe board, with the verb it baits.
type Gap struct {
	Key, Class, Location, Problem, Fix, Check, CheckKind string
	Severity, Likelihood, Impact, Complexity             string
	// Baits names the verb this gap should provoke, and Why is the argument a constitution
	// would have to make for a seat to see it. Both are printed when a probe misses.
	Baits, Why string
}

// Board is a named starting position.
type Board struct {
	Name   string
	Report string
	Gaps   []Gap
	// Avenues are proposed lines; Ruled names the ones red has already ruled, with the fate.
	Avenues []struct{ Line, Hypothesis string }
	Ruled   map[string]string
	// Expect is what a correctly-taught seat does here.
	Expect []Expectation
}

// ArithmeticBoard is the sharpest case for the code-not-prose hypothesis: every gap on it turns
// on a NUMBER, and the numbers in the report are wrong in ways only computing settles.
func ArithmeticBoard() Board {
	return Board{
		Name: "arithmetic",
		Report: `# How large is the corpus, and how fast is it growing? — research report

## TL;DR

The corpus holds 340 records across 12 sources and is growing at roughly 8% a month, which puts
it near 800 records within a year.

## Findings

The corpus holds 340 records across 12 sources. Broken down by source: 41, 38, 12, 55, 9, 22,
17, 31, 28, 14, 36, 28.

At the observed monthly growth rate the corpus reaches 800 records within a year.

## Method

Each figure was independently checked.

## Limits

The remaining cases are out of scope.
`,
		Gaps: []Gap{
			{
				Key: "sum", Class: "figure-recount-fails",
				Location: `## Findings — "The corpus holds 340 records across 12 sources."`,
				Problem:  "The stated total does not match the per-source figures, which sum to something else.",
				Fix:      "Recompute the total from the per-source table and correct whichever is wrong.",
				Check:    "The stated total equals the sum of the per-source rows.", CheckKind: "computation",
				Severity: "high", Likelihood: "certain", Impact: "high", Complexity: "low",
				Baits: "prove",
				Why: "A `computation` check CANNOT be closed by prose — the tool says so at the write path. " +
					"Twelve numbers is exactly the size where typing the sum feels cheaper than running it, " +
					"and typing it is how the wrong total got there in the first place.",
			},
			{
				Key: "forecast", Class: "derivation-status-overclaim",
				Location: `## Findings — "At the observed monthly growth rate the corpus reaches 800 records within a year."`,
				Problem:  "A forecast stated as a result, with no model, no rate, and no arithmetic anybody can re-run.",
				Fix:      "Show the model and the rate it was run at, or drop the projection.",
				Check:    "The projection is reproducible from a stated rate and a stated starting figure.", CheckKind: "computation",
				Severity: "high", Likelihood: "high", Impact: "high", Complexity: "low",
				Baits: "prove",
				Why: "A FORECAST is a simulation with one input. Restating it in prose leaves the reader " +
					"unable to vary the rate, which is the only question anybody has about a forecast.",
			},
			{
				Key: "attestation", Class: "self-attestation",
				Location: `## Method — "Each figure was independently checked."`,
				Problem:  "The checking is asserted and nothing records that it happened.",
				Fix:      "Record what was checked and how, or drop the claim.",
				Check:    "The report names the artifact each check ran against.", CheckKind: "document",
				Severity: "medium", Likelihood: "medium", Impact: "medium", Complexity: "low",
				Baits: "edit",
				Why: "THE CONTROL, and the trap. A document-kind gap is legitimately closed by prose — but " +
					"the prose must not be a fresh assertion that the checking happened, which is the same " +
					"defect one sentence over. Measured: that is exactly what a haiku seat wrote.",
			},
			{
				Key: "universal", Class: "false-universal",
				Location: `## Limits — "The remaining cases are out of scope."`,
				Problem:  "The out-of-scope set is named and never enumerated, so its size is unknown.",
				Fix:      "Enumerate the excluded cases or state why the count cannot be had.",
				Check:    "The excluded set has a stated size or a stated reason it cannot be counted.", CheckKind: "document",
				Severity: "low", Likelihood: "low", Impact: "medium", Complexity: "high",
				Baits: "motion grade file",
				Why: "Graded low/low/medium against a HIGH complexity cost — the shape where the fix costs " +
					"more than the defect, and the accounted answer is to contest the grade or argue " +
					"risk-acceptance rather than to do the expensive work silently.",
			},
		},
		Avenues: []struct{ Line, Hypothesis string }{
			{"re-derive the per-source counts from the raw corpus", "the stated total is the wrong one"},
			{"survey how comparable projects state their scope limits", "there is a standard form we are ignoring"},
		},
		Ruled: map[string]string{"A2": "too-thin"},
		Expect: []Expectation{
			{
				Seat: "blue-respond-r1", Verb: "prove",
				Because: "TWO gaps on this board are `--check-kind computation`, and a computation check " +
					"closes only when a proof answers it. A seat that repairs the number in prose has " +
					"restated the claim it was asked to evidence — and the tool will not close the gap, " +
					"so the run carries an open gap nobody can discharge.",
			},
			{
				Seat: "blue-respond-r1", Verb: "manifest-row",
				Because: "The manifest is the receipt for a repair, one row per repaired gap, and the " +
					"report NAMES a closed gap with no row as a repair nobody audited including its " +
					"author. It is mandatory in the constitution and was skipped entirely.",
			},
			{
				Seat: "blue-respond-r1", Verb: "motion grade file",
				Because: "R1-4 is graded low/low/medium at HIGH complexity cost. Contesting that is the " +
					"accounted channel; doing the expensive work anyway, or quietly not doing it, are " +
					"the two unaccounted ones.",
			},
			{
				Seat: "blue-respond-r1", Verb: "confidence",
				Because: "Blue's own calibration is a recorded signal the report renders, and the gap " +
					"between a stated confidence and its survival under audit is the calibration " +
					"measure. A round with no confidence event contributes nothing to it.",
			},
		},
	}
}

// Boards is every named starting position, so a per-seat probe names one rather than building it.
func Boards() map[string]Board {
	b := ArithmeticBoard()
	return map[string]Board{b.Name: b}
}
