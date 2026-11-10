package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// idOf is the events.id of the first event in runDir whose body is a T that match accepts.
func idOf[T proto.Message](t *testing.T, runDir string, match func(T, string) bool) int64 {
	t.Helper()
	for _, e := range storedOf[T](t, runDir) {
		if body, _ := recordpb.BodyAs[T](e.Ev); match(body, e.Seat) {
			return e.ID
		}
	}
	t.Fatalf("no matching %T in %s", *new(T), runDir)
	return 0
}

// migrated migrates runDir to a fresh directory and returns it with the manifest.
func migrated(t *testing.T, runDir string) (string, map[string]int, map[string]int, []string) {
	t.Helper()
	to := filepath.Join(t.TempDir(), "migrated")
	t.Cleanup(func() { _ = recordsql.CloseUnder(to) })
	man, err := migrate.Migrate(runDir, to, migrate.Entries(), migrate.Options{})
	var refusals []string
	if man != nil {
		for _, r := range man.Refusals {
			refusals = append(refusals, fmt.Sprintf("%d %s: %s", r.OldID, r.Word, r.Err))
		}
	}
	if err != nil && man == nil {
		t.Fatalf("migrate: %v", err)
	}
	return to, man.In, man.Out, refusals
}

// 33. A RECORD WRITTEN AT THIS EPOCH MIGRATES WITH NO REFUSAL: a reading act — Mint, Anchor, Verify,
// written in one transaction by a lens holding gaps of its own — is re-driven whole on replay, each member keeping
// its own duties, and every reference into the act lands on the member it named.
func TestARecordWrittenAtThisEpochMigratesWithNoRefusal(t *testing.T) {
	b := readingRunCast(t, rpReport+"\n# Filler\n\n"+strings.Join(rdFillers, " ")+"\n",
		rpRepeated, "The quarter closed with costs rising steeply.", "Margins held, by the survey's count")
	if _, err := run(t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--key", "E0", "--quote", "costs rising steeply",
		"--new", "costs rising sharply", "--reason", "one wording"); err != nil {
		t.Fatalf("blue's edit: %v", err)
	}
	for _, s := range []string{lensL, lensM, lensN} {
		b.sit(s)
	}
	c := b.cs[1]
	b.fill(lensL)
	cc := "claim_contradicts_source"
	newText := "The quarter closed with costs rising modestly" + anchor.Token(c) + "."
	g := b.mustRead(lensL, citeReading(c, "refutes", cc, "--new", newText)...).Gap.ID
	if r := b.mustRead(lensM, citeReading(c, "refutes", cc)...); r.Gap == nil || r.Gap.ID != g || r.Gap.Minted {
		t.Fatalf("M's reading = %+v, want %s named", r.Gap, g)
	}
	isG := func(m *recordpb.Mint, _ string) bool { return m.GetGapId() == g }
	isGAnchor := func(a *recordpb.Anchor, _ string) bool { return a.GetId() == g }
	isL := func(v *recordpb.Verify, seat string) bool { return seat == lensL && v.GetCarriedBy() == g }
	mintID, anchorID, verifyID := idOf(t, b.runDir, isG), idOf(t, b.runDir, isGAnchor), idOf(t, b.runDir, isL)

	// (1) The chair's plan, whose why names L's reading by event.
	if _, err := run(t, "register", "--run", b.runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "dispatch", "next", "--run", b.runDir, "--seat-id", "red-chair", "--json"); err != nil {
		t.Fatalf("the chair's dispatch: %v", err)
	}
	why := fmt.Sprintf("%s's reading of citation %s at event %d", lensL, c, verifyID)
	planHas := func(dir, want string) bool {
		for _, e := range storedOf[*recordpb.Plan](t, dir) {
			p, _ := recordpb.BodyAs[*recordpb.Plan](e.Ev)
			if strings.Contains(p.GetPrinted(), want) {
				return true
			}
		}
		return false
	}
	if !planHas(b.runDir, why) {
		for _, e := range storedOf[*recordpb.Plan](t, b.runDir) {
			p, _ := recordpb.BodyAs[*recordpb.Plan](e.Ev)
			t.Logf("plan at %d: %s", e.ID, p.GetPrinted())
		}
		t.Fatalf("the stored plan does not name %q", why)
	}
	// (2) A Stale through each member of the act, at the record layer.
	for _, through := range []int64{mintID, anchorID, verifyID} {
		if _, err := record.Append(record.Identity{Run: runtest.Open(t, b.runDir), SeatID: lensM},
			&recordpb.Stale{GapId: proto.String(g), Through: proto.Int64(through)}); err != nil {
			t.Fatalf("stale through %d: %v", through, err)
		}
	}
	// Blue accepts G's verified fix on the repeated sentence: an exact-span edit.
	if err := acceptEdit(t, b.runDir, "A", g, "--accept"); err != nil {
		t.Fatalf("the accept: %v", err)
	}
	if !lastBody(t, b.runDir, &recordpb.BlueEdit{}).GetExactSpan() {
		t.Fatal("the accept is not an exact-span edit, so nothing here exercises one")
	}
	// N mints unsourced_claim from a source it found.
	if r := b.mustRead(lensN, urlReading(rdU3, rdFillers[7], "supports", "unsourced_claim")...); r.Gap == nil || !r.Gap.Minted {
		t.Fatalf("N's found-source reading = %+v, want it to mint", r.Gap)
	}
	openBefore := b.openOf(lensL)
	renderBefore := readReport(t, b.runDir)
	// THE ACTS ABOVE CROSS THE CHECKPOINT SPACING: a reading's placement by sentence_of and by quote,
	// blue's accept at an occurrence and the fill's mints each land ops, and every read through the
	// checkpoints the write path stamped is the whole replay — before the migration and after it.
	if n := checkpointsHold(t, b.runDir); n < 2 {
		t.Fatalf("the acts landed %d report checkpoint(s), so none was stamped past the first op", n)
	}

	to, in, out, refusals := migrated(t, b.runDir)
	if len(refusals) != 0 {
		t.Fatalf("the migration refused:\n%s", strings.Join(refusals, "\n"))
	}
	for _, w := range []string{"mint", "anchor", "verify"} {
		if in[w] == 0 || in[w] != out[w] {
			t.Errorf("%s: %d in, %d out", w, in[w], out[w])
		}
	}
	src, err := record.MergedEvents(runtest.Open(t, b.runDir))
	if err != nil {
		t.Fatal(err)
	}
	dst, err := record.MergedEvents(runtest.Open(t, to))
	if err != nil {
		t.Fatal(err)
	}
	if len(src.Events) != len(dst.Events) {
		t.Fatalf("%d events in, %d out", len(src.Events), len(dst.Events))
	}
	for i := range src.Events {
		s, d := src.Events[i], dst.Events[i]
		// The body, the stamp, the seat and the type are the source's; the key is re-derived by the
		// write the replay drives.
		s, d = proto.Clone(s).(*recordpb.Event), proto.Clone(d).(*recordpb.Event)
		s.Key, d.Key = nil, nil
		if !proto.Equal(s, d) {
			t.Errorf("event %d: %v became %v", i, s, d)
		}
	}
	tb := readingBoard{t: t, runDir: to}
	if got := tb.openOf(lensL); got != openBefore {
		t.Errorf("L's open count %d → %d across the migration", openBefore, got)
	}
	if n := checkpointsHold(t, to); n < 2 {
		t.Fatalf("the migrated record holds %d report checkpoint(s), so its replay stamped none past the first op", n)
	}
	if got := readReport(t, to); got != renderBefore {
		t.Errorf("the migrated render differs:\n%s\n---\n%s", got, renderBefore)
	}
	mVerify := idOf(t, to, isL)
	if !planHas(to, fmt.Sprintf("%s's reading of citation %s at event %d", lensL, c, mVerify)) {
		t.Errorf("the migrated plan does not name L's reading at its migrated event %d", mVerify)
	}
	want := []int64{idOf(t, to, isG), idOf(t, to, isGAnchor), mVerify}
	var got []int64
	for _, e := range storedOf[*recordpb.Stale](t, to) {
		s, _ := recordpb.BodyAs[*recordpb.Stale](e.Ev)
		got = append(got, s.GetThrough())
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("the migrated Stales run through %v, want the act's members %v", got, want)
	}
}

