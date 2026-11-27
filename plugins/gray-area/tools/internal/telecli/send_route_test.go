package telecli

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

// The servers of the incident on 2026-11-26: the daemon before the 03:00 restart, the daemon after
// it, and two standalone language servers.
var (
	oldDaemon = peer.ServerIdentity{PID: 266530, StartTime: 1000, Daemon: true}
	newDaemon = peer.ServerIdentity{PID: 284285, StartTime: 30000, Daemon: true}
	serverS1  = peer.ServerIdentity{PID: 5101, StartTime: 5100}
	serverS2  = peer.ServerIdentity{PID: 5202, StartTime: 5200}
)

// countingServer stands in a language server answering status, and counts the requests it gets.
func countingServer(t *testing.T, status int) (string, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		if status != http.StatusOK {
			http.Error(rw, "boom", status)
			return
		}
		rw.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://"), &hits
}

// unreachable is an address nothing listens on: a send to it fails to connect, so it never arrives.
func unreachable(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

// hangingServer stands in a language server that takes a request and never answers it, and
// counts the requests it gets.
func hangingServer(t *testing.T) (string, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
		case <-done:
		}
	}))
	// Cleanups run last-registered first: the handler is released before Close waits for it.
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(done) })
	return strings.TrimPrefix(srv.URL, "http://"), &hits
}

// linkAs records that conv linked on the server id at addr, as `telepathy link` from its shell does.
func (w *sendWorld) linkAs(t *testing.T, conv, addr string, id peer.ServerIdentity, at time.Time) {
	t.Helper()
	env := map[string]string{peer.EnvConversation: conv, peer.EnvAddress: addr, peer.EnvToken: "tok"}
	if _, _, err := peer.RegisterWith(w.runtime, func(k string) string { return env[k] }, func() peer.ServerIdentity { return id }, at); err != nil {
		t.Fatal(err)
	}
}

// running makes exactly the servers given read Live, and every other record Ended.
func (w *sendWorld) running(ids ...peer.ServerIdentity) {
	w.env.RecordLiveness = func(s peer.Server) catalogue.Liveness {
		for _, id := range ids {
			if s.PID == id.PID && s.StartTime == id.StartTime {
				return catalogue.Live
			}
		}
		return catalogue.Ended
	}
}

// holders answers ConversationBusy by the pid excluded; unlisted exclusions find nobody.
func (w *sendWorld) holders(byExclude map[int]peer.Holder) *[]int {
	var probes []int
	w.env.ConversationBusy = func(_, _ string, excludePID int) peer.Holder {
		probes = append(probes, excludePID)
		return byExclude[excludePID]
	}
	return &probes
}

func dbHolder(pid int) peer.Holder {
	return peer.Holder{PID: pid, Kind: peer.HolderDBOpen, Detail: "db open"}
}

func (w *sendWorld) sendFromClaude(t *testing.T) string {
	t.Helper()
	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}
	out, errOut, code := w.run(t, "send", agyPeer, "hello")
	if code != 0 {
		t.Fatalf("send: exit %d, stderr %q", code, errOut)
	}
	return out
}

func wantQueued(t *testing.T, w *sendWorld, out string, cause string) {
	t.Helper()
	if !strings.Contains(out, "QUEUED for antigravity conversation "+agyPeer) || !strings.Contains(out, "not delivered live: "+cause) {
		t.Errorf("stdout = %q\nwant QUEUED with the cause %q", out, cause)
	}
	if n := len(w.mailbox(t)); n != 1 {
		t.Errorf("%d undelivered markers, want 1", n)
	}
}

// 16:05 ON 2026-11-26. The conversation's own record is the previous daemon's, whose process is
// gone; the current daemon holds the conversation and has not been linked. The message queues
// naming the unlinked daemon, and the current daemon is never called a second writer.
func TestSendIncidentAt1605Queues(t *testing.T) {
	w := newSendWorld(t)
	w.linkAs(t, agyPeer, "127.0.0.1:1", oldDaemon, frozen.Add(-13*time.Hour))
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(newDaemon)
	w.holders(map[int]peer.Holder{0: dbHolder(newDaemon.PID)})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "the daemon (started ")
	if !strings.Contains(out, "has not been linked since it started") || strings.Contains(out, "second writer") {
		t.Errorf("stdout = %q\nwant the unlinked daemon named, and no second writer", out)
	}
}

