package statefile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
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

// AppendRow ACCUMULATES where Write REPLACES, and the difference is not stylistic: a lost
// state file is re-derivable at the next boundary, a lost row is a measurement nothing can
// reconstruct. These are the two records this suite's analysis actually reads.
func TestAppendRowAccumulatesRatherThanReplacing(t *testing.T) {
	p := filepath.Join(t.TempDir(), "deep", "rows.jsonl")
	for i := range 5 {
		if err := AppendRow(p, rec{N: i, Set: true}); err != nil {
			t.Fatalf("row %d: %v", i, err)
		}
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5 — an appender that replaces loses every earlier row", len(lines))
	}
	for i, line := range lines {
		var got rec
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d is not one complete JSON object (%v): %q", i, err, line)
		}
		if got.N != i {
			t.Errorf("line %d holds n=%d", i, got.N)
		}
	}
}

// Parallel writers must not interleave. The client runs an event's hooks CONCURRENTLY
// (hook-surface-spike.md §4b) and all three sealing shims append to seals.jsonl, so a torn
// row here is a corrupt line in the corpus that sets Phase 2's thresholds.
func TestConcurrentAppendersProduceWholeLines(t *testing.T) {
	p := filepath.Join(t.TempDir(), "rows.jsonl")
	const writers, rounds = 4, 100
	var wg sync.WaitGroup
	for w := range writers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range rounds {
				_ = AppendRow(p, rec{N: w*1000 + i, Name: strings.Repeat("x", 64), Set: true})
			}
		}(w)
	}
	wg.Wait()

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) != writers*rounds {
		t.Errorf("got %d lines, want %d — rows were lost or split", len(lines), writers*rounds)
	}
	for i, line := range lines {
		var got rec
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d is torn — a reader hits this and the row is unrecoverable: %q", i, line)
		}
	}
}

func TestAppendRowReportsAnUnusableLocation(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AppendRow(filepath.Join(blocker, "rows.jsonl"), rec{N: 1}); err == nil {
		t.Error("AppendRow reported success with nowhere to write")
	}
}

// A VALUE THAT CANNOT BE ENCODED MUST NOT DESTROY THE RECORD IT WAS MEANT TO REPLACE.
//
// Write truncates nothing until it renames, so an encoding failure has to leave the old
// state intact — and it has to SAY so, because the caller's next move (stamp a baseline,
// suppress an emission) depends on whether the write happened.
//
// These branches were the package's coverage hole: everything that could go wrong after
// the marshal was untested in both functions.
func TestAnUnencodableValueLeavesTheExistingRecordAlone(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.json")
	if err := Write(p, rec{N: 7, Set: true}); err != nil {
		t.Fatal(err)
	}

	// A channel cannot be JSON. This is the shape of any caller that grows a field it
	// cannot encode — an error the author should see, not a record silently emptied.
	if err := Write(p, make(chan int)); err == nil {
		t.Error("Write reported success for a value it cannot encode")
	}

	got, st := Read[rec](p)
	if st != Present || got.N != 7 {
		t.Errorf("after a failed write the record reads (%+v, %v), want the original n=7 present", got, st)
	}
	entries, err := os.ReadDir(filepath.Dir(p))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("a failed write left debris: %d entries, want just state.json", len(entries))
	}
}

// The same for the appender, where the stakes differ: a bad row must not truncate the
// corpus, and must not append a half-line that makes every later read of the file fail.
func TestAnUnencodableRowLeavesTheCorpusReadable(t *testing.T) {
	p := filepath.Join(t.TempDir(), "rows.jsonl")
	if err := AppendRow(p, rec{N: 1, Set: true}); err != nil {
		t.Fatal(err)
	}
	if err := AppendRow(p, make(chan int)); err == nil {
		t.Error("AppendRow reported success for a value it cannot encode")
	}

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 — the failed row was written anyway", len(lines))
	}
	var got rec
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil || got.N != 1 {
		t.Errorf("the surviving row is not the one that was written: %q (%v)", lines[0], err)
	}
}

// A read on a path that is not a file at all — a DIRECTORY where the record should be —
// is UNREADABLE, never Absent. It is the state a botched cleanup leaves, and reporting it
// as "nothing written yet" invites the caller to try to write over a directory forever.
func TestADirectoryWhereTheRecordBelongsIsUnreadable(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.json")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, st := Read[rec](p); st != Unreadable {
		t.Errorf("a directory at the record's path reads as %v, want unreadable", st)
	}
}

