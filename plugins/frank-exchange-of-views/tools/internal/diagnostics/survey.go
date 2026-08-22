package diagnostics

// DID THE SEAT SURVEY ITS SURFACE, OR CONFIRM A CHOICE IT HAD ALREADY MADE?
//
// SEEN answers surface EXPOSURE and Attempted answers USE. Neither answers the question that
// decides whether the help is working: when the seat reached for a command, had it read that
// command's own help FIRST — and did it ever look at anything it did not go on to use?
//
// Those are different questions and the second is the one that matters. A rule that made a seat
// read `--help` for the verb it had already picked would raise flag accuracy and change nothing
// about whether it considered the alternatives; the failure being hunted is a seat that decides
// from memory and then confirms.
//
// # Every counting choice here is a place this could return a flattering zero
//
//	STRICTLY BEFORE   only help read before a command's FIRST invocation counts. A lens read
//	                  `fetch --help` immediately AFTER being refused, and "read at any point"
//	                  would score that as compliance — forgiving precisely the behaviour measured.
//	TOP IS NOT READ   the root listing names every verb and gives no flags. Counting it as having
//	                  read a command's help makes almost every first use compliant, and the metric
//	                  reports a clean board on a sitting that guessed six flags.
//	DEPTHS APART      group help (step 2) and command help (step 3) are separate steps and both
//	                  were skipped. Merged into one bucket, an intervention that moves one and not
//	                  the other is invisible.
//	MISS IS NOT ZERO  a trajectory with no invocations at all reports NotAsked, never a 0% blind
//	                  rate, because "the seat ran nothing" and "the seat ran everything having
//	                  read first" are opposite findings that a bare ratio merges.

