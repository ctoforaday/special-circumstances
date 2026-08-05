package fuzz

// Fuzz the ACTUAL debate.js orchestrator against the REAL feov-record binary. goja runs the
// unmodified script (harness faked); each agent() call is a seat that makes coherent, random,
// VALID tool calls into a real run directory and returns a randomised envelope to drive
// debate.js's own branches (verdict, gap counts, rounds, petitions). A failing run is a real
// finding (in debate.js, the tool, or verify), reproducible from its seed.
//
// COVERAGE CONTRACT. envelopeFor drives every eligible seat to exercise its whole verb surface,
// not a happy path: lens (cite/finding/observe/avenue/friction), merge (position/closing/
// mint/close incl. closed_with_regression/dispose across its full --as domain/regrade any axis/
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
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

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
	// #62 Stage 2: disputes blue RAISED (event emitted + envelope ref), awaiting red's answer
	// next round — mirrors debate.js's pendingDisputes so the fuzz drives the docket machinery
	// through the ENVELOPE, not just the events. Each: {gap_id, dimension, proposed}.
	raised []map[string]any
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
		resp := "accepted"
		if r.coin(50) {
			resp = "rejected"
		}
		if _, err := r.exec("merge", "dispute-respond", "--seat-id", seatID, "--id", id, "--as", resp, "--reason", "respond-rationale-for-"+id+"-by-"+seatID); err != nil {
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
		if r.forceHalt {
			ruling = "halt"
			_, _ = r.exec("bench", "halt", "--seat-id", seatID, "--reason", "fuzz judicial halt — safety boundary")
		} else {
			if r.coin(50) {
				ruling = "denied"
			}
			_, _ = r.exec("bench", "petition-rule", "--seat-id", seatID, "--petitioner", who, "--petition-class", class, "--as", ruling, "--reason", opinion)
		}
		rulings = append(rulings, map[string]any{"petitioner": who, "class": class, "ruling": ruling, "opinion": opinion})
	}
	r.petitioned = nil
	if rulings == nil {
		rulings = arr()
	}
	return map[string]any{"rulings": rulings, "friction": arr()}
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
	args := []string{"--json", "merge", "mint", "--seat-id", seatID, "--problem", "fuzz problem", "--check", "acc", "--likelihood", r.g(), "--impact", r.g()}
	if !r.classMade {
		r.classMade = true
		args = append(args, "--class-new", "fuzzcls", "--definition", "d", "--neighbor", "verification-gap", "--distinguisher", "q")
	} else {
		args = append(args, "--class", "fuzzcls")
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
	_, _ = r.exec("merge", "close", "--seat-id", seatID, "--id", id, "--as", "closed", "--reason", "fuzz close",
		"--anchor-seat", seatID, "--anchor-tool", "fuzz", "--anchor-target", "rec")
}

var avenueStatus = []string{"pursued", "abandoned", "declined"}
var obsKind = []string{"note", "checked-held"}
var disposeAs = []string{"declined", "banked"}

