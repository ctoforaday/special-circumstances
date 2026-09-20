package record

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// liveRegistryRun is a run whose registry stages these classes with these defaults, with a lens
// registered to mint through the real write path — the path that stamps class_material.
func liveRegistryRun(t *testing.T, defaults map[string]recordpb.ClassMaterial) (Run, Identity) {
	t.Helper()
	dir := recordtest.TmpRun(t)
	writeRunConfig(t, dir, `{"mintBudget":50}`)
	run := mustRun(t, dir)
	if err := StageForRunWithDefaults(run, defaults); err != nil {
		t.Fatal(err)
	}
	lens := Identity{Run: run, SeatID: evLens}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	return run, lens
}

func liveMint(id, class string, sev recordpb.Grade) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(id), Class: proto.String(class), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(sev),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
}

// viewMaterial is the gap view's "material" column, per gap.
func viewMaterial(t *testing.T, run Run) map[string]bool {
	t.Helper()
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		t.Fatalf("open: %v", err)
	}
	rows, err := db.Query(`SELECT "gap_id", "material" FROM "gap"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		var m bool
		if err := rows.Scan(&id, &m); err != nil {
			t.Fatal(err)
		}
		out[id] = m
	}
	return out
}

// familyGaps is the Go family fold, per gap.
func familyGaps(t *testing.T, run Run) map[string]*Gap {
	t.Helper()
	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	return GapsByID(fam.Gaps)
}

func appendOK(t *testing.T, id Identity, body proto.Message) {
	t.Helper()
	if _, err := Append(id, body); err != nil {
		t.Fatalf("%T: %v", body, err)
	}
}

// A claim-class defect changes what a reader concludes whatever it is graded, so a LOW gap of an
// `always` class is material on both carriers of the definition.
func TestMaterialAlwaysHoldsLowGrade(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"claimy": recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS})
	appendOK(t, lens, liveMint("G1", "claimy", recordpb.Grade_GRADE_LOW))
	g := familyGaps(t, run)["G1"]
	if g.Mint.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS {
		t.Errorf("the mint was stamped %v, want the registry's always", g.Mint.GetClassMaterial())
	}
	if !g.Material || !viewMaterial(t, run)["G1"] {
		t.Errorf("a low gap of an always class must be material: family %v, view %v", g.Material, viewMaterial(t, run)["G1"])
	}
}

// A template-shape defect changes no conclusion, so even a HIGH gap of a `never` class is not.
func TestMaterialNeverReleasesHighGrade(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"shapey": recordpb.ClassMaterial_CLASS_MATERIAL_NEVER})
	appendOK(t, lens, liveMint("G1", "shapey", recordpb.Grade_GRADE_HIGH))
	if g := familyGaps(t, run)["G1"]; g.Material || viewMaterial(t, run)["G1"] {
		t.Errorf("a high gap of a never class must not be material: family %v, view %v", g.Material, viewMaterial(t, run)["G1"])
	}
}

// THE TWO CARRIERS OF THE ONE DEFINITION AGREE, gap by gap, over every class default and the grades
// either side of the by-grade floor — and on a regrade, because materiality reads the grade NOW.
func TestFamilyGapMaterialMatchesView(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{
		"a": recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS,
		"n": recordpb.ClassMaterial_CLASS_MATERIAL_NEVER,
		"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE,
	})
	want := map[string]bool{}
	n := 0
	for _, c := range []struct {
		class string
		sev   recordpb.Grade
		want  bool
	}{
		{"a", recordpb.Grade_GRADE_TRIVIAL, true}, {"a", recordpb.Grade_GRADE_MEDIUM, true}, {"a", recordpb.Grade_GRADE_HIGH, true},
		{"n", recordpb.Grade_GRADE_TRIVIAL, false}, {"n", recordpb.Grade_GRADE_MEDIUM, false}, {"n", recordpb.Grade_GRADE_CERTAIN, false},
		{"g", recordpb.Grade_GRADE_LOW_MEDIUM, false}, {"g", recordpb.Grade_GRADE_MEDIUM, true}, {"g", recordpb.Grade_GRADE_HIGH, true},
	} {
		n++
		id := "G" + string(rune('0'+n))
		appendOK(t, lens, liveMint(id, c.class, c.sev))
		want[id] = c.want
	}
	// A by_grade gap regraded up across the floor, and one regraded down below it.
	appendOK(t, lens, &recordpb.Regrade{GapId: proto.String("G7"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM_HIGH), Basis: proto.String("b")})
	appendOK(t, lens, &recordpb.Regrade{GapId: proto.String("G9"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("b")})
	want["G7"], want["G9"] = true, false

	view, fam := viewMaterial(t, run), familyGaps(t, run)
	for id, w := range want {
		if fam[id] == nil {
			t.Fatalf("%s is not in the family", id)
		}
		if fam[id].Material != view[id] {
			t.Errorf("%s: the family says material=%v and the view says %v — two carriers of one definition", id, fam[id].Material, view[id])
		}
		if view[id] != w {
			t.Errorf("%s: material=%v, want %v", id, view[id], w)
		}
	}
}

// A hand-built run's registry is `by_grade` throughout, so every existing caller keeps the
// materiality its grade alone gave it.
func TestStageForRunWritesByGrade(t *testing.T) {
	dir := recordtest.TmpRun(t)
	run := mustRun(t, dir)
	if err := StageForRun(run, "a", "b", "a"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(run.Records(), "class-registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	var reg struct {
		Classes []struct {
			Slug            string  `json:"slug"`
			MaterialDefault *string `json:"material_default"`
		} `json:"classes"`
	}
	if err := json.Unmarshal(b, &reg); err != nil {
		t.Fatal(err)
	}
	if len(reg.Classes) != 2 {
		t.Fatalf("staged %d rows, want a and b once each: %s", len(reg.Classes), b)
	}
	for _, c := range reg.Classes {
		if c.MaterialDefault == nil || *c.MaterialDefault != "by_grade" {
			t.Errorf("row %s: material_default %v, want by_grade", c.Slug, c.MaterialDefault)
		}
	}
	lens := Identity{Run: run, SeatID: evLens}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	appendOK(t, lens, liveMint("G1", "a", recordpb.Grade_GRADE_LOW))
	if got := familyGaps(t, run)["G1"].Mint.GetClassMaterial(); got != recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE {
		t.Errorf("a mint of a StageForRun class was stamped %v, want by_grade", got)
	}
}

// A class the run coins carries its default on the record, and a mint of it is stamped from there.
func TestClassNewCarriesMaterialDefault(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	coin := func(slug string, md *recordpb.ClassMaterial) *recordpb.ClassNew {
		return &recordpb.ClassNew{Slug: proto.String(slug), Definition: proto.String("d"), Neighbor: proto.String("g"),
			Distinguisher: proto.String("x"), MaterialDefault: md}
	}
	if _, err := Append(lens, coin("undeclared", nil)); err == nil || !strings.Contains(err.Error(), "--material-default") {
		t.Fatalf("a coining with no material default landed: %v", err)
	}
	appendOK(t, lens, coin("coined", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS.Enum()))
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	var word string
	if err := db.QueryRow(`SELECT "material_default" FROM "class_new" WHERE "slug" = 'coined'`).Scan(&word); err != nil || word != "always" {
		t.Fatalf("class_new.material_default = %q (%v), want always", word, err)
	}
	appendOK(t, lens, liveMint("G1", "coined", recordpb.Grade_GRADE_LOW))
	g := familyGaps(t, run)["G1"]
	if g.Mint.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS || !g.Material {
		t.Errorf("a mint of a coined always class: stamped %v, material %v", g.Mint.GetClassMaterial(), g.Material)
	}
}

// NO SEAT SETS IT. A lens that could choose its gap's materiality could mint every gap harmless;
// a migration's source is the record, so an archived value is kept as recorded.
func TestMintRefusesPresetClassMaterial(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	m := liveMint("G1", "g", recordpb.Grade_GRADE_HIGH)
	m.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_NEVER.Enum()
	if _, err := Append(lens, m); err == nil || !strings.Contains(err.Error(), "stamped by the tool") {
		t.Fatalf("a mint arriving with class_material set landed outside a migration: %v", err)
	}
	Migrating = true
	defer func() { Migrating = false }()
	m2 := liveMint("G2", "g", recordpb.Grade_GRADE_HIGH)
	m2.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS.Enum()
	appendOK(t, lens, m2)
	if got := familyGaps(t, run)["G2"].Mint.GetClassMaterial(); got != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS {
		t.Errorf("a migrated mint's recorded class_material became %v, want it kept as recorded (always)", got)
	}
}

// The refusal a lens meets when it mints ungraded names what severity now decides.
func TestSeverityRefusalWhyIsClassAware(t *testing.T) {
	_, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	m := liveMint("G1", "g", recordpb.Grade_GRADE_LOW)
	m.Severity = nil
	_, err := Append(lens, m)
	if err == nil || !strings.Contains(err.Error(), "--severity") ||
		!strings.Contains(err.Error(), "materiality reads it wherever the class goes by grade") {
		t.Fatalf("the severity refusal must say materiality reads it where the class goes by grade: %v", err)
	}
}

// A CURRENT run's staged registry altered after setup is refused naming the file and the slug. It
// says nothing about migrate: a current run has nothing to migrate, and the remedy is the row.
func TestLoadRegistryRefusalNamesRegistryNotMigrate(t *testing.T) {
	for _, c := range []struct{ name, body, names string }{
		{"missing", `{"classes":[{"slug":"g"}]}`, "no material_default"},
		{"unknown", `{"classes":[{"slug":"g","material_default":"sometimes"}]}`, `"sometimes"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
			p := filepath.Join(run.Records(), "class-registry.json")
			if err := os.WriteFile(p, []byte(c.body), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := Append(lens, liveMint("G1", "g", recordpb.Grade_GRADE_LOW))
			if err == nil {
				t.Fatal("a mint against an altered registry landed")
			}
			for _, want := range []string{p, `"g"`, c.names, "changed since"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal must name %q: %v", want, err)
				}
			}
			if strings.Contains(err.Error(), "migrate") {
				t.Errorf("a current run's registry refusal must not send the seat to migrate: %v", err)
			}
		})
	}
}

