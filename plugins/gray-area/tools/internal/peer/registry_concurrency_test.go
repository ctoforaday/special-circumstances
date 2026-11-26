package peer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// The registry stress fixtures judge a server by its pid alone, identically in every process: a pid
// below 1000 is a server that is gone, any other is running.
func stressLiveness(s Server) catalogue.Liveness {
	if s.PID < 1000 {
		return catalogue.Ended
	}
	return catalogue.Live
}

const (
	stressConversations = 12
	helperEnv           = "GRAY_AREA_REGISTRY_HELPER"
)

// stressWorker links stressConversations conversations on the running server pid, pruning between
// links, as many telepathy processes on one box do.
func stressWorker(rt string, pid int) error {
	for i := 0; i < stressConversations; i++ {
		conv := fmt.Sprintf("conv-%d-%d", pid, i)
		if _, _, err := RegisterWith(rt, agyShell(conv, "localhost:"+strconv.Itoa(pid), "tok"), hostID(pid, 1, false), at); err != nil {
			return err
		}
		if _, err := Prune(rt, stressLiveness); err != nil {
			return err
		}
	}
	return nil
}

// TestRegistryHelperProcess is a child process of TestRegistryConcurrentWritersAndPrunes, not a test.
func TestRegistryHelperProcess(t *testing.T) {
	spec := os.Getenv(helperEnv)
	if spec == "" {
		t.Skip("runs only as a child of TestRegistryConcurrentWritersAndPrunes")
	}
	rt, pidText, _ := strings.Cut(spec, "|")
	pid, err := strconv.Atoi(pidText)
	if err != nil {
		t.Fatal(err)
	}
	if err := stressWorker(rt, pid); err != nil {
		t.Fatal(err)
	}
}

func lockInode(t *testing.T, rt string) os.FileInfo {
	t.Helper()
	fi, err := os.Stat(filepath.Join(RegistryDir(rt), ".register.lock"))
	if err != nil {
		t.Fatalf(".register.lock: %v", err)
	}
	return fi
}

// MANY TELEPATHY PROCESSES, ONE REGISTRY. Goroutines and child processes link conversations on
// running servers and prune at once, over records of gone servers and files of the old shape —
// including old-shape files under the very keys being linked. Afterwards no conversation linked on
// a running server is missing, every record is whole, nothing of a gone server is left, and
// .register.lock is the same file it was.
func TestRegistryConcurrentWritersAndPrunes(t *testing.T) {
	children := []int{3001, 3002}
	if testing.Short() {
		children = nil
	}
	for round := 0; round < 8; round++ {
		rt := t.TempDir()
		goroutines := []int{2001, 2002, 2003, 2004}
		for _, pid := range append(append([]int{}, goroutines...), children...) {
			key := ServerKey(pid, 1, "")
			writeRecord(t, rt, key+".json", fmt.Sprintf(`{"key":%q,"pid":%d,"start_time":1,"address":"a","token":"t","conversations":["old"],"registered_at":1}`, key, pid))
		}
		for pid := 100; pid < 110; pid++ {
			writeRecord(t, rt, ServerKey(pid, 1, "")+".json", record(ServerKey(pid, 1, ""), pid, 1, LinkedConversation{ID: "gone", LinkedAt: 1}))
		}
		writeRecord(t, rt, ".register.lock", "")
		before := lockInode(t, rt)

		var wg sync.WaitGroup
		errs := make(chan error, 16)
		for _, pid := range children {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegistryHelperProcess$", "-test.count=1")
			cmd.Env = append(os.Environ(), helperEnv+"="+rt+"|"+strconv.Itoa(pid))
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := cmd.Wait(); err != nil {
					errs <- fmt.Errorf("child %d: %w", pid, err)
				}
			}()
		}
		for _, pid := range goroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := stressWorker(rt, pid); err != nil {
					errs <- err
				}
			}()
		}
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < stressConversations; j++ {
					if _, err := Prune(rt, stressLiveness); err != nil {
						errs <- err
					}
				}
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			t.Fatal(err)
		}

		for _, pid := range append(goroutines, children...) {
			s, err := LookupServerByKey(rt, ServerKey(pid, 1, ""))
			if err != nil || s == nil {
				t.Fatalf("round %d: the record of running server %d = %+v, %v", round, pid, s, err)
			}
			for i := 0; i < stressConversations; i++ {
				if conv := fmt.Sprintf("conv-%d-%d", pid, i); !s.Lists(conv) {
					t.Errorf("round %d: %s, linked on running server %d, was lost", round, conv, pid)
				}
			}
		}
		entries, err := os.ReadDir(RegistryDir(rt))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			s, err := LookupServerByKey(rt, strings.TrimSuffix(e.Name(), ".json"))
			if err != nil || s == nil {
				t.Errorf("round %d: %s is not a whole record: %v", round, e.Name(), err)
			} else if s.PID < 1000 {
				t.Errorf("round %d: %s, a gone server's record, survived", round, e.Name())
			}
		}
		if after := lockInode(t, rt); !os.SameFile(before, after) {
			t.Errorf("round %d: .register.lock was replaced", round)
		}
	}
}

