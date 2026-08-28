package stopnudge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/freshness"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/unwritable"
)

func at(min int) time.Time { return time.Date(2026, 8, 23, 0, min, 0, 0, time.UTC) }

// bands that are configured, so the nudge exists at all.
func bands() Thresholds {
	return Thresholds{
		TurnsNotice: 30, TurnsWarn: 60, TurnsUrgent: 120,
		GrowthNotice: 100_000, GrowthWarn: 300_000, GrowthUrgent: 600_000,
	}
}

// stale is a reading well past the floor and over the NOTICE edge.
func stale() freshness.Measures {
	return freshness.Measures{Turns: 40, TurnsMeasured: true}
}

func read(t *testing.T, dir string) (State, bool) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "checkpoints", "nudge.json"))
	if os.IsNotExist(err) {
		return State{}, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		t.Fatal(err)
	}
	return st, true
}

// THE TEST THAT MUST NEVER BE DELETED. A spent band emits NOTHING on the next boundary.
//
// Measured, not theorised: an unguarded injector fired 16 times on one prompt, produced
// 35 assistant entries and burned 4,326 output tokens (spike §13). Every other guard in
// this package is redundancy around this one assertion.
func TestASpentBandEmitsNothing(t *testing.T) {
	dir := t.TempDir()
	first := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10))
	if first.Emit == "" {
		t.Fatal("first crossing emitted nothing; the fixture does not exercise the guard")
	}
	second := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(11))
	if second.Emit != "" {
		t.Errorf("a spent band emitted again — this is the loop:\n%s", second.Emit)
	}
}

// The client's own re-entry flag, checked FIRST and needing no state. It is true on all
// fifteen re-entries after an emission (§13), so it stops a loop even if our record is
// gone.
func TestStopHookActiveSuppressesEvenWithNoStateAtAll(t *testing.T) {
	dir := t.TempDir()
	if d := Decide(dir, "s1", true, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
		t.Errorf("emitted inside a Stop-hook continuation:\n%s", d.Emit)
	}
	if _, exists := read(t, dir); exists {
		t.Error("wrote state while suppressed by stop_hook_active")
	}
}

// An unconfigured nudge does not exist — and must not LOOK like it does. The seal record
// derives nudge_enabled from this file's presence, so creating it while inert would make
// every Phase 1 row claim the nudge was live, and criterion 6 would compare an "after"
// against a "before" that is not one.
func TestAnUnconfiguredNudgeWritesNoStateFile(t *testing.T) {
	dir := t.TempDir()
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), Thresholds{}, at(10)); d.Emit != "" {
		t.Errorf("emitted with no thresholds configured:\n%s", d.Emit)
	}
	if _, exists := read(t, dir); exists {
		t.Error("created nudge.json with no thresholds; nudge_enabled would read true while the nudge has never fired")
	}
}

// Write before emit: if the record cannot be written, nothing is said. A guard that emits
// first has re-emitted whenever the write fails, and on Stop that is the loop.
//
// PORTABLE VARIANT. The checkpoints PATH is occupied by a regular file, so MkdirAll fails
// on every platform — where chmod does not: on Windows os.Chmod cannot make a directory
// unwritable, so the write succeeds and the emission is correct. That is a premise
// failing, not an assertion, and CI caught it as one.
func TestAnUnusableStateLocationSuppressesTheEmission(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A FILE where the directory must be.
	if err := os.WriteFile(filepath.Join(dir, ".claude", "checkpoints"), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
		t.Errorf("emitted with nowhere to record the band:\n%s", d.Emit)
	}
}

// The same property against an unwritable DIRECTORY, which is the shape a real machine
// produces (a permissions mistake, a read-only mount). Skipped rather than silently passing
// wherever the chmod does not actually restrict this process — os.Chmod does not restrict a
// directory on Windows, and a sufficiently privileged caller writes through it on Unix — so
// the arm never exercises nothing while reporting green. A test that passes without testing
// is worse than one that says it did not run, and [[unwritable]] holds that by PROBING the
// restriction rather than predicting it.
func TestAnUnwritableStateFileSuppressesTheEmission(t *testing.T) {
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	// unwritable.Dir owns both skips now — Windows, where os.Chmod is a no-op, AND a caller
	// privileged enough to write through 0o500, which this arm did NOT handle and which failed
	// it outright in a root container. TestAnUnusableStateLocationSuppressesTheEmission covers
	// the same property portably in both cases.
	unwritable.Dir(t, cp)

	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
		t.Errorf("emitted despite being unable to record the band:\n%s", d.Emit)
	}
}

