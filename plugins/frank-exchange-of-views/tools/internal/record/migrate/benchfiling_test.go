package migrate_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE BENCH FILES NO DOCKET MOTION, ON ANY MIGRATED RECORD. An archive from before docket motions
// holds bench opinions, and migrate translates each to a docket motion and the bench's ruling on it.
// The bench rules the docket, and the write refuses a filing by the subject's ruler, so the motion
// is filed under the harness: in that era the engine put a gap before the bench and no seat filed
// it. Over every archive: each opinion is one motion filed by the harness and one ruling by a bench
// seat, and no docket motion is filed by a seat of the role that rules it.
func TestNoMigratedDocketMotionIsFiledByTheBench(t *testing.T) {
	ruler, err := recordpb.SubjectRuler(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET)
	if err != nil {
		t.Fatal(err)
	}
	opinions, synthesized := 0, 0
	for _, tb := range archiveTarballs(t) {
		name := filepath.Base(tb)
		res, run := migrateArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
		}
		merged, err := record.MergedEvents(run)
		if err != nil {
			t.Fatal(err)
		}
		ruledBy := map[string]string{} // motion id -> the seat of its ruling
		for _, e := range merged.Events {
			if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok {
				ruledBy[r.GetMotionId()] = e.GetSeatId()
			}
		}
		fromOpinion := 0
		for _, e := range merged.Events {
			m, ok := recordpb.BodyAs[*recordpb.Motion](e)
			if !ok || m.Subject() != recordpb.MotionSubject_MOTION_SUBJECT_DOCKET {
				continue
			}
			if seatclass.RoleOfSeat(e.GetSeatId()) == ruler {
				t.Errorf("%s: docket motion %s is filed by %s, a seat of the role that rules it", name, m.GetMotionId(), e.GetSeatId())
			}
			if !strings.HasPrefix(m.GetBasis(), "migrated from a bench opinion") {
				continue
			}
			fromOpinion++
			if e.GetSeatId() != record.HarnessSeat {
				t.Errorf("%s: the docket motion %s translated from an opinion is filed by %s, want the harness", name, m.GetMotionId(), e.GetSeatId())
			}
			if seat := ruledBy[m.GetMotionId()]; seatclass.RoleOfSeat(seat) != ruler {
				t.Errorf("%s: the ruling on %s, translated from an opinion, is by %q, want a %s seat", name, m.GetMotionId(), seat, ruler)
			}
		}
		if fromOpinion != res.In["opinion"] {
			t.Errorf("%s: %d docket motion(s) translated from %d opinion(s)", name, fromOpinion, res.In["opinion"])
		}
		opinions += res.In["opinion"]
		synthesized += fromOpinion
	}
	if opinions == 0 || synthesized != opinions {
		t.Errorf("the archives hold %d opinion(s) and %d translated docket motion(s); the census needs at least one", opinions, synthesized)
	}
}
