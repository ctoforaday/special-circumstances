package compute

import (
	"strings"
	"testing"
	"time"
)

// A TIMEOUT SAYS HOW LONG THE LAUNCH TOOK AND HOW LONG IT WAITED (#1554). One Windows run timed a
// one-line print out at a minute; the message said only that, so it could not tell a slow
// interpreter start from a hung program.
func TestATimeoutSeparatesLaunchFromWait(t *testing.T) {
	p := program(t, bundle(t, map[string]string{"a": "a"}), "sleep", "60s")
	c, err := resolve(p.Argv)
	if err != nil {
		t.Fatal(err)
	}
	_, err = execute(500*time.Millisecond, t.TempDir()+"/bundle", p, c)
	if err == nil {
		t.Fatal("a program past its limit was not refused")
	}
	for _, want := range []string{"did not finish inside 500ms", "launching it took ", "waiting on it"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the timeout must carry %q: %v", want, err)
		}
	}
}

// A PROGRAM THAT EXITS WHILE A CHILD STILL HOLDS ITS OUTPUT IS NAMED, not waited on: the output the
// tool records is the output at exit, and a child writing after it is not part of the computation.
func TestAChildHoldingTheOutputIsNamed(t *testing.T) {
	p := program(t, bundle(t, map[string]string{"a": "a"}), "fork")
	start := time.Now()
	_, err := Run(t.TempDir(), p)
	if err == nil || !strings.Contains(err.Error(), "still held its output") {
		t.Fatalf("a child holding the output must be named: %v", err)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("the tool waited %s on a child it did not start", d)
	}
}
