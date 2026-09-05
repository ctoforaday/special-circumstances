package dashboard

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// Fingerprint summarises everything BuildModel could read, so a watcher can tell an idle
// fifteen seconds from a busy one.
//
// # Why the watcher needs this
//
// `dashboard --watch` called BuildModel unconditionally every 15 seconds. Each call replays the
// record AND re-reads and re-parses every agent-*.jsonl in the transcript directory — files that
// are append-only and GROW through a run, so the cost climbs while the work stays the same.
// Measured at ~2s of CPU per tick on a long run, ~1,200 ticks over five hours: ~40 minutes of CPU
// spent re-deriving a page that, in most windows, nobody had given it new facts for (#684 F15).
//
// # It walks the ROOTS, and that is the whole design
//
// The obvious implementation lists the files BuildModel opens — run-config.json, journal.jsonl,
// the agent transcripts, the record. That list is a second copy of BuildModel's reads with
// nothing holding the two together, and its failure is the silent one: BuildModel gains an input,
// the fingerprint does not learn about it, and the dashboard stops updating when that input
// changes. A frozen page looks exactly like a quiet run.
//
// So this walks the two DIRECTORIES BuildModel is confined to and hashes what it finds. A new
// input under either root is covered the day it is added, by nobody. The cost of being
// over-broad is a wasted render when an unread file changes, which is what every tick did
// before this existed.
//
// EVERY UNCERTAINTY RENDERS. A walk error, an unreadable directory, anything unexpected returns
// "" — and Changed treats "" as changed. The failure direction is deliberate: rendering when we
// did not need to costs CPU, and skipping when we should not have costs the operator a dashboard
// that quietly stopped telling the truth.
func Fingerprint(runDir, transcriptDir string) string {
	h := sha256.New()
	var lines []string
	for _, root := range []string{runDir, transcriptDir} {
		if root == "" {
			continue
		}
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// A directory that vanished mid-walk is not a reason to render forever, but it
				// IS a reason to render this tick: skip the subtree and let the digest differ.
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if d.IsDir() || skipFromFingerprint(path) {
				return nil
			}
			info, ierr := d.Info()
			if ierr != nil {
				return nil
			}
			lines = append(lines, fmt.Sprintf("%s\x00%d\x00%d", path, info.Size(), info.ModTime().UnixNano()))
			return nil
		})
		if err != nil {
			return "" // could not measure: say so, and let the caller render
		}
	}
	if len(lines) == 0 {
		return "" // nothing found is not the same as nothing changed
	}
	// SORTED, because WalkDir's order is lexical per directory but the two roots may nest or
	// overlap, and a digest that depended on traversal order would report a change that was only
	// a reshuffle.
	sort.Strings(lines)
	for _, l := range lines {
		_, _ = h.Write([]byte(l))
		_, _ = h.Write([]byte("\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// skipFromFingerprint names the ONE thing whose churn is ours rather than the run's.
//
// dashboard.html is what the watcher WRITES. Including it makes every render change the
// fingerprint that decides whether to render, so the watcher chases its own tail and never skips
// a tick — change detection that is inert while looking correct.
//
// # What is deliberately NOT excluded, and why the list is this short
//
// record.db-wal is IN. The record runs journal_mode=WAL: a committed write lands in the WAL and
// may not touch record.db at all until a checkpoint. Measured on a seeded fixture — record.db
// 4 KB and unmoved, record.db-wal 428 KB carrying the events. A fingerprint over record.db alone
// freezes the dashboard exactly when the run is busiest, and
// TestFingerprintSeesAWriteAndIgnoresOurOwnRead fails if this file is dropped.
//
// record.db-shm is also IN, and an earlier draft excluded it on the theory that a WAL reader
// touches the shared-memory index and would dirty the fingerprint its own BuildModel call had
// just taken. MEASURED, AND THE THEORY WAS WRONG: across a seed and a full BuildModel read, -shm
// moved in neither size nor mtime. The exclusion was a precaution carrying a confident comment
// and no evidence, so it is gone. If a reader ever does start touching it, the same test that
// disproved the theory is the one that catches it — the guard is the test, not a defensive
// skip, and the cost of being wrong here is one wasted render rather than a stale page.
func skipFromFingerprint(path string) bool {
	return filepath.Base(path) == outputFileName
}

// outputFileName is the page the watcher writes. It is named here rather than passed in because
// the exclusion is a property of what this package produces, not of who calls it.
const outputFileName = "dashboard.html"

// Changed reports whether a re-render is warranted, and is the ONLY place the empty
// "could not measure" digest is interpreted — so no caller has to remember which way it fails.
func Changed(previous, current string) bool {
	return previous == "" || current == "" || previous != current
}
