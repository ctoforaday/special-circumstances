package strikes

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

var t0 = time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

// The rule's own number: three failures of the same key inside the window is a spin.
func TestRecordTripsOnTheThirdStrike(t *testing.T) {
	s := State{Keys: map[string][]string{}}
	key := Key("Bash", []byte(`{"command":"go test ./..."}`))

	for i := 1; i < Threshold; i++ {
		count, tripped := Record(&s, key, t0.Add(time.Duration(i)*time.Second))
		if tripped {
			t.Fatalf("tripped early, at strike %d", i)
		}
		if count != i {
			t.Errorf("count = %d, want %d", count, i)
		}
	}
	count, tripped := Record(&s, key, t0.Add(3*time.Second))
	if !tripped || count != Threshold {
		t.Fatalf("third strike: count=%d tripped=%v", count, tripped)
	}
	// Reset after tripping: a persistent failure nags at 3, 6, 9 — not on every
	// subsequent attempt. A counter that fires forever is one people learn to ignore,
	// which leaves the rule as unenforced as it was.
	if len(s.Keys[key]) != 0 {
		t.Errorf("key was not reset after tripping: %v", s.Keys[key])
	}
	if _, tripped := Record(&s, key, t0.Add(4*time.Second)); tripped {
		t.Error("the attempt right after a trip must not trip again")
	}
}

// The window is what stands in for "consecutive", which this event cannot observe: a
// success never fires PostToolUseFailure. Three failures across an afternoon are three
// problems, not a loop.
func TestOldFailuresFallOutOfTheWindow(t *testing.T) {
	s := State{Keys: map[string][]string{}}
	key := Key("Bash", []byte(`{"command":"flaky"}`))

	Record(&s, key, t0)
	Record(&s, key, t0.Add(time.Second))
	if _, tripped := Record(&s, key, t0.Add(Window+time.Minute)); tripped {
		t.Error("failures older than the window must not add up to a spin")
	}

	// And inside the window they must.
	s = State{Keys: map[string][]string{}}
	Record(&s, key, t0)
	Record(&s, key, t0.Add(Window-2*time.Second))
	if _, tripped := Record(&s, key, t0.Add(Window-time.Second)); !tripped {
		t.Error("three failures inside the window is a spin")
	}
}

// Different targets are different attempts. The rule forbids re-running the SAME action,
// not failing twice at different things.
func TestDistinctTargetsDoNotAccumulate(t *testing.T) {
	s := State{Keys: map[string][]string{}}
	for _, cmd := range []string{"go test ./a", "go test ./b", "go test ./c"} {
		if _, tripped := Record(&s, Key("Bash", []byte(`{"command":"`+cmd+`"}`)), t0); tripped {
			t.Fatalf("three different commands must not trip: %s", cmd)
		}
	}
	// Same command under a different tool is also a different attempt.
	s = State{Keys: map[string][]string{}}
	for _, tool := range []string{"Bash", "Read", "Grep"} {
		if _, tripped := Record(&s, Key(tool, []byte(`{"command":"x"}`)), t0); tripped {
			t.Fatalf("different tools must not accumulate: %s", tool)
		}
	}
}

func TestKeyIdentifiesTheSameAttempt(t *testing.T) {
	// Whitespace differences are the same attempt; a reformatted command is not a new idea.
	a := Key("Bash", []byte(`{"command":"go   test    ./..."}`))
	b := Key("Bash", []byte(`{"command":"go test ./..."}`))
	if a != b {
		t.Errorf("whitespace must not split a key:\n%q\n%q", a, b)
	}
	// Each tool family's own notion of a subject.
	for _, c := range []struct{ input, want string }{
		{`{"command":"ls"}`, "ls"},
		{`{"file_path":"/p/a.go"}`, "/p/a.go"},
		{`{"path":"/p/b.go"}`, "/p/b.go"},
		{`{"pattern":"func Test"}`, "func Test"},
		{`{"url":"https://x/y"}`, "https://x/y"},
		{`{"query":"how do hooks work"}`, "how do hooks work"},
	} {
		if got := Key("T", []byte(c.input)); got != "T\x00"+c.want {
			t.Errorf("Key(%s) = %q", c.input, got)
		}
	}
	// No recognisable subject degrades to the tool alone rather than to an empty key,
	// so unrelated targetless failures still group under their tool.
	if got := Key("Bash", []byte(`{"unknown":"x"}`)); got != "Bash" {
		t.Errorf("targetless key = %q, want the bare tool", got)
	}
	if got := Key("Bash", nil); got != "Bash" {
		t.Errorf("nil input = %q, want the bare tool", got)
	}
}

// Losing strike history must never cost a tool call.
func TestLoadDegradesToEmpty(t *testing.T) {
	for name, read := range map[string]func(string) ([]byte, error){
		"missing":   func(string) ([]byte, error) { return nil, errors.New("nope") },
		"corrupt":   func(string) ([]byte, error) { return []byte("{{{"), nil },
		"null keys": func(string) ([]byte, error) { return []byte(`{"keys":null}`), nil },
	} {
		t.Run(name, func(t *testing.T) {
			s := Load(read, "x.json")
			if s.Keys == nil {
				t.Fatal("Load must always return a usable map")
			}
			if _, tripped := Record(&s, "k", t0); tripped {
				t.Error("a degraded load must not arrive pre-tripped")
			}
		})
	}
}

// A corrupt timestamp must not pin a key open forever.
func TestUnparseableStampsAreDropped(t *testing.T) {
	s := State{Keys: map[string][]string{"k": {"not-a-time", "also-not"}}}
	if count, tripped := Record(&s, "k", t0); tripped || count != 1 {
		t.Errorf("garbage stamps must not count: count=%d tripped=%v", count, tripped)
	}
}

// A session failing many DIFFERENT things is not spinning, and the state file must not grow
// without bound — the resource-ballooning failure this design has hit once already.
func TestStateStaysBounded(t *testing.T) {
	s := State{Keys: map[string][]string{}}
	for i := 0; i < maxKeys*2; i++ {
		Record(&s, fmt.Sprintf("Bash\x00cmd-%d", i), t0.Add(time.Duration(i)*time.Millisecond))
	}
	if len(s.Keys) > maxKeys {
		t.Errorf("state holds %d keys, want at most %d", len(s.Keys), maxKeys)
	}
	// Expired keys go regardless of count.
	s = State{Keys: map[string][]string{}}
	Record(&s, "old", t0)
	Record(&s, "new", t0.Add(Window+time.Minute))
	if _, ok := s.Keys["old"]; ok {
		t.Error("an expired key must be pruned")
	}
}
