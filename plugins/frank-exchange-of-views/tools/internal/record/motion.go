package record

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// ONE ADJUDICATION MECHANISM, WITH AN ID (#344).
//
// The propose→rule exchange was implemented three times, with three vocabularies and no shared
// identity:
//
//	directions   blue line of inquiry      -> merge line of inquiry-rule       key `ruling`
//	governance   <seat> petition  -> bench petition-rule      key `ruling`
//	grades       blue dispute     -> merge dispute-respond    key `response`
//
// Two spellings of one concept, three renderers, and nothing tying an ask to its answer. That is
// the direct cause of a defect class fixed one instance at a time: #315 found the petition FILING
// unrendered while the line of inquiry RULING was found unrendered SEPARATELY in the same sweep, because
// nothing said they were the same mechanism. #312 is the same root — `petition-rule` joins on
// `(petitioner, class)` with no id, which is why the report renders filings and rulings side by
// side rather than joined: pairing two filings by one seat in one round would be a guess.
//
// A motion has an ID. The ask and its answer join on it, once, and one renderer serves all three.

// MotionSubjects are what a motion can be ABOUT. Each carries its own required payload and its
// own ruler, which is why the CLI subgroups them rather than taking a --on flag: cobra cannot
// express "required only when --on=grade", so a flag-discerned subject would put three divergent
// contracts into hand-written RunE validation — a flag combination policed by prose, which is
// the shape this suite exists to remove.
var MotionSubjects = []string{"grade", "petition", "inquiry"}

// MotionVerdicts are the rulings, per subject. The KEY is `ruling` on every one of them and the
// flag is `--as` on every one of them, which is the point: §I of the plan names
// `ruling`/`ruling`/`response` as the structural defect, and a collapse that kept three
// spellings would have reproduced it inside the new group.
var MotionVerdicts = map[string][]EnumValue{
	"grade": {
		ev("accepted", "the filer is right and the grade moves — say so, then MOVE it with `merge regrade`; accepting without regrading is a channel with no consequence"),
		ev("rejected", "the grade stands. Your --reason is what the filer appeals against, so it carries the argument, not the conclusion"),
	},
	"petition": {
		ev("granted", "the objection holds. The relief BINDS the seats that come after, so state it as an instruction they can follow"),
		ev("denied", "the objection does not hold, and your reason must say why at the leaf — a refusal without one is a decoration the petitioner cannot contest"),
	},
	"inquiry": {
		ev("endorsed", "worth this run's time — blue should take it up"),
		ev("out_of_scope", "a real question, but not THIS question"),
		ev("too_thin", "in scope, and the hypothesis does not carry its budget as stated"),
	},
}

// MotionFields are the ENUMERATED payload fields a subject's filing carries, beyond the verdict.
//
// They live here for the reason MotionVerdicts does: the set is keyed on (subject, key) and
// EnumFields cannot express that — it is keyed by event TYPE, and one `motion` event carries a
// grade `dimension` or a petition `class` depending on what it is about.
//
// THIS TABLE EXISTS BECAUSE DELETING THE OLD VERBS EXPOSED THAT THE NEW ONES HAD LOST THEIR
// CONTRACT. `blue dispute --dimension` was enum-validated and `--proposed` was a pflag.Value that
// refused a non-grade at parse; `motion grade file` registered both as plain strings and accepted
// anything. The additive stage added the verb and nobody diffed its flag contract against the one
// it replaces — which is the half-state that reads as done, one layer below where that rule is
// usually applied.
var MotionFields = map[string]map[string][]EnumValue{
	"grade": {"dimension": {
		ev("severity", "how bad the defect is in itself"),
		ev("likelihood", "how likely the CONSEQUENCE is — never how likely the defect is to BE there, which is what one grade meant before v2 split them"),
		ev("impact", "how bad the consequence is if it lands"),
		// ALL FOUR AXES ARE THEIR FLAG NAMES. This value spent releases spelled `complexity_cost`
		// — the PAYLOAD key — while the flag was `--cx`, and the comment here documented the trap
		// rather than removing it: three dimensions matched their flags and the fourth matched
		// neither, which is a trap a seat walks into by learning the pattern from the other three
		// (measured). The flag is `--complexity` and so is the dimension; the payload key stays
		// `complexity_cost`, which is a schema name no seat types.
		ev("complexity", "what fixing it costs — the axis to contest when the fix is worth more than the defect"),
	}},
	"petition": {
		// FROM THE SCHEMA, not typed here. These four words were hand-listed while PetitionClass
		// carried a different four, and the write path resolves against the enum — so `ethical`
		// and `constitutional` were advertised in --help and REFUSED at the write, 22 of 40 times
		// in the fuzz. Derived, the two cannot disagree.
		"class": evsOf(recordpb.PetitionClass(0).Descriptor()),
		// WHO THE GRANTED RELIEF BINDS. Set on the RULING, not the filing: what a petitioner asks
		// for and what the bench orders are different facts, and only the second binds anyone.
		//
		// Measured (#360): a bench granted a petition in part, issued operative relief for the
		// coming round, and recorded in its own friction that it had "issued a direction to red
		// knowing it has no carrier". The engine threaded relief into exactly ONE prompt — blue's
		// — so relief addressed to red reached nothing. Routing needs an addressee, and there was
		// no field for one.
		// AND THE WORSE HALF: the table said `blue | red | both`, RulingBinds said
		// `all | filer | none`, NOTHING overlapped, and --binds is set exactly when a petition is
		// GRANTED — so no granted petition could be recorded at all.
		"binds": evsOf(recordpb.RulingBinds(0).Descriptor()),
	},
}

