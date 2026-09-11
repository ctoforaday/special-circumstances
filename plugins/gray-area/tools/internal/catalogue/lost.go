package catalogue

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LostSession is one session a restart cut off.
//
// IT IS A CANDIDATE FOR A HUMAN, INFERRED, NEVER MEASURED. A session that exited cleanly after the
// last closure sweep, or one closure's cap deferred, reads exactly like one the restart killed; the
// list is best-effort by design (plans/restart-recovery.md), and the resume attempt is the check —
// `claude --resume` from the wrong directory, or a reattach to an expired environment, fails loudly.
type LostSession struct {
	SessionID string
	// CWD is where to resume from, and Verified says its folder key is the transcript's folder
	// name. An unverified value is a hint — a session that moved folders, a hashed long path — and
	// '' means no record carried one.
	CWD        string
	Verified   bool
	ProjectDir string
	// BridgeSessionID is the captured `session_<suffix>`; '' when capture never saw one.
	BridgeSessionID string
	// LastAt is the session's last sign of life in TRANSCRIPT time. Registered marks the fallback
	// for a session registered with no rows yet, whose LastAt is when it registered.
	LastAt     int64
	Registered bool
	Acts       int
	LastTool   string
	LastToolAt int64
	// Transcript is the session's own transcript, which PermissionMode reads; '' when none is known.
	Transcript string
}

// LostReport is FindLost's answer, with what a reader needs to judge it.
type LostReport struct {
	Boot int64
	// NewestPreBoot is the largest last sign of life below the boot across EVERY session, and
	// NewestSession is whose it is: 0 and '' when nothing in the store precedes the boot.
	NewestPreBoot int64
	NewestSession string
	// Sessions counts every session the store holds, so "nothing precedes the boot" and "the store
	// is empty" are two sentences rather than one silence.
	Sessions int
	Lost     []LostSession
}

// FindLost selects the sessions a restart at `boot` cut off.
//
// EVERY TIME HERE IS TRANSCRIPT TIME. A session's last sign of life is the newest ts across its
// act, word and thought rows — when it DID something — and never ingested_last, which is when the
// catalogue READ it: a backfill, or the rebuild an upgrade triggers, moves that past the boot for
// every session, and a rule over it would list nothing. A session registered with no rows yet falls
// back to ingested_first, which RegisterSession set while it ran.
//
// A session is lost when its last sign of life precedes the boot, lies within `window` of the
// newest pre-boot activity on the box, was not closed by the catalogue between that activity and
// the boot, and is not in `exclude` — the caller's MEASURED `live` sessions, and the `unknown` ones
// that are not ours to judge. A close AFTER the boot does not count: the first SessionStart after a
// restart closes every session no longer advertised, which is exactly the population this lists.
func FindLost(ctx context.Context, db *sql.DB, boot int64, window time.Duration, exclude map[string]bool) (LostReport, error) {
	rep := LostReport{Boot: boot}
	// ts <= 0 is a record whose timestamp did not parse (parseTS), not a sign of life at the epoch.
	rows, err := db.QueryContext(ctx, `
		WITH life AS (
		    SELECT session_id, max(ts) AS last FROM (
		        SELECT session_id, ts FROM act     WHERE ts > 0 UNION ALL
		        SELECT session_id, ts FROM word    WHERE ts > 0 UNION ALL
		        SELECT session_id, ts FROM thought WHERE ts > 0)
		    GROUP BY session_id)
		SELECT s.session_id, s.cwd, s.project_dir, s.bridge_session_id, s.closed_at, s.ingested_first, life.last
		FROM session s LEFT JOIN life ON life.session_id = s.session_id
		ORDER BY s.session_id`)
	if err != nil {
		return rep, fmt.Errorf("catalogue: lost sessions: %w", err)
	}
	type candidate struct {
		LostSession
		closed sql.NullInt64
	}
	var all []candidate
	for rows.Next() {
		var c candidate
		var first int64
		var last sql.NullInt64
		if err := rows.Scan(&c.SessionID, &c.CWD, &c.ProjectDir, &c.BridgeSessionID, &c.closed, &first, &last); err != nil {
			rows.Close()
			return rep, fmt.Errorf("catalogue: lost sessions: %w", err)
		}
		rep.Sessions++
		if last.Valid {
			c.LastAt = last.Int64
		} else {
			c.LastAt, c.Registered = first, true
		}
		// Ties go to the smaller id, so the session the report names does not depend on scan order.
		if c.LastAt < boot && (rep.NewestSession == "" || c.LastAt > rep.NewestPreBoot) {
			rep.NewestPreBoot, rep.NewestSession = c.LastAt, c.SessionID
		}
		all = append(all, c)
	}
	if err := rows.Close(); err != nil {
		return rep, err
	}
	if rep.NewestSession == "" {
		return rep, nil // nothing precedes the boot, so nothing was running when it came
	}
	floor := rep.NewestPreBoot - int64(window/time.Second)
	for _, c := range all {
		switch {
		case c.LastAt >= boot, c.LastAt < floor, exclude[c.SessionID]:
			continue
		case c.closed.Valid && c.closed.Int64 >= c.LastAt && c.closed.Int64 < boot:
			continue // closure settled it after its last act and before the boot: it ended first
		}
		l := c.LostSession
		l.Verified = l.CWD != "" && l.ProjectDir != "" && projectKey(l.CWD) == filepath.Base(l.ProjectDir)
		db.QueryRowContext(ctx, `SELECT count(*) FROM v_action WHERE session_id = ?`, l.SessionID).Scan(&l.Acts)
		db.QueryRowContext(ctx,
			`SELECT tool, ts FROM v_action WHERE session_id = ? ORDER BY ts DESC, seq DESC LIMIT 1`,
			l.SessionID).Scan(&l.LastTool, &l.LastToolAt)
		l.Transcript = ownTranscript(ctx, db, l.SessionID, l.ProjectDir)
		rep.Lost = append(rep.Lost, l)
	}
	sort.SliceStable(rep.Lost, func(i, j int) bool {
		if rep.Lost[i].LastAt != rep.Lost[j].LastAt {
			return rep.Lost[i].LastAt > rep.Lost[j].LastAt
		}
		return rep.Lost[i].SessionID < rep.Lost[j].SessionID
	})
	return rep, nil
}

