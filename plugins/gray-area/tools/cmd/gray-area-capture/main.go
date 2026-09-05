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
//
// SAID IN A FIELD as of schema 2 (#398). Until then this comment was the only place
// the distinction existed: the row itself reported both populations with one error
// string, so a reader counting capture failures got a number that meant nothing and
// three analyses read the permanent category as a transient filesystem error
// (plans/gray-area.md §11.8, §11.9). A comment is not a schema — see capture_category.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/buildid"
	"sort"
)

// schema is the manifest row version. The transcript format is vendor-internal
// and version-unstable, so every row records what produced it.
//
// 2 adds capture_category (#398). The bump is not ceremony: the field is omitempty, so a
// reader cannot otherwise tell "an old binary wrote this row and had no category" from
// "a new binary wrote it and the row resolved" — both omit the key. Gating on schema >= 2
// is what makes the field's ABSENCE mean something, which is the whole point of adding it.
//
// 4 changes what `kind` MEANS (#189), which is the reason it is a bump and not an edit. Rows
// 1-3 called every SubagentStop a seat. Schema 4 does not, and a reader counting seats across
// the boundary would otherwise mix a population of 19 with a population of 165. See kindTurnEnd.
const schema = 4

// The closed set of row kinds. `kind` is the first field any reader filters on, so what it
// MEANS is load-bearing, and until schema 4 it was wrong for 88% of the rows in this
// repository's own manifest.
//
// # kindTurnEnd, and why SubagentStop is not evidence of a subagent (#189)
//
// SubagentStop fires for something other than an `Agent` tool call, and rows 1-3 filed all of
// it as seats. Measured on session 937047bc — 165 SubagentStop rows, 19 with a type:
//
//	IDENTITY    19 typed rows -> 19 distinct agent ids -> 19/19 present under <session>/subagents/.
//	            146 typeless rows -> 146 distinct agent ids -> 0/146 present. Each id fires once.
//	POSITION    139/146 typeless captures land within 120s AFTER a main-agent turn end; the
//	            nearest preceding transcript record is the assistant's closing text (96) or the
//	            Stop hook's own summary (43). 166 of the 197 turn ends in the hook's window are
//	            followed by one — 84%, and 80-90% on every one of the eight days.
//	FALSIFIER   across 3406 mid-turn windows (an assistant tool_use through its tool_result),
//	            0/146 typeless captures land inside one. 2/19 typed do. A subagent completing
//	            is BY CONSTRUCTION mid-turn: the parent is blocked on the Agent call. Nothing
//	            that fires only at turn boundaries is a subagent finishing.
//	MECHANISM   20 transcripts on disk, 19 carrying a .meta.json holding agentType — written
//	            for Agent-tool spawns. 19 metas, 19 typed rows. Supporting, not load-bearing:
//	            the three above stand without it.
//
// So the harness fires this hook at the MAIN agent's Stop with a freshly minted agent id, no
// meta sidecar, and a transcript path it PREDICTS and nothing ever writes to. The path is not
// a broken pointer; it is a forecast.
//
// WHAT THIS OVERTURNS. plans/gray-area.md §11.10 concluded the substrate was blind to 72% of
// seats, and plans/gray-area-phase-2.md scoped Phase 2 around that number. Both counted turn
// ends in the denominator. Every seat this repository ever spawned resolved: 19/19.
//
// WHAT IT DOES NOT CLAIM. 84% is not 100%: some turn ends produce no row (windows where the
// hook binary was absent are the known cause, and are not separately measured). The
// classification here is made from the ROW's own two fields, never from that correlation —
// see buildRow.
const (
	kindSeat    = "seat"     // a subagent's trajectory, handed over at SubagentStop
	kindTurnEnd = "turn-end" // a SubagentStop that named no seat; see above
	kindSession = "session"  // the MAIN session's trajectory, handed over at SessionStart
)

