package qualitygate

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
)

// logPath is where the hook's instrumentation lands under a project dir.
func logPath(projectDir string) string {
	return filepath.Join(projectDir, ".claude", "prosthetic-conscience-hook.log")
}

func readLog(t *testing.T, projectDir string) string {
	t.Helper()
	b, err := os.ReadFile(logPath(projectDir))
	if err != nil {
		t.Fatalf("hook log not written: %v", err)
	}
	return string(b)
}

// fakeQlty stands in for the qlty process: it records the argv of every
// invocation and replays scripted results, so both the clean and the
// issues-found branches are reachable without qlty on the machine.
type fakeQlty struct {
	calls   [][]string
	code    map[string]int    // subcommand -> exit code
	out     map[string]string // subcommand -> combined output
	err     map[string]error  // subcommand -> launch/timeout failure
	rewrite string            // path fmt rewrites, absolute; "" = leaves files alone
}

func (f *fakeQlty) runner() runner {
	return func(_ context.Context, dir, name string, args ...string) (string, int, error) {
		f.calls = append(f.calls, append([]string{name}, args...))
		sub := ""
		if len(args) > 0 {
			sub = args[0]
		}
		if sub == "fmt" && f.rewrite != "" {
			_ = os.WriteFile(f.rewrite, []byte("reformatted\n"), 0o644)
		}
		return f.out[sub], f.code[sub], f.err[sub]
	}
}

func (f *fakeQlty) ran(sub string) bool {
	for _, c := range f.calls {
		if len(c) > 1 && c[1] == sub {
			return true
		}
	}
	return false
}

func (f *fakeQlty) argv(sub string) []string {
	for _, c := range f.calls {
		if len(c) > 1 && c[1] == sub {
			return c
		}
	}
	return nil
}

// call drives the hook as the harness does.
func call(t *testing.T, stdin string, e env, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, b bytes.Buffer
	code = run(args, strings.NewReader(stdin), &o, &b, e)
	return o.String(), b.String(), code
}

func payload(file string) string {
	return `{"tool_name":"Write","tool_input":{"file_path":` + quote(file) + `}}`
}

