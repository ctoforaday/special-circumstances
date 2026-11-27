//go:build linux

package peer

import (
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// serve starts a loopback listener in this process answering every request with status and body.
func serve(t *testing.T, network, addr string, status int, body string) string {
	t.Helper()
	ln, err := net.Listen(network, addr)
	if err != nil {
		t.Skipf("cannot listen on %s %s: %v", network, addr, err)
	}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	return ln.Addr().String()
}

const (
	connectMissing = `{"code":"unauthenticated","message":"unauthenticated: missing CSRF token"}`
	httpsRefusal   = "Client sent an HTTP request to an HTTPS server.\n"
)

// THE HTTP PORT IS THE ONE THAT ASKS FOR A TOKEN: of this process's own listeners, the one
// answering a tokenless request with the Connect "missing CSRF token" error.
func TestDaemonAddressFindsTheHTTPPort(t *testing.T) {
	want := serve(t, "tcp", "127.0.0.1:0", http.StatusUnauthorized, connectMissing)
	serve(t, "tcp", "127.0.0.1:0", http.StatusBadRequest, httpsRefusal)
	serve(t, "tcp", "127.0.0.1:0", http.StatusUnauthorized, `{"code":"unauthenticated","message":"unauthenticated: invalid CSRF token"}`)
	got, err := DaemonAddress(NewAgyClient(2*time.Second), os.Getpid())
	if err != nil || got != want {
		t.Fatalf("DaemonAddress = %q, %v; want %q", got, err, want)
	}
}

func TestDaemonAddressRefusesAmbiguity(t *testing.T) {
	serve(t, "tcp", "127.0.0.1:0", http.StatusUnauthorized, connectMissing)
	serve(t, "tcp", "127.0.0.1:0", http.StatusUnauthorized, connectMissing)
	if got, err := DaemonAddress(NewAgyClient(2*time.Second), os.Getpid()); err == nil || !strings.Contains(err.Error(), "2 of pid") {
		t.Errorf("two token-asking ports = %q, %v; want a refusal naming both", got, err)
	}
}

func TestDaemonAddressNoCandidate(t *testing.T) {
	serve(t, "tcp", "127.0.0.1:0", http.StatusBadRequest, httpsRefusal)
	if got, err := DaemonAddress(NewAgyClient(2*time.Second), os.Getpid()); err == nil || !strings.Contains(err.Error(), "0 of pid") {
		t.Errorf("no token-asking port = %q, %v; want a refusal", got, err)
	}
}

func TestDaemonAddressDialsIPv6Loopback(t *testing.T) {
	want := serve(t, "tcp6", "[::1]:0", http.StatusUnauthorized, connectMissing)
	if got, err := DaemonAddress(NewAgyClient(2*time.Second), os.Getpid()); err != nil || got != want {
		t.Fatalf("DaemonAddress = %q, %v; want %q", got, err, want)
	}
}

func TestListenRowsReadsOnlyLoopbackListeners(t *testing.T) {
	raw := []byte(`  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:B119 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 111 1 0 100 0 0 10 0
   1: 00000000:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 222 1 0 100 0 0 10 0
   2: 0100007F:1F91 0100007F:B119 01 00000000:00000000 00:00000000 00000000  1000        0 333 1 0 100 0 0 10 0
   3: 0100007F:1F92 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 444 1 0 100 0 0 10 0
`)
	got := listenRows(raw, map[string]bool{"111": true, "222": true, "333": true})
	if len(got) != 1 || got[0] != "127.0.0.1:45337" {
		t.Errorf("listenRows = %v, want [127.0.0.1:45337] (a wildcard bind, an established socket and another process's inode excluded)", got)
	}
}
