package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A DISPATCH'S RETURN WRITES WHAT THE HOOKS WRITE, through the writer they spawn: the envelope's
// admission, then the sitting's close, both under the agent its opening carried. Pinned here because
// the sweep reaches it by a draw and the deterministic tests in this package draw it on no seed.
func TestADispatchReturnWritesTheHooksAdmissionAndClose(t *testing.T) {
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"),
		[]byte("# § fuzz\n\nA § fuzz sentence to anchor findings.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	cast, lanes := record.CastFor(1)
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, newLockedRand(7))
	// A forced run closes every sitting, so the never-closed draw cannot take this one.
	r.forceVerified = true
	agent := r.registerTwice("red-chair", "")
	if agent == "" {
		t.Fatalf("the agent-bearing register was refused: %s", r.actErr)
	}
	r.sittingAgent = map[string]string{"red-chair": agent}
	r.dispatchReturns("red-chair", true)
	if r.actErr != "" {
		t.Fatalf("the return's writes: %s", r.actErr)
	}
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, e := range fam.Events {
		if b, ok := recordpb.BodyAs[*recordpb.EnvelopeAdmitted](e); ok && b.GetAgentId() == agent {
			order = append(order, "admitted")
		}
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && b.GetAgentId() == agent {
			order = append(order, "closed")
		}
	}
	if len(order) != 2 || order[0] != "admitted" || order[1] != "closed" {
		t.Fatalf("the record holds %v for %s, want its admission and then its close", order, agent)
	}
	failed, err := record.FailedSittings(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if len(failed) != 0 {
		t.Errorf("a sitting closed with its envelope admitted reads as failed: %+v", failed)
	}
}
