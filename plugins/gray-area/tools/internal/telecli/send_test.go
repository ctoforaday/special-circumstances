package telecli

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

const (
	claudePeer = "c1a0de00-0000-4000-8000-000000000001"
	agyPeer    = "a9e00000-0000-4000-8000-000000000002"
	agySelf    = "a9e5e1f0-0000-4000-8000-000000000003"
)

// sendWorld is a box with one running Claude Code session, one Antigravity conversation, a
// runtime directory, and a caller whose environment the test chooses.
type sendWorld struct {
	*harness
	brain, runtime, socket string
	callerEnv              map[string]string
}

func newSendWorld(t *testing.T) *sendWorld {
	t.Helper()
	agyRoot := t.TempDir()
	w := &sendWorld{harness: newColdHarness(t), brain: filepath.Join(agyRoot, "brain"), runtime: t.TempDir(), callerEnv: map[string]string{}}
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
	w.env.RuntimeDir = w.runtime
	w.env.Getenv = func(k string) string { return w.callerEnv[k] }
	// Isolate unit tests from the host machine's running daemon, /proc and platform: nobody holds a
	// conversation, every record's server runs, and live delivery is judged as on Linux.
	w.env.HostServer = func() peer.ServerIdentity { return peer.ServerIdentity{} }
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return peer.ServerIdentity{}, false }
	w.env.ConversationBusy = func(string, string, int) peer.Holder { return peer.Holder{} }
	w.env.RecordLiveness = func(peer.Server) catalogue.Liveness { return catalogue.Live }
	w.env.GOOS = "linux"
	return w
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

// languageServer stands in the server hosting agyPeer and records what it is sent.
func (w *sendWorld) languageServer(t *testing.T) (addr string, got <-chan map[string]string) {
	t.Helper()
	ch := make(chan map[string]string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-codeium-csrf-token") != "tok" {
			http.Error(rw, "csrf", http.StatusForbidden)
			return
		}
		b, _ := io.ReadAll(r.Body)
		var m map[string]string
		_ = json.Unmarshal(b, &m)
		ch <- m
		rw.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://"), ch
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

func TestSendToAnUnrecordedConversationQueuesAndSaysSo(t *testing.T) {
	w := newSendWorld(t)
	out, errOut, code := w.run(t, "send", "a9e0", "hello")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "QUEUED for antigravity conversation "+agyPeer) {
		t.Errorf("stdout = %q", out)
	}
	if !strings.Contains(out, "no daemon is running") || !strings.Contains(out, "opening a daemon-hosted Antigravity conversation restores the live route") {
		t.Errorf("stdout = %q, want no-daemon + daemon-hosted restore hint", out)
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
	}
}

// THE CALL SITE, not just the function: the record a later send relies on is written by running
// ANY verb from the conversation's own shell. Deleting the root's registration fails this test.
func TestAVerbRunFromAnAntigravityShellMakesItReachableLive(t *testing.T) {
	w := newSendWorld(t)
	addr, got := w.languageServer(t)
	const hostPID = 6006
	w.env.HostServer = func() peer.ServerIdentity {
		return peer.ServerIdentity{PID: hostPID, StartTime: 66666}
	}
	var gotBusyPID int
	w.env.ConversationBusy = func(_, _ string, excludePID int) peer.Holder {
		gotBusyPID = excludePID
		return peer.Holder{}
	}

	// The recipient runs some unrelated verb from its own shell.
	w.callerEnv = map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: addr, peer.EnvToken: "tok"}
	w.run(t, "send", "c1a0", "anything") // fails (no inbox listening); the record is the point
	if s, _ := peer.LookupConversationServer(w.runtime, agyPeer, w.env.RecordLiveness); s == nil || s.Address != addr {
		t.Fatalf("running a verb from the conversation's shell recorded %+v", s)
	}

	// Another session sends to it.
	w.callerEnv = map[string]string{peer.EnvConversation: agySelf}
	out, errOut, code := w.run(t, "send", agyPeer, "wake up", "--title", "Ping")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "delivered to antigravity conversation "+agyPeer+" via antigravity-server-own, the language server at "+addr) {
		t.Errorf("stdout = %q", out)
	}
	if gotBusyPID != hostPID {
		t.Errorf("the probe before sending received excludePID=%d, want %d", gotBusyPID, hostPID)
	}
	m := <-got
	if m["recipient"] != agyPeer || m["senderId"] != agySelf || m["displayTitle"] != "Ping" ||
		!strings.HasSuffix(m["content"], "\n\nwake up") {
		t.Errorf("server was sent %v", m)
	}
	if n := len(w.mailbox(t)); n != 0 {
		t.Errorf("a live delivery also queued %d mailbox copies", n)
	}
}

