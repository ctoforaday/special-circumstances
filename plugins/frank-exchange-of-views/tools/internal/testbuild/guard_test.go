package testbuild

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// EVERY PACKAGE THAT BUILDS MUST ALSO CLEAN UP, and the obligation is checked rather than
// remembered.
//
// Main is the only place the shared build directory can be removed, and nothing about
// calling Binary forces a package to route through it. Five packages were calling Binary and
// none had a TestMain, so one `go test ./...` of this module left five directories holding a
// linked binary apiece in TMPDIR — 186 MB, on a green run, into a 2 GB tmpfs. A sixth
// package added tomorrow would rebuild that leak in silence, because the only thing standing
// between the two states was that somebody remembered.
//
// The scan is over the SYNTAX TREE, not over the source text: `grep "testbuild.Main"` reports
// the same clean board when the convention is honoured and when the walk itself broke, and
// this check exists precisely to catch the case nobody is watching for.
func TestEveryPackageThatBuildsCleansUp(t *testing.T) {
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	self, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}

	builds, cleans := map[string]bool{}, map[string]bool{}
	fset := token.NewFileSet()
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "testdata", "vendor", "node_modules":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, p, nil, 0)
		if perr != nil {
			return perr
		}
		dir := filepath.Dir(p)
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := named(call.Fun); fn {
			case "Binary":
				builds[dir] = true
			case "Main":
				// Only inside TestMain: `Main` is a common enough name that a bare call
				// elsewhere is not evidence the binary's exit was wired to Cleanup.
				if enclosingTestMain(f, call.Pos()) {
					cleans[dir] = true
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// A SCAN THAT FOUND NOTHING IS NOT A CLEAN BOARD. This package's own tests call Binary,
	// so its directory must be in the set on every healthy run; if it is not, the walk or the
	// match broke and every other verdict below is a coin flip reported as a measurement.
	if !builds[self] {
		t.Fatalf("the scan did not find this package's own Binary calls (%s) — it is broken, "+
			"and a broken scan reports the same empty result as a compliant tree", self)
	}

	var missing []string
	for dir := range builds {
		if !cleans[dir] {
			rel, _ := filepath.Rel(root, dir)
			missing = append(missing, rel)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("these packages call testbuild.Binary and never remove what it built:\n  %s\n"+
			"Add to each:\n\n  func TestMain(m *testing.M) { testbuild.Main(m) }\n\n"+
			"Without it the package leaks one directory holding one linked binary per run into "+
			"TMPDIR, which is a tmpfs in CI and in the record-run validation loop.",
			strings.Join(missing, "\n  "))
	}
}

// named returns the identifier a call names — `Binary` for both `Binary(...)` and
// `testbuild.Binary(...)`, so the check reads the same inside this package and outside it.
func named(e ast.Expr) string {
	switch f := e.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		if x, ok := f.X.(*ast.Ident); ok && x.Name == "testbuild" {
			return f.Sel.Name
		}
	}
	return ""
}

// enclosingTestMain reports whether pos falls inside this file's TestMain.
func enclosingTestMain(f *ast.File, pos token.Pos) bool {
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "TestMain" || fn.Recv != nil {
			continue
		}
		if pos > fn.Pos() && pos < fn.End() {
			return true
		}
	}
	return false
}
