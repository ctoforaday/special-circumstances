// Package proof executes a seat's script and records it as evidence.
//
// WHY THIS EXISTS. Every seat has had Write, Edit and Bash since the engine shipped, and
// across six runs of recorded events NO SEAT EVER WROTE A PROGRAM TO SETTLE A QUESTION. The
// capability was never tied — it was never invited, and everything around it (search, cite,
// fetch at the leaf, never improvise) reads as "evidence means sources".
//
// The cost is measured. In the 2026-08-04 smoke red raised R1-2 — "protocol validation uses
// ground truth to validate itself; test it on a false claim, e.g. is 9 prime" — blue answered
// with PROSE asserting the test had happened, and red's R2-2 correctly refused it: "no
// evidence shown". A full round, for something three lines of trial division settle outright
// while leaving an artifact the auditor can re-run.
//
// RISK POSTURE, and it is the one fetchcache already states: a seat could already run this by
// hand through Bash, so this verb adds RECORDING, not new reach. What it buys is that the
// execution becomes an artifact — hashed, cached, re-runnable by the other side — instead of
// an assertion in a seat's context that died with the turn.
//
// A PROOF IS NOT A NEW EVIDENCE CLASS. It is the citation model with the last mile walked:
// the METHOD is cited (trial division, Miller-Rabin — someone else's proof, sourced) and the
// execution applies that method to this instance. Language is language; a proof in English
// and a proof in Python are the same act of argument. What differs is how the auditor checks
// it — prose is re-read, code is RE-RUN, and re-running is the stronger audit.
package proof

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Timeout bounds one execution. A proof settles a question; anything needing longer is a
// project, and belongs in prose with its cost stated.
const Timeout = 60 * time.Second

// The two bases, DERIVED from running twice — never asserted by the seat that benefits from
// the claim, which is the rule fix_basis already follows.
//
//	Reproducible  ran twice, identical output. The strongest evidence the engine can hold,
//	              because the auditor reaches the same bytes by running it again.
//	Observed      ran twice, output MOVED. Still evidence — of a system that changes. A live
//	              network call is the common case and it is legitimate: what an API actually
//	              does beats what its documentation says. But it is a MEASUREMENT, not a
//	              proof, and a report must not read the two as the same thing.
const (
	Reproducible = "reproducible"
	Observed     = "observed"
)

// Result is one recorded execution.
type Result struct {
	SHA    string `json:"sha256"` // of the script — the proof's identity
	Basis  string `json:"basis"`
	Output string `json:"output"`
	Exit   int    `json:"exit"`
	Script string `json:"script_path"`
	Drift  string `json:"drift,omitempty"` // what moved on the second run, when something did
}

// interpreterFor maps a script to how it runs.
//
// BY EXTENSION, deliberately: a shebang does not work on Windows and this suite runs there.
// An unknown extension is REFUSED rather than guessed, because running a file under the wrong
// interpreter produces a confident error that reads exactly like a failed proof.
func interpreterFor(path string) ([]string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".py":
		return []string{"python"}, nil
	case ".js", ".mjs":
		return []string{"node"}, nil
	case ".sh":
		return []string{"bash"}, nil
	case ".go":
		return []string{"go", "run"}, nil
	default:
		return nil, fmt.Errorf("proof: %s has no known interpreter — a proof script must be .py, .js, .mjs, .sh or .go so the tool runs it the way you meant rather than guessing", filepath.Base(path))
	}
}

