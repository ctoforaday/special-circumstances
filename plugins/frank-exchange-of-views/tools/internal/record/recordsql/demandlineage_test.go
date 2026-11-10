package recordsql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// demandLineageOracleDiff holds demand_lineage and demand_reach against the recursive view they replace
// (testdata/oracle_demand_reach.sql), on one connection because the oracle is TEMP. It returns ""
// when they agree row for row, else what differs.
func demandLineageOracleDiff(t *testing.T, db *sql.DB) string {
	t.Helper()
	oracle, err := os.ReadFile("testdata/oracle_demand_reach.sql")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `DROP VIEW IF EXISTS temp."oracle_demand_reach"; DROP VIEW IF EXISTS temp."oracle_demand_lineage";`+string(oracle)); err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, q := range []struct{ what, sql string }{
		{"demand_lineage rows the oracle lacks", `SELECT "gap_id", "from_gap" FROM "demand_lineage" EXCEPT SELECT "gap_id", "from_gap" FROM "oracle_demand_lineage"`},
		{"oracle lineage rows demand_lineage lacks", `SELECT "gap_id", "from_gap" FROM "oracle_demand_lineage" EXCEPT SELECT "gap_id", "from_gap" FROM "demand_lineage"`},
		{"demand_reach rows the oracle lacks", `SELECT * FROM "demand_reach" EXCEPT SELECT * FROM "oracle_demand_reach"`},
		{"oracle demand_reach rows demand_reach lacks", `SELECT * FROM "oracle_demand_reach" EXCEPT SELECT * FROM "demand_reach"`},
		{"row counts (lineage, oracle lineage, reach, oracle reach) that differ", `SELECT a, b, c, d FROM (SELECT
		   (SELECT count(*) FROM "demand_lineage") AS a, (SELECT count(*) FROM (SELECT DISTINCT "gap_id", "from_gap" FROM "oracle_demand_lineage")) AS b,
		   (SELECT count(*) FROM "demand_reach") AS c, (SELECT count(*) FROM "oracle_demand_reach") AS d) WHERE a <> b OR c <> d`},
	} {
		rows, err := conn.QueryContext(ctx, q.sql)
		if err != nil {
			t.Fatal(err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			v := make([]any, len(cols))
			p := make([]any, len(cols))
			for i := range v {
				p[i] = &v[i]
			}
			if err := rows.Scan(p...); err != nil {
				t.Fatal(err)
			}
			out = append(out, fmt.Sprintf("%s: %v", q.what, v))
		}
		rows.Close()
	}
	return strings.Join(out, "\n")
}

// mintOf mints gap of kind about anchor (or a section when about is ""), naming supersedes.
func (r *demandRecord) mintOf(gap string, kind recordpb.CheckKind, about string, supersedes ...string) int64 {
	m := &recordpb.Mint{ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
		GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Problem: proto.String("p"), AcceptanceCheck: proto.String("a"), CheckKind: kind.Enum(),
		Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum(),
		AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), Supersedes: supersedes}
	if about != "" {
		m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), proto.String(about)
	}
	return r.add(reader, recordpb.EventType_EVENT_TYPE_MINT, m)
}

// lineageOf is the ancestors gap carries the demand from, itself included, sorted.
func (r *demandRecord) demandLineageOf(gap string) string {
	r.t.Helper()
	rows, err := r.db.Query(`SELECT "from_gap" FROM "demand_lineage" WHERE "gap_id" = ?`, gap)
	if err != nil {
		r.t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			r.t.Fatal(err)
		}
		got = append(got, g)
	}
	sort.Strings(got)
	return strings.Join(got, ",")
}

