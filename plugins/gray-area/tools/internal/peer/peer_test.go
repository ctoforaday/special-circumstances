package peer

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var at = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

func agyShell(conv, addr, token string) func(string) string {
	return func(k string) string {
		return map[string]string{EnvConversation: conv, EnvAddress: addr, EnvToken: token}[k]
	}
}

func hostID(pid int, start int64, daemon bool) func() ServerIdentity {
	return func() ServerIdentity {
		return ServerIdentity{PID: pid, StartTime: start, Daemon: daemon}
	}
}

func TestRegisterRecordsTheServerPrivatelyKeyedByServer(t *testing.T) {
	rt := t.TempDir()
	s, got, err := RegisterWith(rt, agyShell("conv-1", "localhost:1234", "tok"), hostID(4242, 99999, true), at)
	if err != nil || got != Recorded {
		t.Fatalf("RegisterWith = %v, %v; want Recorded", got, err)
	}
	wantKey := "pid-4242-99999"
	if s.Key != wantKey || s.PID != 4242 || s.StartTime != 99999 || !s.Daemon || s.Address != "localhost:1234" || s.Token != "tok" || s.RegisteredAt != at.Unix() {
		t.Fatalf("returned record = %+v", s)
	}
	byConv, err := LookupConversationServer(rt, "conv-1", allLive)
	if err != nil || byConv == nil || byConv.Key != wantKey {
		t.Fatalf("LookupConversationServer = %+v, %v", byConv, err)
	}
	byDaemon, err := LookupDaemonServer(rt, ServerIdentity{PID: 4242, StartTime: 99999, Daemon: true})
	if err != nil || byDaemon == nil || byDaemon.Address != "localhost:1234" {
		t.Fatalf("LookupDaemonServer = %+v, %v", byDaemon, err)
	}
	if _, err := os.Stat(filepath.Join(RegistryDir(rt), "conv-1.json")); !os.IsNotExist(err) {
		t.Errorf("legacy conversation-keyed file exists; records must be keyed by server only")
	}
	if runtime.GOOS == "windows" {
		return // no Unix permission bits to assert
	}
	for path, want := range map[string]os.FileMode{RegistryDir(rt): 0o700, filepath.Join(RegistryDir(rt), wantKey+".json"): 0o600} {
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Mode().Perm(); got != want {
			t.Errorf("%s is %o, want %o: the token controls the whole language server", path, got, want)
		}
	}
}

func TestRegisterTightensExistingDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no Unix permission bits on Windows")
	}
	rt := t.TempDir()
	dir := RegistryDir(rt)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, got, err := RegisterWith(rt, agyShell("conv-1", "localhost:1", "tok"), hostID(10, 20, true), at); err != nil || got != Recorded {
		t.Fatalf("RegisterWith = %v, %v", got, err)
	}
	st, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := st.Mode().Perm(); perm != 0o700 {
		t.Errorf("registry dir perm = %o, want 0700", perm)
	}
}

func TestRegisterConcurrentLinksPreserveAllConversations(t *testing.T) {
	rt := t.TempDir()
	h := hostID(555, 77777, true)
	const n = 8
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := fmt.Sprintf("conv-%d", idx)
			_, _, err := RegisterWith(rt, agyShell(id, "localhost:555", "tok"), h, at)
			if err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent RegisterWith error: %v", err)
	}
	s, err := LookupDaemonServer(rt, ServerIdentity{PID: 555, StartTime: 77777, Daemon: true})
	if err != nil || s == nil {
		t.Fatalf("LookupDaemonServer = %+v, %v", s, err)
	}
	if len(s.Conversations) != n {
		t.Errorf("got %d conversations (%v), want %d", len(s.Conversations), s.Conversations, n)
	}
}

