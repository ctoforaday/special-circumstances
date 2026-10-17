package cli

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// EVERY VIEW THAT ADVERTISES --match AND --quote SELECTS WITH THEM.
//
// The census is taken from the tree the seat is handed, not from a list kept here: every projection
// under `show` and `inquest` whose page carries the pair is a row. `show inquiries` carried the pair
// for three weeks and answered every pattern with all six inquiries, in both forms (universe m26:
// `--match zzqqxx` and `--match Fermat` each returned the six), because its --json arm wrote its bytes
// past the selector and its bare arm was markdown the selector cannot read. The two other markdown
// views did the same on their bare form, and `inquest telemetry` — JSONL, no document to select
// entries from — did the same on both.
//
// For each row: a pattern that matches nothing returns no entry in any array and says it matched 0
// of the total; a quote taken from one entry returns only entries carrying it. A view the fixture
// leaves empty FAILS rather than passing on nothing — an empty view measures no selector.
func TestEverySelectorThatAViewAdvertisesSelects(t *testing.T) {
	runDir := censusRun(t)
	const nothing = "zzqqxx-matches-no-entry"
	for _, row := range selectorViews(t) {
		t.Run(row.group+" "+row.view, func(t *testing.T) {
			read := func(args ...string) (string, error) {
				return run(t, append([]string{row.group, row.view, "--run", runDir, "--seat-id", "red-chair"}, args...)...)
			}
			whole, err := read("--json")
			if err != nil {
				t.Fatalf("the unselected --json read: %v", err)
			}
			entries := arrayEntries(t, whole)
			if len(entries) == 0 {
				t.Fatalf("the fixture leaves %s %s with no entry, so its selector is not measured:\n%s", row.group, row.view, whole)
			}
			for _, flag := range []string{"--" + flags.Match, "--" + flags.Quote} {
				t.Run(flag+" matching nothing", func(t *testing.T) {
					out, err := read("--json", flag, nothing)
					if err != nil {
						t.Fatal(err)
					}
					if kept := arrayEntries(t, out); len(kept) != 0 {
						t.Errorf("%s %q kept %d of %d entries:\n%s", flag, nothing, len(kept), len(entries), out)
					}
					sel := selectionOf(t, out)
					if sel.Matched != 0 || sel.Of != len(entries) {
						t.Errorf("the selection says %d of %d, want 0 of %d", sel.Matched, sel.Of, len(entries))
					}
					// WHAT IS NOT A LIST OF ENTRIES IS LEFT AS IT WAS. A null is not an empty array: `work`'s
					// `engaged` is an object or null, and a selection rewrote its null as `[]`.
					for key, was := range nonArrays(t, whole) {
						if got := nonArrays(t, out)[key]; got != was {
							t.Errorf("selecting rewrote %q from %s to %s", key, was, got)
						}
					}
				})
			}
			lit := literalOf(t, entries)
			t.Run("--quote one entry's text", func(t *testing.T) {
				out, err := read("--json", "--"+flags.Quote, lit)
				if err != nil {
					t.Fatal(err)
				}
				kept := arrayEntries(t, out)
				if len(kept) == 0 {
					t.Fatalf("--quote %q, taken from an entry, kept nothing:\n%s", lit, out)
				}
				for _, e := range kept {
					if !carries(t, e, lit) {
						t.Errorf("--quote %q kept an entry that does not carry it: %s", lit, e)
					}
				}
				if sel := selectionOf(t, out); sel.Matched != len(kept) || sel.Of != len(entries) {
					t.Errorf("the selection says %d of %d, and the read kept %d of %d", sel.Matched, sel.Of, len(kept), len(entries))
				}
			})
			// THE BARE FORM. Where it is the JSON (a JSON-by-name view) it selects the same way; where
			// it is markdown it has no entries to select, and the pair is refused there by name rather
			// than answered with every entry.
			for _, flag := range []string{"--" + flags.Match, "--" + flags.Quote} {
				t.Run("the bare form "+flag, func(t *testing.T) {
					out, err := read(flag, nothing)
					switch {
					case row.view == "report":
						if err != nil || !strings.Contains(out, "no line of the report matches") {
							t.Errorf("the bare report did not report an empty selection: %v\n%s", err, out)
						}
					case row.markdown:
						if err == nil || !strings.Contains(err.Error(), "--json") {
							t.Errorf("the bare markdown form took --match and did not refuse it toward --json: %v\n%s", err, out)
						}
					default:
						if err != nil {
							t.Fatal(err)
						}
						if kept := arrayEntries(t, out); len(kept) != 0 {
							t.Errorf("the bare form kept %d entries for a pattern matching none:\n%s", len(kept), out)
						}
					}
				})
				// A --format THAT NAMES MARKDOWN is the same markdown, and refuses the pair the same way.
				if row.formats {
					t.Run("--format markdown "+flag, func(t *testing.T) {
						out, err := read("--format", "markdown", flag, nothing)
						if err == nil || !strings.Contains(err.Error(), "--format") {
							t.Errorf("--format markdown took %s and did not refuse it: %v\n%s", flag, err, out)
						}
					})
				}
			}
		})
	}
}

