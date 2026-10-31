// Package setup ports setup-research-run.mjs — the mechanical half of /research
// steps 1-3 — into the feov-record binary. It creates the run's blackboard
// skeleton, pins the evidence base, mirrors the law memory into
// inputs/, writes the .run-live marker, and preflights the record
// binary. Behaviour is preserved byte-for-byte against the mjs it replaces
// (the sole non-deterministic field is the marker's `started` timestamp, and the
// marker lives outside the run dir).
//
// Where the mjs used readdirSync (filesystem order, non-guaranteed), Go sorts the
// entries: a deterministic improvement over the JS's implicit order.
package setup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// dirs are the run's directories. friction.md is deliberately absent (the log lives on the
// record).
//
// `records` is NOT in this list — it is created at its RESOLVED location by BuildSkeleton,
// because a separated run must not be handed an empty records/ in the run directory. An empty
// one is worse than none: it shows up in the seat's listing as the place to look, and it makes
// the run appear unseparated to anyone reading the tree.
//
// SETUP WRITES NO FILE A SEAT FILLS. Untouched-but-present is the worst state a file can be in:
// it reads as an artifact rather than an absence, and a reader takes it for the real thing.
// `blue/candidates/<lane>.md` is each lane's to write and `ingest`'s to read; `report.md` is
// `bench assemble`'s. Setup stubbed the report files once, and in B7 a seat copied its draft onto the
// run-root stub while `ingest` recorded the other one — 40 bytes of heading — as the report. With
// none stubbed, a part written to the wrong path is an ingest that cannot find its file and says
// where it looked.
//
// `scratch` is the seats' working files — drafts, programs, a spool a seat needs on disk — each
// named for its seat. It is made here because a shell redirect into a missing directory fails,
// and a seat's first write is as often `> scratch/…` as a file tool that would make it. Nothing
// composed or archived reads it: the documents come from the record, and capture keeps records/
// and computations/.
var dirs = []string{"blue/candidates", "red", "trajectories", "inputs", "scratch"}

// marshalJSON matches JS `JSON.stringify(x, null, 2) + '\n'`: two-space indent, a
// trailing newline, and — critically — NO HTML escaping. Go's default escapes
// <, >, & to \u00xx; JS does not, and the law banner ("statute > precedent")
// carries '>'. json.Encoder.Encode already appends the newline.
func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// ---- skeleton + pins ----

// BuildSkeleton lays down the run's directories, and no file.
func BuildSkeleton(run record.Run) {
	runDir := run.Dir()
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(runDir, d), 0o755)
	}
	// Resolved once, when the handle was made. This used to re-resolve here and swallow the
	// error on the grounds that a half-built skeleton is worse than an unreachable root — a fair
	// argument, and no longer one anybody has to make: a run that cannot resolve never becomes a
	// record.Run, so run-setup refuses it before laying down a single directory.
	os.MkdirAll(run.Records(), 0o755)
}

// PinResult reports whether PINNED.md was written (false = pre-staged, kept).
type PinResult struct {
	Written bool
	Path    string
}

func BuildPinned(run record.Run, head string, cites []string) PinResult {
	runDir := run.Dir()
	p := filepath.Join(runDir, "inputs", "PINNED.md")
	if exists(p) {
		return PinResult{Written: false, Path: p}
	}
	lines := []string{
		"# PINNED — the evidence base, frozen at launch",
		"",
		"Agents MUST cite these commits for repo and cross-corpus references; the evidence base",
		"does not move under the report. Pushes to cited paths are frozen while the run is live",
		"(.run-live marker; hook-guarded).",
		"",
		"| Corpus | Pin |",
		"|---|---|",
		fmt.Sprintf("| Repo HEAD at launch | `%s` |", head),
	}
	for _, c := range cites {
		path, pin := splitPin(c)
		if pin == "" {
			pin = head
		}
		lines = append(lines, fmt.Sprintf("| `%s` | `%s` |", path, pin))
	}
	lines = append(lines, "")
	os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644)
	return PinResult{Written: true, Path: p}
}

// splitPin splits "path@pin" into (path, pin); "path" alone yields pin "".
func splitPin(c string) (string, string) {
	if i := strings.Index(c, "@"); i >= 0 {
		return c[:i], c[i+1:]
	}
	return c, ""
}

