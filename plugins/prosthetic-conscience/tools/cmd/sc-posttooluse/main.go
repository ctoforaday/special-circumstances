// Command sc-posttooluse is a thin shim: the logic lives in internal/posttooluse, which
// runs every PostToolUse check over one shared context (#201 step 3).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/posttooluse"
)

func main() { os.Exit(posttooluse.Main()) }
