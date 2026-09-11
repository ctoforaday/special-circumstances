package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/diagnostics"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms"
)

// manualName is the command's name. The survey that reads a manual out of a trajectory owns it,
// with the page grammar, so the writer and the reader cannot come to disagree about either.
const manualName = diagnostics.ManualCommand

// newManual prints every command on THIS surface with that command's own help, in one call.
//
// WHY IT EXISTS. Seats learn the tool by reading its help, and walking it a page at a time was
// measured in a real run (#861 B3): 142 help turns across 26 sittings, 14% of all turn time, a cold
// sitting's median seven help turns before its first act. One call replaces the walk.
//
// WHY IT RUNS THE HELP RATHER THAN SHOWING A COPY OF IT. An earlier proposal (#593) staged a
// snapshot of the help tree and was rejected because a snapshot drifts from the tool it describes.
// This executes each command's real `--help` at call time — the same binary, the same seat, the
// same run — so what it prints IS what the command would say, and there is nothing to keep in step.
//
// WHY IT IS LEAN. Printed whole, the pages were larger than all the rest of the seat's text put
// together, and 41% of a blue manual was lines repeated on every page (the Global Flags alone ran
// to ~31k characters across 41 pages). So a block that repeats verbatim is printed once, and a group
// page that only lists commands with pages of their own is left out. Each command's own `--help` is
// untouched; the economy is the manual's, and diagnostics.ExpandManual reverses it exactly.
//
// The surface is the one this tree was built for — the flag, else the seat this agent registered
// as, exactly as every invocation resolves it — so seatID is captured at construction rather than
// re-derived. It records nothing.
func newManual(seatID string) *cobra.Command {
	return &cobra.Command{
		Use:   manualName,
		Short: "every command on your surface, each with its own help, run live — read it once, first thing in a sitting",
		Long: manualName + " prints every command on this surface, in tree order, each under a header naming the command exactly, followed by the help that command prints. The help is produced now, by running this binary with that command and --help, so it cannot disagree with the tool you are running. It records nothing.\n\n" +
			"It is lean on purpose. A block that repeats word for word on several pages — the flags every command inherits, a footer, a flag line several commands share — is printed ONCE, in a SHARED section before the first page, and each page carries a marker line ending `→ SHARED §n` where the block was. A group whose page only lists commands that have pages of their own is left out. Nothing a page said is lost: put each marked block back and you have that command's --help exactly.\n\n" +
			"Before the pages it prints WORDS THIS SURFACE USES: each concept word your surface uses, defined once — the one word every page, prompt and constitution uses for that concept.\n\n" +
			"Read it ONCE, at the start of a sitting and before you decide what to do: it is your whole surface in one call. The surface is your seat's — the seat you registered as, or the one this call names. A page whose help fails to run is printed with its error, never left out, and the command then exits non-zero.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return printManual(cmd, seatID)
		},
	}
}

// leftOutOfManual says whether a command gets no page of its own: `manual` itself, and a GROUP
// whose page would say nothing but the names of commands that each have a page.
//
// DECIDED FROM THE COMMAND, NOT FROM ITS PRINTED HELP. A group carries something of its own when it
// has prose beyond its menu line, a flag besides help, or a bare form that does something (`show`
// answers with the seat's pending work); any one of those keeps its page. What is left is a heading.
func leftOutOfManual(c *cobra.Command) bool {
	if c.Name() == manualName && !c.HasParent() {
		return true
	}
	if c.Name() == manualName && c.Parent() != nil && !c.Parent().HasParent() {
		return true
	}
	if !c.HasSubCommands() {
		return false
	}
	if c.Runnable() && c.Annotations[BareIsACapability] == "yes" {
		return false
	}
	if long := strings.TrimSpace(c.Long); long != "" && long != strings.TrimSpace(c.Short) {
		return false
	}
	ownFlag := false
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			ownFlag = true
		}
	})
	return !ownFlag
}

type manualPage struct {
	path []string
	out  string
	err  error
}

