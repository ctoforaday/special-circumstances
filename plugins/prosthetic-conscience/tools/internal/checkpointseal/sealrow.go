package checkpointseal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/freshness"
)

// Measures is freshness.Measures, aliased so this file reads without qualifying every
// field and so the dependency direction stays visible in one place.
type Measures = freshness.Measures

// sealRow is one line of .claude/checkpoints/seals.jsonl — the RECORD this design
// measures over.
//
// It exists because the seal did not. A seal was an HTML comment prepended to a
// markdown snapshot, in a directory pruned to keepSnapshots, so a baseline over
// twenty boundaries could not be reconstructed from seals at all: the eleventh
// oldest is deleted. And the stamp itself composes five facts into a string and
// recovers them by parsing, which is the shape this suite is named after. The
// snapshot and its stamp stay exactly as they are — they serve human recovery —
// and the machine-read facts move here, into fields a reader cannot mis-parse.
//
// Append-only, deliberately, against the keepSnapshots precedent beside it: a row
// is ~200 bytes, the file is gitignored, and pruning is precisely what made the
// old seal useless as a record.
type sealRow struct {
	At          string `json:"at"`
	Event       string `json:"event"`
	Occasion    string `json:"occasion"`
	SessionID   string `json:"session_id"`
	AgentID     string `json:"agent_id,omitempty"`
	SealTrigger string `json:"seal_trigger"`
	BodySHA     string `json:"body_sha"`

	// LiveHandles is a POINTER so that "the event cannot tell me" and "there was no
	// background work" stay different answers. PreCompact and SessionEnd carry no
	// background_tasks key at all (hook-surface-spike.md §12); only SubagentStop and
	// Stop do, and Stop does not seal. A plain int would write 0 for both cases, on
	// the one column #506's verdict turns on.
	LiveHandles     *int `json:"live_handles,omitempty"`
	HandlesMeasured bool `json:"handles_measured"`

	// The note's AGE, in the three units that fail independently. Each carries its own
	// measured flag and is OMITTED when unmeasured, never written as a zero: criterion 1
	// is a distribution over these, and a zero that meant "could not tell" would pull
	// every median toward fresh.
	NoteAgeTurns  *int `json:"note_age_turns,omitempty"`
	TurnsMeasured bool `json:"turns_measured"`

	NoteGrowthTokens *int `json:"note_growth_tokens,omitempty"`
	GrowthMeasured   bool `json:"growth_measured"`

	NoteBranchCommits *int `json:"note_branch_commits,omitempty"`
	BranchMeasured    bool `json:"branch_measured"`

	// CeilingKnown says whether a proximity figure was computable AT ALL — there is no
	// context-window field anywhere, so the only ceiling is this session's own first
	// compaction. Recorded as its own fact so the baseline can say how often it is
	// absent rather than leaving that to be inferred.
	CeilingKnown bool `json:"ceiling_known"`

	// NudgeAnswered is derived by the SEALER, never self-reported by an agent. In
	// Phase 1 no nudge exists, so every row carries "n/a" — which is a value, not an
	// empty string, because a blank would be indistinguishable from a field nothing
	// wrote.
	NudgeAnswered string `json:"nudge_answered"`
}

// sealTriggers maps the sealing event to the name the baseline groups by. The
// three are separate because the question is which BOUNDARY the note was old at,
// and a seat returning is not a session ending.
var sealTriggers = map[string]string{
	evPreCompact:   "precompact",
	evSessionEnd:   "sessionend",
	evSubagentStop: "seat_return",
}

// handleTask is the shape of one background_tasks entry. Only the discriminator is
// read: a seat that is still running appears in its PARENT's handle list as
// type "subagent" (spike §12), and counting it would answer "is anything running"
// when the question is "did this note miss some background work".
type handleTask struct {
	Type string `json:"type"`
}

// countHandles reports how many non-subagent handles are live, and whether the
// payload could answer at all.
//
// The pointers ARE the mechanism. json.Unmarshal into a []T maps an absent key and
// an empty array to the same nil, which would collapse "PreCompact does not carry
// this" into "nothing was running" — the absent case and the honest zero as the
// same bytes. A present-but-empty array is a MEASURED zero and is the normal seat
// return, so the distinction is not a corner case.
func countHandles(tasks, crons *json.RawMessage) (n int, measured bool) {
	if tasks == nil && crons == nil {
		return 0, false
	}
	if tasks != nil {
		var list []handleTask
		if err := json.Unmarshal(*tasks, &list); err != nil {
			// Present but unreadable is NOT zero: say so rather than report a clean board.
			return 0, false
		}
		for _, t := range list {
			if t.Type != "subagent" {
				n++
			}
		}
	}
	if crons != nil {
		var list []json.RawMessage
		if err := json.Unmarshal(*crons, &list); err != nil {
			return 0, false
		}
		n += len(list)
	}
	return n, true
}

// appendSealRow writes one row for this seal.
//
// Best-effort, matching this package's posture: a recorder that fails loudly on a
// hook path costs a session its restore over a lost observation. (The nudge is the
// opposite — it fails CLOSED — because a lost row costs one measurement while an
// unrecorded emission costs a loop.)
func appendSealRow(dir string, body []byte, now time.Time, event, occ string, in hookInput, stderr io.Writer, age Measures) {
	sum := sha256.Sum256(body)
	n, measured := countHandles(in.BackgroundTasks, in.SessionCrons)
	row := sealRow{
		At:              now.UTC().Format(time.RFC3339),
		Event:           event,
		Occasion:        occ,
		SessionID:       in.SessionID,
		AgentID:         in.AgentID,
		SealTrigger:     sealTriggers[event],
		BodySHA:         hex.EncodeToString(sum[:]),
		HandlesMeasured: measured,
		NudgeAnswered:   "n/a",
		TurnsMeasured:   age.TurnsMeasured,
		GrowthMeasured:  age.GrowthKnown,
		BranchMeasured:  age.BranchKnown,
		CeilingKnown:    age.ProximityKnown,
	}
	if measured {
		row.LiveHandles = &n
	}
	if age.TurnsMeasured {
		t := age.Turns
		row.NoteAgeTurns = &t
	}
	if age.GrowthKnown {
		g := age.Growth
		row.NoteGrowthTokens = &g
	}
	if age.BranchKnown {
		c := age.BranchCommits
		row.NoteBranchCommits = &c
	}

	line, err := json.Marshal(row)
	if err != nil {
		fmt.Fprintln(stderr, "sc-checkpoint-seal: cannot encode seal row:", err)
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "seals.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(stderr, "sc-checkpoint-seal: cannot open seal record:", err)
		return
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		fmt.Fprintln(stderr, "sc-checkpoint-seal: cannot append seal row:", err)
	}
}
