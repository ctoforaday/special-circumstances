package fuzz

// Fuzz the ACTUAL debate.js orchestrator against the REAL feov-record binary. goja runs the
// unmodified script (harness faked); each agent() call is a seat that makes coherent, random,
// VALID tool calls into a real run directory and returns a randomised envelope to drive
// debate.js's own branches (verdict, gap counts, rounds, petitions). A failing run is a real
// finding (in debate.js, the tool, or verify), reproducible from its seed.
//
// COVERAGE CONTRACT. envelopeFor drives every eligible seat to exercise its whole verb surface,
// not a happy path: lens (cite/finding/line of inquiry/friction), merge (position/closing/
// mint/close incl. repaired_with_regression/regrade any axis/
// dispute-respond/spot-check/verdict/petition), blue (position/closing/dispute
// across all four dimensions/manifest-row/line of inquiry/revision/retire/petition), bench
// (opinion/outcome incl. --ended/certify/assemble/petition-rule). The
// petition->petition-rule docket and the disputes docket are driven through the ENVELOPE (see
// maybePetition/rulePetitions, raiseDisputes/answerDisputes), so debate.js's routing runs too.
//
// ORACLES per run: (1) verify passes — whatever path the debate took, the record satisfies every
// invariant; (2) the JSON views (findings/friction/debate) exit 0 and parse; (2c) the six markdown
// views (ledger/archive/debate/changelog/citation-ledger/lines-of-inquiry) render in-memory via
// view.Markdown and exit 0; (3) every dialectic prose renders in the report (the A1-A3
// write-here/read-there class); (4) #111 tier — every seat dispatched on the configured tier.
// COVERAGE GATE: across the run set, every event-emitting verb
// in verbsWithEvents must fire at least once (a regression that silently drops one fails loudly).
// `halt` terminates the run, so it is covered by the dedicated TestFuzzHaltPath, not the sweep.
//
// Run: go test ./integration/fuzz -run TestFuzzDebate -count=1   (respects -short by shrinking N).
// Confidence sweep: FUZZ_N=1000 go test ./integration/fuzz -run TestFuzzDebate -timeout 1200s.
// FUZZ_C overrides concurrency (runs are subprocess-bound, so the default oversubscribes cores).

import (
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Capture only seat-id characters — NOT the trailing "." in "SEAT_ID: red-merge-r1." — or the
// malformed id makes every tool call fail silently and the fuzz degrades to trivial PASS runs.
var seatRe = regexp.MustCompile(`SEAT_ID:\s*([A-Za-z0-9-]+)`)

// THE CITATION AXIS NEEDS A SOURCE TO FETCH (#256). `feov-record fetch` performs a REAL http GET —
// that IS the code under test, caps and all — so this harness serves it from 127.0.0.1 rather than
// stubbing it out. Loopback only: the same discipline internal/fetchcache's own tests use, so CI
// touches no external network while the whole fetch → cache → cite → anchor path runs end to end
// through the real binary. One server per test binary; the path varies so distinct URLs exercise
// cache misses as well as hits.
var (
	sourceSrvOnce sync.Once
	sourceSrvURL  string
)

func sourceURL(path string) string {
	sourceSrvOnce.Do(func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "fuzz source body for %s\n", req.URL.Path)
		}))
		sourceSrvURL = srv.URL // deliberately never closed: it lives for the test binary
	})
	return sourceSrvURL + path
}

// lockedRand is the seed's generator, safe for the concurrent seats phase 3 introduces.
//
// A WRAPPER RATHER THAN A MUTEX AT EVERY CALL SITE: there are 23 draws across this file, all
// Intn, and a lock the caller has to remember is a lock somebody eventually does not. The type
// makes the guarantee structural — a draw cannot be taken without it.
//
// It does NOT preserve per-seed determinism, and that is stated rather than hidden: the SEQUENCE
// is deterministic, but which seat takes which draw depends on the interleaving, so a seed no
// longer reproduces its own run. That cost is #630's, argued there; the run directory is kept on
// failure and the record carries the true ordering, which is what a failing seed is read from.
type lockedRand struct {
	mu sync.Mutex
	r  *rand.Rand
}

func newLockedRand(seed int64) *lockedRand {
	return &lockedRand{r: rand.New(rand.NewSource(seed))}
}

func (l *lockedRand) Intn(n int) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.r.Intn(n)
}

// runner shells the real binary for one run's seats. runDir + bin are fixed per fuzz iteration.
type runner struct {
	// mu guards every mutable field below. The seats run concurrently as of #630 phase 3, and
	// these were single-threaded by accident rather than by design. It is held around the state
	// itself and NEVER across r.exec — a lock spanning a subprocess would put the serialization
	// back one layer down and buy nothing but a slower version of the old behaviour.
	mu          sync.Mutex
	bin, runDir string
	rng         *lockedRand
	registered  map[string]bool
	classMade   bool
	// #277: the gap ids minted with --check-kind computation. Such a gap CANNOT be closed
	// until a proof answers it, so closeGap satisfies it first — otherwise the fuzzer
	// accumulates unclosable gaps and every open-gap-scaled drive grows with them (measured:
	// a uniform draw took the 60-run sweep from 47s to a 900s timeout).
	computationGaps map[string]bool
	// #62 Stage 2: disputes blue RAISED (event emitted + envelope ref), awaiting red's answer
	// next round — mirrors debate.js's pendingDisputes so the fuzz drives the docket machinery
	// through the ENVELOPE, not just the events. Each: {gap_id, dimension, proposed}.
	raised []map[string]any
	// disputedThisRound is what blueRespondTo raised for the CURRENT round's envelope.
	disputedThisRound []map[string]any
	// ruledMotions are motions that HAVE a ruling and can therefore be appealed. Kept apart from
	// the pending set (r.raised) because the two states admit different verbs: an appeal against
	// no ruling is refused, and a list that mixed them would drive the verb without exercising it.
	ruledMotions []string
	// presented records the gaps a responder was actually shown: a gap minted in the terminal
	// round never reaches blue, so its scenario was never dispatched and cannot be asserted.
	presented map[string]bool
	// evaluated records the gaps red actually SAT ON after blue had responded — the only ones
	// whose terminal fate red is answerable for. Blue repairing in the final round leaves a
	// satisfied gap legitimately open, because red never sits again.
	evaluated map[string]bool
	// reproduced records the PROVE gaps whose proof a lens re-ran and confirmed. Red closes on
	// THIS, not on its own assertion that a computation happened.
	reproduced map[string]bool
	// applyMisses counts, by CAUSE, every time the verbatim-apply branch declined to apply.
	// Reported beside the verbatim tally so a zero there names its reason instead of implying one.
	applyMisses map[string]int
	// disputeRefusals is why a grade filing was REFUSED, per gap. The scenario oracle reports a
	// missing dispute event; without this it cannot distinguish a drive that never ran from one
	// that ran and was told no, and those want opposite fixes.
	disputeRefusals map[string]string
	// #111: every model an agent() call carried, one per dispatch ("unset" if absent). The tier
	// oracle asserts all equal the configured tier — map-free, needs no bulk-seat list here.
	models []string
	// petition docket: petitions a party seat RAISED (event emitted + envelope entry), awaiting
	// the judge-petition sitting hearPetitions dispatches next. Each: {who, class}. Mirrors the
	// disputes machinery (raised) so the fuzz drives the petition/petition-rule path end to end.
	petitioned []map[string]any
	// forceUnverified drives the run to the UNVERIFIED terminal verdict, which the random sweep
	// reaches about once in sixty — debate.js computes UNVERIFIED only when the bench declares
	// deadlock AND gaps remain open, and a bench that disposes its whole docket clears the board
	// and earns red a further sitting instead. A 1-in-60 outcome cannot carry an enum gate: the
	// sweep failed on `outcome --as UNVERIFIED` at N=40 having passed at N=60 on luck, which is a
	// flake either way. So it is driven the way HALTED is — deterministically, by a dedicated
	// test — rather than exempted. coverage_test.go says why an exemption is the wrong answer
	// here, and it is right.
	forceUnverified bool
	// forceHalt makes the judge-petition sitting rule HALT (the dedicated halt-path test), instead
	// of the granted/denied the random path uses. A halt ends the run HALTED; kept out of the
	// random oracle on purpose (it reshapes every downstream expectation).
	forceHalt bool
}

func (r *runner) coin(pct int) bool { return r.rng.Intn(100) < pct }

// maybe runs fn with pct% probability — the coin-gated-action idiom, without the `if` nesting.
func (r *runner) maybe(pct int, fn func()) {
	if r.coin(pct) {
		fn()
	}
}

// inquiryIDOf pulls the tool-assigned line-of-inquiry id out of a propose result.
func inquiryIDOf(out string) string {
	m := inquiryIDPat.FindStringSubmatch(out)
	if m == nil {
		return ""
	}
	return m[1]
}

// The record mints Q1, Q2 … — the id comes back on stdout and is never recomposed here.
var inquiryIDPat = regexp.MustCompile(`\b(Q\d+)\b`)

// cmd is a small fluent builder for a seat verb — `<role> <verb> --seat-id <seatID> …` (exec
// appends --run). It collapses the conditional-flag arg-slice boilerplate: set() always adds a
// flag, on() adds it with pct% probability, bare() adds a boolean flag. run() shells the binary.
type cmd struct {
	r    *runner
	args []string
}

// do takes the verb as it is TYPED, so a verb under a subgroup is passed as the words a seat
// would write ("line-of-inquiry propose") rather than needing a second parameter for the group.
// do builds one invocation. THE ROLE IS GONE FROM THE ARGS: the tree is scoped to the seat, so the
// seat id is the only thing that says which verbs exist.
func (r *runner) do(verb, seatID string) *cmd {
	return &cmd{r: r, args: append(strings.Fields(verb), "--seat-id", seatID)}
}
func (c *cmd) set(flag, val string) *cmd { c.args = append(c.args, flag, val); return c }
func (c *cmd) bare(flag string) *cmd     { c.args = append(c.args, flag); return c }
func (c *cmd) on(pct int, flag, val string) *cmd {
	if c.r.coin(pct) {
		c.args = append(c.args, flag, val)
	}
	return c
}
func (c *cmd) run() (string, error) { return c.r.exec(c.args...) }

// noteApplyMiss records WHY the verbatim-apply branch declined, so the estoppel precondition's
// count of zero arrives with a cause rather than inviting the reader to guess one.
func (r *runner) noteApplyMiss(why string) {
	if r.applyMisses == nil {
		r.applyMisses = map[string]int{}
	}
	r.applyMisses[why]++
}

// firstLine keeps a refusal's diagnosis and drops the help cobra staples to it — the tally is a
// histogram, and a full help page per key makes it unreadable.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// dialectic emits the round's transcript onto the record the way Stage 1 seats do — a position
// narrative and a closing per open gap — plus, at random, a regrade, a lineage-carrying
// close-with-regression, and (blue) a grade dispute / (merge) its answer. Unique prose per act
// so the report oracle can prove each one actually rendered.
func (r *runner) dialectic(role, seatID string, open []string) {
	_, _ = r.exec("position", "--seat-id", seatID, "--reason", "narrative from "+seatID)
	for _, id := range open {
		_, _ = r.exec("closing", "--seat-id", seatID, "--id", id, "--reason", "closing-for-"+id+"-by-"+seatID)
	}
	if role == "blue" {
	}
	if role == "merge" && len(open) > 0 && r.coin(30) {
		id := open[r.rng.Intn(len(open))]
		dim := pick(r.rng, regradeDims) // move any grade axis, not only likelihood
		// EVERY AXIS, not three of four. `--complexity` was never passed because disputeDims fed
		// this and the fourth axis was spelled by its payload key rather than by its flag.
		_, _ = r.exec("regrade", "--seat-id", seatID, "--id", id, "--reason", "regrade-basis-for-"+id,
			"--"+dim, r.g(), "--complexity", r.g())
	}
}

// THE SECOND DISPUTE DRIVER IS GONE. `raiseDisputes` sat here, fully written and NEVER CALLED —
// the live driver is in blueRespondTo's dirDisputeWon/dirDisputeLost arm. Migrating it to the
// motion verbs would have produced a second correct-looking copy of a path nothing runs, and the
// command tally would have gone on reporting the verb as driven either way.
//
// answerDisputes is red's Stage-2 answer to blue's pending disputes: EMIT a dispute-respond event
// (rationale on the record) and return the ROUTING REF ({gap_id, dimension, response}) for the
// envelope's dispute_responses. Clears the pending set. A random accept/reject exercises both the
// accepted-delta and rejected-held docket branches.
func (r *runner) answerDisputes(seatID string) []map[string]any {
	var refs []map[string]any
	for _, d := range r.raised {
		id, _ := d["gap_id"].(string)
		dim, _ := d["dimension"].(string)
		// THE ANSWER FOLLOWS THE SCENARIO, not a coin. The gap was minted WON or LOST, and
		// both branches must run: an accepted regrade moves the board's mass (the
		// accepted-delta path), a rejected one is held a round and then dockets. Coining it
		// meant red's answer bore no relation to the dispute it was answering.
		resp := "rejected"
		if r.scenarioOf(id) == dirDisputeWon {
			resp = "accepted"
		}
		mid, _ := d["motion_id"].(string)
		if mid == "" {
			continue
		}
		if _, err := r.exec("motion", "grade", "rule", "--seat-id", seatID, "--id", mid,
			"--as", resp, "--reason", "respond-rationale-for-"+id+"-by-"+seatID); err != nil {
			continue
		}
		// AN APPEAL IS ONLY POSSIBLE ONCE A RULING EXISTS, so the ruled id goes to blue rather
		// than blue appealing what it just filed. The first appeal driver fired in the same
		// breath as the filing and all 14 were REFUSED — driven in the tally, exercising nothing,
		// which is exactly the false green the refusal count was added to expose.
		r.ruledMotions = append(r.ruledMotions, mid)
		refs = append(refs, map[string]any{"gap_id": id, "dimension": dim, "response": resp})
	}
	r.raised = nil
	return refs
}

// maybePetition sometimes files a petition (W2c, the constitutional short-circuit): it emits the
// petition event AND returns the envelope's petitions entry, tracking {who,class} so the
// judge-petition sitting hearPetitions dispatches next can rule on it. Only the seats debate.js
// actually routes to hearPetitions are eligible (blue-synthesize/blue-respond/red-merge); the
// random path never returns a HALT ruling, so the run continues. Returns arr() (no petition) most
// of the time — a petition detours the run through a bench sitting, so it stays occasional.
func (r *runner) maybePetition(role, seatID string) []any {
	if !r.forceHalt && !r.coin(20) { // forceHalt guarantees a petition so the halt sitting fires
		return arr()
	}
	class := pick(r.rng, petitionClasses)
	basis := "fuzz petition basis from " + seatID
	entry := map[string]any{"who": seatID, "class": class}
	out, err := r.exec("--json", "motion", "petition", "file", "--seat-id", seatID,
		"--class", class, "--relief", "fuzz relief", "--reason", basis)
	if err != nil {
		return arr()
	}
	id := motionIDOf(out)
	if id == "" {
		return arr()
	}
	entry["motion"] = id
	r.petitioned = append(r.petitioned, entry)
	return arr(map[string]any{"class": class, "basis": basis, "relief": "fuzz relief"})
}

// rulePetitions is the judge-petition sitting: it rules on every pending petition and returns the
// envelope rulings, clearing the docket. petition-rule takes only granted|denied (a halt is its
// OWN verb), so a forceHalt run emits `bench halt` and returns a halt ruling — the dedicated
// halt-path test; the random path rules granted/denied and the run continues.
func (r *runner) rulePetitions(seatID string) map[string]any {
	var rulings []any
	for _, p := range r.petitioned {
		who, _ := p["who"].(string)
		class, _ := p["class"].(string)
		opinion := "fuzz ruling opinion for " + who
		ruling := "granted"
		if r.coin(50) {
			ruling = "denied"
		}
		// THE PETITION IS ALWAYS RULED ON ITS MERITS. A halt is a separate decision about the
		// RUN, on its own channel (#329) — both can be true, and the petition does not stop
		// being answered because the bench also ended the run.
		id, _ := p["motion"].(string)
		// A GRANT NAMES WHO IT BINDS. Relief with no addressee is what #360 measured — the engine
		// threaded it into one hardcoded prompt, so relief for any other seat reached nothing.
		binds := pick(r.rng, []string{"blue", "red", "both"})
		args := []string{"motion", "petition", "rule", "--seat-id", seatID, "--id", id,
			"--as", ruling, "--reason", opinion}
		if ruling == "granted" {
			args = append(args, "--binds", binds)
		}
		_, _ = r.exec(args...)
		rulings = append(rulings, map[string]any{"petitioner": who, "class": class, "ruling": ruling, "relief": opinion, "binds": binds})
	}
	r.petitioned = nil
	if rulings == nil {
		rulings = arr()
	}
	env := map[string]any{"rulings": rulings, "friction": arr()}
	if r.forceHalt {
		// `bench halt` writes the record; the envelope's halt object is only what stops the
		// engine. The fake already drove the verb correctly before #329 — it was the PROMPT
		// that told a real judge to record a halt through petition-rule, where the enum refuses
		// it. The fuzz stayed green over a production path that could not work.
		haltOpinion := "fuzz judicial halt — safety boundary"
		_, _ = r.exec("halt", "--seat-id", seatID, "--reason", haltOpinion)
		env["halt"] = map[string]any{"opinion": haltOpinion}
	}
	return env
}

