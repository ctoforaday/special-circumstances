package planguard

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Suite plans every statement a package's whole test suite issues, and holds the source to it.
//
// THE SUITE IS THE SWEEP. A sweep written for the guard reaches the statements its author thought
// to call, and a statement nobody thought to call — a write path's validation, a second copy of a
// read in another file — was never planned, so it could never be reported. The package's own
// tests already drive the record from every side, writes included; installed for the whole run,
// the guard plans whatever they issue.
//
// AND THE SOURCE IS THE CENSUS. Every statement text the source packages spell (SourceStatements)
// must have been planned by the end of the run, issued by those packages' own code. A statement
// added where no test reaches fails the run by name, rather than going unplanned in silence.
//
// It judges only a WHOLE run. A run narrowed by -run, -skip or -list issues part of the suite, and
// its missing statements would be the filter's, not the source's; such a run is not measured, and
// says so under -v. A failed run is not measured either: its own failure is the news.
type Suite struct {
	rec      *Recorder
	restore  func()
	dirs     []string          // the source packages, as given
	excluded map[string]string // census statements no run can plan, with the reason
	golden   string            // the directory the goldens live in
}

// SuiteConfig is what a package's TestMain tells the guard about itself.
type SuiteConfig struct {
	// Dirs are the source packages the census reads and whose statements are the suite's; the
	// package under test first, as paths relative to its directory.
	Dirs []string
	// Excluded is each census statement (as SourceStatement.String renders it) that no run can
	// plan, with the reason. An entry the census no longer finds fails the run, so the list
	// cannot outlive its statements.
	Excluded map[string]string
	// Route points the record's opens at the named driver and returns the restore. The caller
	// passes recordsql's driver override, so the only code that installs the guard is a test's
	// (TestNoShippedCodeInstallsTheQueryPlanGuard).
	Route func(driver string) (restore func())
}

// StartSuite installs the guard for every record the suite opens from now on.
func StartSuite(cfg SuiteConfig) (*Suite, error) {
	accept, err := issuerFiles(cfg.Dirs)
	if err != nil {
		return nil, err
	}
	rec := NewRecorder()
	rec.AttributeTo(accept)
	name, err := Install(rec)
	if err != nil {
		return nil, err
	}
	if cfg.Route == nil {
		return nil, errors.New("planguard: SuiteConfig.Route is nil — nothing would route the suite's records through the guard")
	}
	return &Suite{rec: rec, restore: cfg.Route(name), dirs: cfg.Dirs, excluded: cfg.Excluded, golden: "testdata"}, nil
}

// issuerFiles is the predicate a frame's file must satisfy to be one of dirs' non-test sources.
// A frame names its file by absolute path, or by import path under -trimpath; both are accepted.
func issuerFiles(dirs []string) (func(string) bool, error) {
	root, mod, err := moduleOf(".")
	if err != nil {
		return nil, err
	}
	want := map[string]bool{}
	for _, d := range dirs {
		abs, err := filepath.Abs(d)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			return nil, err
		}
		want[filepath.ToSlash(abs)] = true
		want[mod+"/"+filepath.ToSlash(rel)] = true
	}
	return func(file string) bool {
		file = filepath.ToSlash(file)
		if strings.HasSuffix(file, "_test.go") {
			return false
		}
		i := strings.LastIndexByte(file, '/')
		return i > 0 && want[file[:i]]
	}, nil
}

// moduleOf finds the go.mod above dir and returns its directory and module path.
func moduleOf(dir string) (root, mod string, err error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	for d := abs; ; d = filepath.Dir(d) {
		b, err := os.ReadFile(filepath.Join(d, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
					return d, strings.TrimSpace(rest), nil
				}
			}
			return "", "", fmt.Errorf("%s/go.mod names no module", d)
		}
		if filepath.Dir(d) == d {
			return "", "", fmt.Errorf("no go.mod above %s", abs)
		}
	}
}

