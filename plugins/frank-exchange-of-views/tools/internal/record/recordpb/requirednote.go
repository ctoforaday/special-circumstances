package recordpb

// THE REQUIREDNESS MAP IS GONE, and its deletion is the second half of a change that had landed
// only halfway.
//
// `required.go` held `map[protoreflect.FullName]Requirement` — 27 entries pairing a field with the
// flag a seat types for it and what omitting it costs. The `(sql)` annotation carries all three on
// the field itself, and requiredfields.go reads them: CheckRequired refuses at the write, RequiredOf
// lists them for `--help`, and the DDL emits NOT NULL from the same declaration.
//
// The map survived the annotation by weeks with NO non-test reader. That is the shape
// complete-the-concept describes exactly: the primary edit lands, a carrier keeps speaking the old
// model, and the half-state reads as done — its own tests passed, which is what made it invisible.
// The tests went with it: they asked whether every entry named a live field and whether the map was
// complete, and both questions are answered by the annotation being ON the field.
