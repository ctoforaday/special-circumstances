package claims

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Coverage answers the question a seat-scoped inspection must not print a number
// without: does the manifest name every seat transcript that exists?
//
// # Two directions, and only one of them was ever measured
//
// #189 established that every `kind: "seat"` row names a transcript that exists —
// 19 of 19, no phantoms. That is a statement about the ROWS, and it does not bound
// the rows that were never written. Measured the other way, against the files, the
// hook saw 19 of 20: one transcript sits in the session's own subagents directory
// that no manifest row has ever named (#469).
//
// So a seat-scoped answer has two failure modes and this package now covers both:
//
//	no phantom seats   every seat row names a file that exists      confirmed
//	no missed seats    every file on disk is named by a row         THIS
//
// The second is where a plausible zero hides. "No rework found across this run's
// seats" reads identically whether the seats were clean or whether one of them was
// never in the manifest to inspect.
//
// # Why this is not the glob plans/gray-area.md §3 refuses
//
// §3 rules out "sweeping ~/.claude/projects/ guessing at which files belong to
// which seat". The objection is ATTRIBUTION BY GUESSING, and its failure mode is a
// FALSE CITATION — a finding pointing at the wrong seat's transcript, which is the
// worst output this plugin can produce.
//
// This reads ONE directory, whose location is derived from the transcript path the
// harness itself handed over at SessionStart, and it attributes nothing. It cannot
// emit a citation at all: its output is a set difference over ids. The two acts
// share a system call and nothing else — auditing the handover is not replacing it.
//
// The derivation `<transcript>.jsonl` -> `<transcript>/subagents/` is a CONVENTION
// rather than a handover, though, and that is the one thing here that can go
// quietly wrong: if the layout moves, the directory reads empty and "nothing
// unnamed" becomes indistinguishable from "we looked in the wrong place". Measured
// returns false in that case and callers MUST NOT report a number.
type Coverage struct {
	// Dir is the directory examined, so a reader can check it by hand.
	Dir string `json:"dir"`
	// Measured is false when the answer is UNKNOWN rather than clean. Callers must
	// branch on this before printing any coverage figure.
	Measured bool `json:"measured"`
	// Why explains an unmeasured result, in a sentence a human can act on.
	Why string `json:"why,omitempty"`

	// OnDisk is every agent id with a transcript file in Dir.
	OnDisk []string `json:"on_disk"`
	// Named is every agent id the manifest's seat rows name.
	Named []string `json:"named"`
	// Unnamed is OnDisk minus Named: transcripts no row accounts for. THE ANSWER.
	Unnamed []string `json:"unnamed"`
	// Missing is Named minus OnDisk: rows naming a file that is not there. Expected
	// to be empty — #189 measured 0 — and reported so that stops being an
	// assumption the tool cannot contradict.
	Missing []string `json:"missing"`
}

// SubagentDir derives the seat-transcript directory from the session transcript
// path the harness handed over. Returns "" for a path it cannot transform, rather
// than assembling something that would read as an empty directory.
func SubagentDir(sessionTranscript string) string {
	p := strings.TrimSpace(sessionTranscript)
	if p == "" || !strings.HasSuffix(p, ".jsonl") {
		return ""
	}
	return filepath.Join(strings.TrimSuffix(p, ".jsonl"), "subagents")
}

// agentIDFromFile recovers the id from `agent-<id>.jsonl`, or "".
func agentIDFromFile(name string) string {
	if !strings.HasSuffix(name, ".jsonl") || !strings.HasPrefix(name, "agent-") {
		return ""
	}
	return strings.TrimSuffix(name, ".jsonl")
}

