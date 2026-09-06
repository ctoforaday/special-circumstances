package motion

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// file: a seat asks for something, and the tool assigns the id everything else joins on.
func newFile(subject string, required []string) *cobra.Command {
	c := &cobra.Command{
		Use:   "file",
		Short: "file a " + subject + " motion — the tool assigns its id",
		// EACH SUBJECT'S `file` SAYS WHERE THE BOUNDARY IS. The three subjects write ONE event
		// type with different required flags and different gavels, so a seat that opens the wrong
		// one meets a refusal about a flag rather than about the subject it should have named.
		//
		// The tree knowing about the split is not the same as the seat knowing: this page is what
		// a seat actually reads, and it was blank. It went unnoticed because `motion` sat at the
		// ROOT until the surface became seat-scoped, so the sibling gate never reached it.
		Long: "file a " + subject + " motion — the tool assigns its id.\n\n" +
			"THREE SUBJECTS, ONE EVENT, DIFFERENT CONTRACTS. `motion grade file` disputes a gap's " +
			"grade and is ruled by the merge; `motion petition file` raises an ethical, safety, " +
			"integrity or constitutional objection and is ruled by the BENCH, before the debate " +
			"continues; a direction needs no file verb at all — proposing the line of inquiry IS " +
			"the filing, and only its ruling is a motion.\n\n" +
			"Any seat may file. Exactly one rules, and `rule` appears only on that seat's surface.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := seat.Of(cmd)
			// AND IT MUST BE THE RUN THE ENGINE DISPATCHED. Same reason as the seat check below,
			// on the other axis: Begin is not on this path, so the run refusal is not either, and
			// these verbs WRITE — a motion filed against a contradicted run directory is the
			// attribution failure the check exists for, one field over.
			run, err := s.Run()
			if err != nil {
				return err
			}
			// THE FILER MUST BE A SEAT THE ENGINE CREATED. These verbs read the context with
			// seat.Of, which only parses flags — seat.Begin, which runs the identity checks, is
			// never on this path. Measured: `motion grade file --seat-id totally-invented`
			// recorded a motion, while `blue position` with the same id was refused.
			if err := record.RequireDispatchedSeat(s.SeatID); err != nil {
				return err
			}
			basis, err := prose(cmd, "file", "the ASK in your own words — a motion with no argument is a demand, and the ruler has nothing to rule on")
			if err != nil {
				return err
			}
			for _, f := range required {
				if strings.TrimSpace(seat.Str(cmd, f)) == "" {
					return feov.Errorf(feov.Validation,
						"a %s motion requires --%s. The subjects have different contracts, which is why they are subgroups: cobra cannot express \"required only when the subject is %s\", so the requirement lives here where it can refuse",
						subject, f, subject)
				}
			}
			id, err := record.MintMotionID(run)
			if err != nil {
				return err
			}
			// THE UNTYPED LOOP CANNOT SURVIVE A TYPED FILING, and that is the migration working
			// rather than a cost of it. `for _, f := range required { p.Set(payloadKey(f), …) }`
			// wrote whatever the flags happened to hold under whatever key payloadKey returned —
			// which is how `--petition-class` landed under the key `petition-class` while every
			// reader looked for `class`, and nothing could see the two disagree. The filing is now
			// a oneof, so a grade's fields cannot be written onto a petition at all.
			subj, ok := record.MotionSubjectEnum(subject)
			if !ok {
				return feov.Errorf(feov.Validation, "motion %s file: %q is not a motion subject", subject, subject)
			}
			body := &recordpb.Motion{
				MotionId: proto.String(id),
				Subject:  &subj,
				Basis:    proto.String(basis),
			}
			switch subj {
			case recordpb.MotionSubject_MOTION_SUBJECT_GRADE:
				dim, known := record.GradeDimensionOf(seat.Str(cmd, flags.Dimension))
				if !known {
					return feov.Errorf(feov.Validation, "motion grade file: %q is not a grade dimension", seat.Str(cmd, flags.Dimension))
				}
				prop, known := record.GradeOf(seat.Str(cmd, flags.Proposed))
				if !known {
					return feov.Errorf(feov.Validation, "motion grade file: %q is not a grade", seat.Str(cmd, flags.Proposed))
				}
				body.Filing = &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
					GapId:     proto.String(seat.Str(cmd, flags.ID)),
					Dimension: &dim,
					Proposed:  &prop,
				}}
			case recordpb.MotionSubject_MOTION_SUBJECT_PETITION:
				cls, known := record.PetitionClassOf(seat.Str(cmd, flags.Class))
				if !known {
					return feov.Errorf(feov.Validation, "motion petition file: %q is not a petition class", seat.Str(cmd, flags.Class))
				}
				body.Filing = &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: &cls}}
			}
			if r := seat.Str(cmd, flags.Relief); r != "" {
				body.Relief = proto.String(r)
			}
			if _, err := record.Append(s.Identity(), body); err != nil {
				return err
			}
			return seat.Emit(cmd, filed{ID: id, Subject: subject}, nil)
		},
	}
	seat.Prose(c)
	for _, f := range required {
		// THE FLAG'S TYPE IS THE FIRST REFUSAL, and the first draft threw it away. `--proposed`
		// takes a GRADE and is a pflag.Value that refuses a non-grade at parse — before any RunE
		// runs, with the help and the refusal generated from one list. `--dimension` and
		// `--petition-class` take enumerated sets. Registering all three as bare strings, as this
		// loop originally did, accepted anything and deferred the error to a reader that would
		// never look.
		switch {
		case f == flags.Proposed:
			p := new(flags.GradeValue)
			c.Flags().Var(p, f, flags.GradeUsage("the grade you say it should be"))
		default:
			if e, ok := record.MotionFieldEnum(subject, payloadKey(f), f); ok {
				enumhelp.Flag(c, f, e, "REQUIRED for a "+subject+" motion")
				continue
			}
			c.Flags().String(f, "", "REQUIRED for a "+subject+" motion")
		}
	}
	// THE RECORD TYPE, DECLARED. These verbs are built as raw cobra commands rather than through
	// seat.New, so they carried no `records` annotation — and the contract gate joins commands to
	// their declared requirements through exactly that. Three verbs sat outside it: nothing checked
	// that a field the record requires is marked REQUIRED in their help, which is the one place a
	// seat reads what it must supply.
	seat.Records(c, "motion")
	// THE MOTION ID IS ASSIGNED BY THE TOOL on a filing — a seat that chose its own would collide
	// with another seat's, which is the join #312 exists to make reliable. Declared at the verb
	// that does the assigning.
	seat.Supplies(c, "motion_id", "the tool assigns it on filing; a ruling and an appeal name it with --id")
	return c
}

