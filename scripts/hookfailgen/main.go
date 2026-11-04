// Command hookfailgen propagates the one authored hookfailures package into every plugin module.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # Why copies at all
//
// The same reason buildidgen copies: each plugin's tools/ must build alone (`sc-doctor --fix`
// builds from the installed plugin root), so the modules are separate by design and a shared
// import is not available.
//
// # Why this package is worth generating rather than writing three times
//
// It is the one answer to a defect all three plugins had: a hook exits 0 by contract, so every
// failure it reported on stderr reached the debug log and nobody. gray-area's copy (#990) proved
// the shape in the field; three hand-kept copies of it would drift, and the drift would be silent
// in exactly the way the package exists to prevent.
//
// The copies are BYTE-IDENTICAL. What differs per plugin — the record directory's name — is a
// parameter the caller passes to New, never a value written into the file, so `-check` can assert
// the one property that matters: the three modules carry the same code.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/scripts/internal/pkgcopy"
)

// srcDir is the authored package. targets are the modules that need a copy of it.
const srcDir = "scripts/internal/hookfailures"

var targets = []string{
	"plugins/prosthetic-conscience/tools/internal/hookfailures",
	"plugins/gray-area/tools/internal/hookfailures",
	"plugins/frank-exchange-of-views/tools/internal/hookfailures",
}

func main() { os.Exit(pkgcopy.Main("hookfailgen", srcDir, targets)) }
