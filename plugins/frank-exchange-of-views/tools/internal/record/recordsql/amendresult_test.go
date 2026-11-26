package recordsql

import (
	"database/sql"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE INSERT STAMPS AN AMEND WITH THE GAP'S WORDING ONCE IT STANDS, over whatever it carries, and the gap
// view reads the latest amend's stamp as the current wording while the mint's own columns stay the
// mint's. Each amend arrives with a deliberately wrong result. The second amend names only fix_new,
// so its basis is its own (absent) and the problem the first amend restated stands.
func TestAnAmendIsStampedAndTheGapViewReadsItsWording(t *testing.T) {
	db := store(t)
	mintRow(t, db, 0, "G1", recordpb.CheckKind_CHECK_KIND_DOCUMENT) // problem "p", acceptance check "a"
	wrong := proto.String("wrong")
	for i, a := range []*recordpb.Amend{
		{GapId: proto.String("G1"), Problem: proto.String("p2"), FixNew: proto.String("f1"), FixBasis: proto.String("b1")},
		{GapId: proto.String("G1"), FixNew: proto.String("f2")},
	} {
		a.Basis = proto.String("restated")
		a.ResultProblem, a.ResultRequiredFix, a.ResultAcceptanceCheck, a.ResultFixNew, a.ResultFixBasis = wrong, wrong, wrong, wrong, wrong
		if _, err := Insert(db, event(t, int32(1+i), recordpb.EventType_EVENT_TYPE_AMEND, a)); err != nil {
			t.Fatal(err)
		}
	}
	var amendEvent sql.NullInt64
	var problem, fix, check, fixNew, fixBasis, mintProblem sql.NullString
	if err := db.QueryRow(`SELECT "amend_event", "current_problem", "current_required_fix", "current_acceptance_check",
	    "current_fix_new", "current_fix_basis", "problem" FROM "gap" WHERE "gap_id" = 'G1'`).
		Scan(&amendEvent, &problem, &fix, &check, &fixNew, &fixBasis, &mintProblem); err != nil {
		t.Fatal(err)
	}
	got := []any{amendEvent.Valid, problem.String, fix.Valid, check.String, fixNew.String, fixBasis.Valid, mintProblem.String}
	want := []any{true, "p2", false, "a", "f2", false, "p"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("gap view (amend_event set, current problem, required fix set, check, fix_new, fix_basis set, mint problem) = %v, want %v", got, want)
			break
		}
	}
}
