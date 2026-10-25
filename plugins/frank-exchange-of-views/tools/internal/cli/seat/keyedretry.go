package seat

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Asserted is one thing a keyed act states: the flag a seat states it with and the record field it
// lands in. A retry under the key states it again, and RepeatsKeyed holds the two to each other.
type Asserted struct {
	Flag  string
	Field protoreflect.Name
}

// RepeatsKeyed refuses a keyed retry that states other than the act stored under its key, as a
// `conflict` naming the stored values; nil where every asserted field repeats the stored one. id
// is "" for an act the tool names by no id.
//
// A KEY IS ONE ACT'S NAME, as it is one program's for `prove` and one citation's for `cite`: a
// retry that states a different grade, subject or text is a second act reusing the name, and
// answering it with the first leaves the seat holding something other than what it asked for.
// Every fact the act's verb records from a flag is listed by its caller, the location included: a
// retry that quotes other text is another act. An absent field and an empty one are the same
// statement.
func RepeatsKeyed(key, noun, id string, stored, retry proto.Message, asserted ...Asserted) error {
	s, r := stored.ProtoReflect(), retry.ProtoReflect()
	var was, now []string
	for _, a := range asserted {
		fd := s.Descriptor().Fields().ByName(a.Field)
		if fd == nil {
			panic(fmt.Sprintf("%s has no field %s", s.Descriptor().FullName(), a.Field))
		}
		sv, rv := s.Get(fd), r.Get(fd)
		if sv.Equal(rv) {
			continue
		}
		was = append(was, statement(a.Flag, fd, sv))
		now = append(now, statement(a.Flag, fd, rv))
	}
	if was == nil {
		return nil
	}
	// An act the tool gives no id (an edit) is named by its noun alone.
	named := noun + " " + id
	if id == "" {
		named = "an " + noun
	}
	return feov.Errorf(feov.Conflict, "--key %q already names %s on this run, stated with %s; this repeat states %s — "+
		"a key is one %s's name, not a slot to reuse. This is not a retry: recording nothing here would leave you "+
		"with another %s than the one you stated. Give this %s its own --key",
		key, named, strings.Join(was, " "), strings.Join(now, " "), noun, noun, noun)
}

// statement is a field's value as the flag that states it: `--severity high`, `--reason "…"`, or
// `no --about` where the field holds nothing.
func statement(flag string, fd protoreflect.FieldDescriptor, v protoreflect.Value) string {
	if fd.IsList() {
		l := v.List()
		if l.Len() == 0 {
			return "no --" + flag
		}
		parts := make([]string, l.Len())
		for i := range parts {
			parts[i] = l.Get(i).String()
		}
		return fmt.Sprintf("--%s %s", flag, strings.Join(parts, ","))
	}
	switch fd.Kind() {
	case protoreflect.EnumKind:
		if v.Enum() == 0 {
			return "no --" + flag
		}
		if vd := fd.Enum().Values().ByNumber(v.Enum()); vd != nil {
			return fmt.Sprintf("--%s %s", flag, recordpb.Spelling(vd))
		}
	case protoreflect.BoolKind:
		if !v.Bool() {
			return "no --" + flag
		}
		return "--" + flag
	case protoreflect.StringKind:
		if v.String() == "" {
			return "no --" + flag
		}
		return fmt.Sprintf("--%s %q", flag, v.String())
	}
	return fmt.Sprintf("--%s %v", flag, v.Interface())
}
