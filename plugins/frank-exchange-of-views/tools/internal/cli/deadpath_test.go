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

// attributed matches the prose that hands a verb to a NAMED seat: "the chair's `carry`", "Blue's
// `edit`", or the seat as the actor with up to three words between ("is the chair's to restate
// with `carry`", "the gap's originating lens moves it with `regrade`"). The seat it captures is
// the one the span is then resolved against: an attribution is a claim that THAT seat holds the
// verb, and a claim nothing checks lets "the lens's `carry`" through as readily as "the chair's".
// An unattributed span naming another seat's verb is the regression this gate exists for — "use
// `carry --carried-from <epoch>`" in a refusal only a lens ever reads, and the lens that obeys it
// is told no such command exists. Case-insensitive: a sentence may begin with the seat.
var attributed = regexp.MustCompile(`(?i)\b(lens|chair|blue|bench)(?:'s|’s)?(?:\s+[a-z]+){0,3}\s*$`)

// listContinuation is what may sit between one attributed span and the next in the same list:
// "the lens's `mint`, `close` and `regrade`".
var listContinuation = regexp.MustCompile(`^(?:,|,?\s+and|,?\s+or|/)?\s*$`)

// seatSpan is one backtick span in a message, and the seat the prose before it hands it to ("" when
// it is unattributed).
type seatSpan struct {
	text string
	seat string
}

// spansIn reads a message's backtick spans with their attribution. carried is the seat a list
// continues: set by "the chair's `a`", kept across ", `b` and `c`", dropped at the first other
// prose.
func spansIn(text string) []seatSpan {
	var out []seatSpan
	carried := ""
	prevEnd := 0
	for _, m := range backtickSpan.FindAllStringSubmatchIndex(text, -1) {
		here := ""
		if a := attributed.FindStringSubmatch(text[:m[0]]); a != nil {
			here = strings.ToLower(a[1])
		} else if carried != "" && listContinuation.MatchString(text[prevEnd:m[0]]) {
			here = carried
		}
		prevEnd = m[1]
		carried = here
		out = append(out, seatSpan{text: text[m[2]:m[3]], seat: here})
	}
	return out
}

// seatRoles is the order every walk over the trees takes. A map range picked a different tree's
// copy of a shared verb from run to run, so a flag check could pass on one run and fail the next.
var seatRoles = []string{"lens", "chair", "blue", "bench"}

// roleTree is one role's command tree.
type roleTree struct {
	role string
	root *cobra.Command
}

// seatTrees is every seat's command tree and then the operator's, in seatRoles order.
func seatTrees() []roleTree {
	var trees []roleTree
	for _, role := range seatRoles {
		trees = append(trees, roleTree{role, NewRootFor(record.SampleSeatOf(role))})
	}
	return append(trees, roleTree{record.OperatorRole, NewRootFor(record.OperatorRole)})
}

// treeOf is the named role's tree, or nil.
func treeOf(trees []roleTree, role string) *cobra.Command {
	for _, t := range trees {
		if t.role == role {
			return t.root
		}
	}
	return nil
}

// spanCheck is one span resolved on one tree: how many path words the tree consumed (0 when the
// first word is no command there), and the problem with the rest — a path word the command has no
// subcommand for, an argument it refuses, a flag it does not register — or "" when the WHOLE span
// is something this tree accepts.
type spanCheck struct {
	consumed int
	problem  string
}

// checkOnTree resolves words on one tree, all of them: a span is sound for a reader only when its
// whole command path is on that reader's tree and every `--flag` in it is on the command it reaches.
// A prefix is not enough — `motion docket rule --id G1` begins with `motion`, which a lens holds,
// and the lens that obeys it is told `docket` has no `rule` for it.
func checkOnTree(root *cobra.Command, words []string, known map[string]bool) spanCheck {
	cmd, n := resolveIn(root, words)
	if n == 0 {
		return spanCheck{}
	}
	rest := words[n:]
	if len(rest) > 0 && strings.HasPrefix(rest[0], "<") && cmd.HasSubCommands() {
		// A template over a group's subcommands (`motion <subject> rule --id <id>`): the flags
		// belong to whichever leaf the placeholder stands for, and every leaf has its own page.
		return spanCheck{consumed: n}
	}
	// Positionals before the first flag are the command's own to accept.
	var positionals []string
	for len(rest) > 0 && !strings.HasPrefix(rest[0], "--") && isWord(rest[0]) {
		positionals = append(positionals, rest[0])
		rest = rest[1:]
	}
	if len(positionals) > 0 {
		if cmd.HasSubCommands() {
			return spanCheck{n, "`" + strings.Join(positionals, " ") + "` is not a subcommand of `" + cmd.CommandPath() + "`"}
		}
		if cmd.Args != nil {
			if err := cmd.Args(cmd, positionals); err != nil {
				return spanCheck{n, "`" + cmd.CommandPath() + "` refuses the argument(s) `" + strings.Join(positionals, " ") + "`: " + err.Error()}
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
				return spanCheck{n, "`--" + name + "` is not a flag of `" + cmd.CommandPath() + "` (another verb has it)"}
			}
			return spanCheck{n, "`--" + name + "` is a flag no command registers"}
		}
	}
	return spanCheck{consumed: n}
}

