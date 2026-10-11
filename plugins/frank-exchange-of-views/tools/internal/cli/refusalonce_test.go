package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// A REFUSAL NAMES ITS COMMAND ONCE, AND THE RENDERER IS WHAT NAMES IT.
//
// seat.RenderRefusal prints `feov-record: <path>: <diagnosis>` and carries the role and the verb as
// fields of the JSON envelope, both read from the command that resolved. A diagnosis that opens
// with its own role or its own path therefore says it twice — `feov-record: finding: lens finding
// requires --reason` — and a seat reading for what was wrong reads past two names to reach it.
//
// Every refusal any test drives through `run` is read against this (cli_test.go), and the
// binary's own output is in TestEveryRefusalKindHasOneEnvelope — so a new refusal written with a
// hand-typed head fails in whichever test first reaches it.

// layerWords are diagnoses that open with a PACKAGE's name where that name is also the verb that
// reaches the package. They are composed below the command layer — `fetch:` by internal/fetchcache,
// `assemble:` by internal/report, `migrate:` by internal/record/migrate — where the word names the
// layer that failed, as `record:` does; #1266's last phase removes them with the rest of the
// record layer's heads. Each row is one leaf and the one head it may open with.
var layerWords = map[string]string{
	"fetch":    "fetch: ",
	"assemble": "assemble: ",
	"migrate":  "migrate: ",
}

// ownHeadOf is the head a diagnosis opens with when that head is the refused command's own role,
// path or name; "" when the diagnosis starts at what is wrong.
func ownHeadOf(leaf *cobra.Command, diagnosis string) string {
	path := ""
	if leaf != leaf.Root() {
		path = strings.TrimPrefix(leaf.CommandPath(), leaf.Root().Name()+" ")
	}
	if allowed, ok := layerWords[path]; ok && strings.HasPrefix(diagnosis, allowed) {
		return ""
	}
	var names []string
	if path != "" {
		names = append(names, path, leaf.Name())
	}
	var heads []string
	for _, role := range roleWords() {
		heads = append(heads, role+": ")
		for _, n := range names {
			heads = append(heads, role+" "+n+" ", role+" "+n+":")
		}
	}
	for _, n := range names {
		heads = append(heads, n+": ", n+" ")
	}
	for _, h := range heads {
		if strings.HasPrefix(diagnosis, h) {
			return h
		}
	}
	return ""
}

// roleWords are the role names the trees are built for, read from the trees.
func roleWords() []string {
	var out []string
	for role := range AllRoots() {
		out = append(out, role)
	}
	sort.Strings(out)
	return out
}

// The gate reads a doubled head as one, and a diagnosis that starts at what is wrong as none.
func TestTheOnceGateTellsADoubledHeadFromADiagnosis(t *testing.T) {
	root := NewRootFor("red-lens-evidence")
	leaf, _, err := root.Find([]string{"finding"})
	if err != nil {
		t.Fatal(err)
	}
	file, _, err := root.Find([]string{"motion", "docket", "file"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		leaf      *cobra.Command
		diagnosis string
		doubled   bool
	}{
		{leaf, "requires --reason: the explanation red re-audits the repair against", false},
		{leaf, "lens finding requires --reason: the explanation", true},
		{leaf, "lens: requires --reason", true},
		{leaf, "finding: --quote was not found", true},
		{leaf, "finding requires --reason", true},
		{leaf, "findings are raised against the report as it stands", false},
		{leaf, "record: finding requires --reason", false}, // the record layer's word, not the command's
		{file, "motion docket file: no gap G-1", true},
		{file, "lens motion docket file requires --reason", true},
		{file, "a motion's filer does not rule it", false},
		{file, "file: no gap", true},
		{file, "a docket motion names the gap the bench is to decide", false},
		{root, "no command named \"x\" exists", false},
	} {
		if got := ownHeadOf(c.leaf, c.diagnosis) != ""; got != c.doubled {
			t.Errorf("`%s` refused with %q: doubled = %v, want %v", c.leaf.CommandPath(), c.diagnosis, got, c.doubled)
		}
	}
}
