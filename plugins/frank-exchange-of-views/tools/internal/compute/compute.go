// Package compute executes a seat's program and records it as evidence.
//
// WHY THIS EXISTS. Every seat has had Write, Edit and Bash since the engine shipped, and
// across six runs of recorded events NO SEAT EVER WROTE A PROGRAM TO SETTLE A QUESTION. The
// capability was never tied — it was never invited, and everything around it (search, cite,
// fetch at the leaf, never improvise) reads as "evidence means sources".
//
// The cost is measured. In the 2026-08-04 smoke red raised R1-2 — "protocol validation uses
// ground truth to validate itself; test it on a false claim, e.g. is 9 prime" — blue answered
// with PROSE asserting the test had happened, and red's R2-2 correctly refused it: "no
// evidence shown". A full epoch, for something three lines of trial division settle outright
// while leaving an artifact the auditor can re-run.
//
// RISK POSTURE, and it is the one fetchcache already states: a seat could already run this by
// hand through Bash, so this verb adds RECORDING, not new reach. What it buys is that the
// execution becomes an artifact — hashed, cached, re-runnable by the other side — instead of
// an assertion in a seat's context that died with the turn.
//
// A COMPUTATION IS NOT A NEW EVIDENCE CLASS. It is the citation model with the last mile walked:
// the METHOD is cited (trial division, Miller-Rabin — someone else's proof, sourced) and the
// execution applies that method to this instance. Language is language; a proof in English and a
// proof in Python are the same act of argument. What differs is how the auditor checks
// it — prose is re-read, code is RE-RUN, and re-running is the stronger audit.
package compute

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// Timeout bounds one execution. A computation settles a question; anything needing longer is a
// project, and belongs in prose with its cost stated.
const Timeout = 60 * time.Second

// BundleToken is what every copy's root reads as in a program's output: the two copies sit at
// different paths, and a program that prints its own location prints the same line from either.
const BundleToken = "<bundle>"

// The copies' roots under one execution's scratch directory: the same leaf name at two depths, so a
// program that names its own directory prints the same word from both, and one that reaches out of
// its copy by a relative path (`../x`) reaches a different place from each.
var copyRoots = [2]string{filepath.Join("a", "bundle"), filepath.Join("b", "c", "bundle")}

// Result is one recorded execution: what the first of the two runs wrote, and what moved.
type Result struct {
	Bundle string
	Argv   []string
	Output string
	Exit   int
	// Out is every file the first run wrote under out/, by its path inside out/.
	Out []File
	// ToolchainPath and ToolchainSha256 name argv[0] as it resolved on this machine: "" for a bundle
	// file, whose bytes are already in the bundle's identity.
	ToolchainPath, ToolchainSha256 string
	// Drift is what moved between the runs — exit, output or out/ — or "" when both wrote the same.
	Drift string
	// Calls is every feov-record call either run made, each refused: a bundle that calls the tool
	// cannot run in a checker's hands.
	Calls []string
	// Leftover is why the program's copies could not be removed after both runs finished, or "" when
	// they were. It says nothing about the computation: every field above was measured before the
	// removal, so the caller records the act and reports the leftover as the tool's own defect.
	Leftover string
	// Failed is the interpreter/shell error line found in the output, or "" when there is none.
	//
	// It is not an exit code and cannot be one: both programs this check was built for exited 0.
	// A shell script whose `find` cannot see its target prints the error, carries on, and ends
	// clean; the caller decides what to do with that, because a computation MAY legitimately
	// capture an error (see ErrorSignature).
	Failed string
}

// execution is what one copy's run wrote.
type execution struct {
	output string
	exit   int
	out    []File
	calls  []string
}

// command is how a program's argv[0] starts on this machine.
type command struct {
	// bundled is argv[0]'s path inside the bundle, or "" when argv[0] is not a bundle file.
	bundled string
	// path is the file that starts when argv[0] is not a bundle file.
	path            string
	toolchainSha256 string
}

