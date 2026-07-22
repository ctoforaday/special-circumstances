package cli

import (
	"strings"
	"testing"
)

// EVERY CROSS-REFERENCE, CHECKED.
//
// An audit of the whole surface found TWELVE OF TWELVE references unvalidated: only
// `close --id` and `mint --supersedes` were checked, and every other field naming a gap,
// an observation, a finding or a seat was accepted on the seat's say-so.
//
// That is the mechanism behind the 2026-07-18 run's worst damage. Eight judicial closures
// were `opinion --id` events ACCEPTED at write time and silently DROPPED at replay,
// because replay checked what the write path did not. The board was wrong by six gaps for
// three rounds with nothing on the surface saying so.
//
// The split is the defect: a reference replay will refuse must be refused when it is
// WRITTEN, while the seat is still there to fix it. An event accepted into the log and
// discarded on the way out looks recorded and does nothing.

func TestEveryCrossReferenceIsCheckedAtWriteTime(t *testing.T) {
	runDir := seatRun(t)
	real := mintGap(t, runDir, "the-real-gap", "reference-integrity")
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--reason", "a real observation"); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name, wants string
		args        []string
	}{
		{"opinion --id", "no mint event created", []string{"bench", "opinion", "--seat-id", "judge-r1",
			"--id", "R9-9", "--as", "carried", "--principle", "p", "--tension", "t", "--review-flag", "no"}},
		{"dispute --id", "no mint event created", []string{"blue", "dispute", "--seat-id", "blue-respond-r1",
			"--id", "R9-9", "--dimension", "severity", "--proposed", "low", "--reason", "b"}},
		{"dispute-respond --id", "no mint event created", []string{"merge", "dispute-respond", "--seat-id", "red-merge-r1",
			"--id", "R9-9", "--as", "accepted", "--reason", "b"}},
		{"regrade --id", "no mint event created", []string{"merge", "regrade", "--seat-id", "red-merge-r1",
			"--id", "R9-9", "--severity", "low", "--reason", "b"}},
		{"closing --id", "no mint event created", []string{"merge", "closing", "--seat-id", "red-merge-r1",
			"--id", "R9-9", "--reason", "t"}},
		{"manifest-row --id", "no mint event created", []string{"blue", "manifest-row", "--seat-id", "blue-respond-r1",
			"--id", "R9-9", "--row", "r"}},
		{"dispose --observation", "no observe or finding event", []string{"merge", "dispose", "--seat-id", "red-merge-r1",
			"--observation", "NOPE", "--as", "declined", "--reason", "r"}},
		{"dispose --into", "no mint event created", []string{"merge", "dispose", "--seat-id", "red-merge-r1",
			"--observation", "L1-O1", "--as", "folded-into", "--into", "R9-9"}},
		{"close --successor", "no mint event created", []string{"merge", "close", "--seat-id", "red-merge-r1",
			"--id", real, "--as", "closed", "--anchor-seat", "L1", "--anchor-tool", "t",
			"--anchor-target", "x", "--successor", "R9-9"}},
		{"mint --found-by", "no lens recorded", []string{"merge", "mint", "--seat-id", "red-merge-r1",
			"--key", "k2", "--class", "reference-integrity", "--location", "l", "--problem", "p",
			"--fix", "f", "--check", "c", "--severity", "medium", "--likelihood", "medium",
			"--impact", "medium", "--cx", "low", "--found-by", "L9-F9"}},
		{"spot-check --ids", "no mint event created", []string{"merge", "spot-check", "--seat-id", "red-merge-r1",
			"--ids", "R9-9"}},
		{"petition-rule --petitioner", "has recorded nothing in this run", []string{"bench", "petition-rule",
			"--seat-id", "judge-r1", "--petitioner", "ghost-seat", "--petition-class", "safety",
			"--as", "granted", "--reason", "o"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			args := append([]string{c.args[0], c.args[1], "--run", runDir}, c.args[2:]...)
			_, err := run(t, args...)
			if err == nil {
				t.Fatalf("%s accepted a reference to something that does not exist — it would be dropped at replay, and the seat would never know", c.name)
			}
			if !strings.Contains(err.Error(), c.wants) {
				t.Errorf("the refusal must say WHY the reference does not resolve, got: %v", err)
			}
		})
	}
}

