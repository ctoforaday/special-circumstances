// Package checkpoint locates and parses the live CHECKPOINT.md note.
//
// It exists because the seal (PreCompact) and the restore (SessionStart) must
// agree on WHICH file is the note. Two copies of that rule drift, and a restore
// reading a different file than the seal wrote is the failure mode with no
// symptom: both halves report success and the continuity is silently gone.
package checkpoint

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Exists reports whether a path is a readable regular file. Injected so callers
// can test the search order without a filesystem.
type Exists func(string) bool

// NotePath returns the live checkpoint location, or "" when none exists.
// A run/project workspace owns the note when one is active; otherwise it falls
// back to the session-local (gitignored) directory.
//
// Glob is used for the workspace cases and a plain check for the fallback, so
// the caller's Exists is the single authority on whether a candidate counts.
func NotePath(projectDir string, exists Exists, glob func(string) ([]string, error)) string {
	if projectDir == "" {
		return ""
	}
	for _, dir := range []string{"projects", "research"} {
		matches, _ := glob(filepath.Join(projectDir, dir, "*", "CHECKPOINT.md"))
		sort.Strings(matches)
		for _, m := range matches {
			if exists(m) {
				return m
			}
		}
	}
	fallback := filepath.Join(projectDir, ".claude", "checkpoints", "CHECKPOINT.md")
	if exists(fallback) {
		return fallback
	}
	return ""
}

// Note is a parsed checkpoint. Values are taken verbatim from the file; nothing
// is inferred or defaulted, because a restore that invents a field hands the
// resumed session a claim its own note never made.
type Note struct {
	Front    map[string]string // frontmatter keys, values unquoted
	Sections []Section         // "## " sections, in file order
}

// Section is one "## " heading and its body.
type Section struct {
	Heading string
	Body    string
}

// Front returns a frontmatter value, or "" when absent.
func (n Note) Get(key string) string { return n.Front[key] }

// Section returns the named section's body and whether it was present.
// Absent and empty are different: an empty validation loop is a note that
// recorded no checks, and a missing one is a note written before they existed.
func (n Note) Section(heading string) (string, bool) {
	for _, s := range n.Sections {
		if strings.EqualFold(s.Heading, heading) {
			return s.Body, true
		}
	}
	return "", false
}

// NonEmptySection returns the body only when it carries something beyond
// whitespace and schema scaffolding (the "← load-bearing" style annotations the
// skill's template shows). A section holding only its own placeholder is
// nothing established, and restoring it would introduce rather than reinforce.
func (n Note) NonEmptySection(heading string) (string, bool) {
	body, ok := n.Section(heading)
	if !ok {
		return "", false
	}
	var kept []string
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "←") || strings.HasPrefix(t, "<!--") {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return "", false
	}
	return strings.Join(kept, "\n"), true
}

// Parse reads a checkpoint's YAML-ish frontmatter and "## " sections.
//
// Deliberately not a YAML parser: the schema is flat scalars, and pulling a YAML
// dependency into a hook binary that must never fail buys nothing. A line that
// does not look like "key: value" is skipped rather than treated as an error —
// the note is written by an agent under time pressure, and a restore that
// refuses a slightly malformed note is a restore that fires when it is least
// needed.
func Parse(raw string) Note {
	n := Note{Front: map[string]string{}}
	body := raw

	if rest, ok := strings.CutPrefix(raw, "---\n"); ok {
		if front, after, found := strings.Cut(rest, "\n---"); found {
			for _, line := range strings.Split(front, "\n") {
				k, v, ok := strings.Cut(line, ":")
				if !ok {
					continue
				}
				k = strings.TrimSpace(k)
				v = strings.TrimSpace(v)
				v = strings.Trim(v, `"'`)
				if k != "" && !strings.HasPrefix(k, "#") {
					n.Front[k] = v
				}
			}
			body = strings.TrimPrefix(after, "\n")
		}
	}

	var cur *Section
	var buf []string
	flush := func() {
		if cur != nil {
			cur.Body = strings.TrimRight(strings.Join(buf, "\n"), "\n")
			n.Sections = append(n.Sections, *cur)
		}
		buf = nil
	}
	for _, line := range strings.Split(body, "\n") {
		if h, ok := strings.CutPrefix(line, "## "); ok {
			flush()
			s := Section{Heading: strings.TrimSpace(h)}
			cur = &s
			continue
		}
		if cur != nil {
			buf = append(buf, line)
		}
	}
	flush()
	return n
}

