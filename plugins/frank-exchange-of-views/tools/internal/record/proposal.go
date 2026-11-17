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
// on it — a regrade, a close, or a supersede — in that act's own body on the motion (Motion.act). The
// gap's assignee may accept it at its next sitting: its accept is a DocketRuling{accepted: true} and
// the stored act through the act verb's own write, in one transaction. Otherwise the bench answers it:
// accept, reject, or remand. The dispatch's own filings — at impasse, and for grade movement — are
// conditions the record computed, not disputes: an impasse filing stores no act, a grade-movement
// filing the regrade restoring the movement, or no act where nothing moved net.

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
	if _, err := queryRowAt(q, []any{&p.Assignee}, `SELECT COALESCE("assignee", '') FROM "gap" WHERE "gap_id" = ?`, p.GapID); err != nil {
		return p, false, err
	}
	return p, true, nil
}

// ActWord is the word for the act a docket motion proposes: regrade, close, supersede, or "" for none.
func ActWord(m *recordpb.Motion) string {
	switch m.GetAct().(type) {
	case *recordpb.Motion_Regrade:
		return "regrade"
	case *recordpb.Motion_Close:
		return "close"
	case *recordpb.Motion_Supersede:
		return "supersede"
	}
	return ""
}

// ProposedActs is the acts a docket motion proposes, in the order the filing verb mounts them.
var ProposedActs = []string{"regrade", "close", "supersede"}

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
		if c.GetClosureClass() == recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION || c.Successor != nil {
			return feov.Errorf(feov.Validation,
				"a close as repaired_with_regression is the gap's assignee's own act, naming the gap that carries the regression: that successor is a lineage fact no ruling carries, so an accept would drop it. Propose another disposition, or propose the supersede you want. Nothing was recorded")
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
	case *recordpb.Motion_Supersede:
		s := a.Supersede
		if s.GetSupersedes() != g {
			return fmt.Errorf("the proposed successor supersedes %q and the motion is on %s: the tool names the motion's gap as the successor's ancestor", s.GetSupersedes(), g)
		}
		return mintShape(q, s)
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
// assignee (F1), no motion stands unanswered on the gap (F5), a closed gap takes a proposed supersede
// alone (F8) — asked of every writer, and by construction never refusing the dispatch's — and, of
// every writer but the dispatch, one counted exchange on the gap first.
func requireDocketFiling(w writeView, seatID string, m *recordpb.Motion) error {
	q := w.q
	g := m.GetDocket().GetGapId()
	var assignee, standing string
	var open bool
	found, err := queryRowAt(q, []any{&assignee, &open, &standing},
		`SELECT COALESCE("assignee", ''), "open", COALESCE("unruled_docket_motion", '') FROM "gap" WHERE "gap_id" = ?`, g)
	if err != nil || !found {
		return err // an unknown gap is requireGap's refusal
	}
	if assignee == seatID {
		return feov.Errorf(feov.Validation,
			"gap %s is yours — you are its assignee — and a proposal on it asks you to accept your own act. Take the act yourself: regrade it, close it, or mint its successor with --supersedes %s. Nothing was recorded",
			g, g)
	}
	if standing != "" {
		return feov.Errorf(feov.Conflict,
			"motion %s stands unanswered on gap %s, and a gap holds one proposal at a time: argue yours in your closing on %s, or file it once %s is answered. Nothing was recorded",
			standing, g, g, standing)
	}
	if !open {
		if _, ok := m.GetAct().(*recordpb.Motion_Supersede); !ok {
			return feov.Errorf(feov.Conflict,
				"gap %s is closed, and no act but a successor moves a closed gap: propose the successor you want (`motion docket file supersede --id %s`). Nothing was recorded",
				g, g)
		}
	}
	if !w.byDispatch {
		return requireExchangedOn(q, g)
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

// requireDocketAnswer is a docket ruling against the motion it answers (validate): who may write it,
// and which answers fit the motion. The bench answers every docket motion; the gap's assignee writes
// one answer, accepting the act a motion proposes. A motion that proposes an act is answered by
// accepting or rejecting it, or remanded; a closing word is the bench's accept of a proposed close,
// and no other. A motion that proposes none — the dispatch's impasse filing, its no-net grade-movement
// filing, an archived docket — has nothing to accept or reject, and takes a disposition.
func requireDocketAnswer(q recordsql.Querier, seatID string, r *recordpb.MotionRule) error {
	d := r.GetDocket()
	if d == nil {
		return nil
	}
	p, found, err := docketProposalAt(q, r.GetMotionId())
	if err != nil || !found {
		return err // the motion's reference is requireSubjectMatches's refusal
	}
	bench := seatclass.RoleOfSeat(seatID) == "bench"
	act := ActWord(p.Motion)
	if !bench {
		if d.GetAccepted() && p.Assignee == seatID {
			if act == "" {
				return noActRefusal(r.GetMotionId())
			}
			return nil
		}
		return feov.Errorf(feov.RoleViolation,
			"motion %s is answered by the bench, and by one other seat: its gap's assignee, accepting the act it proposes. Gap %s is %s's; you are %s. Nothing was recorded",
			r.GetMotionId(), p.GapID, p.Assignee, seatID)
	}
	if d.Accepted != nil {
		if act == "" {
			return noActRefusal(r.GetMotionId())
		}
		if act == "close" && d.GetAccepted() {
			return fmt.Errorf("the bench's accept of a proposed close is the closing ruling it proposes — --as %s — with --reopens-on or --final; `motion docket rule --accept` writes it", recordpb.Word(p.Motion.GetClose().GetClosureClass()))
		}
		return nil
	}
	if act == "" || d.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
		return nil
	}
	if act == "close" && d.GetDisposition() == p.Motion.GetClose().GetClosureClass() {
		return nil // the bench's accept of the close it proposes
	}
	return feov.Errorf(feov.Validation,
		"motion %s proposes a %s of %s: answer it with --accept or --reject, or remand it. A closing word of your own choosing answers no proposal — accepting a proposed close is how a closing word lands on one. Nothing was recorded",
		r.GetMotionId(), act, p.GapID)
}

// noActRefusal refuses an accept or a reject of a motion that proposes no act, naming the dispositions.
func noActRefusal(motionID string) error {
	return feov.Errorf(feov.Validation,
		"motion %s proposes no act, so there is nothing to accept or reject: rule it with a disposition — one of %s, or remanded. Nothing was recorded",
		motionID, strings.Join(ClosingDispositions(), " | "))
}

// requireBenchHeld is the bench's docket ruling admitted exactly when its current sitting is
// convened to rule the motion (benchHeld): a docket sitting the dispatch convened on the gap after the
// filing, the terminal sitting, or the assembly. Asked of the bench's docket rulings only; the
// assignee's accept is its own act at its own sitting.
func requireBenchHeld(q recordsql.Querier, seatID string, r *recordpb.MotionRule) error {
	if r.GetDocket() == nil || seatclass.RoleOfSeat(seatID) != "bench" {
		return nil
	}
	// AN ANSWERED MOTION IS NO SITTING'S: the first-wins guard (requireUnanswered) refuses its second
	// answer naming the first and what answers it, which is what the bench needs to hear.
	var answered string
	found, err := queryRowAt(q, []any{&answered}, `SELECT "motion_id" FROM "motion_answers" WHERE "motion_id" = ?`, r.GetMotionId())
	if err != nil || found {
		return err
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return err
	}
	if benchHeld(evs, win.IDs(evs), win)[r.GetMotionId()] {
		return nil
	}
	return feov.Errorf(feov.Conflict,
		"motion %s is not before this sitting: the bench rules a docket motion in the docket sitting the dispatch convenes for it once its gap's assignee has had its sitting on it, at the terminal sitting, or at the assembly. Nothing was recorded; the plan readies the bench for it",
		r.GetMotionId())
}

// requireBenchSuccessor is the bench's one Mint: the successor a supersede proposal assigned, on a
// motion the record already holds the bench's accept of. The bench's surface mounts no mint verb, so
// that accept is its only writer, and on replay the ruling precedes the Mint in the stream.
func requireBenchSuccessor(q recordsql.Querier, seatID string, m *recordpb.Mint) error {
	if seatclass.RoleOfSeat(seatID) != "bench" {
		return nil
	}
	var n int64
	found, err := queryRowAt(q, []any{&n},
		`SELECT ms."event_id" FROM "motion_supersede" ms
		   JOIN "motion" mo ON mo."event_id" = ms."event_id"
		   JOIN "motion_rule" mr ON mr."motion_id" = mo."motion_id"
		   JOIN "motion_rule_docket" rd ON rd."event_id" = mr."event_id"
		   JOIN "events" re ON re."id" = mr."event_id"
		  WHERE ms."gap_id" = ? AND rd."accepted" = 1 AND re."seat_id" = ?`, m.GetGapId(), seatID)
	if err != nil || found {
		return err
	}
	return feov.Errorf(feov.RoleViolation,
		"the bench mints no gap of its own: its one Mint is the successor a proposed supersede assigned, written by its accept of that proposal (`motion docket rule --accept`). %s is no such successor. Nothing was recorded",
		m.GetGapId())
}

// actingSeat is which seat acted in a Mint: the seat whose write's checks ran — the event's seat,
// except where the bench wrote it, which it does only as an accepted proposal's successor, with every
// assignee-keyed check run against the gap's assignee (L5). Credited with what the mint did: the debt
// it discharges, the reading behind it, authorship of its text and of a prescription.
func actingSeat(seatID, assignee string) string {
	if seatclass.RoleOfSeat(seatID) == "bench" && assignee != "" {
		return assignee
	}
	return seatID
}

// mintActors is each gap's acting seat on the stream (actingSeat): the Mint event's seat, or, where
// the bench wrote it, the assignee its chain carries — the ancestor's assignee, or the minter of a gap
// naming none. It states on the stream the rule gap_state.assignee stores at the mint
// (recordsql.assigneeSQL), for the readers that fold the stream with no gap rows in reach;
// TestAssigneeAndActingSeatReadersOfASuccessor holds the two to one answer.
func mintActors(evs []*Event) map[string]string {
	_, acting := mintChain(evs)
	return acting
}

// mintChain is each gap's assignee and its mint's acting seat on the stream: a mint naming its one
// ancestor takes the ancestor's assignee, whoever writes it; a mint naming none, its writer's.
func mintChain(evs []*Event) (assignee, acting map[string]string) {
	assignee, acting = map[string]string{}, map[string]string{}
	for _, e := range evs {
		m, ok := recordpb.BodyAs[*recordpb.Mint](e)
		if !ok || m.GetGapId() == "" {
			continue
		}
		if _, seen := assignee[m.GetGapId()]; seen {
			continue
		}
		a := e.GetSeatId()
		if x, ok := assignee[m.GetSupersedes()]; ok {
			a = x
		}
		assignee[m.GetGapId()] = a
		acting[m.GetGapId()] = actingSeat(e.GetSeatId(), a)
	}
	return assignee, acting
}

// actingSeatOf is gap's acting seat on the stream (mintActors), for a reader that names one gap.
func actingSeatOf(evs []*Event, gap string) string { return mintActors(evs)[gap] }

// mintActingSeat is actingSeat for a Mint about to be written: the bench's successor is checked
// against the assignee of the gap it supersedes, whose chain it continues.
func mintActingSeat(q recordsql.Querier, seatID string, m *recordpb.Mint) (string, error) {
	if seatclass.RoleOfSeat(seatID) != "bench" || m.GetSupersedes() == "" {
		return seatID, nil
	}
	var assignee string
	if _, err := queryRowAt(q, []any{&assignee}, `SELECT COALESCE("assignee", '') FROM "gap" WHERE "gap_id" = ?`, m.GetSupersedes()); err != nil {
		return "", err
	}
	return actingSeat(seatID, assignee), nil
}

// benchClosureWord is the word a bench closure states: the ruling's disposition, or `superseded` for
// the bench's accept of a proposed supersede, which closes the gap as a supersession does.
func benchClosureWord(d *recordpb.DocketRuling) string {
	if d.GetAccepted() {
		return ClosedSuperseded
	}
	return recordpb.Word(d.GetDisposition())
}

// StoredAct is the act a docket motion proposes, cloned so the writer that takes it may fill its
// derived fields: a supersede's stamped class material is cleared, since the write stamps it again.
func StoredAct(m *recordpb.Motion) proto.Message {
	switch a := m.GetAct().(type) {
	case *recordpb.Motion_Regrade:
		return proto.Clone(a.Regrade)
	case *recordpb.Motion_Close:
		return proto.Clone(a.Close)
	case *recordpb.Motion_Supersede:
		s := proto.Clone(a.Supersede).(*recordpb.Mint)
		s.ClassMaterial = nil
		return s
	}
	return nil
}
