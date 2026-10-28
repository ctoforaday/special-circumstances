package anchor

import "testing"

// A HUSK IS A LIST MARKER AT MOST, SENTENCE PUNCTUATION AND THE WRAPPERS THAT HELD THE TEXT, and
// nothing else is: assembly and the replay of a retire both delete what this calls a husk, so
// every line here that is not one is a line of a report that one of them would otherwise take,
// and every line here that is one is a line a reader would otherwise be shown — "()", or the
// rule a renderer draws for "****".
func TestAHuskIsMarkersAndSentencePunctuation(t *testing.T) {
	for _, s := range []string{"", "   ", "- ", "- ?", "* .", "+ …", "> ", "> - .", ".", " . ", "?!", ";",
		"()", "(.)", "[]", "**", "**.**", "****", "__", "*.*", "_?_", `""`, "''", "“.”", "‘’", "(**.**)", "- **", "> - (“”).", "() **"} {
		if !Husk(s) {
			t.Errorf("Husk(%q) is false; a marker and sentence punctuation are what an anchor leaves", s)
		}
	}
	for _, s := range []string{
		"| --- | --- |", "|:---|---:|", "| ✓ | ✗ |", "| — | ✓ |", "---", "***", "    }", "→", "—", "- [ ]", "1.", "a",
		"* * *", "* *", "- - -", "___", "(", ")", "*", "(a)", "**a**", "| () | ✓ |", "(]", "“", "( )", `"`,
		"- " + Token("F-00000001"), Token("C-00000001") + ".",
	} {
		if Husk(s) {
			t.Errorf("Husk(%q) is true; that is content, or still holds an anchor", s)
		}
	}
}

// A STRIP TAKES EACH TOKEN AND EACH LINE IT LEAVES A HUSK, AND NOTHING ELSE. Assembly strips with it,
// and the release gate's prose oracle reads a seat's field through assembly's strip, so both see the
// same text: a bullet cut down to its anchors goes whole, a line of content that held a token keeps
// the content, and a token keep reports true for — one in code, one the weave turns into a note —
// stays where it stands.
func TestStripWrittenTakesTokensAndTheHusksTheyLeave(t *testing.T) {
	f, c := Token("F-00000001"), Token("C-00000001")
	in := "kept line\n- " + f + ".\n| --- " + f + "|\nclaim" + c + " here"
	if got, want := StripWritten(in, nil), "kept line\n| --- |\nclaim here"; got != want {
		t.Errorf("StripWritten(%q, nil) = %q, want %q", in, got, want)
	}
	keepCite := func(start, end int) bool { return in[start:end] == c }
	if got, want := StripWritten(in, keepCite), "kept line\n| --- |\nclaim"+c+" here"; got != want {
		t.Errorf("StripWritten keeping the citation = %q, want %q", got, want)
	}
	if got := StripWritten("no token\n- .", nil); got != "no token\n- ." {
		t.Errorf("a line that held no token was taken: %q", got)
	}
}
