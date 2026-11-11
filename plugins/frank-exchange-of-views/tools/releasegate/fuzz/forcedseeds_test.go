package fuzz

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// THE FORCED SEEDS, ONE TABLE. Each seed below is the run a drive forces through its path; the why of
// each is at its drive, in the file named beside it. Every forced seed is declared here and nowhere
// else, and TestForcedSeedsAreOneTable holds that over the whole package: two drives on one seed force
// one run through both, and each then fails the other's oracle on a draw (partsSeed and
// positionLensSeed shared 10, fateTurnSeed and positionGapsSeed 11, and the sweep failed seed 10 on a
// forced-VERIFIED CEILING and seed 11 on its fate turn and its position, by turns).
//
// A new drive takes a free seed in 1..surfaceQuorum, the seeds the sweep runs. sweepPart.owns gives
// seed s to part (s-1) mod 4, so a new seed goes to the part with the fewest forced seeds. A seed of
// 1..6 also runs in the race variant (FUZZ_N=6).
const (
	unverifiedSeed          = 1  // UNVERIFIED terminal verdict (fuzz_test.go)
	estoppelSeed            = 2  // the estoppel conjunction (fuzz_test.go)
	verifiedSeed            = 3  // VERIFIED (fuzz_test.go)
	owedComputationSeed     = 4  // a computation nobody has checked before a plan (fuzz_test.go)
	reopenedComputationSeed = 5  // a reopened computation (fuzz_test.go)
	unsoundComputationSeed  = 6  // a standing unsound judgement a plan readies for (fuzz_test.go)
	limitSeed               = 7  // the epoch-limit exit; past the race variant's six (fuzz_test.go)
	absentSpanSeed          = 8  // a cite of a span its source does not hold (fuzz_test.go)
	benchOrdersSeed         = 9  // every bench-order arm (benchorders_test.go)
	positionLensSeed        = 10 // a position upheld naming a lens (positionheard_test.go)
	positionGapsSeed        = 11 // a position upheld naming open gaps (positionheard_test.go)
	positionOverruledSeed   = 12 // a position overruled (positionheard_test.go)
	lensPetitionSeed        = 13 // a lens petition's hearing (barrierhearing_test.go)
	bluePetitionSeed        = 14 // blue's petition's hearing (barrierhearing_test.go)
	readingDebtSeed         = 15 // a reading debt's whole life on a cited source (readingdebtseed_test.go)
	mintBareSeed            = 16 // a prescription that would leave a citation bare (fuzz_test.go)
	bareEvidenceSeed        = 17 // a cut that leaves an evidence anchor bare (fuzz_test.go)
	sourceDemandSeed        = 18 // the live-answer demand's refusals (sourcedemandseed_test.go)
	lifecycleSeed           = 19 // the anchor lifecycle's acts and refusals (lifecycleseed_test.go)
	readingRemovalSeed      = 25 // the reading debt's discharges no draw reaches (readingremovalseed_test.go)
	partsSeed               = 24 // every part refusal (parts_test.go)
	fateTurnSeed            = 27 // a lens's turn a changed fate brought (fateturnseed_test.go)
)

// stalenessSeeds are the seeds whose runs are FORCED through the staleness arm (judgedsentence_test.go):
// four consecutive seeds, so each of the sweep's four parts runs exactly one.
var stalenessSeeds = []int64{20, 21, 22, 23}

// forcedSeedsFile is the file that declares every forced seed.
const forcedSeedsFile = "forcedseeds_test.go"

// TestForcedSeedsAreOneTable reads every file of the package, so a forced seed declared anywhere is
// held, not only the ones the table lists: a package-level constant or variable named *Seed or *Seeds
// whose value holds an integer literal is a forced seed. Each lives in forcedSeedsFile, each value is an
// integer literal the sweep runs (1..surfaceQuorum), and no two are equal.
func TestForcedSeedsAreOneTable(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	owner := map[int64]string{}
	inTable := 0
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || (gen.Tok != token.CONST && gen.Tok != token.VAR) {
				continue
			}
			for _, spec := range gen.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, name := range vs.Names {
					if !strings.HasSuffix(name.Name, "Seed") && !strings.HasSuffix(name.Name, "Seeds") {
						continue
					}
					var value ast.Expr
					if i < len(vs.Values) {
						value = vs.Values[i]
					}
					at := fset.Position(name.Pos())
					if file != forcedSeedsFile {
						if holdsIntLiteral(value) {
							t.Errorf("%s: %s is a forced seed declared outside %s; declare it in the table there", at, name.Name, forcedSeedsFile)
						}
						continue // one the table's seeds derive (forcedVerifiedSeeds) is no seed of its own
					}
					lits, other := intLiterals(value)
					if len(lits) == 0 || other {
						t.Errorf("%s: %s is not integer literals; the table holds each seed as one, so this test reads it", at, name.Name)
						continue
					}
					inTable++
					for _, s := range lits {
						if s < 1 || s > surfaceQuorum {
							t.Errorf("%s: %s = %d is outside 1..%d, the seeds the sweep runs", at, name.Name, s, surfaceQuorum)
						}
						if prev, dup := owner[s]; dup {
							t.Errorf("%s: %s = %d is %s's seed too; one run would be forced through both drives", at, name.Name, s, prev)
						}
						owner[s] = name.Name
					}
				}
			}
		}
	}
	if inTable == 0 {
		t.Fatalf("%s declares no forced seed the scan could read", forcedSeedsFile)
	}
}

// holdsIntLiteral is whether e holds an integer literal among its values, an array's length aside.
func holdsIntLiteral(e ast.Expr) bool {
	found := false
	if e != nil {
		ast.Inspect(e, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.ArrayType:
				return false // a length, not a value
			case *ast.BasicLit:
				found = v.Kind == token.INT
			}
			return !found
		})
	}
	return found
}

// intLiterals is every integer literal in e, and whether e holds anything but integer literals in a
// slice or array literal.
func intLiterals(e ast.Expr) (lits []int64, other bool) {
	switch v := e.(type) {
	case *ast.BasicLit:
		if v.Kind != token.INT {
			return nil, true
		}
		n, err := strconv.ParseInt(v.Value, 0, 64)
		if err != nil {
			return nil, true
		}
		return []int64{n}, false
	case *ast.CompositeLit:
		for _, el := range v.Elts {
			l, o := intLiterals(el)
			lits = append(lits, l...)
			other = other || o
		}
		return lits, other
	}
	return nil, true
}
