package difftest

import (
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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

func TestGolden(t *testing.T) {
	bin := buildBinary(t)
	for _, sc := range scenarios() {
		t.Run(sc.name, func(t *testing.T) {
			runDir := t.TempDir()
			// EVERY SCENARIO DECLARES ITS CLASS VOCABULARY, as a real run does. These runs are
			// built by hand, and used to be exempt from the class check by accident: no registry
			// staged meant `--class` accepted any string, so every mint in every golden here
			// passed a check that was not running. Staging it is what makes the golden a record
			// of the shipped behaviour rather than of an unconfigured corner of it.
			if err := record.StageForRun(runtest.Open(t, runDir), goldenClasses...); err != nil {
				t.Fatalf("stage the class registry: %v", err)
			}
			seed(t, runDir, sc.seed)
			// A lens finding anchors into blue/report.md (slice 1b) and is rejected
			// unless its --quote quote is present. Ensure a report carrying the
			// scenario heading anchors (## S2 / ## S4) exists; a scenario may override.
			if _, ok := sc.seed["blue/report.md"]; !ok {
				seed(t, runDir, map[string]string{
					"blue/report.md": "## S2\n\nA claim sits under S2.\n\n## S4\n\nA claim sits under S4.\n",
				})
			}
			m := newMapper()

			// INGEST THE ROUND-0 REPORT (#709). The report is the record projection now, so the
			// seeded blue/report.md must be ingested before any verb reads or mutates it: blue-
			// synthesize records the base and the file is deleted, exactly as the engine does after
			// synthesis and before the rounds. Invisible setup, like the class-registry staging and
			// the report seed above — the scenario's own commands are what the golden is about, but
			// the BaseIngest and this register DO appear in the EVENTS section, as they must.
			for _, setup := range []cmd{
				{role: "register", args: []string{"--run", "{RUN}", "--seat-id", "blue-synthesize"}},
				{role: "ingest", args: []string{"--run", "{RUN}", "--seat-id", "blue-synthesize"}},
			} {
				if inv := runGo(bin, runDir, setup); inv.code != 0 {
					t.Fatalf("round-0 setup %v: exit %d\nstderr: %s", setup.args, inv.code, inv.stderr)
				}
				m.observe(filepath.Join(runDir, "records"))
			}

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
			// THE CLOCK IS REPLACED BY POSITION, and position now comes from the record rather
			// than from a sort this harness performs. The events arrive in `events.id` order —
			// one sequence the store assigns at the write — so ranking them is reading them.
			// The old key was (TS, SeatID, Seq) across shard files, and it had to be
			// reconstructed because a run's events lived in several files that each knew only
			// their own order.
			transcript.WriteString("═══ EVENTS ═══\n")
			for i, ev := range st.events {
				if _, ok := ev["ts"]; ok {
					ev["ts"] = i
				}
				b, _ := json.Marshal(ev)
				transcript.Write(b)
				transcript.WriteString("\n")
			}

			// RENDERS: the markdown projections, pulled through the SAME path a seat uses —
			// `show <v>`, which renders in-memory from the record via internal/view (no
			// render-shadow). This byte-pins every projection across every scenario, the coverage
			// the render-shadow removal (#203) dropped when it deleted the materialized snapshot.
			// A view that errors on a given run (e.g. a pure-help/error scenario with no record)
			// contributes nothing, so degenerate runs carry no RENDERS section — as before.
			var renders strings.Builder
			for _, v := range []string{"ledger", "archive", "debate", "changelog", "citation-ledger", "lines-of-inquiry"} {
				// THE SEAT SELECTS THE TREE, so the projection is read the way a merge seat reads
				// it: `show <v> --seat-id red-merge-r1`. This said `cmd{role: "merge", args:
				// {"show", ...}}`, which composes `merge show ledger` — a path that stopped
				// existing when the surface became seat-scoped, and no --seat-id at all, so the
				// root exposed no verbs to begin with.
				//
				// IT FAILED SILENTLY BY DESIGN. The block only appends when `got.code == 0`,
				// because a degenerate run legitimately renders nothing — so every scenario simply
				// lost its RENDERS and REPORT halves, and regenerating would have recorded that as
				// the new expected output. 2278 deletions against 323 insertions across 20 files,
				// and every one of them would have been "the goldens moved with the surface".
				got := normalizeOutput(runGo(bin, runDir, cmd{role: "show", args: []string{v, "--run", runDir, "--seat-id", "red-merge-r1"}}), runDir, m)
				if got.code == 0 {
					fmt.Fprintf(&renders, "-- %s\n%s\n", v, got.stdout)
				}
			}
			if renders.Len() > 0 {
				transcript.WriteString("\n═══ RENDERS ═══\n")
				transcript.WriteString(renders.String())
			}

			// REPORT: the artifact a HUMAN reads, which until now had no golden at all (#447).
			//
			// The RENDERS above pin the projections a SEAT reads back mid-run. `bench assemble`
			// composes something different — the end-of-run document the person the whole run is
			// for actually opens — and nothing byte-pinned it. Measured: renaming the `## Record
			// verification` heading added by #437 left `internal/report`, `internal/difftest` and
			// the whole golden suite GREEN. Its unit test asserts `strings.HasPrefix(got, "##
			// Record verification")`, which any suffix satisfies, so the miss and the pass were
			// the same bytes — `facts-are-fields` clause 3, in a test rather than in a parser.
			//
			// Read from DISK rather than stdout: assemble prints a confirmation line and writes
			// the document to <run>/report.md. A scenario with no terminal outcome or no audited
			// blue/report.md cannot assemble, so it contributes no section — the same
			// degenerate-run rule RENDERS already follows, and the reason this could be added to
			// every scenario rather than needing a fixture of its own.
			if inv := runGo(bin, runDir, cmd{role: "assemble", args: []string{"--run", runDir, "--seat-id", "judge-terminal"}}); inv.code == 0 {
				body, err := os.ReadFile(filepath.Join(runDir, "report.md"))
				if err != nil {
					t.Fatalf("bench assemble exited 0 but wrote no report.md: %v", err)
				}
				transcript.WriteString("\n═══ REPORT ═══\n")
				transcript.WriteString(normalizeOutput(invocation{stdout: string(body)}, runDir, m).stdout)
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
		t.Fatalf("missing golden %s (regenerate with UPDATE_GOLDENS=1): %v", path, err)
	}
	want := strings.ReplaceAll(string(wantBytes), "\r\n", "\n")
	if diff := cmp.Diff(want, got); diff != "" {
		// `UPDATE_GOLDENS=1`, NOT `-update`. The comment at the top of this file exists because
		// someone already wrote `-update` there and no such flag is declared — and this message,
		// the one a reader actually hits, went on saying it anyway. Instructions that name a
		// mechanism nobody implements are the same defect as a refusal naming a flag no verb
		// registers; this one had its own correction written thirty lines above it.
		t.Errorf("%s differs from its golden (-want +got).\nIf this change is INTENTIONAL, regenerate with UPDATE_GOLDENS=1 go test ./internal/difftest and justify the diff in the commit.\n%s",
			name, diff)
	}
}

// goldenClasses is the vocabulary the scenarios mint under. It deliberately EXCLUDES the slugs
// scenarios use to prove an unknown class is refused (`invented`, `invented-class`, `novel`,
// `banana`) — those are the point of their scenarios, and adding them here would turn four
// refusal goldens green while recording nothing.
//
// A scenario that stages its OWN registry still does: seed() runs after this and overwrites the
// file, which is how the class_registry oracle keeps driving a hand-made registry.
var goldenClasses = []string{
	"scope-creep", "attestation-inflation", "citation-drift", "safety", "x", "a",
}
