package sittingwrite

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

// deliveryHolderEnv, set, turns this test binary into a delivery-lock holder (see TestMain).
const deliveryHolderEnv = "SITTINGWRITE_TEST_DELIVERY_HOLDER"

// CALLS RETURNING TOGETHER TAKE THE DELIVERY IN TURN. Each call reads what was handed, then
// rewrites it; without the lock two calls read the same state and one rewrite is lost. Here each
// holder reads a counter and writes it back one higher: every increment survives only if no two
// holders overlap. The lock file stays after the last release.
func TestCallsReturningTogetherTakeTheDeliveryInTurn(t *testing.T) {
	path := filepath.Join(t.TempDir(), deliveryDir, "lens-1.json")
	const n = 16
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock, err := lockDelivery(path)
			if err != nil {
				errs <- err
				return
			}
			defer unlock()
			b, err := os.ReadFile(path)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				errs <- err
				return
			}
			c, _ := strconv.Atoi(string(b))
			time.Sleep(time.Millisecond)
			if err := os.WriteFile(path, []byte(strconv.Itoa(c+1)), 0o644); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("lockDelivery: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != strconv.Itoa(n) {
		t.Errorf("counter = %s after %d locked increments; two holders overlapped", b, n)
	}
	if _, err := os.Stat(path + ".lock"); err != nil {
		t.Errorf("the lock file must stay after release: %v", err)
	}
}

// A CALL KILLED WHILE HOLDING THE DELIVERY DOES NOT BLOCK THE NEXT ONE. The holder is another
// process, as a hook is: while it lives the lock is taken; once it is killed, the operating system
// releases the lock and the next call takes it at once.
func TestAKilledHolderDoesNotBlockTheNextDelivery(t *testing.T) {
	path := filepath.Join(t.TempDir(), deliveryDir, "lens-1.json")
	holder := exec.Command(os.Args[0], "-test.run=^$")
	holder.Env = append(os.Environ(), deliveryHolderEnv+"="+path)
	holder.Stderr = os.Stderr
	out, err := holder.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := holder.Start(); err != nil {
		t.Fatal(err)
	}
	killed := false
	kill := func() {
		if !killed {
			killed = true
			_ = holder.Process.Kill()
			_ = holder.Wait()
		}
	}
	t.Cleanup(kill)
	ready := make(chan string, 1)
	go func() {
		line, _ := bufio.NewReader(out).ReadString('\n')
		ready <- strings.TrimSpace(line)
	}()
	select {
	case line := <-ready:
		if line != "held" {
			t.Fatalf("the holder process did not take the lock: said %q", line)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the holder process never said it held the lock")
	}

	probe := flock.New(path+".lock", flock.SetFlag(os.O_RDWR))
	if ok, err := probe.TryLock(); err != nil || ok {
		_ = probe.Close()
		t.Fatalf("the lock was free while the holder process lived (taken=%v, err=%v)", ok, err)
	}

	kill()
	start := time.Now()
	unlock, err := lockDelivery(path)
	if err != nil {
		t.Fatalf("a killed holder's lock blocked the next delivery: %v", err)
	}
	unlock()
	if waited := time.Since(start); waited >= lockWait {
		t.Errorf("the next delivery waited %s for a dead holder", waited)
	}
}
