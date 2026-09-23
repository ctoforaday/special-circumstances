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
	reg := Registry{
		"friction":      {Translate: frictionEntry},
		"log":           {Translate: logEntry},
		"friction_none": {Translate: frictionNoneEntry},
		"opinion":       {Translate: opinionEntry},
		"cite":          {Translate: citeEntry},
		"register":      {Translate: registerEntry},
	}
	for w, e := range eraEntries() {
		reg[w] = e
	}
	return reg
}

// logEntry: the nominal type is retired, and an archived entry that asserted a clean sitting
// carries forward as an ABSENCE. Clean is derived from a bracketed sitting that filed nothing, so
// keeping the row would restate as data what the record already computes — and the word it carries
// is no longer in the vocabulary, which is why the identity default refuses it.
//
// Every other type passes through by name. A log event is text, type, source and estopped_by; the
// translation is explicit about all four rather than inheriting the identity path, because a word
// this entry exists to DROP must not be able to reach that path by another column.
func logEntry(old OldEvent, _ record.Run) ([]proto.Message, error) {
	l := &recordpb.Log{}
	for col, v := range old.Fields {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("migrate: log.%s holds %T, not text", col, v)
		}
		switch col {
		case "text", "reason":
			l.Text = proto.String(s)
		case "estopped_by":
			l.EstoppedBy = proto.String(s)
		case "source":
			num, err := enumNumberOf(recordpb.LogSource(0).Descriptor(), s)
			if err != nil {
				return nil, fmt.Errorf("migrate: log.source: %w", err)
			}
			l.Source = recordpb.LogSource(num).Enum()
		case "type":
			if s == "nominal" {
				return nil, nil
			}
			num, err := enumNumberOf(recordpb.LogType(0).Descriptor(), s)
			if err != nil {
				return nil, fmt.Errorf("migrate: log.type: %w", err)
			}
			l.Type = recordpb.LogType(num).Enum()
		default:
			return nil, fmt.Errorf("migrate: old column log.%s has no place on Log — extend the log entry", col)
		}
	}
	if l.Source == nil {
		l.Source = recordpb.LogSource_LOG_SOURCE_SEAT.Enum()
	}
	return []proto.Message{l}, nil
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
		case "text", "reason": // `reason` is the shard era's spelling of the same prose
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

// citeEntry: a cite is kept word for word, and the one thing an old cite cannot say is where its
// text came from — no run before epoch 7 recorded it. A guessed `embedded` would tell red an OCR
// quote needs no page check; NOT_RECORDED says the question was never asked. Filled only where
// absent, so a cite that carries an origin keeps it.
func citeEntry(old OldEvent, _ record.Run) ([]proto.Message, error) {
	body, err := identityBody(old)
	if err != nil {
		return nil, err
	}
	c, ok := body.(*recordpb.Cite)
	if !ok {
		return nil, fmt.Errorf("migrate: cite translated to %T, not a Cite", body)
	}
	if c.SourceTextOrigin == nil {
		c.SourceTextOrigin = recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NOT_RECORDED.Enum()
	}
	// AND NO RUN BEFORE EPOCH 10 ASKED AN INDEX WHETHER THE WORK STILL STANDS. NOT_CHECKED would
	// be the answer for a live cite whose source carries no doi; here the truer word is that the
	// field did not exist when this was written, which is what NOT_RECORDED says. Filled only
	// where absent, so a cite that carries a status keeps it.
	if c.WorkStatus == nil {
		c.WorkStatus = recordpb.WorkStatus_WORK_STATUS_NOT_RECORDED.Enum()
	}
	return []proto.Message{c}, nil
}

