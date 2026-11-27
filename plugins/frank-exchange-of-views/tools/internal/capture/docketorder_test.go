package capture

import (
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// T-B33: THE DOCKET-ORDER AUDIT READS DOCKETINGS (audit p7 gap 2). The dispatch writes a gap before the
// bench as a Docketing row and no motion, so the audit reads those rows: an IMPASSE docketing written
// before the gap's lens had its re-audit is named, and a GRADE_MOVEMENT or PROPOSAL docketing at the same
// place is not — grade movement is docketed whatever the lens's turn, and a proposal is a party's. Fails
// if the reader reads motions (zero) or counts another cause.
func TestTheDocketOrderAuditReadsDocketings(t *testing.T) {
	const lens = "red-lens-evidence"

	// THE DISPATCH'S OWN WRITE (record.DispatchNext, which `dispatch next` calls). With the lens's
	// re-audit still to come the plan readies the lens and writes nothing; after it, the plan writes the
	// impasse docketing, no motion, and the audit passes either way.
	t.Run("the dispatch's write", func(t *testing.T) {
		for _, lensSat := range []bool{false, true} {
			run := seedKMax1(t, append(docketOrderBoard(t, lens, lensSat), recordtest.At(t, "red-chair", "register-chair", &recordpb.Register{})))
			if _, _, err := record.DispatchNext(record.Identity{Run: run, SeatID: "red-chair"}); err != nil {
				t.Fatalf("lens sat %v: dispatch next: %v", lensSat, err)
			}
			db := runtest.Open(t, run.Dir())
			docketings, _, err := record.EventsOf(db, recordpb.EventType_EVENT_TYPE_DOCKETING)
			if err != nil {
				t.Fatal(err)
			}
			motions, _, _ := record.EventsOf(db, recordpb.EventType_EVENT_TYPE_MOTION)
			want := 0
			if lensSat {
				want = 1
			}
			if len(docketings) != want || len(motions) != 0 {
				t.Errorf("lens sat %v: the dispatch wrote %d docketing(s) and %d motion(s), want %d and none", lensSat, len(docketings), len(motions), want)
			}
			for _, e := range docketings {
				if d, _ := recordpb.BodyAs[*recordpb.Docketing](e); d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
					t.Errorf("lens sat %v: the dispatch's docketing is %s, want impasse", lensSat, recordpb.Word(d.GetCause()))
				}
			}
			if a := DocketOrderAudit(run); a.Verdict != "PASS" {
				t.Errorf("lens sat %v: a record the dispatch wrote by its rule: %s %s, want PASS", lensSat, a.Verdict, a.Detail)
			}
		}
	})

	// EACH CAUSE WRITTEN BEFORE THE LENS'S RE-AUDIT: only impasse is named.
	for _, tc := range []struct {
		cause recordpb.DocketingCause
		named bool
	}{
		{recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, true},
		{recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, false},
		{recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, false},
	} {
		t.Run(recordpb.Word(tc.cause)+" before the re-audit", func(t *testing.T) {
			evs := docketOrderBoard(t, lens, false)
			motion := ""
			if tc.cause == recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL {
				motion = "M1"
				evs = append(evs, recordtest.At(t, "blue-respond", "propose-G1", &recordpb.Motion{MotionId: proto.String(motion), Basis: proto.String("the bench's to settle"),
					Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
					Act: &recordpb.Motion_Regrade{Regrade: &recordpb.Regrade{GapId: proto.String("G1"),
						Severity: recordpb.Grade_GRADE_LOW.Enum(), Basis: proto.String("the harm needs two failures")}}}))
			}
			d := record.NewID("docketing")
			evs = append(evs, recordtest.At(t, "red-chair", "docket-G1", docketingOf(d, tc.cause, motion)))
			a := DocketOrderAudit(seedKMax1(t, evs))
			if named := a.Verdict == "FAIL" && strings.Contains(a.Detail, d+" on G1"); named != tc.named {
				t.Errorf("a %s docketing before the lens's re-audit: %s %s, want named %v", recordpb.Word(tc.cause), a.Verdict, a.Detail, tc.named)
			}
		})
	}

	// THE ARCHIVES. m37 holds six impasse dockets the dispatch filed before the lens's re-audit, all six
	// on the migrated record (the two on gaps a successor's mint superseded stand on those gaps' ids,
	// each ended by the Amend Translation 7 writes for its link) — and
	// m38 its adversary's G-ca807516 docketed in the final sitting with no re-audit; migrated, each is a
	// docketing of cause impasse, and the audit names them. Translation 3 (an archived dispatch docket
	// becomes a Docketing) is what these read.
	for _, tc := range []struct {
		archive string
		want    int
		gap     string
	}{
		{"bundle-store/2026-10-31_is-91-prime-m37.tar.gz", 6, ""},
		{"bundle-store/2026-11-04_is-91-prime-m38.tar.gz", 0, "G-ca807516"},
	} {
		t.Run(filepath.Base(tc.archive), func(t *testing.T) {
			migrated := filepath.Join(t.TempDir(), "run")
			t.Cleanup(func() { _ = recordsql.CloseUnder(migrated) })
			if _, err := migrate.Migrate(recordtest.RequireArchive(t, tc.archive), migrated, migrate.Entries(), migrate.Options{}); err != nil {
				t.Fatalf("migrating %s: %v", tc.archive, err)
			}
			run, err := record.NewRun(migrated)
			if err != nil {
				t.Fatal(err)
			}
			early, err := record.EarlyDockets(run)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want > 0 && len(early) != tc.want {
				t.Errorf("%d early impasse docketing(s), want %d: %+v", len(early), tc.want, early)
			}
			if tc.gap != "" {
				found := false
				for _, e := range early {
					found = found || e.GapID == tc.gap
				}
				if !found {
					t.Errorf("no early impasse docketing on %s: %+v", tc.gap, early)
				}
			}
			if a := DocketOrderAudit(run); a.Verdict != "FAIL" || (tc.gap != "" && !strings.Contains(a.Detail, " on "+tc.gap)) {
				t.Errorf("the audit on the migrated archive: %s %s, want FAIL naming its early impasse docketings", a.Verdict, a.Detail)
			}
		})
	}
}
