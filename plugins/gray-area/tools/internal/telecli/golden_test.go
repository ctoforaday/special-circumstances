package telecli

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// -update rewrites the golden files. It is a flag rather than an environment variable so
// `go test ./internal/telecli -update` is the whole procedure, and so `go test ./...` cannot
// accidentally be running in update mode — a golden suite that silently records whatever the code
// currently prints asserts nothing at all.
var update = flag.Bool("update", false, "rewrite the golden files from this run's output")

// harness is one fixture world: a corpus, a store built from it, and a frozen clock.
type harness struct {
	env      Env
	projects string
	sessions string
	store    string
}

// newColdHarness builds the fixture world WITHOUT projecting it, for the tests that are about the
// projection itself.
func newColdHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		projects: writeCorpus(t),
		sessions: filepath.Join(t.TempDir(), "sessions"),
		store:    filepath.Join(t.TempDir(), "catalogue.db"),
	}
	sessionFile(t, h.sessions, 4242, alphaID, alphaCWD)
	sessionFile(t, h.sessions, 4243, betaID, betaCWD)
	h.env = Env{Store: h.store, ProjectsDir: h.projects, SessionsDir: h.sessions,
		Now: func() time.Time { return frozen }}
	return h
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := newColdHarness(t)
	// Every read verb is asserted against a store built the way a user builds one: by running the
	// backfill verb through the same argv path they would.
	if _, _, code := h.run(t, "backfill"); code != 0 {
		t.Fatalf("fixture backfill failed with exit %d", code)
	}
	return h
}

// run drives the command tree exactly as main does — argv in, bytes and an exit code out.
func (h *harness) run(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	root := NewRoot(h.env)
	root.SetOut(&o)
	root.SetErr(&e)
	root.SetArgs(args)
	// The same two lines Execute runs, so a test's exit code IS the binary's exit code.
	err := root.Execute()
	if err != nil {
		fmt.Fprintf(&e, "telepathy: %v\n", err)
	}
	return h.scrub(o.String()), h.scrub(e.String()), exitCode(err)
}

// scrub replaces this run's temporary paths with stable placeholders, and normalises the
// separator. Without the first every golden would hold a t.TempDir() path and fail on the next
// run — and a golden nobody can keep passing gets deleted, which is how a suite loses the
// assertions it was written for.
//
// The separator matters for the same reason and is easier to miss: a path this tool PRINTS comes
// from filepath, so `<CORPUS>/-work-alpha` on Linux is `<CORPUS>\-work-alpha` on Windows, and one
// golden cannot hold both. Normalising to forward slashes is safe here because no fixture content
// contains a backslash of its own — if that stops being true, this line starts corrupting the
// comparison rather than stabilising it.
func (h *harness) scrub(s string) string {
	for _, sub := range []struct{ path, name string }{
		{h.projects, "<CORPUS>"}, {h.sessions, "<SESSIONS>"}, {h.store, "<STORE>"},
	} {
		// The %q-ESCAPED form first. Anything cobra renders through %q — a flag default, a
		// quoted argument in an error — arrives with its separators doubled on Windows, so it
		// does not match the plain path and survives to the line below, which turns `\` into
		// `//` and produces a diff nobody can read. Replacing the escaped form first is what
		// stops that; it is a no-op everywhere the separator is already a slash.
		s = strings.ReplaceAll(s, strings.ReplaceAll(sub.path, "\\", "\\\\"), sub.name)
		s = strings.ReplaceAll(s, sub.path, sub.name)
	}
	return strings.ReplaceAll(s, "\\", "/")
}

// assertGolden compares against testdata/golden/<name>.txt.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	// `.golden`, not `.txt`, and deliberately: .gitattributes already pins `*.golden` to LF for
	// exactly this reason, and .txt is outside that list. A Windows checkout converted every one
	// of these to CRLF while the code emits LF, so all 21 differed from themselves on one
	// platform. Using the covered extension keeps a hand-kept roster from having to grow — the
	// roster going stale is the same defect one level up.
	path := filepath.Join("testdata", "golden", name+".golden")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no golden for %s: %v\n--- this run produced ---\n%s\nrun `go test ./internal/telecli -update` if that is correct", name, err, got)
	}
	if string(want) != got {
		// NAME THE LINE-ENDING CASE OUT LOUD. Its diff renders identically on both sides, so a
		// reader compares two blocks of text that look the same and concludes the harness is
		// broken. One line of diagnosis is worth more than the diff here.
		if strings.ReplaceAll(string(want), "\r\n", "\n") == got {
			t.Errorf("%s differs only in LINE ENDINGS — the checkout converted it to CRLF. "+
				"Check that .gitattributes still pins *.golden to eol=lf.", name)
			return
		}
		t.Errorf("%s does not match its golden.\n--- want ---\n%s\n--- got ---\n%s", name, want, got)
	}
}

