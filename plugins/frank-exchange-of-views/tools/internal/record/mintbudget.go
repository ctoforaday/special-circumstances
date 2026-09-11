package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// A LENS'S MINT BUDGET SCALES WITH WHAT IT AUDITS (plans/roundless.md §III.B.2.2 sets the bound;
// this sets its size). A fixed number per lens ignored the document: five mints is generous for a
// page and starves a report of thirty paragraphs, and a smoke run's one mint was one for any
// report at all. So each area names the unit its audit runs over and how many of them buy a mint:
//
//	budget(lens) = max(floor, ceil(units / per))
//
// The floor is the run's MintBudget (--mint-budget at setup), so no lens ever gets less than the
// operator granted. The units are read from the record at EACH mint, so the budget grows as the
// report grows; a superseding mint still counts against it, because lineage is a new gap.
//
// ONE TABLE. The check, the refusal and the tests all read mintScales; an area with no row is an
// error, never a default — a default would size an unknown lens by some other lens's unit.

// mintMeasure is one way of sizing the report for a lens: what it counts, and how to count it.
type mintMeasure struct {
	name  string // what the refusal calls the unit, plural: "citations on the record"
	count func(Run) (int, error)
}

var (
	// Citations: blue's `cite` rows — every source ever authored on the run, retired or not. Each
	// row is one citation (its event key is the citation label), and a retired source was still
	// one the evidence lens had to verify. Counted from the record, not the rendered report, so the
	// evidence budget never depends on the renderer and cannot shrink under a lens mid-run.
	measureCitations = mintMeasure{"citations on the record", countRows(`SELECT count(*) FROM "cite"`)}
	// Proofs: blue's `proof` rows — every computation recorded on the run. A reproduce is red's
	// act on a proof and lives in its own table.
	measureProofs = mintMeasure{"proofs on the record", countRows(`SELECT count(*) FROM "proof"`)}
	// Claims: claimcount.Count on the CURRENT report — citation anchors attached to prose.
	measureClaims = mintMeasure{"claims in the current report", countInReport(claimcount.Count)}
	// Paragraphs: claimcount.Paragraphs on the CURRENT report — blank-line-separated blocks that
	// carry prose; headings, anchor-only lines, footnote definitions and fenced code do not.
	measureParagraphs = mintMeasure{"prose paragraphs in the current report", countInReport(claimcount.Paragraphs)}
)

// mintScale is one area's row: its unit, and how many units buy one mint.
type mintScale struct {
	measure mintMeasure
	per     int
}

// mintScales is the one source of every lens area's budget scale. Every area in flags.LensAreas
// has a row, and no row names an area outside it (TestEveryLensAreaHasAMintScale).
var mintScales = map[string]mintScale{
	"evidence":     {measureCitations, 2},
	"computation":  {measureProofs, 2},
	"logic":        {measureClaims, 3},
	"voice":        {measureParagraphs, 5},
	"dark-side":    {measureParagraphs, 5},
	"adversary":    {measureParagraphs, 5},
	"architecture": {measureParagraphs, 5},
}

// MintScaleSummary states every area's scale in roster order, for help text: read from the table,
// so no page carries a second copy of a ratio.
func MintScaleSummary() string {
	parts := make([]string, 0, len(flags.LensAreas))
	for _, a := range flags.LensAreas {
		if s, ok := mintScales[a]; ok {
			parts = append(parts, fmt.Sprintf("%s 1 per %d %s", a, s.per, s.measure.name))
		}
	}
	return strings.Join(parts, "; ")
}

func mintScaleOf(area string) (mintScale, error) {
	s, ok := mintScales[area]
	if !ok {
		return mintScale{}, fmt.Errorf("lens area %q has no mint-budget scale — only a lens area in the engine's roster can be sized", area)
	}
	return s, nil
}

// LensBudget is one lens's budget, with every term of its arithmetic.
type LensBudget struct {
	Area    string // the lens area: "evidence"
	Measure string // the unit's name: "citations on the record"
	Units   int    // the unit count at the time of reading
	Per     int    // units per mint
	Floor   int    // the run's MintBudget
	Budget  int    // max(Floor, ceil(Units / Per))
}

// Arithmetic states how the budget was reached, for a refusal or a board.
func (b LensBudget) Arithmetic() string {
	return fmt.Sprintf("max(floor %d, ceil(%d %s / %d)) = %d", b.Floor, b.Units, b.Measure, b.Per, b.Budget)
}

// budgetOf is the ruling's arithmetic: max(floor, ceil(units / per)).
func budgetOf(floor, units, per int) int {
	b := (units + per - 1) / per
	if b < floor {
		return floor
	}
	return b
}

