package cli

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// This package drives a binary built by testbuild, and that build directory is shared by
// every caller in this test binary — so the process exit is the only point at which removing
// it is safe. Omitting this leaks one directory holding one linked binary per run into
// TMPDIR, which is a tmpfs on the paths that matter. See testbuild.Main.
func TestMain(m *testing.M) { testbuild.Main(m) }
