package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

// call drives the hook as the harness does.
func call(t *testing.T, stdin, projectDir string, qlty bool, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(stdin), &o, &e, projectDir, qlty)
	return o.String(), e.String(), code
}

// The stated contract: absent tool means ONE warning and no failure; present tool
// means silence. Either way the Edit/Write is never blocked (exit 0).
func TestRunNeverBlocksAndWarnsOnlyWhenDegraded(t *testing.T) {
	cases := []struct {
		name       string
		qlty       bool
		stdin      string
		wantStderr []string // empty = expect silence
		wantInLog  []string
	}{
		{
			"qlty absent warns once and points at the doctor",
			false,
			`{"tool_name":"Write","tool_input":{"file_path":"a.go"}}`,
			[]string{"skipped", "doctor --fix", "a.go"},
			[]string{"sc-quality-gate", "Write", "a.go", "skipped"},
		},
		{
			"qlty present is silent but still logged",
			true,
			`{"tool_name":"Edit","tool_input":{"file_path":"b.md"}}`,
			nil,
			[]string{"sc-quality-gate", "Edit", "b.md", "qlty fmt"},
		},
		{
			"path field is used when file_path is absent",
			true,
			`{"tool_name":"NotebookEdit","tool_input":{"path":"nb.ipynb"}}`,
			nil,
			[]string{"nb.ipynb"},
		},
		{
			"malformed payload still logs and never blocks",
			false,
			`{"tool_name":`,
			[]string{"unknown"},
			[]string{"unknown"},
		},
		{
			"empty stdin still logs and never blocks",
			true,
			``,
			nil,
			[]string{"unknown"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			stdout, stderr, code := call(t, c.stdin, dir, c.qlty)
			if code != 0 {
				t.Fatalf("exit %d; the gate must never fail the tool call", code)
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
			logged := readLog(t, dir)
			for _, w := range c.wantInLog {
				if !strings.Contains(logged, w) {
					t.Errorf("hook log = %q; missing %q", logged, w)
				}
			}
		})
	}
}

// Every firing appends; the log is the evidence that the hook reached here at all
// (including inside a subagent), so a second firing must not overwrite the first.
func TestHookLogAppends(t *testing.T) {
	dir := t.TempDir()
	call(t, `{"tool_name":"Write","tool_input":{"file_path":"first.go"}}`, dir, true)
	call(t, `{"tool_name":"Write","tool_input":{"file_path":"second.go"}}`, dir, true)
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
	if _, _, code := call(t, `{"tool_name":"Write","tool_input":{"file_path":"a.go"}}`, "", true); code != 0 {
		t.Fatalf("exit %d with no project dir", code)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); !os.IsNotExist(err) {
		t.Fatal("empty project dir wrote a relative .claude/ into the working directory")
	}

	// A project dir occupied by a FILE where .claude/ must go: MkdirAll fails.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".claude"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, code := call(t, `{"tool_name":"Write","tool_input":{"file_path":"a.go"}}`, dir, true); code != 0 {
		t.Fatalf("exit %d when the log directory cannot be created", code)
	}
}

func TestVersionFlagPrintsAndSkipsTheGate(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := call(t, "", dir, false, "-version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "sc-quality-gate") || !strings.Contains(stdout, version) {
		t.Fatalf("version output = %q", stdout)
	}
	if stderr != "" {
		t.Errorf("-version must not emit the degraded warning: %q", stderr)
	}
	if _, err := os.Stat(logPath(dir)); !os.IsNotExist(err) {
		t.Error("-version must not record a hook firing")
	}
}

// A flag the binary does not know is an operator typo, not a reason to fail an Edit.
func TestUnknownFlagNeverBlocks(t *testing.T) {
	if _, _, code := call(t, "", t.TempDir(), true, "-nope"); code != 0 {
		t.Fatalf("exit %d on an unknown flag", code)
	}
}
