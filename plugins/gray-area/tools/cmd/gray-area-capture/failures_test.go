package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/hookfailures"
)

// WHAT THIS FILE TESTS, NOW THAT THE RECORD IS SHARED (scripts/internal/hookfailures): that THIS
// hook's call sites reach it. The record's own behaviour — throttle, clearing, an unreadable file,
// no record location at all — is tested once, in the module that authors it, and is not re-tested
// per plugin.
//
// THE WIRING IS TESTED BY BREAKING THE REAL THING.
//
// Every stage below is made to fail the way it fails in the field — a foreign file where the store
// should be, a write the database refuses, a directory that cannot be created — and then repaired,
// and the hook is driven through run() with its real paths. A test that injected a failing function
// would prove the record works and not that any call site reaches it.
//
// NOT REACHED BY ANY FIXTURE, and why: catalogue-enumerate's failure (catalogue.TranscriptFiles walks
// with a callback that swallows every error, so no filesystem state makes it return one — its success
// is tested instead), and three manifest failures no filesystem produces on demand: a row that does
// not encode, an fsync that fails after a write that did not, and a torn tail whose closing newline
// cannot be written onto a file that just accepted a read.

// hookEnv is one isolated machine: a home with one session's transcript, its own state directory,
// and a project directory.
type hookEnv struct {
	home, store, project, sid, transcript string
}

func newHookEnv(t *testing.T) *hookEnv {
	t.Helper()
	sid := "sess-failures"
	home, store := fixtureHome(t, sid, "Read", "Edit")
	return &hookEnv{
		home: home, store: store, project: t.TempDir(), sid: sid,
		transcript: filepath.Join(home, ".claude", "projects", "-w", sid+".jsonl"),
	}
}

func (e *hookEnv) input() string { return e.inputAs(e.sid) }

func (e *hookEnv) inputAs(sid string) string {
	return `{"session_id":` + strconv.Quote(sid) + `,"transcript_path":` + strconv.Quote(e.transcript) +
		`,"cwd":` + strconv.Quote(e.project) + `}`
}

// hook runs one event at a given clock and returns what the client would read. It fails the test
// if the exit code is not 0 or stdout is anything but nothing or exactly one systemMessage object:
// stdout on a hook event is parsed by the client, and a stray line there corrupts the object.
func hook(t *testing.T, stdin, projectDir string, now time.Time, event string) (message, stderr string) {
	t.Helper()
	var o, e bytes.Buffer
	if code := run([]string{"-event", event}, strings.NewReader(stdin), &o, &e, projectDir, now, okStat(9)); code != 0 {
		t.Fatalf("%s exited %d — the hook must never block", event, code)
	}
	return systemMessageOf(t, o.String()), e.String()
}

func systemMessageOf(t *testing.T, stdout string) string {
	t.Helper()
	if stdout == "" {
		return ""
	}
	if strings.Count(stdout, "\n") != 1 || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout is not exactly one line: %q", stdout)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
		t.Fatalf("stdout is not a JSON object: %v (%q)", err, stdout)
	}
	msg, ok := obj["systemMessage"].(string)
	if len(obj) != 1 || !ok || msg == "" {
		t.Fatalf("stdout must hold only a non-empty top-level systemMessage, got %q", stdout)
	}
	return msg
}

func recordPath(t *testing.T) string {
	t.Helper()
	p, err := hookfailures.Path("gray-area")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// recorded reads the record as a reader outside this package would: the JSON file. Absent is nil.
func recorded(t *testing.T) []hookfailures.Failure {
	t.Helper()
	b, err := os.ReadFile(recordPath(t))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var rec hookfailures.Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("the failure record does not parse: %v\n%s", err, b)
	}
	if rec.Schema != hookfailures.Schema {
		t.Fatalf("schema = %d, want %d", rec.Schema, hookfailures.Schema)
	}
	return rec.Failures
}

func entryFor(fs []hookfailures.Failure, s hookfailures.Stage, project string) (hookfailures.Failure, bool) {
	for _, f := range fs {
		if f.Stage == s && f.Scope == project {
			return f, true
		}
	}
	return hookfailures.Failure{}, false
}

