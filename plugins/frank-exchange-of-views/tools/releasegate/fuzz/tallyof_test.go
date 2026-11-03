package fuzz

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

// fill sets every field reachable from v to a value other than its zero, outcome's unexported
// fields included (reached through their address). A kind it cannot fill fails the test, so a new
// outcome field of a new kind is taught here rather than left zero.
func fill(t *testing.T, v reflect.Value, n *int, path string) {
	t.Helper()
	*n++
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanSet() {
				f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
			}
			fill(t, f, n, path+"."+v.Type().Field(i).Name)
		}
	case reflect.Int, reflect.Int64:
		v.SetInt(int64(*n))
	case reflect.String:
		v.SetString(fmt.Sprintf("s%d", *n))
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		fill(t, k, n, path+"[key]")
		e := reflect.New(v.Type().Elem()).Elem()
		fill(t, e, n, path+"[]")
		m.SetMapIndex(k, e)
		v.Set(m)
	case reflect.Slice:
		e := reflect.New(v.Type().Elem()).Elem()
		fill(t, e, n, path+"[]")
		v.Set(reflect.Append(reflect.MakeSlice(v.Type(), 0, 1), e))
	default:
		t.Fatalf("%s is a %s, which fill does not set; teach it, so tallyOf's read of it is checked", path, v.Kind())
	}
}

// TALLYOF CARRIES EVERY RUN FIELD. A field tallyOf leaves out is zero in every part's tally, so the
// fold reads it and the gate judges it — over a zero no run wrote. TestTheFoldReadsEveryRunFieldAndWritesEveryFoldField
// starts from a runTally and cannot see this; this starts from an outcome with every field set.
func TestTallyOfCarriesEveryRunField(t *testing.T) {
	var o outcome
	n := 0
	fill(t, reflect.ValueOf(&o).Elem(), &n, "outcome")
	// hearingsAfter is read by barrier name, so its keys are the three barriers.
	o.hearingsAfter = map[string]int{barrierChair: 1, barrierLens: 2, barrierBlue: 3}
	rv := reflect.ValueOf(tallyOf(o))
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).IsZero() {
			t.Errorf("runTally.%s is zero from an outcome with every field set — tallyOf does not carry it", rv.Type().Field(i).Name)
		}
	}
}
