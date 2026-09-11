package catalogue

import (
	"database/sql"
	"fmt"
	"time"
)

// RegisterSession records, at SessionStart and at Stop, that a session is RUNNING — whether or not
// any of its transcript has been read yet.
//
// # Why a hook registers rather than leaving it to ingest
//
// Ingest writes a session row only once a transcript has bytes to project, and closure marks a
// session shut once it stops being advertised. Neither records which sessions were running when
// the host went down, which is the question `telepathy agents --lost` answers after a restart.
// Registration gives every started session a row (its ingested_first is then when it registered,
// while it ran), reopens a session closure had already settled — closed_at = NULL, because it is
// running again — and records the one fact no transcript reliably holds: the Remote Control cloud
// id from the client's own session file.
//
// A KNOWN CLOUD ID IS NEVER OVERWRITTEN WITH EMPTY. The session file is gone by SessionEnd, and a
// Stop whose read of it failed must not erase what SessionStart captured. project_dir follows the
// same rule and ingest's: an empty value never replaces a known one.
//
// ingested_last is set on insert and left alone after: it means "when the catalogue read this
// session", and a registration reads nothing.
func RegisterSession(db *sql.DB, sessionID, projectDir, bridgeID string, now time.Time) error {
	if sessionID == "" {
		return fmt.Errorf("catalogue: refusing to register a session with no id")
	}
	ts := now.Unix()
	if _, err := db.Exec(`
		INSERT INTO session(session_id, project_dir, ingested_first, ingested_last, bridge_session_id)
		VALUES(?,?,?,?,?)
		ON CONFLICT(session_id) DO UPDATE SET
		    closed_at = NULL,
		    project_dir = CASE WHEN excluded.project_dir != '' THEN excluded.project_dir
		                       ELSE session.project_dir END,
		    bridge_session_id = CASE WHEN excluded.bridge_session_id != '' THEN excluded.bridge_session_id
		                             ELSE session.bridge_session_id END`,
		sessionID, projectDir, ts, ts, bridgeID); err != nil {
		return fmt.Errorf("catalogue: registering %s: %w", sessionID, err)
	}
	return nil
}
