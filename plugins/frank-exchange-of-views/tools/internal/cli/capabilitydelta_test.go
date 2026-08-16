package cli

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE TABLE MUST NOT ROT THE WAY THE CHANGELOG DID (#407).
//
// The changelog this replaced was hand-kept with no mechanism: entries were added by habit and
// nothing failed when one was skipped. This is the mechanism — `Version` moving without a delta
// is exactly the silence a refused preflight cannot afford.
//
// IT LIVES IN PACKAGE cli, NOT package record, and the reason is a trap worth recording.
// record.ToolVersion is a mutable var that cli's init assigns cli.Version INTO; in a
// record-only test binary that init never runs, so it reads its birth value 0.1.0 and the
// check would assert against a version nothing ships. The first draft of this test did exactly
// that and failed pointing at 0.1.0 — the same stale-per-binary-constant class #408 catalogued,
// caught by the check it was written for.
func TestTheRunningVersionDeclaresItsCapabilityDelta(t *testing.T) {
	if Version == "" {
		t.Fatal("cli.Version is empty, so this check reads nothing and would pass on any table")
	}
	if record.CapabilityDelta(Version) == "" {
		t.Errorf("version %s has no capability delta.\n"+
			"A release with no capability change must SAY so (\"no capability change\") rather "+
			"than omit the entry — an absent entry and an unremarkable release must not be the "+
			"same bytes, which is how the changelog this replaced went stale.", Version)
	}
}
