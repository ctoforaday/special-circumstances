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

// LiveSentencesIn is the live sentence of each of ids in report (liveSentenceIn), keyed by id: the
// sentence each backs as the report renders, read once for all of them. An id with none — absent
// from report, or standing bare — has no key.
func LiveSentencesIn(report string, ids []string) map[string]string {
	live, _ := SentencesIn(report, ids)
	return live
}

// SentencesIn is LiveSentencesIn with why an id has no live sentence: bare holds each of ids whose
// token stands bare in report (liveSentenceIn's sentenceBare), so an id in neither map is absent from
// it. One bare rule for every reader that tells the two apart.
func SentencesIn(report string, ids []string) (live map[string]string, bare map[string]bool) {
	live, bare = map[string]string{}, map[string]bool{}
	bareIDs := claimcount.BareAnchorIDs(report)
	for _, id := range ids {
		switch sentence, gone := liveSentenceIn(report, bareIDs, id); gone {
		case sentenceLive:
			live[id] = sentence
		case sentenceBare:
			bare[id] = true
		}
	}
	return live, bare
}

// LiveSentence is the sentence evidence anchor id backs in the report as it renders now — the claim
// a check of it records — or the refusal of a check of an anchor with none, naming the anchor by its
// kind's noun (recordsql.EvidenceTables). A retired anchor's refusal names its retire: the event,
// its reason, and whether the sentence was kept.
//
// THE CLAIM IS THE SENTENCE THE SEAT JUDGED, read before the write: an edit committing between this
// read and the check's landing leaves the check on the sentence before it, which the edit names
// reopened, so the check does not stand and evidence.md says which sentence it was of. A retire
// committing there is refused in the landing (CheckNotRetired).
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
			return "", retiredRefusal(row.Noun, id, r)
		}
		return "", feov.Errorf(feov.Conflict, "%s %s's anchor is not in the report — the anchor never placed — so the report holds no sentence for it and there is nothing to check. "+
			"Check the %ss whose anchors the report carries", row.Noun, id, row.Noun)
	case sentenceBare:
		return "", feov.Errorf(feov.Conflict, "%s %s's anchor stands bare in the report: the sentence it backed was cut, and it backs no prose, so there is nothing to check. "+
			"The anchor leaves with blue's retire of its claim", row.Noun, id)
	}
	return sentence, nil
}

// CheckNotRetired is the guard a check of evidence anchor id lands under: a retire of the anchor
// committing between the check's read of its sentence (LiveSentence) and its write refuses the check,
// as LiveSentence refuses a check of a retired anchor. Asked of the writing transaction by
// retire_anchors_value, so the check and the retire racing it cannot both land.
func CheckNotRetired(id string) Guard {
	return Check(func(l *Landing) error {
		r, found, err := anchorRetiredAt(l.w.q, id)
		if err != nil || !found {
			return err
		}
		return retiredRefusal(EvidenceNoun(id), id, r)
	})
}

// retiredRefusal refuses a check of evidence anchor id, retired by r.
func retiredRefusal(noun, id string, r AnchorRetire) error {
	return feov.Errorf(feov.Conflict, "%s %s is retired — %s — so it is in no document and the report holds no sentence for it to check. "+
		"Check the %ss whose anchors the report carries", noun, id, r.Said(), noun)
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
