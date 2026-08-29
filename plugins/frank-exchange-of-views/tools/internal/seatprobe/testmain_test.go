package seatprobe

import (
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// testbuild.Run removes the directory this package's built binaries live in. Without it they
// are abandoned in the system temp directory — ~38 MB per `go test` invocation, which nothing
// else ever deletes (#643). testbuild.Binary refuses to run without this, so the leak cannot
// come back by omission.
func TestMain(m *testing.M) { os.Exit(testbuild.Run(m)) }