// And the converse, or the test above would pass for a validate that refuses everything.
func TestValidReferencesStillResolve(t *testing.T) {
	runDir := seatRun(t)
	first := mintGap(t, runDir, "first", "reference-integrity")
	second := mintGap(t, runDir, "second", "reference-integrity")
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--reason", "o"); err != nil {
		t.Fatal(err)
	}

	for _, c := range [][]string{
		{"bench", "opinion", "--seat-id", "judge-r1", "--id", first, "--as", "carried",
			"--principle", "p", "--tension", "t", "--review-flag", "no", "--reason", "the ruling"},
		{"merge", "dispose", "--seat-id", "red-merge-r1", "--observation", "L1-O1",
			"--as", "folded-into", "--into", second},
		{"merge", "close", "--seat-id", "red-merge-r1", "--id", first, "--as", "closed",
			"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x", "--successor", second, "--reason", "verified"},
	} {
		args := append([]string{c[0], c[1], "--run", runDir}, c[2:]...)
		if _, err := run(t, args...); err != nil {
			t.Errorf("a reference that DOES resolve was refused: %v (%v)", err, c[1])
		}
	}
}

// STATE, not just existence. A reference can resolve and still be wrong: the thing exists,
// but not in a state where the act means anything.
//
// Every rule here was measured against the 2026-07-18 run before being written, and one
// candidate was DROPPED by that check: "supersedes must name a closed gap" sounded right
// and would have refused NINE OF NINE real mints, because superseding a still-open
// predecessor is the normal pattern. The rules that survived had zero violations.
func TestActsAreRefusedOnTheWrongState(t *testing.T) {
	runDir := seatRun(t)
	open := mintGap(t, runDir, "stays-open", "state-checks")
	closed := mintGap(t, runDir, "gets-closed", "state-checks")
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", closed, "--as", "closed", "--anchor-seat", "L1", "--anchor-tool", "go test",
		"--anchor-target", "./x", "--reason", "closed"); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name, wants string
		args        []string
	}{
		{"close a gap twice", "double-counts closure history", []string{"merge", "close",
			"--seat-id", "red-merge-r1", "--id", closed, "--as", "closed",
			"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x"}},
		{"regrade a closed gap", "changes a number nobody reads", []string{"merge", "regrade",
			"--seat-id", "red-merge-r1", "--id", closed, "--severity", "low", "--reason", "b"}},
		{"dispute a closed gap", "disposition has already been made", []string{"blue", "dispute",
			"--seat-id", "blue-respond-r1", "--id", closed, "--dimension", "severity",
			"--proposed", "low", "--reason", "b"}},
		{"answer a dispute nobody filed", "no dispute was filed", []string{"merge", "dispute-respond",
			"--seat-id", "red-merge-r1", "--id", open, "--as", "accepted", "--reason", "b"}},
		{"spot-check an OPEN gap", "still OPEN", []string{"merge", "spot-check",
			"--seat-id", "red-merge-r1", "--ids", open}},
		{"carry residue into a closed gap", "already finished", []string{"merge", "close",
			"--seat-id", "red-merge-r1", "--id", open, "--as", "closed",
			"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x",
			"--successor", closed}},
	} {
		t.Run(c.name, func(t *testing.T) {
			args := append([]string{c.args[0], c.args[1], "--run", runDir}, c.args[2:]...)
			_, err := run(t, args...)
			if err == nil {
				t.Fatalf("%s was accepted", c.name)
			}
			if !strings.Contains(err.Error(), c.wants) {
				t.Errorf("the refusal must explain what the state makes meaningless, got: %v", err)
			}
		})
	}
}

