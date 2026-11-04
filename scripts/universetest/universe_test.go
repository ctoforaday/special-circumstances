//go:build linux

// Package universetest drives scripts/universe.sh against a stub `claude` on PATH: a script that
// sleeps and emits stream-json events. It covers what the launcher does with the process it starts.
// It says nothing about the engine; `universe.sh run` itself is how that is exercised.
package universetest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	leadSession = "lead-session-1"
	runID       = "wf_stub-run"
	scriptPath  = "/stub/debate.js"
	topic       = "Is 91 prime?"
	waitLimit   = 30 * time.Second
)

// The stub stands for one `claude -p`. A fresh invocation dispatches the Workflow; a --resume
// invocation first emits the empty result a queued notification produces, as m37's resume did, and
// only then dispatches. It holds until the test creates $STUB_DIR/release, with a child that outlives
// it, and writes each fact the test asks about to a file named for the invocation.
const stubClaude = `#!/usr/bin/env bash
n=$(ls "$STUB_DIR" | grep -c '^argv\.' )
n=$((n + 1))
printf '%s\n' "$@" > "$STUB_DIR/argv.$n"
printf '%s\n' "$$" > "$STUB_DIR/pid.$n"
printf '%s\n' "$(ps -o sid= -p $$ | tr -d ' ')" > "$STUB_DIR/sid.$n"
readlink /proc/$$/fd/0 > "$STUB_DIR/stdin.$n"
readlink /proc/$$/fd/1 > "$STUB_DIR/stdout.$n"
pwd > "$STUB_DIR/cwd.$n"
printf '%s\n' "$CLAUDE_CONFIG_DIR" > "$STUB_DIR/config.$n"
printf '%s\n' "$CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS" > "$STUB_DIR/ceiling.$n"
sleep 600 &
printf '%s\n' "$!" > "$STUB_DIR/child.$n"

emit() { printf '%s\n' "$1"; }
workflow='{"type":"assistant","session_id":"` + leadSession + `","message":{"content":[{"type":"tool_use","id":"toolu_1","name":"Workflow","input":{"scriptPath":"` + scriptPath + `","args":{"topic":"` + topic + `","runDir":"'"$STUB_RUNDIR"'","lanes":1}}}]}}'
result='{"type":"user","session_id":"` + leadSession + `","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"launched"}]},"tool_use_result":{"status":"async_launched","runId":"` + runID + `"}}'

emit '{"type":"system","subtype":"init","session_id":"` + leadSession + `"}'
emit '{"type":"user","session_id":"` + leadSession + `","message":{"content":"plain text content"}}'
emit '3'
case " $* " in
  *" --resume "*)
    emit '{"type":"result","subtype":"success","num_turns":0,"total_cost_usd":13.06,"session_id":"` + leadSession + `"}'
    sleep 0.2
    ;;
esac
if [ -z "${STUB_NO_WORKFLOW:-}" ]; then
  emit "$workflow"
  emit "$result"
fi
until [ -e "$STUB_DIR/release" ]; do sleep 0.05; done
emit '{"type":"result","subtype":"success","num_turns":7,"total_cost_usd":1.5,"session_id":"` + leadSession + `"}'
emit '{"type":"system","subtype":"stub_last_event"}'
exit "${STUB_RC:-0}"
`

type universe struct {
	t        *testing.T
	script   string
	workdir  string
	stubDir  string
	runDir   string
	baseEnv  []string
	scratch  string
	extraEnv []string
}