// holdersOf is every SEAT role whose tree accepts the whole span, in seatRoles order.
func holdersOf(trees []roleTree, words []string, known map[string]bool) []string {
	var holders []string
	for _, t := range trees {
		if t.role == record.OperatorRole {
			continue
		}
		if c := checkOnTree(t.root, words, known); c.consumed > 0 && c.problem == "" {
			holders = append(holders, t.role)
		}
	}
	return holders
}

// anyTreeNames says whether the span's first word is a command on any tree at all — the line
// between a command reference and prose that happens to sit in backticks.
func anyTreeNames(trees []roleTree, words []string) bool {
	for _, t := range trees {
		if _, n := resolveIn(t.root, words); n > 0 {
			return true
		}
	}
	return false
}

// attributedProblem checks a span the prose hands to a named seat against THAT seat's tree. It is
// "" when the seat holds the whole span, and for a span no tree names at all (prose).
func attributedProblem(trees []roleTree, seat string, words []string, known map[string]bool) string {
	if !anyTreeNames(trees, words) {
		return ""
	}
	c := checkOnTree(treeOf(trees, seat), words, known)
	switch {
	case c.consumed == 0:
		holders := holdersOf(trees, words, known)
		if len(holders) == 0 {
			return "is attributed to the " + seat + ", which has no such command — and no seat holds it whole"
		}
		return "is attributed to the " + seat + ", which has no such command (it is " + holderPhrase(holders) + ")"
	case c.problem != "":
		return "is attributed to the " + seat + ", and on the " + seat + "'s surface " + c.problem
	}
	return ""
}

var (
	seatTreesOnce   sync.Once
	seatTreesShared []roleTree
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
	seatKnownOnce.Do(func() { seatKnown = knownFlagNames(t) })
	if p := refusalProblems(seatTreesShared, own, err.Error(), seatKnown); len(p) > 0 {
		t.Errorf("seat %s was refused with %s.\n\nrefusal: %v", seatID, strings.Join(p, "; and "), err)
	}
}

var (
	seatKnownOnce sync.Once
	seatKnown     map[string]bool
)

// refusalProblems reads one refusal against the tree of the seat that received it (own), and
// returns each span that seat cannot obey as written. An attributed span must be on the NAMED
// seat's tree, whole; an unattributed one must be on own's, whole — path and every flag.
func refusalProblems(trees []roleTree, own *cobra.Command, text string, known map[string]bool) []string {
	var out []string
	for _, s := range spansIn(text) {
		words := strings.Fields(s.text)
		if len(words) == 0 || words[0] == "feov-record" || strings.HasPrefix(words[0], "-") {
			continue
		}
		if s.seat != "" {
			if p := attributedProblem(trees, s.seat, words, known); p != "" {
				out = append(out, "`"+s.text+"`, which "+p)
			}
			continue
		}
		c := checkOnTree(own, words, known)
		if c.consumed > 0 && c.problem == "" {
			continue
		}
		if !anyTreeNames(trees, words) {
			continue // not a command anywhere: prose, or a dead path TestEveryCommandPathNamedInAStringResolves reports
		}
		holders := holdersOf(trees, words, known)
		switch {
		case len(holders) > 0:
			out = append(out, "`"+s.text+"`, which is not on its surface (it is "+holderPhrase(holders)+") — a seat that obeys it is told no such command exists. Attribute it in prose (\"the "+holders[0]+"'s `"+s.text+"`\") or name the seat's own act")
		case c.consumed > 0:
			out = append(out, "`"+s.text+"`, which this seat cannot type as written: "+c.problem)
		default:
			out = append(out, "`"+s.text+"`, which is not on its surface and which no seat holds whole")
		}
	}
	return out
}