// FROM 16:07:22 the daemon has a record, written by another conversation that linked on it, and it
// delivers: the holder is the current daemon, through its own record.
func TestSendIncidentAfterLinkUsesRoute2(t *testing.T) {
	w := newSendWorld(t)
	addr, hits := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, "127.0.0.1:1", oldDaemon, frozen.Add(-13*time.Hour))
	w.linkAs(t, agySelf, addr, newDaemon, frozen.Add(-time.Minute))
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(newDaemon)
	probes := w.holders(map[int]peer.Holder{0: dbHolder(newDaemon.PID)})
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "delivered to antigravity conversation "+agyPeer+" via antigravity-daemon, the language server at "+addr) {
		t.Errorf("stdout = %q", out)
	}
	if hits.Load() != 1 {
		t.Errorf("the daemon got %d requests, want 1", hits.Load())
	}
	if got := *probes; len(got) != 2 || got[0] != 0 || got[1] != newDaemon.PID {
		t.Errorf("probes excluded %v, want [0 %d]: discovery, then the sole-holder gate", got, newDaemon.PID)
	}
}

// THE DAEMON'S RECORD IS NAMED AS THE DAEMON'S even when it also lists the conversation.
func TestSendRouteNamesTheDaemonOverItsListing(t *testing.T) {
	w := newSendWorld(t)
	addr, _ := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, addr, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "via antigravity-daemon, the language server at "+addr) {
		t.Errorf("stdout = %q", out)
	}
}

// THE HOLDER DELIVERS, NOT THE NEWEST LINK. S1 is the record the conversation linked to last, but
// S2 has it open, through a record that does not list it.
func TestSendHolderIsAnotherLiveRecord(t *testing.T) {
	w := newSendWorld(t)
	a1, h1 := countingServer(t, http.StatusOK)
	a2, h2 := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, a2, serverS2, frozen.Add(-time.Hour))
	w.running(serverS1, serverS2)
	w.holders(map[int]peer.Holder{0: dbHolder(serverS2.PID)})
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "via antigravity-server-holder, the language server at "+a2) {
		t.Errorf("stdout = %q", out)
	}
	if h1.Load() != 0 || h2.Load() != 1 {
		t.Errorf("S1 got %d requests and S2 %d; want 0 and 1", h1.Load(), h2.Load())
	}
}

// AFTER THE UPGRADE DROP no record lists the conversation; the server holding it still delivers.
func TestSendSoleHolderWithNoListingRecord(t *testing.T) {
	w := newSendWorld(t)
	a2, h2 := countingServer(t, http.StatusOK)
	w.linkAs(t, agySelf, a2, serverS2, frozen)
	w.running(serverS2)
	w.holders(map[int]peer.Holder{0: dbHolder(serverS2.PID)})
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "via antigravity-server-holder, the language server at "+a2) || h2.Load() != 1 {
		t.Errorf("stdout = %q, S2 requests %d", out, h2.Load())
	}
	if strings.Contains(out, "also by") {
		t.Errorf("a sole holder was reported with a second one: %q", out)
	}
}

// TWO SERVERS HOLD IT: neither may write, and neither is a second writer to the other's name.
func TestSendTwoHoldersQueues(t *testing.T) {
	w := newSendWorld(t)
	a1, h1 := countingServer(t, http.StatusOK)
	a2, h2 := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, a2, serverS2, frozen)
	w.running(serverS1, serverS2)
	w.holders(map[int]peer.Holder{
		0:            dbHolder(serverS1.PID),
		serverS1.PID: dbHolder(serverS2.PID),
		serverS2.PID: dbHolder(serverS1.PID),
	})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "held by the language server recorded at "+a1+" and also by pid 5202")
	if strings.Contains(out, "second writer") || h1.Load()+h2.Load() != 0 {
		t.Errorf("stdout = %q, requests %d+%d; want no second-writer line and no request", out, h1.Load(), h2.Load())
	}
}

