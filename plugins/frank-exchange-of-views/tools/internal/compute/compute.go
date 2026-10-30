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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// Timeout bounds one execution. A computation settles a question; anything needing longer is a
// project, and belongs in prose with its cost stated.
const Timeout = 60 * time.Second

// Result is one recorded execution: the bundle the tool stored, and what its two runs wrote.
type Result struct {
	// Bundle is the bundle's identity (BundleSha) over Argv and the stored file.
	Bundle string
	Argv   []string
	Output string
	Exit   int
	// Drift is what moved between the runs — the author's two, or the store's copy against them —
	// or "" when every run wrote the same bytes.
	Drift string
	// Calls is every feov-record call the stored copy made, each refused: a bundle that calls the
	// tool cannot run in a checker's hands.
	Calls []string
	// Failed is the interpreter/shell error line found in the output, or "" when there is none.
	//
	// It is not an exit code and cannot be one: both programs this check was built for exited 0.
	// A shell script whose `find` cannot see its target prints the error, carries on, and ends
	// clean; the caller decides what to do with that, because a computation MAY legitimately
	// capture an error (see ErrorSignature).
	Failed string
}

// interpreterFor and its PATH resolution live in interpreter.go, which generalises what stood
// here: per-extension candidate lists, the Windows `py -3` launcher, and a miss that names which
// interpreters this environment DOES have rather than only saying no.

// ScriptPath is the file a --script names: absolute as given, or relative to the run directory —
// the one resolution ScriptBundle, Run and the caller's checks all read.
func ScriptPath(runDir, scriptPath string) string {
	if filepath.IsAbs(scriptPath) {
		return scriptPath
	}
	return filepath.Join(runDir, scriptPath)
}

// ScriptArgv is the command a one-script bundle is named by: the extension's first interpreter and
// the stored file, `script<ext>`. It is the argv the bundle's identity is over, so it names the
// interpreter, never the path one machine resolves it to.
func ScriptArgv(scriptPath string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(scriptPath))
	names := candidates[ext]
	if len(names) == 0 {
		return nil, fmt.Errorf("%s has no known interpreter — a program must be .py, .js, .mjs, .sh or .go so the tool runs it the way you meant rather than guessing", filepath.Base(scriptPath))
	}
	return append(append([]string{names[0]}, argvFor[names[0]]...), "script"+ext), nil
}

// ScriptBundle is the bundle a --script names, and its identity — computed WITHOUT running it.
//
// It exists so `compute` can tell a crash-retry from a key collision before it decides to skip the
// work. Those two look identical at the key and are opposite facts: one is the same program
// arriving twice, the other is a DIFFERENT program under a key an earlier dispatch already burned.
func ScriptBundle(runDir, scriptPath string) (bundle string, argv []string, files []File, err error) {
	body, err := os.ReadFile(ScriptPath(runDir, scriptPath))
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot read %s: %w", scriptPath, err)
	}
	if argv, err = ScriptArgv(scriptPath); err != nil {
		return "", nil, nil, err
	}
	files = []File{{Path: argv[len(argv)-1], Body: body}}
	return BundleSha(argv, Sums(files)), argv, files, nil
}

// Run executes the script TWICE, stores it as a one-file bundle under
// <runDir>/computations/<bundle_sha>/dir/, and runs the stored copy as a checker would.
//
// Twice is the whole mechanism. One run cannot tell a computation from a coin flip, and the
// difference is exactly what a reader needs. Run refuses nothing about what the runs wrote: it
// reports the drift, and the caller decides whether the seat accepted it.
//
// THE STORED COPY RUNS TOO, because the store is where every later reader runs it. Running twice
// from the author's own path proves determinism THERE and says nothing about whether the artifact
// this run ships can be checked by anybody else. In research/2026-09-02_quadratic-formula 28
// programs ran twice from the author's path and matched, while six did not reproduce from the
// store and one exited 0 from it printing INVERTED answers: the author's directory is three levels
// under the run, the store deeper, and `__file__`-relative path arithmetic silently changes
// meaning between them.
func Run(runDir, scriptPath string) (*Result, error) {
	bundle, argv, files, err := ScriptBundle(runDir, scriptPath)
	if err != nil {
		return nil, err
	}
	body := files[0].Body
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("%s is empty — an empty program exits cleanly and establishes nothing, which is the most misleading artifact available", scriptPath)
	}
	abs := ScriptPath(runDir, scriptPath)
	interp, err := interpreterFor(abs)
	if err != nil {
		return nil, err
	}

	first, exit1, err := execute(runDir, interp, abs)
	if err != nil {
		return nil, err
	}
	second, exit2, err := execute(runDir, interp, abs)
	if err != nil {
		return nil, err
	}
	res := &Result{Bundle: bundle, Argv: argv, Output: first, Exit: exit1}
	if first != second || exit1 != exit2 {
		res.Drift = describeDrift(first, second, exit1, exit2)
	}
	res.Failed = ErrorSignature(first)
	if err := WriteBundle(runDir, bundle, argv, files); err != nil {
		return nil, err
	}
	stored, sexit, calls, serr := executeStored(runDir, bundle, argv)
	switch {
	case serr != nil:
		return nil, fmt.Errorf("%s does not run from the store: %w", scriptPath, serr)
	case len(calls) > 0:
		res.Calls = calls
	case res.Drift == "" && (stored != res.Output || sexit != res.Exit):
		res.Drift = "the stored copy says something else — " + describeDrift(res.Output, stored, res.Exit, sexit)
	}
	return res, nil
}

