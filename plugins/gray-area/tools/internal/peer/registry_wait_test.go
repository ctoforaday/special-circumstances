package peer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

// TestMain gives the registry lock a wait no loaded box reaches, in this process and in the helper
// processes the concurrency tests start from this binary: those tests measure exclusion, not the
// bound. TestRegistryLockWaitIsBounded sets its own.
func TestMain(m *testing.M) {
	registryLockWait = time.Minute
	os.Exit(m.Run())
}

// A HOLDER THAT NEVER LETS GO DOES NOT HANG A VERB. With the registry lock held elsewhere, Prune and
// RegisterWith each give up after the bound and say why.
func TestRegistryLockWaitIsBounded(t *testing.T) {
	rt := t.TempDir()
	dir := RegistryDir(rt)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	holder := flock.New(filepath.Join(dir, ".register.lock"), flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))
	if ok, err := holder.TryLock(); err != nil || !ok {
		t.Fatalf("TryLock = (%v, %v), want (true, nil)", ok, err)
	}
	defer holder.Unlock()
	registryLockWait = 100 * time.Millisecond
	defer func() { registryLockWait = time.Minute }()

	for _, op := range []struct {
		name string
		run  func() error
	}{
		{"Prune", func() error { _, err := Prune(rt, stressLiveness); return err }},
		{"RegisterWith", func() error {
			_, _, err := RegisterWith(rt, agyShell("conv-1", "localhost:1", "tok"), hostID(4242, 1, false), at)
			return err
		}},
	} {
		done := make(chan error, 1)
		go func() { done <- op.run() }()
		select {
		case err := <-done:
			if err == nil || !strings.Contains(err.Error(), "for more than 100ms") {
				t.Errorf("%s with the lock held elsewhere: err = %v, want the bound named", op.name, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("%s still waits for the registry lock after 10s", op.name)
		}
	}
}
