package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// project builds a qlty-initialised temp project holding one file, and returns
// the project dir and the absolute path of that file.
func project(t *testing.T, name string) (dir, file string) {
	t.Helper()
	dir = t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".qlty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".qlty", "qlty.toml"), []byte("config_version = \"0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	file = filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, file
}

func healthy(dir string) env {
	return env{projectDir: dir, qltyPresent: true, mode: modeFull, timeout: time.Second, exec: nil}
}

func TestDecideSkips(t *testing.T) {
	dir, file := project(t, "a.go")
	outside := filepath.Join(t.TempDir(), "elsewhere.go")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	bare := t.TempDir() // no .qlty/qlty.toml

	cases := []struct {
		name   string
		env    env
		file   string
		expect string
		warn   bool
	}{
		{"mode off", func() env { e := healthy(dir); e.mode = modeOff; return e }(), file, "SC_QUALITY_GATE=off", false},
		{"no file in payload", healthy(dir), "unknown", "no file", false},
		{"qlty absent warns and points at the doctor",
			func() env { e := healthy(dir); e.qltyPresent = false; return e }(), file, "doctor --fix", true},
		{"no project dir", func() env { e := healthy(dir); e.projectDir = ""; return e }(), file, "CLAUDE_PROJECT_DIR", false},
		{"project not qlty-initialised is silent", healthy(bare), filepath.Join(bare, "a.go"), "qlty init", false},
		{"file outside the project", healthy(dir), outside, "outside the project", false},
		{"file does not exist", healthy(dir), filepath.Join(dir, "ghost.go"), "not a readable file", false},
		{"directory is not a file", healthy(dir), filepath.Join(dir, ".qlty"), "not a readable file", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := decide(c.env, c.file)
			if p.run {
				t.Fatalf("expected a skip, got a run plan %+v", p)
			}
			if !strings.Contains(p.skip, c.expect) {
				t.Errorf("skip = %q; missing %q", p.skip, c.expect)
			}
			if p.warn != c.warn {
				t.Errorf("warn = %v, want %v (only actionable machine state earns stderr)", p.warn, c.warn)
			}
		})
	}
}

func TestDecideRuns(t *testing.T) {
	dir, file := project(t, filepath.Join("pkg", "a.go"))

	p := decide(healthy(dir), file)
	if !p.run || !p.format {
		t.Fatalf("full mode must run fmt + check, got %+v", p)
	}
	if p.rel != filepath.Join("pkg", "a.go") {
		t.Errorf("qlty must be handed a project-relative path, got %q", p.rel)
	}

	e := healthy(dir)
	e.mode = modeCheck
	if p := decide(e, file); !p.run || p.format {
		t.Fatalf("check-only must run without rewriting, got %+v", p)
	}
}

