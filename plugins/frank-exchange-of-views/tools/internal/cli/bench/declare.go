package bench

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
// It had a finding both parties needed on the record and no way to state it. `opinion` is built
// to DISPOSE of a docketed gap: it wants an id and a fate. Every bench act changed something's
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
// # Why no --id, and why that is the point rather than an omission
//
// A declaration that named a gap would be an opinion, and `opinion` already exists. The absence
// of the flag is the verb's whole content: this act is about the record as a whole, and forcing
// it to name one gap would either falsify what it says or send the bench back to prose.
//
// It renders into `show debate` beside the opinions, because the surface a seat reads to catch up
// is the surface a binding statement has to be on. A ruling that reaches no reader is decoration
// — and the bench is this system's ethical and safety boundary, which makes an undelivered one
// worse than an unread finding: an unread finding costs a reader, an undelivered ruling costs
// compliance, and nothing reports that it was not delivered.
func newDeclare() *cobra.Command {
	return seat.Prose(seat.New("declare", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		text, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		if text == "" {
			return nil, feov.Errorf(feov.MissingField, "bench declare requires --reason: the declaration IS the reasoning. A holding with no stated content binds nothing and teaches nobody, and this verb exists because the alternative was prose on a channel nothing reads")
		}
		if _, err := record.Append(s.Identity(), "declare", record.NewPayload().Set("holding", text)); err != nil {
			return nil, err
		}
		return seat.Msg{Message: "declaration recorded — it renders under ### LEAD in `show debate`, where both seats read it"}, nil
	}))
}