// A READER NEEDS NO LOCK. While records of gone servers are written and pruned around it, and the
// running server's record is rewritten, a reader never errors and always finds the running
// server's whole record: a record pruned between listing and reading reads as absent, and a record
// being replaced reads as its old or its new self.
func TestReadersSkipRecordsPrunedUnderThem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows refuses to replace or remove a file another handle has open, so a lock-free reader there can meet a sharing violation; live delivery, which reads the registry while it changes, runs on Linux only")
	}
	rt := t.TempDir()
	if _, _, err := RegisterWith(rt, agyShell("conv-live", "localhost:1", "tok"), hostID(4000, 1, false), at); err != nil {
		t.Fatal(err)
	}
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			for pid := 100; pid < 120; pid++ {
				if _, _, err := RegisterWith(rt, agyShell(fmt.Sprintf("gone-%d", i), "localhost:2", "tok"), hostID(pid, 1, false), at); err != nil {
					t.Error(err)
					return
				}
			}
			if _, err := Prune(rt, stressLiveness); err != nil {
				t.Error(err)
				return
			}
			// The running server's record is rewritten by every relink.
			if _, _, err := RegisterWith(rt, agyShell("conv-live", "localhost:1", "tok"), hostID(4000, 1, false), at.Add(time.Duration(i)*time.Second)); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	deadline := time.Now().Add(400 * time.Millisecond)
	reads := 0
	for time.Now().Before(deadline) {
		if _, err := ListServers(rt); err != nil {
			t.Fatalf("ListServers during prune: %v", err)
		}
		s, err := LookupConversationServer(rt, "conv-live", stressLiveness)
		if err != nil || s == nil {
			t.Fatalf("LookupConversationServer during prune = %+v, %v", s, err)
		}
		if s, err := LookupServerByKey(rt, ServerKey(4000, 1, "")); err != nil || s == nil {
			t.Fatalf("LookupServerByKey of the record being rewritten = %+v, %v", s, err)
		}
		for pid := 100; pid < 120; pid++ {
			if _, err := LookupServerByKey(rt, ServerKey(pid, 1, "")); err != nil {
				t.Fatalf("LookupServerByKey of a record being pruned: %v", err)
			}
		}
		reads++
	}
	close(stop)
	wg.Wait()
	if reads == 0 {
		t.Fatal("no read ran")
	}
}

// A MESSAGE THE CLIENT READS WHILE THE SCAN RUNS IS READ, NOT UNREADABLE. A stand-in client removes
// each marker and then its message file while scans run; no scan may count an unreadable message.
func TestScanSkipsMessagesReadDuringScan(t *testing.T) {
	brain := t.TempDir()
	box := filepath.Join(brain, "conv-busy", ".system_generated", "messages")
	if err := os.MkdirAll(filepath.Join(box, "undelivered"), 0o755); err != nil {
		t.Fatal(err)
	}
	const n = 300
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("m%03d", i)
		body := fmt.Sprintf(`{"id":%q,"sender":"s","timestamp":"2026-09-09T10:00:00Z","content":"c"}`, id)
		if err := os.WriteFile(filepath.Join(box, id+".json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(box, "undelivered", id), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < n; i++ {
			id := fmt.Sprintf("m%03d", i)
			_ = os.Remove(filepath.Join(box, "undelivered", id))
			_ = os.Remove(filepath.Join(box, id+".json"))
		}
	}()
	for {
		ws, errs := ScanMailboxes([]string{brain}, "")
		if len(errs) != 0 {
			t.Fatalf("scan errors: %v", errs)
		}
		for _, w := range ws {
			if w.Unreadable != 0 {
				t.Fatalf("a scan counted %d messages the client was reading as unreadable", w.Unreadable)
			}
		}
		select {
		case <-done:
			return
		default:
		}
	}
}
