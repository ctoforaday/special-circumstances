// Package hookunit runs several checks on ONE hook event, over ONE shared context.
//
// WHY THIS SHAPE (#201 step 3). Until now each check was its own binary and hooks.json
// registered several per event. Measured (plans/hook-surface-spike.md §4b): the client runs
// those in PARALLEL, so process-per-check was optimal for two trivial, independent units —
// wall clock was the max, not the sum.
//
// That is a LOCAL MAXIMUM, and it holds only while units SHARE NOTHING. Every direction this
// suite is heading breaks that:
//
//   - Shared expensive reads. N units each parsing the same transcript, SQLite store or git
//     state is N x the dominant cost, and process parallelism makes it WORSE rather than
//     better: N concurrent readers of one file, N connections, page-cache contention. One
//     process reads once and hands the result to every unit.
//   - A shared prelude. "Validate the VCS state first" is work every unit needs and exactly
//     one should do. Across processes that needs a cache file, which needs staleness
//     handling — a new defect surface invented to avoid doing the work once.
//   - Fan-out inside a unit. By-file goroutines here use ONE pool sized to NumCPU. Across N
//     processes it is N pools oversubscribing the machine with no global scheduler.
//
// So the shared Ctx is the point, not the concurrency: the concurrency merely restores what
// the OS was already giving away. PARITY IS THE FLOOR, not the goal.
//
// # What this package deliberately does not do
//
// It does not merge decisions. Each event's binary owns that, because the events genuinely
// disagree: PreToolUse can DENY, PostToolUse's exit 2 is a feedback channel over a write
// that already happened, and SessionStart composes one response from several. Generalising
// a merge policy from one example would be inventing an abstraction ahead of its instances.
package hookunit

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Ctx is everything the units share, built ONCE per event.
//
// Expensive reads belong here behind Once, so the second unit to want one pays nothing.
// Today the shared state is a small payload and a path; the type exists so that stays true
// as the reads get expensive.
type Ctx struct {
	Event      string
	Raw        []byte // the payload exactly as received
	ProjectDir string
	ToolName   string
	ToolInput  json.RawMessage
	Now        time.Time

	mu    sync.Mutex
	memo  map[string]any
	onces map[string]*sync.Once
}

// NewCtx builds a context. Units must treat it as READ-ONLY apart from Memo.
func NewCtx(event string, raw []byte, projectDir string, now time.Time) *Ctx {
	c := &Ctx{Event: event, Raw: raw, ProjectDir: projectDir, Now: now,
		memo: map[string]any{}, onces: map[string]*sync.Once{}}
	var in struct {
		ToolName  string          `json:"tool_name"`
		ToolInput json.RawMessage `json:"tool_input"`
	}
	_ = json.Unmarshal(raw, &in)
	c.ToolName, c.ToolInput = in.ToolName, in.ToolInput
	return c
}

// Memo runs build exactly once per key, however many units ask for it concurrently, and
// returns the same value to all of them. This is the whole reason the context is shared:
// the second reader of a transcript should pay nothing.
func (c *Ctx) Memo(key string, build func() any) any {
	c.mu.Lock()
	once, ok := c.onces[key]
	if !ok {
		once = &sync.Once{}
		c.onces[key] = once
	}
	c.mu.Unlock()

	once.Do(func() {
		v := build()
		c.mu.Lock()
		c.memo[key] = v
		c.mu.Unlock()
	})

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.memo[key]
}

// Result is what one unit has to say. Exit is the code that unit WANTS; the event's binary
// decides what to do with several of them.
type Result struct {
	Name   string
	Stdout string
	Stderr string
	Exit   int
	// Log is the unit's hook-log line, returned rather than written. Two processes used to
	// append to one log file concurrently — measured to be concurrent BY DESIGN once hooks
	// were known to run in parallel. Collecting the lines and writing them from one place
	// is the fix that consolidation exists to make possible.
	Log string
}

// Unit is one check on one event.
type Unit struct {
	Name string
	// Applies is the unit's OWN matcher. hooks.json now registers the UNION of what an
	// event's units want, so each unit must still say whether this call is its business —
	// otherwise merging would silently widen a matcher (sc-push-freeze-guard watches Bash;
	// sc-secrets-gate also watches WebFetch and WebSearch).
	Applies func(*Ctx) bool
	Run     func(*Ctx) Result
}

// Run executes the applicable units concurrently and returns their results IN UNIT ORDER,
// so output is deterministic however the scheduling falls.
//
// NOT BOUNDED BY NumCPU, and that was measured rather than reasoned: an earlier version
// sized a semaphore to runtime.NumCPU(), and CI on a 2-core runner took 120ms for three
// 60ms units — the third queued. These units are I/O- and SUBPROCESS-bound (sc-quality-gate
// waits on qlty, sc-recall-index on qmd); sizing them to cores makes them queue for a
// resource they are not competing for, which is the serialisation this design exists to
// avoid. N is bounded by the units registered on one event, which is small by construction.
//
// A NumCPU pool belongs one level down, inside a unit that fans out over FILES — that work
// is CPU-bound and does need a bound, and it needs ONE shared pool rather than N of them,
// which is another thing a single process can provide and separate binaries cannot.
//
// A panicking unit yields a Result naming the panic and never takes down the event: losing
// one check must not cost the others, which is the isolation a separate process gave for
// free and this has to provide deliberately.
func Run(ctx *Ctx, units []Unit) []Result {
	out := make([]Result, len(units))
	var wg sync.WaitGroup
	for i, u := range units {
		if u.Applies != nil && !u.Applies(ctx) {
			continue
		}
		wg.Add(1)
		go func(i int, u Unit) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					out[i] = Result{Name: u.Name,
						Stderr: fmt.Sprintf("%s: panicked and was isolated: %v", u.Name, r)}
				}
			}()
			out[i] = u.Run(ctx)
		}(i, u)
	}
	wg.Wait()

	kept := out[:0]
	for _, r := range out {
		if r.Name != "" {
			kept = append(kept, r)
		}
	}
	return kept
}
