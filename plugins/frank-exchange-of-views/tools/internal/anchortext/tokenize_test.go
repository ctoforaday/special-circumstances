package anchortext

import "testing"

// AN ANCHOR IS NOT A WORD. A gap's location carries the anchor its own mint placed, and an edit's
// span the anchors it carries, so text quoted from the report scores against a token set polluted
// by "fx" and the hex id unless the layer is stripped. The failure is a lower score, never an
// error — so it needs a test to be visible.
func TestTokenizeIgnoresTheAnnotationLayer(t *testing.T) {
	plain := Tokenize("Eight authoritative mathematical sources were consulted")
	anchored := Tokenize("Eight authoritative mathematical sources were consulted<!--fx:F-dbd94684-->")
	if len(plain) != len(anchored) {
		t.Errorf("Tokenize saw %d token(s) plain and %d anchored — the anchor contributed vocabulary", len(plain), len(anchored))
	}
	for w := range anchored {
		if !plain[w] {
			t.Errorf("anchored text contributed the token %q, which the prose does not contain", w)
		}
	}
	if got := Jaccard(plain, anchored); got != 1 {
		t.Errorf("Jaccard(plain, anchored) = %v, want 1 — the same sentence must score identically with and without its anchor", got)
	}
}
