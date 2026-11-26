//go:build linux

package peer

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gofrs/flock"
)

const dbHolderEnv = "GRAY_AREA_DB_HOLDER_HELPER"

// TestDBHolderHelperProcess is a child process of the ConversationBusy tests, not a test: it opens
// the file its environment names, says "open", and holds it until its stdin closes.
func TestDBHolderHelperProcess(t *testing.T) {
	path := os.Getenv(dbHolderEnv)
	if path == "" {
		t.Skip("runs only as a child of TestConversationBusyChecksTheDBPastTheExcludedLockHolder")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	os.Stdout.WriteString("open\n")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

// conversationOnDisk makes an Antigravity root holding conversation's database, and returns the
// root and the database's path.
func conversationOnDisk(t *testing.T, conversation string) (root, db string) {
	t.Helper()
	root = t.TempDir()
	for _, d := range []string{"conversations", "presence"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	db = filepath.Join(root, "conversations", conversation+".db")
	if err := os.WriteFile(db, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, db
}

// THE DATABASE CHECK IS PART OF THE PROBE. The daemon holds a conversation through its database
// alone — it takes no presence lock — so a probe that skipped the check would see nobody.
func TestConversationBusyFindsTheDBOpen(t *testing.T) {
	const conv = "conv-busy-db"
	root, db := conversationOnDisk(t, conv)
	f, err := os.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	me := os.Getpid()
	if h := ConversationBusy(root, conv, 0); h.Kind != HolderDBOpen || h.PID != me {
		t.Errorf("ConversationBusy excluding nobody = %+v, want the database open in pid %d", h, me)
	}
	if h := ConversationBusy(root, conv, me); h.Kind != HolderNone {
		t.Errorf("ConversationBusy excluding the only holder, pid %d = %+v, want none", me, h)
	}
}

// EXCLUDING THE LOCK'S HOLDER DOES NOT END THE PROBE. The excluded pid (this process) holds the
// presence lock while another process has the database open: that process is the holder.
func TestConversationBusyChecksTheDBPastTheExcludedLockHolder(t *testing.T) {
	const conv = "conv-busy-lock-and-db"
	root, db := conversationOnDisk(t, conv)
	fl := flock.New(filepath.Join(root, "presence", conv+".lock"), flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))
	if ok, err := fl.TryLock(); err != nil || !ok {
		t.Fatalf("TryLock = (%v, %v), want (true, nil)", ok, err)
	}
	defer fl.Unlock()

	child := exec.Command(os.Args[0], "-test.run=^TestDBHolderHelperProcess$", "-test.count=1")
	child.Env = append(os.Environ(), dbHolderEnv+"="+db)
	stdin, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stdin.Close()
		_ = child.Wait()
	})
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "open\n" {
		t.Fatalf("helper said %q, %v; want open", line, err)
	}

	h := ConversationBusy(root, conv, os.Getpid())
	if h.Kind != HolderDBOpen || h.PID != child.Process.Pid {
		t.Errorf("ConversationBusy excluding the lock's holder = %+v, want the database open in pid %d", h, child.Process.Pid)
	}
}
