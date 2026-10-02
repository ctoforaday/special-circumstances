package record

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Usage/Spelling are what the CLI puts in --help. The help IS the contract a seat is told
// to read, so it has to be generated from the set rather than restated beside it: the
// restated version is what was wrong for every one of these flags.
func TestHelpIsGeneratedFromTheSet(t *testing.T) {
	e := MustEnum("verdict", "verdict")
	if got, want := e.Spelling(), "PASS|FAIL"; got != want {
		t.Errorf("Spelling() = %q, want %q", got, want)
	}
	if got, want := e.Usage("the seat's terminal act"), "PASS | FAIL — the seat's terminal act"; got != want {
		t.Errorf("Usage() = %q, want %q", got, want)
	}
}

// MustEnum panics rather than returning a zero value, because a zero value would render as
// an EMPTY help string — a flag that silently stops advertising its set is the defect this
// whole table exists to remove, arriving by the back door.
func TestMustEnumPanicsRatherThanRenderingAnEmptySet(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustEnum returned quietly for an undeclared key")
		}
	}()
	_ = MustEnum("verdict", "no-such-key")
}

// THE DUPLICATION THIS TEST GUARDED IS GONE, AND THAT IS WHAT IT NOW ASSERTS.
//
// It used to compare `petition.class` against `petition-rule.class` — two declared sets for one
// vocabulary, kept in step by a test because nothing structural kept them in step. A class a seat
// could file but the bench could not rule on was a petition that could never be answered.
//
// After #344 there is ONE table. The filing and the ruling read `MotionFields["petition"]["class"]`
// and `MotionVerdicts["petition"]`, each declared once, so the drift is not detected — it is
// unrepresentable. The test survives inverted: it fails if a second declaration of either
// vocabulary ever reappears in EnumFields, which is how the duplication would come back.
func TestTheAdjudicationVocabulariesHaveExactlyOneSourceEach(t *testing.T) {
	for _, typ := range []string{"petition", "petition-rule", "dispute", "dispute-respond", "avenue-rule"} {
		if _, ok := EnumFields[typ]; ok {
			t.Errorf("EnumFields still declares sets for %q — that event type is retired (#344) and its vocabulary lives in record/motion.go. A second declaration is how the drift this test used to police comes back", typ)
		}
	}
	if len(MotionFields["petition"]["class"]) == 0 {
		t.Error("MotionFields lost the petition classes — the one source the filing and the ruling both read")
	}
	// AND THE CENSUS ABOVE WAS INCOMPLETE, WHICH IS WHY THIS BLOCK EXISTS.
	//
	// This test asserted "there is ONE table … the drift is not detected, it is unrepresentable"
	// while counting only the two tables it knew about. The WRITE PATH resolves against the proto
	// ENUM, a third source it never looked at — and the three disagreed. MotionFields listed
	// `ethical | safety | integrity | constitutional`; PetitionClass carried
	// `integrity | safety | process | scope`. Half the advertised classes were refused at the
	// write for a value the seat had just read in --help. `binds` overlapped in NOTHING, and
	// --binds is set exactly when a petition is granted, so no granted petition could be recorded.
	//
	// A test that names its own completeness is only as good as its census. This one now checks
	// the source it missed: every word the help offers must be a word the write path resolves.
	for _, tc := range []struct {
		subject, key string
		ed           protoreflect.EnumDescriptor
	}{
		{"petition", "class", recordpb.PetitionClass(0).Descriptor()},
		{"petition", "binds", recordpb.RulingBinds(0).Descriptor()},
		{"grade", "dimension", recordpb.GradeDimension(0).Descriptor()},
	} {
		for _, v := range MotionFields[tc.subject][tc.key] {
			if _, ok := recordpb.BySpelling(tc.ed, v.Name); !ok {
				t.Errorf("%s --%s advertises %q and the write path cannot resolve it against %s — "+
					"a seat that reads the help and types the word is REFUSED, which teaches that the help lies rather than what to pass",
					tc.subject, tc.key, v.Name, tc.ed.FullName())
			}
		}
		// And the other direction: a schema value the help never offers is a word nothing can
		// choose, which is how `process` and `scope` sat in the enum unreachable for releases.
		vals := tc.ed.Values()
		for i := 0; i < vals.Len(); i++ {
			if vals.Get(i).Number() == 0 {
				continue
			}
			w := recordpb.Spelling(vals.Get(i))
			if !slices.Contains(Names(MotionFields[tc.subject][tc.key]), w) {
				t.Errorf("%s carries %q and %s --%s never offers it — a value no surface can produce is dead vocabulary, and it reads as a richer set than the one that works",
					tc.ed.FullName(), w, tc.subject, tc.key)
			}
		}
	}
	if len(MotionVerdicts["petition"]) == 0 {
		t.Error("MotionVerdicts lost the petition rulings")
	}
	// The four grade axes, likewise: they were declared on `dispute` and read by
	// `dispute-respond`'s join. One table now.
	if got := strings.Join(Names(MotionFields["grade"]["dimension"]), "|"); got != "severity|likelihood|impact|complexity" {
		t.Errorf("the grade dimensions moved or changed: %q — the ruling is matched to the filing on (gap, dimension), so a change here silently unpairs asks from answers", got)
	}
}

