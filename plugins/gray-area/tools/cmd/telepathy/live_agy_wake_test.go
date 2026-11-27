//go:build linux

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/telecli"
	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// liveModel is the model every live turn here runs on.
const liveModel = "gemini-3.8-flash-low"

// TestLiveAntigravityMailboxWake drives telepathy's Antigravity delivery against a real agy, in a
// throwaway home: a long-lived language server started with a chosen ANTIGRAVITY_SERVER_CSRF_TOKEN
// stands in for the remote-control daemon (Env.CurrentDaemon names it), and send and agy-wake run
// in-process with a token file outside the home, so neither the host's token nor its daemon is ever
// reached, and an agent in the home that ran telepathy would find no token.
//
//   - ChosenTokenIsHonoured: DaemonAddress finds the server's one HTTP port; a wake with the chosen
//     token is taken and one with another token refused. Red when agy stops adopting the variable.
//   - SendWakesIdle: two sends to an idle conversation, the first while no daemon runs (queued, not
//     woken), the second woken: the turn the wake starts injects both queued messages, clears their
//     markers, and the model answers the queued content.
//   - SendHeldElsewhere: a conversation another server holds (it has run a turn there) gets no wake:
//     the message stays queued and the stand-in daemon never opens the conversation.
//   - WakeOnlyRealMail: agy-wake wakes the conversation holding a message from an agent, and not
//     those holding only a wake message or only an agy notice.
//   - SessionRowOnce: gray-area's hooks, built from the checkout and installed in the home, wrote
//     each conversation one session row at SessionStart and none when it resumed on another server.
//
// Gated like every live harness test (HOOKCORE_LIVE=1); it spends about ten agy flash turns.
func TestLiveAntigravityMailboxWake(t *testing.T) {
	r := newWakeRig(t)
	host := r.createConv()
	b, c, m, w, s := r.createConv(), r.createConv(), r.createConv(), r.createConv(), r.createConv()
	srv := r.startServer(host, r.token)
	daemon := peer.ServerIdentity{PID: srv.pid}
	if st, ok := peer.ProcessStartTime(srv.pid); ok {
		daemon.StartTime = st
	}

	t.Run("ChosenTokenIsHonoured", func(t *testing.T) {
		addr, err := peer.DaemonAddress(peer.NewAgyClient(5*time.Second), srv.pid)
		if err != nil {
			t.Fatalf("DaemonAddress on a real language server: %v", err)
		}
		if err := peer.SendWake(peer.NewAgyClient(30*time.Second), addr, "another-token-0000", host); err == nil || !strings.Contains(err.Error(), "401") {
			t.Errorf("a wake with another token = %v, want HTTP 401", err)
		}
		from := len(r.transcript(host))
		if err := peer.SendWake(peer.NewAgyClient(30*time.Second), addr, r.token, host); err != nil {
			t.Fatalf("a wake with the chosen token was refused — agy no longer adopts %s: %v", peer.DaemonTokenVar, err)
		}
		r.waitIdle(host, from, 120*time.Second)
	})

	t.Run("SendWakesIdle", func(t *testing.T) {
		from := len(r.transcript(b))
		first, second := "ALPHA"+r.suffix(), "BRAVO"+r.suffix()
		out := r.send(t, nil, b, "Queued message: remember the word "+first+".")
		if !strings.Contains(out, "not woken: no daemon is running") {
			t.Fatalf("the first send:\n%s", out)
		}
		out = r.send(t, &daemon, b, "Queued message: reply with the word "+first+" and the word "+second+".")
		if !strings.Contains(out, "  woken: the daemon's language server at ") {
			t.Fatalf("the second send:\n%s", out)
		}
		if !waitFor(120*time.Second, func() bool {
			inj, pl := r.stepsWith(b, from, second)
			return len(inj) > 0 && len(pl) > 0 && len(r.markers(b)) == 0
		}) {
			t.Fatalf("the woken turn did not read and answer the queued messages; markers left %v; steps:%s", r.markers(b), r.describe(b, from))
		}
		for _, word := range []string{first, second} {
			if inj, _ := r.stepsWith(b, from, word); len(inj) == 0 {
				t.Errorf("the queued message naming %s was not injected; steps:%s", word, r.describe(b, from))
			}
		}
		r.waitIdle(b, from, 60*time.Second)
	})

	t.Run("SendHeldElsewhere", func(t *testing.T) {
		holder := r.startServer(c, "")
		from := len(r.transcript(c))
		holder.prompt("Reply with the single word OK, then stop.")
		r.waitIdle(c, from, 120*time.Second)
		if h := peer.ConversationBusy(r.agyDir, c, srv.pid); h.PID != holder.pid {
			t.Fatalf("precondition: the holder check does not name the server that ran a turn on the conversation (pid %d): %+v", holder.pid, h)
		}
		out := r.send(t, &daemon, c, "Queued message: reply with the word HELD"+r.suffix()+".")
		if !strings.Contains(out, fmt.Sprintf("not woken: held by pid %d, not the daemon", holder.pid)) {
			t.Errorf("send to a held conversation:\n%s", out)
		}
		time.Sleep(10 * time.Second)
		if r.holdsDB(srv.pid, c) {
			t.Errorf("the stand-in daemon (pid %d) opened the held conversation: a second writer", srv.pid)
		}
		if n := len(r.markers(c)); n != 1 {
			t.Errorf("%d undelivered markers in the held conversation, want the one queued", n)
		}
		holder.stop()
	})

	t.Run("WakeOnlyRealMail", func(t *testing.T) {
		word := "MAIL" + r.suffix()
		r.enqueue(t, m, "e1e1e1e1-0000-4000-8000-0000000000a1", "Queued message: reply with the word "+word+".")
		r.enqueue(t, w, peer.WakeSender, peer.WakeContent)
		r.enqueue(t, s, peer.SystemSender, "[Notice] a notice from agy")
		fromM, fromW, fromS := len(r.transcript(m)), len(r.transcript(w)), len(r.transcript(s))
		var log bytes.Buffer
		root := telecli.NewRoot(r.env(&daemon))
		root.SetOut(&log)
		root.SetErr(&log)
		root.SetArgs([]string{"agy-wake"})
		if err := root.Execute(); err != nil {
			t.Fatalf("agy-wake: %v\n%s", err, log.String())
		}
		t.Logf("agy-wake:\n%s", log.String())
		if !strings.Contains(log.String(), "woke "+m) || strings.Contains(log.String(), w) || strings.Contains(log.String(), s) {
			t.Errorf("agy-wake's lines name the wrong conversations:\n%s", log.String())
		}
		if !waitFor(120*time.Second, func() bool {
			_, pl := r.stepsWith(m, fromM, word)
			return len(pl) > 0 && len(r.markers(m)) == 0
		}) {
			t.Errorf("the conversation with mail was not woken to read it; steps:%s", r.describe(m, fromM))
		}
		r.waitIdle(m, fromM, 60*time.Second)
		for conv, from := range map[string]int{w: fromW, s: fromS} {
			if n := len(r.transcript(conv)); n != from {
				t.Errorf("conversation %s, holding no mail, took a turn; steps:%s", conv, r.describe(conv, from))
			}
			if n := len(r.markers(conv)); n != 1 {
				t.Errorf("conversation %s has %d undelivered markers, want its one entry left unread", conv, n)
			}
		}
	})

	// Every conversation was created by agy -p, and host, b, c and m then resumed on another server.
	t.Run("SessionRowOnce", func(t *testing.T) {
		for _, conv := range []string{host, b, c, m, w, s} {
			if n := r.sessionRows(conv); n != 1 {
				t.Errorf("conversation %s has %d session rows, want SessionStart's one and none on resume", conv, n)
			}
		}
	})
}

