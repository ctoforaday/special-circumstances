package seat

import (
	"fmt"

	"github.com/spf13/cobra"

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

func Register(role, help string) *cobra.Command {
	return New(role, "register", help, func(s Context, _ *cobra.Command) (string, error) {
		nonce, _, err := record.RegisterSeat(s.RunDir, s.SeatID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("registered %s (shard nonce %s)", s.SeatID, nonce), nil
	})
}

func Friction(role, help string) *cobra.Command {
	return Prose(New(role, "friction", help, func(s Context, cmd *cobra.Command) (string, error) {
		text, err := Text(cmd)
		if err != nil {
			return "", err
		}
		if _, err := record.Append(s.RunDir, s.SeatID, "friction", record.NewPayload().Set("text", text)); err != nil {
			return "", err
		}
		return "friction recorded", nil
	}))
}

// Petition takes a suffix because the lens is told one more thing than the
// others: that the bench hears it before the debate continues.
func Petition(role, help, suffix string) *cobra.Command {
	c := New(role, "petition", help, func(s Context, cmd *cobra.Command) (string, error) {
		p := SetSame(cmd, record.NewPayload(), "class", "basis", "relief")
		if _, err := record.Append(s.RunDir, s.SeatID, "petition", p); err != nil {
			return "", err
		}
		return fmt.Sprintf("petition filed (%s)%s", Str(cmd, "class"), suffix), nil
	})
	c.Flags().String("class", "", "ethical | safety | integrity | constitutional")
	c.Flags().String("basis", "", "what happened, and why it reaches the bench")
	c.Flags().String("relief", "", "the relief sought, stated as it would bind the coming seats")
	return c
}

func Position(role, help string) *cobra.Command {
	return Prose(New(role, "position", help, func(s Context, cmd *cobra.Command) (string, error) {
		text, err := Text(cmd)
		if err != nil {
			return "", err
		}
		if _, err := record.Append(s.RunDir, s.SeatID, "position", record.NewPayload().Set("text", text)); err != nil {
			return "", err
		}
		return "position recorded", nil
	}))
}

func Closing(role, help string) *cobra.Command {
	c := Prose(New(role, "closing", help, func(s Context, cmd *cobra.Command) (string, error) {
		text, err := Text(cmd)
		if err != nil {
			return "", err
		}
		p := Set(cmd, record.NewPayload(), "gap_id", "id")
		p.Set("text", text)
		if _, err := record.Append(s.RunDir, s.SeatID, "closing", p); err != nil {
			return "", err
		}
		return fmt.Sprintf("closing filed for %s", Str(cmd, "id")), nil
	}))
	c.Flags().String("id", "", "the gap id this closing argues")
	return c
}

// Render is available to every seat and mutates nothing.
func Render(role string) *cobra.Command {
	c := &cobra.Command{
		Use:          "render",
		Short:        "read-only projection refresh (any seat may invoke)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		runDir := Str(cmd, "run")
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
		fmt.Printf("feov-record %s: rendered to %s (%d open, %d closed%s)\n", role, r.Out, r.Open, r.Closed, extra)
		return nil
	}
	return c
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
		names = append(names, v.Name())
		c.AddCommand(v)
	}
	c.AddCommand(Render(role))
	names = append(names, "render")
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
