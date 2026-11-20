package cli

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// servedRead is one report read on the record: the lens, the head and the sitting it is in.
type servedRead struct {
	seat          string
	head, sitting int64
}

// readsOf is every report read on runDir's record, in order.
func readsOf(t *testing.T, runDir string) []servedRead {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var out []servedRead
	for _, e := range m.Events {
		if r, ok := recordpb.BodyAs[*recordpb.ReportRead](e); ok {
			out = append(out, servedRead{seat: e.GetSeatId(), head: r.GetHead(), sitting: m.At.Of(e).SittingID})
		}
	}
	return out
}

// headOf is runDir's report head.
func headOf(t *testing.T, runDir string) int64 {
	t.Helper()
	_, head, err := record.ReportAtHead(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	return head
}

const servedReport = "# Costs\n\nCosts rose sharply in the first quarter.\n\nVolume grew.\n"

// NO READ OF THE REPORT IS RECORDED; A DELIVERY IS, WHEN THE WRITER HANDS ITS LAST ITEM. `show report`
// serves the report — whole, at an anchor, or a selection, to any seat — and records nothing and moves
// no pin. The tool records a report_read only as the delivery writer hands the last item of a seat's
// delivery, at the head the delivery was cut at, in the seat's sitting; a head move after it is a
// change the seat is delivered at its next sitting, never a second read in this one.
func TestNoReadOfTheReportIsRecordedAndTheWritersLastItemIs(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, servedReport)
	// The mint helper seats the lens and completes its first delivery, so the reads below are in a
	// sitting that holds one; the counts are taken after it.
	id, err := mintQuote(t, runDir, "k1", "Costs rose sharply")
	if err != nil {
		t.Fatal(err)
	}
	edit := func(key, old, new string) int64 {
		t.Helper()
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", key,
			"--quote", old, "--new", new, "--reason", "say how it grew"); err != nil {
			t.Fatal(err)
		}
		return headOf(t, runDir)
	}
	show := func(seat string, extra ...string) string {
		t.Helper()
		out, err := run(t, append([]string{"show", "report", "--run", runDir, "--seat-id", seat}, extra...)...)
		if err != nil {
			t.Fatalf("show report %v as %s: %v", extra, seat, err)
		}
		return out
	}
	pin := func() int64 {
		t.Helper()
		p, err := record.PinAt(openRunDB(t, runDir), lensSeat)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	before, pinBefore := readsOf(t, runDir), pin()
	head := edit("E1", "Volume grew.", "Volume grew steadily.")
	for _, args := range [][]string{
		{"--anchor", id}, {"--anchor", id, "--json"}, {"--match", "Volume"}, {"--quote", "Volume grew"},
		{"--json", "--match", "Volume"}, {"--json", "--quote", "Volume grew"}, {}, {"--json"},
	} {
		show(lensSeat, args...)
	}
	show("red-chair")
	show(blueSeat, "--json")
	if out := show(lensSeat); out != readReport(t, runDir) {
		t.Fatalf("the lens's bare read is not the report:\n%s", out)
	}
	if got := readsOf(t, runDir); len(got) != len(before) || pin() != pinBefore {
		t.Fatalf("a read of the report recorded %+v and moved the lens's pin %d -> %d", got[len(before):], pinBefore, pin())
	}

	// The lens's next sitting owes its full re-read: the writer hands it, item by item.
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatal(err)
	}
	if n := writerHands(t, runDir, "agent_lens_delivered", lensSeat); n == 0 {
		t.Fatal("the writer handed the lens no item, so the read below measures nothing")
	}
	got := readsOf(t, runDir)[len(before):]
	if len(got) != 1 || got[0].seat != lensSeat || got[0].head != head || got[0].sitting == 0 {
		t.Fatalf("reads = %+v, want one by %s at head %d in its sitting", got, lensSeat, head)
	}

	edit("E2", "Volume grew steadily.", "Volume grew steadily all year.")
	show(lensSeat, "--json")
	if n := writerHands(t, runDir, "agent_lens_delivered", lensSeat); n != 0 {
		t.Errorf("the writer handed %d item(s) of report text after the head moved, in a sitting that holds its delivery", n)
	}
	if got = readsOf(t, runDir)[len(before):]; len(got) != 1 {
		t.Fatalf("reads after the head moved = %+v, want the one delivery and no second read in the sitting", got)
	}
}

// A READ AFTER THE RUN'S OUTCOME SERVES THE REPORT AND RECORDS NOTHING: the record takes no seat
// act once the run has its outcome, and a read is not one. Once the run is captured it is served
// too, and still records nothing.
func TestAReadAfterTheOutcomeRecordsNothing(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, servedReport)
	if _, _, err := record.RegisterSeat(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, "", "assemble"); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"},
		&recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("ended")}); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat)
	if err != nil || out != readReport(t, runDir) {
		t.Fatalf("a read after the outcome = %q, %v; want the report", out, err)
	}
	if got := readsOf(t, runDir); len(got) != 0 {
		t.Errorf("a read after the outcome recorded %+v", got)
	}
	if err := record.MarkCaptured(runtest.Open(t, runDir), "2026-10-24T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	out, err = run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat)
	if err != nil || out != readReport(t, runDir) {
		t.Fatalf("a read after the capture = %q, %v; want the report", out, err)
	}
	if got := readsOf(t, runDir); len(got) != 0 {
		t.Errorf("a read after the capture recorded %+v", got)
	}
}

// A READ UNDER ANOTHER WRITER'S LOCK IS SERVED AND RECORDS NOTHING: the read writes nothing, so a lock
// held past the busy budget does not stand in its way. A READ THE RECORD CANNOT ANSWER IS REFUSED AND
// SERVES NOTHING: with the report's own rows unreadable, the read fails loudly and prints no text. The
// budget is shortened so a wait, if the read took one, would cost a fraction of a second.
func TestAReadTheRecordCannotAnswerIsRefusedAndServesNothing(t *testing.T) {
	const budget = 300 * time.Millisecond
	t.Cleanup(recordsql.SetBusyBudgetForTest(budget))
	runDir := newRun(t)
	writeReport(t, runDir, servedReport)
	if err := recordsql.CloseUnder(runDir); err != nil {
		t.Fatal(err)
	}
	holder, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(runDir, "records", "record.db"))+"?_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Close()
	tx, err := holder.Begin()
	if err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat)
	_ = tx.Rollback()
	if err != nil || out != readReport(t, runDir) {
		t.Fatalf("a read under another writer's lock = %q, %v; want the report, served", out, err)
	}
	if got := readsOf(t, runDir); len(got) != 0 {
		t.Errorf("a read recorded %+v", got)
	}

	if _, err := openRunDB(t, runDir).Exec(`DROP TABLE "report_part"`); err != nil {
		t.Fatal(err)
	}
	out, err = run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat)
	if err == nil || out != "" {
		t.Fatalf("a read the record cannot answer = %q, %v; want a refusal and no text", out, err)
	}
}
