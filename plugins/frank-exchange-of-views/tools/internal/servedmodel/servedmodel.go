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
// # Finding the trajectory without composing a path out of parts
//
// The file is `agent-<agent_id>.jsonl`, and `agent_id` is the harness's own handle for the
// subagent — injected by the PreToolUse hook as FEOV_AGENT_ID, never typed by a seat, and the ONE
// payload field that discriminates one seat from another. It is unique BY CONSTRUCTION, so a search
// for that exact name is not a pattern standing in for a schema: it either resolves to the one
// file the harness wrote for this agent, or to nothing. The id never enters a pattern — it is
// compared with each file's name — so no id, whatever it carries, widens the search to another
// seat's file.
//
// NOTHING IS A LOUD ANSWER, not a quiet zero. Locate distinguishes "no such trajectory" from "more
// than one" from "the search root does not exist", because a run whose measurement silently
// returned "" would report the configured tier as served and rebuild the exact defect above.
package servedmodel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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

// searchRoots are the directories a trajectory can live under. CLAUDE_CONFIG_DIR wins where it is
// set, because a client configured elsewhere would otherwise be searched at a path that does not
// exist — which returns nothing, and nothing must not read as "no substitution".
func searchRoots() []string {
	if d := strings.TrimSpace(os.Getenv("CLAUDE_CONFIG_DIR")); d != "" {
		return []string{d}
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	return []string{filepath.Join(home, ".claude")}
}

// Locate resolves an agent id to the one trajectory the harness wrote for it.
//
// Every failure is named. `""` with a nil error is not among the answers.
func Locate(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("servedmodel: no agent id — this process was not rewritten by the PreToolUse hook, so the seat's trajectory cannot be named and the serving model is NOT MEASURED (it is not the configured one by default)")
	}
	roots := searchRoots()
	if len(roots) == 0 {
		return "", fmt.Errorf("servedmodel: no home directory and no CLAUDE_CONFIG_DIR, so there is nowhere to look for agent-%s.jsonl", id)
	}
	name := "agent-" + id + ".jsonl"
	var hits []string
	for _, root := range roots {
		// ONE RULE: the transcript is the file named agent-<id>.jsonl under a session's
		// `projects/<project>/<session>/subagents/` directory, at any depth. The two wildcards are
		// the project slug and the session — neither of which a seat knows — and the depth is the
		// dispatcher's: a plain subagent sits in `subagents/` itself, a Workflow seat in
		// `subagents/workflows/wf_<id>/`. The id is the discriminator and it is exact.
		dirs, err := filepath.Glob(filepath.Join(root, "projects", "*", "*", "subagents"))
		if err != nil {
			return "", fmt.Errorf("servedmodel: searching for %s: %w", name, err)
		}
		for _, dir := range dirs {
			// An unreadable subtree is skipped rather than fatal: what it hides is reported as the
			// miss below, never as a match.
			_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if !d.IsDir() && d.Name() == name {
					hits = append(hits, path)
				}
				return nil
			})
		}
	}
	hits = dedup(hits)
	switch len(hits) {
	case 1:
		return hits[0], nil
	case 0:
		return "", fmt.Errorf("servedmodel: no trajectory named agent-%s.jsonl under any session's subagents directory in %s — the serving model is NOT MEASURED for this seat, which is a different answer from \"it matched what was configured\"", id, strings.Join(roots, ", "))
	default:
		return "", fmt.Errorf("servedmodel: %d files named agent-%s.jsonl (%s) — the agent id is supposed to be unique, so this is a search that has stopped discriminating and its answer is refused rather than picked from", len(hits), id, strings.Join(hits, ", "))
	}
}

func dedup(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// Observe is Locate followed by Read: what the harness served this agent, or a stated reason
// nobody knows.
func Observe(id string) (Observation, error) {
	path, err := Locate(id)
	if err != nil {
		return Observation{}, err
	}
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
