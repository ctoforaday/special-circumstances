package record

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY PROJECTION THAT REPORTS AN EPOCH MUST FETCH THE EVENTS THE EPOCH COMES FROM.
//
// `Clock.Advance` moves the epoch on a REGISTER event and on nothing else. A projection whose
// typed read omits REGISTER therefore leaves its clock at zero and reports `epoch: 0` for every
// row — not an error, a uniform plausible number, and one that reads as a run that never got past
// its first sitting.
//
// MEASURED TWICE, in one function each time. `show findings` reported epoch 0 for all 20 findings
// of a four-epoch run (#852), and the board fold had the same omission (fixed in #856). Both were
// found by accident while doing something else; neither was found by a test, because a projection
// that reports one wrong number for every row is internally consistent.
//
// # What this catches, and what it does not
//
// It catches a COUNT-shaped dependency: the epoch is derived from how many registers the stream
// carried, so withholding them changes the answer for every row at once. It does NOT catch a
// JOIN-shaped one — `minted_as` needs a mint whose found_by matches a finding's label, and a
// stream with the mints withheld reports the honest-looking empty list rather than a wrong number.
// Those are pinned per projection where they arise (see findingminted_test.go); this is the guard
// for the dependency that every epoch-bearing view shares.
func TestEveryProjectionThatReportsAnEpochCanSeeMoreThanOne(t *testing.T) {
	run := mustRun(t, newRun(t))
	chair := Identity{Run: run, SeatID: "red-chair"}
	lens := Identity{Run: run, SeatID: "red-lens-r1-logic"}
	app := func(id Identity, body proto.Message) {
		t.Helper()
		if _, err := Append(id, body); err != nil {
			t.Fatalf("seeding %T: %v", body, err)
		}
	}

	// TWO CHAIR SITTINGS, so the record genuinely spans two epochs. A one-epoch run would make
	// every assertion below pass over the answer `0` being correct.
	app(chair, &recordpb.Register{ToolVersion: proto.String("test")})
	app(lens, &recordpb.Register{ToolVersion: proto.String("test")})
	app(lens, &recordpb.Finding{FindingId: proto.String("f-1"), Label: proto.String("logic-F1"),
		Text: proto.String("first epoch finding"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
	app(chair, &recordpb.Mint{GapId: proto.String("R1-1"), Class: proto.String("self-attestation"),
		Problem: proto.String("p"), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
		CheckKind:  recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		FoundBy: []string{"logic-F1"}})
	app(chair, &recordpb.Log{Text: proto.String("nothing blocked me"),
		Type: recordpb.LogType_LOG_TYPE_NOMINAL.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()})
	app(chair, &recordpb.Motion{MotionId: proto.String("M1"),
		Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Basis:   proto.String("cannot settle R1-1"),
		Filing:  &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("R1-1")}}})

	// SECOND EPOCH. Everything below lands after the chair sits again, so a projection whose
	// clock is stuck reports these rows with the same number as the ones above.
	app(chair, &recordpb.Register{ToolVersion: proto.String("test")})
	app(lens, &recordpb.Finding{FindingId: proto.String("f-2"), Label: proto.String("logic-F2"),
		Text: proto.String("second epoch finding"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
	app(lens, &recordpb.Verify{Url: proto.String("https://example.org"), Claim: proto.String("c"),
		Label: proto.String("v-1"), Title: proto.String("t"),
		Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
		Text:       proto.String("the source states it plainly")})
	app(chair, &recordpb.Position{Text: proto.String("the round narrative")})

	evs, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	if got := CurrentEpochOf(evs.Events); got < 2 {
		t.Fatalf("the seeded run reaches epoch %d — this test needs a run spanning more than one, "+
			"or every assertion below passes because 0 is the right answer", got)
	}

	for _, p := range []struct {
		name string
		read func(Run) ([]byte, error)
	}{
		{"show board", BoardJSONBytes},
		{"show findings", FindingsJSONBytes},
		{"show evidence", EvidenceJSONBytes},
		{"show motions", MotionsJSONBytes},
		{"show debate", DebateJSONBytes},
		{"show log", LogJSONBytes},
	} {
		t.Run(p.name, func(t *testing.T) {
			b, err := p.read(run)
			if err != nil {
				t.Fatalf("%s: %v", p.name, err)
			}
			var tree any
			if err := json.Unmarshal(b, &tree); err != nil {
				t.Fatalf("%s emitted unparseable JSON: %v", p.name, err)
			}
			epochs := epochsIn(tree)
			if len(epochs) == 0 {
				t.Skipf("%s reports no epoch on this run — nothing to check", p.name)
			}
			// THE ASSERTION. Every row at zero on a run that reached epoch 2 is the signature of
			// a clock that never advanced, which means the typed read did not fetch REGISTER.
			allZero := true
			for _, e := range epochs {
				if e != 0 {
					allZero = false
				}
			}
			if allZero {
				t.Errorf("%s reports epoch 0 on all %d row(s) of a run that reached epoch %d — the "+
					"Clock only advances on REGISTER, so this projection's typed read is not fetching it "+
					"and every row carries the same wrong number",
					p.name, len(epochs), CurrentEpochOf(evs.Events))
			}
		})
	}
}

// epochsIn collects every `epoch` value anywhere in a decoded JSON tree, so this asks the question
// of the SHAPE rather than of a list of field paths that would go stale as views change.
func epochsIn(v any) []int {
	var out []int
	switch t := v.(type) {
	case map[string]any:
		for k, sub := range t {
			if k == "epoch" {
				if f, ok := sub.(float64); ok {
					out = append(out, int(f))
					continue
				}
			}
			out = append(out, epochsIn(sub)...)
		}
	case []any:
		for _, sub := range t {
			out = append(out, epochsIn(sub)...)
		}
	}
	return out
}
