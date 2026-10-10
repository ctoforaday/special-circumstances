package report

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// bracket is the hook's SubagentStart: it opens seat's sitting under agent, whether or not the seat
// ever registers.
func bracket(t *testing.T, seat, agent string) *record.Event {
	t.Helper()
	return recordtest.At(t, record.HarnessSeat, "harness:sitting_open:"+agent, &recordpb.SittingOpen{
		AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:" + seat), SeatId: proto.String(seat)})
}

// registersUnder is seat's register under agent, keyed n: it joins agent's bracket when the hook
// bracketed it.
func registersUnder(t *testing.T, seat, agent, n string) *record.Event {
	t.Helper()
	return recordtest.At(t, seat, seat+":register:"+n, &recordpb.Register{AgentId: proto.String(agent)})
}

// THE M15 SHAPE, AND EVERY SURFACE THAT PRINTS AN EPOCH OR A SITTING NUMBER (#1206).
//
// The write path stores which sitting each act is in (events.sitting_id): a hook bracket opens one,
// a register under the agent it bracketed joins it, and a repair joins the sitting it repairs.
// Counting registers over the events misses the first, counts the second twice and gives the third
// a sitting of its own — on universe-m15, 77 of 317 rows, and judgments.md printed `filed by
// red-lens-voice #2` for a motion the record holds in that seat's FIFTH sitting.
//
// The fixture is that run's shapes in miniature:
//   - red-lens-voice: a bracket-only sitting, a bracket with a register that joins it and a second
//     register in the same sitting, then a bracket-only sitting in which it files a docket motion,
//     a finding and a log — its THIRD stored sitting, and its second by any count of registers;
//   - red-chair: a bracketed sitting its register joins (epoch 1), then a bracket-only sitting
//     (epoch 2), so every act after it is in epoch 2 and a count of chair registers says 1;
//   - blue-respond: a sitting, then a sitting-record repair whose acts are the repaired sitting's.
//
// Each expectation below is the stored number, written out, so a surface that recounts fails here by
// name rather than agreeing with an index it was handed.
func TestEveryPrintedEpochAndSittingIsTheStoredOne(t *testing.T) {
	blueOpened := "blue-respond:register:#1"
	evs := []*record.Event{
		bracket(t, "red-lens-voice", "V1"), // voice sitting 1, epoch 0: bracket only
		bracket(t, "red-chair", "C1"),      // epoch 1
		registersUnder(t, "red-chair", "C1", "#1"),
		mintT(t, "red-lens-logic", "G1"),
		bracket(t, "red-lens-voice", "V2"), // voice sitting 2
		registersUnder(t, "red-lens-voice", "V2", "#1"),
		registersUnder(t, "red-lens-voice", "V2", "#2"), // a second register in the same sitting
		bracket(t, "red-chair", "C2"),                   // epoch 2: the chair never registers in it
		bracket(t, "red-lens-voice", "V3"),              // voice sitting 3: bracket only
		recordtest.At(t, "red-lens-voice", "red-lens-voice:motion:M1", &recordpb.Motion{
			MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:  proto.String("red cannot settle G1"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}),
		recordtest.At(t, "red-lens-voice", "red-lens-voice:finding:F1", &recordpb.Finding{
			Id: proto.String("F-00000001"), Text: proto.String("the voice slips")}),
		recordtest.At(t, "red-lens-voice", "red-lens-voice:log:#1", &recordpb.Log{Text: proto.String("the tool refused"),
			Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		registersUnder(t, "blue-respond", "B1", "#1"), // blue sitting 1, epoch 2
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, "blue-respond", "blue-respond:position:#1", &recordpb.Position{Text: proto.String("G1 is answered")}),
		recordtest.At(t, "blue-respond", "blue-respond:avenue:#1", &recordpb.Avenue{AvenueId: proto.String("Q1"),
			Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: proto.String("survey the literature")}),
		recordtest.At(t, record.HarnessSeat, "harness:sitting_close:B1", &recordpb.SittingClose{AgentId: proto.String("B1")}),
		// The repair: a sitting of the seat's own, whose acts are the sitting it repairs.
		recordtest.At(t, "blue-respond", "blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("B2"), RepairsSitting: proto.String(blueOpened)}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e2", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, "blue-respond", "blue-respond:manifest_row:m1", &recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("recomputed the figure")}),
		recordtest.At(t, "blue-respond", "blue-respond:retire:r1", &recordpb.Retire{Claim: proto.String("the old figure"),
			Reason: proto.String("superseded"), RemovalBasis: proto.String(record.RemovalAsserted)}),
		recordtest.At(t, "blue-respond", "blue-respond:revision:#1", &recordpb.Revision{Text: proto.String("recomputed the figure")}),
	}
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	run := runtest.Open(t, dir)
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}

	contains := func(surface, got, want string) {
		t.Helper()
		if !strings.Contains(got, want) {
			t.Errorf("%s does not print the stored number %q:\n%s", surface, want, got)
		}
	}
	absent := func(surface, got, wrong string) {
		t.Helper()
		if strings.Contains(got, wrong) {
			t.Errorf("%s prints %q, a number counted from registers rather than the stored one:\n%s", surface, wrong, got)
		}
	}

	t.Run("judgments.md", func(t *testing.T) {
		got := motions(fam)
		contains("judgments.md", got, "filed by red-lens-voice #3")
	})

	t.Run("the motions view", func(t *testing.T) {
		var mj record.MotionsJSON
		unmarshal(t, must(t)(record.MotionsJSONBytes(run)), &mj)
		if len(mj.Motions) != 1 || mj.Motions[0].Epoch != 2 {
			t.Errorf("show motions = %+v, want M1 at epoch 2", mj.Motions)
		}
	})

	t.Run("the findings view", func(t *testing.T) {
		var fj record.FindingsJSON
		unmarshal(t, must(t)(record.FindingsJSONBytes(run)), &fj)
		if len(fj.Findings) != 1 || fj.Findings[0].Epoch != 2 {
			t.Errorf("show findings = %+v, want F-00000001 at epoch 2", fj.Findings)
		}
		if got := record.FindingsJSONOf(fam.Events, fam.At).Findings; len(got) != 1 || got[0].Epoch != 2 {
			t.Errorf("the findings fold over the family = %+v, want F-00000001 at epoch 2", got)
		}
	})

	t.Run("the log view", func(t *testing.T) {
		var lj record.LogJSON
		unmarshal(t, must(t)(record.LogJSONBytes(run)), &lj)
		if len(lj.Log) != 1 || lj.Log[0].Epoch != 2 {
			t.Errorf("ops log = %+v, want the voice's entry at epoch 2", lj.Log)
		}
		// red-chair, blue-respond: sat and filed nothing. The chair's second sitting has no register,
		// and its first register joined the bracket: the record holds two chair sittings either way.
		if lj.Counts.Clean != 2 {
			t.Errorf("clean = %d, want 2 (red-chair and blue-respond sat and filed nothing)", lj.Counts.Clean)
		}
	})

	t.Run("the debate", func(t *testing.T) {
		var dj record.DebateJSON
		unmarshal(t, must(t)(record.DebateJSONBytes(run)), &dj)
		var at []int
		for _, ep := range dj.Epochs {
			if len(ep.Blue) > 0 {
				at = append(at, ep.Epoch)
			}
		}
		if len(at) != 1 || at[0] != 2 {
			t.Errorf("show debate puts blue's position in epoch(s) %v, want [2]", at)
		}
		d := debate(fam)
		contains("debate.md", d, "### Epoch 2")
		absent("debate.md", d, "### Epoch 1\n")
	})

	t.Run("the avenues view", func(t *testing.T) {
		got := record.AvenuesJSONOf(fam.Events, fam.At).Avenues
		if len(got) != 1 || got[0].Epoch != 2 {
			t.Errorf("show avenues = %+v, want Q1 at epoch 2", got)
		}
	})

	t.Run("show changes", func(t *testing.T) {
		got := string(must(t)(view.Markdown(run, "changes", "")))
		contains("show changes", got, "`blue-respond` #1")
		absent("show changes", got, "`blue-respond` #2")
	})

	t.Run("the changelog", func(t *testing.T) {
		contains("withdrawn claims", withdrawnClaims(fam), "(blue-respond #1)")
		manifest := correctnessManifest(fam)
		contains("correctness manifest", manifest, "**G1** (blue-respond #1)")
		absent("correctness manifest", manifest, "#2")
	})

	// THE CHANGELOG HOLDS NO REVISION PROSE. The record above holds a revision — a migrated archive
	// holds them, and nothing else writes one — and the document a human is handed lists what left
	// the report and nothing a seat narrated about its sitting.
	t.Run("the changelog document", func(t *testing.T) {
		docs, err := AssembleAll(run)
		if err != nil {
			t.Fatal(err)
		}
		var body, blurb string
		for _, d := range docs {
			if d.File == FileChangelog {
				body, blurb = d.Body, d.Blurb
			}
		}
		contains("CHANGELOG.md", body, "the old figure")
		for _, gone := range []string{"Report revision history", "recomputed the figure", "### Epoch"} {
			if strings.Contains(body, gone) {
				t.Errorf("CHANGELOG.md prints %q — a revision's prose:\n%s", gone, body)
			}
		}
		if strings.Contains(strings.ToLower(blurb), "revision") {
			t.Errorf("the changelog's blurb promises revisions: %q", blurb)
		}
	})

	t.Run("the fact box", func(t *testing.T) {
		contains("fact box", factBox(fam), "**Epochs** | 2")
	})

	t.Run("cost's seat bindings", func(t *testing.T) {
		b := cost.SeatBindingsOf(fam.Events, fam.At)
		for agent, want := range map[string]cost.SeatBinding{
			"C1": {SeatID: "red-chair", Epoch: 1, Sitting: 1},
			"V2": {SeatID: "red-lens-voice", Epoch: 1, Sitting: 2},
			// The bracket-only sittings bind by their bracket: no register names these agents.
			"V1": {SeatID: "red-lens-voice", Epoch: 0, Sitting: 1},
			"C2": {SeatID: "red-chair", Epoch: 2, Sitting: 2},
			"V3": {SeatID: "red-lens-voice", Epoch: 2, Sitting: 3},
			"B1": {SeatID: "blue-respond", Epoch: 2, Sitting: 1},
			// The repair agent is labelled with the sitting it completes.
			"B2": {SeatID: "blue-respond", Epoch: 2, Sitting: 1},
		} {
			if got := b[agent]; got != want {
				t.Errorf("%s binds to %+v, want %+v", agent, got, want)
			}
		}
	})

	t.Run("the current epoch", func(t *testing.T) {
		if got := fam.At.CurrentEpoch(fam.Events); got != 2 {
			t.Errorf("current epoch = %d, want 2", got)
		}
	})
}

func must(t *testing.T) func([]byte, error) []byte {
	return func(b []byte, err error) []byte {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
}

func unmarshal(t *testing.T, b []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%v:\n%s", err, b)
	}
}
