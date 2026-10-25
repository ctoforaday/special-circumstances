package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"

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
		// A SITTING'S ENDS AND ITS LIMIT KEEP WHAT THE RECORD READS. A source before epoch 32 copied
		// the hook payload onto them; the copies are not carried.
		"sitting_open": mapped("sitting_open", nil, drops{
			"session_id": "the conversation id was a copy of the hook payload that no reader joins on",
			"prompt_id":  "the dispatch id was a copy of the hook payload that no reader joins on",
		}, nil),
		"sitting_close": mapped("sitting_close", nil, drops{
			"agent_type": "the configuration is the opening end's; the closing end's copy of the payload is not read",
			"session_id": "the conversation id was a copy of the hook payload that no reader joins on",
			"prompt_id":  "the dispatch id was a copy of the hook payload that no reader joins on",
		}, nil),
		"sitting_limit": mapped("sitting_limit", nil, drops{
			"agent_id":   "the agent is the input the seat was resolved from; the limit names the seat",
			"agent_type": "the configuration is the opening end's; the limit names the seat",
		}, nil),
		// The inquiry and its review, under the word an archive recorded them with: the columns
		// are the current ones but the proposal's text, which the record calls its question.
		"avenue":        inquiryEntry(ren{"avenue_id": "inquiry_id", "line": "question"}),
		"avenue_review": mapped("inquiry_review", nil, nil, nil),
		// A spot-check's `none` was the seat's claim that the archive was empty; the board answers
		// that, so the spot-check carries its ids, areas and reason and nothing else.
		"spot_check": mapped("spot_check", nil, drops{"none": "whether the archive was empty is the board's count, replayed at read time; the seat's claim of it is not stored"}, nil),
		// A REVISION TRANSLATES TO NO EVENT. It was a blue sitting's prose account of its change,
		// and the change is the sitting's edits, each with its reason; nothing reads a revision's
		// text. The manifest counts each in events_in and none in events_out.
		"revision": {Translate: toNoEvent},
		// A COINING TRANSLATES TO NO EVENT. A gap's class is a value of the closed GapClass set,
		// whose own definition is the class's; that a run coined a slug, and the coiner's words, are
		// not carried. Its mints carry the slug, which valueRenames maps onto the set.
		"class_new": {Translate: toNoEvent},
		// A MINT'S COINING COLUMNS ARE NOT CARRIED, for the same reason: the jsonl era's mint-carried
		// trio and the sqlite era's class_new flag both arrive here. Nor is a mint's distinct_from
		// list, which a source before epoch 45 may hold.
		"mint": mapped("mint", nil, drops{
			"class_new":     mintCoiningDrop,
			"definition":    mintCoiningDrop,
			"neighbor":      mintCoiningDrop,
			"distinguisher": mintCoiningDrop,
			"distinct_from": "the open gaps a lens named as different defects answered a word-overlap refusal of the mint, which the record does not make: whether a gap is a duplicate is the lens's reading of the board, and the claim is not carried",
		}, nil),
		// A CREDIT TRANSLATES TO NO EVENT. A source at epoch 42 to 44 may hold one: a lens adding its
		// raising finding to an open gap's found_by, which answered the contradiction that finding
		// raised. The record answers a contradiction with the gap minted from its raising finding and
		// with nothing else, so a credited contradiction comes forward unanswered and the credited
		// finding is credited by no gap. The manifest counts each in events_in and none in events_out.
		"credit": {Translate: toNoEvent},
		// A REPRODUCE BEFORE EPOCH 28 RE-RAN A SCRIPT AND JUDGED IT, keyed by the script's sha. Its
		// re-run is the script's; its soundness is carried to every anchor on that script when it was
		// written — the fan-out is the replay's, because it reads the proofs replayed before it
		// (reproduceAnchors in replay.go).
		"reproduce": {Translate: reproduceEntry},
		// A DISPATCH ROW BEFORE EPOCH 31 NAMED THE GAPS WHOSE REMAND EXCHANGE IT READIED. The record
		// stores the plan the row was written from, whose remand_owed names them with the direction,
		// and keeps no copy on the row; an archived row's list is not carried, and its plan was never
		// stored, so the gaps are read off the remand rulings the record replays.
		"dispatch": {Translate: dispatchEntry},
		// AN APPEAL HAS NO TRANSLATION. A ruling stands as made, and the record holds no act that
		// presses one; no archived run holds an appeal, so a record that does is a new fact for its
		// operator to accept by word, not a shape to guess.
		"motion_appeal": {Loss: "the record holds no appeal: a ruling stands as made, and a seat that disagrees with one answers it with a docket motion on the gap. The appeal's reason is not carried"},
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
// Every other type passes through by name. A log event is text, type and source; the translation
// is explicit about each column rather than inheriting the identity path, because a word this entry
// exists to DROP must not be able to reach that path by another column. A source before epoch 32
// may carry `estopped_by`, which is not carried: the estoppel count reads the type, and the entry's
// text names the gap.
func logEntry(old OldEvent) ([]proto.Message, error) {
	l := &recordpb.Log{}
	for col, v := range old.Fields {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("log.%s holds %T, not text", col, v)
		}
		switch col {
		case "text", "reason":
			l.Text = proto.String(s)
		case "estopped_by":
		case "source":
			num, err := enumNumberOf(recordpb.LogSource(0).Descriptor(), s)
			if err != nil {
				return nil, fmt.Errorf("log.source: %w", err)
			}
			l.Source = recordpb.LogSource(num).Enum()
		case "type":
			if s == "nominal" {
				return nil, nil
			}
			num, err := enumNumberOf(recordpb.LogType(0).Descriptor(), s)
			if err != nil {
				return nil, fmt.Errorf("log.type: %w", err)
			}
			l.Type = recordpb.LogType(num).Enum()
		default:
			return nil, fmt.Errorf("old column log.%s has no place on Log — extend the log entry", col)
		}
	}
	if l.Source == nil {
		l.Source = recordpb.LogSource_LOG_SOURCE_SEAT.Enum()
	}
	return []proto.Message{l}, nil
}

