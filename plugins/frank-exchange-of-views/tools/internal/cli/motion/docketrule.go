package motion

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH ANSWERS A DOCKETING BY NAMING A SIDE (#1695 §III.B14). Each docketing puts two statements
// before the bench, each stated by its party: red's text in its closing, blue's word in its closing, a
// held proposal's act, G's grades, G's closure. The answer names a side of the docketing's pair, keeps
// or restores the moved grades at grade movement, or remands; the tool applies the side's statement
// in the ruling's transaction and writes the closing word it gives (record.Positions.AnswerWord). No
// flag here chooses a word, a grade or a text.

// docketAnswers is every answer flag, in the order the page lists them, with the side each names.
var docketAnswers = []struct {
	flag, help string
	side       recordpb.SidedWith
}{
	{flags.ForRed, "side with red: at impasse and on blue's close or supersede, the tool applies the text red's closing states — the edit of the span it quotes, or the anchor's retire — and the gap closes repaired; on blue's regrade, the gap stays at its grades. Refused where red's closing states no text, or where the text applied leaves another gap's standing statement no longer applying", recordpb.SidedWith_SIDED_WITH_RED},
	{flags.ForBlue, "side with blue: at impasse the gap closes with the word blue's closing states, or not_a_defect where it states none — blue's text stands; on blue's proposal the tool applies it as filed: a close with its word, a regrade at its grades, a supersede's successor", recordpb.SidedWith_SIDED_WITH_BLUE},
	{flags.ForFiler, "on a red seat's proposal on another red seat's gap: side with the filer — the tool applies the proposal as filed", recordpb.SidedWith_SIDED_WITH_FILER},
	{flags.ForAssignee, "on a red seat's proposal on another red seat's gap: side with the gap's assignee — against a close or a supersede, the tool applies the text its closing states and the gap closes repaired, refused where it states none or where the text applied leaves another gap's standing statement no longer applying; against a regrade, the gap stays at its grades", recordpb.SidedWith_SIDED_WITH_ASSIGNEE},
	{flags.Keep, "at grade movement: the moved grades stand, and each moved axis is locked at the grade it keeps", 0},
	{flags.Restore, "at grade movement: the tool restores each moved axis to the grade the movement started from, and locks it there", 0},
	{flags.Remand, "send an open gap back for one more exchange on --direction; refused at grade movement and past the remand limit", 0},
}

// docketRuleSays is what the docket ruling's page teaches.
const docketRuleSays = "A DOCKETING PUTS TWO STATEMENTS BEFORE YOU, EACH STATED BY ITS PARTY, AND YOU CHOOSE BETWEEN THEM. Name the docketing with --id (your work list names it, with its cause and basis), and answer with exactly one of its pair — `--for-red | --for-blue`; `--for-filer | --for-assignee` on a red seat's proposal on another red seat's gap; `--keep | --restore` at grade movement — or `--remand`. " +
	"Each side states what it holds in the closing it owes on the docketing: red its text — the span of the report it quotes, ending at the gap's marker, and what it becomes, or the retire of the anchor the gap is about — and blue its closing word. A held proposal is its filer's statement; the gap's grades are the other side's against a regrade proposal; the gap's recorded closure the other side's against a supersede of a closed gap; at grade movement the moved grades stand against the recorded ones. " +
	"THE TOOL APPLIES THE STATEMENT OF THE SIDE YOU NAME, in your ruling's transaction, and writes no word, grade or text of yours: red's text as blue's `edit --accept` applies a prescription, and the gap closes `repaired`; blue's word, or `not_a_defect` where blue's closing states none — blue's text stands; a proposal as filed — a close with its word, a regrade at its grades, a supersede's successor, assigned to the gap's assignee; the restored grades on --restore. " +
	"A SIDE WINS ONLY BY WHAT IT STATED: siding with red (the assignee) where its closing states no text — none filed, `--no-text`, or a gap about an inquiry or a section — is refused naming what is missing, and you answer for the other side or remand. ANOTHER GAP'S STANDING TEXT HOLDS UNTIL ITS OWN RULING: an answer whose text leaves it no longer applying is refused naming it — answer the other side, or rule that one first. " +
	"Every answer but one on a grade closes the gap; a ruling on a grade locks each axis it decided at the grade it leaves, and the assignee may not move it back." +
	"\n\nA REMAND SENDS AN OPEN GAP BACK for one more exchange between its assignee and blue, on --direction, whether or not the gap is material, and a PASS is refused until that exchange has been had: the exchange begins only when blue records an act on the gap, and blue's sitting cannot end until it has; the dispatch readies the assignee on the gap once blue has acted, to audit the answer. " +
	"A REMAND PAST THE LIMIT IS REFUSED — " + record.RemandLimitRule + ". So a gap you remanded at impasse, back before you at impasse, takes a side of its pair. A gap's grade movement takes no remand: the question is grades already on the record."

// newDocketRule is the bench's `motion docket rule`.
func newDocketRule(ruleFlags []string) *cobra.Command {
	c := &cobra.Command{
		Use:   "rule",
		Short: "answer a docketing (the bench seat's): side with one party of its pair, keep or restore the moved grades, or remand — the tool applies the side's statement",
		Long: "answer a docketing — this verb is the bench seat's, and it appears only on that surface.\n\n" +
			"EVERY SUBJECT AND ITS GAVEL: " + GavelRoll() + ". The bench's three each convene a sitting of their own: a petition as soon as nothing is running, in the turn it was filed; a position the same way, after the chair sitting that withheld the PASS; a docketing once the parties' closings on it are filed or no longer awaited.\n\n" +
			docketRuleSays + "\n\n" + record.GradeMovementRule + ".",
		RunE: seat.HandlerRunE(func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			run, err := s.Run()
			if err != nil {
				return nil, err
			}
			id := seat.Str(cmd, flags.ID)
			p, found, err := record.PositionsOf(run, id)
			if err != nil {
				return nil, err
			}
			if !found {
				return nil, feov.Errorf(feov.NotFound, "no docketing %s is on the record: name a docketing your work list names. Nothing was recorded", id)
			}
			opinion, err := prose(cmd, "an unreasoned ruling is the decoration the filer cannot contest, and contesting it is the whole reason a ruling is not a command")
			if err != nil {
				return nil, err
			}
			dr := &recordpb.DocketRuling{
				DocketingId: proto.String(id),
				// AN OMITTED FLAG STAYS ABSENT, so the write refuses omission with the schema's own `why`
				// rather than receiving it as the same "" a passed blank is (#1234).
				Principle:  seat.OptStr(cmd, flags.Principle),
				Tension:    seat.OptStr(cmd, flags.Tension),
				ReviewFlag: seat.OptStr(cmd, flags.ReviewFlag),
				Settled:    seat.OptStr(cmd, flags.Settled),
				Direction:  seat.OptStr(cmd, flags.Direction),
			}
			if seat.Given(cmd, flags.ReopensOn) {
				// GIVEN EMPTY IS NOT ABSENT: an empty --reopens-on read as no flag let --final stand alone
				// over a value the seat typed.
				r := seat.Str(cmd, flags.ReopensOn)
				if strings.TrimSpace(r) == "" {
					return nil, feov.Errorf(feov.Validation, "--reopens-on is empty: state the evidence or condition that would reopen the gap, or pass --final instead to say nothing would")
				}
				dr.ReopensOn = proto.String(r)
			}
			if fin, _ := cmd.Flags().GetBool(flags.Final); fin {
				dr.Final = proto.Bool(true)
			}
			answer, err := docketAnswer(cmd, p)
			if err != nil {
				return nil, err
			}
			switch answer {
			case flags.Remand:
				dr.Disposition = recordpb.Disposition_DISPOSITION_REMANDED.Enum()
			case flags.Keep:
				dr.MovedGrades = recordpb.MovedGrades_MOVED_GRADES_KEPT.Enum()
			case flags.Restore:
				dr.MovedGrades = recordpb.MovedGrades_MOVED_GRADES_RESTORED.Enum()
			default:
				side := sideOf(answer)
				dr.SidedWith = side.Enum()
				dr.Disposition = p.AnswerWord(side)
			}
			body := &recordpb.MotionRule{Opinion: proto.String(opinion), Ruling: &recordpb.MotionRule_Docket{Docket: dr}}
			if p.MotionID != "" {
				body.MotionId = proto.String(p.MotionID)
			}
			if err := applyAnswer(s, run, p, answer, body, opinion); err != nil {
				return nil, err
			}
			return ruled{ID: p.MotionID, Docketing: id, Ruling: record.RulingWord(body)}, nil
		}),
	}
	seat.Returns(c, ruled{})
	seat.Prose(c)
	c.Flags().Var(flags.DocketingID(), flags.ID, "the docketing you answer — your work list names it")
	seat.Require(c, flags.ID)
	for _, a := range docketAnswers {
		c.Flags().Bool(a.flag, false, a.help)
	}
	for _, f := range ruleFlags {
		switch f {
		case flags.Final:
			c.Flags().Bool(f, false, ruleFlagHelp[f])
		default:
			flags.Text(c, f, ruleFlagHelp[f])
		}
	}
	// The bench's opinion: validate refuses a bench ruling without any of the four.
	seat.SaysRequired(c, flags.Principle, flags.Tension, flags.ReviewFlag, flags.Settled)
	seat.Records(c, "motion_rule")
	seat.FillsArm(c, "docket")
	return c
}

