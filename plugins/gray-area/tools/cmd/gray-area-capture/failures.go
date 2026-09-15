package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// A FAILURE THIS HOOK SWALLOWS IS WRITTEN DOWN, AND SAID WHERE A HUMAN LOOKS.
//
// Every path here exits 0 — a hook that cannot write its store must never cost a session its
// turn — and a hook's stderr at exit 0 goes to the client's debug log and nowhere else. So the
// stderr lines alone meant a catalogue that could not open, or a manifest that could not be
// written, failed on every turn for as long as it stayed broken, and nobody was told. The same
// shape left prosthetic-conscience 0.44.0's hooks silent no-ops for hours (hooks/fetch-bin.sh is
// the fix there, and this follows its pattern).
//
// So each failure is also a field on a record (failures.json), keyed by the stage that failed and
// cleared only by that stage's own next success — "no entry" then means the stage last WORKED,
// never "nothing looked". The record is SHOWN only on the events whose top-level systemMessage the
// client displays, SessionStart and Stop, and each stage at most once per notifyEvery.
// SubagentStop and SessionEnd record and never print: SessionEnd output is displayed nowhere, and a
// subagent hook that talks can re-invoke the seat it is attached to.
//
// CONCURRENT HOOKS RACE ON THE RECORD, and that is accepted rather than locked: a hook must never
// wait. Each writer loads, merges and renames over the file, so the last writer wins. A lost failure
// is recorded again the next time the stage fails; a lost clear shows a stale entry once more, until
// that stage's next success clears it again.

// stage names what failed. The closed set is the record's key space: a reader asks "is the catalogue
// opening", never parses a message for it.
type stage string

const (
	stageOpen            stage = "catalogue-open"
	stageBackfill        stage = "catalogue-backfill" // rebuilt empty and not backfilled since
	stageRegister        stage = "catalogue-register"
	stageEnumerate       stage = "catalogue-enumerate"
	stageIngest          stage = "catalogue-ingest"
	stageClosure         stage = "catalogue-closure"
	stageRetention       stage = "catalogue-retention"
	stageRetentionMarker stage = "catalogue-retention-marker"
	stageManifest        stage = "manifest"
	stageProjectRoot     stage = "project-root"
	stageSessionPath     stage = "session-transcript-path"
)

// notifyEvery is how long a stage that is still failing stays unannounced after being shown. Stop
// fires every turn; without it a broken store would speak on every one.
const notifyEvery = 10 * time.Minute

// failuresSchema versions failures.json.
const failuresSchema = 1

// failure is one entry. Its identity is (Stage, Project): the manifest is per project, so a manifest
// that works in one project must not clear the entry of another whose manifest does not. Every other
// stage is about the machine's store or the client and has no Project.
type failure struct {
	Stage    stage     `json:"stage"`
	Project  string    `json:"project,omitempty"`
	Error    string    `json:"error"`
	Event    string    `json:"event"` // the hook event that last hit it
	Since    time.Time `json:"since"` // first failure since the stage last worked
	Last     time.Time `json:"last"`
	Notified time.Time `json:"notified,omitzero"` // last time a displaying event showed it
}

type failureRecord struct {
	Schema   int       `json:"schema"`
	Failures []failure `json:"failures"`
}

type failureKey struct {
	stage   stage
	project string
}

func (f failure) key() failureKey { return failureKey{f.Stage, f.Project} }

// report collects one invocation's outcomes by stage and settles them into the record at the end.
type report struct {
	event  string
	now    time.Time
	stderr io.Writer
	failed map[failureKey]string
	worked map[failureKey]bool
}

func newReport(event string, now time.Time, stderr io.Writer) *report {
	return &report{event: event, now: now, stderr: stderr, failed: map[failureKey]string{}, worked: map[failureKey]bool{}}
}

// fail says the failure on stderr, as before, and records it. A stage that fails and works in one
// invocation settles as failed.
func (r *report) fail(s stage, detail string) { r.failIn(s, "", detail) }

// ok records that a stage worked, which clears its entry.
func (r *report) ok(s stage) { r.okIn(s, "") }

// failIn and okIn are fail and ok for a stage whose entries are per project.
func (r *report) failIn(s stage, project, detail string) {
	fmt.Fprintln(r.stderr, "gray-area-capture: "+detail)
	r.failed[failureKey{s, project}] = detail
}

func (r *report) okIn(s stage, project string) { r.worked[failureKey{s, project}] = true }

// displays reports whether the client shows a top-level systemMessage from this hook event.
func displays(event string) bool { return event == "SessionStart" || event == "Stop" }

// failuresPath is the record's location: beside the catalogue's directory, never inside it, so a
// catalogue directory that cannot be created or written does not take its own failure record down
// with it.
func failuresPath() (string, error) {
	dir, err := catalogue.DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(dir), "gray-area-capture", "failures.json"), nil
}

