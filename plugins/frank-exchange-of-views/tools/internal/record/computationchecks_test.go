package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE RECORD REFUSES A COMPUTATION'S SHAPE THE BASIS DOES NOT ADMIT, and a reproduce's comparison
// that does not say what it found (plans/compute-and-show.md §V V1a.6). Each arm is a CHECK on the
// table, so no write path — the verb's, migrate's or a hand-built one — stores the state the verb
// refuses.
func TestTheRecordRefusesAComputationsUnadmittedShape(t *testing.T) {
	variant, reproducible := recordpb.Basis_BASIS_VARIANT, recordpb.Basis_BASIS_REPRODUCIBLE
	for name, c := range map[string]struct {
		body *recordpb.Computation
		ok   bool
	}{
		"variant with drift and invariant": {&recordpb.Computation{Basis: variant.Enum(), Drift: proto.String("moved"), Invariant: proto.String("holds")}, true},
		"variant without drift":            {&recordpb.Computation{Basis: variant.Enum(), Invariant: proto.String("holds")}, false},
		"variant without invariant":        {&recordpb.Computation{Basis: variant.Enum(), Drift: proto.String("moved")}, false},
		"variant with a blank invariant":   {&recordpb.Computation{Basis: variant.Enum(), Drift: proto.String("moved"), Invariant: proto.String(" \n")}, false},
		"reproducible with neither":        {&recordpb.Computation{Basis: reproducible.Enum()}, true},
		"reproducible with drift":          {&recordpb.Computation{Basis: reproducible.Enum(), Drift: proto.String("moved")}, false},
		"reproducible with an invariant":   {&recordpb.Computation{Basis: reproducible.Enum(), Invariant: proto.String("holds")}, false},
		"no basis":                         {&recordpb.Computation{}, false},
	} {
		t.Run(name, func(t *testing.T) {
			b := c.body
			b.ComputationId, b.BundleSha, b.Argv = proto.String("K-00000001"), proto.String("s1"), []string{"python3", "script.py"}
			_, err := Append(Identity{Run: mustRun(t, newRun(t)), SeatID: "blue-respond"}, b)
			if (err == nil) != c.ok {
				t.Errorf("recorded = %v, want %v: %v", err == nil, c.ok, err)
			}
		})
	}
}

func TestTheRecordRefusesAReproducesUnadmittedShape(t *testing.T) {
	ran := func(rerun recordpb.Rerun) *recordpb.Reproduce {
		return &recordpb.Reproduce{Rerun: rerun.Enum(), Ran: proto.Bool(true), RerunSha: proto.String(rerunOf("s1"))}
	}
	reproduced, differs := recordpb.Rerun_RERUN_REPRODUCED, recordpb.Rerun_RERUN_DIFFERS
	for name, c := range map[string]struct {
		body *recordpb.Reproduce
		ok   bool
	}{
		// A reproduce names its execution: the live write's act judgement (requireNamedExecution).
		"a ran reproduce with neither rerun_exit nor rerun_sha": {func() *recordpb.Reproduce { r := ran(reproduced); r.RerunSha = nil; return r }(), false},
		"reproduced with drift":                                 {func() *recordpb.Reproduce { r := ran(reproduced); r.Drift = proto.String("x"); return r }(), false},
		"differs with drift":                                    {func() *recordpb.Reproduce { r := ran(differs); r.Drift = proto.String("x"); return r }(), true},
		"differs without drift":                                 {ran(differs), false},
		"rerun_exit with ran and rerun_sha": {func() *recordpb.Reproduce {
			r := ran(reproduced)
			r.RerunExit, r.RerunSha = proto.Int32(0), proto.String("r1")
			return r
		}(), true},
		"rerun_exit without rerun_sha": {func() *recordpb.Reproduce {
			r := ran(reproduced)
			r.RerunExit, r.RerunSha = proto.Int32(0), nil
			return r
		}(), false},
		"rerun_exit without ran": {func() *recordpb.Reproduce {
			r := ran(reproduced)
			r.Ran, r.RerunExit, r.RerunSha = proto.Bool(false), proto.Int32(0), proto.String("r1")
			return r
		}(), false},
		"no ran": {func() *recordpb.Reproduce { r := ran(reproduced); r.Ran = nil; return r }(), false},
	} {
		t.Run(name, func(t *testing.T) {
			b := reopenedBoard(t).computation("K-1", "s1")
			if !c.body.GetRan() {
				// An act that ran nothing stands on a re-run the record holds.
				b.reproduce(evLens, "K-1", "s1", reproduced, recordpb.Soundness_SOUNDNESS_SOUND)
			}
			r := c.body
			r.ComputationId, r.BundleSha = proto.String("K-1"), proto.String("s1")
			r.Soundness, r.Note = recordpb.Soundness_SOUNDNESS_SOUND.Enum(), proto.String("read")
			_, err := Append(Identity{Run: b.seed(), SeatID: evLens}, r)
			if (err == nil) != c.ok {
				t.Errorf("recorded = %v, want %v: %v", err == nil, c.ok, err)
			}
		})
	}
}

