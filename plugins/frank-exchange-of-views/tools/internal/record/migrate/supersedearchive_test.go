package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A GAP A SUCCESSOR SUPERSEDED WHILE IT WAS OPEN IS CLOSED BY THAT MINT (ruled by gblock 2026-11-04).
// universe-m30's voice lens minted G-968365ad superseding its open G-97b44713, then regraded the
// ancestor and closed it repaired; its logic lens minted G-5d729d0a superseding its open G-ab17309d,
// then closed the ancestor repaired. Migrated, each ancestor reads closed by its successor's mint,
// with the fate "superseded" and no closing body, and the later regrade and closes, acts on a gap
// already closed, are left out.
func TestAGapSupersededWhileOpenIsClosedByItsSuccessorsMint(t *testing.T) {
	res, dst := migrateArchive(t, "2026-10-19_is-91-prime-m30.tar.gz")
	if got := res.In["close"] - res.Out["close"]; got != 2 {
		t.Errorf("migrate left out %d close(s), want the 2 on the superseded ancestors", got)
	}
	if got := res.In["regrade"] - res.Out["regrade"]; got != 1 {
		t.Errorf("migrate left out %d regrade(s), want G-97b44713's", got)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	successors := map[string]string{"G-97b44713": "G-968365ad", "G-ab17309d": "G-5d729d0a"}
	for anc, succ := range successors {
		g := fam.Gap(anc)
		if g == nil {
			t.Fatalf("%s is not in the migrated family", anc)
		}
		if g.Open || g.SupersededBy != succ || g.Closure != nil {
			t.Errorf("%s: open=%v superseded_by=%q close=%v, want closed by %s's mint with no closing body", anc, g.Open, g.SupersededBy, g.Closure != nil, succ)
		}
		if got := g.ClosureReason(); got != record.ClosedSuperseded {
			t.Errorf("%s: fate %q, want %q", anc, got, record.ClosedSuperseded)
		}
		for i, e := range fam.Events {
			if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok && c.GetGapId() == anc {
				t.Errorf("%s: the migrated record holds a close of the superseded gap at %d", anc, i+1)
			}
			if r, ok := recordpb.BodyAs[*recordpb.Regrade](e); ok && r.GetGapId() == anc {
				t.Errorf("%s: the migrated record holds a regrade of the superseded gap at %d", anc, i+1)
			}
		}
	}
	bj, err := record.BoardJSONOfRun(dst)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, g := range bj.Closed {
		if succ, ok := successors[g.ID]; ok {
			found++
			if g.SupersededBy != succ {
				t.Errorf("board: %s reads superseded_by %q, want %s", g.ID, g.SupersededBy, succ)
			}
		}
	}
	if found != len(successors) {
		t.Errorf("board: %d of the superseded ancestors read closed, want %d; anomalies %v", found, len(successors), bj.Anomalies)
	}
}