// MotionFieldEnum builds the enum entry for one of those fields, so the CLI's help and the write
// check are generated from one table rather than stated twice. It returns ok=false for a
// (subject, key) with no set, which is how a caller distinguishes "free text" from "empty set".
func MotionFieldEnum(subject, key string, flag string) (EnumField, bool) {
	values, ok := MotionFields[subject][key]
	if !ok {
		return EnumField{}, false
	}
	why := map[string]string{
		"dimension": "the ruling is matched to the filing on (gap, dimension) and the gap's grade is then read AT that dimension: an axis outside the four matches no ruling and reads no grade, so the motion is filed against nothing",
		"class":     "the class is what the seat is ASKING the bench to sit on, and the bench is convened per class; a fifth is a petition heard under whichever standard the ruling seat happened to imagine",
	}[key]
	return EnumField{Key: key, Flag: flag, Values: values, Why: why}, true
}

// motionIDOf returns the motion id an event JOINS ON, and whether the event is part of a motion
// exchange at all.
//
// THIS IS THE JOIN, AND IT IS THE ONE THING IN THIS FILE THAT MUST NOT BE INLINED PER CALLER.
// `motion`, `motion-rule` and `motion-appeal` each carry `motion_id` on their OWN body message —
// three different Go types holding one fact — so a reader that reaches for a body's id without
// covering all three drops a whole limb of the exchange silently. And the id it must reach for is
// `motion_id`: a ruling carries NO gap id, so attributing a ruling to a gap from the ruling event
// alone buckets every one of them under the empty key and reports zero forever. The gap lives on
// the FILING (`Motion.grade.gap_id`); the answer knows only the ask's id, and `Motions` is where
// the two are put together.
//
// ok=false means "not a motion event" — distinct from a motion event whose id is empty, which is a
// write the tool should never have produced and which callers still skip on their own terms.
func motionIDOf(e *Event) (string, bool) {
	body, ok := recordpb.Body(e)
	if !ok {
		return "", false
	}
	switch b := body.(type) {
	case *recordpb.Motion:
		return b.GetMotionId(), true
	case *recordpb.MotionRule:
		return b.GetMotionId(), true
	case *recordpb.MotionAppeal:
		return b.GetMotionId(), true
	}
	return "", false
}

// motionSubjectWord renders a subject as the word this tool's surfaces use, and the ONE
// disagreement it has to bridge is stated here rather than in each caller.
//
// The schema calls the third subject `MOTION_SUBJECT_DIRECTION`. Every surface outside the schema
// calls it `inquiry`: the `motion inquiry` subgroup, MotionSubjects above, RequireMotionSubjectRef's
// own branch, report/motions.go's switch, and the goldens that render "direction Q1" from a motion
// whose subject reads `inquiry`. Rendering it as `direction` would compile, and would invert
// RequireSubjectMatches for every direction motion in the run — the CLI passes `inquiry`, the
// record would answer `direction`, and a legitimate ruling would be refused with a message telling
// the seat to use the subgroup it already used.
//
// So the rename is bridged in exactly one place, visibly, until the vocabularies are settled. Every
// other value comes from the schema.
func motionSubjectWord(s recordpb.MotionSubject) string {
	if s == recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
		return "inquiry"
	}
	return enumWord(s)
}

