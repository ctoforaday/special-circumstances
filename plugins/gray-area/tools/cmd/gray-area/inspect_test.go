package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/claims"
)

// A transcript with three back-to-back runs of one command, a file edited twice
// with something between, and a Read that keys to nothing.
const repeats = `{"type":"assistant","uuid":"r-1","timestamp":"2026-08-18T01:00:00Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go test ./..."}}]}}
{"type":"assistant","uuid":"r-2","timestamp":"2026-08-18T01:01:00Z","message":{"content":[{"type":"tool_use","id":"t2","name":"Bash","input":{"command":"go test ./... 2>&1"}}]}}
{"type":"assistant","uuid":"r-3","timestamp":"2026-08-18T01:02:00Z","message":{"content":[{"type":"tool_use","id":"t3","name":"Bash","input":{"command":"go test ./... | tail -5"}}]}}
{"type":"assistant","uuid":"r-4","timestamp":"2026-08-18T01:03:00Z","message":{"content":[{"type":"tool_use","id":"t4","name":"Edit","input":{"file_path":"/w/a.go"}}]}}
{"type":"assistant","uuid":"r-5","timestamp":"2026-08-18T01:04:00Z","message":{"content":[{"type":"tool_use","id":"t5","name":"Read","input":{"file_path":"/w/b.go"}}]}}
{"type":"assistant","uuid":"r-6","timestamp":"2026-08-18T01:05:00Z","message":{"content":[{"type":"tool_use","id":"t6","name":"Edit","input":{"file_path":"/w/a.go"}}]}}
`

// A session where nothing repeats at all.
const clean = `{"type":"assistant","uuid":"c-1","timestamp":"2026-08-18T01:00:00Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go build ./..."}}]}}
{"type":"assistant","uuid":"c-2","timestamp":"2026-08-18T01:01:00Z","message":{"content":[{"type":"tool_use","id":"t2","name":"Edit","input":{"file_path":"/w/a.go"}}]}}
`

func TestReworkGroupsAndCitesEveryOccurrence(t *testing.T) {
	stdout, _, code := call(t, repeats, "rework", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "command x3") {
		t.Errorf("the three runs did not group despite differing only in plumbing:\n%s", stdout)
	}
	if !strings.Contains(stdout, "write   x2") {
		t.Errorf("the two edits to one file did not group:\n%s", stdout)
	}
	// Provenance on every occurrence, not a count.
	for _, u := range []string{"r-1", "r-2", "r-3", "r-4", "r-6"} {
		if !strings.Contains(stdout, u) {
			t.Errorf("occurrence %s is not cited — a count nobody can check:\n%s", u, stdout)
		}
	}
}

// THE PLAUSIBLE-ZERO GATE (spec §5 check 4). A clean session and a session the
// tool could barely read must not print the same thing.
func TestAnEmptyInspectionStatesItsSearch(t *testing.T) {
	for _, verb := range []string{"rework", "stalls"} {
		stdout, _, code := call(t, clean, verb, "/w/s.jsonl")
		if code != 0 {
			t.Fatalf("%s: exit %d", verb, code)
		}
		if !strings.Contains(stdout, "searched 2 citable tool uses") {
			t.Errorf("%s found nothing and did not say what it looked at:\n%s", verb, stdout)
		}
		if !strings.Contains(stdout, "happened exactly once") {
			t.Errorf("%s does not distinguish a clean run from an unreadable one:\n%s", verb, stdout)
		}
	}
	// The contrast: a session of nothing but Reads reports them as unkeyed, so an
	// empty answer over unreadable input cannot pass for an empty answer over a
	// clean one.
	onlyReads := `{"type":"assistant","uuid":"x-1","message":{"content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/w/a.go"}}]}}` + "\n"
	stdout, _, _ := call(t, onlyReads, "rework", "/w/s.jsonl")
	if !strings.Contains(stdout, "1 could not be keyed") {
		t.Errorf("unkeyed events vanished from the coverage line:\n%s", stdout)
	}
}

// The 3-strike rule is CONSECUTIVE. Ordinary iteration must not trip it.
func TestStallsFiresOnARunAndNotOnAScattering(t *testing.T) {
	stdout, _, code := call(t, repeats, "stalls", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "3 IN A ROW") {
		t.Errorf("three back-to-back runs did not trip the limit:\n%s", stdout)
	}
	// The two edits to a.go are separated by a Read; they are 2 occurrences and
	// not consecutive, so they must not appear.
	if strings.Contains(stdout, "/w/a.go") {
		t.Errorf("a scattered pair was reported as a stall:\n%s", stdout)
	}
}

