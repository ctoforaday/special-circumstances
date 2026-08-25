package statefile

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type rec struct {
	N    int    `json:"n"`
	Name string `json:"name"`
	Set  bool   `json:"set"`
}

// THE TRI-STATE, which is the whole reason this package exists rather than a pair of
// helpers. Absent and Unreadable lead to OPPOSITE actions — the first invites a write,
// the second forbids one — so a caller that cannot tell them apart overwrites a record it
// could not see.
func TestAbsentAndUnreadableAreDifferentAnswers(t *testing.T) {
	dir := t.TempDir()

	if got, st := Read[rec](filepath.Join(dir, "nope.json")); st != Absent || got.Set {
		t.Errorf("missing file → (%+v, %v), want (zero, absent)", got, st)
	}

	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, st := Read[rec](bad); st != Unreadable || got.Set {
		t.Errorf("corrupt file → (%+v, %v), want (zero, unreadable)", got, st)
	}

	good := filepath.Join(dir, "good.json")
	if err := Write(good, rec{N: 7, Name: "x", Set: true}); err != nil {
		t.Fatal(err)
	}
	if got, st := Read[rec](good); st != Present || got.N != 7 {
		t.Errorf("written file → (%+v, %v), want (n=7, present)", got, st)
	}
}

// An empty file is UNREADABLE, not empty. This is the exact byte-state a truncating write
// leaves behind mid-flight, and reporting it as Absent is what made a concurrent reader
// re-stamp a baseline it could not see.
func TestAnEmptyFileIsUnreadableRatherThanAbsent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got, st := Read[rec](p); st != Unreadable {
		t.Errorf("empty file → (%+v, %v), want unreadable — a truncating writer leaves exactly this", got, st)
	}
}

// Write creates the directory it needs. Both callers write under .claude/checkpoints/,
// which may not exist on a first run, and a hook that fails because a parent directory is
// missing has failed for a reason nobody would guess from the symptom.
func TestWriteCreatesItsDirectory(t *testing.T) {
	p := filepath.Join(t.TempDir(), "deep", "deeper", "state.json")
	if err := Write(p, rec{N: 1, Set: true}); err != nil {
		t.Fatalf("Write into a missing directory: %v", err)
	}
	if _, st := Read[rec](p); st != Present {
		t.Errorf("read back after creating the directory: %v", st)
	}
}

// An unusable location is an ERROR, not a silent success. The callers differ in what they
// do about it — one refuses to stamp, one refuses to emit — and both need to be told.
func TestWriteReportsAnUnusableLocation(t *testing.T) {
	dir := t.TempDir()
	// A FILE where the directory must be: MkdirAll fails on every platform, unlike chmod.
	blocker := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(filepath.Join(blocker, "state.json"), rec{N: 1}); err == nil {
		t.Error("Write reported success with nowhere to write")
	}
}

// THE PROPERTY THE EXTRACTION EXISTS TO SHARE. Concurrent writers must never leave a
// reader holding an honest-looking empty record.
//
// This is the test `stopnudge` never had: its own copy of this code was written without a
// retry and without the tri-state, and nothing exercised it under concurrency. Both
// callers now inherit the behaviour AND this assertion.
func TestConcurrentWritersNeverYieldAnAbsentReading(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.json")
	if err := Write(p, rec{N: 1, Set: true}); err != nil {
		t.Fatal(err)
	}

	const writers, readers, rounds = 4, 4, 150
	var wg sync.WaitGroup
	var mu sync.Mutex
	var absent, unreadable int

	for w := range writers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range rounds {
				_ = Write(p, rec{N: w*1000 + i, Set: true})
			}
		}(w)
	}
	for range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range rounds {
				_, st := Read[rec](p)
				mu.Lock()
				switch st {
				case Absent:
					absent++
				case Unreadable:
					unreadable++
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	t.Logf("absent=%d unreadable=%d of %d reads", absent, unreadable, readers*rounds)

	if absent != 0 {
		t.Errorf("%d reads reported ABSENT while the file existed throughout; a caller would "+
			"treat that as 'nothing written yet' and overwrite a record it could not see", absent)
	}
}

// Last writer wins is acceptable; a value nobody wrote is not.
func TestTheSurvivingRecordIsOneSomebodyWrote(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.json")
	written := map[int]bool{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := range 4 {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range 50 {
				v := w*1000 + i
				mu.Lock()
				written[v] = true
				mu.Unlock()
				_ = Write(p, rec{N: v, Set: true})
			}
		}(w)
	}
	wg.Wait()

	got, st := Read[rec](p)
	if st != Present {
		t.Fatalf("final read: %v", st)
	}
	if !written[got.N] {
		t.Errorf("final n=%d, which no writer wrote — the file was merged or torn", got.N)
	}
}

// Temp files are an implementation detail and must not survive as one. The directory these
// live in is pruned by pattern elsewhere, so debris there is not inert.
func TestNoTempFilesSurviveASuccessfulWrite(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")
	for i := range 20 {
		if err := Write(p, rec{N: i, Set: true}); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "state.json" {
			t.Errorf("left behind %s", e.Name())
		}
	}
}

// Status names itself, because it reaches humans: it appears in test failures and in the
// stderr of hooks that decline to act. "unreadable" and "absent" are the two words a
// reader needs to tell an overwrite-safe state from an overwrite-forbidden one, so a bare
// integer there would put the burden back on whoever is debugging at the time.
func TestStatusNamesItself(t *testing.T) {
	for st, want := range map[Status]string{Absent: "absent", Present: "present", Unreadable: "unreadable"} {
		if got := st.String(); got != want {
			t.Errorf("Status(%d).String() = %q, want %q", int(st), got, want)
		}
	}
}
