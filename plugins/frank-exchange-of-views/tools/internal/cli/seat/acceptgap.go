package seat

import (
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE AUTHOR TAKES THE RAISER'S STATEMENT ON A GAP ABOUT ITS GAP (--accept G2, §III.B2): the raiser's
// closing on G2 states the wording and grades the author's gap G should carry, and the author's accept
// writes that statement whole — the amend of the wording it states, the regrade of the grades it
// states, or both in one transaction — each answering G2 and accepted. An offer is taken as stated,
// so `amend --accept G2` and `regrade --accept G2` write the same acts; under a docketing on G2 it is
// the author's concession. The record holds the bytes to the statement (requireAuthorAccept).

// GapStatementActs is the acts the author's --accept g2 writes on the gap g2 is about, with reason as
// each act's reason; gapID, where the seat named one with --id, is that gap.
func GapStatementActs(s Context, g2, gapID, reason string) ([]record.Act, string, error) {
	run, err := s.Run()
	if err != nil {
		return nil, "", err
	}
	st, found, err := record.StatementOf(run, g2)
	if err != nil {
		return nil, "", err
	}
	if !found {
		return nil, "", feov.Errorf(feov.NotFound, "--accept: no gap %s on the board", g2)
	}
	under := st.AboutRef
	if anchor.Kind(under) != "gap" {
		return nil, "", feov.Errorf(feov.Validation, "--accept %s names a gap that is not about a gap: a G- id here is another lens's gap about your gap, whose raiser's statement you take; the author's proposal on your gap is taken by its M- id. Nothing was recorded", g2)
	}
	if gapID != "" && gapID != under {
		return nil, "", feov.Errorf(feov.Validation, "--accept %s takes a statement about gap %s, and --id names %s: drop --id — the gap the statement is about names it. Nothing was recorded", g2, under, gapID)
	}
	if st.Wording == nil && st.Grades == nil {
		return nil, "", feov.Errorf(feov.Conflict, "--accept %s: its raiser's closing states no wording or grades for your gap, so there is nothing to take. Nothing was recorded", g2)
	}
	var acts []record.Act
	if w := st.Wording; w != nil {
		acts = append(acts, record.Act{Body: &recordpb.Amend{GapId: proto.String(under), Problem: w.Problem, RequiredFix: w.RequiredFix,
			AcceptanceCheck: w.AcceptanceCheck, Basis: proto.String(reason), Answers: proto.String(g2), Accepted: proto.Bool(true)}})
	}
	if gr := st.Grades; gr != nil {
		acts = append(acts, record.Act{Body: &recordpb.Regrade{GapId: proto.String(under), Severity: gr.Severity, Likelihood: gr.Likelihood,
			Impact: gr.Impact, ComplexityCost: gr.ComplexityCost, Basis: proto.String(reason), Answers: proto.String(g2), Accepted: proto.Bool(true)}})
	}
	return acts, under, nil
}

// WriteGapStatement writes the author's accept of g2's statement (GapStatementActs) with the Fresh
// guard on the gap it changes, as an amend or a regrade of it does.
func WriteGapStatement(s Context, under string, acts []record.Act) error {
	return AppendFresh(s, under, func(guards ...record.Guard) error {
		acts[0].Guards = append(acts[0].Guards, guards...)
		_, err := record.AppendActs(s.Identity(), acts...)
		return err
	})
}
