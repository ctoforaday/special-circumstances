package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH'S ORDERS ON THE RECORD (gblock, 2026-10-30): the bench sets a grade with its own regrade
// while a docket motion on the gap stands before it, the minting lens may not move an axis the bench
// holds locked, the bench moving that axis again releases it to the lens ("Lock returns to lens"),
// and a docket motion on a gap the bench ruled closed is admitted on new grounds only.
//
// The drive reaches each: the bench regrades a docketed gap before it rules it; the minting lens of a
// gap the bench remanded reaches to move the bench's axis and is refused; it moves an axis the bench
// moved again at a later hearing and is admitted; blue files a docket motion on a gap the bench ruled
// closed, first with no grounds (refused) and then on a computation it records since the ruling
// (admitted). A drawn run takes each arm where its path offers one. The sweep gates each count at zero.

// benchOrdersSeed is the seed whose run drives every arm above BY CONSTRUCTION (benchArc), on two gaps
// it mints for the purpose. It is none of the forced seeds before it: their runs are each shaped for
// one fact, and the bench's regrade reshapes the grades every one of them reads.
//
// LEFT TO THE RUN'S PATH, the release was reached on whatever gap the bench happened to hear twice
// with the axis still held, and the seed's own acts decided which: one unrelated change to the drive
// moved the seed off it, and a sweep of 40 counted it zero (PR #1663). The construction owns its gaps,
// so what the rest of the run draws does not reach them.
const benchOrdersSeed = 9

// benchOrdersPct is how often a drawn run takes each arm where it can.
const benchOrdersPct = 20

// benchOrdersArm is whether this run takes a bench-orders arm now: always on the forced seed, never on
// a run forced to a verdict, and by the draw otherwise.
func (r *runner) benchOrdersArm() bool {
	if r.forcedVerdict() {
		return false
	}
	return r.forceBenchOrders || r.coin(benchOrdersPct)
}

// benchSetsGrade is the bench, hearing gapID, moving one axis before it rules: admitted while a docket
// motion on the gap stands unruled. The move locks an axis the bench does not hold and releases one
// it does (benchMoves). A gap benchArc holds is the construction's to move.
func (r *runner) benchSetsGrade(seatID, gapID string) {
	if r.arcHolds(gapID) || !slices.Contains(r.openGaps(), gapID) || len(r.unruledDocketMotions(gapID)) == 0 || !r.benchOrdersArm() {
		return
	}
	// EVERY AXIS, complexity included: the bench's regrade takes the lens's four flags.
	dim := pick(r.rng, benchRegradeDims)
	r.benchMoves(seatID, gapID, dim, r.gradeOtherThan(r.currentGrade(gapID, dim)))
}

// benchMoves is the bench's regrade moving gapID's dim to grade, and the drive's model of the lock
// after it: an axis the bench held is released to the lens (benchLifted), any other is the bench's
// (benchSet).
func (r *runner) benchMoves(seatID, gapID, dim, grade string) bool {
	if _, ok := r.admitFresh(gapID, "regrade", "--seat-id", seatID, "--id", gapID, "--"+dim, grade,
		"--reason", "fuzz: the bench holds "+gapID+"'s "+dim+" is "+grade+", and rules on that grade"); !ok {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.benchSet == nil {
		r.benchSet, r.benchLifted = map[string]map[string]string{}, map[string]map[string]bool{}
	}
	if r.benchSet[gapID] == nil {
		r.benchSet[gapID] = map[string]string{}
	}
	if r.benchLifted[gapID] == nil {
		r.benchLifted[gapID] = map[string]bool{}
	}
	if _, locked := r.benchSet[gapID][dim]; locked {
		delete(r.benchSet[gapID], dim)
		r.benchLifted[gapID][dim] = true
	} else {
		r.benchSet[gapID][dim] = grade
		delete(r.benchLifted[gapID], dim)
	}
	r.benchRegrades++
	return true
}

// benchRegradeDims is the four axes by flag name, as a regrade spells them.
var benchRegradeDims = append(append([]string{}, regradeDims...), "complexity")

// benchGrades is a copy of the axes the bench set on gapID, by flag name.
func (r *runner) benchGrades(gapID string) map[string]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]string{}
	for dim, g := range r.benchSet[gapID] {
		out[dim] = g
	}
	return out
}

