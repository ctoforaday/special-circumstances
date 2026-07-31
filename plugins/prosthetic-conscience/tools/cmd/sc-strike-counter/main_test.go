package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/strikes"
)

var noon = time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

func fire(t *testing.T, dir string, in hookInput, now time.Time) (stdout, stderr string, code int) {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var o, e bytes.Buffer
	code = run(nil, bytes.NewReader(b), &o, &e, dir, now)
	return o.String(), e.String(), code
}

func bash(cmd string) hookInput {
	return hookInput{
		ToolName:  "Bash",
		ToolInput: json.RawMessage(`{"command":"` + cmd + `"}`),
		Error:     "Exit code 2\nsomething went wrong",
	}
}

// The third failure of the same call puts the count in front of the model — and the two
// before it stay quiet, because a hook that speaks on every failure is one nobody reads.
func TestThirdStrikeSpeaks(t *testing.T) {
	dir := t.TempDir()
	in := bash("go test ./...")

	for i := 1; i < strikes.Threshold; i++ {
		stdout, stderr, code := fire(t, dir, in, noon.Add(time.Duration(i)*time.Second))
		if code != 0 {
			t.Fatalf("strike %d: exit %d — this hook must never block", i, code)
		}
		if stdout != "" || stderr != "" {
			t.Fatalf("strike %d spoke early: stdout=%q stderr=%q", i, stdout, stderr)
		}
	}

	stdout, stderr, code := fire(t, dir, in, noon.Add(3*time.Second))
	if code != 0 {
		t.Fatalf("exit %d — the rule says the MODEL stops and escalates; a hook cannot do that for it", code)
	}

	// stderr reaches the operator; the JSON is the model-facing channel.
	for _, want := range []string{"anti-spinning", "3 times", "go test ./...", "STOP", "not a retry"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q:\n%s", want, stderr)
		}
	}
	var out hookOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("stdout is not the documented response shape (%v): %q", err, stdout)
	}
	if out.HookSpecificOutput.HookEventName != "PostToolUseFailure" {
		t.Errorf("hookEventName = %q", out.HookSpecificOutput.HookEventName)
	}
	if !strings.Contains(out.HookSpecificOutput.AdditionalContext, "anti-spinning") {
		t.Errorf("additionalContext = %q", out.HookSpecificOutput.AdditionalContext)
	}
	// The silence limit is stated, not glossed: a green hook is not evidence the loop
	// was broken, because a command can exit 0 while the work still failed.
	if !strings.Contains(stderr, "silence is not evidence") {
		t.Errorf("the coverage limit must travel with the warning:\n%s", stderr)
	}
}

// HONOR THE CANCEL. anti-spinning's other clause, and the field #127 did not anticipate:
// without this, a user pressing stop three times reads as a spin and the counter nags about
// the act of cancelling — the rule's two halves fighting each other.
func TestInterruptsAreNeverCounted(t *testing.T) {
	dir := t.TempDir()
	in := bash("long-running-thing")
	in.IsInterrupt = true

	for i := 0; i < strikes.Threshold+2; i++ {
		stdout, stderr, code := fire(t, dir, in, noon.Add(time.Duration(i)*time.Second))
		if code != 0 || stdout != "" || stderr != "" {
			t.Fatalf("an interrupt spoke: code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	}
	// And nothing was recorded, so a genuine failure afterwards starts from one.
	if _, err := os.Stat(strikes.Path(dir)); err == nil {
		s := strikes.Load(os.ReadFile, strikes.Path(dir))
		if len(s.Keys) != 0 {
			t.Errorf("interrupts left state behind: %v", s.Keys)
		}
	}
}

// A hook that cannot locate the project must do nothing and SAY it did nothing, rather than
// writing state into whatever directory it started in.
func TestNoProjectDirWritesNothing(t *testing.T) {
	in := bash("x")
	var o, e bytes.Buffer
	b, _ := json.Marshal(in)
	if code := run(nil, bytes.NewReader(b), &o, &e, "", noon); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(e.String(), "no project root") {
		t.Errorf("the no-root case must explain itself: %q", e.String())
	}
}

// The payload's cwd is the fallback when the environment carries no root — the class
// internal/hookenv exists to close.
func TestFallsBackToPayloadCWD(t *testing.T) {
	dir := t.TempDir()
	in := bash("y")
	in.CWD = dir
	for i := 0; i < strikes.Threshold; i++ {
		fire(t, "", in, noon.Add(time.Duration(i)*time.Second))
	}
	if _, err := os.Stat(strikes.Path(dir)); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	// Tripping resets the key, so the file exists but the key is gone; the proof the
	// payload cwd was used is that the file landed under dir at all.
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err != nil {
		t.Errorf("state was not written under the payload's cwd: %v", err)
	}
}

// Malformed or empty input must never disturb the session.
func TestDegradesQuietly(t *testing.T) {
	dir := t.TempDir()
	for name, stdin := range map[string]string{
		"empty":        "",
		"not json":     "{{{",
		"no tool name": `{"error":"boom"}`,
	} {
		t.Run(name, func(t *testing.T) {
			var o, e bytes.Buffer
			if code := run(nil, strings.NewReader(stdin), &o, &e, dir, noon); code != 0 {
				t.Fatalf("exit %d", code)
			}
			if o.String() != "" {
				t.Errorf("spoke on malformed input: %q", o.String())
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	var o, e bytes.Buffer
	if code := run([]string{"-version"}, strings.NewReader(""), &o, &e, t.TempDir(), noon); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(o.String(), "sc-strike-counter") {
		t.Errorf("version output = %q", o.String())
	}
}

func TestExcerptBoundsTheError(t *testing.T) {
	if got := excerpt("\n\n  first meaningful line  \nsecond"); got != "first meaningful line" {
		t.Errorf("excerpt = %q", got)
	}
	long := strings.Repeat("x", maxErrExcerpt*2)
	if got := excerpt(long); len(got) > maxErrExcerpt+3 {
		t.Errorf("excerpt not bounded: %d chars", len(got))
	}
	if got := excerpt("   \n  \n"); got != "" {
		t.Errorf("a blank error must yield nothing, got %q", got)
	}
}
