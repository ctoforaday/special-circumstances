//go:build !windows

package runlive

import "os"

// openSharingDelete is os.Open where an open file never blocks its own removal.
func openSharingDelete(path string) (*os.File, error) {
	return os.Open(path)
}
