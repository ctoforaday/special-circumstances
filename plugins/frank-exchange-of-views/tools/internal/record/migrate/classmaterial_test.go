package migrate_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

func openRO(t *testing.T, runDir string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db")+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// THE ARCHIVED b9 RUN's PASS stood over G3, a low_medium gap of cross-section-contradiction — a
// class the shipped table now makes `always`. Migration translates history: it does not re-judge
// that PASS under a rule it predates, so the PASS lands and the gap arrives material.
func TestMigratingAdmitsArchivedPassOverNowAlwaysGap(t *testing.T) {
	res, dst := migrateArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	passes := 0
	for _, e := range fam.Events {
		if g, ok := recordpb.BodyAs[*recordpb.Gate](e); ok && g.GetVerdict() == recordpb.Verdict_VERDICT_PASS {
			passes++
		}
	}
	if passes != 1 {
		t.Fatalf("the migrated record carries %d PASS gates, want the archive's one", passes)
	}
	g := fam.Gap("G3")
	if g == nil || !g.Open || !g.Material || g.Mint.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS ||
		recordpb.GradeMass(g.Severity) >= 2 {
		t.Fatalf("fixture: G3 is not an open, low-graded gap of an always class on the migrated record: %+v", g)
	}
}

// THE SAME BOARD, LIVE, IS REFUSED: the exemption is the migration's, not the gate's.
func TestLivePassOverOpenAlwaysGapRefused(t *testing.T) {
	_, dst := migrateArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	_, err := record.Append(record.Identity{Run: dst, SeatID: "red-chair"}, &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()})
	if err == nil || !strings.Contains(err.Error(), "material gap(s) still OPEN") || !strings.Contains(err.Error(), "G3") {
		t.Fatalf("a live PASS over the open always-class G3 must be refused: %v", err)
	}
}

// A PRE-CHANGE class_new of a slug the shipped table holds takes the table's value. The archived
// sleeper run coined structure-noncompliance, which the registry now carries as `never`.
func TestMigrateClassNewTakesCurrentRegistryDefault(t *testing.T) {
	res, dst := migrateArchive(t, "2026-08-23_sleeper-service-plan.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	var word string
	if err := openRO(t, dst.Dir()).QueryRow(`SELECT "material_default" FROM "class_new" WHERE "slug" = 'structure-noncompliance'`).Scan(&word); err != nil || word != "never" {
		t.Fatalf("structure-noncompliance's coining migrated with material_default %q (%v), want the shipped never", word, err)
	}
}

// ONE HELD BY THE TABLE TAKES ITS VALUE WITH NO FILL; ONE IT LACKS TAKES by_grade AS A STATED FILL.
func TestMigrateClassNewWithoutDefaultFillsFromTable(t *testing.T) {
	res, dst := migrateArchive(t, "2026-08-23_sleeper-service-plan.tar.gz")
	filled := map[string]migrate.StatedFill{}
	for _, f := range res.StatedFills {
		filled[f.Slug] = f
	}
	if _, ok := filled["structure-noncompliance"]; ok {
		t.Errorf("a slug the shipped table holds was recorded as a fill: %+v", filled["structure-noncompliance"])
	}
	rows, err := openRO(t, dst.Dir()).Query(`SELECT "slug", "material_default" FROM "class_new"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	lacking := 0
	for rows.Next() {
		var slug, word string
		if err := rows.Scan(&slug, &word); err != nil {
			t.Fatal(err)
		}
		if _, held := record.ShippedMaterialDefaults[slug]; held {
			continue
		}
		lacking++
		f, ok := filled[slug]
		if word != "by_grade" || !ok || f.Value != "by_grade" || !strings.HasPrefix(f.Where, "class_new event") || f.Why == "" {
			t.Errorf("%s: material_default %q, fill %+v — want by_grade recorded as a stated fill on its class_new", slug, word, f)
		}
	}
	if lacking == 0 || lacking != len(res.StatedFills) {
		t.Errorf("%d coined slugs the table lacks, %d fills — want one fill each and at least one", lacking, len(res.StatedFills))
	}
}

// A STAGED ROW THE TABLE LACKS, IN A REGISTRY THAT PREDATES THE FIELD, takes by_grade and the
// manifest names it; nothing is refused.
func TestMigrateStagedSlugAbsentFromTableFillsByGrade(t *testing.T) {
	src := recordtest.ExtractArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	p := filepath.Join(src, "records", "class-registry.json")
	var reg map[string]json.RawMessage
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &reg); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(reg["classes"], &rows); err != nil {
		t.Fatal(err)
	}
	rows = append(rows, map[string]any{"slug": "not-shipped-anywhere"})
	reg["classes"], _ = json.Marshal(rows)
	out, _ := json.Marshal(reg)
	if err := os.WriteFile(p, out, 0o644); err != nil {
		t.Fatal(err)
	}
	to := recordtest.TmpRun(t)
	m, err := migrate.Migrate(src, to, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	var fill *migrate.StatedFill
	for i, f := range m.StatedFills {
		if f.Slug == "not-shipped-anywhere" {
			fill = &m.StatedFills[i]
		}
	}
	if fill == nil || fill.Where != "staged registry" || fill.Value != "by_grade" {
		t.Fatalf("the manifest must name the staged row it filled: %+v", m.StatedFills)
	}
	got, err := os.ReadFile(filepath.Join(to, "records", "class-registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"slug":"not-shipped-anywhere"`) && !strings.Contains(string(got), `"not-shipped-anywhere"`) {
		t.Fatalf("the translated registry lost the row:\n%s", got)
	}
	var translated struct {
		Classes []struct {
			Slug            string  `json:"slug"`
			MaterialDefault *string `json:"material_default"`
		} `json:"classes"`
	}
	if err := json.Unmarshal(got, &translated); err != nil {
		t.Fatal(err)
	}
	for _, c := range translated.Classes {
		if c.MaterialDefault == nil {
			t.Errorf("row %s still lacks material_default after translation", c.Slug)
		}
		if c.Slug == "not-shipped-anywhere" && (c.MaterialDefault == nil || *c.MaterialDefault != "by_grade") {
			t.Errorf("the unshipped row took %v, want by_grade", c.MaterialDefault)
		}
	}
}

// thisBinaryRun is a source written by THIS binary: a run-coined `always` class and a mint of it,
// a mint of a registry `never` class, and a spot-check naming an area.
func thisBinaryRun(t *testing.T) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	run := runtest.Open(t, dir)
	lens, chair := "red-lens-evidence", "red-chair"
	if _, err := record.Append(record.Identity{Run: run, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: []string{lens, chair}}); err != nil {
		t.Fatal(err)
	}
	if err := record.StageForRunWithDefaults(run, map[string]recordpb.ClassMaterial{
		"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "shapey": recordpb.ClassMaterial_CLASS_MATERIAL_NEVER,
	}); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{lens, chair} {
		if _, _, err := record.RegisterSeat(record.Identity{Run: run, SeatID: s}, ""); err != nil {
			t.Fatal(err)
		}
	}
	mint := func(id, class string, sev recordpb.Grade) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(id), Class: proto.String(class), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: sev.Enum(),
			Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}
	}
	for _, act := range []struct {
		seat string
		body proto.Message
	}{
		{lens, &recordpb.ClassNew{Slug: proto.String("coined-here"), Definition: proto.String("d"), Neighbor: proto.String("g"),
			Distinguisher: proto.String("x"), MaterialDefault: recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS.Enum()}},
		{lens, mint("G1", "coined-here", recordpb.Grade_GRADE_LOW)},
		{lens, mint("G2", "shapey", recordpb.Grade_GRADE_HIGH)},
		{chair, &recordpb.SpotCheck{Areas: []string{lens}, Reason: proto.String("read the changes since the pin")}},
	} {
		if _, err := record.Append(record.Identity{Run: run, SeatID: act.seat}, act.body); err != nil {
			t.Fatalf("%T: %v", act.body, err)
		}
	}
	return dir
}

