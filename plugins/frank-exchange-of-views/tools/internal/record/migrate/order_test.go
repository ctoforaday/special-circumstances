package migrate_test

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

var sectionHeading = regexp.MustCompile(`(?m)^## (.+)$`)

func headingsOf(text string) (out []string) {
	for _, m := range sectionHeading.FindAllStringSubmatch(text, -1) {
		out = append(out, withoutStripped(m[1]))
	}
	return out
}

// withoutStripped is s with every token of a kind assembly strips taken out.
func withoutStripped(s string) string {
	return anchor.Replace(s, func(tok, id string) string {
		if anchor.AssemblyOf(id) == anchor.Strip {
			return ""
		}
		return tok
	})
}

// anchorLines is each anchor in text with the line it stands on, anchors taken out.
func anchorLines(text string) map[string]string {
	out := map[string]string{}
	for _, ln := range strings.Split(text, "\n") {
		for _, id := range anchor.IDs(ln) {
			out[id] = strings.TrimSpace(withoutStripped(ln))
		}
	}
	return out
}

// THE MIGRATED M20 HOLDS ITS TEXT IN THE ORDER IT SHIPPED IN. universe-m20's synthesizer wrote
// TL;DR, Open questions, four sections of its own and then the Catechism; assembly shipped the
// Catechism second and Open questions after the tool's sections, so the order every seat read
// was not the reader's. Assembly moves nothing now, and the migrated record holds the text in the
// shipped order: the base reordered, the seventeen edits replayed over it as recorded, every
// anchor on the line it was on.
func TestTheMigratedM20HoldsItsTextInTheOrderItShips(t *testing.T) {
	const archive = "2026-10-11_is-91-prime-m20.tar.gz"
	m, dst := migrateArchive(t, archive)
	if len(m.Refusals) != 0 {
		t.Fatalf("%d event(s) refused migrating m20; the first: %+v", len(m.Refusals), m.Refusals[0])
	}
	if m.ReportOrder != migrate.OrderReordered {
		t.Fatalf("the manifest says the report's order is %q, want %q", m.ReportOrder, migrate.OrderReordered)
	}
	text, err := reportproj.RenderFromRecord(dst)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"TL;DR", "The Catechism", "The answer and the evidence for it", "Which checks decide it", "Alternatives considered", "Limits of the answer", "Open questions"}
	if got := headingsOf(text); !slices.Equal(got, want) {
		t.Errorf("the migrated m20's sections stand as\n  %q\nwant\n  %q", got, want)
	}
	if mis, out := report.OutOfOrderIn(report.ParseText(""), report.ParseText(text)); out {
		t.Errorf("the migrated m20's text is out of order: %+v", mis)
	}

	// THE SAME REPORT, SECTION FOR SECTION. The migrated record's own acts, replayed over the base
	// as the archive holds it, give the text in its author's order; the two differ by whole
	// sections moved and nothing else, and every anchor stands on the line it stood on.
	var written string
	if err := openRO(t, recordtest.RequireArchive(t, archive)).QueryRow(`SELECT "text" FROM "base_ingest"`).Scan(&written); err != nil {
		t.Fatal(err)
	}
	if got, want := headingsOf(written), []string{"TL;DR", "Open questions", "The answer and the evidence for it", "Which checks decide it", "Alternatives considered", "Limits of the answer", "The Catechism"}; !slices.Equal(got, want) {
		t.Fatalf("the archive's base stands as %q; this test reads the order universe-m20's blue wrote, %q", got, want)
	}
	_, _, ops, err := record.ReportProjection(dst)
	if err != nil {
		t.Fatal(err)
	}
	was, err := reportproj.Replay(written, true, ops)
	if err != nil {
		t.Fatalf("the migrated record's acts do not replay over the base as written: %v", err)
	}
	if text != report.Ordered(was) {
		t.Errorf("the migrated text is not the archived text with its sections in order")
	}
	a, b := strings.Split(was, "\n"), strings.Split(text, "\n")
	slices.Sort(a)
	slices.Sort(b)
	if !slices.Equal(slices.DeleteFunc(a, isBlank), slices.DeleteFunc(b, isBlank)) {
		t.Errorf("the migrated text and the archived text do not hold the same lines")
	}
	before, after := anchorLines(was), anchorLines(text)
	if len(before) < 40 || len(after) != len(before) {
		t.Fatalf("the archived text holds %d anchor(s) and the migrated %d; m20 holds over forty", len(before), len(after))
	}
	for id, ln := range before {
		if after[id] != ln {
			t.Errorf("%s stood on %q and stands on %q", id, ln, after[id])
		}
	}

	// AND THE SET SHIPS AS IT DID: the Catechism ahead of the report's own sections, Open questions
	// after the tool's.
	docs, err := shippedDocs(dst)
	if err != nil {
		t.Fatal(err)
	}
	shipped := headingsOf(docs[0].Body)
	cat, own, alt, open := slices.Index(shipped, "The Catechism"), slices.Index(shipped, "The answer and the evidence for it"), slices.Index(shipped, "Research areas"), slices.Index(shipped, "Open questions")
	if cat != 1 || own != 2 || alt < own || open < alt {
		t.Errorf("report.md's sections ship as %q", shipped)
	}
}

