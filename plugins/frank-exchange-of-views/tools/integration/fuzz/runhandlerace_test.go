package fuzz

import (
	"sync"
	"testing"
)

// THE RUN HANDLE IS RESOLVED ONCE, BY THE CONSTRUCTOR, AND THIS IS WHAT SAYS SO.
//
// It was written lazily — `if !r.runHandle.Valid() { r.runHandle, _ = NewRun(...) }` — on a base
// where the seats were serial, and merged onto one where #630 phase 3 had made them concurrent.
// `envelopeFor` runs on a goroutine per seat and does NOT hold r.mu across the board reads that
// reach run(), so the lazy write is a read at run()'s guard racing a write on its next line.
// Neither branch was wrong alone: the defect existed only in the merge, which is why no test on
// either side could have caught it.
//
// WHY A DEDICATED TEST RATHER THAN TRUSTING THE SWEEP. Measured: the lazy version under
// `-race` through two full debate runs reported NOTHING. The window is real but narrow — the
// first seat to arrive usually writes before any sibling reads — so the fuzz's own scheduling
// is not a detector for it. Eight goroutines with nothing else to do is. Against the lazy
// version this fails with a data race on runHandle; against the constructor version it passes.
//
// It only detects the RACE under `-race`, and CI no longer runs one here: the race leg was
// scoped to the shipped binary's import graph (ruled 2026-09-03 — no race-instrumenting test
// harnesses; a harness race can flake a run but cannot ship). What the plain leg still gets
// from this test is the Valid() assertion; the race half fires only in a local
// `go test -race ./integration/fuzz -run TestTheRunHandleIsImmutableAfterConstruction`,
// which is where to reach when fuzz runs start flaking inexplicably.
func TestTheRunHandleIsImmutableAfterConstruction(t *testing.T) {
	r := newRunner("bin", t.TempDir(), newLockedRand(1))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = r.run() }()
	}
	wg.Wait()
	if !r.run().Valid() {
		t.Fatal("the constructor left the run handle unresolved")
	}
}
