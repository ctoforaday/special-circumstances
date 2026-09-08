package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

const (
	ts1 = "2026-09-02T10:00:01.000000001Z"
	ts2 = "2026-09-02T10:00:02.000000002Z"
	ts3 = "2026-09-02T10:00:03.000000003Z"
	ts4 = "2026-09-02T10:00:04.000000004Z"
)

// The core claim of the whole plan: an old record replays through the CURRENT write path,
// and what lands carries the ORIGINAL stamps — byte for byte, because record.Now is pinned
// and the real stamp() formats it.
func TestReplayLandsOldEventsUnderTheOriginalClock(t *testing.T) {
	old := newOldRecord(t)
	id := old.event(t, "blue-lane-1", 1, ts1, "register")
	old.row(t, "register", `"event_id", "tool_version"`, id, "abc1234")
	id = old.event(t, "blue-lane-1", 1, ts2, "friction")
	old.row(t, "friction", `"event_id", "text", "kind"`, id, "the fetch cache refused a legitimate URL", "defect")
	id = old.event(t, "blue-lane-1", 1, ts3, "friction_none")
	old.row(t, "friction_none", `"event_id", "text"`, id, "clean sitting")

	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals on a fixture built to land: %+v", res.Refusals)
	}
	if res.In["friction"] != 1 || res.Out["log"] != 2 || res.Out["register"] != 1 || res.Out["cast"] != 1 {
		t.Fatalf("counts: in=%v out=%v", res.In, res.Out)
	}

	merged, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	evs := merged.Events
	if len(evs) != 4 {
		t.Fatalf("replayed %d events, want 4 — the three, behind the cast the migration synthesizes", len(evs))
	}
	if evs[0].GetSeatId() != record.HarnessSeat || evs[0].GetTs() != ts1 {
		t.Errorf("the synthesized cast = %s at %s, want the harness's, stamped with the first event's clock %s", evs[0].GetSeatId(), evs[0].GetTs(), ts1)
	}
	for i, want := range []string{ts1, ts2, ts3} {
		if got := evs[i+1].GetTs(); got != want {
			t.Errorf("event %d ts %q, want the ORIGINAL %q — the replay clock leaked", i, got, want)
		}
	}
	if typ := evs[2].GetLog().GetType(); typ != recordpb.LogType_LOG_TYPE_DEFECT {
		t.Errorf("friction kind=defect became %v, want LOG_TYPE_DEFECT", typ)
	}
	if typ := evs[3].GetLog().GetType(); typ != recordpb.LogType_LOG_TYPE_NOMINAL {
		t.Errorf("friction_none became %v, want LOG_TYPE_NOMINAL — the positive empty form", typ)
	}
	if src := evs[2].GetLog().GetSource(); src != recordpb.LogSource_LOG_SOURCE_SEAT {
		t.Errorf("a seat-filed friction became source %v, want SEAT", src)
	}
}

// A word neither the registry nor the current vocabulary knows is a REFUSAL NAMING IT —
// never a skip. The miss and the clean pass must not share bytes.
func TestUnknownWordIsARefusalNamingIt(t *testing.T) {
	old := newOldRecord(t)
	old.db.Exec(`INSERT INTO "enum_event_type" ("value", "means") VALUES ('seance', 'a word no schema ever declared')`)
	old.db.Exec(`CREATE TABLE "seance" ("event_id" INTEGER PRIMARY KEY, "spirit" TEXT)`)
	id := old.event(t, "blue-lane-1", 1, ts1, "seance")
	old.row(t, "seance", `"event_id", "spirit"`, id, "banquo")

	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, `"seance"`) {
		t.Fatalf("an unknown word must refuse by name: %+v", res.Refusals)
	}
}

// An old column the current body does not carry refuses by TABLE AND COLUMN. The identity
// default that silently dropped it would be the plausible-zero shape inside a kept word.
func TestUnmappedColumnOnAKeptWordRefusesByName(t *testing.T) {
	old := newOldRecord(t)
	old.db.Exec(`ALTER TABLE "register" ADD COLUMN "hat_size" TEXT`)
	id := old.event(t, "blue-lane-1", 1, ts1, "register")
	old.row(t, "register", `"event_id", "tool_version", "hat_size"`, id, "abc1234", "L")

	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, "register.hat_size") {
		t.Fatalf("an unmapped column must refuse naming table and column: %+v", res.Refusals)
	}
}

// A documented loss fails the replay unless the operator accepts it BY WORD, and the
// acceptance lands in the result — acknowledged loss is visible loss.
func TestALossIsRefusedUntilAcceptedByWord(t *testing.T) {
	reg := migrate.Registry{
		"seance": {Loss: "the concept has no successor; the record keeps the original"},
	}
	old := newOldRecord(t)
	old.db.Exec(`INSERT INTO "enum_event_type" ("value", "means") VALUES ('seance', 'former')`)
	old.db.Exec(`CREATE TABLE "seance" ("event_id" INTEGER PRIMARY KEY)`)
	old.event(t, "blue-lane-1", 1, ts1, "seance")

	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, reg, dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, "untranslatable") {
		t.Fatalf("an unaccepted loss must be a refusal: %+v", res.Refusals)
	}

	dst2 := runtest.New(t, recordtest.TmpRun(t))
	res, err = migrate.Replay(src, reg, dst2, migrate.Options{AcceptLoss: map[string]bool{"seance": true}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 || res.AcceptedLosses["seance"] == "" {
		t.Fatalf("an accepted loss must move to AcceptedLosses with its reason: %+v / %+v", res.Refusals, res.AcceptedLosses)
	}
}

// The logical source hash is DETERMINISTIC across reads — it is the identity of the record,
// not of a map iteration order.
func TestSourceHashIsStableAcrossReads(t *testing.T) {
	old := newOldRecord(t)
	id := old.event(t, "blue-lane-1", 1, ts1, "friction")
	old.row(t, "friction", `"event_id", "text", "kind"`, id, "x", "friction")

	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	a, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	dst2 := runtest.New(t, recordtest.TmpRun(t))
	b, err := migrate.Replay(src, migrate.Entries(), dst2, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if a.SourceHash == "" || a.SourceHash != b.SourceHash {
		t.Fatalf("source hash must be stable: %q vs %q", a.SourceHash, b.SourceHash)
	}
}
