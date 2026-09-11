package cli

import (
	"testing"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// bodyFor is the body message an event word carries: the Event oneof's field of that name.
func bodyFor(word string) protoreflect.MessageDescriptor {
	fd := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body").Fields().ByName(protoreflect.Name(word))
	if fd == nil {
		return nil
	}
	return fd.Message()
}

// stringFields is every string field of a body, recursing into message fields and oneof arms — so a
// docket ruling's --principle is found at MotionRule.ruling.docket.principle.
func stringFields(md protoreflect.MessageDescriptor) []protoreflect.FieldDescriptor {
	var out []protoreflect.FieldDescriptor
	var walk func(md protoreflect.MessageDescriptor, depth int)
	walk = func(md protoreflect.MessageDescriptor, depth int) {
		fds := md.Fields()
		for i := 0; i < fds.Len(); i++ {
			fd := fds.Get(i)
			switch {
			case fd.Kind() == protoreflect.StringKind:
				out = append(out, fd)
			case fd.Message() != nil && depth < 4:
				walk(fd.Message(), depth+1)
			}
		}
	}
	walk(md, 0)
	return out
}

// EVERY FREE-TEXT FLAG OF A CORRECTABLE VERB FILLS A FIELD THAT SAYS WHETHER A CORRECTION MAY CHANGE
// IT (plans/same-sitting-correction.md III.C.1).
//
// A PROSE-tier correction may change only the fields declaring `(prose) = true`. A free-text field
// that declares nothing would be frozen by omission, so a seat could not fix the very typo the
// mechanism exists for; one that fills a field under ANOTHER name would slip past a walk over
// fields. So this walks FLAGS: each free-text flag must reach at least one field whose flag it is,
// every field it reaches must declare `(prose)`, and a flag reaching several fields must find them
// agreeing. Then the other direction: a `(prose) = true` field no command reaches is stale.
func TestEveryFreeTextFlagOfACorrectableVerbFillsADeclaredField(t *testing.T) {
	byPath := commandsByPath()
	reached := map[protoreflect.FullName]bool{}
	bodies := map[string]protoreflect.MessageDescriptor{}
	checked := 0
	for path, word := range CommandRecords() {
		typ, known := eventTypeWord(word)
		if !known || recordpb.Tier(typ) == recordpb.CorrectionTier_CORRECTION_TIER_NONE {
			continue
		}
		md := bodyFor(word)
		if md == nil {
			t.Errorf("%s records %q, which has no body", path, word)
			continue
		}
		bodies[word] = md
		fields := stringFields(md)
		byPath[path].Flags().VisitAll(func(f *pflag.Flag) {
			if !flags.IsFreeText(f) || len(f.Annotations[seat.CorrectionFlagAnnotation]) > 0 {
				return
			}
			checked++
			var matched []protoreflect.FieldDescriptor
			for _, fd := range fields {
				if recordpb.FlagFor(fd) == f.Name {
					matched = append(matched, fd)
				}
			}
			if len(matched) == 0 {
				t.Errorf("%s --%s fills no field of %s under its own name: annotate the field it fills with (sql).flag = %q", path, f.Name, word, f.Name)
				return
			}
			var agreed *bool
			for _, fd := range matched {
				v, declared := recordpb.IsProse(fd)
				if !declared {
					t.Errorf("%s --%s fills %s, which does not declare (prose) — say whether a correction may change it", path, f.Name, fd.FullName())
					continue
				}
				if agreed != nil && *agreed != v {
					t.Errorf("%s --%s fills fields that disagree about (prose)", path, f.Name)
				}
				agreed = &v
				reached[fd.FullName()] = true
			}
		})
	}
	if checked < 28 {
		t.Fatalf("only %d free-text flags checked — the walk is not seeing the surface", checked)
	}

	// THE CORRECTION'S OWN FREE TEXT is --correction-why, into Correction.why, which is prose.
	why := (&recordpb.Correction{}).ProtoReflect().Descriptor().Fields().ByName("why")
	if recordpb.FlagFor(why) != flags.CorrectionWhy {
		t.Errorf("Correction.why is filled from --%s, want --%s", recordpb.FlagFor(why), flags.CorrectionWhy)
	}
	if v, declared := recordpb.IsProse(why); !v || !declared {
		t.Error("Correction.why does not declare (prose) = true")
	}

	for word, md := range bodies {
		for _, fd := range stringFields(md) {
			if v, declared := recordpb.IsProse(fd); declared && v && !reached[fd.FullName()] {
				t.Errorf("%s declares (prose) = true and no command writing a %s fills it — a stale declaration", fd.FullName(), word)
			}
		}
	}
}
