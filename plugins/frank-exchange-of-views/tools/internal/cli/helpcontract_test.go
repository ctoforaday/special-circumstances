package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A PROMISE IN HELP TEXT IS A CONTRACT, AND EVERY CONTRACT HAS A TEST.
//
// Help is the only teacher a seat has, and it is prose — so it can claim anything, and nothing in
// a normal build disagrees. This file makes the claims checkable: where the help says a thing is
// REQUIRED, the tool must refuse without it; where it promises a closed set, the tool must refuse
// outside it (enum_help_test.go); where a refusal offers a list, that list must carry meanings
// (refusalteaches_test.go).
//
// The measured cost of an unchecked claim is not hypothetical. `lens finding --location` promised
// "a section heading plus a QUOTED sentence" and the tool matched the WHOLE string literally
// against the report, so every seat that obeyed the help was refused — four separators tried in
// one sitting, ten consecutive failures, and the sitting spent on the matcher rather than the
// audit. Nothing failed anywhere in this repo, because the help was prose and the prose was wrong.

// requiredInHelp matches the MARKER CONVENTION, not the word.
//
// The first draft matched `required` case-insensitively anywhere in the usage, and `merge mint
// --fix` reads "the required fix, as prose" — where "required" is part of the FIELD NAME
// (required_fix) and claims nothing about the flag. A gate that fires on prose trains its reader
// to skim, so this matches only what markRequired actually writes.
var requiredInHelp = regexp.MustCompile(`REQUIRED\s+—`)

// satisfiedByAnother are flags whose help says REQUIRED and names an alternative in the same
// breath, with what satisfies them. The help is telling the truth in both halves; the gate's
// model — each required flag independently mandatory — is the thing that cannot express it.
var satisfiedByAnother = map[string]string{
	"problem": "--reason, which the verb copies into the problem field. The help says so in the same sentence: \"(or pass it via --reason)\"",
	"row":     "--reason, which manifest-row falls back to when --row is absent. The help says so in the same sentence, for the same reason problem's does",
}

// notEnforcedAtTheFlag are flags whose help says REQUIRED and whose refusal is DELIBERATELY
// somewhere else, each with where. The distinction matters: a flag required by the record's
// write path is still required, and a seat still cannot proceed without it — what differs is
// which layer says so, and a test that demanded one layer would force the check to move.
var notEnforcedAtTheFlag = map[string]string{
	// The prose field is required by the verbs that carry an argument, and the refusal is the
	// verb's own (it names what the prose is FOR, which a generic flag check could not).
	"reason": "the verb's RunE, which names what the argument is for rather than only that it is missing",
}

