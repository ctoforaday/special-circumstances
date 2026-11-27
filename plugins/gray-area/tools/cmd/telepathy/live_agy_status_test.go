package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// TestLiveAntigravitySendAgentMessageRefusals pins the statuses peer.RefusedError reads as "refused
// before acting" against a real language server: 401 for no token and for a wrong one, 404 for a
// method that does not exist, 405 for a GET. For each, the conversation the request named gains no
// step and never sees the text. It pins too the 400 a body that is not JSON gets, which is not a
// refusal (a method's own validation answers 400 as well). The disconfirming case follows: the same
// request made valid answers 200 and the text arrives.
//
// The server is a long-lived agy in an isolated home, linked by its own conversation, so the token
// is read from the test's own runtime directory and never printed: a body is logged with it
// replaced. Gated like every live harness test (HOOKCORE_LIVE=1); it spends three agy flash turns.
func TestLiveAntigravitySendAgentMessageRefusals(t *testing.T) {
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
	home := testhome.NewLiveHome(t)
	work, runtimeDir, stateDir := t.TempDir(), t.TempDir(), t.TempDir()
	if err := testhome.StageAgyAccessToken(home); err != nil {
		t.Fatalf("antigravity credentials unavailable under HOOKCORE_LIVE=1: %v", err)
	}
	agyDir := filepath.Join(home, ".gemini", "antigravity-cli")
	writeFile(t, filepath.Join(agyDir, "settings.json"), `{"toolPermission":"always-proceed"}`)

	src, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "bin")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", bin+string(filepath.Separator), "./cmd/telepathy")
	build.Dir = filepath.Join(src, "tools")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	telepathy := filepath.Join(bin, "telepathy")
	env := testhome.LiveEnv(t, home, "XDG_RUNTIME_DIR="+runtimeDir, "XDG_STATE_HOME="+stateDir)

	brain := filepath.Join(agyDir, "brain")
	conversations := func() map[string]bool {
		m := map[string]bool{}
		entries, _ := os.ReadDir(brain)
		for _, e := range entries {
			if e.IsDir() {
				m[e.Name()] = true
			}
		}
		return m
	}
	createConv := func() string {
		t.Helper()
		before := conversations()
		cctx, ccancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer ccancel()
		cmd := exec.CommandContext(cctx, agy, "-p", "Reply with the single word OK, then stop.",
			"--model", "gemini-3.8-flash-low", "--print-timeout", "120s")
		cmd.Dir, cmd.Env = work, env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("create a conversation: %v\n%s", err, out)
		}
		for id := range conversations() {
			if !before[id] {
				return id
			}
		}
		t.Fatal("agy created no conversation directory")
		return ""
	}
	transcript := func(conv string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(brain, conv, ".system_generated", "logs", "transcript.jsonl"))
		if err != nil {
			t.Fatalf("transcript of %s: %v", conv, err)
		}
		return string(b)
	}

	// The host conversation links the long-lived server it runs on; the target is another
	// conversation, not loaded there, whose transcript shows whether a request acted.
	host, target := createConv(), createConv()
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer srvCancel()
	srv := exec.CommandContext(srvCtx, agy, "--input-format", "stream-json", "--output-format", "stream-json",
		"--conversation", host, "--model", "gemini-3.8-flash-low", "-p=")
	srv.Dir, srv.Env = work, env
	stdin, err := srv.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := srv.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var srvErr bytes.Buffer
	srv.Stderr = &srvErr
	if err := srv.Start(); err != nil {
		t.Fatalf("start the long-lived agy: %v", err)
	}
	defer func() {
		_ = stdin.Close()
		srvCancel()
		_ = srv.Wait()
	}()
	go func() { _, _ = io.Copy(io.Discard, stdout) }()
	prompt, _ := json.Marshal(map[string]any{"event": "user",
		"message": map[string]any{"content": "Run this exact shell command once and reply DONE: '" + telepathy + "' link"}})
	if _, err := stdin.Write(append(prompt, '\n')); err != nil {
		t.Fatalf("ask the host conversation to link: %v", err)
	}
	var server *peer.Server
	for deadline := time.Now().Add(120 * time.Second); time.Now().Before(deadline) && server == nil; time.Sleep(500 * time.Millisecond) {
		servers, _ := peer.ListServers(runtimeDir)
		for i := range servers {
			if servers[i].Lists(host) {
				server = &servers[i]
			}
		}
	}
	if server == nil {
		t.Fatalf("the host conversation never linked; agy stderr:\n%s", srvErr.String())
	}
	redact := func(s string) string { return strings.ReplaceAll(s, server.Token, "<TOKEN>") }

	const path = "/exa.language_server_pb.LanguageServerService/SendAgentMessage"
	body := func(recipient, content string) []byte {
		b, _ := json.Marshal(map[string]string{"recipient": recipient, "content": content,
			"senderId": "live-test", "displayTitle": "live test"})
		return b
	}
	request := func(method, p string, payload []byte, token string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, "http://"+server.Address+p, bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("x-codeium-csrf-token", token)
		}
		resp, err := peer.NewAgyClient(30 * time.Second).Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, p, redact(err.Error()))
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
		return resp.StatusCode, redact(strings.TrimSpace(string(b)))
	}

	before := transcript(target)
	for _, tc := range []struct {
		name, method, path string
		payload            []byte
		token              string
		want               int
	}{
		{"no token", http.MethodPost, path, body(target, "LIVE_REFUSAL_NO_TOKEN"), "", http.StatusUnauthorized},
		{"wrong token", http.MethodPost, path, body(target, "LIVE_REFUSAL_WRONG_TOKEN"), "not-the-token", http.StatusUnauthorized},
		{"body not JSON", http.MethodPost, path, []byte(`{"recipient": "` + target + `", LIVE_REFUSAL_MALFORMED`), server.Token, http.StatusBadRequest},
		{"no such method", http.MethodPost, "/exa.language_server_pb.LanguageServerService/NoSuchMethod", body(target, "LIVE_REFUSAL_NO_METHOD"), server.Token, http.StatusNotFound},
		{"GET", http.MethodGet, path, nil, server.Token, http.StatusMethodNotAllowed},
	} {
		status, got := request(tc.method, tc.path, tc.payload, tc.token)
		t.Logf("MEASURED %-17s → HTTP %d: %s", tc.name, status, got)
		if status != tc.want {
			t.Errorf("%s: HTTP %d, want %d", tc.name, status, tc.want)
		}
	}
	// The production path classifies a wrong token as a refusal.
	stale := *server
	stale.Token = "not-the-token"
	_, err = peer.SendAgyServer(peer.NewAgyClient(30*time.Second), stale, peer.AgyMessage{Recipient: target, Content: "LIVE_REFUSAL_SEND"}, peer.RouteAgyOwnRecord)
	var refused *peer.RefusedError
	if !errors.As(err, &refused) {
		t.Errorf("SendAgyServer with a wrong token: err = %v, want a RefusedError", err)
	}
	time.Sleep(20 * time.Second)
	if after := transcript(target); after != before {
		t.Errorf("the target conversation's transcript changed after the refused requests (%d → %d bytes)", len(before), len(after))
	}
	if strings.Contains(transcript(target), "LIVE_REFUSAL") {
		t.Error("a refused request's text reached the target conversation")
	}

	// THE DISCONFIRMING CASE: the same request, valid, answers 200 and the text arrives.
	status, got := request(http.MethodPost, path, body(target, "LIVE_ACCEPTED_7 reply with the single word OK"), server.Token)
	t.Logf("MEASURED %-17s → HTTP %d: %s", "valid", status, got)
	if status != http.StatusOK {
		t.Fatalf("a valid SendAgentMessage answered %d: %s", status, got)
	}
	var arrived bool
	for deadline := time.Now().Add(90 * time.Second); time.Now().Before(deadline) && !arrived; time.Sleep(time.Second) {
		arrived = strings.Contains(transcript(target), "LIVE_ACCEPTED_7")
	}
	if !arrived {
		t.Error("a valid SendAgentMessage answered 200 and its text never reached the target conversation")
	}
}
