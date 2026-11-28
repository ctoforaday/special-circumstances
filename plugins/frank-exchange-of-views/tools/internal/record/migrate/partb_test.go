package migrate

import (
	"database/sql"
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// partBRuns is every archive Part B's translations read: the committed set, bundle-store/'s two runs
// among them.
func partBRuns(t *testing.T) []string {
	t.Helper()
	return archivedRuns(t)
}

// migratedMinters is the seat of each migrated gap's Mint event.
func migratedMinters(t *testing.T, dir string) (writer map[string]string, run record.Run) {
	t.Helper()
	run, err := record.OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	writer = map[string]string{}
	for _, e := range fam.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			writer[m.GetGapId()] = e.GetSeatId()
		}
	}
	return writer, run
}

// T-B11. AN ARCHIVED GRADE MOTION LEAVES NO EVENT (EventSchema 80, §III.B7). Each grade motion and each
// ruling on one, by source events.id: the motion subject is gone, and none moved a grade. Both eras
// are counted — the SQLite era's `grade` arm and the flat shard era's `subject = grade` — so an
// archive that adds one changes a row here. Every archive migrates with zero refusals.
func TestAnArchivedGradeMotionLeavesNoEvent(t *testing.T) {
	type motions struct{ motions, rulings []int64 }
	want := map[string]motions{
		"2026-08-23_research-loop-counterparts.tar.gz": {[]int64{304, 506}, []int64{354}},
		"2026-08-23_sleeper-service-plan.tar.gz":       {[]int64{187, 377, 487, 488}, []int64{232, 455}},
		"2026-09-02_quadratic-formula.tar.gz":          {[]int64{276, 443, 444, 642}, []int64{366, 540, 541, 708}},
		"2026-09-11_is-91-prime-b9.tar.gz":             {[]int64{195}, []int64{197}},
		"2026-10-12_is-91-prime-wave-b.tar.gz":         {[]int64{506, 507, 598}, []int64{527, 528, 622}},
		"2026-10-17_is-91-prime-07d.tar.gz":            {[]int64{368}, []int64{375}},
	}
	got := map[string]motions{}
	var nm, nr int
	for _, name := range partBRuns(t) {
		var m motions
		for _, e := range sourceEvents(t, name) {
			_, arm := e.Arms["grade"]
			subject, _ := e.Fields["subject"].(string)
			switch e.Word {
			case "motion":
				if arm || subject == "grade" && len(e.Arms) == 0 {
					m.motions = append(m.motions, e.ID)
				}
			case "motion_rule":
				v, ok := e.Fields["grade"]
				if arm || ok && v != nil || subject == "grade" {
					m.rulings = append(m.rulings, e.ID)
				}
			}
		}
		man, _ := MigratedArchive(t, name)
		if len(man.Refusals) != 0 {
			t.Errorf("%s: %d refusal(s); the first: %+v", name, len(man.Refusals), man.Refusals[0])
		}
		if len(m.motions)+len(m.rulings) > 0 {
			got[name] = m
			nm, nr = nm+len(m.motions), nr+len(m.rulings)
		}
	}
	if !maps.EqualFunc(got, want, func(a, b motions) bool {
		return slices.Equal(a.motions, b.motions) && slices.Equal(a.rulings, b.rulings)
	}) {
		t.Errorf("grade motions and their rulings by source id:\n%v\nwant:\n%v", got, want)
	}
	if nm != 15 || nr != 12 {
		t.Errorf("%d grade motions and %d rulings over the archives, want 15 and 12", nm, nr)
	}
}