func newUniverse(t *testing.T) *universe {
	t.Helper()
	for _, tool := range []string{"setsid", "python3", "pkill", "bash"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH", tool)
		}
	}
	script, err := filepath.Abs("../universe.sh")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	u := &universe{
		t:       t,
		script:  script,
		scratch: scratch,
		workdir: filepath.Join(scratch, "wd"),
		stubDir: filepath.Join(scratch, "stub"),
	}
	u.runDir = filepath.Join(u.workdir, "src", "research", "2026_stub")
	bin := filepath.Join(scratch, "bin")
	for _, d := range []string{bin, u.stubDir, u.workdir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(stubClaude), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "FEOV_") || name == "HOME" || name == "PATH" || name == "WORKDIR" || name == "CLAUDE_CONFIG_DIR" {
			continue
		}
		u.baseEnv = append(u.baseEnv, kv)
	}
	u.baseEnv = append(u.baseEnv,
		"HOME="+scratch,
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"WORKDIR="+u.workdir,
		"STUB_DIR="+u.stubDir,
		"STUB_RUNDIR="+u.runDir,
	)
	// Nothing the stub left may outlive the test: a sleeping child would keep its session alive and
	// hold the test binary's temp directory.
	t.Cleanup(u.killEverything)
	return u
}

func (u *universe) killEverything() {
	for _, name := range []string{"pid", "child"} {
		for n := 1; n < 5; n++ {
			if pid, ok := u.readPid(fmt.Sprintf("%s.%d", name, n)); ok {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
	}
	if b, err := os.ReadFile(filepath.Join(u.workdir, "run.pid")); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
	}
}

func (u *universe) readPid(name string) (int, bool) {
	b, err := os.ReadFile(filepath.Join(u.stubDir, name))
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	return pid, err == nil
}

func (u *universe) command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "bash", append([]string{u.script}, args...)...)
	cmd.Env = append(append([]string{}, u.baseEnv...), u.extraEnv...)
	// A child that inherited the pipe would hold Output open until the deadline, so a detach that
	// shared stdout with its caller fails here as a timeout.
	cmd.WaitDelay = time.Second
	return cmd
}

// verb runs one verb to its end and returns what it printed and its exit status.
func (u *universe) verb(args ...string) (string, int) {
	u.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), waitLimit)
	defer cancel()
	out, err := u.command(ctx, args...).CombinedOutput()
	if ctx.Err() != nil || errors.Is(err, exec.ErrWaitDelay) {
		u.t.Fatalf("%v did not return: %v\n%s", args, err, out)
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), exit.ExitCode()
	}
	if err != nil {
		u.t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return string(out), 0
}

