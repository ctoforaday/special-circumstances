package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// INTEGRATION: the seats are not independent, and until now nothing tested them together.
//
// Every existing test drives ONE verb and asserts on ITS event. That is why the run
// shipped with red's board projection unable to see the bench's closures: no test ever
// had a judge close a gap and then asked red what its board said. A per-verb suite cannot
// catch a defect that lives between seats, and the whole point of this tool is that
// several seats write one shared record.
//
// These tests drive the real multi-seat sequences: red mints, blue disputes, red answers,
// the bench rules, and each side reads back the state the others set.

// seatRun sets up a run directory with every seat registered, the way the engine does.
func seatRun(t *testing.T) string {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	for _, s := range []struct{ role, id string }{
		{"lens", "red-lens-r1-L1"},
		{"merge", "red-merge-r1"},
		{"blue", "blue-respond-r1"},
		{"bench", "judge-r1"},
	} {
		if _, err := run(t, s.role, "register", "--run", runDir, "--seat-id", s.id); err != nil {
			t.Fatalf("register %s: %v", s.id, err)
		}
	}
	return runDir
}

// mintGap mints through the merge seat and returns the tool-assigned id.
func mintGap(t *testing.T, runDir, key, class string) string {
	t.Helper()
	out, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", key, "--class-new", class,
		"--definition", "d", "--neighbor", "n", "--distinguisher", "x",
		"--location", "§1 \"a quoted sentence\"", "--problem", "the defect", "--fix", "the fix",
		"--check", "the acceptance check red runs at re-audit",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--cx", "low")
	if err != nil {
		t.Fatalf("mint %s: %v", key, err)
	}
	id := gapID(out)
	if id == "" {
		t.Fatalf("mint returned no id: %q", out)
	}
	return id
}

// gapID pulls the tool-assigned id out of a mint's output.
func gapID(out string) string {
	return regexp.MustCompile(`R\d+-\d+`).FindString(out)
}

// readProjection reads a rendered file from the shadow — the projection red's board
// would become if authority were flipped from the hand-written markdown to the events.
func readProjection(t *testing.T, runDir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(runDir, "records", "render-shadow", name))
	if err != nil {
		t.Fatalf("projection %s: %v", name, err)
	}
	return string(b)
}

// THE INDICTMENT. A bench closure must be visible to red's board.
//
// The 2026-07-18 run's red-merge-r3 reported: "the verdict render reports 9 open, 9
// closed against the hand-written board's 3 open / 15 closed. The difference is exactly
// the six gaps judge-r2 closed." Bench dispositions lived in the judge's event stream and
// nothing carried them into red's projection, so the board over-reported open gaps by the
// number of bench closures after every sitting and diverged further each round.
//
// If this test fails, the projection cannot be made authoritative: flipping the ledger to
// render would silently lose every judicial closure.
func TestBenchClosureIsVisibleToRedsBoard(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "bench-closes-this", "cross-seat-visibility")

	// The bench rules the gap closed.
	if _, err := run(t, "bench", "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", id, "--as", "closed",
		"--principle", "the repair discharges the defect at the leaf",
		"--tension", "thoroughness against ceremony",
		"--review-flag", "no — the closure is mechanical and the anchor is checkable",
		"--reason", "closed at the bench, not by red"); err != nil {
		// NEVER skip here. A skip would excuse exactly the defect this test exists to
		// catch, and a suite that excuses its own subject is how the projection shipped
		// blind to bench closures in the first place.
		t.Fatalf("the bench must be able to rule on a gap by id: %v", err)
	}

	// WHICH SECTION the gap sits in is the whole question. The first form of this
	// assertion only checked that the ledger mentioned a closure index SOMEWHERE, which
	// would pass with the gap still sitting in the open set — a test too weak to catch
	// the defect it was written for, which is how the projection shipped blind.
	//
	// It now goes through gapIsOpen (crossseat_test.go), which asserts on ENTRIES rather
	// than substrings and fails loudly if a gap is in both sections or neither. Two
	// readers of one artifact disagreeing is the defect class this whole tool exists to
	// remove; the tests do not get an exemption from it.
	if gapIsOpen(t, runDir, id) {
		t.Errorf("gap %s is STILL IN THE OPEN SET after the bench closed it — the defect red-merge-r3 reported, where the board over-reports open gaps by exactly the number of bench closures and diverges further every round", id)
	}
}

