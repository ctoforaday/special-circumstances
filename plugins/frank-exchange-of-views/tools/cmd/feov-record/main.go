// Command feov-record is the seat-side record runtime for the debate engine.
//
// Everything lives in internal/cli: the root command, the flags every seat
// shares, and one package per role whose verbs each build themselves in their
// own file. This file does what a cobra main does — nothing but hand over.
package main

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"

func main() {
	cli.Execute()
}
