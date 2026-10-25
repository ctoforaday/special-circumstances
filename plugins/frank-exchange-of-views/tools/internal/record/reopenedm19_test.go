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

// m19Archive is universe-m19, the run this rule was written from: 496 events migrated, outcome VERIFIED,
// with two citations and a proof whose sentences blue rewrote after their last check.
const m19Archive = "2026-10-10_is-91-prime-m19.tar.gz"

// m19Reopened is what m19 owes at its chair's verdict (event 490), as events.id.
//
// THE IDS ARE LITERALS, DERIVED BY SQL OUTSIDE THE FOLD — the last `blue_edit_reopened` row naming
// each anchor, and the last verify or reproduce of it before that — so a wrong fold cannot agree
// with them by construction. They are the row ids of the archive as migrated, which replays its
// events in order; a migration that renumbers them re-derives the three pairs from the migrated
// record and updates this table in the same commit. The anchors do not move.
var m19Reopened = []reopened{
	{anchor: "C-706fda63", edit: 400, lastCheck: 175},
	{anchor: "C-4dae2507", edit: 370, lastCheck: 332},
	{anchor: "P-cd9f4c5a", proof: true, sha: "10ef5e650b99c5143a5741c500fe37554e061653830365b71518d4f8fcc1dab7", edit: 370, lastCheck: 331},
}

// m19Owner is who owes each reopened anchor's check in m19's cast, which seats every lens: the
// evidence lens a citation's, the computation lens the proof's (reopenedOwnerOf).
func m19Owner(r reopened) string {
	if r.proof {
		return computationSeat
	}
	return evLens
}