// The closed set of reasons a row is unresolved. A counter downstream needs to separate the
// expected population from the alarming one, and a prose error string cannot be counted:
// `capture_error` says it in English for a human, this says it in a word for a machine.
//
// MEASURED (#398, #189): across 69 seats in one session, `agent_type` empty and
// `resolved: false` coincided 50/50 with zero exceptions, and every one of the 19 typed
// seats resolved. The uniform error text read as a filesystem problem — a path that might
// come back on a retry — while it was a permanent property of the seat, knowable before
// anything was stat'ed. That misreading cost three wrong answers (plans/gray-area.md §11.8,
// §11.9) before the correlation was noticed.
const (
	// captureUntypedSeat: the event named no seat — no agent_type, and no transcript at the
	// path it predicted. NOT an alarm, and as of #189 not a seat either: these are main-agent
	// turn ends arriving on the SubagentStop hook. See kindTurnEnd for the measurement.
	captureUntypedSeat = "event-names-no-seat"
	// captureMissing: a TYPED seat's transcript did not stat. This is the alarming one — the
	// population where a file is expected — and one file in 16 arrived after its stat, so the
	// race is real here and must stay visible.
	captureMissing = "typed-seat-transcript-missing"
	// captureNoPath: the hook input carried no path at all, so nothing was even attempted.
	captureNoPath = "hook-input-carried-no-path"
)

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
	// HookEventName is what the HARNESS calls this event. The wiring already knows which hook
	// it registered for and passes it with -event; this is the payload's own claim, recorded
	// so the two can be COMPARED rather than assumed equal. See manifestRow.HookEventName.
	HookEventName string `json:"hook_event_name"`
	// The remaining non-content scalars the payload carries. Recorded together rather than one
	// per investigation round: each round costs a build, a merge and a wait for an event to
	// fire, and the questions these answer are already foreseeable.
	PermissionMode string `json:"permission_mode"`
	PromptID       string `json:"prompt_id"`
	StopHookActive bool   `json:"stop_hook_active"`
}

// manifestRow is one seat's entry. Field names are snake_case to match the
// harness payloads a reader will have alongside them.
type manifestRow struct {
	Schema int `json:"schema"`
	// Kind separates the rows a manifest can hold — see kindSeat, kindTurnEnd, kindSession.
	// A consumer asking "where is this session's transcript" must not be answered with a
	// seat's, and a consumer counting seats must not be handed the turn ends.
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
	// CaptureCategory names WHY the row is unresolved, from the closed set above, so the
	// expected population can be counted apart from the alarming one. Empty when Resolved —
	// a resolved row has nothing to categorise. Present on EVERY unresolved row from
	// schema 2 onward, which is what lets a reader treat its absence as meaningful.
	CaptureCategory string `json:"capture_category,omitempty"`
	// PayloadKeys is the top-level KEY NAMES the hook payload carried, on EVERY seat row.
	//
	// WHY NAMES AND NOT VALUES. #189 is undetermined after four investigations, and the reason
	// the fourth also stalled is that this program records the eight fields it already models
	// and nothing about what else arrived — so a population it cannot explain is a population
	// it cannot even describe. Measured 2026-08-17 on this repository's own manifest: 141 seat
	// rows carry no agent_type and no transcript, and NONE of them corresponds to an `Agent`
	// tool call (20 in the whole session, all 19 typed rows accounted for). Whatever fires
	// SubagentStop for them is not an Agent-tool subagent, and no field here says what it is.
	//
	// Values are NOT recorded, and that is not timidity: the payload carries
	// `last_assistant_message`, which this program deliberately refuses to copy (see the header
	// and plans/gray-area.md G6). Key names describe the SHAPE of an event without spreading
	// conversation content into a second file — which is the whole posture of this plugin.
	//
	// ON EVERY ROW, AND THE FIRST VERSION GOT THIS WRONG. Schema 3 recorded keys on UNRESOLVED
	// rows only, reasoning that a resolved row has nothing to explain. But the stated
	// falsification test for #189 is "compare a typeless row's key set against a TYPED row's" —
	// and typed rows resolve, so they never carried keys. The instrument made the comparison it
	// was built for impossible. A field recorded only where the answer is already suspected
	// cannot distinguish the populations; it can only describe one of them.
	PayloadKeys []string `json:"payload_keys,omitempty"`
	// DeclaredEvent is what the WIRING said (-event, from hooks.json). HookEventName is what the
	// PAYLOAD said. Both are recorded because #189 turns on whether they agree.
	//
	// Schema 3 recorded payload key NAMES and immediately earned its keep: the first typeless row
	// carried `hook_event_name`, `stop_hook_active`, `prompt_id` and `permission_mode` — four
	// fields this program had never modelled, on an event it could not explain. Names alone
	// cannot say WHICH event it was, and that is the question.
	//
	// A VALUE IS RECORDED HERE AND NOT ELSEWHERE, deliberately: `hook_event_name` is harness
	// metadata naming its own event, not conversation content. `last_assistant_message` sits in
	// the same payload and is still refused (G6). The distinction is what the field IS, not how
	// convenient it would be.
	DeclaredEvent string `json:"declared_event,omitempty"`
	HookEventName string `json:"hook_event_name,omitempty"`
	// PermissionMode, PromptID and StopHookActive are the payload's remaining non-content
	// scalars. All three are harness metadata; `last_assistant_message` is the content field in
	// the same payload and stays refused (G6).
	//
	// RECORDED IN ONE ROUND, not one per question. The first two instruments each answered
	// exactly one question and cost a build, a merge and a wait for an event to fire. PromptID
	// in particular is the one that can say whether these events are scoped to a MAIN-SESSION
	// prompt rather than to a subagent — which is the live hypothesis now that declared_event
	// and hook_event_name have been measured to AGREE (both `SubagentStop`, so the wiring is
	// not lying and the harness really does call them that).
	PermissionMode string `json:"permission_mode,omitempty"`
	PromptID       string `json:"prompt_id,omitempty"`
	StopHookActive bool   `json:"stop_hook_active,omitempty"`
}

