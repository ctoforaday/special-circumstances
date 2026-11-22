package migrate

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// partBRuns is every archive Part B's translations read: run-archive/'s runs and the two bundle-store
// runs the docket census reads beside them.
func partBRuns(t *testing.T) []string {
	t.Helper()
	return append(archivedRuns(t), "bundle-store/2026-10-31_is-91-prime-m37.tar.gz", "bundle-store/2026-11-04_is-91-prime-m38.tar.gz")
}

// sourceLink is one archived `supersedes` entry: the successor, the ancestor, the entry's place in
// the successor's list, and each mint's writer as the source records it.
type sourceLink struct {
	successor, ancestor string
	ord                 int
}

func (l sourceLink) String() string { return fmt.Sprintf("%s<-%s", l.successor, l.ancestor) }

// sourceLinksOf is every `supersedes` entry the archive's mints carry, in record order.
func sourceLinksOf(evs []OldEvent) []sourceLink {
	var out []sourceLink
	for _, e := range evs {
		if e.Word != "mint" {
			continue
		}
		g, _ := e.Fields["gap_id"].(string)
		sup := e.Lists["supersedes"]
		if s, ok := e.Fields["supersedes"].(string); ok && s != "" {
			sup = append(sup, s)
		}
		for i, a := range sup {
			out = append(out, sourceLink{g, a, i})
		}
	}
	return out
}

// migratedMints is each migrated gap's ancestor and the seat of its Mint event.
func migratedMints(t *testing.T, dir string) (ancestor, writer map[string]string, run record.Run) {
	t.Helper()
	run, err := record.OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	ancestor, writer = map[string]string{}, map[string]string{}
	for _, e := range fam.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			ancestor[m.GetGapId()], writer[m.GetGapId()] = m.GetSupersedes(), e.GetSeatId()
		}
	}
	return ancestor, writer, run
}

