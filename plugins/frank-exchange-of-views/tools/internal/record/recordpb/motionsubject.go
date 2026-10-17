package recordpb

import "google.golang.org/protobuf/reflect/protoreflect"

// Subject is the motion's subject: the arm its filing sets. The filing IS the subject — a grade
// motion is one whose filing is a GradeMotion — so the record stores it once, in the arm, and every
// reader asks here. UNSPECIFIED is a motion whose filing is unset.
func (x *Motion) Subject() MotionSubject { return armSubject(x, "filing") }

// Subject is the ruling's subject: the arm its ruling sets, read as Motion.Subject reads a filing.
func (x *MotionRule) Subject() MotionSubject { return armSubject(x, "ruling") }

// armSubject is the one reading of a motion arm as a subject: the MotionSubject spelled as the arm
// is named. The arms of `filing` and `ruling` are named for the subjects they carry
// (TestEveryMotionArmIsASubject), so no table here can fall out of step with the schema.
func armSubject(m interface{ ProtoReflect() protoreflect.Message }, oneof protoreflect.Name) MotionSubject {
	if m == nil {
		return MotionSubject_MOTION_SUBJECT_UNSPECIFIED
	}
	msg := m.ProtoReflect()
	if !msg.IsValid() {
		return MotionSubject_MOTION_SUBJECT_UNSPECIFIED
	}
	fd := msg.WhichOneof(msg.Descriptor().Oneofs().ByName(oneof))
	if fd == nil {
		return MotionSubject_MOTION_SUBJECT_UNSPECIFIED
	}
	vd, ok := BySpelling(MotionSubject(0).Descriptor(), string(fd.Name()))
	if !ok {
		return MotionSubject_MOTION_SUBJECT_UNSPECIFIED
	}
	return MotionSubject(vd.Number())
}