// payloadKeys returns the sorted top-level key names of a JSON object, or nil.
//
// Sorted so two rows are comparable at a glance; a set, not a sequence, because JSON object
// order carries no meaning and an unstable list would read as a changing payload.
func payloadKeys(raw []byte) []string {
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// statFunc is the filesystem probe, injected so both the resolved and the
// unresolved branch are reachable in a test regardless of what is on disk.
type statFunc func(string) (int64, error)

// buildRow is the pure decision: given the hook input and a probe, produce the
// row. It never returns an error — an unresolvable path is DATA, recorded as
// resolved=false with the reason, not a failure that loses the row entirely.
func buildRow(in hookInput, raw []byte, declaredEvent string, now time.Time, stat statFunc) manifestRow {
	r := manifestRow{
		Schema:              schema,
		Kind:                kindSeat,
		CapturedAt:          now.UTC().Format(time.RFC3339),
		SessionID:           in.SessionID,
		AgentID:             in.AgentID,
		AgentType:           in.AgentType,
		TranscriptPath:      in.TranscriptPath,
		AgentTranscriptPath: in.AgentTranscriptPath,
		SessionCronCount:    len(in.SessionCrons),
		PayloadKeys:         payloadKeys(raw),
		DeclaredEvent:       declaredEvent,
		HookEventName:       in.HookEventName,
		PermissionMode:      in.PermissionMode,
		PromptID:            in.PromptID,
		StopHookActive:      in.StopHookActive,
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
		r.CaptureCategory = captureNoPath
		return r
	}
	// STAT EVEN THE UNTYPED SEATS, and categorise from what happened rather than from what
	// was predicted. The 50/50 correlation is one session's evidence, not a law, and what
	// makes agent_type empty is NOT determined — six probe agents across every subagent_type
	// (and omitted) all carried a type and all resolved. Skipping the stat for untyped seats
	// would turn a strong correlation into an assumption the row could never contradict, and
	// a row that cannot be surprised has stopped being a measurement.
	size, err := stat(in.AgentTranscriptPath)
	if err != nil {
		r.CaptureError = "agent_transcript_path did not resolve: " + err.Error()
		if in.AgentType == "" {
			// THE CONJUNCTION, AND ONLY THE CONJUNCTION, RECLASSIFIES. No type AND no file is
			// the measured signature of a turn end (kindTurnEnd). Either half alone stays a
			// seat: a typed row that did not stat is the alarm below, and an untyped row that
			// DID stat is a real trajectory whose name is missing — surprising, and a surprise
			// must not be filed under an explanation. Neither has ever been observed, which is
			// exactly why neither may be folded in silently.
			r.Kind = kindTurnEnd
			r.CaptureCategory = captureUntypedSeat
		} else {
			r.CaptureCategory = captureMissing
		}
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
//
// # APPEND-ONLY, AND A REPEAT IS NOT SUPPRESSED HERE
//
// SubagentStop fires again for a seat that is continued — measured on this
// repository's own manifest, seat a703a4ea4a2d4e09d captured at 07:08:23Z and again
// at 07:11:10Z with its transcript grown from 356468 to 452844 bytes. The second row
// is a true observation and is written. Making the repeat harmless is the READER's
// job (claims.SeatCensus); doing it here would need a read-modify-write on a path
// that must never block a subagent, would race every other seat stopping in the same
// moment, and would discard the one fact the second row carries.
//
// # THE TORN TAIL, which is what an interrupted append actually leaves
//
// One row is one write(2) under O_APPEND, so concurrent seats cannot interleave. What
// a kill or an ENOSPC CAN leave is a final line with no newline on it — and the next
// append then lands on the same line, so an interruption that cost one row costs two.
// Terminating an unterminated tail before writing bounds the damage to the row that
// was actually cut. It is idempotent (a file already ending in \n is untouched) and
// safe to race: two healers write two newlines, and a blank line is skipped by every
// reader. The line that was cut stays in the file as an unreadable line, which is
// deliberate — claims.SeatCensus.Unreadable counts it, so a seat lost to a torn write
// can be told from a seat the hook never saw.
//
// Sync is called because the row is the only durable trace that a seat existed: a
// transcript on disk that no row names is #469's signature, and losing the row in the
// page cache on a machine crash manufactures exactly that. One fsync of a ~1 KB
// append, on a path already doing a stat and an open.
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
	// O_RDWR rather than O_WRONLY only so the tail can be READ; O_APPEND still
	// governs every write, which is what keeps concurrent seats from interleaving.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: cannot open manifest:", err)
		return
	}
	defer func() { _ = f.Close() }()
	healTornTail(f, stderr)
	if _, err := f.Write(append(line, '\n')); err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: cannot append row:", err)
		return
	}
	if err := f.Sync(); err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: row appended but not synced:", err)
	}
}

// healTornTail terminates a manifest whose last line has no newline, so the next
// append cannot be concatenated onto a row an interruption cut short.
//
// Silent on every ordinary call, and silent when it cannot look: a probe that fails
// must not cost the row it was protecting. It reports only when it actually repairs
// something, because that is a fact about a PREVIOUS run that nothing else records.
func healTornTail(f *os.File, stderr io.Writer) {
	st, err := f.Stat()
	if err != nil || st.Size() == 0 {
		return
	}
	var last [1]byte
	if _, err := f.ReadAt(last[:], st.Size()-1); err != nil || last[0] == '\n' {
		return
	}
	if _, err := f.Write([]byte{'\n'}); err != nil {
		fmt.Fprintln(stderr, "gray-area-capture: manifest tail is unterminated and could not be closed:", err)
		return
	}
	fmt.Fprintln(stderr, "gray-area-capture: closed an unterminated manifest tail — a previous append was cut short, and that line will not parse")
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
		Kind:              kindSession,
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
		fmt.Fprintln(stdout, buildid.Line("gray-area-capture"))
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

	appendRow(manifestPath(projectDir, in.SessionID), buildRow(in, raw, *event, now, stat), stderr)
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
