package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ID SPELLED ANOTHER WAY IS NEVER TOLD THE RECORD LACKS IT.
//
// MEASURED on universe-m37: `mint --about '<!--computation:K-050249f6-->'` was refused as naming what
// "no citation or computation on this record carries", and K-050249f6 was on the record. Every flag
// here takes an id, each row names one the record holds, and each either takes the spelling and
// stores the bare id or refuses it for its spelling.

// heldRecord is a run holding a citation and a computation, with the lens registered.
func heldRecord(t *testing.T) (runDir, cite, computation string) {
	t.Helper()
	runDir = citedRun(t, "https://held/1")
	cite = citeSentence(t, runDir, "The sky is blue.", "https://held/1")
	computation = computeSentence(t, runDir, "The grass is green.")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	return runDir, cite, computation
}

func mintWith(t *testing.T, runDir string, extra ...string) (string, error) {
	t.Helper()
	return runMint(t, append([]string{"--run", runDir, "--seat-id", lensSeat, "--class", "spec_underspecification",
		"--problem", "p", "--fix", "f", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low"}, extra...)...)
}

// The words of an existence refusal, each of which says the record lacks the id it was handed.
var lacksWords = []string{"no citation or computation", "no mint event", "no lens recorded", "no act on this record", "no inquiry event"}

func refusedForItsSpelling(t *testing.T, err error, value string, says string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%q was accepted", value)
	}
	for _, lacks := range lacksWords {
		if strings.Contains(err.Error(), lacks) {
			t.Errorf("%q was refused as absent from the record: %v", value, err)
		}
	}
	if !strings.Contains(err.Error(), says) {
		t.Errorf("the refusal of %q does not say %q: %v", value, says, err)
	}
}

func TestAboutTakesAnAnchorTokenAndStoresTheBareID(t *testing.T) {
	for _, c := range []struct {
		name  string
		value func(cite, computation string) (value, stored string)
	}{
		{"a computation's token", func(_, k string) (string, string) { return anchor.Token(k), k }},
		{"a citation's token", func(c, _ string) (string, string) { return anchor.Token(c), c }},
		{"a computation's bare id", func(_, k string) (string, string) { return k, k }},
	} {
		t.Run(c.name+" on a gap", func(t *testing.T) {
			// A run per case: one class on one anchor is one gap.
			runDir, cite, computation := heldRecord(t)
			value, stored := c.value(cite, computation)
			if out, err := mintWith(t, runDir, "--about-kind", "anchor", "--about", value); err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			if got := lastBody(t, runDir, &recordpb.Mint{}).GetAboutRef(); got != stored {
				t.Errorf("about_ref is %q, want %q", got, stored)
			}
		})
	}
}

func TestAboutRefusesAnotherSpellingForItsSpelling(t *testing.T) {
	runDir, _, computation := heldRecord(t)
	gap, err := mintQuote(t, runDir, "G", "Fire is hot.")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ name, kind, value, says string }{
		{"the text inside a token's delimiters", "anchor", anchor.Tagged(computation), flags.BareID("K-…")},
		{"a gap's id", "anchor", gap, "`G-` is a gap"},
		{"a citation's id under inquiry", "inquiry", anchor.Token(computation), "an inquiry id is Q-"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := mintWith(t, runDir, "--about-kind", c.kind, "--about", c.value)
			refusedForItsSpelling(t, err, c.value, c.says)
		})
	}
}

// THE EXISTENCE REFUSAL STILL STANDS for an id of the right shape the record does not hold.
func TestAboutRefusesAWellSpelledIDTheRecordDoesNotHold(t *testing.T) {
	runDir, _, _ := heldRecord(t)
	for _, c := range []struct{ kind, value, says string }{
		{"anchor", "K-ffffffff", "no citation or computation on this record carries"},
		{"anchor", anchor.Token("K-ffffffff"), "K-ffffffff, which no citation or computation"},
		{"inquiry", "Q-ffffffff", "--about names inquiry Q-ffffffff, which no inquiry event proposed"},
	} {
		if _, err := mintWith(t, runDir, "--about-kind", c.kind, "--about", c.value); err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("--about-kind %s --about %s answered %v, want %q", c.kind, c.value, err, c.says)
		}
	}
}

// A RETIRE NAMING AN ANCHOR BY ITS TOKEN RETIRES THAT ANCHOR. The token passed the shape and was then
// used as the id, so the report was searched for a token wrapped in a token, and the refusal said the
// report held none.
func TestRetireTakesAnAnchorByItsToken(t *testing.T) {
	runDir := citedRun(t, "https://held/1")
	cite := citeSentence(t, runDir, "The sky is blue.", "https://held/1")
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", anchor.Token(cite), "--reason", "the source does not hold"); err != nil {
		t.Fatalf("retire --anchor with the citation's token: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}).GetAnchors(); len(got) != 1 || got[0] != cite {
		t.Errorf("the retire names %q, want the bare id", got)
	}
}

// EVERY LIST FLAG THAT TAKES IDS HOLDS ITS ITEMS TO A SHAPE, read off the built trees: one left plain
// hands a mistyped id to an existence check, which answers that the record lacks it.
func TestEveryIDListFlagHoldsItsItemsToAShape(t *testing.T) {
	seen := 0
	for _, root := range AllRoots() {
		walk(root, func(c *cobra.Command, path []string) {
			c.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if _, list := f.Value.(*flags.CSV); !list {
					return
				}
				switch f.Name {
				case flags.IDs, flags.Anchor:
				default:
					return
				}
				seen++
				if err := f.Value.Set("banana"); err == nil {
					t.Errorf("%v --%s took an item that is no id's shape", path, f.Name)
				}
			})
		})
	}
	if seen == 0 {
		t.Error("the walk reached no id list")
	}
}
