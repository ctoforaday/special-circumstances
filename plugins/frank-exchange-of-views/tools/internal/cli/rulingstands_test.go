package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RULING STANDS, AND NO SEAT APPEALS ONE.
//
// A seat that disagrees with a ruling says so in an act its surface holds: a docket motion on the
// gap, which the bench rules; a new grade motion while the gap is open; a move of the avenue. This
// file holds the surface, the refusal and the write path to that, for every role's tree rather than
// for the subjects that exist today — so a subject added later cannot grow the verb back.

// roleCommands visits every command of every role's tree, groups included.
func roleCommands(visit func(role string, path []string, c *cobra.Command)) {
	roots := AllRoots()
	roles := make([]string, 0, len(roots))
	for r := range roots {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	var walk func(role string, path []string, c *cobra.Command)
	walk = func(role string, path []string, c *cobra.Command) {
		visit(role, path, c)
		for _, sub := range c.Commands() {
			walk(role, append(append([]string{}, path...), sub.Name()), sub)
		}
	}
	for _, role := range roles {
		walk(role, nil, roots[role])
	}
}

// No command of any role's tree is an appeal, records one, or is a group with nothing under it.
func TestNoSurfaceHoldsAnAppeal(t *testing.T) {
	var seen, motionGroups int
	roleCommands(func(role string, path []string, c *cobra.Command) {
		seen++
		at := role + " " + strings.Join(path, " ")
		if c.Name() == "appeal" || c.HasAlias("appeal") {
			t.Errorf("%s is an appeal: a ruling stands, and the seat that disagrees files a docket motion", at)
		}
		if seat.RecordType(c) == recordpb.Word(recordpb.EventType_EVENT_TYPE_MOTION_APPEAL) {
			t.Errorf("%s records a %s: no command writes one", at, seat.RecordType(c))
		}
		// A SUBJECT WITH NO VERB FOR THIS ROLE IS NOT ON ITS SURFACE. An avenue motion has a ruling
		// and nothing else, so only its ruler's tree holds the group; mounted bare on the others it
		// is a page that offers no act.
		if len(path) == 2 && path[0] == "motion" {
			motionGroups++
			if !c.HasSubCommands() {
				t.Errorf("%s is a motion subject with no verb on this surface", at)
			}
		}
	})
	if seen < 100 || motionGroups < 12 {
		t.Fatalf("the walk visited %d commands and %d motion subjects; it has stopped seeing the surface", seen, motionGroups)
	}
}

// A seat that types an appeal is refused, nothing is written, and the refusal shows what the seat
// holds in its place.
func TestAnAppealIsRefusedWithWhatTheSeatHolds(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, seatID := range []string{"blue-respond", lensSeat, "red-chair", "judge"} {
		t.Run(seatID, func(t *testing.T) {
			runDir := actFixture(t)
			m := fileGradeMotion(t, runDir)
			must(t, runDir, "motion", "grade", "rule", "--seat-id", "red-chair", "--id", m, "--as", "rejected", "--reason", "the evidence does not reach it")
			proposeQ1(t, runDir)
			must(t, runDir, "motion", "avenue", "rule", "--seat-id", "red-chair", "--id", "Q1", "--as", "out_of_scope", "--reason", "not this question")
			d := drive{t: t, run: runDir}

			// The grade subject is on every surface, so the refusal is the subject's own: it names
			// the verbs the seat holds there.
			grade := d.refused("an appeal of the grade ruling", seatID, "motion", "grade", "appeal", "--id", m, "--reason", "the ruling ignored the source")
			if !strings.Contains(grade, "has no `appeal` verb — it has file") {
				t.Errorf("the refusal of a grade appeal does not say what a grade motion holds for %s:\n%s", seatID, grade)
			}
			// The avenue subject is on the ruler's surface alone.
			avenue := d.refused("an appeal of the avenue ruling", seatID, "motion", "avenue", "appeal", "--id", "Q1", "--reason", "it is this question")
			want := "the chair's `motion avenue`"
			if seatID == "red-chair" {
				want = "an avenue motion has no `appeal` verb — it has rule"
			}
			if !strings.Contains(avenue, want) {
				t.Errorf("the refusal of an avenue appeal by %s does not say %q:\n%s", seatID, want, avenue)
			}
			if n := len(eventsOfWord(t, runDir, "motion_appeal")); n != 0 {
				t.Fatalf("the record holds %d appeal(s) after refused invocations", n)
			}

			// WHAT THE SEAT DOES INSTEAD IS ON ITS SURFACE AND ADMITTED: the gap goes before the
			// bench, and blue says what it does with the avenue in a move.
			d.docketMotion("a docket motion on the gap whose grade ruling the seat disagrees with", seatID)
			if seatID == "blue-respond" {
				d.admitted("a move of the avenue against its ruling", seatID, "avenue", "move", "--id", "Q1", "--as", "pursued",
					"--reason", "the ruling reads the question too narrowly, and this line answers it")
			}
		})
	}
}

// A seat's invocation writes no appeal: Append refuses the body outside a migration, whose replay
// of an archived record is the one writer left.
func TestNoSeatWritesAnAppeal(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := actFixture(t)
	m := fileGradeMotion(t, runDir)
	must(t, runDir, "motion", "grade", "rule", "--seat-id", "red-chair", "--id", m, "--as", "rejected", "--reason", "the evidence does not reach it")
	run, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	subject := recordpb.MotionSubject_MOTION_SUBJECT_GRADE
	body := &recordpb.MotionAppeal{MotionId: proto.String(m), Subject: &subject, Reason: proto.String("the ruling ignored the source")}
	for _, seatID := range []string{"blue-respond", lensSeat, "red-chair", "judge"} {
		if _, err := record.Append(record.Identity{Run: run, SeatID: seatID}, body); err == nil || !strings.Contains(err.Error(), "written by migrate") {
			t.Fatalf("Append admitted an appeal by %s outside a migration: %v", seatID, err)
		}
	}
	if n := len(eventsOfWord(t, runDir, "motion_appeal")); n != 0 {
		t.Fatalf("the record holds %d appeal(s)", n)
	}
	record.Migrating = true
	_, err = record.Append(record.Identity{Run: run, SeatID: "blue-respond"}, body)
	record.Migrating = false
	if err != nil {
		t.Fatalf("a migration's replay of an archived appeal was refused: %v", err)
	}
	if n := len(eventsOfWord(t, runDir, "motion_appeal")); n != 1 {
		t.Fatalf("the replay wrote %d appeal(s), want one", n)
	}
}
