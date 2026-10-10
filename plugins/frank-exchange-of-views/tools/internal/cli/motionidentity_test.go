package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// A MOTION IS FILED UNDER THE SEAT THE AGENT REGISTERED AS, OR NOT AT ALL (#1204).
//
// The filer id is what found_by, estoppel and the bench read, so a motion recorded under another
// seat's id is an argument credited to a party that never made it. Measured on a scratch run before
// the fix: an agent registered as red-lens-evidence ran `motion petition file --seat-id
// red-lens-logic` and the record carried `"filer": "red-lens-logic"`, while the same agent's `log`
// under that id was refused.
func TestAMotionIsRefusedUnderASeatTheAgentDidNotRegisterAs(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_registered_as_evidence")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-lens-logic",
		"--class", "safety", "--relief", "r", "--reason", "x")
	if err == nil {
		t.Fatal("a motion was filed under red-lens-logic by an agent registered as red-lens-evidence")
	}
	if got := feov.CodeOf(err); got != string(feov.Conflict) {
		t.Errorf("refused with code %q, want %q (the identity disagreement): %v", got, feov.Conflict, err)
	}
	if n := motionsOnRecord(t, runDir); n != 0 {
		t.Errorf("%d motion(s) on the record after the refusal", n)
	}
}

// AN AGENT THAT NEVER REGISTERED FILES NOTHING BEFORE A REGISTER BINDS IT. Where the surface's
// register is the binding and nothing more, the tool registers it first, silently — which is what
// every other verb does, and what arms the sitting's call cap. Where the register carries something
// only the seat knows (the bench's --occasion), the filing is refused.
func TestAnUnregisteredAgentFilesNoMotionWithoutARegister(t *testing.T) {
	runDir := seatRun(t)

	t.Setenv(seatenv.AgentVar, "agent_lens_never_registered")
	if _, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--class", "safety", "--relief", "r", "--reason", "x"); err != nil {
		t.Fatalf("a lens's first act was refused rather than registered for it: %v", err)
	}
	seatID, found, err := record.RegisteredSeatOfAgent(runtest.Open(t, runDir), "agent_lens_never_registered")
	if err != nil {
		t.Fatal(err)
	}
	if !found || seatID != "red-lens-evidence" {
		t.Errorf("the motion landed and no register binds its agent (bound %q, found %v)", seatID, found)
	}

	t.Setenv(seatenv.AgentVar, "agent_bench_never_registered")
	before := motionsOnRecord(t, runDir)
	if _, err := run(t, "motion", "docket", "file", "--run", runDir, "--seat-id", "judge",
		"--id", "G1", "--reason", "x"); err == nil || !strings.Contains(err.Error(), "register") {
		t.Errorf("an unregistered agent's filing on the bench's surface, whose register the tool cannot perform for it, = %v; want it refused naming the register", err)
	}
	if after := motionsOnRecord(t, runDir); after != before {
		t.Errorf("the refused filing recorded a motion (%d -> %d)", before, after)
	}
}

