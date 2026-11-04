// Command buildidgen propagates the one authored buildid package into every plugin module.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # Why copies at all
//
// Each plugin's tools/ has to build on its own: `sc-doctor --fix` runs
// `go build -C <installedPluginRoot>/tools ./cmd/<name>`, and a consumer with one plugin
// installed has only that plugin's tree on disk. A shared module they all import would make
// that build fail, so the modules stay separate and the code is copied.
//
// # Why generated copies
//
// Because hand-written ones drift, and had. The two existing copies were identical in
// substance and different in comments, and only one of them carried ExeName — so the module
// without it hand-rolled the same `.exe` conditional inside a test, which is how a rule about
// executable naming ends up stated twice and enforced once.
//
// One authored package under scripts/internal/buildid, compiled and tested there; a copy
// written into each module; `-check` fails when a copy has drifted. The copies carry a
// DO NOT EDIT header so the next person edits the source instead of a leaf.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/scripts/internal/pkgcopy"
)

// srcDir is the authored package. targets are the modules that need a copy of it.
const srcDir = "scripts/internal/buildid"

var targets = []string{
	"plugins/prosthetic-conscience/tools/internal/buildid",
	"plugins/gray-area/tools/internal/buildid",
	"plugins/frank-exchange-of-views/tools/internal/buildid",
}

func main() { os.Exit(pkgcopy.Main("buildidgen", srcDir, targets)) }
