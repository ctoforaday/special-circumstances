package shipped_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// One test per row of the table, and each makes ITS call and no other start: no `go` statement,
// no second row. Delete the row and its test leaves the race set.

func closedListener(t *testing.T) net.Listener {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	l.Close()
	return l
}

func closedServer() *http.Server {
	srv := &http.Server{Addr: "127.0.0.1:0"}
	srv.Close()
	return srv
}

func TestRowParallel(t *testing.T) { t.Parallel() }

func TestRowHTTPTestNewServer(t *testing.T) { httptest.NewServer(http.NotFoundHandler()).Close() }

func TestRowHTTPTestNewTLSServer(t *testing.T) {
	httptest.NewTLSServer(http.NotFoundHandler()).Close()
}

func TestRowHTTPTestNewUnstartedServer(t *testing.T) {
	httptest.NewUnstartedServer(http.NotFoundHandler()).Close()
}

func TestRowServerServe(t *testing.T) { _ = closedServer().Serve(closedListener(t)) }

func TestRowServerServeTLS(t *testing.T) { _ = closedServer().ServeTLS(closedListener(t), "", "") }

func TestRowServerListenAndServe(t *testing.T) { _ = closedServer().ListenAndServe() }

func TestRowServerListenAndServeTLS(t *testing.T) { _ = closedServer().ListenAndServeTLS("", "") }

func TestRowHTTPServe(t *testing.T) { _ = http.Serve(closedListener(t), nil) }

func TestRowHTTPServeTLS(t *testing.T) { _ = http.ServeTLS(closedListener(t), nil, "", "") }

func TestRowHTTPListenAndServe(t *testing.T) { _ = http.ListenAndServe("no-such-host.invalid:1", nil) }

func TestRowHTTPListenAndServeTLS(t *testing.T) {
	_ = http.ListenAndServeTLS("no-such-host.invalid:1", "", "", nil)
}

func TestRowTimeAfterFunc(t *testing.T) { time.AfterFunc(time.Hour, func() {}).Stop() }

func TestRowSignalNotify(t *testing.T) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Stop(c)
}

func TestRowContextAfterFunc(t *testing.T) {
	stop := context.AfterFunc(context.Background(), func() {})
	stop()
}

func TestRowErrgroupGo(t *testing.T) {
	var g errgroup.Group
	g.Go(func() error { return nil })
	_ = g.Wait()
}

func TestRowWaitGroupGo(t *testing.T) {
	var wg sync.WaitGroup
	wg.Go(func() {})
	wg.Wait()
}