// disposeInto are the dispositions that FOLD an observation into a gap — they take --into <open
// gap>. Kept separate from disposeAs so the fuzz only picks them when an open gap exists.
var disposeInto = []string{"minted-as", "folded-into"}

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
	avenue := func(role string) {
		// A DECLINED OR ABANDONED avenue requires --reason (record.go: an unexplained
		// non-pursuit is the decoration this verb exists to refuse). Without it two of the
		// three statuses were rejected on every call, so only `pursued` ever reached the
		// record while the verb gate read as covered. Found by the execution tally
		// (blue avenue: 48 of 72 calls refused).
		st := pick(r.rng, avenueStatus)
		c := r.do(role, "avenue", seatID).set("--status", st).set("--line", "fuzz avenue "+seatID).on(50, "--method", "fuzz-method")
		if st != "pursued" {
			c = c.set("--reason", "fuzz: why this line was not pursued")
		}
		c.run()
	}
	switch role {
	case "lens":
		r.maybe(45, func() {
			r.do("lens", "observe", seatID).set("--label", fmt.Sprintf("O%d", r.rng.Intn(1_000_000))).set("--kind", pick(r.rng, obsKind)).set("--reason", "fuzz observation").run()
		})
		// NO avenue DRIVE HERE: the lens role has no avenue verb (register/finding/observe/
		// cite/friction/show). This called it 183 times per sweep, every one refused, while
		// the verb gate stayed green on blue's avenue events — a dead drive that read as
		// coverage. Found by the execution tally (lens avenue: 183 of 183 refused).
	case "merge":
		// W1.8 archive spot-check — the merge's duty. --none is always valid (an empty archive at
		// round start); it exercises the verb and its --none/--reason branch.
		r.maybe(45, func() {
			r.do("merge", "spot-check", seatID).bare("--none").set("--reason", "fuzz: nothing to sample").run()
		})
		// near-match is the screen red runs BEFORE minting, to catch a reopen. Read-only, so it
		// left no event and no gate saw it — while the merge prompt calls it every round.
		r.maybe(40, func() {
			r.readOnly("merge", "near-match", seatID, "--candidate", "fuzz candidate problem text for screening")
		})
	case "blue":
		r.maybe(45, func() { avenue("blue") })
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
			// VERBATIM APPLICATION (#267 stage 4): when a gap carries a concrete proposal,
			// sometimes apply exactly it — the only state that sets applied_verbatim and so
			// the only one that estops red. The rest of the time blue counter-edits, which is
			// the decline path DeclineStats counts. Both must run.
			if id, fo, fn := r.someProposal(); id != "" && r.coin(50) && strings.Contains(string(cur), fo) {
				r.do("blue", "edit", seatID).
					set("--old", fo).set("--new", fn).
					set("--answers", id).
					set("--reason", "fuzz: applying red's proposed text verbatim").
					run()
				return
			}
			c := r.do("blue", "edit", seatID).
				set("--old", oldSpan).
				set("--new", newSpan).
				set("--reason", "fuzz edit: sharper phrasing").
				on(40, "--key", fmt.Sprintf("E%d", 1+r.rng.Intn(2)))
			if len(open) > 0 {
				c = c.set("--answers", pick(r.rng, open))
			}
			c.run()
		})
		// READ-ONLY, PROMPT-CALLED, AND PREVIOUSLY UNFUZZED. claim-index records nothing, so the
		// event-type coverage gate cannot see it — yet blue's response prompt tells blue to call
		// it when propagating a correction to every site of a claim. A crash here costs a real
		// round; no gate would have noticed.
		r.maybe(35, func() { r.readOnly("blue", "claim-index", seatID) })
		r.maybe(40, func() { r.do("blue", "revision", seatID).set("--reason", "fuzz revision").run() })
		r.maybe(30, func() {
			r.do("blue", "retire", seatID).set("--claim", "fuzz claim "+seatID).set("--reason", "fuzz retire").on(50, "--superseded-by", "fuzz replacement claim "+seatID).run()
		})
		if len(open) > 0 {
			r.maybe(40, func() {
				r.do("blue", "manifest-row", seatID).set("--id", pick(r.rng, open)).set("--row", "fuzz manifest row").run()
			})
		}
	case "bench":
		// run-end certification statement — the bench's, additive (a recorded opinion).
		r.maybe(45, func() {
			r.do("bench", "certify", seatID).set("--reason", "fuzz certification statement from "+seatID).run()
		})
	}
}

// disposeObservations gives every observation a FATE — the merge's duty — so a run that
// randomly grew observations still ends clean and exercises observe -> dispose end to end.
func (r *runner) disposeObservations(seatID string) {
	out, err := r.exec("merge", "show", "--view", "board")
	if err != nil {
		return
	}
	var b struct {
		Observations []struct {
			Label    string `json:"label"`
			Disposed bool   `json:"disposed"`
		} `json:"observations"`
		Open []struct {
			ID string `json:"id"`
		} `json:"open"`
	}
	if json.Unmarshal([]byte(out), &b) != nil {
		return
	}
	for _, o := range b.Observations {
		if o.Disposed || o.Label == "" {
			continue
		}
		// The FOLD dispositions (minted-as|folded-into) require a target gap via --into; pick one
		// when an open gap exists, else fall back to the free dispositions (declined|banked). This
		// exercises the full --as domain plus the --into flag.
		if len(b.Open) > 0 && r.coin(50) {
			into := b.Open[r.rng.Intn(len(b.Open))].ID
			_, _ = r.exec("merge", "dispose", "--seat-id", seatID, "--observation", o.Label, "--as", pick(r.rng, disposeInto), "--into", into, "--reason", "fuzz dispose")
		} else {
			_, _ = r.exec("merge", "dispose", "--seat-id", seatID, "--observation", o.Label, "--as", pick(r.rng, disposeAs), "--reason", "fuzz dispose")
		}
	}
}

// arr / obj are goja-friendly envelope builders.
func arr(v ...any) []any { return append([]any{}, v...) }

