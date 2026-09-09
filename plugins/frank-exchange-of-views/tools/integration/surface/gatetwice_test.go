package surface

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// A GATE WITH TWO COPIES IS A GATE THAT CAN REPORT GREEN FROM ITS STALE HALF.
//
// MEASURED (#847). Seven static agreement gates existed in BOTH integration/surface and
// releasegate/fuzz. Six were byte-identical; two had drifted, and one of the two was
// TestEveryEnvelopeEnumAgreesWithTheRecord — whose binding table in the fuzz copy still described
// the judge's disposition vocabulary as it stood several changes earlier. Updating the surface
// copy left the fuzz copy passing against the OLD model, in the same words it uses for a real
// pass. Nothing distinguished "this gate agrees" from "this gate agrees with a vocabulary that no
// longer exists".
//
// The duplication is not a design; it is the residue of a move that copied the files and did not
// delete the originals. This gate makes the next such move fail rather than double.
//
// TestMain is excluded BY SIGNATURE and not by name: Go permits exactly one per package, so a
// package that drives a built binary and a package that does not must each have their own. That
// is a structural fact about the language, not an allowlist somebody keeps by hand.
var reTestFunc = regexp.MustCompile(`(?m)^func (Test\w+)\(t \*testing\.T\)`)

func TestNoGateIsDeclaredInTwoPackages(t *testing.T) {
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, "plugins", "frank-exchange-of-views", "tools")
	dirs := []string{
		filepath.Join(base, "integration", "surface"),
		filepath.Join(base, "releasegate", "fuzz"),
	}
	where := map[string][]string{}
	total := 0
	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			t.Fatalf("cannot read %s: %v — a directory this gate cannot see contributes no names, which reads exactly like a package with no duplicates", d, err)
		}
		seen := 0
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(d, e.Name()))
			if err != nil {
				t.Fatalf("cannot read %s: %v", e.Name(), err)
			}
			for _, m := range reTestFunc.FindAllStringSubmatch(string(b), -1) {
				where[m[1]] = append(where[m[1]], filepath.Base(d)+"/"+e.Name())
				seen++
				total++
			}
		}
		if seen == 0 {
			t.Fatalf("found NO test functions in %s — the pattern or the layout moved and this gate is comparing two empty sets, which reads exactly like a pass", d)
		}
	}
	var dup []string
	for name, files := range where {
		if len(files) > 1 {
			sort.Strings(files)
			dup = append(dup, name+" in "+strings.Join(files, " and "))
		}
	}
	sort.Strings(dup)
	if len(dup) > 0 {
		t.Errorf("the same gate is declared in two packages:\n  %s\n\n"+
			"Delete one. Two copies drift, and the stale one still reports a PASS — in the same words the\n"+
			"live one uses — because a gate measuring an old model has nothing to say about being old.\n"+
			"(%d test functions scanned.)", strings.Join(dup, "\n  "), total)
	}
}
