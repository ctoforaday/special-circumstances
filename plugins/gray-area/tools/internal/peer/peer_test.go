package peer

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
	"time"
)

var at = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

func agyShell(conv, addr, token string) func(string) string {
	return func(k string) string {
		return map[string]string{EnvConversation: conv, EnvAddress: addr, EnvToken: token}[k]
	}
}

func TestRegisterRecordsTheServerPrivately(t *testing.T) {
	rt := t.TempDir()
	_, got, err := Register(rt, agyShell("conv-1", "localhost:1234", "tok"), at)
	if err != nil || got != Recorded {
		t.Fatalf("Register = %v, %v; want Recorded", got, err)
	}
	s, err := LookupServer(rt, "conv-1")
	if err != nil || s == nil {
		t.Fatalf("LookupServer = %v, %v", s, err)
	}
	if *s != (Server{Conversation: "conv-1", Address: "localhost:1234", Token: "tok", RegisteredAt: at.Unix()}) {
		t.Errorf("record = %+v", *s)
	}
	if runtime.GOOS == "windows" {
		return // no Unix permission bits to assert
	}
	for path, want := range map[string]os.FileMode{RegistryDir(rt): 0o700, filepath.Join(RegistryDir(rt), "conv-1.json"): 0o600} {
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Mode().Perm(); got != want {
			t.Errorf("%s is %o, want %o: the token controls the whole language server", path, got, want)
		}
	}
}

// ONCE PER SERVER LAUNCH. Every telepathy verb run from an Antigravity shell calls Register, so a
// record that already names this server must not be rewritten; a restarted server, which has a new
// token, must replace it.
func TestRegisterWritesOncePerServerLaunch(t *testing.T) {
	rt := t.TempDir()
	if _, got, err := Register(rt, agyShell("conv-1", "localhost:1", "tok-a"), at); got != Recorded || err != nil {
		t.Fatalf("first Register = %v, %v; want Recorded", got, err)
	}
	if _, got, err := Register(rt, agyShell("conv-1", "localhost:1", "tok-a"), at.Add(time.Hour)); got != AlreadyLinked || err != nil {
		t.Errorf("same server again: Register = %v, %v; want AlreadyLinked", got, err)
	}
	if s, _ := LookupServer(rt, "conv-1"); s == nil || s.RegisteredAt != at.Unix() {
		t.Errorf("record = %+v; the same server's record was rewritten", s)
	}
	if _, got, err := Register(rt, agyShell("conv-1", "localhost:2", "tok-b"), at.Add(2*time.Hour)); got != Recorded || err != nil {
		t.Errorf("restarted server: Register = %v, %v; want Recorded", got, err)
	}
	if s, _ := LookupServer(rt, "conv-1"); s == nil || s.Token != "tok-b" || s.Address != "localhost:2" {
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
		if _, got, err := Register(rt, env, at); got != NotAnAntigravityShell || err != nil {
			t.Errorf("%s: Register = %v, %v; want NotAnAntigravityShell and no error", name, got, err)
		}
	}
	if _, got, err := Register("", agyShell("conv-1", "localhost:1", "tok"), at); got != NoRuntimeDir || err != nil {
		t.Errorf("no runtime dir: Register = %v, %v; want NoRuntimeDir", got, err)
	}
	if _, err := os.Stat(RegistryDir(rt)); !os.IsNotExist(err) {
		t.Errorf("a registry directory was created with nothing to put in it")
	}
}

func TestRegisterRefusesAnIdThatIsAPath(t *testing.T) {
	rt := t.TempDir()
	for _, id := range []string{"../escape", "a/b", "..", `a\b`} {
		if _, got, err := Register(rt, agyShell(id, "localhost:1", "tok"), at); got == Recorded || err == nil {
			t.Errorf("Register(%q) = %v, %v; want a refusal", id, got, err)
		}
	}
}

func TestLookupOfAnUnrecordedConversationIsNil(t *testing.T) {
	if s, err := LookupServer(t.TempDir(), "never"); s != nil || err != nil {
		t.Errorf("LookupServer = %v, %v; want nil, nil", s, err)
	}
	if s, err := LookupServer("", "never"); s != nil || err != nil {
		t.Errorf("LookupServer with no runtime dir = %v, %v; want nil, nil", s, err)
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
		AgyMessage{Recipient: "conv-1", Sender: "me", Title: "T", Content: "C"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Route != RouteAgyServer || d.Via != addr {
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
	_, err := SendAgyServer(srv.Client(), Server{Address: strings.TrimPrefix(srv.URL, "http://"), Token: "stale"}, AgyMessage{})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want the status named", err)
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
