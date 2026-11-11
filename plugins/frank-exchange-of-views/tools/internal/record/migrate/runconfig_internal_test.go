package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	write(`{"topic":"t","mintBudget":3,"k":2}` + "\n")
	if err := translateRunConfig(from, to); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(to)
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n  \"topic\": \"t\",\n  \"k\": 2\n}\n"; string(got) != want {
		t.Errorf("translated run-config =\n%s\nwant\n%s", got, want)
	}
	for _, tc := range []struct{ name, body, want string }{
		{"truncated", `{"topic":"t"`, "unexpected end of JSON input"},
		{"truncated before its close", `{"topic":"t",`, "unexpected end of JSON input"},
		{"a second object", `{"k":2}{"k":1}`, "content after its object"},
		{"trailing bytes", `{"k":2} garbage`, "content after its object"},
		{"not an object", `["k"]`, "starts with ["},
		{"empty", ``, "is not a JSON object: EOF"},
	} {
		write(tc.body)
		err := translateRunConfig(from, to)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s (%q): %v, want an error naming %q", tc.name, tc.body, err, tc.want)
		}
		if err != nil && strings.Contains(err.Error(), "<nil>") {
			t.Errorf("%s: the refusal reads %q", tc.name, err)
		}
	}
}
