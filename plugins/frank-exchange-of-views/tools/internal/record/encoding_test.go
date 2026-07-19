package record

import (
	"encoding/json"
	"strings"
	"testing"
)

// The serialized event is the record. These tests hold the byte-level contracts
// the package comments declare mandatory — HTML escaping, insertion order, and
// numeric fidelity — because each of them is invisible in a structural
// comparison and each was a real port bug.

// SetEscapeHTML(false) has to hold for the WHOLE line, not just the Event's own
// fields. Seat prose routinely contains angle brackets, and prose lives in the
// payload; escaping half a line is the divergence marshalEvent exists to prevent.
func TestMarshalEventNeverEscapesHTML(t *testing.T) {
	cases := []struct {
		name string
		p    *Payload
	}{
		{"angle brackets in prose", NewPayload().Set("text", "a <seat scratchpad> tag")},
		{"ampersand in prose", NewPayload().Set("text", "this & that")},
		{"all three at once", NewPayload().Set("text", "<a> & <b>")},
		{"in a payload KEY", NewPayload().Set("k<x>&", "v")},
		{"inside a nested payload", NewPayload().Set("by_severity", NewPayload().Set("<undefined>", 1))},
		{"inside a string list", NewPayload().Set("supersedes", []string{"<R1-2>", "a&b"})},
		{"markdown fence with a diff fragment", NewPayload().Set("text", "```diff\n- <old>\n+ <new>\n```")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := Event{SeatID: "red-lens-r1-L1", Type: "finding", Key: "k", Payload: tc.p}
			b, err := marshalEvent(ev)
			if err != nil {
				t.Fatal(err)
			}
			// The escaped forms the encoder must never produce, spelled as
			// interpreted strings so the source carries a literal backslash rather
			// than the character it would escape.
			for _, esc := range []string{"\\u003c", "\\u003e", "\\u0026"} {
				if strings.Contains(string(b), esc) {
					t.Errorf("event carries %s, which JSON.stringify would never emit:\n%s", esc, b)
				}
			}
			// Escaping off must not mean encoding wrong: it still has to parse.
			var back Event
			if err := json.Unmarshal(b, &back); err != nil {
				t.Fatalf("unescaped event does not round-trip: %v\n%s", err, b)
			}
		})
	}
}

func TestMarshalCompactNeverEscapesHTML(t *testing.T) {
	b, err := marshalCompact(NewPayload().Set("max_severity", "<none>").Set("n", 1))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "\\u003c") {
		t.Errorf("telemetry line carries an escaped angle bracket: %s", b)
	}
	if strings.HasSuffix(string(b), "\n") {
		t.Errorf("marshalCompact left a trailing newline: %q", b)
	}
}

// The characters JSON *must* escape are still escaped: turning off HTML escaping
// is not turning off escaping.
func TestMarshalEventStillEscapesWhatJSONRequires(t *testing.T) {
	p := NewPayload().Set("text", "quote \" backslash \\ newline \n tab \t")
	b, err := marshalEvent(Event{Payload: p})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`\"`, `\\`, `\n`, `\t`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("required escape %s is missing from:\n%s", want, b)
		}
	}
	// A literal control character in the output would be invalid JSON.
	if strings.ContainsAny(string(b), "\n\t") {
		t.Errorf("a raw control character survived into the line: %q", b)
	}
	var back Event
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if got := back.Payload.Str("text"); got != p.Str("text") {
		t.Errorf("text did not round-trip: %q", got)
	}
}

// Field order is insertion order, not sorted order: the oracle builds payloads as
// object literals and the render embeds a payload-shaped object.
func TestPayloadMarshalPreservesInsertionOrder(t *testing.T) {
	p := NewPayload().Set("zebra", 1).Set("alpha", 2).Set("mike", 3)
	b, err := p.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"zebra":1,"alpha":2,"mike":3}`; got != want {
		t.Errorf("MarshalJSON() = %s, want %s (Go would sort; the oracle does not)", got, want)
	}
	// Re-setting an existing key updates in place and keeps its original slot.
	p.Set("zebra", 99)
	b, err = p.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"zebra":99,"alpha":2,"mike":3}`; got != want {
		t.Errorf("re-Set moved the key: %s, want %s", got, want)
	}
	if got, want := len(p.Keys()), 3; got != want {
		t.Errorf("re-Set duplicated a key: %d keys, want %d", got, want)
	}
}