// Reconcile compares the seat ids the manifest names against the transcripts on
// disk. namedIDs are the `agent_id` values of `kind: "seat"` rows, WITHOUT the
// `agent-` prefix the filenames carry.
//
// namedIDs IS TREATED AS A SET, and a duplicate is collapsed rather than trusted.
// This is a set difference on both sides — `on` has been a map since the first
// version, because two files cannot share a name — and Named was the one side that
// counted repeats. It matters twice over: `len(Named)` was printed as a seat count
// beside a per-FILE disk count, and a seat captured twice that was also absent from
// disk appeared under MISSING twice. Callers should hand over SeatCensus.IDs, which
// is already distinct; the collapse here is so an exported function cannot be made
// to lie by a caller that assembles the slice itself.
//
// readDir is injected so both the present and the absent directory are reachable
// in a test regardless of what is on the machine running it.
func Reconcile(sessionTranscript string, namedIDs []string, readDir func(string) ([]os.DirEntry, error)) Coverage {
	c := Coverage{Dir: SubagentDir(sessionTranscript), OnDisk: []string{}, Named: []string{}, Unnamed: []string{}, Missing: []string{}}
	named := map[string]bool{}
	for _, id := range namedIDs {
		if id == "" || named["agent-"+id] {
			continue
		}
		named["agent-"+id] = true
		c.Named = append(c.Named, "agent-"+id)
	}
	sort.Strings(c.Named)

	if c.Dir == "" {
		c.Why = "the session row's transcript_path does not end in .jsonl, so the seat-transcript directory cannot be derived from it. Nothing was examined"
		return c
	}
	entries, err := readDir(c.Dir)
	if err != nil {
		// THE DISCRIMINATOR, and it is the whole reason this type has a Measured
		// field. A missing directory in a session that recorded no seats is
		// consistent and boring. A missing directory in a session that recorded
		// seats means the derivation is wrong, and reporting "nothing unnamed"
		// there would be the plausible zero this plugin exists to prevent.
		if len(c.Named) == 0 {
			c.Measured = true
			c.Why = "no seat rows and no seat-transcript directory — consistent, and nothing to reconcile"
			return c
		}
		c.Why = "the manifest names " + itoa(len(c.Named)) + " seat(s) but " + c.Dir +
			" could not be read (" + err.Error() + "). Either the harness's layout moved or this is the wrong directory; " +
			"reporting zero unnamed transcripts from here would be a guess wearing a measurement's clothes"
		return c
	}

	on := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if id := agentIDFromFile(e.Name()); id != "" {
			on[id] = true
			c.OnDisk = append(c.OnDisk, id)
		}
	}
	sort.Strings(c.OnDisk)

	for _, id := range c.OnDisk {
		if !named[id] {
			c.Unnamed = append(c.Unnamed, id)
		}
	}
	for _, id := range c.Named {
		if !on[id] {
			c.Missing = append(c.Missing, id)
		}
	}
	c.Measured = true
	return c
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// SeatCensus is what a manifest says about seats: the DISTINCT seats it names, and
// the facts a bare row count hides.
//
// # ROWS ARE NOT SEATS, because SubagentStop fires again for a seat that is continued
//
// Measured on this repository's own manifest (2026-09-04, trajectories-d628026a…):
// seat `a703a4ea4a2d4e09d` carries TWO rows, 07:08:23Z and 07:11:10Z, with the
// transcript grown from 356468 to 452844 bytes between them. One seat, continued,
// stopped a second time. Nothing in the capture path is wrong there — a row is an
// observation, and the second observation is true.
//
// What was wrong is that every reader downstream treated a row as a seat. Reconcile
// was handed the id twice, `Named` held it twice, and `coverage` printed a per-ROW
// count beside a per-FILE count and called the pair a reconciliation. A seat captured
// twice AND genuinely absent from disk would have been listed under MISSING twice.
//
// So idempotence lives HERE, at the read, and the manifest stays an append-only
// journal. The alternative — having the hook suppress a repeat capture — needs a
// read-modify-write on a path that must never cost a subagent its turn, races with
// every other seat stopping at the same moment, and throws away the one thing the
// second row actually says: this seat did more work. The repeat is REPORTED rather
// than erased.
type SeatCensus struct {
	// IDs is every distinct seat the manifest names, sorted. This is the denominator
	// a coverage claim may be stated over; Rows is not.
	IDs []string `json:"ids"`
	// Rows is how many seat rows produced those ids — always >= len(IDs).
	Rows int `json:"rows"`
	// Repeats maps a seat id to how many times it was captured, holding only the ids
	// captured more than once. Empty on a manifest where every seat stopped once.
	Repeats map[string]int `json:"repeats,omitempty"`
	// Unclassifiable counts seat rows carrying no agent_id: they name nothing and can
	// be reconciled with nothing. Counted, never folded into either side.
	Unclassifiable int `json:"unclassifiable"`
	// Unreadable counts lines that are not JSON, and it exists because skipping them
	// in silence is the wrong report.
	//
	// The manifest is appended to by a hook that can be killed mid-write, so a torn
	// tail is exactly what an interruption leaves behind. The seat row lost with it
	// then surfaces as UNNAMED — which reads as "the hook never fired for that seat",
	// a different fault with a different fix. Counting the unreadable lines is what
	// lets a reader tell the two apart, and 0 here is what makes UNNAMED mean what it
	// says. See appendRow in gray-area-capture for the writer-side half.
	Unreadable int `json:"unreadable"`
}

// Seats reads the census out of a manifest, counting every row that describes a real
// SEAT and returning the seats it names once each.
//
// # `kind` alone is not enough, and the first run proved it
//
// Schema 4 separates `seat` from `turn-end` (#189). Rows written before it called
// every `SubagentStop` a seat, and this repository's own manifest is mostly such
// rows. Filtering on `kind` alone reported **165 seats against 20 transcripts** and
// listed 146 phantom MISSING files — the exact miscount #189 closed, reproduced by
// the tool built to measure its consequences, on its first real run. The comment
// here predicted it and the code did it anyway.
//
// So a pre-schema-4 row is RE-CLASSIFIED from its own two fields, using the same
// conjunction schema 4 applies at capture: no `agent_type` AND not resolved is a
// turn end. Re-deriving beats discarding 90% of the manifest, and it is sound
// because the row carries both fields — this is not a guess about the row, it is
// the row's own evidence read the way the current model reads it.
//
// Unclassifiable is returned rather than folded into either side: a row that
// answers neither way must not quietly become a seat (a phantom MISSING) or
// quietly vanish (an understated denominator).
func Seats(manifest []byte) SeatCensus {
	c := SeatCensus{IDs: []string{}}
	captures := map[string]int{}
	for _, line := range strings.Split(string(manifest), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var r struct {
			Schema    int    `json:"schema"`
			Kind      string `json:"kind"`
			AgentID   string `json:"agent_id"`
			AgentType string `json:"agent_type"`
			Resolved  bool   `json:"resolved"`
		}
		if json.Unmarshal([]byte(line), &r) != nil {
			c.Unreadable++
			continue
		}
		switch r.Kind {
		case "seat":
		case "turn-end", "session":
			continue
		default:
			continue
		}
		if r.AgentID == "" {
			// A seat row with no id names nothing and can be reconciled with
			// nothing. Counted, never dropped.
			c.Unclassifiable++
			continue
		}
		if r.Schema < 4 && r.AgentType == "" && !r.Resolved {
			continue // a turn end, filed as a seat by a binary that had no other word
		}
		c.Rows++
		captures[r.AgentID]++
	}
	for id, n := range captures {
		c.IDs = append(c.IDs, id)
		if n > 1 {
			if c.Repeats == nil {
				c.Repeats = map[string]int{}
			}
			c.Repeats[id] = n
		}
	}
	sort.Strings(c.IDs)
	return c
}
