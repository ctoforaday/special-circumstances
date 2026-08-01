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

// run executes one command against a fresh root and captures what the seat sees.
// Verb output goes to real stdout via fmt.Println, so stdout is redirected
// rather than taken from cobra's writer.
func run(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	// Output goes through cmd.OutOrStdout(), so a buffer set on the root captures every
	// verb's stdout — no os.Stdout swap, no pipe. Errors are returned (SilenceErrors), so
	// stderr carries nothing the tests read.
	root := newRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err = root.Execute()
	return out.String(), err
}

// help captures a command's own help output, which cobra writes to its writer.
func help(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	root := newRoot()
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
		{"lens finding without --run", []string{"lens", "finding", "--seat-id", "red-lens-r1-L1"}, "lens: --run <runDir> is required"},
		{"lens finding without --seat-id", []string{"lens", "finding", "--run", "X"}, "lens: --seat-id is required"},
		{"merge mint without --run", []string{"merge", "mint", "--seat-id", "red-merge-r1"}, "merge: --run <runDir> is required"},
		{"merge mint without --seat-id", []string{"merge", "mint", "--run", "X"}, "merge: --seat-id is required"},
		{"blue revision without --run", []string{"blue", "revision", "--seat-id", "blue-lane-1"}, "blue: --run <runDir> is required"},
		{"bench opinion without --seat-id", []string{"bench", "opinion", "--run", "X"}, "bench: --seat-id is required"},
		{"register is not exempt", []string{"lens", "register", "--run", "X"}, "lens: --seat-id is required"},
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
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// The verb set is the role boundary, and seat identity is bound to its
// namespace: a lens seat may not write through the merge role even though the
// merge role has the verb.
func TestRoleBindingIsEnforcedAtTheCLI(t *testing.T) {
	runDir := t.TempDir()
	_, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err == nil {
		t.Fatal("a LENS seat minted a board gap through the merge role")
	}
	if !strings.Contains(err.Error(), "belongs to the lens role") {
		t.Errorf("the refusal must name the seat's real role: %v", err)
	}
	// And nothing was written under the crossing.
	if len(events(t, runDir)) != 0 {
		t.Errorf("a refused cross-role write still produced events: %+v", events(t, runDir))
	}

	// A seat id no dispatch created is refused too.
	_, err = run(t, "blue", "revision", "--run", runDir, "--seat-id", "hand-invented", "--reason", "t")
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
			_, err := run(t, tc.role, tc.verb, "--run", t.TempDir(), "--seat-id", "x")
			if err == nil {
				t.Fatalf("%s ran the %s verb", tc.role, tc.verb)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("verb %q is outside this seat's role", tc.verb)) {
				t.Errorf("error = %q, want it to name the verb as out of role", err)
			}
			if !strings.Contains(err.Error(), "available:") || !strings.Contains(err.Error(), "show") {
				t.Errorf("the refusal must list what IS available: %q", err)
			}
		})
	}
}

// A role invoked with no verb is a usage error, not a no-op: silently succeeding
// would let a mis-scripted seat believe it recorded something.
func TestARoleWithNoVerbIsAnError(t *testing.T) {
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		t.Run(role, func(t *testing.T) {
			_, err := run(t, role)
			if err == nil {
				t.Fatalf("%s with no verb succeeded", role)
			}
			if !strings.Contains(err.Error(), "a verb is required") {
				t.Errorf("error = %q", err)
			}
		})
	}
}

