package migrate_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// argumentResidue is, per archived run, the closings and positions migrate writes as no event — the
// arguments the write path refuses (record.ClosingRefusal, record.PositionRefusal) — with what the
// migrated record then says of the closing and position duties. It is the residue stated with epoch
// 38 (gblock, 2026-10-23: "Drop all, stated"); run-archive/ holds every dropped argument byte for
// byte.
//
// THE DUTIES ARE THE ONES THE RECORD SAID BEFORE THE DROP. A dropped closing is the chair's, or
// blue's on a gap no unruled docket motion stood on, and a dropped position is no chair's, so none of
// them discharged a duty: a tally that moves here means the drop took an argument a duty read. The
// one exception is a docket on a gap a successor's mint had closed (epoch 62): m29's two, m30's
// one and mersenne's one are left out as acts on a closed gap, so no closing is owed on them, and
// their closings are written as none with them.
var argumentResidue = map[string]struct {
	closings, positions int
	duties, sittings    string
}{
	"2026-08-22_is-7-prime":                       {2, 1, "blue:awaited=2,lens:not_owed=2", "not_owed=2"},
	"2026-08-22_is-7-prime-run2":                  {0, 2, "blue:awaited=7,lens:not_owed=7", "not_owed=2"},
	"2026-08-22_record-store-authority":           {5, 6, "", "not_owed=2"},
	"2026-08-22_sqlite-schema-source":             {4, 3, "blue:awaited=2,lens:not_owed=2", "not_owed=2"},
	"2026-08-23_research-loop-counterparts":       {0, 7, "blue:awaited=3,lens:not_owed=3", "not_owed=4"},
	"2026-08-23_sleeper-service-plan":             {20, 9, "blue:awaited=10,lens:not_owed=10", "not_owed=4"},
	"2026-09-02_quadratic-formula":                {19, 10, "blue:awaited=17,lens:not_owed=17", "not_owed=5"},
	"2026-09-09_is-91-prime-a":                    {1, 5, "blue:not_owed=1,lens:not_owed=1", "not_owed=5"},
	"2026-09-09_is-91-prime-b":                    {0, 2, "", "not_owed=1"},
	"2026-09-10_is-91-prime-b3":                   {0, 6, "blue:awaited=1,lens:awaited=1", "not_owed=1"},
	"2026-09-11_is-91-prime-b4":                   {0, 4, "", "not_owed=1"},
	"2026-09-11_is-91-prime-b5":                   {0, 7, "", "not_owed=1"},
	"2026-09-11_is-91-prime-b6":                   {2, 5, "blue:awaited=2,lens:awaited=2", "not_owed=1"},
	"2026-09-11_is-91-prime-b7":                   {0, 5, "", "not_owed=5"},
	"2026-09-11_is-91-prime-b8":                   {0, 3, "", "not_owed=4"},
	"2026-09-11_is-91-prime-b9":                   {0, 7, "", "not_owed=6"},
	"2026-10-09_is-91-prime-wave-a":               {0, 5, "", "not_owed=6"},
	"2026-10-10_is-91-prime-m19":                  {0, 8, "", "not_owed=10"},
	"2026-10-11_is-91-prime-m20":                  {0, 2, "", "not_owed=7"},
	"2026-10-12_is-91-prime-wave-b":               {3, 6, "blue:filed=3,blue:not_owed=1,lens:missing=2,lens:not_owed=2", "not_owed=9"},
	"2026-10-13_is-91-prime-m23":                  {0, 5, "", "not_owed=8"},
	"2026-10-17_is-91-prime-07d":                  {1, 2, "blue:awaited=1,lens:awaited=1", "not_owed=5"},
	"2026-10-17_is-91-prime-final":                {0, 0, "", "not_owed=2"},
	"2026-10-17_is-91-prime-m26":                  {0, 3, "", "not_owed=8"},
	"2026-10-17_is-91-prime-m27":                  {0, 5, "", "not_owed=10"},
	"2026-10-18_is-91-prime-m28":                  {1, 8, "blue:filed=2,lens:not_owed=2", "not_owed=12"},
	"2026-10-18_is-91-prime-m29":                  {6, 6, "blue:filed=1,blue:not_owed=1,lens:not_owed=2", "not_owed=9"},
	"2026-10-18_is-91-prime-n1":                   {0, 0, "", "not_owed=2"},
	"2026-10-19_is-91-prime-m30":                  {2, 3, "", "not_owed=7"},
	"2026-10-24_is-91-prime-m32":                  {0, 0, "blue:filed=1,lens:filed=1", "filed=1,not_owed=6"},
	"2026-10-24_is-91-prime-1565":                 {0, 0, "blue:filed=3,blue:not_owed=2,lens:filed=5", "not_owed=12"},
	"2026-10-27_is-91-prime-m35":                  {0, 0, "blue:filed=8,lens:filed=7,lens:not_owed=1", "not_owed=6"},
	"2026-10-29_is-91-prime-m36":                  {0, 0, "", "not_owed=8"},
	"2026-10-30_mersenne-31-prime-sum":            {2, 0, "blue:filed=4,lens:filed=4", "not_owed=12"},
	"2026-11-04_eiffel-tower-height":              {0, 0, "blue:filed=3,lens:filed=3", "not_owed=12"},
	"2026-11-11_eiffel-tower-height":              {0, 0, "blue:filed=1,lens:filed=1", "not_owed=12"},
	"2026-11-10_auditing-constitutions-gap-audit": {0, 0, "", "not_owed=12"},
}

