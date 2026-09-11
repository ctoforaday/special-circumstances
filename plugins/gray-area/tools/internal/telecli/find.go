package telecli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// errNoRipgrep is the refusal, held as a value so a test can assert THIS failure rather than
// matching prose that a reword would silently detach the test from.
var errNoRipgrep = errors.New(
	"ripgrep is not installed, so this search cannot run. Refusing rather than reporting zero " +
		"matches, which would be indistinguishable from a real result. Install it (see " +
		"gray-area's requirements.json) or use `telepathy sql`")

// samplesPerFile caps how many matching records are decoded per transcript. A term like "the"
// matches most of the corpus, and decoding all of it would turn a 100 ms search into a minute.
// The cap is STATED when it bites, because a truncated answer presented as a whole one is the
// defect this verb keeps finding in itself.
const samplesPerFile = 200

func newFindCmd(env *Env) *cobra.Command {
	var asRegex, showPaths bool
	var only string
	c := &cobra.Command{
		Use:   "find <term>",
		Short: "search every local transcript (ripgrep, no index)",
		Long: `Searches the raw transcripts the store names, and reports each session's most
recent hit with the text around it.

One row per session and agent, newest first, because the question behind almost
every search here is "who is working on this NOW" — and a list of file paths
ordered by path answers that only after several more queries.

  HITS     matching records in that transcript
  WHEN     the timestamp of the most recent one
  IN       where it landed: assistant, user, peer, notification, lead, harness,
           thinking, tool_use, result, unknown_origin, or ?

IN matters more than it looks, and --in filters on it. This searches the WHOLE
transcript, while the store's word and thought tiers hold only speech and
reasoning — so a hit can be a tool result or a file path. A match in a pasted
log is not somebody saying something.

For text, IN is WHO SPOKE, read from fields the client writes and never from
the text: peer is another session's message, notification a background task's,
lead the lead or a workflow coordinator prompting a seat, harness text the
client injects. A message delivered mid-turn reports its sender the same way.
user is the human — plus what no field separates from them: prompts programs
send to headless sessions, and slash-command and local-command text.

  ?               the term is in a part of the record nothing here models
                  (a cwd, a uuid, queue bookkeeping)
  unknown_origin  a record whose origin, or whose mid-turn delivery mode, this
                  binary does not know — upgrade gray-area

Neither is the liveness word "unknown" that agents prints.

--in keeps a transcript if ANY of its hits is in that channel, and the row
shows the most recent of those; on a term that appears in a seat prompt most
unfiltered rows are exactly that:

  telepathy find 'bench rul' --in assistant   what agents SAID about it

A filtered row still reports the unfiltered HITS count, so narrowing the view
never quietly shrinks the number beside it.

No full-text index is built. A whole-corpus search measures tens of milliseconds
against a corpus of hundreds of megabytes, where an index would cost gigabytes and
need invalidating on every append — the filesystem is already the index, and it
cannot go stale.

The term is a LITERAL by default, and is passed to ripgrep as a single argument
that never reaches a shell. Pass --regex to treat it as a pattern — which is what
you want for a word boundary, and searching for 'roving' rather than '\broving\b'
is how you get every occurrence of "proving" instead.

Anything that stops this from being a completed search REFUSES rather than
printing "no matches": ripgrep absent, a pattern that does not compile, or a
corpus it could not finish reading. A search that could not run must never be
reported as a search that found nothing.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			term := args[0]
			db, err := env.openRead(cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer db.Close()
			paths, err := catalogue.PathsFor(cmd.Context(), db)
			if err != nil {
				return err
			}
			if len(paths) == 0 {
				return errors.New("the catalogue names no transcripts yet — run `telepathy backfill`")
			}
			rg, err := exec.LookPath("rg")
			if err != nil {
				return errNoRipgrep
			}

			// THE ARGV IS SORTED, and that is where the determinism comes from. `paths` is a map,
			// so building the argv by ranging it hands ripgrep a different command line every
			// run — and then the output order, and anything a caller diffs against yesterday's,
			// is an accident of Go's map seed.
			names := make([]string, 0, len(paths))
			for p := range paths {
				names = append(names, p)
			}
			sort.Strings(names)

			// --line-number, not --count-matches: the line NUMBER is what lets the second pass
			// read the record back and say when it was and what it was.
			rgArgs := []string{"--line-number", "--no-heading", "--with-filename", "--only-matching"}
			if !asRegex {
				rgArgs = append(rgArgs, "--fixed-strings")
			}
			rgArgs = append(rgArgs, "--", term)
			rgArgs = append(rgArgs, names...)

			var stderr bytes.Buffer
			run := exec.CommandContext(cmd.Context(), rg, rgArgs...)
			run.Stderr = &stderr
			raw, runErr := run.Output()
			code := 0
			var ee *exec.ExitError
			if errors.As(runErr, &ee) {
				code = ee.ExitCode()
			} else if runErr != nil {
				return fmt.Errorf("could not run ripgrep: %w", runErr)
			}
			// RIPGREP'S THREE EXIT CODES ARE THREE DIFFERENT ANSWERS, and collapsing them is how
			// this verb produced the exact zero it was written to refuse:
			//
			//	0  matched
			//	1  ran to completion, matched nothing        <- the honest zero
			//	2  did not run, or did not finish            <- NOT a zero
			var incomplete string
			if code >= 2 {
				why := strings.TrimSpace(stderr.String())
				if why == "" {
					why = "ripgrep exited " + strconv.Itoa(code) + " without saying why"
				}
				if len(raw) == 0 {
					return fmt.Errorf("the search did not run: %s", why)
				}
				// Partial results are still reported — exit 2 also covers one transcript vanishing
				// mid-walk — but never as if the search were complete.
				incomplete = why
			}

			out := cmd.OutOrStdout()
			locs := parseRipgrep(string(raw))
			if len(locs) == 0 {
				return reportEmpty(out, term, names, asRegex)
			}
			if incomplete != "" {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"telepathy find: INCOMPLETE — ripgrep could not finish, so these hits are a "+
						"floor and not the answer: %s\n", incomplete)
			}
			// REFUSE A CHANNEL THAT CANNOT EXIST. `--in asistant` would otherwise filter every row
			// away and report "none in that channel" — a typo returned as a finding about the
			// corpus, in the same words a real empty result uses.
			if only != "" && !catalogue.ValidChannel(catalogue.Channel(only)) {
				return usagef("--in %q is not a channel; one of: %s", only, channelList())
			}
			rows, matched := decodeHits(locs, paths, term, catalogue.Channel(only))
			if only != "" && len(rows) == 0 {
				// COUNTED IN THE UNIT IT NAMES. `locs` counts matching RECORDS, not transcripts,
				// and printing one as the other is how a row count becomes a subject count.
				fmt.Fprintf(out, "%d transcript(s) contain %q, but none has a hit in %q\n",
					matched, term, only)
				return nil
			}
			render(out, rows, showPaths)
			return nil
		},
	}
	c.Flags().BoolVar(&asRegex, "regex", false,
		"treat the term as a regular expression rather than a literal (needed for \\b word boundaries)")
	c.Flags().StringVar(&only, "in", "",
		"keep only transcripts with ANY hit in this channel, showing the most recent of those: "+channelList())
	c.Flags().BoolVar(&showPaths, "paths", false,
		"print the transcript path under each row, for citing the record rather than summarising it")
	return c
}

// fileLine is one ripgrep hit location.
type fileLine struct {
	path string
	line int
}

// parseRipgrep reads `path:lineno:match` and returns each DISTINCT record, in order.
//
// Distinct, because --only-matching emits one output line per match and a single transcript record
// can hold the term many times. Collapsing here is what makes HITS mean "turns that mention this"
// rather than "times these characters appear", which is the number a reader is actually after: one
// transcript line is one turn, and three mentions inside it are one moment.
func parseRipgrep(raw string) []fileLine {
	var out []fileLine
	seen := map[fileLine]bool{}
	sc := bufio.NewScanner(strings.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		s := sc.Text()
		// path:line:match — the path may contain no ':' on the platforms this runs on, and the
		// match certainly can, so both splits are from the LEFT.
		i := strings.IndexByte(s, ':')
		if i < 0 {
			continue
		}
		j := strings.IndexByte(s[i+1:], ':')
		if j < 0 {
			continue
		}
		n, err := strconv.Atoi(s[i+1 : i+1+j])
		if err != nil {
			continue
		}
		fl := fileLine{path: s[:i], line: n}
		if !seen[fl] {
			seen[fl] = true
			out = append(out, fl)
		}
	}
	return out
}

func channelList() string {
	names := make([]string, 0, len(catalogue.Channels))
	for _, c := range catalogue.Channels {
		names = append(names, string(c))
	}
	return strings.Join(names, ", ")
}

// channelWidth is the IN column's width, DERIVED from the longest channel name. `%-9s` pads but
// never truncates, so a hand-kept width silently ragged-edges SNIPPET the day a longer channel is
// added — which is what `notification` and `unknown_origin` would have done to a width sized for
// `tool_use`.
func channelWidth() int {
	w := len("IN")
	for _, c := range catalogue.Channels {
		w = max(w, len(c))
	}
	return w
}

// row is one session/agent's showing for a term.
type row struct {
	f      catalogue.TranscriptFile
	hits   int
	best   catalogue.Hit
	capped bool
}

// decodeHits reads the matching records back so each row can carry a time, a channel and a
// snippet, and returns the rows with the number of transcripts that matched at all.
//
// UNFILTERED (want == ""), a row describes its newest hit, read from at most samplesPerFile of the
// newest records. FILTERED, every matching record is read and the row describes the newest hit IN
// want; a transcript with none is dropped. Filtering on the newest hit alone answered "whose most
// recent mention was in C" — an earlier match in C vanished, and the empty result read as though
// C had never said it. Nothing is capped under a filter, because a cap there is the same drop.
//
// The row's HITS count is deliberately NOT narrowed: it counts every matching record in that
// transcript either way, so a filtered view cannot make a session look quieter than it is.
func decodeHits(locs []fileLine, paths map[string]catalogue.TranscriptFile, term string, want catalogue.Channel) ([]row, int) {
	byFile := map[string][]int{}
	var order []string
	for _, l := range locs {
		if _, ok := byFile[l.path]; !ok {
			order = append(order, l.path)
		}
		byFile[l.path] = append(byFile[l.path], l.line)
	}
	var rows []row
	matched := 0
	for _, path := range order {
		f, ok := paths[path]
		if !ok {
			continue // a file ripgrep saw and the store does not name; not ours to report
		}
		matched++
		lines := byFile[path]
		r := row{f: f, hits: len(lines)}
		read := lines
		if want == "" && len(read) > samplesPerFile {
			// The tail of a transcript is its recent end, and this row reports the most recent
			// hit — so when the cap bites it must bite on the OLD end.
			read = read[len(read)-samplesPerFile:]
			r.capped = true
		}
		found := false
		for _, h := range recordsAt(path, read, term, f.AgentID != "") {
			if want != "" && h.Channel != want {
				continue
			}
			if !found || h.TS >= r.best.TS {
				r.best, found = h, true
			}
		}
		if found {
			rows = append(rows, r)
		}
	}
	// NEWEST FIRST — the whole point of the change. Ties break on path so the output is stable
	// for a golden and diffable against yesterday's.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].best.TS != rows[j].best.TS {
			return rows[i].best.TS > rows[j].best.TS
		}
		return rows[i].f.Path < rows[j].f.Path
	})
	return rows, matched
}

// recordsAt reads the given 1-indexed lines of a file and decodes each. inSubagent is whether the
// file is a subagent or workflow transcript, which decides who a no-origin prompt came from.
//
// One sequential pass, never a seek per line: a transcript runs to hundreds of megabytes, the
// wanted lines are already ascending, and this stops at the last one it needs.
func recordsAt(path string, lines []int, term string, inSubagent bool) []catalogue.Hit {
	want := map[int]bool{}
	last := 0
	for _, n := range lines {
		want[n] = true
		if n > last {
			last = n
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil // a transcript that vanished between the search and the read
	}
	defer f.Close()
	var out []catalogue.Hit
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 256*1024), 64*1024*1024)
	for n := 1; sc.Scan(); n++ {
		if want[n] {
			h, _ := catalogue.DecodeHit(sc.Text(), term, inSubagent)
			out = append(out, h)
		}
		if n >= last {
			break
		}
	}
	return out
}

func render(out io.Writer, rows []row, showPaths bool) {
	in := channelWidth()
	fmt.Fprintf(out, "%-10s %-18s %5s  %-12s %-*s %s\n",
		"SESSION", "AGENT", "HITS", "WHEN", in, "IN", "SNIPPET")
	capped := false
	for _, r := range rows {
		agent := r.f.AgentID
		if agent == "" {
			agent = "(main)"
		}
		// A record with no usable timestamp is shown as one, not as the epoch — 1970 beside a
		// real date reads as data rather than as the absence of it.
		when := "—"
		if r.best.TS > 0 {
			when = time.Unix(r.best.TS, 0).UTC().Format("01-02 15:04")
		}
		fmt.Fprintf(out, "%-10s %-18s %5d  %-12s %-*s %s\n",
			catalogue.Short(r.f.SessionID), agent, r.hits, when, in, r.best.Channel, r.best.Snippet)
		if showPaths {
			fmt.Fprintf(out, "%-10s %s\n", "", r.f.Path)
		}
		capped = capped || r.capped
	}
	if capped {
		fmt.Fprintf(out, "\n(HITS is exact; only the most recent %d per transcript were read for WHEN and SNIPPET)\n",
			samplesPerFile)
	}
}

// humanBytes picks a unit the number survives. "0 MB" is what a small corpus rounded to, and a
// size that reads as zero defeats the entire point of stating one.
func humanBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.0f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d bytes", n)
	}
}

// reportEmpty gives the zero a SIZE.
//
// "no transcript contains X" is a fact about the term AND a fact about the corpus, and the old
// message carried only the first. A reader could not tell a search of four hundred transcripts
// from a search of none — and a search of none is what a half-built store produces.
func reportEmpty(out io.Writer, term string, names []string, asRegex bool) error {
	var searched int64
	for _, p := range names {
		if st, err := os.Stat(p); err == nil {
			searched += st.Size()
		}
	}
	fmt.Fprintf(out, "no transcript contains %q — searched %d transcripts, %s\n",
		term, len(names), humanBytes(searched))
	if !asRegex {
		fmt.Fprintln(out, "  the term was matched LITERALLY; pass --regex if you meant a pattern")
	}
	return nil
}