import (
	"bufio"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

// Depth is how much help a seat had for a command before it first ran it.
type Depth string

const (
	// DepthNone: nothing. Not even the root listing reached the seat first.
	DepthNone Depth = "none"
	// DepthTop: the command's name appeared in a listing the seat received — it knew the verb
	// existed. It did NOT have the flags, which is where the guesses happen.
	DepthTop Depth = "top"
	// DepthGroup: the seat opened the group the command sits under (step 2).
	DepthGroup Depth = "group"
	// DepthCommand: the seat opened the command's own help (step 3), which is the only depth that
	// states the flags and which are required.
	DepthCommand Depth = "command"
)

// FirstUse is one command the seat invoked, and what it knew when it first did.
type FirstUse struct {
	Command string `json:"command"`
	Depth   Depth  `json:"depth"`
	// Call is the 1-based index of that first invocation among the seat's tool calls, so a
	// sitting that surveys up front is distinguishable from one that surveys as it goes.
	Call int `json:"call"`
	// Refused records whether that first invocation came back an error.
	Refused bool `json:"refused"`
}

// Survey is the sitting's answer.
type Survey struct {
	// Asked is false when the seat invoked no command at all. Every ratio below is meaningless
	// then, and saying so is not the same as reporting zero blind uses.
	Asked     bool       `json:"asked"`
	FirstUses []FirstUse `json:"firstUses"`
	// HelpPages are the distinct help pages the seat opened, by the command path each names —
	// "" for the root. This is the survey breadth: a seat that opened three pages considered at
	// most three commands in any detail, whatever the root listing showed it.
	HelpPages []string `json:"helpPages"`
	// PagesNeverUsed counts help pages the seat opened for commands it did NOT go on to run.
	// A seat that only ever opens the help of verbs it already chose is confirming, not
	// surveying, and this is the number that separates the two.
	PagesNeverUsed int `json:"pagesNeverUsed"`
	// LastHelpCall is the index of the seat's final help read, and TotalCalls the length of the
	// sitting. Help concentrated entirely in the opening calls is the "read once at the top and
	// never return" shape run 5 measured.
	LastHelpCall int `json:"lastHelpCall"`
	TotalCalls   int `json:"totalCalls"`
	// Traversal is how the seat MOVED across the top level of its own tree — the measurement
	// behind "are the groupings leaky".
	Traversal Traversal `json:"traversal"`
	// BashCalls is every Bash call in the trajectory, whether or not it invoked the tool.
	//
	// IT EXISTS TO MAKE ONE FAILURE LOUD. Everything above is keyed on recognising the tool by its
	// basename; hand it the wrong name and no call matches, TotalCalls is 0, Asked is false, and
	// the report reads "this seat invoked nothing" — which is also what a seat that sat on its
	// hands produces. BashCalls > 0 with TotalCalls == 0 can only be a name mismatch.
	BashCalls int `json:"bashCalls"`
}

// Traversal is the seat's movement across the top level of its own tree, over the commands it
// actually RAN — help reads are excluded, because surveying a surface is not an operation on it.
//
// THE QUESTION IT ANSWERS. A command tree groups verbs by what they are ABOUT, and the grouping
// is only earning its keep if the work a seat actually does stays inside one group at a time. If
// an ordinary operation has to hop `motion` -> `show` -> `motion` to complete, those two are
// coupled in practice however they are filed, and the boundary is in the wrong place.
//
// ADJACENCY, NOT SESSIONS, AND THAT IS A REAL LIMIT. A trajectory does not say where one logical
// operation ends and the next begins, so a crossing here is "the seat's next act was in a
// different group", which over-counts a seat that simply finished one job and started another.
// The number to read is not the raw total but the PAIRS: one pair dominating the crossings is a
// coupling, an even spread is a seat working through its docket.
type Traversal struct {
	// Sequence is the top-level word of each invocation in order, with consecutive repeats
	// collapsed — so `show show show motion` reads `show motion`. Legibility, not data; Pairs
	// carries the countable half.
	Sequence []string `json:"sequence"`
	// Crossings and Stays split adjacent invocation pairs by whether the top-level word changed.
	// Both are needed: 20 crossings means nothing without knowing whether the seat made 25 calls
	// or 250.
	Crossings int `json:"crossings"`
	Stays     int `json:"stays"`
	// Unclassified counts invocations the traversal could not place in any group, because
	// CommandWords found no verb path in them.
	//
	// IT EXISTS BECAUSE ZERO CROSSINGS IS THE ANSWER TO TWO DIFFERENT QUESTIONS. A seat that
	// worked entirely inside one group and a seat whose every command was unreadable both report
	// 0 and 0, and the first time this ran over a real corpus it was the second: red-merge-r1
	// made thirty recognised tool calls and one was classified, so the projection said "never
	// left a group" about a seat it could not see. Read `crossings` only when this is small
	// relative to them.
	Unclassified int `json:"unclassified"`
	// Pairs are the crossings by (from, to), most frequent first, then alphabetically so the
	// order is stable across runs and a diff of two runs is readable.
	Pairs []TraversalPair `json:"pairs"`
}

// TraversalPair is one ordered group-to-group move and how often the seat made it.
type TraversalPair struct {
	From string `json:"from"`
	To   string `json:"to"`
	N    int    `json:"n"`
}

// ToolUnrecognised reports the one state that would otherwise pass as an idle sitting: the seat
// ran shell commands and none of them were recognised as the tool.
func (s Survey) ToolUnrecognised() bool { return s.BashCalls > 0 && s.TotalCalls == 0 }

// BlindFirst counts commands first invoked with nothing better than the root listing.
func (s Survey) BlindFirst() int {
	n := 0
	for _, u := range s.FirstUses {
		if u.Depth == DepthNone || u.Depth == DepthTop {
			n++
		}
	}
	return n
}

// RefusedBlind and RefusedRead split the refusals by whether the seat had read the command's own
// help first. The interviews claim every refusal was a guessed flag; this is what settles it.
func (s Survey) RefusedBlind() (blind, read int) {
	for _, u := range s.FirstUses {
		if !u.Refused {
			continue
		}
		if u.Depth == DepthNone || u.Depth == DepthTop {
			blind++
		} else {
			read++
		}
	}
	return blind, read
}

type call struct {
	command string
	isHelp  bool
	path    []string
	result  string
	errored bool
}

// ReadSurvey walks a trajectory in order and answers the survey questions.
//
// binName is the basename the tool was built under; a Bash call that does not invoke it is not a
// command and cannot be a help read either.
func ReadSurvey(trajectoryPath, binName string) (Survey, error) {
	var out Survey
	f, err := os.Open(trajectoryPath)
	if err != nil {
		return out, err
	}
	defer f.Close()

	// One pass to collect calls in order, pairing each tool_use with its tool_result so a refusal
	// is attributable to the invocation that earned it.
	type pending struct{ idx int }
	var calls []call
	byID := map[string]pending{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var ev struct {
			Message struct {
				Content []struct {
					Type      string `json:"type"`
					Name      string `json:"name"`
					ID        string `json:"id"`
					ToolUseID string `json:"tool_use_id"`
					IsError   bool   `json:"is_error"`
					Content   any    `json:"content"`
					Input     struct {
						Command string `json:"command"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		for _, blk := range ev.Message.Content {
			switch blk.Type {
			case "tool_use":
				if blk.Name != "Bash" {
					continue
				}
				out.BashCalls++
				fields := BinFields(blk.Input.Command, binName)
				if fields == nil {
					continue
				}
				c := call{command: blk.Input.Command, path: CommandWords(fields)}
				for _, t := range fields {
					if t == "--help" || t == "-h" {
						c.isHelp = true
						break
					}
				}
				byID[blk.ID] = pending{idx: len(calls)}
				calls = append(calls, c)
			case "tool_result":
				p, ok := byID[blk.ToolUseID]
				if !ok {
					continue
				}
				calls[p.idx].errored = blk.IsError
				calls[p.idx].result = textOf(blk.Content)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return out, err
	}

	out.TotalCalls = len(calls)
	seenPage := map[string]bool{} // help pages opened, by command path
	namesSeen := map[string]bool{}
	firstSeen := map[string]bool{}
	for i, c := range calls {
		key := strings.Join(c.path, " ")
		if c.isHelp {
			if !seenPage[key] {
				seenPage[key] = true
				out.HelpPages = append(out.HelpPages, key)
			}
			out.LastHelpCall = i + 1
			// Names revealed by this listing are known-to-exist, not known-how-to-run.
			for _, n := range listedNames(c.result) {
				namesSeen[strings.TrimSpace(strings.Join(append(append([]string{}, c.path...), n), " "))] = true
				namesSeen[n] = true
			}
			continue
		}
		if len(c.path) == 0 || firstSeen[key] {
			continue
		}
		firstSeen[key] = true
		out.Asked = true
		out.FirstUses = append(out.FirstUses, FirstUse{
			Command: key, Depth: depthFor(c.path, seenPage, namesSeen), Call: i + 1, Refused: c.errored,
		})
	}
	for _, p := range out.HelpPages {
		if p != "" && !firstSeen[p] {
			out.PagesNeverUsed++
		}
	}
	out.Traversal = traversalOf(calls)
	return out, nil
}

// depthFor answers what the seat had for this command path before it ran it.
//
// The command's OWN page beats its group's, which beats a mention in a listing. Checked in that
// order so the strongest true answer is the one reported.
func depthFor(path []string, pages, names map[string]bool) Depth {
	key := strings.Join(path, " ")
	if pages[key] {
		return DepthCommand
	}
	for i := len(path) - 1; i >= 1; i-- {
		if pages[strings.Join(path[:i], " ")] {
			return DepthGroup
		}
	}
	if names[key] || names[path[len(path)-1]] {
		return DepthTop
	}
	return DepthNone
}

// listedNames pulls the command names out of any cobra listing in a help page, reusing the same
// block reader SEEN uses so the two agree about what a listing is.
func listedNames(text string) []string {
	var out []string
	for _, blk := range helpBlocks(text) {
		for _, n := range blk {
			if !housekeeping[n] {
				out = append(out, n)
			}
		}
	}
	return out
}

func textOf(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		var b strings.Builder
		for _, x := range t {
			if m, ok := x.(map[string]any); ok {
				if s, ok := m["text"].(string); ok {
					b.WriteString(s)
					b.WriteString("\n")
				}
			}
		}
		return b.String()
	}
	return ""
}

// traversalOf measures movement across the top level of the tree, over invocations only.
//
// A ROOT-LEVEL VERB IS ITS OWN TOP. `mint` and `show board` both contribute a top-level word —
// `mint` and `show` — because the question is which part of the surface the seat was working in,
// and a verb that sits at the root is a part of the surface. Filing them differently would make
// a seat that never opens a group look like it never moved.
func traversalOf(calls []call) Traversal {
	// EMPTY IS `[]`, NOT `null`. A nil slice marshals as null, and a consumer that iterates the
	// pairs of every seat then crashes on the first seat that never left one group — which is an
	// ordinary sitting, not an error. Caught driving this over a real run: blue-synthesize
	// returned null and the reader died on it.
	t := Traversal{Sequence: []string{}, Pairs: []TraversalPair{}}
	pairs := map[TraversalPair]int{}
	prev := ""
	for _, c := range calls {
		if c.isHelp {
			continue
		}
		if len(c.path) == 0 {
			t.Unclassified++
			continue
		}
		top := c.path[0]
		if prev == "" {
			t.Sequence = append(t.Sequence, top)
			prev = top
			continue
		}
		if top == prev {
			t.Stays++
			continue
		}
		t.Crossings++
		pairs[TraversalPair{From: prev, To: top}]++
		t.Sequence = append(t.Sequence, top)
		prev = top
	}
	for p, n := range pairs {
		p.N = n
		t.Pairs = append(t.Pairs, p)
	}
	// MOST FREQUENT FIRST, THEN ALPHABETICAL. Map iteration order is random, and a projection
	// whose rows move between identical runs cannot be diffed — which is the whole use for this.
	sort.Slice(t.Pairs, func(i, j int) bool {
		a, b := t.Pairs[i], t.Pairs[j]
		if a.N != b.N {
			return a.N > b.N
		}
		if a.From != b.From {
			return a.From < b.From
		}
		return a.To < b.To
	})
	return t
}
