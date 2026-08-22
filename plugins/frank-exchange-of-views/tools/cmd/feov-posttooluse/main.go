// Command feov-posttooluse is the PostToolUse backstop: it blocks and forces a repair when a
// non-author call dropped an immortal anchor from blue/report.md by any mechanism, including the
// exotic Bash the PreToolUse verb-set cannot enumerate.
//
// A DEDICATED BINARY, NOT A VERB — see internal/hookcmd.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookcmd"
)

func main() { os.Exit(hookcmd.Run(hookcmd.Post, os.Stdin, os.Stdout)) }
