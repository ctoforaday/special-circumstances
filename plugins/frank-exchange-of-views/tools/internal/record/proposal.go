package record

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A DOCKET MOTION IS A PROPOSED ACT (#1695 Part B). A seat that disputes a gap files the act it wants
// on it — a regrade or a close — in that act's own body on the motion (Motion.act). The
// gap's assignee may accept it at its next sitting: its accept is a DocketRuling siding with the
// filer, and the stored act through the act verb's own write, in one transaction. Otherwise the tool
// dockets the gap for it (Docketing, docketing.go) and the bench answers it there.

// DocketProposal is a docket motion as its answer reads it: the motion, its event, its filer, and the
// gap's assignee.
type DocketProposal struct {
	Motion   *recordpb.Motion
	EventID  int64
	Filer    string
	GapID    string
	Assignee string
}

// DocketProposalOf is the docket motion id on run's record (ok false where none is filed under that id).
func DocketProposalOf(run Run, motionID string) (p DocketProposal, ok bool, err error) {
	err = readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		p, ok, err = docketProposalAt(q, motionID)
		return err
	})
	return p, ok, err
}

// docketProposalAt is DocketProposalOf asked of q.
func docketProposalAt(q recordsql.Querier, motionID string) (DocketProposal, bool, error) {
	var p DocketProposal
	found, err := queryRowAt(q, []any{&p.EventID, &p.Filer, &p.GapID},
		`SELECT m."event_id", e."seat_id", md."gap_id" FROM "motion" m
		   JOIN "events" e ON e."id" = m."event_id"
		   JOIN "motion_docket" md ON md."event_id" = m."event_id"
		  WHERE m."motion_id" = ?`, motionID)
	if err != nil || !found {
		return p, false, err
	}
	ev, err := recordsql.EventAt(q, p.EventID)
	if err != nil {
		return p, false, err
	}
	m, ok := recordpb.BodyAs[*recordpb.Motion](ev)
	if !ok {
		return p, false, fmt.Errorf("event %d is no motion", p.EventID)
	}
	p.Motion = m
	if _, err := queryRowAt(q, []any{&p.Assignee}, `SELECT COALESCE("minted_by", '') FROM "gap" WHERE "gap_id" = ?`, p.GapID); err != nil {
		return p, false, err
	}
	return p, true, nil
}

// ActWord is the word for the act a docket motion proposes: regrade, close, or "" for none.
func ActWord(m *recordpb.Motion) string {
	switch m.GetAct().(type) {
	case *recordpb.Motion_Regrade:
		return "regrade"
	case *recordpb.Motion_Close:
		return "close"
	}
	return ""
}

// ProposedActs is the acts a docket motion proposes, in the order the filing verb mounts them.
var ProposedActs = []string{"regrade", "close"}

// requireProposedAct is the shape of the act a docket motion proposes (validate): the act's own shape,
// on the motion's gap. Who may take it is asked at the accept, against the seat that takes it — the
// filer cannot pass a check keyed on the assignee.
func requireProposedAct(q recordsql.Querier, m *recordpb.Motion) error {
	g := m.GetDocket().GetGapId()
	switch a := m.GetAct().(type) {
	case nil:
		return nil
	case *recordpb.Motion_Regrade:
		r := a.Regrade
		if r.GetGapId() != g {
			return fmt.Errorf("the proposed regrade names gap %s and the motion is on %s: a proposal moves the gap it is filed on", r.GetGapId(), g)
		}
		if err := requireNoRegradeResult(r); err != nil {
			return err
		}
		return requireRegradeMoves(q, r)
	case *recordpb.Motion_Close:
		c := a.Close
		if c.GetGapId() != g {
			return fmt.Errorf("the proposed close names gap %s and the motion is on %s: a proposal closes the gap it is filed on", c.GetGapId(), g)
		}
		if !recordpb.Closes(c.GetClosureClass()) {
			return fmt.Errorf("--as %s closes nothing; a proposed close states one of %s", recordpb.Word(c.GetClosureClass()), strings.Join(ClosingDispositions(), " | "))
		}
		if c.CarriedFrom != nil {
			return fmt.Errorf("a proposed close is a close, not a carry: it states what was verified, by whom, with what and against what")
		}
		if c.GetAnchorSeat() == "" || c.GetAnchorTool() == "" || c.GetAnchorTarget() == "" {
			return fmt.Errorf("requires the verification triple (--verified-by --verified-with --verified-against) — an unverified closure is unauditable (E0.5a), and the assignee or the bench accepts the close as it is stated")
		}
		return requireClosureAnchors(q, g, c.GetAnchors(), "--anchor")
	}
	return fmt.Errorf("motion carries an act this binary has no shape for — the schema and the write path have diverged")
}

