package blue

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// retire: take one anchor out of the report; its sentence stays word for word.
//
// A claim leaves with its sentence, through `edit`: a cut of whole sentences takes their anchors and
// records them with the edit's reason (removed_anchors), and a rewrite carries every anchor of its
// span, so prose may be compacted, merged and reorganized freely and an anchor leaves only where an act
// names it — this verb, or the cut that took its sentence.
//
// A RETIRE TAKES ONE ANCHOR OFF A SENTENCE THAT STAYS: a citation whose source cannot be read, a
// computation judged unsound, a finding, or a closed gap's marker. An open gap's own marker is refused
// (F14): a cut of its sentence is its raiser's turn.
func newRetire() *cobra.Command {
	named := flags.IDList(flags.AnchorID)
	c := seat.Prose(seat.New("retire", retireResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		if _, err := s.Run(); err != nil {
			return nil, err
		}
		// --reason is the prose channel (Prose provides it); it lands under `reason`, which validate
		// requires — an anchor leaves the report only with its reason.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body := &recordpb.Retire{
			Reason:  proto.String(why),
			Answers: proto.String(seat.Str(cmd, flags.Answers)),
		}
		return retireAnchor(s, body, named.Value())
	}))

	// SUPPLIED, NOT TYPED: the record's claim is the sentence the anchor stood on, which the tool reads
	// off the report.
	seat.Supplies(c, "claim", "the sentence the retired anchor stood on, read off the report; it stays word for word")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap id this retire responds to — the provenance join key an edit carries, so the gap's change log lists the retire beside what red asked for; omit only for a retire that answers no gap")
	c.Flags().Var(named, flags.Anchor, "the one anchor `id`, of any kind, to retire while its sentence stays in the report word for word. An open gap's own marker is refused: carry it, or cut its sentence with an edit, which hands the gap to its lens. A citation a computation still in the report names as its method is refused while that computation stands: take the computation out first, by its own retire or a cut of its sentence")
	return seat.Require(c, flags.Anchor)
}

// retireAnchor takes one anchor of any kind off its sentence, which stays word for word. IT IS HOW BLUE
// ANSWERS AN ANCHOR THAT CANNOT STAND ON A SENTENCE THAT CAN (gblock, 2026-10-25; every kind under
// P12b): a rewrite carries every anchor, and a cut takes the sentence too. THE TOOL READS THE SENTENCE:
// the claim the record holds is the sentence holding the anchor in the report the retire lands on
// (bluedoc.SentenceAround, as `verify` reads it). An open gap's own marker is refused (F14), and a
// citation a live computation applies is refused in the write (ruling L6).
func retireAnchor(s seat.Context, body *recordpb.Retire, named []string) (seat.Result, error) {
	// A BLANK --anchor NAMES NOTHING: the list flag reads "" as the empty list, and a retire of no
	// anchor is the omitted flag, answered as one (universe-m29: a blank value answered as a named one).
	if len(named) == 0 {
		return nil, feov.Errorf(feov.MissingField, "requires --anchor: the one anchor to take off its sentence, which stays word for word — a retire naming no anchor takes nothing out")
	}
	if len(named) != 1 {
		return nil, feov.Errorf(feov.Validation, "--anchor retires one anchor and keeps its sentence, and this names %d: %s. Retire each in its own retire, or cut the sentence they stand on with `edit`", len(named), strings.Join(named, ", "))
	}
	id := named[0]
	noun, tok := record.EvidenceNoun(id), anchor.Token(id)
	guard := func(l *record.Landing) error {
		md, err := l.Report()
		if err != nil {
			return feov.Wrap(feov.Internal, err, "the report could not be read from the record to find the sentence %s %s stands on, so nothing is recorded: %v", noun, id, err)
		}
		sentence, ok := bluedoc.SentenceAround(md, tok)
		if !ok {
			return feov.Errorf(feov.Conflict, "the report holds no %s — %s %s is not on a sentence of the report, so there is no %s to retire from one", tok, noun, id, noun)
		}
		body.Claim = proto.String(sentence)
		body.Anchors = []string{id}
		return nil
	}
	if _, err := record.Append(s.Identity(), body, record.Check(guard)); err != nil {
		return nil, err
	}
	return retireResult{Claim: body.GetClaim(), Anchors: body.Anchors, Noun: noun, Answers: body.GetAnswers()}, nil
}

// quoteCore is a quote or span reduced to what it says: anchors out, the list marker or block-quote
// marker it opens with out — those lie outside every sentence (anchor.Sentences) — then surrounding
// whitespace and emphasis markers and trailing punctuation trimmed until nothing more comes off.
// Internal text is untouched — "the sky" and "sky" stay different claims.
func quoteCore(s string) string {
	s = claimcount.StripAnchors(s)
	if sp := anchor.Sentences(s); len(sp) > 0 {
		s = s[sp[0][0]:]
	}
	for {
		t := strings.TrimRight(strings.Trim(s, " \t\n*_"), anchortext.TrailingPunct)
		if t == s {
			return s
		}
		s = t
	}
}

type retireResult struct {
	// Claim is the sentence the anchor stood on, which stays; Noun is the anchor's kind as
	// EvidenceTables names it.
	Claim   string   `json:"claim"`
	Anchors []string `json:"anchors,omitempty"`
	Noun    string   `json:"-"`
	// Answers is the gap a retire answers, as recorded; empty when it answers none.
	Answers string `json:"answers,omitempty"`
}

func (r retireResult) Human() string {
	answering := ""
	if r.Answers != "" {
		answering = ", answering gap " + r.Answers
	}
	return "retired " + r.Noun + " " + strings.Join(r.Anchors, ", ") + answering + "; its sentence stays in the report word for word, backed by whatever else backs it: " + r.Claim
}