// exec runs one seat command. THE ERROR CARRIES WHAT THE TOOL SAID.
//
// It returned exec.ExitError verbatim — `exit status 2` — so every caller that kept its error
// still learned nothing. The dispute oracle was instrumented to print the refusal and printed
// "exit status 2"; the refusal itself was in the output the error did not mention. Six defects
// this session were a discarded refusal, and the cost was always the distance between the
// refusal and the symptom. Attaching it here shortens that distance for every call site at once.
func (r *runner) exec(args ...string) (string, error) {
	cmd := exec.Command(r.bin, append(args, "--run", r.runDir)...)
	out, err := cmd.CombinedOutput()
	noteExec(args, err, out)
	if err != nil {
		err = fmt.Errorf("%s: %w\n  %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), err
}

func (r *runner) register(role, seatID string) {
	r.mu.Lock()
	if r.registered[seatID] {
		r.mu.Unlock()
		return
	}
	r.registered[seatID] = true
	r.mu.Unlock()
	// OUTSIDE THE LOCK, deliberately: this is the invocation three concurrent lanes make against
	// a run directory whose database may not exist yet, and holding the lock across it would
	// serialize exactly the contention phase 3 exists to produce.
	_, _ = r.exec("register", "--seat-id", seatID)
}

// UNDERSCORES, because that is what the grade enum spells. This was `low-medium` and
// `medium-high`: two of five values refused at every site that generates a grade, which is
// `--proposed` on a grade dispute and both axes of a regrade. Measured — `merge regrade=8(7
// refused)`, `motion grade file=7(2 refused)`, and a scenario oracle three layers away reporting
// `R1-1 has no dispute event` for a contest whose drive was fine.
//
// Sixth instance this session of one value spelled two ways across a boundary with one side
// moved, and the third of them inside this fuzz. The others: the hyphenated direction rulings,
// and five drives naming commands that do not exist.
var grades = []string{"low", "low_medium", "medium", "medium_high", "high"}

func (r *runner) g() string { return grades[r.rng.Intn(len(grades))] }

// currentGrade reads the gap's grade at one axis, in the words a seat types. Empty when the board
// cannot be read or the axis is unknown — the caller treats that as "no constraint".
func (r *runner) currentGrade(gapID, dim string) string {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return ""
	}
	g := b.Gaps[gapID]
	if g == nil {
		return ""
	}
	d, ok := record.GradeDimensionOf(dim)
	if !ok {
		return ""
	}
	cur, known := g.GradeAt(d)
	if !known {
		return ""
	}
	return recordpb.Word(cur)
}

// gradeOtherThan picks a grade that is not the one given, so a dispute always asks for a change.
func (r *runner) gradeOtherThan(cur string) string {
	for i := 0; i < len(grades)*4; i++ {
		if g := r.g(); g != cur {
			return g
		}
	}
	return grades[0] // unreachable while len(grades) > 1; a stated fallback beats a silent loop
}

// verifyOutcomes is the value space of `lens verify --as` — what a source DID for the claim.
//
// It was `--trust high|medium|low`, all three of which mean the source SUPPORTS the claim. The
// negative half is what this list exists to drive: `refutes` and `absent` are the outcomes that
// make the assembly screen fire, and a fuzz that only ever generated supporting verdicts would
// leave that whole path unexercised while the coverage gate read green.
var verifyOutcomes = []string{"supports", "supports_with_bridge", "weak", "refutes", "absent", "unreachable"}

// verifyConfidence is the ORTHOGONAL axis: how sure red is of the outcome it just recorded.
// Driven independently of the outcome, because the pairs that matter most — `refutes` at low
// confidence, `supports` at low confidence — only exist if the two are drawn apart.
var verifyConfidence = []string{"high", "medium", "low"}

// someCitation returns a citation anchor on the record, or "" if blue has cited nothing yet —
// a REAL tool-assigned c-<hex>, the same discipline someFinding uses. `lens verify --anchor`
// refuses an id that names no citation, so a fabricated one would drive only the reject path.
func (r *runner) someCitation() string {
	// A SEAT ID, because `show` only exists inside a seat's tree. Any seat reads the same
	// projection; the lens is the one that acts on citations.
	out, err := r.exec("show", "evidence", "--seat-id", "red-lens-r1-L1")
	if err != nil {
		return ""
	}
	var e struct {
		Sources []struct {
			Anchor string `json:"anchor"`
		} `json:"sources"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &e) != nil || len(e.Sources) == 0 {
		return ""
	}
	return e.Sources[r.rng.Intn(len(e.Sources))].Anchor
}

// mint records a gap and returns the tool-assigned id (R<round>-N). The first mint of a run
// introduces the class; the rest reuse it.
func (r *runner) mint(seatID string) string {
	// #277: the KIND is randomised, not pinned. Pinning it to `document` would leave the
	// computation branch — a gap that CANNOT be closed without a proof answering it — driven by
	// nothing, which is the dead-drive class #276 measured (183/183 refusals on a verb that did
	// not exist, behind a green sweep). Both the refusal and the satisfied path run across a
	// sweep, because the prove drive below answers a computation gap when one is open.
	// THE SCENARIO IS DRAWN HERE, ONCE, and travels in required_fix.
	//
	// The fake used to decide everything downstream by coin flip, and nothing read
	// acceptance_check or required_fix — `grep -c` returned 0 — so the one question that
	// matters, DID THE RESPONSE SATISFY WHAT WAS ASKED, was never posed, because nothing knew
	// what had been asked. A directive fixes that without modelling a seat's judgement: red
	// draws the outcome at mint and every later seat READS IT BACK FROM THE BOARD. We model
	// COMMAND CHAINS, not minds — the seat boundary is only where the JS happens to cut. And
	// because the outcome is chosen rather than emergent, it is ASSERTABLE.
	directive := directives[r.rng.Intn(len(directives))]
	// A FORCED-UNVERIFIED RUN MINTS ONLY CONTESTED GAPS, and this is the condition the first
	// three attempts at this test each missed in turn. debate.js sits the bench only when the
	// docket is contested (`if (contested.length > 0)`), and the bench is what declares the
	// deadlock that makes the exit UNVERIFIED rather than CEILING. Left to the draw, a run whose
	// gaps all happened to be repairable never contested anything, the judge never sat, and the
	// run reached the round ceiling — which is a different verdict and a passing sweep saying so.
	// DISPUTE-LOST specifically: blue contests and red REJECTS, so the docket is contested AND
	// the gap stays open, which are the two things the deadlock arm reads.
	if r.forceUnverified {
		directive = dirDisputeLost
	}
	kind := checkKinds[r.rng.Intn(len(checkKinds))]
	if directive == dirProve || directive == dirProveDrifts {
		// The demand and the answer must agree: a gap settled by computing is minted as a
		// COMPUTATION check, so the tool's own close guard is in force alongside the scenario.
		kind = "computation"
	}
	args := []string{"--json", "mint", "--seat-id", seatID, "--problem", "fuzz problem", "--check-kind", kind,
		"--check", "acc", "--fix", directive, "--likelihood", r.g(), "--impact", r.g(),
		// --quote is the anchor ESTOPPEL keys on and the sweep never passed it; --key is
		// mint's crash-retry idempotency
		// handle; --reason is the prose channel. Four fields on the most consequential verb in
		// the tool, none of them ever exercised by a run.
		// A QUOTE FROM THE SEEDED REPORT, not a composed label. Since 0.63.0 a mint's --location
		// is matched against blue/report.md — the same rule `lens finding` has always been held
		// to — so `"§ fuzz — " + directive[:12]` is refused. The sweep went to ZERO gaps the
		// moment the check landed, and every gap-dependent verb (close, regrade, opinion,
		// closing, manifest-row, reproduce) followed it to zero. The coverage gate caught that;
		// the mint refusal alone would have looked like a quieter run.
		//
		// It quotes the ANCHOR sentence, not the cost sentence: the blue-edit drive swaps the
		// cost sentence between two phrasings, so a mint quoting it fails once an edit has run —
		// which is correct behaviour (red must quote text that is actually there) and useless as
		// a fixture. The anchor sentence is stable, and the invisible anchor layer spliced into
		// it is ignored by the match.
		"--quote", "A § fuzz sentence to anchor findings.",
		"--reason", "fuzz: the argument for raising this"}
	// COINING IS ITS OWN VERB (`merge class new`), so the fuzz drives it as one. It used to be
	// four flags on the first mint, which meant the coining path ran exactly once per run and
	// only ever in company with a mint.
	if !r.classMade {
		r.classMade = true
		// THE COIN'S FAILURE IS NOT OPTIONAL, and swallowing it made this file lie about what it
		// drove. `classMade` is set BEFORE the call and the error was discarded, so a coin that
		// failed for ANY reason left the flag latched and every later `--class fuzzcls` refused
		// with "unknown class" — a run that mints nothing, returns FAIL with an empty gaps array,
		// and is rejected by the engine as a degenerate merge. The cause was three verbs away
		// from the symptom and invisible from the log.
		if _, err := r.exec("class", "new", "--seat-id", seatID,
			// --neighbor names an EXISTING class, and is checked. `verification-gap` was not one;
			// nothing objected while the registry was absent, so the coining path ran green for
			// its whole life against a neighbour that did not exist.
			"--class", "fuzzcls", "--definition", "d", "--neighbor", "self-attestation", "--distinguisher", "q"); err != nil {
			r.noteApplyMiss("class new refused: " + err.Error())
			r.classMade = false // let a later seat try again rather than latching the run dead
		}
	}
	args = append(args, "--class", "fuzzcls")
	if r.coin(40) {
		// --key is mint's crash-retry idempotency handle: a retried mint under the same key
		// returns the first id rather than minting twice. Never driven.
		args = append(args, "--key", fmt.Sprintf("K%d", r.rng.Intn(3)))
	}
	// optional grade/lineage flags — exercised so mint's full flag surface runs, not the minimum.
	if r.coin(50) {
		args = append(args, "--severity", r.g())
	}
	if r.coin(50) {
		args = append(args, "--complexity", r.g())
	}
	if fl := r.someFinding(); fl != "" && r.coin(50) {
		args = append(args, "--found-by", fl) // the lens finding that surfaced it (real TOOL-assigned label)
	}
	if open := r.openGaps(); len(open) > 0 && r.coin(30) {
		args = append(args, "--supersedes", open[r.rng.Intn(len(open))]) // lineage: this gap replaces an ancestor
	}
	// #267 stage 3: a CONCRETE proposed fix, which the tool validates against the live report
	// and which DERIVES fix_basis: verified. It swaps the seeded edit-target sentence, exactly
	// as blue's own edit drive does, so a legal pair exists whichever way the last edit left
	// the file — and both branches (proposal present / prose only) run across the sweep.
	if r.coin(40) {
		if cur, err := os.ReadFile(filepath.Join(r.runDir, "blue", "report.md")); err == nil {
			fixOld, fixNew := "rising over time", "climbing sharply"
			if !strings.Contains(string(cur), fixOld) {
				fixOld, fixNew = fixNew, fixOld
			}
			if strings.Contains(string(cur), fixOld) {
				args = append(args, "--quote", fixOld, "--new", fixNew)
			}
		}
	}
	out, err := r.exec(args...)
	if err != nil {
		return ""
	}
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil {
		return ""
	}
	if kind == "computation" && env.Result.GapID != "" {
		if r.computationGaps == nil {
			r.computationGaps = map[string]bool{}
		}
		r.computationGaps[env.Result.GapID] = true
	}
	return env.Result.GapID
}

// someFinding returns a random lens finding label on the record, or "" if none — feeds mint's
// --found-by with a real TOOL-assigned label (L{role}-F{N}) rather than a fabricated one.
func (r *runner) someFinding() string {
	out, err := r.exec("show", "findings", "--seat-id", "red-merge-r1")
	if err != nil {
		return ""
	}
	var f struct {
		Findings []struct {
			Label string `json:"label"`
		} `json:"findings"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &f) != nil || len(f.Findings) == 0 {
		return ""
	}
	return f.Findings[r.rng.Intn(len(f.Findings))].Label
}

func (r *runner) openGaps() []string {
	out, err := r.exec("show", "board", "--seat-id", "red-merge-r1")
	if err != nil {
		return nil
	}
	var b struct {
		Open []struct {
			ID string `json:"id"`
		} `json:"open"`
	}
	if json.Unmarshal([]byte(out), &b) != nil {
		return nil
	}
	var ids []string
	for _, g := range b.Open {
		ids = append(ids, g.ID)
	}
	return ids
}

func (r *runner) closeGap(seatID, id string, allowReg bool) {
	// A COMPUTATION CHECK IS SATISFIED BEFORE IT IS CLOSED, deterministically.
	//
	// The guard refuses a closure with no proof answering the gap, so leaving it to chance
	// meant unclosable gaps piling up run after run — and every drive that iterates the open
	// set grew with them. Answering it here drives the SATISFIED path on every computation
	// gap, which is the branch a probabilistic prove would mostly skip, and the refusal is
	// still covered by the integration tests that assert it.
	if r.computationGaps[id] {
		name := "fuzz-answer-" + id + ".js"
		if err := os.WriteFile(filepath.Join(r.runDir, name), []byte("console.log('answers "+id+"');"), 0o644); err == nil {
			_, _ = r.exec("prove", "--seat-id", "blue-respond-r1", "--quote", "§ fuzz",
				"--script", name, "--answers", id, "--reason", "fuzz: settling the computation check")
		}
	}
	// A regression close carries lineage forward: it mints a successor and closes WITH it
	// (record.go requires --successor for repaired_with_regression). Only allowed on the first close
	// pass, so the successors it spawns are plain-closed on a later pass and the loop terminates.
	if allowReg && r.coin(35) {
		if succ := r.mint(seatID); succ != "" {
			if _, err := r.exec("close", "--seat-id", seatID, "--id", id, "--as", "repaired_with_regression",
				"--superseded-by", succ, "--reason", "fuzz regression close", "--verified-by", seatID, "--verified-with", "fuzz", "--verified-against", "rec"); err == nil {
				return
			}
		}
	}
	// A CARRY restates last round's verification rather than asserting a fresh act, so it is its
	// own verb and takes --carried-from instead of the verification triple.
	//
	// AND IT CARRIES A GAP THE RECORD ALREADY CLOSED. Driving it against `id` — the gap this call
	// is closing for the FIRST time — was refused every time, because a carry of a gap with no
	// prior closure is a laundering path the record exists to refuse. 20 of 20 refused, and the
	// coverage line read as a driven verb.
	// UNCONDITIONAL WHEN THERE IS SOMETHING TO CARRY. Behind a 30% coin on a precondition that
	// needs a PRIOR round's closure, this fired ONCE across 60 runs — and that one was refused,
	// so the verb's real coverage was zero while the tally read 1. What it was hiding: `merge
	// carry` could not run without --reason at all, because `Close.prose` was annotated
	// unconditionally required and refused before validate's carry exemption could execute. A
	// documented invocation was broken for as long as the drive was too thin to say so.
	// A PRIOR ROUND'S CLOSURE, which is the only thing a carry can restate. `--carried-from`
	// names the round, and a close keys on gap_id, so carrying a gap closed in THIS sitting is a
	// duplicate the record refuses.
	round, _ := record.RoundOf(seatID)
	if prior := r.closedInARoundBefore(round); len(prior) > 0 {
		carried := prior[r.rng.Intn(len(prior))]
		carry := []string{"carry", "--seat-id", seatID, "--id", carried,
			"--carried-from", strconv.Itoa(round - 1), "--as", "repaired"}
		// THE EXEMPTION ITSELF, half the time. A carry restates a closure an earlier round already
		// argued, so it owes no fresh argument — and nothing drove the no-reason form, which is
		// how the annotation deleted it silently. Both shapes are legal and both are driven.
		if r.coin(50) {
			carry = append(carry, "--reason", "fuzz: carried from the prior round")
		}
		// A carry names where the remainder went as often as a close does; the flag is on both
		// verbs and was driven on neither once the two split. The successor is MINTED here for
		// the same reason the regression close mints one: it must be a real, still-open gap, and
		// the record refuses a dead-end forwarding address.
		if r.coin(50) {
			if succ := r.mint(seatID); succ != "" {
				carry = append(carry, "--superseded-by", succ)
			}
		}
		_, _ = r.exec(carry...)
	}
	// THE WHOLE CLOSURE VOCABULARY, not just `closed`. #342 closed the set, so the
	// enum-coverage sweep now demands every value be reached — and three of them
	// (not_a_defect, defect_accepted, defect_owed_elsewhere) had never been driven by
	// anything, on either closing verb, in the tool's life.
	//
	// `amends_prior` takes --supersedes: it names a defect found BETWEEN two repairs that each
	// closed clean earlier, so it needs a prior closure to amend.
	// 20% ON TOP OF "a prior closure exists" left `close --as amends_prior` never driven across
	// 60 runs. The precondition is already the rare part.
	if prior := r.closedGapIDs(); len(prior) > 0 {
		if _, err := r.exec("close", "--seat-id", seatID, "--id", id, "--as", "amends_prior",
			"--supersedes", prior[r.rng.Intn(len(prior))], "--reason", "fuzz: found between two clean repairs",
			"--verified-by", seatID, "--verified-with", "fuzz", "--verified-against", "rec"); err == nil {
			return
		}
	}
	as := pick(r.rng, []string{"repaired", "not_a_defect", "defect_accepted", "defect_owed_elsewhere"})
	_, _ = r.exec("close", "--seat-id", seatID, "--id", id, "--as", as, "--reason", "fuzz close as "+as,
		"--verified-by", seatID, "--verified-with", "fuzz", "--verified-against", "rec")
}

// checkKinds draws uniformly: closeGap satisfies a computation gap before closing it, so the
// kind no longer decides whether a run accumulates dead gaps.
// directives are the scenario vocabulary a mint draws from. The string IS the expected
// behaviour of every seat downstream, and the oracle asserts the record shows it happened.
const (
	dirApply       = "FUZZ-APPLY: blue applies the proposed pair verbatim; red closes it"
	dirCounter     = "FUZZ-COUNTER: blue edits, but not the proposed text; red closes it"
	dirDisputeWon  = "FUZZ-DISPUTE-WON: blue contests the grade and red accepts the regrade"
	dirDisputeLost = "FUZZ-DISPUTE-LOST: blue contests the grade and red rejects it"
	dirIgnore      = "FUZZ-IGNORE: blue does nothing; the gap stays open"
	// dirProve is the COMPUTATION CHAIN, end to end. It is the only axis where the tool
	// re-executes evidence rather than re-reading it, and it was the least represented: `blue
	// prove` fired from a 35% probe, `lens reproduce` picked ANY proof on the record at 35%
	// with no connection to the gap being audited, and red closed the gap BEFORE any of that.
	// The order was inverted — reproduction decorated a decision already taken, when it is the
	// thing that should EARN it. Nothing would have noticed if reproduce returned garbage.
	dirProve = "FUZZ-PROVE: blue settles it by computing; red re-runs the proof and closes only if it holds"
	// dirProveDrifts is the case that makes the re-run MEAN something. A deterministic script
	// always reproduces, so with only dirProve the scenario check and the tool's own
	// computation guard are indistinguishable — probing by deleting the check caught NOTHING.
	// A drifting script (an unseeded random, a live sample) is graded `observed` rather than
	// `reproducible`, and red must NOT close on it: that is the whole difference between a
	// proof and a measurement, and it is the branch a fake with only tidy scripts never shows.
	dirProveDrifts = "FUZZ-PROVE-DRIFTS: blue computes, but the output moves between runs; red cannot close on it"
)

// WEIGHTED so a run can still reach PASS. Every fate is now DERIVED from the directive, so a
// board whose gaps are all IGNORE correctly never closes and the run correctly ends CEILING —
// which is right, and would leave VERIFIED uncovered if the draw were uniform. Four of seven
// draws are satisfiable, which keeps both terminal states in the sweep.
var directives = []string{dirApply, dirApply, dirCounter, dirCounter, dirDisputeWon, dirDisputeLost, dirIgnore, dirProve, dirProve, dirProveDrifts}

// satisfied reports whether a directive means the gap is repaired and red should close it.
func satisfied(directive string) bool { return directive == dirApply || directive == dirCounter }

// contested reports whether the directive routes through a grade dispute rather than a repair.
func contested(directive string) bool {
	return directive == dirDisputeWon || directive == dirDisputeLost
}

// outcomeRe reads the verdict debate.js states in the assembler's prompt.
var outcomeRe = regexp.MustCompile(`Debate outcome: ([A-Z]+)`)

var checkKinds = []string{"document", "computation", "source"}

// `deferred` was added as a fate in #246 and never driven — a value the tool accepts, the
// registry declares, and no run has ever recorded.
var inquiryStatus = []string{"proposed", "pursued", "abandoned", "declined", "deferred"}

// nextInquiryStatus round-robins the undirected line of inquiry's fate so every value is driven BY
// CONSTRUCTION rather than by luck.
//
// It was a uniform random pick behind a 30% branch — a ~6% draw per line of inquiry for any one fate —
// so the enum-coverage gate passed on most seeds and failed the moment an unrelated change
// shifted the RNG stream. A gate whose verdict depends on the draw is sampling coverage, not
// measuring it, and the failure it produces looks like the change's fault rather than its own.
//
// The counter is PACKAGE-LEVEL and atomic, not a field on runner. Per-run it would be worse than
// random: the sweep makes about 1.6 undirected inquiries per run, so a counter resetting each time
// would only ever reach index 0 and 1 and the last three fates would never be driven at all.
// Atomic because the sweep's runs are concurrent.
var inquiryTick atomic.Int64

func nextInquiryStatus() string {
	return inquiryStatus[int(inquiryTick.Add(1)-1)%len(inquiryStatus)]
}

var obsKind = []string{"reason", "checked-held"}

// disputeDims is the full grade-dimension domain — the fuzz must contest each, not only impact.
var disputeDims = []string{"severity", "likelihood", "impact", "complexity"}

// petitionClasses is the full petition-class domain (debate.js PETITIONS enum).
var petitionClasses = []string{"ethical", "safety", "integrity", "constitutional"}

// regradeDims are the grade axes a regrade can move (each maps to its flag below).
var regradeDims = []string{"severity", "likelihood", "impact", "cx"}

func pick[T any](rng *lockedRand, xs []T) T { return xs[rng.Intn(len(xs))] }

// extras fires a RANDOM subset of a role's REMAINING verb surface, so no two fuzz paths
// look alike and every seat exercises far more than its happy path. Only verbs that keep
// the oracles intact (verify passes, the run terminates, the report renders) are here;
// terminal/verdict-shaping acts (halt, certify, outcome, verdict) stay in the structured
// cases above. Reference-taking verbs are gated on a referent existing.
func (r *runner) extras(role, seatID string, open []string) {
	// BOTH ARMS OF THE CHANNEL. --none is the explicit negative a sitting closes with when
	// nothing blocked it, and it writes a DIFFERENT event type — so a sweep that only ever
	// drove the complaint arm would leave the attestation path unexercised, which is the
	// arm the whole change exists to make observable.
	//
	// EXCLUSIVE, AND ONE OF THEM ALWAYS FIRES. These were independent coins at 50% and 30%, so a
	// sitting could close with NEITHER — which the verb's own help forbids ("CLOSE THIS CHANNEL
	// EVERY SITTING"), and a drive that contradicts the contract it exercises is testing a system
	// nobody ships. It also left `bench friction --none` never passed across 60 runs: the bench
	// sits rarely (5 frictions in the whole sweep), and 30% of rarely is a path the coverage gate
	// reports as missing while the drive is right there.
	if r.coin(60) {
		r.do("friction", seatID).set("--reason", "fuzz friction from "+seatID).run()
	} else {
		r.do("friction", seatID).bare("--none").set("--reason", "fuzz: nothing blocked "+seatID).run()
	}
	// line of inquiry carries an optional --method; feed it sometimes so that flag is exercised too.
	// #246: a line of inquiry now has an id and a LIFECYCLE. Propose, then sometimes move it — the
	// move is the path the old one-shot append could not record at all (measured: 0 of 86
	// events across six runs ever changed status).
	inquiry := func(role string) {
		// A DECLINED OR ABANDONED line of inquiry requires --reason (record.go: an unexplained
		// non-pursuit is the decoration this verb exists to refuse). Without it two of the
		// three statuses were rejected on every call, so only `pursued` ever reached the
		// record while the verb gate read as covered. Found by the execution tally
		// (blue line-of-inquiry: 48 of 72 calls refused).
		// A FATE-CARRYING LINE IS BORN `proposed`, so the ruling cycle can run on it: red rules
		// it, blue answers. Creating one directly as `pursued` or `declined` skips the cycle
		// entirely — which the oracle caught, reporting endorsed inquiries that "ended declined"
		// when they had simply never been through a ruling at all.
		//
		// An UNDIRECTED line keeps the random status: proposing something already declined is a
		// real shape (blue weighed it and did not start), and it carries no fate for red to
		// rule from, so the ruling cycle and the oracle both skip it by construction.
		line, st := pick(r.rng, inquiryFates)+" ("+seatID+")", "proposed"
		if r.coin(30) {
			line, st = "fuzz undirected line of inquiry "+seatID, nextInquiryStatus()
		}
		// PROPOSE AND MOVE ARE TWO VERBS. A proposal is born `proposed` — the tool supplies it,
		// so there is no status to pass — and any other fate is reached by MOVING the line it
		// proposed, which is the cycle red rules on and the oracle reads.
		out, err := r.do("line-of-inquiry propose", seatID).set("--reason", line).
			set("--hypothesis", "fuzz: what would be true if "+seatID+" paid off").
			on(50, "--method", "fuzz-method").run()
		// A MOVE BACK TO `proposed` IS A REAL STATE — the line is still open and the seat is
		// saying so rather than settling it — so the fate is drawn from the whole set, not from
		// the set minus its default.
		if err == nil {
			if id := inquiryIDOf(out); id != "" {
				r.do("line-of-inquiry move", seatID).set("--id", id).set("--as", st).
					set("--reason", "fuzz: what changed this line's fate").run()
			}
		}
	}
	switch role {
	case "lens":
		// NO line of inquiry DRIVE HERE: the lens role has no line of inquiry verb (register/finding/
		// cite/friction/show). This called it 183 times per sweep, every one refused, while
		// the verb gate stayed green on blue's line of inquiry events — a dead drive that read as
		// coverage. Found by the execution tally (lens line-of-inquiry: 183 of 183 refused).
	case "merge":
		// THE PER-ROUND REVIEW OF THE REPORT'S ACCOUNT OF ITS OWN RESEARCH. The verb is
		// `inquiry-support` and the event it writes is `inquiry_review` — the names differ, which
		// is why deleting the retired per-line VOTE took this verb's only drive with it. It names
		// no line and casts no verdict: a shortfall in the body is an ordinary gap, so `--reason`
		// is its whole payload.
		// UNCONDITIONAL, BECAUSE IT IS A PER-ROUND DUTY AND A HARD PRECONDITION FOR PASS.
		//
		// This sat behind a 40% coin, so three rounds in five could not pass: the tool refuses a
		// PASS with no review for the round, and debate.js refuses a FAIL that names no gaps, so
		// the seat had NO legal verdict and the run wedged. Before the drive honoured the refusal
		// that was invisible — the error was discarded and the harness was told PASS anyway.
		//
		// A coin is the wrong shape for a duty. What varies between runs is what the review FINDS,
		// not whether the read happened.
		r.do("inquiry-support", seatID).
			set("--reason", "fuzz: read the report against the lines on the record; the treatment matches").run()
		// ONE MOTION DRIVER, AND THERE WERE BRIEFLY TWO.
		//
		// The additive stage added a blue-files/merge-rules pair here beside the `blue dispute`
		// scenario driver, because the old verb was still live and both had to be exercised. When
		// the old verb went, the scenario driver became a motion driver too — and the two minted
		// ids independently while tracking them in separate lists. They crossed: M2 was ruled
		// twice, once by each, and the second ruling OVERWROTE the first at replay. The report
		// showed M2 answered with the reasoning written for another gap.
		//
		// The rulings live in answerDisputes now, where the ask is, so the id never leaves the
		// bookkeeping that minted it.
		// W1.8 archive spot-check — the merge's duty, and now ENFORCED from the board
		// (verify.archiveSpotCheckFloor).
		//
		// THE OLD DRIVER MODELLED A NON-COMPLIANT SEAT AND NOTHING NOTICED. It fired at 45%, so
		// most rounds skipped the duty outright, and its comment claimed "--none is always valid
		// (an empty archive at round start)" — which is false whenever the archive is not empty,
		// and is EXACTLY the self-attestation the duty exists to prevent. Both passed because
		// the event had no reader.
		//
		// It now models a seat that discharges the duty honestly: sample when there is something
		// to sample, and claim emptiness only when the board agrees.
		if closed := r.closedGapIDs(); len(closed) > 0 {
			r.do("spot-check", seatID).set("--ids", closed[r.rng.Intn(len(closed))]).
				set("--reason", "fuzz: re-read the closure record; the anchor still resolves").run()
		} else {
			r.do("spot-check", seatID).bare("--none").
				set("--reason", "fuzz: the archive was empty at round start").run()
		}
		// RED RULES ON BLUE'S DIRECTIONS (#246) — the verb red never had. Across six runs blue
		// rejected 18 of its own 86 inquiries and red rejected none, because it could not.
		// RED RULES EVERY UNRULED LINE OF INQUIRY, and the ruling follows the line it was proposed as.
		//
		// It used to rule ONE random line of inquiry at 45% with a random ruling, so a ruling had no
		// relation to what was proposed and no consequence for what happened next: an
		// out-of-scope line could be pursued and an endorsed one abandoned, and the record kept
		// both halves while joining them nowhere. The line of inquiry's own --line now carries its
		// intended fate, exactly as a gap's --fix carries its scenario.
		r.ruleOpenInquiries(seatID)
		// near-match is the screen red runs BEFORE minting, to catch a reopen. Read-only, so it
		// left no event and no gate saw it — while the merge prompt calls it every round.
		r.maybe(40, func() {
			r.readOnly("near-match", seatID, "--problem", "fuzz candidate problem text for screening", "--quote", "§ fuzz")
		})
	case "blue":
		r.maybe(45, func() { inquiry("blue") })
		// APPEAL WHAT HAS ALREADY BEEN RULED — a round later than the filing, which is when an
		// appeal is possible at all.
		for _, id := range r.ruledMotions {
			if r.coin(40) {
				_, _ = r.exec("motion", "grade", "appeal", "--seat-id", seatID, "--id", id,
					"--reason", "fuzz: pressing it to the bench")
			}
		}
		r.ruledMotions = nil
		// NO RANDOM STATUS MOVE HERE. answerInquiryRulings owns moves now: it answers red's
		// ruling, comply or contest, one decision per inquiry. A second writer sliding statuses
		// at random fought it — the oracle caught it immediately, 24 of 60, reporting endorsed
		// inquiries that ended declined and contests the record never recorded because the move
		// landed before the ruling did. Two writers disagreeing about a fate is the same defect
		// the edit drive had, one axis over.
		// THE CITATION AXIS (#256), driven end to end through the real binary: `blue cite` fetches
		// the source through the run cache and splices an INVISIBLE, IMMORTAL <!--cite:c-…--> anchor
		// at the quoted sentence. --quote must appear VERBATIM in blue/report.md, so it quotes the
		// seeded "§ fuzz" text the same way lens finding does. A --key from a small space exercises
		// the retry short-circuit; a shared URL across seats exercises the cache HIT path (fetch-once),
		// while the per-seat path exercises the MISS path.
		r.maybe(55, func() {
			url := sourceURL("/blue-shared")
			if r.coin(50) {
				url = sourceURL("/" + seatID)
			}
			r.do("cite", seatID).
				set("--quote", "§ fuzz").
				set("--url", url).
				set("--title", "fuzz source "+seatID).
				on(50, "--quote", "fuzz cited claim "+seatID).
				on(40, "--key", fmt.Sprintf("C%d", 1+r.rng.Intn(2))).
				on(50, "--reason", "fuzz: why this source backs the claim").
				run()
		})
		// An UNREACHABLE source is an unusable citation: the cite must be REJECTED and the failure
		// auto-logged as friction. Driving it here proves the reject path never wedges the run.
		r.maybe(15, func() {
			r.do("cite", seatID).
				set("--quote", "§ fuzz").
				set("--url", "http://127.0.0.1:1/unreachable").
				set("--title", "unreachable "+seatID).
				run()
		})
		// THE ONLY WRITE PATH to blue/report.md, driven end to end through the real binary.
		// It swaps the seeded edit-target sentence between two phrasings, so a valid unique
		// --quote exists whichever way the previous edit left the file — no state to thread.
		//
		// --answers carries the PROVENANCE (#267): the gap this edit responds to. Sent when
		// the board has an open gap, omitted otherwise, so BOTH validation branches run —
		// the reference check and the legal no-gap edit. The --reason deliberately names no
		// gap: prose naming a real gap with --answers empty is REFUSED, and this drive must
		// exercise the path that is allowed to succeed.
		r.maybe(45, func() {
			cur, err := os.ReadFile(filepath.Join(r.runDir, "blue", "report.md"))
			if err != nil {
				return
			}
			oldSpan, newSpan := "rising over time", "climbing sharply"
			if !strings.Contains(string(cur), oldSpan) {
				oldSpan, newSpan = newSpan, oldSpan
			}
			if !strings.Contains(string(cur), oldSpan) {
				return // an anchor landed mid-span; skip rather than force a mis-quote
			}
			// THIS DRIVE NO LONGER CLAIMS TO ANSWER A GAP.
			//
			// It used to pick a random open gap for --answers, and the scenario oracle caught it
			// on its first run: 5 of 60 seeds attached provenance to a gap whose directive said
			// blue does nothing. Two writers disagreed about who had repaired what — which is
			// exactly the defect --answers exists to make impossible, reproduced inside the fake.
			//
			// Answering a gap is blueRespondTo's job, one decision per gap, from the directive.
			// What remains is an UNATTRIBUTED edit: blue sharpening its own prose, which is real
			// and must stay legal (--answers' own help says to omit it when no gap is answered).
			r.do("edit", seatID).
				set("--quote", oldSpan).
				set("--new", newSpan).
				set("--reason", "fuzz edit: sharper phrasing, answering no gap").
				on(40, "--key", fmt.Sprintf("E%d", 1+r.rng.Intn(2))).
				run()
		})
		// READ-ONLY, PROMPT-CALLED, AND PREVIOUSLY UNFUZZED. claim-index records nothing, so the
		// event-type coverage gate cannot see it — yet blue's response prompt tells blue to call
		// it when propagating a correction to every site of a claim. A crash here costs a real
		// round; no gate would have noticed.
		// #277: settle a claim by COMPUTING it. The script is written into the run dir and the
		// tool runs it twice — so this drives the real interpreter, the cache, the anchor
		// splice and the reproducible/observed grading, not a stub.
		r.maybe(35, func() {
			name := "fuzz-proof-" + seatID + ".js"
			body := "console.log('fuzz proof for " + seatID + "');"
			if r.coin(25) {
				body = "console.log(Math.random());" // exercises the OBSERVED grade
			}
			if err := os.WriteFile(filepath.Join(r.runDir, name), []byte(body), 0o644); err != nil {
				return
			}
			pv := r.do("prove", seatID).
				set("--quote", "§ fuzz").
				set("--script", name).
				set("--reason", "fuzz: computing rather than arguing").
				on(40, "--key", fmt.Sprintf("P%d", 1+r.rng.Intn(2)))
			// ANSWER a real gap most of the time: a computation check closes only on a proof
			// that names it, so without this the guard would only ever be seen refusing.
			if len(open) > 0 && r.coin(70) {
				pv = pv.set("--answers", pick(r.rng, open))
			}
			pv.run()
		})
		r.maybe(35, func() { r.readOnly("claim-index", seatID) })
		r.maybe(40, func() { r.do("revision", seatID).set("--reason", "fuzz revision").run() })
		r.maybe(30, func() {
			// RETIRE WHAT WAS ACTUALLY REMOVED. This used to retire "fuzz claim <seat>" — a
			// string that was never in the report — 45 times a sweep. A phantom retire is not
			// merely uninformative: the scorecard computes unrecorded_claim_loss as the drop in
			// claim_count MINUS the retire events, so retiring something that was never there
			// cancels real loss and blinds the detector built to catch silent deletion.
			//
			// A real retirement names text a recorded edit took out, which is what makes the
			// removal something the record can SHOW rather than something a seat says.
			if claim := r.recentlyEditedOut(); claim != "" {
				r.do("retire", seatID).set("--quote", claim).set("--reason", "fuzz: the claim went with the edit").
					on(50, "--new", "fuzz replacement claim").run()
			}
		})
		// THE CORRECTNESS MANIFEST: one row per repaired gap, on the record (#318).
		//
		// It used to fire at 40% for ONE random gap, which modelled a seat that mostly skipped
		// its own self-audit — and nothing noticed, because the row had no reader and the
		// coverage metric was counting the ENVELOPE array instead. The verb is now what the
		// prompt asks for and what the scorecard counts, so the fake discharges it the way a
		// compliant blue would: every gap it is repairing this round.
		for _, id := range open {
			r.do("manifest-row", seatID).set("--id", id).
				set("--reason", "fuzz: figures recomputed, universals enumerated, acceptance check run — held").
				on(50, "--reason", "fuzz: the receipt's argument").run()
		}
	case "bench":
		// run-end certification statement — the bench's, additive (a recorded opinion).
		r.maybe(45, func() {
			r.do("certify", seatID).set("--reason", "fuzz certification statement from "+seatID).run()
		})
		// A DECLARATION MOVES NO GAP, which is why it takes no --id and why the coverage walk
		// would otherwise never reach it: every other bench act names something.
		r.maybe(35, func() {
			r.do("declare", seatID).
				set("--reason", "fuzz holding: a term is construed this way for the rest of the run").run()
		})
	}
}

// arr / obj are goja-friendly envelope builders.
func arr(v ...any) []any { return append([]any{}, v...) }

// envelopeFor performs the seat's tool acts and returns the envelope debate.js expects. The
// randomisation here is what drives debate.js's control flow: red's verdict and gap count
// decide whether the loop continues, deadlocks, or terminates.
func (r *runner) envelopeFor(seatID, prompt string) map[string]any {
	switch {
	case strings.HasPrefix(seatID, "blue-synthesize"):
		r.register("blue", seatID)
		r.extras("blue", seatID, nil)
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "petitions": r.maybePetition("blue", seatID), "friction": arr()}

	case strings.HasPrefix(seatID, "red-merge"):
		r.register("merge", seatID)
		// RED NOW EVALUATES THE REPAIR AGAINST WHAT IT ASKED FOR.
		//
		// It used to coin PASS at 40%, mint 1-3 fresh gaps unconditionally, and close every
		// open gap on a PASS regardless of what blue had done — so the verdict bore no relation
		// to whether anything was actually repaired. Each gap now carries the scenario it was
		// minted with, and red reads it back off the board: a repaired gap CLOSES, a contested
		// or ignored one STAYS OPEN. The round's verdict falls out of what is left.
		responses := r.answerDisputes(seatID)
		r.extras("merge", seatID, r.openGaps())

		if r.evaluated == nil {
			r.evaluated = map[string]bool{}
		}
		for _, id := range r.openGaps() {
			d := r.scenarioOf(id)
			if r.presented[id] {
				r.evaluated[id] = true
			}
			switch {
			case satisfied(d):
				// ONE COIN, NOT TWO. This was `r.coin(25)` here AND `r.coin(25)` inside closeGap,
				// so the regression close needed 6.25% of one branch AND a successful mint — and
				// `close --as repaired_with_regression` was reported NEVER DRIVEN across 60 runs,
				// alongside its `--superseded-by`. Coins that multiply read as "sampled" and
				// behave as "off"; the sampling now happens in exactly one place.
				r.closeGap(seatID, id, true)
			case d == dirProve && r.reproduced[id]:
				// THE REPRODUCTION EARNS THE CLOSE. Red does not take blue's word that a
				// computation happened, and does not take its own: a lens re-ran the recorded
				// script and got the same bytes. Without that, the gap stays open — which is
				// also what the tool's own computation guard would enforce.
				r.closeGap(seatID, id, false)
			case d == dirDisputeWon:
				// Red accepted the regrade in answerDisputes; the substance is settled, so the
				// gap closes on the corrected grade.
				r.closeGap(seatID, id, false)
			}
			// dirDisputeLost and dirIgnore: deliberately left open. The first goes to the
			// bench's docket, the second is blue owing work it did not do.
		}

		// MINT BEFORE JUDGING WHETHER ANYTHING REMAINS. Round 1 has an empty board until red
		// puts something on it — evaluating PASS first made every run pass in one round with
		// nothing ever raised, which the coverage gate caught immediately.
		// 0 is legal at round 1: red auditing and raising NOTHING is a real shape, and it is the
		// only way a run reaches PASS in a single round.
		fresh := r.rng.Intn(4)
		if !strings.HasSuffix(seatID, "-r1") {
			fresh = r.rng.Intn(2) // later rounds may raise nothing
		}
		for range fresh {
			r.mint(seatID)
		}

		// RED'S PER-LINE SUPPORT VERDICT USED TO BE DRIVEN HERE, before the round verdict,
		// because `verdict --as PASS` was refused while any line was unvoted. Both the verb and
		// that gate are retired: a line's treatment is an ordinary gap now, so the PASS gate is
		// the gap board and nothing else. See the note where voteInquirySupport was.
		open := r.openGaps()
		if len(open) == 0 {
			r.dialectic("merge", seatID, nil)
			// THE TOOL DECIDES WHETHER THIS IS A PASS, NOT THE SEAT'S OWN GAP COUNT.
			//
			// An empty board is NOT the whole PASS gate: requirePassClosesAllGaps also refuses over
			// an unruled MOTION, which is the case a seat cannot see by counting gaps. This drive
			// discarded the refusal and told the harness `PASS` anyway — so debate.js walked on to
			// `bench outcome --as verified` while the record held NO verdict event at all, and the
			// outcome landed with basis `asserted` because DeriveVerdict had nothing to derive from.
			// Six of sixty seeds, and the fuzz reported them as a derivation failure.
			//
			// The refusal is the production behaviour under test. Honouring it means a clean board
			// with an outstanding motion FAILs the round — which is exactly right, and is the only
			// way the motion arm of that gate is ever exercised end to end.
			// RED DOES NOT PASS IN A FORCED-UNVERIFIED RUN, and that is not a thumb on the
			// scale — debate.js's verdict is `redPASS ? VERIFIED : ceilingUnaudited ? CEILING :
			// UNVERIFIED`, so a run whose red ever passes CANNOT reach the value this drives.
			// Leaving it to the docket is what made the first version of this test flake: with
			// concurrent seats the interleaving decides the draws, so a seeded run no longer
			// reproduces its own board, and the run reached VERIFIED on the second attempt.
			// A terminal path driven on purpose has to be forced at every condition the engine
			// reads, not at one of them.
			// AND IT FAILS OVER SOMETHING REAL. Skipping the PASS branch alone dropped through
			// to the refusal arm below, which returns FAIL with an EMPTY gaps array — and the
			// engine refuses that as a degenerate merge, correctly. On a cleared board there is
			// nothing left to fail over, so this mints one: red keeps the round open the only
			// way the protocol allows, and the gap it mints is also what keeps the board
			// non-empty for the deadlock exit the bench declares later.
			if r.forceUnverified {
				if id := r.mint(seatID); id != "" {
					_, _ = r.exec("verdict", "--seat-id", seatID, "--as", "FAIL")
					return map[string]any{"verdict": "FAIL",
						"gaps":              []any{map[string]any{"id": id, "supersedes": arr()}},
						"closures":          arr(),
						"dispute_responses": responses,
						"petitions":         r.maybePetition("merge", seatID), "friction": arr()}
				}
			}
			if _, err := r.exec("verdict", "--seat-id", seatID, "--as", "PASS"); err == nil {
				return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "petitions": r.maybePetition("merge", seatID), "friction": arr()}
			}
			// Refused over something that is not a gap. Record the verdict the tool WILL take, so
			// the record and the harness agree about how this round ended.
			_, _ = r.exec("verdict", "--seat-id", seatID, "--as", "FAIL")
			return map[string]any{"verdict": "FAIL", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "petitions": r.maybePetition("merge", seatID), "friction": arr()}
		}

		// Something is unrepaired, so the round FAILs.
		r.dialectic("merge", seatID, open)
		var gaps []any
		for _, id := range open {
			gaps = append(gaps, map[string]any{"id": id, "supersedes": arr()})
		}
		// The merge's terminal act on a FAIL too: the checkpoint is what protects the event log
		// from a stray git operation mid-round, and it was only ever driven on a PASS — so the
		// `FAIL` half of a two-value enum had never been recorded by anything.
		_, _ = r.exec("verdict", "--seat-id", seatID, "--as", "FAIL")
		return map[string]any{"verdict": "FAIL", "gaps": gaps, "closures": arr(), "dispute_responses": responses, "petitions": r.maybePetition("merge", seatID), "friction": arr()}

	case strings.HasPrefix(seatID, "blue-respond"):
		r.register("blue", seatID)
		open := r.openGaps()
		r.dialectic("blue", seatID, open) // blue's position and closings
		r.extras("blue", seatID, open)
		// ONE DECISION PER GAP, TAKEN FROM THE GAP. Each open gap carries the scenario it was
		// minted with; blue reads it back off the board and does what it says. This replaces a
		// blanket 40%-per-gap dispute roll plus a 50% verbatim-apply coin, neither of which
		// looked at what red had actually asked for.
		r.answerInquiryRulings(seatID)
		r.disputedThisRound = nil
		if r.presented == nil {
			r.presented = map[string]bool{}
		}
		for _, id := range open {
			r.presented[id] = true
		}
		r.blueRespondTo(seatID, open)
		disputes := r.disputedThisRound
		// debate.js rejects an EMPTY manifest on a round with open gaps — a repair must show its
		// receipt. #318: the envelope is a ROUTING REF now, gap ids only; the rows themselves are
		// on the record, written by the manifest-row calls above against this same `open` set.
		var manifest []any
		for _, id := range open {
			manifest = append(manifest, id)
		}
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "manifest": manifest, "grade_disputes": disputes, "petitions": r.maybePetition("blue", seatID), "friction": arr()}

	case strings.HasPrefix(seatID, "judge-petition"):
		r.register("bench", seatID)
		return r.rulePetitions(seatID) // rule every pending petition (petition-rule events + envelope rulings)

	case strings.HasPrefix(seatID, "judge"): // adjudication + terminal
		r.register("bench", seatID)
		r.extras("bench", seatID, nil)
		// THE BENCH RULES ON WHAT HAPPENED, not on a coin. A gap reaches the docket because
		// its scenario left it open, and the scenario says why: a LOST dispute is a contest
		// red refused and the bench settles it; an IGNORE is blue owing work, which is what
		// `carried` means — "the material needs another round", stated as a decision.
		var res []any
		for _, id := range r.openGaps() {
			disp := "carried"
			// A CARRIED GAP IS STILL A RULING, and the bench records it. The fake used to put
			// `carried` in the envelope and write NO opinion event, so the bench's most common
			// act — deferring a gap with a stated direction — had never been driven once. The
			// enum-coverage gate found it the moment #342 closed the disposition set.
			if r.scenarioOf(id) != dirDisputeLost {
				_, _ = r.exec("opinion", "--seat-id", seatID, "--id", id, "--as", "carried",
					"--principle", "thoroughness", "--tension", "cost", "--review-flag", "false", "--settled", "the proposition this ruling bars", "--final",
					"--reason", "fuzz: carried with a stated direction for "+id)
			}
			if r.scenarioOf(id) == dirDisputeLost {
				// EVERY CLOSING DISPOSITION, not just `closed` (#342). The bench shares red's
				// closure vocabulary now, and the sweep must reach all of it — bench opinion
				// had driven exactly one closing word.
				disp = pick(r.rng, []string{"repaired", "not_a_defect", "defect_accepted", "defect_owed_elsewhere", "amends_prior"})
				_, _ = r.exec("opinion", "--seat-id", seatID, "--id", id, "--as", disp,
					"--principle", "correctness", "--tension", "cost", "--review-flag", "false", "--settled", "the proposition this ruling bars", "--final", "--reason", "opinion-rationale-for-"+id)
			}
			res = append(res, map[string]any{"gap_id": id, "resolution": disp, "reason": "fuzz"})
		}
		// THE DEADLOCK ARM HAD NEVER BEEN DRIVEN, and the exemptions said so as though it were
		// a property of the engine rather than of this fake. `deadlock` was the literal `false`
		// here, so `outcome --as UNVERIFIED` and `--ended deadlock` were both unreachable, both
		// exempted, and the whole judged-termination path — including the assembler's stamp for
		// it — went unfuzzed. A real run produced one on 2026-08-22 (verdict_basis `asserted`,
		// ended `deadlock`), which is what falsified the premise.
		//
		// It is not a coin, for the same reason nothing else here is: the bench rules on what
		// happened. The engine's own precondition is "no gap remains carried" — a carry is the
		// bench saying the material needs another round, which is the opposite of stuck — so
		// deadlock falls out of the dispositions this sitting actually made.
		//
		// This drives BOTH arms of the cleared-board branch: where every open gap got a closing
		// disposition the board empties and the engine must grant red its further sitting, and
		// where gaps remain unruled it must terminate.
		deadlock := true
		for _, x := range res {
			if x.(map[string]any)["resolution"] == "carried" {
				deadlock = false
				break
			}
		}
		if r.forceUnverified {
			// DEADLOCK WITH ONE GAP LEFT UNRULED. Both halves are needed and neither alone does
			// it: deadlock with a CLEARED board hits the relief arm (debate.js grants red one
			// further sitting, which passes, and the run is VERIFIED), while open gaps without
			// deadlock just runs to the ceiling and stamps CEILING.
			//
			// ONE, not all of them. Disposing nothing starves red — it returns FAIL with an
			// empty gaps array in round 3 and the engine refuses the degenerate merge, which is
			// a correct refusal and not the path this drives. Leaving a single gap unruled keeps
			// every other party's flow ordinary and the board non-empty at the exit.
			if n := len(res); n > 0 {
				res = res[:n-1]
			}
			return map[string]any{"resolutions": res, "deadlock": true, "friction": arr()}
		}
		return map[string]any{"resolutions": res, "deadlock": deadlock, "friction": arr()}

	case strings.HasPrefix(seatID, "assemble"):
		r.register("bench", seatID)
		// THE VERDICT IS READ, NOT INVENTED. debate.js computes the terminal outcome and TELLS
		// the assembler ("Debate outcome: <verdict> after N round(s)") — exactly as a real seat
		// is told. The fake used to ignore that and draw from a hat, so `bench outcome --as`
		// could record a verdict contradicting the engine's own computation, and no oracle had
		// ever checked that a board with open gaps yields UNVERIFIED or a closed one VERIFIED.
		// The single most consequential value the engine emits was uncorrelated with everything
		// upstream of it.
		verd := "UNVERIFIED"
		if m := outcomeRe.FindStringSubmatch(prompt); m != nil {
			verd = m[1]
		}
		if r.forceHalt {
			verd = "HALTED" // a halted run's terminal outcome is HALTED (debate.js computes this)
		}
		// --reason is REQUIRED on outcome (#375): it is the run's terminal act, and every claim
		// or judgment act on this record carries its reasoning. Driving it without one leaves
		// `bench outcome` refused on every seed, which this sweep reports as a false green.
		oargs := []string{"outcome", "--seat-id", seatID, "--as", verd,
			"--reason", "fuzz: the run reached " + verd + " and the bench recorded how it ended"}
		// THE TERMINAL MODIFIER IS NOT A COIN, for the same reason the bench's dispositions are
		// not: it is a fact about what happened, and debate.js states it in this very prompt
		// ("by judged deadlock" / "by safety ceiling"). It WAS two 40% coins, and the deadlock
		// arm went undriven across 60 runs because UNVERIFIED now occurs about once in sixty —
		// a cleared docket grants red a further sitting instead of terminating — so a 40% coin
		// on a 1-in-60 verdict is a value the sweep reports as reachable and never reaches.
		//
		// The verdict itself is already read from the prompt, three lines up. Same channel, same
		// determinism. debate.js makes the two mutually exclusive (ceilingUnaudited requires
		// !deadlocked), so at most one lands.
		if !r.forceHalt && strings.Contains(prompt, "by safety ceiling") {
			oargs = append(oargs, "--ended", "ceiling")
		}
		if !r.forceHalt && strings.Contains(prompt, "by judged deadlock") {
			oargs = append(oargs, "--ended", "deadlock")
		}
		_, _ = r.exec(oargs...)
		_, _ = r.exec("assemble", "--seat-id", "assemble-r1")
		open := len(r.openGaps())
		return map[string]any{"synopsis": "fuzz", "open_gaps": open, "friction": arr()}

	default: // frontier, blue lanes, red lenses — register, and lenses record onto the channel
		if seatID != "" {
			// THE DEFAULT IS lens, SO EVERY MAPPING MISS IS A SILENTLY UNREGISTERED SEAT.
			// `frontier` is a BLUE seat (roles.go: blue owns blue-*, frontier) and was mapped
			// to lens here, so its register was REFUSED once per run, in every run, and that
			// seat's whole path went unexercised behind a green sweep. Found by the execution
			// tally (lens register: exactly 60 refusals across 60 runs — one per run is a
			// pattern, not noise).
			role := "lens"
			if strings.HasPrefix(seatID, "blue") || strings.HasPrefix(seatID, "frontier") {
				role = "blue"
			}
			r.register(role, seatID)
			// The non-prose lens channel is the RECORD now, not red/candidates files: a
			// citation lens records cite events, and every red lens records finding events
			// (label TOOL-assigned). This is the ONLY harness that drives that path end to
			// end through the real debate.js + binary, so it must actually exercise it.
			// RED RE-RUNS a recorded proof (#277) — the one audit that does not end in believing
			// bytes someone else chose.
			// RED RE-RUNS THE PROOF FOR THE GAP IT IS AUDITING, and the result is what red
			// closes on. It used to pick ANY proof on the record at 35% with no connection to
			// the gap, AFTER red had already closed it — reproduction decorating a decision
			// instead of earning it. Nothing would have noticed if reproduce returned garbage.
			r.reproveOpenProofs(seatID)
			if strings.HasPrefix(seatID, "red-lens") {
				// BOTH CASES, and the anchored one uses a REAL citation off the record: red
				// adjudicating a citation blue authored, and red corroborating a source it found
				// itself. Falling back to --independent when blue has cited nothing keeps the
				// verb driven from round 0 rather than only after the first cite lands.
				// TWO VERBS, because they are two acts: `verify` adjudicates a citation BLUE
				// authored and requires its anchor, `corroborate` records a source RED found
				// and requires the url and title that are the only thing identifying it.
				// Falling back to corroborate when blue has cited nothing keeps the axis driven
				// from round 0 rather than only after the first cite lands.
				// A REAL SPAN OF THE LIVE REPORT. This quoted `"fuzz claim "+seatID`, which is in
				// no document — and a SUPPORTING corroboration splices a citation anchor at the
				// claim now, so every one of them would be refused and the drive discards its
				// error. The seeded anchor sentence is the right span: the edit target below it
				// is swapped back and forth by blue's edits, this one is left alone precisely so
				// findings and cites have something stable to attach to.
				claim := "A § fuzz sentence to anchor findings."
				outcome := verifyOutcomes[r.rng.Intn(len(verifyOutcomes))]
				axes := func(c *cmd) *cmd {
					return c.set("--quote", claim).
						set("--as", outcome).
						set("--confidence", verifyConfidence[r.rng.Intn(len(verifyConfidence))]).
						set("--reason", "fuzz: what the source actually says").
						on(60, "--access-date", "2026-07-24")
				}
				if anchor := r.someCitation(); anchor != "" && r.coin(70) {
					axes(r.do("verify", seatID)).set("--anchor", anchor).run()
				} else {
					axes(r.do("corroborate", seatID)).
						set("--url", "https://fuzz.invalid/"+seatID).
						set("--title", "fuzz source for "+seatID).run()
				}
				// AND RED RAISES WHAT IT CONTRADICTED. A `refutes` or `absent` reading is red
				// saying the report asserts what its source does not, and a PASS is refused until
				// a finding quotes that claim — the tool will not write the finding itself,
				// because that would mean inventing its three grades. Driving the remedy here is
				// what makes both the gate and the act that clears it real in the sweep; without
				// it every run with a negative reading would wedge at the verdict.
				if outcome == "refutes" || outcome == "absent" {
					_, _ = r.exec("finding", "--seat-id", seatID,
						"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
						"--quote", claim, "--reason", "fuzz: the source says otherwise")
				}
				// --key from a small space so a repeated dispatch exercises retry idempotency.
				_, _ = r.exec("finding", "--seat-id", seatID, "--key", fmt.Sprintf("F%d", 1+r.rng.Intn(2)),
					"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(), "--quote", "§ fuzz", "--reason", "fuzz finding")
				// Red verifies a cited source by reading the CACHED bytes (#256): the same
				// `fetch` any seat uses. Driving it here is what makes the cache path — miss,
				// store, hit — real in the fuzz rather than unit-tested only.
				_, _ = r.exec("fetch", "--seat-id", seatID, "--url", sourceURL("/"+seatID))
				r.extras("lens", seatID, nil)
			}
		}
		return map[string]any{"synopsis": "fuzz", "petitions": arr(), "friction": arr(), "rulings": arr()}
	}
}

