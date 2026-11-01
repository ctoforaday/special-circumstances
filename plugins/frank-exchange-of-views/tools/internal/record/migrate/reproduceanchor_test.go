package migrate

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A REPRODUCE BEFORE EPOCH 28 IS CARRIED TO EVERY ANCHOR ON ITS SCRIPT AT THAT POINT (gblock,
// 2026-10-12): reproduction per sha is kept, and an old check's soundness goes to every computation
// on the bundle written before it. The first such anchor carries the re-run (`ran`), and every other
// one ran nothing and states that re-run's comparison, with the same soundness and note. A
// computation written after the reproduce gets nothing: it was never judged.
//
// AN ARCHIVED `not_run` IS AN ACT THAT RAN NOTHING (epoch 49): it states the comparison of its
// bundle's standing re-run — the latest reproduce before it that ran the bundle.
//
// Asserted on every archive, in order, against the source read directly: the n-th source
// reproduce becomes exactly the expected run of migrated reproduces, each on the bundle its
// script's computation was written with.
func TestEveryArchivedReproduceIsCarriedToTheAnchorsOnItsScript(t *testing.T) {
	type carried struct {
		bundle, rerun string
		ran           bool
		soundness     string
	}
	fanned, shared, kept, stood := 0, 0, 0, 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			computationsOn := map[string]int{} // bundle -> computations written so far
			bundleOf := map[string]string{}    // archived script sha -> bundle
			standing := map[string]string{}    // bundle -> the comparison of the latest reproduce that ran it
			var want []carried
			var named []string // per want: the anchor a reproduce names as written, "" for one carried
			for _, old := range evs {
				switch old.Word {
				case "proof":
					sha, _ := old.Fields["proof_sha"].(string)
					if sha == "" {
						sha, _ = old.Fields["sha256"].(string) // the shard era's spelling
					}
					script, _ := old.Fields["script"].(string)
					_, _, bundle, err := archivedBundle(script, sha)
					if err != nil {
						t.Fatal(err)
					}
					bundleOf[sha] = bundle
					computationsOn[bundle]++
				case "reproduce":
					sha, _ := old.Fields["proof_sha"].(string)
					bundle := bundleOf[sha]
					// A reproduce written at epoch 28 or later names its anchor and passes as written.
					if id, _ := old.Fields["proof_id"].(string); id != "" {
						rerun, _ := old.Fields["rerun"].(string)
						snd, _ := old.Fields["soundness"].(string)
						ran := rerun != "not_run"
						if ran {
							standing[bundle] = rerun
						} else {
							rerun = standing[bundle]
							stood++
						}
						want = append(want, carried{bundle, rerun, ran, snd})
						// The anchor it named keeps its hex under the computation's letter.
						named = append(named, "K-"+id[len("P-"):])
						continue
					}
					rerun := recordpb.Word(recordpb.Rerun_RERUN_DIFFERS)
					if v := old.Fields["reproduced"]; v == int64(1) || v == true {
						rerun = recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED)
					}
					snd, _ := old.Fields["soundness"].(string)
					for i := 0; i < computationsOn[bundle]; i++ {
						if i > 0 {
							shared++
						} else {
							standing[bundle] = rerun
						}
						want = append(want, carried{bundle, rerun, i == 0, snd})
						named = append(named, "")
					}
				}
			}
			_, dir := MigratedArchive(t, name)
			m, err := record.MergedEvents(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			anchorsOn := map[string][]string{} // bundle -> migrated computation anchors, in order
			used := map[string]int{}           // bundle -> anchors used by the current reproduce's run
			var got []carried
			for _, e := range m.Events {
				if p, ok := recordpb.BodyAs[*recordpb.Computation](e); ok {
					anchorsOn[p.GetBundleSha()] = append(anchorsOn[p.GetBundleSha()], p.GetComputationId())
				}
				r, ok := recordpb.BodyAs[*recordpb.Reproduce](e)
				if !ok {
					continue
				}
				got = append(got, carried{r.GetBundleSha(), recordpb.Word(r.GetRerun()), r.GetRan(), recordpb.Word(r.GetSoundness())})
				// A REPRODUCE THAT NAMED ITS ANCHOR KEEPS IT. A lens re-runs whichever computation on
				// a shared bundle it is judging, so its anchor is not the bundle's next in order.
				if j := len(got) - 1; j < len(named) && named[j] != "" {
					if r.GetComputationId() != named[j] {
						t.Errorf("reproduce %d names %q; its source named %q", m.At.Of(e).ID, r.GetComputationId(), named[j])
					}
					kept++
					continue
				}
				// The anchors of one carried run are the computations on the bundle, in the order written.
				if r.GetRan() {
					used[r.GetBundleSha()] = 0
				}
				on := anchorsOn[r.GetBundleSha()]
				if k := used[r.GetBundleSha()]; k >= len(on) || on[k] != r.GetComputationId() {
					t.Errorf("reproduce %d names %q; the computations on its bundle so far are %q", m.At.Of(e).ID, r.GetComputationId(), on)
				}
				used[r.GetBundleSha()]++
			}
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("the migrated reproduces are not the source's carried to their anchors:\n got  %d %v\n want %d %v", len(got), got, len(want), want)
			}
			fanned += len(want)
		})
	}
	// THE TRANSLATION IS EXERCISED: archives hold computations that share a bundle, so some
	// reproduce is carried to more than one anchor, and archives hold `not_run` reproduces. Without
	// either, this test reads the same either way.
	if shared == 0 || fanned == 0 {
		t.Errorf("no archived reproduce was carried to a second anchor (%d carried, %d shared) — the test checks nothing", fanned, shared)
	}
	if stood == 0 {
		t.Error("no archived reproduce ran nothing — the check that it states its standing re-run checks nothing")
	}
	if kept == 0 {
		t.Error("no archived reproduce named its anchor as written — the check that it keeps it checks nothing")
	}
}