// payloadKey maps a flag word to the payload key it writes.
//
// THE TWO DIFFER AND THE FIRST DRAFT ASSUMED THEY DID NOT, which is the composition failure
// facts-are-fields is about, committed inside the change meant to remove one. `--petition-class`
// wrote its value under the key `petition-class`, while record.validate and record.Motions both
// read `class` — so the enum check silently matched nothing and the report's petition head
// rendered empty. Both sides read this function now, and a key with no entry is the flag word,
// which is true for every flag that has not needed a mapping.
//
//   - `--id` names WHAT it identifies: a motion's own id is `motion_id`, so a gap reference
//     cannot also be `id` without the two colliding in one payload.
//   - `--petition-class` is the flag word because `--class` is taken elsewhere in the tree; the
//     RECORD calls it `class`, and the record is what every consumer reads.
func payloadKey(flag string) string {
	switch flag {
	case flags.ID:
		return "gap_id"
	case flags.Class:
		return "class"
	}
	return flag
}

// rule: the ruler answers, on the motion's id.
func newRule(subject, ruler string) *cobra.Command {
	e := record.MotionVerdictEnum(subject)
	c := &cobra.Command{
		Use:   "rule",
		Short: "rule on a " + subject + " motion (the " + ruler + " seat's)",
		// THE THIRD PAGE THAT HAD TO SAY WHICH SUBJECT IT IS — `file` and `appeal` carry one for
		// the same reason. `rule` differs from those two in that it is not on every seat's
		// surface at all, so the page a seat CAN open should also say which ones it cannot.
		Long: "rule on a " + subject + " motion — this verb is the " + ruler + " seat's, and it " +
			"appears only on that surface.\n\n" +
			"THREE SUBJECTS, THREE GAVELS. A grade dispute and a direction are ruled by the MERGE; " +
			"a petition is ruled by the BENCH, before the debate continues. A motion is filed by " +
			"any seat and ruled by one — that asymmetry is the mechanism, not an obstacle, and it " +
			"is why `rule` is missing from the surfaces that do not hold the gavel.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := seat.Of(cmd)
			// AND IT MUST BE THE RUN THE ENGINE DISPATCHED. Same reason as the seat check below,
			// on the other axis: Begin is not on this path, so the run refusal is not either, and
			// these verbs WRITE — a motion filed against a contradicted run directory is the
			// attribution failure the check exists for, one field over.
			run, err := s.Run()
			if err != nil {
				return err
			}
			// THE FILER MUST BE A SEAT THE ENGINE CREATED. These verbs read the context with
			// seat.Of, which only parses flags — seat.Begin, which runs the identity checks, is
			// never on this path. Measured: `motion grade file --seat-id totally-invented`
			// recorded a motion, while `blue position` with the same id was refused.
			if err := record.RequireDispatchedSeat(s.SeatID); err != nil {
				return err
			}
			id := seat.Str(cmd, flags.ID)
			// ORDER IS THE MESSAGE. The motion's own subject is established FIRST, because every
			// later refusal is phrased in terms of it: a lens typing `motion grade rule` at a
			// petition should be told it named the wrong subgroup, not that grade motions belong
			// to the merge — which is true, irrelevant, and sends it to the wrong fix.
			if err := record.RequireMotionSubjectRef(run, mustSubject(subject), id); err != nil {
				return err
			}
			if err := record.RequireSubjectMatches(run, subject, id); err != nil {
				return err
			}
			// NO requireRuler HERE ANY MORE. This verb only exists in the gavel-holder's tree, so
			// a seat that cannot rule this subject cannot name the command — the same boundary the
			// verb set draws everywhere else, instead of a runtime comparison of two copies of the
			// acting role.
			// A motion is answered ONCE; pressing it is an appeal, which keeps both positions.
			if err := record.RequireUnruledMotion(run, id); err != nil {
				return err
			}
			opinion, err := prose(cmd, "rule", "an unreasoned ruling is the decoration the filer cannot contest, and contesting it is the whole reason a ruling is not a command")
			if err != nil {
				return err
			}
			// THE VERDICT SET IS KEYED ON (SUBJECT, RULING) and the schema says so with a oneof:
			// granted|denied can only reach the petition arm, accepted|rejected the grade arm. The
			// old payload carried `subject` and `ruling` as two independent strings that a writer
			// could set inconsistently, and record.go's validate existed to catch exactly that.
			subj := mustSubject(subject)
			body := &recordpb.MotionRule{
				MotionId: proto.String(id),
				Subject:  &subj,
				Opinion:  proto.String(opinion),
			}
			word := seat.Str(cmd, flags.As)
			switch subj {
			case recordpb.MotionSubject_MOTION_SUBJECT_GRADE:
				v, ok := enumOf[recordpb.GradeRuling](recordpb.GradeRuling(0).Descriptor(), word)
				if !ok {
					return feov.Errorf(feov.Validation, "motion grade rule: %q is not a ruling on a grade motion", word)
				}
				body.Ruling = &recordpb.MotionRule_Grade{Grade: v}
			case recordpb.MotionSubject_MOTION_SUBJECT_PETITION:
				v, ok := enumOf[recordpb.PetitionRuling](recordpb.PetitionRuling(0).Descriptor(), word)
				if !ok {
					return feov.Errorf(feov.Validation, "motion petition rule: %q is not a ruling on a petition", word)
				}
				body.Ruling = &recordpb.MotionRule_Petition{Petition: v}
			case recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION:
				v, ok := enumOf[recordpb.DirectionRuling](recordpb.DirectionRuling(0).Descriptor(), word)
				if !ok {
					return feov.Errorf(feov.Validation, "motion inquiry rule: %q is not a ruling on a direction", word)
				}
				body.Ruling = &recordpb.MotionRule_Direction{Direction: v}
			}
			// WHO THE RELIEF BINDS travels with the ruling that grants it (#360). Optional,
			// because a denial binds nobody — and a grant that names no addressee is exactly the
			// state the bench filed friction about: a direction issued knowing it had no carrier.
			if w := seat.Str(cmd, flags.Binds); w != "" {
				b, ok := enumOf[recordpb.RulingBinds](recordpb.RulingBinds(0).Descriptor(), w)
				if !ok {
					return feov.Errorf(feov.Validation, "motion %s rule: %q is not an addressee the relief can bind", subject, w)
				}
				body.Binds = &b
			}
			if _, err := record.Append(s.Identity(), body); err != nil {
				return err
			}
			return seat.Emit(cmd, ruled{ID: id, Ruling: seat.Str(cmd, flags.As)}, nil)
		},
	}
	seat.Prose(c)
	c.Flags().String(flags.ID, "", refHelp(subject))
	enumhelp.Flag(c, flags.As, e, "your ruling")
	if be, ok := record.MotionFieldEnum(subject, "binds", flags.Binds); ok {
		enumhelp.Flag(c, flags.Binds, be, "who granted relief BINDS — set it when you grant, or the relief reaches no prompt and nothing reports that it did not")
	}
	// The record type, for the contract gate — see newFile.
	seat.Records(c, "motion_rule")
	return c
}

