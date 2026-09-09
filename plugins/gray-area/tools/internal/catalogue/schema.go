// Package catalogue is the trajectory miner's store: what every local agent did, said and
// thought, queryable by any other session.
//
// # The views are the contract, the tables are not
//
// Agents write their own SQL against this (plans/gray-area-catalogue.md §II), so a renamed column
// breaks work that is not in this repository and cannot be swept. The base tables are therefore
// free to change and the VIEWS are not: they evolve additively, and `schema_test.go` pins their
// exact column sets so a rename fails here rather than in somebody's query.
//
// That discipline is owed to #819, which measured 0 of 7 run archives readable by any current
// binary after words were retired and an epoch moved. An archive that outlives its source and
// cannot be read is worth nothing.
package catalogue

// Schema is the whole DDL. It is one string rather than a migration chain because the store is
// DERIVED: nothing here is authoritative, so a shape change is answered by deleting the file and
// reprojecting rather than by migrating it. `PRAGMA user_version` records which shape wrote it,
// and Open refuses a database from a newer one rather than reading it wrongly.
const Schema = `
CREATE TABLE IF NOT EXISTS session (
    session_id    TEXT PRIMARY KEY,
    project_dir   TEXT NOT NULL,
    cwd           TEXT NOT NULL DEFAULT '',
    first_seen    INTEGER NOT NULL,
    last_seen     INTEGER NOT NULL,
    -- closed_at is closure's marker: NULL means not yet closed, which is the safe reading
    -- because it means the session is still eligible for the pending queue. Nothing else
    -- records it; the queue itself is a QUERY over this column, not a table.
    closed_at     INTEGER,
    capture_build TEXT NOT NULL DEFAULT ''
);

-- One row per transcript FILE, not per session: a session maps to many files across three tiers
-- (<sid>.jsonl, <sid>/subagents/agent-*.jsonl, and <sid>/subagents/workflows/<wf>/agent-*.jsonl),
-- and ingest resumes each independently. head_sha is the sha256 of the first 4 KiB: if it changes
-- the file was rewritten rather than appended, and the offset is meaningless.
CREATE TABLE IF NOT EXISTS file_offset (
    path      TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    agent_id  TEXT NOT NULL DEFAULT '',
    offset    INTEGER NOT NULL,
    size      INTEGER NOT NULL,
    head_sha  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS act (
    session_id TEXT NOT NULL,
    agent_id   TEXT NOT NULL DEFAULT '',
    seq        INTEGER NOT NULL,
    ts         INTEGER NOT NULL,
    tool       TEXT NOT NULL,
    target     TEXT NOT NULL DEFAULT '',
    -- ok | error | unresolved. NOT a boolean: 12 calls in the measured corpus had no
    -- tool_result at all, and forcing those into ok or error is a lie either way.
    outcome    TEXT NOT NULL,
    PRIMARY KEY (session_id, agent_id, seq)
);

CREATE TABLE IF NOT EXISTS word (
    session_id  TEXT NOT NULL,
    agent_id    TEXT NOT NULL DEFAULT '',
    prompt_id   TEXT,
    -- block_seq is the block's ordinal within its turn, NULL on a payload-sourced row until a
    -- transcript block gives it one. Identity is the four together; (session, prompt_id) is not
    -- unique because 66.2% of measured turns carry more than one block.
    block_seq   INTEGER,
    ts          INTEGER NOT NULL,
    role        TEXT NOT NULL,
    text        TEXT NOT NULL,
    source      TEXT NOT NULL,          -- payload | transcript
    provisional INTEGER NOT NULL DEFAULT 0,
    attribution TEXT                    -- NULL | unverified | ambiguous
);
CREATE INDEX IF NOT EXISTS word_key ON word (session_id, agent_id, prompt_id);

CREATE TABLE IF NOT EXISTS thought (
    session_id TEXT NOT NULL,
    agent_id   TEXT NOT NULL DEFAULT '',
    prompt_id  TEXT,
    block_seq  INTEGER,
    ts         INTEGER NOT NULL,
    text       TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS thought_key ON thought (session_id, agent_id, prompt_id);

-- A skipped turn is a fact a reader needs as much as a stored one: without this, a turn that
-- carried no prompt_id and a turn that was never seen produce the same silence.
CREATE TABLE IF NOT EXISTS provisional_skip (
    session_id TEXT NOT NULL,
    agent_id   TEXT NOT NULL DEFAULT '',
    prompt_id  TEXT,
    reason     TEXT NOT NULL,           -- no-prompt-id | no-last-assistant-message
    at         INTEGER NOT NULL
);

-- Internal bookkeeping, deliberately behind no view: it is the store's own state, not a fact
-- about any agent, and putting it on the published contract would invite queries against it.
CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);

CREATE INDEX IF NOT EXISTS act_target ON act (target);
CREATE INDEX IF NOT EXISTS act_ts     ON act (ts);

-- THE PUBLISHED CONTRACT. Callers query these, never the tables above.
CREATE VIEW IF NOT EXISTS v_session AS
    SELECT session_id, project_dir, cwd, first_seen, last_seen, closed_at, capture_build
    FROM session;
CREATE VIEW IF NOT EXISTS v_action AS
    SELECT session_id, agent_id, seq, ts, tool, target, outcome FROM act;
CREATE VIEW IF NOT EXISTS v_word AS
    SELECT session_id, agent_id, prompt_id, block_seq, ts, role, text, source, provisional,
           attribution
    FROM word;
CREATE VIEW IF NOT EXISTS v_thought AS
    SELECT session_id, agent_id, prompt_id, block_seq, ts, text FROM thought;
CREATE VIEW IF NOT EXISTS v_skip AS
    SELECT session_id, agent_id, prompt_id, reason, at FROM provisional_skip;
`

// UserVersion is the shape this binary writes. Open refuses a database written by a NEWER one:
// reading an unknown shape as though it were this one is how a store starts answering questions
// it cannot actually answer, and the store is derived, so refusing costs a reprojection and
// nothing else.
const UserVersion = 1

// ViewColumns is the contract §V.6 pins, restated here so a test can compare against a
// declaration rather than against the DDL it is testing.
var ViewColumns = map[string][]string{
	"v_session": {"session_id", "project_dir", "cwd", "first_seen", "last_seen", "closed_at", "capture_build"},
	"v_action":  {"session_id", "agent_id", "seq", "ts", "tool", "target", "outcome"},
	"v_word":    {"session_id", "agent_id", "prompt_id", "block_seq", "ts", "role", "text", "source", "provisional", "attribution"},
	"v_thought": {"session_id", "agent_id", "prompt_id", "block_seq", "ts", "text"},
	"v_skip":    {"session_id", "agent_id", "prompt_id", "reason", "at"},
}

// Outcome values for act.outcome.
const (
	OutcomeOK         = "ok"
	OutcomeError      = "error"
	OutcomeUnresolved = "unresolved"
)

// Skip reasons for provisional_skip.reason. Closed set: a counter downstream needs to separate
// the populations, and prose cannot be counted.
const (
	SkipNoPromptID    = "no-prompt-id"
	SkipNoLastAsstMsg = "no-last-assistant-message"
)