// inheritBenchGrades is a successor minted superseding ancestor carrying the bench's axes: the lock
// follows the lineage, so the successor's are the bench's too.
func (r *runner) inheritBenchGrades(successor, ancestor string) {
	held := r.benchGrades(ancestor)
	if len(held) == 0 || successor == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.benchSet[successor] = held
}

// keepingBenchGrades is args with every axis the bench set on any gap it supersedes stated at the
// bench's grade, so a successor mint the drive expects to land does not meet the lock. A flag given
// twice takes the last.
func (r *runner) keepingBenchGrades(args []string) []string {
	for i, a := range args {
		if a != "--supersedes" || i+1 >= len(args) {
			continue
		}
		for dim, g := range r.benchGrades(args[i+1]) {
			args = append(args, "--"+dim, g)
		}
	}
	return args
}

// lensMovesBenchAxis is the minting lens reaching to move an axis the bench set on an open gap of its
// own, which the record refuses (lockRefusesLens). A gap benchArc holds is the construction's.
func (r *runner) lensMovesBenchAxis(seatID string) {
	for _, id := range r.openGaps() {
		held := r.benchGrades(id)
		if len(held) == 0 || r.arcHolds(id) || r.minterOf(id) != seatID || !r.benchOrdersArm() {
			continue
		}
		dims := make([]string, 0, len(held))
		for dim := range held {
			dims = append(dims, dim)
		}
		slices.Sort(dims)
		dim := pick(r.rng, dims)
		r.lockRefusesLens(seatID, id, dim, r.gradeOtherThan(held[dim]))
		return
	}
}

// lockRefusesLens is seatID's regrade moving gapID's dim to grade, an axis the bench holds locked, held
// to the record's refusal as a conflict naming the bench's grade and its place; lockRefused counts each.
// It reports whether the lock refused it.
func (r *runner) lockRefusesLens(seatID, gapID, dim, grade string) bool {
	args := []string{"--json", "regrade", "--seat-id", seatID, "--id", gapID, "--" + dim, grade,
		"--reason", "fuzz: the lens moving an axis the bench set"}
	for range 3 {
		before := r.staleReads(seatID, gapID)
		out, err := r.exec(args...)
		env, ok := refusalEnvelopeIn(out)
		switch {
		case err == nil:
			r.noteActErr(fmt.Sprintf("%s moved %s's %s, which the bench set, to %s, and the record admitted it: %s", seatID, gapID, dim, grade, firstLine(out)))
		case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, "which the bench set at event") && strings.Contains(env.Error, "Nothing was recorded"):
			r.mu.Lock()
			r.lockRefused++
			r.mu.Unlock()
			return true
		case ok && env.Code == string(feov.Conflict) && r.staleReads(seatID, gapID) > before:
			noteEither(args[1:2], "a regrade of a gap another seat changed since this seat's read point, recorded again", false)
			continue
		default:
			r.noteActErr(fmt.Sprintf("%s's move of %s's bench-set %s was refused, and not by the bench's lock: %s", seatID, gapID, dim, firstLine(out)))
		}
		return false
	}
	return false
}

