//go:build !linux

package peer

// CurrentDaemon reports no discoverable daemon on platforms without Linux systemd / /proc.
func CurrentDaemon() (ServerIdentity, bool) { return ServerIdentity{}, false }
