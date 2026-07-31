// Command sc-strike-counter is a thin shim: the logic lives in internal/strikecounter so it can be
// tested, shared, and later composed with the other hooks on its event (#201).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/strikecounter"
)

func main() { os.Exit(strikecounter.Main()) }
