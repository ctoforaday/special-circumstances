package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/refusalcheck"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// The CLI is where the role boundary is ENFORCED rather than described, so these
// tests drive the real command tree in-process: a fresh root per invocation (so
// no flag state leaks between cases), the real record package underneath, and a
// temp run dir standing in for the blackboard.
//
// The golden suite in internal/difftest drives the same tree through a BUILT
// BINARY via os/exec, which is why these packages read 0% statement coverage
// while their behaviour is in fact pinned. What that misses is everything a
// subprocess cannot see: which error a verb returns, whether a payload key was
// written at all, and the absent/present distinction the record format rests on.

// testDispatchedSeat mirrors production's dispatchedSeat for a harness that drives cobra with an
// args slice rather than os.Args: the flag wins, and an agent that has registered resolves to the
// seat it bound itself to.
//
// The run directory comes from the args for the same reason the seat id does — a test's --run
// never reaches the environment the binary reads.
func testDispatchedSeat(args []string) string {
	if id := seatenv.SeatIDIn(args); id != "" {
		return id
	}
	runDir := ""
	for i, a := range args {
		if a == "--run" && i+1 < len(args) {
			runDir = args[i+1]
		}
	}
	if runDir == "" {
		if r, err := seatenv.Resolve("", nil); err == nil {
			runDir = r
		}
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		return seatenv.Dispatched(nil)
	}
	return seatenv.Dispatched(seat.BoundSeat(run))
}

// run executes one command against a fresh root and captures what the seat sees.
// Verb output goes to real stdout via fmt.Println, so stdout is redirected
// rather than taken from cobra's writer.
// flagValue reads a flag's value out of an argv the way the parser will, so the harness and the
// command agree about which run directory this call is against.
func flagValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := cutPrefix(a, flag+"="); ok {
			return v
		}
	}
	return ""
}

func cutPrefix(s, p string) (string, bool) {
	if len(s) >= len(p) && s[:len(p)] == p {
		return s[len(p):], true
	}
	return "", false
}

func run(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	stdout, _, err = runSeen(t, args...)
	return stdout, err
}

// runSeen is run with the stderr the binary would have written: the one line a refusal is in its
// human form. A refusal's --json form is on stdout, as the binary puts it.
func runSeen(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	// Output goes through cmd.OutOrStdout(), so a buffer set on the root captures every
	// verb's stdout — no os.Stdout swap, no pipe. Errors are returned (SilenceErrors), so
	// stderr carries nothing the tests read.
	// THE TREE IS THE SEAT'S, resolved the way the binary resolves it: the flag if one is given,
	// otherwise the seat the record binds this agent to.
	//
	// A HARNESS THAT ONLY READ THE FLAG WOULD MEASURE THE WRONG SURFACE. Production's newRoot
	// calls dispatchedSeat, which falls back to the binding on the record — so a registered seat
	// running with no --seat-id gets its own tree. A harness stopping at the flag hands that
	// same call the EMPTY tree and reports "no such command", which is a fact about the harness
	// and reads as a fact about the seat.
	// RELEASE THIS RUN'S HANDLE WHEN THE TEST ENDS, and do it HERE because this is the one place
	// every test in this package reaches the record. recordsql caches a *sql.DB per database and
	// production never closes one — a seat is a process per command. A test is the other shape,
	// and on Windows an open file cannot be removed, so `t.TempDir` cleanup fails and the test is
	// reported as failing for a reason that has nothing to do with what it asserts.
	//
	// Registering from the harness rather than from 56 fixtures is not a shortcut: the handle is
	// opened by PRODUCTION code deep inside the verb, so the test that owns the directory never
	// sees it. The invocation is the only place that knows both the run and the *testing.T.
	if dir := flagValue(args, "--run"); dir != "" {
		t.Cleanup(func() { _ = recordsql.CloseUnder(dir) })
	}
	seatID := testDispatchedSeat(args)
	root := NewRootFor(seatID)
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	// THE HARNESS REPRODUCES THE BINARY by calling what the binary calls: Invoke is the pre-parse
	// refusal, the tree and the one renderer, in the binary's own order, so a refusal here is in
	// the shape and on the stream the binary gives it.
	leaf, err := Invoke(root, args, seatID)
	assertRefusalSpeaksToItsSeat(t, seatID, root, err)
	if err != nil {
		// BOTH FORMS. The envelope's sentence is not the human one where a help page is named
		// instead of printed, so it is read against the receiving seat's tree as well.
		assertRefusalSpeaksToItsSeat(t, seatID, root, errors.New(seat.EnvelopeOf(leaf, err).Error))
	}
	assertRefusalNamesItsCommandOnce(t, leaf, err)
	if err == nil {
		refuseAnUnreadReason(t, root, args)
	}
	// The error is STILL RETURNED. The binary exits 2 whenever it rendered a refusal, so "this call
	// failed" stays true for the tests that assert one.
	return out.String(), errOut.String(), err
}

// assertRefusalNamesItsCommandOnce reads a refusal `run` returned against the command that
// resolved, in both forms: the diagnosis the human line prints after the renderer's head, and the
// sentence the JSON envelope carries beside its `verb`. Neither names the command again
// (internal/refusalcheck holds the reading; refusalonce_test.go says why).
func assertRefusalNamesItsCommandOnce(t *testing.T, leaf *cobra.Command, err error) {
	t.Helper()
	if err == nil || leaf == nil {
		return
	}
	path := seat.PathOf(leaf)
	for form, said := range map[string]string{"human": seat.Diagnosis(err), "--json": seat.EnvelopeOf(leaf, err).Error} {
		if again := refusalcheck.Again(path, roleWords(), said); len(again) > 0 {
			t.Errorf("the refusal of `%s` names its command %d times in its %s form — the renderer prints it from the command, "+
				"and the diagnosis %s; it starts at what is wrong.\n\nrefusal: %v",
				leaf.CommandPath(), 1+len(again), form, strings.Join(again, ", "), said)
		}
	}
}

// seedBlueReport writes a blue/report.md whose prose contains the quoted sentences
// the finding tests anchor to (slice 1b: a lens finding is rejected unless its
// --quote quote is present in the report). One generous body covers the tokens
// the suite uses (§1..§3, the parser quote, sec 1, a quoted sentence).
func seedBlueReport(t *testing.T, runDir string) {
	t.Helper()
	body := "# One\n\n§1 first — a finding sits in sec 1 here.\n\n" +
		"# Two\n\n§2 the finding prose lands in a quoted sentence.\n\n" +
		"# Three\n\nthe parser accepts an empty body in this line.\n"
	dir := filepath.Join(runDir, "blue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ingestReport(t, runDir)
}

// help captures a command's own help output, which cobra writes to its writer.
func help(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	// THE TREE IS THE SEAT'S, and the seat id is in the args the test is about to send. Reading a
	// surface needs no binding — cobra answers --help without reaching RunE — so the flag alone is
	// the whole resolution here, which is also what an unregistered seat's own help read does.
	root := NewRootFor(seatenv.SeatIDIn(args))
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatalf("help %v: %v", args, err)
	}
	return out.String()
}

// events reads the merged event log the way every projection does.
// withoutToolRefusals is every event but the refusals the TOOL logged. A refused call records one thing — its
// refusal, written by the tool so the operator sees what a seat reached for — and nothing the seat
// meant to write. "A refusal records nothing" is now "a refusal records only itself".
func withoutToolRefusals(evs []*record.Event) []*record.Event {
	var out []*record.Event
	for _, e := range evs {
		if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok && l.GetType() == recordpb.LogType_LOG_TYPE_REFUSAL {
			continue
		}
		out = append(out, e)
	}
	return out
}

func events(t *testing.T, runDir string) []*record.Event {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	return m.Events
}

// lastOfType is LAST IN TIME, and the slice IS in time order now.
//
// It used to re-sort by (TS, SeatID, Seq) because MergedEvents returned events in SHARD order —
// "the tail of the slice" was whichever seat file sorted last, which silently returned a seeded
// event instead of the one the test had just written, reading as the verb writing the wrong body.
// The record is one table read `ORDER BY id`, so the tail of the slice is the last thing that
// happened, and re-sorting it by a wall clock could only move an event away from where it was.
func lastOfType(t *testing.T, runDir string, typ recordpb.EventType) *record.Event {
	t.Helper()
	evs := events(t, runDir)
	for i := len(evs) - 1; i >= 0; i-- {
		if evs[i].GetType() == typ {
			return evs[i]
		}
	}
	t.Fatalf("no %s event in the log", typ)
	return nil
}

// lastBody is lastOfType with the body already typed, which is what every caller wanted.
//
// The type argument and the EventType are the SAME FACT — asking for a *recordpb.Cite and then
// naming EVENT_TYPE_CITE beside it is the agreement this migration keeps removing — so the type is
// derived from the body by the same SetBody the write path uses.
func lastBody[T proto.Message](t *testing.T, runDir string, zero T) T {
	t.Helper()
	typ, err := recordpb.SetBody(&record.Event{}, zero)
	if err != nil {
		t.Fatal(err)
	}
	body, ok := recordpb.BodyAs[T](lastOfType(t, runDir, typ))
	if !ok {
		t.Fatalf("the last %s event does not carry a %T", typ, zero)
	}
	return body
}

// setFields is how the absent/present distinction is asserted: a flag the seat never passed must
// not appear in the event AT ALL.
//
// It was payloadKeys over an untyped map. Under the schema the question is the same and the answer
// is stronger: every field is optional, so PRESENCE is a property the record itself carries, and
// protoreflect.Range visits only the populated ones. A field left unset is absent in the same sense
// the old key was missing — which is the distinction the tests are about, and which a struct with
// zero values could not have expressed at all.
func setFields(body proto.Message) map[string]bool {
	out := map[string]bool{}
	body.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		out[string(fd.Name())] = true
		return true
	})
	return out
}

// ---- preconditions shared by every verb ----

