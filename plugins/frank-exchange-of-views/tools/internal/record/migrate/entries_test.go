package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
	bodies, err := migrate.Entries()["opinion"].Translate(old)
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
	if d.GetDisposition() != recordpb.Disposition_DISPOSITION_REMANDED {
		t.Errorf("disposition: %v", d.GetDisposition())
	}
	if d.GetPrinciple() != "correctness outranks economy" || d.GetDirection() != "blue reports the reproduction" || d.ReopensOn != nil {
		t.Errorf("ruling fields did not carry, or a remand's direction stayed a reopen condition: %+v", d)
	}
	if rule.GetOpinion() != "the computation is unproved" {
		t.Errorf("the old rationale is the ruling's prose (MotionRule.opinion): %q", rule.GetOpinion())
	}
	if d.GetFinal() {
		t.Errorf("final=0 must stay absent/false — the direction is the answer this remand gave")
	}
}

// A SPOT-CHECK MIGRATES WITHOUT ITS EMPTINESS CLAIM. `none` was the seat's word that the archive was
// empty, which the board answers; the ids, the areas and the reason carry as written.
func TestASpotCheckMigratesWithoutItsEmptinessClaim(t *testing.T) {
	for _, none := range []int64{0, 1} {
		old := migrate.OldEvent{ID: 5, SeatID: "red-chair", TS: ts1, Word: "spot_check",
			Fields: map[string]any{"none": none, "reason": "the archive was empty"},
			Lists:  map[string][]string{"areas": {"red-lens-logic"}}}
		bodies, err := migrate.Entries().Translate(old)
		if err != nil {
			t.Fatalf("none=%d: %v", none, err)
		}
		sc, ok := bodies[0].(*recordpb.SpotCheck)
		if !ok || sc.GetReason() != "the archive was empty" || len(sc.GetAreas()) != 1 {
			t.Errorf("none=%d: the spot-check did not carry its reason and areas: %v", none, bodies)
		}
	}
}

// A REMAND'S REOPENS_ON IS ITS DIRECTION (#1253). A source states a remand's research direction in
// the one column it has for what follows a ruling, `reopens_on`; the migrated ruling carries it as
// the direction, and a ruling that closes the gap keeps it as what would reopen it. Both eras that
// write a docket ruling are held to it: an opinion, and a motion_rule with a docket arm.
func TestARemandsReopensOnMigratesAsItsDirection(t *testing.T) {
	ruling := func(disposition string) map[string]any {
		return map[string]any{"disposition": disposition, "principle": "p", "tension": "t",
			"review_flag": "r", "settled": "s", "reopens_on": "the stated text"}
	}
	for _, c := range []struct {
		name, disposition, wantDirection, wantReopensOn string
	}{
		{"a remand", "remanded", "the stated text", ""},
		{"a ruling that closes the gap", "not_a_defect", "", "the stated text"},
	} {
		for _, era := range []struct {
			name string
			old  migrate.OldEvent
			word string
		}{
			{"opinion", migrate.OldEvent{ID: 7, SeatID: "judge", TS: ts1, Word: "opinion",
				Fields: func() map[string]any {
					f := ruling(c.disposition)
					f["gap_id"] = "G1"
					f["rationale"] = "why"
					return f
				}()}, "opinion"},
			{"motion_rule", migrate.OldEvent{ID: 8, SeatID: "judge", TS: ts1, Word: "motion_rule",
				Fields: map[string]any{"motion_id": "M1", "opinion": "why"},
				Arms:   map[string]map[string]any{"docket": ruling(c.disposition)}}, "motion_rule"},
		} {
			t.Run(c.name+"/"+era.name, func(t *testing.T) {
				bodies, err := migrate.Entries().Translate(era.old)
				if err != nil {
					t.Fatal(err)
				}
				var d *recordpb.DocketRuling
				for _, b := range bodies {
					if r, ok := b.(*recordpb.MotionRule); ok {
						d = r.GetDocket()
					}
				}
				if d == nil {
					t.Fatalf("no docket ruling among %v", bodies)
				}
				if d.GetDirection() != c.wantDirection || d.GetReopensOn() != c.wantReopensOn {
					t.Errorf("direction %q, reopens_on %q; want %q, %q", d.GetDirection(), d.GetReopensOn(), c.wantDirection, c.wantReopensOn)
				}
			})
		}
	}
}

// An opinion field this entry does not place refuses by name — the same contract the
// identity default holds kept words to.
func TestOpinionEntryRefusesAColumnItCannotPlace(t *testing.T) {
	old := migrate.OldEvent{
		ID: 7, SeatID: "bench", Round: 2, TS: ts1, Word: "opinion",
		Fields: map[string]any{"gap_id": "R1-1", "disposition": "carried", "gavel_weight": "heavy"},
	}
	if _, err := migrate.Entries()["opinion"].Translate(old); err == nil || !strings.Contains(err.Error(), "opinion.gavel_weight") {
		t.Fatalf("want a refusal naming opinion.gavel_weight, got %v", err)
	}
}

