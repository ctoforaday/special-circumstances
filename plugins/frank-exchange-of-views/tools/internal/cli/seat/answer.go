package seat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// eventTypeOf resolves an event word to its type.
func eventTypeOf(word string) (recordpb.EventType, bool) {
	vs := recordpb.EventType(0).Descriptor().Values()
	for i := 0; i < vs.Len(); i++ {
		t := recordpb.EventType(vs.Get(i).Number())
		if word != "" && recordpb.Word(t) == word {
			return t, true
		}
	}
	return 0, false
}

// keysOnKey annotates a verb with the key fields its act can be keyed on, where the event type it
// records has more than one key shape.
const keysOnKey = "feov.keys_on"

// KeysOn declares which key fields this verb's act can be keyed on. One event type written by two
// verbs with different keys is answered differently per key (record's superseders), and the verb
// is where the key is decided — `verify` names a citation, `corroborate` a source, and the tool
// mints a label when that source backs the claim. teachAnswer reads it to show a verb its own
// rows; TestEveryHelpAnswerIsAdmitted drives each declared field and fails one the verb's act is
// not keyed on.
func KeysOn(c *cobra.Command, fields ...string) *cobra.Command {
	annotate(c, keysOnKey, strings.Join(fields, ","))
	return c
}

// answerTaughtKey annotates a command whose help already carries its answer paragraph.
const answerTaughtKey = "feov.answer_taught"

// AnswerLead opens the paragraph a recording verb's help ends with.
const AnswerLead = "IF WHAT YOU RECORDED WAS WRONG"

// teachAnswer ends a recording verb's help with the act that answers a wrong act of the type it
// records: an act stands as filed, so the page that teaches the verb says what a seat does about
// one it got wrong. The words are the record's (record.AnswersTo), narrowed to what the verb
// fixes — the subject of a ruling or an appeal, the key fields of a verify. A verb whose type holds
// no row gets no paragraph.
//
// IT PANICS when a verb records a type whose repeat the record refuses and the table holds no row
// for it: that refusal would end by sending the seat nowhere, and the mount is where a new verb
// finds out.
func teachAnswer(c *cobra.Command, role string) {
	typ, ok := eventTypeOf(RecordType(c))
	if !ok || c.Annotations[answerTaughtKey] != "" {
		return
	}
	rows := AnswersFor(c, role)
	if len(rows) == 0 {
		if record.RefusesARepeat(typ) {
			panic(fmt.Sprintf("seat: %s records a %s, whose second act in a sitting the record refuses, and no row says what answers a wrong one — add it to record's superseders",
				c.CommandPath(), recordpb.Word(typ)))
		}
		return
	}
	var b strings.Builder
	b.WriteString(AnswerLead)
	for i, r := range rows {
		switch {
		case len(rows) == 1:
			b.WriteString(", say so in " + r.Act)
		case i == 0:
			b.WriteString(": " + r.When + ", say so in " + r.Act)
		default:
			b.WriteString("; " + r.When + ", say so in " + r.Act)
		}
	}
	b.WriteString(".")
	c.Long = strings.TrimRight(c.Long, "\n") + "\n\n" + wrapHelp(b.String(), 100) + "\n"
	annotate(c, answerTaughtKey, "true")
}

// AnswersFor is the rows of the answers table this verb's help states: every row of the type it
// records, less the rows for a motion subject the verb does not rule or appeal, for a key field
// its act is not keyed on, and for an answering act the surface's role does not hold.
func AnswersFor(c *cobra.Command, role string) []record.Answer {
	typ, ok := eventTypeOf(RecordType(c))
	if !ok {
		return nil
	}
	subject := recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED
	if p := c.Parent(); p != nil {
		subject, _ = record.MotionSubjectEnum(p.Name())
	}
	keys := map[string]bool{}
	for _, k := range strings.Split(c.Annotations[keysOnKey], ",") {
		keys[k] = true
	}
	var out []record.Answer
	for _, r := range record.AnswersTo(typ) {
		if r.Subject != recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED && r.Subject != subject {
			continue
		}
		if r.Keyed != "" && !keys[r.Keyed] {
			continue
		}
		if r.Holder != "" && r.Holder != role {
			continue
		}
		out = append(out, r)
	}
	return out
}

// wrapHelp breaks a paragraph at word boundaries so no line runs past width.
func wrapHelp(s string, width int) string {
	var out, line strings.Builder
	for _, w := range strings.Fields(s) {
		if line.Len() > 0 && line.Len()+1+len(w) > width {
			out.WriteString(line.String())
			out.WriteString("\n")
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteString(" ")
		}
		line.WriteString(w)
	}
	out.WriteString(line.String())
	return out.String()
}
