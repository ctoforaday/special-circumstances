package fuzz

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// A SURFACE A SEAT READS IS AGENT-FACING WHEREVER IT LIVES — INCLUDING IN A GO STRING.
//
// The prompt gates walk `skills/`, `agents/` and `commands/`. The help gates walk the cobra tree.
// Between them sits a third kind of carrier nobody was looking at: text a Go constant writes into
// a file the seat then opens.
//
// Measured, 2026-08-13. `.records-elsewhere` — the marker the record separation drops in the run
// directory, whose entire purpose is to tell a seat how to reach a board it cannot see — said:
//
//	Read the board the way a seat does:  feov-record show --run <this directory> --view board
//
// `--view` stopped existing when `show` became a group. Every prompt had been corrected, both
// prompt gates were green, and two probe seats read the marker, followed it, and got
// `unknown flag: --view`. The file exists to unblock a seat and blocked one instead.
//
// So this walks the STRING LITERALS of the tree — not the source text, which would flag every
// comment that explains a rename, including the ones written next to this fix. A literal is
// something the program can emit; a comment is something only a human reads.
//
// It runs the same backward check the help-text gate does, and for the same reason: a forward
// match on `show <token>` cannot tell `show board` from `show the seat`, while a roster of names
// that USED to be projections has no false positives and catches exactly the regression that
// happens — a rename that leaves a carrier behind.
func TestNoStringLiteralNamesARetiredSurface(t *testing.T) {
	retired := retiredSurfaces

	// repotree.ToolSources both FINDS the tree without counting `..` from this file's own
	// location and REFUSES an empty result. Both halves matter to a gate shaped like this one:
	// every assertion below is a negative, so a walk that reached nothing would report a clean
	// tree in the same words it uses for a clean tree.
	//
	// It skips _test.go for us: test files construct retired spellings on purpose — this file
	// names four of them in its own patterns, and the golden fixtures replay old command lines.
	sources, err := repotree.ToolSources()
	if err != nil {
		t.Fatal(err)
	}
	var hits []string
	for _, path := range sources {
		// capability.go is a HISTORY, and the same exemption applies for the same reason (#407).
		// Its entries say what a binary AT AN OLD VERSION cannot do, and they are printed by
		// setup's preflight to an operator whose binary IS at that version — where `--view` and
		// `show ledger` still exist. Naming them is accurate for the only reader who sees them.
		//
		// This guard's own remedy ("move the wording into a comment") is right for a literal
		// EMITTED AT a seat and wrong here: the wording is the payload, and burying it in a
		// comment is exactly the unreachable prose #407 removed.
		if strings.HasSuffix(filepath.ToSlash(path), "internal/record/capability.go") {
			continue
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			t.Fatal(perr)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			for _, re := range retired {
				if m := re.FindString(lit.Value); m != "" {
					hits = append(hits, filepath.ToSlash(path)+": "+strings.TrimSpace(m))
					break
				}
			}
			return true
		})
	}
	sort.Strings(hits)
	if len(hits) > 0 {
		t.Errorf("%d string literal(s) name a surface that no longer exists:\n  %s\n\n"+
			"A seat that reads one and follows it is refused for a mistake the tool made. If the literal is\n"+
			"deliberately historical, move the wording into a comment — a comment is read by people, a\n"+
			"literal is emitted at people.",
			len(hits), strings.Join(hits, "\n  "))
	}
}
