package telecli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// THE ARGV IS A CONTRACT WITH SCRIPTS, and an exit code is the only part of it a script reads
// without parsing prose. 2 must mean "you typed it wrong" and 1 must mean "the world said no",
// because a caller retrying on 1 and reporting on 2 gets both backwards otherwise.
func TestExitCodes(t *testing.T) {
	h := newHarness(t)
	for _, tc := range []struct {
		name string
		args []string
		want int
	}{
		{"a verb that ran", []string{"agents"}, 0},
		{"a verb that found nothing still ran", []string{"touched", "no/such/file"}, 0},
		{"an unknown verb", []string{"telepathise"}, 2},
		{"a verb missing its argument", []string{"touched"}, 2},
		{"a verb given too many arguments", []string{"session", "a", "b"}, 2},
		{"a verb given an argument it takes none of", []string{"agents", "extra"}, 2},
		{"an unknown flag", []string{"agents", "--verbose"}, 2},
		{"a flag given a value of the wrong type", []string{"sql", "--limit", "many", "SELECT 1"}, 2},
		{"a negative limit", []string{"sql", "--limit", "-1", "SELECT 1"}, 2},
		{"invalid SQL is the world saying no, not a typo in the argv", []string{"sql", "SELECT FROM"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, code := h.run(t, tc.args...)
			if code != tc.want {
				t.Errorf("exit %d, want %d", code, tc.want)
			}
		})
	}
}

// A MISSING STORE IS A SENTENCE, NOT AN EMPTY TABLE.
//
// This is the plausible zero the whole plugin is built against: `touched` printing nothing because
// the catalogue does not exist reads exactly like `touched` printing nothing because nobody has
// the file open. One of those is a reason to proceed and the other is not.
func TestAMissingStoreIsWordedNotEmpty(t *testing.T) {
	env := Env{
		Store:       filepath.Join(t.TempDir(), "absent.db"),
		ProjectsDir: t.TempDir(),
		SessionsDir: t.TempDir(),
		Now:         func() time.Time { return frozen },
	}
	h := &harness{env: env, store: env.Store, projects: env.ProjectsDir, sessions: env.SessionsDir}
	for _, verb := range [][]string{{"touched", "x.go"}, {"session", "abc"}, {"agents"}, {"find", "x"}} {
		out, errOut, code := h.run(t, verb...)
		if code == 0 {
			t.Errorf("%v succeeded against a store that does not exist", verb)
		}
		if strings.TrimSpace(out) != "" {
			t.Errorf("%v printed a result from a store that does not exist: %q", verb, out)
		}
		if !strings.Contains(errOut, "empty") && !strings.Contains(errOut, "exist") {
			t.Errorf("%v did not say the store is missing: %q", verb, errOut)
		}
	}
}

// THE SQL SURFACE IS READ-ONLY, AND THAT IS ENFORCED RATHER THAN DOCUMENTED.
//
// `telepathy sql` takes a caller's arbitrary SQL. The store is a derived artifact, so a write
// would not lose anything irreplaceable — but it would silently make every later answer disagree
// with the transcripts, which is the one failure mode this tool cannot have. Three separate
// mechanisms say no; this asserts the behaviour, not any one of them.
func TestSQLCannotWrite(t *testing.T) {
	h := newHarness(t)
	for _, q := range []string{
		`DELETE FROM act`,
		`INSERT INTO act (session_id, seq, ts, tool) VALUES ('x', 0, 0, 'Bash')`,
		`UPDATE session SET project_dir = 'elsewhere'`,
		`DROP TABLE act`,
		`CREATE TABLE mine (a INT)`,
	} {
		out, errOut, code := h.run(t, "sql", q)
		if code == 0 {
			t.Errorf("a write was accepted: %s\n%s", q, out)
			continue
		}
		// WHY it was refused, not merely that it was. `DELETE FROM act` also fails when there is
		// no table called `act`, and a suite that accepted any error here would keep passing after
		// a rename that left the surface writable.
		if !strings.Contains(errOut, "read") && !strings.Contains(errOut, "authorization") {
			t.Errorf("%s was refused for an unrelated reason, so this asserts nothing: %q", q, errOut)
		}
	}
	// `PRAGMA writable_schema = 1` is the case that does NOT refuse, and asserting an error here
	// was wrong: defensive mode makes it a silent no-op, so the statement succeeds and changes
	// nothing. What matters is the value it reads back — 0 means the schema is still sealed. If
	// this ever reads 1, `_defensive` has stopped being applied and the pragma is live again,
	// which no test of the DSN STRING would notice.
	h.run(t, "sql", `PRAGMA writable_schema = 1`)
	back, _, code := h.run(t, "sql", `PRAGMA writable_schema`)
	if code != 0 {
		t.Fatalf("reading writable_schema back failed: exit %d", code)
	}
	if !strings.Contains(back, "0") {
		t.Errorf("writable_schema is settable, so defensive mode is not in force: %q", back)
	}
	// And the store still answers the same way afterwards.
	after, _, code := h.run(t, "sql", "SELECT count(*) FROM v_action")
	if code != 0 {
		t.Fatalf("the store stopped answering after the refused writes: exit %d", code)
	}
	if !strings.Contains(after, "8") {
		t.Errorf("the action count changed under refused writes: %q", after)
	}
}

