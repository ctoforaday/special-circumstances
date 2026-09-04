package surface

import (
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

// A HAND-COUNTED `..` IS A SECOND COPY OF WHERE THE FILE LIVES, and this is the sweep that keeps
// the copies from coming back.
//
// Before repotree, sixteen sites across seven packages reached the plugin tree by counting `..`
// from their own directory — three of them naming the SAME directory (agents/) at three different
// depths, each correct only for its own launch point. The depth was a property of where the file
// happened to sit, so moving a file between packages silently repointed it. At nothing. And the
// gates most likely to be moved were the ones that assert a NEGATIVE over what they discover,
// where pointing at nothing reads exactly like a clean tree.
//
// WHY A GUARD RATHER THAN GENERATION. There is nothing to generate: these are Go source files
// naming a directory, and the fix is an accessor, not a build step. What the guard holds is that
// the accessor is USED — a rule with one exemption, stated in code below rather than kept in a
// list somewhere that has to be maintained alongside it.
//
// WHAT IT DOES NOT CATCH, said plainly rather than left implied: a single `..` or a pair of them.
// Those are ordinary intra-package navigation (testdata/, a sibling package) and flagging them
// would make this a nuisance that gets suppressed. Three is where "I am walking out of the module"
// starts, and every site this replaced used three or more.
func TestNoSourceCountsItsOwnDepthToTheRepository(t *testing.T) {
	sources, err := repotree.ToolSources()
	if err != nil {
		t.Fatal(err)
	}
	// Tests are the majority of the offenders, so the sweep has to include them — ToolSources
	// skips _test.go for the OTHER gates' reasons, and this one needs them back.
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	toolInternal := filepath.Join(root, "plugins", "frank-exchange-of-views", "tools")
	var all []string
	err = filepath.Walk(toolInternal, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(p, "_test.go") {
			all = append(all, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	all = append(all, sources...)
	if len(all) < 50 {
		t.Fatalf("the sweep found only %d file(s) — too few to be this module, and a sweep that reads "+
			"nothing reports every file clean", len(all))
	}

	var hits []string
	for _, path := range all {
		// repotree IS the accessor. It resolves the root by SEARCHING for it rather than by
		// counting, so it has no depth to hard-code — and its own tests navigate deliberately.
		if strings.Contains(filepath.ToSlash(path), "/internal/repotree/") {
			continue
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			t.Fatal(perr)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isFilepathJoin(call.Fun) {
				return true
			}
			if n := countLeadingParentArgs(call.Args); n >= 3 {
				pos := fset.Position(call.Pos())
				hits = append(hits, filepath.Base(path)+":"+strconv.Itoa(pos.Line)+
					" joins "+strconv.Itoa(n)+" parent segments")
			}
			return true
		})
	}
	sort.Strings(hits)
	if len(hits) > 0 {
		t.Errorf("%d site(s) count their own depth to the repository:\n  %s\n\n"+
			"Use internal/repotree — Root(), Plugin(...), Constitutions(), ToolSources(), DebateJS() —\n"+
			"which finds the tree by looking for it. A counted depth is correct only from the directory\n"+
			"it was counted in, and a gate that asserts a negative over what it discovers cannot tell a\n"+
			"repointed path from a clean tree.",
			len(hits), strings.Join(hits, "\n  "))
	}
}

func isFilepathJoin(fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Join" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "filepath"
}

// countLeadingParentArgs counts the ".." string arguments at the head of a Join call. Leading,
// because that is what walking OUT of the tree looks like; a ".." in the middle of a path is
// something else and is not this gate's business.
func countLeadingParentArgs(args []ast.Expr) int {
	n := 0
	for _, a := range args {
		lit, ok := a.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || lit.Value != `".."` {
			// Skip a non-".." first argument (a variable holding a base directory) and keep
			// looking, so `filepath.Join(wd, "..", "..", "..")` is still counted.
			if n == 0 {
				continue
			}
			break
		}
		n++
	}
	return n
}
