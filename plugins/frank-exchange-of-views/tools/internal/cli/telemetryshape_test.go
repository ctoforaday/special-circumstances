package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// shapeNode is one node of a help page's OUTPUT tree: an object with its keys, a list with its
// element, a map (`{<key>:…}`), or a leaf.
type shapeNode struct {
	kind   string // "object", "list", "map", "leaf"
	fields map[string]*shapeNode
	elem   *shapeNode
}

// parseShape reads the tree jsonshape renders — `{a,b:{c},d:[{e,f}],g:{<key>:number}}` — and fails
// on anything left over, so a tree it cannot read is never a tree it agreed with.
func parseShape(t *testing.T, tree string) *shapeNode {
	t.Helper()
	n, rest, err := shapeValue(tree)
	if err != nil || rest != "" {
		t.Fatalf("the OUTPUT tree does not parse (%v, left %q):\n  %s", err, rest, tree)
	}
	return n
}

func shapeValue(s string) (*shapeNode, string, error) {
	switch {
	case strings.HasPrefix(s, "{<key>:"):
		elem, rest, err := shapeValue(strings.TrimPrefix(s, "{<key>:"))
		if err != nil || !strings.HasPrefix(rest, "}") {
			return nil, rest, fmt.Errorf("an unclosed map: %v", err)
		}
		return &shapeNode{kind: "map", elem: elem}, rest[1:], nil
	case strings.HasPrefix(s, "{"):
		n := &shapeNode{kind: "object", fields: map[string]*shapeNode{}}
		s = s[1:]
		for {
			end := strings.IndexAny(s, ",:}")
			if end <= 0 {
				return nil, s, fmt.Errorf("a key is missing")
			}
			key := s[:end]
			s = s[end:]
			child := &shapeNode{kind: "leaf"}
			if s[0] == ':' {
				var err error
				if child, s, err = shapeValue(s[1:]); err != nil {
					return nil, s, err
				}
			}
			n.fields[key] = child
			if s == "" {
				return nil, s, fmt.Errorf("an unclosed object")
			}
			if s[0] == '}' {
				return n, s[1:], nil
			}
			if s[0] != ',' {
				return nil, s, fmt.Errorf("a field does not end at , or }")
			}
			s = s[1:]
		}
	case strings.HasPrefix(s, "["):
		elem, rest, err := shapeValue(s[1:])
		if err != nil || !strings.HasPrefix(rest, "]") {
			return nil, rest, fmt.Errorf("an unclosed list: %v", err)
		}
		return &shapeNode{kind: "list", elem: elem}, rest[1:], nil
	default:
		end := strings.IndexAny(s, ",}]")
		if end < 0 {
			end = len(s)
		}
		return &shapeNode{kind: "leaf"}, s[end:], nil
	}
}

// matchShape walks an emitted value beside the tree and names every node where they part: a key
// one side has and the other lacks, a list that arrived as anything but a list (`null` included),
// an object that arrived as anything but an object.
func matchShape(at string, n *shapeNode, v any) []string {
	switch n.kind {
	case "object":
		obj, ok := v.(map[string]any)
		if !ok {
			return []string{fmt.Sprintf("%s: the tree says an object, the line holds %#v", at, v)}
		}
		var out []string
		for _, k := range sortedKeys(n.fields) {
			val, present := obj[k]
			if !present {
				out = append(out, fmt.Sprintf("%s.%s: the tree names it, the line does not carry it", at, k))
				continue
			}
			if c := n.fields[k]; c.kind != "leaf" && !(c.kind == "map" && val == nil) {
				out = append(out, matchShape(at+"."+k, c, val)...)
			}
		}
		for k := range obj {
			if _, named := n.fields[k]; !named {
				out = append(out, fmt.Sprintf("%s.%s: the line carries it, the tree does not name it", at, k))
			}
		}
		return out
	case "list":
		items, ok := v.([]any)
		if !ok {
			return []string{fmt.Sprintf("%s: the tree says a list, the line holds %#v", at, v)}
		}
		var out []string
		for i, it := range items {
			if n.elem.kind != "leaf" {
				out = append(out, matchShape(fmt.Sprintf("%s[%d]", at, i), n.elem, it)...)
			}
		}
		return out
	case "map":
		if _, ok := v.(map[string]any); !ok {
			return []string{fmt.Sprintf("%s: the tree says a map, the line holds %#v", at, v)}
		}
	}
	return nil
}

func sortedKeys(m map[string]*shapeNode) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// enumWord fails unless v is a JSON string spelling a value of e: never a number, never the
// schema's GRADE_*/GAP_CLASS_* constant.
func enumWord(t *testing.T, at string, e protoreflect.EnumDescriptor, v any) {
	t.Helper()
	s, ok := v.(string)
	if !ok {
		t.Errorf("%s = %#v, want a word of %s", at, v, e.Name())
		return
	}
	if _, ok := recordpb.BySpelling(e, s); !ok || strings.ToUpper(s) == s {
		t.Errorf("%s = %q, which is no word of %s", at, s, e.Name())
	}
}

