//go:build !linux

package peer

import (
	"errors"
	"net/http"
)

// DaemonListens needs /proc; elsewhere it cannot be told.
func DaemonListens(int, string) (bool, error) {
	return false, errors.New("whether the daemon listens is read from /proc, which this platform does not have")
}

// DaemonAddress needs /proc; elsewhere the daemon's port cannot be told.
func DaemonAddress(*http.Client, int) (string, error) {
	return "", errors.New("the daemon's port is read from /proc, which this platform does not have")
}
