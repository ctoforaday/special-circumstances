package catalogue

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TranscriptFile is one file belonging to one session, with the agent it speaks for.
type TranscriptFile struct {
	Path      string
	SessionID string
	AgentID   string // "" for the session's own transcript
}

// TranscriptFiles enumerates a project tree by the vendor's layout, which has THREE tiers:
//
//	<project>/<sid>.jsonl                                     the session's own
//	<project>/<sid>/subagents/agent-<id>.jsonl                a seat
//	<project>/<sid>/subagents/workflows/<wf>/agent-<id>.jsonl a seat inside a workflow
//
// # The third tier is the one that bites
//
// A rule naming only the first two attributes 189 files here and leaves 56 UNATTRIBUTED — every
// one of them presently under a live session. When that session ends they join the non-live
// population, an implementation built to the two-tier rule never enumerates them, their turns go
// uncatalogued, and the pending queue reads EMPTY. A plausible zero produced by a rule that
// disagreed with the `**` glob used to validate it, which is why this is a rule with a test
// rather than a glob.
func TranscriptFiles(projectsRoot string) ([]TranscriptFile, error) {
	var out []TranscriptFile
	err := filepath.WalkDir(projectsRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".jsonl") {
			return nil //nolint:nilerr // an unreadable subtree is not a reason to abandon the rest
		}
		rel, rerr := filepath.Rel(projectsRoot, p)
		if rerr != nil {
			return nil
		}
		parts := strings.Split(rel, string(os.PathSeparator))
		switch {
		case len(parts) == 2:
			out = append(out, TranscriptFile{Path: p, SessionID: strings.TrimSuffix(parts[1], ".jsonl")})
		case len(parts) == 4 && parts[2] == "subagents":
			out = append(out, TranscriptFile{Path: p, SessionID: parts[1], AgentID: agentIDOf(parts[3])})
		case len(parts) == 6 && parts[2] == "subagents" && parts[3] == "workflows":
			out = append(out, TranscriptFile{Path: p, SessionID: parts[1], AgentID: agentIDOf(parts[5])})
		}
		return nil
	})
	return out, err
}

func agentIDOf(base string) string {
	return strings.TrimSuffix(strings.TrimPrefix(base, "agent-"), ".jsonl")
}

// headSHA fingerprints a FIXED-LENGTH prefix — the first min(4096, n) bytes. If it changes, the
// file was rewritten rather than appended and any stored offset points into the middle of a
// record, so the file is reprojected from zero instead of resumed into nonsense.
//
// THE LENGTH MUST BE PINNED TO THE PREFIX ALREADY CONSUMED, not to the current size. Hashing
// "the first 4 KiB or the whole file, whichever is smaller" makes the fingerprint of any file
// under 4 KiB change on every APPEND — the fingerprint grows with the file — so every small
// transcript reprojects from zero on every pass. Caught by TestAppendIsResumedNotReread, which
// is why it asserts the byte count rather than only the row count.
func headSHA(f *os.File, n int64) (string, error) {
	if n > 4096 {
		n = 4096
	}
	if n <= 0 {
		return "", nil
	}
	buf := make([]byte, n)
	got, err := f.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return "", err
	}
	sum := sha256.Sum256(buf[:got])
	return hex.EncodeToString(sum[:]), nil
}

// IngestResult reports what one file's pass did.
type IngestResult struct {
	Acts, Words, Thoughts int
	// Skipped counts records seen and not stored — today, thinking blocks that carried no text.
	// Reported alongside the rest because a projection that drops the majority of one tier in
	// silence is indistinguishable from a corpus that never had it.
	Skipped     int
	Unparsed    int
	Reprojected bool // the head changed, so the file was read from zero
	BytesRead   int64
}

