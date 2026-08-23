package record

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
	runDir := newRun(t)
	const seats = 6
	const perSeat = 4

	var wg sync.WaitGroup
	errs := make(chan error, seats*perSeat*3)
	for s := 1; s <= seats; s++ {
		wg.Add(1)
		go func(s int) {
			defer wg.Done()
			seatID := fmt.Sprintf("red-lens-r1-L%d", s)
			if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: seatID, Round: RoundIn(runDir)(seatID)}, ""); err != nil {
				errs <- err
				return
			}
			for i := 0; i < perSeat; i++ {
				f := &recordpb.Finding{
					Label:      proto.String(fmt.Sprintf("L%d-F%d", s, i)),
					Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
					Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
					Impact:     recordtest.P(recordpb.Grade_GRADE_HIGH),
					Text:       proto.String(strings.Repeat("finding prose ", 20)),
				}
				if _, err := Append(Identity{RunDir: runDir, SeatID: seatID, Round: RoundOf(seatID)}, f); err != nil {
					errs <- err
					continue
				}
				// A concurrent READER (the replay every projection now runs on demand)
				// racing the appenders: no write may be lost to a racing read.
				if _, err := BoardState(runDir); err != nil {
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

	// EVERY EVENT LANDED. This asked whether a shard lost a write to a racing render; it now asks
	// whether a TRANSACTION did, which is the same question against the mechanism that replaced
	// the one the test was written for — and a stronger one, because the writers share a file
	// rather than each owning their own.
	m, err := MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	findings := 0
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_FINDING {
			findings++
		}
	}
	if want := seats * perSeat; findings != want {
		t.Errorf("findings landed = %d, want %d", findings, want)
	}

	// No lock or temp artifact leaked: a stuck .lock- directory would deadlock the
	// next seat for the full bounded wait, and a stray .tmp- would be read as a
	// projection by anything globbing the shadow dir.
	entries, err := os.ReadDir(filepath.Join(runDir, "records"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		// A leftover .lock- FILE is expected and harmless under flock: the lock
		// lives in the kernel, released when the holder exits, and the file is
		// only a handle to lock ON. Under the old mkdir scheme a leftover
		// .lock- DIRECTORY *was* the lock, so its presence blocked the next seat
		// for the full bounded wait — which is why this assertion existed and why
		// it now only guards temp files.
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leaked temp artifact: %s", e.Name())
		}
	}
}

// TestAbandonedLockFileDoesNotBlock: with flock, the lock is the kernel's, not
// the file's. A lock FILE left on disk by a dead holder carries no lock, so the
// next seat acquires immediately — the case the old mkdir implementation had to
// guess about with a ten-second staleness timeout, and could get wrong in both
// directions.
func TestAbandonedLockFileDoesNotBlock(t *testing.T) {
	runDir := newRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}, ""); err != nil {
		t.Fatal(err)
	}
	// An empty lock file for the per-seat pointer lock an append acquires, as a crashed
	// holder would leave behind. Under flock the file carries no lock, so the next
	// append acquires immediately rather than serving the full bounded wait.
	if err := os.WriteFile(filepath.Join(runDir, "records", ".lock-ptr-red-merge-r1"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Finding{Label: proto.String("F1"), Text: proto.String("over an abandoned lock")}); err != nil {
		t.Fatalf("append over an abandoned lock file: %v", err)
	}
	if elapsed := time.Since(start); elapsed > lockWait {
		t.Errorf("append waited %v on an unheld lock — the bounded wait was served instead of acquiring", elapsed)
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_FINDING {
			got++
		}
	}
	if got != 1 {
		t.Errorf("the append did not land: %d finding events", got)
	}
}
