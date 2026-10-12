package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/ctoforaday/special-circumstances/scripts/internal/racescope"
)

// start is one place a goroutine begins: a `go` statement, or a call to a row of the table.
type start struct {
	pkg  string // import path of the package the file compiles into
	file string // module-relative, slash-separated
	line int
	col  int
	fn   string // enclosing declaration; a method is Recv.Name
	what string // "go statement", or the table row's callee
}

func (s start) String() string {
	return fmt.Sprintf("%s:%d (%s, in %s)", s.file, s.line, s.what, s.fn)
}

// owner is the directive's spelling of where a start lives.
func (s start) owner() string { return s.file + ":" + s.fn }

// selected is one test in the race set and the first reason it is there.
type selected struct {
	name string
	why  string
}

// derived is everything the two runs need, computed from the tree and nothing else.
type derived struct {
	module string
	// sets is the race set: test package import path -> the tests selected in it.
	sets map[string][]selected
	// links is, per test package, the module-local packages its test binary links — the
	// evidence run's -coverpkg.
	links map[string][]string
	// owed are the starts in non-test files that some raced test must execute.
	owed []start
	// problems are refusals found before anything runs: a shipped start no test declares, a
	// directive naming nothing, a start no raced binary links.
	problems []string
}

// decl is one function declaration, keyed by the position of its name so the same function
// seen through two package variants (P and P [P.test]) is one node.
type decl struct {
	name     string
	file     string // module-relative
	testFile bool
	inScope  bool   // its package is in the shipped scope
	testPkg  string // for a declaration in a _test.go file: the package `go test` names
	starts   []start
	callees  []string
	covers   []string
	root     bool // a Test, Fuzz or Example function
}

const directive = "//racescope:covers "

var reRoot = regexp.MustCompile(`^(Test|Fuzz|Example)($|[^a-z])`)

