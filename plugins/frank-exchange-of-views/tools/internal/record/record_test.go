package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// THE ORDERING KEY MUST BE FINE ENOUGH TO ORDER.
//
// `ts` is not a field a human reads; it is the key replay sorts by, and when two events
// share it the sort falls through to (SeatID, Seq) — ordering by SEAT NAME, which is the
// exact defect that silently dropped the bench's closures ("judge-r2" sorting before
// "red-merge-r1", so every ruling replayed before the mint it referenced).
//
// The stamp was millisecond-precision, so any two events inside one tick tied and fell
// back to that broken order.
//
// THE CLOCK IS INJECTED, deliberately. The first version of this test appended twenty
// events back-to-back and asserted the stamps were distinct — and it PASSED on the
// millisecond clock it was written to condemn, because file I/O makes each append take
// longer than a tick on this machine. It proved the machine was slow, not that the format
// was adequate, and it would have gone green on the very build CI rejected. Driving Now
// with a fake clock that advances by less than a millisecond asks the real question:
// can this format represent two events that happen close together?
func TestEventStampsResolveSubMillisecondEvents(t *testing.T) {
	orig := Now
	defer func() { Now = orig }()

	base := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	var tick int
	// 100 microseconds apart: far coarser than the clock actually delivers, and still
	// invisible to a millisecond format.
	Now = func() time.Time {
		tick++
		return base.Add(time.Duration(tick) * 100 * time.Microsecond)
	}

	const n = 20
	seen := map[string]bool{}
	var stamps []string
	for i := 0; i < n; i++ {
		s := stamp()
		seen[s] = true
		stamps = append(stamps, s)
	}

	if len(seen) != n {
		t.Errorf("%d events 100µs apart produced only %d distinct stamps — the %d collisions fall back to ordering by SEAT NAME, which is how the bench's closures were dropped",
			n, len(seen), n-len(seen))
	}
	for i := 1; i < len(stamps); i++ {
		if stamps[i] <= stamps[i-1] {
			t.Errorf("stamps are not strictly increasing: %s then %s. Lexicographic order must BE time order, or the fixed-width format is lying about what it encodes", stamps[i-1], stamps[i])
		}
	}
}

// And the whole path, end to end: an appended event carries a stamp at all.
func TestAppendedEventCarriesAStamp(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}); err != nil {
		t.Fatal(err)
	}
	ev, err := Append(Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}, &recordpb.Observe{Label: proto.String("L1-O1")})
	if err != nil {
		t.Fatal(err)
	}
	if ev.GetTs() == "" {
		t.Fatal("an event carries no timestamp; replay would order it by seat name")
	}
}

// THE ORDER IS A PROPERTY OF THE CODE, NOT OF THE MACHINE'S CLOCK — and the code that holds it
// changed, so these two tests ask about the guarantee rather than about its old carrier.
//
// They used to assert that STAMPS strictly increase under a frozen clock and under one running
// backwards. That mattered because replay sorted by (TS, SeatID, Seq): a tie or an inversion in
// the stamp fell through to seat NAME, which is the defect that dropped a whole sitting's bench
// closures. A monotonic clock file existed for exactly that reason.
//
// Order is `events.id` now — assigned by the thing doing the inserting — so `ts` is informational
// and may tie or step backwards without consequence. The guarantee is unchanged and stronger, so
// it is asserted directly: whatever the clock does, the record reads back in the order it was
// written.
func TestTheReadOrderIsTheWriteOrderWhateverTheClockDoes(t *testing.T) {
	for _, c := range []struct {
		name string
		now  func() time.Time
	}{
		{"a frozen clock", func() func() time.Time {
			frozen := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
			return func() time.Time { return frozen }
		}()},
		{"a clock running backwards", func() func() time.Time {
			base := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
			var call int
			return func() time.Time {
				call++
				return base.Add(-time.Duration(call) * time.Second)
			}
		}()},
	} {
		t.Run(c.name, func(t *testing.T) {
			orig := Now
			defer func() { Now = orig }()
			Now = c.now

			runDir := t.TempDir()
			id := Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}
			if _, _, err := RegisterSeat(id); err != nil {
				t.Fatal(err)
			}
			var wrote []string
			for i := 0; i < 6; i++ {
				label := string(rune('a' + i))
				if _, err := Append(id, &recordpb.Observe{Label: proto.String(label)}); err != nil {
					t.Fatal(err)
				}
				wrote = append(wrote, label)
			}

			m, err := MergedEvents(runDir)
			if err != nil {
				t.Fatal(err)
			}
			var read []string
			for _, e := range m.Events {
				if o, ok := recordpb.BodyAs[*recordpb.Observe](e); ok {
					read = append(read, o.GetLabel())
				}
			}
			if len(read) != len(wrote) {
				t.Fatalf("%d observations read, %d written", len(read), len(wrote))
			}
			for i := range wrote {
				if read[i] != wrote[i] {
					t.Fatalf("under %s the record read back in a different order: %v, want %v.\n\n"+
						"The stamps may tie or drift — that is what a real clock does — but the record "+
						"must never lie about WHAT CAME FIRST", c.name, read, wrote)
				}
			}
		})
	}
}

