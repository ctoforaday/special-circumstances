package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// storePath resolves the catalogue's location once, so every verb reports the same path when it
// is absent.
func storePath() (string, error) {
	dir, err := catalogue.DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "catalogue.db"), nil
}

func projectsRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

func sessionsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "sessions")
}

// agentsVerb answers the switchboard's first two questions: which agents are active, and what
// each is working on.
func agentsVerb(stdout, stderr io.Writer) int {
	p, err := storePath()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	db, err := catalogue.OpenRead(p)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer db.Close()

	rows, err := catalogue.Agents(context.Background(), db, sessionsDir())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(rows) == 0 {
		// NOT an empty table. The client writes a session file per live session, so none means
		// nothing is running — a different fact from a store with no rows, and said in words.
		fmt.Fprintln(stdout, "no sessions are advertised in ~/.claude/sessions — nothing is running on this box")
		return 0
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].LastAt > rows[j].LastAt })
	fmt.Fprintf(stdout, "%-10s %-8s %7s  %-14s %s\n", "SESSION", "STATE", "ACTS", "LAST", "CWD")
	for _, a := range rows {
		last := "—"
		if a.LastTool != "" {
			last = a.LastTool
			if a.LastAt > 0 {
				last += " " + humanAgo(a.LastAt)
			}
		}
		fmt.Fprintf(stdout, "%-10s %-8s %7d  %-14s %s\n",
			catalogue.Short(a.SessionID), a.Liveness, a.Acts, last, a.CWD)
	}
	// `unknown` is a real answer and must not be read as a failure: liveness is exact only on
	// Linux, and a session from another pid namespace is not ours to judge.
	for _, a := range rows {
		if a.Liveness == catalogue.Unknown {
			fmt.Fprintln(stdout, "\n  `unknown` means liveness could not be established — a foreign pid namespace, "+
				"or a platform where this is not implemented. It is not `ended`.")
			break
		}
	}
	return 0
}