// THE STORED DEMAND LINEAGE IS THE RECURSIVE VIEW'S, AT EVERY EVENT, ON THE SHAPES NO ARCHIVED RUN HOLDS. The
// archive sweep (migrate) holds demand_lineage and demand_reach equal to the oracle on every archived run,
// and no archived run carries a demand from an ancestor; these fixtures do: a chain of same-kind
// supersessions, a second mint naming an ancestor already named, a supersession across check kinds, a
// non-evidence chain, an ancestor listed twice by one mint, a mint naming itself, and readings of
// the anchor an ancestor is about before and after its heir's mint.
func TestTheStoredDemandLineageIsTheRecursiveViews(t *testing.T) {
	r := newDemandRecord(t)
	src, comp, doc := recordpb.CheckKind_CHECK_KIND_SOURCE, recordpb.CheckKind_CHECK_KIND_COMPUTATION, recordpb.CheckKind_CHECK_KIND_DOCUMENT
	steps := []func(){
		func() { r.cite("C1", "") },
		func() { r.mintOf("A", src, "C1") },
		func() { r.verify(reader, "C1", "A", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES) },
		func() { r.mintOf("B", src, "", "A") },
		func() { r.mintOf("C", src, "", "B") },
		func() { r.verify(other, "C1", "A", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES) },
		func() { r.mintOf("D", src, "", "A") },
		func() { r.mintOf("E", comp, "", "C") },
		func() { r.mintOf("H", doc, "") },
		func() { r.mintOf("I", doc, "", "H") },
		func() { r.mintOf("J", src, "", "D", "D") },
		func() { r.mintOf("K", comp, "K1") },
		func() { r.mintOf("L", comp, "", "L", "K") },
	}
	for i, step := range steps {
		step()
		if d := demandLineageOracleDiff(t, r.db); d != "" {
			t.Fatalf("after step %d the stored demand lineage is not the recursive view's:\n%s", i, d)
		}
	}
	for gap, want := range map[string]string{
		"A": "A",
		"B": "A,B",
		"C": "A,B,C",
		"D": "D",   // A was named first by B: D's naming passes nothing on
		"E": "E",   // computation carries no source demand, and E names C first, so no later mint carries C's
		"I": "I",   // document demands no live answer
		"J": "D,J", // D named twice by one mint is carried once
		"L": "K,L", // a mint naming itself carries itself once
	} {
		if got := r.demandLineageOf(gap); got != want {
			t.Errorf("gap %s carries the demand from %q, want %q", gap, got, want)
		}
	}
	// The anchor the ancestor is about reaches the heir from the heir's mint, checked by the ancestor's
	// minter.
	var n int
	if err := r.db.QueryRow(`SELECT count(*) FROM "demand_reach" d JOIN "mint" m ON m."gap_id" = d."gap_id"
	    WHERE d."gap_id" = 'C' AND d."from_gap" = 'A' AND d."anchor" = 'C1' AND d."seat" = ? AND d."reached" = m."event_id"`, reader).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("C reaches A's anchor %d time(s) from its own mint, want 1", n)
	}
}

// A NAME THAT POINTS AT NO GAP CANNOT BE WRITTEN, because a gap's demand lineage is written at its mint and an
// ancestor minted later could never join it: a supersedes, or an upheld position's gap list, naming a
// gap not yet minted is refused by its foreign key. The demand lineage is append-only, as its rows are.
func TestADemandLineageNamesOnlyMintedGaps(t *testing.T) {
	r := newDemandRecord(t)
	r.mintOf("A", recordpb.CheckKind_CHECK_KIND_SOURCE, "")

	ev := event(t, 99, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
		GapId: proto.String("B"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Problem: proto.String("p"), AcceptanceCheck: proto.String("a"), CheckKind: recordpb.CheckKind_CHECK_KIND_SOURCE.Enum(),
		Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum(),
		Supersedes: []string{"A", "Z"}})
	if _, err := Insert(r.db, ev); err == nil || !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Errorf("a mint superseding unminted gap Z = %v, want refused by the foreign key", err)
	}

	ev = event(t, 98, recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String("A-00000001"),
		Opinion:  proto.String("the objection holds"),
		Ruling: &recordpb.MotionRule_Position{Position: &recordpb.PositionHearing{
			Ruling:    recordpb.PositionRuling_POSITION_RULING_UPHELD.Enum(),
			GapIds:    []string{"A", "Z"},
			Direction: proto.String("source the figure"),
		}},
	})
	if _, err := Insert(r.db, ev); err == nil || !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Errorf("an upheld position naming unminted gap Z = %v, want refused by the foreign key", err)
	}

	for _, stmt := range []string{
		`UPDATE "demand_lineage" SET "from_gap" = 'A' WHERE "gap_id" = 'A'`,
		`DELETE FROM "demand_lineage"`,
	} {
		if _, err := r.db.Exec(stmt); err == nil || !strings.Contains(err.Error(), "fixed at its mint") {
			t.Errorf("%s = %v, want refused: the demand lineage is append-only", stmt, err)
		}
	}
	if got := r.demandLineageOf("A"); got != "A" {
		t.Errorf("A carries %q after the refused writes, want A", got)
	}
}
