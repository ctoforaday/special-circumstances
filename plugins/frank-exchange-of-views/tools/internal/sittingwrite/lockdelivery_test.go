package sittingwrite

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
)

// A DELIVERY ANOTHER CALL HOLDS IS NOT A DELIVERY THAT FAILED: a second lockDelivery on a held path
// returns hookgate.ErrDeliveryHeld, naming the lock; a lock that cannot be taken at all is a real
// error and is not it.
func TestLockDeliveryOnAHeldPathIsHeld(t *testing.T) {
	prev := lockWait
	lockWait = 50 * time.Millisecond
	t.Cleanup(func() { lockWait = prev })

	path := filepath.Join(t.TempDir(), deliveryDir, "a1-3.json")
	unlock, err := lockDelivery(path)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	_, err = lockDelivery(path)
	if !errors.Is(err, hookgate.ErrDeliveryHeld) {
		t.Fatalf("a second lock on a held path returned %v — want hookgate.ErrDeliveryHeld", err)
	}
	if !strings.Contains(err.Error(), path+".lock") {
		t.Errorf("the held error does not name the lock: %v", err)
	}

	// A regular file where the directory should be: the lock cannot be taken, and nothing holds it.
	parent := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(parent, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = lockDelivery(filepath.Join(parent, "a1-3.json"))
	if err == nil || errors.Is(err, hookgate.ErrDeliveryHeld) {
		t.Errorf("a lock that cannot be taken returned %v — want an error that is not ErrDeliveryHeld", err)
	}
}