// wakeRig is one throwaway agy home with the token a stand-in daemon is started with.
type wakeRig struct {
	t                         *testing.T
	agy, home, agyDir, work   string
	env0                      []string
	token, tokenFile, scratch string
}

func newWakeRig(t *testing.T) *wakeRig {
	t.Helper()
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live Antigravity test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live Antigravity test under -short")
	}
	agy, err := exec.LookPath("agy")
	if err != nil {
		t.Fatalf("agy not found on PATH under HOOKCORE_LIVE=1: %v", err)
	}
	r := &wakeRig{t: t, agy: agy, home: testhome.NewLiveHome(t), work: t.TempDir(), scratch: t.TempDir()}
	if err := testhome.StageAgyAccessToken(r.home); err != nil {
		t.Fatalf("antigravity credentials unavailable under HOOKCORE_LIVE=1: %v", err)
	}
	r.agyDir = filepath.Join(r.home, ".gemini", "antigravity-cli")
	writeLiveFile(t, filepath.Join(r.agyDir, "settings.json"), `{"toolPermission":"always-proceed"}`, 0o644)
	r.installGrayArea()
	r.env0 = testhome.LiveEnv(t, r.home, "XDG_RUNTIME_DIR="+t.TempDir(), "XDG_CONFIG_HOME="+t.TempDir(), "XDG_STATE_HOME="+t.TempDir())
	r.token = "live-" + r.suffix() + r.suffix()
	// Outside the home: an agent in it that ran telepathy would find no token there.
	r.tokenFile = filepath.Join(r.scratch, "gray-area", "antigravity-daemon.env")
	if err := os.MkdirAll(filepath.Dir(r.tokenFile), 0o700); err != nil {
		t.Fatal(err)
	}
	writeLiveFile(t, r.tokenFile, peer.DaemonTokenVar+"="+r.token+"\n", 0o600)
	return r
}

