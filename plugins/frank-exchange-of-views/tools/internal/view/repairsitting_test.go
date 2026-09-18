package view

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN EDIT FILED IN A REPAIR IS THE REPAIRED SITTING'S (#1002, gblock 2026-09-18). Both forms of
// this view attribute an edit to a sitting, so both follow the sitting the act belongs to rather
// than the seat's turn count: blue's revision, filed by the re-prompt that exists to put it on the
// record, is grouped under the sitting that owed it — one heading, not two.
func TestARepairsEditIsGroupedUnderTheSittingItCompletes(t *testing.T) {
	runDir := t.TempDir()
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, "red-chair", "red-chair:mint:G1", &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			GapId:           proto.String("G1"),
			Class:           proto.String("overclaim"),
			Problem:         proto.String("independence is overclaimed"),
			RequiredFix:     proto.String("acknowledge the shared definition"),
			AcceptanceCheck: proto.String("the section no longer claims independence"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
	})
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, "blue-respond", "blue-respond:register:1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e1", &recordpb.BlueEdit{
			Answers: proto.String("G1"),
			Old:     proto.String("five independent approaches"),
			New:     proto.String("five approaches"),
			Text:    proto.String("drop the independence claim"),
		}),
	})
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, record.HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
	})
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, "blue-respond", "blue-respond:register:2", &recordpb.Register{AgentId: proto.String("blue-b"), RepairsSitting: proto.String("blue-respond:register:1")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e2", &recordpb.BlueEdit{
			Answers: proto.String("G1"),
			Old:     proto.String("They agree."),
			New:     proto.String("They agree, sharing one definition of primality."),
			Text:    proto.String("name the shared definition"),
		}),
	})

	out := md(t, runDir, "changes")
	if strings.Count(out, "## `blue-respond` #1") != 1 || strings.Contains(out, "## `blue-respond` #2") {
		t.Errorf("the repair's edit opened a heading of its own instead of joining the sitting it completes:\n%s", out)
	}

	b, err := Markdown(runtest.Open(t, runDir), "changes", "G1")
	if err != nil {
		t.Fatalf("scoped changes: %v", err)
	}
	if scoped := string(b); !strings.Contains(scoped, "### 2. `blue-respond` #1 ·") {
		t.Errorf("the scoped comparison numbers the repair's edit under another sitting:\n%s", scoped)
	}
}
