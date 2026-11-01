package difftest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// TWO CHECKERS EXECUTE ONE BUNDLE AT ONCE, AND THE LATER STANDS (plans/compute-and-show.md §III.1.4,
// V1a.5 (c)), driven through two processes of the shipped binary released together on a bundle
// nobody has executed. The program holds each execution until both have started, so both read "no
// standing re-run", both execute, store and append, and the two executions print different lines.
// A later act that executes nothing follows the LATER event's rerun_sha, read off the record:
// tampering the earlier execution's directory leaves its verdict alone, tampering the later's moves
// it.
func TestTwoConcurrentReproducesBothStandAndTheLaterIsCompared(t *testing.T) {
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	m := newMapper()
	prepareRun(t, bin, runDir, m, nil, nil)
	quote := func(s string) string { b, _ := json.Marshal(s); return string(b) }
	arrivals := filepath.Join(runDir, "arrivals.txt")
	dir := filepath.Join(runDir, "together")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Each execution appends its pid and reads its arrival. compute's two runs are the first two and
	// print `same`; a later one waits until four have arrived, then prints its arrival.
	body := `const fs = require("fs");
const at = ` + quote(arrivals) + `;
fs.appendFileSync(at, process.pid + "\n");
const nth = () => fs.readFileSync(at, "utf8").trim().split("\n");
const k = nth().indexOf(String(process.pid)) + 1;
if (k <= 2) { console.log("same"); process.exit(0); }
const until = Date.now() + 20000;
while (nth().length < 4 && Date.now() < until) Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 20);
console.log("run " + k);
`
	if err := os.WriteFile(filepath.Join(dir, "together.js"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	do := func(verb string, args ...string) invocation {
		t.Helper()
		return runGo(bin, runDir, m, cmd{verb: verb, args: append([]string{"--run", "{RUN}"}, args...)})
	}
	if c := do("compute", "--seat-id", "blue-respond", "--quote", "A claim sits under S2.", "--dir", dir, "--reason", "r", "--", "node", "together.js"); c.code != 0 {
		t.Fatalf("compute: exit %d\n%s%s", c.code, c.stdout, c.stderr)
	}
	run := runtest.Open(t, runDir)
	ks, err := record.RecordedComputations(run)
	if err != nil || len(ks) != 1 {
		t.Fatalf("computations: %v %v", ks, err)
	}
	k := ks[0]
	lenses := []string{"red-lens-logic", "red-lens-computation"}
	for _, l := range lenses {
		if r := do("register", "--seat-id", l); r.code != 0 {
			t.Fatalf("register %s: %s", l, r.stderr)
		}
	}
	reproduce := func(seat string) invocation {
		return do("reproduce", "--seat-id", seat, "--anchor", k.Label, "--as", "unsound", "--reason", "it prints its arrival, not the claim", "--json")
	}

	var wg sync.WaitGroup
	results := make([]invocation, 2)
	for i, l := range lenses {
		wg.Add(1)
		go func() { defer wg.Done(); results[i] = reproduce(l) }()
	}
	wg.Wait()
	for i, r := range results {
		if r.code != 0 || !strings.Contains(r.stdout, `"ran":true`) {
			t.Fatalf("concurrent reproduce %d: exit %d — want both to execute and be recorded\n%s%s", i, r.code, r.stdout, r.stderr)
		}
	}
	var ran []*recordpb.Reproduce
	m2, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m2.Events {
		if r, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok && r.GetRan() {
			ran = append(ran, r)
		}
	}
	if len(ran) != 2 || ran[0].GetRerunSha() == ran[1].GetRerunSha() {
		t.Fatalf("two concurrent executions recorded %d `ran` event(s), want two naming different executions: %v", len(ran), ran)
	}
	earlier, later := ran[0].GetRerunSha(), ran[1].GetRerunSha()
	// A DIRECTORY'S AGE IS NOT ITS STANDING: the earlier execution's directory is made the newest, so
	// an act that picked the newest directory would read the execution that does not stand.
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(compute.RerunPath(runDir, k.SHA, earlier), future, future); err != nil {
		t.Fatal(err)
	}

	judge := func() string {
		t.Helper()
		r := reproduce("red-lens-computation")
		if r.code != 0 || !strings.Contains(r.stdout, `"ran":false`) || !strings.Contains(r.stdout, `"rerun_sha":"`+later+`"`) {
			t.Fatalf("an act after both: exit %d — want it to execute nothing and compare against the later execution %s\n%s%s", r.code, later[:12], r.stdout, r.stderr)
		}
		var env struct {
			Result struct {
				Rerun string `json:"rerun"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(r.stdout), &env); err != nil {
			t.Fatal(err)
		}
		return env.Result.Rerun
	}
	if got := judge(); got != "differs" {
		t.Fatalf("against the later execution the act reads %s, want differs", got)
	}
	tamper := func(sha string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(compute.RerunPath(runDir, k.SHA, sha), "output"), []byte("same\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tamper(earlier)
	if got := judge(); got != "differs" {
		t.Errorf("tampering the earlier execution moved the verdict to %s — the act read a re-run that does not stand", got)
	}
	tamper(later)
	if got := judge(); got != "reproduced" {
		t.Errorf("tampering the later execution left the verdict %s — the act did not read the re-run that stands", got)
	}
}