// A PID FROM ANOTHER NAMESPACE NEVER EXCLUDES A LOCAL PROCESS. The record's pid number equals the
// local holder's, but its domain is foreign, so the real liveness reads it Unknown.
func TestSendForeignDomainPidNeverExcluded(t *testing.T) {
	w := newSendWorld(t)
	addr, hits := countingServer(t, http.StatusOK)
	foreign := peer.Server{Shape: peer.RecordShape, Key: "pid-6060-600", PID: 6060, StartTime: 600,
		PidDomain: "linux:another-machine:pid:[1]", Address: addr, Token: "tok", RegisteredAt: frozen.Unix(),
		Conversations: []peer.LinkedConversation{{ID: agyPeer, LinkedAt: frozen.Unix()}}}
	b, _ := json.Marshal(foreign)
	dir := peer.RegistryDir(w.runtime)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, foreign.Key+".json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	w.env.RecordLiveness = peer.RecordLiveness
	w.holders(map[int]peer.Holder{0: dbHolder(6060)})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "held by another process, pid 6060; sending live would create a second writer")
	if hits.Load() != 0 {
		t.Errorf("the foreign record's server got %d requests", hits.Load())
	}
}

func TestSendUnrelatedHolderQueues(t *testing.T) {
	w := newSendWorld(t)
	addr, hits := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, addr, serverS1, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1)
	w.holders(map[int]peer.Holder{0: dbHolder(9999)})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "held by another process, pid 9999; sending live would create a second writer")
	if hits.Load() != 0 {
		t.Errorf("a server got %d requests while an unrelated process held the conversation", hits.Load())
	}
}

func TestSendPresenceLockHolderNamed(t *testing.T) {
	w := newSendWorld(t)
	w.holders(map[int]peer.Holder{0: {PID: 4242, Kind: peer.HolderPresenceLock, Detail: "pid 4242 holds presence/x.lock"}})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "held by another process, pid 4242; sending live would create a second writer")
}

func TestSendUninspectableQueues(t *testing.T) {
	w := newSendWorld(t)
	w.holders(map[int]peer.Holder{0: {Kind: peer.HolderUninspectable, Detail: "cannot read /proc to verify no second writer: denied"}})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "cannot inspect whether another process holds the conversation: cannot read /proc to verify no second writer: denied")
	if strings.Contains(out, "pid 0") {
		t.Errorf("stdout names pid 0: %q", out)
	}
}

// A CONVERSATION TAKEN BEFORE THE FALLBACK ENDS THE SEND. Target 1 cannot be reached, and by the
// probe before the daemon a process holds the conversation; the daemon is never called.
func TestSendReprobesBeforeDaemonFallback(t *testing.T) {
	w := newSendWorld(t)
	a1 := unreachable(t)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	w.holders(map[int]peer.Holder{newDaemon.PID: dbHolder(serverS1.PID)})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "held by the language server recorded at "+a1+" (pid 5101)")
	if hd.Load() != 0 {
		t.Errorf("the daemon got %d requests after target 1 took the conversation", hd.Load())
	}
}

func TestSendGateUninspectableNoAlso(t *testing.T) {
	w := newSendWorld(t)
	a2, h2 := countingServer(t, http.StatusOK)
	w.linkAs(t, agySelf, a2, serverS2, frozen)
	w.running(serverS2)
	w.holders(map[int]peer.Holder{
		0:            dbHolder(serverS2.PID),
		serverS2.PID: {Kind: peer.HolderUninspectable, Detail: "cannot inspect /proc/77/fd"},
	})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "cannot inspect whether another process holds the conversation: cannot inspect /proc/77/fd")
	if strings.Contains(out, "also by") || h2.Load() != 0 {
		t.Errorf("stdout = %q, requests %d", out, h2.Load())
	}
}

// THE HOLDER IS NEVER BYPASSED. X has the conversation open and cannot be reached; the daemon,
// which has a record, is not tried.
func TestSendHolderUnreachedQueues(t *testing.T) {
	w := newSendWorld(t)
	ax := unreachable(t)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, ax, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	w.holders(map[int]peer.Holder{0: dbHolder(serverS1.PID)})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "the language server recorded at "+ax+" cannot be reached")
	if hd.Load() != 0 {
		t.Errorf("the daemon got %d requests while X held the conversation", hd.Load())
	}
}

