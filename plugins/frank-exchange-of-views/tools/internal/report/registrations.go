package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/shared/buildid"
)

// WHICH BUILD, WHICH HOOK, AND WHICH ROUTE TO THE RUN, AS EACH SEAT REGISTERED (run.md).
//
// `register` stamps three engine-observed facts on every seat: the record binary's commit
// (tool_version), the commit of the hook that injected the seat's environment (hook_version), and
// the path that supplied the run directory (run_via). They exist so a run with mixed binaries,
// stale hooks, or seats resolving the run by inference says so in its own record — and until this
// rendered them, saying so took a person reading the database: #751's diagnosis tabulated
// `run_via` across 29 registrations by hand to show the identity hook never fired.
//
// A UNIFORM RUN IS ONE LINE, AND A DISAGREEMENT IS NAMED SEAT BY SEAT. The run's stamp is the one
// most registrations carry; every registration that differs is listed with the differing parts in
// bold.
//
// A HOOK BUILD THAT IS NOT ITS OWN FEOV-RECORD BUILD IS NAMED SEAT BY SEAT TOO (#751). The record
// binary a seat runs and the hook that injected its environment are installed separately, so a
// record binary built from a working tree can run beside hooks from an older install while every
// registration agrees with every other. Each registration that carries both stamps is compared with
// itself; one with no hook stamp has no hook build to compare, and is read only against the others.
// The section says nothing about what was installed: the stamps are what ran.
//
// THE VALUES ARE THE RECORD'S, never re-derived: a registration with no hook stamp renders as
// having none, and `unknown` (an unstamped build) is said to be indistinguishable from another
// unstamped build rather than read as agreement.
func registrations(fam record.Family) string {
	type reg struct {
		seat string
		s    stamp
	}
	var regs []reg
	seats := map[string]bool{}
	for _, e := range fam.Events {
		r, ok := recordpb.BodyAs[*recordpb.Register](e)
		if !ok {
			continue
		}
		regs = append(regs, reg{e.GetSeatId(), stamp{tool: r.GetToolVersion(), hook: r.GetHookVersion(), via: r.GetRunVia()}})
		seats[e.GetSeatId()] = true
	}
	if len(regs) == 0 {
		return ""
	}
	// THE RUN'S STAMP is the one most registrations carry; on a tie, the earliest registered.
	count := map[stamp]int{}
	for _, r := range regs {
		count[r.s]++
	}
	var common stamp
	best := 0
	for _, r := range regs {
		if count[r.s] > best {
			common, best = r.s, count[r.s]
		}
	}

	// THE EXCEPTIONS, grouped by seat and stamp in the order they registered, so a seat that
	// registered twice with one odd stamp is one line with its count. The registrations whose hook
	// build is not their feov-record build are grouped the same way.
	type exception struct {
		seat string
		s    stamp
		n    int
	}
	group := func(list []*exception, at map[reg]*exception, r reg) []*exception {
		if x := at[r]; x != nil {
			x.n++
			return list
		}
		x := &exception{seat: r.seat, s: r.s, n: 1}
		at[r] = x
		return append(list, x)
	}
	var odd, mixed []*exception
	oddAt, mixedAt := map[reg]*exception{}, map[reg]*exception{}
	unknown := common.unknown()
	nMixed := 0
	for _, r := range regs {
		if r.s.mixed() {
			nMixed++
			mixed = group(mixed, mixedAt, r)
		}
		if r.s == common {
			continue
		}
		unknown = unknown || r.s.unknown()
		odd = group(odd, oddAt, r)
	}

	var flags []string
	if len(odd) > 0 {
		flags = append(flags, "NOT UNIFORM")
	}
	if nMixed > 0 {
		flags = append(flags, "MIXED BUILDS")
	}
	var b strings.Builder
	b.WriteString("**Builds and hooks")
	if len(flags) > 0 {
		b.WriteString(" — " + strings.Join(flags, ", "))
	}
	b.WriteString(".** ")
	if len(odd) == 0 {
		fmt.Fprintf(&b, "All %s, from %s, carry one stamp: %s.", plural(len(regs), "registration"), plural(len(seats), "seat"), common.phrase(stamp{}))
	} else {
		fmt.Fprintf(&b, "%d of %s carry the run's most common stamp: %s.", best, plural(len(regs), "registration"), common.phrase(stamp{}))
	}
	if unknown {
		b.WriteString(" `" + buildid.Unknown + "` is a build without a commit stamp, so registrations that agree on it may still have run different binaries.")
	}
	if len(odd) > 0 {
		b.WriteString("\n\nThe registrations that differ, with the differing parts in bold:\n")
		for _, x := range odd {
			times := ""
			if x.n > 1 {
				times = fmt.Sprintf(" (%d registrations)", x.n)
			}
			fmt.Fprintf(&b, "\n- `%s`%s: %s", x.seat, times, x.s.phrase(common))
		}
	}
	if nMixed > 0 {
		fmt.Fprintf(&b, "\n\n%d of %s ran a hook build other than their own feov-record build, so the hooks and the record binary came from different builds:\n", nMixed, plural(len(regs), "registration"))
		for _, x := range mixed {
			times := ""
			if x.n > 1 {
				times = fmt.Sprintf(" (%d registrations)", x.n)
			}
			fmt.Fprintf(&b, "\n- `%s`%s: feov-record build `%s`, hook build `%s`", x.seat, times, x.s.tool, x.s.hook)
		}
	}
	return b.String()
}

// stamp is one registration's three engine-observed facts.
type stamp struct{ tool, hook, via string }

func (s stamp) unknown() bool { return s.tool == buildid.Unknown || s.hook == buildid.Unknown }

// mixed is whether the hook that served the registration is another build than the feov-record
// binary that wrote it. The stamps are compared as stored; a registration no hook stamped has no
// hook build to compare.
func (s stamp) mixed() bool { return s.hook != "" && s.hook != s.tool }

// phrase writes the stamp as a list a singular or a plural subject can both carry, with each part
// that differs from against in bold. A zero against bolds nothing.
func (s stamp) phrase(against stamp) string {
	bold := func(differs bool, text string) string {
		if differs && against != (stamp{}) {
			return "**" + text + "**"
		}
		return text
	}
	tool := "no feov-record build stamped"
	if s.tool != "" {
		tool = "feov-record build `" + s.tool + "`"
	}
	hook := "no hook stamp — no hook served the seat, or one too old to stamp it"
	if s.hook != "" {
		hook = "hook build `" + s.hook + "`"
	}
	return bold(s.tool != against.tool, tool) + ", " + bold(s.hook != against.hook, hook) + ", " + bold(s.via != against.via, viaPhrase(s.via))
}

// viaPhrase says how the run directory reached the seat, one phrase per seatenv.RunSource.
func viaPhrase(via string) string {
	switch seatenv.RunSource(via) {
	case seatenv.RunFromEnv:
		return "run directory from the hook"
	case seatenv.RunFromFlag:
		return "run directory from a `--run` the seat typed"
	case seatenv.RunFromInference:
		return "run directory found by inference — the hook did not supply it"
	case seatenv.RunUnresolved:
		return "no run directory resolved"
	case "":
		return "no route to the run directory recorded (a registration made implicitly, by a seat's first write)"
	}
	return "run directory by `" + via + "`"
}

// plural writes "1 seat", "3 seats".
func plural(n int, word string) string { return fmt.Sprintf("%d %s", n, pluralWord(n, word)) }

func pluralWord(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
