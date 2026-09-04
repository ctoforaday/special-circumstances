package scorecard

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func rowByMetric(rows []Row, metric string) *Row {
	for i := range rows {
		if rows[i].Metric == metric {
			return &rows[i]
		}
	}
	return nil
}

func TestJsToFixed2Num(t *testing.T) {
	cases := map[float64]string{
		0.5: "0.5", 1.0: "1", 10.0: "10", 0.0: "0",
		2.0 / 3.0: "0.67", 0.666: "0.67", 100.0: "100", 1.5: "1.5",
	}
	for in, want := range cases {
		if got := jsToFixed2Num(in); got != want {
			t.Errorf("jsToFixed2Num(%v) = %q, want %q", in, got, want)
		}
	}
}

// The two object-valued rows must serialize in JS INSERTION order — a Go map would sort, which
// is the whole byte-identity trap this port had to solve.
func TestBucketFindingsByRoleKeyOrder(t *testing.T) {
	findings := []record.FindingJSON{
		{Role: "L1", Round: 1, SeatID: "a"},
		{Role: "L1", Round: 1, SeatID: "b"},
		{Role: "L5", Round: 1, SeatID: "c"},
		{Role: "L6", Round: 1, SeatID: "c"}, // darkside, 1 seat
	}
	got, ok := BucketFindingsByRole(findings)
	if !ok {
		t.Fatal("expected findings")
	}
	// citation 2 over 2 seats → 1; logic 1 over 1 → 1; darkside 1 over 1 → 1. Keys in the
	// NON-alphabetical order citation,logic,darkside,seats,per_seat.
	want := `{"1":{"citation":2,"logic":1,"darkside":1,"seats":{"citation":2,"logic":1,"darkside":1},"per_seat":{"citation":1,"logic":1,"darkside":1}}}`
	if string(got) != want {
		t.Errorf("citation_yield JSON:\n got  %s\n want %s", got, want)
	}
	// A darkside-less round → per_seat.darkside is null (not 0), matching JS.
	got2, _ := BucketFindingsByRole([]record.FindingJSON{{Role: "L1", Round: 2, SeatID: "a"}})
	if !strings.Contains(string(got2), `"per_seat":{"citation":1,"logic":null,"darkside":null}`) {
		t.Errorf("empty role must be per_seat null, got %s", got2)
	}
	if _, ok := BucketFindingsByRole(nil); ok {
		t.Error("no findings must return ok=false")
	}
}

// closedGap builds a replayed gap closed by a `close` event carrying the given anchor fields.
func closedGap(c *recordpb.Close) *record.Gap {
	return &record.Gap{HasClosed: true, Closure: c}
}

// benchGap builds a gap the BENCH disposed of: a bench body, no `close` body — the shape that
// can never carry an anchor, and the reason this kernel takes the board.
func benchGap() *record.Gap {
	return &record.Gap{HasClosed: true, BenchClosure: &recordpb.Opinion{}, ClosedByBench: true}
}

func boardOf(gaps map[string]*record.Gap, order ...string) *record.Board {
	return &record.Board{Gaps: gaps, GapOrder: order}
}

func TestComputeAnchoredClosures(t *testing.T) {
	full := &recordpb.Close{AnchorSeat: proto.String("L1"), AnchorTool: proto.String("grep"), AnchorTarget: proto.String("x")}
	partial := &recordpb.Close{AnchorSeat: proto.String("L1"), AnchorTarget: proto.String("x")}

	gaps := map[string]*record.Gap{
		"G1": closedGap(full),                                            // anchored (triple)
		"G2": closedGap(&recordpb.Close{CarriedFrom: proto.String("1")}), // anchored (carried)
		"G3": closedGap(partial),                                         // partial — NOT
		"G4": closedGap(&recordpb.Close{}),                               // NOT
		"G5": {Open: true},                                               // open — not counted at all
	}
	a, total := ComputeAnchoredClosures(boardOf(gaps, "G1", "G2", "G3", "G4", "G5"))
	if a != 2 || total != 4 {
		t.Errorf("ComputeAnchoredClosures = %d/%d, want 2/4", a, total)
	}
}

