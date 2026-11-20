package capture

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A FAILED SITTING LISTS EVERY (TOOL, DEFECT) ROW IN IT AS ITS TOOL DEFECTS, AND CALLS NONE OF THEM ITS
// REASON: a `reproduce` defect fails no sitting, and a delivery that could not advance is one defect
// among them. The join is the row's own source, type and sitting — a (TOOL, FAILURE) row in the same
// sitting is no defect, and a defect in the seat's next sitting is not this one's.
func TestFailedSittingsAuditListsEachSittingsToolDefects(t *testing.T) {
	const (
		lens       = "red-lens-computation"
		reproduce  = "copies left behind in the bundle's run directory: out/"
		stuck      = "the lens's next item could not be handed: the record could not be read"
		failedCall = "feov-record lens reproduce: exit status 1"
		later      = "a defect of the lens's next sitting"
	)
	open := func(agent string) *recordpb.SittingOpen {
		return &recordpb.SittingOpen{AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:" + lens), SeatId: proto.String(lens)}
	}
	log := func(text string, typ recordpb.LogType) *recordpb.Event {
		return recordtest.At(t, lens, "", &recordpb.Log{Text: proto.String(text), Type: typ.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()})
	}
	dir := t.TempDir()
	recordtest.Seed(t, dir,
		recordtest.At(t, record.HarnessSeat, "", &recordpb.Cast{SeatIds: []string{"red-chair", lens}}),
		recordtest.At(t, record.HarnessSeat, "", open("L1")),
		log(reproduce, recordpb.LogType_LOG_TYPE_DEFECT),
		log(failedCall, recordpb.LogType_LOG_TYPE_FAILURE),
		log(stuck, recordpb.LogType_LOG_TYPE_DEFECT),
		recordtest.At(t, record.HarnessSeat, "", &recordpb.SittingClose{AgentId: proto.String("L1")}),
		recordtest.At(t, record.HarnessSeat, "", open("L2")),
		log(later, recordpb.LogType_LOG_TYPE_DEFECT),
		recordtest.At(t, record.HarnessSeat, "", &recordpb.EnvelopeAdmitted{AgentId: proto.String("L2")}),
		recordtest.At(t, record.HarnessSeat, "", &recordpb.SittingClose{AgentId: proto.String("L2")}),
	)
	a := FailedSittingsAudit(runtest.Open(t, dir))
	if a.Verdict != "WARN" || !strings.Contains(a.Detail, lens+" sitting 1") || strings.Contains(a.Detail, lens+" sitting 2") {
		t.Fatalf("want the lens's first sitting alone named failed, audited %s:\n%s", a.Verdict, a.Detail)
	}
	for _, want := range []string{reproduce, stuck} {
		if !strings.Contains(a.Detail, want) {
			t.Errorf("the failed sitting's tool defect %q is not listed:\n%s", want, a.Detail)
		}
	}
	for _, not := range []string{failedCall, later} {
		if strings.Contains(a.Detail, not) {
			t.Errorf("%q is listed as a tool defect of the failed sitting:\n%s", not, a.Detail)
		}
	}
	if !strings.Contains(a.Detail, "tool defects (2)") || strings.Contains(strings.ToLower(a.Detail), "reason") {
		t.Errorf("the defects are not listed as the sitting's tool defects, or one is called its reason:\n%s", a.Detail)
	}
}

// A FAILED SITTING WITH NO TOOL DEFECT SAYS SO, rather than listing nothing.
func TestFailedSittingsAuditSaysWhereASittingHasNoToolDefect(t *testing.T) {
	a := FailedSittingsAudit(stoppedRun(t, time.Date(2026, 10, 27, 17, 0, 0, 0, time.UTC), false))
	if !strings.Contains(a.Detail, "red-chair sitting 1") || !strings.Contains(a.Detail, "no tool defect") {
		t.Errorf("a failed sitting with no tool defect does not say so:\n%s", a.Detail)
	}
}
