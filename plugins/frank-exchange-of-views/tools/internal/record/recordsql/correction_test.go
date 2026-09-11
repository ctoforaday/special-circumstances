package recordsql

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// R5: A PARTIAL UNIQUE INDEX IS NO PARENT KEY. `events.key` is unique only WHERE key IS NOT NULL,
// and SQLite refuses a foreign key onto such an index — which is why Correction.corrects declares
// no `references` and the write checks the target inside its transaction instead. Asserted rather
// than remembered: if SQLite ever accepted it, the reference could move onto the schema, where a
// dangling correction would be refused by the store itself.
func TestAPartialUniqueKeyCannotParentAForeignKey(t *testing.T) {
	db := openScratch(t)
	if _, err := db.Exec(`CREATE TABLE "probe_parent"("id" INTEGER PRIMARY KEY, "key" TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX "probe_parent_key" ON "probe_parent"("key") WHERE "key" IS NOT NULL`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE "probe_child"("ref" TEXT REFERENCES "probe_parent"("key"))`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO "probe_parent"("key") VALUES ('k')`); err != nil {
		t.Fatal(err)
	}
	_, err := db.Exec(`INSERT INTO "probe_child"("ref") VALUES ('k')`)
	if err == nil || !strings.Contains(err.Error(), "foreign key mismatch") {
		t.Fatalf("a foreign key onto a PARTIAL unique index was %v — expected SQLite's foreign key mismatch. If it is accepted now, Correction.corrects can reference events.key and the in-transaction existence check becomes the store's", err)
	}
	// And the schema this binary builds agrees: the correction table references nothing.
	ddl, err := Schema()
	if err != nil {
		t.Fatal(err)
	}
	// Every body's event_id references events(id); what must NOT exist is a reference onto the key.
	if stmt := tableDDL(ddl, "correction"); stmt == "" || strings.Contains(stmt, `REFERENCES "events"("key")`) {
		t.Errorf("the correction table's DDL is %q — it must exist and must not reference events.key", stmt)
	}
}

// A CORRECTABLE BODY CARRIES NO UNIQUE COLUMN. A replacement is a second row of its target's body
// type, carrying the same label; a UNIQUE column on that table would refuse every correction of
// it, at the store, with a constraint message no seat could act on. The tiers are read off the
// schema, so a type made correctable later is checked the moment it is.
func TestNoCorrectableBodyHasAUniqueColumn(t *testing.T) {
	ddl, err := Schema()
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	vs := recordpb.EventType(0).Descriptor().Values()
	for i := 0; i < vs.Len(); i++ {
		typ := recordpb.EventType(vs.Get(i).Number())
		if recordpb.Word(typ) == "" || recordpb.Tier(typ) == recordpb.CorrectionTier_CORRECTION_TIER_NONE {
			continue
		}
		w := recordpb.Word(typ)
		stmt := tableDDL(ddl, w)
		if stmt == "" {
			t.Errorf("no table for the correctable type %s", w)
			continue
		}
		checked++
		if strings.Contains(strings.ToUpper(stmt), "UNIQUE") {
			t.Errorf("the correctable body %s declares a UNIQUE column — its replacement row would be refused:\n%s", w, stmt)
		}
		for _, idx := range indexDDL(ddl, w) {
			if strings.Contains(strings.ToUpper(idx), "UNIQUE") {
				t.Errorf("the correctable body %s has a unique index — its replacement row would be refused:\n%s", w, idx)
			}
		}
	}
	if checked < 17 {
		t.Fatalf("only %d correctable body tables found — the walk is not seeing the schema", checked)
	}
}

func openScratch(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "probe.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(path) })
	return db
}

// tableDDL is the CREATE TABLE statement for one table, out of the schema text.
func tableDDL(ddl, table string) string {
	for _, stmt := range strings.Split(ddl, ";") {
		s := strings.TrimSpace(stmt)
		if strings.HasPrefix(s, `CREATE TABLE "`+table+`"`) || strings.HasPrefix(s, `CREATE TABLE IF NOT EXISTS "`+table+`"`) {
			return s
		}
	}
	return ""
}

// indexDDL is every CREATE INDEX statement on one table.
func indexDDL(ddl, table string) []string {
	var out []string
	for _, stmt := range strings.Split(ddl, ";") {
		s := strings.TrimSpace(stmt)
		if strings.HasPrefix(s, "CREATE") && strings.Contains(s, "INDEX") && strings.Contains(s, ` ON "`+table+`"`) {
			out = append(out, s)
		}
	}
	return out
}
