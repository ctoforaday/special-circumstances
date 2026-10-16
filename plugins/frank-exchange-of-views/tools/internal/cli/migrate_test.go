package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// The publish's warnings are not in the manifest on disk — they arise after it is written — so the
// verb is their only carrier. Each reaches stderr in both modes: under --json the manifest goes to
// stdout as JSON and the text summary is skipped, which must not take the warnings with it.
func TestMigrateWarningsReachStderrInBothModes(t *testing.T) {
	warnings := []string{"the directory /runs/out could not be synced", "the directory /runs could not be synced"}
	for _, jsonMode := range []bool{false, true} {
		c := newMigrate()
		c.Flags().Bool(flags.JSON, false, "")
		if jsonMode {
			if err := c.Flags().Set(flags.JSON, "true"); err != nil {
				t.Fatal(err)
			}
		}
		var stdout, stderr bytes.Buffer
		c.SetOut(&stdout)
		c.SetErr(&stderr)
		reportMigration(c, &migrate.Manifest{SourcePath: "/runs/old", Warnings: warnings})
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
