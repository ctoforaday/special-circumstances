package motion

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// file: a seat asks for something, and the tool assigns the id everything else joins on.
func newFile(subject string, required []string) *cobra.Command {
	c := &cobra.Command{
		Use:   "file",
		Short: "file a " + subject + " motion — the tool assigns its id",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := seat.Of(cmd)
			basis, err := prose(cmd, "file", "the ASK in your own words — a motion with no argument is a demand, and the ruler has nothing to rule on")
			if err != nil {
				return err
			}
			for _, f := range required {
				if strings.TrimSpace(seat.Str(cmd, f)) == "" {
					return feov.Errorf(feov.Validation,
						"a %s motion requires --%s. The subjects have different contracts, which is why they are subgroups: cobra cannot express \"required only when the subject is %s\", so the requirement lives here where it can refuse",
						subject, f, subject)
				}
			}
			id, err := record.MintMotionID(s.RunDir)
			if err != nil {
				return err
			}
			p := record.NewPayload().Set("motion_id", id).Set("subject", subject).Set("basis", basis)
			for _, f := range required {
				// The payload key is the flag word for these, except --id, whose key names WHAT
				// it identifies — a motion's own id is `motion_id`, so a gap reference must not
				// also be `id` or the two collide in one payload.
				key := f
				if f == flags.ID {
					key = "gap_id"
				}
				p.Set(key, seat.Str(cmd, f))
			}
			if r := seat.Str(cmd, flags.Relief); r != "" {
				p.Set("relief", r)
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "motion", p); err != nil {
				return err
			}
			return seat.Emit(cmd, filed{ID: id, Subject: subject}, nil)
		},
	}
	seat.Prose(c)
	for _, f := range required {
		c.Flags().String(f, "", "REQUIRED for a "+subject+" motion")
	}
	return c
}

// rule: the ruler answers, on the motion's id.
func newRule(subject, ruler string) *cobra.Command {
	e := record.MotionVerdictEnum(subject)
	c := &cobra.Command{
		Use:   "rule",
		Short: "rule on a " + subject + " motion (the " + ruler + " seat's)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := seat.Of(cmd)
			if err := requireRuler(subject, ruler, s.SeatID); err != nil {
				return err
			}
			id := seat.Str(cmd, flags.ID)
			if err := record.RequireMotionSubjectRef(s.RunDir, subject, id); err != nil {
				return err
			}
			opinion, err := prose(cmd, "rule", "an unreasoned ruling is the decoration the filer cannot contest, and contesting it is the whole reason a ruling is not a command")
			if err != nil {
				return err
			}
			p := record.NewPayload().Set("motion_id", id).Set("subject", subject).
				Set("ruling", seat.Str(cmd, flags.As)).Set("opinion", opinion)
			if _, err := record.Append(s.RunDir, s.SeatID, "motion-rule", p); err != nil {
				return err
			}
			return seat.Emit(cmd, ruled{ID: id, Ruling: seat.Str(cmd, flags.As)}, nil)
		},
	}
	seat.Prose(c)
	c.Flags().String(flags.ID, "", refHelp(subject))
	c.Flags().String(flags.As, "", e.Usage("REQUIRED — your ruling"))
	return c
}

// appeal: the filer presses on after a ruling.
//
// `contests_ruling` was a bespoke field on ONE of the three exchanges — blue pursuing a direction
// red ruled out-of-scope. Here it is the same act on every subject that has one, which is what
// the collapse buys: re-disputing a rejected grade and pursuing a refused line stop being two
// unrelated mechanisms.
func newAppeal(subject string) *cobra.Command {
	c := &cobra.Command{
		Use:   "appeal",
		Short: "press a " + subject + " motion after a ruling — a ruling is an argument, not a command",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := seat.Of(cmd)
			id := seat.Str(cmd, flags.ID)
			if err := record.RequireRuledMotion(s.RunDir, subject, id); err != nil {
				return err
			}
			reason, err := prose(cmd, "appeal", "why you are pressing on. Going against a ruling without saying why is the disagreement disappearing, which is what the record exists to prevent")
			if err != nil {
				return err
			}
			p := record.NewPayload().Set("motion_id", id).Set("subject", subject).Set("reason", reason)
			if _, err := record.Append(s.RunDir, s.SeatID, "motion-appeal", p); err != nil {
				return err
			}
			return seat.Emit(cmd, appealed{ID: id}, nil)
		},
	}
	seat.Prose(c)
	c.Flags().String(flags.ID, "", refHelp(subject)+" — the motion being appealed, which must already have been ruled")
	return c
}

// refHelp names WHICH id the subject joins on. `direction` has no filing verb, so its id is the
// avenue's; saying "the motion id" there would send a seat looking for an M-number that does not
// exist, and a seat that cannot find the id it was told to pass logs friction and works around the
// verb — losing the capability for the run rather than reporting a wrong flag.
func refHelp(subject string) string {
	if subject == "direction" {
		return "REQUIRED — the AVENUE id (A1, A2 …): a direction's filing is the proposal, so it joins on the avenue's own id, not an M-number"
	}
	return "REQUIRED — the motion id (M1, M2 …)"
}

type filed struct {
	ID      string `json:"motion_id"`
	Subject string `json:"subject"`
}

func (r filed) Human() string { return "motion " + r.ID + " filed (" + r.Subject + ")" }

type ruled struct {
	ID     string `json:"motion_id"`
	Ruling string `json:"ruling"`
}

func (r ruled) Human() string { return "motion " + r.ID + " ruled " + r.Ruling }

type appealed struct {
	ID string `json:"motion_id"`
}

func (r appealed) Human() string { return "motion " + r.ID + " appealed" }
