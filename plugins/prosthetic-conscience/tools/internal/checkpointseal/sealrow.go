package checkpointseal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/freshness"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/statefile"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/stopnudge"
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

	// WrittenAt is the note's OWN claim about when its body last changed, recorded
	// beside the hash so F6 is computable FROM THE RECORD: a written_at that advanced
	// between two rows while body_sha did not is a note claiming to be fresh when it is
	// not. Without this field the hash had no comparable and no reader.
	WrittenAt string `json:"written_at,omitempty"`

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

	// GrowthSince is WHEN the growth baseline was taken. Nothing observes a note at the
	// moment it is written — the callers run at boundaries — so a note written at turn 10
	// and first sealed at turn 200 measures growth from turn 200. The number is real; the
	// interval is not the one the field name suggests, and this makes the difference
	// visible rather than leaving it to be assumed.
	NoteGrowthTokens *int   `json:"note_growth_tokens,omitempty"`
	GrowthMeasured   bool   `json:"growth_measured"`
	GrowthSince      string `json:"growth_since,omitempty"`

	NoteBranchCommits *int `json:"note_branch_commits,omitempty"`
	BranchMeasured    bool `json:"branch_measured"`

	// NudgeEnabled says whether the nudge was live when this row was written. Criterion
	// 6 compares the two populations, and Phase 1 rows are the "before" half — a row
	// that does not carry the key at all makes `select(.nudge_enabled==false)` return
	// nothing, so the kill switch would compare a real "after" against an empty
	// "before" and report whatever it liked.
	NudgeEnabled bool `json:"nudge_enabled"`

	// NudgeMeasured is the tri-state this record keeps everywhere else, applied to the
	// one column that did not have it. nudge.json ABSENT means the nudge is genuinely off;
	// UNREADABLE means it may be on and this row cannot say. Folding the second into
	// `nudge_enabled: false` would file an unreadable state under the "before" population
	// that criterion 6 compares against — a manufactured control.
	NudgeMeasured bool `json:"nudge_measured"`

	// The emission counters criterion 4 is DECIDED by. They are owned by stopnudge and
	// live in nudge.json; this row copies them because the gate queries seals.jsonl.
	//
	// They were missing entirely until the plan's own queries were checked against this
	// struct. `map(.emissions_this_session) | max` over rows that never carried the key is
	// null, `null > 4` is false in jq, and the budget gate for the whole nudge design
	// therefore PASSED by returning nothing.
	//
	// POINTERS, so a row with no reading omits them rather than reporting a zero. A 0
	// emission count is a real and interesting value — it is what a well-behaved session
	// looks like — and it must not share a spelling with "not measured".
	EmissionsThisSession *int `json:"emissions_this_session,omitempty"`
	EmissionBytesMax     *int `json:"emission_bytes_max,omitempty"`

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
func appendSealRow(dir, projectDir string, body []byte, now time.Time, event, occ string, in hookInput, stderr io.Writer, age Measures, writtenAt string) {
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
		WrittenAt:       writtenAt,
		NudgeAnswered:   "n/a",
		TurnsMeasured:   age.TurnsMeasured,
		GrowthMeasured:  age.GrowthKnown,
		BranchMeasured:  age.BranchKnown,
	}
	if measured {
		row.LiveHandles = &n
	}
	readNudgeState(projectDir, &row)
	reportImpossibleWrittenAt(writtenAt, now, binaryFor(event), stderr)
	if age.TurnsMeasured {
		t := age.Turns
		row.NoteAgeTurns = &t
	}
	if age.GrowthKnown {
		g := age.Growth
		row.NoteGrowthTokens = &g
		if !age.GrowthSince.IsZero() {
			row.GrowthSince = age.GrowthSince.UTC().Format(time.RFC3339)
		}
	}
	if age.BranchKnown {
		c := age.BranchCommits
		row.NoteBranchCommits = &c
	}

	if err := statefile.AppendRow(filepath.Join(dir, "seals.jsonl"), row); err != nil {
		// Named for the shim that was INVOKED, not for the binary these three used to be.
		fmt.Fprintln(stderr, binaryFor(event)+": cannot append seal row:", err)
	}
}

// readNudgeState copies the nudge's own counters onto the row.
//
// The sealer is the only writer of seals.jsonl and stopnudge is the only writer of
// nudge.json, so this is a read across an ownership boundary and stays one: it takes the
// path from stopnudge.StatePath rather than spelling it again, and it never writes.
//
// The tri-state is the whole point. Absent is an honest "the nudge is off"; Unreadable is
// "this row cannot say", and the two must not arrive at the same column value.
func readNudgeState(projectDir string, row *sealRow) {
	// The PROJECT ROOT, passed in. It was briefly recovered from the snapshot directory by
	// two filepath.Dir calls — a fact taken back out of a path, which is the shape this
	// design spent three rounds removing from everywhere else.
	st, status := statefile.Read[stopnudge.State](stopnudge.StatePath(projectDir))
	switch status {
	case statefile.Absent:
		row.NudgeMeasured = true // the nudge is off, and that is a reading
	case statefile.Present:
		row.NudgeMeasured = true
		row.NudgeEnabled = true
		e, b := st.Emissions, st.EmissionBytes
		row.EmissionsThisSession = &e
		row.EmissionBytesMax = &b
	default:
		// Unreadable: leave NudgeMeasured false and both counters omitted. Nothing here
		// is written as a zero.
	}
}

// futureGrace is how far ahead of the sealer's clock a note may claim to have been written
// before that claim is called impossible. The agent and the hook read the SAME machine clock,
// so any gap is authorship rather than skew; a minute absorbs a note composed across a seal
// without absorbing the failure this looks for.
const futureGrace = time.Minute

// reportImpossibleWrittenAt complains when a note claims to have been written AFTER the seam
// that is sealing it.
//
// Naming the command that produces `written_at` is policy; this is the mechanism, and the
// rule-sweep gate is right that the first without the second is how a fix stays an instance.
// Nothing could previously distinguish a clock reading from a typed guess, and a real session
// produced four notes running whose stamps were round numbers — one of them seven minutes in the
// FUTURE. That value silently becomes the age every measurement is taken from.
//
// It goes to stderr because that channel is MEASURED to reach the agent: it arrives inside the
// tool result of whatever call was running when the hook fired (spike §2a's attachment settles
// the injected channel; this one was confirmed directly from a live transcript).
func reportImpossibleWrittenAt(writtenAt string, now time.Time, bin string, stderr io.Writer) {
	if writtenAt == "" {
		return // absent is a schema-2 note, not a false claim
	}
	t, err := time.Parse(time.RFC3339, writtenAt)
	if err != nil {
		return // unparsable is a different fault and not this function's to report
	}
	if t.After(now.Add(futureGrace)) {
		fmt.Fprintf(stderr, "%s: this note says written_at %s, which is AFTER the seam sealing it "+
			"at %s — a note cannot be written in the future, so that stamp was typed rather than "+
			"read. The age every measurement is taken from is this field: set it with "+
			"`date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ` before you start composing.\n",
			bin, writtenAt, now.UTC().Format(time.RFC3339))
	}
}
