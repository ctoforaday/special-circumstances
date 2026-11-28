package seat

import (
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AcceptedAct is the act docket motion m proposes, as the raiser's verb takes it (--accept M): the
// motion's stored act, cloned for the verb to fill its own fields. gapID, where the seat named one
// with --id, is the motion's gap. The record asks the rest in the write's transaction — that m still
// stands on the gap, that the seat is its raiser, and that the act is the one filed
// (record.requireAcceptedAsFiled).
func AcceptedAct(s Context, m, gapID string) (proto.Message, error) {
	run, err := s.Run()
	if err != nil {
		return nil, err
	}
	p, found, err := record.DocketProposalOf(run, m)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, feov.Errorf(feov.NotFound, "--accept %s names no proposal on the record: a proposal's id is the M- id its filing returned, and your work list names each one standing on your gaps. Nothing was recorded", m)
	}
	if gapID != "" && gapID != p.GapID {
		return nil, feov.Errorf(feov.Validation, "motion %s proposes an act on gap %s, and --id names %s: drop --id — the proposal names its gap. Nothing was recorded", m, p.GapID, gapID)
	}
	// The stored act names its gap through the motion's filing; the act the raiser writes names it.
	switch a := record.StoredAct(p.Motion).(type) {
	case *recordpb.Regrade:
		a.GapId = proto.String(p.GapID)
		return a, nil
	case *recordpb.Close:
		a.GapId = proto.String(p.GapID)
		return a, nil
	}
	return nil, feov.Errorf(feov.Validation, "motion %s proposes no act: there is nothing to take. Nothing was recorded", m)
}