// THE ACT/RESULT GATE (spec §5 check 5) for stalls. A row that reads "you were
// spinning" from a record that cannot see failure is the false positive this
// phase was told to solve, so the boundary rides on the row itself.
func TestAStallRowCarriesItsOwnLimitInBothOutputs(t *testing.T) {
	stdout, _, _ := call(t, repeats, "stalls", "/w/s.jsonl")
	if !strings.Contains(stdout, "not measured:") || !strings.Contains(stdout, "not their output") {
		t.Errorf("the table output does not state what it could not see:\n%s", stdout)
	}
	jsonOut, _, _ := call(t, repeats, "stalls", "-json", "/w/s.jsonl")
	var row struct {
		Note        string `json:"note"`
		Consecutive int    `json:"consecutive"`
	}
	line := strings.SplitN(strings.TrimSpace(jsonOut), "\n", 2)[0]
	if err := json.Unmarshal([]byte(line), &row); err != nil {
		t.Fatalf("json row did not parse: %v\n%s", err, jsonOut)
	}
	if row.Note == "" {
		t.Error("the JSON row carries no limit — a consumer would read the verdict as complete")
	}
	if row.Consecutive != 3 {
		t.Errorf("consecutive = %d, want 3", row.Consecutive)
	}
}

// -min below 2 would make every act a repeat. Refusing beats emitting a listing
// whose every row is a false positive.
func TestReworkRefusesAMinThatWouldReportEverything(t *testing.T) {
	_, stderr, code := call(t, repeats, "rework", "-min", "1", "/w/s.jsonl")
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(stderr, "false positive") {
		t.Errorf("the refusal does not say why: %q", stderr)
	}
}

// Both new verbs must be reachable, and an unknown one must still be refused
// rather than silently doing nothing.
func TestTheNewVerbsAreDispatchedAndUnknownOnesAreNot(t *testing.T) {
	for _, verb := range []string{"rework", "stalls"} {
		if _, _, code := call(t, clean, verb, "/w/s.jsonl"); code != 0 {
			t.Errorf("%s is not dispatched: exit %d", verb, code)
		}
	}
	_, stderr, code := call(t, clean, "reworking", "/w/s.jsonl")
	if code != 2 || !strings.Contains(stderr, "unknown command") {
		t.Errorf("a near-miss verb was not refused: exit %d %q", code, stderr)
	}
}

// THE ACT/RESULT GATE, end to end (spec §5 check 5). This is the answer to the
// question plans/gray-area-phase-2.md §0 left open and called the builder's
// problem, so it is asserted through the real binary in BOTH output shapes.
func TestAPrBodyClaimWithAnAssertedOutcomeSaysWhatItCouldNotCheck(t *testing.T) {
	body := "Validation: `go run ./check` -> 26 passed, 2 failed.\n"
	trace := `{"type":"assistant","uuid":"p-1","timestamp":"2026-08-18T01:00:00Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"cd scripts && go run ./check"}}]}}` + "\n"
	files := map[string]string{"/w/body.md": body, "/w/s.jsonl": trace}

	stdout, _, code := callFiles(t, files, "pr", "/w/body.md", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d — a pull request body is read AFTER the fact; a finding is not a failure\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "CITED") {
		t.Fatalf("the act was not cited despite being in the trajectory:\n%s", stdout)
	}
	if !strings.Contains(stdout, "NOT MEASURED") || !strings.Contains(stdout, "26 passed") {
		t.Errorf("the table does not name the outcome it could not check — CITED then reads as endorsing the count:\n%s", stdout)
	}

	jsonOut, _, _ := callFiles(t, files, "pr", "-json", "/w/body.md", "/w/s.jsonl")
	var f struct {
		Verdict    string   `json:"verdict"`
		Unmeasured []string `json:"unmeasured"`
	}
	line := strings.SplitN(strings.TrimSpace(jsonOut), "\n", 2)[0]
	if err := json.Unmarshal([]byte(line), &f); err != nil {
		t.Fatalf("json row did not parse: %v\n%s", err, jsonOut)
	}
	if f.Verdict != "CITED" || len(f.Unmeasured) == 0 {
		t.Errorf("json row lost the boundary: %+v", f)
	}
}

