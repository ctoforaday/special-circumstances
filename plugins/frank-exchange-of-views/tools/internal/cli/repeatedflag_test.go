package cli

import (
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A SINGLE-VALUE FLAG STATED TWICE IS REFUSED, BY NAME AND WITH BOTH VALUES, AND NOTHING IS RECORDED.
//
// MEASURED on universe-m37: logic passed `mint --problem` twice, and the record holds the second with
// nothing said, while `close --anchor` twice was refused. These go through the binary's own sequence,
// so deleting the check from Invoke fails them.
func TestAFlagStatedTwiceIsRefusedNamingBothValues(t *testing.T) {
	mint := func(extra ...string) []string {
		return append([]string{"mint", "--seat-id", lensSeat, "--class", "spec_underspecification", "--fix", "f",
			"--check-kind", "document", "--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "Costs rose sharply."}, extra...)
	}
	for _, c := range []struct {
		name        string
		args        []string
		path, flag  string
		first, then string
		do          string
	}{
		{"free text, as m37 did", mint("--problem", "a first complaint", "--problem", "another complaint"),
			"mint", "problem", "a first complaint", "another complaint", "pass it once"},
		{"the value inline", mint("--problem=a first complaint", "--problem", "another complaint"),
			"mint", "problem", "a first complaint", "another complaint", "pass it once"},
		{"the same value twice", mint("--problem", "a complaint", "--problem", "a complaint"),
			"mint", "problem", "a complaint", "a complaint", "pass it once"},
		{"a grade", mint("--problem", "p", "--severity", "high"),
			"mint", "severity", "low", "high", "pass it once"},
		{"a list holds one list", []string{"close", "--seat-id", lensSeat, "--anchor", "F-00000001", "--anchor", "F-00000002"},
			"close", "anchor", "F-00000001", "F-00000002", "comma-separated list"},
		{"a verb under a group", []string{"motion", "docket", "file", "--seat-id", "blue-respond", "--reason", "one", "--reason", "two"},
			"motion docket file", "reason", "one", "two", "pass it once"},
		{"the root's own flag", []string{"show", "board", "--seat-id", lensSeat, "--seat-id", "red-chair"},
			"show board", "seat-id", lensSeat, "red-chair", "pass it once"},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := actFixture(t)
			before := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_MINT)
			args := append(append([]string{}, c.args...), "--run", runDir)
			stdout, stderr, err := runSeen(t, args...)
			if err == nil {
				t.Fatalf("`%s` was accepted", strings.Join(args, " "))
			}
			if stdout != "" {
				t.Errorf("the refusal wrote to stdout: %s", stdout)
			}
			if !strings.HasPrefix(stderr, InvokedAs()+": "+c.path+": --"+c.flag+" was given twice") {
				t.Errorf("the refusal does not open with the flag under its path:\n%s", stderr)
			}
			for _, want := range []string{strconv.Quote(c.first), strconv.Quote(c.then), c.do, "`" + InvokedAs() + " " + c.path + " --help`"} {
				if !strings.Contains(stderr, want) {
					t.Errorf("the refusal does not say %s:\n%s", want, stderr)
				}
			}
			if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != before {
				t.Errorf("the record went from %d mints to %d on a refused call", before, n)
			}
			jsonOut, _, err := runSeen(t, append(args, "--json")...)
			if err == nil {
				t.Fatal("--json: the call was accepted")
			}
			if env := decodeEnvelope(t, jsonOut); env.Code != "validation" || env.OK || !strings.Contains(env.Error, strconv.Quote(c.then)) {
				t.Errorf("envelope code %q ok %v error %q, want validation false naming both", env.Code, env.OK, env.Error)
			}
		})
	}
}

// THE LOG NAMES THE FLAG AND NEVER THE WORDS, as it does for every flag a refused call passed.
func TestARepeatedFlagIsLoggedByNameAndNotByValue(t *testing.T) {
	runDir := actFixture(t)
	const value = "a distinctive complaint no entry may carry"
	if _, err := runAt(t, "mint", "--run", runDir, "--seat-id", lensSeat, "--problem", value, "--problem", "another"); err == nil {
		t.Fatal("the repeat was accepted")
	}
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	var entry string
	for _, e := range evs {
		if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok && l.GetType() == recordpb.LogType_LOG_TYPE_REFUSAL {
			entry = l.GetText()
		}
	}
	if !strings.Contains(entry, "--problem was given twice") || !strings.Contains(entry, "`mint`") {
		t.Errorf("the refusal's log entry does not name the verb and the flag: %q", entry)
	}
	if strings.Contains(entry, value) {
		t.Errorf("the log entry carries the seat's words: %q", entry)
	}
}

// A CALL STATING EACH FLAG ONCE IS NOT TOUCHED by the scan in front of it: the values it parsed are the
// parse's own, and what follows a `--` is a computation's command, in whatever words it has.
func TestACallStatingEachFlagOnceIsLeftToTheParse(t *testing.T) {
	runDir := citedRun(t)
	dir := script(t, runDir, "args.js", "console.log(process.argv.slice(2).join(' '));")
	out, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue.",
		"--dir", dir, "--reason", "the command's own flags repeat", "--", "node", "args.js", "--dir", "x", "--dir", "y")
	if err != nil {
		t.Fatalf("a command's own repeated words were read as this verb's flags: %v\n%s", err, out)
	}
}

// EVERY FLAG OF EVERY VERB ON EVERY SURFACE IS HELD, read off the built trees: a verb grown tomorrow
// is covered without a row, because the check stands in front of the parse and not on the flag.
func TestEverySingleValueFlagOfEveryVerbRefusesASecondStatement(t *testing.T) {
	flagsSeen := 0
	for _, root := range AllRoots() {
		walk(root, func(c *cobra.Command, path []string) {
			c.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if _, appends := f.Value.(pflag.SliceValue); appends {
					return
				}
				flagsSeen++
				argv := append(append([]string{}, path...), "--"+f.Name+"=one", "--"+f.Name+"=two")
				leaf, err := refuseRepeatedFlagNoted(root, argv)
				if err == nil || leaf != c || !strings.Contains(err.Error(), "--"+f.Name+" was given twice") {
					t.Errorf("%v --%s stated twice answered %v from %s", path, f.Name, err, leaf.CommandPath())
				}
			})
		})
	}
	if flagsSeen < 100 {
		t.Errorf("the walk reached %d flags: the trees it reads are not the surfaces", flagsSeen)
	}
}