// THE OUTPUT IS A CONTRACT, so it is pinned byte for byte.
//
// These verbs are read by humans deciding whether to touch a file another agent is in, and by
// scripts. A change to any of them is a change a reviewer should have to look at — which is what a
// golden buys and what "check a few substrings" does not: a substring check passes when a column
// silently starts printing the wrong number beside the right word.
func TestGoldenOutput(t *testing.T) {
	h := newHarness(t)
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"agents", []string{"agents"}},
		{"session-with-reasoning", []string{"session", alphaID}},
		{"session-without-reasoning", []string{"session", betaID}},
		{"touched-hit", []string{"touched", "mint.go"}},
		// view.go is written once by Write and once by a `sed` inside a Bash command. Both are
		// acts on the file, and a reader deciding whether to open it needs to see both.
		{"touched-through-the-shell", []string{"touched", "/work/alpha/view.go"}},
		{"touched-miss", []string{"touched", "nothing/here.go"}},
		{"sql-tools", []string{"sql", "SELECT tool, outcome, count(*) n FROM v_action GROUP BY 1,2 ORDER BY 1,2"}},
		{"sql-thoughts", []string{"sql", "SELECT agent_id, text FROM v_thought ORDER BY block_seq"}},
		{"sql-null-renders-as-the-word", []string{"sql", "SELECT session_id, closed_at FROM v_session ORDER BY 1"}},
		// seq is per-FILE, so `ORDER BY seq` alone leaves ties for SQLite to break however it
		// likes and the golden would be recording an accident of the scan order.
		{"sql-limit-reports-truncation", []string{"sql", "--limit", "2",
			"SELECT session_id, agent_id, seq, tool FROM v_action ORDER BY session_id, agent_id, seq"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, errOut, code := h.run(t, tc.args...)
			if code != 0 {
				t.Fatalf("exit %d, stderr:\n%s", code, errOut)
			}
			assertGolden(t, tc.name, out)
		})
	}
}

// THE HELP IS PART OF THE CONTRACT TOO, and more load-bearing than most output: it is where the
// tool states what it CANNOT tell you. A refactor that drops the 'unknown is not ended' paragraph
// leaves every verb working and every caller worse informed, and no functional test would notice.
func TestGoldenHelp(t *testing.T) {
	h := newHarness(t)
	for _, name := range []string{"", "agents", "session", "touched", "find", "sql", "backfill"} {
		label := "help-root"
		args := []string{"--help"}
		if name != "" {
			label, args = "help-"+name, []string{name, "--help"}
		}
		t.Run(label, func(t *testing.T) {
			out, _, code := h.run(t, args...)
			if code != 0 {
				t.Fatalf("--help exited %d", code)
			}
			assertGolden(t, label, out)
		})
	}
}

// find shells out, so its golden runs only where ripgrep is installed. The skip is LOUD about
// which of the two it is: a silently skipped test and a passing one look identical in a summary.
func TestGoldenFind(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed on this machine, so `find` cannot be exercised here — " +
			"this is a skipped assertion, not a passing one (see requirements.json: ripgrep is required)")
	}
	h := newHarness(t)
	// A term in THREE of the four transcripts, chosen so the output has an order to get wrong:
	// ripgrep walks the argv, the argv is built from a map, and a single-hit golden would pass
	// whether or not the results are sorted at all.
	out, errOut, code := h.run(t, "find", "mint.go")
	if code != 0 {
		t.Fatalf("exit %d, stderr:\n%s", code, errOut)
	}
	assertGolden(t, "find-hit", out)

	out, _, code = h.run(t, "find", "a phrase that appears in no transcript")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	assertGolden(t, "find-miss", out)
}

// BACKFILL HAS TWO OUTPUTS AND BOTH MATTER: what a cold read reports, and what a repeat reports.
//
// The cold one carries the unparsed count, and the fixture's torn line is why it is asserted: an
// unparsed record leaves exactly the same hole as a record that never existed, and that one
// sentence is the only thing that tells them apart. The warm one is the idempotency claim the verb
// makes in its own help — "re-running is safe" — checked rather than believed.
func TestGoldenBackfill(t *testing.T) {
	h := newColdHarness(t)
	first, errOut, code := h.run(t, "backfill")
	if code != 0 {
		t.Fatalf("exit %d, stderr:\n%s", code, errOut)
	}
	assertGolden(t, "backfill-cold", first)
	if !strings.Contains(first, "did not parse") {
		t.Error("the corpus contains a torn line and the cold read did not say so")
	}
	second, _, code := h.run(t, "backfill")
	if code != 0 {
		t.Fatalf("second pass exit %d", code)
	}
	assertGolden(t, "backfill-warm", second)
	// Not merely "the golden matches": the numbers are what the claim is about.
	if !strings.Contains(second, "0 acts, 0 words, 0 thoughts") {
		t.Errorf("re-running re-ingested rows, so the stored offsets are not being honoured:\n%s", second)
	}
}