// printManual writes the manual: one line saying whose surface it is, how many commands it holds
// and that it is lean on purpose, then the SHARED blocks, then the surface's own page and every
// command's page in tree order.
func printManual(cmd *cobra.Command, seatID string) error {
	// NO JSON FORM, SAID RATHER THAN IGNORED. What this prints is help pages — text for a reader —
	// and a --json that was quietly dropped would hand a machine consumer prose on a channel whose
	// contract is that it parses.
	if j, _ := cmd.Flags().GetBool(flags.JSON); j {
		return feov.Errorf(feov.Validation, "%s: there is no JSON form — it prints help pages, which are text to read. Run it without the JSON flag", manualName)
	}
	root := cmd.Root()
	// The surface's own page first: it carries what no command page does — what every inherited
	// flag means, and the footer on what to do about a capability you cannot find.
	pages := []manualPage{{}}
	walkSurface(root, func(path []string, c *cobra.Command) {
		if !leftOutOfManual(c) {
			pages = append(pages, manualPage{path: path})
		}
	})

	runDir, _ := cmd.Flags().GetString(flags.Run)
	for i, p := range pages {
		// THE SEAT IS PASSED EXPLICITLY, and the run with it where this call named one. A child that
		// had to rediscover who is asking could resolve a different surface from the one this call
		// was built for; an injected run arrives through the inherited environment either way.
		argv := append(append([]string{}, p.path...), "--help", "--"+flags.SeatID, seatID)
		if runDir != "" {
			argv = append(argv, "--"+flags.Run, runDir)
		}
		pages[i].out, pages[i].err = runHelp(argv)
	}

	outs := make([]string, len(pages))
	for i, p := range pages {
		outs[i] = p.out
	}
	shared, bodies := shareRepeats(outs)

	w := cmd.OutOrStdout()
	bin := InvokedAs()
	fmt.Fprintf(w, "%s %s — the %s surface, seat %s: %d commands, each under a header naming it and followed by the help it prints, run live just now. LEAN ON PURPOSE: a block that repeats word for word on several pages is printed ONCE under SHARED below, and each page marks where it came out (`→ SHARED §n`); a group whose page would only list commands that have pages of their own is left out, and so is this command's own page. The surface's own page comes first.\n",
		bin, manualName, RoleOfSeat(seatID), seatID, len(pages)-1)
	// THE WORDS, ONCE, BEFORE ANY PAGE. The registry is the single source of the protocol's concept
	// nouns, and a seat reads a constitution, a prompt and these pages in one sitting: a word it
	// meets with no definition is one it can take for a second thing. Printed here rather than on
	// any leaf page, because each command's own --help stays that command's.
	//
	// A registry that does not load is refused, not skipped: a manual without its words would read
	// exactly like a surface that uses none.
	reg, err := terms.Load()
	if err != nil {
		return fmt.Errorf("%s: the terms registry does not load, so this surface's words cannot be printed: %w", manualName, err)
	}
	if words := reg.ForSeat(RoleOfSeat(seatID)); len(words) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, diagnostics.ManualWordsHeading)
		for _, e := range words {
			fmt.Fprintln(w, "  - "+e.Definition)
		}
	}
	if len(shared) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, diagnostics.ManualSharedHeading)
		for i, s := range shared {
			fmt.Fprintln(w, diagnostics.ManualSharedLabel(i+1, s.pages))
			fmt.Fprintln(w, s.text)
			fmt.Fprintln(w)
		}
	}

	var failed []string
	for i, p := range pages {
		fmt.Fprintln(w, diagnostics.ManualRule)
		fmt.Fprintln(w, diagnostics.ManualHeader(bin, p.path))
		if p.err != nil {
			name := strings.Join(p.path, " ")
			if name == "" {
				name = "(the surface's own page)"
			}
			failed = append(failed, name)
			fmt.Fprintf(w, "THIS PAGE DID NOT RUN, so what follows is its error rather than its help: %v\n", p.err)
		}
		fmt.Fprint(w, bodies[i])
		if bodies[i] != "" && !strings.HasSuffix(bodies[i], "\n") {
			fmt.Fprintln(w)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("%s: %d of %d help pages did not run, and each is printed above with its error: %s",
			manualName, len(failed), len(pages), strings.Join(failed, ", "))
	}
	return nil
}

