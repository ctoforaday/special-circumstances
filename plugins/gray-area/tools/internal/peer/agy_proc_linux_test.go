//go:build linux

package peer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	myPID := os.Getpid()
	// Any non-excluded process holding the DB open (here the Go test binary itself, whose comm is
	// not agy/antigravity) must be detected through both realRoot and symRoot.
	for _, root := range []string{realRoot, symRoot} {
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
