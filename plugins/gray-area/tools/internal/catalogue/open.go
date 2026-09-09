package catalogue

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// DefaultDir is the store's home: XDG state, NOT ~/.claude. Writing our store into the vendor's
// directory is squatting in a namespace we do not own — the same principle that stops
// ResolveSession globbing ~/.claude/projects/ to guess which transcript is ours.
func DefaultDir() (string, error) {
	if x := os.Getenv("XDG_STATE_HOME"); x != "" {
		return filepath.Join(x, "special-circumstances", "catalogue"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("catalogue: no home directory: %w", err)
	}
	return filepath.Join(home, ".local", "state", "special-circumstances", "catalogue"), nil
}

// writeDSN carries the pragmas in the order that matters. busy_timeout is set BEFORE
// journal_mode(WAL), because converting to WAL takes an exclusive lock and a concurrent opener
// with no timeout yet fails instantly rather than waiting — the defect fixed in #814, on the same
// engine, which this store would otherwise reproduce.
const writeDSN = "?_txlock=immediate" +
	"&_pragma=busy_timeout(5000)" +
	"&_pragma=journal_mode(WAL)" +
	"&_pragma=synchronous(NORMAL)" +
	"&_pragma=foreign_keys(1)"

// Open opens the store for WRITING, creating it if absent.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("catalogue: %w", err)
	}
	db, err := sql.Open("sqlite", path+writeDSN)
	if err != nil {
		return nil, fmt.Errorf("catalogue: open %s: %w", path, err)
	}
	var got int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&got); err != nil {
		db.Close()
		return nil, fmt.Errorf("catalogue: reading user_version: %w", err)
	}
	// REFUSE A NEWER SHAPE rather than read it as this one. The store is derived, so the cost of
	// refusing is a reprojection; the cost of guessing is a store that answers questions wrongly
	// while looking healthy.
	if got > UserVersion {
		db.Close()
		return nil, fmt.Errorf("catalogue: %s was written by a newer binary (shape %d, this binary writes %d) — "+
			"delete it and let it reproject rather than reading a shape this binary does not know", path, got, UserVersion)
	}
	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("catalogue: schema: %w", err)
	}
	if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, UserVersion)); err != nil {
		db.Close()
		return nil, fmt.Errorf("catalogue: setting user_version: %w", err)
	}
	return db, nil
}

// readDSN is the query path's own. `_defensive=1` makes `PRAGMA writable_schema=1` a silent
// no-op — measured: it reads back 0 with the flag and 1 without, on the same connection — and
// unlike the ATTACH limit below it holds on every pooled connection, because it is applied at
// connection open.
const readDSN = "?mode=ro&_pragma=query_only(1)&_defensive=1&_pragma=busy_timeout(2000)"

// OpenRead opens the store READ-ONLY for the `sql` verb.
//
// It returns a *sql.DB whose pool is capped at one connection, and callers must obtain that
// connection through PinnedConn — never db.Query. See PinnedConn for why.
func OpenRead(path string) (*sql.DB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("catalogue: no store at %s — nothing has been captured yet, "+
			"which is not the same as a store that is empty: run `telepathy backfill` or let a session run", path)
	}
	db, err := sql.Open("sqlite", path+readDSN)
	if err != nil {
		return nil, fmt.Errorf("catalogue: open read-only %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

// PinnedConn hands back a connection carrying SQLITE_LIMIT_ATTACHED = 0.
//
// # Why a pinned connection and not the pool
//
// `sqlite.Limit` binds ONE connection, not the *sql.DB. Measured on v1.57.0: after
// Limit(c1, ATTACHED, 0) the first connection refuses ATTACH while a second from the same pool
// reads the limit back as 10 and attaches successfully, as does db.Exec straight through the
// pool. So a query issued through the pool can land on a connection the limit was never applied
// to, and the hole reopens under concurrency while every single-goroutine test still passes.
//
// mode=ro and query_only already refuse writes; this closes ATTACH, whose only power is reading
// another SQLite file. On this box that is not an escalation — every caller already has Read and
// Bash — but it becomes one the moment the surface is exposed to a caller less privileged than
// the local user, and it costs one call to shut.
func PinnedConn(ctx context.Context, db *sql.DB) (*sql.Conn, error) {
	c, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("catalogue: acquiring connection: %w", err)
	}
	if _, err := sqlite.Limit(c, sqlite3.SQLITE_LIMIT_ATTACHED, 0); err != nil {
		c.Close()
		return nil, fmt.Errorf("catalogue: closing ATTACH on the query connection: %w", err)
	}
	return c, nil
}