const preamble = `
globalThis.phase = function(){}; globalThis.log = function(){};
globalThis.budget = { total:null, spent:function(){return 0;}, remaining:function(){return Infinity;} };
globalThis.parallel = async function(t){ return Promise.all(t.map(function(x){return x();})); };
globalThis.pipeline = async function(items){ var st=Array.prototype.slice.call(arguments,1);
  return Promise.all(items.map(async function(it,i){ var v=it; for(var s=0;s<st.length;s++){ try{v=await st[s](v,it,i);}catch(e){return null;} } return v; })); };
`

func debateWrapped(t *testing.T) string {
	t.Helper()
	p, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read debate.js: %v", err)
	}
	src := strings.Replace(string(b), "export const meta", "const meta", 1)
	return "globalThis.__result = (async () => {\n" + src + "\n})()"
}

type outcome struct {
	seed      int64
	err       string // non-empty = a finding
	runDir    string
	verdict   string         // the terminal verdict this run reached (coverage signal)
	rounds    int            // how many rounds it ran (coverage signal)
	dialectic map[string]int // dialectic events this run left on the record (coverage signal)
	// #256 citation axis: the fetch -> cache -> cite -> anchor chain leaves TWO artifacts a
	// `cite` event alone does not prove — a cached source file and an INVISIBLE anchor spliced
	// into blue/report.md. `cite` is already in verbsWithEvents but is satisfied by red's
	// `lens cite`, which touches neither, so a blue-cite regression would pass the verb gate
	// silently. These two counters are what actually pin the new path.
	citeAnchors int
	cacheFiles  int
	// #267 provenance: blue_edit events carrying --answers. The verb gate below is satisfied
	// by ANY blue_edit, so an edit drive that never sent the flag would pass it silently —
	// exactly the false green the citation counters exist to prevent for cite.
	editAnswers int
	// #267 stage 3: gaps whose fix_basis was EARNED by a validated concrete proposal. The
	// mint verb gate is satisfied by any mint, so a proposal path that stopped validating
	// would pass it while the axis quietly went all-prose.
	verifiedBasis int
	// #267 stage 4: edits that applied red's proposal EXACTLY. Without a counter, the fuzz
	// could counter-edit every time and the estoppel path would never be reached at all.
	verbatimApplied int
	// applyMisses is why it did not, by cause — carried out of the run so a ZERO above arrives
	// with its explanation instead of inviting one.
	applyMisses map[string]int
}

