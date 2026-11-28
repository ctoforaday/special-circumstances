package telecli

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

const (
	claudePeer = "c1a0de00-0000-4000-8000-000000000001"
	agyPeer    = "a9e00000-0000-4000-8000-000000000002"
	agySelf    = "a9e5e1f0-0000-4000-8000-000000000003"
)

// sendWorld is a box with one running Claude Code session, two Antigravity conversations under the
// CLI daemon root, a token file path (no file until a test writes one), and a caller whose
// environment the test chooses.
type sendWorld struct {
	*harness
	brain, tokenFile, socket string
	callerEnv                map[string]string
}

func newSendWorld(t *testing.T) *sendWorld {
	t.Helper()
	agyRoot := t.TempDir()
	w := &sendWorld{harness: newColdHarness(t), brain: filepath.Join(agyRoot, "brain"),
		tokenFile: filepath.Join(t.TempDir(), "gray-area", "antigravity-daemon.env"), callerEnv: map[string]string{}}
	dir, err := os.MkdirTemp("", "tp")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	w.socket = filepath.Join(dir, "s.sock")
	body, _ := json.Marshal(map[string]any{"pid": 5150, "sessionId": claudePeer, "cwd": "/w",
		"procStart": "0", "messagingSocketPath": w.socket})
	if err := os.WriteFile(filepath.Join(w.sessions, "5150.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{agyPeer, agySelf} {
		if err := os.MkdirAll(filepath.Join(w.brain, id, ".system_generated"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	w.env.BrainDirs = []string{w.brain}
	w.env.DaemonRoot = agyRoot
	w.env.TokenFile = w.tokenFile
	w.env.Getenv = func(k string) string { return w.callerEnv[k] }
	// Isolate unit tests from the host machine's running daemon, /proc and platform: no daemon
	// runs, nobody holds a conversation, a port lookup fails the test, and a wake is judged as on
	// Linux.
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return peer.ServerIdentity{}, false }
	w.env.ConversationBusy = func(string, string, int) peer.Holder { return peer.Holder{} }
	w.env.DaemonAddress = func(*http.Client, int) (string, error) {
		t.Error("the daemon's port was looked up where no test expected it")
		return "", errors.New("unexpected")
	}
	w.env.DaemonListens = func(int, string) (bool, error) {
		t.Error("the daemon's port was checked where no test expected it")
		return false, errors.New("unexpected")
	}
	w.env.GOOS = "linux"
	return w
}

// The daemon of the tests, and the token its file holds.
var newDaemon = peer.ServerIdentity{PID: 284285, StartTime: 30000}

const testToken = "test-token-7f3a"

// provisioned writes the token file and stands in the daemon: running as newDaemon, its HTTP port at
// a stand-in server that answers status and records each wake it is sent.
func (w *sendWorld) provisioned(t *testing.T, status int) (addr string, got <-chan map[string]string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the daemon's token file is read on Unix only")
	}
	if err := os.MkdirAll(filepath.Dir(w.tokenFile), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(w.tokenFile, []byte(peer.DaemonTokenVar+"="+testToken+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ch := make(chan map[string]string, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		m := map[string]string{}
		_ = json.Unmarshal(b, &m)
		m["token"] = r.Header.Get("x-codeium-csrf-token")
		ch <- m
		rw.WriteHeader(status)
		_, _ = rw.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)
	addr = strings.TrimPrefix(srv.URL, "http://")
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.env.DaemonAddress = func(_ *http.Client, pid int) (string, error) {
		if pid != newDaemon.PID {
			t.Errorf("the port of pid %d was looked up, not the daemon's", pid)
		}
		return addr, nil
	}
	w.env.DaemonListens = func(pid int, a string) (bool, error) { return pid == newDaemon.PID && a == addr, nil }
	return addr, ch
}

func (w *sendWorld) sendFromClaude(t *testing.T) string {
	t.Helper()
	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}
	out, errOut, code := w.run(t, "send", agyPeer, "hello")
	if code != 0 {
		t.Fatalf("send: exit %d, stderr %q", code, errOut)
	}
	if strings.Contains(out+errOut, testToken) {
		t.Errorf("send printed the token:\n%s%s", out, errOut)
	}
	return out
}

func dbHolder(pid int) peer.Holder {
	return peer.Holder{PID: pid, Kind: peer.HolderDBOpen, Detail: "db open"}
}

// THE MESSAGE IS QUEUED AND THE CONVERSATION WOKEN: the text goes to the mailbox only, and the
// daemon gets the measured wake message with the token from the file.
func TestSendQueuesAndWakesTheIdleConversation(t *testing.T) {
	w := newSendWorld(t)
	addr, got := w.provisioned(t, http.StatusOK)
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "QUEUED for antigravity conversation "+agyPeer) || !strings.Contains(out, "woken: the daemon's language server at "+addr) {
		t.Errorf("stdout = %q", out)
	}
	wake := <-got
	want := map[string]string{"recipient": agyPeer, "content": peer.WakeContent, "senderId": peer.WakeSender, "displayTitle": peer.WakeTitle, "token": testToken}
	for k, v := range want {
		if wake[k] != v {
			t.Errorf("wake %s = %q, want %q", k, wake[k], v)
		}
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
	}
	files, _ := filepath.Glob(filepath.Join(w.brain, agyPeer, ".system_generated", "messages", "*.json"))
	if len(files) != 1 {
		t.Fatalf("%d mailbox messages, want 1", len(files))
	}
	if b, _ := os.ReadFile(files[0]); !strings.Contains(string(b), "hello") {
		t.Errorf("the mailbox message does not hold the text: %s", b)
	}
}

// THE TOKEN FILE IS READ BEFORE ANY CONTACT: with none, no daemon is looked for, no port probed and
// no holder inspected, and the message is queued saying who provisions it.
func TestSendReadsTheTokenBeforeAnyContact(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the daemon's token file is read on Unix only")
	}
	w := newSendWorld(t)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) {
		t.Error("the daemon was looked for without a token file")
		return newDaemon, true
	}
	w.env.ConversationBusy = func(string, string, int) peer.Holder {
		t.Error("a holder was inspected without a token file")
		return peer.Holder{}
	}
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "not woken: no daemon token file at "+filepath.ToSlash(w.tokenFile)+": the human provisions one with telepathy agy-provision") {
		t.Errorf("stdout = %q", out)
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
	}
}

