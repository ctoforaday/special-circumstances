package record

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// NO PROJECTION LIST MAY MARSHAL AS `null`.
//
// Dropping `omitempty` stopped an unset field from vanishing. A nil SLICE re-creates the same
// ambiguity one token along: a reader cannot tell "no ancestors" from "not computed". The record
// already states the rule at the write path — a gap with no ancestors records `supersedes: []`,
// because an absent key would read as lineage UNKNOWN where the truth is lineage NONE — and the
// projection owes the same answer.
//
// `null` is CORRECT for a scalar or a struct pointer: an ungraded gap genuinely has no severity,
// and a run with no sitting genuinely has none. Only lists are policed here.
func TestNoProjectionListMarshalsAsNull(t *testing.T) {
	for name, v := range map[string]any{
		"BoardJSON": mustBoardJSONT(t, mustRun(t, newRun(t))),
		"WorkJSON":  workJSONOfGaps(nil, 0, nil, ""),
	} {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var generic any
		if err := json.Unmarshal(raw, &generic); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, path := range nullListsIn(reflect.ValueOf(v), generic, name) {
			t.Errorf("%s marshalled as null — emit [] so \"none\" is distinguishable "+
				"from \"not computed\"", path)
		}
	}
}

// A LIST FIELD MAY NOT CARRY `omitempty`, which is the same ambiguity one step earlier: the key
// vanishes and the reader cannot tell "none" from "not carried". Checked on the TYPE, so it holds
// for every value, not only the ones a test happens to build.
func TestNoProjectionListIsOmitEmpty(t *testing.T) {
	for name, v := range map[string]any{
		"BoardJSON": BoardJSON{},
		"WorkJSON":  WorkJSON{},
	} {
		for _, f := range omitEmptyLists(reflect.TypeOf(v), name) {
			t.Errorf("%s has `omitempty` on a list — the key vanishes when empty, and \"none\" "+
				"then reads as \"not carried\"", f)
		}
	}
}

func omitEmptyLists(t reflect.Type, path string) []string {
	var bad []string
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("json")
		key := strings.Split(tag, ",")[0]
		if key == "" || key == "-" {
			key = f.Name
		}
		p := path + "." + key
		if f.Type.Kind() == reflect.Slice && strings.Contains(tag, "omitempty") {
			bad = append(bad, p)
			continue
		}
		ft := f.Type
		for ft.Kind() == reflect.Slice || ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && ft != t {
			bad = append(bad, omitEmptyLists(ft, p)...)
		}
	}
	return bad
}

// nullListsIn walks the marshalled shape BESIDE the Go value, so the set of list fields comes from
// the structs themselves rather than a list kept by hand. The hand-kept version missed
// `edited_since`, which is the field a lens reads to learn whether blue moved its gap: omitted, and
// then null, it could not tell "no edits" from "edits not carried" and spent a call finding out.
// A guard whose own allowlist is maintained by hand has reproduced the defect it guards against.
func nullListsIn(v reflect.Value, j any, path string) []string {
	var bad []string
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		obj, _ := j.(map[string]any)
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if f.PkgPath != "" {
				continue
			}
			key := strings.Split(f.Tag.Get("json"), ",")[0]
			if key == "" || key == "-" {
				key = f.Name
			}
			sub, ok := obj[key]
			if !ok {
				continue
			}
			if v.Field(i).Kind() == reflect.Slice && sub == nil {
				bad = append(bad, path+"."+key)
				continue
			}
			bad = append(bad, nullListsIn(v.Field(i), sub, path+"."+key)...)
		}
	case reflect.Slice:
		arr, _ := j.([]any)
		for i := 0; i < v.Len() && i < len(arr); i++ {
			bad = append(bad, nullListsIn(v.Index(i), arr[i], fmt.Sprintf("%s[%d]", path, i))...)
		}
	}
	return bad
}