// Run executes the script TWICE and records it under <runDir>/proofs/<sha>/.
//
// Twice is the whole mechanism. One run cannot tell a computation from a coin flip, and the
// difference is exactly what a reader needs. Neither outcome is refused — the run is GRADED,
// because refusing the moving kind would bar the evidence that beats documentation, and
// accepting it silently would launder a measurement as a proof.
func Run(runDir, scriptPath string) (*Result, error) {
	abs := scriptPath
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(runDir, scriptPath)
	}
	body, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("proof: cannot read %s: %w", scriptPath, err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("proof: %s is empty — an empty script exits cleanly and proves nothing, which is the most misleading artifact available", scriptPath)
	}
	argv, err := interpreterFor(abs)
	if err != nil {
		return nil, err
	}

	first, exit1, err := execute(runDir, argv, abs)
	if err != nil {
		return nil, err
	}
	second, exit2, err := execute(runDir, argv, abs)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(body)
	res := &Result{
		SHA:    hex.EncodeToString(sum[:]),
		Output: first,
		Exit:   exit1,
		Script: scriptPath,
		Basis:  Reproducible,
	}
	if first != second || exit1 != exit2 {
		res.Basis = Observed
		res.Drift = describeDrift(first, second, exit1, exit2)
	}
	if err := store(runDir, res, body); err != nil {
		return nil, err
	}
	return res, nil
}

// execute runs one pass with the run directory as the working directory.
//
// A NON-ZERO EXIT IS NOT AN ERROR HERE. "The check failed" is a result a proof is entitled to
// report: disproving a claim is worth exactly as much as proving one, and turning it into a
// tool error would leave the engine able to record only confirmations.
func execute(runDir string, argv []string, script string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	args := append(append([]string{}, argv[1:]...), script)
	cmd := exec.CommandContext(ctx, argv[0], args...)
	cmd.Dir = runDir
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return "", 0, fmt.Errorf("proof: %s did not finish inside %s — a proof settles a question; anything longer is a project and belongs in prose with its cost stated", filepath.Base(script), Timeout)
	}
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			return "", 0, fmt.Errorf("proof: could not run %s: %w (is %s installed?)", filepath.Base(script), err, argv[0])
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

// store writes the artifact: the script as executed and its output. Keyed by the SCRIPT's hash,
// so one proof run twice is one artifact and a changed script is a different one — the identity
// rule the fetch cache already uses for a source.
//
// IT WRITES NO METADATA. `meta.json` used to sit here holding sha, exit, script path and drift —
// facts about the debate, in a JSON file beside the record that exists to hold facts about the
// debate. Nothing could refuse a wrong value in it, the report rendered fields the record could
// not state, and red's re-run read its expectations out of it. Those fields are on `Proof` now.
// What stays is CONTENT: the script body and the output bytes, which are too large to be facts
// and are addressed by the same sha the record carries.
func store(runDir string, r *Result, body []byte) error {
	dir := filepath.Join(runDir, "proofs", r.SHA)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "script"+filepath.Ext(r.Script)), body, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "output"), []byte(r.Output), 0o644)
}

// Reproduce re-runs a recorded proof and reports whether it still says the same thing.
//
// THIS IS RED'S AUDIT, and it is strictly stronger than any citation check: a cited source is
// re-read and believed; a proof is re-executed and compared. It is also the direct answer to
// the smoke's R2-2, where blue could only assert that a test had happened.
func Reproduce(runDir, sha string) (matches bool, got, want string, err error) {
	dir := filepath.Join(runDir, "proofs", sha)
	// The recorded output is the ARTIFACT, read as bytes. What the proof CLAIMED — its exit, its
	// drift, the script it ran — is on the record, where a reader can refuse it; this function
	// needs only the bytes to compare against.
	recorded, err := os.ReadFile(filepath.Join(dir, "output"))
	if err != nil {
		return false, "", "", fmt.Errorf("proof: no recorded proof %s in this run: %w", sha, err)
	}
	rec := Result{Output: string(recorded)}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, "", "", err
	}
	script := ""
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "script") {
			script = filepath.Join(dir, e.Name())
		}
	}
	if script == "" {
		return false, "", "", fmt.Errorf("proof: %s has no stored script, so it cannot be re-run — which is the one thing a proof is for", sha)
	}
	argv, err := interpreterFor(script)
	if err != nil {
		return false, "", "", err
	}
	out, _, err := execute(runDir, argv, script)
	if err != nil {
		return false, "", "", err
	}
	return out == rec.Output, out, rec.Output, nil
}
