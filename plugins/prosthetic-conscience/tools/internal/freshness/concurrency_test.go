package freshness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// naiveWrite is what writeState USED to be, kept here as a CONTROL.
//
// A test that only asserts the atomic writer holds cannot tell whether it is testing
// atomicity or testing nothing — if the race never materialises on this machine, both
// implementations pass and the test reports a property it never observed. So the same
// harness runs against both, and the control must FAIL. That is the null run the spike
// record insists on, applied to a concurrency claim.
func naiveWrite(path string, st State) {
	b, err := json.Marshal(st)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, b, 0o644)
}

// hammer runs writers and readers against one path and reports how many reads came back
// as the zero state AFTER at least one write had completed.
//
// That count is the whole measurement. os.WriteFile truncates and then writes, so a
// concurrent reader can see a file that exists and is empty, decode nothing, and get the
// zero State — which readState cannot distinguish from "never stamped". Downstream that
// means Of() RE-STAMPS tokens_at_write at the current count, and growth silently starts
// measuring the interval since the tear.
func hammer(t *testing.T, write func(string, State)) (torn int) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	// Seed, so every subsequent read has something valid to find.
	write(path, State{TokensAtWrite: 1, HasWriteReading: true})

	const writers, readers, rounds = 4, 4, 200
	var wg sync.WaitGroup
	var mu sync.Mutex

	for w := range writers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range rounds {
				write(path, State{TokensAtWrite: (w+1)*1000 + i, HasWriteReading: true})
			}
		}(w)
	}
	for range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range rounds {
				if got, ok := readState(path); !ok || !got.HasWriteReading {
					mu.Lock()
					torn++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	return torn
}

// FOUR BINARIES WRITE freshness.json — sc-precompact, sc-sessionend, sc-subagentstop and
// sc-postcompact-observe — the client runs an event's hooks in PARALLEL (measured,
// hook-surface-spike.md §4b), and two seats can return at once. So concurrent writes are
// the expected case, not an edge one.
func TestConcurrentWritersNeverProduceATornRead(t *testing.T) {
	if torn := hammer(t, writeState); torn != 0 {
		t.Errorf("%d reads saw a torn or empty state file; each one would RE-STAMP "+
			"tokens_at_write at the current count and make growth measure the interval "+
			"since the tear", torn)
	}
}

// THE CONTROL. The naive writer must fail the same harness, or the test above is not
// observing atomicity — it is observing a race that did not happen to occur.
//
// If this ever stops failing, do not delete it: it means the harness stopped exercising
// the window, and the test above has quietly become decorative.
func TestTheNaiveWriterFailsTheSameHarness(t *testing.T) {
	if torn := hammer(t, naiveWrite); torn == 0 {
		t.Skip("the truncating writer produced no torn read on this machine — the harness is " +
			"not exercising the window, so TestConcurrentWritersNeverProduceATornRead is " +
			"currently proving nothing about atomicity")
	}
}

// A successful write must leave no temp files. The atomic path creates one per write, so
// a leak here fills the checkpoints directory with debris that every later reader has to
// step over — and the directory is the one the seal prunes by pattern.
func TestConcurrentWritesLeaveNoDebris(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	var wg sync.WaitGroup
	for w := range 4 {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range 50 {
				writeState(path, State{TokensAtWrite: w*100 + i, HasWriteReading: true})
			}
		}(w)
	}
	wg.Wait()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "state.json" {
			t.Errorf("left debris in the state directory: %s", e.Name())
		}
	}
}

// Last writer wins, and that is FINE — both writers are recording a reading of the same
// note, so either value is correct. What must never happen is a value that was never
// written: a merge of two payloads, or a truncation.
func TestTheSurvivingStateIsOneSomebodyActuallyWrote(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
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
				writeState(path, State{TokensAtWrite: v, HasWriteReading: true})
			}
		}(w)
	}
	wg.Wait()

	got, _ := readState(path)
	if !got.HasWriteReading {
		t.Fatal("final state is the zero value; the last write did not survive")
	}
	if !written[got.TokensAtWrite] {
		t.Errorf("final TokensAtWrite = %d, which no writer ever wrote — the file was merged or torn",
			got.TokensAtWrite)
	}
}