// installAgent wires r as the seat backend on vm: it parses the seat id from each agent() prompt
// (falling back to the label), records the model tier the dispatch carried (#111 oracle), and
// resolves the promise with r's envelope for that seat.
func (r *runner) installAgent(loop *eventloop.EventLoop, vm *goja.Runtime) {
	vm.Set("agent", func(call goja.FunctionCall) goja.Value {
		prompt := ""
		if len(call.Arguments) > 0 {
			prompt = call.Argument(0).String()
		}
		seatID := ""
		if m := seatRe.FindStringSubmatch(prompt); m != nil {
			seatID = m[1]
		}
		var opts *goja.Object
		if len(call.Arguments) > 1 {
			opts = call.Argument(1).ToObject(vm)
		}
		if seatID == "" && opts != nil { // fall back to label
			if v := opts.Get("label"); v != nil {
				if fields := strings.Fields(v.String()); len(fields) > 0 {
					seatID = fields[0]
				}
			}
		}
		// #111 tier capture: record the model this dispatch carried ("unset" if absent).
		if opts != nil {
			mdl := "unset"
			if v := opts.Get("model"); v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
				mdl = v.String()
			}
			r.models = append(r.models, mdl)
		}
		// THE SEAT'S WORK LEAVES THE LOOP (#630 phase 3).
		//
		// This built the envelope INLINE and resolved an already-settled promise, so the
		// preamble's `parallel` — `Promise.all(t.map(x => x()))` — ran each thunk to completion
		// before starting the next. debate.js's lane and lens fan-outs were therefore serial in
		// the fuzz however wide they were configured, and the ~270 binary invocations of a run
		// went one at a time. Nothing ever contended for the record's locks, which is why #557's
		// window sat unreachable in the one harness that reaches a seat before the database
		// exists.
		//
		// Now the promise is returned PENDING and the work runs on a goroutine, so a thunk
		// returns immediately and its siblings start. Three lanes hit `register` on a fresh run
		// directory at once, which is the cross-process contention this exists for.
		//
		// goja is single-threaded and the Runtime must not be touched from here: the envelope is
		// built off-loop with no VM access, and vm.ToValue happens back ON the loop.
		p, resolve, reject := vm.NewPromise()
		go func() {
			env := r.envelopeFor(seatID, prompt)
			loop.RunOnLoop(func(vm *goja.Runtime) {
				if env == nil {
					// A seat that produced nothing is a rejection, not a silent fulfilment with
					// an empty envelope — debate.js reads fields off it and would fail further
					// from the cause.
					_ = reject(vm.ToValue("seat " + seatID + " produced no envelope"))
					return
				}
				_ = resolve(vm.ToValue(env))
			})
		}()
		return vm.ToValue(p)
	})
}

// driveDebate runs the wrapped debate.js once under goja with r as the seat backend. It returns the
// fulfilled result map (nil unless the debate settled fulfilled) and a non-empty settledErr for any
// preamble error, rejection, hang, or panic. Shared by the sweep (runOne) and the halt test — the
// one place the goja event loop and __result promise are driven.
func driveDebate(r *runner, wrapped string) (result map[string]any, settledErr string) {
	defer func() {
		if p := recover(); p != nil {
			settledErr = fmt.Sprintf("panic: %v", p)
		}
	}()
	// The fuzz builds its run directly rather than through `setup`, so it must lay down the
	// same inputs/run-config.json setup writes — the ceiling recorded there is what the CEILING
	// verdict is DERIVED against (#308). Without it, every ceiling run recorded an ASSERTED
	// verdict, which is exactly what the tripwire flagged: 24 of 60, purely for a missing file.
	_ = os.MkdirAll(filepath.Join(r.runDir, "inputs"), 0o755)
	_ = os.WriteFile(filepath.Join(r.runDir, "inputs", "run-config.json"),
		[]byte(`{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","maxRounds":"4","lanes":"3"}`), 0o644)

	// STARTED, NOT Run(). `loop.Run` returns once the JS job queue drains, which was correct
	// while agent() resolved inline — the whole debate settled inside one call. Phase 3 (#630)
	// resolves a seat's promise from a goroutine, so the queue can be momentarily empty with
	// work still outstanding, and Run would return mid-debate leaving __result pending. The
	// fuzz would then report "never settled (hang)" for every run: a harness artefact wearing
	// the exact words of a real defect.
	//
	// So the loop runs until the debate SAYS it is done, through __settle below.
	loop := eventloop.NewEventLoop()
	loop.Start()
	defer loop.Stop()

	// Buffered: __settle may fire before the receive, and an unbuffered send on the loop
	// goroutine would deadlock against a loop.Stop() this function has not reached yet.
	done := make(chan string, 1)
	loop.RunOnLoop(func(vm *goja.Runtime) {
		// THE SHAPE PRODUCTION ACTUALLY RUNS (#630). This drove `lanes: 1` under a
		// laneFloorOverride — the --smoke shape, which debate.js:82 REFUSES by default and
		// which the floor exists to stop ("run 2 silently ran under-provisioned at lanes=2").
		// At width one the blue lane fan-out (debate.js:725), red's lens-pass fan-out (:824)
		// and the per-lane method diversity (:560) were all exercised at a width no keeper run
		// uses. Both committed runs' inputs/run-config.json say lanes 3, maxRounds 4; so does
		// this. The override's own branch stays covered by debate.test.mjs, which is where a
		// JS-level guard belongs.
		vm.Set("args", map[string]any{
			"topic": "fuzz", "runDir": r.runDir, "binDir": binDir(r.bin),
			"lanes": 3, "maxRounds": 4,
			// #111: both tiers are REQUIRED — nil refuses dispatch. Both haiku so the tier oracle
			// expects every dispatched seat to carry exactly "haiku".
			"model": "haiku", "judgmentModel": "haiku",
		})
		if _, err := vm.RunString(preamble); err != nil {
			settledErr = "preamble: " + err.Error()
			return
		}
		r.installAgent(loop, vm)
		// __settle is how the debate tells Go it is over, from inside JS. Polling the promise
		// state from outside cannot do it: with async seats there is no moment at which "the
		// queue is empty" means "the debate finished".
		vm.Set("__settle", func(call goja.FunctionCall) goja.Value {
			msg := ""
			if len(call.Arguments) > 0 {
				msg = call.Argument(0).String()
			}
			select {
			case done <- msg:
			default: // already settled; the first word wins
			}
			return goja.Undefined()
		})
		if _, err := vm.RunString(wrapped); err != nil {
			done <- "run: " + err.Error()
			return
		}
		// ATTACHED IN JS, because a rejection has to reach the same channel a fulfilment does.
		// "" is the fulfilled signal; every other string is the failure, so an empty rejection
		// reason cannot read as success.
		if _, err := vm.RunString(`globalThis.__result.then(
			function(){ __settle(""); },
			function(e){ __settle("debate rejected: " + String(e)); })`); err != nil {
			done <- "attaching the settle handler: " + err.Error()
		}
	})

	// THE HANG IS STILL DETECTED, and now it is the only thing this wait can mean. The bound is
	// generous because a production-shaped run drives 270 invocations of a real binary; it is a
	// stuck-run tripwire, not a performance assertion.
	select {
	case msg := <-done:
		if msg != "" {
			return nil, msg
		}
	case <-time.After(10 * time.Minute):
		return nil, "debate never settled (hang)"
	}

	// READ THE RESULT BACK ON THE LOOP. Touching the Runtime from this goroutine is the one
	// thing goja does not allow, and it is not diagnosable when it goes wrong.
	read := make(chan struct{})
	loop.RunOnLoop(func(vm *goja.Runtime) {
		defer close(read)
		v := vm.Get("__result")
		pr, ok := v.Export().(*goja.Promise)
		if v == nil || !ok {
			settledErr = "debate produced no __result promise"
			return
		}
		switch pr.State() {
		case goja.PromiseStateRejected:
			settledErr = "debate rejected: " + truncate(pr.Result().String())
		case goja.PromiseStatePending:
			settledErr = "debate never settled (hang)"
		case goja.PromiseStateFulfilled:
			m, ok := pr.Result().Export().(map[string]any)
			if !ok {
				// A FULFILLED PROMISE THAT IS NOT AN OBJECT IS NOT A SETTLED DEBATE. The old
				// form left `result` nil and `settledErr` empty, which every caller reads as
				// "the run succeeded" — and the verdict then goes missing downstream instead of
				// here, where the actual value is still in hand to name.
				settledErr = fmt.Sprintf("debate fulfilled with %T, not an object: %s",
					pr.Result().Export(), truncate(pr.Result().String()))
				return
			}
			result = m
		}
	})
	<-read
	return result, settledErr
}

// readVerdict pulls the terminal verdict out of debate.js's settled result, and REFUSES a
// result that does not carry one.
//
// THE ABSENCE OF A VERDICT IS A FAILURE, NEVER A CATEGORY (#637). debate.js computes
//
//	const verdict = halted ? 'HALTED' : redEnv && redEnv.verdict === 'PASS' ? 'VERIFIED'
//	              : ceilingUnaudited ? 'CEILING' : 'UNVERIFIED'
//
// and returns it on every path it can reach, so the value is always one of four non-empty
// strings. A missing or non-string verdict therefore means THIS side failed to read the
// result — it never means the debate reached a nameless outcome.
//
// Tolerated, that miss reached the sweep's tally as `verdicts[""]++`, which counts exactly
// like a real verdict and is not one: the run that opened #637 recorded
// `verdicts=map[:29 CEILING:8 VERIFIED:3]`, 29 of 40 runs settling into a key no reader can
// distinguish from an outcome. That is the plausible zero [[facts-are-fields]] clause 3
// forbids — and worse than a bare zero, because the sweep's own coverage gates then pass over
// a distribution whose largest bucket means "we did not look".
//
// The error names the value AND the keys that were present, because "no verdict" alone does
// not separate a result that arrived empty from one that arrived with a different shape.
func readVerdict(result map[string]any) (string, error) {
	raw, present := result["verdict"]
	if !present {
		return "", fmt.Errorf("debate settled with no verdict: result has no %q key (keys present: %v)",
			"verdict", sortedKeys(result))
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("debate settled with a non-string verdict: %#v (%T)", raw, raw)
	}
	if s == "" {
		return "", fmt.Errorf("debate settled with an EMPTY verdict string (keys present: %v)", sortedKeys(result))
	}
	return s, nil
}

