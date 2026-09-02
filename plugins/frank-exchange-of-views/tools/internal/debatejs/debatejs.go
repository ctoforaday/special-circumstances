// Package debatejs runs the shipped debate.js orchestrator under goja and captures the prompts it
// hands its seats.
//
// # Why this exists
//
// The seat probe used to hand a dispatched seat a prompt WRITTEN IN THE HARNESS — a paraphrase of
// what production says. Three defects came out of that paraphrase before the fourth and largest was
// noticed: the probe's prompt was 950 characters where production's is 12,800, and every number the
// probe reported about what a seat finds in its help was a number about a prompt no seat is ever
// given. A measurement of a surface nobody is dispatched onto is not a weak measurement, it is a
// different experiment reported under the first one's name.
//
// So the prompt is no longer written here. It is EXECUTED out of debate.js, which is the same file
// production runs, by the same mechanism the fuzz already uses: goja, with agent() stubbed. A clause
// edited in debate.js reaches the probe on the next run with nothing to keep in step, because there
// is no second copy to keep in step with.
//
// # What a miss returns
//
// Capture returns an error for a seat it did not see. It does NOT return an empty prompt, and the
// caller cannot get one: a blank prompt dispatched to a seat measures the seat's willingness to
// invent a sitting, and reports it in the same column as a real reading. Every seam here is loud.
package debatejs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
)

// Dispatch is one agent() call debate.js made: the prompt a seat is given, and the routing that
// decides which constitution reads it.
type Dispatch struct {
	// SeatID is read from the prompt's own SEAT_ID field — the field recordClause writes and the
	// tool's roster gate reads, so it is the same key at both ends rather than a label parsed
	// back apart.
	SeatID string
	// Label, AgentType and Model are the opts debate.js dispatched under. AgentType is what the
	// probe must pass to `claude --agent`; anything else dispatches a seat under another seat's
	// constitution.
	Label     string
	AgentType string
	Model     string
	Prompt    string
	// Phase is the progress group, kept because a seat id alone does not say which round it sat.
	Phase string
}

// seatRe reads the seat id out of the prompt. recordClause writes `SEAT_ID: <id>.` into every
// dispatched prompt; a dispatch without one is a dispatch no seat can register against, and Capture
// keeps it with an empty SeatID rather than guessing from the label.
var seatRe = regexp.MustCompile(`SEAT_ID:\s*([A-Za-z0-9-]+)`)

// Envelope is what a stubbed seat returns. debate.js branches on these fields — the verdict, the
// gap list, the round record flags — so the backend supplying them decides which dispatch sites the
// capture run reaches at all.
type Envelope map[string]any

// Backend answers one dispatch. seatID is empty when the prompt carried no SEAT_ID (the frontier
// and repair re-prompts); label is always set.
type Backend func(seatID, label, prompt string) Envelope

// Config is the run debate.js is driven through. These are the `args` a real workflow invocation
// supplies; they appear inside the captured prompts, so a value that differs from the probe's own
// fixture produces a prompt naming a directory the seat cannot read.
type Config struct {
	Topic         string
	RunDir        string
	BinDir        string
	Lanes         int
	MaxRounds     int
	Model         string
	JudgmentModel string
	Backend       Backend
	// Timeout bounds the whole drive. A debate that never settles is a hang, and a hang inside a
	// probe build reads as a slow build.
	Timeout time.Duration
}

