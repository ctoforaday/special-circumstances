package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// NO DUTY IS AUTHORED TWICE, and this is the gate whose absence let one survive its own retirement.
//
// The seats' log duty existed as five authored copies. `nominal` was retired — clean became DERIVED
// from having sat and filed nothing — and every one of those copies went on demanding an entry whose
// type had been deleted, because a fix applied to the copy you are looking at leaves the other four.
// It had already drifted into three wordings by then, which is the shape this measures: the same
// bullet, reachable by more than one seat, with no single place to change it.
//
// THE UNIT IS THE BULLET, NOT THE LINE, because wrapping hides the duplication. Two of the five
// copies differed only in where they wrapped, so a line-level comparison reported them distinct and
// a reviewer diffing the sources saw nothing.
//
// THERE IS NO ALLOWLIST, deliberately. A duplicated duty already has its remedy in this generator —
// put it in a fragment and `@include` it — so an exception would only be a place to record that
// someone did not. A guard with a hand-kept exception list reproduces the defect one level up.
func TestNoBulletIsAuthoredTwice(t *testing.T) {
	const srcRoot = "src"
	var files []string
	if err := filepath.WalkDir(srcRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".md") {
			files = append(files, p)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(files) < 10 {
		t.Fatalf("read %d source(s) under %s — this gate is measuring the wrong tree", len(files), srcRoot)
	}

	where := map[string][]string{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, bullet := range bulletsOf(string(b)) {
			// Short bullets are headings and one-line pointers; a duty is long prose, and the
			// threshold keeps the gate off incidental phrasing two seats happen to share.
			if len(bullet) < 80 {
				continue
			}
			if !contains(where[bullet], f) {
				where[bullet] = append(where[bullet], f)
			}
		}
	}
	for bullet, fs := range where {
		if len(fs) < 2 {
			continue
		}
		sort.Strings(fs)
		t.Errorf("one bullet is authored in %d sources — %s\n\n  %.150s…\n\n"+
			"Put it in src/fragments/ and `@include` it from each. Five copies of the log duty is how a "+
			"retired rule went on being demanded after its type was deleted: the fix landed on one copy.",
			len(fs), strings.Join(fs, ", "), bullet)
	}
}

var bulletStart = regexp.MustCompile(`^\s*[-*] `)

// bulletsOf joins each bullet's continuation lines, so two copies that wrap differently read as one.
func bulletsOf(text string) []string {
	var out []string
	cur := ""
	flush := func() {
		if cur != "" {
			out = append(out, strings.Join(strings.Fields(cur), " "))
			cur = ""
		}
	}
	for _, l := range strings.Split(text, "\n") {
		switch {
		case bulletStart.MatchString(l):
			flush()
			cur = strings.TrimSpace(l)
		case cur != "" && strings.TrimSpace(l) != "" && !strings.HasPrefix(l, "#") && !strings.HasPrefix(strings.TrimSpace(l), "@include"):
			cur += " " + strings.TrimSpace(l)
		default:
			flush()
		}
	}
	flush()
	return out
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
