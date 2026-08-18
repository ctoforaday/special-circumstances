package recordpb

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/protocolbuffers/txtpbfmt/parser"
	"google.golang.org/protobuf/encoding/prototext"
)

// CurrentSchemaVersion is what this binary writes and the highest it can read.
const CurrentSchemaVersion = SchemaVersion_SCHEMA_VERSION_1

// LineKind is what a shard line turned out to be.
type LineKind int

const (
	// LineFragment is a torn write from a crashed append. It is DROPPED, inert, exactly as the
	// pre-proto reader dropped it — appendLine's torn-line healing exists to pair with this.
	LineFragment LineKind = iota
	// LineEvent is a well-formed event at a version this binary knows.
	LineEvent
	// LinePreSchema was written before the record had a schema. HARD ERROR.
	LinePreSchema
	// LineFromNewer was written by a later release. HARD ERROR, and the message says NEWER —
	// never "your record is old", which is the lie an unmarshal-failure detector would tell.
	LineFromNewer
	// LineCorrupt claims a version this binary knows but does not parse as it. HARD ERROR.
	LineCorrupt
)

// ErrPreSchema and friends are sentinels so callers can distinguish the fatal classes without
// matching on message text — the string-shaped read this whole schema exists to remove.
var (
	ErrPreSchema = errors.New("recordpb: shard line predates the record schema")
	ErrFromNewer = errors.New("recordpb: shard line was written by a newer release")
	ErrCorrupt   = errors.New("recordpb: shard line claims a known schema version but does not parse as it")
)

// ClassifyLine decides what a line is BEFORE trying to use it.
//
// A TORN FRAGMENT AND A PRE-SCHEMA EVENT ARE BOTH "THIS DOES NOT PARSE", and they must not share
// a fate. The old reader dropped every unparseable line silently, and that tolerance is
// load-bearing: a crash mid-append can leave an unterminated fragment as a shard's last bytes,
// and failing a whole replay over it would trade a plausible zero for a hard outage. But
// dropping a PRE-SCHEMA line silently is how a format break renders as an empty board —
// indistinguishable from a run that did nothing.
//
// THREE PROPERTIES OF THE RULE ARE LOAD-BEARING, and each is a way of being wrong that a
// truncation fuzz over every byte offset will find:
//
//   - A PRE-SCHEMA LINE IS JSON, not textproto missing a field. Keying stage 1 on a missing
//     schema_version drops a real old line as a fragment — silently, which is the plausible zero
//     this function exists to convert into an error.
//   - A LINE WITHOUT A BODY IS NOT AN EVENT. A cut that lands past the envelope leaves valid
//     textproto with no body, and accepting it means replay treats a partial write as real.
//   - A TRUNCATION IS NEVER A VERSION FACT. Since Append reads the seat's own shard to compute
//     the next seq, a torn prefix classified fatal does not fail one replay — it blocks every
//     subsequent append by that seat.
//
// The corrected rule discriminates on what is actually TRUE of each class:
//
//	stage 1  valid JSON object              -> LinePreSchema  (ERROR: the old format WAS json)
//	stage 2  not valid textproto            -> LineFragment   (drop: a torn write)
//	stage 3  no schema_version, or no body  -> LineFragment   (drop: an INCOMPLETE write)
//	stage 4  schema_version > this binary's -> LineFromNewer  (ERROR, naming the version)
//	stage 5  strict parse fails             -> LineCorrupt    (ERROR: not a fragment, not a skew)
//
// Stage 3 is the correction that matters. Completeness is detectable because THIS package writes
// the lines: a canonical event always carries both a schema_version and a body, so a line missing
// either is incomplete by construction rather than by guesswork. The trade is stated plainly —
// a future writer that forgot schema_version would have its lines dropped silently. That is
// guarded at the schema instead: the correspondence test requires every EventType to have a body,
// and Marshal always stamps the version.
func ClassifyLine(line []byte) (LineKind, *Event, error) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return LineFragment, nil, nil
	}

	// Stage 1: is this the OLD FORMAT? Pre-migration shards are JSON objects, one per line. This
	// is checked FIRST and by shape, because a JSON line is not textproto and would otherwise be
	// mistaken for a torn write and dropped without a word.
	if looksLikeJSONObject(trimmed) {
		return LinePreSchema, nil, fmt.Errorf("%w — it was written before the record moved to a "+
			"protobuf schema, and there is no converter. The run's projections (report.md, capture "+
			"output) are unaffected: they were already written", ErrPreSchema)
	}

	// Stage 2: is it textproto SYNTAX at all? This is a schema-FREE parse, and that separation is
	// the correction the fuzz forced.
	//
	// Using a schema-aware unmarshal here conflates two different failures: a torn write, and a
	// line that is perfectly well-formed textproto carrying a value this binary does not know.
	// MEASURED on protobuf-go v1.36.12: prototext rejects an unknown ENUM VALUE regardless of
	// DiscardUnknown (unlike protojson, which tolerates and silently zeroes it — so the reasoning
	// carried over from the protojson draft did not survive the switch). The consequence was that
	// `type: EVENT_TYPE_FROM_THE_FUTURE` failed the permissive parse and was DROPPED AS A
	// FRAGMENT: a corrupt or forward-versioned line vanishing without a word.
	//
	// txtpbfmt parses textproto syntax without knowing the schema, so it answers exactly the
	// question this stage asks and no other.
	if _, err := parser.Parse([]byte(trimmed)); err != nil {
		return LineFragment, nil, nil
	}

	// Stage 3: the line is syntactically textproto, so read the version FROM THE SYNTAX rather
	// than from a decoded message. It has to come from the AST: a line at a future version may
	// carry values this binary cannot decode, and demanding a successful decode before reading
	// the version is what makes an older reader call a newer record corrupt.
	version, hasVersion := schemaVersionFromSyntax(trimmed)
	if !hasVersion || !hasBody(trimmed) {
		// INCOMPLETE, not old. A canonical line always carries both, so a prefix that happens to
		// remain valid textproto is a torn write — drop it inert, exactly as before.
		return LineFragment, nil, nil
	}

	// Stage 4: is it from the future? Refused BY VERSION, before anything tries to decode a value
	// it may not know.
	if version > int64(CurrentSchemaVersion) {
		return LineFromNewer, nil, fmt.Errorf("%w: the line declares schema_version %d and this "+
			"binary reads up to %d. Upgrade feov-record; do not edit the record",
			ErrFromNewer, version, CurrentSchemaVersion)
	}

	// Stage 5: it claims a version we know, so it must parse as that version STRICTLY.
	// DiscardUnknown is false deliberately — an unknown field or enum value at a KNOWN version is
	// corruption, not a skew.
	var ev Event
	if err := (prototext.UnmarshalOptions{DiscardUnknown: false}).Unmarshal([]byte(trimmed), &ev); err != nil {
		return LineCorrupt, nil, fmt.Errorf("%w (declares schema_version %d): %v", ErrCorrupt, version, err)
	}
	return LineEvent, &ev, nil
}

