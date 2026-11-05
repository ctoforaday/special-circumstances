package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

const linkConv = "11111111-2222-4333-8444-555555555555"

func agyPreInvocation(n int) []byte {
	return []byte(fmt.Sprintf(`{"conversationId":%q,"invocationNum":%d,"initialNumSteps":1,"workspacePaths":["/w"]}`, linkConv, n))
}

func recordServer(t *testing.T, runtimeDir, addr string) {
	t.Helper()
	env := map[string]string{peer.EnvConversation: linkConv, peer.EnvAddress: addr, peer.EnvToken: "tok"}
	if _, _, err := peer.Register(runtimeDir, func(k string) string { return env[k] }, time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
}

func TestLinkRequestAsksOnlyWhenItWouldHelp(t *testing.T) {
	up := func(string) bool { return true }
	down := func(string) bool { return false }
	linked, stale, none := t.TempDir(), t.TempDir(), t.TempDir()
	recordServer(t, linked, "localhost:1")
	recordServer(t, stale, "localhost:2")
	claude := hookcore.Read("PreInvocation", strings.NewReader(`{"session_id":"s","hook_event_name":"UserPromptSubmit"}`))
	agy := func(n int) hookcore.Request {
		return hookcore.Read("PreInvocation", strings.NewReader(string(agyPreInvocation(n))))
	}

	for _, c := range []struct {
		name       string
		req        hookcore.Request
		runtimeDir string
		telepathy  string
		dial       func(string) bool
		asks       bool
	}{
		{"no record", agy(0), none, "/bin/telepathy", up, true},
		{"recorded server gone", agy(0), stale, "/bin/telepathy", down, true},
		{"recorded server answers", agy(0), linked, "/bin/telepathy", up, false},
		{"not the turn's first step", agy(1), none, "/bin/telepathy", up, false},
		{"no runtime dir: linking would fail", agy(0), "", "/bin/telepathy", up, false},
		{"no telepathy binary to name", agy(0), none, "", up, false},
		{"claude payload", claude, none, "/bin/telepathy", up, false},
	} {
		got := linkRequest(c.req, c.runtimeDir, c.telepathy, c.dial)
		if (got != "") != c.asks {
			t.Errorf("%s: linkRequest = %q, want asks=%v", c.name, got, c.asks)
		}
		if c.asks && !strings.Contains(got, "'/bin/telepathy' link") {
			t.Errorf("%s: the request does not name the command: %q", c.name, got)
		}
		if c.name == "recorded server gone" && !strings.Contains(got, "no longer running") {
			t.Errorf("a stale link is not said to be stale: %q", got)
		}
	}
}

// THE CALL SITE: the PreInvocation hook's reply carries the request as an ephemeral step, and it
// does none of SessionStart's work — that runs once, from Antigravity's own SessionStart. Deleting
// the request from run fails this.
func TestPreInvocationOnlyAsksToLink(t *testing.T) {
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(transcript, []byte(`{"step_index":0}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	was := findTelepathy
	findTelepathy = func() string { return "/opt/gray-area/bin/telepathy" }
	defer func() { findTelepathy = was }()

	in := fmt.Sprintf(`{"conversationId":%q,"invocationNum":0,"workspacePaths":[%s],"transcriptPath":%s}`,
		linkConv, strconv.Quote(dir), strconv.Quote(transcript))
	stdout, stderr, code := call(t, in, "", okStat(10), "-event", "PreInvocation")
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr)
	}
	if rows := readManifest(t, dir, linkConv); len(rows) != 0 {
		t.Errorf("PreInvocation wrote %d session rows; SessionStart's work runs from SessionStart", len(rows))
	}
	var out struct {
		InjectSteps []map[string]any `json:"injectSteps"`
	}
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("stdout %q: %v", stdout, err)
	}
	found := false
	for _, s := range out.InjectSteps {
		if m, ok := s["ephemeralMessage"].(string); ok && strings.Contains(m, "'/opt/gray-area/bin/telepathy' link") {
			found = true
		}
	}
	if !found {
		t.Errorf("no ephemeral link request in %s", stdout)
	}
}