// LensMintBudget reads a lens's budget off the record: its area from the seat id, the floor from
// the run's terms, the units from the record. A measure that cannot be read is an error, never 0 —
// a zero would read exactly like a report too small to lift the budget off its floor.
func LensMintBudget(run Run, seatID string) (LensBudget, error) {
	p, err := RunParams(run)
	if err != nil {
		return LensBudget{}, err
	}
	return lensBudgetAt(run, RoleOf(seatID), p.MintBudget)
}

func lensBudgetAt(run Run, area string, floor int) (LensBudget, error) {
	s, err := mintScaleOf(area)
	if err != nil {
		return LensBudget{}, err
	}
	b := LensBudget{Area: area, Measure: s.measure.name, Per: s.per, Floor: floor}
	if b.Units, err = s.measure.count(run); err != nil {
		return b, fmt.Errorf("counting %s: %w", s.measure.name, err)
	}
	b.Budget = budgetOf(floor, b.Units, s.per)
	return b, nil
}

func countRows(query string) func(Run) (int, error) {
	return func(run Run) (int, error) {
		var n int
		_, err := queryRow(run, []any{&n}, query)
		return n, err
	}
}

func countInReport(count func(string) int) func(Run) (int, error) {
	return func(run Run) (int, error) {
		if reportRenderer == nil {
			return 0, fmt.Errorf("this binary registers no report renderer (internal/reportproj registers one wherever it is linked)")
		}
		md, err := reportRenderer(run)
		if err != nil {
			return 0, err
		}
		return count(md), nil
	}
}

// reportRenderer renders the current report from the record. It is REGISTERED, not imported: the
// renderer (internal/reportproj) imports this package, so the write path is handed it instead.
var reportRenderer func(Run) (string, error)

// RegisterReportRenderer hands the write path the renderer the report-sized budgets read.
func RegisterReportRenderer(fn func(Run) (string, error)) { reportRenderer = fn }

// requireMintWithinBudget is the run-level bound that is not a clock: each cast lens may mint at
// most its budget (above) in the run, a SUPERSEDING mint included — lineage is a new gap. A lens
// whose budget is spent still sits when the head moves: it verifies, records findings, regrades
// and closes its own gaps. The count is the record's own — this seat's mint events — not a counter
// the seat carries. Only a lens is bounded: the chair mints nothing, and a seed or a migration
// writing under another seat is not a lens spending a budget.
//
// BELOW THE FLOOR NOTHING IS MEASURED. The budget is at least the floor, so a mint under it lands
// whatever the report holds; the units are read only when they can decide the answer.
func requireMintWithinBudget(run Run, seatID string) error {
	if roleOfSeat(seatID) != "lens" {
		return nil
	}
	area := RoleOf(seatID)
	if _, err := mintScaleOf(area); err != nil {
		return feov.Errorf(feov.Validation, "record: mint refused — %s: %v", seatID, err)
	}
	p, err := RunParams(run)
	if err != nil {
		return err
	}
	var n int
	if _, err := queryRow(run, []any{&n},
		`SELECT count(*) FROM "mint" m JOIN "events" e ON e."id" = m."event_id" WHERE e."seat_id" = ?`, seatID); err != nil {
		return err
	}
	if n < p.MintBudget {
		return nil
	}
	b, err := lensBudgetAt(run, area, p.MintBudget)
	if err != nil {
		return feov.Errorf(feov.Validation,
			"record: mint refused — %s has minted %d gap(s), the floor of its budget (mintBudget in inputs/run-config.json is %d), "+
				"and the %s lens's budget above the floor, one mint per %d %s, cannot be read: %v. "+
				"A mint past the floor waits on that measure rather than guessing it. "+
				"You can still verify, record findings, and regrade or close the gaps you minted; what you found now goes in a finding, not a gap",
			seatID, n, p.MintBudget, area, b.Per, b.Measure, err)
	}
	if n < b.Budget {
		return nil
	}
	return feov.Errorf(feov.Validation,
		"record: mint refused — %s has minted %d gap(s), its whole budget. The %s lens's budget is the larger of the run's floor "+
			"(mintBudget in inputs/run-config.json) and one mint per %d %s: %s, and it grows as the report does. "+
			"The budget is where a trifle's cost lands: with %d mints for this report, none of them is spent on a nitpick. "+
			"You can still verify, record findings, and regrade or close the gaps you minted; what you found now goes in a finding, not a gap",
		seatID, n, area, b.Per, b.Measure, b.Arithmetic(), b.Budget)
}