// A RUN-COINED DEFAULT IS KEPT: `coined-here` is absent from the shipped table, so a translation
// that did not keep a recorded value would make it by_grade and record a fill.
func TestMigrateClassNewWithDefaultKeepsIt(t *testing.T) {
	src := thisBinaryRun(t)
	to := recordtest.TmpRun(t)
	m, err := migrate.Migrate(src, to, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	var word string
	if err := openRO(t, to).QueryRow(`SELECT "material_default" FROM "class_new" WHERE "slug" = 'coined-here'`).Scan(&word); err != nil || word != "always" {
		t.Fatalf("coined-here migrated with %q (%v), want the recorded always", word, err)
	}
	if len(m.StatedFills) != 0 {
		t.Errorf("a recorded default produced fills: %+v", m.StatedFills)
	}
}

// MIGRATING A RUN WRITTEN BY THIS BINARY REFUSES NOTHING AND CHANGES NONE OF THE FIELDS THIS PLAN
// ADDED: class_material on each mint, material_default on the coining, the staged registry, and a
// spot-check's areas all arrive as the source holds them.
func TestMigrateRunWrittenByThisBinaryKeepsNewFields(t *testing.T) {
	src := thisBinaryRun(t)
	to := recordtest.TmpRun(t)
	m, err := migrate.Migrate(src, to, migrate.Entries(), migrate.Options{})
	if err != nil || len(m.Refusals) != 0 || len(m.StatedFills) != 0 {
		t.Fatalf("migrate: err=%v refusals=%+v fills=%+v", err, m.Refusals, m.StatedFills)
	}
	query := func(dir, q string) string {
		rows, err := openRO(t, dir).Query(q)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var a, b sql.NullString
			if err := rows.Scan(&a, &b); err != nil {
				t.Fatal(err)
			}
			out = append(out, a.String+"="+b.String)
		}
		return strings.Join(out, ",")
	}
	for _, q := range []string{
		`SELECT "gap_id", "class_material" FROM "mint" ORDER BY "gap_id"`,
		`SELECT "slug", "material_default" FROM "class_new" ORDER BY "slug"`,
		`SELECT "ord", "value" FROM "spot_check_areas" ORDER BY "ord"`,
	} {
		if s, d := query(src, q), query(to, q); s != d || s == "" {
			t.Errorf("%s\n  source: %s\n  migrated: %s", q, s, d)
		}
	}
	a, err := os.ReadFile(filepath.Join(src, "records", "class-registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(to, "records", "class-registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Errorf("the staged registry of a run written by this binary was rewritten:\n%s\n---\n%s", a, b)
	}
}
