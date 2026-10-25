package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A KEY IS ONE ACT'S NAME. A retry under it that states the act again is answered with the stored
// act; one that states anything else is refused as a conflict naming what the record holds. The
// quote of a placing act is not compared: a retry is finished at the location the act stored.
const (
	keyedStored = "The measured latency is bounded."
	keyedOther  = "The queue drains in order."
)

// keyedRun is a run whose report holds the two sentences, with a chair, a lens and blue seated.
func keyedRun(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\n"+keyedStored+"\n\n## J\n\n"+keyedOther+"\n")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	registerBlue(t, runDir)
	return runDir
}

// keyedRetry is one keyed verb: its first call, the retries that state the act otherwise (each with
// the stored statement its refusal must name), and the retries that differ only in what the act
// does not compare.
type keyedRetry struct {
	verb, seat string
	typ        recordpb.EventType
	idField    string
	args       []string            // the first call, after --run and --seat-id
	differ     map[string][]string // flag -> replacement pair, applied to args
	names      map[string]string   // flag -> the stored statement the refusal names
	located    [][]string          // replacement pairs that change only what is not compared
	restated   map[string][]string // whole retries (after --run and --seat-id) -> the stored statement named
	gaps       bool                // mint two gaps and file two findings first: {G1} {G2} {F1} {F2} in the calls
}

