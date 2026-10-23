package fuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// EVERY ENVELOPE IS HELD TO THE SCHEMA ITS DISPATCH CARRIED, AS THE WORKFLOW HARNESS HOLDS IT.
//
// The sweep's chair relays `dispatch next`'s JSON verbatim, and the harness a real chair answers
// to validates that relay against the envelope schema debate.js generates from record.Plan. The
// sweep applied no schema at all, so a plan the verb prints and its own schema refuses — a field,
// a key, a rule the generator states and the verb does not hold — reached the engine here and is
// refused on every relay of a live run. Validated with a JSON-schema implementation rather than
// the engine's own field checks, which stand behind the schema and see only what the engine reads.
//
// A REFUSAL FAILS THE RUN, naming the seat and the clause. In a live run the harness hands the
// refusal back to the seat, which cannot relay the verb's plan any other way.
type envelopeSchemas struct {
	mu       sync.Mutex
	compiled map[string]*jsonschema.Schema
}

// refusal is why schema refuses env, or "" when it accepts it. schema is the dispatch's `schema`
// option as goja exports it.
func (s *envelopeSchemas) refusal(schema any, env map[string]any) string {
	src, err := json.Marshal(schema)
	if err != nil {
		return "the dispatch's schema does not marshal: " + err.Error()
	}
	s.mu.Lock()
	sch, ok := s.compiled[string(src)]
	if !ok {
		doc, derr := jsonschema.UnmarshalJSON(bytes.NewReader(src))
		if derr != nil {
			s.mu.Unlock()
			return "the dispatch's schema is not JSON: " + derr.Error()
		}
		c := jsonschema.NewCompiler()
		if cerr := c.AddResource("envelope.json", doc); cerr != nil {
			s.mu.Unlock()
			return "the dispatch's schema is refused by the compiler: " + cerr.Error()
		}
		if sch, err = c.Compile("envelope.json"); err != nil {
			s.mu.Unlock()
			return "the dispatch's schema does not compile: " + err.Error()
		}
		if s.compiled == nil {
			s.compiled = map[string]*jsonschema.Schema{}
		}
		s.compiled[string(src)] = sch
	}
	s.mu.Unlock()
	body, err := json.Marshal(env)
	if err != nil {
		return "the envelope does not marshal: " + err.Error()
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(body))
	if err != nil {
		return "the envelope is not JSON: " + err.Error()
	}
	if err := sch.Validate(inst); err != nil {
		why := strings.Join(strings.Fields(err.Error()), " ")
		if _, after, ok := strings.Cut(why, "envelope.json#' "); ok {
			why = after
		}
		return why
	}
	return ""
}

// distinctSchemaRefusals is each refusal once, in the order first seen.
func (r *runner) distinctSchemaRefusals() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, why := range r.schemaRefusals {
		if !slices.Contains(out, why) {
			out = append(out, why)
		}
	}
	return out
}

// noteSchemaRefusal records an envelope the dispatch's schema refused.
func (r *runner) noteSchemaRefusal(seatID, why string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schemaRefusals = append(r.schemaRefusals, fmt.Sprintf("%s: %s", seatID, why))
}

// The validator refuses what the harness refuses and accepts what it accepts, on the schema shapes
// debate.js's envelopes use — a closed object, a required key, an anyOf, an enum — so a sweep that
// reports no refusal measured one. The shape is the chair's: the envelope IS the plan, closed, with
// `notes` beside its fields.
func TestTheEnvelopeSchemaCheckRefusesWhatTheHarnessRefuses(t *testing.T) {
	schema := map[string]any{
		"type": "object", "required": []any{"pass_permitted", "blockers"}, "additionalProperties": false,
		"properties": map[string]any{"pass_permitted": map[string]any{"type": "boolean"}, "blockers": map[string]any{"type": "array"},
			"verdict": map[string]any{"type": "string", "enum": []any{"PASS", "FAIL"}}, "notes": map[string]any{"type": "string"}},
		"anyOf": []any{
			map[string]any{"required": []any{"pass_permitted"}, "properties": map[string]any{"pass_permitted": map[string]any{"enum": []any{true}}}},
			map[string]any{"required": []any{"blockers"}, "properties": map[string]any{"blockers": map[string]any{"minItems": 1}}},
		},
	}
	var s envelopeSchemas
	for name, c := range map[string]struct {
		env    map[string]any
		accept bool
	}{
		"a permitted PASS":                {map[string]any{"pass_permitted": true, "blockers": []any{}, "verdict": "PASS"}, true},
		"a blocker named, with notes":     {map[string]any{"pass_permitted": false, "blockers": []any{"b"}, "notes": "n"}, true},
		"nothing holding no PASS":         {map[string]any{"pass_permitted": false, "blockers": []any{}}, false},
		"a key the plan does not hold":    {map[string]any{"pass_permitted": true, "blockers": []any{}, "epochs": 1}, false},
		"the envelope wrapped in a key":   {map[string]any{"plan": map[string]any{"pass_permitted": true, "blockers": []any{}}}, false},
		"a verdict the record never says": {map[string]any{"pass_permitted": true, "blockers": []any{}, "verdict": nil}, false},
		"no plan":                         {map[string]any{}, false},
	} {
		if why := s.refusal(schema, c.env); (why == "") != c.accept {
			t.Errorf("%s: accept=%v, refusal %q", name, c.accept, why)
		}
	}
}
