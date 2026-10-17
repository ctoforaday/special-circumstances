package record

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// PlanField is one field of the relayed plan as record.Plan declares it.
//
// THE DECLARATION IS THE STRUCT, AND THIS IS ITS ONE WALK. `dispatch next` prints the plan
// (encoding/json, off the same tags), the record stores what it printed and computes its key, and
// the engine script's envelope schema tells the chair's harness what a relay must look like. The
// key and the schema read this walk, so a field added to the struct reaches both at once.
//
// WHAT A GO TYPE CANNOT SAY IS A TAG ON THE FIELD: `desc` is the sentence the schema hands the
// relaying seat, `bench` the sentence for the same field of the bench's party, `min` an integer's
// floor and `pattern` a string's, `emptyonlywith` names the boolean of the same struct that is
// true whenever a list is empty, and `key:"-"` leaves a field out of the plan's key (keyed). A tag
// absent leaves the schema looser, never wrong.
type PlanField struct {
	Name string // the Go field
	JSON string // the key the verb prints and a relay carries
	// Type is the JSON type: integer, boolean, string or array.
	Type string
	// Defect is what the declaration states that no schema can carry — a Go kind with no JSON
	// type here, a named string type with no declared set, a floor that is no integer. Empty on a
	// sound field. The schema's generator refuses a plan holding one, and
	// TestEveryPlanFieldIsOneASchemaCanCarry fails on it.
	Defect string
	// Optional: the verb omits the key when the value is empty (`omitempty`).
	Optional bool
	Desc     string
	Bench    string
	// Min is an integer's floor; nil where the field states none.
	Min     *int
	Pattern string
	// EmptyOnlyWith is the JSON key of the boolean, in the same struct, that is true whenever this
	// list is empty; "" where the field states no such rule.
	EmptyOnlyWith string
	// Unkeyed: the field says when the plan was asked, not what it decides, and the plan's key
	// leaves it out (`key:"-"`).
	Unkeyed bool
	// Enum is the closed set of a string field whose Go type declares one.
	Enum []string
	// Elem is the JSON type of an array's scalar elements; empty for an array of structs.
	Elem string
	// Of is the element struct of an array of structs, and Items its fields.
	Of    reflect.Type
	Items []PlanField
}

// planEnums is the closed set of each named string type the plan carries.
var planEnums = map[reflect.Type]func() []string{
	reflect.TypeFor[BlockerKind](): func() []string {
		ks := BlockerKinds()
		out := make([]string, len(ks))
		for i, k := range ks {
			out[i] = string(k)
		}
		return out
	},
}

// PlanFields is record.Plan's relayed fields in declaration order, nested structs included. A
// field tagged `json:"-"` is never relayed and is not listed.
func PlanFields() []PlanField { return fieldsOf(reflect.TypeFor[Plan]()) }

func fieldsOf(t reflect.Type) []PlanField {
	var out []PlanField
	for i := range t.NumField() {
		f := t.Field(i)
		name, opts, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" || !f.IsExported() {
			continue
		}
		p := PlanField{Name: f.Name, JSON: name, Optional: strings.Contains(opts, "omitempty"),
			Desc: f.Tag.Get("desc"), Bench: f.Tag.Get("bench"), Pattern: f.Tag.Get("pattern"),
			EmptyOnlyWith: f.Tag.Get("emptyonlywith"), Unkeyed: f.Tag.Get("key") == "-"}
		if m, ok := f.Tag.Lookup("min"); ok {
			n, err := strconv.Atoi(m)
			if err != nil {
				p.Defect = "its floor " + strconv.Quote(m) + " is no integer"
			}
			p.Min = &n
		}
		p.Type = jsonTypeOf(f.Type)
		if p.Type == "" {
			p.Defect = "its Go type " + f.Type.String() + " has no JSON type here"
		}
		if set, ok := planEnums[f.Type]; ok {
			p.Enum = set()
		} else if f.Type.Kind() == reflect.String && f.Type != reflect.TypeFor[string]() {
			p.Defect = "its Go type " + f.Type.String() + " is a named set of words, and planEnums declares no set for it"
		}
		if f.Type.Kind() == reflect.Slice {
			if e := f.Type.Elem(); e.Kind() == reflect.Struct {
				p.Of, p.Items = e, fieldsOf(e)
			} else if p.Elem = jsonTypeOf(e); p.Elem == "" || p.Elem == "array" {
				p.Defect = "its element type " + e.String() + " has no JSON type here"
			}
		}
		out = append(out, p)
	}
	for i, p := range out {
		if p.EmptyOnlyWith == "" {
			continue
		}
		flag := slices.IndexFunc(out, func(o PlanField) bool { return o.JSON == p.EmptyOnlyWith })
		switch {
		case p.Type != "array":
			out[i].Defect = "it is no list, and only a list is empty"
		case flag < 0 || out[flag].Type != "boolean":
			out[i].Defect = "its emptyonlywith names " + strconv.Quote(p.EmptyOnlyWith) + ", which is no boolean of " + t.String()
		}
	}
	return out
}

// relayable refuses a plan that breaks a rule its declaration states (PlanField.EmptyOnlyWith): the
// engine's envelope schema is written from the same declaration, so such a plan is one the chair's
// harness refuses however faithfully the chair relays it.
func relayable(p Plan) error {
	v := reflect.ValueOf(p)
	for _, f := range PlanFields() {
		if f.EmptyOnlyWith == "" || v.FieldByName(f.Name).Len() > 0 {
			continue
		}
		for _, g := range PlanFields() {
			if g.JSON == f.EmptyOnlyWith && !v.FieldByName(g.Name).Bool() {
				return fmt.Errorf("the plan's `%s` is empty and its `%s` is false; the plan states `%s` empty only when `%s` is true, so this plan is one no relay of it can carry", f.JSON, g.JSON, f.JSON, g.JSON)
			}
		}
	}
	return nil
}

func jsonTypeOf(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "string"
	case reflect.Slice:
		return "array"
	}
	return ""
}
