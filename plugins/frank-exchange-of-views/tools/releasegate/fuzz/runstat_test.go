package fuzz

import (
	"strings"
	"testing"
)

// The RUNSTAT line, which the job log keeps, names a refused act's cause: the envelope exec puts on
// the lines below the exit status is carried onto it.
func TestRunstatNamesTheRefusal(t *testing.T) {
	err := "an act the drive issued: refused, and the drive expects it admitted: mint --seat-id L: exit status 2\n" +
		`  {"verb":"mint","role":"lens","ok":false,"code":"validation","error":"--supersedes names G-1, which no mint event has created — dangling lineage refused"}`
	got := runstatErr(err)
	for _, want := range []string{"exit status 2", "refused (validation)", "dangling lineage refused"} {
		if !strings.Contains(got, want) {
			t.Errorf("RUNSTAT err %q does not say %q", got, want)
		}
	}
	if strings.Contains(got, "\n") {
		t.Errorf("RUNSTAT err spans lines: %q", got)
	}
	if got := runstatErr("forced unsound computation: no plan readied"); got != "forced unsound computation: no plan readied" {
		t.Errorf("an error with no envelope = %q, want its first line", got)
	}
}
