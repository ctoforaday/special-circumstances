package recordsql

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// schemaEpochPath holds the epoch the schema was last stamped at and the hash of that schema.
const schemaEpochPath = "testdata/schema.epoch.json"

// schemaStamp is schemaEpochPath's content: one epoch, one schema.
type schemaStamp struct {
	EventSchema int    `json:"eventSchema"`
	SHA256      string `json:"sha256"`
}

// EVERY DDL CHANGE BUMPS THE EPOCH, because the open trusts the epoch and compares nothing else
// (requireThisEpoch). A table, column or view changed under an unchanged epoch would be served from
// every record created before the change — a stale view answering by its old rule, a missing column
// failing as SQLite's bare "no such column" — with nothing to say why. This test is where that
// change is refused, at build time, instead of at every open.
//
// The hash is of what SQLite stores for each table, index, trigger and view — sqlite_master's text,
// the definitions a record actually holds — so a comment between statements moves nothing, and a
// change to any definition moves it.
//
// UPDATE_GOLDENS=1 restamps the file, and only once the epoch has moved: a changed schema at the
// stamped epoch is refused under UPDATE_GOLDENS too, because restamping it is the bypass this test
// exists to close. Bump `eventSchema` in plugins/frank-exchange-of-views/requirements.json, run
// `go -C scripts run ./schemagen`, then restamp.
func TestEveryDDLChangeBumpsTheEpoch(t *testing.T) {
	sum, err := declaredSchemaSHA256()
	if err != nil {
		t.Fatal(err)
	}
	now := schemaStamp{EventSchema: recordpb.EventSchema, SHA256: sum}

	var stamped schemaStamp
	raw, err := os.ReadFile(schemaEpochPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err == nil {
		if err := json.Unmarshal(raw, &stamped); err != nil {
			t.Fatalf("%s: %v", schemaEpochPath, err)
		}
	}
	if stamped == now {
		return
	}
	if stamped.EventSchema == now.EventSchema {
		t.Fatalf("the record's schema changed and its event-schema epoch did not (still %d).\n\n"+
			"An open trusts the epoch, so every record created before this change would be served "+
			"with the old tables and views. Bump eventSchema in plugins/frank-exchange-of-views/requirements.json, "+
			"run `go -C scripts run ./schemagen`, then restamp with UPDATE_GOLDENS=1 go test ./internal/record/recordsql",
			now.EventSchema)
	}
	if os.Getenv("UPDATE_GOLDENS") != "1" {
		t.Fatalf("%s is stamped at epoch %d and this binary writes %d.\n\n"+
			"Restamp with UPDATE_GOLDENS=1 go test ./internal/record/recordsql",
			schemaEpochPath, stamped.EventSchema, now.EventSchema)
	}
	out, err := json.MarshalIndent(now, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.FromSlash(schemaEpochPath), append(out, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s restamped at epoch %d", schemaEpochPath, now.EventSchema)
}

// declaredSchemaSHA256 applies this binary's schema and views to an in-memory database and hashes
// every definition SQLite stored, in name order.
func declaredSchemaSHA256() (string, error) {
	schema, err := Schema()
	if err != nil {
		return "", err
	}
	mem, err := sql.Open(driverName(), ":memory:")
	if err != nil {
		return "", err
	}
	defer mem.Close()
	mem.SetMaxOpenConns(1) // one connection is one in-memory database
	if _, err := mem.Exec(schema + ViewsDDL); err != nil {
		return "", err
	}
	rows, err := mem.Query(`SELECT type, "name", "sql" FROM sqlite_master WHERE "sql" IS NOT NULL ORDER BY type, "name"`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var b strings.Builder
	for rows.Next() {
		var typ, name, text string
		if err := rows.Scan(&typ, &name, &text); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s %s\n%s\n\n", typ, name, text)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:]), nil
}