// THE TELEMETRY HELP IS THE TELEMETRY LINE, at every depth, read as the bench reads it: the real
// verb's stdout beside the real verb's --help. Epoch 1 mints at high and medium under
// false_universal and the catch-all; epoch 2 is a chair sitting that minted nothing, where every
// list the tree promises is `[]` and both rates are `null`. Every value an enum supplies is a word.
func TestTelemetryHelpTreeMatchesAnEmittedLine(t *testing.T) {
	runDir := castRun(t)
	if _, err := mintClass(t, runDir, "k1", "false_universal", "high", ""); err != nil {
		t.Fatalf("mint false_universal: %v", err)
	}
	if _, err := mintClass(t, runDir, "k2", "unclassified", "medium", "nearest is false_universal, which needs a quantifier; this claim has none"); err != nil {
		t.Fatalf("mint unclassified: %v", err)
	}
	for _, args := range [][]string{{"register"}, {"verdict", "--as", "FAIL"}} {
		if _, err := run(t, append(args, "--run", runDir, "--seat-id", "red-chair")...); err != nil {
			t.Fatalf("chair %v: %v", args, err)
		}
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "docket"); err != nil {
		t.Fatalf("register judge: %v", err)
	}
	out, err := run(t, "inquest", "telemetry", "--run", runDir, "--seat-id", "judge")
	if err != nil {
		t.Fatalf("inquest telemetry: %v\n%s", err, out)
	}
	help, err := run(t, "inquest", "telemetry", "--seat-id", "judge", "--help")
	if err != nil {
		t.Fatalf("inquest telemetry --help: %v", err)
	}
	documented := outputLine.FindStringSubmatch(help)
	if documented == nil {
		t.Fatalf("inquest telemetry --help carries no OUTPUT tree:\n%s", help)
	}
	tree := parseShape(t, documented[1])

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("%d telemetry lines, want 2 (a minting epoch and an empty one):\n%s", len(lines), out)
	}
	decoded := make([]map[string]any, len(lines))
	for i, line := range lines {
		if err := json.Unmarshal([]byte(line), &decoded[i]); err != nil {
			t.Fatalf("epoch %d's line does not parse: %v\n%s", i+1, err, line)
		}
		for _, bad := range matchShape(fmt.Sprintf("epoch %d", i+1), tree, decoded[i]) {
			t.Errorf("the help tree and the emitted line part: %s\n  tree: %s\n  line: %s", bad, documented[1], line)
		}
	}

	grade, class := recordpb.Grade(0).Descriptor(), recordpb.GapClass(0).Descriptor()
	for i, line := range decoded {
		if v := line["max_severity"]; v != nil {
			enumWord(t, fmt.Sprintf("epoch %d max_severity", i+1), grade, v)
		}
		nm, _ := line["new_mint"].(map[string]any)
		sev, _ := nm["by_severity"].([]any)
		for j, e := range sev {
			enumWord(t, fmt.Sprintf("epoch %d by_severity[%d].grade", i+1, j), grade, e.(map[string]any)["grade"])
		}
		cls, _ := nm["by_class"].([]any)
		for j, e := range cls {
			enumWord(t, fmt.Sprintf("epoch %d by_class[%d].class", i+1, j), class, e.(map[string]any)["class"])
		}
	}

	one, _ := decoded[0]["new_mint"].(map[string]any)
	if got := one["unclassified_rate"]; got != 0.5 {
		t.Errorf("epoch 1 unclassified_rate = %v, want 0.5 (one of its two mints is the catch-all)", got)
	}
	var words []string
	for _, e := range one["by_class"].([]any) {
		words = append(words, e.(map[string]any)["class"].(string))
	}
	if want := []string{"false_universal", "unclassified"}; !slices.Equal(words, want) {
		t.Errorf("epoch 1 by_class names %v, want %v in schema order", words, want)
	}

	two, _ := decoded[1]["new_mint"].(map[string]any)
	for _, k := range []string{"by_severity", "by_class"} {
		if l, ok := two[k].([]any); !ok || len(l) != 0 {
			t.Errorf("epoch 2 %s = %#v, want [] — an epoch that minted nothing measured zero of each", k, two[k])
		}
	}
	for _, k := range []string{"class_repeat_rate", "unclassified_rate"} {
		if v, present := two[k]; !present || v != nil {
			t.Errorf("epoch 2 %s = %v (present %v), want null — nothing was minted", k, v, present)
		}
	}
}