// 34. MIGRATE CHECKS A READING'S GAP BY SENTENCE_OF: with the cited sentence's citation gone from the
// stream, the reading's Anchor names a token the migrated render lacks, and gapplace refuses it; the
// act's Mint and Verify are refused with it, and nothing of the gap lands.
func TestMigrateChecksAReadingsGapBySentenceOf(t *testing.T) {
	b := readingRun(t)
	g := b.mustRead(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...).Gap.ID
	isG := func(m *recordpb.Mint, _ string) bool { return m.GetGapId() == g }
	isGAnchor := func(a *recordpb.Anchor, _ string) bool { return a.GetId() == g }
	isL := func(v *recordpb.Verify, seat string) bool { return seat == lensL }
	mintID, anchorID, verifyID := idOf(t, b.runDir, isG), idOf(t, b.runDir, isGAnchor), idOf(t, b.runDir, isL)
	cite := idOf(t, b.runDir, func(c *recordpb.Cite, _ string) bool { return c.GetLabel() == b.c1 })
	citeAnchor := idOf(t, b.runDir, func(a *recordpb.Anchor, _ string) bool { return a.GetId() == b.c1 })
	dropEvents(t, b.runDir, cite, citeAnchor)

	to, _, _, refusals := migrated(t, b.runDir)
	byID := map[int64]string{}
	for _, r := range refusals {
		var id int64
		fmt.Sscanf(r, "%d", &id)
		byID[id] = r
	}
	if r := byID[anchorID]; !strings.Contains(r, g) || !strings.Contains(r, b.c1) || !strings.Contains(r, anchortext.ErrAnchorGone.Error()) {
		t.Errorf("the Anchor's refusal = %q, want gapplace's sentence_of error naming %s and %s", r, g, b.c1)
	}
	for _, id := range []int64{mintID, verifyID} {
		if byID[id] == "" {
			t.Errorf("event %d of the act was not refused; refusals:\n%s", id, strings.Join(refusals, "\n"))
		}
	}
	if mintOf(t, to, g) != nil || len(anchorsOfGap(t, to, g)) != 0 {
		t.Error("the migrated record holds G's Mint or Anchor")
	}
	for _, v := range verifysOf(t, to) {
		if v.GetCarriedBy() == g {
			t.Error("the migrated record holds a Verify carrying G")
		}
	}
}

