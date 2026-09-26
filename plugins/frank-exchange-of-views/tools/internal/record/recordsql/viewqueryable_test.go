package recordsql

import (
	"database/sql"
	"testing"
)

// EVERY VIEW IS QUERIED, NOT MERELY CREATED.
//
// SQLite does not validate a view body at CREATE. A view left reading a table or a column that has
// gone therefore APPLIES CLEANLY and then returns nothing — which reads as an empty run rather than
// as a broken view, and that is the failure this asserts against. views.go already carried the
// warning in prose, from a change where a view left reading a dropped column would have reported
// every disposed gap as undisposed; nothing enforced it.
//
// `LIMIT 0` is deliberate: it resolves every name the body mentions and materialises no rows, so the
// gate stays a schema check and not a fixture.
func TestEveryViewQueriesNotMerelyCreates(t *testing.T) {
	db, err := sql.Open(driverName(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	sch, _ := Schema()
	if _, err := db.Exec(sch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ViewsDDL); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='view' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	rows.Close()
	for _, n := range names {
		if _, err := db.Exec(`SELECT * FROM "` + n + `" LIMIT 0`); err != nil {
			t.Errorf("view %q does not QUERY: %v", n, err)
		}
	}
	t.Logf("%d views, all queryable", len(names))
}
