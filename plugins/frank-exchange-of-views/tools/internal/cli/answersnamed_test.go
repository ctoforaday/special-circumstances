package cli

import (
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// mintAboutKind is the lens's mint of a gap about kind, ref ("" for none).
func mintAboutKind(t *testing.T, runDir, key, kind, ref string) string {
	t.Helper()
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	args := []string{"mint", "--about-kind", kind, "--run", runDir, "--seat-id", lensSeat,
		"--key", key, "--class", "spec_underspecification", "--problem", "the defect", "--fix", "the fix",
		"--check-kind", "document", "--check", "the acceptance check red runs at re-audit",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low"}
	if ref != "" {
		args = append(args, "--about", ref)
	}
	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("mint %s about %s %q: %v %s", key, kind, ref, err, out)
	}
	return gapID(out)
}

// answersOn is the `answers` each event of the record of type T carries, in record order.
func answersOn[T interface {
	proto.Message
	GetAnswers() string
}](t *testing.T, runDir string) []string {
	t.Helper()
	all, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range all.Events {
		if b, ok := recordpb.BodyAs[T](e); ok {
			out = append(out, b.GetAnswers())
		}
	}
	return out
}

// T-B45 (the named arm): AN INQUIRY FATE AND A QUESTION PART ANSWER A GAP ONLY BY NAMING IT, with
// --answers on `inquiry record`, `question cover` and `question split`. The name lands on the record's
// field; one naming a gap whose subject the act does not change, or by a seat that is not the gap's
// author, is refused at the write, naming why.
func TestAFateOrAPartAnswersTheGapItNames(t *testing.T) {
	runDir := seatRun(t)
	registerBlue(t, runDir)
	out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", blueSeat, "--question", "whether trial division to 9 settles it", "--json")
	if err != nil {
		t.Fatalf("inquiry propose: %v %s", err, out)
	}
	q := regexp.MustCompile(`Q-[0-9a-f]{8}`).FindString(out)
	out, err = run(t, "question", "split", "--run", runDir, "--seat-id", blueSeat, "--text", "whether 91 is prime", "--json")
	if err != nil {
		t.Fatalf("question split: %v %s", err, out)
	}
	part := regexp.MustCompile(`S-[0-9a-f]{8}`).FindString(out)
	if q == "" || part == "" {
		t.Fatalf("no inquiry or part id: %q %q", q, part)
	}
	onInquiry := mintAboutKind(t, runDir, "on-inquiry", "inquiry", q)
	onPart := mintAboutKind(t, runDir, "on-part", "question", part)
	onQuestion := mintAboutKind(t, runDir, "on-question", "question", "")
	onSection := mintGap(t, runDir, "on-findings", "spec_underspecification")
	onFuture := mintAboutKind(t, runDir, "on-future", "section", "Future research directions")
	registerBlue(t, runDir)

	for _, c := range []struct {
		name string
		args []string
		want string // "" for landing; else a phrase of the refusal
	}{
		{"a fate naming the gap about its inquiry", []string{"inquiry", "record", "--id", q, "--as", "confirmed", "--reason", "7 divides it", "--answers", onInquiry}, ""},
		{"a fate naming a gap about a section blue writes", []string{"inquiry", "record", "--id", q, "--as", "confirmed", "--reason", "7 divides it, again", "--answers", onSection}, "changes nothing"},
		{"a fate moving its inquiry into a section, naming the gap about that section", []string{"inquiry", "record", "--id", q, "--as", "deferred", "--reason", "a search of printed tables would settle it", "--answers", onFuture}, ""},
		{"a coverage naming the gap about its part", []string{"question", "cover", "--id", part, "--as", "covered", "--reason", "the report answers it", "--answers", onPart}, ""},
		{"a coverage naming the gap about the question", []string{"question", "cover", "--id", part, "--as", "covered", "--reason", "the report answers it, again", "--answers", onQuestion}, "changes nothing"},
		{"a split naming the gap about the question", []string{"question", "split", "--text", "whether 91 is composite", "--answers", onQuestion}, ""},
		{"a split naming the gap about a part", []string{"question", "split", "--text", "whether 91 is odd", "--answers", onPart}, "changes nothing"},
	} {
		t.Run(c.name, func(t *testing.T) {
			args := append(append([]string{}, c.args[:2]...), "--run", runDir, "--seat-id", blueSeat)
			args = append(args, c.args[2:]...)
			out, err := run(t, args...)
			switch {
			case c.want == "" && err != nil:
				t.Fatalf("%v: %v %s", args, err, out)
			case c.want != "" && (err == nil || !strings.Contains(err.Error()+out, c.want)):
				t.Fatalf("%v = %v %q; want refused naming %q", args, err, out, c.want)
			}
		})
	}
	if got := answersOn[*recordpb.Inquiry](t, runDir); !containsAll(got, onInquiry, onFuture) || containsAll(got, onSection) {
		t.Errorf("the inquiry records carry answers %q; want %s and never %s", got, onInquiry, onSection)
	}
	if got := answersOn[*recordpb.QuestionPart](t, runDir); !containsAll(got, onPart, onQuestion) {
		t.Errorf("the question parts carry answers %q; want %s and %s", got, onPart, onQuestion)
	}

	// AN INQUIRY KIND HAS NO STATEMENT TO ACCEPT (#1942, "as built"): none of its verbs mounts --accept.
	for _, verb := range [][]string{{"inquiry", "record", "--id", q, "--as", "confirmed", "--reason", "r"}, {"question", "cover", "--id", part, "--as", "covered", "--reason", "r"}, {"question", "split", "--text", "t"}} {
		args := append(append([]string{}, verb[:2]...), "--run", runDir, "--seat-id", blueSeat, "--accept", onInquiry)
		args = append(args, verb[2:]...)
		if out, err := run(t, args...); err == nil || !strings.Contains(err.Error()+out, "unknown flag: --accept") {
			t.Errorf("%s %s --accept = %v %q; want no such flag", verb[0], verb[1], err, out)
		}
	}

	t.Run("a seat that is not the gap's author", func(t *testing.T) {
		out, err := run(t, "inquiry", "record", "--run", runDir, "--seat-id", "blue-synthesize", "--id", q, "--as", "confirmed", "--reason", "7 divides it, so", "--answers", onInquiry)
		if err == nil || !strings.Contains(err.Error()+out, "only a gap's author answers it") {
			t.Errorf("blue-synthesize answering %s = %v %q; want refused naming the author", onInquiry, err, out)
		}
	})
}

// containsAll reports whether every one of want is in got.
func containsAll(got []string, want ...string) bool {
	for _, w := range want {
		found := false
		for _, g := range got {
			found = found || g == w
		}
		if !found {
			return false
		}
	}
	return true
}
