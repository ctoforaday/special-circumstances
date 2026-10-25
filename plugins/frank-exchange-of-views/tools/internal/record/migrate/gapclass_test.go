package migrate_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/verify"
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
	g := fam.Gap(migratedID(res, "G", "G3"))
	if g == nil || !g.Open || !g.Material || g.Mint.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS ||
		recordpb.GradeMass(g.Severity) >= 2 {
		t.Fatalf("fixture: G3 is not an open, low-graded gap of an always class on the migrated record: %+v", g)
	}
}

// THE SAME BOARD, LIVE, IS REFUSED: the exemption is the migration's, not the gate's.
func TestLivePassOverOpenAlwaysGapRefused(t *testing.T) {
	res, dst := migrateArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	_, err := record.Append(record.Identity{Run: dst, SeatID: "red-chair"}, &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()})
	if err == nil || !strings.Contains(err.Error(), "material gap(s) still OPEN") || !strings.Contains(err.Error(), migratedID(res, "G", "G3")) {
		t.Fatalf("a live PASS over the open always-class G3 must be refused: %v", err)
	}
}

// thisBinaryRun is a source written by THIS binary: a mint of an `always` class, a mint of a
// `never` class, and a spot-check naming an area.
func thisBinaryRun(t *testing.T) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	run := runtest.Open(t, dir)
	lens, chair := "red-lens-evidence", "red-chair"
	if _, err := record.Append(record.Identity{Run: run, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: []string{lens, chair}}); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{lens, chair} {
		if _, _, err := record.RegisterSeat(record.Identity{Run: run, SeatID: s}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	mint := func(id string, class recordpb.GapClass, sev recordpb.Grade) *recordpb.Mint {
		return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(id), Class: class.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: sev.Enum(),
			Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}
	}
	for _, act := range []struct {
		seat string
		body proto.Message
	}{
		{lens, mint("G-00000001", recordpb.GapClass_GAP_CLASS_FALSE_UNIVERSAL, recordpb.Grade_GRADE_LOW)},
		{lens, mint("G-00000002", recordpb.GapClass_GAP_CLASS_STRUCTURE_NONCOMPLIANCE, recordpb.Grade_GRADE_HIGH)},
		{chair, &recordpb.SpotCheck{Areas: []string{lens}, Reason: proto.String("read the changes since the pin")}},
	} {
		if _, err := record.Append(record.Identity{Run: run, SeatID: act.seat}, act.body); err != nil {
			t.Fatalf("%T: %v", act.body, err)
		}
	}
	return dir
}

// MIGRATING A RUN WRITTEN BY THIS BINARY REFUSES NOTHING AND CHANGES NONE OF THE FIELDS THIS PLAN
// ADDED: class_material on each mint and a spot-check's areas arrive as the source holds them, and
// no class registry is written beside the record.
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
		`SELECT "ord", "value" FROM "spot_check_areas" ORDER BY "ord"`,
	} {
		if s, d := query(src, q), query(to, q); s != d || s == "" {
			t.Errorf("%s\n  source: %s\n  migrated: %s", q, s, d)
		}
	}
	if _, err := os.Stat(filepath.Join(to, "records", "class-registry.json")); !os.IsNotExist(err) {
		t.Errorf("a class registry was written beside the migrated record (%v)", err)
	}
}

// migratedPass is the one PASS gate on a migrated record, and the family it reads from.
func migratedPass(t *testing.T, archive string) (*migration.Manifest, record.Family, *recordpb.Gate, record.Run) {
	t.Helper()
	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	var passes []*recordpb.Gate
	for _, e := range fam.Events {
		if g, ok := recordpb.BodyAs[*recordpb.Gate](e); ok && g.GetVerdict() == recordpb.Verdict_VERDICT_PASS {
			passes = append(passes, g)
		}
	}
	if len(passes) != 1 {
		t.Fatalf("the migrated record carries %d PASS gates, want the archive's one", len(passes))
	}
	return res, fam, passes[0], dst
}

