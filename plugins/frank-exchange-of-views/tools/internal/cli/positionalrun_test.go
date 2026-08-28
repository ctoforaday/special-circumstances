package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// A POSITIONAL RUN DIRECTORY IS REFUSED, NOT AUDITED AS AN EMPTY RUN.
//
// capture and dashboard are the two verbs that take the run as a bare argument. Nothing injects
// it — the hook's FEOV_RUN reaches --run, not argv — and nothing validated it, so a mistyped
// path resolved like any other and the verb proceeded against a run that was never dispatched.
//
// WHAT THAT ACTUALLY PRODUCED, measured by null-running this test against the pass-through:
// neither verb refused, and neither reported an empty board either. They ran on and failed
// further down for reasons that name the wrong thing entirely — capture on a missing
// journal.jsonl in the TRANSCRIPT directory, dashboard on being unable to write dashboard.html
// into a directory that does not exist. An operator reading either message goes looking at the
// transcripts or at permissions, and the run directory they actually mistyped is not mentioned.
//
// So the claim here is narrow and is the one the evidence supports: the refusal names the run.
// It is #526's shape in the one place seat-resolution could not reach — the value never passes
// through seat.Context at all — without asserting a board of zeros nobody has observed.
//
// NOTE the limit of the check: OpenRun refuses a path that does not EXIST. `--run /tmp` names a
// real directory that is not a run, and is still accepted.
func TestAPositionalRunDirectoryNobodyMadeIsRefused(t *testing.T) {
	for _, verb := range []string{"capture", "dashboard"} {
		t.Run(verb, func(t *testing.T) {
			typo := filepath.Join(t.TempDir(), "reserach", "2026-01-01_typo")
			// --seat-id operator SELECTS the operator surface. "No identity" is deliberately not
			// a mode here, so omitting it yields an empty tree and a usage dump — a fact about
			// the harness that reads as a fact about the verb.
			out, err := run(t, verb, typo, t.TempDir(), "--seat-id", "operator")
			if err == nil {
				t.Fatalf("%s accepted a path nobody dispatched:\n%s", verb, out)
			}
			// The refusal must be about the RUN. Asserting only err != nil would pass on any
			// unrelated failure — a missing transcript directory, a usage error — and this verb
			// has several.
			if !strings.Contains(err.Error(), "names no run") {
				t.Errorf("%s failed for some other reason than the unresolvable run: %v", verb, err)
			}
		})
	}
}
