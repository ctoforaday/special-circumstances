package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// feovTools is FEOV's Go module, relative to the repository root.
const feovTools = "plugins/frank-exchange-of-views/tools"

// The migration is confined to one binary. migratePkg is the package that reads the record shapes
// every other reader refuses and replays them through the write path; migrateBinary is the one
// directory under cmd/ that links it. Every OTHER directory under cmd/ ships too — the release job,
// fetch-bin.sh and sc-doctor treat each as a binary — and one that links migratePkg is a seat-side
// binary that can replay history.
const (
	migratePkg    = "internal/record/migrate"
	migrateBinary = "feov-migrate"
	recordPkg     = "internal/record"
)

// migrateOnly is the directive that marks a declaration of package record as the migration's alone:
// the replay writer, which writes an archived act without the live write's act judgements. A
// reference to a marked name from any package but migratePkg is a finding. The mark sits on the
// declaration, so a new migrate-only name is gated by being written, never by a list kept here.
const migrateOnly = "//pluginparity:migrate-only"

// shippedPackage is one package of a shipped binary's import graph, as `go list` reports it.
type shippedPackage struct {
	path, name, dir string
	files, imports  []string
	deps            map[string]bool
}

// listShipped asks the go command for every package in the import graph of every directory under
// cmd/, with each package's own files, imports and transitive dependencies. Test files are not in
// a package's GoFiles, and a test-only import is in neither Imports nor Deps — what is listed is
// exactly what a release build links.
func listShipped(tools string) (module string, pkgs []shippedPackage, err error) {
	return listPackages(tools, "-deps", "./cmd/...")
}

// listModule asks the go command for every package of the module, with each package's own files
// and imports: what any of its builds links, a dev command's included. Test files are not listed.
func listModule(tools string) (module string, pkgs []shippedPackage, err error) {
	return listPackages(tools, "./...")
}

// listPackages runs `go list` with args over tools and keeps the module's own packages.
func listPackages(tools string, args ...string) (module string, pkgs []shippedPackage, err error) {
	run := func(args ...string) (string, error) {
		c := exec.Command("go", args...)
		c.Dir = tools
		c.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=")
		out, err := c.Output()
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("go %s: %v\n%s", strings.Join(args, " "), err, ee.Stderr)
		}
		return string(out), err
	}
	m, err := run("list", "-m")
	if err != nil {
		return "", nil, err
	}
	module = strings.TrimSpace(m)
	const sep = "\x1f"
	listArgs := []string{"list", "-f",
		"{{.ImportPath}}" + sep + "{{.Name}}" + sep + "{{.Dir}}" + sep + `{{join .GoFiles " "}}` + sep + `{{join .Imports " "}}` + sep + `{{join .Deps " "}}`}
	out, err := run(append(listArgs, args...)...)
	if err != nil {
		return "", nil, err
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Split(line, sep)
		if len(f) != 6 || !(f[0] == module || strings.HasPrefix(f[0], module+"/")) {
			continue
		}
		p := shippedPackage{path: f[0], name: f[1], dir: f[2], files: strings.Fields(f[3]),
			imports: strings.Fields(f[4]), deps: map[string]bool{}}
		for _, d := range strings.Fields(f[5]) {
			p.deps[d] = true
		}
		pkgs = append(pkgs, p)
	}
	return module, pkgs, nil
}