// A BENCH DISPOSITION IS NOT AN UNANCHORED REPAIR, IT IS NOT A REPAIR. It carries no
// seat|tool|target and no carried_from — nothing was re-run — so counting it in the denominator
// made `target 100` unreachable on any run where the bench closed anything.
func TestBenchDispositionsLeaveBothCounts(t *testing.T) {
	full := &recordpb.Close{AnchorSeat: proto.String("L1"), AnchorTool: proto.String("grep"), AnchorTarget: proto.String("x")}
	a, total := ComputeAnchoredClosures(boardOf(map[string]*record.Gap{
		"G1": closedGap(full),
		"G2": benchGap(),
	}, "G1", "G2"))
	if a != 1 || total != 1 {
		t.Errorf("a bench disposition is still in the ratio: got %d/%d, want 1/1", a, total)
	}
}

// THE ORDERING THAT SEPARATES THE PREDICATE FROM THE FLAG, and the whole reason this does not
// read ClosedByBench. Blue closes with a full anchor triple; the bench afterwards rules on the
// same gap. ClosedByBench is TRUE there — it follows the last closing event — while the closure
// on the board is blue's, anchored. Keyed on the flag this gap vanishes from both counts and a
// genuinely anchored repair goes unmeasured.
func TestABlueCloseTheBenchLaterRuledOnStaysCounted(t *testing.T) {
	g := closedGap(&recordpb.Close{AnchorSeat: proto.String("L1"), AnchorTool: proto.String("grep"), AnchorTarget: proto.String("x")})
	g.BenchClosure = &recordpb.Opinion{}
	g.ClosedByBench = true // the LAST closer was the bench; the closure is still blue's

	a, total := ComputeAnchoredClosures(boardOf(map[string]*record.Gap{"G1": g}, "G1"))
	if a != 1 || total != 1 {
		t.Errorf("blue's anchored close was dropped because the bench later ruled on it: got %d/%d, want 1/1", a, total)
	}
}

// THE ELSE BRANCH HAS TO BE TRUE OF BOTH EMPTY BOARDS. An all-bench board and a board with no
// closures at all now reach the same note, and "no closed gaps this run" was false of the first.
func TestTheEmptyDenominatorNoteIsTrueOfBothWaysToGetOne(t *testing.T) {
	for name, board := range map[string]*record.Board{
		"all bench":      boardOf(map[string]*record.Gap{"G1": benchGap()}, "G1"),
		"nothing closed": boardOf(map[string]*record.Gap{"G1": {Open: true}}, "G1"),
	} {
		row := rowByMetric(redRows(record.Run{}, nil, nil, board), "anchored_closures_pct")
		if row == nil {
			t.Fatalf("%s: no anchored_closures_pct row", name)
		}
		if strings.Contains(row.Note, "no closed gaps this run") {
			t.Errorf("%s: the note claims nothing closed: %q", name, row.Note)
		}
		if !strings.Contains(row.Note, "repair") {
			t.Errorf("%s: the note does not say what the denominator is: %q", name, row.Note)
		}
	}
}

func TestComputeDirectionUptake(t *testing.T) {
	dj := record.DebateJSON{Rounds: []record.DebateRoundJSON{
		{Lead: []record.DebateOpinionJSON{{GapID: "R1-1"}}, Blue: []string{"acted on the judge direction as carried"}},
		{Lead: nil, Blue: []string{"unrelated repair notes"}},
	}}
	lead, blue := ComputeDirectionUptake(dj)
	if lead != 1 || blue != 1 {
		t.Errorf("ComputeDirectionUptake = %d/%d, want 1/1", lead, blue)
	}
}

