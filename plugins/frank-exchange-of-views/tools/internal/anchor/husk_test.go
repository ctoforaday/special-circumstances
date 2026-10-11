package anchor

import "testing"

// A HUSK IS A LIST MARKER AT MOST AND SENTENCE PUNCTUATION, and nothing else is: assembly and the
// replay of a retire both delete what this calls a husk, so every line here that is not one is a
// line of a report that one of them would otherwise take.
func TestAHuskIsMarkersAndSentencePunctuation(t *testing.T) {
	for _, s := range []string{"", "   ", "- ", "- ?", "* .", "+ …", "> ", "> - .", ".", " . ", "?!", ";"} {
		if !Husk(s) {
			t.Errorf("Husk(%q) is false; a marker and sentence punctuation are what an anchor leaves", s)
		}
	}
	for _, s := range []string{
		"| --- | --- |", "|:---|---:|", "| ✓ | ✗ |", "| — | ✓ |", "---", "***", "    }", "→", "—", "()", "**", "- [ ]", "1.", "a",
		"- " + Token("F-00000001"), Token("C-00000001") + ".",
	} {
		if Husk(s) {
			t.Errorf("Husk(%q) is true; that is content, or still holds an anchor", s)
		}
	}
}
