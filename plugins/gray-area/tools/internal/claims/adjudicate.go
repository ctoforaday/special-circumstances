package claims

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

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
}

// bashCommand returns a Bash event's command with heredoc BODIES removed, or "".
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
// settled. That asymmetry is the whole reason this function is not a one-liner.
func bashCommand(e trajectory.Event) string {
	if e.ToolName != "Bash" || len(e.ToolInput) == 0 {
		return ""
	}
	var in struct {
		Command string `json:"command"`
	}
	_ = json.Unmarshal(e.ToolInput, &in)
	return stripHeredocs(in.Command)
}

// heredocOpen finds `<< DELIM`, `<<-DELIM`, `<<'DELIM'` or `<<"DELIM"`.
var heredocOpen = regexp.MustCompile(`<<-?\s*(['"]?)([A-Za-z_][A-Za-z0-9_]*)(['"]?)`)

// stripHeredocs removes each heredoc's body, keeping the command lines around it.
// An unterminated heredoc drops the rest of the string — the shell would have fed
// it all to the reader anyway, so none of it is command.
func stripHeredocs(cmd string) string {
	lines := strings.Split(cmd, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		out = append(out, lines[i])
		m := heredocOpen.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		delim := m[2]
		i++
		for i < len(lines) && strings.TrimSpace(lines[i]) != delim {
			i++
		}
		// i now sits on the delimiter (dropped with the body) or past the end.
	}
	return strings.Join(out, "\n")
}

// writeTarget returns the path a Write/Edit event acted on, or "".
func writeTarget(e trajectory.Event) string {
	switch e.ToolName {
	case "Write", "Edit", "NotebookEdit", "MultiEdit":
	default:
		return ""
	}
	if len(e.ToolInput) == 0 {
		return ""
	}
	var in struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	}
	_ = json.Unmarshal(e.ToolInput, &in)
	if in.FilePath != "" {
		return in.FilePath
	}
	return in.Path
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

// underSurface reports whether a written path falls under a claim's trigger
// surface. Suffix-based rather than filesystem-based on purpose: the trajectory
// holds absolute paths from the machine that ran, and the surface is written
// repo-relative, so the two only ever meet as strings. A surface with no `/` and
// no leading `.` is not treated as a path at all — the same rule the restore
// hook applies, and for the same reason: bare `scripts` is a word, `scripts/` is
// a place.
func underSurface(written, surface string) bool {
	surface = strings.TrimSpace(surface)
	if surface == "" || (!strings.Contains(surface, "/") && !strings.HasPrefix(surface, ".")) {
		return false
	}
	surface = strings.TrimSuffix(surface, "/")
	written = strings.ReplaceAll(written, "\\", "/")
	return strings.Contains(written, "/"+surface+"/") ||
		strings.HasSuffix(written, "/"+surface) ||
		strings.HasPrefix(written, surface+"/") ||
		written == surface
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
		var hit *trajectory.Event
		for i := range citable {
			if cmd := bashCommand(citable[i]); cmd != "" && matches(cmd, sig) {
				hit = &citable[i]
			}
		}
		if hit == nil {
			f.Verdict = NoEvidence
			f.Note = "no Bash invocation in this trajectory contains every token above. That is an ABSENCE OF EVIDENCE, not evidence the check did not run: the command may have been spelled in a form these tokens miss, or run outside this transcript"
			out = append(out, f)
			continue
		}

		f.Verdict, f.File, f.Line, f.UUID, f.At = Cited, hit.File, hit.Line, hit.UUID, hit.Timestamp
		f.Command = matchingLine(bashCommand(*hit), sig)

		// STALENESS, computed without any help from the mechanism that is supposed
		// to report it. The trajectory holds every write with its target and time,
		// so "the claim is older than the last write to its own trigger surface" is
		// answerable here even when FileChanged never fires (#165).
		if c.Surface != "" && c.At != "" {
			if w := lastWriteUnder(citable, c.Surface, c.At); w != nil {
				f.Verdict = Stale
				f.File, f.Line, f.UUID, f.At = w.File, w.Line, w.UUID, w.Timestamp
				f.Command = writeTarget(*w)
				f.Note = "a write under this check's own trigger surface (" + c.Surface + ") happened AFTER the claimed run time. The pass may have been real; it is not current"
			}
		}
		out = append(out, f)
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].NoteLine < out[j].NoteLine })
	return out
}

// lastWriteUnder returns the latest citable write under surface with a timestamp
// strictly after `after`, or nil.
//
// String comparison on RFC3339 timestamps is deliberate: they sort lexically when
// they carry the same offset, which these do (the notes and the transcript both
// write Z). Parsing them would add a failure mode — an unparseable stamp becoming
// a silently skipped comparison — for no gain here.
func lastWriteUnder(events []trajectory.Event, surface, after string) *trajectory.Event {
	var latest *trajectory.Event
	for i := range events {
		p := writeTarget(events[i])
		if p == "" || !underSurface(p, surface) {
			continue
		}
		ts := events[i].Timestamp
		if ts == "" || ts <= after {
			continue
		}
		if latest == nil || ts > latest.Timestamp {
			latest = &events[i]
		}
	}
	return latest
}
