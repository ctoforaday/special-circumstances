package planguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
)

var (
	tableScan = finding{label: "stmt aaaaaaaaaaaa", shape: Shape{Kind: ConstrainedScan, Name: "mint", Count: 1, Where: OnTable}}
	viewScan  = finding{label: "stmt bbbbbbbbbbbb", shape: Shape{Kind: ConstrainedScan, Name: "mint", Count: 2, Where: "via gap"}}
)

// baselineSuite is a suite whose shapes baseline holds lines, in a directory of its own.
func baselineSuite(t *testing.T, unindexed map[string]string, lines ...string) (*Suite, string) {
	t.Helper()
	dir := t.TempDir()
	body := shapesHeader + strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "shapes.golden"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return &Suite{golden: dir, unindexed: unindexed}, filepath.Join(dir, "shapes.golden")
}

func updating(t *testing.T, accept bool) {
	t.Helper()
	was := goldentest.Update
	goldentest.Update = true
	t.Cleanup(func() { goldentest.Update = was })
	if accept {
		t.Setenv(acceptViewCost, "1")
	} else {
		t.Setenv(acceptViewCost, "")
	}
}

func written(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mentions(errs []error, part string) bool {
	for _, e := range errs {
		if strings.Contains(e.Error(), part) {
			return true
		}
	}
	return false
}

// A MISSING INDEX IS NEVER BASELINED: not by listing it, not by regenerating, not even with the
// view-cost flag. The refusal names the fix.
func TestAWalkOfTheStatementsOwnTableIsRefusedWhateverTheBaselineSays(t *testing.T) {
	s, _ := baselineSuite(t, nil, viewScan.line())
	errs := s.compareShapes([]finding{tableScan, viewScan})
	if !mentions(errs, "QueryIndexDDL") || !mentions(errs, tableScan.line()) {
		t.Errorf("an unlisted walk of the statement's own table was not refused with its fix: %v", errs)
	}

	s, _ = baselineSuite(t, nil, viewScan.line(), tableScan.line())
	if errs := s.compareShapes([]finding{tableScan, viewScan}); !mentions(errs, "no baseline may") {
		t.Errorf("a baseline listing a walk of the statement's own table was read as permission: %v", errs)
	}

	updating(t, true)
	s, path := baselineSuite(t, nil, viewScan.line())
	errs = s.compareShapes([]finding{tableScan, viewScan})
	if !mentions(errs, "QueryIndexDDL") {
		t.Errorf("regenerating accepted a missing index: %v", errs)
	}
	if strings.Contains(written(t, path), tableScan.line()) {
		t.Errorf("regenerating wrote a missing index into the baseline:\n%s", written(t, path))
	}
}

// A baseline line marked as the statement's own table fails on read, even where nothing finds it.
func TestABaselineLineOnTheStatementsOwnTableFailsOnRead(t *testing.T) {
	s, _ := baselineSuite(t, nil, tableScan.line())
	if errs := s.compareShapes(nil); !mentions(errs, "no baseline may") {
		t.Errorf("a missing index in the baseline read clean: %v", errs)
	}
}

// THE ALLOWANCE IS THE ONE WAY PAST, AND IT CANNOT OUTLIVE ITS FINDING.
func TestAnUnindexedAllowanceAdmitsItsFindingAndFailsWhenStale(t *testing.T) {
	allow := map[string]string{tableScan.line(): "the reason"}
	s, _ := baselineSuite(t, allow, viewScan.line())
	if errs := s.compareShapes([]finding{tableScan, viewScan}); len(errs) != 0 {
		t.Errorf("an allowed finding was refused: %v", errs)
	}
	s, _ = baselineSuite(t, allow, viewScan.line())
	if errs := s.compareShapes([]finding{viewScan}); !mentions(errs, "names no refused finding") {
		t.Errorf("a stale allowance passed: %v", errs)
	}
}

// A VIEW'S COST SHRINKS UNDER UPDATE_GOLDENS ALONE AND GROWS ONLY WHEN ASKED BY NAME.
func TestRegeneratingShrinksTheBaselineAndGrowsItOnlyWhenAsked(t *testing.T) {
	smaller := viewScan
	smaller.shape.Count = 1
	updating(t, false)
	s, path := baselineSuite(t, nil, viewScan.line())
	if errs := s.compareShapes([]finding{smaller}); len(errs) != 0 {
		t.Errorf("a falling count was refused: %v", errs)
	}
	if !strings.Contains(written(t, path), smaller.line()) {
		t.Errorf("a falling count was not written:\n%s", written(t, path))
	}

	s, path = baselineSuite(t, nil)
	if errs := s.compareShapes([]finding{viewScan}); !mentions(errs, acceptViewCost) {
		t.Errorf("UPDATE_GOLDENS alone grew the baseline: %v", errs)
	}
	if strings.Contains(written(t, path), viewScan.line()) {
		t.Errorf("UPDATE_GOLDENS alone wrote a new line:\n%s", written(t, path))
	}

	updating(t, true)
	s, path = baselineSuite(t, nil)
	if errs := s.compareShapes([]finding{viewScan}); len(errs) != 0 {
		t.Errorf("an accepted view cost was refused: %v", errs)
	}
	if !strings.Contains(written(t, path), viewScan.line()) {
		t.Errorf("an accepted view cost was not written:\n%s", written(t, path))
	}
}

// A count that rises is growth, and a normal run says so; one that falls asks for the shrink.
func TestACountThatMovesIsReadAsGrowthOrShrink(t *testing.T) {
	larger := viewScan
	larger.shape.Count = 3
	smaller := viewScan
	smaller.shape.Count = 1
	s, _ := baselineSuite(t, nil, viewScan.line())
	if errs := s.compareShapes([]finding{larger}); !mentions(errs, "NEW plan-shape finding") {
		t.Errorf("a rising count passed: %v", errs)
	}
	if errs := s.compareShapes([]finding{smaller}); !mentions(errs, "SHRANK") {
		t.Errorf("a falling count was not reported as a shrink: %v", errs)
	}
}