// THE ARCHIVED PASSES MIGRATION ADMITS CARRY THE FACT (fork (a), gblock 2026-09-15). b7's PASS stood
// over G2 (derivation-status-overclaim, low_medium) and b9's over G3 (cross-section-contradiction,
// low_medium), classes the shipped table makes `always`. Each migrated PASS carries exactly that
// gap in migration_admitted_gap_ids, and verify reports the PASS as admitted by migration, naming
// the gap — not the #67 violation. Migrating the migrated record again keeps the field.
func TestMigrationRecordsAdmittedPassOnArchives(t *testing.T) {
	for _, c := range []struct{ archive, gap string }{
		{"2026-09-11_is-91-prime-b7.tar.gz", "G2"},
		{"2026-09-11_is-91-prime-b9.tar.gz", "G3"},
	} {
		t.Run(c.archive, func(t *testing.T) {
			res, fam, pass, dst := migratedPass(t, c.archive)
			c.gap = migratedID(res, "G", c.gap)
			if got := pass.GetMigrationAdmittedGapIds(); len(got) != 1 || got[0] != c.gap {
				t.Fatalf("the migrated PASS carries migration_admitted_gap_ids %v, want [%s]", got, c.gap)
			}
			if g := fam.Gap(c.gap); g == nil || !g.Open || !g.Material || g.Mint.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS {
				t.Fatalf("fixture: %s is not an open, material gap of an always class on the migrated record: %+v", c.gap, g)
			}
			var check verify.Check
			for _, ck := range verify.Run(fam) {
				if ck.Name == "pass-closes-all-gaps" {
					check = ck
				}
			}
			if !check.OK || check.NA || len(check.Violations) != 0 || len(check.Admitted) != 1 || check.Admitted[0] != c.gap ||
				!strings.Contains(check.Detail, "admitted by migration") {
				t.Errorf("verify must report the PASS as admitted by migration over %s, not [FAIL]: %+v", c.gap, check)
			}
			again := recordtest.TmpRun(t)
			if m, err := migrate.Migrate(dst.Dir(), again, migrate.Entries(), migrate.Options{}); err != nil || len(m.Refusals) != 0 {
				t.Fatalf("re-migrating the migrated record: %v (refusals %+v)", err, m.Refusals)
			}
			var kept []string
			if err := func() error {
				rows, err := openRO(t, again).Query(`SELECT "value" FROM "gate_migration_admitted_gap_ids" ORDER BY "ord"`)
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
					var v string
					if err := rows.Scan(&v); err != nil {
						return err
					}
					kept = append(kept, v)
				}
				return rows.Err()
			}(); err != nil || len(kept) != 1 || kept[0] != c.gap {
				t.Errorf("re-migration kept migration_admitted_gap_ids %v (%v), want [%s]", kept, err, c.gap)
			}
		})
	}
}

// A PRE-FIELD MINT TAKES ITS NEAREST VALUE'S DEFAULT (F9). b6's G2 (`inquiry-fate-mismatch`,
// low_medium) was minted before class_material, so the live rule stamps it: the slug translates to
// claim_contradicts_own_record, whose default is `always`, and a stated fill names it. b6's PASS
// stood over G2 and G7 open; as `always` both are material, the PASS gate is exempt under the
// migration, and the migrated PASS records exactly those two as admitted by it.
func TestAPreFieldMintTakesItsNearestValuesDefault(t *testing.T) {
	res, fam, pass, dst := migratedPass(t, "2026-09-11_is-91-prime-b6.tar.gz")
	g2, g7 := migratedID(res, "G", "G2"), migratedID(res, "G", "G7")
	var class, stamp string
	if err := openRO(t, dst.Dir()).QueryRow(`SELECT "class", "class_material" FROM "mint" WHERE "gap_id" = ?`, g2).Scan(&class, &stamp); err != nil {
		t.Fatalf("G2 (%s) on the migrated record: %v", g2, err)
	}
	if class != "claim_contradicts_own_record" || stamp != "always" {
		t.Errorf("G2 migrated as %s stamped %s; want claim_contradicts_own_record stamped always", class, stamp)
	}
	stated := false
	for _, f := range res.StatedFills {
		if f.Where == "mint "+g2 && f.Slug == "inquiry-fate-mismatch" && f.Value == "always" {
			stated = true
		}
	}
	if !stated {
		t.Errorf("no stated fill names G2's stamp: %+v", res.StatedFills)
	}
	got := slices.Sorted(slices.Values(pass.GetMigrationAdmittedGapIds()))
	if want := slices.Sorted(slices.Values([]string{g2, g7})); !slices.Equal(got, want) {
		t.Errorf("the migrated PASS admits %v; want G2's and G7's migrated ids %v", got, want)
	}
	if g := fam.Gap(g2); g == nil || !g.Open || !g.Material {
		t.Errorf("G2 is not open and material on the migrated record: %+v", g)
	}
}

