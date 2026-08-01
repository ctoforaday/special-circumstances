package cli

import (
	"crypto/x509"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
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
