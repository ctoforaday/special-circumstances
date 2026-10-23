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
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !stop.Load() {
				reads.Add(1)
				if runs, readable := readRunLive(proj); !readable || len(runs) == 0 {
					unreadable.Add(1)
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
		t.Fatalf("%d of %d concurrent reads found the marker unreadable or empty — a reader saw a partial write", n, reads.Load())
	}
	// No temporary file is left beside the marker.
	entries, err := os.ReadDir(filepath.Join(proj, ".claude"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "run-live.json" {
		names := []string{}
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf(".claude holds %v, want run-live.json alone", names)
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
