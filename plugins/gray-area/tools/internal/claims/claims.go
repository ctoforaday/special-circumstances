// Package claims reads a sealed checkpoint note as a set of DECLARED CLAIMS.
//
// A validation-loop entry asserts three separable things — a command was run, it
// passed, and it was last run at a stated time — and all three are self-report.
// This package turns that prose into something adjudicable; it judges nothing.
//
// THE PARSER IS RESTATED, NOT SHARED. prosthetic-conscience's
// checkpoint.ParseValidationLoop is the canonical reader of this format, and the
// two plugins are separate Go modules by design (a consumer must be able to take
// the miner without taking continuity, and vice versa). The drift that buys is
// real, so it is handled the only honest way available across a module boundary:
// a note that HAS a validation-loop heading and yields ZERO claims is an error,
// never an empty result. Silent success on an unrecognized format would make
// every downstream row read "nothing to report" when the truth is "could not
// read it".
package claims

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Claim is one validation-loop entry, with the note coordinates that cite it.
type Claim struct {
	File  string `json:"file"`
	Line  int    `json:"line"` // 1-based, the entry's first line
	Index string `json:"index"`

	Text    string `json:"text"`              // the entry's first line, verbatim
	Cmd     string `json:"cmd,omitempty"`     // the backticked command, when there is one
	Surface string `json:"surface,omitempty"` // text after "re-armed by:"

	// The asserted result. Outcome is whatever word followed "last run:" —
	// deliberately NOT parsed into a boolean, because "FAILING, see #165" is a
	// legitimate thing for a note to say and flattening it to false would lose
	// the reference a reader needs.
	Outcome string `json:"outcome,omitempty"`
	At      string `json:"at,omitempty"` // RFC3339-ish timestamp found in the "last run" line
}

// HasCommand reports whether there is anything to look for in a trajectory. A
// prose check ("the continuity loop itself") has no command, and inventing one
// would manufacture the false negative this whole design refuses to produce.
func (c Claim) HasCommand() bool { return strings.TrimSpace(c.Cmd) != "" }

