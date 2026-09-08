// Package migrate is the quarantine: the one place old record shapes are spelled.
//
// Every production reader refuses a record whose vocabulary its schema does not declare, and
// that refusal is correct — for them. This package exists to read exactly what they refuse,
// translate it through an authored registry, and re-drive the CURRENT write path so the output
// is a record the current binary could have written. See plans/replay-migration.md.
//
// Nothing outside this package may learn an old shape. Deleting this package one day deletes
// the capability cleanly.
package migrate

// SourceFile names one file the source was read from, with the hash of its bytes AS FOUND.
// These are file hashes, labelled as such; the manifest's source hash is the LOGICAL one
// (hash over the event stream as read), because file bytes differ across WAL checkpoint
// states of the same record.
type SourceFile struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

// OldEvent is one event as the OLD record spelled it: the envelope, plus the body decomposed
// exactly the way the schema generator decomposes it — scalar columns, list tables, oneof arm
// tables. Values are held loosely (any) because the old record's own bytes are the only
// authority on its shape; the translation registry is where loose meets typed, and where a
// mismatch becomes an error someone reads.
type OldEvent struct {
	ID     int64
	SeatID string
	Round  int32
	TS     string
	Key    string
	Word   string
	// Fields holds the body table's columns for this event, minus event_id, with SQL NULLs
	// omitted — absence in the map is absence in the record.
	Fields map[string]any
	// Lists holds each repeated field's values in ord order, keyed by the field name
	// (the `<word>_<field>` table with columns event_id/ord/value).
	Lists map[string][]string
	// Arms holds each oneof message arm's columns, keyed by the arm name
	// (the `<word>_<arm>` table keyed by event_id).
	Arms map[string]map[string]any
}

// Source yields an old record's events in record order. One adapter per former format:
// the SQLite adapter here in leg 1, the legacy JSONL shard adapter in leg 2.
type Source interface {
	// Events reads the WHOLE record, in the order it was written. A partial read is not
	// offered: the WAL truncation this plan's audit measured (691 of 813 events) is exactly
	// the kind of miss a partial reader folds into a clean answer.
	Events() ([]OldEvent, error)
	// Files lists what was actually read, for the manifest.
	Files() []SourceFile
	Close() error
}
