package migrate

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// A respelling nothing can reach is a respelling that does nothing, and it does nothing silently:
// a key that names no enum, or a target word the vocabulary does not hold, leaves every old record
// refusing exactly as if the entry were absent. And one whose SOURCE word the live vocabulary
// still holds is dead weight — the reverse lookup answers it first. Each entry must be all three
// of: a real enum, an old word the live lookup refuses, a new word it accepts.
func TestValueRenames(t *testing.T) {
	// The deferring disposition's respelling is the entry an archive on the record needs
	// (is-7-prime-run2's R1-1), so it is named rather than inferred from a non-empty table.
	if valueRenames["feov.record.v1.Disposition"]["carried"] != "remanded" {
		t.Fatal("the disposition respelling carried → remanded is gone, and every archive holding the old word refuses")
	}
	for name, words := range valueRenames {
		et, err := protoregistry.GlobalTypes.FindEnumByName(name)
		if err != nil {
			t.Errorf("%s names no enum: %v", name, err)
			continue
		}
		ed := et.Descriptor()
		for from, to := range words {
			if _, err := enumNumberOf(ed, from); err == nil {
				t.Errorf("%s: %q is still a live word, so the respelling to %q never runs", name, from, to)
			}
			want, err := enumNumberOf(ed, to)
			if err != nil {
				t.Errorf("%s: target %q is not in the vocabulary: %v", name, to, err)
				continue
			}
			got, err := enumValue(ed, from)
			if err != nil || got != want {
				t.Errorf("%s: enumValue(%q) = %v, %v; want %v", name, from, got, err, want)
			}
		}
	}
	// A word outside both the vocabulary and the table still refuses: the table translates, it
	// does not widen.
	ed := protoreflect.FullName("feov.record.v1.Disposition")
	et, err := protoregistry.GlobalTypes.FindEnumByName(ed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := enumValue(et.Descriptor(), "postponed"); err == nil {
		t.Error("an unknown disposition word resolved to a value")
	}
}