// The estoppel pairing is structural in the new model, and the translation says it too:
// the tool recorded its own refusal, so the entry is (TOOL, ESTOPPEL), never a seat filing. The
// entry's text, which names the prescribing gap, is what carries; estopped_by is not.
func TestFrictionEstoppelBecomesToolSourced(t *testing.T) {
	old := migrate.OldEvent{
		ID: 3, SeatID: "blue-lane-1", Round: 1, TS: ts1, Word: "friction",
		Fields: map[string]any{"text": "mint refused", "kind": "estoppel", "estopped_by": "R1-2"},
	}
	bodies, err := migrate.Entries()["friction"].Translate(old)
	if err != nil {
		t.Fatal(err)
	}
	l := bodies[0].(*recordpb.Log)
	if l.GetType() != recordpb.LogType_LOG_TYPE_ESTOPPEL || l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL {
		t.Fatalf("an estoppel is the tool's own act: got type %v source %v", l.GetType(), l.GetSource())
	}
	if l.GetText() != "mint refused" {
		t.Errorf("the entry's text did not carry: %q", l.GetText())
	}
}

// Identity translation handles the full decomposition: scalar fields, a oneof ARM table,
// and the `_case` discriminator — which is DERIVED, so it is skipped only where the oneof
// exists to derive it from. The motion's entry drops the stored subject, which the arm states.
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
	bodies, err := migrate.Entries().Translate(old)
	if err != nil {
		t.Fatal(err)
	}
	m := bodies[0].(*recordpb.Motion)
	if m.Subject() != recordpb.MotionSubject_MOTION_SUBJECT_DOCKET {
		t.Errorf("subject word did not convert: %v", m.Subject())
	}
	if m.GetDocket().GetGapId() != "R1-1" {
		t.Errorf("the arm table did not rebuild the oneof: %+v", m)
	}
}

// A CITE FROM BEFORE EPOCH 7 SAYS ITS ORIGIN WAS NEVER RECORDED. A guessed `embedded` would tell
// red an OCR quote needs no page check; the fill states the question was never asked, and a cite
// that carries an origin keeps it.
func TestAnOldCiteStatesItsOriginWasNotRecorded(t *testing.T) {
	for _, tc := range []struct {
		fields map[string]any
		want   recordpb.SourceTextOrigin
	}{
		{map[string]any{"label": "c-1", "url": "https://x"}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NOT_RECORDED},
		{map[string]any{"label": "c-2", "url": "https://x", "source_text_origin": "ocr"}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR},
	} {
		bodies, err := migrate.Entries()["cite"].Translate(migrate.OldEvent{ID: 1, SeatID: "blue-respond", TS: ts1, Word: "cite", Fields: tc.fields})
		if err != nil {
			t.Fatal(err)
		}
		c, ok := bodies[0].(*recordpb.Cite)
		if !ok || len(bodies) != 1 {
			t.Fatalf("a cite translated to %d bodies, the first %T", len(bodies), bodies[0])
		}
		if c.GetSourceTextOrigin() != tc.want {
			t.Errorf("%v: origin = %v, want %v", tc.fields, c.GetSourceTextOrigin(), tc.want)
		}
	}
}

// AND A CITE FROM BEFORE EPOCH 10 SAYS NOBODY ASKED WHETHER THE WORK STILL STANDS. `standing`
// would be a reassurance no index gave, printed against a paper that may have been withdrawn
// years before the run; `not_checked` is the right word for a live cite with no doi, and wrong
// here, where the field did not exist. A cite that carries a status keeps it.
func TestAnOldCiteStatesItsWorkStatusWasNotRecorded(t *testing.T) {
	for _, tc := range []struct {
		fields map[string]any
		want   recordpb.WorkStatus
	}{
		{map[string]any{"label": "c-1", "url": "https://x"}, recordpb.WorkStatus_WORK_STATUS_NOT_RECORDED},
		{map[string]any{"label": "c-2", "url": "https://x", "work_status": "retracted"}, recordpb.WorkStatus_WORK_STATUS_RETRACTED},
	} {
		bodies, err := migrate.Entries()["cite"].Translate(migrate.OldEvent{ID: 1, SeatID: "blue-respond", TS: ts1, Word: "cite", Fields: tc.fields})
		if err != nil {
			t.Fatal(err)
		}
		c, ok := bodies[0].(*recordpb.Cite)
		if !ok || len(bodies) != 1 {
			t.Fatalf("a cite translated to %d bodies, the first %T", len(bodies), bodies[0])
		}
		if c.GetWorkStatus() != tc.want {
			t.Errorf("%v: work_status = %v, want %v", tc.fields, c.GetWorkStatus(), tc.want)
		}
	}
}

