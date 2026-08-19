package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE OPERATOR VERBS WERE DRIVEN BY NOTHING (#185 §6).
//
// `verify.go`'s own doc comment calls it "the human's and CI's instrument … A failed invariant
// exits non-zero so CI can gate on it". Measured at HEAD before this file existed:
//
//	newVerify    10.0%      graph.go:newGraph            23.5%
//	printReport   0.0%      countclaims.go:newCountClaims 11.8%
//	eventTally    0.0%
//	nonEmpty      0.0%
//
// #185 recorded that no difftest scenario named the three verbs. That grep is a FALSE POSITIVE
// today and the correction matters more than the number: `grep '"verify"'` now matches
// `lens verify` — a seat verb that records a citation verification — which is a DIFFERENT
// command with a different newVerify, in internal/cli/lens. Two verbs sharing a word, so the
// issue's own reproduction recipe would report "covered" while the operator verb stayed
// untouched. Checked properly: `graph` and `count-claims` have zero test references, and every
// `verify` reference in the suite is the seat verb.
//
// What these tests hold is the part a human reads and CI would gate on: the report renders, the
// exit code carries the verdict, and the tally's ORDER is the contract eventTally's own comment
// claims ("most-frequent-first").

// A clean run: every invariant holds, the report renders, exit is zero.
func TestVerifyRendersItsReportAndSucceedsOnASoundRun(t *testing.T) {
	runDir := seatRun(t)

	out, err := run(t, "verify", "--run", runDir)
	if err != nil {
		t.Fatalf("verify on a sound run: %v\n%s", err, out)
	}
	for _, want := range []string{"invariants:", "record (verdict", "gaps:", "findings:", "events:"} {
		if !strings.Contains(out, want) {
			t.Errorf("the report is what an operator reads; it is missing %q:\n%s", want, out)
		}
	}
}

// An unrecorded verdict must render as a WORD, not as an empty gap in the line. This is what
// nonEmpty exists for, and it was at 0.0%.
func TestVerifyNamesAnUnrecordedVerdictRatherThanLeavingItBlank(t *testing.T) {
	out, err := run(t, "verify", "--run", seatRun(t))
	if err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	if !strings.Contains(out, "verdict unrecorded") {
		t.Errorf("a run with no verdict must say so; a blank would read as a rendering bug:\n%s", out)
	}
}

// The JSON form is the one a machine would gate on, so it must carry the same answer.
func TestVerifyJSONCarriesTheSameVerdictAsTheReport(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "verify", "--run", runDir, "--json")
	if err != nil {
		t.Fatalf("verify --json: %v\n%s", err, out)
	}
	var got struct {
		OK     bool `json:"ok"`
		Checks []struct {
			Name string `json:"Name"`
			OK   bool   `json:"OK"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("verify --json emitted unparseable JSON: %v\n%s", err, out)
	}
	if !got.OK {
		t.Errorf("ok=false on a sound run; checks: %+v", got.Checks)
	}
	if len(got.Checks) == 0 {
		t.Error("no checks in the JSON — an empty check list reports OK for a run nothing examined, " +
			"which is the plausible zero this verb exists to prevent")
	}
}

// verify writes NOTHING. It is the one property that makes it safe to point at a live run.
func TestVerifyIsReadOnly(t *testing.T) {
	runDir := seatRun(t)
	before, err := run(t, "verify", "--run", runDir, "--json")
	if err != nil {
		t.Fatal(err)
	}
	after, err := run(t, "verify", "--run", runDir, "--json")
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("two consecutive verifies disagreed, so the first changed something:\n%s\n---\n%s", before, after)
	}
}

// A missing --run is a REFUSAL, not an empty report. A verify that says "0 gaps" because it
// looked at nothing is the failure mode the whole verb is aimed at.
func TestVerifyRefusesWithoutARun(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	out, err := run(t, "verify")
	if err == nil {
		t.Fatalf("verify with no run must refuse rather than report an empty board:\n%s", out)
	}
	if !strings.Contains(err.Error(), "--run") {
		t.Errorf("the refusal must name what is missing, got: %v", err)
	}
}

// eventTally's doc comment states the order: most-frequent-first, ties broken alphabetically.
// It was at 0.0%, so the stated contract was unenforced — and an unstable order in a report a
// human diffs between rounds is noise that reads as change.
func TestEventTallyOrdersByFrequencyThenName(t *testing.T) {
	got := eventTally(map[string]int{"finding": 2, "mint": 9, "closing": 2, "verdict": 1})
	const want = "mint=9, closing=2, finding=2, verdict=1"
	if got != want {
		t.Errorf("eventTally = %q, want %q (most-frequent-first, ties alphabetical)", got, want)
	}
}

func TestEventTallyOnNoEvents(t *testing.T) {
	if got := eventTally(map[string]int{}); got != "" {
		t.Errorf("eventTally on an empty map = %q, want empty", got)
	}
}

func TestNonEmptyFallsBackOnlyOnEmpty(t *testing.T) {
	if got := nonEmpty("", "unrecorded"); got != "unrecorded" {
		t.Errorf("nonEmpty(\"\") = %q", got)
	}
	if got := nonEmpty("pass", "unrecorded"); got != "pass" {
		t.Errorf("nonEmpty(\"pass\") = %q — a real value must never be replaced", got)
	}
}

// ---- the two sibling operator verbs, which had ZERO test references ----

func TestGraphRendersARunsBehaviour(t *testing.T) {
	out, err := run(t, "graph", "--run", seatRun(t))
	if err != nil {
		t.Fatalf("graph: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("graph produced nothing; an empty render and a run with no behaviour are the same bytes")
	}
}

func TestCountClaimsRefusesWhenThereIsNoReport(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	// count-claims reads blue/report.md; with no run it must refuse rather than answer 0.
	out, err := run(t, "count-claims")
	if err == nil && strings.Contains(out, "0") {
		t.Error("count-claims answered 0 with no report to count — a claim count of zero and " +
			"a report that does not exist must not be the same answer")
	}
}

// THE NON-ZERO EXIT, which is the entire promise verify.go's doc comment makes to CI:
//
//	A failed invariant exits non-zero so CI can gate on it.
//
// Every other test here drives a SOUND run, so all of them would still pass if that promise were
// removed — I noticed while deciding what to mutate, not while writing them. A suite that only
// exercises the happy path cannot tell a gate from a report.
//
// THE VIOLATION IS WRITTEN BY HAND, and that is not a shortcut. The live gate in record.Append
// refuses `merge verdict --as PASS` while any gap is open ("1 gap(s) still OPEN: R1-1"), so the
// #67 contradiction cannot be produced through the tool at all. A record that reached this state
// some other way — a hand-edited shard, a legacy run, a live gate that regressed — is exactly
// and only what an after-the-fact verifier is for, so that is the record this test builds.
func TestVerifyExitsNonZeroWhenAnInvariantFails(t *testing.T) {
	runDir := seatRun(t)
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "overclaim",
		"--quote", "Five independent verification approaches agree.",
		"--problem", "the defect", "--fix", "drop the independence claim",
		"--check-kind", "document", "--check", "the section no longer claims independence",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low",
		"--quote", "Five independent verification", "--new", "Five verification"); err != nil {
		t.Fatalf("minting the gap this test needs: %v", err)
	}

	// The PASS the writer would refuse, APPENDED TO THE SEAT'S OWN SHARD.
	//
	// Not a new shard. A seat's events live in one shard per nonce, and replay picks a single
	// winner per seat — so writing `events-red-merge-r1-<new nonce>.jsonl` DISPLACES the real
	// shard instead of adding to it. The first version of this test did exactly that, and the
	// board came back with "gaps: 0 total": the fixture deleted the gap it existed to
	// contradict, pass-closes-all-gaps reported ok, and the test passed on a different
	// invariant (register-before-append) firing for a seat whose history had vanished.
	shards, err := filepath.Glob(filepath.Join(runDir, "records", "events-red-merge-r1-*.jsonl"))
	if err != nil || len(shards) != 1 {
		t.Fatalf("expected exactly one shard for red-merge-r1, got %v (%v)", shards, err)
	}
	existing, err := os.ReadFile(shards[0])
	if err != nil {
		t.Fatal(err)
	}
	nonce := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(shards[0]), "events-red-merge-r1-"), ".jsonl")
	ev := record.Event{
		Seq: len(strings.Split(strings.TrimSpace(string(existing)), "\n")),
		TS:  "2099-01-01T00:00:00.000000000Z", SeatID: "red-merge-r1", Nonce: nonce,
		Round: 1, Type: "verdict", Key: "red-merge-r1:verdict",
		Payload: record.NewPayload().Set("verdict", "PASS"),
	}
	line, err := record.MarshalEvent(ev)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shards[0], append(existing, append(line, '\n')...), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "verify", "--run", runDir)
	if err == nil {
		t.Fatalf("verify returned SUCCESS on a record whose PASS left a gap open. The doc comment "+
			"promises CI a non-zero exit; without it the verb is a report wearing a gate's "+
			"description.\n%s", out)
	}
	// NAME the invariant. Asserting only "something failed" let this test pass with the gate
	// under test broken — proved by mutation: reverting passClosesAllGaps to read `outcome`
	// events killed the unit tests and left this one green, because the hand-written shard also
	// trips register-before-append. A test that cannot say which check fired is testing that
	// verify fails, not that the gate works.
	if !strings.Contains(out, "pass-closes-all-gaps") {
		t.Errorf("the report must name the #67 gate as the failure:\n%s", out)
	}
	if !strings.Contains(out, "[FAIL] pass-closes-all-gaps") {
		t.Errorf("pass-closes-all-gaps must be the check marked FAIL:\n%s", out)
	}
	if !strings.Contains(err.Error(), "invariant") {
		t.Errorf("the error must name what failed, got: %v", err)
	}
}
