package migrate_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// The bundle-store archives the census reads beside run-archive/'s top level: m37's six impasse
// dockets and m38's two bench regrades are Translations 3 and 4's cases too.
var docketCensusExtras = []string{
	"bundle-store/2026-10-31_is-91-prime-m37.tar.gz",
	"bundle-store/2026-11-04_is-91-prime-m38.tar.gz",
}

// docketCensus is every archive Translations 3 and 4 read.
func docketCensus(t *testing.T) []string {
	t.Helper()
	var names []string
	for _, tb := range archiveTarballs(t) {
		names = append(names, filepath.Base(tb))
	}
	return append(names, docketCensusExtras...)
}

// archivedDocketing is one docketing of a migrated run, as the test reads it.
type archivedDocketing struct {
	gap, cause string
	standing   bool
	ruling     *recordpb.DocketRuling
}

// docketingsOfMigrated is the migrated run's docketings in record order, and its docket rulings.
func docketingsOfMigrated(t *testing.T, name string) (*migration.Manifest, []archivedDocketing, []*recordpb.MotionRule, []*record.Event) {
	t.Helper()
	res, run := migrateArchive(t, name)
	merged, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []archivedDocketing
	for _, d := range record.DocketingsOf(merged.Events) {
		a := archivedDocketing{gap: d.GapID, cause: recordpb.Word(d.Cause), standing: d.Standing()}
		if d.Ruling != nil {
			a.ruling = d.Ruling.GetDocket()
		}
		out = append(out, a)
	}
	var rulings []*recordpb.MotionRule
	for _, e := range merged.Events {
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetDocket() != nil {
			rulings = append(rulings, r)
		}
	}
	return res, out, rulings, merged.Events
}

// docketResidue is the stated residue a run's manifest carries, per list (migration.DocketResidue).
type docketResidue struct {
	opinions, party, superseded, ended, benchRegrades, spotChecks, inquiryChecks []int64
}

// wantDocketResidue is the measured residue of Translations 3 and 4 per run (probe of 2026-11-22,
// plans/findings-fold.md §III.B14): the source ids each list names. A run absent states none.
var wantDocketResidue = map[string]docketResidue{
	"2026-08-22_is-7-prime-run2.tar.gz":            {opinions: []int64{149, 150, 151, 152, 153, 154, 155}},
	"2026-08-22_is-7-prime.tar.gz":                 {opinions: []int64{147, 148}},
	"2026-08-22_sqlite-schema-source.tar.gz":       {opinions: []int64{210, 211}},
	"2026-08-23_research-loop-counterparts.tar.gz": {opinions: []int64{529, 530, 531}},
	"2026-08-23_sleeper-service-plan.tar.gz": {opinions: []int64{296, 412, 413, 414, 415, 416, 514, 515, 518, 519},
		spotChecks: []int64{474}},
	"2026-09-02_quadratic-formula.tar.gz": {opinions: []int64{456, 457, 458, 459, 460, 461, 462, 463, 659, 660, 661, 662, 663, 664, 665, 741, 804},
		spotChecks: []int64{542, 710, 766}},
	"2026-09-09_is-91-prime-a.tar.gz":                {party: []int64{132, 160}, spotChecks: []int64{163}},
	"2026-09-10_is-91-prime-b3.tar.gz":               {party: []int64{150, 173}},
	"2026-09-11_is-91-prime-b6.tar.gz":               {party: []int64{179, 180, 186, 187}},
	"2026-10-12_is-91-prime-wave-b.tar.gz":           {party: []int64{325, 326, 431, 432, 489, 609, 615, 672}, spotChecks: []int64{626, 698}},
	"2026-10-17_is-91-prime-07d.tar.gz":              {party: []int64{372, 383}},
	"2026-10-18_is-91-prime-m29.tar.gz":              {superseded: []int64{351, 426}},
	"2026-10-24_is-91-prime-1565.tar.gz":             {ended: []int64{557, 558, 559, 560, 872}},
	"2026-10-27_is-91-prime-m35.tar.gz":              {party: []int64{404, 520, 584, 586}, ended: []int64{580, 581, 583}},
	"2026-10-30_mersenne-31-prime-sum.tar.gz":        {party: []int64{541, 542, 667, 668, 780, 836, 881, 882}, ended: []int64{792}},
	"2026-11-04_eiffel-tower-height.tar.gz":          {party: []int64{777, 885}},
	"2026-11-11_eiffel-tower-height.tar.gz":          {benchRegrades: []int64{669}},
	"2026-11-14_is-91-prime-1883.tar.gz":             {party: []int64{266, 459}, inquiryChecks: []int64{489}},
	"bundle-store/2026-10-31_is-91-prime-m37.tar.gz": {superseded: []int64{471, 745}, ended: []int64{971, 1196, 1197}},
	"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": {benchRegrades: []int64{427, 718}},
}

