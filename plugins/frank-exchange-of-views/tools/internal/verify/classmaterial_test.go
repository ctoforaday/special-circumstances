package verify

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// passOverOneOpenGap is a seeded record — through the family fold that computes Gap.Material, not a
// hand-set flag — holding one open gap of the given class and grade under a PASS.
func passOverOneOpenGap(t *testing.T, cm recordpb.ClassMaterial, sev recordpb.Grade) record.Family {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.Event(t, "red-lens-evidence", &recordpb.Register{}),
		recordtest.Event(t, "red-lens-evidence", &recordpb.Mint{GapId: proto.String("G1"), Class: proto.String("x"),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Severity: sev.Enum(), Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum(),
			ClassMaterial: cm.Enum()}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()}),
	)
	fam, err := record.FamilyOf(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	return fam
}

// A PASS over an open gap whose class is never material is not the #67 violation, however it is
// graded; a PASS over a low gap of an always-material class is.
func TestPassOverOpenNeverClassGapIsNotA67Violation(t *testing.T) {
	c := find(t, Run(passOverOneOpenGap(t, recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, recordpb.Grade_GRADE_HIGH)), "pass-closes-all-gaps")
	if !c.OK || c.NA {
		t.Errorf("a PASS over an open never-class gap graded high must hold the check: %+v", c)
	}
	c = find(t, Run(passOverOneOpenGap(t, recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.Grade_GRADE_LOW)), "pass-closes-all-gaps")
	if c.OK {
		t.Errorf("a PASS over an open always-class gap graded low must be the violation: %+v", c)
	}
}

// A PASS OVER AN OPEN MATERIAL GAP WITHOUT THE ADMISSION STILL FAILS (fork (a)), and names no
// admitted gap; the same board whose PASS carries the admission for that gap holds the check and
// names it as admitted by migration. An admission naming a different gap admits nothing.
func TestPassOverOpenMaterialGapWithoutAdmissionFails(t *testing.T) {
	check := func(admitted ...string) Check {
		dir := recordtest.TmpRun(t)
		recordtest.Seed(t, dir,
			recordtest.Event(t, "red-lens-evidence", &recordpb.Register{}),
			recordtest.Event(t, "red-lens-evidence", &recordpb.Mint{GapId: proto.String("G1"), Class: proto.String("x"),
				Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
				Severity: recordpb.Grade_GRADE_LOW.Enum(), Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum(),
				ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS.Enum()}),
			recordtest.Event(t, "red-chair", &recordpb.Register{}),
			recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum(), MigrationAdmittedGapIds: admitted}),
		)
		fam, err := record.FamilyOf(runtest.Open(t, dir))
		if err != nil {
			t.Fatal(err)
		}
		return find(t, Run(fam), "pass-closes-all-gaps")
	}
	for _, c := range []struct {
		name     string
		admitted []string
	}{{"no admission", nil}, {"an admission naming another gap", []string{"G9"}}} {
		got := check(c.admitted...)
		if got.OK || got.NA || len(got.Violations) != 1 || got.Violations[0] != "G1" || len(got.Admitted) != 0 {
			t.Errorf("%s: a PASS over open always-class G1 must FAIL naming G1 and admit nothing: %+v", c.name, got)
		}
	}
	got := check("G1")
	if !got.OK || got.NA || len(got.Violations) != 0 || len(got.Admitted) != 1 || got.Admitted[0] != "G1" ||
		!strings.Contains(got.Detail, "admitted by migration") {
		t.Errorf("a PASS carrying the admission for G1 must hold the check and name G1 as admitted by migration: %+v", got)
	}
}
