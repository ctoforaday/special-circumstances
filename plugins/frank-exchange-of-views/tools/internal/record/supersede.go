package record

import (
	"slices"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE RECORD HOLDS ACTS, AND EVERY ACT STANDS. A seat that recorded something wrong answers it with
// another act, and WHICH act depends on what the first was: a wrong closure is answered by a docket
// motion, a wrong regrade by the next sitting's regrade, a wrong re-run by another re-run.
// superseders is the one statement of that, and it has three readers: the refusal of a second
// once-per-sitting act (requireOncePerSitting), the refusal of a repeated label-keyed act
// (writeIn), the refusal of a second ruling (motion.go) — and each recording verb's help
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
	// role is the answering seat's role, where one act's answer differs by who took it: the chair's
	// carry restates a closure it cannot answer with a gap. Empty matches every role.
	role string
}

const (
	subjectDocket   = recordpb.MotionSubject_MOTION_SUBJECT_DOCKET
	subjectPetition = recordpb.MotionSubject_MOTION_SUBJECT_PETITION
	subjectPosition = recordpb.MotionSubject_MOTION_SUBJECT_POSITION
)

// superseders holds a row for every type whose second act the record refuses and some command
// records (RefusesARepeat), and for the numbered types whose answer a seat would otherwise guess.
//
// NO ROW IS A STATEMENT TOO, and two are deliberate. An anchor and a mint are keyed on an id the
// tool mints per act, so a second with the same key is a retried write and nothing else. No
// verb records an observe. An inquiry's check keys on no label, and a second check while one stands
// is refused by its own rule (requireInquiryCheck), which names what makes it owed again.
//
// A RULING'S ROW IS THE RULER'S, the one seat that can repeat it — and the ruler files nothing on
// the subject it rules (rulerfiles.go), so no row sends it to a new motion of that subject. Nothing
// sits above the bench, so it answers its own ruling on a docket motion as it answers one on a
// petition, in a declaration: a holding every seat is given, which moves no gap — the gap is before
// the bench again only on a new docketing of it, and the row says so.
var superseders = map[supersederKey]superseder{
	{typ: recordpb.EventType_EVENT_TYPE_INQUIRY}: {act: "a new record of the inquiry's fate"},
	{typ: recordpb.EventType_EVENT_TYPE_CERTIFY}: {act: "a new certification", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_CITE}:    {act: "a new citation", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_CLOSE}: {act: "a new gap on the defect where it still stands in the report: " +
		"a closed gap stays closed, and no mint names it"},
	{typ: recordpb.EventType_EVENT_TYPE_CLOSE, role: "chair"}: {act: "a log entry: a carry restates a closure the record holds, " +
		"which stands as its lens made it, and a defect still in the report is a new gap of the lens that finds it"},
	{typ: recordpb.EventType_EVENT_TYPE_CLOSING}: {act: "a new closing on the gap", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_DECLARE}: {act: "a new declaration", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_HALT}:    {act: "a new halt", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_LOG}:     {act: "a new log entry", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_MOTION_RULE, subject: subjectDocket}: {act: "a declaration, which moves no gap: " +
		"the gap comes before you again only on a new docketing of it"},
	{typ: recordpb.EventType_EVENT_TYPE_MOTION_RULE, subject: subjectPetition}: {act: "a declaration"},
	{typ: recordpb.EventType_EVENT_TYPE_MOTION_RULE, subject: subjectPosition}: {act: "a declaration"},
	{typ: recordpb.EventType_EVENT_TYPE_OUTCOME}:                               {whole: "a certification: the outcome stands once it is recorded, and after it the record takes no seat act but your register for the assemble occasion, your declaration, your certification, the assembly and a log entry"},
	{typ: recordpb.EventType_EVENT_TYPE_POSITION}:                              {act: "a new position", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_COMPUTATION}:                           {act: "a new computation", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_REGRADE}:                               {act: "a new regrade", own: true, while: "while the gap is open"},
	{typ: recordpb.EventType_EVENT_TYPE_AMEND}:                                 {act: "a new amend", own: true, while: "while the gap is open"},
	{typ: recordpb.EventType_EVENT_TYPE_REPRODUCE}:                             {act: "a new judgement of the same computation anchor — a seat's latest judgement of a sentence is its verdict on it", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_SPOT_CHECK}:                            {act: "a new spot-check", own: true},
	{typ: recordpb.EventType_EVENT_TYPE_VERDICT}:                               {whole: "a log entry: a PASS ends the debate, so no chair sitting follows it and no act answers it"},
	{typ: recordpb.EventType_EVENT_TYPE_VERIFY, keyed: "anchor"}:               {act: "a new verification of the citation", own: true, when: "for a verification of a citation"},
	// A reading of a source the lens found carries no key field (url keys nothing): its repeat in a
	// sitting is refused by requireNotARetriedReading, which reads this unkeyed row.
	{typ: recordpb.EventType_EVENT_TYPE_VERIFY}: {whole: "a new reading of that source for this sentence, at your next sitting — a reading of another source, or of this source for another sentence, is recorded now",
		when: "for a reading of a source you found"},
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
	// Role is the answering seat's role the row is for; empty when the row is every role's.
	Role string
}

