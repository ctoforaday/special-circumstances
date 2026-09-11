package migrate

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

// valueRenames maps an enum WORD an archived record holds to the word the current vocabulary
// spells the same value with, keyed by the enum's full name.
//
// IT LIVES HERE AND NOWHERE ELSE. The live binary accepts and prints only the current word:
// enumNumberOf is shared with nothing a seat reaches, and the write path resolves against the
// schema's own spelling. An archive is the one place the old word still exists, and migrate is
// the one reader of archives, so the translation is migrate's alone.
//
// A same-number respelling is not a SameWord case (case and separators only), which is why it
// is authored: "carried" and "remanded" share no letters, and a bench disposition that defers
// a gap is exactly the value a silent miss would turn into a refusal on every old run that
// holds one.
var valueRenames = map[protoreflect.FullName]map[string]string{
	"feov.record.v1.Disposition": {"carried": "remanded"},
}

// enumValue resolves an archived enum word: the authored respelling first, then the schema's
// own reverse lookup, which refuses a word neither knows.
func enumValue(ed protoreflect.EnumDescriptor, word string) (protoreflect.EnumNumber, error) {
	if to, ok := valueRenames[ed.FullName()][word]; ok {
		word = to
	}
	return enumNumberOf(ed, word)
}
