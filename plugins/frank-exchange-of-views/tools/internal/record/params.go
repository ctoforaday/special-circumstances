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
//   - MintBudget (M): gaps one lens may mint in the run, superseding mints included.
//   - ConvergenceFraction: the board-mass fraction of the run's peak below which a FAIL over a
//     board with nothing material is refused (§III.B.2.1).
//
// A smoke run is M = 1, KMax = 2 — tighter than the old two-round smoke, and shaped like a run.
type Params struct {
	K                   int     `json:"k"`
	KMax                int     `json:"kMax"`
	MintBudget          int     `json:"mintBudget"`
	ConvergenceFraction float64 `json:"convergenceFraction"`
}

// DefaultParams are the values setup writes when the operator names none. They are a starting
// point the gate run measures (plans/roundless.md §V), stated as such.
var DefaultParams = Params{K: 2, KMax: 6, MintBudget: 5, ConvergenceFraction: 0.25}

// RunParams reads the run's terms. A run with no run-config.json, or one written before these
// were recorded, gets the defaults — a real answer, and the same one setup would have written. A
// value that is PRESENT and unusable (zero or negative, a fraction outside (0, 1]) is an error:
// a limit of zero would put every gap at impasse before its first exchange, and a reader that
// silently repaired it would hide exactly the misconfiguration the file exists to make visible.
func RunParams(run Run) (Params, error) {
	p := DefaultParams
	b, err := os.ReadFile(filepath.Join(run.Dir(), "inputs", "run-config.json"))
	if err != nil {
		return p, nil
	}
	var rc struct {
		K                   *int     `json:"k"`
		KMax                *int     `json:"kMax"`
		MintBudget          *int     `json:"mintBudget"`
		ConvergenceFraction *float64 `json:"convergenceFraction"`
	}
	if err := json.Unmarshal(b, &rc); err != nil {
		return p, fmt.Errorf("record: inputs/run-config.json does not parse: %w", err)
	}
	set := func(dst *int, v *int, name string) error {
		if v == nil {
			return nil
		}
		if *v < 1 {
			return fmt.Errorf("record: run-config.json %s = %d — a limit below 1 puts every gap at impasse before its first exchange; setup writes 1 or more", name, *v)
		}
		*dst = *v
		return nil
	}
	if err := set(&p.K, rc.K, "k"); err != nil {
		return p, err
	}
	if err := set(&p.KMax, rc.KMax, "kMax"); err != nil {
		return p, err
	}
	if err := set(&p.MintBudget, rc.MintBudget, "mintBudget"); err != nil {
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
