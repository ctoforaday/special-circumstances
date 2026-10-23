package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// sliceSource is a source record held as its events, for the shapes no archive this binary can
// write holds: a correction chain longer than one link, a chain whose head is missing, an appeal.
type sliceSource []migrate.OldEvent

func (s sliceSource) Events() ([]migrate.OldEvent, error) { return s, nil }
func (s sliceSource) Files() []migration.SourceFile       { return nil }
func (s sliceSource) Close() error                        { return nil }

// oldStream numbers and stamps a source's events in the order they are added.
type oldStream struct{ evs sliceSource }

func (o *oldStream) add(seat, word, key string, fields map[string]any) {
	n := int64(len(o.evs) + 1)
	o.evs = append(o.evs, migrate.OldEvent{ID: n, SeatID: seat, Word: word, Key: key, Fields: fields,
		TS: "2026-10-01T10:00:" + string(rune('0'+n/10)) + string(rune('0'+n%10)) + ".000000000Z"})
}

func oldLog(text string) map[string]any {
	return map[string]any{"text": text, "type": "defect", "source": "seat"}
}

func replaySlice(t *testing.T, src sliceSource, opt migrate.Options) (*migrate.Result, record.Run) {
	t.Helper()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, opt)
	if err != nil {
		t.Fatal(err)
	}
	return res, dst
}

// A CORRECTION CHAIN MIGRATES AS ITS HEAD, IN ITS ROOT'S PLACE, ONCE (plan 10 §II "The migrate
// translation", items 2-4). The record holds no correction: the act the seat meant is the chain's
// last replacement, and it is written where the chain began, under the root's stamp and with the
// root's key, so an act the seat filed after the root stays after it. Every later link and every
// Correction writes nothing, and the manifest counts each in events_in and none in events_out.
func TestACorrectionChainMigratesAsItsHeadInItsRootsPlace(t *testing.T) {
	var o oldStream
	o.add("blue-respond", "register", "blue-respond:register:#1", map[string]any{"agent_id": "b1"})
	o.add("blue-respond", "log", "blue-respond:log:#1", oldLog("first, wrong"))
	o.add("blue-respond", "log", "blue-respond:log:#1~1", oldLog("first, still wrong"))
	o.add("blue-respond", "correction", "blue-respond:correction:blue-respond:log:#1",
		map[string]any{"corrects": "blue-respond:log:#1", "replacement": "blue-respond:log:#1~1", "why": "a word was lost"})
	o.add("blue-respond", "log", "blue-respond:log:#1~2", oldLog("first, right"))
	o.add("blue-respond", "correction", "blue-respond:correction:blue-respond:log:#1~1",
		map[string]any{"corrects": "blue-respond:log:#1~1", "replacement": "blue-respond:log:#1~2", "why": "and another"})
	o.add("blue-respond", "log", "blue-respond:log:#2", oldLog("second"))
	o.add("blue-respond", "revision", "blue-respond:revision:#1", map[string]any{"text": "the sitting's edits"})

	res, dst := replaySlice(t, o.evs, migrate.Options{})
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	if res.In["log"] != 4 || res.Out["log"] != 2 || res.In["correction"] != 2 || res.Out["correction"] != 0 ||
		res.In["revision"] != 1 || res.Out["revision"] != 0 {
		t.Errorf("in %v, out %v; want 4 logs in and 2 out, 2 corrections and a revision in and none out", res.In, res.Out)
	}
	m, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	var logs []string
	for _, e := range m.Events {
		if strings.Contains(e.GetKey(), "~") {
			t.Errorf("the migrated record holds a replacement key %q", e.GetKey())
		}
		if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok {
			logs = append(logs, e.GetKey()+"="+l.GetText()+"@"+e.GetTs())
		}
	}
	want := []string{
		"blue-respond:log:#1=first, right@" + o.evs[1].TS,
		"blue-respond:log:#2=second@" + o.evs[6].TS,
	}
	if strings.Join(logs, " | ") != strings.Join(want, " | ") {
		t.Errorf("logs = %v\nwant %v — the chain's head at the root's place and stamp, then the later act", logs, want)
	}
}

// A CHAIN WHOSE HEAD THE SOURCE DOES NOT HOLD, AND A HEAD THE ROOT'S PLACE REFUSES, ARE REFUSALS
// NAMING THE CHAIN — loud, never a quiet skip and never the root written in the head's stead.
func TestABrokenCorrectionChainIsRefusedByName(t *testing.T) {
	t.Run("the head is absent", func(t *testing.T) {
		var o oldStream
		o.add("blue-respond", "register", "blue-respond:register:#1", map[string]any{"agent_id": "b1"})
		o.add("blue-respond", "log", "blue-respond:log:#1", oldLog("first"))
		o.add("blue-respond", "correction", "blue-respond:correction:blue-respond:log:#1",
			map[string]any{"corrects": "blue-respond:log:#1", "replacement": "blue-respond:log:#1~1", "why": "w"})
		res, _ := replaySlice(t, o.evs, migrate.Options{})
		if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, "blue-respond:log:#1 -> blue-respond:log:#1~1") ||
			!strings.Contains(res.Refusals[0].Err, "head the source does not hold") {
			t.Errorf("refusals = %+v, want one naming the chain and its missing head", res.Refusals)
		}
	})
	t.Run("the root's place refuses the head", func(t *testing.T) {
		var o oldStream
		o.add("blue-respond", "register", "blue-respond:register:#1", map[string]any{"agent_id": "b1"})
		o.add("blue-respond", "log", "blue-respond:log:#1", oldLog("first"))
		o.add("blue-respond", "log", "blue-respond:log:#1~1", oldLog(""))
		o.add("blue-respond", "correction", "blue-respond:correction:blue-respond:log:#1",
			map[string]any{"corrects": "blue-respond:log:#1", "replacement": "blue-respond:log:#1~1", "why": "w"})
		res, _ := replaySlice(t, o.evs, migrate.Options{})
		if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, "the correction chain blue-respond:log:#1 -> blue-respond:log:#1~1 is written as its head at the root's place") {
			t.Errorf("refusals = %+v, want one naming the chain the root's place refused", res.Refusals)
		}
	})
}

// AN APPEAL HAS NO TRANSLATION: a record that holds one fails to migrate unless its operator accepts
// the loss by word, and the manifest then states it.
func TestAnAppealIsALossTheOperatorAccepts(t *testing.T) {
	var o oldStream
	o.add("blue-respond", "register", "blue-respond:register:#1", map[string]any{"agent_id": "b1"})
	o.add("blue-respond", "motion_appeal", "blue-respond:motion_appeal:#1",
		map[string]any{"motion_id": "M1", "subject": "grade", "reason": "the ruling reads past the argument"})
	res, _ := replaySlice(t, o.evs, migrate.Options{})
	if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, `"motion_appeal" is untranslatable`) {
		t.Errorf("refusals = %+v, want the appeal refused as an untranslatable word", res.Refusals)
	}
	res, _ = replaySlice(t, o.evs, migrate.Options{AcceptLoss: map[string]bool{"motion_appeal": true}})
	if len(res.Refusals) != 0 || res.AcceptedLosses["motion_appeal"] == "" || res.Out["motion_appeal"] != 0 {
		t.Errorf("accepted: refusals %+v, losses %v, out %v; want none refused, the loss stated, nothing written", res.Refusals, res.AcceptedLosses, res.Out)
	}
}