func TestParseMode(t *testing.T) {
	for in, want := range map[string]string{
		"off": modeOff, "OFF": modeOff, "0": modeOff, "false": modeOff, "disabled": modeOff, "no": modeOff,
		"check": modeCheck, "check-only": modeCheck, " CheckOnly ": modeCheck,
		"": modeFull, "on": modeFull, "anything": modeFull,
	} {
		if got := parseMode(in); got != want {
			t.Errorf("parseMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseTimeout(t *testing.T) {
	for in, want := range map[string]time.Duration{
		"":     defaultTimeout,
		"abc":  defaultTimeout,
		"0":    defaultTimeout,
		"-5":   defaultTimeout,
		"45":   45 * time.Second,
		" 5 ":  5 * time.Second,
		"9999": maxTimeout, // an unbounded hook timeout would hang every edit
	} {
		if got := parseTimeout(in); got != want {
			t.Errorf("parseTimeout(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestWithinRejectsEscapes(t *testing.T) {
	dir := t.TempDir()
	if _, ok := within(dir, filepath.Join(dir, "..", "escape.go")); ok {
		t.Error("a path climbing out of the project must be rejected")
	}
	if rel, ok := within(dir, "rel/a.go"); !ok || rel != filepath.Join("rel", "a.go") {
		t.Errorf("a relative path is resolved against the project dir: %q %v", rel, ok)
	}
}

// TestWithinRejectsSymlinkEscape is the case lexical path arithmetic cannot see.
//
// This gate does not merely READ what within() admits: in the default mode it runs
// `qlty fmt`, which REWRITES the file. So a symlink inside the project pointing at a
// sibling tree turns an ordinary Edit into a modification outside the working tree —
// the act agent-guardrails says requires explicit approval.
//
// internal/checkpoint closed exactly this class at 67eb8fa ("containment is os.Root,
// not path arithmetic") after measuring that filepath.Rel reports "contained: true"
// for a symlinked directory. That fix was an INSTANCE fix; this call site kept the
// arithmetic. Containment here is a security property, so it needs the syscall-level
// check, not the string one.
func TestWithinRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privilege on Windows")
	}
	dir, _ := project(t, "a.go")

	outside := t.TempDir()
	victim := filepath.Join(outside, "victim.go")
	if err := os.WriteFile(victim, []byte("not ours to rewrite\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "linked")); err != nil {
		t.Fatal(err)
	}

	through := filepath.Join(dir, "linked", "victim.go")
	if rel, ok := within(dir, through); ok {
		t.Errorf("within() admitted %q as project-relative %q — qlty fmt would rewrite a file outside the project", through, rel)
	}
	if p := decide(healthy(dir), through); p.run {
		t.Errorf("decide() planned to run qlty on a symlinked path outside the project: %+v", p)
	}

	// A file that is genuinely absent but genuinely INSIDE must stay
	// distinguishable from an escape: the operator reads a different reason for
	// each, and conflating them would make this fix report the wrong cause.
	if p := decide(healthy(dir), filepath.Join(dir, "ghost.go")); !strings.Contains(p.skip, "not a readable file") {
		t.Errorf("a missing in-project file must read as unreadable, not as an escape: %q", p.skip)
	}
}

func TestClampBoundsFeedback(t *testing.T) {
	long := strings.Repeat("issue line\n", maxFeedbackLn*2)
	got := clamp(long)
	if n := strings.Count(got, "\n") + 1; n > maxFeedbackLn+2 {
		t.Errorf("clamp left %d lines", n)
	}
	if !strings.Contains(got, "truncated") {
		t.Error("truncation must be visible, not silent")
	}
	if len(clamp(strings.Repeat("x", maxFeedbackLen*2))) > maxFeedbackLen+80 {
		t.Error("character bound not applied")
	}
	if got := clamp("  three issues  "); got != "three issues" {
		t.Errorf("short output must pass through trimmed, got %q", got)
	}
}

func TestFileFrom(t *testing.T) {
	var in hookInput
	in.ToolInput.Path = "p.txt"
	if got := fileFrom(in); got != "p.txt" {
		t.Fatalf("fileFrom should use path: %q", got)
	}
	in.ToolInput.FilePath = "f.txt"
	if got := fileFrom(in); got != "f.txt" {
		t.Fatalf("file_path should win over path: %q", got)
	}
}

// The real process boundary, exercised against a stand-in for qlty so the
// exit-code and timeout paths are proven rather than assumed.
func TestExecQlty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh as a stand-in for qlty")
	}
	dir := t.TempDir()

	out, code, err := execQlty(context.Background(), dir, "sh", "-c", "echo hi; exit 3")
	if err != nil {
		t.Fatalf("a non-zero exit is data, not an error: %v", err)
	}
	if code != 3 || !strings.Contains(out, "hi") {
		t.Fatalf("code=%d out=%q", code, out)
	}

	if _, _, err := execQlty(context.Background(), dir, "definitely-not-a-real-binary-xyz"); err == nil {
		t.Error("an unlaunchable binary must surface as an error")
	}

	// A grandchild holding the output pipe open is the realistic hang: qlty dies
	// on the deadline, the linter it spawned does not. WaitDelay must bound it.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, _, err := execQlty(ctx, dir, "sh", "-c", "sleep 30 & wait"); err == nil {
		t.Error("a hung qlty must surface as an error, not a hang")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("the deadline is not a real ceiling: took %v", elapsed)
	}
}
