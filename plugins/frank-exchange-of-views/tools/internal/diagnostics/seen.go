package diagnostics

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

// SEEN is SURFACE EXPOSURE: the verbs of a seat's own role that appeared in help output the seat
// ACTUALLY RECEIVED. It is not what the seat used — that is Attempted — and conflating the two
// compares a new number against an old one measuring something else.
//
// THIS LIVED IN A SCRATCHPAD PYTHON SCRIPT THAT REGEXED THE PROBE'S RENDERED MARKDOWN, which is
// the defect this repository keeps deleting, committed by the instrument that measures it. Every
// hazard below was found the hard way there, and each is now a test rather than a comment in a
// file nothing runs.
type Seen struct {
	// Top are the role's DIRECT children that reached the seat. This saturates: one root --help
	// reveals every one of them, so per sitting it is all-or-nothing and the figure restates
	// "did this seat open help at all".
	Top map[string]bool
	// Leaf are the ACTS — subgroup children included — so a seat must open the subgroup to see
	// them. This does not saturate, and is the finer question.
	Leaf map[string]bool
	// Blocks is how many help listings were counted, and Rejected how many were read and
	// DISCARDED as not this role's. A run with listings rejected and none counted is a different
	// state from a run where the seat never opened help, and the two must not both report zero.
	Blocks, Rejected int
}

// listingLine matches one row of a cobra "Available Commands:" block.
//
// THE SEPARATOR IS ONE SPACE, NOT TWO. Cobra pads the name column to the width of the LONGEST
// name, which leaves exactly that name followed by a single space — so a `\s{2,}` separator drops
// precisely one verb per role, silently, from numerator and denominator alike. The ratio still
// looked like a measurement. The leading indent is what identifies a row; the gap after the name
// is not.
var listingLine = regexp.MustCompile(`^\s{2,}([a-z][a-z-]+) +(\S.*)$`)

const listingHeader = "Available Commands:"

// housekeeping are the commands cobra generates onto every parent. They belong to no role, so
// they neither credit a seat nor disqualify the listing they appear in.
var housekeeping = map[string]bool{"completion": true, "help": true}

// helpBlocks pulls every cobra command listing out of one text.
func helpBlocks(text string) [][]string {
	var out [][]string
	var cur []string
	inBlock := false
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), listingHeader) {
			inBlock, cur = true, nil
			continue
		}
		if !inBlock {
			continue
		}
		if strings.TrimSpace(line) == "" {
			if len(cur) > 0 {
				out = append(out, cur)
			}
			inBlock, cur = false, nil
			continue
		}
		if m := listingLine.FindStringSubmatch(line); m != nil {
			cur = append(cur, m[1])
		}
	}
	if len(cur) > 0 {
		out = append(out, cur)
	}
	return out
}

// ReadSeen counts the role's own verbs that reached the seat through help output.
//
// COUNTED PER BLOCK, NEVER PER LINE, and that is the difference between a measurement and a
// flattering one. The probe stages ROOT help into a board as setup material, and the root listing
// contains `friction` and `count-claims` — names that are also role verbs. A line-level match
// credits every seat with those before it acts, and flatters exactly the arms that open help
// LEAST, which is the direction the whole experiment is testing. A block counts only when every
// name in it belongs to this role.
func ReadSeen(trajectoryPath string, acts []string) (Seen, error) {
	seen := Seen{Top: map[string]bool{}, Leaf: map[string]bool{}}

	own := map[string]bool{}    // every act, by its last word
	direct := map[string]bool{} // the role's direct children
	for _, v := range acts {
		f := strings.Fields(v)
		own[f[len(f)-1]] = true
		direct[f[0]] = true
	}

	f, err := os.Open(trajectoryPath)
	if err != nil {
		return seen, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var ev struct {
			Message struct {
				Content []struct {
					Type    string          `json:"type"`
					Content json.RawMessage `json:"content"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		for _, blk := range ev.Message.Content {
			if blk.Type != "tool_result" {
				continue
			}
			for _, block := range helpBlocks(resultText(blk.Content)) {
				// EVERY name in the block must be this role's, or the block is not this role's
				// listing and contributes nothing.
				//
				// EXCEPT COBRA'S OWN HOUSEKEEPING, and leaving that out cost the first real
				// reading: `completion` and `help` are generated onto every command, so one of
				// them appearing in a seat's OWN listing failed the all-names test and rejected
				// the block whole. The seat had been handed its entire surface and the measure
				// said it had seen nothing — a false zero produced by the rule written to
				// prevent false credit, which is the same defect facing the other way.
				all := true
				for _, name := range block {
					if housekeeping[name] {
						continue
					}
					if !own[name] && !direct[name] {
						all = false
						break
					}
				}
				if !all {
					seen.Rejected++
					continue
				}
				seen.Blocks++
				for _, name := range block {
					if direct[name] {
						seen.Top[name] = true
					}
					if own[name] {
						seen.Leaf[name] = true
					}
				}
			}
		}
	}
	return seen, sc.Err()
}

// resultText flattens a tool_result's content, which the transcript writes either as a bare
// string or as a list of typed blocks depending on the tool.
func resultText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var blocks []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var b strings.Builder
	for _, blk := range blocks {
		b.WriteString(blk.Text)
		b.WriteString("\n")
	}
	return b.String()
}
