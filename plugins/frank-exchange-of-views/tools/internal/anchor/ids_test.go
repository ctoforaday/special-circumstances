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

// SkipRun steps over exactly the tokens the kinds table spells — a run of them whole, and a
// malformed one not at all, since a token one reader steps over and another refuses to see is a
// span where punctuation gets tidied around something that is not an anchor.
func TestSkipRunReadsTheKindsTable(t *testing.T) {
	for _, tok := range []string{
		"<!--fx:f-seed01-->",   // "s" is not hex — the malformed id that made the fuzzer vacuous
		"<!--fx:f-ABC123-->",   // uppercase is not [0-9a-f]
		"<!--fx:abc123-->",     // no kind prefix on the id
		"<!--cite:f-abc123-->", // the id's prefix is not the tag's kind
		"<!--note:f-abc123-->", // not a tag in the table
		"<!--fx:f--->",         // empty id
		"<!-- fx:f-abc123 -->", // spaced: not the minted spelling
	} {
		if got := SkipRun(tok, 0); got != 0 {
			t.Errorf("SkipRun(%q) = %d, want 0", tok, got)
		}
	}
	pair := Token("f-abc123") + Token("c-d4d4d4") + Token("p-0f0f0f")
	if got := SkipRun(pair+"tail", 0); got != len(pair) {
		t.Errorf("SkipRun over an abutting run = %d, want %d (%q)", got, len(pair), pair)
	}
}