// appeal: a seat presses a motion on after a ruling.
//
// STANDING IS DELIBERATELY OPEN, and the first draft's comment said "the filer" while the code
// checked nobody — so the doc was wrong rather than the code. A motion belongs to the RUN, not to
// the seat that filed it: a lens files a safety petition and it is BLUE that the granted relief
// binds, so restricting the appeal to the filer would leave the seat actually affected with no
// channel. Any registered seat may appeal; who did is on the event.
//
// `contests_ruling` was a bespoke field on ONE of the three exchanges — blue pursuing a direction
// red ruled out-of-scope. Here it is the same act on every subject that has one, which is what
// the collapse buys: re-disputing a rejected grade and pursuing a refused line stop being two
// unrelated mechanisms.
func newAppeal(subject string) *cobra.Command {
	c := &cobra.Command{
		Use:   "appeal",
		Short: "press a " + subject + " motion after a ruling — a ruling is an argument, not a command",
		// SAME EVENT, TWO SUBJECTS, AND THE PAGE HAS TO SAY WHICH IS WHICH — the reason `file`
		// carries one, for the same reason: a seat picks between them by opening one of them.
		Long: "press a " + subject + " motion after a ruling — a ruling is an ARGUMENT, not a " +
			"command, so the losing side may answer it on the record.\n\n" +
			"TWO SUBJECTS TAKE AN APPEAL. `motion grade appeal` presses a grade dispute the merge " +
			"rejected; `motion inquiry appeal` presses a line of inquiry red ruled out-of-scope or " +
			"too-thin, and it is filed whether or not blue also pursues the line — separating the " +
			"argument from the act is the whole point of the verb.\n\n" +
			"A PETITION HAS NO APPEAL, and that absence is the design rather than an omission: it " +
			"is heard by the bench BEFORE the debate continues, so there is nothing to escalate to.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := seat.Of(cmd)
			// AND IT MUST BE THE RUN THE ENGINE DISPATCHED. Same reason as the seat check below,
			// on the other axis: Begin is not on this path, so the run refusal is not either, and
			// these verbs WRITE — a motion filed against a contradicted run directory is the
			// attribution failure the check exists for, one field over.
			run, err := s.Run()
			if err != nil {
				return err
			}
			// THE FILER MUST BE A SEAT THE ENGINE CREATED. These verbs read the context with
			// seat.Of, which only parses flags — seat.Begin, which runs the identity checks, is
			// never on this path. Measured: `motion grade file --seat-id totally-invented`
			// recorded a motion, while `blue position` with the same id was refused.
			if err := record.RequireDispatchedSeat(s.SeatID); err != nil {
				return err
			}
			id := seat.Str(cmd, flags.ID)
			if err := record.RequireRuledMotion(run, mustSubject(subject), id); err != nil {
				return err
			}
			if err := record.RequireSubjectMatches(run, subject, id); err != nil {
				return err
			}
			// #673: a second appeal REPLACED the first in every reader. The state graph found it
			// by probing every act from every state — accepted, the state unchanged, the argument
			// rewritten.
			if err := record.RequireUnappealedMotion(run, id); err != nil {
				return err
			}
			reason, err := prose(cmd, "appeal", "why you are pressing on. Going against a ruling without saying why is the disagreement disappearing, which is what the record exists to prevent")
			if err != nil {
				return err
			}
			subj := mustSubject(subject)
			if _, err := record.Append(s.Identity(), &recordpb.MotionAppeal{
				MotionId: proto.String(id),
				Subject:  &subj,
				Reason:   proto.String(reason),
			}); err != nil {
				return err
			}
			return seat.Emit(cmd, appealed{ID: id}, nil)
		},
	}
	seat.Prose(c)
	c.Flags().String(flags.ID, "", refHelp(subject)+" — the motion being appealed, which must already have been ruled")
	// The record type, for the contract gate — see newFile.
	seat.Records(c, "motion_appeal")
	return c
}