// wantDocketings is each run's docketings, by gap in record order, with the cause the live rule
// gives each (grade movement, else impasse).
var wantDocketings = map[string][]string{
	"2026-10-18_is-91-prime-m28.tar.gz":  {"G-1e4a7bb0 impasse", "G-3a2c77c1 impasse"},
	"2026-10-18_is-91-prime-m29.tar.gz":  {"G-6f6f47fb grade_movement"},
	"2026-10-24_is-91-prime-1565.tar.gz": {"G-f94c09fa impasse", "G-960d94a9 impasse", "G-620d9a8f impasse", "G-fe192a8a impasse", "G-5a88e946 impasse"},
	"2026-10-24_is-91-prime-m32.tar.gz":  {"G-fbdd91f9 impasse"},
	"2026-10-27_is-91-prime-m35.tar.gz": {"G-91f0c8e6 impasse", "G-08b7972e impasse", "G-1f1ae21f impasse", "G-18a9b00c impasse",
		"G-af7292e8 impasse", "G-650ee728 impasse"},
	"2026-10-30_mersenne-31-prime-sum.tar.gz":        {"G-221322e1 impasse"},
	"2026-11-04_eiffel-tower-height.tar.gz":          {"G-72aab1f2 impasse", "G-42584956 impasse"},
	"2026-11-11_eiffel-tower-height.tar.gz":          {"G-6e7c400f impasse"},
	"bundle-store/2026-10-31_is-91-prime-m37.tar.gz": {"G-c3f84f98 impasse", "G-deced912 impasse", "G-deced912 impasse", "G-37347c3c impasse"},
	"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": {"G-845422d8 impasse", "G-ca807516 impasse"},
}

// T-B26. AN ARCHIVED DOCKET BECOMES A DOCKETING (EventSchema 80, §III.B13 Translation 3). Every
// docket the dispatch filed (O3, structural) is a Docketing at its place; its ruling names that
// docketing and no motion; every other docket and its ruling are written as no event (O2), with the
// opinion era's pairs, the dispatch's dockets left standing on a gap a successor superseded, and the
// spot-checks of a gap those drops left open, and the inquiry checks only such a gap holds wanting — each stated in the manifest. No migrated record holds a
// docket motion, and none is refused.
func TestAnArchivedDocketBecomesADocketing(t *testing.T) {
	total := map[string]int{}
	for _, name := range docketCensus(t) {
		res, ks, rulings, evs := docketingsOfMigrated(t, name)
		if len(res.Refusals) != 0 {
			t.Errorf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
		}
		want := wantDocketResidue[name]
		for _, l := range []struct {
			what      string
			got, want []int64
		}{
			{"opinions", res.Dockets.Opinions, want.opinions},
			{"party dockets", res.Dockets.PartyDockets, want.party},
			{"superseded dockets", res.Dockets.Superseded, want.superseded},
			{"ended rulings", res.Dockets.EndedRulings, want.ended},
			{"bench regrades", res.Dockets.BenchRegrades, want.benchRegrades},
			{"spot-checks", res.Dockets.SpotChecks, want.spotChecks},
			{"inquiry checks", res.Dockets.InquiryChecks, want.inquiryChecks},
		} {
			if l.got == nil {
				t.Errorf("%s: the manifest's %s list is null, not a stated list", name, l.what)
			}
			if !slices.Equal(l.got, l.want) && !(len(l.got) == 0 && len(l.want) == 0) {
				t.Errorf("%s: %s written as no event %v, want %v", name, l.what, l.got, l.want)
			}
			total[l.what] += len(l.got)
		}
		var got []string
		for _, k := range ks {
			got = append(got, k.gap+" "+k.cause)
		}
		if !slices.Equal(got, wantDocketings[name]) {
			t.Errorf("%s: docketings %v, want %v", name, got, wantDocketings[name])
		}
		total["docketings"] += len(ks)
		ids := map[string]bool{}
		for _, e := range evs {
			if k, ok := recordpb.BodyAs[*recordpb.Docketing](e); ok {
				ids[k.GetDocketingId()] = true
				if seatclass.RoleOfSeat(e.GetSeatId()) != "chair" {
					t.Errorf("%s: docketing %s is written by %s, want the chair's dispatch", name, k.GetDocketingId(), e.GetSeatId())
				}
			}
			if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok && m.GetDocket() != nil && m.GetAct() == nil {
				t.Errorf("%s: docket motion %s stores no act", name, m.GetMotionId())
			}
		}
		for _, r := range rulings {
			if r.GetMotionId() != "" || !ids[r.GetDocket().GetDocketingId()] {
				t.Errorf("%s: docket ruling names motion %q and docketing %q, want no motion and a docketing on the record", name, r.GetMotionId(), r.GetDocket().GetDocketingId())
			}
		}
		// m35 G-18a9b00c: the dark-side's own docket (ev 404) drops, so the dispatch's is its one
		// docketing, ended by its assignee's close ev 517, and its ruling ev 583 drops.
		if name == "2026-10-27_is-91-prime-m35.tar.gz" {
			n := 0
			for _, k := range ks {
				if k.gap == "G-18a9b00c" {
					n++
					if k.standing || k.ruling != nil {
						t.Errorf("m35 G-18a9b00c: its docketing stands=%v, ruled=%v; want ended by its assignee's close, unruled", k.standing, k.ruling != nil)
					}
				}
			}
			if n != 1 {
				t.Errorf("m35 G-18a9b00c: %d docketing(s), want 1", n)
			}
		}
	}
	t.Logf("written as no event: %v; docketings written: %d", total, total["docketings"])
	for what, want := range map[string]int{"opinions": 41, "party dockets": 34, "superseded dockets": 4, "ended rulings": 12,
		"bench regrades": 3, "spot-checks": 7, "inquiry checks": 1, "docketings": 25} {
		if total[what] != want {
			t.Errorf("census: %s %d, want %d", what, total[what], want)
		}
	}
}