func TestEveryVerbRequiresRunAndSeatID(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"lens finding without --run", []string{"finding", "--seat-id", "red-lens-evidence"}, "--run <runDir> is required"},
		// WITHOUT AN IDENTITY THERE IS NO TREE, so the refusal is about the identity rather
		// than about the verb. That is the contract now: the surface is scoped to whoever is
		// asking, and nothing can be shown — or refused — until that is answered.
		{"a verb with no identity at all", []string{"finding", "--run", "X"}, "--seat-id IS REQUIRED HERE"},
		// EACH CASE CARRIES THE VERB'S OTHER REQUIRED FLAGS. Cobra refuses a missing required
		// flag at PARSE, before Begin reaches the run and seat-id checks, so a case that omits
		// them measures whichever refusal fires first rather than the one it is named for.
		{"lens mint without --run", []string{"mint", "--seat-id", "red-lens-evidence",
			"--check", "c", "--check-kind", "document", "--impact", "medium", "--severity", "medium", "--likelihood", "medium",
			"--class", "spec_underspecification", "--problem", "p"}, "--run <runDir> is required"},
		{"mint with no identity at all", []string{"mint", "--run", "X",
			"--check", "c", "--check-kind", "document", "--impact", "medium", "--severity", "medium", "--likelihood", "medium",
			"--class", "spec_underspecification", "--problem", "p"}, "--seat-id IS REQUIRED HERE"},
		{"blue log without --run", []string{"log", "--seat-id", "blue-lane-1", "--type", "defect",
			"--reason", "what got in the way"}, "--run <runDir> is required"},
		{"opinion with no identity at all", []string{"opinion", "--run", "X",
			"--id", "G1", "--as", "remanded", "--principle", "p", "--tension", "t",
			"--review-flag", "no", "--settled", "the proposition this ruling bars", "--final", "--reason", "r"}, "--seat-id IS REQUIRED HERE"},
		{"register is not exempt", []string{"register", "--run", "X"}, "--seat-id IS REQUIRED HERE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// HERMETIC. --run is now INFERRED from .claude/run-live.json when the flag
			// is absent, and the inference walks up from CLAUDE_PROJECT_DIR (or cwd).
			// Without this, the test's cwd is inside this repo and the walk finds
			// whatever run happens to be live on the developer's machine — the verb
			// then runs, the precondition looks broken, and the suite passes or fails
			// by ambient state. Point it at an empty directory so "no run is live"
			// is a fact of the test rather than of the machine.
			t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
			args := make([]string, len(tc.args))
			copy(args, tc.args)
			for i, a := range args {
				if a == "X" {
					args[i] = recordtest.TmpRun(t)
				}
			}
			_, err := runAt(t, args...)
			if err == nil {
				t.Fatal("the verb ran without its preconditions")
			}
			// SUPPLY EVERY OTHER REQUIRED FLAG, so the refusal under test is this one's. The
			// verbs whose flags cobra marks refuse at PARSE time, before Begin reaches the
			// run/seat-id checks — so a case that omits both measures whichever fires first
			// rather than the one it is named for.
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// The verb set is the role boundary — and it is now the WHOLE of it. The chair may not mint: a
// gap is a lens's finding put on the board by the lens that found it (plans/roundless.md
// §III.B.3), and the chair coalesces nothing. The reason is not a runtime check comparing a seat
// id against a role in the path: the verb is not in the chair's tree. So the refusal names the
// seat that does hold `mint` rather than the role the caller failed to be.
func TestRoleBindingIsEnforcedAtTheCLI(t *testing.T) {
	runDir := newRun(t)
	_, err := run(t, "mint", "--run", runDir, "--seat-id", "red-chair",
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err == nil {
		t.Fatal("the CHAIR minted a board gap — the chair coalesces nothing; a lens mints what it found")
	}
	if !strings.Contains(err.Error(), "the lens seat's verb") {
		t.Errorf("the refusal must name the seat that holds the verb: %v", err)
	}
	// And nothing was written under the crossing but the refusal itself.
	if acts := withoutToolRefusals(events(t, runDir)); len(acts) != 0 {
		t.Errorf("a refused cross-role write still produced events: %+v", acts)
	}

	// A seat id no dispatch created is refused too.
	_, err = run(t, "log", "--run", runDir, "--seat-id", "hand-invented", "--type", "defect", "--reason", "t")
	if err == nil {
		t.Fatal("a hand-invented seat id was accepted")
	}
	if !strings.Contains(err.Error(), "does not belong to any role namespace") {
		t.Errorf("wrong refusal: %v", err)
	}
}

// A verb outside the role answers with what the seat CAN do, not cobra's
// generic "unknown command".
func TestUnknownVerbAnswersWithTheAvailableSet(t *testing.T) {
	// (the seat, a verb that is NOT its). The role is now the seat id, because that is what
	// decides the tree — and the claim got stronger: the verb is absent from the surface this
	// seat was given, not merely refused to it.
	cases := []struct{ role, verb string }{
		{"chair", "mint"},
		{"chair", "close"},
		{"blue", "mint"},
		{"blue", "close"},
		{"bench", "mint"},
		{"chair", "edit"},
	}
	for _, tc := range cases {
		t.Run(tc.role+"/"+tc.verb, func(t *testing.T) {
			out, err := run(t, tc.verb, "--run", recordtest.TmpRun(t), "--seat-id", record.SampleSeatOf(tc.role))
			if err == nil {
				t.Fatalf("%s ran the %s verb", tc.role, tc.verb)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("%q is not on your surface", tc.verb)) {
				t.Errorf("error = %q, want it to say the verb is not on this seat's surface", err)
			}
			// THE LIST MOVED FROM THE ERROR TO THE HELP, and that is the fix rather than a
			// regression. This used to require "available: a, b, c" in the message — a
			// hand-rolled list of bare names beside a cobra help that already had the same
			// verbs WITH their descriptions. A seat learns the surface here, at the refusal,
			// so it now gets the real help: names, meanings, flags and the enumerated values.
			//
			// Checked on the combined output because the help goes to stdout and the refusal
			// follows on the error, which is the order a seat reads them in.
			text := out + "\n" + err.Error()
			if !strings.Contains(text, "Available Commands:") || !strings.Contains(text, "show") {
				t.Errorf("the refusal must show what IS available, with descriptions:\n%s", text)
			}
		})
	}
}

// A role invoked with no verb is a usage error, not a no-op: silently succeeding
// would let a mis-scripted seat believe it recorded something.
// A SEAT THAT NAMES NO VERB IS TOLD SO, AND SHOWN ITS SURFACE.
//
// This used to be `<role>` with nothing after it, refused by the role group. There is no role
// level now, so the same situation is the bare invocation — and the claim is unchanged: naming no
// verb is an error that answers with what this seat can do, never a silent success.
func TestASeatWithNoVerbIsAnError(t *testing.T) {
	for _, role := range []string{"lens", "chair", "blue", "bench"} {
		t.Run(role, func(t *testing.T) {
			out, err := run(t, "--seat-id", record.SampleSeatOf(role))
			if err == nil {
				t.Fatalf("%s with no verb succeeded", role)
			}
			if !strings.Contains(err.Error(), "you named no command") {
				t.Errorf("error = %q, want it to say no command was named", err)
			}
			// AND IT TEACHES: the refusal carries this seat's own verbs, which is the whole
			// point of answering rather than exiting.
			if text := out + "\n" + err.Error(); !strings.Contains(text, "Available Commands:") {
				t.Errorf("the refusal showed no surface:\n%s", text)
			}
		})
	}
}

// The log footer closes the loop the help opens: a missing capability is a
// fact about the tooling, not something to improvise around.
func TestRoleHelpCarriesTheLogFooter(t *testing.T) {
	for _, role := range []string{"lens", "chair", "blue", "bench"} {
		t.Run(role, func(t *testing.T) {
			out := help(t, "--help", "--seat-id", record.SampleSeatOf(role))
			if !strings.Contains(out, "it does not exist for you") {
				t.Errorf("%s help lacks the log footer:\n%s", role, out)
			}
			if !strings.Contains(out, "with 'log', as a request") {
				t.Errorf("%s help does not name the log as where a request goes:\n%s", role, out)
			}
		})
	}
}

// The board verbs exist in the lens role and NOWHERE else. Mint, close and regrade are the
// originating lens's acts (plans/roundless.md §III.B.3): the chair carries, spot-checks, dispatches
// and gives the verdict, and the bench disposes through docket rulings. This is the claim the whole
// engine rests on, so it is asserted over the actual command trees.
func TestBoardVerbsExistOnlyInTheLensRole(t *testing.T) {
	// ASKED OF FOUR TREES, not of four groups in one. The claim is stronger for it: a board verb is
	// absent from the chair's SURFACE, not merely absent from a namespace the chair could still
	// address by typing it.
	verbs := map[string]map[string]bool{}
	for role, r := range AllRoots() {
		set := map[string]bool{}
		for _, v := range r.Commands() {
			set[v.Name()] = true
		}
		verbs[role] = set
	}
	for _, board := range []string{"mint", "close", "regrade"} {
		if !verbs["lens"][board] {
			t.Errorf("the lens role is missing the board verb %q", board)
		}
		for _, other := range []string{"chair", "blue", "bench"} {
			if verbs[other][board] {
				t.Errorf("the %s role has the board verb %q — a gap would have a writer other than the lens that minted it", other, board)
			}
		}
	}
	// The chair's duties are the chair's alone: the lens that minted a gap does not carry it,
	// sample the archive, or pronounce on the run.
	for _, duty := range []string{"carry", "spot-check", "verdict", "dispatch"} {
		if !verbs["chair"][duty] {
			t.Errorf("the chair role is missing the chair duty %q", duty)
		}
		for _, other := range []string{"lens", "blue", "bench"} {
			if verbs[other][duty] {
				t.Errorf("the %s role has the chair duty %q", other, duty)
			}
		}
	}
	// Blue has NO board verbs at all, by topology rather than obedience.
	for _, v := range []string{"mint", "close", "regrade", "carry", "verdict", "spot-check"} {
		if verbs["blue"][v] {
			t.Errorf("blue has %q; blue is additive-only and must not be able to subtract", v)
		}
	}
	// Every role can show, and every role can register.
	for _, role := range []string{"lens", "chair", "blue", "bench"} {
		if !verbs[role]["show"] {
			t.Errorf("%s cannot show", role)
		}
		if !verbs[role]["register"] {
			t.Errorf("%s cannot register", role)
		}
	}
}

// ---- the write path, end to end ----

func TestRegisterThenFindingWritesTheRecord(t *testing.T) {
	runDir := newRun(t)
	seedBlueReport(t, runDir)
	seatID := "red-lens-evidence"

	out, err := run(t, "register", "--run", runDir, "--seat-id", seatID)
	if err != nil {
		t.Fatal(err)
	}
	// "registered <seat>" and NOTHING ELSE on a first dispatch. It used to append "(shard nonce
	// <8 hex>)" — an opaque id naming the file the seat would write to. There is no file and no
	// nonce; a re-dispatch says which attempt it is, which is the part a seat can act on.
	if !strings.Contains(out, "registered "+seatID) {
		t.Errorf("register said %q", out)
	}
	if strings.Contains(out, "nonce") || strings.Contains(out, "dispatch") {
		t.Errorf("a FIRST registration should say nothing beyond the seat: %q", out)
	}

	out, err = run(t, "finding", "--run", runDir, "--seat-id", seatID,
		"--key", "F1", "--severity", "high", "--likelihood", "medium", "--impact", "high",
		"--quote", "§2", "--reason", "the finding prose")
	if err != nil {
		t.Fatal(err)
	}
	// The ID leads the message now: it is what the chair will name, and a seat told only
	// "recorded" has to invent a way to refer to this later.
	ev := lastBody(t, runDir, &recordpb.Finding{})
	if got := ev.GetId(); !regexp.MustCompile(`^F-[0-9a-f]{8}$`).MatchString(got) {
		t.Errorf("id = %q, want F-<8 hex>", got)
	}
	// The id is printed WITH the area that raised it: a random id says nothing about whose
	// finding it is, so the line carries both, exactly.
	if want := "finding recorded: " + ev.GetId() + " (evidence) "; !strings.HasPrefix(out, want) {
		t.Errorf("finding said %q, want it to open %q", out, want)
	}
	if got := ev.GetSeverity(); got != recordpb.Grade_GRADE_HIGH {
		t.Errorf("severity = %q", recordpb.Word(got))
	}
	if got := ev.GetText(); got != "the finding prose" {
		t.Errorf("text = %q", got)
	}
	// The key lives on the ENVELOPE, not on the body — the body is what the seat said, the
	// envelope is what the record stamped. The epoch is on neither: it is a window the record
	// computes over the events (a lens registered and no chair has, so 0). The seat id used to
	// carry a round; it carries nothing now, and no event stamps one.
	env := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING)
	if m, err := record.MergedEvents(runtest.Open(t, runDir)); err != nil {
		t.Fatal(err)
	} else if got := m.At.CurrentEpoch(m.Events); got != 0 {
		t.Errorf("epoch = %d, want 0 — no chair has registered in this run", got)
	}
	if env.GetKey() != seatID+":finding:"+ev.GetId() {
		t.Errorf("key = %q", env.GetKey())
	}
}

// A flag the seat never passed must not appear in the event at all — the
// absent/present distinction the record format rests on.
func TestUnpassedFlagsAreAbsentFromThePayload(t *testing.T) {
	runDir := newRun(t)
	seedBlueReport(t, runDir)
	seatID := "red-lens-evidence"
	// --quote and --reason are required (1b); likelihood/impact are still optional,
	// so they are the unpassed flags whose absence this pins.
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", seatID,
		"--key", "F1", "--severity", "high", "--quote", "§1", "--reason", "t"); err != nil {
		t.Fatal(err)
	}
	// `reason` is the FLAG; the field it lands in is `text`. setFields reads the schema, so it
	// reports what the record holds rather than what the seat typed.
	keys := setFields(lastBody(t, runDir, &recordpb.Finding{}))
	if !keys["id"] || !keys["severity"] || !keys["text"] {
		t.Errorf("a passed flag is missing from the body: %v", keys)
	}
	for _, absent := range []string{"likelihood", "impact"} {
		if keys[absent] {
			t.Errorf("%q appears in the payload though the seat never passed it", absent)
		}
	}
}

