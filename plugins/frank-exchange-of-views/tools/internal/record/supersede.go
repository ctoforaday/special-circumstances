package record

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE RECORD HOLDS ACTS, AND EVERY ACT STANDS. A seat that recorded something wrong answers it with
// another act, and WHICH act depends on what the first was: a wrong closure is answered by a docket
// motion, a wrong regrade by the next sitting's regrade, a wrong re-run by another re-run.
// superseders is the one statement of that, and it has three readers: the refusal of a second
// once-per-sitting act (requireOncePerSitting), the refusal of a repeated label-keyed act
// (insertNumbered), the refusal of a second ruling (motion.go) — and each recording verb's help
// page, through AnswersTo.
//
// A ROW IS A CLAIM ABOUT THE WRITE PATH. A seat told only "say it in a new act" is told nothing: a
// lens sent that way had to work out for itself that a re-run is answered by another re-run, and a
// row that names an act the write path refuses moves clean text into a log. So every act a row's
// words cover is DRIVEN through the command line by TestARefusedRepeatNamesAnActTheWritePathAdmits
// (internal/cli), member by member, at the time the row states. Word a row from what that test
// observes, never from what the act ought to admit.

// superseder is one row: the act that answers an earlier act of the keyed shape.
type superseder struct {
	// act completes "say so in …".
	act string
	// own marks an act that is the wrong act's OWN VERB, run again. Where a sitting holds one of
	// the act, that verb is refused until the seat's next sitting and the phrase says so
	// (oneASitting reads which, from the declarations deriveKey keys on).
	own bool
	// while is a condition the write path puts on the answering act beyond the sitting, as a
	// clause: the regrade's open gap.
	while string
	// whole is a row whose words fit no pattern; it is rendered as written.
	whole string
	// when says which acts of a verb the row is for, where one verb's help states two rows.
	when string
}

// supersederKey is what decides the answer: the event type — and, for a ruling, the motion's
// subject; for a verify, the key field its body fills (one event type, two verbs, three key
// shapes). The zero subject and the empty field match every act of the type.
type supersederKey struct {
	typ     recordpb.EventType
	subject recordpb.MotionSubject
	keyed   protoreflect.Name
}

const (
	subjectGrade    = recordpb.MotionSubject_MOTION_SUBJECT_GRADE
	subjectDocket   = recordpb.MotionSubject_MOTION_SUBJECT_DOCKET
	subjectPetition = recordpb.MotionSubject_MOTION_SUBJECT_PETITION
)

