package doctor

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// These tests drive hooks/fetch-bin.sh the way a guard does — `sh fetch-bin.sh hook <Event>` with
// CLAUDE_PLUGIN_ROOT set — against a file:// release laid out with AssetName, the name the
// release job publishes and fetchRelease asks for. The script is the second fetcher of the same
// assets; this is what keeps the two on one naming contract.

const fixtureVersion = "1.2.3"

type fetchFixture struct {
	root, rel string
	bins      []string
}

// newFetchFixture builds a plugin root and its release. corrupt gives an asset a wrong digest in
// SHA256SUMS; omit leaves an asset out of SHA256SUMS.
func newFetchFixture(t *testing.T, bins []string, corrupt, omit map[string]bool) fetchFixture {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fetch-bin.sh runs under Git Bash on Windows; the fetch-bin-platforms CI job drives it there")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "..", "hooks", "fetch-bin.sh"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(script)
	if err != nil {
		t.Fatalf("the generated hooks/fetch-bin.sh is missing — run `go -C scripts run ./fetchbingen`: %v", err)
	}
	root := t.TempDir()
	// author.name comes first and deeper: the script must read the top-level name, not this one.
	mustWrite(t, filepath.Join(root, ".claude-plugin", "plugin.json"),
		"{\n  \"author\": {\n    \"name\": \"not-the-plugin\"\n  },\n  \"name\": \"fixture\",\n  \"version\": \""+fixtureVersion+"\"\n}\n")
	mustWrite(t, filepath.Join(root, "hooks", "fetch-bin.sh"), string(body))
	mustWrite(t, filepath.Join(root, "bin", ".gitkeep"), "")
	for _, b := range bins {
		if err := os.MkdirAll(filepath.Join(root, "tools", "cmd", b), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	rel := t.TempDir()
	dir := filepath.Join(rel, "fixture--v"+fixtureVersion)
	var sums strings.Builder
	for _, b := range bins {
		asset := AssetName(b, runtime.GOOS, runtime.GOARCH)
		content := "binary " + b + "\n"
		mustWrite(t, filepath.Join(dir, asset), content)
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		if corrupt[b] {
			digest = strings.Repeat("0", 64)
		}
		if !omit[b] {
			fmt.Fprintf(&sums, "%s  %s\n", digest, asset)
		}
	}
	mustWrite(t, filepath.Join(dir, "SHA256SUMS"), sums.String())
	return fetchFixture{root: root, rel: rel, bins: bins}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// run invokes the fixture's script. path, when non-empty, replaces PATH.
func (f fetchFixture) run(t *testing.T, path string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command("sh", append([]string{filepath.Join(f.root, "hooks", "fetch-bin.sh")}, args...)...)
	cmd.Env = append(os.Environ(), "CLAUDE_PLUGIN_ROOT="+f.root, "SC_RELEASE_BASE_URL=file://"+f.rel, "SC_FETCH_LOCKED=")
	if path != "" {
		cmd.Env = append(cmd.Env, "PATH="+path)
	}
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		code = exit.ExitCode()
	case err != nil:
		t.Fatal(err)
	}
	return o.String(), e.String(), code
}

// hook runs `hook <event>` and holds it to the guard's contract: exit 0, and stdout empty or one
// JSON object. It returns that object's systemMessage ("" when nothing was said).
func (f fetchFixture) hook(t *testing.T, event string) (message string, doc map[string]any) {
	t.Helper()
	out, errOut, code := f.run(t, "", "hook", event)
	if code != 0 {
		t.Fatalf("hook %s exited %d (stderr %q) — a guard must never fail a tool call", event, code, errOut)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", nil
	}
	if strings.Contains(out, "\n") {
		t.Fatalf("hook %s printed more than one line: %q", event, out)
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("hook %s printed invalid JSON %q: %v", event, out, err)
	}
	m, _ := doc["systemMessage"].(string)
	if m == "" {
		t.Fatalf("hook %s printed JSON without a top-level systemMessage: %q", event, out)
	}
	return m, doc
}

func (f fetchFixture) state(name string) string { return filepath.Join(f.root, ".fetch", name) }

func (f fetchFixture) holdLock(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(f.state("lock"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func age(t *testing.T, path string, d time.Duration) {
	t.Helper()
	old := time.Now().Add(-d)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
}

func (f fetchFixture) installed(name string) bool {
	b, err := os.ReadFile(filepath.Join(f.root, "bin", name))
	return err == nil && string(b) == "binary "+name+"\n"
}

func (f fetchFixture) waitInstalled(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		all := true
		for _, b := range f.bins {
			all = all && f.installed(b)
		}
		if all {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	log, _ := os.ReadFile(f.state("log"))
	t.Fatalf("binaries never arrived; .fetch/log:\n%s", log)
}

func (f fetchFixture) failedCause(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(f.state("failed"))
	if err != nil {
		t.Fatalf(".fetch/failed was not written: %v", err)
	}
	return string(b)
}

func (f fetchFixture) binEntries(t *testing.T) []string {
	t.Helper()
	es, err := os.ReadDir(filepath.Join(f.root, "bin"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range es {
		names = append(names, e.Name())
	}
	return names
}

// toolPath builds a PATH holding the tools the script uses, less drop, plus any stubs.
func toolPath(t *testing.T, drop []string, stubs map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	skip := map[string]bool{}
	for _, d := range drop {
		skip[d] = true
	}
	for _, tool := range []string{"sh", "sed", "head", "tr", "find", "mkdir", "rm", "mv", "chmod", "awk", "basename", "date", "uname", "sha256sum", "shasum", "curl", "nohup", "cat"} {
		if skip[tool] || stubs[tool] != "" {
			continue
		}
		if p, err := exec.LookPath(tool); err == nil {
			if err := os.Symlink(p, filepath.Join(dir, tool)); err != nil {
				t.Fatal(err)
			}
		}
	}
	for name, body := range stubs {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFetchBinInstallsEveryMissingBinaryFromThePinnedRelease(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha", "beta"}, nil, nil)
	msg, _ := f.hook(t, "PreToolUse")
	if !strings.Contains(msg, "fixture "+fixtureVersion) || !strings.Contains(msg, "installing") {
		t.Fatalf("PreToolUse should announce the install by name and version, said %q", msg)
	}
	// The hook has already returned; the fetch it started must finish without it.
	f.waitInstalled(t)
	if _, err := os.Stat(f.state("failed")); err == nil {
		t.Fatal(".fetch/failed exists after a successful fetch")
	}
	if got := f.binEntries(t); len(got) != 3 {
		t.Fatalf("bin/ should hold .gitkeep and the two binaries only, holds %v", got)
	}
}

func TestFetchBinInstallsNothingWhenOneDigestIsWrong(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha", "beta"}, map[string]bool{"beta": true}, nil)
	if _, _, code := f.run(t, "", "fetch"); code == 0 {
		t.Fatal("fetch succeeded with a wrong digest")
	}
	if f.installed("alpha") || f.installed("beta") {
		t.Fatalf("bin/ gained a binary from a failed fetch: %v", f.binEntries(t))
	}
	if c := f.failedCause(t); !strings.Contains(c, "checksum mismatch for "+AssetName("beta", runtime.GOOS, runtime.GOARCH)) {
		t.Fatalf("cause should name the bad asset, is %q", c)
	}
}

func TestFetchBinNamesAnAssetTheReleaseDoesNotCarry(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha", "beta"}, nil, map[string]bool{"beta": true})
	if _, _, code := f.run(t, "", "fetch"); code == 0 {
		t.Fatal("fetch succeeded with an asset missing from SHA256SUMS")
	}
	if c := f.failedCause(t); !strings.Contains(c, "has no "+AssetName("beta", runtime.GOOS, runtime.GOARCH)) {
		t.Fatalf("cause should name the missing asset, is %q", c)
	}
	if f.installed("alpha") {
		t.Fatal("alpha was installed from a release missing its sibling")
	}
}

func TestFetchBinNamesAMissingPrerequisite(t *testing.T) {
	for _, c := range []struct {
		name string
		drop []string
		want string
	}{
		{"curl", []string{"curl"}, "curl not found"},
		{"checksum tools", []string{"sha256sum", "shasum"}, "neither sha256sum nor shasum found"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFetchFixture(t, []string{"alpha"}, nil, nil)
			if _, _, code := f.run(t, toolPath(t, c.drop, nil), "fetch"); code == 0 {
				t.Fatal("fetch succeeded without its prerequisite")
			}
			if got := f.failedCause(t); !strings.Contains(got, c.want) {
				t.Fatalf("cause should be %q, is %q", c.want, got)
			}
		})
	}
}

func TestFetchBinRefusesAnUnsupportedPlatform(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	path := toolPath(t, nil, map[string]string{"uname": "echo Plan9"})
	if _, _, code := f.run(t, path, "fetch"); code == 0 {
		t.Fatal("fetch succeeded on an unsupported platform")
	}
	if c := f.failedCause(t); !strings.Contains(c, "unsupported platform Plan9") {
		t.Fatalf("cause should name the platform, is %q", c)
	}
}

// Speaking is decided per event: only the events whose systemMessage Claude Code displays may
// speak, and only they spend the 10-minute budget.
func TestFetchBinSpeaksOnlyOnDisplayingEvents(t *testing.T) {
	displaying := map[string]bool{"SessionStart": true, "PreToolUse": true, "PostToolUse": true, "PostToolUseFailure": true, "Stop": true}
	for _, ev := range []string{"SessionStart", "PreToolUse", "PostToolUse", "PostToolUseFailure", "Stop", "SubagentStart", "SubagentStop", "SessionEnd", "PreCompact", "PostCompact", "FileChanged"} {
		t.Run(ev, func(t *testing.T) {
			f := newFetchFixture(t, []string{"alpha"}, nil, nil)
			f.holdLock(t) // a fetch is already running, so this invocation starts none
			msg, doc := f.hook(t, ev)
			_, notified := os.Stat(f.state("notified.installing"))
			if !displaying[ev] {
				if msg != "" || notified == nil {
					t.Fatalf("%s does not display, yet said %q (notified touched: %v)", ev, msg, notified == nil)
				}
				return
			}
			if msg == "" || notified != nil {
				t.Fatalf("%s displays, yet said nothing (notified: %v)", ev, notified)
			}
			if ev == "SessionStart" {
				hso, _ := doc["hookSpecificOutput"].(map[string]any)
				if ctx, _ := hso["additionalContext"].(string); ctx != msg {
					t.Fatalf("SessionStart should also tell the agent, additionalContext = %q", ctx)
				}
			}
		})
	}
}

// A fetch started by a non-displaying event is announced by the first displaying event after it,
// and the budget then holds for 10 minutes.
func TestFetchBinAnnouncesAFetchSomeoneElseStarted(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	f.holdLock(t)
	if msg, _ := f.hook(t, "SubagentStart"); msg != "" {
		t.Fatalf("SubagentStart spoke: %q", msg)
	}
	if msg, _ := f.hook(t, "PreToolUse"); !strings.Contains(msg, "installing") {
		t.Fatalf("PreToolUse after a silent start should announce the install, said %q", msg)
	}
	if msg, _ := f.hook(t, "PostToolUse"); msg != "" {
		t.Fatalf("a second displaying event inside 10 minutes spoke: %q", msg)
	}
	age(t, f.state("notified.installing"), 11*time.Minute)
	if msg, _ := f.hook(t, "Stop"); msg == "" {
		t.Fatal("after 10 minutes the next displaying event should speak again")
	}
}

func TestFetchBinReportsARecentFailureOnTheNextDisplayingEvent(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	mustWrite(t, f.state("failed"), "boom\n")
	if msg, _ := f.hook(t, "SubagentStop"); msg != "" {
		t.Fatalf("SubagentStop spoke: %q", msg)
	}
	msg, _ := f.hook(t, "Stop")
	if !strings.Contains(msg, "boom") || !strings.Contains(msg, "doctor --fix") {
		t.Fatalf("Stop should report the cause and the manual command, said %q", msg)
	}
	if _, err := os.Stat(f.state("log")); err == nil {
		t.Fatal("a failure younger than 5 minutes was retried")
	}
}

func TestFetchBinRetriesAnOldFailure(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	mustWrite(t, f.state("failed"), "boom\n")
	age(t, f.state("failed"), 6*time.Minute)
	f.hook(t, "PreToolUse")
	f.waitInstalled(t)
}

func TestFetchBinBreaksAStaleLockAndLeavesAFreshOne(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	f.holdLock(t)
	f.hook(t, "PreToolUse")
	if _, err := os.Stat(f.state("log")); err == nil {
		t.Fatal("a fetch started while a fresh lock was held")
	}
	age(t, f.state("lock"), 11*time.Minute)
	f.hook(t, "PreToolUse")
	f.waitInstalled(t)
}

func TestFetchBinWithoutAVersionPointsAtTheManualPath(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	mustWrite(t, filepath.Join(f.root, ".claude-plugin", "plugin.json"), "{\n  \"name\": \"fixture\"\n}\n")
	_, errOut, code := f.run(t, "", "hook", "SubagentStop")
	if code != 0 || !strings.Contains(errOut, "doctor --fix") {
		t.Fatalf("exit %d, stderr %q: a missing version must still exit 0 and name the manual path", code, errOut)
	}
	if msg, _ := f.hook(t, "SubagentStop"); msg != "" {
		t.Fatalf("a non-displaying event spoke: %q", msg)
	}
	if msg, _ := f.hook(t, "PreToolUse"); !strings.Contains(msg, "doctor --fix") {
		t.Fatalf("PreToolUse should show the manual path, said %q", msg)
	}
	// The same message is throttled like every other: not repeated on the next tool call.
	if msg, _ := f.hook(t, "PreToolUse"); msg != "" {
		t.Fatalf("the missing-version message repeated inside 10 minutes: %q", msg)
	}
}

// Kinds are throttled apart: a failure that follows an "installing" message is reported at the
// next displaying event, not 10 minutes later.
func TestFetchBinReportsAFailureRightAfterAnnouncingTheInstall(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	f.holdLock(t)
	if msg, _ := f.hook(t, "PreToolUse"); !strings.Contains(msg, "installing") {
		t.Fatalf("expected the install announcement, got %q", msg)
	}
	if err := os.RemoveAll(f.state("lock")); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, f.state("failed"), "boom\n")
	if msg, _ := f.hook(t, "Stop"); !strings.Contains(msg, "boom") {
		t.Fatalf("the failure was held back behind the install message: %q", msg)
	}
}

// holdLockBy leaves a lock that names pid as its holder, aged past the 10-minute mark.
func (f fetchFixture) holdLockBy(t *testing.T, pid int) {
	t.Helper()
	f.holdLock(t)
	mustWrite(t, f.state("lock/pid"), fmt.Sprintf("%d\n", pid))
	age(t, f.state("lock"), 11*time.Minute)
}

// A slow fetch is still running: however old its lock, a second fetch must not start beside it.
func TestFetchBinNeverBreaksALockWhoseHolderIsRunning(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	f.holdLockBy(t, os.Getpid()) // this test process: certainly alive
	f.hook(t, "PreToolUse")
	if _, err := os.Stat(f.state("log")); err == nil {
		t.Fatal("a second fetch started beside a live holder")
	}
	if _, _, code := f.run(t, "", "fetch"); code == 0 {
		t.Fatal("a direct fetch ran while a live holder had the lock")
	}
	if _, err := os.Stat(f.state("lock/pid")); err != nil {
		t.Fatal("a fetch that never held the lock removed it")
	}
}

// A fetch removes only a lock it still owns. The curl stub stands in for a takeover mid-fetch: it
// records another holder in the lock, then fails the download.
func TestFetchBinLeavesALockItNoLongerOwns(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	path := toolPath(t, nil, map[string]string{"curl": `echo 999999 > "$CLAUDE_PLUGIN_ROOT/.fetch/lock/pid"; exit 22`})
	if _, _, code := f.run(t, path, "fetch"); code == 0 {
		t.Fatal("fetch succeeded with a failing curl")
	}
	if b, err := os.ReadFile(f.state("lock/pid")); err != nil || strings.TrimSpace(string(b)) != "999999" {
		t.Fatalf("the new holder's lock was removed or rewritten on exit: %q, %v", b, err)
	}
}

// A holder that has exited leaves a lock that is broken at once, whatever its age.
func TestFetchBinBreaksALockWhoseHolderHasExited(t *testing.T) {
	f := newFetchFixture(t, []string{"alpha"}, nil, nil)
	gone := exec.Command("sh", "-c", "exit 0")
	if err := gone.Run(); err != nil {
		t.Fatal(err)
	}
	f.holdLock(t)
	mustWrite(t, f.state("lock/pid"), fmt.Sprintf("%d\n", gone.Process.Pid))
	f.hook(t, "PreToolUse")
	f.waitInstalled(t)
}
