package dashboard

import (
	"io"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// A BENCH SITTING ITS BRACKET OPENED IS A JUDGE SITTING, register or none (#1262). The bracket is
// the SubagentStart writer's own write.
func TestABenchSittingItsBracketOpenedIsAJudgeSitting(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, registered(t, "red-chair", "C1"))
	if err := sittingwrite.Write(sittingwrite.Sitting{RunDir: dir, Phase: sittingwrite.Open,
		AgentID: "J1", AgentType: "frank-exchange-of-views:lead-judge"}, io.Discard); err != nil {
		t.Fatalf("the hook's bracket for the bench: %v", err)
	}
	if n := BuildModel(runtest.Open(t, dir), t.TempDir(), Config{}, 0).Judiciary.JudgeSittings; n != 1 {
		t.Errorf("JudgeSittings = %d, want 1: the bench's bracket opened a sitting it never registered in", n)
	}
}
