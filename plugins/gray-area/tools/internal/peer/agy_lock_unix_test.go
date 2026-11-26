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
	if h := ConversationBusy(agyRoot, conv, 0); h.Kind != HolderNone {
		t.Fatalf("unlocked file reported held: %+v", h)
	}

	// Acquire an exclusive lock on a separate open file description; ConversationBusy must see it held.
	fl := flock.New(lockPath, flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))
	ok, err := fl.TryLock()
	if err != nil || !ok {
		t.Fatalf("TryLock = (%v, %v), want (true, nil)", ok, err)
	}
	defer fl.Unlock()

	// The holder is named by its pid, read from /proc/locks: this test process.
	h := ConversationBusy(agyRoot, conv, 0)
	if h.Kind != HolderPresenceLock || h.PID != os.Getpid() || !strings.Contains(h.Detail, "presence/"+conv+".lock") {
		t.Errorf("locked file: ConversationBusy = %+v, want the presence lock held by pid %d", h, os.Getpid())
	}

	if err := fl.Unlock(); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if h := ConversationBusy(agyRoot, conv, 0); h.Kind != HolderNone {
		t.Errorf("after unlock: ConversationBusy = %+v, want none", h)
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

	h := ConversationBusy(agyRoot, conv, 0)
	if h.Kind != HolderUninspectable || h.PID != 0 || !strings.Contains(h.Detail, "cannot inspect presence/"+conv+".lock") {
		t.Errorf("unreadable lock: ConversationBusy = %+v, want fail-closed (uninspectable, cannot inspect)", h)
	}
}
