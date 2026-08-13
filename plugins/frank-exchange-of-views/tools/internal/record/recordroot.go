package record

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// WHERE A RUN'S EVENTS LIVE — AND WHY THAT IS NOT ALWAYS UNDER THE RUN.
//
// By default the record is <runDir>/records and everything in this file is inert. The
// separation exists for one measured reason: A SEAT WITH FILESYSTEM ACCESS TO ITS OWN RECORD
// DOES NOT NEED THE TOOL. On the first seat-probe dispatch the seat never called `show` — it
// ran `ls` on the run directory and parsed records/*.jsonl itself, while its constitution says
// the tool's board is the only source of truth for status. Every teaching surface built into
// that path (view names and their descriptions, enum menus, refusals that name the verb you
// meant) reaches a seat that ASKS THE TOOL, and nobody else.
//
// The cost is not the bypass itself. It is that a seat which cannot find a verb reads the
// record, proceeds, and writes NOTHING — so a capability the surface failed to teach is
// indistinguishable from one no seat ever wanted. That is this repo's recurring defect shape:
// the miss and the honest zero are the same bytes. Move the record and a stuck seat has exactly
// two exits, find the verb or file friction, and both of those are signal.
//
// THIS IS NOT A SANDBOX AND MUST NOT BE DESCRIBED AS ONE. `--add-dir` governs Read/Grep/Glob;
// Bash reaches any path the operating system allows, and this package is open source, so a seat
// that goes hunting will find the root. What separation buys is that the record is no longer on
// the CHEAP path: nothing in a listing of the run directory hands it over. That is sized to the
// measured failure, which was satisficing rather than evasion.

const (
	// RecordRootEnv DECLARES a separation, and is read only when a run has no pointer yet.
	// It is deliberately not how the root is resolved afterwards: the harness runs boards
	// concurrently in one process, and a process-global variable cannot answer a per-run
	// question — two boards would resolve to one root and interleave their events.
	RecordRootEnv = "FEOV_RECORD_ROOT"

	// separatedMarker sits in the run directory of a separated run. It never holds the root
	// (that would hand the path straight back to the seat this separates the record FROM).
	// Its whole job is to turn a lost pointer into an error instead of an empty board.
	separatedMarker = ".records-elsewhere"

	// runBackPointer sits INSIDE a record root and names the run it belongs to, so that two
	// runs adopting one root is refused at the second rather than discovered later as one
	// board wearing another's events.
	runBackPointer = ".run"
)

const markerText = `This run's event record was written OUTSIDE this directory.

The path is deliberately not recorded here: seats read this directory, and the separation
exists so that a seat reaches the record through the tool rather than through the filesystem.

Read the board the way a seat does:  feov-record <role> show board --run <this directory>
If that reports the record is unreachable, the resolver's pointer has been lost (a cleaned
cache, or a copied run directory). Re-declare the root:  ` + RecordRootEnv + `=<path> feov-record ...
`

// THIS FILE IS AN AGENT-FACING SURFACE, and it went stale like every other one.
//
// It said `show --run <dir> --view board` — the spelling `show` had before it became a group —
// and a seat that read the marker and followed it got `unknown flag: --view`. Measured on the
// 2026-08-13 probe: two of them, in a run where every prompt had already been corrected, because
// this text is written by the RESOLVER and no prompt gate looks at Go string literals that end up
// in a file inside the run directory.
//
// It is the same rename, the same class, and the fourth carrier of it — after the prompts, the
// help text, and `lens reproduce`'s error message. A surface a seat reads is agent-facing
// wherever it lives, and this one lives in a const.

