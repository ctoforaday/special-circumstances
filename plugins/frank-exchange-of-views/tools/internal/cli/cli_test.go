package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
	return seatenv.Dispatched(seat.BoundSeat(runDir))
}

// run executes one command against a fresh root and captures what the seat sees.
// Verb output goes to real stdout via fmt.Println, so stdout is redirected
// rather than taken from cobra's writer.
func run(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	// Output goes through cmd.OutOrStdout(), so a buffer set on the root captures every
	// verb's stdout — no os.Stdout swap, no pipe. Errors are returned (SilenceErrors), so
	// stderr carries nothing the tests read.
	// THE TREE IS THE SEAT'S, resolved the way the binary resolves it: the flag if one is given,
	// otherwise the seat this agent registered as.
	//
	// A HARNESS THAT ONLY READ THE FLAG WOULD MEASURE THE WRONG SURFACE. Production's newRoot
	// calls dispatchedSeat, which falls back to the binding on the record — so a registered seat
	// running with no --seat-id gets its own tree. A harness stopping at the flag hands that
	// same call the EMPTY tree and reports "no such command", which is a fact about the harness
	// and reads as a fact about the seat.
	seatID := testDispatchedSeat(args)
	root := NewRootFor(seatID)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	// THE HARNESS REPRODUCES THE BINARY, and this pre-check is part of it: Execute runs
	// refuseUnknownCommandFirst before cobra parses anything, so a command named by a caller with
	// no identity is answered about the identity rather than about its flags.
	if err = refuseUnknownCommandFirst(root, append([]string{"feov-record"}, args...), seatID); err == nil {
		err = ExecuteRoot(root)
	}
	// THE HARNESS REPRODUCES THE BINARY. A flag-PARSE refusal never reaches seat.Emit, so
	// Execute() renders it as a --json envelope itself; a harness that skipped that step would
	// measure a wire shape the binary does not produce.
	if err != nil {
		EmitTopLevelError(&out, args, err)
	}
	// The error is STILL RETURNED. The binary exits 2 whether or not it rendered an envelope, so
	// "this call failed" stays true for the tests that assert a refusal; the envelope is an
	// additional rendering, not a substitute for the failure.
	return out.String(), err
}

