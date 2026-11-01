package blue

import "testing"

// A PART IS PROVED BEFORE ITS FILE GOES: the render after ingest must end with the part, joined as
// reportproj.JoinPart joins it, or the file is kept.
func TestVerifyPart(t *testing.T) {
	part := "## Lane\n\nA sentence.\n"
	for _, c := range []struct {
		name, rendered string
		ok             bool
	}{
		{"the first part", part, true},
		{"a part after another", "## Other\n\nB.\n\n" + part, true},
		{"a byte lost", "## Lane\n\nA sentence\n", false},
		{"the bytes there and the join not", "## Other\n\nB.\n" + part, false},
		{"another part's render", "## Other\n\nB.\n", false},
	} {
		if err := VerifyPart(part, c.rendered); (err == nil) != c.ok {
			t.Errorf("%s: VerifyPart = %v, want ok=%v", c.name, err, c.ok)
		}
	}
}
