package seatprobe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
type Naming string

const (
	// NamingPartial names a HANDFUL of the role's verbs and stops — the shape a hand-kept list in
	// a constitution has.
	//
	// IT WAS THE STATUS QUO AND IS NOW A TREATMENT, because the finding landed: the constitutions
	// and the orchestrator's prompts name no verbs at all, so "the constitution exactly as it
	// ships" IS the `none` arm and drawing `partial` from the same bytes would make the two
	// identical. Constructing it keeps the experiment re-runnable as a regression check on the
	// result rather than retiring it into a claim nobody can re-measure.
	NamingPartial Naming = "partial"
	// NamingNone withholds every verb NAME while keeping every situation that calls for one. It is
	// the shipped configuration, and Redact is a belt-and-braces pass over it rather than the
	// treatment it once was.
	NamingNone Naming = "none"
	// NamingComplete states the whole role surface, GENERATED from the command tree.
	NamingComplete Naming = "complete"
)

// NamingArms is every arm, in ladder order — none, partial, complete — which is the order the
// report reads in and the order the hypothesis is stated in.
var NamingArms = []Naming{NamingNone, NamingPartial, NamingComplete}

// ParseNaming resolves an arm name, listing the real ones on a miss rather than defaulting
// silently: an unrecognised arm that quietly ran as `partial` would report a treatment that never
// happened.
func ParseNaming(s string) (Naming, error) {
	for _, a := range NamingArms {
		if Naming(s) == a {
			return a, nil
		}
	}
	var have []string
	for _, a := range NamingArms {
		have = append(have, string(a))
	}
	return "", fmt.Errorf("no naming arm %q — one of %s", s, strings.Join(have, ", "))
}

// HelpDirective is the instruction production carries and the probe never did.
//
// THE SECOND UNCONTROLLED DIFFERENCE. skills/research-protocol/scripts/debate.js tells every seat
// "YOUR CONTRACT IS `feov-record <role> --help` and each verb's own --help … READ THAT LIST BEFORE
// YOUR FIRST ACT, EVERY SITTING". The probe's acting prompt says nothing of the kind. So the
// configuration that was measured exists in no real run, and the configuration that ships has never
// been probed. Making the directive a flag rather than folding it into the arms keeps the two
// questions separable: does NAMING carry the surface, and does the DIRECTIVE carry it.
const HelpDirective = `
BEFORE YOUR FIRST ACT, run the record tool with --help for your role and read the whole list, then
each verb's own --help before you use it. What is listed you may do; what is not listed does not
exist for you. A name you did not read in the help this sitting is a guess.
`

// generatedHeading opens the complete arm's block. It is matched on to keep the arms composable
// (a complete arm is built from the partial text, so the block must be identifiable).
const generatedHeading = "YOUR COMPLETE VERB SURFACE"

// Constitution renders one arm's system prompt from the shipped constitution.
//
// `src` is the real file's bytes; the caller writes the result to a temp file and passes it as
// --system-prompt-file. The shipped file is never modified: an experiment that edits the artifact
// it is measuring cannot be run twice.
func Constitution(src []byte, sf Surface, role string, arm Naming, directive bool) []byte {
	text := string(src)
	switch arm {
	case NamingNone:
		text = Redact(text, sf)
	case NamingPartial:
		text += "\n\n" + PartialSurfaceBlock(sf, role)
	case NamingComplete:
		text += "\n\n" + CompleteSurfaceBlock(sf, role)
	}
	if directive {
		text += "\n" + HelpDirective
	}
	return []byte(text)
}

// CompleteSurfaceBlock is the whole role surface, GENERATED.
//
// Generated rather than written, and that is the point of the arm: if stating the complete surface
// works as well as or better than the partial hand-kept naming, the answer is not a gate over two
// hand-written copies but a carrier produced from the tree — which is what [[facts-are-fields]]
// says to prefer, and it would retire TestEveryRecordingVerbIsNamedInAPrompt rather than repair it.
func CompleteSurfaceBlock(sf Surface, role string) string {
	verbs := sf.Verbs(role)
	var b strings.Builder
	fmt.Fprintf(&b, "%s (generated from the tool's command tree — these %d are the whole set):\n\n", generatedHeading, len(verbs))
	for _, v := range verbs {
		fmt.Fprintf(&b, "  %s %s\n", role, v)
	}
	b.WriteString("\nEach verb's own `--help` carries its flags and what it is for. Nothing outside this list exists for you.\n")
	return b.String()
}