// A body whose claim is not in the trajectory gets NO-EVIDENCE and exit 0. The
// verdict states the search; it does not convict the body.
func TestAnUnmatchedPrClaimIsAnAbsenceNotAConviction(t *testing.T) {
	files := map[string]string{
		"/w/body.md": "Ran `go run ./nonexistent-tool` before pushing.\n",
		"/w/s.jsonl": `{"type":"assistant","uuid":"p-1","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"ls"}}]}}` + "\n",
	}
	stdout, _, code := callFiles(t, files, "pr", "/w/body.md", "/w/s.jsonl")
	if code != 0 {
		t.Errorf("exit %d — NO-EVIDENCE must not fail the process; that turns 'the tokens did not match' into 'the pull request is wrong'", code)
	}
	if !strings.Contains(stdout, "NO-EVIDENCE") || !strings.Contains(stdout, "searched:") {
		t.Errorf("the absence does not state its search:\n%s", stdout)
	}
	if !strings.Contains(stdout, "ABSENCE, not a finding that the body is wrong") {
		t.Errorf("the summary does not disclaim the conviction:\n%s", stdout)
	}
}

// A body with nothing adjudicable must SAY SO. "0 claims checked, all fine" and
// "this body made no checkable claim" are the same empty output otherwise — the
// plausible zero, in the verb most likely to be read by a human in a hurry.
func TestABodyWithNoAdjudicableClaimSaysSoRatherThanReportingNothing(t *testing.T) {
	files := map[string]string{
		"/w/body.md": "This change renames a field and updates the docs. No behaviour moved.\n",
		"/w/s.jsonl": `{"type":"assistant","uuid":"p-1","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"ls"}}]}}` + "\n",
	}
	stdout, _, code := callFiles(t, files, "pr", "/w/body.md", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "no adjudicable claim") {
		t.Errorf("an unclaimable body produced a silent zero:\n%s", stdout)
	}
	if !strings.Contains(stdout, "may be entirely accurate") {
		t.Errorf("the message reads as an accusation rather than as a limit:\n%s", stdout)
	}
}

// THE LINES A HUMAN READS, which had no test until the split that made them
// reachable. Verified against the real manifest as well (plans/gray-area.md §11.12);
// this is the part that stays checked after the manifest is gone.
func TestCoverageReportsDistinctSeatsAndNamesTheRepeats(t *testing.T) {
	census := claims.SeatCensus{
		IDs:     []string{"a703", "once"},
		Rows:    3,
		Repeats: map[string]int{"a703": 2},
	}
	c := claims.Reconcile("/p/sess.jsonl", census.IDs,
		func(string) ([]os.DirEntry, error) { return []os.DirEntry{}, nil })

	var out strings.Builder
	if code := reportCoverage(&out, census, c); code != 0 {
		t.Fatalf("exit %d — it measured, so the exit is 0 whatever it found", code)
	}
	got := out.String()

	// A ROW COUNT IS NOT A SEAT COUNT. Printing 3 beside a per-file disk count is
	// two denominators wearing one comparison.
	if !strings.Contains(got, "2 distinct seat(s) named by 3 row(s)") {
		t.Errorf("headline does not separate seats from rows:\n%s", got)
	}
	if !strings.Contains(got, "REPEAT    agent-a703 — 2 captures") {
		t.Errorf("the repeat was erased rather than reported:\n%s", got)
	}
	// The repeated seat is absent from disk here, and must be listed ONCE.
	if n := strings.Count(got, "MISSING   agent-a703"); n != 1 {
		t.Errorf("agent-a703 listed %d times under MISSING, want 1:\n%s", n, got)
	}
}

// An interrupted append leaves a line that will not parse, and the seat row lost
// with it surfaces as UNNAMED. Saying so is what separates that from a hook that
// never fired — two faults with different fixes and identical-looking output.
func TestCoverageSaysHowManyLinesDidNotParse(t *testing.T) {
	census := claims.SeatCensus{IDs: []string{"a"}, Rows: 1, Unreadable: 2}
	c := claims.Reconcile("/p/sess.jsonl", census.IDs,
		func(string) ([]os.DirEntry, error) { return []os.DirEntry{}, nil })
	var out strings.Builder
	reportCoverage(&out, census, c)
	if !strings.Contains(out.String(), "2 line(s) in this manifest are not JSON") {
		t.Errorf("torn lines were skipped in silence:\n%s", out.String())
	}
}
