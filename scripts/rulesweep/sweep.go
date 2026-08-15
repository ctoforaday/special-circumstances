// Package main implements rule-sweep: a rule patch names its class and sweeps the
// siblings.
//
// Dev tooling for this repository only (CLAUDE.md: dev concerns live at the root).
// Nothing here ships to an installing project.
//
// THE PROBLEM. Every protocol patch that shipped recurred in adjacent form — lens labels
// patched, blue-lane footnote namespaces bit next run; blue-reads-transcript patched, the
// lossy gap summary bit next run; grade enums widened, the mass mapping distorted next
// run; friction-to-file shipped, lens seats still had no write path. Four for four. Each
// fix was correct and each was an INSTANCE fix: the class survived, re-emerged at the
// adjacent seat, and passed every test the patch had added, because those tests were about
// the instance too.
//
// THE RULE. A change to a PROTOCOL SURFACE must name its class from
// feov-memory/protocol-class-registry.md and state the sibling sweep:
//
//	Rule-Class: adjacent-seat-omission
//	Sibling-Sweep: checked blue-lane, blue-synthesize, blue-respond — all three
//	               carry the clause; red-merge does not share this duty
//
// WHY MECHANICAL. A sweep requirement enforced by good intentions is an instance of
// policy-without-mechanism, which is a class in the registry it would be enforcing. This
// is the mechanism; CI runs it on every pull request.
//
// It is deliberately a PROMPT, not a proof: it cannot know whether a sweep was honest,
// only whether one was stated. That is the same tier as the round-parity check —
// presence, not truth — and it is the tier that stops the failure we actually measured,
// which was nobody looking at the siblings at all.
package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// protocolSurfaces are the files that govern how the engine BEHAVES, as opposed to the
// code that implements it. A change to debate.js's dispatch arithmetic is engineering; a
// change to what a seat is TOLD is protocol, and it is the second kind whose fixes kept
// recurring.
var protocolSurfaces = []*regexp.Regexp{
	regexp.MustCompile(`/skills/[^/]+/SKILL\.md$`),
	regexp.MustCompile(`/agents/[^/]+\.md$`),
	regexp.MustCompile(`/references/[^/]+\.md$`),
	regexp.MustCompile(`/commands/[^/]+\.md$`),
	regexp.MustCompile(`/scripts/debate\.js$`),
	regexp.MustCompile(`^law/`),
}

func touchesProtocol(files []string) []string {
	var out []string
	for _, f := range files {
		for _, re := range protocolSurfaces {
			if re.MatchString(f) {
				out = append(out, f)
				break
			}
		}
	}
	return out
}

// classHeading matches the registry's ### headings. Classes come from ONE source, so the
// checker and the document cannot drift into disagreeing about what a valid class is.
var classHeading = regexp.MustCompile(`(?m)^###\s+([a-z][a-z0-9-]+)\s*$`)

func knownClasses(registry string) []string {
	var out []string
	for _, m := range classHeading.FindAllStringSubmatch(registry, -1) {
		out = append(out, m[1])
	}
	return out
}

var (
	trailerLine      = regexp.MustCompile(`^[A-Za-z][A-Za-z-]*:\s+\S`)
	continuationLine = regexp.MustCompile(`^\s+\S`)
	blankLineSplit   = regexp.MustCompile(`\n\s*\n`)
	ruleClassLine    = regexp.MustCompile(`(?i)^Rule-Class:\s*(.+)$`)
	siblingSweepLine = regexp.MustCompile(`(?i)^Sibling-Sweep:\s*(.+)$`)
)

// trailerBlock returns the git trailer paragraphs of a commit message.
//
// The first version of this check scanned the WHOLE message, and the commit introducing
// the gate exposed it: prose EXPLAINING the mechanism ("...must carry Rule-Class: and
// Sibling-Sweep: trailers...") satisfied the gate. A check that passes on a DESCRIPTION of
// the act rather than the act is precisely the defect class this gate exists to prevent,
// which made it an unusually direct way to learn it.
//
// Any paragraph where EVERY line is a `Key: value` or an indented continuation counts —
// not merely the last one: attribution trailers (Co-Authored-By, Claude-Session)
// conventionally sit last, so requiring the final paragraph would force authors to
// interleave unrelated trailers, and a rule that fights the convention it rides on gets
// worked around rather than obeyed.
//
// The strictness that matters is EVERY line, not SOME: a prose paragraph explaining the
// mechanism contains a line starting "Sibling-Sweep:" and must not count.
func trailerBlock(message string) string {
	var kept []string
	for _, para := range blankLineSplit.Split(strings.TrimSpace(message), -1) {
		var lines []string
		for _, l := range strings.Split(para, "\n") {
			if strings.TrimSpace(l) != "" {
				lines = append(lines, l)
			}
		}
		if len(lines) == 0 {
			continue
		}
		all := true
		for _, l := range lines {
			if !trailerLine.MatchString(l) && !continuationLine.MatchString(l) {
				all = false
				break
			}
		}
		if all {
			kept = append(kept, para)
		}
	}
	return strings.Join(kept, "\n")
}

// minSweepLength rejects a token sweep. "done" states nothing about what was checked; the
// threshold is a crude proxy for "an actual sentence", and crude is the honest tier here —
// this gate checks presence, not truth.
const minSweepLength = 10