// A RAISED CONTRADICTION IS RAISED ON THE MIGRATED RECORD. A finding names a contradiction by
// the events.id of the reading that opened it, and every archive renumbers: the cast is written
// first, an anchor lands after the act that placed it. The reference is translated to the id the
// reading was written as; untranslated, it names another event or none.
//
// No archive holds such a finding — the kind arrives at epoch 28 — so one is seeded into the
// quadratic-formula archive, right after its first refuting reading, by that reading's seat.
func TestARaisedContradictionIsRaisedOnTheMigratedRecord(t *testing.T) {
	from := recordtest.ExtractArchive(t, "2026-09-02_quadratic-formula.tar.gz")
	src := archivedSource(t, from)
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	var seeded []OldEvent
	var reading OldEvent
	for _, old := range evs {
		seeded = append(seeded, old)
		if o, _ := old.Fields["outcome"].(string); old.Word == "verify" && reading.ID == 0 && (o == "refutes" || o == "absent") {
			reading = old
			seeded = append(seeded, OldEvent{ID: 1 << 40, SeatID: old.SeatID, Round: old.Round, TS: old.TS, Key: old.SeatID + ":finding:contradiction",
				Word: "finding", Fields: map[string]any{"id": "F-0000c0de", "text": "the source contradicts the sentence",
					"severity": "medium", "likelihood": "medium", "impact": "medium",
					"about_kind": "contradiction", "about_ref": strconv.FormatInt(old.ID, 10)}})
		}
	}
	if reading.ID == 0 {
		t.Fatal("the archive holds no refuting reading")
	}
	// The run's channels, as Migrate stages them before a replay.
	to := recordtest.TmpRun(t)
	if err := copyChannels(from, to); err != nil {
		t.Fatal(err)
	}
	dst := runtest.New(t, to)
	res, err := replayInOrder(sliceOf(seeded), Entries(), dst, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	b, err := record.EvidenceJSONBytes(dst)
	if err != nil {
		t.Fatal(err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal(b, &ev); err != nil {
		t.Fatal(err)
	}
	claim, _ := reading.Fields["claim"].(string)
	m, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	var ref string
	for _, e := range m.Events {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok && f.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_CONTRADICTION {
			ref = f.GetAboutRef()
		}
	}
	// THE ID SHIFTED, or this test is vacuous: an untranslated reference would pass by luck.
	if ref == strconv.FormatInt(reading.ID, 10) || ref == "" {
		t.Fatalf("the migrated finding names contradiction %q, the reading's source id %d — the archive did not renumber, so nothing was tested", ref, reading.ID)
	}
	// The finding raises it and no gap answers it, so it is listed — raised by the seeded finding.
	// A reference left untranslated would raise another reading's contradiction, or none.
	raised := false
	for _, c := range ev.UnansweredContradictions {
		if c.Claim == claim {
			raised = c.RaisedBy == "F-0000c0de"
			if !raised {
				t.Errorf("the raised claim %q is listed raised by %q on the migrated record, want F-0000c0de", claim, c.RaisedBy)
			}
		}
	}
	if !raised {
		t.Errorf("the raised claim %q is not listed raised by F-0000c0de on the migrated record", claim)
	}
	if len(ev.UnansweredContradictions) == 0 {
		t.Errorf("no unanswered contradiction remains: the archive's other readings must still be listed, or the read is blank")
	}
}

// sliceOf is a source over events already read.
type sliceOf []OldEvent

func (s sliceOf) Events() ([]OldEvent, error)   { return s, nil }
func (s sliceOf) Files() []migration.SourceFile { return nil }
func (s sliceOf) Close() error                  { return nil }

// A REFERENCE MIGRATE CANNOT TRANSLATE IS A REFUSAL BY NAME, and a carried reproduce with no
// computation of its bundle before it has nowhere to go.
func TestAnUntranslatableReferenceIsRefusedByName(t *testing.T) {
	if _, err := contradictionRef("17", map[int64]int64{3: 4}); err == nil {
		t.Error("a contradiction reference to a reading the replay wrote nothing for was translated")
	}
	if got, err := contradictionRef("3", map[int64]int64{3: 4}); err != nil || got != "4" {
		t.Errorf("contradictionRef(3) = %q, %v, want 4", got, err)
	}
	if _, err := reproduceAnchors([]proto.Message{&recordpb.Reproduce{BundleSha: proto.String("s1")}}, map[string][]string{}); err == nil {
		t.Error("a reproduce of a bundle no computation was written on was carried to an anchor")
	}
}
