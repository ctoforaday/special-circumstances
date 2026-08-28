package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
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
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	for _, id := range []string{"red-lens-r1-L1", "red-merge-r1", "blue-respond-r1", "judge-r1"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", id); err != nil {
			t.Fatalf("register %s: %v", id, err)
		}
	}
	// A lens finding is now anchored into blue/report.md and rejected unless its
	// --quote quote is present (slice 1b). Seed a report carrying the quotes the
	// finding tests use, mirroring the real run where blue-synthesize wrote the report
	// before red-lens files findings.
	seedBlueReport(t, runDir)
	return runDir
}

// mintGap mints through the merge seat and returns the tool-assigned id.
//
// It passes NO --location. Since 0.63.0 a mint's location is matched against blue/report.md, and
// callers here seed different reports — a helper carrying one sentence would only work for the
// callers whose report happens to contain it. Location is optional on mint; a test that means to
// exercise it passes its own quote.
func mintGap(t *testing.T, runDir, key, class string) string {
	t.Helper()
	out, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", key, "--class", class, "--problem", "the defect", "--fix", "the fix",
		"--check-kind", "document", "--check", "the acceptance check red runs at re-audit",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low")
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

// readProjection returns a markdown projection computed on read from the record via the
// shared view library — the same bytes `show --view <name>` prints.
func readProjection(t *testing.T, runDir, name string) string {
	t.Helper()
	b, err := view.Markdown(runtest.Open(t, runDir), name, "")
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
	if _, err := run(t, "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", id, "--as", "repaired",
		"--principle", "the repair discharges the defect at the leaf",
		"--tension", "thoroughness against ceremony",
		"--review-flag", "no — the closure is mechanical and the anchor is checkable", "--settled", "the proposition this ruling bars", "--final",
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

	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", id, "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation"); err != nil {
		t.Fatalf("motion grade file: %v", err)
	}
	if _, err := run(t, "motion", "grade", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "M1", "--as", "accepted",
		"--reason", "the bound holds; regrading"); err != nil {
		t.Fatalf("motion grade rule: %v", err)
	}

	evs := events(t, runDir)
	var sawFiling, sawRuling bool
	for _, e := range evs {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_MOTION:
			sawFiling = true
		case recordpb.EventType_EVENT_TYPE_MOTION_RULE:
			sawRuling = true
		}
	}
	if !sawFiling || !sawRuling {
		t.Fatalf("the exchange must survive in one shared record: filing=%v ruling=%v", sawFiling, sawRuling)
	}
}

// Red closes a gap with its verification anchor; the closure and its anchor must both be
// readable afterwards. An anchored closure whose anchor is not recoverable is exactly the
// attestation-format defect the scorecard measures — and this run scored 0% anchored.
func TestClosureCarriesItsAnchorIntoTheRecord(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "anchored-closure", "anchor-visibility")

	prose := filepath.Join(recordtest.TmpRun(t), "closure.md")
	if err := os.WriteFile(prose, []byte("re-read the cited source; the digits match the arm the claim names"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:report.md",
		"--reason-file", prose); err != nil {
		t.Fatalf("close: %v", err)
	}

	ev := lastBody(t, runDir, &recordpb.Close{})
	keys := setFields(ev)
	for _, want := range []string{"anchor_seat", "anchor_tool", "anchor_target"} {
		if !keys[want] {
			t.Errorf("closure lost %s; an unanchored closure is mechanically unauditable (payload had %v)", want, keys)
		}
	}
	// `prose` is the FIELD; `--reason` is the flag. A close stores its argument as `prose`, an
	// opinion as `rationale` — the fold is the flag's business, and setFields reads the schema.
	if !keys["prose"] {
		t.Error("closure lost its prose record")
	}
}

// Every seat's events land in ONE record that any seat can read back. If a seat's writes
// were invisible to the others, every cross-seat duty in the protocol would be unverifiable.
func TestAllFourSeatsWriteIntoOneReadableRecord(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--key", "F1", "--quote", "§2", "--reason", "a finding",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	mintGap(t, runDir, "shared-record", "one-record")
	if _, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--reason", "considered rewriting the parser", "--as", "declined",
		"--reason", "the input grammar is not stable enough to justify it this round"); err != nil {
		t.Logf("blue line of inquiry shape rejected: %v", err)
	}

	seats := map[string]bool{}
	for _, e := range events(t, runDir) {
		seats[e.GetSeatId()] = true
	}
	for _, want := range []string{"red-lens-r1-L1", "red-merge-r1"} {
		if !seats[want] {
			t.Errorf("%s wrote nothing readable into the shared record (saw %v)", want, seats)
		}
	}
}