// seedBlueReport writes a blue/report.md whose prose contains the quoted sentences
// the finding tests anchor to (slice 1b: a lens finding is rejected unless its
// --quote quote is present in the report). One generous body covers the tokens
// the suite uses (§1..§3, the parser quote, sec 1, a quoted sentence).
func seedBlueReport(t *testing.T, runDir string) {
	t.Helper()
	body := "# §1\n\n§1 first — a finding sits in sec 1 here.\n\n" +
		"# §2\n\n§2 the finding prose lands in a quoted sentence.\n\n" +
		"# §3\n\nthe parser accepts an empty body in this line.\n"
	dir := filepath.Join(runDir, "blue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
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
func events(t *testing.T, runDir string) []record.Event {
	t.Helper()
	m, err := record.MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	return m.Events
}

// lastOfType is LAST IN TIME, not last in the slice.
//
// MergedEvents returns events in shard order, so "the tail of the slice" is whichever seat
// file sorts last — fine while a test run dir held one seat's events, wrong the moment
// fixtures seed events from several seats. It silently returned the SEEDED dispute instead
// of the one the test had just filed, which reads as the verb writing the wrong payload.
// The canonical order is (TS, SeatID, Seq), the same key replay uses.
func lastOfType(t *testing.T, runDir, typ string) record.Event {
	t.Helper()
	evs := events(t, runDir)
	sort.SliceStable(evs, func(i, j int) bool {
		a, b := evs[i], evs[j]
		if a.TS != b.TS {
			return a.TS < b.TS
		}
		if a.SeatID != b.SeatID {
			return a.SeatID < b.SeatID
		}
		return a.Seq < b.Seq
	})
	for i := len(evs) - 1; i >= 0; i-- {
		if evs[i].Type == typ {
			return evs[i]
		}
	}
	t.Fatalf("no %s event in the log", typ)
	return record.Event{}
}

// payloadKeys is how the absent/present distinction is asserted: a flag the seat
// never passed must not appear in the event AT ALL.
func payloadKeys(ev record.Event) map[string]bool {
	out := map[string]bool{}
	for _, k := range ev.Payload.Keys() {
		out[k] = true
	}
	return out
}

// ---- preconditions shared by every verb ----

func TestEveryVerbRequiresRunAndSeatID(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"lens finding without --run", []string{"finding", "--seat-id", "red-lens-r1-L1"}, "lens: --run <runDir> is required"},
		// WITHOUT AN IDENTITY THERE IS NO TREE, so the refusal is about the identity rather
		// than about the verb. That is the contract now: the surface is scoped to whoever is
		// asking, and nothing can be shown — or refused — until that is answered.
		{"a verb with no identity at all", []string{"finding", "--run", "X"}, "--seat-id IS REQUIRED HERE"},
		// EACH CASE CARRIES THE VERB'S OTHER REQUIRED FLAGS. Cobra refuses a missing required
		// flag at PARSE, before Begin reaches the run and seat-id checks, so a case that omits
		// them measures whichever refusal fires first rather than the one it is named for.
		{"merge mint without --run", []string{"mint", "--seat-id", "red-merge-r1",
			"--check", "c", "--check-kind", "document", "--impact", "medium", "--likelihood", "medium",
			"--class", "x", "--problem", "p"}, "merge: --run <runDir> is required"},
		{"mint with no identity at all", []string{"mint", "--run", "X",
			"--check", "c", "--check-kind", "document", "--impact", "medium", "--likelihood", "medium",
			"--class", "x", "--problem", "p"}, "--seat-id IS REQUIRED HERE"},
		{"blue revision without --run", []string{"revision", "--seat-id", "blue-lane-1",
			"--reason", "what changed this round"}, "blue: --run <runDir> is required"},
		{"opinion with no identity at all", []string{"opinion", "--run", "X",
			"--id", "R1-1", "--as", "carried", "--principle", "p", "--tension", "t",
			"--review-flag", "no", "--reason", "r"}, "--seat-id IS REQUIRED HERE"},
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
			t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
			args := make([]string, len(tc.args))
			copy(args, tc.args)
			for i, a := range args {
				if a == "X" {
					args[i] = t.TempDir()
				}
			}
			_, err := run(t, args...)
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

// The verb set is the role boundary — and it is now the WHOLE of it. A lens seat may not write a
// board gap, and the reason is no longer that a runtime check compared its seat id against the
// role in the path: the verb is not in its tree. The claim this test holds is unchanged and the
// mechanism under it got simpler, so the refusal names the seat that does hold `mint` rather than
// the role the caller failed to be.
func TestRoleBindingIsEnforcedAtTheCLI(t *testing.T) {
	runDir := t.TempDir()
	_, err := run(t, "mint", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err == nil {
		t.Fatal("a LENS seat minted a board gap through the merge role")
	}
	if !strings.Contains(err.Error(), "the merge seat's verb") {
		t.Errorf("the refusal must name the seat that holds the verb: %v", err)
	}
	// And nothing was written under the crossing.
	if len(events(t, runDir)) != 0 {
		t.Errorf("a refused cross-role write still produced events: %+v", events(t, runDir))
	}

	// A seat id no dispatch created is refused too.
	_, err = run(t, "revision", "--run", runDir, "--seat-id", "hand-invented", "--reason", "t")
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
		{"lens", "mint"},
		{"lens", "close"},
		{"blue", "mint"},
		{"blue", "close"},
		{"bench", "mint"},
		{"merge", "revision"},
	}
	for _, tc := range cases {
		t.Run(tc.role+"/"+tc.verb, func(t *testing.T) {
			out, err := run(t, tc.verb, "--run", t.TempDir(), "--seat-id", record.SampleSeatOf(tc.role))
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
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
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

// The friction footer closes the loop the help opens: a missing capability is a
// finding about the tooling, not something to improvise around.
func TestRoleHelpCarriesTheFrictionFooter(t *testing.T) {
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		t.Run(role, func(t *testing.T) {
			out := help(t, "--help", "--seat-id", record.SampleSeatOf(role))
			if !strings.Contains(out, "it does not exist for you") {
				t.Errorf("%s help lacks the friction footer:\n%s", role, out)
			}
			if !strings.Contains(out, "'friction' verb") {
				t.Errorf("%s help does not name the friction channel:\n%s", role, out)
			}
		})
	}
}

// The board verbs exist in the merge role and NOWHERE else. This is the claim
// the whole engine rests on, so it is asserted over the actual command trees.
func TestBoardVerbsExistOnlyInTheMergeRole(t *testing.T) {
	// ASKED OF FOUR TREES, not of four groups in one. The claim is unchanged and is stronger for
	// it: a board verb is absent from a lens's SURFACE, not merely absent from a namespace the
	// lens could still address by typing it.
	verbs := map[string]map[string]bool{}
	for role, r := range AllRoots() {
		set := map[string]bool{}
		for _, v := range r.Commands() {
			set[v.Name()] = true
		}
		verbs[role] = set
	}
	for _, board := range []string{"mint", "close", "regrade"} {
		if !verbs["merge"][board] {
			t.Errorf("the merge role is missing the board verb %q", board)
		}
		for _, other := range []string{"lens", "blue", "bench"} {
			if verbs[other][board] {
				t.Errorf("the %s role has the board verb %q — the board would have more than one writer", other, board)
			}
		}
	}
	// Blue has NO board verbs at all, by topology rather than obedience.
	for _, v := range []string{"mint", "close", "regrade", "verdict", "spot-check"} {
		if verbs["blue"][v] {
			t.Errorf("blue has %q; blue is additive-only and must not be able to subtract", v)
		}
	}
	// Every role can show, and every role can register.
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
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
	runDir := t.TempDir()
	seedBlueReport(t, runDir)
	seatID := "red-lens-r1-L1"

	out, err := run(t, "register", "--run", runDir, "--seat-id", seatID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "registered "+seatID) || !strings.Contains(out, "shard nonce") {
		t.Errorf("register said %q", out)
	}

	out, err = run(t, "finding", "--run", runDir, "--seat-id", seatID,
		"--key", "F1", "--severity", "high", "--likelihood", "medium", "--impact", "high",
		"--quote", "§2", "--reason", "the finding prose")
	if err != nil {
		t.Fatal(err)
	}
	// The ID leads the message now: it is what the merge will name, and a seat told only
	// "recorded" has to invent a way to refer to this later.
	if !strings.Contains(out, "finding recorded:") || !strings.Contains(out, "L1-F1") {
		t.Errorf("finding said %q", out)
	}

	ev := lastOfType(t, runDir, "finding")
	if got := ev.Payload.Str("label"); got != "L1-F1" {
		t.Errorf("label = %q", got)
	}
	if got := ev.Payload.Str("severity"); got != "high" {
		t.Errorf("severity = %q", got)
	}
	if got := ev.Payload.Str("reason"); got != "the finding prose" {
		t.Errorf("text = %q", got)
	}
	if ev.Round != 1 {
		t.Errorf("round = %d, want 1 from the seat id", ev.Round)
	}
	if ev.Key != seatID+":finding:L1-F1" {
		t.Errorf("key = %q", ev.Key)
	}
}

// A flag the seat never passed must not appear in the event at all — the
// absent/present distinction the record format rests on.
func TestUnpassedFlagsAreAbsentFromThePayload(t *testing.T) {
	runDir := t.TempDir()
	seedBlueReport(t, runDir)
	seatID := "red-lens-r1-L1"
	// --quote and --reason are required (1b); likelihood/impact are still optional,
	// so they are the unpassed flags whose absence this pins.
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", seatID,
		"--key", "F1", "--severity", "high", "--quote", "§1", "--reason", "t"); err != nil {
		t.Fatal(err)
	}
	keys := payloadKeys(lastOfType(t, runDir, "finding"))
	if !keys["label"] || !keys["severity"] || !keys["reason"] {
		t.Errorf("a passed flag is missing from the payload: %v", keys)
	}
	for _, absent := range []string{"likelihood", "impact"} {
		if keys[absent] {
			t.Errorf("%q appears in the payload though the seat never passed it", absent)
		}
	}
}

// The CSV fields are the exception: ALWAYS present, even empty, because an
// absent key would read as "lineage unknown" where the truth is "lineage none".
func TestListFieldsAreAlwaysPresentEvenWhenEmpty(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	ev := lastOfType(t, runDir, "mint")
	keys := payloadKeys(ev)
	for _, k := range []string{"supersedes", "found_by"} {
		if !keys[k] {
			t.Errorf("%q is absent; an absent lineage key reads as \"lineage unknown\"", k)
		}
	}
	// And they serialize as arrays, not null.
	b, err := json.Marshal(ev.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"supersedes":[]`) {
		t.Errorf("empty lineage did not serialize as []: %s", b)
	}
}

// Grade validation refuses at the flag, before the payload is built, and the
// message names the values that would have worked.
func TestBadGradeIsRefusedAtParseTimeWithATeachingMessage(t *testing.T) {
	runDir := t.TempDir()
	_, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p", "--severity", "catastrophic")
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
	// Nothing was recorded: the refusal came before the write.
	if len(events(t, runDir)) != 0 {
		t.Errorf("a refused grade still produced events: %+v", events(t, runDir))
	}
}

// Ids are TOOL-assigned and sequential per round; --key makes a crash retry
// idempotent rather than double-minting.
func TestMintAssignsSequentialIdsAndIsIdempotentByKey(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	mint := func(extra ...string) string {
		t.Helper()
		args := append([]string{"mint", "--run", runDir, "--seat-id", seatID,
			"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}, extra...)
		out, err := run(t, args...)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(out)
	}
	if got := mint(); got != "minted R1-1" {
		t.Errorf("first mint said %q", got)
	}
	if got := mint("--key", "L1-F3"); got != "minted R1-2" {
		t.Errorf("second mint said %q", got)
	}
	// The retry: same command, same key, and the EXISTING id comes back.
	if got := mint("--key", "L1-F3"); got != "minted R1-2 (idempotent retry — existing id returned)" {
		t.Errorf("the retry said %q, want the existing id", got)
	}
	mints := 0
	for _, e := range events(t, runDir) {
		if e.Type == "mint" {
			mints++
		}
	}
	if mints != 2 {
		t.Errorf("%d mint events, want 2 — the retry double-minted", mints)
	}
	// A different round has its own id namespace.
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-merge-r2"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r2",
		"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "minted R2-1") {
		t.Errorf("round 2's first mint said %q, want R2-1", out)
	}
}

// --json is the machine contract: a success is a structured ack (read the id from a field,
// not by parsing prose), a failure is a structured error (branch on ok:false), and the
// default mode stays byte-identical prose because the template renders the same fields.
func TestJSONFlagStructuresResultsAndErrors(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seatID); err != nil {
		t.Fatal(err)
	}
	mintArgs := []string{"mint", "--run", runDir, "--seat-id", seatID,
		"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}

	out, err := run(t, append([]string{"--json"}, mintArgs...)...)
	if err != nil {
		t.Fatal(err)
	}
	var ok map[string]any
	if e := json.Unmarshal([]byte(strings.TrimSpace(out)), &ok); e != nil {
		t.Fatalf("mint --json is not valid JSON (%v): %s", e, out)
	}
	result, _ := ok["result"].(map[string]any)
	if ok["verb"] != "mint" || ok["ok"] != true || result["gap_id"] != "R1-1" {
		t.Errorf("mint --json = %v, want {verb:mint, ok:true, result:{gap_id:R1-1}}", ok)
	}

	// Default mode is the unchanged prose — the template reproduces it byte for byte.
	plain, err := run(t, append(mintArgs, "--key", "K2")...)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(plain) != "minted R1-2" {
		t.Errorf("default mint = %q, want unchanged prose 'minted R1-2'", plain)
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
		"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	var viol map[string]any
	if e := json.Unmarshal([]byte(strings.TrimSpace(rv)), &viol); e != nil {
		t.Fatalf("role-violation --json is not valid JSON (%v): %s", e, rv)
	}
	if viol["code"] != "role_violation" {
		t.Errorf("role-violation --json code = %v, want role_violation (env %v)", viol["code"], viol)
	}
}

// COINING IS ITS OWN VERB, and a mint against the coined slug records that this run coined it —
// derived from the registry rather than asserted by a boolean the seat sets.
func TestClassNewCoinsTheSlugInClass(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "class", "new", "--run", runDir, "--seat-id", seatID,
		"--class", "brand-new", "--definition", "d", "--neighbor", "n", "--distinguisher", "q"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "brand-new", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	m := lastOfType(t, runDir, "mint")
	if got := m.Payload.Str("class"); got != "brand-new" {
		t.Errorf("class = %q, want the slug from --class", got)
	}
	if v, _ := m.Payload.Get("class_new"); v != true {
		t.Errorf("class_new = %v, want true", v)
	}
	cn := lastOfType(t, runDir, "class-new")
	if got := cn.Payload.Str("slug"); got != "brand-new" {
		t.Errorf("class-new slug = %q", got)
	}
	for _, k := range []string{"definition", "neighbor", "distinguisher"} {
		if cn.Payload.Str(k) == "" {
			t.Errorf("the class-new event dropped %q", k)
		}
	}
}

// --problem and the prose channel are alternatives; --problem wins when both
// are given, and --file carries anything above trivial size.
func TestProseChannelResolution(t *testing.T) {
	// ONE trailing newline is stripped, and that is deliberate normalization rather than
	// mangling: a file's final newline is a line TERMINATOR that every editor and every
	// heredoc appends, not content the seat chose to record. Keeping it meant a payload
	// passed via --file ended with a stray blank line in every projection while the same
	// text passed via --text did not — one value recorded two ways depending on which
	// spelling the seat happened to use.
	t.Run("--file is read whole, less its terminating newline", func(t *testing.T) {
		runDir := t.TempDir()
		body := "line one\nline two — with unicode ✓ and <angle> brackets"
		f := filepath.Join(t.TempDir(), "prose.md")
		if err := os.WriteFile(f, []byte(body+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := run(t, "position", "--run", runDir, "--seat-id", "red-merge-r1", "--reason-file", f); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "position").Payload.Str("reason"); got != body {
			t.Errorf("text = %q, want the file's content without its terminator %q", got, body)
		}
	})

	// BOTH SPELLINGS IS NOW AN ERROR, not a precedence rule.
	//
	// This used to assert that --file silently wins. Silent precedence means a seat
	// passing both loses one payload without being told which, and the loss is invisible
	// until someone reads a projection rounds later and finds the wrong prose recorded.
	// Refusing costs the seat one turn and tells it exactly what to fix.
	t.Run("--file and --text together are refused, not ranked", func(t *testing.T) {
		runDir := t.TempDir()
		f := filepath.Join(t.TempDir(), "prose.md")
		if err := os.WriteFile(f, []byte("from the file"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := run(t, "position", "--run", runDir, "--seat-id", "red-merge-r1",
			"--reason-file", f, "--reason", "from the flag")
		if err == nil {
			t.Fatal("both spellings were accepted; one payload was silently dropped")
		}
		if !strings.Contains(err.Error(), "exactly one") {
			t.Errorf("the refusal must say which rule was broken, got: %v", err)
		}
	})

	t.Run("a missing --file is an error, not an empty payload", func(t *testing.T) {
		runDir := t.TempDir()
		_, err := run(t, "position", "--run", runDir, "--seat-id", "red-merge-r1",
			"--reason-file", filepath.Join(t.TempDir(), "no-such-file.md"))
		if err == nil {
			t.Fatal("a missing prose file was silently treated as empty")
		}
		if len(events(t, runDir)) != 0 {
			t.Errorf("a failed prose read still recorded an event: %+v", events(t, runDir))
		}
	})

	t.Run("neither channel is not a CHANNEL error", func(t *testing.T) {
		// THE TWO LAYERS ARE SEPARATE, and this is the one that distinguishes them. A missing
		// --reason-file is a channel failure (the case above); passing NEITHER flag is not —
		// the channel returns empty and the VERB decides whether empty is acceptable.
		//
		// Every prose verb now refuses an empty reason, because a duty discharged by nothing
		// still counts as discharged. So the refusal must come from the verb's rule and name
		// the flag, NOT from a failed read: a seat told "cannot read prose file" when it simply
		// omitted an argument goes looking for a file it never named.
		runDir := t.TempDir()
		out, err := run(t, "position", "--run", runDir, "--seat-id", "red-merge-r1")
		if err == nil {
			t.Fatal("a position with no reason was recorded — an empty position is a duty discharged by nothing")
		}
		// Cobra names the pair without dashes ("one of the flags in the group [reason
		// reason-file]"), which is the framework's phrasing and still names both ways in.
		all := out + err.Error()
		if !strings.Contains(all, "reason") {
			t.Errorf("the refusal does not name the flag that fixes it:\n%s", all)
		}
		if strings.Contains(all, "cannot read") || strings.Contains(all, "prose file") {
			t.Errorf("omitting both flags was reported as a FILE read failure, sending the seat after a file it never named:\n%s", all)
		}
		// The seat's own register event is expected; a POSITION event is not.
		for _, e := range events(t, runDir) {
			if e.Type == "position" {
				t.Errorf("a refused position was still recorded: %+v", e)
			}
		}
	})

	t.Run("--problem beats the prose channel on mint", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "from the flag", "--reason", "from the prose channel"); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "mint").Payload.Str("problem"); got != "from the flag" {
			t.Errorf("problem = %q, want --problem to win", got)
		}
	})

	t.Run("the prose channel fills problem when --problem is absent", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--reason", "from the prose channel"); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "mint").Payload.Str("problem"); got != "from the prose channel" {
			t.Errorf("problem = %q", got)
		}
	})
}

// close carries its evidence or it does not happen: the anchor triple is the
// difference between an auditable closure and an assertion.
func TestCloseRequiresItsAnchor(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}

	// --reason is supplied so the refusal under test is the ANCHOR one: cobra refuses a missing
	// required flag at parse, before the handler that checks the anchor ever runs.
	_, err := run(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--reason", "closed after verification")
	if err == nil {
		t.Fatal("an unanchored closure was accepted")
	}
	if !strings.Contains(err.Error(), "verified-by") {
		t.Errorf("error = %q", err)
	}

	// A partial verification is not a verification, and cobra says so at parse.
	_, err = run(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--reason", "closed after verification", "--verified-by", "L1")
	if err == nil {
		t.Fatal("a partial verification was accepted")
	}

	// The full triple closes it.
	out, err := run(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:f",
		"--reason", "verified against the ref")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "closed R1-1 (closed)") {
		t.Errorf("close said %q", out)
	}
	ev := lastOfType(t, runDir, "close")
	if got := ev.Payload.Str("closure_class"); got != "closed" {
		t.Errorf("closure_class = %q, want the default \"closed\"", got)
	}

	// Closing an unknown gap is refused before anything is written.
	_, err = run(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "R9-9",
		"--verified-by", "L1", "--verified-with", "t", "--verified-against", "x",
		"--reason", "supplied so the refusal under test is the reference one")
	if err == nil {
		t.Fatal("a closure of an unknown gap was accepted")
	}
	// THE REFUSAL MOVED EARLIER. `--id` is a typed flag carrying record.GapExists now, so
	// seat.Begin resolves it before the handler runs and the message is the reference one. That
	// is the point of moving the check to the flag: the seat is told the id names nothing,
	// rather than being told about an anchor for a gap that does not exist.
	if !strings.Contains(err.Error(), "names gap R9-9") {
		t.Errorf("error = %q", err)
	}
}

func TestCloseWithRegressionRequiresASuccessor(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	// The successor must EXIST: a successor is a reference like any other and is checked
	// at write time now, so the fixture mints the gap it will name.
	succOut, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--key", "successor-gap", "--class", "x", "--check-kind", "document", "--check", "c",
		"--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	successor := regexp.MustCompile(`R\d+-\d+`).FindString(succOut)
	if successor == "" {
		t.Fatalf("could not read the successor id from %q", succOut)
	}
	base := []string{"close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--verified-by", "L1", "--verified-with", "t", "--verified-against", "x",
		"--reason", "verified", "--as", "closed_with_regression"}
	if _, err := run(t, base...); err == nil {
		t.Fatal("closed_with_regression was accepted without a successor — lineage dropped")
	}
	if _, err := run(t, append(base, "--superseded-by", successor)...); err != nil {
		t.Fatalf("a successor'd regression close was refused: %v", err)
	}
}

// close --file carries the full closure record; a missing one is an error.
func TestCloseFile(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "closure.md")
	if err := os.WriteFile(f, []byte("the whole closure record"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--verified-by", "L1", "--verified-with", "t", "--verified-against", "x", "--reason-file", f); err != nil {
		t.Fatal(err)
	}
	if got := lastOfType(t, runDir, "close").Payload.Str("reason"); got != "the whole closure record" {
		t.Errorf("prose = %q", got)
	}

	_, err := run(t, "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--verified-by", "L1", "--verified-with", "t", "--verified-against", "x",
		"--reason-file", filepath.Join(t.TempDir(), "gone.md"))
	if err == nil {
		t.Fatal("a missing --file was ignored")
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
		{"regrade without --reason", []string{"regrade", "--id", "R1-1", "--severity", "high"}, "reason"},
		{"mint without --check", []string{"mint", "--class", "x", "--problem", "p"}, "check"},
		{"mint without --class", []string{"mint", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}, "class"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			seedBlueReport(t, runDir)
			seatID := "red-merge-r1"
			// The referenced gap and observation must EXIST, or the reference check
			// fires first and this test asserts on the wrong refusal — it is about the
			// missing REASON, not a missing referent.
			if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
				"--key", "k", "--class", "x", "--check-kind", "document", "--check", "c",
				"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
			if _, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
				"--key", "F1", "--quote", "l", "--reason", "t",
				"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
				t.Fatal(err)
			}
			args := append([]string{tc.args[0], tc.args[1], "--run", runDir, "--seat-id", seatID}, tc.args[2:]...)
			_, err := run(t, args...)
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
	runDir := t.TempDir()
	seatID := "blue-lane-1"

	t.Run("retire needs both the claim and the reason", func(t *testing.T) {
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", seatID, "--reason", "refuted"); err == nil {
			t.Error("a retirement with no quoted claim was accepted")
		}
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", seatID, "--quote", "the claim"); err == nil {
			t.Error("a retirement with no reason was accepted")
		}
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", seatID,
			"--quote", "the claim as it stood", "--reason", "refuted"); err != nil {
			t.Errorf("a complete retirement was refused: %v", err)
		}
	})

	t.Run("a line of inquiry is proposed, then moved with what changed", func(t *testing.T) {
		out, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seatID,
			"--reason", "the road taken", "--hypothesis", "it leads somewhere")
		if err != nil {
			t.Fatalf("a proposal needs no fate: %v", err)
		}
		id := inquiryIDOf(out)
		if _, err := run(t, "line-of-inquiry", "move", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--as", "declined"); err == nil {
			t.Error("a move with no reason was accepted — a status that slides silently is decoration")
		}
		if _, err := run(t, "line-of-inquiry", "move", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--as", "invented", "--reason", "r"); err == nil {
			t.Error("an undefined line of inquiry status was accepted; the render groups BY status")
		}
	})

	t.Run("revision records its prose (claim_count moved to count-claims, #70)", func(t *testing.T) {
		if _, err := run(t, "revision", "--run", runDir, "--seat-id", seatID,
			"--reason", "what changed"); err != nil {
			t.Fatal(err)
		}
		ev := lastOfType(t, runDir, "revision")
		if got := ev.Payload.Str("reason"); got != "what changed" {
			t.Errorf("text = %q", got)
		}
	})
}

// The bench rules with reasons or it does not rule: every one of the five
// fields is required, and the refusal names the missing flag.
func TestBenchOpinionRequiresAllFiveFields(t *testing.T) {
	runDir := t.TempDir()
	seatID := "judge-r1"
	// The gap must EXIST: --id is a reference and is checked at write time, so without
	// a real gap this test would assert on the dangling-reference refusal instead of the
	// missing-field one it is about.
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "k", "--class", "x", "--check-kind", "document", "--check", "c",
		"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	full := map[string]string{
		"id": "R1-1", "as": "carried", "principle": "correctness first",
		"tension": "correctness vs economy", "review-flag": "no",
	}
	for missing := range full {
		t.Run("missing --"+missing, func(t *testing.T) {
			args := []string{"opinion", "--run", runDir, "--seat-id", seatID}
			for k, v := range full {
				if k != missing {
					args = append(args, "--"+k, v)
				}
			}
			_, err := run(t, args...)
			if err == nil {
				t.Fatalf("an opinion was accepted without --%s", missing)
			}
			// The refusal NAMES the missing flag. What the field is for lives in
			// `bench opinion --help`, which the seat is required to have read before running
			// the command — restating it in the error would be a second copy of the same text,
			// free to drift from the one cobra generates.
			if !strings.Contains(err.Error(), missing) {
				t.Errorf("the refusal does not name --%s: %v", missing, err)
			}
		})
	}
	args := []string{"opinion", "--run", runDir, "--seat-id", seatID, "--reason-file", writeTemp(t, "the rationale")}
	for k, v := range full {
		args = append(args, "--"+k, v)
	}
	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("a complete opinion was refused: %v", err)
	}
	if !strings.Contains(out, "opinion recorded: R1-1 carried") {
		t.Errorf("opinion said %q", out)
	}
	if got := lastOfType(t, runDir, "opinion").Payload.Str("reason"); got != "the rationale" {
		t.Errorf("rationale = %q", got)
	}
}

// The shared verbs are ONE contract wherever they appear: a friction entry from a
// lens and one from the bench are the same event with the same payload.
func TestSharedVerbsRecordTheSameEventFromEveryRole(t *testing.T) {
	cases := []struct{ role, seatID string }{
		{"lens", "red-lens-r1-L1"},
		{"merge", "red-merge-r1"},
		{"blue", "blue-lane-1"},
		{"bench", "judge-r1"},
	}
	for _, tc := range cases {
		t.Run("friction/"+tc.role, func(t *testing.T) {
			runDir := t.TempDir()
			out, err := run(t, "friction", "--run", runDir, "--seat-id", tc.seatID,
				"--reason", "the capability I needed")
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(out) != "friction recorded" {
				t.Errorf("friction said %q", out)
			}
			ev := lastOfType(t, runDir, "friction")
			if got := ev.Payload.Str("reason"); got != "the capability I needed" {
				t.Errorf("text = %q", got)
			}
			if keys := payloadKeys(ev); len(keys) != 1 || !keys["reason"] {
				t.Errorf("the friction payload is not just the reason: %v", keys)
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
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	// Both gaps must EXIST: a closing names the gap it argues, and that reference is
	// checked at write time.
	for i := 0; i < 2; i++ {
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
			"--key", fmt.Sprintf("k%d", i), "--class", "x", "--check-kind", "document", "--check", "c",
			"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"R1-1", "R1-2"} {
		if _, err := run(t, "closing", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--reason", "argued "+id); err != nil {
			t.Fatal(err)
		}
	}
	closings := 0
	for _, e := range events(t, runDir) {
		if e.Type == "closing" {
			closings++
		}
	}
	if closings != 2 {
		t.Errorf("%d closings survived, want 2 — one was dedup'd", closings)
	}
}

// position is a SINGLETON: a seat has one position, so a re-run replaces rather
// than appending a second.
func TestPositionIsASingletonPerSeat(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	for _, text := range []string{"first", "second"} {
		if _, err := run(t, "position", "--run", runDir, "--seat-id", seatID, "--reason", text); err != nil {
			t.Fatal(err)
		}
	}
	positions := 0
	for _, e := range events(t, runDir) {
		if e.Type == "position" {
			positions++
			if got := e.Payload.Str("reason"); got != "first" {
				t.Errorf("the surviving position is %q, want the first", got)
			}
		}
	}
	if positions != 1 {
		t.Errorf("%d positions survived the merge, want 1", positions)
	}
}

// A PASS is a claim that nothing is left open; the tool refuses one over an open board —
// the 2026-07-20 rubber-stamp (PASS with 9 open gaps) made structurally impossible. FAIL is
// always allowed.
func TestVerdictPASSRefusedOverOpenGaps(t *testing.T) {
	seatID := "red-merge-r1"
	mint2 := func(runDir string) {
		for i := 0; i < 2; i++ {
			if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
				"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
		}
	}

	failDir := t.TempDir()
	mint2(failDir)
	if _, err := run(t, "verdict", "--run", failDir, "--seat-id", seatID, "--as", "FAIL"); err != nil {
		t.Fatalf("FAIL must be allowed with open gaps: %v", err)
	}

	runDir := t.TempDir()
	mint2(runDir)
	_, err := run(t, "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS")
	if err == nil {
		t.Fatal("PASS was recorded over 2 open gaps — the rubber-stamp the guard exists to stop")
	}
	for _, want := range []string{"R1-1", "R1-2", "PASS refused"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the open gaps, got: %v", err)
		}
	}
	for _, id := range []string{"R1-1", "R1-2"} {
		if _, err := run(t, "close", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--as", "closed",
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
	seatID := "red-merge-r1"
	for _, as := range []string{"pass", "Pass", "PASSED", "banana", ""} {
		t.Run("as="+as, func(t *testing.T) {
			runDir := t.TempDir()
			if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
				"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
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
				if e.Type == "verdict" {
					t.Errorf("a refused verdict reached the log anyway: %+v", e.Payload)
				}
			}
		})
	}
}

// verdict is the merge seat's terminal act: it renders and checkpoints.
func TestVerdictRendersAndCheckpoints(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	// A PASS is refused over an open gap, so close it first (the guard is exercised in its
	// own test); this test is about render + checkpoint on a legitimate PASS.
	if _, err := run(t, "close", "--run", runDir, "--seat-id", seatID,
		"--id", "R1-1", "--as", "closed",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "resolved"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS")
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
	mirror := strings.TrimSpace(out[strings.Index(out, "checkpointed to ")+len("checkpointed to "):])
	entries, rerr := os.ReadDir(mirror)
	if rerr != nil {
		t.Fatalf("the checkpoint mirror is not readable: %v", rerr)
	}
	var shards int
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "events-") {
			shards++
		}
	}
	if shards == 0 {
		t.Errorf("the mirror holds no shards: %v", entries)
	}
	// The verdict itself is on the record.
	if got := lastOfType(t, runDir, "verdict").Payload.Str("verdict"); got != "PASS" {
		t.Errorf("verdict payload = %q", got)
	}
}

// A verb takes no positional arguments: a seat that types prose without --text
// must be told, not have it silently dropped.
func TestVerbsRefusePositionalArguments(t *testing.T) {
	runDir := t.TempDir()
	_, err := run(t, "position", "--run", runDir, "--seat-id", "red-merge-r1", "stray prose")
	if err == nil {
		t.Fatal("a positional argument was silently dropped")
	}
	if len(events(t, runDir)) != 0 {
		t.Errorf("a refused invocation still recorded: %+v", events(t, runDir))
	}
}

// The root answers --version with the stamped version, and that version is what
// register writes into the record.
func TestVersionIsStampedOnTheFirstAct(t *testing.T) {
	runDir := t.TempDir()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatal(err)
	}
	reg := lastOfType(t, runDir, "register")
	if got := reg.Payload.Str("tool_version"); got != Version {
		t.Errorf("tool_version = %q, want %q", got, Version)
	}
	if record.ToolVersion != Version {
		t.Errorf("record.ToolVersion = %q, want the CLI's %q — a skewed binary would go unnoticed", record.ToolVersion, Version)
	}
}

func writeTemp(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "prose.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// boardState is the replayed board, for assertions about what the events MEAN
// rather than what they say.
func boardState(t *testing.T, runDir string) (*record.Board, error) {
	t.Helper()
	return record.BoardState(runDir)
}

// The other half of the contract: when a run IS live, a seat that forgot --run is
// served rather than refused. Ten of the first live run's 55 tool-call errors were
// this flag alone, and it cannot be carried in the environment — shell state does not
// persist between tool calls and every subagent shares the parent's session id.
func TestRunDirIsInferredFromTheLiveMarkerWhenTheFlagIsOmitted(t *testing.T) {
	proj := t.TempDir()
	runDir := filepath.Join(proj, "research", "live-run")
	if err := os.MkdirAll(filepath.Join(runDir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(proj, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, ".claude", "run-live.json"),
		[]byte(`{"runDir":"research/live-run"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", proj)

	// No --run. The precondition must pass, so any error is about the verb's own
	// flags rather than the missing run directory.
	_, err := run(t, "register", "--seat-id", "red-lens-r1-L1")
	if err != nil && strings.Contains(err.Error(), "--run <runDir> is required") {
		t.Fatalf("run dir was not inferred: %v", err)
	}
}

// An explicit --run always beats the marker: inference is a fallback for the seat
// that forgot, never a second source of truth that can override what it was told.
func TestExplicitRunDirBeatsTheInferredOne(t *testing.T) {
	proj := t.TempDir()
	marker := filepath.Join(proj, "research", "marker-run")
	explicit := filepath.Join(proj, "research", "explicit-run")
	for _, d := range []string{marker, explicit, filepath.Join(proj, ".claude")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(proj, ".claude", "run-live.json"),
		[]byte(`{"runDir":"research/marker-run"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", proj)

	if got := seat.InferRunDir(proj); got != marker {
		t.Fatalf("precondition: marker should resolve to %q, got %q", marker, got)
	}
	// With the flag present the marker must not be consulted at all.
	if _, err := run(t, "register", "--run", explicit, "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatalf("explicit --run should be honoured: %v", err)
	}
	if _, err := os.Stat(filepath.Join(explicit, "records")); err != nil {
		t.Fatalf("events should have landed in the explicit run dir: %v", err)
	}
}

// The W1.8 spot-check floor keys on the archive's state at round START, so a round
// entering with zero records has nothing to sample. --ids demanded a list, so this run's
// red-merge-r1 could record the discharge only as prose in the ledger — which a later
// audit has to take on trust. An unrecordable discharge is indistinguishable from a
// skipped duty, which is precisely what the event stream exists to prevent.
func TestSpotCheckRecordsAnHonestlyEmptyRound(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--none", "--reason", "archive empty at round start; floor not applicable")
	if err != nil {
		t.Fatalf("an empty archive must be recordable: %v", err)
	}
	if !strings.Contains(out, "nothing to sample") {
		t.Errorf("output %q should say the discharge was empty", out)
	}
	ev := lastOfType(t, runDir, "spot-check")
	keys := payloadKeys(ev)
	if !keys["none"] || !keys["reason"] {
		t.Errorf("the event must carry both the empty marker and its reason; got %v", keys)
	}
}

func TestSpotCheckRefusesAnEmptyDischargeWithNoReason(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	// COBRA REFUSES IT AT PARSE now, because the collapse to one prose channel made --reason
	// unconditionally required: what a sample found and why there was nothing to sample were
	// always the same field, and only the branch that wrote it differed.
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1", "--none"); err == nil ||
		!strings.Contains(err.Error(), "reason") {
		t.Fatalf("--none with no reason must be refused, got %v", err)
	}
}

func TestSpotCheckRefusesContradictoryFlags(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--none", "--reason", "x", "--ids", "R1-4"); err == nil || !strings.Contains(err.Error(), "none") {
		t.Fatalf("claiming both nothing-to-sample and a sample must be refused, got %v", err)
	}
}

// A spot-check with no sample still records an empty array — and now also records WHY, because
// the reason is required in both forms. The bare form used to be legal and recorded an empty
// array with no account of it, so "the archive was empty at round start" and "the seat skipped
// its duty" were the same event.
func TestBareSpotCheckStillRecordsAnEmptyArray(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--reason", "nothing was sampled this round"); err != nil {
		t.Fatalf("the no-sample form must keep working: %v", err)
	}
	if keys := payloadKeys(lastOfType(t, runDir, "spot-check")); !keys["ids"] {
		t.Error("ids must still be present as an empty array")
	}
}

// `merge close` called its payload --prose-file while every other prose-bearing verb
// calls it --file. Seats typed --file here and were refused, twice, in one run. Post-
// collapse the shared word is --reason-file, and this pins that close reads it.
func TestCloseAcceptsTheSharedPayloadFlagName(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	prose := filepath.Join(t.TempDir(), "closure.md")
	if err := os.WriteFile(prose, []byte("verified at the leaf; digits match the cited arm"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	minted, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "some-class",
		"--problem", "p", "--fix", "f", "--check-kind", "document", "--check", "c", "--likelihood", "medium", "--impact", "medium",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low")
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	id := regexp.MustCompile(`R\d+-\d+`).FindString(minted)
	if id == "" {
		t.Fatalf("could not read the minted id from %q", minted)
	}
	// A REAL ANCHOR, not --carried-from. This test's subject is --file, and it used the
	// carry as a shortcut past the anchor requirement — which is exactly how a seat under
	// pressure would use it, and why the carry is now checked against a real earlier
	// closure rather than accepted on its say-so.
	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "closed",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason-file", prose); err != nil {
		t.Fatalf("--file must work on close, as it does on every other prose verb: %v", err)
	}
	if !payloadKeys(lastOfType(t, runDir, "close"))["reason"] {
		t.Error("the prose from --file must reach the event")
	}
}

// inquiryIDOf pulls the tool-assigned line-of-inquiry id out of a propose result.
func inquiryIDOf(out string) string {
	m := regexp.MustCompile(`\b(Q\d+)\b`).FindStringSubmatch(out)
	if m == nil {
		return ""
	}
	return m[1]
}