// migrateConfinementProblems holds FEOV's migration to its own binary, by three rules read from the
// module's import graph — `go list`, never a hand-kept list of binaries.
//
// THE PACKAGE: no directory under cmd/ but feov-migrate has internal/record/migrate in its import
// graph. feov-migrate must have it, so the exemption names a binary that exists and links what it
// is exempted for.
//
// THE MODE: no package-level variable of package record is assigned from outside package record,
// in a package any shipped binary links. A variable of record that another package sets is a mode of the
// write path that every binary's copy of the write path reads, and it is set by code the write path
// cannot see: the value goes to the call that needs it.
//
// THE WRITER: a declaration of record marked migrateOnly — the replay writer — is named by no
// package's non-test code but migratePkg's, a dev command's or a test helper's included.
//
// A graph that lists no binary, a record package with no variables, or one with no marked
// declaration, is refused rather than passed: a rule that read nothing reports what a clean tree
// reports.
func migrateConfinementProblems(tools string) []string {
	module, pkgs, err := listShipped(tools)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot list the shipped binaries' import graph, so the migrate confinement was not checked: %v", feovTools, err)}
	}
	var problems []string
	add := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }
	rel := func(p string) string { return strings.TrimPrefix(strings.TrimPrefix(p, module), "/") }

	migrate := module + "/" + migratePkg
	var binaries []string
	for _, p := range pkgs {
		if p.name != "main" || !strings.HasPrefix(rel(p.path), "cmd/") {
			continue
		}
		bin := strings.TrimPrefix(rel(p.path), "cmd/")
		binaries = append(binaries, bin)
		switch links := p.deps[migrate]; {
		case bin == migrateBinary && !links:
			add("%s/cmd/%s does not link %s — the binary the migration is confined to no longer holds it, so the confinement exempts nothing.", feovTools, bin, migratePkg)
		case bin != migrateBinary && links:
			var via []string
			for _, q := range pkgs {
				if (q.path == p.path || p.deps[q.path]) && contains(q.imports, migrate) {
					via = append(via, rel(q.path))
				}
			}
			sort.Strings(via)
			add("%s/cmd/%s links %s (imported by %s). The migration replays history through the write path, and it ships in %s and nowhere else — a seat-side binary that links it can do what only the operator's tool may. Reach it from cmd/%s, or move the part this binary needs into a package that does not import it (internal/record/migration holds the manifest).",
				feovTools, bin, migratePkg, strings.Join(via, ", "), migrateBinary, migrateBinary)
		}
	}
	if len(binaries) == 0 {
		return []string{fmt.Sprintf("%s: go list named no binary under cmd/, so the migrate confinement checked nothing.", feovTools)}
	}
	if !contains(binaries, migrateBinary) {
		add("%s/cmd/%s does not exist — the migration's binary moved or was renamed, and the confinement names nothing.", feovTools, migrateBinary)
	}

	_, all, err := listModule(tools)
	if err != nil {
		return append(problems, fmt.Sprintf("%s: cannot list the module's packages, so the record variables and the migrate-only writer were not checked: %v", feovTools, err))
	}
	problems = append(problems, recordVarProblems(tools, module, pkgs)...)
	problems = append(problems, migrateOnlyProblems(tools, module, all)...)
	sort.Strings(problems)
	return problems
}

// parsedFile is one non-test file of a package that imports record, parsed, with the name it
// imports record under ("" where it does not).
type parsedFile struct {
	at     string // the file, relative to the module
	fset   *token.FileSet
	file   *ast.File
	record string // the import name of package record
}

// filesImportingRecord parses every file of every package but record that imports record. A file
// that cannot be parsed, or that dot-imports record, is a problem.
func filesImportingRecord(tools, module string, pkgs []shippedPackage, add func(string, ...any)) []parsedFile {
	recordPath := module + "/" + recordPkg
	var out []parsedFile
	for _, p := range pkgs {
		if p.path == recordPath || !contains(p.imports, recordPath) {
			continue
		}
		for _, f := range p.files {
			path := filepath.Join(p.dir, f)
			at, _ := filepath.Rel(tools, path)
			at = filepath.ToSlash(at)
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution|parser.ParseComments)
			if err != nil {
				add("%s/%s: cannot parse it: %v", feovTools, at, err)
				continue
			}
			alias := ""
			for _, im := range file.Imports {
				if v, _ := strconv.Unquote(im.Path.Value); v == recordPath {
					alias = "record"
					if im.Name != nil {
						alias = im.Name.Name
					}
				}
			}
			if alias == "." {
				add("%s/%s dot-imports %s, so a use of its names cannot be told from a local one. Import it by name.", feovTools, at, recordPkg)
				continue
			}
			out = append(out, parsedFile{at: at, fset: fset, file: file, record: alias})
		}
	}
	return out
}

// recordDecls parses package record's own files, the test files excluded, and hands each to visit.
func recordDecls(module string, pkgs []shippedPackage, add func(string, ...any), visit func(*ast.File)) {
	for _, p := range pkgs {
		if p.path != module+"/"+recordPkg {
			continue
		}
		for _, f := range p.files {
			file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(p.dir, f), nil, parser.SkipObjectResolution|parser.ParseComments)
			if err != nil {
				add("%s/%s/%s: cannot parse it: %v", feovTools, recordPkg, f, err)
				continue
			}
			visit(file)
		}
	}
}

