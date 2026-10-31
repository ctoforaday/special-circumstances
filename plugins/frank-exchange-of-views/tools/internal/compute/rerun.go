package compute

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// ONE STANDING RE-RUN PER BUNDLE, JUDGED PER ACT (plans/compute-and-show.md §III.1.4).
//
// A bundle is addressed by its sha and never changes, so a second execution of the same bytes
// measures the environment, not the computation. Rerun executes a bundle ONCE, in one fresh copy, and
// that one execution is what every act on the bundle is compared against — each act against its OWN
// recorded output, exit and out/ (Compare), because two acts on one bundle can have recorded
// different results: a `variant` program records whatever its first run wrote. The execution is
// stored content-addressed at rerun/<rerun_sha>/ (StageRerun), and the record names which one stands.

// Execution is what one execution of a stored bundle wrote: its output with the copy's root read as
// BundleToken, its exit, the files it wrote under out/, and every feov-record call it made, each
// refused. A bundle that made any cannot run in a checker's hands, and nothing it printed is a result.
type Execution struct {
	Output string
	Exit   int
	Out    []File
	Calls  []string
	// Leftover is why the execution's copy could not be removed after it finished, or "" when it was
	// (Result.Leftover): the execution stands, and the caller reports the leftover as a defect.
	Leftover string
}

// Rerun executes a recorded computation's stored bundle once, in one fresh copy outside the run tree,
// as `compute` ran it. It writes nothing to the store: the caller compares, and stores only what it
// records.
//
// THIS IS RED'S AUDIT, and it is strictly stronger than any citation check: a cited source is re-read
// and believed; a computation is re-executed and compared.
func Rerun(runDir, bundle string, argv []string) (e Execution, err error) {
	if len(argv) == 0 {
		return Execution{}, fmt.Errorf("bundle %s names no command, so it cannot be re-run — which is the one thing a computation is for", Short(bundle))
	}
	files, err := Bundle(runDir, bundle)
	if err != nil {
		return Execution{}, fmt.Errorf("bundle %s cannot be re-run, which is the one thing a computation is for: %w", Short(bundle), err)
	}
	c, err := resolve(argv)
	if err != nil {
		return Execution{}, err
	}
	base, err := scratch(runDir)
	if err != nil {
		return Execution{}, err
	}
	defer func() {
		if cerr := clear(base); cerr != nil {
			if err != nil {
				err = errors.Join(err, cerr)
			} else {
				e.Leftover = cerr.Error()
			}
		}
	}()
	x, err := execute(Timeout, filepath.Join(base, copyRoots[0]), &Program{Argv: argv, Files: files, Sha: bundle}, c)
	if err != nil {
		return Execution{}, err
	}
	return Execution{Output: x.output, Exit: x.exit, Out: x.out, Calls: x.calls}, nil
}

