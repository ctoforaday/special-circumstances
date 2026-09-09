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
	LastSeen  int64
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
		db.QueryRowContext(ctx, `SELECT last_seen FROM v_session WHERE session_id = ?`, sf.SessionID).Scan(&a.LastSeen)
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

// Touched answers "is anyone else editing this file". Matching is on a suffix so a caller can
// pass a repo-relative path without knowing the absolute one every agent used.
func Touched(ctx context.Context, db *sql.DB, path string) ([]Touch, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT session_id, agent_id, tool, outcome, max(ts) AS ts, count(*) AS n
		FROM v_action
		WHERE target = ? OR target LIKE ?
		GROUP BY session_id, agent_id, tool, outcome
		ORDER BY ts DESC`, path, "%"+path)
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
}

// Shape summarises a single session.
func Shape(ctx context.Context, db *sql.DB, sessionID string) (SessionShape, error) {
	s := SessionShape{SessionID: sessionID, ByTool: map[string]int{}, Failures: map[string]int{}}
	if err := db.QueryRowContext(ctx,
		`SELECT project_dir, closed_at FROM v_session WHERE session_id = ?`, sessionID).
		Scan(&s.ProjectDir, &s.ClosedAt); err != nil {
		return s, fmt.Errorf("catalogue: no session %s in the store — it may predate capture, "+
			"which is not the same as a session that did nothing", sessionID)
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
