package anchor

import (
	"reflect"
	"testing"
)

// IDs READS WHAT Token WRITES, on the shapes the corpus actually holds.
func TestIDsReadsWhatTokenWrites(t *testing.T) {
	for _, id := range []string{"f-e4bc25ec", "c-a1b2c3d4", "p-99887766"} {
		if got := IDs("text " + Token(id) + " more"); !reflect.DeepEqual(got, []string{id}) {
			t.Errorf("IDs(Token(%q)) = %v, want [%s]", id, got, id)
		}
	}
	// ABUTTING ANCHORS ARE REAL — two lenses anchored one sentence in the 2026-08-22 corpus.
	both := Token("f-e4bc25ec") + Token("f-73a56bd3")
	if got, want := IDs("verification"+both), []string{"f-e4bc25ec", "f-73a56bd3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("abutting anchors = %v, want %v", got, want)
	}
	// Deduplicated: one anchor quoted twice is one anchor.
	if got := IDs(Token("c-a1b2c3d4") + " x " + Token("c-a1b2c3d4")); len(got) != 1 {
		t.Errorf("a repeated anchor = %v, want one entry", got)
	}
	// A STRAY HTML COMMENT IS NOT AN ANCHOR, and a non-hex id is not one either.
	for _, s := range []string{"<!--just a comment-->", "<!--cite:c-zzzz-->", "<!--fx:nope-->"} {
		if got := IDs(s); len(got) != 0 {
			t.Errorf("IDs(%q) = %v, want none", s, got)
		}
	}
	if Kind("c-1") != "citation" || Kind("p-1") != "proof" || Kind("f-1") != "finding" {
		t.Error("Kind disagrees with the prefix classes Token spells")
	}
}
