//go:build !linux

package peer

import (
	"errors"
	"net/http"
)

// DaemonAddress needs /proc; elsewhere the daemon's port cannot be told.
func DaemonAddress(*http.Client, int) (string, error) {
	return "", errors.New("the daemon's port is read from /proc, which this platform does not have")
}
