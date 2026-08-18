package prbody

import (
	"strings"
	"testing"
)

func parse(t *testing.T, body string) []Claim {
	t.Helper()
	cs, err := Parse(strings.NewReader(body), "/w/body.md")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return cs
}

// THE ACT/RESULT SPLIT (spec §2), which is this package's whole reason to exist.
// A claim asserting an outcome must carry that outcome so the finding can name
// what it could not check.
func TestAResultAssertionIsCapturedSeparatelyFromTheAct(t *testing.T) {
	cs := parse(t, "Validation: `go run ./check` → 26 passed, 2 failed, 4 not run.\n")
	if len(cs) != 1 {
		t.Fatalf("want 1 claim, got %d: %+v", len(cs), cs)
	}
	if cs[0].Cmd != "go run ./check" {
		t.Errorf("act = %q", cs[0].Cmd)
	}
	if cs[0].Result == "" {
		t.Fatal("the asserted outcome was dropped — a CITED row would then read as endorsing the count")
	}
	if !strings.Contains(cs[0].Result, "26 passed") {
		t.Errorf("result = %q, does not carry what was asserted", cs[0].Result)
	}
}

// A claim with no outcome asserted must NOT gain a caveat. Attaching one
// everywhere would make the marker noise, and noise is how a real one gets
// skipped.
func TestAnActWithNoAssertedOutcomeCarriesNoResult(t *testing.T) {
	for _, body := range []string{
		"I ran `go run ./check` before pushing.\n",
		"See `git log --oneline -1` for the head.\n",
	} {
		cs := parse(t, body)
		if len(cs) != 1 {
			t.Fatalf("want 1 claim from %q, got %d", body, len(cs))
		}
		if cs[0].Result != "" {
			t.Errorf("invented a result assertion for %q: %q", body, cs[0].Result)
		}
	}
}

// A code span is not automatically a command. Searching the trajectory for a file
// path reports NO-EVIDENCE against a body that never claimed to run anything —
// a false accusation from a guessed referent.
func TestOnlySpansThatReadAsCommandsBecomeClaims(t *testing.T) {
	cs := parse(t, "Touched `plans/gray-area.md` and `main.go`, added `kind: \"turn-end\"`, "+
		"kept `Unmeasured` on the finding. Then ran `go test ./...`.\n")
	if len(cs) != 1 {
		t.Fatalf("want exactly the one real command, got %d: %+v", len(cs), cs)
	}
	if cs[0].Cmd != "go test ./..." {
		t.Errorf("claim = %q", cs[0].Cmd)
	}
}

// Fenced blocks hold OUTPUT as often as commands. Reading a pasted result line as
// a command searches for something nobody claimed to run.
func TestFencedBlocksAreSkipped(t *testing.T) {
	body := "Ran `go run ./check`:\n\n```\ngo test ./... FAILED\n26 passed, 2 failed\n```\n\nAll green.\n"
	cs := parse(t, body)
	if len(cs) != 1 {
		t.Fatalf("want 1 claim, got %d: %+v", len(cs), cs)
	}
	if cs[0].Cmd != "go run ./check" {
		t.Errorf("a fenced output line became a claim: %q", cs[0].Cmd)
	}
}

// Provenance: every claim cites the body line it came from, because the finding
// built on it cites both sides.
func TestEveryClaimCitesItsOwnLine(t *testing.T) {
	cs := parse(t, "intro\n\nran `go build ./...` here\n\nand `go vet ./...` there\n")
	if len(cs) != 2 {
		t.Fatalf("want 2, got %d", len(cs))
	}
	if cs[0].Line != 3 || cs[1].Line != 5 {
		t.Errorf("lines = %d, %d; want 3, 5", cs[0].Line, cs[1].Line)
	}
	for _, c := range cs {
		if c.File == "" || c.Line == 0 {
			t.Errorf("a claim with no provenance: %+v", c)
		}
	}
}

// A body repeating one command must not produce two rows saying the same thing.
func TestARepeatedCommandIsClaimedOnce(t *testing.T) {
	cs := parse(t, "Run `go run ./check`.\n\nAgain: `go run ./check`.\n")
	if len(cs) != 1 {
		t.Errorf("want 1 deduplicated claim, got %d", len(cs))
	}
}

// MEASURED ON A REAL BODY, and the reason the tail is bounded by the next span.
// A line holding three claims gave all three the LAST outcome on it:
//
//	CITED  `gray-area checkpoint`   NOT MEASURED: the asserted outcome (-> 4 checked, clean.)
//	CITED  `rulesweep …`            NOT MEASURED: the asserted outcome (-> 4 checked, clean.)
//	CITED  `versionguard …`         NOT MEASURED: the asserted outcome (-> 4 checked, clean.)
//
// Only the third was right. That is misattribution inside the tool built to catch
// it, and a caveat naming the wrong number is worse than no caveat: a reader
// checks the number and finds the tool wrong.
func TestEachClaimGetsItsOwnOutcomeNotTheLineLast(t *testing.T) {
	cs := parse(t, "`gray-area checkpoint` -> 12/12 CITED. `rulesweep -base origin/main` -> nothing to sweep. `versionguard -base origin/main` -> 4 checked, clean.\n")
	if len(cs) != 3 {
		t.Fatalf("want 3 claims, got %d: %+v", len(cs), cs)
	}
	want := []string{"12/12", "nothing to sweep", "4 checked"}
	for i, w := range want {
		if !strings.Contains(cs[i].Result, w) {
			t.Errorf("claim %d (%s) carries %q, want it to name %q",
				i, cs[i].Cmd, cs[i].Result, w)
		}
	}
	// And the specific wrong answer must not recur: only the last claim may name
	// the last outcome.
	for i := 0; i < 2; i++ {
		if strings.Contains(cs[i].Result, "4 checked") {
			t.Errorf("claim %d (%s) took the LAST outcome on the line: %q", i, cs[i].Cmd, cs[i].Result)
		}
	}
}