// THE CAUSE PRINTED IS THE ONE THAT STOPPED THE LAST TARGET TRIED. The record's server cannot be
// reached, and the daemon, the next target, is not running: that is what queued the message.
func TestSendUnreachedCauseIsLastTarget(t *testing.T) {
	w := newSendWorld(t)
	addr := unreachable(t)
	stale := func(k string) string {
		return map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: addr, peer.EnvToken: "tok"}[k]
	}
	if _, _, err := peer.RegisterWith(w.runtime, stale, w.env.HostServer, frozen); err != nil {
		t.Fatal(err)
	}
	out, errOut, code := w.run(t, "send", agyPeer, "hello")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "QUEUED") || !strings.Contains(out, "not delivered live: no daemon is running") || strings.Contains(out, "cannot be reached") {
		t.Errorf("stdout = %q, want the daemon's cause and not the record's", out)
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
	}
}

func TestSendWhenOwnRecordIsDaemonDoesNotMislabelAsNoDaemonRecorded(t *testing.T) {
	w := newSendWorld(t)
	addr := unreachable(t)
	daemon := peer.ServerIdentity{PID: 3003, StartTime: 33333, Daemon: true}
	w.env.HostServer = func() peer.ServerIdentity { return daemon }
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return daemon, true }
	stale := func(k string) string {
		return map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: addr, peer.EnvToken: "tok"}[k]
	}
	if _, _, err := peer.RegisterWith(w.runtime, stale, w.env.HostServer, frozen); err != nil {
		t.Fatal(err)
	}
	out, errOut, code := w.run(t, "send", agyPeer, "hello")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "QUEUED") || !strings.Contains(out, "the language server recorded at "+addr+" cannot be reached") {
		t.Errorf("stdout = %q", out)
	}
	if strings.Contains(out, "has not been linked") {
		t.Errorf("stdout mislabelled a failed send to the current daemon as the daemon unlinked: %q", out)
	}
}

func TestSendRefusesOwnRouteWhenConversationIsBusy(t *testing.T) {
	w := newSendWorld(t)
	addr, got := w.languageServer(t)
	host := peer.ServerIdentity{PID: 4004, StartTime: 44444, Daemon: true}
	w.env.HostServer = func() peer.ServerIdentity { return host }
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return host, true }
	w.env.ConversationBusy = func(_, conv string, _ int) peer.Holder {
		return peer.Holder{PID: 9999, Kind: peer.HolderDBOpen, Detail: "pid 9999 has conversations/" + conv + ".db open"}
	}
	w.callerEnv = map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: addr, peer.EnvToken: "tok"}
	if _, errOut, code := w.run(t, "link"); code != 0 {
		t.Fatalf("link: exit %d, stderr %q", code, errOut)
	}
	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}
	out, errOut, code := w.run(t, "send", agyPeer, "do not double-write route 1")
	if code != 0 {
		t.Fatalf("send: exit %d, stderr %q", code, errOut)
	}
	if !strings.Contains(out, "QUEUED") || !strings.Contains(out, "held by another process, pid 9999; sending live would create a second writer") {
		t.Errorf("stdout = %q", out)
	}
	select {
	case m := <-got:
		t.Errorf("route 1 delivered live despite ConversationBusy: %v", m)
	default:
	}
}

func TestSendRefusesDaemonRouteForNonCLIRoot(t *testing.T) {
	w := newSendWorld(t)
	daemonAddr, got := w.languageServer(t)
	// Put agyPeer under a separate IDE root, NOT w.env.DaemonRoot.
	ideRoot := t.TempDir()
	ideConv := "1de00000-0000-4000-8000-000000000099"
	if err := os.MkdirAll(filepath.Join(ideRoot, "brain", ideConv, ".system_generated"), 0o755); err != nil {
		t.Fatal(err)
	}
	w.env.BrainDirs = append(w.env.BrainDirs, filepath.Join(ideRoot, "brain"))

	postRestart := peer.ServerIdentity{PID: 2002, StartTime: 22222, Daemon: true}
	w.env.HostServer = func() peer.ServerIdentity { return postRestart }
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return postRestart, true }
	w.callerEnv = map[string]string{peer.EnvConversation: agySelf, peer.EnvAddress: daemonAddr, peer.EnvToken: "tok"}
	if _, errOut, code := w.run(t, "link"); code != 0 {
		t.Fatalf("link: exit %d, stderr %q", code, errOut)
	}

	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}
	out, errOut, code := w.run(t, "send", ideConv, "hello ide")
	if code != 0 {
		t.Fatalf("send: exit %d, stderr %q", code, errOut)
	}
	if !strings.Contains(out, "QUEUED for antigravity conversation "+ideConv) ||
		!strings.Contains(out, "is not the CLI daemon root") {
		t.Errorf("stdout = %q", out)
	}
	select {
	case m := <-got:
		t.Errorf("daemon route sent to non-CLI root conversation: %v", m)
	default:
	}
}