// Outcome is how a driven debate ENDED: the object debate.js returns at the bottom of the
// script, read back out of the settled module promise rather than inferred from the dispatch
// list.
//
// THE TERMINATION MODEL IS NOT WRITTEN DOWN ANYWHERE, and that is deliberate (#535 step 4).
// A hand-authored state machine of the round loop would be a second carrier of the loop's
// semantics, drifting from the loop the moment either moved — the defect class this repository
// names at every other altitude. So the model is the script: a caller enumerates the seat
// behaviours it wants to check, drives the real loop under each, and reads what it actually
// decided. Nothing here restates a rule; every field is a value debate.js computed.
type Outcome struct {
	// Verdict is the terminal stamp: HALTED, VERIFIED, CEILING or UNVERIFIED. debate.js
	// composes it from halted / red's verdict / the ceiling arm.
	Verdict string
	// Rounds is the loop counter at exit, not maxRounds.
	Rounds int
	// Deadlocked and Halted are the two non-ceiling terminators; both can be false at a
	// VERIFIED or CEILING exit.
	Deadlocked bool
	Halted     bool
	// BenchClearedBoard is what became of a board the bench cleared to zero on a deadlock
	// ruling: "" (it never did), "relief_granted" (red was given a further round to verdict
	// against the empty docket), or "ceiling_owed_red_a_sitting" (it cleared on the last round,
	// so there was no round to grant). A boolean here would answer two questions with one
	// false — never cleared, and cleared with nothing left to grant — and those are different
	// terminal facts: the second owes red a sitting.
	BenchClearedBoard string
	// GapsOutstanding is the board's open count as debate.js reports it — the assemble
	// seat's open_gaps where that seat returned an integer, red's docket length otherwise.
	// It is the field the historic UNVERIFIED-with-nothing-open defect was visible in.
	GapsOutstanding int
}

// ScriptPath locates the shipped orchestrator relative to the tools module root. One resolver: the
// probe and the fuzz must read the SAME file, and a second path expression is a second file waiting
// to happen.
func ScriptPath(toolsRoot string) string {
	return filepath.Join(toolsRoot, "..", "skills", "research-protocol", "scripts", "debate.js")
}

// preamble supplies the workflow runtime globals debate.js expects. It is deliberately the minimum:
// anything richer here is behaviour the real runtime may not have, and a prompt captured under it
// would be a prompt production never renders.
const preamble = `
globalThis.phase = function(p){ globalThis.__phase = p; }; globalThis.log = function(){};
globalThis.budget = { total:null, spent:function(){return 0;}, remaining:function(){return Infinity;} };
globalThis.parallel = async function(t){ return Promise.all(t.map(function(x){return x();})); };
globalThis.pipeline = async function(items){ var st=Array.prototype.slice.call(arguments,1);
  return Promise.all(items.map(async function(it,i){ var v=it; for(var s=0;s<st.length;s++){ try{v=await st[s](v,it,i);}catch(e){return null;} } return v; })); };
`

// wrap turns the ES module into something goja can run: the export becomes a plain const, and the
// top-level awaits go inside an async IIFE whose promise is settled after the loop drains.
func wrap(src string) string {
	s := strings.Replace(src, "export const meta", "const meta", 1)
	return "globalThis.__result = (async () => {\n" + s + "\n})()"
}

// programCache holds the COMPILED script, keyed by path and the file's modification time and
// size. Compiling debate.js is the expensive half of a drive — the script is ~1,300 lines — and a
// caller that drives it hundreds of times (the termination enumeration runs it once per schedule)
// otherwise pays that parse on every run for a file that has not changed.
//
// A goja.Program is immutable and safe to run on many runtimes, which is what makes this a cache
// rather than shared state. The key carries mtime AND size so an edit mid-session recompiles: a
// stale program would drive a debate.js nobody has, and report it as the shipped one.
var programCache sync.Map // string -> *goja.Program

func compiled(scriptPath string) (*goja.Program, error) {
	fi, err := os.Stat(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("stat debate.js at %s: %w", scriptPath, err)
	}
	key := fmt.Sprintf("%s|%d|%d", scriptPath, fi.ModTime().UnixNano(), fi.Size())
	if p, ok := programCache.Load(key); ok {
		return p.(*goja.Program), nil
	}
	b, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("read debate.js at %s: %w", scriptPath, err)
	}
	prog, err := goja.Compile(scriptPath, wrap(string(b)), false)
	if err != nil {
		return nil, fmt.Errorf("compile debate.js at %s: %w", scriptPath, err)
	}
	programCache.Store(key, prog)
	return prog, nil
}

// Capture drives debate.js once and returns every agent() call it made, in dispatch order.
//
// A run that throws is NOT an error here. debate.js aborts on envelopes its invariants reject, and
// the prompts dispatched BEFORE the abort are real prompts, rendered by the real clauses. The error
// return is reserved for the seams where nothing was captured at all — an unreadable script, a
// preamble failure, a hang — because those are the cases where a caller would otherwise proceed
// with an empty list that looks exactly like a debate that dispatched nobody.
func Capture(scriptPath string, cfg Config) ([]Dispatch, error) {
	d, err := drive(scriptPath, cfg)
	return d.dispatches, err
}

