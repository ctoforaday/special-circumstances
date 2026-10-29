package surface

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

const (
	sqliteDriver = "modernc.org/sqlite"
	lockSwitch   = "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EVERY PACKAGE THAT OPENS A DATABASE LINKS THE LOCK SWITCH. recordsql's init sets SQLite's
// process-wide lock mode to OFD locks before any open (ofd.go); a package that imports the driver
// without importing recordsql could stand in a binary that never runs that init, and its opens
// would take POSIX locks that a stray close of the record drops. So a shipped package importing the
// driver must import recordsql too, which puts the init in its binary and ahead of its own.
//
// A TEST BINARY IS HELD TO IT TOO. A package's test binary is its non-test and _test.go files
// together (package x and x_test), so a directory whose test files import the driver must import
// recordsql in one of its files; otherwise its tests open databases with POSIX locks and measure a
// record no shipped binary holds.
func TestEveryDriverImporterLinksTheLockSwitch(t *testing.T) {
	shipped := importsByDir(t, repotree.GoSources)
	tests := importsByDir(t, repotree.GoTestSources)
	isSwitch := func(dir string) bool { return strings.HasSuffix(filepath.ToSlash(dir), "internal/record/recordsql") }

	var drivers, missing, missingTests []string
	for dir, set := range shipped {
		if !set[sqliteDriver] {
			continue
		}
		drivers = append(drivers, dir)
		if !set[lockSwitch] && !isSwitch(dir) {
			missing = append(missing, dir)
		}
	}
	var testDrivers []string
	for dir, set := range tests {
		if !set[sqliteDriver] {
			continue
		}
		testDrivers = append(testDrivers, dir)
		if !set[lockSwitch] && !shipped[dir][lockSwitch] && !isSwitch(dir) {
			missingTests = append(missingTests, dir)
		}
	}
	if len(drivers) == 0 {
		t.Fatalf("no shipped package imports %s — this gate measured nothing", sqliteDriver)
	}
	if len(testDrivers) == 0 {
		t.Fatalf("no test file imports %s — the test half of this gate measured nothing", sqliteDriver)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("shipped packages import %s without importing recordsql: %s\n\n"+
			"recordsql's init sets SQLite's lock mode to OFD locks, process-wide, before the first open. "+
			"A binary built from these packages without recordsql opens its databases with POSIX locks, which "+
			"any close of any descriptor on the file drops. Import recordsql (its ofd.go says why).",
			sqliteDriver, strings.Join(missing, ", "))
	}
	sort.Strings(missingTests)
	if len(missingTests) > 0 {
		t.Errorf("test files import %s in packages none of whose files import recordsql: %s\n\n"+
			"Their test binaries never run recordsql's init, so their databases lock POSIX while every "+
			"shipped binary locks the record OFD. Import recordsql in the test (its ofd.go says why).",
			sqliteDriver, strings.Join(missingTests, ", "))
	}
}

// importsByDir parses the imports of every file list returns under FEOV's tools tree, keyed by
// package directory.
func importsByDir(t *testing.T, list func(...string) ([]string, error)) map[string]map[string]bool {
	t.Helper()
	sources, err := list("plugins", "frank-exchange-of-views", "tools")
	if err != nil {
		t.Fatalf("locating the tool's sources: %v", err)
	}
	imports := map[string]map[string]bool{}
	fset := token.NewFileSet()
	for _, path := range sources {
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			t.Errorf("parsing %s: %v", path, perr)
			continue
		}
		dir := filepath.Dir(path)
		if imports[dir] == nil {
			imports[dir] = map[string]bool{}
		}
		for _, spec := range f.Imports {
			p, _ := strconv.Unquote(spec.Path.Value)
			imports[dir][p] = true
		}
	}
	return imports
}
