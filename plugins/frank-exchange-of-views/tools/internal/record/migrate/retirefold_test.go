package migrate

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// Archive names the fold's tables key on.
const (
	b3Run       = "2026-09-10_is-91-prime-b3.tar.gz"
	b6Run       = "2026-09-11_is-91-prime-b6.tar.gz"
	b9Run       = "2026-09-11_is-91-prime-b9.tar.gz"
	waveARun    = "2026-10-09_is-91-prime-wave-a.tar.gz"
	m19Run      = "2026-10-10_is-91-prime-m19.tar.gz"
	m26Run      = "2026-10-17_is-91-prime-m26.tar.gz"
	m28Run      = "2026-10-18_is-91-prime-m28.tar.gz"
	m32Run      = "2026-10-24_is-91-prime-m32.tar.gz"
	m35Run      = "2026-10-27_is-91-prime-m35.tar.gz"
	m36Run      = "2026-10-29_is-91-prime-m36.tar.gz"
	mersenneRun = "2026-10-30_mersenne-31-prime-sum.tar.gz"
	eiffelRun   = "2026-11-04_eiffel-tower-height.tar.gz"
	run1883     = "2026-11-14_is-91-prime-1883.tar.gz"
	m37Run      = "bundle-store/2026-10-31_is-91-prime-m37.tar.gz"
)

// cutsOf is every removed_anchors list the migrated archives hold, by archive, in event order: the
// folds of claim retires into the cuts that bared their tokens (P12b translation 2, F18, F23) and the cuts
// no retire followed (P12b translation 4). No list names a finding: the findings fold places no `F-`
// token (§III.A5), so the cuts that took only finding anchors (m19's, m28's first, m32's, m35's — F22)
// take nothing and list none.
var cutsOf = map[string][][]string{
	b3Run:    {{"G-19b1727b"}},
	b6Run:    {{"G-68c6ecf7"}},
	b9Run:    {{"G-8bd21b5d"}},
	waveARun: {{"C-3743f72c", "G-f6f63173"}},
	m28Run:   {{"G-3a2c77c1"}},
	m37Run:   {{"G-20128984", "G-477feea1"}},
	run1883:  {{"G-05dad92a", "G-3dc8d8e2"}},
}

// migratedArchive is one archive migrated, read: its manifest, events with their ids, the source's
// events, the final report ("" with no part) and the board.
type migratedArchive struct {
	res    *migration.Manifest
	evs    []*recordpb.Event
	ids    []int64
	source []OldEvent
	md     string
	board  record.BoardJSON
}

