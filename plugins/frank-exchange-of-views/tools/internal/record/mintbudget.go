package record

import (
	"fmt"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A LENS'S MINT BUDGET SCALES WITH WHAT IT AUDITS (plans/roundless.md §III.B.2.2 sets the bound;
// this sets its size). A fixed number per lens ignored the document: five mints is generous for a
// page and starves a report of thirty paragraphs, and a floor of one was one for any report at
// all. So each area names the unit its audit runs over and how many mints so many of them buy:
//
//	budget(lens) = max(floor, ceil(units × mints / per))
//
// The floor is the run's MintBudget (--mint-budget at setup), so no lens ever gets less than the
// operator granted. The units are read from the record at EACH mint, so the budget grows as the
// report grows.
//
// THE BUDGET CAPS WHAT A LENS HOLDS OPEN, NOT WHAT IT MINTS IN THE RUN. What is counted against it
// is the lens's gaps that are open on the gap view (lensMintStanding), so a gap that closes frees a
// mint and a defect found after the cap is a finding now and a gap once a slot is free. A count of
// every mint never refilled: a lens that had minted its budget and closed two of the three was
// refused the next real defect while holding one open, and a finding obliges nobody.
//
// A GAP THAT ANSWERS A CONTRADICTION ITS OWN LENS RAISED IS OUTSIDE THE BUDGET (gblock, 2026-10-20:
// the lens can always answer what it raised). A contradiction is raised by a finding and answered
// only by the gap minted from that finding (contradictionsOf); counted, a lens at its cap could
// raise one and never answer it, and the contradiction held the PASS until a slot freed — a CEILING
// made by the bound, not by the report. The link is the mint's own found_by, so no field marks the
// gap: the fold says which gap answered which raise, and whose. Every other mint counts — one
// crediting another seat's raising finding, or a finding that raised nothing.
//
// ONE TABLE. The check, the refusal and the tests all read mintScales; an area with no row is an
// error, never a default — a default would size an unknown lens by some other lens's unit.

// mintMeasure is one way of sizing the report for a lens: what it counts, and how to count it.
type mintMeasure struct {
	name string // what the refusal calls the unit, plural: "citations on the record"
	// count counts the units off the record q holds, and off its report, which report renders.
	count func(q recordsql.Querier, report func() (string, error)) (int, error)
	// readsReport is whether count reads the report: a mint measured by it renders the report before
	// its write takes the lock (validationReadsReport).
	readsReport bool
}

var (
	// Citations: blue's `cite` rows — every source ever authored on the run, retired or not. Each
	// row is one citation (its event key is the citation label), and a retired source was still
	// one the evidence lens had to verify. Counted from the record, not the rendered report, so the
	// evidence budget never depends on the renderer and cannot shrink under a lens mid-run.
	measureCitations = mintMeasure{"citations on the record", countRows(`SELECT count(*) FROM "cite"`), false}
	// Proofs: blue's `proof` rows — every computation recorded on the run. A reproduce is red's
	// act on a proof and lives in its own table.
	measureProofs = mintMeasure{"proofs on the record", countRows(`SELECT count(*) FROM "proof"`), false}
	// Claims: claimcount.Count on the CURRENT report — citation anchors attached to prose.
	measureClaims = mintMeasure{"claims in the current report", countInReport(claimcount.Count), true}
	// Paragraphs: claimcount.Paragraphs on the CURRENT report — blank-line-separated blocks that
	// carry prose; headings, anchor-only lines, footnote definitions and fenced code do not.
	measureParagraphs = mintMeasure{"prose paragraphs in the current report", countInReport(claimcount.Paragraphs), true}
)

// mintScale is one area's row: its unit, and the rate — `mints` mints for every `per` units. Two
// integers because the rates are not all whole: three mints per two citations is no "one per N".
type mintScale struct {
	measure mintMeasure
	mints   int
	per     int
}

// mintsOf names a number of mints: every figure a surface prints of the rate says what it counts.
func mintsOf(n int) string {
	if n == 1 {
		return "1 mint"
	}
	return fmt.Sprintf("%d mints", n)
}

// rate states a scale's ratio in words, for the refusal, the work list and the help.
func rate(mints, per int, unit string) string {
	return fmt.Sprintf("%s for every %d of the %s", mintsOf(mints), per, unit)
}

// mintScales is the one source of every lens area's budget scale. Every area in flags.LensAreas
// has a row, and no row names an area outside it (TestEveryLensAreaHasAMintScale).
var mintScales = map[string]mintScale{
	"evidence":     {measureCitations, 3, 2},
	"computation":  {measureProofs, 3, 2},
	"logic":        {measureClaims, 1, 1},
	"voice":        {measureParagraphs, 3, 5},
	"dark-side":    {measureParagraphs, 3, 5},
	"adversary":    {measureParagraphs, 3, 5},
	"architecture": {measureParagraphs, 3, 5},
}

// MintScaleSummary states every area's scale in roster order, for help text: read from the table,
// so no page carries a second copy of a ratio.
func MintScaleSummary() string {
	parts := make([]string, 0, len(flags.LensAreas))
	for _, a := range flags.LensAreas {
		if s, ok := mintScales[a]; ok {
			parts = append(parts, a+" "+rate(s.mints, s.per, s.measure.name))
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
	Mints   int    // mints bought by every Per units
	Per     int    // the units that buy Mints mints
	Floor   int    // the run's MintBudget
	Budget  int    // max(Floor, ceil(Units × Mints / Per))
}

// Rate states the budget's ratio in words: "3 mints for every 2 of the citations on the record".
func (b LensBudget) Rate() string { return rate(b.Mints, b.Per, b.Measure) }

// Arithmetic states how the budget was reached, for a refusal or a board.
func (b LensBudget) Arithmetic() string {
	return fmt.Sprintf("max(floor %d, ceil(%d %s × %s / %d)) = %d", b.Floor, b.Units, b.Measure, mintsOf(b.Mints), b.Per, b.Budget)
}

// budgetOf is the ruling's arithmetic: max(floor, ceil(units × mints / per)).
func budgetOf(floor, units, mints, per int) int {
	b := (units*mints + per - 1) / per
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
	db, err := openRunForRead(run)
	if err != nil {
		return LensBudget{}, err
	}
	return lensBudgetAt(db, AreaOf(seatID), p.MintBudget)
}

// lensBudgetAt is the area's budget over the record q holds, its report rendered from q.
func lensBudgetAt(q recordsql.Querier, area string, floor int) (LensBudget, error) {
	return lensBudgetOver(q, renderFrom(q), area, floor)
}

// lensBudgetOver is the area's budget over the record q holds — the mint's writing transaction, or
// the run's handle — and the report it renders, which report returns: the write's landing report
// (writeView.report) or a render of q.
func lensBudgetOver(q recordsql.Querier, report func() (string, error), area string, floor int) (LensBudget, error) {
	s, err := mintScaleOf(area)
	if err != nil {
		return LensBudget{}, err
	}
	b := LensBudget{Area: area, Measure: s.measure.name, Mints: s.mints, Per: s.per, Floor: floor}
	if b.Units, err = s.measure.count(q, report); err != nil {
		return b, fmt.Errorf("counting %s: %w", s.measure.name, err)
	}
	b.Budget = budgetOf(floor, b.Units, s.mints, s.per)
	return b, nil
}

// renderFrom is the report q holds, rendered when asked.
func renderFrom(q recordsql.Querier) func() (string, error) {
	return func() (string, error) { return RenderAt(q) }
}

// validationReadsReport is whether the write path's validation of body, written by seatID, reads
// the report: a mint by a lens whose budget is measured on it (requireMintWithinBudget). The
// write renders the report before it takes the lock (appendOwn); a mint whose seat validation turns
// away before the measure renders for nothing.
func validationReadsReport(seatID string, body proto.Message) bool {
	if _, ok := body.(*recordpb.Mint); !ok {
		return false
	}
	s, ok := mintScales[AreaOf(seatID)]
	return ok && s.measure.readsReport
}

func countRows(query string) func(recordsql.Querier, func() (string, error)) (int, error) {
	return func(q recordsql.Querier, _ func() (string, error)) (int, error) {
		var n int
		_, err := queryRowAt(q, []any{&n}, query)
		return n, err
	}
}

func countInReport(count func(string) int) func(recordsql.Querier, func() (string, error)) (int, error) {
	return func(_ recordsql.Querier, report func() (string, error)) (int, error) {
		md, err := report()
		if err != nil {
			return 0, err
		}
		return count(md), nil
	}
}

var errNoRenderer = fmt.Errorf("this binary registers no report renderer (internal/reportproj registers one wherever it is linked)")

// reportRenderer replays the report's frozen base and its ordered mutations — the rows
// ReportProjectionAt reads — into the current report. It is REGISTERED, not imported: the renderer
// (internal/reportproj) imports this package, so the write path is handed it instead. It asks the
// record nothing, so a reader holding a read transaction takes the rows on it and replays after the
// transaction closes.
//
// IT IS A LEFT FOLD OF THE OPS OVER THE BASE: replaying B over the text A replayed to is replaying A
// then B over the base. The write path folds the ops another write added onto a render it made before
// its lock (landingReport.fold) and reads the result as the report; TestTheReportReplayIsAFold
// (internal/record/migrate) holds the renderer to it on archived runs.
var reportRenderer func(base string, haveBase bool, ops []ReportOp) (string, error)

// RegisterReportRenderer hands the write path the renderer the report-sized budgets read.
func RegisterReportRenderer(fn func(base string, haveBase bool, ops []ReportOp) (string, error)) {
	reportRenderer = fn
}

// renderProjection replays a report projection — ReportProjection's or ReportProjectionAt's rows —
// into the current report, or says why it cannot: the rows did not read, or no renderer is linked.
func renderProjection(base string, haveBase bool, ops []ReportOp, err error) (string, error) {
	if err == nil && reportRenderer == nil {
		err = errNoRenderer
	}
	if err != nil {
		return "", err
	}
	return reportRenderer(base, haveBase, ops)
}

// requireMintWithinBudget is the run-level bound that is not a clock: each cast lens may hold at
// most its budget (above) of gaps open at once. A SUPERSEDING mint is a new open gap and takes a
// mint like any other; the gap it replaces holds its own until it closes, which superseding does
// not do — its lens closes it naming the successor. A lens at its budget still sits while the
// record readies it: it verifies, records findings, regrades and closes its own gaps, and each
// close frees a mint; it does not retire while it holds them (lensOwesOf). The
// count is the record's own — the gap view's open gaps this seat minted — not a counter
// the seat carries. Only a lens is bounded: a seat of another role spends no budget. A mint
// attributed to an id that is no seat is refused.
//
// BELOW THE FLOOR THE MEASURE DECIDES NOTHING. The budget is at least the floor, so a mint under it
// lands whatever the report holds — and whether or not the units could be read.
//
// THE STANDING IS lensMintStanding'S, the reading the work list states (WorkOfSeat), so the count a
// lens is shown and the count it is refused on are one read by one function.
func requireMintWithinBudget(w writeView, seatID string, m *recordpb.Mint) error {
	// THE ROLE AND THE AREA ARE THE ROSTER'S, SO AN ID OFF THE ROSTER HAS NEITHER. A seat of another
	// role is not a lens spending a budget; an id with no role is no seat at all, and is refused by
	// name rather than by an area cut off its spelling — `red-lens-banana` is not the `banana` lens.
	// requireDispatchableSeat refuses such an id at `register`; a write that reaches here under one
	// is refused here.
	role := seatclass.RoleOfSeat(seatID)
	if role != "" && role != "lens" {
		return nil
	}
	area := AreaOf(seatID)
	if area == "" {
		return feov.Errorf(feov.Validation,
			"%q is not a lens seat on the engine's roster, so there is no area to size a mint budget for",
			seatID)
	}
	if _, err := mintScaleOf(area); err != nil {
		return feov.Errorf(feov.Validation, "%s: %v", seatID, err)
	}
	// THE STANDING IS READ IN THE MINT'S OWN TRANSACTION. Read before it, two mints from one lens —
	// parallel calls in one sitting — both find the last slot free and both land, and the lens holds
	// one gap more than its budget admits. The report it is measured on is the write's landing
	// report, rendered before the lock where nothing has moved it since.
	// AN ANSWER TO ITS OWN RAISE IS ADMITTED AT ANY STANDING, before the standing is read: it takes
	// no mint, so neither the count nor the measure decides it.
	if own, err := answersOwnRaiseAt(w.q, seatID, m.GetFoundBy()); err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's contradictions could not be read, so whether this mint answers one %s raised cannot be told: %v", seatID, err)
	} else if own != "" {
		return nil
	}
	if own, err := answersOwnUnsoundAt(w.q, seatID, m); err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's proof judgements could not be read, so whether this mint answers one %s made cannot be told: %v", seatID, err)
	} else if own != "" {
		return nil
	}
	st, err := lensMintStanding(w.q, w.report, w.files.params, seatID, area)
	if err != nil {
		return err
	}
	n, b := st.Open, st.Budget
	if n < b.Floor {
		return nil
	}
	if st.MeasureErr != nil {
		return feov.Errorf(feov.Validation,
			"%s holds %d gap(s) open, the floor of its mint budget (mintBudget in inputs/run-config.json is %d), "+
				"and the %s lens's budget above the floor, %s, cannot be read: %v. "+
				"A mint past the floor waits on that measure rather than guessing it, or on a gap you minted closing, which frees a mint under the floor. "+
				"%s", seatID, n, b.Floor, area, b.Rate(), st.MeasureErr, mintRefusedNext)
	}
	if n < b.Budget {
		return nil
	}
	return feov.Errorf(feov.Validation,
		"%s holds %d gap(s) open and its mint budget admits no more. The budget caps the gaps a lens holds open at once: "+
			"the %s lens's is the larger of the run's floor (mintBudget in inputs/run-config.json) and its rate, %s: %s, "+
			"and it grows as the report does. Every open gap you minted takes one mint of it, a trifle as much as a material defect. "+
			"A gap of yours that closes, by your close or by the bench's ruling, frees one. %s",
		seatID, n, area, b.Rate(), b.Arithmetic(), mintRefusedNext)
}

// mintRefusedNext is what a lens refused for budget does with what it found, on both refusals: the
// defect is recorded as a finding, and the finding is what the later mint credits.
const mintRefusedNext = "Record what you found as a finding now, and mint it when a gap of yours has closed, naming that finding in --found-by. " +
	"You can still verify, record findings, and regrade or close the gaps you minted. " +
	mintAnswerOutsideBudget

// mintAnswerOutsideBudget is the exemption as every lens-facing surface states it: the refusal, the
// work list, the PASS blocker and the help say one sentence.
const mintAnswerOutsideBudget = "A gap minted from a finding of yours that raised a contradiction, or about a proof whose sentence your standing judgement calls unsound, answers it and takes no mint — it is admitted at the cap, and does not count while it is open"

// answersOwnRaiseAt is the contradiction a mint by seatID crediting foundBy would answer that seatID's
// own finding raised — its id, or "" where it answers none. It is the fold's own question asked of
// the record before the mint (contradictionsOf), so the mint it admits is the gap the count then
// leaves out (ownAnswersAt). A raise another seat recorded, or one already answered, is none.
func answersOwnRaiseAt(q recordsql.Querier, seatID string, foundBy []string) (string, error) {
	if len(foundBy) == 0 {
		return "", nil
	}
	evs, win, err := eventsOfAt(q, recordpb.EventType_EVENT_TYPE_VERIFY, recordpb.EventType_EVENT_TYPE_FINDING, recordpb.EventType_EVENT_TYPE_MINT)
	if err != nil {
		return "", err
	}
	for _, c := range contradictionsOf(evs, win.IDs(evs)) {
		if c.raiser == seatID && c.answeredBy == "" && slices.Contains(foundBy, c.raisedBy) {
			return c.idWord(), nil
		}
	}
	return "", nil
}

// ownAnswersAt is the gaps that answer a contradiction seatID's own finding raised, or a standing
// unsound judgement of seatID's own (ownUnsoundAnswers). Those seatID
// minted are the gaps its budget does not count (lensMintStanding); one another seat minted from
// that finding answers it all the same, and counts against its own minter's budget.
func ownAnswersAt(q recordsql.Querier, seatID string) (map[string]bool, error) {
	evs, win, err := eventsOfAt(q, recordpb.EventType_EVENT_TYPE_VERIFY, recordpb.EventType_EVENT_TYPE_FINDING, recordpb.EventType_EVENT_TYPE_MINT)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, c := range contradictionsOf(evs, win.IDs(evs)) {
		if c.answeredBy != "" && c.raiser == seatID {
			out[c.answeredBy] = true
		}
	}
	// THE PROOF'S FOLD READS THE WHOLE LEDGER — the checks, the edits and the retires — so it takes
	// the stream the contradictions' narrower read does not hold, and only where seatID has minted a
	// gap about a proof: every mint and every work list asks this.
	if about, err := recordHasAt(q, `SELECT 1 FROM "mint" m JOIN "events" e ON e."id" = m."event_id" WHERE m."about_kind" = ? AND e."seat_id" = ? LIMIT 1`,
		recordpb.Word(recordpb.AboutKind_ABOUT_KIND_PROOF), seatID); err != nil || !about {
		return out, err
	}
	all, allWin, err := eventsAt(q)
	if err != nil {
		return nil, err
	}
	for g := range ownUnsoundAnswers(all, allWin, seatID) {
		out[g] = true
	}
	return out, nil
}

// answersOwnUnsoundAt is the proof a mint by seatID answers — its own standing unsound judgement of
// the proof the mint is about, which no gap answers yet — or "" where it answers none.
func answersOwnUnsoundAt(q recordsql.Querier, seatID string, m *recordpb.Mint) (string, error) {
	if m.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_PROOF {
		return "", nil
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return "", err
	}
	return answersOwnUnsound(evs, win, seatID, m.GetAboutKind(), m.GetAboutRef()), nil
}

// mintStanding is where a lens stands against its mint budget: the gaps it minted that are open
// now, and its budget with every term. MeasureErr is why the allowance above the floor could not be
// read; Budget then holds the area, the unit, the ratio and the floor, and no units or total.
type mintStanding struct {
	Open int
	// Answering is the open gaps this seat minted that answer a contradiction it raised or a proof it
	// judged unsound: open, and outside Open.
	Answering  int
	Budget     LensBudget
	MeasureErr error
}

// Left is how many more mints the write path admits now. Where the measure could not be read it is
// what the floor still admits, and measured is false: a mint past the floor waits on the measure.
func (s mintStanding) Left() (left int, measured bool) {
	limit := s.Budget.Budget
	if s.MeasureErr != nil {
		limit = s.Budget.Floor
	}
	if s.Open >= limit {
		return 0, s.MeasureErr == nil
	}
	return limit - s.Open, s.MeasureErr == nil
}

// lensMintStanding reads a lens's standing off the record: the floor from the run's terms, the
// count from the gap view — the record's, not a counter the seat carries — and the budget from the
// area's scale. IT IS THE ONE READING OF "HOW MANY MINTS HAS THIS LENS LEFT": the write refuses on
// it (requireMintWithinBudget) and the work list states it (WorkOfSeat). An error is the terms or
// the count failing to read; a measure failing to read is MeasureErr, because under the floor a
// mint is admitted without it.
//
// OPEN IS THE GAP VIEW'S "open" (recordsql/views.go), the column the board, the PASS gate and the
// dispatch plan read: no closing act of either arm is on the gap. So a gap its lens closed, on
// any disposition, or the bench disposed of, holds no mint; a non-material gap holds one, because it is open; and a SUPERSEDED gap holds one until its
// lens closes it — superseding closes nothing, and the view calls such a gap stranded.
//
// q is the mint's writing transaction, or the run's handle for the work list; report is the report
// it holds (lensBudgetOver); params reads the run's terms.
func lensMintStanding(q recordsql.Querier, report func() (string, error), params func() (Params, error), seatID, area string) (mintStanding, error) {
	p, err := params()
	if err != nil {
		return mintStanding{}, err
	}
	var st mintStanding
	open, err := openGapsMintedBy(q, seatID)
	if err == nil {
		var own map[string]bool
		if own, err = ownAnswersAt(q, seatID); err == nil {
			for _, id := range open {
				if own[id] {
					st.Answering++
				} else {
					st.Open++
				}
			}
		}
	}
	if err != nil {
		// CODED AND SAID AS THE MINT'S REFUSAL, as the duplicate screen says its own failed read.
		return mintStanding{}, feov.Wrap(feov.Conflict, err, "the board could not be read, so the gaps %s holds open cannot be counted against its mint budget: %v", seatID, err)
	}
	st.Budget, st.MeasureErr = lensBudgetOver(q, report, area, p.MintBudget)
	return st, nil
}

// openGapsMintedBy is the ids of the gaps seatID minted that are open on the gap view.
func openGapsMintedBy(q recordsql.Querier, seatID string) ([]string, error) {
	if noRecord(q) {
		return nil, nil
	}
	rows, err := q.Query(`SELECT "gap_id" FROM "gap" WHERE "open" AND "minted_by" = ?`, seatID)
	if err != nil {
		return nil, fmt.Errorf("asking the record: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// MintBudgetJSON is a lens's standing against its mint budget, as its work list carries it
// (SittingJSON.MintBudget): the terms of LensBudget's arithmetic, what the seat holds open, and
// what is left.
type MintBudgetJSON struct {
	// Open is the gaps this seat minted that are open now (the gap view's "open"), less Answering.
	Open int `json:"open"`
	// Answering is the open gaps this seat minted that answer a contradiction its own finding
	// raised or a proof it judged unsound: they take no mint, so Open leaves them out and Left does
	// not spend on them.
	Answering int `json:"answering"`
	// Budget is max(floor, ceil(units × mints / per)) — or, where Measured is false, the floor alone.
	Budget int `json:"budget"`
	// Left is how many more mints the write path admits now: Budget less Open, never below 0.
	Left  int `json:"left"`
	Floor int `json:"floor"`
	// Unit names what the area's budget is sized by, Units is how many of them the record holds
	// now, and every Per of them buy Mints mints.
	Unit  string `json:"unit"`
	Units int    `json:"units"`
	Mints int    `json:"mints"`
	Per   int    `json:"per"`
	// Measured is false where the units could not be read: Budget and Left are then the floor's,
	// and a mint past the floor is refused until the measure reads. Never omitted — a budget of
	// the floor that was measured and one that was not are different answers.
	Measured bool `json:"measured"`
}

// mintBudgetOf is the work list's two carriers of a lens's standing — the block and its sentence —
// from the one reading the write refuses on. A seat that is no lens on the roster spends no budget
// and has neither. A standing that cannot be read at all is said in the sentence, with no block:
// a block of zeros would read as a lens with nothing left.
func mintBudgetOf(run Run, seatID string) (*MintBudgetJSON, string) {
	area := AreaOf(seatID)
	if _, err := mintScaleOf(area); err != nil {
		return nil, ""
	}
	db, err := openRunForRead(run)
	var st mintStanding
	if err == nil {
		st, err = lensMintStanding(db, renderFrom(db), filesOf(run).params, seatID, area)
	}
	if err != nil {
		return nil, fmt.Sprintf("your mint budget is NOT MEASURED — %v — and that is not a report that no mint is left to you", err)
	}
	b := st.Budget
	left, measured := st.Left()
	j := &MintBudgetJSON{Open: st.Open, Answering: st.Answering, Budget: b.Budget, Left: left, Floor: b.Floor, Unit: b.Measure, Units: b.Units, Mints: b.Mints, Per: b.Per, Measured: measured}
	leftWords := "NO mint is left to you"
	if left > 0 {
		leftWords = fmt.Sprintf("%d mint(s) are left to you", left)
	}
	if !measured {
		j.Budget = b.Floor
		return j, fmt.Sprintf("you hold %d gap(s) open and your mint budget above the run's floor of %d is NOT MEASURED — %s, which cannot be read: %v — so %s under the floor, and a mint past it waits on that measure or on a gap you minted closing; what you find until then goes in a finding, and the finding is minted then. %s",
			st.Open, b.Floor, b.Rate(), st.MeasureErr, leftWords, mintAnswerOutsideBudget)
	}
	tail := "every open gap you minted takes one, a superseding mint too, a gap of yours that closes frees one, and the budget grows as the report does"
	if left == 0 {
		tail = "what you find now goes in a finding, and when a gap of yours closes it frees a mint and the finding is minted then; the budget grows as the report does"
	}
	answering := ""
	if st.Answering > 0 {
		answering = fmt.Sprintf(", and %d more that answer a contradiction you raised or a proof you judged unsound", st.Answering)
	}
	return j, fmt.Sprintf("you hold %d gap(s) open of a mint budget of %d%s — %s — so %s; %s. %s",
		st.Open, b.Budget, answering, b.Arithmetic(), leftWords, tail, mintAnswerOutsideBudget)
}