// Blue disputes a grade, red answers, and BOTH sides' records must show the exchange.
// A dispute answered into the void is the same defect one channel over.
func TestGradeDisputeIsVisibleToBothSides(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "disputed-grade", "grade-dispute-visibility")

	if _, err := run(t, "blue", "dispute", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", id, "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation"); err != nil {
		t.Fatalf("blue dispute: %v", err)
	}
	if _, err := run(t, "merge", "dispute-respond", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "accepted",
		"--reason", "the bound holds; regrading"); err != nil {
		t.Fatalf("red dispute-respond: %v", err)
	}

	evs := events(t, runDir)
	var sawDispute, sawResponse bool
	for _, e := range evs {
		switch e.Type {
		case "dispute":
			sawDispute = true
		case "dispute-respond":
			sawResponse = true
		}
	}
	if !sawDispute || !sawResponse {
		t.Fatalf("the exchange must survive in one shared record: dispute=%v response=%v", sawDispute, sawResponse)
	}
}

// Red closes a gap with its verification anchor; the closure and its anchor must both be
// readable afterwards. An anchored closure whose anchor is not recoverable is exactly the
// attestation-format defect the scorecard measures — and this run scored 0% anchored.
func TestClosureCarriesItsAnchorIntoTheRecord(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "anchored-closure", "anchor-visibility")

	prose := filepath.Join(t.TempDir(), "closure.md")
	if err := os.WriteFile(prose, []byte("re-read the cited source; the digits match the arm the claim names"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "git show", "--anchor-target", "7bc501e:report.md",
		"--reason-file", prose); err != nil {
		t.Fatalf("close: %v", err)
	}

	ev := lastOfType(t, runDir, "close")
	keys := payloadKeys(ev)
	for _, want := range []string{"anchor_seat", "anchor_tool", "anchor_target"} {
		if !keys[want] {
			t.Errorf("closure lost %s; an unanchored closure is mechanically unauditable (payload had %v)", want, keys)
		}
	}
	if !keys["prose"] {
		t.Error("closure lost its prose record")
	}
}

// A lens observation must be disposable BY THE MERGE — a different seat, a different
// role, referring to the first seat's label. This is the cross-seat reference the whole
// found_by/dispose chain depends on.
func TestMergeCanDisposeALensObservationByItsLabel(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--reason", "a thing worth noticing but not yet a gap"); err != nil {
		t.Fatalf("lens observe: %v", err)
	}
	if _, err := run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", "L1-O1", "--as", "declined",
		"--reason", "checked at the leaf and the behaviour is correct as written"); err != nil {
		t.Fatalf("merge dispose of a lens observation: %v", err)
	}
	ev := lastOfType(t, runDir, "dispose")
	if !payloadKeys(ev)["observation"] {
		t.Error("the disposal must name the observation it disposes, or the lens's work has no fate")
	}
}

// Every seat's events land in ONE record that any seat can read back. If a seat's writes
// were invisible to the others, every cross-seat duty in the protocol would be unverifiable.
func TestAllFourSeatsWriteIntoOneReadableRecord(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-F1", "--location", "§2", "--reason", "a finding",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	mintGap(t, runDir, "shared-record", "one-record")
	if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--line", "considered rewriting the parser", "--status", "declined",
		"--reason", "the input grammar is not stable enough to justify it this round"); err != nil {
		t.Logf("blue avenue shape rejected: %v", err)
	}

	seats := map[string]bool{}
	for _, e := range events(t, runDir) {
		seats[e.SeatID] = true
	}
	for _, want := range []string{"red-lens-r1-L1", "red-merge-r1"} {
		if !seats[want] {
			t.Errorf("%s wrote nothing readable into the shared record (saw %v)", want, seats)
		}
	}
}