// CaptureRun drives debate.js once and returns BOTH the dispatches and how the run ended.
//
// It is stricter than Capture in exactly one place, and the difference is the point. Capture
// tolerates a debate that throws, because the prompts rendered before the throw are real. A
// caller asking how the run TERMINATED cannot be handed the same tolerance: a thrown run has no
// terminal verdict, and a zero-valued Outcome would report an empty verdict in the same column a
// real one occupies. So an unsettled or rejected module promise is an error here, naming which.
func CaptureRun(scriptPath string, cfg Config) ([]Dispatch, Outcome, error) {
	d, err := drive(scriptPath, cfg)
	if err != nil {
		return d.dispatches, Outcome{}, err
	}
	if d.outcome == nil {
		return d.dispatches, Outcome{}, fmt.Errorf("debate.js dispatched %d seat(s) and never settled to a terminal result — %s", len(d.dispatches), d.noOutcome)
	}
	return d.dispatches, *d.outcome, nil
}

// driven is one drive of debate.js: the dispatches it made, how it ended, and — when it did not
// end — WHY. The reason is carried rather than dropped: an outcome that is simply absent is the
// silent no-match this codebase keeps finding, and the caller that needs the outcome needs the
// reason more than it needs the nil.
type driven struct {
	dispatches []Dispatch
	outcome    *Outcome
	// noOutcome is the reason there is no outcome, empty when there is one.
	noOutcome string
}

func drive(scriptPath string, cfg Config) (driven, error) {
	prog, err := compiled(scriptPath)
	if err != nil {
		return driven{}, err
	}
	if cfg.Backend == nil {
		return driven{}, fmt.Errorf("debatejs.Capture: no backend — a nil backend resolves every seat to undefined and debate.js aborts on the first envelope, which would report as 'production dispatches one seat'")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	var got []Dispatch
	var settledErr string
	var out *Outcome
	var resultErr string

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if p := recover(); p != nil {
				settledErr = fmt.Sprintf("panic driving debate.js: %v", p)
			}
		}()
		loop := eventloop.NewEventLoop()
		loop.Run(func(vm *goja.Runtime) {
			vm.Set("args", map[string]any{
				"topic": cfg.Topic, "runDir": cfg.RunDir, "binDir": cfg.BinDir,
				"lanes": cfg.Lanes, "laneFloorOverride": "seatprobe",
				"maxRounds": cfg.MaxRounds,
				"model":     cfg.Model, "judgmentModel": cfg.JudgmentModel,
			})
			if _, err := vm.RunString(preamble); err != nil {
				settledErr = "preamble: " + err.Error()
				return
			}
			vm.Set("agent", func(call goja.FunctionCall) goja.Value {
				d := Dispatch{}
				if len(call.Arguments) > 0 {
					d.Prompt = call.Argument(0).String()
				}
				if m := seatRe.FindStringSubmatch(d.Prompt); m != nil {
					d.SeatID = m[1]
				}
				if len(call.Arguments) > 1 {
					o := call.Argument(1).ToObject(vm)
					d.Label = optStr(o, "label")
					d.AgentType = optStr(o, "agentType")
					d.Model = optStr(o, "model")
					d.Phase = optStr(o, "phase")
				}
				got = append(got, d)
				env := cfg.Backend(d.SeatID, d.Label, d.Prompt)
				p, resolve, _ := vm.NewPromise()
				resolve(vm.ToValue(map[string]any(env)))
				return vm.ToValue(p)
			})
			if _, err := vm.RunProgram(prog); err != nil {
				settledErr = "run: " + err.Error()
			}
		})
		if settledErr != "" {
			return
		}
		// Drain. The debate's own rejection is not reported HERE: an aborted debate still
		// rendered every prompt it dispatched before aborting, and those are the ones Capture
		// is after. It IS reported on the outcome channel below, where a caller asking how the
		// run ended must not be handed a zero.
		loop.Run(func(vm *goja.Runtime) {})
		loop.Run(func(vm *goja.Runtime) { out, resultErr = readResult(vm) })
	}()

	select {
	case <-done:
	case <-time.After(cfg.Timeout):
		return driven{}, fmt.Errorf("debate.js never settled within %s — a hang, not an empty run", cfg.Timeout)
	}
	if settledErr != "" && len(got) == 0 {
		return driven{}, fmt.Errorf("debate.js dispatched nobody: %s", settledErr)
	}
	if len(got) == 0 {
		return driven{}, fmt.Errorf("debate.js ran and dispatched nobody — the capture backend never saw a seat, which cannot be right for any config that reaches the frontier")
	}
	if out == nil && resultErr == "" {
		// The drive returned before the result could be read at all. settledErr is the reason
		// where there is one — a synchronous throw out of RunString, or a panic — and carrying
		// it beats the generic sentence, which would report a KNOWN cause as an unknown one.
		if settledErr != "" {
			resultErr = settledErr
		} else {
			resultErr = "the result was never read — readResult returned neither an outcome nor a reason"
		}
	}
	return driven{dispatches: got, outcome: out, noOutcome: resultErr}, nil
}