// migrateOnlyProblems is the third rule of migrateConfinementProblems: every reference to a
// declaration of record marked migrateOnly, from a package other than migratePkg.
func migrateOnlyProblems(tools, module string, pkgs []shippedPackage) []string {
	var problems []string
	add := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }
	marked := func(doc *ast.CommentGroup) bool {
		if doc == nil {
			return false
		}
		for _, c := range doc.List {
			if strings.TrimSpace(c.Text) == migrateOnly {
				return true
			}
		}
		return false
	}
	names := map[string]bool{}
	recordDecls(module, pkgs, add, func(file *ast.File) {
		for _, d := range file.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && marked(d.Doc) {
					names[d.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, sp := range d.Specs {
					doc := d.Doc
					var ids []*ast.Ident
					switch sp := sp.(type) {
					case *ast.TypeSpec:
						if sp.Doc != nil {
							doc = sp.Doc
						}
						ids = []*ast.Ident{sp.Name}
					case *ast.ValueSpec:
						if sp.Doc != nil {
							doc = sp.Doc
						}
						ids = sp.Names
					}
					if marked(doc) {
						for _, id := range ids {
							names[id.Name] = true
						}
					}
				}
			}
		}
	})
	if len(names) == 0 {
		return append(problems, fmt.Sprintf("%s/%s: no declaration carries %s, so no reference to the migration's writer could be found.", feovTools, recordPkg, migrateOnly))
	}
	for _, pf := range filesImportingRecord(tools, module, pkgs, add) {
		if pf.record == "" || path.Dir(pf.at) == migratePkg {
			continue
		}
		ast.Inspect(pf.file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == pf.record && names[sel.Sel.Name] {
				add("%s/%s:%d names record.%s, the migration's writer. It writes an act without the live write's act judgements, and %s alone holds it: write through record.Append.",
					feovTools, pf.at, pf.fset.Position(sel.Pos()).Line, sel.Sel.Name, migratePkg)
			}
			return true
		})
	}
	return problems
}

// recordVarProblems is the second rule of migrateConfinementProblems: every assignment, increment
// or address-of a package-level variable of record, made in a shipped package other than record.
func recordVarProblems(tools, module string, pkgs []shippedPackage) []string {
	var problems []string
	add := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }
	vars := map[string]bool{}
	recordDecls(module, pkgs, add, func(file *ast.File) {
		for _, d := range file.Decls {
			if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.VAR {
				for _, sp := range g.Specs {
					for _, n := range sp.(*ast.ValueSpec).Names {
						if n.IsExported() {
							vars[n.Name] = true
						}
					}
				}
			}
		}
	})
	if len(vars) == 0 {
		return []string{fmt.Sprintf("%s/%s: read no exported package-level variable, so no assignment to one could be found.", feovTools, recordPkg)}
	}
	for _, pf := range filesImportingRecord(tools, module, pkgs, add) {
		if pf.record == "" {
			continue
		}
		written := func(e ast.Expr, how string) {
			sel, ok := e.(*ast.SelectorExpr)
			if !ok {
				return
			}
			if id, ok := sel.X.(*ast.Ident); !ok || id.Name != pf.record || !vars[sel.Sel.Name] {
				return
			}
			add("%s/%s:%d %s record.%s, a package-level variable of record, from outside record. A variable another package sets is a mode of the write path that every binary's copy reads and that the write path cannot see being set: pass the value to the call that needs it.",
				feovTools, pf.at, pf.fset.Position(sel.Pos()).Line, how, sel.Sel.Name)
		}
		ast.Inspect(pf.file, func(n ast.Node) bool {
			switch s := n.(type) {
			case *ast.AssignStmt:
				for _, l := range s.Lhs {
					written(l, "assigns")
				}
			case *ast.IncDecStmt:
				written(s.X, "assigns")
			case *ast.UnaryExpr:
				if s.Op == token.AND {
					written(s.X, "takes the address of")
				}
			}
			return true
		})
	}
	return problems
}