// sameWord is the typo detector behind the closure-class near-miss guard. It must catch
// case and separator differences and NOTHING wider — a wider match would refuse a closure
// class somebody meant, in an enum that is deliberately open.
func TestSameWordCatchesTyposAndNothingWider(t *testing.T) {
	same := []string{"repaired-with-regression", "Repaired_With_Regression", "REPAIREDWITHREGRESSION", "repaired with regression"}
	for _, s := range same {
		if !sameWord(s, "repaired_with_regression") {
			t.Errorf("sameWord(%q) missed a typo", s)
		}
	}
	for _, s := range []string{"repaired", "repaired_with_regressions", "evidence-rebutted", "", "regression"} {
		if sameWord(s, "repaired_with_regression") {
			t.Errorf("sameWord(%q) matched something that is a different word", s)
		}
	}
}

// THE TWO CLOSURE SETS ARE CLOSED NOW (#342), and this test is the inverse of the one it
// replaces. That test asserted `opinion` and `close` must have NO closed set, on the reasoning
// that "closing it means a legitimate act failing hard mid-round" — sound while the candidate
// words were inconsistent, which enums.go recorded as the blocker.
//
// The inconsistency was the thing to fix, and it was worse than the note said: FOUR
// vocabularies for one concept. Now there is one, so an unrecognized class is a typo rather
// than a legitimate act the tool has not heard of.
func TestBothClosureSetsShareOneVocabulary(t *testing.T) {
	closeSet := EnumFields["close"]
	if len(closeSet) != 1 {
		t.Fatal("`merge close` must declare exactly one closed set")
	}
	// THE BENCH'S SET IS KEYED ON (SUBJECT, RULING) NOW, so it is read from the motion table
	// rather than from EnumFields — `bench opinion` was its own event type with its own entry,
	// and the disposition is a docket motion's ruling. Read through the same accessor the CLI's
	// help and the write check use, so a set this test agrees with is the set a seat is offered.
	benchSet := MotionVerdictEnum("docket").Values
	if len(benchSet) == 0 {
		t.Fatal("the bench's docket vocabulary is empty — an empty set would pass every check below without comparing anything")
	}
	closes := map[string]bool{}
	for _, v := range Names(closeSet[0].Values) {
		closes[v] = true
	}
	// Every class red may close with must also be a disposition the bench may rule, or the
	// two verbs mean different things by the same outcome — which is what #342 removed.
	for _, v := range Names(benchSet) {
		if v == DispositionRemanded {
			continue
		}
		if !closes[v] {
			t.Errorf("the bench may rule %q but red cannot close with it — one outcome, two vocabularies again", v)
		}
	}
	for _, v := range closeSet[0].Values {
		found := false
		for _, d := range benchSet {
			if d == v {
				found = true
			}
		}
		if !found {
			t.Errorf("red may close with %q but the bench cannot rule it — one outcome, two vocabularies again", v.Name)
		}
	}
	// `carried` is the ONE word that defers instead of closing, and only the bench has it.
	if closes[DispositionRemanded] {
		t.Error("`carried` is not a closure — red must not be able to close a gap with it")
	}
}