// Finish uninstalls the guard and, on a whole passing run, checks the census and the plans. It
// returns the exit code the run should end with.
func (s *Suite) Finish(code int) int {
	s.restore()
	if code != 0 {
		return code
	}
	if f := narrowed(); f != "" {
		if testingVerbose() {
			fmt.Fprintf(os.Stderr, "planguard: a run narrowed by %s issues part of the suite; the census and the plan goldens were NOT MEASURED\n", f)
		}
		return code
	}
	errs := s.check()
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "planguard: %v\n\n", e)
	}
	if len(errs) > 0 {
		return 1
	}
	return code
}

func narrowed() string {
	for _, name := range []string{"test.run", "test.skip", "test.list"} {
		if f := flag.Lookup(name); f != nil && f.Value.String() != "" {
			return "-" + strings.TrimPrefix(name, "test.")
		}
	}
	return ""
}

func testingVerbose() bool {
	f := flag.Lookup("test.v")
	return f != nil && f.Value.String() != "false"
}

func (s *Suite) check() []error {
	var errs []error
	if pe := s.rec.PlanErrors(); len(pe) > 0 {
		errs = append(errs, fmt.Errorf("%d statement(s) could not be planned, so their findings are unknown:\n%s", len(pe), strings.Join(pe, "\n")))
	}

	var issued []StatementPlan
	var texts []string
	for _, sp := range s.rec.Plans() {
		if sp.Issued {
			issued = append(issued, sp)
			texts = append(texts, NormSQL(sp.SQL))
		}
	}
	if len(issued) == 0 {
		return append(errs, errors.New("the suite issued no statement from its own source through the guard — nothing was planned, and a census or a baseline compared against it would compare nothing"))
	}

	census, err := SourceStatements(s.dirs...)
	if err != nil {
		return append(errs, err)
	}
	found := map[string]bool{}
	var unplanned []string
	for _, st := range census {
		found[st.String()] = true
		if _, ok := s.excluded[st.String()]; ok {
			continue
		}
		if !st.PlannedBy(texts) {
			unplanned = append(unplanned, st.At+"  "+st.String())
		}
	}
	if len(unplanned) > 0 {
		errs = append(errs, fmt.Errorf("%d statement(s) in the source were never planned: no test in this suite reaches them, so no plan of theirs is judged.\n"+
			"Reach each one from a test (plansweep_test.go drives the entry points no other test calls):\n  %s",
			len(unplanned), strings.Join(unplanned, "\n  ")))
	}
	var stale []string
	for k := range s.excluded {
		if !found[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(stale)
	for _, k := range stale {
		errs = append(errs, fmt.Errorf("the census exclusion %q names no statement the source spells any longer — remove it", k))
	}

	targets, an, err := s.targets(issued)
	if err != nil {
		return append(errs, err)
	}
	var golden strings.Builder
	golden.WriteString("# Every view, planned on its own, then every statement this package's code issued during its\n" +
		"# whole test suite, with SQLite's subquery numbering removed. GENERATED by\n" +
		"# UPDATE_GOLDENS=1 go test ./internal/record; read the diff.\n")
	var shapes []string
	for _, tg := range targets {
		golden.WriteString("\n== " + tg.label + "\n")
		if strings.HasPrefix(tg.label, "stmt ") {
			golden.WriteString("-- " + tg.sql + "\n")
		}
		if len(tg.plan.Roots) == 0 {
			golden.WriteString("  (no plan rows: the statement reads no table)\n")
		}
		golden.WriteString(tg.plan.Render())
		for _, sh := range an.Shapes(tg.sql, tg.plan) {
			shapes = append(shapes, tg.label+" | "+sh.String())
		}
	}
	if err := s.assertGolden("plans", golden.String()); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, s.compareShapes(shapes)...)
	return errs
}

// target is one planned statement: a view planned on its own, or a statement the suite issued.
type target struct {
	label, sql string
	plan       Plan
}

// targets is every view planned on its own, then every issued statement by its SQL, and the
// analyzer over the views. The views are planned on a record created for the purpose, so the set
// is the schema's own answer and not whatever the suite's last record held.
func (s *Suite) targets(issued []StatementPlan) ([]target, *Analyzer, error) {
	dir, err := os.MkdirTemp("", "planguard-views-")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(dir)
	db, err := recordsql.Open(filepath.Join(dir, "record.db"))
	if err != nil {
		return nil, nil, err
	}
	defer recordsql.CloseUnder(dir) // before the removal: a handle outliving its directory fails Windows
	ctx := context.Background()
	views, err := Views(ctx, db)
	if err != nil {
		return nil, nil, err
	}
	an, err := NewAnalyzer(views)
	if err != nil {
		return nil, nil, err
	}
	var out []target
	for _, v := range views {
		q := ViewQuery(v.Name)
		p, err := PlanQuery(ctx, db, q)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, target{label: "view " + v.Name, sql: q, plan: p})
	}
	byID := map[string]target{}
	for _, sp := range issued {
		norm := NormSQL(sp.SQL)
		byID[StatementID(norm)] = target{label: "stmt " + StatementID(norm), sql: norm, plan: sp.Plan}
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return byID[ids[i]].sql < byID[ids[j]].sql })
	for _, id := range ids {
		out = append(out, byID[id])
	}
	return out, an, nil
}

