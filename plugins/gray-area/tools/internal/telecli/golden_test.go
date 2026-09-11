package telecli

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
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

// storeVariant names a store in a state the capture hooks do not produce on their own.
type storeVariant int

const (
	// olderShape: the pre-#867 DDL (testdata/shape1.sql) applied to an empty file, stamped 1.
	olderShape storeVariant = iota
	// foreign: an SQLite file holding one table no catalogue has ever had.
	foreign
	// rebuiltPending: a store built as newHarness builds one, then made to read as rebuilt at
	// rebuiltAt and never backfilled since.
	rebuiltPending
)

// rebuiltAt is the rebuild instant the rebuiltPending store carries. A FIXED instant, in the past,
// so the warning's timestamp is stable in a golden and any real backfill's wall clock is later.
var rebuiltAt = time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC)

// fixtureDB opens path read-write for a fixture, through a `file:` URI for the same reason the
// catalogue's own connections use one: a bare path is cut at its first `?` by the driver.
func fixtureDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.ToSlash(abs)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: p}).String())
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func mustExec(t *testing.T, db *sql.DB, stmts ...string) {
	t.Helper()
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("%.60s: %v", s, err)
		}
	}
}

// newStoreHarness is the fixture world with a store in one of the states above.
func newStoreHarness(t *testing.T, v storeVariant) *harness {
	t.Helper()
	switch v {
	case olderShape:
		h := newColdHarness(t)
		ddl, err := os.ReadFile(filepath.Join("testdata", "shape1.sql"))
		if err != nil {
			t.Fatal(err)
		}
		db := fixtureDB(t, h.store)
		defer db.Close()
		mustExec(t, db, string(ddl), `PRAGMA user_version = 1`)
		return h
	case foreign:
		h := newColdHarness(t)
		db := fixtureDB(t, h.store)
		defer db.Close()
		mustExec(t, db, `CREATE TABLE ledger(id INTEGER PRIMARY KEY, amount INTEGER)`)
		return h
	case rebuiltPending:
		h := newHarness(t) // its backfill wrote backfilled_at...
		db, err := catalogue.Open(h.store, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		// ...which is deleted, and the rebuild markers written, so the store reads exactly as one
		// an upgrade rebuilt and nobody has backfilled since.
		mustExec(t, db, `DELETE FROM meta WHERE key = '`+catalogue.MetaBackfilledAt+`'`)
		if err := catalogue.MarkRebuilt(db, 2, rebuiltAt); err != nil {
			t.Fatal(err)
		}
		return h
	}
	t.Fatalf("unknown store variant %d", v)
	return nil
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

// assertGoldenFull pins the WHOLE outcome in one file — exit code, stdout and stderr — for the
// cases where stderr is the point: a refusal, or a warning printed beside a normal answer. A
// stdout-only golden of a refusal is an empty file, and would pass against any refusal at all.
func assertGoldenFull(t *testing.T, name string, code int, stdout, stderr string) {
	t.Helper()
	assertGolden(t, name, fmt.Sprintf("exit: %d\n--- stdout\n%s--- stderr\n%s", code, stdout, stderr))
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
		// first_act/last_act come from the ACTS and are what a reader wanted from the old
		// first_seen/last_seen; ingested_* are the store's own bookkeeping and are now named as
		// such. A session with no acts reports NULL for the pair, which is a real state.
		{"sql-session-span-vs-ingest", []string{"sql",
			"SELECT substr(session_id,1,8) sid, first_act, last_act, (ingested_first = ingested_last) same_pass FROM v_session ORDER BY 1"}},
		// The population that used to be invisible: reasoning the client withheld.
		{"sql-withheld-reasoning", []string{"sql",
			"SELECT reason, count(*) n FROM v_skip GROUP BY 1"}},
		// seq is per-FILE, so `ORDER BY seq` alone leaves ties for SQLite to break however it
		// likes and the golden would be recording an accident of the scan order.
		{"sql-limit-reports-truncation", []string{"sql", "--limit", "2",
			"SELECT session_id, agent_id, seq, tool FROM v_action ORDER BY session_id, agent_id, seq"}},
		// WHO SPOKE, as the word tier stores it: one row per speaker the fixture holds, and no
		// `user` row for a peer, a notification, a seat prompt or harness text.
		{"sql-word-roles", []string{"sql", "SELECT role, count(*) FROM v_word GROUP BY 1 ORDER BY 1"}},
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

	// --in narrows to one channel, and the fixture's mint.go hits are all tool arguments, so
	// filtering to what an agent SAID must come back empty — worded, and counted in transcripts.
	out, _, code = h.run(t, "find", "mint.go", "--in", "assistant")
	if code != 0 {
		t.Fatalf("--in exited %d", code)
	}
	assertGolden(t, "find-in-channel", out)

	// EVERY CHANNEL THE DECODER CAN PRODUCE NEEDS A ROW SOMEWHERE, or the branch that produces it
	// is untested. The mint.go hits above are all tool arguments; these three reach the others,
	// and without them a mutation that classified all assistant text as unknown survived the
	// whole suite.
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"find-channel-assistant", []string{"carriers"}},  // "Widening the four carriers." — the agent speaking
		{"find-channel-user", []string{"widen the gap"}},  // the human's prompt
		{"find-channel-thinking", []string{"four sites"}}, // the one thought that carried text
		// A peer's turn (isMeta, top level) and a peer's mid-turn delivery, in two transcripts.
		{"find-channel-peer", []string{"conduct section", "--in", "peer"}},
		// THE SAME TWO ROWS FROM A PATTERN. IN is read from where ripgrep matched, so a regex is
		// attributed exactly as a literal is — a pattern is not a substring, and re-finding it in
		// the decoded text found no speaker at all.
		{"find-regex-in-peer", []string{"--regex", "conduct sect(ion)", "--in", "peer"}},
		// A notification turn, and one delivered mid-turn in a seat by commandMode alone — the
		// 12-character channel under the widened IN header.
		{"find-channel-notification", []string{"nightly build"}},
		// A seat's own prompt, and a coordinator's mid-turn message in another seat.
		{"find-channel-lead", []string{"every caller"}},
		// isMeta text at both tiers, and a compaction summary (the newer of main's two hits).
		{"find-channel-harness", []string{"task tools"}},
		// The human said it first and the agent repeated it: the row stays, showing the human.
		{"find-in-any-hit", []string{"ledger column", "--in", "user"}},
	} {
		out, errOut, code := h.run(t, append([]string{"find"}, tc.args...)...)
		if code != 0 {
			t.Fatalf("%s exited %d, stderr:\n%s", tc.name, code, errOut)
		}
		assertGolden(t, tc.name, out)
	}

	// THE #885 REPRO: a peer's message is not the human's. `--in user` must not return either
	// transcript that holds it, and must say so in the worded empty result.
	out, _, code = h.run(t, "find", "conduct section", "--in", "user")
	if code != 0 {
		t.Fatalf("--in user exited %d", code)
	}
	if strings.Contains(out, catalogue.Short(gammaID)) || strings.Contains(out, catalogue.Short(deltaID)) {
		t.Errorf("a peer's message was reported as the human's:\n%s", out)
	}
	if !strings.Contains(out, `2 transcript(s) contain "conduct section"`) {
		t.Errorf("the empty result did not count the transcripts that matched:\n%s", out)
	}

	// The any-hit case is only a case if the NEWEST hit is not the human's: unfiltered, the row
	// must show the agent, or find-in-any-hit would pass under the old most-recent filter too.
	out, _, _ = h.run(t, "find", "ledger column")
	if !strings.Contains(out, "Renaming the ledger column now.") {
		t.Errorf("the fixture's newest 'ledger column' hit is not the agent's, so --in user proves nothing:\n%s", out)
	}

	// A channel that cannot exist is a typo, and must be refused rather than filtered to nothing.
	_, errOut, code = h.run(t, "find", "mint.go", "--in", "asistant")
	if code != 2 {
		t.Errorf("an unknown channel exited %d, want 2 (a typo is an argv error)", code)
	}
	if !strings.Contains(errOut, "is not a channel") {
		t.Errorf("the refusal does not name the problem: %q", errOut)
	}

	out, _, code = h.run(t, "find", "a phrase that appears in no transcript")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	assertGolden(t, "find-miss", out)
}