func (u *universe) eventually(what string, ok func() bool) {
	u.t.Helper()
	deadline := time.Now().Add(waitLimit)
	for !ok() {
		if time.Now().After(deadline) {
			u.t.Fatalf("never happened: %s\nlog:\n%s", what, u.log())
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (u *universe) stubStarted(n int) int {
	u.t.Helper()
	var pid int
	u.eventually(fmt.Sprintf("stub invocation %d started", n), func() bool {
		if _, ok := u.readPid(fmt.Sprintf("child.%d", n)); !ok {
			return false
		}
		var ok bool
		pid, ok = u.readPid(fmt.Sprintf("pid.%d", n))
		return ok
	})
	return pid
}

func (u *universe) log() string {
	b, _ := os.ReadFile(filepath.Join(u.workdir, "run-current.log"))
	return string(b)
}

// exitLines returns the log's lines that are exactly an EXIT line, an unparsed count being what a
// test wants: one run, one line.
func (u *universe) exitLines() []string {
	var lines []string
	for _, l := range strings.Split(u.log(), "\n") {
		if strings.HasPrefix(l, "EXIT ") {
			lines = append(lines, l)
		}
	}
	return lines
}

func (u *universe) waitExit(want string) {
	u.t.Helper()
	u.eventually("an EXIT line", func() bool { return len(u.exitLines()) > 0 })
	if got := u.exitLines(); len(got) != 1 || got[0] != want {
		u.t.Fatalf("EXIT lines = %q, want exactly %q\nlog:\n%s", got, want, u.log())
	}
	u.eventually("the session to end", func() bool {
		b, err := os.ReadFile(filepath.Join(u.workdir, "run.pid"))
		if err != nil {
			return false
		}
		out, _ := exec.Command("ps", "-o", "stat=", "-s", strings.TrimSpace(string(b))).Output()
		for _, stat := range strings.Fields(string(out)) {
			if !strings.HasPrefix(stat, "Z") {
				return false
			}
		}
		return true
	})
}

func (u *universe) release() {
	u.t.Helper()
	if err := os.WriteFile(filepath.Join(u.stubDir, "release"), nil, 0o644); err != nil {
		u.t.Fatal(err)
	}
}

func alive(pid int) bool {
	if syscall.Kill(pid, 0) != nil {
		return false
	}
	// A killed process whose parent has not reaped it still answers signal 0.
	out, _ := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(pid)).Output()
	return len(out) > 0 && !strings.HasPrefix(strings.TrimSpace(string(out)), "Z")
}

// record writes the part of a run's record the launcher reads: the outcome table, with the given
// verdict row when there is one.
func (u *universe) record(verdict string) {
	u.t.Helper()
	records := filepath.Join(u.runDir, "records")
	if err := os.MkdirAll(records, 0o755); err != nil {
		u.t.Fatal(err)
	}
	const py = `
import sqlite3, sys
c = sqlite3.connect(sys.argv[1])
c.execute('CREATE TABLE IF NOT EXISTS outcome (verdict TEXT)')
if sys.argv[2]:
    c.execute('INSERT INTO outcome VALUES (?)', (sys.argv[2],))
c.commit()
`
	if out, err := exec.Command("python3", "-c", py, filepath.Join(records, "record.db"), verdict).CombinedOutput(); err != nil {
		u.t.Fatalf("record fixture: %v\n%s", err, out)
	}
}

func (u *universe) readStub(name string) string {
	u.t.Helper()
	b, err := os.ReadFile(filepath.Join(u.stubDir, name))
	if err != nil {
		u.t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

func (u *universe) invocations() int {
	u.t.Helper()
	m, err := filepath.Glob(filepath.Join(u.stubDir, "argv.*"))
	if err != nil {
		u.t.Fatal(err)
	}
	return len(m)
}

// startDetached launches a run with --no-follow and returns the stub's pid, having proved `run`
// returned while the stub was still going.
func (u *universe) startDetached() int {
	u.t.Helper()
	u.record("")
	out, rc := u.verb("run", "--no-follow", topic)
	if rc != 0 {
		u.t.Fatalf("run --no-follow exit %d:\n%s", rc, out)
	}
	return u.stubStarted(1)
}

// truncate leaves a run the way m37's launcher left it: claude killed with no outcome on the record.
func (u *universe) truncate() {
	u.t.Helper()
	pid := u.startDetached()
	u.eventually("the Workflow call in the raw stream", func() bool {
		b, _ := os.ReadFile(filepath.Join(u.workdir, "run-origin.jsonl"))
		return bytes.Contains(b, []byte(runID))
	})
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		u.t.Fatal(err)
	}
	u.waitExit("EXIT 137")
}

func TestKillingTheFollowerLeavesTheRunRunning(t *testing.T) {
	u := newUniverse(t)
	u.record("")

	ctx, cancel := context.WithTimeout(context.Background(), waitLimit)
	defer cancel()
	follower := u.command(ctx, "run", topic)
	follower.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var seen bytes.Buffer
	follower.Stdout, follower.Stderr = &seen, &seen
	if err := follower.Start(); err != nil {
		t.Fatal(err)
	}
	stub := u.stubStarted(1)

	// The whole process group, which is what a launcher's ceiling kills.
	if err := syscall.Kill(-follower.Process.Pid, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	_ = follower.Wait()

	time.Sleep(300 * time.Millisecond)
	if !alive(stub) {
		t.Fatalf("killing the follower killed the run\nlog:\n%s", u.log())
	}
	if lines := u.exitLines(); len(lines) != 0 {
		t.Fatalf("an EXIT line while the run is alive: %q", lines)
	}

	// It shares neither a session nor a standard stream with what launched it.
	pidFile, err := os.ReadFile(filepath.Join(u.workdir, "run.pid"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := u.readStub("sid.1"), strings.TrimSpace(string(pidFile)); got != want {
		t.Errorf("the stub's session %s is not the one in run.pid %s", got, want)
	}
	ours, err := exec.Command("ps", "-o", "sid=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		t.Fatal(err)
	}
	if u.readStub("sid.1") == strings.TrimSpace(string(ours)) {
		t.Errorf("the stub is in its launcher's session %s", ours)
	}
	if got := u.readStub("stdin.1"); got != "/dev/null" {
		t.Errorf("the stub's stdin is %q", got)
	}
	if got := u.readStub("stdout.1"); !strings.HasPrefix(got, filepath.Join(u.workdir, "stream-")) {
		t.Errorf("the stub's stdout is %q, not the raw stream file", got)
	}
	if got := u.readStub("ceiling.1"); got != "0" {
		t.Errorf("CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS = %q", got)
	}
	if got, want := u.readStub("cwd.1"), filepath.Join(u.workdir, "src"); got != want {
		t.Errorf("cwd %q, want %q", got, want)
	}

	// The run goes on to its end, and says so, with no follower.
	u.release()
	u.waitExit("EXIT 0")
	raw, err := os.ReadFile(filepath.Join(u.workdir, "run-origin.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("stub_last_event")) {
		t.Errorf("the raw stream stops short of the last event:\n%s", raw)
	}
	if log := u.log(); !strings.Contains(log, "WORKFLOW DISPATCHED") || strings.Contains(log, "FAILED") {
		t.Errorf("the log:\n%s", log)
	}
}

func TestFollowingReturnsTheRunsExitStatus(t *testing.T) {
	u := newUniverse(t)
	u.record("")
	u.extraEnv = []string{"STUB_RC=7"}
	u.release()
	out, rc := u.verb("run", topic)
	if rc != 7 {
		t.Fatalf("exit %d, want 7:\n%s", rc, out)
	}
	if !strings.Contains(out, "EXIT 7") {
		t.Errorf("the follower did not show the EXIT line:\n%s", out)
	}
	u.waitExit("EXIT 7")
}

func TestARunThatDispatchesNoWorkflowFails(t *testing.T) {
	u := newUniverse(t)
	u.record("")
	u.extraEnv = []string{"STUB_NO_WORKFLOW=1"}
	u.release()
	out, rc := u.verb("run", topic)
	if rc != 2 {
		t.Fatalf("exit %d, want 2:\n%s", rc, out)
	}
	if !strings.Contains(out, "no Workflow was dispatched") {
		t.Errorf("no failure line:\n%s", out)
	}
	u.waitExit("EXIT 2")
}

func TestWatchTellsALiveRunFromATruncatedOne(t *testing.T) {
	u := newUniverse(t)
	pid := u.startDetached()

	// Alive: `watch` loops, so it is cut off after its first pass.
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	live, _ := u.command(ctx, "watch", "1").CombinedOutput()
	if !strings.Contains(string(live), "process: running") || strings.Contains(string(live), "TRUNCATED") {
		t.Fatalf("watch on a live run:\n%s", live)
	}

	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	u.waitExit("EXIT 137")

	// Dead with no outcome: it says so and stops by itself, with the status that means it.
	out, rc := u.verb("watch", "1")
	if rc != 3 || !strings.Contains(out, "TRUNCATED") || !strings.Contains(out, "EXIT 137") {
		t.Fatalf("watch on a killed run, exit %d:\n%s", rc, out)
	}
	if !strings.Contains(out, "resume") {
		t.Errorf("watch does not say how to resume:\n%s", out)
	}

	// What the killed claude left running ended with it.
	child, ok := u.readPid("child.1")
	if !ok {
		t.Fatal("no child pid")
	}
	if alive(child) {
		t.Errorf("the stub's child %d outlived the run", child)
	}
}

func TestWatchOnAFinishedRunIsNotTruncated(t *testing.T) {
	u := newUniverse(t)
	u.startDetached()
	u.record("VERIFIED")
	u.release()
	u.waitExit("EXIT 0")
	out, rc := u.verb("watch", "1")
	if rc != 0 || strings.Contains(out, "TRUNCATED") || !strings.Contains(out, "process: ended") {
		t.Fatalf("watch on a finished run, exit %d:\n%s", rc, out)
	}
}

func TestWatchWithNoPidFileDoesNotGuess(t *testing.T) {
	u := newUniverse(t)
	u.record("")
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	out, _ := u.command(ctx, "watch", "1").CombinedOutput()
	if strings.Contains(string(out), "TRUNCATED") || !strings.Contains(string(out), "nothing says whether a process is alive") {
		t.Fatalf("watch with no pid file:\n%s", out)
	}
}

func TestWatchTakesTheSupervisorsDeathForTruncationOnlyWhenClaudeIsGone(t *testing.T) {
	u := newUniverse(t)
	stub := u.startDetached()
	pid, err := os.ReadFile(filepath.Join(u.workdir, "run.pid"))
	if err != nil {
		t.Fatal(err)
	}
	sup, err := strconv.Atoi(strings.TrimSpace(string(pid)))
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Kill(sup, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	u.eventually("the supervisor to die", func() bool { return !alive(sup) })

	// claude is still writing the record, so a resume must still be refused.
	if !alive(stub) {
		t.Fatal("the stub died with its supervisor")
	}
	out, rc := u.verb("resume")
	if rc != 3 || !strings.Contains(out, "still alive") {
		t.Fatalf("resume beside a live claude, exit %d:\n%s", rc, out)
	}

	// Everything gone and no EXIT line: truncated, and the line's absence is stated.
	u.killEverything()
	u.eventually("the session to end", func() bool { return !alive(stub) })
	child, _ := u.readPid("child.1")
	u.eventually("the child to end", func() bool { return !alive(child) })
	out, rc = u.verb("watch", "1")
	if rc != 3 || !strings.Contains(out, "TRUNCATED") || !strings.Contains(out, "no EXIT line") {
		t.Fatalf("watch, exit %d:\n%s", rc, out)
	}
}

func TestSignallingTheSupervisorEndsTheRunWithItsLine(t *testing.T) {
	u := newUniverse(t)
	stub := u.startDetached()
	pid, err := os.ReadFile(filepath.Join(u.workdir, "run.pid"))
	if err != nil {
		t.Fatal(err)
	}
	sup, err := strconv.Atoi(strings.TrimSpace(string(pid)))
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Kill(sup, syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	u.waitExit("EXIT 143")
	if alive(stub) {
		t.Error("the stub outlived a TERM to its supervisor")
	}
}

func TestResumeRefusesALiveProcess(t *testing.T) {
	u := newUniverse(t)
	pid := u.startDetached()
	out, rc := u.verb("resume")
	if rc != 3 || !strings.Contains(out, "still alive") {
		t.Fatalf("resume on a live run, exit %d:\n%s", rc, out)
	}
	if n := u.invocations(); n != 1 {
		t.Errorf("resume started claude: %d invocations", n)
	}
	if !alive(pid) {
		t.Error("the refused resume hurt the live run")
	}
}

func TestBuildAndRunRefuseALiveProcessEvenWithForce(t *testing.T) {
	u := newUniverse(t)
	u.startDetached()
	for _, args := range [][]string{{"run", topic}, {"run", "--force", topic}, {"build", "--force"}} {
		out, rc := u.verb(args...)
		if rc != 3 || !strings.Contains(out, "process is alive") {
			t.Errorf("%v on a live run, exit %d:\n%s", args, rc, out)
		}
	}
	if n := u.invocations(); n != 1 {
		t.Errorf("%d claude invocations", n)
	}
}

func TestResumeRefusesWithNoPidFile(t *testing.T) {
	u := newUniverse(t)
	out, rc := u.verb("resume")
	if rc != 3 || !strings.Contains(out, "nothing says the process is dead") {
		t.Fatalf("exit %d:\n%s", rc, out)
	}
}

func TestResumeRefusesAnOutcome(t *testing.T) {
	u := newUniverse(t)
	u.truncate()
	u.record("VERIFIED")
	out, rc := u.verb("resume")
	if rc != 3 || !strings.Contains(out, "already has an outcome") || !strings.Contains(out, "VERIFIED") {
		t.Fatalf("exit %d:\n%s", rc, out)
	}
	if n := u.invocations(); n != 1 {
		t.Errorf("resume started claude: %d invocations", n)
	}
}

func TestResumeRefusesAStreamWithNoWorkflow(t *testing.T) {
	u := newUniverse(t)
	u.extraEnv = []string{"STUB_NO_WORKFLOW=1"}
	pid := u.startDetached()
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	u.waitExit("EXIT 137")
	out, rc := u.verb("resume")
	if rc != 3 || !strings.Contains(out, "no Workflow call") {
		t.Fatalf("exit %d:\n%s", rc, out)
	}
}

func TestResumeRefusesARunWithNoRecord(t *testing.T) {
	u := newUniverse(t)
	u.truncate()
	if err := os.RemoveAll(filepath.Join(u.runDir, "records")); err != nil {
		t.Fatal(err)
	}
	out, rc := u.verb("resume")
	if rc != 3 || !strings.Contains(out, "no record at") {
		t.Fatalf("exit %d:\n%s", rc, out)
	}
}

func TestResumeRestartsTheDebateFromTheOriginalCall(t *testing.T) {
	u := newUniverse(t)
	u.truncate()

	out, rc := u.verb("resume", "--no-follow")
	if rc != 0 {
		t.Fatalf("exit %d:\n%s", rc, out)
	}
	stub := u.stubStarted(2)
	u.eventually("the resumed Workflow dispatch", func() bool { return strings.Contains(u.log(), "WORKFLOW DISPATCHED") })

	argv := strings.Split(u.readStub("argv.2"), "\n")
	if len(argv) < 4 || argv[0] != "-p" || argv[1] != "--resume" || argv[2] != leadSession {
		t.Fatalf("argv: %q", argv)
	}
	prompt := argv[3]
	wantArgs, err := json.Marshal(map[string]any{"topic": topic, "runDir": u.runDir, "lanes": 1})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_, argText, found := strings.Cut(prompt, "the original call: ")
	if !found {
		t.Fatalf("no args in the prompt: %s", prompt)
	}
	dec := json.NewDecoder(strings.NewReader(argText))
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("args in the prompt: %v: %s", err, argText)
	}
	gotArgs, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotArgs) != string(wantArgs) {
		t.Errorf("args %s, want %s", gotArgs, wantArgs)
	}
	for _, want := range []string{`resumeFromRunId "` + runID + `"`, `scriptPath "` + scriptPath + `"`, "--from-session " + leadSession} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt lacks %q: %s", want, prompt)
		}
	}
	if got, want := u.readStub("config.2"), filepath.Join(u.workdir, "config"); got != want {
		t.Errorf("CLAUDE_CONFIG_DIR %q, want %q", got, want)
	}
	if got := u.readStub("ceiling.2"); got != "0" {
		t.Errorf("CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS = %q", got)
	}

	// Past the early result the renderer is still there, and the capture goes on to the end.
	if !alive(stub) {
		t.Fatal("the resumed stub ended early")
	}
	logged := u.log()
	if !strings.Contains(logged, "RESULT  success  turns=0") {
		t.Errorf("no early result in the log:\n%s", logged)
	}
	u.release()
	u.waitExit("EXIT 0")
	if log := u.log(); strings.Contains(log, "FAILED") || !strings.Contains(log, "turns=7") {
		t.Errorf("the resumed run's log:\n%s", log)
	}
	stdout := u.readStub("stdout.2")
	raw, err := os.ReadFile(stdout)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("stub_last_event")) {
		t.Errorf("capture stopped short:\n%s", raw)
	}
	if strings.HasSuffix(stdout, mustReadlink(t, filepath.Join(u.workdir, "run-origin.jsonl"))) {
		t.Errorf("the resume wrote the origin stream %s", stdout)
	}
}

func TestAResumedLeadThatDispatchesNothingSaysSo(t *testing.T) {
	u := newUniverse(t)
	u.truncate()
	u.extraEnv = []string{"STUB_NO_WORKFLOW=1"}
	u.release()
	out, rc := u.verb("resume")
	if rc != 2 || !strings.Contains(out, "the debate did not restart") || strings.Contains(out, "plugins did not load") {
		t.Fatalf("exit %d:\n%s", rc, out)
	}
	u.waitExit("EXIT 2")
}

func TestASecondResumeReadsTheSameOriginalCall(t *testing.T) {
	u := newUniverse(t)
	u.truncate()
	if out, rc := u.verb("resume", "--no-follow"); rc != 0 {
		t.Fatalf("first resume, exit %d:\n%s", rc, out)
	}
	stub := u.stubStarted(2)
	if err := syscall.Kill(stub, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	u.waitExit("EXIT 137")
	if out, rc := u.verb("resume", "--no-follow"); rc != 0 {
		t.Fatalf("second resume, exit %d:\n%s", rc, out)
	}
	u.stubStarted(3)
	if got, want := u.readStub("argv.3"), u.readStub("argv.2"); got != want {
		t.Errorf("second resume argv:\n%s\nwant:\n%s", got, want)
	}
}

func mustReadlink(t *testing.T, path string) string {
	t.Helper()
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatal(err)
	}
	return target
}

// The renderer on its own, for what a resumed lead doing nothing looks like: it is a failure, and it
// is not reported as the plugins failing to load, which a resumed session has already proved false.
func TestTheRendererJudgesAResumeByTheWholeStream(t *testing.T) {
	renderer, err := filepath.Abs("../universe-stream.py")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH")
	}
	early := `{"type":"result","subtype":"success","num_turns":0}`
	workflow := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Workflow","input":{}}]}}`
	cases := []struct {
		name, stream string
		args         []string
		rc           int
		want, reject string
	}{
		{"resume, dispatch after an early result", early + "\n" + workflow + "\n", []string{"--resumed"}, 0, "WORKFLOW DISPATCHED", "FAILED"},
		{"fresh, dispatch after an early result", early + "\n" + workflow + "\n", nil, 0, "WORKFLOW DISPATCHED", "FAILED"},
		{"resume, no dispatch", early + "\n", []string{"--resumed"}, 2, "the debate did not restart", "plugins did not load"},
		{"fresh, no dispatch", early + "\n", nil, 2, "plugins did not load", "did not restart"},
		{"text content and a bare value", `{"type":"user","message":{"content":"text"}}` + "\n3\n" + workflow + "\n", nil, 0, "WORKFLOW DISPATCHED", "Traceback"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := exec.Command("python3", append([]string{renderer}, c.args...)...)
			cmd.Stdin = strings.NewReader(c.stream)
			out, err := cmd.CombinedOutput()
			rc := 0
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				rc = exit.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			if rc != c.rc || !strings.Contains(string(out), c.want) || strings.Contains(string(out), c.reject) {
				t.Errorf("exit %d, want %d, want %q, reject %q:\n%s", rc, c.rc, c.want, c.reject, out)
			}
		})
	}
}
