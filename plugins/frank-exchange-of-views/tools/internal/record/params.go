package record

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Params are the run's TERMS (plans/roundless.md §III.B.1, §III.B.2, §III.B.2.2): the limits the
// dispatch verb and the write path read. They are set once, at setup, into inputs/run-config.json
// beside lanes — the file that already holds the run's configuration — so the limits are facts on
// the run rather than constants in a script, and a post-hoc reader can see what bound this run.
//
//   - K: consecutive exchanges on one gap with no movement before it is at impasse.
//   - KMax: total exchanges on one gap before it is at impasse regardless — the monotone bound
//     that a regrade per exchange cannot reset.
//   - MintBudget (M): the FLOOR of the gaps one lens may mint in the run, superseding mints
//     included. A lens's budget is max(M, ceil(units / per)), its units read off the record
//     at each mint — see mintScales (mintbudget.go) for each area's unit and ratio.
//   - ConvergenceFraction: the board-mass fraction of the run's peak below which a FAIL over a
//     board with nothing material is refused (§III.B.2.1).
//   - MaxEpochs: the chair sittings the run gets. The sitting that opens epoch MaxEpochs
//     dispatches nobody, and a run with parties still ready there ends CEILING (PlanDispatch).
//
// A smoke run is M = 1, KMax = 2: one mint per lens unless the report is large enough to buy more.
type Params struct {
	K                   int     `json:"k"`
	KMax                int     `json:"kMax"`
	MintBudget          int     `json:"mintBudget"`
	ConvergenceFraction float64 `json:"convergenceFraction"`
	MaxEpochs           int     `json:"maxEpochs"`
}

// DefaultParams are the values setup writes when the operator names none. They are a starting
// point the gate run measures (plans/roundless.md §V), stated as such.
var DefaultParams = Params{K: 2, KMax: 6, MintBudget: 5, ConvergenceFraction: 0.25, MaxEpochs: 12}

// RunParams reads the run's terms. A run with no run-config.json, or one written before these
// were recorded, gets the defaults — a real answer, and the same one setup would have written —
// with ONE exception: MaxEpochs is 0, no limit, where the file does not state it. An epoch limit
// the run was never held to is not a term of that run, and reading the default into it would
// derive CEILING for a run whose engine never stopped at the limit. A value that is PRESENT and
// unusable (zero or negative, a fraction outside (0, 1]) is an error: a limit of zero would put
// every gap at impasse before its first exchange, and a reader that silently repaired it would
// hide exactly the misconfiguration the file exists to make visible.
func RunParams(run Run) (Params, error) {
	p := DefaultParams
	p.MaxEpochs = 0
	b, err := os.ReadFile(filepath.Join(run.Dir(), "inputs", "run-config.json"))
	if err != nil {
		return p, nil
	}
	var rc struct {
		K                   *int     `json:"k"`
		KMax                *int     `json:"kMax"`
		MintBudget          *int     `json:"mintBudget"`
		ConvergenceFraction *float64 `json:"convergenceFraction"`
		MaxEpochs           *int     `json:"maxEpochs"`
	}
	if err := json.Unmarshal(b, &rc); err != nil {
		return p, fmt.Errorf("record: inputs/run-config.json does not parse: %w", err)
	}
	const gapLimit = "puts every gap at impasse before its first exchange"
	set := func(dst *int, v *int, name, below1 string) error {
		if v == nil {
			return nil
		}
		if *v < 1 {
			return fmt.Errorf("record: run-config.json %s = %d — a limit below 1 %s; setup writes 1 or more", name, *v, below1)
		}
		*dst = *v
		return nil
	}
	if err := set(&p.K, rc.K, "k", gapLimit); err != nil {
		return p, err
	}
	if err := set(&p.KMax, rc.KMax, "kMax", gapLimit); err != nil {
		return p, err
	}
	if err := set(&p.MintBudget, rc.MintBudget, "mintBudget", gapLimit); err != nil {
		return p, err
	}
	if err := set(&p.MaxEpochs, rc.MaxEpochs, "maxEpochs", "ends the run at its first chair sitting, before any seat is dispatched"); err != nil {
		return p, err
	}
	if rc.ConvergenceFraction != nil {
		f := *rc.ConvergenceFraction
		if !(f > 0 && f <= 1) {
			return p, fmt.Errorf("record: run-config.json convergenceFraction = %v — a fraction of the run's peak board mass, in (0, 1]", f)
		}
		p.ConvergenceFraction = f
	}
	return p, nil
}
