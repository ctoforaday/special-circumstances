package seatprobe

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// WHAT THE RECORD CANNOT WITNESS.
//
// Read() counts verbs off the event log, which is the right authority for anything that WRITES.
// It is blind by construction to everything else, and the blindness is not neutral: `show` is a
// read, records no event, and so appeared under "NEVER REACHED" on all nine boards of the first
// separated run — while the trajectories showed eight of those nine seats calling it, up to nine
// times each. The denominator counted it as available and the numerator could never count it as
// used, so every "used N of M" that run produced was wrong in both terms.
//
// That is this project's recurring shape one more time, and this time in the instrument: the
// absence of an event and the absence of the act are the same bytes, and the report chose the
// reading that flattered it. A refused call leaves no event either — so "never reached" was also
// covering seats that reached for a verb and were turned away.
//
// So attempts come from the TRAJECTORY, which records what the seat actually typed, and the two
// sources answer different questions worth keeping apart:
//
//	recorded   the act landed — the board changed
//	attempted  the seat reached for it — whether it landed, was read-only, or was refused
//
// The DIFFERENCE is the interesting column. A verb attempted and never recorded is either a read
// (fine) or a refusal (a defect), and both were invisible before.

// Attempted counts the verbs a seat invoked, read out of a captured trajectory.
//
// binName is the basename the harness built the tool under; a command counts only if it invokes
// that binary, so the seat's own `ls`, `grep` and `cat` do not become phantom verbs.
func Attempted(trajectoryPath, binName string, sf Surface, role string) (map[string]int, error) {
	f, err := os.Open(trajectoryPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	verbs := map[string]bool{}
	for _, v := range sf.Verbs(role) {
		verbs[v] = true
	}
	out := map[string]int{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var ev struct {
			Message struct {
				Content []struct {
					Type  string `json:"type"`
					Name  string `json:"name"`
					Input struct {
						Command string `json:"command"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		for _, blk := range ev.Message.Content {
			if blk.Type != "tool_use" || blk.Name != "Bash" {
				continue
			}
			if v := verbIn(blk.Input.Command, binName, verbs); v != "" {
				out[v]++
			}
		}
	}
	return out, sc.Err()
}

// verbIn finds the verb a shell command invokes, or "".
//
// THE LONGEST MATCHING PREFIX WINS, and that is not fussiness: `motion grade file` and `motion`
// are both real, and taking the first token would file every motion under a verb no expectation
// is written against. The same shortcut in the fuzzer's trajectory reader once reported all seven
// motion verbs as never invoked while 35 motion events sat in the record.
func verbIn(command, binName string, verbs map[string]bool) string {
	fields := strings.Fields(strings.NewReplacer("\"", " ", "'", " ").Replace(command))
	start := -1
	for i, f := range fields {
		base := strings.TrimSuffix(strings.ToLower(filepath.Base(f)), ".exe")
		if base == strings.TrimSuffix(strings.ToLower(binName), ".exe") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	// THE VERB IS NOT ALWAYS THE FIRST TOKEN, and assuming it was cost this counter a whole
	// board. One probed seat wrote its flags first —
	//
	//	fxr.exe --run "." --seat-id red-lens-r1-L1 lens finding --key F1 ...
	//
	// — so a scan that stopped at the first dash found nothing and reported a seat with 22 tool
	// calls as having reached for no verb at all. Which is precisely the reading this file exists
	// to eliminate, one level up, in the instrument written to eliminate it.
	//
	// A token following a flag is that flag's VALUE and can never be the verb. Dropping those is
	// what stops `--seat-id red-lens-r1-L1` and `--reason "close it"` from nominating one.
	var cand []string
	value := false
	for _, f := range fields[start:] {
		if strings.HasPrefix(f, "-") {
			value = true
			continue
		}
		if value {
			value = false // a flag's argument
			continue
		}
		// The role is not part of the verb: the surface names verbs role-relatively (`show`,
		// not `blue show`), because that is how an expectation is written.
		switch f {
		case "blue", "lens", "merge", "bench":
			continue
		}
		cand = append(cand, f)
	}
	for i := range cand {
		for n := min(3, len(cand)-i); n >= 1; n-- {
			if v := strings.Join(cand[i:i+n], " "); verbs[v] {
				return v
			}
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