func TestPayloadMarshalEmptyIsAnObjectNotNull(t *testing.T) {
	b, err := NewPayload().MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != "{}" {
		t.Errorf("empty payload marshals to %s, want {}", got)
	}
	// A zero-value Payload (never through NewPayload) must behave the same.
	var zero Payload
	b, err = zero.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != "{}" {
		t.Errorf("zero payload marshals to %s, want {}", got)
	}
}

// UseNumber keeps the oracle's numeric text verbatim; a float round-trip would
// turn a large integer into exponent notation and change the rendered ledger.
func TestPayloadUnmarshalKeepsNumbersVerbatim(t *testing.T) {
	cases := []string{"1", "0", "-3", "2.50", "1e3", "9007199254740993", "0.1"}
	for _, n := range cases {
		t.Run(n, func(t *testing.T) {
			var p Payload
			if err := p.UnmarshalJSON([]byte(`{"v":` + n + `}`)); err != nil {
				t.Fatal(err)
			}
			v, _ := p.Get("v")
			num, ok := v.(json.Number)
			if !ok {
				t.Fatalf("value decoded as %T, want json.Number", v)
			}
			if num.String() != n {
				t.Errorf("number became %q, want the verbatim %q", num, n)
			}
			b, err := p.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := string(b), `{"v":`+n+`}`; got != want {
				t.Errorf("re-marshal = %s, want %s", got, want)
			}
		})
	}
}

func TestPayloadUnmarshalPreservesOrderAndRejectsNonObjects(t *testing.T) {
	var p Payload
	if err := p.UnmarshalJSON([]byte(`{"z":1,"a":2,"m":3}`)); err != nil {
		t.Fatal(err)
	}
	want := []string{"z", "a", "m"}
	got := p.Keys()
	if len(got) != len(want) {
		t.Fatalf("Keys() = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keys()[%d] = %q, want %q — decode did not preserve order", i, got[i], want[i])
		}
	}

	for _, bad := range []string{`[]`, `"s"`, `3`, `null`, `true`, `{`, `{"a"}`, ``} {
		var q Payload
		if err := q.UnmarshalJSON([]byte(bad)); err == nil {
			t.Errorf("UnmarshalJSON(%q) was accepted; only an object is a payload", bad)
		}
	}
}

// Unmarshal must RESET, not merge: a reused Payload carrying a previous event's
// keys would attribute one seat's fields to another.
func TestPayloadUnmarshalResetsPriorState(t *testing.T) {
	p := NewPayload().Set("stale", "value")
	if err := p.UnmarshalJSON([]byte(`{"fresh":1}`)); err != nil {
		t.Fatal(err)
	}
	if p.Has("stale") {
		t.Error("a key from the previous decode survived into the new payload")
	}
	if got := p.Keys(); len(got) != 1 || got[0] != "fresh" {
		t.Errorf("Keys() = %q, want [fresh]", got)
	}
}