// MotionSubjectEnum resolves a surface's subject word back to the schema value. The inverse of
// motionSubjectWord, sharing its one special case so the pair cannot drift apart.
func MotionSubjectEnum(word string) (recordpb.MotionSubject, bool) {
	if word == "inquiry" {
		return recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION, true
	}
	vd, ok := recordpb.BySpelling(recordpb.MotionSubject(0).Descriptor(), word)
	if !ok {
		return recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED, false
	}
	return recordpb.MotionSubject(vd.Number()), true
}

// motionRulingWord renders whichever ruling the event actually carries.
//
// THE VERDICT SET IS KEYED ON THE SUBJECT — that is what MotionVerdicts above says and what the
// schema's three ruling enums enforce — so there is no single field to read. An arm nobody set
// returns "", which is what Motion.Ruled() tests: a motion filed and never ruled must stay
// distinguishable from one ruled with a verdict this binary does not know.
func motionRulingWord(r *recordpb.MotionRule) string {
	switch v := r.GetRuling().(type) {
	case *recordpb.MotionRule_Grade:
		return enumWord(v.Grade)
	case *recordpb.MotionRule_Petition:
		return enumWord(v.Petition)
	case *recordpb.MotionRule_Direction:
		return enumWord(v.Direction)
	}
	return ""
}

// enumWord is the schema's own spelling of an enum value — recordpb.Spelling, reached through the
// descriptor the generated type already carries.
//
// IT IS DELIBERATELY NOT A SECOND VOCABULARY. Where recordpb.Spelling disagrees with the word a
// seat types (see the return note on separators), that is one defect in one central function and
// fixing it there corrects every caller, including this one, without an edit here.
func enumWord(v protoreflect.Enum) string {
	vd := v.Descriptor().Values().ByNumber(v.Number())
	if vd == nil {
		return ""
	}
	return recordpb.Spelling(vd)
}

// MintMotionID assigns the next run-unique motion id (M1, M2 …).
//
// Run-unique rather than round-scoped, for the reason a line of inquiry's is: a motion OUTLIVES the round
// that filed it — a grade dispute rejected in round 2 is re-disputed in round 3 and appealed to
// the bench in round 4 — so a round-scoped id would have to be re-minted to survive, and the
// re-mint is where the thread breaks.
func MintMotionID(run Run) (string, error) {
	var n int
	if _, err := queryRow(run, []any{&n},
		`SELECT count(*) FROM "motion" WHERE COALESCE("motion_id", '') != ''`); err != nil {
		return "", err
	}
	return fmt.Sprintf("M%d", n+1), nil
}

// Motion is one exchange after replay: what was asked, how it was ruled, and whether the filer
// pressed it further.
type Motion struct {
	ID      string
	Subject string
	Filer   string
	Round   int
	Basis   string // the ask, in the filer's words
	Relief  string // what the filer wants, stated so a seat can act on it without the argument

	// Ruling is empty until ruled. THE ABSENCE IS INFORMATION: a motion filed and never ruled
	// means the sitting did not happen, and the report says so rather than omitting the row.
	Ruling      string
	RulingBy    string
	RulingRound int
	Opinion     string

	// Appeal is the filer pressing on after a ruling — blue pursuing a direction ruled
	// out-of-scope, or re-disputing a rejected grade. `contests_ruling` was a bespoke field on
	// one of the three exchanges; here it is the same act on all of them.
	Appealed     bool
	AppealReason string

	// Subject-specific payload, carried rather than re-derived.
	Fields map[string]string
}

// Ruled reports whether the motion has an answer.
func (m Motion) Ruled() bool { return m.Ruling != "" }

