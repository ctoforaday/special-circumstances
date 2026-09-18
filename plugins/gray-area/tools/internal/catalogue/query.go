package catalogue

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Agent is one row of the switchboard: which agents are active, and what each is working on.
type Agent struct {
	SessionID string
	CWD       string
	Liveness  Liveness
	Acts      int
	LastTool  string
	LastAt    int64
}

// Agents answers the switchboard's first two questions from the store plus the client's own
// session files. Liveness comes from the files (it cannot be derived from a transcript); identity
// and activity come from the store.
func Agents(ctx context.Context, db *sql.DB, sessionDir string) ([]Agent, error) {
	files, err := ReadSessionFiles(sessionDir)
	if err != nil {
		return nil, err
	}
	domain := LocalPidDomain()
	out := make([]Agent, 0, len(files))
	for _, sf := range files {
		a := Agent{SessionID: sf.SessionID, CWD: sf.CWD, Liveness: Of(sf, domain, ProcStart)}
		db.QueryRowContext(ctx, `SELECT count(*) FROM v_action WHERE session_id = ?`, sf.SessionID).Scan(&a.Acts)
		db.QueryRowContext(ctx,
			`SELECT tool, ts FROM v_action WHERE session_id = ? ORDER BY ts DESC, seq DESC LIMIT 1`,
			sf.SessionID).Scan(&a.LastTool, &a.LastAt)
		out = append(out, a)
	}
	return out, nil
}

// Touch is one session's contact with a path.
type Touch struct {
	SessionID string
	AgentID   string
	Tool      string
	Outcome   string
	TS        int64
	N         int
}