// A CONVERSATION ANOTHER PROCESS HOLDS GETS NO WAKE, and the holder check excludes the daemon.
func TestSendHeldElsewhereSendsNoWake(t *testing.T) {
	w := newSendWorld(t)
	_, got := w.provisioned(t, http.StatusOK)
	w.env.ConversationBusy = func(_, _ string, exclude int) peer.Holder {
		if exclude != newDaemon.PID {
			t.Errorf("the holder check excluded pid %d, not the daemon", exclude)
		}
		return dbHolder(9999)
	}
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "not woken: held by pid 9999, not the daemon") {
		t.Errorf("stdout = %q", out)
	}
	select {
	case m := <-got:
		t.Errorf("a wake went to a conversation another process holds: %v", m)
	default:
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
	}
}

// WITH NO CONFIGURATION DIRECTORY THERE IS NO TOKEN FILE, and no daemon is looked for.
func TestSendWithoutAConfigurationDirectory(t *testing.T) {
	w := newSendWorld(t)
	w.env.TokenFile = ""
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) {
		t.Error("the daemon was looked for without a token file")
		return newDaemon, true
	}
	if out := w.sendFromClaude(t); !strings.Contains(out, "not woken: no configuration directory, so no daemon token file") {
		t.Errorf("stdout = %q", out)
	}
}

// THE TOKEN GOES ONLY TO A PORT THE DAEMON STILL HOLDS: checked just before the request, so a port
// the daemon let go of meanwhile, and another process took, never receives it.
func TestSendSendsNoTokenToAnAddressTheDaemonLeft(t *testing.T) {
	w := newSendWorld(t)
	addr, got := w.provisioned(t, http.StatusOK)
	w.env.DaemonListens = func(int, string) (bool, error) { return false, nil }
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "not woken: the daemon (pid 284285) no longer listens at "+addr+", so its token is not sent there") {
		t.Errorf("stdout = %q", out)
	}
	select {
	case m := <-got:
		t.Errorf("the token went to an address the daemon no longer holds: %v", m)
	default:
	}
}

