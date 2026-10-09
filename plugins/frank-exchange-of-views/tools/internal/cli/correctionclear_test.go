package cli

import (
	"slices"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// withoutFlag drops a flag and its value from an argument list.
func withoutFlag(args []string, flag string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		if args[i] == flag {
			i++
			continue
		}
		out = append(out, args[i])
	}
	return out
}

// withFlag gives a flag a new value, in place.
func withFlag(args []string, flag, value string) []string {
	out := slices.Clone(args)
	for i := 0; i+1 < len(out); i++ {
		if out[i] == flag {
			out[i+1] = value
		}
	}
	return out
}

// lastBodyOfWord is the body of the last event of a type named by its word.
func lastBodyOfWord(t *testing.T, runDir, word string) proto.Message {
	t.Helper()
	typ, ok := eventTypeWord(word)
	if !ok {
		t.Fatalf("%q is no event type", word)
	}
	body, ok := recordpb.Body(lastOfType(t, runDir, typ))
	if !ok {
		t.Fatalf("the last %s carries no body", word)
	}
	return body
}

// A CORRECTION CLEARS EXACTLY WHAT ITS HELP OFFERS TO CLEAR, ON EVERY CORRECTABLE COMMAND.
//
// A correction repeats every flag, so wording the act holds and the correction leaves out is
// refused rather than dropped — and the refusal, like the help paragraph, offers the way to drop
// it on purpose: pass the flag empty. That offer is only worth making where the verb admits an
// empty value. Most prose flags are REQUIRED and refuse one, so the paragraph names the flags a
// correction may drop (seat.ClearableProse), and this drives each of them through the command
// line: left out, the correction is refused naming the flag; passed empty, it is recorded and the
// replacement no longer holds the wording.
//
// It also holds the refusal's own premise: the wording an act holds is typed through a flag its
// command registers. A held prose field whose flag the command lacks could be neither repeated nor
// refused, and the check would pass over it in silence.
func TestACorrectionClearsExactlyWhatItsHelpOffers(t *testing.T) {
	rows := corrRows()
	byPath := commandsByPath()
	var paths []string
	for p, c := range byPath {
		if seat.IsCorrectable(c) {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	offered := 0
	for _, p := range paths {
		row, ok := rows[p]
		if !ok {
			continue // TestCorrectionAcceptedPerCorrectablePath fails a path with no row
		}
		c := byPath[p]
		clearable := seat.ClearableProse(c)
		help := strings.Join(strings.Fields(seat.HelpText(c)), " ")
		if !strings.Contains(help, "Repeat your wording too: a correction that leaves out a flag the act holds is refused.") {
			t.Errorf("%s: the help does not say that a correction repeats the seat's wording", p)
		}
		if said := strings.Contains(help, " may be dropped, "); said != (len(clearable) > 0) {
			t.Errorf("%s: the help offers a clear = %v, and the command takes %v empty", p, said, clearable)
		}
		if len(clearable) > 0 && !strings.Contains(help, "Only "+strings.Join(clearable, ", ")+" may be dropped, ") {
			t.Errorf("%s: the help does not name %v as the wording a correction may drop", p, clearable)
		}
		start := func(t *testing.T) (runDir string, act []string, k string) {
			runDir = corrFixture(t)
			var v corrVars
			if row.setup != nil {
				v = row.setup(t, runDir)
			}
			act = row.act(v, "the text as first recorded")
			return runDir, act, correctionKeyOf(t, runDir, row.seat, act)
		}
		correct := func(t *testing.T, runDir, k string, args []string) (string, error) {
			return runAt(t, append(slices.Clone(args), "--run", runDir, "--seat-id", row.seat, "--corrects", k, "--correction-why", "a word was lost")...)
		}
		t.Run(p, func(t *testing.T) {
			runDir, _, _ := start(t)
			held := record.HeldProseFlags(lastBodyOfWord(t, runDir, seat.RecordType(c)))
			for _, f := range held {
				if c.Flags().Lookup(strings.TrimPrefix(f, "--")) == nil {
					t.Errorf("the act holds wording typed as %s, and %s registers no such flag — a correction can neither repeat it nor be refused for leaving it out", f, p)
				}
			}
			for _, f := range clearable {
				if !slices.Contains(held, f) {
					t.Fatalf("the help says %s may be passed empty and this row's act does not hold it (held: %v) — give the row the flag, so the offer is driven", f, held)
				}
			}
		})
		for _, f := range clearable {
			offered++
			t.Run(p+" omits "+f, func(t *testing.T) {
				runDir, act, k := start(t)
				_, err := correct(t, runDir, k, withoutFlag(act, f))
				if err == nil || !strings.Contains(err.Error(), "omits "+f+", which the act you are correcting holds") ||
					!strings.Contains(err.Error(), `pass `+f+` "" to clear it`) {
					t.Fatalf("a correction leaving out %s was not refused by name with the way to clear it: %v", f, err)
				}
			})
			t.Run(p+" clears "+f, func(t *testing.T) {
				runDir, act, k := start(t)
				if _, err := correct(t, runDir, k, withFlag(act, f, "")); err != nil {
					t.Fatalf("the help offers %s \"\" and the correction passing it was refused: %v", f, err)
				}
				if held := record.HeldProseFlags(lastBodyOfWord(t, runDir, seat.RecordType(c))); slices.Contains(held, f) {
					t.Fatalf("the correction passed %s empty and the replacement still holds it", f)
				}
			})
		}
	}
	if offered < 6 {
		t.Fatalf("only %d clearable flag(s) driven — the walk is not seeing the surface", offered)
	}
}
