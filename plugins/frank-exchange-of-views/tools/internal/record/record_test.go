package record

import (
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
	ev, err := Append(Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}, "observe", NewPayload().Set("label", "L1-O1"))
	if err != nil {
		t.Fatal(err)
	}
	if ev.TS == "" {
		t.Fatal("an event carries no timestamp; replay would order it by seat name")
	}
}

// MONOTONICITY IS GUARANTEED IN CODE, NOT BORROWED FROM THE HARDWARE.
//
// Precision only narrows the window in which two events can tie. Two seats can still stamp
// the same instant, and a wall clock can step BACKWARDS under NTP and issue a stamp earlier
// than one already written. Both produce a tie or an inversion in the ordering key, and the
// sort then falls through to seat name — the defect that dropped the bench's closures.
//
// These freeze or reverse the clock outright: conditions no precision can survive, and
// exactly what a real clock does occasionally. The stamps must still strictly increase.
func TestStampsStrictlyIncreaseUnderAFrozenClock(t *testing.T) {
	orig := Now
	defer func() { Now = orig }()
	frozen := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	Now = func() time.Time { return frozen }

	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}); err != nil {
		t.Fatal(err)
	}

	var stamps []string
	for i := 0; i < 10; i++ {
		ev, err := Append(Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}, "observe", NewPayload().Set("label", string(rune('a'+i))))
		if err != nil {
			t.Fatal(err)
		}
		stamps = append(stamps, ev.TS)
	}
	for i := 1; i < len(stamps); i++ {
		if stamps[i] <= stamps[i-1] {
			t.Fatalf("a frozen clock produced non-increasing stamps: %s then %s. Order must be a property of the CODE, not of the machine's clock", stamps[i-1], stamps[i])
		}
	}
}

// The nastier case: the clock runs BACKWARDS. A stamp must never be issued that would sort
// before one already written, or replay reorders events that are already on disk.
func TestStampsStrictlyIncreaseWhenTheClockRunsBackwards(t *testing.T) {
	orig := Now
	defer func() { Now = orig }()
	base := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	var call int
	Now = func() time.Time {
		call++
		return base.Add(-time.Duration(call) * time.Second) // each call earlier than the last
	}

	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}); err != nil {
		t.Fatal(err)
	}

	var stamps []string
	for i := 0; i < 6; i++ {
		ev, err := Append(Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: RoundOf("red-lens-r1-L1")}, "observe", NewPayload().Set("label", string(rune('a'+i))))
		if err != nil {
			t.Fatal(err)
		}
		stamps = append(stamps, ev.TS)
	}
	for i := 1; i < len(stamps); i++ {
		if stamps[i] <= stamps[i-1] {
			t.Fatalf("a backwards-running clock produced non-increasing stamps: %s then %s. The stamps may drift from wall time — that is the accepted trade — but they must never lie about WHAT CAME FIRST", stamps[i-1], stamps[i])
		}
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

	ev, err := Append(Identity{RunDir: dir, SeatID: "judge-terminal", Round: 7}, "friction", NewPayload().Set("reason", "a capability gap"))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Round != 7 {
		t.Errorf("the event must carry the round it was GIVEN, not the one its id looks like: got %d, want 7", ev.Round)
	}
	// And the register the append triggered carries it too — both write sites take the seam.
	evs, err := ReadShard(filepath.Join(dir, "records", "events-judge-terminal-"+ev.Nonce+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range evs {
		if e.Round != 7 {
			t.Errorf("%s event stamped round %d, want 7 — RegisterSeat must take the same seam as Append", e.Type, e.Round)
		}
	}
	// Party is NOT taken from the caller: it stays derived from the seat id, because the
	// caller's Role answers which command group is running, not who is writing.
	if ev.Role != "bench" {
		t.Errorf("the party is the seat's, derived from its id: got %q, want bench", ev.Role)
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
	ev, err := Append(Identity{RunDir: dir, SeatID: "judge-terminal", Round: -1}, "friction", NewPayload().Set("reason", "a capability gap"))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Round != -1 {
		t.Errorf("unknown must stay unknown on the record: got %d, want -1", ev.Round)
	}
}
