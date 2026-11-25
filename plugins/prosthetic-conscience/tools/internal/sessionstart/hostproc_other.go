//go:build !linux

package sessionstart

import (
	"fmt"
	"os"
)

func defaultHostProcessID() string {
	return fmt.Sprintf("ppid:%d", os.Getppid())
}
