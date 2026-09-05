// Package jsonshape renders the FIELD TREE of a Go value as it will be marshalled to JSON —
// the key names a reader gets, in the nesting they arrive in, and nothing else.
//
// # Why this is generated and not written
//
// A seat that does not know a projection's key names discovers them at runtime: dump the
// projection to a scratch file, `python3 -c "print(list(d.keys()))"`, then a second script to
// pull the field it wanted. Measured across one debate run: 251 `json.load` re-parses and 25
// `list(d.keys())` key-dumps against three uses of `jq` (#684 F7). Each of those is a process
// spawn, a hook fire and a full model turn, and the whole class exists because the key names
// were nowhere the seat could read them.
//
// The obvious repair — write the shape into each projection's help prose — is the defect the
// `views` table's own comment records: the JSON-by-name fact lived in prose AND in a hand-kept
// switch, and nothing held them together. Prose describing a struct is a second copy of that
// struct with no writer who can refuse it, and it goes stale on the first field rename with
// every test still green. So the tree is DERIVED from the type that is actually marshalled, and
// pinned by a test against the keys the verb actually emits — see
// TestProjectionShapeMatchesEmittedKeys. Where a shape cannot be generated it must be absent
// rather than approximated: a field tree that is subtly wrong is worse than none, because a
// seat will believe it.
package jsonshape

import (
	"reflect"
	"strings"
)

// maxDepth bounds the walk. Three levels reach every key a seat addresses in practice
// (`{sitting:{open:[{...}]}}`) and stop before the tree stops being readable at a glance. A
// deeper node renders as `…`, which is honest about being elided rather than silently flat.
const maxDepth = 3

// Tree renders v's JSON field tree, e.g. `{open:[{id,severity}],counts:{open,closed}}`.
//
// v is used for its TYPE only; a zero value is the intended argument and no field is read. A
// type with no JSON structure to describe — a scalar, or a struct with no exported tagged
// fields — returns "", which callers MUST treat as "say nothing" rather than printing an empty
// pair of braces.
func Tree(v any) string {
	if v == nil {
		return ""
	}
	s := walk(reflect.TypeOf(v), 0)
	if s == "" || !strings.HasPrefix(s, "{") && !strings.HasPrefix(s, "[") {
		return "" // a scalar describes nothing a seat can navigate
	}
	return s
}

func walk(t reflect.Type, depth int) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		if depth >= maxDepth {
			return "{…}"
		}
		var fields []string
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			// AN EMBEDDED STRUCT IS INLINED BY encoding/json, so its fields arrive at THIS
			// level and rendering it as a nested object would name a key that never appears in
			// the output. Checked BEFORE the exported gate below, because promotion does not
			// require the EMBEDDED TYPE to be exported — only the fields it carries. A lowercase
			// embedded type whose fields are exported marshals them, and gating on the field
			// name first dropped every one of them.
			if f.Anonymous && f.Tag.Get("json") == "" {
				ft := f.Type
				for ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}
				if ft.Kind() == reflect.Struct {
					if inner := walk(ft, depth); strings.HasPrefix(inner, "{") {
						// TrimPrefix/TrimSuffix, NOT Trim: a cutset trim eats the closing brace
						// of a nested object too, turning `{a,b:{c}}` into `a,b:{c`.
						if flat := strings.TrimSuffix(strings.TrimPrefix(inner, "{"), "}"); flat != "" {
							fields = append(fields, flat)
						}
						continue
					}
				}
			}
			name, ok := jsonName(f)
			if !ok {
				continue
			}
			if sub := walk(f.Type, depth+1); sub != "" {
				fields = append(fields, name+":"+sub)
				continue
			}
			fields = append(fields, name)
		}
		if len(fields) == 0 {
			return ""
		}
		return "{" + strings.Join(fields, ",") + "}"
	case reflect.Slice, reflect.Array:
		// []byte marshals as a base64 STRING, not an array — rendering it as a list would
		// describe a shape no consumer will ever see.
		if t.Elem().Kind() == reflect.Uint8 {
			return ""
		}
		if sub := walk(t.Elem(), depth); sub != "" {
			return "[" + sub + "]"
		}
		// A LIST OF SCALARS NAMES ITS ELEMENT TYPE. Rendering it `[]` reads as an empty array
		// — a fact about one response, where this is a statement about every response.
		return "[" + scalarName(t.Elem()) + "]"
	case reflect.Map:
		if depth >= maxDepth {
			return "{…}"
		}
		// The KEYS are the caller's data, not the schema — only the value shape is knowable.
		if sub := walk(t.Elem(), depth+1); sub != "" {
			return "{<key>:" + sub + "}"
		}
		return "{<key>:" + scalarName(t.Elem()) + "}"
	case reflect.Interface:
		// An `any` field's shape is decided at runtime. Saying nothing is the honest answer;
		// guessing one is how a generated tree starts lying.
		return ""
	default:
		return ""
	}
}

// jsonName is the key encoding/json will emit for f, and whether f is emitted at all.
func jsonName(f reflect.StructField) (string, bool) {
	if !f.IsExported() {
		return "", false
	}
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", false
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		name = f.Name
	}
	return name, true
}

// scalarName is the JSON TYPE a non-navigable Go type lands as on the wire. It is used only
// where a key name cannot carry the information on its own — inside a list or as a map value —
// and it names the JSON type rather than the Go one, because the JSON is what the reader sees.
func scalarName(t reflect.Type) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	default:
		// An interface, a channel, a func: decided at runtime or not marshalled at all. The
		// ellipsis says "not knowable from the type", which is the true answer.
		return "…"
	}
}
