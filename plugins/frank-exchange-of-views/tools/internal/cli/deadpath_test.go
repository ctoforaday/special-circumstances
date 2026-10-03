package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
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

// moduleRoot is the module directory, relative to this package's: every non-test Go source under
// it is read. The set used to be a hand-kept list of four packages, and it omitted internal/report
// (which still said `bench outcome`), internal/view and internal/hookgate — a list that names what
// reaches a seat is a second copy of the import graph, and goes stale the way every second copy
// does. The whole module is scanned instead, minus the directories below.
const moduleRoot = "../.."

// notSeatFacing are the module directories whose strings reach no seat, each with the reason.
// Keyed on the path relative to moduleRoot; a directory's subtree is excluded with it.
var notSeatFacing = map[string]string{
	"internal/difftest":       "the golden harness: its strings are expected outputs, not messages",
	"internal/goldentest":     "golden-file helpers for tests",
	"internal/testbuild":      "builds binaries for tests",
	"internal/repotree":       "locates the repository for dev harnesses",
	"internal/record/migrate": "operator-only: brings archived runs forward, and its text is about record shapes",
	"integration":             "test packages",
	"releasegate":             "the release-gate fuzzer",
	"third_party":             "vendored pins",
}

// roleOfDir is the seat a package's text is addressed to, read from the package path: the four
// role packages under internal/cli each build one seat's verbs and nothing else's. Everything else
// is shared — a refusal in internal/record can reach whichever seat wrote the event — and resolves
// against every tree.
func roleOfDir(rel string) string {
	for _, role := range []string{"lens", "chair", "blue", "bench"} {
		if rel == "internal/cli/"+role || strings.HasPrefix(rel, "internal/cli/"+role+"/") {
			return role
		}
	}
	return ""
}

// staleRoles are the words that used to begin a command path and no longer do. A span beginning
// with one is a command reference that has gone dead in the most common way, and it is reported
// as such rather than skipped as prose.
var staleRoles = map[string]bool{"lens": true, "chair": true, "blue": true, "bench": true, "merge": true}

var backtickSpan = regexp.MustCompile("`([^`]+)`")

// attributed matches the prose that hands a verb to ANOTHER seat: "the chair's `carry`", "blue's
// `edit`", or the seat as the actor with up to three words between ("is the chair's to restate
// with `carry`", "the gap's originating lens moves it with `regrade`"). A span so attributed is not an instruction to the reader, so it may name a verb the
// reader's own tree lacks; an unattributed span naming another seat's verb is the regression this
// gate exists for — "use `carry --carried-from <epoch>`" in a refusal only a lens ever reads, and
// the lens that obeys it is told no such command exists.
var attributed = regexp.MustCompile(`\b(?:lens|chair|blue|bench)(?:'s|’s)?(?:\s+[a-z]+){0,3}\s*$`)

// listContinuation is what may sit between one attributed span and the next in the same list:
// "the lens's `mint`, `close` and `regrade`".
var listContinuation = regexp.MustCompile(`^(?:,|,?\s+and|,?\s+or|/)?\s*$`)

// seatSpan is one backtick span in a message, and whether the prose before it hands it to a named
// seat.
type seatSpan struct {
	text       string
	attributed bool
}

// spansIn reads a message's backtick spans with their attribution. carried is the attribution a
// list continues: set by "the chair's `a`", kept across ", `b` and `c`", dropped at the first
// other prose.
func spansIn(text string) []seatSpan {
	var out []seatSpan
	carried := false
	prevEnd := 0
	for _, m := range backtickSpan.FindAllStringSubmatchIndex(text, -1) {
		here := attributed.MatchString(text[:m[0]]) || (carried && listContinuation.MatchString(text[prevEnd:m[0]]))
		prevEnd = m[1]
		carried = here
		out = append(out, seatSpan{text: text[m[2]:m[3]], attributed: here})
	}
	return out
}

