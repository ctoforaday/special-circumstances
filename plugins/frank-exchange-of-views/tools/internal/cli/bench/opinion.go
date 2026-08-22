package bench

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// opinion: a ruling that has to BE an opinion.
//
// principle, tension and review-flag are required by the domain, and the refusal
// says "opinions, not dispositions" because that is the failure being prevented.
// Across runs 4 and 5 the bench ruled `carried` on 64 of 65 items — a router, not
// a judge — and a disposition with no stated principle is indistinguishable from
// a default. Requiring the reasoning is what makes the difference visible.
func newOpinion() *cobra.Command {
	c := seat.Prose(seat.New("opinion",
		// The Use line named two of the seven dispositions and trailed off in an ellipsis. A
		// partial vocabulary is worse than none: a seat that reads `carried|closed|...` has been
		// given enough to act on and not enough to choose well, which is how a bench came to rule
		// `carried` on 64 of 65 items. The set, with what each value is FOR, is one `--help` away
		// and is rendered from the record's own vocabulary.
		`a ruling as an OPINION: --id R3-2 --as <disposition> --principle "..." --tension "correctness vs economy" --review-flag "why a human should look" --reason <rationale>`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			// The word resolves through the ONE vocabulary both closing verbs share. A miss is
			// refused here rather than recorded as the unspecified zero, because an unset
			// disposition reads downstream as a gap the bench never ruled on — which is a
			// quieter, worse outcome than a typo that gets told no.
			as, ok := record.DispositionOf(seat.Str(cmd, flags.As))
			if !ok {
				return nil, feov.Errorf(feov.Validation,
					"bench opinion: %q is not a disposition — the word is what every later reader switches on, and one the record does not know leaves the gap ruled by nobody. Run `feov-record bench opinion --help` for the set and what each value is for", seat.Str(cmd, flags.As))
			}
			body := &recordpb.Opinion{
				GapId:       proto.String(seat.Str(cmd, flags.ID)),
				Disposition: &as,
				Principle:   proto.String(seat.Str(cmd, flags.Principle)),
				Tension:     proto.String(seat.Str(cmd, flags.Tension)),
				ReviewFlag:  proto.String(seat.Str(cmd, flags.ReviewFlag)),
				Rationale:   proto.String(text),
			}
			if _, err := record.Append(s.Identity(), body); err != nil {
				return nil, err
			}
			return opinionResult{ID: seat.Str(cmd, flags.ID), As: seat.Str(cmd, flags.As)}, nil
		}))

	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap being ruled on")
	enumhelp.Flag(c, flags.As, record.MustEnum("opinion", "disposition"), dispositionHelp())
	c.Flags().String(flags.Principle, "", "the principle applied — a ruling is an OPINION, not a disposition")
	c.Flags().String(flags.Tension, "", "the values in tension (e.g. correctness vs economy)")
	c.Flags().String(flags.ReviewFlag, "", "why a human should, or should not, look at this")
	return c
}

type opinionResult struct {
	ID string `json:"id"`
	As string `json:"as"`
}

func (r opinionResult) Human() string { return "opinion recorded: " + r.ID + " " + r.As }

// dispositionHelp states the gap's fate in the seat's own terms, with the deferring words READ OFF
// the vocabulary rather than named here.
//
// The sentence it replaces — "every value ends the gap except `carried`" — was accurate prose and
// a copy of a fact the record now holds as an annotation. Copies of that fact have gone wrong
// before, in the predicate this vocabulary was built to replace and in four agent-facing surfaces
// that named dispositions the record refused. This one cannot: adding a deferring disposition
// rewrites the clause.
// NO BACKTICKS: cobra reads a backquoted word in a flag's usage as the VALUE PLACEHOLDER, so
// "except `carried`" rendered the flag as `--as carried` — the one deferring word displayed where
// the type belongs, in the line a seat skims first.
func dispositionHelp() string {
	deferring := record.Names(record.DeferringDispositions)
	var clause string
	switch len(deferring) {
	case 0:
		// Not reachable today and not silently ignored: a vocabulary in which every word closes
		// would make `bench opinion` unable to defer at all, which is a real change to what the
		// bench can do and should read as one.
		clause = "Every value ends the gap; there is no way to defer one."
	case 1:
		clause = fmt.Sprintf("Every value ends the gap except %s, which defers it to a later round with a stated direction.", deferring[0])
	default:
		clause = fmt.Sprintf("Every value ends the gap except %s, which defer it to a later round with a stated direction.", strings.Join(deferring, " and "))
	}
	return "your ruling AND the gap's fate. " + clause + " One vocabulary with red's closure classes since #342"
}
