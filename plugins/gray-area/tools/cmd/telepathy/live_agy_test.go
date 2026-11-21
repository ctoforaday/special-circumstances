package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/telecli"
)

// TestLiveAntigravityLinkAndSend drives the shipped hook and both binaries through a real agy:
// built from this checkout, installed with the plugin's own hooks.json into an isolated agy home.
//
//  1. A new conversation, told only to say OK, is asked by the PreInvocation hook to link — on every
//     turn until it does — links, and gets its session row from SessionStart.
//  2. Resumed on a new language server, it is asked again (the old link is stale), relinks, and a
//     telepathy send to itself is delivered through its language server as a system message.
//  3. With that server gone, a send from outside any agent session queues to the mailbox and says
//     why.
//  4. The next turn reads the queued message.
//
// Gated like every live harness test (HOOKCORE_LIVE=1); it spends four to eight agy flash turns.
func TestLiveAntigravityLinkAndSend(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live Antigravity test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live Antigravity test under -short")
	}
	agy, err := exec.LookPath("agy")
	if err != nil {
		t.Skipf("agy not found on PATH: %v", err)
	}
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	token := filepath.Join(realHome, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if _, err := os.Stat(token); err != nil {
		t.Skipf("no agy credentials at %s", token)
	}

	home, work, runtimeDir, stateDir := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	agyDir := filepath.Join(home, ".gemini", "antigravity-cli")
	copyFile(t, token, filepath.Join(agyDir, "antigravity-oauth-token"), 0o600)
	for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
		if _, err := os.Stat(filepath.Join(realHome, ".gemini", "antigravity-cli", extra)); err == nil {
			copyFile(t, filepath.Join(realHome, ".gemini", "antigravity-cli", extra), filepath.Join(agyDir, extra), 0o600)
		}
	}
	// Tool calls proceed without prompting: the setting a headless run needs, scoped to this home.
	writeFile(t, filepath.Join(agyDir, "settings.json"), `{"toolPermission":"always-proceed"}`)

	// The plugin as it ships, with binaries built from this checkout.
	src, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	plugin := filepath.Join(home, ".gemini", "config", "plugins", "gray-area")
	copyFile(t, filepath.Join(src, "hooks.json"), filepath.Join(plugin, "hooks.json"), 0o644)
	copyFile(t, filepath.Join(src, ".antigravity-plugin", "plugin.json"), filepath.Join(plugin, ".antigravity-plugin", "plugin.json"), 0o644)
	copyFile(t, filepath.Join(src, "hooks", "fetch-bin.sh"), filepath.Join(plugin, "hooks", "fetch-bin.sh"), 0o755)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(plugin, "bin")+string(filepath.Separator),
		"./cmd/gray-area-capture", "./cmd/telepathy")
	build.Dir = filepath.Join(src, "tools")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	telepathy := filepath.Join(plugin, "bin", "telepathy")

	var env []string
	for _, e := range os.Environ() {
		k, _, _ := strings.Cut(e, "=")
		if k == "HOME" || strings.HasPrefix(k, "CLAUDE_") || strings.HasPrefix(k, "FEOV_") ||
			strings.HasPrefix(k, "ANTIGRAVITY_") || strings.HasPrefix(k, "XDG_") {
			continue
		}
		env = append(env, e)
	}
	env = append(env, "HOME="+home, "XDG_RUNTIME_DIR="+runtimeDir, "XDG_STATE_HOME="+stateDir)

	turn := func(conv, prompt string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()
		args := []string{"-p", prompt, "--model", "gemini-3.8-flash-low", "--print-timeout", "220s"}
		if conv != "" {
			args = append(args, "--conversation", conv)
		}
		cmd := exec.CommandContext(ctx, agy, args...)
		cmd.Dir, cmd.Env = work, env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("agy turn failed: %v\n%s", err, out)
		}
		return string(out)
	}
	record := func(conv string) string {
		t.Helper()
		s, err := peer.LookupConversationServer(runtimeDir, conv)
		if err != nil || s == nil {
			return ""
		}
		return s.Address
	}
	transcript := func(conv string) []map[string]any {
		t.Helper()
		f, err := os.Open(filepath.Join(agyDir, "brain", conv, ".system_generated", "logs", "transcript.jsonl"))
		if err != nil {
			t.Fatalf("transcript: %v", err)
		}
		defer f.Close()
		var steps []map[string]any
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<22), 1<<22)
		for sc.Scan() {
			var m map[string]any
			if json.Unmarshal(sc.Bytes(), &m) == nil {
				steps = append(steps, m)
			}
		}
		return steps
	}
	sessionRows := func(conv string) int {
		b, err := os.ReadFile(filepath.Join(work, ".claude", "gray-area", "trajectories-"+conv+".jsonl"))
		if err != nil {
			return 0
		}
		return strings.Count(strings.TrimSpace(string(b)), "\n") + 1
	}

	// The hook guarantees to ASK on every turn that starts unlinked; whether the model complies on a
	// given turn is the model's (about five turns in six on flash, measured). So each phase gets up to
	// three turns, every unlinked turn must carry the request, and the turns taken are logged.
	requests := func(conv string) int {
		n := 0
		for _, s := range transcript(conv) {
			if s["type"] == "EPHEMERAL_MESSAGE" && strings.Contains(fmt.Sprint(s["content"]), "' link") {
				n++
			}
		}
		return n
	}

	// 1. A new conversation is linked by the hook's request alone.
	turn("", "Reply with the single word OK, then stop.")
	entries, _ := os.ReadDir(filepath.Join(agyDir, "brain"))
	if len(entries) != 1 {
		t.Fatalf("want one conversation, have %d", len(entries))
	}
	conv := entries[0].Name()
	turns := 1
	for record(conv) == "" && turns < 3 {
		if got := requests(conv); got != turns {
			t.Errorf("phase 1: %d link requests after %d unlinked turns, want one per turn", got, turns)
		}
		turn(conv, "Reply with the single word OK, then stop.")
		turns++
	}
	first := record(conv)
	if first == "" {
		t.Fatalf("phase 1: still unlinked after %d turns, each asked; transcript: %v", turns, transcript(conv))
	}
	t.Logf("phase 1: linked to %s after %d turn(s)", first, turns)
	if n := sessionRows(conv); n != 1 {
		t.Errorf("phase 1: %d session rows, want SessionStart's one", n)
	}

	// 2. Resumed on a new server: relinked, and a self-send in the same turn wakes it through that
	//    server. A one-shot's server ends with its turn, so the link and the send share one.
	wakePrompt := "Run this shell command once and report its output verbatim: '" + telepathy +
		"' send \"$ANTIGRAVITY_CONVERSATION_ID\" 'LIVE_WAKE_7'. Then stop."
	var second string
	var viaServer, systemWake bool
	before := requests(conv)
	for turns = 1; turns <= 3; turns++ {
		turn(conv, wakePrompt)
		if requests(conv) <= before {
			t.Errorf("phase 2 turn %d: resumed on a new server with a stale link, and the hook did not ask", turns)
		}
		before = requests(conv)
		second = record(conv)
		for _, s := range transcript(conv) {
			b, _ := json.Marshal(s)
			if second != first && strings.Contains(string(b), "through its language server at "+second) {
				viaServer = true
			}
			if s["type"] == "SYSTEM_MESSAGE" && strings.Contains(string(b), "LIVE_WAKE_7") {
				systemWake = true
			}
		}
		if viaServer && systemWake {
			break
		}
	}
	t.Logf("phase 2: relinked to %s and woken after %d turn(s)", second, turns)
	if second == "" || second == first {
		t.Errorf("phase 2: link %q after resume, want a new server (was %q)", second, first)
	}
	if !viaServer {
		t.Errorf("phase 2: telepathy send never reported delivery through the language server it relinked to")
	}
	if !systemWake {
		t.Errorf("phase 2: LIVE_WAKE_7 never arrived as a SYSTEM_MESSAGE step")
	}
	if n := sessionRows(conv); n != 1 {
		t.Errorf("phase 2: %d session rows; a resume must not repeat SessionStart's work", n)
	}

	// 3. That server is gone: a send from outside any agent session queues and says why.
	send := exec.Command(telepathy, "send", conv, "LIVE_MAILBOX_9")
	send.Env = env
	out, err := send.CombinedOutput()
	if err != nil {
		t.Fatalf("telepathy send: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "QUEUED for antigravity conversation "+conv) ||
		!strings.Contains(string(out), "its language server at "+second+" did not answer") {
		t.Errorf("phase 3: send reported %q, want a mailbox delivery naming the dead server", out)
	}

	// 4. The next turn receives it: a system message carrying the text, and the mailbox marker gone.
	turn(conv, "Reply with the single word OK, then stop.")
	var delivered bool
	for _, s := range transcript(conv) {
		b, _ := json.Marshal(s)
		if s["type"] == "SYSTEM_MESSAGE" && strings.Contains(string(b), "LIVE_MAILBOX_9") {
			delivered = true
		}
	}
	if !delivered {
		t.Errorf("phase 4: the queued message never arrived as a SYSTEM_MESSAGE step on the next turn")
	}
	left, _ := filepath.Glob(filepath.Join(agyDir, "brain", conv, ".system_generated", "messages", "undelivered", "*"))
	if len(left) != 0 {
		t.Errorf("phase 4: %d mailbox markers still undelivered after the turn", len(left))
	}
}

// TestLiveAntigravityDaemonRestartWake verifies #1855 / #1750 end-to-end against a real agy:
//
//  1. Two conversations C1 and C2 are created and their initial servers exit (simulating a restart).
//  2. A fresh long-lived agy server starts for C1; ONLY C1 links (recording the server keyed by
//     PID + start time). C2 has never been loaded on this server.
//  3. telepathy send to C2 uses the restarted server's record, wakes C2, and C2 takes a full turn.
//  4. Negative control: while presence/<C2>.lock is held by another live process, telepathy send
//     to C2 refuses the live route and queues to C2's mailbox instead.
func TestLiveAntigravityDaemonRestartWake(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live Antigravity test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live Antigravity test under -short")
	}
	agy, err := exec.LookPath("agy")
	if err != nil {
		t.Skipf("agy not found on PATH: %v", err)
	}
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	token := filepath.Join(realHome, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if _, err := os.Stat(token); err != nil {
		t.Skipf("no agy credentials at %s", token)
	}

	home, work, runtimeDir, stateDir := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	agyDir := filepath.Join(home, ".gemini", "antigravity-cli")
	copyFile(t, token, filepath.Join(agyDir, "antigravity-oauth-token"), 0o600)
	for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
		if _, err := os.Stat(filepath.Join(realHome, ".gemini", "antigravity-cli", extra)); err == nil {
			copyFile(t, filepath.Join(realHome, ".gemini", "antigravity-cli", extra), filepath.Join(agyDir, extra), 0o600)
		}
	}
	writeFile(t, filepath.Join(agyDir, "settings.json"), `{"toolPermission":"always-proceed"}`)

	src, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	plugin := filepath.Join(home, ".gemini", "config", "plugins", "gray-area")
	copyFile(t, filepath.Join(src, "hooks.json"), filepath.Join(plugin, "hooks.json"), 0o644)
	copyFile(t, filepath.Join(src, ".antigravity-plugin", "plugin.json"), filepath.Join(plugin, ".antigravity-plugin", "plugin.json"), 0o644)
	copyFile(t, filepath.Join(src, "hooks", "fetch-bin.sh"), filepath.Join(plugin, "hooks", "fetch-bin.sh"), 0o755)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(plugin, "bin")+string(filepath.Separator),
		"./cmd/gray-area-capture", "./cmd/telepathy")
	build.Dir = filepath.Join(src, "tools")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	telepathy := filepath.Join(plugin, "bin", "telepathy")

	var env []string
	for _, e := range os.Environ() {
		k, _, _ := strings.Cut(e, "=")
		if k == "HOME" || strings.HasPrefix(k, "CLAUDE_") || strings.HasPrefix(k, "FEOV_") ||
			strings.HasPrefix(k, "ANTIGRAVITY_") || strings.HasPrefix(k, "XDG_") {
			continue
		}
		env = append(env, e)
	}
	env = append(env, "HOME="+home, "XDG_RUNTIME_DIR="+runtimeDir, "XDG_STATE_HOME="+stateDir)

	listBrain := func() map[string]bool {
		m := map[string]bool{}
		entries, _ := os.ReadDir(filepath.Join(agyDir, "brain"))
		for _, e := range entries {
			if e.IsDir() {
				m[e.Name()] = true
			}
		}
		return m
	}
	createConv := func() string {
		t.Helper()
		before := listBrain()
		cctx, ccancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer ccancel()
		cmd := exec.CommandContext(cctx, agy, "-p", "Reply with the single word OK, then stop.",
			"--model", "gemini-3.8-flash-low", "--print-timeout", "120s")
		cmd.Dir, cmd.Env = work, env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("createConv failed: %v\n%s", err, out)
		}
		for id := range listBrain() {
			if !before[id] {
				return id
			}
		}
		t.Fatal("createConv created no conversation directory")
		return ""
	}
	readTranscript := func(conv string) []map[string]any {
		t.Helper()
		f, err := os.Open(filepath.Join(agyDir, "brain", conv, ".system_generated", "logs", "transcript.jsonl"))
		if err != nil {
			t.Fatalf("transcript %s: %v", conv, err)
		}
		defer f.Close()
		var steps []map[string]any
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<22), 1<<22)
		for sc.Scan() {
			var m map[string]any
			if json.Unmarshal(sc.Bytes(), &m) == nil {
				steps = append(steps, m)
			}
		}
		return steps
	}

	// 1. Create two separate conversations whose initial servers exit (simulating pre-restart).
	c1 := createConv()
	c2 := createConv()
	t.Logf("created C1=%s C2=%s", c1, c2)
	c2Before := len(readTranscript(c2))

	// 2. Start a fresh long-lived agy server hosting C1 only (stream-json keeps the server alive
	//    until stdin closes), and link C1 on it.
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer srvCancel()
	srvCmd := exec.CommandContext(srvCtx, agy,
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--conversation", c1,
		"--model", "gemini-3.8-flash-low",
		"--dangerously-skip-permissions",
		"-p=",
	)
	srvCmd.Dir, srvCmd.Env = work, env
	stdin, err := srvCmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := srvCmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var srvStderr bytes.Buffer
	srvCmd.Stderr = &srvStderr
	if err := srvCmd.Start(); err != nil {
		t.Fatalf("start restarted agy server: %v", err)
	}
	defer func() {
		_ = stdin.Close()
		srvCancel()
		_ = srvCmd.Wait()
	}()

	srvStart, ok := peer.ProcessStartTime(srvCmd.Process.Pid)
	if !ok {
		t.Skip("ProcessStartTime unavailable on this platform")
	}
	srvID := peer.ServerIdentity{PID: srvCmd.Process.Pid, StartTime: srvStart, Daemon: true}

	// Drain stdout in background so the stream-json server never blocks on a full pipe.
	go func() { _, _ = io.Copy(io.Discard, stdout) }()

	linkMsg, _ := json.Marshal(map[string]any{
		"event":   "user",
		"message": map[string]any{"content": "Run this exact shell command once and reply DONE: '" + telepathy + "' link"},
	})
	if _, err := stdin.Write(append(linkMsg, '\n')); err != nil {
		t.Fatalf("write link prompt to C1: %v", err)
	}

	deadline := time.Now().Add(90 * time.Second)
	var linkedSrv *peer.Server
	for time.Now().Before(deadline) {
		s, err := peer.LookupDaemonServer(runtimeDir, srvID)
		if err == nil && s != nil {
			linkedSrv = s
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if linkedSrv == nil {
		t.Fatalf("C1 never linked restarted server %+v within 90s; stderr:\n%s", srvID, srvStderr.String())
	}
	t.Logf("C1 linked restarted server key=%s addr=%s", linkedSrv.Key, linkedSrv.Address)

	runSend := func(target, msg string) (string, error) {
		t.Helper()
		sendEnv := telecli.Env{
			BrainDirs:        []string{filepath.Join(agyDir, "brain")},
			AnnotationsDirs:  []string{filepath.Join(agyDir, "annotations")},
			RuntimeDir:       runtimeDir,
			Getenv:           func(k string) string { return map[string]string{"CLAUDE_CODE_SESSION_ID": "test-sender"}[k] },
			HostServer:       func() peer.ServerIdentity { return peer.ServerIdentity{} },
			CurrentDaemon:    func() (peer.ServerIdentity, bool) { return srvID, true },
			ConversationBusy: peer.ConversationBusy,
			Now:              time.Now,
		}
		var out, errBuf bytes.Buffer
		cmd := telecli.NewRoot(sendEnv)
		cmd.SetOut(&out)
		cmd.SetErr(&errBuf)
		cmd.SetArgs([]string{"send", target, msg})
		err := cmd.Execute()
		return out.String() + errBuf.String(), err
	}

	// 3. Negative control: while presence/<C2>.lock is held by another live process, telepathy send
	//    to C2 must refuse the live daemon route and queue to C2's mailbox instead.
	releaseLock := holdPresenceLock(t, agyDir, c2)
	outLocked, err := runSend(c2, "LIVE_LOCK_BLOCKED_1855")
	releaseLock()
	if err != nil {
		t.Fatalf("locked telepathy send failed: %v\n%s", err, outLocked)
	}
	if !strings.Contains(outLocked, "QUEUED for antigravity conversation "+c2) ||
		!strings.Contains(outLocked, "second writer") {
		t.Fatalf("expected locked send to refuse live route and queue, got:\n%s", outLocked)
	}
	left, _ := filepath.Glob(filepath.Join(agyDir, "brain", c2, ".system_generated", "messages", "undelivered", "*"))
	if len(left) == 0 {
		t.Fatalf("expected an undelivered mailbox marker for locked send to C2")
	}

	// 4. With the lock released, send to C2 (never loaded on srvCmd) and assert it wakes and takes
	//    a full turn through the restarted server's record.
	out, err := runSend(c2, "LIVE_RESTART_WAKE_1855: reply with the single word WOKEN.")
	if err != nil {
		t.Fatalf("telepathy send to C2 failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "delivered to antigravity conversation "+c2+" through its language server at "+linkedSrv.Address) {
		t.Fatalf("expected live delivery to C2 through %s, got:\n%s", linkedSrv.Address, out)
	}

	wakeDeadline := time.Now().Add(60 * time.Second)
	var sawWakeMsg, sawPlannerReply bool
	for time.Now().Before(wakeDeadline) {
		steps := readTranscript(c2)
		if len(steps) > c2Before {
			sawWakeMsg, sawPlannerReply = false, false
			for _, st := range steps[c2Before:] {
				b, _ := json.Marshal(st)
				if st["type"] == "SYSTEM_MESSAGE" && strings.Contains(string(b), "LIVE_RESTART_WAKE_1855") {
					sawWakeMsg = true
				}
				if sawWakeMsg && st["type"] == "PLANNER_RESPONSE" {
					sawPlannerReply = true
				}
			}
			if sawWakeMsg && sawPlannerReply {
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !sawWakeMsg || !sawPlannerReply {
		t.Fatalf("C2 did not wake and take a turn after send (sawWakeMsg=%v sawPlannerReply=%v); steps=%d (was %d)",
			sawWakeMsg, sawPlannerReply, len(readTranscript(c2)), c2Before)
	}
}

func copyFile(t *testing.T, from, to string, mode os.FileMode) {
	t.Helper()
	in, err := os.Open(from)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
