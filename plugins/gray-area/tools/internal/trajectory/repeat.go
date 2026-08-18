package trajectory

import (
	"encoding/json"
	"sort"
	"strings"
)

// Repetition is the substrate for Phase 2's rework and stalls inspections: the
// SAME act, done again.
//
// # Why grouping is a package concern and not a caller's loop
//
// Both inspections ask "what happened more than once", and both are one careless
// line away from a plausible zero. A grouping keyed on something that no longer
// appears — a tool renamed, an input field moved — yields zero groups, which is
// byte-identical to a session where nothing was repeated. So the key is derived
// from FIELDS the parser already models, and Group reports how many events it
// could not key at all rather than letting them vanish from the denominator.

// Repeat is one act performed more than once, with every occurrence cited.
//
// Occurrences are in trajectory order, so the first is the original and the last
// is the most recent — which is the direction both inspections read them.
type Repeat struct {
	// Key is what made these the same act. Human-readable on purpose: it is
	// printed, and a reader who disagrees with the grouping must be able to see
	// what the grouping was.
	Key  string `json:"key"`
	Kind string `json:"kind"` // "write" | "command"
	// Target is the file path or the command's signature, whichever Kind says.
	Target string `json:"target"`
	// Occurrences carries EVERY occurrence, not a count. A count cannot be
	// checked; an event with file, line and uuid can.
	Occurrences []Event `json:"occurrences"`
	// Positions is each occurrence's index in the citable, keyed stream — the
	// basis for Consecutive. EXPORTED and serialised on purpose: adjacency is a
	// claim about ordering, and a reader who wants to disagree with Consecutive()
	// needs the numbers it was computed from. An unexported field would make the
	// verdict unfalsifiable from the output.
	Positions []int `json:"positions"`
}

// Count is the number of times the act was performed.
func (r Repeat) Count() int { return len(r.Occurrences) }

