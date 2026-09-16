package main

import (
	"io"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/hookfailures"
)

// The stages this hook can fail at. Each is declared where it can fail — the record's key space is
// this list, so a reader asks "is the catalogue opening" instead of parsing a message for it.
const (
	stageOpen            hookfailures.Stage = "catalogue-open"
	stageBackfill        hookfailures.Stage = "catalogue-backfill" // rebuilt empty and not backfilled since
	stageRegister        hookfailures.Stage = "catalogue-register"
	stageEnumerate       hookfailures.Stage = "catalogue-enumerate"
	stageIngest          hookfailures.Stage = "catalogue-ingest"
	stageClosure         hookfailures.Stage = "catalogue-closure"
	stageRetention       hookfailures.Stage = "catalogue-retention"
	stageRetentionMarker hookfailures.Stage = "catalogue-retention-marker"
	stageManifest        hookfailures.Stage = "manifest"
	stageProjectRoot     hookfailures.Stage = "project-root"
	stageSessionPath     hookfailures.Stage = "session-transcript-path"
)

// report is the recorder plus the stderr the ROUTINE lines keep using — a repaired torn tail and a
// deferred closure backlog are facts for the debug log, not failures anyone must act on.
type report struct {
	*hookfailures.Recorder
	stderr io.Writer
}

func newReport(event string, now time.Time, stderr io.Writer) *report {
	return &report{Recorder: hookfailures.New("gray-area", "gray-area-capture", event, now, stderr), stderr: stderr}
}

func (r *report) fail(s hookfailures.Stage, detail string) { r.Fail(s, detail) }
func (r *report) ok(s hookfailures.Stage)                  { r.OK(s) }
func (r *report) failIn(s hookfailures.Stage, scope, detail string) {
	r.FailIn(s, scope, detail)
}
func (r *report) okIn(s hookfailures.Stage, scope string) { r.OKIn(s, scope) }
