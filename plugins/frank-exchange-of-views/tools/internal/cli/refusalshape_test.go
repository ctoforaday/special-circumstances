package cli

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

// NO MESSAGE IS WRITTEN WITH A COMMAND'S NAME AT ITS HEAD, in any layer a command reaches.
//
// The rendered gate (cli_test.go's `run`, difftest's, the release gate's) reads every refusal a
// test DRIVES. This one reads the source, so a refusal no test reaches is held to the same rule:
// every message string in the layers a command calls into is parsed, and none opens — alone or
// behind the record layer's word — with a role word, a command's name or path, or a `%s` in the
// place a name would stand, followed by a colon, a flag, `requires` or `refused`.
//
// The command names are the trees' own: every path, and every path's last word, on every seat's
// surface and the operator's.

// refusalLayers are the packages a command reaches, below internal/.
var refusalLayers = []string{"cli", "record", "fetchcache", "proof", "report", "bluedoc", "reportproj", "view"}

// messageMakers are the calls whose string argument is a message: the error constructors, and the
// formatters a message is built with before it reaches one.
var messageMakers = map[string]bool{
	"fmt.Errorf": true, "errors.New": true, "feov.Errorf": true, "feov.Wrap": true, "fmt.Sprintf": true,
}

// commandNames is every command path below a root, and every path's last word.
func commandNames() []string {
	set := map[string]bool{}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if p := seat.PathOf(c); p != "" {
			set[p] = true
			set[c.Name()] = true
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	for _, root := range AllRoots() {
		walk(root)
	}
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	// Longest first, so `motion docket rule` is read before `motion`.
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) != len(out[j]) {
			return len(out[i]) > len(out[j])
		}
		return out[i] < out[j]
	})
	return out
}

// composedHead is the head a message opens with when that head names a command or a role the way
// a rendered refusal's own head does; "" for a message that starts at what is wrong.
func composedHead(message string, names, roles []string) string {
	rest := strings.TrimPrefix(message, "record: ")
	for _, role := range roles {
		if strings.HasPrefix(rest, role+": ") {
			return role + ":"
		}
		for _, n := range names {
			if strings.HasPrefix(rest, role+" "+n+" ") || strings.HasPrefix(rest, role+" "+n+":") {
				return role + " " + n
			}
		}
	}
	for _, n := range names {
		for _, tail := range []string{": ", " --", " requires", " refused"} {
			if strings.HasPrefix(rest, n+tail) {
				return n + tail
			}
		}
	}
	// A `%s` WHERE A NAME WOULD STAND. `%s: ` is left out: it heads a table row or wraps a path far
	// more often than it heads a refusal, and the rendered gates read what it prints.
	for _, tail := range []string{" --", " requires", " refused"} {
		if strings.HasPrefix(rest, "%s"+tail) {
			return "%s" + tail
		}
	}
	return ""
}

// messagesIn lists every message string of one file: the first string argument of each message
// maker, read through a concatenation to its leftmost literal — where a message starts.
func messagesIn(t *testing.T, path string) map[token.Position]string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	out := map[token.Position]string{}
	ast.Inspect(file, func(n ast.Node) bool {
		// A SUCCESS LINE HAS NO RENDERER HEAD: what a result's Human prints is the whole line, so a
		// name at its head is the one naming it gets.
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "Human" {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || !messageMakers[pkg.Name+"."+sel.Sel.Name] {
			return true
		}
		for _, arg := range call.Args {
			leftmost := arg
			for {
				bin, ok := leftmost.(*ast.BinaryExpr)
				if !ok {
					break
				}
				leftmost = bin.X
			}
			lit, ok := leftmost.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			text, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatalf("%s: unquoting %s: %v", fset.Position(lit.Pos()), lit.Value, err)
			}
			out[fset.Position(lit.Pos())] = text
			break
		}
		return true
	})
	return out
}

