// Command sc-pretooluse is a thin shim: the logic lives in internal/pretooluse, which runs
// every PreToolUse check over one shared context (#201 step 3).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/pretooluse"
)

func main() { os.Exit(pretooluse.Main()) }
