package record

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Correct is a same-sitting correction the WRITER asked for (plans/same-sitting-correction.md).
//
// The seat re-runs the act's own verb with the text it meant, naming the act it corrects. The verb
// builds its body exactly as it always does; the one body of Type the invocation writes becomes the
// REPLACEMENT, and Append routes it to appendCorrected, which writes it and a Correction event in
// one transaction. Every other body the invocation writes (a tool log a verb emits on the side)
// appends normally.
type Correct struct {
	Type recordpb.EventType // the type the correcting verb writes
	Key  string             // the key of the act being corrected
	Why  string             // what was wrong with it, in the seat's words

	// written is set once the replacement is on the record: a second body of Type in one
	// correcting invocation would be a second act with no key to name it by.
	written bool
}

// CorrectionKeyPrefix is the key segment every Correction event carries: `<seat>:correction:<K>`.
// A retry of the same correction collides on it, which is what makes the retry idempotent.
func correctionKey(seatID, corrects string) string {
	return seatID + ":" + recordpb.Word(recordpb.EventType_EVENT_TYPE_CORRECTION) + ":" + corrects
}

// correctionTarget is the act a correction names, read once.
type correctionTarget struct {
	ID     int64
	Key    string
	SeatID string
	Type   recordpb.EventType
	Body   proto.Message
}

// readTarget reads the act `key` names. Absent is an ERROR: a correction of nothing is a seat that
// mistyped a key, and saying "nothing to correct" would read like success.
func readTarget(db *sql.DB, key string) (*correctionTarget, error) {
	ev, id, found, err := recordsql.EventByKey(db, key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, feov.Errorf(feov.Validation,
			"record: --corrects names %q, and no act on this record carries that key — the key is printed on the act's own success line as [key …]", key)
	}
	body, ok := recordpb.Body(ev)
	if !ok {
		return nil, fmt.Errorf("record: event %d (key %q) carries no body", id, key)
	}
	return &correctionTarget{ID: id, Key: key, SeatID: ev.GetSeatId(), Type: ev.GetType(), Body: body}, nil
}

// TargetBody is the body of the act a correcting verb names, for the handlers that must RE-STATE a
// value the act already holds rather than re-derive it (F13): a line of inquiry's id, a proof's
// reproduced outputs, an outcome's derived verdict. It is loud on every miss — absent, another
// seat's, another type — because a handler that silently fell back to deriving would write a
// replacement the frozen compare then refuses, for a reason the seat cannot see.
func TargetBody(run Run, seatID, key string, typ recordpb.EventType) (proto.Message, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, feov.Errorf(feov.Validation, "record: --corrects names %q, and this run has recorded nothing", key)
	}
	t, err := readTarget(db, key)
	if err != nil {
		return nil, err
	}
	if err := requireSameSeatAndType(t, seatID, typ); err != nil {
		return nil, err
	}
	return t.Body, nil
}

func requireSameSeatAndType(t *correctionTarget, seatID string, typ recordpb.EventType) error {
	if t.Type != typ {
		return feov.Errorf(feov.Validation,
			"record: %s is a %s, and this command writes a %s — a correction re-runs the corrected act's OWN verb, so name a key this command wrote",
			t.Key, recordpb.Word(t.Type), recordpb.Word(typ))
	}
	if t.SeatID != seatID {
		return feov.Errorf(feov.Validation,
			"record: %s was written by %s, not by %s — only the seat that wrote an act may correct it. To answer another seat's act, say so in an act of your own",
			t.Key, t.SeatID, seatID)
	}
	return nil
}

// tierOf is how much of THIS act a correction may change: the type's declared tier, with the ONE
// exception that depends on the body — a log entry the TOOL wrote records what the tool did, and no
// seat may restate it. This is the only home of that exception.
func tierOf(typ recordpb.EventType, body proto.Message) recordpb.CorrectionTier {
	if l, ok := body.(*recordpb.Log); ok && l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL {
		return recordpb.CorrectionTier_CORRECTION_TIER_NONE
	}
	return recordpb.Tier(typ)
}

// Correctable answers whether an act of this type can be corrected at all, for the surfaces that
// must decide whether to offer the correction (the key-collision refusal's pointer).
func Correctable(typ recordpb.EventType, body proto.Message) bool {
	return tierOf(typ, body) != recordpb.CorrectionTier_CORRECTION_TIER_NONE
}

