package checkpointseal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// THE RECORD HAS NO READER IN THIS REPOSITORY. seals.jsonl is written by three binaries and
// read by nothing that compiles — its readers are the `jq` queries in §V of
// plans/checkpoint-freshness.md, run by a human, naming fields in prose.
//
// That is the whole failure mode this test exists for, and it is live rather than theoretical.
// Criterion 4 and criterion 6 are DECIDED by those queries. Rename a JSON tag here and
// `select(.handles_measured)` matches nothing, `map(.emission_bytes_max) | max` is null, and the
// gate reports zero emissions — which is the PASSING answer. The broken query and the clean
// board produce the same output ([[facts-are-fields]] clause 3: ask what a no-match returns).
//
// Generation is not available here: the queries carry LOGIC, not just field names, and
// generating them would mean generating the plan. So this is a guard, and per the same rule it
// says why. What keeps it from reproducing the defect one level up is that its list is not
// hand-kept — it is extracted from the plan on every run, so a query added to §V is checked
// without anyone remembering to add it here.
func TestEveryFieldThePlansQueriesReadExistsOnTheRecord(t *testing.T) {
	plan := filepath.Join("..", "..", "..", "..", "..", "plans", "checkpoint-freshness.md")
	b, err := os.ReadFile(plan)
	if err != nil {
		t.Fatalf("cannot read the plan whose queries are this record's only readers: %v", err)
	}

	read := fieldsReadByJQ(string(b))
	if len(read) == 0 {
		t.Fatal("extracted no field names from the plan's jq blocks; this test would pass " +
			"vacuously, which is the exact failure it exists to catch")
	}
	t.Logf("%d distinct fields read by the plan's queries: %s", len(read), strings.Join(read, " "))

	have := tagsOf(t)
	for _, f := range read {
		if !have[f] {
			t.Errorf("the plan's §V queries read .%s, and no sealRow field carries that JSON tag — "+
				"that query returns null or zero rows, which is indistinguishable from an honest "+
				"empty corpus", f)
		}
	}
}

// jqField matches a field reference in a jq program: `.name` or `.name` inside a pipeline,
// but not the `.[]` iterator and not a decimal in a number.
var jqField = regexp.MustCompile(`\.([a-z][a-z0-9_]*)\b`)

// fenced captures fenced code blocks, which is where the plan keeps its queries.
var fenced = regexp.MustCompile("(?s)```[a-z]*\n(.*?)```")

func fieldsReadByJQ(doc string) []string {
	seen := map[string]bool{}
	for _, block := range fenced.FindAllStringSubmatch(doc, -1) {
		body := block[1]
		if !strings.Contains(body, "jq ") {
			continue
		}
		// Only the seals.jsonl queries: the plan also shows jq over live TRANSCRIPTS, whose
		// fields belong to the client's schema and are not this record's to answer for.
		if !strings.Contains(body, "seals.jsonl") {
			continue
		}
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
