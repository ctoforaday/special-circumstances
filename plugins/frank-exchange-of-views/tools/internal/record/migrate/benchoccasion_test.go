package migrate_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN ARCHIVED BENCH ROW IS A DOCKET. Every archived bench row carries gaps, whatever occasions it named,
// and the write refuses a bench row with none. The census reads every archive: a bench row migrated
// with no gap fails; and an archive set holding no bench row at all fails too, because the translation
// would then be tested by nothing.
func TestAnArchivedBenchDispatchMigratesAsADocket(t *testing.T) {
	tarballs := migrate.ArchivedRuns(t)
	bench := 0
	for _, tb := range tarballs {
		name := tb
		res, dst := migrateArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
		}
		fam, err := record.FamilyOf(dst)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range fam.Events {
			d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
			if !ok {
				continue
			}
			if !record.SeatOwesOccasion(d.GetSeatId()) {
				continue
			}
			bench++
			if len(d.GetGapIds()) == 0 {
				t.Errorf("%s: a bench row migrated with no gap", name)
			}
		}
	}
	if bench == 0 {
		t.Fatal("no archived run holds a bench dispatch row — the translation is exercised by nothing")
	}
	t.Logf("%d archived bench row(s) migrated as a docket", bench)
}

// AN ARCHIVED RUN'S STORED PLANS ARE THE ONES IT STORED, AND ITS ROWS NO REMAND EXCHANGE. An archive
// written before the stored plan (epoch 31) migrates with none: migrate reconstructs none — today's
// fold over the record's prefix is not what the binary that ran it printed — so the relay audit on
// such a run reports that it compared nothing. An archive written at epoch 31 or later carries the
// plans it stored — byte for byte where the replay moved no event a plan names, and otherwise with
// only those numbers and the key changed (TestAStoredPlanCrossesARenumbering). Which numbers moved
// is read from the two databases: a source number whose event sits at that number on the migrated
// record too is one the migrated plan keeps. The census reads every archive and fails on a plan the
// source did not store, a stored plan lost, a word changed or a number that did not move changed,
// or an archive set holding no dispatch row, because the translation would then be tested by
// nothing. A plan stored after its chair sitting recorded the PASS carries it migrated
// (CarryPlanPass) — the one field added, never in the key — and the census counts them, failing
// an archive set in which no plan carries one. A plan stored at EventSchema 78 or later prints
// `passed` itself, and migrates byte for byte.
func TestAnArchivedRunStoresOnlyThePlansItStored(t *testing.T) {
	tarballs := migrate.ArchivedRuns(t)
	rows, carried, identical, verdicts, fails := 0, 0, 0, 0, 0
	for _, tb := range tarballs {
		name := tb
		res, dst := migrateArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
		}
		srcPlans, dstPlans, srcEvs, dstEvs, _ := planPairsOf(t, name)
		var anchors map[int64]string
		if len(srcPlans) > 0 {
			anchors = sourceReadingAnchors(t, recordtest.ExtractArchive(t, name))
		}
		if len(dstPlans) != len(srcPlans) {
			t.Errorf("%s: migrated with %d stored plan(s), the source stored %d — a migrated record stores exactly the source's plans", name, len(dstPlans), len(srcPlans))
			continue
		}
		for i := range srcPlans {
			src := srcPlans[i].printed
			// A PLAN STORED AFTER A FAIL PRINTS IT, and the FAIL translates to no event (EventSchema
			// 76): the migrated plan prints no PASS. One stored after a PASS prints `passed`.
			if strings.HasSuffix(src, `,"verdict":"FAIL"}`) {
				if strings.Contains(dstPlans[i].printed, `"passed":`) {
					t.Errorf("%s: stored plan %d printed a FAIL and migrated printing the PASS: %s", name, i, dstPlans[i].printed)
				}
				fails++
			}
			s, d := readingBlockersTranslated(t, kindsTranslated(t, src), anchors), dstPlans[i].printed
			if v := `,"passed":true}`; !strings.Contains(src, `"verdict":`) && !strings.HasSuffix(src, v) && strings.HasSuffix(d, v) {
				d = strings.TrimSuffix(d, v) + "}"
				verdicts++
			}
			if s == d {
				identical++
				continue
			}
			if maskedPlan(s) != maskedPlan(d) {
				t.Errorf("%s: stored plan %d migrated with more than its numbers and key changed:\nsource   %s\nmigrated %s", name, i, s, d)
				continue
			}
			sn, dn := numeralsOf(s), numeralsOf(d)
			kept := true
			for k := range sn {
				if sn[k].n == dn[k].n {
					continue
				}
				kept = false
				if !moved(sn[k].n, srcEvs, dstEvs) {
					t.Errorf("%s: stored plan %d changed %d to %d, though source event %d is event %d on the migrated record too — byte for byte where nothing a plan names moved", name, i, sn[k].n, dn[k].n, sn[k].n, sn[k].n)
				}
			}
			if kept {
				t.Errorf("%s: stored plan %d changed its key and no number — byte for byte where nothing a plan names moved", name, i)
			}
		}
		carried += len(srcPlans)
		fam, err := record.FamilyOf(dst)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range fam.Events {
			if _, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok {
				rows++
			}
		}
	}
	if rows == 0 {
		t.Fatal("no archived run holds a dispatch row — the translation is exercised by nothing")
	}
	if identical == 0 {
		t.Fatal("no stored plan migrated byte for byte — nothing here holds a plan whose events kept their numbers")
	}
	if verdicts == 0 {
		t.Fatal("no archived plan carries its sitting's PASS migrated — the PASS's carrying is exercised by nothing")
	}
	if fails == 0 {
		t.Fatal("no archived plan prints a FAIL — the stripping of the retired verdict is exercised by nothing")
	}
	t.Logf("%d archived dispatch row(s) migrated; %d stored plan(s) carried, %d byte for byte apart from %d PASS(es) carried and %d FAIL(s) stripped", rows, carried, identical, verdicts, fails)
}

// A ROW'S REMAND EXCHANGE IS NOT CARRIED: the row is translated word for word without it.
func TestADispatchRowsRemandOwedIsNotCarried(t *testing.T) {
	bodies, err := migrate.Entries().Translate(migrate.OldEvent{ID: 9, Word: "dispatch",
		Fields: map[string]any{"pin": int64(4), "seat_id": "blue-respond"},
		Lists:  map[string][]string{"gap_ids": {"G-00000001"}, "remand_owed": {"G-00000001"}}})
	if err != nil || len(bodies) != 1 {
		t.Fatalf("translate = %d bodies, %v", len(bodies), err)
	}
	d, ok := bodies[0].(*recordpb.Dispatch)
	if !ok || d.GetPin() != 4 || d.GetSeatId() != "blue-respond" || !slices.Equal(d.GetGapIds(), []string{"G-00000001"}) {
		t.Fatalf("translated to %+v", bodies[0])
	}
}
