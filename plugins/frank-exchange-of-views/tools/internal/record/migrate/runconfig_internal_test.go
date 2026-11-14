package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
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
		{"drops the substitution consent", `{"topic":"t","eventSchema":74,"allowModelSubstitution":true,"k":2}` + "\n",
			fmt.Sprintf("{\n  \"topic\": \"t\",\n  \"eventSchema\": %d,\n  \"k\": 2\n}\n", record.EventSchema),
			`{"dropped":{"allowModelSubstitution":true},"source_event_schema":74}`},
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
		{"both the class tiers and the seats", `{"model":"sonnet","seats":{}}`, "states both the per-seat"},
		{"a tier that is not a string", `{"model":3}`, `key "model" states 3, not a model`},
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

// AN ARCHIVED CLASS-TIER PAIR IS RESHAPED INTO THE SEATS IT STOOD FOR, not dropped: `model` for every
// bulk seat and `judgmentModel` for every judgment seat, on the claude harness, written as `seats` in
// the place of the pair's first key, through the expansion setup's --bulk and --judgment use. A tier
// stated as "" configures no seat, and the manifest records each tier key with its source value. The
// copy reads through the record's own reader, so a migrated run's tier audit judges each seat.
func TestTranslateRunConfigReshapesTheClassTiers(t *testing.T) {
	dir := t.TempDir()
	from, to := filepath.Join(dir, "from.json"), filepath.Join(dir, "out", "to.json")
	seatsOf := func(bulk, judgment string) string {
		spec := func(m string) modeltier.SeatSpec {
			if m == "" {
				return modeltier.SeatSpec{}
			}
			return modeltier.SeatSpec{Harness: "claude", Model: m}
		}
		b, err := json.Marshal(modeltier.TieredSeats(spec(bulk), spec(judgment)))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	for _, tc := range []struct{ name, body, bulk, judgment, reshaped string }{
		{"both tiers", `{"topic":"t","model":"sonnet","judgmentModel":"opus","k":2,"eventSchema":74}`, "sonnet", "opus",
			`{"judgmentModel":"opus","model":"sonnet"}`},
		{"the bulk tier only", `{"topic":"t","model":"haiku","eventSchema":74}`, "haiku", "",
			`{"model":"haiku"}`},
		{"the judgment tier only", `{"topic":"t","judgmentModel":"opus","eventSchema":74}`, "", "opus",
			`{"judgmentModel":"opus"}`},
		{"an empty bulk tier", `{"topic":"t","model":"","judgmentModel":"opus","eventSchema":74}`, "", "opus",
			`{"judgmentModel":"opus","model":""}`},
	} {
		if err := os.WriteFile(from, []byte(tc.body), 0o644); err != nil {
			t.Fatal(err)
		}
		tr, err := translateRunConfig(from, to)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		got, err := os.ReadFile(to)
		if err != nil {
			t.Fatal(err)
		}
		rc, err := record.DecodeRunConfig(got)
		if err != nil {
			t.Fatalf("%s: the reshaped configuration does not read: %v", tc.name, err)
		}
		b, _ := json.Marshal(rc.Seats)
		if want := seatsOf(tc.bulk, tc.judgment); string(b) != want {
			t.Errorf("%s: seats = %s, want %s", tc.name, b, want)
		}
		if !strings.HasPrefix(string(got), "{\n  \"topic\": \"t\",\n  \"seats\": {") {
			t.Errorf("%s: the seats are not in the place of the pair's first key:\n%s", tc.name, got)
		}
		if r, _ := json.Marshal(tr.Reshaped); string(r) != tc.reshaped {
			t.Errorf("%s: the manifest records the reshaped keys as %s, want %s", tc.name, r, tc.reshaped)
		}
	}
	// A source that states the seats already is copied as written, and records nothing reshaped.
	body := `{"topic":"t","seats":{"blue-lane":{"harness":"claude","model":"sonnet"}},"eventSchema":74}`
	if err := os.WriteFile(from, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	tr, err := translateRunConfig(from, to)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Reshaped != nil {
		t.Errorf("a source stating its seats recorded a reshape: %v", tr.Reshaped)
	}
	got, _ := os.ReadFile(to)
	rc, err := record.DecodeRunConfig(got)
	if err != nil || len(rc.Seats) != 1 || rc.Seats["blue-lane"].Model != "sonnet" {
		t.Errorf("a source stating its seats copied as %s (%v)", got, err)
	}
}
