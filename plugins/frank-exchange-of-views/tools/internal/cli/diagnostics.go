package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/diagnostics"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// `show diagnostics` — DID THE SEATS FIND THEIR SURFACE, answered as fields.
//
// THIS EXISTED, AND IT EXISTED AS THE DEFECT IT MEASURES. The exposure figure every naming
// experiment turns on was computed by a scratchpad Python script that ran a regex over the
// probe's RENDERED MARKDOWN — a fact composed into prose at one end and recovered by matcher at
// the other, whose miss returns a plausible zero indistinguishable from a seat that never opened
// help. It bit that analysis twice: a `\s{2,}` column separator silently dropped exactly one verb
// per role, and line-level counting credited every seat with `friction` from the ROOT listing the
// board stages as setup — flattering precisely the arms that open help least.
//
// So it is a projection now: computed in Go, tested against both of those, and emitted as JSON a
// reader can join on rather than parse.
//
// IT IS THE OPERATOR'S, NOT A SEAT'S. A seat reading its own exposure score is Goodhart bait of
// exactly the kind the scorecards are already kept away from seat prompts for; the question
// "did this seat find its verbs" is one for the human who can retool it.
//
// AND IT IS NOT ONLY FOR PROBE RUNS. Every run has a trajectory, so the same question is askable
// of production — which is the argument for it living in the tool rather than in the harness.
func newShowDiagnostics() *cobra.Command {
	show := &cobra.Command{
		Use:          "show",
		Short:        "read an operator projection over a run (read-only)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	var trajectory, format, sitting string
	d := &cobra.Command{
		Use:   "diagnostics",
		Short: "DID THE SEATS FIND THEIR SURFACE: per seat, the verbs its role offers, the ones that reached it through help output (SEEN), the ones it reached for, and how it used --help. STRUCTURED JSON",
		Long: "diagnostics answers, for each seat that registered in a run, whether the surface reached it at all.\n\n" +
			"SEEN is EXPOSURE — the role's own verbs that appeared in help output the seat received — and it is a\n" +
			"different quantity from USE. A seat can reach for a verb it was never shown (it guessed, or a refusal\n" +
			"taught it) and can be shown verbs it never reaches for. Reporting one as the other is how a naming\n" +
			"experiment measures its own instrument.\n\n" +
			"SEEN is counted per BLOCK, never per line: the root listing carries names that are also role verbs, so\n" +
			"a line-level match credits every seat before it acts.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sc := seat.Of(cmd)
			// BEFORE the inference fallback, never after: falling back on a REFUSED resolution
			// resolves quietly to the real run and hides the contradiction the seat needs to see.
			if sc.RunErr != nil {
				return sc.RunErr
			}
			runDir := sc.RunDir
			if runDir == "" {
				runDir = seat.InferRunDir("")
			}
			if runDir == "" {
				return feov.Errorf(feov.MissingField, "show diagnostics: --run <runDir> is required")
			}
			traj := trajectory
			if traj == "" {
				traj = inferTrajectory(runDir)
			}
			if traj == "" {
				return feov.Errorf(feov.NotFound,
					"show diagnostics: no trajectory found for %s, and it is what carries the help a seat RECEIVED — "+
						"the record says what a seat wrote, never what it was shown. Pass --trajectory <path>", runDir)
			}
			report, err := diagnose(runDir, traj, sitting)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch format {
			case "markdown":
				fmt.Fprint(out, renderDiagnostics(report))
				return nil
			case "json":
				b, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(out, string(b))
				return nil
			default:
				// REFUSED, not rendered as the default. An unknown --format that quietly emits
				// JSON is the plausible-zero shape wearing a rendering hat: the caller asked for
				// something the tool does not have and got something that parses.
				return feov.Errorf(feov.Validation, "show diagnostics: unknown --format %q (json | markdown)", format)
			}
		},
	}
	// --sitting, NOT --seat-id. The persistent --seat-id says WHO IS ASKING and selects the tree
	// this command is even on; this says WHOSE SITTING to report. Spending one flag on both would
	// be the two-contracts-one-verb shape, and it would mean the operator could only ever
	// diagnose itself.
	d.Flags().StringVar(&sitting, flags.Sitting, "", "the seat whose sitting this trajectory is (REQUIRED: one trajectory is one sitting, and the record cannot say which of the registered seats it belongs to)")
	d.Flags().StringVar(&trajectory, flags.Trajectory, "", "path to the captured trajectory JSONL; inferred from the run when omitted")
	d.Flags().StringVar(&format, flags.Format, "json", "json (the form a reader joins on) | markdown")
	show.AddCommand(d)
	return show
}