// The friction footer closes the loop the help opens: a missing capability is a
// finding about the tooling, not something to improvise around.
func TestRoleHelpCarriesTheFrictionFooter(t *testing.T) {
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		t.Run(role, func(t *testing.T) {
			out := help(t, role, "--help")
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
	root := newRoot()
	verbs := map[string]map[string]bool{}
	for _, roleCmd := range root.Commands() {
		set := map[string]bool{}
		for _, v := range roleCmd.Commands() {
			set[v.Name()] = true
		}
		verbs[roleCmd.Name()] = set
	}
	for _, board := range []string{"mint", "close", "dispose", "regrade"} {
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
	for _, v := range []string{"mint", "close", "dispose", "regrade", "verdict", "spot-check"} {
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
	seatID := "red-lens-r1-L1"

	out, err := run(t, "lens", "register", "--run", runDir, "--seat-id", seatID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "registered "+seatID) || !strings.Contains(out, "shard nonce") {
		t.Errorf("register said %q", out)
	}

	out, err = run(t, "lens", "finding", "--run", runDir, "--seat-id", seatID,
		"--key", "F1", "--severity", "high", "--likelihood", "medium", "--impact", "high",
		"--location", "§2", "--reason", "the finding prose")
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
	if got := ev.Payload.Str("text"); got != "the finding prose" {
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
	seatID := "red-lens-r1-L1"
	if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", seatID,
		"--key", "F1", "--severity", "high", "--reason", "t"); err != nil {
		t.Fatal(err)
	}
	keys := payloadKeys(lastOfType(t, runDir, "finding"))
	if !keys["label"] || !keys["severity"] || !keys["text"] {
		t.Errorf("a passed flag is missing from the payload: %v", keys)
	}
	for _, absent := range []string{"likelihood", "impact", "location"} {
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
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "scope-creep", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
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
	_, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p", "--severity", "catastrophic")
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
		args := append([]string{"merge", "mint", "--run", runDir, "--seat-id", seatID,
			"--class", "scope-creep", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}, extra...)
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
	if _, err := run(t, "merge", "register", "--run", runDir, "--seat-id", "red-merge-r2"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r2",
		"--class", "scope-creep", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
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
	if _, err := run(t, "merge", "register", "--run", runDir, "--seat-id", seatID); err != nil {
		t.Fatal(err)
	}
	mintArgs := []string{"merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "scope-creep", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}

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
	bad, _ := run(t, "--json", "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "NOPE")
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
	rv, _ := run(t, "--json", "merge", "mint", "--run", runDir, "--seat-id", "blue-synthesize",
		"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	var viol map[string]any
	if e := json.Unmarshal([]byte(strings.TrimSpace(rv)), &viol); e != nil {
		t.Fatalf("role-violation --json is not valid JSON (%v): %s", e, rv)
	}
	if viol["code"] != "role_violation" {
		t.Errorf("role-violation --json code = %v, want role_violation (env %v)", viol["code"], viol)
	}
}

// --class-new both names the class and mints it, so it wins over --class and
// emits a second event registering the slug.
func TestMintClassNewWinsOverClassAndRecordsTheSlug(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	_, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "ignored", "--class-new", "brand-new",
		"--definition", "d", "--neighbor", "n", "--distinguisher", "q",
		"--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	m := lastOfType(t, runDir, "mint")
	if got := m.Payload.Str("class"); got != "brand-new" {
		t.Errorf("class = %q, want the --class-new slug", got)
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

// --existence writes the leaf-check axis to the board — the write path the board was
// reading (viewjson g.Mint existence) before anything set it. verified|suspected are the
// only legal values; a value outside the enum is refused at parse time (one way, a wrong
// guess errors, never a silent empty axis).
func TestMintExistenceLandsOnBoardAndRejectsBadValue(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "correctness", "--check", "c", "--likelihood", "medium", "--impact", "medium",
		"--problem", "p", "--existence", "verified"); err != nil {
		t.Fatal(err)
	}
	if got := lastOfType(t, runDir, "mint").Payload.Str("existence"); got != "verified" {
		t.Errorf("existence = %q, want verified (the write path must land it on the board)", got)
	}
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "correctness", "--check", "c", "--likelihood", "low", "--impact", "low",
		"--problem", "p2", "--existence", "suspected"); err != nil {
		t.Fatalf("suspected must be accepted: %v", err)
	}
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "correctness", "--check", "c", "--likelihood", "low", "--impact", "low",
		"--problem", "p3", "--existence", "probable"); err == nil {
		t.Error("--existence probable must be rejected (enum is verified|suspected)")
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
		if _, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1", "--reason-file", f); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "position").Payload.Str("text"); got != body {
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
		_, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1",
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
		_, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1",
			"--reason-file", filepath.Join(t.TempDir(), "no-such-file.md"))
		if err == nil {
			t.Fatal("a missing prose file was silently treated as empty")
		}
		if len(events(t, runDir)) != 0 {
			t.Errorf("a failed prose read still recorded an event: %+v", events(t, runDir))
		}
	})

	t.Run("neither channel yields empty text, not an error", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
			t.Fatal(err)
		}
		ev := lastOfType(t, runDir, "position")
		if got := ev.Payload.Str("text"); got != "" {
			t.Errorf("text = %q, want empty", got)
		}
		// text is set unconditionally on a prose verb, so the key is present.
		if !payloadKeys(ev)["text"] {
			t.Error("the text key is absent from a prose verb's payload")
		}
	})

	t.Run("--problem beats the prose channel on mint", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "from the flag", "--reason", "from the prose channel"); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "mint").Payload.Str("problem"); got != "from the flag" {
			t.Errorf("problem = %q, want --problem to win", got)
		}
	})

	t.Run("the prose channel fills problem when --problem is absent", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--reason", "from the prose channel"); err != nil {
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
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}

	_, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1")
	if err == nil {
		t.Fatal("an unanchored closure was accepted")
	}
	if !strings.Contains(err.Error(), "attestation anchor") {
		t.Errorf("error = %q", err)
	}

	// A partial anchor is not an anchor.
	_, err = run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1")
	if err == nil {
		t.Fatal("a partial anchor was accepted")
	}

	// The full triple closes it.
	out, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1", "--anchor-tool", "git show", "--anchor-target", "7bc501e:f",
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
	_, err = run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R9-9",
		"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x")
	if err == nil {
		t.Fatal("a closure of an unknown gap was accepted")
	}
	if !strings.Contains(err.Error(), "close of unknown gap") {
		t.Errorf("error = %q", err)
	}
}

