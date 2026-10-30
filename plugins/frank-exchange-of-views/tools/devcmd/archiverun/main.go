// Command archiverun brings every archived run to this binary's epoch and re-runs every computation
// it holds, comparing each bundle's outcome against the measured outcomes in
// testdata/outcomes.tsv (plans/compute-and-show.md §V V1a.7).
//
// For each run-archive tarball it unpacks the archive (archive.Extract), migrates it
// (migrate.Migrate), executes each recorded bundle once (compute.Rerun) and compares every act on
// it against that execution (compute.Compare), graded by the act's own basis:
//
//	reproduced  the execution wrote the act's recorded output, exit and out/
//	moved       it did not, and the act was recorded variant
//	differs     it did not, and the act was recorded reproducible
//	refused     the bundle could not be executed, or it called the tool
//
// A bundle's row is keyed by the archive and the archived script's sha (the one file's sha256,
// which a migrated bundle keeps), and holds the outcome its acts agree on, or each distinct
// outcome, comma-separated, where they do not.
//
// THE OUTCOMES ARE MEASURED, NEVER FORECAST. They depend on the interpreters the programs run
// under, so the file's header names each interpreter's version and a check under any other
// version fails before a single row is compared. The file is written by `-write` in the
// archive-rerun job of .github/workflows/hooks.yml, on workflow_dispatch, where the interpreters
// are pinned; the job uploads it, and it is committed from that artifact. Off that image the check
// fails by design: a dev box runs `-migrate-only`.
//
// IT RUNS EVERY ARCHIVED PROGRAM with this process's privileges, filesystem and network. In CI it
// runs only in its own job, which holds no write credential and saves no cache.
//
// Run from the tools module; a relative -archives resolves against the repository root:
//
//	go run ./devcmd/archiverun -migrate-only
//	go run ./devcmd/archiverun            # the check, on the pinned image
//	go run ./devcmd/archiverun -write     # the measurement, on the pinned image
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/archive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// outcomesFile is the measured outcomes, relative to the repository root.
const outcomesFile = "plugins/frank-exchange-of-views/tools/devcmd/archiverun/testdata/outcomes.tsv"

func main() {
	archives := flag.String("archives", "run-archive", "the directory of run tarballs (*.tar.gz); relative to the repository root unless absolute")
	outcomes := flag.String("outcomes", outcomesFile, "the measured outcomes; relative to the repository root unless absolute")
	write := flag.Bool("write", false, "measure: re-run every computation and write the outcomes file, its header naming the interpreters measured under")
	migrateOnly := flag.Bool("migrate-only", false, "unpack and migrate every archive and check what migration asserts; execute no program")
	flag.Parse()
	if *write && *migrateOnly {
		fail(errors.New("-write measures re-runs and -migrate-only executes none: give one"))
	}
	root, err := repotree.Root()
	if err != nil {
		fail(err)
	}
	if !filepath.IsAbs(*archives) {
		*archives = filepath.Join(root, *archives)
	}
	if !filepath.IsAbs(*outcomes) {
		*outcomes = filepath.Join(root, *outcomes)
	}
	mode := check
	switch {
	case *write:
		mode = measure
	case *migrateOnly:
		mode = migrateOnlyMode
	}
	if err := run(*archives, *outcomes, mode, os.Stdout); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "archiverun:", err)
	os.Exit(1)
}

type mode int

const (
	check mode = iota
	measure
	migrateOnlyMode
)

func run(archives, outcomes string, m mode, w io.Writer) error {
	tarballs, err := filepath.Glob(filepath.Join(archives, "*.tar.gz"))
	if err != nil {
		return err
	}
	if len(tarballs) == 0 {
		return fmt.Errorf("no *.tar.gz under %s", archives)
	}
	sort.Strings(tarballs)
	work, err := os.MkdirTemp("", "archiverun-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	var problems []string
	var rows []Row
	computations := 0
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		dir := filepath.Join(work, strings.TrimSuffix(name, ".tar.gz"))
		to, n, ps, err := migrated(tb, dir)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		computations += n
		problems = append(problems, ps...)
		if m != migrateOnlyMode && len(ps) == 0 {
			rs, err := rerun(name, to)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			for _, r := range rs {
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Archive, r.OldSha, r.Outcome)
			}
			rows = append(rows, rs...)
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "migrated %d archive(s): %d computation(s), %d problem(s)\n", len(tarballs), computations, len(problems))
	if computations == 0 {
		problems = append(problems, "no archive migrated a computation — a pass that checked nothing")
	}
	if m == migrateOnlyMode || len(problems) > 0 {
		return joined(problems)
	}
	fmt.Fprintln(w, summary(rows))
	versions := Interpreters()
	if m == measure {
		if err := os.MkdirAll(filepath.Dir(outcomes), 0o755); err != nil {
			return err
		}
		return os.WriteFile(outcomes, []byte(Format(versions, rows)), 0o644)
	}
	held, err := os.ReadFile(outcomes)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("NOT MEASURED: %s does not exist. The outcomes are measured on the pinned image: dispatch the archive-rerun job, download its outcomes artifact and commit it", outcomes)
	}
	if err != nil {
		return err
	}
	return joined(Compare(string(held), versions, rows))
}

func joined(problems []string) error {
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%d problem(s):\n  %s", len(problems), strings.Join(problems, "\n  "))
}