// EVERY WORD A TABLE DECLARES MUST BE A WORD THE SCHEMA CARRIES, WITH THE SCHEMA'S MEANING,
// AND EVERY WORD THE SCHEMA CARRIES MUST BE ON THE TABLE.
//
// EnumFields is what `--help` renders and what the inlined manual of every agent definition
// carries, so a value in it that the schema does not have is a value a seat is TOLD to use and
// the record then refuses. That is #342 in one sentence: the engine declared `unresolved`, `moot`
// and `grade_adjusted`, the constitution instructed all three, and the write path rejected every
// one — so a bench following its own constitution could record nothing.
//
// The first pass at this test checked one direction — table ⊆ schema — and only the WORDS. Both
// halves it left unchecked drifted (#1208). The meanings: `LOG_TYPE_FRICTION`'s `(means)` was
// rewritten in PR #1197 and the table kept the old sentence, so the rewrite that PR was for never
// reached a seat; `RUN_OUTCOME_CEILING` lost "at impasse … with nobody ready and PASS not
// permitted" the same way. The coverage: `log.type` listed four of LogType's six values, and a
// one-way subset check cannot see an omission. So the check now runs both ways and compares the
// prose, resolving each word through the schema's own spelling table rather than a string compare
// against a second list, because a second list is the thing that drifted.
// EVERY TABLE IS A VIEW OF ITS DESCRIPTOR, and this is the gate that refuses the next hand-typed
// row. It runs over every set rather than a list of the hand-kept ones: an allowlist of "tables
// that cannot be generated" is itself a hand-kept copy, and today it would be empty — the two
// capitalised sets are the descriptor's words through `loud`, and `close` is narrowed by the
// `closes` facet in dispositionsWhere. The checks that still bite are structural: the field the
// table names exists in the schema and is an enum, no set is empty, every value carries a
// meaning, and a narrowing is declared by a facet (narrowedBy) rather than typed.
func TestEveryDeclaredValueIsAWordTheSchemaCarries(t *testing.T) {
	checked := 0
	for typ, fields := range EnumFields {
		for _, f := range fields {
			ef := f
			t.Run(typ+"."+ef.Key, func(t *testing.T) {
				if len(ef.Values) == 0 {
					t.Errorf("%s.%s declares an empty set — it would render as no choices at all", typ, ef.Key)
				}
				ed := schemaEnumFor(t, typ, ef.Key)
				for _, v := range ef.Values {
					checked++
					if v.Means == "" {
						t.Errorf("%s.%s value %q carries no meaning; a set rendered as bare words leaves a "+
							"seat to guess which situation warrants which", typ, ef.Key, v.Name)
					}
					if ed == nil {
						continue // the one open set: its vocabulary is enforced by checkOpenSets, not by a type
					}
					vd, ok := schemaValue(ed, v.Name)
					if !ok {
						t.Errorf("%s.%s declares %q, which the schema does not carry — a seat reading --help "+
							"is told to use a word the write path refuses", typ, ef.Key, v.Name)
						continue
					}
					// THE MEANING IS THE SCHEMA'S, VERBATIM. The table is a view of the descriptor, and a
					// view that paraphrases is a second author: the `(means)` is where the sentence is
					// edited, and this is what carries the edit to the help a seat reads.
					want, err := recordpb.EnumValueDoc(vd)
					if err != nil {
						t.Errorf("%s.%s: %v", typ, ef.Key, err)
						continue
					}
					if v.Means != want {
						t.Errorf("%s.%s value %q restates the schema's meaning in other words — the help a seat "+
							"reads no longer says what the schema says\n  table:  %q\n  schema: %q",
							typ, ef.Key, v.Name, v.Means, want)
					}
				}
				if ed == nil {
					return
				}
				// EVERY VALUE THE SCHEMA CARRIES IS ON THE TABLE. A tool-only word is listed too — it
				// carries ToolOnly off its facet and SeatFilable drops it from a seat's surface — so the
				// line the tool renders about its own acts is generated from the same row. The only
				// principled narrowing is declared by a FACET, once, in narrowedBy.
				narrow := narrowedBy[[2]string{typ, ef.Key}]
				for i := 0; i < ed.Values().Len(); i++ {
					vd := ed.Values().Get(i)
					if vd.Number() == 0 {
						continue // UNSPECIFIED is the absence of a choice, never offered
					}
					if narrow != nil && narrow(t, vd) {
						continue
					}
					if !tableCarries(ef.Values, vd) {
						t.Errorf("%s.%s omits %s, which the schema carries — a value the table does not list "+
							"reaches no help page, so a seat cannot learn what the word means", typ, ef.Key, vd.Name())
					}
				}
			})
		}
	}
	if checked == 0 {
		t.Fatal("no declared values were checked — an empty traversal passes this test on every set")
	}
}