// wantUnknown asserts a send that ended with its delivery unknown: a non-zero exit saying so, and
// nothing in the mailbox.
func wantUnknown(t *testing.T, w *sendWorld, cause string) string {
	t.Helper()
	w.callerEnv = map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}
	out, errOut, code := w.run(t, "send", agyPeer, "hello")
	if code == 0 || out != "" {
		t.Errorf("send: exit %d, stdout %q; want a non-zero exit and nothing printed as delivered or queued", code, out)
	}
	if !strings.Contains(errOut, "delivery UNKNOWN: "+cause) || !strings.Contains(errOut, "may have arrived") {
		t.Errorf("stderr = %q\nwant delivery UNKNOWN with the cause %q", errOut, cause)
	}
	if n := len(w.mailbox(t)); n != 0 {
		t.Errorf("%d undelivered markers, want none: a queued copy could arrive a second time", n)
	}
	return errOut
}

// A SERVER THAT WAS SENT THE MESSAGE AND DID NOT ANSWER MAY HAVE IT. Target 1 takes the request and
// never answers; the send stops there, with no second server and no mailbox file.
func TestSendTimeoutAfterSendIsUnknown(t *testing.T) {
	w := newSendWorld(t)
	w.env.SendTimeout = 200 * time.Millisecond
	a1, h1 := hangingServer(t)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	probes := w.holders(nil)
	wantUnknown(t, w, "the language server recorded at "+a1+" was sent the message and did not confirm it")
	if h1.Load() != 1 || hd.Load() != 0 {
		t.Errorf("requests: target 1 %d, daemon %d; want 1 and 0", h1.Load(), hd.Load())
	}
	for _, p := range *probes {
		if p == newDaemon.PID {
			t.Errorf("probes %v: the daemon was probed as a next target", *probes)
		}
	}
}

// AN ERROR STATUS IS AN ANSWER FROM A SERVER THAT HAD THE MESSAGE: unknown, not a fallback.
func TestSendErrorStatusIsUnknown(t *testing.T) {
	w := newSendWorld(t)
	a1, h1 := countingServer(t, http.StatusInternalServerError)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	errOut := wantUnknown(t, w, "the language server recorded at "+a1+" answered HTTP 500: boom")
	if h1.Load() != 1 || hd.Load() != 0 {
		t.Errorf("requests: target 1 %d, daemon %d; want 1 and 0", h1.Load(), hd.Load())
	}
	assertGolden(t, "send-unknown", strings.ReplaceAll(errOut, a1, "<S1-ADDR>"))
}

// THE HOLDER'S ERROR IS UNKNOWN TOO: X has the conversation open and answers 500.
func TestSendHolderErrorStatusIsUnknown(t *testing.T) {
	w := newSendWorld(t)
	ax, _ := countingServer(t, http.StatusInternalServerError)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, ax, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	w.holders(map[int]peer.Holder{0: dbHolder(serverS1.PID)})
	wantUnknown(t, w, "the language server recorded at "+ax+" answered HTTP 500")
	if hd.Load() != 0 {
		t.Errorf("the daemon got %d requests while X held the conversation", hd.Load())
	}
}

// THE DAEMON'S ERROR IS UNKNOWN TOO. With no record of its own, the conversation's only target is the
// daemon, which answers 500: no mailbox file, a non-zero exit.
func TestSendDaemonErrorStatusIsUnknown(t *testing.T) {
	w := newSendWorld(t)
	ad, hd := countingServer(t, http.StatusInternalServerError)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(newDaemon)
	wantUnknown(t, w, "the language server recorded at "+ad+" answered HTTP 500")
	if hd.Load() != 1 {
		t.Errorf("the daemon got %d requests, want 1", hd.Load())
	}
}

// AFTER A FALLBACK, THE DAEMON'S SILENCE IS UNKNOWN. Target 1 cannot be reached, so the daemon is
// tried; it takes the request and never answers.
func TestSendDaemonTimeoutAfterFallbackIsUnknown(t *testing.T) {
	w := newSendWorld(t)
	w.env.SendTimeout = 200 * time.Millisecond
	a1 := unreachable(t)
	ad, hd := hangingServer(t)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	wantUnknown(t, w, "the language server recorded at "+ad+" was sent the message and did not confirm it")
	if hd.Load() != 1 {
		t.Errorf("the daemon got %d requests, want 1", hd.Load())
	}
}