// installGrayArea installs the checkout's gray-area plugin into the home: its hooks.json, manifest,
// fetch-bin.sh, and both binaries built from this tree.
func (r *wakeRig) installGrayArea() {
	r.t.Helper()
	src, err := filepath.Abs("../../..")
	if err != nil {
		r.t.Fatal(err)
	}
	plugin := filepath.Join(r.home, ".gemini", "config", "plugins", "gray-area")
	for _, f := range []struct {
		from, to string
		mode     os.FileMode
	}{
		{"hooks.json", "hooks.json", 0o644},
		{".antigravity-plugin/plugin.json", ".antigravity-plugin/plugin.json", 0o644},
		{"hooks/fetch-bin.sh", "hooks/fetch-bin.sh", 0o755},
	} {
		b, err := os.ReadFile(filepath.Join(src, f.from))
		if err != nil {
			r.t.Fatal(err)
		}
		writeLiveFile(r.t, filepath.Join(plugin, f.to), string(b), f.mode)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(plugin, "bin")+string(filepath.Separator),
		"./cmd/gray-area-capture", "./cmd/telepathy")
	build.Dir = filepath.Join(src, "tools")
	if out, err := build.CombinedOutput(); err != nil {
		r.t.Fatalf("build: %v\n%s", err, out)
	}
}

// sessionRows counts the session rows gray-area's capture hook wrote for conv.
func (r *wakeRig) sessionRows(conv string) int {
	b, err := os.ReadFile(filepath.Join(r.work, ".claude", "gray-area", "trajectories-"+conv+".jsonl"))
	if err != nil {
		return 0
	}
	return strings.Count(strings.TrimSpace(string(b)), "\n") + 1
}

func (r *wakeRig) suffix() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

// env is the in-process Env of a send or agy-wake in this home, its daemon the stand-in (nil: none).
func (r *wakeRig) env(daemon *peer.ServerIdentity) telecli.Env {
	return telecli.Env{
		Store:            filepath.Join(r.scratch, "absent.db"),
		ProjectsDir:      r.scratch,
		SessionsDir:      filepath.Join(r.scratch, "sessions"),
		BrainDirs:        []string{filepath.Join(r.agyDir, "brain")},
		AnnotationsDirs:  []string{filepath.Join(r.agyDir, "annotations")},
		Now:              time.Now,
		DaemonRoot:       r.agyDir,
		TokenFile:        r.tokenFile,
		Getenv:           func(string) string { return "" },
		CurrentDaemon:    func() (peer.ServerIdentity, bool) { return derefDaemon(daemon) },
		ConversationBusy: peer.ConversationBusy,
		DaemonAddress:    peer.DaemonAddress,
		GOOS:             "linux",
	}
}

func derefDaemon(d *peer.ServerIdentity) (peer.ServerIdentity, bool) {
	if d == nil {
		return peer.ServerIdentity{}, false
	}
	return *d, true
}

// send runs telepathy send in-process and returns its stdout, failing on any error or on the token
// appearing in its output.
func (r *wakeRig) send(t *testing.T, daemon *peer.ServerIdentity, conv, text string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	root := telecli.NewRoot(r.env(daemon))
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"send", conv, text})
	if err := root.Execute(); err != nil {
		t.Fatalf("send: %v\n%s", err, errOut.String())
	}
	if strings.Contains(out.String()+errOut.String(), r.token) {
		t.Fatal("send printed the token")
	}
	t.Logf("send %s:\n%s", conv, out.String())
	return out.String()
}

func (r *wakeRig) enqueue(t *testing.T, conv, sender, content string) {
	t.Helper()
	if _, err := peer.EnqueueAgy(filepath.Join(r.agyDir, "brain", conv),
		peer.AgyMessage{Recipient: conv, Sender: sender, Title: "live test", Content: content}, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func (r *wakeRig) createConv() string {
	r.t.Helper()
	before := r.convs()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, r.agy, "-p", "Reply with the single word OK, then stop.",
		"--model", liveModel, "--print-timeout", "120s")
	cmd.Dir, cmd.Env = r.work, r.env0
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("creating a conversation: %v\n%s", err, out)
	}
	for id := range r.convs() {
		if !before[id] {
			return id
		}
	}
	r.t.Fatal("agy -p created no conversation directory")
	return ""
}

func (r *wakeRig) convs() map[string]bool {
	m := map[string]bool{}
	entries, _ := os.ReadDir(filepath.Join(r.agyDir, "brain"))
	for _, e := range entries {
		if e.IsDir() {
			m[e.Name()] = true
		}
	}
	return m
}

// liveServer is one long-lived stream-json agy language server.
type liveServer struct {
	pid   int
	stdin io.WriteCloser
	done  chan struct{}
	cmd   *exec.Cmd
}