// Sha is an execution's identity, the name of its directory under rerun/: sha256 over the sorted
// listing of its items, each length-prefixed as BundleSha frames them — `exit <n>`, `output
// <sha256>`, and `out <path> <sha256>` per out/ file — so a newline in a file name forges nothing.
func (e Execution) Sha() string {
	sum := sha256.Sum256([]byte(e.Output))
	items := []string{frame("exit", strconv.Itoa(e.Exit)), frame("output", hex.EncodeToString(sum[:]))}
	for _, f := range Sums(e.Out) {
		items = append(items, frame("out", f.Path, f.Sha256))
	}
	sort.Strings(items)
	h := sha256.New()
	for _, it := range items {
		fmt.Fprintf(h, "%d:%s", len(it), it)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Comparison is one act's recorded result set beside an execution of its bundle.
type Comparison struct {
	// Matches is whether the execution wrote this act's recorded output, exit and out/, byte for
	// byte. Never true where the execution called the tool.
	Matches bool
	// Recorded is this act's own recorded output.
	Recorded string
	// Drift names what differs — out/ files, exit, and what the execution wrote — or "" where it
	// matches.
	Drift string
}

// driftOutputLimit bounds the execution's output a drift carries: the act's output and the
// execution's are both in the store, and the drift says what moved, not a second copy of it.
const driftOutputLimit = 400

// Compare sets one act against an execution of its bundle: the act's own acts/<id>/ output and out/,
// and its recorded exit. An act that wrote no out/ equals an execution whose out/ is empty.
func Compare(runDir, bundle, id string, exit int, e Execution) (Comparison, error) {
	recorded, err := ActOutput(runDir, bundle, id)
	if err != nil {
		// THE REFUSAL NAMES THE REFERENCE, NOT THE FILESYSTEM. A seat that names a computation the
		// store does not hold is handed what the reference was, never a path under the record root
		// it cannot look at.
		return Comparison{}, fmt.Errorf("no recorded output for computation %s in this run's store — `show evidence` lists every computation by its anchor", id)
	}
	recordedOut, err := ActOut(runDir, bundle, id)
	if err != nil {
		return Comparison{}, err
	}
	c := Comparison{Recorded: string(recorded)}
	var moved []string
	if d := outDrift(recordedOut, e.Out, "this act's run", "the re-run"); d != "" {
		moved = append(moved, d)
	}
	if e.Exit != exit {
		moved = append(moved, fmt.Sprintf("exit moved %d -> %d", exit, e.Exit))
	}
	if e.Output != string(recorded) {
		moved = append(moved, "on re-run the program wrote: "+Excerpt(e.Output, driftOutputLimit))
	}
	for i, m := range moved {
		if i > 0 {
			c.Drift += "; "
		}
		c.Drift += m
	}
	c.Matches = len(e.Calls) == 0 && c.Drift == ""
	return c, nil
}

// Outcome is how one act stands against an execution of its bundle, graded by the act's own basis.
type Outcome string

const (
	// Reproduced: the execution wrote the act's recorded output, exit and out/.
	Reproduced Outcome = "reproduced"
	// Moved: it did not, and the act was recorded variant — its two runs already differed, so the
	// execution moving is the recorded fact arriving, not a defect.
	Moved Outcome = "moved"
	// Differs: it did not, and the act was recorded reproducible.
	Differs Outcome = "differs"
)

// Grade is an act's outcome from its comparison and whether it was recorded variant. A comparison
// whose execution called the tool never matches, and its caller refuses it before grading.
func Grade(c Comparison, variant bool) Outcome {
	switch {
	case c.Matches:
		return Reproduced
	case variant:
		return Moved
	default:
		return Differs
	}
}

// RerunPath is the directory a run holds one execution of a bundle in, by the execution's sha.
func RerunPath(runDir, bundle, sha string) string {
	return filepath.Join(runDir, StoreDir, bundle, "rerun", sha)
}

// Staged is an execution written under its bundle's .tmp-* staging, not yet at rerun/<Sha>/.
type Staged struct {
	Sha         string
	dir, runDir string
	bundle      string
	promoted    bool
}

// StageRerun writes an execution's output and out/ into a fresh staging directory beside its
// bundle's. Nothing reads staging: the caller compares and judges, then Promotes what it records or
// Discards what it refuses, so a refused act leaves nothing under rerun/.
func StageRerun(runDir, bundle string, e Execution) (*Staged, error) {
	stage, err := staging(runDir, bundle)
	if err != nil {
		return nil, err
	}
	s := &Staged{Sha: e.Sha(), dir: stage, runDir: runDir, bundle: bundle}
	if err := writeResult(stage, []byte(e.Output), e.Out); err != nil {
		os.RemoveAll(stage)
		return nil, err
	}
	return s, nil
}

// Promote renames the staging to rerun/<Sha>/. A directory already there holds the same bytes by
// construction — its name is the sha of its content — so the staging is removed and the present one
// stands: an earlier act's execution whose event was never appended, or a concurrent act's.
func (s *Staged) Promote() error {
	final := RerunPath(s.runDir, s.bundle, s.Sha)
	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		os.RemoveAll(s.dir)
		return err
	}
	if err := os.Rename(s.dir, final); err != nil {
		os.RemoveAll(s.dir)
		if _, serr := os.Stat(final); serr == nil {
			s.promoted = true
			return nil
		}
		return fmt.Errorf("storing the re-run of bundle %s: %w", Short(s.bundle), err)
	}
	s.promoted = true
	return nil
}

// Discard removes the staging of an execution the caller does not record.
func (s *Staged) Discard() error {
	if s.promoted {
		return nil
	}
	return os.RemoveAll(s.dir)
}

// ReadRerun reads the stored execution the record names by its sha: its output and out/, with exit
// the exit its event records. A directory the record names and the store does not hold is an error
// naming it, never an empty execution.
func ReadRerun(runDir, bundle, sha string, exit int) (Execution, error) {
	dir := RerunPath(runDir, bundle, sha)
	out, err := os.ReadFile(filepath.Join(dir, "output"))
	if err != nil {
		return Execution{}, missing(bundle, "re-run "+Short(sha), dir)
	}
	files, err := readOut(filepath.Join(dir, "out"))
	if err != nil {
		return Execution{}, err
	}
	return Execution{Output: string(out), Exit: exit, Out: files}, nil
}

// writeResult writes an output and its out/ files under dir: the one shape acts/<id>/ and
// rerun/<sha>/ share.
func writeResult(dir string, output []byte, out []File) error {
	if err := os.WriteFile(filepath.Join(dir, "output"), output, 0o644); err != nil {
		return err
	}
	for _, f := range out {
		p := filepath.Join(dir, "out", filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, f.Body, 0o644); err != nil {
			return err
		}
	}
	return nil
}
