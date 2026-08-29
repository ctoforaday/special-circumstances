package cli

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// serveDashboard hosts the dashboard over HTTPS as an EPHEMERAL, CAPABILITY-GATED, self-tearing
// server. Three properties the operator asked for, each defending a different exposure:
//
//   - A per-run SECRET in the URL path (crypto/rand). The server answers ONLY the exact secret
//     path; every other request — a guess, a scan, a bare "/" — gets a flat 404. Knowledge of the
//     URL IS the access grant, so the run is invisible without it.
//   - HTTPS with an EPHEMERAL self-signed cert (generated here, never written to disk). It exists
//     to encrypt the wire — the secret and the run's internals do not cross the LAN in cleartext —
//     not to prove identity, so a browser will warn on the untrusted cert. That is expected.
//   - SELF-TEARDOWN keyed to the .run-live marker: once the run ends (run-capture removes the
//     marker) the socket closes on its own. A dashboard that outlives its run is an exposure with
//     no owner. The server therefore REFUSES TO BIND when the marker is already absent: its
//     lifetime is bounded by the run's at both ends, not just the tail. A finished run is reviewed
//     through the written snapshot, which needs no listening socket.
//
// render is the per-request page (rendered fresh from the record, no materialized file).
func serveDashboard(out io.Writer, run record.Run, port int, render func() string) error {
	// THE SERVER LIVES ONLY WHILE THE RUN DOES. Refuse to start when the run is not live, rather
	// than serving a finished run "until killed" — that mode left a dashboard exposed on the LAN
	// with nothing to end it, which is precisely the exposure-with-no-owner the teardown exists to
	// prevent (observed 2026-08-04: a server outlived its run and had to be killed by hand). A
	// completed run is reviewed through the static snapshot, which needs no listening socket.
	if _, err := os.Stat(runLiveMarker); err != nil {
		return fmt.Errorf("dashboard --serve: this run is not live (no %s) — the served dashboard exists only for the duration of a run; for a finished run drop --serve and read the written %s",
			runLiveMarker, filepath.Join(run.Dir(), "dashboard.html"))
	}
	secret, err := randToken()
	if err != nil {
		return fmt.Errorf("could not generate the run secret: %w", err)
	}
	ips := hostIPv4s()
	cert, err := selfSignedCert(ips)
	if err != nil {
		return fmt.Errorf("could not generate the ephemeral TLS certificate: %w", err)
	}

	srv := &http.Server{
		Addr:      fmt.Sprintf("0.0.0.0:%d", port),
		Handler:   dashboardHandler(secret, render),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
	}

	fmt.Fprintf(out, "serving %s dashboard on port %d over HTTPS — capability URL, rendered fresh per request:\n", filepath.Base(run.Dir()), port)
	if len(ips) == 0 {
		fmt.Fprintf(out, "  https://<this-host-ip>:%d/%s\n", port, secret)
	}
	for _, ip := range ips {
		fmt.Fprintf(out, "  https://%s:%d/%s\n", ip.String(), port, secret)
	}
	fmt.Fprintln(out, "  (self-signed cert — your browser will warn once; the URL secret is the only key; self-tears-down when the run ends)")

	// Self-teardown: watch the run-live marker and close the server on the live→ended transition.
	go watchAndShutdown(srv, run)

	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		return err
	}
	fmt.Fprintln(out, "run ended — dashboard server torn down")
	return nil
}

// dashboardHandler is the capability gate: it answers ONLY the exact secret path and renders the
// page fresh from `render`; every other path — a guess, a scan, a bare "/", the secret with a
// trailing segment — gets a flat 404. Extracted so the gate is unit-testable without a socket.
func dashboardHandler(secret string, render func() string) http.HandlerFunc {
	want := "/" + secret
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != want {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, render())
	}
}

// runLiveMarker is the file run-capture removes when a run ends. serveDashboard requires it to
// exist before binding, so the server's lifetime is bounded by the run's by construction.
var runLiveMarker = filepath.Join(".claude", "run-live.json")

// watchAndShutdown shuts srv down when the run ends, by EITHER of the two things that can say so.
//
// The marker going absent was the only signal, and its only remover is `capture` — so a run that
// was killed, threw, or was simply never captured left this server watching forever, holding a
// listening socket open with nothing left to end it. That is the exposure-with-no-owner the
// teardown above exists to prevent, reached by the one path the teardown could not see (#270).
//
// The RECORD is the second signal and the more truthful one: once the bench records the run's
// outcome, the run is over whatever the filesystem still says.
func watchAndShutdown(srv *http.Server, run record.Run) {
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for range tick.C {
		if ended, _ := runHasEnded(runLiveMarker, run); ended {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = srv.Shutdown(ctx)
			cancel()
			return
		}
	}
}

// randToken is a 24-byte crypto/rand secret as 48 hex chars — the run's capability key.
func randToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hostIPv4s lists the host's non-loopback IPv4 addresses (skipping link-local 169.254/16, which
// is auto-config noise, not a routable LAN address).
func hostIPv4s() []net.IP {
	var out []net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return out
	}
	for _, a := range addrs {
		ipn, ok := a.(*net.IPNet)
		if !ok || ipn.IP.IsLoopback() {
			continue
		}
		v4 := ipn.IP.To4()
		if v4 == nil || v4.IsLinkLocalUnicast() {
			continue
		}
		out = append(out, v4)
	}
	return out
}

// selfSignedCert mints an ephemeral in-memory ECDSA cert covering the host's IPs plus loopback,
// so TLS stands up without a file on disk. It authenticates nothing (self-signed); its only job
// is to encrypt the connection.
func selfSignedCert(ips []net.IP) (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	san := append([]net.IP{}, ips...)
	san = append(san, net.IPv4(127, 0, 0, 1), net.IPv6loopback)
	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "feov-record dashboard (ephemeral)"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           san,
		DNSNames:              []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}, nil
}

// runHasEnded answers the ONE question both watchers ask, in one place: is this run over?
//
// Two signals, because either alone is a half-answer:
//
//   - the run-live marker is gone — the clean end, written by `capture`
//   - the RECORD carries the bench's outcome — the truthful end, which holds even when nothing
//     removed the marker, because `capture` is optional and a killed run never reaches it (#270)
//
// It returns WHY as well as whether, so a watcher can tell the operator that a run ended without
// being captured rather than exiting silently — that gap is the thing worth knowing.
func runHasEnded(marker string, run record.Run) (bool, string) {
	if _, err := os.Stat(marker); err != nil {
		return true, "run-live marker gone"
	}
	if v := record.TerminalVerdict(run); v != "" {
		// IN THE SEAT'S OWN SPELLING. RunOutcome is stored as the schema's word (`unverified`) and
		// typed by the bench in capitals (`--as UNVERIFIED`); this line is read by an OPERATOR
		// looking for the state a seat recorded, so it shouts the word the seat used rather than
		// the one the column happens to hold.
		return true, "the record shows this run ended (" + strings.ToUpper(v) + ") while the run-live marker is still present — capture has not run"
	}
	return false, ""
}
