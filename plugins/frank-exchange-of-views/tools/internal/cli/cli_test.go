package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	r, w, perr := os.Pipe()
	if perr != nil {
		t.Fatal(perr)
	}
	saved := os.Stdout
	os.Stdout = w

	root := newRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(args)
	err = root.Execute()

	os.Stdout = saved
	w.Close()
	var buf bytes.Buffer
	if _, cerr := buf.ReadFrom(r); cerr != nil {
		t.Fatal(cerr)
	}
	r.Close()
	return buf.String(), err
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

func lastOfType(t *testing.T, runDir, typ string) record.Event {
	t.Helper()
	for i, evs := len(events(t, runDir))-1, events(t, runDir); i >= 0; i-- {
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
		"--class", "x", "--check", "c", "--problem", "p")
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
	_, err = run(t, "blue", "revision", "--run", runDir, "--seat-id", "hand-invented", "--text", "t")
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
			if !strings.Contains(err.Error(), "available:") || !strings.Contains(err.Error(), "render") {
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
	// Every role can render, and every role can register.
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		if !verbs[role]["render"] {
			t.Errorf("%s cannot render", role)
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
	// register is EXEMPT from the after-render: it creates the seat rather than
	// changing the board.
	if _, serr := os.Stat(filepath.Join(runDir, "records", "render-shadow")); serr == nil {
		t.Error("register triggered a render; it mutates nothing on the board")
	}

	out, err = run(t, "lens", "finding", "--run", runDir, "--seat-id", seatID,
		"--label", "L1-F1", "--severity", "high", "--likelihood", "medium", "--impact", "high",
		"--location", "§2", "--text", "the finding prose")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "finding L1-F1 recorded") {
		t.Errorf("finding said %q", out)
	}
	// Render-on-mutation keeps projections current after every write.
	if _, serr := os.Stat(filepath.Join(runDir, "records", "render-shadow", "ledger.md")); serr != nil {
		t.Errorf("a mutating verb did not render: %v", serr)
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
		"--label", "L1-F1", "--severity", "high", "--text", "t"); err != nil {
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
		"--class", "scope-creep", "--check", "c", "--problem", "p"); err != nil {
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
		"--class", "x", "--check", "c", "--problem", "p", "--severity", "catastrophic")
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
			"--class", "scope-creep", "--check", "c", "--problem", "p"}, extra...)
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
		"--class", "scope-creep", "--check", "c", "--problem", "p")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "minted R2-1") {
		t.Errorf("round 2's first mint said %q, want R2-1", out)
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
		"--check", "c", "--problem", "p")
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

// --problem and the prose channel are alternatives; --problem wins when both
// are given, and --file carries anything above trivial size.
func TestProseChannelResolution(t *testing.T) {
	t.Run("--file is read whole", func(t *testing.T) {
		runDir := t.TempDir()
		body := "line one\nline two — with unicode ✓ and <angle> brackets\n"
		f := filepath.Join(t.TempDir(), "prose.md")
		if err := os.WriteFile(f, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1", "--file", f); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "position").Payload.Str("text"); got != body {
			t.Errorf("text = %q, want the file verbatim %q", got, body)
		}
	})

	t.Run("--file wins over --text when both are passed", func(t *testing.T) {
		runDir := t.TempDir()
		f := filepath.Join(t.TempDir(), "prose.md")
		if err := os.WriteFile(f, []byte("from the file"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1",
			"--file", f, "--text", "from the flag"); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "position").Payload.Str("text"); got != "from the file" {
			t.Errorf("text = %q, want the file to win", got)
		}
	})

	t.Run("a missing --file is an error, not an empty payload", func(t *testing.T) {
		runDir := t.TempDir()
		_, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1",
			"--file", filepath.Join(t.TempDir(), "no-such-file.md"))
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
			"--class", "x", "--check", "c", "--problem", "from the flag", "--text", "from the prose channel"); err != nil {
			t.Fatal(err)
		}
		if got := lastOfType(t, runDir, "mint").Payload.Str("problem"); got != "from the flag" {
			t.Errorf("problem = %q, want --problem to win", got)
		}
	})

	t.Run("the prose channel fills problem when --problem is absent", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--class", "x", "--check", "c", "--text", "from the prose channel"); err != nil {
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
		"--class", "x", "--check", "c", "--problem", "p"); err != nil {
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
		"--anchor-seat", "L1", "--anchor-tool", "git show", "--anchor-target", "7bc501e:f")
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
		"--class", "x", "--check", "c", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	base := []string{"merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x",
		"--as", "closed_with_regression"}
	if _, err := run(t, base...); err == nil {
		t.Fatal("closed_with_regression was accepted without a successor — lineage dropped")
	}
	if _, err := run(t, append(base, "--successor", "R2-1")...); err != nil {
		t.Fatalf("a successor'd regression close was refused: %v", err)
	}
}

