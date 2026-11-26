package record

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
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
)

// liveSentenceIn is the sentence holding id's token in report, read the way an edit reads it to
// reopen the anchor (bluedoc.SentenceAround), or why there is none. No writer leaves an anchor bare —
// a cut takes the anchors of the sentences it cuts — so a token in the report stands on a sentence.
func liveSentenceIn(report string, id string) (string, sentenceGone) {
	sentence, ok := bluedoc.SentenceAround(report, anchor.Token(id))
	if !ok {
		return "", sentenceAbsent
	}
	return sentence, sentenceLive
}

// SentencesIn is the live sentence of each of ids in report (liveSentenceIn), keyed by id, read once
// for all of them; an id not in the map is absent from the report.
func SentencesIn(report string, ids []string) (live map[string]string) {
	live = map[string]string{}
	for _, id := range ids {
		if sentence, gone := liveSentenceIn(report, id); gone == sentenceLive {
			live[id] = sentence
		}
	}
	return live
}

// LiveSentence is the sentence evidence anchor id backs in the report as it renders now — the claim
// a check of it records — or the refusal of a check of an anchor with none, naming the anchor by its
// kind's noun (recordsql.EvidenceTables). A retired anchor's refusal names its retire: the event,
// its reason, and whether the sentence was kept.
//
// THE CLAIM IS THE SENTENCE THE SEAT JUDGED, read before the write; the check lands only on that
// sentence (CheckOfSentence).
func LiveSentence(run Run, id string) (string, error) {
	return liveSentence(id, func() (string, error) { return RenderRun(run) }, func() (AnchorRemoval, bool, error) { return AnchorRemovedAt(run, id) }, beforeCheck())
}

// CheckOfSentence is the guard a check of evidence anchor id, which judged claim (LiveSentence), lands
// under: the anchor's live sentence in the report the check lands on must be that sentence
// (bluedoc.SameSentence). An edit that changes or cuts it, or a retire of the anchor, committing
// between the read and the write refuses the check, naming what the report now holds, so a check is
// recorded only of the sentence the report holds at its event — the tool never keeps a judgement of a
// sentence the seat did not read. Asked of the writing transaction, so the check and the write racing
// it cannot both land.
func CheckOfSentence(id, claim string) Guard {
	return Check(func(l *Landing) error {
		_, err := liveSentence(id, l.Report, func() (AnchorRemoval, bool, error) { return anchorRemovedAt(l.w.q, id) }, inLanding(claim))
		return err
	})
}

// sentenceRead is which read of an anchor's sentence liveSentence makes: the read before a check
// (beforeCheck), or the read in its landing (inLanding), which holds the sentence the check judged.
type sentenceRead struct {
	landing bool
	judged  string
}

func beforeCheck() sentenceRead            { return sentenceRead{} }
func inLanding(judged string) sentenceRead { return sentenceRead{landing: true, judged: judged} }

// liveSentence is LiveSentence over a report and a removal lookup: the run's now, or a landing's. In
// the landing the sentence must be the one the check judged, and the refusals say the report changed
// while the check was made.
func liveSentence(id string, render func() (string, error), removedAt func() (AnchorRemoval, bool, error), read sentenceRead) (string, error) {
	row, ok := recordsql.EvidenceTableOf(anchor.Kind(id))
	if !ok {
		return "", feov.Errorf(feov.Validation, "%s is not an evidence anchor — a check reads the sentence a citation or a computation backs", id)
	}
	report, err := render()
	if err != nil {
		return "", feov.Wrap(feov.Conflict, err, "the report does not render, so the sentence %s %s marks cannot be read, and nothing can be checked until it does: %v. "+
			"The board's anomalies say why; record that you could not check it with `log`", row.Noun, id, err)
	}
	// while is how a refusal in the landing opens: the check read a live sentence, and the report
	// changed under it.
	while := ""
	if read.landing {
		while = "while this check was made, "
	}
	sentence, gone := liveSentenceIn(report, id)
	switch gone {
	case sentenceAbsent:
		r, found, err := removedAt()
		if err != nil {
			return "", err
		}
		if found {
			return "", removedRefusal(while, row.Noun, id, r)
		}
		if read.landing {
			return "", feov.Errorf(feov.Conflict, "while this check was made, %s %s's anchor left the report, so the report holds no sentence for it and nothing is recorded. "+
				"Check the %ss whose anchors the report carries", row.Noun, id, row.Noun)
		}
		return "", feov.Errorf(feov.Conflict, "%s %s's anchor is not in the report — the anchor never placed — so the report holds no sentence for it and there is nothing to check. "+
			"Check the %ss whose anchors the report carries", row.Noun, id, row.Noun)
	}
	if read.landing && !bluedoc.SameSentence(sentence, read.judged) {
		return "", feov.Errorf(feov.Conflict, "while this check was made, the sentence holding %s %s's anchor changed: the report now holds %s, not %s, the sentence this check read. "+
			"Nothing is recorded: read the sentence as it now stands and judge it again", row.Noun, id, quoted(sentence), quoted(read.judged))
	}
	return sentence, nil
}

// removedRefusal refuses a check of evidence anchor id, taken out of the report by r; while opens it.
func removedRefusal(while, noun, id string, r AnchorRemoval) error {
	return feov.Errorf(feov.Conflict, "%s%s %s is out of the report — %s — so it is in no document and the report holds no sentence for it to check. "+
		"Check the %ss whose anchors the report carries", while, noun, id, r.Said(), noun)
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
