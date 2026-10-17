package record

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// quadraticArchive is the quadratic-formula run. The dark-side lens read C-4e4c92af's source
// against its sentence and found it absent (event 570); blue's edit at 599 rewrote the sentence
// before any finding quoted it.
const quadraticArchive = "2026-09-02_quadratic-formula.tar.gz"

// A CONTRADICTION WHOSE SENTENCE BLUE EDITED IS ANSWERED BY THE CITATION'S STANDING (#1378). On the
// archive as it ended, contradiction 570 holds the PASS: the sentence it was of is gone, and no
// check has read the one that replaced it. A backing check of C-4e4c92af from a sitting opened
// after the edit answers it, and so does a retire of the anchor.
func TestQuadraticFormulasEditedContradictionIsAnsweredByItsCitation(t *testing.T) {
	whole := archivedEvents(t, quadraticArchive)
	const id, anchor = "570", "C-4e4c92af"
	read := whole[569]
	v, ok := read.GetBody().(*recordpb.Event_Verify)
	if !ok || v.Verify.GetAnchor() != anchor || !contradicts(v.Verify.GetOutcome()) || whole[598].GetType() != recordpb.EventType_EVENT_TYPE_BLUE_EDIT {
		t.Fatalf("the archive is not the quadratic-formula run this test reads: event 570 is a %s, 599 a %s", read.GetType(), whole[598].GetType())
	}
	seed := func(extra ...proto.Message) Run {
		dir := recordtest.RequireArchive(t, quadraticArchive)
		for _, f := range []string{"record.db", "record.db-wal", "record.db-shm"} {
			if err := os.Remove(filepath.Join(dir, "records", f)); err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
		evs := append([]*Event{}, whole...)
		for i, body := range extra {
			evs = append(evs, recordtest.At(t, evLens, fmt.Sprintf("%s:after-%d", evLens, i), body))
		}
		recordtest.Seed(t, dir, evs...)
		return mustRun(t, dir)
	}
	held := func(run Run) bool {
		for _, b := range mustBlockers(t, run) {
			if b.Kind == BlockerContradiction && b.Subject == id {
				return true
			}
		}
		return false
	}
	if !held(seed()) {
		t.Fatalf("contradiction %s does not hold the PASS on the archive as it ended", id)
	}
	backs := &recordpb.Verify{Anchor: proto.String(anchor), Claim: proto.String("the sentence as edit 599 left it"),
		Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("read at the leaf")}
	if held(seed(&recordpb.Register{}, backs)) {
		t.Errorf("contradiction %s still holds the PASS after a backing check of %s from a sitting opened after the edit", id, anchor)
	}
	retire := &recordpb.Retire{Claim: proto.String("the sentence"), Reason: proto.String("the source does not hold it"), Anchors: []string{anchor}}
	if held(seed(retire)) {
		t.Errorf("contradiction %s still holds the PASS after the retire of %s", id, anchor)
	}
}
