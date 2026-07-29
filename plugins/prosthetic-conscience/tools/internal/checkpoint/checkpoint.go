// Package checkpoint locates and parses the live CHECKPOINT.md note.
//
// It exists because the seal (PreCompact) and the restore (SessionStart) must
// agree on WHICH file is the note. Two copies of that rule drift, and a restore
// reading a different file than the seal wrote is the failure mode with no
// symptom: both halves report success and the continuity is silently gone.
package checkpoint

import (
	"path/filepath"
	"sort"
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
