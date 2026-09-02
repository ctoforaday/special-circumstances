package recordsql

import (
	"path/filepath"
	"sync"
	"testing"
)

// TestConcurrentOpenOnFreshDatabase is the gate #557 said had to arrive with its fix.
//
// `Open` reads whether `events` exists and applies the schema if it does not. With nothing
// between the two, two openers of a FRESH database both saw zero and both applied; the second
// failed with "table … already exists". The issue records why nothing ever hit it, and both
// reasons are accidents rather than guarantees:
//
//   - the in-process handle cache means one process cannot race itself, so this test calls
//     openUncached — the cache is exactly what has to be bypassed to ask the question;
//   - the cross-process test seeds the schema before spawning its children, so the racing
//     branch is unreachable there by construction.
//
// Production was covered by the same accident: `setup` creates the run directory and its
// database before any seat is dispatched. The fuzz removes it — it builds its run directory
// directly, so the first seat command creates the schema, and with concurrent lanes (#630)
// that is several processes at once on round 0 of every run.
func TestConcurrentOpenOnFreshDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fresh.db")

	// SIX, not two. Two openers can pass by luck of scheduling; the failure this guards is a
	// window, and more racers is more of the window sampled per run.
	const openers = 6
	var wg sync.WaitGroup
	errs := make([]error, openers)
	start := make(chan struct{})
	for i := range openers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start // released together, so they contend rather than queue politely
			db, err := openUncached(path)
			errs[i] = err
			if db != nil {
				_ = db.Close()
			}
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("opener %d failed on a fresh database: %v", i, err)
		}
	}

	// AND THE SCHEMA IS USABLE, not merely present. A losing opener that swallowed its error
	// would satisfy the loop above while handing back a database nothing can write.
	db, err := openUncached(path)
	if err != nil {
		t.Fatalf("reopening after the race: %v", err)
	}
	defer db.Close()
	if !hasEvents(db) {
		t.Fatal("no events table after six concurrent opens — every opener thought another had made it")
	}
}
