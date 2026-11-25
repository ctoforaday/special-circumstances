//go:build !linux

package peer

// HostServer returns an empty ServerIdentity on platforms without /proc inspection.
func HostServer() ServerIdentity { return ServerIdentity{} }

// CurrentDaemon reports no discoverable daemon on platforms without Linux systemd / /proc.
func CurrentDaemon() (ServerIdentity, bool) { return ServerIdentity{}, false }

// ProcessStartTime is not measurable on non-Linux platforms.
func ProcessStartTime(int) (int64, bool) { return 0, false }