// lensMovesReleasedAxis is the minting lens moving an axis the bench moved again on an open gap of its
// own: the bench's later move released it, so the record admits the lens's move, and lockLifted
// counts each one admitted. The axis is then the lens's, as any unlocked axis is. It returns the gap
// it regraded, "" for none, and leaves skip alone (this sitting already regraded it) and every gap
// benchArc holds.
func (r *runner) lensMovesReleasedAxis(seatID, skip string) string {
	for _, id := range r.openGaps() {
		if id == skip || r.arcHolds(id) {
			continue
		}
		r.mu.Lock()
		var dims []string
		for dim := range r.benchLifted[id] {
			dims = append(dims, dim)
		}
		r.mu.Unlock()
		if len(dims) == 0 || r.minterOf(id) != seatID || !r.benchOrdersArm() {
			continue
		}
		slices.Sort(dims)
		dim := pick(r.rng, dims)
		if _, ok := r.admitFresh(id, "regrade", "--seat-id", seatID, "--id", id, "--"+dim, r.gradeOtherThan(r.currentGrade(id, dim)),
			"--reason", "fuzz: the lens moving an axis the bench released"); !ok {
			return ""
		}
		r.mu.Lock()
		delete(r.benchLifted[id], dim)
		r.lockLifted++
		r.mu.Unlock()
		return id
	}
	return ""
}

// reconsider is blue putting a gap the bench ruled closed before it again: with no grounds, or on a
// quote of report text no edit answering the gap wrote since the ruling, refused; then on a computation of
// its own answering the gap, recorded since the ruling, admitted. A gap a docket motion already
// stands on is left to that motion.
func (r *runner) reconsider(seatID string) {
	if !r.benchOrdersArm() {
		return
	}
	w, err := record.WorkOfSeat(r.run(), "blue", seatID)
	if err != nil {
		r.noteActErr("blue's work list could not be read to reconsider a gap the bench ruled closed: " + err.Error())
		return
	}
	var ruled []string
	for _, e := range w.Estopped {
		if len(r.unruledDocketMotions(e.ID)) == 0 {
			ruled = append(ruled, e.ID)
		}
	}
	// benchOrdersSeed's construction reconsiders the gap it had the bench rule closed.
	id, arc := r.arcAwaitsReconsideration(), false
	switch {
	case id != "" && !slices.Contains(ruled, id):
		r.noteActErr(fmt.Sprintf("the bench ruled %s closed for the bench-orders construction, and blue's work list holds no ruling on it to reconsider (estopped, no motion standing: %v)", id, ruled))
		return
	case id != "":
		arc = true
	case len(ruled) == 0:
		return
	default:
		id = ruled[r.rng.Intn(len(ruled))]
	}
	file := func(extra ...string) (string, error) {
		return r.exec(append([]string{"--json", "motion", "docket", "file", "--seat-id", seatID, "--id", id,
			"--reason", "fuzz: what was recorded since the bench ruled " + id + " closed is new"}, extra...)...)
	}
	out, err := file()
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s, which the bench ruled closed, was admitted with no new grounds: %s", id, firstLine(out)))
		return
	case !ok || env.Code != string(feov.Validation) || !strings.Contains(env.Error, "admitted on NEW GROUNDS only") || !strings.Contains(env.Error, "the bench ruled gap "+id):
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s, which the bench ruled closed, was refused, and not for wanting new grounds: %s", id, firstLine(out)))
		return
	}
	r.mu.Lock()
	r.groundsRefused++
	r.mu.Unlock()

	// A QUOTE THE RULING ALREADY HAD: the seeded sentence stands in the report, and no edit answering
	// the gap wrote it after the ruling.
	out, err = file("--quote", "§ fuzz sentence")
	env, ok = refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s quoting text no edit answering it wrote since the ruling was admitted: %s", id, firstLine(out)))
		return
	case !ok || env.Code != string(feov.Validation) || !strings.Contains(env.Error, "is not new grounds on gap "+id):
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s quoting text the ruling already had was refused, and not as old grounds: %s", id, firstLine(out)))
		return
	}
	r.mu.Lock()
	r.groundsRefused++
	r.mu.Unlock()

	name := "fuzz-grounds-" + id + ".js"
	dir, argv, err := r.writeBundle(name, "console.log('what the ruling on "+id+" did not have');")
	if err != nil {
		return
	}
	if _, ok := r.admit(append([]string{"compute", "--seat-id", seatID, "--quote", "§ fuzz sentence", "--dir", dir, "--answers", id,
		"--reason", "fuzz: new grounds on " + id, "--"}, argv...)...); !ok {
		return
	}
	computation := r.latestComputationAnswering(id)
	if computation == "" {
		r.noteActErr("blue computed new grounds on " + id + " and no computation answering it is on the record")
		return
	}
	if _, ok := r.admitFresh(id, "motion", "docket", "file", "--seat-id", seatID, "--id", id, "--grounds", computation,
		"--reason", "fuzz: computation "+computation+" was recorded since the bench ruled "+id+" closed"); ok {
		r.mu.Lock()
		r.groundsAdmitted++
		if arc {
			r.arc.ruled.stage = arcReconsidered
		}
		r.mu.Unlock()
	}
}

