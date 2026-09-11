package consistency

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// corrected is an act, its replacement and the correction between them, keyed as the record keys
// them — what one same-sitting correction leaves on the record.
func corrected(t *testing.T, seat, key string, was, is proto.Message) []*recordpb.Event {
	t.Helper()
	return corrChain(t, seat, key, was, is)
}

// corrChain is a chain: the act, then each replacement in turn, each struck by the next.
func corrChain(t *testing.T, seat, key string, versions ...proto.Message) []*recordpb.Event {
	t.Helper()
	evs := []*recordpb.Event{recordtest.At(t, seat, key, versions[0])}
	prev := key
	for i, v := range versions[1:] {
		next := key + "~" + string(rune('1'+i))
		evs = append(evs,
			recordtest.At(t, seat, next, v),
			recordtest.At(t, seat, seat+":correction:"+prev, &recordpb.Correction{
				Corrects: proto.String(prev), Replacement: proto.String(next), Why: proto.String("a word was lost")}))
		prev = next
	}
	return evs
}

// A RECORD HOLDING CORRECTIONS IS READ THE SAME WAY BY EVERY READER: the ground truth folds the
// acts that stand, and the SQL views and the Go overlay agree about what was struck and where each
// act stands. A regrade corrected to a different severity is the sharpest case — the walk, the
// gap view's current grade and the board must all read the replacement's.
func TestARecordWithCorrectionsIsConsistent(t *testing.T) {
	lens := "red-lens-evidence"
	regrade := func(sev recordpb.Grade, basis string) *recordpb.Regrade {
		return &recordpb.Regrade{GapId: proto.String("G1"), Severity: sev.Enum(), Basis: proto.String(basis)}
	}
	closure := func(prose string) *recordpb.Close {
		return &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./..."), Prose: proto.String(prose)}
	}
	pos := func(s string) *recordpb.Position { return &recordpb.Position{Text: proto.String(s)} }
	row := func(s string) *recordpb.ManifestRow {
		return &recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String(s)}
	}
	evs := []*recordpb.Event{mint(t, lens, "G1")}
	evs = append(evs, corrected(t, lens, lens+":regrade:#0:G1",
		regrade(recordpb.Grade_GRADE_LOW, "the consequence is  bounded"), regrade(recordpb.Grade_GRADE_HIGH, "the consequence reaches every caller"))...)
	evs = append(evs, corrected(t, lens, lens+":close:G1", closure("verified at the  leaf"), closure("verified at the leaf"))...)
	evs = append(evs, corrected(t, "blue-respond", "blue-respond:manifest_row:#0:G1", row("checked the  thing"), row("checked the recorded thing"))...)
	evs = append(evs, corrChain(t, "blue-respond", "blue-respond:position:#0",
		pos("the report is  now"), pos("the report is sound  now"), pos("the report is sound now"))...)

	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir, evs...)
	v, err := Check(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(v) > 0 {
		t.Errorf("%d violation(s) on a record whose corrections every reader should agree about:\n  %s", len(v), strings.Join(v, "\n  "))
	}
}