// A PRE-CHANGE RUN — its database and its staged registry both written before the field — is
// refused at the first write, naming what the database lacks and `migrate`.
func TestPreChangeRegistryRefusedNamingMigrate(t *testing.T) {
	dir := recordtest.ExtractArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	b, err := os.ReadFile(filepath.Join(dir, "records", "class-registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "material_default") {
		t.Fatal("fixture: the archived registry already carries material_default, so it is not a pre-change run")
	}
	run, err := NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Append(Identity{Run: run, SeatID: evLens}, liveMint("G99", "false-universal", recordpb.Grade_GRADE_LOW))
	if err == nil {
		t.Fatal("a mint into a pre-change run landed")
	}
	for _, want := range []string{`"mint" table has no "class_material" column`, "migrate"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name %q: %v", want, err)
		}
	}
}

// A LIVE VERDICT CANNOT CLAIM A MIGRATION'S ADMISSION (fork (a)). On a board where a plain PASS
// appends, a PASS carrying migration_admitted_gap_ids is refused, naming the field; so is one
// claiming admission over an open material gap, where the refusal must be the admission's and not
// the material gate's; so is a FAIL carrying it. The plain PASS appends and carries no admission.
func TestLiveGateCarryingMigrationAdmissionRefused(t *testing.T) {
	board := func(severity string) Run {
		b := chairBoard(t, true)
		b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, severity))
		// THE EARLIER SITTING'S FAIL, and it sits before this sitting's register because a verdict
		// is a once-per-sitting act: recorded after it, the live PASS below is the chair's SECOND
		// verdict in one sitting and requireOncePerSitting refuses it for that instead.
		b.add("red-chair", &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)})
		openChairSitting(b)
		return b.seed()
	}
	claim := func(v recordpb.Verdict) *recordpb.Gate {
		return &recordpb.Gate{Verdict: v.Enum(), MigrationAdmittedGapIds: []string{"G1"}}
	}
	for _, c := range []struct {
		name     string
		severity string
		gate     *recordpb.Gate
	}{
		{"PASS on a board a plain PASS clears", "low", claim(recordpb.Verdict_VERDICT_PASS)},
		{"PASS over an open material gap", "medium", claim(recordpb.Verdict_VERDICT_PASS)},
		{"FAIL", "medium", claim(recordpb.Verdict_VERDICT_FAIL)},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Append(Identity{Run: board(c.severity), SeatID: "red-chair"}, c.gate)
			if err == nil || !strings.Contains(err.Error(), "migration_admitted_gap_ids") || !strings.Contains(err.Error(), "only a migration writes") {
				t.Fatalf("a live verdict carrying migration_admitted_gap_ids must be refused for it: %v", err)
			}
		})
	}
	run := board("low")
	ev, err := Append(Identity{Run: run, SeatID: "red-chair"}, proto.Clone(passGate).(*recordpb.Gate))
	if err != nil {
		t.Fatalf("control: a plain PASS on this board must append: %v", err)
	}
	if g, _ := recordpb.BodyAs[*recordpb.Gate](ev); len(g.GetMigrationAdmittedGapIds()) != 0 {
		t.Errorf("a live PASS was stamped with an admission: %v", g.GetMigrationAdmittedGapIds())
	}
}
