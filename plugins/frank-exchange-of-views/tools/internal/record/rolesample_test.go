package record

import "testing"

// A ROLE WITH NO SAMPLE BUILDS NO COMMAND TREE, AND A SAMPLE NO DISPATCH PRODUCES WALKS A
// SURFACE THAT DOES NOT EXIST. roleSample is the last hand-written table in the roster; this is
// what keeps it from being the defect it replaced.
func TestEveryRoleHasADispatchableSample(t *testing.T) {
	if err := rolesAndSamplesAgree(); err != nil {
		t.Error(err)
	}
}
