package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// EVERYTHING HERE IS BEST-EFFORT AND SAYS SO ON STDERR.
//
// A hook that cannot write its store must never cost a session its turn — the contract
// `appendRow` has held since this binary existed. Every failure below is reported and swallowed,
// and the exit code is unaffected.
//
// AND `Stop`/`SessionEnd` WRITE ZERO MANIFEST ROWS. main.go's final line routes every
// non-`SessionStart` event to `appendRow(buildRow(...))`, a seat row. A `Stop` binding that merely
// added ingest without leaving that path would write one `event-names-no-seat` row per turn into
// the manifest — inflating the very population #189 measured. These are explicit branches for
// that reason, not a fall-through.

func openCatalogue(stderr io.Writer) *sql.DB {
	dir, err := catalogue.DefaultDir()
	if err != nil {
		return nil
	}
	db, err := catalogue.Open(filepath.Join(dir, "catalogue.db"))
	if err != nil {
		fmt.Fprintf(stderr, "gray-area-capture: catalogue unavailable: %v (capture unaffected)\n", err)
		return nil
	}
	return db
}

func projectsRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// ingestForSession brings one session's transcripts up to date. Called at Stop and SessionEnd.
func ingestForSession(sessionID string, stderr io.Writer) {
	if sessionID == "" {
		return
	}
	root := projectsRoot()
	if root == "" {
		return
	}
	db := openCatalogue(stderr)
	if db == nil {
		return
	}
	defer db.Close()

	files, err := catalogue.TranscriptFiles(root)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area-capture: catalogue: enumerating transcripts: %v (capture unaffected)\n", err)
		return
	}
	for _, f := range files {
		if f.SessionID != sessionID {
			continue
		}
		if _, err := catalogue.IngestFile(db, f); err != nil {
			fmt.Fprintf(stderr, "gray-area-capture: catalogue: %v (capture unaffected)\n", err)
			return
		}
	}
}

// sweep runs at SessionStart: closure on EVERY invocation, retention at most once a UTC day.
//
// The two are separate because they cost differently — a closure tail read measured 2.4 ms, a
// retention DELETE 325 ms — and conflating them made an earlier draft of the design claim both
// "the next SessionStart resolves it" and "a no-op on every SessionStart but the first of a day".
func sweep(stderr io.Writer) {
	root := projectsRoot()
	if root == "" {
		return
	}
	db := openCatalogue(stderr)
	if db == nil {
		return
	}
	defer db.Close()

	files, err := catalogue.TranscriptFiles(root)
	if err != nil {
		return
	}
	home, _ := os.UserHomeDir()
	live := map[string]bool{}
	if sfs, err := catalogue.ReadSessionFiles(filepath.Join(home, ".claude", "sessions")); err == nil {
		for _, sf := range sfs {
			live[sf.SessionID] = true
		}
	}
	lim := catalogue.DefaultLimits()
	if st, err := catalogue.Closure(db, files, live, lim); err != nil {
		fmt.Fprintf(stderr, "gray-area-capture: catalogue closure: %v (capture unaffected)\n", err)
	} else if st.Deferred > 0 {
		// SAID OUT LOUD: a backlog that drains silently is indistinguishable from an empty queue.
		fmt.Fprintf(stderr, "gray-area-capture: catalogue closed %d session(s), %d deferred to the next SessionStart\n",
			st.SessionsShut, st.Deferred)
	}
	now := time.Now()
	if catalogue.DueForRetention(db, now) {
		if _, err := catalogue.Retain(db, lim, now); err != nil {
			fmt.Fprintf(stderr, "gray-area-capture: catalogue retention: %v (capture unaffected)\n", err)
			return
		}
		if err := catalogue.MarkRetained(db, now); err != nil {
			fmt.Fprintf(stderr, "gray-area-capture: catalogue retention marker: %v\n", err)
		}
	}
}
