// Package checkpoint locates and parses the live CHECKPOINT.md note.
//
// It exists because the seal (PreCompact) and the restore (SessionStart) must
// agree on WHICH file is the note. Two copies of that rule drift, and a restore
// reading a different file than the seal wrote is the failure mode with no
// symptom: both halves report success and the continuity is silently gone.
package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
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

// loopEntry matches a line that a reader would take for a validation-loop entry,
// including the malformed spellings ParseValidationLoop declines to open a check
// for. Detecting what was NOT parsed is the whole point of LoopProblems.
var loopEntry = regexp.MustCompile(`^(\d+)([a-z]*)\.(\s|$)`)

// LoopProblems reports validation-loop entries whose written label disagrees
// with their ordinal position, and entries a reader would count that the parser
// does not.
//
// The schema (skills/context-checkpointing/SKILL.md) numbers the loop `1.`, `2.`,
// `3.` … so the written label and the ordinal are the same number. Nothing
// enforced that, and two consequences followed, both live in this repo:
//
//   - A note numbered from `0.` made rearmed.json's key 2 mean the note's `1.`
//     Every human and agent reader used the written label; the state file used
//     the ordinal. Off by one at the seam, silently (#192).
//   - Lettered sub-entries (`0b.`, `7b.`) open no check in EITHER this parser or
//     gray-area's, so a loop a reader counts as eleven checks parses as nine.
//     Two checks were neither watched nor adjudicated, and nothing said so (#193).
//
// Both are reported, never repaired. Silently renumbering would restore exactly
// the ambiguity this exists to expose: the reader's number and the tool's number
// have to be the same number, or the disagreement has to be visible.
//
// This does not refuse. It is called from hooks, and a hook that stopped working
// because a hand-written note has a bad label would trade a numbering slip for a
// loss of continuity — a worse failure than the one it reports. The ordinal
// remains authoritative for attribution; the complaint rides out in the digest.
// gray-area's parser makes the opposite call, correctly: it is a command a human
// invokes, and it can afford to stop.
func LoopProblems(body string) []string {
	var problems []string
	ordinal := 0
	for _, line := range strings.Split(body, "\n") {
		m := loopEntry.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		if m[2] != "" {
			// A lettered label never opens a check, so it does not advance the
			// ordinal either — it is invisible, which is what makes it dangerous.
			problems = append(problems, fmt.Sprintf(
				"entry %q is not an ordinal: labels must be 1, 2, 3 … — this entry opens no check and is neither watched nor adjudicated",
				m[1]+m[2]+"."))
			continue
		}
		ordinal++
		if n, err := strconv.Atoi(m[1]); err == nil && n != ordinal {
			problems = append(problems, fmt.Sprintf(
				"entry %q is in position %d: the written label and the ordinal must be the same number, or state keyed by ordinal will disagree with every reader",
				m[1]+".", ordinal))
		}
	}
	return problems
}

// WithinRoot reports whether a project-relative path exists INSIDE the project.
//
// The production implementation is os.Root.Stat. That matters: containment here
// is a security property, not a tidiness one — the checkpoint is agent-written,
// and this design treats it as a claim rather than a fact, so a note must not be
// able to aim the file watcher at an arbitrary directory.
//
// A lexical check (filepath.Rel, or the strings.HasPrefix it replaced) cannot do
// this. Measured: with `proj/watched` a symlink to an outside tree, Rel reports
// "contained: true" while os.Root reports "path escapes from parent". Lexical
// path arithmetic is blind to symlinks by construction.
type WithinRoot func(rel string) bool

