package fuzz

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
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
	// The spellings that stopped existing. Each entry is a rename this tree has already made, and
	// the list is the cheapest possible memory of it.
	retired := []*regexp.Regexp{
		// `show --view <name>` — the flag form, retired when show became a group (0.56.0).
		regexp.MustCompile(`--view\s`),
		// `show --run <dir> show <name>` — the doubled verb the group restructure produced.
		regexp.MustCompile(`\bshow\s+(?:--\S+\s+\S+\s+)+show\b`),
		// Projections that were renamed or retired out of the seat menu.
		regexp.MustCompile(`\bshow\s+(citation-ledger|ledger|archive|changelog|proofs|friction)\b`),
		// The verification grade before it got its own name back (0.60.0).
		regexp.MustCompile(`--trust\s`),
	}

	root := filepath.Join("..", "..", "internal")
	var hits []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Test files construct retired spellings on purpose — this file names four of them in its
		// own patterns, and the golden fixtures replay old command lines.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
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
		return nil
	})
	if err != nil {
		t.Fatal(err)
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
