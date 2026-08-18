package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// The rule, and the reason it is a gate rather than a habit.
//
// A surface an agent reads carries prose about a thing that NO LONGER EXISTS — a retired verb, a
// deleted file, a mode that was removed — so every seat spends context learning what it cannot
// use. That is `archaeology-in-live-surface` in feov-memory/protocol-class-registry.md, and its
// sweep question is the one that decides each site: could a seat ACT on this sentence today?
//
// It was swept by hand once. The count went UP during the same session, because an unrelated
// branch landed fresh sites while the sweep was in flight. A sweep with nothing checking it is
// `policy-without-mechanism` — a class in that same registry — so this is the mechanism.
//
// # ADDED LINES ONLY, and that is not leniency
//
// Several hundred sites predate the rule. A gate that failed on all of them would fail every
// pull request for a reason its author did not cause, which is how a gate gets turned off. This
// asks the narrower question a diff can answer honestly: did THIS change add archaeology?
//
// # WHAT IT DOES NOT REACH, stated so the gap is not mistaken for coverage
//
// Go and test comments are OUT OF SCOPE. The cost this gate exists to stop is paid per seat per
// round — bytes that enter a prompt every sitting — and a code comment is read by whoever opens
// the file. The same phrases are also load-bearing in code far more often ("this used to fail
// when…" justifying a guard that still stands), so gating them would train an author to route
// around the gate rather than to write plainly.

// agentFacing are the globs whose bytes reach a seat or an operator agent, relative to the repo
// root. Enumerated per plugin rather than by a bare `plugins/**` so that adding a plugin with a
// new surface shape is a deliberate act here, not a silent gap.
var agentFacing = []string{
	"plugins/*/agents/*.md",
	"plugins/*/commands/*.md",
	"plugins/*/skills/*/SKILL.md",
	"plugins/*/skills/*/*.md",
	"plugins/*/skills/*/scripts/*.js",
	"CLAUDE.md",
}

// marker is one archaeology phrase and the plain-language reason it reads as one.
type marker struct {
	re  *regexp.Regexp
	why string
}

// markers are deliberately HIGH-PRECISION rather than broad.
//
// A bare `no longer` is ordinary live prose — "a realized risk is no longer a probability" states
// a current fact — so matching it would produce false alarms, and false alarms are what teach an
// author to ignore a gate. Each pattern here names a thing in the PAST TENSE that a seat cannot
// act on now.
var markers = []marker{
	{regexp.MustCompile(`\bRETIRED\b`), "a retirement notice: the seat cannot use what it names"},
	{regexp.MustCompile(`\b(is|was|are|were)\s+retired\b`), "a retirement notice: the seat cannot use what it names"},
	{regexp.MustCompile(`\bused to\b`), "describes how something behaved before; a seat can only act on how it behaves now"},
	{regexp.MustCompile(`\bno longer\s+(exists?|creates?|writes?|replays?|runs?|fires?|reads?|carries|carry)\b`), "names a capability that is gone"},
	{regexp.MustCompile(`\b(has|have)\s+been\s+(removed|retired|deleted)\b`), "names a thing that is gone"},
	{regexp.MustCompile(`\b(was|were)\s+(removed|deleted)\b`), "names a thing that is gone"},
}

// finding is one archaeology phrase on one added line.
type finding struct {
	file, text, why string
}

func (f finding) String() string {
	return fmt.Sprintf("%s\n      + %s\n      ^ %s", f.file, strings.TrimSpace(f.text), f.why)
}

// isAgentFacing reports whether a repo-relative path is a surface an agent reads.
func isAgentFacing(path string) bool {
	path = filepath.ToSlash(path)
	for _, g := range agentFacing {
		if ok, _ := filepath.Match(g, path); ok {
			return true
		}
	}
	return false
}

// scan finds archaeology in the ADDED lines of a unified diff.
//
// It reads the diff rather than the files so that a pre-existing site does not fail a change that
// merely touched the same file. `+++ b/<path>` sets the current file; a line starting with a
// single `+` is an addition.
func scan(diff string) []finding {
	var out []finding
	file := ""
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+++ b/"):
			file = strings.TrimPrefix(line, "+++ b/")
			continue
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			continue
		case !strings.HasPrefix(line, "+"):
			continue
		}
		if file == "" || !isAgentFacing(file) {
			continue
		}
		text := strings.TrimPrefix(line, "+")
		for _, m := range markers {
			if m.re.MatchString(text) {
				out = append(out, finding{file: file, text: text, why: m.why})
				break // one finding per line; the first reason is enough to act on
			}
		}
	}
	return out
}