// SeatDiagnostic is one seat's answer, as fields.
type SeatDiagnostic struct {
	SeatID string `json:"seatId"`
	Role   string `json:"role"`
	// Offers is every act the seat's role affords — the denominator.
	Offers []string `json:"offers"`
	// SeenTop saturates at one root help; SeenLeaf counts acts and does not.
	SeenTop  []string `json:"seenTop"`
	SeenLeaf []string `json:"seenLeaf"`
	// Blocks counted, and listings read and DISCARDED as not this role's. Both are carried
	// because "no listing reached this seat" and "every listing was somebody else's" are
	// different facts that a single zero would merge.
	HelpBlocks   int `json:"helpBlocks"`
	HelpRejected int `json:"helpRejected"`
	// Survey answers the question the three fields above cannot: when the seat reached for a
	// command, had it read THAT COMMAND's help first — and did it open any page for a verb it did
	// not go on to use? SEEN saturates on one root --help, so it restates "did this seat open
	// help at all"; a seat can score full exposure and still guess every flag it typed.
	Survey diagnostics.Survey `json:"survey"`
	// Groups are the pages a full traversal opens; GroupsUnopened the ones the seat never did.
	//
	// THE COMPLIANCE MEASURE FOR A TRAVERSAL RULE, and it did not exist while the rule was a
	// lookup rule: steps 2 and 3 both fired on "you are about to run this", so a seat could obey
	// perfectly and open nothing it did not already intend to use. Under that shape there was
	// nothing for this to find.
	Groups         []string `json:"groups"`
	GroupsUnopened []string `json:"groupsUnopened"`
	// The headline ratios, computed here so a reader does not have to and so two readers cannot
	// disagree about how they were derived.
	BlindFirst       int `json:"blindFirst"`
	RefusedBlind     int `json:"refusedBlind"`
	RefusedAfterRead int `json:"refusedAfterRead"`
}

// RunDiagnostic is the whole run's answer.
type RunDiagnostic struct {
	RunDir     string           `json:"runDir"`
	Trajectory string           `json:"trajectory"`
	Seats      []SeatDiagnostic `json:"seats"`
}

// toolName is how the seat's shell commands name this binary. Taken from the running executable
// rather than written down: `show diagnostics` IS the tool, so it already knows.
func toolName() string {
	exe, err := os.Executable()
	if err != nil {
		return "feov-record"
	}
	return filepath.Base(exe)
}

func diagnose(runDir, traj, seatID string) (RunDiagnostic, error) {
	out := RunDiagnostic{RunDir: runDir, Trajectory: traj}
	m, err := record.MergedEvents(runDir)
	if err != nil {
		return out, err
	}
	// ONE TRAJECTORY IS ONE SITTING, SO IT IS ONE SEAT'S — and the caller says which.
	//
	// THIS REPORTED FOUR SEATS ON ITS FIRST RUN AND THREE OF THEM NEVER SAT. A board's BUILD
	// registers every seat so the fixture has parties to hang events on; only one is then
	// dispatched. Reporting "the seats that registered" therefore credited three seats with
	// SEEN 0 — which reads as three seats that were shown nothing, when they were never asked
	// anything. That is the plausible zero, produced by the projection built to remove it, and it
	// is why this refuses rather than guesses: nothing in a trajectory attributes an invocation
	// to one of several registered seats, so there is no honest way to split it.
	ids := []string{seatID}
	if seatID == "" {
		var registered []string
		for _, e := range m.Events {
			if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
				registered = append(registered, e.GetSeatId())
			}
		}
		sort.Strings(registered)
		return out, feov.Errorf(feov.MissingField,
			"show diagnostics: --sitting names whose sitting this trajectory is. A run's record carries every seat "+
				"the BUILD registered (%s), and only one of them sat; reporting them all would credit the rest with "+
				"an exposure of zero they were never given the chance to earn",
			strings.Join(dedupeSeats(registered), ", "))
	}
	if !registeredIn(m.Events, seatID) {
		return out, feov.Errorf(feov.NotFound,
			"show diagnostics: %q never registered in this run, so there is no sitting to report on", seatID)
	}

	for _, id := range ids {
		role := RoleOfSeat(id)
		acts := actsOf(role)
		s, err := diagnostics.ReadSeen(traj, acts)
		if err != nil {
			return out, err
		}
		// The traversal is checked against the seat's OWN verb set, so a prose word that leaked
		// out of a --reason body cannot be counted as a group the seat visited.
		tops := map[string]bool{}
		for _, a := range acts {
			tops[strings.Fields(a)[0]] = true
		}
		sv, err := diagnostics.ReadSurvey(traj, toolName(), tops)
		if err != nil {
			return out, err
		}
		if sv.ToolUnrecognised() {
			return out, feov.Errorf(feov.Validation,
				"show diagnostics: the trajectory has %d shell calls and none of them invoke %q, so every survey figure "+
					"would read as a seat that ran nothing. Either this trajectory is not a sitting with this tool, or the "+
					"tool was built under another name", sv.BashCalls, toolName())
		}
		groups := groupsOf(role)
		opened := map[string]bool{}
		for _, p := range sv.HelpPages {
			opened[p] = true
		}
		// EMPTY, NOT NIL. A nil slice marshals to `null`, and a consumer then cannot tell "this
		// seat opened every group" from "this field was never computed" — the same bytes for the
		// best outcome and a broken one. It cost the first run-6 comparison, which crashed on the
		// good news.
		unopened := []string{}
		for _, g := range groups {
			if !opened[g] {
				unopened = append(unopened, g)
			}
		}
		rb, rr := sv.RefusedBlind()
		out.Seats = append(out.Seats, SeatDiagnostic{
			SeatID: id, Role: role, Offers: acts,
			SeenTop: keys(s.Top), SeenLeaf: keys(s.Leaf),
			HelpBlocks: s.Blocks, HelpRejected: s.Rejected,
			Survey: sv, Groups: groups, GroupsUnopened: unopened, BlindFirst: sv.BlindFirst(),
			RefusedBlind: rb, RefusedAfterRead: rr,
		})
	}
	return out, nil
}

