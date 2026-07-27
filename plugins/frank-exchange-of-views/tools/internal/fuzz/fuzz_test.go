package fuzz

// Fuzz the ACTUAL debate.js orchestrator against the REAL feov-record binary. goja runs the
// unmodified script (harness faked); each agent() call is a seat that makes coherent, random,
// VALID tool calls into a real run directory and returns a randomised envelope to drive
// debate.js's own branches (verdict, gap counts, rounds). The oracle per run is `verify`:
// whatever random path the debate takes, the record it leaves must satisfy every invariant —
// plus the run must terminate and assemble. A failing run is a real finding (in debate.js, the
// tool, or verify), reproducible from its seed.
//
// Run: go test ./internal/fuzz -run TestFuzzDebate -count=1   (respects -short by shrinking N)

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
}

func (r *runner) coin(pct int) bool { return r.rng.Intn(100) < pct }

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
		_, _ = r.exec("merge", "regrade", "--seat-id", seatID, "--id", id, "--reason", "regrade-basis-for-"+id, "--likelihood", r.g())
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
		if _, err := r.exec("blue", "dispute", "--seat-id", seatID, "--id", id, "--dimension", "impact", "--proposed", proposed, "--reason", "dispute-evidence-for-"+id+"-by-"+seatID); err != nil {
			continue
		}
		ref := map[string]any{"gap_id": id, "dimension": "impact", "proposed": proposed}
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

