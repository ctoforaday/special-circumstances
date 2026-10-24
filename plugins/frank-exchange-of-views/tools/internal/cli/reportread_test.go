package cli

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

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

// THE WHOLE REPORT SERVED TO A LENS IS RECORDED, ONCE PER HEAD IN A SITTING, AND NOTHING ELSE IS. Both
// whole forms record it; a read at an anchor, a selection and the assembled set serve part of it or
// none of its text, and record nothing; no other seat's read is a lens's pin. A head move is a new
// read.
func TestTheWholeReportServedToALensIsRecordedAndNothingElseIs(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, servedReport)
	// The mint helper seats the lens and serves it the report, as an audit does, so the head moves
	// before the reads below: a read at the head the lens was already served records nothing anyway.
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
	before := readsOf(t, runDir)
	head := edit("E1", "Volume grew.", "Volume grew steadily.")
	for _, partial := range [][]string{
		{"--anchor", id}, {"--anchor", id, "--json"}, {"--match", "Volume"}, {"--quote", "Volume grew"},
		{"--json", "--match", "Volume"}, {"--json", "--quote", "Volume grew"}, {"--assembled"}, {"--assembled", "--json"},
	} {
		show(lensSeat, partial...)
	}
	show("red-chair")
	show(blueSeat, "--json")
	if got := readsOf(t, runDir); len(got) != len(before) {
		t.Fatalf("a read at an anchor, a selection, the assembled set or another seat's whole read recorded %+v", got[len(before):])
	}

	if out := show(lensSeat); out != readReport(t, runDir) {
		t.Fatalf("the lens's bare read is not the report:\n%s", out)
	}
	show(lensSeat, "--json")
	show(lensSeat)
	got := readsOf(t, runDir)[len(before):]
	if len(got) != 1 || got[0].seat != lensSeat || got[0].head != head || got[0].sitting == 0 {
		t.Fatalf("reads = %+v, want one by %s at head %d in its sitting", got, lensSeat, head)
	}

	moved := edit("E2", "Volume grew steadily.", "Volume grew steadily all year.")
	show(lensSeat, "--json")
	if got = readsOf(t, runDir)[len(before):]; len(got) != 2 || got[1].head != moved || got[1].sitting != got[0].sitting {
		t.Fatalf("reads after the head moved to %d = %+v, want a second read at it in the same sitting", moved, got)
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

// A READ THE RECORD CANNOT TAKE IS REFUSED, NOT SERVED UNRECORDED. Another writer holds the lock past
// the busy budget: the read fails loudly, prints nothing, and nothing is on the record.
func TestAReadTheRecordCannotTakeIsRefusedAndServesNothing(t *testing.T) {
	if testing.Short() {
		t.Skip("waits out the record's busy budget")
	}
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
	if err == nil || out != "" {
		t.Fatalf("a read under another writer's lock = %q, %v; want a refusal and no text", out, err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "lock") && !strings.Contains(strings.ToLower(err.Error()), "busy") {
		t.Errorf("the refusal does not say the record was locked: %v", err)
	}
	if got := readsOf(t, runDir); len(got) != 0 {
		t.Errorf("a refused read recorded %+v", got)
	}
}
