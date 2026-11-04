package recordpb

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// THE TABLE IS THE SCHEMA'S, ROW FOR ROW. Every value's annotations are derived afresh off its
// descriptor options and compared with the generated row, so an edit to record.proto that was not
// regenerated fails here as well as at vocabgen -check. The count is asserted because a walk that
// met no row would compare nothing and pass.
func TestTheAnnotationTableIsWhatTheSchemaDeclares(t *testing.T) {
	var walk func(es protoreflect.EnumDescriptors, ms protoreflect.MessageDescriptors)
	compared := 0
	walk = func(es protoreflect.EnumDescriptors, ms protoreflect.MessageDescriptors) {
		for i := 0; i < es.Len(); i++ {
			ed := es.Get(i)
			if got, want := len(generatedAnnotations(ed.FullName())), ed.Values().Len(); got != want {
				t.Errorf("%s: the table holds %d rows and the schema declares %d values", ed.FullName(), got, want)
				continue
			}
			for j := 0; j < ed.Values().Len(); j++ {
				vd := ed.Values().Get(j)
				row, err := annotationsOf(vd)
				if err != nil {
					t.Error(err)
					continue
				}
				if want := AnnotationsFromOptions(vd); *row != want {
					t.Errorf("%s: the table holds %+v and the schema declares %+v — run `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`", vd.FullName(), *row, want)
				}
				compared++
			}
		}
		for i := 0; i < ms.Len(); i++ {
			walk(ms.Get(i).Enums(), ms.Get(i).Messages())
		}
	}
	walk(File_record_proto.Enums(), File_record_proto.Messages())
	if compared == 0 {
		t.Fatal("compared zero rows — the walk found nothing, which must fail rather than pass")
	}
	t.Logf("compared %d rows", compared)
}
