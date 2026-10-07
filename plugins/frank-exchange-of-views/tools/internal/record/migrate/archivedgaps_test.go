package migrate_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// EVERY ARCHIVED GAP HAS A NAMED LOCATION. Migrated, every quote gap carries a location state and
// every about gap none and its reference; in a run with a base every `marked` gap's anchor stands
// once, with prose before it in its sentence; a run with no base reads `unrendered`. The golden names
// each gap's state, so a `gone` gap — a cut its run made, or a quote that never placed — and a gap
// placed only by the fallback are listed by name and read in review.
func TestEveryArchivedGapHasANamedLocation(t *testing.T) {
	var lines []string
	for _, tb := range archiveTarballs(t) {
		run := strings.TrimSuffix(filepath.Base(tb), ".tar.gz")
		res, dst := migrateArchive(t, filepath.Base(tb))
		_, haveBase, _, err := record.ReportProjection(dst)
		if err != nil {
			t.Fatal(err)
		}
		md := ""
		if haveBase {
			if md, err = reportproj.RenderFromRecord(dst); err != nil {
				t.Fatalf("%s: %v", run, err)
			}
		}
		board, err := record.BoardJSONOfRun(dst)
		if err != nil {
			t.Fatalf("%s: %v", run, err)
		}
		line := []string{run}
		for _, g := range append(board.Open, board.Closed...) {
			if g.AboutKind != "" {
				if g.LocationState != "" || g.AboutRef == "" {
					t.Errorf("%s %s: an about gap reads state %q, about_ref %q", run, g.ID, g.LocationState, g.AboutRef)
				}
				line = append(line, g.ID+"=about")
				continue
			}
			if g.Location == "" {
				continue
			}
			switch {
			case !haveBase && g.LocationState != record.LocationUnrendered:
				t.Errorf("%s %s: a run with no base reads %q, want unrendered", run, g.ID, g.LocationState)
			case g.LocationState == record.LocationMarked:
				tok := anchor.Token(g.ID)
				at := strings.Index(md, tok)
				if strings.Count(md, tok) != 1 {
					t.Errorf("%s %s: a marked gap's anchor stands %d times", run, g.ID, strings.Count(md, tok))
				}
				for _, sp := range anchor.Sentences(md) {
					if sp[0] <= at && at < sp[1] && !claimcount.HasProse(md[sp[0]:at]) {
						t.Errorf("%s %s: a marked gap's anchor has no prose before it in its sentence %q", run, g.ID, md[sp[0]:sp[1]])
					}
				}
			case g.LocationState == "":
				t.Errorf("%s %s: a quote gap carries no location state", run, g.ID)
			}
			name := g.ID + "=" + g.LocationState
			if slices.Contains(res.GapAnchors.Fallback, g.ID) {
				name += "+fallback"
			}
			line = append(line, name)
		}
		lines = append(lines, strings.Join(line, " "))
	}
	goldentest.Assert(t, "archived_gaps", strings.Join(lines, "\n")+"\n")
}

// archiveTarballs is every archived run.
func archiveTarballs(t *testing.T) []string {
	t.Helper()
	tbs, err := filepath.Glob(filepath.Join(filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz")), "*.tar.gz"))
	if err != nil || len(tbs) == 0 {
		t.Fatalf("no archived run found: %v", err)
	}
	return tbs
}

// THE TWO GAPS THE REPLAY LOST, RESOLVED BY NAME. b9 G4's sentence was duplicated elsewhere and the
// replay followed the duplicate's edit; b5 G5's edits overlapped its quote without containing it,
// and the replay saw none. Migrated onto anchors, b9 G4 is marked on the sentence holding its minted
// text, and b5 G5's edited_since is not empty.
func TestB9AndB5GapsResolveByName(t *testing.T) {
	_, b9 := migrateArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	board, err := record.BoardJSONOfRun(b9)
	if err != nil {
		t.Fatal(err)
	}
	fam, err := record.FamilyOf(b9)
	if err != nil {
		t.Fatal(err)
	}
	var g4 *record.GapJSON
	for _, g := range append(board.Open, board.Closed...) {
		if g.ID == "G4" {
			g4 = &g
		}
	}
	if g4 == nil || g4.LocationState != record.LocationMarked || !strings.Contains(g4.Location, fam.Gap("G4").Mint.GetLocation()) {
		t.Errorf("b9 G4 is not marked on the sentence holding its minted text %q: %+v", fam.Gap("G4").Mint.GetLocation(), g4)
	}

	_, b5 := migrateArchive(t, "2026-09-11_is-91-prime-b5.tar.gz")
	reopened := 0
	for _, e := range edits(t, b5) {
		if slices.Contains(e.GetReopened(), "G5") {
			reopened++
		}
	}
	if reopened == 0 {
		t.Error("b5 G5's edited_since is empty: no migrated edit records that it changed G5's sentence")
	}
}