// envelopeFor performs the seat's tool acts and returns the envelope debate.js expects. The
// randomisation here is what drives debate.js's control flow: red's verdict and gap count
// decide whether the loop continues, deadlocks, or terminates.
func (r *runner) envelopeFor(seatID string) map[string]any {
	switch {
	case strings.HasPrefix(seatID, "blue-synthesize"):
		r.register("blue", seatID)
		r.emitConfidence(seatID) // round-0 calibration
		r.extras("blue", seatID, nil)
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "petitions": r.maybePetition("blue", seatID), "friction": arr()}

	case strings.HasPrefix(seatID, "red-merge"):
		r.register("merge", seatID)
		// Answer blue's pending disputes FIRST (debate.js processes dispute_responses before the
		// PASS/FAIL branch) — emits the events + returns the routing refs, whatever the verdict.
		responses := r.answerDisputes(seatID)
		r.extras("merge", seatID, r.openGaps())
		// PASS ~40%: run the round's dialectic, dispose any loose observations, then close every
		// open gap so the #67 gate (and verify) holds.
		if r.coin(40) {
			open := r.openGaps()
			r.dialectic("merge", seatID, open)
			r.disposeObservations(seatID)
			// Close every open gap so the #67 gate (and verify) holds. The FIRST pass may
			// regression-close (minting successors) and dispose may have minted-as new gaps; re-read
			// and plain-close until the board is empty. Bounded: only the first pass regression-closes.
			for first := true; ; first = false {
				rem := r.openGaps()
				if len(rem) == 0 {
					break
				}
				for _, id := range rem {
					r.closeGap(seatID, id, first)
				}
			}
			// The merge's TERMINAL act on a PASS: checkpoints records/ (the event log) to the
			// recovery mirror. A PASS ends the debate, so this round is terminal.
			_, _ = r.exec("merge", "verdict", "--seat-id", seatID, "--as", "PASS")
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "corroboration": arr(), "petitions": r.maybePetition("merge", seatID), "friction": arr()}
		}
		// FAIL: mint 1-3 fresh gaps (a FAIL with empty gaps is a degenerate merge debate.js
		// rejects on purpose). Report the FULL open set as the docket.
		for range r.rng.Intn(3) + 1 {
			r.mint(seatID)
		}
		open := r.openGaps()
		if len(open) > 0 {
			r.dialectic("merge", seatID, open)
		}
		var gaps []any
		for _, id := range r.openGaps() {
			gaps = append(gaps, map[string]any{"id": id, "supersedes": arr()})
		}
		if len(gaps) == 0 {
			// The mints did not take (a tool refusal we could not satisfy) — degrade to PASS
			// rather than fabricate a docket. A FAIL with an empty gaps array is exactly the
			// degenerate merge debate.js rejects, and we must not hand it one.
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "corroboration": arr(), "petitions": r.maybePetition("merge", seatID), "friction": arr()}
		}
		return map[string]any{"verdict": "FAIL", "gaps": gaps, "closures": arr(), "dispute_responses": responses, "corroboration": arr(), "petitions": r.maybePetition("merge", seatID), "friction": arr()}

	case strings.HasPrefix(seatID, "blue-respond"):
		r.register("blue", seatID)
		open := r.openGaps()
		r.dialectic("blue", seatID, open) // blue's position, closings, confidence
		r.extras("blue", seatID, open)
		disputes := r.raiseDisputes(seatID, open) // emit dispute events + envelope routing refs
		// debate.js rejects an EMPTY manifest on a round with open gaps — a repair must show its
		// receipt. One row per open gap it is repairing this round.
		var manifest []any
		for _, id := range open {
			manifest = append(manifest, map[string]any{"gap_id": id, "row": "fuzz: checked"})
		}
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "manifest": manifest, "grade_disputes": disputes, "petitions": r.maybePetition("blue", seatID), "friction": arr()}

	case strings.HasPrefix(seatID, "judge-petition"):
		r.register("bench", seatID)
		return r.rulePetitions(seatID) // rule every pending petition (petition-rule events + envelope rulings)

	case strings.HasPrefix(seatID, "judge"): // adjudication + terminal
		r.register("bench", seatID)
		r.extras("bench", seatID, nil)
		var res []any
		for _, id := range r.openGaps() {
			disp := "carried"
			if r.rng.Intn(2) == 0 {
				disp = "closed"
				_, _ = r.exec("bench", "opinion", "--seat-id", seatID, "--id", id, "--as", "closed",
					"--principle", "correctness", "--tension", "cost", "--review-flag", "false", "--reason", "opinion-rationale-for-"+id)
			}
			res = append(res, map[string]any{"gap_id": id, "resolution": disp, "rationale": "fuzz"})
		}
		return map[string]any{"resolutions": res, "deadlock": false, "friction": arr()}

	case strings.HasPrefix(seatID, "assemble"):
		r.register("bench", seatID)
		verd := []string{"VERIFIED", "CEILING", "UNVERIFIED"}[r.rng.Intn(3)]
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
			if strings.HasPrefix(seatID, "red-lens") {
				_, _ = r.exec("lens", "cite", "--seat-id", seatID, "--claim", "fuzz claim "+seatID,
					"--reference", "https://fuzz.invalid/"+seatID, "--confidence", confGrades[r.rng.Intn(len(confGrades))], "--access-date", "2026-07-24")
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
		env := r.envelopeFor(seatID)
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
	for _, v := range []string{"ledger", "archive", "debate", "changelog", "changes", "citation-ledger", "lines-of-inquiry"} {
		if out, err := tracked(bin, "merge", "show", "--view", v, "--run", runDir); err != nil {
			res.err = "show --view " + v + " (markdown projection) failed:\n" + truncate(string(out))
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
	"confidence", "cite", "finding", "observe", "avenue", "friction", "revision", "retire",
	"manifest-row", "dispose", "petition", "petition-rule", "verdict", "spot-check", "certify", "halt",
	// Added 2026-08-04 by a census of every type record.Append can write: these three were
	// APPENDABLE BUT UNGATED, so a regression that stopped emitting any of them would have
	// left the sweep green. `anchor` is the finding-marker's own record (the immortal-marker
	// detector's EXPECTED set is exactly these), `class-new` is the growing gap registry's
	// write, `outcome` is the bench's.
	"blue_edit", "anchor", "class-new", "outcome",
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

var dialecticProseKey = map[string]string{
	"closing": "text", "opinion": "rationale", "dispute": "evidence",
	"dispute-respond": "rationale", "petition-rule": "opinion",
	// confidence's "prose" is its claim label — it must render in the report's confidence
	// self-assessment section, or blue's calibration silently vanishes (the dead-letter it was).
	"confidence": "label",
}

func proseRenders(board *record.Board, runDir string) string {
	report, err := os.ReadFile(filepath.Join(runDir, "report.md"))
	if err != nil {
		return "no report.md: " + err.Error()
	}
	rpt := string(report)
	var missing []string
	for _, e := range board.Events {
		key, ok := dialecticProseKey[e.Type]
		if !ok {
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

// ---- read-only coverage: the class the event gate is BLIND to ----

// readOnlyCalls tallies every invocation of a verb that RECORDS NOTHING, keyed "<role> <verb>".
//
// WHY A SECOND TALLY EXISTS. The coverage gate below counts EVENT TYPES, so it can only see a
// verb that writes to the record. Every read-only surface — the four roles' `show`, blue's
// claim-index, the merge's near-match screen, and the operator renders (graph, count-claims,
// scorecard, dashboard) — was structurally invisible to it: a regression that made any of them
// crash on a real run shape would have shipped behind a green fuzz. Measured 2026-08-04: of 44
// seat verbs, the ones with zero fuzz coverage were ALL of this kind, plus 7 of 9 root commands.
//
// Guarded by a mutex because seats run concurrently within a run.
var (
	readOnlyMu    sync.Mutex
	readOnlyCalls = map[string]int{}
)

func noteReadOnly(key string) {
	readOnlyMu.Lock()
	readOnlyCalls[key]++
	readOnlyMu.Unlock()
}

// readOnly runs a record-nothing seat verb and TALLIES it. A non-zero exit is a finding: these
// verbs are called mid-round by real prompts, so a crash costs a paid round.
func (r *runner) readOnly(role, verb, seatID string, extra ...string) {
	args := append([]string{role, verb, "--seat-id", seatID}, extra...)
	if _, err := r.exec(args...); err == nil {
		noteReadOnly(role + " " + verb)
	}
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
	noteReadOnly("dashboard")
	for _, argv := range readOnlySurfaces {
		args := append(append([]string{}, argv...), "--run", runDir)
		if len(argv) == 2 && argv[1] == "show" {
			args = append(args, "--seat-id", seatFor(argv[0]))
		}
		if out, err := tracked(bin, args...); err != nil {
			return "read-only surface `" + strings.Join(argv, " ") + "` failed on a real run shape:\n" + truncate(string(out))
		}
		noteReadOnly(strings.Join(argv, " "))
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