// The rule the evidence KILLED. Nine of nine mints in the run superseded a still-open
// predecessor, so requiring a closed one would have refused every real lineage claim.
func TestSupersedingAnOpenGapIsNormal(t *testing.T) {
	runDir := seatRun(t)
	ancestor := mintGap(t, runDir, "the-ancestor", "state-checks")
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "the-successor", "--class", "state-checks", "--location", "l",
		"--problem", "p", "--fix", "f", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--cx", "low",
		"--supersedes", ancestor); err != nil {
		t.Errorf("superseding an OPEN gap must be accepted — it is what 9 of 9 real mints did: %v", err)
	}
}

// ONE FINDING, ONE FATE — checkable only now that identity is assigned.
func TestAFindingCannotBeGivenASecondFate(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--reason", "worth noticing")
	if err != nil {
		t.Fatal(err)
	}
	id := findingID.FindString(out)
	if _, err := run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", id, "--as", "declined", "--reason", "checked at the leaf"); err != nil {
		t.Fatal(err)
	}
	_, err = run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", id, "--as", "banked", "--reason", "changed my mind")
	if err == nil {
		t.Fatal("a finding was given a second fate; the first is already in the counts and the projection silently shows whichever replayed last")
	}
	if !strings.Contains(err.Error(), "friction") {
		t.Errorf("the refusal must point at the channel for a missing capability, since there is no verb for revising a disposal: %v", err)
	}
}

// SUPERSEDING IS A PROMISE TO REPLACE, and the promise is due at the seat's terminal act.
//
// INVESTIGATED, NOT ASSUMED. 9 mints in the 2026-07-18 run superseded an open gap, and the
// frequency first read as proof it was intended. Tracing each one splits them:
//
//	7 are STRUCTURALLY REQUIRED — the protocol mints the successor, then closes the
//	  ancestor naming it, so the ancestor MUST still be open at mint time.
//	2 are a DEFECT — R3-1 superseded R2-1 and R2-5 and closed neither, so all three
//	  finished open. The run reported 9 open gaps; 7 were distinct.
//
// So the rule is not about mint-time state, which would refuse all 9. It is a completion
// duty, and this test holds both halves: the 7 stay legal, the 2 are caught.
func TestVerdictRefusesWhileASupersededGapIsStillOpen(t *testing.T) {
	runDir := seatRun(t)
	ancestor := mintGap(t, runDir, "the-ancestor", "lineage-completion")

	// THE LEGAL 7: mint a successor while the ancestor is still open.
	out, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "the-successor", "--class", "lineage-completion", "--location", "l",
		"--problem", "p", "--fix", "f", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--cx", "low",
		"--supersedes", ancestor)
	if err != nil {
		t.Fatalf("superseding an OPEN gap is what 7 of 9 real mints did, and the protocol requires it: %v", err)
	}
	successor := gapID(out)

	// THE DEFECT: finish without closing the ancestor.
	_, err = run(t, "merge", "verdict", "--run", runDir, "--seat-id", "red-merge-r1", "--as", "PASS")
	if err == nil {
		t.Fatal("the seat finished with a superseded gap still open — the same defect counted twice, which is how a board reported 9 open gaps for 7 distinct ones")
	}
	for _, want := range []string{ancestor, successor} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the stranded ancestor AND what claimed to replace it, got: %v", err)
		}
	}

	// KEEPING THE PROMISE clears it.
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", ancestor, "--as", "closed", "--successor", successor,
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./x",
		"--reason", "replaced by its successor"); err != nil {
		t.Fatal(err)
	}
	// The successor is now the live gap; PASS still requires it closed (the all-gaps guard).
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", successor, "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./y",
		"--reason", "successor resolved"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "verdict", "--run", runDir, "--seat-id", "red-merge-r1", "--as", "PASS"); err != nil {
		t.Errorf("with the ancestor and successor both closed the verdict must go through: %v", err)
	}
}