// sideOf is the side an answer flag names, 0 for a grade answer or a remand.
func sideOf(flag string) recordpb.SidedWith {
	for _, a := range docketAnswers {
		if a.flag == flag {
			return a.side
		}
	}
	return 0
}

// docketAnswer is the one answer flag passed, refused where it is none, more than one, or outside the
// docketing's pair — naming the pair, so the bench is told what it may answer.
func docketAnswer(cmd *cobra.Command, p record.Positions) (string, error) {
	var given []string
	for _, a := range docketAnswers {
		if on, _ := cmd.Flags().GetBool(a.flag); on {
			given = append(given, a.flag)
		}
	}
	pair := p.PairWords()
	if len(given) != 1 {
		return "", feov.Errorf(feov.Validation, "docketing %s is answered with exactly one of its pair, %s, or --remand; this passes %d. Nothing was recorded", p.ID, pair, len(given))
	}
	answer := given[0]
	if answer != flags.Remand && !strings.Contains(" "+pair+" ", " --"+answer+" ") {
		return "", feov.Errorf(feov.Validation, "docketing %s is answered with one of its pair, %s, or --remand; --%s is no answer to it. Nothing was recorded", p.ID, pair, answer)
	}
	return answer, nil
}

// applyAnswer writes the ruling with the statement its answer applies, in one transaction: red's
// (the assignee's) text as the bench's edit or retire before the ruling; a held regrade, or the
// restoring regrade, as the bench's regrade before it; a held supersede's successor after it; the
// ruling alone for every other answer. A side that stated nothing writes the ruling alone, which the
// record refuses naming what is missing.
func applyAnswer(s seat.Context, run record.Run, p record.Positions, answer string, body *recordpb.MotionRule, opinion string) error {
	side := sideOf(answer)
	ruling := record.Act{Body: body}
	switch {
	case answer == flags.Restore && p.Restoring != nil:
		r := proto.Clone(p.Restoring).(*recordpb.Regrade)
		r.Basis = proto.String(opinion)
		_, err := record.AppendActs(s.Identity(), record.Act{Body: r}, ruling)
		return err
	case side != 0 && side == p.TextSide && p.Stated(side):
		return applyText(s, run, p, ruling, opinion)
	case side != 0 && side == p.FilerSide:
		switch p.Act.(type) {
		case *recordpb.Regrade, *recordpb.Mint:
			return applyAccepted(s, run, p.MotionID, ruling)
		}
	}
	_, err := record.AppendActs(s.Identity(), ruling)
	return err
}

