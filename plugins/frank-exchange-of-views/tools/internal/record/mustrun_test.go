package record

import "testing"

// mustRun resolves a directory as a run for record's OWN tests.
//
// Not runtest.Open: these files are `package record`, and runtest imports record. The helper is
// duplicated here rather than shared for that reason alone.
func mustRun(t *testing.T, dir string) Run {
	t.Helper()
	r, err := NewRun(dir)
	if err != nil {
		t.Fatalf("resolving %q as a run: %v", dir, err)
	}
	return r
}
