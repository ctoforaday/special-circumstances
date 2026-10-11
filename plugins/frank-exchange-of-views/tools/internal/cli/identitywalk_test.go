package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// EVERY COMMAND ON EVERY SEAT'S SURFACE REFUSES BOTH IDENTITY DISAGREEMENTS — the ones that record
// and the ones that only read.
//
// Two things a call can say about who is asking that the record or the dispatch contradicts: a
// --seat-id that is not the seat this agent registered as, and a --run that is not the run the
// seat was dispatched into. A recording verb that ignores either writes under the wrong name or
// into the wrong run. A READ that ignores either does something quieter: it answers, from another
// seat's surface or another run's record, in bytes that look exactly like the right answer.
//
// THE TREES ARE WALKED, so the table is the surface itself: a command added tomorrow is held to
// both refusals without anyone listing it. Each command's RunE is called directly, with only the
// run and the seat id parsed onto it — cobra's own required-flag check would otherwise answer
// first, for a reason that has nothing to do with identity.
//
// Neither refusal may write: the call is refused, and the record is as it was.
func TestEveryLeafRefusesBothIdentityDisagreements(t *testing.T) {
	runDir := seatRun(t)
	// A REAL directory, so the refusal is about contradicting the dispatch rather than about a path
	// that happens not to exist.
	elsewhereRun := filepath.Join(t.TempDir(), "another-run")
	if err := os.MkdirAll(elsewhereRun, 0o755); err != nil {
		t.Fatal(err)
	}

	roles := make([]string, 0, len(AllRoots()))
	for role := range AllRoots() {
		roles = append(roles, role)
	}
	sort.Strings(roles)

	walked, reads := 0, 0
	for _, role := range roles {
		surface := seat.DispatchedAs(AllRoots()[role])
		// Neither owes an --occasion: the bench is the one seat whose register carries one.
		bound := "red-lens-evidence"
		if surface == bound {
			bound = "red-chair"
		}
		boundAgent := "agent_walking_" + role + "_bound_as_" + bound
		t.Setenv(seatenv.AgentVar, boundAgent)
		if _, err := run(t, "register", "--run", runDir, "--seat-id", bound); err != nil {
			t.Fatalf("register %s: %v", bound, err)
		}
		t.Setenv(seatenv.AgentVar, "")

		for _, path := range runnableCommands(t, surface, runDir) {
			walked++
			leaf, _, err := NewRootFor(surface).Find(path)
			if err != nil {
				t.Fatal(err)
			}
			if seat.RecordType(leaf) == "" {
				reads++
			}
			name := role + "/" + strings.Join(path, " ")

			t.Run(name+"/seat id disagrees with the registration", func(t *testing.T) {
				t.Setenv(seatenv.AgentVar, boundAgent)
				before := len(events(t, runDir))
				out, err := callLeaf(t, surface, path, runDir)
				if after := len(events(t, runDir)); after != before {
					t.Errorf("the refused call appended %d event(s)", after-before)
				}
				if err == nil {
					t.Fatalf("ran as %s for an agent the record binds to %s, and answered:\n%s", surface, bound, clipTo(out, 300))
				}
				if got := feov.CodeOf(err); got != string(feov.Conflict) || !strings.Contains(err.Error(), "disagrees with the seat this agent registered as") {
					t.Errorf("refused for another reason (code %q): %v", got, err)
				}
			})

			t.Run(name+"/run disagrees with the dispatch", func(t *testing.T) {
				t.Setenv(seatenv.Var, "")
				t.Setenv(seatenv.VarWrapper, runDir)
				before := len(events(t, runDir))
				out, err := callLeaf(t, surface, path, elsewhereRun)
				if after := len(events(t, runDir)); after != before {
					t.Errorf("the refused call appended %d event(s) to the dispatched run", after-before)
				}
				if entries, _ := os.ReadDir(elsewhereRun); len(entries) != 0 {
					t.Errorf("the refused call wrote into the run it named: %v", entries)
				}
				if err == nil {
					t.Fatalf("ran against a --run the dispatch contradicts, and answered:\n%s", clipTo(out, 300))
				}
				if !strings.Contains(err.Error(), "disagrees with the run this seat was dispatched into") {
					t.Errorf("refused for another reason: %v", err)
				}
				if strings.TrimSpace(out) != "" {
					t.Errorf("a refused call printed from a run:\n%s", clipTo(out, 300))
				}
			})
		}
	}
	// THE FLOOR IS WHAT THE WALK FOUND when it was written, less a margin for verbs that retire: a
	// walk that silently stops reaching a tree, or the reads, measures a fraction and reports green.
	if walked < 90 || reads < 40 {
		t.Fatalf("walked %d command(s), %d of them reads — the walk is not reaching the surfaces", walked, reads)
	}
}

// runnableCommands is every command on the surface that runs something of its own: each leaf, and
// a group that answers when called bare. Cobra's own scaffolding is left out, and so is a group
// whose bare call reads nothing (bareCallReadsNothing).
func runnableCommands(t *testing.T, surface, runDir string) [][]string {
	t.Helper()
	var out [][]string
	var walk func(c *cobra.Command, path []string)
	walk = func(c *cobra.Command, path []string) {
		for _, sub := range c.Commands() {
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			p := append(append([]string{}, path...), sub.Name())
			if sub.Runnable() && !(sub.HasSubCommands() && bareCallReadsNothing(t, surface, p, runDir)) {
				out = append(out, p)
			}
			walk(sub, p)
		}
	}
	walk(NewRootFor(surface), nil)
	return out
}

// bareCallReadsNothing is a group whose bare call is its own help page, or a refusal that it was
// given no verb: it answers the same with a run and a registered seat as it does with neither, so
// it reads no record and there is no identity for it to get wrong. A group that reads — bare
// `show` is the seat's work list — answers differently, and is walked.
func bareCallReadsNothing(t *testing.T, surface string, path []string, runDir string) bool {
	t.Helper()
	root := NewRootFor(surface)
	var blind bytes.Buffer
	root.SetOut(&blind)
	root.SetErr(&blind)
	group, _, err := root.Find(path)
	if err != nil || group.RunE == nil {
		return false
	}
	blindErr := group.RunE(group, nil)
	seen, seenErr := callLeaf(t, surface, path, runDir)
	text := func(e error) string {
		if e == nil {
			return ""
		}
		return e.Error()
	}
	return blind.String() == seen && text(blindErr) == text(seenErr)
}

// callLeaf builds the surface's tree, parses the run and the seat id onto the command at path, and
// calls its RunE — the command's own preamble, with nothing of cobra's in front of it.
func callLeaf(t *testing.T, surface string, path []string, runDir string) (string, error) {
	t.Helper()
	root := NewRootFor(surface)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	leaf, rest, err := root.Find(path)
	if err != nil || len(rest) != 0 {
		t.Fatalf("find %v: %v (rest %v)", path, err, rest)
	}
	if err := leaf.ParseFlags([]string{"--run", runDir, "--seat-id", surface}); err != nil {
		t.Fatalf("parse %v: %v", path, err)
	}
	if leaf.RunE == nil {
		t.Fatalf("%v runs through Run, which cannot refuse", path)
	}
	runErr := leaf.RunE(leaf, nil)
	return out.String(), runErr
}