// seatTrees is every seat's command tree and the operator's, keyed by role.
func seatTrees() map[string]*cobra.Command {
	trees := map[string]*cobra.Command{record.OperatorRole: NewRootFor(record.OperatorRole)}
	for _, role := range []string{"lens", "chair", "blue", "bench"} {
		trees[role] = NewRootFor(record.SampleSeatOf(role))
	}
	return trees
}

var (
	seatTreesOnce   sync.Once
	seatTreesShared map[string]*cobra.Command
)

// EVERY REFUSAL A SEAT RECEIVES NAMES ONLY VERBS ON THAT SEAT'S SURFACE, OR HANDS THE VERB TO ITS
// HOLDER IN PROSE.
//
// The source scan cannot see who reads a refusal raised in a shared package: record's validate
// answers the lens's `close` and the chair's `carry` from one switch arm, keyed on the body type
// both write. The harness can — it knows the seat it dispatched. So every refusal any test drives
// through `run` is read here against the tree of the seat that received it. #1226 stripped the
// role word from these refusals and left the lens told to `carry --carried-from`, the chair told
// to `close` and to raise a `finding`: each resolved on SOME tree, so the source scan was green,
// and a seat that obeyed got "no command named …".
func assertRefusalSpeaksToItsSeat(t *testing.T, seatID string, own *cobra.Command, err error) {
	t.Helper()
	if seatID == "" || err == nil {
		return
	}
	seatTreesOnce.Do(func() { seatTreesShared = seatTrees() })
	for _, s := range spansIn(err.Error()) {
		words := strings.Fields(s.text)
		if s.attributed || len(words) == 0 || words[0] == "feov-record" || strings.HasPrefix(words[0], "-") {
			continue
		}
		if _, n := resolveIn(own, words); n > 0 {
			continue
		}
		var holders []string
		for role, r := range seatTreesShared {
			if role == record.OperatorRole {
				continue
			}
			if _, n := resolveIn(r, words); n > 0 {
				holders = append(holders, role)
			}
		}
		if len(holders) == 0 {
			continue // not a command anywhere: prose, or a dead path TestEveryCommandPathNamedInAStringResolves reports
		}
		sort.Strings(holders)
		t.Errorf("seat %s was refused with `%s`, which is not on its surface (it is %s) — a seat that obeys it is told no such command exists. Attribute it in prose (\"the %s's `%s`\") or name the seat's own act.\n\nrefusal: %v",
			seatID, s.text, holderPhrase(holders), holders[0], s.text, err)
	}
}

