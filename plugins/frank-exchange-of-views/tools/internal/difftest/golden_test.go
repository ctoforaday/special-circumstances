package difftest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Golden tests are what SURVIVES the oracle.
//
// The differential gate proved the port faithful; it cannot outlive the mjs
// implementation it compares against, and that implementation is being retired
// (it was never used in a run — no run directory contains records/, and the
// engine's toolsDir defaulted to null throughout). But the scenarios ARE the
// executable statement of these semantics, and deleting the oracle without first
// capturing them would throw away the specification along with the thing being
// specified.
//
// So: these goldens were generated while the gate was still green, which means
// every byte in testdata/ is oracle-validated behaviour. From here they are the
// contract — a diff in a golden is a deliberate semantics change, and must be
// justified in the commit that regenerates it.
//
//	go run ./golden -review          (from scripts/ — verify, review each diff, accept what you read)
//
// NOT a -update FLAG: this line used to read `go test ./internal/difftest -run TestGolden
// -update`, and no such flag is declared here — see the env-var note below. A documented
// command that cannot run is worse than none: it sends the reader looking for the mechanism
// somewhere other than where it lives.

// Updating is driven by an ENV VAR rather than a test flag so that one command
// can drive both languages: a Go test flag is package-scoped (passing -update to
// a package that does not declare it is a hard error), while UPDATE_GOLDENS=1
// reaches every suite in the repo, Go and mjs alike. scripts/golden.mjs relies
// on that.
var update = os.Getenv("UPDATE_GOLDENS") == "1"

// eventRanks maps each event to its position in the canonical merge order, counting from
// zero, keyed by the event's identity (shard file + seq).
//
// RANKED BY POSITION, NOT BY DISTINCT TIMESTAMP. The first version ranked the set of
// distinct clock values, which was reproducible on one machine and NOT on another: when
// two events land inside the same clock tick there is one fewer distinct instant, so every
// later rank shifts by one and the whole golden moves. CI found it immediately — `ts:9`
// locally against `ts:8` on the runner — and the cause was speed, not semantics. Ranking
// distinct instants fixed the machine-PATH dependence these goldens used to have and left
// a machine-SPEED dependence in its place.
//
// The canonical order is (TS, SeatID, Seq), the same key BoardState replays by. Ranking
// position in THAT order is stable no matter how coarse the clock is: events sharing a
// tick are separated by seat and sequence, deterministically. A real reordering still
// changes the golden, which is the property this field was added to protect after
// replay-by-filename silently dropped the bench's closures.
func eventRanks(events map[string][]map[string]any) map[string]int {
	type ref struct {
		ts, seat string
		seq      float64
		id       string
	}
	var all []ref
	for name, evs := range events {
		for _, ev := range evs {
			r := ref{id: name}
			r.ts, _ = ev["ts"].(string)
			r.seat, _ = ev["seatId"].(string)
			r.seq, _ = ev["seq"].(float64)
			r.id = fmt.Sprintf("%s#%v", name, ev["seq"])
			all = append(all, r)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		a, b := all[i], all[j]
		if a.ts != b.ts {
			return a.ts < b.ts
		}
		if a.seat != b.seat {
			return a.seat < b.seat
		}
		return a.seq < b.seq
	})
	rank := make(map[string]int, len(all))
	for i, r := range all {
		rank[r.id] = i
	}
	return rank
}

func TestGolden(t *testing.T) {
	bin := buildBinary(t)
	for _, sc := range scenarios() {
		t.Run(sc.name, func(t *testing.T) {
			runDir := t.TempDir()
			seed(t, runDir, sc.seed)
			// A lens finding anchors into blue/report.md (slice 1b) and is rejected
			// unless its --location quote is present. Ensure a report carrying the
			// scenario heading anchors (## S2 / ## S4) exists; a scenario may override.
			if _, ok := sc.seed["blue/report.md"]; !ok {
				seed(t, runDir, map[string]string{
					"blue/report.md": "## S2\n\nA claim sits under S2.\n\n## S4\n\nA claim sits under S4.\n",
				})
			}
			m := newMapper()
			var transcript strings.Builder

			for _, c := range sc.cmds {
				inv := runGo(bin, runDir, c)
				m.observe(filepath.Join(runDir, "records"))
				applyMtimes(t, runDir, m, c)
				got := normalizeOutput(inv, runDir, m)
				fmt.Fprintf(&transcript, "$ %s %s\nexit %d\n", c.role, strings.Join(c.args, " "), got.code)
				if got.stdout != "" {
					fmt.Fprintf(&transcript, "stdout:\n%s", got.stdout)
				}
				if got.stderr != "" {
					fmt.Fprintf(&transcript, "stderr:\n%s", got.stderr)
				}
				transcript.WriteString("\n")
			}

			st := collect(t, runDir, m)

			// TIMESTAMPS BECOME THEIR RANK, not a placeholder.
			//
			// Every event carries the wall clock, which differs on every run, so a raw
			// value makes the golden unreproducible — the machine-dependence that once
			// baked a developer's home directory into these files. Blanking it to a
			// constant would fix that and throw away the ORDER, which is precisely what
			// this field was added to establish after replay-by-filename silently
			// dropped the bench's closures. Ranking keeps it: an event's POSITION in
			// the canonical (TS, SeatID, Seq) order is written in place of its clock,
			// so two events swapping places CHANGES the golden and a reordering cannot
			// slip past as "just the clock".
			//
			// Normalized here rather than faked in the binary: a production code path
			// whose only purpose is to lie about the clock is the worse trade.
			rank := eventRanks(st.events)

			transcript.WriteString("═══ EVENTS ═══\n")
			for _, name := range sortedKeys(st.events) {
				fmt.Fprintf(&transcript, "-- %s\n", name)
				for _, ev := range st.events[name] {
					if _, ok := ev["ts"]; ok {
						ev["ts"] = rank[fmt.Sprintf("%s#%v", name, ev["seq"])]
					}
					b, _ := json.Marshal(ev)
					transcript.Write(b)
					transcript.WriteString("\n")
				}
			}

			// RENDERS: the markdown projections, pulled through the SAME path a seat uses —
			// `show <v>`, which renders in-memory from the record via internal/view (no
			// render-shadow). This byte-pins every projection across every scenario, the coverage
			// the render-shadow removal (#203) dropped when it deleted the materialized snapshot.
			// A view that errors on a given run (e.g. a pure-help/error scenario with no record)
			// contributes nothing, so degenerate runs carry no RENDERS section — as before.
			var renders strings.Builder
			for _, v := range []string{"ledger", "archive", "debate", "changelog", "citation-ledger", "lines-of-inquiry"} {
				got := normalizeOutput(runGo(bin, runDir, cmd{role: "merge", args: []string{"show", v, "--run", runDir}}), runDir, m)
				if got.code == 0 {
					fmt.Fprintf(&renders, "-- %s\n%s\n", v, got.stdout)
				}
			}
			if renders.Len() > 0 {
				transcript.WriteString("\n═══ RENDERS ═══\n")
				transcript.WriteString(renders.String())
			}
			compareGolden(t, sc.name, transcript.String())
		})
	}
}

func compareGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("regenerated %s", path)
		return
	}
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden %s (run with -update): %v", path, err)
	}
	want := strings.ReplaceAll(string(wantBytes), "\r\n", "\n")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("%s differs from its golden (-want +got).\nIf this change is INTENTIONAL, regenerate with -update and justify it in the commit.\n%s",
			name, diff)
	}
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
