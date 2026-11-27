//go:build !linux

package peer

import "errors"

// HostServer returns an empty ServerIdentity on platforms without /proc inspection.
func HostServer() ServerIdentity { return ServerIdentity{} }

// CurrentDaemon reports no discoverable daemon on platforms without Linux systemd / /proc.
func CurrentDaemon() (ServerIdentity, bool) { return ServerIdentity{}, false }

// ProcessStartTime is not measurable on non-Linux platforms.
func ProcessStartTime(int) (int64, bool) { return 0, false }

// procStartTime cannot read a start time here; its error is never fs.ErrNotExist, so a record
// reads Unknown rather than Ended.
func procStartTime(int) (int64, error) {
	return 0, errors.New("process start times are not measurable on this platform")
}