// superseders holds a row for every type whose second act the record refuses and some command
// records (RefusesARepeat), and for the numbered types whose answer a seat would otherwise guess.
//
// NO ROW IS A STATEMENT TOO, and three are deliberate. The chair holds no act that answers its own
// ruling on an AVENUE: a docket motion names a gap, an avenue is not one, and no verb files a
// motion on an avenue — so that refusal ends without a tail. An anchor, a finding and a mint are
// keyed on an id the tool mints per act, so a second with the same key is a retried write and
// nothing else. No verb records an observe.
//
// A RULING'S ROW IS THE RULER'S, the one seat that can repeat it. What a seat that DISAGREES with a
// ruling does is the same act on a grade — a docket motion on the gap, which the bench rules, or a
// new grade motion while the gap is open — and it reaches that seat through its prompt and the
// `motion` pages, since the record refuses it nothing.
var superseders = map[supersederKey]superseder{
	{typ: recordpb.EventType_EVENT_TYPE_AVENUE}:                                {act: "a move of the avenue"},
	{typ: recordpb.EventType_EVENT_TYPE_AVENUE_REVIEW}:                         {act: "a new review of the avenues", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_CERTIFY}:                               {act: "a new certification", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_CITE}:                                  {act: "a new citation", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_CLOSE}:                                 {act: "a docket motion on the gap"},
	{typ: recordpb.EventType_EVENT_TYPE_CLOSING}:                               {act: "a new closing on the gap", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_DECLARE}:                               {act: "a new declaration", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_HALT}:                                  {act: "a new halt", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_LOG}:                                   {act: "a new log entry", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_MANIFEST_ROW}:                          {act: "a new manifest row for the gap", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_MOTION_RULE, subject: subjectGrade}:    {act: "a docket motion on the gap, or a new grade motion while the gap is open"},
	{typ: recordpb.EventType_EVENT_TYPE_MOTION_RULE, subject: subjectDocket}:   {act: "a new docket motion on the gap"},
	{typ: recordpb.EventType_EVENT_TYPE_MOTION_RULE, subject: subjectPetition}: {act: "a declaration"},
	{typ: recordpb.EventType_EVENT_TYPE_OUTCOME}:                               {act: "a new outcome", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_POSITION}:                              {act: "a new position", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_PROOF}:                                 {act: "a new proof", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_REGRADE}:                               {act: "a new regrade", own: true, while: "while the gap is open"},
	{typ: recordpb.EventType_EVENT_TYPE_REPRODUCE}:                             {act: "a new re-run of the proof, whose reason says which earlier re-run it replaces", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_SPOT_CHECK}:                            {act: "a new spot-check", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_VERDICT}: {whole: "the verdict of your next sitting (a sitting holds one, and this sitting's stands); " +
		"after a PASS no chair sitting follows and nothing answers it"},
	{typ: recordpb.EventType_EVENT_TYPE_VERIFY, keyed: "anchor"}: {act: "a new verification of the citation", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_VERIFY, keyed: "url"}:    {act: "a new corroboration of the source", own: true, when: "for a determination that does not back the claim"},
	// A corroboration that backs the claim is keyed on a label the tool mints, and its repeat is
	// answered idempotently rather than refused: this row reaches a seat through the verb's help.
	{typ: recordpb.EventType_EVENT_TYPE_VERIFY, keyed: "label"}: {whole: "a verification of the citation this corroboration placed, at your next sitting",
		when: "for a determination that backs the claim, which placed a citation"},
}

// Answer is one row as a surface renders it.
type Answer struct {
	// Subject is the motion subject the row is for; unspecified when the type has one row for all.
	Subject recordpb.MotionSubject
	// Keyed is the key field the row is for; empty when the type has one row for all.
	Keyed string
	// Act completes "say so in …".
	Act string
	// When says which acts of a verb the row is for, where one verb's help states two rows.
	When string
}

// AnswerTo is the act that answers a wrong act of this type and body, or false where the table
// holds no row — a refusal then ends without a tail rather than send a seat to an act it has to
// guess.
func AnswerTo(typ recordpb.EventType, body proto.Message) (string, bool) {
	k := supersederKey{typ: typ}
	if b, ok := body.(*recordpb.MotionRule); ok {
		k.subject = b.GetSubject()
	}
	field := keyField(body)
	k.keyed = field
	row, ok := superseders[k]
	if !ok {
		k.keyed = ""
		row, ok = superseders[k]
	}
	if !ok {
		return "", false
	}
	return row.render(typ, field), true
}

// AnswersTo is every row the table holds for a type, in a stable order — what a verb's help reads,
// narrowed by the verb to the subject it rules or the key field its flags fill.
func AnswersTo(typ recordpb.EventType) []Answer {
	var out []Answer
	subjects := append([]recordpb.MotionSubject{recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED}, subjectOrder()...)
	fields := append([]protoreflect.Name{""}, keyFields...)
	for _, s := range subjects {
		for _, f := range fields {
			row, ok := superseders[supersederKey{typ: typ, subject: s, keyed: f}]
			if !ok {
				continue
			}
			field := f
			if field == "" {
				field = typeKeyField(typ)
			}
			out = append(out, Answer{Subject: s, Keyed: string(f), Act: row.render(typ, field), When: row.when})
		}
	}
	return out
}

func subjectOrder() []recordpb.MotionSubject {
	vs := recordpb.MotionSubject(0).Descriptor().Values()
	var out []recordpb.MotionSubject
	for i := 0; i < vs.Len(); i++ {
		if n := recordpb.MotionSubject(vs.Get(i).Number()); n != recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED {
			out = append(out, n)
		}
	}
	return out
}

