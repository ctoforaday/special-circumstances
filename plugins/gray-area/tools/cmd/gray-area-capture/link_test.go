package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

const linkConv = "11111111-2222-4333-8444-555555555555"

func liveAs(l catalogue.Liveness) func(peer.Server) catalogue.Liveness {
	return func(peer.Server) catalogue.Liveness { return l }
}

func agyPreInvocation(n int) []byte {
	return []byte(fmt.Sprintf(`{"conversationId":%q,"invocationNum":%d,"initialNumSteps":1,"workspacePaths":["/w"]}`, linkConv, n))
}

func recordServer(t *testing.T, runtimeDir, addr string) {
	t.Helper()
	env := map[string]string{peer.EnvConversation: linkConv, peer.EnvAddress: addr, peer.EnvToken: "tok"}
	if _, _, err := peer.RegisterWith(runtimeDir, func(k string) string { return env[k] }, func() peer.ServerIdentity { return peer.ServerIdentity{} }, time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
}

func TestLinkRequestSkipsWhenHostServerAlreadyLinkedByAnotherConversation(t *testing.T) {
	rt := t.TempDir()
	host := peer.ServerIdentity{PID: 7777, StartTime: 88888, Daemon: true}
	otherEnv := map[string]string{peer.EnvConversation: "other-conv", peer.EnvAddress: "localhost:9999", peer.EnvToken: "tok"}
	if _, _, err := peer.RegisterWith(rt, func(k string) string { return otherEnv[k] }, func() peer.ServerIdentity { return host }, time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
	req := hookcore.Read("PreInvocation", strings.NewReader(string(agyPreInvocation(0))))
	if got := linkRequestWithHost(req, rt, "/bin/telepathy", func() peer.ServerIdentity { return host }, liveAs(catalogue.Live)); got != "" {
		t.Errorf("another conv already linked this host server, yet linkRequest asked: %q", got)
	}
	// The host server's record counts only while its server runs: a record of a server that is gone
	// is no reason to stay quiet.
	if got := linkRequestWithHost(req, rt, "/bin/telepathy", func() peer.ServerIdentity { return host }, liveAs(catalogue.Ended)); !strings.HasPrefix(got, firstLinkWhy) {
		t.Errorf("the host server's record reads ended, yet linkRequest = %q", got)
	}
	// Early returns must not invoke hostFn (/proc walk).
	called := false
	reqStep1 := hookcore.Read("PreInvocation", strings.NewReader(string(agyPreInvocation(1))))
	_ = linkRequestWithHost(reqStep1, rt, "/bin/telepathy", func() peer.ServerIdentity { called = true; return host }, liveAs(catalogue.Live))
	if called {
		t.Errorf("hostFn was called on invocationNum=1 early return")
	}
}

func TestLinkRequestAsksOnlyWhenItWouldHelp(t *testing.T) {
	up := liveAs(catalogue.Live)
	down := liveAs(catalogue.Ended)
	unknown := liveAs(catalogue.Unknown)
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
		liveness   func(peer.Server) catalogue.Liveness
		asks       bool
	}{
		{"no record", agy(0), none, "/bin/telepathy", up, true},
		{"recorded server gone", agy(0), stale, "/bin/telepathy", down, true},
		{"recorded server cannot be judged", agy(0), stale, "/bin/telepathy", unknown, true},
		{"recorded server answers", agy(0), linked, "/bin/telepathy", up, false},
		{"not the turn's first step", agy(1), none, "/bin/telepathy", up, false},
		{"no runtime dir: linking would fail", agy(0), "", "/bin/telepathy", up, false},
		{"no telepathy binary to name", agy(0), none, "", up, false},
		{"claude payload", claude, none, "/bin/telepathy", up, false},
	} {
		got := linkRequest(c.req, c.runtimeDir, c.telepathy, c.liveness)
		if (got != "") != c.asks {
			t.Errorf("%s: linkRequest = %q, want asks=%v", c.name, got, c.asks)
		}
		if c.asks && !strings.Contains(got, "'/bin/telepathy' link") {
			t.Errorf("%s: the request does not name the command: %q", c.name, got)
		}
		// A conversation that linked before is told its link is stale; one that never linked is
		// asked plainly.
		want := staleLinkWhy
		if c.runtimeDir == none {
			want = firstLinkWhy
		}
		if c.asks && !strings.HasPrefix(got, want) {
			t.Errorf("%s: linkRequest = %q, want it to open %q", c.name, got, want)
		}
	}
}

// THE REQUEST NEEDS NO RECORD TO NAME A STALE LINK. Prune deletes the record of a server that is
// gone; the conversation's linked marker, which prune leaves, still says it linked before.
func TestLinkRequestAsksAfterTheStaleRecordIsPruned(t *testing.T) {
	rt := t.TempDir()
	env := map[string]string{peer.EnvConversation: linkConv, peer.EnvAddress: "localhost:3", peer.EnvToken: "tok"}
	gone := peer.ServerIdentity{PID: 4321, StartTime: 98765}
	if _, _, err := peer.RegisterWith(rt, func(k string) string { return env[k] }, func() peer.ServerIdentity { return gone }, time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
	if removed, err := peer.Prune(rt, liveAs(catalogue.Ended)); err != nil || len(removed) != 1 {
		t.Fatalf("Prune = %v, %v; want the pid-keyed record removed", removed, err)
	}
	req := hookcore.Read("PreInvocation", strings.NewReader(string(agyPreInvocation(0))))
	got := linkRequestWithHost(req, rt, "/bin/telepathy", func() peer.ServerIdentity { return peer.ServerIdentity{} }, liveAs(catalogue.Live))
	if !strings.HasPrefix(got, staleLinkWhy) || !strings.Contains(got, "'/bin/telepathy' link") {
		t.Errorf("with the stale record pruned, linkRequest = %q; want the request naming the stale link", got)
	}
}

// A CONVERSATION THAT NEVER LINKED IS ASKED PLAINLY, even beside the record of another conversation
// on a server that is gone: the stale wording is for a link this conversation made.
func TestLinkRequestIsPlainForAConversationThatNeverLinked(t *testing.T) {
	rt := t.TempDir()
	env := map[string]string{peer.EnvConversation: "another-conversation", peer.EnvAddress: "localhost:3", peer.EnvToken: "tok"}
	if _, _, err := peer.RegisterWith(rt, func(k string) string { return env[k] }, func() peer.ServerIdentity { return peer.ServerIdentity{PID: 4321, StartTime: 98765} }, time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
	req := hookcore.Read("PreInvocation", strings.NewReader(string(agyPreInvocation(0))))
	got := linkRequestWithHost(req, rt, "/bin/telepathy", func() peer.ServerIdentity { return peer.ServerIdentity{} }, liveAs(catalogue.Ended))
	if !strings.HasPrefix(got, firstLinkWhy) || !strings.Contains(got, "'/bin/telepathy' link") {
		t.Errorf("a never-linked conversation: linkRequest = %q; want the plain request", got)
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

// THE CALL SITE JUDGES A RECORD BY ITS PROCESS. A record of a server that is running — here this
// test process, by its real pid, start time and pid domain — stops the request from the hook's
// own entry point; a record judged any other way would ask.
func TestPreInvocationDoesNotAskWhileTheConversationsServerRuns(t *testing.T) {
	if runtime.GOOS != "linux" || catalogue.LocalPidDomain() == "" {
		t.Skip("server liveness is measured on Linux with a machine id only")
	}
	start, ok := peer.ProcessStartTime(os.Getpid())
	if !ok {
		t.Fatal("this process's start time is unmeasurable")
	}
	rt := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", rt)
	env := map[string]string{peer.EnvConversation: linkConv, peer.EnvAddress: "localhost:4", peer.EnvToken: "tok"}
	me := peer.ServerIdentity{PID: os.Getpid(), StartTime: start}
	if _, _, err := peer.RegisterWith(rt, func(k string) string { return env[k] }, func() peer.ServerIdentity { return me }, time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
	was := findTelepathy
	findTelepathy = func() string { return "/opt/gray-area/bin/telepathy" }
	defer func() { findTelepathy = was }()
	dir := t.TempDir()
	in := fmt.Sprintf(`{"conversationId":%q,"invocationNum":0,"workspacePaths":[%s]}`, linkConv, strconv.Quote(dir))
	stdout, stderr, code := call(t, in, "", okStat(10), "-event", "PreInvocation")
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr)
	}
	if strings.Contains(stdout, "telepathy' link") {
		t.Errorf("the conversation's server is running and recorded, yet the hook asked to link: %s", stdout)
	}
}

// THE TWO OPENINGS ARE MEASURED SENTENCES, pinned word for word: the plain one is what a fresh
// conversation acted on, and the stale one is what a conversation whose history showed `linked`
// acted on after the plain one was ignored.
// WITHOUT A PID DOMAIN NO RECORD READS LIVE, so the hook does not ask: it would ask on every turn.
func TestLinkRequestNeverAsksWithoutAPidDomain(t *testing.T) {
	was := localPidDomain
	defer func() { localPidDomain = was }()
	req := hookcore.Read("PreInvocation", strings.NewReader(string(agyPreInvocation(0))))
	rt := t.TempDir()
	if got := linkRequestWithHost(req, rt, "/bin/telepathy", func() peer.ServerIdentity { return peer.ServerIdentity{} }, liveAs(catalogue.Ended)); got == "" {
		t.Fatal("with a pid domain, an unlinked conversation is not asked; the control is broken")
	}
	localPidDomain = func() string { return "" }
	if got := linkRequestWithHost(req, rt, "/bin/telepathy", func() peer.ServerIdentity { return peer.ServerIdentity{} }, liveAs(catalogue.Ended)); got != "" {
		t.Errorf("with no pid domain, linkRequest = %q, want no request", got)
	}
}

func TestLinkRequestWordings(t *testing.T) {
	if firstLinkWhy != "Other agents cannot yet wake this conversation with `telepathy send`." {
		t.Errorf("the first-link opening changed: %q", firstLinkWhy)
	}
	if staleLinkWhy != "This conversation's earlier `telepathy link` points at a language server that is no longer running, so other agents cannot wake it." {
		t.Errorf("the stale-link opening changed: %q", staleLinkWhy)
	}
}
