package nonet

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// THE GUARD REFUSES THE OUTSIDE AND ADMITS LOOPBACK, and both halves are the test.
//
// A guard that refused everything would pass the first assertion and break every package that
// serves its own fixture — which is most of them — so the loopback arm is not a courtesy, it is
// what makes the refusal safe to install suite-wide.
func TestOnlyLoopbackRefusesTheOutsideAndAdmitsLocalhost(t *testing.T) {
	OnlyLoopback()

	// A host we do not control. No DNS lookup happens: the name is refused before resolution,
	// so this assertion costs nothing even on a machine with no network at all.
	_, err := http.Get("https://example.com/")
	if err == nil {
		t.Fatal("the suite reached example.com — a test can still depend on a host nobody here controls")
	}
	if !strings.Contains(err.Error(), "nonet:") {
		t.Errorf("the refusal did not come from this guard, so something else failed and the guard is unproven: %v", err)
	}
	// The message has to say what to do, not only what was refused: the reader is someone whose
	// test just failed for a reason they did not expect.
	for _, want := range []string{"httptest.NewServer", "SERVE IT"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not tell the reader how to fix it (missing %q): %v", want, err)
		}
	}

	// AND THE ANTI-VACUITY HALF. httptest binds 127.0.0.1, which is exactly what a test that
	// needs a document should be doing.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("the pinned source"))
	}))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("the guard blocked a loopback server, which is how every fixture serves itself: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("loopback fetch returned %d", resp.StatusCode)
	}
}

// CALLING IT TWICE IS SAFE. A package with its own TestMain may install the guard after a shared
// one already did, and a second clone of an already-guarded transport would nest the check and
// report the refusal twice.
func TestOnlyLoopbackIsIdempotent(t *testing.T) {
	OnlyLoopback()
	OnlyLoopback()
	_, err := http.Get("https://example.com/")
	if err == nil {
		t.Fatal("the guard stopped refusing after a second install")
	}
	if n := strings.Count(err.Error(), "nonet:"); n != 1 {
		t.Errorf("the refusal is reported %d times — the guard has been layered over itself: %v", n, err)
	}
}