func TestAKeyedRetryIsHeldToTheActItRepeats(t *testing.T) {
	for _, c := range []keyedRetry{
		{verb: "finding", seat: lensSeat, typ: recordpb.EventType_EVENT_TYPE_FINDING, idField: "finding_id",
			args: []string{"--key", "K1", "--quote", keyedStored, "--reason", "the bound is unsourced",
				"--severity", "low", "--likelihood", "low", "--impact", "low"},
			differ: map[string][]string{
				"--quote":      {"--quote", keyedOther},
				"--reason":     {"--reason", "another claim"},
				"--severity":   {"--severity", "high"},
				"--likelihood": {"--likelihood", "high"},
				"--impact":     {"--impact", "high"},
			},
			names: map[string]string{
				"--quote":  `--quote "` + keyedStored + `"`,
				"--reason": `--reason "the bound is unsourced"`, "--severity": "--severity low",
				"--likelihood": "--likelihood low", "--impact": "--impact low",
			}},
		{verb: "finding", seat: lensSeat, typ: recordpb.EventType_EVENT_TYPE_FINDING, idField: "finding_id",
			args: []string{"--key", "K2", "--about-kind", "section", "--about", "H", "--reason", "a case is missing",
				"--severity", "low", "--likelihood", "low", "--impact", "low"},
			differ: map[string][]string{"--about": {"--about", "J"}},
			names:  map[string]string{"--about": `--about "H"`}},
		{verb: "mint", seat: lensSeat, typ: recordpb.EventType_EVENT_TYPE_MINT, idField: "gap_id",
			args: []string{"--key", "K1", "--quote", keyedStored, "--class", "spec_underspecification",
				"--problem", "the bound is unsourced", "--fix", "source the bound", "--check-kind", "document",
				"--check", "the bound cites a source", "--severity", "medium", "--likelihood", "medium",
				"--impact", "medium", "--complexity", "low", "--reason", "red's argument"},
			differ: map[string][]string{
				"--problem":    {"--problem", "another defect"},
				"--fix":        {"--fix", "another fix"},
				"--check":      {"--check", "another check"},
				"--check-kind": {"--check-kind", "computation"},
				"--severity":   {"--severity", "high"},
				"--likelihood": {"--likelihood", "high"},
				"--impact":     {"--impact", "high"},
				"--complexity": {"--complexity", "high"},
				"--class":      {"--class", "unsourced_claim"},
				"--quote":      {"--quote", keyedOther},
				"--reason":     {"--reason", "the argument, put again"},
			},
			names: map[string]string{
				"--problem": `--problem "the bound is unsourced"`, "--fix": `--fix "source the bound"`,
				"--check": `--check "the bound cites a source"`, "--check-kind": "--check-kind document",
				"--severity": "--severity medium", "--likelihood": "--likelihood medium", "--impact": "--impact medium",
				"--complexity": "--complexity low",
				"--class":      "--class spec_underspecification",
				"--quote":      `--quote "` + keyedStored + `"`,
				"--reason":     `--reason "red's argument"`,
			}},
		{verb: "mint", seat: lensSeat, typ: recordpb.EventType_EVENT_TYPE_MINT, idField: "gap_id", gaps: true,
			args: []string{"--key", "K2", "--quote", keyedStored, "--new", "The measured latency is capped.",
				"--class", "spec_underspecification", "--problem", "the word is loose", "--fix", "tighten it",
				"--check-kind", "document", "--check", "the word is exact", "--severity", "low", "--likelihood", "low",
				"--impact", "low", "--complexity", "low", "--supersedes", "{G1}", "--found-by", "{F1}"},
			differ: map[string][]string{
				"--new":        {"--new", "The measured latency is limited."},
				"--supersedes": {"--supersedes", "{G2}"},
				"--found-by":   {"--found-by", "{F2}"},
			},
			names: map[string]string{"--new": `--new "The measured latency is capped."`, "--supersedes": "--supersedes {G1}",
				"--found-by": "--found-by {F1}"}},
		{verb: "mint", seat: lensSeat, typ: recordpb.EventType_EVENT_TYPE_MINT, idField: "gap_id",
			args: []string{"--key", "K3", "--about-kind", "section", "--about", "H", "--class", "spec_underspecification",
				"--problem", "a case is missing", "--fix", "add it", "--check-kind", "document", "--check", "the case is there",
				"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low"},
			differ: map[string][]string{"--about": {"--about", "J"}},
			names:  map[string]string{"--about": `--about "H"`}},
		{verb: "edit", seat: blueSeat, typ: recordpb.EventType_EVENT_TYPE_BLUE_EDIT,
			args: []string{"--key", "E1", "--quote", "is bounded", "--new", "is capped", "--reason", "plainer"},
			differ: map[string][]string{
				"--new":    {"--new", "is limited"},
				"--quote":  {"--quote", "drains in order"},
				"--reason": {"--reason", "plainer, put again"},
			},
			names: map[string]string{"--new": `--new "is capped"`, "--quote": `--quote "is bounded"`, "--reason": `--reason "plainer"`}},
		{verb: "edit", seat: blueSeat, typ: recordpb.EventType_EVENT_TYPE_BLUE_EDIT, gaps: true,
			args:   []string{"--key", "E2", "--quote", "drains in order", "--new", "drains first in, first out", "--answers", "{G1}", "--reason", "exact"},
			differ: map[string][]string{"--answers": {"--answers", "{G2}"}},
			names:  map[string]string{"--answers": `--answers "{G1}"`},
			restated: map[string][]string{
				"no --accept": {"--key", "E2", "--answers", "{G1}", "--accept", "--reason", "exact"},
			}},
	} {
		t.Run(c.verb+" "+c.args[1], func(t *testing.T) {
			runDir := keyedRun(t)
			gap := strings.NewReplacer()
			if c.gaps {
				g1, err := mintQuote(t, runDir, "g1", keyedStored)
				if err != nil {
					t.Fatal(err)
				}
				g2, err := mintQuote(t, runDir, "g2", keyedOther)
				if err != nil {
					t.Fatal(err)
				}
				var f [2]string
				for i, q := range []string{keyedStored, keyedOther} {
					out, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "f"+q, "--quote", q,
						"--reason", "r", "--severity", "low", "--likelihood", "low", "--impact", "low", "--json")
					if err != nil {
						t.Fatal(err)
					}
					var env struct {
						Result struct {
							ID string `json:"finding_id"`
						} `json:"result"`
					}
					if err := json.Unmarshal([]byte(out), &env); err != nil {
						t.Fatal(err)
					}
					f[i] = env.Result.ID
				}
				gap = strings.NewReplacer("{G1}", g1, "{G2}", g2, "{F1}", f[0], "{F2}", f[1])
			}
			call := func(args []string) (string, error) {
				full := []string{c.verb, "--run", runDir, "--seat-id", c.seat, "--json"}
				for _, a := range args {
					full = append(full, gap.Replace(a))
				}
				return run(t, full...)
			}
			idOf := func(out string) string {
				var env struct {
					Result map[string]any `json:"result"`
				}
				if err := json.Unmarshal([]byte(out), &env); err != nil {
					t.Fatalf("read the result: %v\n%s", err, out)
				}
				if env.Result["idempotent"] == true {
					return "idempotent:" + str(env.Result[c.idField])
				}
				return str(env.Result[c.idField])
			}
			out, err := call(c.args)
			if err != nil {
				t.Fatalf("the first %s was refused: %v", c.verb, err)
			}
			first := idOf(out)

			// An identical retry, and one differing only in what is not compared, is the stored act.
			for _, swap := range append([][]string{nil}, c.located...) {
				out, err := call(replaced(c.args, swap))
				if err != nil {
					t.Fatalf("a retry changing %v was refused: %v", swap, err)
				}
				if got := idOf(out); got != "idempotent:"+first {
					t.Errorf("a retry changing %v answered %q; want the stored %s %q, idempotent", swap, got, c.verb, first)
				}
			}
			// A retry stating the act otherwise is refused, naming the stored statement.
			for flag, swap := range c.differ {
				_, err := call(replaced(c.args, swap))
				if want := gap.Replace(c.names[flag]); err == nil || feov.CodeOf(err) != string(feov.Conflict) || !strings.Contains(err.Error(), want) {
					t.Errorf("a retry with another %s answered %v; want a conflict naming %s", flag, err, want)
				}
			}
			for named, args := range c.restated {
				if _, err := call(args); err == nil || feov.CodeOf(err) != string(feov.Conflict) || !strings.Contains(err.Error(), named) {
					t.Errorf("the retry %v answered %v; want a conflict naming %s", args, err, named)
				}
			}
			want := 1
			if c.gaps && c.typ == recordpb.EventType_EVENT_TYPE_MINT {
				want = 3
			}
			if n := countType(t, runDir, c.typ); n != want {
				t.Errorf("the record holds %d %s events, want %d", n, c.typ, want)
			}
		})
	}
}

