package migrate_test

import (
	"database/sql"
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// archivedInquiryRun is what an archive written at EventSchema 55 holds of the chair's inquiry duties
// and of the inquiries themselves, each counted from the archive by the test before it is compared.
type archivedInquiryRun struct {
	archive                     string
	reviews, rulings, filings   int // inquiry_review rows; motion_rule rows ruling on an inquiry; motion_inquiry rows
	inquiries, acts             int // distinct inquiry ids; inquiry acts, proposals and fates
	outcomes, gates             []string
	motionRules, motionsArchive int
}

// THE LAST RUNS THAT HELD THE CHAIR'S INQUIRY DUTIES LOSE THOSE AND NOTHING ELSE. m36 and m35 were
// written at EventSchema 55: the chair reviewed the inquiries and ruled on each. Migrated, every
// review and every ruling on an inquiry is gone, every other ruling and motion arrives, every inquiry
// arrives with each of its acts under the fate it was filed with, and the run's stored outcome and
// gate verdicts are the archive's. m35's engine died before a verdict, so it has neither.
func TestTheChairsInquiryDutiesLeaveAnArchivedRun(t *testing.T) {
	for _, c := range []archivedInquiryRun{
		{archive: "2026-10-29_is-91-prime-m36.tar.gz", reviews: 1, rulings: 5, inquiries: 5, acts: 12,
			outcomes: []string{"verified"}, gates: []string{"pass"}, motionRules: 5},
		{archive: "2026-10-27_is-91-prime-m35.tar.gz", reviews: 1, rulings: 6, inquiries: 7, acts: 22,
			motionRules: 13, motionsArchive: 8},
	} {
		t.Run(c.archive, func(t *testing.T) {
			src := openRO(t, recordtest.RequireArchive(t, c.archive))
			got := archivedInquiryRun{
				archive:        c.archive,
				reviews:        countOf(t, src, `SELECT count(*) FROM "inquiry_review"`),
				rulings:        countOf(t, src, `SELECT count(*) FROM "motion_rule" WHERE "inquiry" IS NOT NULL`),
				filings:        countOf(t, src, `SELECT count(*) FROM "motion_inquiry"`),
				inquiries:      countOf(t, src, `SELECT count(DISTINCT "inquiry_id") FROM "inquiry"`),
				acts:           countOf(t, src, `SELECT count(*) FROM "inquiry"`),
				outcomes:       wordsOf(t, src, `SELECT "verdict" FROM "outcome" ORDER BY "event_id"`),
				gates:          wordsOf(t, src, `SELECT "verdict" FROM "gate" ORDER BY "event_id"`),
				motionRules:    countOf(t, src, `SELECT count(*) FROM "motion_rule"`),
				motionsArchive: countOf(t, src, `SELECT count(*) FROM "motion"`),
			}
			if !slices.Equal(got.outcomes, c.outcomes) || !slices.Equal(got.gates, c.gates) ||
				got.reviews != c.reviews || got.rulings != c.rulings || got.filings != c.filings ||
				got.inquiries != c.inquiries || got.acts != c.acts || got.motionRules != c.motionRules || got.motionsArchive != c.motionsArchive {
				t.Fatalf("the archive holds %+v, this test was written against %+v", got, c)
			}

			res, dst := migrateArchive(t, c.archive)
			if len(res.Refusals) != 0 {
				t.Fatalf("refusals: %+v", res.Refusals)
			}
			ddb := openRO(t, dst.Dir())
			if res.In["inquiry_review"] != c.reviews || res.Out["inquiry_review"] != 0 {
				t.Errorf("reviews: %d in, %d out; want %d in and none out", res.In["inquiry_review"], res.Out["inquiry_review"], c.reviews)
			}
			if n := countOf(t, ddb, `SELECT count(*) FROM "motion_rule"`); n != c.motionRules-c.rulings {
				t.Errorf("the migrated record holds %d ruling(s), want the archive's %d less its %d on an inquiry", n, c.motionRules, c.rulings)
			}
			if n := countOf(t, ddb, `SELECT count(*) FROM "motion"`); n != c.motionsArchive-c.filings {
				t.Errorf("the migrated record holds %d motion(s), want the archive's %d less its %d on an inquiry", n, c.motionsArchive, c.filings)
			}

			// Every inquiry arrives under its migrated id, with every act under the fate it was filed with.
			var archivedIDs []string
			for _, id := range wordsOf(t, src, `SELECT DISTINCT "inquiry_id" FROM "inquiry" ORDER BY "inquiry_id"`) {
				if to, ok := res.IDs[id]; ok {
					id = to
				}
				archivedIDs = append(archivedIDs, id)
			}
			slices.Sort(archivedIDs)
			if ids := wordsOf(t, ddb, `SELECT DISTINCT "inquiry_id" FROM "inquiry" ORDER BY "inquiry_id"`); !slices.Equal(ids, archivedIDs) {
				t.Errorf("the migrated inquiries are %v, the archive's under their migrated ids %v", ids, archivedIDs)
			}
			if n := countOf(t, ddb, `SELECT count(*) FROM "inquiry"`); n != c.acts {
				t.Errorf("%d archived inquiry act(s) arrived as %d", c.acts, n)
			}
			if before, after := fatesIn(t, src, "inquiry"), fatesIn(t, ddb, "inquiry"); !maps.Equal(before, after) {
				t.Errorf("inquiry acts by fate: the archive holds %v, the migrated record %v", before, after)
			}

			if o := wordsOf(t, ddb, `SELECT "verdict" FROM "outcome" ORDER BY "event_id"`); !slices.Equal(o, c.outcomes) {
				t.Errorf("the migrated outcome(s) %v, the archive's %v", o, c.outcomes)
			}
			if g := wordsOf(t, ddb, `SELECT "verdict" FROM "gate" ORDER BY "event_id"`); !slices.Equal(g, c.gates) {
				t.Errorf("the migrated gate verdict(s) %v, the archive's %v", g, c.gates)
			}
		})
	}
}

// THE CHAIR'S INQUIRY-REVIEW BLOCKER LEAVES EVERY STORED PLAN (EventSchema 57), and nothing else
// does: each migrated plan holds its archived blockers less that kind, keeps its pass_permitted, is
// re-keyed, and is one a relay can carry — no archived plan that held the kind is left with no
// blocker and no PASS permitted.
func TestTheInquiryReviewBlockerLeavesEveryStoredPlan(t *testing.T) {
	for _, name := range []string{"2026-10-29_is-91-prime-m36.tar.gz", "2026-10-27_is-91-prime-m35.tar.gz"} {
		t.Run(name, func(t *testing.T) {
			res, _ := migrateArchive(t, name)
			if len(res.Refusals) != 0 {
				t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
			}
			srcPlans, dstPlans, _, _, _ := planPairsOf(t, name)
			if len(srcPlans) != len(dstPlans) {
				t.Fatalf("%d archived plan(s), %d migrated", len(srcPlans), len(dstPlans))
			}
			held := 0
			for i := range srcPlans {
				// An archived plan may print a field the current plan does not declare, so it is read
				// for its blockers and its pass_permitted alone.
				var s struct {
					Blockers      []struct{ Kind string } `json:"blockers"`
					PassPermitted bool                    `json:"pass_permitted"`
				}
				if err := json.Unmarshal([]byte(srcPlans[i].printed), &s); err != nil {
					t.Fatalf("archived plan %d does not read: %v", i, err)
				}
				d, _, err := record.ReadPlan([]byte(dstPlans[i].printed))
				if err != nil {
					t.Fatalf("migrated plan %d does not read: %v", i, err)
				}
				want := 0
				for _, b := range s.Blockers {
					if b.Kind == "inquiry_review" {
						held++
						continue
					}
					want++
				}
				for _, b := range d.Blockers {
					if b.Kind == "inquiry_review" {
						t.Errorf("migrated plan %d still holds the chair's inquiry review: %+v", i, b)
					}
				}
				if len(d.Blockers) != want {
					t.Errorf("migrated plan %d holds %d blocker(s), want the archive's %d less its inquiry review", i, len(d.Blockers), want)
				}
				if d.PassPermitted != s.PassPermitted {
					t.Errorf("migrated plan %d: pass_permitted %v, archived %v", i, d.PassPermitted, s.PassPermitted)
				}
				if len(d.Blockers) == 0 && !d.PassPermitted {
					t.Errorf("migrated plan %d holds no blocker and permits no PASS: no relay can carry it", i)
				}
				if k := d.WithKey().Key; d.Key != k {
					t.Errorf("migrated plan %d's key is %q and the plan's own is %q", i, d.Key, k)
				}
			}
			if held == 0 {
				t.Fatal("fixture: no archived plan holds an inquiry review, so the translation is tested by nothing")
			}
		})
	}
}

func countOf(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func wordsOf(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			t.Fatal(err)
		}
		out = append(out, w)
	}
	return out
}
