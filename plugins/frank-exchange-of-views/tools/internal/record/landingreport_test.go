package record

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// everyInput stands in for the renderer with one whose text is every input it is handed, so two
// renders are equal exactly when their inputs are, and it counts the renders. It is a left fold of the
// ops over the base, as the registered renderer is, and an op whose new text is "boom" does not apply,
// refused with its position in the replay, as reportproj.Replay refuses.
func everyInput(t *testing.T) *int {
	t.Helper()
	calls := new(int)
	withRenderer(t, func(base string, haveBase bool, ops []ReportOp) (string, error) {
		*calls++
		if !haveBase {
			return "no base", nil
		}
		text := base
		for i, op := range ops {
			if op.B == "boom" {
				return "", fmt.Errorf("replaying mutation %d of %d: it does not apply", i+1, len(ops))
			}
			text += fmt.Sprintf(" | %s %q %q %v", op.Kind, op.A, op.B, op.Exact)
		}
		return text, nil
	})
	return calls
}

// renderAndVersion is the run's report and its version, read off the run's handle.
func renderAndVersion(t *testing.T, run Run) (string, reportVersion) {
	t.Helper()
	db := recordOf(t, run)
	v, err := reportVersionAt(db)
	if err != nil {
		t.Fatal(err)
	}
	md, err := RenderAt(db)
	if err != nil {
		t.Fatal(err)
	}
	return md, v
}

// A RENDER AT AN EQUAL VERSION IS AN EQUAL RENDER. The version is what lets a write reuse the report
// it rendered before taking the lock (landingReport), so a write that changes what the render reads
// and leaves the version where it was would be decided on a report it does not land on. Each row
// changes the record one way, and the render — which here is its inputs, verbatim — must move exactly
// when the version does: every input of the render moves it, and an act the render does not read
// leaves it, which is what lets most writes reuse the render.
func TestTheReportVersionMovesWithEveryInputOfTheRender(t *testing.T) {
	everyInput(t)
	for _, c := range []struct {
		name   string
		change func(t *testing.T, dir string)
		moves  bool
	}{
		{"the base is ingested", func(t *testing.T, dir string) {
			db := seatWriter(t, filepath.Join(dir, "records", "record.db"))
			if _, err := db.Exec(`DELETE FROM "base_ingest"`); err != nil {
				t.Fatal(err)
			}
			md0, v0 := renderAndVersion(t, mustRun(t, dir))
			recordtest.Seed(t, dir, recordtest.Event(t, HarnessSeat, &recordpb.BaseIngest{Text: proto.String("# report, ingested again")}))
			if md1, v1 := renderAndVersion(t, mustRun(t, dir)); md1 == md0 || v1 == v0 {
				t.Errorf("ingesting the base moved the render %v and the version %v, want both", md1 != md0, v1 != v0)
			}
		}, true},
		{"an edit", func(t *testing.T, dir string) {
			recordtest.Seed(t, dir, recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Old: proto.String("report"), New: proto.String("account")}))
		}, true},
		{"an anchor placed", func(t *testing.T, dir string) {
			recordtest.Seed(t, dir, recordtest.Event(t, evLens, &recordpb.Anchor{Id: proto.String("F-2"), Location: proto.String("report")}))
		}, true},
		{"a retire taking an anchor out", func(t *testing.T, dir string) {
			recordtest.Seed(t, dir, recordtest.Event(t, "blue-respond", &recordpb.Retire{Claim: proto.String("c"), Reason: proto.String("r"), Anchors: []string{"F-1"}}))
		}, true},
		// The two rows below are no write path's: none updates or deletes a row the render reads. They
		// hold the version's two halves each to its own case — a row gone under a newer one moves only
		// the count, and one gone with another added moves only the newest id.
		{"an anchor's row removed", func(t *testing.T, dir string) {
			db := seatWriter(t, filepath.Join(dir, "records", "record.db"))
			if _, err := db.Exec(`DELETE FROM "anchor" WHERE "id" = 'F-1'`); err != nil {
				t.Fatal(err)
			}
		}, true},
		{"an anchor's row replaced by another's", func(t *testing.T, dir string) {
			db := seatWriter(t, filepath.Join(dir, "records", "record.db"))
			if _, err := db.Exec(`DELETE FROM "anchor" WHERE "id" = 'F-1'`); err != nil {
				t.Fatal(err)
			}
			recordtest.Seed(t, dir, recordtest.Event(t, evLens, &recordpb.Anchor{Id: proto.String("F-2"), Location: proto.String("report")}))
		}, true},
		{"an anchor with no location", func(t *testing.T, dir string) {
			recordtest.Seed(t, dir, recordtest.Event(t, evLens, &recordpb.Anchor{Id: proto.String("F-3")}))
		}, false},
		{"a mint", func(t *testing.T, dir string) {
			recordtest.Seed(t, dir, recordtest.Event(t, evLens, highMint("G7")))
		}, false},
		{"a sitting", func(t *testing.T, dir string) {
			recordtest.Seed(t, dir, recordtest.Event(t, "blue-respond", &recordpb.Register{}))
		}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := newRun(t)
			recordtest.Seed(t, dir,
				recordtest.Event(t, HarnessSeat, &recordpb.BaseIngest{Text: proto.String("# report")}),
				recordtest.Event(t, evLens, &recordpb.Anchor{Id: proto.String("F-1"), Location: proto.String("# report")}),
				recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Old: proto.String("# report"), New: proto.String("# the report")}))
			run := mustRun(t, dir)
			md0, v0 := renderAndVersion(t, run)
			c.change(t, dir)
			md1, v1 := renderAndVersion(t, run)
			if moved := v1 != v0; moved != c.moves {
				t.Errorf("the version moved = %v (%+v → %+v), want %v", moved, v0, v1, c.moves)
			}
			if rendered := md1 != md0; rendered != c.moves {
				t.Errorf("the render moved = %v, want %v:\nbefore %s\nafter  %s", rendered, c.moves, md0, md1)
			}
		})
	}
}