// THE CSV FIELDS ARE ALWAYS RENDERED, even empty, because an absent key reads as "lineage
// unknown" where the truth is "lineage none".
//
// THE QUESTION MOVED FROM THE BODY TO THE PROJECTION, and it had to. This asserted that the mint
// EVENT carried `supersedes` and `found_by` as present-but-empty keys — a distinction a payload
// map could hold and a proto message cannot: a repeated field has no presence, so an empty list
// and an unset one are the same bytes. There is nothing on the body left to assert.
//
// What a reader actually sees is the projection, and there the distinction is real and preserved:
// viewjson seeds both as `[]string{}` so they render as `[]` rather than being omitted. That is
// the guarantee this test was always about, asked where it can still be answered — and where it
// would fail if someone dropped the seeding.
func TestListFieldsAreAlwaysRenderedEvenWhenEmpty(t *testing.T) {
	// newRun, not a bare TempDir: a run stages its gap-class vocabulary, and a mint against a run
	// with no registry is refused rather than waved through.
	runDir := newRun(t)
	seatID := lensSeat
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "policy_without_mechanism", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	b, err := record.BoardJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, want := range []string{`"supersedes": []`, `"found_by": []`} {
		if !strings.Contains(s, want) {
			t.Errorf("the board does not render %s — an absent lineage key reads as \"lineage unknown\" "+
				"where the truth is \"lineage none\":\n%s", want, s)
		}
	}
	// And a gap that HAS lineage renders it, so the check above is not passing on an empty board.
	if strings.Count(s, `"supersedes"`) != 1 {
		t.Errorf("expected exactly one gap on the board, got %d supersedes keys", strings.Count(s, `"supersedes"`))
	}
}

// Grade validation refuses at the flag, before the payload is built, and the
// message names the values that would have worked.
func TestBadGradeIsRefusedAtParseTimeWithATeachingMessage(t *testing.T) {
	runDir := newRun(t)
	// No register here: the mint is refused at PARSE time and never reaches the write path, and
	// this test asserts the record is left untouched — a register would be an event.
	_, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p", "--severity", "catastrophic")
	if err == nil {
		t.Fatal("an invalid grade was accepted")
	}
	if !strings.Contains(err.Error(), "is not a grade") {
		t.Errorf("error = %q", err)
	}
	for _, g := range []string{"low", "medium", "high", "certain", "realized", "trivial"} {
		if !strings.Contains(err.Error(), g) {
			t.Errorf("the refusal does not offer %q: %v", g, err)
		}
	}
	// Nothing was recorded but the refusal: it came before the write.
	if acts := withoutToolRefusals(events(t, runDir)); len(acts) != 0 {
		t.Errorf("a refused grade still produced events: %+v", acts)
	}
}

// Ids are TOOL-assigned, one per gap and no two alike; --key makes a crash retry
// idempotent rather than double-minting.
func TestMintAssignsAnIDPerGapAndIsIdempotentByKey(t *testing.T) {
	runDir := newRun(t)
	seatID := lensSeat
	registerChairOnce(t, runDir) // the chair sits first: the epoch is its sitting count
	registerLensOnce(t, runDir)  // and the lens mints
	mint := func(extra ...string) string {
		t.Helper()
		args := append([]string{"mint", "--run", runDir, "--seat-id", seatID,
			"--class", "policy_without_mechanism", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}, extra...)
		out, err := run(t, args...)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(out)
	}
	// THE ID IS THE FIRST LINE; the acceptance check is echoed under it. The echo exists because
	// a shell can rewrite a --check before the tool sees it, and the mint used to print only the
	// id — so a corrupted contract was invisible until someone read the board (measured,
	// red-chair, 2026-09-02_quadratic-formula).
	firstLine := func(s string) string { return strings.SplitN(s, "\n", 2)[0] }
	if got := firstLine(mint()); got != handleText(t, runDir, "minted G1") {
		t.Errorf("first mint said %q", got)
	}
	if got := firstLine(mint("--key", "L1-F3")); got != handleText(t, runDir, "minted G2") {
		t.Errorf("second mint said %q", got)
	}
	// The retry: same command, same key, and the EXISTING id comes back.
	// FIRST LINE, because every write now also prints where the seat stands (seat/standing.go) and
	// the subject here is what the verb said about the act.
	if got := firstLine(mint("--key", "L1-F3")); got != handleText(t, runDir, "minted G2 (idempotent retry — existing id returned)") {
		t.Errorf("the retry said %q, want the existing id", got)
	}
	mints := 0
	for _, e := range events(t, runDir) {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_MINT {
			mints++
		}
	}
	if mints != 2 {
		t.Errorf("%d mint events, want 2 — the retry double-minted", mints)
	}
	// A new epoch — the chair sits again — continues the ONE run-global sequence.
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "policy_without_mechanism", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, handleText(t, runDir, "minted G3")) {
		t.Errorf("epoch 2's first mint said %q, want G3", out)
	}
}