// sortedKeys renders a result's keys in a stable order, so two failures of the same shape
// produce the same message and can be told apart from two different shapes.
func sortedKeys(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// unverifiedSeed is the one seed in the sweep whose run is FORCED to the UNVERIFIED terminal
// verdict rather than left to the draw.
//
// THE SWEEP WAS STILL HOPING. #630 built the machinery to drive this deterministically —
// `runner.forceUnverified`, and TestFuzzUnverifiedPath which sets it — and the comment at
// forceUnverified says exactly why it is needed: "Left to the draw, a run whose gaps all
// happened to be repairable never contested anything, the judge never sat, and the run reached
// the round ceiling." But `execValues` is in-memory package state and TestFuzzUnverifiedPath
// sits ~600 lines BELOW TestFuzzDebate's coverage assertion, with no t.Parallel() in the file.
// Go runs them in source order, so the deterministic driver could never contribute to the gate
// that needs it, and the gate went on asking the 40-run draw for a value that occurs about once
// in sixty.
//
// MEASURED on origin/main at 65b27a9d with no other diff: red 3 of 4 runs, failing on
// `outcome --as UNVERIFIED`. One forced run costs one seed's worth of natural variation and
// makes the value a fact rather than a coin flip (#637).
const unverifiedSeed = 1

func runOne(t *testing.T, wrapped, bin string, seed int64, forceUnverified bool) (res outcome) {
	runDir, _ := os.MkdirTemp("", "fuzz-run-")
	// A lens finding anchors into blue/report.md and is rejected unless its --location
	// quote is present (slice 1b). Seed a report carrying the fuzzer's finding quote
	// ("§ fuzz") so findings are accepted and the coverage gate sees them.
	_ = os.MkdirAll(filepath.Join(runDir, "blue"), 0o755)
	// The trailing sentence is the EDIT TARGET: `blue edit` swaps it between two fixed
	// phrasings, so every round has a valid unique span to replace whichever way the last
	// edit left it, while the anchor quote above stays untouched for findings and cites.
	_ = os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte("# § fuzz\n\nA § fuzz sentence to anchor findings.\n\nThe cost is rising over time.\n"), 0o644)
	// THE RUN DECLARES ITS CLASS VOCABULARY, because a real run does. The fuzz used to be exempt
	// by accident: no registry staged meant `validateClass` accepted every slug, so the fuzz drove
	// `mint` for its whole life without ever exercising the check the flag's help describes. When
	// the exemption went, the coverage gate here reported it in one line — ZERO mint events across
	// 60 runs — which is what that gate is for.
	if err := record.StageForRun(runDir, fuzzClasses...); err != nil {
		return outcome{seed: seed, runDir: runDir, err: "stage the class registry: " + err.Error()}
	}
	r := &runner{bin: bin, runDir: runDir, rng: newLockedRand(seed), registered: map[string]bool{}, forceUnverified: forceUnverified}

	res = outcome{seed: seed, runDir: runDir}
	// NAMED RETURN + defer, because this function returns from a dozen places — an oracle that
	// fires early would otherwise drop the run's apply-miss causes on the floor, which is the
	// same silent loss the causes exist to explain.
	defer func() { res.applyMisses = r.applyMisses }()
	result, settledErr := driveDebate(r, wrapped)
	if settledErr != "" {
		res.err = settledErr
		return res
	}
	verdict, verr := readVerdict(result)
	if verr != nil {
		res.err = verr.Error()
		return res
	}
	res.verdict = verdict
	if n, ok := result["rounds"].(int64); ok {
		res.rounds = int(n)
	}
	// Oracle #111 (map-free): both tiers were configured haiku, so every dispatched seat must have
	// carried model "haiku" — none unset, none on another tier. This needs no bulk-seat list; the
	// SEAT_CLASS<->dispatch binding is proved separately by debate-dispatch.test.mjs.
	if len(r.models) == 0 {
		res.err = "no agent() call carried a model — the resolver regressed to unset"
		return res
	}
	for _, mdl := range r.models {
		if mdl != "haiku" {
			res.err = "a seat dispatched on tier " + mdl + ", not the configured haiku"
			return res
		}
	}
	// Oracle 1: the record the run left must pass verify.
	if out, err := tracked(bin, "verify", "--seat-id", "operator", "--run", runDir); err != nil {
		res.err = "verify FAILED:\n" + truncate(string(out))
		return res
	}
	// Oracle 1b: the JSON views the operator side reads in json-mode must exit 0 and parse.
	// `board` is already exercised above; `findings`/`friction` are JSON by name; `debate --json`
	// is the structured debate the capture audits count sections from. A broken view is what
	// would silently blank a dashboard tile or make an audit read an empty transcript.
	ids := mintedGapIDs(runDir)
	// EVERY PROJECTION, ON EVERY ROLE. `show` became a GROUP (0.56.0), so each projection is its
	// own command path — and the coverage gate immediately reported 34 paths never invoked,
	// because the sweep had only ever driven the handful it asserted on. A view that only the
	// merge reads is a view nobody checks from the seat that actually reads it.
	// EVERY SEAT READS EVERY PROJECTION, and the seat id is what says which tree `show` is in.
	for role, sid := range map[string]string{
		"blue": "blue-respond-r1", "lens": "red-lens-r1-L1", "merge": "red-merge-r1", "bench": "judge-r1",
	} {
		for _, v := range viewNamesForFuzz {
			args := []string{"show", v, "--run", runDir, "--seat-id", sid}
			if v == "changes" && len(ids) > 0 {
				args = append(args, "--id", ids[0])
			}
			if _, err := tracked(bin, args...); err != nil {
				res.err = role + " show " + v + " failed: " + err.Error()
				return res
			}
		}
		if _, err := tracked(bin, "show", "--run", runDir, "--seat-id", sid); err != nil {
			res.err = role + " show (bare, the seat's pending work) failed: " + err.Error()
			return res
		}
	}
	// READING THE REPORT AT AN ANCHOR, over the anchors this run actually minted.
	//
	// The unit tests drive a fixture; this drives whatever shape the sweep reached — a report
	// whose anchors sit at the top, at the bottom, adjacent to each other, or inside a section
	// that a later edit rewrote around them. The window's line arithmetic is where an
	// off-by-one lives, and a fixture with the anchor comfortably in the middle never asks.
	if a := someReportAnchor(runDir); a != "" {
		// EVERY ROLE, BOTH FLAGS. `show report` is defined once in internal/cli/seat, so a
		// window that worked on the merge and not on blue would mean the projection had grown
		// a per-role surface — and --window 0 is the degenerate size that must still resolve to
		// the anchored line rather than to nothing.
		// A SEAT ID PER ROLE, because the tree is scoped to the dispatched identity: without one
		// the binary builds the OPERATOR surface, where `show` does not exist at all.
		for _, sid := range map[string]string{
			"blue": "blue-respond-r1", "lens": "red-lens-r1-L1", "merge": "red-merge-r1", "bench": "judge-r1",
		} {
			for _, extra := range [][]string{nil, {"--window", "0"}} {
				args := append([]string{"show", "report", "--anchor", a, "--run", runDir, "--seat-id", sid}, extra...)
				out, err := tracked(bin, args...)
				if err != nil {
					res.err = strings.Join(args, " ") + " failed:\n" + truncate(string(out))
					return res
				}
				// THE WINDOW MUST CONTAIN ITS OWN ANCHOR. An exit-0 window that does not is
				// the plausible zero: a read that reports success and shows the wrong place.
				if !strings.Contains(string(out), a) {
					res.err = strings.Join(args, " ") + " returned a window WITHOUT the anchor it was addressed by:\n" + truncate(string(out))
					return res
				}
			}
		}
	}
	// AN ANCHOR NOBODY MINTED IS REFUSED, NOT READ EMPTY — the read-side twin of `show changes
	// --id R9-99`. An empty window says "the report has nothing here", which is a different
	// fact from "that anchor is not in this report".
	if out, err := tracked(bin, "show", "report", "--anchor", "f-ffffffff", "--run", runDir, "--seat-id", "blue-respond-r1"); err == nil {
		res.err = "show report --anchor f-ffffffff SUCCEEDED on an anchor nobody minted — a window over nothing:\n" + truncate(string(out))
		return res
	}
	// The OPERATOR's friction read — seats write the channel, the human reads it back.
	if _, err := tracked(bin, "friction", "--run", runDir, "--seat-id", "operator"); err != nil {
		res.err = "operator friction read failed: " + err.Error()
		return res
	}
	// THE WRONG-ADDRESS DRIVE IS GONE, because the address collision is. A seat's `friction
	// --reason` used to land on the OPERATOR's read, which takes no --reason, so it died in the
	// parser before any message could teach — and the channel it could not reach is the one for
	// reporting exactly that. The two frictions are on different trees now: a seat's carries its
	// write verb and not the read, the operator's the read and no seat verbs. There is nothing
	// left to land on by accident, so there is nothing here to drive.
	// `friction` left the SEAT menu (0.57.0) — it is the operator's read. The verb stays on
	// every role; only the view moved.
	for _, v := range []string{"findings"} {
		out, err := tracked(bin, "show", v, "--run", runDir, "--seat-id", "red-merge-r1")
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show " + v + " did not return valid JSON:\n" + truncate(string(out))
			return res
		}
	}
	{
		out, err := tracked(bin, "show", "debate", "--json", "--run", runDir, "--seat-id", "red-merge-r1")
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show debate --json did not return valid JSON:\n" + truncate(string(out))
			return res
		}
	}
	// Oracle 1c: the markdown views now render in-memory (view.Markdown) with no render-shadow
	// round-trip. Each must exit 0 on whatever state the run reached — this restores the randomized
	// coverage the `render` verb + the difftest RENDERS section used to give the projection renderer,
	// which the render-shadow removal (#203) took away (view_test.go pins the bytes on fixed
	// fixtures; only the fuzz drives them across arbitrary run shapes).
	//
	// DRIVEN FROM cli.ViewNames(), not a list maintained here. A hand-kept roster of what
	// exists is how the coverage gap this oracle closes was created in the first place: a
	// view ships, nobody adds it to the list, and the sweep reports full coverage of a
	// surface it never drove. The JSON-by-name views take their own dispatch branches above.
	// EVERY ROLE'S show, not merge's alone. Each role has a different DEFAULT view and its own
	// role gate, so driving only `merge show --view` left the other three reachable but never
	// reached — and --id (the scoped form) never passed at all.
	for _, role := range []string{"blue", "lens", "bench"} {
		if out, err := tracked(bin, "show", "debate", "--run", runDir, "--seat-id", seatOfRole(role)); err != nil {
			res.err = role + " show debate failed:\n" + truncate(string(out))
			return res
		}
	}
	if ids := mintedGapIDs(runDir); len(ids) > 0 {
		for _, role := range []string{"blue", "lens", "bench"} {
			_, _ = tracked(bin, "show", "changes", "--id", ids[0], "--run", runDir, "--seat-id", seatOfRole(role))
		}
	}
	// THE SET IS DERIVED, because this exclusion was four names written here by hand and two of
	// them were wrong in opposite directions: `friction` stopped being a view entirely (it is
	// the OPERATOR's read of the channel, never a projection), so the case excluded nothing;
	// and `motions`, `telemetry` and `evidence` are JSON by name and were NOT excluded, so this
	// loop drove three projections down "the markdown path" while the comment said their own
	// oracles had them. An exclusion list that names a view which does not exist, and misses
	// half the ones it is about, reads as a considered boundary and is neither.
	jsonByName := map[string]bool{}
	for _, v := range cli.JSONByNameViews() {
		jsonByName[v] = true
	}
	for _, v := range cli.ViewNames() {
		if jsonByName[v] {
			continue // JSON by name — driven by their own oracles, not the markdown path
		}
		if out, err := tracked(bin, "show", v, "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
			res.err = "show " + v + " (projection) failed:\n" + truncate(string(out))
			return res
		}
	}
	// And the SCOPED form, over a gap the run actually minted: `--view changes --id <gap>`
	// takes a different path (it resolves the gap and renders the comparison), so the unscoped
	// render above proves nothing about it. A gap the board does not know must be REFUSED, not
	// rendered empty — the read-side twin of requireGap.
	if ids := mintedGapIDs(runDir); len(ids) > 0 {
		if out, err := tracked(bin, "show", "changes", "--id", ids[0], "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
			res.err = "show changes --id " + ids[0] + " failed:\n" + truncate(string(out))
			return res
		}
		if out, err := tracked(bin, "show", "changes", "--id", "R9-99", "--run", runDir, "--seat-id", "red-merge-r1"); err == nil {
			res.err = "show changes --id R9-99 SUCCEEDED on a gap nobody minted — a view that invents a comparison:\n" + truncate(string(out))
			return res
		}
	}
	// Oracle 1d: THE READ-ONLY CENSUS. Every record-nothing surface must survive whatever
	// shape this run reached. These emit no events, so the verb-coverage gate below is blind
	// to them by construction — this is the only thing that exercises them at all.
	if msg := sweepReadOnly(bin, runDir); msg != "" {
		res.err = msg
		return res
	}
	// SCENARIO ORACLE: the record must SHOW the decision the scenario specified.
	//
	// This is what the directives buy. Every other oracle here asks "did anything break" —
	// verify passes, the views parse, prose renders. None can tell a chain that CARRIED a
	// decision from one that dropped it, because until now no decision was written down before
	// it was taken. Each gap is now minted with the outcome it should reach, so the terminal
	// record is checkable rather than merely well-formed.
	//
	// IGNORE is the sharpest: if --answers stopped recording provenance, an ignored gap and a
	// repaired one would look identical — and that join is what #267's whole measurement axis
	// is built on.
	if board, err := record.BoardState(runDir); err == nil {
		answered, disputed, proved := map[string]bool{}, map[string]bool{}, map[string]bool{}
		// THE SWITCH IS ON THE BODY, not on a type string beside it. Each arm reaches straight for
		// a field, so binding the message and the type in one step removes the pair that could
		// disagree — and a grade motion's gap id now comes off the GradeMotion arm rather than
		// from a payload key that petitions and directions also had.
		for _, e := range board.Events {
			body, ok := recordpb.Body(e)
			if !ok {
				continue
			}
			switch t := body.(type) {
			case *recordpb.Proof:
				if t.GetAnswers() != "" {
					proved[t.GetAnswers()] = true
				}
			case *recordpb.BlueEdit:
				if t.GetAnswers() != "" {
					answered[t.GetAnswers()] = true
				}
			case *recordpb.Motion:
				if g := t.GetGrade(); g != nil && g.GetGapId() != "" {
					disputed[g.GetGapId()] = true
				}
			}
		}
		for _, id := range board.GapOrder {
			g := board.Gaps[id]
			// Only gaps a responder was SHOWN can be judged: one minted in the terminal round
			// never reaches blue, so its scenario was never dispatched.
			if g == nil || g.Mint == nil || !r.presented[id] {
				continue
			}
			switch g.Mint.GetRequiredFix() {
			case dirIgnore:
				if answered[id] {
					res.err = "scenario IGNORE: " + id + " was answered by a blue_edit, but the scenario said blue does nothing — either a writer ignored the directive or --answers is attaching provenance nobody claimed"
					return res
				}
			case dirDisputeWon, dirDisputeLost:
				if !disputed[id] {
					res.err = "scenario DISPUTE: " + id + " has no dispute event — the contest the scenario specified is absent from the record"
					if why := r.disputeRefusals[id]; why != "" {
						res.err += "\nthe filing was REFUSED, and this is what it said:\n  " + why
					} else {
						res.err += "\nno refusal was recorded for it either, so the drive never reached `motion grade file` at all"
					}
					return res
				}
			case dirProve, dirProveDrifts:
				// The chain must be VISIBLE ON THE RECORD, not merely to have happened: a proof
				// answering this gap, and the gap closed only behind a reproduction that held.
				if !proved[id] {
					res.err = "scenario PROVE: " + id + " has no proof answering it — the computation the scenario specified left no artifact red could re-run"
					return res
				}
				if r.evaluated[id] && !g.Open && !r.reproduced[id] {
					res.err = "scenario PROVE: " + id + " was CLOSED without its proof reproducing — red accepted a computation on somebody's word, which is the one thing re-running exists to prevent"
					return res
				}
			case dirApply:
				if !answered[id] {
					res.err = "scenario APPLY: " + id + " has no blue_edit answering it — a repair the scenario specified left no provenance"
					return res
				}
			}
			// THE TERMINAL FATE, which is the assertion the first pass was missing. Found by
			// probe: making red never close a repaired gap produced ZERO oracle failures — the
			// run simply ended CEILING with work sitting done-but-open, and nothing said so.
			// Scoped to gaps red actually sat on after blue answered: a repair landed in the
			// final round leaves the gap legitimately open, because red never sits again.
			if satisfied(g.Mint.GetRequiredFix()) && r.evaluated[id] && g.Open {
				res.err = "scenario " + g.Mint.GetRequiredFix() + ": " + id + " is still OPEN after red sat on the repaired board — work was done and the board never recorded it as finished"
				return res
			}
		}
	}

	// A RULING HAS A CONSEQUENCE, AND A CONTEST IS VISIBLE AS ONE.
	//
	// Red's ruling and the line of inquiry's fate were both on the record and joined NOWHERE, so blue
	// pursuing a line red called out-of-scope looked exactly like pursuing one red endorsed.
	// The design says a ruling is an ARGUMENT blue may contest. That contest used to be
	// `contests_ruling`, a field set as a SIDE EFFECT of moving a line to `pursued` — so it could
	// only record disagreement that won, and a seat that argued and then yielded fell out of the
	// count silently (measured on a real run; see #344). It is `motion direction appeal` now: its
	// own event, with its own reason, filed against the ruling and independent of what the line's
	// status does next.
	if board, err := record.BoardState(runDir); err == nil {
		contests := map[string]string{}
		// THE PRE-#344 ARM IS GONE. It read a `line-of-inquiry` event carrying `contests_ruling`,
		// kept because "a stored record carries it and the oracle runs against replayed records as
		// well as fresh ones". No stored record carries it: the event type does not exist in the
		// schema, so nothing can hold that shape and the arm could only ever be dead.
		for _, e := range board.Events {
			a, ok := recordpb.BodyAs[*recordpb.MotionAppeal](e)
			if ok && a.GetSubject() == recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
				contests[a.GetMotionId()] = "appealed"
			}
		}
		for _, a := range record.Inquiries(board) {
			ruling := record.InquiryRuling(runDir, a.ID)
			if ruling == "" || a.Status == "proposed" {
				continue // never ruled, or blue has not answered yet
			}
			switch {
			case strings.HasPrefix(a.Line, avContest):
				if a.Status != "pursued" {
					res.err = "line of inquiry " + a.ID + " was proposed as CONTESTED but ended " + a.Status + " — blue was to pursue it against the ruling"
					return res
				}
				if contests[a.ID] == "" {
					res.err = "line of inquiry " + a.ID + " was pursued AGAINST a " + ruling + " ruling and the record does not say so — the disagreement is invisible, which is the state this join exists to end"
					return res
				}
			case ruling == "endorsed":
				if a.Status != "pursued" {
					res.err = "line of inquiry " + a.ID + " was ENDORSED and ended " + a.Status + " — a ruling with no consequence"
					return res
				}
			default:
				if a.Status == "pursued" && contests[a.ID] == "" {
					res.err = "line of inquiry " + a.ID + " was ruled " + ruling + " and pursued anyway with nothing recording the contest"
					return res
				}
			}
		}
	}

	// EVERY RETIREMENT IS EVIDENCED. The fake retires only text a recorded edit removed, so an
	// `asserted` retire means either the fake regressed to phantom retirements or the
	// edit-tracking that evidences them broke. A phantom retire cancels real claim loss in the
	// scorecard's additive-integrity detector, so it must not pass unnoticed here either.
	if board, err := record.BoardState(runDir); err == nil {
		for _, e := range board.Events {
			r, ok := recordpb.BodyAs[*recordpb.Retire](e)
			if ok && r.GetRemovalBasis() != record.RemovalVerified {
				res.err = "a retire recorded removal_basis=" + r.GetRemovalBasis() +
					" — nothing on the record shows that claim was ever in the report, and an unevidenced retirement cancels real claim loss in the additive-integrity detector"
				return res
			}
		}
	}

	// AN ASSERTED VERDICT IS EITHER A JUDGED DEADLOCK OR A BROKEN DERIVATION, and this is the
	// tripwire that keeps the second from hiding behind the first.
	//
	// The tool refuses an --as that contradicts the record (#308), so a recorded verdict is
	// either DERIVED or came from the one case the record cannot decide: a judged deadlock.
	//
	// THIS USED TO ASSERT THAT NO DEADLOCK COULD EXIST — "debate.js cannot produce one,
	// `deadlock` is hardcoded false" — and treated every asserted basis as derivation failure.
	// That premise was false by the time anyone checked it: the 2026-08-22 sqlite-schema run
	// stamped `verdict_basis: asserted` / `ended: deadlock` off a real bench call. The comment
	// predicted its own obsolescence and said the right response was to update rather than
	// widen it, so that is what this is — the assertion now turns on whether the RECORD shows a
	// deadlock, not on a claim about what the engine can reach.
	//
	// An asserted verdict with `ended: deadlock` is the sanctioned case and passes. An asserted
	// verdict WITHOUT one still means the derivation stopped working, and still fails here.
	if board, err := record.BoardState(runDir); err == nil {
		for _, e := range board.Events {
			o, ok := recordpb.BodyAs[*recordpb.Outcome](e)
			if !ok {
				continue
			}
			switch o.GetVerdictBasis() {
			case record.VerdictDerived:
			case "":
				res.err = "the outcome event carries no verdict_basis — the field that says whether the verdict was computed or claimed has gone missing"
				return res
			default:
				// THE SANCTIONED CASE, WHICH THE COMMENT ABOVE ALREADY DESCRIBED AND THE CODE NEVER
				// CHECKED. DeriveVerdict reports "cannot answer" on a judged DEADLOCK by design —
				// that is a real answer, not a gap to paper over — so the bench asserts the outcome
				// and stamps `ended: deadlock` to say why. This arm fired on it anyway.
				//
				// IT PASSED FOR YEARS BECAUSE THE ARM WAS UNREACHABLE. Its own message says
				// "debate.js cannot produce a judged deadlock while it is hardcoded false (#289)",
				// and that was true when it was written. The stub judge now derives deadlock from
				// the dispositions it actually made, precisely so both arms of the cleared-board
				// branch are driven — and the moment deadlock became reachable, this gate started
				// failing the healthy outcome. A premise that expires without the assertion
				// noticing is the shape this suite exists to catch, met in the suite itself.
				if o.GetEnded() == "deadlock" {
					continue
				}
				res.err = "the run recorded an ASSERTED verdict (" + recordpb.Word(o.GetVerdict()) + ") and did NOT stamp `ended: deadlock` — a judged deadlock is the one case the record cannot derive a verdict for, and without it an asserted verdict means the derivation stopped working"
				return res
			}
		}
	}

	// VERDICT ORACLE: the recorded outcome must agree with the board it describes.
	//
	// The terminal verdict is the single most consequential value the engine emits, and until
	// now the fake DREW IT FROM A HAT — so `bench outcome --as` could record VERIFIED over a
	// board with open gaps and nothing anywhere would notice. The assembler is now told the
	// verdict the way a real seat is (debate.js states it in the prompt), which makes the
	// agreement between outcome and board checkable for the first time.
	if board, err := record.BoardState(runDir); err == nil {
		openCount := 0
		for _, id := range board.GapOrder {
			if g := board.Gaps[id]; g != nil && g.Open {
				openCount++
			}
		}
		recorded := ""
		for _, e := range board.Events {
			if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok {
				// UPPERCASE, because that is the word this oracle compares against and the
				// schema's spelling is lowercase. The fold is stated here rather than left to
				// a reader to notice.
				recorded = strings.ToUpper(recordpb.Word(o.GetVerdict()))
			}
		}
		if recorded == "VERIFIED" && openCount > 0 {
			res.err = fmt.Sprintf("verdict oracle: the run recorded VERIFIED with %d gap(s) still open — a pass over unfinished work", openCount)
			return res
		}
		// THE CONVERSE NEEDS A CARVE-OUT, and finding out why is the point of running it.
		//
		// It fired 2 of 60 on "CEILING with every gap closed", which looked like a verdict
		// contradicting its board — and is not. The BENCH closes gaps at the terminal sitting,
		// AFTER the ceiling has already been determined: the verdict describes the debate's
		// exit, not the post-adjudication board. That is why Gap.ClosedByBench exists.
		//
		// So the check is only sound over gaps RED closed. A non-passing verdict whose board
		// was cleaned entirely by the bench is correct; one whose board red itself emptied is
		// the contradiction worth catching.
		if recorded != "" && recorded != "VERIFIED" && recorded != "HALTED" && openCount == 0 && len(board.GapOrder) > 0 {
			benchClosed := 0
			for _, id := range board.GapOrder {
				if g := board.Gaps[id]; g != nil && g.ClosedByBench {
					benchClosed++
				}
			}
			if benchClosed == 0 {
				res.err = "verdict oracle: the run recorded " + recorded + " with every gap closed BY RED — an unfinished verdict over work red itself finished"
				return res
			}
		}
	}

	// Oracle 2: every dialectic event's prose must actually RENDER in the report — the A1-A3
	// class (prose written under one key, read under another) is invisible to verify but caught
	// here, on every run.
	// One replay of the record serves both remaining oracles (prose + coverage tally).
	board, berr := record.BoardState(runDir)
	if berr != nil {
		res.err = "board: " + berr.Error()
		return res
	}
	if m := proseRenders(t, board, runDir); m != "" {
		res.err = m
	}
	if res.err == "" {
		if m := basisRenders(t, board, runDir); m != "" {
			res.err = m
		}
	}
	res.dialectic = tallyDialectic(board)
	// #256: count the citation axis's two real artifacts (see outcome). A cite that fetched and
	// anchored leaves BOTH; a cite event alone leaves neither.
	if md, err := os.ReadFile(filepath.Join(runDir, "blue", "report.md")); err == nil {
		res.citeAnchors = strings.Count(string(md), "<!--cite:")
	}
	for _, e := range board.Events {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch t := body.(type) {
		case *recordpb.BlueEdit:
			if t.GetAnswers() != "" {
				res.editAnswers++
			}
			if t.GetAppliedVerbatim() {
				res.verbatimApplied++
			}
		case *recordpb.Mint:
			if t.GetFixBasis() == "verified" {
				res.verifiedBasis++
			}
		}
	}
	if ents, err := os.ReadDir(filepath.Join(runDir, "cache")); err == nil {
		for _, e := range ents {
			if !e.IsDir() && e.Name() != "index" {
				res.cacheFiles++
			}
		}
	}
	return res
}

