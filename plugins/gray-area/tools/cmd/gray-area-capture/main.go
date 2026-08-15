// gray-area-capture is a gray-area SubagentStop hook (Go binary).
//
// Contract (Design by Contract):
//
//	AFTER a subagent finishes, it MUST record one manifest row naming that seat's
//	own trajectory file, so every downstream consumer is HANDED the path instead
//	of globbing ~/.claude/projects and guessing which file belongs to which seat.
//	During capture it MUST verify the path resolves and record what it found —
//	a path that silently fails to resolve later is the failure mode the miner
//	must never swallow.
//	It MUST NOT block the subagent: every path exits 0.
//
// The manifest is an INDEX, not a copy. It records where a trajectory is and what
// was true of it at capture time; it never duplicates conversation content. The
// transcript already holds the text, and Gray Area's whole posture is that a
// finding cites the primary record rather than a summary of it. Concretely:
// last_assistant_message is present in the hook input and is deliberately NOT
// recorded — copying it here would spread conversation content into a second
// file for no evidentiary gain (see plans/gray-area.md G6).
//
// Measured basis for the fields, from the Phase 0 spike
// (plans/hook-surface-spike.md): SubagentStop carries agent_id, agent_type and
// agent_transcript_path.
//
// SCOPED 2026-08-15 (spike §7a, #189): "and that path resolves to a real per-seat
// file" is true of the seats that carry an agent_type and FALSE of the ones that
// do not. Measured across 69 seat rows in one session: 19 carry a type and all 19
// resolve; 50 carry none and none resolve — zero exceptions, both populations
// interleaved seconds apart, so the correlate is the SEAT, not the session and not
// the environment. The spike saw only the good case because its single subagent
// carried a type.
//
// The path is announced either way, which is what makes this worth writing down:
// agent_transcript_path being present says nothing about the file existing. Every
// row carries a non-empty agent_id in both populations — the typeless events have
// an identity and no type, pointing at a file nobody wrote.
//
// What produces the typeless population is NOT determined. Six probe agents across
// every subagent_type available (and omitted) all carried a type and all resolved.
// Do NOT add a fallback path search: with 50 of 69 seats having no file, that is
// guessing three times in four, and a wrong file confidently attributed to a seat
// is the false citation the adjudicator exists to refuse.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const version = "0.1.0"

// schema is the manifest row version. The transcript format is vendor-internal
// and version-unstable, so every row records what produced it.
const schema = 1

type backgroundTask struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type hookInput struct {
	SessionID           string            `json:"session_id"`
	TranscriptPath      string            `json:"transcript_path"`
	CWD                 string            `json:"cwd"`
	AgentID             string            `json:"agent_id"`
	AgentType           string            `json:"agent_type"`
	AgentTranscriptPath string            `json:"agent_transcript_path"`
	BackgroundTasks     []backgroundTask  `json:"background_tasks"`
	SessionCrons        []json.RawMessage `json:"session_crons"`
	Effort              struct {
		Level string `json:"level"`
	} `json:"effort"`
}

// manifestRow is one seat's entry. Field names are snake_case to match the
// harness payloads a reader will have alongside them.
type manifestRow struct {
	Schema int `json:"schema"`
	// Kind separates the two rows a manifest can hold. "seat" is a subagent's
	// trajectory, handed over at SubagentStop; "session" is the MAIN session's,
	// handed over at SessionStart. They are one file because they belong to one
	// run, and they are distinguished because a consumer asking "where is this
	// session's transcript" must not be answered with a seat's.
	Kind                string   `json:"kind"`
	CapturedAt          string   `json:"captured_at"`
	SessionID           string   `json:"session_id"`
	AgentID             string   `json:"agent_id"`
	AgentType           string   `json:"agent_type"`
	TranscriptPath      string   `json:"transcript_path"`
	AgentTranscriptPath string   `json:"agent_transcript_path"`
	Resolved            bool     `json:"resolved"`
	SizeBytes           int64    `json:"size_bytes"`
	BackgroundTaskIDs   []string `json:"background_task_ids"`
	SessionCronCount    int      `json:"session_cron_count"`
	Effort              string   `json:"effort,omitempty"`
	CaptureError        string   `json:"capture_error,omitempty"`
}

// statFunc is the filesystem probe, injected so both the resolved and the
// unresolved branch are reachable in a test regardless of what is on disk.
type statFunc func(string) (int64, error)

// buildRow is the pure decision: given the hook input and a probe, produce the
// row. It never returns an error — an unresolvable path is DATA, recorded as
// resolved=false with the reason, not a failure that loses the row entirely.
func buildRow(in hookInput, now time.Time, stat statFunc) manifestRow {
	r := manifestRow{
		Schema:              schema,
		Kind:                "seat",
		CapturedAt:          now.UTC().Format(time.RFC3339),
		SessionID:           in.SessionID,
		AgentID:             in.AgentID,
		AgentType:           in.AgentType,
		TranscriptPath:      in.TranscriptPath,
		AgentTranscriptPath: in.AgentTranscriptPath,
		SessionCronCount:    len(in.SessionCrons),
		Effort:              in.Effort.Level,
		BackgroundTaskIDs:   []string{},
	}
	for _, t := range in.BackgroundTasks {
		if t.ID != "" {
			r.BackgroundTaskIDs = append(r.BackgroundTaskIDs, t.ID)
		}
	}
	if in.AgentTranscriptPath == "" {
		r.CaptureError = "hook input carried no agent_transcript_path"
		return r
	}
	size, err := stat(in.AgentTranscriptPath)
	if err != nil {
		r.CaptureError = "agent_transcript_path did not resolve: " + err.Error()
		return r
	}
	r.Resolved = true
	r.SizeBytes = size
	return r
}

