//go:build !windows && !linux

package dashboard

import "os"

// seatStart falls back to ModTime where birthtime is not portably available.
func seatStart(_ string, fi os.FileInfo) float64 {
	return float64(fi.ModTime().UnixNano()) / 1e6
}