// #111: debate.js must REFUSE to dispatch when either model tier is unset — the engine never
// guesses or inherits a tier. The guard throws at the top of the async IIFE, so __result rejects
// with the flag-named message. No binary needed: the throw precedes any tool act.
func TestDispatchRefusesUnsetModel(t *testing.T) {
	wrapped := debateWrapped(t)
	cases := []struct {
		name, want string
		args       map[string]any
	}{
		{"model unset", "refusing dispatch — model unset", map[string]any{"topic": "x", "runDir": "research/2026-01-01_t", "lanes": 1, "laneFloorOverride": "t", "judgmentModel": "haiku"}},
		{"judgmentModel unset", "refusing dispatch — judgmentModel unset", map[string]any{"topic": "x", "runDir": "research/2026-01-01_t", "lanes": 1, "laneFloorOverride": "t", "model": "haiku"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			loop := eventloop.NewEventLoop()
			var got string
			loop.Run(func(vm *goja.Runtime) {
				vm.Set("args", c.args)
				vm.Set("agent", func(goja.FunctionCall) goja.Value { return goja.Undefined() })
				if _, err := vm.RunString(preamble); err != nil {
					t.Fatalf("preamble: %v", err)
				}
				if _, err := vm.RunString(wrapped); err != nil {
					got = "run: " + err.Error()
				}
			})
			loop.Run(func(vm *goja.Runtime) {
				if v := vm.Get("__result"); v != nil {
					if pr, ok := v.Export().(*goja.Promise); ok && pr.State() == goja.PromiseStateRejected {
						got = pr.Result().String()
					}
				}
			})
			if !strings.Contains(got, c.want) {
				t.Errorf("expected rejection containing %q, got %q", c.want, got)
			}
		})
	}
}

// verbsWithEvents are the record-event-emitting verbs the fuzz is meant to drive — the whole seat
// surface bar the read-only ones (register/show emit no board event). The coverage gate
// asserts each fired at least once across the run set, so a regression that silently STOPS
// emitting one fails loudly (the old gate guarded only cite/finding). `halt` terminates the run
// and is covered by TestFuzzHaltPath, not the random sweep — the gate skips it (see coverExempt).
var verbsWithEvents = []string{
	"closing", "position", "opinion", "regrade", "mint", "close",
	"cite", "verify", "finding", "avenue", "reproduce", "friction", "revision", "retire",
	"manifest_row", "verdict", "spot_check", "certify", "declare", "halt",
	// friction-none is the EXPLICIT NEGATIVE arm of the friction verb — a distinct event type,
	// so a gate listing only "friction" would report the channel covered while the arm that
	// makes an empty log meaningful went undriven.
	"friction_none",
	// The motion collapse (#344): filed by any seat, ruled by one, appealed by the filer.
	"motion", "motion_rule", "motion_appeal",
	// Added 2026-08-04 by a census of every type record.Append can write: these three were
	// APPENDABLE BUT UNGATED, so a regression that stopped emitting any of them would have
	// left the sweep green. `anchor` is the finding-marker's own record (the immortal-marker
	// detector's EXPECTED set is exactly these), `class-new` is the growing gap registry's
	// write, `outcome` is the bench's.
	"blue_edit", "anchor", "class_new", "outcome", "proof",
	// The remaining schema types, named so the census below has a complete list to check against.
	"closing", "inquiry_review", "register",
}

// coverExempt names verbs tallied but NOT required in the random-sweep coverage gate.
var coverExempt = map[string]bool{
	"halt": true, // terminal — covered by TestFuzzHaltPath
	// NO VERB WRITES `observe`. The event type is in the schema and nothing in the command tree
	// produces it: `lens observe` does not exist, and the only Appends of a recordpb.Observe are
	// in record's own tests. It is exempted rather than driven, because a drive would have to
	// invent a verb — and named here rather than dropped from the list, so the type's homelessness
	// is a line somebody has to read instead of a silence.
	//
	// The `Observation` a board carries is built from FINDING events, not from this type.
	"observe": true,
}

// TestFuzzHaltPath drives the JUDICIAL HALT terminal path — kept OUT of the random sweep because a
// halt ends the run and reshapes every downstream oracle. blue-synthesize files a petition
// (forceHalt guarantees it); the judge-petition sitting rules HALT via the `halt` verb; debate.js
// sets halted, breaks the loop, and the run ends HALTED. Asserts: the run settles as HALTED, a
// halt event is on the record, and the halted run STILL passes verify (a safety exit is a valid
// record, not a broken one). This is the dedicated coverage for `halt`, which the sweep exempts.
func TestFuzzHaltPath(t *testing.T) {
	bin := buildBinary(t)
	wrapped := debateWrapped(t)
	runDir, _ := os.MkdirTemp("", "fuzz-halt-")
	defer os.RemoveAll(runDir)
	r := &runner{bin: bin, runDir: runDir, rng: newLockedRand(1), registered: map[string]bool{}, forceHalt: true}

	result, settledErr := driveDebate(r, wrapped)
	if settledErr != "" {
		t.Fatalf("halt run did not settle cleanly: %s", settledErr)
	}
	if v, _ := result["verdict"].(string); v != "HALTED" {
		t.Fatalf("expected verdict HALTED, got %q", v)
	}
	if h, _ := result["halted"].(bool); !h {
		t.Fatal("forceHalt run did not end halted — the judicial-halt terminal path is unexercised")
	}
	board, err := record.BoardState(runDir)
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	halts := 0
	for _, e := range board.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_HALT {
			halts++
		}
	}
	if halts == 0 {
		t.Fatal("no halt event on the record — the halt verb never ran")
	}
	if out, err := tracked(bin, "verify", "--seat-id", "operator", "--run", runDir); err != nil {
		t.Fatalf("verify FAILED on the halted run (a safety exit must still be a valid record):\n%s", truncate(string(out)))
	}
}

// tallyDialectic counts the events that prove the fuzz exercised the paths it claims to.
func tallyDialectic(board *record.Board) map[string]int {
	want := map[string]bool{}
	for _, v := range verbsWithEvents {
		want[v] = true
	}
	m := map[string]int{}
	for _, e := range board.Events {
		// The type is an enum now; the tally is keyed on the WORD, which is what verbsWithEvents
		// carries and what the failure message names.
		w := recordpb.Word(e.GetType())
		if want[w] {
			m[w]++
		}
	}
	return m
}

// THIS MAP WAS AN ALLOWLIST, AND THAT IS WHY IT STOPPED WORKING.
//
// It held six event types. A type absent from it was not checked, and absence was silent — so
// every event type added after it was written defaulted to "the report need not render this",
// with nothing anywhere saying so. A 2026-08-08 trace of all 30 types found four whole exchanges
// reaching the record and never the reader (petition filings, line of inquiry rulings, retirements,
// observations) plus two verbs read by nothing at all, under a green sweep.
//
// It is now EXHAUSTIVE over the types a run actually produces: every event type seen on the
// record must appear here or in reportExemptions with a stated reason, and an unclassified type
// FAILS. A new verb cannot be added without deciding, in writing, what its reader gets.
var dialecticProseKey = map[string]string{
	// The debate proper.
	// THE FIELD NAMES, NOT THE FLAG NAMES. Every entry here said `reason` because that is what a
	// seat TYPES; the schema spells the prose field per verb — a closing stores `text`, an opinion
	// `rationale`, a regrade `basis`. Read against a payload map the wrong name simply never
	// matched, so this gate was inert for most of the types it claims to cover, reporting "the
	// report renders every recorded prose" over rules that could not fire. fieldStr makes the miss
	// loud, which is how these were found.
	"position": "text", "closing": "text", "opinion": "rationale",
	// Motions: the ASK and the ANSWER, on one id. `petition-rule` alone used to be the half-dialog — the
	// reader got the bench's ruling with no question attached.
	// The board.
	"mint": "problem", "finding": "text", "regrade": "basis",
	// Directions: the line, red's ruling on it, and blue's reason for its fate.
	"avenue": "line",
	// Red's per-round verdict that the REPORT still carries the line. It renders ON THE LINE'S OWN
	// ROW in the three research areas rather than in a section of its own, because the claim it
	// answers ("we pursued X") lives there and a verdict a reader has to go and find is a verdict
	// most readers do not find. `reason` is what red read at that line — the quoted text, not the
	// grade, which is the half a reader can check.

	// The lens's below-the-bar work and the fate the merge gave it.
	// Substance leaving the report, on the record, with its reason.
	"retire": "claim",
	// Run-level voices.
	"friction": "text", "revision": "text", "halt": "opinion", "certify": "statement", "declare": "holding",
	// The friction channel's EXPLICIT NEGATIVE. It renders for the same reason the complaint
	// does, and arguably a stronger one: a reader weighing "no friction this run" needs to know
	// whether the seats looked and said so, or never used the channel. Those were the same
	// bytes for eighteen recorded sittings, and this event is what separates them.
	"friction_none": "text",
	// Blue's self-audit receipt, one per repaired gap. `row` is what blue checked and what
	// checking it showed — the receipt reached no reader for a year, because the coverage metric
	// counted the ENVELOPE array and the verb was named in no prompt at all (#318).
	"manifest_row": "row",
	// The motion group. `basis` is the ASK, `opinion` the answer, `reason` the appeal — all three
	// render in the report's one Motions section, joined on the motion id (#344).
	"motion":        "basis",
	"motion_rule":   "opinion",
	"motion_appeal": "reason",
	// Red re-reading its own closure archive. The prose is what the sample FOUND — the whole
	// point of sampling — and it reached no reader at all until the floor was enforced (#317).
	"spot_check": "reason",
}

// reportExemptions are event types whose prose is deliberately NOT expected in the report, each
// with its reason. Stated rather than omitted: an absence with no reason is indistinguishable
// from an oversight, which is precisely how this gate decayed.
var reportExemptions = map[string]string{
	"register":  "a seat announcing itself to the run — attribution machinery, and the attribution reaches the reader on every act that seat records, never as an entry of its own",
	"anchor":    "an estoppel key spliced INTO blue/report.md — it is machinery for the edit path, and the text it anchors is the lifted content itself",
	"blue_edit": "mutates blue/report.md, which assembly lifts verbatim; the edit's effect IS in the report, and rendering the old/new spans again would duplicate the document",
	"class_new": "registers a gap class; the class reaches the reader on every gap that carries it, not as an entry of its own",
	// Red's independent re-run. The NOTE is its judgement; whether it reproduced is computed
	// by the tool and rendered beside the proof either way (#343).
	"reproduce": "reason",
	"verify":    "red adjudicating ONE citation. It reaches the reader through the evidence view rather than the report body: the report carries BLUE's citations (woven into the bibliography), and red's verdict on them is audit provenance rather than a claim the reader acts on. It IS surfaced — attached to the source it names in `show evidence`, in the board's citations count, and, for `refutes` and `absent`, as an assembly-screen FAILURE if the report still cites what red found against",
	"cite":      "resolved rather than rendered — the anchor becomes a visible [^N] and the source becomes a ## Bibliography line (weaveCitations)",
	"proof":     "resolved rather than rendered — weaveProofs splices the computation at its anchor",
	"close":     "the closure's prose is red's acceptance argument and reaches the reader only as an index row today; rendering it in full is tracked, not silently accepted",
	// The per-round review of the report against the lines on the record. Its READER IS A GATE,
	// not the document: record.InquiryReviewDue asks whether this round's review exists and
	// refuses the sitting until it does. Rendering it in the report would put a process check
	// where the debate goes — the review says "I read the report against the record", and what
	// a reader wants is the outcome of that reading, which arrives as an ordinary GAP when the
	// treatment falls short. Named here rather than left silent, because the gate is right that
	// an unclassified event type is how a report loses a whole exchange.
	"inquiry_review": "read by record.InquiryReviewDue as a per-round duty gate, not rendered: a shortfall the review finds is minted as an ordinary gap, which is what reaches the reader",
	"outcome":        "composed into the verdict stamp by verdictStamp, from the payload's verdict/deadlocked/exhausted fields rather than a prose field",
	"verdict":        "red's per-round PASS/FAIL, consumed by DeriveVerdict into the terminal outcome; the round-by-round spine is not yet a transcript section",
}

// basisFields are the DERIVED-NOT-ASSERTED fields, mapped to the event that carries each and a
// phrase the report must contain when that value is on the record.
//
// Every one of these exists because a seat asked to self-report reports the flattering value —
// so each is computed by the tool at the write. Each then gated something (estoppel, the outcome
// cross-check, the claim-loss detector) and reached the reader as NOTHING, which collapsed the
// distinction at the only point where a human could act on it: a verdict the record itself
// decided printed the same word as one the bench merely asserted.
//
// prose keys cannot cover this — a basis is a field on an event whose OTHER prose renders fine,
// so the prose gate is green while the qualifier is gone.
var basisFields = []struct{ evType, key, value, want string }{
	{"outcome", "verdict_basis", "derived", "derived from the record"},
	{"outcome", "verdict_basis", "asserted", "asserted by the bench"},
	{"mint", "fix_basis", "verified", "with the text in front of it"},
	{"mint", "fix_basis", "proposed", "nothing checked this demand"},
	{"retire", "removal_basis", "verified", "the record shows it leaving"},
	{"retire", "removal_basis", "asserted", "nothing on the record shows it was ever present"},
	// proof_basis was already rendered before this gate existed (weaveProofs prints it with an
	// explanation of reproducible vs observed). It is here so the coverage is stated rather
	// than assumed — the one that was fine is the easiest to break unnoticed.
	{"proof", "proof_basis", "reproducible", "reproducible"},
	{"proof", "proof_basis", "observed", "observed"},
}

func basisRenders(t *testing.T, board *record.Board, runDir string) string {
	report, err := os.ReadFile(filepath.Join(runDir, "report.md"))
	if err != nil {
		return "no report.md: " + err.Error()
	}
	rpt := string(report)
	var missing []string
	for _, bf := range basisFields {
		present := false
		for _, e := range board.Events {
			// A CLOSED gap's mint is still on the record but its fix_basis is no longer shown —
			// the demand was met, and the report's closure index is a one-line index by design.
			// Only OPEN gaps are checked, which is where the qualifier changes what a reader does.
			if recordpb.Word(e.GetType()) != bf.evType {
				continue
			}
			if fieldStr(t, e, bf.key) == bf.value {
				if bf.evType == "mint" {
					m, _ := recordpb.BodyAs[*recordpb.Mint](e)
					if !gapIsOpen(board, m.GetGapId()) {
						continue
					}
				}
				present = true
				break
			}
		}
		if present && !strings.Contains(rpt, bf.want) {
			missing = append(missing, fmt.Sprintf("%s.%s=%s is on the record and the report never says so (looked for %q)", bf.evType, bf.key, bf.value, bf.want))
		}
	}
	if len(missing) > 0 {
		return "derived-basis absent from report — the qualifier the field exists to preserve:\n" + strings.Join(missing, "\n")
	}
	return ""
}

func gapIsOpen(board *record.Board, id string) bool {
	g := board.Gaps[id]
	return g != nil && g.Open
}

