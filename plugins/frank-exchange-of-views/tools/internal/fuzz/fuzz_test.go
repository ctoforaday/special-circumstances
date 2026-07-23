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
	disputed    map[string]bool // gaps blue has disputed, awaiting red's dispute-respond
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
	if role == "blue" && len(open) > 0 && r.coin(40) {
		id := open[r.rng.Intn(len(open))]
		if _, err := r.exec("blue", "dispute", "--seat-id", seatID, "--id", id, "--dimension", "impact", "--proposed", r.g(), "--reason", "dispute-evidence-for-"+id); err == nil {
			r.disputed[id] = true
		}
	}
	if role == "merge" {
		for id := range r.disputed { // answer every open dispute (the tool refuses a dangling one)
			_, _ = r.exec("merge", "dispute-respond", "--seat-id", seatID, "--id", id, "--as", "rejected", "--reason", "respond-rationale-for-"+id)
			delete(r.disputed, id)
		}
		if len(open) > 0 && r.coin(30) {
			id := open[r.rng.Intn(len(open))]
			_, _ = r.exec("merge", "regrade", "--seat-id", seatID, "--id", id, "--reason", "regrade-basis-for-"+id, "--likelihood", r.g())
		}
	}
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

// arr / obj are goja-friendly envelope builders.
func arr(v ...any) []any { return append([]any{}, v...) }

// envelopeFor performs the seat's tool acts and returns the envelope debate.js expects. The
// randomisation here is what drives debate.js's control flow: red's verdict and gap count
// decide whether the loop continues, deadlocks, or terminates.
func (r *runner) envelopeFor(seatID string) map[string]any {
	switch {
	case strings.HasPrefix(seatID, "blue-synthesize"):
		r.register("blue", seatID)
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "petitions": arr(), "friction": arr()}

	case strings.HasPrefix(seatID, "red-merge"):
		r.register("merge", seatID)
		// PASS ~40%: run the round's dialectic, then close every open gap so the #67 gate
		// (and verify) holds.
		if r.coin(40) {
			open := r.openGaps()
			r.dialectic("merge", seatID, open)
			for _, id := range r.openGaps() { // re-read: a regression close may have added a successor
				r.closeGap(seatID, id)
			}
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": arr(), "corroboration": arr(), "petitions": arr(), "friction": arr()}
		}
		// FAIL: mint 1-3 fresh gaps (a FAIL with empty gaps is a degenerate merge debate.js
		// rejects on purpose). Report the FULL open set as the docket.
		n := r.rng.Intn(3) + 1
		for i := 0; i < n; i++ {
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
			return map[string]any{"verdict": "PASS", "gaps": arr(), "closures": arr(), "dispute_responses": arr(), "corroboration": arr(), "petitions": arr(), "friction": arr()}
		}
		return map[string]any{"verdict": "FAIL", "gaps": gaps, "closures": arr(), "dispute_responses": arr(), "corroboration": arr(), "petitions": arr(), "friction": arr()}

	case strings.HasPrefix(seatID, "blue-respond"):
		r.register("blue", seatID)
		open := r.openGaps()
		r.dialectic("blue", seatID, open) // blue's position, closings, and maybe a grade dispute
		// debate.js rejects an EMPTY manifest on a round with open gaps — a repair must show its
		// receipt. One row per open gap it is repairing this round.
		var manifest []any
		for _, id := range open {
			manifest = append(manifest, map[string]any{"gap_id": id, "row": "fuzz: checked"})
		}
		return map[string]any{"round_record_appended": true, "claim_count": r.rng.Intn(40) + 10, "manifest": manifest, "grade_disputes": arr(), "petitions": arr(), "friction": arr()}

	case strings.HasPrefix(seatID, "judge-petition"):
		r.register("bench", seatID)
		return map[string]any{"rulings": arr(), "friction": arr()}

	case strings.HasPrefix(seatID, "judge"): // adjudication + terminal
		r.register("bench", seatID)
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

	default: // frontier, blue lanes, red lenses — register if identifiable, minimal envelope
		if seatID != "" {
			role := "lens"
			switch {
			case strings.HasPrefix(seatID, "blue"):
				role = "blue"
			case strings.HasPrefix(seatID, "frontier"):
				role = "lens"
			}
			r.register(role, seatID)
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
	r := &runner{bin: bin, runDir: runDir, rng: rand.New(rand.NewSource(seed)), registered: map[string]bool{}, disputed: map[string]bool{}}

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
				"model": nil, "judgmentModel": nil,
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
	// Oracle 1: the record the run left must pass verify.
	if out, err := exec.Command(bin, "verify", "--run", runDir).CombinedOutput(); err != nil {
		res.err = "verify FAILED:\n" + truncate(string(out))
		return res
	}
	// Oracle 2: every dialectic event's prose must actually RENDER in the report — the A1-A3
	// class (prose written under one key, read under another) is invisible to verify but caught
	// here, on every run.
	if m := proseRenders(runDir); m != "" {
		res.err = m
	}
	res.dialectic = tallyDialectic(runDir)
	return res
}

// tallyDialectic counts the events that prove the fuzz exercised the paths it claims to.
func tallyDialectic(runDir string) map[string]int {
	board, err := record.BoardState(runDir)
	if err != nil {
		return nil
	}
	want := map[string]bool{"closing": true, "position": true, "dispute": true, "dispute-respond": true, "opinion": true, "regrade": true, "mint": true, "close": true}
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
}

func proseRenders(runDir string) string {
	report, err := os.ReadFile(filepath.Join(runDir, "report.md"))
	if err != nil {
		return "no report.md: " + err.Error()
	}
	rpt := string(report)
	board, err := record.BoardState(runDir)
	if err != nil {
		return "board: " + err.Error()
	}
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
	concurrency := 12
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
