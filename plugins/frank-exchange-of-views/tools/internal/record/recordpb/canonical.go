package recordpb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/protocolbuffers/txtpbfmt/parser"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// THE SHARD LINE IS TEXTPROTO, ONE EVENT PER LINE, CANONICALIZED.
//
// WHY NOT JSON, having very nearly shipped it. protojson works and was the standing choice for
// several drafts. It was replaced on one argument, which is the only one of three that survives
// scrutiny:
//
//	JSON HAS NO ENUM TYPE. protojson renders a closed-set value as a QUOTED STRING, so on the
//	wire `"type":"EVENT_TYPE_VERDICT"` is lexically indistinguishable from `"role":"red"` or
//	from any prose payload that happens to contain those characters. Measured: a generic reader
//	unmarshalling the JSON gets Go `string` for the free-text field and Go `string` for the
//	enum. The encoding UNDOES, at the boundary, the exact distinction this schema exists to
//	create — a closed set that is no longer a string. textproto emits `type: EVENT_TYPE_VERDICT`
//	as a bare identifier: a symbol, visibly not a string.
//
// The two arguments that did NOT carry it, recorded so they are not re-litigated as though they
// had:
//
//   - "Strong enum typing." True and IRRELEVANT to the choice: EventType is an int32-based Go
//     constant under either encoding. That is the win over the old Payload's p.Str("verdict"),
//     not a reason to prefer textproto.
//   - "Efficiency." Measured on one event: binary=10, protojson-with-numbers=37, textproto=39,
//     protojson-with-names=49. textproto beats protojson by ~20%, but every efficiency argument
//     in this space terminates at "then use binary", which is 4x smaller and was declined for
//     inspectability. Choosing textproto FOR the bytes would be choosing the wrong thing.
//
// AND A NON-ARGUMENT, which is why this comment exists at all: `jq`. Losing it on shards was
// weighed as a real cost for two drafts. It is not one. `jq` is declared ONLY in
// prosthetic-conscience, at tier `optional`, for "JSON slicing for diagnostics and key-scoped
// settings reads"; frank-exchange-of-views — which owns the record — declares node and uv. Its
// presence on any given machine is chance, and a format decision resting on an undeclared,
// doctor-unchecked tool is reasoning from an accident of the environment. The structured read
// path is `feov-record show board`, which emits JSON as a PROJECTION and is unaffected by what
// shards hold.
//
// CANONICALIZATION IS MANDATORY, exactly as json.Compact was for protojson. prototext's
// whitespace comes from internal/detrand, whose header states it "does not change within a
// program" while being "unstable across different builds" — measured here as one space between
// fields in one binary and two in another, from identical source. The goldens compare bytes, so
// an uncanonicalized writer re-records every golden on somebody else's machine.
//
// txtpbfmt is that canonicalizer, and it is a real parser rather than a whitespace heuristic —
// which matters, because textproto separators sit between quoted strings that may themselves
// contain spaces, newlines and quotes. A naive collapse would corrupt seat prose. Verified:
// both detrand spellings converge to identical bytes, the result is idempotent, and prose
// carrying embedded newlines, quotes and space runs round-trips exactly.

// marshalOpts is Multiline:false because the shard is LINE-DELIMITED and that is load-bearing:
// appendLine's torn-line healing, ReadShard's split, and the four-stage read all assume one line
// is one record. txtpbfmt.Format then expands to one field per line and joinLines folds it back,
// which is the only step this package owns.
var marshalOpts = prototext.MarshalOptions{Multiline: false}

// Marshal renders an event as one canonical line, with no trailing newline.
func Marshal(ev *Event) ([]byte, error) {
	raw, err := marshalOpts.Marshal(ev)
	if err != nil {
		return nil, fmt.Errorf("recordpb: marshalling event: %w", err)
	}
	return canonicalize(raw)
}

// MarshalMessage is Marshal for the projections (the telemetry line), which are not Events.
func MarshalMessage(m proto.Message) ([]byte, error) {
	raw, err := marshalOpts.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("recordpb: marshalling %T: %w", m, err)
	}
	return canonicalize(raw)
}

// canonicalize turns detrand-variable textproto into one stable line.
//
// txtpbfmt.Format emits ONE FIELD PER LINE with every string literal quoted and escaped, so
// folding on newlines is safe by construction: an embedded newline inside a value is already
// "\n" in the source text and never a real line break. That property is the whole reason this
// is a join rather than a parse — and it is asserted by a test rather than assumed, because it
// is an invariant of somebody else's formatter.
func canonicalize(raw []byte) ([]byte, error) {
	formatted, err := parser.Format(raw)
	if err != nil {
		// A format failure means the bytes prototext just produced are not parseable as
		// textproto, which is a bug in this package or its dependency — never a seat's doing.
		// It is returned rather than swallowed to the unformatted form: silently writing an
		// uncanonical line would pass every test on THIS build and diverge on the next.
		return nil, fmt.Errorf("recordpb: canonicalizing textproto: %w", err)
	}
	return joinLines(formatted), nil
}

func joinLines(b []byte) []byte {
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	out := lines[:0]
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return []byte(strings.Join(out, " "))
}

// Unmarshal reads one shard line back.
//
// DiscardUnknown is FALSE, deliberately, and §II.5 of plans/record-protobuf.md carries the
// argument: with it true, an unknown enum value does not fail — it is silently dropped to the
// zero value, so a newer binary's `check_kind` reads as UNSPECIFIED in an older one without a
// word. The version gate refuses such a line before the body is ever parsed, which is why the
// tolerant setting is neither needed nor safe here.
var unmarshalOpts = prototext.UnmarshalOptions{DiscardUnknown: false}

func Unmarshal(line []byte, ev *Event) error {
	return unmarshalOpts.Unmarshal(bytes.TrimSpace(line), ev)
}