// Headings lists the "## " headings of a note, in file order.
func Headings(raw string) []string {
	var out []string
	for _, s := range Parse(raw).Sections {
		out = append(out, s.Heading)
	}
	return out
}

// Check is one entry in the note's validation loop.
//
// The schema puts the trigger surface on the same line as the command:
//
//  1. `go test ./...`  → all packages ok  · re-armed by: any .go edit under tools/
//     last run: pass
//
// ReArmedBy is free prose by design — the skill asks the agent to say what makes
// a check fire, not to write a matcher. Turning that prose into something a file
// watcher can register is WatchTargets' job, and it is allowed to fail.
type Check struct {
	Line      string // the check's first line, verbatim, for citation
	Index     int    // 1-based position in the loop
	ReArmedBy string // text after "re-armed by:", trimmed; "" when absent
}

// reArmedMarker is matched case-insensitively; the schema writes "re-armed by:"
// but a note is hand-written and "Re-armed by:" is the same intent.
const reArmedMarker = "re-armed by:"

// ParseValidationLoop pulls the numbered checks out of a validation-loop body.
//
// A line starting "<n>." opens a check; everything until the next such line
// belongs to it. Continuation lines are scanned for the marker too, because an
// agent under pressure wraps.
func ParseValidationLoop(body string) []Check {
	var out []Check
	numbered := func(s string) bool {
		t := strings.TrimSpace(s)
		i := 0
		for i < len(t) && t[i] >= '0' && t[i] <= '9' {
			i++
		}
		return i > 0 && i < len(t) && t[i] == '.'
	}
	for _, line := range strings.Split(body, "\n") {
		if numbered(line) {
			out = append(out, Check{Line: strings.TrimSpace(line), Index: len(out) + 1})
		}
		if len(out) == 0 {
			continue
		}
		cur := &out[len(out)-1]
		if i := strings.Index(strings.ToLower(line), reArmedMarker); i >= 0 && cur.ReArmedBy == "" {
			rest := strings.TrimSpace(line[i+len(reArmedMarker):])
			// The schema separates trailing clauses with "·"; keep only the first.
			if j := strings.Index(rest, "·"); j >= 0 {
				rest = strings.TrimSpace(rest[:j])
			}
			cur.ReArmedBy = rest
		}
	}
	return out
}

