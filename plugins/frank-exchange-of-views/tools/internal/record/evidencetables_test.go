package record

import (
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EVERY BACKING KIND HAS ONE EVIDENCE ROW, AND EVERY ROW NAMES WHAT THE RECORD HOLDS. The anchor kinds
// table says which kinds back their sentence; recordsql.EvidenceTables says what each such kind is on
// the record. The two are held level here, and each column to what reads it: the generated DDL for
// the tables, columns and the holds predicate, the check-kind vocabulary for `check`, and the anchor
// ledger for `failed` — the state a check of the kind that finds against its sentence leaves.
func TestEveryBackingKindHasAnEvidenceRow(t *testing.T) {
	var backs []string
	for _, k := range anchor.Kinds() {
		// Kinds are named; Backs reads an id, so ask it of an id of the kind.
		if anchor.Backs(idOfKind(t, k)) {
			backs = append(backs, k)
		}
	}
	var rows []string
	for _, r := range recordsql.EvidenceTables {
		rows = append(rows, r.Kind)
	}
	if !slices.Equal(backs, rows) {
		t.Fatalf("the backing kinds are %v and the evidence rows %v: every kind that backs its sentence has one row, in the kinds table's order, and no row names a kind that backs nothing", backs, rows)
	}

	// EVERY CHECK KIND THAT DEMANDS EVIDENCE HAS ONE ROW: each word but `document`, which a prose
	// answer settles, and no two rows answer one check kind.
	var demanded []string
	for _, w := range recordpb.Words(recordpb.CheckKind(0).Descriptor()) {
		if w != recordpb.Word(recordpb.CheckKind_CHECK_KIND_DOCUMENT) {
			demanded = append(demanded, w)
		}
	}
	var checks []string
	for _, r := range recordsql.EvidenceTables {
		checks = append(checks, r.Check)
	}
	slices.Sort(demanded)
	slices.Sort(checks)
	if !slices.Equal(demanded, checks) {
		t.Errorf("the check kinds an evidence anchor answers are %v, and the check kinds that demand one are %v", checks, demanded)
	}

	// A STANDING CHECK AGAINST THE SENTENCE IS A STATE OF THE LEDGER FOR A COMPUTATION ALONE: a
	// standing unsound judgement leaves the computation unsound; a citation's standing reading that
	// finds against leaves the citation holding (the reading is its lens's debt, readingDebtsOf).
	for _, r := range recordsql.EvidenceTables {
		id := idOfKind(t, r.Kind)
		s := &anchorStanding{anchor: id, computation: r.Kind == "computation",
			checks: []anchorCheck{{id: 1, seat: "red-lens-x", reads: true}}}
		want := stateHolds
		if r.Kind == "computation" {
			want = stateUnsound
		}
		if got := s.state(); got != want {
			t.Errorf("%s: a standing check against its sentence leaves the ledger state %v, want %v", r.Kind, got, want)
		}
		if r.Noun == "" {
			t.Errorf("%s: no noun — every refusal and result names the kind by it", r.Kind)
		}
	}

	// EVERY TABLE AND COLUMN IS ONE THE GENERATED DDL HOLDS, and the holds predicate reads that DDL.
	db := scratchStore(t)
	hasColumn := func(table, col string) bool {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM pragma_table_info(?) WHERE "name" = ?`, table, col).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n == 1
	}
	for _, r := range recordsql.EvidenceTables {
		for _, tc := range [][2]string{{r.Table, r.IDColumn}, {r.CheckTable, r.CheckIDColumn}} {
			if !hasColumn(tc[0], tc[1]) {
				t.Errorf("%s: the generated DDL holds no column %q.%q", r.Kind, tc[0], tc[1])
			}
		}
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM "` + r.CheckTable + `" c WHERE ` + r.Holds("c")).Scan(&n); err != nil {
			t.Errorf("%s: the holds predicate does not read the DDL: %v", r.Kind, err)
		}
	}
}

// THE HOLDS PREDICATE IS THE LEDGER'S HOLDS. For every value of what a check records — a reading's
// outcome and completeness, a judgement's soundness — the SQL predicate over the check's row and the
// ledger's own reading of the same check agree.
func TestEachEvidenceRowsHoldsIsTheLedgers(t *testing.T) {
	db := scratchStore(t)
	// holds asks the row's predicate of one check row, built from the values a check records.
	holds := func(kind, cols string, args ...any) bool {
		t.Helper()
		r, _ := recordsql.EvidenceTableOf(kind)
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM (SELECT `+cols+`) c WHERE `+r.Holds("c"), args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n == 1
	}
	outcomes := recordpb.SourceOutcome(0).Descriptor().Values()
	completeness := recordpb.SourceCompleteness(0).Descriptor().Values()
	for i := 0; i < outcomes.Len(); i++ {
		o := recordpb.SourceOutcome(outcomes.Get(i).Number())
		if o == recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED {
			continue
		}
		// Completeness is optional on a reading: its absent value is a NULL column.
		for j := 0; j < completeness.Len(); j++ {
			c := recordpb.SourceCompleteness(completeness.Get(j).Number())
			v := &recordpb.Verify{Outcome: o.Enum()}
			var col any
			if c != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_UNSPECIFIED {
				v.SourceCompleteness, col = c.Enum(), recordpb.Word(c)
			}
			want := VerifyReads(v) && !FindsAgainst(v)
			if got := holds("citation", `? AS "outcome", ? AS "source_completeness"`, recordpb.Word(o), col); got != want {
				t.Errorf("a reading %s of a copy %q: the row's holds says %v, the ledger %v", recordpb.Word(o), col, got, want)
			}
		}
	}
	sound := recordpb.Soundness(0).Descriptor().Values()
	for i := 0; i < sound.Len(); i++ {
		s := recordpb.Soundness(sound.Get(i).Number())
		var col any
		if s != recordpb.Soundness_SOUNDNESS_UNSPECIFIED {
			col = recordpb.Word(s)
		}
		if got, want := holds("computation", `? AS "soundness"`, col), s == recordpb.Soundness_SOUNDNESS_SOUND; got != want {
			t.Errorf("a judgement %q: the row's holds says %v, the ledger %v", col, got, want)
		}
	}
}

// idOfKind is an id of the anchor kind named k.
func idOfKind(t *testing.T, k string) string {
	t.Helper()
	return anchor.ID(k, [4]byte{0, 0, 0, 1})
}

// scratchStore is an empty record store under the test's temp dir, holding the generated DDL.
func scratchStore(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "records", "record.db")
	db, err := recordsql.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recordsql.Close(path) })
	return db
}
