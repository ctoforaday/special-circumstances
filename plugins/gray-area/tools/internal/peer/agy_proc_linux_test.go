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
		busy, why := otherDBHolder(root, conv, myPID+1)
		if !busy || !strings.Contains(why, fmt.Sprintf("pid %d", myPID)) {
			t.Errorf("otherDBHolder(%q, %q, %d) = (%v, %q), want (true, naming pid %d)", root, conv, myPID+1, busy, why, myPID)
		}
		if busy, why := otherDBHolder(root, conv, myPID); busy {
			t.Errorf("otherDBHolder excluding holder pid %d = (%v, %q), want (false, \"\")", myPID, busy, why)
		}
	}

	// Closing the file clears the busy state.
	f.Close()
	if busy, why := otherDBHolder(realRoot, conv, myPID+1); busy {
		t.Errorf("after closing db, otherDBHolder = (%v, %q), want false", busy, why)
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
	if busy, why := ConversationBusy(agyRoot, conv, myPID); busy {
		t.Errorf("ConversationBusy with holder pid %d excluded = (true, %q), want (false, \"\")", myPID, why)
	}
	if busy, why := ConversationBusy(agyRoot, conv, myPID+1); !busy || !strings.Contains(why, "presence/"+conv+".lock") {
		t.Errorf("ConversationBusy with other pid excluded = (%v, %q), want (true, naming presence/%s.lock)", busy, why, conv)
	}
}