// validateCorrection is what may change between an act and its replacement: nothing at all for a
// NONE act, only the declared prose for a PROSE act, everything but the key label for a FULL one —
// and never nothing, because a replacement equal to its target corrects nothing.
func validateCorrection(t *correctionTarget, seatID string, typ recordpb.EventType, body proto.Message) error {
	if err := requireSameSeatAndType(t, seatID, typ); err != nil {
		return err
	}
	tier := tierOf(t.Type, t.Body)
	if tier == recordpb.CorrectionTier_CORRECTION_TIER_NONE {
		why := "it creates an identity other acts refer to, decides a fate no restatement may move, or records what the tool or the harness did"
		if l, ok := t.Body.(*recordpb.Log); ok && l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL {
			why = "the tool wrote it, as its own record of what it did, and a seat cannot restate that"
		}
		return feov.Errorf(feov.Validation,
			"record: %s is a %s, which cannot be corrected — %s. Say it in the act that supersedes it; the record is append-only and both stay visible",
			t.Key, recordpb.Word(t.Type), why)
	}
	if a, b := keyLabel(t.Body), keyLabel(body); a != b {
		return feov.Errorf(feov.Validation,
			"record: a correction keeps the act's label — %s is on %q and this replacement is on %q. An act about something else is a new act, not a correction",
			t.Key, a, b)
	}
	if proto.Equal(t.Body, body) {
		return feov.Errorf(feov.Validation,
			"record: this correction changes nothing — the replacement equals event %d; correct it with the text you meant", t.ID)
	}
	if tier == recordpb.CorrectionTier_CORRECTION_TIER_PROSE {
		if diff := frozenDiff(t.Body, body); len(diff) > 0 {
			return feov.Errorf(feov.Validation,
				"record: a %s correction may change only the seat's own wording (%s); this one changes %s, which must stay as event %d recorded it. What the act decided is not restated by a correction — say it in a new act",
				recordpb.Word(t.Type), strings.Join(proseFlags(t.Body.ProtoReflect().Descriptor()), ", "), strings.Join(diff, ", "), t.ID)
		}
	}
	return nil
}

