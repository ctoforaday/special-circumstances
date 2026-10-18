package claims

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/trajectory"
)

// Verdict is what the adjudication says about one claim. The names are chosen so
// that reading the output aloud cannot overstate it.
type Verdict string

const (
	// Cited: a plausible invocation is in the trajectory, and its uuid is here.
	Cited Verdict = "CITED"
	// Stale: a write under the claim's own trigger surface happened AFTER the
	// claimed run time. The pass may have been real and is no longer current.
	Stale Verdict = "STALE"
	// NoEvidence is NOT "did not run" — see the Finding docs. It is the delicate
	// one, and the naming is the point.
	NoEvidence Verdict = "NO-EVIDENCE"
	// Uncheckable: the claim names no command, so there is nothing to look for.
	Uncheckable Verdict = "UNCHECKABLE"
)

// Finding is one adjudicated claim. Provenance is not optional: every row cites
// the note it came from, and every row that asserts something about the
// trajectory cites the trajectory too.
//
// ON `NO-EVIDENCE`. It records what was searched (Searched) and how much was
// searched (EventsSeen), because a miner that reports absence as fact is a
// trusted component that can lie silently — the risk this plugin's own plan
// names G4. The asymmetry is the reason: a matcher that is too narrow produces a
// confident false accusation, while one that is too broad produces a citation a
// reader can see is wrong. So the tool states its search and the reader convicts.
type Finding struct {
	Verdict Verdict `json:"verdict"`

	// The claim's side of the provenance.
	NoteFile  string `json:"note_file"`
	NoteLine  int    `json:"note_line"`
	Index     string `json:"index"`
	Claim     string `json:"claim"`
	Claimed   string `json:"claimed_outcome,omitempty"`
	ClaimedAt string `json:"claimed_at,omitempty"`

	// AsOf is the time staleness was measured from, and AsOfFrom says where it
	// came from: "last run" (the entry's own run time) or "written_at" (the
	// note's, an upper bound on every run it reports). Empty when staleness was
	// not measured — Unmeasured then says why.
	AsOf     string `json:"as_of,omitempty"`
	AsOfFrom string `json:"as_of_from,omitempty"`

	// The trajectory's side, when there is one.
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	UUID    string `json:"uuid,omitempty"`
	At      string `json:"at,omitempty"`
	Command string `json:"command,omitempty"`

	// How the answer was reached. Present on EVERY row, including the negative
	// ones — an unstated search is indistinguishable from no search.
	Searched   []string `json:"searched,omitempty"`
	EventsSeen int      `json:"events_seen"`
	Note       string   `json:"note,omitempty"`

	// Unmeasured names the parts of the claim THIS RECORD CANNOT SPEAK TO, so a
	// verdict on the rest cannot be read as a verdict on the whole.
	//
	// The case it exists for: a claim asserts an ACT and a RESULT — "`go run
	// ./check` → 26 passed". The trajectory holds invocations and not their
	// output (result bodies are conversation content, refused under G4/G6), so
	// the act is adjudicable and the count is not. Silence would let CITED read
	// as endorsing the number, which is the false positive
	// plans/gray-area-phase-2.md §0 named as the builder's problem.
	//
	// [[facts-are-fields]] clause 3, aimed at this tool's own output: where the
	// miss cannot be made refusable, make it LOUD rather than let it fold into
	// the honest answer.
	Unmeasured []string `json:"unmeasured,omitempty"`
}

// bashCommand returns a Bash event's command with heredoc BODIES removed, "" for
// any other event, and an error when the command does not parse as shell.
//
// The stripping is not tidiness, it is correctness. Matching is token containment
// over the command string, and a heredoc body is data, not command — so a
// `python3 - <<PY` script that merely MENTIONS plugins/prosthetic-conscience/tools
// matched a claim to run the tests there, and the tool reported CITED against an
// event that ran no test at all. Measured on this repo's own transcript: two of
// nine claims were cited to heredocs that only quoted their paths.
//
// A false citation is worse than a miss. A miss says "no evidence" and shows its
// search, so a reader checks; a wrong citation says "here it is" and looks
// settled. So a command whose bodies cannot be told from its command lines is
// not searched at all, and the miss row counts it.
func bashCommand(e trajectory.Event) (string, error) {
	cmd := trajectory.BashCommand(e)
	if cmd == "" {
		return "", nil
	}
	stripped, _, err := trajectory.StripHeredocs(cmd)
	return stripped, err
}

// matches reports whether a trajectory command contains every signature token.
func matches(command string, sig []string) bool {
	if len(sig) == 0 {
		return false
	}
	for _, t := range sig {
		if !strings.Contains(command, t) {
			return false
		}
	}
	return true
}