// selectorRow is one projection that carries the selector pair, as the seat's tree declares it.
type selectorRow struct {
	group, view string
	// markdown says the bare form is markdown: the page's JSON flag is a choice of form, not a no-op.
	markdown bool
	// formats says the view takes --format, whose markdown value is a markdown form too.
	formats bool
}

// selectorViews walks the chair's tree, which mounts every `show` projection and the `inquest` group.
func selectorViews(t *testing.T) []selectorRow {
	t.Helper()
	root := NewRootFor("red-chair")
	var rows []selectorRow
	for _, group := range []string{"show", "inquest"} {
		g, _, err := root.Find([]string{group})
		if err != nil || g == nil || g.Name() != group {
			t.Fatalf("the chair's tree has no %s group: %v", group, err)
		}
		for _, v := range g.Commands() {
			if v.Flags().Lookup(flags.Match) == nil && v.Flags().Lookup(flags.Quote) == nil {
				continue
			}
			if v.Flags().Lookup(flags.Match) == nil || v.Flags().Lookup(flags.Quote) == nil {
				t.Errorf("%s %s carries one of --match and --quote without the other", group, v.Name())
			}
			rows = append(rows, selectorRow{group: group, view: v.Name(), markdown: bareIsMarkdown(v), formats: v.Flags().Lookup(flags.Format) != nil})
		}
	}
	if len(rows) < 2 {
		t.Fatalf("the census found %d views with a selector — the walk is not reading the tree", len(rows))
	}
	return rows
}

// bareIsMarkdown reads the page the seat reads: a markdown view's output tree is labelled as the
// --json form.
func bareIsMarkdown(c *cobra.Command) bool {
	return c.Name() != "report" && strings.Contains(c.Long, "the bare call is the markdown form")
}

// arrayEntries is every element of every top-level array of a JSON document, as its JSON text.
func arrayEntries(t *testing.T, doc string) []string {
	t.Helper()
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(doc), &top); err != nil {
		t.Fatalf("not one JSON document (%v):\n%s", err, doc)
	}
	var out []string
	for _, raw := range top {
		var items []json.RawMessage
		if json.Unmarshal(raw, &items) != nil {
			continue
		}
		for _, it := range items {
			out = append(out, string(it))
		}
	}
	return out
}

// nonArrays is every top-level key of a JSON document whose value is not an array, as compact JSON.
func nonArrays(t *testing.T, doc string) map[string]string {
	t.Helper()
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(doc), &top); err != nil {
		t.Fatalf("not one JSON document (%v):\n%s", err, doc)
	}
	out := map[string]string{}
	for k, raw := range top {
		var b bytes.Buffer
		if err := json.Compact(&b, raw); err != nil {
			t.Fatal(err)
		}
		if k != "selection" && !strings.HasPrefix(b.String(), "[") {
			out[k] = b.String()
		}
	}
	return out
}