// registerEntry: a register is kept word for word and opens a sitting of its own. An epoch-7
// register table holds event_id, tool_version, hook_version, agent_id, run_via and agent_type, so
// no archived register can say it repairs a sitting; inferring a repair from the stream would put a
// claim on the record no seat made and no writer checked.
//
// A ROW THAT CARRIES THE COLUMN IS REFUSED, NOT REWRITTEN. Dropping the value silently is the move
// that reads identically whether the source was an honest epoch-7 run or a record claiming a field
// its epoch could not hold — and a translation that cannot fail on any real row is a comment. The
// refusal names what it found, which is how this repository turns away data by its content.
func registerEntry(old OldEvent, _ record.Run) ([]proto.Message, error) {
	if v, ok := old.Fields["repairs_sitting"]; ok {
		return nil, fmt.Errorf("migrate: register %s carries repairs_sitting (%v), which no run before epoch 8 could record — "+
			"this row did not come from the epoch this step reads, and translating it would put a repair on the record no writer checked", old.Key, v)
	}
	body, err := identityBody(old)
	if err != nil {
		return nil, err
	}
	r, ok := body.(*recordpb.Register)
	if !ok {
		return nil, fmt.Errorf("migrate: register translated to %T, not a Register", body)
	}
	return []proto.Message{r}, nil
}

// frictionNoneEntry: the explicit empty form carried "none" as an entry. The type that said so is
// retired and clean is derived from a bracketed sitting that filed nothing, so the row translates
// to no event at all — the statement survives, computed rather than stored.
func frictionNoneEntry(old OldEvent, _ record.Run) ([]proto.Message, error) {
	for col := range old.Fields {
		if col != "text" && col != "reason" { // `reason` is the shard era's spelling
			return nil, fmt.Errorf("migrate: old column friction_none.%s has no place on Log — extend the friction_none entry", col)
		}
	}
	return nil, nil
}

// opinionEntry is the one CONCEPT translation: the bench opinion was retired (fd9e6970) for
// a docket MOTION naming the gap and a docket RULING carrying the bench's answer. One old
// event becomes two, both stamped with the opinion's own ts and seat — the bench raised the
// question and the bench answered it, which is what an opinion was.
func opinionEntry(old OldEvent, _ record.Run) ([]proto.Message, error) {
	// UNIQUE BY CONSTRUCTION, AND VISIBLY A MIGRATION. MintMotionID counts the motions
	// already replayed, and the first drive of the real archive collided exactly there: a
	// synthesized M6 met the old run's own M6 five events later. Old event ids are unique
	// and the M-mig- spelling is disjoint from the native M%d family, so the pair's id says
	// what its basis text says — this exchange was synthesized from an opinion.
	id := fmt.Sprintf("M-mig-%d", old.ID)
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
			w := v.(string)
			if to, ok := dispositionRename[w]; ok { // the a1e8e260 rename, for shard-era opinions
				w = to
			}
			num, err := enumValue(recordpb.Disposition(0).Descriptor(), w)
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
		case "rationale", "reason": // `reason` is the shard era's spelling of the rationale
			// The old free-prose rationale is the ruling's own words — MotionRule.opinion,
			// which (the name is the history) is where ruling prose lives now.
			rule.Opinion = proto.String(v.(string))
		default:
			return nil, fmt.Errorf("migrate: old column opinion.%s has no place on the docket pair — extend the opinion entry", col)
		}
	}
	// THE STATED MISS, for rulings that predate a requirement. The earliest opinions carry
	// neither `settled` nor the reopens_on/final pair — the era had not asked yet. An empty
	// string would read as "answered: nothing" and a forged `final` would decide what nobody
	// decided; a value that SAYS it was never recorded is the one honest shape
	// ([[facts-are-fields]]: make the miss loud, or a stated "not measured"). Filled only
	// where absent — a ruling that answered keeps its answer.
	if ruling.Settled == nil {
		ruling.Settled = proto.String("not stated: this ruling predates the settled requirement (migrated from a bench opinion)")
	}
	if ruling.Tension == nil {
		ruling.Tension = proto.String("not stated: this ruling predates the tension requirement (migrated from a bench opinion)")
	}
	if ruling.ReviewFlag == nil {
		ruling.ReviewFlag = proto.String("not stated: this ruling predates the review-flag requirement (migrated from a bench opinion)")
	}
	if ruling.ReopensOn == nil && ruling.Final == nil {
		ruling.ReopensOn = proto.String("not stated: this ruling predates the reopens-on/final requirement (migrated from a bench opinion)")
	}
	// The motion's ask, in the only words the record has for it: the bench took the gap up
	// on its own motion. Synthesized text says so rather than posing as a filing.
	motion.Basis = proto.String("migrated from a bench opinion: the bench took this gap up sua sponte")
	return []proto.Message{motion, rule}, nil
}
