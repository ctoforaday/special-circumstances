package record

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// NARRATION OF THE RUN IS ALWAYS MATERIAL (gblock, 2026-11-05). The report is subject prose only (#880,
// m30), so a sentence addressed to the run holds the PASS whatever it is graded. On universe-m39 the
// voice lens graded "Every number the run computed" low (finding ev386) and minted nothing, because its
// class went by grade.
func TestAReportVoiceGapIsMaterialAtAnyGrade(t *testing.T) {
	if got := recordpb.MaterialDefault(recordpb.GapClass_GAP_CLASS_REPORT_VOICE); got != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS {
		t.Fatalf("report_voice's material default = %v, want always", got)
	}
	run, lens := liveRun(t)
	appendOK(t, lens, liveMint("G1", recordpb.GapClass_GAP_CLASS_REPORT_VOICE, recordpb.Grade_GRADE_LOW))
	if g := familyGaps(t, run)["G1"]; g.Mint.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS || !g.Material {
		t.Errorf("a low report_voice gap: stamped %v, material %v; want always and material", g.Mint.GetClassMaterial(), g.Material)
	}
	if !viewMaterial(t, run)["G1"] {
		t.Error("the gap view reads a low report_voice gap as not material")
	}
}

// m39's chair withholding: a lens retired behind the head, a PASS the plan permits, and a position over
// a defect the chair's spot-check found in the retired lens's area, naming the defect's class.
func m39Withholding(t *testing.T, class *recordpb.GapClass) Run {
	t.Helper()
	b, _ := retiredBehindHead(t)
	b.register("red-chair").add(chairSeat, &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("two sentences narrate the run: \"the tests the run computed\"")})
	b.permittedPlan().add(chairSeat, &recordpb.Position{PositionId: proto.String(heardID), Class: class,
		Text: proto.String("the PASS is withheld: the report narrates the run in two sentences, and the lens whose area it is has retired")})
	return b.benchRegister(recordpb.Occasion_OCCASION_POSITION).seed()
}

// UNIVERSE-M39, REPLAYED IN SHAPE. The bench overruled the chair's position on the narration as "not
// material" (ev585) and declared it so (ev586), believing a retired lens could not be named. A position
// whose class is always material is not overruled: the class has decided materiality. The bench names
// the retired lens, the ruling readies it, and its gap holds the PASS until the sentence is fixed.
func TestM39ANarrationPositionIsUpheldNeverOverruled(t *testing.T) {
	voice := recordpb.GapClass_GAP_CLASS_REPORT_VOICE
	run := m39Withholding(t, &voice)
	bench := Identity{Run: run, SeatID: benchSeat}
	_, err := Append(bench, positionRuling(heardID, overruled, "", nil, ""))
	mustRefuse(t, err, "class report_voice", "always", "no ruling makes a defect of it immaterial", "a retired one included", "Nothing was recorded")

	fam := mustFamily(t, run)
	if slices.ContainsFunc(fam.Events, func(e *Event) bool { _, ok := recordpb.BodyAs[*recordpb.MotionRule](e); return ok }) {
		t.Fatal("the refused overrule left a ruling on the record")
	}
	if f := lensStateOf(t, run); !f.retired() {
		t.Fatalf("fixture: the lens is not retired: %+v", f)
	}

	mustAppend(t, bench, positionRuling(heardID, upheld, evLens, nil, "audit the report's voice: two sentences narrate the run"))
	if p := runPlan(t, run); p.PassPermitted || !slices.Contains(partySeats(p), evLens) {
		t.Fatalf("plan after the upheld ruling = %+v, want the retired lens readied and the PASS held", p)
	}
	lens := Identity{Run: run, SeatID: evLens}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	appendOK(t, lens, liveMint("G1", voice, recordpb.Grade_GRADE_LOW))
	if !familyGaps(t, run)["G1"].Material {
		t.Fatal("the voice lens's low narration gap is not material")
	}
	if p := runPlan(t, run); p.PassPermitted {
		t.Errorf("plan with a narration gap open = %+v, want the PASS held", p)
	}
}

// THE REFUSAL IS THE CLASS'S, NOT THE POSITION'S: a position naming a by_grade class, or none, is
// overruled as before — whether that defect matters is the bench's to weigh.
func TestAPositionOfAByGradeClassOrNoneMayBeOverruled(t *testing.T) {
	spec := recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION
	for name, class := range map[string]*recordpb.GapClass{"by_grade": &spec, "none": nil} {
		t.Run(name, func(t *testing.T) {
			run := m39Withholding(t, class)
			mustAppend(t, Identity{Run: run, SeatID: benchSeat}, positionRuling(heardID, overruled, "", nil, ""))
		})
	}
}
