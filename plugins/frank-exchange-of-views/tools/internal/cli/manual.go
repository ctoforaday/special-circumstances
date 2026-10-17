package cli

import (
	"bytes"
	"cmp"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/diagnostics"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
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
// The surface is the one this tree was built for — the flag, else the seat the record binds this
// agent to, exactly as every invocation resolves it — so seatID is captured at construction rather than
// re-derived. It records nothing.
func newManual(seatID string) *cobra.Command {
	var forSeat string
	c := &cobra.Command{
		Use:   manualName,
		Short: "render a surface's commands, each with its own help, run live — for generating a seat's constitution",
		Long: manualName + " prints every command on this surface, in tree order, each under a header naming the command exactly, followed by the help that command prints. The help is produced now, by running this binary with that command and --help, so it cannot disagree with the tool you are running. It records nothing.\n\n" +
			"It is lean on purpose. A block that repeats word for word on several pages — a footer, a flag's description several commands share, the sentences that close several descriptions — is printed ONCE, in a SHARED section before the first page, and each page carries a marker ending `→ SHARED §n` where the block was. The flags every command inherits are the last block of every page, so they are printed once and marked only on a page that does not end with them. What every page would say the same way — a Usage that only restates the command, the help flag — is left off and said once. A group whose page only lists commands that have pages of their own is left out. Nothing a page said is lost: put each block and each left-off line back and you have that command's --help exactly, and a manual that does not read back that way is refused rather than printed.\n\n" +
			"Before the pages it prints WORDS THIS SURFACE USES: each concept word your surface uses, defined once — the one word every page, prompt and constitution uses for that concept.\n\n" +
			"THIS IS A GENERATOR'S COMMAND, NOT A SEAT'S. A seat does not run it and cannot: every command on a seat's surface, with this same help, is written into that seat's constitution by scripts/agentgen, so the seat is already holding what this prints. Pass --for <seat> to render a surface other than your own, which is how that generation is done. A page whose help fails to run is printed with its error, never left out, and the command then exits non-zero.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// --for RENDERS ANOTHER SEAT'S SURFACE, which is the whole reason this command still
			// exists. Its readers are generators, not seats: scripts/agentgen writes each seat's
			// surface into that seat's constitution, and it cannot ask the seat to render it —
			// the seat has no `manual` to run.
			if forSeat != "" {
				return printManualFor(cmd, NewRootFor(forSeat), forSeat)
			}
			return printManualFor(cmd, cmd.Root(), seatID)
		},
	}
	c.Flags().StringVar(&forSeat, flags.For, "", "render THIS seat's surface instead of your own — for generating a seat's documentation, since a seat cannot render its own")
	return c
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
// and that it is lean on purpose, then the words, what every page leaves off, the SHARED blocks,
// then the surface's own page and every command's page in tree order.
func printManualFor(cmd *cobra.Command, root *cobra.Command, seatID string) error {
	// NO JSON FORM, SAID RATHER THAN IGNORED. What this prints is help pages — text for a reader —
	// and a --json that was quietly dropped would hand a machine consumer prose on a channel whose
	// contract is that it parses.
	if j, _ := cmd.Flags().GetBool(flags.JSON); j {
		return feov.Errorf(feov.Validation, "there is no JSON form — it prints help pages, which are text to read. Run it without the JSON flag")
	}
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

	bin := InvokedAs()
	nameOf := func(p manualPage) string {
		if len(p.path) == 0 {
			return "(the surface's own page)"
		}
		return strings.Join(p.path, " ")
	}
	outs := make([]string, len(pages))
	asPrinted := make([]bool, len(pages))
	for i, p := range pages {
		// A PAGE THAT DID NOT RUN IS PRINTED AS IT CAME. Its text is an error and whatever the
		// command got out before it, not a help page, so there is no Usage or help flag to leave off
		// and nothing the reader should put back.
		if outs[i], asPrinted[i] = p.out, p.err != nil; asPrinted[i] {
			continue
		}
		outs[i] = diagnostics.LeanManualPage(bin, p.path, p.out)
	}
	shared, bodies := shareRepeats(outs, asPrinted)

	var b bytes.Buffer
	fmt.Fprintf(&b, "%s %s — the %s surface, seat %s: %d commands, each under a header naming it and followed by the help it prints, run live just now. LEAN ON PURPOSE: a block that repeats word for word on several pages is printed ONCE under SHARED below, and each page marks where it came out (`→ SHARED §n`); what every page would say the same way is said once, before SHARED; a group whose page would only list commands that have pages of their own is left out, and so is this command's own page. The surface's own page comes first.\n",
		bin, manualName, seatclass.RoleOfSeat(seatID), seatID, len(pages)-1)
	// THE WORDS, ONCE, BEFORE ANY PAGE. The registry is the single source of the protocol's concept
	// nouns, and a seat reads a constitution, a prompt and these pages in one sitting: a word it
	// meets with no definition is one it can take for a second thing. Printed here rather than on
	// any leaf page, because each command's own --help stays that command's.
	//
	// A registry that does not load is refused, not skipped: a manual without its words would read
	// exactly like a surface that uses none.
	reg, err := terms.Load()
	if err != nil {
		return fmt.Errorf("the terms registry does not load, so this surface's words cannot be printed: %w", err)
	}
	if words := reg.ForSeat(seatclass.RoleOfSeat(seatID)); len(words) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, diagnostics.ManualWordsHeading)
		for _, e := range words {
			fmt.Fprintln(&b, "  - "+e.Definition)
		}
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, diagnostics.ManualLeftOff)
	if len(shared) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, diagnostics.ManualSharedHeading)
		for i, s := range shared {
			if s.closes > 0 {
				fmt.Fprintln(&b, diagnostics.ManualClosingLabel(i+1, s.closes))
			} else {
				fmt.Fprintln(&b, diagnostics.ManualSharedLabel(i+1, s.marks))
			}
			fmt.Fprintln(&b, s.text)
			fmt.Fprintln(&b)
		}
	}

	var failed []string
	for i, p := range pages {
		fmt.Fprintln(&b, diagnostics.ManualRule)
		fmt.Fprintln(&b, diagnostics.ManualHeader(bin, p.path))
		if p.err != nil {
			failed = append(failed, nameOf(p))
			fmt.Fprintf(&b, diagnostics.ManualFailedLead+", so what follows is its error rather than its help: %v\n", p.err)
		}
		fmt.Fprint(&b, bodies[i])
		if bodies[i] != "" && !strings.HasSuffix(bodies[i], "\n") {
			fmt.Fprintln(&b)
		}
	}

	// NOTHING IS PRINTED THAT CANNOT BE PUT BACK. The pages above stand for each command's help, and
	// the reader's own expansion is run on them here, against the help that was just run: a lifted
	// block, a closing block or a left-off line that does not come back exactly is refused before a
	// seat is handed it, rather than found by a survey reading a page that has quietly lost a line.
	expanded, err := diagnostics.ExpandManual(b.String(), bin)
	if err != nil {
		return fmt.Errorf("the manual does not expand back into its pages: %w", err)
	}
	if len(expanded) != len(pages) {
		return fmt.Errorf("the manual holds %d pages and reads back as %d", len(pages), len(expanded))
	}
	for i, p := range pages {
		want := p.out
		if want != "" && !strings.HasSuffix(want, "\n") {
			want += "\n"
		}
		if p.err == nil && expanded[i].Body != want {
			return fmt.Errorf("the page for %s does not read back as the help that command prints, so the manual is not printed", nameOf(p))
		}
	}
	if _, err := cmd.OutOrStdout().Write(b.Bytes()); err != nil {
		return err
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d of %d help pages did not run, and each is printed above with its error: %s",
			len(failed), len(pages), strings.Join(failed, ", "))
	}
	return nil
}

