package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY COMMAND PATH A SEAT IS TOLD TO TYPE MUST RESOLVE IN THE TREE IT WILL TYPE IT INTO.
//
// A refusal is the moment a seat learns what it should have done, and the instruction inside it is
// prose: `blue prove --location …`, `chair carry --carried-from <epoch>`, `lens show evidence`. Nine
// of these named a path or a flag the binary no longer has — the role words left the command path
// when the surface became seat-scoped, `--location` became `--quote` — and nothing failed, because a
// string is not a symbol. Probed: `feov-record --seat-id red-chair chair carry` → `no command named
// "chair" exists`. A seat that obeys the refusal is refused again, and now for a mistake that is not
// its own.
//
// The string IS the artifact, so this reads the strings: every backtick span in an interpreted
// string literal under the packages whose text reaches a seat, resolved against the real command
// trees (every seat's, and the operator's for a `feov-record …` form). Backtick spans are how this
// codebase marks something a seat types; a span whose first word is no command at all is prose and
// is skipped — this gate fires only on a span that BEGINS as a command reference and then goes wrong.
//
// TestNoRefusalNamesAFlagThatDoesNotExist checks the flags a refusal names, but only in refusals a
// bare invocation reaches, and never the path words. This one reads the source, so it sees the
// message behind every precondition and the path as well as the flag.

// seatFacingDirs are the packages whose string literals render into a seat's help, refusal or
// hint, relative to this package's directory.
var seatFacingDirs = []string{".", "../record", "../flags", "../bluedoc"}

// staleRoles are the words that used to begin a command path and no longer do. A span beginning
// with one is a command reference that has gone dead in the most common way, and it is reported
// as such rather than skipped as prose.
var staleRoles = map[string]bool{"lens": true, "chair": true, "blue": true, "bench": true, "merge": true}

var backtickSpan = regexp.MustCompile("`([^`]+)`")

func TestEveryCommandPathNamedInAStringResolves(t *testing.T) {
	roots := []*cobra.Command{newRoot()}
	for _, role := range []string{"lens", "chair", "blue", "bench"} {
		roots = append(roots, NewRootFor(record.SampleSeatOf(role)))
	}
	known := knownFlagNames(t)

	var spans, references int
	for _, dir := range seatFacingDirs {
		for _, lit := range stringLiterals(t, dir) {
			for _, m := range backtickSpan.FindAllStringSubmatch(lit.text, -1) {
				spans++
				problem, isRef := resolveSpan(roots, known, m[1])
				if isRef {
					references++
				}
				if problem != "" {
					t.Errorf("%s: `%s` — %s\n\nA seat that obeys this instruction is refused again, for a mistake that is not its own.", lit.at, m[1], problem)
				}
			}
		}
	}
	if references < 40 {
		t.Fatalf("only %d command references found in %d backtick spans — the walk is not reaching the sources, and a gate that finds nothing passes forever", references, spans)
	}
}

// EVERY FLAG A HELP PAGE NAMES IS ON THAT PAGE.
//
// A usage line is read on ONE page. `near-match --quote` carried the shared --quote contract,
// whose last sentence says "Name the section in --reason" — and near-match has no --reason. A seat
// following the page it is reading types a flag the page does not list, and the refusal it gets is
// cobra's "unknown flag", which teaches nothing about where the sentence was meant to go.
//
// Inherited flags count as on the page (cobra prints them under Global Flags). A flag that no
// verb at all registers is TestNoRefusalNamesAFlagThatDoesNotExist's business; this one is about
// the flag being on the WRONG page.
func TestEveryFlagAHelpPageNamesIsOnThatPage(t *testing.T) {
	var checked int
	for role, r := range AllRoots() {
		if !isSeatRole(role) {
			continue
		}
		walk(r, func(c *cobra.Command, path []string) {
			if !c.Runnable() {
				return
			}
			c.LocalNonPersistentFlags().VisitAll(func(f *pflag.Flag) {
				for _, m := range flagToken.FindAllStringSubmatch(f.Usage, -1) {
					checked++
					name := m[1]
					if c.Flags().Lookup(name) != nil || c.InheritedFlags().Lookup(name) != nil {
						continue
					}
					t.Errorf("%s: `%s --%s` names --%s, which is not a flag of this verb:\n\n%s", role, strings.Join(path, " "), f.Name, name, f.Usage)
				}
			})
		})
	}
	if checked < 20 {
		t.Fatalf("only %d flag mentions found across every seat's help — the walk is not reaching the tree", checked)
	}
}

