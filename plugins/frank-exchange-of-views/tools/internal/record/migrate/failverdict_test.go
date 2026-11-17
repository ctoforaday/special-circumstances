package migrate

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A FAIL TRANSLATES TO NO EVENT, AND THE PLAN STORED AFTER IT PRINTS NO PASS (EventSchema 78).
// universe 2026-11-11 eiffel recorded one FAIL at its final sitting (register, plan, FAIL, plan),
// and the plan it stored after the FAIL prints `"verdict":"FAIL"`. Migrated, the record holds no
// gate, the manifest counts the FAIL in and not out, every plan the archive stored is still stored,
// and none prints the PASS: the record at each plan's place holds none.
func TestAnArchivedFailMigratesToNoEvent(t *testing.T) {
	const name = "2026-11-11_eiffel-tower-height.tar.gz"
	src, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
	if err != nil {
		t.Fatal(err)
	}
	fails, plans, printedFail := 0, 0, 0
	for _, old := range src {
		switch old.Word {
		case "verdict":
			if w, _ := old.Fields["verdict"].(string); recordpb.SameWord(w, retiredVerdict) {
				fails++
			}
		case "plan":
			plans++
			if p, _ := old.Fields["printed"].(string); strings.Contains(p, `"verdict":"FAIL"`) {
				printedFail++
			}
		}
	}
	if fails != 1 || printedFail == 0 {
		t.Fatalf("fixture: the archive holds %d FAIL(s) and %d plan(s) printing one, want 1 and at least 1", fails, printedFail)
	}

	res, dir := MigratedArchive(t, name)
	if len(res.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	if in, out := res.In["verdict"], res.Out["verdict"]; in != fails || out != 0 {
		t.Errorf("manifest verdicts in %d out %d, want %d in and none out", in, out, fails)
	}
	fam, err := record.FamilyOf(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range fam.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_VERDICT {
			t.Errorf("the migrated record holds a gate at event %d; the archive's only verdict was a FAIL", fam.At.Of(e).ID)
		}
	}
	stored, err := record.StoredPlans(fam.Events)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != plans {
		t.Errorf("the migrated record stores %d plan(s), the archive %d", len(stored), plans)
	}
	for _, p := range stored {
		if p.Plan.Passed {
			t.Errorf("the plan stored at %d prints the PASS over a record that holds none", p.At)
		}
	}
}

// THE RESIDUE IS EVERY ARCHIVED FAIL, and nothing else: each archive's migrated record holds as many
// gates as its source held PASSes, and no stored plan prints the PASS on a record that holds none.
// The count is stated so a change to it is read, not passed: 25 FAILs in 11 runs — 16 in the six
// August round-era runs and quadratic-formula's 5, each a round's continue signal, and one at the
// final sitting of each of 1565, mersenne-31 and both eiffel runs.
func TestEveryArchivedFailIsTheResidue(t *testing.T) {
	const residue = 25
	total := 0
	for _, name := range archivedRuns(t) {
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			src, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			kept, dropped := 0, 0
			for _, old := range src {
				if old.Word != "verdict" {
					continue
				}
				// EVERY ARCHIVE PREDATES EPOCH 78, so each of its verdicts states its word: a wordless
				// gate is the live shape, and one here would pass through the replay as a PASS unread.
				raw, worded := old.Fields["verdict"]
				if !worded {
					t.Errorf("archived verdict event %d states no word", old.ID)
				}
				if w, _ := raw.(string); recordpb.SameWord(w, retiredVerdict) {
					dropped++
				} else {
					kept++
				}
			}
			total += dropped
			if dropped > 0 {
				t.Logf("%d FAIL(s) dropped", dropped)
			}
			res, dir := MigratedArchive(t, name)
			if len(res.Refusals) != 0 {
				t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
			}
			fam, err := record.FamilyOf(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			gates := 0
			for _, e := range fam.Events {
				if _, ok := recordpb.BodyAs[*recordpb.Gate](e); ok {
					gates++
				}
			}
			if gates != kept {
				t.Errorf("the migrated record holds %d gate(s), the archive %d PASS(es)", gates, kept)
			}
			stored, err := record.StoredPlans(fam.Events)
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range stored {
				if p.Plan.Passed && gates == 0 {
					t.Errorf("the plan stored at %d prints the PASS on a record that holds none", p.At)
				}
			}
		})
	}
	if total != residue {
		t.Errorf("the archives hold %d FAIL(s); the stated residue is %d", total, residue)
	}
}
