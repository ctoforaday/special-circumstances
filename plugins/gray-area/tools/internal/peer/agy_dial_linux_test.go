//go:build linux

package peer

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"testing"
	"time"
)

// fullBacklog is the address of a listening socket whose accept queue is full and never drained,
// so a new connection's SYN is dropped and its dial hangs.
func fullBacklog(t *testing.T) string {
	t.Helper()
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { syscall.Close(fd) })
	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Listen(fd, 0); err != nil {
		t.Fatal(err)
	}
	sa, err := syscall.Getsockname(fd)
	if err != nil {
		t.Fatal(err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", sa.(*syscall.SockaddrInet4).Port)
	for i := 0; i < 8; i++ {
		c, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err != nil {
			return addr // the queue is full: this dial hung
		}
		t.Cleanup(func() { c.Close() })
	}
	t.Skip("the accept queue never filled, so no dial hangs here")
	return ""
}

// A DIAL THAT NEVER COMPLETES NEVER SENT THE MESSAGE. The dial's own bound, below the request's, makes
// it fail as a dial — UnreachedError — instead of as the request's timeout, which replaces the
// dial's error and reads as a message that may have arrived.
func TestSendAgyServerDialTimeoutIsUnreached(t *testing.T) {
	addr := fullBacklog(t)
	start := time.Now()
	_, err := SendAgyServer(NewAgyClient(agyDialTimeout+3*time.Second), Server{Address: addr, Token: "tok"}, AgyMessage{}, RouteAgyOwnRecord)
	var u *UnreachedError
	if !errors.As(err, &u) {
		t.Errorf("err = %v after %s, want UnreachedError", err, time.Since(start).Round(time.Millisecond))
	}
}