func TestNoMessageOpensWithACommandOrARole(t *testing.T) {
	names, roles := commandNames(), roleWords()
	if len(names) < 60 {
		t.Fatalf("the trees gave %d command names — the walk is not reading the surface", len(names))
	}
	files, messages := 0, 0
	var found []string
	for _, layer := range refusalLayers {
		err := filepath.WalkDir(filepath.Join("..", layer), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") {
				return nil
			}
			files++
			for pos, message := range messagesIn(t, path) {
				messages++
				if head := composedHead(message, names, roles); head != "" {
					found = append(found, fmt.Sprintf("%s:%d: opens %q: %.90q", filepath.ToSlash(pos.Filename), pos.Line, head, message))
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if files < 100 || messages < 500 {
		t.Fatalf("read %d files and %d messages — the walk is not reading the layers", files, messages)
	}
	sort.Strings(found)
	if len(found) > 0 {
		t.Errorf("%d of %d messages open with a command or a role, which the renderer prints from the command that resolved — "+
			"a message starts at what is wrong:\n  %s", len(found), messages, strings.Join(found, "\n  "))
	}
}

// The reading tells a composed head from a sentence.
func TestTheMessageGateTellsAHeadFromASentence(t *testing.T) {
	names, roles := commandNames(), roleWords()
	for message, composed := range map[string]bool{
		"record: mint refused — the first sentence":          true,
		"record: %s requires --%s":                           true,
		"%s refused — the originator":                        true,
		"record: close --id names gap %s":                    true,
		"record: motion docket rule requires --reopens-on":   true,
		"lens finding requires --reason":                     true,
		"chair: no such gap":                                 true,
		"fetch: %s returned %d":                              true,
		"migrate: reading %s: %w":                            true,
		"requires --reason: the explanation":                 false,
		"record: %s has already recorded a %s this sitting":  false,
		"the --id names gap %s, which no mint event created": false,
		"a carry restates an earlier closure":                false,
		"closing it twice double-counts closure history":     false,
		"recordsql: opening %s: %w":                          false,
		"%s was minted --check-kind computation":             false,
		"mints are budgeted: this lens has %d left":          false,
		"record: a %s names the gap it is about":             false,
		"--accept: no gap %s on the board":                   false,
		"reading the report to ingest at %s: %w":             false,
		"record: another seat's verdict stands this sitting": false,
		"record: unknown class %s — use a registry slug":     false,
		"%s ran to exit %d, but its output carries an error": false,
	} {
		if got := composedHead(message, names, roles) != ""; got != composed {
			t.Errorf("%q: composed = %v, want %v", message, got, composed)
		}
	}
}

// A SEAT VERB'S REFUSAL IS MINTED BY ONE CONSTRUCTOR, WITH ITS KIND. In the six packages that hold
// the seat verbs a refusal is feov.Errorf or feov.Wrap, so every one carries a code a `--json`
// consumer can branch on. fmt.Errorf stays for one thing: passing a failure up from below with
// `%w`, which keeps the code the failure came with — a read that failed is not given a code it
// does not have. A tool defect with no cause beneath it says so with feov.Internal.
func TestASeatVerbsRefusalIsMintedWithItsCode(t *testing.T) {
	files, found := 0, []string{}
	for _, pkg := range []string{"blue", "lens", "chair", "bench", "motion", "seat"} {
		paths, err := filepath.Glob(filepath.Join(pkg, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			files++
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatalf("parsing %s: %v", path, err)
			}
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkgName, _ := sel.X.(*ast.Ident)
				if pkgName == nil {
					return true
				}
				at := fmt.Sprintf("%s:%d", filepath.ToSlash(path), fset.Position(call.Pos()).Line)
				switch pkgName.Name + "." + sel.Sel.Name {
				case "errors.New":
					found = append(found, at+": errors.New")
				case "fmt.Errorf":
					if !strings.Contains(formatOf(call), "%w") {
						found = append(found, at+": fmt.Errorf with no %w")
					}
				}
				return true
			})
		}
	}
	if files < 30 {
		t.Fatalf("read %d files — the walk is not reading the seat packages", files)
	}
	if len(found) > 0 {
		t.Errorf("%d refusal(s) in the seat packages are minted without a code — use feov.Errorf with the refusal's kind, "+
			"or wrap the failure beneath it with %%w:\n  %s", len(found), strings.Join(found, "\n  "))
	}
}

// formatOf is a call's first argument as source text, every literal of a concatenation joined.
func formatOf(call *ast.CallExpr) string {
	if len(call.Args) == 0 {
		return ""
	}
	var b strings.Builder
	ast.Inspect(call.Args[0], func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			b.WriteString(lit.Value)
		}
		return true
	})
	return b.String()
}

// NO HELPER BELOW THE COMMAND LAYER IS HANDED THE COMMAND'S NAME. The reference checks, the
// required-field check and the report's own layers are reached from several verbs each, so a name
// passed down to compose into a refusal is wrong for every verb but one — and a second naming for
// that one. A parameter called `verb` in those files is that hand-off.
func TestNoSharedCheckTakesTheCommandsName(t *testing.T) {
	var files []string
	for _, pattern := range []string{"../record/refs.go", "../record/about.go", "../record/recordpb/requiredfields.go", "../bluedoc/*.go", "../reportproj/*.go"} {
		paths, err := filepath.Glob(pattern)
		if err != nil || len(paths) == 0 {
			t.Fatalf("%s matches no file (%v) — the gate is reading nothing there", pattern, err)
		}
		files = append(files, paths...)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			for _, field := range fn.Type.Params.List {
				for _, name := range field.Names {
					if name.Name == "verb" {
						t.Errorf("%s:%d: %s takes a `verb` — the renderer names the command; the check says what is wrong",
							filepath.ToSlash(path), fset.Position(fn.Pos()).Line, fn.Name.Name)
					}
				}
			}
		}
	}
}
