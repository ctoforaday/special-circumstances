package recallindex

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

// spyUpdate stands in for `qmd update` and counts how often the hook reached for it.
// No test in this package may run the real thing: it mutates the developer's index.
func spyUpdate(calls *int, result string) func() string {
	return func() string {
		*calls++
		return result
	}
}

type callResult struct {
	stdout, stderr string
	code           int
	updates        int
}

func call(t *testing.T, stdin, projectDir string, qmd bool, now time.Time, updateResult string, args ...string) callResult {
	t.Helper()
	var o, e bytes.Buffer
	calls := 0
	code := run(args, strings.NewReader(stdin), &o, &e, projectDir, qmd, now, spyUpdate(&calls, updateResult))
	return callResult{o.String(), e.String(), code, calls}
}

func writePayload(file string) string {
	return `{"tool_name":"Write","tool_input":{"file_path":"` + file + `"}}`
}

// The stated contract: index only markdown, only when qmd exists, never fail the
// tool call, and never nag on stdout/stderr about an optional capability.
func TestRunIndexesOnlyMarkdownAndStaysSilent(t *testing.T) {
	cases := []struct {
		name        string
		qmd         bool
		stdin       string
		wantUpdates int
		wantInLog   []string
	}{
		{"markdown write indexes", true, writePayload("blue/report.md"), 1, []string{"qmd update ok", "blue/report.md"}},
		{"uppercase extension indexes", true, writePayload("NOTES.MD"), 1, []string{"qmd update ok"}},
		{"code write never pays the index cost", true, writePayload("main.go"), 0, []string{"not markdown", "main.go"}},
		{"qmd absent never runs anything", false, writePayload("report.md"), 0, []string{"doctor --fix"}},
		{"malformed payload never indexes", true, `{"tool_name":`, 0, []string{"unknown", "not markdown"}},
		{"empty stdin never indexes", true, ``, 0, []string{"unknown"}},
		{"extensionless file never indexes", true, writePayload("Makefile"), 0, []string{"not markdown"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			got := call(t, c.stdin, dir, c.qmd, time.Now(), "qmd update ok")
			if got.code != 0 {
				t.Fatalf("exit %d; the hook must never fail the Write/Edit", got.code)
			}
			if got.updates != c.wantUpdates {
				t.Fatalf("ran the index %d time(s); want %d", got.updates, c.wantUpdates)
			}
			if got.stdout != "" || got.stderr != "" {
				t.Errorf("an optional capability must not nag: stdout=%q stderr=%q", got.stdout, got.stderr)
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

// The debounce is the reason this hook stopped costing 14 agent-minutes a run. It is
// carried by a stamp FILE precisely so it survives across hook processes, so the
// second call here is a genuinely separate invocation.
func TestDebounceSuppressesTheSecondUpdate(t *testing.T) {
	dir := t.TempDir()
	base := time.Now()

	if got := call(t, writePayload("a.md"), dir, true, base, "qmd update ok"); got.updates != 1 {
		t.Fatalf("first markdown write must update, ran %d", got.updates)
	}
	stamp := filepath.Join(dir, ".claude", stampName)
	if _, err := os.Stat(stamp); err != nil {
		t.Fatalf("a completed update must leave a stamp: %v", err)
	}

	got := call(t, writePayload("b.md"), dir, true, base.Add(time.Second), "qmd update ok")
	if got.updates != 0 {
		t.Fatalf("a write inside the window must be debounced, ran %d", got.updates)
	}
	if !strings.Contains(readLog(t, dir), "debounced") {
		t.Error("the debounce must be visible in the hook log")
	}
}

// Past the window the index has to refresh again, or staleness is unbounded.
func TestUpdateResumesAfterTheWindow(t *testing.T) {
	dir := t.TempDir()
	if got := call(t, writePayload("a.md"), dir, true, time.Now(), "qmd update ok"); got.updates != 1 {
		t.Fatalf("first write must update")
	}
	// Age the stamp past the interval rather than sleeping for it.
	old := time.Now().Add(-2 * debounceInterval)
	stamp := filepath.Join(dir, ".claude", stampName)
	if err := os.Chtimes(stamp, old, old); err != nil {
		t.Fatal(err)
	}
	if got := call(t, writePayload("b.md"), dir, true, time.Now(), "qmd update ok"); got.updates != 1 {
		t.Fatalf("a write past the window must update, ran %d", got.updates)
	}
}

// A failing `qmd update` is reported in the log, not escalated to the tool call.
func TestUpdateFailureIsLoggedNotFatal(t *testing.T) {
	dir := t.TempDir()
	got := call(t, writePayload("a.md"), dir, true, time.Now(), "qmd update failed: exit status 1")
	if got.code != 0 {
		t.Fatalf("exit %d; an index failure must not fail the Write", got.code)
	}
	if !strings.Contains(readLog(t, dir), "qmd update failed") {
		t.Error("the failure must be recorded for the operator")
	}
}

// With no project dir there is nowhere to keep state, and the hook must not invent
// one in the working directory — neither the stamp nor the log.
func TestNoProjectDirWritesNothingRelative(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	got := call(t, writePayload("a.md"), "", true, time.Now(), "qmd update ok")
	if got.code != 0 {
		t.Fatalf("exit %d", got.code)
	}
	if got.updates != 1 {
		t.Fatalf("without a stamp the index must still refresh, ran %d", got.updates)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); !os.IsNotExist(err) {
		t.Fatal("the hook wrote a relative .claude/ into the working directory")
	}
}

func TestReadStampOnMissingDirIsZero(t *testing.T) {
	if !readStamp(filepath.Join(t.TempDir(), "nope")).IsZero() {
		t.Fatal("a missing stamp must read as the zero time (never indexed)")
	}
	if !readStamp("").IsZero() {
		t.Fatal("an empty dir must read as the zero time, not resolve relatively")
	}
}

func TestVersionFlagSkipsTheHook(t *testing.T) {
	dir := t.TempDir()
	got := call(t, "", dir, true, time.Now(), "qmd update ok", "-version")
	if got.code != 0 {
		t.Fatalf("exit %d", got.code)
	}
	if !strings.Contains(got.stdout, "sc-recall-index") || !strings.Contains(got.stdout, version) {
		t.Fatalf("version output = %q", got.stdout)
	}
	if got.updates != 0 {
		t.Error("-version must not touch the index")
	}
	if _, err := os.Stat(logPath(dir)); !os.IsNotExist(err) {
		t.Error("-version must not record a hook firing")
	}
}

func TestUnknownFlagNeverBlocks(t *testing.T) {
	got := call(t, "", t.TempDir(), true, time.Now(), "qmd update ok", "-nope")
	if got.code != 0 {
		t.Fatalf("exit %d on an unknown flag", got.code)
	}
	if got.updates != 0 {
		t.Error("a bad flag must not fall through into an index run")
	}
}