func proseRenders(t *testing.T, board *record.Board, runDir string) string {
	report, err := os.ReadFile(filepath.Join(runDir, "report.md"))
	if err != nil {
		return "no report.md: " + err.Error()
	}
	rpt := string(report)
	var missing, unclassified []string
	seen := map[string]bool{}
	for _, e := range board.Events {
		w := recordpb.Word(e.GetType())
		key, ok := dialecticProseKey[w]
		if !ok {
			// An event type in neither table is a DECISION NOBODY MADE. Registering it here
			// costs one line; leaving it unclassified is how four exchanges reached the record
			// and never the reader.
			if reportExemptions[w] == "" && !seen[w] {
				seen[w] = true
				unclassified = append(unclassified, w)
			}
			continue
		}
		prose := strings.TrimSpace(fieldStr(t, e, key))
		if prose == "" || strings.Contains(rpt, prose) {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s/%s prose absent from report: %q", e.GetSeatId(), w, prose))
		if len(missing) >= 5 {
			break
		}
	}
	if len(unclassified) > 0 {
		sort.Strings(unclassified)
		return "event type(s) on the record with NO entry in dialecticProseKey and NO stated exemption: " + strings.Join(unclassified, ", ") +
			"\nDecide what the reader gets from each — render it (add the prose key) or say why not (add the exemption). Silence here is how the report loses a whole exchange."
	}
	if len(missing) > 0 {
		return "prose-not-rendered (A1-A3 class):\n" + strings.Join(missing, "\n")
	}
	return ""
}

// motionIDOf pulls the tool-assigned motion id out of a filing's JSON envelope.
func motionIDOf(out string) string {
	var env struct {
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		return ""
	}
	return env.Result.MotionID
}

func binDir(bin string) string { return filepath.ToSlash(filepath.Dir(bin)) }

func truncate(s string) string {
	if len(s) > 600 {
		return s[:600] + "…"
	}
	return s
}

// viewNamesForFuzz is the projection set, taken from the command tree so a view added or removed
// is swept without anyone remembering to edit a list here.
var viewNamesForFuzz = seat.ViewNames()

// surfaceQuorum is the run count at or above which the coverage gates can hold the sweep to the
// FULL surface. Below it a low-frequency path can flake to zero and fail an honest run.
//
// MEASURED, not guessed: across a default 60-run sweep the scarcest path is `motion inquiry
// appeal` at 9 invocations (0.15/run), which gives ~10% odds of never firing at N=15 against
// ~0.25% at N=40. The number is shape-dependent — lane-driven verbs scale with `lanes`, dispute-
// driven ones with `maxRounds` — so it is a floor for THIS shape, and a shape change re-tunes it.
// That is why the gates below assert the tally rather than trusting this constant to stand in
// for coverage (#630).
const surfaceQuorum = 40

func TestFuzzDebate(t *testing.T) {
	bin := buildBinary(t)
	wrapped := debateWrapped(t)

	// The DEFAULT is a modest smoke that proves the harness and catches gross regressions in
	// ~15s on Linux — which is what the Linux legs, main pushes, tag builds and the nightly
	// all run. Windows PULL REQUESTS run -short (hooks.yml): each run shells the real binary
	// ~50-70 times, and Windows pays process-spawn cost on every one, which made this
	// package that leg's critical path (573s, measured 2026-08-25) after the RAM disk took
	// I/O out of the bill. What a run proves about the PLATFORM — the .exe build, the
	// spawns, the temp plumbing — run 15 proves as well as run 60; depth past that is
	// statistics on debate semantics, which are platform-independent and keep full depth on
	// Linux. The full 1000-run confidence sweep is on demand:
	// FUZZ_N=1000 go test ./integration/fuzz -run TestFuzzDebate -timeout 1200s.
	// THE DEFAULT IS THE QUORUM, so the default sweep is always one that can assert the surface.
	// 60 was a number the gates did not depend on; at production shape a run costs ~2.2x what the
	// smoke shape cost, and the sweep needs exactly enough runs for the coverage gates to hold.
	n := surfaceQuorum
	if testing.Short() {
		n = 15
	}
	if v := os.Getenv("FUZZ_N"); v != "" {
		if k, err := strconv.Atoi(v); err == nil && k > 0 {
			n = k
		}
	}
	// Each run is process-spawn-bound (it shells the real binary ~50-70 times), not CPU-bound,
	// so the goroutines mostly WAIT on subprocesses — oversubscribe past core count to keep every
	// core busy spawning. FUZZ_C overrides.
	concurrency := runtime.NumCPU() * 3
	if v := os.Getenv("FUZZ_C"); v != "" {
		if k, err := strconv.Atoi(v); err == nil && k > 0 {
			concurrency = k
		}
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []outcome
	var completed int
	verdicts := map[string]int{}
	roundHist := map[int]int{}
	dcov := map[string]int{}
	citeAnchors, cacheFiles := 0, 0 // dialectic-event coverage across all runs (proves the fuzz emits them)
	editAnswers := 0                // #267: blue_edit events that carried the provenance key
	verifiedBasis := 0              // #267 stage 3: gaps whose fix_basis was EARNED by a validated pair
	verbatimApplied := 0            // #267 stage 4: edits that applied red's proposal exactly (the estoppel precondition)
	applyMisses := map[string]int{} // and why it did not, by cause — a bare 0 above named none of them

	for i := 0; i < n; i++ {
		seed := int64(i) + 1
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			o := runOne(t, wrapped, bin, seed, seed == unverifiedSeed)
			// THE ORACLE RUNS ON EVERY RECORD THE SWEEP PRODUCES. The drives above assert that
			// each command SUCCEEDED; the oracle asserts that the record those commands built is
			// one every projection agrees about — the cross-reader class the unit suites cannot
			// see, because each exercises one reader against its own fixture. A violation keeps
			// the run directory like any other failure, so the disagreement can be inspected.
			if o.err == "" {
				if violations, cerr := consistency.Check(runtest.Open(t, o.runDir)); cerr != nil {
					o.err = "consistency oracle: " + cerr.Error()
				} else if len(violations) > 0 {
					o.err = "consistency violations:\n  " + strings.Join(violations, "\n  ")
				}
			}
			// The oracle opened this run's cached handle in-process; release it or the
			// RemoveAll below fails on Windows and the sweep leaks one handle per seed.
			_ = recordsql.CloseUnder(o.runDir)
			mu.Lock()
			completed++
			verdicts[o.verdict]++
			roundHist[o.rounds]++
			for k, v := range o.dialectic {
				dcov[k] += v
			}
			citeAnchors += o.citeAnchors
			cacheFiles += o.cacheFiles
			editAnswers += o.editAnswers
			verifiedBasis += o.verifiedBasis
			verbatimApplied += o.verbatimApplied
			for why, n := range o.applyMisses {
				applyMisses[why] += n
			}
			if o.err != "" {
				failures = append(failures, o)
			} else {
				os.RemoveAll(o.runDir) // keep only failing runs, for inspection
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// AND WHY IT DID NOT, WHENEVER IT DID NOT. `%d edits applied a proposal verbatim` read 0
	// across 60 runs and named no cause: the branch could have gone untaken, found no proposal
	// on the gap, found the span already edited away, or run and been refused. Four causes, one
	// zero, in the instrument whose job is to detect exactly that shape.
	misses := ""
	if len(applyMisses) > 0 {
		causes := make([]string, 0, len(applyMisses))
		for why, n := range applyMisses {
			causes = append(causes, fmt.Sprintf("%d× %s", n, why))
		}
		sort.Strings(causes)
		misses = "\n  verbatim-apply declined: " + strings.Join(causes, "\n                          ")
	}
	t.Logf("fuzzed %d debate runs · %d failed · verdicts=%v · rounds=%v\n  dialectic events emitted: %v\n  citation axis: %d anchors spliced · %d sources cached\n  provenance: %d of %d blue_edit ops carried --answers · %d of %d gaps earned fix_basis=verified · %d edits applied a proposal verbatim%s",
		completed, len(failures), verdicts, roundHist, dcov, citeAnchors, cacheFiles, editAnswers, dcov["blue_edit"], verifiedBasis, dcov["mint"], verbatimApplied, misses)
	// FULL-SURFACE COVERAGE GATE. A green fuzz that never drove a verb is a false green (the lens
	// stub emitted neither cite nor finding for the whole life of PR-1, unexercised end to end).
	// Assert EVERY event-emitting seat verb fired at least once across the run set — so a
	// regression that silently stops emitting one (a dropped envelope branch, a broken dispatch)
	// fails here, not silently. `halt` is exempt (terminal; covered by TestFuzzHaltPath). Each
	// gated verb fires with per-run probability well above ~20%, so P(missed across the default 60
	// runs) is negligible; if a verb ever flakes here it is a real coverage regression, not noise.
	// The full surface needs enough runs for every verb (incl. the ~10%/run ones like regrade) to
	// fire; below the quorum a low-frequency verb could flake to zero, so the -short smoke keeps
	// only the cite/finding floor (the original false-green guard) and the default+ size asserts
	// all.
	//
	// A SWEEP UNDER QUORUM NOW SAYS SO. These gates used to fall silent below 40 and print the
	// same green as a sweep that ran them — so `-short` (N=15) has been running this package with
	// its three coverage gates inactive and nothing on the record saying which. That is #428's
	// class (a gate that did not fire rendered as a gate that held) inside the sweep built to
	// detect exactly it.
	measured := completed >= surfaceQuorum
	if !measured {
		t.Logf("NOT MEASURED: %d runs is under the surface quorum of %d, so these gates did NOT run: "+
			"the per-verb event gate, the citation/provenance floors, the full-surface command gate, "+
			"and the flag/enum coverage sweeps. This is not a pass over them — only the cite/finding "+
			"floor below was checked. Run the default sweep to assert the surface.", completed, surfaceQuorum)
	}
	if measured {
		for _, k := range verbsWithEvents {
			if coverExempt[k] {
				continue
			}
			if dcov[k] == 0 {
				t.Errorf("fuzz drove ZERO %s events across %d runs — that seat verb is unexercised (false green); a generator branch dropped it", k, completed)
			}
		}
	}
	// #256 CITATION-CHAIN GATE. `cite` in the verb gate above is satisfied by red's `lens cite`,
	// which neither fetches nor anchors — so a blue-cite regression would sail through it. These
	// two assert the chain that actually matters ran end to end through the real binary: a source
	// was FETCHED into <run>/cache (loopback server, see sourceURL) and an INVISIBLE anchor was
	// spliced into blue/report.md. Zero of either across the whole sweep is a false green.
	if measured {
		if citeAnchors == 0 {
			t.Errorf("fuzz spliced ZERO <!--cite:--> anchors across %d runs — `blue cite` never anchored (false green); the citation axis is unexercised", completed)
		}
		if cacheFiles == 0 {
			t.Errorf("fuzz cached ZERO sources across %d runs — `fetch`/`blue cite` never populated <run>/cache (false green); the fetch path is unexercised", completed)
		}
		// #267 PROVENANCE GATE. The verb gate above accepts any blue_edit, so an edit drive
		// that stopped sending --answers would satisfy it while the join key every #267
		// measurement reads went silently missing.
		if editAnswers == 0 {
			t.Errorf("fuzz recorded ZERO blue_edit events carrying --answers across %d runs — the provenance key is unexercised (false green)", completed)
		}
		if verifiedBasis == 0 {
			t.Errorf("fuzz minted ZERO gaps with fix_basis=verified across %d runs — no concrete proposal ever validated, so the whole stage-3 path is unexercised (false green)", completed)
		}
		if verbatimApplied == 0 {
			t.Errorf("fuzz recorded ZERO verbatim applications across %d runs — nothing ever estopped red, so the stage-4 guard is unexercised (false green)", completed)
		}
	}
	if !measured {
		for _, k := range []string{"cite", "finding"} {
			if dcov[k] == 0 {
				t.Errorf("fuzz drove ZERO %s events across %d runs — the lens record channel is unexercised (false green)", k, completed)
			}
		}
	}
	// FULL-SURFACE GATE, DERIVED FROM THE COMMAND TREE.
	//
	// The harness tallies EVERY invocation it makes (noteExec), so "what did the fuzz drive"
	// is observed rather than declared. cli.CommandPaths() walks the real cobra tree for
	// "what exists". Nothing hand-written states the surface, which is the point: the gap
	// this closes was created by a hand list going stale — 18 of 44 seat verbs and 7 of 9
	// root commands undriven, every undriven seat verb one that records nothing and so was
	// invisible to the event gate above.
	//
	// The only hand-written list left is exemptSurfaces, and each entry states its reason.
	if measured {
		if un := unreachedSurfaces(); len(un) > 0 {
			t.Errorf("%d command path(s) exist in the tool and were NEVER invoked across %d runs (false green — add a drive, or an exemption with its reason):\n  %s",
				len(un), completed, strings.Join(un, "\n  "))
		}
		// DRIVING A VERB IS NOT DRIVING ITS SURFACE. The gate above asks whether every verb
		// ran; these ask whether the flags those verbs take were ever passed, and whether
		// every value of every closed set was ever recorded. An unpassed flag is code no run
		// has executed, and an unreached enum member is a word nothing has written — both
		// sitting behind a sweep that reports full coverage.
		if un := unreachedFlags(); len(un) > 0 {
			t.Errorf("%d flag(s) on verbs the sweep DOES drive were never passed:\n  %s", len(un), strings.Join(un, "\n  "))
		}
		if un := unreachedEnumValues(); len(un) > 0 {
			t.Errorf("%d enum value(s) were never driven:\n  %s", len(un), strings.Join(un, "\n  "))
		}
	}
	// AT EVERY SIZE, both of these. The tally is the sweep's own trajectory and costs nothing to
	// print; withholding it under quorum hid the evidence for the gate below on exactly the runs
	// most likely to need it. And a path that never succeeded is a coverage hole at ANY N — it
	// needs no quorum, because one refusal per invocation is not a sampling accident.
	t.Log(execReport())
	if un := neverSucceeded(); len(un) > 0 {
		t.Errorf("%d command path(s) were driven and NEVER SUCCEEDED — the surface gate counts invocations, so these read as covered while their success path has never run:\n  %s",
			len(un), strings.Join(un, "\n  "))
	}
	if len(failures) > 0 {
		// EIGHT, AND IT SAYS SO WHEN IT TRUNCATES. Sixty seeds failing the same way is a wall of
		// identical text, so the cap is right — but "17/60 failed, see seeds above" over eight
		// printed seeds invites a reader to take the eight as the whole distribution and count
		// classes from them. A silent cap reads as full coverage; this one names what it dropped.
		show := failures
		if len(show) > 8 {
			show = show[:8]
		}
		for _, f := range show {
			t.Errorf("seed %d FAILED (runDir %s):\n%s", f.seed, f.runDir, f.err)
		}
		more := ""
		if len(failures) > len(show) {
			var rest []string
			for _, f := range failures[len(show):] {
				rest = append(rest, fmt.Sprintf("%d", f.seed))
			}
			more = fmt.Sprintf(" — %d of them NOT printed above (seeds %s); re-run one of those directly to see its failure",
				len(rest), strings.Join(rest, " "))
		}
		t.Fatalf("%d/%d fuzz runs failed — see seeds above (reproduce with that seed)%s", len(failures), n, more)
	}
}

// buildBinary is testbuild.Binary: built once per test binary instead of once per call, and
// named by the platform convention rather than an unconditional ".exe".
func buildBinary(t *testing.T) string {
	t.Helper()
	return testbuild.Binary(t, "feov-record")
}

// mintedGapIDs lists the gaps a run actually created, so a scoped-view oracle asks about
// something real rather than about a shape the fuzz happened not to produce this seed.
// someReportAnchor returns one anchor id this run actually put in the report, or "" before any
// finding has been recorded. Read from the RECORD (the `finding` event's label) rather than by
// scanning report.md for a token, so the oracle is not testing the reader with the reader.
func someReportAnchor(runDir string) string {
	b, err := record.BoardState(runDir)
	if err != nil {
		return ""
	}
	for _, e := range b.Events {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok && f.GetFindingId() != "" {
			return f.GetFindingId()
		}
	}
	return ""
}

func mintedGapIDs(runDir string) []string {
	b, err := record.BoardState(runDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range b.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			out = append(out, m.GetGapId())
		}
	}
	return out
}

// someProposal returns a gap carrying a concrete proposal, with its exact pair, so the fuzz
// can drive the VERBATIM-application path rather than only the counter-edit one.

// scenarioOf reads a gap's directive BACK FROM THE BOARD, not from Go memory. The round trip
// is part of what is tested: if required_fix stopped surviving a mint, every seat would
// silently revert to coin-flip behaviour and the oracle would catch it.

// reproveOpenProofs re-runs the proof answering each open PROVE gap and records whether it
// held. This is the lens's audit — the one that does not end in believing bytes somebody else
// chose — and merge closes on its result rather than on its own say-so.

// recentlyEditedOut returns text a recorded edit removed and which is absent from the report
// now — a claim whose retirement the record can evidence.
func (r *runner) recentlyEditedOut() string {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return ""
	}
	cur, err := os.ReadFile(filepath.Join(r.runDir, "blue", "report.md"))
	if err != nil {
		return ""
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		e := b.Events[i]
		be, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		if !ok {
			continue
		}
		old := be.GetOld()
		if old != "" && !strings.Contains(string(cur), old) {
			return old
		}
	}
	return ""
}

// inquiryFates are the scenarios a line of inquiry is proposed under. The line itself carries it, the
// way a gap's required_fix does — red rules from it, and blue's next move answers the ruling.
const (
	avEndorse = "FUZZ-INQUIRY-ENDORSE: a line worth the run's time"
	avScope   = "FUZZ-INQUIRY-OUT-OF-SCOPE: a real question, but not this run's"
	avThin    = "FUZZ-INQUIRY-TOO-THIN: in scope, but the hypothesis does not carry its budget"
	// avContest is the case gb's design named and the tool could not express: red rules the
	// line out, and blue PURSUES IT ANYWAY with an argument. A ruling is an argument, never a
	// command — `blue dispute --id A1` is refused outright, because disputes are gap-shaped —
	// so the move itself is the contest, and it is now recorded as one.
	avContest = "FUZZ-INQUIRY-CONTESTED: red rules it out and blue pursues it anyway, with reasons"
)

var inquiryFates = []string{avEndorse, avEndorse, avScope, avThin, avContest}

// rulingFor maps a proposed line to the ruling red should give it.
//
// UNDERSCORES, because that is what the tool accepts. These were `out-of-scope` and `too-thin`,
// and DirectionRuling spells `out_of_scope` / `too_thin` — so 17 of 27 rulings across 60 runs were
// refused, and everything below a ruling starved with them: a contested line can only be appealed
// after it is ruled, which is why `motion inquiry appeal` reported as an unreached path.
//
// Same family as `--as supports-with-bridge` advertised in help and refused by the write path: one
// value spelled two ways across a boundary, with only one side moved. The refusals were discarded
// by the drive, so the only thing that noticed was a coverage gate reporting a MISSING drive for a
// path whose drive was fine.
func rulingFor(line string) string {
	switch {
	case strings.HasPrefix(line, avEndorse):
		return "endorsed"
	case strings.HasPrefix(line, avScope):
		return "out_of_scope"
	case strings.HasPrefix(line, avThin):
		return "too_thin"
	case strings.HasPrefix(line, avContest):
		return "out_of_scope"
	}
	return ""
}

// ruleOpenInquiries has red rule every line of inquiry that has no ruling yet, from the line it carries.
func (r *runner) ruleOpenInquiries(seatID string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return
	}
	for _, a := range record.Inquiries(b) {
		if rulingFor(a.Line) == "" || directionRuling(b, r.runDir, a.ID) != "" {
			continue
		}
		_, _ = r.exec("motion", "inquiry", "rule", "--seat-id", seatID, "--id", a.ID,
			"--as", rulingFor(a.Line), "--reason", "fuzz: ruling as the line was proposed")
	}
}

// directionRuling reports a line of inquiry's ruling under EITHER vocabulary.
//
// Asking only record.InquiryRuling would have seen the legacy events alone, so every
// motion-ruled line of inquiry would read as unruled and be ruled again each round — the drive would
// have looked correct and measured nothing, because a second ruling on a settled line is not a
// path the run takes.
func directionRuling(b *record.Board, runDir, inquiryID string) string {
	if v := record.InquiryRuling(runDir, inquiryID); v != "" {
		return v
	}
	for _, m := range record.Motions(b) {
		if m.Subject == "inquiry" && m.Fields["inquiry_id"] == inquiryID && m.Ruled() {
			return m.Ruling
		}
	}
	return ""
}

// answerInquiryRulings is blue's move after red has ruled: comply, or CONTEST by pursuing
// anyway with an argument. Which one is decided by the line, not by a coin.
func (r *runner) answerInquiryRulings(seatID string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return
	}
	for _, a := range record.Inquiries(b) {
		ruling := directionRuling(b, r.runDir, a.ID)
		if ruling == "" || a.Status != "proposed" {
			continue
		}
		switch {
		case strings.HasPrefix(a.Line, avContest):
			// THE CONTEST, IN BOTH VOCABULARIES. `contests_ruling` is a field the move sets as a
			// side effect; `motion direction appeal` is the act named as itself. Both are live
			// during the additive stage, so the move still happens either way — what the appeal
			// adds is that the disagreement has its own event instead of riding on a status
			// change, which is the whole reason the collapse treats an appeal as one act across
			// all three subjects.
			// UNCONDITIONAL, not a coin. Behind a 50% gate this reached 2 invocations across 60
			// runs — contests are already rare — and a drive that thin flakes to ZERO, which the
			// unreached-path gate reports as a missing drive in CI and nowhere else. The legacy
			// `contests_ruling` field is still exercised by the move below either way.
			_, _ = r.exec("motion", "inquiry", "appeal", "--seat-id", seatID, "--id", a.ID,
				"--reason", "fuzz: the scope call is wrong, this bears on the core claim")
			r.do("line-of-inquiry move", seatID).set("--id", a.ID).set("--as", "pursued").
				set("--reason", "fuzz: the scope call is wrong, this bears on the core claim").run()
		case ruling == "endorsed":
			r.do("line-of-inquiry move", seatID).set("--id", a.ID).set("--as", "pursued").
				set("--reason", "fuzz: endorsed, taking it up").run()
		default:
			r.do("line-of-inquiry move", seatID).set("--id", a.ID).set("--as", "declined").
				set("--reason", "fuzz: accepting the ruling").run()
		}
	}
}

