package anchor

import (
	"reflect"
	"regexp"
	"testing"
)

// EVERY COLUMN OF THE KINDS TABLE IS READ. A row's meaning reaches its readers only through these
// functions, so each column is pinned per kind: deleting a row, or flipping a column, fails here.
func TestTheKindsTableSaysWhatEachKindMeans(t *testing.T) {
	for _, c := range []struct {
		id, kind, token string
		assembly        Assembly
		claim, backs    bool
	}{
		{"f-1a2b", "finding", "<!--fx:f-1a2b-->", Strip, false, false},
		{"c-1a2b", "citation", "<!--cite:c-1a2b-->", WeaveSource, true, true},
		{"p-1a2b", "proof", "<!--proof:p-1a2b-->", WeaveProof, false, true},
	} {
		if got := Kind(c.id); got != c.kind {
			t.Errorf("Kind(%q) = %q, want %q", c.id, got, c.kind)
		}
		if got := Token(c.id); got != c.token {
			t.Errorf("Token(%q) = %q, want %q", c.id, got, c.token)
		}
		if got := AssemblyOf(c.id); got != c.assembly {
			t.Errorf("AssemblyOf(%q) = %v, want %v", c.id, got, c.assembly)
		}
		if got := CountsAsClaim(c.id); got != c.claim {
			t.Errorf("CountsAsClaim(%q) = %v, want %v", c.id, got, c.claim)
		}
		if got := Backs(c.id); got != c.backs {
			t.Errorf("Backs(%q) = %v, want %v", c.id, got, c.backs)
		}
	}
	if got, want := Kinds(), []string{"finding", "citation", "proof"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Kinds() = %v, want %v — the walk order every reader reports in", got, want)
	}
	// AN ID NO ROW CLAIMS READS AS A FINDING, and its label passes through unchanged.
	if Kind("G3") != "finding" || Backs("G3") || Token("G3") != "<!--fx:G3-->" || Label("G3") != "G3" {
		t.Errorf("an unclaimed id: Kind=%q Backs=%v Token=%q Label=%q", Kind("G3"), Backs("G3"), Token("G3"), Label("G3"))
	}
	if got := Label("f-1a2b"); got != "finding-marker f-1a2b" {
		t.Errorf("Label(f-1a2b) = %q", got)
	}
}

// The id pattern matches what the table spells and nothing else.
func TestIDPatternReadsTheTable(t *testing.T) {
	id := regexp.MustCompile(`^` + IDPattern() + `$`)
	for s, want := range map[string]bool{"f-1a2b": true, "c-00": true, "p-ff": true, "g-1a": false, "c-": false, "c-XY": false, "1a2b": false} {
		if got := id.MatchString(s); got != want {
			t.Errorf("IDPattern on %q = %v, want %v", s, got, want)
		}
	}
}

// Replace and StripAssembled. The assembly strips a stripped kind's token whatever stands inside
// it — a seat may copy one into record prose with its id elided — and leaves every other kind.
func TestReplaceAndStripAssembled(t *testing.T) {
	in := "A<!--fx:f-1--> b<!--cite:c-2-->. C<!--proof:p-3-->."
	if got, want := Replace(in, func(_, id string) string { return "[" + id + "]" }), "A[f-1] b[c-2]. C[p-3]."; got != want {
		t.Errorf("Replace = %q, want %q", got, want)
	}
	if got := Replace("no anchors", func(string, string) string { return "x" }); got != "no anchors" {
		t.Errorf("Replace with no token = %q", got)
	}
	if got, want := StripAssembled("A<!--fx:f-1--> b<!--fx:…--><!--cite:c-2-->."), "A b<!--cite:c-2-->."; got != want {
		t.Errorf("StripAssembled = %q, want %q", got, want)
	}
}
