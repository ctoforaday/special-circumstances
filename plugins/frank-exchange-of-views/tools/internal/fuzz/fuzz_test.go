package fuzz

// Fuzz the ACTUAL debate.js orchestrator against the REAL feov-record binary. goja runs the
// unmodified script (harness faked); each agent() call is a seat that makes coherent, random,
// VALID tool calls into a real run directory and returns a randomised envelope to drive
// debate.js's own branches (verdict, gap counts, rounds, petitions). A failing run is a real
// finding (in debate.js, the tool, or verify), reproducible from its seed.
//
// COVERAGE CONTRACT. envelopeFor drives every eligible seat to exercise its whole verb surface,
// not a happy path: lens (cite/finding/avenue/friction), merge (position/closing/
// mint/close incl. closed_with_regression/regrade any axis/
// dispute-respond/spot-check/verdict/petition), blue (position/closing/confidence/dispute
// across all four dimensions/manifest-row/avenue/revision/retire/petition), bench
// (opinion/outcome incl. --exhausted/--deadlocked/certify/assemble/petition-rule). The
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
// Run: go test ./internal/fuzz -run TestFuzzDebate -count=1   (respects -short by shrinking N).
// Confidence sweep: FUZZ_N=1000 go test ./internal/fuzz -run TestFuzzDebate -timeout 1200s.
// FUZZ_C overrides concurrency (runs are subprocess-bound, so the default oversubscribes cores).

import (
	"encoding/json"
	"fmt"
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
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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

// runner shells the real binary for one run's seats. runDir + bin are fixed per fuzz iteration.
type runner struct {
	bin, runDir string
	rng         *rand.Rand
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
	// #111: every model an agent() call carried, one per dispatch ("unset" if absent). The tier
	// oracle asserts all equal the configured tier — map-free, needs no bulk-seat list here.
	models []string
	// petition docket: petitions a party seat RAISED (event emitted + envelope entry), awaiting
	// the judge-petition sitting hearPetitions dispatches next. Each: {who, class}. Mirrors the
	// disputes machinery (raised) so the fuzz drives the petition/petition-rule path end to end.
	petitioned []map[string]any
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

// cmd is a small fluent builder for a seat verb — `<role> <verb> --seat-id <seatID> …` (exec
// appends --run). It collapses the conditional-flag arg-slice boilerplate: set() always adds a
// flag, on() adds it with pct% probability, bare() adds a boolean flag. run() shells the binary.
type cmd struct {
	r    *runner
	args []string
}

func (r *runner) do(role, verb, seatID string) *cmd {
	return &cmd{r: r, args: []string{role, verb, "--seat-id", seatID}}
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

// dialectic emits the round's transcript onto the record the way Stage 1 seats do — a position
// narrative and a closing per open gap — plus, at random, a regrade, a lineage-carrying
// close-with-regression, and (blue) a grade dispute / (merge) its answer. Unique prose per act
// so the report oracle can prove each one actually rendered.
func (r *runner) dialectic(role, seatID string, open []string) {
	_, _ = r.exec(role, "position", "--seat-id", seatID, "--reason", "narrative from "+seatID)
	for _, id := range open {
		_, _ = r.exec(role, "closing", "--seat-id", seatID, "--id", id, "--reason", "closing-for-"+id+"-by-"+seatID)
	}
	if role == "blue" {
		r.emitConfidence(seatID) // blue calibrates its claims every round
	}
	if role == "merge" && len(open) > 0 && r.coin(30) {
		id := open[r.rng.Intn(len(open))]
		dim := pick(r.rng, regradeDims) // move any grade axis, not only likelihood
		_, _ = r.exec("merge", "regrade", "--seat-id", seatID, "--id", id, "--reason", "regrade-basis-for-"+id, "--"+dim, r.g())
	}
}

// raiseDisputes is blue's Stage-2 contest path: EMIT a dispute event (evidence on the record)
// and return the ROUTING REF ({gap_id, dimension, proposed}) for the envelope's grade_disputes.
// The raised set is remembered so red answers it next round (drives the docket machinery). Unique
// evidence prose per gap so the report oracle can prove it rendered.
func (r *runner) raiseDisputes(seatID string, open []string) []map[string]any {
	var refs []map[string]any
	for _, id := range open {
		if !r.coin(40) {
			continue
		}
		proposed := r.g()
		dim := pick(r.rng, disputeDims) // contest each dimension over a run, not only impact
		if _, err := r.exec("blue", "dispute", "--seat-id", seatID, "--id", id, "--dimension", dim, "--proposed", proposed, "--reason", "dispute-evidence-for-"+id+"-by-"+seatID); err != nil {
			continue
		}
		ref := map[string]any{"gap_id": id, "dimension": dim, "proposed": proposed}
		refs = append(refs, ref)
		r.raised = append(r.raised, ref)
	}
	return refs
}

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
		if _, err := r.exec("merge", "dispute-respond", "--seat-id", seatID, "--id", id, "--dimension", dim, "--as", resp, "--reason", "respond-rationale-for-"+id+"-by-"+seatID); err != nil {
			continue
		}
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
	if _, err := r.exec(role, "petition", "--seat-id", seatID, "--petition-class", class, "--reason", basis, "--relief", "fuzz relief"); err != nil {
		return arr()
	}
	r.petitioned = append(r.petitioned, map[string]any{"who": seatID, "class": class})
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
		_, _ = r.exec("bench", "petition-rule", "--seat-id", seatID, "--petitioner", who, "--petition-class", class, "--as", ruling, "--reason", opinion)
		rulings = append(rulings, map[string]any{"petitioner": who, "class": class, "ruling": ruling, "relief": opinion})
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
		_, _ = r.exec("bench", "halt", "--seat-id", seatID, "--reason", haltOpinion)
		env["halt"] = map[string]any{"opinion": haltOpinion}
	}
	return env
}

func (r *runner) exec(args ...string) (string, error) {
	cmd := exec.Command(r.bin, append(args, "--run", r.runDir)...)
	out, err := cmd.CombinedOutput()
	noteExec(args, err)
	return string(out), err
}

func (r *runner) register(role, seatID string) {
	if r.registered[seatID] {
		return
	}
	r.registered[seatID] = true
	_, _ = r.exec(role, "register", "--seat-id", seatID)
}

var grades = []string{"low", "low-medium", "medium", "medium-high", "high"}

func (r *runner) g() string { return grades[r.rng.Intn(len(grades))] }

var confGrades = []string{"high", "medium", "low"}

// emitConfidence records blue's per-claim calibration — the NON-AUTHORITATIVE signal wired in
// 0.13.0. Unique labels per act so the report oracle can prove each one actually rendered in the
// "Blue's confidence self-assessment" section (and never leaked into the risk matrix).
func (r *runner) emitConfidence(seatID string) {
	for i := 0; i <= r.rng.Intn(2); i++ {
		claim := fmt.Sprintf("conf-claim-%d-by-%s", i, seatID)
		_, _ = r.exec("blue", "confidence", "--seat-id", seatID, "--claim", claim, "--confidence", confGrades[r.rng.Intn(len(confGrades))])
	}
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
	kind := checkKinds[r.rng.Intn(len(checkKinds))]
	if directive == dirProve || directive == dirProveDrifts {
		// The demand and the answer must agree: a gap settled by computing is minted as a
		// COMPUTATION check, so the tool's own close guard is in force alongside the scenario.
		kind = "computation"
	}
	args := []string{"--json", "merge", "mint", "--seat-id", seatID, "--problem", "fuzz problem", "--check-kind", kind,
		"--check", "acc", "--fix", directive, "--likelihood", r.g(), "--impact", r.g(),
		// --location is the anchor ESTOPPEL keys on and the sweep never passed it; --existence
		// is grading v2's verified/suspected split; --key is mint's crash-retry idempotency
		// handle; --reason is the prose channel. Four fields on the most consequential verb in
		// the tool, none of them ever exercised by a run.
		"--location", "§ fuzz — " + directive[:12],
		"--existence", pick(r.rng, []string{"verified", "suspected"}),
		"--reason", "fuzz: the argument for raising this"}
	if !r.classMade {
		r.classMade = true
		args = append(args, "--class-new", "fuzzcls", "--definition", "d", "--neighbor", "verification-gap", "--distinguisher", "q")
	} else {
		args = append(args, "--class", "fuzzcls")
	}
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
		args = append(args, "--cx", r.g())
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
				args = append(args, "--fix-old", fixOld, "--fix-new", fixNew)
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
	out, err := r.exec("merge", "show", "--view", "findings")
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
	out, err := r.exec("merge", "show", "--view", "board")
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
			_, _ = r.exec("blue", "prove", "--seat-id", "blue-respond-r1", "--location", "§ fuzz",
				"--script", name, "--answers", id, "--reason", "fuzz: settling the computation check")
		}
	}
	// A regression close carries lineage forward: it mints a successor and closes WITH it
	// (record.go requires --successor for closed_with_regression). Only allowed on the first close
	// pass, so the successors it spawns are plain-closed on a later pass and the loop terminates.
	if allowReg && r.coin(25) {
		if succ := r.mint(seatID); succ != "" {
			if _, err := r.exec("merge", "close", "--seat-id", seatID, "--id", id, "--as", "closed_with_regression",
				"--successor", succ, "--reason", "fuzz regression close", "--anchor-seat", seatID, "--anchor-tool", "fuzz", "--anchor-target", "rec"); err == nil {
				return
			}
		}
	}
	// A CARRIED closure restates last round's verification rather than asserting a fresh act,
	// and takes --carried-from instead of an anchor. Never driven, though the attestation audit
	// treats it as a distinct case and skips it deliberately.
	if r.coin(20) {
		if _, err := r.exec("merge", "close", "--seat-id", seatID, "--id", id, "--as", "closed",
			"--reason", "fuzz: carried from the prior round", "--carried-from", "1"); err == nil {
			return
		}
	}
	// THE WHOLE CLOSURE VOCABULARY, not just `closed`. #342 closed the set, so the
	// enum-coverage sweep now demands every value be reached — and three of them
	// (rebuttal_sustained, risk_accepted, routed_to_infrastructure) had never been driven by
	// anything, on either closing verb, in the tool's life.
	//
	// `amends_prior` takes --supersedes: it names a defect found BETWEEN two repairs that each
	// closed clean earlier, so it needs a prior closure to amend.
	if prior := r.closedGapIDs(); len(prior) > 0 && r.coin(20) {
		if _, err := r.exec("merge", "close", "--seat-id", seatID, "--id", id, "--as", "amends_prior",
			"--supersedes", prior[r.rng.Intn(len(prior))], "--reason", "fuzz: found between two clean repairs",
			"--anchor-seat", seatID, "--anchor-tool", "fuzz", "--anchor-target", "rec"); err == nil {
			return
		}
	}
	as := pick(r.rng, []string{"closed", "rebuttal_sustained", "risk_accepted", "routed_to_infrastructure"})
	_, _ = r.exec("merge", "close", "--seat-id", seatID, "--id", id, "--as", as, "--reason", "fuzz close as "+as,
		"--anchor-seat", seatID, "--anchor-tool", "fuzz", "--anchor-target", "rec")
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
var avenueStatus = []string{"proposed", "pursued", "abandoned", "declined", "deferred"}
var obsKind = []string{"note", "checked-held"}

