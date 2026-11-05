package gcpolicy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
)

// state reads the GC percent and memory limit without changing either.
func state() (percent int, limit int64) {
	percent = debug.SetGCPercent(100)
	debug.SetGCPercent(percent)
	return percent, debug.SetMemoryLimit(-1)
}

// isolate restores the process's GC settings after the test and clears the operator variables,
// so each case starts from the runtime's defaults whatever the shell running the suite set.
func isolate(t *testing.T) {
	t.Helper()
	p, l := state()
	t.Cleanup(func() { debug.SetGCPercent(p); debug.SetMemoryLimit(l) })
	t.Setenv("GOGC", "")
	t.Setenv("GOMEMLIMIT", "")
	debug.SetGCPercent(100)
	debug.SetMemoryLimit(math.MaxInt64)
}

func TestShortTurnsTheGCOffUnderTheCap(t *testing.T) {
	isolate(t)
	Short()
	if p, l := state(); p != -1 || l != Cap {
		t.Fatalf("after Short: GC percent %d, limit %d; want -1, %d", p, l, Cap)
	}
}

func TestLongRestoresTheDefaultAfterShort(t *testing.T) {
	isolate(t)
	Short()
	Long()
	if p, l := state(); p != 100 || l != math.MaxInt64 {
		t.Fatalf("after Short then Long: GC percent %d, limit %d; want 100, %d", p, l, int64(math.MaxInt64))
	}
}

// An operator's GOGC or GOMEMLIMIT has already been applied by the runtime when main runs; neither
// function may overwrite it. The settings below stand in for what the runtime applied.
func TestAnOperatorSettingIsLeftAlone(t *testing.T) {
	for _, v := range []string{"GOGC", "GOMEMLIMIT"} {
		for name, call := range map[string]func(){"Short": Short, "Long": Long} {
			t.Run(v+"/"+name, func(t *testing.T) {
				isolate(t)
				t.Setenv(v, "50")
				debug.SetGCPercent(50)
				debug.SetMemoryLimit(1 << 30)
				call()
				if p, l := state(); p != 50 || l != 1<<30 {
					t.Fatalf("%s with %s set changed the GC: percent %d, limit %d", name, v, p, l)
				}
			})
		}
	}
}

func TestAnEmptyVariableIsNotAnOperatorSetting(t *testing.T) {
	isolate(t)
	Short()
	if p, _ := state(); p != -1 {
		t.Fatalf("empty GOGC and GOMEMLIMIT kept Short from running: GC percent %d", p)
	}
}

// TestEveryShippedMainStatesItsGCClass derives the shipped binaries from the tree — every directory
// under plugins/*/tools/cmd, which is what the release job and fetch-bin.sh ship — and requires the
// first statement of each one's main to be its own module's gcpolicy.Short() or gcpolicy.Long().
// First, because anything main does before it runs under the other policy.
func TestEveryShippedMainStatesItsGCClass(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	mods, err := filepath.Glob(filepath.Join(root, "plugins", "*", "tools", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	var mains int
	for _, gomod := range mods {
		want := "github.com/ctoforaday/special-circumstances/shared/gcpolicy"
		cmds, err := filepath.Glob(filepath.Join(filepath.Dir(gomod), "cmd", "*"))
		if err != nil {
			t.Fatal(err)
		}
		for _, dir := range cmds {
			if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
				continue
			}
			rel, _ := filepath.Rel(root, dir)
			fn, file := findMain(t, dir)
			if fn == nil {
				t.Errorf("%s: no func main", rel)
				continue
			}
			mains++
			alias := importName(file, want)
			if alias == "" {
				t.Errorf("%s: main does not import %s", rel, want)
				continue
			}
			if got := firstCall(fn); got != alias+".Short" && got != alias+".Long" {
				t.Errorf("%s: main's first statement is %q, want %s.Short() or %s.Long()", rel, got, alias, alias)
			}
		}
	}
	// A glob that matched nothing would pass every assertion above.
	if len(mods) < 3 || mains < 20 {
		t.Fatalf("found %d plugin modules and %d mains — the tree walk is not seeing the shipped binaries", len(mods), mains)
	}
}

func findMain(t *testing.T, dir string) (*ast.FuncDecl, *ast.File) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), f, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range file.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "main" && file.Name.Name == "main" {
				return fn, file
			}
		}
	}
	return nil, nil
}

// importName is the name path is imported under in file, or "" when it is not imported.
func importName(file *ast.File, path string) string {
	for _, imp := range file.Imports {
		if p, _ := strconv.Unquote(imp.Path.Value); p == path {
			if imp.Name != nil {
				return imp.Name.Name
			}
			return filepath.Base(path)
		}
	}
	return ""
}

// firstCall renders main's first statement as pkg.Func when it is a bare call of one, else "".
func firstCall(fn *ast.FuncDecl) string {
	if fn.Body == nil || len(fn.Body.List) == 0 {
		return ""
	}
	es, ok := fn.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return ""
	}
	call, ok := es.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return ""
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return x.Name + "." + sel.Sel.Name
}
