package bench

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// declare: the bench states something that binds how the record is READ, without moving any
// gap's fate.
//
// # Found by a bench seat, ruling a real petition
//
// From research/2026-08-10_pre-motion-real-record, in the bench's own friction (#361):
//
//	No declaratory disposition — `bench opinion` requires --id plus a fate-changing --as, so
//	the one channel both seats read was unavailable for the one holding both seats most need.
//
// Quoted as written. That verb is `motion docket rule` now (#681) and the argument is unchanged by
// the rename: it had a finding both parties needed on the record and no way to state it. Disposing
// of a docketed gap wants an id and a fate. Every bench act changed something's
// fate, so a holding that changes none had no verb — and the bench put it in a petition ruling's
// opinion text, which was the channel least likely to be read (#360).
//
// # What is a declaration, and why it is a real category
//
// Three shapes, all met in one run:
//
//   - A CONSTRUCTION of a term. "verified means an act of looking, and may never be set from
//     confidence in the critique." Nobody's gap moves; how every gap is read does.
//   - A CORRECTION of the record's meaning without touching its contents.
//   - A HOLDING offered for promotion to precedent. This one is instructive: `law/proposed/` is
//     fed by a harvest that reads DISPOSITIONS, so a holding nobody disposed of is a holding the
//     harvest cannot see. It was offered in prose and went nowhere.
//
// # Why --ids and not --id
//
// A declaration that DISPOSED of a gap would be a ruling, and `motion docket rule` already does
// that, with --id. A holding moves no fate, and is often about the record as a whole — but when it
// is about gaps, the gaps it is about are a fact the work list needs: it sets each beside where the
// board says that gap stands now, so a status the holding's prose states cannot outlive the gap's
// move. --ids carries them as a field the write checks, each a gap the board holds; prose that names
// a gap is still prose, and nothing is read out of it.
//
// It reaches every seat on its work list (record.DeclarationJSON), the read a seat holds before its
// first act, with the board's status beside each gap it names; and it renders into `inquest debate`
// beside the bench's dispositions, where the chair and the bench read what the bench ruled. A ruling that reaches no reader is decoration
// — and the bench is this system's ethical and safety boundary, which makes an undelivered one
// worse than an unread finding: an unread finding costs a reader, an undelivered ruling costs
// compliance, and nothing reports that it was not delivered.
func newDeclare() *cobra.Command {
	// The handler above refuses an empty holding; the marker says so where the seat reads.
	var ids flags.CSV
	c := seat.SaysRequired(seat.Prose(seat.New("declare", seat.Msg{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		text, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		if text == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --reason: the declaration IS the reasoning. A holding with no stated content binds nothing and teaches nobody, and this verb exists because the alternative was prose on a channel nothing reads")
		}
		gaps := ids.Value()
		for _, id := range gaps {
			if err := flags.GapID().Set(id); err != nil {
				return nil, feov.Errorf(feov.Validation, "--%s: %v", flags.IDs, err)
			}
		}
		if _, err := record.Append(s.Identity(), &recordpb.Declare{Holding: proto.String(text), GapIds: gaps}); err != nil {
			return nil, err
		}
		return seat.Msg{Message: "declaration recorded — every seat's work list carries it, with the board's status beside each gap --ids names, and it renders under ### LEAD in `inquest debate`"}, nil
	})), flags.Reason)
	c.Flags().Var(&ids, flags.IDs, "the `gap-ids` of the gaps this holding is about, comma-separated — each a gap the board holds (`show board` lists them). Every seat's work list sets each beside where the board says it stands now; a gap named only in --reason is not one of them")
	return c
}
