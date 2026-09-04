package fuzz

import (
	"os"
	"testing"
)

// releaseGate marks a test as a RELEASE-BOUNDARY sweep: it runs on the nightly and on a
// feov tag (both set FEOV_RELEASE_GATE=1) and SKIPS on pull requests, where only this
// package's cheap literal-surface guards run (ruled 2026-09-04; the directory name states
// the same fact).
//
// A RUNTIME SKIP, NOT A BUILD TAG, on two grounds. The sweeps keep compiling on every pull
// request — a tag would let them rot unbuilt — and a skip is VISIBLE in the output where an
// untagged absence is indistinguishable from a pass. What a skip cannot do is prove the
// release leg really runs the sweep: an env var typo'd in the workflow would skip there too,
// forever, green. That is the plausible zero, and it is answered where it lives — the
// release step greps for the sweep's own `--- PASS` line and fails loudly without it.
func releaseGate(t *testing.T) {
	t.Helper()
	if os.Getenv("FEOV_RELEASE_GATE") == "" {
		t.Skip("release-boundary sweep: runs on the nightly and on a feov tag (FEOV_RELEASE_GATE=1)")
	}
}
