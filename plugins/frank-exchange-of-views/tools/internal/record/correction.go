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

// Correct is a correction an ARCHIVED record holds, as migrate replays it: the act it struck, the
// act that replaced it, and why. No seat makes one — no verb takes a correction, and Append refuses
// an identity that carries one outside a migration. An act a seat got wrong is answered by another
// act (supersede.go).
//
// The replayed replacement is the one body of Type written through the identity; Append routes it
// to appendCorrected, which writes it and a Correction event in one transaction.
type Correct struct {
	Type recordpb.EventType // the type of the struck act and of its replacement
	Key  string             // the migrated key of the struck act
	Why  string             // what was wrong with it, in the seat's words

	// written is set once the replacement is on the record: a second body of Type through the
	// same identity would be a second act with no key to name it by.
	written bool
}

// sittingBeforeAndNowSQL is "which sitting was this act filed in, and which sitting is the seat in
// now" — the two numbers the same-sitting check of a replayed correction compares, bound as
// (seat, event id, seat). It counts the sittings the record holds, which is the `sittings` view.
const sittingBeforeAndNowSQL = `SELECT
    (SELECT count(*) FROM "sittings" WHERE "seat_id" = ?1 AND "id" < ?2),
    (SELECT count(*) FROM "sittings" WHERE "seat_id" = ?1)`

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
			"record: a correction names %q, and no act on this record carries that key", key)
	}
	body, ok := recordpb.Body(ev)
	if !ok {
		return nil, fmt.Errorf("record: event %d (key %q) carries no body", id, key)
	}
	return &correctionTarget{ID: id, Key: key, SeatID: ev.GetSeatId(), Type: ev.GetType(), Body: body}, nil
}

