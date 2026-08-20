package seatprobe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// HOW MUCH OF THE SURFACE THE SEAT IS TOLD, AS A VARIABLE RATHER THAN A CONSTANT.
//
// # The confound this exists to remove
//
// menu.go states, as the basis for making the REFUSAL the teaching channel: "MEASURED across nine
// probe sittings: seats do not learn this tool from `--help`. Every one of them read it once or
// twice in twenty to forty tool calls." That observation is real. What it cannot support is the
// conclusion drawn from it, because every one of those nine sittings ran with a PARTIAL list of
// verbs already in front of the seat — and a partial list is a plausible answer to the question
// `--help` answers completely.
//
// Measured 2026-08-15, counting distinct `<role> <verb>` forms in the constitutions the probe
// actually dispatches under (cmd/seatprobe.constitutionFor):
//
//	blue   blue-researcher.md   2 named   of 18 reachable
//	bench  lead-judge.md        2 named   of 11 reachable
//	merge  red-auditor.md       4 named   of 16 reachable
//	lens   red-auditor.md       1 named   of  9 reachable
//
// The acting-arm prompt itself (cmd/seatprobe.dispatch) names ZERO verbs. So the seat's entire
// nomination set was those two-to-four names, and the measured result was a seat using about four
// verbs. "Seats do not read `--help`" and "seats stop when the partial list runs out" produce that
// same number and want opposite fixes — the first is a fact about seats, the second is a fact
// about the constitution, and nothing in nine sittings could tell them apart.
//
// # Why this is a treatment and not a rewrite
//
// The `none` arm is produced by REDACTING the real constitution rather than by writing a new one.
// A hand-authored no-names constitution differs from the original in its prose as well as in its
// naming, and the prose would then be the uncontrolled variable — which is the same error one
// level up from the one being corrected. Redaction changes exactly one thing: whether the NAME is
// present. The situation binding around it ("a grade moves through …, and only through it") stays,
// because that clause is what a constitution is FOR and removing it would test a different
// question.
//
// # And the treatment is measured, not assumed
//
// NamesSurviving counts live verb names still in the text after an arm is applied. A redactor whose
// pattern matches nothing yields a `none` arm byte-identical to `partial`, both arms then report
// the same behaviour, and the experiment concludes "naming does not matter" — a null result
// manufactured by the instrument. That is this repository's recurring shape, and it would arrive
// here wearing the clothes of a finding. So the count is printed with the result, and the tests
// refuse a redaction that removed nothing from an input that named something.

// Constitution is the seat's system prompt: THE SHIPPED BYTES, unmodified.
//
// It used to assemble one of three arms — withhold the verb names, name a handful, name them all
// — plus a fourth axis that added or removed the surface-discovery directive. Those are gone. The
// question they answered is settled: the constitutions name no verb and carry the directive, and
// keeping the alternatives dispatchable made the probe an instrument for re-litigating a decision
// instead of one for measuring what ships.
//
// The shipped file is never modified, and that still matters: an experiment that edits the
// artifact it is measuring cannot be run twice.
func Constitution(src []byte) []byte { return src }

// NamesSurviving counts the role verbs a constitution NAMES, and it is the gate behind the claim
// that they name none.
//
// ONE RULE NOW, WHERE THERE WERE TWO. A name in a GENERATED block used to be counted structurally,
// because the arms rendered one verb per indented line under their own heading. No arm renders
// anything any more, so that half read a shape nothing produces.
//
// A name WRITTEN INTO PROSE has to be marked as an invocation by something or it cannot be told
// from English: `close`, `verify`, `position`, `finding` and `show` are ordinary words, and making
// the role prefix optional once took the bench constitution from 0 surviving names to 11, every one
// of them prose. The role in front is that mark, and it stays required even though a seat no longer
// types it — this counts the SPELLING a hand-written catalogue uses, not an invocation.
func NamesSurviving(text string, sf Surface) map[string]int {
	out := map[string]int{}
	seen := map[string]bool{}
	for _, role := range Roles {
		for _, v := range sf.Verbs(role) {
			if seen[v] {
				continue
			}
			seen[v] = true
			n := len(regexp.MustCompile(`\b(lens|merge|blue|bench)\s+`+regexp.QuoteMeta(v)+`\b`).FindAllString(text, -1))
			if n > 0 {
				out[v] = n
			}
		}
	}
	return out
}