// A corrupt record is not an empty one. Treating it as empty re-emits every band the
// session already spent; refusing costs one nudge.
func TestACorruptStateFileSuppressesRatherThanResets(t *testing.T) {
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cp, "nudge.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
		t.Errorf("emitted over a corrupt record:\n%s", d.Emit)
	}
}

// Criterion 4's cap suppresses unconditionally, whatever the bands say — bands re-arm on
// an answer, so without a non-resetting counter two checkpoint writes already permit six.
func TestTheHardCapSuppressesEvenWhenBandsHaveReArmed(t *testing.T) {
	dir := t.TempDir()
	note := at(1)
	for i := range maxEmissions {
		d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", note, bands(), at(10+i))
		if d.Emit == "" {
			t.Fatalf("emission %d suppressed early", i+1)
		}
		note = note.Add(time.Minute) // the note is answered each time, re-arming the band
	}
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", note, bands(), at(99)); d.Emit != "" {
		t.Errorf("emitted past the cap of %d:\n%s", maxEmissions, d.Emit)
	}
	st, _ := read(t, dir)
	if st.Emissions != maxEmissions {
		t.Errorf("Emissions = %d, want %d", st.Emissions, maxEmissions)
	}
}

// A band re-arms when the note moves — an answer of either kind — but the RECORD is not
// cleared, because nudge_answered's derivation reads it to know a band fired at all.
func TestAnAnsweredNoteReArmsTheBand(t *testing.T) {
	dir := t.TempDir()
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit == "" {
		t.Fatal("first crossing emitted nothing")
	}
	// Same note: silent.
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(11)); d.Emit != "" {
		t.Fatal("spent band emitted again")
	}
	// The note is answered — written or re-affirmed — so the band re-arms.
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(5), bands(), at(12)); d.Emit == "" {
		t.Error("band did not re-arm after the note was answered")
	}
}

// Session state does not outlive its session: a file from another session resets the
// counters, or four emissions EVER would suppress the nudge permanently in a project.
func TestADifferentSessionStartsWithAFreshBudget(t *testing.T) {
	dir := t.TempDir()
	for i := range maxEmissions {
		Decide(dir, "old", false, stale(), "CHECKPOINT.md", at(1+i), bands(), at(10+i))
	}
	if d := Decide(dir, "old", false, stale(), "CHECKPOINT.md", at(50), bands(), at(50)); d.Emit != "" {
		t.Fatal("the old session is not actually capped; fixture is wrong")
	}
	if d := Decide(dir, "new", false, stale(), "CHECKPOINT.md", at(51), bands(), at(51)); d.Emit == "" {
		t.Error("a new session inherited the old session's spent budget")
	}
}

// F7's floor: a session that has done too little has nothing worth interrupting for.
func TestBelowTheFloorNothingIsEmitted(t *testing.T) {
	dir := t.TempDir()
	small := freshness.Measures{Turns: floorTurns - 1, TurnsMeasured: true}
	if d := Decide(dir, "s1", false, small, "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
		t.Errorf("emitted below the floor:\n%s", d.Emit)
	}
	if _, exists := read(t, dir); exists {
		t.Error("wrote state for a session below the floor")
	}
}

// An unmeasured figure crosses nothing. A band fired from an unmeasured measure would be
// a nudge about a number nobody has.
func TestUnmeasuredFiguresCrossNoBand(t *testing.T) {
	dir := t.TempDir()
	unmeasured := freshness.Measures{Turns: 0, TurnsMeasured: false, Growth: 0, GrowthKnown: false}
	if d := Decide(dir, "s1", false, unmeasured, "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
		t.Errorf("emitted from unmeasured figures:\n%s", d.Emit)
	}
}

