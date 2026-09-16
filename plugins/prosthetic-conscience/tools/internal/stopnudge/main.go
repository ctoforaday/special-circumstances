package stopnudge

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/ctxusage"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/freshness"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookmain"
)

// hookInput is the Stop payload. Every field here was MEASURED on client 2.1.240
// (hook-surface-spike.md §9c, §13) rather than taken from documentation.
//
// Note what is absent: no agent_id and no agent_type. Stop carries none of the agent_*
// fields, which is why the debounce is keyed on session_id alone — a composite key would
// have kept half of itself permanently empty.
type hookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	StopHookActive bool   `json:"stop_hook_active"`
}

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext,omitempty"`
	} `json:"hookSpecificOutput"`
	// SystemMessage is for the HUMAN, and is a different audience from AdditionalContext, which is
	// for the agent. The failure record travels here: a broken hook is not something to ask the
	// agent to relay.
	SystemMessage string `json:"systemMessage,omitempty"`
}

// configured reads the band edges.
//
// THE NUMBERS ARE THE PHASE 1 BASELINE AT §III'S PREREGISTERED RULE, and nothing here was chosen
// after seeing them. §III fixed the rule before any data existed — three measures, each banded at
// P50/P75/P90 of ITS OWN distribution, combined ANY-OF taking the max — precisely so the edges
// could not be picked to flatter the result. These are the nearest-rank percentiles of the 64-row
// baseline (2026-08-23..28, all nudge_enabled:false), unrounded, because no rounding rule was
// preregistered and rounding after the fact is the freedom §III forecloses.
//
//	measure              measured n    NOTICE(P50)   WARN(P75)   URGENT(P90)
//	note_age_turns          22 / 64            16          39            47
//	note_growth_tokens      57 / 64        66_764     109_902       157_844
//	note_branch_commits     64 / 64            10          12            17
//
// TWO THINGS THE DATA SAYS THAT THE RULE DOES NOT FIX, reported here rather than smoothed away —
// §III's own instruction for a distribution that comes out awkward is that it is "a finding to
// report and re-decide with the human", never a licence to quietly choose different percentiles.
//
//  1. TURNS NOTICE IS UNREACHABLE. floorTurns is 20 and the measured P50 is 16, so nothing can
//     cross the turns NOTICE edge without already clearing the floor at 20. The edge is left at
//     its preregistered value rather than nudged to 20: moving it would be choosing a number after
//     seeing the data, which is the one thing this design was built to prevent. The effective
//     turns NOTICE is the floor.
//  2. THE TURNS DISTRIBUTION IS RIGHT-CENSORED. Turns were measured on 22 of 64 rows, and the rows
//     where they were NOT measured are the stale ones — mean wall age 10.5h against 2.0h where
//     they were. So the stalest notes are exactly the ones these percentiles never saw, and the
//     turns edges undershoot by an unknown margin. Growth (57/64) and branch (64/64) are close to
//     complete and carry the gate in practice.
//
// The baseline is one machine over five days. Phase 3 falsifies these edges against a nudge-on
// population; until then they are the honest reading of the only distribution that exists.
//
// THE CONTEXT EDGES HAVE NO BASELINE, and say so. No session without a note was ever measured, so
// there was no distribution to take percentiles of; gblock chose 150k / 300k / 600k live tokens on
// 2026-09-15, weighing interruptions against cost — every turn re-reads the whole context, so a
// heavy context is paid for on each one.
func configured() Thresholds {
	return Thresholds{
		TurnsNotice: 16, TurnsWarn: 39, TurnsUrgent: 47,
		GrowthNotice: 66_764, GrowthWarn: 109_902, GrowthUrgent: 157_844,
		BranchNotice: 10, BranchWarn: 12, BranchUrgent: 17,
		ContextNotice: 150_000, ContextWarn: 300_000, ContextUrgent: 600_000,
	}
}