// disputeDims is the full grade-dimension domain — the fuzz must contest each, not only impact.
var disputeDims = []string{"severity", "likelihood", "impact", "complexity_cost"}

// petitionClasses is the full petition-class domain (debate.js PETITIONS enum).
var petitionClasses = []string{"ethical", "safety", "integrity", "constitutional"}

// regradeDims are the grade axes a regrade can move (each maps to its flag below).
var regradeDims = []string{"severity", "likelihood", "impact", "cx"}

func pick[T any](rng *rand.Rand, xs []T) T { return xs[rng.Intn(len(xs))] }

// extras fires a RANDOM subset of a role's REMAINING verb surface, so no two fuzz paths
// look alike and every seat exercises far more than its happy path. Only verbs that keep
// the oracles intact (verify passes, the run terminates, the report renders) are here;
// terminal/verdict-shaping acts (halt, certify, outcome, verdict) stay in the structured
// cases above. Reference-taking verbs are gated on a referent existing.
func (r *runner) extras(role, seatID string, open []string) {
	r.maybe(50, func() { r.do(role, "friction", seatID).set("--reason", "fuzz friction from "+seatID).run() })
	// avenue carries an optional --method; feed it sometimes so that flag is exercised too.
	// #246: an avenue now has an id and a LIFECYCLE. Propose, then sometimes move it — the
	// move is the path the old one-shot append could not record at all (measured: 0 of 86
	// events across six runs ever changed status).
	avenue := func(role string) {
		// A DECLINED OR ABANDONED avenue requires --reason (record.go: an unexplained
		// non-pursuit is the decoration this verb exists to refuse). Without it two of the
		// three statuses were rejected on every call, so only `pursued` ever reached the
		// record while the verb gate read as covered. Found by the execution tally
		// (blue avenue: 48 of 72 calls refused).
		// A FATE-CARRYING LINE IS BORN `proposed`, so the ruling cycle can run on it: red rules
		// it, blue answers. Creating one directly as `pursued` or `declined` skips the cycle
		// entirely — which the oracle caught, reporting endorsed avenues that "ended declined"
		// when they had simply never been through a ruling at all.
		//
		// An UNDIRECTED line keeps the random status: proposing something already declined is a
		// real shape (blue weighed it and did not start), and it carries no fate for red to
		// rule from, so the ruling cycle and the oracle both skip it by construction.
		line, st := pick(r.rng, avenueFates)+" ("+seatID+")", "proposed"
		if r.coin(30) {
			line, st = "fuzz undirected avenue "+seatID, pick(r.rng, avenueStatus)
		}
		c := r.do(role, "avenue", seatID).set("--status", st).set("--line", line).
			set("--hypothesis", "fuzz: what would be true if "+seatID+" paid off").
			on(50, "--method", "fuzz-method")
		if st != "pursued" && st != "proposed" {
			c = c.set("--reason", "fuzz: why this line was not pursued")
		}
		c.run()
	}
	switch role {
	case "lens":
		// NO avenue DRIVE HERE: the lens role has no avenue verb (register/finding/
		// cite/friction/show). This called it 183 times per sweep, every one refused, while
		// the verb gate stayed green on blue's avenue events — a dead drive that read as
		// coverage. Found by the execution tally (lens avenue: 183 of 183 refused).
	case "merge":
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
			r.do("merge", "spot-check", seatID).set("--ids", closed[r.rng.Intn(len(closed))]).
				set("--notes", "fuzz: re-read the closure record; the anchor still resolves").run()
		} else {
			r.do("merge", "spot-check", seatID).bare("--none").
				set("--reason", "fuzz: the archive was empty at round start").run()
		}
		// RED RULES ON BLUE'S DIRECTIONS (#246) — the verb red never had. Across six runs blue
		// rejected 18 of its own 86 avenues and red rejected none, because it could not.
		// RED RULES EVERY UNRULED AVENUE, and the ruling follows the line it was proposed as.
		//
		// It used to rule ONE random avenue at 45% with a random ruling, so a ruling had no
		// relation to what was proposed and no consequence for what happened next: an
		// out-of-scope line could be pursued and an endorsed one abandoned, and the record kept
		// both halves while joining them nowhere. The avenue's own --line now carries its
		// intended fate, exactly as a gap's --fix carries its scenario.
		r.ruleOpenAvenues(seatID)
		// near-match is the screen red runs BEFORE minting, to catch a reopen. Read-only, so it
		// left no event and no gate saw it — while the merge prompt calls it every round.
		r.maybe(40, func() {
			r.readOnly("merge", "near-match", seatID, "--candidate", "fuzz candidate problem text for screening", "--location", "§ fuzz")
		})
	case "blue":
		r.maybe(45, func() { avenue("blue") })
		// NO RANDOM AVENUE MOVE HERE. answerAvenueRulings owns moves now: it answers red's
		// ruling, comply or contest, one decision per avenue. A second writer sliding statuses
		// at random fought it — the oracle caught it immediately, 24 of 60, reporting endorsed
		// avenues that ended declined and contests the record never recorded because the move
		// landed before the ruling did. Two writers disagreeing about a fate is the same defect
		// the edit drive had, one axis over.
		// THE CITATION AXIS (#256), driven end to end through the real binary: `blue cite` fetches
		// the source through the run cache and splices an INVISIBLE, IMMORTAL <!--cite:c-…--> anchor
		// at the quoted sentence. --location must appear VERBATIM in blue/report.md, so it quotes the
		// seeded "§ fuzz" text the same way lens finding does. A --key from a small space exercises
		// the retry short-circuit; a shared URL across seats exercises the cache HIT path (fetch-once),
		// while the per-seat path exercises the MISS path.
		r.maybe(55, func() {
			url := sourceURL("/blue-shared")
			if r.coin(50) {
				url = sourceURL("/" + seatID)
			}
			r.do("blue", "cite", seatID).
				set("--location", "§ fuzz").
				set("--url", url).
				set("--title", "fuzz source "+seatID).
				on(50, "--claim", "fuzz cited claim "+seatID).
				on(40, "--key", fmt.Sprintf("C%d", 1+r.rng.Intn(2))).
				on(50, "--reason", "fuzz: why this source backs the claim").
				run()
		})
		// An UNREACHABLE source is an unusable citation: the cite must be REJECTED and the failure
		// auto-logged as friction. Driving it here proves the reject path never wedges the run.
		r.maybe(15, func() {
			r.do("blue", "cite", seatID).
				set("--location", "§ fuzz").
				set("--url", "http://127.0.0.1:1/unreachable").
				set("--title", "unreachable "+seatID).
				run()
		})
		// THE ONLY WRITE PATH to blue/report.md, driven end to end through the real binary.
		// It swaps the seeded edit-target sentence between two phrasings, so a valid unique
		// --old exists whichever way the previous edit left the file — no state to thread.
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
			r.do("blue", "edit", seatID).
				set("--old", oldSpan).
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
			pv := r.do("blue", "prove", seatID).
				set("--location", "§ fuzz").
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
		r.maybe(35, func() { r.readOnly("blue", "claim-index", seatID) })
		r.maybe(40, func() { r.do("blue", "revision", seatID).set("--reason", "fuzz revision").run() })
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
				r.do("blue", "retire", seatID).set("--claim", claim).set("--reason", "fuzz: the claim went with the edit").
					on(50, "--superseded-by", "fuzz replacement claim").run()
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
			r.do("blue", "manifest-row", seatID).set("--id", id).
				set("--row", "fuzz: figures recomputed, universals enumerated, acceptance check run — held").
				on(50, "--reason", "fuzz: the receipt's argument").run()
		}
	case "bench":
		// run-end certification statement — the bench's, additive (a recorded opinion).
		r.maybe(45, func() {
			r.do("bench", "certify", seatID).set("--reason", "fuzz certification statement from "+seatID).run()
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
		r.emitConfidence(seatID) // round-0 calibration
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
				r.closeGap(seatID, id, r.coin(25)) // the regression-close variant is still sampled
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

		open := r.openGaps()
		if len(open) == 0 {
			r.dialectic("merge", seatID, nil)
			// The merge's TERMINAL act on a PASS: checkpoints the event log to the recovery mirror.
			_, _ = r.exec("merge", "verdict", "--seat-id", seatID, "--as", "PASS")
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "petitions": r.maybePetition("merge", seatID), "friction": arr()}
		}

		// Something is unrepaired, so the round FAILs.
		r.dialectic("merge", seatID, open)
		var gaps []any
		for _, id := range open {
			gaps = append(gaps, map[string]any{"id": id, "supersedes": arr()})
		}
		if len(gaps) == 0 {
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "petitions": r.maybePetition("merge", seatID), "friction": arr()}
		}
		// The merge's terminal act on a FAIL too: the checkpoint is what protects the event log
		// from a stray git operation mid-round, and it was only ever driven on a PASS — so the
		// `FAIL` half of a two-value enum had never been recorded by anything.
		_, _ = r.exec("merge", "verdict", "--seat-id", seatID, "--as", "FAIL")
		return map[string]any{"verdict": "FAIL", "gaps": gaps, "closures": arr(), "dispute_responses": responses, "petitions": r.maybePetition("merge", seatID), "friction": arr()}

	case strings.HasPrefix(seatID, "blue-respond"):
		r.register("blue", seatID)
		open := r.openGaps()
		r.dialectic("blue", seatID, open) // blue's position, closings, confidence
		r.extras("blue", seatID, open)
		// ONE DECISION PER GAP, TAKEN FROM THE GAP. Each open gap carries the scenario it was
		// minted with; blue reads it back off the board and does what it says. This replaces a
		// blanket 40%-per-gap dispute roll plus a 50% verbatim-apply coin, neither of which
		// looked at what red had actually asked for.
		r.answerAvenueRulings(seatID)
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
				_, _ = r.exec("bench", "opinion", "--seat-id", seatID, "--id", id, "--as", "carried",
					"--principle", "thoroughness", "--tension", "cost", "--review-flag", "false",
					"--reason", "fuzz: carried with a stated direction for "+id)
			}
			if r.scenarioOf(id) == dirDisputeLost {
				// EVERY CLOSING DISPOSITION, not just `closed` (#342). The bench shares red's
				// closure vocabulary now, and the sweep must reach all of it — bench opinion
				// had driven exactly one closing word.
				disp = pick(r.rng, []string{"closed", "rebuttal_sustained", "risk_accepted", "routed_to_infrastructure", "amends_prior"})
				_, _ = r.exec("bench", "opinion", "--seat-id", seatID, "--id", id, "--as", disp,
					"--principle", "correctness", "--tension", "cost", "--review-flag", "false", "--reason", "opinion-rationale-for-"+id)
			}
			res = append(res, map[string]any{"gap_id": id, "resolution": disp, "rationale": "fuzz"})
		}
		return map[string]any{"resolutions": res, "deadlock": false, "friction": arr()}

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
		oargs := []string{"bench", "outcome", "--seat-id", seatID, "--as", verd}
		// the terminal-outcome modifiers — a non-VERIFIED end may be by safety ceiling or deadlock
		// (not on a halt, whose outcome stands alone).
		if !r.forceHalt && verd != "VERIFIED" && r.coin(40) {
			oargs = append(oargs, "--exhausted")
		}
		if !r.forceHalt && verd == "UNVERIFIED" && r.coin(40) {
			oargs = append(oargs, "--deadlocked")
		}
		_, _ = r.exec(oargs...)
		_, _ = r.exec("bench", "assemble")
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
				_, _ = r.exec("lens", "verify", "--seat-id", seatID, "--claim", "fuzz claim "+seatID,
					"--reference", "https://fuzz.invalid/"+seatID, "--trust", confGrades[r.rng.Intn(len(confGrades))], "--access-date", "2026-07-24")
				// --key from a small space so a repeated dispatch exercises retry idempotency.
				_, _ = r.exec("lens", "finding", "--seat-id", seatID, "--key", fmt.Sprintf("F%d", 1+r.rng.Intn(2)),
					"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(), "--location", "§ fuzz", "--reason", "fuzz finding")
				// Red verifies a cited source by reading the CACHED bytes (#256): the same
				// `fetch` any seat uses. Driving it here is what makes the cache path — miss,
				// store, hit — real in the fuzz rather than unit-tested only.
				_, _ = r.exec("fetch", "--url", sourceURL("/"+seatID))
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
	p := filepath.Join("..", "..", "..", "skills", "research-protocol", "scripts", "debate.js")
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
}

