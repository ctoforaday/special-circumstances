package difftest

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
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
//	go test ./internal/difftest -run TestGolden -update

var update = flag.Bool("update", false, "regenerate golden files (a deliberate semantics change)")

func TestGolden(t *testing.T) {
	bin := buildBinary(t)
	for _, sc := range scenarios() {
		t.Run(sc.name, func(t *testing.T) {
			runDir := t.TempDir()
			seed(t, runDir, sc.seed)
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
			transcript.WriteString("═══ EVENTS ═══\n")
			for _, name := range sortedKeys(st.events) {
				fmt.Fprintf(&transcript, "-- %s\n", name)
				for _, ev := range st.events[name] {
					b, _ := json.Marshal(ev)
					transcript.Write(b)
					transcript.WriteString("\n")
				}
			}
			transcript.WriteString("\n═══ RENDERS ═══\n")
			for _, name := range sortedKeys(st.renders) {
				fmt.Fprintf(&transcript, "-- %s\n%s\n", name, st.renders[name])
			}

			compareGolden(t, sc.name, transcript.String())
		})
	}
}

func compareGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
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
	if got != want {
		t.Errorf("%s differs from its golden.\nIf this change is INTENTIONAL, regenerate with -update and justify it in the commit.\n%s",
			name, firstDiff(got, want))
	}
}

// firstDiff reports the first differing line with context, rather than dumping
// two multi-kilobyte transcripts into the failure output.
func firstDiff(got, want string) string {
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		gl, wl := "", ""
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			start := max(0, i-3)
			var b strings.Builder
			fmt.Fprintf(&b, "first difference at line %d:\n", i+1)
			for j := start; j < i; j++ {
				fmt.Fprintf(&b, "  %s\n", w[j])
			}
			fmt.Fprintf(&b, "- want: %s\n+ got:  %s\n", wl, gl)
			return b.String()
		}
	}
	return "(transcripts differ only in trailing whitespace)"
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
