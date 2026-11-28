package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// propose runs `motion docket file <act> --id gap` as filer with the act's own flags.
func propose(t *testing.T, runDir, filer, act, gap string, flagsOf ...string) (string, error) {
	t.Helper()
	args := append([]string{"motion", "docket", "file", act, "--run", runDir, "--seat-id", filer, "--id", gap,
		"--reason", "the act this gap wants", "--json"}, flagsOf...)
	return runAt(t, args...)
}

// closeFlags is a close proposal's own flags: its disposition and the verification triple.
func closeFlags(class string) []string {
	return []string{"--as", class, "--verified-by", "blue-respond", "--verified-with", "the report read",
		"--verified-against", "the report's sentence"}
}

// proposalRow is the act a stored motion carries, read off the record's act tables.
func proposalRow(t *testing.T, runDir, motionID string) (act, actGap string) {
	t.Helper()
	p, found, err := record.DocketProposalOf(runtest.Open(t, runDir), motionID)
	if err != nil || !found {
		t.Fatalf("motion %s: found %v, %v", motionID, found, err)
	}
	switch a := p.Motion.GetAct().(type) {
	case *recordpb.Motion_Regrade:
		return "regrade", a.Regrade.GetGapId()
	case *recordpb.Motion_Close:
		return "close", a.Close.GetGapId()
	}
	return "", ""
}

// T-B2: A DOCKET MOTION PROPOSES ONE ACT, stored in that act's own table. A bare `file` is refused
// naming the acts; each act lands; each act's shape binds at the filing.
func TestADocketMotionStoresTheActItProposes(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "proposal-shape", "spec_underspecification")
	exchangeOn(t, runDir, g)

	if out, err := runAt(t, "motion", "docket", "file", "--run", runDir, "--seat-id", blueSeat, "--id", g, "--reason", "r"); err == nil ||
		!strings.Contains(err.Error()+out, "regrade, close") {
		t.Errorf("a bare `motion docket file` = %v %q, want refused naming the acts", err, out)
	}

	// Each act's shape binds at the filing, and nothing is stored.
	for name, c := range map[string]struct {
		act   string
		flags []string
		want  string
	}{
		"a regrade that moves no axis": {"regrade", []string{"--severity", "medium"}, "moves no axis"},
	} {
		if out, err := propose(t, runDir, blueSeat, c.act, g, c.flags...); err == nil || !strings.Contains(err.Error()+out, c.want) {
			t.Errorf("%s = %v %q, want refused with %q", name, err, out, c.want)
		}
	}
	// Each act lands in its own table, on the motion's gap.
	for _, c := range []struct {
		act   string
		flags []string
	}{
		{"regrade", []string{"--severity", "low"}},
		{"close", closeFlags("not_a_defect")},
	} {
		h := mintGap(t, runDir, "proposal-"+c.act, "policy_without_mechanism")
		exchangeOn(t, runDir, h)
		out, err := propose(t, runDir, blueSeat, c.act, h, c.flags...)
		if err != nil {
			t.Fatalf("%s proposal: %v", c.act, err)
		}
		act, actGap := proposalRow(t, runDir, filedMotionID(t, out))
		switch {
		case act != c.act:
			t.Errorf("the %s proposal stored act %q", c.act, act)
		case actGap != h:
			t.Errorf("the %s proposal names gap %q, want %q", c.act, actGap, h)
		}
	}
}

// T-B3: A PROPOSAL IS REFUSED where no seat should file it: on the filer's own gap (its assignee
// takes the act itself), before the gap's first exchange — NOT MEASURED where its sittings cannot be
// closed — while another motion stands on it, and on a closed gap.
func TestAProposalIsRefusedWhereItCannotStand(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "proposal-refused", "spec_underspecification")

	if out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low"); err == nil || !strings.Contains(err.Error()+out, "0 exchange(s)") {
		t.Errorf("a proposal before any exchange = %v %q, want refused at 0 exchanges", err, out)
	}
	// Blue engaged and sat once: its sitting is open, so the count is not measured, never zero.
	ensureCast(t, runDir)
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"}, dispatchOf("blue-respond", g)); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-respond"); err != nil {
		t.Fatal(err)
	}
	if out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low"); err == nil || !strings.Contains(err.Error()+out, "NOT MEASURED") {
		t.Errorf("a proposal on a gap whose sitting the record cannot close = %v %q, want NOT MEASURED", err, out)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-respond"); err != nil {
		t.Fatal(err)
	}

	if out, err := propose(t, runDir, lensSeat, "regrade", g, "--severity", "low"); err == nil || !strings.Contains(err.Error()+out, "you are its raiser") {
		t.Errorf("the assignee's proposal on its own gap = %v %q, want refused naming its own act", err, out)
	}
	out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
	if err != nil {
		t.Fatalf("blue's proposal after one exchange: %v", err)
	}
	first := filedMotionID(t, out)
	if out, err := propose(t, runDir, blueSeat, "close", g, closeFlags("not_a_defect")...); err == nil || !strings.Contains(err.Error()+out, first+" stands unanswered") {
		t.Errorf("a second proposal while %s stands = %v %q, want refused naming it", first, err, out)
	}

	h := mintGap(t, runDir, "proposal-closed", "policy_without_mechanism")
	exchangeOn(t, runDir, h)
	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", h, "--as", "not_a_defect",
		"--verified-by", lensSeat, "--verified-with", "the read", "--verified-against", "the sentence", "--reason", "nothing to repair"); err != nil {
		t.Fatalf("closing %s: %v", h, err)
	}
	if out, err := propose(t, runDir, blueSeat, "regrade", h, "--severity", "low"); err == nil || !strings.Contains(err.Error()+out, "is closed") {
		t.Errorf("a regrade proposal on a closed gap = %v %q, want refused", err, out)
	}
}