func humanAgo(ts int64) string {
	d := time.Since(time.Unix(ts, 0))
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// touchedVerb answers "is anyone else editing this file".
func touchedVerb(stdout, stderr io.Writer, path string) int {
	db, code := openStore(stderr)
	if db == nil {
		return code
	}
	defer db.Close()
	hits, err := catalogue.Touched(context.Background(), db, path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(hits) == 0 {
		fmt.Fprintf(stdout, "no recorded act names %s\n", path)
		fmt.Fprintln(stdout, "  ^ that is a statement about the CATALOGUE, not about the file: a session whose "+
			"transcript was never ingested contributes nothing here. `gray-area backfill` reads what exists.")
		return 0
	}
	fmt.Fprintf(stdout, "%-10s %-18s %-8s %-6s %5s  %s\n", "SESSION", "AGENT", "TOOL", "OUT", "N", "WHEN")
	for _, h := range hits {
		agent := h.AgentID
		if agent == "" {
			agent = "(main)"
		}
		fmt.Fprintf(stdout, "%-10s %-18s %-8s %-6s %5d  %s\n",
			catalogue.Short(h.SessionID), agent, h.Tool, h.Outcome, h.N, humanAgo(h.TS))
	}
	return 0
}

// sessionVerb summarises one session's shape.
func sessionVerb(stdout, stderr io.Writer, id string) int {
	db, code := openStore(stderr)
	if db == nil {
		return code
	}
	defer db.Close()
	s, err := catalogue.Shape(context.Background(), db, id)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	state := "open"
	if s.ClosedAt.Valid {
		state = "closed " + humanAgo(s.ClosedAt.Int64)
	}
	fmt.Fprintf(stdout, "session %s  (%s)\n  %s\n\n", catalogue.Short(s.SessionID), state, s.ProjectDir)
	tools := make([]string, 0, len(s.ByTool))
	for k := range s.ByTool {
		tools = append(tools, k)
	}
	sort.Slice(tools, func(i, j int) bool { return s.ByTool[tools[i]] > s.ByTool[tools[j]] })
	fmt.Fprintf(stdout, "  %-16s %7s %7s\n", "TOOL", "CALLS", "ERRORS")
	for _, tl := range tools {
		fmt.Fprintf(stdout, "  %-16s %7d %7d\n", tl, s.ByTool[tl], s.Failures[tl])
	}
	fmt.Fprintf(stdout, "\n  %d words, %d thoughts\n", s.Words, s.Thoughts)
	if s.Thoughts == 0 && s.Words > 0 {
		// The distinction the doctor's warning exists for: reasoning absent because it was never
		// captured is not reasoning absent because there was none.
		fmt.Fprintln(stdout, "  ^ no reasoning captured for this session. If `showThinkingSummaries` was off "+
			"when it ran, that is why — see /prosthetic-conscience:doctor.")
	}
	return 0
}

// sqlVerb is the raw read-only surface: the questions worth asking are not knowable in advance,
// so the schema is published and callers write their own.
func sqlVerb(stdout, stderr io.Writer, query string, limit int) int {
	db, code := openStore(stderr)
	if db == nil {
		return code
	}
	defer db.Close()
	cols, rows, err := catalogue.RawQuery(context.Background(), db, query, limit)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, strings.Join(cols, "\t"))
	for _, r := range rows {
		fmt.Fprintln(stdout, strings.Join(r, "\t"))
	}
	if limit > 0 && len(rows) == limit {
		fmt.Fprintf(stdout, "\n(%d rows shown — the limit, so there may be more)\n", limit)
	}
	return 0
}

// findVerb is the full-text half: ripgrep over the transcripts the catalogue names, joined back
// to who and when. No index is built — a whole-corpus search measures tens of milliseconds.
func findVerb(stdout, stderr io.Writer, term string) int {
	db, code := openStore(stderr)
	if db == nil {
		return code
	}
	defer db.Close()
	paths, err := catalogue.PathsFor(context.Background(), db)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "gray-area find: the catalogue names no transcripts yet — run `gray-area backfill`")
		return 1
	}
	rg, err := exec.LookPath("rg")
	if err != nil {
		// A MISSING BINARY IS AN ERROR, NEVER ZERO HITS. A search that cannot run reports the
		// same "no matches" as one that ran and found none, and every "nobody else touched this"
		// would be a lie.
		fmt.Fprintln(stderr, "gray-area find: ripgrep is not installed, so this search cannot run. "+
			"Refusing rather than reporting zero matches, which would be indistinguishable from a real result. "+
			"Install it (see requirements.json) or use `gray-area sql`.")
		return 1
	}
	args := []string{"--count-matches", "--no-messages", "--", term}
	for p := range paths {
		args = append(args, p)
	}
	// The term is a single argv element and never reaches a shell.
	out, err := exec.Command(rg, args...).Output()
	if err != nil && len(out) == 0 {
		fmt.Fprintf(stdout, "no transcript contains %q\n", term)
		return 0
	}
	type hit struct {
		f catalogue.TranscriptFile
		n string
	}
	var hits []hit
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		i := strings.LastIndex(line, ":")
		if i < 0 {
			continue
		}
		if f, ok := paths[line[:i]]; ok {
			hits = append(hits, hit{f, line[i+1:]})
		}
	}
	if len(hits) == 0 {
		fmt.Fprintf(stdout, "no transcript contains %q\n", term)
		return 0
	}
	fmt.Fprintf(stdout, "%-10s %-18s %6s  %s\n", "SESSION", "AGENT", "HITS", "PATH")
	for _, h := range hits {
		agent := h.f.AgentID
		if agent == "" {
			agent = "(main)"
		}
		fmt.Fprintf(stdout, "%-10s %-18s %6s  %s\n", catalogue.Short(h.f.SessionID), agent, h.n, h.f.Path)
	}
	return 0
}

// backfillVerb projects the existing corpus once. Explicit, never from a hook: it is the cold
// read that closure's per-invocation cap exists to avoid facing.
func backfillVerb(stdout, stderr io.Writer) int {
	p, err := storePath()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	db, err := catalogue.Open(p)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer db.Close()
	files, err := catalogue.TranscriptFiles(projectsRoot())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	start := time.Now()
	var acts, words, thoughts, unparsed int
	var bytes int64
	for _, f := range files {
		r, err := catalogue.IngestFile(db, f)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		acts += r.Acts
		words += r.Words
		thoughts += r.Thoughts
		unparsed += r.Unparsed
		bytes += r.BytesRead
	}
	fmt.Fprintf(stdout, "backfilled %d transcripts, %.0f MB in %.1fs\n  %d acts, %d words, %d thoughts\n",
		len(files), float64(bytes)/1e6, time.Since(start).Seconds(), acts, words, thoughts)
	if unparsed > 0 {
		// Reported, never swallowed: a torn line and an absent line leave the same missing row.
		fmt.Fprintf(stdout, "  %d line(s) did not parse — those records are ABSENT from the store, "+
			"which is a different fault from a session that did less\n", unparsed)
	}
	fmt.Fprintln(stdout, "  re-running is safe: unchanged files are skipped by their stored offset")
	return 0
}

func openStore(stderr io.Writer) (*sql.DB, int) {
	p, err := storePath()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return nil, 1
	}
	db, err := catalogue.OpenRead(p)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return nil, 1
	}
	return db, 0
}
