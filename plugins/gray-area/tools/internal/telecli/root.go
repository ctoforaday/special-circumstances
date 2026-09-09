// Package telecli is telepathy's command tree.
//
// The layout is the one feov-record established and for the same reason: `main` hands over and
// does nothing else, the root and the shared flags are declared once here, and each verb builds
// itself in its own file. A verb is then readable on its own, and adding one cannot quietly change
// another's flags.
//
// Contract (Design by Contract):
//
//	Every absence MUST be worded. No store, no rows for a path, no sessions running and no
//	reasoning captured are four different facts, and a caller that cannot tell them apart will
//	read one as another. YOU MUST NOT print an empty table where a sentence is owed.
//	A search that cannot RUN MUST refuse, never report zero matches.
package telecli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// Env is everything the command tree reads from OUTSIDE its arguments — the two client
// directories, the store, and the clock.
//
// It is a struct rather than four calls to os.UserHomeDir and time.Now scattered through the verbs
// because those calls are what makes a command-line tool untestable: a golden file cannot be
// written against "however long ago that was in real time". Defaults come from Defaults(); a test
// supplies a fixture corpus and a frozen clock, and then exercises the SAME code path a user gets.
type Env struct {
	Store       string // the catalogue database
	ProjectsDir string // ~/.claude/projects — the transcripts, which we read and never write
	SessionsDir string // ~/.claude/sessions — the client's liveness advertisements
	Now         func() time.Time
}

// Defaults resolves the environment a real invocation runs in.
func Defaults() Env {
	e := Env{Now: time.Now}
	if dir, err := catalogue.DefaultDir(); err == nil {
		e.Store = filepath.Join(dir, "catalogue.db")
	}
	if home, err := os.UserHomeDir(); err == nil {
		e.ProjectsDir = filepath.Join(home, ".claude", "projects")
		e.SessionsDir = filepath.Join(home, ".claude", "sessions")
	}
	return e
}

const longDescription = `telepathy — what every agent on this box did, said and thought

Gray Area is the ship; telepathy is what it does. This reads ACROSS every agent on
the host and answers questions of fact about what ran, from a store built by the
capture hooks out of transcripts the harness already writes.

WHY IT EXISTS. SendMessage asks an agent what it BELIEVES, and needs it alive and
willing to answer. This asks the record what HAPPENED.

WHAT IT CANNOT TELL YOU. An 'unknown' liveness is not 'ended' — liveness is exact
on Linux only, and a session in another pid namespace is not ours to judge. A
session's final turn may be missing if its transcript lagged the last hook. And
reasoning is present only where it was captured: showThinkingSummaries defaults
off, so a session with no thoughts may have had them and not recorded them.

The five views ARE the contract, and telepathy sql reads them directly:
v_session, v_action, v_word, v_thought, v_skip.`

// usageError marks a refusal that is the CALLER's phrasing rather than a fact about the world, so
// Execute can exit 2 for it and 1 for everything else. Conflating the two makes a script unable to
// tell "you typed it wrong" from "the search found nothing and said so".
type usageError struct{ error }

func usagef(format string, a ...any) error { return usageError{fmt.Errorf(format, a...)} }

// NewRoot builds the command tree against a given environment.
func NewRoot(env Env) *cobra.Command {
	if env.Now == nil {
		env.Now = time.Now
	}
	root := &cobra.Command{
		Use:   "telepathy",
		Short: "what every agent on this box did, said and thought",
		Long:  longDescription,
		Example: `  telepathy agents
  telepathy touched internal/catalogue/schema.go
  telepathy sql "SELECT tool, count(*) n FROM v_action GROUP BY 1 ORDER BY 2 DESC"`,
		Version: buildid.Line("telepathy"),
		// The tree prints its own refusals through one path, so a wrong argument does not also
		// dump the full usage block over the answer.
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("{{.Version}}\n")

	// THE THREE PATHS EVERY VERB RESOLVES, declared once and inherited. They exist as flags
	// mainly so this tool can be pointed at a corpus that is not the caller's own — which is what
	// the golden tests do, and what anyone debugging a colleague's exported transcripts needs.
	root.PersistentFlags().StringVar(&env.Store, "store", env.Store,
		"the catalogue database the hooks write and every read verb queries")
	root.PersistentFlags().StringVar(&env.ProjectsDir, "projects", env.ProjectsDir,
		"the transcript root, which this reads and never writes")
	root.PersistentFlags().StringVar(&env.SessionsDir, "sessions", env.SessionsDir,
		"where the client advertises live sessions, which is the only source of liveness")

	// THE HELP STATES THE RULE, NOT THIS MACHINE'S ANSWER TO IT.
	//
	// pflag prints each flag's resolved default, so `--help` carried three absolute paths out of
	// the caller's own home directory — different on every machine, and quoted with %q, so on
	// Windows they arrived with escaped separators. Help that varies per machine cannot be a
	// contract, and the resolved value is not the useful half anyway: a reader wants to know WHERE
	// the tool looks, and a caller who needs the literal path gets it from the error that names it
	// ("no store at ..."). DefValue is display only; the values above are the actual defaults.
	for flag, rule := range map[string]string{
		"store":    "$XDG_STATE_HOME/special-circumstances/catalogue/catalogue.db",
		"projects": "~/.claude/projects",
		"sessions": "~/.claude/sessions",
	} {
		root.PersistentFlags().Lookup(flag).DefValue = rule
	}

	root.AddCommand(
		newAgentsCmd(&env),
		newSessionCmd(&env),
		newTouchedCmd(&env),
		newFindCmd(&env),
		newSQLCmd(&env),
		newBackfillCmd(&env),
	)
	return root
}

// Execute runs the tree against the real environment and returns a process exit code.
func Execute() int {
	root := NewRoot(Defaults())
	err := root.Execute()
	if err == nil {
		return 0
	}
	fmt.Fprintf(root.ErrOrStderr(), "telepathy: %v\n", err)
	return exitCode(err)
}

// exitCode is the ONE mapping from a refusal to a process status, so the golden tests assert the
// same rule the binary applies rather than a re-implementation of it that can drift from it.
//
//	0  the verb ran
//	2  the ARGV was wrong — a caller can fix this by typing something else
//	1  everything else: the store is missing, the query is invalid, ripgrep is absent
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ue usageError
	if errors.As(err, &ue) || isFlagError(err) {
		return 2
	}
	return 1
}

// isFlagError reports whether cobra refused the ARGV rather than the work. cobra does not export a
// type for this, so the check is on the prefix its own messages use; a miss costs an exit code of
// 1 where 2 was owed, never a wrong answer.
func isFlagError(err error) bool {
	for _, p := range []string{"unknown flag", "unknown shorthand flag", "flag needs an argument",
		"invalid argument", "unknown command", "accepts ", "requires "} {
		if strings.HasPrefix(err.Error(), p) {
			return true
		}
	}
	return false
}