// ONE RECORD PER SERVER LAUNCH. Every telepathy verb run from an Antigravity shell calls Register:
// a conversation already listed is AlreadyLinked and keeps the record's first registration time; a
// second conversation on the same server shares the same server record; a restarted server writes
// a new server record.
func TestRegisterKeepsOneRecordPerServerLaunchAndSharesAcrossConversations(t *testing.T) {
	rt := t.TempDir()
	h1 := hostID(100, 5000, true)
	if _, got, err := RegisterWith(rt, agyShell("conv-1", "localhost:1", "tok-a"), h1, at); got != Recorded || err != nil {
		t.Fatalf("first RegisterWith = %v, %v; want Recorded", got, err)
	}
	if _, got, err := RegisterWith(rt, agyShell("conv-1", "localhost:1", "tok-a"), h1, at.Add(time.Hour)); got != AlreadyLinked || err != nil {
		t.Errorf("same server and conv again: RegisterWith = %v, %v; want AlreadyLinked", got, err)
	}
	if s, _ := LookupConversationServer(rt, "conv-1", allLive); s == nil || s.RegisteredAt != at.Unix() {
		t.Errorf("record = %+v; the server's first registration time moved", s)
	}
	// Second conversation on the same server joins the same record without bumping RegisteredAt.
	if _, got, err := RegisterWith(rt, agyShell("conv-2", "localhost:1", "tok-a"), h1, at.Add(time.Hour)); got != Recorded || err != nil {
		t.Errorf("second conv on same server: RegisterWith = %v, %v; want Recorded", got, err)
	}
	if servers, _ := ListServers(rt); len(servers) != 1 || len(servers[0].Conversations) != 2 {
		t.Errorf("servers = %+v; want 1 shared server record with 2 conversations", servers)
	}
	// Restarted server (new PID/start time): creates a new record; LookupConversationServer returns the newest.
	h2 := hostID(200, 6000, true)
	if _, got, err := RegisterWith(rt, agyShell("conv-1", "localhost:2", "tok-b"), h2, at.Add(2*time.Hour)); got != Recorded || err != nil {
		t.Errorf("restarted server: RegisterWith = %v, %v; want Recorded", got, err)
	}
	if s, _ := LookupConversationServer(rt, "conv-1", allLive); s == nil || s.Token != "tok-b" || s.Address != "localhost:2" || s.PID != 200 {
		t.Errorf("record = %+v; want the restarted server", s)
	}
}

func TestRegisterOutsideAnAntigravityShellRecordsNothing(t *testing.T) {
	rt := t.TempDir()
	for name, env := range map[string]func(string) string{
		"no variables":    agyShell("", "", ""),
		"no token":        agyShell("conv-1", "localhost:1", ""),
		"no address":      agyShell("conv-1", "", "tok"),
		"no conversation": agyShell("", "localhost:1", "tok"),
	} {
		if _, got, err := RegisterWith(rt, env, hostID(10, 20, false), at); got != NotAnAntigravityShell || err != nil {
			t.Errorf("%s: RegisterWith = %v, %v; want NotAnAntigravityShell and no error", name, got, err)
		}
	}
	if _, got, err := RegisterWith("", agyShell("conv-1", "localhost:1", "tok"), hostID(10, 20, false), at); got != NoRuntimeDir || err != nil {
		t.Errorf("no runtime dir: RegisterWith = %v, %v; want NoRuntimeDir", got, err)
	}
	if _, err := os.Stat(RegistryDir(rt)); !os.IsNotExist(err) {
		t.Errorf("a registry directory was created with nothing to put in it")
	}
}

func TestRegisterRefusesAnIdThatIsAPath(t *testing.T) {
	rt := t.TempDir()
	for _, id := range []string{"../escape", "a/b", "..", `a\b`} {
		if _, got, err := RegisterWith(rt, agyShell(id, "localhost:1", "tok"), hostID(10, 20, false), at); got == Recorded || err == nil {
			t.Errorf("RegisterWith(%q) = %v, %v; want a refusal", id, got, err)
		}
	}
}

func TestLookupOfAnUnrecordedConversationIsNil(t *testing.T) {
	if s, err := LookupConversationServer(t.TempDir(), "never", allLive); s != nil || err != nil {
		t.Errorf("LookupConversationServer = %v, %v; want nil, nil", s, err)
	}
	if s, err := LookupConversationServer("", "never", allLive); s != nil || err != nil {
		t.Errorf("LookupConversationServer with no runtime dir = %v, %v; want nil, nil", s, err)
	}
	if s, err := LookupDaemonServer(t.TempDir(), ServerIdentity{PID: 1, StartTime: 2}); s != nil || err != nil {
		t.Errorf("LookupDaemonServer = %v, %v; want nil, nil", s, err)
	}
}

// shortSocket returns a socket path short enough for sun_path on every platform.
func shortSocket(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "tp")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