// frozenDiff names, in the words a seat types, every non-prose field that differs between an act
// and its replacement — recursing into message fields and oneof arms, so a docket ruling's
// disposition is found under MotionRule.ruling.docket.
func frozenDiff(a, b proto.Message) []string {
	seen := map[string]bool{}
	var out []string
	var walk func(x, y protoreflect.Message)
	walk = func(x, y protoreflect.Message) {
		fds := x.Descriptor().Fields()
		for i := 0; i < fds.Len(); i++ {
			fd := fds.Get(i)
			if p, _ := recordpb.IsProse(fd); p {
				continue
			}
			hx, hy := x.Has(fd), y.Has(fd)
			if fd.Message() != nil && !fd.IsList() && !fd.IsMap() && hx && hy {
				walk(x.Get(fd).Message(), y.Get(fd).Message())
				continue
			}
			if hx == hy && (!hx || x.Get(fd).Equal(y.Get(fd))) {
				continue
			}
			name := "--" + recordpb.FlagFor(fd)
			if fd.Message() != nil && !fd.IsList() {
				name = string(fd.Name())
			}
			if !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	walk(a.ProtoReflect(), b.ProtoReflect())
	sort.Strings(out)
	return out
}

// proseFlags names the flags that fill a body's prose fields, recursively.
func proseFlags(md protoreflect.MessageDescriptor) []string {
	seen := map[string]bool{}
	var out []string
	var walk func(md protoreflect.MessageDescriptor, depth int)
	walk = func(md protoreflect.MessageDescriptor, depth int) {
		if depth > 4 {
			return
		}
		fds := md.Fields()
		for i := 0; i < fds.Len(); i++ {
			fd := fds.Get(i)
			if fd.Message() != nil {
				walk(fd.Message(), depth+1)
				continue
			}
			if p, _ := recordpb.IsProse(fd); p {
				f := "--" + recordpb.FlagFor(fd)
				if !seen[f] {
					seen[f] = true
					out = append(out, f)
				}
			}
		}
	}
	walk(md, 0)
	sort.Strings(out)
	return out
}

// ProseFlags is proseFlags for an event type, for the generated help paragraph.
func ProseFlags(typ recordpb.EventType) []string {
	md, ok := bodyDescriptor(recordpb.Word(typ))
	if !ok {
		return nil
	}
	return proseFlags(md)
}

// supersedingAct is what a seat does instead when a correction is refused for time — the act that
// answers an earlier one without restating it.
func supersedingAct(typ recordpb.EventType) string {
	switch typ {
	case recordpb.EventType_EVENT_TYPE_MOTION_RULE:
		return "an appeal, or a new motion"
	case recordpb.EventType_EVENT_TYPE_MOTION_APPEAL:
		return "a new motion"
	case recordpb.EventType_EVENT_TYPE_REGRADE:
		return "a new regrade"
	case recordpb.EventType_EVENT_TYPE_CLOSE:
		return "a motion on the gap"
	case recordpb.EventType_EVENT_TYPE_AVENUE:
		return "a move of the line"
	}
	return "a new act"
}

// appendCorrected writes a replacement and its Correction in ONE transaction.
//
// # What is checked where, and why the order
//
// The retry is recognised FIRST, before anything that could have changed since: a crash between a
// correction's commit and the seat seeing it must re-answer the same, even if another seat has acted
// in between. Then what cannot change — the target's type, seat, tier and fields — is checked
// against the target read once. Then the record-level guards run against the target (a closure's
// correction names a closed gap by definition). Finally, INSIDE the transaction whose BEGIN took
// the write lock, what can change: the sitting, whether another seat has acted since (F7), and the
// chain the replacement's key extends. A reliance check made before the transaction would race a
// lens writing in parallel.
func appendCorrected(id Identity, db *sql.DB, ev *Event, typ recordpb.EventType, body proto.Message) (*Event, error) {
	c := id.Correct
	run, seatID := id.Run, id.SeatID
	if strings.TrimSpace(c.Why) == "" {
		return nil, feov.Errorf(feov.MissingField,
			"record: a correction requires --correction-why — what was wrong with the act; a reader sees it beside the struck text")
	}
	if existing, err := correctionRetry(db, seatID, c.Key, body); existing != nil || err != nil {
		return existing, err
	}
	target, err := readTarget(db, c.Key)
	if err != nil {
		return nil, err
	}
	if err := validateCorrection(target, seatID, typ, body); err != nil {
		return nil, err
	}
	if err := validateAgainst(run, seatID, typ, body, target); err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := recordsql.RequireTable(tx, recordpb.Word(recordpb.EventType_EVENT_TYPE_CORRECTION)); err != nil {
		return nil, err
	}
	// THE RETRY AGAIN, under the lock: two invocations of one correction racing each other must
	// resolve to one write and one idempotent answer, not two refusals or two replacements.
	var prior string
	switch err := tx.QueryRow(`SELECT c."replacement" FROM "correction" c JOIN "events" e ON e."id" = c."event_id" WHERE e."key" = ?`,
		correctionKey(seatID, c.Key)).Scan(&prior); {
	case err == nil:
		_ = tx.Rollback()
		return correctionRetry(db, seatID, c.Key, body)
	case !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}
	// F6: THE SAME SITTING. The target's sitting is the writer's registers before it; the writer's
	// current sitting is all of them — counted the way deriveKey counts.
	var before, now int
	if err := tx.QueryRow(`SELECT
	    (SELECT count(*) FROM "events" WHERE "seat_id" = ? AND "type" = 'register' AND "id" < ?),
	    (SELECT count(*) FROM "events" WHERE "seat_id" = ? AND "type" = 'register')`,
		seatID, target.ID, seatID).Scan(&before, &now); err != nil {
		return nil, fmt.Errorf("record: counting %s's sittings: %w", seatID, err)
	}
	if before != now {
		return nil, feov.Errorf(feov.Validation,
			"record: %s was written in an earlier sitting of %s (sitting %d; this is sitting %d) — a correction is for the sitting that wrote the act, while it is still yours alone. Say it in %s; the record keeps both",
			target.Key, seatID, before, now, supersedingAct(target.Type))
	}
	// F7: NOT ONCE RELIED ON. Any act by another seat after the target counts, whatever it was —
	// it may have read what this would change. The harness's span events and the tool's own log
	// entries are not a seat acting.
	var fid int64
	var fseat, ftype string
	switch err := tx.QueryRow(`SELECT e."id", e."seat_id", e."type" FROM "events" e
	      LEFT JOIN "log" l ON l."event_id" = e."id"
	     WHERE e."id" > ? AND e."seat_id" NOT IN (?, ?)
	       AND (l."source" IS NULL OR l."source" <> ?)
	     ORDER BY e."id" LIMIT 1`,
		target.ID, seatID, HarnessSeat, recordpb.Word(recordpb.LogSource_LOG_SOURCE_TOOL)).Scan(&fid, &fseat, &ftype); {
	case err == nil:
		return nil, feov.Errorf(feov.Validation,
			"record: another seat has acted since this %s (event %d, a %s by %s); what you would correct is now on the record they read — say it in a new act",
			recordpb.Word(target.Type), fid, ftype, fseat)
	case !errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("record: asking whether another seat has acted since %s: %w", target.Key, err)
	}
	// THE CHAIN: the replacement's key is the chain's ROOT key and its depth, both walked here from
	// the correction rows — never parsed out of a key.
	root, depth := target.Key, 0
	for {
		var prev string
		err := tx.QueryRow(`SELECT "corrects" FROM "correction" WHERE "replacement" = ?`, root).Scan(&prev)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("record: walking the correction chain of %s: %w", target.Key, err)
		}
		root, depth = prev, depth+1
	}
	ts := Now().UTC().Format(stampLayout)
	envelope(ev, ts, seatID, fmt.Sprintf("%s~%d", root, depth+1))
	if _, err := recordsql.InsertTx(tx, ev); err != nil {
		return nil, err
	}
	corr := &Event{}
	if _, err := recordpb.SetBody(corr, &recordpb.Correction{
		Corrects:    proto.String(target.Key),
		Replacement: proto.String(ev.GetKey()),
		Why:         proto.String(c.Why),
	}); err != nil {
		return nil, err
	}
	envelope(corr, ts, seatID, correctionKey(seatID, target.Key))
	if _, err := recordsql.InsertTx(tx, corr); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if id.OnWrite != nil {
		id.OnWrite(ev)
		id.OnWrite(corr)
	}
	return ev, nil
}

// closedByTarget answers whether the act a correction replaces is the one that closed this gap —
// the case where "the gap is already closed" is the correction's own premise, not a refusal. A read
// that fails answers false, so the ordinary guard runs and says what it says.
func closedByTarget(run Run, gapID string, target *correctionTarget) bool {
	if target == nil || gapID == "" {
		return false
	}
	var seq sql.NullInt64
	found, err := queryRow(run, []any{&seq}, `SELECT "merge_closed_seq" FROM "gap" WHERE "gap_id" = ?`, gapID)
	return err == nil && found && seq.Valid && seq.Int64 == target.ID
}

// correctionRetry answers a correction this seat already made of this key: the SAME replacement is
// the idempotent success (0 new events, the stored replacement returned so its key is shown), a
// different one is refused naming the act that stands now. (nil, nil) means no prior correction.
func correctionRetry(db *sql.DB, seatID, key string, body proto.Message) (*Event, error) {
	has, err := recordsql.HasTable(db, recordpb.Word(recordpb.EventType_EVENT_TYPE_CORRECTION))
	if err != nil || !has {
		return nil, err
	}
	var replacement string
	switch err := db.QueryRow(`SELECT c."replacement" FROM "correction" c JOIN "events" e ON e."id" = c."event_id" WHERE e."key" = ?`,
		correctionKey(seatID, key)).Scan(&replacement); {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, err
	}
	stored, _, found, err := recordsql.EventByKey(db, replacement)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("record: %s names replacement %s, which is not on the record", correctionKey(seatID, key), replacement)
	}
	if sb, ok := recordpb.Body(stored); ok && proto.Equal(sb, body) {
		return stored, nil
	}
	head, err := liveHead(db, key)
	if err != nil {
		return nil, err
	}
	return nil, feov.Errorf(feov.Validation,
		"record: %s was already corrected — the act that stands now is %s; name %s to correct again", key, head, head)
}

