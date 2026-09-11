package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// corrected is an act, its replacement and the correction between them, keyed as the record keys
// them — the three events one same-sitting correction leaves.
func corrected(t *testing.T, seat, key string, was, is proto.Message) []*record.Event {
	t.Helper()
	return []*record.Event{
		recordtest.At(t, seat, key, was),
		recordtest.At(t, seat, key+"~1", is),
		recordtest.At(t, seat, seat+":correction:"+key, &recordpb.Correction{
			Corrects: proto.String(key), Replacement: proto.String(key + "~1"), Why: proto.String("a word was lost")}),
	}
}

// THE REPORT SHOWS A CORRECTED ACT STRUCK, BESIDE THE ACT THAT REPLACED IT — never hidden, and
// never counted as a second act. B3's manifest rows are the case that made this necessary.
func TestTheReportShowsACorrectedActStruck(t *testing.T) {
	evs := blueSatOn(t, []string{"G1"}, "G1")
	evs = append(evs, corrected(t, "blue-respond", "blue-respond:manifest_row:#1:G1",
		&recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("G1 is reproducible via ")},
		&recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("G1 is reproducible via the recorded proof")})...)
	evs = append(evs, corrected(t, "blue-respond", "blue-respond:position:#1",
		&recordpb.Position{Text: proto.String("the report is  now")},
		&recordpb.Position{Text: proto.String("the report is sound now")})...)
	evs = append(evs, corrected(t, "blue-respond", "blue-respond:log:#1",
		&recordpb.Log{Text: proto.String("the tool  refused"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()},
		&recordpb.Log{Text: proto.String("the tool refused the cite"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()})...)
	evs = append(evs, corrected(t, "judge", "judge:outcome:#1",
		&recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("it stopped because  refused")},
		&recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("it stopped because the gate refused")})...)
	fam := (&boardT{GapOrder: []string{"G1"},
		Gaps:   map[string]*record.Gap{"G1": {ID: "G1", HasClosed: true, Closure: &recordpb.Close{}}},
		Events: evs}).fam()

	struck := func(text string) string { return "~~" + text + "~~ (struck by blue-respond: a word was lost)" }
	sections := map[string]string{
		"correctness manifest": correctnessManifest(fam),
		"debate":               debate(fam, evs),
		"log":                  logSection(evs),
	}
	for name, want := range map[string][]string{
		"correctness manifest": {"### Blue's correctness manifest (1)", struck("G1 is reproducible via"), "G1 is reproducible via the recorded proof"},
		"debate":               {struck("the report is  now"), "the report is sound now", "(struck by judge: a word was lost)"},
		"log":                  {struck("the tool  refused"), "the tool refused the cite"},
	} {
		for _, w := range want {
			if !strings.Contains(sections[name], w) {
				t.Errorf("the %s does not show %q:\n%s", name, w, sections[name])
			}
		}
	}
	if got := outcomeOf(evs).GetProse(); got != "it stopped because the gate refused" {
		t.Errorf("the terminal outcome read = %q, want the one that stands", got)
	}
}

// JUDGMENTS SHOW A CORRECTED RULING STRUCK, then the ruling that stands — the answer is the
// corrected one, and the struck wording is not hidden behind it.
func TestJudgmentsShowACorrectedRulingStruck(t *testing.T) {
	ruling := func(opinion string) *recordpb.MotionRule {
		return &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String(opinion),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CARRIED), Principle: proto.String("p"),
				Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s"), ReopensOn: proto.String("r")}}}
	}
	evs := []*record.Event{
		recordtest.At(t, "red-chair", "red-chair:motion:#1", &recordpb.Motion{MotionId: proto.String("M1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET), Basis: proto.String("red cannot settle G1"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}),
	}
	evs = append(evs, corrected(t, "judge", "judge:motion_rule:#1", ruling("because  refuses"), ruling("because the gate refuses"))...)
	got := motions((&boardT{Events: evs}).fam())
	struck := strings.Index(got, "~~ruled by judge — because  refuses~~ (struck by judge: a word was lost)")
	stands := strings.Index(got, "because the gate refuses")
	if struck < 0 || stands < 0 || stands < struck {
		t.Errorf("judgments must show the struck ruling, marked, and then the ruling that stands:\n%s", got)
	}
}

// A CORRECTION TAKES ITS TARGET'S PLACE (F12). The judge records outcome #1, then #2, then corrects
// #1: the terminal outcome is still #2. Read in raw order, #1's replacement is the last outcome on
// the record and would be taken for the run's ending.
func TestACorrectedEarlierOutcomeDoesNotBecomeTheTerminalOne(t *testing.T) {
	out := func(prose string) *recordpb.Outcome {
		return &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String(prose)}
	}
	evs := []*record.Event{
		recordtest.At(t, "judge", "judge:outcome:#1", out("first account, with a word  lost")),
		recordtest.At(t, "judge", "judge:outcome:#2", out("the account that ends the run")),
		recordtest.At(t, "judge", "judge:outcome:#1~1", out("first account, whole")),
		recordtest.At(t, "judge", "judge:correction:judge:outcome:#1", &recordpb.Correction{
			Corrects: proto.String("judge:outcome:#1"), Replacement: proto.String("judge:outcome:#1~1"), Why: proto.String("a word was lost")}),
	}
	if got := outcomeOf(evs).GetProse(); got != "the account that ends the run" {
		t.Errorf("the terminal outcome read = %q, want #2 — a correction of #1 stands in #1's place", got)
	}
}