// replaced is args with the value after swap[0] set to swap[1]; nil swap is args unchanged.
func replaced(args, swap []string) []string {
	out := append([]string(nil), args...)
	if swap == nil {
		return out
	}
	for i := 0; i+1 < len(out); i++ {
		if out[i] == swap[0] {
			out[i+1] = swap[1]
			return out
		}
	}
	panic("no " + swap[0] + " in the call")
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

// AN EDIT'S STORED REPLACEMENT CARRIES THE ANCHOR THE TOOL PUT BACK, which the seat never typed. Its
// identical retry is the stored edit all the same.
func TestAnEditRetryIsComparedWithoutTheAnchorsPutBack(t *testing.T) {
	runDir := keyedRun(t)
	out, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--quote", keyedStored,
		"--reason", "the bound is unsourced", "--severity", "low", "--likelihood", "low", "--impact", "low", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Result struct {
			ID string `json:"finding_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.ID == "" {
		t.Fatalf("no finding id: %v\n%s", err, out)
	}
	// The quote carries the finding's anchor and the replacement leaves it out, keeping its sentence:
	// the tool puts the anchor back.
	edit := []string{"edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
		"--quote", "The measured latency is bounded" + anchor.Token(env.Result.ID),
		"--new", "Under load. The measured latency is bounded", "--reason", "context first"}
	if _, err := run(t, edit...); err != nil {
		t.Fatalf("the edit was refused: %v", err)
	}
	if stored := lastBody(t, runDir, &recordpb.BlueEdit{}).GetNew(); !strings.Contains(stored, env.Result.ID) {
		t.Fatalf("the edit stored %q: no anchor was put back, so this pins nothing", stored)
	}
	if out, err := run(t, edit...); err != nil || !strings.Contains(out, "idempotent") {
		t.Errorf("the identical retry answered %q, %v; want the stored edit", out, err)
	}
}