// resolveSpan checks one backtick span. It returns the problem (empty when the span is sound or is
// not a command reference at all) and whether the span was a command reference.
func resolveSpan(roots []*cobra.Command, known map[string]bool, span string) (problem string, isRef bool) {
	words := strings.Fields(span)
	if len(words) == 0 {
		return "", false
	}
	if words[0] == "$" {
		words = words[1:]
	}
	if len(words) > 0 && words[0] == "feov-record" {
		words = words[1:]
		isRef = true
	}
	if len(words) == 0 {
		return "", isRef
	}
	if staleRoles[words[0]] && len(words) > 1 && isWord(words[1]) {
		return "begins with the role word `" + words[0] + "`, which left the command path when the surface became seat-scoped — a seat types the verb alone", true
	}
	// The longest command path any tree resolves.
	var cmd *cobra.Command
	consumed := 0
	for _, r := range roots {
		c, n := resolveIn(r, words)
		if n > consumed {
			cmd, consumed = c, n
		}
	}
	if consumed == 0 {
		if strings.HasPrefix(words[0], "--") {
			// A bare flag span (`--as FAIL`): the flag must exist somewhere in the tree.
			name := strings.TrimPrefix(words[0], "--")
			if isWord(name) && !known[name] {
				return "`--" + name + "` is a flag no command registers", true
			}
			return "", isWord(name)
		}
		if !isRef {
			return "", false // prose, not a command reference
		}
		return "names no command in any tree", true
	}
	isRef = true
	rest := words[consumed:]
	if len(rest) > 0 && strings.HasPrefix(rest[0], "<") && cmd.HasSubCommands() {
		// A template over a group's subcommands (`motion <subject> rule --id <id>`): the flags
		// belong to whichever leaf the placeholder stands for, and every leaf has its own page.
		return "", true
	}
	// Positionals before the first flag are the command's own to accept.
	var positionals []string
	for len(rest) > 0 && !strings.HasPrefix(rest[0], "--") && isWord(rest[0]) {
		positionals = append(positionals, rest[0])
		rest = rest[1:]
	}
	if len(positionals) > 0 {
		if cmd.HasSubCommands() {
			return "`" + strings.Join(positionals, " ") + "` is not a subcommand of `" + cmd.CommandPath() + "`", true
		}
		if cmd.Args != nil {
			if err := cmd.Args(cmd, positionals); err != nil {
				return "`" + cmd.CommandPath() + "` refuses the argument(s) `" + strings.Join(positionals, " ") + "`: " + err.Error(), true
			}
		}
	}
	for _, w := range rest {
		if !strings.HasPrefix(w, "--") {
			continue // a value, a placeholder, or prose after the command
		}
		name := strings.TrimPrefix(w, "--")
		if i := strings.IndexByte(name, '='); i >= 0 {
			name = name[:i]
		}
		if !isWord(name) || name == "help" {
			continue
		}
		if cmd.Flags().Lookup(name) == nil && cmd.InheritedFlags().Lookup(name) == nil {
			if known[name] {
				return "`--" + name + "` is not a flag of `" + cmd.CommandPath() + "` (another verb has it)", true
			}
			return "`--" + name + "` is a flag no command registers", true
		}
	}
	return "", true
}

// resolveIn walks words down a tree from its root, returning the deepest command reached and how
// many words it consumed. Zero means the first word is not a command here.
func resolveIn(root *cobra.Command, words []string) (*cobra.Command, int) {
	c := root
	n := 0
	for _, w := range words {
		if !isWord(w) {
			break
		}
		next := cmdAt(c, []string{w})
		if next == nil || w == "help" {
			break
		}
		c = next
		n++
	}
	if n == 0 {
		return nil, 0
	}
	return c, n
}

func isWord(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r == '-' && i > 0:
		default:
			return false
		}
	}
	return true
}

type sourceLiteral struct {
	at   string // file:line
	text string
}

// stringLiterals is every interpreted string literal in the package directory's non-test sources,
// with adjacent `+` operands joined — a long refusal is written as several literals, and a span may
// straddle the seam ("Settle each with `blue prove … --script <path> " + "--answers <gap>`").
// A non-literal operand joins as a space, so a span around one is empty and skipped.
func stringLiterals(t *testing.T, dir string) []sourceLiteral {
	t.Helper()
	var out []sourceLiteral
	fset := token.NewFileSet()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != dir && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		joined := map[ast.Expr]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			switch e := n.(type) {
			case *ast.BinaryExpr:
				if e.Op != token.ADD || joined[e] {
					return true
				}
				var parts []string
				var leaves []ast.Expr
				flatten(e, &parts, &leaves)
				if len(parts) == 0 {
					return true
				}
				for _, l := range leaves {
					joined[l] = true
				}
				out = append(out, sourceLiteral{at: fset.Position(e.Pos()).String(), text: strings.Join(parts, "")})
				return false
			case *ast.BasicLit:
				if e.Kind == token.STRING && !joined[e] && strings.HasPrefix(e.Value, `"`) {
					if s, err := strconv.Unquote(e.Value); err == nil {
						out = append(out, sourceLiteral{at: fset.Position(e.Pos()).String(), text: s})
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].at < out[j].at })
	return out
}

// flatten collects the operands of a `+` chain: interpreted string literals as their text, anything
// else as one space. It records every BinaryExpr and literal it visited so the walk does not read
// them twice.
func flatten(e ast.Expr, parts *[]string, leaves *[]ast.Expr) {
	switch x := e.(type) {
	case *ast.BinaryExpr:
		if x.Op == token.ADD {
			*leaves = append(*leaves, x)
			flatten(x.X, parts, leaves)
			flatten(x.Y, parts, leaves)
			return
		}
	case *ast.ParenExpr:
		flatten(x.X, parts, leaves)
		return
	case *ast.BasicLit:
		if x.Kind == token.STRING && strings.HasPrefix(x.Value, `"`) {
			if s, err := strconv.Unquote(x.Value); err == nil {
				*leaves = append(*leaves, x)
				*parts = append(*parts, s)
				return
			}
		}
	}
	*parts = append(*parts, " ")
}
