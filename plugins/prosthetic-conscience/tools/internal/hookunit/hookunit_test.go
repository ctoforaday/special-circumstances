package hookunit

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var noon = time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

func ctx(raw string) *Ctx { return NewCtx("PostToolUse", []byte(raw), "/proj", noon) }

func TestNewCtxParsesThePayloadOnce(t *testing.T) {
	c := ctx(`{"tool_name":"Edit","tool_input":{"file_path":"/proj/a.go"}}`)
	if c.ToolName != "Edit" {
		t.Errorf("ToolName = %q", c.ToolName)
	}
	if !strings.Contains(string(c.ToolInput), "a.go") {
		t.Errorf("ToolInput = %q", c.ToolInput)
	}
	// Malformed input must not panic; units decide what to do with an empty payload.
	if c := ctx("{{{"); c.ToolName != "" {
		t.Errorf("malformed payload should yield an empty tool name, got %q", c.ToolName)
	}
}

// THE REASON THE CONTEXT IS SHARED: the second unit to want an expensive read pays nothing.
// Asserted under concurrency, because that is the only condition it has to hold in.
func TestMemoBuildsExactlyOnceUnderConcurrency(t *testing.T) {
	c := ctx("{}")
	var builds int32
	var wg sync.WaitGroup
	got := make([]any, 50)
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i] = c.Memo("transcript", func() any {
				atomic.AddInt32(&builds, 1)
				time.Sleep(2 * time.Millisecond) // a read worth memoising
				return "parsed-once"
			})
		}(i)
	}
	wg.Wait()

	if n := atomic.LoadInt32(&builds); n != 1 {
		t.Errorf("built %d times, want exactly 1 — this is the whole point of the shared context", n)
	}
	for i, v := range got {
		if v != "parsed-once" {
			t.Fatalf("caller %d got %v, want every caller to receive the same value", i, v)
		}
	}
	// A different key is a different read.
	c.Memo("other", func() any { return 1 })
	if n := atomic.LoadInt32(&builds); n != 1 {
		t.Errorf("an unrelated key rebuilt the first: %d", n)
	}
}

func TestRunHonoursEachUnitsOwnMatcher(t *testing.T) {
	var ran []string
	var mu sync.Mutex
	unit := func(name string, applies bool) Unit {
		return Unit{Name: name,
			Applies: func(*Ctx) bool { return applies },
			Run: func(*Ctx) Result {
				mu.Lock()
				ran = append(ran, name)
				mu.Unlock()
				return Result{Name: name}
			}}
	}
	res := Run(ctx("{}"), []Unit{unit("yes", true), unit("no", false), unit("also", true)})
	if len(res) != 2 || res[0].Name != "yes" || res[1].Name != "also" {
		t.Fatalf("results = %+v; want only the applicable units, in unit order", res)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, n := range ran {
		if n == "no" {
			t.Error("a unit whose matcher said no was run — merging must not widen a matcher")
		}
	}
}

// Results must be deterministic in UNIT order however the scheduler interleaves them,
// or the operator sees a different message order run to run.
func TestRunReturnsResultsInUnitOrder(t *testing.T) {
	slowThenFast := []Unit{
		{Name: "slow", Run: func(*Ctx) Result {
			time.Sleep(20 * time.Millisecond)
			return Result{Name: "slow"}
		}},
		{Name: "fast", Run: func(*Ctx) Result { return Result{Name: "fast"} }},
	}
	for i := 0; i < 20; i++ {
		res := Run(ctx("{}"), slowThenFast)
		if len(res) != 2 || res[0].Name != "slow" || res[1].Name != "fast" {
			t.Fatalf("run %d: %+v — order must follow the unit list, not completion", i, res)
		}
	}
}

// Units must actually overlap. Parity with what the OS gave away for free is the FLOOR of
// this design; a sequential runner would be a regression, not a refactor.
func TestRunExecutesUnitsConcurrently(t *testing.T) {
	const hold = 60 * time.Millisecond
	sleeper := func(name string) Unit {
		return Unit{Name: name, Run: func(*Ctx) Result {
			time.Sleep(hold)
			return Result{Name: name}
		}}
	}
	start := time.Now()
	Run(ctx("{}"), []Unit{sleeper("a"), sleeper("b"), sleeper("c")})
	elapsed := time.Since(start)

	if elapsed > 2*hold {
		t.Errorf("three %v units took %v — that is serial, and serial is a REGRESSION against process-per-hook", hold, elapsed)
	}
}

// A panicking unit must not take the event. A separate process gave this isolation away
// for free; one binary has to provide it deliberately.
func TestRunIsolatesAPanickingUnit(t *testing.T) {
	res := Run(ctx("{}"), []Unit{
		{Name: "boom", Run: func(*Ctx) Result { panic("unit exploded") }},
		{Name: "survivor", Run: func(*Ctx) Result { return Result{Name: "survivor", Stderr: "still here"} }},
	})
	if len(res) != 2 {
		t.Fatalf("results = %+v; a panic must not remove the other unit's result", res)
	}
	if !strings.Contains(res[0].Stderr, "panicked and was isolated") || !strings.Contains(res[0].Stderr, "unit exploded") {
		t.Errorf("the panic must be reported, with its value: %q", res[0].Stderr)
	}
	if res[1].Stderr != "still here" {
		t.Errorf("the surviving unit lost its result: %+v", res[1])
	}
}

func TestRunOnNoUnits(t *testing.T) {
	if res := Run(ctx("{}"), nil); len(res) != 0 {
		t.Errorf("no units must yield no results, got %+v", res)
	}
}