// A WAKE THAT FAILS LEAVES THE MESSAGE QUEUED, says why, and the send still succeeds.
func TestSendWakeFailureLeavesItQueued(t *testing.T) {
	w := newSendWorld(t)
	w.provisioned(t, http.StatusUnauthorized)
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "not woken: the wake failed: ") || !strings.Contains(out, "HTTP 401") {
		t.Errorf("stdout = %q", out)
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
	}
}

// inbox stands in the Claude Code session's socket and returns the first line it receives.
func (w *sendWorld) inbox(t *testing.T) <-chan string {
	t.Helper()
	ln, err := net.Listen("unix", w.socket)
	if err != nil {
		t.Skipf("this platform cannot listen on a Unix socket: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	got := make(chan string, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		line, _ := bufio.NewReader(c).ReadString('\n')
		got <- line
	}()
	return got
}

func (w *sendWorld) mailbox(t *testing.T) []string {
	t.Helper()
	got, _ := filepath.Glob(filepath.Join(w.brain, agyPeer, ".system_generated", "messages", "undelivered", "*"))
	return got
}

func TestSendToAClaudeSessionUsesItsInboxSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native Windows refuses every post to a Claude Code inbox; the peer package tests the refusal")
	}
	w := newSendWorld(t)
	got := w.inbox(t)
	w.callerEnv["CLAUDE_CODE_SESSION_ID"] = "caller-session"
	out, errOut, code := w.run(t, "send", "c1a0", "rebase after #1730")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "delivered to claude session "+claudePeer+" over its inbox socket") {
		t.Errorf("stdout = %q", out)
	}
	var line struct {
		Type    string
		Message struct{ Role, Content string }
	}
	if err := json.Unmarshal([]byte(<-got), &line); err != nil {
		t.Fatal(err)
	}
	want := "[telepathy] message from claude session caller-session; reply with: telepathy send caller-session '<reply>'\n\nrebase after #1730"
	if line.Type != "user" || line.Message.Role != "user" || line.Message.Content != want {
		t.Errorf("line = %+v, want content %q", line, want)
	}
}

func TestSendRefusesAnAmbiguousPrefixNamingEach(t *testing.T) {
	w := newSendWorld(t)
	_, errOut, code := w.run(t, "send", "a9e", "hello")
	if code != 2 || !strings.Contains(errOut, "names 2 sessions") ||
		!strings.Contains(errOut, "antigravity "+agyPeer) || !strings.Contains(errOut, "antigravity "+agySelf) {
		t.Errorf("exit %d, stderr %q", code, errOut)
	}
	if n := len(w.mailbox(t)); n != 0 {
		t.Errorf("an ambiguous send queued %d messages", n)
	}
}

func TestSendToNobodyIsWorded(t *testing.T) {
	w := newSendWorld(t)
	_, errOut, code := w.run(t, "send", "ffff", "hello")
	if code != 1 || !strings.Contains(errOut, `no running Claude Code session (advertised in <SESSIONS>) and no Antigravity conversation`) {
		t.Errorf("exit %d, stderr %q", code, errOut)
	}
}

func TestSendRefusesAPathOrAnEmptyMessage(t *testing.T) {
	w := newSendWorld(t)
	for _, args := range [][]string{{"send", "../x", "hi"}, {"send", "a9e0", "  "}} {
		if _, _, code := w.run(t, args...); code != 2 {
			t.Errorf("%v exited %d, want 2", args, code)
		}
	}
}

func TestSendShowsAgyTitleInReportAndRejectsTitleAsTarget(t *testing.T) {
	w := newSendWorld(t)
	annDir := t.TempDir()
	w.env.AnnotationsDirs = []string{annDir}
	if err := os.WriteFile(filepath.Join(annDir, agyPeer+".pbtxt"), []byte(`title:"Cross Platform Plugin Configuration"`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, errOut, code := w.run(t, "send", "a9e0", "hello")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, `QUEUED for antigravity conversation `+agyPeer+` ("Cross Platform Plugin Configuration") in its mailbox`) {
		t.Errorf("stdout = %q", out)
	}
	// Display only: a title is never accepted as a target.
	if _, _, code := w.run(t, "send", "Cross Platform Plugin Configuration", "hello"); code == 0 {
		t.Errorf("send accepted title as target")
	}
}
