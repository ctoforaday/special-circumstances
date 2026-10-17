package record

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// m19Archive is universe-m19, the run this rule was written from: 492 events, outcome VERIFIED,
// with two citations and a proof whose sentences blue rewrote after their last check.
const m19Archive = "2026-10-10_is-91-prime-m19.tar.gz"

// m19Reopened is what m19 owes at its chair's verdict (event 486), as events.id.
//
// THE IDS ARE LITERALS, DERIVED BY SQL OUTSIDE THE FOLD — the last `blue_edit_reopened` row naming
// each anchor, and the last verify or reproduce of it before that — so a wrong fold cannot agree
// with them by construction. They are the row ids of the archive as migrated, which replays its
// events in order; a migration that renumbers them re-derives the three pairs from the migrated
// record and updates this table in the same commit. The anchors do not move.
var m19Reopened = []reopened{
	{anchor: "C-706fda63", edit: 394, lastCheck: 168},
	{anchor: "C-4dae2507", edit: 363, lastCheck: 325},
	{anchor: "P-cd9f4c5a", sha: "10ef5e650b99c5143a5741c500fe37554e061653830365b71518d4f8fcc1dab7", edit: 363, lastCheck: 324},
}

// m19At is m19 as it stood after event n: the archive's first n events replayed, in order, into a
// run of its own beside the archive's inputs and class registry. The replay stores each act's
// sitting by the write path's own rule, so every reader below reads a record, not a slice.
func m19At(t *testing.T, whole []*Event, n int) Run {
	t.Helper()
	return archiveAt(t, m19Archive, whole, n)
}

// archiveMigrate brings an archived run forward into a fresh run directory; archivemigrate_test.go
// installs migrate here.
var archiveMigrate func(src, dst string) error

