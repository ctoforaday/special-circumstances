package cli

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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
	roadSupersede = "motion docket file supersede|regrade|close --id "
	roadOnBoard   = "nothing more is recorded"
	roadExchange  = "a proposal on it is admitted once it has had one exchange"
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
// §III.A2, T-A3). The refused seat's own gap is restated by superseding it and reweighed by regrading
// it; another seat's gap already carries a point that agrees with it, and a dispute of it is a
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
				[]string{held, "is yours", "`--supersedes " + held + "`", "`regrade --id " + held + "`", "Nothing was written"},
				[]string{"motion docket", roadOnBoard, roadExchange})
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
				[]string{held, "assigned to " + lensM, roadOnBoard, roadSupersede + held, roadExchange, "Nothing was written"},
				[]string{"is yours", "regrade --id"})
		})
	}

	// The exchange sentence is printed while the gap has had no exchange, worded NOT MEASURED where
	// the record cannot close a sitting engaged on it, and not at all once it has had one.
	t.Run("the exchange sentence", func(t *testing.T) {
		b := readingRun(t)
		held, out, err := mint(b, lensM, "--quote", rdFree2)
		if err != nil {
			t.Fatalf("the first mint: %v\n%s", err, out)
		}
		cycle := func(n int) []*record.Event {
			_, head, err := record.ReportAtHead(runtest.Open(t, b.runDir))
			if err != nil {
				t.Fatal(err)
			}
			k := 0
			at := func(seat string, body proto.Message) *record.Event {
				k++
				return recordtest.At(t, seat, fmt.Sprintf("%s:road:%d:%d", seat, n, k), body)
			}
			return []*record.Event{
				at("red-chair", &recordpb.Register{}),
				at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(head), SeatId: proto.String(lensM), GapIds: []string{held}}),
				at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(head), SeatId: proto.String(blueSeat), GapIds: []string{held}}),
				at(lensM, &recordpb.Register{}),
				at(blueSeat, &recordpb.Register{}),
			}
		}
		// Both parties sat and neither has sat again: the record cannot close either sitting.
		recordtest.Seed(t, b.runDir, cycle(1)...)
		_, out, err = mint(b, lensL, "--quote", rdFree2)
		roadRefusal(t, err, out, []string{roadSupersede + held, roadExchange, "NOT MEASURED"}, nil)
		// Their next sittings close the first: one exchange, and no sentence.
		recordtest.Seed(t, b.runDir, cycle(2)...)
		_, out, err = mint(b, lensL, "--quote", rdFree2)
		roadRefusal(t, err, out, []string{roadSupersede + held, roadOnBoard}, []string{roadExchange, "NOT MEASURED"})
	})

	// A mint superseding another seat's open gap is refused with the proposal road named first.
	t.Run("a supersede of another seat's open gap", func(t *testing.T) {
		b := readingRun(t)
		held, out, err := mint(b, lensM, "--quote", rdFree1)
		if err != nil {
			t.Fatalf("the first mint: %v\n%s", err, out)
		}
		_, out, err = mint(b, lensL, "--quote", rdFree1, "--supersedes", held)
		roadRefusal(t, err, out, []string{held, "assigned to " + lensM, roadSupersede + held, roadExchange, roadOnBoard}, nil)
		if i, j := strings.Index(err.Error(), roadSupersede), strings.Index(err.Error(), roadOnBoard); i > j {
			t.Errorf("the refusal names the proposal after the board, want it first:\n%v", err)
		}
	})

	// A gap is not the subject of another gap: `--about-kind gap` is refused naming both roads.
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
			"If you dispute a gap — its problem, fix, check or weight — propose the act you want on it " +
			"(`motion docket file supersede|regrade|close --id <gap>`), argued in --reason; its assignee may accept it, else the bench rules. " +
			"If you restate a gap's defect, mint superseding it (`--supersedes <gap>`) — your own open gap, or any closed gap. Nothing was written"
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