// resolve finds what argv[0] starts. A relative path with a separator is a bundle file, run from the
// copy; an absolute path is that file; a bare name is looked up on PATH. A miss is named, because a
// seat told only "no" writes the same command again.
func resolve(argv []string) (command, error) {
	arg0 := argv[0]
	if rel, ok := bundled(arg0); ok {
		return command{bundled: rel}, nil
	}
	path := arg0
	if !filepath.IsAbs(arg0) {
		p, err := exec.LookPath(arg0)
		if err != nil {
			return command{}, fmt.Errorf("%s is not on PATH in this environment, so the computation cannot run here — name a program this environment has, or one in the bundle as ./<file>", arg0)
		}
		if path, err = filepath.Abs(p); err != nil {
			return command{}, err
		}
	}
	sum, err := fileSha(path)
	if err != nil {
		return command{}, fmt.Errorf("cannot read %s, the program the command names: %w", path, err)
	}
	return command{path: path, toolchainSha256: sum}, nil
}

func fileSha(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Run executes the program TWICE, each in its own fresh copy of the bundle outside the run tree, and
// reports what the first wrote and what moved between the two. It stores nothing: the caller decides
// whether the seat's act is accepted, and writes the store only then.
//
// Twice is the whole mechanism. One run cannot tell a computation from a coin flip, and the
// difference is exactly what a reader needs.
//
// TWO COPIES AT DIFFERENT DEPTHS, NEITHER THE AUTHOR'S, because a later reader runs a copy too. In
// research/2026-09-02_quadratic-formula 28 programs ran twice from the author's path and matched,
// while six did not reproduce from the store and one exited 0 from it printing INVERTED answers:
// `__file__`-relative path arithmetic silently changes meaning between directories. Two copies at
// different depths make that arithmetic move between the runs, where it is seen.
func Run(runDir string, p *Program) (res *Result, err error) {
	c, err := resolve(p.Argv)
	if err != nil {
		return nil, err
	}
	base, err := scratch(runDir)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := clear(base); cerr != nil {
			if err != nil || res == nil {
				err = errors.Join(err, cerr)
			} else {
				res.Leftover = cerr.Error()
			}
		}
	}()
	var runs [2]execution
	for i, sub := range copyRoots {
		if runs[i], err = execute(Timeout, filepath.Join(base, sub), p, c); err != nil {
			return nil, err
		}
	}
	first, second := runs[0], runs[1]
	res = &Result{Bundle: p.Sha, Argv: p.Argv, Output: first.output, Exit: first.exit, Out: first.out,
		ToolchainPath: c.path, ToolchainSha256: c.toolchainSha256}
	if c.bundled != "" {
		res.ToolchainPath, res.ToolchainSha256 = "", ""
	}
	res.Drift = describeDrift(first, second)
	res.Calls = union(first.calls, second.calls)
	res.Failed = ErrorSignature(first.output)
	return res, nil
}

// scratch is a fresh directory for one act's copies, outside the run tree: a copy inside it would
// let a program reach the run by a relative path that no other machine holds.
func scratch(runDir string) (string, error) {
	base, err := os.MkdirTemp("", "feov-compute-*")
	if err != nil {
		return "", fmt.Errorf("could not make a directory for the program's copies: %w", err)
	}
	canonBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		os.RemoveAll(base)
		return "", err
	}
	if canonRun, err := filepath.EvalSymlinks(runDir); err == nil {
		if rel, err := filepath.Rel(canonRun, canonBase); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			os.RemoveAll(base)
			return "", fmt.Errorf("the system's temporary directory %s is inside the run directory %s, and a program's copies must sit outside the run tree — point TMPDIR elsewhere", canonBase, canonRun)
		}
	}
	return base, nil
}

// removeCopies removes an execution's scratch directory once its read-only bits are cleared; a test
// replaces it to make the removal fail.
var removeCopies = os.RemoveAll

