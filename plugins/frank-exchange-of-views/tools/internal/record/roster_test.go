package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE COMMENTS ABOVE THE GATE ARE CLAIMS ABOUT IT, AND THEY ROTTED. Measured 2026-09-16 on the
// tree as it then stood: roster.go said the gate "bounds the SHAPE, never the membership" and gave
// `red-lens-L4` as a well-formed id, agentrole.go repeated it with `red-lens-L9` — and
// dispatchableSeatID refused both, because the lens pattern is lower-case and the area must be in
// LensAreas. Forty lines below the first sentence, the same file said the opposite under
// "MEMBERSHIP, NOT JUST SHAPE". A file that states a rule and its own contradiction teaches whichever
// paragraph the reader happens to open at.
//
// So the corrected sentences are pinned here rather than left as prose: each case below is one
// claim those comments now make, and rewording a comment without this table failing is how the drift
// gets in again. `red-merge-r1` is kept because agentrole.go used to cite it: it names both the
// retired `-r<N>-` sitting segment and the `merge` role that #831 renamed to `chair`.
func TestTheSeatIdCommentsDescribeWhatTheGateDoes(t *testing.T) {
	for _, c := range []struct {
		id    string
		want  bool
		claim string
	}{
		{"red-lens-evidence", true, "a real area is dispatchable"},
		{"red-lens-evidence-oops", false, "the area is bounded by membership, so a hyphenated near-miss is refused"},
		{"red-lens-L4", false, "the lens pattern is lower-case: an old-style L<n> id is not well formed"},
		{"red-chair", true, "the chair keeps its party prefix and IS dispatched — which is why the attestation table, not the shape, is what catches the wrong agent registering as it"},
		{"red-merge-r1", false, "neither the retired -r<N>- sitting segment nor the pre-#831 merge role survives"},
	} {
		if got, err := dispatchableSeatID(mustRun(t, t.TempDir()), c.id); err != nil || got != c.want {
			t.Errorf("dispatchableSeatID(%q) = %v, want %v — %s", c.id, got, c.want, c.claim)
		}
	}
}

// A CAST THAT CANNOT BE READ IS NOT A CAST THAT NAMES NO LANES. The register gate asks the cast
// whether a lane id is one this run seats; where the record cannot be read, the register is
// refused with the read error named — not told its id is one no dispatch produces, which sends a
// seat to re-copy an id it had right. The record here holds a former format's shard and no
// database: a run this binary cannot read.
func TestALaneRegisterOverAnUnreadableCastNamesTheRead(t *testing.T) {
	dir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "records", "events-blue-0123abcd.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run := mustRun(t, dir)
	err := requireDispatchableSeat(run, "blue-lane-1")
	if err == nil || !strings.Contains(err.Error(), "could not be read to establish whether") ||
		!strings.Contains(err.Error(), "FORMER record format") {
		t.Errorf("a lane register over an unreadable cast = %v, want the refusal naming the read", err)
	}
	if c, err := TierClassOfSeat(run, "blue-lane-1"); err == nil {
		t.Errorf("TierClassOfSeat over an unreadable cast = %q with no error — \"\" says the seat rides no tier", c)
	}
	// A roster seat is answered by the roster, so the cast is never asked.
	if err := requireDispatchableSeat(run, "red-chair"); err != nil {
		t.Errorf("a roster seat was refused over the cast it never reads: %v", err)
	}
}