// RecordsDir resolves where a run's events live, and REFUSES rather than guessing.
//
// Order matters and each step earns its place:
//
//	pointer   per-run, so concurrent runs in one process resolve independently
//	env       declares a NEW separation; adopted once, then never consulted again
//	marker    a separated run whose pointer is gone — an error, never an empty board
//	default   <runDir>/records, the ordinary case
func RecordsDir(runDir string) (string, error) {
	abs, err := filepath.Abs(runDir)
	if err != nil {
		return "", err
	}
	ptr, hasPtr, err := readRootPointer(abs)
	if err != nil {
		return "", err
	}
	env := strings.TrimSpace(os.Getenv(RecordRootEnv))

	// A POINTER WITHOUT ITS MARKER IS A POINTER TO A RUN THAT NO LONGER EXISTS.
	//
	// The binding has two halves: the pointer in the user cache, and the marker in the run
	// directory. Deleting a run directory removes one and leaves the other, and a NEW run at the
	// same path then inherits its predecessor's root — which is not a hypothetical: the seat
	// probe rebuilds nine boards at fixed paths on every invocation, and the second run refused
	// all nine with a conflict naming two temp directories the operator never chose.
	//
	// Re-adopting is right rather than merely convenient. The pointer is a cache; the marker is
	// on the artifact. When they disagree, the artifact wins.
	if hasPtr {
		if _, err := os.Stat(filepath.Join(abs, separatedMarker)); os.IsNotExist(err) {
			hasPtr = false
			_ = forgetRootPointer(abs)
		}
	}

	switch {
	case hasPtr && env != "" && !samePath(env, ptr):
		// Silently preferring either one writes half a run into each root.
		return "", fmt.Errorf("record: %s is set to %s but run %s already keeps its record at %s — "+
			"a run has ONE record root; unset %s to use the established one, or point it at that same path",
			RecordRootEnv, env, abs, ptr, RecordRootEnv)
	case hasPtr:
		// A DANGLING POINTER IS THE SAME LIE ONE LEVEL UP. adoptRoot creates the root, so it
		// existing is an invariant — and if it has since been deleted, returning the path would
		// send MergedEvents into its IsNotExist arm, which flattens to an empty merge and reports
		// a run with a full history as a board of zeros. Measured shape, not a hypothetical: the
		// seat probe builds each board on a temp root it removes afterwards, leaving exactly this.
		if _, err := os.Stat(ptr); err != nil {
			return "", fmt.Errorf("record: run %s keeps its record at %s, which is no longer there (%v) — "+
				"the events are gone or the directory moved. This is NOT an empty run and will not be reported as one; "+
				"re-declare the root with %s=<path> if it moved", abs, ptr, err, RecordRootEnv)
		}
		return ptr, nil
	case env != "":
		return adoptRoot(abs, env)
	}

	if _, err := os.Stat(filepath.Join(abs, separatedMarker)); err == nil {
		return "", fmt.Errorf("record: run %s was written with its record OUTSIDE the run directory "+
			"(%s is present) and the resolver has no pointer for it — the pointer lives in the user cache and "+
			"does not survive a cleaned cache or a copied run directory. Re-declare the root: %s=<path>. "+
			"REPORTING AN EMPTY BOARD HERE WOULD BE A LIE — the record exists and is simply not reachable",
			abs, separatedMarker, RecordRootEnv)
	}
	return filepath.Join(abs, "records"), nil
}

