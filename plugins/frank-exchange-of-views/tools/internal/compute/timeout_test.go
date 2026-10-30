package compute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A TIMEOUT SAYS HOW LONG THE LAUNCH TOOK AND HOW LONG IT WAITED (#1554). One Windows run timed a
// one-line print out at a minute; the message said only that, so it could not tell a slow
// interpreter start from a hung script.
func TestATimeoutSeparatesLaunchFromWait(t *testing.T) {
	dir := t.TempDir()
	// ONE PROCESS, so the kill ends everything the script started. A `go run` script leaves its
	// built binary running after the kill, and on Windows that orphan holds the test's directory.
	script := filepath.Join(dir, "slow.py")
	if err := os.WriteFile(script, []byte("import time\ntime.sleep(60)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	argv, err := interpreterFor(script)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = executeWithin(500*time.Millisecond, dir, argv, script, nil)
	if err == nil {
		t.Fatal("a script past its limit was not refused")
	}
	for _, want := range []string{"did not finish inside 500ms", "launching ", " took ", "waiting on it"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the timeout must carry %q: %v", want, err)
		}
	}
}

// A SCRIPT THAT EXITS WHILE A CHILD STILL HOLDS ITS OUTPUT IS NAMED, not waited on: the output the
// tool records is the output at exit, and a child writing after it is not part of the proof.
func TestAChildHoldingTheOutputIsNamed(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "forks.sh")
	if err := os.WriteFile(script, []byte("sleep 20 &\necho done\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	argv, err := interpreterFor(script)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, _, err = executeWithin(time.Minute, dir, argv, script, nil)
	if err == nil || !strings.Contains(err.Error(), "still held its output") {
		t.Fatalf("a child holding the output must be named: %v", err)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("the tool waited %s on a child it did not start", d)
	}
}