// executeStored runs a bundle's stored copy as a checker runs it: `computations/<sha>/dir/` is the
// only copy that outlives the run directory's working files and the only one a later audit can
// reach.
func executeStored(runDir, bundle string, argv []string) (string, int, []string, error) {
	stored := filepath.Join(BundlePath(runDir, bundle), argv[len(argv)-1])
	interp, err := interpreterFor(stored)
	if err != nil {
		return "", 0, nil, err
	}
	return executeChecked(runDir, interp, stored)
}

// executeChecked is execute as a CHECKER runs it: the script's environment names a fresh marker
// file under seatenv.VarComputeRerun, feov-record refuses every call made under it and appends the
// call there, and the calls come back beside the output. A script that called the tool cannot run
// in a checker's hands, whatever it then did with the refusal — a `jq` over the refusal turns it
// into an ordinary-looking number.
func executeChecked(runDir string, argv []string, script string) (string, int, []string, error) {
	marker, err := os.CreateTemp("", "feov-computation-rerun-*")
	if err != nil {
		return "", 0, nil, fmt.Errorf("could not stage the re-run's tool-call marker: %w", err)
	}
	path := marker.Name()
	marker.Close()
	defer os.Remove(path)
	out, code, err := executeIn(runDir, argv, script, append(os.Environ(), seatenv.VarComputeRerun+"="+path))
	if err != nil {
		return "", 0, nil, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", 0, nil, fmt.Errorf("could not read the re-run's tool-call marker: %w", err)
	}
	var calls []string
	for _, line := range strings.Split(string(body), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			calls = append(calls, line)
		}
	}
	return out, code, calls, nil
}

// execute runs one pass with the run directory as the working directory.
//
// A NON-ZERO EXIT IS NOT AN ERROR HERE. "The check failed" is a result a computation is entitled to
// report: disproving a claim is worth exactly as much as proving one, and turning it into a
// tool error would leave the engine able to record only confirmations.
func execute(runDir string, argv []string, script string) (string, int, error) {
	return executeIn(runDir, argv, script, nil)
}

// executeIn is execute under env, or under this process's environment when env is nil.
func executeIn(runDir string, argv []string, script string, env []string) (string, int, error) {
	return executeWithin(Timeout, runDir, argv, script, env)
}

// executeWithin is executeIn under a given limit; only a test asks for one other than Timeout.
func executeWithin(limit time.Duration, runDir string, argv []string, script string, env []string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	args := append(append([]string{}, argv[1:]...), script)
	cmd := exec.CommandContext(ctx, argv[0], args...)
	cmd.Dir = runDir
	cmd.Env = env
	// CombinedOutput, split so the launch is timed apart from the run: a timeout has to say whether
	// the interpreter was slow to START or the script was slow to FINISH, which are different
	// problems (#1554).
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	// THE LIMIT HOLDS AGAINST A SCRIPT'S CHILDREN TOO. The context kills the interpreter, not what it
	// started; a child still holding the output pipe kept Wait blocked until IT exited — measured: a
	// `go run` computation killed at 500ms returned after the full minute its built binary slept. WaitDelay
	// closes the pipe that long after the kill, so the limit is the limit.
	cmd.WaitDelay = time.Second
	t0 := time.Now()
	err := cmd.Start()
	launch := time.Since(t0)
	if err == nil {
		err = cmd.Wait()
	}
	waited := time.Since(t0) - launch
	out := buf.Bytes()
	if ctx.Err() != nil {
		return "", 0, fmt.Errorf("%s did not finish inside %s (launching %s took %s, then %s waiting on it) — a computation settles a question; anything longer is a project and belongs in prose with its cost stated",
			filepath.Base(script), limit, argv[0], launch.Round(time.Millisecond), waited.Round(time.Millisecond))
	}
	if errors.Is(err, exec.ErrWaitDelay) {
		return "", 0, fmt.Errorf("%s exited, but a process it started still held its output %s later — a computation's output is complete when the script exits, so wait for what it starts",
			filepath.Base(script), cmd.WaitDelay)
	}
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			return "", 0, fmt.Errorf("could not run %s: %w (is %s installed?)", filepath.Base(script), err, argv[0])
		}
		code = ee.ExitCode()
	}
	return string(out), code, nil
}