// --in IS NEVER CAPPED. Unfiltered, a row reads at most the newest samplesPerFile records; under a
// filter that cap is the drop this fixes — a human's hit older than 200 agent repetitions vanished
// and the transcript read as though the human had never said it. Driven through decodeHits and
// render directly, because the golden corpus is too small for the cap to bite.
func TestInReadsEveryHitAndStatesNoCap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "S.jsonl")
	body := []string{turnLine("h1", "S", "/w", 10*time.Hour, origin("human"), "the sleeper term, first")}
	for i := 0; i < samplesPerFile+50; i++ {
		body = append(body, assistantLine(fmt.Sprintf("a%d", i), "h1", "S", "/w", time.Duration(samplesPerFile+50-i)*time.Minute,
			map[string]any{"type": "text", "text": "repeating the sleeper term"}))
	}
	if err := os.WriteFile(path, []byte(strings.Join(body, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// What ripgrep would report: each record's match at its ABSOLUTE offset in the file written.
	var recs []ripRecord
	var start int64
	for n, line := range body {
		abs := start + int64(strings.Index(line, "sleeper term"))
		recs = append(recs, ripRecord{path: path, line: n + 1, matches: []rawMatch{{abs: abs, text: "sleeper term"}}})
		start += int64(len(line) + 1)
	}
	paths := map[string]catalogue.TranscriptFile{path: {Path: path, SessionID: "S"}}

	rows, matched, stale := decodeHits(recs, paths, "")
	if len(rows) != 1 || matched != 1 || stale != 0 || !rows[0].capped || rows[0].best.Channel != catalogue.ChannelAssistant {
		t.Fatalf("unfiltered: %d rows, %d matched, %d stale, %+v — want one capped assistant row", len(rows), matched, stale, rows)
	}

	rows, matched, stale = decodeHits(recs, paths, catalogue.ChannelUser)
	if len(rows) != 1 || matched != 1 || stale != 0 {
		t.Fatalf("--in user: %d rows of %d matched, %d stale, want the one transcript kept", len(rows), matched, stale)
	}
	r := rows[0]
	if r.capped || r.hits != len(body) || r.best.Channel != catalogue.ChannelUser || !strings.Contains(r.best.Snippet, "first") {
		t.Errorf("--in user: capped=%v hits=%d channel=%q snippet=%q — want uncapped, %d hits, the human's words",
			r.capped, r.hits, r.best.Channel, r.best.Snippet, len(body))
	}
	var out bytes.Buffer
	render(&out, rows, false)
	if strings.Contains(out.String(), "were read for WHEN and SNIPPET") {
		t.Errorf("a filtered search printed the cap footer:\n%s", out.String())
	}
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

// A STORE IN THE WRONG STATE IS SAID IN WORDS, on stderr, with the exit code a script reads.
//
// An older store used to reach `sql` and fail inside SQLite ("no such column: s.ingested_first");
// a foreign file used to be queried as though it were a catalogue. Both are now refused before a
// query runs. A rebuilt store answers — its rows are true — and says beside the answer that rows
// from before the rebuild are missing.
func TestGoldenStoreStates(t *testing.T) {
	for _, tc := range []struct {
		name    string
		variant storeVariant
		args    []string
		code    int
	}{
		{"sql-older-shape", olderShape, []string{"sql", "SELECT count(*) FROM v_session"}, 1},
		{"sql-not-a-catalogue", foreign, []string{"sql", "SELECT count(*) FROM v_session"}, 1},
		{"agents-rebuilt-warning", rebuiltPending, []string{"agents"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newStoreHarness(t, tc.variant)
			out, errOut, code := h.run(t, tc.args...)
			if code != tc.code {
				t.Errorf("exit %d, want %d; stderr:\n%s", code, tc.code, errOut)
			}
			assertGoldenFull(t, tc.name, code, out, errOut)
		})
	}
}

// rebuildWarning is the one fragment of the warning this test counts by. The golden above pins
// the whole line; this only needs to find it once per verb, and zero times after a backfill.
const rebuildWarning = "has not been backfilled since"

// EVERY READ VERB CARRIES THE WARNING, because each reaches the store through openRead and a verb
// that skipped it would answer from a gutted store with no word said. After a backfill, none does.
func TestTheRebuildWarningReachesEveryReadVerb(t *testing.T) {
	h := newStoreHarness(t, rebuiltPending)
	_, noRipgrep := exec.LookPath("rg")
	verbs := [][]string{
		{"agents"},
		{"touched", "mint.go"},
		{"sql", "SELECT 1"},
		{"find", "mint.go"},
		{"session", alphaID},
	}
	check := func(wantWarnings int) {
		t.Helper()
		for _, args := range verbs {
			_, errOut, code := h.run(t, args...)
			if got := strings.Count(errOut, rebuildWarning); got != wantWarnings {
				t.Errorf("%v printed the rebuild warning %d times, want %d; stderr:\n%s", args, got, wantWarnings, errOut)
			}
			switch {
			case args[0] == "find" && noRipgrep != nil:
				// openRead runs before the ripgrep lookup, so the warning is still owed; the verb
				// then refuses in its own words rather than reporting zero matches.
				if code == 0 || !strings.Contains(errOut, errNoRipgrep.Error()) {
					t.Errorf("find without ripgrep: exit %d, stderr:\n%s", code, errOut)
				}
			case code != 0:
				t.Errorf("%v exited %d; stderr:\n%s", args, code, errOut)
			}
		}
	}
	check(1)
	if _, errOut, code := h.run(t, "backfill"); code != 0 {
		t.Fatalf("backfill exited %d:\n%s", code, errOut)
	}
	check(0)
}