// Motions replays the motion events into current state, in filing order.
//
// It reads the `motion` types only. Pre-collapse records carry `dispute`/`dispute-respond`,
// `petition`/`petition-rule` and `avenue-rule` instead, and those are handled by the DUAL-READ in
// compat.go — deliberately separate, so the shape of a motion is not bent to accommodate the
// three shapes it replaced.
func Motions(b *Board) []*Motion {
	byID := map[string]*Motion{}
	var order []string

	// Inquiry PROPOSALS, indexed by their id — the filing half of every direction motion. Gathered
	// in the same pass and read only when a ruling arrives, because a proposal nobody ruled on is
	// not a motion (see the motion-rule arm).
	type proposal struct {
		filer, basis string
		round        int
	}
	proposals := map[string]proposal{}

	// TWO PASSES, FOR THE REASON compat.go SPELLS OUT, AND I WROTE THE BUG HERE ANYWAY.
	//
	// Each seat writes its own shard and replay orders by TIMESTAMP across all of them, so the
	// shards interleave and A RULING CAN REPLAY BEFORE ITS FILING — the bench rules a petition blue
	// filed, and the two events live in different shards. A single pass that files-then-rules drops
	// every ruling that lands first, silently: the motion still renders, as an ask nobody answered.
	//
	// I documented exactly this in compat.go after the fixture caught it, and then shipped the same
	// single pass in the function compat.go exists to be the legacy twin of. What caught it was the
	// prose gate (#320) — "judge-petition/motion-rule prose absent from report" on 25 of 60 seeds —
	// not the reasoning that had already been written down one file over.
	// A PASS OF ITS OWN, for the same interleaving reason: blue proposes the line of inquiry and the merge
	// rules it, so the two live in different shards and the ruling can replay first. Gathered
	// inside pass 1 this map was read before it was filled, and a direction motion came out with
	// no filer, no round and no ask — rendering as an answer to a question nobody asked.
	for _, e := range b.Events {
		av, ok := recordpb.BodyAs[*recordpb.Avenue](e)
		// A MOVE IS NOT A PROPOSAL, and the discriminator is PRESENCE. The schema states
		// `supersedes_status` marks the event as a move, so a proposal does not carry the field at
		// all — `GetSupersedesStatus() == ""` would fold a field somebody set to the empty string
		// back into the proposal set, which is the read this migration exists to remove.
		if !ok || av.SupersedesStatus != nil {
			continue
		}
		if a := av.GetAvenueId(); a != "" {
			proposals[a] = proposal{filer: e.GetSeatId(), basis: av.GetLine(), round: int(e.GetRound())}
		}
	}

	for _, e := range b.Events {
		body, ok := recordpb.Body(e)
		if !ok {
			// No body at all. Not an empty motion — an event this pass has nothing to read, and
			// falling through would ask a nil message for an id.
			continue
		}
		switch f := body.(type) {
		case *recordpb.Motion:
			// THE FILING CARRIES THE ID EVERYTHING ELSE JOINS ON. It is read from the filing's own
			// message rather than from a key that might be present on anything.
			id := f.GetMotionId()
			if id == "" {
				continue
			}
			m, ok := byID[id]
			if !ok {
				m = &Motion{ID: id, Fields: map[string]string{}}
				byID[id] = m
				order = append(order, id)
			}
			m.Subject, m.Filer, m.Round = motionSubjectWord(f.GetSubject()), e.GetSeatId(), int(e.GetRound())
			m.Basis, m.Relief = f.GetBasis(), f.GetRelief()
			// THE SUBJECT-SPECIFIC FIELDS COME OFF THE SUBJECT'S OWN MESSAGE, which is what the
			// `filing` oneof buys: a grade motion cannot carry a petition's `class`, so the loop
			// over five payload keys that used to stand in for that guarantee is gone. The KEYS
			// stay as they were — report/motions.go and motionview.go read `gap_id`, `dimension`,
			// `proposed`, `class` and `inquiry_id` by name, and `inquiry_id` in particular is the
			// surface spelling of what the schema calls `avenue_id`.
			//
			// Each is set only when PRESENT. The old read skipped an empty string; an absent enum
			// would otherwise render as its UNSPECIFIED spelling and put a word on the report where
			// the seat wrote nothing.
			switch fil := f.GetFiling().(type) {
			case *recordpb.Motion_Grade:
				if fil.Grade.GapId != nil {
					m.Fields["gap_id"] = fil.Grade.GetGapId()
				}
				if fil.Grade.Dimension != nil {
					m.Fields["dimension"] = enumWord(fil.Grade.GetDimension())
				}
				if fil.Grade.Proposed != nil {
					m.Fields["proposed"] = enumWord(fil.Grade.GetProposed())
				}
			case *recordpb.Motion_Petition:
				if fil.Petition.Class != nil {
					m.Fields["class"] = enumWord(fil.Petition.GetClass())
				}
			case *recordpb.Motion_Direction:
				if fil.Direction.AvenueId != nil {
					m.Fields["inquiry_id"] = fil.Direction.GetAvenueId()
				}
			}
		case *recordpb.MotionRule:
			id := f.GetMotionId()
			if id == "" {
				continue
			}
			// PASS 1 only CREATES. A direction motion has no filing event of its own, so its
			// ruling is also its creation; every other subject is created by its `motion` event,
			// which this pass may not have reached yet. The ruling itself is attached in pass 2.
			if _, ok := byID[id]; !ok {
				if f.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
					continue // a ruling naming no filing; RequireMotionSubjectRef refuses this at the write
				}
				// A DIRECTION MOTION IS CREATED BY ITS RULING, and that is not a special case
				// bolted on — it is the only shape that does not invent identity. `direction` has
				// no `file` verb because the PROPOSAL is the filing (`blue line of inquiry`), so the id it
				// joins on is the line of inquiry's own A-id: minted, refusable, already on the record.
				//
				// The rejected alternative was minting an M-id at propose time. It would have
				// filed a motion for every line blue ever floated — ~60 a run in the fuzz against
				// 45 rulings — so the report's "N motions received no ruling" would have been
				// dominated by lines nobody was ever going to rule on, and the one signal that
				// message exists to carry (a sitting that did not happen) would have drowned in it.
				//
				// compat.go builds the legacy direction motion the same way, from `avenue-rule`.
				// One shape for both vocabularies, which is what makes the dual-read a translation
				// rather than a second model.
				m := &Motion{ID: id, Subject: "inquiry", Fields: map[string]string{"inquiry_id": id}}
				if p, ok := proposals[id]; ok {
					m.Filer, m.Round, m.Basis = p.filer, p.round, p.basis
				}
				byID[id] = m
				order = append(order, id)
			}
		}
	}

	// PASS 2 attaches every answer. Order-independent by construction, which is the only safe
	// assumption about a multi-shard log.
	//
	// THE ANSWER KNOWS ONLY THE ASK'S ID. A `motion-rule` carries `motion_id` and nothing else that
	// identifies what it is about — no gap id, no avenue id — so this lookup IS the attribution.
	// Reading a ruling's subject matter from the ruling event alone would key every one of them on
	// the empty string and report a board with no rulings on it.
	for _, e := range b.Events {
		id, ok := motionIDOf(e)
		if !ok || id == "" {
			continue
		}
		m, ok := byID[id]
		if !ok {
			continue // a ruling or appeal naming no filing; the write-side ref checks refuse these
		}
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch f := body.(type) {
		case *recordpb.MotionRule:
			m.Ruling, m.RulingBy, m.RulingRound = motionRulingWord(f), e.GetSeatId(), int(e.GetRound())
			m.Opinion = f.GetOpinion()
		case *recordpb.MotionAppeal:
			m.Appealed, m.AppealReason = true, f.GetReason()
		}
	}

	out := make([]*Motion, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// RequireMotionSubjectRef refuses a ruling or appeal naming a motion no filing created — the same
// discipline every other cross-reference gets, for the same reason: a dangling reference is
// accepted at write time and dropped at replay, where nobody sees it go.
//
// It takes the SUBJECT because the subjects do not share a filing verb. `grade` and `petition` are
// filed by `motion <subject> file` and join on the M-id it mints; `direction` has no file verb —
// the proposal is the filing — so it joins on the LINE's own id, and the thing that must exist
// is the line of inquiry, not a motion event. Passing the subject keeps that difference in one place
// instead of pushing it into each RunE.
func RequireMotionSubjectRef(run Run, subject recordpb.MotionSubject, id string) error {
	if id == "" {
		return fmt.Errorf("record: --id is required — a ruling names the motion it answers, and that join is the whole of #312")
	}
	if subject == recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
		return RequireInquiryRef(run, id)
	}
	found, err := recordHas(run, `SELECT 1 FROM "motion" WHERE "motion_id" = ? LIMIT 1`, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("record: --id names motion %s, which no filing created — a dangling reference is accepted here and dropped at replay", id)
}

// RequireSubjectMatches refuses a ruling whose SUBGROUP disagrees with the motion's own subject.
//
// MEASURED BY PROBING, and it is the defect the collapse was supposed to remove, reintroduced one
// level down. `motion <subject> rule` takes its subject from its POSITION IN THE TREE — which
// subgroup you typed — and validated the verdict against that, never against the subject the
// motion was actually filed under. So `motion grade rule --id M1` on a PETITION motion was
// accepted: it bypassed the gavel (the merge ruling what only the bench may rule) AND the verdict
// vocabulary (`accepted`, which is not a petition ruling at all). The report then rendered
// "petition (safety) … ruled accepted".
//
// A fact recovered from tree position rather than read from the record is exactly what
// facts-are-fields is about. The record carries the subject; this reads it.
func RequireSubjectMatches(run Run, subject, id string) error {
	got, err := motionSubjectOf(run, id)
	if err != nil {
		return err
	}
	if got != subject {
		return fmt.Errorf("record: motion %s was filed as a %s motion and you are ruling it as a %s — rule it under `motion %s rule`. The subject decides BOTH who holds the gavel and which verdicts exist, so ruling under the wrong one answers with a vocabulary the motion does not have",
			id, got, subject, got)
	}
	return nil
}

// motionSubjectOf reports what a motion is ABOUT, from the record rather than from the caller.
//
// A direction has no filing event — the proposal is the filing — so an id that resolves to an
// line of inquiry IS a direction motion by construction.
func motionSubjectOf(run Run, id string) (string, error) {
	// The stored word is the schema's spelling; the SURFACE word goes through the same
	// motionSubjectWord mapping the fold used (direction speaks as `inquiry`).
	var word sql.NullString
	found, err := queryRow(run, []any{&word},
		`SELECT "subject" FROM "motion" WHERE "motion_id" = ? ORDER BY "event_id" LIMIT 1`, id)
	if err != nil {
		return "", err
	}
	if found {
		if vd, ok := recordpb.BySpelling(recordpb.MotionSubject(0).Descriptor(), word.String); ok {
			return motionSubjectWord(recordpb.MotionSubject(vd.Number())), nil
		}
		return word.String, nil
	}
	// ANY avenue event, proposal or move: the question is whether this id names a line of
	// inquiry at all, and a line that has since moved status is still a line.
	isLine, err := recordHas(run, `SELECT 1 FROM "avenue" WHERE "avenue_id" = ? LIMIT 1`, id)
	if err != nil {
		return "", err
	}
	if isLine {
		return "inquiry", nil
	}
	return "", fmt.Errorf("record: --id names motion %s, which no filing created — a dangling reference is accepted here and dropped at replay", id)
}

// RequireUnruledMotion refuses a SECOND ruling on a motion already answered.
//
// MEASURED BY PROBING: a petition ruled `accepted` by the merge was then ruled `denied` by the
// bench, both accepted, and the report showed ONE of them with no sign the other existed. A
// direction ruled `endorsed` was re-ruled `out-of-scope` the same way. Replay keeps whichever the
// ordering happens to favour, so the answer a reader sees is decided by shard interleaving.
//
// Two writers disagreeing about one fate is the defect the line of inquiry code already guards against by
// giving moves a single writer; a ruling had no such guard. The escalation path is an APPEAL,
// which is a new event that preserves both positions, rather than a second ruling that erases one.
func RequireUnruledMotion(run Run, id string) error {
	// motion_answers is the one statement of first-wins: the FIRST ruling is the one quoted,
	// its word whichever arm the rule carried, "" when it carried none. A row whose ruled_by
	// is NULL is an appeal with no ruling — not a ruling, so it does not refuse.
	var word, seat sql.NullString
	found, err := queryRow(run, []any{&word, &seat},
		`SELECT "ruling", "ruled_by" FROM "motion_answers" WHERE "motion_id" = ? AND "ruled_by" IS NOT NULL`, id)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("record: motion %s is already ruled %q by %s. A second ruling does not overturn the first — the second is simply the one a later reader sees, and the first stops being the answer. To press it, `appeal` it: an appeal keeps both positions on the record, which is the whole reason a ruling is an argument rather than a command",
			id, word.String, seat.String)
	}
	return nil
}

