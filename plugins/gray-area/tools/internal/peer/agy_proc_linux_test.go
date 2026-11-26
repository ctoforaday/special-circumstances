//go:build linux

package peer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofrs/flock"
)

func TestOtherDBHolderDetectsOpenDBAndHandlesSymlinkedRoot(t *testing.T) {
	realRoot := t.TempDir()
	const conv = "conv-db-holder-1"
	convDir := filepath.Join(realRoot, "conversations")
	if err := os.MkdirAll(convDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(convDir, conv+".db")
	f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Create a symlink pointing to realRoot to verify EvalSymlinks canonicalisation.
	linkParent := t.TempDir()
	symRoot := filepath.Join(linkParent, "agy-link")
	if err := os.Symlink(realRoot, symRoot); err != nil {
		t.Fatal(err)
	}

	// Also verify a symlinked conversations/ directory inside a non-symlinked root.
	symConvRoot := t.TempDir()
	if err := os.Symlink(convDir, filepath.Join(symConvRoot, "conversations")); err != nil {
		t.Fatal(err)
	}

	myPID := os.Getpid()
	// Any non-excluded process holding the DB open (here the Go test binary itself, whose comm is
	// not agy/antigravity) must be detected through realRoot, symRoot, and symConvRoot.
	for _, root := range []string{realRoot, symRoot, symConvRoot} {
		h := otherDBHolder(root, conv, myPID+1)
		if h.Kind != HolderDBOpen || h.PID != myPID || !strings.Contains(h.Detail, fmt.Sprintf("pid %d", myPID)) {
			t.Errorf("otherDBHolder(%q, %q, %d) = %+v, want the db open in pid %d", root, conv, myPID+1, h, myPID)
		}
		if h := otherDBHolder(root, conv, myPID); h.Kind != HolderNone {
			t.Errorf("otherDBHolder excluding holder pid %d = %+v, want none", myPID, h)
		}
	}

	// Closing the file clears the busy state.
	f.Close()
	if h := otherDBHolder(realRoot, conv, myPID+1); h.Kind != HolderNone {
		t.Errorf("after closing db, otherDBHolder = %+v, want none", h)
	}
}

func TestConversationBusyAllowsLockHeldOnlyByExcludedPID(t *testing.T) {
	agyRoot := t.TempDir()
	lockDir := filepath.Join(agyRoot, "presence")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const conv = "conv-lock-own-pid"
	lockPath := filepath.Join(lockDir, conv+".lock")
	fl := flock.New(lockPath, flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))
	ok, err := fl.TryLock()
	if err != nil || !ok {
		t.Fatalf("TryLock = (%v, %v), want (true, nil)", ok, err)
	}
	defer fl.Unlock()

	myPID := os.Getpid()
	if h := ConversationBusy(agyRoot, conv, myPID); h.Kind != HolderNone {
		t.Errorf("ConversationBusy with holder pid %d excluded = %+v, want none", myPID, h)
	}
	if h := ConversationBusy(agyRoot, conv, myPID+1); h.Kind != HolderPresenceLock || h.PID != myPID {
		t.Errorf("ConversationBusy with other pid excluded = %+v, want the presence lock held by pid %d", h, myPID)
	}
}

// A PRESENCE-LOCK HOLDER IS NAMED BY ITS PID FROM /proc/locks, and a waiter is not a holder.
func TestLockHolderReadsProcLocks(t *testing.T) {
	const locks = `1: POSIX  ADVISORY  WRITE 900 00:2a:77 0 EOF
2: FLOCK  ADVISORY  WRITE 4242 08:01:1234 0 EOF
2: -> FLOCK  ADVISORY  WRITE 5151 08:01:1234 0 EOF
3: FLOCK  ADVISORY  WRITE 777 08:01:9999 0 EOF
`
	for _, tc := range []struct {
		name    string
		locks   string
		exclude int
		want    int
		wantErr bool
	}{
		{"holder named", locks, 0, 4242, false},
		{"holder excluded", locks, 4242, 0, false},
		{"holder outside this pid namespace", "2: FLOCK  ADVISORY  WRITE 0 08:01:1234 0 EOF\n", 0, 0, true},
		{"no holder listed", "3: FLOCK  ADVISORY  WRITE 777 08:01:9999 0 EOF\n", 0, 0, true},
	} {
		got, err := lockHolder(tc.locks, "1234", tc.exclude)
		if got != tc.want || (err != nil) != tc.wantErr {
			t.Errorf("%s: lockHolder = %d, %v; want %d, error %v", tc.name, got, err, tc.want, tc.wantErr)
		}
	}
}
