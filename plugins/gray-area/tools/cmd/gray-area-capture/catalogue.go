package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
	db, err := catalogue.Open(filepath.Join(dir, "catalogue.db"), stderr)
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

// ingestForSession brings one session's transcripts up to date. Called at SessionEnd, which only
// ingests: the session file the registration reads is already gone by then.
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
	ingestInto(db, root, sessionID, stderr)
}

// atStop runs at every turn end: ONE open, then the registration, then the ingest. Registering
// here as well as at SessionStart is what keeps a session resumed after closure reopened, and
// catches the cloud id of a session whose SessionStart ran before capture was installed.
func atStop(in hookInput, stderr io.Writer) {
	if in.SessionID == "" {
		return
	}
	db := openCatalogue(stderr)
	if db == nil {
		return
	}
	defer db.Close()
	register(db, in, stderr)
	if root := projectsRoot(); root != "" {
		ingestInto(db, root, in.SessionID, stderr)
	}
}

// register records this session as running (catalogue.RegisterSession). Best-effort, like
// everything here: a failure is one stderr line and the hook's exit code is untouched.
func register(db *sql.DB, in hookInput, stderr io.Writer) {
	if in.SessionID == "" {
		return
	}
	bridge := ""
	if home, err := os.UserHomeDir(); err == nil {
		bridge = bridgeIDFor(filepath.Join(home, ".claude", "sessions"), os.Getppid(), in.SessionID)
	}
	if err := catalogue.RegisterSession(db, in.SessionID, projectDirOf(in.TranscriptPath), bridge, time.Now()); err != nil {
		fmt.Fprintf(stderr, "gray-area-capture: %v (capture unaffected)\n", err)
	}
}

// projectDirOf is the folder the session's transcript lives in, or ” — NEVER ".", which is what
// filepath.Dir makes of an empty path and which would then read as a directory.
func projectDirOf(transcriptPath string) string {
	if transcriptPath == "" {
		return ""
	}
	return filepath.Dir(transcriptPath)
}

// bridgeIDFor reads the Remote Control cloud id from THIS session's own file.
//
// The file is `<sessionsDir>/<ppid>.json`: the hook command `exec`s this binary, so its parent is
// the client process that fired the hook, and the client names each session file by its pid. The
// id is taken ONLY when that file's sessionId is the payload's — a pid is a guess until the file
// confirms it, and a cloud id attributed to the wrong session would send a reattach to someone
// else's conversation. Any miss is ”, which RegisterSession never lets overwrite a known id.
func bridgeIDFor(sessionsDir string, ppid int, sessionID string) string {
	body, err := os.ReadFile(filepath.Join(sessionsDir, strconv.Itoa(ppid)+".json"))
	if err != nil {
		return ""
	}
	var sf catalogue.SessionFile
	if json.Unmarshal(body, &sf) != nil || sf.SessionID != sessionID {
		return ""
	}
	return sf.BridgeSessionID
}

// ingestInto reads every transcript file of one session into an open store.
func ingestInto(db *sql.DB, root, sessionID string, stderr io.Writer) {
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
//
// THIS SESSION REGISTERS FIRST AND COUNTS ITSELF LIVE. Registration reopens a resumed session that
// closure had settled; the closure pass right after it must then not settle it again, and the
// session's own file is not proof of that — so the payload's id joins the live set directly.
func sweep(in hookInput, stderr io.Writer) {
	root := projectsRoot()
	if root == "" {
		return
	}
	db := openCatalogue(stderr)
	if db == nil {
		return
	}
	defer db.Close()
	register(db, in, stderr)

	files, err := catalogue.TranscriptFiles(root)
	if err != nil {
		return
	}
	home, _ := os.UserHomeDir()
	live := map[string]bool{}
	if in.SessionID != "" {
		live[in.SessionID] = true
	}
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
