package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

// rawLockCalls is every OS lock primitive a plugin could call directly, by import path. File
// locking goes through github.com/gofrs/flock, which takes flock(2) on Unix and LockFileEx on byte
// 0, length 1, on Windows: one primitive keeps every plugin's locks interlocking with every
// other's by construction. A hand-written lock that takes another primitive, or another byte range,
// shares a lock file with a gofrs holder and excludes nothing.
var rawLockCalls = map[string]map[string]bool{
	"syscall":                  {"Flock": true},
	"golang.org/x/sys/unix":    {"Flock": true},
	"golang.org/x/sys/windows": {"LockFileEx": true, "UnlockFileEx": true},
}

// rawLockRoots are the trees whose Go code ships to a consumer.
var rawLockRoots = []string{"plugins", "shared"}

// rawLockProblems names every reference, in a non-test Go file under rawLockRoots, to a raw lock
// primitive. Files are parsed straight from the tree, whatever their build constraints, so a
// Windows-only file is read on Linux. A reference is a selector on the file's own name for the
// package, aliased or not; a dot import of one of the packages hides its references, so the import
// itself is a finding. Directories the go command ignores — testdata, and names starting with "."
// or "_" — are skipped. A root that does not exist holds nothing to read; a walk with no Go file
// read at all is a finding, never a pass.
func rawLockProblems(root string) []string {
	var problems []string
	add := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }
	read := 0
	for _, top := range rawLockRoots {
		err := filepath.WalkDir(filepath.Join(root, top), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := d.Name()
			if d.IsDir() {
				if p != filepath.Join(root, top) && (name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			fset := token.NewFileSet()
			file, perr := parser.ParseFile(fset, p, nil, parser.SkipObjectResolution)
			if perr != nil {
				add("%s: does not parse, so its lock calls could not be read: %v", rel, perr)
				return nil
			}
			read++
			for _, f := range rawLockCallsIn(fset, file) {
				add("%s:%s. File locks go through github.com/gofrs/flock (flock.New, then Lock, RLock or TryLockContext): one primitive keeps every plugin's locks interlocking.", rel, f)
			}
			return nil
		})
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			add("%s: could not be walked for raw lock calls: %v", top, err)
		}
	}
	if read == 0 {
		add("%s: no Go file was read, so no raw lock call could be found.", strings.Join(rawLockRoots, ", "))
	}
	return problems
}

// rawLockCallsIn returns "line: what" for each raw lock reference in file.
func rawLockCallsIn(fset *token.FileSet, file *ast.File) []string {
	local := map[string]string{} // the file's name for a package -> its import path
	var found []string
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil || rawLockCalls[path] == nil {
			continue
		}
		name := path[strings.LastIndex(path, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if name == "." {
			found = append(found, fmt.Sprintf("%d dot-imports %s, which hides any raw lock call made through it", fset.Position(imp.Pos()).Line, path))
			continue
		}
		local[name] = path
	}
	if len(local) == 0 {
		return found
	}
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if path, ok := local[id.Name]; ok && rawLockCalls[path][sel.Sel.Name] {
			found = append(found, fmt.Sprintf("%d names %s.%s, a raw lock primitive", fset.Position(sel.Pos()).Line, path, sel.Sel.Name))
		}
		return true
	})
	return found
}