// T-B43. A SUPERSEDE LINK BECOMES AN AMEND, OR NOTHING (Translation 7, §III.B8). Every archived link is
// classed on the record as migrated: an own-seat link to an open ancestor that keeps the class,
// subject, location and check kind is an Amend of the chain's first gap (a Regrade beside it where the
// grades differ), and every other drops, the successor standing as its own gap. Counted per outcome
// over every archive, with zero refusals; an amended successor holds no mint, and an amended
// ancestor's gap view reads the amend.
func TestASupersedeLinkBecomesAnAmend(t *testing.T) {
	outcomes, regraded := map[string]int{}, 0
	for _, name := range partBRuns(t) {
		man, dir := MigratedArchive(t, name)
		if len(man.Refusals) != 0 {
			t.Errorf("%s: %d refusal(s); the first: %+v", name, len(man.Refusals), man.Refusals[0])
		}
		db, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "records", "record.db")+"?mode=ro")
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		for _, l := range man.SupersedeLinks {
			outcomes[l.Outcome]++
			t.Logf("%s: %d %s<-%s %s %v regraded=%v", name, l.OldID, l.Successor, l.Ancestor, l.Outcome, l.Moved, l.Regraded)
			if l.Outcome != linkAmended {
				continue
			}
			if l.Regraded {
				regraded++
			}
			var minted, amended int
			if err := db.QueryRow(`SELECT count(*) FROM "mint" WHERE "gap_id" = ?`, l.Successor).Scan(&minted); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT count(*) FROM "gap" WHERE "gap_id" = ? AND "amend_event" IS NOT NULL`, l.Ancestor).Scan(&amended); err != nil {
				t.Fatal(err)
			}
			if minted != 0 || amended != 1 {
				t.Errorf("%s: %s amended into %s: %d mint(s) of the successor (want 0), ancestor reads an amend: %v", name, l.Successor, l.Ancestor, minted, amended == 1)
			}
		}
	}
	t.Logf("links by outcome %v, %d amended with a regrade", outcomes, regraded)
}

// T-B11. EVERY ARCHIVED GAP IS ITS MINTER'S (§III.R B5). Each migrated gap's minted_by, as the board
// reads it, is the seat of its own Mint event: no link hands a gap to another seat.
func TestEveryArchivedGapIsItsMinters(t *testing.T) {
	gaps := 0
	for _, name := range partBRuns(t) {
		_, dir := MigratedArchive(t, name)
		writer, run := migratedMinters(t, dir)
		bj, err := record.BoardJSONOfRun(run)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range append(bj.Open, bj.Closed...) {
			gaps++
			if g.Assignee != writer[g.ID] {
				t.Errorf("%s %s: minted by %q, its mint written by %q", name, g.ID, g.Assignee, writer[g.ID])
			}
		}
	}
	if gaps == 0 {
		t.Fatal("no migrated gap was read")
	}
	t.Logf("%d migrated gaps read", gaps)
}

// T-B20. ARCHIVED RUNS CARRY THEIR TOMBSTONES (§III.B12). The replay's write path writes each row right
// after the act that owes it: an estoppel placement after every archived blue edit that applied a
// gap's fix_new verbatim, an end after each reword of such a sentence, a ruling placement after each
// non-remand docket ruling the replay keeps, and a ruling's end after blue's reword of its gap's
// sentence (m29's two, on docketings Translation 7 keeps). Counted per run, bundle-store's m37 and m38
// included, the ends and the rulings by gap; no
// source holds a row, so events_in counts none and events_out every one; and no dispatch pin, stale
// read point or served head names a row.
func TestArchivedRunsCarryTheirTombstones(t *testing.T) {
	wantPlaced := map[string]int{
		"2026-08-22_record-store-authority.tar.gz":       3,
		"2026-08-23_research-loop-counterparts.tar.gz":   1,
		"2026-09-02_quadratic-formula.tar.gz":            2,
		"2026-10-17_is-91-prime-m26.tar.gz":              2,
		"2026-10-17_is-91-prime-m27.tar.gz":              2,
		"2026-10-18_is-91-prime-m28.tar.gz":              2,
		"2026-10-19_is-91-prime-m30.tar.gz":              1,
		"2026-10-24_is-91-prime-1565.tar.gz":             3,
		"2026-10-24_is-91-prime-m32.tar.gz":              2,
		"2026-10-29_is-91-prime-m36.tar.gz":              1,
		"2026-10-30_mersenne-31-prime-sum.tar.gz":        2,
		"2026-11-11_eiffel-tower-height.tar.gz":          3,
		"2026-11-12_is-91-prime-1838.tar.gz":             6,
		"2026-11-14_is-91-prime-1883.tar.gz":             3,
		"bundle-store/2026-10-31_is-91-prime-m37.tar.gz": 2,
		"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": 2,
	}
	wantEnded := map[string][]string{
		"2026-10-24_is-91-prime-1565.tar.gz":             {"G-82c43e74"},
		"2026-11-11_eiffel-tower-height.tar.gz":          {"G-2bb7b6ec"},
		"2026-11-12_is-91-prime-1838.tar.gz":             {"G-530b345f"},
		"bundle-store/2026-10-31_is-91-prime-m37.tar.gz": {"G-08b93634", "G-f822ce4a"},
	}
	wantRuled := map[string][]string{
		"2026-10-18_is-91-prime-m28.tar.gz":              {"G-1e4a7bb0", "G-3a2c77c1"},
		"2026-10-18_is-91-prime-m29.tar.gz":              {"G-bde527eb", "G-c55e607d", "G-237fed96"},
		"2026-10-19_is-91-prime-m30.tar.gz":              {"G-97b44713"},
		"2026-10-24_is-91-prime-m32.tar.gz":              {"G-fbdd91f9"},
		"2026-10-27_is-91-prime-m35.tar.gz":              {"G-1f1ae21f"},
		"2026-11-04_eiffel-tower-height.tar.gz":          {"G-42584956"},
		"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": {"G-ca807516"},
	}
	wantRuledEnded := map[string][]string{
		"2026-10-18_is-91-prime-m29.tar.gz": {"G-bde527eb", "G-c55e607d"},
	}
	gotPlaced, gotEnded, gotRuled, gotRuledEnded := map[string]int{}, map[string][]string{}, map[string][]string{}, map[string][]string{}
	total := 0
	for _, name := range partBRuns(t) {
		man, dir := MigratedArchive(t, name)
		run, err := record.OpenRun(dir)
		if err != nil {
			t.Fatal(err)
		}
		merged, err := record.MergedEvents(run)
		if err != nil {
			t.Fatal(err)
		}
		rows := map[int64]bool{}
		var prev any
		n := 0
		for _, e := range merged.Events {
			tb, ok := recordpb.BodyAs[*recordpb.Tombstone](e)
			if !ok {
				prev, _ = recordpb.Body(e)
				continue
			}
			rows[merged.At.Of(e).ID] = true
			n++
			switch tb.GetKind() {
			case recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL:
				if _, edit := prev.(*recordpb.BlueEdit); !edit {
					t.Errorf("%s: the estoppel row on %s follows %T, not a blue edit", name, tb.GetGapId(), prev)
				}
				if tb.GetEnds() {
					gotEnded[name] = append(gotEnded[name], tb.GetGapId())
				} else {
					gotPlaced[name]++
				}
			case recordpb.TombstoneKind_TOMBSTONE_KIND_RULING:
				if tb.GetEnds() {
					if _, edit := prev.(*recordpb.BlueEdit); !edit {
						t.Errorf("%s: the ruling's end on %s follows %T, not a blue edit", name, tb.GetGapId(), prev)
					}
					gotRuledEnded[name] = append(gotRuledEnded[name], tb.GetGapId())
					continue
				}
				if _, ruling := prev.(*recordpb.MotionRule); !ruling {
					t.Errorf("%s: the ruling row on %s follows %T, not a docket ruling", name, tb.GetGapId(), prev)
				}
				gotRuled[name] = append(gotRuled[name], tb.GetGapId())
			}
		}
		total += n
		if man.In["tombstone"] != 0 || man.Out["tombstone"] != n {
			t.Errorf("%s: the manifest counts %d tombstone(s) in and %d out; the record holds %d and the source none", name, man.In["tombstone"], man.Out["tombstone"], n)
		}
		for _, e := range merged.Events {
			var ref int64
			switch b, _ := recordpb.Body(e); b := b.(type) {
			case *recordpb.Dispatch:
				ref = b.GetPin()
			case *recordpb.Stale:
				ref = b.GetThrough()
			case *recordpb.ReportRead:
				ref = b.GetHead()
			}
			if rows[ref] {
				t.Errorf("%s: event %d (%s) names the tombstone row %d", name, merged.At.Of(e).ID, e.GetKey(), ref)
			}
		}
	}
	if !maps.Equal(gotPlaced, wantPlaced) {
		t.Errorf("estoppel placements per run %v, want %v", gotPlaced, wantPlaced)
	}
	if !maps.EqualFunc(gotEnded, wantEnded, slices.Equal) {
		t.Errorf("estoppel ends %v, want %v", gotEnded, wantEnded)
	}
	if !maps.EqualFunc(gotRuled, wantRuled, slices.Equal) {
		t.Errorf("ruling placements %v, want %v", gotRuled, wantRuled)
	}
	if !maps.EqualFunc(gotRuledEnded, wantRuledEnded, slices.Equal) {
		t.Errorf("ruling ends %v, want %v", gotRuledEnded, wantRuledEnded)
	}
	if total != 54 {
		t.Errorf("%d tombstone rows over the archives, want 54: 37 estoppel placements, 5 ends, 10 rulings, 2 ruling ends", total)
	}
}

// T-B13 (sides). NO ARCHIVE HOLDS AN ACCEPT AS A RULING (§III.B8, Sides). Every migrated docket ruling
// is the bench's, names the docketing it answers, and sides with the author or with no one (a remand);
// no migrated act takes a motion (accepts) or concedes a docketing (accepted) — the raiser's and the
// author's accepts never ran in an archived run. Counted over every archive, so a zero is loud.
func TestNoArchiveHoldsAnAcceptAsARuling(t *testing.T) {
	rulings, sided := 0, 0
	for _, name := range partBRuns(t) {
		_, dir := MigratedArchive(t, name)
		run, err := record.OpenRun(dir)
		if err != nil {
			t.Fatal(err)
		}
		fam, err := record.FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		for i, e := range fam.Events {
			body, ok := recordpb.Body(e)
			if !ok {
				continue
			}
			switch b := body.(type) {
			case *recordpb.MotionRule:
				d := b.GetDocket()
				if d == nil {
					continue
				}
				rulings++
				if d.GetDocketingId() == "" || seatclass.RoleOfSeat(e.GetSeatId()) != "bench" {
					t.Errorf("%s event %d: a docket ruling by %s naming docketing %q: every docket ruling is the bench's, of a docketing", name, i, e.GetSeatId(), d.GetDocketingId())
				}
				switch d.GetSidedWith() {
				case recordpb.SidedWith_SIDED_WITH_AUTHOR:
					sided++
				case recordpb.SidedWith_SIDED_WITH_UNSPECIFIED:
				default:
					t.Errorf("%s event %d: a docket ruling sided %s; archived rulings side with the author", name, i, recordpb.Word(d.GetSidedWith()))
				}
			case *recordpb.Regrade:
				if b.GetAccepts() != "" || b.GetAccepted() || b.GetAnswers() != "" {
					t.Errorf("%s event %d: a regrade carrying an accept or an answer: %v", name, i, b)
				}
			case *recordpb.Close:
				if b.GetAccepts() != "" {
					t.Errorf("%s event %d: a close taking motion %s", name, i, b.GetAccepts())
				}
			case *recordpb.Retire:
				if b.GetAccepted() {
					t.Errorf("%s event %d: a retire carrying accepted", name, i)
				}
			}
		}
	}
	if rulings == 0 || sided == 0 {
		t.Fatalf("%d docket rulings, %d sided with the author, over the archives: the assertion read nothing", rulings, sided)
	}
	t.Logf("%d docket rulings, %d sided with the author", rulings, sided)
}

// T-B43, the amend's end (X18): an archived docketing a Translation 7 amend of its gap follows stands no
// longer, and its archived ruling after the amend is written as no event. Counted over every archive:
// the docketings an amend follows, unruled before it, on the migrated record — none (measured at R1's
// Commit C: no amended link follows a docketing of its gap) — and of those any still read as standing.
func TestAnAmendEndsTheArchivedDocketingItFollows(t *testing.T) {
	followed, standing := 0, 0
	for _, name := range partBRuns(t) {
		_, dir := MigratedArchive(t, name)
		db, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "records", "record.db")+"?mode=ro")
		if err != nil {
			t.Fatal(err)
		}
		var f, s int
		err = db.QueryRow(`SELECT count(*), coalesce(sum(g."standing_docketing" = dk."docketing_id"), 0)
		  FROM "docketing" dk JOIN "gap_state" g ON g."gap_id" = dk."gap_id"
		  WHERE dk."cause" <> 'grade_movement'
		    AND EXISTS (SELECT 1 FROM "amend" a WHERE a."gap_id" = dk."gap_id" AND a."event_id" > dk."event_id")
		    AND NOT EXISTS (SELECT 1 FROM "motion_rule_docket" rd JOIN "amend" a ON a."gap_id" = dk."gap_id"
		                    WHERE rd."docketing_id" = dk."docketing_id" AND rd."event_id" < a."event_id")`).Scan(&f, &s)
		db.Close()
		if err != nil {
			t.Fatal(err)
		}
		if f > 0 {
			t.Logf("%s: %d docketing(s) an amend follows, %d standing", name, f, s)
		}
		followed, standing = followed+f, standing+s
	}
	if standing != 0 {
		t.Errorf("%d docketing(s) an amend of their gap follows still stand", standing)
	}
	if followed != 0 {
		t.Errorf("%d archived docketing(s) an amend of their gap follows, want 0: state each outcome here", followed)
	}
	t.Logf("%d archived docketing(s) an amend of their gap follows", followed)
}