// installAgent wires r as the seat backend on vm: it parses the seat id from each agent() prompt
// (falling back to the label), records the model tier the dispatch carried (#111 oracle), and
// resolves the promise with r's envelope for that seat.
func (r *runner) installAgent(vm *goja.Runtime) {
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
		env := r.envelopeFor(seatID, prompt)
		p, resolve, _ := vm.NewPromise()
		resolve(vm.ToValue(env))
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
		[]byte(`{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","maxRounds":"3","lanes":"1"}`), 0o644)

	loop := eventloop.NewEventLoop()
	loop.Run(func(vm *goja.Runtime) {
		vm.Set("args", map[string]any{
			"topic": "fuzz", "runDir": r.runDir, "binDir": binDir(r.bin),
			"lanes": 1, "laneFloorOverride": "fuzz", "maxRounds": 3,
			// #111: both tiers are REQUIRED — nil refuses dispatch. Both haiku so the tier oracle
			// expects every dispatched seat to carry exactly "haiku".
			"model": "haiku", "judgmentModel": "haiku",
		})
		if _, err := vm.RunString(preamble); err != nil {
			settledErr = "preamble: " + err.Error()
			return
		}
		r.installAgent(vm)
		if _, err := vm.RunString(wrapped); err != nil {
			settledErr = "run: " + err.Error()
		}
	})
	if settledErr != "" {
		return nil, settledErr
	}
	loop.Run(func(vm *goja.Runtime) {
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
			if m, ok := pr.Result().Export().(map[string]any); ok {
				result = m
			}
		}
	})
	return result, settledErr
}