// TestEveryRequiredFlagIsActuallyRefused walks the tree and, for every flag whose help says
// REQUIRED, invokes the verb WITHOUT it and demands a refusal that names the flag.
//
// The failure this prevents: a help string that says REQUIRED beside a flag the tool happily
// omits. A seat reading it supplies the flag and nothing is wrong; a seat that misses it records
// an event with a hole in it, and the hole surfaces rounds later as an empty section.
func TestEveryRequiredFlagIsActuallyRefused(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := seatRunForContracts(t)

	var checked int
	// PER TREE, AND THE ROLE COMES FROM THE TREE. It used to be path[0], because a verb's path
	// began with its role group; the path is now the verb alone and the role is which surface it
	// was found on.
	for role, r := range AllRoots() {
		if !isSeatRole(role) {
			continue // operator commands take a run directory rather than a seat
		}
		walk(r, func(c *cobra.Command, path []string) {
			if !c.Runnable() || len(path) < 1 {
				return
			}
			// TWO SUBTREES ARE STILL OUT OF SCOPE, and now they have to say so.
			//
			// `motion` was excluded by ACCIDENT: it sat at the root, so its path began "motion",
			// isSeatRole said no, and the walk skipped it. Moving it inside each seat's tree
			// brought it into a gate that has never covered it — the subjects carry different
			// required sets, which is why they are subgroups, and `--id` is enforced in the
			// handler rather than marked at the flag.
			//
			// `reproduce` needs a proof ON THE BOARD before its own flags are reached, so the
			// placeholder id produces a refusal about the missing proof rather than the missing
			// flag — a precondition this gate does not build.
			//
			// Both are real gaps in this gate's coverage, named rather than left looking covered.
			if path[0] == "motion" || path[0] == "reproduce" {
				return
			}
			c.Flags().VisitAll(func(f *pflag.Flag) {
				if !requiredInHelp.MatchString(f.Usage) || notEnforcedAtTheFlag[f.Name] != "" || satisfiedByAnother[f.Name] != "" {
					return
				}
				checked++
				t.Run(strings.Join(path, " ")+"/--"+f.Name, func(t *testing.T) {
					args := append(append([]string{}, path...), "--run", runDir, "--seat-id", seatFor(role))
					// --reason is required by most verbs' own RunE rather than by the marker, so
					// it is supplied unconditionally: without it the prose refusal fires first
					// and this gate measures the wrong one.
					if f.Name != "reason" && c.Flags().Lookup("reason") != nil {
						args = append(args, "--reason", "the argument for this act")
					}
					// Supply every OTHER required flag, so the refusal under test is this one's.
					c.Flags().VisitAll(func(o *pflag.Flag) {
						if o.Name == f.Name || !requiredInHelp.MatchString(o.Usage) {
							return
						}
						args = append(args, "--"+o.Name, placeholderFor(c, o, path))
					})
					_, err := run(t, args...)
					if err == nil {
						t.Fatalf("--%s says REQUIRED in its help and the verb ran without it.\n\nusage: %s\n\nA seat reading that supplies it and nothing is wrong; a seat that misses it records an event with a hole, and the hole surfaces rounds later as an empty section nobody can trace.", f.Name, f.Usage)
					}
					if !strings.Contains(err.Error(), f.Name) {
						t.Errorf("--%s was refused and the message does not NAME it:\n\n%v\n\nA seat that cannot tell which flag it missed guesses, and the guess costs a round.", f.Name, err)
					}
					// AND EVERY FLAG THE REFUSAL NAMES MUST EXIST. Checked HERE as well as in
					// TestNoRefusalNamesAFlagThatDoesNotExist because that one invokes each verb bare
					// and therefore only ever sees cobra's "required flag(s) not set" — the record
					// layer's teaching messages, which are the ones that name flags in prose, are
					// reached only by satisfying the earlier refusals, which is what this loop does.
					// Measured: renaming `mint requires --problem` to `--problem-statement` left the
					// bare-invocation gate green.
					assertNamedFlagsExist(t, err, knownFlagNames(t))
				})
			})
		})
	}

	if checked < 15 {
		t.Fatalf("only %d REQUIRED flags checked — the walk is not reaching the tree, and a walk that finds nothing passes this forever", checked)
	}
}