// manifestPath keys the manifest by session so concurrent runs cannot interleave
// into one file. Subagents share the parent's session_id, which is exactly what
// makes it the right grouping key here: one manifest per run, one row per seat.
func manifestPath(projectDir, sessionID string) string {
	name := "trajectories-" + sessionID + ".jsonl"
	if sessionID == "" {
		name = "trajectories-unknown-session.jsonl"
	}
	return filepath.Join(projectDir, ".claude", "gray-area", name)
}

// appendRow is best-effort: a manifest that cannot be written must never cost
// the subagent its turn, so failures go to stderr and never change the exit code.
func appendRow(path string, r manifestRow, stderr io.Writer) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: cannot create manifest dir:", err)
		return
	}
	line, err := json.Marshal(r)
	if err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: cannot encode row:", err)
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: cannot open manifest:", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(line, '\n')); err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: cannot append row:", err)
	}
}

// buildSessionRow records what SessionStart handed over: where THIS session's
// transcript is.
//
// It exists because the alternative is a glob of ~/.claude/projects/, which
// plans/gray-area.md §3 rules out on principle — the design is that the harness
// hands the path over deterministically rather than the tool guessing which file
// belongs to whom. SubagentStop already does that for seats; this is the same
// move one event earlier, for the session itself.
func buildSessionRow(in hookInput, now time.Time, stat statFunc) manifestRow {
	r := manifestRow{
		Schema:            schema,
		Kind:              "session",
		CapturedAt:        now.UTC().Format(time.RFC3339),
		SessionID:         in.SessionID,
		TranscriptPath:    in.TranscriptPath,
		BackgroundTaskIDs: []string{},
	}
	if size, err := stat(in.TranscriptPath); err == nil {
		r.Resolved = true
		r.SizeBytes = size
	} else {
		r.CaptureError = "transcript_path did not resolve: " + err.Error()
	}
	return r
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, stat statFunc) int {
	fs := flag.NewFlagSet("gray-area-capture", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	// TAKEN FROM hooks.json, NOT INFERRED FROM THE PAYLOAD. The precedent is
	// prosthetic-conscience's sc-checkpoint-seal: the wiring already knows which
	// event it registered for, and inferring it from which fields happen to be
	// present is a guess that goes wrong silently when the payload shape moves.
	event := fs.String("event", "SubagentStop", "SubagentStop | SessionStart — which hook registered this call")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag is never worth costing a subagent its turn
	}
	if *showVersion {
		fmt.Fprintln(stdout, "gray-area-capture", version)
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)

	// THE PAYLOAD ALSO SAYS WHERE THE PROJECT IS, and this used the environment alone.
	// The refusal below is right — better than the silent no-op the same class produced
	// in prosthetic-conscience's hooks — but it fired while `cwd` sat parsed and unread
	// two lines up, so a session with CLAUDE_PROJECT_DIR unset lost its trajectory for
	// want of a value it had been handed.
	//
	// Resolution order matches prosthetic-conscience's internal/hookenv, which is the
	// canonical statement of this rule; it is restated here rather than imported because
	// the two plugins are separate Go modules by design. In particular there is NO
	// os.Getwd() fallback: "I do not know where the project is" must not become "I will
	// write somewhere", and the somewhere is unrelated to the session.
	if projectDir == "" {
		projectDir = in.CWD
	}
	// No project dir means no durable place to key the manifest; say so once
	// rather than writing a relative .claude/ wherever the process happens to be.
	if projectDir == "" {
		fmt.Fprintln(stderr, "gray-area-capture: no project root (CLAUDE_PROJECT_DIR unset and the hook payload carried no cwd) — trajectory not captured")
		return 0
	}

	if *event == "SessionStart" {
		// THE ALARM (plan §11.3). hook-surface-spike.md §3 states every event carries
		// transcript_path, but that was not re-measured for SessionStart, and a
		// running session cannot fire one to check. So an absent field writes NO ROW
		// and says so: a manifest row whose path is empty would resolve to nothing
		// later, silently, which is the failure this plugin exists to avoid. If this
		// line is ever seen in the wild, the spike's claim is wrong for this event.
		if in.TranscriptPath == "" {
			fmt.Fprintln(stderr, "gray-area-capture: SessionStart carried no transcript_path — no session row written (see plans/gray-area.md §11.3; this refutes the spike's claim that every event carries it)")
			return 0
		}
		appendRow(manifestPath(projectDir, in.SessionID), buildSessionRow(in, now, stat), stderr)
		return 0
	}

	appendRow(manifestPath(projectDir, in.SessionID), buildRow(in, now, stat), stderr)
	return 0
}

func statSize(p string) (int64, error) {
	st, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	if st.IsDir() {
		return 0, fmt.Errorf("is a directory")
	}
	return st.Size(), nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now(), statSize))
}
