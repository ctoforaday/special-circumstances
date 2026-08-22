// Package motion is the ONE adjudication mechanism (#344).
//
// Three exchanges — a grade dispute, a petition, a ruling on a proposed direction — were three
// verb pairs with three vocabularies and no shared identity. `ruling`/`ruling`/`response` for one
// concept, three renderers, and nothing tying an ask to its answer. #315 found the petition
// FILING unrendered while the line of inquiry RULING was found unrendered SEPARATELY in the same sweep,
// because nothing said they were the same mechanism; #312 is the same root.
//
// A motion has an ID, and the ask and its answer join on it.
//
// # Why this is a top-level group and not a role's verb set
//
// The other groups hang off a role node, and `seat.Of` reads the role from the command's
// POSITION — a verb's parent IS its role. That cannot work here: under `motion grade file` the
// parent is `grade`. And it should not work here, because a motion is not one seat's act. All
// four seats file; merge and bench rule.
//
// So these verbs take the acting role from the SEAT'S IDENTITY rather than from tree position,
// which is the model plans/command-groups.md stage 6 moves the whole tree to. It arrives early
// for the one group that cannot be expressed the old way.
//
// # Why the subjects are SUBGROUPS and not a --on flag
//
// Their contracts diverge: a grade motion needs --dimension and --proposed and is ruled by the
// merge; a petition needs --class and --relief and is ruled by the BENCH; a direction needs
// neither and has no filing verb at all (the proposal IS the filing). Cobra cannot express
// "required only when --on=grade", so a flag-discerned subject would put three contracts into
// hand-written RunE validation — a flag combination policed by prose, which is the shape this
// suite exists to remove. As subgroups, cobra refuses the nonsense at parse.
package motion

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "motion",
		Short:        "file and rule on a motion — a grade dispute, a petition, or a ruling on a direction. One mechanism, one id.",
		SilenceUsage: true,
	}
	// THE GAVEL IS NOT TYPED HERE. It was — `subject("petition", …, "bench")` — and the PASS gate
	// in internal/record, which cannot import this package, told blocked seats to rule motions
	// without knowing whose ruling it would be. Both readers take it off the MotionSubject enum
	// now, so a subject cannot be added with a gavel in one place and not the other.
	c.AddCommand(subject("grade",
		"contest a gap's grade: the merge rules, and a rejected dispute may be appealed to the bench",
		[]string{flags.ID, flags.Dimension, flags.Proposed}))
	c.AddCommand(subject("petition",
		"an ethical | safety | integrity | constitutional objection: the BENCH rules, before the debate continues",
		[]string{flags.Class, flags.Relief}))
	c.AddCommand(subject("inquiry",
		"rule on a line blue proposed: the merge rules. There is NO file verb — the proposal is the filing (`blue line of inquiry`)",
		nil))
	return c
}

// rulerFor resolves the seat holding a subgroup's gavel from the schema.
//
// It PANICS on an unknown subject or an unannotated one, and that is deliberate: this runs at
// command construction, so the failure is at startup for every seat rather than at the moment one
// tries to rule. A subgroup whose gavel nobody declared would otherwise accept rulings from
// everybody, which is the arm requireRuler exists to close.
func rulerFor(name string) string {
	subj, ok := record.MotionSubjectEnum(name)
	if !ok {
		panic("motion: subgroup " + name + " names no MotionSubject — the subgroup and the schema have diverged")
	}
	r, err := recordpb.SubjectRuler(subj)
	if err != nil {
		panic("motion: " + err.Error())
	}
	return r
}

// subject builds one subgroup. `direction` gets no `file`: red rules on a line blue already
// proposed, so a filing verb here would be a second way to say what `blue line of inquiry` already says.
func subject(name, short string, fileFlags []string) *cobra.Command {
	ruler := rulerFor(name)
	c := &cobra.Command{
		Use: name, Short: short, SilenceUsage: true,
		// Tolerate unknown flags AT THE GROUP LEVEL so the Args check below is what answers.
		// Cobra parses flags before Args, so `motion petition appeal --id M1` died on `--id`
		// (unknown to the group) and never reached the message that explains the real problem.
		// The group has no RunE, so anything landing here is already an error path — the only
		// question is which error the seat is told about.
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	// A VERB THIS SUBJECT DOES NOT HAVE MUST SAY SO. Without this, cobra parsed the verb word as
	// a positional ARGUMENT to the subgroup and then failed on the first flag: `motion petition
	// appeal --id M1` answered "unknown flag: --id", which sends a seat looking at its flags for
	// a problem that is not there. Measured by probing.
	//
	// The absent verb is a real design statement — a petition is heard BEFORE the debate
	// continues, so there is nothing to escalate to — and the refusal is where a seat actually
	// meets it, so it is where the reason belongs.
	c.RunE = func(_ *cobra.Command, args []string) error {
		var have []string
		for _, sub := range c.Commands() {
			have = append(have, sub.Name())
		}
		named := "names no verb"
		if len(args) > 0 {
			named = "has no `" + args[0] + "` verb"
		}
		return feov.Errorf(feov.Validation,
			"a %s motion %s — it has %s. If you expected one, its absence is the design, not an omission: see `motion %s --help`",
			name, named, strings.Join(have, ", "), name)
	}
	if fileFlags != nil {
		c.AddCommand(newFile(name, fileFlags))
	}
	c.AddCommand(newRule(name, ruler))
	if name != "petition" {
		// A petition has no appeal: it is heard BEFORE the debate continues, so there is
		// nothing to escalate to. Expressed by absence rather than by a runtime refusal.
		c.AddCommand(newAppeal(name))
	}
	return c
}

// actingRole reports the role of the seat running the command, from its identity.
//
// Not from tree position (see the package comment) and not from a flag — the seat id is injected
// and cross-checked (#348), so it is the one identity fact a seat cannot quietly get wrong.
func actingRole(seatID string) string {
	for _, r := range []string{"lens", "merge", "blue", "bench"} {
		if record.CheckSeatRole(r, seatID) == nil {
			return r
		}
	}
	return ""
}

// requireRuler refuses a ruling from a seat that does not hold the gavel for this subject.
//
// The message NAMES the seat that does. Under a scoped surface "not yours" would otherwise read
// as "does not exist" — the measured failure where a seat handed an unavailable verb logs
// friction and works around it, losing the capability for the run.
func requireRuler(subject, ruler, seatID string) error {
	if got := actingRole(seatID); got != ruler {
		return feov.Errorf(feov.Validation,
			"a %s motion is ruled by the %s seat; you are %s (%s). A motion is filed by any seat and ruled by one — that asymmetry is the mechanism, not an obstacle",
			subject, ruler, seatID, got)
	}
	return nil
}

// prose pulls the one prose field, refusing an empty one by name.
func prose(cmd *cobra.Command, verb, why string) (string, error) {
	text, err := seat.Reason(cmd)
	if err != nil {
		return "", err
	}
	if text == "" {
		return "", feov.Errorf(feov.Validation, "motion %s requires --reason: %s", verb, why)
	}
	return text, nil
}
