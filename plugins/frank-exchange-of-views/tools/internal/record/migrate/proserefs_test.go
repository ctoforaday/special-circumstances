package migrate_test

import (
	"database/sql"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN ARCHIVED RECORD'S PROSE IDS BECOME FIELDS, ONCE, KEEPING EACH THE RECORD HOLDS. Before the
// fields, a declaration's gaps were read out of its holding and a closure's anchors out of its
// --verified-with words at every read; migrate reads them from that text once. Every migrated
// declaration's gap_ids, and every closure's anchors, are exactly the ids its text names that the
// migrated record holds — and the archive holds at least one of each, or nothing was checked.
func TestArchivedProseIDsBecomeFields(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	sort.Strings(tarballs)
	gapRe := regexp.MustCompile(`\b` + anchor.IDPattern("gap") + `\b`)
	anyRe := regexp.MustCompile(`\b` + anchor.IDPattern() + `\b`)
	var declared, anchored, unheld int
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		_, run := migrateArchive(t, name)
		fam, err := record.FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		heldIDs := map[string]bool{}
		for _, e := range fam.Events {
			if id := bodyID(e); id != "" {
				heldIDs[id] = true
			}
		}
		want := func(text string, re *regexp.Regexp) []string {
			var out []string
			for _, id := range re.FindAllString(text, -1) {
				if slices.Contains(out, id) {
					continue
				}
				if !heldIDs[id] {
					unheld++
					continue
				}
				out = append(out, id)
			}
			return out
		}
		for _, e := range fam.Events {
			if d, ok := recordpb.BodyAs[*recordpb.Declare](e); ok {
				if w := want(d.GetHolding(), gapRe); !slices.Equal(d.GetGapIds(), w) {
					t.Errorf("%s: a declaration's gap_ids are %v, want %v — the held gaps its holding names", name, d.GetGapIds(), w)
				}
				if len(d.GetGapIds()) > 0 {
					declared++
				}
			}
			if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok {
				if w := want(c.GetAnchorTool(), anyRe); !slices.Equal(c.GetAnchors(), w) {
					t.Errorf("%s: a closure's anchors are %v, want %v — the held anchors its --verified-with names", name, c.GetAnchors(), w)
				}
				if len(c.GetAnchors()) > 0 {
					anchored++
				}
			}
		}
	}
	t.Logf("across the archive: %d declaration(s) about gaps, %d closure(s) naming anchors, %d id(s) named in prose and held by no act", declared, anchored, unheld)
	if declared == 0 || anchored == 0 {
		t.Errorf("the archive yields %d declaration(s) with gaps and %d closure(s) with anchors; with none of either, the translation was not exercised", declared, anchored)
	}
}

// bodyID is the anchor id an act mints, or "".
func bodyID(e *recordpb.Event) string {
	if b, ok := recordpb.BodyAs[*recordpb.Finding](e); ok {
		return b.GetId()
	}
	if b, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
		return b.GetGapId()
	}
	if b, ok := recordpb.BodyAs[*recordpb.Cite](e); ok {
		return b.GetLabel()
	}
	if b, ok := recordpb.BodyAs[*recordpb.Proof](e); ok {
		return b.GetProofId()
	}
	if b, ok := recordpb.BodyAs[*recordpb.Verify](e); ok {
		return b.GetLabel()
	}
	return ""
}

// A SOURCE THAT HAS THE FIELD STATES IT, NONE INCLUDED. Re-migrating a record whose schema holds
// gap_ids and anchors carries each as written: an empty list there is "names none", and the prose is
// not read again for an id — if it were, a declaration the bench made about no gap would gain one at
// every later migration.
func TestAStatedListIsCarriedAndProseIsNotReadAgain(t *testing.T) {
	_, run := migrateArchive(t, lastWithBoth(t))
	db, err := sql.Open("sqlite", "file:"+filepath.Join(run.Dir(), "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"declare_gap_ids", "close_anchors"} {
		if _, err := db.Exec(`DELETE FROM "` + table + `"`); err != nil {
			t.Fatalf("emptying %s: %v", table, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	to := filepath.Join(recordtest.TmpRun(t), "run")
	m, err := migrate.Migrate(run.Dir(), to, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Refusals) != 0 {
		t.Fatalf("re-migrating refused %d event(s); the first: %+v", len(m.Refusals), m.Refusals[0])
	}
	dst := runtest.Open(t, to)
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range fam.Events {
		if d, ok := recordpb.BodyAs[*recordpb.Declare](e); ok && len(d.GetGapIds()) > 0 {
			t.Errorf("a declaration stated about no gap was given %v from its prose", d.GetGapIds())
		}
		if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok && len(c.GetAnchors()) > 0 {
			t.Errorf("a closure stated to name no anchor was given %v from its prose", c.GetAnchors())
		}
	}
}

// lastWithBoth is the newest archived run holding a declaration whose holding names a gap and a
// closure whose --verified-with names an anchor.
func lastWithBoth(t *testing.T) string {
	t.Helper()
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, _ := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	sort.Sort(sort.Reverse(sort.StringSlice(tarballs)))
	for _, tb := range tarballs {
		_, run := migrateArchive(t, filepath.Base(tb))
		fam, err := record.FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		var d, c bool
		for _, e := range fam.Events {
			if x, ok := recordpb.BodyAs[*recordpb.Declare](e); ok && len(x.GetGapIds()) > 0 {
				d = true
			}
			if x, ok := recordpb.BodyAs[*recordpb.Close](e); ok && len(x.GetAnchors()) > 0 {
				c = true
			}
		}
		if d && c {
			return filepath.Base(tb)
		}
	}
	t.Fatal("no archived run holds both a declaration about a gap and a closure naming an anchor")
	return ""
}