func TestCloseWithRegressionRequiresASuccessor(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	// The successor must EXIST: a successor is a reference like any other and is checked
	// at write time now, so the fixture mints the gap it will name.
	succOut, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--key", "successor-gap", "--class", "x", "--check", "c",
		"--likelihood", "medium", "--impact", "medium", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	successor := regexp.MustCompile(`R\d+-\d+`).FindString(succOut)
	if successor == "" {
		t.Fatalf("could not read the successor id from %q", succOut)
	}
	base := []string{"merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x",
		"--reason", "verified", "--as", "closed_with_regression"}
	if _, err := run(t, base...); err == nil {
		t.Fatal("closed_with_regression was accepted without a successor — lineage dropped")
	}
	if _, err := run(t, append(base, "--successor", successor)...); err != nil {
		t.Fatalf("a successor'd regression close was refused: %v", err)
	}
}

// close --file carries the full closure record; a missing one is an error.
func TestCloseFile(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "closure.md")
	if err := os.WriteFile(f, []byte("the whole closure record"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x", "--reason-file", f); err != nil {
		t.Fatal(err)
	}
	if got := lastOfType(t, runDir, "close").Payload.Str("prose"); got != "the whole closure record" {
		t.Errorf("prose = %q", got)
	}

	_, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x",
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
		{"regrade without --reason", []string{"merge", "regrade", "--id", "R1-1", "--severity", "high"}, "regrade requires --reason"},
		{"dispose without --as", []string{"merge", "dispose", "--observation", "L1-F1"}, "dispose requires --as"},
		{"mint without --check", []string{"merge", "mint", "--class", "x", "--problem", "p"}, "mint requires --check"},
		{"mint without --class", []string{"merge", "mint", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}, "mint requires --class"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			seatID := "red-merge-r1"
			// The referenced gap and observation must EXIST, or the reference check
			// fires first and this test asserts on the wrong refusal — it is about the
			// missing REASON, not a missing referent.
			if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
				"--key", "k", "--class", "x", "--check", "c",
				"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
			if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
				"--key", "F1", "--location", "l", "--reason", "t",
				"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
				t.Fatal(err)
			}
			args := append([]string{tc.args[0], tc.args[1], "--run", runDir, "--seat-id", seatID}, tc.args[2:]...)
			_, err := run(t, args...)
			if err == nil {
				t.Fatal("the verb was accepted without its reason")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantErr)
			}
			// The role leads the message, so the seat knows which contract it broke.
			if !strings.HasPrefix(err.Error(), "merge:") {
				t.Errorf("the error does not lead with the role: %q", err)
			}
		})
	}
}