// RequireUnappealedMotion refuses a SECOND appeal on a motion already appealed.
//
// FOUND BY THE STATE GRAPH (#673), and it is RequireUnruledMotion's defect one verb over. Probing
// every act from every state of a motion, `appeal` on an already-appealed motion was ACCEPTED and
// left the state alone while rewriting `appeal_reason` — three appeals in a row, each silently
// replacing the last, the report showing only the third.
//
// The reasoning above applies unchanged: both events stay on the record and replay keeps whichever
// came last, so the earlier position simply stops being the answer. It is WORSE here than for a
// ruling, because an appeal is not an answer to be overturned — debate.js tells blue "the appeal is
// where your ARGUMENT is recorded", so the thing quietly dropped is the argument itself.
//
// A SECOND APPEAL IS NOT THE MOVE FOR NEW GROUNDS, and the refusal says what is: a grade dispute
// pressed again belongs in a NEW motion for the new round, which is the path the engine already
// drives (`grade_dispute_re_raised`). That keeps both arguments, which is the whole point of an
// appeal being an event rather than a field.
func RequireUnappealedMotion(run Run, id string) error {
	var seat, reason sql.NullString
	found, err := queryRow(run, []any{&seat, &reason},
		`SELECT "appealed_by", "appeal_reason" FROM "motion_answers"
		  WHERE "motion_id" = ? AND "appealed_by" IS NOT NULL`, id)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("record: motion %s is already appealed by %s (%q). A second appeal does not add to the first — it REPLACES it in every reader, and the argument already on the record stops being the one anybody sees. If you are pressing on new grounds, file a NEW motion for this round: two motions keep two arguments, which is what an appeal being an event rather than a field is for",
			id, seat.String, reason.String)
	}
	return nil
}

