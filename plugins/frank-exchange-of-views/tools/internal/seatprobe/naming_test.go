package seatprobe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// THE CONSTITUTIONS NAME NO VERB, AND THAT IS NOW THE INVARIANT.
//
// This asserted the opposite: that the redactor REMOVED names from files that carried them, back
// when a constitution's hand-kept list was the shipped configuration. The finding it was built to
// measure landed — the help page is the only page that instructs — so the shipped bytes ARE the
// `none` arm, and the property worth holding is that they stay that way.
//
// The redactor is kept and still asserted below: it is what makes the arm total against a file
// that acquires a name, and a treatment nobody exercises is one nobody would notice breaking.
func TestTheShippedConstitutionsNameNoVerb(t *testing.T) {
	sf := NewSurface(cli.CommandPaths())
	for _, name := range []string{"red-auditor.md", "blue-researcher.md", "blue-synthesizer.md", "lead-judge.md"} {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("..", "..", "..", "agents", name))
			if err != nil {
				t.Fatal(err)
			}
			if named := NamesSurviving(string(b), sf); len(named) > 0 {
				var left []string
				for v, n := range named {
					left = append(left, v+"×"+itoa(n))
				}
				t.Errorf("%s names %d verb(s): %s\n\nA constitution that names a slice of the surface does not under-inform a seat, it SATISFIES it — 58%% surface exposure against 95%% with the names removed (2026-08-15), and re-measured 2026-08-19 over two models the arm that names the WHOLE surface is the worst of the four. Name the ACT; the verb is the help's to state.",
					name, len(named), strings.Join(left, ", "))
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

// THE DUTY-DELIVERY READER IS ITSELF ASSERTED.
//
// It exists because an unmeasured channel co-varied with a treatment and a result was published on
// top of it. A reader built in that repair and never checked would be the same defect once more:
// its no-match and its honest zero are the same number, and "this seat opened no projection" is a
// real outcome the report prints.
func TestViewReadsCountsTheBareFormAsTheWorkList(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "traj.jsonl")
	// A bare `show` resolves to the role default, which is the work list for every role. Counting
	// it as an unnamed view would undercount the ONE carrier of the duty list — the number this
	// whole measurement exists to report.
	lines := []string{
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show --run /r"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show board"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show work"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"grep show board /etc/passwd"}}]}}`,
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := ReadViewReads(p, "feov-record")
	if err != nil {
		t.Fatal(err)
	}
	if v.Work != 2 {
		t.Errorf("work list = %d, want 2 (the bare form plus the explicit one)", v.Work)
	}
	if v.Board != 1 {
		t.Errorf("board = %d, want 1", v.Board)
	}
	// The seat's own grep is not a read of this surface, exactly as Attempted scopes by binary.
	if v.Total != 3 {
		t.Errorf("total = %d, want 3 — a command that never invokes the tool is not a projection read", v.Total)
	}
}