func isBlank(s string) bool { return strings.TrimSpace(s) == "" }

const (
	movHead = "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n"
	movCat  = "## The Catechism\n\nThese two factors settle it.\n\n"
	movOwn  = "## Is 91 divisible by a small prime?\n\nYes: 7 divides it.\n\n"
	movOpen = "## Open questions\n\nNone.\n"
)

// sourceRun is a run whose record holds base as its report and the edits over it, written at the
// record layer as an archive's are read: no write there asks about the order.
func sourceRun(t *testing.T, base string, edits ...[2]string) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	run := runtest.Open(t, dir)
	for _, seat := range []string{"blue-synthesize", "blue-respond"} {
		if _, _, err := record.RegisterSeat(record.Identity{Run: run, SeatID: seat}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	if base == "" {
		return dir
	}
	if _, err := record.Append(record.Identity{Run: run, SeatID: "blue-synthesize"}, &recordpb.BaseIngest{Text: proto.String(base)}); err != nil {
		t.Fatal(err)
	}
	for i, e := range edits {
		if _, err := record.Append(record.Identity{Run: run, SeatID: "blue-respond"}, &recordpb.BlueEdit{
			EditKey: proto.String("E" + string(rune('1'+i))), Old: proto.String(e[0]), New: proto.String(e[1]), Text: proto.String("a change"),
		}); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// THE BASE IS REORDERED ONLY WHERE EVERY ACT REPLAYS OVER IT TO THE SAME REPORT. Otherwise the
// record is replayed as written: nothing is refused, nothing is lost, the text keeps its author's
// order and the manifest says so.
func TestMigrateReordersTheBaseOnlyWhereEveryActReplays(t *testing.T) {
	outOfOrder := movHead + movOwn + movCat + movOpen
	for name, tc := range map[string]struct {
		base  string
		edits [][2]string
		want  string
	}{
		"no report": {"", nil, migrate.OrderNoReport},
		"in order":  {movHead + movCat + movOwn + movOpen, [][2]string{{"7 divides it", "7 divides it, 13 times"}}, migrate.OrderInOrder},
		"out of order, edits inside sections": {outOfOrder,
			[][2]string{{"7 divides it", "7 divides it, 13 times"}, {"These two factors", "The two factors"}}, migrate.OrderReordered},
		// The quote runs from one section into the heading of the next, and the two are no longer
		// neighbours once the Catechism stands in its place.
		"an edit across two sections that move apart": {outOfOrder,
			[][2]string{{"Yes: 7 divides it.\n\n## The Catechism", "Yes: 7 divides it, 13 times.\n\n## The Catechism"}}, migrate.OrderAsWritten},
		// The edit writes a section after the one that closes the text: reordering the base leaves
		// that section where the edit put it.
		"an edit that adds a section out of order": {outOfOrder,
			[][2]string{{"None.", "None.\n\n## Appendix\n\nMore arithmetic."}}, migrate.OrderAsWritten},
		"in order until an edit adds a section out of order": {movHead + movCat + movOwn + movOpen,
			[][2]string{{"None.", "None.\n\n## Appendix\n\nMore arithmetic."}}, migrate.OrderAsWritten},
	} {
		t.Run(name, func(t *testing.T) {
			from := sourceRun(t, tc.base, tc.edits...)
			to := recordtest.TmpRun(t)
			m, err := migrate.Migrate(from, to, migrate.Entries(), migrate.Options{})
			if err != nil {
				t.Fatalf("migrate: %v", err)
			}
			if m.ReportOrder != tc.want {
				t.Errorf("the manifest says the report's order is %q, want %q", m.ReportOrder, tc.want)
			}
			// THE TRIAL LEAVES NOTHING. Where the reordered replay is not kept it is rolled back, and the
			// record holds exactly the events the replay as written reports writing — one base, not two.
			written := 0
			for _, n := range m.Out {
				written += n
			}
			var events, bases int
			if err := openRO(t, to).QueryRow(`SELECT count(*), (SELECT count(*) FROM "base_ingest") FROM "events"`).Scan(&events, &bases); err != nil {
				t.Fatal(err)
			}
			if events != written || (tc.base != "" && bases != 1) {
				t.Errorf("the migrated record holds %d event(s) and %d base(s); the manifest reports writing %d", events, bases, written)
			}
			if tc.base == "" {
				return
			}
			was, err := reportproj.RenderFromRecord(runtest.Open(t, from))
			if err != nil {
				t.Fatal(err)
			}
			got, err := reportproj.RenderFromRecord(runtest.Open(t, to))
			if err != nil {
				t.Fatal(err)
			}
			_, out := report.OutOfOrderIn(report.ParseText(""), report.ParseText(got))
			switch tc.want {
			case migrate.OrderReordered:
				if got != report.Ordered(was) || out {
					t.Errorf("the migrated text is not the source's with its sections in order:\n%s", got)
				}
			case migrate.OrderAsWritten:
				if got != was || !out {
					t.Errorf("the migrated text is not the source's as written:\n%s", got)
				}
			default:
				if got != was || out {
					t.Errorf("a text in order did not migrate as written:\n%s", got)
				}
			}
			if n := m.Out["blue_edit"]; n != len(tc.edits) || m.In["blue_edit"] != n {
				t.Errorf("%d edit(s) in and %d out, of %d", m.In["blue_edit"], n, len(tc.edits))
			}
		})
	}
}

// A REPLAY THAT REFUSES AN ACT IS NOT THE SAME REPORT, whatever its text reads: the reordered
// record is kept only when nothing is refused. Here the refused event is one no order of the text
// could land, so both replays refuse it; the migration fails as it would have, and the record it
// leaves for review holds the text as written.
func TestAReorderedReplayThatRefusesAnActIsNotKept(t *testing.T) {
	old := newOldRecord(t)
	for _, ddl := range []string{
		`INSERT INTO "enum_event_type" ("value", "means") VALUES ('base_ingest', 'the report, frozen'), ('seance', 'a word no schema ever declared')`,
		`CREATE TABLE "base_ingest" ("event_id" INTEGER PRIMARY KEY, "text" TEXT)`,
		`CREATE TABLE "seance" ("event_id" INTEGER PRIMARY KEY, "spirit" TEXT)`,
	} {
		if _, err := old.db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	old.row(t, "register", `"event_id", "tool_version"`, old.event(t, "blue-synthesize", 0, ts1, "register"), "0.1.0")
	written := movHead + movOwn + movCat + movOpen
	old.row(t, "base_ingest", `"event_id", "text"`, old.event(t, "blue-synthesize", 0, ts1, "base_ingest"), written)
	old.row(t, "seance", `"event_id", "spirit"`, old.event(t, "blue-synthesize", 0, ts1, "seance"), "banquo")

	to := recordtest.TmpRun(t)
	m, err := migrate.Migrate(old.dir, to, migrate.Entries(), migrate.Options{})
	if err == nil || m == nil || len(m.Refusals) != 1 {
		t.Fatalf("migrating a record that holds an unknown word: manifest %+v, error %v", m, err)
	}
	if m.ReportOrder != migrate.OrderAsWritten {
		t.Errorf("the manifest says the report's order is %q, want %q", m.ReportOrder, migrate.OrderAsWritten)
	}
	if got, err := reportproj.RenderFromRecord(runtest.Open(t, to)); err != nil || got != written {
		t.Errorf("the record left for review does not hold the text as written (%v):\n%s", err, got)
	}
}

// A TRIAL WHOSE BATCH FAILS FAILS THE MIGRATION. The failure is the batch's, not a refusal of the
// reordering: replaying as written over it would publish a record the failed batch says nothing
// about, so the migration returns the failure and publishes nothing.
func TestAFailedTrialBatchFailsTheMigration(t *testing.T) {
	from := sourceRun(t, movHead+movOwn+movCat+movOpen, [2]string{"7 divides it", "7 divides it, 13 times"})
	to := recordtest.TmpRun(t)
	savepoints := 0
	t.Cleanup(func() { record.SavepointTestHook = nil })
	record.SavepointTestHook = func(stmt string) error {
		if strings.HasPrefix(stmt, "SAVEPOINT") {
			// The fifth act lands after the report's base: the trial has a base to reorder, so a
			// failure here is the case where replaying as written would otherwise be tried.
			if savepoints++; savepoints == 5 {
				return errors.New("disk I/O error")
			}
		}
		return nil
	}
	m, err := migrate.Migrate(from, to, migrate.Entries(), migrate.Options{})
	if !errors.Is(err, record.ErrBatchFailed) {
		t.Fatalf("a migration whose trial batch failed answered %v (manifest %+v); want the batch's failure", err, m)
	}
	if m != nil {
		t.Errorf("a migration whose trial batch failed returned a manifest, report order %q", m.ReportOrder)
	}
	if _, err := os.Stat(filepath.Join(to, "inputs", migration.ManifestName)); !os.IsNotExist(err) {
		t.Errorf("a migration whose trial batch failed published a manifest at --to (%v)", err)
	}
}

// A migration hands back the clock it was given: every pass pins record.Now to its events, and
// none of those stamps may outlive it. (That each pass also STARTS from the caller's clock is held
// by replay's own restore; it shows only for a source whose first stamp does not parse, which no
// fixture has.)
func TestAMigrationRestoresTheClock(t *testing.T) {
	from := sourceRun(t, movHead+movOwn+movCat+movOpen, [2]string{"7 divides it", "7 divides it, 13 times"})
	fixed := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	was := record.Now
	record.Now = func() time.Time { return fixed }
	t.Cleanup(func() { record.Now = was })
	if _, err := migrate.Migrate(from, recordtest.TmpRun(t), migrate.Entries(), migrate.Options{}); err != nil {
		t.Fatal(err)
	}
	if got := record.Now(); !got.Equal(fixed) {
		t.Errorf("after a migration record.Now answers %v, want the caller's %v — a pass left its pinned stamp behind", got, fixed)
	}
}

// shippedDocs is the run's report set as it ships: the run assembled, and each document it wrote — its
// name, and its .md, masthead and all, in Body — in reading order.
func shippedDocs(run record.Run) ([]report.Doc, error) {
	if _, err := report.Assemble(run); err != nil {
		return nil, err
	}
	var docs []report.Doc
	for _, f := range report.Files() {
		b, err := os.ReadFile(filepath.Join(run.Dir(), f))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, report.Doc{File: f, Body: string(b)})
	}
	return docs, nil
}
