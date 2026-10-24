package migrate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// sourceOutcomes is every outcome event of the named archive's source record, in stream order, read
// through the adapter migrate itself picks for it.
func sourceOutcomes(t *testing.T, name string) []migrate.OldEvent {
	t.Helper()
	recs := filepath.Join(recordtest.RequireArchive(t, name), "records")
	var src migrate.Source
	if _, err := os.Stat(filepath.Join(recs, "record.db")); err == nil {
		sq, err := migrate.OpenSQLite(recs, t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = sq.Close() })
		src = sq
	} else {
		js, err := migrate.OpenJSONL(recs)
		if err != nil {
			t.Fatal(err)
		}
		src = js
	}
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	var out []migrate.OldEvent
	for _, e := range evs {
		if e.Word == "outcome" {
			out = append(out, e)
		}
	}
	return out
}

// sourceProse is an archived outcome's prose under whichever name its era wrote it.
func sourceProse(e migrate.OldEvent) string {
	if p, ok := e.Fields["prose"].(string); ok {
		return p
	}
	p, _ := e.Fields["reason"].(string)
	return p
}

// respelled is an archived outcome's prose with the migration's own stated rewrites of that field
// applied in the order it made them: the archived ids its prose names, respelled to the migrated
// ids. Only the written outcome's prose is rewritten, so these are its rewrites alone.
func respelled(prose string, rw []migration.ProseRewrite) string {
	var b strings.Builder
	for _, r := range rw {
		if r.Field != "Outcome.prose" {
			continue
		}
		i := strings.Index(prose, r.Old)
		if i < 0 {
			return "(a stated rewrite of " + r.Old + " has no occurrence left in the source prose)"
		}
		b.WriteString(prose[:i] + r.New)
		prose = prose[i+len(r.Old):]
	}
	return b.String() + prose
}

// A MIGRATED RECORD HOLDS ONE OUTCOME, THE SOURCE'S LATEST (outcomes.go). Each of these archives
// holds a second assembly sitting after its first outcome — a register, then an outcome with other
// prose — and the write closes a run with an outcome to seat acts, so replaying the earlier outcome
// refuses the later sitting's register and outcome. The earlier one is not written: the archive
// migrates with nothing refused, its one outcome is the latest the source stamped, verdict and
// prose, and the manifest names the outcome it dropped and the one that replaced it.
func TestASupersededOutcomeIsNotWritten(t *testing.T) {
	for _, name := range []string{
		"2026-08-23_research-loop-counterparts.tar.gz",
		"2026-08-23_sleeper-service-plan.tar.gz",
		"2026-09-02_quadratic-formula.tar.gz",
	} {
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			src := sourceOutcomes(t, name)
			if len(src) != 2 {
				t.Fatalf("the source holds %d outcome(s); this test reads an archive with two", len(src))
			}
			first, latest := src[0], src[1]
			if sourceProse(first) == sourceProse(latest) {
				t.Fatalf("the two source outcomes carry the same prose, so this test cannot tell which landed")
			}
			res, run := migrateArchive(t, name)
			if len(res.Refusals) != 0 {
				t.Errorf("%d event(s) refused; the first: %+v", len(res.Refusals), res.Refusals[0])
			}
			merged, err := record.MergedEvents(run)
			if err != nil {
				t.Fatal(err)
			}
			var got []*recordpb.Outcome
			for _, e := range merged.Events {
				if o := e.GetOutcome(); o != nil {
					got = append(got, o)
				}
			}
			if len(got) != 1 {
				t.Fatalf("the migrated record holds %d outcomes, want 1", len(got))
			}
			wantVerdict, _ := latest.Fields["verdict"].(string)
			firstVerdict, _ := first.Fields["verdict"].(string)
			if wantVerdict == "" || firstVerdict == "" {
				t.Fatalf("a source outcome states no verdict (first %q, latest %q)", firstVerdict, wantVerdict)
			}
			if v := recordpb.Word(got[0].GetVerdict()); !strings.EqualFold(v, wantVerdict) {
				t.Errorf("the migrated outcome's verdict is %q, the source's latest %q", v, wantVerdict)
			}
			if want := respelled(sourceProse(latest), res.Rewritten); got[0].GetProse() != want {
				t.Errorf("the migrated outcome's prose is not the source's latest:\n got  %q\n want %q", got[0].GetProse(), want)
			}
			want := migration.SupersededOutcome{OldID: first.ID, Seat: first.SeatID, Verdict: strings.ToLower(firstVerdict), SupersededBy: latest.ID}
			if len(res.SupersededOutcomes) != 1 || res.SupersededOutcomes[0] != want {
				t.Errorf("the manifest states superseded outcomes %+v, want [%+v]", res.SupersededOutcomes, want)
			}
		})
	}
}

// A RECORD WITH ONE OUTCOME DROPS NONE, and says so: the manifest states an empty list.
func TestASingleOutcomeIsKept(t *testing.T) {
	res, run := migrateArchive(t, "2026-10-19_is-91-prime-m30.tar.gz")
	if res.SupersededOutcomes == nil || len(res.SupersededOutcomes) != 0 {
		t.Errorf("superseded outcomes %#v, want a stated empty list", res.SupersededOutcomes)
	}
	merged, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range merged.Events {
		if e.GetOutcome() != nil {
			n++
		}
	}
	if want := len(sourceOutcomes(t, "2026-10-19_is-91-prime-m30.tar.gz")); n != want || want != 1 {
		t.Errorf("the migrated record holds %d outcome(s), the source %d; want one each", n, want)
	}
}
