package recordtest_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE GUARD'S OWN ABSENCE MUST BE LOUD (#666).
//
// recordtest.Main catches the leak exactly — it measures a cached handle whose file is gone — but
// only in a package that declares it. A package added tomorrow with tests that open records and no
// TestMain is back to where this started: green on Linux, refused on Windows, one push and one
// round-trip later. The runtime check cannot report a package it was never installed in, and that
// silence reads exactly like a clean one.
//
// So the roster is checked rather than trusted. This is an IMPORT-GRAPH question, not a pattern
// one: a package whose test files import the record layer either declares the guard or it does
// not, and both halves are read out of the syntax tree rather than guessed from a regex. There is
// no data-flow heuristic here and so no false-positive exemption list to keep — the one list is
// the packages that touch records, and it is derived.

// guardCalls are the ways a package may declare the post-suite check. A package with its own
// TestMain passes it to testbuild.Main as a hook rather than giving up the TestMain it has;
// recordsql declares OrphanedHandles itself, because recordtest imports it and cannot be imported
// back.
var guardCalls = []string{
	"recordtest.Main",
	"recordtest.CheckOrphanedHandles",
	"OrphanedHandles",
}

// recordImports are the import paths that mean "this package's tests can open a record". Any one
// of them puts the package on the roster.
var recordImports = []string{
	"internal/record",
	"internal/record/recordsql",
	"internal/record/recordtest",
	"internal/runtest",
}

func touchesRecord(importPath string) bool {
	for _, p := range recordImports {
		if strings.HasSuffix(importPath, p) {
			return true
		}
	}
	return false
}

func TestEveryPackageWhoseTestsOpenARecordDeclaresTheHandleGuard(t *testing.T) {
	root, err := repotree.Root()
	if err != nil {
		t.Fatalf("locating the repository: %v", err)
	}
	toolsDir := filepath.Join(root, "plugins", "frank-exchange-of-views", "tools")

	// dir -> whether its test files import the record layer, and whether any declares the guard.
	type pkgState struct{ onRoster, guarded bool }
	pkgs := map[string]*pkgState{}

	fset := token.NewFileSet()
	walkErr := filepath.Walk(toolsDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, p, nil, 0)
		if perr != nil {
			return fmt.Errorf("parsing %s: %w", p, perr)
		}
		dir := filepath.Dir(p)
		st := pkgs[dir]
		if st == nil {
			st = &pkgState{}
			pkgs[dir] = st
		}
		for _, im := range f.Imports {
			path, uerr := strconv.Unquote(im.Path.Value)
			if uerr != nil {
				continue
			}
			if touchesRecord(path) {
				st.onRoster = true
			}
		}
		// The guard is NAMED, not necessarily CALLED, and the difference is not academic: a
		// package with its own TestMain declares it by PASSING it —
		// `testbuild.Main(m, recordtest.CheckOrphanedHandles)` — where it is an argument, not a
		// call target. Matching only call targets read all four of those packages as unguarded,
		// which is how this sweep's first draft reported five defects that were four false ones.
		//
		// So any reference to the name counts. A package that merely imports recordtest for
		// fixtures still does not qualify, which is the distinction being drawn: the import puts
		// it on the roster, naming the guard is what takes it off the missing list.
		ast.Inspect(f, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.Ident, *ast.SelectorExpr:
			default:
				return true
			}
			text := exprText(n.(ast.Expr))
			for _, want := range guardCalls {
				if text == want {
					st.guarded = true
				}
			}
			return true
		})
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walking the tools module: %v", walkErr)
	}

	// THE FLOOR. Every assertion below is vacuously true over an empty roster, which is the shape
	// that lets a sweep pass by traversing nothing.
	var roster, missing []string
	for dir, st := range pkgs {
		if !st.onRoster {
			continue
		}
		rel, _ := filepath.Rel(toolsDir, dir)
		roster = append(roster, rel)
		if !st.guarded {
			missing = append(missing, rel)
		}
	}
	sort.Strings(roster)
	sort.Strings(missing)

	if len(pkgs) == 0 {
		t.Fatal("the walk parsed NO test files under the tools module — the sweep measured nothing, and an empty roster reads exactly like a fully-guarded one")
	}
	if len(roster) < 2 {
		t.Fatalf("only %d package(s) were found to touch the record layer (%v) — the import check is no longer matching, so this sweep is guarding almost nothing", len(roster), roster)
	}
	t.Logf("handle guard (#666): %d package(s) whose tests open a record, %d unguarded", len(roster), len(missing))

	if len(missing) > 0 {
		t.Errorf("these packages' tests open records but declare no orphaned-handle guard: %s\n\n"+
			"Each needs a TestMain — `func TestMain(m *testing.M) { recordtest.Main(m) }` — or, if it\n"+
			"already has one, recordtest.CheckOrphanedHandles passed to testbuild.Main as a hook.\n"+
			"Without it a run directory taken from t.TempDir() instead of recordtest.TmpRun leaks a\n"+
			"cached database handle, which passes on Linux and fails the Windows leg. Nine tests have\n"+
			"made that mistake; this is what catches the tenth.", strings.Join(missing, ", "))
	}
}

// exprText renders a call target as written — `recordtest.Main` or `OrphanedHandles` — without
// resolving it. The guard names are unambiguous in this module, and a resolver here would be more
// machinery than the question needs.
func exprText(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return exprText(v.X) + "." + v.Sel.Name
	}
	return ""
}