// ownTranscript is the session's own transcript: the file the store read it from (the newest by
// modification time if a moved session left two), else `<project_dir>/<session>.jsonl` for a
// session registered with no rows, else ”. No new column: the store already holds the path.
func ownTranscript(ctx context.Context, db *sql.DB, sessionID, projectDir string) string {
	best, bestAt := "", time.Time{}
	if rows, err := db.QueryContext(ctx,
		`SELECT path FROM file_offset WHERE session_id = ? AND agent_id = '' ORDER BY path`, sessionID); err == nil {
		for rows.Next() {
			var p string
			if rows.Scan(&p) != nil {
				continue
			}
			st, err := os.Stat(p)
			if err != nil {
				continue // a path the store read and the client has since removed
			}
			if best == "" || st.ModTime().After(bestAt) {
				best, bestAt = p, st.ModTime()
			}
		}
		rows.Close()
	}
	if best == "" && projectDir != "" {
		best = filepath.Join(projectDir, sessionID+".jsonl")
	}
	return best
}

// PermissionMode is a session's last RECORDED permission mode — the `permissionMode` of the newest
// `user` record THAT CARRIES THE FIELD, read from its own transcript in file order. ” means
// unknown: no file, or no record carrying it.
//
// THE NEWEST `user` RECORD IS USUALLY THE WRONG ONE. Tool results are `user` records too and none
// carries the field (11,141 measured, 0 with it), so the newest `user` record of a session killed
// mid-tool-loop says nothing; the mode lives on the prompt before it. Session files carry no mode
// at all, which is why this reads the transcript.
func PermissionMode(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	// The initial buffer is capped too: bufio takes the LARGER of it and max as the real limit.
	sc.Buffer(make([]byte, 0, min(256*1024, permissionModeLineCap)), permissionModeLineCap)
	key := []byte(`"permissionMode"`)
	mode := ""
	for sc.Scan() {
		line := sc.Bytes()
		if !bytes.Contains(line, key) {
			continue // the decode below is the cost; most lines cannot carry the field
		}
		var r struct {
			Type           string  `json:"type"`
			PermissionMode *string `json:"permissionMode"`
		}
		if json.Unmarshal(line, &r) != nil || r.Type != "user" || r.PermissionMode == nil || *r.PermissionMode == "" {
			continue
		}
		mode = *r.PermissionMode
	}
	// A READ THAT STOPPED EARLY IS NOT A READ OF THE WHOLE FILE. A line past the cap (or an I/O
	// error) ends Scan silently, and the mode found before it may not be the newest — printing it
	// would resume the session in a state it had left. Unknown is the honest answer.
	if sc.Err() != nil {
		return ""
	}
	return mode
}

// permissionModeLineCap bounds one transcript line in PermissionMode; a var so a test can make a
// line exceed it without writing 64 MB.
var permissionModeLineCap = 64 * 1024 * 1024

// CloudID turns a captured `session_<suffix>` into the `cse_<suffix>` the reattach flag takes, and
// anything else into ” — which the rows print as unknown rather than guess a shape.
func CloudID(bridge string) string {
	if s, ok := strings.CutPrefix(bridge, "session_"); ok && s != "" {
		return "cse_" + s
	}
	return ""
}

// ModeUnknownLine is the MODE row of a session with no recorded mode. It NAMES NO MODE: what a
// resume without the flag starts in is the CLI's choice, not a fact this store holds.
const ModeUnknownLine = "MODE unknown — the resume starts in whatever mode the CLI chooses"

// ModeLine renders the MODE row for a recorded mode, ” being unknown.
func ModeLine(mode string) string {
	if mode == "" {
		return ModeUnknownLine
	}
	return "MODE " + mode
}

// RecoveryCommands renders the commands that bring a lost session back, ready to paste. reattach
// is ” when no cloud id is known.
//
// A RECOVERED SESSION RESUMES IN THE STATE IT BEGAN IN (gblock, 2026-09-11). The REATTACH carries
// no mode flag because it needs none: a server-respawned session kept its own recorded `auto`
// under a `--permission-mode default` on its child's command line (measured 2026-09-11). The LOCAL
// resume is GIVEN the mode, which was measured to override the CLI's own default — and is given
// none when the mode is unknown. A recorded `default` is written `manual`, the spelling
// `claude --help` lists; `claude --resume` accepts both today.
func RecoveryCommands(sessionID, cloud, mode string) (reattach, local string) {
	if cloud != "" {
		reattach = "claude remote-control --session-id " + cloud
	}
	local = "claude --resume " + sessionID
	if mode != "" {
		if mode == "default" {
			mode = "manual"
		}
		local += " --permission-mode " + mode
	}
	return reattach, local + " --remote-control"
}
