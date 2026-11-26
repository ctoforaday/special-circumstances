package cli

import (
	"io"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/deliverytest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"regexp"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
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

// seatRun sets up a run directory with every seat registered, the way the engine does: on the cast
// setup writes, which the chair's dispatch plans against (a bench convened on a docket, conveneBench).
func seatRun(t *testing.T) string {
	return seatRunReport(t, "")
}

// seatRunReport is seatRun with the round-0 report body a caller needs. body=="" seeds the generic
// report carrying the quotes the tests anchor to; a non-empty body is ingested instead, for
// a test whose --quote is specific. Either way the report is INGESTED (#709), mirroring the real run
// where blue-synthesize wrote and froze the report before red-lens mints gaps — so a caller must
// NOT writeReport again afterwards, which would be a second, refused ingest.
func seatRunReport(t *testing.T, body string) string {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	// The cast names every seat on the roster, so a test may register any lens beside these four.
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: seatclass.SeatIDs()}))
	for _, id := range []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"} {
		// The bench owes an OCCASION and no other seat may pass one — ask the same question the
		// write path asks rather than keeping a list of which ids are the bench.
		args := []string{"register", "--run", runDir, "--seat-id", id}
		if record.SeatOwesOccasion(id) {
			args = append(args, "--occasion", "docket")
		}
		if _, err := run(t, args...); err != nil {
			t.Fatalf("register %s: %v", id, err)
		}
	}
	if body == "" {
		seedBlueReport(t, runDir)
	} else {
		writeReport(t, runDir, body)
	}
	return runDir
}

// mintGap mints through the LENS seat and returns the tool-assigned id.
//
// Mint is a lens verb (plans/roundless.md §III.B.3): the chair coalesces nothing and puts nothing
// on the board itself. Every fixture here mints from ONE lens, `lensSeat`, so that the same seat
// can close or regrade the gap later — the originator closes, and any other seat is refused.
//
// It passes NO --location. Since 0.63.0 a mint's location is matched against blue/report.md, and
// callers here seed different reports — a helper carrying one sentence would only work for the
// callers whose report happens to contain it. Location is optional on mint; a test that means to
// exercise it passes its own quote.
func mintGap(t *testing.T, runDir, key, class string) string {
	t.Helper()
	// The chair sits before anyone mints: the epoch is the count of chair sittings, and a gap's
	// minted_epoch is read by the convergence checks. Then the lens sits, because it is the seat
	// that mints.
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	out, err := run(t, "mint", "--about-kind", "section", "--about", "Findings", "--run", runDir, "--seat-id", lensSeat,
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

// runMint is `mint` with args.
func runMint(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return run(t, append([]string{"mint"}, args...)...)
}

// gapID pulls the tool-assigned id out of a mint's output.
func gapID(out string) string {
	return regexp.MustCompile(`G-[0-9a-f]{8}`).FindString(out)
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
// The 2026-07-18 run's red-chair reported: "the verdict render reports 9 open, 9
// closed against the hand-written board's 3 open / 15 closed. The difference is exactly
// the six gaps judge closed." Bench dispositions lived in the judge's event stream and
// nothing carried them into red's projection, so the board over-reported open gaps by the
// number of bench closures after every sitting and diverged further each round.
//
// If this test fails, the projection cannot be made authoritative: flipping the ledger to
// render would silently lose every judicial closure.
func TestBenchClosureIsVisibleToRedsBoard(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "bench-closes-this", "spec_underspecification")

	// The bench rules the gap closed — a docket motion filed and ruled.
	//
	// NEVER skip on a failure in here. A skip would excuse exactly the defect this test exists
	// to catch, and a suite that excuses its own subject is how the projection shipped blind to
	// bench closures in the first place. benchDisposes fails the test rather than skipping.
	benchDisposes(t, runDir, id, "repaired", "the repair discharges the defect at the leaf")

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
		t.Errorf("gap %s is STILL IN THE OPEN SET after the bench closed it — the defect red-chair reported, where the board over-reports open gaps by exactly the number of bench closures and diverges further every round", id)
	}
}

// A gap's assignee closes it with its verification anchor; the closure and its anchor
// must both be readable afterwards. An anchored closure whose anchor is not recoverable is exactly the
// attestation-format defect the scorecard measures — and this run scored 0% anchored.
func TestClosureCarriesItsAnchorIntoTheRecord(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "anchored-closure", "risk_coverage_omission")

	const prose = "re-read the cited source; the digits match the arm the claim names"
	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat,
		"--id", id, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:report.md",
		"--reason", prose); err != nil {
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
	mintGap(t, runDir, "shared-record", "doctrine_vs_implementation")
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond",
		"--question", "considered rewriting the parser", "--as", "declined",
		"--reason", "the input grammar is not stable enough to justify it this round"); err != nil {
		t.Logf("blue inquiry shape rejected: %v", err)
	}

	seats := map[string]bool{}
	for _, e := range events(t, runDir) {
		seats[e.GetSeatId()] = true
	}
	for _, want := range []string{"red-lens-evidence", "red-chair"} {
		if !seats[want] {
			t.Errorf("%s wrote nothing readable into the shared record (saw %v)", want, seats)
		}
	}
}