// render words a row for an act keyed on field ("" for a numbered or singleton act).
func (s superseder) render(typ recordpb.EventType, field protoreflect.Name) string {
	if s.whole != "" {
		return s.whole
	}
	if !s.own {
		return s.act
	}
	held, per := oneASitting(typ, field)
	if !held {
		return s.act
	}
	out := s.act + " at your next sitting"
	if s.while != "" {
		out += ", " + s.while
	}
	if per == "" {
		return out + " (a sitting holds one, and this sitting's stands)"
	}
	return out + " (a sitting holds one per " + per + ", and this sitting's stands)"
}

// SupersedingAnswer words the tail every refusal of a repeated act ends with, or "" where the
// table holds no row.
func SupersedingAnswer(typ recordpb.EventType, body proto.Message) string {
	act, ok := AnswerTo(typ, body)
	if !ok {
		return ""
	}
	return "The first stands. If it was wrong, say so in " + act + "; the record is append-only and both stay visible"
}

// oneASitting answers whether a sitting holds ONE act of this type keyed on field, and the flag
// that names what it holds one PER ("" for a singleton). It is read from what deriveKey keys on,
// never restated: a singleton, or a label key that references something rather than defining it.
func oneASitting(typ recordpb.EventType, field protoreflect.Name) (held bool, per string) {
	switch {
	case singleton[typ]:
		return true, ""
	case field == "" || defines[typ]:
		return false, ""
	}
	md, ok := bodyDescriptor(recordpb.Word(typ))
	if !ok {
		return false, ""
	}
	fd := md.Fields().ByName(field)
	if fd == nil {
		return false, ""
	}
	return true, "--" + flagOf(fd)
}

// RefusesARepeat answers whether the record refuses a second act of this type in one sitting on
// the strength of its KEY: a singleton, or a body that can carry a key field and references what
// the field names rather than defining it. Derived from the three declarations deriveKey reads
// (singleton, keyFields, defines), so a type that joins one of them owes a row in superseders
// without a list to extend — a verb that records such a type and finds no row refuses to mount
// (cli/seat teachAnswer).
func RefusesARepeat(typ recordpb.EventType) bool {
	if singleton[typ] {
		return true
	}
	return !defines[typ] && typeKeyField(typ) != ""
}

// typeKeyField is the first key field a body of this type can carry, or "".
func typeKeyField(typ recordpb.EventType) protoreflect.Name {
	md, ok := bodyDescriptor(recordpb.Word(typ))
	if !ok {
		return ""
	}
	for _, name := range keyFields {
		if fd := md.Fields().ByName(name); fd != nil && fd.Kind() == protoreflect.StringKind && !fd.IsList() {
			return name
		}
	}
	return ""
}

// keyField is the key field this body fills — keyLabel's field rather than its value — or "".
func keyField(body proto.Message) protoreflect.Name {
	if body == nil {
		return ""
	}
	m := body.ProtoReflect()
	fds := m.Descriptor().Fields()
	for _, name := range keyFields {
		fd := fds.ByName(name)
		if fd == nil || fd.Kind() != protoreflect.StringKind || fd.IsList() {
			continue
		}
		if m.Get(fd).String() != "" {
			return name
		}
	}
	return ""
}

// flagOf is the word a seat types for a field: the field's own `(sql).flag` when it declares one,
// else the payload-key map, which knows that a gap id is typed --id and a status --as. (FlagFor's
// fallback is the field name itself, which would name --gap-id — a flag no verb has.)
//
// AN ENUM ARM OF A ONEOF IS TYPED --as. A ruling's word lands on MotionRule.grade, .petition or
// .direction — one arm per subject — and every one of them is set through --as; named by its field,
// a refusal would send the seat to --petition, which no verb has.
func flagOf(fd protoreflect.FieldDescriptor) string {
	if o, _ := proto.GetExtension(fd.Options(), recordpb.E_Sql).(*recordpb.Sql); o.GetFlag() != "" {
		return o.GetFlag()
	}
	if fd.ContainingOneof() != nil && fd.Enum() != nil {
		return flags.As
	}
	return flags.ForPayloadKey(string(fd.Name()))
}
