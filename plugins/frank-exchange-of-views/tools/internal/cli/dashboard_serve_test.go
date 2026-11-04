package cli

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/dashboard"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

// The capability gate is the whole security model: the exact secret path renders, everything else
// is a flat 404. A regression that answered a prefix, a query variant, or "/" would silently open
// the run to the network, so each is pinned.
func TestDashboardHandlerCapabilityGate(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef0123456789abcdef"
	rendered := 0
	h := dashboardHandler(secret, func() string { rendered++; return "<html>DASHBOARD</html>" })

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}

	// The exact secret path: 200 + the rendered page.
	if rec := get("/" + secret); rec.Code != http.StatusOK || rec.Body.String() != "<html>DASHBOARD</html>" {
		t.Errorf("secret path = %d %q, want 200 and the page", rec.Code, rec.Body.String())
	}
	// Everything without the exact secret: 404, and the page is NEVER rendered.
	rendered = 0
	for _, bad := range []string{"/", "/favicon.ico", "/wrongsecret", "/" + secret + "/extra", "/" + secret + "x"} {
		if rec := get(bad); rec.Code != http.StatusNotFound {
			t.Errorf("path %q = %d, want 404 (no secret, no dashboard)", bad, rec.Code)
		}
	}
	if rendered != 0 {
		t.Errorf("the page rendered %d time(s) for non-secret paths — it must never render without the exact URL", rendered)
	}
}

func TestRandTokenIsUnguessableAndUnique(t *testing.T) {
	hex48 := regexp.MustCompile(`^[0-9a-f]{48}$`)
	a, err := randToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := randToken()
	if err != nil {
		t.Fatal(err)
	}
	if !hex48.MatchString(a) {
		t.Errorf("token %q is not 48 hex chars (24 crypto/rand bytes)", a)
	}
	if a == b {
		t.Error("two tokens collided — the secret is not per-invocation random")
	}
}

func TestSelfSignedCertCoversHostAndLoopback(t *testing.T) {
	ip := net.IPv4(192, 168, 4, 55)
	cert, err := selfSignedCert([]net.IP{ip})
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("the generated cert does not parse: %v", err)
	}
	// It authenticates nothing, but TLS still needs the connect IP in the SANs or the handshake
	// name-check fails outright (separate from the untrusted-CA warning the browser shows).
	if leaf.VerifyHostname("192.168.4.55") != nil {
		t.Error("the host IP is not in the cert SANs")
	}
	if leaf.VerifyHostname("127.0.0.1") != nil {
		t.Error("loopback is not in the cert SANs")
	}
}

func TestHostIPv4sSkipsLoopbackAndLinkLocal(t *testing.T) {
	// Can't assert the machine's real addresses, but whatever it returns must be routable IPv4:
	// no loopback (127/8) and no link-local (169.254/16 auto-config noise).
	for _, ip := range hostIPv4s() {
		if ip.To4() == nil {
			t.Errorf("%v is not IPv4", ip)
		}
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			t.Errorf("%v should have been filtered (loopback/link-local)", ip)
		}
	}
}

// THE HELP MUST NOT CONTRADICT THE TRANSPORT. `--serve`'s help once read "over HTTP …
// UNAUTHENTICATED" while serveDashboard has always run ListenAndServeTLS behind a per-run secret.
// Two co-resident statements of one contract, disagreeing on a SECURITY property — the class the
// protocol registry calls co-resident-rules-disagree. A golden pins the STRING; nothing pinned the
// AGREEMENT, so a reviewer regenerating a golden would not notice a security claim flip.
func TestServeHelpMatchesTheTransport(t *testing.T) {
	h := help(t, "dashboard", "--help", "--seat-id", "operator")
	if !strings.Contains(h, "HTTPS") {
		t.Error("--serve help does not say HTTPS, but serveDashboard uses ListenAndServeTLS")
	}
	for _, lie := range []string{"UNAUTHENTICATED", "over HTTP on this port"} {
		if strings.Contains(h, lie) {
			t.Errorf("--serve help claims %q — the server is TLS + secret-gated; the help must not understate it", lie)
		}
	}
	if !strings.Contains(h, "secret") {
		t.Error("--serve help does not mention the per-run secret URL — the only access control there is")
	}
}