// schemaVersionFromSyntax reads schema_version out of the textproto AST.
//
// Both spellings are accepted because both are legal textproto for an enum field and a record may
// carry either: the symbolic name this package writes (SCHEMA_VERSION_1) and the bare number a
// hand-edit or another implementation might use.
func schemaVersionFromSyntax(line string) (int64, bool) {
	nodes, err := parser.Parse([]byte(line))
	if err != nil {
		return 0, false
	}
	for _, n := range nodes {
		if n.Name != "schema_version" || len(n.Values) == 0 {
			continue
		}
		raw := strings.TrimSpace(n.Values[0].Value)
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return v, true
		}
		if v, ok := SchemaVersion_value[raw]; ok {
			return int64(v), true
		}
		// Present but unreadable: a version we cannot interpret is NOT a missing one. Report it
		// as impossibly high so stage 4 refuses it loudly rather than stage 3 dropping it.
		return int64(CurrentSchemaVersion) + 1, true
	}
	return 0, false
}

// hasBody reports whether the line carries one of the oneof body fields, derived from the
// descriptor rather than a hand-kept list — a list here would be a third carrier of the event-type
// census and would rot the moment a body was added.
func hasBody(line string) bool {
	nodes, err := parser.Parse([]byte(line))
	if err != nil {
		return false
	}
	od := (&Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		return false
	}
	bodies := make(map[string]bool, od.Fields().Len())
	for i := 0; i < od.Fields().Len(); i++ {
		bodies[string(od.Fields().Get(i).Name())] = true
	}
	for _, n := range nodes {
		if bodies[n.Name] {
			return true
		}
	}
	return false
}

// looksLikeJSONObject reports whether a line is the pre-migration format.
//
// Shape, not a full parse: the question is "was this written by the old binary", and the old
// binary emitted exactly one compact JSON object per line. A `{` opening and `}` closing is
// enough to separate it from textproto, which never begins with a brace — and being wrong in the
// permissive direction costs nothing here, because a non-JSON line that somehow matched would
// still be refused rather than silently dropped.
func looksLikeJSONObject(s string) bool {
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
}

// IsFatal reports whether a classification must stop the caller.
//
// ONLY THE TWO VERSION FACTS ARE FATAL, and LineCorrupt is deliberately not one of them.
//
// A truncated body and a corrupted body are the same bytes: measured, 27 of 227 truncation
// offsets parse as textproto syntax while failing the strict schema parse. Nothing in the line
// tells them apart, so the only question is which way to be wrong.
//
// POSITION CANNOT RESCUE IT. "Only a shard's last line can be torn" holds at the instant of
// tearing and fails immediately after, because `appendLine` heals a fragment by terminating it
// and writing the next event AFTER it (durability_test.go pins the shard as register / sealed
// fragment / new event, fragment at lines[1]). A position rule therefore makes a mid-shard line
// permanently fatal, and since `Append` reads the seat's own shard for the next seq, that seat
// never writes again and every replay of the run fails with it.
//
// So availability wins: a false fatal is an outage, a false fragment repeats the tolerance the
// record has had for its whole life.
//
// It is NOT silently dropped. LineCorrupt carries its error to the caller, which surfaces it in
// the render anomaly footer (viewjson.Anomalies) — visible in the report, never blocking a write.
// Loudness and fatality are different properties and this is where they separate.
//
// What the hard break actually needs stays fatal: a PRE-SCHEMA line (a format break, which would
// otherwise render as an empty board) and a line FROM A NEWER RELEASE (a version skew this binary
// cannot honestly interpret). Both are version FACTS read off a field, not guesses about damage.
func (k LineKind) IsFatal() bool {
	return k == LinePreSchema || k == LineFromNewer
}

// IsAnomaly reports whether a line should be surfaced in the render anomaly footer. A corrupt
// line is inert for the replay and VISIBLE in the report — the pre-existing reader dropped it
// with no trace at all.
func (k LineKind) IsAnomaly() bool { return k == LineCorrupt }

func (k LineKind) String() string {
	switch k {
	case LineFragment:
		return "fragment"
	case LineEvent:
		return "event"
	case LinePreSchema:
		return "pre-schema"
	case LineFromNewer:
		return "from-newer"
	case LineCorrupt:
		return "corrupt"
	}
	return "unknown"
}