// lensSeat is the ONE lens every fixture in this package mints from. A gap belongs to the lens
// that minted it for its whole life — close and regrade are refused to any other seat — so a
// fixture that mints here and disposes there has to be the same seat at both ends.
const lensSeat = "red-lens-evidence"

// registerChairOnce makes sure red-chair has registered on this run — ONCE, whatever path did it.
// The epoch is the count of chair sittings, so a fixture with no chair on the record writes every
// gap into epoch 0, and one that registered blindly on a run whose staged board had already seated
// the chair opened epoch 2. Both are wrong. The record is asked, not a Go map: the map cannot see a
// register the board builder wrote.
func registerChairOnce(t *testing.T, runDir string) {
	t.Helper()
	registerSeatOnce(t, runDir, "red-chair")
}

// registerLensOnce seats `lensSeat` — ONCE, whatever path did it — and completes its delivery of
// the whole report where one is ingested, as a lens's first sitting does. A seat registers before it
// acts, and a gap is its assignee's to close, so every mint helper here sits the same lens rather
// than a fresh one per call. The delivery is the lens's pin (report_read, written
// by the delivery writer as it hands the last item): a lens that sat and was never delivered the report
// is still ready at the head and owes its first whole delivery, and every PASS these fixtures drive
// waits on it.
func registerLensOnce(t *testing.T, runDir string) {
	t.Helper()
	registerSeatOnce(t, runDir, lensSeat)
	deliverIfOwed(t, runDir, lensSeat)
}

// lensRereads opens `lensSeat`'s next sitting and completes its delivery there: the whole report once
// more, its full re-read. A lens is delivered the whole report at its first sitting and again at its
// next, and a PASS is held (whole_report_owed) while a lens that has sat lacks either, so a fixture
// that drives a PASS sits its lens through both.
func lensRereads(t *testing.T, runDir string) {
	t.Helper()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatalf("register %s for its full re-read: %v", lensSeat, err)
	}
	deliverIfOwed(t, runDir, lensSeat)
}

// deliverIfOwed completes seatID's delivery in its open sitting where its list holds report text it is
// owed — the state the delivery writer leaves once it has handed the last item. A seat owed none (no
// report yet, its sitting already holds its delivery) is left as it stands.
func deliverIfOwed(t *testing.T, runDir, seatID string) {
	t.Helper()
	deliverytest.DeliverIfOwed(t, runtest.Open(t, runDir), seatID)
}

func registerSeatOnce(t *testing.T, runDir, seatID string) {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err == nil {
		for _, e := range m.Events {
			if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER && e.GetSeatId() == seatID {
				return
			}
		}
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seatID); err != nil {
		t.Fatalf("register %s: %v", seatID, err)
	}
}

// sitToAssemble opens the bench's assembly sitting, the one sitting whose outcome the write takes.
func sitToAssemble(t *testing.T, runDir string) {
	t.Helper()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "assemble"); err != nil {
		t.Fatalf("registering the bench's assembly sitting: %v", err)
	}
}

// writerHands runs the shipped delivery writer's next phase for seatID under agent, as the PostToolUse
// hook runs it after each of the seat's calls, until the head of the seat's queue is no
// report text: one item per call, and the report_read written as the last is handed. It returns the
// number of items handed. An agent the record does not bind is resolved by the seat's configuration.
func writerHands(t *testing.T, runDir, agent, seatID string) int {
	t.Helper()
	owes := func() bool {
		w, err := record.WorkOfSeat(runtest.Open(t, runDir), seatclass.RoleOfSeat(seatID), seatID)
		if err != nil {
			t.Fatal(err)
		}
		q := record.QueueOf(w)
		return len(q) > 0 && q[0].Kind == record.ItemReportText
	}
	for n := 0; n < 200; n++ {
		if !owes() {
			return n
		}
		if err := sittingwrite.WriteNextItem(runDir, agent, seatclass.Seats[seatID].AgentType, string(hookcore.EngineClaude), io.Discard); err != nil {
			t.Fatalf("the writer's next phase for %s: %v", seatID, err)
		}
	}
	t.Fatalf("%s is still owed report text after 200 items: the writer stopped handing it", seatID)
	return 0
}