// ---- mirrors ----

type MirrorResult struct {
	Written bool
	Reason  string
	Files   int
	Sources int
}

// MirrorLaw stages the repo's law/ corpus read-only into inputs/law.
func MirrorLaw(repoLawDir string, run record.Run) MirrorResult {
	outDir := record.LawDir(run)
	if exists(outDir) {
		return MirrorResult{Written: false, Reason: "already staged"}
	}
	if repoLawDir == "" || !exists(repoLawDir) {
		return MirrorResult{Written: false, Reason: "no law dir"}
	}
	// THE DIRECTORY IS MADE HERE, AND EVERY WRITE IS CHECKED. This discarded the MkdirAll error,
	// discarded each WriteFile error, and incremented the count regardless, so a caller reading
	// `Written: true, Files: 3` could not tell it from three files on disk.
	//
	// A COUNT THAT OUTRUNS THE DISK IS THE FAILURE, not a missing law corpus: a run with no law is
	// ordinary and says so, while a run that reports a staged corpus it does not have looks
	// complete to every reader after it.
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return MirrorResult{Written: false, Reason: "could not create inputs/law: " + err.Error()}
	}
	files := 0
	for _, f := range mdFiles(repoLawDir, nil) {
		body, err := os.ReadFile(filepath.Join(repoLawDir, f))
		if err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(outDir, f), append([]byte("<!-- mirrored from law/ at run setup — read-only copy -->\n"), body...), 0o644); err != nil {
			return MirrorResult{Written: false, Reason: "could not write " + f + ": " + err.Error()}
		}
		files++
	}
	return MirrorResult{Written: files > 0, Files: files}
}

// ---- pin validation ----

// GitResult mirrors the fields the mjs reads off spawnSync: an exit Status and an Err,
// plus the Stdout `rev-parse` needs. Stdout exists so HEAD resolution goes through the SAME
// injected seam as every other git call — it used to shell out directly, which meant the pin
// gate could not be reached in a test at all: a temp cwd is not a repo, HEAD came back
// "unknown", and ValidatePins skipped itself. A gate whose arming depends on ambient state
// nobody can control is the same defect this file's preflight had.
type GitResult struct {
	Status int
	Err    error
	Stdout string
}

// GitFunc runs a git invocation; injectable for tests, defaults to the real git.
type GitFunc func(args []string) GitResult

func realGit(cwd string) GitFunc {
	return func(args []string) GitResult {
		c := exec.Command("git", args...)
		c.Dir = cwd
		out, err := c.Output()
		status := 0
		if err != nil {
			status = 1
			if ee, ok := err.(*exec.ExitError); ok {
				status = ee.ExitCode()
			}
		}
		return GitResult{Status: status, Err: nil, Stdout: string(out)}
	}
}

// PinCheck names a cite whose path is absent at its pin.
type PinCheck struct {
	Path string
	Pin  string
}

// PinValidation is the pin-validation verdict; Skipped set == non-git context.
type PinValidation struct {
	Checked int
	Missing []PinCheck
	Skipped string
}

func ValidatePins(cites []string, head string, git GitFunc) PinValidation {
	if head == "" || head == "unknown" {
		return PinValidation{Checked: 0, Missing: nil, Skipped: "not a git repo — pins UNVALIDATED"}
	}
	var missing []PinCheck
	for _, c := range cites {
		path, pin := splitPin(c)
		at := pin
		if at == "" {
			at = head
		}
		r := git([]string{"cat-file", "-e", fmt.Sprintf("%s:%s", at, path)})
		if r.Err != nil || r.Status != 0 {
			missing = append(missing, PinCheck{Path: path, Pin: at})
		}
	}
	return PinValidation{Checked: len(cites), Missing: missing}
}

// ---- purge, preflight, version ----

// The mirror purge is record.PurgeStaleMirrors, beside the path it purges: the write, the purge
// and capture's reap read one composition of ~/.cache/feov/run-mirror. Two hand-assembled copies
// of that location would disagree as a purge that returns 0 and reads exactly like a clean board.

