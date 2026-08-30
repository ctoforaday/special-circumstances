package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// mirrorFixture puts a sandboxed HOME in place and returns the mirror root under it.
func mirrorFixture(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root, err := record.MirrorRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// THE RUN CAPTURE JUST FINISHED KEEPS ITS MIRROR. This is the safety property, not a detail:
// capture writes an UNTRACKED tarball and does not commit, so when it returns, the record's
// only durable copy does not exist yet. A reap that took this run's mirror would remove the
// recovery path at the moment its replacement is most exposed.
func TestTheReapSparesTheRunThatJustFinished(t *testing.T) {
	mirrorFixture(t)
	thisRun, err := record.MirrorDir("/home/dev/research/2026-08-29_the-run-being-captured")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(thisRun, 0o755); err != nil {
		t.Fatal(err)
	}

	if line := reapOrphanMirrors(time.Now()); line != "" {
		t.Errorf("the reap reported %q on a board holding only the live run", line)
	}
	if _, err := os.Stat(thisRun); err != nil {
		t.Fatalf("capture reaped the mirror of the run it just captured — the record's only "+
			"committed copy does not exist yet at this point: %v", err)
	}
}

// An orphan — a run that stopped writing weeks ago — is what this exists to take.
func TestTheReapTakesOrphansAndSaysSo(t *testing.T) {
	root := mirrorFixture(t)
	orphan := filepath.Join(root, "aaaaaaaaaaaa")
	live := filepath.Join(root, "bbbbbbbbbbbb")
	for _, d := range []string{orphan, live} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-(mirrorOrphanDays + 10) * 24 * time.Hour)
	if err := os.Chtimes(orphan, old, old); err != nil {
		t.Fatal(err)
	}

	line := reapOrphanMirrors(time.Now())
	if !strings.Contains(line, "1 stale checkpoint mirror(s) removed") {
		t.Errorf("the reap reported %q, want it to name the one mirror it took", line)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Error("the orphan survived")
	}
	if _, err := os.Stat(live); err != nil {
		t.Errorf("a fresh mirror was taken: %v", err)
	}
}

// A REAP THAT COULD NOT RUN SAYS SO, rather than returning the silence a clean board returns.
// With no home there is no mirror root to read, and "nothing stale" would be a measurement
// nobody made.
func TestTheReapIsLoudWhenItCouldNotRun(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	line := reapOrphanMirrors(time.Now())
	if !strings.HasPrefix(line, "mirror purge: NOT RUN") {
		t.Errorf("with no resolvable home the reap reported %q — a zero that did not check "+
			"reads exactly like a zero that found nothing", line)
	}
}
