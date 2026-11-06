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
	"strconv"
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
	rec       *Recorder
	restore   func()
	dirs      []string          // the source packages, as given
	excluded  map[string]string // census statements no run can plan, with the reason
	unindexed map[string]string // refused findings allowed by name, with the reason
	golden    string            // the directory the goldens live in
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
	// Unindexed is each refused finding (Shape.Refused: a walk of, or an automatic index on, a
	// growing table the statement reads itself) that no index can serve, keyed by its whole
	// baseline line and valued by the reason. No baseline holds such a finding; this table is the
	// one way past the refusal, so each entry says in code review why the index cannot exist. An
	// entry whose finding no longer occurs fails the run, so the table cannot outlive its reason.
	Unindexed map[string]string
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
		return nil, errors.New("SuiteConfig.Route is nil: nothing would route the suite's records through the query-plan guard")
	}
	return &Suite{rec: rec, restore: cfg.Route(name), dirs: cfg.Dirs, excluded: cfg.Excluded, unindexed: cfg.Unindexed, golden: "testdata"}, nil
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
	var shapes []finding
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
			shapes = append(shapes, finding{label: tg.label, shape: sh})
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
#   <view NAME | stmt ID> | <kind> | <name> | x<plan nodes showing it> | <whose cost: table, or via V>
# A stmt ID is the head of the SHA-256 of the statement's whitespace-normalised SQL; plans.golden
# prints the SQL under it. GENERATED by UPDATE_GOLDENS=1 go test ./internal/record — never edited
# by hand. Every line here is the cost of a view or a CTE ("via"). A walk of, or an automatic index
# on, a growing table the statement reads itself ("table") is a missing index, and no line here
# holds one. The record suite's TestMain fails on a finding not listed here and on a line here that
# no longer occurs. UPDATE_GOLDENS=1 writes only what the baseline already holds, less what the
# fix removed; a new line takes PLANGUARD_ACCEPT_VIEW_COST=1 as well, and shows in the diff.
`

// acceptViewCost is the environment variable that lets UPDATE_GOLDENS=1 add a line to the shapes
// baseline. Exact "1", as UPDATE_GOLDENS is.
const acceptViewCost = "PLANGUARD_ACCEPT_VIEW_COST"

// finding is one shape in one planned target, as the baseline names it.
type finding struct {
	label string
	shape Shape
}

func (f finding) line() string { return f.label + " | " + f.shape.String() }

// key is the line without its count, so a count that falls is told from a finding that is new.
func (f finding) key() string {
	return f.label + " | " + string(f.shape.Kind) + " | " + f.shape.Name + " | " + f.shape.Where
}

// baselineLine is one line of the shapes baseline, read back into its parts.
type baselineLine struct {
	key   string
	count int
	where string
	kind  ShapeKind
}

// parseBaselineLine reads "<label> | <kind> | <name> | x<n> | <where>". The name may itself hold
// " | " (an ambiguous alias lists its candidates with "|"), so the parts are read from both ends.
func parseBaselineLine(line string) (baselineLine, error) {
	parts := strings.Split(line, " | ")
	if len(parts) < 5 {
		return baselineLine{}, fmt.Errorf("the baseline line %q has %d of its five fields", line, len(parts))
	}
	n := len(parts)
	count, err := strconv.Atoi(strings.TrimPrefix(parts[n-2], "x"))
	if err != nil || !strings.HasPrefix(parts[n-2], "x") {
		return baselineLine{}, fmt.Errorf("the baseline line %q has no count x<n> before its last field", line)
	}
	label, kind, where := parts[0], parts[1], parts[n-1]
	name := strings.Join(parts[2:n-2], " | ")
	return baselineLine{key: label + " | " + kind + " | " + name + " | " + where, count: count, where: where, kind: ShapeKind(kind)}, nil
}

// refusal is the message for a finding no baseline holds. It is what the author of the statement
// reads, so it says what to change and where.
func refusal(line string) string {
	return "the statement walks, or indexes on the fly, a growing table it reads itself — a missing index, which no baseline holds:\n  " + line + "\n" +
		"Add the index its plan needs to QueryIndexDDL in internal/record/recordsql/schema.go (EnvelopeDDL for events), " +
		"bump eventSchema in plugins/frank-exchange-of-views/requirements.json, run go -C scripts run ./schemagen and " +
		"restamp recordsql's goldens with UPDATE_GOLDENS=1 — or rewrite the statement to read an index the table has. " +
		"The plan is under its id in plans.golden. Where no index can serve the question, the Unindexed table in the " +
		"suite's TestMain takes the line with the reason."
}

// compareShapes holds the findings to the baseline and to the refusal.
//
// A REFUSED FINDING IS NEVER BASELINED. It is a missing index (Shape.Refused); a baseline that
// held it would let the next one land the same way. It fails the run whether or not the baseline
// names it, UPDATE_GOLDENS=1 does not write it, and a baseline line marked OnTable fails on read.
// The suite's Unindexed table is the only way past, one reasoned entry at a time, and an entry
// whose finding is gone fails.
//
// THE BASELINE ONLY SHRINKS UNDER UPDATE_GOLDENS=1 ALONE. A regeneration run by reflex — after an
// unrelated golden moved — would otherwise add a view's new cost with nothing but a diff line to
// show for it. Growth takes PLANGUARD_ACCEPT_VIEW_COST=1 as well: the author asks for it by name,
// and the added line is in the diff the reviewer reads. A count that falls is a shrink.
func (s *Suite) compareShapes(found []finding) []error {
	path := filepath.Join(s.golden, "shapes.golden")
	var errs []error
	var baselinable []finding
	allowed := map[string]bool{}
	for _, f := range found {
		line := f.line()
		if !f.shape.Refused() {
			baselinable = append(baselinable, f)
			continue
		}
		if _, ok := s.unindexed[line]; ok {
			allowed[line] = true
			continue
		}
		errs = append(errs, errors.New(refusal(line)))
	}
	var stale []string
	for line := range s.unindexed {
		if !allowed[line] {
			stale = append(stale, line)
		}
	}
	sort.Strings(stale)
	for _, line := range stale {
		errs = append(errs, fmt.Errorf("the Unindexed allowance %q names no refused finding this run made — the index landed or the statement changed; remove the entry", line))
	}
	sort.Slice(baselinable, func(i, j int) bool { return baselinable[i].line() < baselinable[j].line() })

	raw, err := os.ReadFile(path)
	if err != nil && !(goldentest.Update && errors.Is(err, os.ErrNotExist)) {
		return append(errs, fmt.Errorf("no findings baseline: %v\nGenerate it with UPDATE_GOLDENS=1 PLANGUARD_ACCEPT_VIEW_COST=1 go test ./internal/record and read it", err))
	}
	want := map[string]baselineLine{}
	for _, line := range strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		b, err := parseBaselineLine(line)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if b.where == OnTable && (Shape{Kind: b.kind, Where: b.where}).Refused() {
			errs = append(errs, fmt.Errorf("%s holds a missing index, which no baseline may:\n  %s\nRemove the line; %s", path, line, refusal(line)))
			continue
		}
		want[b.key] = b
	}

	if goldentest.Update {
		grow := os.Getenv(acceptViewCost) == "1"
		var keep []string
		for _, f := range baselinable {
			if b, ok := want[f.key()]; ok && f.shape.Count <= b.count || grow {
				keep = append(keep, f.line())
				continue
			}
			errs = append(errs, fmt.Errorf("the baseline would grow by a view's cost, and UPDATE_GOLDENS=1 alone only shrinks it:\n  %s\n"+
				"Make the view or the statement cheaper, or accept the cost: rerun with %s=1 as well and say why in the commit; the line shows in the diff.", f.line(), acceptViewCost))
		}
		if err := os.WriteFile(path, []byte(shapesHeader+strings.Join(keep, "\n")+"\n"), 0o644); err != nil {
			errs = append(errs, err)
		}
		return errs
	}

	got := map[string]bool{}
	for _, f := range baselinable {
		got[f.key()] = true
		b, ok := want[f.key()]
		if ok && f.shape.Count < b.count {
			errs = append(errs, fmt.Errorf("baseline finding SHRANK from x%d:\n  %s\n"+
				"Regenerate with UPDATE_GOLDENS=1 go test ./internal/record so the baseline shrinks with the fix.", b.count, f.line()))
			continue
		}
		if !ok || f.shape.Count > b.count {
			errs = append(errs, fmt.Errorf("NEW plan-shape finding, not in %s:\n  %s\n"+
				"It is the cost of what it is read through; make that cheaper, or accept it with UPDATE_GOLDENS=1 %s=1 go test ./internal/record. The plan is in plans.golden.", path, f.line(), acceptViewCost))
		}
	}
	var gone []string
	for k, b := range want {
		if !got[k] {
			gone = append(gone, k+" (x"+strconv.Itoa(b.count)+")")
		}
	}
	sort.Strings(gone)
	for _, w := range gone {
		errs = append(errs, fmt.Errorf("baseline finding NO LONGER OCCURS:\n  %s\n"+
			"Regenerate with UPDATE_GOLDENS=1 go test ./internal/record so the baseline shrinks with the fix.", w))
	}
	return errs
}
