//go:build unix

package peer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofrs/flock"
)

func TestConversationBusyDetectsHeldPresenceLock(t *testing.T) {
	agyRoot := t.TempDir()
	lockDir := filepath.Join(agyRoot, "presence")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const conv = "conv-lock-1"
	lockPath := filepath.Join(lockDir, conv+".lock")
	if err := os.WriteFile(lockPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	// Unlocked file is not busy.
	if busy, why := ConversationBusy(agyRoot, conv, 0); busy {
		t.Fatalf("unlocked file reported busy: %s", why)
	}

	// Acquire an exclusive lock on a separate open file description; ConversationBusy must see it held.
	fl := flock.New(lockPath, flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))
	ok, err := fl.TryLock()
	if err != nil || !ok {
		t.Fatalf("TryLock = (%v, %v), want (true, nil)", ok, err)
	}
	defer fl.Unlock()

	busy, why := ConversationBusy(agyRoot, conv, 0)
	if !busy || !strings.Contains(why, "presence/"+conv+".lock") {
		t.Errorf("locked file: ConversationBusy = (%v, %q), want (true, naming presence/%s.lock)", busy, why, conv)
	}

	if err := fl.Unlock(); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if busy, why := ConversationBusy(agyRoot, conv, 0); busy {
		t.Errorf("after unlock: ConversationBusy = (%v, %q), want false", busy, why)
	}
}

func TestConversationBusyFailsClosedOnUnreadableLock(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses 0000 file permissions")
	}
	agyRoot := t.TempDir()
	lockDir := filepath.Join(agyRoot, "presence")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const conv = "conv-lock-unreadable"
	lockPath := filepath.Join(lockDir, conv+".lock")
	if err := os.WriteFile(lockPath, nil, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(lockPath, 0o600) })

	busy, why := ConversationBusy(agyRoot, conv, 0)
	if !busy || !strings.Contains(why, "cannot inspect presence/"+conv+".lock") {
		t.Errorf("unreadable lock: ConversationBusy = (%v, %q), want fail-closed (true, cannot inspect)", busy, why)
	}
}