// frictionEntry: the friction channel became the typed log (#755). Old columns text/kind carry
// over by name, and estopped_by is not carried, as on a log; `kind` was the old FrictionKind word
// and the LogType vocabulary kept those words. What the rename ADDED is `source` — the old channel
// was seat-filed except for the tool's own estoppel refusals, which is exactly the
// (TOOL, ESTOPPEL) pairing the new model spells out.
func frictionEntry(old OldEvent) ([]proto.Message, error) {
	l := &recordpb.Log{Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
	for col, v := range old.Fields {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("friction.%s holds %T, not text", col, v)
		}
		switch col {
		case "text", "reason": // `reason` is the shard era's spelling of the same prose
			l.Text = proto.String(s)
		case "estopped_by":
		case "kind":
			num, err := enumNumberOf(recordpb.LogType(0).Descriptor(), s)
			if err != nil {
				return nil, fmt.Errorf("friction.kind: %w", err)
			}
			l.Type = recordpb.LogType(num).Enum()
		default:
			return nil, fmt.Errorf("old column friction.%s has no place on Log — extend the friction entry", col)
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

// citeEntry: a cite is kept word for word but `source_text_read`, a source before epoch 32's
// assertion of how much of the source the seat read, which nothing reads. And the one thing an old
// cite cannot say is where its text came from — no run before epoch 7 recorded it. A guessed `embedded` would tell red an OCR
// quote needs no page check; NOT_RECORDED says the question was never asked. Filled only where
// absent, so a cite that carries an origin keeps it.
//
// A SOURCE BEFORE EPOCH 37 named an OCR citation's span `ocr_quote`; it is the citation's
// source_span, carried word for word. What no archived cite holds is the span's offsets or the
// source's text around it — the cache they were read from is not archived — so a migrated cite
// carries a span only where an OCR quote stood, and never the text a lens reads beside it.
func citeEntry(old OldEvent) ([]proto.Message, error) {
	fields := map[string]any{}
	for k, v := range old.Fields {
		switch k {
		case "source_text_read":
		case "ocr_quote":
			fields["source_span"] = v
		default:
			fields[k] = v
		}
	}
	old.Fields = fields
	body, err := identityBody(old)
	if err != nil {
		return nil, err
	}
	c, ok := body.(*recordpb.Cite)
	if !ok {
		return nil, fmt.Errorf("cite translated to %T, not a Cite", body)
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
	// NOR, BEFORE EPOCH 15, WHICH PART OF THE WORK THE COPY WAS. NOT_ASKED would claim the fetch
	// looked and could not tell; the field simply did not exist.
	if c.SourceCompleteness == nil {
		c.SourceCompleteness = recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_RECORDED.Enum()
	}
	// NOR, BEFORE EPOCH 40, WHOSE BYTES THE CITATION READ. An archived cite may rest on a followed
	// pointer, a recovered copy or a fuller one, and nothing it holds says which — the fetch index
	// that knew is not archived. NONE would claim the typed url's own bytes; NOT_RECORDED says the
	// question was not asked, and no copy_url is written: the residue is that a migrated cite names
	// no copy.
	if c.CopySubstitution == nil {
		c.CopySubstitution = recordpb.CopySubstitution_COPY_SUBSTITUTION_NOT_RECORDED.Enum()
	}
	return []proto.Message{c}, nil
}

// dispatchEntry: a dispatch row is kept word for word but its remand_owed list.
func dispatchEntry(old OldEvent) ([]proto.Message, error) {
	if _, ok := old.Lists["remand_owed"]; ok {
		lists := map[string][]string{}
		for k, v := range old.Lists {
			if k != "remand_owed" {
				lists[k] = v
			}
		}
		old.Lists = lists
	}
	body, err := identityBody(old)
	if err != nil {
		return nil, err
	}
	return []proto.Message{body}, nil
}

// registerEntry: a register is kept word for word and opens a sitting of its own. A source at epoch
// 8 to 26 may hold a register naming the sitting it repaired (`repairs_sitting`); the record holds
// no repair, so the register is written without it and opens a sitting of its own, as every
// register does. Every act filed under it is kept, in that sitting rather than the one it repaired
// — the residue the plan states (plans/v2/10 §II, item 8). No archived run holds one.
func registerEntry(old OldEvent) ([]proto.Message, error) {
	if _, ok := old.Fields["repairs_sitting"]; ok {
		fields := map[string]any{}
		for k, v := range old.Fields {
			if k != "repairs_sitting" {
				fields[k] = v
			}
		}
		old.Fields = fields
	}
	body, err := identityBody(old)
	if err != nil {
		return nil, err
	}
	r, ok := body.(*recordpb.Register)
	if !ok {
		return nil, fmt.Errorf("register translated to %T, not a Register", body)
	}
	return []proto.Message{r}, nil
}

// remandDirection carries a remand's research direction to the field that holds it. A source at
// epoch 26 or earlier states it as `reopens_on`, its one column for what follows a ruling; a ruling
// that closes the gap keeps its `reopens_on`, which is what would reopen it. The disposition says
// which the source meant.
func remandDirection(r *recordpb.DocketRuling) {
	if r.GetDisposition() != recordpb.Disposition_DISPOSITION_REMANDED || r.ReopensOn == nil {
		return
	}
	r.Direction, r.ReopensOn = r.ReopensOn, nil
}

// toNoEvent translates an archived event to none, whatever its era's columns.
func toNoEvent(OldEvent) ([]proto.Message, error) { return nil, nil }

// mintCoiningDrop is why a mint's coining columns are not carried.
const mintCoiningDrop = "a class is a value of the closed set; whether a run coined it, and the coiner's definition, are not carried — the value's own definition is"

// reproduceEntry: `reproduced`, a bool, becomes `rerun` — the act re-ran the script, and it
// reproduced or it differed. A record written at epoch 28 or later states `rerun` and `anchor`
// itself, and passes by name. `reason` is the shard era's spelling of `note`.
func reproduceEntry(old OldEvent) ([]proto.Message, error) {
	out := OldEvent{ID: old.ID, SeatID: old.SeatID, Round: old.Round, TS: old.TS, Key: old.Key,
		Word: "reproduce", Lists: old.Lists, Arms: old.Arms, Fields: map[string]any{}}
	for col, v := range old.Fields {
		switch col {
		case "reason":
			out.Fields["note"] = v
		case "reproduced":
			var ok bool
			switch b := v.(type) {
			case int64:
				ok = b == 1
				if b != 0 && b != 1 {
					return nil, fmt.Errorf("reproduce.reproduced holds %v, not a 0/1 bool", v)
				}
			case bool:
				ok = b
			default:
				return nil, fmt.Errorf("reproduce.reproduced holds %v (%T), not a 0/1 bool", v, v)
			}
			r := recordpb.Rerun_RERUN_DIFFERS
			if ok {
				r = recordpb.Rerun_RERUN_REPRODUCED
			}
			out.Fields["rerun"] = recordpb.Word(r)
		default:
			out.Fields[col] = v
		}
	}
	body, err := identityBody(out)
	if err != nil {
		return nil, err
	}
	return []proto.Message{body}, nil
}

// frictionNoneEntry: the explicit empty form carried "none" as an entry. The type that said so is
// retired and clean is derived from a bracketed sitting that filed nothing, so the row translates
// to no event at all — the statement survives, computed rather than stored.
func frictionNoneEntry(old OldEvent) ([]proto.Message, error) {
	for col := range old.Fields {
		if col != "text" && col != "reason" { // `reason` is the shard era's spelling
			return nil, fmt.Errorf("old column friction_none.%s has no place on Log — extend the friction_none entry", col)
		}
	}
	return nil, nil
}

// opinionEntry is the one CONCEPT translation: the bench opinion was retired (fd9e6970) for
// a docket MOTION naming the gap and a docket RULING carrying the bench's answer. One old
// event becomes two, both stamped with the opinion's own ts and seat — the bench raised the
// question and the bench answered it, which is what an opinion was.
func opinionEntry(old OldEvent) ([]proto.Message, error) {
	// UNIQUE BY CONSTRUCTION. The pair is named for the opinion's own event id, which no other
	// event shares and no motion the archive filed is spelled like; the remap then gives that
	// name its id in the one shape, as it does every archived id.
	id := fmt.Sprintf("M-mig-%d", old.ID)
	docket := &recordpb.DocketMotion{}
	motion := &recordpb.Motion{
		MotionId: proto.String(id),
		Filing:   &recordpb.Motion_Docket{Docket: docket},
	}
	ruling := &recordpb.DocketRuling{}
	rule := &recordpb.MotionRule{
		MotionId: proto.String(id),
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
				return nil, fmt.Errorf("opinion.disposition: %w", err)
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
				return nil, fmt.Errorf("opinion.final holds %v (%T), not a 0/1 bool", v, v)
			}
			if n == 1 {
				ruling.Final = proto.Bool(true)
			}
		case "rationale", "reason": // `reason` is the shard era's spelling of the rationale
			// The old free-prose rationale is the ruling's own words — MotionRule.opinion,
			// which (the name is the history) is where ruling prose lives now.
			rule.Opinion = proto.String(v.(string))
		default:
			return nil, fmt.Errorf("old column opinion.%s has no place on the docket pair — extend the opinion entry", col)
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
	if ruling.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED && ruling.ReopensOn == nil {
		ruling.ReopensOn = proto.String("not stated: this remand predates the direction requirement (migrated from a bench opinion)")
	}
	if ruling.ReopensOn == nil && ruling.Final == nil {
		ruling.ReopensOn = proto.String("not stated: this ruling predates the reopens-on/final requirement (migrated from a bench opinion)")
	}
	remandDirection(ruling)
	// The motion's ask, in the only words the record has for it: the bench took the gap up
	// on its own motion. Synthesized text says so rather than posing as a filing.
	motion.Basis = proto.String("migrated from a bench opinion: the bench took this gap up sua sponte")
	return []proto.Message{motion, rule}, nil
}