// A QUERY MUST NOT REACH A SECOND DATABASE. ATTACH would let a caller's SQL read any file the
// process can — the store's read-only opening says nothing about what else is on the disk.
func TestSQLCannotAttach(t *testing.T) {
	h := newHarness(t)
	other := filepath.Join(t.TempDir(), "other.db")
	if err := os.WriteFile(other, []byte("not a database, and it must never be opened anyway"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, code := h.run(t, "sql", "ATTACH DATABASE '"+other+"' AS other")
	if code == 0 {
		t.Fatal("ATTACH was accepted")
	}
	if !strings.Contains(strings.ToLower(errOut), "attach") && !strings.Contains(strings.ToLower(errOut), "many") {
		t.Errorf("ATTACH failed for an unrelated reason, so this is not testing what it claims: %q", errOut)
	}
}

// A SEARCH THAT CANNOT RUN MUST REFUSE.
//
// With ripgrep absent, `find` must exit non-zero and must NOT print the same "no transcript
// contains" line it prints after a real search. The two are indistinguishable to a reader, and
// every "nobody else touched this" built on the wrong one is a lie.
func TestFindRefusesWithoutRipgrep(t *testing.T) {
	h := newHarness(t)
	t.Setenv("PATH", t.TempDir()) // a PATH with no rg on it
	out, errOut, code := h.run(t, "find", "mint.go")
	if code == 0 {
		t.Fatal("find reported success with no ripgrep installed")
	}
	if strings.Contains(out, "no transcript contains") {
		t.Error("find reported an empty RESULT for a search that never ran")
	}
	if !strings.Contains(errOut, "ripgrep") {
		t.Errorf("the refusal does not name ripgrep: %q", errOut)
	}
	if !strings.Contains(errOut, "sql") {
		t.Errorf("the refusal does not offer the way around it: %q", errOut)
	}
}

// find must also distinguish "the store names no transcripts" from "the search found nothing".
func TestFindRefusesAnEmptyStore(t *testing.T) {
	h := newColdHarness(t)
	// A store that EXISTS and names nothing — the state a fresh install is in before any session
	// has run, and the one most likely to be met by a first-time user.
	h.env.ProjectsDir = t.TempDir()
	if _, _, code := h.run(t, "backfill"); code != 0 {
		t.Fatal("backfill of an empty corpus should succeed")
	}
	out, errOut, code := h.run(t, "find", "anything")
	if code == 0 {
		t.Fatal("find succeeded against a store naming no transcripts")
	}
	if strings.Contains(out, "no transcript contains") {
		t.Error("an unsearchable store was reported as a search that found nothing")
	}
	if !strings.Contains(errOut, "backfill") {
		t.Errorf("the refusal does not say what would fix it: %q", errOut)
	}
}

// --version answers without a store, because the first thing anyone runs against a binary they do
// not trust is --version, and needing a working catalogue to get one is a preflight that cannot
// preflight.
func TestVersionNeedsNothing(t *testing.T) {
	env := Env{Store: filepath.Join(t.TempDir(), "absent.db"), Now: func() time.Time { return frozen }}
	h := &harness{env: env, store: env.Store, projects: t.TempDir(), sessions: t.TempDir()}
	out, _, code := h.run(t, "--version")
	if code != 0 {
		t.Fatalf("--version exited %d", code)
	}
	if !strings.Contains(out, "telepathy") {
		t.Errorf("--version does not name the binary: %q", out)
	}
	if strings.Count(out, "\n") != 1 {
		t.Errorf("--version should be one line, got %q", out)
	}
}

// EVERY VERB DOCUMENTS ITSELF. A cobra tree makes it cheap to add a verb and equally cheap to add
// one with a one-word Short and no Long, and the help IS this tool's statement of what it cannot
// tell you. This is the roster check: a new verb fails here until it has been written up.
func TestEveryVerbIsDocumented(t *testing.T) {
	root := NewRoot(Defaults())
	var seen []string
	for _, c := range root.Commands() {
		if c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		seen = append(seen, c.Name())
		if len(c.Short) < 10 {
			t.Errorf("%s has no useful Short", c.Name())
		}
		if len(c.Long) < 120 {
			t.Errorf("%s has no Long: the help is where this tool says what it cannot tell you", c.Name())
		}
		if c.Args == nil {
			t.Errorf("%s does not constrain its arguments, so a typo becomes a silent extra word", c.Name())
		}
		if c.RunE == nil {
			t.Errorf("%s has no RunE, so its failures cannot reach the exit-code mapping", c.Name())
		}
	}
	// The roster itself. A verb removed by accident is as much a regression as one added
	// undocumented, and nothing else in this package would notice.
	want := []string{"agents", "backfill", "find", "session", "sql", "touched"}
	if strings.Join(sorted(seen), " ") != strings.Join(want, " ") {
		t.Errorf("the verb set is %v, want %v", sorted(seen), want)
	}
}

func sorted(s []string) []string {
	out := append([]string(nil), s...)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// The three shared flags must be PERSISTENT. Declared per-verb they drift, and `--store` working
// on four verbs and not the fifth is the kind of thing nobody finds until they need it.
func TestTheSharedFlagsArePersistent(t *testing.T) {
	root := NewRoot(Defaults())
	for _, f := range []string{"store", "projects", "sessions"} {
		if root.PersistentFlags().Lookup(f) == nil {
			t.Errorf("--%s is not a persistent flag", f)
		}
		for _, c := range root.Commands() {
			if c.Name() == "help" || c.Name() == "completion" {
				continue
			}
			if c.InheritedFlags().Lookup(f) == nil {
				t.Errorf("%s does not inherit --%s", c.Name(), f)
			}
		}
	}
}

// exitCode is the mapping the binary applies, so it is worth its own table: the golden harness and
// Execute both route through it, and a change here changes both at once.
func TestExitCodeMapping(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{nil, 0},
		{usagef("you typed it wrong"), 2},
		{errors.New(`unknown command "x" for "telepathy"`), 2},
		{errors.New("accepts 1 arg(s), received 0"), 2},
		{errNoRipgrep, 1},
		{errors.New("catalogue: no such table: v_nothing"), 1},
	} {
		if got := exitCode(tc.err); got != tc.want {
			t.Errorf("exitCode(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}

// THE HELP OVERRIDE IS DISPLAY ONLY.
//
// The three shared flags print the RULE they resolve by rather than this machine's answer to it,
// because a help text carrying absolute paths out of the caller's home directory cannot be a
// contract. That is done by rewriting pflag's DefValue, which is one field away from the value the
// flag actually starts with — and if it ever became the real default, the tool would look for its
// store at a literal path called "$XDG_STATE_HOME/..." and report an empty catalogue on a machine
// that has one.
func TestTheDefaultShownIsNotTheDefaultUsed(t *testing.T) {
	env := Defaults()
	root := NewRoot(env)
	shown := root.PersistentFlags().Lookup("store").DefValue
	if !strings.Contains(shown, "$XDG_STATE_HOME") {
		t.Errorf("--store no longer documents the rule it resolves by: %q", shown)
	}
	if env.Store == "" || strings.Contains(env.Store, "$") {
		t.Fatalf("the resolved store path is %q — the display override reached the real value", env.Store)
	}
	if !filepath.IsAbs(env.Store) {
		t.Errorf("the resolved store path is not absolute: %q", env.Store)
	}
}

// A PATTERN THAT DOES NOT COMPILE IS NOT AN EMPTY RESULT.
//
// ripgrep has three exit codes and this verb used to collapse them into two: 0 matched, 1 matched
// nothing, 2 DID NOT RUN. `find 'Mint('` took the third path — unclosed group — and printed "no
// transcript contains", exit 0. That is the same plausible zero the missing-ripgrep branch ten
// lines above was written to refuse, reached by a different road.
func TestFindRefusesAPatternThatDoesNotCompile(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed, so `find` cannot be exercised here — a skipped assertion, not a passing one")
	}
	h := newHarness(t)
	out, errOut, code := h.run(t, "find", "--regex", "Mint(")
	if code == 0 {
		t.Fatalf("an uncompilable pattern was reported as a completed search:\n%s", out)
	}
	if strings.Contains(out, "no transcript contains") {
		t.Error("a search that never ran was reported as a search that found nothing")
	}
	// The reason reaches the caller. --no-messages used to swallow it, so even the refusal
	// could not say what was wrong with the pattern.
	if !strings.Contains(errOut, "unclosed group") {
		t.Errorf("the refusal does not relay ripgrep's reason: %q", errOut)
	}
}

// A PATTERN IS ATTRIBUTED WHERE RIPGREP MATCHED IT, like a literal. `\<` is ripgrep's word start,
// which Go's regexp misreads — so this also proves no second engine re-finds the match.
func TestFindRegexReportsWhoSpoke(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed, so `find` cannot be exercised here — a skipped assertion, not a passing one")
	}
	h := newHarness(t)
	out, errOut, code := h.run(t, "find", "--regex", `\<conduct sect`, "--in", "peer")
	if code != 0 {
		t.Fatalf("exit %d, stderr:\n%s", code, errOut)
	}
	for _, id := range []string{gammaID, deltaID} {
		if !strings.Contains(out, catalogue.Short(id)) {
			t.Errorf("the peer row for %s is missing:\n%s", catalogue.Short(id), out)
		}
	}
	if strings.Contains(out, "none has a hit") {
		t.Errorf("a pattern found no peer where the literal finds two:\n%s", out)
	}
}

// THE CALLER'S RIPGREP CONFIG DOES NOT REACH find. A config holding --ignore-case would make a
// literal's matches depend on the box it runs on — and any flag there can move the offsets.
func TestFindIgnoresRipgrepConfig(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed, so `find` cannot be exercised here — a skipped assertion, not a passing one")
	}
	h := newHarness(t)
	cfg := filepath.Join(t.TempDir(), "ripgreprc")
	if err := os.WriteFile(cfg, []byte("--ignore-case\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RIPGREP_CONFIG_PATH", cfg)
	out, errOut, code := h.run(t, "find", "CONDUCT SECTION")
	if code != 0 {
		t.Fatalf("exit %d, stderr:\n%s", code, errOut)
	}
	if !strings.Contains(out, "no transcript contains") {
		t.Errorf("a ripgrep config file changed what find matched:\n%s", out)
	}
}

// THE TERM IS A LITERAL BY DEFAULT, which is what removes the class above for the common case:
// `Mint(` is a reasonable thing to search a Go corpus for and is not a valid regex. The fixture
// contains it as a Grep pattern precisely so this can be asserted.
func TestFindTreatsTheTermAsALiteralUnlessAsked(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed, so `find` cannot be exercised here — a skipped assertion, not a passing one")
	}
	h := newHarness(t)
	out, errOut, code := h.run(t, "find", "Mint(")
	if code != 0 {
		t.Fatalf("a literal containing a regex metacharacter was refused: exit %d, %s", code, errOut)
	}
	if strings.Contains(out, "no transcript contains") {
		t.Errorf("the literal was not found, so it was compiled as a pattern after all:\n%s", out)
	}
	// And --regex genuinely changes the meaning, rather than being a flag that does nothing.
	// A dot matches any character as a pattern and only itself as a literal.
	asPattern, _, _ := h.run(t, "find", "--regex", "Mint.")
	asLiteral, _, _ := h.run(t, "find", "Mint.")
	if asPattern == asLiteral {
		t.Errorf("--regex did not change how the term is interpreted:\n%s", asPattern)
	}
}
