package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// This file is the SHARD ERA's translation set (2026-08-22/23 archives): the vocabulary as
// it stood before the SQLite cutover. Every mapping here is either the schema's own recorded
// history (the disposition rename table is commit a1e8e260's, verbatim; inquiry-support's
// field deletions are commit f516261a's ruling) or a field rename the current message spells
// out. Nothing here guesses: a word or value neither era declares still refuses by name.

// dispositionRename is the closure-class → disposition rename, exactly as commit a1e8e260
// authored it. The era wrote the words on the left; the schema renamed them because they
// named the procedural route rather than the outcome.
var dispositionRename = map[string]string{
	"closed":                   "repaired",
	"closed_with_regression":   "repaired_with_regression",
	"rebuttal_sustained":       "not_a_defect",
	"risk_accepted":            "defect_accepted",
	"routed_to_infrastructure": "defect_owed_elsewhere",
}

// eraEntries is the shard-era overlay. Words also written by the SQLite era (friction,
// friction_none, opinion) are handled by their entries in entries.go, which accept both
// eras' columns.
func eraEntries() Registry {
	return Registry{
		// The era's universal `reason` column, renamed onto each body's own prose field.
		"blue_edit": mapped("blue_edit", ren{"reason": "text"}, nil, nil),
		"verify":    mapped("verify", ren{"reason": "text"}, nil, nil),
		"finding":   mapped("finding", ren{"reason": "text"}, nil, nil),
		"position":  mapped("position", ren{"reason": "text"}, nil, nil),
		"revision":  mapped("revision", ren{"reason": "text"}, nil, nil),
		"closing":   mapped("closing", ren{"reason": "text"}, nil, nil),
		"reproduce": mapped("reproduce", ren{"reason": "note"}, nil, nil),
		"certify":   mapped("certify", ren{"reason": "statement"}, nil, nil),
		"regrade":   mapped("regrade", ren{"reason": "basis"}, nil, nil),
		"outcome":   mapped("outcome", ren{"reason": "prose"}, nil, nil),
		"close": mapped("close", ren{"reason": "prose"}, nil,
			vals{"closure_class": dispositionRename}),
		// The era's proof carried its output INLINE beside its sha; the current record keeps
		// the bytes in the content-addressed proofs/ store, which the channel copy preserves,
		// and proof_sha is the tie. Carrying the bytes twice is the two-copies shape.
		"proof": mapped("proof", ren{"reason": "text", "sha256": "proof_sha"},
			drops{"output": "the output's bytes live in the copied proofs/ store under this proof's sha; a second inline copy is a fork waiting to disagree"}, nil),
		// line-of-inquiry became Avenue wholesale — same lifecycle, same fields, one rename.
		"line_of_inquiry": mapped("avenue", ren{"inquiry_id": "avenue_id"}, nil, nil),
		// inquiry-support became inquiry-review, and the schema's own history says why the
		// two fields die: the four-value `as` answered "does the report still carry this
		// line", and presence stopped being a question when the lines became GENERATED onto
		// the page (f516261a). The reason — the seat's actual reading — is the survivor.
		"inquiry_support": mapped("inquiry_review", nil, drops{
			"inquiry_id": "inquiry-review is a per-round statement about the whole live set; the per-line id answered the presence question the schema retired (f516261a)",
			"as":         "the closed set {supported, weakened, unsupported, absent} answered report-presence, which is generated now and cannot be cut — the schema deleted the flag rather than accept-and-discard it (f516261a)",
		}, nil),
		"motion":      {Translate: eraMotionEntry},
		"motion_rule": {Translate: eraMotionRuleEntry},
	}
}

type (
	ren   map[string]string
	drops map[string]string
	vals  map[string]map[string]string
)

// mapped is the workhorse era entry: rename old columns to their current names, drop the
// ones whose retirement the schema recorded (each with its reason, which lands in the
// refusal if the drop map is wrong), remap recorded value renames, then hold the result to
// the same identity contract every kept word gets.
func mapped(currentWord string, renames ren, dropped drops, valueMaps vals) Entry {
	return Entry{Translate: func(old OldEvent, _ record.Run) ([]proto.Message, error) {
		out := OldEvent{ID: old.ID, SeatID: old.SeatID, Round: old.Round, TS: old.TS, Key: old.Key,
			Word: currentWord, Lists: old.Lists, Arms: old.Arms}
		for col, v := range old.Fields {
			if _, gone := dropped[col]; gone {
				continue
			}
			name := col
			if to, ok := renames[col]; ok {
				name = to
			}
			if vm, ok := valueMaps[col]; ok {
				s, isStr := v.(string)
				if !isStr {
					return nil, fmt.Errorf("migrate: %s.%s holds %T where a word was expected", old.Word, col, v)
				}
				if to, ok := vm[s]; ok {
					v = to
				}
			}
			if out.Fields == nil {
				out.Fields = map[string]any{}
			}
			out.Fields[name] = v
		}
		body, err := identityBody(out)
		if err != nil {
			return nil, err
		}
		return []proto.Message{body}, nil
	}}
}