// A MOTION'S REFUSAL HAS EVERY OTHER VERB'S SHAPE. Its handler returned the error raw, so a --json
// caller got the top-level parse envelope — no verb, no role — for a refusal the handler made.
func TestAMotionRefusalIsTheVerbEnvelope(t *testing.T) {
	runDir := seatRun(t)
	// A HANDLER'S refusal, past parsing: no gap G99 exists to put before the bench.
	// --reason is passed so the refusal reached is the record's, and so the harness's unread-reason
	// check cannot stand in for it.
	out, err := runAt(t, "motion", "docket", "file", "--json", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--id", "G99", "--reason", "the bench should decide the gap")
	if err == nil {
		t.Fatalf("a --json refusal returned no error, so the call exits 0 and the refusal log never sees it:\n%s", out)
	}
	var env struct {
		Verb string  `json:"verb"`
		Role *string `json:"role"`
		OK   bool    `json:"ok"`
		Code string  `json:"code"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &env); err != nil {
		t.Fatalf("not one JSON envelope: %v\n%s", err, out)
	}
	if env.OK || env.Code == "" {
		t.Fatalf("not a refusal: %s", out)
	}
	if env.Verb != "file" || env.Role == nil || *env.Role == "" {
		t.Errorf("a motion refusal is missing the verb envelope's verb/role: %s", out)
	}
}

// EVERY WRITING VERB ON EVERY SEAT SURFACE REFUSES WHAT seat.Begin REFUSES.
//
// The motion verbs re-implemented two of Begin's checks by hand, each patched in after a measured
// hole, and lacked the rest — the identity disagreement and register-first. This walks the real
// trees, so a verb added later that reads its context without going through Begin fails here
// rather than in a run.
//
// The reference is Begin itself, called on the same command under the same conditions: where Begin
// refuses, the verb must refuse with the same coded error; where Begin registers the agent for it,
// the agent must be bound after the verb ran. Calling RunE directly skips cobra's required-flag
// check on purpose, so the only thing standing between the call and the handler is the verb's own
// preamble.
func TestEveryWritingVerbRefusesWhatBeginRefuses(t *testing.T) {
	runDir := seatRun(t)
	r := runtest.Open(t, runDir)
	n := 0
	for role, root := range AllRoots() {
		surface := seat.DispatchedAs(root)
		// Neither owes an --occasion: the bench is the one seat whose register carries one.
		elsewhere := "red-lens-evidence"
		if surface == elsewhere {
			elsewhere = "red-chair"
		}
		boundAgent := "agent_bound_as_" + elsewhere + "_for_" + role
		t.Setenv(seatenv.AgentVar, boundAgent)
		if _, err := run(t, "register", "--run", runDir, "--seat-id", elsewhere); err != nil {
			t.Fatalf("register %s: %v", elsewhere, err)
		}

		for _, path := range writingLeaves(root) {
			n++
			name := role + "/" + strings.Join(path, " ")
			t.Run(name+"/seat-id disagrees with the binding", func(t *testing.T) {
				t.Setenv(seatenv.AgentVar, boundAgent)
				want := beginOn(t, surface, path, runDir)
				if want == nil {
					t.Fatalf("Begin accepted --seat-id %s from an agent bound as %s — the reference is wrong", surface, elsewhere)
				}
				sameRefusal(t, invokeLeaf(t, surface, path, runDir), want)
			})
			t.Run(name+"/agent never registered", func(t *testing.T) {
				if path[len(path)-1] == "register" {
					// The one verb that may run unbound: it is what creates the binding.
					t.Skip("register creates the binding the other verbs require")
				}
				agent := fmt.Sprintf("agent_unregistered_%d", n)
				t.Setenv(seatenv.AgentVar, agent)
				got := invokeLeaf(t, surface, path, runDir)
				t.Setenv(seatenv.AgentVar, agent+"_reference")
				want := beginOn(t, surface, path, runDir)
				if want != nil {
					sameRefusal(t, got, want)
					return
				}
				if _, found, err := record.RegisteredSeatOfAgent(r, agent); err != nil || !found {
					t.Errorf("Begin registers an unregistered agent first on this surface; the verb ran and no register binds it (err %v)", err)
				}
			})
		}
	}
	if n == 0 {
		t.Fatal("walked no writing verb — the walk measured nothing")
	}
}

// writingLeaves is every command under root that declares the record type it writes — a parent
// with subcommands included, since nothing stops one from writing too.
func writingLeaves(root *cobra.Command) [][]string {
	var out [][]string
	var walk func(c *cobra.Command, path []string)
	walk = func(c *cobra.Command, path []string) {
		for _, sub := range c.Commands() {
			p := append(append([]string{}, path...), sub.Name())
			if seat.RecordType(sub) != "" {
				out = append(out, p)
			}
			walk(sub, p)
		}
	}
	walk(root, nil)
	return out
}

// leafOf builds the surface's real tree and parses only the run and the seat id onto the leaf.
func leafOf(t *testing.T, surface string, path []string, runDir string) *cobra.Command {
	t.Helper()
	root := NewRootFor(surface)
	var sink bytes.Buffer
	root.SetOut(&sink)
	root.SetErr(&sink)
	leaf, rest, err := root.Find(path)
	if err != nil || len(rest) != 0 {
		t.Fatalf("find %v: %v (rest %v)", path, err, rest)
	}
	if err := leaf.ParseFlags([]string{"--run", runDir, "--seat-id", surface}); err != nil {
		t.Fatalf("parse %v: %v", path, err)
	}
	return leaf
}

func invokeLeaf(t *testing.T, surface string, path []string, runDir string) error {
	t.Helper()
	leaf := leafOf(t, surface, path, runDir)
	return leaf.RunE(leaf, nil)
}

func beginOn(t *testing.T, surface string, path []string, runDir string) error {
	t.Helper()
	_, err := seat.Begin(leafOf(t, surface, path, runDir))
	return err
}

func sameRefusal(t *testing.T, got, want error) {
	t.Helper()
	var g, w *feov.Error
	if !errors.As(want, &w) {
		t.Fatalf("Begin's refusal is not coded: %v", want)
	}
	if !errors.As(got, &g) || g.Code != w.Code || g.Error() != w.Error() {
		t.Errorf("the verb does not refuse what Begin refuses.\n got: %v\nwant: %s: %v", got, w.Code, w)
	}
}

func motionsOnRecord(t *testing.T, runDir string) int {
	t.Helper()
	n := 0
	for _, ev := range events(t, runDir) {
		if _, ok := recordpb.BodyAs[*recordpb.Motion](ev); ok {
			n++
		}
	}
	return n
}

// A BINDING LOOKUP THAT CANNOT ANSWER REFUSES A WRITE, NOT A READ.
//
// Only a binding that DISAGREES with --seat-id is a refusal for every verb. A record the lookup
// cannot read — an archived run, a schema this binary does not know — says nothing about who the
// agent is, and a read that does not need the record proceeds as it would with no agent handle.
func TestAnUnreadableBindingRefusesAWriteAndNotARead(t *testing.T) {
	no := false
	dir, sha := cacheScan(t, &no, "application/pdf")
	db := filepath.Join(dir, "records", "record.db")
	if err := os.MkdirAll(filepath.Dir(db), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(db, []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(seatenv.AgentVar, "agent_on_an_unreadable_record")
	if _, _, err := record.RegisteredSeatOfAgent(runtest.Open(t, dir), "agent_on_an_unreadable_record"); err == nil {
		t.Fatal("the binding lookup answered — this test is not exercising a lookup that fails")
	}
	if out, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--dpi", "72", "--run", dir); err != nil {
		t.Fatalf("a read that needs no record was refused because the binding could not be read: %v\n%s", err, out)
	}
	if _, err := run(t, "log", "--run", dir, "--seat-id", "red-lens-evidence", "--type", "defect", "--reason", "x"); err == nil {
		t.Fatal("a write was accepted with no readable binding to attribute it to")
	}
}