func TestUnrecordedClaimLossCountsRetireEventsNotEnvelope(t *testing.T) {
	// The additive-integrity detector: a claim_count fall the retire EVENTS don't account for is
	// substance leaving silently. Retires come from the RECORD (retire events on the board), NOT a
	// BLUE_ENVELOPE `retired` field — that field never existed, so the old code counted zero and
	// flagged every legitimate retirement as a violation.
	results := []map[string]any{
		{"claim_count": float64(10)},
		{"claim_count": float64(7)},
	}
	board := &record.Board{Events: []*record.Event{recordtest.Event(t, "", 0, &recordpb.Retire{})}} // one recorded retirement
	r := rowByMetric(blueRows(record.Run{}, results, nil, board), "unrecorded_claim_loss")
	if r == nil || r.Value == nil {
		t.Fatalf("row not computed: %+v", r)
	}
	if v, _ := r.Value.(int); v != 2 { // drop 3, 1 retire event → max(0, 3-1)=2
		t.Errorf("unrecorded_claim_loss = %v, want 2 (drop 3 − 1 retire event on the record)", r.Value)
	}
	if !strings.Contains(r.Note, "3 claim(s) lost across rounds, 1 retired") {
		t.Errorf("note = %q", r.Note)
	}

	// A `retired` field on the ENVELOPE must be IGNORED — the fix guards against the phantom-field
	// source. No board + a bogus envelope `retired` → retires stays 0 → the whole drop is flagged.
	phantom := []map[string]any{
		{"claim_count": float64(10)},
		{"claim_count": float64(7), "retired": []any{map[string]any{}, map[string]any{}}},
	}
	rp := rowByMetric(blueRows(record.Run{}, phantom, nil, nil), "unrecorded_claim_loss")
	if v, _ := rp.Value.(int); v != 3 {
		t.Errorf("a phantom envelope `retired` field must not count: want lost=3 (drop 3 − 0), got %v", rp.Value)
	}

	// A single round → the not-computed note.
	r1 := rowByMetric(blueRows(record.Run{}, []map[string]any{{"claim_count": float64(5)}}, nil, nil), "unrecorded_claim_loss")
	if r1.Value != nil {
		t.Errorf("single round must not compute: %+v", r1)
	}
}

// headline ranks a tripped detector first, drops zero detectors and uncomputed rows.
func TestHeadlineRanking(t *testing.T) {
	rows := []Row{
		{Metric: "benchy", Cls: "benchmark", Value: 1},
		{Metric: "quiet", Cls: "detector", Value: 0},
		{Metric: "live", Cls: "detector", Value: 3},
		{Metric: "blocked", Cls: "benchmark", Note: "BLOCKED"},
	}
	h := headline(rows, 3)
	if len(h) == 0 || !strings.HasPrefix(h[0], "live 3") {
		t.Errorf("a tripped detector must rank first: %v", h)
	}
	joined := strings.Join(h, " ")
	if strings.Contains(joined, "quiet") {
		t.Error("a zero detector is not headline news")
	}
	if strings.Contains(joined, "blocked") {
		t.Error("an uncomputed row is never a headline")
	}
}

// renderChair reproduces the exact markdown, including a not-computed row and an object value.
func TestRenderChairFormat(t *testing.T) {
	rows := []Row{
		{Clause: "Durable repairs", Metric: "repair_regression_ratio", Cls: "benchmark", Value: 0.5, Joint: "reads WITH red rigour"},
		{Clause: "Alternatives explored", Metric: "lines_of_inquiry", Cls: "diagnostic", Value: objJSON(`{"pursued":2,"abandoned":1}`)},
	}
	out := RenderChair("blue", rows, "this run")
	for _, want := range []string{
		"## this run",
		"- `repair_regression_ratio` [benchmark] — Durable repairs: **0.5**",
		"  - reads WITH red rigour",
		"- `lines_of_inquiry` [diagnostic] — Alternatives explored: **{\"pursued\":2,\"abandoned\":1}**",
		"HEADLINE: repair_regression_ratio 0.5 [BENCHMARK] · lines_of_inquiry {\"pursued\":2,\"abandoned\":1} [DIAGNOSTIC]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("renderChair missing %q:\n%s", want, out)
		}
	}
}
