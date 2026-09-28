package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
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
		// EACH VIEW IS READ WHERE IT LIVES, BY A SEAT THAT HOLDS IT. The bench's projections are
		// `inquest` verbs on the judge's surface; asked for under `show` as the chair they were
		// refused, the refusal counted as "emitted nothing", and debate, motions and telemetry
		// passed this test for their whole lives without one byte of theirs being compared.
		group, reader := seat.GroupOf(view), "red-chair"
		if group == "inquest" {
			reader = "judge"
		}
		trimmed, found := "", false
		for _, args := range [][]string{{view}, {view, "--json"}} {
			out, err := run(t, append([]string{group, "--run", runDir, "--seat-id", reader}, args...)...)
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

		help, herr := run(t, group, view, "--seat-id", reader, "--help")
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

		// EVERY ARRAY THE TREE DOCUMENTS ARRIVES AS AN ARRAY, at every depth. The tree says
		// `red_closings:[…]`; a nil slice marshals as null and an omitempty one vanishes, and a
		// seat that believed the help — `.red_closings | map(.gap_id)` — was told "Cannot iterate
		// over null" on universe-m13 and blamed its own jq. Top-level keys alone never saw it:
		// the null was inside `epochs[]`.
		for _, doc := range jsonDocuments(trimmed) {
			var v any
			if err := json.Unmarshal([]byte(doc), &v); err != nil {
				t.Errorf("%s %s: emitted a document that does not parse: %v", group, view, err)
				continue
			}
			for _, bad := range arrayViolations(view, reflect.TypeOf(seat.ShapeOf(view)), v) {
				t.Errorf("%s %s documents an array that it does not emit as one: %s", group, view, bad)
			}
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

// jsonDocuments is the projection's output as the documents a reader parses: the whole output when
// it is one document, each non-empty line when it is JSONL.
func jsonDocuments(trimmed string) []string {
	if json.Valid([]byte(trimmed)) {
		return []string{trimmed}
	}
	var out []string
	for _, line := range strings.Split(trimmed, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

// arrayViolations walks an emitted value beside the Go type its help was generated from and names
// every list-typed field that arrived null or absent. It reads the TYPE for what was
// promised and the VALUE for what arrived, so the check cannot agree with itself.
func arrayViolations(path string, t reflect.Type, v any) []string {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Implements(reflect.TypeOf((*json.Marshaler)(nil)).Elem()) {
		return nil
	}
	var out []string
	switch t.Kind() {
	case reflect.Struct:
		obj, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if f.Anonymous && name == "" {
				out = append(out, arrayViolations(path, f.Type, v)...) // inlined by encoding/json
				continue
			}
			if name == "" {
				name = f.Name
			}
			ft := f.Type
			for ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			val, present := obj[name]
			at := path + "." + name
			// A LIST, NOT A MAP. A map-typed field here is an optional RECORD (a gap's closure, null
			// while it is open), and reading a key off null is safe in every consumer; iterating a
			// null list is not.
			isList := ft.Kind() == reflect.Slice && ft.Elem().Kind() != reflect.Uint8
			switch {
			case isList && !present:
				out = append(out, fmt.Sprintf("%s is missing", at))
			case isList && val == nil:
				out = append(out, fmt.Sprintf("%s is null", at))
			case present && val != nil:
				out = append(out, arrayViolations(at, f.Type, val)...)
			}
		}
	case reflect.Slice, reflect.Array:
		items, ok := v.([]any)
		if !ok {
			return nil
		}
		for i, it := range items {
			out = append(out, arrayViolations(fmt.Sprintf("%s[%d]", path, i), t.Elem(), it)...)
		}
	case reflect.Map:
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		for k, it := range m {
			out = append(out, arrayViolations(path+"."+k, t.Elem(), it)...)
		}
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