// liveHead follows a key's correction chain to the act that stands now. A key nobody corrected is
// its own head.
func liveHead(q interface {
	QueryRow(string, ...any) *sql.Row
}, key string) (string, error) {
	head := key
	for i := 0; i < 1<<16; i++ {
		var next string
		err := q.QueryRow(`SELECT "replacement" FROM "correction" WHERE "corrects" = ?`, head).Scan(&next)
		if errors.Is(err, sql.ErrNoRows) {
			return head, nil
		}
		if err != nil {
			return "", err
		}
		head = next
	}
	return "", fmt.Errorf("record: the correction chain of %s does not end", key)
}

// correctionPointer is the tail a key-collision refusal gains for a correctable act: the key of
// the act that stands now, and the invocation that corrects it. Fail-safe to the literal key — a
// chain that cannot be read (an older record has no correction table) points at the key the
// collision named, and a correction of it answers loudly if it was already corrected.
func correctionPointer(q interface {
	QueryRow(string, ...any) *sql.Row
}, key string) string {
	head, err := liveHead(q, key)
	if err != nil || head == "" {
		head = key
	}
	return fmt.Sprintf("If it was wrong, correct it now: run the same command with --corrects %s --correction-why <what was wrong>. The corrected act stays on the record, struck, beside the replacement", head)
}

