//go:build unix

package peer

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestConversationBusyDetectsHeldPresenceLock(t *testing.T) {
	agyRoot := t.TempDir()
	lockDir := filepath.Join(agyRoot, "presence")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const conv = "conv-lock-1"
	lockPath := filepath.Join(lockDir, conv+".lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Unlocked file is not busy.
	if busy, why := ConversationBusy(agyRoot, conv, 0); busy {
		t.Fatalf("unlocked file reported busy: %s", why)
	}

	// Acquire an exclusive flock on this open file description; ConversationBusy opens a separate
	// file description and must see EWOULDBLOCK.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("flock: %v", err)
	}
	busy, why := ConversationBusy(agyRoot, conv, 0)
	if !busy || !strings.Contains(why, "presence/"+conv+".lock") {
		t.Errorf("locked file: ConversationBusy = (%v, %q), want (true, naming presence/%s.lock)", busy, why, conv)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	if busy, why := ConversationBusy(agyRoot, conv, 0); busy {
		t.Errorf("after unlock: ConversationBusy = (%v, %q), want false", busy, why)
	}
}
