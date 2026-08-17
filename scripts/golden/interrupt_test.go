package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// THE PROMISE THESE HOLD. -review materialises proposals and takes back everything it was not
// told to keep. Between those moments the tree is rewritten, and #432 friction 4 measured what
// a signal in that window did: `golden -review | head -200` left THIRTEEN rewritten goldens
// behind and exited 141, under a tool whose own result line reads "the tree is exactly as you
// found it".
//
// These tests hold the two halves of the fix separately, because they fail separately: the
// restore must take back what the run created, and it must NOT take back what the run was told
// to keep.

func resetInterruptState(t *testing.T) {
	t.Helper()
	interrupted.mu.Lock()
	interrupted.got, interrupted.cmd = nil, nil
	interrupted.mu.Unlock()
	held.mu.Lock()
	held.root, held.props = "", nil
	held.mu.Unlock()
}

// A run that materialised proposals and then died must put every one of them back — the
// modified ones to their committed content, the new ones removed entirely.
func TestRestoreTakesBackEveryHeldProposal(t *testing.T) {
	resetInterruptState(t)
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "a proposal this run wrote\n")
	write(t, dir, "testdata/new.golden", "a new golden this run wrote\n")

	hold(dir, []proposal{
		{path: "testdata/one.golden"},
		{path: "testdata/new.golden", isNew: true},
	})

	if n := restoreHeld(); n != 2 {
		t.Fatalf("restored %d, want 2", n)
	}
	b, err := os.ReadFile(filepath.Join(dir, "testdata/one.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "recorded\n" {
		t.Errorf("tracked golden = %q, want the committed content back", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "testdata/new.golden")); !os.IsNotExist(err) {
		t.Error("a new golden the run created must be gone after a restore")
	}
}

// THE HALF THAT WOULD BE WORSE THAN THE BUG. The restore runs from a defer, so it fires on the
// SUCCESS path too. If release() did not clear the claim, an accepted golden would be reverted
// out from under the reviewer who had just named it — turning a crash-safety fix into silent
// data loss on the one path that matters.
func TestReleaseStopsTheRestoreFromUndoingAnAccept(t *testing.T) {
	resetInterruptState(t)
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "the change the reviewer accepted\n")

	hold(dir, []proposal{{path: "testdata/one.golden"}})
	release()

	if n := restoreHeld(); n != 0 {
		t.Fatalf("restored %d after release; an accepted golden must survive", n)
	}
	b, err := os.ReadFile(filepath.Join(dir, "testdata/one.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "the change the reviewer accepted\n" {
		t.Errorf("golden = %q, want the accepted content kept", b)
	}
}

// Idempotent, because it runs from a defer on paths that already restored explicitly. A second
// call reverting a second time would fight whatever came after the first.
func TestRestoreIsIdempotent(t *testing.T) {
	resetInterruptState(t)
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "a proposal\n")
	hold(dir, []proposal{{path: "testdata/one.golden"}})

	if n := restoreHeld(); n != 1 {
		t.Fatalf("first restore returned %d, want 1", n)
	}
	if n := restoreHeld(); n != 0 {
		t.Errorf("second restore returned %d; the claim must be cleared by the first", n)
	}
}

// A restore that cannot put ONE file back must still put the other twelve back. Modelled with a
// path that does not exist, which is what a half-written tree looks like.
func TestRestoreKeepsGoingPastAFileItCannotPutBack(t *testing.T) {
	resetInterruptState(t)
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "a proposal\n")
	hold(dir, []proposal{
		{path: "testdata/gone.golden", isNew: true}, // never existed; removing it fails
		{path: "testdata/one.golden"},
	})

	if n := restoreHeld(); n != 1 {
		t.Fatalf("restored %d, want 1 — the reachable file must still be restored", n)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "testdata/one.golden"))
	if string(b) != "recorded\n" {
		t.Errorf("golden = %q; one failure must not strand the rest", b)
	}
}

// THE CALLER'S VIEW MUST NOT CHANGE. The fix makes SIGPIPE non-fatal so the restore can run,
// which means the process now chooses its own exit code where the kernel used to. 128+signal is
// what a shell reports for a signalled process, so `| head` still sees 141 and a script that
// keys on it keeps working.
func TestStoppedReportsTheShellsExitCodeForTheSignal(t *testing.T) {
	for _, c := range []struct {
		sig  syscall.Signal
		want int
	}{
		{syscall.SIGPIPE, 141},
		{syscall.SIGINT, 130},
		{syscall.SIGTERM, 143},
	} {
		resetInterruptState(t)
		interrupted.mu.Lock()
		interrupted.got = c.sig
		interrupted.mu.Unlock()

		code, yes := stopped()
		if !yes {
			t.Errorf("%v: stopped() said no", c.sig)
			continue
		}
		if code != c.want {
			t.Errorf("%v: exit code %d, want %d", c.sig, code, c.want)
		}
		if signalName() == "" {
			t.Errorf("%v: the message has no signal to name", c.sig)
		}
	}
}

// No signal is the ordinary case and must not be reported as one, or every clean run would exit
// non-zero.
func TestStoppedIsSilentWithoutASignal(t *testing.T) {
	resetInterruptState(t)
	if code, yes := stopped(); yes || code != 0 {
		t.Errorf("stopped() = (%d, %v) with no signal; want (0, false)", code, yes)
	}
	if n := signalName(); n != "" {
		t.Errorf("signalName() = %q with no signal, want empty", n)
	}
}

// The restore must only ever touch what THIS run created. proposalsBetween already excludes
// pre-existing dirt, and this pins the two together: a file the author had modified before the
// run is not in the held set, so no crash path can revert their work.
func TestRestoreNeverClaimsWorkTheRunDidNotCreate(t *testing.T) {
	resetInterruptState(t)
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "the author's own uncommitted edit\n")

	before, err := readGoldenState(dir)
	if err != nil {
		t.Fatal(err)
	}
	write(t, dir, "testdata/two.golden", "a proposal this run wrote\n")
	after, err := readGoldenState(dir)
	if err != nil {
		t.Fatal(err)
	}

	p := proposalsBetween(before, after)
	hold(dir, p)
	restoreHeld()

	b, err := os.ReadFile(filepath.Join(dir, "testdata/one.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "the author's own uncommitted edit") {
		t.Errorf("the restore reverted an edit it did not make: %q", b)
	}
}
