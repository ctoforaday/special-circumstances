//go:build windows

package compute

import (
	"io/fs"
	"testing"

	"golang.org/x/sys/windows"
)

// The classifier keys on ERROR_SHARING_VIOLATION alone, wrapped the way os.Open returns it.
func TestSharingViolationIsTheOneErrno(t *testing.T) {
	for _, c := range []struct {
		err  error
		want bool
	}{
		{&fs.PathError{Op: "open", Path: "dir", Err: windows.ERROR_SHARING_VIOLATION}, true},
		{&fs.PathError{Op: "open", Path: "dir", Err: windows.ERROR_LOCK_VIOLATION}, false},
		{&fs.PathError{Op: "open", Path: "dir", Err: windows.ERROR_ACCESS_DENIED}, false},
		{&fs.PathError{Op: "open", Path: "dir", Err: windows.ERROR_FILE_NOT_FOUND}, false},
		{fs.ErrNotExist, false},
	} {
		if got := sharingViolation(c.err); got != c.want {
			t.Errorf("sharingViolation(%v) = %v, want %v", c.err, got, c.want)
		}
	}
}
