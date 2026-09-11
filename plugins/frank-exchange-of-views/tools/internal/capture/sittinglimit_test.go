package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func limitAuditRun(t *testing.T) record.Run {
	t.Helper()
	dir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recordsql.CloseUnder(dir) })
	r, err := record.NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	// one ordinary event so the record exists
	if _, err := record.Append(record.Identity{Run: r, SeatID: record.HarnessSeat},
		&recordpb.SittingOpen{AgentId: proto.String("a1"), AgentType: proto.String("t")}); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAStoppedSittingIsNamedByCapture(t *testing.T) {
	r := limitAuditRun(t)
	if a := SittingLimitAudit(r); a.Verdict != "PASS" {
		t.Errorf("a run with no stopped sitting: %s %s, want PASS", a.Verdict, a.Detail)
	}
	if _, err := record.Append(record.Identity{Run: r, SeatID: record.HarnessSeat}, &recordpb.SittingLimit{
		AgentId: proto.String("a1"), SeatId: proto.String("red-lens-evidence-L1"),
		Sitting: proto.Int32(3), Limit: proto.Int32(150),
	}); err != nil {
		t.Fatal(err)
	}
	a := SittingLimitAudit(r)
	if a.Verdict != "WARN" || !strings.Contains(a.Detail, "red-lens-evidence-L1 sitting 3 at 150 calls") {
		t.Errorf("a stopped sitting: %s %q, want WARN naming red-lens-evidence-L1 sitting 3 at 150 calls", a.Verdict, a.Detail)
	}
}
