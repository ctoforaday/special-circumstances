package migrate

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A SOURCE BEFORE EPOCH 32 CARRIES COLUMNS NO READER READ, and each translates to the same body
// without them: every other column comes forward as it was. Stated as data, one row per word.
func TestAColumnNoReaderReadIsNotCarried(t *testing.T) {
	for _, tc := range []struct {
		word   string
		fields map[string]any
		want   proto.Message
	}{
		{"verify", map[string]any{"claim": "c", "url": "https://ex/a", "title": "A", "independent": int64(1),
			"outcome": "supports", "confidence": "high", "text": "read it"},
			&recordpb.Verify{Claim: proto.String("c"), Url: proto.String("https://ex/a"), Title: proto.String("A"),
				Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
				Text: proto.String("read it")}},
		{"cite", map[string]any{"label": "C-00000001", "url": "https://ex/a", "source_text_read": "leaf",
			"source_text_origin": "embedded", "work_status": "standing", "source_completeness": "full"},
			&recordpb.Cite{Label: proto.String("C-00000001"), Url: proto.String("https://ex/a"),
				SourceTextOrigin:   recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(),
				WorkStatus:         recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(),
				SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum()}},
		{"log", map[string]any{"text": "estoppel — G-00000001", "type": "estoppel", "source": "tool", "estopped_by": "G-00000001"},
			&recordpb.Log{Text: proto.String("estoppel — G-00000001"), Type: recordpb.LogType_LOG_TYPE_ESTOPPEL.Enum(),
				Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}},
		{"sitting_open", map[string]any{"agent_id": "a1", "agent_type": "frank-exchange-of-views:red-chair",
			"seat_id": "red-chair", "session_id": "s1", "prompt_id": "p1"},
			&recordpb.SittingOpen{AgentId: proto.String("a1"), AgentType: proto.String("frank-exchange-of-views:red-chair"),
				SeatId: proto.String("red-chair")}},
		{"sitting_close", map[string]any{"agent_id": "a1", "agent_type": "frank-exchange-of-views:red-chair",
			"agent_transcript_path": "/t/agent-a1.jsonl", "session_id": "s1", "prompt_id": "p1"},
			&recordpb.SittingClose{AgentId: proto.String("a1"), AgentTranscriptPath: proto.String("/t/agent-a1.jsonl")}},
		{"sitting_limit", map[string]any{"agent_id": "a1", "agent_type": "frank-exchange-of-views:red-chair",
			"seat_id": "red-chair", "sitting": int64(2), "limit": int64(150)},
			&recordpb.SittingLimit{SeatId: proto.String("red-chair"), Sitting: proto.Int32(2), Limit: proto.Int32(150)}},
	} {
		t.Run(tc.word, func(t *testing.T) {
			bodies, err := Entries().Translate(OldEvent{ID: 5, SeatID: "s", Word: tc.word, Fields: tc.fields})
			if err != nil {
				t.Fatal(err)
			}
			if len(bodies) != 1 || !proto.Equal(bodies[0], tc.want) {
				t.Errorf("translated to %v, want %v", bodies, tc.want)
			}
		})
	}
}

// EVERY ARCHIVED VERIFICATION MIGRATES, AND EVERY ONE THAT SAID `independent` IS A CORROBORATION. The
// bool is not carried because the empty anchor already says it: the census reads every archive and
// fails on a source row where the two disagree — an independent verification naming an anchor, or an
// anchorless one not marked independent — because then the bool carried a fact and dropping it would
// lose it. It also fails on a migrated record holding a different number of verifications, cites,
// sitting ends or limits than its source, or a different number of corroborations than the source's
// independent rows, and on an archive set holding no independent row at all, which would leave the
// translation tested by nothing.
func TestEveryArchivedIndependentVerificationMigratesAsACorroboration(t *testing.T) {
	independent := 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]int{}
			wantIndependent := 0
			for _, old := range evs {
				switch old.Word {
				case "verify", "cite", "sitting_open", "sitting_close", "sitting_limit":
					want[old.Word]++
				default:
					continue
				}
				if old.Word != "verify" {
					continue
				}
				ind := false
				switch v := old.Fields["independent"].(type) {
				case int64:
					ind = v == 1
				case bool:
					ind = v
				}
				anchor, _ := old.Fields["anchor"].(string)
				if ind != (anchor == "") {
					t.Errorf("source verify %d: independent=%v with anchor %q — the bool carries a fact the anchor does not", old.ID, ind, anchor)
				}
				if ind {
					wantIndependent++
				}
			}
			_, dir := MigratedArchive(t, name)
			m, err := record.MergedEvents(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]int{}
			gotCorroborations := 0
			for _, e := range m.Events {
				switch {
				case e.GetVerify() != nil:
					got["verify"]++
					if e.GetVerify().GetAnchor() == "" {
						gotCorroborations++
					}
				case e.GetCite() != nil:
					got["cite"]++
				case e.GetSittingOpen() != nil:
					got["sitting_open"]++
				case e.GetSittingClose() != nil:
					got["sitting_close"]++
				case e.GetSittingLimit() != nil:
					got["sitting_limit"]++
				}
			}
			for w, n := range want {
				if got[w] != n {
					t.Errorf("migrated %d %s event(s), the source holds %d", got[w], w, n)
				}
			}
			if gotCorroborations != wantIndependent {
				t.Errorf("migrated %d corroboration(s), the source marks %d verification(s) independent", gotCorroborations, wantIndependent)
			}
			independent += wantIndependent
		})
	}
	if independent == 0 {
		t.Fatal("no archived run holds an independent verification, so the translation was tested by nothing")
	}
	t.Logf("%d archived independent verification(s) migrated as corroborations", independent)
}