// WatchTargets turns a check's re-armed-by prose into paths a file watcher will
// actually accept, plus a reason when it cannot.
//
// MEASURED CONSTRAINT (plans/hook-surface-spike.md §6): watchPaths takes paths —
// files or directories — and NO pattern of any kind. Not globs, not regex. A
// pattern registers nothing, silently. Directories are watched recursively and
// fire `add` for files created later, so the right move for "any .go edit under
// tools/" is to watch `tools/`, not to expand it.
//
// Results are project-relative and slash-separated, because the re-arm hook
// compares them against a slash-normalised file_path on every platform.
func WatchTargets(reArmedBy string, within WithinRoot) (paths []string, reason string) {
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
		// A cheap lexical pre-filter. NOT the containment check — `within` is —
		// but it drops the obvious cases before touching the filesystem and
		// keeps an absolute path from being handed to a root-relative Stat.
		if strings.HasPrefix(p, "..") || filepath.IsAbs(p) {
			continue
		}
		if !within(p) {
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

// RootedWithin builds a WithinRoot backed by os.Root, so containment survives a
// symlink. A project directory that cannot be opened yields a checker that
// admits nothing — refusing to watch is the safe failure, and the caller reports
// the surface as unresolved rather than watching something it could not verify.
//
// The returned closer must be called; it is the open directory handle.
func RootedWithin(projectDir string) (WithinRoot, func() error) {
	root, err := os.OpenRoot(projectDir)
	if err != nil {
		return func(string) bool { return false }, func() error { return nil }
	}
	return func(rel string) bool {
		_, err := root.Stat(filepath.FromSlash(rel))
		return err == nil
	}, root.Close
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

// RearmState is the whole file: newest entry per check IDENTITY. A check
// re-armed twice is re-armed; the count is not the signal, the fact is.
//
// LastEvent is stamped on every delivered event, matched or not. That is the
// only field here that can distinguish "nothing has changed" from "nothing has
// ARRIVED" — the distinction #165 cost three sessions for want of. An unmatched
// event still records nothing per-check, deliberately (a directory watch is
// coarser than the surface it stands for, so unclaimed changes are expected and
// are not evidence); but it does prove the watcher is alive, and that is worth
// exactly one timestamp.
type RearmState struct {
	Rearmed   map[string]Rearm `json:"rearmed"`
	LastEvent string           `json:"last_event,omitempty"` // UTC RFC3339, any delivered event
}

// CheckKey is a check's identity: its line with the ordinal label stripped.
//
// KEYED BY IDENTITY, NOT POSITION. Keying by ordinal made this file's key `2`
// mean the note's `1.` (#192), and made every renumber orphan every record
// silently — measured: promoting two sub-entries shifted nine keys at once, and
// the only reason it was visible at all is that each record also stores its
// check line. If the line is what identifies the record, the position never
// needed to be the key.
//
// The tradeoff, stated: editing a check's first line changes its identity and
// orphans its record. That is the right failure — an edited check has a
// different expectation, so its old `last run` should not carry over — and it
// costs one stale entry rather than a whole file re-pointed at the wrong checks.
func CheckKey(line string) string {
	return strings.TrimSpace(loopEntry.ReplaceAllString(strings.TrimSpace(line), ""))
}

// RearmPath is where the state lives for a project.
func RearmPath(projectDir string) string {
	return filepath.Join(projectDir, ".claude", "checkpoints", RearmFile)
}

// LoadRearm reads the state, distinguishing ABSENT from UNREADABLE.
//
// A missing file is an empty state and no error: the first run has no history.
// Anything else — unreadable, corrupt, truncated — is an ERROR, and the caller
// must not save over it.
//
// The old behaviour returned an empty state for both, and `save` then wrote a
// single-key file. So any damage to this file silently reset coverage to a lone
// record beside surfaces that were demonstrably being edited — which is exactly
// the symptom #165 described and spent three sessions failing to explain. A
// state file whose loss is indistinguishable from the bug it reports on is worse
// than no state file.
//
// Migration is automatic and quiet: records written under the old ordinal keys
// carry their check line, so identity is recoverable from the record itself.
func LoadRearm(read func(string) ([]byte, error), path string) (RearmState, error) {
	s := RearmState{Rearmed: map[string]Rearm{}}
	b, err := read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, fmt.Errorf("re-arm state at %s exists but could not be read: %w — refusing to overwrite it; delete it deliberately if that is what you want", path, err)
	}
	var parsed RearmState
	if err := json.Unmarshal(b, &parsed); err != nil || parsed.Rearmed == nil {
		return s, fmt.Errorf("re-arm state at %s is present but unparsable — refusing to overwrite it; delete it deliberately if that is what you want", path)
	}
	s.LastEvent = parsed.LastEvent
	for k, r := range parsed.Rearmed {
		// Re-key on load. Old files keyed by ordinal migrate here, because the
		// record has always carried the line that identifies it.
		if want := CheckKey(r.Check); r.Check != "" && want != k {
			k = want
		}
		s.Rearmed[k] = r
	}
	return s, nil
}

// UpdateRearm is the ONLY safe way to modify the state file.
//
// MEASURED, not theorised: six FileChanged events fired concurrently against a
// six-check note recorded FOUR records. Two were lost. Every hook process was
// doing its own read-modify-write on one file — each loaded the state as it
// stood, added its own key, and wrote the whole map back, so the last writer
// erased whatever landed after its own read. The state file also tore: a short
// write over a longer file left the previous document's tail behind, producing
// JSON that parses as garbage. A live file in this repo was found in exactly
// that condition, with five records stamped the same second.
//
// That is #165. The symptom was "coverage stops advancing" and the suspected
// cause was someone deleting the file; the actual cause is that concurrent
// events destroy each other's records, and the old LoadRearm turned the
// resulting corruption into a silent reset.
//
// So: lock, read, mutate, write atomically, unlock. The read MUST happen inside
// the lock — loading before acquiring it reintroduces the same lost update with
// extra steps.
func UpdateRearm(path string, mutate func(*RearmState)) error {
	unlock, err := lockRearm(path)
	if err != nil {
		return err
	}
	defer unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cannot create state dir: %w", err)
	}
	s, err := LoadRearm(os.ReadFile, path)
	if err != nil {
		return err
	}
	mutate(&s)

	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	// Write-then-rename. os.WriteFile truncates, so it cannot tear on its own —
	// but two of them racing can, because truncate and write are not one step.
	// A rename is.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return fmt.Errorf("cannot write state: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("cannot commit state: %w", err)
	}
	return nil
}