// migrated unpacks one tarball and migrates it into dir/to, and returns the migrated run, how many
// computations it holds, and what migration asserts and did not hold: a refused event, or an
// artifact the source lacked — every act must have its bundle's dir/ and its own output.
func migrated(tarball, dir string) (to string, computations int, problems []string, err error) {
	name := filepath.Base(tarball)
	src, to := filepath.Join(dir, "src"), filepath.Join(dir, "to")
	if err := archive.Extract(tarball, src); err != nil {
		return "", 0, nil, err
	}
	res, merr := migrate.Migrate(src, to, migrate.Entries(), migrate.Options{})
	if cerr := recordsql.CloseUnder(to); cerr != nil && merr == nil {
		merr = cerr
	}
	if merr != nil {
		return "", 0, []string{fmt.Sprintf("%s: migrate: %v", name, merr)}, nil
	}
	for _, r := range res.Refusals {
		problems = append(problems, fmt.Sprintf("%s: refused source event %d (%s): %s", name, r.OldID, r.Word, r.Err))
	}
	for _, a := range res.MissingArtifacts {
		problems = append(problems, fmt.Sprintf("%s: the source lacks %s", name, a))
	}
	run, err := record.OpenRun(to)
	if err != nil {
		return "", 0, nil, err
	}
	defer recordsql.CloseUnder(to)
	fam, err := record.FamilyOf(run)
	if err != nil {
		return "", 0, nil, err
	}
	for _, e := range fam.Events {
		c, ok := recordpb.BodyAs[*recordpb.Computation](e)
		if !ok {
			continue
		}
		computations++
		if _, err := compute.Bundle(to, c.GetBundleSha()); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %s has no dir/: %v", name, c.GetComputationId(), err))
		}
		if _, err := compute.ActOutput(to, c.GetBundleSha(), c.GetComputationId()); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %s has no acts/<K-id>/output: %v", name, c.GetComputationId(), err))
		}
	}
	return to, computations, problems, nil
}

// rerun executes each recorded bundle of a migrated run once and grades every act on it.
func rerun(archiveName, runDir string) ([]Row, error) {
	run, err := record.OpenRun(runDir)
	if err != nil {
		return nil, err
	}
	defer recordsql.CloseUnder(runDir)
	fam, err := record.FamilyOf(run)
	if err != nil {
		return nil, err
	}
	type bundle struct {
		sha  string
		argv []string
		acts []*recordpb.Computation
	}
	var order []*bundle
	of := map[string]*bundle{}
	for _, e := range fam.Events {
		c, ok := recordpb.BodyAs[*recordpb.Computation](e)
		if !ok {
			continue
		}
		b := of[c.GetBundleSha()]
		if b == nil {
			b = &bundle{sha: c.GetBundleSha(), argv: c.GetArgv()}
			of[b.sha] = b
			order = append(order, b)
		}
		b.acts = append(b.acts, c)
	}
	var rows []Row
	for _, b := range order {
		old, err := archivedSha(runDir, b.sha)
		if err != nil {
			return nil, err
		}
		outcomes := map[string]bool{}
		e, rerr := compute.Rerun(runDir, b.sha, b.argv)
		for _, a := range b.acts {
			outcomes[grade(runDir, b.sha, a, e, rerr)] = true
		}
		rows = append(rows, Row{Archive: archiveName, OldSha: old, Outcome: joinOutcomes(outcomes)})
	}
	return rows, nil
}

// archivedSha is the archived script's sha: a migrated bundle holds the one file it stored, whose
// sha256 the archive named it by.
func archivedSha(runDir, bundle string) (string, error) {
	files, err := compute.Bundle(runDir, bundle)
	if err != nil {
		return "", err
	}
	if len(files) != 1 {
		return "", fmt.Errorf("bundle %s holds %d files; a migrated computation's bundle holds its one script", bundle, len(files))
	}
	return compute.Sums(files)[0].Sha256, nil
}

// grade is one act's outcome against its bundle's execution.
func grade(runDir, bundle string, act *recordpb.Computation, e compute.Execution, rerr error) string {
	if rerr != nil || len(e.Calls) > 0 {
		return Refused
	}
	c, err := compute.Compare(runDir, bundle, act.GetComputationId(), int(act.GetExit()), e)
	switch {
	case err != nil:
		return Refused
	case c.Matches:
		return Reproduced
	case act.GetBasis() == recordpb.Basis_BASIS_VARIANT:
		return Moved
	default:
		return Differs
	}
}

// Interpreters is the version each command an archived program names reports, as the first line
// it prints; "absent" where the command is not on PATH or does not answer.
func Interpreters() map[string]string {
	out := map[string]string{}
	for name, argv := range versionCommands {
		v, err := exec.Command(argv[0], argv[1:]...).Output()
		line := strings.TrimSpace(strings.SplitN(string(v), "\n", 2)[0])
		if err != nil || line == "" {
			line = "absent"
		}
		out[name] = line
	}
	return out
}

// versionCommands asks each command an archived argv names for its version: every interpreter in
// migrate's frozen table of what the archive ran.
var versionCommands = map[string][]string{
	"python3": {"python3", "--version"},
	"node":    {"node", "--version"},
	"bash":    {"bash", "--version"},
	"go":      {"go", "version"},
}