// --json is the machine contract: a success is a structured ack (read the id from a field,
// not by parsing prose), a failure is a structured error (branch on ok:false), and the
// default mode stays byte-identical prose because the template renders the same fields.
// AN ACT'S RESULT NAMES ITS ID BY KIND. mint, regrade, close and manifest-row carry `gap_id`,
// inquiry propose `inquiry_id`, prove `proof_id`, motion file and rule `motion_id`; finding carried a
// bare `id` and closing named its gap `id`. m25 ev472: a lens read the mint it had just made at the
// finding's `.result.id`, got null for a mint that landed, and recovered the id from `show work`.
func TestActResultsNameTheirIDByKind(t *testing.T) {
	runDir := newRun(t)
	seedBlueReport(t, runDir)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	result := func(args ...string) map[string]any {
		t.Helper()
		out, err := run(t, append([]string{"--json"}, args...)...)
		if err != nil {
			t.Fatalf("%s: %v", args[0], err)
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &env); err != nil {
			t.Fatalf("%s --json is not JSON (%v): %s", args[0], err, out)
		}
		r, _ := env["result"].(map[string]any)
		if _, bare := r["id"]; bare {
			t.Errorf("%s --json carries a bare id: %v", args[0], r)
		}
		return r
	}
	f := result("finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--severity", "low",
		"--likelihood", "low", "--impact", "low", "--quote", "§1", "--reason", "r")
	fid, _ := f["finding_id"].(string)
	if want := lastBody(t, runDir, &recordpb.Finding{}).GetId(); fid != want {
		t.Errorf("finding --json result %v, want finding_id %s", f, want)
	}
	g := result("mint", "--run", runDir, "--seat-id", lensSeat, "--key", "G1", "--class", "spec_underspecification", "--check-kind", "document",
		"--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p", "--found-by", fid)
	gid, _ := g["gap_id"].(string)
	if gid != handleText(t, runDir, "G1") {
		t.Errorf("mint --json result %v, want gap_id G1", g)
	}
	docketFile(t, runDir, "blue-respond", gid, "the bench's to settle")
	if c := result("closing", "--run", runDir, "--seat-id", lensSeat, "--id", gid, "--reason", "argued"); c["gap_id"] != gid {
		t.Errorf("closing --json result %v, want gap_id %s", c, gid)
	}
}

func TestJSONFlagStructuresResultsAndErrors(t *testing.T) {
	runDir := newRun(t)
	seatID := lensSeat
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seatID); err != nil {
		t.Fatal(err)
	}
	mintArgs := []string{"mint", "--run", runDir, "--seat-id", seatID,
		"--class", "policy_without_mechanism", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}

	out, err := run(t, append([]string{"--json"}, mintArgs...)...)
	if err != nil {
		t.Fatal(err)
	}
	var ok map[string]any
	if e := json.Unmarshal([]byte(strings.TrimSpace(out)), &ok); e != nil {
		t.Fatalf("mint --json is not valid JSON (%v): %s", e, out)
	}
	result, _ := ok["result"].(map[string]any)
	if ok["verb"] != "mint" || ok["ok"] != true || result["gap_id"] != handleText(t, runDir, "G1") {
		t.Errorf("mint --json = %v, want {verb:mint, ok:true, result:{gap_id:G1}}", ok)
	}

	// Default mode is the unchanged prose — the template reproduces it byte for byte.
	plain, err := run(t, append(mintArgs, "--key", "K2")...)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.SplitN(strings.TrimSpace(plain), "\n", 2)[0]; got != handleText(t, runDir, "minted G2") {
		t.Errorf("default mint = %q, want unchanged prose 'minted G2' on the first line", got)
	}

	// A handler failure under --json is structured: ok:false and a message, not a bare exit.
	// Closing a nonexistent gap with no verification anchor fails inside the handler.
	bad, _ := run(t, "--json", "close", "--run", runDir, "--seat-id", seatID, "--id", "NOPE")
	var fail map[string]any
	if e := json.Unmarshal([]byte(strings.TrimSpace(bad)), &fail); e != nil {
		t.Fatalf("mint --json error is not valid JSON (%v): %s", e, bad)
	}
	if fail["ok"] != false {
		t.Errorf("error --json = %v, want ok:false", fail)
	}
	if _, has := fail["error"]; !has {
		t.Errorf("error --json missing an 'error' field: %v", fail)
	}
	if _, has := fail["code"]; !has {
		t.Errorf("error --json missing a 'code' field: %v", fail)
	}

	// A coded domain fault carries its SPECIFIC code, so a consumer branches on the KIND of
	// failure, not the message. A seat-id from another role is a role_violation.
	rv, _ := run(t, "--json", "mint", "--run", runDir, "--seat-id", "blue-synthesize",
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	var viol map[string]any
	if e := json.Unmarshal([]byte(strings.TrimSpace(rv)), &viol); e != nil {
		t.Fatalf("role-violation --json is not valid JSON (%v): %s", e, rv)
	}
	if viol["code"] != "role_violation" {
		t.Errorf("role-violation --json code = %v, want role_violation (env %v)", viol["code"], viol)
	}
}

// --problem and the prose channel are alternatives; --problem wins when both are given.
func TestProseChannelResolution(t *testing.T) {
	// ONE trailing newline is stripped, and that is deliberate normalization rather than
	// mangling: a heredoc's final newline is a line TERMINATOR, not content the seat chose to
	// record. Keeping it would put a stray blank line in every projection of prose captured one
	// way and not of the same text captured another.
	t.Run("--reason is recorded whole, less its terminating newline", func(t *testing.T) {
		runDir := newRun(t)
		body := "line one\nline two — with unicode ✓ and <angle> brackets"
		if _, err := run(t, "position", "--run", runDir, "--seat-id", "red-chair", "--reason", body+"\n"); err != nil {
			t.Fatal(err)
		}
		if got := lastBody(t, runDir, &recordpb.Position{}).GetText(); got != body {
			t.Errorf("text = %q, want the prose without its terminator %q", got, body)
		}
	})

	t.Run("neither channel is not a CHANNEL error", func(t *testing.T) {
		// THE TWO LAYERS ARE SEPARATE, and this is the one that distinguishes them. Omitting
		// --reason is not a channel failure — the channel returns empty (or cobra refuses the
		// missing required flag) and the VERB decides whether empty is acceptable.
		//
		// Every prose verb now refuses an empty reason, because a duty discharged by nothing
		// still counts as discharged. So the refusal must come from the verb's rule and name
		// the flag, NOT from a failed read: a seat told "cannot read prose file" when it simply
		// omitted an argument goes looking for a file it never named.
		runDir := newRun(t)
		out, err := run(t, "position", "--run", runDir, "--seat-id", "red-chair")
		if err == nil {
			t.Fatal("a position with no reason was recorded — an empty position is a duty discharged by nothing")
		}
		// The refusal names the flag owed and carries its usage line ("requires --reason").
		all := out + err.Error()
		if !strings.Contains(all, "reason") {
			t.Errorf("the refusal does not name the flag that fixes it:\n%s", all)
		}
		if strings.Contains(all, "cannot read") || strings.Contains(all, "prose file") {
			t.Errorf("omitting --reason was reported as a FILE read failure, sending the seat after a file it never named:\n%s", all)
		}
		// The seat's own register event is expected; a POSITION event is not.
		for _, e := range events(t, runDir) {
			if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION {
				t.Errorf("a refused position was still recorded: %+v", e)
			}
		}
	})

	t.Run("--problem beats the prose channel on mint", func(t *testing.T) {
		runDir := newRun(t)
		registerLensOnce(t, runDir)
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
			"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "from the flag", "--reason", "from the prose channel"); err != nil {
			t.Fatal(err)
		}
		if got := lastBody(t, runDir, &recordpb.Mint{}).GetProblem(); got != "from the flag" {
			t.Errorf("problem = %q, want --problem to win", got)
		}
	})

	t.Run("the prose channel never fills problem", func(t *testing.T) {
		runDir := newRun(t)
		registerLensOnce(t, runDir)
		_, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
			"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--reason", "from the prose channel")
		if err == nil || !strings.Contains(err.Error(), "requires --problem") {
			t.Errorf("a mint with --reason and no --problem = %v, want it refused for --problem", err)
		}
	})
}

// close carries its evidence or it does not happen: the anchor triple is the
// difference between an auditable closure and an assertion.
func TestCloseRequiresItsAnchor(t *testing.T) {
	runDir := newRun(t)
	seatID := lensSeat // the lens that mints is the one that closes
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}

	// --reason is supplied so the refusal under test is the ANCHOR one: cobra refuses a missing
	// required flag at parse, before the handler that checks the anchor ever runs.
	_, err := runAt(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "G1",
		"--reason", "closed after verification")
	if err == nil {
		t.Fatal("an unanchored closure was accepted")
	}
	if !strings.Contains(err.Error(), "verified-by") {
		t.Errorf("error = %q", err)
	}

	// A partial verification is not a verification, and cobra says so at parse.
	_, err = runAt(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "G1",
		"--reason", "closed after verification", "--verified-by", "L1")
	if err == nil {
		t.Fatal("a partial verification was accepted")
	}

	// The full triple closes it.
	out, err := runAt(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "G1",
		"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:f",
		"--reason", "verified against the ref", "--as", "repaired")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, handleText(t, runDir, "closed G1 (repaired)")) {
		t.Errorf("close said %q", out)
	}
	ev := lastBody(t, runDir, &recordpb.Close{})
	if got := ev.GetClosureClass(); got != recordpb.Disposition_DISPOSITION_REPAIRED {
		t.Errorf("closure_class = %q, want the default \"closed\"", got)
	}

	// Closing an unknown gap is refused before anything is written.
	_, err = runAt(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "G2", "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "t", "--verified-against", "x",
		"--reason", "supplied so the refusal under test is the reference one")
	if err == nil {
		t.Fatal("a closure of an unknown gap was accepted")
	}
	// THE REFUSAL MOVED EARLIER. `--id` is a typed flag carrying record.GapExists now, so
	// seat.Begin resolves it before the handler runs and the message is the reference one. That
	// is the point of moving the check to the flag: the seat is told the id names nothing,
	// rather than being told about an anchor for a gap that does not exist.
	if !strings.Contains(err.Error(), handleText(t, runDir, "names gap G2")) {
		t.Errorf("error = %q", err)
	}
}