// T-B31. AN ARCHIVED BENCH WORD TAKES ITS SIDE (EventSchema 80, §III.B14 Translation 4). Each
// docketing's cause is the live rule's, printed beside the archived basis's impasse: every one is
// impasse but m29 G-6f6f47fb, whose lens's regrade (ev 540) moved it past the grade-movement threshold
// (ruled 2026-11-22: the live cause stands). A kept closing word is a for-blue answer with blue stating
// none; a remand stays a remand with neither side nor grades; no ruling sets moved_grades; no bench
// Regrade is on any migrated record; a docketing a Close ended reads closed by its assignee; m35
// G-650ee728 stands unruled at the run's end, and so does m29 G-6f6f47fb: its raiser's close (ev 561)
// and, with it, its ruling (ev 576) are lost acts (the close stands on no delivered text; the ruling,
// on a grade-movement docketing, keeps or restores no grade).
func TestAnArchivedBenchWordTakesItsSide(t *testing.T) {
	type kept struct{ gap, answer string }
	wantKept := map[string][]kept{
		"2026-10-18_is-91-prime-m28.tar.gz":              {{"G-1e4a7bb0", "blue not_a_defect"}, {"G-3a2c77c1", "blue not_a_defect"}},
		"2026-10-24_is-91-prime-m32.tar.gz":              {{"G-fbdd91f9", "blue not_a_defect"}},
		"2026-10-27_is-91-prime-m35.tar.gz":              {{"G-91f0c8e6", "remanded"}, {"G-1f1ae21f", "blue not_a_defect"}},
		"2026-11-04_eiffel-tower-height.tar.gz":          {{"G-72aab1f2", "remanded"}, {"G-42584956", "blue not_a_defect"}},
		"2026-11-11_eiffel-tower-height.tar.gz":          {{"G-6e7c400f", "remanded"}},
		"bundle-store/2026-10-31_is-91-prime-m37.tar.gz": {{"G-deced912", "remanded"}},
		"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": {{"G-845422d8", "remanded"}, {"G-ca807516", "blue not_a_defect"}},
	}
	for _, name := range docketCensus(t) {
		_, ks, _, evs := docketingsOfMigrated(t, name)
		run := runOf(t, name)
		fam, err := record.FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		var got []kept
		for _, k := range ks {
			t.Logf("%s %s: live cause %s, archived basis impasse", name, k.gap, k.cause)
			wantCause := "impasse"
			if name == "2026-10-18_is-91-prime-m29.tar.gz" && k.gap == "G-6f6f47fb" {
				wantCause = "grade_movement"
			}
			if k.cause != wantCause {
				t.Errorf("%s %s: cause %s, want %s", name, k.gap, k.cause, wantCause)
			}
			switch {
			case k.ruling != nil:
				r := k.ruling
				if r.MovedGrades != nil {
					t.Errorf("%s %s: a migrated ruling sets moved_grades", name, k.gap)
				}
				answer := recordpb.Word(r.GetDisposition())
				if r.SidedWith != nil {
					answer = fmt.Sprintf("%s %s", recordpb.Word(r.GetSidedWith()), answer)
				}
				got = append(got, kept{k.gap, answer})
			case k.standing:
				if !(name == "2026-10-27_is-91-prime-m35.tar.gz" && k.gap == "G-650ee728") && !(name == "2026-10-18_is-91-prime-m29.tar.gz" && k.gap == "G-6f6f47fb") {
					t.Errorf("%s %s: a docketing stands at the run's end", name, k.gap)
				}
			default:
				g := fam.Gap(k.gap)
				if g == nil || g.Open || g.ClosedByBench {
					t.Errorf("%s %s: its docketing ended unruled and the gap is not closed by its assignee", name, k.gap)
				}
			}
		}
		if !slices.Equal(got, wantKept[name]) && !(len(got) == 0 && len(wantKept[name]) == 0) {
			t.Errorf("%s: kept rulings %v, want %v", name, got, wantKept[name])
		}
		for _, e := range evs {
			if _, ok := recordpb.BodyAs[*recordpb.Regrade](e); ok && seatclass.RoleOfSeat(e.GetSeatId()) == "bench" {
				t.Errorf("%s: a bench regrade is on the migrated record", name)
			}
		}
	}
}

// runOf is the migrated run of the named archive.
func runOf(t *testing.T, name string) record.Run {
	t.Helper()
	_, run := migrateArchive(t, name)
	return run
}
