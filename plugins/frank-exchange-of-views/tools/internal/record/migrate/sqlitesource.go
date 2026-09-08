package migrate

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // the module's own driver; opened directly, never through recordsql.Open

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// dbSiblings is the FILE SET a SQLite record is. The database alone is not the record: the
// plan's audit measured the archived record.db reading clean at 691 events while 122 more —
// most of the friction events this migration exists to translate — lived only in the 3.1 MB
// -wal beside it. A truncated source is the plausible-zero shape, so the set is copied
// together and SQLite replays the WAL on the copy.
var dbSiblings = []string{"record.db", "record.db-wal", "record.db-shm"}

// SQLiteSource reads a former-vocabulary record.db. It operates on a COPY: the originals are
// never opened, because opening a database can checkpoint its WAL and the archive is the raw
// record.
type SQLiteSource struct {
	db    *sql.DB
	files []SourceFile
	// unclassified names tables the decomposition rules cannot place. Empty ones are noise
	// (a body table of a word no event used reads identically); one holding rows is refused
	// at read, because rows nobody can attribute are events nobody replays.
	unclassified []string
}

// OpenSQLite copies the record's file set from recordsDir into scratchDir and opens the copy.
func OpenSQLite(recordsDir, scratchDir string) (*SQLiteSource, error) {
	if _, err := os.Stat(filepath.Join(recordsDir, dbSiblings[0])); err != nil {
		return nil, fmt.Errorf("migrate: no %s in %s: %w", dbSiblings[0], recordsDir, err)
	}
	if err := os.MkdirAll(scratchDir, 0o755); err != nil {
		return nil, err
	}
	s := &SQLiteSource{}
	for _, name := range dbSiblings {
		from := filepath.Join(recordsDir, name)
		st, err := os.Stat(from)
		if os.IsNotExist(err) {
			continue // a checkpointed record legitimately has no -wal/-shm
		} else if err != nil {
			return nil, err
		}
		sum, err := copyHashing(from, filepath.Join(scratchDir, name))
		if err != nil {
			return nil, err
		}
		s.files = append(s.files, SourceFile{Name: name, SHA256: sum, Bytes: st.Size()})
	}
	// A plain open, read-write ON THE COPY: SQLite needs write access to replay the WAL, and
	// replaying it is the point.
	db, err := sql.Open("sqlite", filepath.Join(scratchDir, dbSiblings[0]))
	if err != nil {
		return nil, err
	}
	s.db = db
	return s, nil
}

func copyHashing(from, to string) (string, error) {
	in, err := os.Open(from)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(to)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, h), in); err != nil {
		out.Close()
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *SQLiteSource) Files() []SourceFile { return s.files }
func (s *SQLiteSource) Close() error        { return s.db.Close() }

// Events reads the whole old record, mirroring the schema generator's decomposition in
// reverse: body table per word, `<word>_<field>` list tables (event_id/ord/value),
// `<word>_<arm>` oneof arm tables (event_id primary key).
func (s *SQLiteSource) Events() ([]OldEvent, error) {
	words, err := s.vocabulary()
	if err != nil {
		return nil, err
	}
	evs, err := s.envelopes()
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]*OldEvent, len(evs))
	for i := range evs {
		byID[evs[i].ID] = &evs[i]
	}
	tables, err := s.tables()
	if err != nil {
		return nil, err
	}
	// A body table is named for its MESSAGE, not its word — the `verdict` word's body lives
	// in `round_verdict` (message RoundVerdict). Where the word survives, the current schema
	// is the authority on that spelling; a retired word's table is its own name, and a
	// renamed message under a kept word would surface below as a row-bearing table nobody
	// can place — loud, not folded.
	byTable := map[string]string{}
	for w := range words {
		byTable[tableForWord(w)] = w
	}
	for _, t := range tables {
		if err := s.classify(t, byTable, byID); err != nil {
			return nil, err
		}
	}
	return evs, nil
}

func tableForWord(word string) string {
	if fd := bodyFieldFor(word); fd != nil {
		return recordsql.TableName(fd.Message())
	}
	return word
}

// vocabulary is the OLD record's own event-type set: the enum_event_type table where the
// record carries one, else the distinct types actually written. The fallback is sound for
// classification because a side table can only be misread as a body table (or vice versa)
// when its word holds rows, and rows imply the word appears in events.
func (s *SQLiteSource) vocabulary() (map[string]bool, error) {
	words := map[string]bool{}
	rows, err := s.db.Query(`SELECT "value" FROM "enum_event_type"`)
	if err != nil {
		rows, err = s.db.Query(`SELECT DISTINCT "type" FROM "events"`)
		if err != nil {
			return nil, fmt.Errorf("migrate: reading the old vocabulary: %w", err)
		}
	}
	defer rows.Close()
	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			return nil, err
		}
		words[w] = true
	}
	return words, rows.Err()
}