func extractTrailers(messages []string) (classes, sweeps []string) {
	for _, m := range messages {
		for _, line := range strings.Split(trailerBlock(m), "\n") {
			line = strings.TrimSpace(line)
			if c := ruleClassLine.FindStringSubmatch(line); c != nil {
				for _, part := range strings.Split(c[1], ",") {
					if p := strings.TrimSpace(part); p != "" {
						classes = append(classes, p)
					}
				}
			}
			if s := siblingSweepLine.FindStringSubmatch(line); s != nil {
				if v := strings.TrimSpace(s[1]); len(v) > minSweepLength {
					sweeps = append(sweeps, v)
				}
			}
		}
	}
	return classes, sweeps
}

// input is what evaluate needs; keeping it a value makes every branch reachable from a
// test without a git repository or a network.
type input struct {
	files    []string
	messages []string
	classes  []string
	// dirty is the set of files modified in the WORKING TREE but not committed.
	// files above comes from `git diff base...HEAD`, which is committed-only, so a
	// branch whose protocol edits are all uncommitted produced an empty files and
	// the cheerful "nothing to sweep" — see unmeasurable below.
	dirty []string
}

type verdict struct {
	ok      bool
	skipped bool
	// unmeasurable marks the one state this gate cannot judge: protocol surfaces
	// ARE modified, but only in the working tree. Trailers live in commit messages,
	// so there is nothing to read yet. That is not a pass and it is not a failure —
	// it is the gate saying it did not fire, which is the whole point (#409).
	unmeasurable bool
	reason       string
	touched      []string
	problems     []string
	named        []string
	sweeps       []string
}

func evaluate(in input) verdict {
	touched := touchesProtocol(in.files)
	if len(touched) == 0 {
		if pending := touchesProtocol(in.dirty); len(pending) > 0 {
			return verdict{
				unmeasurable: true,
				touched:      pending,
				reason: "protocol surfaces are modified but UNCOMMITTED. Trailers are read from commit " +
					"messages, so this gate has nothing to check yet — it did not pass, it did not run. Commit, then re-run.",
			}
		}
		return verdict{ok: true, skipped: true, reason: "no protocol surface touched"}
	}

	named, sweeps := extractTrailers(in.messages)
	var problems []string
	if len(named) == 0 {
		problems = append(problems, "no Rule-Class: trailer — name the class this defect belongs to (feov-memory/protocol-class-registry.md)")
	}
	var unknown []string
	for _, c := range named {
		found := false
		for _, k := range in.classes {
			if c == k {
				found = true
				break
			}
		}
		if !found {
			unknown = append(unknown, c)
		}
	}
	if len(unknown) > 0 {
		problems = append(problems, fmt.Sprintf("unknown class(es): %s — use a registry slug, or add the class WITH its definition, neighbour and distinguishing question", strings.Join(unknown, ", ")))
	}
	if len(sweeps) == 0 {
		problems = append(problems, "no Sibling-Sweep: trailer — state which adjacent seats or surfaces carry this same duty and what you found at each")
	}
	return verdict{ok: len(problems) == 0, touched: touched, problems: problems, named: named, sweeps: sweeps}
}

// ranges: THE TWO RANGES ARE NOT THE SAME RANGE, and conflating them was a live defect
// this gate shipped with (#146).
//
//	git diff A...B   changes on B since the merge base        <- what a pull request is
//	git log  A...B   the SYMMETRIC DIFFERENCE — commits on
//	                 EITHER side and not both                 <- not that
//
// When the branch is behind its base, the three-dot LOG range silently includes commits
// that are on main and not in the branch at all. If any of them carries a Rule-Class:
// trailer, the gate passes on somebody else's sweep. Measured, not theorised: a
// trailer-less commit was waved through on the trailer of 4b4fdca, from an unrelated pull
// request. CI passes --base origin/main with fetch-depth 0, so origin/main is the CURRENT
// tip rather than the branch point — which made the gate EASIER TO PASS THE BUSIER THE
// REPO WAS. Exactly backwards, and invisible, because the output still named a class.
func ranges(base string) (diff, log string) {
	return base + "...HEAD", base + "..HEAD"
}

// collect gathers evaluate's two inputs from a real repository. Separated from main so a
// test can drive it against a scratch repo — range construction was previously unreachable
// from any test, which is how the wrong range survived.
func collect(base, dir string) (files, messages, dirty []string, err error) {
	diffRange, logRange := ranges(base)

	out, err := gitx.Run(dir, "diff", "--name-only", diffRange)
	if err != nil {
		return nil, nil, nil, err
	}
	files = gitx.Lines(out)

	out, err = gitx.Run(dir, "log", "--format=%B%x00", logRange)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, m := range strings.Split(out, "\x00") {
		if strings.TrimSpace(m) != "" {
			messages = append(messages, m)
		}
	}

	// The working tree, staged and unstaged, against HEAD. This is NOT part of the
	// sweep — trailers cannot exist for an uncommitted change. It exists so the gate
	// can tell "nothing to sweep" apart from "cannot sweep yet" (#409).
	out, err = gitx.Run(dir, "diff", "--name-only", "HEAD")
	if err != nil {
		return nil, nil, nil, err
	}
	dirty = gitx.Lines(out)

	return files, messages, dirty, nil
}
