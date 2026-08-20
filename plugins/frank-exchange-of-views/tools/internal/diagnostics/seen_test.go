package diagnostics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A trajectory line carrying one tool_result whose text is `body`.
func resultLine(t *testing.T, body string) string {
	t.Helper()
	ev := map[string]any{"message": map[string]any{"content": []any{
		map[string]any{"type": "tool_result", "content": body},
	}}}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func trajectory(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "trajectory.jsonl")
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// lensActs is the role's act list, as cli.CommandPaths() renders it with the role stripped.
func lensActs() []string {
	return []string{"finding", "verify", "corroborate", "reproduce", "friction", "register", "show"}
}

// THE PADDED NAME IS NOT DROPPED. Cobra pads the name column to the width of the LONGEST name,
// leaving exactly that one name followed by a SINGLE space. A `\s{2,}` separator therefore loses
// precisely one verb per role — silently, from numerator and denominator alike, so the ratio
// still reads like a measurement. `corroborate` is the longest here and so the one at risk.
func TestTheLongestNameIsNotDroppedByTheColumnPadding(t *testing.T) {
	help := "Available Commands:\n" +
		"  corroborate a claim from a source YOU found\n" +
		"  finding     a lens finding\n" +
		"  verify      adjudicate ONE citation\n" +
		"\n"
	got, err := ReadSeen(trajectory(t, resultLine(t, help)), lensActs())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Leaf["corroborate"] {
		t.Error("the longest name was dropped — cobra pads to it, so it is the one a two-space separator loses")
	}
	if len(got.Leaf) != 3 {
		t.Errorf("saw %d verbs, want 3: %v", len(got.Leaf), got.Leaf)
	}
}

// A LISTING THAT IS NOT THIS ROLE'S CONTRIBUTES NOTHING, and is counted as REJECTED rather than
// ignored. The probe stages ROOT help into a board as setup material, and root lists `friction`
// and `count-claims` — names that are also role verbs. Counting line by line credits every seat
// with those before it acts, and flatters exactly the arms that open help LEAST.
func TestARootListingDoesNotCreditTheRole(t *testing.T) {
	root := "Available Commands:\n" +
		"  friction    the operator's read of the friction channel\n" +
		"  setup       build a research run's blackboard\n" +
		"  dashboard   the live run dashboard\n" +
		"\n"
	got, err := ReadSeen(trajectory(t, resultLine(t, root)), lensActs())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Leaf) != 0 {
		t.Errorf("a root listing credited the lens with %v — `friction` is in both, which is exactly the trap", got.Leaf)
	}
	if got.Rejected != 1 {
		t.Errorf("rejected = %d, want 1: a listing read and discarded is a different state from no listing at all, and reporting both as zero is the plausible zero this measures", got.Rejected)
	}
}

// TOP SATURATES AND LEAF DOES NOT, which is why both are reported. One root help reveals every
// direct child, so Top is all-or-nothing per sitting; a subgroup's children need the subgroup
// opened, so Leaf is the finer question.
func TestTopAndLeafAnswerDifferentQuestions(t *testing.T) {
	acts := []string{"line-of-inquiry propose", "line-of-inquiry move", "edit"}
	top := "Available Commands:\n" +
		"  edit            change the report\n" +
		"  line-of-inquiry propose and move a direction\n" +
		"\n"
	got, err := ReadSeen(trajectory(t, resultLine(t, top)), acts)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Top) != 2 {
		t.Errorf("Top = %v, want the two direct children", got.Top)
	}
	if got.Leaf["propose"] || got.Leaf["move"] {
		t.Error("a subgroup's CHILDREN were credited from the parent listing; opening the group is the act that reveals them")
	}
}

// NO HELP AT ALL IS ZERO WITH ZERO BLOCKS — distinguishable from a run whose listings were all
// rejected, which is the whole reason Blocks and Rejected are both carried.
func TestASeatThatNeverOpenedHelpIsDistinguishableFromOneWhoseListingsWereRejected(t *testing.T) {
	got, err := ReadSeen(trajectory(t, resultLine(t, "ok, minted R1-1")), lensActs())
	if err != nil {
		t.Fatal(err)
	}
	if got.Blocks != 0 || got.Rejected != 0 || len(got.Leaf) != 0 {
		t.Errorf("a seat that never opened help reported %+v", got)
	}
}

// COBRA'S HOUSEKEEPING NEITHER CREDITS NOR DISQUALIFIES. `completion` and `help` are generated
// onto every parent command, so they appear inside a seat's OWN listing — and an all-names test
// that did not know them rejected the whole block, reporting a seat that had just been handed its
// entire surface as having seen nothing.
//
// MEASURED, on the first real trajectory this was run against: the lens received its own listing
// through a refusal, and the block was discarded over the single word `completion`.
func TestCobrasHousekeepingDoesNotDisqualifyTheRolesOwnListing(t *testing.T) {
	help := "Available Commands:\n" +
		"  completion  Generate the autocompletion script for the specified shell\n" +
		"  corroborate a claim from a source YOU found\n" +
		"  finding     a lens finding\n" +
		"  help        Help about any command\n" +
		"  verify      adjudicate ONE citation\n" +
		"\n"
	got, err := ReadSeen(trajectory(t, resultLine(t, help)), lensActs())
	if err != nil {
		t.Fatal(err)
	}
	if got.Blocks != 1 {
		t.Fatalf("the role's own listing was rejected over cobra's generated entries: %+v", got)
	}
	if len(got.Leaf) != 3 {
		t.Errorf("saw %v, want the three role verbs — housekeeping must not be counted either", got.Leaf)
	}
}