// A MIGRATED PASS OVER ONLY NON-MATERIAL OPEN GAPS CARRIES NO ADMISSION. A seeded source: a lens
// mints a spec_underspecification gap (by_grade) at low, and the chair's PASS stands over it open.
// The gate the migration is exempt from would not have refused, so nothing is admitted.
func TestMigratedPassOverOnlyNonMaterialGapsCarriesNoAdmission(t *testing.T) {
	dir := recordtest.TmpRun(t)
	run := runtest.Open(t, dir)
	lens, chair := "red-lens-logic", "red-chair"
	if _, err := record.Append(record.Identity{Run: run, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: []string{lens, chair}}); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{lens, chair} {
		if _, _, err := record.RegisterSeat(record.Identity{Run: run, SeatID: s}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := record.Append(record.Identity{Run: run, SeatID: lens}, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String("G-00000001"),
		Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: recordpb.Grade_GRADE_LOW.Enum(),
		Likelihood: recordpb.Grade_GRADE_LOW.Enum(), Impact: recordpb.Grade_GRADE_LOW.Enum()}); err != nil {
		t.Fatal(err)
	}
	// Seeded, not appended: the live gate's own rule is not this test's subject, and the stamp is
	// the write's, so the replay reads it back.
	pass := recordtest.Event(t, chair, &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()})
	pass.Ts = proto.String(record.Stamp())
	recordtest.Seed(t, dir, pass)
	to := recordtest.TmpRun(t)
	m, err := migrate.Migrate(dir, to, migrate.Entries(), migrate.Options{})
	if err != nil || len(m.Refusals) != 0 {
		t.Fatalf("migrate: %v (refusals %+v)", err, m.Refusals)
	}
	fam, err := record.FamilyOf(runtest.Open(t, to))
	if err != nil {
		t.Fatal(err)
	}
	if g := fam.Gap("G-00000001"); g == nil || !g.Open || g.Material {
		t.Fatalf("fixture: the gap is not open and non-material on the migrated record: %+v", g)
	}
	passes := 0
	for _, e := range fam.Events {
		if g, ok := recordpb.BodyAs[*recordpb.Gate](e); ok && g.GetVerdict() == recordpb.Verdict_VERDICT_PASS {
			passes++
			if got := g.GetMigrationAdmittedGapIds(); len(got) != 0 {
				t.Errorf("a PASS over only non-material open gaps carries migration_admitted_gap_ids %v", got)
			}
		}
	}
	if passes != 1 {
		t.Fatalf("the migrated record carries %d PASS gates, want the source's one", passes)
	}
}

// classWords is every GapClass word.
func classWords() map[string]bool {
	out := map[string]bool{}
	vs := recordpb.GapClass(0).Descriptor().Values()
	for i := 1; i < vs.Len(); i++ {
		out[recordpb.Spelling(vs.Get(i))] = true
	}
	return out
}