// matchingLine returns the line of a multi-line command that carries the match,
// so the evidence a reader sees is the evidence the match was made on.
//
// Without this, a compound command that edits a file and THEN runs the check
// displayed as `python3 - <<'PY'` — a correct citation presented as an obviously
// wrong one, which costs exactly the trust the provenance exists to build.
func matchingLine(cmd string, sig []string) string {
	for _, l := range strings.Split(cmd, "\n") {
		if matches(l, sig) {
			return strings.TrimSpace(l)
		}
	}
	// No single line carries every token — the match spans the command, so show it
	// whole rather than picking a line that would misrepresent where it came from.
	return cmd
}

// surfacePaths reads the paths out of a trigger surface written as prose —
// "agentgen src, scripts/" names one path, `scripts`. A token is a path when it
// holds a `/` or starts with `.`: bare `scripts` is a word, `scripts/` is a
// place. A glob keeps its directory part ("tools/**/*.go" is tools). The rule is
// prosthetic-conscience's checkpoint.WatchTargets token rule, restated across
// the module boundary; a surface resolving to the root names nothing.
func surfacePaths(surface string) []string {
	s := strings.NewReplacer("`", " ", "\"", " ", "'", " ", ",", " ", ";", " ", "(", " ", ")", " ").Replace(surface)
	var out []string
	seen := map[string]bool{}
	for _, f := range strings.Fields(s) {
		f = strings.TrimRight(f, ".,:;")
		if f == "" || !(strings.ContainsAny(f, "/*?[") || strings.HasPrefix(f, ".")) {
			continue
		}
		var kept []string
		for _, seg := range strings.Split(strings.TrimPrefix(f, "./"), "/") {
			if strings.ContainsAny(seg, "*?[") {
				break
			}
			if seg != "" {
				kept = append(kept, seg)
			}
		}
		p := strings.Join(kept, "/")
		if p == "" || p == "." || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

// underSurface reports whether a written path falls under one path read by
// surfacePaths. Suffix-based rather than filesystem-based on
// purpose: the trajectory holds absolute paths from the machine that ran, and
// the surface is written repo-relative, so the two only ever meet as strings.
func underSurface(written, surface string) bool {
	written = strings.ReplaceAll(written, "\\", "/")
	return strings.Contains(written, "/"+surface+"/") ||
		strings.HasSuffix(written, "/"+surface) ||
		strings.HasPrefix(written, surface+"/") ||
		written == surface
}

// underAny reports whether a written path falls under any of paths.
func underAny(written string, paths []string) bool {
	for _, p := range paths {
		if underSurface(written, p) {
			return true
		}
	}
	return false
}

// Adjudicate puts each claim against the trajectory and returns one finding per
// claim, in the note's own order.
//
// It concludes nothing beyond what the two records say when placed side by side.
func Adjudicate(cs []Claim, res trajectory.Result) []Finding {
	uses := res.ToolUses()

	// Only citable events can support a finding — a row nobody can check is the
	// thing this tool exists not to emit.
	var citable []trajectory.Event
	for _, e := range uses {
		if e.Cited() {
			citable = append(citable, e)
		}
	}

	out := make([]Finding, 0, len(cs))
	for _, c := range cs {
		f := Finding{
			NoteFile: c.File, NoteLine: c.Line, Index: c.Index,
			Claim: c.Text, Claimed: c.Outcome, ClaimedAt: c.At,
			EventsSeen: len(citable),
		}

		if !c.HasCommand() {
			f.Verdict = Uncheckable
			f.Note = "the claim names no command, so there is nothing to look for in the trajectory. Guessing one would manufacture a false negative"
			out = append(out, f)
			continue
		}

		sig := Signature(c.Cmd)
		f.Searched = sig

		// The LAST matching invocation is the one that matters: a check run twice
		// is current as of the second run.
		var (
			hit      *trajectory.Event
			hitCmd   string
			unparsed int
		)
		for i := range citable {
			cmd, err := bashCommand(citable[i])
			if err != nil {
				unparsed++
				continue
			}
			if cmd != "" && matches(cmd, sig) {
				hit, hitCmd = &citable[i], cmd
			}
		}
		if hit == nil {
			f.Verdict = NoEvidence
			f.Note = "no Bash invocation in this trajectory contains every token above. That is an ABSENCE OF EVIDENCE, not evidence the check did not run: the command may have been spelled in a form these tokens miss, or run outside this transcript"
			if unparsed > 0 {
				f.Unmeasured = append(f.Unmeasured, fmt.Sprintf(
					"%d Bash command(s) in this trajectory do not parse as shell, so their heredoc bodies cannot be told from their command lines; they were not searched", unparsed))
			}
			out = append(out, f)
			continue
		}

		f.Verdict, f.File, f.Line, f.UUID, f.At = Cited, hit.File, hit.Line, hit.UUID, hit.Timestamp
		f.Command = matchingLine(hitCmd, sig)

		// THE ACT/RESULT BOUNDARY, and it belongs HERE rather than in one caller.
		//
		// A validation-loop entry asserts an ACT and an OUTCOME — "`go run ./check`
		// … last run: pass". This record can adjudicate the act and cannot see the
		// outcome: the trajectory holds invocations, never their output (result
		// bodies are conversation content, refused under G4/G6). A CITED row that
		// says nothing about the outcome reads as endorsing the claimed pass.
		//
		// It was built into the pull request reader first and left out of this one
		// — `adjacent-seat-omission`, the class the sweep exists to catch, in the
		// tool that adjudicates. Putting it in Adjudicate rather than in each verb
		// is what stops the third caller from omitting it again.
		if c.Outcome != "" {
			f.Unmeasured = append(f.Unmeasured,
				"the claimed outcome ("+strings.TrimSpace(c.Outcome)+") — the trajectory records invocations, not their output")
		}

		// STALENESS, computed without any help from the mechanism that is supposed
		// to report it. The trajectory holds every write with its target and time,
		// so "the claim is older than the last write to its own trigger surface" is
		// answerable here whatever FileChanged does or does not report (#165).
		// That issue first claimed the events never fire; withdrawn — they do, but
		// coverage stops advancing mid-session for reasons still unexplained. This
		// computation depends on none of it, which is the point.
		//
		// Every input it needs can be missing, and each miss is NAMED: a CITED row
		// that skipped the check silently reads as a current pass.
		staleness(&f, c, citable)
		out = append(out, f)
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].NoteLine < out[j].NoteLine })
	return out
}

