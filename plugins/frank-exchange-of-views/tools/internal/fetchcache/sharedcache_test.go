package fetchcache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// THE SHARED CACHE IS SHARED, WHICH MEANS NO HOLDER MAY DELETE IT.
//
// It used to be torn down by whichever test binary finished first (#717). `go test ./...` runs
// package binaries concurrently, both this package and internal/cli acquire the same fixed
// directory, and the release did os.RemoveAll — so the loser's wazero compile lost its staging
// file mid-write:
//
//	could not compile webassembly module: open …/.wazero/…/1a47c473….tmp: no such file or directory
//
// Three unforced runs found it on three DIFFERENT OCR tests, which is why it read as a flaky test
// rather than as a fixed defect. Forced, it needs no luck at all: `go test ./internal/fetchcache
// -run XXXNothing` runs zero tests, and its TestMain teardown still deleted the directory a
// concurrent internal/cli run was compiling out of.
//
// The comment on UseSharedModuleCache had the safety argument almost right — "two concurrent test
// binaries sharing it can at worst make each other recompile; there is no state here to corrupt" —
// and it is true of SHARING and false of the teardown that sharing implies. Nothing is corrupted;
// the directory is simply removed out from under a live compile, and the loser does not recompile,
// it fails.

// TestAcquiringTheCacheDoesNotDestroyWhatAnotherHolderIsUsing guards the REAP, not the original
// defect, and the distinction is worth stating because the test name invites the other reading.
//
// #717 lived in the teardown, and no test fixes it: the release is GONE from this package's API,
// so there is nothing left to call that removes a directory another binary holds. That is
// enforcement at the write rather than a check at the read, and it is stronger than any assertion
// here could be.
//
// What this DOES pin is the new hazard the fix introduces. The reap is a destructive act on the
// same shared directory, and a reap that fired on a live cache would recreate #717 in a new place.
// A second binary starting up is a second acquire, and the first binary's compiled module has to
// survive it. Verified by mutation: dropping the age check fails this test. The sentinel stands in
// for the module so the test needs no 3,968 ms PDFium compile to ask the question.
func TestAcquiringTheCacheDoesNotDestroyWhatAnotherHolderIsUsing(t *testing.T) {
	if err := UseSharedModuleCache(); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	first := moduleCacheDir
	sentinel := filepath.Join(first, "stands-in-for-a-compiled-module")
	if err := os.WriteFile(sentinel, []byte("compiled"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(sentinel) })

	// The second test binary starts.
	if err := UseSharedModuleCache(); err != nil {
		t.Fatalf("second acquire: %v", err)
	}
	if moduleCacheDir != first {
		t.Errorf("second acquire moved the cache: %q -> %q; the whole point is that both hold ONE directory",
			first, moduleCacheDir)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("acquiring the shared cache destroyed what a live holder was using (%v) — this is #717", err)
	}
}

// TestAnIdleCacheIsReapedAndALiveOneIsNot pins both halves of the age rule, because a reap that
// never fires and a reap that fires on a live cache are the same code with one comparison wrong,
// and only one of them is loud.
func TestAnIdleCacheIsReapedAndALiveOneIsNot(t *testing.T) {
	now := time.Now()
	for _, c := range []struct {
		name      string
		lastUsed  time.Time
		wantReap  bool
		wantWhyOK string
	}{
		{
			name:      "idle past the threshold",
			lastUsed:  now.Add(-cacheIdleTTL - time.Hour),
			wantReap:  true,
			wantWhyOK: "nothing has acquired it for longer than any test binary lives",
		},
		{
			name:      "used a minute ago",
			lastUsed:  now.Add(-time.Minute),
			wantReap:  false,
			wantWhyOK: "a concurrent binary acquired it and may be compiling into it right now",
		},
		{
			// THE BOUNDARY, on the safe side of it. A cache used exactly at the threshold is one
			// whose holder could still be alive, and an off-by-one here is a reap that races.
			name:     "used exactly at the threshold",
			lastUsed: now.Add(-cacheIdleTTL),
			wantReap: false,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "cache")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(dir, "compiled")
			if err := os.WriteFile(sentinel, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(dir, c.lastUsed, c.lastUsed); err != nil {
				t.Fatal(err)
			}

			if got := reapIfIdle(dir, now); got != c.wantReap {
				t.Fatalf("reapIfIdle = %v, want %v (%s)", got, c.wantReap, c.wantWhyOK)
			}
			_, err := os.Stat(sentinel)
			if c.wantReap && err == nil {
				t.Error("reported a reap and left the contents behind")
			}
			if !c.wantReap && err != nil {
				t.Errorf("left the cache alone but its contents are gone: %v", err)
			}
		})
	}
}

// TestReapingIsSilentOnACacheThatIsNotThere — the first run on a machine, and every run after a
// /tmp sweep. A missing directory is the ordinary case, not a fault, and it must not be reported
// as a reap: a caller that logged it would say "removed a stale cache" about a machine that has
// never had one.
func TestReapingIsSilentOnACacheThatIsNotThere(t *testing.T) {
	if reapIfIdle(filepath.Join(t.TempDir(), "never-existed"), time.Now()) {
		t.Error("reported reaping a directory that does not exist")
	}
}

// TestAcquiringTouchesTheCacheSoALiveHolderKeepsItAlive is the other half of the age rule, and the
// half a reader is most likely to assume rather than check: reading a compiled module does not
// update the directory's mtime, so without an explicit touch a cache in continuous use across a
// long suite would age into the reap window and be deleted by the next binary to start.
//
// IT BACKDATES TO JUST INSIDE THE THRESHOLD, and that is the whole design of the test. The obvious
// version backdates PAST it — and then passes whether or not the touch exists, because the reap
// fires, MkdirAll builds a fresh directory, and a fresh directory has a fresh mtime. Written that
// way it cannot distinguish "touched" from "deleted and rebuilt", which is to say it cannot fail.
// Caught by mutation: deleting the Chtimes left that version green.
//
// Inside the threshold nothing is reaped, so the mtime can only have moved if something moved it,
// and the surviving sentinel says the directory is the same one.
func TestAcquiringTouchesTheCacheSoALiveHolderKeepsItAlive(t *testing.T) {
	if err := UseSharedModuleCache(); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	dir := moduleCacheDir
	sentinel := filepath.Join(dir, "stands-in-for-a-compiled-module-2")
	if err := os.WriteFile(sentinel, []byte("compiled"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(sentinel) })

	aging := time.Now().Add(-cacheIdleTTL + time.Hour)
	if err := os.Chtimes(dir, aging, aging); err != nil {
		t.Fatalf("backdate: %v", err)
	}

	if err := UseSharedModuleCache(); err != nil {
		t.Fatalf("re-acquire: %v", err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("the cache was rebuilt, not touched — this test can no longer measure the touch: %v", err)
	}
	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat after re-acquire: %v", err)
	}
	if age := time.Since(fi.ModTime()); age > time.Minute {
		t.Errorf("acquire left the cache looking %v old; a live holder would age into the reap window and "+
			"be deleted by the next binary to start", age)
	}
}