func readMigrated(t *testing.T, name string) migratedArchive {
	t.Helper()
	res, dir := MigratedArchive(t, name)
	src, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
	if err != nil {
		t.Fatalf("%s: reading the source: %v", name, err)
	}
	db, err := recordsql.OpenSeparate(filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	evs, ws, err := recordsql.Events(db)
	if err != nil {
		t.Fatal(err)
	}
	m := migratedArchive{res: res, evs: evs, source: src}
	for _, w := range ws {
		m.ids = append(m.ids, w.ID)
	}
	run, err := record.OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if has, err := record.ReportHasParts(run); err != nil {
		t.Fatal(err)
	} else if has {
		if m.md, err = record.RenderRun(run); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if m.board, err = record.BoardJSONOfRun(run); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return m
}

// cuts is every edit's removed_anchors list, in event order.
func (m migratedArchive) cuts() [][]string {
	var out [][]string
	for _, e := range m.evs {
		if b, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok && len(b.GetRemovedAnchors()) > 0 {
			out = append(out, b.GetRemovedAnchors())
		}
	}
	return out
}

// gap is the named gap on the board.
func (m migratedArchive) gap(id string) (record.GapJSON, bool) {
	for _, g := range append(m.board.Open, m.board.Closed...) {
		if g.ID == id {
			return g, true
		}
	}
	return record.GapJSON{}, false
}

// A CLAIM RETIRE FOLDS INTO ITS CUT. Every archived claim retire is classified from the source
// against the migrated record: one carrying no anchor is a lost act, listed in lost_acts (R19: every
// dropped act is in the manifest's lost_acts); one every token of
// which is in the removed_anchors of the edit that bared it is folded, and written as no event; one
// taking anchors off a kept sentence is the anchor-only retire it was in effect; and one taking an
// open gap's own marker is a lost act (F14). The split is the plan's, archive by archive, and the gaps
// whose marker a fold took keep the fates their replayed acts give them. A retire's anchors are read
// as the migrated record spells them, less the findings', which leave every list they stood in.
// anchorlessBy is each archive's anchorless claim retires, each listed in its lost_acts: 99 in all —
// m32's 377 and m35's 245 took only a finding's anchor, so they carry none.
var anchorlessBy = map[string]int{
	"2026-09-02_quadratic-formula.tar.gz":            1,
	"2026-09-10_is-91-prime-b3.tar.gz":               1,
	"2026-09-11_is-91-prime-b9.tar.gz":               2,
	"2026-10-09_is-91-prime-wave-a.tar.gz":           6,
	"2026-10-10_is-91-prime-m19.tar.gz":              2,
	"2026-10-12_is-91-prime-wave-b.tar.gz":           6,
	"2026-10-13_is-91-prime-m23.tar.gz":              1,
	"2026-10-17_is-91-prime-m26.tar.gz":              12,
	"2026-10-17_is-91-prime-m27.tar.gz":              2,
	"2026-10-18_is-91-prime-m28.tar.gz":              11,
	"2026-10-18_is-91-prime-m29.tar.gz":              11,
	"2026-10-24_is-91-prime-m32.tar.gz":              4,
	"2026-10-27_is-91-prime-m35.tar.gz":              14,
	"2026-10-30_mersenne-31-prime-sum.tar.gz":        3,
	"2026-11-04_eiffel-tower-height.tar.gz":          14,
	"2026-11-11_eiffel-tower-height.tar.gz":          1,
	"2026-11-12_is-91-prime-1838.tar.gz":             5,
	"bundle-store/2026-11-04_is-91-prime-m38.tar.gz": 3,
}

func TestAClaimRetireFoldsIntoItsCut(t *testing.T) {
	want := map[string]map[string]int{ // archive -> kind -> claim retires
		waveARun:    {"folded": 1},
		m19Run:      {"anchor-only": 3},
		m28Run:      {"folded": 1},
		m36Run:      {"F14": 1},
		mersenneRun: {"anchor-only": 4},
		eiffelRun:   {"anchor-only": 2},
		m37Run:      {"folded": 1, "anchor-only": 1},
		run1883:     {"folded": 1},
	}
	total, anchorless := map[string]int{}, map[string]int{}
	for _, name := range archivedRuns(t) {
		m := readMigrated(t, name)
		findings := findingNamesOf(m.source)
		folded := map[string]bool{}
		for _, c := range m.cuts() {
			for _, id := range c {
				folded[id] = true
			}
		}
		var retires [][]string
		for _, e := range m.evs {
			if r, ok := recordpb.BodyAs[*recordpb.Retire](e); ok {
				retires = append(retires, r.GetAnchors())
			}
		}
		got := map[string]int{}
		for _, ev := range m.source {
			if ev.Word != "retire" || truthy(ev.Fields["claim_kept"]) {
				continue
			}
			var anchors []string
			for _, a := range ev.Lists["anchors"] {
				if findings[a] {
					continue
				}
				if to, ok := m.res.IDs[a]; ok {
					a = to
				}
				anchors = append(anchors, a)
			}
			lost := slices.ContainsFunc(m.res.LostActs, func(r migration.Refusal) bool { return r.OldID == ev.ID && r.Word == "retire" })
			kind := ""
			switch {
			case len(anchors) == 0:
				kind = "anchorless"
			case !slices.ContainsFunc(anchors, func(a string) bool { return !folded[a] }):
				kind = "folded"
			case lost:
				kind = "F14"
			case slices.ContainsFunc(retires, func(r []string) bool { return slices.Equal(r, anchors) }):
				kind = "anchor-only"
			default:
				t.Errorf("%s: claim retire %d (%v) is neither folded, kept as an anchor-only retire, nor lost", name, ev.ID, anchors)
				continue
			}
			switch listed := kind == "F14" || kind == "anchorless"; {
			case listed && !lost:
				t.Errorf("%s: claim retire %d is %s and not listed in lost_acts", name, ev.ID, kind)
			case !listed && lost:
				t.Errorf("%s: claim retire %d is %s and also listed in lost_acts", name, ev.ID, kind)
			}
			total[kind]++
			if kind != "anchorless" {
				got[kind]++
			} else {
				anchorless[name]++
			}
		}
		if n := LostByWord(m.res)["retire"]; n != anchorless[name]+got["F14"] {
			t.Errorf("%s: %d retires in lost_acts, want its %d anchorless and %d F14", name, n, anchorless[name], got["F14"])
		}
		if fmt.Sprint(got) != fmt.Sprint(want[name]) && (len(got) > 0 || len(want[name]) > 0) {
			t.Errorf("%s: anchored claim retires %v, want %v", name, got, want[name])
		}
		if c := m.cuts(); fmt.Sprint(c) != fmt.Sprint(cutsOf[name]) && (len(c) > 0 || len(cutsOf[name]) > 0) {
			t.Errorf("%s: removed_anchors %v, want %v", name, c, cutsOf[name])
		}
	}
	if w := map[string]int{"anchorless": 99, "folded": 4, "anchor-only": 10, "F14": 1}; fmt.Sprint(total) != fmt.Sprint(w) {
		t.Errorf("the claim retires classify as %v, want %v", total, w)
	}
	if fmt.Sprint(anchorless) != fmt.Sprint(anchorlessBy) {
		t.Errorf("anchorless claim retires in lost_acts, by archive: %v, want %v", anchorless, anchorlessBy)
	}

	// F18's six gaps keep the fates they migrated with before the fold: no Close is written, an open
	// one reads gone, and no act on any of them is lost but the ones P3's delivery rule already lost.
	// m28's G-3a2c77c1 is the exception, below.
	for _, c := range []struct {
		run, gap string
		open     bool
		class    string
		lost     []int64 // source ids of the lost acts on the gap
	}{
		{waveARun, "G-f6f63173", true, "", []int64{328}},
		{m37Run, "G-20128984", true, "", nil},
		{m37Run, "G-477feea1", false, "repaired", nil},
		{run1883, "G-05dad92a", false, "repaired", nil},
		{run1883, "G-3dc8d8e2", false, "repaired", nil},
	} {
		m := readMigrated(t, c.run)
		assertGapFate(t, m, c.run, c.gap, c.open, c.class, c.lost)
	}
	// m28's G-3a2c77c1 is ended by its docket ruling (source 659), which the bench makes on the
	// parties' evidence and takes no delivery for (§V R1): the ruling comes forward, sided with the
	// author (Translation 4), and the gap reads closed not_a_defect by its docketing; only its lens's
	// close (632) is lost.
	m28 := readMigrated(t, m28Run)
	if g, ok := m28.gap("G-3a2c77c1"); !ok || g.Open || fmt.Sprint(g.Closure["disposition"]) != "not_a_defect" || g.Closure["docketing_id"] == nil {
		t.Errorf("%s: G-3a2c77c1 reads %+v; want closed not_a_defect by its docketing's ruling", m28Run, g)
	}
	var m28Lost []int64
	for _, r := range m28.res.LostActs {
		if strings.Contains(r.Err, "G-3a2c77c1") {
			m28Lost = append(m28Lost, r.OldID)
		}
	}
	if !slices.Equal(m28Lost, []int64{632}) {
		t.Errorf("%s: the lost acts on G-3a2c77c1 are source %v, want [632]", m28Run, m28Lost)
	}
	// m36's retire 442 took both of its gaps' markers while they were open; the markers stay on their
	// kept sentence, and the fates their lenses recorded stand.
	m := readMigrated(t, m36Run)
	for _, id := range []string{"G-34314cff", "G-e24f46e2"} {
		if !strings.Contains(m.md, anchor.Token(id)) {
			t.Errorf("m36: %s left the report; retire 442 is written as no event under F14", id)
		}
	}
	assertGapFate(t, m, m36Run, "G-34314cff", false, "repaired", []int64{449}) // retire 442 itself, source id 449
	assertGapFate(t, m, m36Run, "G-e24f46e2", false, "moot", nil)
}

// assertGapFate holds a gap to its fate: open and reading gone, or closed with class, and the source
// ids of the acts on it that are lost.
func assertGapFate(t *testing.T, m migratedArchive, run, id string, open bool, class string, lost []int64) {
	t.Helper()
	g, ok := m.gap(id)
	if !ok {
		t.Errorf("%s: %s is not on the board", run, id)
		return
	}
	switch {
	case open && (!g.Open || g.LocationState != record.LocationGone):
		t.Errorf("%s: %s reads open=%v, location %q; want open and gone", run, id, g.Open, g.LocationState)
	case !open && (g.Open || !strings.EqualFold(fmt.Sprint(g.Closure["closure_class"]), class)):
		t.Errorf("%s: %s reads open=%v, closure %v; want closed %s", run, id, g.Open, g.Closure, class)
	}
	var got []int64
	for _, r := range m.res.LostActs {
		if strings.Contains(r.Err, id) {
			got = append(got, r.OldID)
		}
	}
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	if !slices.Equal(got, lost) {
		t.Errorf("%s: the lost acts on %s are source %v, want %v", run, id, got, lost)
	}
}

// NO MIGRATED ARCHIVE HOLDS A BARE ANCHOR. No live write leaves a token on no words, and migrate
// refuses a record whose final report holds one: every committed archive's final render reads no
// bare token. Each cut no retire followed takes its gap's marker in the order its new holds it (P12b
// translation 4), and that gap reads open and gone with no Close written. No migrated edit names a
// token in both reopened and removed_anchors. Every token a baring edit or a claim retire touched
// takes the treatment its row names.
func TestNoMigratedArchiveHoldsABareAnchor(t *testing.T) {
	for _, name := range archivedRuns(t) {
		m := readMigrated(t, name)
		if bare := claimcount.BareAnchorIDs(m.md); len(bare) > 0 {
			t.Errorf("%s: the migrated report holds %v on no words", name, bare)
		}
		for i, e := range m.evs {
			b, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
			if !ok {
				continue
			}
			for _, id := range b.GetRemovedAnchors() {
				if slices.Contains(b.GetReopened(), id) {
					t.Errorf("%s: edit %d names %s in both reopened and removed_anchors", name, m.ids[i], id)
				}
				if strings.Contains(m.md, anchor.Token(id)) {
					t.Errorf("%s: edit %d took %s, and the final report holds it", name, m.ids[i], id)
				}
			}
		}
		if c := m.cuts(); fmt.Sprint(c) != fmt.Sprint(cutsOf[name]) && (len(c) > 0 || len(cutsOf[name]) > 0) {
			t.Errorf("%s: removed_anchors %v, want %v", name, c, cutsOf[name])
		}
	}
	// P12b translation 4's gaps stay open, reading gone, and no act on them is lost but the ones P3's
	// delivery rule already lost (b6's and b9's lens closes; source ids). b3's docket ruling (173) is
	// the bench's, which takes no delivery (§V R1), so it comes forward.
	for _, c := range []struct {
		run, gap string
		lost     []int64
	}{
		{b3Run, "G-19b1727b", nil}, {b6Run, "G-68c6ecf7", []int64{173}}, {b9Run, "G-8bd21b5d", []int64{136}},
	} {
		assertGapFate(t, readMigrated(t, c.run), c.run, c.gap, true, "", c.lost)
	}
	// The tokens no cut takes beside a finding's: eiffel 11-04's claim retire 394 is the anchor-only
	// retire of its gap's marker alone (its finding anchor is never placed), and m26's re-seated gap
	// markers stand in the report.
	eiffel := readMigrated(t, eiffelRun)
	if !slices.ContainsFunc(eiffel.evs, func(e *recordpb.Event) bool {
		r, ok := recordpb.BodyAs[*recordpb.Retire](e)
		return ok && slices.Equal(r.GetAnchors(), []string{"G-8dbfce13"})
	}) {
		t.Error("eiffel 11-04: no anchor-only retire carries G-8dbfce13 alone")
	}
	m26 := readMigrated(t, m26Run)
	for _, id := range []string{"G-59e61f8c", "G-0f2d72fd"} {
		if !strings.Contains(m26.md, anchor.Token(id)) {
			t.Errorf("m26: %s left the report; P12b translation 6 re-seats it", id)
		}
	}
}

// A REWRITE KEEPS ITS TOKEN IN ITS SENTENCE (P12b translation 6, F20). m26's rewrites left four tokens
// after their sentences' full stops — two gaps' markers and two citations', the seven finding anchors
// beside them never placed — and no retire took them: each is re-seated before the full stop of the
// sentence it trails, counted once as `reseated`, and no other archive counts one. Both gaps keep
// their fates, and nothing removes any of the four.
func TestARewriteKeepsItsTokenInItsSentence(t *testing.T) {
	for _, name := range archivedRuns(t) {
		res, _ := MigratedArchive(t, name)
		want := 0
		if name == m26Run {
			want = 4
		}
		if got := res.GapAnchors.Shapes["reseated"]; got != want {
			t.Errorf("%s: %d tokens re-seated, want %d", name, got, want)
		}
	}
	m := readMigrated(t, m26Run)
	tok := func(ids ...string) string {
		var b strings.Builder
		for _, id := range ids {
			b.WriteString(anchor.Token(id))
		}
		return b.String()
	}
	for _, c := range []struct{ sentence, run string }{
		{"This is a limit on the claim, not a finding that no such source exists", tok("G-59e61f8c")},
		{"does not prove that none exists", tok("G-0f2d72fd")},
		{"its data block runs to 271 without listing 91", tok("C-8a9958e3")},
		{"so no single base settles primality", tok("C-8c470447")},
	} {
		if !strings.Contains(m.md, c.sentence+c.run+".") {
			t.Errorf("m26: %q does not end %s before its full stop", c.sentence, c.run)
		}
	}
	// G-59e61f8c's one supersede link (source mint 266, G-4a4093c0 naming it) moved its successor's
	// location, so Translation 7 drops it: G-4a4093c0 stands as its own gap, and G-59e61f8c, its
	// lens's close (269) a lost act, stays open on its re-seated marker.
	if l, want := m.res.SupersedeLinks, []migration.SupersedeLink{{OldID: 266, Successor: "G-4a4093c0", Ancestor: "G-59e61f8c",
		Outcome: linkMovedField, Moved: []string{"location"}}}; fmt.Sprint(l) != fmt.Sprint(want) {
		t.Errorf("m26: supersede links %+v, want %+v", l, want)
	}
	for _, id := range []string{"G-59e61f8c", "G-4a4093c0"} {
		if g, ok := m.gap(id); !ok || !g.Open || g.LocationState != record.LocationMarked {
			t.Errorf("m26: %s reads %+v; want open and marked", id, g)
		}
	}
	if g, ok := m.gap("G-0f2d72fd"); !ok || !g.Open {
		t.Errorf("m26: G-0f2d72fd reads %+v; want open", g)
	}
	if c := m.cuts(); len(c) > 0 {
		t.Errorf("m26: removed_anchors %v; P12b translation 6 removes nothing", c)
	}
	// AN EDIT TAKING A STRAY STOP IS SETTLED, NOT LOST. Each run m26 re-seated was followed by the stray
	// full stop its sentence had left after the run. Source edits 219 and 222 each took one of those
	// stops out: 219's quote no longer locates, and 222's locates on the re-seated sentence's own full
	// stop, after the anchor run it abuts. Both are settled (findingsfold_test reads them), so no
	// archive lists an edit in lost_acts.
	for _, name := range archivedRuns(t) {
		res, _ := MigratedArchive(t, name)
		var got []int64
		for _, a := range res.LostActs {
			if a.Word == "blue_edit" {
				got = append(got, a.OldID)
			}
		}
		if len(got) > 0 {
			t.Errorf("%s: edits in lost_acts %v, want none", name, got)
		}
	}
	for _, end := range []string{
		"This is a limit on the claim, not a finding that no such source exists" + tok("G-59e61f8c") + ".\n",
		"does not prove that none exists" + tok("G-0f2d72fd") + ".\n",
	} {
		if strings.Count(m.md, end) != 1 {
			t.Errorf("m26: the report does not hold %q once, its sentence ending on one full stop", end)
		}
	}
}