// THE SERVED DASHBOARD MAY NOT OUTLIVE ITS RUN. It used to serve a finished run "until killed",
// which on 2026-08-04 left a dashboard exposed on the LAN after its run ended, with nothing to
// close it. Binding is now refused unless the run is live.
func TestServeRefusesWhenTheRunIsNotLive(t *testing.T) {
	dir := recordtest.TmpRun(t)
	prev := dashboardProject
	dashboardProject = t.TempDir() // a project with no marker: this run has no row
	t.Cleanup(func() { dashboardProject = prev })

	err := serveDashboard(io.Discard, runtest.Open(t, dir), 0, func() string { return "<html></html>" })
	if err == nil {
		t.Fatal("serveDashboard bound a socket for a run that is not live")
	}
	if !strings.Contains(err.Error(), "not live") {
		t.Errorf("error = %v, want it to say the run is not live", err)
	}
	if !strings.Contains(err.Error(), "dashboard.html") {
		t.Error("the refusal should point at the static snapshot as the way to read a finished run")
	}

	// ANOTHER RUN'S ROW IS NOT THIS RUN'S. The file alone used to be the test, so a sibling's row
	// kept a finished run's server binding.
	other := t.TempDir()
	liveProjectAt(t, dashboardProject, other)
	if err := serveDashboard(io.Discard, runtest.Open(t, dir), 0, func() string { return "<html></html>" }); err == nil ||
		!strings.Contains(err.Error(), "not live") {
		t.Errorf("a marker naming only another run let this one bind: %v", err)
	}
}

// liveProject makes a project whose live-run marker holds one row, for runDir, and returns the
// project and its marker file.
func liveProject(t *testing.T, runDir string) (proj, marker string) {
	t.Helper()
	proj = t.TempDir()
	return proj, liveProjectAt(t, proj, runDir)
}