// rateLimited stands in a language server that answers 429 to its first n requests, with the given
// Retry-After header when it is not empty, and 200 after; it counts the requests.
func rateLimited(t *testing.T, n int32, retryAfter string) (string, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) <= n {
			if retryAfter != "" {
				rw.Header().Set("Retry-After", retryAfter)
			}
			http.Error(rw, "rate limited", http.StatusTooManyRequests)
			return
		}
		rw.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://"), &hits
}

// rateLimitWorld links agyPeer to S1 at a1 with a running daemon at another server, and records
// the waits send takes.
func (w *sendWorld) rateLimitWorld(t *testing.T, a1 string) (*[]time.Duration, *atomic.Int32) {
	t.Helper()
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	var waits []time.Duration
	w.env.Sleep = func(d time.Duration) { waits = append(waits, d) }
	return &waits, hd
}

// A 429 DECLINED WITHOUT ACTING: the same server is asked again after the wait it named, and
// delivers; the probe runs again before the second request.
func TestSendRateLimitedThenDelivers(t *testing.T) {
	w := newSendWorld(t)
	a1, h1 := rateLimited(t, 1, "2")
	waits, hd := w.rateLimitWorld(t, a1)
	probes := w.holders(nil)
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "via antigravity-server-own, the language server at "+a1) || h1.Load() != 2 || hd.Load() != 0 {
		t.Errorf("stdout = %q, requests %d to S1 and %d to the daemon; want delivery by S1 on its second request", out, h1.Load(), hd.Load())
	}
	if fmt.Sprint(*waits) != "[2s]" {
		t.Errorf("waits %v, want [2s]", *waits)
	}
	if fmt.Sprint(*probes) != fmt.Sprint([]int{0, serverS1.PID, serverS1.PID}) {
		t.Errorf("probes %v, want discovery, then S1's before each request", *probes)
	}
}

// RATE LIMITED EVERY TIME: after rateLimitAttempts requests the message queues, naming the 429s,
// and no other server is tried. With no Retry-After the wait is a second.
func TestSendRateLimitedQueues(t *testing.T) {
	w := newSendWorld(t)
	a1, h1 := rateLimited(t, 99, "")
	waits, hd := w.rateLimitWorld(t, a1)
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, fmt.Sprintf("the language server recorded at %s answered 429 (rate limited) 3 times", a1))
	if h1.Load() != 3 || hd.Load() != 0 {
		t.Errorf("requests %d to S1 and %d to the daemon, want 3 and 0", h1.Load(), hd.Load())
	}
	if fmt.Sprint(*waits) != "[1s 1s]" {
		t.Errorf("waits %v, want [1s 1s]", *waits)
	}
}

// A LONG Retry-After IS CAPPED.
func TestSendRateLimitWaitIsCapped(t *testing.T) {
	w := newSendWorld(t)
	a1, _ := rateLimited(t, 1, "120")
	waits, _ := w.rateLimitWorld(t, a1)
	w.sendFromClaude(t)
	if fmt.Sprint(*waits) != "[5s]" {
		t.Errorf("waits %v, want [5s]", *waits)
	}
}

// THE CONVERSATION TAKEN DURING THE WAIT ENDS THE SEND: the probe before the retry runs after the
// wait, the window in which a holder can appear, and finds one.
func TestSendRateLimitedReprobes(t *testing.T) {
	w := newSendWorld(t)
	a1, h1 := rateLimited(t, 1, "")
	_, hd := w.rateLimitWorld(t, a1)
	var events []string
	w.env.Sleep = func(time.Duration) { events = append(events, "sleep") }
	w.env.ConversationBusy = func(_, _ string, excludePID int) peer.Holder {
		events = append(events, fmt.Sprintf("probe(%d)", excludePID))
		if len(events) == 4 { // discovery, the probe before request 1, the wait, then the retry's probe
			return dbHolder(9999)
		}
		return peer.Holder{}
	}
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "held by another process, pid 9999")
	if h1.Load() != 1 || hd.Load() != 0 {
		t.Errorf("requests %d to S1 and %d to the daemon, want 1 and 0", h1.Load(), hd.Load())
	}
	if want := fmt.Sprintf("[probe(0) probe(%d) sleep probe(%d)]", serverS1.PID, serverS1.PID); fmt.Sprint(events) != want {
		t.Errorf("events %v, want %s", events, want)
	}
}