// requireRegradeMoves refuses a regrade that moves no axis of the gap as it stands: a proposal that
// changes nothing asks nobody to do anything.
func requireRegradeMoves(q recordsql.Querier, r *recordpb.Regrade) error {
	var sev, lik, imp, cx string
	found, err := queryRowAt(q, []any{&sev, &lik, &imp, &cx},
		`SELECT COALESCE("current_severity", ''), COALESCE("current_likelihood", ''), COALESCE("current_impact", ''), COALESCE("current_complexity_cost", '')
		   FROM "gap" WHERE "gap_id" = ?`, r.GetGapId())
	if err != nil || !found {
		return err // an unknown gap is requireGap's refusal
	}
	for _, ax := range []struct {
		set bool
		to  recordpb.Grade
		now string
	}{{r.Severity != nil, r.GetSeverity(), sev}, {r.Likelihood != nil, r.GetLikelihood(), lik}, {r.Impact != nil, r.GetImpact(), imp}, {r.ComplexityCost != nil, r.GetComplexityCost(), cx}} {
		if ax.set && recordpb.Word(ax.to) != ax.now {
			return nil
		}
	}
	return feov.Errorf(feov.Validation,
		"the proposed regrade moves no axis of %s as it stands (severity %s, likelihood %s, impact %s, complexity %s): propose the grade you hold it should carry. Nothing was recorded",
		r.GetGapId(), orUngraded(sev), orUngraded(lik), orUngraded(imp), orUngraded(cx))
}

func orUngraded(w string) string {
	if w == "" {
		return "ungraded"
	}
	return w
}

// requireDocketFiling is the act judgements of a docket filing (judgeAct): the filer is not the gap's
// assignee (F1), no motion stands unanswered on the gap (F5), a closed gap takes none (F8) — asked of
// every writer, and by construction never refusing the dispatch's — and, of
// every writer but the dispatch, one counted exchange on the gap first.
func requireDocketFiling(w writeView, seatID string, m *recordpb.Motion) error {
	q := w.q
	g := m.GetDocket().GetGapId()
	var assignee, standing string
	var open bool
	found, err := queryRowAt(q, []any{&assignee, &open, &standing},
		`SELECT COALESCE("minted_by", ''), "open", COALESCE("unruled_docket_motion", '') FROM "gap" WHERE "gap_id" = ?`, g)
	if err != nil || !found {
		return err // an unknown gap is requireGap's refusal
	}
	if assignee == seatID {
		return feov.Errorf(feov.Validation,
			"gap %s is yours — you are its assignee — and a proposal on it asks you to accept your own act. Take the act yourself: regrade it or close it. Nothing was recorded",
			g)
	}
	if err := requireNotEstoppedOn(q, g); err != nil {
		return err
	}
	if err := requireNoDocketingOnFiling(q, g); err != nil {
		return err
	}
	if standing != "" {
		return feov.Errorf(feov.Conflict,
			"motion %s stands unanswered on gap %s, and a gap holds one proposal at a time: argue yours in your closing on %s, or file it once %s is answered. Nothing was recorded",
			standing, g, g, standing)
	}
	if !open {
		return feov.Errorf(feov.Conflict,
			"gap %s is closed, and no proposed act moves a closed gap: a defect still in the report is a new gap, minted by the lens whose class it is. Nothing was recorded", g)
	}
	if err := requireExchangedOn(q, g); err != nil {
		return err
	}
	return nil
}

