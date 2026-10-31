package fuzz

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// WHEN THE CHAIR HOLDS BACK A PASS ITS PLAN PERMITS, THE BENCH HEARS ITS POSITION (gblock,
// 2026-10-30). The drive stands in for each seat of that exchange: the chair's withheld PASS files the
// position (chairPosition), the plan it asks for next convenes the bench, and the bench's position
// sitting — or the terminal sitting, at the epoch limit — rules it upheld or overruled on a coin. An
// upheld ruling names a cast lens or open gaps, with a direction; an overruled one names nothing, and
// the chair's next sitting records its verdict, after the tool refuses a second position there.
// positionOracle holds every position the chair withheld a PASS on to that route.

// positionIDOf pulls the tool-assigned position id out of a position's JSON envelope.
func positionIDOf(out string) string {
	var env struct {
		Result struct {
			PositionID string `json:"position_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		return ""
	}
	return env.Result.PositionID
}

// unruledPositions is every position motion the record holds unruled: the positions the bench hears.
func (r *runner) unruledPositions() []string {
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		r.noteActErr("the record could not be read for the positions the bench hears: " + err.Error())
		return nil
	}
	var ids []string
	for _, m := range record.MotionsOf(fam.Events, fam.At) {
		if m != nil && m.Subject == "position" && !m.Ruled() {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

// THE THREE RULINGS ARE DRIVEN IN EVERY SWEEP, NOT DRAWN. A withheld PASS is rare on its own (one in
// forty runs), and a coin over a rare event drove neither upheld form in a whole sweep, so `--ids`
// went unpassed and the tally failed. Each sweep process owes each ruling — upheld naming a lens,
// upheld naming open gaps, and overruled — once: while one is owed, a chair whose plan permits the
// PASS withholds it (owesWithhold), and the bench rules the ruling still owed (rulePositions). Once
// all three are driven, both fall back to the coin. The sweep's tally gates each count at zero.
//
// The ledger is the SWEEP's: runOne hands sweepPositionForms to the runs it drives. A runner built
// for a deterministic test holds none, owes nothing, and draws on the coin alone, so no test's
// outcome depends on which tests ran before it in the process.
type positionFormLedger struct {
	mu                    sync.Mutex
	lens, gaps, overruled bool
}

var sweepPositionForms = &positionFormLedger{}

// owed reports which rulings the ledger has not yet seen driven; a nil ledger owes none.
func (l *positionFormLedger) owed() (lens, gaps, overruled bool) {
	if l == nil {
		return false, false, false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return !l.lens, !l.gaps, !l.overruled
}

func (l *positionFormLedger) driven(form string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	switch form {
	case "lens":
		l.lens = true
	case "gaps":
		l.gaps = true
	case "overruled":
		l.overruled = true
	}
}

// owesWithhold reports whether a chair whose plan permits the PASS withholds it to reach a ruling
// still owed: upheld on a lens or overruled, always; upheld on gaps only where open gaps stand for
// the bench to name.
func (r *runner) owesWithhold() bool {
	lens, gaps, overruled := r.positionForms.owed()
	return lens || overruled || (gaps && len(r.openGaps()) > 0)
}

// rulePositions rules every position the bench hears, at the sitting of the given occasion. A forced
// run's bench overrules. Otherwise the bench rules what this sweep still owes — upheld on open gaps
// where some stand, upheld on a lens, overruled — and past those, upheld or overruled on a coin.
// Upheld names open gaps the tool accepts, or a cast lens.
func (r *runner) rulePositions(seatID, occasion string) {
	if r.positionRuledAt == nil {
		r.positionRuledAt = map[string]string{}
	}
	for _, id := range r.unruledPositions() {
		const opinion = "fuzz: the bench's reasoning on the PASS withheld"
		owedLens, owedGaps, owedOverruled := r.positionForms.owed()
		open := r.openGaps()
		var form string
		switch {
		// A FORCED RUN'S BENCH OVERRULES, and drives no refusal: its chair's next sitting decides on
		// the PASS, so the forced run keeps the exit it is forced to (UNVERIFIED: its chair leaves its
		// own motions unruled, so that verdict is a FAIL or none).
		case r.forcedVerdict():
			form = "overruled"
		case owedGaps && len(open) > 0:
			form = "gaps"
		case owedLens:
			form = "lens"
		case owedOverruled:
			form = "overruled"
		case r.coin(50):
			form = "overruled"
		case len(open) > 0 && r.coin(50):
			form = "gaps"
		default:
			form = "lens"
		}
		if form == "overruled" {
			// AN OVERRULED RULING NAMES NO WORK, and one that names a lens is refused.
			if !r.forcedVerdict() {
				r.maybe(20, func() {
					r.positionRulingRefused(seatID, id, "overruled naming a lens", "--as", "overruled", "--lens", r.castLens(), "--reason", opinion)
				})
			}
			if _, err := r.exec("motion", "position", "rule", "--seat-id", seatID, "--id", id, "--as", "overruled", "--reason", opinion); err != nil {
				r.noteActErr(fmt.Sprintf("the bench's overruling of position %s was refused: %v", id, err))
				continue
			}
			r.positionForms.driven("overruled")
			r.mu.Lock()
			r.positionRuledAt[id] = occasion
			r.positionsOverruled++
			r.overruledPending = id
			r.mu.Unlock()
			continue
		}
		// AN UPHELD RULING NAMES ITS WORK ONE WAY: naming neither is refused.
		r.maybe(20, func() {
			r.positionRulingRefused(seatID, id, "upheld naming no work", "--as", "upheld", "--direction", "fuzz direction", "--reason", opinion)
		})
		const direction = "fuzz: audit the claim the chair would not pass"
		if form == "gaps" {
			// THE FIRST OPEN GAP THE TOOL ACCEPTS: one remanded at impasse already, or standing
			// docketed unruled, is refused, and the next is named instead.
			form = "lens"
			for _, g := range open {
				if _, err := r.exec("motion", "position", "rule", "--seat-id", seatID, "--id", id, "--as", "upheld",
					"--ids", g, "--direction", direction, "--reason", opinion); err == nil {
					form = "gaps"
					r.remandRouted(g)
					break
				}
			}
		}
		if form == "lens" {
			if _, err := r.exec("motion", "position", "rule", "--seat-id", seatID, "--id", id, "--as", "upheld",
				"--lens", r.castLens(), "--direction", direction, "--reason", opinion); err != nil {
				r.noteActErr(fmt.Sprintf("the bench's upholding of position %s, naming a cast lens, was refused: %v", id, err))
				continue
			}
		}
		r.positionForms.driven(form)
		r.mu.Lock()
		r.positionRuledAt[id] = occasion
		if form == "gaps" {
			r.positionsUpheldGaps++
		} else {
			r.positionsUpheldLens++
		}
		r.mu.Unlock()
	}
}

// castLens is a lens of the run's cast, drawn.
func (r *runner) castLens() string {
	cast, err := record.CastOf(r.run())
	if err != nil {
		return fuzzLensSeats[0]
	}
	var lenses []string
	for _, s := range cast {
		if strings.HasPrefix(s, "red-lens-") {
			lenses = append(lenses, s)
		}
	}
	if len(lenses) == 0 {
		return fuzzLensSeats[0]
	}
	return pick(r.rng, lenses)
}

// positionRulingRefused drives a ruling on position id that names the wrong work, and holds the tool
// to refusing it.
func (r *runner) positionRulingRefused(seatID, id, what string, args ...string) {
	out, err := r.exec(append([]string{"motion", "position", "rule", "--seat-id", seatID, "--id", id}, args...)...)
	if err == nil {
		r.noteActErr(fmt.Sprintf("the bench's ruling on position %s, %s, was admitted: %s", id, what, firstLine(out)))
		return
	}
	r.mu.Lock()
	r.positionRulingsRefused++
	r.mu.Unlock()
}

// repeatedPositionRefused files the chair's position at its first sitting after the bench overruled
// its last one, and holds the tool to refusing it as a conflict naming the overruled position.
func (r *runner) repeatedPositionRefused(seatID, overruled string) {
	out, err := r.exec("--json", "position", "--seat-id", seatID, "--reason", "fuzz: the same objection again")
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("the chair's position at its first sitting after the bench overruled %s was admitted: %s", overruled, firstLine(out)))
	case !ok || env.Code != string(feov.Conflict) || !strings.Contains(env.Error, overruled):
		r.noteActErr(fmt.Sprintf("the chair's repeated position was refused without naming the overruled position %s as a conflict: %s", overruled, firstLine(out)))
	default:
		r.mu.Lock()
		r.repeatedPositionsRefused++
		r.mu.Unlock()
	}
}

// takeOverruled is the position the bench overruled since the chair last sat, cleared: the chair's
// next sitting owes the verdict.
func (r *runner) takeOverruled() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.overruledPending
	r.overruledPending = ""
	return id
}

// positionOracle holds every position the chair filed on a PASS it withheld to its route: the bench
// rules it, at the position sitting a plan convened or at the terminal sitting. A halted run is
// exempt, as the petition oracle exempts it. "" when it holds.
func (r *runner) positionOracle(halted bool) string {
	if halted {
		return ""
	}
	for _, id := range r.positionsHeard {
		switch at, ruled := r.positionRuledAt[id]; {
		case !ruled:
			return fmt.Sprintf("position oracle: the chair withheld a PASS its plan permitted and filed position %s, and the bench never ruled it", id)
		case at != "position" && at != "terminal":
			return fmt.Sprintf("position oracle: position %s was ruled at the %s sitting — the plan convenes the bench for it, or the terminal sitting hears it at the exit", id, at)
		}
	}
	return ""
}

// THE DRIVE'S BENCH RULES THE POSITION IT HEARS, and the oracle holds it to the route. Pinned outside
// the sweep: a PASS withheld at the final sitting is ruled by the drive's terminal bench, the oracle
// passes it, and a position it never ruled fails the oracle.
func TestTheDrivesBenchRulesThePositionItHears(t *testing.T) {
	r, _ := converseRun(t, redClosesAndWithholds)
	ids := r.unruledPositions()
	if len(ids) != 1 {
		t.Fatalf("the record holds %v unruled positions, want the one the chair withheld the PASS on", ids)
	}
	r.positionsHeard = ids
	if m := r.positionOracle(false); m == "" || !strings.Contains(m, ids[0]) {
		t.Errorf("the oracle held over a position the bench never ruled: %q", m)
	}
	r.dispatchSits("bench", "judge", "terminal")
	r.rulePositions("judge", "terminal")
	r.dispatchReturns("judge", false)
	if r.actErr != "" {
		t.Fatalf("the drive's ruling: %s", r.actErr)
	}
	if left := r.unruledPositions(); len(left) != 0 {
		t.Fatalf("the drive's bench left %v unruled", left)
	}
	if r.positionRuledAt[ids[0]] != "terminal" || r.positionsUpheldLens+r.positionsUpheldGaps+r.positionsOverruled != 1 {
		t.Errorf("ruled at %q, %d upheld on a lens, %d on gaps and %d overruled; want one ruling at the terminal sitting", r.positionRuledAt[ids[0]], r.positionsUpheldLens, r.positionsUpheldGaps, r.positionsOverruled)
	}
	if m := r.positionOracle(false); m != "" {
		t.Errorf("the oracle fired over a position ruled at the terminal sitting: %s", m)
	}
}

// AN OVERRULED POSITION SENDS THE PASS BACK TO THE CHAIR, through the verbs the drive drives: the
// bench's position sitting overrules it, the chair's next sitting is refused a second position naming
// the overruled one, and records the PASS its plan permits.
func TestTheDrivesChairDecidesThePassAfterAnOverruling(t *testing.T) {
	r, last := converseRun(t, redClosesAndWithholdsEarly)
	if last.Ceiling || len(last.Parties) != 1 || last.Parties[0].SeatID != "judge" || len(last.Parties[0].Occasions) != 1 || last.Parties[0].Occasions[0] != "position" {
		t.Fatalf("the plan after the position = %+v, want the bench convened for the position alone", last)
	}
	id := r.unruledPositions()[0]
	r.dispatchSits("bench", "judge", "position")
	if _, err := r.exec("motion", "position", "rule", "--seat-id", "judge", "--id", id, "--as", "overruled", "--reason", "fuzz: the objection does not hold"); err != nil {
		t.Fatal(err)
	}
	r.dispatchReturns("judge", false)
	r.dispatchSits("chair", "red-chair")
	out, err := r.exec("--json", "dispatch", "next", "--seat-id", "red-chair")
	if err != nil || !strings.Contains(out, `"pass_permitted":true`) {
		t.Fatalf("the chair's next plan: %v\n%s", err, out)
	}
	r.repeatedPositionRefused("red-chair", id)
	if r.actErr != "" || r.repeatedPositionsRefused != 1 {
		t.Fatalf("the repeated position: %q, %d refused", r.actErr, r.repeatedPositionsRefused)
	}
	if _, err := r.exec("--json", "verdict", "--seat-id", "red-chair", "--as", "PASS"); err != nil {
		t.Fatalf("the PASS the overruling sent back: %v", err)
	}
	r.dispatchReturns("red-chair", false)
	if got, _, ok, err := record.DeriveVerdict(r.run()); err != nil || !ok || got != "VERIFIED" {
		t.Fatalf("derived (%q, %v, %v), want VERIFIED", got, ok, err)
	}
}

// THE SWEEP'S LEDGER PICKS THE RULING, not the coin: a bench whose sweep still owes one ruling makes
// that one, and marks it driven, so a sweep reaches each ruling however rarely a PASS is withheld.
func TestTheDrivesBenchRulesTheRulingTheSweepOwes(t *testing.T) {
	// EACH ROW'S COIN LANDS AGAINST ITS RULING — every coin heads where lens is owed, every coin tails
	// where overruled is — so the coin alone would rule the other way, and only the ledger rules this.
	for owed, draw := range map[string]int64{"lens": 0, "overruled": 99} {
		t.Run(owed, func(t *testing.T) {
			r, _ := converseRun(t, redClosesAndWithholds)
			r.rng = &lockedRand{r: rand.New(fixedSource(draw))}
			r.positionForms = &positionFormLedger{lens: owed != "lens", gaps: true, overruled: owed != "overruled"}
			if !r.owesWithhold() {
				t.Fatalf("a sweep owing the %s ruling does not withhold a permitted PASS", owed)
			}
			r.positionsHeard = r.unruledPositions()
			r.dispatchSits("bench", "judge", "terminal")
			r.rulePositions("judge", "terminal")
			r.dispatchReturns("judge", false)
			if r.actErr != "" {
				t.Fatalf("the drive's ruling: %s", r.actErr)
			}
			got := map[string]int{"lens": r.positionsUpheldLens, "gaps": r.positionsUpheldGaps, "overruled": r.positionsOverruled}
			if got[owed] != 1 || got["lens"]+got["gaps"]+got["overruled"] != 1 {
				t.Errorf("rulings %v, want the one %s ruling the sweep owed", got, owed)
			}
			if r.owesWithhold() {
				t.Errorf("the ledger still owes a ruling after the %s ruling was driven", owed)
			}
		})
	}
}

// fixedSource draws n on every Intn(100): its 31-bit draw is n, below every rejection bound.
type fixedSource int64

func (f fixedSource) Int63() int64 { return int64(f) << 32 }
func (fixedSource) Seed(int64)     {}
