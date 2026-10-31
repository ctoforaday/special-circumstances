package migrate_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// kindsTranslated is a source plan's printing as TranslatePlanKinds leaves it: the comparison a
// census makes of an archived plan with its migrated one is of what the other translations did to it.
func kindsTranslated(t *testing.T, printed string) string {
	t.Helper()
	pl := &recordpb.Plan{Printed: proto.String(printed)}
	if err := migrate.TranslatePlanKinds(pl); err != nil {
		t.Fatalf("a stored plan does not translate: %v", err)
	}
	return pl.GetPrinted()
}

// haltedCarried is a source plan as CarryPlanHalted leaves it, carrying the `halted` the migrated
// plan prints — the record's fact at the plan's place — and re-keyed over the plan as printed.
func haltedCarried(t *testing.T, src, migrated string) string {
	t.Helper()
	p, _, err := record.ReadPlan([]byte(src))
	if err != nil {
		t.Fatalf("a source plan does not read as a plan: %v", err)
	}
	p.Halted = dstHalted(t, migrated)
	return p.WithKey().Printed()
}

// THE ARCHIVED BLOCKER KINDS ARE TRANSLATED, AND NOTHING ELSE IS (EventSchema 48). Four archives
// store plans printing the kind `reopened` (m28, m30, 1565, m32): every one migrates printing
// `unchecked` in its place, re-keyed, with no archived kind word left on any migrated plan, and each
// plan's other fields as the census tests hold them (TestAnArchivedRunStoresOnlyThePlansItStored).
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
				archived += strings.Count(srcPlans[i].printed, `"kind":"reopened"`) + strings.Count(srcPlans[i].printed, `"kind":"unsound_proof"`)
				d := dstPlans[i].printed
				if strings.Contains(d, `"kind":"reopened"`) || strings.Contains(d, `"kind":"unsound_proof"`) {
					t.Errorf("migrated plan %d still prints an archived kind: %s", i, d)
				}
				p, _, err := record.ReadPlan([]byte(d))
				if err != nil {
					t.Fatalf("migrated plan %d does not read: %v", i, err)
				}
				if want := p.WithKey().Key; p.Key != want {
					t.Errorf("migrated plan %d's key is %q and the plan's own is %q", i, p.Key, want)
				}
				translated += strings.Count(d, `"kind":"unchecked"`) + strings.Count(d, `"kind":"failed_check"`)
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
