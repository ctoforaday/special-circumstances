package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"regexp"
	"strings"
	"testing"
)

// A SEAT NAVIGATES BY WHAT IT READS, SO EVERY VIEW MUST NAME THE VERB THAT FILLS IT.
//
// MEASURED, 2026-08-10, from a haiku seat's transcript on the seatprobe board:
//
//	FEOV blue show --view lines-of-inquiry     <- it read the projection
//	FEOV blue line-of-inquiry --id A1 --decision pursued --conclusion "..."   <- INVENTED
//	FEOV blue line-of-inquiry --id A1 --decision pursued --conclusion "..."            <- invented flags
//	FEOV blue line of inquiry --help                    <- only NOW did it read the contract
//
// It built a verb name out of the VIEW name, singularised, and invented two flags to go with it.
// The verb is `line of inquiry`; nothing in the projection it had just read says so. In 32 tool calls that
// seat read `--help` exactly once, and only after two invented calls failed — so the working
// theory that a seat learns its surface from `<role> --help` is false in practice. It learns it
// from what the tool prints back.
//
// That makes a projection's name a de facto API, and a projection that names a concept without
// naming the verb that writes it is teaching a name that does not work.
//
// # What this test does and does not claim
//
// It does NOT require the view and the verb to share a name — `lines-of-inquiry` is the right
// word for the reader of a report and `line of inquiry` is the right word for the seat proposing one, and
// collapsing them would cost more than it saves. It requires only that the view's own help SAY
// which verb fills it, so the seat that reads the projection has the verb in the same breath.

// viewWriters maps each projection to the verb(s) whose events fill it. It is declared rather
// than derived because no structure in the tree connects them today — which is precisely the gap
// this test exists to make visible rather than to paper over.
var viewWriters = map[string][]string{
	// The round-0 synthesis authors it; every later change is an `edit`. `cite`, `finding` and
	// `prove` splice the anchor layer into it, which is why they are named here too — a seat
	// reading the report meets their tokens and has to carry them.
	"report": {"edit", "cite", "finding", "prove"},
	// The motion group fills it: the ask, the answer, and the press-on. All three, because a
	// view that named only `file` would leave a reader wondering where a ruling comes from.
	"motions":  {"motion", "rule", "appeal"},
	"work":     {"mint"},
	"findings": {"finding"},
	"debate":   {"position"},
	"changes":  {"edit"},
	// All four, because the view answers two different questions with one table: what blue
	// offered as backing (`cite`, `prove`) and what red made of it (`verify`, `reproduce`).
	"evidence":         {"cite", "prove", "verify", "reproduce"},
	"lines-of-inquiry": {"line-of-inquiry", "propose"},
	"telemetry":        {},
	// Computed from the record at read time, not written by any verb — same as telemetry. The
	// description says so in as many words, so a seat cannot go looking for a `scorecard` verb.
	"scorecard": {},
	"board":     {"mint", "close", "regrade", "retire"},
}

// EVERY VIEW'S HELP NAMES THE VERB THAT FILLS IT.
//
// The failure this prevents is the one measured above: a seat reads a projection, forms a concept
// name from it, and types a verb that does not exist. It then either finds the right one by
// failing twice, or — the expensive case — decides the capability is missing and works around it
// in prose, which is a capability lost for the whole run and reported nowhere.
func TestEveryViewNamesTheVerbThatFillsIt(t *testing.T) {
	// ASKED OF A SEAT, because `show` only exists inside one. Any seat will do: the view table is
	// shared, and this gate is about the DESCRIPTIONS, which do not vary by role.
	help, err := run(t, "show", "--help", "--seat-id", record.SampleSeatOf("blue"))
	if err != nil {
		t.Fatalf("show --help: %v", err)
	}

	names := ViewNames()
	if len(names) == 0 {
		t.Fatal("no views found — a walk that finds nothing passes this forever")
	}
	for _, view := range names {
		writers, declared := viewWriters[view]
		if !declared {
			t.Errorf("view %q is in the tree and not in viewWriters — say which verb fills it, or that nothing does. An undeclared view is one nobody has asked the discoverability question about", view)
			continue
		}
		if len(writers) == 0 {
			continue // telemetry and friends: written by the engine, not by a seat verb
		}
		var named bool
		for _, w := range writers {
			if strings.Contains(help, w) {
				named = true
				break
			}
		}
		if !named {
			t.Errorf("`show --help` describes view %q without naming any verb that fills it (%s).\n"+
				"A seat reads the projection and builds a verb name out of it: measured, one read `--view lines-of-inquiry` and typed `blue line-of-inquiry`, which does not exist. Name the verb in the view's own description so the two arrive together.",
				view, strings.Join(writers, ", "))
		}
	}
}