// refHelp names WHICH id the subject joins on. `direction` has no filing verb, so its id is the
// line of inquiry's; saying "the motion id" there would send a seat looking for an M-number that does not
// exist, and a seat that cannot find the id it was told to pass logs friction and works around the
// verb — losing the capability for the run rather than reporting a wrong flag.
// refHelp is the --id line, and it carries the REQUIRED marker because the flag is.
//
// It did not, and nothing caught that until the surface became seat-scoped: `motion` used to sit
// at the ROOT, so the marker gate — which walks each seat's tree and skips anything whose path
// does not begin inside one — never reached it. Moving motion inside each seat's tree brought
// three subgroups into a gate that had never covered them, and all three were unmarked.
//
// A seat reading an unmarked flag supplies it or does not, learns which by being refused, and
// spends the turn. The refusal itself is real (RequireMotionSubjectRef resolves the id against the
// record), so this is the help disagreeing with the enforcement rather than a missing rule.
// refHelp says WHICH id the verb wants. It no longer writes "REQUIRED — " itself: seat.MarkTree
// walks this tree now and markRequired is the one writer of that word, so a hand-written copy here
// rendered "REQUIRED — REQUIRED — the motion id".
func refHelp(subject string) string {
	if subject == "inquiry" {
		return "the LINE-OF-INQUIRY id (Q1, Q2 …): a direction's filing is the proposal, so it joins on the line of inquiry's own id, not an M-number"
	}
	return "the motion id (M1, M2 …)"
}