// SetRemoveCopiesForTest has every execution remove its copies with f, and returns the restore.
func SetRemoveCopiesForTest(f func(string) error) (restore func()) {
	prev := removeCopies
	removeCopies = f
	return func() { removeCopies = prev }
}

// clear removes an execution's scratch directory, clearing every read-only bit first: a program may
// leave a directory it cannot be deleted from, and Windows refuses to remove a read-only file.
func clear(base string) error {
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		if d.IsDir() {
			os.Chmod(p, 0o700)
		} else if d.Type().IsRegular() {
			os.Chmod(p, 0o600)
		}
		return nil
	})
	if err := removeCopies(base); err != nil {
		return fmt.Errorf("could not remove the program's copies at %s: %w", base, err)
	}
	return nil
}

// execute runs the program once in a fresh copy of its bundle at root.
//
// THE COPY IS WRITTEN WITH EXPLICIT MODES, never the umask's: every file 0644, argv[0]'s bundle file
// 0755, every directory 0755. The mode is not part of the bundle's identity, so it is the tool's to
// set the same way in every copy on every machine.
//
// THE ROOT IS READ BACK CANONICAL (EvalSymlinks; on Windows the long form of an 8.3 path), because
// that is the spelling the program sees: getcwd and an interpreter's own file path resolve every
// symlink on the way. Replacing any other spelling with BundleToken would leave the copy's real path
// in the output, different in each copy.
func execute(limit time.Duration, root string, p *Program, c command) (execution, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return execution{}, err
	}
	canon, err := filepath.EvalSymlinks(root)
	if err != nil {
		return execution{}, err
	}
	for _, f := range p.Files {
		path := filepath.Join(canon, filepath.FromSlash(f.Path))
		if err := mkdirs(canon, filepath.Dir(path)); err != nil {
			return execution{}, err
		}
		if err := os.WriteFile(path, f.Body, 0o644); err != nil {
			return execution{}, err
		}
		mode := os.FileMode(0o644)
		if f.Path == c.bundled {
			mode = 0o755
		}
		if err := os.Chmod(path, mode); err != nil {
			return execution{}, err
		}
	}
	if err := mkdirs(canon, filepath.Join(canon, "out")); err != nil {
		return execution{}, err
	}
	// The checker's marker and the plotting config sit beside the copy, never in it.
	parent := filepath.Dir(canon)
	marker := filepath.Join(parent, "calls")
	rc := filepath.Join(parent, "matplotlibrc")
	if err := os.WriteFile(marker, nil, 0o600); err != nil {
		return execution{}, err
	}
	// svg.hashsalt fixes the ids matplotlib otherwise salts at random into every SVG it writes.
	if err := os.WriteFile(rc, []byte("svg.hashsalt: feov\n"), 0o644); err != nil {
		return execution{}, err
	}
	path := c.path
	if c.bundled != "" {
		path = filepath.Join(canon, filepath.FromSlash(c.bundled))
	}
	output, exit, err := launch(limit, canon, path, p.Argv, checkerEnv(canon, marker, rc))
	if err != nil {
		return execution{}, err
	}
	out, err := readOut(filepath.Join(canon, "out"))
	if err != nil {
		return execution{}, err
	}
	body, err := os.ReadFile(marker)
	if err != nil {
		return execution{}, fmt.Errorf("could not read the run's tool-call marker: %w", err)
	}
	// Each call is the arguments the program passed the tool, which may be any bytes.
	var calls []string
	for _, line := range strings.Split(string(body), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			calls = append(calls, Printable(line))
		}
	}
	return execution{output: strings.ReplaceAll(output, canon, BundleToken), exit: exit, out: out, calls: calls}, nil
}

