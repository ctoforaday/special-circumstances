package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE RUN'S CONFIGURATION IS COPIED AS ONE OBJECT OR REFUSED. The record reads the file strictly, so
// a copy that read past a truncation, a second object or trailing bytes would publish a run whose
// configuration does not read; and the dropped keys go while every other key keeps its place.
func TestTranslateRunConfig(t *testing.T) {
	dir := t.TempDir()
	from, to := filepath.Join(dir, "from.json"), filepath.Join(dir, "out", "to.json")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(from, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// THE EPOCH THE FILE STATES IS THE ONE THE MIGRATED RECORD IS AT: rewritten in its place where
	// the source states one, added at the end where it states none. WHAT WAS CHANGED IS RETURNED FOR
	// THE MANIFEST with the values the source wrote: a source record may state no epoch of its own,
	// and then the file's is the only statement of it.
	for _, tc := range []struct{ name, body, want, translation string }{
		{"states its epoch", `{"topic":"t","eventSchema":5,"mintBudget":3,"k":2}` + "\n",
			fmt.Sprintf("{\n  \"topic\": \"t\",\n  \"eventSchema\": %d,\n  \"k\": 2\n}\n", record.EventSchema),
			`{"dropped":{"mintBudget":3},"source_event_schema":5}`},
		{"states none", `{"topic":"t","mintBudget":3,"k":2}` + "\n",
			fmt.Sprintf("{\n  \"topic\": \"t\",\n  \"k\": 2,\n  \"eventSchema\": %d\n}\n", record.EventSchema),
			`{"dropped":{"mintBudget":3}}`},
		{"drops the call limit", `{"topic":"t","eventSchema":73,"maxSittingCalls":150}` + "\n",
			fmt.Sprintf("{\n  \"topic\": \"t\",\n  \"eventSchema\": %d\n}\n", record.EventSchema),
			`{"dropped":{"maxSittingCalls":150},"source_event_schema":73}`},
		{"drops nothing", `{"topic":"t","eventSchema":73}` + "\n",
			fmt.Sprintf("{\n  \"topic\": \"t\",\n  \"eventSchema\": %d\n}\n", record.EventSchema),
			`{"dropped":{},"source_event_schema":73}`},
	} {
		write(tc.body)
		tr, err := translateRunConfig(from, to)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(to)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != tc.want {
			t.Errorf("%s: translated run-config =\n%s\nwant\n%s", tc.name, got, tc.want)
		}
		if b, err := json.Marshal(tr); err != nil || string(b) != tc.translation {
			t.Errorf("%s: the translation reads %s (%v), want %s", tc.name, b, err, tc.translation)
		}
	}
	for _, tc := range []struct{ name, body, want string }{
		{"truncated", `{"topic":"t"`, "unexpected EOF"},
		{"truncated before its close", `{"topic":"t",`, "unexpected EOF"},
		{"a second object", `{"k":2}{"k":1}`, "content after the run's configuration"},
		{"trailing bytes", `{"k":2} garbage`, "content after the run's configuration"},
		{"not an object", `["k"]`, "it holds an array"},
		{"null", `null`, "it holds null"},
		{"a number", ` 3 `, "it holds a number"},
		{"empty", ``, "does not parse as the run's configuration: EOF"},
		{"a key the configuration does not declare", `{"k":2,"rounds":3}`, "unknown field"},
		{"lanes as a number", `{"lanes":3}`, "the migrated configuration does not parse"},
	} {
		write(tc.body)
		_, err := translateRunConfig(from, to)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s (%q): %v, want an error naming %q", tc.name, tc.body, err, tc.want)
		}
		if err != nil && strings.Contains(err.Error(), "<nil>") {
			t.Errorf("%s: the refusal reads %q", tc.name, err)
		}
	}
}
