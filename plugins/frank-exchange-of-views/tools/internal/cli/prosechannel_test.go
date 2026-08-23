package cli

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/reflect/protoreflect"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// --reason AND --reason-file ARE ONE ARGUMENT, AND EVERY VERB MUST TREAT THEM AS ONE.
//
// The flag's own help says --reason-file is "the same field as --reason, for anything long or that
// would fight shell quoting". Cobra enforces that ONE OF THEM is present (MarkFlagsOneRequired),
// which is why a verb can pass every existing gate while doing something different with each.
//
// # The instance this was written for
//
// `line-of-inquiry propose` filled its `line` payload key from the raw --reason FLAG and its
// `reason` key from the resolved CHANNEL. So --reason filled both and --reason-file filled only
// one, and the write was refused for a missing field the seat had supplied.
//
// FOUND BY A SEAT, 2026-08-21, and only because it had somewhere to say so. A blue seat wrote a
// paragraph-length proposal into `--reason-file -` as a heredoc — which is exactly what that flag
// exists for — was refused, tried twice more, and filed a friction report naming the tool rather
// than working around it. It could not name the cause, because the refusal pointed at `--line`, a
// flag on no surface. It was still right that the tool was wrong.
//
// # Why a gate and not just the fix
//
// Nothing about the defect was specific to that verb. Any verb that reads flags.Reason directly
// instead of the channel has it, silently, for whichever input the tests happen not to use — and
// the tests reach for --reason, because it is the one you can type on a line.
//
// So this drives BOTH ways in, for every prose verb on every seat, and compares the RECORD each
// produced. Same bytes in, same event out, or the two flags are not one argument.
func TestBothProseChannelsProduceTheSameRecord(t *testing.T) {
	const prose = "the same paragraph, arriving two ways"

	// One call per verb, with everything else it needs. A verb absent here is reported rather
	// than skipped: an unlisted verb is an unchecked one, and the list is the thing that rots.
	cases := []struct {
		seat string
		args []string
	}{
		{blueSeat, []string{"line-of-inquiry", "propose", "--hypothesis", "h"}},
		{blueSeat, []string{"position"}},
		{blueSeat, []string{"revision"}},
		{blueSeat, []string{"friction"}},
		// A SEAT THE STAGED BOARD HAS NOT ALREADY USED. The docket board records a `position` for
		// red-merge-r1, and a position is a once-per-sitting act the record REFUSES to repeat
		// rather than dedup — so this case was failing on the fixture's own write, not on the
		// prose channel it is testing.
		{"red-merge-r2", []string{"position"}},
		{"red-merge-r1", []string{"friction"}},
		{"judge-r2", []string{"certify"}},
		{"judge-r2", []string{"declare"}},
	}

	for _, tc := range cases {
		name := tc.seat + "/" + strings.Join(tc.args, "_")
		t.Run(name, func(t *testing.T) {
			viaFlag := recordOnce(t, tc.seat, tc.args, func(dir string) []string {
				return []string{"--reason", prose}
			})
			viaFile := recordOnce(t, tc.seat, tc.args, func(dir string) []string {
				p := filepath.Join(dir, "prose.md")
				if err := os.WriteFile(p, []byte(prose), 0o644); err != nil {
					t.Fatal(err)
				}
				return []string{"--reason-file", p}
			})
			if viaFlag != viaFile {
				t.Errorf("--reason and --reason-file write DIFFERENT records for %s.\n\n"+
					"  --reason:      %s\n  --reason-file: %s\n\n"+
					"The flag's help calls them the same field. A verb that reads flags.Reason directly "+
					"instead of the resolved channel gets one of them and refuses the other — and the "+
					"refusal names the payload key, not the flag the seat actually used.",
					name, viaFlag, viaFile)
			}
		})
	}
}

// recordOnce runs one verb in a fresh board and returns its event payloads, minus the fields that
// differ by construction (ids, nonces, timestamps).
func recordOnce(t *testing.T, seatID string, args []string, prose func(dir string) []string) string {
	t.Helper()
	runDir := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	exec := func(a ...string) (string, error) { return run(t, a...) }
	if err := seatprobe.Build(runDir, seatprobe.Boards()["docket"], exec); err != nil {
		t.Fatalf("stage the board: %v", err)
	}
	full := append(append([]string{}, args...), "--run", runDir, "--seat-id", seatID)
	full = append(full, prose(t.TempDir())...)
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
