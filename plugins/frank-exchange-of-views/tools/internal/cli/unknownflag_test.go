package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A SEAT THAT TYPED THE RECORD'S WORD IS TOLD THE FLAG'S, and only when that is actually true.
//
// Measured on the 2026-09-20 smoke: a lens typed `--complexity_cost` at `mint` TWELVE times, each
// costing a turn. That is the key the record stores and shows the seat in a gap's own JSON before
// it mints, so the seat used the vocabulary it had just been handed — the most defensible mistake
// available. The translation was already in internal/flags and was consulted only when the TOOL
// named a field, never when a SEAT did.
func TestAnUnknownFlagNamesTheFlagTheSeatMeant(t *testing.T) {
	runDir := seatRun(t)

	out, err := run(t, "mint", "--run", runDir, "--seat-id", "red-lens-evidence", "--complexity_cost", "high")
	if err == nil {
		t.Fatal("mint --complexity_cost was accepted; it is not a flag")
	}
	msg := err.Error() + out
	// THE ASSERTION MUST NOT BE SATISFIED BY COBRA'S OWN MESSAGE, and the obvious one is:
	// "unknown flag: --complexity_cost" already contains the substring "--complexity", because the
	// wrong word has the right word as a PREFIX. A mutation that deleted the translation entirely
	// passed against that check. What is pinned instead is the sentence only the translation can
	// produce.
	for _, want := range []string{
		"the flag that sets it is --complexity.",
		"the record stores that field as `complexity_cost`",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q — a seat is left to guess a second time:\n%s", want, msg)
		}
	}
}

// AND IT DOES NOT INVENT ONE. A key with no entry gets cobra's plain refusal: a suggested spelling
// the parser rejects is worse than no suggestion, which is the defect the payload map was built to
// stop in the first place.
func TestAnUnknownFlagWithNoTranslationIsNotGivenOne(t *testing.T) {
	runDir := seatRun(t)

	out, err := run(t, "mint", "--run", runDir, "--seat-id", "red-lens-evidence", "--not_a_real_field", "x")
	if err == nil {
		t.Fatal("an invented flag was accepted")
	}
	if strings.Contains(err.Error()+out, "the flag that sets it is") {
		t.Errorf("a flag with no known translation was given one anyway:\n%s", err.Error()+out)
	}
}

// AND THE SUGGESTION IS SCOPED TO THE FAILING VERB. The payload map is GLOBAL while payload keys
// are not globally unique — its own comment says so — so a suggestion is made only where the verb
// that refused actually carries that flag. `log` has no --complexity, and must not be told it does.
func TestTheSuggestionIsNotMadeForAFlagTheVerbDoesNotHave(t *testing.T) {
	runDir := seatRun(t)

	out, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-evidence", "--complexity_cost", "high")
	if err == nil {
		t.Fatal("log --complexity_cost was accepted")
	}
	if strings.Contains(err.Error()+out, "the flag that sets it is --complexity") {
		t.Errorf("log was told to type --complexity, which it does not have:\n%s", err.Error()+out)
	}
}

