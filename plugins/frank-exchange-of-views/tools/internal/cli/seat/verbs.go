package seat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// The verbs every role shares, defined once.
//
// register, friction, petition, position and closing are the SAME contract
// wherever they appear — a friction entry from a lens and one from the bench are
// the same event with the same payload. Restating them per role would be four
// copies to drift apart, and the drift would be silent because each copy would
// still pass its own tests.
//
// Each role still supplies its OWN help text, because the contract is shared
// while the guidance is not: a lens is told petitions do not pause its duties,
// the bench is told the same verb from the other side.

func Register(help string) *cobra.Command {
	return New("register", help, func(s Context, _ *cobra.Command) (Result, error) {
		nonce, _, err := record.RegisterSeat(s.RunDir, s.SeatID)
		if err != nil {
			return nil, err
		}
		return registerResult{SeatID: s.SeatID, Nonce: nonce}, nil
	})
}

func Friction(help string) *cobra.Command {
	return Prose(New("friction", help, func(s Context, cmd *cobra.Command) (Result, error) {
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		if _, err := record.Append(s.RunDir, s.SeatID, "friction", record.NewPayload().Set("text", text)); err != nil {
			return nil, err
		}
		return Msg{Message: "friction recorded"}, nil
	}))
}

// Petition takes a suffix because the lens is told one more thing than the
// others: that the bench hears it before the debate continues.
func Petition(help, suffix string) *cobra.Command {
	c := New("petition", help, func(s Context, cmd *cobra.Command) (Result, error) {
		p := SetSame(cmd, record.NewPayload(), flags.Relief)
		// Flag word --reason (the one prose word), payload key stays basis.
		reason, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		if reason != "" {
			p.Set("basis", reason)
		}
		Set(cmd, p, "class", flags.PetitionClass)
		if _, err := record.Append(s.RunDir, s.SeatID, "petition", p); err != nil {
			return nil, err
		}
		return petitionResult{Class: Str(cmd, flags.PetitionClass), Suffix: suffix}, nil
	})
	c.Flags().String(flags.PetitionClass, "", record.MustEnum("petition", "class").Usage("what you are asking the bench to sit on"))
	c.Flags().String(flags.Relief, "", "the relief sought, stated as it would bind the coming seats")
	return Prose(c)
}

func Position(help string) *cobra.Command {
	return Prose(New("position", help, func(s Context, cmd *cobra.Command) (Result, error) {
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		if _, err := record.Append(s.RunDir, s.SeatID, "position", record.NewPayload().Set("text", text)); err != nil {
			return nil, err
		}
		return Msg{Message: "position recorded"}, nil
	}))
}

func Closing(help string) *cobra.Command {
	c := Prose(New("closing", help, func(s Context, cmd *cobra.Command) (Result, error) {
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		p := Set(cmd, record.NewPayload(), "gap_id", flags.ID)
		p.Set("text", text)
		if _, err := record.Append(s.RunDir, s.SeatID, "closing", p); err != nil {
			return nil, err
		}
		return closingResult{ID: Str(cmd, flags.ID)}, nil
	}))
	c.Flags().String(flags.ID, "", "the gap id this closing argues")
	return c
}

// Render is available to every seat and mutates nothing.
func Render() *cobra.Command {
	c := &cobra.Command{
		Use:          "render",
		Short:        "read-only projection refresh (any seat may invoke)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		role := roleOf(cmd)
		runDir := Str(cmd, flags.Run)
		if runDir == "" {
			return fmt.Errorf("%s: --run <runDir> is required", role)
		}
		r, err := record.Render(runDir, "")
		if err != nil {
			return err
		}
		extra := ""
		if r.Anomalies > 0 {
			extra = fmt.Sprintf(", %d anomalies", r.Anomalies)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "feov-record %s: rendered to %s (%d open, %d closed%s)\n", role, r.Out, r.Open, r.Closed, extra)
		return nil
	}
	return c
}