// mkdirs makes dir and every directory between root and it, each 0755 whatever the umask.
func mkdirs(root, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for d := dir; d != root && strings.HasPrefix(d, root); d = filepath.Dir(d) {
		if err := os.Chmod(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// checkerEnv is the environment a program runs under: the caller's, with the checker's marker
// (seatenv.VarComputeRerun — feov-record refuses every call made under it and appends the call
// there) and the pins that take the commonest sources of moving output out of a run. PWD names the
// copy, since the inherited one names the caller's directory, which no other machine has.
func checkerEnv(root, marker, rc string) []string {
	return append(os.Environ(),
		seatenv.VarComputeRerun+"="+marker,
		"SOURCE_DATE_EPOCH=0",
		"PYTHONHASHSEED=0",
		"TZ=UTC",
		"LC_ALL=C",
		"PYTHONUTF8=1",
		"MATPLOTLIBRC="+rc,
		"PWD="+root,
	)
}

// launch runs path with argv (argv[0] as the seat typed it) in dir, stdout and stderr as one stream.
//
// A NON-ZERO EXIT IS NOT AN ERROR HERE. "The check failed" is a result a computation is entitled to
// report: disproving a claim is worth exactly as much as proving one.
func launch(limit time.Duration, dir, path string, argv []string, env []string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, path)
	cmd.Args = append([]string{}, argv...)
	cmd.Dir = dir
	cmd.Env = env
	// One stream, split from CombinedOutput so the launch is timed apart from the run: a timeout has
	// to say whether the program was slow to START or slow to FINISH, which are different problems
	// (#1554).
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	// THE LIMIT HOLDS AGAINST A PROGRAM'S CHILDREN TOO. The context kills the process, not what it
	// started; a child still holding the output pipe kept Wait blocked until IT exited — measured: a
	// `go run` computation killed at 500ms returned after the full minute its built binary slept.
	// WaitDelay closes the pipe that long after the kill, so the limit is the limit.
	cmd.WaitDelay = time.Second
	t0 := time.Now()
	if err := cmd.Start(); err != nil {
		return "", 0, fmt.Errorf("the system cannot start %s: %v — a bundle file the command runs directly must be a program this system executes (on Windows, an .exe; elsewhere, a binary or a script with a #! line); run a script through its interpreter instead, as `-- bash run.sh`", argv[0], err)
	}
	launched := time.Since(t0)
	err := cmd.Wait()
	waited := time.Since(t0) - launched
	if ctx.Err() != nil {
		return "", 0, fmt.Errorf("%s did not finish inside %s (launching it took %s, then %s waiting on it) — a computation settles a question; anything longer is a project and belongs in prose with its cost stated",
			argv[0], limit, launched.Round(time.Millisecond), waited.Round(time.Millisecond))
	}
	if errors.Is(err, exec.ErrWaitDelay) {
		return "", 0, fmt.Errorf("%s exited, but a process it started still held its output %s later — a computation's output is complete when the program exits, so wait for what it starts",
			argv[0], cmd.WaitDelay)
	}
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			return "", 0, fmt.Errorf("could not run %s: %w", argv[0], err)
		}
		code = ee.ExitCode()
	}
	return buf.String(), code, nil
}

// readOut reads every file a run wrote under out/, by its slash-separated path inside out/. A
// symlink or other non-file there is refused: the act keeps out/'s bytes, and those have none.
func readOut(dir string) ([]File, error) {
	var out []File
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("the program left out/%s, which is not a file — out/ holds the files the act keeps, and this has no bytes to keep", Printable(filepath.ToSlash(rel)))
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out = append(out, File{Path: filepath.ToSlash(rel), Body: body})
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, err
}

func union(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range append(append([]string{}, a...), b...) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// describeDrift names WHAT moved, so a reader can judge whether the movement is the point (a live
// measurement) or a defect (an unseeded random, a timestamp printed into the output); "" when
// nothing did.
func describeDrift(a, b execution) string {
	var moved []string
	if a.exit != b.exit {
		moved = append(moved, fmt.Sprintf("exit code moved %d -> %d between runs", a.exit, b.exit))
	}
	if a.output != b.output {
		moved = append(moved, outputDrift(a.output, b.output))
	}
	if d := outDrift(a.out, b.out, "the first", "the second"); d != "" {
		moved = append(moved, d)
	}
	return strings.Join(moved, "; ")
}

func outputDrift(a, b string) string {
	if len(a) != len(b) {
		return fmt.Sprintf("output length moved %d -> %d bytes between runs", len(a), len(b))
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return fmt.Sprintf("output differs from byte %d between runs", i)
		}
	}
	return "output differs between runs"
}