// expectedSchema is the epoch the plugin beside this binary declares.
//
// THE AUTHORITY IS THE MANIFEST, AND ONLY THE MANIFEST. It ships with the prompts this run
// will drive, so it is the thing the binary has to agree with. The previous shape fell back
// to the binary's own compiled-in number when the manifest could not be read — which, in the
// documented flow where setup and the seats are the same binary, reduced the check to a
// number compared with itself and passed by construction.
//
// So an unreadable manifest is a REFUSAL, not a default. A missing epoch means the check
// cannot be made, and a check that cannot be made must not report success.
func expectedSchema(recordBin string) (int, error) {
	manifest := filepath.Join(filepath.Dir(filepath.Dir(recordBin)), "requirements.json")
	b, err := os.ReadFile(manifest)
	if err != nil {
		return 0, fmt.Errorf("cannot read %s (%v), so the binary's event schema cannot be checked against the plugin's", manifest, err)
	}
	var m struct {
		EventSchema *int `json:"eventSchema"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return 0, fmt.Errorf("%s is unparseable (%v), so the binary's event schema cannot be checked against the plugin's", manifest, err)
	}
	if m.EventSchema == nil {
		return 0, fmt.Errorf("%s declares no eventSchema, so there is nothing to check the binary against", manifest)
	}
	return *m.EventSchema, nil
}

// ExecResult mirrors the spawnSync fields the mjs preflight reads.
type ExecResult struct {
	Status int
	Err    error
	Stdout string
}

// ExecFunc runs a binary with args; injectable for tests.
type ExecFunc func(bin string, args []string) ExecResult

func realExec(bin string, args []string) ExecResult {
	out, err := exec.Command(bin, args...).Output()
	res := ExecResult{Stdout: string(out)}
	if err == nil {
		return res
	}
	if ee, ok := err.(*exec.ExitError); ok {
		res.Status = ee.ExitCode()
		return res
	}
	// Could not run at all. Node's spawnSync surfaces error.code (ENOENT) for a missing
	// binary; match it so the "not runnable" detail is byte-identical across the port.
	res.Status = 1
	if errors.Is(err, exec.ErrNotFound) {
		res.Err = errors.New("ENOENT")
	} else {
		res.Err = err
	}
	return res
}

// Preflight is the record-binary preflight verdict.
type Preflight struct {
	OK      bool
	Reason  string
	Remedy  string
	Version string
}

func PreflightRecordBinary(expectSchema int, bin string, run ExecFunc) Preflight {
	r := run(bin, []string{"--schema"})
	if r.Err != nil || r.Status != 0 {
		detail := fmt.Sprintf("exit %d", r.Status)
		if r.Err != nil {
			detail = r.Err.Error()
		}
		return Preflight{
			OK:     false,
			Reason: fmt.Sprintf("%s not runnable (%s)", bin, detail),
			Remedy: "install it with /prosthetic-conscience:doctor --fix, or build it: " +
				"(cd plugins/frank-exchange-of-views/tools && go build ./cmd/feov-record)",
		}
	}
	// A DEDICATED FLAG, so nothing parses a number out of prose. --version answers "which
	// build is this" and --schema answers "what shape does it write"; reading the second out
	// of the first is the string-shaped hope this repository keeps removing.
	got, err := strconv.Atoi(strings.TrimSpace(r.Stdout))
	if err != nil {
		return Preflight{
			OK:     false,
			Reason: fmt.Sprintf("%s --schema printed %q, which is not an epoch — this binary predates the check or is not feov-record", bin, strings.TrimSpace(r.Stdout)),
			Remedy: "refresh it with /prosthetic-conscience:doctor --fix (never mid-run)",
		}
	}
	if got != expectSchema {
		// NO DELTA LIST. Nothing here promises backwards compatibility, so "how far behind"
		// is not a question with an answer worth maintaining — the 49-entry changelog that
		// used to answer it was a second copy of the git history, kept by hand.
		return Preflight{
			OK: false,
			Reason: fmt.Sprintf("%s writes event schema %d, the plugin reads %d — the shapes differ, so this binary's events would not be readable by the prompts this run drives",
				bin, got, expectSchema),
			Remedy: "refresh it with /prosthetic-conscience:doctor --fix (never mid-run)",
		}
	}
	return Preflight{OK: true, Version: strconv.Itoa(got)}
}

// ---- small helpers ----

// mdFiles lists *.md files in dir, SORTED (deterministic — the JS used readdir order),
// excluding the names in skip.
func mdFiles(dir string, skip map[string]bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".md") || skip[n] {
			continue
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