// staleness measures one cited claim against later writes under its trigger
// surface, turning the row STALE when one exists, or names in Unmeasured the
// input it lacked.
//
// The run time is the entry's own "last run" stamp when it has one. Without one
// it is the note's written_at: every run the note reports happened before the
// note was written, so a write after written_at is certainly after the run. The
// bound is one-sided on purpose — a write between the run and the note is
// missed, never invented.
func staleness(f *Finding, c Claim, citable []trajectory.Event) {
	paths := surfacePaths(c.Surface)
	asOf, from := c.At, "last run"
	if asOf == "" {
		// The value up to its first space: the template writes a trailing
		// "# comment" after it.
		if v := strings.Fields(c.NoteWrittenAt); len(v) > 0 {
			asOf = v[0]
		}
		from = "written_at"
	}
	switch {
	case strings.TrimSpace(c.Surface) == "":
		f.Unmeasured = append(f.Unmeasured,
			"staleness — the entry names no trigger surface (re-armed by:), so no later write can be tied to it")
		return
	case len(paths) == 0:
		f.Unmeasured = append(f.Unmeasured,
			"staleness — the trigger surface ("+strings.TrimSpace(c.Surface)+") names no path, so no write can be placed under it")
		return
	case asOf == "":
		f.Unmeasured = append(f.Unmeasured,
			"staleness — the entry states no run time and the note no written_at, so no write can be placed after the run")
		return
	}
	t, ok := ParseTime(asOf)
	if !ok {
		f.Unmeasured = append(f.Unmeasured,
			"staleness — the note's written_at ("+asOf+") is not an RFC 3339 time, and the entry states no run time")
		return
	}
	f.AsOf, f.AsOfFrom = asOf, from

	w, unstamped := lastWriteUnder(citable, paths, t)
	if unstamped > 0 {
		f.Unmeasured = append(f.Unmeasured, fmt.Sprintf(
			"staleness over %d write(s) under %s — their timestamps do not parse, so they cannot be placed before or after the run", unstamped, strings.Join(paths, ", ")))
	}
	if w == nil {
		return
	}
	f.Verdict = Stale
	f.File, f.Line, f.UUID, f.At = w.File, w.Line, w.UUID, w.Timestamp
	f.Command = trajectory.WriteTarget(*w)
	if from == "written_at" {
		f.Note = "a write under this check's own trigger surface (" + c.Surface + ") happened AFTER the note was written (written_at " + asOf + "; the entry states no run time, and every run it reports came before the note). The pass may have been real; it is not current"
		return
	}
	f.Note = "a write under this check's own trigger surface (" + c.Surface + ") happened AFTER the claimed run time. The pass may have been real; it is not current"
}

// lastWriteUnder returns the latest citable write under surface strictly after
// `after`, or nil, and how many writes under it carry no parseable timestamp.
//
// Times are parsed, not compared as strings: "04:41Z" sorts after "04:41:30Z"
// lexically, and a note's minute-precision stamp meets a transcript's
// millisecond one here.
func lastWriteUnder(events []trajectory.Event, paths []string, after time.Time) (*trajectory.Event, int) {
	var (
		latest   *trajectory.Event
		latestAt time.Time
		bad      int
	)
	for i := range events {
		p := trajectory.WriteTarget(events[i])
		if p == "" || !underAny(p, paths) {
			continue
		}
		ts, ok := ParseTime(events[i].Timestamp)
		if !ok {
			bad++
			continue
		}
		if !ts.After(after) {
			continue
		}
		if latest == nil || ts.After(latestAt) {
			latest, latestAt = &events[i], ts
		}
	}
	return latest, bad
}