// IngestFile brings one transcript file up to date, resuming from its stored offset.
//
// Everything it writes is derived from the transcript, so a re-run is idempotent by construction:
// acts key on (session, agent, seq), and words and thoughts on their turn ordinal. A second pass
// over the same bytes produces the same rows.
func IngestFile(db *sql.DB, tf TranscriptFile) (IngestResult, error) {
	var res IngestResult
	f, err := os.Open(tf.Path)
	if err != nil {
		return res, nil // a transcript that vanished mid-sweep is data, not a failure
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return res, nil
	}
	var offset, size int64
	var storedHead string
	row := db.QueryRow(`SELECT offset, size, head_sha FROM file_offset WHERE path = ?`, tf.Path)
	switch err := row.Scan(&offset, &size, &storedHead); {
	case err == sql.ErrNoRows:
		offset, storedHead = 0, ""
	case err != nil:
		return res, fmt.Errorf("catalogue: reading offset for %s: %w", tf.Path, err)
	}
	// Compare over the prefix we already consumed, so an append cannot move the fingerprint.
	if offset > 0 {
		cur, err := headSHA(f, offset)
		if err != nil {
			return res, fmt.Errorf("catalogue: fingerprinting %s: %w", tf.Path, err)
		}
		if cur != storedHead || st.Size() < offset {
			offset, res.Reprojected = 0, true
		}
	}
	// The fingerprint STORED covers the prefix this pass will have consumed.
	head, err := headSHA(f, st.Size())
	if err != nil {
		return res, fmt.Errorf("catalogue: fingerprinting %s: %w", tf.Path, err)
	}
	if st.Size() == offset {
		return res, nil
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return res, fmt.Errorf("catalogue: seeking %s: %w", tf.Path, err)
	}
	res.BytesRead = st.Size() - offset

	// seq continues from what this file already contributed, so acts stay unique per agent.
	var startSeq int
	db.QueryRow(`SELECT COALESCE(MAX(seq)+1,0) FROM act WHERE session_id=? AND agent_id=?`,
		tf.SessionID, tf.AgentID).Scan(&startSeq)

	p := Project(f, startSeq)
	res.Acts, res.Words, res.Thoughts, res.Unparsed = len(p.Acts), len(p.Words), len(p.Thoughts), p.Unparsed
	res.Skipped = len(p.Skips)

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("catalogue: begin: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // a committed tx makes this a no-op

	if res.Reprojected {
		for _, q := range []string{
			`DELETE FROM act WHERE session_id=? AND agent_id=?`,
			`DELETE FROM word WHERE session_id=? AND agent_id=? AND source='transcript'`,
			`DELETE FROM thought WHERE session_id=? AND agent_id=?`,
			`DELETE FROM skip WHERE session_id=? AND agent_id=?`,
		} {
			if _, err := tx.Exec(q, tf.SessionID, tf.AgentID); err != nil {
				return res, fmt.Errorf("catalogue: clearing for reprojection: %w", err)
			}
		}
	}
	for _, a := range p.Acts {
		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO act(session_id,agent_id,seq,ts,tool,target,outcome) VALUES(?,?,?,?,?,?,?)`,
			tf.SessionID, tf.AgentID, a.Seq, a.TS, a.Tool, a.Target, a.Outcome); err != nil {
			return res, fmt.Errorf("catalogue: act: %w", err)
		}
	}
	for _, w := range p.Words {
		if _, err := tx.Exec(
			`INSERT INTO word(session_id,agent_id,prompt_id,block_seq,ts,role,text,source,provisional)
			 VALUES(?,?,?,?,?,?,?,'transcript',0)`,
			tf.SessionID, tf.AgentID, nullable(w.PromptID), w.BlockSeq, w.TS, w.Role, w.Text); err != nil {
			return res, fmt.Errorf("catalogue: word: %w", err)
		}
	}
	for _, th := range p.Thoughts {
		if _, err := tx.Exec(
			`INSERT INTO thought(session_id,agent_id,prompt_id,block_seq,ts,text) VALUES(?,?,?,?,?,?)`,
			tf.SessionID, tf.AgentID, nullable(th.PromptID), th.BlockSeq, th.TS, th.Text); err != nil {
			return res, fmt.Errorf("catalogue: thought: %w", err)
		}
	}
	for _, sk := range p.Skips {
		if _, err := tx.Exec(
			`INSERT INTO skip(session_id,agent_id,prompt_id,reason,at) VALUES(?,?,?,?,?)`,
			tf.SessionID, tf.AgentID, nullable(sk.PromptID), sk.Reason, sk.At); err != nil {
			return res, fmt.Errorf("catalogue: skip: %w", err)
		}
	}
	if _, err := tx.Exec(
		`INSERT INTO file_offset(path,session_id,agent_id,offset,size,head_sha) VALUES(?,?,?,?,?,?)
		 ON CONFLICT(path) DO UPDATE SET offset=excluded.offset, size=excluded.size, head_sha=excluded.head_sha`,
		tf.Path, tf.SessionID, tf.AgentID, st.Size(), st.Size(), head); err != nil {
		return res, fmt.Errorf("catalogue: offset: %w", err)
	}
	now := time.Now().Unix()
	// PROJECT DIR COMES FROM THE SESSION'S OWN TRANSCRIPT, never from whichever file happened to
	// be ingested last. A session with subagents has most of its files under <sid>/subagents/, so
	// taking the last one made the row report the subagents directory as the project — a field
	// that was simply false, and false in a way that reads as plausible.
	dir := ""
	if tf.AgentID == "" {
		dir = filepath.Dir(tf.Path)
	}
	if _, err := tx.Exec(
		`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES(?,?,?,?)
		 ON CONFLICT(session_id) DO UPDATE SET ingested_last=excluded.ingested_last,
		     project_dir=CASE WHEN excluded.project_dir != '' THEN excluded.project_dir ELSE session.project_dir END`,
		tf.SessionID, dir, now, now); err != nil {
		return res, fmt.Errorf("catalogue: session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return res, fmt.Errorf("catalogue: commit: %w", err)
	}
	return res, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