// landingWrite is what a guard that reads the report was handed, and what its write did under the
// lock to hand it over.
type landingWrite struct {
	saw              string
	sawErr           error
	folded, rendered int
}

// reportStage is a run with a base, an anchor placed on it, and a lens in a sitting, and a write whose
// one guard reads the report and writes nothing (ErrWritesNothing), so the question is only what the
// guard was handed.
func reportStage(t *testing.T) (Run, func() error, *landingWrite) {
	t.Helper()
	run := newStage(t).cast(evLens, "blue-respond").ingest().
		add(evLens, &recordpb.Anchor{Id: proto.String("F-1"), Location: proto.String("# report")}).register(evLens).seed()
	got := &landingWrite{}
	write := func() error {
		_, err := Append(Identity{Run: run, SeatID: evLens}, &recordpb.Anchor{Id: proto.String("F-9"), Location: proto.String("# report")},
			CheckOnReport(func(l *Landing) error {
				got.saw, got.sawErr = l.Report()
				got.folded, got.rendered = l.w.rendering.foldedInside, l.w.rendering.renderedInside
				return ErrWritesNothing
			}))
		return err
	}
	return run, write, got
}

// fullRender is the report as RenderAt renders it from the base, and its failure.
func fullRender(t *testing.T, run Run) (string, error) {
	t.Helper()
	return RenderAt(recordOf(t, run))
}