// AND NO DECLARED VIEW HAS GONE AWAY.
//
// The inverse, for the same reason every table in this package has one: an entry for a view that
// no longer exists is checked coverage of nothing, and it reads exactly like coverage.
func TestViewWritersHasNoStaleEntries(t *testing.T) {
	live := map[string]bool{}
	for _, v := range ViewNames() {
		live[v] = true
	}
	for view := range viewWriters {
		if !live[view] {
			t.Errorf("viewWriters names %q, which is not a view in the tree — remove it", view)
		}
	}
}

// NO MESSAGE NAMES A VIEW THAT DOES NOT EXIST.
//
// The inverse of the contract above, and it caught a real one: `lens reproduce --id` told a seat
// the proofs were listed by `show --view proofs`. There is no `proofs` view. A seat following that
// runs a command which is refused, and the refusal it gets — "unknown view" — reads as its own
// mistake rather than the tool's.
//
// This is the `--view lines-of-inquiry` failure in reverse. There, a seat invented a VERB from a
// view name; here the tool invents a VIEW name for a seat. Both come from a projection vocabulary
// referenced in prose that nothing checks, which is why the check is mechanical: every `--view X`
// written anywhere in the tree must name a projection that exists.
func TestNoHelpOrErrorNamesAViewThatDoesNotExist(t *testing.T) {
	live := map[string]bool{}
	for _, v := range ViewNames() {
		live[v] = true
	}

	// Every help string in the tree, plus the Long/Short of every command.
	var texts []string
	var walkAll func(*cobra.Command)
	walkAll = func(c *cobra.Command) {
		texts = append(texts, c.Short, c.Long)
		c.Flags().VisitAll(func(f *pflag.Flag) { texts = append(texts, f.Usage) })
		for _, sub := range c.Commands() {
			walkAll(sub)
		}
	}
	for _, r := range AllRoots() {
		walkAll(r)
	}

	named := regexp.MustCompile(`--view\s+([a-z][a-z-]*)`)
	seen := map[string]bool{}
	for _, text := range texts {
		for _, m := range named.FindAllStringSubmatch(text, -1) {
			name := m[1]
			if name == "changes" || live[name] {
				continue
			}
			if seen[name] {
				continue
			}
			seen[name] = true
			t.Errorf("a help string names `--view %s`, which is not a projection. A seat that follows it is refused for a mistake the tool made, and reads the refusal as its own.\n\nthe projections are: %s",
				name, strings.Join(ViewNames(), ", "))
		}
	}

	// AND THE SAME CHECK FOR THE SUBCOMMAND FORM, which the `--view` regex above cannot see.
	//
	// `show` became a group, so the way a message names a projection is now `show evidence`, not
	// `--view evidence` — and the moment that happened, the gate above went on passing while
	// covering a syntax nothing writes any more. The miss is indistinguishable from the honest
	// zero: no findings, which reads exactly like a clean tree.
	//
	// A forward match on `show <token>` cannot work — "show board" and "show the seat" are the
	// same shape — so this runs BACKWARD, from a roster of names that USED to be projections. It
	// catches the regression that actually happens (a view is renamed and a help string keeps the
	// old word) with no guesswork about which English words are view names.
	// AND NO HELP STRING DOUBLES THE VERB. `show --run <dir> show evidence` parses as the show
	// group with the argument "show" and is refused. It is not a hypothetical: the restructure
	// wrote it into twelve prompt sites, and while writing the fix for those I wrote it again into
	// this package's own `lens reproduce` message. A shape you reproduce while repairing it is one
	// that needs a check rather than care.
	doubled := regexp.MustCompile(`\bshow\s+(?:--\S+\s+\S+\s+)+show\b`)
	for _, text := range texts {
		if m := doubled.FindString(text); m != "" {
			t.Errorf("a help string doubles the verb: `%s` — it parses as the show group with the argument \"show\", and the tool refuses it. Put the projection first: `show <view> --run <dir>`",
				strings.Join(strings.Fields(m), " "))
			break
		}
	}

	for _, gone := range retiredViews {
		if live[gone] {
			t.Errorf("retiredViews names %q, which IS a live projection — remove it from the roster before it starts failing every rename", gone)
			continue
		}
		stale := regexp.MustCompile(`(--view|show)\s+` + regexp.QuoteMeta(gone) + `\b`)
		for _, text := range texts {
			if stale.MatchString(text) {
				t.Errorf("a help string still names the retired projection %q. It was renamed and this carrier kept the old word — a seat following it gets 'unknown view' and reads the refusal as its own mistake.\n\nthe projections are: %s",
					gone, strings.Join(ViewNames(), ", "))
				break
			}
		}
	}
}

// retiredViews are projection names that existed and do not any more. Every rename adds one, and
// the entry is the whole point: a name nobody is watching for is how a stale reference survives a
// rename in prose that compiles fine.
var retiredViews = []string{"citation-ledger", "ledger", "archive", "changelog", "proofs"}
