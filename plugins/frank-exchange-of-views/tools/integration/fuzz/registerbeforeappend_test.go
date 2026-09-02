package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A SEAT REGISTERS BEFORE IT APPENDS — pinned deterministically, because the sweep only hit this
// about once in two hundred runs and a rate that low is indistinguishable from a flake (#664).
//
// closeGap is driven from the RED-MERGE branch and proves as `blue-respond-r1`. `blue prove`
// RECORDS a `prove` event, so that is an append — and in round 1 red-merge-r1 runs before
// blue-respond-r1 has ever registered. Nothing on that path called register() at all, which is
// why the #656 waiter did not help: it ordered a seat against ITSELF for callers that went
// through register, and this caller never did.
//
// The reproduction removes the rarity rather than chasing it: mint a computation gap, close it
// with ONLY the red seat registered, and ask `verify` — the same invariant the sweep's oracle
// runs. Before the fix this fails every time; the sweep needed a computation gap to become
// closable in round 1, which is what made it rare.
func TestClosingAComputationGapRegistersTheProvingSeatFirst(t *testing.T) {
	bin := buildBinary(t)
	// recordtest.TmpRun, NOT t.TempDir. This test opens a record, and the record layer caches a
	// *sql.DB per run: on Linux an open file can still be unlinked so a missed release is
	// invisible, and on Windows TempDir's own RemoveAll cannot delete the locked record.db and
	// the test fails in cleanup. TmpRun exists precisely for that, and its comment notes it had
	// already been copied eight times without its release — this was the ninth, caught by the
	// windows leg of CI and by nothing local.
	runDir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"),
		[]byte("# § fuzz\n\nA § fuzz sentence to anchor findings.\n\nThe cost is rising over time.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatalf("resolving the run: %v", err)
	}
	if err := record.StageForRun(stageRun, fuzzClasses...); err != nil {
		t.Fatalf("staging the class registry: %v", err)
	}

	r := newRunner(bin, runDir, newLockedRand(7))
	// ONLY THE ACTING SEAT. blue-respond-r1 is deliberately left unregistered: the whole point is
	// that closeGap appends as it, so closeGap is what must register it.
	r.register("merge", "red-merge-r1")

	// `mint` draws the gap KIND at random, so take gaps until a computation one appears. Bounded
	// so a change that stops producing them fails here rather than hanging.
	var gapID string
	for i := 0; i < 40 && gapID == ""; i++ {
		id := r.mint("red-merge-r1")
		if id != "" && r.computationGaps[id] {
			gapID = id
		}
	}
	if gapID == "" {
		t.Fatal("no computation gap after 40 mints — this test cannot exercise the prove path, and passing on that would be a green that measured nothing")
	}
	if r.registered["blue-respond-r1"] {
		t.Fatal("blue-respond-r1 is already registered before closeGap ran — the setup no longer reproduces the ordering this pins")
	}

	r.closeGap("red-merge-r1", gapID, false)

	out, _ := r.exec("verify", "--seat-id", "operator")
	// THE VERDICT, NOT THE LABEL. `verify` lists every invariant by name whether it passed or
	// failed, so `Contains(out, "register-before-append")` matches a CLEAN run too — it read as
	// a failure on the first attempt at this test, against a fix that was working. A check keyed
	// on a name that is always present measures nothing; the marker is what carries the verdict.
	// AND THE INVARIANT MUST ACTUALLY HAVE RUN. `verify` marks one `[n/a ]` when the run gave it
	// nothing to judge, and an n/a reads exactly like a pass to any substring check — so assert
	// that it was EVALUATED before trusting what it says.
	ok := strings.Contains(out, "[ok  ] register-before-append")
	failed := strings.Contains(out, "[FAIL] register-before-append")
	if !ok && !failed {
		t.Fatalf("verify did not EVALUATE register-before-append on this run, so this test proved nothing:\n%s", out)
	}
	if failed {
		t.Errorf("verify reports register-before-append after closing a computation gap:\n%s\n\n"+
			"closeGap proves as blue-respond-r1, and `blue prove` records an event, so that seat must be registered before the proof is appended", out)
	}
}
