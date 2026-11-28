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
	// AN ACT RUN WITHOUT --json prints its refusal as text, and RUNSTAT carries its first line.
	text := "an act the drive issued: refused, and the drive expects it admitted: motion docket rule --for-raiser: exit status 2\n" +
		"  the quoted span was not found in report.md — quote the EXACT current text\nmore"
	if got := runstatErr(text); !strings.Contains(got, "exit status 2 — the quoted span was not found") || strings.Contains(got, "more") {
		t.Errorf("RUNSTAT err of a text refusal = %q, want the exit status and the refusal's first line", got)
	}
	if got := runstatErr("forced unsound computation: no plan readied"); got != "forced unsound computation: no plan readied" {
		t.Errorf("an error with no envelope = %q, want its first line", got)
	}
}