// blue's verbs refuse the removals and empty avenues the engine cares about.
func TestBlueVerbContracts(t *testing.T) {
	runDir := t.TempDir()
	seatID := "blue-lane-1"

	t.Run("retire needs both the claim and the reason", func(t *testing.T) {
		if _, err := run(t, "blue", "retire", "--run", runDir, "--seat-id", seatID, "--reason", "refuted"); err == nil {
			t.Error("a retirement with no quoted claim was accepted")
		}
		if _, err := run(t, "blue", "retire", "--run", runDir, "--seat-id", seatID, "--claim", "the claim"); err == nil {
			t.Error("a retirement with no reason was accepted")
		}
		if _, err := run(t, "blue", "retire", "--run", runDir, "--seat-id", seatID,
			"--claim", "the claim as it stood", "--reason", "refuted"); err != nil {
			t.Errorf("a complete retirement was refused: %v", err)
		}
	})

	t.Run("a declined avenue needs its reason", func(t *testing.T) {
		if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", seatID,
			"--status", "declined", "--line", "the road not taken"); err == nil {
			t.Error("a declined avenue with no reason was accepted — that is decoration")
		}
		if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", seatID,
			"--status", "pursued", "--line", "the road taken"); err != nil {
			t.Errorf("a pursued avenue needs no reason: %v", err)
		}
		if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", seatID,
			"--status", "invented", "--line", "l", "--reason", "r"); err == nil {
			t.Error("an undefined avenue status was accepted; the render groups BY status")
		}
	})

	t.Run("revision records its prose (claim_count moved to count-claims, #70)", func(t *testing.T) {
		if _, err := run(t, "blue", "revision", "--run", runDir, "--seat-id", seatID,
			"--reason", "what changed"); err != nil {
			t.Fatal(err)
		}
		ev := lastOfType(t, runDir, "revision")
		if got := ev.Payload.Str("text"); got != "what changed" {
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
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "k", "--class", "x", "--check", "c",
		"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	full := map[string]string{
		"id": "R1-1", "as": "carried", "principle": "correctness first",
		"tension": "correctness vs economy", "review-flag": "no",
	}
	for missing := range full {
		t.Run("missing --"+missing, func(t *testing.T) {
			args := []string{"bench", "opinion", "--run", runDir, "--seat-id", seatID}
			for k, v := range full {
				if k != missing {
					args = append(args, "--"+k, v)
				}
			}
			_, err := run(t, args...)
			if err == nil {
				t.Fatalf("an opinion was accepted without --%s", missing)
			}
			if !strings.Contains(err.Error(), "opinions, not dispositions") {
				t.Errorf("the refusal must say why: %v", err)
			}
		})
	}
	args := []string{"bench", "opinion", "--run", runDir, "--seat-id", seatID, "--reason-file", writeTemp(t, "the rationale")}
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
	if got := lastOfType(t, runDir, "opinion").Payload.Str("rationale"); got != "the rationale" {
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
			out, err := run(t, tc.role, "friction", "--run", runDir, "--seat-id", tc.seatID,
				"--reason", "the capability I needed")
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(out) != "friction recorded" {
				t.Errorf("friction said %q", out)
			}
			ev := lastOfType(t, runDir, "friction")
			if got := ev.Payload.Str("text"); got != "the capability I needed" {
				t.Errorf("text = %q", got)
			}
			if keys := payloadKeys(ev); len(keys) != 1 || !keys["text"] {
				t.Errorf("the friction payload is not just text: %v", keys)
			}
		})
	}

	// petition records the same event across the seats that can file it. Only merge
	// and blue carry the verb (cases[1:3]); lens surfaces findings, it does not petition.
	for _, tc := range cases[1:3] {
		t.Run("petition/"+tc.role, func(t *testing.T) {
			runDir := t.TempDir()
			out, err := run(t, tc.role, "petition", "--run", runDir, "--seat-id", tc.seatID,
				"--petition-class", "safety", "--reason", "what happened", "--relief", "the relief sought")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, "petition filed (safety)") {
				t.Errorf("petition said %q", out)
			}
			ev := lastOfType(t, runDir, "petition")
			for k, want := range map[string]string{"class": "safety", "basis": "what happened", "relief": "the relief sought"} {
				if got := ev.Payload.Str(k); got != want {
					t.Errorf("%s = %q, want %q", k, got, want)
				}
			}
		})
	}
}