// ANY-OF, taking the max: the measures fail independently, so a session that burned
// tokens in few turns and one that took many turns without growth are both stale.
func TestGrowthAloneCanCrossABand(t *testing.T) {
	dir := t.TempDir()
	burned := freshness.Measures{Growth: 700_000, GrowthKnown: true}
	d := Decide(dir, "s1", false, burned, "CHECKPOINT.md", at(1), bands(), at(10))
	if d.Emit == "" {
		t.Fatal("growth alone crossed no band")
	}
	if d.Band != BandUrgent {
		t.Errorf("Band = %q, want URGENT — the highest edge any measure crosses", d.Band)
	}
}

// Band selection is the policy, so every edge is asserted from both measures. ANY-OF
// taking the MAX: the measures fail independently, so whichever is worse decides, and a
// reading that crosses two bands reports the higher one rather than the first matched.
func TestEveryBandEdgeFromEitherMeasure(t *testing.T) {
	th := bands() // turns 30/60/120, growth 100k/300k/600k
	for _, tc := range []struct {
		name string
		m    freshness.Measures
		want Band
		any  bool
	}{
		{"below every edge", freshness.Measures{Turns: 29, TurnsMeasured: true}, "", false},
		{"turns at NOTICE", freshness.Measures{Turns: 30, TurnsMeasured: true}, BandNotice, true},
		{"turns at WARN", freshness.Measures{Turns: 60, TurnsMeasured: true}, BandWarn, true},
		{"turns at URGENT", freshness.Measures{Turns: 120, TurnsMeasured: true}, BandUrgent, true},
		{"growth at NOTICE", freshness.Measures{Growth: 100_000, GrowthKnown: true}, BandNotice, true},
		{"growth at WARN", freshness.Measures{Growth: 300_000, GrowthKnown: true}, BandWarn, true},
		{"growth at URGENT", freshness.Measures{Growth: 600_000, GrowthKnown: true}, BandUrgent, true},
		// The lopsided case three measures exist for: few turns, enormous growth. An
		// all-of rule would call this fresh.
		{"few turns, huge growth", freshness.Measures{
			Turns: 5, TurnsMeasured: true, Growth: 700_000, GrowthKnown: true,
		}, BandUrgent, true},
		// And the mirror: many turns, no growth measured at all.
		{"many turns, growth unmeasured", freshness.Measures{Turns: 200, TurnsMeasured: true}, BandUrgent, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := highestBand(tc.m, th)
			if ok != tc.any || got != tc.want {
				t.Errorf("highestBand = (%q, %v), want (%q, %v)", got, ok, tc.want, tc.any)
			}
		})
	}
}

// WHAT THIS PACKAGE INHERITED, asserted here rather than assumed from the shared code.
//
// stopnudge's own copy of the state IO had no retry and no absent/unreadable distinction;
// freshness's had both, because CI exercised that one and the fix went where the red was.
// Moving to internal/statefile hands this package the stricter behaviour — and an
// inheritance nobody tests is a claim about another package's tests.
//
// The property that matters here is not the retry itself but what it protects: a spent
// band must survive concurrent access. A lost band is a re-emission, and on Stop a
// re-emission is the sixteen-firing loop rather than a duplicate nudge.
func TestASpentBandSurvivesConcurrentDecisions(t *testing.T) {
	dir := t.TempDir()
	if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit == "" {
		t.Fatal("first crossing emitted nothing; the fixture does not exercise the guard")
	}

	// Many boundaries arriving at once against the same spent band. Hooks on one event run
	// in PARALLEL (measured, hook-surface-spike.md §4b), so this is the expected shape.
	const n = 24
	var wg sync.WaitGroup
	var mu sync.Mutex
	var emissions int
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if d := Decide(dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(11+i)); d.Emit != "" {
				mu.Lock()
				emissions++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if emissions != 0 {
		t.Errorf("%d of %d concurrent boundaries re-emitted a spent band; each one is a turn the "+
			"model did not need to take", emissions, n)
	}
	// And the record still says what it should, rather than having been torn by the race.
	st, ok := read(t, dir)
	if !ok {
		t.Fatal("state file gone after concurrent access")
	}
	if len(st.BandsSpent) != 1 {
		t.Errorf("BandsSpent = %v, want exactly one band recorded", st.BandsSpent)
	}
}