func quote(s string) string { return `"` + strings.ReplaceAll(s, `\`, `\\`) + `"` }

// The headline behaviour: the gate actually executes qlty.
func TestRunExecutesQltyAndReportsIssues(t *testing.T) {
	dir, file := project(t, "a.py")
	f := &fakeQlty{
		code: map[string]int{"check": 1},
		out:  map[string]string{"check": "a.py:1 medium unused import  ruff:F401"},
	}
	e := healthy(dir)
	e.exec = f.runner()

	_, stderr, code := call(t, payload(file), e)

	if !f.ran("fmt") || !f.ran("check") {
		t.Fatalf("both qlty commands must run, got %v", f.calls)
	}
	if code != 2 {
		t.Fatalf("findings must reach the model (exit 2), got %d", code)
	}
	if !strings.Contains(stderr, "ruff:F401") || !strings.Contains(stderr, "a.py") {
		t.Errorf("stderr must carry the findings, got %q", stderr)
	}
	if !strings.Contains(readLog(t, dir), "issues reported") {
		t.Errorf("log = %q", readLog(t, dir))
	}
}

// Flags are part of the contract: --no-fix keeps a future qlty default from
// rewriting files behind the model's back, --no-progress keeps the spinner out.
func TestCheckIsInvokedWithGuardFlags(t *testing.T) {
	dir, file := project(t, "a.py")
	f := &fakeQlty{}
	e := healthy(dir)
	e.exec = f.runner()

	call(t, payload(file), e)

	argv := strings.Join(f.argv("check"), " ")
	for _, want := range []string{"qlty check", "--no-progress", "--no-fix", "--no-error", "a.py"} {
		if !strings.Contains(argv, want) {
			t.Errorf("check argv = %q; missing %q", argv, want)
		}
	}
	if argv := strings.Join(f.argv("fmt"), " "); !strings.Contains(argv, "--trigger agent") {
		t.Errorf("fmt argv = %q; must declare the agent trigger", argv)
	}
}

// A clean file is silent: no feedback, no exit 2, but still logged.
func TestCleanFileIsSilent(t *testing.T) {
	dir, file := project(t, "a.py")
	f := &fakeQlty{}
	e := healthy(dir)
	e.exec = f.runner()

	_, stderr, code := call(t, payload(file), e)
	if code != 0 || stderr != "" {
		t.Fatalf("clean must be silent: code=%d stderr=%q", code, stderr)
	}
	if !strings.Contains(readLog(t, dir), "clean") {
		t.Errorf("log = %q", readLog(t, dir))
	}
}

// qlty fmt rewrites the file AFTER the Write, which makes the model's cached
// copy stale — the next Edit would fail on a string mismatch. Say so.
func TestFormatRewriteIsAnnounced(t *testing.T) {
	dir, file := project(t, "a.py")
	f := &fakeQlty{rewrite: file}
	e := healthy(dir)
	e.exec = f.runner()

	_, stderr, code := call(t, payload(file), e)
	if code != 2 {
		t.Fatalf("a rewritten file must be announced (exit 2), got %d", code)
	}
	low := strings.ToLower(stderr)
	if !strings.Contains(low, "rewrote") || !strings.Contains(low, "re-read") {
		t.Errorf("stderr must tell the model to re-read: %q", stderr)
	}
	// Asserted on meaning, not casing: the emphasis is deliberate and may move.
	if !strings.Contains(low, "stale") {
		t.Errorf("it must say WHY re-reading matters — a stale cached copy: %q", stderr)
	}
	if !strings.Contains(readLog(t, dir), "formatted") {
		t.Errorf("log = %q", readLog(t, dir))
	}
}

func TestCheckOnlyModeNeverRewrites(t *testing.T) {
	dir, file := project(t, "a.py")
	f := &fakeQlty{}
	e := healthy(dir)
	e.mode = modeCheck
	e.exec = f.runner()

	call(t, payload(file), e)
	if f.ran("fmt") {
		t.Fatal("check-only must not run the formatter")
	}
	if !f.ran("check") {
		t.Fatal("check-only must still check")
	}
}

// Infrastructure failure is never the Edit's problem: a timeout or an
// unlaunchable qlty exits 0 and leaves a trace in the log instead.
func TestInfrastructureFailureNeverBlocks(t *testing.T) {
	cases := []struct {
		name    string
		fake    *fakeQlty
		wantLog string
	}{
		{"check times out", &fakeQlty{err: map[string]error{"check": context.DeadlineExceeded}}, "qlty check could not run"},
		{"fmt cannot launch", &fakeQlty{err: map[string]error{"fmt": errors.New("exec: not found")}}, "qlty fmt could not run"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, file := project(t, "a.py")
			e := healthy(dir)
			e.exec = c.fake.runner()

			_, _, code := call(t, payload(file), e)
			if code != 0 {
				t.Fatalf("exit %d; infrastructure failure must not cost the tool call", code)
			}
			if !strings.Contains(readLog(t, dir), c.wantLog) {
				t.Errorf("log = %q; missing %q", readLog(t, dir), c.wantLog)
			}
		})
	}
}

// A fmt rewrite still has to be reported even when the check afterwards fails
// to run — the stale-cache warning is the part the model cannot recover without.
func TestRewriteSurvivesACheckFailure(t *testing.T) {
	dir, file := project(t, "a.py")
	f := &fakeQlty{rewrite: file, err: map[string]error{"check": context.DeadlineExceeded}}
	e := healthy(dir)
	e.exec = f.runner()

	_, stderr, code := call(t, payload(file), e)
	if code != 2 || !strings.Contains(stderr, "rewrote") {
		t.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

// The degraded path predates this change and must not regress: one warning
// line, exit 0, and qlty is never invoked.
func TestDegradedPathsWarnOnceAndNeverInvokeQlty(t *testing.T) {
	dir, file := project(t, "a.go")
	bare := t.TempDir()

	cases := []struct {
		name       string
		env        func(env) env
		file       string
		wantStderr []string // empty = expect silence
		wantInLog  []string
	}{
		{"qlty absent", func(e env) env { e.qltyPresent = false; return e }, file,
			[]string{"skipped", "doctor --fix", "a.go"}, []string{"sc-quality-gate", "Write", "skipped"}},
		{"gate switched off", func(e env) env { e.mode = modeOff; return e }, file,
			nil, []string{"SC_QUALITY_GATE=off"}},
		{"project not initialised", func(e env) env { e.projectDir = bare; return e }, filepath.Join(bare, "a.go"),
			nil, []string{"qlty init"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeQlty{}
			e := c.env(healthy(dir))
			e.exec = f.runner()

			stdout, stderr, code := call(t, payload(c.file), e)
			if code != 0 {
				t.Fatalf("exit %d; the gate must never fail the tool call", code)
			}
			if len(f.calls) != 0 {
				t.Fatalf("qlty must not be invoked: %v", f.calls)
			}
			if stdout != "" {
				t.Errorf("stdout must stay clean (PostToolUse), got %q", stdout)
			}
			if len(c.wantStderr) == 0 {
				if stderr != "" {
					t.Errorf("want silence on stderr, got %q", stderr)
				}
			} else {
				for _, w := range c.wantStderr {
					if !strings.Contains(stderr, w) {
						t.Errorf("stderr = %q; missing %q", stderr, w)
					}
				}
				if n := strings.Count(strings.TrimRight(stderr, "\n"), "\n"); n != 0 {
					t.Errorf("the warning must be one line, got %q", stderr)
				}
			}
			logged := readLog(t, e.projectDir)
			for _, w := range c.wantInLog {
				if !strings.Contains(logged, w) {
					t.Errorf("hook log = %q; missing %q", logged, w)
				}
			}
		})
	}
}

func TestMalformedPayloadStillLogsAndNeverBlocks(t *testing.T) {
	for _, stdin := range []string{`{"tool_name":`, ``, `{"tool_name":"Write","tool_input":{}}`} {
		dir, _ := project(t, "a.go")
		f := &fakeQlty{}
		e := healthy(dir)
		e.exec = f.runner()

		_, stderr, code := call(t, stdin, e)
		if code != 0 || stderr != "" {
			t.Errorf("stdin %q: code=%d stderr=%q", stdin, code, stderr)
		}
		if len(f.calls) != 0 {
			t.Errorf("stdin %q invoked qlty with no file: %v", stdin, f.calls)
		}
		if !strings.Contains(readLog(t, dir), "unknown") {
			t.Errorf("stdin %q: log = %q", stdin, readLog(t, dir))
		}
	}
}

// The path field is what NotebookEdit sends; it must reach qlty too.
func TestPathFieldIsUsedWhenFilePathIsAbsent(t *testing.T) {
	dir, file := project(t, "nb.ipynb")
	f := &fakeQlty{}
	e := healthy(dir)
	e.exec = f.runner()

	call(t, `{"tool_name":"NotebookEdit","tool_input":{"path":`+quote(file)+`}}`, e)
	if argv := strings.Join(f.argv("check"), " "); !strings.Contains(argv, "nb.ipynb") {
		t.Fatalf("check argv = %q", argv)
	}
}

// Every firing appends; the log is the evidence that the hook reached here at all
// (including inside a subagent), so a second firing must not overwrite the first.
func TestHookLogAppends(t *testing.T) {
	dir, first := project(t, "first.go")
	second := filepath.Join(dir, "second.go")
	if err := os.WriteFile(second, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	e := healthy(dir)
	e.exec = (&fakeQlty{}).runner()

	call(t, payload(first), e)
	call(t, payload(second), e)

	logged := readLog(t, dir)
	if !strings.Contains(logged, "first.go") || !strings.Contains(logged, "second.go") {
		t.Fatalf("second firing clobbered the first: %q", logged)
	}
	if n := strings.Count(strings.TrimRight(logged, "\n"), "\n"); n != 1 {
		t.Fatalf("want 2 log lines, got %q", logged)
	}
}

// Instrumentation is best-effort: no project dir, and no unwritable path, may cost
// the tool call. An empty dir must also NOT write a relative .claude/ into the cwd.
func TestLoggingFailureNeverBlocks(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	e := env{qltyPresent: true, mode: modeFull, timeout: time.Second, exec: (&fakeQlty{}).runner()}
	if _, _, code := call(t, `{"tool_name":"Write","tool_input":{"file_path":"a.go"}}`, e); code != 0 {
		t.Fatalf("exit %d with no project dir", code)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); !os.IsNotExist(err) {
		t.Fatal("empty project dir wrote a relative .claude/ into the working directory")
	}

	// A project dir occupied by a FILE where .claude/ must go: MkdirAll fails.
	dir, file := project(t, "a.go")
	if err := os.WriteFile(filepath.Join(dir, ".claude"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	e = healthy(dir)
	e.exec = (&fakeQlty{}).runner()
	if _, _, code := call(t, payload(file), e); code != 0 {
		t.Fatalf("exit %d when the log directory cannot be created", code)
	}
}

func TestVersionFlagPrintsAndSkipsTheGate(t *testing.T) {
	dir, _ := project(t, "a.go")
	f := &fakeQlty{}
	e := healthy(dir)
	e.exec = f.runner()

	stdout, stderr, code := call(t, "", e, "-version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "sc-quality-gate") || !strings.Contains(stdout, buildid.Revision()) {
		t.Fatalf("version output = %q", stdout)
	}
	if stderr != "" {
		t.Errorf("-version must not emit the degraded warning: %q", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("-version must not invoke qlty: %v", f.calls)
	}
	if _, err := os.Stat(logPath(dir)); !os.IsNotExist(err) {
		t.Error("-version must not record a hook firing")
	}
}

// A flag the binary does not know is an operator typo, not a reason to fail an Edit.
func TestUnknownFlagNeverBlocks(t *testing.T) {
	dir, _ := project(t, "a.go")
	e := healthy(dir)
	e.exec = (&fakeQlty{}).runner()
	if _, _, code := call(t, "", e, "-nope"); code != 0 {
		t.Fatalf("exit %d on an unknown flag", code)
	}
}

// Exit 2 at PostToolUse is a FEEDBACK channel, not a rejection: the write already happened
// and was not reverted. An agent that reads a non-zero exit as "my edit failed" re-applies
// it, duplicating the change or fighting the formatter. The message must say so BEFORE the
// findings, because that is the part read first (#212).
func TestFindingsSayTheWriteAlreadyHappened(t *testing.T) {
	dir, file := project(t, "a.go")
	e := healthy(dir)
	e.mode = modeCheck // check only: isolate the findings path from the rewrite path
	e.exec = func(_ context.Context, _, _ string, args ...string) (string, int, error) {
		return "a.go:1 some finding", 1, nil
	}

	_, stderr, code := call(t, payload(file), e)
	if code != 2 {
		t.Fatalf("findings must reach the model via exit 2, got %d", code)
	}
	low := strings.ToLower(stderr)
	for _, want := range []string{"already happened", "not a rejection", "do not re-apply"} {
		if !strings.Contains(low, want) {
			t.Errorf("the feedback must say the write stands — missing %q:\n%s", want, stderr)
		}
	}
	if !strings.Contains(low, "qlty check --no-fix") {
		t.Errorf("it must name how to see the full set: %q", stderr)
	}
	if strings.Index(low, "already happened") > strings.Index(low, "some finding") {
		t.Error("the reassurance must precede the findings — it is what an agent reads first")
	}
}
