package recordsql

import (
	"database/sql"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
	"testing"
)

func open(t *testing.T) *sql.DB {
	t.Helper()
	s, err := Schema()
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+tmpRun(t)+"/r.db?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(s); err != nil {
		t.Fatalf("the derived schema is not valid SQLite: %v", err)
	}
	return db
}

// EVERY BODY MESSAGE GETS A TABLE AND EVERY FIELD A COLUMN.
//
// The derivation walks the descriptors, so this cannot drift the way a committed .sql would — but
// it CAN silently skip: a kind with no SQL mapping returns an error at generation, and an arm the
// walk never reaches produces a schema that applies cleanly and holds nothing. That is the shape
// this repository keeps finding, so the coverage is asserted rather than assumed.
func TestEveryBodyAndFieldReachesTheSchema(t *testing.T) {
	db := open(t)
	bodies, err := Bodies()
	if err != nil {
		t.Fatal(err)
	}
	if len(bodies) == 0 {
		t.Fatal("no body messages — an empty schema applies cleanly and holds nothing")
	}
	for _, md := range bodies {
		table := TableName(md)
		cols := map[string]bool{}
		rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				t.Fatal(err)
			}
			cols[n] = true
		}
		rows.Close()
		if len(cols) == 0 {
			t.Errorf("%s has no table — its body can be recorded and never read", table)
			continue
		}
		for i := 0; i < md.Fields().Len(); i++ {
			fd := md.Fields().Get(i)
			name := string(fd.Name())
			switch {
			case fd.IsList():
				// repeated fields live in their own table, so they stay joinable
				var n int
				_ = db.QueryRow(`SELECT count(*) FROM pragma_table_info(?)`, table+"_"+name).Scan(&n)
				if n == 0 {
					t.Errorf("%s.%s is repeated and has no child table", table, name)
				}
			case fd.Message() != nil:
				var n int
				_ = db.QueryRow(`SELECT count(*) FROM pragma_table_info(?)`, table+"_"+name).Scan(&n)
				if n == 0 {
					t.Errorf("%s.%s is a message arm and has no table", table, name)
				}
			default:
				if !cols[name] {
					t.Errorf("%s.%s has no column — the field can be set on a body and lost on the way to the record", table, name)
				}
			}
		}
	}
}