// Absence is meaningful and distinct from empty — SetIf is the `undefined` drop.
func TestPayloadSetIfDropsAbsentValues(t *testing.T) {
	p := NewPayload().SetIf(false, "absent", "x").SetIf(true, "present", "").SetIf(true, "zero", 0)
	if p.Has("absent") {
		t.Error("SetIf(false) wrote the key; a flag never passed must not appear in the event")
	}
	if !p.Has("present") {
		t.Error("SetIf(true) dropped an EMPTY value; empty is not absent")
	}
	if !p.Has("zero") {
		t.Error("SetIf(true) dropped a ZERO value")
	}
	b, err := p.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"present":"","zero":0}`; got != want {
		t.Errorf("MarshalJSON() = %s, want %s", got, want)
	}
}

// Str deliberately does NOT coerce: a bare boolean flag must fail enum
// validation the same way `true` does in the oracle, not silently become "true".
func TestPayloadStrDoesNotCoerce(t *testing.T) {
	p := NewPayload().Set("b", true).Set("n", 3).Set("s", "x").Set("nil", nil)
	if got := p.Str("b"); got != "" {
		t.Errorf("Str on a bool = %q, want empty — coercion would smuggle it past the enum", got)
	}
	if got := p.Str("n"); got != "" {
		t.Errorf("Str on a number = %q, want empty", got)
	}
	if got := p.Str("nil"); got != "" {
		t.Errorf("Str on nil = %q, want empty", got)
	}
	if got := p.Str("s"); got != "x" {
		t.Errorf("Str on a string = %q, want %q", got, "x")
	}
	if got := p.Str("missing"); got != "" {
		t.Errorf("Str on an absent key = %q, want empty", got)
	}
	// Has must still see the key: absent and non-string are different answers.
	if !p.Has("b") || !p.Has("nil") {
		t.Error("Has() lost a key that Str() declines to coerce")
	}
}

// A nil *Payload is reachable from a decoded event whose payload field was null,
// so the accessors must not panic on it.
func TestPayloadNilReceiverIsSafe(t *testing.T) {
	var p *Payload
	if _, ok := p.Get("k"); ok {
		t.Error("Get on a nil payload reported a hit")
	}
	if p.Str("k") != "" {
		t.Error("Str on a nil payload returned a value")
	}
	if p.Has("k") {
		t.Error("Has on a nil payload returned true")
	}
	if p.StrList("k") != nil {
		t.Error("StrList on a nil payload returned a list")
	}
}

func TestPayloadStrList(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want []string
	}{
		{"native []string", []string{"a", "b"}, []string{"a", "b"}},
		{"decoded []any", []any{"a", "b"}, []string{"a", "b"}},
		{"decoded []any with non-strings filtered", []any{"a", 3, nil, "b"}, []string{"a", "b"}},
		{"empty []string stays empty, not nil-ish", []string{}, []string{}},
		{"a bare string is NOT a list", "a,b", nil},
		{"a number is not a list", 3, nil},
		{"nil is not a list", nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewPayload().Set("k", tc.v).StrList("k")
			if len(got) != len(tc.want) {
				t.Fatalf("StrList() = %q, want %q", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("StrList()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
	if got := NewPayload().StrList("missing"); got != nil {
		t.Errorf("StrList on an absent key = %v, want nil", got)
	}
}

// Keys() is handed to callers; it must not alias the payload's own order slice.
func TestPayloadKeysIsACopy(t *testing.T) {
	p := NewPayload().Set("a", 1).Set("b", 2)
	k := p.Keys()
	k[0] = "clobbered"
	if p.Keys()[0] != "a" {
		t.Error("Keys() aliases the payload's key order; a caller can corrupt serialization")
	}
}

// The whole event round-trips through the same encoder the shard uses, including
// a payload holding every value shape the CLI can produce.
func TestEventRoundTripsThroughShardEncoding(t *testing.T) {
	p := NewPayload().
		Set("gap_id", "R1-1").
		Set("class_new", false).
		Set("supersedes", []string{"R0-1"}).
		Set("problem", "an <em>emphatic</em> & unicode — 日本語 ✓ problem").
		Set("nested", NewPayload().Set("count", 2))
	ev := Event{Seq: 7, SeatID: "red-merge-r1", Nonce: "deadbeef", Round: 1, Type: "mint", Key: "red-merge-r1:mint:R1-1", Payload: p}
	b, err := marshalEvent(ev)
	if err != nil {
		t.Fatal(err)
	}
	var back Event
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	if back.Seq != ev.Seq || back.SeatID != ev.SeatID || back.Nonce != ev.Nonce ||
		back.Round != ev.Round || back.Type != ev.Type || back.Key != ev.Key {
		t.Errorf("event header did not round-trip: %+v", back)
	}
	if got := back.Payload.Str("problem"); got != p.Str("problem") {
		t.Errorf("prose did not round-trip:\n got %q\nwant %q", got, p.Str("problem"))
	}
	if got := back.Payload.StrList("supersedes"); len(got) != 1 || got[0] != "R0-1" {
		t.Errorf("lineage did not round-trip: %q", got)
	}
	// Field order on the wire is now OURS to choose. It mirrored the retired oracle's
	// object literal; `ts` is inserted straight after `seq` because the two together are
	// the ordering key replay sorts by, and a reader scanning a shard wants them
	// adjacent. The order is still PINNED — a silent reshuffle would be a byte-level
	// divergence in every golden, which is the class this test exists to catch.
	if !strings.HasPrefix(string(b), `{"seq":7,"ts":"","seatId":"red-merge-r1","nonce":"deadbeef","round":1,"type":"mint",`) {
		t.Errorf("event field order changed:\n%s", b)
	}
}
