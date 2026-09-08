package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Entries is the authored translation set for the vocabulary the current schema retired.
// Each entry is the statement of what an old word MEANS in the current model; the tests
// beside this file state each translation as data (old fields in, new event out).
func Entries() Registry {
	return Registry{
		"friction":      {Translate: frictionEntry},
		"friction_none": {Translate: frictionNoneEntry},
		"opinion":       {Translate: opinionEntry},
	}
}

// frictionEntry: the friction channel became the typed log (#755). Old columns
// text/kind/estopped_by carry over by name; `kind` was the old FrictionKind word and the
// LogType vocabulary kept those words. What the rename ADDED is `source` — the old channel
// was seat-filed except for the tool's own estoppel refusals, which is exactly the
// (TOOL, ESTOPPEL) pairing the new model spells out.
func frictionEntry(old OldEvent, _ record.Run) ([]proto.Message, error) {
	l := &recordpb.Log{Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
	for col, v := range old.Fields {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("migrate: friction.%s holds %T, not text", col, v)
		}
		switch col {
		case "text":
			l.Text = proto.String(s)
		case "estopped_by":
			l.EstoppedBy = proto.String(s)
		case "kind":
			num, err := enumNumberOf(recordpb.LogType(0).Descriptor(), s)
			if err != nil {
				return nil, fmt.Errorf("migrate: friction.kind: %w", err)
			}
			l.Type = recordpb.LogType(num).Enum()
		default:
			return nil, fmt.Errorf("migrate: old column friction.%s has no place on Log — extend the friction entry", col)
		}
	}
	// An untyped old entry predates the kind column. FRICTION is the channel's namesake and
	// the honest generic: "the work was impeded and you are noting it".
	if l.Type == nil {
		l.Type = recordpb.LogType_LOG_TYPE_FRICTION.Enum()
	}
	// An estoppel was never seat-filed: the tool recorded its own refusal. The new model
	// makes that pairing structural, so the translation says it too.
	if l.GetType() == recordpb.LogType_LOG_TYPE_ESTOPPEL {
		l.Source = recordpb.LogSource_LOG_SOURCE_TOOL.Enum()
	}
	return []proto.Message{l}, nil
}

// frictionNoneEntry: the explicit empty form became the POSITIVE nominal entry — "none" is
// an answer, and an entry that says so is still an entry.
func frictionNoneEntry(old OldEvent, _ record.Run) ([]proto.Message, error) {
	l := &recordpb.Log{
		Type:   recordpb.LogType_LOG_TYPE_NOMINAL.Enum(),
		Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum(),
	}
	for col, v := range old.Fields {
		s, ok := v.(string)
		if !ok || col != "text" {
			return nil, fmt.Errorf("migrate: old column friction_none.%s has no place on Log — extend the friction_none entry", col)
		}
		l.Text = proto.String(s)
	}
	return []proto.Message{l}, nil
}

// opinionEntry is the one CONCEPT translation: the bench opinion was retired (fd9e6970) for
// a docket MOTION naming the gap and a docket RULING carrying the bench's answer. One old
// event becomes two, both stamped with the opinion's own ts and seat — the bench raised the
// question and the bench answered it, which is what an opinion was.
func opinionEntry(old OldEvent, dst record.Run) ([]proto.Message, error) {
	id, err := record.MintMotionID(dst)
	if err != nil {
		return nil, err
	}
	docket := &recordpb.DocketMotion{}
	motion := &recordpb.Motion{
		MotionId: proto.String(id),
		Subject:  recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
		Filing:   &recordpb.Motion_Docket{Docket: docket},
	}
	ruling := &recordpb.DocketRuling{}
	rule := &recordpb.MotionRule{
		MotionId: proto.String(id),
		Subject:  recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
		Ruling:   &recordpb.MotionRule_Docket{Docket: ruling},
	}
	for col, v := range old.Fields {
		switch col {
		case "gap_id":
			docket.GapId = proto.String(v.(string))
		case "disposition":
			num, err := enumNumberOf(recordpb.Disposition(0).Descriptor(), v.(string))
			if err != nil {
				return nil, fmt.Errorf("migrate: opinion.disposition: %w", err)
			}
			ruling.Disposition = recordpb.Disposition(num).Enum()
		case "principle":
			ruling.Principle = proto.String(v.(string))
		case "tension":
			ruling.Tension = proto.String(v.(string))
		case "review_flag":
			ruling.ReviewFlag = proto.String(v.(string))
		case "settled":
			ruling.Settled = proto.String(v.(string))
		case "reopens_on":
			ruling.ReopensOn = proto.String(v.(string))
		case "final":
			n, ok := v.(int64)
			if !ok || (n != 0 && n != 1) {
				return nil, fmt.Errorf("migrate: opinion.final holds %v (%T), not a 0/1 bool", v, v)
			}
			if n == 1 {
				ruling.Final = proto.Bool(true)
			}
		case "rationale":
			// The old free-prose rationale is the ruling's own words — MotionRule.opinion,
			// which (the name is the history) is where ruling prose lives now.
			rule.Opinion = proto.String(v.(string))
		default:
			return nil, fmt.Errorf("migrate: old column opinion.%s has no place on the docket pair — extend the opinion entry", col)
		}
	}
	// The motion's ask, in the only words the record has for it: the bench took the gap up
	// on its own motion. Synthesized text says so rather than posing as a filing.
	motion.Basis = proto.String("migrated from a bench opinion: the bench took this gap up sua sponte")
	return []proto.Message{motion, rule}, nil
}