// EVERY ARCHIVED RUN'S REFUSED ARGUMENTS ARE WRITTEN AS NO EVENT, AND ONLY THEY. Per run, the
// manifest's events_in less its events_out for closing and position is the residue above; every
// closing the migrated record holds is a party's, and every position the chair's; and the closing
// and position duties read as they did before the drop.
func TestEveryArchivedRunDropsOnlyTheArgumentsTheWriteRefuses(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	seen := map[string]bool{}
	for _, tb := range tarballs {
		name := strings.TrimSuffix(filepath.Base(tb), ".tar.gz")
		seen[name] = true
		t.Run(name, func(t *testing.T) {
			want, stated := argumentResidue[name]
			if !stated {
				t.Fatalf("%s is archived and its residue is not stated in argumentResidue", name)
			}
			res, run := migrateArchive(t, name+".tar.gz")
			fam, err := record.FamilyOf(run)
			if err != nil {
				t.Fatal(err)
			}
			// A LOST ACT MOVES THE DUTIES ON ITS GAP, AND ONLY THERE. A bench ruling lost for want of
			// a delivery leaves its docket motion unruled, so the docketing never ends: a closing
			// filed after the archived ruling answers it (the write accepts it, so it is written
			// rather than dropped), a party's later sitting reads it missing, and a later close or
			// supersession reads it not owed. A lost close takes the closing act that made a duty not
			// owed. touched is each gap a lost close closed or a lost ruling's motion docketed.
			touched := migrate.LostCloseGaps(t, name+".tar.gz", res)
			lostRulings := migrate.LostRulingMotions(t, name+".tar.gz", res)
			for _, e := range fam.Events {
				if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok && lostRulings[m.GetMotionId()] && m.GetDocket().GetGapId() != "" {
					touched[m.GetDocket().GetGapId()] = true
				}
			}
			closings := res.In["closing"] - res.Out["closing"]
			positions := res.In["position"] - res.Out["position"]
			reaccepted := 0
			for _, e := range fam.Events {
				if c, ok := recordpb.BodyAs[*recordpb.Closing](e); ok && touched[c.GetGapId()] {
					reaccepted++
				}
			}
			if positions != want.positions || closings > want.closings || want.closings-closings > reaccepted {
				t.Errorf("migrate wrote %d closing(s) and %d position(s) as no event, want %d and %d, less at most the %d closing(s) written on a gap whose ruling or close is lost",
					closings, positions, want.closings, want.positions, reaccepted)
			}
			minter := map[string]string{}
			for _, e := range fam.Events {
				if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
					minter[m.GetGapId()] = e.GetSeatId()
				}
			}
			for _, e := range fam.Events {
				if c, ok := recordpb.BodyAs[*recordpb.Closing](e); ok && !record.IsClosingParty(e.GetSeatId(), minter[c.GetGapId()]) {
					t.Errorf("the migrated record holds %s's closing on %s, minted by %s — no party's", e.GetSeatId(), c.GetGapId(), minter[c.GetGapId()])
				}
				if _, ok := recordpb.BodyAs[*recordpb.Position](e); ok && record.PositionRefusal(e.GetSeatId()) != nil {
					t.Errorf("the migrated record holds a position from %s, which is not the chair", e.GetSeatId())
				}
			}
			duties, moved := map[string]int{}, map[string]int{}
			for _, d := range record.ClosingDuties(fam.Events, fam.At) {
				party := "lens"
				if d.Party == "blue-respond" {
					party = "blue"
				}
				if touched[d.GapID] {
					moved[party]++
					continue
				}
				duties[party+":"+string(d.State)]++
			}
			sittings := map[string]int{}
			for _, s := range record.PositionSittings(fam.Events, fam.At, record.AfterTheRun) {
				sittings[string(s.State)]++
			}
			// A duty on no touched gap reads as stated; the stated tally less those is the touched
			// duties' share, which must hold exactly as many duties per party as the fold now reads
			// on touched gaps. A duty that moves on an untouched gap breaks the first.
			if len(moved) == 0 {
				if got := tally(duties); got != want.duties {
					t.Errorf("closing duties read %q, want %q", got, want.duties)
				}
			} else {
				rest := parseTally(t, want.duties)
				for k, n := range duties {
					if rest[k] -= n; rest[k] < 0 {
						t.Errorf("closing duties on gaps no lost act names read %q, more %s than the stated %q", tally(duties), k, want.duties)
					}
				}
				share := map[string]int{}
				for k, n := range rest {
					share[strings.SplitN(k, ":", 2)[0]] += n
				}
				if tally(share) != tally(moved) {
					t.Errorf("the stated duties %q less those on untouched gaps %q leave %v per party; the fold reads %v on gaps a lost act names", want.duties, tally(duties), share, moved)
				}
			}
			if got := tally(sittings); got != want.sittings {
				t.Errorf("chair sittings' positions read %q, want %q", got, want.sittings)
			}
		})
	}
	for name := range argumentResidue {
		if !seen[name] {
			t.Errorf("argumentResidue states %s, which is not archived", name)
		}
	}
}

// tally renders counts as "k=n,…" in key order.
func tally(m map[string]int) string {
	var parts []string
	for k, n := range m {
		parts = append(parts, fmt.Sprintf("%s=%d", k, n))
	}
	slices.Sort(parts)
	return strings.Join(parts, ",")
}

// parseTally reads tally's "k=n,…" back into counts.
func parseTally(t *testing.T, s string) map[string]int {
	t.Helper()
	out := map[string]int{}
	if s == "" {
		return out
	}
	for _, kv := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(kv, "=")
		n, err := strconv.Atoi(v)
		if !ok || err != nil {
			t.Fatalf("stated tally %q does not read: %q", s, kv)
		}
		out[k] = n
	}
	return out
}