var (
	heading  = regexp.MustCompile(`(?i)^#{1,6}\s+validation loop\s*$`)
	nextHead = regexp.MustCompile(`^#{1,6}\s+`)
	// An entry opens with "N." at the left margin. Indented continuation lines
	// belong to the entry above.
	entry = regexp.MustCompile(`^(\d+)\.\s+(.*)$`)
	// Anything a reader would count as an entry, including the spellings `entry`
	// declines. Detecting what is NOT parsed is the point: a lettered label was
	// invisible to both parsers, so a loop of eleven checks adjudicated as nine.
	malformed = regexp.MustCompile(`^(\d+)([a-z]*)\.(\s|$)`)
	backticks = regexp.MustCompile("`([^`]+)`")
	rearmed   = regexp.MustCompile(`(?i)re-armed by:\s*(.+?)\s*$`)
	lastRun   = regexp.MustCompile(`(?i)last run:\s*(.+?)\s*$`)
	// Timestamps as the notes write them: 2026-07-30T04:41Z, optionally with seconds.
	stamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?Z`)
)

// Parse extracts every claim under the note's validation-loop heading.
//
// It returns an error when the heading is present and nothing parses beneath it,
// which is the drift alarm described in the package comment.
func Parse(r io.Reader, file string) ([]Claim, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var (
		out       []Claim
		inLoop    bool
		sawLoop   bool
		lineNo    int
		headingAt int
	)
	for sc.Scan() {
		lineNo++
		line := sc.Text()

		if heading.MatchString(strings.TrimSpace(line)) {
			inLoop, sawLoop, headingAt = true, true, lineNo
			continue
		}
		if !inLoop {
			continue
		}
		if nextHead.MatchString(line) {
			inLoop = false
			continue
		}

		// A label that is not its own ordinal makes every index in this output a
		// lie, and the output's whole job is to be citable. Refused rather than
		// reported: unlike prosthetic-conscience's parser, which runs inside
		// hooks and must degrade rather than break continuity, this runs in a
		// command a human invoked and can afford to stop. See #192, #193.
		if m := malformed.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			if m[2] != "" {
				return nil, fmt.Errorf("%s:%d: entry %q is not an ordinal — validation-loop labels must be 1, 2, 3 …; this entry opens no check, so adjudicating would silently omit it", file, lineNo, m[1]+m[2]+".")
			}
			if n, err := strconv.Atoi(m[1]); err == nil && n != len(out)+1 {
				return nil, fmt.Errorf("%s:%d: entry %q is in position %d — the written label and the ordinal must be the same number, or every index this command prints disagrees with the state keyed by ordinal", file, lineNo, m[1]+".", len(out)+1)
			}
		}

		if m := entry.FindStringSubmatch(line); m != nil {
			c := Claim{File: file, Line: lineNo, Index: m[1], Text: strings.TrimSpace(line)}
			if b := backticks.FindStringSubmatch(m[2]); b != nil {
				c.Cmd = b[1]
			}
			if s := rearmed.FindStringSubmatch(m[2]); s != nil {
				c.Surface = s[1]
			}
			absorb(&c, m[2])
			out = append(out, c)
			continue
		}

		// A continuation line extends the entry above it: the "last run" line
		// lives there, and so does a trigger surface that wrapped.
		if len(out) > 0 && strings.TrimSpace(line) != "" && (line[0] == ' ' || line[0] == '\t') {
			last := &out[len(out)-1]
			// THE COMMAND COMES FROM THE ENTRY'S FIRST LINE, NEVER A CONTINUATION.
			//
			// Continuations carry `last run:` and wrapped surfaces, and they quote
			// things in backticks that are not commands. A real note's prose entry
			// continues "...re-armed by an `add` event", and taking that as the
			// command made the adjudicator search the trajectory for "add" and report
			// NO-EVIDENCE against a check it had invented a command for — the exact
			// confident false accusation this design refuses to produce.
			//
			// The asymmetry decides it: missing a command yields UNCHECKABLE, which is
			// honest, while inventing one yields an accusation that reads as fact.
			if last.Surface == "" {
				if s := rearmed.FindStringSubmatch(line); s != nil {
					last.Surface = s[1]
				}
			}
			absorb(last, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("claims: reading %s: %w", file, err)
	}
	if sawLoop && len(out) == 0 {
		return nil, fmt.Errorf("claims: %s has a validation-loop heading at line %d but no entry parsed beneath it — the note format has changed and this reader has drifted from it. Reporting nothing here would read as \"no claims to check\" when the truth is \"could not read the claims\"", file, headingAt)
	}
	return out, nil
}

// absorb pulls the asserted outcome and timestamp out of a line, without
// deciding what they mean.
func absorb(c *Claim, s string) {
	if m := lastRun.FindStringSubmatch(s); m != nil && c.Outcome == "" {
		c.Outcome = strings.TrimSpace(m[1])
	}
	if c.At == "" {
		// The timestamp is taken from the "last run" text when there is one, so a
		// date mentioned elsewhere in the entry cannot be mistaken for the run time.
		src := s
		if m := lastRun.FindStringSubmatch(s); m != nil {
			src = m[1]
		} else if c.Outcome == "" {
			return
		}
		if t := stamp.FindString(src); t != "" {
			c.At = t
		}
	}
}

// Signature reduces a claimed command to the tokens a trajectory command must
// contain to be a plausible match: the binary, its subcommand, and anything
// path-shaped.
//
// Token containment rather than equality, because the two spellings are never
// the same string — `go test -C x ./...` and `cd x && go test ./... | grep FAIL`
// are one check. The signature is EMITTED with every verdict, so an approximate
// match shows its working instead of asserting a precision it does not have.
func Signature(cmd string) []string {
	var sig []string
	seen := map[string]bool{}
	add := func(t string) {
		if t != "" && !seen[t] {
			seen[t] = true
			sig = append(sig, t)
		}
	}
	fields := strings.Fields(cmd)
	for i, f := range fields {
		f = strings.Trim(f, `"'`)
		switch {
		case i == 0:
			add(f) // the binary
		case i == 1 && !strings.HasPrefix(f, "-"):
			add(f) // its subcommand
		case strings.Contains(f, "/") && f != "./..." && !strings.HasPrefix(f, "-"):
			add(strings.TrimSuffix(strings.TrimSuffix(f, "/..."), "/"))
		}
	}
	return sig
}