// THE RULER'S MENU IS THE SCHEMA'S SENTENCE, for the motion vocabularies as for EnumFields.
// MotionVerdicts is keyed on the subject, which EnumFields cannot express, so the check above does
// not reach it — and `GRADE_RULING_ACCEPTED` said "the proposed grade stands" while the chair's
// menu said the lens moves the grade with `regrade`, for as long as nothing compared them. Each
// subject's enum is resolved through MotionRule's `ruling` oneof, so a new arm fails here rather
// than slipping in with a set nothing checks.
func TestEveryMotionVerdictCarriesTheSchemasMeaning(t *testing.T) {
	od := (&recordpb.MotionRule{}).ProtoReflect().Descriptor().Oneofs().ByName("ruling")
	if od == nil {
		t.Fatal("MotionRule carries no `ruling` oneof")
	}
	checked := 0
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		subject := string(fd.Name())
		if fd.Kind() != protoreflect.EnumKind {
			// `docket` is a message, and its disposition is the shared Disposition set, which
			// TestBothClosureSetsShareOneVocabulary already holds to the descriptor.
			continue
		}
		vs, ok := MotionVerdicts[subject]
		if !ok {
			t.Errorf("MotionRule rules on %q and MotionVerdicts carries no set for it", subject)
			continue
		}
		ed := fd.Enum()
		for _, v := range vs {
			checked++
			vd, found := recordpb.BySpelling(ed, v.Name)
			if !found {
				t.Errorf("MotionVerdicts[%q] offers %q, which %s does not carry", subject, v.Name, ed.FullName())
				continue
			}
			want, err := recordpb.EnumValueDoc(vd)
			if err != nil {
				t.Errorf("MotionVerdicts[%q]: %v", subject, err)
				continue
			}
			if v.Means != want {
				t.Errorf("MotionVerdicts[%q] value %q restates the schema's meaning in other words — the "+
					"ruler's menu no longer says what the schema says\n  table:  %q\n  schema: %q",
					subject, v.Name, v.Means, want)
			}
		}
		for j := 0; j < ed.Values().Len(); j++ {
			vd := ed.Values().Get(j)
			if vd.Number() != 0 && !tableCarries(vs, vd) {
				t.Errorf("MotionVerdicts[%q] omits %s, which the schema carries", subject, vd.Name())
			}
		}
	}
	if checked == 0 {
		t.Fatal("no motion verdict was checked — an empty traversal passes this test on every set")
	}
}

