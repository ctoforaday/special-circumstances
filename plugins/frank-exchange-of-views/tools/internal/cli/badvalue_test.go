package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A VALUE A FLAG DOES NOT TAKE, A WORD WITH NO FLAG AND A FLAG GROUP VIOLATED ARE EACH REFUSED IN A
// FEW LINES, as a flag left out and a flag the verb does not have are.
//
// MEASURED on the wave B smoke: a lens passed `finding --about-kind ""` and read 5 776 bytes — the
// parser's wrapper and the flag's own sentence each naming the flag and the value, then the verb's
// whole page. What answers the call is the flag, what it was given and what it takes.

// assertShortRefusal reads what every one of these shares: the line under the tool's name and the
// path, the page named and not printed, and one text in both forms.
func assertShortRefusal(t *testing.T, seatID, path string, args []string, opens string) string {
	t.Helper()
	runDir := actFixture(t)
	args = append(args, "--run", runDir, "--seat-id", seatID)
	stdout, stderr, err := runSeen(t, args...)
	if err == nil {
		t.Fatalf("`%s` was accepted", strings.Join(args, " "))
	}
	if stdout != "" {
		t.Errorf("the refusal wrote to stdout: %s", stdout)
	}
	head := InvokedAs() + ": " + path + ": "
	if !strings.HasPrefix(stderr, head+opens) {
		t.Errorf("the refusal does not open %q:\n%s", head+opens, stderr)
	}
	if !strings.Contains(stderr, "`"+InvokedAs()+" "+path+" --help`") {
		t.Errorf("the refusal does not name the page that lists every flag:\n%s", stderr)
	}
	if len(stderr) >= 1500 || strings.Contains(stderr, "Usage:\n") || strings.Contains(stderr, "\nFlags:") {
		t.Errorf("the refusal is %d bytes or carries the page — it is a few lines and names the page:\n%s", len(stderr), stderr)
	}
	jsonOut, _, err := runSeen(t, append(args, "--json")...)
	if err == nil {
		t.Fatal("--json: the call was accepted")
	}
	env := decodeEnvelope(t, jsonOut)
	if env.Code != "validation" || env.OK {
		t.Errorf("envelope code %q ok %v, want validation false", env.Code, env.OK)
	}
	human := strings.TrimSuffix(strings.TrimPrefix(stderr, head), "\n")
	if env.Error != human {
		t.Errorf("the two forms differ.\n human: %s\n  json: %s", human, env.Error)
	}
	return human
}

func TestAnEnumeratedFlagsBadValueIsRefusedWithItsSet(t *testing.T) {
	said := assertShortRefusal(t, "red-lens-evidence", "finding", []string{"finding", "--about-kind", "", "--reason", "r"},
		"--about-kind takes one of section|inquiry|gap|contradiction and was given nothing. ")
	// THE FLAG AND THE VALUE ONCE: the parser's wrapper names both in front of a sentence that does.
	for _, twice := range []string{"invalid argument", `(got "")`, "nothing was passed, and"} {
		if strings.Contains(said, twice) {
			t.Errorf("the refusal carries %q — the parser's wrapper, or the flag's old sentence under it:\n%s", twice, said)
		}
	}
	if n := strings.Count(said, "--about-kind"); n != 1 {
		t.Errorf("the refusal names --about-kind %d times, want once:\n%s", n, said)
	}
	// WHAT IT ACCEPTS, value by value, from the table the page's own menu renders.
	for _, v := range record.MustEnum("finding", "about_kind").Values {
		if !strings.Contains(said, v.Name+"  ") || !strings.Contains(said, v.Means) {
			t.Errorf("the refusal does not carry `%s` with what it is for (%q):\n%s", v.Name, v.Means, said)
		}
	}

	near := assertShortRefusal(t, "red-lens-evidence", "finding", []string{"finding", "--about-kind", "Section", "--about", "x", "--reason", "r"},
		`--about-kind takes one of section|inquiry|gap|contradiction, not "Section": it differs from "section" only in case. `)
	if strings.Contains(near, "invalid argument") {
		t.Errorf("the near miss is refused under the parser's wrapper:\n%s", near)
	}
}

