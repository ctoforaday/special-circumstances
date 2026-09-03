package recordsql

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A MIXED RECORD COMES BACK WHOLE — every recovery path at once, not one per test.
//
// The single-event round-trip tests each pin one mechanism in isolation: scalars, presence, a
// list. Reading is about to be restructured from per-event queries to per-table scans, and the
// shape that restructuring can break is exactly the one no isolated test holds: several events of
// the SAME type interleaved with other types, where a batched read has to hand each row, each list
// in its ord order, and each oneof arm back to the event that wrote it rather than to a neighbour.
func TestAMixedRecordSurvivesTheRoundTrip(t *testing.T) {
	db := store(t)

	mint := func(id string, supersedes ...string) *recordpb.Mint {
		return &recordpb.Mint{
			GapId:           proto.String(id),
			Class:           proto.String("scope-creep"),
			Problem:         proto.String("a problem for " + id),
			AcceptanceCheck: proto.String("a check for " + id),
			CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
			Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
			Supersedes:      supersedes,
		}
	}
	originals := []proto.Message{
		mint("R1-1", "R0-4", "R0-9"),
		&recordpb.Position{Text: proto.String("red speaks")},
		// A grade motion: the oneof arm WITH scalar columns, referencing the mint above.
		&recordpb.Motion{
			MotionId: proto.String("M-1"),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
			Basis:    proto.String("severity is understated"),
			Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
				GapId:     proto.String("R1-1"),
				Dimension: recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY.Enum(),
				Proposed:  recordpb.Grade_GRADE_HIGH.Enum(),
			}},
		},
		mint("R1-2"),
		// A petition: the oneof arm whose row holds NO scalar values — the arm must still come
		// back SET, because "filed a petition" and "filed nothing" are different acts.
		&recordpb.Motion{
			MotionId: proto.String("M-2"),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_PETITION.Enum(),
			Filing:   &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{}},
		},
		// A motion with NO filing arm at all: the NULL discriminator path.
		&recordpb.Motion{
			MotionId: proto.String("M-3"),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
			Basis:    proto.String("filed bare"),
		},
	}
	for i, body := range originals {
		ev := &recordpb.Event{
			SeatId: proto.String("red-merge-r1"),
			Round:  proto.Int32(1),
			Ts:     proto.String("2026-01-01T00:00:00Z"),
			Key:    proto.String(fmt.Sprintf("red-merge-r1:act:#%d", i)),
		}
		typ, err := recordpb.SetBody(ev, body)
		if err != nil {
			t.Fatal(err)
		}
		ev.Type = &typ
		if _, err := Insert(db, ev); err != nil {
			t.Fatalf("event %d: %v", i, err)
		}
	}

	evs, err := Events(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != len(originals) {
		t.Fatalf("read %d events, want %d", len(evs), len(originals))
	}
	for i, want := range originals {
		got, ok := recordpb.Body(evs[i])
		if !ok {
			t.Fatalf("event %d came back with no body", i)
		}
		if !proto.Equal(want, got) {
			t.Errorf("event %d did not survive the round trip:\n before %v\n after  %v", i, want, got)
		}
	}
	// Stated separately, because proto.Equal cannot say WHICH mechanism dropped the arm.
	if m, ok := recordpb.BodyAs[*recordpb.Motion](evs[2]); !ok || m.GetGrade() == nil {
		t.Error("the grade motion's filing arm did not come home")
	}
	if m, ok := recordpb.BodyAs[*recordpb.Motion](evs[4]); !ok || m.GetPetition() == nil {
		t.Error("the empty petition arm came back unset — an empty row is still a filing")
	}
	if m, ok := recordpb.BodyAs[*recordpb.Motion](evs[5]); !ok || m.GetFiling() != nil {
		t.Error("the bare motion grew a filing arm it never had")
	}
}

// AN EVENT ROW WITH NO BODY ROW IS A REFUSAL, NOT AN EMPTY REPLAY.
//
// Insert writes both in one transaction, so this shape needs hand-written SQL to exist at all —
// which is the point: if the record disagrees with itself, the read must say so rather than replay
// a seat that did nothing.
func TestAnEventRowWithNoBodyRowIsRefused(t *testing.T) {
	db := store(t)
	if _, err := db.Exec(`INSERT INTO "events" ("seat_id", "round", "ts", "type", "key")
		VALUES ('red-merge-r1', 1, '2026-01-01T00:00:00Z', 'mint', 'red-merge-r1:act:#0')`); err != nil {
		t.Fatal(err)
	}
	_, err := Events(db)
	if err == nil || !strings.Contains(err.Error(), "has no mint body") {
		t.Fatalf("err = %v, want the missing-body refusal", err)
	}
}

// The read path's price, on a record shaped like a real run: scalars, lists and oneof arms mixed.
// Run with: go test ./internal/record/recordsql -bench Events -benchmem -run XXX
func BenchmarkEvents(b *testing.B) {
	path := filepath.Join(b.TempDir(), "record.db")
	db, err := Open(path)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = Close(path) })
	for i := 0; i < 400; i++ {
		gap := fmt.Sprintf("R1-%d", i+1)
		bodies := []proto.Message{
			&recordpb.Mint{
				GapId:           proto.String(gap),
				Class:           proto.String("scope-creep"),
				Problem:         proto.String("a problem"),
				AcceptanceCheck: proto.String("a check"),
				CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
				Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
				Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
				Supersedes:      []string{"R0-1", "R0-2"},
			},
			&recordpb.Position{Text: proto.String("a position")},
			&recordpb.Motion{
				MotionId: proto.String(fmt.Sprintf("M-%d", i+1)),
				Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
				Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
					GapId:     proto.String(gap),
					Dimension: recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY.Enum(),
					Proposed:  recordpb.Grade_GRADE_HIGH.Enum(),
				}},
			},
		}
		for j, body := range bodies {
			ev := &recordpb.Event{
				SeatId: proto.String("red-merge-r1"),
				Round:  proto.Int32(1),
				Ts:     proto.String("2026-01-01T00:00:00Z"),
				Key:    proto.String(fmt.Sprintf("red-merge-r1:act:#%d", i*len(bodies)+j)),
			}
			typ, err := recordpb.SetBody(ev, body)
			if err != nil {
				b.Fatal(err)
			}
			ev.Type = &typ
			if _, err := Insert(db, ev); err != nil {
				b.Fatal(err)
			}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evs, err := Events(db)
		if err != nil {
			b.Fatal(err)
		}
		if len(evs) != 1200 {
			b.Fatalf("read %d events, want 1200", len(evs))
		}
	}
}