// outDrift names every out/ file that differs between two sets, or "" when they hold the same
// bytes, naming the side that alone wrote a file as aName or bName. An absent out/ and an empty one
// are the same set.
func outDrift(a, b []File, aName, bName string) string {
	am, bm := map[string][]byte{}, map[string][]byte{}
	var names []string
	for _, f := range a {
		am[f.Path] = f.Body
		names = append(names, f.Path)
	}
	for _, f := range b {
		if _, ok := am[f.Path]; !ok {
			names = append(names, f.Path)
		}
		bm[f.Path] = f.Body
	}
	sort.Strings(names)
	var moved []string
	for _, n := range names {
		x, inA := am[n]
		y, inB := bm[n]
		name := "out/" + Printable(n)
		switch {
		case !inB:
			moved = append(moved, name+" written by "+aName+" only")
		case !inA:
			moved = append(moved, name+" written by "+bName+" only")
		case !bytes.Equal(x, y):
			moved = append(moved, name+" differs")
		}
	}
	return strings.Join(moved, ", ")
}

// ErrorSignature reports the first line of output that is the ENVIRONMENT failing rather than the
// computation answering, or "" when the output holds none.
//
// WHY THIS IS NOT THE EXIT CODE. Both computations that produced it exited 0. Two lane scripts were
// authored against the repo root (their own comments say "run from repo root"), so
// `find ./plugins/…` printed "No such file or directory", the script carried on past it, and the run
// ended clean. One recorded a file enumeration that enumerated nothing — every "ABSENT" verdict below
// it vacuous — and the other recorded "0 files, 0 bytes" immediately above its own hard-coded line
// asserting the opposite. Both were cited in a shipped report as re-runnable evidence.
//
// THE SIGNATURES ARE THE SHELL'S AND THE INTERPRETER'S, not the program's vocabulary. A computation
// about error handling may legitimately print the word "error"; none of the lines below are ones a
// program produces on purpose, and the one case that does — a computation whose POINT is to capture
// a failing command — says so with `--expect-error` rather than being guessed at from the text.
var errorSignatures = []string{
	"No such file or directory",
	"command not found",
	"Permission denied",
	"Traceback (most recent call last)",
	"ModuleNotFoundError",
	"cannot access",
	": not found",
}

func ErrorSignature(output string) string {
	for _, line := range strings.Split(output, "\n") {
		for _, sig := range errorSignatures {
			if strings.Contains(line, sig) {
				return strings.TrimSpace(line)
			}
		}
	}
	return ""
}

// Stored reads a recorded act's artifact: the bundle as `reproduce` will execute it, and the
// output this act's run recorded.
//
// A MISSING HALF IS AN ERROR, NEVER AN EMPTY STRING. The caller is a seat about to judge whether
// the program establishes a claim, and "" reads as a program that computes nothing — a fault in
// the store, recorded as a verdict on the computation. The error names the directory because this
// is the store's failure and not the seat's mistake: the sha came off the record, so the one who
// can act on it is whoever holds the run directory.
func Stored(runDir, bundle, id string) (files []File, output []byte, err error) {
	if files, err = Bundle(runDir, bundle); err != nil {
		return nil, nil, err
	}
	if output, err = ActOutput(runDir, bundle, id); err != nil {
		return nil, nil, err
	}
	return files, output, nil
}

// missing is the one refusal for an artifact the record names and the store does not hold.
func missing(sha, what, at string) error {
	return fmt.Errorf("bundle %s is on the record and its %s is not in the store: %s not found at %s. "+
		"That is a fault in the run, not in the computation — it cannot be read, and a judgement of it would be a judgement of nothing. Report it with `log` as a defect",
		Short(sha), what, what, at)
}
