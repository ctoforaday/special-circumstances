// Command telepathy reads the minds of every agent on this box.
//
// # Why this is its own binary
//
// Gray Area is the SHIP — the plugin, named for the Culture's mind-reader. Telepathy is what it
// does. `gray-area agents` reads as plugin-name plus verb; `telepathy agents` names the act, which
// is what a reader needs when deciding whether to reach for it.
//
// The split is also real rather than cosmetic. `gray-area` adjudicates ONE trajectory against a
// document — a checkpoint's claims, a pull request body — and every row it prints cites file, line
// and uuid so a human can check it. This reads ACROSS every agent on the host and answers questions
// of fact about what ran. Different question, different subject, different store.
//
// Everything lives in internal/telecli. This file does what a cobra main does — nothing but hand
// over.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/telecli"
)

func main() { os.Exit(telecli.Execute()) }
