package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// The one CONCEPT translation, stated as data: an opinion's fields in, the docket pair out.
// The bench raised the question and the bench answered it — one old event, two new ones,
// sharing the opinion's identity.
func TestOpinionBecomesTheDocketPair(t *testing.T) {
	old := migrate.OldEvent{
		ID: 7, SeatID: "bench", Round: 2, TS: ts1, Word: "opinion",
		Fields: map[string]any{
			"gap_id":      "R1-1",
			"disposition": "carried",
			"principle":   "correctness outranks economy",
			"tension":     "",
			"review_flag": "",
			"rationale":   "the computation is unproved",
			"settled":     "",
			"reopens_on":  "blue reports the reproduction",
		},
	}
	dst := runtest.New(t, recordtest.TmpRun(t))
	bodies, err := migrate.Entries()["opinion"].Translate(old, dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 2 {
		t.Fatalf("an opinion is a motion AND its ruling: got %d bodies", len(bodies))
	}
	motion, ok := bodies[0].(*recordpb.Motion)
	if !ok {
		t.Fatalf("first body is %T, want Motion", bodies[0])
	}
	rule, ok := bodies[1].(*recordpb.MotionRule)
	if !ok {
		t.Fatalf("second body is %T, want MotionRule", bodies[1])
	}
	if motion.GetMotionId() == "" || motion.GetMotionId() != rule.GetMotionId() {
		t.Fatalf("the pair must share a minted id: motion %q rule %q", motion.GetMotionId(), rule.GetMotionId())
	}
	if motion.GetDocket().GetGapId() != "R1-1" {
		t.Errorf("the motion does not name the opinion's gap: %q", motion.GetDocket().GetGapId())
	}
	if !strings.Contains(motion.GetBasis(), "migrated") {
		t.Errorf("a synthesized filing must say it was synthesized, not pose as a filing: %q", motion.GetBasis())
	}
	d := rule.GetDocket()
	if d.GetDisposition() != recordpb.Disposition_DISPOSITION_CARRIED {
		t.Errorf("disposition: %v", d.GetDisposition())
	}
	if d.GetPrinciple() != "correctness outranks economy" || d.GetReopensOn() != "blue reports the reproduction" {
		t.Errorf("ruling fields did not carry: %+v", d)
	}
	if rule.GetOpinion() != "the computation is unproved" {
		t.Errorf("the old rationale is the ruling's prose (MotionRule.opinion): %q", rule.GetOpinion())
	}
	if d.GetFinal() {
		t.Errorf("final=0 must stay absent/false — reopens_on is the answer this ruling gave")
	}
}

// An opinion field this entry does not place refuses by name — the same contract the
// identity default holds kept words to.
func TestOpinionEntryRefusesAColumnItCannotPlace(t *testing.T) {
	old := migrate.OldEvent{
		ID: 7, SeatID: "bench", Round: 2, TS: ts1, Word: "opinion",
		Fields: map[string]any{"gap_id": "R1-1", "disposition": "carried", "gavel_weight": "heavy"},
	}
	dst := runtest.New(t, recordtest.TmpRun(t))
	if _, err := migrate.Entries()["opinion"].Translate(old, dst); err == nil || !strings.Contains(err.Error(), "opinion.gavel_weight") {
		t.Fatalf("want a refusal naming opinion.gavel_weight, got %v", err)
	}
}

// The estoppel pairing is structural in the new model, and the translation says it too:
// the tool recorded its own refusal, so the entry is (TOOL, ESTOPPEL), never a seat filing.
func TestFrictionEstoppelBecomesToolSourced(t *testing.T) {
	old := migrate.OldEvent{
		ID: 3, SeatID: "blue-lane-1", Round: 1, TS: ts1, Word: "friction",
		Fields: map[string]any{"text": "mint refused", "kind": "estoppel", "estopped_by": "R1-2"},
	}
	dst := runtest.New(t, recordtest.TmpRun(t))
	bodies, err := migrate.Entries()["friction"].Translate(old, dst)
	if err != nil {
		t.Fatal(err)
	}
	l := bodies[0].(*recordpb.Log)
	if l.GetType() != recordpb.LogType_LOG_TYPE_ESTOPPEL || l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL {
		t.Fatalf("an estoppel is the tool's own act: got type %v source %v", l.GetType(), l.GetSource())
	}
	if l.GetEstoppedBy() != "R1-2" {
		t.Errorf("estopped_by did not carry: %q", l.GetEstoppedBy())
	}
}

// Identity translation handles the full decomposition: scalar fields, a oneof ARM table,
// and the `_case` discriminator — which is DERIVED, so it is skipped only where the oneof
// exists to derive it from.
func TestIdentityRebuildsAOneofArm(t *testing.T) {
	old := migrate.OldEvent{
		ID: 9, SeatID: "red-merge-r1", Round: 1, TS: ts1, Word: "motion",
		Fields: map[string]any{
			"motion_id":   "M1",
			"subject":     "docket",
			"basis":       "put R1-1 before the bench",
			"filing_case": "docket",
		},
		Arms: map[string]map[string]any{"docket": {"gap_id": "R1-1"}},
	}
	dst := runtest.New(t, recordtest.TmpRun(t))
	bodies, err := (migrate.Registry{}).Translate(old, dst)
	if err != nil {
		t.Fatal(err)
	}
	m := bodies[0].(*recordpb.Motion)
	if m.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DOCKET {
		t.Errorf("subject word did not convert: %v", m.GetSubject())
	}
	if m.GetDocket().GetGapId() != "R1-1" {
		t.Errorf("the arm table did not rebuild the oneof: %+v", m)
	}
}