// views are the projections a seat may read, and the role whose view each one is by
// default. The default exists so a seat can type `show` and get the artifact it works
// against, without having to learn the file layout of a directory it should not be
// reading directly in the first place.
var views = []struct {
	name, desc, defaultFor string
}{
	{"board", "STRUCTURED JSON: open and closed gaps with grades, closures, anchors, observations and their fates, counts, and any replay anomalies — the form a seat acts on", "merge"},
	{"findings", "STRUCTURED JSON: every lens finding on the record (label, seat, round, role, grades, location, text) — the merge coalesces these into gaps; replaces the red/candidates/*.md files", "merge"},
	{"friction", "STRUCTURED JSON: every friction event on the record (seat, round, text) — capability/protocol complaints as events; read by the dashboard instead of parsing a markdown file", ""},
	{"ledger", "the board as markdown, for a human verification pass", ""},
	{"archive", "closed gaps with their closure records and anchors", ""},
	{"debate", "the round-by-round transcript, every seat's sections in order (add --json for the STRUCTURED form: rounds with red/blue/lead sections as data, for the audits)", "bench"},
	{"changelog", "blue's revision record, per round", "blue"},
	{"citation-ledger", "verified claims with source, confidence and access date", "lens"},
	{"lines-of-inquiry", "the exploration space: avenues taken, declined and abandoned", ""},
}

func viewNames() []string {
	out := make([]string, 0, len(views))
	for _, v := range views {
		out = append(out, v.name)
	}
	return out
}

