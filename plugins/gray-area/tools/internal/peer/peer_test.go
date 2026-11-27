package peer

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var at = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

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