// settle merges this invocation into the record at path and, on a displaying event, prints the
// stages due an announcement as one systemMessage. path is "" when no location resolved.
func (r *report) settle(path string, pathErr error, stdout io.Writer) {
	rec, persistErr := r.merge(path, pathErr)
	var due []failure
	if persistErr != nil {
		// NO RECORD, SO NO THROTTLE: show what this invocation saw, and why it could not be kept.
		fmt.Fprintf(r.stderr, "gray-area-capture: failure record not kept: %v\n", persistErr)
		if !displays(r.event) || len(r.failed) == 0 {
			return
		}
		for k, d := range r.failed {
			due = append(due, failure{Stage: k.stage, Project: k.project, Error: d, Event: r.event, Since: r.now, Last: r.now})
		}
		sortFailures(due)
		writeSystemMessage(stdout, renderFailures(due, "", persistErr))
		return
	}
	if !displays(r.event) {
		return
	}
	for i, f := range rec.Failures {
		if f.Notified.IsZero() || r.now.Sub(f.Notified) >= notifyEvery {
			rec.Failures[i].Notified = r.now
			due = append(due, rec.Failures[i])
		}
	}
	if len(due) == 0 {
		return
	}
	if err := saveFailures(path, rec); err != nil {
		fmt.Fprintf(r.stderr, "gray-area-capture: failure record not kept: %v\n", err)
	}
	writeSystemMessage(stdout, renderFailures(due, path, nil))
}

// merge loads the record, applies this invocation and saves it when anything changed. It returns
// the merged record. A healthy invocation with no record on disk writes nothing.
func (r *report) merge(path string, pathErr error) (failureRecord, error) {
	if pathErr != nil {
		return failureRecord{}, pathErr
	}
	rec, err := loadFailures(path)
	changed := false
	var unreadable errUnreadableRecord
	switch {
	case errors.As(err, &unreadable):
		// Replaced, not kept broken: an unreadable record would otherwise disable the throttle for
		// good. What it held is lost; any stage still failing is recorded again as it fails.
		fmt.Fprintf(r.stderr, "gray-area-capture: replacing the failure record: %v\n", err)
		changed = true
	case err != nil:
		return failureRecord{}, err
	}
	kept := rec.Failures[:0]
	for _, f := range rec.Failures {
		_, failedNow := r.failed[f.key()]
		if r.worked[f.key()] && !failedNow {
			changed = true
			continue
		}
		kept = append(kept, f)
	}
	rec.Failures = kept
	for k, d := range r.failed {
		i := slices.IndexFunc(rec.Failures, func(f failure) bool { return f.key() == k })
		if i < 0 {
			rec.Failures = append(rec.Failures, failure{Stage: k.stage, Project: k.project, Since: r.now})
			i = len(rec.Failures) - 1
		}
		f := &rec.Failures[i]
		f.Error, f.Event, f.Last = d, r.event, r.now
		changed = true
	}
	sortFailures(rec.Failures)
	if changed {
		if err := saveFailures(path, rec); err != nil {
			return rec, err
		}
	}
	return rec, nil
}

// loadFailures reads the record; an absent file is an empty record. A file that does not parse is
// an ERROR alongside the empty record, never a silent empty record: merge says so on stderr and
// replaces it with what this invocation knows.
func loadFailures(path string) (failureRecord, error) {
	rec := failureRecord{Schema: failuresSchema}
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return rec, nil
	}
	if err != nil {
		return rec, err
	}
	var got failureRecord
	if err := json.Unmarshal(b, &got); err != nil || got.Schema != failuresSchema {
		return rec, errUnreadableRecord{path: path, err: err, schema: got.Schema}
	}
	rec.Failures = got.Failures
	return rec, nil
}

type errUnreadableRecord struct {
	path   string
	err    error
	schema int
}

func (e errUnreadableRecord) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s does not parse: %v", e.path, e.err)
	}
	return fmt.Sprintf("%s has schema %d, want %d", e.path, e.schema, failuresSchema)
}

// saveFailures replaces the record by rename, so a reader never sees a half-written file. An empty
// record removes the file.
func saveFailures(path string, rec failureRecord) error {
	if len(rec.Failures) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	rec.Schema = failuresSchema
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "failures-*.json")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}

func sortFailures(fs []failure) {
	slices.SortFunc(fs, func(a, b failure) int {
		return cmp.Or(cmp.Compare(a.Stage, b.Stage), cmp.Compare(a.Project, b.Project))
	})
}

func renderFailures(due []failure, path string, persistErr error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "gray-area-capture: %d stage(s) failing — the session is unaffected; what gray-area records is not:", len(due))
	for _, f := range due {
		where := ""
		if f.Project != "" {
			where = " in " + f.Project
		}
		fmt.Fprintf(&b, "\n- %s%s: %s (failing since %s, last at %s)", f.Stage, where, f.Error, f.Since.UTC().Format(time.RFC3339), f.Event)
	}
	if persistErr != nil {
		fmt.Fprintf(&b, "\nThis is shown on every SessionStart and Stop while it lasts: the failure record cannot be kept (%v).", persistErr)
	} else {
		fmt.Fprintf(&b, "\nRecord: %s. Each entry clears when its stage next works; a failing stage is shown again after %d minutes.", path, int(notifyEvery.Minutes()))
	}
	return b.String()
}

// writeSystemMessage prints the one top-level object the client displays. Nothing else in this
// binary writes stdout on a hook event.
func writeSystemMessage(stdout io.Writer, msg string) {
	b, err := json.Marshal(struct {
		SystemMessage string `json:"systemMessage"`
	}{msg})
	if err != nil {
		return
	}
	fmt.Fprintln(stdout, string(b))
}