// minShared is the smallest block worth lifting. A marker line costs ~30 characters on every page
// it replaces a block on, so a short repeat (`Usage:`, a one-word heading) would cost more lifted
// than left.
const minShared = 80

type sharedBlock struct {
	text  string
	pages int
}

var (
	sectionHeading = regexp.MustCompile(`^[A-Z][A-Za-z ]*:$`)
	flagEntryStart = regexp.MustCompile(`^\s+-`)
)

// shareRepeats lifts every block that repeats VERBATIM on two or more pages, returning the lifted
// blocks in order of first appearance and each page with a marker line where each one was.
//
// GENERIC BY CONSTRUCTION: nothing here names a flag, a footer or a heading. A block is a paragraph
// — lines between blank lines — or, inside a section that opens with a heading line such as
// `Flags:`, one flag entry with its continuation lines. Paragraphs are lifted first, so the whole
// Global Flags section goes as one block; a flag section that differs between pages still gives up
// the individual entries it shares (`--reason` on every verb that takes one).
func shareRepeats(pages []string) ([]sharedBlock, []string) {
	type span struct{ start, end int } // [start, end) in a page's lines
	lines := make([][]string, len(pages))
	paras := make([][]span, len(pages))
	for i, p := range pages {
		lines[i] = strings.Split(p, "\n")
		for j := 0; j < len(lines[i]); {
			if lines[i][j] == "" {
				j++
				continue
			}
			k := j
			for k < len(lines[i]) && lines[i][k] != "" {
				k++
			}
			paras[i] = append(paras[i], span{j, k})
			j = k
		}
	}
	text := func(i int, s span) string { return strings.Join(lines[i][s.start:s.end], "\n") }
	onPages := func(count map[string]map[int]bool, t string, i int) {
		if count[t] == nil {
			count[t] = map[int]bool{}
		}
		count[t][i] = true
	}
	lifts := func(count map[string]map[int]bool, t string) bool {
		return len(t) >= minShared && len(count[t]) >= 2
	}

	paraCount := map[string]map[int]bool{}
	for i := range pages {
		for _, s := range paras[i] {
			onPages(paraCount, text(i, s), i)
		}
	}
	// entriesOf splits a section paragraph into its flag entries; nil for any other paragraph.
	entriesOf := func(i int, s span) []span {
		if !sectionHeading.MatchString(lines[i][s.start]) {
			return nil
		}
		var out []span
		for j := s.start + 1; j < s.end; j++ {
			if flagEntryStart.MatchString(lines[i][j]) || len(out) == 0 {
				out = append(out, span{j, j + 1})
				continue
			}
			out[len(out)-1].end = j + 1
		}
		return out
	}
	entryCount := map[string]map[int]bool{}
	for i := range pages {
		for _, s := range paras[i] {
			if lifts(paraCount, text(i, s)) {
				continue
			}
			for _, e := range entriesOf(i, s) {
				if _, desc, ok := entryParts(text(i, e)); ok {
					onPages(entryCount, desc, i)
				}
			}
		}
	}

	var shared []sharedBlock
	ids := map[string]int{}
	idOf := func(t string, n int) int {
		if id, ok := ids[t]; ok {
			return id
		}
		shared = append(shared, sharedBlock{text: t, pages: n})
		ids[t] = len(shared)
		return ids[t]
	}
	bodies := make([]string, len(pages))
	for i := range pages {
		replace := map[int]string{} // first line of a lifted block -> its marker
		skip := map[int]bool{}      // the rest of that block's lines
		mark := func(s span, lead string, id int) {
			replace[s.start] = diagnostics.ManualMarker(lead, id)
			for j := s.start + 1; j < s.end; j++ {
				skip[j] = true
			}
		}
		for _, s := range paras[i] {
			if t := text(i, s); lifts(paraCount, t) {
				mark(s, "("+clip(strings.TrimSpace(lines[i][s.start]), 32)+") ", idOf(t, len(paraCount[t])))
				continue
			}
			for _, e := range entriesOf(i, s) {
				if prefix, desc, ok := entryParts(text(i, e)); ok && lifts(entryCount, desc) {
					mark(e, prefix, idOf(desc, len(entryCount[desc])))
				}
			}
		}
		var out []string
		for j, l := range lines[i] {
			if skip[j] {
				continue
			}
			if m, ok := replace[j]; ok {
				l = m
			}
			out = append(out, l)
		}
		bodies[i] = strings.Join(out, "\n")
	}
	return shared, bodies
}

