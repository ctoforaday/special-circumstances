package peer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicksToTime(t *testing.T) {
	boot := time.Date(2026, 11, 26, 3, 0, 0, 0, time.UTC).Unix()
	if got, want := StartTimeAt(boot, 12345), time.Date(2026, 11, 26, 3, 2, 3, 450_000_000, time.UTC); !got.Equal(want) {
		t.Errorf("StartTimeAt(boot, 12345) = %s, want %s", got.UTC(), want)
	}
}

// A MESSAGE THE CLIENT READS WHILE THE SCAN RUNS IS READ, NOT UNREADABLE. A stand-in client removes
// each marker and then its message file while scans run; no scan may count an unreadable message.
func TestScanSkipsMessagesReadDuringScan(t *testing.T) {
	brain := t.TempDir()
	box := filepath.Join(brain, "conv-busy", ".system_generated", "messages")
	if err := os.MkdirAll(filepath.Join(box, "undelivered"), 0o755); err != nil {
		t.Fatal(err)
	}
	const n = 300
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("m%03d", i)
		body := fmt.Sprintf(`{"id":%q,"sender":"s","timestamp":"2026-09-09T10:00:00Z","content":"c"}`, id)
		if err := os.WriteFile(filepath.Join(box, id+".json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(box, "undelivered", id), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < n; i++ {
			id := fmt.Sprintf("m%03d", i)
			_ = os.Remove(filepath.Join(box, "undelivered", id))
			_ = os.Remove(filepath.Join(box, id+".json"))
		}
	}()
	for {
		ws, errs := ScanMailboxes([]string{brain}, "")
		if len(errs) != 0 {
			t.Fatalf("scan errors: %v", errs)
		}
		for _, w := range ws {
			if w.Unreadable != 0 {
				t.Fatalf("a scan counted %d messages the client was reading as unreadable", w.Unreadable)
			}
		}
		select {
		case <-done:
			return
		default:
		}
	}
}

// A WAKE IS NOT MAIL. A mailbox's wake entries are counted apart, never in Messages or Oldest, and
// a mailbox that holds only wakes is still listed.
func TestScanCountsWakesApart(t *testing.T) {
	brain := t.TempDir()
	put := func(conv, id, sender, at string) {
		box := filepath.Join(brain, conv, ".system_generated", "messages")
		if err := os.MkdirAll(filepath.Join(box, "undelivered"), 0o755); err != nil {
			t.Fatal(err)
		}
		body := fmt.Sprintf(`{"id":%q,"sender":%q,"timestamp":%q,"content":"c"}`, id, sender, at)
		if err := os.WriteFile(filepath.Join(box, id+".json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(box, "undelivered", id), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("mixed", "w1", WakeSender, "2026-11-01T00:00:00Z")
	put("mixed", "m1", "agent-x", "2026-11-02T00:00:00Z")
	put("mixed", "s1", SystemSender, "2026-10-01T00:00:00Z")
	put("wakeonly", "w2", WakeSender, "2026-11-03T00:00:00Z")
	ws, errs := ScanMailboxes([]string{brain}, "")
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	got := map[string]Waiting{}
	for _, w := range ws {
		got[w.Conversation] = w
	}
	m := got["mixed"]
	if m.Messages != 1 || m.Wakes != 1 || m.Notices != 1 || !m.Oldest.Equal(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("mixed = %+v, want 1 message (oldest 11-02), 1 wake, 1 notice", m)
	}
	w, ok := got["wakeonly"]
	if !ok || w.Messages != 0 || w.Wakes != 1 || !w.Oldest.IsZero() {
		t.Errorf("wakeonly = %+v (listed %v), want listed with 1 wake and no message", w, ok)
	}
}

// THE WAKE IS THE MEASURED MESSAGE, with the server's token, and any answer but 200 is an error that
// does not carry the token.
func TestSendWake(t *testing.T) {
	const token = "tok-123-abc"
	var got map[string]string
	var header string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/exa.language_server_pb.LanguageServerService/SendAgentMessage" {
			http.NotFound(w, r)
			return
		}
		header = r.Header.Get("x-codeium-csrf-token")
		_ = json.NewDecoder(r.Body).Decode(&got)
		if header != token {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":"unauthenticated","message":"unauthenticated: invalid CSRF token"}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	addr := strings.TrimPrefix(srv.URL, "http://")
	if err := SendWake(NewAgyClient(5*time.Second), addr, token, "conv-1"); err != nil {
		t.Fatalf("SendWake: %v", err)
	}
	want := map[string]string{"recipient": "conv-1", "content": WakeContent, "senderId": WakeSender, "displayTitle": WakeTitle}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("request %s = %q, want %q", k, got[k], v)
		}
	}
	if WakeContent != "You have queued messages." || WakeSender != "telepathy-wake" || WakeTitle != "telepathy" {
		t.Errorf("the wake is not the measured message: %q from %q titled %q", WakeContent, WakeSender, WakeTitle)
	}
	err := SendWake(NewAgyClient(5*time.Second), addr, "wrong-token", "conv-1")
	if err == nil || !strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "wrong-token") {
		t.Errorf("a refused wake = %v, want an HTTP 401 error without the token", err)
	}
}
