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
func TestEveryDriverImporterLinksTheLockSwitch(t *testing.T) {
	sources, err := repotree.GoSources("plugins", "frank-exchange-of-views", "tools")
	if err != nil {
		t.Fatalf("locating the tool's sources: %v", err)
	}
	if len(sources) == 0 {
		t.Fatal("no Go sources found — this gate measured nothing")
	}
	imports := map[string]map[string]bool{} // package directory -> import paths of its files
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
	var drivers, missing []string
	for dir, set := range imports {
		if !set[sqliteDriver] {
			continue
		}
		drivers = append(drivers, dir)
		if !set[lockSwitch] && !strings.HasSuffix(filepath.ToSlash(dir), "internal/record/recordsql") {
			missing = append(missing, dir)
		}
	}
	if len(drivers) == 0 {
		t.Fatalf("no shipped package imports %s — this gate measured nothing", sqliteDriver)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("shipped packages import %s without importing recordsql: %s\n\n"+
			"recordsql's init sets SQLite's lock mode to OFD locks, process-wide, before the first open. "+
			"A binary built from these packages without recordsql opens its databases with POSIX locks, which "+
			"any close of any descriptor on the file drops. Import recordsql (its ofd.go says why).",
			sqliteDriver, strings.Join(missing, ", "))
	}
}