// PartialNamed is how many verbs the partial arm names.
//
// THREE, because that is the size the hand-kept lists actually were when they existed: measured
// over the four dispatched constitutions, 2 of 18 reachable for blue, 2 of 11 for the bench, 4 of
// 16 for the merge and 1 of 9 for a lens. The arm reproduces the SHAPE of the defect — a few names
// in front of the seat and the rest unmentioned — not any particular historical list.
const PartialNamed = 3

// PartialSurfaceBlock names the first few of the role's verbs and stops, which is what a hand-kept
// list in a constitution does.
//
// TAKEN FROM THE TREE IN ITS OWN ORDER, so the arm cannot drift from the surface it is a subset
// of. Which three is arbitrary and stated as arbitrary: the treatment under test is the PARTIALITY
// — a plausible answer to the question --help answers completely — not the identity of the verbs.
func PartialSurfaceBlock(sf Surface, role string) string {
	verbs := sf.Verbs(role)
	if len(verbs) > PartialNamed {
		verbs = verbs[:PartialNamed]
	}
	var b strings.Builder
	b.WriteString("THE VERBS YOU WILL MOSTLY NEED:\n\n")
	for _, v := range verbs {
		fmt.Fprintf(&b, "  %s %s\n", role, v)
	}
	return b.String()
}

// Redact removes verb NAMES while leaving the situations that call for them.
//
// WHAT IT REACHES, STATED SO THE GAP IS NOT MISTAKEN FOR COVERAGE:
//
//	`<role> <verb>`   the invocation form, with or without the binary in front of it
//	`motion <subject> <verb>`   the motion paths, which carry their subject
//	`` `<verb>` ``    the backticked bare name — code voice, which is how a constitution teaches
//	                  a name outside an invocation ("A grade moves through `regrade`")
//
// WHAT IT DELIBERATELY DOES NOT REACH: the bare word in prose. Most verb names are ordinary
// English — position, close, verify, observe — and redacting every occurrence would rewrite the
// document's argument rather than withhold its vocabulary, which is a different treatment wearing
// this one's name. The residue is reported by NamesSurviving instead of being assumed to be zero.
func Redact(text string, sf Surface) string {
	// Longest first, so `motion grade file` is redacted whole rather than leaving `grade file`
	// behind as debris that still names the act.
	var verbs []string
	seen := map[string]bool{}
	for _, role := range Roles {
		for _, v := range sf.Verbs(role) {
			if !seen[v] {
				seen[v] = true
				verbs = append(verbs, v)
			}
		}
	}
	sort.Slice(verbs, func(i, j int) bool { return len(verbs[i]) > len(verbs[j]) })

	const withheld = "⟨verb withheld⟩"
	for _, v := range verbs {
		q := regexp.QuoteMeta(v)
		// The invocation form: `blue prove`, `"…/feov-record" merge regrade`, `} bench opinion`.
		text = regexp.MustCompile(`\b(lens|merge|blue|bench)\s+`+q+`\b`).ReplaceAllString(text, "$1 "+withheld)
		// The motion paths carry their subject rather than a role.
		if strings.HasPrefix(v, "motion ") {
			text = regexp.MustCompile(`\b`+q+`\b`).ReplaceAllString(text, "motion "+withheld)
			continue
		}
		// The backticked bare name — code voice teaching a name outside an invocation.
		text = regexp.MustCompile("`"+q+"`").ReplaceAllString(text, "`"+withheld+"`")
	}
	return text
}

// NamesSurviving counts live verb names still present, so an arm's treatment is a measurement
// rather than a claim. See the type doc: a redactor that matched nothing would manufacture a null
// result and report it as a finding.
func NamesSurviving(text string, sf Surface) map[string]int {
	out := map[string]int{}
	seen := map[string]bool{}
	for _, role := range Roles {
		for _, v := range sf.Verbs(role) {
			if seen[v] {
				continue
			}
			seen[v] = true
			q := regexp.QuoteMeta(v)
			n := len(regexp.MustCompile(`\b(lens|merge|blue|bench)\s+`+q+`\b`).FindAllString(text, -1))
			if strings.HasPrefix(v, "motion ") {
				n = len(regexp.MustCompile(`\b`+q+`\b`).FindAllString(text, -1))
			} else {
				n += len(regexp.MustCompile("`"+q+"`").FindAllString(text, -1))
			}
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
