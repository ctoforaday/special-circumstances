//go:build linux

package dashboard

import (
	"os"

	"golang.org/x/sys/unix"
)

// seatStart returns the transcript's birth time in ms (the seat's start) via statx; falls
// back to ModTime where the filesystem does not record btime.
func seatStart(tp string, fi os.FileInfo) float64 {
	var st unix.Statx_t
	if err := unix.Statx(unix.AT_FDCWD, tp, 0, unix.STATX_BTIME, &st); err == nil && st.Btime.Sec > 0 {
		return float64(st.Btime.Sec)*1000 + float64(st.Btime.Nsec)/1e6
	}
	return float64(fi.ModTime().UnixNano()) / 1e6
}