// WatchTargets turns a check's re-armed-by prose into paths a file watcher will
// actually accept, plus a reason when it cannot.
//
// MEASURED CONSTRAINT (plans/hook-surface-spike.md §6): watchPaths takes paths —
// files or directories — and NO pattern of any kind. Not globs, not regex. A
// pattern registers nothing, silently. Directories are watched recursively and
// fire `add` for files created later, so the right move for "any .go edit under
// tools/" is to watch `tools/`, not to expand it.
//
// exists reports whether a path is present and whether it is a directory.
// Everything is resolved relative to projectDir; a target that escapes it is
// dropped, because a checkpoint should not be able to point the watcher at /etc.
func WatchTargets(reArmedBy, projectDir string, exists func(string) (isDir bool, ok bool)) (paths []string, reason string) {
	if strings.TrimSpace(reArmedBy) == "" {
		return nil, "no trigger surface recorded"
	}
	seen := map[string]bool{}
	for _, tok := range candidateTokens(reArmedBy) {
		p := watchableDir(tok)
		if p == "" {
			continue
		}
		// "**"-rooted patterns resolve to the project root. Watching the root is
		// measured to catch the hooks' own writes and re-trigger them, so it is
		// refused rather than silently accepted.
		if p == "." || p == "/" {
			reason = "trigger surface resolves to the project root; refused (a root watch re-triggers on hook output)"
			continue
		}
		full := filepath.Join(projectDir, p)
		// Prefix comparison is NOT containment: "/w2/x" has the prefix "/w" and
		// is not inside it. Rel is the honest test, and it is also the one that
		// behaves the same on Windows, where Join produces backslashes.
		if rel, err := filepath.Rel(projectDir, full); err != nil ||
			rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if _, ok := exists(full); !ok {
			continue
		}
		if !seen[p] {
			seen[p] = true
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 && reason == "" {
		reason = "no resolvable path in " + strconv.Quote(reArmedBy)
	}
	if len(paths) > 0 {
		reason = ""
	}
	return paths, reason
}

// candidateTokens splits prose into things that might be paths. Deliberately
// generous: WatchTargets discards what does not resolve, and a missed surface is
// cheaper than a wrong one only because the miss is REPORTED.
func candidateTokens(s string) []string {
	s = strings.NewReplacer("`", " ", "\"", " ", "'", " ", ",", " ", ";", " ", "(", " ", ")", " ").Replace(s)
	var out []string
	for _, f := range strings.Fields(s) {
		// Trim TRAILING punctuation only. Trimming leading dots destroys every
		// dotfile path — ".qlty/qlty.toml" became "qlty/qlty.toml", which
		// resolves to nothing, silently, which is the failure mode this whole
		// function exists to avoid.
		f = strings.TrimRight(f, ".,:;")
		if f == "" {
			continue
		}
		if strings.ContainsAny(f, "/*?[") || strings.HasPrefix(f, ".") {
			out = append(out, f)
		}
	}
	return out
}

// watchableDir reduces a path-ish token to the deepest directory that contains
// no wildcard. "manifest/*.yml" -> "manifest"; "tools/" -> "tools";
// "**/*.go" -> "." (which WatchTargets refuses); "cmd/x/main.go" -> unchanged.
func watchableDir(tok string) string {
	tok = strings.TrimPrefix(tok, "./")
	if tok == "" {
		return ""
	}
	segs := strings.Split(tok, "/")
	var kept []string
	for _, s := range segs {
		if strings.ContainsAny(s, "*?[") {
			break
		}
		if s != "" {
			kept = append(kept, s)
		}
	}
	if len(kept) == 0 {
		return "."
	}
	joined := strings.Join(kept, "/")
	// A bare filename with an extension and no separator is not a useful watch
	// target on its own — it is prose like "go.mod" only when it really exists,
	// which the caller checks.
	return joined
}

// --- re-arm state -----------------------------------------------------------
//
// Written by the FileChanged hook, read by the restore hook. It lives here for
// the same reason NotePath does: two copies of a format drift, and a reader that
// disagrees with its writer fails with both halves reporting success.

// RearmFile is the state file's name, under .claude/checkpoints/.
const RearmFile = "rearmed.json"

// Rearm records that one validation check's trigger surface moved, so its
// recorded `last run` result is stale.
type Rearm struct {
	Index int    `json:"index"` // 1-based position in the validation loop
	Check string `json:"check"` // the check's line, verbatim, for citation
	By    string `json:"by"`    // the file that re-armed it, project-relative
	Event string `json:"event"` // change | add | unlink
	At    string `json:"at"`    // UTC RFC3339
}

// RearmState is the whole file: newest entry per check index. A check re-armed
// twice is re-armed; the count is not the signal, the fact is.
type RearmState struct {
	Rearmed map[string]Rearm `json:"rearmed"`
}

// RearmPath is where the state lives for a project.
func RearmPath(projectDir string) string {
	return filepath.Join(projectDir, ".claude", "checkpoints", RearmFile)
}

// LoadRearm reads the state. A missing or corrupt file yields an empty state
// rather than an error: losing re-arm history must never cost a tool call, and
// must never block a session start either.
func LoadRearm(read func(string) ([]byte, error), path string) RearmState {
	s := RearmState{Rearmed: map[string]Rearm{}}
	b, err := read(path)
	if err != nil {
		return s
	}
	var parsed RearmState
	if json.Unmarshal(b, &parsed) == nil && parsed.Rearmed != nil {
		s = parsed
	}
	return s
}

// Sorted returns the re-arm records in check order, so a digest lists them the
// way the note does rather than in map order.
func (s RearmState) Sorted() []Rearm {
	out := make([]Rearm, 0, len(s.Rearmed))
	for _, r := range s.Rearmed {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out
}