// HelpUse is what a trajectory says about the seat's reading of `--help`.
//
// THE DEPENDENT VARIABLE THE ORIGINAL FINDING WAS ABOUT. "Seats read it once or twice in twenty to
// forty tool calls" is a claim about COUNT and about POSITION, and neither was ever extracted
// mechanically — it was read off trajectories by hand, once. Both are in the capture already.
type HelpUse struct {
	// Calls is how many times the seat asked the tool for help.
	Calls int
	// BinCalls is how many times it invoked the tool at all — the denominator that makes Calls
	// mean anything.
	BinCalls int
	// FirstHelpAt is the 1-based index, among tool invocations, of the first help read; 0 when
	// the seat never asked. BEFORE-OR-AFTER-THE-FIRST-ACT is the interesting half: a seat that
	// reads help at call 1 was orienting, and one that reads it at call 9 was recovering.
	FirstHelpAt int
	// Refusals counts invocations that came back as an error the tool authored. A seat that
	// learns from refusals rather than from help should show these leading its help reads.
	Refusals int
}

// helpFlag matches an explicit request for the surface, and nothing else. `-h` is included because
// cobra accepts it; a bare `help` subcommand is not, because the role groups answer it as a verb.
var helpFlag = regexp.MustCompile(`(^|\s)(--help|-h)(\s|$)`)

