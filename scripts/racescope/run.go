package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ctoforaday/special-circumstances/scripts/internal/testjson"
)

// verdict is one package's raced run, read from `go test -json`.
type verdict struct {
	pkg     string
	pass    int
	failed  []string
	skipped []string
	missing []string // testjson.Answers.Missing: asked for, not answered
	race    bool
	runErr  error
	output  string
}

func (v verdict) red() bool {
	return v.runErr != nil || len(v.failed) > 0 || len(v.missing) > 0 || v.race
}

// readVerdict reads a package's raced run for the selected names: every selected name must
// answer (testjson.Answers says why), and a race the detector printed is red whatever the
// verdicts say.
func readVerdict(pkg string, names []string, stream []byte, runErr error) verdict {
	s := testjson.Read(stream)
	a := s.Answer(pkg, names)
	v := verdict{pkg: pkg, runErr: runErr, pass: len(a.Pass), failed: a.Fail, skipped: a.Skip, missing: a.Missing}
	// What did not arrive as the package's events — a `go` error, a build failure — comes first.
	v.output = s.Stray
	if p := s.Packages[pkg]; p != nil {
		v.output += p.Output
	}
	v.race = strings.Contains(v.output, "WARNING: DATA RACE")
	return v
}

func runPattern(names []string) string { return "^(" + strings.Join(names, "|") + ")$" }

func setNames(set []selected) []string {
	names := make([]string, len(set))
	for i, s := range set {
		names[i] = s.name
	}
	return names
}

// each runs fn over the packages with a bounded number at once and returns when all are done.
func each(pkgs []string, parallel int, fn func(i int, pkg string)) {
	if parallel < 1 {
		parallel = 1
	}
	var wg sync.WaitGroup
	slots := make(chan struct{}, parallel)
	for i, pkg := range pkgs {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-slots }()
			fn(i, pkg)
		}()
	}
	wg.Wait()
}

// raceRun is the VERDICT run: the set under the race detector, with no coverage. Coverage
// under -race counts with atomics, and the detector reads each atomic as a happens-before
// edge — a planted race reported 3 runs of 3 without a profile is reported 0 of 3 with one.
func raceRun(moduleDir string, d *derived, pkgs []string, timeout string, parallel int) []verdict {
	out := make([]verdict, len(pkgs))
	each(pkgs, parallel, func(i int, pkg string) {
		names := setNames(d.sets[pkg])
		cmd := exec.Command("go", "test", "-race", "-count=1", "-json", "-timeout="+timeout, "-run", runPattern(names), pkg)
		cmd.Dir = moduleDir
		stream, err := cmd.CombinedOutput()
		out[i] = readVerdict(pkg, names, stream, err)
	})
	return out
}

// evidenceRun is the EVIDENCE run: the same set with no race detector and a coverage profile
// over everything its binary links. It answers one question — did the set execute the starts —
// and its verdicts are not the gate's. A run that fails here has still written what it reached,
// and a start it did not reach is reported from the profile.
func evidenceRun(moduleDir string, d *derived, pkgs []string, timeout, dir string, parallel int) (profiles []string, errs []string) {
	profiles = make([]string, len(pkgs))
	failures := make([]string, len(pkgs))
	each(pkgs, parallel, func(i int, pkg string) {
		profile := filepath.Join(dir, strconv.Itoa(i)+".cover")
		profiles[i] = profile
		cmd := exec.Command("go", "test", "-count=1", "-timeout="+timeout, "-covermode=set",
			"-coverpkg="+strings.Join(d.links[pkg], ","), "-coverprofile="+profile,
			"-run", runPattern(setNames(d.sets[pkg])), pkg)
		cmd.Dir = moduleDir
		if out, err := cmd.CombinedOutput(); err != nil {
			failures[i] = fmt.Sprintf("the evidence run of %s failed (%v):\n%s", pkg, err, tail(string(out), 40))
		}
	})
	for _, f := range failures {
		if f != "" {
			errs = append(errs, f)
		}
	}
	return profiles, errs
}

// block is one counted region of a coverage profile.
type block struct {
	sl, sc, el, ec int
	count          int
}

// readProfiles parses `go test -coverprofile` output into blocks per profile file name
// (import path + "/" + base name). A profile that is missing, empty or malformed is an ERROR:
// read as "no blocks", it would report every start unexecuted for the wrong reason, and a
// reader that tolerated it would one day report them all executed.
func readProfiles(paths []string) (map[string][]block, error) {
	blocks := map[string][]block{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("coverage profile not written: %w", err)
		}
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")
		if len(lines) < 2 || !strings.HasPrefix(lines[0], "mode: ") {
			return nil, fmt.Errorf("coverage profile %s holds no blocks", p)
		}
		for _, line := range lines[1:] {
			name, bl, err := parseBlock(line)
			if err != nil {
				return nil, fmt.Errorf("coverage profile %s: %w", p, err)
			}
			blocks[name] = append(blocks[name], bl)
		}
	}
	return blocks, nil
}

// parseBlock reads `name.go:line.col,line.col statements count`.
func parseBlock(line string) (string, block, error) {
	bad := func() (string, block, error) {
		return "", block{}, fmt.Errorf("unreadable line %q", line)
	}
	colon := strings.LastIndex(line, ":")
	if colon < 0 {
		return bad()
	}
	fields := strings.Fields(line[colon+1:])
	if len(fields) != 3 {
		return bad()
	}
	span := strings.Split(fields[0], ",")
	if len(span) != 2 {
		return bad()
	}
	var b block
	var err error
	if b.sl, b.sc, err = lineCol(span[0]); err != nil {
		return bad()
	}
	if b.el, b.ec, err = lineCol(span[1]); err != nil {
		return bad()
	}
	if b.count, err = strconv.Atoi(fields[2]); err != nil {
		return bad()
	}
	return line[:colon], b, nil
}

func lineCol(s string) (int, int, error) {
	l, c, ok := strings.Cut(s, ".")
	if !ok {
		return 0, 0, fmt.Errorf("no column in %q", s)
	}
	line, err := strconv.Atoi(l)
	if err != nil {
		return 0, 0, err
	}
	col, err := strconv.Atoi(c)
	return line, col, err
}

// prove splits the owed starts into those a raced test executed and those none did. A start
// inside no counted block at all is unproven too, and says so: it was not measured.
func prove(owed []start, blocks map[string][]block) (proved []start, unproven []string) {
	for _, s := range owed {
		name := s.pkg + "/" + path.Base(s.file)
		found, ran := false, false
		for _, b := range blocks[name] {
			after := s.line > b.sl || (s.line == b.sl && s.col >= b.sc)
			before := s.line < b.el || (s.line == b.el && s.col < b.ec)
			if after && before {
				found = true
				ran = ran || b.count > 0
			}
		}
		switch {
		case ran:
			proved = append(proved, s)
		case found:
			unproven = append(unproven, s.String()+": no raced test executed this start")
		default:
			unproven = append(unproven, s.String()+": not measured — no coverage block holds this position")
		}
	}
	sort.Strings(unproven)
	return proved, unproven
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