// flagEntryHead splits a flag entry's first line into the flag's name with its exact padding, and
// the description that starts after the gap.
var flagEntryHead = regexp.MustCompile(`^(\s+-\S.*?\S\s{2,})(\S.*)$`)

// entryParts splits one flag entry into the part that stays on the page — `      --reason string   `,
// padding included, so the page still says which flag it takes and expansion is exact — and the
// description, with any continuation lines, which is what several pages share.
//
// KEYED ON THE DESCRIPTION, NOT THE LINE, because pflag pads each flag set to its own column: the
// same `--reason` said the same way is a different line on a page whose longest flag is longer, and
// keying on the line lifted it once per padding.
func entryParts(entry string) (prefix, desc string, ok bool) {
	first, rest, more := strings.Cut(entry, "\n")
	m := flagEntryHead.FindStringSubmatch(first)
	if m == nil {
		return "", "", false
	}
	desc = m[2]
	if more {
		desc += "\n" + rest
	}
	return m[1], desc, true
}

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// runHelp runs this tool with argv and returns everything it printed. It is a variable so a test
// can stand a failing one in; production never reassigns it.
//
// Assigned in init rather than at the declaration: the executor builds the tree, the tree mounts
// manual, and manual calls the executor, which Go refuses as an initialization cycle.
var runHelp func(argv []string) (string, error)

func init() { runHelp = selfExecHelp }

// selfExecHelp runs THIS BINARY again with argv — the real help, from the real tool, at call time.
//
// NOT INSIDE A TEST BINARY, and the reason is not speed. os.Executable() there is the test binary,
// and running it with a command path does not print that command's help — it runs the whole suite
// again, which calls manual again. The fuzz harness drives commands in-process through the same
// tree, so a test binary reaches this path without any test asking for it. In one, the pages come
// from the same tree built in-process, which is what the child would have built.
func selfExecHelp(argv []string) (string, error) {
	if testing.Testing() {
		return inProcessHelp(argv)
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot locate this binary to run its help: %w", err)
	}
	c := exec.Command(exe, argv...)
	// ARGV[0] IS KEPT AS THIS PROCESS HAD IT, because InvokedAs reads it and every usage line on the
	// page names the tool through it. os.Executable resolves symlinks, so without this a tool run
	// through a link would print pages naming a program the seat never typed.
	c.Args[0] = os.Args[0]
	var out bytes.Buffer
	c.Stdout, c.Stderr = &out, &out
	if err := c.Run(); err != nil {
		return out.String(), fmt.Errorf("`%s` failed: %w", strings.Join(argv, " "), err)
	}
	return out.String(), nil
}

// inProcessHelp builds the tree for the seat argv names and runs argv against it — what the binary
// does around Execute, minus the process: the pre-parse refusal, the same ExecuteRoot, and the
// error line Execute prints.
func inProcessHelp(argv []string) (string, error) {
	seatID := seatenv.SeatIDIn(argv)
	root := NewRootFor(seatID)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(argv)
	err := refuseUnknownCommandFirst(root, append([]string{InvokedAs()}, argv...), seatID)
	if err == nil {
		err = ExecuteRoot(root)
	}
	if err != nil {
		fmt.Fprintf(&out, "%s: %v\n", InvokedAs(), err)
	}
	return out.String(), err
}
