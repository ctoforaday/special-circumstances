package cli

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// A VERB'S PROSE LANDS IN EVERY FIELD THAT READS THE CHANNEL.
//
// # The instance this was written for
//
// `line-of-inquiry propose` filled its `line` payload key from the raw --reason FLAG and its
// `reason` key from the resolved CHANNEL. When the channel had a file spelling too, that spelling
// filled only one of the two, and the write was refused for a missing field the seat had supplied.
//
// FOUND BY A SEAT, 2026-08-21, and only because it had somewhere to say so. It filed a friction
// report naming the tool rather than working around it, though it could not name the cause — the
// refusal pointed at `--line`, a flag on no surface.
//
// # What is still true with one spelling
//
// The channel is --reason alone now, so two spellings can no longer disagree. What remains is the
// half that was always the point: every field a verb fills from its prose must receive the prose.
// So this drives each cheap-to-drive prose verb with --reason and asserts the text reached the
// record — and, for propose, that it reached BOTH `line` and `reason`.
// TestNoVerbReadsTheProseFlagDirectly holds the same invariant at every site by source scan.
func TestProseLandsInEveryFieldThatReadsTheChannel(t *testing.T) {
	const prose = "the same paragraph, in every field that reads it"

	// One call per verb, with everything else it needs. fields, where given, are the payload
	// fields that must each hold the prose exactly; otherwise the prose must appear somewhere.
	cases := []struct {
		seat   string
		args   []string
		fields []string
	}{
		{blueSeat, []string{"line-of-inquiry", "propose", "--hypothesis", "h"}, []string{"line", "reason"}},
		{blueSeat, []string{"position"}, nil},
		{blueSeat, []string{"revision"}, nil},
		{blueSeat, []string{"log", "--type", "defect"}, nil},
		// A SEAT THE STAGED BOARD HAS NOT ALREADY USED. The docket board records a `position` for
		// red-chair, and a position is a once-per-sitting act the record REFUSES to repeat
		// rather than dedup — so this case was failing on the fixture's own write, not on the
		// prose channel it is testing.
		{"red-chair", []string{"position"}, nil},
		{"red-chair", []string{"log", "--type", "defect"}, nil},
		{"judge", []string{"certify"}, nil},
		{"judge", []string{"declare"}, nil},
	}

	for _, tc := range cases {
		name := tc.seat + "/" + strings.Join(tc.args, "_")
		t.Run(name, func(t *testing.T) {
			got := recordOnce(t, tc.seat, tc.args, prose)
			quoted := fmt.Sprintf("%q", prose)
			if !strings.Contains(got, quoted) {
				t.Errorf("%s: --reason did not reach the record.\n  payload: %s", name, got)
			}
			for _, f := range tc.fields {
				if !strings.Contains(" "+got+" ", " "+f+"="+quoted+" ") {
					t.Errorf("%s: field %q does not hold the prose — a verb that reads flags.Reason "+
						"directly for one key and the resolved channel for another can disagree, and the "+
						"refusal names the payload key, not the flag the seat used.\n  payload: %s", name, f, got)
				}
			}
		})
	}
}

// recordOnce runs one verb in a fresh board with --reason prose and returns its event payloads,
// minus the fields that differ by construction (ids, nonces, timestamps).
func recordOnce(t *testing.T, seatID string, args []string, prose string) string {
	t.Helper()
	runDir := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	exec := func(a ...string) (string, error) { return run(t, a...) }
	if err := seatprobe.Build(runtest.Open(t, runDir), seatprobe.Boards()["docket"], exec); err != nil {
		t.Fatalf("stage the board: %v", err)
	}
	// THE SEAT SITS BEFORE IT ACTS. The staged docket board already seated red-chair and recorded its
	// position, and a position is once per SITTING — so this register is what makes the act legal:
	// a new register is a new sitting, which is exactly what a re-dispatched seat does. (It used to
	// dodge the collision by acting as red-chair-r2, a second seat; there is one chair now.)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seatID); err != nil {
		t.Fatalf("register %s: %v", seatID, err)
	}
	full := append(append([]string{}, args...), "--run", runDir, "--seat-id", seatID, "--reason", prose)
	if _, err := run(t, full...); err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return payloadOfLast(t, runDir, recordTypeOf(t, args))
}

// recordTypeOf is the event a verb writes.
//
// SMALL, AND CHECKED — which is the part that was missing either way. It began as a hand-kept
// switch defended as "SMALL on purpose". I replaced it with a bare schema lookup on the verb name,
// on the argument that the table was a second list beside the descriptor. That was wrong: a verb
// and the event it writes are DIFFERENT NAMES, deliberately — `line-of-inquiry` writes an `avenue`
// — so the lookup resolved nothing and the test asserted against a word the schema does not carry.
//
// The mapping is real and irreducible, so it stays; what it now does is RESOLVE through the
// descriptor, so a stale entry fails here naming the word it could not find rather than passing
// on a miss. See [[facts-are-fields]] clause 4: find every reader before removing an encoding.
func recordTypeOf(t *testing.T, args []string) recordpb.EventType {
	t.Helper()
	word := args[0]
	if w, ok := map[string]string{"line-of-inquiry": "avenue"}[word]; ok {
		word = w
	}
	vd, ok := recordpb.BySpelling(recordpb.EventType(0).Descriptor(), strings.ReplaceAll(word, "-", "_"))
	if !ok {
		t.Fatalf("%q maps to %q, which is not an event type the schema declares — the verb->event mapping above is stale", args[0], word)
	}
	return recordpb.EventType(vd.Number())
}

// payloadOfLast renders the event's payload as stable text, dropping the fields that differ by
// construction between two separate runs.
func payloadOfLast(t *testing.T, runDir string, typ recordpb.EventType) string {
	t.Helper()
	ev := lastOfType(t, runDir, typ)
	body, ok := recordpb.Body(ev)
	if !ok {
		t.Fatalf("the %s event carries no body", typ)
	}
	// FIELDS OFF THE DESCRIPTOR, not keys off a map. The payload is a typed message now, so
	// "which fields are set" is presence on the message rather than membership in a string map —
	// and an unset field is absent here for the same reason it is absent from the record.
	var keys []string
	seen := map[string]string{}
	body.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch string(fd.Name()) {
		case "inquiry_id", "gap_id":
			return true // minted per run; equality here would be a test of the id generator
		}
		keys = append(keys, string(fd.Name()))
		seen[string(fd.Name())] = v.String()
		return true
	})
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%q ", k, seen[k])
	}
	return strings.TrimSpace(b.String())
}
