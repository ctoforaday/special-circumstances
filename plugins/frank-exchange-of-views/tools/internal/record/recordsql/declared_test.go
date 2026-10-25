package recordsql

import (
	"reflect"
	"testing"
)

// The open compares a database against the compiled-in declared set, so a declared_gen.go that
// lags the DDL makes every open refuse a record the binary wrote, or admit one missing a column.
// `declaredgen -check` is the CI gate; this test makes plain `go test` notice too.
func TestTheCompiledInDeclaredSetIsWhatTheDDLDerives(t *testing.T) {
	derived, err := DeriveDeclared()
	if err != nil {
		t.Fatal(err)
	}
	const regen = "regenerate with `go -C plugins/frank-exchange-of-views/tools run ./devcmd/declaredgen`"
	if !reflect.DeepEqual(declared.Tables, derived.Tables) {
		t.Errorf("declared_gen.go's tables differ from the DDL's — %s\n compiled in: %v\n derived:     %v", regen, declared.Tables, derived.Tables)
	}
	if !reflect.DeepEqual(declared.Columns, derived.Columns) {
		for tbl, cols := range derived.Columns {
			if !reflect.DeepEqual(declared.Columns[tbl], cols) {
				t.Errorf("declared_gen.go's columns for %q differ from the DDL's — %s\n compiled in: %v\n derived:     %v", tbl, regen, declared.Columns[tbl], cols)
			}
		}
		for tbl := range declared.Columns {
			if _, ok := derived.Columns[tbl]; !ok {
				t.Errorf("declared_gen.go declares columns for %q, which the DDL does not create — %s", tbl, regen)
			}
		}
	}
	if !reflect.DeepEqual(declared.Views, derived.Views) {
		for v, text := range derived.Views {
			if declared.Views[v] != text {
				t.Errorf("declared_gen.go's view %q differs from the DDL's — %s", v, regen)
			}
		}
		for v := range declared.Views {
			if _, ok := derived.Views[v]; !ok {
				t.Errorf("declared_gen.go declares view %q, which the DDL does not create — %s", v, regen)
			}
		}
	}
}