// migratedGapID is the id the migrated record gives an archived gap id: respelled where the
// manifest says so, else the same.
func migratedGapID(man *migration.Manifest, id string) string {
	if v, ok := man.IDs[id]; ok {
		return v
	}
	return id
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

// T-B11. A MINT KEEPS ITS FIRST ANCESTOR (Translation 1, §III.B11). Three archived mints name a second
// ancestor, each dropped; the first stands where it is the writer's own. quadratic-formula's R2-7 names
// R1-4, and R1-9, which R2-3 also superseded, reads closed on the migrated record.
func TestAMintKeepsItsFirstAncestor(t *testing.T) {
	want := map[string][]string{
		"2026-08-23_sleeper-service-plan.tar.gz": {"R3-3<-R2-3"},
		"2026-09-02_quadratic-formula.tar.gz":    {"R2-7<-R1-9"},
		"2026-11-12_is-91-prime-1838.tar.gz":     {"G-a9991ec0<-G-6a8778c0"},
	}
	got := map[string][]string{}
	for _, name := range partBRuns(t) {
		man, dir := MigratedArchive(t, name)
		ancestor, _, _ := migratedMints(t, dir)
		for _, l := range sourceLinksOf(sourceEvents(t, name)) {
			if l.ord == 0 {
				continue
			}
			got[name] = append(got[name], l.String())
			if a := ancestor[migratedGapID(man, l.successor)]; a == migratedGapID(man, l.ancestor) {
				t.Errorf("%s: %s still names its second ancestor %s", name, l.successor, l.ancestor)
			}
		}
	}
	if !maps.EqualFunc(got, want, slices.Equal) {
		t.Errorf("second ancestors dropped:\n%v\nwant:\n%v", got, want)
	}
	man, dir := MigratedArchive(t, "2026-09-02_quadratic-formula.tar.gz")
	ancestor, _, run := migratedMints(t, dir)
	r27, r14, r19 := migratedGapID(man, "R2-7"), migratedGapID(man, "R1-4"), migratedGapID(man, "R1-9")
	if ancestor[r27] != r14 {
		t.Errorf("quadratic-formula R2-7 (%s) names %q, want R1-4 (%s)", r27, ancestor[r27], r14)
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	if g := fam.Gap(r19); g == nil || g.Open {
		t.Errorf("quadratic-formula R1-9 (%s) is not closed on the migrated record: %+v", r19, g)
	}
}

// T-B11. EVERY CROSS-SEAT LINEAGE LINK DROPS (Translation 2, §III.B7). A first ancestor written by
// another seat — seats compared as the migrated record holds them, so the merge era's links, all the
// chair's, stand — is dropped; each is named here by run, successor and ancestor. A blue engagement
// whose sitting then owes no successor is written as no event, by source id. Every other first
// ancestor stands, and every archive migrates with zero refusals.
func TestACrossSeatLineageLinkDrops(t *testing.T) {
	wantDropped := map[string][]string{
		"2026-10-18_is-91-prime-m28.tar.gz":              {"G-79f965fd<-G-1e4a7bb0"},
		"2026-10-24_is-91-prime-1565.tar.gz":             {"G-bf2de55d<-G-82c43e74"},
		"2026-10-24_is-91-prime-m32.tar.gz":              {"G-cc9fe904<-G-11c8fa75"},
		"2026-10-27_is-91-prime-m35.tar.gz":              {"G-86478f5a<-G-af7292e8"},
		"2026-11-04_eiffel-tower-height.tar.gz":          {"G-50d8e3db<-G-164778cb", "G-00a137f3<-G-2d17fcf4", "G-92a782b9<-G-65119bba", "G-b7596566<-G-cec1c0f9", "G-b2ba7d66<-G-b0c8d93b"},
		"2026-11-11_eiffel-tower-height.tar.gz":          {"G-4338072a<-G-2bb7b6ec", "G-581f7470<-G-2bb7b6ec"},
		"2026-11-12_is-91-prime-1838.tar.gz":             {"G-6cafa44e<-G-7395c02f"},
		"bundle-store/2026-10-31_is-91-prime-m37.tar.gz": {"G-011bc585<-G-f822ce4a", "G-37347c3c<-G-e8929d64", "G-4febbcc9<-G-55eb14f9"},
		"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": {"G-29f07d26<-G-eb79c13d"},
	}
	wantUnowed := map[string][]int64{
		"2026-10-24_is-91-prime-1565.tar.gz":             {906},
		"2026-11-11_eiffel-tower-height.tar.gz":          {647},
		"bundle-store/2026-10-31_is-91-prime-m37.tar.gz": {1186},
	}
	gotDropped, gotUnowed, kept := map[string][]string{}, map[string][]int64{}, map[string]int{}
	for _, name := range partBRuns(t) {
		man, dir := MigratedArchive(t, name)
		if len(man.Refusals) != 0 {
			t.Errorf("%s: %d refusal(s); the first: %+v", name, len(man.Refusals), man.Refusals[0])
		}
		ancestor, _, _ := migratedMints(t, dir)
		for _, l := range sourceLinksOf(sourceEvents(t, name)) {
			if l.ord != 0 {
				continue
			}
			switch ancestor[migratedGapID(man, l.successor)] {
			case migratedGapID(man, l.ancestor):
				kept[name]++
			case "":
				gotDropped[name] = append(gotDropped[name], l.String())
			default:
				t.Errorf("%s: %s names %q, neither its first ancestor %s nor none", name, l.successor, ancestor[migratedGapID(man, l.successor)], l.ancestor)
			}
		}
		if len(man.UnowedEngagements) > 0 {
			gotUnowed[name] = man.UnowedEngagements
		}
	}
	if !maps.EqualFunc(gotDropped, wantDropped, slices.Equal) {
		t.Errorf("cross-seat links dropped:\n%#v\nwant:\n%#v", gotDropped, wantDropped)
	}
	if !maps.EqualFunc(gotUnowed, wantUnowed, slices.Equal) {
		t.Errorf("engagements written as no event:\n%#v\nwant:\n%#v", gotUnowed, wantUnowed)
	}
	n := 0
	for _, l := range gotDropped {
		n += len(l)
	}
	if n != 16 {
		t.Errorf("%d cross-seat links dropped, want 16", n)
	}
	if kept["2026-09-02_quadratic-formula.tar.gz"] != 14 {
		t.Errorf("quadratic-formula keeps %d chair links, want 14", kept["2026-09-02_quadratic-formula.tar.gz"])
	}
}

// T-B11. EVERY ARCHIVED GAP IS ASSIGNED TO ITS CHAIN'S ORIGIN (§III.B4a, §III.B7). Each migrated gap's
// stored assignee, as the board reads it, is the writer of the first gap of its chain recomputed from
// the migrated mints — and, with every kept link one seat's, its own writer.
func TestEveryArchivedGapIsAssignedToItsChainsOrigin(t *testing.T) {
	gaps := 0
	for _, name := range partBRuns(t) {
		_, dir := MigratedArchive(t, name)
		ancestor, writer, run := migratedMints(t, dir)
		bj, err := record.BoardJSONOfRun(run)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range append(bj.Open, bj.Closed...) {
			gaps++
			origin := g.ID
			for ancestor[origin] != "" {
				origin = ancestor[origin]
			}
			if g.Assignee != writer[origin] || g.Assignee != writer[g.ID] {
				t.Errorf("%s %s: assignee %q, its chain's origin %s written by %q, itself by %q", name, g.ID, g.Assignee, origin, writer[origin], writer[g.ID])
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
// gap's fix_new verbatim, an end after each reword of such a sentence, and a ruling placement after the
// two non-remand docket rulings the replay keeps. Counted per run, bundle-store's m37 and m38 included, the ends and the ruling by gap; no
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
		"2026-11-04_eiffel-tower-height.tar.gz":          {"G-42584956"},
		"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": {"G-ca807516"},
	}
	gotPlaced, gotEnded, gotRuled := map[string]int{}, map[string][]string{}, map[string][]string{}
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
				if _, ruling := prev.(*recordpb.MotionRule); !ruling || tb.GetEnds() {
					t.Errorf("%s: the ruling row on %s (ends %v) follows %T, not a docket ruling", name, tb.GetGapId(), tb.GetEnds(), prev)
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
	if total != 44 {
		t.Errorf("%d tombstone rows over the archives, want 44: 37 estoppel placements, 5 ends, 2 rulings", total)
	}
}