// A TRANSIENTLY MISSING NAME IS NOT AN ABSENT RECORD, and collapsing the two is the
// failure this package exists to prevent — in the one direction that overwrites data.
//
// Windows can report ErrNotExist for an instant while a rename binds over a name. The
// first version of Read trusted that instant and returned Absent, which tells the caller
// "nothing has been stamped yet, go ahead and write" about a record that was there the
// whole time. It went unseen for as long as CI ran on slow storage: the window is real on
// any substrate, but a RAM disk samples it often enough to catch — 3 of 800 reads, on a
// file that existed before and after.
//
// The test drives the same shape deterministically: read a name that does not exist yet
// and appears mid-retry, and require Read to find it rather than call it absent.
func TestATransientlyMissingFileIsRetriedRatherThanCalledAbsent(t *testing.T) {
	// Widen the window so a loaded runner cannot turn this into a coin flip: the writer
	// lands well inside the first backoff.
	orig := readBackoff
	readBackoff = 25 * time.Millisecond
	t.Cleanup(func() { readBackoff = orig })

	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	written := make(chan struct{})
	go func() {
		defer close(written)
		time.Sleep(5 * time.Millisecond)
		if err := Write(path, rec{N: 7, Name: "x", Set: true}); err != nil {
			t.Errorf("write: %v", err)
		}
	}()

	got, st := Read[rec](path)
	<-written
	if st != Present || got.N != 7 {
		t.Errorf("a file that appeared mid-read → (%+v, %v), want (n=7, present) — "+
			"reporting it absent tells the caller to overwrite a record that exists", got, st)
	}
}

// AND THE HONEST ABSENCE STILL RESOLVES ABSENT, which is what stops the fix above from
// being a retry that swallows the common case: a name that never appears must not come
// back Unreadable, or every first run would be forbidden from writing its first record.
func TestAFileThatNeverAppearsIsAbsentNotUnreadable(t *testing.T) {
	got, st := Read[rec](filepath.Join(t.TempDir(), "never.json"))
	if st != Absent || got.Set {
		t.Errorf("a name that never appears → (%+v, %v), want (zero, absent)", got, st)
	}
}

// THE EXPORTED READER RETRIES TOO, because the callers that need it most do not come
// through Read.
//
// checkpoint.LoadRearm takes its reader as a parameter, and the one caller reading the
// re-arm record outside the re-arm lock passed os.ReadFile bare — so the transient miss
// that Read now absorbs went straight through to a caller that reads it as "no re-arm
// history". Same window, same wrong conclusion, one signature away.
func TestReadFileFindsAFileThatAppearsMidRetry(t *testing.T) {
	orig := readBackoff
	readBackoff = 25 * time.Millisecond
	t.Cleanup(func() { readBackoff = orig })

	dir := t.TempDir()
	path := filepath.Join(dir, "rearm.json")

	written := make(chan struct{})
	go func() {
		defer close(written)
		time.Sleep(5 * time.Millisecond)
		if err := os.WriteFile(path, []byte(`{"n":7}`), 0o644); err != nil {
			t.Errorf("write: %v", err)
		}
	}()

	b, err := ReadFile(path)
	<-written
	if err != nil || !strings.Contains(string(b), `"n":7`) {
		t.Errorf("a file that appeared mid-read → (%q, %v), want the contents — a bare read "+
			"here reports ErrNotExist and the caller concludes the record is empty", b, err)
	}
}

// AND A NAME THAT NEVER APPEARS STILL REPORTS ErrNotExist, which is what lets callers keep
// telling absent from unreadable after the retry.
func TestReadFileStillReportsAGenuineAbsence(t *testing.T) {
	if _, err := ReadFile(filepath.Join(t.TempDir(), "never.json")); !os.IsNotExist(err) {
		t.Errorf("a name that never appears → %v, want a not-exist error", err)
	}
}

// A RENAME THAT FAILS MUST BE REPORTED AS A FAILURE. This is the last step of the atomic
// write, and it is the one where "I told you I saved it" and "I saved it" come apart: the
// temp file holds the new bytes, the real path still holds the old ones, and a caller told
// `nil` believes its state is durable when nothing moved.
//
// Found by mutation testing, not by coverage — the branch was EXECUTED by every successful
// write, so it read as covered while the failing half had never run. Inverting the
// condition left the suite green.
//
// A DIRECTORY at the destination is the portable way to make rename fail: POSIX gives
// EISDIR/ENOTEMPTY and Windows refuses it too, where a permission trick would not.
func TestAFailedRenameIsReportedAndLeavesNoDebris(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "state.json")
	if err := os.MkdirAll(filepath.Join(dest, "occupied"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Write(dest, rec{N: 1, Set: true}); err == nil {
		t.Error("Write reported success while the record was never moved into place")
	}

	// The temp file must not survive: this directory is pruned by pattern elsewhere, and a
	// caller that retries would otherwise accumulate one orphan per attempt.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "state.json" {
			t.Errorf("a failed write left %s behind", e.Name())
		}
	}
}