func requireSameSeatAndType(t *correctionTarget, seatID string, typ recordpb.EventType) error {
	if t.Type != typ {
		return feov.Errorf(feov.Validation,
			"record: %s is a %s, and its replacement is a %s — a correction replaces an act with one of its own type",
			t.Key, recordpb.Word(t.Type), recordpb.Word(typ))
	}
	if t.SeatID != seatID {
		return feov.Errorf(feov.Validation,
			"record: %s was written by %s, not by %s — a correction strikes only its own seat's act",
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
			"record: %s is a %s, which no correction strikes — %s",
			t.Key, recordpb.Word(t.Type), why)
	}
	if a, b := keyLabel(t.Body), keyLabel(body); a != b {
		return feov.Errorf(feov.Validation,
			"record: a correction keeps the act's label — %s is on %q and this replacement is on %q. An act about something else is a new act, not a correction",
			t.Key, a, b)
	}
	if proto.Equal(t.Body, body) {
		return feov.Errorf(feov.Validation,
			"record: this correction changes nothing — the replacement equals event %d", t.ID)
	}
	if tier == recordpb.CorrectionTier_CORRECTION_TIER_PROSE {
		if diff := frozenDiff(t.Body, body); len(diff) > 0 {
			return feov.Errorf(feov.Validation,
				"record: a %s correction changes only the seat's own wording (%s); this one changes %s, which stays as event %d recorded it",
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
			name := "--" + flagOf(fd)
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
				f := "--" + flagOf(fd)
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
			"record: a correction says what was wrong with the act it strikes, and this one says nothing; a reader sees it beside the struck text")
	}
	if existing, err := correctionRetry(db, seatID, c.Key, body); existing != nil || err != nil {
		// THE RETRY IS ANSWERED WITH THE ACT THAT STANDS.
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
	// F6: THE SAME SITTING. The target's sitting is the writer's sittings opened before it; the
	// writer's current sitting is all of them.
	var before, now int
	if err := tx.QueryRow(sittingBeforeAndNowSQL,
		seatID, target.ID).Scan(&before, &now); err != nil {
		return nil, fmt.Errorf("record: counting %s's sittings: %w", seatID, err)
	}
	if before != now {
		return nil, feov.Errorf(feov.Validation,
			"record: %s was written in an earlier sitting of %s (sitting %d; this is sitting %d) — a correction belongs to the sitting that wrote the act it strikes",
			target.Key, seatID, before, now)
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
			"record: another seat has acted since this %s (event %d, a %s by %s); a correction strikes an act only before any other seat acts",
			recordpb.Word(target.Type), fid, ftype, fseat)
	case !errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("record: asking whether another seat has acted since %s: %w", target.Key, err)
	}
	// A MOTION IS ANSWERED ONCE, and a replayed ruling or appeal does not ask again: both types are
	// PROSE tier, so validateCorrection has held the replacement to its target's motion, and the
	// target is that motion's one answer — the ordinary write refused any other (requireUnanswered).
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
		"record: %s was already corrected — the act that stands now is %s", key, head)
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

// Listed is one act as a LISTING shows it: the act, and — when a correction struck it — who struck
// it and why. A listing never hides a struck act; it marks it.
type Listed struct {
	*Event
	Struck *Struck
}

// Markdown renders an act's text the way every markdown listing shows it: as written, or struck
// through with who struck it and why.
func (l Listed) Markdown(text string) string { return StruckMarkdown(text, l.Struck) }

// StruckMarkdown is the ONE rendering of a struck act's text in markdown: the text struck through,
// then who struck it and why; the text as written when s is nil. The struck text is TRIMMED, because
// a closing ~~ after whitespace does not close in CommonMark — `~~text ~~` renders as literal
// tildes, and the strike a reader is owed does not show.
func StruckMarkdown(text string, s *Struck) string {
	if s == nil {
		return text
	}
	return strikeThrough(text) + " (struck by " + s.By + ": " + s.Why + ")"
}

func strikeThrough(text string) string {
	if t := strings.TrimSpace(text); t != "" {
		return "~~" + t + "~~"
	}
	return "~~(no text)~~"
}

// Strike is Markdown without the note, for a line rendered in parts: every part of a struck act is
// struck, and the one part that carries Markdown says who struck it and why.
func (l Listed) Strike(text string) string {
	if l.Struck == nil {
		return text
	}
	return strikeThrough(text)
}

// Listing is the stream a LISTING renders — the ONE home of that order, as Live is of a fold's.
// The acts that stand, in Live's order, each replacement preceded by the acts it struck (oldest
// first), each of those marked. So a reader sees the struck wording where the act stood, the
// corrected one right after it, and never two acts where the seat made one.
func Listing(evs []*Event) []Listed {
	idx := StruckIndexOf(evs)
	if idx.Empty() {
		out := make([]Listed, len(evs))
		for i, e := range evs {
			out[i] = Listed{Event: e}
		}
		return out
	}
	byKey := make(map[string]*Event, len(evs))
	for _, e := range evs {
		if k := e.GetKey(); k != "" {
			byKey[k] = e
		}
	}
	struckBy := map[string]string{} // replacement key -> the key it struck
	for k, s := range idx.byKey {
		struckBy[s.Replacement] = k
	}
	var out []Listed
	for _, e := range Live(evs) {
		var chain []Listed
		for k := struckBy[e.GetKey()]; k != ""; k = struckBy[k] {
			if old := byKey[k]; old != nil {
				s, _ := idx.Of(k)
				chain = append([]Listed{{Event: old, Struck: &s}}, chain...)
			}
		}
		out = append(out, chain...)
		out = append(out, Listed{Event: e})
	}
	return out
}

// Live is the stream a WINNER or DUTY reader folds: each struck act replaced IN PLACE by the act
// that stands now, and each replacement dropped from its own later position — so a correction takes
// its target's place in every ordering (F12) and a last-wins reader cannot pick a replacement over a
// later act by the same seat. Correction events themselves stay: a correction is an act.
//
// The stream must carry the Correction events (a full read does; a reader narrowed to a few types
// adds `correction` to its words). A replacement the stream does not carry — narrowed away — leaves
// its struck act out rather than standing in for it.
func Live(evs []*Event) []*Event {
	out, _ := liveAt(evs, nil)
	return out
}

// liveAt is Live carrying each standing act's PLACE beside it: seq[i] is evs[i]'s place (events.id,
// or the stream position), and a replacement standing in its target's place takes that place. A
// nil seq carries none.
func liveAt(evs []*Event, seq []int64) ([]*Event, []int64) {
	idx := StruckIndexOf(evs)
	if idx.Empty() {
		return evs, seq
	}
	byKey := make(map[string]*Event, len(evs))
	for _, e := range evs {
		if k := e.GetKey(); k != "" {
			byKey[k] = e
		}
	}
	out := make([]*Event, 0, len(evs))
	var places []int64
	keep := func(e *Event, i int) {
		out = append(out, e)
		if seq != nil {
			places = append(places, seq[i])
		}
	}
	for i, e := range evs {
		k := e.GetKey()
		if idx.IsReplacement(k) {
			continue
		}
		if idx.IsStruck(k) {
			if h := byKey[idx.Head(k)]; h != nil {
				keep(h, i)
			}
			continue
		}
		keep(e, i)
	}
	return out, places
}
