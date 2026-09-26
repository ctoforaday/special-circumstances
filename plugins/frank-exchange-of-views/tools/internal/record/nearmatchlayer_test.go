package record

import "testing"

// AN ANCHOR IS NOT A WORD. A gap's location carries the anchor its own mint placed, so a candidate
// quoted from the report scores against a token set polluted by "fx" and the hex id unless the layer
// is stripped. The failure is a lower score, never an error — so it needs a test to be visible.
func TestNearMatchIgnoresTheAnnotationLayer(t *testing.T) {
	plain := tokenize("Eight authoritative mathematical sources were consulted")
	anchored := tokenize("Eight authoritative mathematical sources were consulted<!--fx:f-dbd94684-->")
	if len(plain) != len(anchored) {
		t.Errorf("tokenize saw %d token(s) plain and %d anchored — the anchor contributed vocabulary", len(plain), len(anchored))
	}
	for w := range anchored {
		if !plain[w] {
			t.Errorf("anchored text contributed the token %q, which the prose does not contain", w)
		}
	}
	if got := jaccard(plain, anchored); got != 1 {
		t.Errorf("jaccard(plain, anchored) = %v, want 1 — the same sentence must score identically with and without its anchor", got)
	}
}
