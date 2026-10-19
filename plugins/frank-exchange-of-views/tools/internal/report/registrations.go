package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookprovenance"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// WHICH BUILD, WHICH HOOK, AND WHICH ROUTE TO THE RUN, AS EACH SEAT REGISTERED (run.md).
//
// `register` stamps three engine-observed facts on every seat: the record binary's commit
// (tool_version), the commit of the hook that injected the seat's environment (hook_version), and
// the path that supplied the run directory (run_via). They exist so a run with mixed binaries,
// stale hooks, or seats resolving the run by inference says so in its own record — and until this
// rendered them, saying so took a person reading the database: #751's diagnosis tabulated
// `run_via: wrapper (29/29)` by hand to show the identity hook never fired.
//
// A UNIFORM RUN IS ONE LINE, AND A DISAGREEMENT IS NAMED SEAT BY SEAT. The run's stamp is the one
// most registrations carry; every registration that differs is listed with the differing parts in
// bold. Setup's hook provenance (inputs/run-config.json) is set beside the stamps, and a
// contradiction between the two — hooks setup saw none of stamping seats, or hooks setup saw
// installed and seats no hook served — is stated rather than left for the reader to cross-check.
//
// THE VALUES ARE THE RECORD'S, never re-derived: a registration with no hook stamp renders as
// having none, and `unknown` (an unstamped build) is said to be indistinguishable from another
// unstamped build rather than read as agreement.
func registrations(run record.Run, fam record.Family) string {
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
	// registered twice with one odd stamp is one line with its count.
	type exception struct {
		seat string
		s    stamp
		n    int
	}
	var odd []*exception
	at := map[reg]*exception{}
	unknown := common.unknown()
	stamped := 0 // registrations a hook stamped
	for _, r := range regs {
		if r.s.hook != "" {
			stamped++
		}
		if r.s == common {
			continue
		}
		unknown = unknown || r.s.unknown()
		if x := at[r]; x != nil {
			x.n++
			continue
		}
		x := &exception{seat: r.seat, s: r.s, n: 1}
		at[r] = x
		odd = append(odd, x)
	}

	verdict, contradicts := hookSide(run, stamped, len(regs))
	var b strings.Builder
	b.WriteString("**Builds and hooks")
	if len(odd) > 0 || contradicts {
		b.WriteString(" — NOT UNIFORM")
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
	if verdict != "" {
		b.WriteString(" " + verdict)
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
	return b.String()
}

// stamp is one registration's three engine-observed facts.
type stamp struct{ tool, hook, via string }

func (s stamp) unknown() bool { return s.tool == buildid.Unknown || s.hook == buildid.Unknown }

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
	case seatenv.RunFromWrapper:
		return "run directory from the per-run wrapper — no hook reached the seat"
	case seatenv.RunFromInference:
		return "run directory found by inference — neither the hook nor the wrapper supplied it"
	case seatenv.RunUnresolved:
		return "no run directory resolved"
	case "":
		return "no route to the run directory recorded (a registration made implicitly, by a seat's first write)"
	}
	return "run directory by `" + via + "`"
}

// hookSide is setup's hook provenance in one sentence, and whether it contradicts the stamps:
// hooks setup saw none of while a hook stamped a registration, or hooks setup saw installed while
// a registration carries no hook stamp. A handle with no run directory (a family built in a test)
// has no configuration to read, and says nothing.
func hookSide(run record.Run, stamped, total int) (string, bool) {
	if run.Dir() == "" {
		return "", false
	}
	p, ok, err := hookprovenance.Read(run.Dir())
	if err != nil {
		return "**Setup's account of the installed hooks cannot be read:** " + err.Error() + ".", true
	}
	if !ok {
		return "**Setup recorded nothing about the installed hooks**, so nothing here can be compared with what it saw.", true
	}
	var b strings.Builder
	contradicts := false
	switch p.State {
	case "installed":
		fmt.Fprintf(&b, "Setup found the hooks installed, cached as %s %s", pluralWord(len(p.Versions), "version"), andList(p.Versions))
		if len(p.Versions) > 1 {
			b.WriteString("; which of them served a seat is not something setup can see")
		}
		b.WriteString(".")
		var short []string
		for _, v := range sortedVersions(p.Missing) {
			if len(p.Missing[v]) > 0 {
				short = append(short, v+" registers "+strings.Join(p.Missing[v], ", "))
			}
		}
		if len(short) > 0 {
			contradicts = true
			b.WriteString(" **A cached version registers hook binaries it does not ship:** " + strings.Join(short, "; ") + ".")
		}
		if unstamped := total - stamped; unstamped > 0 {
			contradicts = true
			fmt.Fprintf(&b, " **Yet %d of %d registrations carry no hook stamp.**", unstamped, total)
		}
	default:
		contradicts = stamped > 0
		fmt.Fprintf(&b, "**Setup found the hooks %s:** %s.", p.State, strings.TrimSuffix(p.Why, "."))
		if stamped > 0 {
			fmt.Fprintf(&b, " **Yet %d of %d registrations were stamped by a hook**, so the hooks that served this run are not the ones setup could see.", stamped, total)
		}
	}
	return b.String(), contradicts
}

func sortedVersions(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// plural writes "1 seat", "3 seats".
func plural(n int, word string) string { return fmt.Sprintf("%d %s", n, pluralWord(n, word)) }

func pluralWord(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// andList writes "a", "a and b", "a, b and c".
func andList(words []string) string {
	if len(words) < 2 {
		return strings.Join(words, "")
	}
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}
