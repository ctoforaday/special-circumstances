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

// errStaleZero refuses a zero it could not measure: no row survived, and some matching records
// could not be read back as ripgrep matched them. Printing "none has a hit" or an empty table
// there would report the records find FAILED to read as records that were read and did not match.
var errStaleZero = errors.New(
	"no row survives, and some matching records could not be read back as ripgrep matched them, " +
		"so this is not a completed search. Refusing rather than reporting no hits; run it again " +
		"once the transcripts stop changing")

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

IN is read from the JSON value ripgrep's match lies in, at the byte it
reported — the same for --regex as for a literal. tool_use is a match anywhere
inside a call's arguments, result one anywhere inside what a tool returned;
SNIPPET is centred on the match.

For text, IN is WHO SPOKE, read from fields the client writes and never from
the text: peer is another session's message, notification a background task's,
lead the lead or a workflow coordinator prompting a seat, harness text the
client injects. A message delivered mid-turn reports its sender the same way.
user is the human — plus what no field separates from them: prompts programs
send to headless sessions, and slash-command and local-command text.

  ?               the match is in a part of the record nothing here models
                  (a cwd, a uuid, a key name, queue bookkeeping)
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

The term is matched against the raw JSON of each line, so a term containing "
or \ does not match text that holds it (the transcript stores \" and \\), and
ripgrep refuses a term containing a line break. ripgrep configuration files
are ignored, and transcripts are read as raw bytes: no encoding detection, and
a byte-order mark is not skipped.

Anything that stops this from being a completed search REFUSES rather than
printing "no matches": ripgrep absent, a pattern that does not compile, a
corpus it could not finish reading, or no row surviving while some matching
records could not be read back as ripgrep matched them (the file changed or
went away since the search, or the record no longer parses). When rows do
survive, one line on stderr counts those records. A search that could not run
must never be reported as a search that found nothing.`,
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

			var stderr bytes.Buffer
			run := exec.CommandContext(cmd.Context(), rg, ripgrepArgs(term, asRegex, names)...)
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
			recs := parseRipgrep(string(raw))
			if len(recs) == 0 {
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
			rows, matched, stale := decodeHits(recs, paths, catalogue.Channel(only))
			return settle(out, cmd.ErrOrStderr(), rows, matched, stale, term, catalogue.Channel(only), showPaths)
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

// ripgrepArgs is find's whole ripgrep command line after the binary, extracted so a test runs
// real ripgrep with exactly these flags.
//
//   - --line-number, not --count-matches: the line NUMBER is what lets the second pass read the
//     record back and say when it was and what it was.
//   - --byte-offset with --only-matching is the ABSOLUTE file offset of each match itself, which is
//     how the second pass knows WHERE in the record it matched rather than guessing it back from
//     the text. --null ends the path with a NUL, so a Windows drive letter's `C:` is not read as
//     the separator.
//   - --no-config and --encoding none make those offsets the file's raw bytes on every box: a
//     caller's config can add --ignore-case or anything else, and ripgrep's encoding sniffing
//     strips a byte-order mark and reports offsets after it.
func ripgrepArgs(term string, asRegex bool, names []string) []string {
	args := []string{"--no-config", "--encoding", "none", "--line-number", "--no-heading",
		"--with-filename", "--only-matching", "--byte-offset", "--null"}
	if !asRegex {
		args = append(args, "--fixed-strings")
	}
	args = append(args, "--", term)
	return append(args, names...)
}

// rawMatch is one ripgrep match: its absolute byte offset in the file, and the bytes it matched.
type rawMatch struct {
	abs  int64
	text string
}

// ripRecord is one matching transcript RECORD and every match ripgrep reported inside it.
type ripRecord struct {
	path    string
	line    int
	matches []rawMatch
}

// parseRipgrep reads `path\0line:offset:match` and returns each DISTINCT record, in order.
//
// Distinct, because --only-matching emits one output line per match and a single transcript record
// can hold the term many times. Collapsing here is what makes HITS mean "turns that mention this"
// rather than "times these characters appear", which is the number a reader is actually after: one
// transcript line is one turn, and three mentions inside it are one moment. The matches are kept
// on the record, because each one's offset is where the record is attributed.
//
// NO LINE CAP. ripgrep's output is already in memory, and a bufio.Scanner over it silently dropped
// every result after the first match longer than its buffer — which `--regex '.*'` on a
// megabyte transcript line produces.
func parseRipgrep(raw string) []ripRecord {
	type key struct {
		path string
		line int
	}
	var out []ripRecord
	at := map[key]int{}
	for rest := raw; rest != ""; {
		var s string
		s, rest, _ = strings.Cut(rest, "\n")
		path, loc, ok := strings.Cut(s, "\x00")
		if !ok {
			continue
		}
		// line:offset:match — the match can hold ':', so both splits are from the LEFT.
		ln, loc, ok1 := strings.Cut(loc, ":")
		off, text, ok2 := strings.Cut(loc, ":")
		n, err1 := strconv.Atoi(ln)
		abs, err2 := strconv.ParseInt(off, 10, 64)
		if !ok1 || !ok2 || err1 != nil || err2 != nil {
			continue
		}
		k := key{path, n}
		i, seen := at[k]
		if !seen {
			i = len(out)
			at[k] = i
			out = append(out, ripRecord{path: path, line: n})
		}
		out[i].matches = append(out[i].matches, rawMatch{abs: abs, text: text})
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
// snippet, and returns the rows with the number of transcripts that matched at all and the number
// of matching records it could not read back as ripgrep matched them.
//
// UNFILTERED (want == ""), a row describes its newest hit, read from at most samplesPerFile of the
// newest records. FILTERED, every matching record is read and the row describes the newest hit IN
// want; a transcript with none is dropped. Filtering on the newest hit alone answered "whose most
// recent mention was in C" — an earlier match in C vanished, and the empty result read as though
// C had never said it. Nothing is capped under a filter, because a cap there is the same drop.
//
// The row's HITS count is deliberately NOT narrowed: it counts every matching record in that
// transcript either way, so a filtered view cannot make a session look quieter than it is.
//
// A STALE HIT IS NEVER A ROW'S BEST. A record that changed, went away or no longer parses cannot
// say where it matched, and a transcript whose every hit is stale yields no row — it is in the
// stale count instead, which settle states.
func decodeHits(recs []ripRecord, paths map[string]catalogue.TranscriptFile, want catalogue.Channel) (rows []row, matched, stale int) {
	byFile := map[string][]ripRecord{}
	var order []string
	for _, r := range recs {
		if _, ok := byFile[r.path]; !ok {
			order = append(order, r.path)
		}
		byFile[r.path] = append(byFile[r.path], r)
	}
	for _, path := range order {
		f, ok := paths[path]
		if !ok {
			continue // a file ripgrep saw and the store does not name; not ours to report
		}
		matched++
		mine := byFile[path]
		r := row{f: f, hits: len(mine)}
		read := mine
		if want == "" && len(read) > samplesPerFile {
			// The tail of a transcript is its recent end, and this row reports the most recent
			// hit — so when the cap bites it must bite on the OLD end.
			read = read[len(read)-samplesPerFile:]
			r.capped = true
		}
		found := false
		hits, unread := recordsAt(path, read, f.AgentID != "")
		stale += unread
		for _, h := range hits {
			if h.Stale || (want != "" && h.Channel != want) {
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
	return rows, matched, stale
}

// recordsAt reads the given records of one file and decodes each at the offsets ripgrep matched.
// inSubagent is whether the file is a subagent or workflow transcript, which decides who a
// no-origin prompt came from.
//
// One sequential pass, never a seek per line: a transcript runs to hundreds of megabytes, the
// wanted lines are already ascending, and this stops at the last one it needs. ReadBytes keeps
// each line's delimiter, so every line's starting offset is exact, and ripgrep's absolute offsets
// become offsets inside the line.
//
// stale COUNTS WHAT IT COULD NOT READ BACK AS RIPGREP MATCHED IT: every record when the file does
// not open, every wanted line past the end of the file, every record whose bytes at a match
// differ, and every record that no longer parses — a transcript cut after the match but before
// that line's end is the last of these, and discarding DecodeHit's ok would fold it into a
// plain `?`. Those that were read are returned with Stale set; the rest are only counted.
func recordsAt(path string, recs []ripRecord, inSubagent bool) (hits []catalogue.Hit, stale int) {
	want := map[int]ripRecord{}
	last := 0
	for _, r := range recs {
		want[r.line] = r
		last = max(last, r.line)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, len(want) // a transcript that vanished between the search and the read
	}
	defer f.Close()
	rd := bufio.NewReaderSize(f, 256*1024)
	var start int64
	for n := 1; n <= last; n++ {
		b, err := rd.ReadBytes('\n')
		if len(b) == 0 && err != nil {
			break
		}
		if r, ok := want[n]; ok {
			delete(want, n)
			spans := make([]catalogue.Span, len(r.matches))
			for i, m := range r.matches {
				s := int(m.abs - start)
				spans[i] = catalogue.Span{Start: s, End: s + len(m.text), Text: m.text}
			}
			h, ok := catalogue.DecodeHit(string(bytes.TrimSuffix(b, []byte("\n"))), spans, inSubagent)
			if !ok {
				h.Stale = true
			}
			if h.Stale {
				stale++
			}
			hits = append(hits, h)
		}
		start += int64(len(b))
		if err != nil {
			break
		}
	}
	return hits, stale + len(want) // what is left in want lay past the end of the file
}

// settle is everything after decoding: the rows, the worded empty result, or a refusal.
//
// A RECORD NOT READ BACK IS COUNTED OUT LOUD, and a zero that includes one is refused. Under --in,
// "none has a hit in C" with an unread record is the silent zero this verb exists to refuse: the
// record might have been C's, and saying none is a guess reported as a finding. Unfiltered, an
// empty table says the same thing with no words at all.
func settle(out, errw io.Writer, rows []row, matched, stale int, term string, only catalogue.Channel, showPaths bool) error {
	if stale > 0 {
		fmt.Fprintf(errw, "telepathy find: %d matching record(s) could not be read back as ripgrep "+
			"matched them (the file changed or went away since the search, or the record no longer "+
			"parses); they are not counted in any row's IN\n", stale)
	}
	if len(rows) == 0 && stale > 0 {
		return errStaleZero
	}
	if only != "" && len(rows) == 0 {
		// COUNTED IN THE UNIT IT NAMES. ripgrep's records are matching RECORDS, not transcripts,
		// and printing one as the other is how a row count becomes a subject count.
		fmt.Fprintf(out, "%d transcript(s) contain %q, but none has a hit in %q\n", matched, term, only)
		return nil
	}
	render(out, rows, showPaths)
	return nil
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
