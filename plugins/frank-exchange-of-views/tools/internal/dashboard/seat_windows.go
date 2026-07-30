//go:build windows

package dashboard

import (
	"os"
	"syscall"
)

// seatStart returns the transcript's creation time in ms (the seat's start). Windows exposes
// CreationTime directly; ModTime is the fallback if the attribute data is absent.
func seatStart(tp string, fi os.FileInfo) float64 {
	if d, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok {
		return float64(d.CreationTime.Nanoseconds()) / 1e6
	}
	return float64(fi.ModTime().UnixNano()) / 1e6
}