// StatementID is the head of the SHA-256 of a statement's NormSQL text: the name shapes.golden
// gives it, under which plans.golden prints the SQL.
func StatementID(norm string) string {
	sum := sha256.Sum256([]byte(norm))
	return hex.EncodeToString(sum[:])[:12]
}

func (s *Suite) assertGolden(name, got string) error {
	path := filepath.Join(s.golden, name+".golden")
	if goldentest.Update {
		if err := os.MkdirAll(s.golden, 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(got), 0o644)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("no %s: %v\nGenerate it with UPDATE_GOLDENS=1 go test ./internal/record and read it", path, err)
	}
	want := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if want == got {
		return nil
	}
	wl, gl := strings.Split(want, "\n"), strings.Split(got, "\n")
	i := 0
	for i < len(wl) && i < len(gl) && wl[i] == gl[i] {
		i++
	}
	line := func(ls []string) string {
		if i < len(ls) {
			return ls[i]
		}
		return "(end of file)"
	}
	return fmt.Errorf("%s differs from the plans this suite recorded, first at line %d:\n  golden: %s\n  run:    %s\n"+
		"Regenerate with UPDATE_GOLDENS=1 go test ./internal/record and read every changed plan.", path, i+1, line(wl), line(gl))
}

const shapesHeader = `# Plan-shape findings the record has today, one per line:
#   <view NAME | stmt ID> | <kind> | <name> | x<plan nodes showing it>
# A stmt ID is the head of the SHA-256 of the statement's whitespace-normalised SQL; plans.golden
# prints the SQL under it. GENERATED by UPDATE_GOLDENS=1 go test ./internal/record — never edited
# by hand. The record suite's TestMain fails on a finding not listed here and on a line here that
# no longer occurs, so a fix removes its line and nothing adds one unread.
`

// compareShapes fails on a finding the baseline does not list and on a baseline line that no
// longer occurs. Under UPDATE_GOLDENS=1 it writes the baseline instead.
func (s *Suite) compareShapes(found []string) []error {
	sort.Strings(found)
	path := filepath.Join(s.golden, "shapes.golden")
	if goldentest.Update {
		if err := os.WriteFile(path, []byte(shapesHeader+strings.Join(found, "\n")+"\n"), 0o644); err != nil {
			return []error{err}
		}
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return []error{fmt.Errorf("no findings baseline: %v\nGenerate it with UPDATE_GOLDENS=1 go test ./internal/record and read it", err)}
	}
	want := map[string]bool{}
	for _, line := range strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n") {
		if line != "" && !strings.HasPrefix(line, "#") {
			want[line] = true
		}
	}
	var errs []error
	got := map[string]bool{}
	for _, f := range found {
		got[f] = true
		if !want[f] {
			errs = append(errs, fmt.Errorf("NEW plan-shape finding, not in %s:\n  %s\n"+
				"Fix the statement or the index it needs; the plan is in plans.golden.", path, f))
		}
	}
	var gone []string
	for w := range want {
		if !got[w] {
			gone = append(gone, w)
		}
	}
	sort.Strings(gone)
	for _, w := range gone {
		errs = append(errs, fmt.Errorf("baseline finding NO LONGER OCCURS:\n  %s\n"+
			"Regenerate with UPDATE_GOLDENS=1 go test ./internal/record so the baseline shrinks with the fix.", w))
	}
	return errs
}