func (r *runner) exec(args ...string) (string, error) {
	cmd := exec.Command(r.bin, append(args, "--run", r.runDir)...)
	out, err := cmd.CombinedOutput()
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

func (r *runner) closeGap(seatID, id string) {
	_, _ = r.exec("merge", "close", "--seat-id", seatID, "--id", id, "--as", "closed", "--reason", "fuzz close",
		"--anchor-seat", seatID, "--anchor-tool", "fuzz", "--anchor-target", "rec")
}

var avenueStatus = []string{"pursued", "abandoned", "declined"}
var obsKind = []string{"note", "checked-held"}
var disposeAs = []string{"declined", "banked"}

func pick(rng *rand.Rand, xs []string) string { return xs[rng.Intn(len(xs))] }

// extras fires a RANDOM subset of a role's REMAINING verb surface, so no two fuzz paths
// look alike and every seat exercises far more than its happy path. Only verbs that keep
// the oracles intact (verify passes, the run terminates, the report renders) are here;
// terminal/verdict-shaping acts (halt, certify, outcome, verdict) stay in the structured
// cases above. Reference-taking verbs are gated on a referent existing.
func (r *runner) extras(role, seatID string, open []string) {
	if r.coin(50) {
		_, _ = r.exec(role, "friction", "--seat-id", seatID, "--reason", "fuzz friction from "+seatID)
	}
	switch role {
	case "lens":
		if r.coin(45) {
			_, _ = r.exec("lens", "observe", "--seat-id", seatID, "--label", fmt.Sprintf("O%d", r.rng.Intn(1_000_000)),
				"--kind", pick(r.rng, obsKind), "--reason", "fuzz observation")
		}
		if r.coin(45) {
			_, _ = r.exec("lens", "avenue", "--seat-id", seatID, "--status", pick(r.rng, avenueStatus), "--line", "fuzz avenue "+seatID)
		}
	case "blue":
		if r.coin(45) {
			_, _ = r.exec("blue", "avenue", "--seat-id", seatID, "--status", pick(r.rng, avenueStatus), "--line", "fuzz avenue "+seatID)
		}
		if r.coin(40) {
			_, _ = r.exec("blue", "revision", "--seat-id", seatID, "--reason", "fuzz revision")
		}
		if r.coin(30) {
			_, _ = r.exec("blue", "retire", "--seat-id", seatID, "--claim", "fuzz claim "+seatID, "--reason", "fuzz retire")
		}
		if len(open) > 0 && r.coin(40) {
			_, _ = r.exec("blue", "manifest-row", "--seat-id", seatID, "--id", open[r.rng.Intn(len(open))], "--row", "fuzz manifest row")
		}
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
	}
	if json.Unmarshal([]byte(out), &b) != nil {
		return
	}
	for _, o := range b.Observations {
		if !o.Disposed && o.Label != "" {
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
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "petitions": arr(), "friction": arr()}

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
			for _, id := range r.openGaps() { // re-read: a regression close may have added a successor
				r.closeGap(seatID, id)
			}
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "corroboration": arr(), "petitions": arr(), "friction": arr()}
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
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": responses, "corroboration": arr(), "petitions": arr(), "friction": arr()}
		}
		return map[string]any{"verdict": "FAIL", "gaps": gaps, "closures": arr(), "dispute_responses": responses, "corroboration": arr(), "petitions": arr(), "friction": arr()}

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
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "manifest": manifest, "grade_disputes": disputes, "petitions": arr(), "friction": arr()}

	case strings.HasPrefix(seatID, "judge-petition"):
		r.register("bench", seatID)
		return map[string]any{"rulings": arr(), "friction": arr()}

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
		_, _ = r.exec("bench", "outcome", "--seat-id", seatID, "--as", verd)
		_, _ = r.exec("bench", "assemble")
		open := len(r.openGaps())
		return map[string]any{"synopsis": "fuzz", "open_gaps": open, "friction": arr()}

	default: // frontier, blue lanes, red lenses — register, and lenses record onto the channel
		if seatID != "" {
			role := "lens"
			switch {
			case strings.HasPrefix(seatID, "blue"):
				role = "blue"
			case strings.HasPrefix(seatID, "frontier"):
				role = "lens"
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
}

func runOne(wrapped, bin string, seed int64) outcome {
	runDir, _ := os.MkdirTemp("", "fuzz-run-")
	r := &runner{bin: bin, runDir: runDir, rng: rand.New(rand.NewSource(seed)), registered: map[string]bool{}}

	res := outcome{seed: seed, runDir: runDir}
	func() {
		defer func() {
			if p := recover(); p != nil {
				res.err = fmt.Sprintf("panic: %v", p)
			}
		}()
		loop := eventloop.NewEventLoop()
		var rejected string
		loop.Run(func(vm *goja.Runtime) {
			vm.Set("args", map[string]any{
				"topic": "fuzz", "runDir": runDir, "binDir": binDir(bin),
				"lanes": 1, "laneFloorOverride": "fuzz", "maxRounds": 3,
				// #111: both tiers are now REQUIRED — nil would refuse dispatch. Both haiku so the
				// tier oracle expects every dispatched seat to carry exactly "haiku".
				"model": "haiku", "judgmentModel": "haiku",
			})
			if _, err := vm.RunString(preamble); err != nil {
				rejected = "preamble: " + err.Error()
				return
			}
			vm.Set("agent", func(call goja.FunctionCall) goja.Value {
				prompt := ""
				if len(call.Arguments) > 0 {
					prompt = call.Argument(0).String()
				}
				seatID := ""
				if m := seatRe.FindStringSubmatch(prompt); m != nil {
					seatID = m[1]
				}
				if seatID == "" && len(call.Arguments) > 1 { // fall back to label
					if o := call.Argument(1).ToObject(vm); o != nil {
						if v := o.Get("label"); v != nil {
							if fields := strings.Fields(v.String()); len(fields) > 0 {
								seatID = fields[0]
							}
						}
					}
				}
				// #111 tier capture: record the model this dispatch carried ("unset" if absent).
				if len(call.Arguments) > 1 {
					if o := call.Argument(1).ToObject(vm); o != nil {
						mdl := "unset"
						if v := o.Get("model"); v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
							mdl = v.String()
						}
						r.models = append(r.models, mdl)
					}
				}
				env := r.envelopeFor(seatID)
				p, resolve, _ := vm.NewPromise()
				resolve(vm.ToValue(env))
				return vm.ToValue(p)
			})
			if _, err := vm.RunString(wrapped); err != nil {
				rejected = "run: " + err.Error()
			}
		})
		if rejected != "" {
			res.err = rejected
			return
		}
		loop.Run(func(vm *goja.Runtime) {
			if v := vm.Get("__result"); v != nil {
				if pr, ok := v.Export().(*goja.Promise); ok {
					switch pr.State() {
					case goja.PromiseStateRejected:
						res.err = "debate rejected: " + truncate(pr.Result().String())
					case goja.PromiseStatePending:
						res.err = "debate never settled (hang)"
					case goja.PromiseStateFulfilled:
						if m, ok := pr.Result().Export().(map[string]any); ok {
							if s, ok := m["verdict"].(string); ok {
								res.verdict = s
							}
							if n, ok := m["rounds"].(int64); ok {
								res.rounds = int(n)
							}
						}
					}
				}
			}
		})
	}()
	if res.err != "" {
		return res
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
	if out, err := exec.Command(bin, "verify", "--run", runDir).CombinedOutput(); err != nil {
		res.err = "verify FAILED:\n" + truncate(string(out))
		return res
	}
	// Oracle 1b: the JSON views the operator side reads in json-mode must exit 0 and parse.
	// `board` is already exercised above; `findings`/`friction` are JSON by name; `debate --json`
	// is the structured debate the capture audits count sections from. A broken view is what
	// would silently blank a dashboard tile or make an audit read an empty transcript.
	for _, v := range []string{"findings", "friction"} {
		out, err := exec.Command(bin, "merge", "show", "--view", v, "--run", runDir).CombinedOutput()
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show --view " + v + " did not return valid JSON:\n" + truncate(string(out))
			return res
		}
	}
	{
		out, err := exec.Command(bin, "merge", "show", "--view", "debate", "--json", "--run", runDir).CombinedOutput()
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show --view debate --json did not return valid JSON:\n" + truncate(string(out))
			return res
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
	res.dialectic = tallyDialectic(board)
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

// tallyDialectic counts the events that prove the fuzz exercised the paths it claims to.
func tallyDialectic(board *record.Board) map[string]int {
	want := map[string]bool{"closing": true, "position": true, "dispute": true, "dispute-respond": true, "opinion": true, "regrade": true, "mint": true, "close": true, "confidence": true, "cite": true, "finding": true, "observe": true, "avenue": true, "friction": true, "revision": true, "retire": true, "manifest-row": true, "dispose": true}
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
	dcov := map[string]int{} // dialectic-event coverage across all runs (proves the fuzz emits them)

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
			if o.err != "" {
				failures = append(failures, o)
			} else {
				os.RemoveAll(o.runDir) // keep only failing runs, for inspection
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	t.Logf("fuzzed %d debate runs · %d failed · verdicts=%v · rounds=%v\n  dialectic events emitted: %v", completed, len(failures), verdicts, roundHist, dcov)
	// A green fuzz that drove NEITHER the cite nor the finding path is a false green: the
	// lens stub emitted neither for the whole life of PR-1, so the record-only channel went
	// unexercised end-to-end. Assert the paths actually ran.
	for _, k := range []string{"cite", "finding"} {
		if dcov[k] == 0 {
			t.Errorf("fuzz drove ZERO %s events across %d runs — the lens record channel is unexercised (false green)", k, completed)
		}
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