// A SERVER THAT ANSWERED KEEPS THE SEND. S1 answers 429 and is gone by the retry; the message
// queues, and the daemon is never tried.
func TestSendRateLimitedThenUnreachedQueues(t *testing.T) {
	w := newSendWorld(t)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(rw, "rate limited", http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)
	a1 := strings.TrimPrefix(srv.URL, "http://")
	_, hd := w.rateLimitWorld(t, a1)
	w.env.Sleep = func(time.Duration) { srv.Close() }
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "the language server recorded at "+a1+" cannot be reached")
	if hits.Load() != 1 || hd.Load() != 0 {
		t.Errorf("requests %d to S1 and %d to the daemon, want 1 and 0", hits.Load(), hd.Load())
	}
}

// A RETRY OPENS A NEW CONNECTION: the one the 429 came on may be closed during the wait, and a
// request written into it would fail as one that may have arrived.
func TestSendRateLimitRetryOnANewConnection(t *testing.T) {
	w := newSendWorld(t)
	var hits, conns atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) == 1 {
			http.Error(rw, "rate limited", http.StatusTooManyRequests)
			return
		}
		rw.Write([]byte("{}"))
	}))
	srv.Config.ConnState = func(_ net.Conn, st http.ConnState) {
		if st == http.StateNew {
			conns.Add(1)
		}
	}
	srv.Start()
	t.Cleanup(srv.Close)
	a1 := strings.TrimPrefix(srv.URL, "http://")
	w.rateLimitWorld(t, a1)
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "delivered") || hits.Load() != 2 || conns.Load() != 2 {
		t.Errorf("stdout %q, %d requests on %d connections; want delivery by the second request on a second connection", out, hits.Load(), conns.Load())
	}
}

// THE HOLDER'S 429s QUEUE TOO, and never pass the message to another server.
func TestSendHolderRateLimitedQueues(t *testing.T) {
	w := newSendWorld(t)
	ax, hx := rateLimited(t, 99, "")
	_, hd := w.rateLimitWorld(t, ax)
	w.holders(map[int]peer.Holder{0: dbHolder(serverS1.PID)})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, fmt.Sprintf("the language server recorded at %s answered 429 (rate limited) 3 times", ax))
	if hx.Load() != 3 || hd.Load() != 0 {
		t.Errorf("requests %d to the holder and %d to the daemon, want 3 and 0", hx.Load(), hd.Load())
	}
}

// csrfRefusal stands in a language server whose CSRF check refuses every request, in agy 1.5.18's
// measured shape, and counts the requests.
func csrfRefusal(t *testing.T) (string, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		rw.WriteHeader(http.StatusUnauthorized)
		io.WriteString(rw, `{"code":"unauthenticated","message":"unauthenticated: invalid CSRF token"}`)
	}))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://"), &hits
}

// THE DAEMON'S REFUSAL QUEUES: it is the only target, and its CSRF check refuses.
func TestSendDaemonRefusedQueues(t *testing.T) {
	w := newSendWorld(t)
	ad, hd := csrfRefusal(t)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(newDaemon)
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "the language server recorded at "+ad+" refused the message before acting on it, HTTP 401")
	if hd.Load() != 1 {
		t.Errorf("the daemon got %d requests, want 1", hd.Load())
	}
}

// deliver's client follows no redirect: a server that read the message and redirects to a port
// nobody listens on leaves the delivery UNKNOWN, and the daemon is never tried.
func TestSendFollowsNoRedirect(t *testing.T) {
	w := newSendWorld(t)
	closed := unreachable(t)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		http.Redirect(rw, r, "http://"+closed+r.URL.Path, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(srv.Close)
	a1 := strings.TrimPrefix(srv.URL, "http://")
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	wantUnknown(t, w, "the language server recorded at "+a1+" answered HTTP 307")
	if hits.Load() != 1 || hd.Load() != 0 {
		t.Errorf("requests %d to S1 and %d to the daemon, want 1 and 0", hits.Load(), hd.Load())
	}
}

// A REFUSAL BEFORE HANDLING DID NOT DELIVER: the message queues, naming it, and no other server is
// tried.
func TestSendRefusedQueues(t *testing.T) {
	w := newSendWorld(t)
	a1, h1 := csrfRefusal(t)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "the language server recorded at "+a1+" refused the message before acting on it, HTTP 401")
	if h1.Load() != 1 || hd.Load() != 0 {
		t.Errorf("requests %d to S1 and %d to the daemon, want 1 and 0", h1.Load(), hd.Load())
	}
}