// THE CONSTRAINTS ARE THE POINT, so each is driven rather than read.
//
// A schema that parses and refuses nothing is the plausible zero one layer down: every write
// succeeds, every reader sees a clean board, and the guarantees exist only in the DDL's prose.
func TestTheSchemaRefusesWhatTheRecordCannotHold(t *testing.T) {
	db := open(t)
	seed := func(t *testing.T) int64 {
		t.Helper()
		res, err := db.Exec(`INSERT INTO events (seat_id, round, key, ts, type) VALUES ('red-merge-r1', 1, 'red-merge-r1:mint:#' || ?, '2026-01-01T00:00:00Z', 'mint')`, seq())
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		return id
	}

	t.Run("an unknown enum value is refused", func(t *testing.T) {
		id := seed(t)
		_, err := db.Exec(`INSERT INTO mint (event_id, gap_id, class, problem, acceptance_check, check_kind, likelihood, impact)
			VALUES (?, 'R9-1', 'c', 'p', 'a', 'guesswork', 'medium', 'medium')`, id)
		if err == nil {
			t.Fatal("`guesswork` was accepted as a check kind — an unrecognised value lands in no bucket and the gap reads as checked by nothing")
		}
	})

	t.Run("a required field omitted is refused", func(t *testing.T) {
		id := seed(t)
		// gap_id is supplied: it is required too now, and a fixture omitting BOTH tests whichever
		// constraint SQLite reports first, which is not the one this case is about.
		_, err := db.Exec(`INSERT INTO mint (event_id, gap_id, class, acceptance_check, check_kind, likelihood, impact)
			VALUES (?, 'R9-2', 'c', 'a', 'document', 'medium', 'medium')`, id)
		if err == nil || !strings.Contains(err.Error(), "problem") {
			t.Fatalf("a mint with no problem was accepted (%v) — required.go says the verb may not omit it, and the two must not disagree", err)
		}
	})

	t.Run("a ruling cannot answer two subjects at once", func(t *testing.T) {
		res, _ := db.Exec(`INSERT INTO events (seat_id, round, key, ts, type) VALUES ('judge-r1', 1, 'judge-r1:motion_rule:#' || ?, '2026-01-01T00:00:01Z', 'motion_rule')`, seq())
		id, _ := res.LastInsertId()
		_, err := db.Exec(`INSERT INTO motion_rule (event_id, motion_id, subject, grade, petition) VALUES (?, 'M1', 'grade', 'accepted', 'granted')`, id)
		if err == nil {
			t.Fatal("a motion rule carried BOTH a grade verdict and a petition verdict — the oneof exists so that a grade motion cannot be answered with a petition's word")
		}
	})

	t.Run("the record is append-only", func(t *testing.T) {
		id := seed(t)
		if _, err := db.Exec(`UPDATE events SET seat_id = 'someone-else' WHERE id = ?`, id); err == nil {
			t.Error("an event was edited after it was written — the record is evidence in an adversarial process, and evidence that can be rewritten is not evidence")
		}
		if _, err := db.Exec(`DELETE FROM events WHERE id = ?`, id); err == nil {
			t.Error("an event was deleted after it was written")
		}
	})

	t.Run("a body cannot exist without its event", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO mint (event_id, class, problem, acceptance_check, check_kind, likelihood, impact)
			VALUES (99999, 'c', 'p', 'a', 'document', 'medium', 'medium')`)
		if err == nil {
			t.Fatal("a mint body attached to no event — this is the foreign key the shard design could not have, and the reason a ruling can no longer precede its filing")
		}
	})
}

// The counter names each fixture act uniquely. It was a `seq` column; the record no longer keeps
// a per-seat counter, and the key is the fact that has to be distinct.
var seqN int

func seq() int { seqN++; return seqN }

// THE VOCABULARY IS IN THE DATABASE, WITH ITS MEANINGS.
//
// A `CHECK (x IN (…))` enforces the same set and keeps only the words. descriptions.go exists to
// stop precisely that — "A VALUE CARRIES ITS OWN MEANING, OR THE SET IS A LIST OF NOUNS" — and the
// meanings had already been lost once, sitting in source comments no seat could read. Putting only
// the words into the schema would lose them again at the layer that IS the record.
func TestEnumVocabulariesAreQueryableWithTheirMeanings(t *testing.T) {
	db := open(t)

	var means string
	err := db.QueryRow(`SELECT means FROM enum_disposition WHERE value = 'defect_accepted'`).Scan(&means)
	if err != nil {
		t.Fatalf("the closure classes are not in the database: %v", err)
	}
	if !strings.Contains(means, "risk") {
		t.Errorf("defect_accepted means %q — the row carries a word and not what it is FOR", means)
	}

	// Every vocabulary row must say something. A blank means is a value with no meaning, which is
	// the state the docs table refuses at the Go layer and which must not reappear here.
	var blank int
	if err := db.QueryRow(`SELECT count(*) FROM enum_disposition WHERE trim(means) = ''`).Scan(&blank); err != nil {
		t.Fatal(err)
	}
	if blank != 0 {
		t.Errorf("%d closure classes carry an empty meaning", blank)
	}
}

// THE FOREIGN KEY DOES THE WORK THE CHECK USED TO.
//
// Swapping a CHECK for a lookup table is only safe if the refusal survives the swap, so it is
// driven rather than assumed: an unknown word must still be refused, and a known one still accepted.
func TestAnEnumColumnStillRefusesAnUnknownWord(t *testing.T) {
	db := open(t)
	mk := func(t *testing.T, typ string) int64 {
		t.Helper()
		res, err := db.Exec(`INSERT INTO events (seat_id, round, key, ts, type) VALUES ('red-merge-r1', 1, 'red-merge-r1:act:#' || ?, '2026-01-01T00:00:00Z', ?)`, seq(), typ)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		return id
	}

	// The gaps must EXIST before anything can close them: gap_id is a foreign key onto mint now, so
	// a fixture that closes a gap nobody minted is refused for a reason that has nothing to do with
	// the vocabulary this test is about.
	mintGap := func(t *testing.T, gapID string) {
		t.Helper()
		mid := mk(t, "mint")
		if _, err := db.Exec(`INSERT INTO mint (event_id, gap_id, class, problem, acceptance_check, check_kind, likelihood, impact)
			VALUES (?, ?, 'c', 'p', 'a', 'document', 'medium', 'medium')`, mid, gapID); err != nil {
			t.Fatal(err)
		}
	}
	mintGap(t, "R1-1")
	mintGap(t, "R1-2")

	id := mk(t, "close")
	if _, err := db.Exec(`INSERT INTO close (event_id, gap_id, closure_class, prose) VALUES (?, 'R1-1', 'evidence-rebutted', 'x')`, id); err == nil {
		t.Fatal("`evidence-rebutted` was accepted as a closure class — it is not one, and a graph fixture " +
			"asserted against it for months while the counters read zero")
	}

	id = mk(t, "close")
	if _, err := db.Exec(`INSERT INTO close (event_id, gap_id, closure_class, prose) VALUES (?, 'R1-2', 'defect_accepted', 'x')`, id); err != nil {
		t.Fatalf("a real closure class was refused: %v — the vocabulary table is a wall rather than a gate", err)
	}
}

// EVERY ENUM COLUMN HAS ITS VOCABULARY BEHIND IT, INCLUDING THE ONES IN ONEOF ARMS.
//
// This is the check the golden made me write. `armTable` emitted columns and CHECKs and dropped
// foreign keys entirely, so `motion_grade.dimension`, `motion_grade.proposed` and
// `motion_petition.class` sat in the record with no vocabulary — three of the columns a seat
// actually argues about, unconstrained, next to `motion_rule.binds` which was constrained. The
// whole suite was green. It had to be: every other test asserts a constraint somebody thought of,
// and this was a constraint that was never emitted at all.
//
// It asks SQLITE, not the DDL string. `strings.Contains(ddl, "FOREIGN KEY")` would pass on a
// foreign key attached to the wrong column, in the wrong table, or naming a vocabulary that does
// not exist — and the point of this test is that the generator's output is not trusted.
func TestEveryEnumColumnIsBackedByItsVocabulary(t *testing.T) {
	db := store(t)

	// What the descriptors say SHOULD be constrained: every non-repeated enum field of every body,
	// and of every message arm of every oneof, with the table name each one lands in.
	type col struct{ table, column, vocab string }
	var want []col
	var walk func(md protoreflect.MessageDescriptor, table string)
	walk = func(md protoreflect.MessageDescriptor, table string) {
		for i := 0; i < md.Fields().Len(); i++ {
			fd := md.Fields().Get(i)
			if fd.IsList() {
				continue // list values live in their own child table, which has no enum arm today
			}
			if m := fd.Message(); m != nil {
				walk(m, table+"_"+string(fd.Name()))
				continue
			}
			if fd.Kind() == protoreflect.EnumKind {
				want = append(want, col{table, string(fd.Name()), EnumTableName(fd.Enum())})
			}
		}
	}
	bodies, err := Bodies()
	if err != nil {
		t.Fatal(err)
	}
	for _, md := range bodies {
		walk(md, TableName(md))
	}
	if len(want) == 0 {
		t.Fatal("no enum columns found — a broken walk would pass every assertion below")
	}

	for _, c := range want {
		rows, err := db.Query(`SELECT "table", "from" FROM pragma_foreign_key_list(?)`, c.table)
		if err != nil {
			t.Errorf("%s: %v", c.table, err)
			continue
		}
		found := false
		for rows.Next() {
			var target, from string
			if err := rows.Scan(&target, &from); err != nil {
				t.Fatal(err)
			}
			if from == c.column && target == c.vocab {
				found = true
			}
		}
		rows.Close()
		if !found {
			t.Errorf("%s.%s holds values from %s and no foreign key enforces it — the column accepts "+
				"any text, so a word no seat could ever type through the CLI is storable by anything "+
				"writing SQL, and the record is EVIDENCE",
				c.table, c.column, c.vocab)
		}
	}
}
