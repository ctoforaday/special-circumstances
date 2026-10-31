package record

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE READ IS THE RECORD'S, AT EACH POINT THE RUN REACHES: a filing is owed until a sitting of its
// occasion opens after it, and a filing that sitting left unruled is not owed again, so nothing
// re-enters a hearing. A halt and an outcome read once they stand, and each prefix reads as it read
// then, which is what capture's audit recomputes.
func TestTheReadIsTheRecordAtEachEvent(t *testing.T) {
	type want struct {
		owed    []BenchHearing
		halted  bool
		outcome RecordedVerdict
	}
	b := permittingBoard(t)
	marks := []struct {
		name string
		at   int
		want want
	}{{name: "nothing filed", at: len(b.evs), want: want{owed: []BenchHearing{}}}}
	mark := func(name string, w want) {
		marks = append(marks, struct {
			name string
			at   int
			want want
		}{name, len(b.evs), w})
	}
	b.add("blue-respond", petitionMotion("M1"))
	mark("a petition filed", want{owed: []BenchHearing{"petition"}})
	b.withhold("A-00000001")
	mark("a position beside it, heard in the engine's order", want{owed: []BenchHearing{"petition", "position"}})
	b.add("judge", &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_PETITION.Enum()})
	mark("the petition sitting opened, and the position still owed", want{owed: []BenchHearing{"position"}})
	b.add("judge", &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_POSITION.Enum()})
	mark("left unruled by both, and neither re-enters", want{owed: []BenchHearing{}})
	b.add("judge", &recordpb.Halt{Opinion: proto.String("consent gate")})
	mark("a halt", want{owed: []BenchHearing{}, halted: true})
	b.add("judge", &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_HALTED.Enum(), Prose: proto.String("halted")})
	mark("an outcome", want{owed: []BenchHearing{}, halted: true, outcome: "HALTED"})
	run := b.seed()

	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	ids := fam.At.IDs(fam.Events)
	for _, m := range marks {
		got, err := RecordReadAt(fam.Events, fam.At, ids[m.at-1])
		if err != nil {
			t.Fatalf("%s: %v", m.name, err)
		}
		if got.Event != ids[m.at-1] || !slices.Equal(got.BenchOwed, m.want.owed) || got.Halted != m.want.halted || got.Outcome != m.want.outcome {
			t.Errorf("%s: read %+v, want %+v at event %d", m.name, got, m.want, ids[m.at-1])
		}
	}
	now, err := ReadRecord(run)
	if err != nil {
		t.Fatal(err)
	}
	if now.Event != ids[len(ids)-1] || !now.Halted || now.Outcome != "HALTED" {
		t.Errorf("the read as it stands = %+v, want the last event's", now)
	}
	if _, err := RecordReadAt(fam.Events, fam.At, ids[len(ids)-1]+1); err == nil {
		t.Error("a read naming an event the stream does not hold was recomputed")
	}
}

// A READ NAMES AN EVENT: a record holds none below 1, and the floor on RecordRead.Event refuses what a
// seat types there. A stream's event 0 is no prefix to fold, and a record with no event has no read.
func TestEveryReadNamesAnEventOnTheRecord(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair").ingest().seed()
	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	if first := fam.At.IDs(fam.Events)[0]; first < 1 {
		t.Errorf("the record's first event is %d", first)
	}
	if _, err := RecordReadAt(fam.Events, fam.At, 0); err == nil {
		t.Error("a read naming event 0 was recomputed")
	}
	if _, err := ReadRecord(newStage(t).seed()); err == nil || !strings.Contains(err.Error(), "holds no event") {
		t.Errorf("a record with no event read as %v", err)
	}
	var floor *int
	for _, f := range RecordReadFields() {
		if f.JSON == "event" {
			floor = f.Min
		}
	}
	if floor == nil || *floor != 1 {
		t.Error("RecordRead.Event's floor is not 1")
	}
}

// EMPTY AND ABSENT STAY DISTINCT ON THE WIRE: an owed list with nothing in it prints as an empty list,
// which a seat's envelope reads as a read that found nothing, and an outcome not yet stamped is no key.
func TestAReadOfNothingPrintsAnEmptyListAndNoOutcome(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().seed()
	read, err := ReadRecord(run)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(read)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	if string(doc["bench_owed"]) != "[]" {
		t.Errorf("bench_owed = %s, want an empty list", doc["bench_owed"])
	}
	if _, carried := doc["outcome"]; carried {
		t.Errorf("a record with no outcome printed one: %s", b)
	}
}