func TestEveryCommandPathNamedInAStringResolves(t *testing.T) {
	trees := seatTrees()
	known := knownFlagNames(t)

	var spans, references int
	for _, lit := range moduleLiterals(t) {
		role := roleOfDir(lit.pkg)
		for _, s := range spansIn(lit.text) {
			spans++
			problem, isRef := resolveSpan(trees, role, known, s.text, s.attributed)
			if isRef {
				references++
			}
			if problem != "" {
				t.Errorf("%s: `%s` — %s\n\nA seat that obeys this instruction is refused again, for a mistake that is not its own.", lit.at, s.text, problem)
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

// resolveSpan checks one backtick span from a literal addressed to `role` ("" when the literal is
// shared and may reach any seat). It returns the problem (empty when the span is sound or is not a
// command reference at all) and whether the span was a command reference.
//
// A span resolves PER TREE where the producing site's role is known. In a role package it must
// resolve on that role's tree: a verb another seat holds is a dead path for this reader unless the
// prose attributes it ("the chair's `carry`"). In a shared package the reader is not knowable from
// the source — `show`'s listing says which verbs WRITE a view, to every seat — so the span need
// only resolve on some tree here, and assertRefusalSpeaksToItsSeat holds every refusal a test
// delivers to the seat that received it. The flags are checked on whichever tree holds the verb.
func resolveSpan(trees map[string]*cobra.Command, role string, known map[string]bool, span string, attributedHere bool) (problem string, isRef bool) {
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
		role = "" // the operator's spelling reaches whichever tree the --seat-id selects
	}
	if len(words) == 0 {
		return "", isRef
	}
	if staleRoles[words[0]] && len(words) > 1 && isWord(words[1]) {
		return "begins with the role word `" + words[0] + "`, which left the command path when the surface became seat-scoped — a seat types the verb alone, and prose hands another seat's verb to it as \"the " + words[0] + "'s `" + strings.Join(words[1:], " ") + "`\"", true
	}
	// The longest command path any tree resolves, and which trees reach it.
	var cmd *cobra.Command
	consumed := 0
	var holders []string
	for name, r := range trees {
		c, n := resolveIn(r, words)
		if n == 0 {
			continue
		}
		if n > consumed {
			cmd, consumed, holders = c, n, nil
		}
		if n == consumed {
			holders = append(holders, name)
		}
	}
	sort.Strings(holders)
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
	if !attributedHere && role != "" && !slices.Contains(holders, role) {
		return "is not on the " + role + "'s surface (it is " + holderPhrase(holders) + ") — a " + role + " reading this has no such command; attribute it in prose (\"the " + holders[0] + "'s `" + span + "`\") if it is another seat's act", true
	}
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

// holderPhrase names the trees that hold a verb: "the chair's", "the chair's and the bench's".
func holderPhrase(holders []string) string {
	var parts []string
	for _, h := range holders {
		parts = append(parts, "the "+h+"'s")
	}
	return strings.Join(parts, " and ")
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
	pkg  string // the package directory, relative to the module root
	text string
}

// moduleLiterals is every interpreted string literal in the module's non-test, non-generated
// sources outside notSeatFacing, with adjacent `+` operands joined — a long refusal is written as
// several literals, and a span may straddle the seam ("Settle each with `blue prove … --script
// <path> " + "--answers <gap>`"). A non-literal operand (a %s argument, a constant from another
// package, a function call) joins as one space, so a span that straddles one is read with a gap
// where the value goes — `show --run %s --view <name>` still names its command and its flags.
func moduleLiterals(t *testing.T) []sourceLiteral {
	t.Helper()
	var out []sourceLiteral
	fset := token.NewFileSet()
	var excluded int
	err := filepath.WalkDir(moduleRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(moduleRoot, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if path != moduleRoot && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			if notSeatFacing[rel] != "" {
				excluded++
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
		pkg := filepath.ToSlash(filepath.Dir(rel))
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
				out = append(out, sourceLiteral{at: fset.Position(e.Pos()).String(), pkg: pkg, text: strings.Join(parts, "")})
				// The walk CONTINUES into the non-literal operands: a call or a composite
				// inside the chain can carry its own literals, with their own spans.
				return true
			case *ast.BasicLit:
				if e.Kind == token.STRING && !joined[e] && strings.HasPrefix(e.Value, `"`) {
					if s, err := strconv.Unquote(e.Value); err == nil {
						out = append(out, sourceLiteral{at: fset.Position(e.Pos()).String(), pkg: pkg, text: s})
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
	if excluded != len(notSeatFacing) {
		t.Fatalf("notSeatFacing names %d directories and the walk met %d — an entry names a directory that no longer exists, and the exclusion reads as an exclusion of nothing", len(notSeatFacing), excluded)
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

// THE CHAIR'S PASS OVER AN UNRAISED CONTRADICTION IS ANSWERED IN THE CHAIR'S TERMS.
//
// No other test delivers this refusal to the chair, so assertRefusalSpeaksToItsSeat never read it,
// and it shipped telling the chair to raise a `finding` — a lens's verb the chair does not hold.
// Driven here through the chair's own tree so the harness reads it against that tree.
func TestThePassRefusalOverAContradictionNamesWhoRaisesIt(t *testing.T) {
	runDir := corroborateRun(t)
	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/contradicts", "--title", "A source that disagrees",
		"--quote", corroborated, "--as", "refutes", "--confidence", "high",
		"--reason", "the source says the opposite at page 9"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "verdict", "--run", runDir, "--seat-id", "red-chair", "--as", "PASS")
	if err == nil || !strings.Contains(err.Error(), corroborated) {
		t.Fatalf("the chair's PASS over a contradiction nobody raised was not refused over it: %v", err)
	}
}