// adoptRoot binds a run to a record root for good.
//
// It runs on a READ as readily as a write, because the first thing anything does with a
// freshly declared root is read an empty board off it. Everything here is idempotent.
func adoptRoot(absRun, root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if samePath(root, filepath.Join(absRun, "records")) {
		// Not an error worth failing on — it is the default spelled out longhand — but the
		// marker and pointer must NOT be written, or an ordinary run acquires the machinery
		// of a separated one and breaks the moment the cache is cleaned.
		return root, nil
	}
	// A RUN THAT ALREADY HAS EVENTS IN PLACE CANNOT BE SEPARATED MID-LIFE. Adopting anyway would
	// leave the earlier shards orphaned in <runDir>/records — still on disk, never read again —
	// and every projection would render a board missing its own history with no anomaly to show
	// for it. The events are not lost, so the fix is to move them deliberately, not to guess here.
	if ents, err := os.ReadDir(filepath.Join(absRun, "records")); err == nil && len(ents) > 0 {
		return "", fmt.Errorf("record: run %s already keeps events in its own records/ and cannot be "+
			"separated now — those shards would be orphaned where nothing reads them. Separate a run BEFORE "+
			"its first register, or move records/ to %s yourself and delete the original",
			absRun, root)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}

	back := filepath.Join(root, runBackPointer)
	switch b, err := os.ReadFile(back); {
	case err == nil && !samePath(strings.TrimSpace(string(b)), absRun):
		return "", fmt.Errorf("record: %s already holds the record of run %s — two runs cannot share one "+
			"record root, their events would merge into one board wearing both runs' history; give run %s its own",
			root, strings.TrimSpace(string(b)), absRun)
	case err != nil && !os.IsNotExist(err):
		return "", err
	case os.IsNotExist(err):
		if err := os.WriteFile(back, []byte(absRun+"\n"), 0o644); err != nil {
			return "", err
		}
	}

	if err := writeRootPointer(absRun, root); err != nil {
		return "", err
	}
	// REQUIRED, NOT BEST EFFORT. It was best-effort, on the reasoning that a read against a
	// read-only run directory should still resolve — but the marker is now load-bearing in both
	// directions: its ABSENCE is what tells a later resolve that this run directory was deleted
	// and rebuilt, so a marker that silently failed to write would make every subsequent resolve
	// re-adopt and hand the run a fresh empty root. Adoption happens at the first write to the
	// record, so a run directory nobody can write to is not a case worth degrading for.
	if err := os.WriteFile(filepath.Join(absRun, separatedMarker), []byte(markerText), 0o644); err != nil {
		return "", fmt.Errorf("record: cannot mark %s as separated (%w) — without the marker a later "+
			"read cannot tell this run from one whose directory was deleted, and would silently adopt a "+
			"fresh empty root in place of the record", absRun, err)
	}
	return root, nil
}

// rootPointerPath keys the pointer by a hash of the run's absolute path.
//
// The hash is not obfuscation — it is that a run directory is an arbitrarily deep path and a
// cache filename is not. Mirrors internal/cli/merge's run-mirror, which keys the same way.
// cacheDirFn is indirected so tests resolve against a temp directory. Without the seam a test
// run would leave pointers in the developer's real cache, and one stale pointer there outlives
// the test that wrote it — resolving a LATER run to a deleted temp root.
var cacheDirFn = os.UserCacheDir

func rootPointerPath(absRun string) (string, error) {
	cache, err := cacheDirFn()
	if err != nil {
		return "", err
	}
	key := absRun
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	sum := sha256.Sum256([]byte(filepath.Clean(key)))
	return filepath.Join(cache, "feov", "record-root", hex.EncodeToString(sum[:])[:12]), nil
}

func readRootPointer(absRun string) (string, bool, error) {
	p, err := rootPointerPath(absRun)
	if err != nil {
		return "", false, err
	}
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	root := strings.TrimSpace(string(b))
	if root == "" {
		return "", false, fmt.Errorf("record: the record-root pointer at %s is empty — delete it and re-declare %s", p, RecordRootEnv)
	}
	return root, true, nil
}

// forgetRootPointer drops a run's binding. Used when the run directory it named is gone.
func forgetRootPointer(absRun string) error {
	p, err := rootPointerPath(absRun)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeRootPointer(absRun, root string) error {
	p, err := rootPointerPath(absRun)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(root+"\n"), 0o644)
}

// samePath compares two paths as the filesystem would, which on Windows means case-insensitively.
// Symlinks are NOT resolved: EvalSymlinks fails on a path that does not exist yet, and every
// caller here is comparing against a root that may be about to be created.
func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