// lockRearmTimeout bounds how long a hook waits for its turn, and how old a lock
// must be before it is assumed abandoned. A hook that blocks a file event is
// worse than a hook that occasionally loses one, so this is short.
const lockRearmTimeout = 2 * time.Second

// lockSeq distinguishes concurrent holders inside ONE process, where the pid is
// the same for all of them.
var lockSeq atomic.Uint64

// lockNonce identifies this acquisition. pid alone is not enough — the tests and
// the hook can both hold this lock from several goroutines at once.
func lockNonce() string {
	return strconv.Itoa(os.Getpid()) + "-" + strconv.FormatUint(lockSeq.Add(1), 10)
}

// lockRearm takes an exclusive lock via O_EXCL create, which works on every
// platform this ships to — unlike flock, which would need a Windows variant for
// the CI this repo actually runs.
//
// # THE LOCK NAMES ITS HOLDER, AND THAT IS NOT DECORATION (#215)
//
// The first version wrote an empty lock file and broke a stale one by removing
// it. That can violate the mutual exclusion it exists to provide:
//
//	A acquires at T. B arrives at T+2.1s, judges the lock abandoned, removes it,
//	creates its own. A finishes and its unlock removes THE LOCK B IS HOLDING.
//	C now acquires while B is still mid-update — two writers, no exclusion.
//
// So the file carries a nonce, and both dangerous operations are conditional on
// it: unlock removes only a lock that is still ours, and a stale break removes
// only the exact lock that was observed to be stale.
//
// Residual, stated rather than hidden: there is no portable filesystem
// compare-and-delete, so between reading a nonce and removing that file another
// process could in principle replace it. The window is a syscall wide instead of
// seconds wide, and a loser of that race fails its O_EXCL create and retries.
func lockRearm(path string) (func(), error) {
	lock := path + ".lock"
	me := lockNonce()
	deadline := time.Now().Add(lockRearmTimeout)
	for {
		f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, writeErr := f.WriteString(me)
			_ = f.Close()
			if writeErr != nil {
				// A lock nobody can identify is worse than no lock: it can never
				// be safely broken or released.
				_ = os.Remove(lock)
				return nil, fmt.Errorf("could not write the re-arm lock at %s: %w", lock, writeErr)
			}
			return func() {
				// Ours only. After a stale break this file belongs to someone
				// else, and removing it would hand a third process a lock the
				// second one still thinks it holds.
				if held, readErr := os.ReadFile(lock); readErr == nil && string(held) == me {
					_ = os.Remove(lock)
				}
			}, nil
		}

		// A process that died holding the lock must not wedge the mechanism
		// forever — losing a re-arm is cheap and losing the file is not.
		if fi, statErr := os.Stat(lock); statErr == nil && time.Since(fi.ModTime()) > lockRearmTimeout {
			stale, readErr := os.ReadFile(lock)
			if readErr == nil {
				// Re-read immediately before removing: if the nonce moved, the
				// lock was replaced since it looked abandoned and is live.
				if now, err2 := os.ReadFile(lock); err2 == nil && string(now) == string(stale) {
					_ = os.Remove(lock)
				}
			}
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("re-arm state at %s is locked by another hook and did not free within %s — this event is not recorded, nothing is damaged", path, lockRearmTimeout)
		}
		time.Sleep(5 * time.Millisecond)
	}
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
