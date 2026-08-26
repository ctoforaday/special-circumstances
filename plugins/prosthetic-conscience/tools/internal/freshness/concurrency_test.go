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

// hammer runs writers and readers against one path and reports what the readers saw.
//
//	empty   — readState returned (zero, ok=true): an HONEST EMPTY STATE, which after a
//	          write has landed is a LIE. This is the dangerous outcome: Of() would stamp
//	          over a reading it could not see, and growth would measure from that moment.
//	unread  — readState returned ok=false: "the record exists and I could not read it".
//	          Costly but SAFE — Of() refuses to stamp, and one growth figure is lost.
func hammer(t *testing.T, write func(string, State)) (empty, unread, reads int) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
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
				got, ok := readState(path)
				mu.Lock()
				switch {
				case !ok:
					unread++
				case !got.HasWriteReading:
					empty++
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	// READS IS RETURNED, not recomputed by the caller. A guard expressed as a fraction of the
	// reads that actually happened cannot drift out of step with the harness the way a
	// hand-copied denominator would.
	return empty, unread, readers * rounds
}

// THE SAFETY PROPERTY, and it is portable: a read must NEVER report an honest empty state
// once a write has landed.
//
// This is the one that protects the measurement. "Nothing has been stamped yet" and "I
// could not read what was stamped" lead to opposite actions — the first says stamp now,
// the second says do not touch it — so a reader that confuses them re-stamps
// tokens_at_write at the current count and growth silently measures from there.
//
// It must hold for BOTH writers, which is why it is asserted against both: the truncating
// writer is unsafe for a different reason, and this property is not the one that
// distinguishes them.
func TestNoReaderEverSeesAnHonestEmptyStateAfterAWrite(t *testing.T) {
	for _, w := range []struct {
		name  string
		write func(string, State)
	}{{"atomic", writeState}, {"truncating", naiveWrite}} {
		t.Run(w.name, func(t *testing.T) {
			empty, _, _ := hammer(t, w.write)
			if empty != 0 {
				t.Errorf("%d reads reported an EMPTY state after a write had landed; each one "+
					"would re-stamp tokens_at_write and make growth measure from that moment", empty)
			}
		})
	}
}

// WHAT ATOMICITY ACTUALLY BUYS, stated as a comparison because an absolute is not portable.
//
// An earlier version of this test asserted that the atomic writer produces ZERO unreadable
// reads. That is true on Linux and false on Windows — rename is atomic for the FILE, but a
// concurrent open can still hit a sharing violation while it lands, and CI measured 21 of
// 800, then 3 of 800 once the read retried. Neither is a defect: an unreadable read is
// safe, it just costs one measurement.
//
// So the claim is relative, and it is the claim worth making: truncating writes are
// unreadable far more often than atomic ones.
//
// # THE TWO PLATFORMS DO NOT MEASURE THE SAME THING, and that is what made this flaky
//
// The comparison only means something where the truncation window is what the readers are
// hitting. Measured:
//
//	Linux    atomic 0 unreadable, truncating 638 of 800 — five runs, identical
//	Windows  atomic 1 unreadable, truncating   2 of 800 — CI, on a commit that also passed
//
// The mechanisms differ. `naiveWrite` truncates in place, so a Linux reader catches the file
// mid-write, `os.ReadFile` SUCCEEDS on short bytes, and `json.Unmarshal` fails — statefile's
// retry cannot help, because nothing errored. On Windows the same instant usually produces a
// SHARING VIOLATION instead: `os.ReadFile` errors, and `statefile.ReadFile` retries it away —
// that retry exists, by its own comment, to "absorb the transient cases measured on Windows".
// So on Windows both writers are reduced to the same small residue of retry exhaustion, and
// the ratio between two residues is noise.
//
// That is why 1-against-2 failed a run where atomicity did its job, and why a bigger sample
// cannot fix it: more rounds scale both residues together. The guard has to key on whether
// the TRUNCATION WINDOW was exercised at all, which is a fraction of reads, not a count.
//
// If that stops being true where it IS exercised, either the write stopped being atomic or
// the harness stopped opening the window — and both are worth a failure.
func TestAtomicWritesAreReadableFarMoreOftenThanTruncatingOnes(t *testing.T) {
	atomicEmpty, atomicUnread, reads := hammer(t, writeState)
	naiveEmpty, naiveUnread, _ := hammer(t, naiveWrite)
	t.Logf("atomic: %d empty, %d unreadable | truncating: %d empty, %d unreadable | %d reads each",
		atomicEmpty, atomicUnread, naiveEmpty, naiveUnread, reads)

	// THE THRESHOLD IS SET BY WHAT THE TWO REGIMES MEASURE, not picked for comfort. Retry
	// residue has been observed as high as 21 of 800 (2.6%); a real truncation window shows
	// 638 of 800 (80%). A tenth of the reads sits four times above the worst residue ever
	// seen here and eight times below the signal, so it separates them without being close
	// to either.
	if naiveUnread*10 < reads {
		t.Skipf("the truncating writer was unreadable on only %d of %d reads — below the tenth "+
			"that marks a truncation window actually being hit, so this platform is measuring "+
			"retry exhaustion rather than atomicity, and the ratio below would compare two "+
			"residues", naiveUnread, reads)
	}
	if atomicUnread*4 > naiveUnread {
		t.Errorf("atomic writes were unreadable %d times against the truncating writer's %d; "+
			"atomicity should make this rare, not merely less common", atomicUnread, naiveUnread)
	}
}
