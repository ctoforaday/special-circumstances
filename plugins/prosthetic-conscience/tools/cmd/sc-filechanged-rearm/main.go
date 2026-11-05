// Command sc-filechanged-rearm is a thin shim: the logic lives in internal/filechangedrearm so it can be
// tested, shared, and later composed with the other hooks on its event (#201).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/filechangedrearm"
	"github.com/ctoforaday/special-circumstances/shared/gcpolicy"
)

func main() {
	gcpolicy.Short()
	os.Exit(filechangedrearm.Main())
}
