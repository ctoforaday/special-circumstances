package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// liveRun is a run with a lens registered to mint through the real write path — the path that
// stamps class_material from the class's material default.
func liveRun(t *testing.T) (Run, Identity) {
	t.Helper()
	dir := recordtest.TmpRun(t)
	writeRunConfig(t, dir, `{"mintBudget":50}`)
	run := mustRun(t, dir)
	lens := Identity{Run: run, SeatID: evLens}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	return run, lens
}

func liveMint(id string, class recordpb.GapClass, sev recordpb.Grade) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(id), Class: class.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
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

// familyGaps is the family FamilyOf assembles, per gap.
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
// `always` class is material.
func TestMaterialAlwaysHoldsLowGrade(t *testing.T) {
	run, lens := liveRun(t)
	appendOK(t, lens, liveMint("G1", recordpb.GapClass_GAP_CLASS_FALSE_UNIVERSAL, recordpb.Grade_GRADE_LOW))
	g := familyGaps(t, run)["G1"]
	if g.Mint.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS {
		t.Errorf("the mint was stamped %v, want false_universal's always", g.Mint.GetClassMaterial())
	}
	if !viewMaterial(t, run)["G1"] {
		t.Error("a low gap of an always class must be material, and the gap view says it is not")
	}
}

// A template-shape defect changes no conclusion, so even a HIGH gap of a `never` class is not.
func TestMaterialNeverReleasesHighGrade(t *testing.T) {
	run, lens := liveRun(t)
	appendOK(t, lens, liveMint("G1", recordpb.GapClass_GAP_CLASS_STRUCTURE_NONCOMPLIANCE, recordpb.Grade_GRADE_HIGH))
	if viewMaterial(t, run)["G1"] {
		t.Error("a high gap of a never class must not be material, and the gap view says it is")
	}
}

// THE GAP VIEW STATES MATERIALITY, gap by gap, over every class default and the grades either side
// of the by-grade floor — and on a regrade, because materiality reads the grade NOW.
func TestTheGapViewStatesMaterialAcrossTheFloor(t *testing.T) {
	run, lens := liveRun(t)
	a, nv, g := recordpb.GapClass_GAP_CLASS_FALSE_UNIVERSAL, recordpb.GapClass_GAP_CLASS_STRUCTURE_NONCOMPLIANCE, recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION
	want := map[string]bool{}
	n := 0
	for _, c := range []struct {
		class recordpb.GapClass
		sev   recordpb.Grade
		want  bool
	}{
		{a, recordpb.Grade_GRADE_TRIVIAL, true}, {a, recordpb.Grade_GRADE_MEDIUM, true}, {a, recordpb.Grade_GRADE_HIGH, true},
		{nv, recordpb.Grade_GRADE_TRIVIAL, false}, {nv, recordpb.Grade_GRADE_MEDIUM, false}, {nv, recordpb.Grade_GRADE_CERTAIN, false},
		{g, recordpb.Grade_GRADE_LOW_MEDIUM, false}, {g, recordpb.Grade_GRADE_MEDIUM, true}, {g, recordpb.Grade_GRADE_HIGH, true},
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

	view := viewMaterial(t, run)
	for id, w := range want {
		if view[id] != w {
			t.Errorf("%s: material=%v, want %v", id, view[id], w)
		}
	}
}

// NO SEAT SETS IT. A lens that could choose its gap's materiality could mint every gap harmless;
// a migration's source is the record, so an archived value is kept as recorded.
func TestMintRefusesPresetClassMaterial(t *testing.T) {
	run, lens := liveRun(t)
	m := liveMint("G1", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_HIGH)
	m.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_NEVER.Enum()
	if _, err := Append(lens, m); err == nil || !strings.Contains(err.Error(), "stamped by the tool") {
		t.Fatalf("a mint arriving with class_material set landed outside a migration: %v", err)
	}
	Migrating = true
	defer func() { Migrating = false }()
	m2 := liveMint("G2", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_HIGH)
	m2.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS.Enum()
	appendOK(t, lens, m2)
	if got := familyGaps(t, run)["G2"].Mint.GetClassMaterial(); got != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS {
		t.Errorf("a migrated mint's recorded class_material became %v, want it kept as recorded (always)", got)
	}
}

// The refusal a lens meets when it mints ungraded names what severity now decides.
func TestSeverityRefusalWhyIsClassAware(t *testing.T) {
	_, lens := liveRun(t)
	m := liveMint("G1", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_LOW)
	m.Severity = nil
	_, err := Append(lens, m)
	if err == nil || !strings.Contains(err.Error(), "--severity") ||
		!strings.Contains(err.Error(), "materiality reads it wherever the class goes by grade") {
		t.Fatalf("the severity refusal must say materiality reads it where the class goes by grade: %v", err)
	}
}

// A PRE-CHANGE RUN — its database written before the field — is refused at the first write, naming
// what the database lacks and `migrate`.
func TestPreChangeRunRefusedNamingMigrate(t *testing.T) {
	dir := recordtest.ExtractArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	run, err := NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Append(Identity{Run: run, SeatID: evLens}, liveMint("G99", recordpb.GapClass_GAP_CLASS_FALSE_UNIVERSAL, recordpb.Grade_GRADE_LOW))
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
