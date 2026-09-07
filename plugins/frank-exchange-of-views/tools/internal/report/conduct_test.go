package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// registers builds a board of register events, one per (seat, served, requested) triple. An empty
// requested means the harness declared no swap; an empty served means NOTHING MEASURED.
func registers(t *testing.T, rows ...[3]string) *record.Board {
	t.Helper()
	b := &record.Board{}
	for _, r := range rows {
		reg := &recordpb.Register{}
		if r[1] != "" {
			reg.ServedModel = proto.String(r[1])
		}
		if r[2] != "" {
			reg.RequestedModel = proto.String(r[2])
		}
		b.Events = append(b.Events, recordtest.Event(t, r[0], 1, reg))
	}
	return b
}

// THE SUBSTITUTION HAS TO REACH THE READER, because the reader is the only party who can act on
// it and the report is the only thing they see.
//
// MEASURED: run B's certified report said "the pairing this run actually used — blue on `fable`,
// red on `sonnet`" and then reasoned for two paragraphs about same-vendor bias from that premise,
// while every one of its 44 bulk seats had been answered by claude-opus-4-8. The claim was not
// careless — configuration is the only model fact a seat can see.
func TestASubstitutedTierIsNamedInTheReport(t *testing.T) {
	got := conduct(registers(t,
		[3]string{"blue-lane-1", "claude-opus-4-8", "claude-fable-5"},
		[3]string{"red-lens-r1-evidence", "claude-opus-4-8", "claude-fable-5"},
		[3]string{"red-merge-r1", "claude-sonnet-5", ""},
	))
	for _, want := range []string{"claude-opus-4-8", "claude-fable-5", "SUBSTITUTED", "claude-sonnet-5"} {
		if !strings.Contains(got, want) {
			t.Errorf("the conduct section never names %q:\n%s", want, got)
		}
	}
	// AND IT SAYS WHAT THE SUBSTITUTION COSTS THE DOCUMENT. Naming the model without saying that
	// the report's own reasoning about models was written against the other one leaves the reader
	// to notice the contradiction themselves, which is what nobody did for $379.
	if !strings.Contains(got, "does not describe") {
		t.Errorf("a substituted run does not tell the reader that the document's model reasoning is stale:\n%s", got)
	}
}

// THE KEY MISMATCH THIS TEST WAS WRITTEN FOR. The first cut kept requested counts in ONE map
// keyed by requested name and then asked it `requested[served]` — a lookup of a served name among
// requested names. It answered 0 for every real substitution, so the section rendered a clean run
// on precisely the input it exists to convict, and every assertion about the healthy case passed.
func TestTheSubstitutionIsFoundBySERVEDModelNotByCoincidence(t *testing.T) {
	got := conduct(registers(t, [3]string{"blue-lane-1", "served-x", "asked-y"}))
	if !strings.Contains(got, "SUBSTITUTED") {
		t.Fatalf("a seat whose served and requested models share no substring was reported as un-substituted — "+
			"the join is on the wrong key:\n%s", got)
	}
	if !strings.Contains(got, "configured as `asked-y`") {
		t.Errorf("the section names the served model but not what was ASKED for:\n%s", got)
	}
}

// NOT MEASURED IS NOT MATCHED. The absent case and the healthy case are the same bytes unless one
// of them says so — and a report that silently omits the seats nobody looked at reads as a run
// whose every seat was checked.
func TestSeatsWithNoMeasurementAreReportedAsNotMeasured(t *testing.T) {
	got := conduct(registers(t,
		[3]string{"blue-lane-1", "claude-opus-4-8", ""},
		[3]string{"blue-lane-2", "", ""},
	))
	if !strings.Contains(got, "NOT MEASURED") {
		t.Errorf("a seat whose serving model nobody observed vanished from the table:\n%s", got)
	}
	if strings.Contains(got, "SUBSTITUTED") {
		t.Errorf("an unmeasured seat was reported as a substitution — the three states must stay apart:\n%s", got)
	}
}

// A CLEAN RUN STILL SAYS WHAT ANSWERED, and says it without the alarm. The section is provenance
// first and a warning second: a reader who cannot see the models on a healthy run has no baseline
// against which the substituted one means anything.
func TestAnUnsubstitutedRunStillNamesWhatAnswered(t *testing.T) {
	got := conduct(registers(t,
		[3]string{"blue-lane-1", "claude-fable-5", ""},
		[3]string{"red-merge-r1", "claude-sonnet-5", ""},
	))
	if !strings.Contains(got, "claude-fable-5") || !strings.Contains(got, "claude-sonnet-5") {
		t.Errorf("a clean run does not name its own models:\n%s", got)
	}
	for _, unwanted := range []string{"SUBSTITUTED", "NOT MEASURED"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("a clean run raised %q:\n%s", unwanted, got)
		}
	}
}

// A RUN WITH NO REGISTERS RENDERS NOTHING, rather than an empty table. Assembly runs over
// fixtures and partial runs; a heading with no rows under it is a claim that the question was
// asked and came back blank, which is not what happened.
func TestNoRegistersRendersNoSection(t *testing.T) {
	if got := conduct(&record.Board{}); got != "" {
		t.Errorf("a board with no register events rendered a section:\n%s", got)
	}
}
