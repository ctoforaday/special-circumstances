package recordpb

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// EVERY MOTION ARM IS A SUBJECT, AND EVERY SUBJECT HAS ITS ARM ON BOTH SIDES. Subject reads a
// motion's subject off the arm its oneof sets, by the arm's name; an arm named for no subject would
// read UNSPECIFIED, and a subject with no arm could be neither filed nor ruled.
func TestEveryMotionArmIsASubject(t *testing.T) {
	subjects := MotionSubject(0).Descriptor().Values()
	for _, side := range []struct {
		msg   protoreflect.Message
		oneof protoreflect.Name
		read  func(protoreflect.FieldDescriptor) MotionSubject
	}{
		{(&Motion{}).ProtoReflect(), "filing", func(fd protoreflect.FieldDescriptor) MotionSubject {
			m := &Motion{}
			m.ProtoReflect().Set(fd, m.ProtoReflect().NewField(fd))
			return m.Subject()
		}},
		{(&MotionRule{}).ProtoReflect(), "ruling", func(fd protoreflect.FieldDescriptor) MotionSubject {
			r := &MotionRule{}
			r.ProtoReflect().Set(fd, r.ProtoReflect().NewField(fd))
			return r.Subject()
		}},
	} {
		od := side.msg.Descriptor().Oneofs().ByName(side.oneof)
		if od == nil {
			t.Fatalf("%s has no oneof %q", side.msg.Descriptor().Name(), side.oneof)
		}
		seen := map[MotionSubject]bool{}
		for i := 0; i < od.Fields().Len(); i++ {
			fd := od.Fields().Get(i)
			got := side.read(fd)
			if got == MotionSubject_MOTION_SUBJECT_UNSPECIFIED {
				t.Errorf("%s.%s arm %q reads as no subject", side.msg.Descriptor().Name(), side.oneof, fd.Name())
			}
			seen[got] = true
		}
		for i := 0; i < subjects.Len(); i++ {
			s := MotionSubject(subjects.Get(i).Number())
			if s != MotionSubject_MOTION_SUBJECT_UNSPECIFIED && !seen[s] {
				t.Errorf("subject %s has no %s.%s arm", Word(s), side.msg.Descriptor().Name(), side.oneof)
			}
		}
	}
	if (&Motion{}).Subject() != MotionSubject_MOTION_SUBJECT_UNSPECIFIED || (*MotionRule)(nil).Subject() != MotionSubject_MOTION_SUBJECT_UNSPECIFIED {
		t.Error("a motion or ruling with no arm set reads as a subject")
	}
}