// actsOf is what THIS SEAT'S TREE offers, walked from the real root.
//
// IT WAS THE role-PREFIXED SUBSET OF CommandPaths(), AND THAT IS A DIFFERENT SET. `motion`,
// `fetch` and `count-claims` are mounted on every seat's root and carry no role in their join
// key, so they were missing from the denominator — and because SEEN rejects a listing containing
// a name the role does not own, their presence in the seat's OWN help discarded the block whole.
// The measure reported a lens that had just been handed its entire surface as having seen
// nothing.
//
// The tree is the authority for what a seat is offered, exactly as it is for everything else
// here. A join key composed for a table is not the same fact as a command a seat can reach.
// groupsOf is the acts that HAVE CHILDREN — the pages a survey has to open to reach leaf depth.
//
// THE DENOMINATOR OF THE TRAVERSAL, and it is small: two or three per seat, against roots listing
// 12 to 19 commands. That size is the argument for asking a seat to open all of them rather than
// the one holding the verb it already picked — the whole tree to leaf depth costs 3-4 calls in a
// sitting of 16 to 32.
//
// `completion` and `help` are cobra's, generated onto every parent, and belong to no role.
func groupsOf(role string) []string {
	root := NewRootFor(record.SampleSeatOf(role))
	if root == nil {
		return nil
	}
	var out []string
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		for _, sub := range c.Commands() {
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			path := strings.TrimSpace(prefix + " " + sub.Name())
			if sub.HasSubCommands() {
				out = append(out, path)
			}
			walk(sub, path)
		}
	}
	walk(root, "")
	sort.Strings(out)
	return out
}

func actsOf(role string) []string {
	root := NewRootFor(record.SampleSeatOf(role))
	if root == nil {
		return nil
	}
	var out []string
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		for _, sub := range c.Commands() {
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			path := strings.TrimSpace(prefix + " " + sub.Name())
			out = append(out, path)
			walk(sub, path)
		}
	}
	walk(root, "")
	sort.Strings(out)
	return out
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// inferTrajectory looks where a probe puts one, then where a run does. Empty when neither is
// there — a MISSING trajectory must refuse rather than report a run in which no seat saw anything.
func inferTrajectory(runDir string) string {
	probe := filepath.Join(filepath.Dir(runDir), ".probe", filepath.Base(runDir)+".jsonl")
	if st, err := os.Stat(probe); err == nil && !st.IsDir() {
		return probe
	}
	inRun := filepath.Join(runDir, "trajectories", "journal.jsonl")
	if st, err := os.Stat(inRun); err == nil && !st.IsDir() {
		return inRun
	}
	return ""
}

func renderDiagnostics(r RunDiagnostic) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Diagnostics — %s\n\n", r.RunDir)
	for _, s := range r.Seats {
		fmt.Fprintf(&b, "## %s (%s)\n\n", s.SeatID, s.Role)
		fmt.Fprintf(&b, "- offers %d act(s)\n", len(s.Offers))
		fmt.Fprintf(&b, "- SEEN top %d, leaf %d\n", len(s.SeenTop), len(s.SeenLeaf))
		fmt.Fprintf(&b, "- help listings counted %d, rejected %d\n\n", s.HelpBlocks, s.HelpRejected)
	}
	return b.String()
}

func dedupeSeats(in []string) []string {
	seen, out := map[string]bool{}, []string{}
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func registeredIn(evs []*record.Event, seatID string) bool {
	for _, e := range evs {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER && e.GetSeatId() == seatID {
			return true
		}
	}
	return false
}