// AN ACT THAT RAN NOTHING NAMES THE STANDING RE-RUN, and is refused where it names another or where
// nothing ever ran the bundle.
func TestAReproduceThatRanNothingStandsOnTheBundlesReRun(t *testing.T) {
	notRan := func(rerun recordpb.Rerun) *recordpb.Reproduce {
		r := &recordpb.Reproduce{ComputationId: proto.String("K-1"), BundleSha: proto.String("s1"), Rerun: rerun.Enum(), Ran: proto.Bool(false),
			Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Note: proto.String("read"), RerunSha: proto.String(rerunOf("s1"))}
		if rerun == recordpb.Rerun_RERUN_DIFFERS {
			r.Drift = proto.String("x")
		}
		return r
	}
	never := reopenedBoard(t).computation("K-1", "s1").seed()
	if _, err := Append(Identity{Run: never, SeatID: evLens}, notRan(recordpb.Rerun_RERUN_REPRODUCED)); err == nil {
		t.Error("a reproduce that ran nothing was recorded on a bundle nothing ever ran")
	}
	stood := reopenedBoard(t).computation("K-1", "s1").reproduce(evLens, "K-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND).seed()
	stale := notRan(recordpb.Rerun_RERUN_REPRODUCED)
	stale.RerunSha = proto.String("r-elsewhere")
	if _, err := Append(Identity{Run: stood, SeatID: evLens}, stale); err == nil {
		t.Error("a reproduce that ran nothing was recorded naming a re-run other than the standing one")
	}
	if _, err := Append(Identity{Run: stood, SeatID: evLens}, notRan(recordpb.Rerun_RERUN_REPRODUCED)); err != nil {
		t.Errorf("a reproduce that ran nothing and states the standing re-run's comparison was refused: %v", err)
	}
}

// AN ARCHIVED REPRODUCE CAN NAME NO EXECUTION: its archive holds the re-run's output only in part. The replay
// writes one, ran or not, as the archive holds it; the live write refuses it (requireNamedExecution).
func TestTheReplayWritesAReproduceThatNamesNoExecution(t *testing.T) {
	for _, ran := range []bool{true, false} {
		b := reopenedBoard(t).computation("K-1", "s1")
		r := &recordpb.Reproduce{ComputationId: proto.String("K-1"), BundleSha: proto.String("s1"), Rerun: recordpb.Rerun_RERUN_REPRODUCED.Enum(), Ran: proto.Bool(ran),
			Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Note: proto.String("read")}
		run := b.seed()
		if err := wouldReplay(t, run, evLens, recordpb.EventType_EVENT_TYPE_REPRODUCE, r); err != nil {
			t.Errorf("ran=%v: the replay refused an archived reproduce that names no execution: %v", ran, err)
		}
		if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_REPRODUCE, r); err == nil {
			t.Errorf("ran=%v: the live write admitted a reproduce that names no execution", ran)
		}
	}
}

// AN ACT WITH NO COMPARISON STANDS ON A NAMED EXECUTION OR ON NONE. A bundle's standing re-run is the
// latest reproduce that ran it and names its execution; a migrated re-run that differed names none,
// so another act on that bundle that nobody compared is not_rerun, and one on a bundle whose re-run
// is named is not_compared.
func TestAnActWithNoComparisonStandsOnlyOnANamedExecution(t *testing.T) {
	for _, tc := range []struct {
		sha   string
		rerun recordpb.Rerun
		want  string
	}{
		{"", recordpb.Rerun_RERUN_DIFFERS, RerunNone},
		{"e1", recordpb.Rerun_RERUN_REPRODUCED, RerunNotCompared},
	} {
		ev := &Event{SeatId: proto.String(evLens)}
		r := &recordpb.Reproduce{ComputationId: proto.String("K-1"), BundleSha: proto.String("s1"), Ran: proto.Bool(true), Rerun: tc.rerun.Enum()}
		if tc.sha != "" {
			r.RerunSha = proto.String(tc.sha)
		}
		if _, err := recordpb.SetBody(ev, r); err != nil {
			t.Fatal(err)
		}
		evs := []*Event{ev}
		if got := computationVerdictOf("K-2", "s1", nil, standingRerunsOf(evs), actVerdictsOf(evs)).rerunWord(); got != tc.want {
			t.Errorf("an act with no comparison on a bundle whose re-run %s and names %q reads %s, want %s", recordpb.Word(tc.rerun), tc.sha, got, tc.want)
		}
	}
}
