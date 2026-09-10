package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// WHAT ANSWERED THE SEATS IS ENVELOPE, SO IT IS IN run.md AND NOT IN report.md.
//
// gblock, 2026-09-10: "It's envelope. Move out of report.md." The section sat in report.md "in the same
// breath" as the verdict, and its tool-authored heading was the one voice tell no run could remove
// from the research document (#792, §V item 3). The difftest goldens render report.md only, so after
// the move they showed the section LEAVING and nothing showing it ARRIVE — a section composed
// nowhere would have passed every one of them. This asserts both halves on an assembled run.
//
// ONE SEAT IS REGISTERED ON PURPOSE. conduct() renders nothing for a run with no seats, so a seatless
// fixture would pass this test whichever document the section was composed into.
func TestWhatAnsweredTheSeatsIsInRunMdNotReportMd(t *testing.T) {
	runDir := newRun(t)
	blue := "# Is 91 prime? — research report\n\n## TL;DR\n91 = 7 x 13, so it is composite.\n"
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(blue), 0o644); err != nil {
		t.Fatal(err)
	}
	id := record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, ""); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(blue)}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if _, err := Assemble(runtest.Open(t, runDir)); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(runDir, name))
		if err != nil {
			t.Fatalf("assemble did not write %s: %v", name, err)
		}
		return string(b)
	}
	report, runmd := read(FileReport), read(FileRun)

	const heading = "## How this run was conducted"
	if strings.Contains(report, heading) {
		t.Errorf("report.md still carries the conduct section — it is envelope, not research:\n%s", report)
	}
	if !strings.Contains(runmd, heading) {
		t.Errorf("run.md does not carry the conduct section, so the served-model fact is composed nowhere:\n%s", runmd)
	}
	// The TABLE, not only the heading: a heading with no body would satisfy the line above.
	if !strings.Contains(runmd, "what answered") {
		t.Errorf("run.md's conduct section has no measured-model table:\n%s", runmd)
	}
}

// NO VERDICT'S GLOSS IS IN report.md; EVERY VERDICT'S IS IN run.md.
//
// gblock, 2026-09-10, three rulings that add up to one rule: a halted run's gloss "Move it out of
// report.md"; CEILING's "Move it out like HALTED"; the basis "State on the stamp, prose to run.md".
// So report.md carries the stamp — the word, and the basis as state — and nothing that explains
// either. Each row names a phrase ONLY the gloss prose contains, and asserts it on both sides: gone
// from report.md, present in run.md's verdict-basis section. A gloss composed nowhere fails the
// second half; a gloss left in place fails the first.
func TestNoVerdictsGlossIsInReportMdAndEveryOnesIsInRunMd(t *testing.T) {
	for _, tc := range []struct {
		name    string
		outcome *recordpb.Outcome
		stamp   string // what report.md must say
		prose   string // a phrase only the gloss says — must be absent from report.md, present in run.md
	}{
		{"halted", &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED),
			Prose: proto.String("the bench halted the run on a safety petition")},
			"**Verdict:** HALTED", "the bench ended this run"},
		{"ceiling, derived", &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING),
			VerdictBasis: proto.String(record.VerdictDerived), Prose: proto.String("every gap reached its limit")},
			"**Verdict:** CEILING-TERMINATED (derived from the record)", "never audited by a red pass"},
		{"unverified, asserted", &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED),
			VerdictBasis: proto.String(record.VerdictAsserted), Prose: proto.String("the run ended before a terminal state")},
			"**Verdict:** UNVERIFIED (asserted by the bench)", "The record holds no terminal state"},
		// NO OUTCOME AT ALL (gblock: "State only"). The stamp says NONE; what went missing — the bench
		// never ran `bench outcome` — is a fact about the run, and run.md says it.
		{"no outcome", nil, "**Verdict:** NONE (no terminal outcome on the record)", "was not run before assembly"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report, runmd := assembleWithOutcome(t, tc.outcome)
			if !strings.Contains(report, tc.stamp) {
				t.Errorf("report.md does not carry the stamp %q — the STATE stays, only the prose moves:\n%s", tc.stamp, report)
			}
			if strings.Contains(report, tc.prose) {
				t.Errorf("report.md still carries the gloss (%q):\n%s", tc.prose, report)
			}
			if !strings.Contains(runmd, "## The verdict's basis") || !strings.Contains(runmd, tc.prose) {
				t.Errorf("run.md does not carry the gloss (%q), so it is composed nowhere:\n%s", tc.prose, runmd)
			}
		})
	}
}

// assembleWithOutcome assembles a one-seat run with the given terminal outcome and returns
// report.md and run.md.
func assembleWithOutcome(t *testing.T, outcome *recordpb.Outcome) (report, runmd string) {
	t.Helper()
	runDir := newRun(t)
	blue := "# Is 91 prime? — research report\n\n## TL;DR\n91 = 7 x 13, so it is composite.\n"
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(blue), 0o644); err != nil {
		t.Fatal(err)
	}
	add := func(seat string, body proto.Message) {
		t.Helper()
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seat}
		if _, _, err := record.RegisterSeat(id, ""); err != nil {
			t.Fatalf("register %s: %v", seat, err)
		}
		if _, err := record.Append(id, body); err != nil {
			t.Fatalf("append %s/%T: %v", seat, body, err)
		}
	}
	add("blue-synthesize", &recordpb.BaseIngest{Text: proto.String(blue)})
	if outcome != nil {
		add("judge-terminal", outcome)
	}
	if _, err := Assemble(runtest.Open(t, runDir)); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(runDir, name))
		if err != nil {
			t.Fatalf("assemble did not write %s: %v", name, err)
		}
		return string(b)
	}
	return read(FileReport), read(FileRun)
}
