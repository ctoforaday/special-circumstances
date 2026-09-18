package lens

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// class: the gap-class registry, and coining an entry in it.
//
// IT WAS FOUR FLAGS ON `mint`, and that is what a verb looks like when it is wearing another
// verb's clothes. `--class-new` was a boolean whose whole meaning was "I also passed
// --definition, --neighbor and --distinguisher"; the three of them were optional to cobra and
// policed in the mint handler; and a mint that coined a class wrote TWO events, one of which had
// nothing to do with putting a gap on the board.
//
// A cluster of flags that must travel together, changes what else is required, and writes its
// own event, is a verb. As one, cobra refuses an incomplete coining at parse and `--help` can
// say what a class IS without that competing for room with what a gap is.
func newClass() *cobra.Command {
	c := &cobra.Command{
		Use:          "class",
		Short:        "the gap-class registry: what KINDS of defect this run recognises",
		SilenceUsage: true,
	}
	c.AddCommand(newClassNew())
	c.AddCommand(newClassList())
	return c
}

// class list: the vocabulary `--class` and `--neighbor` are checked against.
//
// `class` had one subcommand, and coining was the only thing a seat could do with the registry —
// so the check refused a `--neighbor` the seat had no way to look up. The remedy a lens actually
// used was to leave the record tool, go outside the run directory, and grep the checked-in Go
// source and feov-memory/class-registry.json for the vocabulary. Every other read on this surface
// is a projection of the record; this was the one that was not, because it did not exist.
//
// READ-ONLY, and it records nothing: a seat asking what the words are has not yet done anything.
func newClassList() *cobra.Command {
	return seat.New("list", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		rows, err := record.ClassRoster(run)
		if err != nil {
			return nil, err
		}
		return classListResult{Classes: rows}, nil
	})
}

type classListResult struct {
	Classes []record.ClassRow `json:"classes"`
}

func (r classListResult) Human() string {
	// AN EMPTY ROSTER IS NOT A CLEAN BOARD. ClassRoster refuses a run with no registry staged
	// rather than returning nothing, so reaching here empty would mean a registry that parsed and
	// held no rows — which is a broken vocabulary, not an absent one, and is said as such.
	if len(r.Classes) == 0 {
		return "class list: the staged registry holds NO classes — every `--class` would have nothing to match, so this is a broken registry rather than an empty one"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "class list: %d class(es) this run accepts for --class and --neighbor.", len(r.Classes))
	sb.WriteString("\n  The material default decides whether a gap of that class holds the PASS gate.")
	for _, c := range r.Classes {
		md := c.MaterialDefault
		if md == "" {
			md = "—"
		}
		origin := ""
		if c.Coined {
			origin = "  (coined on this run)"
		}
		fmt.Fprintf(&sb, "\n  %-42s %-9s%s", c.Slug, md, origin)
		// THE COINED ROWS CARRY WHAT A NEIGHBOUR CHOICE NEEDS, and the shipped rows do not — setup
		// stages only the slug and the default. Printing the definition where it exists, and
		// nothing where it does not, is the honest shape; a blank line under every shipped class
		// would read as missing data rather than as data the registry never staged.
		if c.Definition != "" {
			fmt.Fprintf(&sb, "\n      %s", c.Definition)
		}
		if c.Neighbor != "" {
			fmt.Fprintf(&sb, "\n      nearest: %s", c.Neighbor)
		}
	}
	return sb.String()
}

func newClassNew() *cobra.Command {
	c := seat.Records(seat.New("new", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		// THE DEFAULT IS ALWAYS WRITTEN. A coined class with no default on the record would leave
		// every mint of it without the value its materiality starts from, so the verb supplies
		// `by_grade` when the coiner says nothing.
		word := seat.Str(cmd, flags.MaterialDefault)
		if word == "" {
			word = recordpb.Word(recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE)
		}
		md, ok := record.ClassMaterialOf(word)
		if !ok {
			return nil, fmt.Errorf("class new: %q is not a material default — always | never | by_grade", word)
		}
		body := &recordpb.ClassNew{
			Slug:            proto.String(seat.Str(cmd, flags.Class)),
			Definition:      proto.String(seat.Str(cmd, flags.Definition)),
			Neighbor:        proto.String(seat.Str(cmd, flags.Neighbor)),
			Distinguisher:   proto.String(seat.Str(cmd, flags.Distinguisher)),
			MaterialDefault: md.Enum(),
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return classResult{Slug: seat.Str(cmd, flags.Class)}, nil
	}), "class_new")

	c.Flags().String(flags.Class, "", "the slug you are coining — lowercase, hyphenated, and NOT one the registry already has")
	flags.Text(c, flags.Definition, "what this class is, in one line")
	c.Flags().String(flags.Neighbor, "", "the existing class it sits closest to")
	flags.Text(c, flags.Distinguisher, "the tie-break question that tells the two apart — without it a new class is a synonym, and the registry stops discriminating")
	enumhelp.Flag(c, flags.MaterialDefault, record.MustEnum("class_new", "material_default"),
		"where materiality starts for every gap of this class (default by_grade) — always for a class whose defects change a conclusion or a figure, never for one whose defects change none")
	_ = c.MarkFlagRequired(flags.Class)
	// COBRA SAYS THE TRIO TRAVELS TOGETHER. It was a boolean plus three optional flags checked
	// in a handler, so a coining missing its distinguisher was composed, sent, and refused after
	// the fact — if it was refused at all.
	c.MarkFlagsRequiredTogether(flags.Definition, flags.Neighbor, flags.Distinguisher)
	_ = c.MarkFlagRequired(flags.Definition)
	return c
}

type classResult struct {
	Slug string `json:"slug"`
}

func (r classResult) Human() string {
	return "class " + r.Slug + " coined — mint against it with `mint --class " + r.Slug + "`"
}
