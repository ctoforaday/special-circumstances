package seatprobe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE PROBE'S OWN PROMPT IS A CARRIER OF THE IDENTITY CONTRACT, and it spoke the old model for a
// full run after the contract changed.
//
// It said "Your identity and run are INJECTED: do not pass --run or --seat-id." That was TRUE
// while Build registered each seat under the handle it would later be dispatched with — the
// binding existed before the seat ran, so the seat genuinely never needed the flag.
//
// When Build stopped pre-binding, the seat began arriving UNBOUND and had to pass --seat-id once,
// at register. The prompt forbade exactly that. MEASURED: 8 of 9 seats were refused on their first
// or second call — "the surface is scoped to whoever is asking" — every one of them obeying an
// instruction the tool now refuses. They all recovered on the next call, which is why it read as a
// finding about seats rather than as a defect in the harness.
//
// A gate rather than a comment, because this carrier is prose in a format string: nothing compiles
// it against the contract, and the last one to go stale cost a nine-board run.
func TestTheActingPromptTeachesTheBindingContract(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "cmd", "seatprobe", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	i := strings.Index(src, "You are the %s seat in a frank-exchange-of-views run")
	if i < 0 {
		t.Fatal("the acting prompt is not where this gate looks — a scan that finds nothing passes forever")
	}
	prompt := src[i:]
	if j := strings.Index(prompt, "`,"); j > 0 {
		prompt = prompt[:j]
	}

	// The run IS injected, and saying so is what stops a seat typing an absolute path it can
	// mistype — the measured defect FEOV_RUN exists for.
	if !strings.Contains(prompt, "never pass --run") {
		t.Error("the prompt no longer tells the seat its run is injected")
	}
	// The identity is NOT injected. Saying it is sends the seat into a refusal on its first act.
	if strings.Contains(prompt, "do not pass --run or --seat-id") {
		t.Error("the prompt still says the identity is injected — it is bound at register, and a seat obeying this is refused on its opening call")
	}
	for _, want := range []string{"register", "--seat-id", "ONCE"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not carry %q: the seat has to be told it states its id once, at register", want)
		}
	}
}
