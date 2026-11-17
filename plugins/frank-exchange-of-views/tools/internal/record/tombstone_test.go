package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A TOMBSTONE HAS ITS SHAPE AND NOTHING MORE (validate's Tombstone arm): its gap is a gap, a ruling
// placement names a docket motion on that gap, and an estoppel row names none. And the one that
// stands is the latest row on its (gap, kind): an end row ends it, a later placement stands again.
func TestATombstoneHasItsShapeAndTheLatestRowStands(t *testing.T) {
	run := motionGuardRun(t)
	stone := func(gap string, kind recordpb.TombstoneKind, motion string, ends bool) *recordpb.Tombstone {
		ts := &recordpb.Tombstone{GapId: proto.String(gap), Kind: kind.Enum()}
		if motion != "" {
			ts.MotionId = proto.String(motion)
		}
		if ends {
			ts.Ends = proto.Bool(true)
		}
		return ts
	}
	ruling, estoppel := recordpb.TombstoneKind_TOMBSTONE_KIND_RULING, recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL
	for _, c := range []struct {
		name  string
		row   *recordpb.Tombstone
		wants string
	}{
		{"a ruling row naming no motion", stone("G1", ruling, "", false), "names the motion"},
		{"a ruling row naming another gap's motion", stone("G1", ruling, "M2", false), "no docket motion on that gap"},
		{"a ruling row naming a petition", stone("G1", ruling, "M3", false), "no docket motion on that gap"},
		{"an estoppel row naming a motion", stone("G1", estoppel, "M1", false), "name no motion"},
		{"an end row naming a motion", stone("G1", ruling, "M1", true), "name no motion"},
		{"a row on no gap", stone("G9", estoppel, "", false), "G9"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Append(sit(t, run, "judge"), c.row)
			mustRefuse(t, err, c.wants)
		})
	}
	standing := func() map[string]bool {
		t.Helper()
		db, err := openRunForRead(run)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := db.Query(`SELECT "gap_id" || ' ' || "kind" FROM "tombstone_standing"`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		out := map[string]bool{}
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				t.Fatal(err)
			}
			out[s] = true
		}
		return out
	}
	judge := sit(t, run, "judge")
	for i, step := range []struct {
		row  *recordpb.Tombstone
		want []string
	}{
		{stone("G1", ruling, "M1", false), []string{"G1 ruling"}},
		{stone("G2", estoppel, "", false), []string{"G1 ruling", "G2 estoppel"}},
		{stone("G1", ruling, "", true), []string{"G2 estoppel"}},
		{stone("G1", ruling, "M1", false), []string{"G1 ruling", "G2 estoppel"}},
		{stone("G2", estoppel, "", true), []string{"G1 ruling"}},
	} {
		mustAppend(t, judge, step.row)
		got := standing()
		if len(got) != len(step.want) {
			t.Fatalf("after step %d the standing tombstones are %v, want %v", i, got, step.want)
		}
		for _, w := range step.want {
			if !got[w] {
				t.Errorf("after step %d the standing tombstones are %v, want %v", i, got, step.want)
			}
		}
	}
}

// ONLY THE BENCH OWES AN OPINION (requireBenchOpinion): a bench ruling missing any of its four
// opinion fields is refused, and the gap's assignee's accept carrying one is refused.
func TestTheBenchAloneOwesTheOpinionFields(t *testing.T) {
	for _, field := range []string{"principle", "tension", "review_flag", "settled"} {
		t.Run("the bench without "+field, func(t *testing.T) {
			run := motionGuardRun(t)
			r := actRule("M1", false, "judge")
			d := r.GetDocket()
			switch field {
			case "principle":
				d.Principle = nil
			case "tension":
				d.Tension = nil
			case "review_flag":
				d.ReviewFlag = nil
			case "settled":
				d.Settled = nil
			}
			_, err := Append(sit(t, run, "judge"), r)
			mustRefuse(t, err, "requires --")
		})
	}
	t.Run("the assignee with an opinion field", func(t *testing.T) {
		run := motionGuardRun(t)
		r := actRule("M1", true, evLens)
		r.GetDocket().Principle = proto.String("a rule")
		_, err := Append(sit(t, run, evLens), r)
		mustRefuse(t, err, "the bench's opinion")
	})
}