// A SERVER THAT CANNOT BE REACHED NEVER GOT THE MESSAGE, so the daemon delivers it.
func TestSendDialFailureFallsBack(t *testing.T) {
	w := newSendWorld(t)
	a1 := unreachable(t)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agyPeer, a1, serverS1, frozen)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.running(serverS1, newDaemon)
	out := w.sendFromClaude(t)
	if !strings.Contains(out, "via antigravity-daemon, the language server at "+ad) || hd.Load() != 1 {
		t.Errorf("stdout = %q, daemon requests %d; want delivery through the daemon", out, hd.Load())
	}
	if n := len(w.mailbox(t)); n != 0 {
		t.Errorf("%d undelivered markers, want none", n)
	}
}

// THE DAEMON HOLDS IT AND ITS RECORD CANNOT BE JUDGED: not a second writer, and not "unlinked".
func TestSendDaemonHolderWithUnknownRecordQueues(t *testing.T) {
	w := newSendWorld(t)
	ad, hd := countingServer(t, http.StatusOK)
	w.linkAs(t, agySelf, ad, newDaemon, frozen)
	w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
	w.env.RecordLiveness = func(peer.Server) catalogue.Liveness { return catalogue.Unknown }
	w.holders(map[int]peer.Holder{0: dbHolder(newDaemon.PID)})
	out := w.sendFromClaude(t)
	wantQueued(t, w, out, "the daemon's server record reads unknown, and only a record whose server is running is used")
	if strings.Contains(out, "second writer") || hd.Load() != 0 {
		t.Errorf("stdout = %q, daemon requests %d", out, hd.Load())
	}
}

// THE CALL SITE OF PRUNE: any verb, from a shell that is not Antigravity's, deletes the record of a
// server that is gone.
func TestEveryVerbPrunesEndedRecords(t *testing.T) {
	w := newSendWorld(t)
	w.linkAs(t, agyPeer, "127.0.0.1:1", oldDaemon, frozen)
	w.running()
	w.callerEnv = map[string]string{}
	w.run(t, "send", "c1a0", "anything") // fails (no inbox listening); the prune is the point
	if s, err := peer.LookupServerByKey(w.runtime, peer.ServerKey(oldDaemon.PID, oldDaemon.StartTime, "")); s != nil || err != nil {
		t.Errorf("after a verb ran, the Ended record = %+v, %v; want it pruned", s, err)
	}
}

// THE CALL SITE OF PRUNE IN LINK, which records and prunes itself.
func TestLinkPrunesEndedRecords(t *testing.T) {
	w := newSendWorld(t)
	w.linkAs(t, agySelf, "127.0.0.1:1", oldDaemon, frozen)
	w.running(newDaemon)
	w.env.HostServer = func() peer.ServerIdentity { return newDaemon }
	w.callerEnv = map[string]string{peer.EnvConversation: agyPeer, peer.EnvAddress: "localhost:9", peer.EnvToken: "tok"}
	if _, errOut, code := w.run(t, "link"); code != 0 {
		t.Fatalf("link: exit %d, stderr %q", code, errOut)
	}
	if s, _ := peer.LookupServerByKey(w.runtime, peer.ServerKey(oldDaemon.PID, oldDaemon.StartTime, "")); s != nil {
		t.Errorf("link left the Ended record in place")
	}
	if s, _ := peer.LookupServerByKey(w.runtime, peer.ServerKey(newDaemon.PID, newDaemon.StartTime, "")); s == nil {
		t.Errorf("link's own record is missing")
	}
}

func TestPruneFailureIsOneLine(t *testing.T) {
	w := newSendWorld(t)
	// A registry directory whose lock cannot be created: the lock path is a directory.
	if err := os.MkdirAll(filepath.Join(peer.RegistryDir(w.runtime), ".register.lock", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	_, errOut, _ := w.run(t, "send", agyPeer, "hello")
	if strings.Count(errOut, "telepathy: server records not pruned: ") != 1 {
		t.Errorf("stderr = %q, want one 'server records not pruned' line", errOut)
	}
}