// TestEveryRefusalNamesTheProblemBeforeTheHelp holds the ordering rule.
//
// The first draft of the teaching refusal printed cobra's help and THEN the error, because that is
// the order cobra emits them. For a reader that is backwards: a page of help arrives with no
// statement of what went wrong, and a seat has to infer its own mistake from a list of things it
// did not do. A diagnosis that comes after the remedy is not a diagnosis.
func TestEveryRefusalNamesTheProblemBeforeTheHelp(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := newRun(t)

	for _, tc := range []struct {
		name string
		args []string
		says string
	}{
		{"a verb that is another seat's", []string{"mint", "--seat-id", "red-lens-r1-L1"}, `"mint" is not on your surface`},
		// There is no role level to name, so the old "a role with no verb" case is now a caller
		// with no identity — the same shape of mistake at the level that still exists.
		{"a command with no identity", []string{"mint"}, "--seat-id IS REQUIRED HERE"},
		{"an unknown command", []string{"frobnicate", "--seat-id", "blue-respond-r1"}, `no command named "frobnicate"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := run(t, append(append([]string{}, tc.args...), "--run", runDir)...)
			if err == nil {
				t.Fatal("expected a refusal")
			}
			text := out + err.Error()
			idx := strings.Index(text, tc.says)
			if idx < 0 {
				t.Fatalf("the refusal never says what went wrong (wanted %q):\n%s", tc.says, text)
			}
			// The diagnosis must come before any help block.
			for _, marker := range []string{"Usage:", "Available Commands:", "Flags:"} {
				if h := strings.Index(text, marker); h >= 0 && h < idx {
					t.Errorf("the help (%q at %d) comes BEFORE the diagnosis (at %d).\n\nA seat reading top-down meets a page of options with no statement of its mistake, and has to infer what it did wrong from what it did not do.", marker, h, idx)
				}
			}
		})
	}
}

func walk(c *cobra.Command, fn func(*cobra.Command, []string)) {
	var rec func(*cobra.Command, []string)
	rec = func(cur *cobra.Command, path []string) {
		for _, sub := range cur.Commands() {
			n := sub.Name()
			if n == "help" || n == "completion" {
				continue
			}
			p := append(append([]string{}, path...), n)
			fn(sub, p)
			rec(sub, p)
		}
	}
	rec(c, nil)
}

func isSeatRole(s string) bool {
	switch s {
	case "lens", "merge", "blue", "bench":
		return true
	}
	return false
}

// seatHolding answers which seat can run this verb, by looking for it. It replaces reading the
// role out of a command PATH, which stopped being possible when the role left the path.
func seatHolding(verb string) string {
	// SEATS FIRST, AND IN A FIXED ORDER. `verify` is a lens verb AND the operator's whole-record
	// cross-check, and `fetch`/`count-claims` sit in both too — so ranging a map returned whichever
	// tree came up first and the answer changed between runs. This asks about SEAT verbs, so the
	// operator's tree is not a candidate at all.
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		for _, c := range NewRootFor(record.SampleSeatOf(role)).Commands() {
			if c.Name() == verb {
				return seatFor(role)
			}
		}
	}
	return ""
}

func seatFor(role string) string {
	return map[string]string{
		"lens": "red-lens-r1-L1", "merge": "red-merge-r1",
		"blue": "blue-respond-r1", "bench": "judge-r1",
	}[role]
}

// placeholderFor supplies a value a flag will ACCEPT, so the refusal under test is the missing
// one rather than a rejected neighbour. Enum flags need a legal member; everything else takes a
// token.
func placeholderFor(c *cobra.Command, f *pflag.Flag, path []string) string {
	// AN ENUM'S PLACEHOLDER COMES FROM THE COMMAND'S OWN SET. It used to be a switch on the flag
	// WORD — `case "as": return "closed"` — which is the same mistake this vocabulary keeps
	// making one layer up: --as is one word with a value space per verb, so a single answer is
	// right for `close` and wrong for `verify`, and the wrong one made this gate report the flag
	// under test as unnamed when the refusal was about the placeholder.
	if vals := enumhelp.Registered(c)[f.Name]; len(vals) > 0 {
		return vals[0].Name
	}
	// AN ID MUST NAME SOMETHING THAT EXISTS, or the reference check fires first and masks the
	// refusal under test. Which id depends on what the verb points AT — a gap, a line of inquiry, or a
	// motion — and getting that wrong made this gate report every bench-opinion flag as
	// unnamed when the real refusal was about a gap called "placeholder".
	if f.Name == "script" {
		return scriptPath
	}
	if f.Name == "id" {
		// THE SHAPE DECIDES, NOT THE PATH. A flag declares the id kind it accepts
		// (flags.InquiryID() types itself `inquiry-id`), so reading the declaration answers the
		// question the path was being used to guess at. The path form went stale the moment
		// `motion direction` became `motion inquiry` — this branch used to say `motion direction`
		// and silently stopped matching, feeding a gap id to a flag that wanted a line's.
		if f.Value.Type() == "inquiry-id" {
			return "Q1"
		}
		joined := strings.Join(path, " ")
		switch {
		case strings.HasPrefix(joined, "motion") && !strings.HasSuffix(joined, "file"):
			return "M1"
		case joined == "lens reproduce":
			// A PROOF SHA, not a gap id. Passing R1-1 made the proof lookup fail first and
			// the gate reported --as as unnamed when the refusal was about a missing proof.
			return proofSHA
		}
		return "R1-1"
	}
	if vals := strings.Split(f.Value.Type(), "|"); len(vals) > 1 {
		return vals[0]
	}
	switch f.Name {
	case "severity", "likelihood", "impact", "complexity", "proposed":
		return "medium"
	case "check-kind":
		return "document"
	case "dimension":
		return "severity"
	case "url":
		return "https://example.test/source"
	}
	return "placeholder"
}

func seatRunForContracts(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	for _, id := range []string{"red-lens-r1-L1", "red-merge-r1", "blue-respond-r1", "judge-r1"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", id); err != nil {
			t.Fatalf("register %s: %v", id, err)
		}
	}
	seedBlueReport(t, runDir)
	// A REAL SCRIPT ON DISK, because `blue prove` RUNS what --script names.
	scriptPath = filepath.Join(t.TempDir(), "contract-proof.py")
	if err := os.WriteFile(scriptPath, []byte("print('contract seed')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// REAL REFERENTS, so an --id in a probe names something. Without these the reference checks
	// fire before the flag-specific ones and this gate measures the wrong refusal.
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "contract-seed", "--class", "self-attestation",
		"--problem", "p", "--fix", "f",
		"--check", "c", "--check-kind", "document",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low",
		"--reason", "the gap a probe's --id names"); err != nil {
		t.Fatalf("seed gap: %v", err)
	}
	if _, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--reason", "a seeded line", "--hypothesis", "it would settle something"); err != nil {
		t.Fatalf("seed line of inquiry: %v", err)
	}
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low",
		"--reason", "the motion a probe's --id names"); err != nil {
		t.Fatalf("seed motion: %v", err)
	}
	// A RECORDED PROOF, so `lens reproduce --id` names one.
	//
	// The sha comes from the RECORD, not from parsing stdout: a human-readable line is not an
	// interface, and reading a fact out of prose is the shape this suite exists to remove. The
	// first draft regexed a 64-hex out of the message, found nothing, and the gate then reported
	// `--as` as unnamed when the real refusal was a missing --id.
	if _, err := run(t, "prove", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "a quoted sentence", "--script", scriptPath,
		"--reason", "the computation a probe re-runs"); err != nil {
		t.Fatalf("seed proof: %v", err)
	}
	b, err := record.BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range b.Events {
		if pf, ok := recordpb.BodyAs[*recordpb.Proof](e); ok && pf.GetProofSha() != "" {
			proofSHA = pf.GetProofSha()
		}
	}
	if proofSHA == "" {
		t.Fatal("no proof landed on the record — `lens reproduce --id` would then be probed with an empty id, and the gate would measure the wrong refusal")
	}
	return runDir
}

// scriptPath and proofSHA are seeded once by seatRunForContracts and read by placeholderFor.
var (
	scriptPath string
	proofSHA   string
)

// A SUMMARY LINE THAT RESTATES A SET MUST RESTATE IT CORRECTLY.
//
// The `Short` of every verb is what a seat reads FIRST — it is the line in `Available Commands`,
// which is now also the line a refusal prints. Several restate an enum inline to show the flag
// shape: `--as sound|unsound`, `--as supports|refutes|…`, `--as proposed|pursued|…`.
//
// That is a SECOND rendering of a set the flag help already carries properly, and the second one
// is the lossy copy: it has no meanings and nothing regenerates it. The `--view` list in `show`'s
// Short was exactly this — it survived the whole enum-menu change untouched, so a seat reading
// the summary still met twelve bare nouns.
//
// This does not forbid the shape: showing the flag shape in a summary is genuinely useful, and a
// seat that must open a second help page to learn what `--as` takes is a seat that guesses. What
// it forbids is the copy DIVERGING — a Short that offers a value the tool refuses, or omits one it
// accepts, teaches a set that does not exist.
func TestEverySetRestatedInASummaryMatchesTheRealOne(t *testing.T) {
	// THE SET IS RESOLVED PER COMMAND, NOT PER FLAG NAME. `--as` carries a different vocabulary
	// on every verb that uses it — closure classes on `merge close`, dispositions on `bench
	// opinion`, soundness on `lens reproduce` — which is the whole reason record.EnumFields is
	// keyed by event type. A map from flag name to values holds whichever verb was registered
	// last, and the first draft of this test compared `bench opinion --as carried|closed`
	// against `sound|unsound` and reported both values as unreal.
	//
	// enumhelp.Registered answers for THIS command, which is exact by construction.
	restated := regexp.MustCompile(`--([a-z-]+)\s+([a-z][a-zA-Z_-]*(?:\|[a-zA-Z_][a-zA-Z_-]*)+)`)

	// INVERTED, BECAUSE THE CONVENTION IT POLICED IS GONE.
	//
	// This checked that a set restated in a help summary matched the real one — the summary being
	// "the line a seat reads first". Help text no longer restates sets at all: cobra prints an
	// `Enumerated values:` block glossing every value of every enum flag, generated from the flags
	// themselves, and a prose copy beside it is a second version nothing keeps in step. Removing
	// those copies emptied this gate, which tripped its own "a walk that finds nothing" guard
	// rather than passing vacuously — the guard working exactly as written.
	//
	// So it asserts the ABSENCE now, and keeps a live tripwire: the walk must still find commands
	// declaring enums, or the traversal is broken and this proves nothing in either direction.
	enumCommands := 0
	for _, r := range AllRoots() {
		walk(r, func(c *cobra.Command, path []string) {
			registered := enumhelp.Registered(c)
			if len(registered) > 0 {
				enumCommands++
			}
			for _, m := range restated.FindAllStringSubmatch(c.Short+"\n"+c.Long, -1) {
				flag := m[1]
				var want []string
				switch {
				case registered[flag] != nil:
					want = record.Names(registered[flag])
				case flag == "view":
					want = ViewNames()
				default:
					continue // not an enum this command declares: a shape hint, not a set
				}
				t.Errorf("`%s` restates --%s as %q in its help.\n\nCobra prints an Enumerated values block for this flag, glossing each value, derived from the flag itself. A prose copy is a second set nothing compiles against the first — say what the field MEANS and let the block carry the vocabulary. The real set is (%s).",
					strings.Join(path, " "), flag, m[2], strings.Join(want, "|"))
			}
		})
	}
	if enumCommands == 0 {
		t.Fatal("walked every root and found no command declaring an enum flag — the walk is broken, and this gate would report a clean board on a tree it never read")
	}
}

// flagToken matches a `--flag` reference inside a refusal message.
//
// THE LEADING BOUNDARY IS LOAD-BEARING. Without it this matched `--proof` inside the anchor
// token `<!--proof:p-…-->`, and reported `lens reproduce` — whose message is correct — for a
// flag it never named. A seat types a flag after a space, a backtick, a quote or an open
// paren; nothing else introduces one. Go's regexp has no lookbehind, so the boundary is a
// non-capturing alternation and the NAME is group 1.
//
// Trailing punctuation is excluded by the character class, so `--as FAIL`, "`--quote`," and
// "--id." all yield the bare name. A bare `--` is not a name: one name character is required.
var flagToken = regexp.MustCompile("(?:^|[\\s(\"'`])--([a-z][a-z0-9-]*)")

// EVERY FLAG A REFUSAL NAMES MUST BE A FLAG SOMETHING REGISTERS.
//
// FOUR INSTANCES ON ONE BRANCH, each found by a different accident:
//
//	`--as supports-with-bridge`  advertised in help, refused by the write path
//	`--id Q1 --as supported|…`   advertised by inquiry-support after the schema retired both
//	`retire requires --claim`    the FIELD name in a message that must carry the FLAG word
//	`out-of-scope` / `too-thin`  the fuzz typing an enum's field spelling at the flag
//
// The shape is always the same: two spellings of one fact across a boundary, one side moved.
// A seat obeys the message, gets a cobra refusal for a flag nobody accepts, and learns nothing
// about what it actually needed. It is not catchable by reading the message — it reads fine.
//
// This is the CHEAP HALF of the check, and deliberately so: it asks only whether the name exists
// ANYWHERE in the tree, because a refusal legitimately names other verbs' flags ("close them with
// `close --id <id>`"). Scoping each mention to its own verb would need to know which sentences
// are self-referential, and a gate that guesses that would fire on prose. A name nothing at all
// registers is unambiguous, needs no judgement, and is what all four instances were.
func TestNoRefusalNamesAFlagThatDoesNotExist(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := seatRunForContracts(t)

	known := knownFlagNames(t)

	var checked int
	walk(newRoot(), func(c *cobra.Command, path []string) {
		if !c.Runnable() || len(path) < 2 || !isSeatRole(path[0]) {
			return
		}
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			// Invoked with identity and NOTHING else, so whatever the verb requires is what
			// refuses. A verb that runs clean here simply has nothing to say and is skipped.
			_, err := run(t, append(append([]string{}, path...), "--run", runDir, "--seat-id", seatFor(path[0]))...)
			if err == nil {
				return
			}
			checked++
			// THE DIAGNOSIS, NOT THE HELP COBRA STAPLES TO IT. A verb's own usage block lists its
			// flags (fine) and its prose sometimes names a RETIRED one on purpose — spot-check's
			// --reason explains that it absorbed --notes, which is exactly the history a seat
			// needs. Scanning that turns a gate about broken instructions into a gate against
			// explaining anything. TestEveryRefusalNamesTheProblemBeforeTheHelp guarantees the
			// diagnosis comes FIRST, which is what makes this cut safe to make.
			assertNamedFlagsExist(t, err, known)
		})
	})
	if checked < 10 {
		t.Fatalf("only %d verbs refused anything — this gate reads refusals, so a walk that produces none passes it forever", checked)
	}
}