func TestCloseWithRegressionRequiresASuccessor(t *testing.T) {
	runDir := newRun(t)
	seatID := lensSeat // the lens that mints is the one that closes
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	// The successor must EXIST: a successor is a reference like any other and is checked
	// at write time now, so the fixture mints the gap it will name.
	succOut, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--key", "successor-gap", "--class", "spec_underspecification", "--check-kind", "document", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	successor := regexp.MustCompile(`G-[0-9a-f]{8}`).FindString(succOut)
	if successor == "" {
		t.Fatalf("could not read the successor id from %q", succOut)
	}
	base := []string{"close", "--run", runDir, "--seat-id", seatID, "--id", "G1",
		"--verified-by", "L1", "--verified-with", "t", "--verified-against", "x",
		"--reason", "verified", "--as", "repaired_with_regression"}
	if _, err := runAt(t, base...); err == nil {
		t.Fatal("repaired_with_regression was accepted without a successor — lineage dropped")
	}
	if _, err := runAt(t, append(base, "--superseded-by", successor)...); err != nil {
		t.Fatalf("a successor'd regression close was refused: %v", err)
	}
}

// close --file carries the full closure record; a missing one is an error.
func TestCloseRecordsItsProse(t *testing.T) {
	runDir := newRun(t)
	seatID := lensSeat // the lens that mints is the one that closes
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "G1",
		"--verified-by", "L1", "--verified-with", "t", "--verified-against", "x", "--reason", "the whole closure record", "--as", "repaired"); err != nil {
		t.Fatal(err)
	}
	if got := lastBody(t, runDir, &recordpb.Close{}).GetProse(); got != "the whole closure record" {
		t.Errorf("prose = %q", got)
	}
}

func TestVerbsThatRefuseWithoutTheirReason(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		// The expectation is the FLAG NAME. Cobra refuses these now, and its message names what
		// is missing without restating what it is for — which the command's own --help already
		// documents, and which the seat is required to have read before running the command.
		{"regrade without --reason", []string{"regrade", "--id", "G1", "--severity", "high"}, "reason"},
		{"mint without --check", []string{"mint", "--class", "spec_underspecification", "--problem", "p"}, "check"},
		{"mint without --class", []string{"mint", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}, "class"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			seedBlueReport(t, runDir)
			seatID := lensSeat
			// The referenced gap must EXIST, or the reference check
			// fires first and this test asserts on the wrong refusal — it is about the
			// missing REASON, not a missing referent.
			registerLensOnce(t, runDir)
			if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
				"--key", "k", "--class", "spec_underspecification", "--check-kind", "document", "--check", "c",
				"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
			if _, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-evidence",
				"--key", "F1", "--quote", "the parser accepts an empty body in this line.", "--reason", "t",
				"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
				t.Fatal(err)
			}
			// EVERY OTHER VALUE IS ONE THE VERB TAKES, or the refusal is of that value and the
			// missing flag is never looked for: the gap is the one minted above, by its own id.
			args := append([]string{tc.args[0], "--run", runDir, "--seat-id", seatID}, tc.args[1:]...)
			for i, a := range args {
				if a == "G1" {
					args[i] = onlyID(t, runDir, "G")
				}
			}
			_, err := runAt(t, args...)
			if err == nil {
				t.Fatal("the verb was accepted without its reason")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to name the missing flag %q", err, tc.wantErr)
			}
		})
	}
}

// blue's verbs refuse the removals and empty inquiries the engine cares about.
func TestBlueVerbContracts(t *testing.T) {
	runDir := newRun(t)
	seatID := "blue-lane-1"

	t.Run("retire needs both the claim and the reason", func(t *testing.T) {
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", seatID, "--reason", "refuted"); err == nil {
			t.Error("a retirement with no quoted claim was accepted")
		}
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", seatID, "--quote", "the claim"); err == nil {
			t.Error("a retirement with no reason was accepted")
		}
		// AND A COMPLETE ONE OVER A REPORT THE RECORD CANNOT RENDER IS REFUSED, naming why: this
		// run has ingested no base, so nothing can say whether the claim ever stood or has left.
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", seatID,
			"--quote", "the claim as it stood", "--reason", "refuted"); err == nil ||
			!strings.Contains(err.Error(), "could not be read from the record") || !strings.Contains(err.Error(), "no base has been ingested") {
			t.Errorf("a retirement over a run with no report was answered %v; want it refused, naming the read", err)
		}
	})

	t.Run("an inquiry is proposed, then has its fate recorded with what changed", func(t *testing.T) {
		out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seatID,
			"--question", "the road taken", "--hypothesis", "it leads somewhere")
		if err != nil {
			t.Fatalf("a proposal needs no fate: %v", err)
		}
		id := inquiryIDOf(out)
		if _, err := run(t, "inquiry", "record", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--as", "declined"); err == nil {
			t.Error("a fate recorded with no reason was accepted — a status that slides silently is decoration")
		}
		if _, err := run(t, "inquiry", "record", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--as", "invented", "--reason", "r"); err == nil {
			t.Error("an undefined inquiry status was accepted; the render groups BY status")
		}
	})
}

// The bench rules with reasons or it does not rule: each field the verb enforces is refused when
// omitted, and the refusal names the missing flag.
//
// IT IS `motion docket rule` NOW. The contract is the deleted `bench opinion`'s, unchanged: the
// same six words, the same "name the flag a seat can type" duty. What moved is the carrier and
// what `--id` MEANS — it named the gap on the old verb and names the MOTION here, because the
// gap rides the filing. That is why the fixture files a docket motion first: without it the
// refusal under test would be the dangling-motion one.
func TestBenchDocketRuleRequiresEachUnconditionalField(t *testing.T) {
	runDir := newRun(t)
	seatID := "judge"
	// The gap must EXIST: `motion docket file --id` is a reference and is checked at write time,
	// so without a real gap the FILING below would be refused and this test would never reach the
	// ruling it is about.
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "k", "--class", "spec_underspecification", "--check-kind", "document", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "motion", "docket", "file", "--run", runDir, "--seat-id", "red-chair",
		"--id", "G1", "--reason", "contested, and not mine to close"); err != nil {
		t.Fatalf("the docket filing was refused, so the ruling has nothing to answer: %v", err)
	}
	// EVERY ENTRY IN `required` IS UNCONDITIONALLY REQUIRED, because the loop below omits each in
	// turn and asserts the refusal names it. --direction is NOT unconditional (a remand's alone, as
	// --reopens-on and --final are a closing ruling's), so it is supplied separately for every call
	// and is not a subtest of its own (#502).
	//
	// --tension, --review-flag and --settled are required by PRESENCE (#1234): an omitted one is
	// refused, and an empty one is an answer — TestBenchDocketRuleTakesAnHonestBlank holds that half.
	full := map[string]string{"id": "M1", "as": "remanded", "principle": "correctness first",
		"tension": "correctness vs economy", "review-flag": "no",
		"settled": "blue must repair C-65ca0a9e"}
	// THE FOUR DocketRuling FIELDS ARE REFUSED BY THE RECORD WRITE, with the schema's own `why`,
	// and as OMITTED — not as "said nothing", which is the refusal an omitted flag written as ""
	// would get instead (#1234). The `why` is read off the field, so this asserts the refusal is
	// the schema's rather than restating its text.
	docketWhy := map[string]string{}
	for _, name := range []protoreflect.Name{"principle", "tension", "review_flag", "settled"} {
		fd := (&recordpb.DocketRuling{}).ProtoReflect().Descriptor().Fields().ByName(name)
		docketWhy[recordpb.FlagFor(fd)] = proto.GetExtension(fd.Options(), recordpb.E_Sql).(*recordpb.Sql).GetWhy()
	}
	for missing := range full {
		t.Run("missing --"+missing, func(t *testing.T) {
			// The remand's --direction is stated, so the refusal reached is the omitted field's own.
			// Without it the record refuses on the direction first, and only a refusal that carried
			// the verb's whole help page could still name the omitted flag.
			args := []string{"motion", "docket", "rule", "--run", runDir, "--seat-id", seatID,
				"--reason", "the rationale", "--direction", "a reproduction on a clean tree"}
			for k, v := range full {
				if k != missing {
					args = append(args, "--"+k, v)
				}
			}
			_, err := runAt(t, args...)
			if err == nil {
				t.Fatalf("a docket ruling was accepted without --%s", missing)
			}
			// The refusal NAMES the missing flag. What the field is for lives in
			// `motion docket rule --help`, which the seat is required to have read before running
			// the command — restating it in the error would be a second copy of the same text,
			// free to drift from the one cobra generates.
			if !strings.Contains(err.Error(), missing) {
				t.Errorf("the refusal does not name --%s: %v", missing, err)
			}
			if why, ok := docketWhy[missing]; ok {
				want := "requires --" + missing + " (" + why + ")"
				if !strings.Contains(err.Error(), want) {
					t.Errorf("omitting --%s is not refused with the schema's reason as an OMISSION.\nwant: %s\ngot:  %v", missing, want, err)
				}
			}
		})
	}
	// AND THE RULER'S OWN PROSE, which the docket ruling made `required` on MotionRule for every
	// subject. It is not in the map above because it is supplied by the prose channel (--reason)
	// rather than by the fields the loop omits by name.
	t.Run("missing --reason", func(t *testing.T) {
		args := []string{"motion", "docket", "rule", "--run", runDir, "--seat-id", seatID}
		for k, v := range full {
			args = append(args, "--"+k, v)
		}
		args = append(args, "--direction", "a reproduction on a clean tree")
		if _, err := runAt(t, args...); err == nil {
			t.Fatal("a ruling with no reasoning was accepted — a verdict the losing party cannot answer")
		} else if !strings.Contains(err.Error(), "reason") {
			t.Errorf("the refusal does not name --reason: %v", err)
		}
	})

	args := []string{"motion", "docket", "rule", "--run", runDir, "--seat-id", seatID,
		"--reason", "the rationale", "--direction", "a reproduction on a clean tree"}
	for k, v := range full {
		args = append(args, "--"+k, v)
	}
	out, err := runAt(t, args...)
	if err != nil {
		t.Fatalf("a complete docket ruling was refused: %v", err)
	}
	if !strings.Contains(out, handleText(t, runDir, "motion M1 ruled remanded")) {
		t.Errorf("the docket ruling said %q", out)
	}
	// THE RULER'S ARGUMENT IS `MotionRule.opinion` — the prose channel every subject's ruling
	// carries — where the retired verb kept it on `Opinion.rationale`.
	if got := lastBody(t, runDir, &recordpb.MotionRule{}).GetOpinion(); got != "the rationale" {
		t.Errorf("the ruler's argument = %q", got)
	}
}