func (r *runner) reproveOpenProofs(seatID string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return
	}
	if r.reproduced == nil {
		r.reproduced = map[string]bool{}
	}
	proofFor := map[string]string{} // gap -> sha
	for _, e := range b.Events {
		if p, ok := recordpb.BodyAs[*recordpb.Proof](e); ok && p.GetAnswers() != "" && p.GetProofSha() != "" {
			proofFor[p.GetAnswers()] = p.GetProofSha()
		}
	}
	for _, id := range r.openGaps() {
		if d := r.scenarioOf(id); d != dirProve && d != dirProveDrifts {
			continue
		}
		sha := proofFor[id]
		if sha == "" {
			continue // blue has not answered it yet; red has nothing to re-run
		}
		out, err := r.exec("--json", "reproduce", "--seat-id", seatID, "--id", sha, "--as", pick(r.rng, []string{"sound", "unsound"}), "--reason", "fuzz: read the script and re-ran it")
		if err != nil {
			continue
		}
		var env struct {
			Result struct {
				Matches bool `json:"matches"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) == nil && env.Result.Matches {
			r.mu.Lock()
			r.reproduced[id] = true
			r.mu.Unlock()
		}
	}
}

func (r *runner) scenarioOf(gapID string) string {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return ""
	}
	g := b.Gaps[gapID]
	if g == nil || g.Mint == nil {
		return ""
	}
	return g.Mint.GetRequiredFix()
}

// blueRespondTo carries out the scenario each open gap was minted with — one decision per gap,
// taken from the gap rather than from a coin, which is what makes the terminal state assertable.
func (r *runner) blueRespondTo(seatID string, open []string) {
	for _, id := range open {
		switch r.scenarioOf(id) {
		case dirApply:
			// The only branch that sets applied_verbatim, and so the only one that estops red.
			//
			// EVERY WAY IT CAN FAIL IS COUNTED, because the fallthrough below is silent and the
			// gate downstream reported "0 verbatim applications" across 60 runs without being
			// able to say whether the branch was never taken, took and found no proposal, took
			// and found the span already edited away, or ran and was REFUSED. Four causes, one
			// zero — the plausible-zero shape, in the instrument meant to detect it.
			gid, fo, fn := r.proposalFor(id)
			switch {
			case gid == "":
				r.noteApplyMiss("no proposal on the gap (red minted prose-only)")
			default:
				cur, err := os.ReadFile(filepath.Join(r.runDir, "blue", "report.md"))
				switch {
				case err != nil:
					r.noteApplyMiss("report unreadable: " + err.Error())
				case !strings.Contains(string(cur), fo):
					r.noteApplyMiss("the proposed span is no longer in the report (an earlier edit moved it)")
				default:
					if _, err := r.do("edit", seatID).set("--quote", fo).set("--new", fn).
						set("--answers", id).set("--reason", "fuzz: applying red's proposed text verbatim").run(); err != nil {
						r.noteApplyMiss("blue edit REFUSED: " + firstLine(err.Error()))
						break
					}
					continue
				}
			}
			// No applicable pair — degrade to a counter-edit rather than fake one.
			fallthrough
		case dirCounter:
			r.counterEdit(seatID, id)
		case dirDisputeWon, dirDisputeLost:
			dim := pick(r.rng, disputeDims)
			// A DIFFERENT GRADE FROM THE ONE ON THE BOARD. A motion proposing the grade already
			// there is refused now (it asks for no change), and a drive that hits that 1 time in
			// 5 would report as a missing dispute event three layers away — which is exactly the
			// shape the hyphenated grade words produced.
			proposed := r.gradeOtherThan(r.currentGrade(id, dim))
			out, err := r.exec("--json", "motion", "grade", "file", "--seat-id", seatID, "--id", id,
				"--dimension", dim, "--proposed", proposed, "--reason", "fuzz: contesting the grade on "+id)
			// THE REFUSAL IS KEPT, because the oracle downstream reports its ABSENCE and cannot
			// say why. `scenario DISPUTE: R1-1 has no dispute event` is what a discarded error
			// looks like from the far end — a coverage report about a drive that ran fine, which
			// is the same shape as the hyphenated ruling words that made `motion inquiry appeal`
			// read as unreached. 2 of 7 filings were refused across 60 runs and nothing said so.
			if err != nil {
				if r.disputeRefusals == nil {
					r.disputeRefusals = map[string]string{}
				}
				r.disputeRefusals[id] = err.Error()
			}
			if err == nil {
				if mid := motionIDOf(out); mid != "" {
					// The envelope ref keeps (gap_id, dimension) because debate.js routes the
					// DOCKET on them; the motion id rides alongside so the ruling can name the
					// ask it answers. Not redundant: the pair says which grade, the id says
					// which ASK about it, and #312 is the case where one gap and one dimension
					// had two asks and the pair could not tell them apart.
					ref := map[string]any{"gap_id": id, "dimension": dim, "proposed": proposed, "motion_id": mid}
					r.raised = append(r.raised, ref)
					r.disputedThisRound = append(r.disputedThisRound, ref)
				}
			}
		case dirProve, dirProveDrifts:
			// SETTLE IT BY COMPUTING. The script is written into the run dir and the tool runs
			// it TWICE, so this drives the real interpreter, the cache, the anchor splice and
			// the reproducible/observed grading — and --answers ties the execution to the gap
			// it settles, which is what makes red's re-run targetable rather than arbitrary.
			name := "fuzz-answer-" + id + ".js"
			body := "console.log('settles " + id + "');"
			if r.scenarioOf(id) == dirProveDrifts {
				body = "console.log(Math.random());" // graded `observed`, and it will not re-run the same
			}
			if err := os.WriteFile(filepath.Join(r.runDir, name), []byte(body), 0o644); err == nil {
				prove := r.do("prove", seatID).
					set("--quote", "§ fuzz").
					set("--script", name).
					set("--answers", id).
					set("--reason", "fuzz: computing rather than arguing")
				// The METHOD is cited and the instance computed — that pairing is the whole
				// claim of the proof axis. It passed the FABRICATED label "c-fuzz", which named
				// no citation on the record: a proof claiming a provenance it did not have, and
				// the report rendered the link. `requireCitation` refuses it now, so this takes
				// a real anchor off the record — the same discipline someFinding and someCitation
				// already impose everywhere else a label is passed.
				if anchor := r.someCitation(); anchor != "" && r.coin(50) {
					prove.set("--cites", anchor)
				}
				prove.run()
			}
		case dirIgnore:
			// Deliberately nothing. The oracle asserts NO edit answers this gap, which is what
			// proves --answers records provenance rather than decorating it.
		}
	}
}

// counterEdit makes a real edit that is NOT red's proposed text.
func (r *runner) counterEdit(seatID, gapID string) {
	cur, err := os.ReadFile(filepath.Join(r.runDir, "blue", "report.md"))
	if err != nil {
		return
	}
	oldSpan, newSpan := "rising over time", "climbing sharply"
	if !strings.Contains(string(cur), oldSpan) {
		oldSpan, newSpan = newSpan, oldSpan
	}
	if !strings.Contains(string(cur), oldSpan) {
		return
	}
	r.do("edit", seatID).set("--quote", oldSpan).set("--new", newSpan).
		set("--answers", gapID).set("--reason", "fuzz: counter-edit, not red's text").run()
}

// proposalFor returns the concrete pair for ONE gap, when it carries one.
func (r *runner) proposalFor(gapID string) (string, string, string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return "", "", ""
	}
	g := b.Gaps[gapID]
	if g == nil || g.Mint == nil || g.Mint.GetFixBasis() != "verified" {
		return "", "", ""
	}
	// THE SPAN IS THE GAP'S OWN `location`. `fix_old` was a second copy of it, matched by a
	// second matcher; a proposal is --quote (the span, required anyway) plus --new.
	return gapID, g.Mint.GetLocation(), g.Mint.GetFixNew()
}

func (r *runner) someProposal() (string, string, string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return "", "", ""
	}
	for _, e := range b.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetFixBasis() == "verified" {
			return m.GetGapId(), m.GetLocation(), m.GetFixNew()
		}
	}
	return "", "", ""
}

// ---- read-only INVOCATION: the surfaces the event gate is blind to ----
//
// These verbs RECORD NOTHING, so the event-type gate cannot see them: of 44 seat verbs, every
// one with zero fuzz coverage was of this kind, plus 7 of 9 root commands. What follows only
// CALLS them — the counting is done by the execution tally (noteExec), which observes every
// invocation the harness makes rather than keeping a second ledger of its own.

// readOnly invokes a record-nothing seat verb. r.exec tallies it; a non-zero exit shows up in
// the report's refusal count, which is where the dead `lens line of inquiry` drive and the misrouted
// `frontier` registration were both found.
func (r *runner) readOnly(verb, seatID string, extra ...string) {
	args := append([]string{verb, "--seat-id", seatID}, extra...)
	_, _ = r.exec(args...)
}

// readOnlySurfaces is the CENSUS, not a sample: every record-nothing surface a seat or the
// operator can reach, with the argv that exercises it. Adding a read-only verb without adding
// it here leaves it unfuzzed, and the gate below is what says so out loud.
//
// `setup` and `capture` are deliberately absent and that is stated rather than silently
// omitted: setup CREATES a run (the fuzz builds its run dir directly, so driving setup would
// mean fuzzing a different thing), and capture needs a transcript directory the harness has no
// analogue for. Both carry their own package tests. `hook` reads a JSON payload on stdin rather
// than argv and is covered by internal/cli's hook tests.
var readOnlySurfaces = [][]string{
	// Every SEAT's `show` with no view, which resolves that seat's DEFAULT. Only the merge's was
	// ever driven, so a regression in any other default was invisible. The seat id is what selects
	// the tree now, so it is what distinguishes these four rather than a role word in front.
	{"show", "--seat-id", "blue-respond-r1"},
	{"show", "--seat-id", "red-lens-r1-L1"},
	{"show", "--seat-id", "red-merge-r1"},
	{"show", "--seat-id", "judge-r1"},
	// The operator renders, over whatever shape the run actually reached.
	{"graph", "--format", "mermaid"},
	{"graph", "--format", "dot"},
	{"count-claims"},
	{"scorecard", "--chair", "red"},
}

// dashboardArgv is separate because `dashboard` is POSITIONAL — `dashboard <runDir>
// <transcript-dir>` — not `--run`. Discovered by this very census: the first version of it
// passed --run and dashboard failed on all 60 runs with a usage error, which is the census
// working (the surface had never been driven, so nothing here was known).
func dashboardArgv(runDir string) []string {
	return []string{"dashboard", runDir, runDir, "--seat-id", "operator"}
}

func containsFlag(argv []string, flag string) bool {
	for _, a := range argv {
		if a == flag || strings.HasPrefix(a, flag+"=") {
			return true
		}
	}
	return false
}

// sweepReadOnly runs the census against a finished run and reports the first surface that
// fails. Run AFTER the debate so the state is arbitrary rather than empty — an empty run
// proves only that nothing crashed on nothing.
func sweepReadOnly(bin, runDir string) string {
	// The transcript dir is the run dir here: dashboard tolerates finding no agent-*.jsonl,
	// and what is under test is that it RENDERS whatever board shape the run reached.
	if out, err := tracked(bin, dashboardArgv(runDir)...); err != nil {
		return "read-only surface `dashboard` failed on a real run shape:\n" + truncate(string(out))
	}
	for _, argv := range readOnlySurfaces {
		args := append(append([]string{}, argv...), "--run", runDir)
		// THE OPERATOR SAYS SO TOO. graph, count-claims, scorecard and the rest are the operator's,
		// and a surface scoped to whoever is asking has nothing to show a caller who has not said.
		if !containsFlag(argv, "--seat-id") {
			args = append(args, "--seat-id", "operator")
		}
		if out, err := tracked(bin, args...); err != nil {
			return "read-only surface `" + strings.Join(argv, " ") + "` failed on a real run shape:\n" + truncate(string(out))
		}
	}
	return ""
}

// seatFor returns a seat id bound to the role, since `show` is role-scoped like any verb.
func seatFor(role string) string {
	switch role {
	case "lens":
		return "red-lens-r1-L1"
	case "merge":
		return "red-merge-r1"
	case "blue":
		return "blue-respond-r1"
	default:
		return "judge-r1"
	}
}

func sumCounts(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

// THE INQUIRY-SUPPORT VOTE IS GONE, and it is the verb that went, not just the fuzz action.
//
// It cast a per-line verdict — supported / weakened / unsupported / absent — on every line of
// inquiry. `record.UnvotedInquiries` fed it and was deleted deliberately with nothing replacing
// it: every one of those vocabularies made PRESENCE the question, and a line is on the worklist
// because the record says so, not because a seat voted it there. What is genuinely open — whether
// blue's body delivered the research a line claims — is an ORDINARY GAP now, with the id, the
// grade, the blue duty and the PASS gate every gap already has.
//
// A fuzz action driving a retired verb is not coverage. It would spend every run issuing a command
// the tool refuses, and the refusals would read as the tool working.

// fieldStr reads a named string field off an event's body, by DESCRIPTOR.
//
// The basisFields table is keyed on field names, which is the right shape for it — the same rule
// stated once per (event, field, value) triple. Reading them off typed bodies with a switch would
// be seven near-identical arms, and reading them out of a map is what the migration removed.
//
// THE MISS IS LOUD. A name the schema does not carry FAILS THE TEST rather than returning "", and
// that is the whole reason this is not a bare lookup: a stale key silently never matches, so the
// gate would report "the report never says so" for a field that no longer exists — or, worse,
// report nothing at all and read as covered. record.go's keyFields carries the same warning about
// the same shape.
func fieldStr(t *testing.T, e *record.Event, name string) string {
	t.Helper()
	body, ok := recordpb.Body(e)
	if !ok {
		return ""
	}
	m := body.ProtoReflect()
	fd := m.Descriptor().Fields().ByName(protoreflect.Name(name))
	if fd == nil {
		t.Fatalf("fuzz: %s has no field %q — the basisFields table names a field the schema does not "+
			"carry, so its rule can never match and the gate would read as covered",
			m.Descriptor().FullName(), name)
	}
	if !m.Has(fd) {
		return ""
	}
	return m.Get(fd).String()
}

// seatOfRole is a dispatched seat of the given role, for drives that iterate roles. The role is no
// longer part of an invocation; the seat id is what selects the tree it runs in.
func seatOfRole(role string) string {
	return map[string]string{
		"blue": "blue-respond-r1", "lens": "red-lens-r1-L1",
		"merge": "red-merge-r1", "bench": "judge-r1",
	}[role]
}

// fuzzClasses is the vocabulary every fuzz run declares. Deliberately SMALL: the fuzz mints
// under one coined slug, and the registry is here so the coining path has a real neighbour to
// name and the mint has a real registry to be checked against — not to give the generator a
// menu to pick from, which would test the picking rather than the check.
var fuzzClasses = []string{"self-attestation", "policy-without-mechanism", "metric-conflation"}

// TestFuzzUnverifiedPath drives the UNVERIFIED terminal verdict deterministically.
//
// IT USED TO BE LUCK. debate.js computes UNVERIFIED only for a run that is not halted, whose red
// never passed, and which did NOT stop at the ceiling — the judged-deadlock exit with gaps still
// open. In the random sweep that happens about once in sixty runs, because a bench that disposes
// its whole docket clears the board and the engine grants red a further sitting instead, which
// then passes. So `outcome --as UNVERIFIED` was an enum value the coverage gate demanded and the
// sweep supplied by chance: it passed at N=60 with 2 occurrences and 1 at N=40, then failed at
// N=40 with none. Roughly a one-in-three flake, and it had been one all along.
//
// coverage_test.go already refused the other answer, in writing: `outcome --as UNVERIFIED` was
// once exempted on a claim that the engine had no path to it, a real run falsified that, and the
// note left behind says an unreachable value should make the sweep SAY so rather than earn an
// exemption explaining why it never was. So this drives it, the way TestFuzzHaltPath drives
// HALTED — a terminal shape too disruptive for the random sweep gets a dedicated run, and its
// invocations land in the same package-level tally the gate reads.
func TestFuzzUnverifiedPath(t *testing.T) {
	bin := buildBinary(t)
	wrapped := debateWrapped(t)
	runDir, _ := os.MkdirTemp("", "fuzz-unverified-")
	defer os.RemoveAll(runDir)
	// THE SAME RUN THE SWEEP BUILDS, because driveDebate is only half of one. Without the seeded
	// report a lens finding has no anchor quote to attach to and is refused, so red mints nothing
	// and round 3 is a FAIL with an empty gaps array — the engine's degenerate-merge refusal,
	// which is correct and is not this path. Without the class registry every `mint` is refused
	// for an unknown class. runOne does both before it drives; so does this.
	_ = os.MkdirAll(filepath.Join(runDir, "blue"), 0o755)
	_ = os.WriteFile(filepath.Join(runDir, "blue", "report.md"),
		[]byte("# § fuzz\n\nA § fuzz sentence to anchor findings.\n\nThe cost is rising over time.\n"), 0o644)
	if err := record.StageForRun(runDir, fuzzClasses...); err != nil {
		t.Fatalf("staging the class registry: %v", err)
	}
	r := &runner{bin: bin, runDir: runDir, rng: newLockedRand(1), registered: map[string]bool{}, forceUnverified: true}

	result, settledErr := driveDebate(r, wrapped)
	if settledErr != "" {
		t.Fatalf("unverified run did not settle cleanly: %s", settledErr)
	}
	if v, _ := result["verdict"].(string); v != "UNVERIFIED" {
		t.Fatalf("expected verdict UNVERIFIED, got %q — the terminal path this test exists to drive did not run", v)
	}
	// AND THE RECORD SAYS SO, not just the returned map. This is the assertion the enum gate
	// actually depends on: `outcome --as UNVERIFIED` has to be a value the tool WROTE.
	//
	// It asserted `len(r.openGaps()) > 0` first, on the reasoning that UNVERIFIED-with-nothing-
	// open is a contradiction. That reasoning is debate.js's and it holds where debate.js
	// applies it — at the decision — but not here: the bench disposes the docket in its terminal
	// sitting, so an empty board AFTER the run is ordinary. Measured, it failed 12 of 12 while
	// the verdict itself was right every time, which makes it an assertion about the wrong
	// moment rather than a defect it caught.
	board, err := record.BoardState(runDir)
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	recorded, ended := "", ""
	for _, e := range board.Events {
		if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok {
			recorded = strings.ToUpper(recordpb.Word(o.GetVerdict()))
			ended = o.GetEnded()
		}
	}
	if recorded != "UNVERIFIED" {
		t.Fatalf("the record carries outcome %q, not UNVERIFIED — the enum value this test exists to write was not written", recorded)
	}
	// THE REASON, not only the stamp. UNVERIFIED is reachable two ways in principle and this
	// test drives exactly one of them: the judged-deadlock exit. Without this, a run that
	// reached the same word down some other path would satisfy the test and teach nothing.
	if ended != "deadlock" {
		t.Fatalf("outcome ended %q, expected deadlock — UNVERIFIED was reached, but not by the path this drives", ended)
	}
}

// THE GUARD THAT WOULD HAVE NAMED #637's 29 RUNS. Every case below reached the sweep's
// distribution as `verdicts[""]++` before this change — a bucket that counts like a verdict,
// sorts like a verdict, and satisfies a coverage gate like a verdict, while meaning "this side
// could not read the result".
//
// The gate is on readVerdict rather than on a driven debate deliberately: the miss is a
// PROPERTY OF THE READ, not of any particular run, and a test that had to reproduce the
// upstream conditions would be exactly as unreliable as the symptom it chases.
func TestReadVerdictRefusesAResultThatCarriesNoVerdict(t *testing.T) {
	for _, tc := range []struct {
		name    string
		result  map[string]any
		wantErr string // a substring the message must carry
	}{
		{
			// The measured shape: driveDebate returned (nil, "") because the promise fulfilled
			// with something that was not an object, and every caller read that as success.
			name:    "a nil result is not a settled debate",
			result:  nil,
			wantErr: "no verdict",
		},
		{
			name:    "a result carrying other keys but no verdict",
			result:  map[string]any{"rounds": int64(3), "runDir": "/tmp/x"},
			wantErr: "no verdict",
		},
		{
			name:    "a verdict that is not a string",
			result:  map[string]any{"verdict": 7},
			wantErr: "non-string verdict",
		},
		{
			name:    "an explicitly empty verdict",
			result:  map[string]any{"verdict": ""},
			wantErr: "EMPTY verdict",
		},
	} {
		got, err := readVerdict(tc.result)
		if err == nil {
			t.Errorf("%s: readVerdict returned %q and NO error — the miss is still indistinguishable "+
				"from a terminal verdict", tc.name, got)
			continue
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("%s: error = %q, want it to mention %q", tc.name, err, tc.wantErr)
		}
		if got != "" {
			t.Errorf("%s: readVerdict returned %q alongside an error; a refused read yields nothing", tc.name, got)
		}
	}
}

// AND THE FOUR REAL VERDICTS STILL PASS. debate.js computes exactly these; a guard that also
// rejected one of them would turn a working sweep red and be indistinguishable, from the
// outside, from the bug it replaced.
func TestReadVerdictAcceptsEveryVerdictDebateJSComputes(t *testing.T) {
	for _, want := range []string{"HALTED", "VERIFIED", "CEILING", "UNVERIFIED"} {
		got, err := readVerdict(map[string]any{"verdict": want, "rounds": int64(1)})
		if err != nil {
			t.Errorf("readVerdict rejected %q, which debate.js line 1166 can produce: %v", want, err)
		}
		if got != want {
			t.Errorf("readVerdict(%q) = %q", want, got)
		}
	}
}

// The keys are named in the failure, and named STABLY — two failures of one shape must produce
// one message, or a reader cannot tell a recurring defect from a family of them.
func TestSortedKeysIsStableAcrossMapIterationOrder(t *testing.T) {
	m := map[string]any{"runDir": "", "verdict": "", "rounds": 0, "lanes": 0, "deadlocked": false}
	first := fmt.Sprint(sortedKeys(m))
	for i := 0; i < 50; i++ {
		if got := fmt.Sprint(sortedKeys(m)); got != first {
			t.Fatalf("sortedKeys is order-dependent: %s vs %s", got, first)
		}
	}
	if first != "[deadlocked lanes rounds runDir verdict]" {
		t.Errorf("sortedKeys = %s, want sorted order", first)
	}
}