// narrowedBy names the ONE set that lists fewer words than its enum carries, and the facet that
// says why: red closes a gap with a Disposition that `closes`, and `remanded` is declared not to.
// A narrowing that no facet declares is a hand-restated set, and the coverage check above refuses it.
var narrowedBy = map[[2]string]func(t *testing.T, vd protoreflect.EnumValueDescriptor) bool{
	{"close", "closure_class"}: func(t *testing.T, vd protoreflect.EnumValueDescriptor) bool {
		closes, declared, err := recordpb.Facet(vd, "closes")
		if err != nil || !declared {
			t.Fatalf("Disposition %s declares no `closes` facet: %v", vd.Name(), err)
		}
		return !closes
	},
}

// tableCarries answers whether the table lists this schema value, under the schema's spelling or
// the LOUDER one two converters accept (see schemaValue).
func tableCarries(vs []EnumValue, vd protoreflect.EnumValueDescriptor) bool {
	word := recordpb.Spelling(vd)
	for _, v := range vs {
		if v.Name == word || strings.ToLower(v.Name) == word {
			return true
		}
	}
	return false
}

// schemaEnumFor resolves (event type, payload key) to the enum the record holds there, by asking
// the DESCRIPTOR rather than a list beside it. nil is the one open set, whose field is a string.
func schemaEnumFor(t *testing.T, typ, key string) protoreflect.EnumDescriptor {
	t.Helper()
	md, ok := bodyDescriptorFor(typ)
	if !ok {
		t.Fatalf("EnumFields names the event type %q, which the schema does not declare", typ)
	}
	fd := md.Fields().ByName(protoreflect.Name(key))
	if fd == nil {
		t.Fatalf("%s has no field %q — the table names a field the schema does not carry, so its "+
			"whole set is advertised against nothing", typ, key)
	}
	if fd.Kind() != protoreflect.EnumKind {
		return nil
	}
	return fd.Enum()
}

// schemaValue resolves a table word to the schema value it names.
//
// A CONVERTER MAY FOLD CASE, and two deliberately do. `chair verdict --as PASS` and `bench
// outcome --as VERIFIED` are the seat's words in capitals — that is the surface, and VerdictOf
// and RunOutcomeOf lowercase before resolving. BySpelling is exact by design (its own test
// pins that `PASS` does not resolve to `pass`), so the fold is checked here rather than
// weakened there. It is one-way: a declared word may be louder than the schema's, never
// different from it.
func schemaValue(ed protoreflect.EnumDescriptor, word string) (protoreflect.EnumValueDescriptor, bool) {
	if vd, found := recordpb.BySpelling(ed, word); found {
		return vd, true
	}
	return recordpb.BySpelling(ed, strings.ToLower(word))
}

// bodyDescriptorFor resolves an event type's WORD to the body message the schema pairs with it,
// through the `body` oneof — the one place that pairing is declared.
func bodyDescriptorFor(typ string) (protoreflect.MessageDescriptor, bool) {
	od := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		return nil, false
	}
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if string(fd.Name()) == typ && fd.Message() != nil {
			return fd.Message(), true
		}
	}
	return nil, false
}