// #396: THE ROUND IS CARRIED TO THE WRITE, NOT RECOVERED AT IT.
//
// `Append` used to stamp `Round: RoundOf(seatID)` — a regex over the seat id, 0 on a miss —
// while the caller had already resolved the round as a field and `Begin` had already refused an
// unresolvable seat. The fact was in hand and thrown away one frame later.
//
// This is the regression guard for that seam: if the write ever goes back to deriving, the
// injected round stops arriving and this fails. `judge-terminal` is the right probe because it
// carries no `-r<N>`, so the two answers are distinguishable — the derivation says 0.
func TestAppendStampsTheRoundItIsGiven(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if RoundOf("judge-terminal") != 0 {
		t.Fatal("fixture assumption: the seat id derivation answers 0 for judge-terminal")
	}

	ev, err := Append(Identity{RunDir: dir, SeatID: "judge-terminal", Round: 7}, &recordpb.Friction{Text: proto.String("a capability gap")})
	if err != nil {
		t.Fatal(err)
	}
	if ev.GetRound() != 7 {
		t.Errorf("the event must carry the round it was GIVEN, not the one its id looks like: got %d, want 7", ev.GetRound())
	}
	// And every other event this seat wrote carries it too — both write sites take the seam. Read
	// from the record rather than from a named shard file: there is no filename to compose, which
	// is one fewer place for the test to encode where the events live.
	m, err := MergedEvents(dir)
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, e := range m.Events {
		if e.GetSeatId() != "judge-terminal" {
			continue
		}
		seen++
		if e.GetRound() != 7 {
			t.Errorf("%s event stamped round %d, want 7 — RegisterSeat must take the same seam as Append",
				recordpb.Word(e.GetType()), e.GetRound())
		}
	}
	if seen == 0 {
		t.Fatal("the seat wrote nothing — an empty traversal passes the assertion above on every event")
	}
	// Party is NOT taken from the caller: it stays derived from the seat id, because the
	// caller's Role answers which command group is running, not who is writing.
	if ev.GetRole() != "bench" {
		t.Errorf("the party is the seat's, derived from its id: got %q, want bench", ev.GetRole())
	}
}

// -1 IS A REAL VALUE ON THIS SEAM and it means unknown, which is not round 0. Nothing produces
// it today: every verb reached through `Begin` has a resolved round, and the three `motion`
// verbs — which deliberately skip `Begin` (see cli/motion/verbs.go) — still resolve one whenever
// a seat id is present. It becomes reachable only when an injected identity CONFLICTS with a
// typed --seat-id, and nothing sets FEOV_SEAT yet (#290).
//
// Pinned so the behaviour is a decision rather than a discovery: an unknown round is written as
// unknown. It is NOT quietly converted to 0, which is synthesis and a real round — the
// conflation that produced the phantom archive in #327.
func TestAnUnknownRoundIsWrittenAsUnknownNotAsZero(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	ev, err := Append(Identity{RunDir: dir, SeatID: "judge-terminal", Round: -1}, &recordpb.Friction{Text: proto.String("a capability gap")})
	if err != nil {
		t.Fatal(err)
	}
	if ev.GetRound() != -1 {
		t.Errorf("unknown must stay unknown on the record: got %d, want -1", ev.GetRound())
	}
}
