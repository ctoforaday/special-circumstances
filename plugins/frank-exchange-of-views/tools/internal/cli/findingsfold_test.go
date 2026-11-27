package cli

import (
	"strings"
	"testing"
)

// A LENS REPORTS A DEFECT ONE WAY: AS A GAP (plans/findings-fold.md Part A). There is no finding
// verb, no findings view and no attribution flag to name one.
func TestNoSurfaceMountsAFinding(t *testing.T) {
	runDir := seatRun(t)
	for _, c := range []struct {
		args []string
		want string // the refusal that says the surface holds no such thing
	}{
		{[]string{"finding", "--run", runDir, "--seat-id", lensSeat, "--quote", "x", "--reason", "y"},
			`no command named "finding" exists, on your surface or on any other seat's`},
		{[]string{"show", "findings", "--run", runDir, "--seat-id", lensSeat}, `no projection named "findings"`},
		{[]string{"mint", "--run", runDir, "--seat-id", lensSeat, "--found-by", "F-00000001"}, "unknown flag: --found-by"},
	} {
		out, err := run(t, c.args...)
		if err == nil {
			t.Errorf("%q was accepted: %s", strings.Join(c.args, " "), out)
			continue
		}
		if msg := err.Error() + out; !strings.Contains(msg, c.want) {
			t.Errorf("%q was refused without %q:\n%s", strings.Join(c.args, " "), c.want, msg)
		}
	}
}

// The roads, in the words each case's refusal states them; none names a finding, a no-act docket
// motion or reconsideration grounds.
const (
	roadParties = "a gap has two parties"
	roadOnBoard = "nothing more is recorded"
)

var roadNever = []string{"finding", "--grounds", "--about-kind gap"}

// roadRefusal fails the test unless err is a refusal holding every want and none of never.
func roadRefusal(t *testing.T, err error, out string, want, never []string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the mint was admitted: %s", out)
	}
	msg := err.Error()
	for _, w := range want {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal does not say %q:\n%s", w, msg)
		}
	}
	for _, n := range append(append([]string{}, roadNever...), never...) {
		if strings.Contains(msg, n) {
			t.Errorf("the refusal says %q:\n%s", n, msg)
		}
	}
}

// THE ONE-GAP REFUSAL NAMES THE REFUSED SEAT'S ROAD, per case, on every arm of the rule (plan
// §III.A2, T-A3). The refused seat's own gap is reweighed by regrading it; another seat's gap already carries a point that agrees with it, and a dispute of it is a
// proposed act, with the exchange sentence while the gap has had none.
func TestTheOneGapRefusalNamesTheSeatsRoad(t *testing.T) {
	const plain = "spec_underspecification"
	mint := func(b readingBoard, seat string, subject ...string) (string, string, error) {
		return b.mintSource(seat, plain, subject...)
	}
	arms := []struct {
		name          string
		first, second func(b readingBoard, q string) []string
	}{
		{"the inquiry arm",
			func(_ readingBoard, q string) []string { return []string{"--about-kind", "inquiry", "--about", q} },
			func(_ readingBoard, q string) []string { return []string{"--about-kind", "inquiry", "--about", q} }},
		{"the anchor arm",
			func(b readingBoard, _ string) []string { return about(b.c1) },
			func(b readingBoard, _ string) []string { return about(b.c1) }},
		{"the cross-arm, about over quote",
			func(readingBoard, string) []string { return []string{"--quote", rdBacked} },
			func(b readingBoard, _ string) []string { return about(b.c2) }},
		{"the quote arm",
			func(readingBoard, string) []string { return []string{"--quote", rdFree1} },
			func(readingBoard, string) []string { return []string{"--quote", rdFree1} }},
		{"the cross-arm, quote over about",
			func(b readingBoard, _ string) []string { return about(b.c1) },
			func(readingBoard, string) []string { return []string{"--quote", rdCited} }},
	}
	for _, a := range arms {
		t.Run(a.name+", own gap", func(t *testing.T) {
			b := readingRun(t)
			q := proposeInquiry(t, b.runDir, "whether costs rose")
			held, out, err := mint(b, lensL, a.first(b, q)...)
			if err != nil {
				t.Fatalf("the first mint: %v\n%s", err, out)
			}
			_, out, err = mint(b, lensL, a.second(b, q)...)
			roadRefusal(t, err, out,
				[]string{held, "is yours", "`regrade --id " + held + "`", "Nothing was written"},
				[]string{"motion docket", roadOnBoard, roadParties})
		})
		t.Run(a.name+", another seat's gap", func(t *testing.T) {
			b := readingRun(t)
			q := proposeInquiry(t, b.runDir, "whether costs rose")
			held, out, err := mint(b, lensM, a.first(b, q)...)
			if err != nil {
				t.Fatalf("the first mint: %v\n%s", err, out)
			}
			_, out, err = mint(b, lensL, a.second(b, q)...)
			roadRefusal(t, err, out,
				[]string{held, "raised by " + lensM, roadOnBoard, roadParties, "Nothing was written"},
				[]string{"is yours", "`regrade --id"})
		})
	}

	// A gap is not the subject of another gap: `--about-kind gap` is refused naming the parties.
	t.Run("about-kind gap", func(t *testing.T) {
		b := readingRun(t)
		held, out, err := mint(b, lensM, "--quote", rdFree1)
		if err != nil {
			t.Fatalf("the first mint: %v\n%s", err, out)
		}
		_, out, err = mint(b, lensL, "--about-kind", "gap", "--about", held)
		if err == nil {
			t.Fatalf("--about-kind gap was admitted: %s", out)
		}
		// The whole first line, so a prefix the flag frame adds and the refusal repeats is caught.
		want := `--about-kind takes no "gap": a gap is not the subject of another gap. ` +
			"A gap has two parties — the lens that raised it and its subject's author — and only they act on it: " +
			"what you see wrong in another lens's open gap is its raiser's to answer, and a defect that stands in the report after it closes is a new gap of yours. Nothing was written"
		if got, _, _ := strings.Cut(err.Error(), "\n"); got != want {
			t.Errorf("the refusal reads:\n%s\nwant:\n%s", got, want)
		}
		for _, n := range []string{"finding", "--grounds"} {
			if strings.Contains(err.Error(), n) {
				t.Errorf("the refusal says %q:\n%v", n, err)
			}
		}
	})
}
