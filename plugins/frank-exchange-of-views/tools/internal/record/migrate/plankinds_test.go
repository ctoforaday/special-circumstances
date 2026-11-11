package migrate_test

import (
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// readingBlockersTranslated is a source plan's printing with each archived reading blocker carried to
// the debt TranslateReadingBlockers writes for it — kind reading_debt, subject the anchor the
// source's reading at that events.id names (anchors, sourceReadingAnchors) — or, for a reading that
// names no anchor, dropped; re-keyed.
func readingBlockersTranslated(t *testing.T, printed string, anchors map[int64]string) string {
	t.Helper()
	p, _, err := record.ReadPlan([]byte(printed))
	if err != nil {
		t.Fatalf("a stored plan does not read: %v", err)
	}
	changed := false
	kept := p.Blockers[:0:0]
	for _, b := range p.Blockers {
		if b.Kind != "contradiction" {
			kept = append(kept, b)
			continue
		}
		n, err := strconv.ParseInt(b.Subject, 10, 64)
		a, read := anchors[n]
		if err != nil || !read {
			t.Fatalf("a stored plan's reading blocker names %q, which is no reading in the source", b.Subject)
		}
		changed = true
		if a != "" {
			b.Kind, b.Subject = record.BlockerReadingDebt, a
			kept = append(kept, b)
		}
	}
	if !changed {
		return printed
	}
	p.Blockers = kept
	return p.WithKey().Printed()
}

// sourceReadingAnchors is the anchor each of a source record's readings names, by events.id.
func sourceReadingAnchors(t *testing.T, srcDir string) map[int64]string {
	t.Helper()
	rows, err := openRO(t, srcDir).Query(`SELECT "event_id", COALESCE("anchor", '') FROM "verify"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[int64]string{}
	for rows.Next() {
		var id int64
		var a string
		if err := rows.Scan(&id, &a); err != nil {
			t.Fatal(err)
		}
		out[id] = a
	}
	return out
}

// kindsTranslated is a source plan's printing as TranslatePlanKinds leaves it: the comparison a
// census makes of an archived plan with its migrated one is of what the other translations did to it.
func kindsTranslated(t *testing.T, printed string) string {
	t.Helper()
	pl := &recordpb.Plan{Printed: proto.String(printed)}
	if err := migrate.TranslatePlanParties(pl); err != nil {
		t.Fatalf("a stored plan's parties do not translate: %v", err)
	}
	if err := migrate.TranslatePlanKinds(pl); err != nil {
		t.Fatalf("a stored plan does not translate: %v", err)
	}
	return pl.GetPrinted()
}

// THE ARCHIVED BLOCKER KINDS ARE TRANSLATED, AND NOTHING ELSE IS (EventSchema 48, 59). Four archives
// store plans printing the kind `reopened` (m28, m30, 1565, m32): every one migrates printing
// `unchecked` in its place, re-keyed, with no archived kind word left on any migrated plan, and each
// plan's other fields as the census tests hold them (TestAnArchivedRunStoresOnlyThePlansItStored).
// m32's plans also print the archived reading blocker `contradiction`, which migrates as the reader's
// reading_debt on the citation the reading read (TranslateReadingBlockers).
func TestArchivedBlockerKindsAreTranslated(t *testing.T) {
	for _, name := range []string{"2026-10-18_is-91-prime-m28.tar.gz", "2026-10-19_is-91-prime-m30.tar.gz", "2026-10-24_is-91-prime-1565.tar.gz", "2026-10-24_is-91-prime-m32.tar.gz"} {
		t.Run(name, func(t *testing.T) {
			res, _ := migrateArchive(t, name)
			if len(res.Refusals) != 0 {
				t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
			}
			srcPlans, dstPlans, _, _, _ := planPairsOf(t, name)
			archived, translated := 0, 0
			for i := range srcPlans {
				archived += strings.Count(srcPlans[i].printed, `"kind":"reopened"`) + strings.Count(srcPlans[i].printed, `"kind":"unsound_proof"`) +
					strings.Count(srcPlans[i].printed, `"kind":"contradiction"`)
				d := dstPlans[i].printed
				if strings.Contains(d, `"kind":"reopened"`) || strings.Contains(d, `"kind":"unsound_proof"`) || strings.Contains(d, `"kind":"contradiction"`) {
					t.Errorf("migrated plan %d still prints an archived kind: %s", i, d)
				}
				p, _, err := record.ReadPlan([]byte(d))
				if err != nil {
					t.Fatalf("migrated plan %d does not read: %v", i, err)
				}
				if want := p.WithKey().Key; p.Key != want {
					t.Errorf("migrated plan %d's key is %q and the plan's own is %q", i, p.Key, want)
				}
				translated += strings.Count(d, `"kind":"unchecked"`) + strings.Count(d, `"kind":"failed_check"`) + strings.Count(d, `"kind":"reading_debt"`)
			}
			if archived == 0 {
				t.Fatal("fixture: the archive's stored plans print no archived kind, so the translation is tested by nothing")
			}
			if translated != archived {
				t.Errorf("%d archived kind word(s) became %d current one(s)", archived, translated)
			}
		})
	}
}

// A GAP ABOUT A PROOF IS A GAP ABOUT ITS ANCHOR (EventSchema 48): 1565's gap about the proof its
// computation lens judged unsound migrates about the anchor.
func TestTheArchivedAboutKindProofIsAnAnchor(t *testing.T) {
	res, dst := migrateArchive(t, "2026-10-24_is-91-prime-1565.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	about := 0
	for _, e := range fam.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR {
			about++
			if !strings.HasPrefix(m.GetAboutRef(), "K-") {
				t.Errorf("gap %s is about %q, want the computation anchor the proof it was about translates to", m.GetGapId(), m.GetAboutRef())
			}
		}
	}
	if about != 1 {
		t.Errorf("the migrated 1565 holds %d gap(s) about an anchor, want its one gap about a proof", about)
	}
}
