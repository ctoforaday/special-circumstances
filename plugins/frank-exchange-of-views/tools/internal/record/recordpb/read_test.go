package recordpb_test

import (
	"errors"
	"strings"
	"testing"

	pb "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

func TestClassifiesAWellFormedEvent(t *testing.T) {
	line := marshal(t, sample())
	kind, ev, err := pb.ClassifyLine([]byte(line))
	if err != nil || kind != pb.LineEvent {
		t.Fatalf("kind=%v err=%v, want event/nil", kind, err)
	}
	if ev.GetSeatId() != "red-merge-r1" {
		t.Errorf("seat_id = %q", ev.GetSeatId())
	}
	if kind.IsFatal() {
		t.Error("a good line must not be fatal")
	}
}

// A PRE-SCHEMA LINE ERRORS. This is the whole point of the hard break: it must not render as an
// empty board, which is indistinguishable from a run that produced nothing.
func TestPreSchemaLineIsFatal(t *testing.T) {
	// Well-formed textproto, no schema_version — what a pre-migration binary would produce if it
	// spoke textproto. The JSON case reaches stage 1 and classifies as a fragment, which is
	// covered separately below.
	// The real pre-migration format is JSON, which is what stage 1 keys on. A textproto line
	// missing schema_version is an INCOMPLETE write, not an old one — see the truncation fuzz.
	kind, _, err := pb.ClassifyLine([]byte(`{"seq":1,"seatId":"red-merge-r1","role":"red"}`))
	if kind != pb.LinePreSchema || !kind.IsFatal() {
		t.Fatalf("kind=%v fatal=%v, want pre-schema/fatal", kind, kind.IsFatal())
	}
	if !errors.Is(err, pb.ErrPreSchema) {
		t.Errorf("err = %v, want ErrPreSchema", err)
	}
	if !strings.Contains(err.Error(), "no converter") {
		t.Errorf("the refusal must say there is no converter, so nobody goes looking: %v", err)
	}
}

// A NEWER LINE SAYS *NEWER*. An unmarshal-failure detector would report this as OLDER than the
// binary reading it — the exact opposite of the truth, and the reason the discriminator is a
// positive field.
func TestLineFromNewerReleaseIsFatalAndSaysSo(t *testing.T) {
	line := strings.Replace(marshal(t, sample()),
		"schema_version: SCHEMA_VERSION_1", "schema_version: 99", 1)
	kind, _, err := pb.ClassifyLine([]byte(line))
	if kind != pb.LineFromNewer || !kind.IsFatal() {
		t.Fatalf("kind=%v, want from-newer", kind)
	}
	if !errors.Is(err, pb.ErrFromNewer) {
		t.Errorf("err = %v, want ErrFromNewer", err)
	}
	if !strings.Contains(err.Error(), "Upgrade feov-record") {
		t.Errorf("the refusal must point at the binary, not the record: %v", err)
	}
}

// A KNOWN VERSION THAT DOES NOT PARSE IS CORRUPT — not a fragment, not a skew. Tolerating it
// would silently zero the offending field.
func TestKnownVersionWithUnknownFieldIsCorrupt(t *testing.T) {
	line := marshal(t, sample()) + ` not_a_real_field: "x"`
	kind, _, err := pb.ClassifyLine([]byte(line))
	if kind != pb.LineCorrupt {
		t.Fatalf("kind=%v, want corrupt", kind)
	}
	if !errors.Is(err, pb.ErrCorrupt) {
		t.Errorf("err = %v, want ErrCorrupt", err)
	}
	// Corruption is an ANOMALY, not a fatality: it is indistinguishable from a torn write, and
	// appendLine's healing leaves torn writes mid-shard forever.
	if kind.IsFatal() {
		t.Error("corrupt must not be fatal — it would block the seat's next append")
	}
	if !kind.IsAnomaly() {
		t.Error("corrupt must reach the anomaly footer")
	}
}

// An unknown ENUM VALUE at a known version is corruption too. Measured on protobuf-go v1.36.12:
// with DiscardUnknown=true this does NOT fail — the field is silently dropped to its zero. That
// is why stage 4 is strict.
func TestUnknownEnumValueAtKnownVersionIsCorrupt(t *testing.T) {
	line := strings.Replace(marshal(t, sample()),
		"type: EVENT_TYPE_VERDICT", "type: EVENT_TYPE_FROM_THE_FUTURE", 1)
	kind, _, err := pb.ClassifyLine([]byte(line))
	if kind != pb.LineCorrupt {
		t.Fatalf("kind=%v err=%v, want corrupt — a silently zeroed enum is the plausible zero "+
			"this schema exists to delete", kind, err)
	}
}

func TestEmptyAndBlankLinesAreFragments(t *testing.T) {
	for _, l := range []string{"", "   ", "\t"} {
		if kind, _, err := pb.ClassifyLine([]byte(l)); kind != pb.LineFragment || err != nil {
			t.Errorf("ClassifyLine(%q) = %v, %v — want fragment/nil", l, kind, err)
		}
	}
}

// THE FUZZ THE PLAN OWES: truncate a real line at EVERY byte offset and assert the split holds.
//
// The residue §II.5 states is that a truncation leaving syntactically valid textproto would be
// misclassified as pre-schema and ERROR, when it should drop inert. This is the check that says
// how often that actually happens — and because ReadShard is on the WRITE path (Append reads the
// seat's own shard for the next seq), a false fatal does not fail one replay, it blocks every
// subsequent append by that seat.
//
// A truncated prefix can never be LineEvent: schema_version sits late in the envelope, so any
// prefix short enough to lose the body also loses the version.
func TestTruncationAtEveryOffsetNeverYieldsAFalseEvent(t *testing.T) {
	full := marshal(t, sample())
	var fragments, preSchema, corrupt, events int

	for n := 0; n < len(full); n++ {
		kind, _, _ := pb.ClassifyLine([]byte(full[:n]))
		switch kind {
		case pb.LineFragment:
			fragments++
		case pb.LinePreSchema:
			preSchema++
		case pb.LineCorrupt:
			corrupt++
		case pb.LineEvent:
			events++
			t.Errorf("offset %d: a TRUNCATED line classified as a complete event — replay would "+
				"accept a partial write as real:\n%s", n, full[:n])
		case pb.LineFromNewer:
			t.Errorf("offset %d: truncation produced a from-newer classification", n)
		}
	}
	t.Logf("over %d truncations: fragment=%d pre-schema=%d corrupt=%d event=%d",
		len(full), fragments, preSchema, corrupt, events)

	// The residue, MEASURED rather than asserted away. Any truncation classified pre-schema is a
	// prefix that happened to stay valid textproto while losing schema_version — a false fatal.
	// It is reported as a number so a change in that number is visible, not hidden behind a
	// boolean.
	// NO FALSE FATALS. A truncated line must never be fatal: Append reads the seat's own shard
	// for the next seq, so a false fatal blocks every subsequent write by that seat. The first
	// draft produced 16 of these and the number is now zero BY RULE (stage 3 treats a missing
	// schema_version or body as incomplete), not by luck.
	// NO FALSE PRE-SCHEMA and NO FALSE FROM-NEWER: a truncation must never be mistaken for a
	// version fact. Those two would send an operator hunting a converter or an upgrade.
	if preSchema > 0 {
		t.Errorf("%d truncation points classified pre-schema — a torn write is not an old record", preSchema)
	}

	// THE `corrupt` RESIDUE IS REAL AND MUST NEVER BE FATAL — regardless of position.
	//
	// POSITION CANNOT RESCUE THEM. "Only a shard's last line can be torn" is false the moment
	// appendLine HEALS the fragment — it terminates the torn line and writes the next event AFTER
	// it (durability_test.go: register / sealed fragment / new event, fragment at lines[1]), so a
	// mid-shard fragment is the normal steady state after any crash. Fatal here fails `Append`
	// for that seat forever.
	for n := 0; n < len(full); n++ {
		kind, _, _ := pb.ClassifyLine([]byte(full[:n]))
		if kind.IsFatal() {
			t.Errorf("offset %d: a torn line is FATAL (%v) — after appendLine seals it, this "+
				"blocks every subsequent append by the seat", n, kind)
		}
	}
	// It is inert, not invisible: corrupt lines reach the render anomaly footer.
	if corrupt > 0 {
		var anomalous int
		for n := 0; n < len(full); n++ {
			if kind, _, _ := pb.ClassifyLine([]byte(full[:n])); kind.IsAnomaly() {
				anomalous++
			}
		}
		if anomalous != corrupt {
			t.Errorf("%d corrupt truncations but %d reported as anomalies — a dropped line with "+
				"no trace is the pre-existing behaviour this replaces", corrupt, anomalous)
		}
	}
}

// ONLY THE TWO VERSION FACTS ARE FATAL. Corruption is an ANOMALY: surfaced in the report, inert
// for the replay. Append reads the seat's own shard for the next seq, so anything fatal here
// blocks that seat's writes — a cost only a genuine format break earns.
func TestOnlyVersionFactsAreFatal(t *testing.T) {
	fatal := map[pb.LineKind]bool{
		pb.LineFragment: false, pb.LineEvent: false, pb.LineCorrupt: false,
		pb.LinePreSchema: true, pb.LineFromNewer: true,
	}
	for kind, want := range fatal {
		if got := kind.IsFatal(); got != want {
			t.Errorf("%v.IsFatal() = %v, want %v", kind, got, want)
		}
	}
	if !pb.LineCorrupt.IsAnomaly() {
		t.Error("a corrupt line must reach the anomaly footer — inert is not the same as invisible")
	}
	for _, k := range []pb.LineKind{pb.LineFragment, pb.LineEvent, pb.LinePreSchema, pb.LineFromNewer} {
		if k.IsAnomaly() {
			t.Errorf("%v must not be an anomaly", k)
		}
	}
}

// A JSON line — what the record ACTUALLY held before this migration — must not be silently
// dropped as a fragment. It is the real-world pre-schema case, so it is checked with real bytes.
func TestActualPreMigrationJSONLine(t *testing.T) {
	const jsonLine = `{"seq":0,"ts":"2026-08-10T02:51:23.598112700Z","seatId":"blue-lane-1",` +
		`"nonce":"4b0a6542","round":0,"role":"blue","type":"register",` +
		`"key":"blue-lane-1:register:4b0a6542","payload":{"tool_version":"0.47.0"}}`
	kind, _, err := pb.ClassifyLine([]byte(jsonLine))
	if kind == pb.LineEvent {
		t.Fatal("a pre-migration JSON line parsed as a current event")
	}
	if !kind.IsFatal() {
		t.Errorf("a pre-migration JSON line classified as %v (non-fatal) — it would be dropped "+
			"silently, and a shard of them renders as an empty board, which is exactly the "+
			"plausible zero the hard break exists to convert into an error. err=%v", kind, err)
	}
}
