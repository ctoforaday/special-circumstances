//go:build !windows

package compute

// sharingViolation is Windows' alone: a POSIX open takes no share mode, so no error here is one.
func sharingViolation(error) bool { return false }