// derive loads every package of the module with its tests and computes the race set.
//
// The module is loaded with type information because the two questions asked — "is this call a
// row of the table" and "which declaration does this call reach" — are questions about what a
// name RESOLVES to. Text cannot answer them: `srv.Serve` is a start on an *http.Server and
// nothing at all on a type of ours that happens to have a Serve method.
func derive(moduleDir, target string) (*derived, error) {
	module, scopePkgs, err := racescope.Scope(moduleDir, target)
	if err != nil {
		return nil, err
	}
	inScope := map[string]bool{}
	for _, p := range scopePkgs {
		inScope[p] = true
	}
	absModule, err := filepath.Abs(moduleDir)
	if err != nil {
		return nil, err
	}
	if resolved, err := filepath.EvalSymlinks(absModule); err == nil {
		absModule = resolved
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedForTest,
		Dir:   moduleDir,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", moduleDir, err)
	}
	// A package that did not type-check resolves some calls to nothing, and a call that
	// resolves to nothing is a start the set silently lacks.
	var loadErrs []string
	for _, p := range pkgs {
		for _, e := range p.Errors {
			loadErrs = append(loadErrs, p.ID+": "+e.Error())
		}
	}
	if len(loadErrs) > 0 {
		sort.Strings(loadErrs)
		return nil, fmt.Errorf("the module does not load cleanly, so the race set is not measured:\n  %s",
			strings.Join(loadErrs, "\n  "))
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].ID < pkgs[j].ID })

	local := func(path string) bool { return path == module || strings.HasPrefix(path, module+"/") }
	rel := func(filename string) (string, bool) {
		if resolved, err := filepath.EvalSymlinks(filename); err == nil {
			filename = resolved
		}
		r, err := filepath.Rel(absModule, filename)
		if err != nil || strings.HasPrefix(r, "..") {
			return "", false
		}
		return filepath.ToSlash(r), true
	}

	decls := map[string]*decl{}
	seenFile := map[string]bool{}
	for _, p := range pkgs {
		// `go test` names the package under test; an external test package (P_test) and the
		// in-package variant both belong to P.
		home := p.PkgPath
		if p.ForTest != "" {
			home = p.ForTest
		}
		for _, f := range p.Syntax {
			filename := p.Fset.Position(f.Package).Filename
			file, ok := rel(filename)
			if !ok || seenFile[file] {
				continue // generated in the build cache (a test main, cgo output), or seen in another variant
			}
			seenFile[file] = true
			testFile := strings.HasSuffix(file, "_test.go")
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok || fd.Body == nil {
					continue
				}
				n := &decl{name: declName(fd), file: file, testFile: testFile, inScope: inScope[home]}
				if testFile {
					n.testPkg = home
					n.root = fd.Recv == nil && reRoot.MatchString(fd.Name.Name) && fd.Name.Name != "TestMain"
					if fd.Doc != nil {
						for _, c := range fd.Doc.List {
							if strings.HasPrefix(c.Text, directive) {
								n.covers = append(n.covers, strings.TrimSpace(strings.TrimPrefix(c.Text, directive)))
							}
						}
					}
				}
				// A function literal's body is walked as part of the declaration enclosing it,
				// so a `go` statement inside a closure, a t.Run body or a handler literal
				// belongs to that declaration.
				ast.Inspect(fd.Body, func(node ast.Node) bool {
					switch x := node.(type) {
					case *ast.GoStmt:
						n.starts = append(n.starts, startAt(p.Fset, x.Pos(), home, file, n.name, "go statement"))
					case *ast.CallExpr:
						fn, ok := typeutil.Callee(p.TypesInfo, x).(*types.Func)
						if !ok {
							return true // a function value, a conversion, a builtin
						}
						fn = fn.Origin()
						if row, ok := racescope.Lookup(fn.FullName()); ok {
							n.starts = append(n.starts, startAt(p.Fset, x.Pos(), home, file, n.name, row.Callee))
							return true
						}
						if sig, ok := fn.Type().(*types.Signature); ok && sig.Recv() != nil && types.IsInterface(sig.Recv().Type()) {
							return true // dispatched at run time; not an edge
						}
						if fn.Pkg() != nil && local(fn.Pkg().Path()) {
							n.callees = append(n.callees, p.Fset.Position(fn.Pos()).String())
						}
					}
					return true
				})
				decls[p.Fset.Position(fd.Name.Pos()).String()] = n
			}
		}
	}

	d := &derived{module: module, sets: map[string][]selected{}, links: map[string][]string{}}

	// followable is where reachability may go: test files anywhere in the module, and the
	// non-test files of packages OUTSIDE the shipped scope (the test-support packages).
	// It stops at shipped source: every test that builds the root command reaches every verb,
	// so a graph that entered source would select the whole of a package and the scoping
	// would be gone. A start in shipped source is owned by a directive instead.
	followable := func(n *decl) bool { return n.testFile || !n.inScope }

	declared := map[string][]string{} // owner -> tests declaring it
	var keys []string
	for k := range decls {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		root := decls[k]
		if len(root.covers) > 0 && !(root.root && inScope[root.testPkg]) {
			d.problems = append(d.problems, fmt.Sprintf("%s: %s carries a racescope:covers directive and is not a test of a package in the race scope",
				root.file, root.name))
			continue
		}
		if !root.root || !inScope[root.testPkg] {
			continue
		}
		why := ""
		for _, c := range root.covers {
			declared[c] = append(declared[c], root.name)
			if why == "" {
				why = "declares " + c
			}
		}
		if why == "" {
			seen := map[string]bool{k: true}
			queue := []string{k}
			for len(queue) > 0 && why == "" {
				n := decls[queue[0]]
				queue = queue[1:]
				if len(n.starts) > 0 {
					why = n.starts[0].String()
					break
				}
				for _, c := range n.callees {
					if next, ok := decls[c]; ok && !seen[c] && followable(next) {
						seen[c] = true
						queue = append(queue, c)
					}
				}
			}
		}
		if why != "" {
			d.sets[root.testPkg] = append(d.sets[root.testPkg], selected{name: root.name, why: why})
		}
	}
	for pkg := range d.sets {
		sort.Slice(d.sets[pkg], func(i, j int) bool { return d.sets[pkg][i].name < d.sets[pkg][j].name })
	}

	// What each selected package's test binary links, walked over the loaded import graph.
	linked := map[string]bool{}
	for _, p := range pkgs {
		if p.ForTest == "" || len(d.sets[p.ForTest]) == 0 {
			continue
		}
		seen := map[string]bool{}
		var walk func(q *packages.Package)
		walk = func(q *packages.Package) {
			if seen[q.ID] {
				return
			}
			seen[q.ID] = true
			for _, imp := range q.Imports {
				if local(imp.PkgPath) {
					walk(imp)
				}
			}
		}
		walk(p)
		have := map[string]bool{}
		for _, l := range d.links[p.ForTest] {
			have[l] = true
		}
		for _, q := range pkgs {
			// An external test package's own path (P_test) is not a package -coverpkg can name.
			if seen[q.ID] && !have[q.PkgPath] && !strings.HasSuffix(q.PkgPath, "_test") && !strings.HasSuffix(q.PkgPath, ".test") {
				have[q.PkgPath] = true
				d.links[p.ForTest] = append(d.links[p.ForTest], q.PkgPath)
				linked[q.PkgPath] = true
			}
		}
	}
	for pkg := range d.links {
		sort.Strings(d.links[pkg])
	}

	// The starts some raced test must execute: every one in a non-test file of the shipped
	// scope, and every one in a non-test file of a support package a raced binary links.
	owners := map[string]bool{}
	for _, k := range keys {
		n := decls[k]
		if n.testFile {
			continue
		}
		for _, s := range n.starts {
			if !n.inScope && !linked[s.pkg] {
				continue
			}
			d.owed = append(d.owed, s)
			owners[s.owner()] = true
			if !linked[s.pkg] {
				d.problems = append(d.problems, fmt.Sprintf("%s: no raced test binary links %s, so nothing executes this start under the detector", s, s.pkg))
			}
			if n.inScope && len(declared[s.owner()]) == 0 {
				d.problems = append(d.problems, fmt.Sprintf("%s: a start in shipped source that no test declares — put `%s%s` above the test that runs it",
					s, directive, s.owner()))
			}
		}
	}
	sort.Slice(d.owed, func(i, j int) bool {
		if d.owed[i].file != d.owed[j].file {
			return d.owed[i].file < d.owed[j].file
		}
		return d.owed[i].line < d.owed[j].line
	})
	var names []string
	for c := range declared {
		names = append(names, c)
	}
	sort.Strings(names)
	for _, c := range names {
		if !owners[c] {
			d.problems = append(d.problems, fmt.Sprintf("%s declares it covers %s, and no goroutine starts there",
				strings.Join(declared[c], ", "), c))
		}
	}
	sort.Strings(d.problems)
	return d, nil
}

func startAt(fset *token.FileSet, pos token.Pos, pkg, file, fn, what string) start {
	p := fset.Position(pos)
	return start{pkg: pkg, file: file, line: p.Line, col: p.Column, fn: fn, what: what}
}

// declName is the directive's spelling of a declaration: Name, or Recv.Name for a method.
func declName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	t := fd.Recv.List[0].Type
	for {
		switch x := t.(type) {
		case *ast.StarExpr:
			t = x.X
			continue
		case *ast.IndexExpr:
			t = x.X
			continue
		case *ast.IndexListExpr:
			t = x.X
			continue
		case *ast.ParenExpr:
			t = x.X
			continue
		case *ast.Ident:
			return x.Name + "." + fd.Name.Name
		}
		return fd.Name.Name
	}
}
