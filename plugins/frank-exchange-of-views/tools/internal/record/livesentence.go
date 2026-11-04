package record

import (
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A CHECK NEEDS A LIVE SENTENCE. A check of an evidence anchor — `verify` of a citation, `reproduce`
// of a computation — is a judgement of the sentence the anchor backs, so an anchor with no such
// sentence has nothing to judge: its token is not in the report (blue retired it, or it never
// placed), or it stands bare (the sentence it backed was cut). One predicate answers it for every
// kind, and the work list asks the same one for why no check can reach an anchor (unverifiableIn).

// sentenceGone is why an evidence anchor has no live sentence in a rendered report, or sentenceLive.
type sentenceGone int

const (
	sentenceLive sentenceGone = iota
	// sentenceAbsent: the anchor's token is not in the report.
	sentenceAbsent
	// sentenceBare: the token stands with no prose before it in its sentence.
	sentenceBare
)

// liveSentenceIn is the sentence holding id's token in report, read the way an edit reads it to
// reopen the anchor (bluedoc.SentenceAround), or why there is none. bare is the report's bare
// anchors (claimcount.BareAnchorIDs), which a caller asking of many anchors computes once.
func liveSentenceIn(report string, bare []string, id string) (string, sentenceGone) {
	sentence, ok := bluedoc.SentenceAround(report, anchor.Token(id))
	switch {
	case !ok:
		return "", sentenceAbsent
	case slices.Contains(bare, id):
		return "", sentenceBare
	}
	return sentence, sentenceLive
}

// LiveSentence is the sentence evidence anchor id backs in the report as it renders now — the claim
// a check of it records — or the refusal of a check of an anchor with none, naming the anchor by its
// kind's noun (recordsql.EvidenceTables). A retired anchor's refusal names its retire: the event,
// its reason, and whether the sentence was kept.
func LiveSentence(run Run, id string) (string, error) {
	row, ok := recordsql.EvidenceTableOf(anchor.Kind(id))
	if !ok {
		return "", feov.Errorf(feov.Validation, "%s is not an evidence anchor — a check reads the sentence a citation or a computation backs", id)
	}
	report, err := RenderRun(run)
	if err != nil {
		return "", feov.Wrap(feov.Conflict, err, "the report does not render, so the sentence %s %s marks cannot be read, and nothing can be checked until it does: %v. "+
			"The board's anomalies say why; record that you could not check it with `log`", row.Noun, id, err)
	}
	sentence, gone := liveSentenceIn(report, claimcount.BareAnchorIDs(report), id)
	switch gone {
	case sentenceAbsent:
		r, found, err := AnchorRetiredAt(run, id)
		if err != nil {
			return "", err
		}
		if found {
			return "", feov.Errorf(feov.Conflict, "%s %s is retired — %s — so it is in no document and the report holds no sentence for it to check. "+
				"Check the %ss whose anchors the report carries", row.Noun, id, r.Said(), row.Noun)
		}
		return "", feov.Errorf(feov.Conflict, "%s %s's anchor is not in the report — the anchor never placed — so the report holds no sentence for it and there is nothing to check. "+
			"Check the %ss whose anchors the report carries", row.Noun, id, row.Noun)
	case sentenceBare:
		return "", feov.Errorf(feov.Conflict, "%s %s's anchor stands bare in the report: the sentence it backed was cut, and it backs no prose, so there is nothing to check. "+
			"The anchor leaves with blue's retire of its claim", row.Noun, id)
	}
	return sentence, nil
}

// EvidenceNoun is the word an evidence anchor's kind names it by (recordsql.EvidenceTables), and the
// kind's own name for an anchor of a kind that backs nothing.
func EvidenceNoun(id string) string {
	if row, ok := recordsql.EvidenceTableOf(anchor.Kind(id)); ok {
		return row.Noun
	}
	return anchor.Kind(id)
}

// quoted is s in double quotes, flattened to one line.
func quoted(s string) string { return `"` + strings.Join(strings.Fields(s), " ") + `"` }

// reasonOf states a retire's recorded reason, or that it recorded none.
func reasonOf(reason string) string {
	if strings.TrimSpace(reason) == "" {
		return "no reason recorded"
	}
	return "reason: " + quoted(reason)
}
