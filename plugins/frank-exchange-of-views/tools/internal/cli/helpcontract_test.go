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
	walk(newRoot(), func(c *cobra.Command, path []string) {
		if !c.Runnable() || len(path) < 2 {
			return
		}
		role := path[0]
		if !isSeatRole(role) {
			return // operator commands take a run directory rather than a seat
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
					args = append(args, "--"+o.Name, placeholderFor(o, path))
				})
				_, err := run(t, args...)
				if err == nil {
					t.Fatalf("--%s says REQUIRED in its help and the verb ran without it.\n\nusage: %s\n\nA seat reading that supplies it and nothing is wrong; a seat that misses it records an event with a hole, and the hole surfaces rounds later as an empty section nobody can trace.", f.Name, f.Usage)
				}
				if !strings.Contains(err.Error(), f.Name) {
					t.Errorf("--%s was refused and the message does not NAME it:\n\n%v\n\nA seat that cannot tell which flag it missed guesses, and the guess costs a round.", f.Name, err)
				}
			})
		})
	})

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
	runDir := t.TempDir()

	for _, tc := range []struct {
		name string
		args []string
		says string
	}{
		{"a verb outside the role", []string{"lens", "mint", "--seat-id", "red-lens-r1-L1"}, `verb "mint" is outside`},
		{"a role with no verb", []string{"blue"}, "verb is required"},
		{"an unknown top-level command", []string{"frobnicate"}, `no command named "frobnicate"`},
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

func seatFor(role string) string {
	return map[string]string{
		"lens": "red-lens-r1-L1", "merge": "red-merge-r1",
		"blue": "blue-respond-r1", "bench": "judge-r1",
	}[role]
}

// placeholderFor supplies a value a flag will ACCEPT, so the refusal under test is the missing
// one rather than a rejected neighbour. Enum flags need a legal member; everything else takes a
// token.
func placeholderFor(f *pflag.Flag, path []string) string {
	// AN ID MUST NAME SOMETHING THAT EXISTS, or the reference check fires first and masks the
	// refusal under test. Which id depends on what the verb points AT — a gap, an avenue, or a
	// motion — and getting that wrong made this gate report every bench-opinion flag as
	// unnamed when the real refusal was about a gap called "placeholder".
	if f.Name == "script" {
		return scriptPath
	}
	if f.Name == "id" {
		joined := strings.Join(path, " ")
		switch {
		case strings.HasPrefix(joined, "motion direction"):
			return "A1"
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
	case "severity", "likelihood", "impact", "cx", "proposed":
		return "medium"
	case "check-kind":
		return "document"
	case "existence":
		return "verified"
	case "petition-class":
		return "safety"
	case "dimension":
		return "severity"
	case "as":
		return "closed"
	case "confidence":
		return "high"
	}
	return "placeholder"
}

func seatRunForContracts(t *testing.T) string {
	t.Helper()
	runDir := t.TempDir()
	for _, s := range []struct{ role, id string }{
		{"lens", "red-lens-r1-L1"}, {"merge", "red-merge-r1"},
		{"blue", "blue-respond-r1"}, {"bench", "judge-r1"},
	} {
		if _, err := run(t, s.role, "register", "--run", runDir, "--seat-id", s.id); err != nil {
			t.Fatalf("register %s: %v", s.id, err)
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
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "contract-seed", "--class", "self-attestation",
		"--problem", "p", "--fix", "f",
		"--check", "c", "--check-kind", "document",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--cx", "low",
		"--existence", "verified", "--reason", "the gap a probe's --id names"); err != nil {
		t.Fatalf("seed gap: %v", err)
	}
	if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--line", "a seeded line", "--hypothesis", "it would settle something"); err != nil {
		t.Fatalf("seed avenue: %v", err)
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
	if _, err := run(t, "blue", "prove", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--location", "a quoted sentence", "--script", scriptPath,
		"--reason", "the computation a probe re-runs"); err != nil {
		t.Fatalf("seed proof: %v", err)
	}
	b, err := record.BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range b.Events {
		if e.Type == "proof" && e.Payload.Str("sha256") != "" {
			proofSHA = e.Payload.Str("sha256")
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
// shape: `--as sound|unsound`, `--as supports|refutes|…`, `--status proposed|pursued|…`.
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

	var checked int
	walk(newRoot(), func(c *cobra.Command, path []string) {
		registered := enumhelp.Registered(c)
		for _, m := range restated.FindAllStringSubmatch(c.Short, -1) {
			flag, listed := m[1], strings.Split(m[2], "|")
			var want []string
			switch {
			case registered[flag] != nil:
				want = record.Names(registered[flag])
			case flag == "view":
				want = ViewNames()
			default:
				continue // not an enum this command declares: a shape hint, not a set
			}
			checked++
			real := map[string]bool{}
			for _, v := range want {
				real[v] = true
			}
			for _, v := range listed {
				if v == "" || v == "..." {
					continue
				}
				if !real[v] {
					t.Errorf("`%s` summarises --%s as %q, and %q is NOT in the real set (%s).\n\nThe summary is the line a seat reads first — in Available Commands and now in every refusal — so a value offered here and refused by the tool is a set that does not exist, taught at the moment a seat is deciding what to type.",
						strings.Join(path, " "), flag, m[2], v, strings.Join(want, "|"))
				}
			}
		}
	})
	if checked == 0 {
		t.Fatal("no summary restated any known set — either the convention is gone or the walk is broken, and a walk that finds nothing passes this forever")
	}
}
