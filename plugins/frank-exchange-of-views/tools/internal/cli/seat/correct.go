package seat

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// CorrectionFlagAnnotation marks the two flags Correctable registers. They fill the Correction event,
// not the verb's own body, so a walk from a verb's flags to the fields they fill skips them by this
// mark rather than by name.
const CorrectionFlagAnnotation = "feov_correction_flag"

// correctableKey annotates a command that takes a same-sitting correction, with the event word.
const correctableKey = "feov.correctable"

// invocation is one run of a correctable command: every event it wrote, and the correction it was
// asked to make. It is held per command because one process runs one command; Correctable opens it
// when the command starts and closes it when the command returns.
type invocation struct {
	mu      sync.Mutex
	written []*record.Event
	correct *record.Correct
}

func (inv *invocation) add(ev *record.Event) {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.written = append(inv.written, ev)
}

var (
	invMu       sync.Mutex
	invocations = map[*cobra.Command]*invocation{}
)

func invocationOf(cmd *cobra.Command) *invocation {
	if cmd == nil {
		return nil
	}
	invMu.Lock()
	defer invMu.Unlock()
	return invocations[cmd]
}

// IsCorrectable reports whether a command takes a same-sitting correction.
func IsCorrectable(c *cobra.Command) bool { return c != nil && c.Annotations[correctableKey] != "" }

// eventTypeOf resolves an event word to its type.
func eventTypeOf(word string) (recordpb.EventType, bool) {
	vs := recordpb.EventType(0).Descriptor().Values()
	for i := 0; i < vs.Len(); i++ {
		t := recordpb.EventType(vs.Get(i).Number())
		if word != "" && recordpb.Word(t) == word {
			return t, true
		}
	}
	return 0, false
}

// Correctable gives a recording verb the same-sitting correction: --corrects and --correction-why,
// the paragraph that says how in its help, and the write log that lets its success line show the key
// a correction names. Call it LAST in the verb's constructor — after Records and after every flag,
// because the help paragraph names the verb's own flags.
//
// It panics when the verb records no correctable type: a correction offered on an act the record
// refuses to correct is a flag that can only ever fail.
func Correctable(c *cobra.Command) *cobra.Command {
	typ, ok := eventTypeOf(RecordType(c))
	if !ok || recordpb.Tier(typ) == recordpb.CorrectionTier_CORRECTION_TIER_NONE {
		panic(fmt.Sprintf("seat: %s records %q, which is not a correctable event type — call Records first if the verb's name is not its event, and offer a correction only on a FULL or PROSE type",
			c.CommandPath(), RecordType(c)))
	}
	c.Flags().String(flags.Corrects, "", flags.DescCorrects)
	_ = c.Flags().SetAnnotation(flags.Corrects, CorrectionFlagAnnotation, []string{"true"})
	flags.Text(c, flags.CorrectionWhy, flags.DescCorrectionWhy)
	_ = c.Flags().SetAnnotation(flags.CorrectionWhy, CorrectionFlagAnnotation, []string{"true"})
	c.Long = strings.TrimRight(c.Long, "\n") + "\n\n" + correctionHelp(c, typ) + "\n"
	annotate(c, correctableKey, recordpb.Word(typ))

	inner := c.RunE
	c.RunE = func(cmd *cobra.Command, args []string) error {
		err := beginInvocation(cmd, typ)
		defer func() {
			invMu.Lock()
			delete(invocations, cmd)
			invMu.Unlock()
		}()
		if err != nil {
			return Emit(cmd, nil, err)
		}
		return inner(cmd, args)
	}
	return c
}

func beginInvocation(cmd *cobra.Command, typ recordpb.EventType) error {
	inv := &invocation{}
	invMu.Lock()
	invocations[cmd] = inv
	invMu.Unlock()
	key, why := strings.TrimSpace(Str(cmd, flags.Corrects)), Str(cmd, flags.CorrectionWhy)
	switch {
	case Given(cmd, flags.Corrects):
		if key == "" {
			return feov.Errorf(feov.MissingField, "--corrects names the act you are correcting by its key, printed as [key …] on the line that recorded it; it is empty here")
		}
		if strings.TrimSpace(why) == "" {
			return feov.Errorf(feov.MissingField, "a correction requires --correction-why — what was wrong with the act; a reader sees it beside the struck text")
		}
		inv.correct = &record.Correct{Type: typ, Key: key, Why: why}
	case Given(cmd, flags.CorrectionWhy):
		return feov.Errorf(feov.Validation, "--correction-why says what was wrong with the act --corrects names, and --corrects is not given — pass both to correct an act, or neither to record a new one")
	}
	return nil
}