// eraMotionEntry: the era's motion was FLAT — subject, gap_id, dimension, proposed and the
// reason side by side — and the schema made the subject a oneof so a grade motion carrying a
// petition's field is unrepresentable. The era only ever filed grade motions; anything else
// refuses by name rather than guessing an arm.
func eraMotionEntry(old OldEvent, dst record.Run) ([]proto.Message, error) {
	// TWO ERAS SHARE THIS WORD. The SQLite era already decomposed the subject into oneof arm
	// tables — that shape is the CURRENT one and translates by identity; only the flat shard
	// shape (dimension/proposed beside the subject) needs this entry.
	if len(old.Arms) > 0 || old.Fields["dimension"] == nil && old.Fields["proposed"] == nil && old.Fields["gap_id"] == nil {
		return (Registry{}).Translate(old, dst)
	}
	if s, _ := old.Fields["subject"].(string); s != "grade" {
		return nil, fmt.Errorf("migrate: era motion subject %q has no authored arm mapping — the era's census held only grade motions, so this is a new fact to rule on, not a shape to guess", s)
	}
	m := &recordpb.Motion{Subject: recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum()}
	g := &recordpb.GradeMotion{}
	m.Filing = &recordpb.Motion_Grade{Grade: g}
	for col, v := range old.Fields {
		switch col {
		case "subject":
		case "motion_id":
			m.MotionId = proto.String(v.(string))
		case "reason":
			m.Basis = proto.String(v.(string))
		case "gap_id":
			g.GapId = proto.String(v.(string))
		case "dimension":
			num, err := enumNumberOf(recordpb.GradeDimension(0).Descriptor(), v.(string))
			if err != nil {
				return nil, fmt.Errorf("migrate: motion.dimension: %w", err)
			}
			g.Dimension = recordpb.GradeDimension(num).Enum()
		case "proposed":
			num, err := enumNumberOf(recordpb.Grade(0).Descriptor(), v.(string))
			if err != nil {
				return nil, fmt.Errorf("migrate: motion.proposed: %w", err)
			}
			g.Proposed = recordpb.Grade(num).Enum()
		default:
			return nil, fmt.Errorf("migrate: old column motion.%s has no place on the grade filing — extend the era motion entry", col)
		}
	}
	return []proto.Message{m}, nil
}

// eraMotionRuleEntry: the era's ruling was a WORD beside a subject; the schema keyed the
// verdict set on the subject as one oneof arm per subject. The era's `inquiry` subject is
// today's `direction` — the ruling on a line of inquiry, whose motion_id is the avenue's own
// id, which is exactly what the era wrote there.
func eraMotionRuleEntry(old OldEvent, dst record.Run) ([]proto.Message, error) {
	// Same two-era split as the motion: an arm-shaped ruling is the current decomposition
	// and translates by identity; only the flat `ruling`-word shape is this era's.
	if _, flat := old.Fields["ruling"]; !flat {
		return (Registry{}).Translate(old, dst)
	}
	r := &recordpb.MotionRule{}
	subject, _ := old.Fields["subject"].(string)
	ruling, _ := old.Fields["ruling"].(string)
	for col, v := range old.Fields {
		switch col {
		case "subject", "ruling":
		case "motion_id":
			r.MotionId = proto.String(v.(string))
		case "reason":
			r.Opinion = proto.String(v.(string))
		default:
			return nil, fmt.Errorf("migrate: old column motion_rule.%s has no place on the ruling — extend the era motion-rule entry", col)
		}
	}
	switch subject {
	case "grade":
		r.Subject = recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum()
		num, err := enumNumberOf(recordpb.GradeRuling(0).Descriptor(), ruling)
		if err != nil {
			return nil, fmt.Errorf("migrate: motion_rule.ruling (grade): %w", err)
		}
		r.Ruling = &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling(num)}
	case "inquiry":
		r.Subject = recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION.Enum()
		num, err := enumNumberOf(recordpb.DirectionRuling(0).Descriptor(), ruling)
		if err != nil {
			return nil, fmt.Errorf("migrate: motion_rule.ruling (inquiry->direction): %w", err)
		}
		r.Ruling = &recordpb.MotionRule_Direction{Direction: recordpb.DirectionRuling(num)}
	default:
		return nil, fmt.Errorf("migrate: era motion-rule subject %q has no authored mapping — the era's census held grade and inquiry, so this is a new fact to rule on", subject)
	}
	return []proto.Message{r}, nil
}