// latestComputationAnswering is the id of the newest computation on the record answering gapID, ""
// where none does.
func (r *runner) latestComputationAnswering(gapID string) string {
	b, ok := r.board()
	if !ok {
		return ""
	}
	id := ""
	for _, e := range b.Events {
		if p, ok := recordpb.BodyAs[*recordpb.Computation](e); ok && p.GetAnswers() == gapID {
			id = p.GetComputationId()
		}
	}
	return id
}

// THE CONSTRUCTION (benchOrdersSeed). Two gaps, each minted by a lens at its first sitting for the
// purpose, each with a scenario under which no seat's own drive moves or closes it (dirIgnore), at grades
// that keep it material whatever the bench does to its likelihood, so the dispatch keeps readying its
// lens and blue for it. arcHolds keeps every other arm off them until their part is done.
//
// THE LOCKING GAP, in order, each step on the record's answer to the last:
//
//  1. blue files a docket motion on it (benchArcBlue), unless one already stands;
//  2. the bench, hearing it, moves its likelihood: LOCKED (benchArcHears), and remands it;
//  3. its lens, in the remand's exchange, moves the likelihood and is REFUSED, then moves its impact,
//     which the record admits: movement, so the exchange leaves the gap short of impasse and a
//     remand at the next hearing is not past the limit (benchArcLens);
//  4. blue files a docket motion on it again, and the bench, hearing it, moves the likelihood again:
//     RELEASED, and remands it;
//  5. its lens moves the likelihood and is ADMITTED: lockLifted.
//
// One regrade per gap per sitting: each step is one sitting's act on the gap. Every bench step needs
// the docket motion its own step filed or the dispatch did; every bench move moves the axis.
//
// THE RULED GAP rides the locking gap's sittings: blue files a docket motion on it with step 1, the
// bench rules it closed (not_a_defect) at that hearing, and blue's next sitting — the remand of step 2
// readies blue — reconsiders it: refused with no grounds and on old grounds, admitted on a computation.
//
// A run that ends with either gap short of its last step fails, naming the step (arcUnfinished).

// arcAxis is the axis the bench locks and releases, arcFreeAxis the one the lens moves while it is
// locked. Severity stays high, so the gap stays material.
const (
	arcAxis     = "likelihood"
	arcFreeAxis = "impact"
)

// arcStage is how far one of the construction's gaps has come.
type arcStage string

const (
	arcUnminted     arcStage = "not minted"
	arcMinted       arcStage = "minted"
	arcLocked       arcStage = "locked by the bench"
	arcRefused      arcStage = "the lens's move refused"
	arcReleased     arcStage = "released by the bench"
	arcLifted       arcStage = "the lens's move admitted"
	arcRuledClosed  arcStage = "ruled closed by the bench"
	arcReconsidered arcStage = "reconsidered on new grounds"
)

// arcGap is one of the construction's gaps: its id, its minting lens and its stage. why is the last
// refusal of its mint, for a run whose gap never landed.
type arcGap struct {
	gap, lens, why string
	stage          arcStage
}

// benchArc is the construction's two gaps, guarded by runner.mu; mintMu serializes their mints, which
// lenses sitting in parallel reach together.
type benchArc struct {
	locking, ruled arcGap
	mintMu         sync.Mutex
}