// Consecutive reports the longest run of occurrences that were ADJACENT in the
// trajectory's citable order, which is what separates "ran the suite at three
// points in the session" from "ran the suite three times in a row".
//
// The distinction is the whole of the stalls inspection: [[anti-spinning]]'s
// 3-strike rule is about consecutive repair attempts, and a rule counted over a
// whole session would fire on ordinary iteration.
func (r Repeat) Consecutive() int {
	best, run := 0, 0
	prev := -1
	for _, p := range r.Positions {
		if prev >= 0 && p == prev+1 {
			run++
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
		prev = p
	}
	return best
}

// GroupResult is what Group answers with. Unkeyed is present so a reader can
// tell "nothing repeated" from "most of it could not be keyed" — the two would
// otherwise produce the same empty list.
type GroupResult struct {
	Repeats []Repeat `json:"repeats"`
	// Considered is how many citable tool uses were examined.
	Considered int `json:"considered"`
	// Unkeyed is how many of those produced no key: a tool this grouping does not
	// model, or an input whose target field was absent. NOT an error — a Read is
	// legitimately unkeyed — but it bounds every answer derived from Repeats.
	Unkeyed int `json:"unkeyed"`
	// Singles is how many keys occurred exactly once. Reported because a run with
	// 400 distinct acts and no repeats is a very different session from one with
	// 3, and both give len(Repeats) == 0.
	Singles int `json:"singles"`
}

// WriteTarget returns the path a file-modifying tool acted on, or "".
//
// Exported because both this package's grouping and claims' staleness check need
// the same answer, and two copies of "which field holds the path" is exactly the
// drift that makes one of them silently stop matching.
func WriteTarget(e Event) string {
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
	if json.Unmarshal(e.ToolInput, &in) != nil {
		return ""
	}
	if in.FilePath != "" {
		return in.FilePath
	}
	return in.Path
}

// BashCommand returns a Bash event's command string, or "".
func BashCommand(e Event) string {
	if e.ToolName != "Bash" || len(e.ToolInput) == 0 {
		return ""
	}
	var in struct {
		Command string `json:"command"`
	}
	if json.Unmarshal(e.ToolInput, &in) != nil {
		return ""
	}
	return in.Command
}

// Group buckets citable tool uses into repeated acts.
//
// Only events that can be CITED are considered: a row nobody can check is the
// thing this tool exists not to emit, and letting an uncitable event into a
// count would put it behind a number where nobody could see it.
func Group(events []Event) GroupResult {
	var res GroupResult
	buckets := map[string]*Repeat{}
	var order []string

	seq := 0
	for _, e := range events {
		if e.Kind != "tool_use" || !e.Cited() {
			continue
		}
		res.Considered++
		kind, target := "", ""
		if p := WriteTarget(e); p != "" {
			kind, target = "write", p
		} else if c := BashCommand(e); c != "" {
			if s := CommandKey(c); s != "" {
				kind, target = "command", s
			}
		}
		if kind == "" {
			res.Unkeyed++
			continue
		}
		key := kind + ":" + target
		b, ok := buckets[key]
		if !ok {
			b = &Repeat{Key: key, Kind: kind, Target: target}
			buckets[key] = b
			order = append(order, key)
		}
		b.Occurrences = append(b.Occurrences, e)
		b.Positions = append(b.Positions, seq)
		seq++
	}

	for _, k := range order {
		b := buckets[k]
		if b.Count() < 2 {
			res.Singles++
			continue
		}
		res.Repeats = append(res.Repeats, *b)
	}
	// Most-repeated first, then by key so the order is stable across runs — a
	// listing that reshuffles between identical runs cannot be diffed.
	sort.SliceStable(res.Repeats, func(i, j int) bool {
		if a, b := res.Repeats[i].Count(), res.Repeats[j].Count(); a != b {
			return a > b
		}
		return res.Repeats[i].Key < res.Repeats[j].Key
	})
	return res
}

// CommandKey reduces a shell command to the act it performs, so two runs of the
// same check group together despite incidental differences.
//
// IT MAY MISS A GROUP; IT MAY NOT INVENT ONE. Two commands with different
// arguments are usually different acts, and merging them reports rework that
// never happened. A missed group is a smaller lie than an invented one — the same
// asymmetry NO-EVIDENCE rests on.
//
// # TRUNCATION IS WHAT INVENTS GROUPS, and two rounds of measurement to learn it
//
// The first two versions took the first meaningful line and cut it down —
// dropping a `cd` prefix, an assignment, a redirect. Each round passed its own
// unit tests and each round produced fresh invented groups on this repository's
// real session:
//
//	round 1  "8 IN A ROW  cd /home/user/special-circumstances"   a prefix
//	         "5 IN A ROW  python3 - <<'PY'"                      a heredoc opener
//	round 2  "4 IN A ROW  s=open(p).read()"                      a heredoc BODY line
//	         "3 IN A ROW  import json,os,datetime"               likewise, via `cd ...; python3 -c "`
//
// Every one of them is the same defect: the key threw away the part that told two
// commands apart. So the rule is now the opposite of what it was — **keep the
// whole command** and normalise only what a shell would call plumbing. Under it,
// `cd repo && go test` and `cd repo && git push` cannot collide, because nothing
// is discarded that distinguishes them, and no prefix grammar has to be right.
//
// The cost is accepted and stated: `go test ./...` and `cd repo && go test ./...`
// now key differently, so a group is missed. That is the direction the contract
// permits. Long keys are truncated FOR DISPLAY by the caller, never for keying —
// the distinction the first two versions collapsed.
//
// The one thing still refused is a command whose act lives in a heredoc body:
// after the body is stripped every such command reads alike, so keying them would
// merge every inline script in a session. Refusing lands the event in Unkeyed,
// where the coverage line reports it — visibly not grouped, rather than grouped
// wrongly.
func CommandKey(cmd string) string {
	stripped := StripHeredocs(cmd)
	// A surviving opener means the act was in the body, which is now gone.
	for _, line := range strings.Split(stripped, "\n") {
		if OpensHeredoc(line) {
			return ""
		}
	}
	// Drop comment-only lines: a comment is not part of the act, and keeping it
	// makes two runs of one command differ because the note above them changed.
	var body []string
	for _, line := range strings.Split(stripped, "\n") {
		if t := strings.TrimSpace(line); t != "" && !strings.HasPrefix(t, "#") {
			body = append(body, t)
		}
	}
	// Collapse whitespace so a reformatted-but-identical command still groups.
	key := strings.Join(strings.Fields(strings.Join(body, " ")), " ")
	// Drop what a shell would treat as plumbing rather than as the act, so two
	// invocations differing only in a redirect group.
	for _, cut := range []string{" > ", " >> ", " 2>", " | ", " && echo ", " ; echo "} {
		if i := strings.Index(key, cut); i > 0 {
			key = key[:i]
		}
	}
	key = strings.TrimSpace(key)
	// Navigation alone is not an act.
	if key == "" || key == "cd" || strings.HasPrefix(key, "cd ") && !strings.ContainsAny(key, "&;|") {
		return ""
	}
	return key
}
