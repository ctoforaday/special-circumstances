package fuzz

import (
	"strings"
	"testing"
)

// A DEBATE THAT SETTLES FULFILLED WITHOUT A VERDICT MUST BE A FAILURE, NOT A MAP KEY.
//
// Both reads off the result map were `if v, ok := ...; ok {}` with no else, so a fulfilment
// this side could not read left `result` nil and `settledErr` empty, and runOne recorded a
// PASSING run whose verdict was "". The sweep's tally then carried that "" next to the empty
// verdicts of genuinely failed runs — `verdicts=map[:29 CEILING:8 VERIFIED:3]` in #637 — with
// nothing in the line able to say which kind it had counted.
//
// These drive driveDebate directly: no seat is dispatched, so no binary is needed and the
// whole file runs in milliseconds. The point is the refusal, which is otherwise a branch
// nothing executes until the next incident.
func TestDriveDebateRefusesAFulfilmentItCannotRead(t *testing.T) {
	for _, tc := range []struct {
		name    string
		wrapped string
		want    string
	}{
		{"a number", `globalThis.__result = Promise.resolve(42);`, "non-object result"},
		{"a string", `globalThis.__result = Promise.resolve("done");`, "non-object result"},
		{"null", `globalThis.__result = Promise.resolve(null);`, "non-object result"},
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
