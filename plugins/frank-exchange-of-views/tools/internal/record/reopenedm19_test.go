package record

import (
	"database/sql"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// m19Archive is universe-m19, the run this rule was written from: 526 events migrated, outcome VERIFIED,
// with two citations and a computation whose sentences blue rewrote after their last check.
const m19Archive = "2026-10-10_is-91-prime-m19.tar.gz"

// m19Reopened is what m19 owes at its chair's verdict (event 504), as events.id.
//
// THE IDS ARE LITERALS, DERIVED BY SQL OUTSIDE THE FOLD — the last `blue_edit_reopened` row naming
// each anchor, and the last verify or reproduce of it before that — so a wrong fold cannot agree
// with them by construction. They are the row ids of the archive as migrated, which replays its
// events in order; a migration that renumbers them re-derives the three pairs from the migrated
// record and updates this table in the same commit. The anchors do not move.
var m19Reopened = []unchecked{
	{anchor: "C-706fda63", edit: 415, lastCheck: 180},
	{anchor: "C-4dae2507", edit: 389, lastCheck: 348},
	{anchor: "K-cd9f4c5a", computation: true, sha: "078d75b87b5cf2c5319d398a8d1c736ff5ecfbc1780abed8a448b0057ede9a70", edit: 389, lastCheck: 347},
}

// owesClause is whether a lens's reason holds the fold's clause on an unretired lens that owes
// (lensFold.retiresWhy): past its barren limit it does not retire while it owes, and below it, it
// retires at the limit only owing nothing. A sitting that checks is productive (plan constitution-gaps
// §III.3, "Productive sitting (retirement)"), so m19's owners, whose sittings hold their checks, are
// below the limit and carry the second.
func owesClause(why string) bool {
	return strings.Contains(why, "it does not retire while it owes ") || strings.Contains(why, fmt.Sprintf("it retires at %d owing nothing, and it owes ", retiresAfter))
}

// m19Owner is who owes each reopened anchor's check: the evidence lens a citation's, the computation
// lens the computation's (anchorOwnerOf).
func m19Owner(r unchecked) string {
	if r.computation {
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

// archiveMigrate brings an archived run forward into a fresh run directory, and returns the acts its
// manifest lists in lost_acts; archivemigrate_test.go installs migrate here.
var archiveMigrate func(src, dst string) ([]LostAct, error)

// LostAct is one archived act the migration wrote as no event (the manifest's lost_acts): the
// write refuses it, for want of a delivery or for a gap such a loss took with it. Gap and Motion
// are the migrated ids it names, empty where it names none.
type LostAct struct {
	OldID          int64
	Word, TS, Seat string
	Gap, Motion    string
}

// archivedEvents is the named archive's whole stream, migrated to this epoch and read off the
// migrated record's event tables on a plain read-only handle. Each test replays a prefix of them
// into a record of this binary's making (archiveAt), so no view of the archive's is ever asked.
func archivedEvents(t *testing.T, archive string) []*Event {
	t.Helper()
	evs, _ := archivedRun(t, archive)
	return evs
}

// archivedRun is archivedEvents with the acts the migration lost.
func archivedRun(t *testing.T, archive string) ([]*Event, []LostAct) {
	t.Helper()
	if archiveMigrate == nil {
		t.Fatal("no migration installed: archivemigrate_test.go installs migrate before any test runs")
	}
	migrated := filepath.Join(t.TempDir(), "run")
	lost, err := archiveMigrate(recordtest.RequireArchive(t, archive), migrated)
	if err != nil {
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
	return evs, lost
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
// C-706fda63 (event 112); blue rewrote the sentence to answer it (edit 415); no seat read the new
// sentence against the source, and the chair's PASS at 504 was admitted. Under this rule the plans
// that followed the edits ready each anchor's owner — the evidence lens for the citations, the
// computation lens for the computation — and the PASS is refused naming all three.
func TestM19ReopenedEvidenceHoldsThePass(t *testing.T) {
	whole, lost := archivedRun(t, m19Archive)
	if len(whole) != 512 || whole[503].GetType() != recordpb.EventType_EVENT_TYPE_VERDICT {
		t.Fatalf("the archive holds %d events and its event 504 is a %s, want 512 and the chair's verdict", len(whole), whole[503].GetType())
	}
	fact := func(r unchecked) string {
		return fmt.Sprintf("%s's sentence changed at edit %d and no sitting opened since has checked it (its last check is event %d)", r.anchor, r.edit, r.lastCheck)
	}

	// The record before the chair's verdict.
	run := m19At(t, whole, 503)
	// WHAT THE MIGRATION AND THE DELIVERY MODEL ADD, derived rather than listed. Each close the
	// migration lost (its manifest's lost_acts) leaves its gap open, so its material_gap holds the
	// PASS, and a settled inquiry such a gap is about is held through that row, not its own
	// (inquiryBlockersOf). And every lens that sat owes the whole report (plan constitution-gaps
	// §III.3): the archive's lenses hold fewer than two report reads each, so each owes it.
	lostClose := map[string]bool{}
	for _, a := range lost {
		if a.Word == "close" && a.TS < whole[503].GetTs() {
			lostClose[a.Gap] = true
		}
	}
	heldByLostClose := map[string]bool{}
	sat, reads := map[string]bool{}, map[string]int{}
	for _, e := range whole[:503] {
		if m := e.GetMint(); m != nil && m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY && lostClose[m.GetGapId()] {
			heldByLostClose[m.GetAboutRef()] = true
		}
		if seatclass.RoleOfSeat(e.GetSeatId()) != "lens" {
			continue
		}
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_REGISTER:
			sat[e.GetSeatId()] = true
		case recordpb.EventType_EVENT_TYPE_REPORT_READ:
			reads[e.GetSeatId()]++
		}
	}
	owesWhole := map[string]bool{}
	for l := range sat {
		if reads[l] < 2 {
			owesWhole[l] = true
		}
	}
	if len(lostClose) == 0 || len(owesWhole) == 0 {
		t.Fatalf("m19 lost no close (%v) or no lens owes the whole report (%v): the arms below run on nothing", lostClose, owesWhole)
	}
	gotOpen, gotOwes := map[string]bool{}, map[string]bool{}
	blockers, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	// AND EACH OWNER IS READY: it owes the check, so it is not retired, and its pin is behind the head
	// (retirement.go). THE ARCHIVE PREDATES THE REPORT READ, so no lens of it was ever served the
	// report on the record and every unretired lens is behind the head: an owner's reason names its
	// reopened check, any other's that it was never served the report.
	owners := map[string]bool{computationSeat: true, evLens: true}
	var got []unchecked
	var ready []string
	inquiries := 0
	for _, b := range blockers {
		if b.Kind == BlockerLensReady {
			switch {
			case !strings.Contains(b.foldWhy, "no sitting it sat through holds a complete delivery of report text"):
				t.Errorf("%s is ready for %q, want it behind the head: the archive records no report read", b.Subject, b.foldWhy)
			case owners[b.Subject] && (!owesClause(b.foldWhy) || !strings.Contains(b.foldWhy, "unchecked ")):
				t.Errorf("%s is ready for %q, want it unretired because it owes a reopened check", b.Subject, b.foldWhy)
			}
			ready = append(ready, b.Subject)
			continue
		}
		// AND EVERY INQUIRY IT SETTLED: m19's chair read its inquiries and no check of any stands, so
		// each is owed the logic lens's.
		if b.Kind == BlockerUncheckedInquiry && b.Owner == inquiryOwner {
			inquiries++
			continue
		}
		if b.Kind == BlockerMaterialGap && lostClose[b.Subject] {
			gotOpen[b.Subject] = true
			continue
		}
		if b.Kind == BlockerWholeReportOwed && owesWhole[b.Subject] && b.Owner == b.Subject {
			gotOwes[b.Subject] = true
			continue
		}
		if b.Kind != BlockerUnchecked {
			t.Errorf("%s (%q, owned by %q) holds the PASS as well, want only the reopened anchors, their owners' readiness and the settled inquiries", b.Kind, b.Subject, b.Owner)
			continue
		}
		r := unchecked{anchor: b.Subject, computation: b.Detail != "", sha: b.Detail, edit: b.since, lastCheck: b.lastCheck}
		if b.Owner != m19Owner(r) {
			t.Errorf("reopened %s is owned by %q, want %s", b.Subject, b.Owner, m19Owner(r))
		}
		got = append(got, r)
	}
	if !maps.Equal(gotOpen, lostClose) {
		t.Errorf("material gaps holding the PASS = %v, want each gap whose close the migration lost: %v", gotOpen, lostClose)
	}
	if !maps.Equal(gotOwes, owesWhole) {
		t.Errorf("lenses owing the whole report = %v, want every lens that sat with fewer than two report reads: %v", gotOwes, owesWhole)
	}
	if want := 10 - len(heldByLostClose); inquiries != want {
		t.Errorf("%d settled inquiries hold the PASS unchecked, want m19's 10 less the %d a lost close's open gap holds instead (%v)", inquiries, len(heldByLostClose), heldByLostClose)
	}
	if !slices.Equal(got, m19Reopened) {
		t.Fatalf("reopened blockers = %+v\nwant %+v", got, m19Reopened)
	}
	if slices.Sort(ready); !slices.Contains(ready, computationSeat) || !slices.Contains(ready, evLens) {
		t.Errorf("ready lenses = %q, want the two owners among them", ready)
	}
	refusal := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if refusal == nil {
		t.Fatal("the gate admits the PASS m19's chair recorded at event 504")
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
		owed := uncheckedOwed(Blocker{Owner: m19Owner(r)})
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
		if err != nil || at.Unchecked == nil || *at.Unchecked != (EvidenceUncheckedJSON{Anchor: r.anchor, Edit: r.edit, LastCheck: r.lastCheck}) {
			t.Errorf("show evidence --anchor %s: reopened = %+v (%v), want edit %d, last check %d", r.anchor, at.Unchecked, err, r.edit, r.lastCheck)
		}
	}
	if !strings.Contains(chair, "`reproduce --anchor "+m19Reopened[2].anchor+"` records your verdict on its sentence, executing nothing once the bundle has a standing re-run") {
		t.Errorf("the computation's item does not carry the anchor its judgement takes:\n%s", chair)
	}
	var rows []unchecked
	for _, r := range standingOf(t, run) {
		rows = append(rows, unchecked{anchor: r.Anchor, edit: r.Edit, lastCheck: r.LastCheck})
	}
	if want := []unchecked{{anchor: "C-706fda63", edit: 415, lastCheck: 180}, {anchor: "C-4dae2507", edit: 389, lastCheck: 348}, {anchor: "K-cd9f4c5a", edit: 389, lastCheck: 347}}; !slices.Equal(rows, want) {
		t.Errorf("show evidence reopened = %+v\nwant %+v", rows, want)
	}
	// A computation reopened and retired in one blue sitting (edits 448 and 451, retire 452) owes nothing.
	if at, err := EvidenceAt(run, "K-c1c0989f"); err != nil || at.Computation == nil || at.Unchecked != nil {
		t.Errorf("show evidence --anchor K-c1c0989f: reopened = %+v (%v), want the computation with no reopened row", at.Unchecked, err)
	}

	// The two plans that followed the edits: the chair's dispatches at events 396 and 422 readied
	// architecture, voice and blue, then architecture alone. Each now readies the evidence lens for
	// the citations and the computation lens for the computation.
	for _, c := range []struct {
		before int
		owed   []string
	}{
		{396, []string{"C-706fda63", "C-4dae2507", "K-cd9f4c5a"}},
		{422, []string{"C-706fda63", "C-4dae2507", "K-cd9f4c5a"}},
	} {
		plan, err := PlanDispatch(m19At(t, whole, c.before-1))
		if err != nil {
			t.Fatal(err)
		}
		if _, ready := partyOf(plan, evLens); !ready || plan.PassPermitted {
			t.Errorf("the plan at event %d: parties %+v, pass_permitted %v — want %s readied and no PASS", c.before, plan.Parties, plan.PassPermitted, evLens)
		}
		if _, ready := partyOf(plan, computationSeat); !ready {
			t.Errorf("the plan at event %d: parties %+v — want %s readied for the computation it owns", c.before, plan.Parties, computationSeat)
		}
		// Each owner is readied because it owes the anchors' checks: its reason names every one it owns.
		for _, r := range m19Reopened {
			if !slices.ContainsFunc(plan.Why, func(w string) bool {
				return strings.HasPrefix(w, m19Owner(r)+": ") && strings.Contains(w, ", behind head ") && owesClause(w) && strings.Contains(w, "unchecked "+r.anchor)
			}) {
				t.Errorf("the plan at event %d does not ready %s for %s: %q", c.before, m19Owner(r), r.anchor, plan.Why)
			}
		}
		var owed []string
		for _, b := range plan.Blockers {
			if b.Kind == BlockerUnchecked {
				owed = append(owed, b.Subject)
			}
		}
		if !slices.Equal(owed, c.owed) {
			t.Errorf("the plan at event %d holds reopened %q, want %q", c.before, owed, c.owed)
		}
	}
}