// Struck is one corrected act: the act that replaced it, who corrected it and why.
type Struck struct {
	Replacement string `json:"replacement"`
	By          string `json:"by"`
	Why         string `json:"why"`
}

// StruckIndex is every correction on a stream, keyed by the corrected act's key — built once from
// the events a reader already holds.
type StruckIndex struct {
	byKey       map[string]Struck
	replacement map[string]bool
}

// StruckIndexOf indexes the Correction events on a stream.
func StruckIndexOf(evs []*Event) StruckIndex {
	idx := StruckIndex{}
	for _, e := range evs {
		c, ok := recordpb.BodyAs[*recordpb.Correction](e)
		if !ok {
			continue
		}
		if idx.byKey == nil {
			idx.byKey, idx.replacement = map[string]Struck{}, map[string]bool{}
		}
		idx.byKey[c.GetCorrects()] = Struck{Replacement: c.GetReplacement(), By: e.GetSeatId(), Why: c.GetWhy()}
		idx.replacement[c.GetReplacement()] = true
	}
	return idx
}

// Of answers whether the act keyed key was struck, and by what.
func (x StruckIndex) Of(key string) (Struck, bool) {
	s, ok := x.byKey[key]
	return s, ok
}

// IsStruck answers whether the act keyed key was corrected.
func (x StruckIndex) IsStruck(key string) bool { _, ok := x.byKey[key]; return ok }

// IsReplacement answers whether the act keyed key replaced another.
func (x StruckIndex) IsReplacement(key string) bool { return x.replacement[key] }

// Empty answers whether the stream carried no correction at all.
func (x StruckIndex) Empty() bool { return len(x.byKey) == 0 }

// Head follows key's chain to the act that stands now.
func (x StruckIndex) Head(key string) string {
	for i := 0; i < len(x.byKey)+1; i++ {
		s, ok := x.byKey[key]
		if !ok {
			return key
		}
		key = s.Replacement
	}
	return key
}

// IsStruck answers whether ev was corrected, on a stream's index.
func IsStruck(idx StruckIndex, ev *Event) bool { return idx.IsStruck(ev.GetKey()) }

// Live is the stream a WINNER or DUTY reader folds: each struck act replaced IN PLACE by the act
// that stands now, and each replacement dropped from its own later position — so a correction takes
// its target's place in every ordering (F12) and a last-wins reader cannot pick a replacement over a
// later act by the same seat. Correction events themselves stay: a correction is an act.
//
// The stream must carry the Correction events (a full read does; a reader narrowed to a few types
// adds `correction` to its words). A replacement the stream does not carry — narrowed away — leaves
// its struck act out rather than standing in for it.
func Live(evs []*Event) []*Event {
	idx := StruckIndexOf(evs)
	if idx.Empty() {
		return evs
	}
	byKey := make(map[string]*Event, len(evs))
	for _, e := range evs {
		if k := e.GetKey(); k != "" {
			byKey[k] = e
		}
	}
	out := make([]*Event, 0, len(evs))
	for _, e := range evs {
		k := e.GetKey()
		if idx.IsReplacement(k) {
			continue
		}
		if idx.IsStruck(k) {
			if h := byKey[idx.Head(k)]; h != nil {
				out = append(out, h)
			}
			continue
		}
		out = append(out, e)
	}
	return out
}
