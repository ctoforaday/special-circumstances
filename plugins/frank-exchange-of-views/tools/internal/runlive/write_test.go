package runlive

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// A CONCURRENT READER NEVER SEES A PARTIAL MARKER (#1552). The marker is polled on every hook call
// and every dashboard tick while setup and capture rewrite it. An in-place write truncates first,
// and a read inside that window found an empty or half-written file: readRunLive called it
// unreadable, the dashboard watcher took that as the run's row gone and tore down a live run's
// server, and InferRunDir reported a broken marker.
func TestAConcurrentReaderNeverSeesAPartialMarker(t *testing.T) {
	proj := t.TempDir()
	t0 := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	if _, err := WriteRunLiveMarker(proj, "research/base", nil, t0, "", ""); err != nil {
		t.Fatal(err)
	}
	var stop atomic.Bool
	var wg sync.WaitGroup
	var unreadable, reads atomic.Int64
	// What each failed read got: an open error, empty or torn bytes, or a file with no rows.
	var kindsMu sync.Mutex
	kinds := map[string]int{}
	marker := filepath.Join(proj, ".claude", "run-live.json")
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !stop.Load() {
				reads.Add(1)
				// readRunLive's own steps — shared lock, read, decode — kept apart so a failure says
				// which one failed.
				kind := ""
				release, lerr := rlockMarker(proj)
				b, err := readSharingDelete(marker)
				if lerr == nil {
					release()
				}
				if lerr != nil {
					kind = "lock: " + lerr.Error()
				} else if err != nil {
					kind = "open/read: " + err.Error()
				} else if runs, readable := decodeRunLive(b); !readable {
					kind = fmt.Sprintf("unparseable, %d bytes", len(b))
				} else if len(runs) == 0 {
					kind = "no rows"
				}
				if kind != "" {
					unreadable.Add(1)
					kindsMu.Lock()
					kinds[kind]++
					kindsMu.Unlock()
				}
			}
		}()
	}
	// Rows of differing lengths, so a torn write cannot happen to parse.
	var werr error
	for i := 0; i < 400 && werr == nil; i++ {
		run := fmt.Sprintf("research/run-%d", i%7)
		pins := make([]string, i%13)
		for j := range pins {
			pins[j] = fmt.Sprintf("ideas/pin-%d-%d.md", i, j)
		}
		if _, werr = WriteRunLiveMarker(proj, run, pins, t0, "", ""); werr == nil {
			_, _, werr = RemoveRunLiveEntry(proj, run)
		}
	}
	stop.Store(true)
	wg.Wait()
	if werr != nil {
		t.Fatalf("a write failed while readers polled: %v", werr)
	}
	if n := unreadable.Load(); n != 0 {
		t.Fatalf("%d of %d concurrent reads found the marker unreadable or empty — a reader saw a partial write; the reads got %v", n, reads.Load(), kinds)
	}
	// No temporary file is left beside the marker and its lock.
	entries, err := os.ReadDir(filepath.Join(proj, ".claude"))
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 2 || names[0] != "run-live.json" || names[1] != MarkerLock {
		t.Fatalf(".claude holds %v, want run-live.json and %s", names, MarkerLock)
	}
}

// A FAILED REMOVAL OR WRITE IS REPORTED, never dropped (#1552). It used to be discarded, so a
// capture whose removal failed reported the run closed while its row stayed live. The marker path is
// made a non-empty directory: neither a remove nor a rename can take it, on any platform.
func TestAFailedMarkerWriteIsReported(t *testing.T) {
	blocked := func(t *testing.T) string {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "run-live.json", "occupant"), 0o755); err != nil {
			t.Fatal(err)
		}
		return proj
	}
	t.Run("remove", func(t *testing.T) {
		if _, _, err := RemoveRunLiveEntry(blocked(t), "research/x"); err == nil {
			t.Fatal("a removal that failed reported success")
		}
	})
	t.Run("write", func(t *testing.T) {
		if _, err := WriteRunLiveMarker(blocked(t), "research/x", nil, time.Now(), "", ""); err == nil {
			t.Fatal("a write that failed reported success")
		}
	})
}

// A READER WAITS OUT A WRITER. A writer holds MarkerLock exclusive from its read to its rename;
// readRunLive takes it shared, so its open can never land inside the rename, where on Windows it
// finds no file. While the lock is held the read does not return; once released it sees the marker.
func TestAReaderWaitsForAWriterHoldingTheMarkerLock(t *testing.T) {
	proj := t.TempDir()
	if _, err := WriteRunLiveMarker(proj, "research/x", nil, time.Now(), "", ""); err != nil {
		t.Fatal(err)
	}
	release, err := lockMarker(proj)
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		runs     []RunLiveMarker
		readable bool
	}
	done := make(chan result, 1)
	go func() {
		runs, readable := readRunLive(proj)
		done <- result{runs, readable}
	}()
	select {
	case <-done:
		release()
		t.Fatal("readRunLive returned while a writer held the marker lock")
	case <-time.After(200 * time.Millisecond):
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-done:
		if !r.readable || len(r.runs) != 1 {
			t.Fatalf("after the writer released, readRunLive saw %+v", r)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("readRunLive did not return after the writer released the lock")
	}
}