type filed struct {
	ID      string `json:"motion_id"`
	Subject string `json:"subject"`
}

func (r filed) Human() string { return "motion " + r.ID + " filed (" + r.Subject + ")" }

type ruled struct {
	ID     string `json:"motion_id"`
	Ruling string `json:"ruling"`
}

func (r ruled) Human() string { return "motion " + r.ID + " ruled " + r.Ruling }

type appealed struct {
	ID string `json:"motion_id"`
}

func (r appealed) Human() string { return "motion " + r.ID + " appealed" }

// mustSubject resolves a subject word registered by this package's own command tree, where an
// unknown word is a programming error rather than seat input: subject() is called with literals.
// The zero would be UNSPECIFIED, which reads downstream as a motion about nothing.
func mustSubject(word string) recordpb.MotionSubject {
	subj, ok := record.MotionSubjectEnum(word)
	if !ok {
		panic("motion: " + word + " is registered as a subgroup and is not a motion subject — the command tree and the schema disagree")
	}
	return subj
}

// enumOf resolves a seat's word to a typed ruling. It mirrors record.enumOf, which is unexported
// there for the same reason it is unexported here: the EXPORTED pairs (GradeOf, ConfidenceOf …)
// carry each enum's own fold rules, and a caller reaching past them to a generic would lose those.
// The three ruling sets have no folds — they are the schema's spelling exactly.
func enumOf[E ~int32](d protoreflect.EnumDescriptor, word string) (E, bool) {
	vd, ok := recordpb.BySpelling(d, word)
	if !ok {
		return E(0), false
	}
	return E(vd.Number()), true
}
