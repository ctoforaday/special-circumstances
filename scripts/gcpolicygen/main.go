// Command gcpolicygen propagates the one authored gcpolicy package into every plugin module.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// Every shipped binary states its garbage-collector class in its first statement, and the class
// means the same thing in every plugin, so the code that applies it is one package: authored and
// tested under scripts/internal/gcpolicy, copied into each module by pkgcopy, and `-check` fails a
// copy that has drifted.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/scripts/internal/pkgcopy"
)

// srcDir is the authored package. targets are the modules that need a copy of it.
const srcDir = "scripts/internal/gcpolicy"

var targets = []string{
	"plugins/prosthetic-conscience/tools/internal/gcpolicy",
	"plugins/gray-area/tools/internal/gcpolicy",
	"plugins/frank-exchange-of-views/tools/internal/gcpolicy",
}

func main() { os.Exit(pkgcopy.Main("gcpolicygen", srcDir, targets)) }
