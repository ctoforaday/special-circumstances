package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// storeEntries counts every directory under the run's computation store, so a test can say a call
// wrote nothing there.
func storeEntries(t *testing.T, runDir string) int {
	t.Helper()
	n := 0
	filepath.WalkDir(filepath.Join(runDir, compute.StoreDir), func(_ string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			n++
		}
		return nil
	})
	return n
}

// A KEYED RETRY IS SETTLED BY THE BUNDLE'S IDENTITY, WITHOUT RUNNING ANYTHING. An identical retry is
// idempotent and executes nothing (the program counts its runs in a file outside the bundle) and
// writes nothing; the same key over other bytes or another command is a conflict naming both shas;
// another --variance is a conflict naming the stored one. None of the refused retries records.
func TestAKeyedComputeRetryIsSettledByTheBundle(t *testing.T) {
	runDir := newRun(t)
	const sentence = "The sampled value varies between runs."
	seat := computeSeat(t, runDir, "# H\n\n"+sentence+"\n")
	counter := filepath.Join(t.TempDir(), "runs")
	body := "const fs=require('fs');fs.appendFileSync(process.argv[2],'x');console.log(Math.random());"
	s := script(t, runDir, "sample.js", body)
	args := []string{"compute", "--run", runDir, "--seat-id", seat, "--key", "S1", "--dir", s, "--quote", sentence,
		"--reason", "sampling a live system", "--variance", "a number in [0, 1)", "--", "node", "sample.js", counter}
	if out, err := run(t, args...); err != nil || !strings.Contains(out, "variant") {
		t.Fatalf("compute = %q, %v; want a variant computation", out, err)
	}
	ran := func() int {
		b, _ := os.ReadFile(counter)
		return len(b)
	}
	if ran() != 2 {
		t.Fatalf("the program ran %d times, want 2", ran())
	}
	events, store := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION), storeEntries(t, runDir)

	if out, err := run(t, args...); err != nil || !strings.Contains(out, "idempotent") {
		t.Errorf("an identical retry = %q, %v; want the stored computation, idempotent", out, err)
	}
	if ran() != 2 || storeEntries(t, runDir) != store {
		t.Errorf("an identical retry ran the program (%d runs) or wrote the store (%d → %d entries)", ran(), store, storeEntries(t, runDir))
	}

	swap := func(from, to string) []string {
		out := append([]string(nil), args...)
		for i := range out {
			if out[i] == from {
				out[i] = to
			}
		}
		return out
	}
	other := filepath.Join(runDir, s, "sample.js")
	// In order: the bundle's bytes change last, since every case after it would meet that first.
	for _, c := range []struct {
		name string
		args []string
		says string
		prep func()
	}{
		{"another --variance", swap("a number in [0, 1)", "a number in [0, 2)"), `--variance "a number in [0, 1)"`, nil},
		{"another argument", append(append([]string(nil), args...), "extra"), "hashes to", nil},
		{"an argument holding a space", swap(counter, counter+" x"), "hashes to", nil},
		{"other bytes", args, "hashes to", func() {
			if err := os.WriteFile(other, []byte(body+"\n// changed"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if c.prep != nil {
				c.prep()
			}
			_, err := run(t, c.args...)
			if err == nil || !strings.Contains(err.Error(), c.says) || !strings.Contains(err.Error(), "conflict") && !strings.Contains(err.Error(), "already names") {
				t.Errorf("err = %v; want a conflict saying %q", err, c.says)
			}
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION); n != events {
				t.Errorf("the refused retry recorded %d computation(s)", n-events)
			}
			if ran() != 2 {
				t.Errorf("the refused retry ran the program: %d runs", ran())
			}
		})
	}
}
