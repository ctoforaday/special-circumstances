package record

import (
	"encoding/json"
	"maps"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY PROJECTION THAT REPORTS AN EPOCH CAN SEE MORE THAN ONE.
//
// Each row's epoch is the stored window the loader reads beside it (WindowIndex), so a projection
// that narrows its read to a few event families still prints the chair sitting each row is in. A
// projection that printed epochs from anything the narrowed slice must carry — a count of the
// chair's registers over the events — reports `epoch: 0` for every row when the slice does not
// carry them: not an error, a uniform plausible number, and one that reads as a run that never got
// past its first sitting.
//
// # What this catches, and what it does not
//
// It catches a COUNT-shaped dependency: an epoch derived from what the stream carried changes for
// every row at once when the read withholds it. It does NOT catch a JOIN-shaped one — a field
// computed by matching rows of two families reports an honest-looking empty value when the read
// withholds one of them, rather than a wrong number. Those are pinned per projection where they
// arise; this is the guard for the dependency that every epoch-bearing view shares.
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
	// The evidence lens mints the gap the dispatch dockets: the chair mints nothing.
	minter := Identity{Run: run, SeatID: evLens}
	app(minter, &recordpb.Register{ToolVersion: proto.String("test")})
	app(minter, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String("R1-1"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Problem: proto.String("p"), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
		CheckKind:  recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
	app(chair, &recordpb.Log{Text: proto.String("no verb renders a gap's lineage in one read"),
		Type: recordpb.LogType_LOG_TYPE_REQUEST.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()})
	// The dispatch dockets R1-1, the lens's gap, as it files a docket at impasse.
	mustFileByDispatch(t, run, &recordpb.Motion{MotionId: proto.String("M1"),
		Basis:  proto.String("cannot settle R1-1"),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("R1-1")}}})

	// SECOND EPOCH. Everything below lands after the chair sits again, so a projection whose
	// epoch is stuck reports these rows with the same number as the ones above.
	app(chair, &recordpb.Register{ToolVersion: proto.String("test")})
	app(lens, &recordpb.Verify{CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(), Url: proto.String("https://example.org"), Claim: proto.String("c"),
		Title:      proto.String("t"),
		Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
		Text:       proto.String("the source says nothing of it")})
	app(chair, &recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("the round narrative")})

	evs, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	if got := evs.At.CurrentEpoch(evs.Events); got < 2 {
		t.Fatalf("the seeded run reaches epoch %d — this test needs a run spanning more than one, "+
			"or every assertion below passes because 0 is the right answer", got)
	}

	for _, p := range []struct {
		name string
		read func(Run) ([]byte, error)
	}{
		{"show board", BoardJSONBytes},
		{"show evidence", EvidenceJSONBytes},
		{"show motions", MotionsJSONBytes},
		{"show debate", DebateJSONBytes},
		{"ops log", LogJSONBytes},
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
			// an epoch counted over a slice that did not carry what it counts.
			allZero := true
			for _, e := range epochs {
				if e != 0 {
					allZero = false
				}
			}
			if allZero {
				t.Errorf("%s reports epoch 0 on all %d row(s) of a run that reached epoch %d — the "+
					"projection is not printing each row's stored window, so every row carries the same "+
					"wrong number",
					p.name, len(epochs), evs.At.CurrentEpoch(evs.Events))
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

// A NARROWED READ CARRIES THE STORED EPOCH. The evidence view reads no act that opens a sitting, and
// each reading still reports the chair sitting the record holds it in — including a chair sitting the
// hook opened with no register, which no count over the slice could see. A loader that dropped the
// windows from the narrowed read would put every reading at 0.
func TestANarrowedReadCarriesTheStoredEpoch(t *testing.T) {
	read := func(claim string) *Event {
		return recordtest.Event(t, "red-lens-r1-L5", &recordpb.Verify{
			CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
			Url:              proto.String("https://example.org/" + claim), Claim: proto.String(claim), Title: proto.String("t"),
			Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT.Enum(),
			Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
			Text:       proto.String("the source says nothing of it"),
		})
	}
	dir := newRun(t)
	recordtest.Seed(t, dir,
		read("A"), // epoch 0: before the chair has sat
		recordtest.Event(t, "red-chair", &recordpb.Register{ToolVersion: proto.String("test")}),
		read("B"), // epoch 1: the chair's register opened it
		recordtest.Event(t, HarnessSeat, &recordpb.SittingOpen{AgentId: proto.String("chair-2"),
			AgentType: proto.String("frank-exchange-of-views:red-chair"), SeatId: proto.String("red-chair")}),
		read("C"), // epoch 2: the hook's bracket opened it, and the chair never registered
	)
	run := mustRun(t, dir)
	evs, win, err := EventsOf(run, evidenceView.families...)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, v := range EvidenceJSONOf(evs, win, "", nil).Independent {
		got[v.Claim] = v.Epoch
	}
	if want := map[string]int{"A": 0, "B": 1, "C": 2}; !maps.Equal(got, want) {
		t.Errorf("reading epochs off the narrowed read = %v, want %v (the chair sittings the record holds each in)", got, want)
	}
	out, err := EvidenceJSONBytes(run)
	if err != nil {
		t.Fatal(err)
	}
	var ej EvidenceJSON
	if err := json.Unmarshal(out, &ej); err != nil {
		t.Fatal(err)
	}
	if len(ej.Independent) != len(got) {
		t.Fatalf("show evidence prints %d reading(s); the narrowed read holds %d", len(ej.Independent), len(got))
	}
	for _, v := range ej.Independent {
		if v.Epoch != got[v.Claim] {
			t.Errorf("show evidence prints %s at epoch %d; the record holds it in %d", v.Claim, v.Epoch, got[v.Claim])
		}
	}
}