// A RENDER MADE BEFORE THE LOCK IS CARRIED TO THE REPORT THE WRITE LANDS ON. Another seat's edit and
// anchor commit at the instant the write opens its transaction — after the render the write made
// before the lock, before the guard reads it. The guard must read the report with both in it: their
// ops folded onto that render under the lock, which is the report rendered from the base.
func TestARenderBeforeTheLockIsFoldedOntoTheReportItLandsOn(t *testing.T) {
	calls := everyInput(t)
	run, write, got := reportStage(t)
	err := overtaken(t, run, []*Event{
		recordtest.At(t, "blue-respond", "race:edit", &recordpb.BlueEdit{Old: proto.String("report"), New: proto.String("account"), EditKey: proto.String("E1")}),
		recordtest.At(t, "red-lens-logic", "race:anchor", &recordpb.Anchor{Id: proto.String("F-2"), Location: proto.String("# account")}),
	}, write)
	if !errors.Is(err, ErrWritesNothing) {
		t.Fatalf("the write = %v, want the guard's ErrWritesNothing", err)
	}
	if got.folded != 1 || got.rendered != 0 || *calls != 2 {
		t.Errorf("folds under the lock = %d, renders under it = %d, renderer calls = %d; want 1, 0 and 2 (the render before the lock, the fold)",
			got.folded, got.rendered, *calls)
	}
	want, werr := fullRender(t, run)
	if got.sawErr != nil || werr != nil || got.saw != want {
		t.Errorf("the guard read %q (%v), want the report rendered from the base, %q (%v)", got.saw, got.sawErr, want, werr)
	}
}

// A RENDER MADE BEFORE THE LOCK IS THE ONE THE GUARD READS WHERE NOTHING MOVED THE REPORT. Another
// seat's act that the render does not read commits at the same instant; the write renders once,
// before the lock, and the guard reads that render.
func TestARenderBeforeTheLockIsReusedWhereTheReportStayed(t *testing.T) {
	calls := everyInput(t)
	run, write, got := reportStage(t)
	err := overtaken(t, run, []*Event{recordtest.At(t, "blue-respond", "race:register", &recordpb.Register{})}, write)
	if !errors.Is(err, ErrWritesNothing) {
		t.Fatalf("the write = %v, want the guard's ErrWritesNothing", err)
	}
	if got.folded != 0 || got.rendered != 0 || *calls != 1 {
		t.Errorf("folds under the lock = %d, renders under it = %d, renderer calls = %d; want 0, 0 and 1", got.folded, got.rendered, *calls)
	}
	if want, _ := fullRender(t, run); got.saw != want {
		t.Errorf("the guard read %s, want the report %s", got.saw, want)
	}
}

// A FOLD IS MADE ONLY OVER THE ROWS IT WAS RENDERED FROM. A row the render read is gone as another
// write's edit lands — no write path does that, and the count is what says it happened — so the ops
// above the render's newest id are not all that moved, and the write renders the report from the base.
func TestARenderBeforeTheLockIsNotFoldedWhereARowItReadMoved(t *testing.T) {
	everyInput(t)
	run, write, got := reportStage(t)
	db := seatWriter(t, filepath.Join(run.Records(), "record.db"))
	err := overtaken(t, run, []*Event{
		recordtest.At(t, "blue-respond", "race:edit", &recordpb.BlueEdit{Old: proto.String("report"), New: proto.String("account"), EditKey: proto.String("E1")}),
	}, func() error {
		landsOthers := WriteBegins
		WriteBegins = func() {
			landsOthers()
			if _, err := db.Exec(`DELETE FROM "anchor" WHERE "id" = 'F-1'`); err != nil {
				t.Fatal(err)
			}
		}
		return write()
	})
	if !errors.Is(err, ErrWritesNothing) {
		t.Fatalf("the write = %v, want the guard's ErrWritesNothing", err)
	}
	if got.folded != 0 || got.rendered != 1 {
		t.Errorf("folds under the lock = %d, renders under it = %d; want 0 and 1", got.folded, got.rendered)
	}
	if want, _ := fullRender(t, run); got.saw != want || strings.Contains(got.saw, "F-1") {
		t.Errorf("the guard read %q, want the report rendered from the base without F-1, %q", got.saw, want)
	}
}