// m19At is m19 as it stood after event n: the archive's first n events replayed, in order, into a
// run of its own beside the archive's inputs. The replay stores each act's
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
// C-706fda63 (event 112); blue rewrote the sentence to answer it (edit 400); no seat read the new
// sentence against the source, and the chair's PASS at 490 was admitted. Under this rule the plans
// that followed the edits ready each anchor's owner — the evidence lens for the citations, the
// computation lens for the proof — and the PASS is refused naming all three.
func TestM19ReopenedEvidenceHoldsThePass(t *testing.T) {
	whole := archivedEvents(t, m19Archive)
	if len(whole) != 496 || whole[489].GetType() != recordpb.EventType_EVENT_TYPE_VERDICT {
		t.Fatalf("the archive holds %d events and its event 490 is a %s, want 496 and the chair's verdict", len(whole), whole[489].GetType())
	}
	fact := func(r reopened) string {
		return fmt.Sprintf("%s's sentence changed at edit %d and no sitting opened since has checked it (its last check is event %d)", r.anchor, r.edit, r.lastCheck)
	}

	// The record before the chair's verdict.
	run := m19At(t, whole, 489)
	blockers, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	// AND EACH OWNER IS READY: it owes the check, so it is not retired, and its pin is behind the head
	// (retirement.go). THE ARCHIVE PREDATES THE REPORT READ, so no lens of it was ever served the
	// report on the record and every unretired lens is behind the head: an owner's reason names its
	// reopened check, any other's that it was never served the report.
	owners := map[string]bool{computationSeat: true, evLens: true}
	var got []reopened
	var ready []string
	unraised := false
	for _, b := range blockers {
		if b.Kind == BlockerLensReady {
			switch {
			case !strings.Contains(b.foldWhy, "no sitting it sat through was served the whole report"):
				t.Errorf("%s is ready for %q, want it behind the head: the archive records no report read", b.Subject, b.foldWhy)
			case owners[b.Subject] && (!strings.Contains(b.foldWhy, "it does not retire while it owes") || !strings.Contains(b.foldWhy, "reopened ")):
				t.Errorf("%s is ready for %q, want it unretired because it owes a reopened check", b.Subject, b.foldWhy)
			}
			ready = append(ready, b.Subject)
			continue
		}
		// AND A CONTRADICTION NO FINDING NAMED: the reading at 112. The archive's finding about it
		// quotes the claim and names no contradiction, and a quote raises nothing.
		if b.Kind == BlockerContradiction && b.Subject == "112" && b.Owner == evLens && b.raisedBy == "" {
			unraised = true
			continue
		}
		if b.Kind != BlockerReopened {
			t.Errorf("%s (%q, owned by %q) holds the PASS as well, want only the reopened anchors, their owners' readiness and contradiction 112", b.Kind, b.Subject, b.Owner)
			continue
		}
		r := reopened{anchor: b.Subject, proof: b.Detail != "", sha: b.Detail, edit: b.since, lastCheck: b.lastCheck}
		if b.Owner != m19Owner(r) {
			t.Errorf("reopened %s is owned by %q, want %s", b.Subject, b.Owner, m19Owner(r))
		}
		got = append(got, r)
	}
	if !unraised {
		t.Error("contradiction 112, which no finding named, does not hold the PASS unraised")
	}
	if !slices.Equal(got, m19Reopened) {
		t.Fatalf("reopened blockers = %+v\nwant %+v", got, m19Reopened)
	}
	if slices.Sort(ready); !slices.Contains(ready, computationSeat) || !slices.Contains(ready, evLens) {
		t.Errorf("ready lenses = %q, want the two owners among them", ready)
	}
	refusal := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if refusal == nil {
		t.Fatal("the gate admits the PASS m19's chair recorded at event 490")
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
		owed := reopenedOwed(Blocker{Owner: m19Owner(r)})
		if !strings.Contains(refusal.Error(), fact(r)+" — "+m19Owner(r)+" owes the check and is the seat readied for it") || !strings.Contains(refusal.Error(), "clears it, whichever lens records it; ") {
			t.Errorf("the PASS refusal does not state %q:\n%v", fact(r), refusal)
		}
		if !strings.Contains(chair, fact(r)+" — PASS is refused until a lens checks it from a sitting opened after the edit, by ") || !strings.Contains(chair, owed) {
			t.Errorf("the chair's work list does not hold %q:\n%s", fact(r), chair)
		}
		if !slices.Contains(plan.Why, fact(r)+" — PASS is not permitted until a lens checks it from a sitting opened after the edit. "+owed) {
			t.Errorf("the plan's reasons do not state %q: %q", fact(r), plan.Why)
		}
		at, err := EvidenceAt(run, r.anchor)
		if err != nil || at.Reopened == nil || *at.Reopened != (EvidenceReopenedJSON{Anchor: r.anchor, Edit: r.edit, LastCheck: r.lastCheck}) {
			t.Errorf("show evidence --anchor %s: reopened = %+v (%v), want edit %d, last check %d", r.anchor, at.Reopened, err, r.edit, r.lastCheck)
		}
	}
	if !strings.Contains(chair, "(`reproduce --anchor "+m19Reopened[2].anchor+"`, which re-runs nothing once the script has a re-run)") {
		t.Errorf("the proof's item does not carry the anchor its judgement takes:\n%s", chair)
	}
	var rows []reopened
	for _, r := range standingOf(t, run) {
		rows = append(rows, reopened{anchor: r.Anchor, edit: r.Edit, lastCheck: r.LastCheck})
	}
	if want := []reopened{{anchor: "C-706fda63", edit: 400, lastCheck: 175}, {anchor: "C-4dae2507", edit: 370, lastCheck: 332}, {anchor: "P-cd9f4c5a", edit: 370, lastCheck: 331}}; !slices.Equal(rows, want) {
		t.Errorf("show evidence reopened = %+v\nwant %+v", rows, want)
	}
	// A proof reopened and retired in one blue sitting (edits 432 and 435, retire 436) owes nothing.
	if at, err := EvidenceAt(run, "P-c1c0989f"); err != nil || at.Proof == nil || at.Reopened != nil {
		t.Errorf("show evidence --anchor P-c1c0989f: reopened = %+v (%v), want the proof with no reopened row", at.Reopened, err)
	}

	// The two plans that followed the edits: the chair's dispatches at events 379 and 407 readied
	// architecture, voice and blue, then architecture alone. Each now readies the evidence lens for
	// the citations and the computation lens for the proof.
	for _, c := range []struct {
		before int
		owed   []string
	}{
		{379, []string{"C-706fda63", "C-4dae2507", "P-cd9f4c5a"}},
		{407, []string{"C-706fda63", "C-4dae2507", "P-cd9f4c5a"}},
	} {
		plan, err := PlanDispatch(m19At(t, whole, c.before-1))
		if err != nil {
			t.Fatal(err)
		}
		if _, ready := partyOf(plan, evLens); !ready || plan.PassPermitted {
			t.Errorf("the plan at event %d: parties %+v, pass_permitted %v — want %s readied and no PASS", c.before, plan.Parties, plan.PassPermitted, evLens)
		}
		if _, ready := partyOf(plan, computationSeat); !ready {
			t.Errorf("the plan at event %d: parties %+v — want %s readied for the proof it owns", c.before, plan.Parties, computationSeat)
		}
		// Each owner is readied because it owes the anchors' checks: its reason names every one it owns.
		for _, r := range m19Reopened {
			if !slices.ContainsFunc(plan.Why, func(w string) bool {
				return strings.HasPrefix(w, m19Owner(r)+": ") && strings.Contains(w, ", behind head ") && strings.Contains(w, "it does not retire while it owes ") && strings.Contains(w, "reopened "+r.anchor)
			}) {
				t.Errorf("the plan at event %d does not ready %s for %s: %q", c.before, m19Owner(r), r.anchor, plan.Why)
			}
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
