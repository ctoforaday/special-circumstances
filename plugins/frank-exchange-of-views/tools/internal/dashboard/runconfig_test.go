package dashboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE BOARD READS THE RUN'S CONFIGURATION THROUGH THE RECORD, AND SHOWS A REFUSAL WHERE IT DOES NOT
// READ. A file the record refuses — here a key the configuration does not declare — shows why, never
// the values a lenient decode would have picked out of it.
func TestTheBoardShowsAnUnreadableConfigurationAsUnread(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "inputs", "run-config.json")
	if err := os.WriteFile(cfg, []byte(`{"model":"claude-fable-5","lanes":"2"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	m := BuildModel(runtest.Open(t, dir), t.TempDir(), Config{}, 0)
	if m.ConfigUnread != "" || m.Config.Model != "claude-fable-5" || m.Config.Lanes != "2" {
		t.Fatalf("a readable configuration read as %+v (unread %q)", m.Config, m.ConfigUnread)
	}

	if err := os.WriteFile(cfg, []byte(`{"model":"claude-fable-5","noSuchKey":3}`), 0o644); err != nil {
		t.Fatal(err)
	}
	m = BuildModel(runtest.Open(t, dir), t.TempDir(), Config{}, 0)
	if !strings.Contains(m.ConfigUnread, "noSuchKey") {
		t.Errorf("an unreadable configuration read as unread=%q", m.ConfigUnread)
	}
	html := RenderHTML(m)
	if !strings.Contains(html, "the run&#39;s configuration could not be read") && !strings.Contains(html, "the run's configuration could not be read") {
		t.Errorf("the board does not say the configuration could not be read")
	}
	if strings.Contains(html, "claude-fable-5") {
		t.Errorf("the board shows a value from a configuration that does not read")
	}

	// The flags the board was given are the operator's own statement, and still show.
	m = BuildModel(runtest.Open(t, dir), t.TempDir(), Config{Model: "operator-bulk", Lanes: "4"}, 0)
	html = RenderHTML(m)
	if m.ConfigUnread == "" || !strings.Contains(html, "operator-bulk") || !strings.Contains(html, ">4<") {
		t.Errorf("the board dropped the flags it was given beside an unread configuration (unread %q)", m.ConfigUnread)
	}
}
