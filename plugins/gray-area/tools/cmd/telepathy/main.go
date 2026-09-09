// telepathy reads the minds of every agent on this box.
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
// Contract (Design by Contract):
//
//	Every absence MUST be worded. No store, no rows for a path, no sessions running and no
//	reasoning captured are four different facts, and a caller that cannot tell them apart will
//	read one as another. YOU MUST NOT print an empty table where a sentence is owed.
//	A search that cannot RUN MUST refuse, never report zero matches.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/buildid"
)

const usage = `telepathy — what every agent on this box did, said and thought

  telepathy agents                 who is running, in which worktree, doing what
  telepathy session <id>           one session's shape: calls and errors by tool
  telepathy touched <path>         which sessions acted on a path, and when
  telepathy find <term>            search every local transcript (ripgrep, no index)
  telepathy sql '<SELECT ...>'     read-only SQL over the published views
  telepathy backfill               read the existing corpus into the store

Views, which are the contract: v_session, v_action, v_word, v_thought, v_skip.

  telepathy sql "SELECT tool, count(*) n FROM v_action GROUP BY 1 ORDER BY 2 DESC"

WHY IT EXISTS. SendMessage asks an agent what it BELIEVES, and needs it alive and
willing to answer. This asks the record what HAPPENED.

WHAT IT CANNOT TELL YOU. An 'unknown' liveness is not 'ended' — liveness is exact
on Linux only, and a session in another pid namespace is not ours to judge. A
session's final turn may be missing if its transcript lagged the last hook. And
reasoning is present only where it was captured: showThinkingSummaries defaults
off.

Flags:
  -version   print version and exit
`

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	if args[0] == "-version" || args[0] == "--version" {
		fmt.Fprintln(stdout, buildid.Line("telepathy"))
		return 0
	}
	cmd, rest := args[0], args[1:]
	need := func(what string) bool {
		if len(rest) < 1 {
			fmt.Fprintf(stderr, "telepathy %s: %s is required\n", cmd, what)
			return false
		}
		return true
	}
	switch cmd {
	case "agents":
		return agentsVerb(stdout, stderr)
	case "backfill":
		return backfillVerb(stdout, stderr)
	case "session":
		if !need("a session id (see `telepathy agents`)") {
			return 2
		}
		return sessionVerb(stdout, stderr, rest[0])
	case "touched":
		if !need("a path") {
			return 2
		}
		return touchedVerb(stdout, stderr, rest[0])
	case "find":
		if !need("a search term") {
			return 2
		}
		return findVerb(stdout, stderr, rest[0])
	case "sql":
		if !need("a query") {
			return 2
		}
		return sqlVerb(stdout, stderr, rest[0], 200)
	default:
		fmt.Fprintf(stderr, "telepathy: unknown command %q\n", cmd)
		fmt.Fprint(stderr, usage)
		return 2
	}
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
