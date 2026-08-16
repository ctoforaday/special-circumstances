package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TWO SITTINGS THE RECORD CANNOT SEPARATE MUST RESOLVE THE SAME WAY EVERY TIME.
//
// MEASURED 2026-08-16: TestDiscardedEventsAudit passed on ubuntu-latest and FAILED on
// windows-latest at the same commit. Two shards written back to back land on distinct mtimes under
// Linux and on IDENTICAL ones under Windows' coarser timestamp resolution, and the selection was
// `s.mtime.After(winner.mtime)` — strictly after — so a tie left pool[0] standing and the surviving
// sitting was whichever shard os.ReadDir yielded first.
//
// THE PLATFORM IS NOT THE POINT. Equal mtimes are reachable anywhere; Windows just reaches them
// reliably. This test manufactures the condition with Chtimes so the property is checked on every
// platform, rather than trusting a Linux run that cannot produce the tie to say anything about it.
func TestAnMtimeTieResolvesDeterministically(t *testing.T) {
	shared := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	build := func(t *testing.T, order ...string) (string, string) {
		t.Helper()
		runDir := t.TempDir()
		t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
		dir := filepath.Join(runDir, "records")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, nonce := range order {
			var lines []string
			for _, e := range []Event{
				{TS: "2026-01-01T00:00:00Z", SeatID: "judge-petition", Nonce: nonce, Type: "register", Key: "judge-petition:register:" + nonce, Payload: NewPayload()},
				{TS: "2026-01-01T00:00:00Z", SeatID: "judge-petition", Nonce: nonce, Type: "petition-rule", Key: "judge-petition:petition-rule:" + nonce, Payload: NewPayload()},
			} {
				b, err := MarshalEvent(e)
				if err != nil {
					t.Fatal(err)
				}
				lines = append(lines, string(b))
			}
			p := filepath.Join(dir, "events-judge-petition-"+nonce+".jsonl")
			if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			// THE TIE, MANUFACTURED. Both shards claim the same instant, which is what Windows
			// hands the replay for free.
			if err := os.Chtimes(p, shared, shared); err != nil {
				t.Fatal(err)
			}
		}
		m, err := MergedEvents(runDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Discarded) != 1 {
			t.Fatalf("expected one seat to have discarded work, got %d", len(m.Discarded))
		}
		return m.Discarded[0].Winner, strings.Join(m.Anomalies, " | ")
	}

	// THE WINNER IS THE STATED TIE-BREAK, not an accident of enumeration.
	//
	// NOT an order-invariance check, and the distinction is worth writing down: os.ReadDir returns
	// entries sorted by name, so write order never reached the selection and an assertion about it
	// could not fail. What actually varied across platforms was whether the CLOCK could separate
	// the two writes — Linux gave distinct mtimes and picked the later shard, Windows gave equal
	// ones and kept whichever sorted first. Pinning the tie-break is what makes both platforms
	// answer the same way.
	winnerAB, note := build(t, "aaaaaaaa", "bbbbbbbb")
	winnerBA, _ := build(t, "bbbbbbbb", "aaaaaaaa")
	if winnerAB != winnerBA {
		t.Errorf("the same two sittings resolved differently: %q vs %q", winnerAB, winnerBA)
	}
	if winnerAB != "bbbbbbbb" {
		t.Errorf("winner = %q, want \"bbbbbbbb\" — on a tie the greater nonce wins, which is arbitrary but stable, and it is what keeps a tied record resolving the same way on every filesystem", winnerAB)
	}

	if !strings.Contains(note, "TIED") {
		t.Errorf("the anomaly does not disclose that the mtimes tied, so it reports a decision the clock never made:\n%s", note)
	}
}
