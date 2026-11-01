package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// `assemble --json` ANSWERS IN THE ENVELOPE ITS REFUSAL DOES. On #1677's smoke it printed its prose
// line on success, the seat's jq failed on it, and the seat assembled the run a second time for
// nothing.
func TestAssembleJSONIsTheEnvelope(t *testing.T) {
	runDir := seatRun(t)
	assembleAndStamp(t, runDir, "agent_bench_assemble")
	out, err := run(t, "--json", "assemble", "--run", runDir, "--seat-id", "judge")
	if err != nil {
		t.Fatalf("assemble --json: %v\n%s", err, out)
	}
	var env struct {
		Verb   string `json:"verb"`
		OK     bool   `json:"ok"`
		Result struct {
			Documents int    `json:"documents"`
			Verdict   string `json:"verdict"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &env); err != nil {
		t.Fatalf("assemble --json printed no JSON (%v):\n%s", err, out)
	}
	if env.Verb != "assemble" || !env.OK || env.Result.Documents == 0 || env.Result.Verdict != "UNVERIFIED" {
		t.Errorf("the envelope does not state the assembly: %+v\n%s", env, out)
	}
	bare, err := run(t, "assemble", "--run", runDir, "--seat-id", "judge")
	if err != nil {
		t.Fatalf("assemble: %v\n%s", err, bare)
	}
	if strings.HasPrefix(strings.TrimSpace(bare), "{") || !strings.Contains(bare, "the verdict stamped UNVERIFIED") {
		t.Errorf("the bare call is the human form, and it printed:\n%s", bare)
	}
}
