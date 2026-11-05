package planguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	gotoken "go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// sqlVerbs are the heads of a statement SQLite plans; driver.go's isExplainable reads the same set.
var sqlVerbs = []string{"SELECT", "WITH", "INSERT", "UPDATE", "DELETE", "REPLACE"}

// piece is one stretch of a statement's text as the source spells it: a constant, or a hole where
// the text is computed at run time (a variable, a call, a %-verb of a format).
type piece struct {
	text string
	hole bool
}

// SourceStatement is one statement text the source can issue, as the constant stretches it is
// built from, in order.
type SourceStatement struct {
	At    string   // file:line of its first spelling
	Frags []string // the constant stretches, whitespace-normalised; holes lie between them
	Holes bool     // whether any part is computed
	text  string   // the statement as String renders it
}

// String is the statement's text, whitespace-normalised, each computed stretch shown as "…".
func (s SourceStatement) String() string { return s.text }

// constTable resolves a package's string constants, across its files and inside its functions.
type constTable struct {
	decls map[string]ast.Expr
	other map[string]*constTable // an imported package's table, by the name the source calls it
	memo  map[string]*string
}

func (c *constTable) value(name string) (string, bool) {
	if v, ok := c.memo[name]; ok {
		if v == nil {
			return "", false
		}
		return *v, true
	}
	c.memo[name] = nil // a cycle resolves to "not a constant"
	e, ok := c.decls[name]
	if !ok {
		return "", false
	}
	ps := c.pieces(e)
	if len(ps) != 1 || ps[0].hole {
		return "", false
	}
	v := ps[0].text
	c.memo[name] = &v
	return v, true
}

// pieces is e's text as constant stretches and holes, adjacent constants joined.
func (c *constTable) pieces(e ast.Expr) []piece {
	var out []piece
	add := func(ps ...piece) {
		for _, p := range ps {
			if !p.hole && len(out) > 0 && !out[len(out)-1].hole {
				out[len(out)-1].text += p.text
				continue
			}
			if p.hole && len(out) > 0 && out[len(out)-1].hole {
				continue
			}
			out = append(out, p)
		}
	}
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind == gotoken.STRING {
			if s, err := strconv.Unquote(x.Value); err == nil {
				add(piece{text: s})
				return out
			}
		}
	case *ast.ParenExpr:
		return c.pieces(x.X)
	case *ast.BinaryExpr:
		if x.Op == gotoken.ADD {
			add(c.pieces(x.X)...)
			add(c.pieces(x.Y)...)
			return out
		}
	case *ast.Ident:
		if v, ok := c.value(x.Name); ok {
			add(piece{text: v})
			return out
		}
	case *ast.SelectorExpr:
		if pkg, ok := x.X.(*ast.Ident); ok {
			if o := c.other[pkg.Name]; o != nil {
				if v, ok := o.value(x.Sel.Name); ok {
					add(piece{text: v})
					return out
				}
			}
		}
	case *ast.CallExpr:
		if len(x.Args) > 0 {
			if isSprintf(x) {
				f := c.pieces(x.Args[0])
				if len(f) == 1 && !f[0].hole {
					add(formatPieces(f[0].text)...)
					return out
				}
			}
		}
	}
	add(piece{hole: true})
	return out
}

// formatPieces splits a format string at its verbs, each verb a hole. %% is a literal percent.
func formatPieces(f string) []piece {
	var out []piece
	var b strings.Builder
	for i := 0; i < len(f); i++ {
		if f[i] != '%' {
			b.WriteByte(f[i])
			continue
		}
		if i+1 < len(f) && f[i+1] == '%' {
			b.WriteByte('%')
			i++
			continue
		}
		out = append(out, piece{text: b.String()}, piece{hole: true})
		b.Reset()
		for i++; i < len(f) && !strings.ContainsRune("vTtbcdoOqxXUeEfFgGsp", rune(f[i])); i++ {
		}
	}
	return append(out, piece{text: b.String()})
}

// isStatementHead is whether s begins a statement SQLite plans: a statement verb followed by more
// text. A verb alone is a word being compared, not a statement.
func isStatementHead(s string) bool {
	f := strings.Fields(strings.ToUpper(s))
	if len(f) < 2 {
		return false
	}
	for _, v := range sqlVerbs {
		if f[0] == v {
			return true
		}
	}
	return false
}

// NormSQL is a statement's text with its whitespace collapsed: the form statements are compared in.
func NormSQL(s string) string { return strings.Join(strings.Fields(s), " ") }