// listenInbox stands in a Claude Code session's inbox socket and returns the first line it
// receives.
func listenInbox(t *testing.T, path string) <-chan string {
	t.Helper()
	ln, err := net.Listen("unix", path)
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

func TestSendClaudeWritesOneUserMessageLine(t *testing.T) {
	if nativeWindows {
		t.Skip("native Windows refuses every post; TestSendClaudeRefusesOnNativeWindows covers it")
	}
	sock := shortSocket(t)
	got := listenInbox(t, sock)
	d, err := SendClaude(sock, "hello\nworld")
	if err != nil {
		t.Fatal(err)
	}
	if d.Route != RouteClaudeSocket || d.Via != sock {
		t.Errorf("delivery = %+v", d)
	}
	line := <-got
	want := `{"type":"user","message":{"role":"user","content":"hello\nworld"}}` + "\n"
	if line != want {
		t.Errorf("line = %q, want %q", line, want)
	}
}

func TestSendClaudeRefusesOnNativeWindows(t *testing.T) {
	was := nativeWindows
	nativeWindows = true
	defer func() { nativeWindows = was }()
	if _, err := SendClaude("pipe", "x"); err == nil || !strings.Contains(err.Error(), "native Windows") {
		t.Errorf("err = %v, want the Windows refusal", err)
	}
}

func TestSendClaudeToADeadSocketIsAnError(t *testing.T) {
	if nativeWindows {
		t.Skip("native Windows refuses before dialling")
	}
	if _, err := SendClaude(shortSocket(t), "x"); err == nil || !strings.Contains(err.Error(), "refused") {
		t.Errorf("err = %v, want a refusal naming the socket", err)
	}
	if _, err := SendClaude("", "x"); err == nil {
		t.Errorf("an empty socket path was not refused")
	}
}

func TestSendAgyServerSpeaksSendAgentMessage(t *testing.T) {
	var gotPath, gotToken string
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotToken = r.URL.Path, r.Header.Get("x-codeium-csrf-token")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()
	addr := strings.TrimPrefix(srv.URL, "http://")
	d, err := SendAgyServer(srv.Client(), Server{Address: addr, Token: "tok"},
		AgyMessage{Recipient: "conv-1", Sender: "me", Title: "T", Content: "C"}, RouteAgyHolderRecord)
	if err != nil {
		t.Fatal(err)
	}
	if d.Route != RouteAgyHolderRecord || d.Via != addr {
		t.Errorf("delivery = %+v", d)
	}
	if gotPath != "/exa.language_server_pb.LanguageServerService/SendAgentMessage" || gotToken != "tok" {
		t.Errorf("path %q token %q", gotPath, gotToken)
	}
	want := map[string]string{"recipient": "conv-1", "senderId": "me", "displayTitle": "T", "content": "C"}
	for k, v := range want {
		if gotBody[k] != v {
			t.Errorf("body[%s] = %q, want %q", k, gotBody[k], v)
		}
	}
}

func TestSendAgyServerRefusalIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "csrf", http.StatusForbidden)
	}))
	defer srv.Close()
	_, err := SendAgyServer(srv.Client(), Server{Address: strings.TrimPrefix(srv.URL, "http://"), Token: "stale"}, AgyMessage{}, RouteAgyOwnRecord)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want the status named", err)
	}
}

// ONLY A CONNECTION NEVER OPENED PROVES THE MESSAGE DID NOT ARRIVE. A refused connection is an
// UnreachedError; a server that took the request and did not answer, or answered an error status,
// is not.
func TestSendAgyServerSaysWhetherTheMessageCanHaveArrived(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := ln.Addr().String()
	ln.Close()

	release := make(chan struct{})
	hang := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer hang.Close()
	defer close(release)
	fail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer fail.Close()

	client := &http.Client{Timeout: 200 * time.Millisecond}
	for _, tc := range []struct {
		name, addr string
		unreached  bool
	}{
		{"refused", closed, true},
		{"no answer", strings.TrimPrefix(hang.URL, "http://"), false},
		{"error status", strings.TrimPrefix(fail.URL, "http://"), false},
	} {
		_, err := SendAgyServer(client, Server{Address: tc.addr, Token: "tok"}, AgyMessage{}, RouteAgyOwnRecord)
		var u *UnreachedError
		if err == nil || errors.As(err, &u) != tc.unreached {
			t.Errorf("%s: err = %v, want unreached=%v", tc.name, err, tc.unreached)
		}
	}
}

func TestEnqueueAgyWritesTheBodyAndTheMarker(t *testing.T) {
	conv := filepath.Join(t.TempDir(), "conv-1")
	d, err := EnqueueAgy(conv, AgyMessage{Recipient: "conv-1", Sender: "me", Title: "T", Content: "C"}, at)
	if err != nil {
		t.Fatal(err)
	}
	if d.Route != RouteAgyMailbox {
		t.Errorf("route = %s", d.Route)
	}
	b, err := os.ReadFile(d.Via)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	id, _ := m["id"].(string)
	if filepath.Base(d.Via) != id+".json" || len(id) != 36 {
		t.Errorf("file %s does not carry its id %q", d.Via, id)
	}
	for k, v := range map[string]string{"recipient": "conv-1", "sender": "me", "content": "C",
		"priority": "MESSAGE_PRIORITY_HIGH", "timestamp": "2026-11-04T12:00:00Z"} {
		if m[k] != v {
			t.Errorf("%s = %v, want %q", k, m[k], v)
		}
	}
	if rd, _ := m["renderDetails"].(map[string]any); rd["messageTitle"] != "T" {
		t.Errorf("renderDetails = %v", m["renderDetails"])
	}
	box := filepath.Join(conv, ".system_generated", "messages")
	if _, err := os.Stat(filepath.Join(box, "undelivered", id)); err != nil {
		t.Errorf("no undelivered marker: %v", err)
	}
	left, _ := filepath.Glob(filepath.Join(box, ".*.tmp"))
	if len(left) != 0 {
		t.Errorf("temporary files left behind: %v", left)
	}
}