func liveProjectAt(t *testing.T, proj, runDir string) string {
	t.Helper()
	marker := filepath.Join(proj, ".claude", "run-live.json")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(map[string]any{"runs": []map[string]string{{"runDir": runDir}}})
	if err := os.WriteFile(marker, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return marker
}

// A VERDICT THE BOARD DERIVES IS NOT THE OUTCOME. On #1677's smoke `--watch` exited on the final
// chair's register, printing that the record held the outcome, 4m17s before the assembly sitting
// recorded it; the terminal and assembly sittings ran unwatched, and the lead repeated the line.
func TestAVerdictTheBoardDerivesDoesNotEndTheWatch(t *testing.T) {
	runDir := newRun(t)
	proj, _ := liveProject(t, runDir)
	recordtest.Seed(t, runDir, recordtest.At(t, "judge", "judge:halt:1",
		&recordpb.Halt{Opinion: proto.String("a boundary the run may not cross")}))
	run := runtest.Open(t, runDir)
	if v, err := record.TerminalVerdict(run); err != nil || v == "" {
		t.Fatalf("the seeded board derives no verdict (%q, %v), so this measures nothing", v, err)
	}
	if over, why := watchIsOver(proj, run); over {
		t.Errorf("a derived verdict with no outcome recorded ended the watch: %q", why)
	}
}

// #270: the watchers must end on ANY signal. The row's only remover is `capture`, which is
// optional — so a killed run used to leave the served dashboard holding a socket, and `--watch`
// regenerating a dead run, forever.
func TestTheWatchIsOverOnEitherSignal(t *testing.T) {
	// A live run: its row present, no outcome on the record. Neither watcher may exit.
	runDir := newRun(t)
	proj, _ := liveProject(t, runDir)
	if over, why := watchIsOver(proj, runtest.Open(t, runDir)); over {
		t.Fatalf("a live run must not end the watch: %q", why)
	}

	// The clean end: capture removed the row.
	over, why := watchIsOver(t.TempDir(), runtest.Open(t, runDir))
	if !over || !strings.Contains(why, "row is gone") {
		t.Errorf("an absent row is the clean end: over=%v why=%q", over, why)
	}

	// The row is no longer live — here, its outcome mark — whatever the record could be read to say.
	markedDir := newRun(t)
	markedProj, _ := liveProject(t, markedDir)
	if err := os.WriteFile(filepath.Join(markedDir, runlive.OutcomeMark), []byte(`{"agentId":""}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if over, why := watchIsOver(markedProj, runtest.Open(t, markedDir)); !over || !strings.Contains(why, "outcome is recorded") {
		t.Errorf("a row that is no longer live ends the watch: over=%v why=%q", over, why)
	}

	// THE CASE THIS EXISTS FOR: the bench recorded the run's outcome and the marker is STILL
	// there, because nothing ran capture. The record is the truthful signal.
	recordtest.Seed(t, runDir, recordtest.At(t, "judge", "judge:outcome:1",
		&recordpb.Outcome{
			Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED),
			Prose:   proto.String("the run ended without the question being answered"),
		}))
	over, why = watchIsOver(proj, runtest.Open(t, runDir))
	if !over {
		t.Fatal("a run whose outcome is on the record leaves nothing to watch, whatever the filesystem says")
	}
	// And it says WHY, because "this run was decided without being captured" is the part worth knowing.
	for _, want := range []string{"UNVERIFIED", "capture has not run"} {
		if !strings.Contains(why, want) {
			t.Errorf("the reason must carry %q so the operator learns capture was skipped: %q", want, why)
		}
	}
}

// lockedBuffer is serveDashboard's output, written on the serving goroutine and read by the test.
type lockedBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

// THE SERVED DASHBOARD, RUN TO ITS OWN TEARDOWN. serveDashboard is the one place the shipped
// binary starts a goroutine of its own, and it does it twice over: the watcher that polls for the
// run's end, and the connection goroutines that render the page from the same record. This serves
// a live run on loopback, reads the page through the real socket while the watcher is polling —
// so a render on a connection goroutine overlaps watchIsOver on the watcher's — then ends the run
// and waits for the server to close itself.
//
// The directive is what puts this test in the race gate and holds it there: scripts/racescope
// refuses a goroutine start in shipped source that no test declares, and refuses this
// declaration if the test stops executing those starts.
//
//racescope:covers internal/cli/dashboard_serve.go:serveDashboard
func TestTheServedDashboardAnswersWhileWatchedAndClosesWhenTheRunEnds(t *testing.T) {
	run := runtest.Open(t, newRun(t))
	proj, marker := liveProject(t, run.Dir())
	prevProject, prevHost, prevInterval := dashboardProject, dashboardBindHost, dashboardWatchInterval
	dashboardProject, dashboardBindHost, dashboardWatchInterval = proj, "127.0.0.1", 2*time.Millisecond
	t.Cleanup(func() {
		dashboardProject, dashboardBindHost, dashboardWatchInterval = prevProject, prevHost, prevInterval
	})
	render := func() string {
		return dashboard.RenderHTML(dashboard.BuildModel(run, "", dashboard.Config{}, float64(time.Now().UnixMilli())))
	}
	// The certificate is minted per invocation and never leaves the process, so there is no root
	// to verify it against; the secret in the path is the only credential this server has.
	client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}}

	// serveDashboard binds the port it is given, so the test has to name a free one, and a port
	// that was free a moment ago can be taken. A bind that loses that race is retried on another.
	var (
		out  *lockedBuffer
		done chan error
		base string
	)
	for attempt := 0; base == ""; attempt++ {
		probe, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		port := probe.Addr().(*net.TCPAddr).Port
		probe.Close()

		out, done = &lockedBuffer{}, make(chan error, 1)
		go func() { done <- serveDashboard(out, run, port, render) }()
		secret := regexp.MustCompile(fmt.Sprintf(`:%d/([0-9a-f]{48})`, port))
		deadline := time.Now().Add(30 * time.Second)
		for base == "" {
			select {
			case err := <-done:
				if attempt < 5 && err != nil && strings.Contains(err.Error(), "address already in use") {
					deadline = time.Time{}
				} else {
					t.Fatalf("serveDashboard returned before the run ended: %v\n%s", err, out.String())
				}
			default:
			}
			if deadline.IsZero() {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("the served dashboard never answered on port %d:\n%s", port, out.String())
			}
			if m := secret.FindStringSubmatch(out.String()); m != nil {
				if resp, err := client.Get(fmt.Sprintf("https://127.0.0.1:%d/%s", port, m[1])); err == nil {
					resp.Body.Close()
					base = fmt.Sprintf("https://127.0.0.1:%d/%s", port, m[1])
					break
				}
			}
			time.Sleep(time.Millisecond)
		}
	}
	// From here the server is up. Whatever happens next, end the run and wait for the watcher
	// before the package variables are put back: it reads them on its own goroutine.
	ended := false
	t.Cleanup(func() {
		if ended {
			return
		}
		// A removal that failed leaves the run live and the watcher running: waiting on it
		// unbounded turned that failure into a job timeout naming no step.
		if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
			t.Errorf("ending the run in cleanup: %v", err)
		}
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Error("cleanup ended the run and the served dashboard is still listening")
		}
	})

	// The page, through the socket, while the watcher polls every 2ms: enough requests that many
	// polls land inside them.
	for i := 0; i < 25; i++ {
		resp, err := client.Get(base)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "generated by feov-record dashboard") {
			t.Fatalf("request %d = %d, %d bytes; want 200 and the rendered page", i, resp.StatusCode, len(body))
		}
	}
	if resp, err := client.Get(base + "x"); err != nil {
		t.Fatal(err)
	} else if resp.Body.Close(); resp.StatusCode != http.StatusNotFound {
		t.Errorf("a path that is not the secret = %d through the socket, want 404", resp.StatusCode)
	}
	// Still serving: the run is live, so nothing may have torn it down.
	select {
	case err := <-done:
		t.Fatalf("the server closed while its run was live: %v", err)
	default:
	}

	// The run ends. The watcher is the only thing that can close the server now.
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		ended = true
		if err != nil {
			t.Errorf("serveDashboard = %v, want a clean return once the run ended", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the run ended and the served dashboard is still listening")
	}
	if !strings.Contains(out.String(), "torn down") {
		t.Errorf("the teardown is not reported:\n%s", out.String())
	}
}