// THE VERB REFUSES A BLANK --tension OR --review-flag AND A REMAND WITH NO DIRECTION, and takes a
// blank --settled (gblock, 2026-10-04).
//
// An empty tension or review flag cannot be told from a bench that skipped the question; a remand
// sends the gap back for one more exchange owing the direction its --direction states, and
// --reopens-on and --final say what would reopen a gap the remand never closed. A ruling may bar no
// proposition, so --settled "" stays an answer — and the verb writes it present, not absent
// (#1234): omission and the blank are kept apart.
func TestBenchDocketRuleRefusesABlankTensionOrReviewFlagAndADirectionlessRemand(t *testing.T) {
	stated := map[string]string{
		"--tension":     "correctness against economy",
		"--review-flag": "none: the reproduction the remand asks for settles it",
		"--settled":     "",
		"--direction":   "a reproduction on the shipped binary",
	}
	rule := func(t *testing.T, runDir, m string, set map[string]string, final bool) error {
		args := []string{"motion", "docket", "rule", "--run", runDir, "--seat-id", "judge",
			"--id", m, "--as", "remanded", "--principle", "correctness first", "--reason", "the rationale"}
		for _, f := range []string{"--tension", "--review-flag", "--settled", "--direction", "--reopens-on"} {
			if v, ok := set[f]; ok {
				args = append(args, f, v)
			}
		}
		if final {
			args = append(args, "--final")
		}
		_, err := runAt(t, args...)
		return err
	}
	docket := func(t *testing.T) (string, string) {
		runDir := newRun(t)
		registerChairOnce(t, runDir)
		registerLensOnce(t, runDir)
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
			"--key", "k", "--class", "spec_underspecification", "--check-kind", "document", "--check", "c",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
			t.Fatal(err)
		}
		return runDir, docketFile(t, runDir, "red-chair", "G1", "contested, and not mine to close")
	}
	// with is the stated set with f answered v; omitting is the stated set with f never passed.
	with := func(f, v string) map[string]string {
		out := maps.Clone(stated)
		out[f] = v
		return out
	}
	omitting := func(f string) map[string]string {
		out := maps.Clone(stated)
		delete(out, f)
		return out
	}
	for _, c := range []struct {
		name, want string
		set        map[string]string
		final      bool
	}{
		{"blank --tension", "requires --tension to say something", with("--tension", ""), false},
		{"whitespace --tension", "requires --tension to say something", with("--tension", "  \t "), false},
		{"blank --review-flag", "requires --review-flag to say something", with("--review-flag", ""), false},
		{"whitespace --review-flag", "requires --review-flag to say something", with("--review-flag", "\n "), false},
		{"remand with --final", "--as remanded requires --direction", omitting("--direction"), true},
		{"remand with a direction and --final", "--as remanded requires --direction, and never --reopens-on or --final", stated, true},
		{"remand with a direction and --reopens-on", "--as remanded requires --direction, and never --reopens-on or --final", with("--reopens-on", "new evidence"), false},
		{"remand with neither", "--as remanded requires --direction", omitting("--direction"), false},
		{"a blank --reopens-on is not an absent one", "--reopens-on is empty", with("--reopens-on", ""), false},
		{"remand with a blank direction", "--as remanded requires --direction", with("--direction", ""), false},
		{"remand with a whitespace direction", "--as remanded requires --direction", with("--direction", " \t"), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir, m := docket(t)
			err := rule(t, runDir, m, c.set, c.final)
			if err == nil {
				t.Fatalf("the verb recorded a ruling with %s", c.name)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("the refusal does not say %q: %v", c.want, err)
			}
		})
	}
	t.Run("stated, with a blank --settled", func(t *testing.T) {
		runDir, m := docket(t)
		if err := rule(t, runDir, m, stated, false); err != nil {
			t.Fatalf("a remand stating its tension, review flag and direction was refused: %v", err)
		}
		d := lastBody(t, runDir, &recordpb.MotionRule{}).GetDocket()
		if d.Settled == nil || *d.Settled != "" {
			t.Errorf("settled = %v, want present and empty — the blank the bench passed", d.Settled)
		}
		if d.GetDirection() != stated["--direction"] || d.ReopensOn != nil {
			t.Errorf("direction = %q, reopens_on = %v; want the direction the bench stated and no reopen condition", d.GetDirection(), d.ReopensOn)
		}
	})
}

// THE SUBJECT IS ASKED BEFORE THE FIELDS. A docket ruling naming a PETITION motion, with a
// required docket field omitted, is refused for the subject: the field's refusal would send the
// seat to supply a flag for a ruling it cannot make under this subgroup at all (#1236). A
// parse-time requirement on the flag would answer first, which is why the four are refused at the
// write instead (#1234).
func TestBenchDocketRuleNamesTheWrongSubjectBeforeAMissingField(t *testing.T) {
	runDir := newRun(t)
	registerLensOnce(t, runDir)
	m := motionID(t, runDir, "motion", "petition", "file", "--seat-id", lensSeat,
		"--class", "safety", "--relief", "halt before the next round", "--reason", "a consent gate is missing")
	for _, omit := range []string{"principle", "tension", "review-flag", "settled"} {
		t.Run("omitted --"+omit, func(t *testing.T) {
			args := []string{"motion", "docket", "rule", "--run", runDir, "--seat-id", "judge",
				"--id", m, "--as", "remanded", "--direction", "a reproduction on the shipped binary", "--reason", "the rationale"}
			for k, v := range map[string]string{"principle": "p", "tension": "t", "review-flag": "r", "settled": "s"} {
				if k != omit {
					args = append(args, "--"+k, v)
				}
			}
			_, err := runAt(t, args...)
			if err == nil {
				t.Fatal("a docket ruling on a petition motion was accepted")
			}
			if !strings.Contains(err.Error(), "filed as a petition motion") || strings.Contains(err.Error(), "--"+omit) {
				t.Errorf("the refusal is not the subject's, first and alone: %v", err)
			}
		})
	}
}

// The shared verbs are ONE contract wherever they appear: a log entry from a
// lens and one from the bench are the same event with the same payload.
func TestSharedVerbsRecordTheSameEventFromEveryRole(t *testing.T) {
	cases := []struct{ role, seatID string }{
		{"lens", "red-lens-evidence"},
		{"chair", "red-chair"},
		{"blue", "blue-lane-1"},
		{"bench", "judge"},
	}
	for _, tc := range cases {
		t.Run("log/"+tc.role, func(t *testing.T) {
			runDir := newRun(t)
			out, err := run(t, "log", "--run", runDir, "--seat-id", tc.seatID,
				"--reason", "the capability I needed", "--type", "defect")
			if err != nil {
				t.Fatal(err)
			}
			// THE FIRST LINE IS THE VERB'S; the lines after it are where the seat now stands, printed
			// by every write (seat/standing.go).
			if want := "log entry recorded: defect"; firstLines(out, 1) != want {
				t.Errorf("log said %q, want %q", out, want)
			}
			ev := lastBody(t, runDir, &recordpb.Log{})
			if got := ev.GetText(); got != "the capability I needed" {
				t.Errorf("text = %q", got)
			}
			// `text`, not `reason`. `--reason` is the word a SEAT types; the field it lands in is
			// spelled per verb, and a log entry stores `text`. setFields reads the schema, so it
			// reports what the record holds rather than what the seat typed.
			//
			// TYPE AND SOURCE RIDE WITH IT, and that is the point of the channel: the prose is
			// what the seat said, the type is what it ASSERTS, and the source is who recorded it.
			// A body carrying only prose is the shape that made an operator read 142,891
			// characters to learn which entries were actionable.
			keys := setFields(ev)
			if len(keys) != 3 || !keys["text"] || !keys["type"] || !keys["source"] {
				t.Errorf("the log body is not prose + type + source: %v", keys)
			}
			if ev.GetType() != recordpb.LogType_LOG_TYPE_DEFECT {
				t.Errorf("type = %v", ev.GetType())
			}
			if ev.GetSource() != recordpb.LogSource_LOG_SOURCE_SEAT {
				t.Errorf("a seat's own entry must record source SEAT, got %v", ev.GetSource())
			}
		})
	}

	// THE PETITION SUB-TEST IS GONE, AND ITS PROPERTY MOVED (#344).
	//
	// It asserted that `<seat> petition` recorded the same event from merge and from blue — the
	// shared-verb property this whole test exists for. That verb is retired: a petition is now
	// `motion petition file`, which is NOT a per-role verb at all. It is one command any seat may
	// run, taking the acting role from the seat's identity rather than from tree position, so
	// "the same event from every role" is not a property that can drift — there is one code path.
	//
	// motion_test.go covers what remains checkable: that the filing mints an id, and that the
	// ruler is the seat holding the gavel for the subject.
}

