package terms

import (
	"reflect"
	"testing"
)

// THE TABLE IS THE REGISTRY'S, ENTRY FOR ENTRY, in registry order: an edit to terms.json that was
// not regenerated fails here as well as at vocabgen -check.
func TestTheDefinitionTableIsTheRegistry(t *testing.T) {
	reg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(definitions) != len(reg.Entries) {
		t.Fatalf("the table holds %d definitions and the registry %d entries — run `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`", len(definitions), len(reg.Entries))
	}
	for i, e := range reg.Entries {
		if want := DefinitionOf(e); !reflect.DeepEqual(definitions[i], want) {
			t.Errorf("definition %d is %+v and the registry's entry is %+v — run `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`", i, definitions[i], want)
		}
	}
	if def, ok := Defined(reg.Entries[0].Term); !ok || def != reg.Entries[0].Definition {
		t.Errorf("Defined(%q) = %q, %t", reg.Entries[0].Term, def, ok)
	}
}