// Show prints a projection to STDOUT.
//
// READS BELONG IN THE TOOL. Every seat could already WRITE through the tool while still
// having to READ the run by opening markdown files at paths it learned from a prompt —
// which is how a seat comes to trust a hand-written artifact over the event log, and how
// the two came to disagree. A seat that asks the tool gets the answer the events support.
//
// It does NOT re-derive the markdown. It renders through the exact path `render` uses and
// then prints the resulting file, so the bytes on stdout and the bytes on disk cannot
// differ. A second renderer would be a second reader of one artifact, which is the defect
// class this whole tool exists to remove — writing one to serve reads would reintroduce it
// at the read surface.
func Show() *cobra.Command {
	c := &cobra.Command{
		Use:          "show",
		Short:        "read a projection on STDOUT (the tool is the read path; the .md files are for human verification): --view " + strings.Join(viewNames(), "|"),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		role := roleOf(cmd)
		runDir := Str(cmd, flags.Run)
		if runDir == "" {
			return fmt.Errorf("%s: --run <runDir> is required", role)
		}
		want := Str(cmd, flags.View)
		if want == "" {
			for _, v := range views {
				if v.defaultFor == role {
					want = v.name
				}
			}
		}

		// --json on a read opts into the STRUCTURED form of a view whose native form is
		// markdown. It is the same inherited flag the mutating verbs use for their JSON
		// envelopes; on reads it had been ignored (reads were view-selected only). It now
		// selects the structured debate — the sole view with both a markdown transcript (for
		// the human) and a JSON form (for the audits that used to regex the sections).
		//
		// One canonical way to each form ([[one-way-no-aliases]]): --json is an ERROR on the
		// views already JSON by name (board/findings/friction — `--view board` is the single
		// way to board JSON, no alias) and on markdown views with no JSON form. It is checked
		// BEFORE the board/findings/friction branches so `--view board --json` refuses rather
		// than falling through to the flagless JSON. A wrong guess fails loudly.
		if asJSON, _ := cmd.Flags().GetBool(flags.JSON); asJSON {
			switch want {
			case "debate":
				b, err := record.DebateJSONBytes(runDir)
				if err != nil {
					return err
				}
				cmd.OutOrStdout().Write(b)
				return nil
			case "board", "findings", "friction":
				return fmt.Errorf("%s show: --view %s is already JSON by name — drop --json (it is the single way to that view's JSON)", role, want)
			case "":
				return fmt.Errorf("%s show: --view is required for this role (one of: %s)", role, strings.Join(viewNames(), ", "))
			default:
				return fmt.Errorf("%s show: --view %s has no --json form (only 'debate' does; board/findings/friction are JSON by name)", role, want)
			}
		}

		// The BOARD is served as structured JSON, because a seat ACTS on it.
		//
		// Every other view is prose a seat reads; the board is state a seat has to make
		// decisions against — which gap is open, what its grades are, whether a closure
		// carried its anchor. Parsing that back out of markdown is precisely where the
		// scorecard defects came from (anchored_closures_pct read 0 against an 89
		// baseline because it parsed sentences while the anchors sat in structured
		// fields). Markdown for the same state stays available behind --view ledger.
		//
		// Resolved AFTER the role default, not before: `merge show` with no flags must
		// reach this branch, and checking the raw flag first sent the merge seat's own
		// default view looking for a board.md that no renderer writes.
		if want == "board" {
			b, err := record.BoardJSONBytes(runDir)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(b)
			return nil
		}
		// findings is served as JSON too, and for the same reason: the merge ACTS on it
		// (coalesces findings into gaps), so it reads structured fields, not prose it must
		// parse. This is the channel that replaced red/candidates/*.md.
		if want == "findings" {
			b, err := record.FindingsJSONBytes(runDir)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(b)
			return nil
		}
		if want == "friction" {
			b, err := record.FrictionJSONBytes(runDir)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(b)
			return nil
		}
		if want == "" {
			return fmt.Errorf("%s show: --view is required for this role (one of: %s)", role, strings.Join(viewNames(), ", "))
		}
		var file string
		for _, v := range views {
			if v.name == want {
				file = viewFile(v.name)
			}
		}
		if file == "" {
			return fmt.Errorf("%s show: unknown view %q (one of: %s)", role, want, strings.Join(viewNames(), ", "))
		}

		r, err := record.Render(runDir, "")
		if err != nil {
			return err
		}
		b, err := os.ReadFile(filepath.Join(r.Out, file))
		if err != nil {
			return fmt.Errorf("%s show: the %s projection is not on disk after a render — this is a renderer defect, not a missing artifact: %w", role, want, err)
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	c.Flags().String(flags.View, "", "which projection to read: "+strings.Join(viewNames(), " | ")+" (defaults to this role's own)")
	return c
}

// viewFile maps a view name to the file the renderer writes. The renderer's filenames
// are its own business; a seat should never have to know them.
func viewFile(view string) string {
	switch view {
	case "changelog":
		return "CHANGELOG.md"
	default:
		return view + ".md"
	}
}

// Role assembles a role's command from the verbs it was given.
//
// The role knows only its own name, its own one-line contract, and which verbs
// it has. That list IS the role boundary — a lens has no mint verb to call —
// and a seat that reaches past it is answered with what it CAN do rather than
// cobra's generic "unknown command".
func Role(role, short string, verbs ...*cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   role,
		Short: short,
		Long:  short + "\n" + FrictionFooter,
		// ArbitraryArgs with flag parsing off at the ROLE level so an unrecognised
		// verb reaches RunE instead of failing on a flag the role does not own:
		// `lens mint --run x` must answer "the lens has no mint verb", not
		// "unknown flag: --run". Subcommand resolution happens before flag
		// parsing, so each verb still parses and validates its own flags strictly.
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		SilenceUsage:       true,
	}
	names := make([]string, 0, len(verbs)+1)
	for _, v := range verbs {
		// Applied HERE rather than in each verb: a verb that had to remember to mark its
		// own required flags is a verb that can forget, and the forgetting is silent —
		// the help simply looks like everything is optional. Every role passes through
		// this loop, so every verb is annotated by construction.
		markRequired(v, v.Name())
		names = append(names, v.Name())
		c.AddCommand(v)
	}
	c.AddCommand(Render())
	names = append(names, "render")
	c.AddCommand(Show())
	names = append(names, "show")
	available := join(names)

	c.RunE = func(cmd *cobra.Command, args []string) error {
		for _, a := range args {
			if a == "--help" || a == "-h" || a == "help" {
				return cmd.Help()
			}
		}
		if len(args) > 0 {
			return fmt.Errorf("verb %q is outside this seat's role (available: %s)", args[0], available)
		}
		// A role invoked with no verb is a usage error, not a no-op: silently
		// succeeding would let a mis-scripted seat believe it recorded something.
		return fmt.Errorf("%s: a verb is required (%s)", role, available)
	}
	return c
}

func join(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}

// registerResult, petitionResult and closingResult are the shared-verb results: these
// three verbs are built by seat for every role, so their result types live beside them.
type registerResult struct {
	SeatID string `json:"seat_id"`
	Nonce  string `json:"nonce"`
}

func (r registerResult) Human() string {
	return "registered " + r.SeatID + " (shard nonce " + r.Nonce + ")"
}

type petitionResult struct {
	Class  string `json:"class"`
	Suffix string `json:"suffix,omitempty"`
}

func (r petitionResult) Human() string { return "petition filed (" + r.Class + ")" + r.Suffix }

type closingResult struct {
	ID string `json:"id"`
}

func (r closingResult) Human() string { return "closing filed for " + r.ID }