// closing is keyed on the gap it argues, so two closings in one seat do not
// collide and neither is dedup'd away.
func TestClosingIsKeyedPerGap(t *testing.T) {
	runDir := newRun(t)
	seatID := lensSeat
	// Both gaps must EXIST and be DOCKETED: a closing names the gap it argues, and both the reference
	// and the docketing are checked at write time. The lens mints them and argues each.
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	for i := 0; i < 2; i++ {
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
			"--key", fmt.Sprintf("k%d", i), "--class", "spec_underspecification", "--check-kind", "document", "--check", "c",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range strings.Fields(handleText(t, runDir, "G1 G2")) {
		docketFile(t, runDir, "blue-respond", id, "the bench's to settle")
		if _, err := run(t, "closing", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--reason", "argued "+id); err != nil {
			t.Fatal(err)
		}
	}
	closings := 0
	for _, e := range events(t, runDir) {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_CLOSING {
			closings++
		}
	}
	if closings != 2 {
		t.Errorf("%d closings survived, want 2 — one was dedup'd", closings)
	}
}

// position is a SINGLETON: a seat has one position, and a second is REFUSED rather than quietly
// replacing it.
//
// The refusal is the change and it is the honest half. The shard record deduped on READ — two
// events with one key, one silently discarded — so this test's own header said "replaces" while
// its assertion said the FIRST survives, and a seat that stated its position twice learned
// neither. `events.key` is UNIQUE now, so the second write fails and the seat is told why.
func TestPositionIsASingletonPerSeat(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-chair"
	if _, err := run(t, "position", "--run", runDir, "--seat-id", seatID, "--reason", "first"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "position", "--run", runDir, "--seat-id", seatID, "--reason", "second")
	if err == nil {
		t.Fatal("a second position was accepted — the record would carry two answers to a once-per-sitting question")
	}
	if !strings.Contains(err.Error(), "once-per-sitting") {
		t.Errorf("the refusal does not teach what was wrong:\n%v", err)
	}
	positions := 0
	for _, e := range events(t, runDir) {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION {
			positions++
			p, _ := recordpb.BodyAs[*recordpb.Position](e)
			if got := p.GetText(); got != "first" {
				t.Errorf("the surviving position is %q, want the first", got)
			}
		}
	}
	if positions != 1 {
		t.Errorf("%d positions survived the chair, want 1", positions)
	}
}

// A PASS is a claim that nothing left on the board holds the gate; the tool refuses one over an open material gap —
// the 2026-07-20 rubber-stamp (PASS with 9 open gaps) made structurally impossible. FAIL is
// always allowed.
func TestVerdictPASSRefusedOverOpenGaps(t *testing.T) {
	seatID := "red-chair"
	mint2 := func(runDir string) {
		registerChairOnce(t, runDir)
		registerLensOnce(t, runDir)
		for i := 0; i < 2; i++ {
			// MATERIAL gaps: severity medium holds the gate; a trifle would not (roundless §III.B.2.1).
			if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
				"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
		}
		chairSitsAgain(t, runDir)
	}

	failDir := castRun(t)
	mint2(failDir)
	if _, err := run(t, "verdict", "--run", failDir, "--seat-id", seatID, "--as", "FAIL"); err != nil {
		t.Fatalf("FAIL must be allowed with open gaps: %v", err)
	}

	runDir := castRun(t)
	mint2(runDir)
	_, err := run(t, "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS")
	if err == nil {
		t.Fatal("PASS was recorded over 2 open gaps — the rubber-stamp the guard exists to stop")
	}
	for _, want := range []string{handleText(t, runDir, "G1"), handleText(t, runDir, "G2"), "PASS refused"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the open gaps, got: %v", err)
		}
	}
	// The originator closes: the lens that minted the gaps is the seat that closes them.
	for _, id := range strings.Fields(handleText(t, runDir, "G1 G2")) {
		if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat,
			"--id", id, "--as", "repaired",
			"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "resolved"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := run(t, "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS"); err != nil {
		t.Errorf("PASS must go through once every gap is closed: %v", err)
	}
}

// ...AND THE GATE CANNOT BE SPELLED PAST.
//
// The guard above compares --as to the literal "PASS", so for as long as it has existed
// every OTHER spelling of a pass simply missed it: measured on the shipped tree, `--as
// pass`, `--as Pass` and `--as banana` were all recorded over an open board with the
// check never running. That is a worse failure than the rubber-stamp it was written to
// stop — the rubber-stamp at least left "PASS" on the record, where a re-audit could see
// it. The set is closed at the write path now (record.EnumFields), and this pins the
// route the original guard is reached BY, not just the guard.
func TestVerdictGateCannotBeSpelledPast(t *testing.T) {
	seatID := "red-chair"
	for _, as := range []string{"pass", "Pass", "PASSED", "banana", ""} {
		t.Run("as="+as, func(t *testing.T) {
			runDir := newRun(t)
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
				"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
			args := []string{"verdict", "--run", runDir, "--seat-id", seatID}
			if as != "" {
				args = append(args, "--as", as)
			}
			if _, err := run(t, args...); err == nil {
				t.Fatalf("`verdict --as %q` was recorded over an open gap", as)
			} else if !strings.Contains(err.Error(), "PASS|FAIL") {
				t.Errorf("the refusal must name what would have worked, got: %v", err)
			}
			// AND NOTHING WAS WRITTEN. A refusal that still appends leaves a verdict on
			// the record for every projection to read.
			for _, e := range events(t, runDir) {
				if e.GetType() == recordpb.EventType_EVENT_TYPE_VERDICT {
					t.Errorf("a refused verdict reached the log anyway: %+v", e)
				}
			}
		})
	}
}

// verdict is the chair's terminal act: it renders and checkpoints.
func TestVerdictRendersAndCheckpoints(t *testing.T) {
	runDir := castRun(t)
	seatID := lensSeat // the lens that mints is the one that closes
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	// A PASS is refused over an open material gap, so close it first (the guard is exercised in its
	// own test); this test is about render + checkpoint on a legitimate PASS.
	if _, err := runAt(t, "close", "--run", runDir, "--seat-id", seatID,
		"--id", "G1", "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "resolved"); err != nil {
		t.Fatal(err)
	}
	chairSitsAgain(t, runDir)
	out, err := run(t, "verdict", "--run", runDir, "--seat-id", "red-chair", "--as", "PASS")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "verdict PASS") || !strings.Contains(out, "(0 open, 1 closed)") {
		t.Errorf("verdict said %q", out)
	}
	if !strings.Contains(out, "checkpointed to ") {
		t.Errorf("verdict did not report the mirror: %q", out)
	}
	// The mirror carries the events, so they survive the working tree.
	// TO THE END OF THE LINE, not the end of the output. Every write now prints where the seat stands
	// after it, so slicing to the end swept that text into the path — which is this repository's own
	// lesson about recovering a fact from prose, arriving in a test.
	mirror := strings.TrimSpace(firstLines(out[strings.Index(out, "checkpointed to ")+len("checkpointed to "):], 1))
	entries, rerr := os.ReadDir(mirror)
	if rerr != nil {
		t.Fatalf("the checkpoint mirror is not readable: %v", rerr)
	}
	// THE MIRROR HOLDS THE RECORD, which is one database rather than a set of `events-*.jsonl`
	// files. The checkpoint's purpose is unchanged — a copy of the record survives the run
	// directory — so the assertion is that the record is THERE, by the name it now has.
	var found bool
	for _, e := range entries {
		if e.Name() == "record.db" {
			found = true
		}
	}
	if !found {
		t.Errorf("the mirror does not hold the record: %v", entries)
	}
	// The verdict itself is on the record.
	if got := lastBody(t, runDir, &recordpb.Gate{}).GetVerdict(); got != recordpb.Verdict_VERDICT_PASS {
		t.Errorf("verdict payload = %q", got)
	}
}

// A verb takes no positional arguments: a seat that types prose without --text
// must be told, not have it silently dropped.
func TestVerbsRefusePositionalArguments(t *testing.T) {
	runDir := newRun(t)
	_, err := run(t, "position", "--run", runDir, "--seat-id", "red-chair", "stray prose")
	if err == nil {
		t.Fatal("a positional argument was silently dropped")
	}
	if acts := withoutToolRefusals(events(t, runDir)); len(acts) != 0 {
		t.Errorf("a refused invocation still recorded: %+v", acts)
	}
}

// The register event carries the BUILD this binary came from, so a run that somehow mixes
// binaries says which one wrote what.
//
// It used to carry a hand-maintained semver, which identified an intent to ship rather than
// the artefact: two builds of the same release are indistinguishable under it, and that is
// exactly the case tool_version exists to expose. A revision comes from the build itself, so
// nothing has to remember to move it.
func TestTheBuildIsStampedOnTheFirstAct(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatal(err)
	}
	reg := lastBody(t, runDir, &recordpb.Register{})
	if got := reg.GetToolVersion(); got != buildid.Revision() {
		t.Errorf("tool_version = %q, want the build's revision %q", got, buildid.Revision())
	}
	if reg.GetToolVersion() == "" {
		t.Error("tool_version is empty — a run that mixed binaries could not be told apart from one that did not")
	}
}

// boardState is the replayed board, for assertions about what the events MEAN
// rather than what they say.
func boardState(t *testing.T, runDir string) (record.Family, error) {
	t.Helper()
	return record.FamilyOf(runtest.Open(t, runDir))
}

