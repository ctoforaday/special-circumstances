package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A REPEATED PROPOSAL IS THE SAME LINE OF INQUIRY, because a line of inquiry is identified by what it
// says, and the seat that repeats one is retrying rather than proposing.
//
// MEASURED on universe-m12: a blue seat's propose script ran, wrote five lines, and returned NO
// OUTPUT AT ALL to the seat. It ran the identical script again — the only reasonable act on an answer
// you never saw — and five more ids were minted for five byte-identical hypotheses. The run then
// carried TEN lines of inquiry for five questions, blue moved all ten to pursued, and a lens filed a
// defect about the method count being redundant. Nothing refused any of it: the idempotency key is
// `<seat>:<verb>:#<ordinal>`, so a repeat is by design a new act, and `propose` had no --key.
func TestARepeatedProposalReturnsTheSameLine(t *testing.T) {
	runDir := seatRun(t)
	const line = "trial division to the square root settles it"
	const hyp = "every composite has a factor at or below its square root"

	first, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond",
		"--reason", line, "--hypothesis", hyp)
	if err != nil {
		t.Fatal(err)
	}
	second, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond",
		"--reason", line, "--hypothesis", hyp)
	if err != nil {
		t.Fatalf("the retry was refused rather than answered: %v", err)
	}

	id := func(out string) string {
		for _, f := range strings.Fields(out) {
			if strings.HasPrefix(f, "Q") {
				return f
			}
		}
		return ""
	}
	if a, b := id(first), id(second); a == "" || a != b {
		t.Errorf("the retry recorded %q where the first call recorded %q — two ids for one line, which is "+
			"how a run ends up with ten lines of inquiry for five questions", b, a)
	}
	// AND IT SAYS SO, because a seat that cannot tell a retry from a fresh record will not know its
	// ledger is intact.
	if !strings.Contains(second, "idempotent") {
		t.Errorf("the retry does not say it wrote nothing: %q", second)
	}
	// NOTHING WAS WRITTEN. The id matching is not enough on its own: an implementation that wrote a
	// second event carrying the FIRST id would pass the check above and still double the ledger.
	proposals := 0
	for _, e := range events(t, runDir) {
		if a, ok := recordpb.BodyAs[*recordpb.Avenue](e); ok && a.GetStatus() == recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED {
			proposals++
		}
	}
	if proposals != 1 {
		t.Errorf("the record holds %d proposals for one line", proposals)
	}

	// A DIFFERENT HYPOTHESIS IS A DIFFERENT LINE, and this is the arm that keeps the dedup honest:
	// --reason is required and --hypothesis is not, so matching the line alone would collapse two
	// proposals that share required prose and test different claims.
	third, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond",
		"--reason", line, "--hypothesis", "a different claim entirely")
	if err != nil {
		t.Fatal(err)
	}
	if id(third) == id(first) {
		t.Errorf("a proposal testing a different claim was folded into %s — the line is the same prose, "+
			"but the hypothesis is what a later abandonment is judged against", id(first))
	}
}