// THE DOCKET AXIS AND THE ARTIFACT AXIS ARE DIFFERENT, and for a long time one word carried both.
//
// Measured 2026-08-22 on the sqlite-schema run: the board said open:0 while assembly-screen
// reported a source red had found AGAINST still cited in the shipped report. Two gates, one run,
// contradicting each other in English — because "closed" answers whether the DISPUTE ended, and
// three of the six classes end the dispute while leaving a live defect in the artifact.
func TestArtifactStateSeparatesTheDisputeFromTheDefect(t *testing.T) {
	for _, c := range []struct {
		class string
		want  ArtifactState
	}{
		{"repaired", ArtifactRepaired},
		{"not_a_defect", ArtifactNoDefect},
		{"defect_accepted", ArtifactDefectLive},
		{"defect_owed_elsewhere", ArtifactDefectLive},
		{"repaired_with_regression", ArtifactDefectLive},
		{DispositionRemanded, ArtifactUnexamined},
	} {
		got, ok := ArtifactStateOf(c.class)
		if !ok {
			t.Errorf("%s: ArtifactStateOf reported it cannot answer, but only amends_prior may", c.class)
			continue
		}
		if got != c.want {
			t.Errorf("%s -> %s, want %s", c.class, got, c.want)
		}
	}

	// THE THREE THAT SHIP A LIVE DEFECT must not be readable as repaired. This is the assertion
	// the board could not make before, and the one the run needed.
	for _, class := range []string{"defect_accepted", "defect_owed_elsewhere", "repaired_with_regression"} {
		if got, _ := ArtifactStateOf(class); got == ArtifactRepaired || got == ArtifactNoDefect {
			t.Errorf("%s reads as %s — a settled dispute is being reported as a sound artifact, "+
				"which is how a board says open:0 while the report ships a defect red found against", class, got)
		}
	}

	// AMENDS_PRIOR REFUSES TO ANSWER rather than guessing. It is defined relative to an earlier
	// ruling, so a table row for it would be a fabricated value wearing a derived one's authority.
	if _, ok := ArtifactStateOf("amends_prior"); ok {
		t.Error("amends_prior answered from the class alone — it inherits from the ruling it amends, " +
			"and a plausible answer here is worse than an honest refusal")
	}

	// A WORD OUTSIDE THE VOCABULARY REPORTS ITSELF. No record can carry one — the schema's CHECK
	// is generated from the same enum artifactByClass is checked against below — so this pins the
	// behaviour for input that reached the function without passing the schema.
	//
	// THE EXAMPLE USED TO BE `moot`, which the engine offered and the record refused. It is a
	// record word now (#847), so the example had to become one that is genuinely outside — and
	// this is the reason the example must be invented rather than borrowed from a real surface:
	// borrowed, it stops testing what it claims the moment the surfaces agree.
	if got, ok := ArtifactStateOf("adjudicated_sideways"); !ok || got != ArtifactUnknown {
		t.Errorf("a word the schema does not carry read as %q/%v, want unknown — it must not "+
			"fold into a healthy value", got, ok)
	}

	// EVERY class in the live vocabulary is covered, so adding one without deciding its artifact
	// meaning fails here rather than defaulting to unknown in a projection nobody re-reads.
	for _, name := range closureClassNames() {
		if name == "amends_prior" {
			continue
		}
		if got, _ := ArtifactStateOf(name); got == ArtifactUnknown {
			t.Errorf("closure class %q has no artifact meaning — decide it here, not at the read", name)
		}
	}
}

