package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// stalenessArm is one state a standing check puts its anchor in, and the check that puts it there.
type stalenessArm struct {
	// kind is the EvidenceTables row the arm's anchor is of; state is the word the oracle names the
	// check's state by ("holds", or the row's `failed`), "" for a check that never stands.
	kind, state string
	check       func(t *testing.T, runDir, id string) (string, error)
}

// stalenessArms is every state #1647's ledger gives a check, once each: a check that holds, of
// either kind; an `unsound` judgement; a `found_against` reading, carried by the gap it minted; and an
// `unread` reading, which never stands.
func stalenessArms() map[string]stalenessArm {
	reproduce := func(as string) func(t *testing.T, runDir, id string) (string, error) {
		return func(t *testing.T, runDir, id string) (string, error) {
			return run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", id,
				"--as", as, "--reason", "read the script against the sentence")
		}
	}
	verify := func(as string) func(t *testing.T, runDir, id string) (string, error) {
		return func(t *testing.T, runDir, id string) (string, error) { return verifyAs(t, runDir, id, as) }
	}
	return map[string]stalenessArm{
		"citation holds":         {kind: "citation", state: "holds", check: verify("supports")},
		"computation holds":      {kind: "computation", state: "holds", check: reproduce("sound")},
		"computation unsound":    {kind: "computation", state: "unsound", check: reproduce("unsound")},
		"citation found_against": {kind: "citation", state: "found_against", check: verify("refutes")},
		"citation unread":        {kind: "citation", state: "", check: verify("unreachable")},
	}
}

// staleViolations is every staleness violation the consistency oracle reports of runDir.
func staleViolations(t *testing.T, runDir string) (consistency.Result, []consistency.Violation) {
	t.Helper()
	res, err := consistency.Check(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var stale []consistency.Violation
	for _, v := range res.Violations {
		if v.Rule == "staleness" {
			stale = append(stale, v)
		}
	}
	return res, stale
}

// lastCheckEvent is the events.id of the latest check of id in the check table of its kind's row.
func lastCheckEvent(t *testing.T, runDir string, row recordsql.EvidenceTable, id string) int64 {
	t.Helper()
	var ev int64
	if err := openRunDB(t, runDir).QueryRow(`SELECT max("event_id") FROM "`+row.CheckTable+`" WHERE "`+row.CheckIDColumn+`" = ?`, id).Scan(&ev); err != nil {
		t.Fatalf("read %s's last check: %v", id, err)
	}
	return ev
}

// A STANDING CHECK OF A SENTENCE THE REPORT NO LONGER CARRIES IS A VIOLATION (plans/compute-and-show.md
// III.1″.1 Staleness, V1a″.1), once per state the ledger gives a standing check, over every
// EvidenceTables row. The honest record — the edit names the anchor reopened, so the check no longer
// stands — is no violation; the test seam then takes the anchor out of that edit's `reopened`, the
// check stands again on a wording the report no longer holds, and the oracle reports exactly one
// violation naming the anchor, the check's event, its state and both sentences. An `unread` reading
// never stands, so it is outside the rule either way.
func TestAStandingCheckOfAnEarlierWordingIsAViolation(t *testing.T) {
	kinds := judgedKinds()
	arms := stalenessArms()
	for _, row := range recordsql.EvidenceTables {
		if !hasArm(arms, row.Kind) {
			t.Errorf("backing kind %s has no staleness arm: its checks are never forged against the oracle", row.Kind)
		}
	}
	for name, arm := range arms {
		t.Run(name, func(t *testing.T) {
			row, ok := recordsql.EvidenceTableOf(arm.kind)
			if !ok {
				t.Fatalf("no EvidenceTables row for %s", arm.kind)
			}
			k := kinds[arm.kind]
			runDir, id := k.setup(t)
			sit(t, runDir, k.seat)
			old, err := record.LiveSentence(runtest.Open(t, runDir), id)
			if err != nil {
				t.Fatal(err)
			}
			if out, err := arm.check(t, runDir, id); err != nil {
				t.Fatalf("%s: %v\n%s", row.CheckTable, err, out)
			}
			ev := lastCheckEvent(t, runDir, row, id)
			if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", k.edit[0], "--new", k.edit[1],
				"--reason", "rewrote the checked sentence"); err != nil {
				t.Fatalf("edit: %v\n%s", err, out)
			}
			now, err := record.LiveSentence(runtest.Open(t, runDir), id)
			if err != nil {
				t.Fatal(err)
			}
			if _, stale := staleViolations(t, runDir); len(stale) != 0 {
				t.Fatalf("the edit named %s reopened, so its check no longer stands, and the oracle reports %v", id, stale)
			}

			edits, err := runtest.DropReopened(runtest.Open(t, runDir), id)
			if err != nil || len(edits) != 1 {
				t.Fatalf("the seam took %s out of %v: %v", id, edits, err)
			}
			_, stale := staleViolations(t, runDir)
			if arm.state == "" {
				if len(stale) != 0 {
					t.Errorf("a check that never stands is held to the sentence: %v", stale)
				}
				return
			}
			if len(stale) != 1 {
				t.Fatalf("the oracle reports %d staleness violation(s) of a forged standing %s check, want 1: %v", len(stale), arm.state, stale)
			}
			v := stale[0]
			for _, want := range []string{id, fmt.Sprintf("event %d", ev), arm.state, fmt.Sprintf("%q", old), fmt.Sprintf("%q", now)} {
				if !strings.Contains(v.Detail, want) {
					t.Errorf("the violation does not name %s: %s", want, v)
				}
			}
			if v.Anchor != id {
				t.Errorf("the violation is about %q, want %s", v.Anchor, id)
			}
		})
	}
}