func TestABadValueIsRefusedWithTheFlagsOwnLine(t *testing.T) {
	said := assertShortRefusal(t, "red-lens-evidence", "finding", []string{"finding", "--severity", "huge", "--quote", "q", "--reason", "r"},
		`--severity "huge" is not a grade`)
	severity := cmdAt(NewRootFor("red-lens-evidence"), []string{"finding"}).Flags().Lookup("severity")
	if !strings.Contains(said, "--severity grade ") || !strings.Contains(said, severity.Usage) {
		t.Errorf("the refusal does not carry --severity's own usage line (%q):\n%s", severity.Usage, said)
	}
	if strings.Contains(said, "--likelihood") || strings.Contains(said, "invalid argument") {
		t.Errorf("the refusal lists a flag the call got right, or the parser's wrapper:\n%s", said)
	}

	// A STANDARD-LIBRARY VALUE refuses in the library's words, which name a Go function.
	lib := assertShortRefusal(t, "red-lens-evidence", "show report", []string{"show", "report", "--window", "abc"},
		`--window takes an int, not "abc"`)
	if strings.Contains(lib, "strconv") || strings.Contains(lib, "invalid argument") {
		t.Errorf("the refusal carries the library's parse error:\n%s", lib)
	}
	if !strings.Contains(lib, "--window int ") {
		t.Errorf("the refusal does not carry --window's own usage line:\n%s", lib)
	}
}

func TestAStrayWordIsRefusedWithTheVerbsFlags(t *testing.T) {
	said := assertShortRefusal(t, "red-lens-evidence", "finding", []string{"finding", "stray", "--quote", "q", "--reason", "r"},
		"takes no arguments; got \"stray\"\nThis verb's flags: --about, --about-kind, ")
	if !strings.Contains(said, "--severity.") {
		t.Errorf("the flags a value can go behind are not all listed:\n%s", said)
	}
}

// NO SEAT VERB DECLARES A GROUP A CALL CAN VIOLATE NOW: mint's [problem reason] was the last, and
// it went with the --reason-as-problem fallback; class new's and close's groups are each also
// required flag by flag, which cobra refuses first. The refusal is held on a command built here,
// so a group added to a verb later is answered the way this says.
func TestAFlagGroupViolatedIsRefusedWithItsFlagsLines(t *testing.T) {
	c := &cobra.Command{Use: "probe", RunE: func(*cobra.Command, []string) error { return nil }}
	c.Flags().String("problem", "", "what is wrong")
	c.Flags().String("reason", "", "why it matters")
	c.Flags().String("fix", "", "the fix")
	c.MarkFlagsOneRequired("problem", "reason")
	c.SetArgs([]string{"--fix", "f"})
	c.SilenceErrors, c.SilenceUsage = true, true
	err := c.Execute()
	if err == nil {
		t.Fatal("cobra accepted a call that set no flag of a one-required group")
	}
	r := seat.RefuseFlagGroup(c, err.Error(), err)
	if r == nil {
		t.Fatal("RefuseFlagGroup found no violated group")
	}
	said := seat.Diagnosis(r)
	if !strings.HasPrefix(said, "at least one of the flags in the group [problem reason] is required") {
		t.Errorf("the refusal does not open with cobra's sentence:\n%s", said)
	}
	for _, name := range []string{"problem", "reason"} {
		if u := c.Flags().Lookup(name).Usage; !strings.Contains(said, "--"+name+" ") || !strings.Contains(said, u) {
			t.Errorf("the refusal does not carry --%s's usage line (%q):\n%s", name, u, said)
		}
	}
	if strings.Contains(said, "--fix ") {
		t.Errorf("the refusal lists --fix, which is in no group:\n%s", said)
	}
}