func TestEveryCommandPathNamedInAStringResolves(t *testing.T) {
	trees := seatTrees()
	known := knownFlagNames(t)

	var spans, references int
	for _, lit := range moduleLiterals(t) {
		role := roleOfDir(lit.pkg)
		for _, s := range spansIn(lit.text) {
			spans++
			problem, isRef := resolveSpan(trees, role, known, s.text, s.seat)
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
// A span resolves on ONE NAMED TREE wherever the source says whose it is, and only a shared,
// unattributed span falls back to the whole set:
//   - attributed in prose ("the chair's `carry`"): on the named seat's tree, path and flags;
//   - in a role package: on that role's tree, and a verb another seat holds is a dead path for this
//     reader unless the prose attributes it;
//   - shared and unattributed: the reader is not knowable from the source — `show`'s listing says
//     which verbs WRITE a view, to every seat — so some tree must accept it whole, and
//     assertRefusalSpeaksToItsSeat holds every refusal a test delivers to the seat that received it.
//
// Trees are walked in seatRoles order, so which copy of a shared verb a flag is checked against
// never depends on a map's iteration order.
func resolveSpan(trees []roleTree, role string, known map[string]bool, span, attributedTo string) (problem string, isRef bool) {
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
		role, attributedTo = "", "" // the operator's spelling reaches whichever tree the --seat-id selects
	}
	if len(words) == 0 {
		return "", isRef
	}
	if staleRoles[words[0]] && len(words) > 1 && isWord(words[1]) {
		return "begins with the role word `" + words[0] + "`, which left the command path when the surface became seat-scoped — a seat types the verb alone, and prose hands another seat's verb to it as \"the " + words[0] + "'s `" + strings.Join(words[1:], " ") + "`\"", true
	}
	if !anyTreeNames(trees, words) {
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
	if attributedTo != "" {
		return attributedProblem(trees, attributedTo, words, known), true
	}
	if role != "" {
		c := checkOnTree(treeOf(trees, role), words, known)
		if c.consumed > 0 && c.problem == "" {
			return "", true
		}
		if holders := holdersOf(trees, words, known); len(holders) > 0 {
			return "is not on the " + role + "'s surface (it is " + holderPhrase(holders) + ") — a " + role + " reading this has no such command; attribute it in prose (\"the " + holders[0] + "'s `" + span + "`\") if it is another seat's act", true
		}
		if c.consumed > 0 {
			return c.problem, true
		}
	}
	// Shared and unattributed, or a role-package span no tree accepts whole: the problem the
	// deepest-reaching tree reports, the first in seatRoles order on a tie.
	best := spanCheck{}
	for _, t := range trees {
		c := checkOnTree(t.root, words, known)
		if c.consumed > 0 && c.problem == "" {
			return "", true
		}
		if c.consumed > best.consumed {
			best = c
		}
	}
	return best.problem, true
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

// THE GATE ITSELF, on the refusals it must refuse. Each case is a span a seat could not obey as
// written; a gate that read only the first word, never the flags, or took an attribution on
// trust passed every one of them.
func TestARefusalASeatCannotObeyFailsTheGate(t *testing.T) {
	trees := seatTrees()
	known := knownFlagNames(t)
	lens := treeOf(trees, "lens")
	for _, c := range []struct {
		name, text string
		bad        bool
	}{
		{"another seat's path under a group the lens holds", "rule it with `motion docket rule --id G1`", true},
		{"the lens's own verb with the chair's flag", "use `close --carried-from <epoch>` instead", true},
		{"an attribution to a seat that does not hold the verb", "the lens's `carry --carried-from <epoch>`", true},
		{"an attribution whose flag is not on the named seat's verb", "the chair's `carry --anchor C-00000001`", true},
		{"an attribution, capitalised, to the holder", "The Chair's `carry --carried-from <epoch>`", false},
		{"a list continuing an attribution", "the bench's `motion docket rule` and `inquest debate`", false},
		{"the lens's own verb, whole", "close it with `close --id G1 --as repaired`", false},
		{"another seat's verb, attributed to its holder", "the bench's `motion docket rule --id G1`", false},
		{"prose in backticks", "the `awaiting_docket` field", false},
	} {
		got := refusalProblems(trees, lens, c.text, known)
		if (len(got) > 0) != c.bad {
			t.Errorf("%s: %q to a lens — problems %q, want a problem: %v", c.name, c.text, got, c.bad)
		}
	}
}

// THE UNKNOWN-VIEW REFUSAL, DELIVERED TO EVERY SEAT. It carries its group's listing, and each row
// says which verbs write that view — so it is a refusal in every seat's terms at once, and only a
// test that delivers it to each seat lets the harness read it against each tree. The inquest
// listing reached only the bench and the chair, and no test sent it to the bench, so its
// "Written by `position`, `closing`" — verbs the bench does not hold — went unread.
func TestTheUnknownViewRefusalSpeaksToEverySeat(t *testing.T) {
	runDir := seatRun(t)
	for _, role := range seatRoles {
		id := record.SampleSeatOf(role)
		groups := []string{"show"}
		if cmdAt(NewRootFor(id), []string{"inquest"}) != nil {
			groups = append(groups, "inquest")
		}
		for _, g := range groups {
			// run hands the refusal to assertRefusalSpeaksToItsSeat with this seat's tree.
			_, err := run(t, g, "nonesuch", "--run", runDir, "--seat-id", id)
			if err == nil || !strings.Contains(err.Error(), "nonesuch") {
				t.Errorf("%s: `%s nonesuch` was not refused about the view it named: %v", role, g, err)
			}
		}
	}
}