// arcHolds reports whether gapID is one of the construction's gaps short of its last step.
func (r *runner) arcHolds(gapID string) bool {
	if !r.forceBenchOrders || gapID == "" {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return (gapID == r.arc.locking.gap && r.arc.locking.stage != arcLifted) ||
		(gapID == r.arc.ruled.gap && r.arc.ruled.stage != arcReconsidered)
}

// arcMint is, at seatID's sitting, the mint of whichever construction gap no lens has minted yet: the
// locking gap first, the ruled gap by another lens. A refused mint leaves it to the next lens sitting.
func (r *runner) arcMint(seatID string) {
	r.arc.mintMu.Lock()
	defer r.arc.mintMu.Unlock()
	r.mu.Lock()
	var slot *arcGap
	switch {
	case r.arc.locking.gap == "":
		slot = &r.arc.locking
	case r.arc.ruled.gap == "" && r.arc.locking.lens != seatID:
		slot = &r.arc.ruled
	}
	r.mu.Unlock()
	if slot == nil {
		return
	}
	gap, why := r.mintArcGap(seatID)
	r.mu.Lock()
	defer r.mu.Unlock()
	if gap == "" {
		slot.why = why
		return
	}
	*slot = arcGap{gap: gap, lens: seatID, stage: arcMinted}
}

// mintArcGap mints, as seatID, one construction gap on the anchor sentence under the first pool class no
// open gap holds there. It returns the gap, or "" and the refusal.
func (r *runner) mintArcGap(seatID string) (string, string) {
	args := []string{"--json", "mint", "--seat-id", seatID, "--problem", "fuzz: the bench-orders construction's gap",
		"--check-kind", "document", "--check", "acc", "--fix", dirIgnore,
		"--severity", "high", "--likelihood", "high", "--impact", "high",
		"--quote", "A § fuzz sentence to anchor findings.", "--class", "",
		"--reason", "fuzz: a gap the bench sets a grade on"}
	at := slices.Index(args, "--class") + 1
	for _, c := range mintPoolClasses {
		if open, _ := r.sameClassAndSubject(c, args); len(open) == 0 {
			args[at] = c
			break
		}
	}
	if args[at] == "" {
		return "", "every pool class stands open on the anchor sentence"
	}
	out, err := r.exec(args...)
	if err != nil {
		return "", firstLine(out)
	}
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		return "", "the mint's result named no gap: " + firstLine(out)
	}
	r.noteMint(seatID, env.Result.GapID)
	r.checkOneGap(env.Result.GapID, args)
	return env.Result.GapID, ""
}

// benchArcLens is the construction at seatID's lens sitting: a mint where a gap is still owed, then on
// the locking gap it minted, step 3 or step 5. It returns the gap it regraded, "" for none.
func (r *runner) benchArcLens(seatID string) string {
	if !r.forceBenchOrders {
		return ""
	}
	r.arcMint(seatID)
	r.mu.Lock()
	a := r.arc.locking
	r.mu.Unlock()
	if a.lens != seatID || !slices.Contains(r.openGaps(), a.gap) {
		return ""
	}
	switch a.stage {
	case arcLocked:
		if !r.lockRefusesLens(seatID, a.gap, arcAxis, "low") {
			return ""
		}
		if _, ok := r.admitFresh(a.gap, "regrade", "--seat-id", seatID, "--id", a.gap, "--"+arcFreeAxis, "medium_high",
			"--reason", "fuzz: the lens moving an axis the bench does not hold"); !ok {
			return a.gap
		}
		r.setArcStage(a.gap, arcRefused)
		return a.gap
	case arcReleased:
		if _, ok := r.admitFresh(a.gap, "regrade", "--seat-id", seatID, "--id", a.gap, "--"+arcAxis, "medium",
			"--reason", "fuzz: the lens moving an axis the bench released"); !ok {
			return a.gap
		}
		r.mu.Lock()
		delete(r.benchLifted[a.gap], arcAxis)
		r.lockLifted++
		r.arc.locking.stage = arcLifted
		r.mu.Unlock()
		return a.gap
	}
	return ""
}