type selection struct {
	Matched int `json:"matched"`
	Of      int `json:"of"`
}

func selectionOf(t *testing.T, doc string) selection {
	t.Helper()
	var d struct {
		Selection *selection `json:"selection"`
	}
	if err := json.Unmarshal([]byte(doc), &d); err != nil || d.Selection == nil {
		t.Fatalf("a selected read carries no selection block (%v):\n%s", err, doc)
	}
	return *d.Selection
}

// plainText is a string a literal can be taken from without a JSON escape or an anchor in it.
var plainText = regexp.MustCompile(`^[A-Za-z0-9 ,.;:'?()!/-]{6,}$`)

// literalOf is the longest plain string value in the first entry that has one — text that entry
// certainly carries.
func literalOf(t *testing.T, entries []string) string {
	t.Helper()
	for _, e := range entries {
		if lit := plainStringIn(t, e); lit != "" {
			return lit
		}
	}
	t.Fatalf("no entry has plain text to quote: %v", entries)
	return ""
}

// plainStringIn is the longest string value of the entry that reads as plain text with its anchors out.
func plainStringIn(t *testing.T, entry string) string {
	t.Helper()
	best := ""
	for _, s := range stringsIn(t, entry) {
		if s = visible(s); plainText.MatchString(s) && len(s) > len(best) {
			best = s
		}
	}
	return best
}

// carries reports whether one of the entry's string values holds lit, read as a seat reads it.
func carries(t *testing.T, entry, lit string) bool {
	t.Helper()
	for _, s := range stringsIn(t, entry) {
		if strings.Contains(strings.ToLower(visible(s)), strings.ToLower(lit)) {
			return true
		}
	}
	return false
}

// visible is a string with its anchors out and its whitespace runs collapsed.
func visible(s string) string {
	return strings.Join(strings.Fields(claimcount.StripAnchors(s)), " ")
}

// stringsIn is every string value in one entry, at any depth.
func stringsIn(t *testing.T, entry string) []string {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(entry), &v); err != nil {
		t.Fatal(err)
	}
	var out []string
	var walk func(any)
	walk = func(x any) {
		switch x := x.(type) {
		case string:
			out = append(out, x)
		case []any:
			for _, e := range x {
				walk(e)
			}
		case map[string]any:
			for _, e := range x {
				walk(e)
			}
		}
	}
	walk(v)
	return out
}

// censusRun is a run with at least one entry in every projection a seat can select from: gaps on the
// board and the lens's list, findings, a motion, a position in the debate, an edit, a citation and
// inquiries.
func censusRun(t *testing.T) string {
	t.Helper()
	runDir, gap := answeredBoard(t)
	editAnswering(t, runDir, gap)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://census/1": []byte("<html>the cost is moderate and rising</html>")}})
	for _, args := range [][]string{
		{"finding", "--seat-id", lensSeat, "--key", "F1", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "The cost is moderate and rising.", "--reason", "the rising trend is asserted without a series"},
		{"position", "--seat-id", "red-chair", "--reason", "the chair's position on the open gap"},
		{"inquiry", "propose", "--seat-id", blueSeat, "--question", "does a second source give the same figure", "--hypothesis", "two sources agree within ten percent"},
		{"inquiry", "propose", "--seat-id", blueSeat, "--question", "is the trend seasonal", "--hypothesis", "the rise repeats every winter"},
		{"motion", "grade", "file", "--seat-id", blueSeat, "--id", gap, "--dimension", "likelihood", "--proposed", "low", "--reason", "the harm needs two failures"},
		{"cite", "--seat-id", blueSeat, "--quote", "The cost is moderate and rising.", "--url", "https://census/1", "--title", "Census source"},
	} {
		if out, err := run(t, append(args, "--run", runDir)...); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(args[:2], " "), err, out)
		}
	}
	return runDir
}
