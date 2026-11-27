package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RULER DOES NOT FILE THE SUBJECT IT RULES. Each subject has one gavel, so a motion filed by the
// seat that holds it is a motion whose only ruler is its filer: nothing but that seat's own word
// answers it, and while it stands unruled it holds the PASS gate with no other seat able to clear
// it. These tests read the gavel from the schema (recordpb.SubjectRuler) and the verbs from the
// trees a seat is handed, so a subject added later is held to the same rule without a row here.

// filingOf reports whether a leaf FILES a motion on the subject: the subject's own `file`, which
// records a motion — or, for the one subject whose filing is another act, the proposal.
func filingOf(subject string, path []string, c *cobra.Command) bool {
	if len(path) >= 2 && path[0] == "motion" && path[1] == subject && seat.RecordType(c) == "motion" {
		return true
	}
	// The inquiry's filing is the proposal, and the position's is the chair's position.
	return (subject == "inquiry" && seat.RecordType(c) == "inquiry") || (subject == "position" && seat.RecordType(c) == "position")
}

func TestNoGavelHolderFilesItsOwnSubject(t *testing.T) {
	for _, subject := range record.MotionSubjects {
		subj, ok := record.MotionSubjectEnum(subject)
		if !ok {
			t.Fatalf("%s is in MotionSubjects and names no MotionSubject", subject)
		}
		ruler, err := recordpb.SubjectRuler(subj)
		if err != nil {
			t.Fatal(err)
		}
		filers := map[string]bool{}
		rules := map[string]bool{}
		accepts := map[string]bool{}
		roleLeaves(func(role string, path []string, c *cobra.Command) {
			if filingOf(subject, path, c) {
				filers[role] = true
			}
			if len(path) == 3 && path[0] == "motion" && path[1] == subject && seat.RecordType(c) == "motion_rule" {
				// A docket proposal's raiser takes it by its own act carrying the motion (accepts),
				// never a ruling, so no motion verb but `rule` writes a motion_rule.
				if path[2] == "accept" {
					accepts[role] = true
					return
				}
				rules[role] = true
			}
		})
		if len(accepts) != 0 {
			t.Errorf("%s: `accept` is on %v: a proposal is taken by its raiser's own act, never a ruling", subject, accepts)
		}
		if filers[ruler] {
			t.Errorf("%s: the %s seat rules it and holds a verb that files it — a motion its own ruler filed is one no other seat can rule", subject, ruler)
		}
		if len(rules) != 1 || !rules[ruler] {
			t.Errorf("%s: `rule` is on %v, want the %s surface alone", subject, rules, ruler)
		}
		// Not vacuous: somebody files every subject, or the walk stopped reaching the verbs.
		if len(filers) == 0 {
			t.Errorf("%s: no seat holds a verb that files it — the walk or the trees moved", subject)
		}
	}
}

func TestARulerFilingItsOwnSubjectIsRefusedAtTheSurface(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	fresh := func(t *testing.T) drive {
		runDir := actFixture(t)
		if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
			t.Fatal(err)
		}
		return drive{t: t, run: runDir}
	}
	docket := []string{"motion", "docket", "file", "--id", "G1", "--reason", "the bench should decide"}
	petition := []string{"motion", "petition", "file", "--class", "safety", "--relief", "halt", "--reason", "a consent gate is missing"}
	docketRule := func(m string) []string {
		return []string{"motion", "docket", "rule", "--id", m, "--remand", "--principle", "thoroughness over speed",
			"--tension", "cost against certainty", "--review-flag", "none", "--settled", "nothing yet", "--direction", "a reproduction", "--reason", "one more exchange"}
	}

	for _, tc := range []struct {
		name, seatID string
		act          []string
		wants        []string
	}{
		{"the bench files a docket motion", "judge", docket,
			[]string{"a docket motion is ruled by the bench seat", "`file` is not on your surface", "you are judge (bench)",
				"The blue and lens seats file one"}},
		{"the bench files a petition", "judge", petition,
			[]string{"a petition motion is ruled by the bench seat", "`file` is not on your surface", "you are judge (bench)",
				"The blue, chair, lane and lens seats file one", "`halt` is yours"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := fresh(t)
			refusal := d.refused(tc.name, tc.seatID, tc.act...)
			for _, w := range tc.wants {
				if !strings.Contains(refusal, w) {
					t.Errorf("the refusal does not say %q:\n%s", w, refusal)
				}
			}
			if n := len(eventsOfWord(t, d.run, "motion")); n != 0 {
				t.Errorf("the refused filing left %d motion(s) on the record", n)
			}
		})
	}

	t.Run("the author files a docket motion and the bench rules it", func(t *testing.T) {
		d := fresh(t)
		g := handleText(t, d.run, "G1")
		m := docketFile(t, d.run, blueSeat, g, "red cannot settle G1")
		k := docketNow(t, d.run, g, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
		conveneBench(t, d.run, g)
		d.admitted("the bench's ruling", "judge", docketRule(k)...)
	})
	for _, filer := range []string{"blue-respond", lensSeat, "red-chair"} {
		t.Run("a petition filed by "+filer+" is ruled by the bench", func(t *testing.T) {
			d := fresh(t)
			m := motionID(t, d.run, append(append([]string{}, petition...), "--seat-id", filer)...)
			d.admitted("the bench's ruling", "judge", "motion", "petition", "rule", "--id", m, "--as", "denied", "--reason", "no boundary is crossed")
		})
		// G1 is the lens's gap: its raiser takes the act itself, and its proposal is refused.
		if filer == lensSeat {
			t.Run("a docket motion on its own gap, filed by "+filer+", is refused", func(t *testing.T) {
				d := fresh(t)
				g := handleText(t, d.run, "G1")
				exchangeOn(t, d.run, g)
				out, err := runAt(t, closeProposalArgs(d.run, filer, g, "not_a_defect", "the bench should decide G1")...)
				if err == nil || !strings.Contains(err.Error()+out, "you are its raiser") {
					t.Errorf("the assignee's proposal on its own gap = %v %q, want refused naming its own act", err, out)
				}
			})
			continue
		}
		// A docket motion is its gap's author's alone (X6): the chair files none.
		if filer != blueSeat {
			continue
		}
		t.Run("a docket motion filed by "+filer+" is ruled by the bench", func(t *testing.T) {
			d := fresh(t)
			g := handleText(t, d.run, "G1")
			m := docketFile(t, d.run, filer, g, "the bench should decide G1")
			k := docketNow(t, d.run, g, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
			conveneBench(t, d.run, g)
			d.admitted("the bench's ruling", "judge", docketRule(k)...)
		})
	}
}
