// Command feov-posttooluse is the PostToolUse hook: a seat's call returning its envelope that
// succeeded is recorded as the envelope's admission (see hookcmd.Post).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookcmd"
)

func main() {
	os.Exit(hookcmd.Run("feov-posttooluse", "PostToolUse", hookcmd.Post, os.Stdin, os.Stdout))
}