// A FOLD IS MADE ONLY OVER THE BASE IT WAS RENDERED FROM. A second base lands with another write's
// edit — the write path refuses a second ingest, and the render refuses a record that holds two — so
// the report does not render, and folding the edit onto the render made before the lock would hand the
// guard a report the record does not have.
func TestARenderBeforeTheLockIsNotFoldedWhereTheBaseMoved(t *testing.T) {
	everyInput(t)
	run, write, got := reportStage(t)
	if err := overtaken(t, run, []*Event{
		recordtest.At(t, HarnessSeat, "race:base", &recordpb.BaseIngest{Text: proto.String("# another report")}),
		recordtest.At(t, "blue-respond", "race:edit", &recordpb.BlueEdit{Old: proto.String("report"), New: proto.String("account"), EditKey: proto.String("E1")}),
	}, write); !errors.Is(err, ErrWritesNothing) {
		t.Fatalf("the write = %v, want the guard's ErrWritesNothing", err)
	}
	if got.sawErr == nil || !strings.Contains(got.sawErr.Error(), "two base_ingest") || got.folded != 0 {
		t.Errorf("the guard read %q (%v), %d folds; want the refusal of two bases and no fold", got.saw, got.sawErr, got.folded)
	}
}

// A FOLD THAT DOES NOT APPLY IS NOT THE REFUSAL. Another write's op does not replay; folded onto the
// render, the failure would name its place among the ops folded, not among the report's. The write
// renders from the base, and the guard reads the full render's failure in its own words.
func TestAFoldThatDoesNotApplyIsRenderedFromTheBase(t *testing.T) {
	everyInput(t)
	run, write, got := reportStage(t)
	if err := overtaken(t, run, []*Event{
		recordtest.At(t, "blue-respond", "race:edit", &recordpb.BlueEdit{Old: proto.String("report"), New: proto.String("boom"), EditKey: proto.String("E1")}),
	}, write); !errors.Is(err, ErrWritesNothing) {
		t.Fatalf("the write = %v, want the guard's ErrWritesNothing", err)
	}
	_, want := fullRender(t, run)
	if want == nil || got.sawErr == nil || got.sawErr.Error() != want.Error() {
		t.Fatalf("the guard read the failure %v, want the full render's, %v", got.sawErr, want)
	}
	if !strings.Contains(want.Error(), "mutation 2 of 2") || got.rendered != 1 || got.folded != 0 {
		t.Errorf("the failure %q, renders under the lock %d, folds %d; want mutation 2 of 2, 1 and 0", want, got.rendered, got.folded)
	}
}

// A GUARD THAT READS THE REPORT WITHOUT SAYING SO IS REFUSED. Its write made no render before the
// lock, so answering it would render under the lock: the declaration is CheckOnReport.
func TestAGuardThatReadsTheReportUndeclaredIsRefused(t *testing.T) {
	everyInput(t)
	run, _, _ := reportStage(t)
	_, err := Append(Identity{Run: run, SeatID: evLens}, &recordpb.Anchor{Id: proto.String("F-9"), Location: proto.String("# report")},
		Check(func(l *Landing) error {
			_, err := l.Report()
			return err
		}))
	var fe *feov.Error
	if !errors.As(err, &fe) || fe.Code != feov.Internal || !strings.Contains(fe.Error(), "CheckOnReport") {
		t.Fatalf("an undeclared read of the report = %v, want an internal refusal naming CheckOnReport", err)
	}
}

// A MINT MEASURED ON THE REPORT RENDERS IT BEFORE THE LOCK. The budget of a lens whose unit is in the
// report reads the report in the mint's validation (requireMintWithinBudget); validationReadsReport
// is what makes the write render it before it takes the lock, and the measure reads that render.
func TestAMintMeasuredOnTheReportRendersItBeforeTheLock(t *testing.T) {
	calls := everyInput(t)
	run := newStage(t).cast("red-lens-voice").ingest().register("red-lens-voice").seed()
	if _, err := Append(Identity{Run: run, SeatID: "red-lens-voice"}, highMint("G1")); err != nil {
		t.Fatalf("the mint: %v", err)
	}
	if *calls != 1 {
		t.Errorf("a voice lens's mint rendered the report %d times, want once, before the lock", *calls)
	}
	*calls = 0
	g2 := highMint("G2")
	if _, err := Append(Identity{Run: run, SeatID: evLens}, g2); err != nil {
		t.Fatalf("the evidence mint: %v", err)
	}
	if *calls != 0 {
		t.Errorf("an evidence lens's mint, measured on citations, rendered the report %d times, want none", *calls)
	}
}