// applyText writes the text red's (the assignee's) closing states as the bench's act, then the ruling:
// the anchor's retire where the closing states one, else the edit blue's `edit --accept` makes of its
// pair, planned on the report the write lands on. The freeze holds it as it holds blue's edit: where
// another docketing's or proposal's road no longer plans on the report it leaves, it is refused naming
// that one (blue.BenchRoadsHold).
func applyText(s seat.Context, run record.Run, p record.Positions, ruling record.Act, opinion string) error {
	if p.Road.GetRetire() {
		// RED'S TEXT ALREADY STANDS: a retire of the anchor its road retires matches it, and the freeze
		// admits it (blue.roadPlans), so the answer for red writes its ruling alone.
		if _, retired, err := record.AnchorRetiredAt(run, p.AboutRef); err != nil || retired {
			if err != nil {
				return err
			}
			_, err := record.AppendActs(s.Identity(), ruling)
			return err
		}
		rt := &recordpb.Retire{Reason: proto.String(opinion), Answers: proto.String(p.GapID)}
		_, err := record.AppendActs(s.Identity(), record.Act{Body: rt, Guards: []record.Guard{record.Check(blue.RetireAnchor(rt, p.AboutRef)), blue.BenchRoadsHold}}, ruling)
		return err
	}
	// RED'S TEXT ALREADY STANDS: the report holds the pair's replacement at the gap's place, as an edit
	// of blue's that matched it left it, so the answer for red writes its ruling alone.
	if done, err := record.RedTextApplied(run, p.GapID, p.Road.GetLocation(), p.Road.GetFixNew()); err != nil {
		return err
	} else if done {
		_, err := record.AppendActs(s.Identity(), ruling)
		return err
	}
	ed := &recordpb.BlueEdit{Answers: proto.String(p.GapID), Text: proto.String(opinion)}
	plan := record.Check(func(l *record.Landing) error {
		return blue.AcceptPair(l, ed, p.GapID, p.Road.GetLocation(), p.Road.GetFixNew())
	})
	_, err := record.AppendActs(s.Identity(), record.Act{Body: ed, Guards: []record.Guard{plan, blue.BenchRoadsHold}}, ruling)
	return err
}
