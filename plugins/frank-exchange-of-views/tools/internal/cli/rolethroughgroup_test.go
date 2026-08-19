package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// THE ROLE MUST SURVIVE A COMMAND GROUP.
//
// `blue show work list` reported its role as "show" from the day `show` became a group, because
// roleOf took the immediate parent. record.SittingOf switches on that role, a switch that matches
// nothing falls through in silence, and the result was that every seat's duty list lost every
// role-specific line — leaving only the friction duty, which sits above the switch.
//
// This asserts the two halves separately, because they fail for different reasons: the role that
// reaches the projection, and the duties that depend on it. A test for only the first would pass a
// tree where the role arrives and nothing reads it.
func TestTheRoleSurvivesTheShowGroup(t *testing.T) {
	runDir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	board := seatprobe.Boards()["audit"]
	exec := func(args ...string) (string, error) { return run(t, args...) }
	if err := seatprobe.Build(runDir, board, exec); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "show", "work", "--run", runDir, "--seat-id", "red-merge-r1")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Sitting record.SittingJSON `json:"sitting"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("work list did not parse: %v\n%s", err, out)
	}
	if got.Sitting.Role != "merge" {
		t.Errorf("role = %q, want \"merge\" — a projection read under a command group must still know whose sitting it is", got.Sitting.Role)
	}

	// THE DUTIES THAT DEPEND ON IT. This board carries open gaps, so a merge seat owes the
	// disposal and PASS is refused while they stand. Before the fix the only line here was
	// friction, and `complete` went true the moment friction was filed.
	var sawOpenGap bool
	for _, d := range got.Sitting.Open {
		if d.Blocks && strings.Contains(d.What, "is open") {
			sawOpenGap = true
		}
	}
	if !sawOpenGap {
		var what []string
		for _, d := range got.Sitting.Open {
			what = append(what, d.What)
		}
		t.Errorf("a merge sitting on a board with open gaps listed no open-gap duty; outstanding was:\n  %s\n\n"+
			"`verdict --as PASS` refuses over those gaps, so a work list that omits them tells a seat it is finished while the write path holds it.",
			strings.Join(what, "\n  "))
	}
}
