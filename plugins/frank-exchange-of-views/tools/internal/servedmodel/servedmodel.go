// Package servedmodel answers a question the engine cannot ask and the seat must not be trusted
// to answer: WHICH MODEL ACTUALLY REPLIED.
//
// # The measured failure
//
// Both 2026-08-23 research runs configured `claude-fable-5` for the bulk tier. Every one of the
// 44 bulk seats was served `claude-opus-4-8` instead, on every turn, for ~$379 of spend. Nothing
// in the loop knew. `inputs/run-config.json` records what was REQUESTED, the seats and the
// assembler read only that, and so run B's certified report states its own methodology wrong —
// "the blue/red pairing this run used is fable and sonnet" — a false claim in a signed artifact.
// The one component that noticed was the post-hoc capture audit, hours later, and it graded the
// substitution WARN rather than FAIL because the served tier was CHEAPER than the configured one.
//
// # Why this is a measurement and not an inference
//
// The harness declares the substitution. A seat's own trajectory opens with an assistant line
// whose content is a single `fallback` block naming both ends as fields:
//
//	{"type":"fallback","from":{"model":"claude-fable-5"},"to":{"model":"claude-opus-4-8"}}
//
// and every assistant line thereafter carries `message.model` — the model that answered, not the
// one that was asked for. So the request and the service are two fields on a record somebody else
// wrote, which is the shape this repository asks for everywhere: not a tier guessed from a price,
// not a model a seat reports about itself.
//
// # Where the trajectory is
//
// The record says. The SubagentStop hook is told each seat's transcript path and stores it on the
// agent's sitting_close; record.TranscriptOf reads it back, and this package opens the path it is
// handed. It searches nothing: a search over the harness's layouts missed every Workflow seat for
// two runs (#1438), and its miss read like a seat that never ran.
//
// NOTHING IS A LOUD ANSWER, not a quiet zero. A path that cannot be opened, and a trajectory that
// names no model yet, are each an error that says so, because a run whose measurement silently
// returned "" would report the configured tier as served and rebuild the exact defect above.
package servedmodel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Observation is what a seat's own trajectory says about the model that answered it.
//
// Requested is set ONLY when the harness declared a substitution. Its absence therefore means
// "nothing was substituted, as far as anything said", and never "the request equalled the
// service" — those are the same outcome but not the same evidence, and Declared keeps them apart.
type Observation struct {
	Served    string `json:"served"`
	Requested string `json:"requested,omitempty"`
	// Declared is true when a `fallback` block named both ends. A substitution inferred by
	// comparing Served to run-config is a different, weaker fact.
	Declared bool   `json:"declared"`
	Path     string `json:"path"`
}

// Substituted reports whether the harness said it served something other than what was asked.
func (o Observation) Substituted() bool {
	return o.Declared && o.Requested != "" && o.Requested != o.Served
}

// Read scans one trajectory and returns what it says about the serving model.
//
// The FIRST model-bearing assistant line wins, because that is the line that exists by the time a
// seat makes its first tool call — measured: the opening three lines of a real seat's trajectory
// are the fallback notice, a thinking block and the tool_use, all before any tool_result. A seat
// can therefore measure itself at `register` rather than at capture.
//
// A malformed line is SKIPPED, never fatal: a trajectory is appended to live and its final line
// may be half-written while this reads.
func Read(r io.Reader) (Observation, bool) {
	var o Observation
	sc := bufio.NewScanner(r)
	// Trajectory lines carry whole tool results and run far past the 64KiB default.
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var j struct {
			Message struct {
				Model   string `json:"model"`
				Content []struct {
					Type string `json:"type"`
					From struct {
						Model string `json:"model"`
					} `json:"from"`
					To struct {
						Model string `json:"model"`
					} `json:"to"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &j) != nil {
			continue
		}
		// The DECLARATION outranks the observation, and is looked for on every line rather than
		// only the first: a fallback can be announced mid-sitting, and that is precisely the case
		// the issue calls out — "a substitution mid-run should not silently change the adversary's
		// strength".
		for _, c := range j.Message.Content {
			if c.Type == "fallback" && c.From.Model != "" && c.To.Model != "" {
				o.Requested, o.Served, o.Declared = c.From.Model, c.To.Model, true
				return o, true
			}
		}
		if o.Served == "" && j.Message.Model != "" {
			o.Served = j.Message.Model
		}
	}
	// A scan error over a partially-written file still yields what was read before it.
	return o, o.Served != ""
}

// Observe reads the trajectory at path, which the record names (record.TranscriptOf): what the
// harness served its agent, or a stated reason nobody knows.
func Observe(path string) (Observation, error) {
	f, err := os.Open(path)
	if err != nil {
		return Observation{}, fmt.Errorf("servedmodel: opening %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	o, ok := Read(f)
	if !ok {
		return Observation{}, fmt.Errorf("servedmodel: %s carries no assistant turn naming a model yet — NOT MEASURED, not unsubstituted", path)
	}
	o.Path = path
	return o, nil
}