// Clause is the row as a sentence part, after "if it was wrong, …": the act to say so in.
func (a Answer) Clause() string {
	return "say so in " + a.Act
}

// AnswerTo is the act that answers a wrong act of this type and body, or false where the table
// holds no row — a refusal then ends without a tail rather than send a seat to an act it has to
// guess.
func AnswerTo(role string, typ recordpb.EventType, body proto.Message) (Answer, bool) {
	k := supersederKey{typ: typ, role: role}
	if b, ok := body.(*recordpb.MotionRule); ok {
		k.subject = b.Subject()
	}
	field := keyField(body)
	for _, r := range []string{role, ""} {
		for _, f := range []protoreflect.Name{field, ""} {
			k.role, k.keyed = r, f
			if row, ok := superseders[k]; ok {
				return Answer{Subject: k.subject, Keyed: string(k.keyed), Act: row.render(typ, field), When: row.when, Role: r}, true
			}
		}
	}
	return Answer{}, false
}

// AnswersTo is every row the table holds for a type, in a stable order — what a verb's help reads,
// narrowed by the verb to the subject it rules or the key field its flags fill.
func AnswersTo(typ recordpb.EventType) []Answer {
	var out []Answer
	subjects := append([]recordpb.MotionSubject{recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED}, subjectOrder()...)
	fields := append([]protoreflect.Name{""}, keyFields...)
	roles := []string{""}
	for k := range superseders {
		if k.typ == typ && k.role != "" && !slices.Contains(roles, k.role) {
			roles = append(roles, k.role)
		}
	}
	slices.Sort(roles[1:])
	for _, s := range subjects {
		for _, f := range fields {
			for _, r := range roles {
				row, ok := superseders[supersederKey{typ: typ, subject: s, keyed: f, role: r}]
				if !ok {
					continue
				}
				field := f
				if field == "" {
					field = typeKeyField(typ)
				}
				out = append(out, Answer{Subject: s, Keyed: string(f), Act: row.render(typ, field), When: row.when, Role: r})
			}
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
func SupersedingAnswer(seatID string, typ recordpb.EventType, body proto.Message) string {
	a, ok := AnswerTo(seatclass.RoleOfSeat(seatID), typ, body)
	if !ok {
		return ""
	}
	return "The first stands. If it was wrong, " + a.Clause() + "; the record is append-only and both stay visible"
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
	body, ok := recordpb.BodyField(recordpb.Word(typ))
	if !ok {
		return false, ""
	}
	fd := body.Message().Fields().ByName(field)
	if fd == nil {
		return false, ""
	}
	return true, "--" + flagOf(fd)
}

// RefusesARepeat answers whether the record refuses a second act of this type in one sitting on
// the strength of its KEY: a singleton, or a body that can carry a key field and references what
// the field names rather than defining it — unless the type is numbered, whose every act keys on
// the seat's ordinal. Derived from the three declarations deriveKey reads
// (singleton, keyFields, defines), so a type that joins one of them owes a row in superseders
// without a list to extend — a verb that records such a type and finds no row refuses to mount
// (cli/seat teachAnswer).
func RefusesARepeat(typ recordpb.EventType) bool {
	if singleton[typ] {
		return true
	}
	return !defines[typ] && !numbered[typ] && typeKeyField(typ) != ""
}

// typeKeyField is the first key field a body of this type can carry, or "".
func typeKeyField(typ recordpb.EventType) protoreflect.Name {
	body, ok := recordpb.BodyField(recordpb.Word(typ))
	if !ok {
		return ""
	}
	for _, name := range keyFields {
		if fd := body.Message().Fields().ByName(name); fd != nil && fd.Kind() == protoreflect.StringKind && !fd.IsList() {
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
// AN ENUM ARM OF A ONEOF IS TYPED --as. A ruling's word lands on MotionRule.grade or .petition — one
// arm per subject — and every one of them is set through --as; named by its field,
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
