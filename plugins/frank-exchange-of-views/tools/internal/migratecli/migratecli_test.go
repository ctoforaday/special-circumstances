package migratecli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
)

// The publish's warnings are not in the manifest on disk — they arise after it is written — so the
// command is their only carrier. Each reaches stderr in both modes: under --json the manifest goes to
// stdout as JSON and the text summary is skipped, which must not take the warnings with it.
func TestMigrateWarningsReachStderrInBothModes(t *testing.T) {
	warnings := []string{"the directory /runs/out could not be synced", "the directory /runs could not be synced"}
	for _, jsonMode := range []bool{false, true} {
		c := New()
		var stdout, stderr bytes.Buffer
		c.SetOut(&stdout)
		c.SetErr(&stderr)
		reportMigration(c, jsonMode, &migration.Manifest{SourcePath: "/runs/old", Warnings: warnings})
		for _, w := range warnings {
			if !strings.Contains(stderr.String(), w) {
				t.Errorf("json=%v: warning %q is not on stderr: %q", jsonMode, w, stderr.String())
			}
			if strings.Contains(stdout.String(), w) {
				t.Errorf("json=%v: warning %q went to stdout", jsonMode, w)
			}
		}
		if jsonMode {
			var m map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &m); err != nil {
				t.Errorf("json=true: stdout is not the manifest as JSON: %v: %q", err, stdout.String())
			}
		} else if !strings.Contains(stdout.String(), "migrated /runs/old") {
			t.Errorf("json=false: stdout is not the manifest's summary: %q", stdout.String())
		}
	}
}

// A REFUSAL IS EXIT 2 IN BOTH FORMS: the diagnosis on stderr, named by the tool, or the record tool's
// refusal envelope on stdout under --json — the shape a machine consumer of feov-record already reads.
func TestARefusalExitsTwoInBothForms(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Main(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("no --from/--to: exit %d, want 2", code)
	}
	if want := Name + ": --from and --to are both required"; !strings.HasPrefix(stderr.String(), want) {
		t.Errorf("stderr = %q, want it to open %q", stderr.String(), want)
	}
	if stdout.Len() != 0 {
		t.Errorf("a text refusal wrote stdout: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Main([]string{"--json"}, &stdout, &stderr); code != 2 {
		t.Fatalf("--json with no --from/--to: exit %d, want 2", code)
	}
	env, ok := feov.DecodeRefusal(strings.TrimSpace(stdout.String()))
	if !ok {
		t.Fatalf("stdout is not the refusal envelope: %q", stdout.String())
	}
	if env.Code != string(feov.MissingField) {
		t.Errorf("envelope code = %q, want %q", env.Code, feov.MissingField)
	}
	if stderr.Len() != 0 {
		t.Errorf("a --json refusal wrote stderr: %q", stderr.String())
	}
}
