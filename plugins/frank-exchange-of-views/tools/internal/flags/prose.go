package flags

import (
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

// Prose is the prose payload channel AS ONE VALUE.
//
// # Why a type and not a flag plus a resolver
//
// It used to be one argument arriving three ways — `--reason` inline, `--reason-file <path>`, and
// `--reason-file -` from stdin — and held as loose flags that fact was true only for verbs that
// remembered every spelling. Three did not: `line-of-inquiry propose` filled its `line` from the raw
// inline flag, so the file form was refused for a field the seat had supplied; `spot-check` and
// `outcome` registered the inline flag by hand and shipped with no file form at all. A convention
// that has to be remembered at fifty call sites is not a mechanism, so the channel became a value
// that owns its own registration and its own reading.
//
// # ONE SPELLING, because the second one was the wrong fix for the right problem
//
// The file forms existed so long prose did not have to fight the shell. They never said WHICH
// fight, and the one that mattered was not quoting at all: bash RUNS a backtick inside double
// quotes, before this tool sees the text, and the command's output — or nothing — is recorded in
// its place. Measured over #861's smoke runs (2026-09-10): 409 double-quoted --reason values, 8 with
// backticks, every one of them substituted and stored silently — `prove` ran Perl's test harness,
// `factor 91` pasted its output into a finding — while the same seats used the file form 7 times.
// Twenty-five other free-text flags (--quote, --new, --problem, --fix, --check …) have no file form
// and never did, so a second spelling of ONE flag was never going to close the class.
//
// What closes it is one way for EVERY free-text value — capture it with a quoted heredoc, pass the
// variable — stated once in the help of every verb that takes free text (ProseFooter, attached by
// Text), and a PreToolUse
// deny of any tool command carrying a backtick the shell would run (internal/hookgate). With that
// rule in place the file spelling is a second way to write the same field, which is the one-way
// rule's own definition of an alias, and it went.
type Prose struct {
	inline string
}

// registry maps a command to the prose channel it registered.
//
// Cobra's Annotations are string-valued, and the alternative — threading the struct through every
// verb constructor — is the convention this type exists to replace. Guarded because the role trees
// are built in parallel by the gates, which is how an unguarded map in the help cache became a
// `concurrent map read and map write` in this package's sibling.
var (
	registryMu sync.RWMutex
	registry   = map[*cobra.Command]*Prose{}
)

// Register attaches the channel to a command: the flag, its canonical wording, the quoting rule in
// the verb's help, and the binding that makes this struct the thing the flag writes to.
//
// The prose channel is a free-text flag like any other, so it registers through TextVar, which is
// what attaches the quoting rule.
func (p *Prose) Register(c *cobra.Command) {
	TextVar(c, &p.inline, Reason, DescReason)
	registryMu.Lock()
	registry[c] = p
	registryMu.Unlock()
}

// RegisterRequired attaches the channel and makes it mandatory.
func (p *Prose) RegisterRequired(c *cobra.Command) {
	p.Register(c)
	_ = c.MarkFlagRequired(Reason)
}

// ProseOf returns the channel a command registered, or nil if it registered none.
func ProseOf(c *cobra.Command) *Prose {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[c]
}

// Read resolves the channel to one string.
//
// The trailing newline a captured heredoc can leave behind is trimmed, because a seat following
// the quoting rule should not have to think about it. `$(…)` already strips it; a variable filled
// some other way may not.
func (p *Prose) Read() string {
	return strings.TrimRight(p.inline, "\n")
}

// String is the resolved representation.
func (p *Prose) String() string { return p.Read() }
