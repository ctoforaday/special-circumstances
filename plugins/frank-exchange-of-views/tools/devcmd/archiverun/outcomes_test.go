package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

var (
	pinned = map[string]string{"python3": "Python 3.12.16", "node": "v24.19.0"}
	rows   = []Row{{Archive: "a.tar.gz", OldSha: "0b15", Outcome: Reproduced}, {Archive: "a.tar.gz", OldSha: "62d0", Outcome: Moved}, {Archive: "b.tar.gz", OldSha: "ab6a", Outcome: Differs}}
)

// WHAT -write WRITES, THE CHECK ACCEPTS, and only under the interpreters it names.
func TestTheWrittenOutcomesAreTheCheckedOnes(t *testing.T) {
	if p := Compare(Format(pinned, rows), pinned, rows); len(p) != 0 {
		t.Fatalf("the written file disagrees with its own measurement: %q", p)
	}
}

// EVERY DISAGREEMENT IS NAMED, AND A FILE THAT MEASURED NOTHING FAILS: an empty or header-only file
// is NOT MEASURED, never a pass over zero rows.
func TestEveryDisagreementIsNamed(t *testing.T) {
	held := Format(pinned, rows)
	for _, c := range []struct {
		name, held string
		versions   map[string]string
		rows       []Row
		want       string
	}{
		{"an empty file", "", pinned, rows, "NOT MEASURED"},
		{"a header and no rows", Format(pinned, nil), pinned, rows, "NOT MEASURED"},
		{"another interpreter", held, map[string]string{"python3": "Python 3.12.3", "node": "v24.19.0"}, rows, `interpreter python3 is "Python 3.12.3"`},
		{"an absent interpreter", held, map[string]string{"python3": "absent", "node": "v24.19.0"}, rows, `interpreter python3 is "absent"`},
		{"an interpreter the file lacks", held, map[string]string{"python3": "Python 3.12.16", "node": "v24.19.0", "bash": "x"}, rows, "interpreter bash"},
		{"an interpreter the check lacks", held, map[string]string{"python3": "Python 3.12.16"}, rows, "names interpreter node"},
		{"a moved outcome", held, pinned, []Row{rows[0], {Archive: "a.tar.gz", OldSha: "62d0", Outcome: Differs}, rows[2]}, "re-ran differs, the file holds moved"},
		{"a row the file lacks", held, pinned, append(append([]Row{}, rows...), Row{Archive: "c.tar.gz", OldSha: "ffff", Outcome: Reproduced}), "holds no row"},
		{"a row no archive re-ran", held, pinned, rows[:2], "b.tar.gz ab6a: the file holds a row no archive re-ran"},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := Compare(c.held, c.versions, c.rows)
			if !strings.Contains(strings.Join(p, "\n"), c.want) {
				t.Errorf("problems %q, want one naming %q", p, c.want)
			}
		})
	}
}

// A VARIANT ACT'S RE-RUN MAY OR MAY NOT WRITE ITS RECORDED OUTPUT, so either of its non-failing
// outcomes stands for the other; a reproducible act's outcome, and any failing one, is held exactly.
func TestAVariantActsOutcomeIsEitherNonFailingOne(t *testing.T) {
	variant := func(o string) Act { return Act{Outcome: o, Variant: true} }
	fixed := func(o string) Act { return Act{Outcome: o} }
	row := func(acts ...Act) Row {
		set := map[string]bool{}
		for _, a := range acts {
			set[a.Outcome] = true
		}
		return Row{Archive: "a.tar.gz", OldSha: "62d0", Outcome: joinOutcomes(set), Acts: acts}
	}
	for _, c := range []struct {
		name, held string
		measured   Row
		ok         bool
	}{
		{"held moved, re-ran reproduced", Moved, row(variant(Reproduced)), true},
		{"held reproduced, re-ran moved", Reproduced, row(variant(Moved)), true},
		{"held moved, two variant acts split", Moved, row(variant(Moved), variant(Reproduced)), true},
		{"held both, two variant acts agree", Moved + "," + Reproduced, row(variant(Reproduced), variant(Reproduced)), true},
		{"held moved, a reproducible act beside it reproduced", Moved + "," + Reproduced, row(fixed(Reproduced), variant(Reproduced)), true},
		{"held moved, a variant act refused", Moved, row(variant(Refused)), false},
		{"held reproduced, a reproducible act differs", Reproduced, row(fixed(Differs)), false},
		{"held differs, a reproducible act reproduced", Differs, row(fixed(Reproduced)), false},
		{"held reproduced and differs, one variant act", Differs + "," + Reproduced, row(variant(Reproduced)), false},
		{"held both, one variant act", Moved + "," + Reproduced, row(variant(Moved)), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := Compare(Format(pinned, []Row{{Archive: "a.tar.gz", OldSha: "62d0", Outcome: c.held}}), pinned, []Row{c.measured})
			if got := len(p) == 0; got != c.ok {
				t.Errorf("held %q, measured %+v: problems %q, want ok=%v", c.held, c.measured.Acts, p, c.ok)
			}
		})
	}
}

// AN ARCHIVE AND ITS OUTCOMES ARRIVE TOGETHER. The re-run executes every archived program, so it
// runs only in the archive-rerun job, nightly; a tarball committed without its measured rows would
// otherwise pass every pull request and fail that job on every night after. This holds the names on
// every pull request, executing nothing.
func TestEveryArchiveIsMeasured(t *testing.T) {
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	held, err := os.ReadFile(filepath.Join(root, outcomesFile))
	if err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join(root, "run-archive", "*.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no run-archive/*.tar.gz: a check over no archive checks nothing")
	}
	var tarballs []string
	for _, p := range paths {
		tarballs = append(tarballs, filepath.Base(p))
	}
	if p := Unmeasured(string(held), tarballs); len(p) != 0 {
		t.Errorf("%d problem(s):\n  %s", len(p), strings.Join(p, "\n  "))
	}
}

// EACH WAY THE ARCHIVES AND THE FILE PART IS NAMED: a tarball with no row, and rows with no tarball.
func TestEveryUnmeasuredArchiveIsNamed(t *testing.T) {
	held := Format(pinned, rows)
	if p := Unmeasured(held, []string{"a.tar.gz", "b.tar.gz"}); len(p) != 0 {
		t.Fatalf("the archives the file measured: problems %q", p)
	}
	for _, c := range []struct {
		name     string
		tarballs []string
		want     string
	}{
		{"a tarball the file lacks", []string{"a.tar.gz", "b.tar.gz", "c.tar.gz"}, "c.tar.gz: the outcomes file holds no row for this archive"},
		{"rows for a tarball that is gone", []string{"a.tar.gz"}, "b.tar.gz: the outcomes file holds rows for an archive the archive directory does not hold"},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := Unmeasured(held, c.tarballs)
			if len(p) != 1 || !strings.Contains(p[0], c.want) || !strings.Contains(p[0], "dispatch the archive-rerun job") {
				t.Errorf("problems %q, want exactly one naming %q and the remedy", p, c.want)
			}
		})
	}
}
