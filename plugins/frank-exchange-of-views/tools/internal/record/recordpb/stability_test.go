package recordpb_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

func s(v string) *string { return &v }
func i(v int32) *int32   { return &v }

func sample() *pb.Event {
	sv := pb.SchemaVersion_SCHEMA_VERSION_1
	et := pb.EventType_EVENT_TYPE_VERDICT
	vd := pb.Verdict_VERDICT_PASS
	return &pb.Event{
		Seq: i(0), Ts: s("2026-08-17T10:00:00.000000000Z"), SeatId: s("red-merge-r1"),
		Nonce: s("deadbeef"), Round: i(1), Role: s("red"), Type: &et,
		Key: s("red-merge-r1:verdict"), SchemaVersion: &sv,
		Body: &pb.Event_Verdict{Verdict: &pb.RoundVerdict{Verdict: &vd}},
	}
}

func marshal(t *testing.T, ev *pb.Event) string {
	t.Helper()
	b, err := pb.Marshal(ev)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return string(b)
}

// THE TEST WITH TEETH: a FIXED expected byte string.
//
// A marshal-N-times determinism test passes on EVERY build and cannot see the axis detrand
// actually varies — it is seeded per-binary, so output is constant within a process and differs
// across builds of identical source. Only a pinned expectation catches a canonicalization that
// stopped canonicalizing.
func TestEventMarshalsToExactBytes(t *testing.T) {
	const want = `seq: 0 ts: "2026-08-17T10:00:00.000000000Z" seat_id: "red-merge-r1" ` +
		`nonce: "deadbeef" round: 1 role: "red" type: EVENT_TYPE_VERDICT ` +
		`key: "red-merge-r1:verdict" schema_version: SCHEMA_VERSION_1 ` +
		`verdict: { verdict: VERDICT_PASS }`
	if got := marshal(t, sample()); got != want {
		t.Errorf("canonical byte shape moved:\n got: %s\nwant: %s", got, want)
	}
}

// THE ENUM IS NOT A STRING ON THE WIRE. This is the property the whole encoding was chosen for:
// JSON has no enum type, so protojson would render these as `"type":"EVENT_TYPE_VERDICT"` —
// lexically identical to a prose field. A bare identifier says "closed set" to any reader,
// including grep.
func TestEnumsRenderUnquoted(t *testing.T) {
	got := marshal(t, sample())
	for _, want := range []string{
		"type: EVENT_TYPE_VERDICT",
		"schema_version: SCHEMA_VERSION_1",
		"verdict: VERDICT_PASS",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("want unquoted enum %q in:\n%s", want, got)
		}
	}
	for _, never := range []string{`"EVENT_TYPE_VERDICT"`, `"VERDICT_PASS"`} {
		if strings.Contains(got, never) {
			t.Errorf("enum %s is QUOTED — it has been re-flattened into a string, which is the "+
				"one thing this encoding was chosen to prevent:\n%s", never, got)
		}
	}
	// The free-text fields ARE quoted, which is what makes the distinction visible.
	if !strings.Contains(got, `role: "red"`) {
		t.Errorf("free text should stay quoted, so it is distinguishable from an enum:\n%s", got)
	}
}

// seq: 0 MUST SURVIVE. It is the register event's real sequence number, and difftest's rank key
// reads it. This is what `optional` on the envelope buys: without explicit presence a zero is
// indistinguishable from unset and vanishes from the line.
func TestZeroSeqIsEmitted(t *testing.T) {
	if !strings.Contains(marshal(t, sample()), "seq: 0") {
		t.Error("seq: 0 was omitted — a zero without explicit presence vanishes from the record")
	}
}

func TestStableAcrossManyMarshals(t *testing.T) {
	first := marshal(t, sample())
	for n := 0; n < 200; n++ {
		if got := marshal(t, sample()); got != first {
			t.Fatalf("marshal %d differs:\n got: %s\nwant: %s", n, got, first)
		}
	}
}

// ONE LINE PER EVENT IS LOAD-BEARING, not cosmetic: appendLine's torn-line healing, ReadShard's
// split on "\n", and the four-stage read all assume one line is one record. A payload carrying a
// literal newline must therefore escape it, never break the line.
func TestProseWithNewlinesStaysOnOneLine(t *testing.T) {
	ev := sample()
	ev.Body = &pb.Event_Halt{Halt: &pb.Halt{
		Opinion: s("first line\nsecond line\twith tab and \"quotes\" and  runs   of  spaces"),
	}}
	et := pb.EventType_EVENT_TYPE_HALT
	ev.Type = &et

	got := marshal(t, ev)
	if strings.Contains(got, "\n") {
		t.Fatalf("a record spans multiple lines — torn-line healing and ReadShard both break:\n%s", got)
	}

	var rt pb.Event
	if err := pb.Unmarshal([]byte(got), &rt); err != nil {
		t.Fatalf("canonical line does not reparse: %v\n%s", err, got)
	}
	if !proto.Equal(ev, &rt) {
		t.Errorf("round trip lost or altered content:\n got: %v\nwant: %v", &rt, ev)
	}
}

