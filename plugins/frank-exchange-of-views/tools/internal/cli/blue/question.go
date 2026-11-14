package blue

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// question: the parts of the question the run was asked, and how the report covers each (ruled by
// gblock 2026-11-11, E and J). Whether the report answered its question lived in prose nothing could
// refuse, and a run passed VERIFIED over an unanswered quarter of it. On the record, each part has a
// coverage the logic lens checks, a PASS over a part not covered derives QUALIFIED, and the tool
// states each such part in report.md from its coverage.
//
// TWO ACTS, AND THEY ARE TWO VERBS, as an inquiry's are: `part` splits a part out with its text, and
// `cover` records its coverage with the reason. The logic lens's `question check` is the group's
// third act, on its own surface.
func newQuestion() *cobra.Command {
	c := &cobra.Command{
		Use:          "question",
		Short:        "split the question the run was asked into question parts, and record how the report covers each — `split` splits one out, `cover` records its coverage",
		SilenceUsage: true,
	}
	c.AddCommand(newQuestionSplit(), newQuestionCover())
	return c
}

func newQuestionSplit() *cobra.Command {
	c := seat.Records(seat.NewKeyed("split", "question-split", questionPartResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		text := strings.TrimRight(seat.Str(cmd, flags.PartText), "\n")
		id := record.NewID("question part")
		if _, err := record.Append(s.Identity(), &recordpb.QuestionPart{PartId: proto.String(id), Text: proto.String(text)}); err != nil {
			return nil, err
		}
		return questionPartResult{ID: id, Text: text, VoiceTells: spanVoiceTells(text)}, nil
	}), "question_part")
	seat.Supplies(c, "coverage", "a part is split out with its text alone, and has no coverage until `question cover` records one")
	flags.Text(c, flags.PartText, "the PART of the question this names, as a reader of the SUBJECT would ask it answered — a standard, a population, a period or an alternative the question names, in the subject's terms. Where the report does not answer it, report.md prints it word for word in the sentence that says so")
	seat.Require(c, flags.PartText)
	return c
}

func newQuestionCover() *cobra.Command {
	c := seat.Prose(seat.Records(seat.NewKeyed("cover", "question-cover", questionPartResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		id := seat.Str(cmd, flags.ID)
		cov, ok := record.PartCoverageOf(seat.Str(cmd, flags.As))
		if !ok {
			return nil, feov.Errorf(feov.Validation, "%q is not a coverage a question part can take", seat.Str(cmd, flags.As))
		}
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body := &recordpb.QuestionPart{PartId: proto.String(id), SupersedesCoverage: proto.String("1"), Coverage: &cov, Reason: proto.String(why)}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return questionPartResult{ID: id, Coverage: recordpb.Word(cov), Recorded: true}, nil
	}), "question_part"))
	c.Flags().Var(flags.QuestionPartID().WithCheck(record.QuestionPartExists), flags.ID, "the question part whose coverage you record — its `S-` id, which its split returned and show question lists with every part's coverage")
	seat.Require(c, flags.ID)
	enumhelp.Flag(c, flags.As, record.MustEnum("question_part", "coverage"), "how the report covers this part")
	seat.Require(c, flags.As)
	c.Flags().Lookup(flags.Reason).Usage = "why the coverage holds — what in the report answers the part, which sources could not be obtained and what was tried, or why the part lies outside the report's answer. The logic lens judges the coverage against it; it is printed on the record's surfaces, never in report.md"
	return seat.SaysRequired(c, flags.Reason)
}

type questionPartResult struct {
	ID       string `json:"part_id"`
	Text     string `json:"text,omitempty"`
	Coverage string `json:"coverage,omitempty"`
	Recorded bool   `json:"recorded,omitempty"`
	// VoiceTells is ADVICE: the part's text is printed in report.md where the report does not answer
	// it, written for a reader of the subject. It refuses nothing.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r questionPartResult) Human() string {
	if r.Recorded {
		return "question part " + r.ID + " recorded as " + r.Coverage
	}
	head := "question part " + r.ID + " recorded: " + r.Text
	if len(r.VoiceTells) == 0 {
		return head
	}
	return head + "\n\nNOTE — this text is printed in report.md where the report does not answer the part, and in\n" +
		"places sounds like the run rather than the subject. It is recorded; this is not a refusal:\n  - " +
		strings.Join(r.VoiceTells, "\n  - ")
}
