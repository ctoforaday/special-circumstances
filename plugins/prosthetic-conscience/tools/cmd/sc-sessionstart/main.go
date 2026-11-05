// Command sc-sessionstart is a thin shim: the logic lives in internal/sessionstart, which
// composes one response from every SessionStart check (#201 step 3).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/sessionstart"
	"github.com/ctoforaday/special-circumstances/shared/gcpolicy"
)

func main() {
	gcpolicy.Short()
	os.Exit(sessionstart.Main())
}
