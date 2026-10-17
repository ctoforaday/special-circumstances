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
	return subject == "inquiry" && seat.RecordType(c) == "inquiry"
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
		roleLeaves(func(role string, path []string, c *cobra.Command) {
			if filingOf(subject, path, c) {
				filers[role] = true
			}
			if len(path) == 3 && path[0] == "motion" && path[1] == subject && seat.RecordType(c) == "motion_rule" {
				rules[role] = true
			}
		})
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
	grade := []string{"motion", "grade", "file", "--id", "G1", "--dimension", "severity", "--proposed", "high", "--reason", "the consequence is larger"}
	docket := []string{"motion", "docket", "file", "--id", "G1", "--reason", "the bench should decide"}
	petition := []string{"motion", "petition", "file", "--class", "safety", "--relief", "halt", "--reason", "a consent gate is missing"}
	docketRule := func(m string) []string {
		return []string{"motion", "docket", "rule", "--id", m, "--as", "remanded", "--principle", "thoroughness over speed",
			"--tension", "cost against certainty", "--review-flag", "none", "--settled", "nothing yet", "--direction", "a reproduction", "--reason", "one more exchange"}
	}

	for _, tc := range []struct {
		name, seatID string
		act          []string
		wants        []string
	}{
		{"the chair files a grade motion", "red-chair", grade,
			[]string{"a grade motion is ruled by the chair seat", "`file` is not on your surface", "you are red-chair (chair)",
				"The bench, blue and lens seats file one", "file `motion docket file` on the gap, which the bench rules"}},
		{"the bench files a docket motion", "judge", docket,
			[]string{"a docket motion is ruled by the bench seat", "`file` is not on your surface", "you are judge (bench)",
				"The blue, chair and lens seats file one"}},
		{"the bench files a petition", "judge", petition,
			[]string{"a petition motion is ruled by the bench seat", "`file` is not on your surface", "you are judge (bench)",
				"The blue, chair and lens seats file one", "`halt` is yours"}},
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

	t.Run("the chair files a docket motion and the bench rules it", func(t *testing.T) {
		d := fresh(t)
		m := docketFile(t, d.run, "red-chair", "G1", "red cannot settle G1")
		d.admitted("the bench's ruling", "judge", docketRule(m)...)
	})
	for _, filer := range []string{"blue-respond", lensSeat, "judge"} {
		t.Run("a grade motion filed by "+filer+" is ruled by the chair", func(t *testing.T) {
			d := fresh(t)
			m := motionID(t, d.run, append(append([]string{}, grade...), "--seat-id", filer)...)
			d.admitted("the chair's ruling", "red-chair", "motion", "grade", "rule", "--id", m, "--as", "rejected", "--reason", "the evidence does not reach it")
		})
	}
	for _, filer := range []string{"blue-respond", lensSeat, "red-chair"} {
		t.Run("a petition filed by "+filer+" is ruled by the bench", func(t *testing.T) {
			d := fresh(t)
			m := motionID(t, d.run, append(append([]string{}, petition...), "--seat-id", filer)...)
			d.admitted("the bench's ruling", "judge", "motion", "petition", "rule", "--id", m, "--as", "denied", "--reason", "no boundary is crossed")
		})
		t.Run("a docket motion filed by "+filer+" is ruled by the bench", func(t *testing.T) {
			d := fresh(t)
			m := docketFile(t, d.run, filer, "G1", "the bench should decide G1")
			d.admitted("the bench's ruling", "judge", docketRule(m)...)
		})
	}
}