// CorrectionTarget is the body of the act this invocation corrects, or nil when it corrects nothing.
// A handler that ASSIGNS an identity or COMPUTES a field the act already holds reads it from here
// rather than re-deriving it: a correction re-states the act, and a re-derived value that came out
// different would be refused as a change to a frozen field.
func (c Context) CorrectionTarget() (proto.Message, error) {
	inv := invocationOf(c.cmd)
	if inv == nil || inv.correct == nil {
		return nil, nil
	}
	run, err := c.Run()
	if err != nil {
		return nil, err
	}
	return record.TargetBody(run, c.SeatID, inv.correct.Key, inv.correct.Type)
}

// CorrectionKey is the key of the act this invocation corrects, "" when it corrects nothing.
func (c Context) CorrectionKey() string {
	if inv := invocationOf(c.cmd); inv != nil && inv.correct != nil {
		return inv.correct.Key
	}
	return ""
}

// writtenKey is the key of the last event of the command's own type this invocation wrote.
// tracked is false for a command that keeps no write log (not correctable).
func writtenKey(cmd *cobra.Command) (key string, correcting, tracked bool) {
	inv := invocationOf(cmd)
	if inv == nil {
		return "", false, false
	}
	want := RecordType(cmd)
	inv.mu.Lock()
	defer inv.mu.Unlock()
	for i := len(inv.written) - 1; i >= 0; i-- {
		if recordpb.Word(inv.written[i].GetType()) == want {
			return inv.written[i].GetKey(), inv.correct != nil, true
		}
	}
	return "", inv.correct != nil, true
}

// correctionHelp is the paragraph a correctable verb's help carries: how, who, when, and what may
// change — the last computed from the type's tier and this verb's own flags.
func correctionHelp(c *cobra.Command, typ recordpb.EventType) string {
	var b strings.Builder
	b.WriteString("CORRECTING WHAT YOU RECORDED. If an act this command recorded came out wrong — a lost word, a wrong " +
		"figure — run it again with what you meant, adding --corrects <key> (the key its success line printed as " +
		"[key …]) and --correction-why <what was wrong>. The record keeps the first act, shown struck beside its " +
		"replacement. You may correct only your own act, only in the sitting that recorded it, and only until " +
		"another seat has acted; after that, say it in a new act.")
	own := func(fs []string) []string {
		var out []string
		for _, f := range fs {
			if c.Flags().Lookup(strings.TrimPrefix(f, "--")) != nil {
				out = append(out, f)
			}
		}
		return out
	}
	switch recordpb.Tier(typ) {
	case recordpb.CorrectionTier_CORRECTION_TIER_FULL:
		if l := own([]string{record.LabelFlag(typ)}); len(l) == 1 {
			fmt.Fprintf(&b, " A correction may change anything the act says except %s, which names what it is about.", l[0])
		} else {
			b.WriteString(" A correction may change anything the act says.")
		}
	case recordpb.CorrectionTier_CORRECTION_TIER_PROSE:
		if ps := own(record.ProseFlags(typ)); len(ps) > 0 {
			fmt.Fprintf(&b, " A correction may change only your wording (%s); every other flag must repeat what the act recorded.", strings.Join(ps, ", "))
		} else {
			b.WriteString(" A correction may change only your wording; every other flag must repeat what the act recorded.")
		}
	}
	return wrapHelp(b.String(), 100)
}

// wrapHelp breaks a paragraph at word boundaries so no line runs past width.
func wrapHelp(s string, width int) string {
	var out, line strings.Builder
	for _, w := range strings.Fields(s) {
		if line.Len() > 0 && line.Len()+1+len(w) > width {
			out.WriteString(line.String())
			out.WriteString("\n")
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteString(" ")
		}
		line.WriteString(w)
	}
	out.WriteString(line.String())
	return out.String()
}
