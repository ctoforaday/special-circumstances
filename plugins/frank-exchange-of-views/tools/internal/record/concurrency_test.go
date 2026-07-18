package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestConcurrentSeatsRace is what the port buys that the oracle could not have:
// `go test -race` over the real lock and render paths. The mjs suite could only
// SPAWN six processes and check that nothing was visibly lost — a black-box
// assertion that passes right up until an interleaving it did not happen to hit.
//
// Six lens seats writing while every verb triggers a render is the live shape:
// six shards (single-writer each, so no shard race) contending on the SHARED
// surfaces — the seat pointer and the projection files.
func TestConcurrentSeatsRace(t *testing.T) {
	runDir := t.TempDir()
	const seats = 6
	const perSeat = 4

	var wg sync.WaitGroup
	errs := make(chan error, seats*perSeat*3)
	for s := 1; s <= seats; s++ {
		wg.Add(1)
		go func(s int) {
			defer wg.Done()
			seatID := fmt.Sprintf("red-lens-r1-L%d", s)
			if _, _, err := RegisterSeat(runDir, seatID); err != nil {
				errs <- err
				return
			}
			for i := 0; i < perSeat; i++ {
				p := NewPayload().
					Set("label", fmt.Sprintf("L%d-F%d", s, i)).
					Set("severity", "medium").Set("likelihood", "medium").Set("impact", "high").
					Set("text", strings.Repeat("finding prose ", 20))
				if _, err := Append(runDir, seatID, "finding", p); err != nil {
					errs <- err
					continue
				}
				if _, err := Render(runDir, ""); err != nil {
					errs <- err
				}
			}
		}(s)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent seat: %v", err)
	}

	// Every event landed: no shard lost a write to a racing render.
	m, err := MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	findings := 0
	for _, e := range m.Events {
		if e.Type == "finding" {
			findings++
		}
	}
	if want := seats * perSeat; findings != want {
		t.Errorf("findings landed = %d, want %d", findings, want)
	}
	if len(m.Anomalies) != 0 {
		t.Errorf("unexpected anomalies under clean concurrency: %v", m.Anomalies)
	}

	// No lock or temp artifact leaked: a stuck .lock- directory would deadlock the
	// next seat for the full bounded wait, and a stray .tmp- would be read as a
	// projection by anything globbing the shadow dir.
	entries, err := os.ReadDir(filepath.Join(runDir, "records"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".lock-") || strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leaked artifact: %s", e.Name())
		}
	}
	shadow, err := os.ReadDir(filepath.Join(runDir, "records", "render-shadow"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range shadow {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leaked temp render: %s", e.Name())
		}
	}
}

// TestStaleLockIsStolen: a crashed holder must not deadlock the run. The lock is
// stolen after staleMS by mtime, and the bounded wait then proceeds rather than
// blocking a seat forever.
func TestStaleLockIsStolen(t *testing.T) {
	runDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(runDir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(runDir, "records", ".lock-render")
	if err := os.Mkdir(lock, 0o755); err != nil {
		t.Fatal(err)
	}
	old := timeNowMinus(staleMS * 2)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(runDir, "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := Render(runDir, ""); err != nil {
			t.Errorf("render under stale lock: %v", err)
		}
	}()
	<-done
	if _, err := os.Stat(filepath.Join(runDir, "records", "render-shadow", "ledger.md")); err != nil {
		t.Errorf("render did not complete through a stale lock: %v", err)
	}
}
