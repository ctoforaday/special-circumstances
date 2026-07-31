package posttooluse

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hooklog"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

var noon = time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

func unit(name string, r hookunit.Result) hookunit.Unit {
	r.Name = name
	return hookunit.Unit{Name: name, Run: func(*hookunit.Ctx) hookunit.Result { return r }}
}

func fire(t *testing.T, dir string, units []hookunit.Unit) (stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(nil, strings.NewReader(`{"tool_name":"Edit","tool_input":{"file_path":"a.go"}}`),
		&o, &e, dir, noon, units)
	return e.String(), code
}

// Exit 2 at PostToolUse is a FEEDBACK channel — the write already happened. Any unit
// wanting it wins, and no unit's message may be lost because another had nothing to say.
func TestMergeKeepsEveryMessageAndTheHighestExit(t *testing.T) {
	stderr, exit, logs := merge([]hookunit.Result{
		{Name: "a", Stderr: "quality findings", Exit: 2, Log: "a-log\n"},
		{Name: "b", Stderr: "", Exit: 0, Log: "b-log\n"},
		{Name: "c", Stderr: "index note", Exit: 0, Log: "c-log\n"},
	})
	if exit != 2 {
		t.Errorf("exit = %d; any unit wanting 2 must win", exit)
	}
	for _, want := range []string{"quality findings", "index note"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lost %q: %q", want, stderr)
		}
	}
	if len(logs) != 3 {
		t.Errorf("logs = %v; every unit's line must reach the single writer", logs)
	}
	// All quiet means quiet.
	if s, e, _ := merge([]hookunit.Result{{Name: "a"}, {Name: "b"}}); s != "" || e != 0 {
		t.Errorf("silent units must produce silence: %q %d", s, e)
	}
}

// The event's contract: a finding reaches the model via exit 2 and stderr.
func TestRunSurfacesFindingsAndExitsTwo(t *testing.T) {
	dir := t.TempDir()
	stderr, code := fire(t, dir, []hookunit.Unit{
		unit("gate", hookunit.Result{Stderr: "issues in a.go", Exit: 2, Log: "gate fired\n"}),
		unit("index", hookunit.Result{Log: "index fired\n"}),
	})
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "issues in a.go") {
		t.Errorf("stderr = %q", stderr)
	}
}

// ONE writer, in unit order — the hazard the merge exists to remove.
func TestRunWritesEveryLogLineOnce(t *testing.T) {
	dir := t.TempDir()
	fire(t, dir, []hookunit.Unit{
		unit("gate", hookunit.Result{Log: "first\n"}),
		unit("index", hookunit.Result{Log: "second\n"}),
	})
	b, err := os.ReadFile(hooklog.Path(dir))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "first\nsecond\n" {
		t.Errorf("log = %q; want both lines once, in unit order", string(b))
	}
}

// A panicking unit must not cost the other unit's finding, nor the event.
func TestRunSurvivesAPanickingUnit(t *testing.T) {
	dir := t.TempDir()
	stderr, code := fire(t, dir, []hookunit.Unit{
		{Name: "boom", Run: func(*hookunit.Ctx) hookunit.Result { panic("gate exploded") }},
		unit("index", hookunit.Result{Stderr: "index still ran", Log: "index\n"}),
	})
	if code != 0 {
		t.Errorf("a panic must not change the exit code: %d", code)
	}
	if !strings.Contains(stderr, "panicked and was isolated") {
		t.Errorf("the panic must be reported: %q", stderr)
	}
	if !strings.Contains(stderr, "index still ran") {
		t.Errorf("the surviving unit's message was lost: %q", stderr)
	}
}

// A hook that cannot locate the project must not write a relative .claude/ wherever it
// started — and must still not fail the tool call.
func TestNoProjectDirWritesNothingRelative(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)
	var o, e bytes.Buffer
	if code := run(nil, strings.NewReader(`{"tool_name":"Edit"}`), &o, &e, "", noon,
		[]hookunit.Unit{unit("x", hookunit.Result{Log: "x\n"})}); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if _, err := os.Stat(filepath.Join(wd, ".claude")); !os.IsNotExist(err) {
		t.Error("wrote a relative .claude/ with no project root")
	}
}

func TestVersionAndBadFlagNeverBlock(t *testing.T) {
	var o, e bytes.Buffer
	if code := run([]string{"-version"}, strings.NewReader(""), &o, &e, t.TempDir(), noon, nil); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(o.String(), "sc-posttooluse") {
		t.Errorf("version output = %q", o.String())
	}
	if code := run([]string{"-nonsense"}, strings.NewReader(""), &o, &e, t.TempDir(), noon, nil); code != 0 {
		t.Errorf("a bad flag must never fail the Edit/Write: %d", code)
	}
}

// The real units must carry their own matchers, so merging cannot widen what they act on.
func TestRealUnitsKeepTheirOwnMatchers(t *testing.T) {
	for _, u := range Units() {
		if u.Applies == nil {
			t.Fatalf("%s has no matcher — a merged binary registers the UNION, so each unit must still say whether a call is its business", u.Name)
		}
		if u.Applies(hookunit.NewCtx("PostToolUse", []byte(`{"tool_name":"Bash"}`), "/p", noon)) {
			t.Errorf("%s claimed a Bash call; it is a Write|Edit unit", u.Name)
		}
		if !u.Applies(hookunit.NewCtx("PostToolUse", []byte(`{"tool_name":"Edit"}`), "/p", noon)) {
			t.Errorf("%s refused an Edit call", u.Name)
		}
	}
}
