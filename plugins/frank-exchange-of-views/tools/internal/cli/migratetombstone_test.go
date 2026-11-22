package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// tombstoneLines is each Tombstone row as a migration must carry it: the gap, the kind, the docketing a
// ruling row names, whether it ends, and the type of the act it follows.
func tombstoneLines(t *testing.T, runDir string) []string {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	prev := ""
	for _, e := range m.Events {
		if tb, ok := recordpb.BodyAs[*recordpb.Tombstone](e); ok {
			out = append(out, fmt.Sprintf("%s %s docketing=%q ends=%v after %s", tb.GetGapId(), recordpb.Word(tb.GetKind()), tb.GetDocketingId(), tb.GetEnds(), prev))
			continue
		}
		typ, _ := recordpb.Body(e)
		prev = fmt.Sprintf("%T", typ)
	}
	return out
}

// standingTombstones is each gap's standing tombstones as the board reads them from
// tombstone_standing: the ruling's docketing and word, and whether an estoppel stands.
func standingTombstones(t *testing.T, runDir string) []string {
	t.Helper()
	bj, err := record.BoardJSONOfRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, g := range append(bj.Open, bj.Closed...) {
		if r := g.RulingTombstone; r != nil {
			out = append(out, fmt.Sprintf("%s ruling %s %s on=%q", g.ID, r.Docketing, r.Word, r.On))
		}
		if g.EstoppelTombstone != nil {
			out = append(out, g.ID+" estoppel")
		}
	}
	slices.Sort(out)
	return out
}

// eventRefs is every events.id a pin, a stale read point or a report read's head names, as the key
// and type of the event it names: a reference that moved to a Tombstone row, or to another act,
// reads differently here.
func eventRefs(t *testing.T, runDir string) []string {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	// An event is named by its type and its place among the events of that type: a migration re-keys
	// events and renumbers them, and keeps both of these.
	byID, nth, seen := map[int64]string{}, map[*record.Event]string{}, map[string]int{}
	for _, e := range m.Events {
		typ, _ := recordpb.Body(e)
		w := fmt.Sprintf("%T", typ)
		seen[w]++
		byID[m.At.Of(e).ID] = fmt.Sprintf("%s #%d", w, seen[w])
		nth[e] = byID[m.At.Of(e).ID]
	}
	name := func(id int64) string {
		if n, ok := byID[id]; ok {
			return n
		}
		return fmt.Sprintf("no event %d", id)
	}
	var out []string
	for _, e := range m.Events {
		switch b, _ := recordpb.Body(e); b := b.(type) {
		case *recordpb.Dispatch:
			if b.Pin != nil {
				out = append(out, nth[e]+" pins "+name(b.GetPin()))
			}
		case *recordpb.Stale:
			out = append(out, nth[e]+" reads through "+name(b.GetThrough()))
		case *recordpb.ReportRead:
			out = append(out, nth[e]+" was served "+name(b.GetHead()))
		}
	}
	return out
}

// T-B21. A MIGRATED RECORD WRITES EACH TOMBSTONE ONCE (§III.B12, §III.B7). A record built through the
// verbs at this epoch, holding the rows each placing and ending act writes — a ruling for blue and a
// ruling for red at impasse, blue's verbatim application of a gap's fix_new, its reword ending that
// estoppel and its re-application — with a dispatch plan pinned at the head and a lens's report read
// each after a placing act, migrates through feov-migrate with zero refusals: the rows come forward
// in order, each after the same act, none written twice; the manifest counts as many in as out;
// the standing tombstones are the source's; and every pin, read point and served head names the act
// it named, never a row.
func TestAMigratedRecordWritesEachTombstoneOnce(t *testing.T) {
	for _, c := range []struct {
		name  string
		build func(t *testing.T) string
	}{
		{"a ruling for blue", func(t *testing.T) string {
			runDir := newRun(t)
			quoteImpasseBoard(t, runDir)
			ruleAtImpasse(t, runDir, false)
			afterPlacing(t, runDir)
			return runDir
		}},
		{"a ruling for red", func(t *testing.T) string {
			runDir := newRun(t)
			quoteImpasseBoard(t, runDir)
			ruleForRedAtImpasse(t, runDir)
			afterPlacing(t, runDir)
			return runDir
		}},
		{"an estoppel placed, ended and placed again", func(t *testing.T) string {
			runDir := newRun(t)
			if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
				&recordpb.Cast{SeatIds: []string{lensSeat, otherLens, "red-chair", "blue-synthesize", blueSeat, "judge"}}); err != nil {
				t.Fatalf("the cast: %v", err)
			}
			gap := seedProposalApplied(t, runDir)
			afterPlacing(t, runDir)
			const quote = "Five verification approaches agree, all sharing one definition of primality"
			tok := "<!--gap:" + gap + "-->"
			asSeat(t, runDir, blueSeat, "edit", "--key", "W1", "--quote", quote+tok+".", "--new", quote+", on every input tried"+tok+".", "--reason", "r")
			asSeat(t, runDir, blueSeat, "edit", "--key", "W2", "--quote", quote+", on every input tried"+tok+".", "--new", "Five independent verification approaches agree"+tok+".", "--reason", "r")
			asSeat(t, runDir, blueSeat, "edit", "--key", "A2", "--answers", gap, "--accept", "--reason", "red's wording after all")
			afterPlacing(t, runDir)
			return runDir
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			src := c.build(t)
			rows := tombstoneLines(t, src)
			if len(rows) == 0 {
				t.Fatal("the source holds no tombstone row")
			}
			dst := recordtest.TmpRun(t)
			out, merr := runMigrate(t, "--from", src, "--to", dst)
			raw, err := os.ReadFile(filepath.Join(dst, "inputs", migration.ManifestName))
			if err != nil {
				t.Fatal(err)
			}
			var man migration.Manifest
			if err := json.Unmarshal(raw, &man); err != nil {
				t.Fatal(err)
			}
			if merr != nil || len(man.Refusals) != 0 {
				first := migration.Refusal{}
				if len(man.Refusals) > 0 {
					first = man.Refusals[0]
				}
				t.Fatalf("migrate: %v; %d refusal(s), the first: %+v\n%s", merr, len(man.Refusals), first, out)
			}
			if man.In["tombstone"] != len(rows) || man.Out["tombstone"] != len(rows) {
				t.Errorf("the manifest counts %d tombstone(s) in and %d out, want %d each", man.In["tombstone"], man.Out["tombstone"], len(rows))
			}
			if got := tombstoneLines(t, dst); !slices.Equal(got, rows) {
				t.Errorf("the migrated rows, in order:\n%v\nwant the source's:\n%v", got, rows)
			}
			if got, want := standingTombstones(t, dst), standingTombstones(t, src); !slices.Equal(got, want) {
				t.Errorf("the standing tombstones after %v, before %v", got, want)
			}
			if got, want := eventRefs(t, dst), eventRefs(t, src); !slices.Equal(got, want) {
				t.Errorf("the references after:\n%v\nbefore:\n%v", got, want)
			}
		})
	}
}

// afterPlacing puts the references a row's id shift could move right after a placing act: the chair's
// plan and its dispatches pinned at the head, and a lens's read of the report. A stale refusal's read
// point is held the same way by eventRefs where a record holds one; these sources hold none.
func afterPlacing(t *testing.T, runDir string) {
	t.Helper()
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	asSeat(t, runDir, otherLens, "register")
	asSeat(t, runDir, otherLens, "show", "report")
}