// startServer starts a stream-json server on conv (with token as its server token when not empty)
// and waits until it listens.
func (r *wakeRig) startServer(conv, token string) *liveServer {
	r.t.Helper()
	env := append([]string{}, r.env0...)
	if token != "" {
		env = append(env, peer.DaemonTokenVar+"="+token)
	}
	cmd := exec.Command(r.agy, "--input-format", "stream-json", "--output-format", "stream-json",
		"--model", liveModel, "--conversation", conv, "-p=")
	cmd.Dir, cmd.Env = r.work, env
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		r.t.Fatalf("starting a language server: %v", err)
	}
	s := &liveServer{pid: cmd.Process.Pid, stdin: stdin, done: make(chan struct{}), cmd: cmd}
	go func() { _, _ = io.Copy(io.Discard, stdout) }()
	go func() { _ = cmd.Wait(); close(s.done) }()
	r.t.Cleanup(s.stop)
	if !waitFor(60*time.Second, func() bool {
		_, err := peer.DaemonAddress(peer.NewAgyClient(2*time.Second), s.pid)
		return err == nil
	}) {
		r.t.Fatalf("the language server (pid %d) never listened", s.pid)
	}
	return s
}

func (s *liveServer) prompt(text string) {
	b, _ := json.Marshal(map[string]any{"event": "user", "message": map[string]any{"content": text}})
	_, _ = s.stdin.Write(append(b, '\n'))
}

func (s *liveServer) stop() {
	_ = s.stdin.Close()
	select {
	case <-s.done:
	case <-time.After(30 * time.Second):
		_ = syscall.Kill(s.pid, syscall.SIGKILL)
		<-s.done
	}
}

func (r *wakeRig) transcript(conv string) []map[string]any {
	f, err := os.Open(filepath.Join(r.agyDir, "brain", conv, ".system_generated", "logs", "transcript.jsonl"))
	if err != nil {
		return nil
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

func stepText(s map[string]any) string {
	tc, _ := json.Marshal(s["tool_calls"])
	return fmt.Sprint(s["content"]) + " " + string(tc)
}

// stepsWith lists conv's steps from from on that mention word: injected (any type but the model's)
// and the model's own responses.
func (r *wakeRig) stepsWith(conv string, from int, word string) (injected, planner []int) {
	steps := r.transcript(conv)
	for i := from; i < len(steps); i++ {
		if !strings.Contains(stepText(steps[i]), word) {
			continue
		}
		if steps[i]["type"] == "PLANNER_RESPONSE" {
			planner = append(planner, i)
		} else {
			injected = append(injected, i)
		}
	}
	return injected, planner
}

func (r *wakeRig) describe(conv string, from int) string {
	var b strings.Builder
	steps := r.transcript(conv)
	for i := from; i < len(steps); i++ {
		c := strings.ReplaceAll(stepText(steps[i]), "\n", " ")
		if len(c) > 300 {
			c = c[:300]
		}
		fmt.Fprintf(&b, "\n  [%d] %v %q", i, steps[i]["type"], c)
	}
	return b.String()
}

// waitIdle waits until conv has a model response after from and its transcript has not grown for
// five seconds, so a server is never stopped mid-turn.
func (r *wakeRig) waitIdle(conv string, from int, d time.Duration) {
	last, changed := len(r.transcript(conv)), time.Now()
	for end := time.Now().Add(d); time.Now().Before(end); time.Sleep(time.Second) {
		steps := r.transcript(conv)
		if len(steps) != last {
			last, changed = len(steps), time.Now()
			continue
		}
		for i := from; i < len(steps); i++ {
			if steps[i]["type"] == "PLANNER_RESPONSE" && time.Since(changed) >= 5*time.Second {
				return
			}
		}
	}
}

func (r *wakeRig) markers(conv string) []string {
	entries, _ := os.ReadDir(filepath.Join(r.agyDir, "brain", conv, ".system_generated", "messages", "undelivered"))
	var ids []string
	for _, e := range entries {
		ids = append(ids, e.Name())
	}
	return ids
}

// holdsDB reports whether pid has conv's database (or its -wal/-shm) open.
func (r *wakeRig) holdsDB(pid int, conv string) bool {
	fds, _ := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	want := filepath.Join(r.agyDir, "conversations", conv+".db")
	for _, f := range fds {
		if target, err := os.Readlink(fmt.Sprintf("/proc/%d/fd/%s", pid, f.Name())); err == nil && strings.HasPrefix(target, want) {
			return true
		}
	}
	return false
}

func waitFor(d time.Duration, f func() bool) bool {
	for end := time.Now().Add(d); time.Now().Before(end); time.Sleep(500 * time.Millisecond) {
		if f() {
			return true
		}
	}
	return f()
}

func writeLiveFile(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}