// RequireRuledMotion additionally refuses an appeal against a motion NOBODY HAS RULED.
//
// An appeal is the filer pressing on AFTER an answer; against no answer there is nothing to press
// against, and the event would replay as a motion both unruled and appealed — a state the report
// has no honest sentence for. The check covers every subject rather than the one that surfaced it:
// appealing an unruled grade is the same nonsense as appealing an unruled direction, and fixing
// only the instance is how the class survives.
func RequireRuledMotion(run Run, subject recordpb.MotionSubject, id string) error {
	if err := RequireMotionSubjectRef(run, subject, id); err != nil {
		return err
	}
	found, err := recordHas(run,
		`SELECT 1 FROM "motion_answers" WHERE "motion_id" = ? AND "ruled_by" IS NOT NULL`, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("record: %s motion %s has no ruling to appeal — an appeal presses on after an answer, and there is no answer on the record yet", subject, id)
}

// MotionVerdictEnum builds the enum entry for a subject's ruling, so the CLI's help and the write
// check are generated from one table rather than stated twice.
func MotionVerdictEnum(subject string) EnumField {
	return EnumField{
		Key: "ruling", Flag: flags.As, Values: MotionVerdicts[subject],
		Why: "the ruling is what BINDS the coming seats, and every downstream reader switches on it; an unrecognized verdict reads as no ruling at all, so a refusal silently becomes permission",
	}
}

// THE DUAL-READ IS GONE, and this is where it was.
//
// Before the motion collapse (#344) the three exchanges were written as `dispute`/`dispute-respond`,
// `petition`/`petition-rule` and `avenue`/`avenue-rule`. `compat.go` translated those retired shapes
// through an `AllMotions` seam so a stored pre-collapse record would not render `0 filed / 0 ruled`
// — a plausible zero indistinguishable from a run that genuinely had no disputes.
//
// That reasoning was sound and its PREMISE was not. It declared itself permanent on the grounds
// that "stored runs are re-read" and that installing projects hold records this repo cannot see.
// Checked 2026-08-16 rather than assumed: the only records anywhere carrying the retired types were
// `internal/record/testdata/pre-motion-run` and `pre-motion-real-run` — two fixtures created to
// exercise the dual-read — plus one research run referenced solely in code comments. The
// compatibility code's entire evidence base was the fixtures written to justify it.
//
// Two sessions reached that conclusion independently on the same instruction and deleted it in
// parallel; the merge kept `Motions` as the single read and dropped the `AllMotions` seam with it.
//
// THE COST IF THE PREMISE IS WRONG, stated rather than left to be discovered: a record written
// before #344 renders an empty motions section and `0 filed / 0 ruled`, exactly the plausible zero
// compat.go existed to prevent.
