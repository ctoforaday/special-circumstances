//go:build linux

package peer

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
		h := otherDBHolder(root, conv, myPID+1)
		if h.Kind != HolderDBOpen || h.PID != myPID || !strings.Contains(h.Detail, fmt.Sprintf("pid %d", myPID)) {
			t.Errorf("otherDBHolder(%q, %q, %d) = %+v, want the db open in pid %d", root, conv, myPID+1, h, myPID)
		}
		if h := otherDBHolder(root, conv, myPID); h.Kind != HolderNone {
			t.Errorf("otherDBHolder excluding holder pid %d = %+v, want none", myPID, h)
		}
	}

	// Closing the file clears the busy state.
	f.Close()
	if h := otherDBHolder(realRoot, conv, myPID+1); h.Kind != HolderNone {
		t.Errorf("after closing db, otherDBHolder = %+v, want none", h)
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
	if h := ConversationBusy(agyRoot, conv, myPID); h.Kind != HolderNone {
		t.Errorf("ConversationBusy with holder pid %d excluded = %+v, want none", myPID, h)
	}
	if h := ConversationBusy(agyRoot, conv, myPID+1); h.Kind != HolderPresenceLock || h.PID != myPID {
		t.Errorf("ConversationBusy with other pid excluded = %+v, want the presence lock held by pid %d", h, myPID)
	}
}

// A PRESENCE-LOCK HOLDER IS NAMED BY ITS PID FROM /proc/locks, and a waiter is not a holder.
func TestLockHoldersReadsProcLocks(t *testing.T) {
	const locks = `1: POSIX  ADVISORY  WRITE 900 00:2a:1234 0 EOF
2: FLOCK  ADVISORY  WRITE 4242 08:01:1234 0 EOF
2: -> FLOCK  ADVISORY  WRITE 5151 08:01:1234 0 EOF
3: FLOCK  ADVISORY  WRITE 777 08:01:9999 0 EOF
4: FLOCK  ADVISORY  READ 0 08:01:1234 0 EOF
`
	got, err := lockHolders(locks, "1234")
	if err != nil || fmt.Sprint(got) != "[4242 0]" {
		t.Errorf("lockHolders = %v, %v; want [4242 0]: the holder and the one outside this pid namespace, not the waiter", got, err)
	}
	if got, err := lockHolders(locks, "5555"); err != nil || len(got) != 0 {
		t.Errorf("lockHolders of an unlocked inode = %v, %v; want none", got, err)
	}
}

// A HOLDER THIS NAMESPACE CANNOT NAME IS UNINSPECTABLE, never "pid 0".
// WHERE /proc/locks CANNOT SHOW THE HOLDER, THE PROBE FAILS CLOSED. Here the process that took
// the lock (flock(1)) has exited, and the shell it shared the descriptor with holds it.
func TestConversationBusyLockerThatExitedIsUninspectable(t *testing.T) {
	flockBin, err := exec.LookPath("flock")
	if err != nil {
		t.Skip("no flock(1) on this box")
	}
	agyRoot := t.TempDir()
	const conv = "conv-passed-on"
	lockPath := filepath.Join(agyRoot, "presence", conv+".lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}
	sh := exec.Command("sh", "-c", `exec 9>>"$1"; "$2" -x 9; echo locked; exec sleep 60`, "sh", lockPath, flockBin)
	out, err := sh.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := sh.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sh.Process.Kill(); _ = sh.Wait() })
	if line, err := bufio.NewReader(out).ReadString('\n'); err != nil || line != "locked\n" {
		t.Fatalf("shell said %q, %v; want locked", line, err)
	}
	h := ConversationBusy(agyRoot, conv, 0)
	if h.Kind != HolderUninspectable || !strings.Contains(h.Detail, "which has exited") {
		t.Errorf("ConversationBusy = %+v, want uninspectable naming the exited locker", h)
	}
}

func TestBlindSpot(t *testing.T) {
	const flat, nested, ext4 = "NSpid:\t4242\n", "NSpid:\t4242\t17\n", 0xEF53
	for _, tc := range []struct {
		status string
		magic  uint32
		want   string
	}{
		{flat, ext4, ""},
		{nested, ext4, "this process runs in a nested pid namespace"},
		{"Pid:\t4242\n", ext4, "/proc/self/status has no NSpid line"},
		{flat, 0x6969, "the lock is on NFS"},
	} {
		if got := blindSpot(tc.status, tc.magic); (tc.want == "") != (got == "") || !strings.HasPrefix(got, tc.want) {
			t.Errorf("blindSpot(%q, %#x) = %q, want %q", tc.status, tc.magic, got, tc.want)
		}
	}
}

// AN UNHELD LOCK IN A BLIND SPOT IS UNINSPECTABLE, not free: here the blind spot is stood in.
func TestConversationBusyUnheldLockInABlindSpot(t *testing.T) {
	agyRoot := t.TempDir()
	const conv = "conv-blind"
	if err := os.MkdirAll(filepath.Join(agyRoot, "presence"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agyRoot, "presence", conv+".lock"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	was := blindSpotOf
	defer func() { blindSpotOf = was }()
	blindSpotOf = func(string) string { return "the lock is on NFS" }
	if h := ConversationBusy(agyRoot, conv, 0); h.Kind != HolderUninspectable || !strings.Contains(h.Detail, "the lock is on NFS") {
		t.Errorf("ConversationBusy = %+v, want uninspectable naming the blind spot", h)
	}
}

func TestNestedPidNamespace(t *testing.T) {
	for _, tc := range []struct {
		status        string
		nested, known bool
	}{
		{"Name:\tx\nNSpid:\t4242\n", false, true},
		{"Name:\tx\nNSpid:\t4242\t17\n", true, true},
		{"Name:\tx\nPid:\t4242\n", false, false},
	} {
		if nested, known := nestedPidNamespace(tc.status); nested != tc.nested || known != tc.known {
			t.Errorf("nestedPidNamespace(%q) = (%v, %v), want (%v, %v)", tc.status, nested, known, tc.nested, tc.known)
		}
	}
}

func TestNetworkFS(t *testing.T) {
	for m, want := range map[uint32]string{0x6969: "NFS", 0xFF534D42: "SMB", 0xFE534D42: "SMB", 0x517B: "SMB", 0x00C36400: "Ceph",
		0x5346414F: "AFS", 0x6B414653: "AFS", 0x01021997: "9P", 0x01161970: "GFS2", 0x7461636F: "OCFS2", 0x0BD00BD0: "Lustre",
		0x73757245: "Coda", 0xEF53: "", 0x01021994: "", 0x58465342: ""} {
		if got := networkFS(m); got != want {
			t.Errorf("networkFS(%#x) = %q, want %q", m, got, want)
		}
	}
}

// ON A LOCAL FILESYSTEM, IN THIS NAMESPACE, AN UNHELD LOCK FILE IS NOT A BLIND SPOT.
func TestLocksBlindSpotHere(t *testing.T) {
	status, err := os.ReadFile("/proc/self/status")
	if err != nil {
		t.Skip(err)
	}
	if nested, _ := nestedPidNamespace(string(status)); nested {
		t.Skip("this test runs in a nested pid namespace")
	}
	if why := locksBlindSpot(filepath.Join(t.TempDir(), "f.lock")); why != "" {
		t.Errorf("locksBlindSpot on a temporary directory = %q, want none", why)
	}
}
