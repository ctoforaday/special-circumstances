// Command feov-posttoolusefailure is the PostToolUseFailure hook: a seat's failed tool call goes to
// the run's log as the tool's `failure` entry (see hookcmd.PostFailure).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookcmd"
)

func main() {
	os.Exit(hookcmd.Run("feov-posttoolusefailure", "PostToolUseFailure", hookcmd.PostFailure, os.Stdin, os.Stdout))
}
