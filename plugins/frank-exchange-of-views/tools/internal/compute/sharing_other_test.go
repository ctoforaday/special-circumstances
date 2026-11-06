//go:build !windows

package compute

import (
	"io/fs"
	"syscall"
	"testing"
)

// Off Windows no error is a sharing violation — not even the errno whose number Windows uses for
// one, which here is EPIPE.
func TestSharingViolationIsWindowsOnly(t *testing.T) {
	if sharingViolation(&fs.PathError{Op: "open", Path: "dir", Err: syscall.Errno(32)}) {
		t.Error("errno 32 read as a sharing violation off Windows")
	}
}