// archivedEvents is the named archive's whole stream, migrated to this epoch and read off the
// migrated record's event tables on a plain read-only handle. Each test replays a prefix of them
// into a record of this binary's making (archiveAt), so no view of the archive's is ever asked.
func archivedEvents(t *testing.T, archive string) []*Event {
	t.Helper()
	if archiveMigrate == nil {
		t.Fatal("no migration installed: archivemigrate_test.go installs migrate before any test runs")
	}
	migrated := filepath.Join(t.TempDir(), "run")
	if err := archiveMigrate(recordtest.RequireArchive(t, archive), migrated); err != nil {
		t.Fatalf("migrating %s: %v", archive, err)
	}
	if err := recordsql.CloseUnder(migrated); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(migrated, "records", "record.db")
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	evs, _, err := eventsAt(db)
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

// archiveAt is the named archive as it stood after event n, replayed as m19At says.
func archiveAt(t *testing.T, archive string, whole []*Event, n int) Run {
	t.Helper()
	dir := recordtest.RequireArchive(t, archive)
	for _, f := range []string{"record.db", "record.db-wal", "record.db-shm"} {
		if err := os.Remove(filepath.Join(dir, "records", f)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	recordtest.Seed(t, dir, whole[:n]...)
	return mustRun(t, dir)
}

// M19 PASSED OVER THREE CHECKS OF SENTENCES IT NO LONGER CARRIED. Its evidence lens found against
// C-706fda63 (event 111); blue rewrote the sentence to answer it (edit 394); no seat read the new
// sentence against the source, and the chair's PASS at 486 was admitted. Under this rule the plans
// that followed the edits ready the evidence lens, and the PASS is refused naming all three.
func TestM19ReopenedEvidenceHoldsThePass(t *testing.T) {
	whole := archivedEvents(t, m19Archive)
	if len(whole) != 492 || whole[485].GetType() != recordpb.EventType_EVENT_TYPE_VERDICT {
		t.Fatalf("the archive holds %d events and its event 486 is a %s, want 492 and the chair's verdict", len(whole), whole[485].GetType())
	}
	fact := func(r reopened) string {
		return fmt.Sprintf("%s's sentence changed at edit %d and no sitting opened since has checked it (its last check is event %d)", r.anchor, r.edit, r.lastCheck)
	}

	// The record before the chair's verdict.
	run := m19At(t, whole, 485)
	blockers, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var got []reopened
	for _, b := range blockers {
		if b.Kind != BlockerReopened || b.Owner != evLens {
			t.Errorf("%s (%q, owned by %q) holds the PASS as well, want only the reopened anchors, owned by %s", b.Kind, b.Subject, b.Owner, evLens)
			continue
		}
		got = append(got, reopened{anchor: b.Subject, sha: b.Detail, edit: b.since, lastCheck: b.lastCheck})
	}
	if !slices.Equal(got, m19Reopened) {
		t.Fatalf("reopened blockers = %+v\nwant %+v", got, m19Reopened)
	}
	refusal := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if refusal == nil {
		t.Fatal("the gate admits the PASS m19's chair recorded at event 486")
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Errorf("pass_permitted is true before the verdict: %q", plan.Why)
	}
	chair := strings.Join(blockingItems(sittingOfRunT(t, run, "chair", "red-chair")), "\n")
	for _, r := range m19Reopened {
		if !strings.Contains(refusal.Error(), fact(r)) || !strings.Contains(refusal.Error(), "clears it, whichever lens records it. "+reopenedOwed) {
			t.Errorf("the PASS refusal does not state %q:\n%v", fact(r), refusal)
		}
		if !strings.Contains(chair, fact(r)+" — PASS is refused until a lens checks it from a sitting opened after the edit, by re") || !strings.Contains(chair, reopenedOwed) {
			t.Errorf("the chair's work list does not hold %q:\n%s", fact(r), chair)
		}
		if !slices.Contains(plan.Why, fact(r)+" — PASS is not permitted until a lens checks it from a sitting opened after the edit. "+reopenedOwed) {
			t.Errorf("the plan's reasons do not state %q: %q", fact(r), plan.Why)
		}
		at, err := EvidenceAt(run, r.anchor)
		if err != nil || at.Reopened == nil || *at.Reopened != (EvidenceReopenedJSON{Anchor: r.anchor, Edit: r.edit, LastCheck: r.lastCheck}) {
			t.Errorf("show evidence --anchor %s: reopened = %+v (%v), want edit %d, last check %d", r.anchor, at.Reopened, err, r.edit, r.lastCheck)
		}
	}
	if !strings.Contains(chair, "by re-running it (`reproduce --id "+m19Reopened[2].sha+"`)") {
		t.Errorf("the proof's item does not carry the sha its re-run takes:\n%s", chair)
	}
	var rows []reopened
	for _, r := range standingOf(t, run) {
		rows = append(rows, reopened{anchor: r.Anchor, edit: r.Edit, lastCheck: r.LastCheck})
	}
	if want := []reopened{{anchor: "C-706fda63", edit: 394, lastCheck: 168}, {anchor: "C-4dae2507", edit: 363, lastCheck: 325}, {anchor: "P-cd9f4c5a", edit: 363, lastCheck: 324}}; !slices.Equal(rows, want) {
		t.Errorf("show evidence reopened = %+v\nwant %+v", rows, want)
	}
	// A proof reopened and retired in one blue sitting (edits 427 and 430, retire 431) owes nothing.
	if at, err := EvidenceAt(run, "P-c1c0989f"); err != nil || at.Proof == nil || at.Reopened != nil {
		t.Errorf("show evidence --anchor P-c1c0989f: reopened = %+v (%v), want the proof with no reopened row", at.Reopened, err)
	}

	// The two plans that followed the edits: the chair's dispatches at events 373 and 402 readied
	// architecture, voice and blue, then architecture alone. Each now readies the evidence lens.
	for _, c := range []struct {
		before int
		edit   int64
		owed   []string
	}{
		{373, 363, []string{"C-706fda63", "C-4dae2507", "P-cd9f4c5a"}},
		{402, 394, []string{"C-706fda63", "C-4dae2507", "P-cd9f4c5a"}},
	} {
		plan, err := PlanDispatch(m19At(t, whole, c.before-1))
		if err != nil {
			t.Fatal(err)
		}
		if _, ready := partyOf(plan, evLens); !ready || plan.PassPermitted {
			t.Errorf("the plan at event %d: parties %+v, pass_permitted %v — want %s readied and no PASS", c.before, plan.Parties, plan.PassPermitted, evLens)
		}
		if want := fmt.Sprintf("%s has not sat since edit %d,", evLens, c.edit); !strings.Contains(whyOf(plan), want) {
			t.Errorf("the plan at event %d does not say %q: %q", c.before, want, plan.Why)
		}
		var owed []string
		for _, b := range plan.Blockers {
			if b.Kind == BlockerReopened {
				owed = append(owed, b.Subject)
			}
		}
		if !slices.Equal(owed, c.owed) {
			t.Errorf("the plan at event %d holds reopened %q, want %q", c.before, owed, c.owed)
		}
	}
}