// SourceStatements reads every statement text the Go source in dirs spells, sorted by position.
// Test files are not read: a test's own statements are not the record's.
//
// A STATEMENT IS AN EXPRESSION WHOSE TEXT BEGINS WITH A STATEMENT VERB, read at its widest: the
// whole concatenation or format, never one of its operands, so `base + where` is one statement
// with a hole and not two. String constants are inlined where they are used — across the packages
// read together, so record's use of a recordsql constant resolves — and a constant's declaration
// is not a statement of its own: a constant nobody issues is not a statement the record runs.
//
// WHAT IT CANNOT READ is a statement whose head is computed — a variable, a call's result, a
// strings.Builder that writes the verb in a separate call. Those are not in the census; the caller
// learns of them only by the plans it records.
func SourceStatements(dirs ...string) ([]SourceStatement, error) {
	fset := gotoken.NewFileSet()
	type pkgSrc struct {
		files []*ast.File
		tab   *constTable
	}
	var pkgs []*pkgSrc
	byName := map[string]*constTable{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		p := &pkgSrc{tab: &constTable{decls: map[string]ast.Expr{}, other: byName, memo: map[string]*string{}}}
		for _, e := range entries {
			n := e.Name()
			if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, filepath.Join(dir, n), nil, parser.SkipObjectResolution)
			if err != nil {
				return nil, err
			}
			p.files = append(p.files, f)
			byName[f.Name.Name] = p.tab
			ast.Inspect(f, func(n ast.Node) bool {
				if g, ok := n.(*ast.GenDecl); ok && g.Tok == gotoken.CONST {
					for _, s := range g.Specs {
						vs := s.(*ast.ValueSpec)
						for i, name := range vs.Names {
							if i < len(vs.Values) {
								p.tab.decls[name.Name] = vs.Values[i]
							}
						}
					}
				}
				return true
			})
		}
		if len(p.files) == 0 {
			return nil, fmt.Errorf("no Go source in %s — a census of it would read nothing and pass", dir)
		}
		pkgs = append(pkgs, p)
	}

	seen := map[string]bool{}
	var out []SourceStatement
	for _, p := range pkgs {
		for _, f := range p.files {
			var stack []ast.Node
			ast.Inspect(f, func(n ast.Node) bool {
				if n == nil {
					stack = stack[:len(stack)-1]
					return true
				}
				var parent ast.Node
				if len(stack) > 0 {
					parent = stack[len(stack)-1]
				}
				if g, ok := n.(*ast.GenDecl); ok && g.Tok == gotoken.CONST {
					return false // inlined at its uses; returning false skips the nil visit, so nothing is pushed
				}
				stack = append(stack, n)
				e, ok := n.(ast.Expr)
				if !ok || !widest(e, parent) {
					return true
				}
				ps := p.tab.pieces(e)
				if len(ps) == 0 || ps[0].hole || !isStatementHead(ps[0].text) {
					return true
				}
				st := SourceStatement{At: shortPos(fset.Position(e.Pos()))}
				var shown []string
				for _, pc := range ps {
					if pc.hole {
						st.Holes = true
						shown = append(shown, "…")
						continue
					}
					if fr := NormSQL(pc.text); fr != "" {
						st.Frags = append(st.Frags, fr)
						shown = append(shown, fr)
					}
				}
				st.text = strings.Join(shown, " ")
				if key := st.String(); !seen[key] {
					seen[key] = true
					out = append(out, st)
				}
				return true
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At < out[j].At })
	return out, nil
}

// widest reports whether e is read whole here, rather than as an operand of a wider string.
func widest(e ast.Expr, parent ast.Node) bool {
	switch e.(type) {
	case *ast.BasicLit, *ast.BinaryExpr, *ast.Ident, *ast.SelectorExpr, *ast.CallExpr:
	default:
		return false
	}
	switch p := parent.(type) {
	case *ast.BinaryExpr:
		return p.Op != gotoken.ADD
	case *ast.ParenExpr:
		return false
	case *ast.SelectorExpr:
		return false // the X or Sel of a selector is a name, not a string
	case *ast.CallExpr:
		// A format is read whole at its call; its first argument is not a statement of its own.
		if isSprintf(p) && len(p.Args) > 0 && p.Args[0] == e {
			return false
		}
	}
	return true
}

func isSprintf(c *ast.CallExpr) bool {
	sel, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "fmt" && sel.Sel.Name == "Sprintf"
}

func shortPos(p gotoken.Position) string {
	dir, file := filepath.Split(p.Filename)
	return filepath.Base(filepath.Clean(dir)) + "/" + file + ":" + strconv.Itoa(p.Line)
}

// PlannedBy reports whether one of texts (each NormSQL'd) is this statement: it begins with the
// statement's first constant stretch and holds the rest in order. A wholly constant statement must
// BEGIN the text, not merely occur inside it; a helper may append a clause to the text it is
// handed, so it need not END it.
func (s SourceStatement) PlannedBy(texts []string) bool {
	for _, p := range texts {
		if !strings.HasPrefix(p, s.Frags[0]) {
			continue
		}
		rest, ok := p[len(s.Frags[0]):], true
		for _, fr := range s.Frags[1:] {
			i := strings.Index(rest, fr)
			if i < 0 {
				ok = false
				break
			}
			rest = rest[i+len(fr):]
		}
		if ok {
			return true
		}
	}
	return false
}
