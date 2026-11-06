//go:build windows

package compute

import (
	"errors"

	"golang.org/x/sys/windows"
)

// sharingViolation reports an open Windows refused because another handle on the same object does
// not share with it. os.Open asks for FILE_SHARE_READ|FILE_SHARE_WRITE and not FILE_SHARE_DELETE,
// so any handle holding DELETE access — a rename in flight holds one on the directory it moves —
// refuses the open until that handle closes.
func sharingViolation(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}
