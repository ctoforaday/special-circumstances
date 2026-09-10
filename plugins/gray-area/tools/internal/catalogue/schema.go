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
// DERIVED: nothing here is authoritative, so a shape change is answered by REBUILDING rather than
// migrating. `PRAGMA user_version` records which shape wrote it, and Open acts on what it finds:
//
//   - an OLDER catalogue is rebuilt — every table, view and index dropped, this DDL applied,
//     meta.rebuilt_at/rebuilt_from written — and must then be backfilled; every read verb warns
//     until it has been (RebuildPending);
//   - a file that is NOT RECOGNISED as a catalogue is never written, by Open or by OpenRead;
//   - a NEWER stamp is refused, never read as this shape.
//
// `CREATE TABLE IF NOT EXISTS` cannot carry a shape change on its own: it skips an existing table,
// which is how #867's column rename landed on disk as a stamp 2 over shape-1 columns.
//
// THE RECOGNITION CONTRACT (recognise, open.go). Every future shape MUST keep tables named
// `session` and `file_offset`: a stamp newer than this binary's is recognised as a catalogue by
// those two names ALONE, because a future shape may add tables or rename `session`'s columns, and
// a newer catalogue mistaken for a foreign file would be refused with the wrong words during every
// version skew. A stamp of UserVersion or below is recognised by `session`'s columns including
// session_id, project_dir, cwd, closed_at and capture_build — the five every shape has carried —
// and by every table name being one a catalogue has used.
const Schema = `
CREATE TABLE IF NOT EXISTS session (
    session_id    TEXT PRIMARY KEY,
    project_dir   TEXT NOT NULL,
    cwd           TEXT NOT NULL DEFAULT '',
    -- WHEN THE CATALOGUE SAW THIS SESSION, which is not when the session ran. These were called
    -- first_seen/last_seen and read as the session's lifetime; every session backfilled in one
    -- pass carries the same pair of timestamps, and the author of this store misread his own
    -- column that way inside a week. The session's actual span is first_act/last_act on the view,
    -- derived from the acts.
    ingested_first INTEGER NOT NULL,
    ingested_last  INTEGER NOT NULL,
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

-- WHAT WE SAW AND DID NOT STORE. Without this, a record the projection dropped and a record that
-- never existed produce the same silence — and the biggest such population is invisible without
-- it: 911 thinking blocks in one measured transcript, 79 ever carrying text, none since
-- 2026-09-08. v_thought reports 0 for that session, identically to a session that emitted no
-- thinking blocks at all. One is the client withholding reasoning; the other is an agent that did
-- not reason, and gray-area exists to refuse exactly that conflation.
--
-- The table was provisional_skip, serving the payload-recovery machinery that was ruled out
-- before it was built, and nothing ever wrote a row. Dropped rather than left on the contract as a
-- view that returns zero for all time.
DROP VIEW IF EXISTS v_skip;
DROP TABLE IF EXISTS provisional_skip;
CREATE TABLE IF NOT EXISTS skip (
    session_id TEXT NOT NULL,
    agent_id   TEXT NOT NULL DEFAULT '',
    prompt_id  TEXT,
    reason     TEXT NOT NULL,           -- see the Skip* constants; a closed set, so it counts
    at         INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS skip_key ON skip (session_id, agent_id);

-- Internal bookkeeping, deliberately behind no view: it is the store's own state, not a fact
-- about any agent, and putting it on the published contract would invite queries against it.
CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);

CREATE INDEX IF NOT EXISTS act_target ON act (target);
CREATE INDEX IF NOT EXISTS act_ts     ON act (ts);
-- first_act/last_act on v_session are min/max per session; without this they are a scan per row.
CREATE INDEX IF NOT EXISTS act_session_ts ON act (session_id, ts);

-- THE PUBLISHED CONTRACT. Callers query these, never the tables above.
-- first_act/last_act are NULL for a session with no acts — a real state (a session that only
-- spoke), and distinct from a session whose acts are simply old.
DROP VIEW IF EXISTS v_session;
CREATE VIEW IF NOT EXISTS v_session AS
    SELECT s.session_id, s.project_dir, s.cwd,
           (SELECT min(ts) FROM act WHERE act.session_id = s.session_id) AS first_act,
           (SELECT max(ts) FROM act WHERE act.session_id = s.session_id) AS last_act,
           s.ingested_first, s.ingested_last, s.closed_at, s.capture_build
    FROM session s;
CREATE VIEW IF NOT EXISTS v_action AS
    SELECT session_id, agent_id, seq, ts, tool, target, outcome FROM act;
CREATE VIEW IF NOT EXISTS v_word AS
    SELECT session_id, agent_id, prompt_id, block_seq, ts, role, text, source, provisional,
           attribution
    FROM word;
CREATE VIEW IF NOT EXISTS v_thought AS
    SELECT session_id, agent_id, prompt_id, block_seq, ts, text FROM thought;
CREATE VIEW IF NOT EXISTS v_skip AS
    SELECT session_id, agent_id, prompt_id, reason, at FROM skip;
`

// UserVersion is the shape this binary writes. Open refuses a database written by a NEWER one:
// reading an unknown shape as though it were this one is how a store starts answering questions
// it cannot actually answer. An OLDER one is rebuilt.
//
// WHY 3, WITH NO DDL CHANGE SINCE 2. The stamp 2 cannot be trusted: before the rebuild existed,
// Open stamped the current version onto any older store after a `CREATE TABLE IF NOT EXISTS` that
// skipped its existing tables, so every store created before #867 carries a 2 over shape-1
// columns (first_seen/last_seen), and its v_session fails on the first query. Bumping past it is
// what lets every such store be recognised as older and rebuilt once.
const UserVersion = 3

// ViewColumns is the contract §V.6 pins, restated here so a test can compare against a
// declaration rather than against the DDL it is testing.
var ViewColumns = map[string][]string{
	"v_session": {"session_id", "project_dir", "cwd", "first_act", "last_act", "ingested_first", "ingested_last", "closed_at", "capture_build"},
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

// Skip reasons. A CLOSED SET, and deliberately only the reasons something actually writes: a
// constant nothing can produce is a population that always counts zero, which reads as a clean
// board forever. The two that used to live here — no-prompt-id, no-last-assistant-message —
// belonged to machinery that was ruled out before it was built.
const (
	// SkipThinkingEmpty: the record carried a thinking block whose text was empty. The client
	// emits these with a signature and no content; they are the difference between "this agent
	// did not reason" and "its reasoning was withheld from the transcript".
	SkipThinkingEmpty = "thinking-empty"
)