// dropEvents deletes the events ids from runDir's record, from every table holding an event_id.
func dropEvents(t *testing.T, runDir string, ids ...int64) {
	t.Helper()
	if err := recordsql.CloseUnder(runDir); err != nil {
		t.Fatal(err)
	}
	db := otherSeat(t, runDir)
	// The scratch record stands for an archive that never held these events; the append-only
	// trigger guards a live record, not this fixture.
	if _, err := db.Exec(`DROP TRIGGER "events_are_append_only_delete"`); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT m.name FROM sqlite_master m WHERE m.type = 'table' AND EXISTS (SELECT 1 FROM pragma_table_info(m.name) WHERE name = 'event_id')`)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, n)
	}
	rows.Close()
	for _, id := range ids {
		for _, tb := range tables {
			if _, err := db.Exec(`DELETE FROM "`+tb+`" WHERE event_id = ?`, id); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := db.Exec(`DELETE FROM events WHERE id = ?`, id); err != nil {
			t.Fatal(err)
		}
	}
}

// checkpointsHold holds runDir's report checkpoints to the whole replay: the report as of every event
// that adds an op, read from the checkpoint at or below it and the ops after it, is the replay of every
// op up to it from the empty text. It returns how many checkpoints the record stores.
func checkpointsHold(t *testing.T, runDir string) int {
	t.Helper()
	ops, err := record.ReportProjection(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	db, err := recordsql.Open(filepath.Join(runDir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	for k, op := range ops {
		if k+1 < len(ops) && ops[k+1].EventID == op.EventID {
			continue // the event's last op closes it
		}
		want, werr := reportproj.Replay("", false, ops[:k+1])
		src, err := recordsql.ReportSourceAt(db, op.EventID)
		if err != nil {
			t.Fatalf("the report as of event %d: %v", op.EventID, err)
		}
		got, gerr := src.Fold(reportproj.Replay)
		if got != want || fmt.Sprint(gerr) != fmt.Sprint(werr) {
			t.Fatalf("the report as of event %d read through its checkpoint is (%q, %v), the whole replay (%q, %v)", op.EventID, got, gerr, want, werr)
		}
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "report_checkpoint"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
