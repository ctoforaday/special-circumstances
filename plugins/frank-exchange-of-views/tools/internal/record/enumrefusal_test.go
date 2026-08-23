package record

import (
	"strings"
	"testing"
)

// THE REFUSAL NAMES THE NEAR MISS, INCLUDING THE ONE THE SURFACE'S OWN CONVENTIONS INVITE.
//
// Flag NAMES take dashes (`--reopens-on`, `--check-kind`) and VALUES take the schema's spelling,
// which is underscored because recordpb.Word derives it from the generated constant. Both
// conventions are right and they sit on the same command line, so `--as too-thin` is the mistake
// the surface asks for.
//
// Case was already detected and said so. The separator half was not: a seat got the bare legal set
// with no hint that its WORD was right and only its punctuation wrong — the one fact that turns a
// refusal into a correction. recordpb.SameWord already decides exactly this class and is what
// NearMiss uses, so the machinery was present and only the sentence was missing.
//
// The tool is the instruction: a refusal that can name the right word and does not is a turn spent
// re-reading a command that was almost correct.
func TestARefusalNamesWhichKindOfNearMissItWas(t *testing.T) {
	vs := []EnumValue{Ev("endorsed", "worth this run's time"), Ev("too_thin", "does not carry its budget"), Ev("out_of_scope", "not THIS question")}

	for _, c := range []struct{ name, got, want string }{
		{"separator", "too-thin", `"too-thin" is "too_thin" with different punctuation`},
		{"case", "TOO_THIN", `"TOO_THIN" differs from "too_thin" only in case`},
		{"separator on a longer value", "out-of-scope", `"out-of-scope" is "out_of_scope" with different punctuation`},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := Refuse("as", c.got, vs, "the ruling binds the coming seats")
			if err == nil {
				t.Fatalf("%q was accepted", c.got)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("the refusal does not name the near miss:\nwant to contain: %s\ngot: %v", c.want, err)
			}
		})
	}

	// AND A GENUINELY WRONG WORD GETS NO NEAR-MISS CLAIM. Telling a seat its value is "the right
	// word with different punctuation" when it is not sends it to look for a typo that is not
	// there — worse than the bare set, because it carries the tool's authority.
	err := Refuse("as", "banana", vs, "the ruling binds the coming seats")
	for _, unwanted := range []string{"different punctuation", "only in case"} {
		if strings.Contains(err.Error(), unwanted) {
			t.Errorf("a wrong word was reported as a near miss (%q):\n%v", unwanted, err)
		}
	}
	// It must still list what WOULD have worked.
	if !strings.Contains(err.Error(), "too_thin") {
		t.Errorf("the refusal does not carry the legal set:\n%v", err)
	}
}
