package cli

import (
	"encoding/json"
	"testing"
)

// THE RECORDED VERDICT REACHES THE DEBATE THE BENCH READS. The structured debate carries each
// epoch's verdict beside the chair's prose because the prose can say PASS while the record holds
// fail; an empty verdict means "none recorded", so a projection that never loaded the gates prints
// that on every epoch of every run, in the bytes it would use for an epoch nobody judged (#1243).
//
// Driven through the real verb on a run directory — the bench's `inquest debate --json` — because
// the run path is where the projection narrows its read; the whole-stream callers (capture, the
// dashboard) always carried the gates.
func TestInquestDebateCarriesEachEpochsRecordedVerdict(t *testing.T) {
	runDir := newRun(t)
	chair := func(args ...string) {
		t.Helper()
		if _, err := run(t, append(args, "--run", runDir, "--seat-id", "red-chair")...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	// Epoch 1 fails; the chair sits again and epoch 2 passes.
	chair("register")
	chair("verdict", "--as", "FAIL")
	chair("register")
	chair("verdict", "--as", "PASS")

	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "docket"); err != nil {
		t.Fatalf("register judge: %v", err)
	}
	out, err := run(t, "inquest", "debate", "--json", "--run", runDir, "--seat-id", "judge")
	if err != nil {
		t.Fatalf("inquest debate --json: %v\n%s", err, out)
	}
	var dj struct {
		Epochs []struct {
			Epoch   int    `json:"epoch"`
			Verdict string `json:"verdict"`
		} `json:"epochs"`
	}
	if err := json.Unmarshal([]byte(out), &dj); err != nil {
		t.Fatalf("inquest debate --json emitted unparseable JSON: %v\n%s", err, out)
	}
	got := map[int]string{}
	for _, e := range dj.Epochs {
		got[e.Epoch] = e.Verdict
	}
	for epoch, want := range map[int]string{1: "fail", 2: "pass"} {
		if got[epoch] != want {
			t.Errorf("epoch %d: inquest debate reports verdict %q, the record holds %q\n%s", epoch, got[epoch], want, out)
		}
	}
}