// The other half of the contract: when a run IS live, a seat that forgot --run is
// served rather than refused. Ten of the first live run's 55 tool-call errors were
// this flag alone, and it cannot be carried in the environment — shell state does not
// persist between tool calls and every subagent shares the parent's session id.
func TestRunDirIsInferredFromTheLiveMarkerWhenTheFlagIsOmitted(t *testing.T) {
	proj := recordtest.TmpRun(t)
	runDir := filepath.Join(proj, "research", "live-run")
	if err := os.MkdirAll(filepath.Join(runDir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(proj, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, ".claude", "run-live.json"),
		[]byte(`{"runs":[{"runDir":"research/live-run"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", proj)

	// No --run. The precondition must pass, so any error is about the verb's own
	// flags rather than the missing run directory.
	_, err := run(t, "register", "--seat-id", "red-lens-evidence")
	if err != nil && strings.Contains(err.Error(), "--run <runDir> is required") {
		t.Fatalf("run dir was not inferred: %v", err)
	}
}

// An explicit --run always beats the marker: inference is a fallback for the seat
// that forgot, never a second source of truth that can override what it was told.
func TestExplicitRunDirBeatsTheInferredOne(t *testing.T) {
	proj := recordtest.TmpRun(t)
	marker := filepath.Join(proj, "research", "marker-run")
	explicit := filepath.Join(proj, "research", "explicit-run")
	for _, d := range []string{marker, explicit, filepath.Join(proj, ".claude")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(proj, ".claude", "run-live.json"),
		[]byte(`{"runs":[{"runDir":"research/marker-run"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", proj)

	if got := runlive.InferRunDir(proj, "").Dir; got != marker {
		t.Fatalf("precondition: marker should resolve to %q, got %q", marker, got)
	}
	// With the flag present the marker must not be consulted at all.
	if _, err := run(t, "register", "--run", explicit, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatalf("explicit --run should be honoured: %v", err)
	}
	if _, err := os.Stat(filepath.Join(explicit, "records")); err != nil {
		t.Fatalf("events should have landed in the explicit run dir: %v", err)
	}
}

// The W1.8 spot-check floor keys on the archive's state at round START, so a round
// entering with zero records has nothing to sample. --ids demanded a list, so this run's
// red-chair could record the discharge only as prose in the ledger — which a later
// audit has to take on trust. An unrecordable discharge is indistinguishable from a
// skipped duty, which is precisely what the event stream exists to prevent.
func TestSpotCheckRecordsAnHonestlyEmptySitting(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-chair",
		"--reason", "archive empty at round start; floor not applicable")
	if err != nil {
		t.Fatalf("an empty archive must be recordable: %v", err)
	}
	if !strings.Contains(out, "nothing to sample") {
		t.Errorf("output %q should say the discharge was empty", out)
	}
	ev := lastBody(t, runDir, &recordpb.SpotCheck{})
	keys := setFields(ev)
	if !keys["reason"] || keys["ids"] {
		t.Errorf("the event must carry its reason and name no closure; got %v", keys)
	}
}

func TestSpotCheckRefusesAnEmptyDischargeWithNoReason(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	// COBRA REFUSES IT AT PARSE now, because the collapse to one prose channel made --reason
	// unconditionally required: what a sample found and why there was nothing to sample were
	// always the same field, and only the branch that wrote it differed.
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-chair"); err == nil ||
		!strings.Contains(err.Error(), "reason") {
		t.Fatalf("a spot-check with no reason must be refused, got %v", err)
	}
}

// A spot-check with no sample still records an empty array — and now also records WHY, because
// the reason is required in both forms. The bare form used to be legal and recorded an empty
// array with no account of it, so "the archive was empty at round start" and "the seat skipped
// its duty" were the same event.
func TestBareSpotCheckStillRecordsAnEmptyArray(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-chair",
		"--reason", "nothing was sampled this round"); err != nil {
		t.Fatalf("the no-sample form must keep working: %v", err)
	}
	// AN EMPTY REPEATED FIELD HAS NO PRESENCE, so this cannot be asked of the body: a proto
	// message cannot tell "sampled nothing" from "field unset". The distinction is real and it
	// lives in the PROJECTION, which renders `ids` as `[]` — a reader distinguishes "checked
	// nothing" from "did not check" by seeing the empty array.
	// The BARE form (no --ids) still records the duty and its reason. A repeated field has no
	// presence, so "sampled nothing" is this event with its reason — see
	// TestSpotCheckIdsAreAlwaysAnArray, which pins all three states.
	sc := lastBody(t, runDir, &recordpb.SpotCheck{})
	if len(sc.GetIds()) != 0 {
		t.Errorf("the bare form sampled %v", sc.GetIds())
	}
	if sc.GetReason() == "" {
		t.Error("the bare form lost its reason — an unexplained empty round is indistinguishable from a skipped one")
	}
}

// `merge close` called its payload --prose-file while every other prose-bearing verb
// calls it --file. Seats typed --file here and were refused, twice, in one run. The shared
// word is --reason now, the one prose spelling, and this pins that close reads it.
func TestCloseAcceptsTheSharedPayloadFlagName(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	const prose = "verified at the leaf; digits match the cited arm"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatal(err)
	}
	minted, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--class", "risk_coverage_omission",
		"--problem", "p", "--fix", "f", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low")
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	id := regexp.MustCompile(`G-[0-9a-f]{8}`).FindString(minted)
	if id == "" {
		t.Fatalf("could not read the minted id from %q", minted)
	}
	// A REAL ANCHOR, not --carried-from. This test's subject is the prose flag, and it used the
	// carry as a shortcut past the anchor requirement — which is exactly how a seat under
	// pressure would use it, and why the carry is now checked against a real earlier
	// closure rather than accepted on its say-so.
	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat,
		"--id", id, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason", prose); err != nil {
		t.Fatalf("--reason must work on close, as it does on every other prose verb: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Close{}).GetProse(); got != prose {
		t.Errorf("the prose from --reason must reach the event: got %q", got)
	}
}

// inquiryIDOf pulls the tool-assigned inquiry id out of a propose result.
func inquiryIDOf(out string) string {
	m := regexp.MustCompile(`\b(Q-[0-9a-f]{8})\b`).FindStringSubmatch(out)
	if m == nil {
		return ""
	}
	return m[1]
}

// fieldText renders one field of an event's body as the WORD the record holds.
//
// Test tables state expectations as (field, value) pairs, which is the right shape — one line per
// fact a verb must record. Reading them off typed bodies with a switch per verb would be forty
// near-identical arms; reading them out of a payload map is what this migration removed.
//
// THE MISS IS LOUD. A name the body does not carry FAILS rather than returning "", because a stale
// field name silently never matches and the assertion then passes for every value — a table row
// that cannot fail reads exactly like one that holds.
//
// Enums render through recordpb.Word, not the generated String(): the table states the vocabulary's
// word, which is what a seat types and what every projection prints.
func fieldText(t *testing.T, body proto.Message, name string) string {
	t.Helper()
	m := body.ProtoReflect()
	fd := m.Descriptor().Fields().ByName(protoreflect.Name(name))
	if fd == nil {
		// THE SUBSTANCE MAY BE ON A ONEOF ARM. A motion's `proposed` grade lives on GradeMotion
		// and a ruling's verdict on the MotionRule_Grade arm — that separation is what makes a
		// ruling from the wrong subject's vocabulary unrepresentable. A seat thinks of them as
		// fields of the act, so the lookup descends into whichever arm is set.
		if arm, af := setArmField(m, name); af != nil {
			m, fd = arm, af
		}
	}
	if fd == nil {
		t.Fatalf("%s has no field %q, on the message or on any set arm — the expectation names a "+
			"field the schema does not carry, so it could never have matched", m.Descriptor().FullName(), name)
	}
	if !m.Has(fd) {
		return ""
	}
	v := m.Get(fd)
	if fd.Kind() == protoreflect.EnumKind {
		return recordpb.Spelling(fd.Enum().Values().ByNumber(v.Enum()))
	}
	return v.String()
}

// setArmField finds a field on whichever message-typed oneof arm is set.
//
// The record models a motion's substance on its subject's arm rather than flat on the message, so
// "the motion proposed `low`" is GradeMotion.proposed. A test table states the fact the way a seat
// would, and this is the one join between the two.
func setArmField(m protoreflect.Message, name string) (protoreflect.Message, protoreflect.FieldDescriptor) {
	md := m.Descriptor()
	for i := 0; i < md.Oneofs().Len(); i++ {
		od := md.Oneofs().Get(i)
		if od.IsSynthetic() {
			continue
		}
		set := m.WhichOneof(od)
		if set == nil || set.Message() == nil {
			continue
		}
		arm := m.Get(set).Message()
		if fd := arm.Descriptor().Fields().ByName(protoreflect.Name(name)); fd != nil {
			return arm, fd
		}
	}
	return m, nil
}

// seedCast puts on runDir the cast setup writes before any seat registers: the four seats these
// fixtures register — the evidence lens, the chair, blue and the bench — and blue-synthesize, the
// seat that ingests the report. A verdict prints the plan the record folds (record.RecordVerdict),
// and the record folds a plan only on a cast.
func seedCast(t *testing.T, runDir string) {
	t.Helper()
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-synthesize", "blue-respond", "judge"}}))
}

// castRun is a run as a real one stands when its lens is first sent to sit: the cast setup writes,
// the report blue-synthesize ingests (seedBlueReport), and the chair's first sitting, whose `dispatch
// next` records the plan naming each cast lens at the report head. A verdict records the plan, which
// pins the report head, so it needs one; and a lens is SAT only through a sitting opened after a
// dispatch naming it — a lens that registers here audits the head, and one that never does holds a
// PASS as "never sat".
func castRun(t *testing.T) string {
	t.Helper()
	dir := newRun(t)
	seedCast(t, dir)
	seedBlueReport(t, dir)
	registerChairOnce(t, dir)
	if out, err := run(t, "dispatch", "next", "--run", dir, "--seat-id", "red-chair"); err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, out)
	}
	return dir
}

// castSeats sits the seats castRun's cast names beside blue-synthesize: the lens castRun dispatched,
// then the chair's next sitting, then blue and the bench.
func castSeats(t *testing.T, runDir string) {
	t.Helper()
	registerLensOnce(t, runDir)
	chairSitsAgain(t, runDir)
	registerSeatOnce(t, runDir, "blue-respond")
	// The bench owes an OCCASION and no other seat may pass one.
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "docket"); err != nil {
		t.Fatalf("register judge: %v", err)
	}
}

// chairSitsAgain opens the chair's next sitting: the workflow comes back to the chair once the seats
// its dispatch named have sat, and a verdict — which records the next plan — is refused in the sitting
// whose dispatch they sat for.
func chairSitsAgain(t *testing.T, runDir string) {
	t.Helper()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatalf("register red-chair: %v", err)
	}
}

// newRun makes a run directory as `setup` leaves it before any seat sits: the empty record. A gap's
// class is a value of the closed GapClass set, so nothing is staged for it.
func newRun(t *testing.T) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.CreatedRecord(t, dir)
	return dir
}