// requireExchangedOn refuses a seat's proposal on a gap that has had no counted exchange: the dispute
// is argued between the gap's parties first. A gap whose sittings the record cannot close is NOT
// MEASURED, never zero (GapExchanges.Counted).
func requireExchangedOn(q recordsql.Querier, g string) error {
	evs, win, err := eventsAt(q)
	if err != nil {
		return err
	}
	remands, err := remandRulingsOf(q)
	if err != nil {
		return err
	}
	x := exchangeFold(evs, win.IDs(evs), win, WhileRunning, remands)[g]
	if x == nil {
		x = &GapExchanges{GapID: g}
	}
	if x.Exchanges >= 1 {
		return nil
	}
	return feov.Errorf(feov.Conflict,
		"gap %s stands at %s, and a proposal on a gap is admitted once it has had one exchange: its assignee and blue argue it first, and your dispute reaches them through the debate. Nothing was recorded",
		g, x.Counted())
}

// requireBenchHeld is the bench's docket ruling admitted exactly when its current sitting is
// convened to rule the docketing it names (benchHeld): a docket sitting the dispatch convened on the
// docketing's gap, the terminal sitting, or the assembly. Asked of the bench's docket rulings only; the
// assignee's accept is its own act at its own sitting.
func requireBenchHeld(q recordsql.Querier, seatID string, r *recordpb.MotionRule) error {
	if r.GetDocket() == nil || seatclass.RoleOfSeat(seatID) != "bench" {
		return nil
	}
	d := r.GetDocket().GetDocketingId()
	// AN ANSWERED OR UNKNOWN DOCKETING IS NO SITTING'S: the first-wins guard (requireUnanswered) refuses
	// a second answer naming the first, and the shape check an id the record does not hold, which is
	// what the bench needs to hear.
	if d == "" {
		return nil
	}
	answered, err := recordHasAt(q, `SELECT 1 FROM "motion_rule_docket" WHERE "docketing_id" = ?`, d)
	if err != nil || answered {
		return err
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return err
	}
	if benchHeld(evs, win.IDs(evs), win)[d] {
		return nil
	}
	return feov.Errorf(feov.Conflict,
		"docketing %s is not before this sitting: the bench rules a docketing in the docket sitting the dispatch convenes for it once the closings owed on it are in, at the terminal sitting, or at the assembly. Nothing was recorded; the plan readies the bench for it",
		d)
}

// mintersOf is each gap's minter on the stream: the seat that wrote its Mint. It states on the stream
// what gap_state.minted_by stores, for the readers that fold the stream with no gap rows in reach.
func mintersOf(evs []*Event) map[string]string {
	out := map[string]string{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetGapId() != "" {
			if _, seen := out[m.GetGapId()]; !seen {
				out[m.GetGapId()] = e.GetSeatId()
			}
		}
	}
	return out
}

// StoredAct is the act a docket motion proposes, cloned so the writer that takes it may fill its
// derived fields.
func StoredAct(m *recordpb.Motion) proto.Message {
	switch a := m.GetAct().(type) {
	case *recordpb.Motion_Regrade:
		return proto.Clone(a.Regrade)
	case *recordpb.Motion_Close:
		return proto.Clone(a.Close)
	}
	return nil
}

// requireNotEstoppedOn refuses a proposal on a gap a standing ruling tombstone protects. A bench ruling but a remand is final
// for the ruled gap's text while its tombstone stands, whoever files; the dispatch files no proposal,
// so its docketings are not barred.
func requireNotEstoppedOn(q recordsql.Querier, g string) error {
	t, found, err := rulingProtecting(q, g)
	if err != nil || !found {
		return err
	}
	var aboutKind, aboutRef string
	if _, err := queryRowAt(q, []any{&aboutKind, &aboutRef},
		`SELECT COALESCE("about_kind", ''), COALESCE("about_ref", '') FROM "mint" WHERE "gap_id" = ?`, t.Gap); err != nil {
		return err
	}
	return feov.Errorf(feov.Conflict,
		"the bench ruled gap %s %s on docketing %s (event %d); its tombstone stands, and a ruling is final for the ruled gap's text while it does: a proposal on %s relitigates it. "+
			"%s ends it; blue's evidence belongs in the exchange before the bench rules. Nothing was recorded",
		t.Gap, t.Word, t.Docketing, t.Event, g, capitalized(tombstoneEnds(aboutKind, aboutRef)))
}
