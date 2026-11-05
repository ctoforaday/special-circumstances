package main

import (
	"bufio"
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
		b, err := os.ReadFile(filepath.Join(runtimeDir, "gray-area", "antigravity-servers", conv+".json"))
		if err != nil {
			return ""
		}
		var s struct{ Address string }
		_ = json.Unmarshal(b, &s)
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