// THE JOIN RESTS ON SOMEBODY ELSE'S INVARIANT, so it is asserted rather than assumed.
//
// canonicalize() folds txtpbfmt's output on newlines. That is safe only because Format emits one
// field per line with every string literal quoted and escaped — an embedded newline is already
// "\n" in the source text and never a real break. If a future txtpbfmt wrapped long strings
// across lines (it has WrapStringsAtColumn), the join would silently corrupt prose. This test is
// what turns that from a hope into a check.
func TestCanonicalFormNeverWrapsLongStrings(t *testing.T) {
	ev := sample()
	et := pb.EventType_EVENT_TYPE_HALT
	ev.Type = &et
	ev.Body = &pb.Event_Halt{Halt: &pb.Halt{Opinion: s(strings.Repeat("long prose ", 60))}}

	got := marshal(t, ev)
	if strings.Contains(got, "\n") {
		t.Fatal("long string was wrapped across lines — the join in canonicalize() would corrupt it")
	}
	var rt pb.Event
	if err := pb.Unmarshal([]byte(got), &rt); err != nil {
		t.Fatalf("long-string line does not reparse: %v", err)
	}
	if rt.GetHalt().GetOpinion() != ev.GetHalt().GetOpinion() {
		t.Error("long prose was altered by canonicalization")
	}
}

func TestUnmarshalRejectsUnknownField(t *testing.T) {
	var ev pb.Event
	err := pb.Unmarshal([]byte(`seq: 1 not_a_real_field: "x"`), &ev)
	if err == nil {
		t.Error("an unknown field was accepted — DiscardUnknown must stay false so a version " +
			"skew is refused by the gate rather than silently dropped")
	}
}

func f64(v float64) *float64 { return &v }

// THE TELEMETRY LINE IS A SECOND BYTE-COMPARED CONTRACT, and floats are where a text encoding is
// most likely to surprise. Pinned: `7.5` not `7.500000`, `1` not `1.0`, full precision on a
// repeating fraction, and no scientific notation until it is genuinely needed.
func TestTelemetryLineMarshalsToExactBytes(t *testing.T) {
	g := pb.Grade_GRADE_HIGH
	two, three, one := int32(2), int32(3), int32(1)
	tl := &pb.TelemetryLine{
		Round: i(1), MappingVersion: s("v2"), OpenCount: &three, MaxSeverity: &g,
		NewMint: &pb.NewMint{
			Count:      &two,
			BySeverity: []*pb.SeverityTally{{Grade: &g, Count: &two}},
			ByClass:    map[string]int32{"zeta": one, "alpha": three, "mid": two},
		},
		Mass:             f64(7.5),
		RepairRegression: &pb.RepairRegression{Closures: &two, LineageMints: &one},
		EdgeDeltas:       &pb.EdgeDeltas{DownMass: f64(0.5), UpMass: f64(0)},
	}
	const want = `round: 1 mapping_version: "v2" open_count: 3 max_severity: GRADE_HIGH ` +
		`new_mint: { count: 2 by_severity: { grade: GRADE_HIGH count: 2 } ` +
		`by_class: { key: "alpha" value: 3 } by_class: { key: "mid" value: 2 } ` +
		`by_class: { key: "zeta" value: 1 } } mass: 7.5 ` +
		`repair_regression: { closures: 2 lineage_mints: 1 } ` +
		`edge_deltas: { down_mass: 0.5 up_mass: 0 }`

	b, err := pb.MarshalMessage(tl)
	if err != nil {
		t.Fatalf("MarshalMessage: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("telemetry byte shape moved:\n got: %s\nwant: %s", got, want)
	}
}

// MAP ORDER IS DETERMINISTIC, and this is the check that says so out loud.
//
// by_class is the one map in the schema (the gap-class registry is open, so a repeated form
// would need an invented sort). Its determinism rests on the encoder sorting map entries by key
// — implementation behaviour, not a documented contract, which is why the protobuf version is
// pinned in go.mod and asserted here rather than assumed.
func TestByClassMapOrderIsStable(t *testing.T) {
	mk := func() *pb.TelemetryLine {
		return &pb.TelemetryLine{NewMint: &pb.NewMint{
			ByClass: map[string]int32{"zeta": 1, "alpha": 3, "mid": 2, "beta": 4, "omega": 5},
		}}
	}
	first, err := pb.MarshalMessage(mk())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), `by_class: { key: "alpha" value: 3 } by_class: { key: "beta" value: 4 }`) {
		t.Errorf("map keys are not in sorted order:\n%s", first)
	}
	for n := 0; n < 100; n++ {
		again, err := pb.MarshalMessage(mk())
		if err != nil {
			t.Fatal(err)
		}
		if string(again) != string(first) {
			t.Fatalf("map ordering varied on marshal %d:\n got: %s\nwant: %s", n, again, first)
		}
	}
}

// realized_open = 0 must SURVIVE: "no realized gaps are open" is a measurement, and it must not
// be indistinguishable from "this round did not compute it".
func TestTelemetryZeroesSurvive(t *testing.T) {
	zero := int32(0)
	b, err := pb.MarshalMessage(&pb.TelemetryLine{RealizedOpen: &zero, Mass: f64(0)})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"realized_open: 0", "mass: 0"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("want %q in %q — a zero measurement must not vanish", want, b)
		}
	}
}
