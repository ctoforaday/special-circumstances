package main

import (
	"strings"
	"testing"
)

const workflowWithTest = `
      - run: >-
          node --test
          plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs
          plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs
`

func TestListedExtractsEverySuite(t *testing.T) {
	got := listed(workflowWithTest)
	if len(got) != 2 {
		t.Fatalf("expected 2 listed suites, got %d: %v", len(got), got)
	}
	if got[0] != "plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs" {
		t.Errorf("listed suites must be sorted and complete: %v", got)
	}
}

func TestCompareAgrees(t *testing.T) {
	both := []string{"plugins/a/tests/x.test.mjs"}
	if p := compare("node --test", both, both); len(p) != 0 {
		t.Errorf("matching lists must produce no problems: %v", p)
	}
}

// THE case that made this guard necessary: CI names a suite that does not exist. `node
// --test` exits 0 on a missing path, so it counted as a pass while testing nothing.
func TestCompareCatchesAListedButMissingSuite(t *testing.T) {
	p := compare("node --test", nil, []string{"plugins/a/tests/ghost.test.mjs"})
	if len(p) != 1 {
		t.Fatalf("expected one problem, got %v", p)
	}
	if !strings.Contains(p[0], "does not exist") || !strings.Contains(p[0], "testing nothing") {
		t.Errorf("the message must name the consequence: %q", p[0])
	}
}

// The other direction: a live suite nobody runs. record.test.mjs went its whole life this
// way and never ran in CI once.
func TestCompareCatchesAnUnlistedSuite(t *testing.T) {
	// `have` carries a real entry so this isolates the unlisted-suite check: an EMPTY
	// listed set legitimately trips the "read nothing" guard as well, which is correct
	// behaviour and a different finding.
	listedOne := "plugins/a/tests/other.test.mjs"
	p := compare("node --test", []string{listedOne, "plugins/a/tests/real.test.mjs"}, []string{listedOne})
	if len(p) != 1 {
		t.Fatalf("expected one problem, got %v", p)
	}
	if !strings.Contains(p[0], "never run in CI") {
		t.Errorf("the message must say it has never run: %q", p[0])
	}
}

func TestCompareCatchesADuplicate(t *testing.T) {
	s := "plugins/a/tests/x.test.mjs"
	p := compare("node --test", []string{s}, []string{s, s})
	if len(p) != 1 || !strings.Contains(p[0], "more than once") {
		t.Errorf("a duplicated entry must be reported: %v", p)
	}
}

// A guard that stops reading its input must SAY SO, not pass. This is the same failure
// shape it exists to catch, one level up — and it is how this guard retires cleanly when
// the last .mjs suite is finally ported: loudly, as a decision, rather than by drifting
// into vacuous success.
func TestCompareFailsLoudlyWhenItCannotRead(t *testing.T) {
	p := compare("a workflow with no node test step", nil, nil)
	if len(p) != 1 {
		t.Fatalf("expected one problem, got %v", p)
	}
	if !strings.Contains(p[0], "no `node --test` step found") {
		t.Errorf("must report that it read nothing: %q", p[0])
	}

	p = compare("node --test", nil, nil)
	if len(p) != 1 || !strings.Contains(p[0], "could not extract a single suite path") {
		t.Errorf("a step present but unparseable must be reported: %v", p)
	}
}