// A TOOL-WRITTEN LOG TYPE IS NOT ON THE SEAT'S SURFACE, AND CANNOT BE FILED UNDER A SEAT'S NAME.
//
// `estoppel` MEANS "the tool refused a mint against text the other side prescribed" — its own
// `means` says "Recorded by the tool, not filed by the seat". The seat verb accepted it anyway
// and stamped source=SEAT, so a seat could record a refusal that never happened; and the generated
// help OFFERED the word beside the four a seat should use, which is what made it likely rather
// than hypothetical (#782). A reader keying on `type` alone — which is what the operator triage
// channel is documented to do — could not tell the two apart.
//
// THREE ARMS, because the fix has three surfaces and any one of them alone leaves the hole open:
// the schema must carry the fact, the seat's surface must not offer it, and the write path must
// refuse it under a seat's name while still admitting the tool's own.
//
// THE TOOL-ONLY WORDS ARE READ OFF THE FACET, not listed here: a hand map beside the facet is a
// second copy of the fact, and the first version of this test checked `estoppel` alone while the
// surface had grown two more tool-written words.
func TestAToolWrittenLogTypeIsOffTheSeatSurface(t *testing.T) {
	// 1. THE SCHEMA CARRIES IT, and every value answers. A facet declared on some values and not
	// others is refused at schema build, so this also pins that nobody added a word without an
	// answer — the exact way `grade_adjusted` once acquired a closing meaning nobody chose.
	var seatWords, toolWords []string
	vals := recordpb.LogType(0).Descriptor().Values()
	for i := 0; i < vals.Len(); i++ {
		lt := recordpb.LogType(vals.Get(i).Number())
		if lt == recordpb.LogType_LOG_TYPE_UNSPECIFIED {
			continue
		}
		may, declared, err := recordpb.SeatMayFile(lt)
		if err != nil {
			t.Fatal(err)
		}
		if !declared {
			t.Errorf("log type %q does not say whether a seat may file it — defaulting that would "+
				"answer on behalf of whoever added the word", recordpb.Word(lt))
		}
		if may {
			seatWords = append(seatWords, recordpb.Word(lt))
		} else {
			toolWords = append(toolWords, recordpb.Word(lt))
		}
	}
	// Anti-vacuity: both halves exist, or the loops below check nothing. The tool writes its own
	// entries (an estoppel, every other refusal it gives, a seat's failed call the harness saw)
	// and a seat files the rest.
	if len(toolWords) == 0 || len(seatWords) == 0 {
		t.Fatalf("LogType must carry both tool-only and seat-filable words: tool=%v seat=%v", toolWords, seatWords)
	}

	// 2. THE SEAT'S SURFACE OMITS EVERY TOOL-ONLY WORD and carries every seat-filable one, in
	// schema order — the one derivation (SeatLogTypeEnum, through the ToolOnly facet init stamps)
	// agrees with the facet read directly.
	if help := Names(SeatLogTypeEnum().Values); !slices.Equal(help, seatWords) {
		t.Errorf("the seat's log surface disagrees with the seat_may_file facet: help=%v facet=%v", help, seatWords)
	}
	// 3. The FULL table still carries every tool-only word — this narrows the seat's surface, not
	// the record's. The vocabulary table and every reader still know the words.
	for _, w := range toolWords {
		if !slices.Contains(Names(MustEnum("log", "type").Values), w) {
			t.Errorf("narrowing the seat surface removed %q from the record's vocabulary — the "+
				"tool still writes it and every reader still has to resolve it", w)
		}
	}
}

// AND THE WRITE PATH HOLDS THE INVARIANT FOR EVERY WRITER, which is where it belongs: the CLI's
// own enum set already refuses the word at flag parse, so a check there could never fire.
func TestAToolOnlyLogTypeIsRefusedUnderASeatsName(t *testing.T) {
	run := mustRun(t, newRun(t))
	id := Identity{Run: run, SeatID: "red-chair"}
	est := recordpb.LogType_LOG_TYPE_ESTOPPEL
	seat, tool := recordpb.LogSource_LOG_SOURCE_SEAT, recordpb.LogSource_LOG_SOURCE_TOOL

	if _, err := Append(id, &recordpb.Log{
		Text: proto.String("a refusal that never happened"), Type: &est, Source: &seat,
	}); err == nil {
		t.Error("a seat filed an `estoppel` entry — it can now claim the tool refused a mint it never refused")
	}
	// THE TOOL'S OWN WRITE STILL PASSES. Without this the fix could be a blanket ban on the word,
	// which would delete the estoppel guard's only record rather than protect it.
	if _, err := Append(id, &recordpb.Log{
		Text: proto.String("merge mint: estoppel — this quotes text you prescribed"),
		Type: &est, Source: &tool, EstoppedBy: proto.String("G1"),
	}); err != nil {
		t.Errorf("the TOOL's own estoppel record was refused: %v", err)
	}
	// A seat-filable word under a seat's name is untouched.
	nom := recordpb.LogType_LOG_TYPE_FRICTION
	if _, err := Append(id, &recordpb.Log{
		Text: proto.String("the board view paged awkwardly for this gap"), Type: &nom, Source: &seat,
	}); err != nil {
		t.Errorf("an ordinary seat log was refused: %v", err)
	}
}