// hasArm is whether arms holds an arm of kind.
func hasArm(arms map[string]stalenessArm, kind string) bool {
	for _, a := range arms {
		if a.kind == kind {
			return true
		}
	}
	return false
}

// A READING GAP'S TOKEN AT THE CHECKED SENTENCE'S END IS NO CHANGE OF WORDING. A `found_against`
// reading mints its gap at the sentence it read, after the check recorded that sentence; with no edit
// the reading stands of the sentence the report carries, and the oracle reports nothing. A raw
// comparison would read the token as a new wording.
func TestAReadingGapTokenIsNoStaleness(t *testing.T) {
	runDir, id := citedClause(t)
	sit(t, runDir, lensSeat)
	if out, err := verifyAs(t, runDir, id, "refutes"); err != nil {
		t.Fatalf("verify --as refutes: %v\n%s", err, out)
	}
	claim := firstVerify(t, runDir).GetClaim()
	live, err := record.LiveSentence(runtest.Open(t, runDir), id)
	if err != nil {
		t.Fatal(err)
	}
	report := readReport(t, runDir)
	if !strings.Contains(report, "<!--gap:") {
		t.Fatalf("the refuting reading placed no gap token in the report:\n%s", report)
	}
	if !strings.Contains(sentenceHolding(report, id), "<!--gap:") {
		t.Fatalf("the reading's gap token is not in the sentence %s marks:\n%s", id, report)
	}
	if claim == "" || live == "" {
		t.Fatalf("claim %q, live sentence %q", claim, live)
	}
	if _, stale := staleViolations(t, runDir); len(stale) != 0 {
		t.Errorf("a gap token placed at the checked sentence reads as a change of wording: %v", stale)
	}
}

// sentenceHolding is the raw line of report holding id's token, anchors included.
func sentenceHolding(report, id string) string {
	for _, ln := range strings.Split(report, "\n") {
		if strings.Contains(ln, id) {
			return ln
		}
	}
	return ""
}

// A CHECK THAT RECORDED NO SENTENCE IS COUNTED, NEVER FOLDED INTO AGREEMENT. A migrated reproduce holds
// no `claim`; the oracle cannot say whether it is of the sentence the report carries, so it reports it
// unmeasured, and no violation.
func TestAJudgementWithNoSentenceIsUnmeasured(t *testing.T) {
	k := judgedKinds()["computation"]
	runDir, id := k.setup(t)
	sit(t, runDir, k.seat)
	if out, err := k.check(t, runDir, id); err != nil {
		t.Fatalf("reproduce: %v\n%s", err, out)
	}
	if res, stale := staleViolations(t, runDir); len(stale) != 0 || len(res.Unmeasured) != 0 {
		t.Fatalf("a judgement of the sentence the report carries: violations %v, unmeasured %v", stale, res.Unmeasured)
	}
	// The judgement as migrate leaves it: the body row with no sentence.
	if _, err := openRunDB(t, runDir).Exec(`UPDATE "reproduce" SET "claim" = NULL WHERE "computation_id" = ?`, id); err != nil {
		t.Fatal(err)
	}
	if err := recordsql.Close(filepath.Join(runDir, "records", "record.db")); err != nil {
		t.Fatal(err)
	}
	res, stale := staleViolations(t, runDir)
	if len(stale) != 0 {
		t.Errorf("a judgement that recorded no sentence is a violation: %v", stale)
	}
	if len(res.Unmeasured) != 1 || !strings.Contains(res.Unmeasured[0], id) || !strings.Contains(res.Unmeasured[0], "not measured") {
		t.Errorf("the oracle's unmeasured = %q, want one entry naming %s as not measured", res.Unmeasured, id)
	}
}
