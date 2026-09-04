package fuzz

import (
	"strings"
	"testing"
)

// THE GOJA-LEVEL HALF OF #639's REFUSAL, WHICH ITS OWN TESTS DO NOT REACH.
//
// `readVerdict` is covered thoroughly by TestReadVerdictRefusesAResultThatCarriesNoVerdict, and
// that is the right level for the map it inspects. But driveDebate has a SEPARATE refusal one
// layer down — a promise that fulfils with something that is not an object at all — and nothing
// drives it. It is the branch that produced the shape #637 measured: an unreadable fulfilment
// left `result` nil AND `settledErr` empty, so every caller read it as a run that succeeded and
// reported nothing, which is byte-identical to a genuinely failed run's empty verdict key.
//
// Reaching it needs the real event loop, so these drive driveDebate itself. No seat is
// dispatched, so no binary is needed and the whole file runs in milliseconds.
func TestDriveDebateRefusesAFulfilmentItCannotRead(t *testing.T) {
	for _, tc := range []struct {
		name    string
		wrapped string
		want    string
	}{
		{"a number", `globalThis.__result = Promise.resolve(42);`, "not an object"},
		{"a string", `globalThis.__result = Promise.resolve("done");`, "not an object"},
		{"null", `globalThis.__result = Promise.resolve(null);`, "not an object"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &runner{runDir: t.TempDir(), rng: newLockedRand(1), registered: map[string]bool{}}
			result, settledErr := driveDebate(r, tc.wrapped)
			if result != nil {
				t.Errorf("result should be nil when the fulfilment is unreadable, got %v", result)
			}
			if !strings.Contains(settledErr, tc.want) {
				t.Errorf("settledErr = %q, want it to name %q — a fulfilment this side cannot read must not pass as an empty run", settledErr, tc.want)
			}
		})
	}
}

// THE POSITIVE CONTROL, so the refusal above is not passing because driveDebate refuses
// everything. A readable fulfilment still comes back as a map with no error.
func TestDriveDebateAcceptsAReadableFulfilment(t *testing.T) {
	r := &runner{runDir: t.TempDir(), rng: newLockedRand(1), registered: map[string]bool{}}
	result, settledErr := driveDebate(r, `globalThis.__result = Promise.resolve({verdict: "VERIFIED", rounds: 2});`)
	if settledErr != "" {
		t.Fatalf("settledErr = %q, want empty for a readable fulfilment", settledErr)
	}
	if v, _ := result["verdict"].(string); v != "VERIFIED" {
		t.Errorf("verdict = %q, want VERIFIED", v)
	}
}
