package recordsql

import (
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

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

// THE STORED DEMAND LINEAGE IS EACH GAP'S ANCESTORS BY ITS DEMAND, ON THE SHAPES NO ARCHIVED RUN HOLDS.
// No archived run carries a demand from an ancestor; these fixtures do: a chain of same-kind
// supersessions, a second mint naming an ancestor already named, a supersession across check kinds, a
// non-evidence chain, an ancestor listed twice by one mint, a mint naming itself, and a reading
// carried by an ancestor before and after its heir's mint.
func TestTheStoredDemandLineageIsEachGapsAncestry(t *testing.T) {
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
	for _, step := range steps {
		step()
	}
	for gap, want := range map[string]string{
		"A": "A",
		"B": "A,B",
		"C": "A,B,C",
		"D": "D", // A was named first by B: D's naming passes nothing on
		"H": "H",
		"E": "E",   // computation carries no source demand, and E names C first, so no later mint carries C's
		"I": "I",   // document demands no live answer
		"J": "D,J", // D named twice by one mint is carried once
		"K": "K",
		"L": "K,L", // a mint naming itself carries itself once
	} {
		if got := r.demandLineageOf(gap); got != want {
			t.Errorf("gap %s carries the demand from %q, want %q", gap, got, want)
		}
	}
	// A reading carried by the ancestor reaches the heir from the heir's mint, whether the reading came
	// before the heir's mint (reader's) or after it (other's).
	var n int
	if err := r.db.QueryRow(`SELECT count(*) FROM "demand_reach" d JOIN "mint" m ON m."gap_id" = d."gap_id"
	    WHERE d."gap_id" = 'C' AND d."from_gap" = 'A' AND d."via" = 'reading' AND d."reached" = m."event_id"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("C reaches %d of A's two carried readings from its own mint, want 2", n)
	}
	// THE WHOLE REACH: each gap's own about and readings reached at their own acts, and every row it
	// inherits reached at its own mint ("mint" below); a gap about a section reaches nothing of its own.
	rows, err := r.db.Query(`SELECT d."gap_id" || ' ' || d."from_gap" || ' ' || d."via" || ' ' || d."anchor" || ' ' || d."seat" || ' ' ||
	    CASE WHEN d."reached" = d."minted" THEN 'mint' ELSE (SELECT e."type" FROM "events" e WHERE e."id" = d."reached") END
	  FROM "demand_reach" d ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var reach []string
	for rows.Next() {
		var row string
		if err := rows.Scan(&row); err != nil {
			t.Fatal(err)
		}
		reach = append(reach, row)
	}
	want := []string{
		"A A about C1 " + reader + " mint",
		"A A reading C1 " + reader + " verify",
		"A A reading C1 " + other + " verify",
		"B A about C1 " + reader + " mint",
		"B A reading C1 " + reader + " mint",
		"B A reading C1 " + other + " mint",
		"C A about C1 " + reader + " mint",
		"C A reading C1 " + reader + " mint",
		"C A reading C1 " + other + " mint",
		"K K about K1 " + reader + " mint",
		"L K about K1 " + reader + " mint",
	}
	if got, w := strings.Join(reach, "\n"), strings.Join(want, "\n"); got != w {
		t.Errorf("demand_reach holds\n%s\nwant\n%s", got, w)
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
