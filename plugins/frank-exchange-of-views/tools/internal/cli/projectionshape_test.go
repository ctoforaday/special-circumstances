package cli

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

// outputLine finds the generated field tree on a projection's help page.
//
// A PATTERN READ OFF A RENDERED PAGE, which is the shape this repo distrusts — so the miss is
// made LOUD rather than folded into a pass. The caller below FAILS a projection that emits JSON
// and carries no such line; a no-match can never read as agreement here.
var outputLine = regexp.MustCompile(`(?m)^OUTPUT \(JSONL?[^)]*\): (.+)$`)

// THE GENERATED TREE IS PINNED AGAINST WHAT THE VERB ACTUALLY EMITS.
//
// A shape rendered from a Go type and checked against that same Go type agrees with itself and
// discriminates nothing — the defect TestJSONByNameMarkMatchesWhatTheBareViewEmits records and
// the reason the mark beside this one is behaviour-pinned. So the tree is compared to the top
// level keys the projection puts on stdout: rename a field and the tree moves with it, but drop
// the field from the marshalled type without moving the other, and this fails.
//
// It also holds the TOTAL mapping in both directions. A projection that emits JSON and carries
// no OUTPUT line is undocumented — the runtime key-discovery this exists to remove (#684 F7)
// comes straight back for that view, silently. A projection that emits prose and carries one is
// describing a shape no reader receives.
func TestProjectionShapeMatchesEmittedKeys(t *testing.T) {
	runDir := seatRun(t)
	// A GAP, so the run has a round: `telemetry` is dense over rounds 1..current and emits
	// nothing at all on a fixture where nothing was minted — an unchecked view, in the one shape
	// (JSONL) a whole-document parse cannot reach.
	mintGap(t, runDir, "projection-shape", "read-surface")

	for _, view := range seat.ViewNames() {
		// BOTH FORMS, BECAUSE THE JSON IS NOT ALWAYS THE BARE ONE. `debate` is markdown bare and
		// JSON behind --json; driving only the bare call convicted it of documenting a shape it
		// does emit. Which form carries the JSON is discovered here rather than listed, so a view
		// that changes form is re-measured instead of re-agreeing with a stale list.
		trimmed, found := "", false
		for _, args := range [][]string{{view}, {view, "--json"}} {
			out, err := run(t, append([]string{"show", "--run", runDir, "--seat-id", "red-merge-r1"}, args...)...)
			if err != nil {
				continue // this seat cannot open it, or the form is refused — neither is evidence
			}
			candidate := strings.TrimSpace(out)
			if candidate != "" && emitsJSONOrJSONL(candidate) {
				trimmed, found = candidate, true
				break
			}
			if candidate != "" && trimmed == "" {
				trimmed = candidate // the prose form, kept in case NO form is JSON
			}
		}
		if trimmed == "" {
			t.Logf("show %s emitted nothing on this fixture — shape UNCHECKED here, not confirmed", view)
			continue
		}

		help, herr := run(t, "show", view, "--seat-id", "red-merge-r1", "--help")
		if herr != nil {
			t.Errorf("show %s --help: %v", view, herr)
			continue
		}
		documented := outputLine.FindStringSubmatch(help)
		emitsJSON := found

		if !emitsJSON {
			if documented != nil {
				t.Errorf("show %s emits prose and its help carries an OUTPUT field tree — it "+
					"describes a shape no reader of this projection receives:\n  %s", view, documented[1])
			}
			continue
		}
		if documented == nil {
			t.Errorf("show %s emits JSON and its help carries NO OUTPUT field tree. A seat that "+
				"cannot read the key names discovers them at runtime, which is the spawn class "+
				"#684 F7 measured — add the projection's marshalled type to the views table's "+
				"`shape` field", view)
			continue
		}

		emitted := topLevelKeys(t, view, trimmed)
		if len(emitted) == 0 {
			t.Logf("show %s emitted no top-level keys on this fixture — UNCHECKED", view)
			continue
		}
		tree := documented[1]
		for _, key := range emitted {
			if !treeNamesTopLevel(tree, key) {
				t.Errorf("show %s emits top-level key %q that its documented tree does not name:\n  %s",
					view, key, tree)
			}
		}
		for _, key := range treeTopLevelKeys(tree) {
			if !contains(emitted, key) {
				t.Errorf("show %s documents top-level key %q that it does not emit — a seat "+
					"reading the help would reach for a field that is not there:\n  %s", view, key, tree)
			}
		}
	}
}

// topLevelKeys reads the keys of the projection's outermost object. For JSONL it reads the
// FIRST LINE, because the tree documents one line rather than the stream.
func topLevelKeys(t *testing.T, view, trimmed string) []string {
	t.Helper()
	doc := trimmed
	var whole map[string]json.RawMessage
	if json.Unmarshal([]byte(doc), &whole) != nil {
		for _, line := range strings.Split(trimmed, "\n") {
			if strings.TrimSpace(line) != "" {
				doc = line
				break
			}
		}
		if json.Unmarshal([]byte(doc), &whole) != nil {
			t.Logf("show %s: neither the document nor its first line is an object — UNCHECKED", view)
			return nil
		}
	}
	out := make([]string, 0, len(whole))
	for k := range whole {
		out = append(out, k)
	}
	return out
}

// treeTopLevelKeys splits `{a,b:{c,d},e:[{f}]}` into a, b, e — the keys at depth one only,
// which is the level the emitted document can be compared against.
func treeTopLevelKeys(tree string) []string {
	body := strings.TrimSuffix(strings.TrimPrefix(tree, "{"), "}")
	var out []string
	depth, start := 0, 0
	for i, r := range body {
		switch r {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, keyOf(body[start:i]))
				start = i + 1
			}
		}
	}
	if start < len(body) {
		out = append(out, keyOf(body[start:]))
	}
	return out
}

func keyOf(field string) string {
	name, _, _ := strings.Cut(strings.TrimSpace(field), ":")
	return name
}

func treeNamesTopLevel(tree, key string) bool { return contains(treeTopLevelKeys(tree), key) }

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
