package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// mintOnGap is seat's mint of a gap about gap on, through the command tree.
func mintOnGap(t *testing.T, runDir, seat, key, on string) (string, error) {
	t.Helper()
	return mintOnGapChecked(t, runDir, seat, key, on, "document")
}

// mintOnGapChecked is mintOnGap with the check kind named.
func mintOnGapChecked(t *testing.T, runDir, seat, key, on, checkKind string) (string, error) {
	t.Helper()
	out, err := run(t, "mint", "--run", runDir, "--seat-id", seat, "--key", key, "--about-kind", "gap", "--about", on,
		"--class", "spec_underspecification", "--problem", "the gap's check cannot fail", "--fix", "state a check that can fail",
		"--check-kind", checkKind, "--check", "the check names an observable outcome",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low")
	return gapID(out), withOut(err, out)
}

// withOut is err carrying out, so a refusal's envelope text is read with it.
func withOut(err error, out string) error {
	if err == nil {
		return nil
	}
	return &outErr{err, out}
}

type outErr struct {
	err error
	out string
}

func (e *outErr) Error() string { return e.err.Error() + "\n" + e.out }

// T-B44 (the mint): A GAP ON A GAP IS ANOTHER LENS'S OPEN GAP. A lens mints a gap about another lens's
// open gap — a defect in its printed wording or grades — and that gap's minter is the new gap's author.
// Refused, each recording nothing: a gap nothing minted, the minter's own gap (its road is amend or
// regrade), a gap that is itself about a gap, and a closed gap.
func TestAGapOnAGapIsAnotherLenssOpenGap(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "under", "spec_underspecification")
	asSeat(t, runDir, otherLens, "register")
	g2, err := mintOnGap(t, runDir, otherLens, "on-gap", g)
	if err != nil {
		t.Fatalf("a gap on another lens's open gap: %v", err)
	}
	if author, err := record.AuthorOfGap(runtest.Open(t, runDir), g2); err != nil || author != lensSeat {
		t.Errorf("the author of %s, about %s, is %q (%v), want %s's minter %s", g2, g, author, err, g, lensSeat)
	}
	refusedChecked := func(name, seat, on, checkKind string, want ...string) {
		t.Helper()
		mints := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT)
		_, err := mintOnGapChecked(t, runDir, seat, "refused-"+strings.ReplaceAll(name, " ", "-"), on, checkKind)
		if err == nil {
			t.Errorf("%s: the mint landed", name)
			return
		}
		for _, w := range want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("%s: the refusal does not say %q: %v", name, w, err)
			}
		}
		if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); got != mints {
			t.Errorf("%s: mints on the record %d, want %d", name, got, mints)
		}
	}
	refused := func(name, seat, on string, want ...string) {
		t.Helper()
		refusedChecked(name, seat, on, "document", want...)
	}
	// A GAP ON A GAP TAKES CHECK KIND document: its answer, an amend or regrade, places no evidence.
	for _, ck := range []string{"source", "computation"} {
		refusedChecked("check kind "+ck, otherLens, g, ck, "--check-kind "+ck, "takes check kind document only")
	}
	refused("a gap nothing minted", otherLens, "G-0badf00d", "G-0badf00d", "no mint event created")
	refused("the minter's own gap", lensSeat, g, "your own gap", "amend --id "+g, "regrade --id "+g)
	refused("a gap on a gap on a gap", lensSeat, g2, "itself a gap about gap "+g, "gap on a gap on a gap is refused")
	if out, err := runAt(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--as", "not_a_defect",
		"--verified-by", lensSeat, "--verified-with", "a re-read of the section", "--verified-against", "the report",
		"--reason", "the section states it after all"); err != nil {
		t.Fatalf("the close of %s: %v %s", g, err, out)
	}
	refused("a closed gap", otherLens, g, g, "CLOSED")
}

// T-B44 (the answer): G'S MINTER ANSWERS A GAP ON G BY CHANGING G AND NAMING IT. A's regrade of G with
// --answers G2 lands carrying the answer; --answers naming a gap that is not about G — another gap of
// the report, or G itself — is refused and records nothing, and so does B's regrade of A's G.
func TestTheMintersRegradeNamingAGapOnItsGapAnswersIt(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "under", "spec_underspecification")
	other := mintGap(t, runDir, "other", "unsourced_claim")
	asSeat(t, runDir, otherLens, "register")
	g2, err := mintOnGap(t, runDir, otherLens, "on-gap", g)
	if err != nil {
		t.Fatalf("a gap on %s: %v", g, err)
	}
	onOther, err := mintOnGap(t, runDir, otherLens, "on-other", other)
	if err != nil {
		t.Fatalf("a gap on %s: %v", other, err)
	}
	regrade := func(seat, answers string) (string, error) {
		return runAt(t, "regrade", "--run", runDir, "--seat-id", seat, "--id", g, "--answers", answers,
			"--severity", "low", "--reason", "the gap on it is right about the grade")
	}
	for _, c := range []struct{ name, seat, answers, want string }{
		{"a gap about the report", lensSeat, other, "only a gap's author answers it"},
		{"the regraded gap itself", lensSeat, g, "only a gap's author answers it"},
		{"a gap on another gap of the minter's", lensSeat, onOther, "it changes nothing " + onOther + " is about"},
		{"another lens's regrade of the gap", otherLens, g2, otherLens + " cannot"},
	} {
		regrades := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REGRADE)
		out, err := regrade(c.seat, c.answers)
		heldRefusal(t, c.name, out, err, c.want)
		if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REGRADE); got != regrades {
			t.Errorf("%s: regrades on the record %d, want %d", c.name, got, regrades)
		}
	}
	if out, err := regrade(lensSeat, g2); err != nil {
		t.Fatalf("the minter's regrade naming %s: %v %s", g2, err, out)
	}
	if rg := lastBody(t, runDir, &recordpb.Regrade{}); rg.GetGapId() != g || rg.GetAnswers() != g2 {
		t.Errorf("the regrade on the record is %v, want of %s answering %s", rg, g, g2)
	}
}