// describeDrift names WHAT moved, so a reader can judge whether the movement is the point (a
// live measurement) or a defect (an unseeded random, a timestamp printed into the output).
func describeDrift(a, b string, ea, eb int) string {
	switch {
	case ea != eb:
		return fmt.Sprintf("exit code moved %d -> %d between runs", ea, eb)
	case len(a) != len(b):
		return fmt.Sprintf("output length moved %d -> %d bytes between runs", len(a), len(b))
	default:
		for i := range a {
			if a[i] != b[i] {
				return fmt.Sprintf("output differs from byte %d between runs", i)
			}
		}
		return "output differs between runs"
	}
}

// ErrorSignature reports the first line of output that is the ENVIRONMENT failing rather than the
// computation answering, or "" when the output holds none.
//
// WHY THIS IS NOT THE EXIT CODE. Both computations that produced it exited 0. `Run` executes with the
// RUN directory as its working directory; two lane scripts were authored against the repo root
// (their own comments say "run from repo root"), so `find ./plugins/…` printed
// "No such file or directory", the script carried on past it, and the run ended clean. One
// recorded a file enumeration that enumerated nothing — every "ABSENT" verdict below it vacuous —
// and the other recorded "0 files, 0 bytes" immediately above its own hard-coded line asserting
// the opposite. Both were cited in a shipped report as re-runnable evidence.
//
// THE SIGNATURES ARE THE SHELL'S AND THE INTERPRETER'S, not the script's vocabulary. A computation about
// error handling may legitimately print the word "error"; none of the lines below are ones a
// script produces on purpose, and the one case that does — a computation whose POINT is to capture a
// failing command — says so with `--expect-error` rather than being guessed at from the text.
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
		short(sha), what, what, at)
}

// short is the sha's leading twelve characters, the form every message about a computation prints.
func short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// Rerun is one re-run of a recorded computation's stored bundle, against one act's recorded result.
type Rerun struct {
	// Matches is whether it wrote the act's recorded output and exit byte for byte. Never true where
	// the bundle called the tool.
	Matches bool
	// Got and Exit are what it wrote; Want is what the act's run recorded.
	Got, Want string
	Exit      int
	// ToolCalls is every feov-record call the bundle made, each one refused (seatenv.VarComputeRerun).
	// A bundle that made any cannot run in a checker's hands, and nothing it printed is a result.
	ToolCalls []string
}

// Reproduce re-runs a recorded computation's bundle and reports whether it still writes what the
// act recorded: its stored output and its exit, which the record holds.
//
// THIS IS RED'S AUDIT, and it is strictly stronger than any citation check: a cited source is
// re-read and believed; a computation is re-executed and compared. It is also the direct answer to
// the smoke's R2-2, where blue could only assert that a test had happened.
//
// IT RUNS AS A CHECKER (executeChecked), so a bundle that calls the tool is reported as one, never
// as a match.
func Reproduce(runDir, bundle string, argv []string, id string, exit int) (Rerun, error) {
	recorded, err := ActOutput(runDir, bundle, id)
	if err != nil {
		// THE REFUSAL NAMES THE REFERENCE, NOT THE FILESYSTEM. A seat that names a computation the
		// store does not hold is handed what the reference was and what it means that nothing
		// carries it, never a path under the record root it cannot look at.
		return Rerun{}, fmt.Errorf("no recorded output for computation %s in this run's store — `show evidence` lists every computation by its anchor", id)
	}
	if len(argv) == 0 {
		return Rerun{}, fmt.Errorf("computation %s names no command, so it cannot be re-run — which is the one thing a computation is for", id)
	}
	if _, err := os.Stat(filepath.Join(BundlePath(runDir, bundle), argv[len(argv)-1])); err != nil {
		return Rerun{}, fmt.Errorf("%s has no stored program, so it cannot be re-run — which is the one thing a computation is for", short(bundle))
	}
	out, code, calls, err := executeStored(runDir, bundle, argv)
	if err != nil {
		return Rerun{}, err
	}
	return Rerun{Matches: len(calls) == 0 && out == string(recorded) && code == exit, Got: out, Want: string(recorded), Exit: code, ToolCalls: calls}, nil
}