// benchArcBlue is the construction at blue's sitting: a docket motion on each of its gaps whose next
// step is the bench's — the ruled gap once minted, the locking gap once minted and once its lens's
// move was refused — unless one already stands.
func (r *runner) benchArcBlue(seatID string) {
	if !r.forceBenchOrders {
		return
	}
	r.mu.Lock()
	ruled, locking := r.arc.ruled, r.arc.locking
	r.mu.Unlock()
	for _, a := range []arcGap{ruled, locking} {
		if a.stage != arcMinted && a.stage != arcRefused {
			continue
		}
		if !slices.Contains(r.openGaps(), a.gap) || len(r.unruledDocketMotions(a.gap)) > 0 {
			continue
		}
		r.admitFresh(a.gap, "motion", "docket", "file", "--seat-id", seatID, "--id", a.gap,
			"--reason", "fuzz: "+a.gap+" is the bench's to hear")
	}
}

// benchArcHears is the construction at the bench's hearing of gapID: the ruled gap ruled closed, the
// locking gap's likelihood moved — locked at step 2, released at step 4. It reports whether it ruled
// gapID itself; the locking gap is ruled as any gap is.
func (r *runner) benchArcHears(seatID, gapID string) bool {
	if !r.forceBenchOrders {
		return false
	}
	r.mu.Lock()
	ruled, locking := r.arc.ruled, r.arc.locking
	r.mu.Unlock()
	if !slices.Contains(r.openGaps(), gapID) || len(r.unruledDocketMotions(gapID)) == 0 {
		return false
	}
	switch {
	case gapID == ruled.gap && ruled.stage == arcMinted:
		r.benchDisposes(seatID, gapID, "not_a_defect", "--final", "--reason", "fuzz: the bench rules "+gapID+" closed")
		if slices.Contains(r.openGaps(), gapID) {
			r.noteActErr("the bench ruled " + gapID + " not_a_defect for the bench-orders construction, and it is still open")
			return true
		}
		r.setArcStage(gapID, arcRuledClosed)
		return true
	case gapID == locking.gap && locking.stage == arcMinted:
		if r.benchMoves(seatID, gapID, arcAxis, "medium_high") {
			r.setArcStage(gapID, arcLocked)
		}
	case gapID == locking.gap && locking.stage == arcRefused:
		if r.benchMoves(seatID, gapID, arcAxis, "high") {
			r.setArcStage(gapID, arcReleased)
		}
	}
	return false
}

// setArcStage moves the construction gap gapID to stage.
func (r *runner) setArcStage(gapID string, stage arcStage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch gapID {
	case r.arc.locking.gap:
		r.arc.locking.stage = stage
	case r.arc.ruled.gap:
		r.arc.ruled.stage = stage
	}
}

// arcAwaitsReconsideration is the ruled gap while blue owes its reconsideration, "" otherwise.
func (r *runner) arcAwaitsReconsideration() string {
	if !r.forceBenchOrders {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.arc.ruled.stage == arcRuledClosed {
		return r.arc.ruled.gap
	}
	return ""
}

// arcUnfinished is, on benchOrdersSeed's run, why the construction did not reach its last step on each
// gap; "" where it did, and on every other run.
func (r *runner) arcUnfinished() string {
	if !r.forceBenchOrders {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var short []string
	for _, a := range []struct {
		name string
		gap  arcGap
		last arcStage
	}{{"locking", r.arc.locking, arcLifted}, {"ruled", r.arc.ruled, arcReconsidered}} {
		if a.gap.stage == a.last {
			continue
		}
		stage := a.gap.stage
		if stage == "" {
			stage = arcUnminted
		}
		s := fmt.Sprintf("the %s gap %q (lens %q) stopped at %q, short of %q", a.name, a.gap.gap, a.gap.lens, stage, a.last)
		if a.gap.why != "" {
			s += " — its mint was last refused: " + a.gap.why
		}
		short = append(short, s)
	}
	if len(short) == 0 {
		return ""
	}
	return "benchOrdersSeed's construction did not finish: " + strings.Join(short, "; ")
}
