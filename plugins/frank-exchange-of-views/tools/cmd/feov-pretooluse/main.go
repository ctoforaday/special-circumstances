// Command feov-pretooluse is the PreToolUse hook backend for the blue-report lockdown and the
// run-directory injection. It reads the hook event on stdin and writes its decision to stdout.
//
// A DEDICATED BINARY, NOT A VERB. See internal/hookcmd for why: a hook has no seat identity, and
// hanging it off an identity-scoped command tree cost an exemption at every layer — identity,
// help visibility, and the surface census — before the shipped invocation stopped exiting 2 and
// denying every mutating tool call in the session.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookcmd"
)

func main() { os.Exit(hookcmd.Run(hookcmd.Pre, os.Stdin, os.Stdout)) }