// Touched answers "is anyone else editing this file".
//
// Matching is on a SUBSTRING, not a suffix, and the difference is the whole usefulness of the
// verb. A tool target is `file_path` for Read/Edit/Write — where a suffix match is exactly right —
// but for Bash it is the COMMAND, and an edit made with `sed -i … path` or a python heredoc names
// the path in the middle of that string. Under a suffix match every such edit reported nothing,
// which is the plausible zero this plugin exists to refuse: a caller asking "is anyone in this
// file" got silence from a shell rewriting it.
//
// The residue is stated rather than engineered around: targets are truncated at 200 characters, so
// a path buried past that in a long command is still invisible here. `find` is the fallback that
// reads the transcripts themselves.
func Touched(ctx context.Context, db *sql.DB, path string) ([]Touch, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT session_id, agent_id, tool, outcome, max(ts) AS ts, count(*) AS n
		FROM v_action
		WHERE target = ? OR instr(target, ?) > 0
		GROUP BY session_id, agent_id, tool, outcome
		ORDER BY ts DESC, session_id, agent_id, tool`, path, path)
	if err != nil {
		return nil, fmt.Errorf("catalogue: touched: %w", err)
	}
	defer rows.Close()
	var out []Touch
	for rows.Next() {
		var t Touch
		if err := rows.Scan(&t.SessionID, &t.AgentID, &t.Tool, &t.Outcome, &t.TS, &t.N); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// SessionShape is what one session did, by tool.
type SessionShape struct {
	SessionID  string
	ProjectDir string
	ClosedAt   sql.NullInt64
	ByTool     map[string]int
	Failures   map[string]int
	Words      int
	Thoughts   int
	// Skipped counts, by reason, what the projection SAW and did not store. A session showing 0
	// thoughts and 300 thinking-empty skips is a different fact from one showing 0 and 0, and the
	// verb that prints them has to be able to tell a reader which it is looking at.
	Skipped map[string]int
}

// ResolveSession turns what a reader typed into the one session id it names: the full id, or any
// prefix of exactly one — Short's eight characters being the form every verb here prints.
//
// A FULL-ID-ONLY LOOKUP REFUSED THE IDS THIS TOOL HANDS OUT. `agents` and `find` print Short, and
// `session <that>` answered "no session in the store — it may predate capture" about a session the
// store held: a worded absence naming the wrong cause, which sent a reader off to backfill and
// then to conclude the session was gone. A prefix naming several is refused with the candidates,
// never resolved by guessing one.
func ResolveSession(ctx context.Context, db *sql.DB, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("catalogue: an empty session id names nothing")
	}
	rows, err := db.QueryContext(ctx,
		`SELECT session_id FROM v_session WHERE substr(session_id, 1, length(?1)) = ?1 ORDER BY session_id`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return "", err
		}
		if sid == id {
			return sid, nil
		}
		ids = append(ids, sid)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	switch len(ids) {
	case 0:
		return "", fmt.Errorf("catalogue: no session %s in the store, by full id or prefix — it may "+
			"predate capture, which is not the same as a session that did nothing", id)
	case 1:
		return ids[0], nil
	}
	return "", fmt.Errorf("catalogue: %q is a prefix of %d sessions (%s) — give more of the id",
		id, len(ids), strings.Join(ids, ", "))
}

// Shape summarises a single session, named by its full id or a unique prefix (ResolveSession).
func Shape(ctx context.Context, db *sql.DB, id string) (SessionShape, error) {
	sessionID, err := ResolveSession(ctx, db, id)
	if err != nil {
		return SessionShape{SessionID: id}, err
	}
	s := SessionShape{SessionID: sessionID, ByTool: map[string]int{}, Failures: map[string]int{}, Skipped: map[string]int{}}
	if err := db.QueryRowContext(ctx,
		`SELECT project_dir, closed_at FROM v_session WHERE session_id = ?`, sessionID).
		Scan(&s.ProjectDir, &s.ClosedAt); err != nil {
		return s, err
	}
	rows, err := db.QueryContext(ctx,
		`SELECT tool, outcome, count(*) FROM v_action WHERE session_id = ? GROUP BY tool, outcome`, sessionID)
	if err != nil {
		return s, err
	}
	for rows.Next() {
		var tool, outcome string
		var n int
		if err := rows.Scan(&tool, &outcome, &n); err != nil {
			rows.Close()
			return s, err
		}
		s.ByTool[tool] += n
		if outcome == OutcomeError {
			s.Failures[tool] += n
		}
	}
	rows.Close()
	db.QueryRowContext(ctx, `SELECT count(*) FROM v_word WHERE session_id = ?`, sessionID).Scan(&s.Words)
	db.QueryRowContext(ctx, `SELECT count(*) FROM v_thought WHERE session_id = ?`, sessionID).Scan(&s.Thoughts)
	if srows, err := db.QueryContext(ctx,
		`SELECT reason, count(*) FROM v_skip WHERE session_id = ? GROUP BY reason`, sessionID); err == nil {
		for srows.Next() {
			var reason string
			var n int
			if srows.Scan(&reason, &n) == nil {
				s.Skipped[reason] += n
			}
		}
		srows.Close()
	}
	return s, nil
}

// RawQuery runs a caller's SQL on a connection carrying the ATTACH limit.
//
// It returns column names and rows as strings: the surface is "run this SELECT and show me",
// and typing it further would be inventing a schema the caller already has in the views.
func RawQuery(ctx context.Context, db *sql.DB, query string, limit int) ([]string, [][]string, error) {
	c, err := PinnedConn(ctx, db)
	if err != nil {
		return nil, nil, err
	}
	defer c.Close()

	rows, err := c.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("catalogue: %w", err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	var out [][]string
	for rows.Next() {
		if limit > 0 && len(out) >= limit {
			break
		}
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		rec := make([]string, len(cols))
		for i, v := range vals {
			rec[i] = render(v)
		}
		out = append(out, rec)
	}
	return cols, out, rows.Err()
}

// render turns a driver value into display text. NULL is rendered as the word, not as an empty
// string: "" and NULL are different answers and a reader must be able to tell them apart.
func render(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

// Grep is the full-text half: ripgrep over the transcript files the catalogue names, joined back
// to who and when. No index is built — a whole-corpus search measured 33–54 ms, against an
// estimated 1–2 GB for an FTS index that would need invalidating on every append.
type Grep struct {
	SessionID string
	AgentID   string
	Path      string
	Hits      int
}

// PathsFor lists the transcript files the store knows for a session, for `find` to search.
func PathsFor(ctx context.Context, db *sql.DB) (map[string]TranscriptFile, error) {
	rows, err := db.QueryContext(ctx, `SELECT path, session_id, agent_id FROM file_offset`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]TranscriptFile{}
	for rows.Next() {
		var f TranscriptFile
		if err := rows.Scan(&f.Path, &f.SessionID, &f.AgentID); err != nil {
			return nil, err
		}
		out[f.Path] = f
	}
	return out, rows.Err()
}

// Short renders a session id at the length humans use in this repository.
func Short(id string) string {
	if i := strings.IndexByte(id, '-'); i > 0 {
		return id[:i]
	}
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