func (s *SQLiteSource) envelopes() ([]OldEvent, error) {
	rows, err := s.db.Query(`SELECT "id", "seat_id", "round", "ts", "type", "key" FROM "events" ORDER BY "id"`)
	if err != nil {
		return nil, fmt.Errorf("migrate: reading the old envelope: %w", err)
	}
	defer rows.Close()
	var out []OldEvent
	for rows.Next() {
		var ev OldEvent
		var key *string
		if err := rows.Scan(&ev.ID, &ev.SeatID, &ev.Round, &ev.TS, &ev.Word, &key); err != nil {
			return nil, err
		}
		if key != nil {
			ev.Key = *key
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

func (s *SQLiteSource) tables() ([]string, error) {
	rows, err := s.db.Query(`SELECT "name" FROM "sqlite_master" WHERE "type" = 'table' ORDER BY "name"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// classify routes one table's rows to the events they belong to. The longest word that
// prefixes the name wins, so `motion_rule_docket` is an arm of the WORD `motion_rule`, not
// of `motion` — the same both-directions care the legacy shard regex documents.
func (s *SQLiteSource) classify(t string, byTable map[string]string, byID map[int64]*OldEvent) error {
	switch {
	case t == "events" || t == "seat_turn" || strings.HasPrefix(t, "sqlite_") || strings.HasPrefix(t, "enum_"):
		return nil
	case byTable[t] != "":
		return s.readBody(t, byID, byTable[t])
	}
	owner, ownerTable, field := "", "", ""
	for bt, w := range byTable {
		if strings.HasPrefix(t, bt+"_") && len(bt) > len(ownerTable) {
			owner, ownerTable, field = w, bt, t[len(bt)+1:]
		}
	}
	if owner == "" {
		// Not the envelope, not a vocabulary, not any word's table. Empty is noise; rows are
		// events nobody can replay, and that must not fold into a clean run.
		var n int
		if err := s.db.QueryRow(fmt.Sprintf(`SELECT count(*) FROM %q`, t)).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("migrate: table %q holds %d row(s) and belongs to no word this record declares — "+
				"rows nobody can attribute are events nobody replays, so this is a refusal rather than a skip", t, n)
		}
		s.unclassified = append(s.unclassified, t)
		return nil
	}
	cols, err := s.columns(t)
	if err != nil {
		return err
	}
	if isListShape(cols) {
		return s.readList(t, owner, field, byID)
	}
	return s.readArm(t, owner, field, byID)
}

func isListShape(cols []string) bool {
	if len(cols) != 3 {
		return false
	}
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	return set["event_id"] && set["ord"] && set["value"]
}

func (s *SQLiteSource) columns(t string) ([]string, error) {
	rows, err := s.db.Query(`SELECT "name" FROM pragma_table_info(?)`, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *SQLiteSource) readBody(t string, byID map[int64]*OldEvent, word string) error {
	return s.eachRow(t, func(id int64, fields map[string]any) error {
		ev, ok := byID[id]
		if !ok {
			return fmt.Errorf("migrate: %q row for event %d, which the envelope does not hold", t, id)
		}
		if ev.Word != word {
			return fmt.Errorf("migrate: %q row for event %d, whose word is %q not %q", t, id, ev.Word, word)
		}
		ev.Fields = fields
		return nil
	})
}

func (s *SQLiteSource) readArm(t, owner, arm string, byID map[int64]*OldEvent) error {
	return s.eachRow(t, func(id int64, fields map[string]any) error {
		ev, ok := byID[id]
		if !ok {
			return fmt.Errorf("migrate: %q row for event %d, which the envelope does not hold", t, id)
		}
		if ev.Word != owner {
			return fmt.Errorf("migrate: %q row for event %d, whose word is %q not %q", t, id, ev.Word, owner)
		}
		if ev.Arms == nil {
			ev.Arms = map[string]map[string]any{}
		}
		ev.Arms[arm] = fields
		return nil
	})
}

func (s *SQLiteSource) readList(t, owner, field string, byID map[int64]*OldEvent) error {
	rows, err := s.db.Query(fmt.Sprintf(`SELECT "event_id", "ord", "value" FROM %q ORDER BY "event_id", "ord"`, t))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, ord int64
		var v string
		if err := rows.Scan(&id, &ord, &v); err != nil {
			return err
		}
		ev, ok := byID[id]
		if !ok {
			return fmt.Errorf("migrate: %q row for event %d, which the envelope does not hold", t, id)
		}
		if ev.Word != owner {
			return fmt.Errorf("migrate: %q row for event %d, whose word is %q not %q", t, id, ev.Word, owner)
		}
		if ev.Lists == nil {
			ev.Lists = map[string][]string{}
		}
		ev.Lists[field] = append(ev.Lists[field], v)
	}
	return rows.Err()
}

// eachRow scans a whole event_id-keyed table, handing each row over as loose fields with
// event_id extracted and SQL NULLs omitted.
func (s *SQLiteSource) eachRow(t string, fn func(id int64, fields map[string]any) error) error {
	rows, err := s.db.Query(fmt.Sprintf(`SELECT * FROM %q ORDER BY "event_id"`, t))
	if err != nil {
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		var id int64
		fields := map[string]any{}
		for i, c := range cols {
			v := vals[i]
			if v == nil {
				continue
			}
			if c == "event_id" {
				n, ok := v.(int64)
				if !ok {
					return fmt.Errorf("migrate: %q event_id is %T, not an integer", t, v)
				}
				id = n
				continue
			}
			fields[c] = v
		}
		if err := fn(id, fields); err != nil {
			return err
		}
	}
	return rows.Err()
}

// Unclassified lists the empty tables no rule could place, for the manifest.
func (s *SQLiteSource) Unclassified() []string {
	out := append([]string(nil), s.unclassified...)
	sort.Strings(out)
	return out
}
