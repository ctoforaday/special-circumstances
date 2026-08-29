package cli

import (
	"crypto/x509"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

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
	prev := runLiveMarker
	runLiveMarker = filepath.Join(dir, "absent-run-live.json")
	t.Cleanup(func() { runLiveMarker = prev })

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
}

// #270: the watchers must end on EITHER signal. The marker's only remover is `capture`, which is
// optional — so a killed run used to leave the served dashboard holding a socket, and `--watch`
// regenerating a dead run, forever.
func TestRunHasEndedTakesEitherSignal(t *testing.T) {
	// A live run: marker present, no outcome on the record. Neither watcher may exit.
	runDir := newRun(t)
	marker := filepath.Join(recordtest.TmpRun(t), "run-live.json")
	if err := os.WriteFile(marker, []byte(`{"runs":[{"runDir":"x"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if ended, why := runHasEnded(marker, runtest.Open(t, runDir)); ended {
		t.Fatalf("a live run must not end the watch: %q", why)
	}

	// The clean end: capture removed the marker.
	gone := filepath.Join(recordtest.TmpRun(t), "absent.json")
	ended, why := runHasEnded(gone, runtest.Open(t, runDir))
	if !ended || !strings.Contains(why, "marker gone") {
		t.Errorf("an absent marker is the clean end: ended=%v why=%q", ended, why)
	}

	// THE CASE THIS EXISTS FOR: the bench recorded the run's outcome and the marker is STILL
	// there, because nothing ran capture. The record is the truthful signal.
	recordtest.Seed(t, runDir, recordtest.At(t, "judge-terminal", 1, "judge-terminal:outcome:1",
		&recordpb.Outcome{
			Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED),
			Prose:   proto.String("the run ended without the question being answered"),
		}))
	ended, why = runHasEnded(marker, runtest.Open(t, runDir))
	if !ended {
		t.Fatal("a run whose outcome is on the record has ended, whatever the filesystem says")
	}
	// And it says WHY, because "this run ended without being captured" is the part worth knowing.
	for _, want := range []string{"UNVERIFIED", "capture has not run"} {
		if !strings.Contains(why, want) {
			t.Errorf("the reason must carry %q so the operator learns capture was skipped: %q", want, why)
		}
	}
}
