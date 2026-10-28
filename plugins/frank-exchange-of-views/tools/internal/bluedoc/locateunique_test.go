package bluedoc

import (
	"strings"
	"testing"
)

// EDIT'S AMBIGUITY REFUSAL NAMES EACH SITE, as the placing verbs' does: an edit whose quote repeats
// is told where each copy stands, so it can quote the one it means or make one edit per site.
func TestLocateUniqueNamesEachSiteOfARepeatedQuote(t *testing.T) {
	_, _, err := LocateUnique("# A\n\nSo costs rose.\n\n# B\n\nThen costs rose.\n", "costs rose")
	if err == nil {
		t.Fatal("a repeated quote was located")
	}
	for _, want := range []string{"MORE THAN ONCE", "it stands 2 times", `under the heading "A"`, `under the heading "B", in the sentence "Then costs rose"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q:\n%v", want, err)
		}
	}
}