// diagnosisOf returns the part of a refusal that states the problem, cutting cobra's appended
// help. The boundary is the usage block, or the standing footer that precedes it.
func diagnosisOf(err error) string {
	msg := err.Error()
	for _, cut := range []string{"\nUsage:", "\nIf you need a verb or a flag that is not listed here"} {
		if i := strings.Index(msg, cut); i >= 0 {
			msg = msg[:i]
		}
	}
	return msg
}

// knownFlagNames is every flag name the whole tree registers — local, persistent AND inherited.
//
// The first draft collected only local flags and reported `--run`, `--seat-id`, `--json` and
// `--help` on nineteen verbs: the globals every refusal legitimately names. A gate whose own
// vocabulary is incomplete manufactures the exact defect it exists to find.
func knownFlagNames(t *testing.T) map[string]bool {
	t.Helper()
	root := newRoot()
	known := map[string]bool{"help": true}
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) { known[f.Name] = true })
	walk(root, func(c *cobra.Command, path []string) {
		c.Flags().VisitAll(func(f *pflag.Flag) { known[f.Name] = true })
		c.PersistentFlags().VisitAll(func(f *pflag.Flag) { known[f.Name] = true })
		c.InheritedFlags().VisitAll(func(f *pflag.Flag) { known[f.Name] = true })
	})
	if len(known) < 20 {
		t.Fatalf("only %d flags found in the tree — the walk is not reaching it, and an empty vocabulary accepts every message forever", len(known))
	}
	return known
}

// assertNamedFlagsExist reports any `--flag` the refusal's DIAGNOSIS names that nothing registers.
func assertNamedFlagsExist(t *testing.T, err error, known map[string]bool) {
	t.Helper()
	for _, m := range flagToken.FindAllStringSubmatch(diagnosisOf(err), -1) {
		if !known[m[1]] {
			t.Errorf("the refusal names `--%s`, which NO command in the tree registers:\n\n%v\n\n"+
				"A seat that obeys this gets a cobra refusal for a flag nobody accepts, and still does not know what it needed. "+
				"Either the flag was renamed and the message was not, or the message is naming a record FIELD where a seat types a FLAG word.", m[1], err)
		}
	}
}