func runOne(wrapped, bin string, seed int64) outcome {
	runDir, _ := os.MkdirTemp("", "fuzz-run-")
	// A lens finding anchors into blue/report.md and is rejected unless its --location
	// quote is present (slice 1b). Seed a report carrying the fuzzer's finding quote
	// ("§ fuzz") so findings are accepted and the coverage gate sees them.
	_ = os.MkdirAll(filepath.Join(runDir, "blue"), 0o755)
	// The trailing sentence is the EDIT TARGET: `blue edit` swaps it between two fixed
	// phrasings, so every round has a valid unique span to replace whichever way the last
	// edit left it, while the anchor quote above stays untouched for findings and cites.
	_ = os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte("# § fuzz\n\nA § fuzz sentence to anchor findings.\n\nThe cost is rising over time.\n"), 0o644)
	r := &runner{bin: bin, runDir: runDir, rng: rand.New(rand.NewSource(seed)), registered: map[string]bool{}}

	res := outcome{seed: seed, runDir: runDir}
	result, settledErr := driveDebate(r, wrapped)
	if settledErr != "" {
		res.err = settledErr
		return res
	}
	if s, ok := result["verdict"].(string); ok {
		res.verdict = s
	}
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
	if out, err := tracked(bin, "verify", "--run", runDir); err != nil {
		res.err = "verify FAILED:\n" + truncate(string(out))
		return res
	}
	// Oracle 1b: the JSON views the operator side reads in json-mode must exit 0 and parse.
	// `board` is already exercised above; `findings`/`friction` are JSON by name; `debate --json`
	// is the structured debate the capture audits count sections from. A broken view is what
	// would silently blank a dashboard tile or make an audit read an empty transcript.
	for _, v := range []string{"findings", "friction"} {
		out, err := tracked(bin, "merge", "show", "--view", v, "--run", runDir)
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show --view " + v + " did not return valid JSON:\n" + truncate(string(out))
			return res
		}
	}
	{
		out, err := tracked(bin, "merge", "show", "--view", "debate", "--json", "--run", runDir)
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show --view debate --json did not return valid JSON:\n" + truncate(string(out))
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
		if out, err := tracked(bin, role, "show", "--view", "debate", "--run", runDir); err != nil {
			res.err = role + " show --view debate failed:\n" + truncate(string(out))
			return res
		}
	}
	if ids := mintedGapIDs(runDir); len(ids) > 0 {
		for _, role := range []string{"blue", "lens", "bench"} {
			_, _ = tracked(bin, role, "show", "--view", "changes", "--id", ids[0], "--run", runDir)
		}
	}
	for _, v := range cli.ViewNames() {
		switch v {
		case "board", "findings", "friction", "worklist":
			continue // JSON by name — driven by their own oracles, not the markdown path
		}
		if out, err := tracked(bin, "merge", "show", "--view", v, "--run", runDir); err != nil {
			res.err = "show --view " + v + " (projection) failed:\n" + truncate(string(out))
			return res
		}
	}
	// And the SCOPED form, over a gap the run actually minted: `--view changes --id <gap>`
	// takes a different path (it resolves the gap and renders the comparison), so the unscoped
	// render above proves nothing about it. A gap the board does not know must be REFUSED, not
	// rendered empty — the read-side twin of requireGap.
	if ids := mintedGapIDs(runDir); len(ids) > 0 {
		if out, err := tracked(bin, "merge", "show", "--view", "changes", "--id", ids[0], "--run", runDir); err != nil {
			res.err = "show --view changes --id " + ids[0] + " failed:\n" + truncate(string(out))
			return res
		}
		if out, err := tracked(bin, "merge", "show", "--view", "changes", "--id", "R9-99", "--run", runDir); err == nil {
			res.err = "show --view changes --id R9-99 SUCCEEDED on a gap nobody minted — a view that invents a comparison:\n" + truncate(string(out))
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
		for _, e := range board.Events {
			if e.Type == "proof" && e.Payload.Str("answers") != "" {
				proved[e.Payload.Str("answers")] = true
			}
			switch e.Type {
			case "blue_edit":
				if id := e.Payload.Str("answers"); id != "" {
					answered[id] = true
				}
			case "dispute":
				if id := e.Payload.Str("gap_id"); id != "" {
					disputed[id] = true
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
			switch g.Mint.Str("required_fix") {
			case dirIgnore:
				if answered[id] {
					res.err = "scenario IGNORE: " + id + " was answered by a blue_edit, but the scenario said blue does nothing — either a writer ignored the directive or --answers is attaching provenance nobody claimed"
					return res
				}
			case dirDisputeWon, dirDisputeLost:
				if !disputed[id] {
					res.err = "scenario DISPUTE: " + id + " has no dispute event — the contest the scenario specified is absent from the record"
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
			if satisfied(g.Mint.Str("required_fix")) && r.evaluated[id] && g.Open {
				res.err = "scenario " + g.Mint.Str("required_fix") + ": " + id + " is still OPEN after red sat on the repaired board — work was done and the board never recorded it as finished"
				return res
			}
		}
	}

	// A RULING HAS A CONSEQUENCE, AND A CONTEST IS VISIBLE AS ONE.
	//
	// Red's ruling and the avenue's fate were both on the record and joined NOWHERE, so blue
	// pursuing a line red called out-of-scope looked exactly like pursuing one red endorsed.
	// The design says a ruling is an ARGUMENT blue may contest — but `blue dispute --id A1` is
	// refused outright (disputes are gap-shaped and take grade dimensions), so the only channel
	// is the move itself, and until now it said nothing about what it was answering.
	if board, err := record.BoardState(runDir); err == nil {
		contests := map[string]string{}
		for _, e := range board.Events {
			if e.Type == "avenue" && e.Payload.Str("contests_ruling") != "" {
				contests[e.Payload.Str("avenue_id")] = e.Payload.Str("contests_ruling")
			}
		}
		for _, a := range record.Avenues(board) {
			ruling := record.AvenueRuling(runDir, a.ID)
			if ruling == "" || a.Status == "proposed" {
				continue // never ruled, or blue has not answered yet
			}
			switch {
			case strings.HasPrefix(a.Line, avContest):
				if a.Status != "pursued" {
					res.err = "avenue " + a.ID + " was proposed as CONTESTED but ended " + a.Status + " — blue was to pursue it against the ruling"
					return res
				}
				if contests[a.ID] == "" {
					res.err = "avenue " + a.ID + " was pursued AGAINST a " + ruling + " ruling and the record does not say so — the disagreement is invisible, which is the state this join exists to end"
					return res
				}
			case ruling == "endorsed":
				if a.Status != "pursued" {
					res.err = "avenue " + a.ID + " was ENDORSED and ended " + a.Status + " — a ruling with no consequence"
					return res
				}
			default:
				if a.Status == "pursued" && contests[a.ID] == "" {
					res.err = "avenue " + a.ID + " was ruled " + ruling + " and pursued anyway with nothing recording the contest"
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
			if e.Type == "retire" && e.Payload.Str("removal_basis") != record.RemovalVerified {
				res.err = "a retire recorded removal_basis=" + e.Payload.Str("removal_basis") +
					" — nothing on the record shows that claim was ever in the report, and an unevidenced retirement cancels real claim loss in the additive-integrity detector"
				return res
			}
		}
	}

	// EVERY VERDICT IS DERIVABLE TODAY, and this is the tripwire on that fact.
	//
	// The tool refuses an --as that contradicts the record (#308), so a recorded verdict is
	// either DERIVED or came from the one case the record cannot decide: a judged deadlock.
	// debate.js cannot currently produce one — `deadlock` is hardcoded false whenever red
	// raised anything new (#289) — so an `asserted` basis anywhere in a fuzz run means the
	// derivation stopped working, not that a deadlock happened.
	//
	// When #289 gives the bench a real stopping call, this WILL fail, loudly and correctly, and
	// the right response is to update it rather than to widen it: an asserted verdict is
	// exactly the thing that should be rare enough to notice.
	if board, err := record.BoardState(runDir); err == nil {
		for _, e := range board.Events {
			if e.Type != "outcome" {
				continue
			}
			switch e.Payload.Str("verdict_basis") {
			case record.VerdictDerived:
			case "":
				res.err = "the outcome event carries no verdict_basis — the field that says whether the verdict was computed or claimed has gone missing"
				return res
			default:
				res.err = "the run recorded an ASSERTED verdict (" + e.Payload.Str("verdict") + "), which today can only mean the derivation failed: debate.js cannot produce a judged deadlock while it is hardcoded false (#289)"
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
			if e.Type == "outcome" {
				recorded = e.Payload.Str("verdict")
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
	if m := proseRenders(board, runDir); m != "" {
		res.err = m
	}
	if res.err == "" {
		if m := basisRenders(board, runDir); m != "" {
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
		if e.Type == "blue_edit" && e.Payload.Str("answers") != "" {
			res.editAnswers++
		}
		if e.Type == "mint" && e.Payload.Str("fix_basis") == "verified" {
			res.verifiedBasis++
		}
		if e.Type == "blue_edit" && e.Payload.Bool("applied_verbatim") {
			res.verbatimApplied++
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
	"closing", "position", "dispute", "dispute-respond", "opinion", "regrade", "mint", "close",
	"confidence", "cite", "verify", "finding", "avenue", "friction", "revision", "retire",
	"manifest-row", "petition", "petition-rule", "verdict", "spot-check", "certify", "halt",
	// Added 2026-08-04 by a census of every type record.Append can write: these three were
	// APPENDABLE BUT UNGATED, so a regression that stopped emitting any of them would have
	// left the sweep green. `anchor` is the finding-marker's own record (the immortal-marker
	// detector's EXPECTED set is exactly these), `class-new` is the growing gap registry's
	// write, `outcome` is the bench's.
	"blue_edit", "anchor", "class-new", "outcome", "avenue-rule", "proof",
}

// coverExempt names verbs tallied but NOT required in the random-sweep coverage gate.
var coverExempt = map[string]bool{"halt": true} // terminal — covered by TestFuzzHaltPath

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
	r := &runner{bin: bin, runDir: runDir, rng: rand.New(rand.NewSource(1)), registered: map[string]bool{}, forceHalt: true}

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
		if e.Type == "halt" {
			halts++
		}
	}
	if halts == 0 {
		t.Fatal("no halt event on the record — the halt verb never ran")
	}
	if out, err := tracked(bin, "verify", "--run", runDir); err != nil {
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
		if want[e.Type] {
			m[e.Type]++
		}
	}
	return m
}

// THIS MAP WAS AN ALLOWLIST, AND THAT IS WHY IT STOPPED WORKING.
//
// It held six event types. A type absent from it was not checked, and absence was silent — so
// every event type added after it was written defaulted to "the report need not render this",
// with nothing anywhere saying so. A 2026-08-08 trace of all 30 types found four whole exchanges
// reaching the record and never the reader (petition filings, avenue rulings, retirements,
// observations) plus two verbs read by nothing at all, under a green sweep.
//
// It is now EXHAUSTIVE over the types a run actually produces: every event type seen on the
// record must appear here or in reportExemptions with a stated reason, and an unclassified type
// FAILS. A new verb cannot be added without deciding, in writing, what its reader gets.
var dialecticProseKey = map[string]string{
	// The debate proper.
	"position": "text", "closing": "text", "opinion": "rationale",
	"dispute": "evidence", "dispute-respond": "rationale",
	// Petitions: the ASK and the ANSWER. `petition-rule` alone was the half-dialog — the
	// reader got the bench's ruling with no question attached.
	"petition": "basis", "petition-rule": "opinion",
	// The board.
	"mint": "problem", "finding": "text", "regrade": "basis",
	// Directions: the line, red's ruling on it, and blue's reason for its fate.
	"avenue": "line", "avenue-rule": "reason",
	// The lens's below-the-bar work and the fate the merge gave it.
	// Substance leaving the report, on the record, with its reason.
	"retire": "claim",
	// confidence's "prose" is its claim label — it must render in the report's confidence
	// self-assessment section, or blue's calibration silently vanishes (the dead-letter it was).
	"confidence": "label",
	// Run-level voices.
	"friction": "text", "revision": "text", "halt": "opinion", "certify": "statement",
	// Blue's self-audit receipt, one per repaired gap. `row` is what blue checked and what
	// checking it showed — the receipt reached no reader for a year, because the coverage metric
	// counted the ENVELOPE array and the verb was named in no prompt at all (#318).
	"manifest-row": "row",
	// Red re-reading its own closure archive. `notes` is what the sample FOUND — the whole
	// point of sampling — and it reached no reader at all until the floor was enforced (#317).
	"spot-check": "notes",
}

// reportExemptions are event types whose prose is deliberately NOT expected in the report, each
// with its reason. Stated rather than omitted: an absence with no reason is indistinguishable
// from an oversight, which is precisely how this gate decayed.
var reportExemptions = map[string]string{
	"register":  "a seat announcing itself to the run — attribution machinery, and the attribution reaches the reader on every act that seat records, never as an entry of its own",
	"anchor":    "an estoppel key spliced INTO blue/report.md — it is machinery for the edit path, and the text it anchors is the lifted content itself",
	"blue_edit": "mutates blue/report.md, which assembly lifts verbatim; the edit's effect IS in the report, and rendering the old/new spans again would duplicate the document",
	"class-new": "registers a gap class; the class reaches the reader on every gap that carries it, not as an entry of its own",
	"verify":    "red recording that it CHECKED a claim against its source. It reaches the reader through the citation-ledger view rather than the report: the report carries BLUE's citations (woven into the bibliography), and red verifying them is audit provenance rather than a claim the reader acts on. The trust grade IS surfaced — as the board's citations count, which #341 stopped inflating with blue's authored cites",
	"cite":      "resolved rather than rendered — the anchor becomes a visible [^N] and the source becomes a ## Bibliography line (weaveCitations)",
	"proof":     "resolved rather than rendered — weaveProofs splices the computation at its anchor",
	"close":     "the closure's prose is red's acceptance argument and reaches the reader only as an index row today; rendering it in full is tracked, not silently accepted",
	"outcome":   "composed into the verdict stamp by verdictStamp, from the payload's verdict/deadlocked/exhausted fields rather than a prose field",
	"verdict":   "red's per-round PASS/FAIL, consumed by DeriveVerdict into the terminal outcome; the round-by-round spine is not yet a transcript section",
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
	{"mint", "existence", "verified", "red checked this defect at the leaf"},
	{"mint", "existence", "suspected", "the grades below are conditional on the defect being real"},
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

func basisRenders(board *record.Board, runDir string) string {
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
			if e.Type == bf.evType && e.Payload.Str(bf.key) == bf.value {
				if bf.evType == "mint" && !gapIsOpen(board, e.Payload.Str("gap_id")) {
					continue
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

func proseRenders(board *record.Board, runDir string) string {
	report, err := os.ReadFile(filepath.Join(runDir, "report.md"))
	if err != nil {
		return "no report.md: " + err.Error()
	}
	rpt := string(report)
	var missing, unclassified []string
	seen := map[string]bool{}
	for _, e := range board.Events {
		key, ok := dialecticProseKey[e.Type]
		if !ok {
			// An event type in neither table is a DECISION NOBODY MADE. Registering it here
			// costs one line; leaving it unclassified is how four exchanges reached the record
			// and never the reader.
			if reportExemptions[e.Type] == "" && !seen[e.Type] {
				seen[e.Type] = true
				unclassified = append(unclassified, e.Type)
			}
			continue
		}
		prose := strings.TrimSpace(e.Payload.Str(key))
		if prose == "" || strings.Contains(rpt, prose) {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s/%s prose absent from report: %q", e.SeatID, e.Type, prose))
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

func binDir(bin string) string { return filepath.ToSlash(filepath.Dir(bin)) }

func truncate(s string) string {
	if len(s) > 600 {
		return s[:600] + "…"
	}
	return s
}

func TestFuzzDebate(t *testing.T) {
	bin := buildBinary(t)
	wrapped := debateWrapped(t)

	// CI runs `go test ./...` (no -short) on four jobs, so the DEFAULT is a modest smoke that
	// proves the harness and catches gross regressions in ~15s. The full 1000-run confidence
	// sweep is on demand: FUZZ_N=1000 go test ./internal/fuzz -run TestFuzzDebate -timeout 600s.
	n := 60
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

	for i := 0; i < n; i++ {
		seed := int64(i) + 1
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			o := runOne(wrapped, bin, seed)
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
			if o.err != "" {
				failures = append(failures, o)
			} else {
				os.RemoveAll(o.runDir) // keep only failing runs, for inspection
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	t.Logf("fuzzed %d debate runs · %d failed · verdicts=%v · rounds=%v\n  dialectic events emitted: %v\n  citation axis: %d anchors spliced · %d sources cached\n  provenance: %d of %d blue_edit ops carried --answers · %d of %d gaps earned fix_basis=verified · %d edits applied a proposal verbatim",
		completed, len(failures), verdicts, roundHist, dcov, citeAnchors, cacheFiles, editAnswers, dcov["blue_edit"], verifiedBasis, dcov["mint"], verbatimApplied)
	// FULL-SURFACE COVERAGE GATE. A green fuzz that never drove a verb is a false green (the lens
	// stub emitted neither cite nor finding for the whole life of PR-1, unexercised end to end).
	// Assert EVERY event-emitting seat verb fired at least once across the run set — so a
	// regression that silently stops emitting one (a dropped envelope branch, a broken dispatch)
	// fails here, not silently. `halt` is exempt (terminal; covered by TestFuzzHaltPath). Each
	// gated verb fires with per-run probability well above ~20%, so P(missed across the default 60
	// runs) is negligible; if a verb ever flakes here it is a real coverage regression, not noise.
	// The full surface needs enough runs for every verb (incl. the ~10%/run ones like regrade) to
	// fire; below ~40 runs a low-frequency verb could flake to zero, so the -short smoke keeps only
	// the cite/finding floor (the original false-green guard) and the default+ size asserts all.
	if completed >= 40 {
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
	if completed >= 40 {
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
	if completed < 40 {
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
	if completed >= 40 {
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
		t.Log(execReport())
	}
	if len(failures) > 0 {
		show := failures
		if len(show) > 8 {
			show = show[:8]
		}
		for _, f := range show {
			t.Errorf("seed %d FAILED (runDir %s):\n%s", f.seed, f.runDir, f.err)
		}
		t.Fatalf("%d/%d fuzz runs failed — see seeds above (reproduce with that seed)", len(failures), n)
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "feov-record.exe")
	out, err := exec.Command("go", "build", "-o", bin, "../../cmd/feov-record").CombinedOutput()
	if err != nil {
		t.Fatalf("build feov-record: %v\n%s", err, out)
	}
	return bin
}

// mintedGapIDs lists the gaps a run actually created, so a scoped-view oracle asks about
// something real rather than about a shape the fuzz happened not to produce this seed.
func mintedGapIDs(runDir string) []string {
	b, err := record.BoardState(runDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range b.Events {
		if e.Type == "mint" {
			out = append(out, e.Payload.Str("gap_id"))
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
		if e.Type != "blue_edit" {
			continue
		}
		old := e.Payload.Str("old")
		if old != "" && !strings.Contains(string(cur), old) {
			return old
		}
	}
	return ""
}

// avenueFates are the scenarios an avenue is proposed under. The line itself carries it, the
// way a gap's required_fix does — red rules from it, and blue's next move answers the ruling.
const (
	avEndorse = "FUZZ-AVENUE-ENDORSE: a line worth the run's time"
	avScope   = "FUZZ-AVENUE-OUT-OF-SCOPE: a real question, but not this run's"
	avThin    = "FUZZ-AVENUE-TOO-THIN: in scope, but the hypothesis does not carry its budget"
	// avContest is the case gb's design named and the tool could not express: red rules the
	// line out, and blue PURSUES IT ANYWAY with an argument. A ruling is an argument, never a
	// command — `blue dispute --id A1` is refused outright, because disputes are gap-shaped —
	// so the move itself is the contest, and it is now recorded as one.
	avContest = "FUZZ-AVENUE-CONTESTED: red rules it out and blue pursues it anyway, with reasons"
)

var avenueFates = []string{avEndorse, avEndorse, avScope, avThin, avContest}

// rulingFor maps a proposed line to the ruling red should give it.
func rulingFor(line string) string {
	switch {
	case strings.HasPrefix(line, avEndorse):
		return "endorsed"
	case strings.HasPrefix(line, avScope):
		return "out-of-scope"
	case strings.HasPrefix(line, avThin):
		return "too-thin"
	case strings.HasPrefix(line, avContest):
		return "out-of-scope"
	}
	return ""
}

// ruleOpenAvenues has red rule every avenue that has no ruling yet, from the line it carries.
func (r *runner) ruleOpenAvenues(seatID string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return
	}
	for _, a := range record.Avenues(b) {
		if rulingFor(a.Line) == "" || record.AvenueRuling(r.runDir, a.ID) != "" {
			continue
		}
		r.do("merge", "avenue-rule", seatID).set("--id", a.ID).
			set("--ruling", rulingFor(a.Line)).
			set("--reason", "fuzz: ruling as the line was proposed").run()
	}
}

// answerAvenueRulings is blue's move after red has ruled: comply, or CONTEST by pursuing
// anyway with an argument. Which one is decided by the line, not by a coin.
func (r *runner) answerAvenueRulings(seatID string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return
	}
	for _, a := range record.Avenues(b) {
		ruling := record.AvenueRuling(r.runDir, a.ID)
		if ruling == "" || a.Status != "proposed" {
			continue
		}
		switch {
		case strings.HasPrefix(a.Line, avContest):
			r.do("blue", "avenue", seatID).set("--id", a.ID).set("--status", "pursued").
				set("--reason", "fuzz: the scope call is wrong, this bears on the core claim").run()
		case ruling == "endorsed":
			r.do("blue", "avenue", seatID).set("--id", a.ID).set("--status", "pursued").
				set("--reason", "fuzz: endorsed, taking it up").run()
		default:
			r.do("blue", "avenue", seatID).set("--id", a.ID).set("--status", "declined").
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
		if e.Type == "proof" && e.Payload.Str("answers") != "" && e.Payload.Str("sha256") != "" {
			proofFor[e.Payload.Str("answers")] = e.Payload.Str("sha256")
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
		out, err := r.exec("--json", "lens", "reproduce", "--seat-id", seatID, "--id", sha)
		if err != nil {
			continue
		}
		var env struct {
			Result struct {
				Matches bool `json:"matches"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) == nil && env.Result.Matches {
			r.reproduced[id] = true
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
	return g.Mint.Str("required_fix")
}

// blueRespondTo carries out the scenario each open gap was minted with — one decision per gap,
// taken from the gap rather than from a coin, which is what makes the terminal state assertable.
func (r *runner) blueRespondTo(seatID string, open []string) {
	for _, id := range open {
		switch r.scenarioOf(id) {
		case dirApply:
			// The only branch that sets applied_verbatim, and so the only one that estops red.
			if gid, fo, fn := r.proposalFor(id); gid != "" {
				if cur, err := os.ReadFile(filepath.Join(r.runDir, "blue", "report.md")); err == nil && strings.Contains(string(cur), fo) {
					r.do("blue", "edit", seatID).set("--old", fo).set("--new", fn).
						set("--answers", id).set("--reason", "fuzz: applying red's proposed text verbatim").run()
					continue
				}
			}
			// No concrete pair to apply — degrade to a counter-edit rather than fake one.
			fallthrough
		case dirCounter:
			r.counterEdit(seatID, id)
		case dirDisputeWon, dirDisputeLost:
			dim := pick(r.rng, disputeDims)
			proposed := r.g()
			if _, err := r.exec("blue", "dispute", "--seat-id", seatID, "--id", id,
				"--dimension", dim, "--proposed", proposed, "--reason", "fuzz: contesting the grade on "+id); err == nil {
				ref := map[string]any{"gap_id": id, "dimension": dim, "proposed": proposed}
				r.raised = append(r.raised, ref)
				r.disputedThisRound = append(r.disputedThisRound, ref)
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
				r.do("blue", "prove", seatID).
					set("--location", "§ fuzz").
					set("--script", name).
					set("--answers", id).
					// The METHOD is cited and the instance computed — that pairing is the whole
					// claim of the proof axis, and --cites was never passed.
					on(50, "--cites", "c-fuzz").
					set("--reason", "fuzz: computing rather than arguing").
					run()
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
	r.do("blue", "edit", seatID).set("--old", oldSpan).set("--new", newSpan).
		set("--answers", gapID).set("--reason", "fuzz: counter-edit, not red's text").run()
}

// proposalFor returns the concrete pair for ONE gap, when it carries one.
func (r *runner) proposalFor(gapID string) (string, string, string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return "", "", ""
	}
	g := b.Gaps[gapID]
	if g == nil || g.Mint == nil || g.Mint.Str("fix_basis") != "verified" {
		return "", "", ""
	}
	return gapID, g.Mint.Str("fix_old"), g.Mint.Str("fix_new")
}

func (r *runner) someProposal() (string, string, string) {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return "", "", ""
	}
	for _, e := range b.Events {
		if e.Type == "mint" && e.Payload.Str("fix_basis") == "verified" {
			return e.Payload.Str("gap_id"), e.Payload.Str("fix_old"), e.Payload.Str("fix_new")
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
// the report's refusal count, which is where the dead `lens avenue` drive and the misrouted
// `frontier` registration were both found.
func (r *runner) readOnly(role, verb, seatID string, extra ...string) {
	args := append([]string{role, verb, "--seat-id", seatID}, extra...)
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
	// Every role's `show` with NO --view, which resolves that role's DEFAULT view. Only the
	// merge's was ever driven, so a regression in any other role's default was invisible.
	{"lens", "show"}, {"merge", "show"}, {"blue", "show"}, {"bench", "show"},
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
func dashboardArgv(runDir string) []string { return []string{"dashboard", runDir, runDir} }

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
		if len(argv) == 2 && argv[1] == "show" {
			args = append(args, "--seat-id", seatFor(argv[0]))
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