func TestSendWakesUnlinkedConversationThroughRestartedDaemonRecord(t *testing.T) {
	w := newSendWorld(t)
	daemonAddr, got := w.languageServer(t)

	// agyPeer had a stale pre-restart record on a dead server.
	preRestart := peer.ServerIdentity{PID: 1001, StartTime: 11111, Daemon: true}
	if _, _, err := peer.RegisterWith(w.runtime, func(k string) string {
		return map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: "127.0.0.1:1", peer.EnvToken: "old-tok"}[k]
	}, func() peer.ServerIdentity { return preRestart }, frozen); err != nil {
		t.Fatal(err)
	}

	// After restart, ONLY agySelf links on the new daemon.
	postRestart := peer.ServerIdentity{PID: 2002, StartTime: 22222, Daemon: true}
	w.env.HostServer = func() peer.ServerIdentity { return postRestart }
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return postRestart, true }
	var busyPIDs []int
	w.env.ConversationBusy = func(_, _ string, excludePID int) peer.Holder {
		busyPIDs = append(busyPIDs, excludePID)
		return peer.Holder{}
	}
	// The pre-restart server is gone.
	w.env.RecordLiveness = func(s peer.Server) catalogue.Liveness {
		if s.PID == preRestart.PID {
			return catalogue.Ended
		}
		return catalogue.Live
	}
	w.callerEnv = map[string]string{peer.EnvConversation: agySelf, peer.EnvAddress: daemonAddr, peer.EnvToken: "tok"}
	if _, errOut, code := w.run(t, "link"); code != 0 {
		t.Fatalf("link agySelf: exit %d, stderr %q", code, errOut)
	}

	// External send to agyPeer falls through to the restarted daemon's record and delivers live.
	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}
	out, errOut, code := w.run(t, "send", agyPeer, "wake after restart")
	if code != 0 {
		t.Fatalf("send: exit %d, stderr %q", code, errOut)
	}
	if !strings.Contains(out, "delivered to antigravity conversation "+agyPeer+" via antigravity-daemon, the language server at "+daemonAddr) {
		t.Errorf("stdout = %q", out)
	}
	// One discovery excluding nobody, then the probe excluding only the daemon it sends through.
	if len(busyPIDs) != 2 || busyPIDs[0] != 0 || busyPIDs[1] != postRestart.PID {
		t.Errorf("ConversationBusy excludePIDs = %v, want [0, %d]", busyPIDs, postRestart.PID)
	}
	m := <-got
	if m["recipient"] != agyPeer || !strings.HasSuffix(m["content"], "\n\nwake after restart") {
		t.Errorf("daemon received %v", m)
	}
	if n := len(w.mailbox(t)); n != 0 {
		t.Errorf("live delivery queued %d mailbox copies", n)
	}
}

func TestSendRefusesDaemonRouteWhenPresenceLockIsHeld(t *testing.T) {
	w := newSendWorld(t)
	daemonAddr, got := w.languageServer(t)

	postRestart := peer.ServerIdentity{PID: 2002, StartTime: 22222, Daemon: true}
	w.env.HostServer = func() peer.ServerIdentity { return postRestart }
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return postRestart, true }
	w.env.ConversationBusy = func(_, conv string, _ int) peer.Holder {
		return peer.Holder{PID: 7777, Kind: peer.HolderPresenceLock, Detail: "pid 7777 holds presence/" + conv + ".lock"}
	}

	// agySelf links on the restarted daemon.
	w.callerEnv = map[string]string{peer.EnvConversation: agySelf, peer.EnvAddress: daemonAddr, peer.EnvToken: "tok"}
	if _, errOut, code := w.run(t, "link"); code != 0 {
		t.Fatalf("link: exit %d, stderr %q", code, errOut)
	}

	// Sending to agyPeer must NOT use the daemon route while its lock is held.
	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}
	out, errOut, code := w.run(t, "send", agyPeer, "do not double-write")
	if code != 0 {
		t.Fatalf("send: exit %d, stderr %q", code, errOut)
	}
	if !strings.Contains(out, "QUEUED for antigravity conversation "+agyPeer) ||
		!strings.Contains(out, "second writer") {
		t.Errorf("stdout = %q", out)
	}
	select {
	case m := <-got:
		t.Errorf("daemon received live send despite held presence lock: %v", m)
	default:
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
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

func TestLinkReportsWhatItFound(t *testing.T) {
	w := newSendWorld(t)
	w.callerEnv = map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: "localhost:9", peer.EnvToken: "tok"}
	out, errOut, code := w.run(t, "link")
	if code != 0 || !strings.Contains(out, "linked antigravity conversation "+agyPeer+" to its language server at localhost:9") {
		t.Errorf("first link: exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	out, _, code = w.run(t, "link")
	if code != 0 || !strings.Contains(out, "already linked") {
		t.Errorf("second link: exit %d, stdout %q", code, out)
	}
	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "c"}
	if _, errOut, code := w.run(t, "link"); code != 1 || !strings.Contains(errOut, "not an Antigravity agent's shell") {
		t.Errorf("from a Claude shell: exit %d, stderr %q", code, errOut)
	}
	w.env.RuntimeDir = ""
	w.callerEnv = map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: "localhost:9", peer.EnvToken: "tok"}
	if _, errOut, code := w.run(t, "link"); code != 1 || !strings.Contains(errOut, "XDG_RUNTIME_DIR is unset") {
		t.Errorf("no runtime dir: exit %d, stderr %q", code, errOut)
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
