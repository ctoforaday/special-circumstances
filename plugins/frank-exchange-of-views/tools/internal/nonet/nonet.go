// Package nonet keeps the test suite off the public internet.
//
// A test that fetches a host we do not control fails for reasons that have nothing to do with the
// code, and the failure arrives wearing the message of whatever it interrupted. Measured, on the
// nightly of 2026-09-06: `TestEveryProbeBoardStillBuilds/sources` timed out reaching
// example.com and reported "this fixture speaks a retired model" — the red said the fixture was
// stale when it meant the network was slow. At a gate that only runs at the boundary that is the
// worst available confusion, because a flake is then indistinguishable from the thing the gate
// exists to catch, and it lands when the cost of misreading it is highest.
//
// #807 fixed the two call sites. This removes the CLASS: a third one cannot be added without the
// suite saying so, at the dial, in its own words.
//
// # Why a dialer and not a sweep over the source
//
// The obvious alternative is a check-gate grepping `_test.go` for outbound URL literals. It needs
// an allowlist on its first run — `board_test.go` carries example.com as corroboration REFERENCE
// STRINGS on a path that never fetches, and claimcount's fixtures carry them inside markdown — and
// a hand-kept allowlist beside a rule is the shape facts-are-fields exists to refuse. Worse, it
// measures the wrong thing: what a test CONTAINS rather than what it DOES.
//
// The dialer asks the question that actually matters, and a package that genuinely needs a server
// stands one up on loopback, which several already do (`httptest`, and the fuzz's own sourceURL).
package nonet

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// OnlyLoopback replaces the default transport's dialer with one that refuses any address outside
// loopback. It is idempotent, so a TestMain that calls it twice — or a package whose own TestMain
// calls it after a shared one already did — is not a problem.
//
// IT GUARDS THE DEFAULT TRANSPORT, which is what the production fetcher uses: NewHTTPFetcher
// builds an http.Client with a Timeout and no Transport, so every fetch this module makes goes
// through here. A fetcher that grew its own Transport would escape this, and that is worth
// knowing rather than papering over — the test below pins the current arrangement.
func OnlyLoopback() {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok || guarded {
		return
	}
	t := base.Clone()
	inner := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if !loopback(addr) {
			return nil, fmt.Errorf("nonet: this suite refused a connection to %q.\n\n"+
				"A test must not depend on a host we do not control: it fails for reasons unrelated "+
				"to the code, and the failure is reported by whatever it interrupted rather than as "+
				"a network fault. If this test needs a document, SERVE IT — httptest.NewServer, or "+
				"seatprobe.serveSource for a probe board — and pass that URL. If it needs to prove "+
				"the fetch path handles a real refusal, point it at a closed loopback port.", addr)
		}
		return inner.DialContext(ctx, network, addr)
	}
	http.DefaultTransport = t
	guarded = true
}

// guarded makes OnlyLoopback idempotent. Without it a second call clones the ALREADY-GUARDED
// transport and nests the check, which still works but makes the refusal message arrive twice.
var guarded bool

// loopback reports whether an address is one this suite may reach. It reads the HOST STRING rather
// than resolving it: a name is refused before DNS, so the suite cannot be slowed by a lookup for a
// host it was never going to be allowed to dial.
func loopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
