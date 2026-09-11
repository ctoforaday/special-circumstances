package migrate_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

func seatLog(text string) *recordpb.Log {
	return &recordpb.Log{Text: proto.String(text), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
}

func appendOK(t *testing.T, id record.Identity, body proto.Message) *record.Event {
	t.Helper()
	ev, err := record.Append(id, body)
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

func correcting(id record.Identity, typ recordpb.EventType, key string) record.Identity {
	id.Correct = &record.Correct{Type: typ, Key: key, Why: "a word was lost"}
	return id
}

// liveTexts is what a reader of the acts that stand sees, per event word, in record order.
func liveTexts(t *testing.T, run record.Run) map[string][]string {
	t.Helper()
	m, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	for _, e := range record.Live(m.Events) {
		switch b := mustBody(e).(type) {
		case *recordpb.Log:
			out["log"] = append(out["log"], b.GetText())
		case *recordpb.Position:
			out["position"] = append(out["position"], b.GetText())
		case *recordpb.ManifestRow:
			out["manifest_row"] = append(out["manifest_row"], b.GetGapId()+": "+b.GetRow())
		}
	}
	return out
}

func mustBody(e *record.Event) proto.Message {
	b, _ := recordpb.Body(e)
	return b
}

// assertCorrectionsResolve holds every correction on a migrated record to naming two acts THAT
// RECORD carries — the struck one and its replacement — and returns how many there are.
func assertCorrectionsResolve(t *testing.T, run record.Run) int {
	t.Helper()
	m, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	for _, e := range m.Events {
		keys[e.GetKey()] = true
	}
	n := 0
	for _, e := range m.Events {
		if c, ok := recordpb.BodyAs[*recordpb.Correction](e); ok {
			n++
			if !keys[c.GetCorrects()] || !keys[c.GetReplacement()] {
				t.Errorf("a migrated correction names %s -> %s, and the migrated record does not carry both", c.GetCorrects(), c.GetReplacement())
			}
		}
	}
	return n
}

// A CORRECTION CHAIN SURVIVES MIGRATION AS A CHAIN: every replacement lands as the correction of
// its target's migrated key, and a reader of the acts that stand sees what it saw before.
func TestMigrationCarriesACorrectionChain(t *testing.T) {
	srcDir := recordtest.TmpRun(t)
	src := runtest.New(t, srcDir)
	blue := record.Identity{Run: src, SeatID: "blue-respond"}
	if _, _, err := record.RegisterSeat(blue, ""); err != nil {
		t.Fatal(err)
	}
	k := appendOK(t, blue, seatLog("the tool  refused")).GetKey()
	r1 := appendOK(t, correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k), seatLog("the tool refused the cite"))
	appendOK(t, correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, r1.GetKey()), seatLog("the tool refused the cite, twice"))
	p := appendOK(t, blue, &recordpb.Position{Text: proto.String("the board is  going in")}).GetKey()
	appendOK(t, correcting(blue, recordpb.EventType_EVENT_TYPE_POSITION, p), &recordpb.Position{Text: proto.String("the board is clean going in")})
	want := liveTexts(t, src)

	toDir := recordtest.TmpRun(t)
	res, err := migrate.Migrate(srcDir, toDir, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("migrate: %v (refusals %+v)", err, resRefusals(res))
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	for _, w := range []string{"log", "position", "correction", "register"} {
		if res.In[w] != res.Out[w] {
			t.Errorf("%s: %d in, %d out", w, res.In[w], res.Out[w])
		}
	}
	if res.In["correction"] != 3 {
		t.Errorf("the source held %d corrections, want 3", res.In["correction"])
	}
	dst := runtest.Open(t, toDir)
	if n := assertCorrectionsResolve(t, dst); n != 3 {
		t.Errorf("the migrated record holds %d corrections, want 3", n)
	}
	got := liveTexts(t, dst)
	for w, texts := range want {
		if strings.Join(got[w], " | ") != strings.Join(texts, " | ") {
			t.Errorf("%s: the migrated record's standing acts read %q, the source's %q", w, got[w], texts)
		}
	}
}

func resRefusals(res *migrate.Manifest) any {
	if res == nil {
		return nil
	}
	return res.Refusals
}

// A CORRECTION WHOSE REPLACEMENT IS NOT ON THE SOURCE IS A REFUSAL, never a quiet skip.
func TestAnOrphanCorrectionIsRefusedByName(t *testing.T) {
	srcDir := recordtest.TmpRun(t)
	src := runtest.New(t, srcDir)
	blue := record.Identity{Run: src, SeatID: "blue-respond"}
	if _, _, err := record.RegisterSeat(blue, ""); err != nil {
		t.Fatal(err)
	}
	k := appendOK(t, blue, seatLog("an act")).GetKey()
	recordtest.Seed(t, srcDir, recordtest.At(t, "blue-respond", "blue-respond:correction:"+k,
		&recordpb.Correction{Corrects: proto.String(k), Replacement: proto.String(k + "~1"), Why: proto.String("w")}))
	res, _ := migrate.Migrate(srcDir, recordtest.TmpRun(t), migrate.Entries(), migrate.Options{})
	if res == nil {
		t.Fatal("no manifest")
	}
	found := false
	for _, r := range res.Refusals {
		if r.Word == "correction" && strings.Contains(r.Err, "did not land as a correction") {
			found = true
		}
	}
	if !found {
		t.Errorf("an orphan correction was not refused by name: %+v", res.Refusals)
	}
}

// THE REAL RECORD, ROUND TRIP: B3 migrates clean; a seat corrects a manifest row on the migrated
// record — the act #861's B3 could not repair — and that record migrates again with the pair intact.
func TestB3ArchiveRoundTripsWithACorrection(t *testing.T) {
	tarball := archivePath(t, "2026-09-10_is-91-prime-b3.tar.gz")
	runDir := t.TempDir()
	untar(t, tarball, runDir)
	mid := recordtest.TmpRun(t)
	res, err := migrate.Migrate(runDir, mid, migrate.Entries(), migrate.Options{})
	if err != nil || len(res.Refusals) != 0 {
		t.Fatalf("B3 did not migrate clean: %v %+v", err, resRefusals(res))
	}
	run := runtest.Open(t, mid)
	blue := record.Identity{Run: run, SeatID: "blue-respond"}
	if _, _, err := record.RegisterSeat(blue, ""); err != nil {
		t.Fatal(err)
	}
	k := appendOK(t, blue, &recordpb.ManifestRow{GapId: proto.String("G2"), Row: proto.String("G2 is reproducible via ")}).GetKey()
	appendOK(t, correcting(blue, recordpb.EventType_EVENT_TYPE_MANIFEST_ROW, k),
		&recordpb.ManifestRow{GapId: proto.String("G2"), Row: proto.String("G2 is reproducible via the recorded proof")})
	want := liveTexts(t, run)

	to := recordtest.TmpRun(t)
	res2, err := migrate.Migrate(mid, to, migrate.Entries(), migrate.Options{})
	if err != nil || len(res2.Refusals) != 0 {
		t.Fatalf("the corrected record did not migrate clean: %v %+v", err, resRefusals(res2))
	}
	in, out := 0, 0
	for _, n := range res2.In {
		in += n
	}
	for _, n := range res2.Out {
		out += n
	}
	if in != out || res2.Out["correction"] != 1 {
		t.Errorf("%d in, %d out, %d correction(s) — want every event carried and the one correction with it", in, out, res2.Out["correction"])
	}
	dst := runtest.Open(t, to)
	assertCorrectionsResolve(t, dst)
	if got := liveTexts(t, dst)["manifest_row"]; strings.Join(got, " | ") != strings.Join(want["manifest_row"], " | ") {
		t.Errorf("manifest rows standing after the round trip = %q, want %q", got, want["manifest_row"])
	}
}