func execStore(t *testing.T, store string, stmts ...string) {
	t.Helper()
	db, err := sql.Open("sqlite", store)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
}

// createStore makes an empty, current catalogue at the environment's store path.
func createStore(t *testing.T, e *hookEnv) {
	t.Helper()
	db, err := catalogue.Open(e.store, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

// refuse installs a trigger that makes the named write fail, which is how a store that opens and
// then cannot be written behaves — and a trigger is not a table, so the store still classifies as a
// catalogue.
func refuse(t *testing.T, e *hookEnv, name, on string) {
	t.Helper()
	execStore(t, e.store, fmt.Sprintf(`CREATE TRIGGER %s %s BEGIN SELECT RAISE(ABORT, 'forced: %s'); END`, name, on, name))
}

func allow(t *testing.T, e *hookEnv, names ...string) {
	t.Helper()
	for _, n := range names {
		execStore(t, e.store, `DROP TRIGGER `+n)
	}
}

// EVERY STAGE: its failure is recorded under its own name, shown only by an event whose message the
// client displays, and cleared by that stage's own next success.
func TestEveryStageIsRecordedShownAndClearedByItsOwnSuccess(t *testing.T) {
	cases := []struct {
		name    string // "" names the case by its stage
		stage   hookfailures.Stage
		event   string // the event that hits the failure
		project bool   // the entry is per project
		prep    func(t *testing.T, e *hookEnv)
		brk     func(t *testing.T, e *hookEnv)
		fix     func(t *testing.T, e *hookEnv)
		broken  func(e *hookEnv) (stdin, projectDir string) // nil: the ordinary input
		fixed   func(e *hookEnv) (stdin, projectDir string) // nil: the ordinary input
	}{
		{
			stage: stageOpen, event: "Stop",
			brk: func(t *testing.T, e *hookEnv) {
				if err := os.MkdirAll(filepath.Dir(e.store), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(e.store, []byte("this is not a SQLite database"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			fix: func(t *testing.T, e *hookEnv) {
				if err := os.Remove(e.store); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			stage: stageBackfill, event: "Stop",
			prep: createStore,
			// A store stamped one shape back is what an upgrade leaves: the next open rebuilds it empty.
			brk: func(t *testing.T, e *hookEnv) {
				execStore(t, e.store, fmt.Sprintf(`PRAGMA user_version = %d`, catalogue.UserVersion-1))
			},
			fix: func(t *testing.T, e *hookEnv) {
				db, err := catalogue.Open(e.store, io.Discard)
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				if err := catalogue.MarkBackfilled(db, time.Now().Add(time.Second)); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			// A marker that exists and does not parse is a failure of the check, never "not pending".
			name: "catalogue-backfill/unreadable-markers", stage: stageBackfill, event: "Stop",
			prep: createStore,
			brk: func(t *testing.T, e *hookEnv) {
				execStore(t, e.store, `INSERT INTO meta(key,value) VALUES('rebuilt_at','yesterday')`)
			},
			fix: func(t *testing.T, e *hookEnv) { execStore(t, e.store, `DELETE FROM meta WHERE key = 'rebuilt_at'`) },
		},
		{
			stage: stageRegister, event: "Stop",
			prep: createStore,
			brk: func(t *testing.T, e *hookEnv) {
				refuse(t, e, "no_session_insert", "BEFORE INSERT ON session")
				refuse(t, e, "no_session_update", "BEFORE UPDATE ON session")
			},
			fix: func(t *testing.T, e *hookEnv) { allow(t, e, "no_session_insert", "no_session_update") },
		},
		{
			stage: stageIngest, event: "SessionEnd",
			prep: createStore,
			brk:  func(t *testing.T, e *hookEnv) { refuse(t, e, "no_act", "BEFORE INSERT ON act") },
			fix:  func(t *testing.T, e *hookEnv) { allow(t, e, "no_act") },
		},
		{
			// Closure settles sessions that are not live: the fixture's session, ingested at a Stop, is
			// closed by the SessionStart of a different one.
			stage: stageClosure, event: "SessionStart",
			prep: func(t *testing.T, e *hookEnv) {
				hook(t, e.input(), e.project, noon, "Stop")
			},
			brk: func(t *testing.T, e *hookEnv) {
				refuse(t, e, "no_close", "BEFORE UPDATE OF closed_at ON session")
			},
			fix:    func(t *testing.T, e *hookEnv) { allow(t, e, "no_close") },
			broken: func(e *hookEnv) (string, string) { return e.inputAs("sess-next"), e.project },
			fixed:  func(e *hookEnv) (string, string) { return e.inputAs("sess-next"), e.project },
		},
		{
			// Retention deletes rows older than the window; a row at the epoch is always one.
			stage: stageRetention, event: "SessionStart",
			prep: func(t *testing.T, e *hookEnv) {
				hook(t, e.input(), e.project, noon, "Stop")
				execStore(t, e.store, `UPDATE act SET ts = 0`)
			},
			brk: func(t *testing.T, e *hookEnv) { refuse(t, e, "no_delete", "BEFORE DELETE ON act") },
			fix: func(t *testing.T, e *hookEnv) { allow(t, e, "no_delete") },
		},
		{
			stage: stageRetentionMarker, event: "SessionStart",
			prep: createStore,
			brk: func(t *testing.T, e *hookEnv) {
				refuse(t, e, "no_mark", "BEFORE INSERT ON meta WHEN NEW.key = 'retained_on'")
			},
			fix: func(t *testing.T, e *hookEnv) { allow(t, e, "no_mark") },
		},
		{
			stage: stageManifest, event: "SubagentStop", project: true,
			brk: func(t *testing.T, e *hookEnv) {
				if err := os.MkdirAll(filepath.Join(e.project, ".claude"), 0o755); err != nil {
					t.Fatal(err)
				}
				// A FILE where the manifest's directory belongs: MkdirAll cannot create it.
				if err := os.WriteFile(filepath.Join(e.project, ".claude", "gray-area"), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			fix: func(t *testing.T, e *hookEnv) {
				if err := os.Remove(filepath.Join(e.project, ".claude", "gray-area")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			// A DIRECTORY where the manifest file belongs: the directory exists, the open fails.
			name: "manifest/open", stage: stageManifest, event: "SessionStart", project: true,
			brk: func(t *testing.T, e *hookEnv) {
				if err := os.MkdirAll(manifestPath(e.project, e.sid), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			fix: func(t *testing.T, e *hookEnv) {
				if err := os.Remove(manifestPath(e.project, e.sid)); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			stage: stageProjectRoot, event: "SubagentStop",
			broken: func(e *hookEnv) (string, string) { return `{"session_id":"` + e.sid + `"}`, "" },
		},
		{
			stage: stageSessionPath, event: "SessionStart",
			broken: func(e *hookEnv) (string, string) {
				return `{"session_id":"` + e.sid + `","cwd":` + strconv.Quote(e.project) + `}`, e.project
			},
		},
	}
	for _, tc := range cases {
		name := tc.name
		if name == "" {
			name = string(tc.stage)
		}
		t.Run(name, func(t *testing.T) {
			e := newHookEnv(t)
			project := ""
			if tc.project {
				project = e.project
			}
			if tc.prep != nil {
				tc.prep(t, e)
			}
			if tc.brk != nil {
				tc.brk(t, e)
			}
			stdin, dir := e.input(), e.project
			if tc.broken != nil {
				stdin, dir = tc.broken(e)
			}
			msg, stderr := hook(t, stdin, dir, noon, tc.event)

			f, ok := entryFor(recorded(t), tc.stage, project)
			if !ok {
				t.Fatalf("%s failed and nothing recorded it; record: %+v\nstderr: %s", tc.stage, recorded(t), stderr)
			}
			if f.Error == "" || f.Event != tc.event || !f.Since.Equal(noon) || !f.Last.Equal(noon) {
				t.Errorf("entry = %+v, want an error, event %s, since and last at the hook's clock", f, tc.event)
			}
			if !strings.Contains(stderr, "gray-area-capture: "+f.Error) {
				t.Errorf("the debug log lost its line: stderr %q does not carry %q", stderr, f.Error)
			}
			if hookfailures.Displays(tc.event) {
				if !strings.Contains(msg, "- "+string(tc.stage)) || !strings.Contains(msg, f.Error) {
					t.Errorf("%s displays, and its message does not name the failure: %q", tc.event, msg)
				}
				if f.Notified.IsZero() {
					t.Error("shown, but the entry does not record that it was")
				}
			} else if msg != "" {
				t.Errorf("%s must never print a message, printed %q", tc.event, msg)
			}

			if tc.fix != nil {
				tc.fix(t, e)
			}
			stdin, dir = e.input(), e.project
			if tc.fixed != nil {
				stdin, dir = tc.fixed(e)
			}
			hook(t, stdin, dir, noon.Add(time.Minute), tc.event)
			if f, ok := entryFor(recorded(t), tc.stage, project); ok {
				t.Errorf("the stage worked and its entry stayed: %+v", f)
			}
		})
	}
}

// A HEALTHY HOOK IS INVISIBLE. No message on any event, and no record — not even its directory —
// because a file every turn writes is a cost every turn pays, and "no file" is what healthy means.
func TestAHealthyHookWritesNoRecordAndSaysNothing(t *testing.T) {
	e := newHookEnv(t)
	for _, ev := range []string{"SessionStart", "SubagentStop", "Stop", "SessionEnd"} {
		if msg, stderr := hook(t, e.input(), e.project, noon, ev); msg != "" {
			t.Errorf("%s spoke on a healthy machine: %q (stderr %q)", ev, msg, stderr)
		}
	}
	if _, err := os.Stat(filepath.Dir(recordPath(t))); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a healthy run created the record's directory (stat err %v)", err)
	}
}

// THE RECORD IS WHAT CARRIES A FAILURE FROM AN EVENT NOBODY SEES TO ONE SOMEBODY DOES. The Stop below
// runs in a different, healthy project: the message can only have come from the record.
func TestASilentEventsFailureIsShownByTheNextDisplayingEvent(t *testing.T) {
	e := newHookEnv(t)
	broken := t.TempDir()
	if err := os.MkdirAll(filepath.Join(broken, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, ".claude", "gray-area"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	in := `{"session_id":"s","agent_id":"a","agent_type":"t","cwd":` + strconv.Quote(broken) + `}`
	if msg, _ := hook(t, in, broken, noon, "SubagentStop"); msg != "" {
		t.Fatalf("SubagentStop printed %q", msg)
	}

	msg, _ := hook(t, e.input(), e.project, noon.Add(time.Minute), "Stop")
	if !strings.Contains(msg, "- manifest in "+broken+": cannot create manifest dir") {
		t.Fatalf("the next Stop did not show the SubagentStop failure: %q", msg)
	}
	if !strings.Contains(msg, "last at SubagentStop") || !strings.Contains(msg, recordPath(t)) {
		t.Errorf("the message does not say where the failure happened or where the record is: %q", msg)
	}
}

// A MANIFEST THAT WORKS IN ONE PROJECT SAYS NOTHING ABOUT ANOTHER'S.
func TestManifestEntriesArePerProject(t *testing.T) {
	e := newHookEnv(t)
	broken := t.TempDir()
	if err := os.MkdirAll(filepath.Join(broken, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, ".claude", "gray-area"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	hook(t, `{"session_id":"s","agent_id":"a"}`, broken, noon, "SubagentStop")
	hook(t, `{"session_id":"s","agent_id":"b"}`, e.project, noon, "SubagentStop")
	if _, ok := entryFor(recorded(t), stageManifest, broken); !ok {
		t.Fatalf("a success in %s cleared the failure in %s: %+v", e.project, broken, recorded(t))
	}
	if _, ok := entryFor(recorded(t), stageManifest, e.project); ok {
		t.Errorf("the working project has an entry: %+v", recorded(t))
	}
}

// Stop fires every turn. A failure that lasts is shown once, again after notifyEvery, and not
// between — and a NEW failure is not held back by an older one's quiet period.
func TestAFailingStageIsShownAtMostOnceEveryNotifyPeriod(t *testing.T) {
	e := newHookEnv(t)
	if err := os.MkdirAll(filepath.Dir(e.store), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(e.store, []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}

	if msg, _ := hook(t, e.input(), e.project, noon, "Stop"); !strings.Contains(msg, "- catalogue-open") {
		t.Fatalf("first Stop: %q", msg)
	}
	if msg, _ := hook(t, e.input(), e.project, noon.Add(2*time.Minute), "Stop"); msg != "" {
		t.Fatalf("shown again 2 minutes later: %q", msg)
	}

	// A new failure at +5m: the SessionStart shows the transcript-path alarm, and only that.
	noPath := `{"session_id":"` + e.sid + `","cwd":` + strconv.Quote(e.project) + `}`
	msg, _ := hook(t, noPath, e.project, noon.Add(5*time.Minute), "SessionStart")
	if !strings.Contains(msg, "- session-transcript-path") || strings.Contains(msg, "catalogue-open") {
		t.Fatalf("the new failure should be shown alone: %q", msg)
	}
	if !strings.HasPrefix(msg, "gray-area-capture: 1 stage(s)") {
		t.Errorf("the count does not match what is listed: %q", msg)
	}

	if msg, _ := hook(t, e.input(), e.project, noon.Add(hookfailures.NotifyEvery-time.Second), "Stop"); msg != "" {
		t.Fatalf("shown a second before the period ended: %q", msg)
	}
	msg, _ = hook(t, e.input(), e.project, noon.Add(hookfailures.NotifyEvery), "Stop")
	if !strings.Contains(msg, "- catalogue-open") || strings.Contains(msg, "session-transcript-path") {
		t.Fatalf("at the period, only the stage last shown at noon is due: %q", msg)
	}
	f, _ := entryFor(recorded(t), stageOpen, "")
	if !f.Since.Equal(noon) || !f.Last.Equal(noon.Add(hookfailures.NotifyEvery)) {
		t.Errorf("since/last = %s/%s: a failure that persists keeps its first time and moves its last", f.Since, f.Last)
	}
}

// WITHOUT A RECORD THERE IS NO THROTTLE, AND THE FAILURE IS STILL SHOWN. A record that cannot be
// written must not become a second silent failure hiding the first.
func TestWithNoRecordEveryDisplayingEventStillShowsTheFailure(t *testing.T) {
	e := newHookEnv(t)
	if err := os.MkdirAll(filepath.Dir(filepath.Dir(recordPath(t))), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Dir(recordPath(t)), []byte("a file where the record's directory belongs"), 0o644); err != nil {
		t.Fatal(err)
	}
	noPath := `{"session_id":"` + e.sid + `","cwd":` + strconv.Quote(e.project) + `}`
	for i := range 2 {
		msg, stderr := hook(t, noPath, e.project, noon.Add(time.Duration(i)*time.Second), "SessionStart")
		if !strings.Contains(msg, "- session-transcript-path") || !strings.Contains(msg, "the failure record cannot be kept") {
			t.Fatalf("SessionStart %d: %q", i, msg)
		}
		if !strings.Contains(stderr, "failure record not kept") {
			t.Errorf("the debug log does not say the record was not kept: %q", stderr)
		}
	}
	if msg, _ := hook(t, `{"session_id":"s"}`, "", noon, "SubagentStop"); msg != "" {
		t.Errorf("SubagentStop printed %q", msg)
	}
}

// A WRITE THAT FAILS after the manifest opened: /dev/full accepts the open and refuses every write.
func TestAManifestAppendThatFailsIsRecorded(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("no /dev/full on this platform")
	}
	e := newHookEnv(t)
	p := manifestPath(e.project, e.sid)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/dev/full", p); err != nil {
		t.Skip("cannot symlink:", err)
	}
	hook(t, `{"session_id":"`+e.sid+`","agent_id":"a"}`, e.project, noon, "SubagentStop")
	f, ok := entryFor(recorded(t), stageManifest, e.project)
	if !ok || !strings.Contains(f.Error, "cannot append row") {
		t.Fatalf("a failed append is not recorded as one: %+v", recorded(t))
	}
}
