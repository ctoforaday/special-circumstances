package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The cast is the record's, written once under harness; membership admits a seat and its
// petition sitting, and "no cast" is told apart from "not in it".
func TestTheCastIsReadOffTheRecordAndAdmitsPetitionSittings(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	if cast, err := CastOf(run); err != nil || cast != nil {
		t.Fatalf("an empty record has no cast: (%v, %v)", cast, err)
	}
	if member, has, err := InCast(run, "red-chair"); err != nil || member || has {
		t.Fatalf("no cast: (%v, %v, %v), want (false, false, nil)", member, has, err)
	}
	recordtest.Seed(t, runDir, recordtest.At(t, "harness", "harness:cast:#1", &recordpb.Cast{
		SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge"},
	}))
	cast, err := CastOf(run)
	if err != nil || len(cast) != 5 || cast[0] != "red-lens-evidence" || cast[4] != "judge" {
		t.Fatalf("CastOf = (%v, %v)", cast, err)
	}
	for seat, want := range map[string]bool{
		"red-chair":                true,
		"judge-petition-red-chair": true, // a petition sitting by a cast seat
		"red-lens-voice":           false,
		"judge-petition-frontier":  false, // frontier is not in this cast
	} {
		member, has, err := InCast(run, seat)
		if err != nil || !has || member != want {
			t.Errorf("InCast(%q) = (%v, %v, %v), want member=%v with a cast present", seat, member, has, err, want)
		}
	}
}