// readResult reads the value debate.js returned out of the settled module promise.
//
// EVERY MISS IS A REASON, NEVER A ZERO. A pending promise, a rejection, a non-object result and a
// result with no verdict are four different failures, and an Outcome{} would report all four as a
// run that ended with an empty verdict and nothing outstanding — which is a state the loop can
// genuinely produce, so the honest zero and the failed read would be the same bytes.
func readResult(vm *goja.Runtime) (*Outcome, string) {
	v := vm.Get("__result")
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return nil, "debate.js left no __result — the wrapper did not install one"
	}
	p, ok := v.Export().(*goja.Promise)
	if !ok {
		return nil, fmt.Sprintf("__result is %T, not a promise — the wrapper changed shape", v.Export())
	}
	switch p.State() {
	case goja.PromiseStatePending:
		return nil, "the debate promise is still PENDING after the loop drained — the run hung rather than ending"
	case goja.PromiseStateRejected:
		return nil, "the debate THREW: " + fmt.Sprint(p.Result())
	}
	m, ok := p.Result().Export().(map[string]any)
	if !ok {
		return nil, fmt.Sprintf("the debate settled to %T, not the object debate.js returns", p.Result().Export())
	}
	verdict, _ := m["verdict"].(string)
	if verdict == "" {
		return nil, "the debate settled to an object with no verdict — the return shape moved and this reader did not"
	}
	return &Outcome{
		Verdict:           verdict,
		Rounds:            asInt(m["rounds"]),
		Deadlocked:        asBool(m["deadlocked"]),
		Halted:            asBool(m["halted"]),
		BenchClearedBoard: asString(m["bench_cleared_board"]),
		GapsOutstanding:   asInt(m["gaps_outstanding"]),
	}, ""
}

// asInt reads a JS number back. goja exports integral numbers as int64 and the rest as float64, so
// both are read rather than one being assumed.
func asInt(v any) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func asBool(v any) bool { b, _ := v.(bool); return b }

// asString reads a nullable JS string back. debate.js sends null for "the bench never cleared the
// board", which exports as nil, so the empty string is that state and nothing else.
func asString(v any) string { s, _ := v.(string); return s }

// optStr reads a string option, returning "" for absent/undefined/null so an unset opt and an empty
// one are the same to the caller — which they are: both dispatch under the default.
func optStr(o *goja.Object, k string) string {
	if o == nil {
		return ""
	}
	v := o.Get(k)
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return ""
	}
	return v.String()
}

// For returns the dispatch debate.js made for seatID.
//
// THE MISS IS AN ERROR, NOT A ZERO. A seat id the debate never dispatched is a probe pointed at a
// seat production does not seat — the exact defect this package was built after. Returning an empty
// Dispatch would let the probe dispatch a blank prompt and score the answer.
func For(ds []Dispatch, seatID string) (Dispatch, error) {
	for _, d := range ds {
		if d.SeatID == seatID {
			return d, nil
		}
	}
	seen := map[string]bool{}
	var ids []string
	for _, d := range ds {
		if d.SeatID != "" && !seen[d.SeatID] {
			seen[d.SeatID] = true
			ids = append(ids, d.SeatID)
		}
	}
	return Dispatch{}, fmt.Errorf("debate.js dispatched no seat %q in this run; it dispatched: %s", seatID, strings.Join(ids, ", "))
}
