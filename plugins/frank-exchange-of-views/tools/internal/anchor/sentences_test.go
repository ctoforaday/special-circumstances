package anchor

import (
	"reflect"
	"testing"
)

// ONE SENTENCE RULE. Each row is a shape a reader of an anchor's sentence met: a terminator that
// ends nothing ("3.5%", "e.g.,", "verify_91.py"), a soft wrap, a closer before an anchor, and each
// condition by which markdown opens a block. The stated splits ("Dr.", "et al.") are pinned so
// changing them is a decision.
func TestSentences(t *testing.T) {
	const c = "<!--cite:c-1-->"
	P, L, H, T, F := Paragraph, ListItem, Heading, TableRow, Fence
	for _, r := range []struct {
		in    string
		want  []string
		kinds []BlockKind // checked when set
	}{
		// Terminators.
		{"Prices rose 3.5% in 2024. Then fell.", []string{"Prices rose 3.5% in 2024.", "Then fell."}, nil},
		{"Some bases (e.g., 3) fool it.", []string{"Some bases (e.g., 3) fool it."}, nil},
		{"Run scripts/verify_91.py to check.", []string{"Run scripts/verify_91.py to check."}, nil},
		{"Sources include numbers.education and OEIS.", []string{"Sources include numbers.education and OEIS."}, nil},
		{"√91 ≈ 9.54, so we test 2–9.", []string{"√91 ≈ 9.54, so we test 2–9."}, nil},
		{"see e.g.\nfoo bar.", []string{"see e.g.\nfoo bar."}, nil},
		{"He said “Stop.” Then he left.", []string{"He said “Stop.”", "Then he left."}, nil},
		{"Dr. Smith agrees.", []string{"Dr.", "Smith agrees."}, nil},
		{"Smith et al. (2020) found X.", []string{"Smith et al.", "(2020) found X."}, nil},
		// A soft wrap is whitespace.
		{"methods\nwith zero disagreement among them. The", []string{"methods\nwith zero disagreement among them.", "The"}, nil},
		{"First.\nSecond.", []string{"First.", "Second."}, []BlockKind{P}},
		{"A list:\n- one\n- two", []string{"A list:", "one", "two"}, []BlockKind{P, L, L}},
		{"Intro.\n```\nA. B.\n```\nAfter.", []string{"Intro.", "```\nA. B.\n```", "After."}, []BlockKind{P, F, P}},
		// An anchor after a closer belongs to the closer's sentence; the boundary is tested again
		// after it. After whitespace, or flush after a bare terminator, it opens the next one.
		{"**Wet.**" + c + " Next sentence.", []string{"**Wet.**" + c, "Next sentence."}, nil},
		{"**Wet.**" + c + " and more.", []string{"**Wet.**" + c + " and more."}, nil},
		{"Seen (p. 3.)" + c + " Next.", []string{"Seen (p. 3.)" + c, "Next."}, nil},
		{"One. " + c + ". Two.", []string{"One.", c + ".", "Two."}, nil},
		{"One." + c + " Two.", []string{"One.", c + " Two."}, nil},
		// Block openers, one row per condition.
		{"Text\n# Heading\nMore", []string{"Text", "# Heading", "More"}, []BlockKind{P, H, P}},
		{"Text\n#\nMore", []string{"Text", "#", "More"}, []BlockKind{P, H, P}},
		{"Fixed in\n#552 after review. Next.", []string{"Fixed in\n#552 after review.", "Next."}, []BlockKind{P}},
		{"Values differ\n| sharply here. Then more.\nAnd so on.", []string{"Values differ\n| sharply here.", "Then more.", "And so on."}, []BlockKind{P}},
		{"| a | b |\n| --- | --- |\n| c | d |", []string{"| a | b |", "| --- | --- |", "| c | d |"}, []BlockKind{T, T, T}},
		{"The steps are\n2. compute it. Done.", []string{"The steps are\n2. compute it.", "Done."}, []BlockKind{P}},
		{"Steps:\n1. compute it.\n2. check it.", []string{"Steps:", "compute it.", "check it."}, []BlockKind{P, L, L}},
		{"- one\n2. two", []string{"one\n2. two"}, []BlockKind{L}},
		{"Intro.\n\n3. starts a list.", []string{"Intro.", "starts a list."}, []BlockKind{P, L}},
		{"Claimed here.\n[^a]: https://example.org", []string{"Claimed here.", "[^a]: https://example.org"}, []BlockKind{P, FootnoteDef}},
		// Block quotes: a continuation ">" reads as whitespace; a ">"-only line ends the paragraph;
		// openers are read after the ">"; a lazy line continues the quoted paragraph.
		{"> a long\n> sentence.", []string{"a long\n> sentence."}, []BlockKind{P}},
		{"> One.\n> Two.", []string{"One.", "Two."}, []BlockKind{P}},
		{"> a\n>\n> b", []string{"a", "b"}, []BlockKind{P, P}},
		{"> x:\n> - y", []string{"x:", "y"}, []BlockKind{P, L}},
		{"> # h\n> body text.", []string{"# h", "body text."}, []BlockKind{H, P}},
		{"Text before.\n> quoted. Here.\nlazy line.", []string{"Text before.", "quoted.", "Here.\nlazy line."}, []BlockKind{P, P}},
	} {
		var got []string
		for _, sp := range Sentences(r.in) {
			got = append(got, r.in[sp[0]:sp[1]])
		}
		if !reflect.DeepEqual(got, r.want) {
			t.Errorf("Sentences(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		if r.kinds == nil {
			continue
		}
		var kinds []BlockKind
		for _, b := range Blocks(r.in) {
			kinds = append(kinds, b.Kind)
		}
		if !reflect.DeepEqual(kinds, r.kinds) {
			t.Errorf("Blocks(%q) kinds = %v, want %v", r.in, kinds, r.kinds)
		}
	}
}