// minShared is the smallest block worth lifting. A marker line costs ~30 characters on every page
// it replaces a block on, so a short repeat (`Usage:`, a one-word heading) would cost more lifted
// than left.
const minShared = 80

// minTailEntries is how many flag entries must close with the same sentences before those
// sentences are lifted. A whole description lifts from two; a tail needs three, because its marker
// leaves a description in two places — the flag's own sentences on the page, the shared ones under
// SHARED — and on two entries that buys about forty characters.
const minTailEntries = 3

type sharedBlock struct {
	text   string
	marks  int // the markers that stand for it, over all pages
	closes int // the pages it closes with no marker — non-zero for at most one block
}

var (
	sectionHeading = regexp.MustCompile(`^[A-Z][A-Za-z ]*:$`)
	flagEntryStart = regexp.MustCompile(`^\s+-`)
)

// shareRepeats lifts every block that repeats VERBATIM on two or more pages, returning the lifted
// blocks in order of first appearance and each page with a marker where each one was.
//
// GENERIC BY CONSTRUCTION: nothing here names a flag or a footer, and it names one heading only —
// a command's own `Flags:` section is never lifted whole, so every page still lists the flags its
// command takes. A block is one of
//
//   - a paragraph — lines between blank lines — or a run of paragraphs that are never apart: where
//     one always follows another, the two are lifted as ONE block under one marker;
//   - inside a section that opens with a heading line such as `Flags:`, one flag entry's description
//     with its continuation lines, or the closing sentences of it that several entries share — the
//     sentences only that entry says stay on its line;
//   - the entries of a lifted section, where a page lists them under a heading of its own.
//
// Paragraphs are lifted first, so the whole Global Flags section goes as one block; a flag section
// that differs between pages still gives up the individual entries it shares (`--reason` on every
// verb that takes one).
//
// ONE BLOCK CLOSES ITS PAGES UNMARKED: the lifted paragraph that is the last block of the most
// pages is taken off them with no marker, and its label says so. It is chosen only where every page
// either ends with it or marks it, so "a page with no marker for it ends with it" has no exception;
// asPrinted names the pages exempt from that — a page that did not run keeps a marker.
func shareRepeats(pages []string, asPrinted []bool) ([]sharedBlock, []string) {
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

	// PARAGRAPHS THAT ARE NEVER APART ARE ONE BLOCK. Where every appearance of a lifted paragraph is
	// followed, after one blank line, by the same lifted paragraph, and that one is never seen
	// without it, the two spans are joined and counted again — so a chain of three closes in two
	// passes, and the loop ends when a pass joins nothing.
	const apart = "\x00"
	var paraCount map[string]map[int]bool
	for {
		paraCount = map[string]map[int]bool{}
		for i := range pages {
			for _, s := range paras[i] {
				// A PAGE KEEPS ITS OWN FLAG LIST. Two commands with the same flags print the same
				// section, and lifted whole it would leave each page a marker where its flags were
				// — a page that no longer says what its command takes. Left uncounted it never
				// lifts, and its entries give up their descriptions one by one, names staying.
				if lines[i][s.start] == diagnostics.ManualFlagsHeading {
					continue
				}
				onPages(paraCount, text(i, s), i)
			}
		}
		after, before := map[string]string{}, map[string]string{}
		note := func(m map[string]string, t, v string) {
			if old, seen := m[t]; seen && old != v {
				v = apart
			}
			m[t] = v
		}
		for i := range pages {
			for k, s := range paras[i] {
				t := text(i, s)
				if !lifts(paraCount, t) {
					continue
				}
				next, prev := apart, apart
				if k+1 < len(paras[i]) && paras[i][k+1].start == s.end+1 {
					if n := text(i, paras[i][k+1]); lifts(paraCount, n) && n != t {
						next = n
					}
				}
				if k > 0 && paras[i][k-1].end+1 == s.start {
					if p := text(i, paras[i][k-1]); lifts(paraCount, p) && p != t {
						prev = p
					}
				}
				note(after, t, next)
				note(before, t, prev)
			}
		}
		joined := false
		for i := range pages {
			var out []span
			for k := 0; k < len(paras[i]); k++ {
				s := paras[i][k]
				if next, ok := after[text(i, s)]; ok && next != apart && before[next] == text(i, s) {
					s.end = paras[i][k+1].end
					k++
					joined = true
				}
				out = append(out, s)
			}
			paras[i] = out
		}
		if !joined {
			break
		}
	}

	// A lifted section — a heading line and its entries — can also be met as entries under another
	// page's own heading. In order of first appearance, so the output is the same every run.
	var sections []string
	for i := range pages {
		for _, s := range paras[i] {
			t := text(i, s)
			_, entries, _ := strings.Cut(t, "\n")
			if lifts(paraCount, t) && sectionHeading.MatchString(lines[i][s.start]) && len(entries) >= minShared && !slices.Contains(sections, t) {
				sections = append(sections, t)
			}
		}
	}

	// A mark is one replacement on a page: the lines it takes out, what stays before the arrow, and
	// the block the arrow names.
	type mark struct {
		span
		lead, block string
	}
	marks := make([][]mark, len(pages))
	// entriesOf splits a section paragraph into its flag entries, leaving out the lines a mark
	// already covers; nil for any other paragraph.
	entriesOf := func(i int, s span, taken func(int) bool) []span {
		if !sectionHeading.MatchString(lines[i][s.start]) {
			return nil
		}
		var out []span
		open := false
		for j := s.start + 1; j < s.end; j++ {
			if taken(j) {
				open = false
				continue
			}
			if flagEntryStart.MatchString(lines[i][j]) || !open {
				out = append(out, span{j, j + 1})
				open = true
				continue
			}
			out[len(out)-1].end = j + 1
		}
		return out
	}
	type entry struct {
		page int
		span
		prefix, desc string
	}
	var entries []entry
	entryCount := map[string]map[int]bool{}
	for i := range pages {
		for _, s := range paras[i] {
			if t := text(i, s); lifts(paraCount, t) {
				marks[i] = append(marks[i], mark{s, "(" + clip(strings.TrimSpace(lines[i][s.start]), 32) + ") ", t})
				continue
			}
			var runs []span
			if sectionHeading.MatchString(lines[i][s.start]) {
				for _, sec := range sections {
					want := strings.Split(sec, "\n")[1:]
					for j := s.start + 1; j+len(want) <= s.end; j++ {
						if slices.Equal(lines[i][j:j+len(want)], want) {
							runs = append(runs, span{j, j + len(want)})
							marks[i] = append(marks[i], mark{span{j, j + len(want)}, diagnostics.ManualEntriesLead, sec})
							break
						}
					}
				}
			}
			taken := func(j int) bool {
				return slices.ContainsFunc(runs, func(r span) bool { return r.start <= j && j < r.end })
			}
			for _, e := range entriesOf(i, s, taken) {
				if prefix, desc, ok := entryParts(text(i, e)); ok {
					entries = append(entries, entry{i, e, prefix, desc})
					onPages(entryCount, desc, i)
				}
			}
		}
	}
	// A description several pages give whole is lifted whole. Of the rest, the closing sentences
	// several share are lifted, longest first, each entry giving up one tail at most.
	tails := map[string][]int{} // a closing run of sentences -> the entries that end with it
	for n, e := range entries {
		if lifts(entryCount, e.desc) {
			marks[e.page] = append(marks[e.page], mark{e.span, e.prefix, e.desc})
			continue
		}
		for _, at := range sentenceStarts(e.desc) {
			if tail := e.desc[at:]; len(tail) >= minShared {
				tails[tail] = append(tails[tail], n)
			}
		}
	}
	byLength := slices.SortedFunc(maps.Keys(tails), func(a, b string) int {
		return cmp.Or(cmp.Compare(len(b), len(a)), cmp.Compare(a, b))
	})
	gaveUp := map[int]bool{}
	for _, tail := range byLength {
		var free []int
		for _, n := range tails[tail] {
			if !gaveUp[n] {
				free = append(free, n)
			}
		}
		if len(free) < minTailEntries {
			continue
		}
		for _, n := range free {
			e := entries[n]
			gaveUp[n] = true
			marks[e.page] = append(marks[e.page], mark{e.span, e.prefix + strings.TrimSuffix(e.desc, tail), tail})
		}
	}
	for i := range marks {
		slices.SortFunc(marks[i], func(a, b mark) int { return cmp.Compare(a.start, b.start) })
	}

	// THE CLOSING BLOCK. closes[i] is the paragraph mark page i gives up for it.
	closing, closes := closingBlock(len(pages), asPrinted, func(i int) (string, bool) {
		if len(paras[i]) == 0 {
			return "", false
		}
		s := paras[i][len(paras[i])-1]
		// Exactly one blank line before it and the page's final newline after it, which is what
		// the reader puts back around it.
		tidy := s.start >= 2 && lines[i][s.start-1] == "" && lines[i][s.start-2] != "" &&
			s.end == len(lines[i])-1
		return text(i, s), tidy && lifts(paraCount, text(i, s))
	}, func(i int, t string) int {
		n := 0
		for _, m := range marks[i] {
			if m.block == t {
				n++
			}
		}
		return n
	})

	var shared []sharedBlock
	ids := map[string]int{}
	idOf := func(t string) int {
		if _, ok := ids[t]; !ok {
			shared = append(shared, sharedBlock{text: t})
			ids[t] = len(shared)
		}
		return ids[t]
	}
	bodies := make([]string, len(pages))
	for i := range pages {
		replace := map[int]string{} // first line of a lifted block -> its marker
		skip := map[int]bool{}      // the rest of that block's lines
		for _, m := range marks[i] {
			id := idOf(m.block)
			if closes[i] && m.block == closing {
				// Taken off with the blank line before it, and no marker.
				shared[id-1].closes++
				for j := m.start - 1; j < m.end; j++ {
					skip[j] = true
				}
				continue
			}
			shared[id-1].marks++
			replace[m.start] = diagnostics.ManualMarker(m.lead, id)
			for j := m.start + 1; j < m.end; j++ {
				skip[j] = true
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

// closingBlock picks the block that closes pages unmarked, and says which pages it closes.
//
// last(i) is page i's last paragraph and whether it is a lifted one the reader can put back at the
// page's end; marksFor(i, t) is how many marks page i carries for block t. The block is the one that
// is last on the most pages — two at least, the earliest page winning a tie — and it is taken only
// where the rule it states has no exception: every page that is not printed as it came either ends
// with the block or marks it. A page that ends with it AND marks it elsewhere keeps its marker.
func closingBlock(pages int, asPrinted []bool, last func(int) (string, bool), marksFor func(int, string) int) (string, []bool) {
	none := make([]bool, pages)
	ends := map[string]int{}
	best := ""
	for i := range pages {
		if t, ok := last(i); ok && !asPrinted[i] {
			ends[t]++
		}
	}
	for i := range pages {
		if t, ok := last(i); ok && !asPrinted[i] && ends[t] >= 2 && ends[t] > ends[best] {
			best = t
		}
	}
	if best == "" {
		return "", none
	}
	closes := make([]bool, pages)
	for i := range pages {
		if asPrinted[i] {
			continue
		}
		t, ok := last(i)
		n := marksFor(i, best)
		closes[i] = ok && t == best && n == 1
		if n == 0 {
			return "", none
		}
	}
	return best, closes
}

// sentenceStarts is every offset in a description where a tail may be cut off it: where a later
// sentence opens — after a full stop and a space, at a capital letter — and after a spaced dash,
// which these descriptions use to hang an explanation on a clause that is already complete. A tail
// lifted from either leaves on the page words that stand on their own, and puts under SHARED words
// that read on from them; a cut anywhere else would leave half a clause in each place.
func sentenceStarts(desc string) []int {
	const dash = " — "
	var out []int
	for at := 2; at < len(desc); at++ {
		if desc[at-2] == '.' && desc[at-1] == ' ' && desc[at] >= 'A' && desc[at] <= 'Z' {
			out = append(out, at)
		}
		if at >= len(dash) && desc[at-len(dash):at] == dash {
			out = append(out, at)
		}
	}
	return out
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
// does around Execute, minus the process: Invoke, which is the binary's own sequence.
func inProcessHelp(argv []string) (string, error) {
	seatID := seatenv.SeatIDIn(argv)
	root := NewRootFor(seatID)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	_, err := Invoke(root, argv, seatID)
	return out.String(), err
}