// ReadHelpUse extracts the help-reading behaviour from a captured trajectory.
//
// binName scopes it to the tool, exactly as Attempted does: the seat's own `grep --help` is not a
// read of this surface, and counting it would inflate the number the whole experiment turns on.
func ReadHelpUse(trajectoryPath, binName string) (HelpUse, error) {
	var u HelpUse
	f, err := os.Open(trajectoryPath)
	if err != nil {
		return u, err
	}
	defer f.Close()

	want := strings.TrimSuffix(strings.ToLower(binName), ".exe")
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var ev struct {
			Message struct {
				Content []struct {
					Type    string `json:"type"`
					Name    string `json:"name"`
					IsError bool   `json:"is_error"`
					Input   struct {
						Command string `json:"command"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		for _, blk := range ev.Message.Content {
			if blk.Type == "tool_result" && blk.IsError {
				u.Refusals++
				continue
			}
			if blk.Type != "tool_use" || blk.Name != "Bash" {
				continue
			}
			cmd := blk.Input.Command
			if !invokesBin(cmd, want) {
				continue
			}
			u.BinCalls++
			if helpFlag.MatchString(cmd) {
				u.Calls++
				if u.FirstHelpAt == 0 {
					u.FirstHelpAt = u.BinCalls
				}
			}
		}
	}
	return u, sc.Err()
}

// ViewReads is which projections a seat actually opened.
//
// # Why a naming experiment has to measure this
//
// THE DUTY LIST IS A THIRD CHANNEL, AND IT WAS UNCONTROLLED. record.SittingOf derives a
// Duty{What, How} for each live circumstance — the situation, its consequence, and the exact
// command that discharges it — which is the same fact the constitution's naming carries, arriving
// only when it applies. It rides on ONE projection: `work list`. `show board` does not carry it.
//
// Measured 2026-08-15, across the 24 dispatches of the first naming matrix:
//
//	arm             work list reads/cell   board reads/cell   surface reached
//	none                   2.00                 4.00              8.83
//	partial                0.67                 3.17              6.83
//	complete               0.33                 4.33              8.33
//	none+directive         1.83                 2.67              7.50
//
// So the channel did not merely go unmeasured — it CO-VARIED WITH THE TREATMENT, threefold, in an
// experiment whose whole claim was about how a seat learns its surface. The reported `none` over
// `partial` advantage sits on top of that difference and cannot be attributed to the constitution.
// A covariate that moves with the arm is not a constant to be waved off; it is a rival explanation
// wearing the result's clothes.
//
// It is NOT a clean mediator either, and this must not be swapped for the opposite over-claim:
// `complete` had the FEWEST work list reads and the second-highest reach, so "reading the work list
// raises reach" is refuted by the same table that raises the confound.
//
// The instrument's job is to stop any future run reporting a naming effect without the channel that
// competes with it. Hence this, printed beside HelpUse on every probe report.
//
// AND THE DESIGN FINDING IS INDEPENDENT OF THE EXPERIMENT: `board` is described in this tool's own
// words as "the form a seat acts on" and is read 2.7–4.3 times a sitting; `work list` carries every
// duty and is read 0.33–2.00 times. The one channel that delivers situation-plus-verb at the moment
// it applies is the one the tool steers seats away from.
type ViewReads struct {
	// ByView counts each projection the seat opened. A bare `show` resolves to the role's
	// default, which is `work list` for every role, and is counted as such.
	ByView map[string]int
	// Work is the seat's one work-list read; Board is the one seats reach for instead.
	Work, Board, Total int
}

// showCall matches a projection read. The bare form (`<role> show` with flags or nothing after it)
// is the seat's pending work — the work list — so a token starting with a dash is not a view name.
var showCall = regexp.MustCompile(`(?:lens|merge|blue|bench)\s+show(?:\s+([a-z][a-z-]*))?`)

// ReadViewReads extracts which projections were opened, from a captured trajectory.
func ReadViewReads(trajectoryPath, binName string) (ViewReads, error) {
	v := ViewReads{ByView: map[string]int{}}
	f, err := os.Open(trajectoryPath)
	if err != nil {
		return v, err
	}
	defer f.Close()

	want := strings.TrimSuffix(strings.ToLower(binName), ".exe")
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
			cmd := blk.Input.Command
			if !invokesBin(cmd, want) {
				continue
			}
			m := showCall.FindStringSubmatch(cmd)
			if m == nil {
				continue
			}
			view := m[1]
			if view == "" {
				// The bare form is the role default, and that default is the work list for
				// every role. Counting it as "unknown" would undercount the one channel this
				// measurement exists for.
				view = "work"
			}
			v.ByView[view]++
			v.Total++
			switch view {
			case "work":
				v.Work++
			case "board":
				v.Board++
			}
		}
	}
	return v, sc.Err()
}

// Line renders the duty-delivery result for the probe report.
func (v ViewReads) Line() string {
	if v.Total == 0 {
		return "no projection opened at all — the duty list reached this seat through no channel"
	}
	return fmt.Sprintf("work list×%d (the ONLY carrier of the duty list), board×%d, %d projection reads total",
		v.Work, v.Board, v.Total)
}

func invokesBin(command, want string) bool {
	for _, f := range strings.Fields(strings.NewReplacer("\"", " ", "'", " ").Replace(command)) {
		if strings.TrimSuffix(strings.ToLower(filepath.Base(f)), ".exe") == want {
			return true
		}
	}
	return false
}

// Line renders one arm's result as a single comparable row. The arms are only worth anything
// side by side, so the harness prints them in one table rather than one report per run.
func (u HelpUse) Line() string {
	at := "never"
	if u.FirstHelpAt > 0 {
		at = fmt.Sprintf("call %d", u.FirstHelpAt)
	}
	return fmt.Sprintf("help×%d of %d tool calls (first: %s), refusals %d", u.Calls, u.BinCalls, at, u.Refusals)
}