// AND A CITE FROM BEFORE EPOCH 15 SAYS NOBODY RECORDED WHICH PART OF THE WORK ITS COPY WAS.
// `not_asked` would claim the fetch looked and could not tell; the field did not exist. A cite
// that carries one keeps it.
func TestAnOldCiteStatesItsCompletenessWasNotRecorded(t *testing.T) {
	for _, tc := range []struct {
		fields map[string]any
		want   recordpb.SourceCompleteness
	}{
		{map[string]any{"label": "c-1", "url": "https://x"}, recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_RECORDED},
		{map[string]any{"label": "c-2", "url": "https://x", "source_completeness": "abstract"}, recordpb.SourceCompleteness_SOURCE_COMPLETENESS_ABSTRACT},
	} {
		bodies, err := migrate.Entries()["cite"].Translate(migrate.OldEvent{ID: 1, SeatID: "blue-respond", TS: ts1, Word: "cite", Fields: tc.fields})
		if err != nil {
			t.Fatal(err)
		}
		c, ok := bodies[0].(*recordpb.Cite)
		if !ok || len(bodies) != 1 {
			t.Fatalf("a cite translated to %d bodies, the first %T", len(bodies), bodies[0])
		}
		if c.GetSourceCompleteness() != tc.want {
			t.Errorf("%v: source_completeness = %v, want %v", tc.fields, c.GetSourceCompleteness(), tc.want)
		}
	}
}

// AND A CITE FROM BEFORE EPOCH 40 SAYS NOBODY RECORDED WHOSE BYTES IT READ. `none` would claim the
// typed url's own; the fetch index that knew whether a copy stood in is not archived. No copy url is
// written — that is the residue. A cite that carries a substitution keeps it.
func TestAnOldCiteStatesItsCopySubstitutionWasNotRecorded(t *testing.T) {
	for _, tc := range []struct {
		fields map[string]any
		want   recordpb.CopySubstitution
	}{
		{map[string]any{"label": "c-1", "url": "https://x"}, recordpb.CopySubstitution_COPY_SUBSTITUTION_NOT_RECORDED},
		{map[string]any{"label": "c-2", "url": "https://x", "copy_substitution": "recovered", "copy_url": "https://y"}, recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED},
	} {
		bodies, err := migrate.Entries()["cite"].Translate(migrate.OldEvent{ID: 1, SeatID: "blue-respond", TS: ts1, Word: "cite", Fields: tc.fields})
		if err != nil {
			t.Fatal(err)
		}
		c, ok := bodies[0].(*recordpb.Cite)
		if !ok || len(bodies) != 1 {
			t.Fatalf("a cite translated to %d bodies, the first %T", len(bodies), bodies[0])
		}
		if c.GetCopySubstitution() != tc.want {
			t.Errorf("%v: copy_substitution = %v, want %v", tc.fields, c.GetCopySubstitution(), tc.want)
		}
		if tc.want == recordpb.CopySubstitution_COPY_SUBSTITUTION_NOT_RECORDED && c.CopyUrl != nil {
			t.Errorf("a pre-field cite was given a copy url it never recorded: %q", c.GetCopyUrl())
		}
	}
}

// A CITE FROM BEFORE EPOCH 37 named an OCR citation's span `ocr_quote`: it is the citation's
// source_span, word for word, beside the reading it was found in. No archived cite holds the span's
// offsets or the source's text around it — the cache is not archived — so none is invented.
func TestAnOldOCRQuoteIsTheCitationsSourceSpan(t *testing.T) {
	bodies, err := migrate.Entries()["cite"].Translate(migrate.OldEvent{ID: 1, SeatID: "blue-respond", TS: ts1, Word: "cite", Fields: map[string]any{
		"label": "c-1", "url": "https://x", "source_text_origin": "ocr", "ocr_quote": "Annexes A through J",
		"ocr_engine": "e", "ocr_text_sha": "t"}})
	if err != nil {
		t.Fatal(err)
	}
	c := bodies[0].(*recordpb.Cite)
	if c.GetSourceSpan() != "Annexes A through J" || c.GetOcrEngine() != "e" || c.SpanStart != nil || c.SourceMatched != nil {
		t.Errorf("translated cite: span %q, engine %q, offsets %v, matched %v", c.GetSourceSpan(), c.GetOcrEngine(), c.SpanStart, c.SourceMatched)
	}
}
