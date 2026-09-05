package jsonshape

import (
	"encoding/json"
	"strings"
	"testing"
)

type inner struct {
	A string `json:"a"`
	B int    `json:"b"`
}

type embedded struct {
	Carried string `json:"carried"`
}

type sample struct {
	Name     string            `json:"name"`
	Count    int               `json:"count"`
	Items    []inner           `json:"items"`
	Tags     []string          `json:"tags"`
	Nested   inner             `json:"nested"`
	Lookup   map[string]inner  `json:"lookup"`
	Untagged string            // no tag: the FIELD name is the key
	Skipped  string            `json:"-"`
	unseen   string            //nolint:unused // unexported: never marshalled
	Omitted  string            `json:"omitted,omitempty"`
	Blob     []byte            `json:"blob"`
	Anything any               `json:"anything"`
	Strings  map[string]string `json:"strings"`
	embedded
}

// THE TREE IS CHECKED AGAINST WHAT encoding/json ACTUALLY EMITS, not against a string someone
// believed. A generated shape that agrees only with itself is the failure mode this package
// exists to remove, so the assertion goes through a real marshal.
func TestTreeNamesEveryKeyEncodingJSONEmits(t *testing.T) {
	got := Tree(sample{})

	var emitted map[string]json.RawMessage
	b, err := json.Marshal(sample{})
	if err != nil {
		t.Fatalf("marshalling the sample: %v", err)
	}
	if err := json.Unmarshal(b, &emitted); err != nil {
		t.Fatalf("unmarshalling the sample: %v", err)
	}

	for key := range emitted {
		if !strings.Contains(got, key) {
			t.Errorf("encoding/json emits key %q and the tree does not name it:\n  %s", key, got)
		}
	}
	// And the other direction: a key in the tree that no marshal produces would send a seat
	// looking for a field that is not there.
	for _, key := range []string{"Skipped", "unseen"} {
		if strings.Contains(got, key) {
			t.Errorf("tree names %q, which is never marshalled:\n  %s", key, got)
		}
	}
}

func TestTreeShapes(t *testing.T) {
	got := Tree(sample{})
	for _, want := range []string{
		"items:[{a,b}]",        // a slice of structs
		"tags:[string]",        // A LIST OF SCALARS NAMES ITS ELEMENT TYPE: `[]` would read as an empty array — a fact about one response, where this is a claim about every response
		"nested:{a,b}",         // a nested struct
		"lookup:{<key>:{a,b}}", // a map: keys are data, only the value shape is knowable
		"Untagged",             // untagged exported field keys on the FIELD name
		"omitted",              // omitempty is an encoding option, not part of the key
		"carried",              // an embedded struct is INLINED, not nested under a key
	} {
		if !strings.Contains(got, want) {
			t.Errorf("tree missing %q:\n  %s", want, got)
		}
	}
	// `blob` is []byte — base64 STRING on the wire. Describing it as a list would name a shape
	// no consumer sees.
	if strings.Contains(got, "blob:[") {
		t.Errorf("[]byte rendered as a list; it marshals as a string:\n  %s", got)
	}
	// `anything` is an interface: its shape is decided at runtime and guessing one is how a
	// generated tree starts lying. The KEY is still named.
	if strings.Contains(got, "anything:") {
		t.Errorf("an interface field was given a shape:\n  %s", got)
	}
	if !strings.Contains(got, "anything") {
		t.Errorf("an interface field's KEY must still be named:\n  %s", got)
	}
}

// SAYING NOTHING IS A RESULT, and callers branch on it. A type with no navigable JSON structure
// must return "" rather than `{}` — an empty pair of braces in a help page reads as "this
// projection emits an empty object", which is a different and false claim.
func TestTreeSaysNothingRatherThanGuessing(t *testing.T) {
	for name, v := range map[string]any{
		"nil":            nil,
		"scalar":         "",
		"int":            0,
		"empty struct":   struct{}{},
		"untagged-only":  struct{ unexported string }{}, //nolint:unused
		"slice of bytes": []byte(nil),
	} {
		if got := Tree(v); got != "" {
			t.Errorf("%s: Tree = %q, want \"\" (a shape it cannot describe must be absent, not approximated)", name, got)
		}
	}
}

// DEPTH IS BOUNDED AND THE ELISION IS VISIBLE. A tree that silently flattened at depth would
// tell a seat a nested object has no fields.
func TestTreeElidesDeepNodesVisibly(t *testing.T) {
	type d4 struct {
		Leaf string `json:"leaf"`
	}
	type d3 struct {
		Down d4 `json:"down"`
	}
	type d2 struct {
		Down d3 `json:"down"`
	}
	type d1 struct {
		Down d2 `json:"down"`
	}
	got := Tree(d1{})
	if !strings.Contains(got, "…") {
		t.Errorf("a tree deeper than the bound must SAY it was elided:\n  %s", got)
	}
}
