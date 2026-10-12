package record

import "testing"

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
		if got := dispatchableSeatID(mustRun(t, t.TempDir()), c.id); got != c.want {
			t.Errorf("dispatchableSeatID(%q) = %v, want %v — %s", c.id, got, c.want, c.claim)
		}
	}
}
