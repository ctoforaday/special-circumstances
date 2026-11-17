package migrate

import (
	"fmt"
	"maps"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// findingNamesOf is every name an archived finding carries in the archive — its id, its archived id
// and its label — in both letter cases of the one-shape id, which a reference written before or
// after the source respelled it may carry.
func findingNamesOf(evs []OldEvent) map[string]bool {
	out := map[string]bool{}
	for _, n := range findingNames(evs) {
		out[n] = true
		if wholeFindingID.MatchString(strings.ToUpper(n[:1]) + n[1:]) {
			out[strings.ToUpper(n[:1])+n[1:]], out[strings.ToLower(n[:1])+n[1:]] = true, true
		}
	}
	return out
}

// onlyFindings reports whether ids is non-empty and every element names a finding (or is skip).
func onlyFindings(ids []string, names map[string]bool, skip string) bool {
	named := false
	for _, id := range ids {
		switch {
		case id == skip && skip != "":
		case names[id]:
			named = true
		default:
			return false
		}
	}
	return named
}

// findingTokenOld is an archived finding token: its tag and an id in either era's shape. A comment
// that only shows the form (`<!--fx:...-->`) is a seat's prose, not a token.
var findingTokenOld = regexp.MustCompile(`<!--fx:[fF]-[0-9a-f]+-->`)

// proseFindingID is a finding id in the one shape, in either letter case, at word boundaries.
var proseFindingID = regexp.MustCompile(`\b[fF]-[0-9a-f]{8}\b`)

// reportText is each archived column that holds report text: a part's, and an edit's old and new.
var reportText = map[string][]string{"report_part": {"text"}, "base_ingest": {"text"}, "blue_edit": {"old", "new"}}

// A LENS FINDING LEAVES AN ARCHIVED RUN (plans/findings-fold.md §III.A5, T-A5). A finding is written as
// no event, the Anchor that placed its token is written as none, its id leaves every list it stood in
// and its token every field. What it leaves behind is the residue, counted here over every archive:
// the findings, their Anchors, the gaps' found_by links, the tokens in report text, and the acts whose
// only anchors were findings — a repaired close, a blue edit's reopened list, a claim retire — which
// come forward with none.
//
// The plan's census (§II.4: 600 Anchors, 392 found_by links, 558 tokens) read the sqlite archives'
// tables and missed the six JSONL-era archives of 2026-08-22/23, which hold 114 finding Anchors, 87
// found_by links and 82 tokens in report text; the counts here are over all 36, as migrate reads them.
func TestALensFindingLeavesAnArchivedRun(t *testing.T) {
	var findings, anchors, foundBy, tokens, repairedBare, reopenedBare, retireBare, runs int
	stated := map[string]string{} // archive -> the acts written as no event because a finding's absence settled them
	for _, name := range archivedRuns(t) {
		evs := sourceEvents(t, name)
		names := findingNamesOf(evs)
		held := false
		for _, old := range evs {
			for _, col := range reportText[old.Word] {
				if v, ok := old.Fields[col].(string); ok {
					tokens += len(findingTokenOld.FindAllString(v, -1))
				}
			}
			switch old.Word {
			case "finding":
				findings++
				held = true
			case "anchor":
				if id, _ := old.Fields["id"].(string); names[id] {
					anchors++
				}
			case "mint":
				foundBy += len(old.Lists["found_by"])
			case "close":
				gap, _ := old.Fields["gap_id"].(string)
				if cls, _ := old.Fields["closure_class"].(string); cls == "repaired" && onlyFindings(old.Lists["anchors"], names, gap) {
					repairedBare++
				}
			case "blue_edit":
				if onlyFindings(old.Lists["reopened"], names, "") {
					reopenedBare++
				}
			case "retire":
				if claim, _ := old.Fields["claim"].(string); claim != "" && onlyFindings(old.Lists["anchors"], names, "") {
					retireBare++
				}
			}
		}
		if held {
			runs++
		}
		man, _ := MigratedArchive(t, name)
		if len(man.Refusals) != 0 {
			t.Errorf("%s: %d refusal(s); the first: %+v", name, len(man.Refusals), man.Refusals[0])
		}
		if man.Out["finding"] != 0 {
			t.Errorf("%s: the migrated record holds %d finding(s)", name, man.Out["finding"])
		}
		if len(man.QuotelessReadings)+len(man.SettledEdits) != 0 {
			stated[name] = fmt.Sprintf("quoteless readings %v, settled edits %v", man.QuotelessReadings, man.SettledEdits)
		}
	}
	// A reading whose claim was only finding tokens, and an edit of a stray a finding's token left,
	// are written as no event; these are every one, by source id.
	if want := map[string]string{
		"2026-09-02_quadratic-formula.tar.gz": "quoteless readings [758], settled edits []",
		"2026-10-17_is-91-prime-m26.tar.gz":   "quoteless readings [], settled edits [219 222]",
	}; !maps.Equal(stated, want) {
		t.Errorf("the acts a finding's absence settled, by archive:\n%v\nwant:\n%v", stated, want)
	}
	got := [8]int{runs, findings, anchors, foundBy, tokens, repairedBare, reopenedBare, retireBare}
	if want := [8]int{36, 1037, 714, 479, 640, 26, 138, 2}; got != want {
		t.Errorf("the residue over the archives: runs, findings, finding Anchors, found_by links, tokens in report text, "+
			"repaired closes left bare, edits left with an empty reopened, claim retires left bare = %v, want %v", got, want)
	}
}

// NO MIGRATED ARCHIVE HOLDS A FINDING ID OR TOKEN (T-A5): no id field — every field respellIDs walks
// as an id, Close.anchors, BlueEdit.reopened, Retire.anchors, SpotCheck.ids and Anchor.id among
// them — holds a name an archived finding carried, and no field holds a finding's token. A seat's
// prose keeps the id or label it wrote, which names nothing in the migrated record: the one-shape ids
// in it are counted, not asserted absent.
func TestNoMigratedArchiveHoldsAFindingIDOrToken(t *testing.T) {
	prose := 0
	for _, name := range archivedRuns(t) {
		names := findingNamesOf(sourceEvents(t, name))
		bodies, _, _ := migratedBodies(t, name)
		for _, b := range bodies {
			eachString(b.(proto.Message).ProtoReflect(), func(field, v string) {
				switch {
				case idFields[field] && names[v]:
					t.Errorf("%s: %s holds the finding id %q", name, field, v)
				case findingTokenOld.MatchString(v):
					t.Errorf("%s: %s holds a finding token: %q", name, field, findingTokenOld.FindString(v))
				case argumentFields[field]:
					prose += len(proseFindingID.FindAllString(v, -1))
				}
			})
		}
	}
	// 444: the two in grade rulings' opinions (sleeper-service-plan, b9) leave with the rulings,
	// which translate to no event (EventSchema 79).
	if prose != 444 {
		t.Errorf("%d finding ids stand in the migrated records' prose fields, want 444", prose)
	}
}

// docketText is a docket's or a ruling's text as the fold compares it: every anchor token out, every
// id in any era's shape one placeholder, whitespace runs one space. What remains is what a seat
// wrote, which the fold leaves as it was.
func docketText(v string) string {
	v = docketToken.ReplaceAllString(v, "")
	v = anyIDShape.ReplaceAllString(v, "ID")
	return strings.Join(strings.Fields(v), " ")
}

var (
	docketToken = regexp.MustCompile(`<!--[^>]*-->`)
	anyIDShape  = regexp.MustCompile(`\b(?:[A-Za-z]{1,2}-[0-9a-f]{4,}|[GQMRP]\d+(?:-\d+)?|[A-Za-z0-9]+-F\d+)\b`)
)

// stringsOf is every string value of an archived event — its columns and its arms' — of at least
// twenty characters holding a space, which leaves out ids, enumerated words and dates.
func stringsOf(old OldEvent) []string {
	var out []string
	add := func(v any) {
		if s, ok := v.(string); ok && len(s) >= 20 && strings.Contains(s, " ") {
			out = append(out, s)
		}
	}
	for _, v := range old.Fields {
		add(v)
	}
	for _, arm := range old.Arms {
		for _, v := range arm {
			add(v)
		}
	}
	return out
}

// AN ARCHIVED DOCKET IS UNCHANGED BY THE FOLD (T-A6) but for the finding tokens its text held. Every
// archived docket motion and every ruling on one that comes forward holds every text the archive did,
// tokens and ids aside, and no docket's grounds named a finding, so nothing of a docket leaves with
// the findings. The 39 archived dockets and their 38 rulings are counted, and so are the 4 dockets
// and 4 rulings left out as acts on a gap a mint had superseded (actOnSupersededGap), which the fold
// does not touch, and the finding tokens the dockets' and rulings' text held.
func TestAnArchivedDocketIsUnchanged(t *testing.T) {
	var dockets, rulings, leftOut, tokens int
	for _, name := range archivedRuns(t) {
		evs := sourceEvents(t, name)
		names := findingNamesOf(evs)
		man, _ := MigratedArchive(t, name)
		bodies, _, _ := migratedBodies(t, name)
		migrated := map[string][]string{} // migrated motion id -> the texts of the motion and its rulings
		for _, b := range bodies {
			var id string
			switch x := b.(type) {
			case *recordpb.Motion:
				id = x.GetMotionId()
			case *recordpb.MotionRule:
				id = x.GetMotionId()
			default:
				continue
			}
			eachString(b.(proto.Message).ProtoReflect(), func(field, v string) {
				if findingTokenOld.MatchString(v) {
					t.Errorf("%s: %s of motion %s holds a finding token", name, field, id)
				}
				migrated[id] = append(migrated[id], docketText(v))
			})
		}
		docket := map[string]bool{} // archived motion ids filed on the docket
		for _, old := range evs {
			if old.Word == "motion" && old.Arms["docket"] != nil {
				id, _ := old.Fields["motion_id"].(string)
				docket[id] = true
			}
		}
		for _, old := range evs {
			id, _ := old.Fields["motion_id"].(string)
			if !docket[id] || (old.Word != "motion" && (old.Word != "motion_rule" || old.Arms["docket"] == nil)) {
				continue
			}
			if old.Word == "motion" {
				dockets++
				if g, _ := old.Arms["docket"]["grounds"].(string); names[g] {
					t.Errorf("%s: docket motion %s names finding %s in its grounds", name, id, g)
				}
			} else {
				rulings++
			}
			for _, v := range stringsOf(old) {
				tokens += len(findingTokenOld.FindAllString(v, -1))
			}
			to := man.IDs[id]
			if to == "" {
				to = id
			}
			held, ok := migrated[to]
			if !ok {
				leftOut++
				continue
			}
			for _, v := range stringsOf(old) {
				want := docketText(v)
				found := false
				for _, h := range held {
					found = found || h == want
				}
				if !found {
					t.Errorf("%s: archived %s %d on motion %s held text the migrated act does not:\n%.300s", name, old.Word, old.ID, id, want)
				}
			}
		}
	}
	if got, want := [4]int{dockets, rulings, leftOut, tokens}, [4]int{39, 38, 8, 5}; got != want {
		t.Errorf("dockets, rulings, left out as acts on a superseded gap, finding tokens = %v, want %v", got, want)
	}
}

// A STALE REFUSAL THAT READ THROUGH A FINDING READS THROUGH THE EVENT BEFORE IT (§III.A5, pulled
// forward into A1). Its read point named an act the migrated record does not hold; it comes forward
// naming the last event written before that act, so every act after the finding is news to the seat
// as it was, and the finding, which is gone, is none. Asserted on every such read point in the
// archives: the migrated read point is the event the source wrote just before the finding — same
// clock, same word — and their count is the 15 the fold translates.
func TestAStaleReadPointAtAFindingIsTheEventBeforeIt(t *testing.T) {
	translated := 0
	for _, name := range archivedRuns(t) {
		evs := sourceEvents(t, name)
		names := findingNamesOf(evs)
		dropped := map[int64]bool{} // source ids written as no event by the fold: findings and their Anchors
		for _, old := range evs {
			id, _ := old.Fields["id"].(string)
			if old.Word == "finding" || (old.Word == "anchor" && names[id]) {
				dropped[old.ID] = true
			}
		}
		var want []OldEvent // per source stale refusal, in order: the event its migrated read point names, or a zero OldEvent
		named := false
		for _, old := range evs {
			if old.Word != "stale" {
				continue
			}
			through, _ := old.Fields["through"].(int64)
			if !dropped[through] {
				want = append(want, OldEvent{})
				continue
			}
			var before OldEvent
			for _, e := range evs {
				if e.ID < through && !dropped[e.ID] && e.ID > before.ID {
					before = e
				}
			}
			want = append(want, before)
			named = true
		}
		if !named {
			continue
		}
		m, _, _, _ := migratedEvents(t, name)
		ids := m.At.IDs(m.Events)
		at := map[int64]int{}
		for i := range m.Events {
			at[ids[i]] = i
		}
		n := 0
		for _, e := range m.Events {
			st := e.GetStale()
			if st == nil {
				continue
			}
			if n >= len(want) {
				t.Fatalf("%s: the migrated record holds more stale refusals than the source's %d", name, len(want))
			}
			w := want[n]
			n++
			if w.ID == 0 {
				continue
			}
			translated++
			got := m.Events[at[st.GetThrough()]]
			wantTS, err := record.ParseStamp(w.TS)
			if err != nil {
				t.Fatal(err)
			}
			gotTS, err := record.ParseStamp(got.GetTs())
			if err != nil {
				t.Fatal(err)
			}
			if !gotTS.Equal(wantTS) {
				t.Errorf("%s: a stale refusal that read through a finding reads through %s at %s; the source wrote %s %d just before the finding, at %s",
					name, recordpb.Word(got.GetType()), got.GetTs(), w.Word, w.ID, w.TS)
			}
		}
	}
	if translated != 15 {
		t.Errorf("%d stale read points named a finding or its Anchor, want 15", translated)
	}
}