// close --prose-file carries the full closure record; a missing one is an error.
func TestCloseProseFile(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check", "c", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "closure.md")
	if err := os.WriteFile(f, []byte("the whole closure record"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x", "--prose-file", f); err != nil {
		t.Fatal(err)
	}
	if got := lastOfType(t, runDir, "close").Payload.Str("prose"); got != "the whole closure record" {
		t.Errorf("prose = %q", got)
	}

	_, err := run(t, "merge", "close", "--run", runDir, "--seat-id", seatID, "--id", "R1-1",
		"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x",
		"--prose-file", filepath.Join(t.TempDir(), "gone.md"))
	if err == nil {
		t.Fatal("a missing --prose-file was ignored")
	}
}

func TestVerbsThatRefuseWithoutTheirReason(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"regrade without --basis", []string{"merge", "regrade", "--id", "R1-1", "--severity", "high"}, "regrade requires --basis"},
		{"dispose without --as", []string{"merge", "dispose", "--observation", "F1"}, "dispose requires --as"},
		{"mint without --check", []string{"merge", "mint", "--class", "x", "--problem", "p"}, "mint requires --check"},
		{"mint without --class", []string{"merge", "mint", "--check", "c", "--problem", "p"}, "mint requires --class"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			seatID := "red-merge-r1"
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

	t.Run("revision records the claim count alongside the prose", func(t *testing.T) {
		if _, err := run(t, "blue", "revision", "--run", runDir, "--seat-id", seatID,
			"--text", "what changed", "--claim-count", "12"); err != nil {
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
	full := map[string]string{
		"gap-id": "R1-1", "disposition": "carried", "principle": "correctness first",
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
	args := []string{"bench", "opinion", "--run", runDir, "--seat-id", seatID, "--file", writeTemp(t, "the rationale")}
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

// render is available to every seat, mutates nothing, and needs only --run.
func TestRenderVerbIsAvailableToEverySeatAndNeedsOnlyRun(t *testing.T) {
	runDir := t.TempDir()
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "x", "--check", "c", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		t.Run(role, func(t *testing.T) {
			// No --seat-id: render is read-only and belongs to no seat.
			out, err := run(t, role, "render", "--run", runDir)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if !strings.Contains(out, "feov-record "+role+": rendered to ") {
				t.Errorf("render said %q", out)
			}
			if !strings.Contains(out, "(1 open, 0 closed)") {
				t.Errorf("render did not report the board: %q", out)
			}
		})
	}
	_, err := run(t, "lens", "render")
	if err == nil {
		t.Fatal("render ran without --run")
	}
	if !strings.Contains(err.Error(), "--run <runDir> is required") {
		t.Errorf("error = %q", err)
	}
}

// The anomaly count reaches the seat's own output: a render that hides a double
// dispatch is how a divergence goes unexplained.
func TestRenderReportsAnomalies(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-lens-r1-L1"
	if _, err := run(t, "lens", "register", "--run", runDir, "--seat-id", seatID); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", seatID, "--label", "F1", "--text", "a"); err != nil {
		t.Fatal(err)
	}
	// A second dispatch of the same seat: a crash re-run.
	if _, err := run(t, "lens", "register", "--run", runDir, "--seat-id", seatID); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", seatID, "--label", "F2", "--text", "b"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "lens", "render", "--run", runDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "anomalies") {
		t.Errorf("render did not report the double dispatch: %q", out)
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
				"--text", "the capability I needed")
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

	// petition takes a role-specific SUFFIX while recording the same event.
	for _, tc := range cases[:3] {
		t.Run("petition/"+tc.role, func(t *testing.T) {
			runDir := t.TempDir()
			out, err := run(t, tc.role, "petition", "--run", runDir, "--seat-id", tc.seatID,
				"--class", "safety", "--basis", "what happened", "--relief", "the relief sought")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, "petition filed (safety)") {
				t.Errorf("petition said %q", out)
			}
			if tc.role == "lens" && !strings.Contains(out, "the bench hears it before the debate continues") {
				t.Errorf("the lens suffix is missing: %q", out)
			}
			if tc.role != "lens" && strings.Contains(out, "the bench hears it") {
				t.Errorf("%s got the lens's suffix: %q", tc.role, out)
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
	for _, id := range []string{"R1-1", "R1-2"} {
		if _, err := run(t, "merge", "closing", "--run", runDir, "--seat-id", seatID,
			"--id", id, "--text", "argued "+id); err != nil {
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
		if _, err := run(t, "merge", "position", "--run", runDir, "--seat-id", seatID, "--text", text); err != nil {
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

// verdict is the merge seat's terminal act: it renders and checkpoints.
func TestVerdictRendersAndCheckpoints(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check", "c", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "merge", "verdict", "--run", runDir, "--seat-id", seatID, "--as", "PASS")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "verdict PASS") || !strings.Contains(out, "(1 open, 0 closed)") {
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