// A FLAG CARRIED FROM A SIBLING VERB IS ANSWERED FROM THE TREE, IN A FEW LINES.
//
// Measured on the 2026-10-09 smoke: blue typed `retire --key` four times in one batch, having used
// `--key` on fourteen edits, and a lens typed `finding --problem` twice, fourteen seconds after
// `near-match --problem` was accepted. Each refusal was the verb's whole page — 3 506 bytes per
// retire, 9 029 for the two findings — and none said the one thing the seat had wrong: the flag
// is a real one, on the verbs beside this one.
//
// The refusal says which verbs on the ASKING SEAT's surface take the flag, what this verb takes,
// and where the page is. It is the same text in both forms, under 1 KB, and it accepts nothing:
// the call is refused and the verb writes no event.
func TestAnUnknownFlagIsAnsweredFromTheSeatsOwnTree(t *testing.T) {
	runDir := actFixture(t)
	for _, tc := range []struct {
		name, seat string
		args       []string
		verb, role string
		// writes is the event type the verb writes when it runs, "" for a verb that writes none.
		writes string
		says   []string
		silent []string
	}{
		{name: "a handle on a verb with none", seat: "blue-respond",
			args: []string{"retire", "--key", "R1", "--quote", "a claim that is no longer in the report.", "--reason", "r"},
			verb: "retire", role: "blue", writes: "retire",
			says: []string{"unknown flag: --key", "On this surface `cite`, `edit` and `prove` take it.",
				"This verb's flags: --anchor, --new, --quote, --reason.", "`" + InvokedAs() + " retire --help`"},
			silent: []string{"`finding`", "`mint`"}}, // the lens's verbs take a handle too, and are not blue's
		{name: "the neighbour's problem flag", seat: "red-lens-evidence",
			args: []string{"finding", "--quote", "§1 first — a finding sits in sec 1 here.", "--severity", "low", "--likelihood", "low", "--impact", "low", "--reason", "r", "--problem", "p"},
			verb: "finding", role: "lens", writes: "finding",
			says: []string{"unknown flag: --problem", "On this surface `mint` and `near-match` take it.",
				"This verb's flags: --about, --about-kind, --impact, --key, --likelihood, --quote, --reason, --severity.", "`" + InvokedAs() + " finding --help`"}},
		{name: "a flag nothing takes", seat: "red-lens-evidence",
			args: []string{"close", "--nonesuch", "x"},
			verb: "close", role: "lens", writes: "close",
			says: []string{"unknown flag: --nonesuch", "no verb on this surface takes it", "This verb's flags: --anchor, --as, --id,", "`" + InvokedAs() + " close --help`"}},
		{name: "the operator's surface", seat: "operator",
			args: []string{"graph", "--nonesuch", "x"},
			verb: "graph", role: "",
			says: []string{"unknown flag: --nonesuch", "This verb's flags: --format.", "`" + InvokedAs() + " graph --help`"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			written := func() int {
				if tc.writes == "" {
					return 0
				}
				return len(eventsOfWord(t, runDir, tc.writes))
			}
			before := written()
			args := append(append([]string{}, tc.args...), "--run", runDir, "--seat-id", tc.seat)
			stdout, stderr, err := runSeen(t, args...)
			if err == nil {
				t.Fatal("a flag the verb does not have was accepted")
			}
			if stdout != "" {
				t.Errorf("the refusal wrote to stdout: %s", stdout)
			}
			for _, w := range tc.says {
				if !strings.Contains(stderr, w) {
					t.Errorf("the refusal does not say %q:\n%s", w, stderr)
				}
			}
			for _, w := range tc.silent {
				if strings.Contains(stderr, w) {
					t.Errorf("the refusal names %s, which is not on this seat's surface:\n%s", w, stderr)
				}
			}
			if len(stderr) >= 1024 || strings.Contains(stderr, "\nUsage:") || strings.Contains(stderr, "\nFlags:") {
				t.Errorf("the refusal is %d bytes or carries the page — it is a few lines and names the page:\n%s", len(stderr), stderr)
			}

			jsonOut, _, err := runSeen(t, append(args, "--json")...)
			if err == nil {
				t.Fatal("--json: a flag the verb does not have was accepted")
			}
			env := decodeEnvelope(t, jsonOut)
			if env.Verb != tc.verb || env.Role != tc.role || env.Code != "validation" || env.OK {
				t.Errorf("envelope verb %q role %q code %q ok %v, want %q %q validation false", env.Verb, env.Role, env.Code, env.OK, tc.verb, tc.role)
			}
			if human := strings.TrimSuffix(strings.TrimPrefix(stderr, InvokedAs()+": "+tc.verb+": "), "\n"); env.Error != human {
				t.Errorf("the two forms differ.\n human: %s\n  json: %s", human, env.Error)
			}
			if n := written() - before; n != 0 {
				t.Errorf("the refused calls wrote %d %s event(s) — a flag the verb does not have must not reach the verb", n, tc.writes)
			}
		})
	}
}

// A VERB'S FLAG TYPED BEFORE THE VERB IS NAMED AS THAT, and the call is refused, not reordered. On
// the #1565 smoke an adversary lens typed `--key adv-f-tldr finding …` five times in one call; the
// command name is read at the first word that is no flag's, so "adv-f-tldr" was the command, and each
// refusal said "no command named adv-f-tldr exists … there is no seat to switch to" above the whole
// surface's help — naming neither the flag nor the word it had taken as the name.
func TestAVerbsFlagBeforeTheVerbIsNamedAndRefused(t *testing.T) {
	runDir := seatRun(t)
	registerLensOnce(t, runDir)
	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING)

	out, err := run(t, "--run", runDir, "--seat-id", lensSeat, "--key", "adv-f-tldr", "finding",
		"--quote", "x", "--reason", "y")
	if err == nil {
		t.Fatal("a verb's flag before the verb was accepted — the tool chose which verb it belonged to")
	}
	msg := err.Error() + out
	for _, want := range []string{
		"--key came before the verb",
		`"adv-f-tldr", was read as the command name`,
		"`finding`", // a verb on this surface that takes --key, read from the tree
		"Nothing ran.",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "no seat to switch to") || strings.Contains(msg, "Usage:") {
		t.Errorf("the refusal still answers as an unknown command, or prints the whole surface:\n%s", msg)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING); n != before {
		t.Errorf("a refused call recorded a finding (%d -> %d)", before, n)
	}

	// A WORD AFTER A ROOT FLAG'S VALUE IS STILL A COMMAND NAME, and an unknown one is still told so.
	if _, err := run(t, "--run", runDir, "--seat-id", lensSeat, "banana"); err == nil || !strings.Contains(err.Error(), `no command named "banana"`) {
		t.Errorf("an unknown command after root flags is not refused as one: %v", err)
	}
}