// THE GAP CLASS MIGRATES ONTO THE ENUM, on the two eras that coined classes, each driven as the
// shipped path runs — Migrate on the extracted archive, its channels and all. m28 (current epoch,
// two class_new events, two invalid-inference mints, every mint stamped): every class a GapClass
// word, every stamp the source's, nothing refused, one slug through an authored row. The sleeper
// run (jsonl era, coinings carried on the mint, no class_material): a coined
// citation-support-mismatch lands on citation_figure_misattribution stamped `always` and the three
// guardrail-scope-overclaim mints on partial_control_coverage `by_grade`, each with a stated fill,
// and a slug that is a current word (structure-noncompliance) gets no translation row. Neither
// migrated record holds a class_new event or a staged registry.
func TestTheGapClassMigratesOntoTheEnum(t *testing.T) {
	words := classWords()
	mintsOf := func(t *testing.T, dir string) map[string][2]string {
		rows, err := openRO(t, dir).Query(`SELECT "gap_id", "class", "class_material" FROM "mint"`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		out := map[string][2]string{}
		for rows.Next() {
			var id, class, stamp sql.NullString
			if err := rows.Scan(&id, &class, &stamp); err != nil {
				t.Fatal(err)
			}
			out[id.String] = [2]string{class.String, stamp.String}
		}
		return out
	}
	common := func(t *testing.T, name string) (*migration.Manifest, string, map[string][2]string) {
		res, dir := migrate.MigratedArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
		}
		got := mintsOf(t, dir)
		for id, m := range got {
			if !words[m[0]] {
				t.Errorf("mint %s migrated with class %q, which is no GapClass word", id, m[0])
			}
		}
		var n int
		if err := openRO(t, dir).QueryRow(`SELECT count(*) FROM "events" WHERE "type" = 'class_new'`).Scan(&n); err != nil || n != 0 || res.Out["class_new"] != 0 {
			t.Errorf("the migrated record holds %d class_new event(s) (%v)", n, err)
		}
		if _, err := os.Stat(filepath.Join(dir, "records", "class-registry.json")); !os.IsNotExist(err) {
			t.Errorf("a class registry is staged beside the migrated record (%v)", err)
		}
		return res, dir, got
	}
	t.Run("m28", func(t *testing.T) {
		const name = "2026-10-18_is-91-prime-m28.tar.gz"
		res, _, got := common(t, name)
		src := mintsOf(t, recordtest.RequireArchive(t, name))
		if len(got) != len(src) || len(src) == 0 {
			t.Fatalf("%d mint(s) migrated of the source's %d", len(got), len(src))
		}
		for id, s := range src {
			if got[id][1] != s[1] {
				t.Errorf("mint %s migrated stamped %q; the source stamped %q", id, got[id][1], s[1])
			}
		}
		if want := []migration.ClassTranslation{{From: "audited-artifact-sibling-halo", To: "verification_scope_blindspot", Mints: 1}}; !slices.Equal(res.ClassTranslations, want) {
			t.Errorf("class_translations = %+v, want %+v", res.ClassTranslations, want)
		}
		if len(res.StatedFills) != 0 {
			t.Errorf("every m28 mint carries its stamp, yet %d fill(s) are stated: %+v", len(res.StatedFills), res.StatedFills)
		}
	})
	t.Run("sleeper", func(t *testing.T) {
		const name = "2026-08-23_sleeper-service-plan.tar.gz"
		res, _, got := common(t, name)
		js, err := migrate.OpenJSONL(filepath.Join(recordtest.RequireArchive(t, name), "records"))
		if err != nil {
			t.Fatal(err)
		}
		evs, err := js.Events()
		if err != nil {
			t.Fatal(err)
		}
		want := map[string][2]string{
			"citation-support-mismatch": {"citation_figure_misattribution", "always"},
			"guardrail-scope-overclaim": {"partial_control_coverage", "by_grade"},
		}
		seen := map[string]int{}
		for _, old := range evs {
			slug, _ := old.Fields["class"].(string)
			w, ok := want[slug]
			if old.Word != "mint" || !ok {
				continue
			}
			archived, _ := old.Fields["gap_id"].(string)
			id := res.IDs[archived]
			seen[slug]++
			if got[id] != w {
				t.Errorf("%s (%s, %s) migrated as %v, want %v", archived, slug, id, got[id], w)
			}
			if !slices.ContainsFunc(res.StatedFills, func(f migration.StatedFill) bool {
				return f.Where == "mint "+id && f.Slug == slug && f.Value == w[1]
			}) {
				t.Errorf("no stated fill names %s's stamp", id)
			}
		}
		if seen["citation-support-mismatch"] != 1 || seen["guardrail-scope-overclaim"] != 3 {
			t.Fatalf("fixture: the sleeper run mints %v; want one citation-support-mismatch and three guardrail-scope-overclaim", seen)
		}
		var guardrail bool
		for _, r := range res.ClassTranslations {
			if r == (migration.ClassTranslation{From: "guardrail-scope-overclaim", To: "partial_control_coverage", Mints: 3}) {
				guardrail = true
			}
			if r.From == "structure-noncompliance" {
				t.Errorf("structure-noncompliance resolves by the fold, yet has a translation row: %+v", r)
			}
		}
		if !guardrail {
			t.Errorf("class_translations lacks guardrail-scope-overclaim -> partial_control_coverage (3): %+v", res.ClassTranslations)
		}
	})
}

// A REFUSED PRE-FIELD MINT STATES NO FILL. The stamp is supplied at the write, so a mint the write
// refuses stamped nothing: here a lens mints before any cast or report, the write refuses it, and the
// manifest states the refusal and no fill, as it counts no class translation.
func TestARefusedPreFieldMintStatesNoFill(t *testing.T) {
	old := newOldRecord(t)
	for _, ddl := range []string{
		`INSERT INTO "enum_event_type" ("value", "means") VALUES ('mint', 'a gap')`,
		`CREATE TABLE "mint" ("event_id" INTEGER PRIMARY KEY, "gap_id" TEXT, "class" TEXT, "location" TEXT, "problem" TEXT, "required_fix" TEXT, "acceptance_check" TEXT, "severity" TEXT, "mint_reason" TEXT)`,
	} {
		if _, err := old.db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	old.row(t, "register", `"event_id", "tool_version"`, old.event(t, "red-lens-logic", 0, ts1, "register"), "0.1.0")
	old.row(t, "mint", `"event_id", "gap_id", "class", "location", "problem", "required_fix", "acceptance_check", "severity", "mint_reason"`,
		old.event(t, "red-lens-logic", 0, ts2, "mint"), "G1", "guardrail-scope-overclaim", "the sentence", "the problem", "the fix", "the check", "high", "the reason")

	m, err := migrate.Migrate(old.dir, recordtest.TmpRun(t), migrate.Entries(), migrate.Options{})
	if m == nil || len(m.Refusals) != 1 || m.Refusals[0].Word != "mint" {
		t.Fatalf("migrating a mint the write refuses: manifest %+v, error %v", m, err)
	}
	if len(m.StatedFills) != 0 || len(m.ClassTranslations) != 0 {
		t.Errorf("the refused mint left fills %+v and translations %+v; it stamped nothing", m.StatedFills, m.ClassTranslations)
	}
}