// closing is keyed on the gap it argues, so two closings in one seat do not
// collide and neither is dedup'd away.
func TestClosingIsKeyedPerGap(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	// Both gaps must EXIST: a closing names the gap it argues, and that reference is
	// checked at write time.
	for i := 0; i < 2; i++ {
		if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
			"--key", fmt.Sprintf("k%d", i), "--class", "x", "--check", "c",
			"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"R1-1", "R1-2"} {
		if _, err := run(t, "merge", "closing", "--run", runDir, "--seat-id", seatID,
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
		if _, err := run(t, "merge", "position", "--run", runDir, "--seat-id", seatID, "--reason", text); err != nil {
			t.Fatal(err)
		}
	}
	positions := 0
	for _, e := range events(t, runDir) {
		if e.Type == "position" {
			positions++
			if got := e.Payload.Str("text"); got != "first" {
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
			if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
				"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
		}
	}

	failDir := t.TempDir()
	mint2(failDir)
	if _, err := run(t, "merge", "verdict", "--run", failDir, "--seat-id", seatID, "--as", "FAIL"); err != nil {
		t.Fatalf("FAIL must be allowed with open gaps: %v", err)
	}

	runDir := t.TempDir()
	mint2(runDir)
	_, err := run(t, "merge", "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS")
	if err == nil {
		t.Fatal("PASS was recorded over 2 open gaps — the rubber-stamp the guard exists to stop")
	}
	for _, want := range []string{"R1-1", "R1-2", "PASS refused"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the open gaps, got: %v", err)
		}
	}
	for _, id := range []string{"R1-1", "R1-2"} {
		if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--as", "closed",
			"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./x", "--reason", "resolved"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := run(t, "merge", "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS"); err != nil {
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
			if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
				"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
				t.Fatal(err)
			}
			args := []string{"merge", "verdict", "--run", runDir, "--seat-id", seatID}
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
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check", "c", "--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	// A PASS is refused over an open gap, so close it first (the guard is exercised in its
	// own test); this test is about render + checkpoint on a legitimate PASS.
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID,
		"--id", "R1-1", "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./x", "--reason", "resolved"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "merge", "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS")
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
	_, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1", "stray prose")
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
	if _, err := run(t, "lens", "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"); err != nil {
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
	_, err := run(t, "lens", "register", "--seat-id", "red-lens-r1-L1")
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
	if _, err := run(t, "lens", "register", "--run", explicit, "--seat-id", "red-lens-r1-L1"); err != nil {
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
	if _, err := run(t, "merge", "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
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
	if _, err := run(t, "merge", "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1", "--none"); err == nil ||
		!strings.Contains(err.Error(), "--reason") {
		t.Fatalf("--none with no reason must be refused, got %v", err)
	}
}

func TestSpotCheckRefusesContradictoryFlags(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	if _, err := run(t, "merge", "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--none", "--reason", "x", "--ids", "R1-4"); err == nil || !strings.Contains(err.Error(), "contradictory") {
		t.Fatalf("claiming both nothing-to-sample and a sample must be refused, got %v", err)
	}
}

// The bare form still works and still records an empty array. The friction claimed the
// tool could not record an empty round; it could, and TestSpotCheckIdsAreAlwaysAnArray
// caught the attempt to break that. What was missing was the REASON, not the record.
func TestBareSpotCheckStillRecordsAnEmptyArray(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()
	if _, err := run(t, "merge", "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatalf("the bare form must keep working: %v", err)
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
	if _, err := run(t, "merge", "register", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	minted, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class-new", "some-class", "--definition", "d", "--neighbor", "n", "--distinguisher", "x",
		"--location", "sec 1", "--problem", "p", "--fix", "f", "--check", "c", "--likelihood", "medium", "--impact", "medium",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--cx", "low")
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
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./internal/x",
		"--reason-file", prose); err != nil {
		t.Fatalf("--file must work on close, as it does on every other prose verb: %v", err)
	}
	if !payloadKeys(lastOfType(t, runDir, "close"))["prose"] {
		t.Error("the prose from --file must reach the event")
	}
}
