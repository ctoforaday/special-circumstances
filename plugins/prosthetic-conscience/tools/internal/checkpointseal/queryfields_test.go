package checkpointseal

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// THE RECORD'S ONLY READERS ARE THE EMBEDDED jq PROGRAMS, so this checks they still resolve.
//
// seals.jsonl is written by three binaries and read by nothing that compiles. `queries/*.jq` are
// what read it, and criteria 4 and 6 are DECIDED by them. Rename a JSON tag on sealRow and
// `select(.handles_measured)` matches nothing, `map(.emission_bytes_max) | max` is null, and the
// gate reports zero emissions — the PASSING answer. The broken query and the clean board produce
// the same output ([[facts-are-fields]] clause 3: ask what a no-match returns).
//
// THIS USED TO PARSE A PLAN. The queries lived in §V of plans/checkpoint-freshness.md and this
// test reached five directories up to extract field names out of that markdown — a design
// document made load-bearing for a build, where editing prose could fail CI and moving the file
// would fail it for a reason no reader would guess. The queries are files now; this reads the
// embedded copy, which is the same one that would be run.
//
// Generation is still not available: the programs carry LOGIC, not just field names, and
// generating them would mean generating the analysis. So this stays a guard, and per the same
// rule it says why. What keeps it from reproducing the defect one level up is that its list is
// not hand-kept — it is extracted from the queries themselves on every run, so a field a new
// query reads is checked without anyone remembering to add it here.
func TestEveryFieldTheQueriesReadExistsOnTheRecord(t *testing.T) {
	queries := Queries()
	if len(queries) == 0 {
		t.Fatal("no embedded queries; this test would pass vacuously, which is the exact " +
			"failure it exists to catch")
	}

	read := fieldsReadByJQ(queries)
	if len(read) == 0 {
		t.Fatalf("extracted no field names from %d embedded queries; vacuous pass", len(queries))
	}
	t.Logf("%d queries read %d distinct fields: %s", len(queries), len(read), strings.Join(read, " "))

	have := tagsOf(t)
	for _, f := range read {
		if !have[f] {
			t.Errorf("queries/*.jq read .%s, and no sealRow field carries that JSON tag — "+
				"that query returns null or zero rows, which is indistinguishable from an honest "+
				"empty corpus", f)
		}
	}
}

// jqField matches a field reference in a jq program: `.name` or `.name` inside a pipeline,
// but not the `.[]` iterator and not a decimal in a number.
var jqField = regexp.MustCompile(`\.([a-z][a-z0-9_]*)\b`)

func fieldsReadByJQ(queries map[string]string) []string {
	seen := map[string]bool{}
	for _, body := range queries {
		// Comments are PROSE, and prose has full stops. "…, and no render may exceed 200
		// bytes." yielded a field called `and` on the first run of this test.
		body = stripShellComments(body)
		for _, m := range jqField.FindAllStringSubmatch(body, -1) {
			seen[m[1]] = true
		}
	}
	// jq builtins and path components reached through a dot are not record fields.
	for _, notAField := range []string{"jsonl", "claude", "checkpoints", "md", "length", "max", "min"} {
		delete(seen, notAField)
	}
	out := make([]string, 0, len(seen))
	for f := range seen {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// tagsOf asks the ENCODER, not the source text. A tag read out of the struct declaration by
// regex would be one more fact recovered from a string; marshalling a row and reading its keys
// is what the queries will actually meet.
func tagsOf(t *testing.T) map[string]bool {
	t.Helper()
	// Every optional field populated: omitempty means a zero row would not show them, and the
	// absent field is precisely what this test must not mistake for a present one.
	n := 1
	row := sealRow{
		At: "t", Event: "PreCompact", Occasion: "compact", SessionID: "s", AgentID: "a",
		SealTrigger: "auto", BodySHA: "sha", WrittenAt: "t",
		LiveHandles: &n, NoteAgeTurns: &n, NoteGrowthTokens: &n, NoteBranchCommits: &n,
		GrowthSince: "t", NudgeAnswered: "yes",
		EmissionsThisSession: &n, EmissionBytesMax: &n,
	}
	b, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for k := range m {
		have[k] = true
	}
	return have
}

// stripShellComments removes `#` to end-of-line. The queries live in bash blocks whose
// comments state what each gate decides, and those sentences end in full stops.
func stripShellComments(body string) string {
	var kept []string
	for _, line := range strings.Split(body, "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