// run takes its thresholds as an ARGUMENT rather than reading configured() directly, so
// the emission path can be driven in a test.
//
// That is not test scaffolding for its own sake. With a hardcoded source, a test could reach
// the emit path only by building a session that crosses the live edges — and the branch that
// would go untested is the one that composes the response the client actually reads. A hook
// whose emit path has never run is a hook whose first real emission is its first
// execution of that code.
// The stages sc-stop can fail at, each cleared by its own next success.
const (
	StageNoteRead    hookfailures.Stage = "note-read"
	StageEncode      hookfailures.Stage = "encode"
	StageStateRead   hookfailures.Stage = "nudge-state-read"
	StageStateWrite  hookfailures.Stage = "nudge-state-write"
	StageStateRemove hookfailures.Stage = "nudge-state-remove"
)

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, th Thresholds) int {
	if hookmain.Preamble(args, stdout, stderr, hookmain.Named("sc-stop")) {
		return 0 // a bad flag is never worth disturbing the session over
	}
	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	rec := hookfailures.New("prosthetic-conscience", "sc-stop", "Stop", now, stderr)
	if !hookenv.Explain(projectDir, rec, "sc-stop") {
		// Recorded, and NOT said: a hook with no project root writes nothing to stdout, which
		// projectroot_test.go pins for every binary here. The next event with a root says it.
		_ = rec.Settle()
		return 0
	}

	notePath := checkpoint.NotePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}, filepath.Glob)
	var d Decision
	if notePath == "" {
		// No note: nothing has an age, but the context still has a size. The zero time keeps the
		// read to a backward scan for the newest usage figure — measured p95 1.2 ms on a 36 MB
		// transcript, against 6.6 ms for the with-note gauge every noted session already pays.
		u, _ := ctxusage.Read(in.TranscriptPath, time.Time{})
		d = DecideContext(rec, projectDir, in.SessionID, in.StopHookActive, u, th)
	} else {
		body, err := os.ReadFile(notePath)
		if err != nil {
			// The note is THERE — NotePath stat'd it — and unreadable, so the nudge cannot measure
			// anything. Silence here reads exactly like a session with no note.
			rec.FailIn(StageNoteRead, projectDir, "cannot read the checkpoint note "+notePath+": "+err.Error())
			emitResponse(stdout, "", rec)
			return 0
		}
		rec.OKIn(StageNoteRead, projectDir)
		n := checkpoint.Parse(string(body))
		m := freshness.Of(projectDir, in.TranscriptPath, string(body),
			freshness.BranchWork(n.Get("head")), now)
		d = Decide(rec, projectDir, in.SessionID, in.StopHookActive, m, notePath, newest(n), th, now)
	}
	emitResponse(stdout, d.Emit, rec)
	return 0
}

// emitResponse writes this event's ONE document: the nudge for the agent, the record's message for
// the human, or nothing at all. Stop displays a systemMessage, which makes this hook the carrier for
// every failure recorded on an event that displays none.
func emitResponse(stdout io.Writer, emit string, rec *hookfailures.Recorder) {
	msg := rec.Settle()
	if emit == "" && msg == "" {
		return // silence is a valid outcome and MUST stay silent
	}
	var out hookOutput
	out.HookSpecificOutput.HookEventName = "Stop"
	out.HookSpecificOutput.AdditionalContext = emit
	out.SystemMessage = msg
	if err := json.NewEncoder(stdout).Encode(out); err != nil {
		// A response that cannot be encoded is a nudge that never arrives. Recorded rather than
		// said, because the channel for saying it is the thing that just failed; the next
		// displaying event reads it out.
		rec.Fail(StageEncode, "cannot encode response: "+err.Error())
		_ = rec.Settle()
	}
}

// newest is the later of written_at and reaffirmed_at — the note's most recent ANSWER of
// either kind. Bands re-arm against this, so a re-affirmation closes a band exactly as a
// rewrite does, which is what the skill clause promises when it calls a reasoned "still
// accurate" a valid answer.
func newest(n checkpoint.Note) time.Time {
	var out time.Time
	for _, k := range []string{"written_at", "reaffirmed_at"} {
		if t, err := time.Parse(time.RFC3339, n.Get(k)); err == nil && t.After(out) {
			out = t
		}
	}
	return out
}

// Main is the process boundary: it wires the real environment in and returns the exit
// code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now(), configured())
}
